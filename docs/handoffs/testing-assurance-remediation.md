# Testing assurance remediation

## Baseline and authority

Implementation starts at `00d24aa5354284dfcf6533b8f650e9eb0291ad44`, with a clean
index and worktree. The user approved the nine-finding implementation plan,
coordinated Reporting 1.4.0 / Composition 1.3.0 adoption, a clean pre-production
cutover, and TSV rather than Excel-specific AC-003 conformance.

This handoff is human review material, never executable input. Specification
adoption, implementation completion and assurance completion are separate.
No unexecuted test or historical passing run proves current conformance.

## Current release cleanup blocker and remaining work

The remaining execution-cost windows and baseline publication are waived by the
user. They are not a performance pass. Production rendering and sandboxing are
implemented; coordinated adoption remains unapplied while final assurance is
incomplete.

`remediation-final-release-20261017-04` ran public `make release-check` with
normal cache and CI output at clean implementation commit
`810069fe610fcc29f9bb27ecfaab73d98f1dffb9`. Its executable snapshot is
`59068a74158462c1b9bb3c42ea86fab124dedbb551f41c35e46b5feaec47e10e` and its graph is
`103d9845dcd21fbf9a8bbadb276e541b8a917fcf2ee2deba203dcc10b3462b83`.
The native command returned exit 2 after 6,214,884 ms: 1,272 work units passed,
two failed and one was dependency-skipped. Cache accounting is 236 hits,
447 misses and 591 bypasses; a hit is not fresh execution.

The exact product row `module.imports.measurement.xlsx_discovery` and its browser
group passed. Its owned browser fixture, `lease-000049`, then failed release
with `harness/cleanup_error`, unit exit 12. The corresponding measurement-summary
unit failed with `artifact/artifact_error`, exit 11, because the retained
observation was absent. The measurement target summary was dependency-skipped.
The aggregate therefore reports `artifact/artifact_error`; these are downstream
effects of one cleanup incident, not two independent product assertion failures.

Of 1,743 cleanup records, 1,739 completed, three failed and one was blocked.
The retained-secret scan passed over 11,992 files. Every component target
projection except `browser-e2e-measurement` passed, including native fixture
lifecycle, production rendering/confinement, migrations, portability and recovery.
Those projections do not turn the failed release invocation into a pass or supply
a successful full-owner audit root.

The exact retained acquisition had closed producer and frontend launches; its
backend launch lacked a closure receipt, and no terminal settlement receipt was
present. That locates the unresolved teardown stage but does not establish the
original native process adapter's error. Historical cleanup incidents remain
unattributed; neither a permission-denial cause nor an intermittent failure rate
is inferred by association. Review found that ordinary process settlement fails
immediately on a proof-read error while diagnostic settlement can reobserve
permission denial within its existing deadline. The adopted rule permits bounded
reobservation; this difference alone does not prove the recorded incident's cause
or justify changing lifecycle behavior without a reproducer.

Exact-owner recovery used public `make test-run-recover` with this original result
root and the validated private runtime/lease binding. The separate root
`remediation-release04-recover-20261017-01` returned native exit 0 in 2,360 ms.
Recovery settled the resources and leaves the original release result failed.
No broad service reset or timing qualification was performed by recovery.

Remaining work is:

1. Establish the cleanup failure mechanism, implement any justified correction,
   and validate the affected lifecycle and required release acceptance.
2. Complete the 23 prepared full-owner evidence audits using successful,
   source-compatible roots. The release04 manifests were prepared but not audited
   because that invocation failed.
3. Apply coordinated Reporting 1.4.0, Composition 1.3.0 and Core companion adoption
   after the applicable conditions are satisfied; pin the immutable adoption
   commit separately from the implementation commit.
4. Finish the current closure matrix and reproducible handoff, retaining the
   timing waiver and all failed-run dispositions.

No implementation or adoption edits were made during this status continuation.
The documentation checkpoint uses `make lint-markdown`; the previous implementation
finalizer's retained-run maintenance remains skipped because `RESULTS_DIR` was
unset. No cost portfolio or speed-up claim is made.

## Earlier final-release repair checkpoint

`remediation-final-release-20261017-03` completed at clean
`a3d33e127ee04e96751ee7ee83067aea4b3bad07`, executable snapshot
`e137355bf5225450bc98bd4bb21767937613e5ce4502ebbbc86cc6c8c4e2e06e`.
The native `make release-check` command exited 2 after 6,279.139 s: 1,274
passed and one failed work unit, zero skipped/cancelled units, and 13 cache hits.
All 1,967 cleanup operations completed and the retained-secret scan passed over
12,050 files. Every component target projection except `backend-integration` passed,
including browser acceptance, visuals/accessibility, installed rendering,
packaging, security and independent recovery. This remains a failed aggregate;
passing projections do not authorize an owner evidence audit of this root.

The single failing row was
`harness.browser.integration.native_fixture_lifecycle`, unit
`go:874f3be29c25ce14:0fffc159bbf9`. The borrowed subcase reported
`state=stale reason=tool_changed compatible=false borrowers=0 services=2 running=2`.
Both services survived and the borrower count settled, but the shared workspace
test-services executable changed. The exact publishing invocation was not
retained. Review found that this test's ordinary Go context disables CGO while
other preparation contexts can rebuild the same workspace output with different
build settings. The strict compatibility rejection is correct; the fixture's
assumption that a shared build output remains unchanged throughout its lease is
not. This explains this observed mismatch without attributing every earlier
borrowed-suite allegation to it.

The borrowed fixture now owns a copy of the prepared executable for its complete
lease lifetime. `physicalFixtureExecutable` opens one prepared file, checks that
it is a regular executable, copies it to an exclusively created private test
file, and checks the copy length and completion. The test-owned directory lasts
through cleanup. The existing suite compatibility checks, assertions, owner,
runner and real services remain unchanged; no retry, timeout, skip or product
behavior changes. This isolates a test input under TH-HARNESS-REQ-018/853 rather
than weakening `tool_changed` rejection. The helper has no independent entrypoint.

`remediation-borrowed-tool-lifetime-20261017-01` passed the exact lifecycle row
with cache disabled, native exit 0, 117.069 s, one work unit, all ten cleanup
operations complete and a passing 95-file retained-secret scan. It includes the
borrowed, owned and failed-cleanup cases plus standalone native-session coverage.
Executable snapshot:
`59068a74158462c1b9bb3c42ea86fab124dedbb551f41c35e46b5feaec47e10e`.
Final integrated validation and coordinated adoption remain pending. Remaining
execution-cost qualification is still waived; no new cost window is collected.

`remediation-borrowed-tool-format-20261017-01` and
`remediation-borrowed-tool-finalize-20261017-01` passed with native exit 0
in 7.389 s and 22.343 s. Finalization changed no generated artifact.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset; no
compatible successful full warm check root exists at this executable revision.
Review of the actual release projections confirms that all fifteen target
partitions for the 75-row `harness.browser` owner are present. Its complete
owner audit can therefore use the next successful release root; the earlier
assumption that optional review partitions were absent does not apply.

The follow-up owner review found a production failure-classification race in
the same serving-lease incident. Core 04 REQ-04-113 and Extensions EXT-REQ-213
retain the serving fence and exact acquisition-expiry outcome. A driver's
socket deadline can expire before `context.Err()` publishes cancellation;
checking only that publication returned the transport error instead of the
required active-process reason. Both session-open and lock-attempt failures
now use the same absolute acquisition deadline at error observation. Failures
before that boundary remain unchanged; cleanup time cannot alter the observed
classification. No acquisition/loss-detection timeout or retry policy changes.
`TestLeaseAcquisitionDeadlineClassification`, under the new ordinary Go unit
row `app.server.unit.process_lease_acquisition_deadline`, covers delayed
cancellation publication at equality/after the deadline, pre-deadline transport
failure and cancellation, explicit deadline causes, and unrelated timeouts.
These cases use deterministic observation instants, without sleeps.

The already-started `remediation-final-release-20261017-02` was cancelled
before changing that production source. Native exit was 130, with 135 passed,
1,140 cancelled and zero failed work units after 238.077 s. All 176 cleanup
operations completed; the retained-secret scan passed over 4,849 files. Its
source was clean `ea42436b198437d65b822a02b0e813255e2a2431`, executable snapshot
`8ad29eada03a31cbd3ef9c588f3ff89c665e5b1287e14abd7de3257de090dbec`.
This cancellation avoids qualifying an obsolete source; it is not a failed
test retry or passing evidence. The deadline repair passed
`remediation-lease-deadline-focused-20261017-01`, native exit 0, 16.204 s,
five work units, all 18 cleanup operations and a passing retained-secret scan
over 35 files. It selects the new deadline unit row, both existing PostgreSQL
process/serving-lease integration rows, and the assembly fencing row. Source:
`db35c15ecda29425cf1e5538ee0c443a93ca39eb3ab27b92d03da21974a000f7`.
The existing real database acquisition tests retain their 20 ms deadlines.
The direct assembly assertions and deadline classifier have distinct
responsibilities; neither replaces the other's mechanism. Final integrated
validation remains required before coordinated adoption.
`remediation-lease-deadline-finalize-20261017-01` passed with native exit 0
in 22.049 s. Retained-run maintenance was skipped because `RESULTS_DIR` was
unset; no compatible successful full warm check root exists for this revision.

Final review removed the asynchronous context-status/error-type fallback as
well: a shorter dependency timeout must not impersonate the acquisition
deadline. `remediation-lease-deadline-focused-20261017-02` passed the same four
rows, native exit 0, 17.131 s, five work units, all 18 cleanup operations and
the 35-file retained-secret scan. The final executable snapshot is
`e137355bf5225450bc98bd4bb21767937613e5ce4502ebbbc86cc6c8c4e2e06e`.
Failure classification now depends on the actual acquisition deadline at
observation, preserving every earlier transport/cancellation failure. This
fix changes no deadline, retry policy, lease identity or mutation boundary.
Finalization `remediation-lease-deadline-finalize-20261017-02` passed, native
exit 0, 22.228 s. Retained-run maintenance was skipped because `RESULTS_DIR`
was unset. The implementation is frozen for the next integrated release check.

`remediation-final-release-20261017-01` ran `make release-check` from clean
`dfa95e49a3c3cdca8120d979cc0fc92f084e8ae1`, executable snapshot
`58db062dcc1905ac09a665d4b8d773cd4484a5df7f40a2c325c00196c645a536`.
The native command exited 2 after 6,378.056 s: 1,272 passed and three failed
work units, no skipped or cancelled units, 128 cache hits. All 1,853 cleanup
operations completed, and the retained-secret scan passed over 12,000 files.
This invocation is failed evidence, not final release acceptance. Remaining
execution-cost qualification remains waived; this was a correctness invocation.

The failures have two independent mechanisms:

- `app.server.integration.single_active_process_fencing_6a57fb3221` failed
  while identifying a PostgreSQL session under the test's 20 ms acquisition
  deadline, before observing lock exclusion. The assembly test now holds a
  real exclusive serving session, verifies application-start denial, releases
  it, directly observes exclusion while the application serves, and proves
  acquisition becomes possible after application shutdown. Application overlap
  still checks `ErrApplicationProcessActive`; the platform's acquisition-policy
  integration tests retain that classification. The same test identity, real
  PostgreSQL/object-store fixture and runner remain. This removes an incidental
  round-trip timing assumption, not the serving-fence obligation. That
  assertion consolidation changes no production code; the separate owner
  repair above preserves the required acquisition-expiry failure reason.
- The `web.design.visual.incident_import_workflow` renderer failed at Docker
  start, before producing group evidence. The visual aggregate consequently
  failed with missing group results; there was no screenshot mismatch. The
  old diagnostic retained no cause, so port collision is unconfirmed for that
  historical failure. Review nevertheless found a concrete check-then-bind
  race: a host port was released before Docker could reserve it. Docker now
  atomically assigns host ports, and the harness admits only exact, unique
  IPv4 loopback bindings for the requested container ports. Ordinary Playwright
  and optional diagnostic endpoints translate the admitted container/host pair.
  Invalid binding, startup failure and partial acquisition still fail closed
  with owned cleanup, and bounded diagnostics expose closed reason tokens.
  No retry or golden refresh was added. Docker's
  [published-port contract](https://docs.docker.com/get-started/docker-concepts/running-containers/publishing-ports/)
  supplies the ephemeral host-port mechanism; the harness owns admission and
  lifecycle under TH-HARNESS-REQ-255/552/853.

Both behavioral and execution reviews above were performed by Codex under the
user-authorized plan. Narrow checks passed with native exit 0:
`remediation-serving-fencing-focused-20261017-01` (17.147 s) and
`remediation-renderer-ports-contract-20261017-01` (3.416 s). The latter covers
ordinary/diagnostic bindings, malformed/non-loopback/duplicate ports, partial
acquisition cleanup, exact endpoint translation and value-free diagnostics.
Actual renderer and diagnostic-path checks also passed with native exit 0:
`remediation-renderer-ports-visual-20261017-01` (90.459 s) and
`remediation-renderer-ports-diagnostic-20261017-01` (93.328 s). The former
compares the unchanged incident-import golden; the latter exercises the seeded
public review workflow and its translated diagnostic endpoint. Both bind
executable snapshot `5d43ae800c41e121451d40f7a69d457a0e19c16572c551ffbfbe01512705c4c5`.
The historical Docker-start cause remains unconfirmed. Coordinated adoption
and final integrated source applicability remain open.

All four focused runs completed cleanup (14, 11, 23 and 12 operations,
respectively) and passed retained-secret scans (24, 14, 74 and 17 files).
Finalization attempt `remediation-final-release-repairs-finalize-20261017-01`
exited 2 because the authored harness changes made the generated topology
index stale. `remediation-renderer-ports-generate-20261017-01` then passed
`make generate`, native exit 0, 10.970 s, updating that index through its owner.
The failed finalizer root remains retained; no generated file was hand-edited.
The successor `remediation-final-release-repairs-finalize-20261017-02` passed
with native exit 0 in 22.194 s. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset; there is no compatible successful full warm check root
for this executable revision. The earlier successful maintenance checkpoint
remains retained. These repairs are ready for final integrated verification.

## Historical implementation checkpoints

Latest implementation: Reporting production PDF/SVG rendering, confinement, frozen
assets and render recipes have qualified owner and packaged evidence in the
[Reporting review](reporting-cutover-adoption-review.md#render-recipe-owner-qualification).
Evidence now verifies uploaded bytes in a bounded, cancelable Base Job before
finalization; its [cutover review](evidence-verification-cutover.md#qualified-physical-verification-observation)
records the qualified 125.2 ms physical observation, ordinary upload/recovery
checks, accessibility and cancellation-race evidence. The exact Reference Pack
predicates now have full-owner and physical qualification in the
[timed assurance review](timed-product-assurance-review.md#reference-pack-progress-implementation).

The [current adoption register](reporting-coordinated-adoption.md) reviews every
amended Reporting/Composition clause, companion decision and assurance family.
Composition local validation v2, ordinary lifecycle and current-authority replay
checks now pass the full owner suite. The repaired integrated `make check`
passed all 799 work units at its recorded source checkpoint. The
[performance evidence review](performance-portfolio-remediation.md) records the
additional retained-context corrections before qualified windows are collected.

Coordinated Reporting adoption, final source applicability and reviewed browser
consolidation remain open. The user waived remaining execution-cost qualification;
the partial portfolio is retained without a publication or speedup claim.
Earlier implementation
checkpoints retain their own source identities and do not automatically qualify
later source states. All nine findings are not yet closed.

The final retained-policy repair has now passed the complete Reporting owner
and its explicit evidence audit. Current packaged Reference Pack qualification
passes production rendering, portability, encrypted export, source loss and
independent recovery through all fourteen scenarios. The reviewed visual refresh
has both required ordinary passes against the same final manifest; see the
[golden review](ui-ux/remediation-golden-review-2026-10-16.md). Integrated acceptance
used clean `7ccec27a9d905bcd4d8024df58d67c7d5a73f53d`. The cold release invocation
`remediation-cost-reference-release-check-20261016-01-cold` passed with native
exit 0 in 6,257.716 seconds: 1,275/1,275 work units, no cache hits, all 1,980
cleanup operations complete and retained-secret scan passed. The preceding cold check passed 799/799 work units
at its recorded source, but is excluded from the successor cost window.
The [performance review](performance-portfolio-remediation.md) records the
validated stability guards and subsequent retained-finalizer/support-projection
repairs. Their focused evidence-contract and broader harness-contract checks
pass. The new full warm check reports 799 passed work units and a passing scan;
its native outer exit was unavailable after context recovery. Actual retained
maintenance then exposed an incorrect equality between scanner and instrumentation
policies. The corrected reader and adversarial fixtures pass their focused row;
fresh full-check maintenance remains required before closure. No subsystem candidate has yet been promoted and the
Saved View scenario remains pending reference-cost qualification.

The successor clean cold check at `776b5ad23` recorded one private-child-capture
artifact failure, 798 passed units, complete cleanup and a passing retained scan.
The affected OpenAPI row passes its smaller fresh diagnostic context. Its cause
remains unconfirmed; the capture now reports closed operation/error-code labels
without exposing private cause text. The performance review records the failed
root and excludes it from qualification. A new frozen source is required for the
next cost window; no failed observation is relabeled as a warm-up or successful run.

The next cold check passed all 799 units at `8b8aa5c53`. Its warm-up exposed an
Auth test assumption: a random bootstrap-admin UUID can sort before the test's
named viewer. The corrected test checks the searched user's later-page position
and explicitly includes an earlier unrelated user. Its fresh service-backed
row passes; UUID/email matching, cursor-scope denial and non-admin denial remain
exercised. The failed and superseded check roots remain excluded from cost
qualification. The finalizer's seven-observation window is collected, with
portfolio qualification still pending. Exact outcomes are in the cost review.

At `f338b0c35`, the cold check again passed all 799 work units. The full warm
context passed the repaired Auth search row but failed a separate application-
assembly capture before row-result publication. The diagnostic identifies
`process_registration`; native rejection/transport classification is being
validated before a focused batch reproduction. The subsequent bounded diagnostic
fixtures and smaller `test-fast` batch passed. All seven cost-window invocations
at clean `4dd7d3b8f` passed, with five measured intervals at p50 110.256 s and
MAD 3.746 s. No capture failure recurred; its cause remains unresolved. Cost
collection and actual retained maintenance remain open; the failed warm root
cannot satisfy either. The performance review records exact identities,
cleanup, cache posture and the next full-check context.

The complete `remediation-cost-reference-check-20261016-05-*` window now passes
at clean `a603188c3`: all seven invocations exited 0, all 799 work units passed
in each, cleanup completed and retained scans passed. Its five measured intervals
have p50 675.676 s and MAD 20.033 s. Actual retained maintenance then passed in
`remediation-retained-finalizer-maintenance-20261016-02`, native exit 0,
29.909 s, selecting the final measured root as valid/latest and changing no
files. This closes the retained-finalizer validation prerequisite. Three provider
windows are collected; the full portfolio, browser consolidation and coordinated
adoption remain open. Historical capture causes remain unresolved and are not
declared repaired by these passes.

The migration, API compatibility, object-store release-evidence, full Auth slice
and service-backed Auth slice windows are also collected at clean `d672cfc53`.
All seven invocations in each completed with native exit 0, complete cleanup
and passing retained scans. Their exact scopes, durations and variability are
in the performance review. Eight of the twelve provider windows are collected;
`test`, `ci`, `browser-e2e` and `release-check` remain. This is acquisition
progress, not portfolio qualification or closure of the remaining findings.

The first full-test cold observation was rejected after native fixture/review,
System views setup and change-set cleanup failures. The
[performance incident record](performance-portfolio-remediation.md#rejected-full-test-cold-observation-and-targeted-repairs)
retains all failures, exact-owner recovery and applicability limits. Temporary
native ownership-proof denial now stays pending within the existing settlement
deadline; persistent uncertainty remains fail-closed. The browser setup waits
for incident-access admission before focusing its entry tab. Native ownership,
native fixture/review, System views and change-set measurement slices passed
with cache disabled and completed cleanup. Borrowed-suite and historical
change-set cleanup causes remain unattributed. No finite pass streak is treated
as proof of stability.

The native cleanup repair implements Harness TH-HARNESS-REQ-804's single
physical cleanup attempt and TH-HARNESS-REQ-853's exact ownership/settlement
proof. `TestSessionRuntimeProof` reaches the temporary and persistent procfs
denial cases through `testDiagnosticAuthority`, under
`harness.evidence_accounting.behavior.native_session`. The native fixture and
review contexts remain `harness.browser.integration.native_fixture_lifecycle`
and `harness.browser.unit.native_review`. Their ordinary fixtures, Linux native
runner and retained cleanup outcomes are preserved. Reobservation within the
original deadline is not another cleanup attempt or proof of absence. The
existing failed incident remains retained; new passing evidence does not
attribute every historical cleanup failure to this mechanism.

The repair finalizer `remediation-full-test-repairs-finalize-20261016-01`
passed with native exit 0 in 22.789 s and no generated changes. Retained-run
maintenance was skipped because `RESULTS_DIR` was unset: the prior successful
full warm check has the earlier executable digest. Its earlier successful
maintenance checkpoint remains retained. Harness evidence contracts and Markdown
lint passed before the new frozen performance window.

## Workstream status

| Finding | Owner | Status | Exit evidence |
| --- | --- | --- | --- |
| 1: discovery and selection | Harness catalog and source owners | CLOSED for admitted execution profiles | Bidirectional supported-source closure, reviewed source dispositions and fresh catalog-owner validation in `remediation-template-assets-catalog-20261015-01`; behavior coverage and release qualification remain separate |
| 2: timed assurance | Timeline and Core 04 | IN_PROGRESS | AC-003 v2 and reviewed timed-predicate evidence |
| 3: obligation traceability | Behavior and harness owners | IN_PROGRESS | Reviewed assertions, contexts and removal survivors |
| 4: execution cost | Harness observability | QUALIFICATION WAIVED BY USER | Roster and validator repairs implemented; eight provider windows retained. Remaining windows/publication were explicitly waived. No qualified portfolio or speedup claim. |
| 5: stability evidence | Incident-specific behavior/harness owner | CLOSED for reviewed incidents | Local-asset imports and embedded-entrypoint CommonJS preparation reproduced the native PDF failure mechanism. Content-sensitive negative tests and fresh native/full-owner execution passed after repair; see the Reporting renderer-contract-15 checkpoint. Earlier allegations retain their recorded dispositions; no general stability guarantee is inferred. |
| 6: structural assertions | Workbook and architecture owners | CLOSED for identified assertions | Identified spelling guards retired after reviewed survivors; fresh owner/detector checks in the structural closure record below; no claim that static checks prove arbitrary runtime behavior |
| 7: browser consolidation | Workbook, Timeline and harness browser | IN_PROGRESS | Obligation matrix and evidence-preserving survivor decisions |
| 8: legacy/support inventory | Harness diagnostics | CLOSED | Wrapper retired after predecessor/survivor review and execution; no alias or active routing; external provenance and current non-traversal checks retained |
| 9: Reporting adoption | Reporting, Composition and companion owners | IN_PROGRESS | Clause comparison, production validation and coordinated adoption record |

## Immutable Reporting baselines

At commit `78a4effc23aac4b46ca414e11b91e17892f790c7`,
`docs/reporting-subsystem-nlspec.md` is adopted Reporting 1.3.0 and
`docs/report-composition-nlspec.md` is adopted Composition 1.2.0. Their immutable
text supplies the comparison baseline; it is not proof that every companion
dependency or current implementation satisfies that baseline. The current
successors remain proposed until their promotion conditions have been reviewed
and validated. Proposal-only tests cannot close an adopted-base discrepancy.

## Session evidence

| Action | Result | Interpretation |
| --- | --- | --- |
| Baseline Git inspection | Clean at the commit above | No pre-existing modifications |
| `make help-all` | Passed | Public command discovery only |
| `make task-guide ROLE=module-author OWNER=harness.test_catalog` | Passed | Narrow selection: `test-slice`; broader: `harness-graph-contract` |
| `make task-guide ROLE=module-author OWNER=module.reporting` | Passed | Unit and service-backed owner slices |
| Source/canonical-selector comparison | 225 unselected Go functions | Investigation inventory, not test execution or a permanent count target |

## Closure and handoff

Overall implementation and assurance closure are incomplete. Production
rendering, confinement, owner validation, packaged portability/recovery and
visual refresh and integrated release have passed at their recorded sources.
The retained-finalizer preflight repair now passes actual maintenance against
the compatible final warm-check root. The qualified performance portfolio,
reviewed Saved View removal and coordinated adoption remain open. The first real
`RESULTS_DIR` maintenance attempt rejected before mutation; its diagnosis,
correction and successful successor are retained in the performance review.
No adoption or portfolio qualification claim is made by this handoff.

## Closure applicability

Findings 1, 5, 6 and 8 have distinct closure claims. Finding 1 closes selection for
the admitted execution profiles; the current catalog owner passed with native
exit 0 in `remediation-current-catalog-owner-20261015-01`, 2/2 work units,
29.462 seconds and all 11 cleanup steps complete. The source reconciliation
records that source identity and the new Reporting helper dispositions.
Finding 5's earlier fragmented-evidence review is complete: the incident table explicitly
retains unconfirmed historical/cleanup allegations, separates deterministic
defects from intermittent outcomes, and does not infer a rate from changed-source
passes. It does not declare the unconfirmed cleanup symptom repaired. The native PDF incident was subsequently reproduced and repaired under the
renderer-contract-15 checkpoint; its fresh native and owner validation closes
that incident without implying a general stability guarantee.
Finding 6 closes the identified structural assertions through the reviewed survivors
and fresh detector checks in the structural continuation below. It does not
claim static analysis proves arbitrary runtime behavior.
Finding 8 closes retirement and external-support scope through the recorded
predecessor review, passing named survivor, absent active wrapper routing, and
current discovery boundary checks. No external reference tool was adopted.

The subsequent generation change-set omission, missing profile-scope admission
check and incomplete private-adjunct test profile have exact failed and repaired
roots in the Reporting review. They are deterministic findings; no retry-to-green
policy, fresh-execution denominator or intermittent failure rate is inferred
from those changed-source runs. Required behavioral, timing, cost, consolidation
and coordinated adoption work remains open. The final handoff must revalidate
source applicability, not silently treat these records as evidence for later
changes. The closure documentation passed `make lint-markdown` with native
exit 0 in `remediation-current-closure-markdown-20261015-01` before this
applicability paragraph.

## Performance membership review

The Section 4.3 command registry now owns performance membership in its existing
target rows. The initial projection had 52 entries. Workload review removes the
owner-evidence audit and optional benchmark-claim validator from fixed performance
membership; the resulting cardinality is derived, never an acceptance target.
Product test, static/security, migration/generation validation, package
and recovery, aggregate and finalization commands retain
their distinct costs. Read-only help, installation, destructive reset, baseline
publication and raw build readiness commands remain outside this measured
portfolio; their prerequisite cost remains inside the invoking command envelope.
Four existing machine commands lacked normative registry rows: credential
capacity assessment, cryptographic policy assessment, package inspection and
local release packaging. Their entries now state scope and side effects. This
review authorizes measurement scope, not successful qualification or external
publication. No qualified windows have yet been produced.

The two removed timing bindings do not remove validation duties. Owner audits
consume caller-selected partitions, and the benchmark validator can return a
no-claim pass without validating a publication. Neither defines a stable workload.
This follows the existing scheduler-diagnostic disposition. The artificial
`retained-module.auth-slice-evidence` pathname and unused audit timing profile
are retired without replacement. Costs inside measured parent invocations remain
included. Historical baseline rows remain incompatible investigative evidence;
the existing v3 writer must publish a newly qualified portfolio.

## Reviewed obligation and removal matrix

Reviewer for the following implementation reviews: Codex, under the approved
remediation plan. These reviews do not substitute for required execution or
coordinated adoption evidence. Exact catalog row IDs for new source selections
are in [the source reconciliation](testing-assurance-source-reconciliation.md).

| Obligation and authority | Applicability and implementation | Exact surviving assertion/context | Review and status |
| --- | --- | --- | --- |
| Harness REQ-800/AC-063: supported source selection is complete, unambiguous and bounded | All admitted Go, Vitest and Playwright roots; catalog resolver/discovery | `catalog_selector_scope` in `contract-suite-support.mjs`, including source-discovery fixtures; `catalog package closure rejects a silently removed helper test` in `test-catalog-metadata-fixtures.mjs`; isolated filesystem fixtures, no services | Fresh catalog owner pass and 90-test evidence-contract pass; supported-source routing demonstrated, behavior coverage remains separately reviewed |
| Harness REQ-019/800: repeated catalog loads observe current source and pin identity | Resolver cache lifetime | `catalog_selector_scope`: source replacement, imported suite edits, pin change, duplicate registration and symlink checks | Facts remain per load; scratch roots moved outside the checkout |
| Core 04 AC-003, REQ-157/158: supported TSV insertion and visible completion below two seconds | Base Timeline, large-grid/25-session fixture | Existing sentinel renamed to `pastes 20 rows of two-column TSV starting at an existing Timeline row`; browser stack, ordinary auth/CSRF | The sentinel remains functional coverage; the separate v2 measurement passed in `remediation-timeline-measurements-20261014-01` with 356.3 ms and qualified evidence |
| Inspector ownership, semantic focus and capability boundary | Workbook/Timeline composition; grid-adapter interface | `Workbook inspector owner boundaries` exact five expanded cases in `workbookInspectorOwnershipArchitecture.test.ts`; browser-unit, no services | AST imports reject forbidden coordinator/construction dependencies even through aliases; allowed owner calls may be renamed or extracted. Private call spelling and model-property placement assertions retired |
| Semantic grid identity, separate shell fallback identity | Same inspector test | Existing semantic registry guards and required `GridHandle` types | The newly enrolled old test incorrectly forbade all `dataTestIdSelector` use. Current continuity uses it only for fixed shell fallback controls; row/editor identity is supplied by registries. Guard narrowed to row/editor selectors; no product behavior change |
| Harness REQ-053 and current target-plan v4 contract | Diagnostics, no services | `test-canonical-target-plan.mjs`: repeat-byte identity, unique units, dependency/fixture closure, compact/detail output, unknown-target and option denial | Recovered predecessor at `aecef149d^:tools/harness/smoke/target-plan-smoke-suite.sh`. Retired v3 flat-plan shapes, raw package aggregates and old fixture-budget names are not compatibility obligations. Current Go scheduling/fixture contracts retain their own tests. Survivor passed in `remediation-target-plan-survivor-20261013-05/harness-smoke-print-target-plan`; enclosing `make run-harness-smoke-extended` exited 0. Broken private wrapper deleted without an alias; no active routing references existed |
| Domain authority boundary and Harness REQ-800 | External reference material | Documentation-tree pruning exercised by non-traversable sentinel links in discovery tests | UI/UX upstream Python tests remain external support, never executable project assurance |

### Cross-workstream obligation register

Unless a row specifies an immutable subsystem version, Core and Harness references
mean the baseline text at `00d24aa5354284dfcf6533b8f650e9eb0291ad44` together with
this remediation's explicitly described amendments. The final adoption record
must pin the resulting commit. Reviewer is Codex under the user-approved plan;
that designation records this review, not independent approval by another owner.
The tables linked below are detail views of this register, not executable catalogs.

| Scope / governing obligation | Implementation and applicable context | Exact evidence / assertions | Coverage status and remaining discrepancy |
| --- | --- | --- | --- |
| 1 / Harness REQ-800, REQ-013: independent inventory and operational selection | Catalog facade, supported Go build profile and shared TS parser; no service setup before selection validation | Catalog selector fixtures reject uncataloged sources, unsupported registration, duplicate underlying identities, stale selection and forbidden roots; the source reconciliation lists exact operational rows and their fixtures | Routing demonstrated by the integrated check and adversarial contracts. Previously unselected browser rows have the exact passing roots below; no inference of all product behavior coverage |
| 2 / Core 04 REQ-157–159, AC-003 v2: insertion, exact mapped cells, real acceptance-to-paint time | Timeline paste owner and single-observation adapter, ordinary browser stack and 25-session large grid | `module.timeline.measurement.paste_20x5_v2`; `timeline-paste.spec.ts` checks 20 new IDs and all 100 values; observer negatives reject stale/offscreen/one-frame completion and threshold equality; exact roots in timed assurance review | Demonstrated for retained source. Synthesized clipboard event only. No Excel application or Core 05 publication claim |
| 2 / other timed Core 04 criteria | Each owner retains its distinct predicate, fixture, start/stop and sampling | [Timed assurance review](timed-product-assurance-review.md) contains one row per timed criterion, exact AC-043 tests and missing mechanism-specific evidence | AC-043, AC-044, presence, rollback and anchored-trace predicates demonstrated for their retained sources. Mechanism-specific Job predicates now have retained qualified observations; the Evidence metadata-shell v2 physical workload qualified in the retained run recorded below; frozen-source integrated applicability still requires review |
| 3 / Harness REQ-800 removal rule and REQ-013 metadata boundary | Human review; behavior and execution/evidence evaluated separately | Every actual removal is in the preceding table or the six named Network Flow helper conversions in source reconciliation. No machine contract consumes this register | Demonstrated for removals already made; future consolidation requires its own surviving assertions and contexts. Unmapped tests remain retained for review |
| 4 / Harness REQ-377–380 and adopted Section 4.3 membership | Canonical performance owner, v3 writer and exact provider/target bindings | `test-canonical-performance.mjs` validates statistics, gates and closure; canonical-event cases validate ordering, interval accounting and hostile input; retained run manifest/summary/event inputs use protected reads | Validator assertions pass and eight provider windows are retained. Remaining qualification/publication explicitly waived by the user; no qualified portfolio or speedup claim |
| 5 / Harness §§8–11: outcome, cause and evidence comparability | Actual product/test/fixture/runtime owner of each incident | Incident tables below name failed roots, cause, repairs and fresh successor roots; no cached success enters a fresh-execution denominator | Attributed expectation and fixture failures have focused repairs. Historical instability and isolated cleanup failure remain explicitly unconfirmed; no disablement or skip is inferred |
| 6 / architecture ownership and Harness REQ-020 implementation freedom | Inspector composition imports and semantic capability interfaces | `workbookInspectorOwnershipArchitecture.test.ts`, `WorkbookInspectorDeclaredPanelList.test.tsx`: forbidden ownership dependencies fail, declared panel order/ref/focus and concealed/deleted/creation cases survive harmless call refactors | Demonstrated for this candidate. Other structural assertions require invariant-level review before retirement; exact public identities remain legitimate checks |
| 7 / Core 03, Design and dedicated evidence owners | Workbook/Timeline pure model, storage/transport, component and real-browser layers | Mechanism register below identifies contexts that cannot be replaced merely by shared choreography | Reviewed distinct mechanisms remain. Only the redundant SavedView API scenario is removed under the completed equivalence review below; all three fresh survivors passed after removal. Final integrated applicability remains open. Reference cost qualification is waived, with no savings claim |
| 8 / Harness REQ-053 reachability; Domain authority boundary | Current target-plan suite and source-discovery pruning | Named target-plan survivor and external-reference provenance above; predecessor obligations reviewed before deleting wrapper | Implemented and exercised. No alias, new runner or incidental external-test execution |
| 9 / Reporting 1.4.0, Composition 1.3.0 candidate; immutable adopted 1.3.0 / 1.2.0 bases | Reporting/Composition admission, storage, rendering and companion owners | [Clause and promotion register](reporting-cutover-adoption-review.md) maps every changed clause, identity/pack/canonical/portability assertion and adoption condition | Production rendering, template/deck derivation, confinement and selected recovery/portability have retained qualification. Direct Composition review repaired ignored validation context through explicit local validation v2, strict immutable validation, stage ordering and ordinary lifecycle/authority evidence; the full Composition owner passes. Final integrated validation and coordinated adoption remain open |

### Workbook/Timeline mechanism register

All browser rows below use the catalog-owned Chromium context with ordinary
application authentication, PostgreSQL and object storage. Owner manifests remain
authoritative for exact fixtures, isolation and service dependencies. Component
survivors do not replace these contexts. These distinct mechanisms are retained;
the separate SavedView review below authorizes only its named duplicate's removal.

| Observable obligation / coverage dimensions | Smallest preserving layer and exact existing coverage | Consolidation decision / required evidence |
| --- | --- | --- |
| Typed parsing, mapping and validation; scalar and collection editor families | Pure model/unit and hook tests; `useTimelineClipboardPasteController.test.ts` asserts only admitted gestures reserve creation presentation | Keep pure cases narrow. Native clipboard, managed focus, virtualization and server insertion remain browser obligations |
| Roles and authorization changes: viewer/editor/reviewer/admin, incident loss versus account loss | Server owner integration plus `workbook-ordinary-security.spec.ts`; newly enrolled viewer-reference scenario also checks disabled actions | No pairwise reduction of authorization with mutation/replay. Review each distinct denial and retained-authoring effect |
| Uncertain mutation replay preserves captured identity and exact request bytes | `module.workbook.browser.inspector_exact_replay`, exact title `Inspector uncertain recovery replays exact requests without consuming newer authoring` | Retain real browser/transport context; a component promise stub cannot prove admission/replay semantics |
| Successful mutation followed by refresh failure retains acknowledgment and newer focus | `module.workbook.browser.inspector_detached_refresh`, exact title `Inspector acknowledged refresh recovery preserves newer grid focus and never dispatches another patch` | Distinct from uncertain failure; cannot merge merely because both expose recovery controls |
| Typed query membership, accepted versus requested state and failed/delayed replacements | `module.workbook.browser_stateful.exact_literal_tag_set_membership_fcb35d55a3` and `module.workbook.browser_stateful.group_requested_replacement`; literal Timeline/Notes memberships plus Hosts/group keyboard cases | Preserve collection-membership and asynchronous replacement mechanisms; these functional cases do not establish AC-044 latency |
| History subject replacement, rollback and destructive failure effects | `module.revisions.browser.source_history_spec_ts_45f9ef2541` retargets an already open history panel; existing history/rollback owner suites retain mutation assertions | Retargeting is not equivalent to reopening the panel. Keep source identity and old-content removal assertions together |
| Inspector geometry, narrow/zoomed layouts, keyboard shortcuts and retained reference drafts | `module.workbook.browser.source_sentinel_spec_ts_38af00ea72` and `module.workbook.accessibility.source_workbook_inspector_edit_spec_ts_34633fc46b`; declared-panel component cases cover owner-provided structure | Geometry and browser accessibility are independent evidence. Reference summaries consume accepted wire display metadata; long-text controlled rendering is separately covered in `GenericInspectorReferenceSummary.test.tsx` |
| Collaboration, focus/viewport continuity and version bindings during live updates | Collaboration/Timeline browser owners and the dedicated AC-047 seeded large-grid trace recorded in the timed assurance review | Retain independent editor family and concurrency interactions until the trace and obligation review establish equivalence |
| Visual, accessibility and timed acceptance | Dedicated `workbook.visual.spec.ts`, `workbook.a11y.spec.ts` and measurement contexts | No cross-context substitution. Timing windows become incomparable when selection/workload changes unless separately qualified |

### Saved-view consolidation candidate review

The route-foundation browser scenario
`module.savedviews.browser.the_browser_authenticated_public_saved_view_list_3ccd8d0d58`
performs public HTTP calls after opening Workbook; it does not exercise saved-view
interaction. Behavior-owner review recommends consolidating it into the existing
HTTP/storage and actual UI contexts below. Harness-owner review preserves their
ordinary authentication/CSRF, real persistence, catalog reachability and browser
integration. The missing initial-empty assertion and named survivors passed
before removal. The user subsequently waived remaining timing qualification,
removing the reference-portfolio prerequisite. The single duplicate scenario and
its catalog row are now removed. Generated projections were refreshed through
Make, and the fresh survivor checks below passed.

The governing Base Profile obligations are Core 01 REQ-01-138–145 and
Core 03 REQ-03-012/013/022/023 for resource identity, visible listing and
canonical create defaults (AC-146–153/231); Core 03 REQ-03-295 owns saved layout
application/capture. Core 01 REQ-01-146/147 and Core 03 REQ-03-024–026 own the
retained update/delete/duplicate behavior (AC-152/231/480). The initial empty
list is fixture characterization through the public route, not a new product
requirement. The versions are the adopted Core text at the implementation
baseline together with this remediation's recorded amendments. Codex reviewed
both behavior and execution/evidence equivalence under the authorized plan.

| Existing assertion | Named survivor and actual mechanism | Decision |
| --- | --- | --- |
| Fresh incident lists no saved views | `TestSavedViewCreateDefaults_Unit`, row `module.savedviews.store.ordinary_saved_view_create_defaults_omitted_scop_c2e245f7d7` | Add an ordinary authenticated initial GET assertion before creating a saved view; do not replace it with a database count. |
| POST defaults private scope, trims display name, binds Timeline and canonicalizes empty query/layout | Same real-server/PostgreSQL test, including `requireEmptyCanonicalQueryJSON`, `requireDefaultLayoutJSON`, stored-resource equality and exact listed names | Existing assertions preserve and extend this behavior, including ordinary viewer membership and denial cases. Despite its historical function suffix, the catalog correctly classifies this as service-backed integration evidence. |
| Created view appears through the public list route | Same real-server list assertions plus browser row `module.savedviews.browser.verify_saved_view_create_update_select_default_u_ce0bb1647d` | Retain browser selection/create/update/default interactions and their actual public requests. |
| Workbook opens using canonical saved-view query and persisted layout | UI row above and `module.savedviews.browser_stateful.browser_saved_view_query_layout_state_user_home_cb9c681674` | Retain actual browser rendering, routing, reload and state context. |
| Duplicate/delete buttons are absent before the action menu opens | No adopted requirement forbids those supported lifecycle actions; current UI owner tests cover their authorized menu flow | Retire this incidental pending-feature assertion with the obsolete route-foundation scenario. It is not negative authorization evidence. |

This removes one API-only browser scenario, not a distinct focus, concurrency,
visual, accessibility or timing mechanism. Cost accounting must distinguish its
actual execution interval from shared browser setup; no full fixture cost may be
charged as a per-scenario saving. The selected portfolio changes after removal,
so changed selection cannot be presented as unchanged-workload executor speedup.

The initial-empty public-list assertion is now present in the service-backed
survivor. `remediation-saved-view-survivors-20261016-01` ran all three exact
survivor rows above through `make test-slice OWNER=module.savedviews` and passed
with outer exit 0 in 82.466 seconds. The source snapshot is
`9988fd0ba8a1ff90e52290bb5758215fde5dd2c937f40695dbbf42d3f0a125b4`.
This is fresh assertion/context evidence, not a qualified execution-cost window.
This evidence precedes removal. The timing waiver permits the reviewed
consolidation to proceed without claiming a measured saving or equivalent
selection for the earlier cost windows.

After removal, the same three exact survivor rows passed in
`remediation-saved-view-consolidation-survivors-20261016-01`, outer exit 0,
76.592 s, 15/15 work units and all 30 cleanup operations completed. The retained
secret scan passed over 109 files. The executable snapshot is
`58db062dcc1905ac09a665d4b8d773cd4484a5df7f40a2c325c00196c645a536`.
`make generate` and `make format` passed in the matching `consolidation-generate`
and `consolidation-format` roots (11.210 s and 7.015 s). This closes the reviewed
removal's behavior and execution equivalence; integrated source applicability
remains the final release check's responsibility. No unique security, focus,
concurrency, clipboard, visual, accessibility or timing context was removed.

The upstream UI/UX reference source is `ui-ux-pro-max` release `v2.15.0`, commit
`a38d04c3d5c298c851dbe5e6ee1965ee3de42cb5`, retained with the MIT license and
original metadata under `docs/cartulary-ui-ux-refactor-digest/upstream/`.
Its ten Python test modules remain provenance-bearing external reference
material. This remediation neither modifies nor operationally adopts that tool.

## Stability incident dispositions

| Incident | Evidence and compatibility | Disposition and next proof |
| --- | --- | --- |
| Historical browser-instability allegations in the supplied audit | No exact compatible pass/fail run pair supplied | Suspected, unconfirmed; no test disabled and no retry policy added. A future incident requires source/selection/seed/order/environment, fresh-run denominator, failure classification and cleanup outcome |
| Catalog slice failure `remediation-catalog-20261013-01` | Deterministic assertion expected obsolete package-only error wording; verbose diagnostic root `remediation-catalog-diagnostic-20261013-02` identifies the assertion | Attributed test defect; corrected expected missing source identity. Fresh `remediation-catalog-20261013-02` passed. The failed roots remain evidence, not performance samples |
| Inspector slice failure `remediation-inspector-20261013-01` | Newly enrolled structural guard rejects legitimate shell fallback control lookup | Attributed overbroad test invariant; narrowed to semantic grid/editor identity. Its detector cases passed within the complete fast run `remediation-complete-fast-20261014-01`; that enclosing run failed three unrelated assertions and is not an aggregate pass |
| Missing shell assertion detail in retained catalog failure | Graph row has null diagnostic; direct command summary references absent logs and selects an unrelated success line | Attributed diagnostic defect. Error headlines now prefer actual diagnostics over source excerpts. Finalization captures bounded private child output and retains redacted diagnostics before cleanup, preserving earlier failures when cleanup also fails. Contract checks passed in `remediation-diagnostic-evidence-20261014-01`, `remediation-finalizer-diagnostics-20261014-01` and `remediation-finalizer-cleanup-precedence-20261014-01`; actual finalization passed in `remediation-registration-finalize-20261014-01`. Prior failed evidence remains unchanged |

## Browser consolidation decision

No browser scenario is removed merely because its file is large. Newly
reconciled history, inspector, accessibility, and support scenarios remain
active. Parsing/state transitions remain unit-owned, transport/persistence and
authorization remain integration-owned, while focus, clipboard, virtualization,
collaboration, accessibility, visuals and timing retain browser-specific
contexts. Consolidation requires measured cost and a reviewed named survivor
for every assertion and context. The reference portfolio is not qualified yet,
so no reduction is represented as an executor speedup.

### Remaining structural review

The initial inspector replacement does not close every structural assertion.
Further source inspection identified the obligations below. This table records
the earlier review state; the structural retirement continuation later in this
handoff supplies their final dispositions and execution evidence.

| Test owner and assertion | Classification and retained invariant | Required disposition |
| --- | --- | --- |
| `workbookSurfaceOwnershipPolicy.test.ts`: literal inspector composition calls, JSX component spelling, `config.panels.map`, `return children(sections)` | Incidental realization around a legitimate inspector ownership and panel-admission boundary | Retired after `remediation-inspector-invariants-survivors-20261014-01` passed (3.140 s, outer exit 0): AST owner-import/capability checks plus `Declared inspector panels` component assertions preserve declared order, focus/ref forwarding, concealed/missing-role access, deleted/history-only state and creation attachment. Other domain-workflow and transport assertions in this file remain pending review. |
| Same file: exact registration owner IDs | Contract identity in the authored Workbook registration owner | Retain exact bounded-context identifiers and uniqueness review; these are operational identities rather than local variable names |
| `workbookLayoutPolicy.test.ts`: `syntheticGridRows` and row-count variable names | Incidental realization around the adopted geometry/no-filler behavior | Use existing layout metric and browser surface geometry assertions as named survivors; retain AST geometry-owner checks with renaming/comment counterexamples |
| `clientTransactionIdPolicy.test.ts`: proximity regex and `clientTxnRef` | Incidental spelling around a security-relevant transaction-ID contract | Bind the detector to generator/import ownership and use randomness/uniqueness behavior tests; a harmless comment or private ref rename must not decide security conformance |
| `networkFlowBoundaryPolicy.test.ts`: `openClaimedNetworkAnalysis` helper spelling and raw coordinate variable names | Incidental realization around claimed-profile evidence and semantic grid interfaces | Preserve runtime-profile qualification and adapter type/import boundaries; replace helper-name checks with actual admitted-context assertions |
| Network Flow generated schema/token and design-variable identities; font-bundle manifests | Public contract or adopted design identity | Retain exact identities where contractual. Separate executable code/token inspection from arbitrary comments and labels |
| Transport and source-ownership policy tests | Adopted ownership and dependency constraints | Existing AST/import-manifest checks retain independent detection value; review direct/global/aliased references and benign property-name counterexamples |
| Entity-merge and administrative audit fixtures | Behavioral contract fixtures, not source-spelling guards | Retain; filesystem fixture loading alone is not a structural-test defect |

The shared frontend import policy already owns vendor isolation, assessment-to-
Timeline direction, stateless inspector presentation and composition boundaries.
New checks should extend that owner where it expresses the invariant; a second
architecture registry is not needed. The structural retirement continuation below closes the identified follow-ups.
This is not a claim that static checks prove every architectural obligation.

### Reporting structural-check retirement review

The four private-name scans in `internal/modules/reporting/boundary_guard_test.go`
are retired after the named survivors passed in
`remediation-reporting-invariants-survivors-20261014-01` (2.779 s, outer exit 0).
This is an implementation review by the remediation agent, not human specification
adoption. Behavioral review and execution review are recorded separately below.
No production behavior, browser scenario or evidence context is removed.

| Retired assertion | Behavior review and named survivors | Execution/evidence review |
| --- | --- | --- |
| `TestReportingProviderPackagesExposeTypedFacts`: exact filenames and `CollectFactsTx(` spelling | REQ-RPT-026e requires typed, fully bound owner contributions. The production `FieldProvider` interface enforces method/result types at compilation. `TestReportingMaterializerUsesOwnerContracts` exercises borrowed context/transaction, exact incident/support input, explicit identity, sorted output, isolated metadata, owner failure and invalid/duplicate contribution rejection without partial publication. `TestReportingMaterializerRejectsInvalidOwnerRegistration` rejects missing/duplicate owners. Existing `TestProviderOutputValidateAndBindings` and `TestProviderOutputValidateRejectsFlatOrIncompleteFacts` retain complete fact-shape validation. Harmless adapter renaming or file extraction is permitted. | New row `module.reporting.unit.source_owner_orchestration`, existing provider row, Go unit runner, no services/fixture, fast tier. No source-provider integration fixture is removed; full Reporting owner still includes them. |
| `TestReportingCurrentPathsDoNotUseLegacyCompatibilityFields`: private cache accessor spelling | Candidate REQ-RPT-026e requires structured source consumption, and REQ-RPT-039a rejects retired wire formats. `TestRedactionProfilePrecedenceActionsAndManifest_Unit` (including its structured-source and legacy-wire cases), `TestBuildExportModelUsesReportingOwnedMetadataSnapshotStableHash`, `TestSnapshotRecordMaterialization`, and `TestReportingAdmittedModelIdentityAndTime` exercise structured consumption, legacy wire rejection and immutable model identity. An accessor's name has no separate contract value. | Existing materialization/admitted-model rows retain their unit runner, fixtures and evidence. This does not establish the still-open canonical post-redaction-model cutover. |
| `TestReportingProductionDoesNotReintroduceLegacyExportOrRenderPaths`: three private function-name bans | REQ-RPT-019a requires an admitted, unchanged participant output; `TestRenderExportParticipantValidatesClosedOutputAndBounds` rejects altered/retired output and cancellation. Existing bundle/identity tests retain their actual assertions. A same-named helper cannot establish bypass, and renamed placeholder rendering passed the old scan. The real renderer remains an explicit adoption blocker. | Existing participant and render-artifact unit rows remain. No claim is made that either the old name scan or these unit fixtures demonstrates a real Slidev/Mermaid process or sandbox. |
| `TestApplicationServiceDoesNotOwnReportingJobExecutionLoop`: execution-method names in one file | Core Jobs and REQ-RPT-019a require durable execution, cancellation, terminal proof and no partial publication; they do not require a particular Go file. `TestReportingJobWorkerCancelsBeforeWork`, `TestReportingJobWorkerTimeoutRetainsFailedRelease`, `TestReportingJobWorkerRenderFailureSummaryIsDeterministic`, `TestReportingJobWorkerParticipantFailurePublishesNoRenderOutput`, and `TestReportingPreviewWorkerHonorsCancellationAndParticipantFailure` retain the observable worker assertions. Real snapshot/release replay integration remains enrolled. | Existing job-worker and participant rows retain Go unit fixtures; snapshot/release integration retains PostgreSQL/object-store contexts. Pure file extraction no longer fails a nominal architecture gate. |

The old operational row
`module.reporting.unit.reporting_production_does_not_reintroduce_legacy_36f39083f7`
is retired; its identity is not reused for changed execution semantics. Reporting's
AST owner-import guard remains active because dependency direction is a genuine
boundary. The table-reference guard also remains active: those table identities
are ownership contracts, although its limited textual detection is not complete
SQL analysis. This slice neither adds a parser nor claims to close all structural
checks across the repository.

Initial format invocation `remediation-reporting-invariants-format-20261014-01`
rejected before graph publication (`config / configuration_error`, outer exit 2).
`make target-plan-json TARGET=format` identified an unsorted authored catalog row;
sorting the row and regenerating corrected that setup defect. The failed run had
no published summary for `explain-run` to inspect. Generation then passed in
`remediation-reporting-invariants-generate-20261014-01` (9.120 s), and format in
`remediation-reporting-invariants-format-20261014-02` (4.697 s), both outer exit 0.

The first complete Reporting owner passed in
`remediation-reporting-invariants-owner-20261014-01` (22.618 s). Lint in
`remediation-reporting-invariants-lint-20261014-01` failed (22.589 s, outer exit 2,
`product / test_assertion_failure`, 10/11 units passed, all 20 cleanup steps
completed): the new test imported the Revisions source-boundary implementation
outside its admitted importer list. The constructor test now uses Reporting's
existing `SourceBoundaryResolver` interface. No owner/import-policy exemption
was added. It deliberately cannot resolve or operate on source state.

After that repair, Reporting passed in
`remediation-reporting-invariants-owner-20261014-02` (22.935 s, 8/8 units with one
compatible cache hit), and lint passed in
`remediation-reporting-invariants-lint-20261014-02` (23.978 s, 11/11 units).
Both bind source `sha256:37254cb12352ac84499ebd87a403162d803cc5071e61c8a6961ca4353ea33bfa`.
Generation drift, artifact policy and JSON shape passed in the corresponding
`remediation-reporting-invariants-{drift,policy,shape}-20261014-01` roots
(15.288 s, 0.817 s and 4.354 s). Every final owning exit was 0.
`agent-finalize` passed in both `remediation-reporting-invariants-finalize-20261014-01`
and `-02` (18.161 s and 18.062 s); both skipped retained-run maintenance because
`RESULTS_DIR` was unset.

The frontend inspector retirement keeps every existing test identity and runner
context. Only duplicate private-call/JSX/statement-placement subassertions in
the surface policy test are removed. Behavior review: the retained AST tests
reject forbidden coordinator construction through aliases and permit renamed
owner composition; the component suite exercises the panel behavior directly.
Execution review: those two exact existing rows remain
`web.workbook.boundary_support.source_workbookinspectorownershiparchitecture_test_t_8849e53c39`
and `web.workbook.regression.inspector_declared_panel_behavior`, both browser-unit,
no services, no fixture capability. No real-browser, visual, accessibility or
performance scenario was consolidated. These implementation reviews do not
substitute for independent human adoption review or the remaining structural
inventory review.

Inspector post-retirement selection passed in
`remediation-inspector-invariants-verify-20261014-01` (4.637 s); full lint passed in
`remediation-inspector-invariants-lint-20261014-01` (21.088 s). Format and finalizer
passed in the same prefix's `format-20261014-01` and `finalize-20261014-01` roots
(4.847 s and 18.048 s). All owning exits were 0; finalizer maintenance was skipped
because `RESULTS_DIR` was unset.

## Current verification and reconciliation (2026-10-13)

The first complete enrolled backend run (`remediation-backend-unit-20261013-01`) failed in ten runner groups. The frontend run (`remediation-frontend-unit-20261013-01`) failed in eight groups (446 of 454 work units passed). These are deterministic failures observed during newly complete selection, not evidence of intermittent outcomes. Backend test data fixes include microsecond Reporting timestamps, required report status, Collaboration import dependency injection, current schema inputs, canonical `Etc/UTC`, exact HTTP methods, and ordinary request-ID middleware. The SQLC guard now constrains source-owner packages and exported type boundaries rather than private filenames, and only traverses approved Go source roots.

`remediation-reporting-unit-20261013-01` passed all 20 Reporting owner rows, including six service-backed rows, after fixture repairs. The independently selected six service rows also passed in `remediation-reporting-services-20261013-01`. These results do not by themselves establish companion, recovery or adoption closure. `remediation-generated-owners-20261013-01` exposed further obsolete negative-test fixture inputs; those are being corrected without weakening production validation. The failed nested graph `native-engine/selected-graph-77163-basic-second` had two successful work units and a `runtime_close` cleanup failure; its underlying cause remains unconfirmed.

The presentation-source state-hook assertion moves to the existing inspector architecture test, retaining its invariant with import-origin checks (including aliasing) and no browser DOM dependency. The old exact-call and exact-label expectations are retired because the adopted owner requires ownership, role, route and effects, not private spelling or button copy. Registry immutability failure identified an unfrozen Import target projection; its generator now shares the existing recursive freeze emitter. Generated output is refreshed through Make.


### Follow-up verification

- `remediation-catalog-20261013-03`: `make test-slice OWNER=harness.test_catalog` passed (26.253 s), including Go ignored-file/directory discovery boundaries.
- `remediation-composition-cutover-20261013-02`: `make test-slice OWNER=module.reportcomposition` passed (22.761 s). The release-binding fixture stopped inserting the removed `rendered_output` column; assertions still prove immutable version binding.
- `remediation-reporting-portability-20261013-01`: the selected Incident Bundles integration row passed (48.826 s), exercising unavailable-pack round-trip, native snapshot references and superseded Timeline replacement.
- `remediation-workbook-corrections-20261013-04`: support and collection-focus row passed. Mock layout now represents one CSS scale consistently; semantic focus uses the shared reveal operation in the test grid adapter.
- `remediation-workbook-corrections-20261013-05`: grid-provenance row passed after waiting for the selected surface and providing its actual public schema.
- `remediation-history-contract-20261013-01`: history row passed (6.589 s), exercising the current closed rollback error codes, reviewer authorization, authority suspension, opaque history selectors and version binding.
- `remediation-target-plan-20261013-02`: extended smoke failed at JSON shape validation because selector/discovery topology fingerprints were stale. The target-plan survivor was dependency-skipped; this run provides no survivor proof.
- `remediation-finalize-20261013-03`: failed its initial JSON shape validation for the same stale topology, before refresh. `remediation-projections-20261013-02` subsequently passed `make generate` (9.940 s). Retained-run maintenance remains skipped because `RESULTS_DIR` was unset.

Failed roots are retained. These are implementation verification runs on changing sources, not qualified execution-cost observations. Full current-owner and integrated verification remains required after the remaining changes.


### Measurement implementation slice

Core 04 now defines AC-003 v2's exact TSV insertion workload. Harness REQ-806 now owns discriminated single/percentile evidence, retained-sample recomputation, retry rejection, and the narrowly scoped synthetic insertion identity proof. Current evidence schemas advance to observation v3, summary v4 and aggregate v4; the shared Timeline large-grid execution profile advances to `timeline_large_grid_snapshot_v2`. AC-043 keeps its own one-warm-up/100-observation protocol. The fixture's semantic large-grid population remains unchanged.

`remediation-paste-observer-20261013-01` passed both the new AC-003 and existing AC-043 paint qualification rows (3.743 s). `remediation-measurement-graph-20261013-01` and `remediation-measurement-evidence-20261013-01` passed graph/evidence contract checks, including statistic tampering, wrong insertion proof, retry rejection and the strict 2,000 ms boundary. These are detector tests, not product performance qualification. Full browser measurement remains pending.

`remediation-target-plan-20261013-03` reached the replacement target-plan test and exposed an overbroad added assertion: command-owned UI review integration checks correctly reserve PostgreSQL capacity while acquiring their fixtures through the command boundary. The assertion now checks enclosing capacity claims and retained PostgreSQL fixture work; it no longer assumes every service consumer uses a graph fixture capability. Fresh survivor execution subsequently passed in `remediation-target-plan-survivor-20261013-05`; the wrapper is retired.

The [Reporting clause review](reporting-cutover-adoption-review.md) enumerates changed clauses against the immutable baseline and explicitly separates existing passing evidence from unresolved promotion conditions. No candidate is promoted yet.

### Publication and discovery follow-up

`remediation-ac003-browser-20261013-01` failed before the browser action. A concurrent frontend producer published unignored `.frontend-publish-*` files while another producer hashed source inputs; those transient files disappeared during hashing. The temporary publication tree is now Git-ignored, and `remediation-publication-fix-20261013-01` passed `make harness-graph-contract`, including unchanged source-digest assertions for both staging and displaced output. This is an attributed harness boundary defect; it is not a measured product failure.

Source discovery prunes measurement output and publication staging before stat/read. Reverse Go completeness ignores declared generated outputs while explicit existing catalog selections still resolve generated schema parity tests. Unsupported Unicode Go entrypoint names now produce an explicit admission error instead of disappearing from completeness accounting, and the valid bare `Test` name is recognized. `remediation-discovery-projections-20261013-01` rejected an overbroad draft exclusion that also blocked the explicitly enrolled generated parity test; the separation above repaired it. `remediation-discovery-projections-20261013-02` then passed generation.

`remediation-catalog-20261013-04` passed the full catalog owner (25.554 s); `remediation-finalize-20261013-06` passed finalization (26.161 s). Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

### Recordless paste insertion

`remediation-ac003-browser-20261013-02` completed both frontend builds, populated fixture preparation and 25-session qualification, then failed because no paste acceptance mark was produced. It retained an incomplete single-observation attachment and completed cleanup; no duration satisfies AC-003. Source review showed that the measurement dispatched to a native blank-row editor, whose required native paste behavior deliberately excludes table transport. The production grid also lacked a distinct recordless insertion target.

Core 03 REQ-03-147 now explicitly owns that target: Escape in an empty blank-row scalar editor selects the insertion cell; Enter returns to the editor. The adapter keeps the target outside committed record/range identity and sequential tab order. A shared production/test binding handler decodes table input, resolves current writable columns and creation capability, and invokes a typed draft-paste intent. Timeline revalidates current targets and authority before using its existing clipboard batch port. Core 03 native editor paste remains intact. The Design owner defines the visible focus ring and action announcement. AC-003 preparation now reaches this target through the user keyboard path before timing.

The first two new interaction runs (`remediation-draft-insertion-20261014-01` and `-02`) exposed that vendor summary cells were not focusable and the vendor did not forward capture handlers. The final handlers live on the adapter-owned event boundary. `remediation-draft-insertion-20261014-03` passed the new exact row. `remediation-grid-adapter-20261014-01` passed the full adapter owner (40.357 s); `remediation-native-draft-20261014-01` passed both native scalar draft cases (6.872 s); `remediation-paste-translation-20261014-01` passed the existing sorted/filtered/grouped clipboard translation rows (7.149 s). `remediation-draft-paste-types-20261014-02` passed frontend type checking. These functional results do not replace the pending current browser measurement.

`remediation-finalize-20261014-07` passed finalization (23.674 s); retained-run maintenance was skipped because `RESULTS_DIR` was unset. `remediation-ac003-browser-20261014-03` then accepted the single paste but failed visible completion with zero qualified rows (162.112 s total). Its classification is `product` / `test_assertion_failure`; its observation is incomplete, and all cleanup activities completed. The retained browser state still displayed the original query window. The receipt application path appends creates locally, but its query refresh has no batch creation presentation pin and can remove them. No AC-003 performance pass is established.

Core 03 REQ-03-224 now extends the existing bounded creation pin to the latest creation operation, including paste-created rows in receipt order. This explicit temporary presentation exception preserves canonical query identity, freshest row versions, unique identities, and scope/authority boundaries. The clipboard adapter provides a cancellable, one-time receipt observation after source settlement. Timeline captures the originating query scope and creation generation, retains only the latest creation block, and reveals its end only if the user's focus and scroll have not changed. Measurement still dispatches once and never scrolls through results. Implementation and focused regression validation are in progress.

### Pre-enumeration documentation boundary

Both source-snapshot Git enumeration and executable-policy scanning now pass the shared documentation exclusions to Git before it inspects untracked directories. Post-enumeration filters remain defensive, and file existence checks follow admission. `remediation-source-boundary-20261014-01` passed the full catalog owner (23.281 s), including pre-enumeration argument checks and forbidden filesystem probes for synthetic documentation fixtures. This fixes an implementation boundary; no executable consumer reads human adoption records.

### Creation-block and target-evidence validation

`remediation-batch-presentation-20261014-01` passed the Timeline row coordinator (3.934 s), including bounded receipt-order pins, deduplication, freshest query versions, stale generation rejection, query changes and detached presentation. `remediation-paste-receipt-20261014-01` passed the clipboard adapter (3.408 s), including exactly-once observation after settlement, cancellation, rejected receipts and no resubmission. `remediation-batch-presentation-types-20261014-01` passed frontend type checking (20.171 s).

`remediation-finalize-20261014-08` failed source-policy validation because the negative fixture used literal restricted-documentation paths. `remediation-shape-diagnosis-20261014-01` retained the exact diagnosis. The test now supplies a synthetic `human-material` restricted root in its temporary repository; the production policy is unchanged. Generation passed, then `remediation-finalize-20261014-09` passed (23.500 s). Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

`remediation-ac003-browser-20261014-04` produced one successful browser observation: 394.59999990463257 ms, 20 new record/version bindings and all 100 visible cells, with qualified geometry/traffic, zero scheduler overlap and completed cleanup. Its row and profile summary passed. The enclosing target failed because `browser_target_result.v4` still admitted only a v3 aggregate artifact kind, despite the new v4 measurement aggregate. The finalizer's unhandled exception was incorrectly classified as a product failure. This is an attributed artifact-contract defect; the enclosing run remains failed and cannot be presented as fully accepted evidence. Source digest: `sha256:69893f99a47d8a68126e5727a6b3b8175c30c53be0d8f24ae45db9d96284e275`.

Browser target evidence now advances to `cartulary.browser_target_result.v5`, with current aggregate kinds and no retired-reader fallback. Visual readers and fixtures move with it; the unchanged visual reconciliation identity remains v3. Target-finalizer exceptions use the canonical artifact failure boundary. Contract regression tests and fresh integrated qualification are pending; retained failed evidence is not rewritten.

The subsequent exact AC-003 slice `remediation-ac003-browser-20261014-05`
passed end to end (150.744 s): one fresh observation **381.5 ms**, strict
threshold 2,000 ms, qualified measurement aggregate, v5 target result, no
scheduler overlap, and successful cleanup. Its source digest is
`sha256:45644954844d983a2998f04a1007647cfd83f391a2905e186ffb43a7be023ec6`.
This is retained single-observation conformance evidence for that source, not
a Core 05 publication claim. The v5 target schema regression suite
`remediation-browser-target-contract-20261014-01` passed (21.880 s), followed by
`remediation-finalize-20261014-10` (23.513 s). Retained-run maintenance was skipped
because `RESULTS_DIR` was unset. Later executable-source changes require an
explicit applicability review or fresh qualification.

### Static registration and executable-input boundaries

The finalizer now snapshots and restores admitted tracked executable inputs through
one restricted Git enumeration, preserving file permissions and symlink identity
without following link targets. Markdown and restricted roots are excluded before
Git enumeration. Shell and Go static-analysis discovery use the same policy and
propagate discovery failures instead of treating failed discovery as an empty pass.
`remediation-finalizer-boundary-20261014-01` passed `harness-evidence-contract`
(8.040 s), including forbidden filesystem probes and rollback preservation.

Frontend discovery now resolves lexical constant bindings and named runner aliases,
rejects conditional/indirect registrations and unsupported modifiers/formatters,
and requires imported registration suites to have one unconditional call to the
actual named export. Three profile-conditional browser registrations were made
unconditional; existing catalog rows retain their exact runtime profiles and titles.
`remediation-parser-contract-20261014-02` passed the graph contract suite. The pure
formatter is an explicit pinned workspace dependency (`@vitest/utils` 4.1.4),
shared by ordinary and scratch generation. Catalog command failure publication now
lives in execution composition, preserving the pure catalog import boundary.

`remediation-discovery-finalize-20261014-01` failed JSON-shape validation because
failure publication was initially imported inside the catalog owner; the CLI was
moved to execution composition. `remediation-discovery-finalize-20261014-02` and
its narrow diagnostic `remediation-discovery-drift-20261014-01` failed scratch
generation because the formatter was reachable only through the web package.
The direct dependency addresses that prerequisite; these failed roots remain
retained and do not supply passing evidence. Further finalization is pending.

`remediation-discovery-finalize-20261014-04` passed finalization (23.290 s) after
regeneration; retained-run maintenance was skipped because `RESULTS_DIR` was
unset. The prior `-03` attempt rejected stale topology inputs, as confirmed by
`remediation-discovery-shape-20261014-02`; `remediation-discovery-projections-20261014-03`
regenerated them. `remediation-diagnostic-evidence-20261014-01` passed the evidence
contract suite (8.000 s), including error-headline selection ahead of Node source
excerpts. In `remediation-discovery-smoke-20261014-01`, the lint-shell check passed
(8.040 s) with a restricted-root script and non-traversable README/Markdown links.
Full smoke completion and the fast-tier result remain pending.

### Complete fast-tier reconciliation

`remediation-discovery-smoke-20261014-01` completed successfully, including the
native service boundary (155.150 s). The subsequent `remediation-complete-fast-20261014-01`
completed all 565 work units in 358.548 s with three assertion failures; all 574
cleanup operations completed. The failures were the public View Schema registry
omitting the adopted `note_associations`/`note_associations_route` vocabulary,
a layout test retaining the old 45vw-only expectation instead of the required
720px grid remainder, and a Timeline refresh test expecting immediate server order
instead of REQ-03-224's temporary local creation block. Expectations are being
reconciled against those owners, without removing their assertions.

Review also found that a rejected second paste gesture could reserve a presentation
generation and suppress an earlier admitted receipt. Presentation is now reserved
only after admission succeeds; a focused hook regression covers that ordering and
unmount cancellation. Its input port now requires only the pending-save reference
it actually consumes. Narrow verification remains pending.

All four focused repairs passed: `remediation-paste-admission-20261014-01`
(2.785 s), `remediation-viewschema-vocabulary-20261014-01` (1.820 s),
`remediation-layout-metrics-20261014-01`, and
`remediation-query-creation-order-20261014-01`. The first type check exposed
composition deriving its broader pending-save type from the newly narrowed
clipboard port. Composition now declares its actual model type;
`remediation-paste-admission-types-20261014-02` passed (20.482 s).

The executable-input boundary now rejects a tracked descendant whose ancestor
has become a symlink before statting or opening the descendant. Static consumers,
source snapshots and finalization share that admission check. The catalog also
checks path ancestry before existence checks on selected files and imported suites.
`remediation-boundary-contract-20261014-01` passed (23.010 s), including
non-traversal assertions. Additional parser cases reject computed API calls,
destructured aliases, API storage in objects and API passage to arbitrary helpers;
`remediation-registration-contract-20261014-01` passed. During implementation,
`remediation-registration-escape-20261014-01` through `-03` exposed an overbroad
detector that confused ordinary property names and runtime metadata with runner
bindings; `-04` passed after separating those identities. These failed attempts
remain investigation evidence, not product failures or timing samples.

`remediation-registration-finalize-20261014-01` passed (24.287 s). Finalization
retains bounded redacted substep logs and keeps a cleanup failure secondary to
an earlier child failure. `remediation-finalizer-cleanup-precedence-20261014-01`
passed its regression suite. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset. The subsequent complete fast run and AC-003/AC-043 measurement results are recorded below.

The build-context review found that Go discovery and invocation needed one shared Linux/amd64 profile. The correction and its adversarial validation are recorded below; conditional source admission is no longer based on an independently assumed host environment.

The complete fast tier `remediation-complete-fast-20261014-02` passed all 566 work units in 359.285 s, outer exit 0; all 575 cleanup operations completed. Its source digest matches the five-case Timeline measurement window recorded in [the timed assurance review](timed-product-assurance-review.md). These are execution results for the recorded source and selection, not full remediation/adoption or cost-publication closure.

The Go build-context follow-up now shares one operational profile between selector discovery and adapter execution. It binds Linux/amd64.v1, gc, cgo disabled, the pinned toolchain, no custom flags or experiments, and disables user Go environment/workspace selection. Explicit experimental constraints fail configuration rather than being silently interpreted against incomplete tags. Focused validation follows in the retained records below.

Go-context validation: `remediation-go-context-graph-20261014-01` passed `harness-graph-contract` in 24.390 s. In `remediation-go-context-runner-20261014-01`, the Go adapter checks passed but the containing command failed because its Vitest fixture was under an excluded transient root. The fixture was moved to explicit authored support data without relaxing selector containment. `remediation-go-context-runner-20261014-02` then failed a frontend-artifact source digest check while this remediation was editing executable sources. It is an incompatible observation, not evidence of an intermittent product failure. The subsequent frozen-source rerun is recorded below.

The frozen-source `remediation-go-context-runner-20261014-03` passed `harness-command-surface-contract` in 58.930 s, outer exit 0. `remediation-panel-survivors-20261014-02` passed the new panel behavior and existing ownership checks in 3.871 s; `remediation-panel-survivor-types-20261014-01` passed type checking in 21.462 s. Finalization initially rejected a catalog-to-execution import (`remediation-context-panel-finalize-20261014-01`); the shared Go build context now belongs to the catalog owner and is consumed through its public facade. Boundary enforcement was preserved.

After the Go profile change, `remediation-go-context-backend-20261014-01` passed the complete `backend-unit` selection: 199 work units, 130.736 s, outer exit 0, and 212 completed cleanup operations. `remediation-context-panel-finalize-20261014-03` passed in 24.566 s with retained-run maintenance skipped because `RESULTS_DIR` was unset. Markdown lint passed in `remediation-handoff-markdown-20261014-01`. The first frontend boundary run identified a support-fixture naming issue and an overbroad production-controller rule that included its own tests. The fixture now uses the existing `TestSupport.ts` classification; the controller rule excludes ordinary test files while preserving its production dependency prohibition.

The complete backend integration run `remediation-complete-integration-20261014-01` finished with 297 passed and four failed catalog rows (112 passed / four failed work units); all 341 cleanup operations completed. Outer Make exit was 2; retained classification is `product/test_assertion_failure`. Source was `sha256:d86543fb61af0a2e18b7e0c7625c727bf9832ff5dbb0f6950b661a7277ef2712`, graph `sha256:ef8cd531ca7d7aeb5d38fac41e2848c8c2ff9e9348d9fc73c091bc6159582779`, duration 477.407 s. Passing rows retain their actual scope; this run does not establish integrated acceptance.

The four failures have concrete repair dispositions:

- `platform.viewschema.integration.view_schemas_discovery_http_036651fc36`: update obsolete assertions to Core 01's closed inspector token set and `cursor_query_mismatch` for a changed collection-route contract. Production validation remains unchanged.
- `module.extensions.static.behavior_routing_e80c9e3dc7`: retain every named owner-boundary assertion, retire the copied total-row formula. Bidirectional catalog reconciliation owns completeness; additional justified owner tests must not fail a second inventory.
- `module.extensions.integration.extension_coordination_state_and_staged_objects_5a1db013c5`: the store test must not infer a profile claim from bootstrap migrations. Assert absent metadata before explicit fixture initialization, then verify the inserted metadata and existing staged-object transitions. Extensions §21 owns initialization; production state admission remains unchanged.
- `module.auth.integration.route_idempotency_actor_scoped_uniqueness_afb930c9c4`: use `postgres_dedicated`, because the assertion commits independent transactions to test database uniqueness. Remove its erroneous transaction-fixture approval. The first generation attempt exposed that stale approval (`remediation-integration-repairs-generate-20261014-01`); corrected generation passed in `-02`.

Measurement evidence now uses descriptor-relative, no-follow reads for each path component, rejects documentation tokens before filesystem access, and rejects aliased/nonregular/oversized or changing files. `remediation-artifact-reader-graph-20261014-01` passed in 23.200 s, including an ancestor-rename adversarial case. The first finalization flagged hostile documentation path literals in the negative tests as executable references. They are now represented explicitly as rejected input tokens while a filesystem probe stub proves no traversal; the production boundary remains intact. Follow-up finalization and repaired integration results are recorded below.

Focused repair evidence: `remediation-integration-repairs-extensions-20261014-01` passed both selected rows in 21.282 s with 17 cleanup operations complete; `remediation-integration-repairs-auth-20261014-01` passed in 19.518 s with 14 complete. Both used source `sha256:0717405ee779ffc5f8cd0f38f9c5fa7fa4f0c3b6b8c572180a5d9108eecc74bb`. The first schema rerun exposed two further obsolete expectations: public `create_writable` was incorrectly treated as a leak, and the closed route-kind table omitted Indicator observations/lifecycle. The test now requires a boolean create capability and the adopted route vocabulary while continuing to reject internal write/storage metadata. `remediation-integration-repairs-viewschema-20261014-02` passed in 14.495 s with 14 cleanup operations complete at source `sha256:a76003e360c36ca18ede67854715861318d02ccb563310bf7a9ecb5dd5793e88`.

`remediation-artifact-reader-command-20261014-01` passed the command-surface contract in 58.600 s. `remediation-integration-repairs-finalize-20261014-02` rejected stale generated topology after a finalizer import correction; regenerated inputs passed finalization in `-03` (24.389 s). Retained-run maintenance was skipped because `RESULTS_DIR` was unset. All originally enrolled Go inventory rows now have an explicit fresh selected-row pass in the reconciliation ledger; this establishes execution reachability for those sources, not full behavioral assurance or a complete integrated pass.

## Stability incident dispositions

All roots below are retained under `.cartulary/test-results/`; exact source, graph, toolchain, environment and selection identities belong to each root's manifest and selected-row results. Missing historical bindings are unknown, not reconstructed. No cached result is counted as a fresh pass, no retry-to-green policy was introduced, and a successful repair run is not treated as proof of stability.

| Incident / evidence roots | Symptom and cause disposition | Comparability / denominator | Repair and closure |
| --- | --- | --- | --- |
| Initial complete backend/frontend selection, `remediation-backend-unit-20261013-01` and `remediation-frontend-unit-20261013-01` | Observed failures in newly reached tests. Reviewed causes include obsolete fixture fields, invalid admission timestamps, missing ordinary middleware and stale structural expectations. These are test/fixture defects where separately demonstrated, not a generic flakiness finding. | Different source after repairs; no intermittent-outcome rate is inferable. | Exact successor rows and source identities are in the inventory and preceding evidence records. Unrelated uncovered obligations remain open. |
| Nested `native-engine/selected-graph-77163-basic-second` in the first backend run | Both work units succeeded; `runtime_close` reported a cleanup failure. Underlying mechanism remains unconfirmed. | One retained failure; no source-compatible pass/fail pair established. Later successful cleanup on changed sources is not a reproduction. | Suspected cleanup instability, not an attributed product/test defect. Keep the original cleanup evidence; investigate the same lifecycle context if reproduced. No disabling or authorized skip is justified. |
| `remediation-go-context-runner-20261014-01` | Vitest diagnostic support data was placed under an excluded source root. Attributed harness-fixture error. | Deterministic configuration failure; Go adapter assertions had passed. | Authored support fixture uses the existing `TestSupport.ts` classification. `remediation-artifact-reader-command-20261014-01` passed after the final filename/boundary correction. |
| `remediation-go-context-runner-20261014-02` | Artifact source identity changed during a test invocation. Attributed concurrent-edit contamination. | Incompatible source window; excluded from intermittent comparisons and performance sampling. | Source frozen for `-03`, which passed. Freeze is an execution precondition, not an automatic test retry policy. |
| `remediation-complete-integration-20261014-01` | Four assertion failures; no cleanup failures. Causes are the documented stale schema/initialization/count assertions and incorrectly selected transaction fixture. | 301 selected rows: 297 passed, four failed; 341/341 cleanup operations completed. This is an invocation outcome, not a failure-rate estimate across repeated compatible observations. | Mechanism-preserving assertion/fixture corrections and fresh focused passes are recorded above. The failed root remains retained. |
| Earlier externally supplied historical stability allegations | Suspected instability only where exact source, selection, fixture, policy and fresh-execution bindings are absent. | Unknown denominator; no Cartulary-specific intermittent rate can be claimed. | Explicitly unconfirmed. Required bounded surface-selection synchronization remains; measured product actions are not retried. |

The first integrated `check`, `remediation-selection-integrated-check-20261014-01`, completed at source `sha256:a76003e360c36ca18ede67854715861318d02ccb563310bf7a9ecb5dd5793e88` and graph `sha256:65553953ab39eb6ace1185bfcd683b5154f18a7a8ac2eb419404c76bc52cf46d`. Of 768 work units, 765 passed and three failed; 1,521 catalog rows passed and two failed. All 1,043 cleanup operations completed. Cache accounting was 171 bypass / 597 miss / zero hit. Duration was 556.815 s; outer Make exit was 2, with retained primary `product/test_assertion_failure`.

The failures were introduced test/routing maintenance omissions, not intermittent outcomes: the inspector label assertion used a named heading as a presence selector; three new TypeScript files lacked source-owner manifest entries; and the new panel characterization used three non-null assertions rejected by Biome. The corrections preserve the production boundaries: select the panel by its stable builder before asserting heading content, assign the new harness support fixture to `web.testing` and the two Workbook tests to `web.workbook`, and make the panel-fixture precondition explicit. This failed integrated root remains evidence; focused repairs and a later passing integrated gate are required before claiming integrated closure.

The UI/UX refactor skill's maintained overlay was consulted for the Workbook assurance review. It remains advisory. Its acceptance criteria are not yet all demonstrated: the cross-mechanism obligation matrix, relevant visual/accessibility evidence, complete timed predicates and reviewed consolidation survivors remain open. No browser matrix has been retired by inference from file size or the integrated row count.

The integrated-policy corrections passed their focused checks: `remediation-integrated-policy-architecture-20261014-01` (4.860 s), `remediation-integrated-policy-panels-20261014-01` (3.692 s), and `remediation-integrated-policy-biome-20261014-01` (2.045 s), each outer exit 0. Generated projections passed in `remediation-integrated-policy-generate-20261014-01` (10.250 s); finalization passed in `remediation-integrated-policy-finalize-20261014-01` (23.975 s). Retained-run maintenance was skipped because `RESULTS_DIR` was unset. The complete repaired `check`, `remediation-selection-integrated-check-20261014-02`, passed in 564.668 s with outer exit 0. Source was `sha256:03f179b688258f0b827469d9448d4a8dae0c67c751e4156585cca7abf8c55d22`; graph remained `sha256:65553953ab39eb6ace1185bfcd683b5154f18a7a8ac2eb419404c76bc52cf46d`. All 768 work units and 1,523 catalog rows passed; all 843 cleanup operations completed. Cache accounting was 171 bypass / 200 hit / 397 miss. Cached rows are reused valid execution evidence, not fresh successful observations for stability rates. The root has not yet received an owner evidence audit and is not a clean-source performance window.

The retained evidence reader and runtime private-file operations now share one descriptor-relative directory traversal. Performance manifest/summary readers and the canonical event stream reject documentation paths before filesystem access, reject symlink ancestors and aliased files, and detect mutation while reading. Prefix-based live-event observation keeps its separate append-only inode authority. `remediation-shared-evidence-contract-20261014-01` passed in 8.150 s, including canonical event and performance adversarial cases; `remediation-shared-evidence-graph-20261014-01` passed in 23.160 s. Generation passed in `remediation-shared-evidence-generate-20261014-01` (10.230 s). These changes occurred after the full check and require finalization and affected verification. The harness evidence-generation clause now names observation v3 and summary/aggregate v4, matching the approved single-observation extension; old evidence is not translated.

The shared-reader follow-up passed `harness-command-surface-contract` in
`remediation-shared-evidence-command-20261014-01` (58.420 s), finalization in
`remediation-shared-evidence-finalize-20261014-01` (24.263 s), and Markdown lint
in `remediation-evidence-docs-20261014-01` (9.200 s), each outer exit 0.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

Fresh execution of previously unselected browser rows exposed three obsolete
expectations. `remediation-enrolled-workbook-browser-20261014-01` failed in
37.786 s (all 26 cleanup operations completed); its viewer-denial scenario
passed, while the inspector geometry and editable-reference scenarios failed.
`remediation-enrolled-history-browser-20261014-01` failed in 35.094 s (all 22
cleanup operations completed). Both outer Make exits were 2, with retained
`product/test_assertion_failure`, at source
`sha256:4d4f4f96286a4a40f229e554013188ea70f80e577ef4a613a7458681142f4842`.
These are attributed expectation defects, not confirmed intermittent outcomes.

The history test now checks the typed technical record identity and the actual
retargeted history content through a committed-cell gesture, rather than retired
`Record <id>` prose. The geometry test accounts for the Core 03/Design-owned
Workbook browsing controls within the primary grid slot; it still requires the
inspector and complete grid work area to fill the shell and scroll independently.
The reference test compares the accepted collection item's `display_text` and
stable linked identity rather than assuming a referenced Evidence title is its
wire label. Long accepted text remains separately exercised by
`GenericInspectorReferenceSummary.test.tsx`; no cross-record title lookup was
added to presentation. Fresh repaired browser results remain required.

The repaired history selection passed in
`remediation-enrolled-history-browser-20261014-02` (29.758 s, 22 completed cleanup
operations). The geometry scenario passed in
`remediation-enrolled-workbook-browser-20261014-02`; that enclosing run remained
failed (34.699 s, all 26 cleanup operations completed) because the editable
reference scenario subsequently exposed a globally scoped `Manage Evidence`
selector matching three panels. Both runs used source
`sha256:c2ce4bca65c9c034f25e9a8b0985b9c5d724a150bcbe02a56e33d1f03bc12da9`.
The helper now accepts the already scoped Evidence-panel action, and its zoomed
assertions use the same accepted display text. The complete two-scenario
accessibility row passed in `remediation-enrolled-reference-browser-20261014-03`
(28.626 s, outer exit 0). All previously unselected frontend entries now have
explicit retained execution, with unit/component cache applicability and browser
fresh execution distinguished in the reconciliation ledger. This does not fill
missing timed or publication evidence.

The final accessibility source was `sha256:11119af25cd655283e7cecbf9d999db478db41347b10bdf7e247e9d3579b7342`, graph `sha256:2bc42e703adc307f972721595dfae5179480ec22c18a61a174fdbe43ca56dfd2`; all 22 cleanup operations completed. Type checking passed in `remediation-enrolled-browser-types-20261014-01` (23.203 s). The intermediate failed roots remain retained.

## Integrated source enrollment and audit closure

`remediation-source-enrollment-check-20261014-01` passed in 543.505 s, outer exit
0, at source `sha256:11119af25cd655283e7cecbf9d999db478db41347b10bdf7e247e9d3579b7342`
and graph `sha256:65553953ab39eb6ace1185bfcd683b5154f18a7a8ac2eb419404c76bc52cf46d`.
All 768 work units and 1,523 catalog rows passed; all 1,043 cleanup operations
completed. Cache accounting was 171 bypass / zero hit / 597 miss. This closes
integrated execution for that source and the separately retained newly enrolled
browser rows. It remains dirty-source implementation evidence, not a qualified
cost window or Reporting adoption evidence.

The shared retained-artifact boundary now also protects canonical audit manifests,
referenced-artifact presence checks, and private local evidence readers. Tests
reject documentation before filesystem access, directory/leaf symlinks, hardlink
aliases and changed bytes. Presence checks do not reread large event streams.
`remediation-audit-boundary-contract-20261014-01` passed in 8.110 s. Reviewed roster
contracts passed in `remediation-reviewed-roster-evidence-20261014-01` (8.150 s)
and `remediation-reviewed-roster-graph-20261014-01` (23.640 s).

The first reviewed-roster generation rejected the plural section token in the two
new dispositions. Both now use the closed `Section 4` token; generation passed in
`remediation-reviewed-roster-generate-20261014-02` (10.290 s). The first finalizer
rejected literal restricted paths in a hostile-input test. Those cases now build
synthetic path tokens without referencing repository documents; the zero-access
assertion remains. `remediation-reviewed-roster-finalize-20261014-02` passed in
24.204 s. Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

Review found that the old owner auditor counted referenced row filenames without
validating successful outcomes or current compatibility. It now reconstructs the
current graph from retained selection/capacity, compares unit semantic identities
and row registrations, checks successful unit/row outcomes and completed cleanup,
and publishes a closed `cartulary.harness_owner_evidence_audit.v1` result. Required,
accepted, rejected and unused partitions remain explicit. Cached child results
remain reused evidence. Source changes during the audit reject acceptance.

Harness REQ-121 and REQ-668 previously disagreed about requiring both a full-owner
slice and leaf partitions. The amendment preserves REQ-121's atomic full-owner
alternative, rejects filtered slices as full-owner closure, and keeps independent
project security/release/publication gates additional. The current row schema is
v3; copied v2 normative names were corrected without adding historical readers.
`remediation-owner-audit-generate-20261014-01` passed in 11.210 s;
`remediation-owner-audit-contract-20261014-01` passed in 8.320 s. Additional negative
cases cover missing cleanup and omitted unit evidence. Finalization passed in
`remediation-owner-audit-finalize-20261014-01` (25.141 s); retained-run maintenance
was skipped because `RESULTS_DIR` was unset. Fresh public-command audit validation
is in progress; this record does not yet claim it passed.

The first public audit (`remediation-owner-audit-public-20261014-01`) rejected
before execution because graph preparation discarded required argv inputs. The
graph compiler now receives declared inputs explicitly; Node owner commands use
the existing dispatcher with literal environment transport. Inputs enter the
unit identity; undeclared values are excluded. The first generation attempt
exceeded the existing Make binding budget by 256 bytes. Sharing the literal argv
transport macro removed that duplication without raising the budget. Generation
passed in `remediation-owner-audit-dispatch-generate-20261014-02` (10.270 s), and
finalization passed in `remediation-owner-audit-dispatch-finalize-20261014-03`
(24.407 s). Earlier finalizer attempts rejected stale generated projections.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

Command-surface and graph contracts passed in
`remediation-owner-audit-dispatch-command-20261014-01` (58.810 s) and
`remediation-owner-audit-dispatch-graph-20261014-01` (24.960 s). Their deliberately
failed child fixtures are negative checks, not failures of those contract runs.
The second public audit (`remediation-owner-audit-public-20261014-02`) retained
both inputs but rejected artifact identity initialization because the CLI no
longer had their Make origin after argv conversion. The CLI now identifies its
two parsed arguments as explicit CLI inputs. An automated dispatcher-to-CLI
regression test checks this boundary; successful public validation is still
required. The small `platform.pagination` provider slices passed in the first
and second attempts (811 ms and 728 ms); they are not audit acceptance.

The dispatcher-to-CLI negative fixture initially inherited the outer contract
run's prepared identity; `remediation-owner-audit-argv-contract-20261014-01`
failed for that fixture defect (9.660 s). The fixture now creates an independent
identity. `remediation-owner-audit-registration-contract-20261014-01` passed
(9.840 s), including tampered kind, dependency, resource and service registration
rejection. Generation and finalization passed in the corresponding
`remediation-owner-audit-registration-*-20261014-01` roots (10.350 s and 24.446 s);
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

The fresh `platform.pagination` slice passed in
`remediation-owner-audit-pagination-20261014-04` (580 ms), followed by successful
public audit `remediation-owner-audit-public-20261014-04` (2.453 s, both outer
exits 0). The typed report accepts both exact owner rows and their current graph,
unit/row outcomes, declared artifacts and cleanup. Source was
`sha256:e8c9eb333a217fa5a0131f70b4691f99bd5e3db78b9010c492d8a8bbb1b07d21`;
catalog was `sha256:113f38c046f5ec7749c2d3fd61b2ac7c6935b3d44a64ffbaf8fe7464b9aa2c98`.
The preceding public attempt `remediation-owner-audit-public-20261014-03` also
passed (2.422 s) before registration validation was tightened. These are public
audit implementation checks, not project-wide assurance or qualified cost
windows. Subsequent executable-source changes require compatible new evidence.

## Renderer-source verification and standalone runtime readiness

`remediation-audit-svg-integrated-check-20261014-01` finished in 559.799 s with
767 passing units and one failed Go-format unit (outer exit 2). All 1,043 cleanup
operations completed. Its source was
`sha256:e780b98723148dd9448d45b98b736eb43b1cf9f8b832156d05282634ba4c90a1`.
Formatting was corrected through Make; the aggregate remains failed evidence.
The Reporting owner slice subsequently passed in
`remediation-reporting-source-owner-20261014-01` (28.407 s, outer exit 0), covering
the closed SVG validator and corrected Mermaid serialization. Real browser
rendering and its sandbox evidence remain adoption blockers.

Standalone `make lint-go` first failed in
`remediation-reporting-source-lint-20261014-01`: the Go-format child exited 127
because constrained discovery invoked ambient `node`. Graph execution had
supplied Node and masked the missing standalone dependency. The authored task
recipes now declare pinned Node readiness and pass `NODE_BIN` for Go formatting,
shell discovery and the targeted Go security scanner. Shell discovery has a
regression fixture with a failing ambient Node and a valid explicitly selected
runtime. Generation passed in `remediation-static-node-generate-20261014-02`
(10.620 s), standalone Go lint passed in
`remediation-static-node-lint-20261014-01` (outer exit 0), and finalization passed
in `remediation-static-node-finalize-20261014-01` (24.507 s). Retained-run
maintenance was skipped because `RESULTS_DIR` was unset. An attempted direct
`harness-smoke-lint-shell` invocation was rejected because that is a private
check identity, not a public Make target; it produced no test evidence. The
public command-surface contract owns its execution.

`remediation-static-node-command-contract-20261014-01` passed (58.500 s, outer
exit 0), including the declared-runtime shell regression. Markdown lint passed
in `remediation-static-node-markdown-20261014-01` (9.180 s), before this annotation.

Composition source admission now has one owner facade for immutable versions
and preview descriptors. Consumers validate closed canonical bytes, nested
authoring input, recomputed digests and retained identity bindings before
projecting fields for rendering. Version previews also reconstruct and verify
the bound immutable-version preimage. A service test alters retained version
bytes while preserving the digest column and confirms rejection.

Reporting's closed selection-rule type is shared with Composition through the
Reporting-owned `diagramcontract` leaf package. Draft admission and immutable
creation materialize declared defaults and NFC-normalize authored text bodies;
retained documents that need repair reject. Opaque source refs and semantic
anchor identifiers no longer fail merely because they contain renderer names or
prefixes. Actual target membership still requires Reporting resolution; source
shape validation does not establish complete overlay/render behavior.

The complete Composition slice passed in
`remediation-composition-admission-owner-20261014-02` (20.000 s), and the affected
Reporting source row passed in `remediation-composition-admission-reporting-20261014-02`
(6.317 s), both outer exit 0. Earlier corresponding runs also passed (19.689 s
and 6.233 s) before shared selection/default validation was added. Formatting
passed in `remediation-composition-admission-format-20261014-03` (9.238 s) and
generation in `remediation-composition-admission-generate-20261014-02` (10.290 s).

The full Reporting owner passed in
`remediation-composition-admission-full-reporting-20261014-01` (29.228 s), and
Go lint, module-boundary validation and Markdown lint passed in the matching
`lint`, `boundary` and `markdown` roots (outer exits 0). Finalization passed in
`remediation-composition-admission-finalize-20261014-01` (24.450 s);
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

## Shared fixture and view-change measurement work

The unchanged large-grid fixture now has its own typed source contract,
`contracts/performance/large-grid.v1.json`. AC-043's projection is v2 and binds
that owner; its individual predicate identities, sampling and thresholds are
unchanged. AC-003 binds the same fixture. Fixture contribution identities and
semantic counts are preserved; regenerated source bindings invalidate older
snapshot keys instead of accepting a legacy projection.

The first generation attempt (`remediation-shared-large-grid-generate-20261014-01`)
rejected a contribution still bound to the retired AC-043 contract. The first
focused unit run (`remediation-shared-large-grid-unit-20261014-01`) then failed
with `infra` / `preflight_error`: the public UI-contracts barrel did not export
the new fixture constant. This was an introduced load defect, not external
infrastructure instability. All 13 cleanup activities completed. Both inputs
were corrected. Generation `-03` passed (10.360 s), formatting `-02` passed
(9.461 s), and the four fixture/paint rows passed in unit run `-02` (4.009 s,
outer exit 0).

Published summaries for `remediation-shared-large-grid-finalize-20261014-01`,
`remediation-shared-large-grid-types-20261014-01` and
`remediation-shared-large-grid-graph-20261014-01` report passes in 24.230 s,
20.228 s and 24.490 s respectively. Their outer session exits were not retained
across context recovery, so these summary results are not represented as
independently observed successful command exits. Finalization did not receive
`RESULTS_DIR`; retained-run maintenance was skipped.

AC-044 now specifies independent Timeline sort, filter and group workloads,
each with first-useful and stable completion identities. The former two generic
view-change identities are retired rather than reinterpreted. The existing
fixture exercises one expanded `disabled` activity-time group; this does not
establish multiple-group behavior or performance on other surfaces. Product
acceptance marks bind a subsequent committed query revision, and the browser
observer requires a filled visible row window, valid bindings, keyboard entry
and two consecutive qualifying frames. Stable completion additionally checks
final order and settled query state. Execution qualification is still pending.

`remediation-view-change-generate-20261014-01` failed because appended catalog
rows were not sorted. The catalog was corrected; generation `-02` passed in
10.380 s (outer exit 0). Task-guide calls made against the unsorted catalog also
rejected before selection; they produced no execution evidence.

The AC-044 adversarial paint row passed in
`remediation-view-change-paint-20261014-01` (2.834 s); the four affected Workbook
query/controller rows passed in `remediation-view-change-query-20261014-01`
(3.308 s). Generation `remediation-view-change-generate-20261014-03` passed
(10.430 s), followed by finalization `-01` (24.630 s). Type checking passed in
`remediation-view-change-types-20261014-01` (21.677 s), graph contracts in
`remediation-view-change-graph-20261014-01` (25.790 s), and Markdown lint in
`remediation-view-change-markdown-20261014-01` (22.710 s), all outer exits 0.
The browser case was then strengthened to require an actual keyboard-anchor
change after the measured completion. Browser qualification remains pending;
these unit and contract passes do not satisfy AC-044.

The first real sort/stable observation,
`remediation-view-change-browser-20261014-01`, failed in 161.582 s (outer exit 2,
`product` / `test_assertion_failure`). It collected zero samples and retained
an incomplete observation; all 27 cleanup activities completed. Inspection
attributed the failure to two new measurement defects: the header sort command
bypassed the toolbar acceptance mark, and `display: contents` row wrappers have
no painted bounds. The shared query timing hook now covers both command entry
points; the observer derives row geometry from its visible cells while keeping
record/version and mounted-row identity checks. The negative DOM fixture now
uses boxless row wrappers. This failed run does not establish a product timing
failure or an intermittent outcome. Finalization immediately before that run
(`remediation-view-change-finalize-20261014-02`) passed in 24.464 s;
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

`remediation-view-change-browser-20261014-02` passed the sort/stable row in
273.897 s (outer exit 0), with one warm-up, 100 measured samples, p95 270 ms,
threshold 1,000 ms, qualified evidence and zero scheduler overlap. All 27 cleanup
activities completed. Source:
`sha256:1cc0d5d4ac5333add1238ba2054d89136c1bbc75d7a0e620cae865f5c37d8af4`;
graph: `sha256:34c892d35d4c65963afe65fbec27d8fe7e0abf5f26b8354629b19ae138245904`.
Finalization `remediation-view-change-finalize-20261014-03` passed in 24.817 s
with maintenance skipped (`RESULTS_DIR` unset); its last few seconds overlapped
browser setup, before the published measurement interval. The browser source
and zero-overlap qualification above are the evidence boundaries.

Subsequent preparation keeps the warmed foreground page mounted and restores
the default query through ordinary controls. This avoids restarting the whole
application between every operation; preparation must still wait for the
default committed viewport. The observer samples one visible cell's geometry
per row, because cells in the fixed-height grid share row geometry. It still
validates every visible record/version and the filled row window. Fresh evidence
is required for the revised preparation and the other five predicate rows.

The six-row run `remediation-view-change-browser-20261014-03` was cancelled
through native session 64242 after source review proved that its filter setup
selected an unadmitted field/operator pair. Timeline synopsis is sortable but
has no filter operators; `contains` is not an admitted scalar filter operator.
The test's 30-minute total timeout also left selector actions unbounded. The run
exited 130 after 1,171.405 s, classified `interrupted` /
`cancelled_or_interrupted`; 11 setup units passed and 13 units were cancelled.
All 25 recorded cleanup operations completed. This is invalid setup, not product
latency evidence, an intermittent outcome, or a pass for the remaining rows.

The workload now uses admitted `timeline.capture_state = rough`, a
nonselective filter over the existing fixture. Its limitation is explicit in
Core 04. Generation checks each sort/filter/group capability against the typed
Timeline view contract; negative checks reject unsupported fields/operators.
Browser actions now have a 10-second diagnostic deadline (navigation 30 seconds)
independent of the unchanged performance thresholds. Accepted query metadata
also exposes bounded filter/sort counts and group-field identity, so identical
row membership cannot let the default query satisfy the new filter's stop state.
No raw filter value enters product timing metadata.

The corrected filter run `remediation-view-change-filter-20261014-01` failed in
198.916 s (outer exit 2, `product` / `test_assertion_failure`) after one warm-up
and 40 measured samples. The incomplete observation has no p95. Its preparation
check found that repeated post-measurement ArrowDown probes had advanced the
selection until the grid scrolled past the first default row. The probe now
selects the first rendered cell before each navigation check, and preparation
restores scroll position before the next measured action. All 27 cleanup
activities completed. No measured action is retried or scrolled by this repair.

Capability-check generation passed in
`remediation-view-change-generate-20261014-04` (10.280 s), the strengthened paint
row passed in `remediation-view-change-paint-20261014-04` (2.501 s), finalization
`-05` passed (24.520 s; maintenance skipped because `RESULTS_DIR` was unset),
type checking `remediation-view-change-types-20261014-03` passed (20.799 s), and
graph contracts `remediation-view-change-graph-20261014-02` passed (25.230 s),
all outer exits 0. The setup-state repair still requires fresh browser evidence.

`remediation-view-change-filter-20261014-02` failed in 157.284 s (outer exit 2,
`product` / `test_assertion_failure`) at the first post-observation keyboard
probe, with no complete measured sample set. The probe had selected the bulk
checkbox column rather than a semantic data cell. It now selects the first
committed summary cell and requires ArrowDown to reach the exact second record
in the same column. All 27 cleanup activities completed. The retained page also
showed zero visible collaborators; preparation now requires all 24 background
analysts in the foreground presence summary, with the same check before and
after each observation. Until this fixture check is demonstrated, previous
nominal session counts alone do not establish the intended presence load.

Finalization runs `remediation-view-change-finalize-20261014-06` and `-07`
passed in 24.379 s and 24.497 s (outer exits 0), respectively, with maintenance
skipped because `RESULTS_DIR` was unset. Format `-10` published a pass in
9.192 s, but its original outer exit was not retained across context recovery;
formats `-11` and `-12` passed in 9.079 s and 9.164 s with observed exits 0.

`remediation-view-change-filter-20261014-03` passed the corrected filter/first
useful row in 255.678 s (outer exit 0): one discarded warm-up, 100 measured
samples, p50 71.6 ms and p95 151.4 ms against 250 ms. Qualification is
`qualified`, scheduler overlap is zero, and all 27 cleanup activities completed.
Source: `sha256:d483562d4649036fcb1fc1a2a0b7e85427691bb0a5fae103ccdb7a764ec521e8`;
graph: `sha256:9d3325c37fa5f56777d421607313acb0b00ca36b581268088b592fdd390ed46e`.
The 24-collaborator assertions passed before and after each observation. The
earlier failure snapshot was captured after the driver's `finally` closed its
background contexts; its zero count therefore does not by itself establish
missing presence during measurement. No current presence defect was reproduced.
Finalization `remediation-view-change-finalize-20261014-08` passed in 24.719 s
(outer exit 0; maintenance skipped because `RESULTS_DIR` was unset).

After the user's Windows update, the checkout remained at the same baseline
commit and retained this completed run. The next invocation selects only the
five remaining AC-044 predicate rows, with a fresh run identity. No interrupted
run is relabeled or resumed as fresh evidence.

`remediation-view-change-remaining-20261014-01` passed all five remaining
AC-044 rows in 713.241 s (outer exit 0), with 22 work units passed and all
43 cleanup activities completed. Every predicate retained 100 measured samples
and qualified evidence with zero overlap: filter/stable p95 220.7 ms;
group/first-useful 150.4 ms; group/stable 165.4 ms; sort/first-useful 161.8 ms;
sort/stable 194.5 ms. Source matches the earlier filter/first-useful pass:
`sha256:d483562d4649036fcb1fc1a2a0b7e85427691bb0a5fae103ccdb7a764ec521e8`.
Graph: `sha256:c9a839e0c0c1da2f2cd09a0b92fe4c534889387bd0fc0be7d6670fec3cab6acc`.
This closes execution for the six exact Timeline workloads, with their declared
nonselective-filter and single-group limitations. It does not close the other
timed criteria, all surface applicability, Reporting adoption, or cost evidence.


## Presence and history continuation

The three exact presence workloads now have qualified implementation evidence
at `remediation-presence-browser-20261014-04`, including real active-cell focus
publication and correct current-generation evidence finalization. The
[timed assurance review](timed-product-assurance-review.md) records each
predicate, source identity, observation, failed predecessor and cleanup outcome.
This does not qualify the unmeasured timed criteria or a clean cost portfolio.

History measurement exposed an incident-wide projection rebuild in interactive
rollback/delete/restore. Revisions now receives only exact row-refresh
capability; source-owned Links and Entities dependency readers bound the direct
projection scope. Existing source/history mutation identities remain separate.
Transactional, lock-precedence, unrelated-row, retained-link restore and
projection-owner regressions have passing evidence; the full Revisions run's
three missing assembly rows passed after its test double adopted the new port.
The three browser predicates have current v2 qualification: change-set 348 ms,
single-entry 241.3 ms, and row restore 231.7 ms in the exact roots recorded in
the timed review. V2 fixes
grid scroll and geometry while measuring ordinary inspector focus reveal, with
one prepared keyboard confirmation and no action retries. Both failed browser
roots and their diagnoses remain in the timed review. Reporting adoption,
remaining timed criteria, cost windows and browser consolidation remain open.

The resumed Reporting slice adds closed retained-job decoding, exact admitted
request/model/row consistency, and one shared output-options codec for releases
and Composition previews. Defaults now precede preview idempotency and
persistence. New and extended tests, their initial compile/boundary failures,
and the retained pass roots are recorded in the Reporting cutover handoff.
Reporting's import guard now constrains permitted package interfaces rather
than private filenames. A separate boundary check found and removed the direct
Records read from Entities' history-projection dependency reader; Revisions'
incident validation now has explicit adversarial unit coverage. No coordinated
adoption or overall nine-finding completion is claimed.

## AC-047 continuation

The exact seeded large-grid edit/refresh trace is implemented and demonstrated
by `remediation-anchor-trace-browser-20261014-04`: 11 steps, 154 observed frames,
versions 2→3→4, no invariant violations, maximum 42 mounted rows, zero scheduler
overlap, all 27 cleanup operations completed, and outer exit 0. The
[product assurance review](timed-product-assurance-review.md#qualified-ac-047-implementation-evidence)
records source applicability, the new non-timed evidence contracts, rejected
detector assumptions and retained failure roots. This advances Gap 2 without
claiming overall remediation completion. Other timed job/Evidence criteria,
Reporting renderer/adoption conditions, the clean cost portfolio and final
integrated evidence remain open.

## Reporting selected-rule and token continuation

Reporting now rejects unresolved external disclosure partitions without rewriting
selected actions, separates selection from token assignment, and requires explicit
snapshot-bound token rules. Deterministic per-family/neutral ordinal display tokens
replace hash-fragment labels. The
[Reporting review](reporting-cutover-adoption-review.md#selected-rule-and-subject-token-corrections-after-restart)
records the exact scope, owner and portability passes, the repaired test compile
and lint errors, and an unattributed Composition runtime-cleanup failure. Canonical
model/schema and real-renderer obligations still block coordinated adoption.

## Restart continuation: Reporting owner boundaries

After the forced Windows update, source-binding verification was recovered and
the provider constructor was changed to consume explicit source identities.
Unused flat provider wrappers were removed; the source query boundary preserves
exact integer tokens and validates complete record/subject/Timeline bindings.
Full Reporting and Composition owners, selected Incident Bundles portability,
lint and generated-artifact checks passed for that slice. Exact roots and
reviewed survivor assertions are in the Reporting cutover review.

The following slice removes the hidden flat snapshot-field cache and redaction
classification/path fallbacks. Structured fixtures now exercise production
materialization. Current Reporting and Composition owner checks, lint and Markdown lint pass;
two fixture assertion failures and their corrections remain recorded. Coordinated adoption
is withheld, including the newly explicit conflict between Core disclosure
classes and Reporting's presentation vocabulary. Real rendering, full canonical
models, remaining timed criteria, qualified cost windows and the final integrated
handoff remain required open work.

## Canonical cache evidence continuation

The Reporting owner audit exposed a deterministic reader mismatch: canonical
cache outcomes are separate `cache_hit`, `cache_miss` and `cache_bypass` events,
but the terminal projection attempted to copy a nonexistent `cache_status`
member. The event schema correctly forbids that member. The shared reducer now
derives each terminal's reuse status from the canonical outcome event, with
null for no recorded outcome. It rejects duplicate/conflicting or late outcomes
and a child start after a hit. No persisted schema, legacy reader, audit waiver,
or invented child exit is introduced.

The owner-audit cache regression now feeds actual current-schema events through
the shared reducer instead of assigning `cache_status` in a mock terminal.
The regression root `remediation-cache-evidence-regression-20261014-01`
reproduced both failures (10.404 s, outer exit 2,
`product / test_assertion_failure`, 11 completed cleanup steps). The focused
repair root passed in 10.520 s. The complete evidence-accounting owner passed
at `remediation-cache-evidence-owner-20261014-01` (13.562 s, 10/10 units,
19 completed cleanup steps). Formatting, finalization and lint passed in their
corresponding `remediation-cache-evidence-*-20261014-01` roots (4.752 s,
18.066 s and 24.090 s). Retained-run maintenance was skipped because
`RESULTS_DIR` was unset. All successful owning exits were 0.

At source
`sha256:39aaa93e3a74081ebc5f13237e1a073fcecff10d4e258233e7052ff157f85e27`,
Reporting and Composition full-owner slices passed in the corresponding
`reporting` and `composition` roots (24.400 s and 15.472 s). Their explicit
owner audits passed in `reporting-audit` and `composition-audit` roots (2.532 s
and 2.537 s). The Reporting pass used no cache hits, so an additional
`remediation-cache-evidence-reporting-warm-20261014-01` run exercised four
recorded hits and four fresh units (19.616 s). Its audit
`remediation-cache-evidence-reporting-warm-audit-20261014-01` passed (2.405 s).
This demonstrates the original reuse path without converting cached results
into fresh executions. The earlier failed audit remains retained and attributed
in the Reporting handoff. These durations are diagnostics, not a qualified
cost window; owner-row closure does not adopt Reporting or prove every clause.

## Evidence metadata prerequisite (AC-045 remains open)

Core 01 REQ-01-328a now owns a bounded, authenticated metadata read for an
explicit incident-scoped Evidence window. The additive POST accepts 1–100
ordered distinct canonical UUIDs and returns current record versions, nullable
titles, lifecycle/blob state, and preview eligibility. Eligibility is explicitly
not an issued handle or object-availability proof. One database statement
provides the batch snapshot; any missing, deleted, or foreign identity rejects
the entire selection. Ordinary session/CSRF/member checks remain in force.
The route has no object-store call or mutation path. Existing issuance continues
to verify objects and redemption rechecks authority and version.

Authored OpenAPI, error registration and TypeScript operation inputs were
updated, then generated through Make. Six additive compatibility entries were
reviewed (one route, five schemas). No database migration or compatibility
reader is needed; rollback removes the new consumer before removing the route.
Tests reject malformed/oversized/dynamic identity selections, verify ordering
and null shape, exercise cross-incident rejection and membership loss, and
compare records, evidence, blob metadata, custody, idempotency, revisions and
projections before/after reads. The shared handle policy retains its existing
adversarial checks.

Execution roots under `.cartulary/test-results/` use prefix
`remediation-evidence-metadata-`. At executable source
`sha256:b01a24d0b63fbddfc04db3339376344b9bc41860075b56b2f4914ed7c76cbd38`:

| Suffix | Public command | Outcome |
| --- | --- | --- |
| `generate-20261014-01` | `make generate` | Failed: reviewed OpenAPI changes covered 399 of 405 findings; retained, not waived |
| `generate-20261014-02` | `make generate` | Pass, 10.070 s, after reviewing six additions |
| `format-20261014-02` | `make format` | Pass, 4.739 s |
| `adversarial-20261014-01` | Exact metadata unit + handle integration rows via `make test-slice OWNER=module.evidence` | Pass, 13.018 s |
| `finalize-20261014-01` | `make agent-finalize` | Pass, 18.282 s; maintenance skipped because `RESULTS_DIR` was unset |
| `handles-20261014-01` | Eight metadata/access/handle rows via `make test-slice OWNER=module.evidence` | Pass, 18.073 s; 6 units, 21 cleanup steps, no cleanup failure |
| `lint-20261014-01` | `make lint` | Pass, 32.667 s |
| `contracts-20261014-01` | `make generate-drift` followed by other Make goals | Drift passed, 17.048 s; outer exit 2 because the next goal correctly rejected reuse of the nonempty run ID |
| `artifact-policy-20261014-01` | `make generated-artifact-policy-check` | Pass, 0.816 s, fresh identity |
| `shapes-20261014-01` | `make json-shape-check` | Pass, 4.430 s, fresh identity |
| `openapi-20261014-01` | `make openapi-compatibility-check` | Pass, 0.868 s, fresh identity |

Except the explicitly retained failures, owning commands exited 0. Exact row
selections remain in each run manifest. Earlier generation/format invocations
precede the final adversarial-test source state; they are not falsely bound to
that later digest. The inspector consumer, real evidence-heavy fixture, exact
visible stop predicate and 100-operation qualification remain required. This
backend prerequisite does not close AC-045, the Evidence owner, or any overall
remediation finding.

## Timeline Evidence metadata window continuation

Core 03 REQ-03-248a now specifies a five-record window over the canonical
linked-record order, with total linked records distinct from attached-file
count. The inspector reads only that window, displays current labels, lifecycle,
file state and preview eligibility, and preserves an eligible download when
only preview is blocked. Loading, read failure and authority loss cannot enable
file actions. Explicit local retry performs a read; metadata never mints a
handle. Pending reads and accepted results are invalidated on page, source
observation, incident, session, scope, close/reopen or adapter replacement.
Revisiting a prior window cannot resurrect its earlier accepted observation.

The HTTP adapter owns envelope validation, exact incident/ordered-ID binding,
version and eligibility consistency, aborts and the existing 30-second read
observation bound. Its semantic port carries domain metadata rather than a
protocol-envelope type. Timeline composition wires independent metadata and
handle owners; controllers do not import sibling controllers. This corrects
the first implementation's rejected coupling without changing architecture
policy or adding exemptions. Existing attachment operations keep their own
lifetime. The new interface is additive and requires no data migration. A
rollback must remove the consumer before removing the endpoint.

The old isolated panel assertion assumed access actions without a metadata
owner. It now verifies that admission is required. Its label/action detection
survives in `module.evidence.frontend.metadata_window_presentation`, with the
actual metadata and handle owners, and in
`module.evidence.browser.linked_metadata_window`. The browser test verifies
seven real records, canonical order, five then two rendered items, separate
zero attached-file/seven linked-record counts, keyboard pagination, and zero
handle requests. Existing real preview, mixed blocked preview/Space, and
transient-storage recovery paths remain selected. A blocked preview is now
identified before invocation; its direct issuance denial remains covered by
backend handle tests, and Space discovery still exercises current admission.

At source
`sha256:f5f77ff5c2d7204ac6e2b39800a80ab27a1614c67421811083c0871dff3f5e1e`,
all roots below are under `.cartulary/test-results/` with prefix
`remediation-evidence-window-`:

| Suffix | Command / scope | Result |
| --- | --- | --- |
| `generate-20261014-05` | `make generate` after authored selectors | Pass, 9.030 s |
| `format-20261014-04` | `make format` | Pass, 4.825 s |
| `focused-20261014-04` | Three exact metadata frontend rows via `make test-slice OWNER=module.evidence` | Pass, 4.170 s |
| `panel-20261014-02` | Exact `module.workbook` panel row | Pass, 3.011 s |
| `imports-20261014-02` | `make frontend-import-boundary-check` | Pass, 6.151 s |
| `finalize-20261014-02` | `make agent-finalize` | Pass, 18.198 s; maintenance skipped because `RESULTS_DIR` was unset |
| `browser-20261014-02` | Metadata paging, linked text review, mixed blocked preview, transient recovery via four exact `module.evidence` rows | Pass, 35.872 s, 11 units, 22 successful cleanup steps |
| `lint-20261014-02` | `make lint`, including TypeScript and boundaries | Pass, 28.701 s |
| `drift-20261014-01` | `make generate-drift` | Pass, 17.912 s |

All successful owning exits were 0. Generation/format preceding the final
source projection are retained at their own manifest identities. Earlier
failures remain retained: `types-01/02` identified missing test-fixture wiring,
operation-family registration and mock type inference; `imports-01` and
`lint-01` identified the four boundary violations described above; `panel-01`
identified the old no-metadata expectation. `generate-04` rejected its stale
selector after that test was renamed. Their dates use `20261014`; no failed
root was overwritten or treated as a pass. The earlier `regression-01` passed
existing handle-presentation and Space resolution tests (3.959 s); those owner
sources were restored unchanged when sibling-controller coupling was removed.

Seeded UI review used `REVIEW_PROFILE=default`, session
`uireview-9042ce2aa47ee2ba26ac10fecb1aa870`, at
`remediation-evidence-window-review-20261014-01`. Operation 12 captured the
actual Timeline Evidence inspector with two linked available records at
1440×900. The image and observation byte counts/digests were verified before
inspection. Labels, state text, preview/download controls and attachment entry
were readable in the adjacent inspector while the grid stayed visible. Axe
completed with no violations and no unassessed frames; color-contrast checks
remained incomplete for existing grid/view controls, so this is not a blanket
accessibility pass. The page-window keyboard mechanism is covered separately
by the real browser row; this two-record review does not visually qualify a
five-record or 100-record workload. Raw review artifacts are private and expire
on cleanup; retained operation receipts identify the review.

Digest acceptance assessment: A001–A003 and A026 are supported by the Core
01/03 owner amendments, semantic adapter/composition review, and additive
interface/rollback record. A004–A005 use existing tokens/theme without changes.
A010–A011, A016–A018 and A023 are supported by current-source lifetime, binding,
blocked-state, real interaction and selector tests. A024–A025 are supported by
unchanged documentation-free executable boundaries and generated drift. A006,
A008–A009, A019–A022 remain limited to the unchanged geometry and observed
states above; expanded performance/visual qualification is not inferred.
A007 and A012–A015 do not change: creation, transaction identity, acknowledgement,
editing and conflict ownership remain with their existing owners. A027 is this
retained handoff. No overall completion or AC-045 timing claim follows from
these implementation checks. The evidence-heavy snapshot, exact predicate
review and 100-operation p95 qualification remain the next dependent work.

### UI review console and cleanup disposition

The first seeded review completed cleanup and exited 0. Its terminal receipt
reported two console errors, so a second default-profile review,
`remediation-evidence-window-review-console-20261014-01`, inspected the initial
unauthenticated page before opening any workbook. Operation 1 contained exactly
two untruncated console entries, correlated with the session read returning 401
and Enterprise provider discovery returning 404. The default profile does not
claim Enterprise Authentication; Core 01 reserved-family dispatch, the platform
HTTP route tests and the browser discovery owner explicitly handle
`extension_profile_not_claimed`. These initial responses do not establish an
Evidence-panel failure. This diagnosis does not retroactively identify every
request in the first review, whose raw observations had already expired.

The second session, `uireview-b30f083cf2b719a5871df8794e7c2313`, was stopped through
its exact locator. Its terminal receipt reports `cleanup=complete`, no primary
failure, zero failed requests, and exit 0; the foreground owner also exited 0.
Only this task's request scratch file and directory were removed.

A documentation consistency review also corrected Harness REQ-815's copied
measurement generation list to observation v4, summary/aggregate v5 and browser
wrapper v6, matching the already implemented typed owners. Fixture key, runtime,
build and lease remain v2. Single, percentile and continuous-trace observations
retain their distinct sampling semantics; no executable consumer reads this
prose.

## Structural retirement continuation

### Current closure

Finding 6 is closed for the identified assertions. Each candidate in this
review has a contractual invariant and a named behavioral, interface or
import-boundary survivor. Exact contract identities remain exact; incidental
names/layout no longer decide conformance. The counterexamples exercise
forbidden imports/capabilities as well as harmless aliasing, extraction,
shadowing and comments. No browser scenario was removed in this workstream.
Future removals still require an obligation/evidence review; this closure is
not a blanket endorsement of every static assertion in the repository.

Fresh Make verification used the following exact run IDs under
`.cartulary/test-results/`; all owning native exits were 0. These are functional
checks, not cost windows. The Reporting boundary replacements also passed in
the full Reporting owner at `remediation-subject-model-owner-20261015-01`.

| Run | Selection | Result |
| --- | --- | --- |
| `remediation-structural-current-architecture-20261015-01` | `make test-slice OWNER=web.architecture` | Pass, 10.925 s; complete architecture owner |
| `remediation-structural-current-workbook-20261015-01` | `web.workbook.boundary_support.declarative_surface_policy` and `web.workbook.boundary_support.source_workbookinspectorownershiparchitecture_test_t_8849e53c39` | Pass, 3.186 s |
| `remediation-structural-current-network-20261015-01` | `web.networkflow.boundary_support.semantic_contract_boundaries` | Pass, 3.836 s |
| `remediation-structural-current-boundaries-20261015-01` | `harness.command_surface.behavior.owner_boundaries` | Pass, 21.061 s; actual-checker counterexamples included |

### Retirement and survivor review

The identified frontend guards now test observable contracts and actual imported
capabilities. No product behavior, interface, evidence profile or browser scenario
was removed. Candidate replacements passed before retirement. The reviews below
were performed by the remediation agent and are retained for human review; they
are not specification adoption or an independent human sign-off.

| Candidate / classification | Behavior review and named survivor | Execution and evidence review |
| --- | --- | --- |
| Layout row-count/private filler names: incidental spelling around Core 03 REQ-03-286 | `workbook layout policy preserves exactly the supplied record population without layout filler` checks zero, one, three and 50 real records, identities and versions. The existing `Workbook inspector fills the shell work area across row counts` measures actual grid/inspector/work-area bounds, document overflow and independent scroll at zero, one, three and many records. AST surface-geometry ownership remains. | Replacement unit test and existing real-browser row passed before the name/proximity regex was removed. The new operational row is `web.architecture.boundary_support.workbook_geometry_contract`; the browser row remains `module.workbook.browser.source_sentinel_spec_ts_38af00ea72`. No browser fixture or selection is reduced. |
| `clientTxnRef` and 180-character entropy proximity: incidental spelling around Core 03 REQ-03-301 | `createClientTransactionId` tests retain platform UUID, secure-byte fallback and unavailable-crypto behavior. The new facade test rejects clock/Math randomness, observes fresh secure IDs and local failure without Web Crypto. `semantic mutation command ports` proves exact owner request identities and secure-ID failure without transport. Source ownership now recognizes the actual generator imports and namespace capabilities; unrelated local labels with the old names pass. | `module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_16d9b2afa3` and `web.workbook.regression.workbookz_mutation_runtime_semantic_command_ports_7b2f1d02a4` passed before deleting `clientTransactionIdPolicy.test.ts` and its `...f313f22abb` row. Source ownership remains in its existing `web.architecture` row. No retry/re-key behavior or assertion is deleted. |
| Network Flow helper name and private coordinate names: incidental realization | The real claimed-empty, unclaimed and complete claimed-analysis browser tests exercise profile admission and the semantic grid. Existing vendor import isolation rejects aliased/dynamic imports outside the Grid Adapter. AST checks reject the contractual `setContent` capability, including aliases/destructuring, in the real scenario and its support sources. Exact schema/type identities remain checked as syntax; comments and helper renaming pass. | All three real browser rows passed before removal. `web.networkflow.boundary_support.semantic_contract_boundaries` now owns the retained syntax/design checks. No synthesized browser page counts as authoritative application evidence. Public token-variable identities remain legitimate exact checks. |
| Surface-policy transport keywords, workflow helper names and Assessment symbol names: incidental realization | The shared import owner enforces policy transport isolation, vendor isolation and both Assessment/Timeline directions. `frontend owner boundaries survive aliasing extraction and comments` runs the actual checker against forbidden aliased/re-exported/dynamic imports and harmless extraction. Runtime registration tests retain exact owner IDs and verify recursively immutable plain policy data with no executable capabilities. Owner-local coordination and Evidence tests retain role gates, draft guards, retained command lifetime and current access behavior. | New `web.workbook.boundary_support.declarative_surface_policy` replaces the two old policy rows. Canonical `harness.command_surface.behavior.owner_boundaries`, `module.tasksdecisions.frontend_unit.coordination_workflow_lifecycle_7a8d13c2f0` and `web.workbook.regression.evidence_access_request_lifetime_ea01` passed before retirement. The generic renderer filename is not promoted to a normative workflow boundary. |
| Raw `fetch` token: retained architecture invariant with corrected detection | `frontend transport boundary policy` now binds ambient `fetch`, browser-global properties and aliases rather than unrelated property keys or shadowing parameters. Direct/global/computed/destructured/aliased acquisition outside the service owner fails; benign same-named local data/functions pass. | Existing operational row and source scope remain. This is a bounded syntax/symbol detector, not a proof of arbitrary dynamic JavaScript behavior. Type/interface, import and behavioral checks remain independent assurance. |

Seven missing frontend ownership entries from the prior metadata/timing work
were also assigned explicitly to `web.workbook`: the metadata port, adapter and
its test, metadata hook and its test, metadata presentation test, and query-action
timing hook. Their existing catalog selection was not replaced by this manifest.

All roots below are under `.cartulary/test-results/`, prefixed
`remediation-structural-`, dated `20261014`. Successful owning exits are 0.
These are verification runs, not qualified cost observations.

| Suffix before date / ordinal | Command and scope | Result |
| --- | --- | --- |
| `survivors-generate` / `01` | `make generate` | Pass, 10.110 s |
| `survivors-format` / `01` | `make format` | Pass, 4.749 s |
| `architecture` / `01` | Three exact `web.architecture` rows | Failed, 3.737 s, outer 2; new probe used invalid view identity, and seven ownership entries were missing; retained |
| `architecture` / `02` | Same rows after the two diagnosed repairs | Pass, 4.068 s |
| `workbook` / `01` | Two original policy rows plus new data invariant | Pass, 2.750 s |
| `network` / `01` | Original Network Flow row plus replacement syntax checks | Pass, 3.989 s |
| `crypto` / `01` | Generator and browser-facade row | Pass, 3.457 s |
| `survivors-finalize` / `01` | `make agent-finalize` | Pass, 18.092 s; `RESULTS_DIR` unset, maintenance skipped |
| `imports` / `01` | `make frontend-import-boundary-check` | Pass, 6.508 s |
| `boundary-detectors` / `01` | Canonical harness boundary row | Pass, 20.305 s; before the new import counterexamples were enrolled |
| `boundary-detectors` / `02` | Same row including actual-checker counterexamples | Pass, 18.783 s |
| `layout-browser` / `01` | Exact real geometry row | Pass, 29.267 s |
| `network-browser` / `01` | Three exact real profile/mechanism rows | Pass, 34.519 s |
| `owner-mechanisms` / `01` | Exact mutation-port and Evidence lifetime rows | Pass, 4.046 s |
| `coordination` / `01` | Exact coordination workflow row | Pass, 4.015 s |
| `retirement-generate` / `01` | `make generate` after reviewed removals | Pass, 8.920 s |
| `retirement-format` / `01` | `make format` | Pass, 4.885 s |
| `architecture` / `03` | Final geometry/source/transport rows | Pass, 4.289 s |
| `workbook` / `02` | Final declarative policy row | Pass, 2.458 s |
| `network` / `02` | Final Network Flow boundary row | Pass, 3.649 s |
| `retirement-finalize` / `01` | `make agent-finalize` | Pass, 19.231 s; `RESULTS_DIR` unset, maintenance skipped |
| `retirement-drift` / `01` | `make generate-drift` | Pass, 13.854 s |
| `retirement-lint` / `01` | `make lint`, including TypeScript and import boundaries | Pass, 22.053 s |
| `retirement-markdown` / `01` | `make lint-markdown` | Pass, 21.020 s; before this result-only record was appended |

The final focused source identity is
`sha256:5696efe9c6ffe2efa3cfad3d95e6dd73f73a6b5e301976bc1b97baabe0f432dd`.
Earlier survivor roots retain their own exact source identities and selectors in
run manifests. Removed operational row IDs are not reused for changed detector
suites. Rollback is source/catalog/projection only; no data migration is involved.
No consolidation speedup, full-current-source gate or overall remediation closure
follows from these focused results. AC-045 object-backed snapshots, other timed
criteria, Reporting production/adoption work and qualified cost windows remain
required.

## Object snapshot lifecycle continuation — 2026-10-14

This continuation resumes the interrupted remediation after the Windows update.
It closes the generic object-storage prerequisite for AC-045; it does not admit
an Evidence-heavy profile or claim AC-045 performance conformance.

TH-HARNESS-REQ-814, REQ-815 and new REQ-820 now specify an immutable binary
population alongside the populated database. The active profile registry,
build and lease schemas advance to v3. Runtime credentials and snapshot-key
envelopes retain their unchanged v2 shapes; observations remain v4, summaries
and aggregates v5. Retired registry/build/lease v2 schemas are removed from the
supported surface. Existing bytes remain historical evidence and cannot be
translated or used to qualify the new lifecycle.

The v3 registry declares exact physical `object_population` counts and bytes.
Those expectations enter the profile source-contract digest. Source-owner
assembly receives an ordinary object-store port backed by a private filesystem
store. After assembly and semantic validation, the builder hashes the actual
bytes and seals an RFC 8785 manifest with an exclusive write. No raw object key,
content or runtime path enters the retained projection. The current large-grid
profile declares an explicit empty population; the generic integration fixture
uses a separate bounded synthetic profile with actual object bytes.

Before allocating a clone, admission reads this run's closed, bounded, regular
build artifact and checks the exact identity, contribution receipts, semantic
results, population and current schema. Each clone restores into its own empty
S3 bucket, checks the source and then reads back the destination bytes and MIME
metadata. Build, clone, summary and aggregate provenance must agree on the
manifest digest and physical population. Required zero-valued proof fields
cannot be omitted or null. Failed restoration returns no usable fixture and
retains actual acquired-resource cleanup outcomes. Suite cleanup removes the
private payloads, manifest and credentials together.

Behavior review: object identity, byte integrity, MIME metadata, isolation and
admission are enforced by generic lifecycle mechanics; Evidence meanings and
workload counts remain source-owner/profile responsibilities. Execution review:
the existing lifecycle and harness graph rows remain the canonical surfaces;
the new `harness.browser.unit.performance_fixture_objects` row owns adversarial
object-integrity cases. These are agent reviews for human projection-fidelity
review, not an invented independent sign-off.

All run IDs below are under `.cartulary/test-results/` and were invoked from the
repository root with explicit `CARTULARY_TEST_RUN_ID`. Focused final evidence uses
source digest
`sha256:2aa360027c4e6fdd3f775a434cb112ce60365e7cb4074eab4928c27a9a72f573`.

| Run ID | Public command / selection | Result and applicability |
| --- | --- | --- |
| `remediation-fixture-objects-v3-generate-20261014-02` | `make generate` | PASS, 9.160 s; authored registry/schema/generator projections |
| `remediation-fixture-objects-v3-format-20261014-05` | `make format` | PASS, 4.775 s |
| `remediation-fixture-objects-v3-unit-20261014-03` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.performance_fixture_lifecycle,harness.browser.unit.performance_fixture_objects,harness.browser.unit.testservices_lifecycle_contract` | PASS, 3.249 s; strict admission, corruption, unsafe paths, sparse S3 listings, required proof fields and cleanup evidence |
| `remediation-fixture-objects-v3-integration-20261014-04` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.performance_fixture_snapshot_lifecycle` | PASS, 16.203 s; four real isolated PostgreSQL/S3 clones, byte/MIME checks, mutation isolation, corrupt-source rejection, cleanup and private-root removal |
| `remediation-fixture-objects-v3-graph-20261014-02` | `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | PASS, 26.070 s; graph and measurement evidence contracts including mismatched object proofs |

The first integration run,
`remediation-fixture-objects-v3-integration-20261014-01`, failed with
`product/test_assertion_failure`: the new generic verifier incorrectly assumed
S3 listing entries included MIME metadata. Listings establish membership/size;
read metadata now supplies the actual MIME value, which remains checked against
the sealed manifest. A sparse-listing regression and real S3 restoration cover
the repair. The second integration run, ending `-02`, reached the failure-path
assertion but used the production registry resolver for the deliberately
synthetic test profile; its test was corrected to use the same explicit-profile
boundary as the existing integration preparation/cleanup helpers. Both failed
runs had 14 successful cleanup steps, zero failed or blocked cleanup steps, and
outer Make exit 2. They are retained deterministic development failures, not
claims of intermittent behavior. Passing runs have outer exit 0.

No data reset, schema migration or application downgrade is required for this
harness-only generation change. Existing invocation roots remain immutable;
new runs construct fresh private resources. Rollback requires matching harness
sources and fresh execution, never relabeling old evidence. The remaining
AC-045 work is its populated owner fixture, exact revised predicate/workload,
product marks, paint proof, adversarial measurement validation and actual
100-operation qualification. Reporting promotion, remaining timed criteria,
qualified cost windows and reviewed browser consolidation remain open.

Post-finalizer validation used the same implementation source digest:

| Run ID | Command | Result |
| --- | --- | --- |
| `remediation-fixture-objects-v3-finalize-20261014-01` | `make agent-finalize` | PASS, 18.028 s; schema/shape and generation work completed. Retained-run maintenance was skipped because `RESULTS_DIR` was unset. |
| `remediation-fixture-objects-v3-lint-20261014-01` | `make lint` | PASS, 24.314 s |
| `remediation-fixture-objects-v3-drift-20261014-01` | `make generate-drift` | PASS, 15.534 s |
| `remediation-fixture-objects-v3-boundaries-20261014-01` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries` | PASS, 22.910 s; active/retired generation and generic-profile boundaries |
| `remediation-fixture-objects-v3-paste-20261014-01` | `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.measurement.paste_20x5_v2` | PASS, 149.046 s; 14/14 work units, 27 successful cleanup steps, zero failed/blocked cleanup. Production 20,000-row builder, explicit empty object population, real browser admission and immutable v3 build/lease evidence all exercised. |

The paste invocation is current-source AC-003 implementation evidence and a
lifecycle regression, not a Core 05 publication qualification or an AC-045 test.
Previously retained v2 build/lease runs remain useful historical investigations
but do not qualify this active evidence generation. No current full-check,
release or portfolio qualification is claimed by these focused runs.

`remediation-fixture-objects-v3-policy-20261014-01` ran
`make generated-artifact-policy-check`: PASS, 0.803 s, outer exit 0. The production
paste observation was one synthesized TSV paste at 409 ms, below its strict
2,000 ms bound; all 20 new identities and 100 visible mapped cells were required
by its existing predicate. No Excel application interoperability is implied.

## Evidence object contribution prerequisite — 2026-10-14

Evidence now owns an isolated fixture-construction capability, consumed only by
application test composition. It writes actual object bytes and upload-complete
blob metadata, then uses normal Evidence create admission and mutations for
each envelope, lifecycle, revision and workbook projection. It does not add an
HTTP route, broaden Workbook dispatch, invent upload leases, or mint preview
handles. The populated-snapshot builder remains responsible for discarding a
partially constructed database and private object namespace.

The Evidence-owned payload helper produces a deterministic, fully decodable
2 by 2 PNG of exactly the requested physical size, up to 2 MiB. A valid PNG text
chunk supplies the synthetic payload mass. The integration test uses three
separate 2 MiB objects, verifies actual bytes and SHA-256 metadata, checks normal
revision/projection binding, and rejects invalid or truncated PNGs, unsupported
sizes, duplicate construction, cancellation and an object-store write failure.
The five exported fixture capability names have an explicit Evidence surface
classification; this is an application-construction capability, not a new
product interface.

The current production large-grid adapter receives the ordinary typed object
store through application composition. The same lifecycle integration still
passes with this dependency. No Evidence-heavy profile has yet been enrolled:
10,000 Evidence records, 5,000 Timeline rows, the 100-record attachment set and
AC-045 measurement remain required subsequent work. Three objects are fixture
capability evidence, never evidence-heavy workload qualification.

| Run ID | Command / selection | Result |
| --- | --- | --- |
| `remediation-evidence-fixture-generate-20261014-01` | `make generate` | PASS, 9.160 s |
| `remediation-evidence-fixture-format-20261014-04` | `make format` | PASS, 4.737 s |
| `remediation-evidence-fixture-integration-20261014-02` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.evidence_object_fixture,harness.browser.integration.performance_fixture_snapshot_lifecycle` | PASS, 14.635 s; actual Evidence records and four-clone generic lifecycle |
| `remediation-evidence-fixture-surface-20261014-01` | `make test-slice OWNER=module.evidence ROWS=module.evidence.surface.exported_surface_reachability` | PASS, 4.326 s; reviewed exported capability boundary |
| `remediation-evidence-fixture-integration-20261014-04` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.evidence_object_fixture` | PASS, 15.048 s; final complete-PNG validation and mutation/projection assertions |

The last integration source digest is
`sha256:0473ba7f4dbb65497e2b7cc97be3eeb01c5bbe3dd3ebe672659639b388d78ae5`.
All listed invocations have outer exit 0 and reside under
`.cartulary/test-results/`. These tests use disposable resources; no user data
reset or compatibility reader was introduced.

Continuation after the workstation restart: the retained
`remediation-evidence-fixture-finalize-20261014-02` summary reports PASS
(17.868 s), as does its retained-secret scan. Its native process exit was not
recoverable after restart, so complete invocation success is not inferred from
that summary alone. The preceding `-01` finalizer failed because the new
Evidence test-support root was absent from the support inventory; the authored
inventory now assigns it to Evidence and includes support scanning.

Fresh validation caught and corrected capitalized Go error strings in the new
Evidence fixture helper (`staticcheck` ST1005). The failed lint invocation
`remediation-evidence-fixture-lint-20261014-01` has outer exit 2,
`harness/child_target_failure`, 10/11 units passing and all 20 cleanup steps
successful. This deterministic implementation issue is not an instability
allegation. Validation after that correction:

| Run ID | Command | Result |
| --- | --- | --- |
| `remediation-evidence-fixture-format-20261014-05` | `make format` | PASS, 4.821 s, outer exit 0 |
| `remediation-evidence-fixture-finalize-20261014-03` | `make agent-finalize` | PASS, 18.261 s, outer exit 0; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `remediation-evidence-fixture-lint-20261014-02` | `make lint` | PASS, 26.165 s, outer exit 0 |
| `remediation-evidence-fixture-policy-20261014-01` | `make generated-artifact-policy-check` | PASS, 0.960 s, outer exit 0 |

`remediation-evidence-fixture-drift-20261014-01` ran `make generate-drift`:
PASS, 12.227 s, outer exit 0.

## Inspector fixture attachment construction — 2026-10-14

Timeline test support now constructs background Timeline rows through its
existing isolated bulk capability and the selected Evidence-inspector target
through ordinary create admission and mutations. The 100 links obey the
existing 64-action limit: create commits 64 links, followed by a versioned patch
for the remaining 36. Both operations retain normal Evidence validation,
Links mutations, revisions and source-owner projection updates. A partial
construction is a failed disposable fixture, never a usable measurement seed.
No public mutation API, raised action limit or direct cross-owner link insert
was introduced.

The dedicated integration test creates 5,000 Timeline rows and 100 real Evidence
objects, checks both owners' projections, target versions 1 and 2, all 100
attachment mutation entries, exact link membership, malformed populations,
cancellation, duplicate construction and missing/cross-incident references.
Its 256-byte objects deliberately qualify construction and attachment scale
only. The required 10,000-object/20,971,520,000-byte population and 100 measured
inspector operations remain open.

The initial integration run failed `product/test_assertion_failure`, outer exit
2, because the helper attempted 100 actions in one create request. Ordinary
admission correctly rejected it. The helper now uses the existing create/patch
protocol. All 14 cleanup steps completed in that failed run. The corrected
integration passed but spent 113.140 s in the test: construction cost remains a
measured diagnostic requiring investigation when the full profile is assembled,
not a claim that fixture preparation is inexpensive.

| Run ID | Public command / selection | Result |
| --- | --- | --- |
| `remediation-inspector-fixture-generate-20261014-01` | `make generate` | PASS, 9.110 s, outer exit 0 |
| `remediation-inspector-fixture-format-20261014-02` | `make format` | PASS, 4.742 s, outer exit 0 |
| `remediation-inspector-fixture-integration-20261014-02` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.evidence_inspector_fixture` | PASS, 122.526 s, outer exit 0 |
| `remediation-inspector-fixture-finalize-20261014-01` | `make agent-finalize` | PASS, 18.097 s, outer exit 0; retained-run maintenance skipped because `RESULTS_DIR` was unset |

The integration source identity is
`sha256:49e7e2b2b535178f551d0bb656c5e908d63d1fcc692518c933a917bb48b8fb00`.
All roots are retained beneath `.cartulary/test-results/`.

Core 04 now defines `cartulary.perf.evidence_heavy.v2` and
`perf.evidence_inspector.metadata_shell.v2`: exact physical PNG population,
one ordinary foreground editor, required preflight of actual preview issuance
and redemption for all 100 links, and an independently timed five-item metadata
window. Inspector opening exposes current eligibility without automatically
issuing capabilities. The 1920 by 1080 viewport, ordinary `timeline.has_evidence=true` filter,
closed starting state, fresh metadata read, exact bindings, clipped inspector
visibility and two-frame stop are explicit. One warm-up, 100 operations and
p95 <=300 ms remain required. The former fixture/predicate IDs are retired;
there is no claim that eligibility equals handle possession or that historical
v1 evidence qualifies v2. This is authored specification closure for review;
typed projections, full snapshot enrollment, measurement and actual qualifying
evidence are still required.


Post-finalizer inspector-construction validation:

| Run ID | Command | Result |
| --- | --- | --- |
| `remediation-inspector-fixture-lint-20261014-01` | `make lint` | PASS, 21.474 s, outer exit 0 |
| `remediation-inspector-fixture-drift-20261014-01` | `make generate-drift` | PASS, 12.103 s, outer exit 0 |
| `remediation-inspector-fixture-policy-20261014-01` | `make generated-artifact-policy-check` | PASS, 0.817 s, outer exit 0 |
| `remediation-inspector-fixture-markdown-20261014-01` | `make lint-markdown` | PASS, 8.870 s, outer exit 0; precedes the subsequent supported-filter wording correction |

Workload projection review found that Synopsis deliberately has no filter
operators. AC-045 preparation now uses the already supported
`timeline.has_evidence=true` equality filter; only the target has Evidence links.
No filter capability was added to make the measurement possible.

## AC-045 v2 typed workload — 2026-10-14

`contracts/performance/ac045.v2.json` projects the revised Core 04 workload
through a closed `cartulary.performance_evidence_inspector_contract.v1` schema.
It binds the exact physical population, foreground session, preview preflight,
metadata window, sampling, threshold, viewport and supported boolean filter.
The independent Evidence-inspector generator checks the admitted view capability
and shared p95 policy and emits a dedicated frontend constant; the ordinary
contract generator also embeds the authored JSON for backend fixture consumers.
Neither generator reads specification Markdown. Human fidelity review remains
separate from these automated projection checks.

The existing routing-contract suite now rejects retired predicate/fixture IDs,
wrong object counts/bytes, unknown stubbing options, insufficient preview
preflight, automatic handle requests during timing, missing fresh reads,
one-frame/grid-only visibility, unsupported Synopsis filtering, changed sampling,
relaxed thresholds and retries. It also detects drift in the companion view and
measurement policy. No measurement row or Evidence-heavy snapshot is enrolled
by this projection change alone.

| Run ID | Command | Result |
| --- | --- | --- |
| `remediation-ac045-contract-generate-20261014-01` | `make generate` | PASS, 9.330 s, outer exit 0 |
| `remediation-ac045-contract-format-20261014-01` | `make format` | PASS, 4.854 s, outer exit 0 |
| `remediation-ac045-contract-routing-20261014-01` | `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | PASS, 24.992 s, outer exit 0 |

AC-045 workload projection post-finalizer checks:

| Run ID | Command | Result |
| --- | --- | --- |
| `remediation-ac045-contract-finalize-20261014-01` | `make agent-finalize` | PASS, 18.190 s, outer exit 0; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `remediation-ac045-contract-lint-20261014-01` | `make lint` | PASS, 22.106 s, outer exit 0 |
| `remediation-ac045-contract-markdown-20261014-01` | `make lint-markdown` | PASS, 10.090 s, outer exit 0 |

`remediation-ac045-contract-drift-20261014-01` (`make generate-drift`) passed
in 12.083 s; `remediation-ac045-contract-policy-20261014-01`
(`make generated-artifact-policy-check`) passed in 0.808 s. Both outer exits
were 0.

## Evidence/Timeline contribution wiring — 2026-10-14

Evidence test support now exposes a closed fixture contribution with exact
record, object and physical-byte count receipts. Record identities remain in a
private, copied application capability. Timeline's inspector contribution
consumes that capability, chooses the prescribed initial attachment set and
uses the validated ordinary mutation application. Shared fixture state and the
generic snapshot lifecycle receive no Evidence-specific identity fields.
Descriptors and returned identity slices are copied, and both contributions
reject a second attempt rather than implicitly retrying partial construction.

The existing inspector integration row exercises this composition with the same
5,000-row/100-object bounded population. It verifies receipts, descriptor and
identity isolation, no-retry behavior, all prior attachment/projection/history
assertions and rejection cases. It remains construction evidence only. The full
AC-045 profile, one-editor credential assembly, exact semantic validation,
sealed physical population, clone preparation, browser predicate/proof and
100-operation qualification are still open.

| Run ID | Command / selection | Result |
| --- | --- | --- |
| `remediation-evidence-providers-format-20261014-01` | `make format` | PASS, 4.744 s, outer exit 0 |
| `remediation-evidence-providers-integration-20261014-01` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.evidence_inspector_fixture` | PASS, 118.229 s, outer exit 0; 3/3 work units and all 14 cleanup steps successful |
| `remediation-evidence-providers-finalize-20261014-01` | `make agent-finalize` | PASS, 18.160 s, outer exit 0; retained-run maintenance skipped because `RESULTS_DIR` was unset |

The integration source identity is
`sha256:74c918e7a7f50e6e81f300ce234b4b3587cac33b5600a1da7163e23238ee262e`.
Per-stage test log messages are not retained for this successful invocation;
therefore no Evidence-versus-Timeline cost attribution is claimed. The enrolled
snapshot builder's canonical contribution diagnostics must supply that evidence.

Provider wiring post-finalizer checks passed with outer exit 0:
`remediation-evidence-providers-lint-20261014-01` (`make lint`, 21.021 s),
`remediation-evidence-providers-drift-20261014-01` (`make generate-drift`,
12.868 s), and `remediation-evidence-providers-markdown-20261014-01`
(`make lint-markdown`, 9.060 s).

## Fixture principal and workspace composition — 2026-10-14

Private fixture assembly now uses `PrincipalUserIDs` and Auth's
`CreateFixtureAnalyst` capability. Auth receives an explicit display-name prefix;
Incidents receives an explicit workspace identity with key prefix and title.
This permits the forthcoming ordinary foreground editor and Evidence-heavy
workspace without duplicating either owner's account, membership or workspace
creation. The large-grid adapter still supplies its existing background labels,
`AC043-PERF` incident prefix, title, 24 accounts, zero sessions and memberships.
The credential-set identity, runtime schema, receipts, seed and semantic profile
remain unchanged; no compatibility alias or reader was added.

| Run ID | Command / selection | Result |
| --- | --- | --- |
| `remediation-fixture-principals-format-20261014-01` | `make format` | PASS, 4.835 s, outer exit 0 |
| `remediation-fixture-principals-unit-20261014-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.performance_fixture_runtime_and_assembler,harness.browser.unit.performance_fixture_lifecycle` | PASS, 1.581 s, outer exit 0 |
| `remediation-fixture-principals-assembly-20261014-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.source_owner_contribution_assembler` | PASS, 7.640 s, outer exit 0 |
| `remediation-fixture-principals-integration-20261014-01` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.performance_fixture_snapshot_lifecycle` | PASS, 15.505 s, outer exit 0; four real isolated PostgreSQL/S3 clones and all 14 cleanup steps complete |
| `remediation-fixture-principals-finalize-20261014-01` | `make agent-finalize` | PASS, 18.108 s, outer exit 0; retained-run maintenance skipped because `RESULTS_DIR` was unset |

The integration source digest is
`sha256:5a987ad1601f950d2947e78bb99be9ff5d30c8d4573ca6367f88c4b7469a74c8`.

Principal/workspace post-finalizer checks passed with outer exit 0:
`remediation-fixture-principals-lint-20261014-01` (`make lint`, 21.577 s),
`remediation-fixture-principals-drift-20261014-01` (`make generate-drift`,
12.816 s), and `remediation-fixture-principals-markdown-20261014-01`
(`make lint-markdown`, 9.500 s).

A companion publication review remains necessary: the current
`tools/release-evidence/check-benchmark-claim.mjs` contains the retired
Evidence-inspector v1 and generic view-change IDs in a hard-coded p95 list,
and skips unknown predicates when checking sample counts. Its test fixture
uses an arbitrary fixture ID. Core 05 REQ-05-011 requires exact current
criterion/predicate/fixture binding. REQ-05-012's unconditional large-grid
traffic/warm-up wording also needs reconciliation with independently defined
Core 04 single-observation and Evidence-heavy workloads. No Core 05 publication
claim is established by these implementation runs; this known companion gap
must not be hidden by a no-manifest pass.

## Evidence-heavy profile and AC-045 measurement implementation — 2026-10-14

The active `evidence_heavy_snapshot_v2` profile now composes Auth, Incidents,
Evidence, Timeline and Projections through their existing owner contributions.
It requires one ordinary foreground editor, no pre-created sessions, 5,000
Timeline records, 10,000 Evidence records and 10,000 physical 2,097,152-byte
PNG objects (20,971,520,000 bytes total). The selected row has 100 ordinary
Evidence links and version 2. Independent semantic queries check populations,
versions, projections, exact target/link bindings, available-object hashes,
default workbook preferences, absence of sessions and absence of access handles
or upload leases. The fixture's private credential parser is shared with the
large-grid adapter while preserving each profile's exact account set and count.

The catalog now enrolls `module.timeline.measurement.evidence_inspector_v2`
and `module.timeline.frontend.ac045_paint_qualification`. The browser workload
uses the ordinary editor session and verifies complete preview bodies for all
100 linked records before timing. The measured action uses a product acceptance
mark, one fresh bounded metadata request, exact target and first-five item
bindings, and two consecutive visible frames. It performs one discarded warm-up
and 100 measured openings without action retries. Trace, video and automatic
screenshots are disabled to exclude private preview capabilities and bodies.
The closed Evidence inspector proof is validated during collection and retained
evidence review. Synthetic record identities are admitted only in its target
and five item bindings; unrelated sensitive fields still fail. Multi-profile
graph tests now prove each profile has exactly one appropriately bound builder,
without relying on array position or a global one-builder assumption.

These changes establish implementation and detector coverage. They do not yet
establish the full physical fixture, browser latency or Core 05 publication.
The smaller integration fixture deliberately uses 100 physical 256-byte objects
and retains the actual 5,000 Timeline rows and 100 links. Its result is not a
substitute for the full workload.

All roots below are beneath `.cartulary/test-results/`. Commands ran from the
repository root with the named `CARTULARY_TEST_RUN_ID`; PASS rows have outer exit
0 except the interrupted-session generation observation explicitly identified.

| Run ID | Command / scope | Result |
| --- | --- | --- |
| `remediation-evidence-heavy-integration-20261014-01` | `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.evidence_inspector_fixture` | PASS, 118.271 s; source `sha256:2918209450c3afa1baf28c348d07a15e6e27e69e5cd576cfcc7ac0fa815cc227` |
| `remediation-evidence-heavy-integration-20261014-02` | Same row, including rejection of altered default preferences | PASS, 119.333 s; source `sha256:cd08032e9ca88cfe25a5036295cdc18a0bd428b19aa231662851737b2372f76f` |
| `remediation-evidence-measurement-assembler-20261014-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.source_owner_contribution_assembler` | PASS, 26.557 s |
| `remediation-evidence-measurement-unit-20261014-02` | `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.ac045_paint_qualification,module.timeline.frontend.ac043_snapshot_supported_envelope_shape_cbe446f6e6,module.timeline.frontend.source_ac043snapshot_test_ts_f08ef1d7fa` | PASS, 4.012 s; detector rejection and both private runtime profiles |
| `remediation-evidence-measurement-routing-20261014-04` | `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | PASS, 26.688 s; profile routing, complete proof, adversarial proof and retained-artifact checks |
| `remediation-evidence-measurement-generate-20261014-04` | `make generate` | Retained tool summary PASS, 9.110 s, recorded exit 0; original outer process observation unavailable after interruption |
| `remediation-evidence-measurement-finalize-20261014-03` | `make agent-finalize` | PASS, 18.211 s; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `remediation-evidence-measurement-lint-20261014-02` | `make lint` | PASS, 20.772 s |
| `remediation-evidence-measurement-drift-20261014-02` | `make generate-drift` | PASS, 12.751 s |
| `remediation-evidence-measurement-shape-20261014-02` | `make json-shape-check` | PASS, 4.979 s |
| `remediation-evidence-measurement-policy-20261014-01` | `make generated-artifact-policy-check` | PASS, 0.814 s |

The earlier failures remain retained. `remediation-evidence-measurement-generate-20261014-01`
rejected the missing profile cleanup policy; `-02` rejected unsorted frontend
titles (`harness/unknown_failure`). Generation `-03` passed after correcting
authored inputs. Routing `-01` rejected the stale global one-builder assertion;
routing `-03` rejected a retained-proof test helper's large-grid object-count
assumption (`product/test_assertion_failure`, all 11 cleanup operations complete).
Routing `-02` and `-04` passed after their respective corrections. Lint `-01`
reported a missing package-root contract export, nullable metadata typing and
non-null assertion warnings; those were corrected before lint `-02` passed.
Shape `-01` and drift `-01` rejected the stale topology render index after the
redaction implementation changed. Explicit generation `-04` and their `-02`
checks resolved that drift. These are deterministic implementation/test-input
errors, not evidence of intermittent product behavior.

### Full physical workload and preparation corrections

`remediation-ac045-measurement-v2-20261014-01` failed before browser execution
in 188.763 s, outer exit 2, `product/test_assertion_failure`. All 23 cleanup
operations completed. The Evidence contribution failed during a physical write;
the retained adapter error says endpoint unreachable, but construction uses a
local rooted filesystem. The default `/tmp` is a 16 GiB tmpfs on this host,
smaller than the 20,971,520,000-byte population. The wrapped diagnostic does not
preserve errno, so a network failure is not established. The missing measurement
summary is a downstream artifact failure, not another product failure.

The builder now checks user-available filesystem capacity before cloning the
populated database or calling source-owner contributions. It rejects unavailable
capacity and populations exceeding that lower bound without private path
disclosure. Block rounding avoids overflow; zero-object fixtures remain valid.
Capacity is not a reservation or a guarantee against concurrent writers. Linux
is the supported object-fixture host; the non-Linux helper rejects populated
objects explicitly. The Linux test entrypoint is enrolled in
`harness.browser.unit.performance_fixture_objects`. No sparse objects, smaller
population or automatic alternate root is admitted.

The full run uses the existing external scratch setting explicitly:
`CARTULARY_HARNESS_SCRATCH_ROOT=/home/jochi/.cache/cartulary/harness-scratch`.
This host's disk has sufficient capacity; private permissions, ownership,
containment and suite cleanup still apply. The setting must be recorded and
matched for comparable future windows.

`remediation-ac045-measurement-v2-20261014-02` used that setting and successfully
sealed the exact physical population. Source:
`sha256:19dd475f65b4a4d467a71033957f28e56d4b04580870528d2eb5790a4e8bbfe8`.
Construction took 298.869 s, including Evidence 288.839 s, Timeline 1.275 s,
Auth 65 ms, Incidents 13 ms, Projections 95 ms and semantic validation 41 ms.
These contribution diagnostics do not allocate all remaining builder work to
individual tests. The builder's sealed artifact contains the exact physical
object proof; no browser latency follows from it.

Clone preparation then failed before service startup. The run finished in
439.859 s, outer exit 2, `infra/service_start_error`, with all 24 cleanup
operations complete. The retained failed clone reports database and bucket
cleanup complete and no acquired credential copy, process or session. Its
creation and overall lease state remain failed. The failure occurred about
126 seconds after attachment; inspection found `prepare-web-e2e` imposed the
ordinary two-minute migrated-template budget on a full object restore. The
old startup diagnostic pointed to a non-retained private log, so its underlying
error cannot be recovered from this run. Deadline exhaustion is the supported
cause hypothesis, not a claim that the unavailable log was inspected.

Object-backed clone preparation now uses the same bounded 15-minute
administrative budget as physical snapshot construction. Empty-object
preparation retains two minutes. Unknown profiles reject. Wrapped or observed
deadline exhaustion propagates exit 13 and `timing/timeout_failure` through
startup diagnostics; configuration rejection propagates exit 2. Diagnostics
no longer direct reviewers to discarded private logs. This does not change
product timing thresholds, retries or sampling. The dedicated
`harness.browser.boundary_support.preparation_contract` row runs the existing
adapter's failure paths through the new internal
`harness-browser-preparation-contract` command. It verifies that failed
preparation never reads the unpublished environment or admits dependent state.
The public-target roster is unchanged. Legacy smoke-wrapper assertions were
not used as proof of this new route.

| Run ID | Command / scope | Result |
| --- | --- | --- |
| `remediation-fixture-capacity-generate-20261014-01` | `make generate` | PASS, 9.020 s |
| `remediation-fixture-capacity-format-20261014-02` | `make format` | PASS, 4.775 s |
| `remediation-fixture-capacity-unit-20261014-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.performance_fixture_objects,harness.browser.unit.performance_fixture_lifecycle` | PASS, 1.957 s |
| `remediation-fixture-capacity-finalize-20261014-01` | `make agent-finalize` | PASS, 17.538 s; maintenance skipped because `RESULTS_DIR` was unset |
| `remediation-fixture-capacity-lint-20261014-01` | `make lint` | PASS, 25.326 s |
| `remediation-fixture-capacity-drift-20261014-01` | `make generate-drift` | PASS, 17.328 s |
| `remediation-fixture-capacity-markdown-20261014-01` | `make lint-markdown` | PASS, 10.820 s |
| `remediation-fixture-preparation-generate-20261014-02` | `make generate` | PASS, 9.070 s |
| `remediation-fixture-preparation-format-20261014-02` | `make format` | PASS, 4.823 s |
| `remediation-fixture-preparation-unit-20261014-02` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.testservices_lifecycle_contract` | PASS, 1.548 s |
| `remediation-fixture-preparation-shell-20261014-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.preparation_contract` | PASS, 1.008 s |

All PASS rows above have outer exit 0. Capacity format `-01` rejected an unsorted
new selector before retained execution (`config/configuration_error`, outer
exit 2, no run summary); the authored selector was sorted and regenerated.
Preparation unit `-01` failed because its fake active-service lease and requested
suite had different identities (`product/test_assertion_failure`, 2.378 s,
outer exit 2, all 10 cleanup operations complete). Correcting the fixture to
the declared lease identity made `-02` pass; authorization was not bypassed.
Neither failure establishes intermittent behavior. Full clone/browser
qualification remains open until a fresh run with these corrections passes.


### Full AC-045 restore and browser investigation

The preparation corrections passed `remediation-fixture-preparation-finalize-20261014-01`
(18.234 s; maintenance skipped because `RESULTS_DIR` was unset), followed by
`remediation-fixture-preparation-lint-20261014-01` (27.279 s),
`remediation-fixture-preparation-routing-20261014-01` (34.092 s),
`remediation-fixture-preparation-drift-20261014-01` (19.133 s), and
`remediation-fixture-preparation-markdown-20261014-01` (12.280 s). All outer
exits were 0.

`CARTULARY_HARNESS_SCRATCH_ROOT=/home/jochi/.cache/cartulary/harness-scratch
CARTULARY_TEST_RUN_ID=remediation-ac045-measurement-v2-20261014-03 make
service-backed-test-slice OWNER=module.timeline
ROWS=module.timeline.measurement.evidence_inspector_v2` completed in 540.235 s
with outer exit 2. Source was
`sha256:b15f88cece2976505aff43dbada420641ffa0cf03e5b78666949a2140d9a0525`;
graph was `sha256:e8ccbc1d0780f9dadb8ab147aa7e46f1c2068049ae5eb88996b950eaaa84fac0`.
The full physical snapshot sealed in 301.079 s and restored successfully:
the v3 lease records `created`, `isolated`, 10,000 objects and 20,971,520,000
bytes. Every lease cleanup resource and all 26 run cleanup steps completed.
This is positive full-population restore evidence, not AC-045 timing closure.

The browser row failed `product/test_assertion_failure` on its first open:
the observer rejected a contaminated measured request set. No timing sample
completed. The aggregate was `artifact/artifact_error` because the observation
attachment lacked the predicate suffix required by the existing collector.
The attachment name now follows that interface. Incomplete observations no
longer assert a completed predicate proof. The observer now uses the normative
clipped-viewport intersection rule rather than full rectangle containment,
and its request failure gives bounded metadata/access counts without URLs,
capabilities or payloads. Preparation explicitly awaits its metadata response
completion. These changes do not relax the product threshold or permit retries.

The existing `module.evidence.browser.linked_metadata_window` scenario now checks
exact request cardinality for opening, paging and reopening, fresh metadata
rather than retained window content, and absence of access-handle requests.
`remediation-evidence-reopen-browser-20261014-01` passed in 25.976 s, outer exit 0.
It uses ordinary real services and seven requested Evidence records; it is a
focused request-lifecycle regression, not a substitute for the heavy fixture.
`remediation-evidence-observer-unit-20261014-01` passed in 3.614 s, outer exit 0,
including a partially clipped but intersecting element and the existing
wrong-binding, hidden/offscreen, one-frame, repeated-action and contaminated-read
negative cases. The full-workload request failure remains unclassified pending
its bounded diagnostics; the smaller pass alone does not establish instability.

The UI/UX refactor guidance applies to the inspected inspector boundary under
Core 04 REQ-157–159 and the existing Evidence/Workbook owners. This slice changes
only test choreography, observer semantics and evidence serialization; product
presentation, tokens, density, authorization and mutation semantics are unchanged.
Relevant acceptance concerns are exact ownership, semantic selectors, inspector
read lifetime and truthful evidence. Rendered visual/a11y and product-refactor
acceptance are not claimed by these detector checks.

### Construction publications and AC-045 contamination

The full AC-045 run `remediation-ac045-measurement-v2-20261014-04` failed
in 540.906 s with outer exit 2, `product/test_assertion_failure`, 12 passed
and two failed work units, and all 27 cleanup operations complete. The full
physical build and restore passed. Its first open observed two metadata reads
and zero access-handle reads; no timed sample completed. The corrected attachment
was collected. Source was
`sha256:876667b97cd9b0bc6ba700cf128ce21510f9ee0abadb2e36494361d23d210baf`;
graph was `sha256:e8ccbc1d0780f9dadb8ab147aa7e46f1c2068049ae5eb88996b950eaaa84fac0`.
This failure does not establish an intermittent product defect.

The focused Evidence row now also exercises 100 genuinely available linked
Evidence records, an ordinary editor, the same filter and geometry, and exact
metadata/access request cardinality across close and reopen. The objects are
small functional fixtures, not the AC-045 physical workload. Both selected
scenarios passed at `remediation-evidence-reopen-browser-20261014-03` in
34.858 s, outer exit 0. The earlier admin variant passed in `-02` (35.939 s).
Its first two generation attempts rejected a missing scenario identity and
unsorted identities respectively (`harness/unknown_failure`, outer exit 2);
`remediation-evidence-reopen-generate-20261014-03` passed after authored routing
was corrected. No failed result was replaced.

Inspection found that ordinary Evidence and Timeline construction leaves durable
Collaboration intents pending while the fixture server is not running. The
production dispatcher starts its live tail at the sequenced high-water mark;
sequencing those old intents after restore therefore creates live construction
traffic. A final Collaboration-owned fixture contribution now sequences pending
construction intents through the normal durable stream before sealing. It does
not delete history, synthesize delivery, bypass validation/backoff, or suppress
production events. Independent semantic validation requires no pending intent,
no quarantined stream, and exact replay bindings. The changed contribution and
proof invalidate older snapshot keys. The product workload and predicate remain
v2: this repairs construction quiescence, without changing target population,
visible completion, sampling or threshold. Whether this caused the observed
extra browser request remains a hypothesis until full-workload qualification.

`remediation-fixture-publications-unit-20261014-01` passed the source-owner
assembler row in 5.580 s. `remediation-fixture-publications-integration-20261014-01`
passed the real Evidence inspector fixture row in 121.275 s. Both outer exits
were 0. The integration regression first rejects pending construction history,
then proves exact retained intent/replay counts and high-water state, repeated
finalization without duplicate replay, cancellation, stalled sequencing,
quarantine rejection, and no mutation of unrelated pending publications.
Generation `remediation-fixture-publications-generate-20261014-03` passed in
8.930 s; formatting `remediation-fixture-publications-format-20261014-01`
passed in 4.900 s. Full-workload browser qualification follows separately.

Publication finalizer `remediation-fixture-publications-finalize-20261014-01`
failed in 4.508 s (`harness/child_target_failure`, outer exit 2): the new
Collaboration support root was not yet in the authored support inventory.
After explicit admission and generation `-04` (9.030 s), finalizer `-02`
passed in 18.006 s. Retained-run maintenance was skipped because `RESULTS_DIR`
was unset. The publication-slice lint, routing, generation-drift and Markdown
roots ending `-20261014-01` passed in 24.257, 31.502, 15.558 and 26.100 s
respectively; all outer exits were 0.

The full run `remediation-ac045-measurement-v2-20261014-05` failed in
568.215 s (`product/test_assertion_failure`, outer exit 2), with 12 passed and
two failed units and all 27 cleanup operations complete. Source was
`sha256:e0ae2fdcc85049b1e4cc91ba336c8e63457d9be254c0c101c4da72381ca2a65a`;
graph was `sha256:da8f84bd2aac03ea996ed132855f5818a7d7f46587a3c6dd4d9db08a0775b857`.
Snapshot `5dca49a1ee203b1fe6abc34101685e1866eb629050a4b7a9035cc57ba4dff077`
sealed in 314.394 s, including 18.161 s of ordinary Collaboration sequencing.
Its independent semantic check and full clone passed. The contaminated-read
failure did not recur; the first open instead failed to satisfy visible
completion, with zero completed samples. This is not a performance pass.

The small 100-link browser regression now also calibrates the exact two-frame
observer against real rendering using a deliberately pre-dispatch detector
marker. That marker cannot qualify product latency, and this regression emits
no measurement observation. Bounded detector diagnostics identify condition,
part ordinal and numeric clipping geometry without content or capabilities.
Detector browser `-01` reproduced an offscreen attached-count element (49.199 s);
`-02` proved that the Evidence section was no longer selected after reopening
(41.750 s). Both failed with outer exit 2 and `product/test_assertion_failure`,
and all 22 cleanup steps completed in each. Unit checks `-01` and `-02` passed
in 3.842 and 3.637 s.

The cause is test choreography: the row-menu Inspect command explicitly opens
Details, which correctly takes precedence over retained section selection under
Design §12.7. The AC-045 prepared state requires Evidence. Measurement and the
new calibration regression now reopen the same selected subject through the
ordinary workbook inspector toggle, preserving Evidence without test-driven
scrolling or section changes during the interval. Core 04 clarifies that start
state; the v2 predicate, workload, threshold, sampling and product behavior are
unchanged. `remediation-evidence-detector-browser-20261014-03` passed both
selected scenarios and the real-rendering detector in 34.831 s, outer exit 0.
No product layout change, scenario removal or retry-to-green policy was used.


## AC-045 qualified full workload

After resuming, finalizer `remediation-evidence-detector-finalize-20261014-01`
was confirmed to have published successful finalization and retained-secret
results; its original native exit was unavailable after session loss. Lint
`remediation-evidence-detector-lint-20261014-01` failed in 21.016 s
(`harness/child_target_failure`, outer exit 2) because the toggle correction
left two unused imports in `evidence.spec.ts`. Removing those imports required
no behavior change. Finalizer `-02` passed in 18.204 s and lint `-02` passed
in 20.814 s, both with outer exit 0. Retained-run maintenance was skipped
because `RESULTS_DIR` was unset. Markdown
`remediation-evidence-detector-markdown-20261014-01` passed in 22.320 s.

`remediation-ac045-measurement-v2-20261014-06` passed through
`make service-backed-test-slice OWNER=module.timeline
ROWS=module.timeline.measurement.evidence_inspector_v2`, using
`CARTULARY_HARNESS_SCRATCH_ROOT=/home/jochi/.cache/cartulary/harness-scratch`.
The exact run root is `.cartulary/test-results/` followed by that run ID.
The outer exit was 0, duration 600.980 s, all 14 work units passed, and all
27 cleanup steps completed. The retained-secret scan passed. Source:
`sha256:109cec88b4f539f3888c3027e7caf35444245846273cf9bf762bc30c04ec5f72`;
graph: `sha256:da8f84bd2aac03ea996ed132855f5818a7d7f46587a3c6dd4d9db08a0775b857`.

The current physical snapshot generation and full 10,000-object / 20 GiB
population qualified. After one discarded warm-up, all 100 measured opens
completed with fresh bounded metadata, exact bindings, no access-handle
requests in the measured interval, and two visible frames. P50 was
45.100000001490116 ms; p95 was 57.100000001490116 ms, below the 300 ms
limit. The finalized v5 summary reports `qualified` and zero scheduler overlap.
This establishes the AC-045 v2 implementation predicate for this source state.
It is not Core 05 publication qualification or a clean-source harness-cost
window. Earlier failed roots remain retained; this pass does not establish a
failure rate or prove the whole subsystem stable.


## Benchmark artifact-read boundary and routing

The benchmark claim checker now uses the existing secure retained-artifact
reader for its registry, selected manifest and bundle. The actual retained
bundle bytes must hash to the declared SHA-256. The metadata-only hash bypass
and repository-relative fallback are removed. Manifest-relative bundle paths
are normalized and contained; absolute paths, traversal, symlink ancestors,
leaf symlinks, hard links, non-regular files, oversize inputs and changing files
are rejected. Documentation and Markdown paths are rejected before filesystem
access. Strict JSON parsing rejects duplicate members. Core 05 REQ-05-009
records these evidence-admission requirements. This is a correction to artifact
proof, not a new measurement predicate or sampling interpretation.

The previous `harness-smoke-benchmark-claim-check` definition and tier references
did not provide a current catalog route. Its complete script now survives in
`harness.release.behavior.benchmark_claim`, using the dedicated existing-runner
command `cartulary.harness.command.harness_benchmark_claim_contract.v1` and
verification `harness.release.verification.benchmark_claim`. All existing
profile, sample, security and hash regressions remain in the same script; the
new cases add path and alias rejection, metadata-only rejection, duplicate JSON,
and a filesystem-open spy proving no traversal of forbidden documentation paths.
The stale smoke entry is removed rather than aliased. Behavior and execution
review: Codex under the approved remediation scope; the same test script,
no-service fixture and shell runner survive, with reachable catalog evidence.

`remediation-benchmark-boundary-generate-20261014-01` passed in 9.670 s and
`remediation-benchmark-boundary-contract-20261014-01` passed the exact new
owner row in 1.060 s, both outer exit 0. The profile/sampling/publication
protocol still needs reconciliation with the current Core 04 typed predicates;
this security slice alone does not close Core 05 or authorize a benchmark claim.


Boundary finalizer `remediation-benchmark-boundary-finalize-20261014-01`
failed in 2.093 s (`harness/child_target_failure`, outer exit 2), with all ten
cleanup steps complete: the static input guard interpreted literal negative
`docs/` fixture paths as documentation dependencies. The tests now construct
those hostile tokens as the existing secure-reader tests do; they still assert
zero filesystem opens. Finalizer `-02` passed in 17.883 s. Retained-run
maintenance was skipped because `RESULTS_DIR` was unset. The exact benchmark
row `remediation-benchmark-boundary-contract-20261014-02` passed in 1.039 s;
full lint `remediation-benchmark-boundary-lint-20261014-01` passed in 22.905 s;
routing `remediation-benchmark-boundary-routing-20261014-01` passed in 27.268 s.
All three native outer exits and the second finalizer exit were 0.

Further publication-protocol reconciliation must also admit `BENCHMARK_MANIFEST`
in the public target's typed input contract. The Make recipe currently forwards
it and the root Makefile supplies a default, but the target contract does not
expose that selector. No publication profile or predicate completeness claim is
made from the boundary-only pass.


The missing benchmark selector is now explicitly admitted in the Harness
normative input matrix and task-surface owner: optional `BENCHMARK_MANIFEST`,
path type, default `.cartulary/benchmark/benchmark_manifest.json`, empty value
rejected, forwarded as argv. The existing no-claim default-absence behavior is
unchanged. The benchmark regression exercises the real public-input resolver
for default and explicit values, empty rejection, and rejection on unrelated
`lint`. Generation `remediation-benchmark-selector-generate-20261014-01`
passed in 8.820 s and the exact benchmark row
`remediation-benchmark-selector-contract-20261014-01` passed in 1.049 s,
both outer exit 0. The earlier boundary Markdown check
`remediation-benchmark-boundary-markdown-20261014-01` passed in 21.870 s.


Selector finalizer `remediation-benchmark-selector-finalize-20261014-01`
passed in 18.179 s, outer exit 0; retained-run maintenance was skipped because
`RESULTS_DIR` was unset. Real public invocation
`remediation-benchmark-selector-default-20261014-01` passed in 117 ms with the
explicit no-claim diagnostic for the absent default manifest. It is not
publication evidence. Real explicit selection
`remediation-benchmark-selector-missing-20261014-01` intentionally selected
`BENCHMARK_MANIFEST=tmp/remediation-benchmark-deliberately-absent.json` and
rejected in 125 ms, outer exit 2, `harness/child_target_failure`, with the
`benchmark manifest missing` diagnostic and all ten cleanup steps complete.
This proves the admitted selector reaches the validator without changing the
existing missing-non-default failure contract.


## Current benchmark publication protocol

Core 05 now owns a closed v2 publication envelope and recursively digested v1
bundle index. The single current profile is desktop reference v3; retired
profile registries and flat manifests have no compatibility reader. The browser
and Playwright versions follow the installed pin, while criterion-owned
fixtures, traffic, geometry, warm-up and sampling no longer inherit global
defaults. Current predicate membership comes from the existing typed product
contracts and active fixture/catalog bindings, not another execution catalog.

The reader validates actual retained bytes, current physical fixture receipts,
run/source/selection identity, fresh passing outcomes, complete predicate sets,
raw browser results, sample statistics, cleanup and actual browser receipts.
Single-observation AC-003 remains distinct from percentile criteria. Browser
build is bound to the managed renderer attestation; mode is observed from the
browser-level Chromium version response without retaining raw user-agent data.
Hardware/deployment qualification and implementation-conformance review remain
separate publication prerequisites. No Core 05 claim has been produced.

Generation `remediation-benchmark-v2-generate-20261014-01` passed in 10.270 s.
Contract runs `-01` and `-02` failed: the first incorrectly requested a schema
attachment for the separately owned pin registry; the second supplied an
incomplete failed-cleanup fixture. Both defects are corrected without weakening
admission. The exact row passed in `remediation-benchmark-v2-contract-20261014-03`
(26.123 s, outer exit 0). The native exit of `-02` was lost across session
recovery; its retained product/test_assertion_failure and all ten cleanup
results are available. Format passed in `remediation-benchmark-v2-format-20261014-01`
(4.868 s). Finalizer `-01` rejected stale generated topology inputs after the
shared reader edit; generation `-02` passed (9.440 s), then finalizer `-02`
passed (18.297 s). Retained-run maintenance was skipped because `RESULTS_DIR`
was unset. Further source changes and browser receipt validation still require
fresh checks. Selector Markdown lint `remediation-benchmark-selector-markdown-20261014-01`
also passed (8.330 s).


The first real receipt run, `remediation-benchmark-v2-paste-20261014-01`, failed
before product action (133.420 s; outer exit 2; artifact/artifact_error, with
underlying Playwright test_assertion_failure). Chromium headless shell does not
support `Browser.getBrowserCommandLine` without an automation switch. All 26
acquired cleanup operations completed. No paste measurement was produced.
The receipt now checks the pinned managed renderer and browser-level
`Browser.getVersion` identity; it does not alter browser launch arguments or
trust a local fixture default for a remote browser. The Chromium source's
[headless product identity](https://chromium.googlesource.com/chromium/src/+/main/headless/lib/browser/headless_browser_impl.cc)
informed this implementation; no executable input reads that reference.
Finalizer `-03` passed (17.909 s), routing `-01` passed (30.102 s), and lint
`-01` passed (22.401 s) before this runtime correction.


Fresh browser receipt validation passed in
`remediation-benchmark-v2-paste-20261014-02` through the exact service-backed
AC-003 row: 148.335 s, outer exit 0, all 14 units passed, all 27 cleanup
operations completed, and retained-secret scanning passed (86 files). The
single observation qualified at 352.5 ms with zero scheduler overlap.
Source: `sha256:fc464a302bd0e4627cb1bc79760aa5cfae2828aa46f2c910e3eb7ef4c72a61d7`.
Graph: `sha256:99d850dba4bbd5b360c050c0dbe168af313926e0dc6ac3c38b7bc68d35fdb0d8`.
The retained and inline browser receipt reports Chromium 153.0.8010.12,
Playwright 1.63.0, Node v24.15.0, Linux/x64 and `browser_mode=headless`.
This run cannot satisfy the headed Core 05 publication profile. It is a fresh
implementation-predicate observation with accurate execution context.

Finalizer `remediation-benchmark-v2-finalize-20261014-04` passed (18.891 s),
with retained-run maintenance skipped because `RESULTS_DIR` was unset. Lint
`remediation-benchmark-v2-lint-20261014-02` passed (22.363 s), and Markdown lint
`remediation-benchmark-v2-markdown-20261014-01` passed (23.700 s), before this
annotation. No publication claim or qualified execution-cost portfolio is
inferred from these results.


## Presence criterion binding correction

Core 04 already assigns the same three Timeline presence predicates to AC-008
and AC-132. The typed presence contract now has a closed v2 shape with both
criterion IDs. The existing predicates, workload, timing and observation label
remain unchanged. Publication derives many-to-many criterion membership from
that owner and requires the complete predicate set once, even when both
criteria are declared. It cannot infer AC-132's other functional obligations.
The single-criterion v1 contract shape is retired; no translation or duplicate
browser execution is introduced. Generation
`remediation-presence-binding-generate-20261014-01` passed (9.090 s). Focused
validation and current physical-fixture presence measurements are pending.


Presence binding validation passed through the exact public owner routes:
`remediation-presence-binding-contract-20261014-01` (33.748 s),
`remediation-presence-binding-routing-20261014-01` (30.959 s), and
`remediation-presence-binding-lint-20261014-01` (22.371 s), all outer exit 0.
Finalizer `remediation-presence-binding-finalize-20261014-01` passed (18.415 s);
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

The exact three-row service-backed run
`remediation-presence-binding-browser-20261014-01` passed (205.828 s, outer
exit 0): 18/18 units, 35 completed cleanup operations, one shared physical
snapshot and three isolated clones, zero scheduler overlap. The qualified
single observations were surface 46.100000001490116 ms, row
179.99985351413488 ms, and cell 205.79995116591454 ms, all <=1,000 ms.
Source: `sha256:32bf06bcb19c3e3182497e47f7fa67ddc4cdaec2c6d5642dd3fadaa4261c47be`.
Graph: `sha256:14a7c11704839f544c15830dc4931c168a23b6bdddb708f8c62bd939f8a3e378`.
Snapshot: `856723721e445931dd81d7d9d2c9ea4793c9f13088765088d4a5d8c7ecb5e2a7`.
This supplies the timed presence predicates for AC-008/AC-132 at the retained
source. It does not replace other AC-132 functional contexts or qualify a
headed Core 05 publication or clean-source execution-cost claim.


## Exact browser execution identity

A final publication-reader review found a shared execution gap: browser command
construction still used unanchored substring grep, and result accounting joined
only file/leaf title, omitting project and suite path. The shared parser now
retains suite components, rejects unrepresentable exact-selection names and
suite-prefix collisions during configuration preflight, and derives the pinned
Playwright test-list input directly from existing owner rows. No additional
catalog is introduced. Result accounting joins project/file/full suite path and
rejects unexpected executed identities. Publication reuses that same adapter.
Generation `remediation-playwright-identity-generate-20261014-01` passed
(9.210 s), and the initial focused adapter run passed (1.784 s). Additional
pinned-runner, preflight and publication checks remain pending after adding
nested/tagged/sub-string collision regressions.


The pinned Playwright invocation regression passed in adapter run `-02`
(2.396 s): only the selected nested tagged test ran, while same-leaf and
suffix-named tests remained unexecuted. Generation `-02` passed (9.280 s)
and finalizer `-01` passed (18.576 s); retained-run maintenance was skipped
because `RESULTS_DIR` was unset. Lint `-01` passed (22.967 s). Retained
routing and publication `-01` summaries report passing fresh execution
(31.184 s and 36.521 s respectively); their native exits were unavailable
after session recovery. Those summaries do not establish the lost outer exits.

Real browser run `remediation-playwright-identity-browser-20261014-01` passed
through the exact AC-003 service-backed row (147.430 s, native exit 0):
14/14 units, 27 completed cleanup operations, retained-secret scan passed
(87 files). The retained test list selects exactly the Chromium project,
measurement file and full AC-003 title. The single observation qualified at
383.69999999552965 ms with zero scheduler overlap.
Source: `sha256:24a60c393126f9f440539799f95552b2a678990a314ba4beba172e25d63370f7`.
Graph: `sha256:0a0380c5defbbbdc1b68164fd963cc448015256806e3acffe8f166a2f8f5677d`.
The source was dirty; this is scoped engineering evidence, not a clean-source
execution-cost window or headed Core 05 publication claim.

## Reporting source and frozen-field boundaries after restart

The closed source-query and frozen-field slices are implemented and validated
for source `sha256:af76b0af7874e14801895eccd7797fa58a70ece999a68b43b6ebf5f2e396e5e5`.
Nine source tables now contribute explicitly selected fields; adding a private
storage column does not expand snapshots or releases. Artifact fields are
limited to their own family. Canonical fields reject invalid state/value pairs,
raw digests, null reference arrays and inconsistent section/record/Timeline
copies. Reporting coalesces repeated source support targets during
materialization while preserving the Links source contract. Aggregation owns
its disclosure-label storage, preventing later sorting from relabeling fields.

Full Reporting, Composition, selected Incident Bundles portability, lint,
Markdown, generated drift/policy, JSON shape and both explicit-root owner audits
passed. Exact commands, source/graph identities, run roots, cleanup and the two
repaired failure mechanisms are retained in the
[Reporting adoption review](reporting-cutover-adoption-review.md#frozen-field-copy-admission).
Finalization passed; retained-run maintenance was skipped because `RESULTS_DIR`
was unset. This is implementation evidence for these boundaries. Reporting's
full canonical model, Timeline sort metadata, template/deck derivation, actual
renderer/sandbox and required adoption evidence remain open, alongside the
other outstanding workstreams above.

## Timeline ordering remediation after restart

Reporting now receives source-owned Timeline ordering facts through a versioned
provider contract and retains them in the exact event shape. Snapshot/replay
admission rejects missing or inconsistent facts, altered sort keys, wrong order,
omitted required-null members and retired shapes. Source timestamp interpretation
stays with Timeline; display selection/redaction stays with Reporting. The
[adoption review](reporting-cutover-adoption-review.md#timeline-source-ordering-and-retained-event-admission)
records exact clauses, source identities, commands, owner audits, retained
compile/import-boundary failures and subsequent passes. Reporting, Composition
and selected artifact portability passed; final lint passed after admitting the
two exact Timeline-owned/parser dependencies. Both finalizations skipped
retained-run maintenance because `RESULTS_DIR` was unset. Reporting adoption,
render/display completion and overall nine-finding assurance remain open.

The subsequent Reporting memory-isolation slice also passed current-source
Reporting and Composition owner runs and evidence audits. Snapshot, sibling
render, preview identity and caller payload bytes are now isolated from nested
mutations across render/participant boundaries. The adoption review records the
regression, exact roots and an evidence-source mismatch that was resolved with
fresh owner execution. Overall remediation and coordinated adoption remain open.

External composition-authored text now honors the selected profile's default
denial before any render artifact is published. Unit coverage includes all three
roles and unused authored nodes; a real job regression compares internal success
with external denial for the same immutable composition and a valid recipient.
Fresh Reporting/Composition owner runs, lint and both evidence audits passed.
The [permission-boundary record](reporting-cutover-adoption-review.md#external-authored-text-permission-boundary)
retains exact commands, source identity and the interrupted finalizer's native
exit limitation. Opt-in text redaction/partitioning and overall adoption remain
open; finalization skipped retained-run maintenance with `RESULTS_DIR` unset.

The next coordinated specification slice resolves Core/Reporting's conflicting
content-class vocabularies. Source classifications remain unchanged; obsolete
Reporting aliases are retired, generated categories cannot be source input,
and v2 section/block summaries count actual blocks. Current Reporting,
Composition, portability, lint and both owner audits passed. The
[classification record](reporting-cutover-adoption-review.md#source-classification-and-section-summary-reconciliation)
retains exact source and run identities plus the remaining model/adoption limits.

Native snapshot readiness now checks original admission content and exact
historical Reference Pack bindings in one read-only transaction. Focused
Reporting and Reference Data backup/restore checks, the Reporting owner, lint
and its evidence audit passed. The full server owner failed during shared
object-store readiness before affected tests executed; that assurance remains
open. The [native-readiness record](reporting-cutover-adoption-review.md#native-snapshot-readiness-validation)
retains the failure, cleanup distinctions, interrupted finalizer exit limitation
and outstanding recovery/adoption boundaries.

The [closed snapshot Job admission slice](reporting-cutover-adoption-review.md#closed-snapshot-job-admission)
adds strict retained-input and independent Jobs binding checks, delegated source
boundary validation, and successful-tombstone compatibility. Reporting owner,
lint, generated-output gates, its evidence audit and a narrow server restart
case pass. Composition's two service-backed units encountered the same shared
object-store readiness symptom as the earlier full server run; its full owner
assurance and the infrastructure cause remain open. Failed runs and corrected
test-fixture mistakes are retained separately from successful evidence.

Reporting private reveal maps are now separate from shareable release/preview
bundles, with catalog v2 portability, clean-cutover migration 69, native private
readiness validation and a fresh encrypted backup/restore rehearsal. Full
Reporting, Composition and server owner suites and their evidence audits pass
for source `sha256:e09864a795a0e52639da53821e6689308cbdc71b16f0d3df88574df19f7553e5`.
See [private artifact closure evidence](reporting-cutover-adoption-review.md#private-reveal-map-separation-and-recovery)
for exact runs, preserved failures, sensitive retention and remaining limits.
This advances Gap 9 implementation and assurance without adopting the proposed
specifications or closing renderer, model, timed-job, performance-portfolio or
independent-review obligations. Earlier object-store incidents remain unattributed.

## Native release binding follow-up

The [Reporting adoption review](reporting-cutover-adoption-review.md#native-release-completion-and-recovery)
now records release completion/recovery binding checks, Composition-owned pin
validation, adversarial and expired-history tests, the extended restore rehearsal,
and passing full Reporting, Composition and server owner audits. Source digest:
`sha256:f5420061d9bfee0e4b8451663d760de19126a7ae55e0e7509c66fec72f67fb1a`.
The Reporting run reused one compatible cached unit; other full-owner units in
that record executed freshly. Overall nine-finding closure and coordinated
adoption remain incomplete. Preview/pending-Job recovery, profile admission,
actual renderer/sandbox conformance and qualified cost windows remain open.

## Frozen release profile follow-up

The [frozen-profile admission record](reporting-cutover-adoption-review.md#frozen-release-profile-admission)
closes the failed-release sentinel debt with exact profile bytes and an
independent retained digest, release Job v3, and clean-cutover migration 70.
Full Reporting, Composition and server owners, owner evidence audits, production
DDL, generated contracts and lint pass on source
`sha256:e2906d130587935b796ffb7374bd169fa5de0f835fe6eadc8d4803bbdea8f497`.
The initial shared-idempotency digest mismatch is retained as a diagnosed design
failure. The narrow preview run's process exit was lost after restart; the later
full Composition run provides fresh execution and observed successful exit.
Overall closure and coordinated adoption remain open; the record identifies the
remaining renderer, preview, model, worker, timing and review obligations.

The [worker execution-boundary follow-up](reporting-cutover-adoption-review.md#reporting-worker-execution-boundary)
fixes swallowed finalization errors, failed-observation continuation, shutdown
ownership and raw public failure text. Focused regression tests, full Reporting
and Composition owners, their evidence audits, real Jobs-supervisor tests and
lint/generated checks pass on source
`sha256:f19e113bf1b07dfa8295a87fe9ce6fc266339729ea70f46babab9a22954d02b2`.
It adds no retry policy and does not close the remaining preview or renderer
requirements.


### Frozen preview admission follow-through

The [Reporting adoption review](reporting-cutover-adoption-review.md#frozen-preview-admission-and-recovery)
records preview Job v3, migration 71, independently retained admission, completion
and recovery validation, the normalized-request defect and its repair, exact
owner runs and evidence audits. The preview integration retained after the desktop
restart has a passing summary and secret scan but no recovered native exit; the
subsequent full owner run has observed exit 0. This closes the exercised preview
input/output integrity slice, not Reporting adoption or all nine remediation
findings. Live-execution payload fencing and real rendering remain open.


### Live execution and unfinished-preview follow-through

The [Reporting review](reporting-cutover-adoption-review.md#live-execution-fencing-and-unfinished-preview-recovery)
records handler input fencing under the Jobs transition lock, current-time lease
validation, atomic Graph read/renewal, and real unfinished-preview recovery cases.
Full Reporting, Composition, Jobs and server owners and explicit evidence audits
passed on source `sha256:94436b76791692c1d80f1d760398a35f3eb56ab9c48228b47940631336d15080`.
New guard regressions have one active owner row; existing Graph and preview
scenarios retain their original operational identities. Earlier fixture/build
failures are retained with their classifications and cleanup results. This closes
these exercised mechanisms only; Reporting adoption and overall remediation
remain open, principally on actual rendering and remaining model/UI assurance.


### Standalone renderer follow-through

The [Reporting review](reporting-cutover-adoption-review.md#standalone-slidev-runtime-characterization)
records a pinned real Slidev/PDF adapter, click-page and determinism
characterization, browser request-denial tests, and explicit renderer execution
routing without product-service coupling. Full Reporting and test-catalog owners,
extended harness checks, lint, generated checks and owner evidence audits passed
on source `sha256:e624bafa95fab61b3fee178690fc04167523d4e433f6226f6164c7be668841d2`.
This is an intermediate implementation checkpoint: production still uses the old
renderer, and Mermaid rendering, OS sandbox enforcement, packaged runtime,
source/model fidelity and adoption remain open. Browser interception does not
satisfy native sandbox evidence. No required finding is declared closed solely
from this package's successful tests.

### Generated-text safety follow-through

The [Reporting text-boundary record](reporting-cutover-adoption-review.md#inert-generated-renderer-text)
records source escaping v2 and actual renderer tests proving that expression-like
text, entities, HTML and URLs remain inert while declared click behavior works.
Full Reporting and Composition owner suites and evidence audits, lint and
generated checks pass on source
`sha256:cb70ecc22b60cac1ba17c8853602d14818870bcfbf54b0ee96b6e8f54c991544`.
The restart-lost native lint exit is recorded separately from its passing summary;
a fresh lint invocation completed with observed exit 0. Production rendering,
OS sandboxing, full model fidelity and coordinated adoption remain open.

### Canonical source and template cutover follow-through

The [Reporting source-framing record](reporting-cutover-adoption-review.md#canonical-source-framing-and-template-metadata-cutover)
documents source v2, manifest v2, template `@2`, the proposed Core 01 companion,
migration 72 and exact-source owner evidence. Shared Go/Slidev parser vectors
detect extra cover/blank slides. The migration hash assertion now uses a fixed
synthetic protocol vector while retaining every production migration's manifest
parity check; its original failure and named surviving assertion are recorded.
Reporting, Composition, migration and server owner suites/audits, lint and
generated checks pass on
`sha256:afe57b34a0975542ae9de9f6104791f22ec5e841106df34a48a00a55714415c7`.
Compatible cache hits are disclosed. The full nine-finding remediation and
coordinated adoption remain incomplete.

### Logical support-reference follow-through

The [strict support-reference record](reporting-cutover-adoption-review.md#strict-logical-support-references)
records removal of invented unknown-target identities, exact selector validation,
incident binding and retained-model rejection. Full Reporting and Composition
owner suites and audits, lint and generated checks pass on source
`sha256:d6e0272ed562e31664de7abad20ba2bed3559daccd969fb63166eccdf7e2a527`.
The interrupted narrow run's missing native exit is disclosed separately.
General source/support reachability and schema fidelity remain open, along with
production renderer, timing, cost qualification and coordinated adoption work.

### Token-use and frozen-redaction follow-through

The [token integrity record](reporting-cutover-adoption-review.md#exact-token-use-and-frozen-redaction-binding)
binds private uses to selected rules and disclosure, and completed artifacts to
deterministic redaction of independently frozen admission. Recomputed digests no
longer conceal changed rules or invented private values. Full Reporting,
Composition and server suites/audits, lint and generated checks pass on source
`sha256:2d9845c4cafa2a994315e3c0c4bb6cc0bc97bb9f569dc5312651876505a15569`.
The separate token/subject schema discrepancy remains explicitly open. This is
an integrity repair, not coordinated adoption or nine-finding completion.

### Integrated regression repairs

`CARTULARY_TEST_RUN_ID=remediation-integrated-check-20261015-01 make check`
ran against source
`sha256:2d9845c4cafa2a994315e3c0c4bb6cc0bc97bb9f569dc5312651876505a15569`.
It failed with outer Make exit 2: 772 of 781 work units passed, nine failed,
none were skipped or cancelled, and all 1,043 cleanup steps completed.
The owning summary reports `product / test_assertion_failure`, 504.048 seconds.
The telemetry child separately reports `harness / child_target_failure`.
Exact retained root: `.cartulary/test-results/remediation-integrated-check-20261015-01`.

The following repairs preserve the affected obligations rather than disabling
their checks. Reviewer: Codex under the authorized remediation scope.

| Failure | Repair and retained assurance |
| --- | --- |
| Entities export boundary | Retain the explicit `ProjectionDependencies` read port and method in the exact export inventory. Entities reads its own mention dependencies; application assembly injects this consumer-owned Revisions port. The boundary check remains strict. |
| Operator evidence byte hash | Replace the changing whole-catalog digest with an independently fixed one-migration CLI serialization vector, including exact JSON order, manifest/source hashes, evidence-only flags, safe binding, ledger and finding. Current-catalog transport, projection and relocation tests remain. This retires incidental inventory spelling, not the byte protocol. |
| Viewer Evidence inspector fixture | Admit an explicit metadata response and wait for eligible read actions. The viewer/closed-incident upload denial remains asserted. Metadata admission is not bypassed. |
| Timeline foundation fixture | Supply its required Workbook browsing provider; stable adapter, row, query and pending-state assertions remain. |
| Workbook overlapping mounts | Separate presentation registration from updates. Only the current token may update readiness or detach, and obsolete cleanup cannot remove a successor. Existing composition, navigation and semantic-focus tests now exercise this boundary. |
| Unused frontend export | Keep the Evidence access helper private to its presentation owner. No behavior or external interface is removed. |
| Records provider numeric expectation | Require `json.Number("3")`, preserving the exact canonical integer rather than converting through a float. |
| Network Flow restore fixture | Resolve current Reporting profile/template bindings through their owners. Corrupt only Graph references for the atomic-rollback assertion. This fixture proves Graph restore reconciliation, not full native Reporting render readiness. |
| Telemetry public-error completeness | Classify the admitted `invalid_evidence_metadata_request` as `request_invalid` in the existing authored registry. No new class or payload dimension is introduced. |

All focused runs below exited 0 and use source
`sha256:9140407343a595569523e8d1cacc225139d5b706a549984b12cd255fc5fcab79`.
Run IDs have prefix `remediation-integration-repairs-`, suffix `-20261015-01`,
and roots under `.cartulary/test-results/`.

| Run middle | Command selection | Duration |
| --- | --- | --- |
| `operator` | `test-slice OWNER=app.operator ROWS=app.operator.unit.migration_evidence_cli_transport_and_redaction_906768574c` | 5.430 s |
| `entities` | `test-slice OWNER=module.entities ROWS=module.entities.unit.root_boundary_guards_e292ec1361` | 4.240 s |
| `records` | `service-backed-test-slice OWNER=module.records ROWS=module.records.reporting.envelope_facts` | 15.098 s |
| `networkflow` | `service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2` | 70.790 s |
| `workbook` | `test-slice OWNER=web.workbook`, the six exact rows listed below | 11.438 s |

Workbook rows: `web.workbook.regression.surface_composition_debb2dc3cb`,
`web.workbook.regression.timeline_first_open_evidence_focus_78b3aba0fd`,
`web.workbook.regression.timeline_surface_foundation_owns_stable_adapter_02f2dc6f6f`,
`web.workbook.regression.useworkbookstartupcontroller_zsemantic_grid_focus_6f31bf7e92`,
`web.workbook.regression.workbench_navigation_admission`, and
`web.workbook.regression.workbench_navigation_intent`.

The Workbook correction follows Core 03 §2's obsolete-entry fencing and the
existing layout/query ownership boundary; it changes no geometry, theme,
density, authorization, retained writes, or persisted data. All internal callers
use explicit registration; no compatibility fallback remains. The UI digest
assessment is PASS for A001–A003, A010–A011, A016, A023–A026 for this narrow
boundary, supported by inspected owners and the focused rows. A004–A009,
A012–A015, A017–A022 are not applicable to changed presentation-lifetime mechanics
or fixture-only corrections; this is not a visual/accessibility qualification.
A027 passes for this slice with the integrated result below. No upstream design advice was applied.

Format and generation passed in `format` (5.253 s) and `generate` (9.600 s).
Finalization passed in `finalize` (18.852 s, outer exit 0); retained-run
maintenance was skipped because `RESULTS_DIR` was unset. Markdown lint passed
in `markdown` (8.910 s, outer exit 0).

`CARTULARY_TEST_RUN_ID=remediation-integrated-check-20261015-02 make check`
passed on the same source in 505.881 s, outer exit 0: all 781 work units passed,
none failed, skipped or cancelled, and all 952 cleanup steps completed. The
cache split was 173 bypass, 106 hits and 502 misses; cache hits are not fresh
execution. Its graph digest is
`sha256:cb032beaa6f9e0f15a5d305269b635251a28742c0dfb3f901527ec691965779b`.
This repairs the earlier failed invocation; it does not erase that evidence or
establish an intermittent outcome under unchanged source.

`make test-evidence-audit OWNER=<owner> EVIDENCE_ROOTS_FILE=<manifest>` passed
for Reporting, Composition, Records, Workbook and Telemetry using the integrated
run's exact target partitions. Audit IDs use the same repair prefix/suffix and
middles `reporting-audit`, `composition-audit`, `records-audit`, `workbook-audit`
and `telemetry-audit`. Their explicit root manifests are retained in
`.cartulary/remediation-evidence/integration-repairs-<owner>-20261015.json`.

The first `operator-audit` failed with outer exit 2 and
`artifact / artifact_error`: the integrated `backend-process` selection does
not contain Operator's complete owner partition. It was not a product failure.
The complete `test-slice OWNER=app.operator` passed in `operator-owner`
(11.284 s, outer exit 0), and `operator-audit-20261015-02` then passed using
`integration-repairs-app.operator-owner-20261015.json`. Both audit attempts
remain retained. These audits establish the named current owner coverage;
release, clean-source performance-portfolio and coordinated adoption remain open.

### Frozen support target follow-through

The [support target closure record](reporting-cutover-adoption-review.md#frozen-support-target-closure)
records snapshot-bound support identities, exact frozen target reachability,
indexed references and rejection of malformed or missing historical support.
Full Reporting, Composition and server owner suites/audits, lint and generated
checks pass on source
`sha256:8308cc4507fe0349261a46fad153c36a447f73c60c1147aa4dc0199c1c04d79b`.
The earlier integrated `check` remains bound to its earlier source; it is not a
fresh integrated pass for this change. Source-index/disclosure fidelity and the
larger renderer, timing, cost and adoption obligations remain open.


### Renderer supervision continuation

The [renderer process and live-cancellation record](reporting-cutover-adoption-review.md#renderer-process-supervision-and-live-cancellation)
records the new owner-routed native confinement tests, closed renderer stage
input, real PDF artifact checks, and production worker cancellation during an
active render. Every added executable identity is cataloged through the existing
Reporting owner; no new execution catalog, fallback runner, skip, or compatibility
reader was introduced. The existing formatter now keeps informational diagnostics
from displacing blocking errors.

The record distinguishes native confinement, browser observation, direct adapter
execution and production integration. The placeholder production renderer is
still present; native attempted-network accounting, complete packaging and
production wiring remain required. This continuation does not close the overall
nine-finding remediation, release qualification, qualified performance windows
or coordinated Reporting/Composition adoption.


## 2026-10-15 renderer and native-sandbox continuation

The Reporting adoption handoff now records real Mermaid rendering and passive
SVG admission, shared source-v2 vectors, the narrow pinned Mermaid correction,
stage protocol v2, catalog-v8 explicit preparation, native Linux connection
mediation and Go-to-Node real PDF repeatability. See
[the current checkpoint](reporting-cutover-adoption-review.md#2026-10-15-continuation-real-mermaid-stages-and-native-network-mediation)
for source bindings, exact run roots, retained failures and evidence limits.

The production candidate and packaged runtime integration remain open. These
engineering results do not close adoption or all nine findings, and no current
full-owner evidence audit or qualified cost portfolio is claimed. Prior audits
that rejected source mismatch remain unsuccessful. Finalize retained-run
maintenance was skipped because `RESULTS_DIR` was unset.

## Packaged native rendering checkpoint

The [Reporting adoption review](reporting-cutover-adoption-review.md#installed-stage-validation-checkpoint)
now retains passing real PDF/SVG execution inside the installed native sandbox,
fresh-attempt byte comparisons, native denial and cancellation checks, installation
tamper rejection, full non-service Reporting routing, harness routing, lint,
generation drift, and targeted security checks. The review also preserves the
startup and teardown failures that led to those repairs. Production candidate
wiring and coordinated adoption remain open; this checkpoint does not close
all nine findings or qualify a performance portfolio.

## Durable Reporting abort evidence checkpoint

The [Reporting adoption record](reporting-cutover-adoption-review.md#durable-abort-diagnostics-and-recovery)
now records canonical failed/canceled-attempt diagnostics, atomic owner/Jobs
finalization, expiry-safe retention, queued cancellation, cleanup priority,
recovery enumeration and rejection of rehashed or missing history. Migration 74
uses the approved clean pre-production cutover. Full Reporting, DDL recurrence and
shared finalization tests passed; the record retains a reproduced canceled-Job
expiry defect and its repair. Exhaustive validation, structured post-redaction
rendering, remaining timing predicates, qualified cost windows and coordinated
adoption remain open. No nine-finding completion claim is made.

### Integrated abort-diagnostics validation and structural guard correction

The [Reporting checkpoint](reporting-cutover-adoption-review.md#integrated-validation-after-abort-diagnostics)
records the complete `remediation-render-terminal-integrated-20261015-01`
`make check` outcome: outer 2, 785/790 units, five diagnosed test-projection or
fixture failures. Each now has a passing targeted successor. The Jobs declaration
guard replaces the overbroad repository-wide call-spelling assertion in its
existing catalog row, with positive/negative detector cases and both assurance
reviews retained. No retry-to-green or instability claim is made.

Coordinated adoption remains open. Source review found that the required
Host/Identity Party-assignment facts are missing, in addition to the already
recorded flat post-redaction model and incomplete template/Composition semantics.
These prerequisites must be implemented and exercised; a companion document's
`closed` label cannot supply the missing product evidence.
## Current rendering checkpoint

See [bounded stage diagnostics and admission](reporting-cutover-adoption-review.md#shared-model-validation-boundary)
for the completed 789/790 integrated run, the independently repaired boundary
false positive, action-parameter admission, exhaustive Slidev/Mermaid source
diagnostics, and the subsequent successful 24/24 Reporting owner run. This
checkpoint does not close the remaining source-assignment, structured model,
Composition, timing, qualified cost, consolidation, or adoption obligations.


## Asynchronous Reporting source and render preparation

The [current render recipe checkpoint](reporting-cutover-adoption-review.md#render-recipes-wired-into-production--2026-10-16)
records compact snapshot capture, v4 release/preview admission, cancelable
preparation, independently sealed derivation, detached candidate validation,
preparation-aware failure/recovery, migration 94, and fresh slice evidence.
Production rendering and sandbox protections remain active. Broader owner,
large-grid timing, qualified portfolio and coordinated-adoption obligations
remain open. The latest finalizer passed; retained-run maintenance was skipped
because `RESULTS_DIR` was unset.
