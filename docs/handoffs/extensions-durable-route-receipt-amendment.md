# Durable route receipt Recovery amendment

Status: adopted on 2026-11-02 under the user's standing amendment authorization.
Reviewed candidate SHA-256: `9cf2d52a6b24d84ee08038f820512c8650566998d6a6109f17653f47bb908259`.
Core 01 and Extensions owner text was applied before implementation. This
document records review and adoption; normative requirements remain in the
owning specifications.

## Problem and selected correction

Current operational Recovery excludes and invalidates `route_idempotency` as
security-ephemeral state. These rows are committed domain mutation outcomes,
including the receipts required by pending job finalization and retained owner
admission. Current capture/restore therefore preserves jobs and their input but
removes a required part of their authoritative identity. Both Reporting release
and preview fail finalization after structural restore.

Preserve the existing authoritative receipt once, through the current exact-row
PostgreSQL capture and restore path. Do not synthesize replacement receipts,
relax finalizer checks, introduce a second receipt store, or restore sessions.
This supports continuing production uses: duplicate prevention, exact committed
replay, and atomic owner/job completion. It is not historical codec support.

## Concrete owner amendments

1. Core 01 REQ-01-647: `route_idempotency`, owned by `module.auth`, is
   `authoritative`, `authoritative_required`, and `restore`, using the current
   exact-row PostgreSQL snapshot codec and no invalidation/rebuild algorithm.
   Preserve every committed receipt at the selected consistency point, including
   actor, route, scope, client transaction, normalized request hash, response
   status and owner response. A receipt is not an authentication grant. Replay
   remains subject to the initiating route's current authentication,
   authorization and profile-admission rules. Sessions, pending TOTP enrollment
   and enterprise authentication transactions remain excluded and invalidated.
2. Core 01 REQ-01-661: retain 147 catalog tables and require 115 authoritative
   table units. No DDL change is required. Also reconcile REQ-01-647's obsolete
   four-table Graph wording to the already adopted three-result-table contract.
3. Extensions EXT-REQ-136: required idempotency references include the committed
   initiating route receipt for human-owned jobs. Operational Recovery preserves
   these receipts even while the owning profile is unclaimed. Preservation does
   not invoke the profile, bypass current authorization, or fabricate a human
   receipt for local-operator jobs.

## Cutover, validation and review

The current catalog/contribution and dependent generation digests change
coherently. Prior catalog/digest pairs reject before target mutation; no reader,
converter, environment reset, automatic deletion or deployment is authorized.
Public HTTP shapes, route replay contracts, finalizer requirements, snapshot v3,
journal v7 and completion shapes remain unchanged.

Validate actual current capture/restore with pending frozen release and preview,
exact retained receipt identities, absent restored sessions, rejected old session
cookies, and successful current authorization. Preserve strict malformed or
missing-receipt rejection. Verify the assembled catalog and packaged schema,
current-only admission, generated drift, boundaries and Markdown.

Review must compare the source amendment with Core's authoritative backup rule,
Extensions' required idempotency preservation, and actual Auth receipt writers.
Auth user/credential/enterprise mutation writers store safe user responses;
session and handshake grants use separate excluded tables. Owner-specific
receipt contents and public replay semantics remain with their initiating owner.
No executable requirement, generator or product test consumes this Markdown.
