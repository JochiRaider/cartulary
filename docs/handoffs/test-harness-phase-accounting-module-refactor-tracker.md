# test-harness-phase-accounting Next Simplification Tracker

## 1. Status and Authority

- Target label: `test-harness-phase-accounting`
- Target path: `tools/harness/phase-accounting`
- Tracker path:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Refresh date: 2026-07-05
- Current iteration status: complete. This iteration implemented the G-23
  through G-27 simplification/removal pass with the tracker updated after each
  workstream before the next workstream started.
- Active gap range: none for this tracker pass; G-23 through G-27 are complete.
- Completed historical baseline: G-01 through G-22 are complete unless current
  code, owner docs, or validation evidence proves a regression.
- Standing validation debt: OR-008 remains a live-browser retained-root audit
  closure requirement outside this tracker refresh. It is not a blocker for
  tracker-only planning.

Authority and source hierarchy:

1. `docs/testing-harness-nlspec.md` owns harness mechanics, including command
   invocation, target selection, scheduling, fixture lifecycle, artifact
   emission, cleanup, retained roots, summary emission, failure mapping, helper
   ownership, smoke tiers, cache behavior, and verification gates.
2. Core 00 through Core 04 own product behavior. Core 05 applies only to
   claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
3. `docs/domain.md` owns vocabulary and concept-boundary interpretation.
4. `docs/design.md` owns frontend design direction and token definitions, but
   it is not product-conformance evidence by itself.
5. Task-surface, topology, schedule, ledger, and generated Make outputs are
   downstream of owner inputs.

## 2. Current Completed Baseline

The previous Phase-Accounting remediation passes are complete history. Future
sessions should not reopen them merely because older handoff text mentioned
them.

- G-01 through G-09 completed the first structural remediation pass: frontend
  manifest decomposition, future-growth guardrails, shared frontend row scope,
  evidence audit library split, phase-slice planner decomposition, table-based
  CLI dispatch, import-boundary enforcement, phase-slice v1 work-unit openness,
  FE-P to base phase bridge documentation, generated metadata refresh, and
  targeted validation.
- G-10 through G-16 completed the second remediation pass: frontend phase
  validation decomposition, root-scoped cache/context tests, fail-closed
  frontend row accounting owner-data loading, selected-phase target-driven
  `frontend-evidence-audit` retained-root requiredness, base manifest validator
  decomposition, fixture-policy audit, and phase-accounting smoke scratch
  migration to `harness-scratch.sh`.
- G-17 through G-22 completed the last simplification pass: removal of private
  frontend re-export layers, unsupported-private registry contraction,
  package-reset fallback contraction, row-accounting compatibility field
  retirement through v4, smoke compatibility pruning, and generated metadata
  backing-path review.
- Current public compatibility posture remains frozen unless a workstream is
  explicitly marked spec-first: preserve public Make targets, stable
  `command_id` values, retained artifact paths, failure mapping, cleanup
  semantics, Make-wrapper behavior, and current schema IDs.
- Private helper compatibility is not preserved by default. A private helper
  path, shim, forwarding layer, or compatibility alias should remain only when
  current callers or clear ownership value justify it.

## 3. Sources and Evidence

This refresh is grounded in:

- Owner documents:
  `docs/testing-harness-nlspec.md`, `docs/domain.md`, and this tracker.
- Current implementation and tests:
  `tools/harness/phase-accounting/**`.
- Adjacent consumers:
  harness smoke tests, release evidence readers/tests, frontend evidence audit,
  test-output phase artifacts, Make wrappers, task-surface/topology manifests,
  schema attachments, generated metadata, phase maps, frontend phase maps,
  registry data, and fixture-policy owner data.

Commands and inspections used for this refresh:

- `git status --short`
- `find tools/harness/phase-accounting -maxdepth 4 -type f | sort`
- `rg --files docs tools/harness`
- Targeted `sed -n` reads of owner docs, tracker sections, implementation
  modules, wrappers, schemas, release-evidence readers, and smoke tests.
- Targeted `rg -n` searches for Phase-Accounting helper paths, compatibility
  terms, old row-accounting shapes, `package_reset`, `local_input_stamp`,
  generated metadata references, CLI command callers, frontend phase-map
  references, and smoke assertions.
- Node inventories confirming service-backed fixture-policy completeness and
  `package_reset` usage across current phase maps.

Key facts recorded for the next implementation session:

- Current phase maps have zero service-backed Go rows missing explicit
  Postgres fixture policy or budget.
- Current phase maps have zero `package_reset` rows.
- `package_reset` remains owner-supported compatibility behavior, but no
  current owner data depends on it.
- `cartulary.frontend_row_accounting.v4` is the current row-accounting schema.
  The v3 schema file is historical/diagnostic and must not be treated as
  current closure evidence.
- Release-readiness legacy artifact failure behavior is current contract and
  must be preserved unless superseded by a spec-first change.
- Several Phase-Manifest CLI subcommands appear private and unused by Make
  wrappers or adjacent harness callers: `go-postgres-*`, `support-go-*`,
  `go-family-*`, `support-go-family-*`, `go-verify-log`,
  `playwright-verify-*`, and `vitest-verify-run`.
- The frontend numeric bridge from `FE-P<N>` to `phase<N>` is current
  owner-specified behavior and must be changed only through a spec-first
  workstream.

## 4. Simplification Inventory

### Unused modules, exports, helpers, CLIs, and fixtures

- Candidate: unused private subcommands in
  `tools/harness/phase-accounting/phase-manifest-cli.mjs`.
  Evidence: caller inventory found current Make wrappers using commands such as
  `go-count`, `go-regex`, `playwright-files`, `playwright-grep`,
  `playwright-selection-report`, `vitest-files`, `vitest-grep`,
  `phase-policy-exceptions-validate`, and list/validate commands; the
  `go-postgres-*`, `support-go-*`, family-count/regex, verify-log, and
  `*-verify-*` helpers appear private and unused.
- Candidate: unused frontend manifest CLI command `phases`.
  Evidence: current caller search did not find a public wrapper dependency.
  Treat this as a candidate only until a final caller inventory confirms it.
- Candidate: helper exports in `phase-fixture-policy.mjs` used only by private
  CLI subcommands, including fixture assignment and reset-table assignment
  helpers. Remove only after the CLI callers are deleted or proven absent.
- Candidate: smoke fixtures that assert missing service-backed fixture policy
  should pass under defaults. Evidence: NLSpec requires explicit service-backed
  Postgres policy and budget; current owner phase maps already comply.

### Legacy compatibility branches and old artifact support

- Candidate: implicit service-backed Postgres fixture policy and budget
  defaults. Evidence: current owner phase maps do not need the fallback, and
  the NLSpec requires explicit declarations for service-backed rows.
- Candidate: `cartulary.frontend_row_accounting.v3` schema attachment.
  Evidence: v4 is current; v1/v2/v3 row-accounting artifacts are diagnostic
  only and cannot close current frontend rows. The implementation plan selects
  deletion of the v3 schema attachment rather than deprecation.
- Preserve: release-readiness rejection of unsupported legacy row-accounting
  shapes. Evidence: this is current failure behavior, not legacy support.
- Preserve: explicit `package_reset` validation. Evidence: `package_reset` is
  compatibility-only but still owner-supported when explicitly declared.

### Private facades, forwarding layers, and aliases

- Preserve current owner facades unless the owner boundary changes spec-first:
  `frontend-phase-manifest.mjs`, `frontend-readiness.mjs`,
  `frontend-row-accounting.mjs`, `frontend/index.mjs`,
  `phase-manifest.mjs`, `phase-registry.mjs`, and
  `phase-slice-plan.mjs`.
- Delete or simplify private forwarding layers when they only preserve a
  historical import path and have no current caller. The G-17 precedent applies
  to future private frontend or phase-manifest forwarding layers.

### Smoke assertions and generated metadata

- Smoke assertions should protect current contracts, not retired default
  behavior. Any fixture retained solely to prove a removed compatibility path
  must be deleted or rewritten as a current-contract failure case.
- Generated metadata must not reference deleted, moved, or legacy helper paths.
  If a cleanup moves an owner input, update the owner input and regenerate
  through Make-owned targets. Do not hand-edit generated outputs.

### Phase-map, frontend phase-map, registry, and fixture-policy behavior

- Fixture policy should be explicit for service-backed rows.
- `package_reset` should remain explicit-only compatibility behavior.
- The frontend `FE-P<N>` to base `phase<N>` mapping is a future-growth risk.
  The implementation plan selects a spec-first replacement with an explicit
  frontend-registry-owned base-phase join before code changes.
- Frontend phase maps, registry data, fixture registry data, row accounting,
  and evidence audit must continue to agree on current v4 row-accounting
  semantics.

### Adjacent consumers

- Release evidence: preserve v4 row-accounting validation and unsupported
  legacy failure records.
- Frontend evidence audit: preserve current artifact-ref behavior and current
  selected-target requiredness.
- Task-surface/topology: after any path deletion or command removal, verify no
  generated metadata still references the removed private surface.
- Contract tests and smoke tests: update only where they protect retired
  behavior or private helper surfaces.

## 5. Active Gap Records

### G-23 Remove implicit service-backed fixture defaults

- Remediation: remove fallback/default Postgres fixture policy and budget
  behavior from Phase-Accounting implementation and tests. Service-backed rows
  must rely on explicit owner phase-map declarations.
- Areas: implementation and tests; documentation only if current wording still
  describes direct helper defaults as supported current behavior.
- Rationale: current owner data already declares explicit service-backed
  policies and budgets. The fallback now creates hidden behavior rather than
  current value.
- Long-term benefit: future service-backed phase growth becomes explicit,
  reviewable, and easier to reason about.
- Compatibility and migration impact: current phase maps are unaffected.
  Synthetic fixtures and future service-backed rows must declare policy and
  budget explicitly.
- Risks if unresolved: new rows can silently inherit broad fixture behavior,
  weakening fixture-budget review and making phase growth more brittle.
- Validation criteria:
  inventory confirms zero current service-backed rows missing explicit policy
  or budget; `make phase-map-check`; `make phase-test-name-check`;
  `make harness-contract`; `make run-harness-smoke-extended`;
  representative `make phase-slice PHASE=phase4 JSON=1`; representative
  `make service-backed-slice PHASE=phase4 JSON=1`.

### G-24 Prune private unused Phase-Manifest and frontend CLI surfaces

- Remediation: delete unused private CLI subcommands after a final caller
  inventory. Remove helper exports/imports that exist only to support deleted
  commands.
- Areas: implementation and tests.
- Rationale: private helper command surfaces create compatibility expectations
  without public ownership value.
- Long-term benefit: smaller CLI dispatch, fewer helper exports, simpler tests,
  and lower chance of obsolete behavior being mistaken for contract.
- Compatibility and migration impact: no public Make target, stable
  `command_id`, schema ID, retained artifact path, failure mapping, or cleanup
  behavior should change.
- Risks if unresolved: dead private commands continue to preserve obsolete
  fixture-policy and verification paths.
- Validation criteria:
  final `rg` caller inventory recorded in the tracker; `node --check` for
  touched Phase-Accounting modules; `make phase-map-check`;
  `make harness-contract`; `make run-harness-smoke-extended`;
  `make lint-scripts`.

### G-25 Retire historical frontend row-accounting v3 schema attachment

- Remediation: delete
  `tools/schemas/cartulary.frontend_row_accounting.v3.schema.json`; keep v4 as
  the only current schema attachment. Preserve release-readiness rejection of
  unsupported legacy artifacts.
- Areas: schema, documentation, and tests.
- Rationale: old row-accounting shapes are diagnostic history and cannot close
  current frontend row accounting.
- Long-term benefit: retired fields such as old `target`, `rows`, and `counts`
  stop appearing as quasi-supported contract.
- Compatibility and migration impact: current v4 artifacts are unaffected.
  Historical v3 artifact validation by repo-local schema may go away if the
  schema is deleted.
- Risks if unresolved: stale schema files mislead future implementers and keep
  compatibility burden alive.
- Validation criteria:
  `rg cartulary.frontend_row_accounting.v3`; `make json-shape-check`;
  `make harness-contract`; release-readiness evidence tests if touched;
  `make lint-markdown` if docs change.

### G-26 Replace frontend numeric bridge before divergence

- Remediation: make a spec-first change that replaces the implicit
  `FE-P<N>` to `phase<N>` bridge with a registry-owned base-phase join before
  frontend and base phase numbering diverge.
- Areas: specification, schema, owner data, implementation, and tests; owner
  docs must change before implementation behavior changes.
- Rationale: the current numeric bridge is intentional current behavior, but it
  can become brittle as future frontend phases grow.
- Long-term benefit: frontend phase growth can remain append-only without
  relying on accidental numeric namespace equivalence.
- Compatibility and migration impact: high. This affects frontend registry
  schema/data, browser readiness, browser duration/baseline joins, generated
  freshness data, and related tests.
- Risks if unresolved: later frontend phase expansion may require a larger,
  riskier migration.
- Validation criteria after any implementation:
  `make phase-map-check`; relevant browser e2e target or targets for touched
  behavior; representative phase/service-backed slices; generated drift checks
  if owner inputs change.

### G-27 Consolidate Phase-Accounting smoke fixtures

- Remediation: remove smoke assertions that protect retired defaults and factor
  large synthetic fixture blobs into focused current-contract cases.
- Areas: tests.
- Rationale: smoke tests should make current contracts obvious and should not
  force obsolete compatibility paths to remain implemented.
- Long-term benefit: easier future simplification, smaller smoke fixtures, and
  clearer failure diagnosis.
- Compatibility and migration impact: none if current contract coverage remains
  equivalent.
- Risks if unresolved: future cleanup work will have to preserve retired
  behavior only because smoke fixtures still encode it.
- Validation criteria:
  `make run-harness-smoke-extended`; `make harness-contract`;
  `make lint-shell`; targeted `rg` evidence that retired default assertions
  are gone or rewritten as current-contract failures.

## 6. Workstreams

### WS-01 Characterization Inventory

- Scope and covered gaps: confirm caller usage, schema references, generated
  metadata references, fixture-policy completeness, `package_reset` usage,
  smoke assertions, and adjacent consumers before deleting code.
- Dependencies: current tracker baseline.
- Sequencing: complete before specification, code, or schema changes.
- Risks: deleting a private surface with an undiscovered caller could break a
  public wrapper indirectly.
- Exit criteria: tracker records exact inventory commands, current facts, and
  blockers.
- Required validation: targeted `rg`/Node inventories.
- Generated-artifact handling: inspect generated metadata but do not hand-edit
  it.
- Tracker update requirements: append exact commands and summarized results.

### WS-02 Spec and Documentation Cleanup

- Scope and covered gaps: update the NLSpec and live guides for explicit
  service-backed fixture declarations, v4 row-accounting wording, and explicit
  frontend/base phase mapping.
- Dependencies: WS-01.
- Sequencing: complete before behavior changes for G-23 or G-26.
- Risks: implementation may drift from owner docs if code changes first.
- Exit criteria: owner docs are internally consistent and the tracker records
  the explicit mapping decision.
- Required validation: markdown lint when available; targeted `rg` checks for
  retired current-doc wording.
- Generated-artifact handling: none unless owner inputs change.
- Tracker update requirements: record doc files changed and decision details.

### WS-03 Fixture Policy Simplification

- Scope and covered gaps: G-23 plus related smoke fixture updates.
- Dependencies: WS-02.
- Sequencing: before private helper pruning so dead fixture helper exports can
  be removed cleanly.
- Risks: tests may still assert missing policy defaults; update those tests to
  assert explicit current contracts.
- Exit criteria: no implicit service-backed Postgres fixture defaults remain;
  current phase maps still pass validation.
- Required validation: G-23 validation criteria.
- Generated-artifact handling: regenerate only if owner inputs change.
- Tracker update requirements: record removed helpers, test updates, command
  results, and retained explicit compatibility behavior.

### WS-04 Private CLI and Helper Pruning

- Scope and covered gaps: G-24.
- Dependencies: WS-01 and WS-03.
- Sequencing: after caller inventory and fixture default cleanup.
- Risks: a private command may be consumed by an undocumented local workflow.
  Public Make and harness wrappers are the compatibility boundary.
- Exit criteria: unused private commands and exclusive helper exports are gone;
  public wrappers continue to pass.
- Required validation: G-24 validation criteria.
- Generated-artifact handling: if generated metadata references removed private
  paths, update owner inputs and regenerate through Make.
- Tracker update requirements: list deleted commands and the caller inventory
  that justified deletion.

### WS-05 Historical Schema Cleanup

- Scope and covered gaps: G-25.
- Dependencies: WS-02.
- Sequencing: after schema reference inventory and before final validation.
- Risks: docs or tools may still mention v3 as a current closure schema.
- Exit criteria: v4 remains the only current row-accounting schema attachment,
  while release-readiness legacy failure behavior is preserved.
- Required validation: G-25 validation criteria.
- Generated-artifact handling: delete the historical schema attachment; do not
  revise v4 unless a separate schema change is specified.
- Tracker update requirements: record schema deletion, references updated, and
  release-readiness behavior preserved.

### WS-06 Frontend Phase Mapping Implementation

- Scope and covered gaps: G-26.
- Dependencies: WS-02.
- Sequencing: after owner docs declare explicit registry-owned mapping.
- Risks: treating the bridge as private cleanup would break current browser and
  frontend phase behavior.
- Exit criteria: current FE-P0 through FE-P11 behavior is preserved through
  explicit registry data, and synthetic divergence is test-covered.
- Required validation: frontend registry/map validation, browser selection or
  duration tests, representative slices, and generated drift checks when owner
  inputs change.
- Generated-artifact handling: update owner inputs and regenerate through Make
  when generated frontend ledgers or freshness data changes.
- Tracker update requirements: record schema/data/code/test migration and any
  generated-artifact handling.

### WS-07 Smoke and Metadata Review

- Scope and covered gaps: G-27 plus generated metadata references to deleted,
  moved, or legacy paths.
- Dependencies: WS-03 through WS-06.
- Sequencing: near the end of the cleanup pass.
- Risks: smoke suite may lose meaningful current-contract coverage while
  removing obsolete fixtures.
- Exit criteria: smoke assertions protect current contracts and generated
  metadata has no stale Phase-Accounting references.
- Required validation: G-27 validation criteria plus generated artifact policy
  checks if owner inputs or generated outputs are affected.
- Generated-artifact handling: never hand-edit generated outputs; update owner
  inputs first.
- Tracker update requirements: record fixtures removed or rewritten and any
  generated metadata drift checks.

### WS-08 Final Validation and Handoff

- Scope and covered gaps: completion accounting for the simplification pass.
- Dependencies: all attempted workstreams.
- Sequencing: last.
- Risks: incomplete evidence can make later sessions repeat characterization.
- Exit criteria: tracker records completed gaps, remaining gaps, files changed,
  commands run, skipped checks, retained risks, and next actions.
- Required validation: `make agent-finalize`; broaden to `make check` only if
  public scheduler behavior, schemas, service-backed behavior, browser
  execution, generated metadata, or release gates changed.
- Generated-artifact handling: pass `RESULTS_DIR` to `make agent-finalize` only
  when a successful retained full warm check run root is available.
- Tracker update requirements: append final session evidence and completion
  state.

## 7. Validation Matrix

Use the narrowest Make-owned target that covers the proposed change, then
broaden only when the change affects public scheduler behavior, schemas,
service-backed behavior, browser execution, generated metadata, or release
gates.

| Change area | Required validation |
| --- | --- |
| Tracker/docs-only refresh | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check` when schema references are touched |
| Fixture policy or phase-map behavior | `make phase-map-check`; `make phase-test-name-check`; representative `make phase-slice PHASE=phase4 JSON=1`; representative `make service-backed-slice PHASE=phase4 JSON=1` |
| CLI/helper pruning | Final `rg` caller inventory; `node --check` for touched modules; `make phase-map-check`; `make harness-contract`; `make run-harness-smoke-extended`; `make lint-scripts` |
| Schema cleanup | `rg cartulary.frontend_row_accounting.v3`; `make json-shape-check`; `make harness-contract`; release-readiness evidence tests if touched |
| Smoke fixture cleanup | `make run-harness-smoke-extended`; `make harness-contract`; `make lint-shell` |
| Browser/frontend phase mapping | `make phase-map-check`; frontend registry/map validation; targeted browser selection or duration tests; representative slices; generated drift checks if owner inputs change |
| Generated metadata movement | Owner-input update; Make-owned generator or drift target; `make generated-artifact-policy-check`; relevant topology/task-surface drift target |
| Final handoff | `make agent-finalize`; `RESULTS_DIR=<successful full warm check run root>` only when retained evidence exists |

## 8. Public Compatibility Rules for Implementation Sessions

Implementation sessions must preserve these surfaces unless a workstream is
explicitly spec-first:

- Public Make targets.
- Stable `command_id` values.
- Retained artifact paths and artifact-ref semantics.
- Failure mapping and exit-code behavior.
- Cleanup semantics.
- Make-wrapper behavior.
- Current schema IDs, especially `cartulary.frontend_row_accounting.v4` and
  `cartulary.phase_slice_plan.v1`.

Private surfaces may be removed when evidence shows they are unused,
legacy-only, duplicate, or unsupported by owner docs. Do not keep behavior
merely because it exists.

## 9. Risks and Guardrails

- Do not delete owner facades solely because they are thin. Owner facades are
  part of the boundary model until owner docs change.
- Do not remove release-readiness legacy-failure behavior while cleaning up old
  row-accounting schema attachments.
- Do not treat `package_reset` as deleted behavior. The current cleanup target
  is implicit fallback support, not explicit compatibility validation.
- Do not change frontend phase numeric bridge behavior without a spec-first
  workstream.
- Do not hand-edit generated outputs. Generated metadata cleanup must flow from
  owner inputs and Make-owned generation.

## 10. Session Handoff Log

### 2026-07-05 Implementation session - WS-01 characterization

- Workstream completed:
  WS-01 Characterization Inventory.
- Files inspected:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`,
  `docs/testing-harness-nlspec.md`, `tools/harness/phase-accounting/**`,
  `tools/harness/smoke/target-plan-smoke-suite.sh`,
  `tools/schemas/cartulary.frontend_row_accounting.v3.schema.json`,
  `tools/schemas/cartulary.frontend_row_accounting.v4.schema.json`, and
  release-readiness evidence readers/tests.
- Commands run:
  `git status --short`; targeted `rg -n` caller inventories for unused private
  Phase-Manifest commands and the frontend `phases` command; targeted `rg -n`
  searches for `cartulary.frontend_row_accounting.v3`, `v3 row accounting`,
  fixture defaults, `effective*PostgresFixture*`, and missing-policy smoke
  assertions; Node inventory of service-backed fixture-policy completeness and
  `package_reset` usage.
- Findings:
  worktree was clean before implementation; current 12 base phase-map files
  have zero service-backed Go rows missing explicit Postgres fixture policy or
  budget; current phase maps have zero `package_reset` rows. The private
  Phase-Manifest command candidates are referenced only by their defining CLI
  module and this tracker. The frontend `phases` command is referenced only by
  its defining CLI module. Current non-archive v3 row-accounting references are
  the NLSpec diagnostic-only statement, two frontend guide references that need
  current wording, one frontend validation diagnostic string, and the historical
  v3 schema attachment. The smoke suite still contains a missing-policy fixture
  that expects service-backed fixture defaults to pass.
- Generated-artifact handling:
  no generated outputs were edited during WS-01.
- Next action:
  start WS-02 by updating owner docs and live guide wording before behavior or
  schema changes.

### 2026-07-05 Implementation session - WS-02 spec and docs

- Workstream completed:
  WS-02 Spec and Documentation Cleanup.
- Files changed:
  `docs/testing-harness-nlspec.md`,
  `docs/guides/cartulary_frontend_implementation_testing_guide.md`,
  `tools/harness/phase-accounting/frontend/phase-map-validation.mjs`, and this
  tracker.
- Substantive edits:
  the harness NLSpec now makes `cartulary.frontend_phase_registry.v4` the
  current frontend registry contract, requires each frontend registry phase to
  declare explicit `base_phase_join`, requires browser readiness/duration/base
  Playwright selection joins to use that field instead of numeric suffix
  derivation, and removes the service-backed Postgres helper-default allowance.
  The live frontend guide now describes preflight row accounting as current v4
  row accounting and lists explicit `base_phase_join` as part of frontend
  registry validation. The frontend phase-map validation diagnostic no longer
  says current preflight row accounting is v3.
- Commands run:
  targeted `rg -n` checks for `cartulary.frontend_phase_registry.v3`,
  `preflight v3 row accounting`, `helper defaults are only`, and the retired
  numeric compatibility bridge wording in live docs.
- Validation:
  targeted current-doc wording checks passed. Markdown lint is deferred until
  final validation because later workstreams will update docs and the tracker
  again.
- Generated-artifact handling:
  no generated outputs were edited during WS-02.
- Next action:
  start WS-03 by removing implicit service-backed Postgres fixture defaults
  from implementation consumers and rewriting the missing-policy smoke fixture
  as a fail-closed current-contract check.

### 2026-07-05 Implementation session - WS-03 fixture policy

- Workstream completed:
  WS-03 Fixture Policy Simplification.
- Files changed:
  `tools/harness/phase-accounting/phase-fixture-policy.mjs`,
  `tools/harness/phase-accounting/phase-manifest-validation.mjs`,
  `tools/harness/phase-accounting/phase-manifest.mjs`,
  `tools/harness/phase-accounting/phase-manifest-cli.mjs`,
  `tools/harness/backend/target-plan.mjs`,
  `tools/harness/smoke/target-plan-smoke-suite.sh`, and this tracker.
- Substantive edits:
  removed the default Postgres fixture policy and budget functions from
  Phase-Accounting; rewired manifest validation and backend target-plan rows to
  consume explicit manifest fixture policy and budget only; removed the old
  `effective*PostgresFixture*` facade exports; and inverted the synthetic
  missing-policy smoke fixture so missing service-backed policy now fails
  closed.
- Commands run:
  `node --check` on touched Phase-Accounting/backend modules;
  `make phase-test-name-check`; `make phase-map-check VERBOSE=1`;
  direct `node tools/harness/phase-accounting/phase-map-check-cli.mjs phase4`;
  target-plan JSON inventory confirming service-backed Go fixture policy and
  budget coverage; targeted `rg -n` for removed default/effective fixture
  symbols; representative `make phase-slice PHASE=phase4 JSON=1`; and
  representative `make service-backed-slice PHASE=phase4 JSON=1`.
- Validation:
  syntax checks passed; `make phase-test-name-check` passed; direct phase4 map
  validation passed; representative phase and service-backed slice plan
  commands passed; target-plan inventory found 221 service-backed Go rows with
  explicit fixture policy and budget; removed default/effective fixture symbols
  no longer appear. Full `make phase-map-check` currently fails because WS-02
  changed the frontend guide digest and frontend registry freshness has not yet
  been migrated; rerun after WS-06 updates registry/freshness owner data.
- Generated-artifact handling:
  no generated outputs were edited during WS-03.
- Next action:
  start WS-04 by pruning unused private Phase-Manifest and frontend CLI
  surfaces now that fixture default helper exports are unnecessary.

### 2026-07-05 Implementation session - WS-04 private CLI pruning

- Workstream completed:
  WS-04 Private CLI and Helper Pruning.
- Files changed:
  `tools/harness/phase-accounting/phase-manifest-cli.mjs`,
  `tools/harness/phase-accounting/phase-fixture-policy.mjs`,
  `tools/harness/phase-accounting/frontend/cli.mjs`, and this tracker.
- Substantive edits:
  removed unused private Phase-Manifest subcommands for `go-postgres-*`,
  `support-go-*`, `go-family-*`, `support-go-family-*`, `go-verify-log`,
  `playwright-verify-*`, and `vitest-verify-run`; removed the unused frontend
  manifest `phases` command; and removed fixture assignment/reset-table helper
  exports that existed only for deleted private command paths.
- Commands run:
  final targeted `rg -n` caller inventory for deleted private commands;
  targeted `rg -n` for deleted helper exports and frontend `phases` usage;
  `node --check` for `phase-manifest-cli.mjs`, `phase-fixture-policy.mjs`, and
  `frontend/cli.mjs`.
- Validation:
  syntax checks passed. Deleted private command names now appear only in this
  tracker's historical inventory. Deleted helper exports and the frontend
  `phases` command no longer appear in current Phase-Accounting code.
- Generated-artifact handling:
  no generated outputs were edited during WS-04.
- Next action:
  start WS-05 by deleting the historical frontend row-accounting v3 schema
  attachment and preserving current release-readiness legacy rejection.

### 2026-07-05 Tracker refresh session

- Files inspected:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`,
  `docs/testing-harness-nlspec.md`, `docs/domain.md`,
  `tools/harness/phase-accounting/**`,
  `tools/schemas/cartulary.frontend_row_accounting.v4.schema.json`,
  `tools/schemas/cartulary.frontend_row_accounting.v3.schema.json`,
  release-evidence readers/tests, test-output phase artifacts, wrapper
  scripts, generated metadata manifests, and smoke suites.
- Commands run:
  `git status --short`; `rg --files`; `find`; `sed -n`; targeted `rg -n`
  searches for Phase-Accounting paths, CLI callers, schema IDs, compatibility
  fields, generated metadata, `package_reset`, and fixture-policy terms; Node
  inventories for explicit service-backed fixture-policy coverage and
  `package_reset` usage.
- Verification run for this tracker refresh:
  `make lint-markdown`, `make generated-artifact-policy-check`,
  `make json-shape-check`, `git diff --check`, and `make agent-finalize`
  passed during the tracker refresh.
- Findings:
  G-01 through G-22 remain completed baseline; next active gaps are G-23
  through G-27. Current phase maps have zero service-backed rows missing
  explicit policy or budget and zero `package_reset` rows. v4 is the current
  row-accounting schema; v3 is historical. Several private CLI commands appear
  unused. The frontend numeric bridge is current spec-owned behavior.
- Files changed in this tracker refresh:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`.
- Substantive edit:
  replaced the prior G-17 through G-22 active tracker with a G-23 through G-27
  simplification/removal iteration, including workstreams, compatibility
  posture, validation matrix, risks, and handoff evidence.
- Skipped checks:
  implementation, schema-edit, generated-output, browser, service-backed, and
  broad `make check` validation are skipped because this pass intentionally
  updates only the handoff tracker. Retained-run maintenance was skipped
  because `RESULTS_DIR` was unset.
- Next action:
  start WS-01 characterization inventory before implementing any deletion or
  simplification work.

### 2026-07-05 Implementation session - WS-05 historical schema cleanup

- Workstream completed:
  WS-05 Historical Schema Cleanup.
- Files changed:
  deleted `tools/schemas/cartulary.frontend_row_accounting.v3.schema.json`;
  updated `tools/release-evidence/tests/test-release-readiness-evidence.sh`;
  and updated this tracker.
- Substantive edits:
  removed the historical v3 row-accounting schema attachment from the current
  schema set and extended release-readiness evidence coverage so both
  `cartulary.frontend_row_accounting.v2` and
  `cartulary.frontend_row_accounting.v3` retained artifacts are rejected as
  legacy diagnostic-only evidence.
- Commands run:
  `find tools/schemas -maxdepth 1 -type f -name
  'cartulary.frontend_row_accounting.v*.schema.json'`; targeted `rg -n`
  searches for current non-archive v3 row-accounting wording; `bash -n
  tools/release-evidence/tests/test-release-readiness-evidence.sh`; generated
  metadata inspection for the release-readiness evidence smoke owner; attempted
  `make harness-smoke-release-readiness-evidence`; `make help-all | rg -n
  'release-readiness|harness-smoke-release'`; `make run-harness-smoke-extended`;
  and `make explain-run RESULTS_DIR=.cartulary/test-results/20260705T192325Z-p200201`.
- Validation:
  only `tools/schemas/cartulary.frontend_row_accounting.v4.schema.json` remains
  in the current row-accounting schema attachment set. Non-archive v3
  row-accounting references are limited to this tracker, the NLSpec
  diagnostic-only retained-artifact statement, and the explicit v3 legacy
  rejection test. Shell syntax passed. The release-readiness evidence smoke
  child passed inside run root
  `.cartulary/test-results/20260705T192325Z-p200201`. The aggregate
  `make run-harness-smoke-extended` failed at
  `harness-smoke-print-target-plan` because
  `tools/frontend_phase_registry.json.guide_digest` is stale after WS-02; this
  is the expected pending WS-06 freshness migration, not a release-readiness
  rejection failure. The direct target name
  `harness-smoke-release-readiness-evidence` is metadata-owned but not exposed
  as a public Make target.
- Generated-artifact handling:
  no generated outputs were edited during WS-05.
- Next action:
  start WS-06 by implementing the explicit frontend registry-owned
  `base_phase_join` mapping, migrating the frontend registry schema/data, and
  refreshing frontend freshness evidence from owner inputs.

### 2026-07-05 Implementation session - WS-06 frontend phase mapping

- Workstream completed:
  WS-06 Frontend Phase Mapping Implementation.
- Files changed:
  `docs/testing-harness-nlspec.md`,
  `docs/guides/cartulary_frontend_implementation_testing_guide.md`,
  `tools/schemas/cartulary.frontend_phase_registry.v4.schema.json`,
  `tools/frontend_phase_registry.json`,
  `tools/frontend_phase_maps/fe_p*_test_map.json`,
  generated frontend ledgers under
  `docs/testing/frontend_phase_coverage_ledgers/`,
  `tools/harness/phase-accounting/frontend/{constants,freshness,ledger,phase-ids,registry-loader}.mjs`,
  `tools/harness/phase-accounting/phase-selection.mjs`,
  `tools/harness/browser/browser-duration-accounting.mjs`,
  `tools/harness/generated-artifacts/check-json-shapes.mjs`,
  `tools/harness/phase-accounting/tests/test-run-phase-slice.sh`, and this
  tracker.
- Substantive edits:
  introduced `cartulary.frontend_phase_registry.v4`; made
  `base_phase_join` a required nullable registry field; populated current
  FE-P0 through FE-P11 joins to `phase0` through `phase11`; validated non-null
  joins against `tools/phase_registry.json`; replaced the old numeric
  `FE-P<N> -> phase<N>` browser selection and duration bridge with explicit
  registry data; included the join in frontend freshness digests; rendered the
  join in generated frontend ledgers; and refreshed guide, manifest, ledger,
  and evidence freshness digests from owner inputs.
- Commands run:
  `sha256sum` for changed guide/schema inputs; `node --check` for touched
  frontend registry, phase-selection, browser-duration, freshness, and JSON
  shape modules; `make phase-ledgers`; `node
  tools/harness/phase-accounting/frontend-phase-manifest.mjs validate`;
  targeted `rg -n` checks for the retired numeric bridge and stale v3 registry
  references; `make phase-map-check`; `make json-shape-check`; `make
  phase-ledger-drift`; `make phase-slice PHASE=phase4 JSON=1`; `make
  service-backed-slice PHASE=phase4 JSON=1`; `make
  browser-e2e-duration-baseline-drift`; direct browser shard-plan selection
  checks; direct frontend base-join helper checks; and `make
  phase-test-name-check`.
- Validation:
  frontend artifact validation passed; phase map, JSON shape, generated ledger
  drift, representative base phase-slice, representative service-backed slice,
  direct browser shard planning, direct selected frontend row planning, direct
  base-join helper checks, and phase test-name validation passed. Direct
  shard planning selected 22 FE browser entries into explicit base phases and
  selected `FE-I-P5-01` into `phase5`. `make
  browser-e2e-duration-baseline-drift` failed with
  `retained browser duration evidence has no observed Playwright timings`;
  this command is not valid without retained browser timing evidence and did
  not indicate a join-selection drift.
- Generated-artifact handling:
  generated frontend coverage ledgers were refreshed only by the public `make
  phase-ledgers` target, then verified with `make phase-ledger-drift`.
- Next action:
  start WS-07 by reviewing smoke fixtures and generated metadata for stale
  deleted helper paths, retired default assertions, stale v3 row-accounting
  schema references, and explicit frontend join coverage.

### 2026-07-05 Implementation session - WS-07 smoke and metadata review

- Workstream completed:
  WS-07 Smoke and Metadata Review.
- Files changed:
  `tools/harness/phase-accounting/frontend/registry-loader.mjs`,
  `tools/harness/smoke/target-plan-smoke-suite.sh`, and this tracker.
- Substantive edits:
  changed frontend/base join validation to read the base phase registry from
  the same root as `tools/frontend_phase_registry.json` instead of the ambient
  `CARTULARY_PHASE_MANIFEST_ROOT`; this keeps frontend owner validation stable
  while synthetic base-phase smoke fixtures redirect only base manifest data.
  Updated positive smoke fixtures that create service-backed synthetic rows so
  they declare explicit transaction fixture policy and budget. The dedicated
  missing-policy and missing-budget negative fixtures remain fail-closed.
- Metadata and stale-reference review:
  exact deleted private command searches found no current references outside
  this tracker; deleted Phase-Accounting helper exports no longer appear in
  current Phase-Accounting code. Remaining `fixturePolicyAssignments` and
  `resetTableAssignments` references are live backend aggregate helper
  surfaces used by public wrappers, not the deleted Phase-Accounting private
  helper exports. Current non-archive v3 row-accounting references are limited
  to the NLSpec diagnostic-only retained-artifact statement and the explicit
  v3 release-readiness rejection test.
- Commands run:
  targeted `rg -n` stale-reference inventories for deleted private commands,
  deleted helper exports, retired fixture defaults, v3 row-accounting schema
  paths, and frontend numeric-bridge terms; `node --check
  tools/harness/phase-accounting/frontend/registry-loader.mjs`; frontend
  artifact validation with and without ambient `CARTULARY_PHASE_MANIFEST_ROOT`;
  `make generated-artifact-policy-check`; `make lint-shell`; `make
  run-harness-smoke-extended`; and isolated `make harness-contract` reruns.
- Validation:
  `make run-harness-smoke-extended` passed with run root
  `.cartulary/test-results/20260705T193943Z-p319751`; `make lint-shell`
  passed; `make generated-artifact-policy-check` passed; and the final
  isolated `make harness-contract` rerun passed. Earlier WS-07 validation
  attempts exposed the stale synthetic-root join coupling and stale
  missing-policy smoke fixtures; both were fixed before the passing smoke run.
  One intermediate `harness-contract` run failed on a SeaweedFS readiness
  timeout in `test-dev-services-lifecycle.sh`; an immediate isolated rerun
  passed, so no code change was made for that transient infrastructure
  failure.
- Generated-artifact handling:
  no generated outputs were edited during WS-07.
- Next action:
  start WS-08 final validation and handoff: run `make agent-finalize`, rerun
  focused drift/contract checks as needed, then update this tracker with final
  retained-run status, remaining risks, and handoff summary.

### 2026-07-05 Implementation session - WS-08 final validation and handoff

- Workstream completed:
  WS-08 Final Validation and Handoff.
- Files changed:
  this tracker was updated to mark the G-23 through G-27 pass complete and to
  record final validation evidence.
- Final scope summary:
  G-23 now fails closed for service-backed Go rows without explicit
  `fixture_policy.postgres` and `fixture_budget.postgres`; G-24 private CLI
  and helper-only surfaces were pruned while public wrappers remained intact;
  G-25 removed the current v3 frontend row-accounting schema attachment while
  preserving explicit legacy rejection; G-26 migrated the frontend registry to
  v4 with explicit `base_phase_join`; and G-27 smoke fixtures now protect the
  current contract instead of retired defaults.
- Final commands run:
  `make agent-finalize`; `git diff --check`; `make lint-markdown`; `make
  phase-map-check`; `make json-shape-check`; `make phase-test-name-check`;
  `make generated-artifact-policy-check`; `make phase-ledger-drift`; `make
  harness-contract`; `make check`; isolated `make frontend-unit`; and a final
  rerun of `make check`.
- Final validation results:
  all focused final checks passed. `make agent-finalize` passed with
  `results_dir=-`, so retained-run maintenance was skipped because
  `RESULTS_DIR` was unset. The first broad `make check` run failed in
  `frontend-unit` on two frontend product test assertions
  (`App.landing.test.tsx` and
  `WorkbookShell.phase4.support.test.tsx`); isolated `make frontend-unit`
  immediately passed at
  `.cartulary/test-results/20260705T194723Z-p499364`. The final broad `make
  check` rerun passed at
  `.cartulary/test-results/20260705T194759Z-p501131` with 252/252 work units,
  981 tests, and 0 failures.
- Final inventories:
  no current non-archive code/docs references the retired numeric
  `frontendPhaseToBasePhase` bridge or `cartulary.frontend_phase_registry.v3`
  outside the retained historical v3 registry schema file. Current
  non-archive v3 row-accounting references are limited to the NLSpec
  diagnostic-only retained-artifact statement and the explicit
  release-readiness v3 rejection test. The current row-accounting schema
  attachment set contains only v4.
- Retained-run status:
  no retained successful full warm-check root was supplied to
  `make agent-finalize`, so retained-run duration/baseline maintenance was
  skipped by design for this handoff.
- Remaining risks:
  no open implementation gaps remain for G-23 through G-27. The only observed
  residual instability was transient validation behavior: one intermediate
  `harness-contract` run hit a SeaweedFS readiness timeout and one first broad
  `make check` run hit frontend-unit assertions; both passed on isolated or
  final rerun without code changes.
- Handoff status:
  complete. Future work should treat `cartulary.frontend_phase_registry.v4`,
  `base_phase_join`, explicit service-backed fixture policy/budget data, and
  `cartulary.frontend_row_accounting.v4` as the current contracts.
