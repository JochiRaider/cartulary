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

## Frozen acquisition checkpoint

The initial reference source is local commit
`b4830d0732ead2002e684504cb1cf4c29efc03cd` on
`codex/specification-remediation-20261016`. The complete worktree was clean at
freeze. Subsequent supporting-document edits are outside executable provenance;
do not change executable files or move HEAD within a provider window.
The retained upstream OFL license preserves its original trailing whitespace;
the remaining staged diff passed whitespace inspection. Final documentation
lint passed in `remediation-freeze-records-markdown-20261016-02`, native exit 0,
9.680 s.

Acquisition began with public `make check`, explicit
`CARTULARY_HARNESS_CACHE_MODE=cold`, `CARTULARY_OUTPUT_MODE=ci` and physical-fixture
scratch root `/home/jochi/.cache/cartulary/remediation-harness-scratch`.
Run ID: `remediation-cost-reference-check-20261016-01-cold`. Its native session
is 98967 while this checkpoint is live. No outcome is established by launch.
After a successful cold observation, run the manifest's exact warm-up and five
measured roots with cache mode `normal`, preserving the same command, source and
inputs. A failed observation stays rejected; investigate before further work.
The twelve provider windows and fifty target bindings remain prospective until
the sole writer qualifies their retained evidence.

### Qualification guards found during the first cold run

The first frozen `check` remains running while this issue is recorded; let it
drain without changing executable source. Review of
`canonical-performance.mjs` found two remaining enforcement gaps under
REQ-377: the `stability_extension_reason` token is schema-checked but its stated
cause is never reconstructed, and leave-one-out analysis removes corresponding
reference/candidate observations together rather than each single observation
independently. Stable five-sample windows could therefore claim an unearned
sixth, and paired removal could conceal a gate-sensitive individual observation.

Before continuing acquisition, enforce the actual first-five trigger and its
exact declared reason, including matched comparison windows and shared aggregate
bindings. For comparison sensitivity, remove each reference observation with the
candidate unchanged, then each candidate observation with the reference unchanged.
Keep sixth-window instability rejection and the existing statistical gates.
Add adversarial contract fixtures, run the public harness contract, finalize,
and create a successor clean freeze. Retain the first cold root as correctness
evidence if successful and as superseded cost evidence; do not mix it into the
successor window. No threshold, count limit or retry policy is relaxed.

The initial cold check completed successfully: native exit 0, 707.449 s,
799/799 work units, all 1,080 cleanup operations completed, and retained-secret
scan passed over 5,383 files. Its clean source snapshot is
`e50407fac1489b4ac9d04b58b088f1cb17f10f79be37ea92b0e24a8c2c405c4a`
at the frozen commit above; graph
`01038448d002f6e15ad1cd371db15e61e5686391bb7e86ae14ff9b5b53918fab`.
It is correctness evidence for that source, excluded from the successor cost
window as `superseded_validation_policy`. The replacement validator reconstructs
the first-five trigger, propagates matching requirements through aggregate
provider bindings and tests individual-observation sensitivity. Public
`make harness-contract` passed in
`remediation-performance-stability-contract-20261016-01` (53.018 s) and its
successor `remediation-performance-stability-contract-20261016-02` (53.170 s),
both native exit 0. The successor includes complete retained comparison fixtures:
an individual outlier triggers a matched sixth even when paired removal would
hide it; the stabilizing sixth is accepted; an unearned sixth and false extension
reasons reject. The standalone high-MAD extension is independently exercised.

`remediation-performance-stability-finalize-20261016-01` passed with native
exit 0 in 21.320 s. Retained-run maintenance was skipped because `RESULTS_DIR`
was unset; the earlier cold check is not a compatible full warm-check root.
Documentation lint passed in `remediation-performance-stability-markdown-20261016-01`
with native exit 0 in 9.880 s before this result entry. The successor freeze
starts with the complete `release-check` provider to expose integration defects
before collecting the remaining portfolio windows.

The successor source is clean commit
`7ccec27a9d905bcd4d8024df58d67c7d5a73f53d`, executable snapshot
`b3463443139a16a16aa1ec2d820c9ab712295bfdf19ca70f4f6b45fb8642fcb2`.
The first cold release invocation is
`remediation-cost-reference-release-check-20261016-01-cold`, native session
33967, graph `195b03c43c1af2124fcb2108cfdf55cd04a1f2d5ee09621d9b71f0e6480addc9`.
It selects 1,275 work units. At launch no outcome, adoption or qualified cost
window is established. The pending check window uses suffix `02` and excludes
the initial cold check above. Other planned provider identities remain unchanged.

The cold release completed with native exit 0, 6,257.716 seconds, 1,275/1,275
passed work units and no cache hits. All 1,980 cleanup operations completed;
the retained-secret scan passed over 12,070 files. Every target projection
passed, including production rendering, package/independent recovery, visual,
accessibility and all product measurement contexts. The observed dependency
critical path is 6,254.287 seconds. Canonical non-overlapping buckets are
setup 0.047 s, fixture 793.891 s, execution 5,301.945 s, collation 155.738 s,
wrapper 3.230 s and unattributed 2.865 s. Resource-blocking time is contextual,
not an additional charge. Diagnostics were off; no measured CPU or memory cost
is inferred. This is one cold observation and source-bound integrated acceptance,
not a qualified multi-observation portfolio or a Core 05 publication claim.

### Retained finalizer admission follow-up

Review of `agent-finalize-retained-run.mjs` found that its preflight requires a
successful complete `check` roster and matching source but does not enforce
`manifest.cache_mode='normal'`. Harness §8.3 explicitly rejects cold retained
roots before mutation. Add a narrow rejection and an adversarial retained-root
case before claiming final handoff/maintenance closure. Also review the ordering
of complete retained-evidence validation against the first mutating action;
the existing partial preflight must not substitute for complete canonical
admission. This does not affect the active release invocation, which supplies
no finalizer `RESULTS_DIR`, and does not invalidate its ordinary assertions.
Preserve the frozen provider window while it runs; perform the maintenance
repair at the next source-change boundary and retain exact source applicability.

The normative cleanup now states the exact warm cache mode and removes the stale
command-table references to the retired duration-refresh and duration-drift
finalizer actions.
The current closed action registry remains unchanged: generated-structure
refresh is its only mutating action; standalone duration writers keep their own
public admission rules. The implementation follow-up must reuse canonical
retained-evidence validation before that mutation, validate unit/row closure,
and reject prohibited evidence paths before filesystem traversal. No new runtime
catalog or document dependency is appropriate. These guards remain unimplemented
at this checkpoint and must be validated before final remediation closure.

### Finalizer and support-projection implementation checkpoint

The source-change boundary after the successful cold release implements the
pre-mutation guard. The finalizer now uses the canonical event/context readers,
current compiled full-check graph, current host/capacity and shared unit/row and
cleanup validators. It requires normal cache mode and a passing retained-secret
scan. Safe descriptor-relative reads replace its direct retained-file traversal;
the unused recursive walker is removed. Latest eligible sibling selection first
filters manifests and validates candidates without rerunning work. Explicit older
root override remains unchanged.

The integrated Timeline owner audit exposed a separate partition defect:
`remediation-integrated-audit-timeline-20261016-01` failed with native exit 2 in
1.627 s, `config/configuration_error`, all ten cleanup operations complete, and
`unknown public evidence target`. Its catalog-required backend/browser support
projections exist in the successful release root. REQ-121 requires known catalog
partitions; the auditor incorrectly restricted all partitions to public command
entries. The correction admits only current required support projections from
a public provider, with exact compiled graph, command and row closure. An
internal projection without its own command identity binds the enclosing
provider's identity, matching canonical publication. Unrelated private targets
and direct private providers reject. This does not add a command, alias or runner.

The shared validator keeps owner audits and finalizer maintenance consistent.
`remediation-retained-finalizer-contract-20261016-01` passed the exact
`harness.evidence_accounting.behavior.current_epoch_evidence` row through
`make test-slice`, native exit 0, 11.188 s. Cases cover cold/cache-off input,
partial/current-graph and command mismatch, host/capacity mismatch, altered or
absent unit/row evidence, context/scan/cleanup failure, unsafe paths rejected
before traversal, retained sibling selection and both support projection forms.
`remediation-retained-finalizer-finalize-20261016-01` passed with native exit 0,
22.100 s. `RESULTS_DIR` was unset, so retained maintenance remains skipped until
a new compatible full warm check passes. Broader harness validation and the
actual retained-maintenance invocation are still required.

The cold release remains valid integrated evidence for its recorded source.
It is excluded from the successor cost window as `superseded_validation_policy`;
the operational manifest now uses release suffix `02`, retaining the rejected
root and reason. Do not mix the new harness assertions into the older window.
Other unstarted provider windows retain their planned identities. No qualifying
cost portfolio has yet been published.

After the frozen windows, collect one explicitly instrumented full warm check
with `HARNESS_DIAGNOSTICS=basic` for the final resource-cost handoff. Keep it
separate from the default-off performance windows. Use the public performance
and resource explainers to report observed process/group CPU, memory, shared
service context and missing attribution. It must not invent per-test resource
cost or establish an instrumentation-overhead or product benchmark claim.

### Production retained-maintenance admission

The broader `remediation-retained-finalizer-harness-20261016-01` passed
`make harness-contract` with native exit 0 in 55.968 s. The subsequent full
normal-cache `remediation-retained-finalizer-warm-check-20261016-01` reports
799/799 work units passed in 673.556 s, no cache hits, and a passing retained
scan over 5,383 files. Its executable snapshot is
`e8e21db1a28155f29ad025142e995a9e046894a5bc8d4a2fdce3c306222c4524`.
The native session handle was unavailable after context recovery, so its outer
exit is unavailable; the terminal summary is not substituted for that exit.

Actual maintenance against that root,
`remediation-retained-finalizer-maintenance-20261016-01`, exited 2 in 12.954 s.
The finalizer's precise failure was `artifact/artifact_error`: retained scan
policy differs from execution. All ten cleanup operations completed and the
mutating refresh was skipped. The outer target reports `harness/child_target_failure`.
Investigation found an incorrect equality between two independent identities:
the native scanner hashes its scanner patterns, whereas instrumentation hashes
the resource-collection policy. No scan failure occurred.

The finalizer and performance reader now validate the scan schema and successful
outcome without equating those independent policies. Performance windows still
reject mixed scanner identities. The scanner remains the only scan-policy
owner; no patterns or compatibility identities are copied into JavaScript.
Fixtures deliberately use different scanner and instrumentation digests and
retain failed/malformed-scan rejection. The exact current-epoch evidence row
passed in `remediation-retained-scan-policy-contract-20261016-01`, native exit 0,
11.248 s. Fresh full-check maintenance at the corrected source remains required.

The corrected slice's `make agent-finalize` passed in
`remediation-retained-scan-policy-finalize-20261016-01`, native exit 0,
21.820 s; retained maintenance was skipped because `RESULTS_DIR` was unset.
Broader `make harness-contract` passed in
`remediation-retained-scan-policy-harness-20261016-01`, native exit 0,
54.852 s. Documentation lint passed in
`remediation-retained-scan-policy-markdown-20261016-01`, native exit 0,
10.750 s. The successor check cost window will provide the new compatible
normal-cache root for actual maintenance; no old root is relabeled or altered.

### Isolated capture failure during the successor cold check

The corrected freeze is `776b5ad23ab9e0c3ea52d238083b43840a0287c5`, executable
digest `d001bcb0a384ba451d670e4921d624a188490ee52b3c624978bc293fd2dee57a`.
`remediation-cost-reference-check-20261016-02-cold` exited 2, with canonical
`artifact/artifact_error`, 696.031 s, 798 passed and one failed work unit.
All 1,080 cleanup operations completed; the retained scan passed over 5,382 files.
The failed unit `go:7868cac819e3077b:4e0b012e5ef2` selected
`platform.openapi.unit.runtime_contract_parity_4fc98c2b71`. Its bounded diagnostic
is `private child capture failed`; no row result was published. This is not
evidence of a product assertion failure. The capture's cause is unresolved.

The capture adapter now exposes only a closed operation label and allowlisted
error code. Unknown codes and raw cause text remain private. Existing capture
fault fixtures check the operation labels and non-disclosure. No lifecycle,
retry, timeout or capture-security rule was weakened. The isolated affected row
passed with fresh cache-off execution in
`remediation-openapi-capture-diagnosis-20261016-01`, native exit 0, 1.249 s.
That smaller context does not reproduce the failed full-check context or prove
its cause repaired. Preserve the incident as unconfirmed, with one full-context
failure and one narrower diagnostic pass, not a combined failure-rate estimate.

The failed root is excluded as `artifact_error_private_child_capture`; the
successor check window uses suffix `03`. No automatic retry-to-green occurred.
Further full-context execution uses the diagnostic correction and a new source
freeze so a recurrence can identify the failed lifecycle operation.

Diagnostic validation passed with native exit 0 in
`remediation-private-capture-diagnostic-contract-20261016-01` (22.709 s),
`remediation-private-capture-diagnostic-finalize-20261016-01` (21.644 s),
`remediation-private-capture-diagnostic-harness-20261016-01` (54.148 s), and
`remediation-private-capture-diagnostic-markdown-20261016-01` (10.920 s).
Retained maintenance was skipped by that finalizer because `RESULTS_DIR` was
unset. These are diagnostic implementation checks, not a repaired-cause claim
or qualified cost observations.

### First complete provider window and Auth fixture correction

At clean `8b8aa5c53085d0d089d96a7dc85e443fde5ac1a3`, executable snapshot
`297bff537501816b3e0bd5e5810e550fbfddeec6bd6e28a65c526e35fb519ffc`, all seven
`remediation-cost-reference-agent-finalize-20261016-01-*` observations exited 0.
Each completed all ten cleanup operations and passed its retained scan. Canonical
target intervals were cold 21.557 s, discarded warm-up 21.225 s, and measured
21.962 / 21.363 / 21.856 / 21.513 / 21.637 s. Derived median is 21.637 s,
nearest-rank p90 21.962 s and MAD 0.219 s. These are collected observations;
the complete portfolio writer has not yet qualified or published them.
Their canonical finalizer invocation omits `RESULTS_DIR` as prescribed by its
measurement profile; it is not evidence of retained-maintenance acceptance.

The same-source `remediation-cost-reference-check-20261016-03-cold` passed with
native exit 0, 679.694 s, 799/799 work units, no cache hits, all 1,080 cleanup
operations complete and a passing scan over 5,383 files. The subsequent warm-up
`remediation-cost-reference-check-20261016-03-warmup` exited 2 with
`product/test_assertion_failure`: 798 passed and one failed work unit, 665.487 s,
623 cache hits, all 457 cleanup operations complete, scan passed over 4,439 files.

The exact failure is the Auth integration row
`module.auth.integration.user_list_search_tokenizes_identifiers_and_binds_45c3e062a8`.
`TestUserListSearchTokenizesIdentifiersAndBindsCursorScope` incorrectly required
its named viewer to be the first unfiltered result. The server's bootstrap admin
has a random UUID and can legitimately sort before that viewer under the existing
ascending-ID contract. The retained assertion shows exactly that case. The test
must establish that its searched user is beyond the first unfiltered page, not
assert the incidental identity of the preceding user.

The correction retains the exact UUID/email search, pagination-scope denial and
non-admin authorization assertions. It adds an explicit earlier unrelated user
so the former assertion would fail deterministically, then checks the actual
later-page precondition. No product ordering, authorization or cursor behavior
changes. This attributes the observed failure to a test-fixture assumption;
the cold/warm cache policies differ and do not establish an intermittent rate.
The earlier capture incident remains independently unresolved.

Both check roots are retained outside the successor window, respectively as
`superseded_auth_test_fixture` and `test_assertion_failure_auth_search_order`.
The replacement check window uses suffix `04`; the completed finalizer window
retains its original exact-source applicability.

The corrected Auth row passed in
`remediation-auth-search-order-contract-20261016-01`, native exit 0, 13.716 s,
through the exact service-backed slice with cache off. `make agent-finalize`
passed in `remediation-auth-search-order-finalize-20261016-01`, native exit 0,
22.335 s; retained maintenance was skipped because `RESULTS_DIR` was unset.
Markdown lint passed in `remediation-auth-search-order-markdown-20261016-01`,
native exit 0, 9.630 s. Full warm-check validation remains the next required
triggering-context evidence.

### Registration failure isolated from the Auth fixture

The Auth correction is frozen at
`f338b0c3556cee2d59bf2b457f823c96b3ae62ed`, executable snapshot
`5ecfb504aa01eb0b25c22c7ce0c69c6c57c66823d6aeebd0727dda9cf91bd17f`.
`remediation-cost-reference-check-20261016-04-cold` passed with native exit 0,
662.419 s, all 799 work units, no cache hits, all 1,080 cleanup operations complete
and a passing retained scan over 5,383 files. Its warm-up
`remediation-cost-reference-check-20261016-04-warmup` exited 2 with
`artifact/artifact_error`, 694.117 s, 798 passed and one failed work unit,
623 cache hits, all 457 cleanup operations complete and a passing scan over
4,438 files. The Auth search row passed in this full warm context.

The new failure selected
`app.server.unit.evidence_cleanup_activation_and_telemetry_a1b2c3d4e5`, unit
`go:571ba902511e922a:5d3edc6d7053`. Its diagnostic is
`private child capture failed (stage=process_registration, code=unclassified)`;
no row result was published. This localizes the failure to capture ownership
registration, not an application assertion. The earlier OpenAPI capture failure
did not retain a stage and cannot yet be attributed to this same mechanism.

Registration can reject during schema validation, resource recording or host
admission. The next diagnostic change retains only allowlisted native rejection
categories, bounded errno and allowlisted transport codes; process identities,
proof-file names, request values and raw causes remain private. Do not substitute
a guessed process-race repair for evidence of the actual rejection. Both current
check roots remain outside the successor window, respectively as
`superseded_capture_diagnostics` and `artifact_error_process_registration`.
The pending check window uses suffix `05`; it must not start before diagnosis
and validation are complete.

The bounded native diagnostic cases passed in
`remediation-registration-diagnostic-contract-20261016-01`, native exit 0,
22.208 s. The exact affected assembly row passed fresh cache-off execution in
`remediation-registration-appserver-diagnosis-20261016-01`, native exit 0,
8.439 s. That isolated pass does not repair or explain the failed registration.
Public target discovery confirms `test-fast` includes both affected capture
rows and requires no services. It is the next smaller batch context for this
investigation. The public performance explainer rejected the failed warm root
because the assembly row result is absent; no duration or resource attribution
is invented for that unit.

The native diagnostic slice's finalizer passed in
`remediation-registration-diagnostic-finalize-20261016-01`, native exit 0,
21.934 s; retained maintenance was skipped because `RESULTS_DIR` was unset.
Broader harness validation passed in
`remediation-registration-diagnostic-harness-20261016-01`, native exit 0,
53.751 s, and Markdown lint passed in
`remediation-registration-diagnostic-markdown-20261016-01`, native exit 0,
10.650 s. The next frozen provider is `test-fast`; its observations must stop
for diagnosis on any failure, not continue as a retry policy.

### Complete test-fast window and capture investigation

All seven `remediation-cost-reference-test-fast-20261016-01-*` invocations
completed with native exit 0 at clean commit
`4dd7d3b8fceb7164227ef40a8263ea69c8105afa`, executable source digest
`6fc8b6b5057f89135a449380b4760a29140c2cf92728ac78e7c58d446d609c88`
and graph digest
`f8289e3d8dae59385dd0ee78cfe88de7f7870b0c9bbe3b6cd4111de623853785`.
Each passed all 581 work units and its retained-secret scan. The cold run
completed all 590 cleanup operations; each normal-cache run completed all 12.
The cold run
recorded 578 misses and three bypasses; each normal-cache run recorded 578 hits
and three bypasses. These are cache-policy observations, not 581 fresh
executions per warm run.

| Observation suffix | Canonical invocation wall (ms) | Selected interval union (ms) |
| --- | --- | --- |
| `cold` | 396663 | 396080 |
| `warmup` | 114019 | 113473 |
| `measured-1` | 104727 | 104370 |
| `measured-2` | 111059 | 110539 |
| `measured-3` | 103978 | 103625 |
| `measured-4` | 110623 | 110256 |
| `measured-5` | 114557 | 114002 |

The public `explain-run DETAIL=performance JSON=1` reader supplied these
intervals. For the five selected-interval observations, p50 is 110256 ms,
nearest-rank p90 is 114002 ms and MAD is 3746 ms, below ten percent of p50.
No sixth observation is earned by baseline variability. This is the second
collected provider window; qualification and publication still require the
existing v3 writer to validate the complete portfolio.

The application-assembly row executed freshly in the smaller warm batch;
the OpenAPI row was cached. Neither capture failure recurred in this window.
That result does not identify or repair either historical failure mechanism.
The incident remains explicitly unresolved, with bounded native diagnostics
installed and validated for the next occurrence. The full `check` window at
suffix `05` is the next required context; any failure stops acquisition for
diagnosis. Successful full normal-cache evidence is also required for actual
retained-run maintenance.

### Complete full-check window 05 and retained maintenance

All seven `remediation-cost-reference-check-20261016-05-*` invocations completed
with native exit 0 at clean commit `a603188c3d2afacf58e8a39894ac9ed5e1ef6704`,
executable source
`6fc8b6b5057f89135a449380b4760a29140c2cf92728ac78e7c58d446d609c88`
and graph `01038448d002f6e15ad1cd371db15e61e5686391bb7e86ae14ff9b5b53918fab`.
Every run passed all 799 work units. The cold observation recorded 623 misses,
176 bypasses, all 1080 cleanup operations complete and a passing retained-secret
scan over 5383 files. Each normal-cache observation recorded 623 hits and 176
bypasses, all 457 cleanup operations complete and a passing scan: 4439 files for
the discarded warm-up and 4436 for each measured run. Native exits were observed
separately after scanning.

| Observation suffix | Canonical invocation wall (ms) | Selected interval union (ms) |
| --- | --- | --- |
| `cold` | 688590 | 686822 |
| `warmup` | 670974 | 669722 |
| `measured-1` | 703733 | 702456 |
| `measured-2` | 676988 | 675676 |
| `measured-3` | 656940 | 655643 |
| `measured-4` | 651816 | 650636 |
| `measured-5` | 693326 | 692029 |

The public performance reader validated every root and supplied the intervals.
The five measured selected intervals have p50 675676 ms, nearest-rank p90
702456 ms and MAD 20033 ms, below ten percent of p50. No baseline-variability
sixth observation is earned. This is the third collected provider window;
portfolio qualification and publication remain pending the existing v3 writer.
The earlier capture failures did not recur. Fresh assembly-row execution in this
full warm context does not identify or repair their historical causes.

After the window completed,
`CARTULARY_TEST_RUN_ID=remediation-retained-finalizer-maintenance-20261016-02 CARTULARY_OUTPUT_MODE=ci make agent-finalize RESULTS_DIR=.cartulary/test-results/remediation-cost-reference-check-20261016-05-measured-5`
passed with native exit 0 in 29.909 s. Its finalization artifact identifies the
supplied root as valid and latest, with no older-root override. Canonical evidence,
scheduler drift, schema shape, tier coverage and generated-structure actions all
passed; `updated_files` is empty. All ten cleanup operations completed and the
retained-secret scan passed over 23 files. This exercises the repaired preflight
and distinct scanner-policy handling against a real compatible full warm check.
It supersedes the failed maintenance attempt as current validation, without
removing that attempt's evidence. Retained maintenance was performed here rather
than reported skipped for an unset `RESULTS_DIR`.

The completed `test-fast` record passed `make lint-markdown` in
`remediation-test-fast-window-markdown-20261016-01`, native exit 0, 9.740 s,
before this full-check/maintenance entry. The remaining nine provider windows,
full portfolio publication, browser consolidation and coordinated adoption are
still open.

### Complete migration and API compatibility windows

At clean `d672cfc53ee412718ecb7f9455cbe86590e0a0ad`, executable digest
`6fc8b6b5057f89135a449380b4760a29140c2cf92728ac78e7c58d446d609c88`, all seven
invocations in each window below completed with native exit 0. Cold and warm-up
observations remain separate from the five measured samples. The public
performance reader validated every retained root; neither window earns a sixth
observation from baseline variability.

`remediation-cost-reference-migration-drift-20261016-01-*` uses graph
`4d1f16f9f3962cf093eb970b21fbe9f1ffef90a649b97190161b04c869b3c744`.
Every invocation passed six work units, all cache-bypassed, completed all 17
cleanup operations and passed the retained scan over 33 files. The real
PostgreSQL schema bootstrap remains part of this command's workload.

| Observation suffix | Canonical invocation wall (ms) | Selected interval union (ms) |
| --- | --- | --- |
| `cold` | 11536 | 11076 |
| `warmup` | 11662 | 11255 |
| `measured-1` | 12163 | 11442 |
| `measured-2` | 13512 | 12974 |
| `measured-3` | 13793 | 13231 |
| `measured-4` | 13678 | 13133 |
| `measured-5` | 11264 | 10856 |

Measured selected intervals: p50 12974 ms, nearest-rank p90 13231 ms,
MAD 257 ms.

`remediation-cost-reference-openapi-compatibility-check-20261016-01-*` uses
graph `6f2e6b1a10da78a5149b87c02a93bb4f760ff4448e3551650596be84911352b2`.
Every invocation passed four work units, all cache-bypassed, completed all 13
cleanup operations and passed the retained scan over 20 files. No immutable
release schema or declared API change set was altered to obtain these results.

| Observation suffix | Canonical invocation wall (ms) | Selected interval union (ms) |
| --- | --- | --- |
| `cold` | 1027 | 1013 |
| `warmup` | 1011 | 992 |
| `measured-1` | 1017 | 1001 |
| `measured-2` | 1012 | 994 |
| `measured-3` | 935 | 919 |
| `measured-4` | 1023 | 1008 |
| `measured-5` | 1029 | 1015 |

Measured selected intervals: p50 1001 ms, nearest-rank p90 1015 ms,
MAD 7 ms. Five provider windows are collected; complete portfolio qualification
and publication remain pending. The preceding full-check and retained-maintenance
record passed `make lint-markdown` in
`remediation-check-window-markdown-20261016-01`, native exit 0, 10.000 s,
before this entry.

### Complete object-store release-evidence window

All seven `remediation-cost-reference-seaweedfs-release-evidence-20261016-01-*`
invocations completed with native exit 0 at the same clean `d672cfc53` source
above, using graph
`23039836c35953a33c310b3fdbb5a4e01a6aacb98bee568fa7d386db85a3e2b4`.
Each passed nine work units and the retained scan over 42 files. The cold run
recorded one cache miss and eight bypasses, completing all 20 cleanup operations.
Each normal-cache run recorded one hit and eight bypasses, completing all 19
cleanup operations. Compatibility, license and software-bill-of-materials
prerequisites remained in the public command's workload.

| Observation suffix | Canonical invocation wall (ms) | Selected interval union (ms) |
| --- | --- | --- |
| `cold` | 10557 | 10148 |
| `warmup` | 10384 | 9988 |
| `measured-1` | 10378 | 9972 |
| `measured-2` | 10452 | 10004 |
| `measured-3` | 10524 | 10079 |
| `measured-4` | 10474 | 10053 |
| `measured-5` | 10397 | 9963 |

The public performance reader validated every root. Measured selected intervals
have p50 10004 ms, nearest-rank p90 10079 ms and MAD 41 ms. No sixth observation
is earned by baseline variability. Six provider windows are collected; the full
portfolio is not yet qualified or published.

### Complete full-owner slice window

The canonical `test-slice` workload is `OWNER=module.auth`, with `ROWS` omitted.
Public owner discovery reports 69 catalog rows, including 37 service-backed
rows; these compile to 38 execution work units in the full-owner slice. The
service-backed command has a separate window and cannot substitute for this one.
Two read-only attempts to pass `OWNER` to `explain-target` rejected before
execution with `config/usage_error`, exit 2; no run root was reported. The corrected public
`make explain-test-owner OWNER=module.auth JSON=1` and task guide supplied the
owner selection; no command interface was widened to accommodate the typo.

All seven `remediation-cost-reference-test-slice-20261016-01-*` invocations
completed with native exit 0 at clean `d672cfc53ee412718ecb7f9455cbe86590e0a0ad`,
executable digest `6fc8b6b5057f89135a449380b4760a29140c2cf92728ac78e7c58d446d609c88`,
and graph `cd1978901617968e61029e70261e10bceb89b701d2f062f794e5d020d65c70cd`.
Each passed all 38 work units. Cold execution recorded eight misses and 30
bypasses, all 75 cleanup operations complete and a passing scan over 405 files.
Each normal-cache invocation recorded eight hits and 30 bypasses, all 67 cleanup
operations complete and a passing scan over 397 files.

| Observation suffix | Canonical invocation wall (ms) | Selected interval union (ms) |
| --- | --- | --- |
| `cold` | 130934 | 130146 |
| `warmup` | 122357 | 121551 |
| `measured-1` | 123155 | 122343 |
| `measured-2` | 122098 | 121331 |
| `measured-3` | 124041 | 123323 |
| `measured-4` | 158293 | 157480 |
| `measured-5` | 173848 | 172900 |

The public performance reader validated every root. Measured selected intervals
have p50 123323 ms, nearest-rank p90 172900 ms and MAD 1992 ms. Both longer
observations remain in the window; none was discarded or retried to lower the
reported duration. MAD remains below ten percent of p50, so baseline variability
does not earn a sixth observation. Seven provider windows are collected; the
complete portfolio remains unqualified and unpublished.

### Complete service-backed slice window

The canonical command is `make service-backed-test-slice OWNER=module.auth`,
with `ROWS` omitted. The public command supplies its intrinsic
`SERVICE_BACKED_ONLY=1` selection; no undeclared user override was introduced.
The 37 service-backed catalog rows compile to 30 work units, all cache-bypassed.

All seven
`remediation-cost-reference-service-backed-test-slice-20261016-01-*`
invocations completed with native exit 0 at the same clean `d672cfc53` commit
and `6fc8b6b5` executable digest recorded above. Graph:
`90506ab791764fb6a55ad70d86e82f48bd75e6d01b95c3a9ac677755df731f36`.
Every invocation passed all 30 work units, completed all 67 cleanup operations
and passed its retained-secret scan over 341 files.

| Observation suffix | Canonical invocation wall (ms) | Selected interval union (ms) |
| --- | --- | --- |
| `cold` | 183858 | 182725 |
| `warmup` | 155338 | 154653 |
| `measured-1` | 124719 | 123928 |
| `measured-2` | 124498 | 123721 |
| `measured-3` | 123379 | 122661 |
| `measured-4` | 123683 | 122862 |
| `measured-5` | 126618 | 125771 |

The public performance reader validated every root. Measured selected intervals
have p50 123721 ms, nearest-rank p90 125771 ms and MAD 859 ms. No sixth observation
is earned by baseline variability. Eight provider windows are collected. The
remaining windows are `test`, `ci`, `browser-e2e` and `release-check`; the full
portfolio remains unqualified and unpublished until the existing writer
validates every named binding and retained observation.

### Rejected full-test cold observation and targeted repairs

The next frozen provider window started at clean
`051cdb1e59216b6a22d189c964a9ef1fd1f31e34` with executable
digest `6fc8b6b5057f89135a449380b4760a29140c2cf92728ac78e7c58d446d609c88`:
`remediation-cost-reference-test-20261016-01-cold`, public `make test`, forced-cold
cache, CI output and the same external physical-fixture scratch root. Its graph is
`88085d7c0eabd501c8f6977f292c30f87ef683b5baa15f285b596b2059934601`.

The completed row `harness.browser.integration.native_fixture_lifecycle` failed
with `product/test_assertion_failure`, normalized row/unit exit 10. The exact unit
is `go:874f3be29c25ce14:0fffc159bbf9`. `TestNativeFixturePhysicalLifecycle/borrowed`
reported `borrowed suite did not survive exit status 1` at
`internal/testutil/harnessruntime/boundary/fixtures_linux_test.go:276`; the owned
and failed-cleanup cases and `TestStandaloneNativeSession` passed within the same
unit. The failing assertion combines active state, zero live borrowers and a
nonempty service inventory, but the current diagnostic does not identify which
predicate failed. Actual service termination, compatibility change and status
inspection failure remain unconfirmed alternatives. No lifecycle fix or
intermittency claim follows from this error alone.

The native command (session `81884`) completed with outer Make exit 2 after
5,142,670 ms. Its 1,235 work units comprise 1,228 passed, six failed and one
skipped; no work was cancelled. The primary classification is
`product/test_assertion_failure`. Of 1,938 cleanup operations, three failed and
one was blocked; the retained-secret scan passed over 11,641 files. Independent
work and finalizers drained without cancellation. Earlier compact observations
omitted `publication_mode`; this failed root is excluded from performance
qualification and must not be used to claim uncontaminated measurement evidence.

The other failures have distinct dispositions:

| Affected context | Observed mechanism and remediation |
| --- | --- |
| `harness.browser.unit.native_review`, `TestReviewIPC/unresponsive_initialize_write_deadline` | Native settlement returned unresolved after permission denial reading relevant `/proc` ownership proof. Settlement now reobserves within its existing deadline, without signaling an uncertain identity, extending the budget or starting another cleanup attempt. Persistent denial remains unresolved. The historical process's precise exec/exit transition is not established. |
| `module.workbook.browser.verify_system_views_switcher_keyboard_entry_rovi_90a2f62956` | Setup called Playwright `focus()` on the roving tab after shell mount without waiting for incident-access admission. Tabs are disabled during that admission, and `focus()` does not wait for enabled state. The helper now asserts enabled state before one focus operation; keyboard journey assertions and deadlines remain unchanged. The original report does not record the tab's disabled state at the instant of failure, so the historical causal attribution remains a supported setup explanation rather than a captured browser fact. |
| `module.revisions.measurement.change_set` | The product row and browser group passed, but fixture release for `lease-000040` failed, provider close failed and service close was blocked. Measurement rollup and target failures are downstream effects. The retained cleanup projection contains no detailed cause, so this incident is not attributed to the proof-read failure by association. |

Exact-owner recovery used `make test-run-recover` with the original run root and
its verified runtime root/lease binding. It completed with native exit 0 in
2,220 ms as `remediation-full-test-cleanup-recovery-20261016-01`. No service-wide
reset was performed; the original failed result remains unchanged.

Fresh targeted validation of the settlement mechanism passed with cache disabled:
`remediation-scope-settlement-focused-20261016-01`, public `test-slice` for
`harness.evidence_accounting.behavior.native_session`, native exit 0, 1,222 ms.
It exercises both a real temporary Linux proof-read denial and persistent denial
that exhausts the existing deadline while leaving the uncertain process alive.
The two affected native browser-harness rows then passed as
`remediation-native-lifecycle-focused-20261016-01`, native exit 0, 123,297 ms.
The borrowed-suite assertion now emits only the schema's closed state/reason,
compatibility, borrower count and service counts; its historical failure remains
unattributed despite this fresh pass. Neither focused run establishes full-suite
stability.

The exact System views row passed in
`remediation-system-views-entry-focused-20261016-01`, native exit 0, 82,386 ms,
12/12 work units, all 23 cleanup operations completed and 76 files scanned.
The exact change-set measurement passed in
`remediation-change-set-cleanup-focused-20261016-01`, native exit 0, 205,200 ms,
15/15 work units, all 28 cleanup operations completed and 91 files scanned.
Both used cache-disabled execution and passed retained-secret scanning. These
are separate fresh context observations; they do not establish the cause of the
historical cleanup failure. The two native-harness rows completed all 11 cleanup
operations with a passing 96-file scan; the native-session regression completed
all ten cleanup operations with a passing 14-file scan.
