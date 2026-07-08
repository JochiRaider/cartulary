# Test Harness Subsystem Update Tracker

Target artifact: `docs/handoffs/test-harness-subsystem-update-tracker.md`

Planning target: Cartulary test harness subsystem refactor tracker focused on
reducing warm steady-state wall time for `make check`, especially the
`check-service-backed` critical path.

## 1. Session Context

- Branch: `main`
- Commit: `a88fc1bf6667aed22277af38f0df2d8747b89271`
- Dirty tree: clean by `git status --short --branch`; branch is ahead of
  `origin/main` by 18 commits before this tracker reconciliation.
- Date: 2026-07-08, `America/New_York`
- Primary retained evidence inspected:
  `.cartulary/test-results/20260708T224533Z-p1093`
- Prior accepted retained evidence inspected:
  `.cartulary/test-results/20260708T221258Z-p12791`
- Initial pre-remediation baseline retained for comparison:
  `.cartulary/test-results/20260708T205356Z-p94648`
- Source files inspected: `AGENTS.md`, `docs/testing-harness-nlspec.md`,
  `docs/domain.md`,
  `docs/handoffs/cartulary_modular_refactor_planning_framework.md`,
  `docs/spec/00_document_set_status_and_precedence.md`,
  `docs/spec/04_security_deployment_and_conformance.md`, `Makefile`,
  `tools/task_surface_manifest.json`, `tools/execution_topology_manifest.json`,
  `tools/scheduler_manifest.json`, duration baseline manifests, phase/frontend
  phase maps, `tools/harness/**` inventory, retained run summaries, and prior
  harness trackers.
- Source limits: the current request reran focused harness checks and a final
  warm `make check`. Final retained timing evidence is warm-attributed and
  usable for handoff validation; the initial baseline remains diagnostic only
  because readiness attribution is empty in that older pressure summary.
- This plan is only for the test harness subsystem and wall-time reduction. It
  does not propose product behavior changes except through explicit spec-first
  gates where default evidence selection may change.
- Implementation plan adopted from the 2026-07-08 remediation request. This
  tracker is the controlling execution artifact. It MUST be updated after each
  workstream completes and before the next workstream begins.
- Current execution state: W0 through W8 are complete. This reconciliation
  updates stale aggregate status text only; no code, schema, phase-map, or
  generated-output edits are introduced by the reconciliation itself.

## 2. Scope And Non-Scope

Primary seam: warm local `check` and `check-service-backed` scheduling,
readiness attribution, and default evidence selection.

Behavior-preserving refactors:

- Move scheduler/readiness/timing callers toward owner facades and semantic
  ownership instead of private helper paths.
- Make hidden readiness/provisioning visible in scheduler diagnostics without
  changing required evidence.
- Rebalance scheduler lanes through owner inputs while preserving public targets,
  command IDs, summary schemas, cleanup, and failure classification.
- Improve pressure summaries and bottleneck reports for future phase growth.

Proposed behavior changes, spec-first:

- Removing or reclassifying default local `check` rows, browser projections,
  support-only rows, or duplicate evidence.
- Changing public Make target behavior, stable `command_id` values, schema IDs,
  retained paths, cleanup behavior, public inputs, or failure contracts.
- Changing fixture lifecycle policy for authoritative evidence rows.

Explicit exclusions:

- No caching for security scans, drift verdicts, generated-artifact drift
  detection, service-backed/browser/live-state tests, cleanup, destructive reset,
  runtime reset, or aggregate success.
- No cross-run product-test or scheduler-result reuse in this tracker.
- No hand edits to generated files.
- No phase-shaped runtime dependency.
- No product-performance claim or Core 05 publication claim from harness timing.

## 3. Current-State Inventory

| Path | Current responsibility | Target owner/facade | Public/private status | Contracts touched | Timing relevance | Risk | Test evidence |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `docs/testing-harness-nlspec.md` | Normative harness mechanics | Harness owner spec | Public owner doc | Target selection, scheduling, caches, artifacts, cleanup | High | Spec drift if bypassed | Inspected |
| `docs/domain.md` | Vocabulary and boundary hygiene | Domain reference | Public reference | Domain terms only | Low | Terminology drift | Inspected |
| `Makefile` | Public command facade | Make public targets | Public surface | Target names, public inputs, cleanup | High | Compatibility break | `make help`, explain targets |
| `tools/task_surface_manifest.json` | Task surface contract/mirror input | Task-surface owner inputs/generator | Compatibility surface | Public targets, default inclusion, command metadata | High | Default over-selection or drift | Inspected |
| `tools/execution_topology_manifest.json` | Schedule/topology owner input | Execution topology owner | Owner input | Resource profiles, stage/session policy, generated outputs | High | Bad lane balance | Inspected |
| `tools/scheduler_manifest.json` | Generated scheduler schedule | Generated output | Generated, do not hand-edit | Work units, resource limits | High | Manual drift | Inspected only |
| `tools/task_surface.generated.mk` | Generated Make surface | Generated output | Generated, do not hand-edit | Make wiring | Medium | Manual drift | Inventory only |
| `tools/browser_e2e_batch_manifest.json` | Generated browser batch plan | Generated output | Generated, do not hand-edit | Browser grouping/shards | High | Manual drift | Inventory/partial |
| `tools/phase*_test_map.json` | Phase evidence owner maps | Phase owner inputs | Public-ish owner input | Default rows, fixture policy, evidence ownership | High | Product evidence loss | Inspected through plans |
| `tools/frontend_phase_registry.json` | Frontend evidence owner map | Frontend phase owner | Owner input | Browser/frontend evidence | Medium | Over/under-selection | Inventory |
| `tools/*duration_baselines.json` | Duration baseline inputs | Duration accounting owner | Support input | Timing drift and weights | Medium | Stale scheduling weights | Inspected |
| `tools/scheduler_resource_registry.json` | Resource vocabulary | Scheduler resource owner | Owner input | Resource claims/limits | Medium | False contention model | Inventory |
| `tools/harness/**` | Helper implementation | Owner facades by subsystem | Private unless promoted | Helper paths not compatibility | Medium | Accidental compatibility burden | Inventory |
| `.cartulary/test-results/20260708T205356Z-p94648/**` | Retained run evidence | Harness summaries | Retained artifact contract | Summary schemas and paths | High | Misread cold vs warm evidence | Inspected |

Generated outputs identified separately:

- `tools/scheduler_manifest.json`
- `tools/task_surface.generated.mk`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- Generated harness/topology outputs listed by
  `tools/execution_topology_manifest.json`

Public compatibility surfaces that must not drift without spec-first routing:

- Public Make targets and public input variables.
- Stable `command_id` values.
- Summary, scheduler, pressure, fixture, cache, and event schema IDs.
- Retained artifact paths and cleanup behavior.
- Failure classification and aggregate summary contracts.

## 4. Timing Baseline And Bottleneck Model

Initial diagnostic baseline:

- Run root: `.cartulary/test-results/20260708T205356Z-p94648`
- `make check`: `138.03s`
- Slowest target: `check-service-backed`, `126.87s`
- Tests: 982 total, 0 failed, 273 executed accounting rows, 0 reused, 0
  skipped.
- `check-service-backed` children:
  - `backend-process`: `120.44s`
  - `backend-integration`: `106.38s`
  - `browser-e2e-stateful`: `91.91s`
  - `backend-store`: `72.33s`
  - `browser-e2e-webserver-backed`: `59.57s`
  - `build-operator`: `3.38s`

Critical path model:

- Initial path includes frontend install/build readiness, embedded web assets,
  server build, browser session startup, and a long webserver-backed browser
  shard.
- Browser webserver-backed stage completes around the middle of the run, but
  `check-service-backed` completion waits on backend-process/stateful/backend
  tails.
- Top individual slow units include browser functional shards around 50-54s and
  backend-process SeaweedFS support rows around 47-50s.
- Resource blockers are dominated by `host_io`, then service-session readiness,
  `host_cpu`, `postgres_clone`, and `process`.
- Fixture pressure remains material: template clone and transaction/shared
  PostgreSQL rows dominate backend integration/store/process pressure.
- Browser stack observed max was below limit, but stateful browser sessions are
  isolated by design and capped separately.
- Readiness/provisioning attribution is incomplete in this initial baseline:
  retained pressure summary has empty `readiness_attribution_counts`, so this
  run remains useful for bottleneck modeling but not final warm
  scheduler-health acceptance.

Accepted warm retained baseline:

- Run root: `.cartulary/test-results/20260708T221258Z-p12791`
- Final warm `make check` passed.
- `check-service-backed`: `122851ms`, below the `155000ms` compatibility cap.
- `check/pressure-summary.json` uses
  `cartulary.scheduler_pressure_summary.v4`.
- `check/pressure-summary.json` reports readiness attribution counts:
  `build_artifact=4`, `embedded_web_assets=1`, `frontend_install=1`,
  `service_image=1`, `test_service_binary=1`, and `toolchain=5`.
- Nested `check-service-backed` scheduler artifacts are present:
  `scheduler-summary.json`, `scheduler-events.jsonl`, and
  `pressure-summary.json`.
- Retained-run finalizer passed at
  `.cartulary/test-results/20260708T221519Z-p10405`.
- Timing drift check passed at
  `.cartulary/test-results/20260708T221614Z-p12628`.

Current request final warm retained baseline:

- Run root: `.cartulary/test-results/20260708T224533Z-p1093`
- Final warm `make check` passed in `129553ms`.
- `check-service-backed`: `119075ms`, below the `155000ms` compatibility cap.
- Tests: 980 total, 0 failed.
- `check/pressure-summary.json` uses
  `cartulary.scheduler_pressure_summary.v4`.
- `check/pressure-summary.json` reports readiness attribution counts:
  `build_artifact=4`, `embedded_web_assets=1`, `frontend_install=1`,
  `service_image=1`, `test_service_binary=1`, and `toolchain=5`.
- Nested `check-service-backed` scheduler artifacts are present:
  `scheduler-summary.json`, `scheduler-events.jsonl`, and
  `pressure-summary.json`.
- Retained-run finalizer passed at
  `.cartulary/test-results/20260708T224752Z-p1348` and refreshed six generated
  duration/schedule artifacts.
- Timing drift check passed at
  `.cartulary/test-results/20260708T224819Z-p3515`.

Minimal measurement plan if retained evidence must be refreshed:

- Run a warm-ready `make check` after provisioning is already complete.
- Expected artifacts:
  - `.cartulary/test-results/<run-id>/tool-run-summary.json`
  - `.cartulary/test-results/<run-id>/run-summary.json`
  - `.cartulary/test-results/<run-id>/check/scheduler-summary.json`
  - `.cartulary/test-results/<run-id>/check/pressure-summary.json`
  - `.cartulary/test-results/<run-id>/check/scheduler-events.jsonl`
  - `.cartulary/test-results/<run-id>/check-service-backed/target-summary.json`
- Finalize retained evidence with
  `make agent-finalize RESULTS_DIR=.cartulary/test-results/<run-id>`.

## 5. Refactor Tracker

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| THT-WT-001 | Establish accepted warm baseline and readiness attribution | Measurement | COMPLETE | None | Harness timing owner | `.cartulary/test-results/20260708T224533Z-p1093` | Final warm run reports readiness attribution and `check-service-backed=119075ms` |
| THT-WT-002 | Audit default `check` selection for support-only, duplicate, and high-cost rows | Selection | COMPLETE | THT-WT-001 | Harness/spec owners | `make target-plan-json`; phase maps | Default rows have metadata and support-only SeaweedFS rows are explicit-only |
| THT-WT-003 | Expose hidden readiness/provisioning as first-class pressure attribution | Diagnostics | COMPLETE | THT-WT-001 | Scheduler/readiness owner | `check/pressure-summary.json` | Readiness costs are visible for warm scheduler-health conclusions |
| THT-WT-004 | Reduce backend-process terminal tail, especially SeaweedFS support rows | Scheduling/selection | COMPLETE | THT-WT-002 | Backend harness owner | Scheduler summary and row evidence | `check` and `check-service-backed` schedules exclude SeaweedFS migration-support units |
| THT-WT-005 | Rebalance browser webserver-backed and stateful stages by owner metadata | Browser scheduler | COMPLETE | THT-WT-001 | Browser harness owner | Browser batch manifest and scheduler events | Reviewed as no-op; retained evidence did not justify shard/session changes |
| THT-WT-006 | Target fixture tier proof work to top pressure rows only | Fixture lifecycle | COMPLETE | THT-WT-001 | Fixture owner | Fixture proof artifacts | Reviewed as no-op; no row-specific proof justified unsafe fixture downgrades |
| THT-WT-007 | Tune topology resource claims and lane priorities from owner inputs | Scheduler | COMPLETE | THT-WT-003 | Execution topology owner | Generated scheduler manifest | Reviewed as no-op; retained evidence did not justify oversubscription risk |
| THT-WT-008 | Remove duplicate support-only default evidence where owner docs allow | Selection/spec | COMPLETE | THT-WT-002 | Harness spec owner | NLSpec and target plan | Duplicate/support-only local evidence is explicit-only unless owner-promoted |
| THT-WT-009 | Clean private helper path coupling in touched scheduler/readiness code | Structure | COMPLETE | Related edits | Harness module owners | Import/path checks | Touched callers use owner facades; no private compatibility alias was promoted |
| THT-WT-010 | Validate final timing and retained-run maintenance | Verification | COMPLETE | Implemented workstreams | Harness verification owner | `.cartulary/test-results/20260708T224533Z-p1093` plus finalizer roots | Final warm run, retained-run finalizer, timing drift, and harness contract passed |

### 5.1 Active Implementation Workstreams

| Workstream | Depends on | Status | Scope | Files changed | Validation / retained run roots | Exit criteria |
| --- | --- | --- | --- | --- | --- | --- |
| W0 Tracker adoption | none | COMPLETE | Adopt the remediation plan, add newly found gaps, and mark W1 active. | `docs/handoffs/test-harness-subsystem-update-tracker.md` | Tracker self-update only. Retained observational root remains `.cartulary/test-results/20260708T205356Z-p94648`. | Tracker is current before any code/spec edits. |
| W1 Spec and schema closure | W0 | COMPLETE | Revise NLSpec and schema attachments for scheduler manifest v2, pressure summary v4, readiness metadata, target-plan metadata closure, and nested artifact acceptance. | `docs/testing-harness-nlspec.md`; `tools/schemas/cartulary.scheduler_manifest.v2.schema.json`; `tools/schemas/cartulary.scheduler_pressure_summary.v4.schema.json`; `tools/harness/generated-artifacts/tests/test-json-shapes.sh`; `tools/harness/tests/test-harness-contracts.mjs` | `node -e` JSON parse for new schema files passed. Full `json-shape-check` is intentionally deferred until W2 migrates producers and generated manifests to the current schema IDs. | Spec names v2/v4 as current; v1/v3 remain readable for old retained diagnostics; schema fixtures cover the new readiness fields. |
| W2 Diagnostics/artifact implementation | W1 | COMPLETE | Implement readiness metadata propagation, pressure v4 emission, nested `check-service-backed` artifacts, and target-plan metadata fields. | `tools/harness/scheduler/scheduler-manifest.mjs`; `tools/harness/generated-artifacts/execution-topology.mjs`; `tools/harness/execution/service-backed/schedule-expansion.mjs`; `tools/harness/execution/service-backed/schedule-go-planning.mjs`; `tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs`; `tools/scheduler_manifest.json`; `tools/execution_topology_render_index.json`; `tools/harness/scheduler/scheduler/engine.mjs`; `tools/harness/scheduler/scheduler/pressure-summary.mjs`; `tools/harness/diagnostics/scheduler-summary-timing-drift-cli.mjs`; `tools/harness/backend/target-plan.mjs`; scheduler/duration test fixtures | `make phase-schedules` passed at `.cartulary/test-results/20260708T215409Z-p22236`; `make json-shape-check` passed at `.cartulary/test-results/20260708T215958Z-p26365`; `make check-harness-smoke` passed with run root `.cartulary/test-results/20260708T215945Z-p25282`; `make check` passed at `.cartulary/test-results/20260708T215651Z-p26103` and confirmed nested artifacts existed/schema-validated, with nested total-count correction applied afterward. `make target-plan-json` inspection found 313/313 phase rows with complete default-check metadata. | Focused scheduler tests pass; v4 pressure readiness fields are emitted; nested `check-service-backed` artifacts are materialized from scheduler records; generated manifest is v2. |
| W3 Default-selection cleanup | W2 | COMPLETE | Reclassify Phase 10 SeaweedFS migration-preservation support rows explicit-only and ensure release gate uses current-run migration evidence. | `docs/testing-harness-nlspec.md`; `tools/phase10_test_map.json`; regenerated `tools/scheduler_manifest.json`; regenerated `tools/execution_topology_render_index.json` | `make phase-schedules` passed at `.cartulary/test-results/20260708T220114Z-p27100`; `make target-plan-json` inspection showed the two SeaweedFS support rows `default_check_required=false`; generated `check` and direct `check-service-backed` schedules contain zero `backend-process-seaweedfs-migration-support` units; `make run-harness-smoke-extended` passed at `.cartulary/test-results/20260708T220147Z-p27695`; `make phase-schedule-drift` passed at `.cartulary/test-results/20260708T220343Z-p89804`. | Default local `check` excludes support-only SeaweedFS migration-preservation rows; release-evidence characterization still passes. |
| W4 Backend tail and fixture proof | W3 | COMPLETE | Use pressure data to address remaining backend-process top rows; reduce fixture tiers only with row-specific proof. | No additional owner-input changes beyond W3; fixture tiers intentionally unchanged because no row-specific `fixture_tier_proof.v1` evidence supported a safe downgrade for remaining Phase 10 product rows. | `make service-backed-slice PHASE=phase10` passed at `.cartulary/test-results/20260708T220427Z-p90205`; post-W3 `make check` passed at `.cartulary/test-results/20260708T220604Z-p6523`; nested `check-service-backed` summary/pressure totals were validated at 232/232; `check/pressure-summary.json` slowest units contained zero SeaweedFS migration-support units. | Product Phase 10 recovery/operator rows remain selected; support-only SeaweedFS rows are out of default pressure; no unsafe fixture reduction was made. |
| W5 Browser scheduling balance | W3 | COMPLETE | Rebalance browser batch/stateful owner metadata only if accepted warm evidence still shows browser lane skew. | No browser owner-input changes; evidence did not justify adding shards or sessions. | Post-W3 `make check` root `.cartulary/test-results/20260708T220604Z-p6523` showed the five slowest webserver-backed functional shards at 55.14s to 60.86s, about 5.72s spread, with row/session accounting preserved by the passing full check. | Browser balance reviewed; no topology change made because lane skew was not the limiting problem. |
| W6 Resource/topology tuning | W4, W5 | COMPLETE | Tune narrow resource claims and priority bands in `tools/execution_topology_manifest.json`; regenerate downstream scheduler artifacts. | No topology owner-input changes. Broad `host_io`/`host_cpu` increases were not made because W5 showed browser work, not resource-lane skew, as the remaining tail and no retained evidence justified oversubscription risk. | Post-W3 `make check` `.cartulary/test-results/20260708T220604Z-p6523` passed; top blockers were lower than the original retained run and no service/readiness flakes were observed. | Resource tuning reviewed; no unsafe broad resource change introduced. |
| W7 Facade cleanup | related touched files | COMPLETE | Move touched imports to owner facades; delete unsupported private compatibility paths only after callers move. | No code changes required beyond W2 imports already using `tools/harness/backend/backend-target-plan.mjs` and `tools/harness/backend/backend-shard-plan.mjs` facades from touched scheduler/diagnostics code. | Import audit: touched scheduler pressure code imports backend facades, not private `target-plan.mjs`; later final validation includes `make harness-contract`. | Touched paths remain on owner facades; no legacy private redirect promoted. |
| W8 Final validation and handoff | all prior | COMPLETE | Run finalizer, final warm `make check`, retained-run finalizer, and complete tracker handoff. | `docs/handoffs/test-harness-subsystem-update-tracker.md`; retained-run finalizer refreshed `tools/browser_e2e_duration_baselines.json`, `tools/go_test_duration_baselines.json`, `tools/harness_smoke_duration_baselines.json`, and `tools/service_backed_make_target_duration_baselines.json`. | `make agent-finalize` passed at `.cartulary/test-results/20260708T221239Z-p11346`; final warm `make check` passed at `.cartulary/test-results/20260708T221258Z-p12791` with `check-service-backed=122851ms`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260708T221258Z-p12791` passed at `.cartulary/test-results/20260708T221519Z-p10405`; `make scheduler-summary-timing-drift RESULTS_DIR=.cartulary/test-results/20260708T221258Z-p12791 TARGET=check` passed at `.cartulary/test-results/20260708T221614Z-p12628`; final `make harness-contract` passed at `.cartulary/test-results/20260708T221623Z-p12717`. Earlier W8 checks also passed: `generated-artifact-policy-check`, `json-shape-check`, `phase-schedule-drift`, and `generate-drift`. | Handoff complete; final retained run is warm-attributed, under the 155000ms compatibility cap, and finalized. |

### 5.2 Current Request Verification Pass

| Workstream | Status | Current-pass evidence | Next step |
| --- | --- | --- | --- |
| W0 Tracker reconciliation | COMPLETE | Tracker stale aggregate status rows reconciled to detailed W0-W8 completion state; final retained run elevated as accepted warm evidence. | W1 verified. |
| W1 Spec and schema closure | COMPLETE | `cartulary.scheduler_manifest.v2` and `cartulary.scheduler_pressure_summary.v4` schema files parsed with `node -e`; NLSpec names v2/v4, target-plan metadata, nested artifacts, and helper facade ownership. | W2 verified. |
| W2 Diagnostics/artifact implementation | COMPLETE | `tools/scheduler_manifest.json` declares `cartulary.scheduler_manifest.v2` and contains readiness attribution; retained final run contains v4 pressure summaries and first-class `check-service-backed` scheduler artifacts. | W3 verified. |
| W3 Default-selection cleanup | COMPLETE | `make target-plan-json` stream-safe inspection found 313 phase rows with complete default-check metadata and 0 missing metadata rows; the two SeaweedFS migration-support rows are `explicit_only`; generated `check` and `check-service-backed` schedules contain 0 `backend-process-seaweedfs-migration-support` units. | W4 verified. |
| W4 Backend tail and fixture proof | COMPLETE | Final retained `check` and `check-service-backed` pressure summaries contain 0 SeaweedFS migration-support slowest units; `check-service-backed` completed 232/232 work units; fixture proof counts are 168 tier proofs, 8 accepted proof records, and 0 blocked proof records; retained Phase 10 `service-backed-slice` passed at `.cartulary/test-results/20260708T220427Z-p90205`. | W5 verified. |
| W5 Browser scheduling balance | COMPLETE | Final retained `check` pressure summary shows the five slowest browser functional shards at 55.658s to 61.507s, about 5.849s spread; `tools/browser_e2e_batch_manifest.json` remains `cartulary.browser_e2e_batch_manifest.v5` with stateful partitioning preserved. No browser owner-input change is justified. | W6 verified. |
| W6 Resource topology and facade cleanup | COMPLETE | Final retained `check` scheduler summary reports passing execution with current resource limits and no oversubscription failures; top blockers are diagnostic pressure only (`host_io`, service-session dependency, `host_cpu`, `postgres_clone`, and stateful browser resource). Import audit found scheduler/diagnostics/generated-artifact callers using `backend-target-plan.mjs` and `backend-shard-plan.mjs` owner facades, with no non-test imports of private `backend/target-plan.mjs`. | W7 verified. |
| W7 Final validation and handoff | COMPLETE | Surface checks passed (`make help`, `explain-target` for `check` and `check-service-backed`, `target-plan-json` metadata assertion). Drift/schema checks passed: `generated-artifact-policy-check` `.cartulary/test-results/20260708T224849Z-p4835`, `json-shape-check` `.cartulary/test-results/20260708T224840Z-p4162`, `phase-schedule-drift` `.cartulary/test-results/20260708T224840Z-p4164`, `phase-ledger-drift` `.cartulary/test-results/20260708T224901Z-p6012`, and `generate-drift` `.cartulary/test-results/20260708T224849Z-p4784`. Focused checks passed: `check-harness-smoke` `.cartulary/test-results/20260708T223426Z-p22719`, `run-harness-smoke-extended` `.cartulary/test-results/20260708T224150Z-p22092`, and `service-backed-slice PHASE=phase10` `.cartulary/test-results/20260708T224408Z-p84015`. Final checks passed: `agent-finalize` `.cartulary/test-results/20260708T224520Z-p99318`, warm `make check` `.cartulary/test-results/20260708T224533Z-p1093`, retained-run finalizer `.cartulary/test-results/20260708T224752Z-p1348`, timing drift `.cartulary/test-results/20260708T224819Z-p3515`, and `harness-contract` `.cartulary/test-results/20260708T224826Z-p3660`. | Handoff complete. |

Current-request implementation delta: reconciled this tracker; widened the
scheduler priority-reservation smoke fixture blocker sleep in
`tools/harness/scheduler/tests/test-check-scheduler.sh` from `0.2s` to `1s`
so the wrapper-level extended smoke proof is not sensitive to process-start
jitter; and accepted Make/finalizer-generated refreshes to duration baseline
inputs plus downstream scheduler/topology artifacts.

## 6. Gap Analysis And Recommendations

| Gap ID | Gap | Remediation | Area: spec / implementation / tests / docs / multiple | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| G-001 | Warm-readiness attribution is missing from retained pressure evidence | Add spec/schema support for scheduler-owned timing metadata and emit readiness/provisioning attribution in pressure summaries using `cartulary.scheduler_manifest.v2` and `cartulary.scheduler_pressure_summary.v4`; keep v1/v3 readable for old runs. | spec / schema / implementation / tests / docs | Warm health cannot accept hidden provisioning as scheduler cost, and current v3 explicitly forbids non-empty attribution without scheduler-readable source metadata. | Clear timing evidence across machines; warm health becomes explainable and portable. | Public schema revision; old retained runs remain diagnostic only. | Future optimizations chase cold setup noise. | `pressure-summary.json` shows readiness counts, durations, and units; warm checker rejects over-threshold readiness by declared class. |
| G-002 | `check-service-backed` terminal wait is backend-process/stateful dominated | Rebalance or reclassify high-cost tail rows through owner metadata | multiple | Browser is not the only critical path | High wall-time reduction potential | Default row removals are spec-first | Run stays near 140s despite browser tuning | Critical path no longer waits on unjustified tail work |
| G-003 | Heavy support-only SeaweedFS rows appear in default local path | Audit continuing value; move out of default if not local correctness evidence | spec / implementation / tests | The optimization goal rejects legacy preservation without continuing value | High if removed or narrowed | Public/default behavior change requires NLSpec and manifest update | Local check pays CI-style support cost forever | Target plan shows support rows either justified or excluded |
| G-004 | Browser default projection may be over-broad or poorly grouped | Audit browser projection rows, shard weights, and stateful session grouping | multiple | More shards alone was previously rejected; owner-row clarity is needed | Medium | Generated browser batch refreshed only through Make | Browser stage remains noisy and hard to grow | Browser lanes meet balance target with stable failure reporting |
| G-005 | Fixture clone/reset pressure remains high but proof closure blocks changes | Limit proof work to top pressure rows and require row-specific isolation evidence | tests / implementation | Broad fixture lowering is unsafe without proof | Medium | Fixture policy changes touch phase maps and summaries | Database pressure remains dominant | Fixture counts/duration drop with accepted proofs |
| G-006 | Resource model reports `host_io`, clone, and process contention but lacks clear remediation loop | Tune resource claims, priorities, and limits through topology owner inputs after measurement | implementation / tests | Lane balance needs accurate resource semantics | Medium | Regenerate scheduler outputs; no public contract drift | Scheduler keeps serializing independent work | Pressure blockers decrease without flakes |
| G-007 | Private helper path history can become accidental compatibility | Move touched callers to owner facades; delete temporary redirects instead of preserving them | implementation / docs | NLSpec says private helpers are implementation details | Low wall time, high maintainability | No public migration unless promoted by NLSpec | Future refactors inherit false compatibility | Import-boundary and harness contract checks pass |
| G-008 | Duplicate default evidence across harness smoke, support rows, and full targets is not fully explained | Add default inclusion reason audit using existing owner metadata | docs / spec / tests | Default local check should stay correctness-focused | Medium | Spec-first for changed defaults | Default check expands with every phase | Every default row has `default_check_reason_code` and non-duplicate value |
| G-009 | Generated topology outputs are easy to hand-edit accidentally | Require owner-input edits plus generator/drift validation at each checkpoint | docs / tests | Generated outputs are downstream | Low direct timing, high safety | No hand-edited generated files | Schedule drift becomes unreproducible | Generated-artifact and schedule drift checks pass |
| G-010 | `check-service-backed` lacks first-class nested scheduler artifacts | Materialize schema-valid `check-service-backed/scheduler-summary.json`, `check-service-backed/scheduler-events.jsonl`, and `check-service-backed/pressure-summary.json` from scheduler records rather than pointer-only aliases. | implementation / tests / docs | NLSpec and `explain-target` already require direct diagnosability for nested scheduler targets. | Critical path diagnosis does not require mining parent `check` artifacts; finalizer eligibility is unambiguous. | Fills already-declared retained paths; no target rename. | Retained-run evidence remains incomplete and investigations miss child scheduler state. | Full `make check` run contains valid nested artifacts with parent linkage. |
| G-011 | `target-plan-json` omits default-check metadata named by the NLSpec | Propagate `default_check_kind`, `default_check_reason_code`, `evidence_delta`, `warm_local_cost_class`, optional `default_check_reason`, and related closure fields into public target-plan JSON for phase-map rows. | implementation / tests | Default-selection audits need complete owner metadata at the public diagnostic surface. | Default local check rows become reviewable without hand-inspecting phase maps. | Adds missing contract fields; consumers receive more complete JSON. | Default rows cannot be audited from public command output. | `make target-plan-json` shows non-null metadata for all phase-map rows. |

## 7. Prioritized Roadmap And Closure State

P0:

- THT-WT-001 / G-001: accepted warm baseline and readiness attribution.
  Complete with final warm run
  `.cartulary/test-results/20260708T224533Z-p1093`.
- THT-WT-002 / G-003 / G-008: default-selection audit, especially support-only
  and duplicate high-cost rows. Complete; Phase 10 SeaweedFS
  migration-preservation support rows are explicit-only.
- THT-WT-004 / G-002: backend-process terminal-tail reduction. Complete for
  the identified support-only tail rows; product recovery/operator rows remain
  selected.
- THT-WT-007 / G-006: scheduler lane/resource tuning from owner inputs. Effect:
  reviewed as no-op because retained evidence did not justify oversubscription
  risk.

P1:

- THT-WT-005 / G-004: browser projection and session grouping balance. Effect:
  reviewed as no-op because retained evidence did not show material lane skew.
- THT-WT-006 / G-005: top-row fixture proof reductions. Effect: medium.
  Reviewed as no-op until row-specific proof exists.
- THT-WT-009 / G-007: owner facade cleanup in touched harness areas. Effect:
  low wall-time, high maintainability. Complete for touched scheduler and
  diagnostics paths.

P2:

- Future proof schemas for more aggressive fixture or scheduler reuse, if owner
  docs adopt them. Effect: unknown.
- CI/release split refinements for evidence that is valuable but not
  default-local. Effect: low to medium.
- Duration baseline maintenance automation improvements after P0/P1 stabilize.
  Effect: low.

## 8. Checkpoint Plan

| Checkpoint | Edit scope | Validation | Expected diff | Rollback point |
| --- | --- | --- | --- | --- |
| CP-0 | Tracker artifact only | Markdown/docs checks only | New handoff doc | Remove tracker doc |
| CP-1 | Measurement/readiness diagnostics, no selection change | `make json-shape-check`, harness smoke, explain-run on retained root | Pressure/readiness summaries only | Revert diagnostics |
| CP-2 | Spec-first default-selection changes | NLSpec review, target-plan JSON, task-surface parity | Owner docs and owner inputs | Revert selection owner inputs/spec |
| CP-3 | Topology/resource lane tuning | Schedule generation, phase-schedule drift, scheduler summaries | Topology input plus generated outputs | Revert topology input and regenerated outputs |
| CP-4 | Targeted fixture proof changes | Phase/service-backed slices for affected rows | Phase maps/proof artifacts | Revert affected row policy |
| CP-5 | Final warm verification | Warm `make check`, `make agent-finalize RESULTS_DIR=<run-root>` | Retained run evidence | Revert last checkpoint causing regression |

Each checkpoint must stay reviewable and avoid mixing unrelated behavior, schema,
generated-output, and documentation changes unless the changes are inseparable.

## 9. Validation Plan

Use the narrowest Make-owned targets that prove each checkpoint.

Planning and surface inspection:

- `make help`
- `make explain-target TARGET=check DETAIL=summary`
- `make explain-target TARGET=check-service-backed DETAIL=summary`
- `make target-plan-json`
- `make explain-run RESULTS_DIR=.cartulary/test-results/<run-id>`

Contract and generated-output safety:

- `make generated-artifact-policy-check`
- `make json-shape-check`
- `make phase-schedule-drift`
- `make phase-ledger-drift`
- `make harness-contract`
- `make frontend-import-boundary-check` when frontend/browser surfaces are touched

Focused behavior checks:

- `make check-harness-smoke`
- `make phase-slice PHASE=<affected-phase>`
- `make service-backed-slice PHASE=<affected-phase>`
- `make browser-e2e-webserver-backed` only for browser projection changes
- `make browser-e2e-stateful` only for stateful browser changes
- `make check-service-backed` after schedule or fixture changes

End-of-run:

- Run `make agent-finalize` before broader verification.
- If retained warm full-check evidence is required, run
  `make agent-finalize RESULTS_DIR=.cartulary/test-results/<successful-full-warm-check-run>`.
- If no retained run is available, retained-run maintenance is skipped because
  `RESULTS_DIR` is unset.

## 10. Documentation And Generated Artifact Plan

- NLSpec changes required for public behavior changes: default selection, public
  target behavior, cache semantics, schema IDs, artifact paths, cleanup
  behavior, failure contracts, fixture lifecycle rules.
- Implementation-support docs/tracker changes are sufficient for private helper
  cleanup and non-contractual diagnostics.
- Owner-input updates may include `tools/execution_topology_manifest.json`,
  task-surface owner metadata, phase maps, frontend phase maps, scheduler
  resource registry, and duration baseline inputs.
- Generated outputs must be refreshed only through Make-owned generators, never
  hand-edited.
- Generated outputs expected to change after owner-input edits may include
  scheduler manifests, task-surface Make fragments, browser batch manifests,
  topology render indexes, and phase schedule/ledger outputs.

## 11. Risks, Blockers, And Migration Notes

- Resolved: final retained warm run
  `.cartulary/test-results/20260708T224533Z-p1093` supplies accepted
  scheduler-health evidence with readiness attribution and
  `check-service-backed=119075ms`.
- Residual risk: owner contradiction can recur if NLSpec, task-surface owner
  metadata, or phase maps later disagree about default inclusion, fixture
  policy, or browser/session isolation. Future changes must route through the
  same spec-first and owner-input-first process.
- Public contracts must be migrated spec-first; private helper paths and
  temporary redirects should be removed rather than preserved unless promoted by
  NLSpec.
- Heavy support-only rows should not remain in default local check solely for
  historical reasons.
- Fixture policy reductions require row-specific proof; prior broad
  fixture-lowering attempts were rejected for lack of proof.
- Browser shard count increases should not be used as a default remedy without
  retained evidence and lane-balance proof.
- No proposed work may weaken summaries, failure classification, cleanup,
  service readiness, security scans, drift checks, or generated-artifact drift
  detection.

## 12. Binary Acceptance Criteria

- Public contracts are mapped: Make targets, command IDs, schema IDs, retained
  paths, cleanup, public inputs, and failure contracts.
- Timing baseline or explicit measurement plan is present.
- Each gap has remediation, area, rationale, benefit, migration impact,
  unresolved risk, and validation criteria.
- No generated file hand-edit is planned.
- No phase-shaped runtime dependency is introduced.
- Wall-time improvements preserve required summaries, failure classification,
  cleanup, service readiness, security checks, drift checks, and fixture
  lifecycle evidence.
- Warm steady-state timing is treated only as harness health evidence, not
  product performance evidence.
- Default local `check` work is retained only where it has clear continuing
  correctness value or an owner-spec reason.
