# Runtime-owned private child capture remediation

Execution baseline: 2026-09-23, clean working tree at
`e1f9edfee641dba23272c624266a2f9a3da35657`. The user authorized W1–W4
implementation, including the specification clarification and final gates.
This follow-up does not reopen completed web cleanup iterations.

Outcome: W1–W4 DONE. All required gates pass on the completed implementation;
the four original AC-043 measurements are qualified with zero scheduler overlap.

## Scope and owners

TH-HARNESS-REQ-603 owns private allocation, permissions and cleanup; Section 9
owns failure precedence. AC-011/014/015 cover capture, failure preservation and
retained evidence. `docs/domain.md` excludes harness implementation vocabulary
and remains unchanged. Tests and generators consume machine inputs, never
Markdown. Public commands, semantic identities, artifact schemas/paths,
measurement thresholds and fixture policies remain unchanged.

G1 clarifies allocation ownership. G2 removes caller-supplied physical identities.
G3 makes acquisition and cleanup exception-safe. G4 closes the capture-to-launch
coverage gap. All belong to implementation, tests and documentation; G1 also
changes the adopted specification. No compatibility shim or data migration is
needed. Leaving these gaps open permits valid catalog names to prevent launch,
leaks private resources on failures, and allows routing tests to imply execution.

The shared owner is `tools/harness/runtime/private-child-process.mjs`. Its four
production callers are:

- `tools/harness/browser/browser-catalog-group-cli.mjs`
- `tools/harness/execution/runners/row-runner-cli.mjs`
- `tools/harness/execution/run-make-node-tool-cli.mjs`
- `tools/harness/performance-fixture/snapshot-builder-cli.mjs`

The confirmed prior failure root is
`.cartulary/test-results/20260923T032235Z-p53291`. Four browser groups failed with
`artifact_error`, before Playwright launched, because sanitized capture identities
were 144–145 characters against the 128-character private component bound.
`apps/web/e2e/measurement/timeline-grid.spec.ts` supplies these unchanged rows in
`tools/test_families/module.timeline.json`:

- `module.timeline.measurement.committed_timeline_summary_typing_acknowledgment_b615aabfe6`
- `module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13`
- `module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3`
- `module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95`

The historical `enter_focus` identity exercises pointer focus. Actual AC-043
performance remains unknown until successful execution. Static advisories are
outside this effort.

## Sequencing and exits

| Workstream | Dependency | Status | Exit |
| --- | --- | --- | --- |
| W1 specification and checkpoint | none | DONE | Owner clarification, inventory, routing and baseline recorded |
| W2 runtime and atomic caller cutover | W1 | DONE | All callers migrated; lifecycle/security regressions pass |
| W3 regression and verification integration | W2 | DONE | Real caller execution; authored/generated accounting agrees |
| W4 integrated validation and handoff | W1–W3 | DONE | Required gates pass; final handoff recorded |

Before each workstream its prerequisites and IN_PROGRESS state are saved here;
its evidence and DONE/BLOCKED exit must be saved before its successor starts.
Rollback restores the helper, four callers, tests and authored verification
inputs together, then regenerates projections. Suite roots remain borrowed;
scheduler process-group cancellation remains the process-lifecycle owner.

## W1 execution checkpoint

Prerequisites satisfied: clean baseline inspected; root AGENTS.md, specification,
source owner, four callers, catalog rows and prior failure inspected.
`make task-guide ROLE=module-author OWNER=harness.browser` and the equivalent
`OWNER=module.timeline` passed using `tmp/node-runtime/bin` on PATH. Owner slices
and `harness-contract` provide focused routing; measurement uses the four rows
above and the public `browser-e2e-measurement` target.

Generated membership is controlled by `tools/generated_artifact_policy.json`;
task bindings/backing scripts are authored in `tools/task_surface_owner.json`.
New boundary cases join the existing public harness contract suite; authored
accounting changes precede `make generate`. No new public target is planned.

Observed source defects: existing directory permissions are repaired before
validation; second-stream open is outside cleanup guarding; tail-read failure
occurs before cleanup transfer; cleanup stops on the first failure and is not
idempotent. Deterministic characterization and remediation belong to W2/W3.

Validation sequence: focused contracts and affected owner slices; `lint-scripts`,
`test-catalog-check`, `harness-contract`, `generate-drift`,
`generated-artifact-policy-check`, `json-shape-check`; `agent-finalize` before
`check`; sequential webserver-backed, stateful, a11y, visual and measurement
browser targets; final `lint-markdown` and staged/unstaged whitespace checks.
All repository commands run from the repository root using public Make targets
and the pinned Node runtime. No qualifying retained full warm check is selected
for finalization yet; if RESULTS_DIR is unset, retained-run maintenance is skipped.

### W1 completed exit

W1 DONE. REQ-603 now separates semantic metadata from runtime storage allocation,
defines exception-safe acquisition and ownership transfer, and preserves failure
precedence. AC-011 includes real subprocess and injected lifecycle failures.
Editorial source/owner review found no conflicting owner clause. `git diff --check`
passed; product tests were not needed for this editorial checkpoint. Public
contracts and domain vocabulary are unchanged. Revert the added prose/link to
roll back W1; subsequent implementation depends on this ownership clarification.

## W2 execution checkpoint

W2 IN_PROGRESS after the saved W1 exit. All four consumers and their error paths
are inventoried. Implement runtime allocation and guarded lifecycle, migrate
callers atomically, then validate deterministic real-subprocess regressions before
starting W3. The principal risks are masking a child failure with cleanup and
leaking a descriptor or private file on partial acquisition.

W2 first contract run failed at `.cartulary/test-results/20260923T130743Z-p70051`:
the new test used an incorrect relative catalog path (corrected), and the existing
JSON-shape contract rejected stale topology fingerprints after caller edits.
Both failures relate to this change. Minimal authored backing-script accounting
and generation are pulled forward from W3 so W2 can satisfy its contract exit;
W3 retains integration and final routing verification. No generated output is
edited by hand.

### W2 completed exit

W2 DONE. The runtime allocates unique `capture-*` directories with fixed stream
names, validates permissions/ownership without repair, guards descriptors and
owned paths, validates bounded reads, and transfers idempotent cleanup only with
a complete result. All four production callers no longer supply captureID.
Caller cleanup attempts remaining releases and preserves earlier primary errors;
cleanup-only errors retain the existing cleanup taxonomy. No compatibility alias
or public identity change was introduced.

`make generate` passed at `.cartulary/test-results/20260923T130839Z-p73479`.
`make harness-contract` passed at
`.cartulary/test-results/20260923T130900Z-p76633`, including fifteen isolated
real-subprocess lifecycle/security scenarios. Injected failures cover second
stream open, synchronous/asynchronous spawn, close, read and partial unlink;
concurrent captures retain all four original semantic row IDs. Full browser
execution remains W3/W4 evidence. Rollback requires all four callers together
with the shared helper, cases, authored backing scripts and their projections.

## W3 execution checkpoint

W3 IN_PROGRESS after the saved W2 exit. Add controlled process-group cancellation
coverage, confirm semantic graph identities, run affected owner and measurement
rows, and reconcile generated accounting. Successful graph construction alone
will not close this workstream; actual Playwright evidence is required.

W3 validation correction: the first `test-slice OWNER=harness.browser` launch
failed before creating a run root because it overlapped `make generate` and saw
an ephemeral generated Go file during fingerprinting. This was orchestration
related, not a product failure. Generation and dependent verification are now
strictly sequential; the slice is rerun after generation completes.

W3 evidence: `harness-contract` passed at
`.cartulary/test-results/20260923T131202Z-p87559`, including controlled scheduler
cancellation after captured-child readiness and reaping before suite cleanup.
`test-slice OWNER=harness.browser` passed 38/38 units at
`.cartulary/test-results/20260923T131202Z-p87472`. The four exact Timeline rows
passed 20/20 graph units at `.cartulary/test-results/20260923T131350Z-p28175`;
each browser group passed and produced its existing
`frontend-measurement-summary.v3.json`. This executes both browser-group capture
and the performance snapshot builder, while owner slices execute row capture and
Make tooling exercises the Node wrapper. Retained-secret scans passed.

An attempted `make target-plan TARGET=test-slice OWNER=module.timeline ROWS=...`
was rejected before execution (usage exit 2, OWNER is not an input to target-plan,
no run root). The supported `test-slice` command above supplied the actual proof;
this discovery failure is unrelated to product behavior.

Final W3 contracts include eighteen isolated capture scenarios (including invalid
component characters/length and cancellation), plus failed-row capture/context
release assertions in the existing command-failure suite. Imports remain inert.
`make generate` passed at `.cartulary/test-results/20260923T131854Z-p70689`;
`lint-scripts` passed at `.cartulary/test-results/20260923T131928Z-p73851`;
`test-catalog-check` passed at `.cartulary/test-results/20260923T131930Z-p74186`;
`harness-contract` passed at `.cartulary/test-results/20260923T131928Z-p73855`.

An unnecessary direct invocation of the internal candidate target
`harness-command-surface-contract` failed at
`.cartulary/test-results/20260923T132000Z-p77791`: its Vitest diagnostic fixture
requires the enclosing suite-runtime environment. This is an invocation error,
not a capture regression. The same cases, including the new failed-row release
assertions, passed through the supported `make harness-contract` immediately
before it. No internal target is retried standalone; public routing remains the
required evidence. The interrupted gate sequence resumes at `generate-drift`.

### W3 completed exit

W3 DONE. All four real measurement scenarios and the affected owner slice pass;
all four caller paths execute successfully. New lifecycle cases and existing
caller/graph coverage retain complete semantic identities and prove bounded,
exclusive allocation, failure cleanup, retry, privacy and process-group
cancellation. Authored backing scripts precede generated projections.
`generate-drift` passed at `.cartulary/test-results/20260923T132059Z-p78670`,
`generated-artifact-policy-check` at `20260923T132059Z-p78672`, and
`json-shape-check` at `20260923T132059Z-p78674` (all roots under
`.cartulary/test-results`). No unresolved W3 blocker remains. Policies, thresholds,
public commands and schemas are unchanged; no migration is required.

## W4 execution checkpoint

W4 IN_PROGRESS after the saved W1–W3 exits. Run `agent-finalize` now, without
RESULTS_DIR: retained-run maintenance is skipped because no qualifying successful
full warm check root was selected. Then run `check` and, sequentially,
`browser-e2e-webserver-backed`, `browser-e2e-stateful`, `browser-e2e-a11y`,
`browser-e2e-visual`, and `browser-e2e-measurement`. Measurement runs alone.
Finish source/compatibility/rollback accounting and the saved handoff before
`lint-markdown` and staged/unstaged whitespace checks. Newly exposed performance
failures would block readiness and belong to Timeline; no threshold adjustment
is authorized.

### W4 verification and change accounting

`agent-finalize` passed at `.cartulary/test-results/20260923T132152Z-p83230`.
Its summary reports generated files unchanged, zero updated files, no rollback
needed, and retained-run/performance maintenance skipped with RESULTS_DIR unset.

The change set comprises these authored and generated paths:

- `docs/handoffs/harness-private-capture-remediation.md`
- `docs/handoffs/web-apps-cleanup-tracker.md`
- `docs/testing-harness-nlspec.md`
- `tools/execution_topology_render_index.json`
- `tools/harness/browser/browser-catalog-group-cli.mjs`
- `tools/harness/execution/run-make-node-tool-cli.mjs`
- `tools/harness/execution/runners/row-runner-cli.mjs`
- `tools/harness/performance-fixture/snapshot-builder-cli.mjs`
- `tools/harness/runtime/private-child-process.mjs`
- `tools/harness/tests/contract-suite-support.mjs`
- `tools/harness/tests/private-child-capture-cases.mjs`
- `tools/harness/tests/test-command-failure.mjs`
- `tools/task_surface_manifest.json`
- `tools/task_surface_owner.json`

The specification and two handoffs establish ownership and preserve history.
The shared runtime and four callers implement allocation, release and failure
precedence. The new boundary module, suite registration, command-failure test and
browser graph assertions provide regression coverage. `task_surface_owner.json`
is authored backing-script accounting; `task_surface_manifest.json` and
`execution_topology_render_index.json` are generator-produced projections.
No test-family row, public schema, artifact path, command ID, measurement policy,
fixture policy, domain vocabulary, dependency or product source changed.
The unrelated visual fixture's `captureID` local variable denotes public visual
capture metadata and is not a consumer of the removed runtime option.

Compatibility is an atomic private API cutover; runtime storage is ephemeral and
needs no migration. Rollback restores this helper and all four callers, tests and
authored accounting together, regenerates projections, and revisits REQ-603's
implementation mapping. Do not revert only the helper or only its callers.
Existing suite cleanup owns older leftovers. Static advisories, UI optimization,
dependency upgrades, deployment and release/benchmark certification stay outside
scope. No thresholds, assertions, visual tolerances or privacy checks were relaxed.

`make check` passed 956/956 units at
`.cartulary/test-results/20260923T132224Z-p87198` after finalization. This includes
integrated row execution, harness contracts and normal repository checks. Final
browser gates now execute sequentially in the requested order, with no competing
verification job during measurement.

`browser-e2e-webserver-backed` passed 138/138 units at
`.cartulary/test-results/20260923T133211Z-p67450`, including all 66 browser groups.
The sequential command continues with stateful, accessibility, visual and
measurement gates. No code changed after finalization or during these gates.

`browser-e2e-stateful` passed 42/42 units at
`.cartulary/test-results/20260923T133211Z-p67452`, including all 17 browser groups.

`browser-e2e-a11y` passed 20/20 units at
`.cartulary/test-results/20260923T133211Z-p67498`.
`browser-e2e-visual` passed 12/12 units (44 rows) at
`.cartulary/test-results/20260923T133211Z-p67456`, with unchanged tolerances.

W4 review identified two remaining reporting edges to close before completion:
the snapshot builder has no scheduler command-failure channel, so a cleanup-only
exit needs an explicit existing-taxonomy marker; the Make wrapper's generic
capture diagnostic drops the helper's secondary cleanup-failure annotation.
Add real caller subprocess regressions and preserve the helper's bounded message.
The active browser sequence finishes without source mutation; focused contracts,
finalization and final gates will validate the completed reporting behavior.

The initial full `browser-e2e-measurement` passed 24/24 units (eight groups) at
`.cartulary/test-results/20260923T133211Z-p67454`. The entire requested browser
sequence passed before the reporting corrections. W4 remains IN_PROGRESS until
caller regressions, finalization and final gates validate the completed source.
The corrections add no failure code: snapshot cleanup-only errors publish the
existing class/reason marker; the Make wrapper preserves the runtime's bounded
primary/secondary diagnostic. Real CLI subprocess fixtures exercise cleanup-only,
child-plus-cleanup, and capture-plus-cleanup precedence without service mocks in
production or new test flags in the runtime.

The first expanded caller regression run failed at
`.cartulary/test-results/20260923T140358Z-p22063`: its read-open injector compared
flags to bare O_RDONLY and missed O_NOFOLLOW. The wrapper therefore correctly
preserved the fixture child's duration-baseline failure (13), rather than the
intended injected capture failure (11). Corrected the injection to distinguish
read-open from exclusive creation; no assertion or production behavior was
weakened. The snapshot cleanup-only/primary-failure regressions passed in that run.

The full semantic-identity graph assertions now live in the public graph contract
suite, rather than adding changes only to an extended-tier smoke script. This
makes all new regression assertions reachable through `make harness-contract`.

The next `harness-contract` run at
`.cartulary/test-results/20260923T140611Z-p29250` reached and passed the caller
failure-precedence assertions, then failed a fixture assumption about a configurable
wrapper scratch path. The Make wrapper owns its runtime allocation and does not
use that environment override. The regression now records the actual allocation
inside its private fixture and asserts that exact owned directory is gone; this
strengthens cleanup proof without changing runtime configuration or assertions
about failure precedence. The new public semantic-identity graph case passed.

The completed reporting fixes pass twenty isolated capture/caller scenarios and
the public semantic-identity graph case. Final focused gates passed:

| Target | Run root below `.cartulary/test-results` | Result |
| --- | --- | --- |
| lint-scripts | `20260923T140830Z-p36572` | PASS |
| test-catalog-check | `20260923T140832Z-p36859` | PASS |
| harness-contract | `20260923T140830Z-p36576` | PASS, including all new assertions |
| generate-drift | `20260923T140830Z-p36486` | PASS |
| generated-artifact-policy-check | `20260923T140830Z-p36488` | PASS |
| json-shape-check | `20260923T140830Z-p36490` | PASS |

No unresolved implementation blocker remains. Finalization, full check and the
full sequential browser gates are renewed against this completed source before
W4 is marked DONE. Earlier passing runs remain historical evidence.

Renewed `agent-finalize` passed at
`.cartulary/test-results/20260923T140949Z-p44763`; RESULTS_DIR remains unset,
retained-run maintenance remains skipped, and generated artifacts were unchanged.
The final full check and browser sequence follow this finalization.

The renewed `check` passed 956/956 units at
`.cartulary/test-results/20260923T140949Z-p44867`. The final source now proceeds
through all five browser gates in order. No further production or test changes
are planned; only validation accounting and final documentation remain.

Final-source browser validation: `browser-e2e-webserver-backed` passed 138/138
units and all 66 groups at `.cartulary/test-results/20260923T142020Z-p28933`.

## Final handoff assessment

| Gap | Closure evidence | Assessment |
| --- | --- | --- |
| G1 specification ownership | REQ-603 and AC-011 distinguish semantic metadata, runtime allocation, acquisition, cleanup transfer and failure precedence; no executable Markdown dependency | PASS |
| G2 physical identity coupling | Runtime-owned exclusive allocation; four caller removals; concurrent/repeated capture cases; unchanged four catalog identities and real browser launch | PASS |
| G3 incomplete lifecycle | Twenty isolated lifecycle/caller scenarios; descriptor and owned-path cleanup; unsafe-resource rejection; scheduler cancellation; primary and cleanup-only taxonomy through real callers | PASS |
| G4 launch-boundary coverage | Public capture and graph contract cases, failed-row cleanup assertion, owner slice and final qualified AC-043 evidence with zero scheduler overlap | PASS |

Environment prerequisites used: repository-root public Make commands, the pinned
Node runtime from `tmp/node-runtime/bin` on PATH, pinned Go and frontend tooling,
and the existing Docker-backed test-service/browser prerequisites on this POSIX
host. All required gates use current-run evidence. No platform expansion,
deployment, release certification or benchmark claim publication is included.
Retained-run finalizer maintenance is the only planned maintenance exclusion:
RESULTS_DIR was unset. Static-analysis advisories remain independently owned.

Next engineering action after completion is review and land the atomic change
set through the normal repository process. No data conversion, compatibility
reader, dependency update or follow-up product optimization is required by this
remediation. Preserve the listed failure roots as diagnostic history; do not
reinterpret their failures as performance measurements or successful evidence.

Final-source `browser-e2e-stateful` passed 42/42 units at
`.cartulary/test-results/20260923T142020Z-p28935`.

Final-source `browser-e2e-a11y` passed 20/20 units at
`.cartulary/test-results/20260923T142020Z-p28981`.

Final-source `browser-e2e-visual` passed 12/12 units and 44 rows at
`.cartulary/test-results/20260923T142020Z-p28939`, with unchanged goldens and
tolerances. The final measurement gate runs alone.

Final-source `browser-e2e-measurement` passed 24/24 units and all eight groups at
`.cartulary/test-results/20260923T142020Z-p28937`. Each original Timeline row has
`qualification_outcome: qualified` and `scheduler_overlap_count: 0` in its
`frontend-measurement-summary.v3.json`. The final retained-secret scan passed
across 420 files. No original capture launch failure, newly exposed AC-043 failure,
changed threshold or relaxed tolerance remains.

### W4 handoff prepared

All implementation, contract, generated-accounting, integrated-check and browser
exits pass on the completed source. G1–G4 are closed. The final handoff is saved;
W4 remains IN_PROGRESS only until Markdown and staged/unstaged whitespace checks
confirm this document. No applicable product blocker remains. Every failed target
and its relationship to this work is recorded above; superseding successful runs
are identified explicitly. The next action is normal review of the atomic diff.

### W4 completed exit

W4 DONE. G1–G4 are closed, all required gates passed on the final implementation,
and no applicable blocker remains. The final specification mapping, four caller
migration, twenty capture/caller scenarios, public graph assertions, generated
projections, compatibility and rollback accounting are complete. The existing
web-cleanup advisory links this completed follow-up without rewriting its history.

`make lint-markdown` passed at
`.cartulary/test-results/20260923T145146Z-p79784` (summary:
`adhoc/lint-markdown/tool-run-summary.json`). Staged and unstaged `git diff --check`
passed. Added-file `git diff --no-index --check` produced no whitespace diagnostic;
its exit 1 reports the expected new-file difference, not a failing Make target.
Exact final accounting remains fourteen paths, including two new authored files;
the index was not changed.

The post-completion Markdown revalidation uses the explicit run identity
`capture-handoff-20260923-final`, retained under `.cartulary/test-results`, followed
by staged, unstaged and added-file whitespace checks. Its terminal summary is
`adhoc/lint-markdown/tool-run-summary.json` below that root. This last check covers
the saved completion text and advisory link; it does not change product evidence.

Retained-run maintenance was skipped because RESULTS_DIR was unset. All other
required checks were executed. Static advisories and release/benchmark
certification remain outside scope. The only next engineering action is normal
review and atomic landing; no further remediation is required for this effort.
