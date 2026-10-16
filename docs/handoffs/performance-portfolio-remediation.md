# Performance portfolio remediation

Status: implementation validation in progress; no qualified portfolio published.

This is supporting review material. Harness REQ-277–281 and REQ-377–380 own
execution evidence and qualification. The reviewed Section 4.3 named membership
projects to the task-surface owner; its cardinality is derived. No test or
generator consumes this document.

## Retained execution boundary

The previous performance writer accepted declared digests without enough retained
preimages to reconstruct them. It also checked timing totals without using the
complete canonical timing/critical-path validator, rejected ordinary cached units
for lacking fresh starts, and allowed baseline publication with excessive MAD
after six observations. These are deterministic validation defects, not measured
performance regressions.

The execution producer now publishes an immutable
`cartulary.harness_execution_context.v1` before child execution. The private
preparation interface moves to v2 and rejects v1 without a compatibility reader.
The context retains the exact resolved graph, projections, source-entry preimage,
manifest semantic digest, bounded host profile and instrumentation-policy bytes.
It excludes launch bindings, cache internals and credentials. The host profile
contains platform, architecture, kernel release and sorted CPU models; it does
not contain hostnames or an environment dump.

Qualification reconstructs the source, graph, per-unit, toolchain, capability,
policy and workload bindings, then reuses the canonical evidence reducer for
intervals, timing buckets, critical paths and pressure. Policy bytes must first
match their source-entry byte digest. The comparison's system profile binds host
and capability facts. Duplicate identities, overlapping/out-of-order samples,
incomplete cleanup, context omission and unknown fields reject. Six-sample high
MAD rejects publication. Retained contexts are not signatures or independent
proof against wholesale fabrication by a trusted producer.

The ordinary service-backed slice has a distinct measurement profile preserving
its intrinsic `SERVICE_BACKED_ONLY=1` input. It still selects `module.auth` with
rows omitted. Optional worker inputs must match the adopted profile, not be
silently discarded to make observations comparable.

Cache hits remain reused work. A warm invocation may measure current waiting
time and acceleration; no cached outcome becomes another fresh behavioral
execution. Shared intervals are unions, never charged once per dependent test.

## Validation record

All roots are under `.cartulary/test-results/`; outer native exits were observed.

| Run ID | Result |
| --- | --- |
| `remediation-retained-context-generate-20261016-01` | `make generate`, exit 0, 9.750 s; authored schemas and registry projected through Make. |
| `remediation-retained-context-format-20261016-01` | `make format`, exit 0, 6.133 s. |
| `remediation-retained-context-native-20261016-01` | Exact `harness.command_surface.unit.native_runtime` row, exit 0, 49.565 s. Publication is present before child execution, immutable, and excludes private launch state. |
| `remediation-retained-context-evidence-20261016-01` | `make harness-evidence-contract`, exit 0, 10.170 s. Complete retained fixtures exercise valid publication and altered graph, source, policy, workload, timing, host, context, cleanup, source cleanliness, toolchain, capacity and ordering rejection. Six-sample high MAD rejects. |
| `remediation-retained-context-finalize-20261016-01` | Exit 2, 13.639 s; supplying the preceding full-check root correctly failed source-digest compatibility preflight after these harness edits. All ten cleanup operations completed. The root is rejected maintenance evidence, not a product regression. |
| `remediation-retained-context-finalize-20261016-02` | `make agent-finalize`, exit 0, 21.339 s. `RESULTS_DIR` unset: retained-run maintenance skipped pending a compatible current full check. |

The finalization checkpoint also validates policy grammar and semantic identity
before native execution. Its additional rejection cases still require current
native/integrated execution. Older functional evidence without a context remains
historical functional evidence; it cannot qualify a new performance portfolio.

## Remaining qualification

Freeze clean executable sources after release repairs. Record one cold, one
discarded warm-up and five consecutive measured observations per provider; use a
sixth only under the existing variability/sensitivity rules. Retain rejected
roots and reasons. Bind all required targets exactly once through direct roots
or valid aggregate projections. Publish only through the v3 baseline writer.
No comparative improvement claim follows from a standalone current portfolio.

The handoff must include per-target waiting time, critical path, non-overlapping
timing buckets, queue pressure, cleanup and cache outcomes. Resource consumption
is reported only at observed process/group granularity; absent instrumentation
or short-lived process attribution remains explicitly unknown. Selection changes
from approved browser consolidation are reported separately from unchanged
workload executor improvements.


## Release integration findings under investigation

`remediation-coordinated-release-20261016-01` completed with outer Make exit 2,
5,868.690 seconds, 1,256 passed / 17 failed / two skipped work units. Its
source digest is `3b1fb28748ceae2500f3cc355adf2a34a04211797a1b0473575fd84c51b80ee4`.
The retained-secret scan passed. Six of 1,981 cleanup records failed or blocked:
both Import discovery fixtures failed release (two records each), provider close
failed, and service close was blocked. Exact owned recovery subsequently passed
in `remediation-release-cleanup-recovery-20261016-01` (outer exit 0, 2.850 seconds),
retiring both browser stacks and their managed suite. This does not repair or
reclassify the original cleanup failure. Its benchmark-claim
contract row rejects `unresolved or duplicate current predicate binding`.
Inspection identifies one missing owner reference: `evidence_heavy_snapshot_v2`
binds `perf.evidence_verification.visible_with_cancel.v1` but omits
`contracts/performance/evidence-verification.v1.json` from `source_contract_refs`.
Add the existing typed contract to that authored profile after the invocation
ends, regenerate, and run the exact release-contract row. No benchmark claim
has been published, and relaxing unresolved-binding rejection is not a remedy.

Before performance windows, reconcile explicit default worker values emitted
by Make with canonical omission policy. Accept only the public owner's declared
default when the profile does not specify that input, retain observed inputs in
the execution-policy record, and reject nondefault concurrency. This uses the
current owner as qualification policy, not to fabricate historical actual inputs.
The current root must drain before executable edits or further regeneration.

The same release root also exposes a private browser-helper import from the
Reference Pack fixture adapter, a stale stateful Extensions profile expectation,
and two Go files edited after the formatting checkpoint. These have direct owner
repairs queued; no compatibility exception is appropriate.

The selected-services boundary's `pure` child expired after its five-minute
context with no row result. Its retained graph requested 19 CPU tokens while
the containing Go unit reserved four. The other three service selections passed.
This is a nested worker-budget/admission investigation, not evidence that the
selected configuration helper assertion failed. The deadline remains unchanged.
The `DETAIL=accounting` observer succeeded; `DETAIL=performance` could not read
the absent row directory and supplies no successful performance analysis.

CSV discovery failed the two-frame progress/cancellation predicate. Its trace
shows the correct editor draft and a queued admission card, but the controller
keeps cancellation disabled until its first current Job read, which returned
terminal success. A fresh admission response, historical replay receipt and
current cancellation authorization must remain distinct in the repair. The
strict predicate, workload and no-delay/no-retry rules remain unchanged. XLSX
discovery passed its browser body but failed physical fixture cleanup; its
observation is not qualified. Cleanup attribution awaits retained terminal
diagnostics. Both seeded UI-review workflows also failed preparation; their
Evidence fixture still attaches uploaded bytes without completing the new
verification Job. None of these failures is classified as an intermittent test
defect from this single source checkpoint.

The inspector visual row subsequently reported differences in
`workbook-inspector-narrow-technical-details`, `workbook-inspector-compact-actions`,
`workbook-inspector-history`, `workbook-inspector-rollback-preview` and
`workbook-inspector-public-error`. An artifacts-only UI review inspected the
technical-details actual, expected and diff images. Differences are localized to
text/metadata placement; the review does not establish their cause or justify
golden replacement. Canonical import was unavailable before enclosing run
finalization, so explicit report-referenced PNGs were inspected as reference-only
images, without DOM or accessibility claims. The private review
`remediation-release-visual-review-20261016-01` stopped successfully with complete
cleanup. The exact inspector row passed at the original clean commit in the
isolated `remediation-original-inspector-visual-20261016-01` invocation (native
exit 0, 47.865 seconds). Retained renderer diagnostics match fonts, styles,
viewport and device scale. Geometry records show changed internal scroll heights
from the Evidence panel's current metadata presentation, with fractional anchor
positions differing by about 0.11 CSS pixels in History. This establishes a
current-source visual delta, not a renderer upgrade. Review of all affected
captures and the complete reconciliation remains required before any refresh.

The credential-capacity assessment failed before producing credential latency
observations, at its container resource-report command. The fixture compiles its
bound client under `umask 077` but does not grant execute permission to the
container's nonroot user. The packaged renderer probe already handles this
boundary explicitly. A narrow executable-mode repair and fresh capacity run are
pending; the generic retained failure alone does not prove that this is the only
prerequisite defect.

Before the final retained-context checkpoint, remove the old cache-off omission
of source entries. Source identity is evidence independent of cache eligibility;
cache-off must retain and validate the same complete source preimage while
keeping cache dependency indices empty. Also reverse the schema dependency:
the public context owns its shared property shapes, and the private runtime-plan
schema references those properties. A public retained format must not depend on
a private execution-plan version.

## Repair checkpoint after release diagnostics

The authored predicate reference, browser facade, current Extensions expectation,
verified Evidence seeding, nonroot capacity-probe executable mode and retained
snapshot observer have been corrected. The Reference Pack package test queried
the retired `export_model_json` column after successful snapshot creation; its
observer now joins the immutable source capture used by the current owner. The
same exact binding assertion remains before and after recovery.

Fresh in-time canonical Import admission enables the cancellation affordance;
replay and late acknowledgement do not. Explicit cancellation still obtains
current authorized Job state before mutation. The focused lifecycle/paint rows
passed in `remediation-import-admission-unit-20261016-01` (outer exit 0,
3.636 seconds, three work units). Native plan, cache-off preimage and owner-boundary
checks passed in `remediation-retained-context-native-20261016-02` (outer exit 0,
70.204 seconds). Generation and format passed. The evidence-contract checkpoint
passed before the subsequent scan/public-envelope correction and is not evidence
for that later change.

A further measurement-boundary defect was identified: the reader used canonical
graph wall time for `public_invocation_envelope`. That path now fails closed.
The optional wrapper receipt excludes initial shell/Make startup and receipt
publication and is explicitly diagnostic under REQ-847; it cannot silently become
complete public timing. The current portfolio will use the existing qualified
`canonical_unit_interval_union` source, including for directly invoked targets.
Its successful intervals and complete dependency projections are reconstructed
from retained canonical events. This is one of REQ-378's admitted timing sources;
it requires no new acquisition runner or diagnostic-to-canonical promotion.
Graph invocation wall time and critical path remain separately reportable at
their defined boundaries. Full caller-to-shell-return waiting time is unavailable
and must be identified as such, not inferred from those values. Basic-mode wrapper
and resource observations, when collected, remain diagnostic and disclose their
exclusions. No complete public-invocation or per-test resource claim follows.
All windows must also contain a successful retained-secret scan bound to the
execution policy. No qualified portfolio has been published.

The public `make harness-contract` checkpoint
`remediation-release-repairs-harness-20261016-01` passed with outer exit 0 in
52.608 seconds, covering the scan/default-input/window/root-reuse/timing-source
rejections. The earlier direct `harness-evidence-contract` command is an internal
target and is not the canonical verification claim. Browser harness repairs
passed in `remediation-release-repairs-browser-harness-20261016-01`: outer exit 0,
193.736 seconds, including the nested selected-service lifecycle and both seeded
UI-review contexts. Its source qualification remains separate from later golden
promotion and any ensuing implementation repair.

`remediation-extensions-profile-repair-20261016-01` passed the exact stateful
availability row with outer exit 0 in 105.177 seconds (12/12 work units), binding
source `9988fd0ba8a1ff90e52290bb5758215fde5dd2c937f40695dbbf42d3f0a125b4`.
The benchmark binding rerun `remediation-benchmark-binding-repair-20261016-01`
failed with outer exit 2 in 13.320 seconds; all ten cleanup records completed.
The missing source-contract reference is fixed, allowing the validator tests to
reach a stale negative case. That case still treats AC-016 as unbound, although
the new Evidence-verification predicate now binds it. Its replacement must test
an explicitly synthetic unbound criterion and separately retain rejection of
wrong evidence supplied for AC-016. The original failure is retained and is not
a measured instability outcome.

The nonroot probe repair passed the actual fixed-profile assessment in
`remediation-credential-capacity-repair-20261016-01`: outer exit 0, 310.954 seconds,
13/13 work units and all 22 cleanup operations complete. This qualifies the
selected Windows/WSL2 package assessment, not additional platforms or formal
cryptographic certification.

`remediation-reference-package-repair-20261016-01` returned outer exit 2 in
304.199 seconds, 14/15 work units, all 24 cleanup operations complete. The repaired
source-capture observer now passes exact snapshot pin binding. Eight package
scenarios completed, including activation, restart, transport rejection and
terminal failure auditing. The production-rendering scenario then rejected its
retired template version 3; its redaction version 4 is also retired. Current
explicit package-fixture inputs must select template version 4 and redaction
version 5. Do not admit legacy versions or relax public admission. Rendering,
portability and recovery after that point remain unexecuted in this root.

The benchmark-contract fixture now uses synthetic `AC-000` for the unbound
criterion case and separately rejects AC-043 evidence relabeled AC-016. Its exact
row passed in `remediation-benchmark-binding-repair-20261016-02`, outer exit 0,
41.133 seconds. The package fixture now requests current template 4 and redaction
5 explicitly. Post-golden finalization passed in
`remediation-post-golden-finalize-20261016-01`, outer exit 0, 22.203 seconds;
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

`remediation-reference-package-repair-20261016-02` returned outer exit 2 in
320.335 seconds, 14/15 work units, all 24 cleanup operations complete. The current
template/profile tuple passed actual historical review and external PDF
rendering. Portability then reached a stale observer for retired
`export_model_path`; the v3 catalog retains immutable `source_capture_path`.
Repair that observer without restoring the retired format. The subsequent
complete encrypted backup, independent source-loss restoration and exact retry
remain unexecuted in this root.

`remediation-import-discovery-physical-20261016-01` passed both exact Import
discovery measurement rows with outer exit 0 in 422.623 seconds: 17/17 work units
and all 32 cleanup operations complete. CSV qualified at 604.6 ms and XLSX at
371.4 ms, one fresh observation each, threshold 1,000 ms and zero scheduler
overlap. This validates the fresh-admission cancellation affordance at source
`700d45dd9e4772407495836300b4525ac4887972c792d3f487d19319bd8c3815`.
The earlier release cleanup symptom remains unattributed historical evidence;
this changed-source pass does not establish that symptom's cause or repair.

The corrected v3 portable-capture observer and retained-policy repair passed
`remediation-reference-package-repair-20261016-03`: native exit 0, 522.683 s,
15/15 work units, all 24 cleanup operations complete. All 14 packaged scenarios
completed through fresh installation, production rendering, portability,
encrypted export, source loss and independent recovery. Exact evidence and
source applicability are recorded in the coordinated adoption register. This
dirty-source correctness run is not a qualified execution-cost sample.

The operational acquisition manifest
`.cartulary/remediation-evidence/performance-reference-20261016.json` now names
the prospective cold, warm-up and five measured roots for twelve providers.
The reviewed public roster derives to 50 target bindings; exact aggregate leaf
projections are used where admitted and the remaining commands are direct.
The manifest is a plan of unexecuted roots until those invocations finish, not
evidence of any observation or a published baseline. Every binding uses
`canonical_unit_interval_union`; any qualified result still requires the sole
writer to validate complete retained contexts and all roster/statistical gates.

`remediation-source-freeze-finalize-20261016-01` passed `make agent-finalize`
with native exit 0 in 21.744 s after the complete owner/package/visual validation
above. Final source cleanup removed extra blank lines at EOF from the
terminal-receipt reader and interactive large-grid fixture adapter; behavior is
unchanged. A second finalization, `remediation-source-freeze-finalize-20261016-02`,
passed with native exit 0 in 21.402 s after the latter cleanup. Retained-run maintenance
was skipped because `RESULTS_DIR` was unset. A clean local freeze commit follows;
the subsequent cost roots must retain that exact commit and executable snapshot.
