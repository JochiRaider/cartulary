# test-harness-phase-accounting Next Simplification Tracker

## 1. Status and Authority

- Target label: `test-harness-phase-accounting`
- Target path: `tools/harness/phase-accounting`
- Tracker path:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Refresh date: 2026-07-05
- Current iteration status: planning tracker refresh only. This iteration
  identifies the next simplification and removal pass; it does not implement
  Phase-Accounting code, schema, generated metadata, or test changes.
- Active gap range: G-23 and later.
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
  only and cannot close current frontend rows.
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
- The frontend `FE-P<N>` to base `phase<N>` mapping is a future-growth risk but
  is current owner behavior. Treat any change as spec-first.
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

- Remediation: delete or explicitly deprecate
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

### G-26 Decide frontend phase growth mapping before numeric divergence

- Remediation: make a spec-first decision to either preserve the
  `FE-P<N>` to `phase<N>` bridge as intentional compatibility or replace it
  with registry-owned mapping before frontend and base phase numbering diverge.
- Areas: specification first; implementation and tests only after owner docs
  change.
- Rationale: the current numeric bridge is intentional current behavior, but it
  can become brittle as future frontend phases grow.
- Long-term benefit: frontend phase growth can remain append-only without
  relying on accidental numeric namespace equivalence.
- Compatibility and migration impact: high. This affects frontend phase maps,
  browser readiness, browser duration/baseline joins, and related tests.
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

### WS-00 Tracker Refresh

- Scope and covered gaps: update this tracker with the G-23 through G-27
  simplification iteration.
- Dependencies: none.
- Sequencing: complete before implementation work.
- Risks: stale tracker state could cause future sessions to reopen completed
  G-01 through G-22 work.
- Exit criteria: tracker names G-01 through G-22 as historical baseline,
  identifies G-23 through G-27 as active, and records current compatibility
  posture.
- Required validation: Markdown lint if available for docs-only change.
- Generated-artifact handling: none.
- Tracker update requirements: record files inspected, commands run, findings,
  skipped checks, and next actions.

### WS-01 Characterization Inventory

- Scope and covered gaps: confirm caller usage, schema references, generated
  metadata references, fixture-policy completeness, `package_reset` usage,
  smoke assertions, and adjacent consumers before deleting code.
- Dependencies: WS-00.
- Sequencing: complete immediately before WS-02 through WS-06.
- Risks: deleting a private surface with an undiscovered caller could break a
  public wrapper indirectly.
- Exit criteria: tracker records current inventories and any newly discovered
  blockers.
- Required validation: targeted `rg`/Node inventories; `make phase-map-check`
  when phase-map behavior is in scope.
- Generated-artifact handling: inspect generated metadata but do not hand-edit
  it.
- Tracker update requirements: append exact commands and summarized results.

### WS-02 Fixture Policy Simplification

- Scope and covered gaps: G-23 plus related smoke fixture updates.
- Dependencies: WS-01.
- Sequencing: run before or alongside WS-03 so dead fixture helpers can be
  removed cleanly.
- Risks: tests may still assert missing policy defaults; update those tests to
  assert explicit current contracts.
- Exit criteria: no implicit service-backed Postgres fixture defaults remain;
  current phase maps still pass validation.
- Required validation: G-23 validation criteria.
- Generated-artifact handling: regenerate only if owner inputs change.
- Tracker update requirements: record removed helpers, test updates, command
  results, and any retained explicit compatibility behavior.

### WS-03 Private CLI and Helper Pruning

- Scope and covered gaps: G-24.
- Dependencies: WS-01; coordinate with WS-02 for fixture helper exports.
- Sequencing: after caller inventory and after any fixture default cleanup that
  makes helper exports unused.
- Risks: a private command may be consumed by an undocumented local workflow.
  Public Make and harness wrappers are the compatibility boundary.
- Exit criteria: unused private commands and exclusive helper exports are gone;
  public wrappers continue to pass.
- Required validation: G-24 validation criteria.
- Generated-artifact handling: if any generated metadata references removed
  private paths, update owner inputs and regenerate through Make.
- Tracker update requirements: list deleted commands and the caller inventory
  that justified deletion.

### WS-04 Historical Schema Cleanup

- Scope and covered gaps: G-25.
- Dependencies: WS-01.
- Sequencing: after schema reference inventory and before final validation.
- Risks: docs or tools may still mention v3 as a retained diagnostic schema.
  Decide whether to delete or explicitly deprecate before editing.
- Exit criteria: old row-accounting schema support is removed or explicitly
  marked diagnostic-only, while v4 remains current.
- Required validation: G-25 validation criteria.
- Generated-artifact handling: preserve current v4 schema ID unless a
  deliberate schema revision is proposed.
- Tracker update requirements: record schema decision, references updated, and
  release-readiness behavior preserved.

### WS-05 Spec-First Future Growth Decision

- Scope and covered gaps: G-26.
- Dependencies: owner-doc decision.
- Sequencing: do not implement behavior changes until the NLSpec or other owner
  document explicitly changes the mapping contract.
- Risks: treating the bridge as private cleanup would break current browser and
  frontend phase behavior.
- Exit criteria: tracker clearly states whether the numeric bridge is preserved
  or queued for a future owner-spec change.
- Required validation: owner-doc review only if deferred; browser/frontend
  validation if implemented.
- Generated-artifact handling: update owner inputs and regenerate only after a
  spec-first decision.
- Tracker update requirements: record the decision and any deferred follow-up.

### WS-06 Smoke and Metadata Review

- Scope and covered gaps: G-27 plus generated metadata references to deleted,
  moved, or legacy paths.
- Dependencies: WS-01 and any implementation work from WS-02 through WS-04.
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

### WS-07 Final Handoff

- Scope and covered gaps: completion accounting for the simplification pass.
- Dependencies: all active workstreams that were attempted in the session.
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
| Browser/frontend phase mapping | `make phase-map-check`; relevant `make browser-e2e-*` target or targets; generated drift checks if owner inputs change |
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
