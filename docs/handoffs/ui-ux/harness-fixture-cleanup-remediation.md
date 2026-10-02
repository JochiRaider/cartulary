# Fixture cleanup lifecycle remediation

Status: COMPLETE. G1–G6 and W1–W4 pass current verification, including the full
uncached visual comparison and seeded real-stack failure/recovery scenarios.
This bounded repair preserves the existing Workbook implementation and refreshed
visual golden. Earlier visual failures remain recorded below; no UI audit was
reopened.

## Baseline and authority

Baseline: dirty `main` at `3833f7543d56ea2c12e82422c2da41ef7246da37`.
Existing frontend, authored browser-routing and visual-manifest changes belong
to the Workbook enum slice and are preserved.

[Testing Harness NLSpec](../../testing-harness-nlspec.md) §2.1 and
TH-HARNESS-REQ-804/806 own fixture lifetimes; TH-HARNESS-REQ-020 leaves internal
signatures unspecified; §9.1 and TH-HARNESS-REQ-300/301/304/310 own classification
and precedence; §11.3/TH-HARNESS-REQ-402 owns proof and ordered teardown;
TH-HARNESS-REQ-603 and the private-runtime rules own protected evidence.
[Domain](../../domain.md) supplies navigation and remains unchanged. Root
AGENTS.md and generated-artifact policy were reviewed; no nested instructions
apply to the touched harness paths. Tests, generators and runtime consume machine
inputs, never Markdown. Human owner review establishes the projection mapping.

Owner clarifications define single settlement, sticky unhealthy state, ownership
through publication, independent safe finalization, root-owned suites, private
proof retention, and the closed companion artifact. They repair missing explicit
guarantees within the adopted lifecycle; they do not authorize the old defect.
No owner contradiction was found. No internal hook signature was prescribed in
the specification, and no domain, product or Core owner was amended.

Historical failure root: `.cartulary/test-results/20261002T005458Z-p19980`.
The visual assertion was primary; lease-000002 falsely reported cleanup failure
after physical stack stop succeeded. Synchronous quarantine returned undefined,
so nullish fallback invoked destroy and stopped the same lease again.
Diagnostic reproduction: `20261002T015243Z-p81030`. Historical roots remain
immutable. Earlier passing visual roots after the golden refresh are historical
evidence; the final-source comparison below validates this repair.

## Workstream exits

| Phase | Dependency | State | Required exit |
| --- | --- | --- | --- |
| W1 owner and characterization | none | PASS | Owner clarification and pre-fix failing regression recorded |
| W2 allocation cutover | W1 | PASS | One release, stable outcomes, healthy reuse only |
| W3 ordered finalization and evidence | W2 | PASS | Independent cleanup, protected recovery, correct failures |
| W4 final verification and handoff | W1–W3 | PASS | Fresh lifecycle, production recovery, policy and full visual comparison pass; current evidence and earlier failures recorded |

## Gap ledger

| Gap | Final remediation and evidence boundary |
| --- | --- |
| G1 return-value dispatch | Owned allocations require `release({ healthy })`; borrowed allocations require `detach`; `none` needs neither. Admission validates hooks. Allocation `quarantine`/`destroy` aliases and lease `quarantine()` are retired. Pool destruction stays private to the pool adapter. Synchronous undefined and ordinary returns, promises, throws and rejections are covered. |
| G2 unstable repeated outcomes | Each lease and physical allocation memoizes its operation and rejection. Contradictory repeated dispositions reject. Unhealthy shared allocations leave affinity reuse immediately; remaining borrowers drain and cannot restore health. Warm reuse remains healthy-only. Broker and pool close drain in-flight admission/reset/release and retain failures arriving during drain. Physical settlement uses an explicit failure flag, including falsy thrown values. |
| G3 fragmented shutdown | Graph runner and review preparation use the shared suite-controller guard. Fixture providers no longer close the suite. Children drain before broker allocations and provider reserves; the composition root then closes its owned suite, followed by private runtime and evidence. Every independent release is attempted; unsafe dependent service closure is recorded blocked. Repeated close observes its original outcome. |
| G4 lost proof | Ownership is registered before publication and remains tracked through cleanup. Acquisition-publication failure performs one rollback and preserves both failures. Browser stop retains its original lease until acknowledgement; managed-service and proxy owners consume a protected snapshot so their input deletion cannot erase the registered original. Unsettled resources prevent private-runtime deletion. Exact validated recovery proof survives; closed-consumer detail is purged. Missing proof alone never authorizes success or destruction. |
| G5 lost failures/evidence | Ordinary cleanup failures normalize to harness/cleanup_error, exit 12; declared cleanup timeout stays timing/timeout_failure, exit 13. Earlier work failures stay primary. All rejected finalizers survive aggregation, including execution exceptions and boundary fallback. Unit publication/cache admission follow logical release; warm cache admission additionally waits for physical broker/service finalization. Current runs publish the closed companion and investigation tooling validates and displays it. |
| G6 weak production coverage | Existing fixture/scheduler smoke rows now invoke permanent characterization matrices, actual `runWorkGraph`, controlled subprocess owners through production providers, and seeded real-stack work/cleanup-failure scenarios. Tests assert invocation counts, listener closure, ownership, retained outcomes, primary/secondary classifications and recovery rather than reproducing provider teardown. |

### Consumers and source provenance

The shared broker boundary is in
`tools/harness/scheduler/fixture-broker/{index,providers,cleanup-lifecycle}.mjs`.
Migrated production consumers are
`tools/harness/scheduler/work-graph/{scheduler,runner-cli}.mjs` and
`tools/harness/browser/review-preparation.mjs`. The dedicated pool adapter and
fixtures in `tools/harness/tests/test-work-graph-cutover.mjs` migrated atomically.
Cache quarantine is a separate cache operation and was not renamed.

Shared secure ownership proof moved from review-only code to
`tools/harness/runtime/resource-recovery.mjs`; review ownership delegates to it.
`suite-runtime.mjs` guards deletion and purges closed detail during recovery
preservation. `start-web-e2e.sh` retains private lease material on failed cleanup
and supports acknowledgement by its existing owned provider.
The proxy stop invocation retains its explicit configuration arguments; its Go
CLI was inspected because it normalizes configuration before dispatching stop.
No Go source or object-store behavior was changed.

`cleanup-lifecycle.mjs` owns the small suite guard, failure aggregation and
structural ledger. `cartulary.harness_cleanup_results.v1.schema.json` is a closed
machine contract: ordered operations, opaque identities, normalized failures
and safe run-relative lease references. It excludes raw exceptions, credentials,
private paths and resource handles. Cardinality follows admitted lease/allocation
operations and declared finalizers. Existing event, result and lease schema
identities remain intact; lease outcomes are checked for semantic consistency.
`failure-taxonomy.mjs` keeps earlier work/interruption primary over cleanup.
`diagnostics/explain-run-cli.mjs` securely validates the companion, including
run identity, and supports a cleanup-only investigation root.

New tests are `fixture-cleanup-cases.mjs`, `fixture-stack-driver.mjs`, and
`browser/tests/ui-review-fixture-cleanup.mjs`. The existing seeded-default smoke
invokes the live cases; review lifecycle tests settle their fake pending proof.
Authored helper ownership, schema attachments and task-surface backing/dependency
inputs register the new files. `make generate` refreshed only their declared
task-surface/render projections alongside preserved Workbook routing. No
generated topology was hand-edited. The implementation testing guide documents
the release protocol, ordering, recovery and companion investigation.

### Journeys and compatibility

Previously a failed browser group selected `quarantine() ?? destroy()`. A
synchronous successful stop returned undefined, causing a second stop against
the deleted lease and a false secondary cleanup failure. Now the broker chooses
one explicit unhealthy release. Successful teardown preserves the work failure
without adding cleanup_error. A genuinely failed stop retains its rejection,
records cleanup_error, prevents dependent service teardown and retains protected
proof for exact resource-owner recovery. Reobserving release/close does not retry.

Publication is also part of settlement: if acknowledgement fails after physical
stop, the registered original proof remains; a second broker close does not stop
again. Explicit owner recovery is a separate lifetime and validates the exact
lease before acting. No generic destructive janitor or absence-based fallback
was added. Proof shape, identity, containment, owner-only permissions and symlink
rejection are tested. Stale-runtime maintenance skips unresolved recovery roots.

Borrowed services are detached and stay outside provider destruction. Owned
healthy pooling and warm reuse remain supported. The API cutover is internal and
all repository callers migrated; there is no legacy adapter. Historical runs
are inspected as recorded and receive no companion backfill. No application data,
schema, route, saved-view, storage, cursor or visual migration is required. No
dependency was added, and no further golden was refreshed for this repair.

Rollback reverts this harness slice and regenerates derived routing, preserving
the Workbook changes and refreshed golden.

## Verification ledger

Commands run from the repository root. Current task surface and owner rows were
rediscovered through help/task-guide and authored machine inputs. Narrow existing
rows route the new matrices; no hook name became a public Make target.

| Command / selection | Result and retained run root |
| --- | --- |
| `make run-harness-smoke-fast` | PASS, `20261002T053232Z-p12043`; fixture, scheduler and wrapper smoke against final code. |
| `make generate` | PASS, `20261002T053329Z-p15076`; authored-input projections regenerated. |
| `make format` | PASS, `20261002T045559Z-p37915`, 2/2 units; formatter diff inspected, no additional frontend/Go scope. |
| `make agent-finalize` | PASS, `20261002T053556Z-p18472`; RESULTS_DIR unset. Retained-run maintenance skipped: no qualifying successful full warm check. |
| `make harness-contract CARTULARY_HARNESS_CACHE_MODE=off` | PASS, `20261002T054158Z-p24624`, 2/2 execution units, 74/74 contract tests. |
| Affected `harness.browser` slice, cache off | PASS, `20261002T054157Z-p24246`, 6/6 units; fixture suite, UI-review contract/execution/lifecycle, test-services lifecycle rows below. |
| `make harness-ui-review-seeded-default CARTULARY_HARNESS_CACHE_MODE=off` | PASS, `20261002T055639Z-p77465`, exit 0; seeded editor/viewer and artifact workflows, plus final live work/cleanup-failure cases. |
| `make browser-e2e-visual CARTULARY_HARNESS_CACHE_MODE=off` | PASS, `20261002T054655Z-p39835`, 12/12 graph units; default group 46/46 Playwright tests with no retries/flaky results, claimed group passes. All 16 cleanup steps completed. Reconciliation maps all 255 goldens with no missing/orphan entries. Ordinary comparison only; no refresh. |
| `make lint-scripts CARTULARY_HARNESS_CACHE_MODE=off` | PASS, `20261002T054158Z-p24577`, 2/2 units. |
| `make lint-shell CARTULARY_HARNESS_CACHE_MODE=off` | PASS, `20261002T054158Z-p24622`, 4/4 units. |
| `make json-shape-check CARTULARY_HARNESS_CACHE_MODE=off` | PASS, `20261002T054157Z-p24127`, 3/3 units. |
| `make generated-artifact-policy-check CARTULARY_HARNESS_CACHE_MODE=off` | PASS, `20261002T054157Z-p24105`, 3/3 units. |
| `make generate-drift CARTULARY_HARNESS_CACHE_MODE=off` | PASS, `20261002T054157Z-p24099`, 4/4 units. |
| Command-surface registry and catalog routing slices, cache off | PASS, `20261002T050133Z-p65249` and `20261002T050133Z-p65270`, 1/1 each; exact rows below. |
| `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` | PASS, check=pass; scratch generations byte-identical. |
| `make lint-markdown`; `git diff --check` | PASS, `20261002T061346Z-p76211`, Markdown tool summary under `adhoc/lint-markdown`; final diff/scope review passes. |

Run roots above are under `.cartulary/test-results/`. Graph targets publish
`run-summary.json` and `cleanup-results.json`; tool targets publish their named
`tool-run-summary.json`. Use `make explain-run RESULTS_DIR=<exact-root>`.

Exact affected browser rows used with
`make test-slice OWNER=harness.browser ROWS=<comma-separated-rows>` and
`CARTULARY_HARNESS_CACHE_MODE=off`:

- `harness.browser.boundary_support.fixtures_suite_b570a8a829`
- `harness.browser.boundary_support.ui_review_contract`
- `harness.browser.boundary_support.ui_review_execution`
- `harness.browser.boundary_support.ui_review_lifecycle`
- `harness.browser.unit.testservices_lifecycle_contract`

Other affected policy selections:
`harness.command_surface.behavior.public_registry_parity` with owner
`harness.command_surface`, and `harness.test_catalog.behavior.routing_contract`
with owner `harness.test_catalog`.

### Final production lifecycle proof

The final seeded Make run includes
`fixture-cleanup-1790920685705-product` and
`fixture-cleanup-1790920748543-unresolved`. Their retained
`cleanup-results.json`, `review-session.json` and lease records reconcile with
the assertions in `ui-review-fixture-cleanup.mjs` and the passing smoke log.
The inner work graphs retain product/test_assertion_failure and normalized 10;
the surrounding prepared-review record describes its resource-owner lifecycle.

The product case records one completed fixture release followed by completed
service closure, with no cleanup failure. The genuine failure case verifies a
live owned HTTP listener and both browser/suite proofs after the failed release,
then records one failed fixture release and blocked dependent service closure.
It checks closed login-consumer credentials have disappeared before invoking
existing exact-resource recovery. Recovery removes the listener, settles both
proofs and removes the private runtime. The failed receipt remains unchanged
after recovery; it is not relabeled successful.

The public seeded editor/viewer and artifact workflows also assert terminal
cleanup complete, foreground exit 0, repeated stop preserving the same receipt,
and removal of private references. Their caller-owned scratch is removed in a
finally path; expired private images/credentials are not linked here. Borrowed
service preservation is covered by the affected UI lifecycle tests.

Final process inspection finds no foreground review session or controlled
fixture listener remaining. Protected runtime owner markers contain no remaining
roots for this slice's controlled/live cases or final browser runs. Caller-owned
task-surface report scratch was removed. The working-tree review confirms the
preserved frontend diff and golden hash; unrelated resources were not touched.

`make explain-run` succeeds for the final visual root and the cleanup-only
genuine-failure root. It reports zero failed/blocked cleanup steps for the passing
visual run and one failed/one blocked step for the genuine cleanup failure.
Schema validation and secure reads reject unsafe companions.

### Failure history and recovery

- Pre-fix focused roots `20261002T023914Z-p22662` and
  `20261002T024146Z-p24557` reproduce two invocations instead of one. The first
  also exposed an inherited wrapper-input issue; the second isolates the
  regression. Final smoke passes the same permanent assertion.
- Development controlled-process runs exposed Node argument parsing of
  `--env-file`; the process fixture now passes an explicit `--` before its script.
  A diagnostic stop wrapper in the failed development run left one owned
  listener with valid proof. Exact owner recovery validated the protected lease
  and run identity, stopped that listener and removed its runtime. This was a
  recovery action, not a substitute for public Make verification. No private
  handles or paths are committed here.
- Early `agent-finalize` roots `20261002T034752Z-p28807` and
  `20261002T034848Z-p29493`, and contract root `20261002T033335Z-p95836`, rejected
  stale generated source metadata while inputs were changing. Regeneration
  repaired the projections; fresh final checks above pass.
- Final failure-matrix extension at `20261002T045354Z-p31538` proved that an
  undefined thrown value was incorrectly considered completed physical
  settlement. The broker now records failure independently of exception
  truthiness, and ledger attempts normalize undefined/null/zero/empty-string
  rejections before recording them. The permanent regression passes in
  `20261002T045520Z-p34708`; this is part of G2/G5, not a separate audit.
- Ordinary visual comparison `20261002T040501Z-p79230` failed one Evidence
  fixture, `evidence-timeline-evidence-count`, with product/test_assertion_failure
  (10/12 graph units). Inspected expected/actual captures differ in horizontal
  grid position: the actual starts at Analyst while the expected includes Date
  Entered. This is outside the harness repair; its root cause is not established
  by the captures alone. Its receipt records both fixture releases, services and
  private runtime completed, with zero cleanup failures. No golden or fixture was
  changed to conceal this mismatch. Final comparison is recorded separately.
- Comparison `20261002T043159Z-p41272` failed
  `module.workbook.visual.contextual_task_decision_creation` at
  `contextual-decision-recovery`, with 996 differing pixels and
  product/test_assertion_failure. Inspected captures show the recovery-heading
  focus outline absent in the actual image. Its 16-step cleanup companion has
  zero failed or blocked steps. The Evidence fixture from the preceding run
  passed in this run. Neither fixture was modified; their rendering root causes
  remain outside this repair and are not established by capture inspection.
- Seeded smoke `20261002T043159Z-p41679` failed at capture admission with
  infra/resource_conflict, diagnostic capacity_exceeded, while checks I started
  concurrently were consuming capacity. The public workflow's finally path
  stopped its session. An isolated rerun passed at `20261002T044140Z-p33907`,
  including live cases `fixture-cleanup-1790916186338-product` and
  `fixture-cleanup-1790916249372-unresolved`. Required final-source browser work
  is serialized; its fresh results are recorded above.
- Ordinary visual comparison `20261002T050319Z-p72539`, run without concurrent
  browser suites, again failed
  `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`
  at `evidence-timeline-evidence-count`: 27,793 pixels differ (the fixture permits
  8,000). Its actual PNG is byte-identical to the inspected actual from
  `20261002T040501Z-p79230`; it starts at Analyst instead of Date Entered. All
  other Playwright tests in that group pass (45/46), including the earlier
  contextual Decision fixture. Both fixture leases, services and private runtime
  complete cleanup; the run retains product/test_assertion_failure and no false
  secondary cleanup_error. The subsequent final-source comparison
  `20261002T054655Z-p39835` passes all 12 units, including this fixture, without
  fixture or golden changes.

The refreshed Workbook golden remains SHA-256
`9c6ab774fd3d95456caea9efb22cf33c269c052a3f4dde5c0b721527dc375692`.
The unrelated Evidence and Decision fixtures/goldens were not changed. The final
full comparison passes; the earlier scroll-position and focus-outline variation
remains documented, with no rendering root cause claimed. One passing execution
does not establish stability across future runs. Investigating that variation is
outside this completed lifecycle repair.

### Scoped acceptance and limitations

Acceptance applies to this lifecycle slice, not the complete subsystem matrix.

| Testing Harness acceptance | State | Evidence / unchanged scope |
| --- | --- | --- |
| AC-007 service lifetimes | PASS scoped | Ownership, independent finalizers, taint, production provider smoke and final live teardown/recovery. Backend eager database/bucket performance thresholds are unchanged and were not remeasured. |
| AC-010 proof gates | PASS scoped | Exact proof, unsafe identity/containment/symlink rejection and unresolved-root deletion refusal; existing stale-owner contract matrix. No destructive generic janitor introduced. |
| AC-014 exit matrix | PASS scoped | Product failure plus cleanup success/failure, cleanup-only 12, deadline 13, thrown execution failure and all secondary failures. Full contract matrix passes. |
| AC-015 artifact identity/private lifetime | PASS scoped | Closed companion, safe identities/references, secret sentinel exclusion, protected proof, purged closed-consumer detail and final seeded cleanup receipts. |
| AC-016 editorial boundary | PASS scoped | Human owner clarification review complete; no executable Markdown dependency. Markdown maintenance is recorded separately and is not product conformance. |
| AC-017 lifecycle | PASS scoped | Current test-services lifecycle row passes in the 6/6 affected slice; fixture/graph memoized terminal outcomes and retained failure proofs covered. Go lifecycle schema and state transitions unchanged. |
| AC-032 primary determinism | PASS scoped | Earlier work/interruption remains primary; cleanup-only failures use ordered finalizer sequence and existing selector. Concurrent/failure matrix assertions pass. |
| AC-077 graph/browser scheduling | PASS scoped | Actual runWorkGraph matrix, late terminal/cache admission, final live recovery, full visual comparison and cleanup of all started visual stacks pass. Scheduling capacity policy is unchanged. |
| AC-086 broker | PASS scoped | Owned/borrowed/none, warm/shared/pooled, synchronous/asynchronous, taint, contradictory repetition, publication rollback and production process closure. Object-store semantics and probe contracts unchanged. |

Frontend product unit, backend-wide, release and performance suites are skipped:
this repair changes harness lifetime/evidence only. The existing Workbook code
and refreshed golden are preserved; the required current visual comparison
passes. Historical artifacts remain immutable. No full warm-check evidence is
claimed, and no visual or Axe result is treated as accessibility certification
or a Core 05 publication claim.
