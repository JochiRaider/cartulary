# Release reliability and recovery remediation

The implementation is on `main`. The qualified implementation candidate is
`36ed8c239981b576caf58b09a7bf772c4884a536`, with source fingerprint
`sha256:8aff5143719ce3efd33e6298814653f101f42a788f2bc3d71a30ce6b01fbc4e8`.
Work began from clean `2e4fe6f14ca4816ce9c1ebd7acd3d95b1f757781`.
The fresh complete release run `20261004T225028Z-p76574` passed all 1,260 required
units, with no failures, skips or cancellations.

## Contracts and resulting behavior

| Workstream | Adopted owner and implementation | Acceptance evidence |
| --- | --- | --- |
| Browser acquisition | Harness REQ-804/416/502, AC-110/111: an immutable private acquisition record precedes startup; process and port ownership survives absent ready leases; settlement proves release or retains exact dependencies. | Owner-boundary fault injection, broker matrix, lifecycle row, both real seeded profiles and performance-fixture retirement. |
| Failure preservation | Harness REQ-301/302: preparation, structured IPC, command diagnostics and test finalization preserve the primary failure and collect cleanup failures separately. | Classified CLI failures, malformed/dead IPC, failed reaping, live-stack cleanup failure and named command-propagation cases. |
| Timeline coverage | Existing Core 03 section 13.4: independent vertical, horizontal, paging and active-gesture unmount scenarios preserve product obligations. | The stable range-scrolling row selects all four exact scenarios; the grid-adapter disposal row checks frames, capture, listeners and no later selection commit. |
| Test lifecycle and concurrency | Harness runner rules and AC-006/014/032: imports define/register cases; named cases await children and finalization; deliberate timer expiry is controlled at the private executor boundary. | Import-purity probe, Make/row propagation, timeout/cancellation/escalation/reaping cases, worker/claim consistency and the extended harness contract. |

`cartulary.browser_acquisition.v1` is private recovery authority. The ready lease
continues to represent attachment readiness. One browser lifecycle owner reaps
the producer, inspects late resource registrations, stops exact service processes,
releases exact port reservations and retires fixture metadata. The managed suite
continues to own its service ledger and containers, including acquisition that
stops before fixture metadata publication. No resource names are reconstructed.
An exit trap cannot declare terminal release before its producer is reaped.
Every managed browser process has an immutable launch ticket before spawn. The creator records
its exact identity before acknowledging launch; the child records it before
starting dependent work. An unbound ticket blocks release unless its creator
proves no child was created or reaps its exact group. Stopping fences new launches,
and late launchers register their identity and exit without starting payloads.
Managed browser services use this protocol directly; redundant PID-only parent
monitors were removed from that path.

Preparation tracks its own resources and exact suite dependency even in a borrowed
runtime. Directory ownership determines deletion authority; unrelated controller
records do not block a suite. Independent releases continue after a failure.
Ordinary provider settlement is memoized; explicit later recovery is a separate
attempt. Missing, altered, cross-run or ambiguous proof fails closed.

Ordinary retirement takes its run, result root and target from the validated,
digest-bound managed-suite proof. It cannot fall back to an ambient or `adhoc`
run. Explicit recovery uses a new private lifecycle output directory per attempt,
so it neither backfills historical evidence nor collides with earlier immutable
recovery output. Proof validation precedes launch and destructive settlement.

Timeline fixtures use 105/4/405/4 rows. Creation remains serial through public
APIs, with one authentication snapshot and returned row identities. Clock work
uses batches of eight 16-ms frames within the original frame budget. Horizontal
coverage creates real overflow through the public width control and requires
more than 500 pixels of movement. Unmount uses an application surface transition
while the document remains alive. Product interfaces and the 60-second deadline
are unchanged; assertions against implicit queries and document scrolling remain.

Affected Node invocations use one test-file worker. The `harness_contract` profile
reserves four CPU slots, twenty process slots, 1536 MiB and one I/O slot, including
nested Make/diagnostic wrappers and deliberate three-child races. Process-tree
sampling observed peaks of 16 processes in the command slice and 15 in the full
contract; maximum summed RSS in the latter was 888.4 MiB. These observations
support the reservation, not a proof of historical contention. Diagnostic children
retain the 10-second deadline. Production deadline start and escalation grace are
unchanged; deterministic deadline tests retain an independent real-time watchdog.

## Changed owners and files

| Area | Principal files |
| --- | --- |
| Lifecycle and recovery | `tools/harness/browser/browser-acquisition.mjs`, `review-preparation.mjs`, `start-web-e2e.sh`, `lifecycle/ports-and-token.sh`, `ui-review/preparation.mjs`, `ui-review/session.mjs`, and `tools/harness/scheduler/fixture-broker/providers.mjs`. |
| Lifecycle regression support | `tools/harness/browser/tests/test-browser-acquisition.mjs`, `test-ui-review-preparation.mjs`, `test-ui-review-seeded.mjs`, `ui-review-fixture-cleanup.mjs`, `ui-review-public-workflow.mjs`; `tools/harness/tests/fixture-cleanup-cases.mjs` and `fixture-stack-driver.mjs`. |
| Timeline | `apps/web/e2e/timeline-range-selection.spec.ts`, `apps/web/e2e/timeline-auto-resolution-feedback.spec.ts`, `apps/web/e2e/support/timeline/fixtures.ts`, `packages/grid-adapter/src/gridInteraction.test.ts`, and `tools/test_families/module.timeline.json`. |
| Command lifecycle | `tools/harness/tests/test-command-failure.mjs`, `command-failure-fixture.mjs`, `test-frontend-producer-graph.mjs`, `test-frontend-producer-lifecycle.mjs`, `frontend-producer-fixture.mjs`, `contract-initialization-cases.mjs`, `test-harness-command-surface-contracts.mjs`, and `tools/harness/scheduler/work-graph/executor.mjs`. |
| Machine projections | The acquisition schema, schema attachments, task-surface owner, work-graph owner, command-surface family, catalog resource-profile enumeration, and corresponding schema enumerations under `tools/`. |
| Generated outputs | `tools/task_surface.generated.mk`, `tools/task_surface_manifest.json`, `tools/execution_topology_manifest.json`, and `tools/execution_topology_render_index.json`, refreshed only through Make. |
| Specification and operations | `docs/testing-harness-nlspec.md` and `docs/guides/cartulary_browser_design_readiness_workflow.md`. |

Domain vocabulary and Core 03 are unchanged. No product API, bulk fixture endpoint,
legacy recovery reader, automatic translation, dependency lockfile change or
public timeout increase was introduced. Executable checks consume machine
projections, never Markdown. Human review checked the projections against the
adopted owners; routing success alone is not specification completeness.

Implementation was split into specification/projection, recovery, test-structure,
and subsequent regression-driven correction commits. The final handoff is a
separate documentation change. Earlier handoffs remain historical:
[preparation remediation](ui-review-preparation-remediation.md) and
[fixture cleanup](ui-ux/harness-fixture-cleanup-remediation.md).

## Final focused qualification

The following commands each passed three consecutive times without retries on
`07798a24009de3445666a58e083819e2cfdae12c`, fingerprint
`sha256:496026f8fac7e7754ec0b5b5c8cb768278374e50fd663e4cbbff872c63811f9e`.
Their recovery, command-lifecycle, range-scrolling and measurement sources are
unchanged in the final candidate. Run IDs resolve under `.cartulary/test-results`.

| Command | Pass 1 | Pass 2 | Pass 3 |
| --- | --- | --- | --- |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_contract,harness.browser.boundary_support.ui_review_lifecycle` | `20261004T211415Z-p80882` | `20261004T212823Z-p35504` | `20261004T214216Z-p88456` |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.ui_review_seeded_default,harness.browser.integration.ui_review_seeded_network_flow_claimed` | `20261004T211527Z-p84709` | `20261004T212935Z-p39332` | `20261004T214328Z-p92289` |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_scrolling` | `20261004T211940Z-p43650` | `20261004T213348Z-p97413` | `20261004T214738Z-p50973` |
| `make test-slice OWNER=harness.command_surface` | `20261004T212102Z-p80644` | `20261004T213508Z-p35046` | `20261004T214858Z-p88049` |
| `make harness-contract` | `20261004T212144Z-p83900` | `20261004T213550Z-p38300` | `20261004T214940Z-p91330` |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.measurement.committed_timeline_summary_typing_acknowledgment_b615aabfe6,module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13,module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95,module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3` | `20261004T212320Z-p87805` | `20261004T213723Z-p42233` | `20261004T215113Z-p95263` |
| `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.range_pointer_binding` | `20261004T212817Z-p34946` | `20261004T214212Z-p88026` | `20261004T215606Z-p42260` |

All 21 manifests bind that focused candidate identity. Every required
unit passed; all 184 cleanup steps completed. Browser reports show first-attempt
success and exactly four selected range scenarios. Each measurement run contains
four finalized leases with complete cleanup in its own run. No late asynchronous
activity warning was found. A read-only audit after the sequence found no runtime
created since qualification began remaining under the harness scratch root.

| Timeline scenario | Three durations (ms) | Three setup durations (ms) | Serial row creates | Clock calls |
| --- | --- | --- | ---: | ---: |
| Vertical | 12533, 12424, 12198 | 1249, 1274, 1185 | 105 | 22 |
| Horizontal | 4020, 3975, 3956 | 70, 66, 59 | 4 | 21 |
| Paging | 8003, 7575, 7435 | 4439, 4185, 4055 | 405 | 0 |
| Active-gesture unmount | 2544, 2540, 2560 | 50, 47, 44 | 4 | 0 |

Each fixture operation used one authentication snapshot. These are functional
verification observations, not a product performance claim. Clock and deadline
budgets are unchanged. In the final complete release, the four scenarios took
26,266/7,588/17,356/4,236 ms under shared execution. Earlier scheduler evidence showed
fully reserved CPU and service lanes during the range group; reservation evidence does
not establish actual host utilization or the cause of the historical timeout.

## Additional Undo boundary qualification

The final candidate additionally corrects the Undo continuity fixture in
`apps/web/e2e/timeline-auto-resolution-feedback.spec.ts`. Missing timestamps had
allowed random record IDs to determine row order. Explicit sequential timestamps
now keep the source first and the newer editor last. The native-editor scenario
forces the lower scroll boundary before the Undo disclosure disappears.

The retained geometry is exact: content height stays at 1,248 pixels while the
viewport grows from 527 to 570 pixels. The legal maximum therefore falls from
721 to 678 pixels. The test requires that precise browser clamp and unchanged
horizontal scroll, while retaining the original DOM node, focus, draft text and
backward selection assertions. It changes no product behavior or deadline.

Three consecutive unretried passes on the final candidate cover all twelve cases
in `module.timeline.browser.auto_resolution_feedback_characterization`, including
all seven cases sharing this fixture:

| Command | Pass 1 | Pass 2 | Pass 3 |
| --- | --- | --- | --- |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.browser.auto_resolution_feedback_characterization` | `20261004T224350Z-p63214` | `20261004T224600Z-p1079` | `20261004T224810Z-p39017` |

All 33 required unit executions passed and all 42 cleanup steps completed. Every browser
case passed on its first attempt. Each run retained the before/after geometry above and
binds the final candidate commit and fingerprint. The later complete release qualifies
the combined source.

## Finalization and static checks

| Command | Passing run | Scope |
| --- | --- | --- |
| `make agent-finalize` | `20261004T224041Z-p52449` | Final source content. |
| `make frontend-typecheck` | `20261004T224102Z-p56492` | Final source content. |
| `make lint-biome` | `20261004T224124Z-p57073` | Final source content. |
| `make generate-drift` | `20261004T224128Z-p57513` | Final source content. |
| `make generated-artifact-policy-check` | `20261004T224143Z-p61675` | Final source content. |
| `make json-shape-check` | `20261004T224148Z-p62171` | Final source content. |
| `make lint-scripts` | `20261004T211207Z-p71800` | Unchanged harness scripts. |
| `make lint-shell` | `20261004T211210Z-p72252` | Unchanged harness scripts. |

`make generate` passed after the final test correction. Retained-run maintenance
was skipped by `agent-finalize` because `RESULTS_DIR` was unset. The completed
handoff also received `make lint-markdown` before its separate documentation
commit. Documentation is outside executable source fingerprints.

## Complete release gate

`make release-check` passed at `20261004T225028Z-p76574` with normal declared capacity,
no filters, no retries and no capacity override. All 1,260 required units passed; all
1,215 cleanup steps completed. Its manifest binds the candidate identity above. All 607
browser cases across 107 reports passed on their first attempt. The required unit set
exactly matches the independently generated release plan; no filtered or omitted unit is
counted as complete. Normal cache policy reused 216 units, with 627 misses and 417
bypasses; no cache override was supplied. Cleanup accounting contains 1,044 host
releases, one for each unit that did not hit cache, plus 22 fixture releases, 147
fixture detachments, and service/runtime closure. No new suite runtime remains, and no
late asynchronous activity warning was found.

## Retained failures and regression evidence

Before the final test correction, candidate `07798a24009de3445666a58e083819e2cfdae12c`,
fingerprint
`sha256:496026f8fac7e7754ec0b5b5c8cb768278374e50fd663e4cbbff872c63811f9e`,
completed `20261004T215617Z-p42831` with 1,258 passed, two failed and no skipped
or cancelled units. Both failed units came from one Undo scroll assertion and
its aggregate summary; all 1,430 cleanup steps completed. The original three
reported problem areas passed in that run. It remains failed evidence.

The Undo assertion required scrollTop 721 after the viewport's legal maximum
had shrunk to 678. The deterministic red run `20261004T223455Z-p72607` forced
that boundary and retained the old equality assertion: exactly the native-editor
case failed, with 9/11 work units passing. Its recorded geometry established
browser clamping without a product continuity failure. With the precise clamp
assertions applied, `20261004T223756Z-p14174` passed all twelve scenarios,
11/11 work units and all fourteen cleanup steps. The final three focused passes
and complete release above supersede that development pass.

The intermediate candidate `7cd71e0ca3484aff54bfb88b8ca964194b04c8f3` passed a
complete release at `20261004T201612Z-p63041`: 1,260/1,260 units and all 1,429
cleanup steps. Its fingerprint was
`sha256:40adb7c1d8318dbc3f08284a9049b3aba2d240c529395db6ddeddf5a2d35e998`.
A final ownership review then identified the process-launch interruption window.
A deterministic regression paused the real producer before module initialization;
recovery incorrectly reported release while the producer was alive. Run
`20261004T205846Z-p82192` retained exactly that failure with 52 other cases passing.
The launch-ticket correction and late-launch/pre-spawn cases passed in the
lifecycle slice at `20261004T210545Z-p3754`; both real seeded profiles passed
4/4 at `20261004T210658Z-p9060`, with all six cleanup steps complete.
Development runs `20261004T210236Z-p89799` and `20261004T210416Z-p96501`
caught optional-callback and missing-proof-classification integration mistakes;
both were corrected before the final candidate was committed.
The final qualification above supersedes these intermediate passes.

The first complete candidate run, `20261004T184238Z-p2511`, completed with 1,255
passed, four failed, one dependency-skipped unit and no cancellations. All 1,429
cleanup steps completed. Its source was `c2d4900a86223efcf3cff7eaa6d5a328369c6531`,
fingerprint `sha256:03ea6bd354f9e8acf2fd541ca1ece40829ff2c96f978817d5f15ce6e7605606a`.
The new cleanup subprocess had omitted run identity, so four measurement clone
leases were written under `adhoc` and their summaries failed with `artifact_error`.
The failed run and misplaced artifacts remain unchanged. They were not backfilled
or treated as qualification.

Three retirement-routing/isolation regressions failed against that implementation
at `20261004T192050Z-p29128`: exactly three failures, 49 passes, no cancellations
or skips. The fixed lifecycle slice passed at `20261004T192258Z-p36105`; the real
measurement slice passed 20/20 at `20261004T192516Z-p43812`, with four correct
leases and all 27 cleanup steps complete. The final focused table supersedes
these development passes for qualification.

Earlier intentional counterfactuals restored early release before producer
reaping and loss of the primary failure when IPC/reaping both fail. The exact two
cases failed at `20261004T174306Z-p98108`, with 45 others passing. The classified
CLI diagnostic regression failed before its fix at `20261004T175736Z-p2443`,
with 48 others passing. Fixed sources were restored and regenerated immediately.
These negative tests are regression evidence, never successful qualification.

Other development failures were repaired before qualification: synthetic cleanup
and IPC precedence (`20261004T165217Z-p68892`); schema enum/attachment ordering
(`20261004T170347Z-p78859`, `20261004T171324Z-p91164`); the launcher's missing Node
argument separator (`20261004T170728Z-p97274`, `20261004T170838Z-p68024`,
`20261004T171027Z-p4564`, `20261004T171143Z-p39612`); and stale generated metadata
(`20261004T173007Z-p15779`). An attempted direct invocation of the internal
`harness-smoke-fixture-broker-smoke` name reported no public Make target; its
matrix was verified through `make harness-contract`, without a raw-script bypass.

## Original evidence and migration status

The originally reported failures tested dirty `54f4219` with source fingerprint
`sha256:5c9373a4378d601d416e9a2371c7275a784640d4ff1560eef8467312432c73fb`.
The original browser startup trigger and the contribution of host contention
remain unknown. Neither is inferred from passing retries or the later recovery
error.

| Case | Failed retained run | Passing diagnostic run |
| --- | --- | --- |
| UI review recovery | `20261004T152543Z-p28254` | `20261004T160912Z-p17945` |
| Timeline scrolling | `20261004T144412Z-p77087` | `20261004T151912Z-p37134` |
| Child diagnostics | `20261004T152104Z-p74695` (cancelled) | `20261004T152459Z-p24229` |

The registered old supervisor `uireview-80a8dffa57b0d366a8c58eb839f70d66` was
stopped with the old tooling using its exact locator at
`/tmp/cartulary-review-supervisors-5jc9Ak/both/ui-review/session.json`. Stop reported
`session_lost` and removed the registry entry. Historical receipts were unchanged.

The following older runtime directories remain preserved under
`/tmp/cartulary-harness-scratch/suite-runtime`:

- `suite-20261002T221420Z-p56807-br4R8j`
- `suite-20261003T005539Z-p6721-2V50CA`
- `suite-20261003T012636Z-p23799-63UATt`
- `suite-20261003T030043Z-p5210-lgZ0YJ`
- `suite-20261003T040642Z-p63009-9jDn6q`
- `suite-20261004T042948Z-p63192-PHe0zy`
- `suite-fixture-cleanup-1790962532794-unresolved-9MCDyG`
- `suite-fixture-cleanup-1791130012864-product-ShMirv`

The last is the originally damaged record: pending browser ownership without
its ready lease, after managed-suite cleanup. Available proof does not establish
exact resource clearance. Cutover on an environment containing unresolved old
ownership remains blocked until exact-owner reconciliation establishes clearance.
Do not fabricate a lease, infer names or delete by age. Fresh qualification runs
use independent runtime identities and do not translate these old records.

Drain sessions with their running version before upgrading or rolling back.
Replace producer, provider, recovery reader and private IPC together, then start
fresh sessions only after that environment is cleared. Rollback likewise drains
sessions created by the version being removed. Successful later recovery removes
only exact owned resources and private proof; it never rewrites the original
failed receipt. No production deployment or unsupported old-record migration is
claimed by this handoff.
