# Evidence verification cutover

This is a supporting implementation and evidence record for remediation finding
2. It is not an executable contract, a timing qualification, or a coordinated
Reporting adoption record. Core 01 REQ-01-452a, Core 02's blob lifecycle, Core 03's
file workflow, and Core 04 AC-154 own the amended behavior. Core 04 specifies
the exact AC-016/AC-046 physical workload. Its qualified observation and source
identity are recorded in the final section below; later integrated applicability
is tracked in the coordinated adoption register.

## Decision and compatibility

Stored-object hashing now belongs to the Evidence background worker. The public
`POST /api/v1/object-blobs/{object_blob_id}/verify` operation admits an
incident-scoped Common Job without reading object bytes. The existing initial
row-create and attach transactions require its retained successful observation,
bound to the exact completed single-use upload lease. They retain their atomic
record, revision, projection and association effects. Verification itself does
not create an Evidence record, change custody or mark the blob available.

The Base Profile worker is independently selected when optional extensions are
unclaimed. `contracts/jobs/base.v1.json` is the authored operational projection;
the application composes it with recognized and selected extension definitions.
The generic Jobs catalog remains the sole execution authority. No fake optional
extension or second runtime runner was added.

The old synchronous finalization-attempt counter and fourth-failure transition
are retired. Infrastructure failures do not establish corrupt bytes. Missing
verification has no record or blob effect; explicit failed/canceled verification
can be followed by a new Job while the slot remains eligible. Size and expected
SHA-256 mismatches remain terminal on first detection. Pending expiry and cleanup
deadlines remain enforced. This replaces the old attempt-count assertion with
effect-free repeated finalization, real verification failure, and retained-proof
integrity assertions; it does not remove the surrounding lifecycle tests.

Migration 95 locks the source relation, rejects used legacy retry history before
schema mutation, drops the counter, updates the cleanup-protection trigger, and
creates the verification observation relation. It invents no Job or proof for
existing data. Down migration rejects; rollback requires a matching application
and data pair or fresh initialization. Reset authorization remains limited to
explicitly identified disposable environments.

The current recovery generation is
`recovery.current.reference_pack_v2.graph_v4.evidence_verify_v1`. Verification
observations are excluded security-ephemeral state, invalidated with upload
leases. Already available blobs retain their existing observed metadata. The
new catalog digest is
`c433c2f2ab403e4a734b57ab73737ab8ee49c015e9d0bc57dff67fba73951068`.
Previous generations are not translated.

## Mechanism and validation map

| Obligation | Implementation and exact exercising test | Status |
| --- | --- | --- |
| Atomic proof consumption, concealment, single association, exact create replay | `blob_attachment.go`, `initial_blob_create.go`; `TestAttachBlobValidation_Unit`, `TestObjectUploadAtomicEvidenceCreate_Integration` | Focused pass in `remediation-evidence-jobs-atomic-20261016-03`; later source applicability requires integrated review |
| Current incident role, authentication/CSRF order, no deployment-admin bypass | `route_verification_handler.go`, transactional incident admission; `TestEvidenceAuthorizationMatrixAC543_Integration` | Passed within `remediation-evidence-jobs-cancel-20261016-02`; that invocation still failed its separate controlled-read fixture |
| No Jobs/blob lock across object reads; cancellation drains before terminal publication | `blob_verification_worker.go`, `blob_verification_object.go`; `TestEvidenceVerificationCancelsAndDrainsStoredRead_Integration` | Pass in `remediation-evidence-jobs-cancel-20261016-03`; controlled fault test, not latency evidence |
| Unknown/retired contract rejection and base selection without optional claims | `internal/app/jobassembly`; `TestBaseJobContractRejectsUnknownAndInconsistentInputs`, `TestBaseJobsRemainSelectedWithoutOptionalClaims` | Passed in `remediation-evidence-jobs-catalog-20261016-01` (native 0, 1.274 s) |
| Captured admission replay, cancellation replay, stale read rejection, explicit retry, retirement fences | `EvidenceBlobVerification.ts`, `EvidenceUploadSession.ts`; exact retained-file recovery row | Pass in `remediation-evidence-jobs-frontend-20261016-01`, also passed in the mixed cancellation invocation |
| Real browser progress, cancellation, grid continuity and physical timed predicate | File recovery presentation and real browser measurement | Qualified physical root `remediation-evidence-measurement-browser-20261016-04`; separate retained functional cancellation and accessibility checks below |
| Migration rejection leaves data and migration ledger unchanged | Migration 95 preflight | Passed in `remediation-evidence-jobs-source-runtime-20261016-03`; both legacy preflight cases preserve data and migration state |
| Recovery invalidation and graph/catalog binding | Evidence recovery contribution and current recovery fixtures | Affected inventory, process restore and browser restore passed in `remediation-evidence-jobs-recovery-repair-20261016-01` |

Every root above is under `.cartulary/test-results/`. Native outer exits were
collected separately. The successful atomic root exited 0 in 13.849 seconds;
the controlled cancellation root exited 0 in 12.129 seconds. The frontend
recovery root exited 0 in 5.178 seconds. Frontend type checking passed in
`remediation-evidence-jobs-types-20261016-02`, native 0, 22.326 seconds.
These are execution durations, not performance-window acceptance samples.

## Failed observations retained

Generation attempts 01–06 found migration naming/ownership, missing reviewed
OpenAPI additions, and inconsistent recovery catalog/generation bindings.
Attempt 07 passed. Later attempt 09 rejected replacement DDL; explicit
transactional trigger/function replacement corrected it. Attempt 15 passed.
All use the `remediation-evidence-jobs-generate-20261016-` prefix.

The construction root `remediation-evidence-jobs-construction-20261016-01`
failed an outdated recovery-table expectation. Atomic attempt 01 failed a
fixture compile error; attempt 02 exposed a cleanup trigger still referencing
the retired counter. Both causes were corrected before attempt 03 passed.
Cancellation attempt 01 used a wrapper without the required typed object-store
capability; attempt 02 supplied a content type inconsistent with its slot.
Attempt 03 exercised the intended blocked read and passed. These are attributed
implementation/fixture failures, not evidence of intermittent product behavior.

The first type check rejected an explicitly undefined request body and unsafe
test-array indexing. Both were fixed. Initial finalize attempts stopped on stale
generated routing or the missing Jobs family declaration. The subsequent
`remediation-evidence-jobs-finalize-20261016-06` passed with native exit 0 in
20.203 seconds. `RESULTS_DIR` was unset, so retained-run maintenance was skipped.

## Remaining handoff obligations

Complete the full Evidence owner, Jobs composition, migration, recovery and
security checks; inspect the rendered progress/cancellation states; qualify the
AC-016/AC-046 workload; reconcile final source applicability and obligation
review. No all-nine closure, Reporting promotion or benchmark claim is made by
this intermediate record.

## Source capability split and current regression evidence

Evidence now separates transaction-bound `SourceRuntime` capabilities from the
full HTTP/object-storage/worker `OwnerRuntime`. The full runtime composes the
source runtime. Source-only Timeline, Workbook and import fixtures construct the
narrow source capabilities, without fake Jobs or an unavailable-object-store
implementation. The performance fixture contribution explicitly borrows its
source capabilities, database and typed object store; it does not start workers
or transport. This preserves production mutation, revision and projection owners
while removing unrelated fixture dependencies.

`remediation-evidence-jobs-source-runtime-20261016-03` passed the two formerly
failing source fixtures, dependency validation, exported-surface review and the
new `TestEvidenceVerificationCutoverRejectsLegacyHistoryWithoutMutation_Integration`.
The migration test independently rehearses pending legacy attempts and exhausted
legacy attempts at schema 94. Both reject migration 95, preserve the full blob
row and migration ledger, and leave the verification relation absent. All work
is against test-owned disposable databases. Native outer exit was 0; harness
duration was 16.880 seconds. Attempts 01 and 02 retained compile failures from
the capability split (the fixture's old runtime field and an omitted clock
import); both were repaired before this pass.

The full Evidence owner root
`remediation-evidence-jobs-complete-owner-20261016-01` completed all 52 work units
with 51 passing and one failing, native exit 2, harness duration 199.366 seconds.
All 95 cleanup operations completed. Backend, real browser, accessibility and
visual units passed. The failing unit contained two assertions in
`workbookEvidence.test.ts`: the old upload mocks supplied no verification Job.
The corrected mocks preserve the custody/private-target assertions, include
queued admission and successful observation, and check the exact authenticated
verification request. Their focused root
`remediation-evidence-jobs-frontend-20261016-02` passed, native exit 0, harness
3.776 seconds. This focused repair does not relabel the failed aggregate root.

The compact Timeline progress presentation now keeps verification and Cancel
visible with recovery details collapsed. It shares the same source-owned Job
view with expanded recovery and does not move focus. The targeted component
regression in `remediation-evidence-jobs-progress-unit-20261016-01` passed,
native exit 0, harness 2.661 seconds. A new ordinary-browser regression retains
real worker/admission/cancellation behavior while delaying only status delivery
to exercise a completed-worker/cancellation-intent race. Its execution and
rendered inspection remain pending at this checkpoint. No delayed-delivery test
is performance evidence.

## Recovery and compact progress validation

The real browser cancellation-race row passed in
`remediation-evidence-jobs-progress-browser-20261016-02` (native 0, 72.472 s).
Its bounded browser response hold is functional fault evidence, never latency
qualification. The worker succeeds before the cancellation arrives, so the
correct public response is 409 `already_terminal`. Local cancellation intent
still prevents automatic Evidence creation; explicit Resume remains available.
The first attempt expected 200 and failed; the test expectation was corrected
without weakening the product contract.

Artifact review `remediation-evidence-jobs-ui-review-20261016-01` inspected the
1280×720 state with details collapsed, the inspector closed, readable Cancel,
and an independently focused Timeline draft. The private artifact session
completed cleanup and exited 0. This image-only inspection establishes no Axe
result. A later filename label and narrow-layout accessibility test require
fresh inspection.

The full Recovery owner root
`remediation-evidence-jobs-recovery-owner-20261016-01` failed (native 2,
86.416 s; all 53 cleanup operations complete). Causes were an outdated expected
table inventory, an upload fixture still attaching before verification, and a
browser restore fixture lacking the production renderer installation receipt.
Those fixtures now follow the current interfaces. The affected three rows pass
in `remediation-evidence-jobs-recovery-repair-20261016-01` (native 0, 78.501 s),
source `sha256:020476a98e5a8b2aaa865ee60e7b01058f6b1f2eb43654263814f27ad58f83a1`.
The process fixture independently asserts one real verification proof before
capture and zero after restore, then verifies the restored Evidence projection
through the packaged server. This subset closes the three diagnosed failures;
it is not a new full-owner run.

`remediation-evidence-jobs-finalize-20261016-09` passed (native 0, 20.470 s).
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
`remediation-evidence-jobs-types-20261016-06` passed (native 0, 23.817 s).
The preceding invocation used a nonexistent `typecheck` target and was rejected
before execution; the public target is `frontend-typecheck`.
Generation root `remediation-evidence-jobs-generate-20261016-20` reports success
in 9.940 s, but its native session outcome was unavailable after context
recovery, so that outer exit is not independently established.

The narrow progress accessibility row passed in
`remediation-evidence-jobs-progress-a11y-20261016-01` (native 0, 87.332 s,
all 23 cleanup operations complete). It exercises 1280×720, 768×640 and 390×480,
keyboard cancellation, visible focus, control reachability, contrast and scoped
Axe WCAG A/AA checks. Its held Job read is deliberate functional synchronization,
not performance evidence. Image review through artifact session
`remediation-evidence-jobs-ui-review-20261016-02` inspected the 1280 and 390
captures: the long filename wraps, progress and Cancel remain readable, and the
focus outline is visible without overlaying the grid. That session cleaned up
and exited 0 in 67.138 s. Private images expired on stop.

The packaged-server upload check passed in
`remediation-evidence-jobs-process-upload-20261016-01` (native 0, 78.165 s).
Reference Pack degradation passed in
`remediation-evidence-jobs-reference-degradation-20261016-01` (native 0, 28.604 s).
The shared route fixture root `remediation-evidence-jobs-support-routes-20261016-01`
failed compilation because two fixture login structs differ. The verification
helper now accepts only the session and CSRF cookies it needs; its callers retain
their own authentication fixture types. The failed root (native 2, 11.093 s,
`infra` / `preflight_error`, 14 successful cleanup operations) remains retained;
repaired route verification is pending.

The repaired six-part route envelope suite passed in
`remediation-evidence-jobs-support-routes-20261016-02` (native 0, 73.772 s),
including CSRF, current authorization, exact replay/divergence, query-field
conformance, projections and WebSocket consequences. The cookie-only helper
works with each fixture's existing authentication types.

## Exact verification measurement implementation

Core 04 now binds `perf.evidence_verification.visible_with_cancel.v1` jointly to
AC-016 and AC-046. The versioned typed workload references the existing physical
Evidence-heavy v2 fixture without redefining its population. It uses the full
100-object preview preflight, one ordinary editor, a fresh 2 MiB File paste and
verification admission, a distinct first-page Timeline target, and two-frame
progress/cancel/edit proof. Synthesized clipboard evidence is labeled explicitly.
No response hold, worker delay, action retry or larger file is part of timing.

Product marks bind the single accepted transaction to dispatch, admitted Job
and blob. The shared Job paint observer additionally checks the displayed blob
identity. Retained proof admits only the closed synthetic bindings; no source
bytes or capability URLs are serialized. The existing single-observation policy
and evidence finalizer remain the sole timing/evidence path. The per-predicate
proof schema and profile adapter validate the exact fixture digest, traffic,
threshold, distinct source/edit records and complete sampling.

`remediation-evidence-measurement-generate-20261016-03` passed (native 0,
10.010 s), after the first attempt found missing schema registration and the
second found that a unit test incorrectly shared the physical measurement
verification identity. The unit test now maps to ordinary behavior assurance;
only the browser row claims physical measurement. Format 01 passed (6.268 s)
and finalize 01 passed (20.696 s); retained-run maintenance was skipped because
`RESULTS_DIR` was unset.

Detector and retained-upload instrumentation checks passed in
`remediation-evidence-measurement-unit-20261016-01` (native 0, 5.692 s).
Type checking 01 rejected an inferred string instead of the closed Job purpose;
graph validation 01 rejected use of the Timeline-specific fixture digest mapper
for Evidence's different physical fixture shape. The adapter now hashes the
Evidence fixture with its existing canonical representation. Both failures
remain retained; native exits were 2, with durations 23.252 s and 29.208 s.
Fresh corrected validation and physical measurement are still required.

Corrected type checking passed in
`remediation-evidence-measurement-types-20261016-02` (native 0, 21.729 s),
and graph/evidence rejection validation passed in
`remediation-evidence-measurement-graph-20261016-02` (native 0, 37.210 s).
Generate 04 and finalize 02 passed in 10.290 s and 20.749 s; retained-run
maintenance remained skipped because `RESULTS_DIR` was unset.

The first physical measurement root,
`remediation-evidence-measurement-browser-20261016-01`, rejected setup because
`/tmp` is a 16 GiB memory filesystem and the physical population requires about
20 GiB before overhead. Native exit was 2, duration 59.808 s; the producer
classified it `product` / `test_assertion_failure`, while the diagnosed cause is
scratch capacity. No browser sample ran; all 24 cleanup operations completed.
The supported `CARTULARY_HARNESS_SCRATCH_ROOT` override now selects
`/home/jochi/.cache/cartulary/remediation-harness-scratch`, outside the checkout
on the sufficiently sized local disk. No population reduction or cleanup bypass
was used.

Root `remediation-evidence-measurement-browser-20261016-02` sealed the full
fixture (390.194 s construction, 0.123 s semantic validation), but the browser
test timed out after 300 s. The failure snapshot shows the source Analyst editor
focused and no upload progress. Clicking had replaced the display-cell element,
while the driver still awaited that removed display locator for its File paste.
The driver now targets the focused source editor and bounds ordinary interaction
waits at 10 s (navigation 30 s). This changes no workload, timing start, threshold
or retry policy. Source for the failed root:
`sha256:e97967c5f95f03de681ef294ffe8a2c9e04f1e1edfa57654c659fadd3692579f`.
Native exit was 2, total duration 985.717 s, and all 28 cleanup operations
completed. No passing latency claim is made from either failed root.

Root `remediation-evidence-measurement-browser-20261016-03` produced a qualified
168.7 ms observation with zero scheduler overlap. The command nevertheless failed
(native 2, 705.053 s): cleanup checked that Cancel was visible, then attempted to
click after the Job had finished and removed that button. All 28 harness cleanup
operations completed. Its source was
`sha256:81db21c7f4bde7454b10e602e99ae744d610563465145ee684c4b7259ccae494`.
This failed invocation does not close assurance. Cleanup now observes the already
bounded worker to its terminal state and requires success, without introducing a
second action that races completion. Cancellation-race behavior remains covered
by its separate functional test. Workload, predicate and threshold are unchanged.


## Qualified physical verification observation

`remediation-evidence-measurement-browser-20261016-04` passed with observed
native exit 0 in 665.215 s: all 15 work units and all 28 cleanup operations
completed. The one fresh observation was 125.2 ms against the inclusive 1,000 ms
threshold, with zero scheduler overlap and a qualified retained measurement
summary. The exact admitted Job subsequently succeeded. Source:
`sha256:a7f44a54f475b9e46e9fd08a6890823892d84ef0115d8359d2d82624518048fb`.
The preceding format 03 and finalize 04 passed in 6.565 s and 20.980 s; retained-run
maintenance was skipped because `RESULTS_DIR` was unset. The failed roots remain
retained; they are not additional successful observations or a stability sample.
This closes the exact AC-016/AC-046 verification timing workload at that source,
not all remaining remediation or Core 05 publication requirements.
