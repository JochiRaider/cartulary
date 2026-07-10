# Auth Module Refactor Tracker - Next Iteration

Last updated: 2026-07-09.

Target implementation boundary: `internal/modules/auth`, with supporting platform
boundaries in `internal/platform/authn`, `internal/platform/enterpriseauth`,
`internal/platform/httpapi`, `internal/platform/httpauth`, and assembly in
`internal/app`.

This tracker supersedes the previous contents of
`docs/handoffs/auth-module-module-refactor-tracker-next.md`. The previous tracker
is now a post-remediation baseline and provenance source, not a list of open work.
Do not reopen resolved items unless current code or owner documents show a
regression.

## 1. Authority and inspection baseline

Authority order for this tracker:

1. Core 00 through Core 04 for current implementation conformance.
2. `docs/testing-harness-nlspec.md` for harness mechanics, evidence, cleanup, and
   verification gates.
3. `docs/domain.md` for vocabulary and ownership boundaries.
4. `docs/design.md` only for design direction when frontend evidence is relevant.
5. Current repository code and tests for implementation state.
6. Prior auth handoffs as evidence only.

Inputs inspected for this iteration:

- `docs/handoffs/auth-module-module-refactor-tracker-next.md`.
- `docs/domain.md`.
- `docs/spec/00_document_set_status_and_precedence.md` through
  `docs/spec/04_security_and_permissions_nlspec.md`.
- `docs/testing-harness-nlspec.md`.
- `internal/modules/auth`.
- `internal/platform/authn`.
- `internal/platform/enterpriseauth`.
- `internal/app`.
- `contracts/errors/index.json`.
- `tools/phase1_test_map.json`.
- `tools/phase11_test_map.json`.

Scope constraints:

- Keep feature behavior out of `cmd/*`.
- Keep auth domain and application behavior in `internal/modules/auth`.
- Keep transport, runtime, storage, and auth primitives in `internal/platform/*`.
- Do not hand-edit generated roots or generated harness/topology outputs.
- Do not introduce a new module, public API, migration, compatibility layer, audit
  module, store split, or generated contract change unless an owner document gives
  a durable reason.
- Prefer clean structural fixes over tactical patches, but do not preserve behavior
  merely because it already exists.

## 2. Current state summary

The 2026-07-09 remediation closed the prior immediate Auth Module blockers:

- Production Enterprise Authentication verifier boundaries now exist in
  `internal/platform/enterpriseauth`.
- Provider-manifest loading and reconciliation run through `internal/app` before
  route serving for claimed Enterprise Authentication.
- `Service` now depends on handler-family ports rather than one broad private
  store seam.
- Bootstrap credential authentication is isolated in
  `bootstrapCredentialAuthenticator`.
- Deployment-scoped administrative audit remains in auth/admin ownership, while
  incident membership audit remains outside auth.
- Auth internals call platform-owned HTTP API/auth helpers directly instead of
  exported auth compatibility wrappers.

Those items should stay closed. The next structural gaps are narrower and should
be handled in sequence: request-contract correction, strict request decoding,
enterprise protocol test-flow separation, then helper cohesion.

## 3. Immediate implementation candidates

### AUTH-ITER-001: Enterprise binding reason contract cleanup

Current state:

- Enterprise binding create, rotate, and retire decoders in
  `internal/modules/auth/enterprise_routes.go` use `normalizedOptionalReason`.
- That helper treats non-string `reason` values as omitted and trims strings only.
- Core 01 requires optional enterprise binding reasons to normalize under
  `reason_note_v1`.
- Enterprise binding client-transaction conflicts are currently written with
  `httpapi.ClientTxnConflictError("")`, losing the submitted `client_txn_id` in
  the public error detail.

Recommended remediation:

- Replace `normalizedOptionalReason` with an error-returning decoder for optional
  binding reasons.
- Match existing admin-user action semantics: omitted or `null` reason is absent,
  non-string reason returns `invalid_mutation_payload` for field `reason`, and
  string reason values pass through `authn.NormalizeReasonNote`.
- Preserve current behavior that normalized-empty reasons compare as absent for
  idempotency.
- Pass the submitted `client_txn_id` into enterprise binding conflict errors.
- Do not add new public error codes or reason codes for this cleanup.

Area:

- Implementation and tests.
- Specification or contract work only if owners choose to add a new public reason
  code; default remediation must reuse existing registered surfaces.

Rationale:

- This is active contract drift, not stylistic cleanup.
- The current decoder silently drops malformed operator intent.

Expected long-term benefit:

- Enterprise binding request handling stays aligned with the shared auth/admin
  reason model.
- Future Phase 11 evidence can rely on normalized request state and consistent
  idempotency behavior.

Compatibility or migration impact:

- Previously accepted malformed `reason` payloads become rejected or normalized.
- No database migration, generated contract edit, route change, or compatibility
  layer is required.

Risks of leaving unresolved:

- Binding audit evidence can lose operator-supplied reason data.
- Idempotency behavior can depend on loosely parsed request data.
- Future conformance tests may encode the wrong contract.

Validation criteria:

- Phase 11 tests cover omitted, `null`, empty, normalized-empty, normalized string,
  and non-string enterprise binding reasons.
- Tests assert that client transaction conflict responses include the submitted
  `client_txn_id` detail where the platform helper supports it.
- Existing successful binding create, rotate, and retire flows remain unchanged.

Dependencies and sequencing:

- Do this before broader enterprise handler refactors so later tests lock the
  corrected request contract.
- No owner decision is needed unless implementation requires a new public reason
  code.

### AUTH-ITER-002: Auth request decoding consolidation

Current state:

- Auth request decoding is split across local helpers in `api.go`, `users.go`, and
  `enterprise_routes.go`.
- These helpers decode JSON objects but do not consistently reject duplicate
  object members or trailing JSON.
- Other modules already use `internal/platform/httpapi.DecodeStrictJSONObject` for
  strict request-object decoding.

Recommended remediation:

- Route auth request-object decoding through `internal/platform/httpapi` strict
  object decoding.
- Preserve existing route-owned public error envelopes.
- Map duplicate object members, trailing JSON, malformed JSON, and non-object
  requests to existing registered auth or mutation error surfaces.
- Use `request_not_object` for mutation-family strict JSON failures where that is
  the existing registered route surface.
- Do not introduce a new duplicate-member reason code unless Core 01 and
  `contracts/errors/index.json` owners approve it.

Area:

- Implementation and tests.
- Specification and contract changes only if owners approve new public reason
  codes.

Rationale:

- Core 01 rejects undeclared or ambiguous request-object shape, and local decoder
  drift will compound as auth routes grow.
- Strict decoding is already a platform primitive, so auth should not maintain a
  parallel weaker parser.

Expected long-term benefit:

- One decoding rule for login, credentials, current-account routes, user-admin
  routes, administrative audit mutators if any are added later, and Enterprise
  Authentication routes.
- Smaller future review surface for request-contract changes.

Compatibility or migration impact:

- Ambiguous payloads with duplicate members or trailing JSON fail closed.
- No storage migration, public route change, generated contract edit, or frontend
  compatibility layer is required.

Risks of leaving unresolved:

- Route families can diverge on malformed JSON.
- Security-sensitive fields such as password, TOTP, provider assertion, binding
  subject, and idempotency keys can be interpreted inconsistently.
- Future tests may prove only the helper used by one route family.

Validation criteria:

- Phase 1 coverage rejects duplicate members and trailing JSON for login,
  credential, account, and deployment-user admin request bodies.
- Phase 11 coverage rejects duplicate members and trailing JSON for enterprise
  begin and binding request bodies.
- Existing public error codes and registered reason codes remain stable.

Dependencies and sequencing:

- May land with AUTH-ITER-001, but should precede enterprise protocol-flow cleanup.
- If strict decoding exposes a missing public reason code, pause for owner decision
  rather than inventing a compatibility layer.

### AUTH-ITER-003: Enterprise test verifier flow separation

Current state:

- Deterministic OIDC/SAML verifier overrides are injected through
  `ModuleOverrides` for tests.
- Production handlers branch on `enterpriseVerifierOverride` in Enterprise
  Authentication begin, OIDC callback, and SAML ACS flows.
- Some override branches skip or reorder production-shaped transaction lookup and
  redirect construction.

Recommended remediation:

- Keep verifier injection internal, but make route-handler control flow
  production-shaped in all cases.
- Enterprise begin should create the transaction before redirect construction.
- OIDC callback should resolve transaction state and browser binding before
  verifier execution.
- SAML ACS should resolve relay/provider/request context before verifier
  execution.
- Move deterministic fixture differences into test verifier or redirect adapters,
  not into production handler branches.
- Do not add a new module or public protocol API for this cleanup.

Area:

- Implementation and tests.

Rationale:

- Test-only branches inside production handlers weaken the evidence value of
  Phase 11 fixture tests.
- Enterprise Auth is security-sensitive enough that deterministic tests should
  exercise the same transaction-ordering shape as production verifier flows.

Expected long-term benefit:

- Future Enterprise Authentication changes extend one protocol flow rather than
  maintaining fixture-specific behavior.
- Replay, state, nonce, PKCE, RelayState, and browser-binding checks remain easier
  to reason about.

Compatibility or migration impact:

- Public routes, cookies, redirects, sessions, and provider binding resources stay
  unchanged.
- Deterministic test helper URLs or setup helpers may change.
- No migration or public compatibility layer is required.

Risks of leaving unresolved:

- Fixture evidence can pass through paths that do not prove production transaction
  ordering.
- Later production verifier changes can accidentally preserve behavior only for
  deterministic overrides.

Validation criteria:

- Phase 11 enterprise tests pass with deterministic fixtures while exercising the
  same transaction-first control flow used by production verifier adapters.
- Negative-path tests still prove no session issuance after state, relay,
  browser-binding, nonce, assertion, or verifier failure.

Dependencies and sequencing:

- Do after AUTH-ITER-002 so enterprise request decoding behavior is stable.
- Do before any work that attempts to claim production Enterprise Authentication
  profile evidence.

### AUTH-ITER-004: Auth helper cohesion after contract fixes

Current state:

- Handler-family ports are already narrow and should be preserved.
- Helper logic remains mixed across broad files: request decoding, resource
  building, idempotency hashes, audit resource construction, path parsing, error
  writing, and protocol helpers are interleaved.
- This is manageable today, but it raises future phase-growth cost.

Recommended remediation:

- After AUTH-ITER-001 and AUTH-ITER-002, reorganize helpers inside
  `internal/modules/auth` by concern.
- Keep the same package; do not create a new package, module, or public API.
- Keep route handlers as the HTTP facade and keep durable auth state in
  `internal/platform/authn`.
- Do not split the concrete auth store or introduce a generic audit abstraction.

Area:

- Implementation and regression tests.

Rationale:

- The previous broad service seam was fixed, but local helper cohesion still
  affects maintainability.
- Cohesion work is useful only after behavior-correcting contract fixes are locked
  by tests.

Expected long-term benefit:

- Smaller, more reviewable change surfaces for credential, account, admin, audit,
  and enterprise auth route growth.
- Less risk that future route additions reuse the wrong helper family.

Compatibility or migration impact:

- None expected; this should be behavior-preserving.
- No migration, generated edit, route change, or public API change is authorized.

Risks of leaving unresolved:

- Future auth additions can accumulate in broad files and increase coupling across
  unrelated route families.
- Security-sensitive helper behavior can become harder to audit.

Validation criteria:

- Phase 1 and Phase 11 slices pass with no public route, envelope, cookie, audit,
  idempotency, or error-registry drift.
- Import boundaries still show auth callers using platform-owned helpers directly
  where appropriate.

Dependencies and sequencing:

- Do only after AUTH-ITER-001 and AUTH-ITER-002.
- If AUTH-ITER-003 touches the same enterprise helper area, complete that first to
  avoid a behavior refactor hidden inside file movement.

## 4. Watchpoints that should not trigger refactor work yet

### Runtime enterprise startup ordering

Current state:

- `internal/app/runtime.go` validates config, applies extension claim state,
  reconciles enterprise provider manifests, assembles routes, and starts serving
  in the required order.

Watchpoint:

- Do not refactor this path solely for style.
- If enterprise profile claim or reconciliation behavior changes, extract a
  private `internal/app` helper that makes startup ordering explicit and testable.

Reason to wait:

- Current behavior is correct, and unnecessary movement in runtime startup can
  create claim-gate risk.

### Bootstrap credential authentication

Current state:

- Bootstrap credential auth remains a narrow internal route-limited bridge for
  credential setup.

Watchpoint:

- Do not widen bootstrap tokens into ordinary authenticated principals.
- Preserve mixed session/bootstrap rejection, consumed/expired rejection,
  route allowlists, and no direct session issuance.

Reason to wait:

- The current boundary has continuing security value and should not be generalized
  without owner-backed behavior.

### Auth store backing

Current state:

- Narrow service ports are backed by one concrete `platform/authn.Store`.

Watchpoint:

- Do not split the store, add migrations, or move SQL into auth handlers unless a
  concrete transaction or ownership problem appears.

Reason to wait:

- A store split would increase migration and transaction complexity without a
  current durable benefit.

### Deployment administrative audit

Current state:

- Deployment-scoped administrative audit read remains in auth/admin.
- Incident membership audit remains outside auth.

Watchpoint:

- Do not introduce a generic audit module.
- Add audit shape, redaction, ordering, and authorization tests only when audit
  behavior changes.

Reason to wait:

- `docs/domain.md` and Core 01/Core 04 distinguish deployment administrative
  audit from incident revision or membership history.

### Enterprise Authentication production claim posture

Current state:

- Production verifier and manifest boundaries exist, but the Enterprise
  Authentication profile remains unclaimed by default.

Watchpoint:

- Expanded provider-manifest and reconciliation evidence is claim-readiness work,
  not an Auth Module refactor trigger by itself.

Reason to wait:

- Claiming the profile requires owner-approved evidence across the full AC gate,
  not just handler cleanup.

### Reserved extension route families

Current state:

- Current reserved extension route matching is sufficient for the active route
  surface.

Watchpoint:

- Revisit matching only when new extension route families are added under auth or
  users.

Reason to wait:

- Preemptive route-family abstraction would add complexity without a current
  route conflict.

## 5. Owner decisions required before implementation

The following decisions are not granted by this tracker:

- New public error codes or reason codes for duplicate object members, invalid
  enterprise binding reasons, or more granular enterprise auth failures.
- Claiming Enterprise Authentication production conformance.
- Runtime provider-definition mutation.
- Browser provider editors.
- Self-service enterprise linking.
- JIT users, SCIM, group-to-role mapping, or enterprise-only users.
- Passkeys or WebAuthn.
- Generic audit module creation.
- Auth store splitting or new auth persistence migrations.
- New compatibility layers around corrected request-contract behavior.

Default decisions for this tracker:

- Reuse existing registered public error surfaces.
- Keep Enterprise Authentication unclaimed by default.
- Keep deterministic Enterprise Authentication fixtures test-only.
- Keep auth behavior in `internal/modules/auth` and auth primitives/storage in
  `internal/platform/*`.

## 6. Closed and resolved provenance

Retain these items only as provenance. They are not active work.

| ID | Closed state |
| --- | --- |
| `RB-001` | Production OIDC/SAML verifier boundary was structurally remediated through `internal/platform/enterpriseauth`, manifest reconciliation, and runtime claim handling. |
| `RB-002` | The broad private auth store seam was replaced by handler-family ports backed by `platform/authn.Store`. |
| `RB-003` | Bootstrap-aware credential auth was isolated in `bootstrapCredentialAuthenticator`. |
| `RB-004` | Deployment administrative audit remains in auth/admin; incident membership audit remains outside auth. |
| `AUTH-NEXT-001` | Closed: production enterprise verifier failures normalize to Core 01 closed registries. |
| `AUTH-NEXT-002` | Closed: production verifier negative-path and registry guard coverage was added. |
| `AUTH-NEXT-003` | Closed: Phase 11 claim-gate metadata names the full gate and remains unclaimed by default. |
| `AUTH-NEXT-004` | Closed: Phase 1 E-1-08 metadata uses authenticated `403 authorization_denied` with `required_capability=deployment_admin`. |
| `AUTH-NEXT-007` | Closed: generic auth compatibility wrappers were de-exported or narrowed after caller cleanup. |
| `AUTH-NEXT-012` | Closed for the prior slice: validation commands and run roots were recorded in that handoff. |

Preserve these prior items only as watchpoints:

| ID | Watchpoint state |
| --- | --- |
| `AUTH-NEXT-005` | No broad handler split is currently authorized; use cohesion cleanup only after contract fixes. |
| `AUTH-NEXT-006` | Bootstrap credential auth remains internal and route-limited. |
| `AUTH-NEXT-008` | One concrete `platform/authn.Store` remains behind narrow service ports. |
| `AUTH-NEXT-009` | Deployment audit remains in auth/admin; no generic audit module is authorized. |
| `AUTH-NEXT-010` | Add audit shape tests only when audit behavior changes. |
| `AUTH-NEXT-011` | Preserve runtime startup ordering; extract a private helper only if enterprise startup behavior changes. |

## 7. Workstreams, dependencies, and exit criteria

### Workstream A: request-contract correction

Includes:

- AUTH-ITER-001.

Status:

- Completed 2026-07-09.
- Substantive edits: enterprise binding create, rotate, and retire now reject
  non-string `reason` values through the existing `invalid_mutation_payload`
  surface, normalize string reasons through `reason_note_v1`, preserve
  omitted/null/normalized-empty reason equivalence, and include the submitted
  `client_txn_id` in binding idempotency conflict details.
- Files changed: `internal/modules/auth/enterprise_routes.go` and
  `internal/modules/auth/phase11_enterprise_auth_integration_test.go`.
- Compatibility impact: malformed non-string enterprise binding reasons now fail
  closed; no public route, migration, generated contract, or error-registry
  change was introduced.
- Validation: `make phase-slice PHASE=phase11` passed with run root
  `.cartulary/test-results/20260709T131359Z-p59552`
  (`tests=55 failed=0`).
- Next workstream: Workstream B, strict auth request decoding.

Dependencies:

- Current Core 01 enterprise binding contract.
- Existing `authn.NormalizeReasonNote` behavior.
- Existing registered mutation error surfaces.

Exit criteria:

- Enterprise binding reasons are decoded and normalized consistently.
- Non-string binding reasons fail through the existing mutation error surface.
- Client transaction conflict details include the submitted transaction ID.
- Phase 11 enterprise request behavior is covered.

### Workstream B: strict auth request decoding

Includes:

- AUTH-ITER-002.

Status:

- Completed 2026-07-09.
- Substantive edits: local auth, credential, account, deployment-user admin,
  enterprise begin, and enterprise binding request-object decoding now routes
  through `internal/platform/httpapi.DecodeStrictJSONObject`; duplicate object
  members, trailing JSON, malformed JSON, and non-object bodies fail closed
  through existing route-owned error surfaces.
- Files changed: `internal/modules/auth/api.go`,
  `internal/modules/auth/enterprise_routes.go`,
  `internal/modules/auth/phase1_request_test.go`, and
  `internal/modules/auth/phase11_enterprise_auth_integration_test.go`.
- Compatibility impact: ambiguous JSON request bodies now fail closed; no public
  route, storage, generated contract, compatibility layer, or error-registry
  change was introduced.
- Validation:
  - `make phase-slice PHASE=phase1` passed with run root
    `.cartulary/test-results/20260709T131708Z-p88052`
    (`tests=82 failed=0`).
  - `make phase-slice PHASE=phase11` passed with run root
    `.cartulary/test-results/20260709T131901Z-p18248`
    (`tests=55 failed=0`).
- Next workstream: Workstream C, enterprise protocol flow cleanup.

Dependencies:

- Workstream A may land first or with this workstream.
- `internal/platform/httpapi.DecodeStrictJSONObject`.
- Current error registries.

Exit criteria:

- Auth route families no longer maintain weaker local JSON object decoders.
- Duplicate members and trailing JSON fail closed.
- Public error envelopes remain stable.
- Phase 1 and Phase 11 slices cover representative strict-decoding failures.

### Workstream C: enterprise protocol flow cleanup

Includes:

- AUTH-ITER-003.

Status:

- Completed 2026-07-09.
- Substantive edits: Enterprise Authentication begin now persists the
  transaction before redirect construction; deterministic redirect behavior is
  selected through a private service redirect adapter instead of handler
  branches; OIDC callback and SAML ACS handlers now resolve transaction context
  before verifier execution for both production and deterministic verifiers.
- Supporting edits: OIDC and SAML transaction lookups preserve explicit
  state/provider/browser/relay mismatch semantics when the browser-bound or
  relay-bound transaction exists; deterministic test verifiers validate against
  the resolved transaction.
- Files changed: `internal/modules/auth/enterprise_routes.go`,
  `internal/modules/auth/routes.go`, `internal/modules/auth/enterprise_protocol.go`,
  `internal/platform/authn/enterprise_store.go`,
  `internal/platform/enterpriseauth/enterpriseauth.go`, and
  `internal/testutil/enterpriseauthtest/deterministic.go`.
- Compatibility impact: public routes, cookies, redirects, sessions, binding
  resources, and error registries remain unchanged; no migration or public
  protocol API was introduced.
- Validation:
  - Initial `make phase-slice PHASE=phase11` failed at run root
    `.cartulary/test-results/20260709T132339Z-p52732` with
    `phase11_enterprise_auth_integration_test.go:455` because SAML provider
    mismatch surfaced as `provider_response_rejected`; the lookup was corrected
    to preserve `enterprise_auth_transaction_rejected/provider_mismatch`.
  - Rerun `make phase-slice PHASE=phase11` passed with run root
    `.cartulary/test-results/20260709T132504Z-p76414`
    (`tests=55 failed=0`).
  - `make service-backed-slice PHASE=phase11` was skipped because this
    workstream did not change runtime enterprise startup ordering.
- Next workstream: Workstream D, auth helper cohesion.

Dependencies:

- Workstream B should land first.
- Deterministic verifier fixtures in test support.
- Existing production verifier adapters.

Exit criteria:

- Enterprise handlers use production-shaped transaction-first control flow for
  both production and deterministic verifiers.
- Deterministic fixture differences live in test adapters, not handler branches.
- Phase 11 enterprise tests still pass and retain negative-path no-session
  assertions.

### Workstream D: auth helper cohesion

Includes:

- AUTH-ITER-004.

Status:

- Completed 2026-07-09.
- Substantive edits: strict request-object/scalar decoding helpers now live in
  `internal/modules/auth/request_decoding.go`; enterprise protocol path,
  return-to, completion URL, and deterministic redirect helpers now live in
  `internal/modules/auth/enterprise_protocol_helpers.go`; enterprise public
  error helper constructors now live in
  `internal/modules/auth/enterprise_errors.go`.
- Files changed: `internal/modules/auth/api.go`,
  `internal/modules/auth/enterprise_routes.go`,
  `internal/modules/auth/request_decoding.go`,
  `internal/modules/auth/enterprise_protocol_helpers.go`, and
  `internal/modules/auth/enterprise_errors.go`.
- Compatibility impact: behavior-preserving helper movement only; no public
  route, envelope, cookie, audit, idempotency, generated contract, migration,
  store split, or generic audit module change was introduced.
- Validation:
  - `make phase-slice PHASE=phase1` passed with run root
    `.cartulary/test-results/20260709T132817Z-p8079`
    (`tests=82 failed=0`).
  - `make phase-slice PHASE=phase11` passed with run root
    `.cartulary/test-results/20260709T133004Z-p37107`
    (`tests=55 failed=0`).
  - `make test-fast` passed with run root
    `.cartulary/test-results/20260709T133051Z-p55487`
    (`tests=979 failed=0`).
- Next workstream: final validation and handoff completion.

Dependencies:

- Workstreams A and B must be complete.
- Workstream C should be complete if touching the same enterprise helpers.

Exit criteria:

- Helpers are grouped by concern inside `internal/modules/auth`.
- No public API, route, generated contract, migration, audit module, or store split
  is introduced.
- Phase 1 and Phase 11 slices pass without wire-visible drift.

### Final validation and handoff completion

Status:

- Completed 2026-07-09.
- Final validation:
  - `make agent-finalize` passed with run root
    `.cartulary/test-results/20260709T133346Z-p18930`; retained-run
    maintenance was skipped because `RESULTS_DIR` was unset.
  - `make lint-markdown` passed for the tracker edits.
  - User-run `make check` failed at
    `.cartulary/test-results/20260709T134407Z-p62834` because the supplemental
    strict-decoding regression used an authoritative-looking Phase 1 test name
    without a manifest-owned row.
  - The strict-decoding regression was moved to the existing Phase 1 support test
    surface as `TestSupportPhase1_StrictAuthRequestDecoding`, preserving the
    authoritative Phase 1 ledger while retaining the regression coverage.
  - `make phase-test-name-check`, `make phase-map-check`, and
    `make phase-ledger-drift` passed after the support-test placement correction;
    the phase-ledger drift run root was
    `.cartulary/test-results/20260709T135206Z-p71123`.
  - `make phase-slice PHASE=phase1` passed after the support-test placement
    correction with run root `.cartulary/test-results/20260709T135213Z-p72479`
    (`tests=82 failed=0`).
  - `make check` passed with run root
    `.cartulary/test-results/20260709T135402Z-p1852`
    (`work_units=260/260 tests=947 failed=0`).
- Generated contract, generated-artifact, schedule, migration, lockfile, and
  toolchain drift targets were not rerun because those owner inputs were not
  changed.
- Enterprise Authentication remains unclaimed by default; deterministic provider
  behavior remains test-only.
- No owner decisions were required; existing public error surfaces were reused.

## 8. Risks

Primary risks:

- Accidentally widening public error registries while fixing strict decoding.
- Hiding enterprise protocol behavior changes inside test fixture cleanup.
- Turning cohesion work into a behavior refactor before contract tests are stable.
- Creating a generic audit or store abstraction without owner-backed need.
- Treating Enterprise Authentication implementation presence as a production
  conformance claim.

Risk controls:

- Use existing registered error surfaces by default.
- Keep deterministic fixtures test-only and production-shaped.
- Land behavior-correcting work before file movement.
- Keep storage and transport primitives in platform packages.
- Require owner decisions before profile claims, public API changes, migrations, or
  generated contract edits.

## 9. Recommended validation

For this documentation-only tracker update:

- `make lint-markdown`

For later implementation work:

- AUTH-ITER-001 and enterprise binding request behavior:
  `make phase-slice PHASE=phase11`
- AUTH-ITER-002 Phase 1 auth/admin decoding:
  `make phase-slice PHASE=phase1`
- AUTH-ITER-002 enterprise decoding:
  `make phase-slice PHASE=phase11`
- AUTH-ITER-003 protocol flow or runtime-visible enterprise changes:
  `make phase-slice PHASE=phase11`
- Enterprise startup or service-backed route changes:
  `make service-backed-slice PHASE=phase11`
- Shared backend regression confidence:
  `make test-fast`
- Contract, phase-map, or generated-owner input changes only if explicitly made:
  `make generated-artifact-policy-check`
  `make json-shape-check`
  `make generate-drift`
  `make phase-ledger-drift`
  `make phase-schedule-drift`
- Broader end-of-run verification when risk warrants:
  `make agent-finalize`
  `make check`

When a command fails, record the failing target, run root or summary artifact when
available, and whether the failure appears related to the auth tracker work.

## 10. Handoff notes for the next engineer

- Start with AUTH-ITER-001; it is the only candidate that corrects known
  request-contract drift.
- Treat AUTH-ITER-002 as the foundation for future auth route growth.
- Keep AUTH-ITER-003 focused on removing fixture-specific control flow from
  production handlers without changing public Enterprise Authentication routes.
- Do not begin AUTH-ITER-004 until behavior-correcting tests are in place.
- Do not edit generated files, contracts, migrations, phase maps, or lockfiles for
  the tracker update itself.
- Do not introduce new compatibility layers to preserve malformed request payloads.
- If a remediation requires a new public error reason, stop for Core 01 and
  contract-owner approval before implementation.
