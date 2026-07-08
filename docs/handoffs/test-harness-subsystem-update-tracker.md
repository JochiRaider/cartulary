# Test Harness Subsystem Update Tracker

Target artifact: `docs/handoffs/test-harness-subsystem-update-tracker.md`

Planning target: Cartulary test harness subsystem refactor tracker focused on
reducing warm steady-state wall time for `make check`, especially the
`check-service-backed` critical path.

## 1. Session Context

- Branch: `main`
- Commit: `d1b261e02a77d8cb11d10c7191c55f5a08a35e4e`
- Dirty tree: clean by `git status --short --branch`; branch is ahead of
  `origin/main` by 16 commits.
- Date: 2026-07-08, `America/New_York`
- Primary retained evidence inspected:
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
- Source limits: no broad suite was run for discovery. Existing retained timing
  evidence is usable for bottleneck modeling, but warm-readiness eligibility is
  incomplete because readiness attribution is empty in the retained pressure
  summary.
- This plan is only for the test harness subsystem and wall-time reduction. It
  does not propose product behavior changes except through explicit spec-first
  gates where default evidence selection may change.

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

Retained warm-ish baseline:

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
- Readiness/provisioning attribution is incomplete: retained pressure summary
  has empty `readiness_attribution_counts`, so warm scheduler-health evidence
  needs a follow-up measurement with explicit readiness classification.

Minimal measurement plan if retained evidence is not accepted:

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
| THT-WT-001 | Establish accepted warm baseline and readiness attribution | Measurement | TODO: retained warm run required | None | Harness timing owner | Warm `make check` run root | Baseline separates warm scheduler health from provisioning |
| THT-WT-002 | Audit default `check` selection for support-only, duplicate, and high-cost rows | Selection | Proposed | THT-WT-001 | Harness/spec owners | Target plan JSON and phase maps | Every default row has continuing local value or is routed spec-first out of default |
| THT-WT-003 | Expose hidden readiness/provisioning as first-class pressure attribution | Diagnostics | Proposed | THT-WT-001 | Scheduler/readiness owner | `pressure-summary.json` | Readiness costs are visible and excluded from warm scheduler-health conclusions |
| THT-WT-004 | Reduce backend-process terminal tail, especially SeaweedFS support rows | Scheduling/selection | Proposed | THT-WT-002 | Backend harness owner | Scheduler summary and row evidence | Backend-process no longer dominates terminal wait without evidence loss |
| THT-WT-005 | Rebalance browser webserver-backed and stateful stages by owner metadata | Browser scheduler | Proposed | THT-WT-001 | Browser harness owner | Browser batch manifest and scheduler events | Browser lanes are balanced without adding accidental isolation |
| THT-WT-006 | Target fixture tier proof work to top pressure rows only | Fixture lifecycle | Proposed | THT-WT-001 | Fixture owner | Fixture proof artifacts | Clone/reset reductions are proof-backed and row-specific |
| THT-WT-007 | Tune topology resource claims and lane priorities from owner inputs | Scheduler | Proposed | THT-WT-003 | Execution topology owner | Generated scheduler manifest | Host I/O, clone, and process blockers decrease without oversubscription failures |
| THT-WT-008 | Remove duplicate support-only default evidence where owner docs allow | Selection/spec | Proposed | THT-WT-002 | Harness spec owner | NLSpec and target plan | Duplicate local evidence is moved to explicit or CI/release targets |
| THT-WT-009 | Clean private helper path coupling in touched scheduler/readiness code | Structure | Proposed | Related edits | Harness module owners | Import/path checks | Callers use owner facades; private compatibility aliases are not preserved |
| THT-WT-010 | Validate final timing and retained-run maintenance | Verification | Proposed | Implemented workstreams | Harness verification owner | Final run root and finalizer output | `make check` wall time is reduced and required summaries remain intact |

## 6. Gap Analysis And Recommendations

| Gap ID | Gap | Remediation | Area: spec / implementation / tests / docs / multiple | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| G-001 | Warm-readiness attribution is missing from retained pressure evidence | Add first-class readiness/provisioning attribution in scheduler pressure summaries | implementation / tests / docs | Warm health cannot accept hidden provisioning as scheduler cost | Clear timing evidence across machines | Preserve summary schemas or version spec-first | Future optimizations chase cold setup noise | Pressure summary identifies readiness/provisioning classes |
| G-002 | `check-service-backed` terminal wait is backend-process/stateful dominated | Rebalance or reclassify high-cost tail rows through owner metadata | multiple | Browser is not the only critical path | High wall-time reduction potential | Default row removals are spec-first | Run stays near 140s despite browser tuning | Critical path no longer waits on unjustified tail work |
| G-003 | Heavy support-only SeaweedFS rows appear in default local path | Audit continuing value; move out of default if not local correctness evidence | spec / implementation / tests | The optimization goal rejects legacy preservation without continuing value | High if removed or narrowed | Public/default behavior change requires NLSpec and manifest update | Local check pays CI-style support cost forever | Target plan shows support rows either justified or excluded |
| G-004 | Browser default projection may be over-broad or poorly grouped | Audit browser projection rows, shard weights, and stateful session grouping | multiple | More shards alone was previously rejected; owner-row clarity is needed | Medium | Generated browser batch refreshed only through Make | Browser stage remains noisy and hard to grow | Browser lanes meet balance target with stable failure reporting |
| G-005 | Fixture clone/reset pressure remains high but proof closure blocks changes | Limit proof work to top pressure rows and require row-specific isolation evidence | tests / implementation | Broad fixture lowering is unsafe without proof | Medium | Fixture policy changes touch phase maps and summaries | Database pressure remains dominant | Fixture counts/duration drop with accepted proofs |
| G-006 | Resource model reports `host_io`, clone, and process contention but lacks clear remediation loop | Tune resource claims, priorities, and limits through topology owner inputs after measurement | implementation / tests | Lane balance needs accurate resource semantics | Medium | Regenerate scheduler outputs; no public contract drift | Scheduler keeps serializing independent work | Pressure blockers decrease without flakes |
| G-007 | Private helper path history can become accidental compatibility | Move touched callers to owner facades; delete temporary redirects instead of preserving them | implementation / docs | NLSpec says private helpers are implementation details | Low wall time, high maintainability | No public migration unless promoted by NLSpec | Future refactors inherit false compatibility | Import-boundary and harness contract checks pass |
| G-008 | Duplicate default evidence across harness smoke, support rows, and full targets is not fully explained | Add default inclusion reason audit using existing owner metadata | docs / spec / tests | Default local check should stay correctness-focused | Medium | Spec-first for changed defaults | Default check expands with every phase | Every default row has `default_check_reason_code` and non-duplicate value |
| G-009 | Generated topology outputs are easy to hand-edit accidentally | Require owner-input edits plus generator/drift validation at each checkpoint | docs / tests | Generated outputs are downstream | Low direct timing, high safety | No hand-edited generated files | Schedule drift becomes unreproducible | Generated-artifact and schedule drift checks pass |

## 7. Prioritized Roadmap

P0:

- THT-WT-001 / G-001: accepted warm baseline and readiness attribution. Effect:
  unknown pending baseline.
- THT-WT-002 / G-003 / G-008: default-selection audit, especially support-only
  and duplicate high-cost rows. Effect: high.
- THT-WT-004 / G-002: backend-process terminal-tail reduction. Effect: high.
- THT-WT-007 / G-006: scheduler lane/resource tuning from owner inputs. Effect:
  medium.

P1:

- THT-WT-005 / G-004: browser projection and session grouping balance. Effect:
  medium.
- THT-WT-006 / G-005: top-row fixture proof reductions. Effect: medium.
- THT-WT-009 / G-007: owner facade cleanup in touched harness areas. Effect:
  low wall-time, high maintainability.

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

- TODO: retained warm run required before claiming accepted scheduler-health
  improvement, because current retained evidence lacks readiness attribution
  even though it records `138.03s`.
- BLOCKED: owner contradiction if NLSpec, task-surface owner metadata, or phase
  maps disagree about default inclusion, fixture policy, or browser/session
  isolation.
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
