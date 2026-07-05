# test-harness-phase-accounting Next Refactor Tracker

## 1. Current Status and Authority

- Target label: `test-harness-phase-accounting`
- Target path: `tools/harness/phase-accounting`
- Tracker path:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Status: next simplification tracker refresh started on
  2026-07-05T14:04:10-04:00.
- Tracker purpose: identify the next structural Phase-Accounting refactor
  iteration after completed G-01 through G-16 work. This is tracker-only work;
  implementation changes belong to a later session.
- Current iteration: G-17 and later are active planning records. G-01 through
  G-16 are completed historical baseline unless live code, owner docs, or
  validation proves a regression.
- Public posture: preserve public Make targets, stable `command_id` values,
  schema IDs, retained artifact paths, failure mapping, cleanup behavior, and
  Make-owned wrapper semantics unless a future workstream is explicitly
  marked spec-first.
- Private posture: private helper compatibility is not preserved by default.
  A private helper path, shim, re-export layer, or compatibility alias stays
  only when it has clear continuing value.
- Generated posture: generated files and generated outputs are downstream
  evidence. Update owner inputs first, then use Make-owned generation.
- Domain posture: `docs/domain.md` was consulted only for vocabulary and
  concept boundaries. Domain vocabulary is unchanged.

Authority and source hierarchy:

1. `docs/testing-harness-nlspec.md` owns harness mechanics, including command
   invocation, target selection, scheduling, fixture lifecycle, artifact
   emission, cleanup, retained roots, summary emission, failure mapping, helper
   ownership, smoke tiers, cache behavior, and verification gates.
2. Core 00 through Core 04 own product behavior. Core 05 applies only to
   claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
3. `docs/domain.md` owns vocabulary and concept-boundary interpretation only.
4. `docs/design.md` owns frontend design direction and token definitions, but
   it is not product-conformance evidence by itself.
5. Task-surface, topology, schedule, ledger, and generated Make outputs are
   downstream of owner inputs.

## 2. Sources Inspected

This refresh is grounded in these current sources and commands:

- Repository procedure and tracker history:
  `AGENTS.md`,
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`,
  `git status --short`, and `date -Iseconds`.
- Harness and vocabulary owners:
  `docs/testing-harness-nlspec.md`, especially harness import-boundary,
  unsupported-private helper, frontend accounting, phase-slice, fixture
  policy, smoke tier, cache, and retired-behavior sections; `docs/domain.md`
  for vocabulary boundaries only.
- Phase-Accounting implementation:
  `tools/harness/phase-accounting/**`.
- Phase-Accounting tests:
  `tools/harness/phase-accounting/tests/**`.
- Static import-boundary rules and tests:
  `tools/harness/static-analysis/harness-import-boundary.mjs`,
  `tools/harness/tests/test-harness-contracts.mjs`, and
  `tools/harness/static-analysis/tests/test-frontend-import-boundaries.sh`.
- Smoke tier implementation and generated metadata:
  `tools/harness/smoke/run-harness-smoke-cli.mjs`,
  `tools/task_surface_manifest.json`,
  `tools/execution_topology_manifest.json`,
  `tools/task_surface.generated.mk`,
  `tools/scheduler_manifest.json`,
  `tools/browser_e2e_batch_manifest.json`, and
  `tools/execution_topology_render_index.json`.
- Generated-output policy and schemas:
  `tools/generated_artifact_policy.json`,
  `tools/schemas/cartulary.frontend_row_accounting.v3.schema.json`,
  `tools/schemas/cartulary.phase_slice_plan.v1.schema.json`, and adjacent
  frontend registry/map schemas.
- Frontend owner data:
  `tools/frontend_phase_registry.json`,
  `tools/frontend_phase_maps/*.json`, and
  `tools/frontend_visual_fixture_registry.json`.
- Current fact-gathering commands included `rg --files`,
  targeted `rg -n` searches for Phase-Accounting helper paths and
  compatibility terms, `wc -l` on Phase-Accounting modules/tests, direct
  import-boundary collector output, frontend registry/fixture summaries, and
  generated metadata reference searches.

## 3. Completed Historical Baseline

The following work is complete history. Do not reopen it merely because older
tracker text mentions it.

- G-01 through G-09 and WS-00 through WS-09 completed the first structural
  remediation pass: frontend manifest decomposition, future-growth guardrails,
  shared frontend row scope, evidence audit library split, phase-slice planner
  decomposition, table-based CLI dispatch, import-boundary enforcement,
  phase-slice v1 work-unit openness decision, FE-P to base phase bridge
  documentation, generated metadata refresh, and final targeted validation.
- G-10 through G-16 completed the second remediation pass on 2026-07-05:
  frontend phase validation decomposition, root-scoped cache/context tests,
  fail-closed frontend row accounting owner-data loading,
  selected-phase target-driven `frontend-evidence-audit` retained-root
  requiredness, base manifest validator decomposition, fixture-policy audit,
  and phase-accounting smoke scratch migration to `harness-scratch.sh`.
- Current code confirms the second pass landed. For example,
  `frontend/phase-manifest-core.mjs` is now a 14-line re-export layer,
  `phase-manifest-validation.mjs` is 377 lines and delegates runner, ID,
  source-scan, profile-claim, shape, frontend-fixture, and fixture-policy
  checks, and phase-accounting smoke fixtures now use the harness scratch
  helper.
- OR-008 remains standing validation debt only for live browser retained-root
  audit closure. It is not a blocker for this tracker-only refresh.

## 4. Simplification Policy

- Preserve public contracts. Public Make target names, command IDs, schema IDs,
  retained artifact paths, public input variables, failure mapping, cleanup,
  and Make wrapper semantics remain frozen unless a workstream is explicitly
  spec-first.
- Do not preserve private compatibility by default. Private helper imports,
  exact legacy path denylist rows, owner-local shims, and compatibility
  re-export layers must justify their existence with current callers or clear
  future value.
- Prefer deletion over relocation when a private layer only forwards to another
  private layer and no owner facade or generated metadata depends on it.
- Prefer owner facades over private helper paths for cross-subsystem callers.
  Do not delete owner facades merely because they are thin.
- Keep future `phaseN` and `FE-P<N>` growth append-only where required by the
  NLSpec and testable with isolated fixture roots.
- Treat fallback behavior as debt unless it is an owner-defined current
  profile behavior, a public compatibility contract, or a test fixture that
  proves a still-supported failure mode.
- When generated metadata must move, update owner inputs and regenerate through
  Make. Never hand-edit generated outputs to remove a stale reference.

## 5. Current Findings

- The direct harness import-boundary collector reports zero violations. It
  recognizes these Phase-Accounting owner facades:
  `frontend-phase-manifest.mjs`, `frontend-readiness.mjs`,
  `frontend-row-accounting.mjs`, `frontend/index.mjs`,
  `phase-manifest.mjs`, `phase-registry.mjs`, and
  `phase-slice-plan.mjs`.
- The same collector reports 45 `unsupported_private` helper paths, and all
  45 are currently absent from the working tree. The exact-path registry is
  therefore a historical denylist, not a live compatibility shim inventory.
- `tools/harness/frontend` is not present. Current protection for reintroduced
  frontend catch-all helpers comes from the private frontend catch-all rule
  plus exact unsupported-private entries.
- Current Phase-Accounting frontend private re-export layers remain:
  `frontend/phase-manifest-core.mjs`, `frontend/registry.mjs`,
  `frontend/validation.mjs`, and `frontend/visual-fixtures.mjs`.
  External callers use the owner facade `frontend-phase-manifest.mjs`, but
  owner-local modules still import some private re-export layers.
- `frontend-readiness.mjs` is a one-line facade over
  `frontend-phase-slice-plan.mjs` and is listed as an owner facade. It should
  not be deleted as private cleanup unless the owner boundary changes
  spec-first.
- `phase-fixture-policy.mjs` still contains package-reset default-budget code,
  dirty-table parsing, reset-conformance handling, and explicit package-reset
  validation. Current `tools/phase*_test_map.json` files declare no
  `fixture_policy.postgres="package_reset"` rows. Package-reset validation is
  exercised by smoke fixtures only.
- `cartulary.frontend_row_accounting.v3` currently requires `target`, `rows`,
  and `counts` compatibility fields in addition to the newer
  `scenario_results`, `row_results`, and `rollup` fields. The implementation
  emits the compatibility fields, and execution/browser/phase-accounting tests
  still read `accounting.rows`.
- The NLSpec says target and tool-run summaries must reference
  `frontend-row-accounting.json` rather than duplicating row details. It also
  records the v3 artifact as retaining compatibility row/count copies. Removing
  those artifact fields is therefore public schema work, not private cleanup.
- All four phase-accounting shell tests under
  `tools/harness/phase-accounting/tests/**` are referenced by current
  task-surface/topology harness-smoke metadata. No phase-accounting shell test
  is currently manual-only.
- Current generated metadata references active Phase-Accounting helper paths,
  including `check-phase-maps.sh`, `frontend-evidence-audit-cli.mjs`,
  `frontend-phase-manifest.mjs`, `frontend-row-accounting.mjs`,
  `phase-manifest.mjs`, `phase-manifest-cli.mjs`, `phase-registry.mjs`,
  `phase-slice-cli.mjs`, `phase-slice-plan.mjs`,
  `frontend-phase-slice-plan.mjs`, and phase-accounting smoke scripts.
  No generated metadata reference to a missing Phase-Accounting path was
  found.
- Current generated topology outputs are declared in
  `tools/execution_topology_manifest.json`: `tools/task_surface.generated.mk`,
  `tools/scheduler_manifest.json`, `tools/browser_e2e_batch_manifest.json`,
  and `tools/execution_topology_render_index.json`.
- The frontend registry currently declares `FE-P0` through `FE-P11`, all
  `status=active` and `row_rollup_state=active_green`. The frontend visual
  fixture registry currently has 15 `current` fixtures and 6 `missing`
  fixtures.

## 6. Public Compatibility Freeze

Future implementation sessions must preserve these public harness surfaces
unless a workstream is explicitly marked spec-first:

- Public Make targets, especially `phase-slice`, `service-backed-slice`,
  `frontend-evidence-audit`, `phase-map-check`, `phase-test-name-check`,
  `phase-ledgers`, `phase-ledger-drift`, `phase-schedules`,
  `phase-schedule-drift`, `harness-contract`, public smoke targets, and
  diagnostic/report targets.
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
  `frontend-evidence-audit-summary.json`, scheduler summaries/events,
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

## 7. Private and Legacy Deletion Policy

- Delete owner-local re-export layers once all owner-local imports use the real
  implementation modules and cross-owner callers use declared facades.
- Delete exact unsupported-private denylist rows only after an owner decision
  confirms the current protection is provided by owner-directory pattern rules
  or by a smaller current denylist. Because the NLSpec registry is closed,
  this contraction is spec-first.
- Delete fallback branches only when current owner data, current schemas, and
  characterization tests prove the branch is unreachable or unsupported.
- Do not keep a smoke fixture solely to prove behavior that has been retired
  or is now covered by an owner-controlled schema/import-boundary test.
- Do not delete owner facades simply because they are thin. Thin facades remain
  valuable when they are the declared cross-owner boundary.
- Do not hand-edit generated files to remove references. If a helper backing
  path is moved or deleted, update task-surface/topology owner inputs and run
  Make-owned generation and drift checks.

## 8. Out-of-Scope Boundaries

- Product HTTP routes, WebSocket behavior, workbook behavior, saved-view
  behavior, storage behavior, revision/history semantics, authentication,
  authorization, release publication behavior, and domain model changes.
- Domain vocabulary changes. This refresh records `domain vocabulary
  unchanged`.
- SQL migrations, DB queries, generated product contracts, dependency locks,
  `go.sum`, `pnpm-lock.yaml`, and tool-managed dependency/install artifacts.
- Hand-editing generated roots or generated outputs, including
  `internal/gen/**`, `packages/protocol-ts/src/generated/**`,
  `packages/ui-contracts/src/generated/**`,
  `tools/task_surface.generated.mk`, `tools/scheduler_manifest.json`,
  `tools/browser_e2e_batch_manifest.json`, and
  `tools/execution_topology_render_index.json`.
- Broad refactors outside Phase-Accounting and directly adjacent harness owner
  metadata unless required by import-boundary enforcement or generated
  metadata movement.
- Treating visual, accessibility, design-direction, cache, smoke, or other
  implementation-support evidence as product conformance without an owner
  boundary.
- Reopening completed G-01 through G-16 work without current evidence.

## 9. Detailed Gap Records

### G-17 Owner-Local Frontend Re-Export Layer Deletion

- Remediation: remove private frontend re-export layers that only forward to
  other owner-local modules, starting with `frontend/phase-manifest-core.mjs`,
  `frontend/registry.mjs`, `frontend/validation.mjs`, and
  `frontend/visual-fixtures.mjs`, after owner-local imports are moved to the
  true implementation modules. Preserve `frontend-phase-manifest.mjs` and
  `frontend/index.mjs` as owner facades.
- Affected areas: Phase-Accounting frontend modules, import-boundary fixtures
  that currently name `frontend/validation.mjs`, generated metadata only if
  owner inputs reference a moved path.
- Rationale: these private re-export layers were useful during the G-10 split
  but now add indirection without being public contracts.
- Expected long-term benefit: fewer private module names to preserve, clearer
  ownership of registry loading, validation, artifact validation, and visual
  fixture validation, and less chance that new code imports a compatibility
  layer instead of the real module.
- Compatibility or migration impact: no public Make target, schema ID, CLI
  behavior, retained artifact path, or owner facade export may change.
  Owner-local imports and harness-contract fixtures may change.
- Risks of leaving unresolved: private compatibility layers become de facto
  stable and future frontend validation changes have to preserve unnecessary
  aliases.
- Validation criteria: `node --check` for changed modules, direct import
  inventory proving no remaining imports of deleted private layers,
  harness import-boundary fixture update, `make phase-map-check`, `make
  json-shape-check`, `make harness-contract`, and `make lint-scripts`.

### G-18 Unsupported-Private Registry Contraction

- Remediation: replace the 45-entry exact-path unsupported-private historical
  denylist with the smallest owner-approved rule set that still rejects
  reintroduced legacy backend, scheduler, and frontend catch-all helper
  imports. Because `docs/testing-harness-nlspec.md` defines a closed helper
  ownership registry, this is spec-first.
- Affected areas: `docs/testing-harness-nlspec.md`,
  `tools/harness/static-analysis/harness-import-boundary.mjs`,
  `tools/harness/tests/test-harness-contracts.mjs`, task-surface/topology
  owner inputs only if backing paths move.
- Rationale: all currently listed unsupported-private helper paths are absent
  from the tree. Exact missing-path lists carry maintenance burden and can make
  the import-boundary rule look like it supports historical path names.
- Expected long-term benefit: import-boundary enforcement stays semantic
  rather than path-archival, future helper moves need fewer registry edits, and
  legacy paths are less likely to be treated as durable compatibility names.
- Compatibility or migration impact: public command behavior must not change.
  The machine-readable import-boundary report shape should remain stable, but
  `unsupported_private_helpers` contents may change only after owner approval.
- Risks of leaving unresolved: stale exact paths keep accumulating and the
  project spends future review time preserving names that no longer exist.
- Validation criteria: accepted NLSpec update, current live caller inventory,
  synthetic fixtures proving reintroduced legacy frontend/scheduler/backend
  imports still fail, direct import-boundary collector output with zero live
  violations, `make harness-contract`, `make frontend-import-boundary-check`
  when relevant, and generated drift checks if owner metadata moves.

### G-19 Package-Reset Fallback Contraction

- Remediation: delete unreachable package-reset default-budget fallback code
  while preserving explicit package-reset validation if the owner spec still
  admits the token. Current phase maps declare no `package_reset` rows; smoke
  fixtures should be reduced to the smallest checks that prove explicit
  package-reset policy remains rejected or validated as owner-approved.
- Affected areas: `phase-fixture-policy.mjs`,
  `phase-manifest-validation.mjs`, phase-accounting smoke fixtures, and
  fixture-policy schema/tests if owner policy changes.
- Rationale: the current implementation includes a default package-reset budget
  path even though no default policy returns `package_reset` and no phase-map
  row uses it. That is legacy tolerance, not current owner data.
- Expected long-term benefit: fixture-policy code becomes easier to reason
  about, future phase growth is pushed toward `transaction`, `group_clone`,
  `template_clone`, or explicit owner-approved reset policy, and broad reset
  compatibility does not become an accidental default.
- Compatibility or migration impact: no current phase-map behavior should
  change. Removing the `package_reset` token entirely would be spec-first and
  is not part of this gap unless the owner explicitly retires it.
- Risks of leaving unresolved: broad reset defaults remain available as an
  attractive but unsupported future shortcut, increasing fixture cost and
  ambiguity.
- Validation criteria: current phase-map audit still finds zero
  `package_reset` rows, focused fixture-policy tests cover explicit
  package-reset validation or rejection, `make phase-map-check`, `make
  phase-test-name-check`, `make harness-contract`, and service-backed slice
  JSON checks for affected phases if policy behavior changes.

### G-20 Frontend Row-Accounting Compatibility Field Retirement

- Remediation: make an owner decision on retiring the v3 artifact
  compatibility fields `target`, `rows`, and `counts`. If retirement is
  accepted, create a spec-first schema revision, move tests and consumers to
  `target_name`, `scenario_results`, `row_results`, and `rollup`, and keep old
  v3 artifacts diagnostic-only for old retained runs.
- Affected areas: `docs/testing-harness-nlspec.md`,
  `tools/schemas/cartulary.frontend_row_accounting.*.schema.json`,
  `frontend-row-accounting.mjs`, frontend evidence audit fixtures, execution
  and browser smoke tests that read `accounting.rows`, JSON shape fixtures, and
  release-readiness/evidence consumers if present.
- Rationale: v3 currently requires both the current normalized fields and
  compatibility copies. That doubles the artifact surface and keeps tests tied
  to legacy `rows`/`counts` semantics.
- Expected long-term benefit: row accounting has one authoritative shape,
  frontend evidence audit and release readiness consume the same normalized
  fields, and future schema evolution is less brittle.
- Compatibility or migration impact: this is public schema work. Do not remove
  v3 fields under the same schema ID unless the owner explicitly approves a
  same-ID compatibility break. Preferred path is a new schema ID with old v3
  diagnostic handling retained only where owner docs require it.
- Risks of leaving unresolved: new code can keep reading compatibility rows,
  making later schema cleanup harder and increasing the chance of mismatched
  rollup versus row-level semantics.
- Validation criteria: accepted schema/spec decision, updated JSON schema and
  shape fixtures, consumer inventory proving no current code reads retired
  fields, targeted frontend-unit/browser row-accounting tests, `make
  json-shape-check`, `make frontend-evidence-audit` with Make-owned retained
  roots when available, `make harness-contract`, and generated drift checks if
  metadata changes.

### G-21 Phase-Accounting Smoke Compatibility Pruning

- Remediation: prune or consolidate smoke assertions that exist only for
  retired compatibility behavior, while keeping all active Phase-Accounting
  smoke coverage reachable from owner-controlled tiers. Candidate areas are
  legacy row-name expectations in `test-check-phase-test-names.sh`,
  open-object compatibility assertions in `test-run-phase-slice.sh`, and v3
  row-accounting compatibility fixture generation in
  `test-frontend-evidence-audit.sh`.
- Affected areas: phase-accounting smoke scripts, harness-smoke metadata only
  if check membership changes, JSON shape fixtures if row-accounting schema
  changes.
- Rationale: all phase-accounting shell tests are currently reachable from
  generated smoke metadata, so the problem is not manual-only tests. The next
  simplification target is test content that protects compatibility behavior
  after the owner no longer values it.
- Expected long-term benefit: smaller, clearer smoke tests that protect
  current contracts instead of historical implementation shapes, with less
  friction when deleting private compatibility code.
- Compatibility or migration impact: public smoke target names and tier
  membership should remain stable unless owner metadata changes. Deleting
  active coverage requires named replacement coverage in an owner-controlled
  tier.
- Risks of leaving unresolved: compatibility-only assertions can block
  legitimate simplification and make future agents preserve behavior because a
  smoke test still mentions it.
- Validation criteria: per-assertion inventory, replacement-coverage notes for
  any deleted assertion, direct changed smoke scripts, `make
  run-harness-smoke-extended`, `make harness-contract`, `make lint-shell`, and
  generated drift checks if smoke metadata changes.

### G-22 Generated Metadata Backing-Path Review

- Remediation: audit Phase-Accounting backing paths in
  `tools/task_surface_manifest.json` and
  `tools/execution_topology_manifest.json` after any helper deletion or move.
  Remove stale references by updating owner inputs and running Make-owned
  generation, not by hand-editing generated outputs.
- Affected areas: task-surface/topology owner inputs, generated
  `tools/task_surface.generated.mk`, `tools/scheduler_manifest.json`,
  `tools/browser_e2e_batch_manifest.json`, and
  `tools/execution_topology_render_index.json`.
- Rationale: current generated metadata references active helper paths and no
  missing Phase-Accounting path was found. That can change when G-17 through
  G-21 move or delete helper files.
- Expected long-term benefit: generated metadata remains a truthful dependency
  inventory, drift checks catch stale backing paths, and future helpers are not
  kept alive solely because generated metadata still names them.
- Compatibility or migration impact: public Make target behavior and
  `command_id` values must not change. Backing script lists may change only
  through owner inputs and generated refresh.
- Risks of leaving unresolved: deleted helper paths can survive in generated
  metadata, causing drift, broken smoke dependencies, or accidental path
  resurrection.
- Validation criteria: owner-input diff review, `make phase-schedules` when
  schedule/topology inputs move, `make phase-schedule-drift`, `make
  generate-drift`, `make generated-artifact-policy-check`, `make
  json-shape-check`, and `make task-surface-report
  TASK_SURFACE_REPORT_ARGS=--all`.

## 10. Workstream Sequencing

### WS-00 Tracker Refresh

- Depends on: none.
- Sequencing: complete before implementation work.
- Changes: replace the completed G-10 through G-16 next-work content with this
  G-17 through G-22 simplification iteration.
- Risks: documentation drift only.
- Exit criteria: required sections, current findings, gap records,
  generated-artifact rules, validation matrix, and session log are current.
- Required validation: `make lint-markdown`.
- Generated-artifact handling: none.
- Tracker update requirements: record validation results and skipped checks.

### WS-01 Characterization and Import/Usage Inventory

- Depends on: WS-00.
- Sequencing: run before code deletion.
- Changes: no implementation changes; capture current imports, generated
  references, schema consumers, smoke reachability, package-reset usage, and
  public behavior.
- Risks: unrelated harness failures can obscure deletion risk.
- Exit criteria: inventory proves which private layers and compatibility fields
  have live callers, and identifies all generated metadata references.
- Required validation: direct import-boundary collector output, `rg` inventory
  for candidate paths/fields, `make phase-map-check`, `make
  phase-test-name-check`, `make json-shape-check`, `make harness-contract`,
  representative base and frontend `phase-slice` JSON checks.
- Generated-artifact handling: none unless inventory proves drift.
- Tracker update requirements: append inventory results before WS-02 starts.

### WS-02 Legacy and Private Helper Deletion Plan

- Depends on: WS-01.
- Sequencing: address G-17 first. Address G-18 only after an owner spec
  decision because the unsupported-private registry is closed in the NLSpec.
- Changes: delete owner-local frontend private re-export layers with no
  continuing value; optionally contract unsupported-private exact-path rules
  spec-first.
- Risks: accidental facade deletion, changed import-boundary diagnostics, or
  generated metadata still naming moved helpers.
- Exit criteria: cross-owner callers use owner facades, owner-local callers use
  real modules, deleted private layers have no imports, and import-boundary
  fixtures reject reintroduction.
- Required validation: `node --check` for changed modules, direct import
  inventory, import-boundary collector output, `make phase-map-check`, `make
  json-shape-check`, `make harness-contract`, and `make lint-scripts`.
- Generated-artifact handling: run WS-05 if any owner input or backing path
  moves.
- Tracker update requirements: mark G-17 complete/deferred and G-18
  complete/deferred with the owner decision.

### WS-03 Fallback and Fixture-Policy Simplification

- Depends on: WS-01.
- Sequencing: address G-19 before broad fixture-policy edits. Treat token
  retirement or phase-map policy changes as spec-first.
- Changes: delete unreachable package-reset default fallback behavior and
  tighten smoke coverage around explicit package-reset validation.
- Risks: service-backed fixture behavior drift, accidental rejection of an
  owner-approved explicit package-reset fixture, or insufficient coverage for
  current fixture policy tokens.
- Exit criteria: current phase maps still validate, no hidden package-reset
  default remains, and explicit package-reset handling is either still covered
  or owner-retired.
- Required validation: package-reset usage audit, focused fixture-policy
  fixtures, `make phase-map-check`, `make phase-test-name-check`, `make
  harness-contract`, and service-backed slice JSON checks if fixture behavior
  changes.
- Generated-artifact handling: refresh ledgers/schedules only if owner phase
  maps or topology inputs change.
- Tracker update requirements: record whether package reset remains an
  explicit owner-supported token or is deferred for retirement.

### WS-04 Test and Smoke Coverage Pruning or Consolidation

- Depends on: WS-01; G-20 requires spec-first schema decision before artifact
  shape pruning.
- Sequencing: prune compatibility-only assertions after the code/schema
  behavior they protect is deleted or owner-retired.
- Changes: remove obsolete compatibility assertions, migrate tests to current
  owner fields, and keep active coverage reachable from `extended` or another
  owner-controlled tier.
- Risks: deleting real coverage, breaking generated smoke metadata, or
  changing public smoke target names.
- Exit criteria: every deleted assertion has named replacement coverage or a
  recorded owner-retirement decision; no phase-accounting test is manual-only
  unless intentionally deleted.
- Required validation: direct changed smoke scripts, `make
  run-harness-smoke-extended`, `make harness-contract`, `make lint-shell`,
  `make json-shape-check` when schemas or JSON fixtures change.
- Generated-artifact handling: run WS-05 if harness smoke metadata changes.
- Tracker update requirements: list deleted assertions and replacement
  coverage.

### WS-05 Generated Metadata Review

- Depends on: any workstream that moves helper paths, owner metadata, schemas,
  task-surface inputs, topology inputs, phase maps, schedules, or smoke
  membership.
- Sequencing: run after owner input edits and before final validation.
- Changes: regenerate downstream outputs through Make-owned targets only.
- Risks: hand-editing generated outputs, stale backing paths, or command ID
  drift.
- Exit criteria: generated outputs match owner inputs, no generated metadata
  references deleted Phase-Accounting paths, and drift checks pass.
- Required validation: `make phase-schedules` when schedule/topology inputs
  move, `make phase-schedule-drift`, `make generate-drift`, `make
  generated-artifact-policy-check`, `make json-shape-check`, and `make
  task-surface-report TASK_SURFACE_REPORT_ARGS=--all`.
- Generated-artifact handling: never hand-edit generated outputs.
- Tracker update requirements: list owner inputs changed, generated outputs
  refreshed, and drift results.

### WS-06 Final Handoff and Finalize

- Depends on: completed implementation workstreams.
- Sequencing: run after targeted validation for the implementation slice.
- Changes: update this tracker or successor handoff with final status.
- Risks: incomplete handoff context, missing generated-artifact notes, or
  retained-run claims without evidence.
- Exit criteria: binary completion criteria are checked, open risks are
  explicit, and validation results are recorded.
- Required validation: `make agent-finalize`; broader `make check` only when
  implementation breadth warrants it.
- Generated-artifact handling: confirm generated outputs are unchanged or
  refreshed through Make.
- Tracker update requirements: record files changed, substantive edits,
  validations, skipped checks, run roots, failures, and next action.

## 11. Validation Matrix

- Tracker-only refresh: `make lint-markdown`; then `make agent-finalize`
  before final handoff; rerun `make lint-markdown` if the tracker log changes.
- Private frontend re-export deletion: `node --check` for changed modules,
  import inventory, import-boundary collector output, `make phase-map-check`,
  `make json-shape-check`, `make harness-contract`, `make lint-scripts`.
- Unsupported-private registry contraction: accepted NLSpec change, synthetic
  import-boundary fixtures, direct collector output, `make harness-contract`,
  and generated drift checks if owner metadata changes.
- Fixture-policy fallback deletion: package-reset usage audit, focused
  fixture-policy tests, `make phase-map-check`, `make phase-test-name-check`,
  `make harness-contract`, and targeted service-backed slice JSON checks if
  fixture behavior changes.
- Frontend row-accounting schema simplification: spec/schema decision, JSON
  shape fixtures, consumer inventory, targeted frontend-unit/browser
  row-accounting tests, `make json-shape-check`, `make harness-contract`, and
  `make frontend-evidence-audit` with Make-owned retained roots when
  available.
- Smoke pruning: direct changed smoke scripts, `make
  run-harness-smoke-extended`, `make harness-contract`, and `make lint-shell`.
- Generated metadata movement: `make phase-schedules`, `make
  phase-schedule-drift`, `make generate-drift`, `make
  generated-artifact-policy-check`, `make json-shape-check`, and `make
  task-surface-report TASK_SURFACE_REPORT_ARGS=--all`.
- Broad `make check`: require only when implementation breadth affects public
  scheduler behavior, retained artifacts, schema IDs, service-backed behavior,
  browser execution, or release gates. Otherwise targeted validation is
  sufficient and the skip reason must be recorded.

## 12. Open Risks and Blockers

- OR-018: private frontend re-export layers can become accidental stable
  import paths. Resolution path: complete G-17 through WS-02. Status: open.
- OR-019: the unsupported-private registry is currently a 45-path historical
  exact denylist with no live files. Resolution path: complete G-18
  spec-first or record that explicit archival denylisting is intentional.
  Status: open.
- OR-020: package-reset fallback code remains present despite no current
  phase-map usage. Resolution path: complete G-19 through WS-03. Status: open.
- OR-021: frontend row-accounting v3 still carries compatibility fields in the
  schema-owned artifact. Resolution path: complete G-20 spec-first or record
  that compatibility copies remain intentional for v3. Status: open.
- OR-022: compatibility-only smoke assertions can block deletion even when all
  tests are owner-tier reachable. Resolution path: complete G-21 through
  WS-04. Status: open.
- OR-023: generated metadata can retain deleted helper paths after future code
  movement. Resolution path: complete G-22 through WS-05 whenever owner inputs
  move. Status: standing generated-artifact risk.
- OR-008: live `frontend-evidence-audit` closure still requires fresh
  Make-owned retained roots for selected browser targets. Resolution path:
  produce or supply the retained roots before claiming live browser audit
  closure. Status: standing validation requirement, not a tracker blocker.

## 13. Binary Completion Criteria

This tracker refresh is complete only when all of the following are true:

- Current status and authority identify G-17+ as active work and G-01 through
  G-16 as historical baseline.
- Sources inspected include the owner docs, current tracker/session log,
  Phase-Accounting implementation/tests, import-boundary rules/tests,
  task-surface/topology owner inputs, generated-output policy, and current
  generated metadata references.
- Simplification policy explicitly preserves public contracts and rejects
  private compatibility preservation by default.
- Current findings are grounded in live files and command output, including
  zero import-boundary violations, 45 absent unsupported-private paths,
  private frontend re-export layers, package-reset non-usage in phase maps, v3
  row-accounting compatibility fields, smoke reachability, generated metadata
  references, and current frontend phase/fixture registry facts.
- Public compatibility freeze, private deletion policy, and out-of-scope
  boundaries are explicit.
- New gaps G-17 through G-22 each include remediation, affected areas,
  rationale, expected long-term benefit, compatibility or migration impact,
  risks of leaving unresolved, and validation criteria.
- Workstreams WS-00 through WS-06 include dependencies, sequencing, risks, exit
  criteria, required validation, generated-artifact handling, and tracker
  update requirements.
- Validation matrix, open risks/blockers, and session handoff log are current.
- `make lint-markdown` and `make agent-finalize` results are recorded for this
  tracker-only refresh.

Future implementation slices are complete only when their gap records are
marked completed or intentionally deferred, public compatibility impact is
recorded, generated-artifact rules were followed, and validation commands
passed or failures are classified with run roots.

## 14. Session Handoff Log

- 2026-07-05T10:37:26-04:00 through 2026-07-05T13:09:24-04:00,
  Codex GPT-5 historical baseline: completed G-01 through G-16 across two
  remediation iterations. Prior targeted harness validation passed. Treat this
  as historical baseline unless current code or owner docs prove regression.
- 2026-07-05T14:04:10-04:00, Codex GPT-5 WS-00 tracker refresh started.
  Files changed:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`.
  Commands run before edit: `git status --short`, tracker inspection,
  `docs/testing-harness-nlspec.md` and `docs/domain.md` inspection,
  Phase-Accounting file inventory, module/test line counts, compatibility
  term searches, static import-boundary rule/test inspection, direct
  import-boundary collector output, frontend registry and visual fixture
  summaries, generated-artifact policy inspection, task-surface/topology
  generated metadata reference searches, and smoke tier metadata inspection.
  Result: tracker replacement started from a completed G-10 through G-16
  implementation handoff. Next action: run `make lint-markdown`, run `make
  agent-finalize`, record results, and rerun markdown validation if this log
  changes.
- 2026-07-05T14:04:10-04:00 through 2026-07-05T14:07:35-04:00,
  Codex GPT-5 WS-00 tracker refresh completed. Files changed:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`.
  Commands run: `make lint-markdown`; `make agent-finalize`. Results:
  `make lint-markdown` passed with no output. `make agent-finalize` passed at
  `.cartulary/test-results/20260705T180731Z-p9` with
  `generated=unchanged files=0`, `duration=skipped`, `run_checks=skipped`,
  `reused=3`, `cache_hits=3`, and `results_dir=-`. Retained-run maintenance
  was skipped because `RESULTS_DIR` was unset. Broad `make check` was not run
  because this was a tracker-only documentation refresh and targeted
  validation covered the changed file. A post-log `make lint-markdown` rerun
  passed with no output. Next action: future implementation should start with
  WS-01 characterization and import/usage inventory before deleting code.
