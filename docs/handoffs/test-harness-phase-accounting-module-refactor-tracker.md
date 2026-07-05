# test-harness-phase-accounting Next Refactor Tracker

## 1. Current Status and Authority

- Target label: `test-harness-phase-accounting`
- Target path: `tools/harness/phase-accounting`
- Tracker path:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Status: next structural refactor tracker refresh in progress on
  2026-07-05. Prior remediation work completed earlier on 2026-07-05 is
  historical baseline, not open work.
- Tracker purpose: identify the next forward-looking structural refactor
  iteration for phase-accounting after the completed G-01 through G-09 and
  WS-00 through WS-09 remediation pass.
- Public posture: preserve public harness contracts unless a future workstream
  is explicitly marked spec-first.
- Generated posture: generated files are downstream evidence and must not be
  hand-edited.
- Domain posture: `docs/domain.md` was inspected for vocabulary and concept
  boundaries only. Domain vocabulary is unchanged by this tracker refresh.

Authority and source hierarchy:

1. `docs/testing-harness-nlspec.md` owns harness mechanics, including command
   invocation, target selection, scheduling, fixture lifecycle, artifact
   emission, cleanup, retained roots, summary emission, failure mapping, helper
   ownership, and verification gates.
2. `docs/domain.md` owns vocabulary and concept boundaries only. Harness helper
   names, module names, validation rows, and tracker labels remain
   implementation-support terms unless an owner spec promotes them.
3. Product behavior is out of scope unless Core 00 through Core 04 or an
   adopted owner spec explicitly requires it. Core 05 applies only to
   claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
4. `docs/design.md` owns frontend design direction and token definitions, but
   it does not by itself establish Base Profile or extension-profile
   conformance.
5. Generated task-surface, topology, schedule, and Make artifacts are
   downstream of owner inputs.

## 2. Inspected Sources

This tracker refresh is based on inspection of:

- Controlling tracker:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Harness and vocabulary owners:
  `docs/testing-harness-nlspec.md`, `docs/domain.md`
- Phase-accounting implementation:
  `tools/harness/phase-accounting/**`
- Phase-accounting tests:
  `tools/harness/phase-accounting/tests/**`
- Adjacent harness boundaries:
  `tools/harness/scheduler/**`, `tools/harness/backend/**`,
  `tools/harness/browser/**`, `tools/harness/diagnostics/**`,
  `tools/harness/generated-artifacts/**`, `tools/harness/output/**`,
  `tools/harness/static-analysis/**`, and `tools/harness/tests/**`
- Frontend owner data:
  `tools/frontend_phase_registry.json`,
  `tools/frontend_phase_maps/*.json`,
  `tools/frontend_visual_fixture_registry.json`
- Relevant schemas:
  `tools/schemas/cartulary.phase_slice_plan.v1.schema.json`,
  `tools/schemas/cartulary.frontend_phase_registry.v3.schema.json`,
  `tools/schemas/cartulary.frontend_phase_test_map.v3.schema.json`,
  `tools/schemas/cartulary.frontend_row_accounting.v3.schema.json`,
  `tools/schemas/cartulary.frontend_visual_fixture_registry.v3.schema.json`,
  `tools/schemas/cartulary.frontend_evidence_audit_summary.v1.schema.json`,
  `tools/schemas/cartulary.scheduler_manifest.v1.schema.json`
- Generated-artifact and topology metadata:
  `tools/generated_artifact_policy.json`,
  `tools/task_surface_manifest.json`,
  `tools/task_surface.generated.mk`,
  `tools/execution_topology_manifest.json`,
  `tools/execution_topology_render_index.json`,
  `tools/scheduler_manifest.json`,
  `tools/browser_e2e_batch_manifest.json`
- Harness scratch support:
  `tools/harness/test-support/harness-scratch.sh`
- Current validation/tooling facts:
  `git status --short`, current registry and fixture summaries, and direct
  harness import-boundary collector output.

## 3. Completed Historical Baseline

The prior tracker and implementation pass completed these items. Treat them as
historical baseline unless current code, an owner-spec change, or validation
evidence proves regression:

- G-01/WS-03 frontend manifest decomposition moved freshness digest
  calculation, guide restatement validation, ledger rendering, scenario grep,
  and frontend CLI dispatch out of `frontend/phase-manifest-core.mjs`.
- G-02/WS-02 future-growth guardrails removed live-code fixed ceilings for
  `FE-P0..FE-P11` and `FE-VFIX-01..FE-VFIX-21`.
- G-03/WS-04 shared frontend row scope extracted row-ID parsing, selected-row
  scope construction, through-phase selection, and active/planned filtering to
  `frontend/row-scope.mjs`.
- G-04/WS-04 evidence audit library split moved reusable audit behavior into
  `frontend/evidence-audit.mjs`; the CLI remains public input and summary
  plumbing.
- G-05/WS-05 phase-slice planner decomposition moved backend and browser
  work-unit construction into `phase-slice-planning/**` modules behind the
  stable planner facade.
- G-06/WS-06 phase-manifest CLI dispatch cleanup replaced branch-heavy command
  dispatch with a table-based command handler while preserving public behavior.
- G-07/WS-06 import-boundary enforcement moved cross-owner imports to declared
  phase-accounting facades and added private-import rejection fixtures.
- G-08/WS-07 phase-slice work-unit schema decision recorded
  `cartulary.phase_slice_plan.v1` work-unit openness as private extension
  space, not a public work-unit-internals contract.
- G-09/WS-05 frontend/base phase-number coupling audit centralized the current
  `FE-P<N>` to `phase<N>` compatibility bridge in `frontend/phase-ids.mjs` and
  documented it in `docs/testing-harness-nlspec.md`.
- WS-08 refreshed generated metadata through Make-owned generation when the
  previous implementation moved generator inputs.
- WS-09 final validation passed targeted harness checks. Fresh retained browser
  roots remained a standing requirement before making live retained-root audit
  closure claims.

Do not reopen these as current work without evidence of regression or an owner
spec change.

## 4. Current Findings

- `tools/harness/phase-accounting/frontend/phase-manifest-core.mjs` remains a
  1,458-line mixed validator. It still combines schema constants, registry
  loading, phase-map validation, owner-ref validation, target-ref validation,
  row metadata validation, visual fixture registry validation, artifact checks,
  and freshness orchestration.
- `tools/harness/phase-accounting/phase-manifest-validation.mjs` remains a
  large base-manifest validator, and `phase-fixture-policy.mjs` remains a large
  fixture-policy module. Both are structurally harder to extend than the newer
  split planner modules.
- The frontend registry currently declares `FE-P0` through `FE-P11`, all with
  `status=active` and `row_rollup_state=active_green`.
- The frontend visual fixture registry currently declares 21 fixtures.
  `FE-VFIX-01` through `FE-VFIX-15` are `current`; `FE-VFIX-16` through
  `FE-VFIX-21` are `missing`.
- The harness import-boundary collector currently reports zero violations and
  recognizes the phase-accounting facades:
  `frontend/index.mjs`, `frontend-phase-manifest.mjs`,
  `frontend-readiness.mjs`, `frontend-row-accounting.mjs`,
  `phase-manifest.mjs`, `phase-registry.mjs`, and `phase-slice-plan.mjs`.
- Some frontend validation helpers still depend on ambient `process.cwd()` or
  root-insensitive caches while accepting explicit roots elsewhere. This makes
  fixture-root tests and future phase-growth checks harder to reason about.
- Frontend row accounting contains broad fallback behavior that can return no
  rows when registry loading fails. Future work should distinguish absent
  optional fixtures from malformed or stale owner data.
- `frontend-evidence-audit` owner text, task-surface metadata, and CLI behavior
  need a focused retained-root requiredness review. The owner text says
  preflight and measurement roots are required when selected-phase rows require
  those targets; current CLI requiredness is broader for check, support, visual,
  and a11y, and optional for preflight and measurement.
- Phase-accounting smoke fixtures still create repo-local `tmp/` scratch trees.
  This is not currently failing the fast-smoke scratch enforcement, which is
  scoped to the fast tier, but it is hardening debt relative to the NLSpec
  scratch direction for disposable repo-shaped fixtures.
- Generated-output references in the old tracker used stale `.generated` names
  for scheduler and browser batch manifests. Current generated topology outputs
  are `tools/scheduler_manifest.json`,
  `tools/browser_e2e_batch_manifest.json`,
  `tools/execution_topology_render_index.json`, and
  `tools/task_surface.generated.mk`.

## 5. Compatibility Freeze

Future implementation sessions must preserve these public harness contracts
unless a workstream is explicitly spec-first:

- Public Make target names, especially `phase-slice`,
  `service-backed-slice`, `frontend-evidence-audit`, `phase-ledgers`,
  `phase-ledger-drift`, `phase-schedules`, `phase-schedule-drift`,
  `phase-map-check`, `phase-test-name-check`, `harness-contract`, public smoke
  targets, and report/diagnostic targets.
- Stable `command_id` values, including
  `cartulary.harness.command.phase_slice.v1`,
  `cartulary.harness.command.service_backed_slice.v1`, and
  `cartulary.harness.command.frontend_evidence_audit.v1`.
- Schema IDs, including `cartulary.phase_slice_plan.v1`,
  `cartulary.frontend_phase_registry.v3`,
  `cartulary.frontend_phase_test_map.v3`,
  `cartulary.frontend_row_accounting.v3`,
  `cartulary.frontend_visual_fixture_registry.v3`,
  `cartulary.frontend_evidence_audit_summary.v1`,
  `cartulary.scheduler_manifest.v1`, and
  `cartulary.browser_e2e_batch_manifest.v5`.
- Retained artifact paths, including `frontend-row-accounting.json`,
  `frontend-evidence-audit-summary.json`, scheduler events and summaries,
  tool-run summaries, target summaries, generated ledgers, phase-slice plan
  output, scheduler manifests, and browser batch manifests.
- Public inputs, including `PHASE`, `PHASE_NAMESPACE`, `ROWS`, `JSON`,
  `CHECK_RESULTS_DIR`, `BROWSER_SUPPORT_RESULTS_DIR`,
  `BROWSER_VISUAL_RESULTS_DIR`, `BROWSER_A11Y_RESULTS_DIR`,
  `BROWSER_A11Y_PREFLIGHT_RESULTS_DIR`, and
  `BROWSER_MEASUREMENT_RESULTS_DIR`.
- Failure class/reason mapping, cleanup behavior, retained artifact retention,
  service ownership, runtime reset behavior, scheduler resource semantics,
  work-unit ordering semantics, and Make-owned wrapper behavior.

Any change to `cartulary.phase_slice_plan.v1`, frontend registry/map schemas,
`frontend-evidence-audit` env requiredness, retained artifact paths, Make target
behavior, task-surface metadata, or generated topology outputs must be
spec-first and recorded before implementation.

## 6. Out-of-Scope Boundaries

- Product HTTP routes, WebSocket behavior, workbook behavior, saved-view
  behavior, storage behavior, revision/history semantics, authentication,
  authorization, release publication behavior, and domain model changes.
- Domain vocabulary changes. This refresh records `domain vocabulary unchanged`.
- SQL migrations, DB queries, generated product contracts, dependency locks,
  `go.sum`, `pnpm-lock.yaml`, and tool-managed dependency/install artifacts.
- Hand-editing generated roots or generated outputs, including
  `internal/gen/**`, `packages/protocol-ts/src/generated/**`,
  `packages/ui-contracts/src/generated/**`,
  `tools/task_surface.generated.mk`, `tools/scheduler_manifest.json`,
  `tools/browser_e2e_batch_manifest.json`, and
  `tools/execution_topology_render_index.json`.
- Broad refactors outside phase-accounting/harness ownership unless required
  by owner metadata or import-boundary enforcement.
- Treating visual, accessibility, design-direction, or implementation-support
  evidence as product conformance without a Core 05 or product-owner boundary.
- Reopening completed G-01 through G-09 work without concrete current evidence.

## 7. Detailed Gap Records

### G-10 Frontend Validator Decomposition

- Remediation: split real owner-local modules out of
  `frontend/phase-manifest-core.mjs` for schema constants, registry loading,
  phase-map validation, owner refs, target refs, row metadata, visual fixture
  registry validation, and artifact/freshness orchestration. Keep
  `frontend-phase-manifest.mjs`, `frontend/index.mjs`, and current public
  facade exports stable.
- Affected areas: implementation, tests, import-boundary metadata if new
  facades are introduced, generated metadata only if backing script paths move.
- Rationale: the current frontend core still concentrates unrelated validation
  rules in one 1,458-line module after the prior decomposition pass.
- Expected long-term benefit: smaller validation modules, narrower tests,
  lower review risk, and easier frontend phase/fixture growth.
- Compatibility or migration impact: no public Make target, schema ID, retained
  path, CLI behavior, or facade export may change. Private imports may move
  within the phase-accounting owner boundary.
- Risks of leaving unresolved: future frontend owner-data changes will continue
  to collect in a catch-all validator, raising the chance of accidental public
  behavior drift.
- Validation criteria: `node --check` for changed modules,
  `make phase-map-check`, `make json-shape-check`, `make harness-contract`,
  `make lint-scripts`, and import-boundary checks when facade lists change.

### G-11 Root-Scoped Loader And Cache Context

- Remediation: replace ambient `process.cwd()` use and root-insensitive global
  caches in frontend validation and phase selection with an explicit
  root-scoped context for registry/maps, task-surface entries, base Playwright
  title ownership, Playwright source files, and freshness inputs.
- Affected areas: implementation, tests, fixture-root validation, frontend
  phase selection, browser/frontend row-evidence joins.
- Rationale: helpers that accept `root` should not silently read from the
  process working tree or reuse cache entries from another root.
- Expected long-term benefit: synthetic fixture tests become trustworthy,
  future phase-growth tests can use isolated roots, and validation behavior is
  easier to reason about under parallel or repeated invocations.
- Compatibility or migration impact: no public output change is intended.
  Private helper signatures may gain context parameters inside the owner
  boundary.
- Risks of leaving unresolved: fixture-root tests can pass for the wrong reason
  or fail non-deterministically after a previous validation populated a global
  cache.
- Validation criteria: targeted synthetic-root tests for frontend phase-map
  validation and Playwright title ownership, `node --check` for changed modules,
  `make phase-map-check`, `make harness-contract`, and representative frontend
  namespace `phase-slice` JSON checks.

### G-12 Fail-Closed Frontend Accounting Integration

- Remediation: replace broad "catch and return no rows" behavior in frontend
  row accounting and Playwright frontend selection with explicit handling for
  absent optional fixture data versus malformed, stale, or invalid owner data.
  Malformed current-profile registry/map data should fail closed rather than
  silently disabling frontend row closure.
- Affected areas: implementation, tests, frontend row-accounting retained
  output, target-summary failure injection, Playwright selection.
- Rationale: current fallback behavior can hide frontend owner-data breakage by
  producing no accounting rows.
- Expected long-term benefit: clearer diagnostics, fewer false-green target
  summaries, and stronger confidence that frontend readiness evidence is
  actually evaluated.
- Compatibility or migration impact: malformed current-profile frontend data may
  fail targets that previously skipped accounting. This is a behavior-tightening
  change and should be validated against NLSpec wording before implementation.
- Risks of leaving unresolved: broken frontend registry or map data could make
  required row accounting disappear instead of failing the owning target.
- Validation criteria: focused fixtures for missing registry, malformed
  registry, stale digest, and valid no-row target cases; `make phase-map-check`;
  `make harness-contract`; `make run-harness-smoke-extended`; targeted
  frontend-unit/browser row-accounting smokes.

### G-13 Frontend Evidence Audit Retained-Root Policy

- Remediation: reconcile NLSpec retained-root requiredness, task-surface input
  metadata, and `frontend-evidence-audit-cli.mjs` behavior for
  `CHECK_RESULTS_DIR`, support, visual, a11y, preflight, and measurement roots.
  Prefer selected-phase required-target driven validation unless the owner spec
  intentionally keeps broader required inputs.
- Affected areas: specification if requiredness changes, implementation, tests,
  task-surface metadata, execution topology metadata, generated outputs if
  owner inputs change.
- Rationale: retained-root requiredness is public command behavior. The current
  owner text requires preflight and measurement only when selected-phase rows
  require those targets, while existing CLI behavior has a different required
  set.
- Expected long-term benefit: callers get precise missing-root diagnostics and
  future frontend targets can be added without broad always-required retained
  roots.
- Compatibility or migration impact: any env requiredness change is public and
  must be spec-first. If no change is accepted, the tracker should record that
  the broader current CLI requiredness is intentional.
- Risks of leaving unresolved: future evidence-audit callers can be blocked by
  unnecessary roots or, worse, omit roots that are required for selected-phase
  closure.
- Validation criteria: accepted NLSpec/task-surface decision, retained-root
  fixture tests for required and optional target roots, `make
  frontend-evidence-audit` with Make-owned roots when available,
  `make harness-contract`, `make json-shape-check`, and generated drift checks
  if metadata changes.

### G-14 Base Phase Manifest Validator Split

- Remediation: decompose `phase-manifest-validation.mjs` into owner-local
  modules for expected-ID and guide parity, runner-specific validation,
  source scanning, support-Go validation, profile-claim validation, and
  orchestration behind stable `validateManifest` exports.
- Affected areas: implementation, tests, import-boundary metadata if new
  facades are introduced, generated metadata only if public backing paths move.
- Rationale: base manifest validation mixes ID contracts, guide extraction,
  source traversal, runner semantics, fixture policy, support entries, and
  profile claims in one large module.
- Expected long-term benefit: easier addition of future base phases, clearer
  runner-specific validation, and safer review of fixture-policy changes.
- Compatibility or migration impact: preserve `phase-map-check`,
  `phase-test-name-check`, `phase-manifest.mjs` facade exports, validation
  diagnostics unless tests intentionally update them, and schema IDs.
- Risks of leaving unresolved: future runner or fixture additions will keep
  expanding a hard-to-test validator and can accidentally shift unrelated
  validation behavior.
- Validation criteria: `node --check` for changed modules,
  `make phase-map-check`, `make phase-test-name-check`,
  `bash tools/harness/phase-accounting/tests/test-check-phase-test-names.sh`,
  `make harness-contract`, and `make lint-scripts`.

### G-15 Fixture Policy Compatibility Contraction

- Remediation: audit broad `package_reset` and other Postgres fixture
  compatibility paths in `phase-fixture-policy.mjs` and base phase maps.
  Identify rows that can move to targeted reset or `group_clone` policies.
  Keep current behavior until a spec-first decision approves any public policy
  tightening. Current implementation audit result: no `package_reset` rows are
  declared in `tools/phase*_test_map.json`; no phase-map contraction was
  performed.
- Affected areas: specification, phase maps, implementation, tests, fixture
  budget metadata, generated ledgers/schedules if owner data changes.
- Rationale: the NLSpec treats broad package reset as a compatibility path and
  prefers narrower fixture policies when the touched surface is small and
  stable.
- Expected long-term benefit: lower service-backed fixture cost, clearer
  Postgres isolation intent, and less compatibility burden around broad mutable
  reset behavior.
- Compatibility or migration impact: fixture-policy changes can affect
  service-backed execution and timing evidence. Treat policy contraction as
  spec-first when public row behavior or fixture budgets change.
- Risks of leaving unresolved: broad reset behavior remains a long-term default
  and future phase growth inherits unnecessary fixture cost and ambiguity.
- Validation criteria: documented audit inventory, accepted owner decision for
  any contraction, `make phase-map-check`, `make service-backed-slice
  PHASE=phase4 JSON=1`, service-backed slice checks for affected phases,
  `make harness-contract`, and timing evidence only when execution behavior
  changes.

### G-16 Phase-Accounting Smoke Scratch Hardening

- Remediation: migrate phase-accounting smoke fixtures that create repo-shaped
  temp trees under repo-local `tmp/` to
  `tools/harness/test-support/harness-scratch.sh`, scoped to
  `tools/harness/phase-accounting/tests/**`. Do not start a broad adjacent
  harness scratch migration from this tracker.
- Affected areas: tests and documentation/handoff only unless target metadata
  changes.
- Rationale: the NLSpec directs disposable repo-shaped smoke fixtures to use
  harness scratch outside the repository checkout. Phase-accounting tests are
  not currently part of the fast scratch enforcement set, but they are still
  better isolated outside the checkout.
- Expected long-term benefit: less risk of concurrent source traversal seeing
  transient fake packages, clearer cleanup behavior, and consistency with the
  scratch helper pattern already used by fast smoke tests.
- Compatibility or migration impact: no public command behavior change.
  `CARTULARY_HARNESS_SCRATCH_ROOT` must remain optional and must still resolve
  outside the repository.
- Risks of leaving unresolved: future tier changes could turn current
  repo-local scratch behavior into a conformance blocker; transient fixture
  trees can also interfere with concurrent source scans.
- Validation criteria: direct phase-accounting smoke scripts,
  `make run-harness-smoke-extended`, `make harness-contract`,
  `make lint-shell`, and no broad adjacent-test edits unless a failure proves
  they are required.

## 8. Workstream Sequencing

### WS-00 Tracker Refresh

- Depends on: none.
- Sequencing: complete before any code movement.
- Changes: replace stale open G-01 through G-09 sections with this
  forward-looking G-10 through G-16 tracker and correct generated-output path
  names.
- Risks: documentation drift only.
- Exit criteria: required sections, gap records, generated-artifact rules, and
  session log are current.
- Required validation: `make lint-markdown`.
- Generated-artifact handling: none.
- Tracker update requirements: record validation results and skipped checks.

### WS-01 Characterization Baseline

- Depends on: WS-00.
- Sequencing: run before implementation code movement.
- Changes: no code changes; capture current behavior.
- Risks: unrelated existing harness failure may obscure refactor risk.
- Exit criteria: baseline command results and run roots are recorded before
  code movement.
- Required validation: `make phase-map-check`; `make phase-test-name-check`;
  `make harness-contract`; `make phase-slice PHASE=phase4 JSON=1`;
  `make service-backed-slice PHASE=phase4 JSON=1`;
  `make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`;
  `make service-backed-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`.
- Generated-artifact handling: none unless a command proves drift.
- Tracker update requirements: append results and any failures before starting
  WS-02.

### WS-02 Frontend Context And Decomposition

- Depends on: WS-01.
- Sequencing: address G-10 and G-11 first so later frontend accounting changes
  build on root-scoped helpers.
- Changes: split frontend validation modules and introduce root-scoped context
  for caches/loaders.
- Risks: export drift, cache behavior drift, fixture-root behavior changes, or
  accidental generated metadata path changes.
- Exit criteria: `frontend/phase-manifest-core.mjs` no longer owns unrelated
  frontend validation concerns, and root-scoped tests exercise isolated roots.
- Required validation: `node --check` for changed modules;
  `make phase-map-check`; `make json-shape-check`; `make harness-contract`;
  `make lint-scripts`; representative frontend namespace phase-slice JSON
  checks.
- Generated-artifact handling: run generated drift checks only if owner inputs
  or public backing paths change.
- Tracker update requirements: mark G-10/G-11 complete or deferred with
  validation evidence.

### WS-03 Frontend Accounting/Audit Semantics

- Depends on: WS-02.
- Sequencing: address G-12 before or alongside G-13 so retained-root audit
  behavior relies on fail-closed accounting semantics.
- Changes: tighten frontend accounting fallbacks and reconcile audit
  retained-root requiredness.
- Risks: public env requiredness drift, false failures for valid optional
  fixtures, or missed row closure.
- Exit criteria: missing optional fixtures, malformed owner data, stale
  digests, and missing retained roots have distinct diagnostics.
- Required validation: focused row-accounting and audit fixtures;
  `make frontend-evidence-audit` with Make-owned retained roots when available;
  `make harness-contract`; `make run-harness-smoke-extended`;
  `make json-shape-check`.
- Generated-artifact handling: update task-surface/topology owner inputs first
  if requiredness metadata changes, then regenerate through Make.
- Tracker update requirements: record whether G-13 made a spec-first public
  change or confirmed current behavior.

### WS-04 Base Validator And Fixture Policy

- Depends on: WS-01; WS-02 may run in parallel if files do not overlap.
- Sequencing: address G-14 decomposition before G-15 policy contraction.
- Changes: split base manifest validation modules and audit Postgres fixture
  compatibility paths.
- Risks: validation diagnostic drift, support-Go behavior drift, service-backed
  fixture budget changes, or timing changes.
- Exit criteria: base validator responsibilities are separated and any fixture
  policy contraction has a recorded owner decision.
- Required validation: `node --check` for changed modules;
  `make phase-map-check`; `make phase-test-name-check`;
  `bash tools/harness/phase-accounting/tests/test-check-phase-test-names.sh`;
  `make harness-contract`; service-backed JSON checks for affected phases.
- Generated-artifact handling: refresh generated ledgers/schedules only when
  owner phase maps or topology inputs change.
- Tracker update requirements: mark G-14 complete and G-15 complete/deferred
  with the policy decision.

### WS-05 Smoke Fixture Hardening

- Depends on: WS-01.
- Sequencing: can run after validator movement, or earlier if isolated.
- Changes: migrate phase-accounting smoke tests from repo-local `tmp/` scratch
  to `harness-scratch.sh`.
- Risks: test-only path assumptions or cleanup behavior drift.
- Exit criteria: phase-accounting smoke tests use external scratch for
  disposable repo-shaped fixtures and still clean up.
- Required validation: direct changed smoke scripts;
  `make run-harness-smoke-extended`; `make harness-contract`;
  `make lint-shell`.
- Generated-artifact handling: none unless task metadata changes.
- Tracker update requirements: record scripts moved and confirm no broad
  adjacent-test scratch migration was performed.

### WS-06 Generated Metadata Review

- Depends on: any workstream with owner input or backing-script movement.
- Sequencing: run only when needed.
- Changes: update owner inputs first, then regenerate downstream outputs through
  Make.
- Risks: hand-editing generated outputs or leaving stale backing paths.
- Exit criteria: generated outputs match owner inputs and drift checks pass.
- Required validation: `make phase-schedules`;
  `make phase-schedule-drift`; `make generate-drift`;
  `make generated-artifact-policy-check`; `make json-shape-check`;
  `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`.
- Generated-artifact handling: never hand-edit generated outputs.
- Tracker update requirements: list owner inputs and generated outputs touched.

### WS-07 Final Handoff

- Depends on: completed implementation workstreams.
- Sequencing: run after validation for the implementation slice.
- Changes: update tracker or successor handoff with final status.
- Risks: incomplete handoff context for the next agent.
- Exit criteria: binary completion criteria are checked and remaining risks are
  explicit.
- Required validation: `make agent-finalize`; broader `make check` only when
  implementation breadth warrants it.
- Generated-artifact handling: confirm generated outputs are unchanged or
  refreshed through Make.
- Tracker update requirements: record files changed, validations, skipped
  checks, run roots, failures, and next action.

## 9. Generated-Artifact Handling Rules

- Do not hand-edit generated files or generated roots.
- Generated roots from policy include `internal/gen/**`,
  `packages/protocol-ts/src/generated/**`, and
  `packages/ui-contracts/src/generated/**`.
- Current generated topology/task-surface outputs include
  `tools/task_surface.generated.mk`, `tools/scheduler_manifest.json`,
  `tools/browser_e2e_batch_manifest.json`, and
  `tools/execution_topology_render_index.json`.
- Generated outputs are downstream of owner inputs such as
  `tools/execution_topology_manifest.json`,
  `tools/task_surface_manifest.json`, frontend phase registries/maps, and
  generator scripts.
- If helper backing paths, owner metadata, phase maps, or topology inputs
  change, update owner inputs first and run Make-owned generation. Do not edit
  generated outputs as the source of truth.
- Use `make phase-schedules` for schedule/topology/task-surface refreshes when
  relevant.
- Verify with `make phase-schedule-drift`, `make generate-drift`,
  `make generated-artifact-policy-check`, and `make json-shape-check`.
- If generated metadata still references an old helper path after a move, the
  workstream is incomplete.

## 10. Validation Matrix

- Tracker-only docs update: `make lint-markdown`.
- End-of-run tracker handoff: `make agent-finalize`; rerun
  `make lint-markdown` after recording final validation results.
- Harness docs, schema, or metadata change: `make json-shape-check`;
  `make generated-artifact-policy-check`; `make harness-contract`.
- Phase map or registry change: `make phase-map-check`;
  `make phase-test-name-check`; `make phase-ledger-drift`;
  `make phase-schedule-drift`.
- Frontend validator/context refactor: `node --check` for changed modules;
  `make phase-map-check`; `make json-shape-check`; `make harness-contract`;
  representative frontend namespace `phase-slice` JSON checks.
- Base manifest validator refactor: `node --check` for changed modules;
  `make phase-map-check`; `make phase-test-name-check`;
  `bash tools/harness/phase-accounting/tests/test-check-phase-test-names.sh`;
  `make harness-contract`.
- Frontend row-accounting or evidence-audit behavior: focused audit and
  row-accounting fixtures; `make frontend-evidence-audit` with Make-owned roots
  when available; `make run-harness-smoke-extended`; `make harness-contract`.
- Phase-slice planner behavior: base and frontend `phase-slice` and
  `service-backed-slice` JSON checks; `make harness-contract`.
- Smoke fixture scratch hardening: direct changed scripts;
  `make run-harness-smoke-extended`; `make harness-contract`;
  `make lint-shell`.
- Owner input or backing-script path movement: `make phase-schedules`;
  `make phase-schedule-drift`; `make generate-drift`;
  `make generated-artifact-policy-check`; `make json-shape-check`;
  `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`.
- Broad `make check`: run only when implementation breadth or ownership impact
  warrants it; otherwise record why targeted validation is sufficient.

## 11. Handoff Update Requirements

Every future implementation session using this tracker must update this file or
a successor handoff with:

- Workstream status and exact started/completed timestamps.
- Files changed, grouped by owner area.
- Public contracts preserved, or spec-first public changes made.
- Generated-artifact handling: owner inputs changed, generation commands run,
  generated outputs refreshed, and drift results.
- Verification commands, pass/fail status, run roots when emitted, and failure
  classification if any command fails.
- Skipped checks with concrete reasons.
- Retained-root requirements or fixture limitations that prevent stronger audit
  claims.
- Remaining blockers, risks, or owner-spec decisions.

Required tracker updates by workstream:

- WS-00 records the markdown validation result.
- WS-01 records baseline command results before code movement.
- WS-02 through WS-05 mark touched gaps open, in progress, completed, or
  deferred with validation evidence.
- WS-06 lists owner inputs and generated outputs touched.
- WS-07 updates open risks, binary criteria, and next recommended action.

## 12. Open Risks and Blockers

- OR-011: frontend validator core remains a high-change mixed module.
  Resolution path: complete G-10 through WS-02.
  Status: completed on 2026-07-05; `frontend/phase-manifest-core.mjs` is now a
  compatibility re-export layer backed by owner-local modules.
- OR-012: root-scoped validation is not consistently enforced.
  Resolution path: complete G-11 with root-context tests.
  Status: completed on 2026-07-05; root-scoped cache/context tests were added
  for Go module and Playwright source discovery.
- OR-013: frontend accounting can silently skip rows when owner data loading
  fails.
  Resolution path: complete G-12 with fail-closed fixture coverage.
  Status: completed on 2026-07-05; frontend-aware row accounting now fails
  closed on owner-data load errors.
- OR-014: frontend evidence-audit retained-root requiredness needs an explicit
  current-profile decision.
  Resolution path: complete G-13 spec-first or record that current behavior is
  intentional.
  Status: completed on 2026-07-05; selected-phase target-driven retained-root
  requiredness was specified and implemented.
- OR-015: base manifest validation remains concentrated in a large mixed
  module.
  Resolution path: complete G-14 through WS-04.
  Status: completed on 2026-07-05; base validation now delegates runner,
  ID/guide, source-scan, and profile-claim checks to focused modules.
- OR-016: broad fixture-policy compatibility paths may become default future
  practice.
  Resolution path: complete G-15 audit and owner decision before contraction.
  Status: completed on 2026-07-05; current phase maps declare no
  `package_reset` rows and no unapproved contraction was made.
- OR-017: phase-accounting smoke scratch still uses repo-local `tmp/`.
  Resolution path: complete G-16 through WS-05.
  Status: completed on 2026-07-05; phase-accounting smoke fixtures now use
  `harness-scratch.sh`.
- OR-008: direct retained-root audit still depends on fresh Make-owned browser
  roots.
  Resolution path: produce or supply retained roots before claiming live
  browser audit closure.
  Status: standing validation requirement.

## 13. Binary Completion Criteria

This tracker refresh is complete only when all of the following are true:

- Prior G-01 through G-09 remediation is summarized as completed history and no
  longer presented as current work.
- Current public compatibility surfaces are explicitly frozen.
- Product, dependency, domain-vocabulary, and generated-file boundaries are
  explicit.
- Current findings record the 1,458-line frontend core, large base validation
  modules, current frontend phase/fixture registry facts, zero import-boundary
  violations, and corrected generated-output path names.
- New gap records G-10 through G-16 each include remediation, affected areas,
  rationale, expected long-term benefit, compatibility or migration impact,
  risks of leaving unresolved, and validation criteria.
- Workstreams include dependencies, sequencing, risks, exit criteria, required
  validation, generated-artifact handling, and tracker update requirements.
- Open risks and blockers are listed with resolution paths.
- Validation matrix covers tracker-only, schema/metadata, phase map, frontend
  validator/context, base validator, audit/accounting, planner, smoke scratch,
  generated-artifact, and finalization changes.
- The session handoff log records `make lint-markdown` and finalization results
  for this tracker refresh.

Future implementation slices are complete only when their gap records are
marked completed or intentionally deferred, their public compatibility impact
is recorded, generated-artifact rules were followed, and their validation
commands passed or failures are classified with run roots.

## 14. Session Handoff Log

- 2026-07-05T10:37:26-04:00 through 2026-07-05T11:30:11-04:00,
  Codex GPT-5 remediation baseline: prior tracker refresh and implementation
  pass completed G-01 through G-09 and WS-00 through WS-09. Targeted harness
  validation passed; `make agent-finalize` passed at
  `.cartulary/test-results/20260705T152755Z-p81448`; retained-run maintenance
  was skipped because `RESULTS_DIR` was unset. Treat this as historical
  baseline.
- 2026-07-05T11:53:29-04:00, Codex GPT-5 next-iteration tracker refresh
  started. Files changed:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`.
  Commands run before edit: `git status --short`, `date -Iseconds`, tracker
  inspection, owner/spec/code/schema/generated-metadata inspection, current
  frontend registry summary, current visual fixture registry summary, generated
  output inventory, and direct harness import-boundary collector. Result:
  tracker replacement started from an already modified tracker file and treats
  that prior modified content as superseded by this refresh. Next action: run
  `make lint-markdown`, `make agent-finalize`, record results, and rerun
  markdown validation.
- 2026-07-05T11:53:29-04:00 through 2026-07-05T11:56:49-04:00,
  Codex GPT-5 WS-00 tracker refresh completed. Files changed:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`.
  Commands run: `make lint-markdown`; `make agent-finalize`. Result:
  `make lint-markdown` passed with no output; `make agent-finalize` passed at
  `.cartulary/test-results/20260705T155634Z-p52242` with
  `generated=unchanged files=0`, `run_checks=skipped`, and `results_dir=-`.
  Retained-run maintenance was skipped because `RESULTS_DIR` was unset. Broad
  `make check` was not run because this was a tracker-only documentation
  update and targeted tracker validation passed. A post-log `make
  lint-markdown` rerun passed with no output. Next action: future
  implementation should start with WS-01 characterization baseline before code
  movement.
- 2026-07-05T12:21:37-04:00 through 2026-07-05T13:09:24-04:00,
  Codex GPT-5 G-10 through G-16 remediation implementation. Substantive
  changes: frontend phase validation was decomposed behind stable facades;
  root-scoped cache helpers were added; frontend row accounting now fails
  closed for frontend-aware owner-data load failures; `frontend-evidence-audit`
  retained-root requiredness is selected-phase target-driven; base manifest
  validation was split into focused helper modules; phase-accounting smoke
  scratch moved to `harness-scratch.sh`; task-surface/topology owner metadata
  and generated outputs were refreshed through `make phase-schedules`.
  Fixture-policy audit result: no `package_reset` rows are declared in current
  phase maps, so no phase-map contraction was performed. Commands run:
  baseline `make phase-map-check`, `make phase-test-name-check`, `make
  harness-contract`, `make phase-slice PHASE=phase4 JSON=1`, `make
  service-backed-slice PHASE=phase4 JSON=1`, `make phase-slice
  PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`, and `make service-backed-slice
  PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`; implementation validation
  direct `node --check` and import checks for changed modules; direct
  phase-accounting smoke scripts; `make phase-schedules`; `make
  phase-map-check`; `make phase-test-name-check`; `make json-shape-check`;
  `make phase-schedule-drift`; `make generate-drift`; `make
  generated-artifact-policy-check`; `make harness-contract`; `make
  lint-scripts`; `make lint-shell`; `make lint-markdown`; `make
  run-harness-smoke-extended`; post-change base/frontend phase-slice checks;
  and `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`. Results:
  all listed commands passed. OR-008 remains a standing validation requirement:
  live retained browser-root audit closure was not claimed because fresh
  Make-owned `check` plus browser retained roots were not produced in this
  session.
