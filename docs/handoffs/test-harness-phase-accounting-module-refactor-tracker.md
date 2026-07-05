# test-harness-phase-accounting Next Refactor Tracker

## 1. Current Status and Authority

- Target label: `test-harness-phase-accounting`
- Target path: `tools/harness/phase-accounting`
- Tracker path: `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Status: next refactor iteration planned. The prior 2026-07-05
  remediation is treated as completed historical baseline, not current work.
- Tracker purpose: identify remaining specification, implementation, test,
  documentation, metadata, and generated-artifact gaps for the next structural
  refactor iteration.
- Public posture: preserve public harness contracts unless a future workstream
  is explicitly marked spec-first.
- Generated posture: generated files are downstream evidence and must not be
  hand-edited.

Authority and source hierarchy:

1. `docs/testing-harness-nlspec.md` owns harness mechanics, including command
   invocation, target selection, scheduling, fixture lifecycle, service
   ownership, artifact emission, cleanup, summary emission, failure mapping,
   and harness verification gates.
2. `docs/domain.md` owns domain vocabulary and concept boundaries only. Harness
   module names, helper files, validation targets, and tracker labels remain
   implementation-support terms unless an owner spec promotes them.
3. Core 00 through Core 04 own product behavior. Core 05 applies only to
   claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
4. `docs/design.md` owns frontend design direction and token definitions, but
   does not by itself establish Base Profile or extension-profile conformance.
5. Generated task-surface, topology, schedule, and Make artifacts are downstream
   of owner inputs.

## 2. Inspected Sources

This tracker refresh was based on inspection of:

- Historical tracker:
  `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Harness and vocabulary owners:
  `docs/testing-harness-nlspec.md`, `docs/domain.md`
- Frontend implementation-support guides:
  `docs/guides/cartulary_frontend_implementation_testing_guide.md`,
  `docs/guides/cartulary_visual_golden_maintenance.md`
- Phase-accounting implementation:
  `tools/harness/phase-accounting/**`
- Adjacent harness ownership boundaries:
  `tools/harness/scheduler/**`, `tools/harness/backend/**`,
  `tools/harness/diagnostics/**`, `tools/harness/output/test-output/**`,
  and `tools/harness/tests/**`
- Frontend owner data:
  `tools/frontend_phase_registry.json`,
  `tools/frontend_phase_maps/*.json`,
  `tools/frontend_visual_fixture_registry.json`
- Schemas:
  `tools/schemas/cartulary.phase_slice_plan.v1.schema.json`,
  `tools/schemas/cartulary.frontend_phase_registry.v3.schema.json`,
  `tools/schemas/cartulary.frontend_phase_test_map.v3.schema.json`,
  `tools/schemas/cartulary.frontend_row_accounting.v3.schema.json`,
  `tools/schemas/cartulary.frontend_visual_fixture_registry.v3.schema.json`,
  and `tools/schemas/cartulary.frontend_evidence_audit_summary.v1.schema.json`
- Generated-artifact and execution metadata:
  `tools/generated_artifact_policy.json`,
  `tools/task_surface_manifest.json`,
  `tools/task_surface.generated.mk`,
  `tools/execution_topology_manifest.json`,
  `tools/execution_topology_render_index.json`,
  `tools/scheduler_manifest.json.generated`,
  and `tools/browser_e2e_batch_manifest.json.generated`
- Public validation targets discovered through `make help-all`,
  `make task-guide ROLE=phase-author PHASE=phase4`, and
  `make explain-target TARGET=phase-slice DETAIL=summary`

## 3. Completed Historical Baseline

The prior tracker recorded a completed remediation pass. Carry this forward as
baseline unless a current validation target proves regression:

- Runtime execution for `phase-slice` and `service-backed-slice` was moved out
  of the phase-accounting CLI and into scheduler-owned execution helpers.
- Frontend row accounting was routed through normalized test-output
  observations rather than re-exporting test-output indexes from a frontend
  phase-accounting facade.
- Target-plan smoke coverage was split into owner-aligned diagnostics, backend,
  and phase-accounting wrappers behind the unchanged aggregate target.
- Retained-root `frontend-evidence-audit` validation was completed against
  broad, support, visual, and accessibility retained roots in the prior run.
- A v3 frontend phase registry and v3 visual fixture registry were introduced
  as append-only contiguous owner-data contracts.
- `frontend-phase-manifest.mjs` and `phase-manifest.mjs` were narrowed into
  compatibility facades with CLI or owner-local helper movement behind them.
- Generated metadata changes from that remediation were refreshed through
  Make-owned generation and drift checks.

Do not repeat that completed work as an open workstream. Re-open a historical
item only when a current owner spec changes or a current validation target fails.

## 4. Current Findings

- `tools/harness/phase-accounting/frontend/phase-manifest-core.mjs` remains a
  large cross-concern module even after facade narrowing. It still owns registry
  parsing, map validation, owner-ref validation, target-ref validation, visual
  fixture checks, freshness digests, guide restatement checks, ledger rendering,
  scenario grep, and CLI dispatch support.
- The frontend fixture path still has growth coupling. At least one helper uses
  a hard-coded `FE-VFIX-01` through `FE-VFIX-21` pattern even though the v3
  registry is append-only and should permit future contiguous `FE-VFIX-22+`
  owner data.
- Some diagnostics and guide language still present current counts such as
  `FE-P0` through `FE-P11`, `FE-VFIX-01` through `FE-VFIX-21`, or "all 12
  frontend phases" as if they were permanent limits.
- Frontend row selection and selected-row scope behavior are duplicated between
  frontend phase-slice planning and frontend row accounting.
- `frontend-evidence-audit-cli.mjs` mixes public CLI/env handling with reusable
  audit logic, retained-root resolution, digest checks, row-target audit, and
  summary writing.
- `phase-slice-plan.mjs` still constructs backend and browser work units while
  coordinating backend shard planning, browser batch/topology metadata,
  scheduler resources, diagnostics guidance, and plan serialization.
- `phase-manifest-cli.mjs` remains a large command switch and has minor
  duplicate branch noise. The compatibility facade is improved, but command
  handling is still harder to extend than necessary.
- Import-boundary metadata does not yet strongly protect future private
  phase-accounting internals from non-owner imports after the next split.
- `cartulary.phase_slice_plan.v1` leaves individual work-unit records open with
  `additionalProperties: true`. Tightening this could improve command-shape
  closure, but it is a public contract decision and must be spec-first.
- Frontend phase IDs currently map numerically to base `phaseN` selection in
  adjacent browser/base helpers. That coupling must be understood before adding
  phases whose frontend and base numbering diverge.

## 5. Compatibility Freeze

Future implementation sessions must preserve these public harness contracts
unless a workstream is explicitly marked spec-first:

- Make target names, especially `phase-slice`, `service-backed-slice`,
  `frontend-evidence-audit`, `phase-ledgers`, `phase-ledger-drift`,
  `phase-schedules`, `phase-schedule-drift`, `phase-map-check`,
  `phase-test-name-check`, `harness-contract`, and public smoke/report targets.
- Stable `command_id` values, including
  `cartulary.harness.command.phase_slice.v1`,
  `cartulary.harness.command.service_backed_slice.v1`, and
  `cartulary.harness.command.frontend_evidence_audit.v1`.
- Schema IDs, including `cartulary.phase_slice_plan.v1`,
  `cartulary.frontend_phase_registry.v3`,
  `cartulary.frontend_phase_test_map.v3`,
  `cartulary.frontend_row_accounting.v3`,
  `cartulary.frontend_visual_fixture_registry.v3`, and
  `cartulary.frontend_evidence_audit_summary.v1`.
- Retained artifact paths, including `frontend-row-accounting.json`,
  `frontend-evidence-audit-summary.json`, scheduler events and summaries,
  tool-run summaries, target summaries, generated ledgers, and phase-slice plan
  output.
- Public inputs, including `PHASE`, `PHASE_NAMESPACE`, `ROWS`, `JSON`,
  `CHECK_RESULTS_DIR`, `BROWSER_SUPPORT_RESULTS_DIR`,
  `BROWSER_VISUAL_RESULTS_DIR`, `BROWSER_A11Y_RESULTS_DIR`,
  `BROWSER_A11Y_PREFLIGHT_RESULTS_DIR`, and
  `BROWSER_MEASUREMENT_RESULTS_DIR`.
- Failure class/reason mapping, cleanup behavior, retained artifact retention,
  service ownership, runtime reset behavior, and scheduler resource semantics.

## 6. Out-of-Scope Boundaries

- Product HTTP routes, WebSocket behavior, workbook behavior, saved-view
  behavior, storage behavior, revision/history semantics, authentication,
  authorization, release publication behavior, and domain model changes.
- SQL migrations, DB queries, generated product contracts, dependency locks,
  `go.sum`, `pnpm-lock.yaml`, and tool-managed dependency/install artifacts.
- Hand-editing generated roots or generated outputs, including
  `internal/gen/**`, `packages/protocol-ts/src/generated/**`,
  `packages/ui-contracts/src/generated/**`,
  `tools/task_surface.generated.mk`,
  `tools/scheduler_manifest.json.generated`,
  `tools/browser_e2e_batch_manifest.json.generated`, and
  `tools/execution_topology_render_index.json`.
- Preserving private helper paths solely because they existed historically.
- Broad refactors outside phase-accounting/harness ownership unless required by
  owner metadata or import-boundary enforcement.
- Treating visual, accessibility, design-direction, or implementation-support
  evidence as product conformance without a Core 05 or product-owner boundary.

## 7. Detailed Gap Records

### G-01 Frontend Phase-Manifest Core Decomposition

- Remediation: split `frontend/phase-manifest-core.mjs` into owner-local modules
  for registry loading, phase-map validation, owner/root refs, target refs and
  audit routing, visual fixture validation, freshness digests, ledger rendering,
  scenario grep, and CLI support. Keep public facade exports and direct CLI
  behavior stable.
- Affected areas: implementation, tests, import-boundary metadata, and generated
  topology/task-surface metadata only if backing script paths move.
- Rationale: the current core still combines unrelated concerns, which makes
  future phase or fixture growth require high-risk edits in one large file.
- Expected long-term benefit: smaller modules with clearer ownership, narrower
  tests, easier review, and less risk when adding future frontend phases,
  visual fixtures, ledger behavior, or audit-route validation.
- Compatibility or migration impact: no public Make target, command output,
  schema ID, retained path, or exported facade name may change. Private import
  callers should move to owner facades or new owner-local modules intentionally.
- Risks of leaving unresolved: new frontend growth work will keep accumulating
  in a catch-all module, increasing duplicate validation logic and accidental
  public behavior drift.
- Validation criteria: `node --check` for changed modules,
  `make phase-map-check`, `make json-shape-check`, `make harness-contract`,
  `make lint-scripts`, and generated drift checks if metadata changes.

### G-02 Frontend Phase and Fixture Growth Guardrails

- Remediation: replace hard-coded fixture and phase count assumptions with
  registry-derived checks. Specifically, remove remaining fixed
  `FE-VFIX-01..FE-VFIX-21` validation from base phase fixture helpers, make
  unknown-frontend-phase diagnostics derive from the declared registry, and
  update guides that describe current counts as permanent limits.
- Affected areas: specification text if behavior changes, implementation,
  documentation, fixtures/tests, frontend owner data validation, and generated
  ledgers/schedules only when owner inputs change.
- Rationale: future phase growth is a core design constraint. The v3 registries
  already express append-only contiguous growth, but stale hard-coded ranges can
  still block future `FE-P12+` or `FE-VFIX-22+` data.
- Expected long-term benefit: adding future phases and fixtures becomes owner
  data plus validation evidence, not a hunt for scattered range constants.
- Compatibility or migration impact: current `FE-P0..FE-P11` and
  `FE-VFIX-01..FE-VFIX-21` data must continue to pass unchanged. Accepting new
  IDs requires owner data, maps, freshness digests, ledgers, and validation
  evidence; it must not silently infer missing maps or artifacts.
- Risks of leaving unresolved: the next frontend phase or fixture addition will
  require tactical patches across docs and validators, increasing drift between
  schemas, implementation, and guides.
- Validation criteria: add or retain tests with synthetic next IDs where useful,
  then run `node --check` for changed modules, `make phase-map-check`,
  `make json-shape-check`, `make phase-ledger-drift`,
  `make phase-schedule-drift`, and `make generated-artifact-policy-check`.

### G-03 Shared Frontend Row Scope and Selection Library

- Remediation: extract selected-row parsing, phase-number parsing, active versus
  planned row inclusion, through-phase selection, and frontend row-ID validation
  into a shared owner-local module used by frontend phase-slice planning and
  frontend row accounting.
- Affected areas: implementation, unit/smoke tests, row-accounting retained
  output, and frontend namespace phase-slice planning.
- Rationale: duplicated row selection logic risks subtle divergence between the
  plan emitted before execution and the accounting/audit pass that evaluates
  retained evidence afterward.
- Expected long-term benefit: one definition of selected frontend row scope,
  easier testing for future phase growth, and simpler reasoning about planned,
  active, stale, blocked, retired, and explicitly selected rows.
- Compatibility or migration impact: preserve `cartulary.frontend_row_accounting.v3`,
  `frontend-row-accounting.json`, selected `ROWS` behavior, direct frontend
  namespace phase-slice behavior, and target-summary failure mapping.
- Risks of leaving unresolved: future changes can make planner rows and
  accounting rows disagree, producing false audit failures or missed evidence.
- Validation criteria: `node --check` for changed modules,
  `make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`,
  `make service-backed-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`,
  `make run-harness-smoke-extended`, `make harness-contract`, and any focused
  row-accounting tests.

### G-04 Frontend Evidence Audit CLI/Library Split

- Remediation: move pure audit behavior from `frontend-evidence-audit-cli.mjs`
  into an owner-local audit module. Keep the CLI as the public env/input,
  argument, retained-root, summary-writing, and exit-code wrapper.
- Affected areas: implementation, tests, task-surface metadata only if backing
  script paths change, and retained audit fixtures.
- Rationale: audit logic is useful to tests and validators, while CLI/env
  handling is public command plumbing. Combining them makes behavior harder to
  test without invoking the public target shape.
- Expected long-term benefit: clearer tests for audit behavior, less risk when
  adding retained-root routes, and a thinner public command wrapper.
- Compatibility or migration impact: preserve `frontend-evidence-audit`, all
  supported retained-root inputs, `cartulary.frontend_evidence_audit_summary.v1`,
  retained summary path, digest checks, failure classes, and cleanup behavior.
- Risks of leaving unresolved: future audit-route changes will keep touching CLI
  plumbing and may accidentally change public input or failure behavior.
- Validation criteria: `node --check` for changed modules,
  `make frontend-evidence-audit` with Make-owned retained roots when available,
  `make run-harness-smoke-extended`, `make harness-contract`, and
  `make lint-scripts`.

### G-05 Phase-Slice Planner Work-Unit Decomposition

- Remediation: extract backend work-unit construction and browser work-unit
  construction into owner-local planning helpers behind the existing
  `phase-slice-plan.mjs` facade. Keep row selection, resource claims, ordering,
  dependency edges, service-wrapper semantics, and JSON serialization stable.
- Affected areas: implementation, planner tests, scheduler resource semantics,
  backend/browser helper imports, and generated metadata only if backing script
  paths move.
- Rationale: the current planner still coordinates multiple owner domains in one
  module, which makes future runner or phase namespace additions brittle.
- Expected long-term benefit: phase-slice planning becomes easier to extend and
  test by runner family without weakening scheduler ownership.
- Compatibility or migration impact: preserve `cartulary.phase_slice_plan.v1`,
  target summaries, emitted work-unit order, scheduler resources, service
  ownership, cleanup/reset behavior, and public Make target behavior.
- Risks of leaving unresolved: backend, browser, frontend, and scheduler changes
  will continue to collide in one planner module, making future expansion harder
  to review and validate.
- Validation criteria: `node --check` for changed modules,
  `make phase-slice PHASE=phase4 JSON=1`,
  `make service-backed-slice PHASE=phase4 JSON=1`,
  `make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`,
  `make service-backed-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`,
  `make run-harness-smoke-extended`, `make harness-contract`, and drift checks
  when metadata changes.

### G-06 Phase-Manifest CLI Command Dispatch Cleanup

- Remediation: replace the large command switch in `phase-manifest-cli.mjs` with
  a small command table or grouped handler modules, and remove duplicate branch
  noise. Keep command names, arguments, stdout/stderr shape, and exit behavior
  unchanged.
- Affected areas: implementation, CLI tests/smokes, docs only if command help
  text changes, and generated metadata only if backing script paths move.
- Rationale: the facade was narrowed, but command dispatch remains difficult to
  extend and review.
- Expected long-term benefit: adding future validation or reporting commands
  becomes a local command registration change instead of another branch in a
  large switch.
- Compatibility or migration impact: no public command behavior change. Existing
  scripts and Make targets that invoke `phase-manifest.mjs` or
  `phase-manifest-cli.mjs` must continue to work.
- Risks of leaving unresolved: command handling will continue to collect
  tactical branches and increase the chance of copy/paste defects.
- Validation criteria: `node --check` for changed modules,
  direct smoke of representative CLI commands such as `list-phases`,
  `make phase-test-name-check`, `make phase-map-check`,
  `make harness-contract`, and `make lint-scripts`.

### G-07 Phase-Accounting Import-Boundary Enforcement

- Remediation: after private module boundaries are split, update harness import
  boundary metadata/tests so non-owner modules import declared phase-accounting
  facades rather than private internals. Allow owner-local imports inside the
  phase-accounting subtree.
- Affected areas: static-analysis metadata, tests, implementation imports, and
  task-surface/topology metadata only when public backing paths move.
- Rationale: decomposition only helps if new private internals do not become
  accidental cross-owner dependencies.
- Expected long-term benefit: clearer module ownership, safer future movement,
  and less compatibility burden for private helper paths.
- Compatibility or migration impact: public facades remain stable. Non-owner
  private imports may need mechanical migration to declared facades.
- Risks of leaving unresolved: newly split modules can immediately become
  de-facto public APIs, making future cleanup harder.
- Validation criteria: import-boundary test coverage for allowed facades and
  rejected private paths, `make harness-contract`, `make lint-scripts`, and
  `node --check` for changed static-analysis scripts.

### G-08 Phase-Slice Work-Unit Schema Closure Decision

- Remediation: run a spec-first decision on whether
  `cartulary.phase_slice_plan.v1` should keep open work-unit objects, tighten
  same-ID validation through implementation only, or introduce a future schema
  revision. Do not silently tighten the public schema without NLSpec alignment.
- Affected areas: specification, schema, validator implementation, generated
  metadata, scheduler consumers, and tests.
- Rationale: NLSpec command-shape closure favors precise public contracts, while
  the current schema leaves `work_unit` open for extensibility.
- Expected long-term benefit: explicit contract posture for planner output,
  fewer accidental scheduler fields, and a clear migration path if a v2 schema
  is needed.
- Compatibility or migration impact: any schema ID or wire-shape change is a
  public contract change and must be spec-first. Same-ID validator tightening
  may still affect callers if they emit extra fields.
- Risks of leaving unresolved: work-unit shape can drift under a stable schema
  ID, weakening scheduler contract evidence.
- Validation criteria: accepted NLSpec decision, schema validation through
  `make json-shape-check`, planner JSON checks, `make harness-contract`, and
  drift checks if schema or metadata owners change.

### G-09 Frontend/Base Phase Number Coupling Audit

- Remediation: audit all helpers that convert `FE-P<N>` to `phase<N>` or
  otherwise assume frontend and base phase numbers advance together. Document
  the intended behavior before accepting frontend phases that diverge from base
  phase numbering.
- Affected areas: frontend phase selection, browser duration accounting, base
  phase maps, frontend phase maps, planner behavior, and documentation.
- Rationale: current data declares `FE-P0` through `FE-P11`, matching the
  current base phase range. Future frontend-only growth could expose hidden
  coupling.
- Expected long-term benefit: future frontend phases can be added deliberately
  without breaking base Playwright selection, browser duration accounting, or
  generated schedules.
- Compatibility or migration impact: no current behavior change is required.
  Any decoupling must preserve existing `FE-P0..FE-P11` selection behavior and
  public target inputs.
- Risks of leaving unresolved: adding `FE-P12+` can fail late in browser/base
  helper paths even if frontend registry validation passes.
- Validation criteria: audit notes in this tracker or successor handoff,
  targeted tests for divergent synthetic or fixture data if introduced,
  frontend namespace phase-slice JSON checks, `make phase-map-check`, and
  `make harness-contract`.

## 8. Workstream Sequencing

| Workstream | Depends on | Class | Main changes | Risks | Exit criteria | Required validation |
| --- | --- | --- | --- | --- | --- | --- |
| WS-00 Tracker refresh | none | docs | Replace completed-work tracker with this forward-looking handoff and record lint result. | Documentation drift only. | Required sections, gap records, generated-artifact handling, and session log are current. | `make lint-markdown` |
| WS-01 Characterization baseline | WS-00 | validation | Capture current behavior before implementation movement. | Existing unrelated harness failure may obscure refactor risk. | Baseline command results and run roots are recorded before code edits. | `make phase-map-check`; `make phase-test-name-check`; `make harness-contract`; targeted phase-slice JSON checks as needed |
| WS-02 Future-growth guardrails | WS-01 | implementation/spec if public behavior changes | Address G-02 before more splits. | Accidentally changing accepted current data or public diagnostics. | Current data passes and future IDs are blocked only by missing owner data/evidence. | `node --check` for changed modules; `make phase-map-check`; `make json-shape-check`; drift checks if owner inputs change |
| WS-03 Frontend manifest decomposition | WS-01, WS-02 when growth validators move | implementation | Address G-01 behind stable facades. | Export, CLI, or task-surface backing-path drift. | Core concerns are split and facade callers still pass. | `node --check`; `make phase-map-check`; `make json-shape-check`; `make harness-contract`; `make lint-scripts` |
| WS-04 Shared row scope and audit library | WS-03 | implementation | Address G-03 and G-04. | Retained evidence audit or selected-row semantics drift. | Planner, row accounting, and audit use shared owner-local logic where appropriate. | Frontend phase-slice JSON checks; `make run-harness-smoke-extended`; `make harness-contract`; audit retained-root checks when roots are available |
| WS-05 Phase-slice planner decomposition | WS-01, WS-03 | implementation | Address G-05 and include G-09 audit where planner behavior depends on phase numbering. | Work-unit order, resources, service ownership, or cleanup changes. | Base and frontend JSON plans remain compatible and scheduler semantics are preserved. | Base/frontend `phase-slice` and `service-backed-slice` JSON checks; `make run-harness-smoke-extended`; `make harness-contract` |
| WS-06 CLI and import-boundary cleanup | WS-03, WS-05 | implementation | Address G-06 and G-07 after private boundaries exist. | Breaking private callers before facade migration. | Command behavior is unchanged and private imports are enforced. | `node --check`; `make phase-test-name-check`; `make phase-map-check`; `make harness-contract`; `make lint-scripts` |
| WS-07 Spec-first plan contract decision | WS-01 | spec-first | Resolve G-08 before schema tightening. | Public schema or scheduler shape changes without owner approval. | Decision recorded: keep v1 open, tighten validators, or define a versioned migration. | `make json-shape-check`; planner JSON checks; `make harness-contract`; drift checks if owners change |
| WS-08 Generated metadata refresh | Any workstream with owner input or backing-path changes | generated | Update owner inputs first, then regenerate downstream artifacts through Make. | Hand-editing generated files or leaving stale backing paths. | Generated outputs match owner inputs and drift checks pass. | `make phase-schedules`; `make phase-schedule-drift`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` |
| WS-09 Final handoff | Completed implementation workstreams | handoff | Record files changed, validations, skipped checks, run roots, and blockers. | Incomplete context for the next agent. | Tracker/successor handoff is current and binary criteria are checked. | `make agent-finalize`; broader `make check` only when implementation breadth warrants it |

## 9. Generated-Artifact Handling Rules

- Do not hand-edit generated files or generated roots.
- Generated roots from policy include `internal/gen/**`,
  `packages/protocol-ts/src/generated/**`, and
  `packages/ui-contracts/src/generated/**`.
- Generated harness/topology outputs include `tools/task_surface.generated.mk`,
  `tools/scheduler_manifest.json.generated`,
  `tools/browser_e2e_batch_manifest.json.generated`, and
  `tools/execution_topology_render_index.json`.
- Update owner inputs before generated outputs. Path or backing-script changes
  usually start in `tools/execution_topology_manifest.json` or other
  task-surface/topology owner inputs.
- Run Make-owned generation after owner input changes. Use `make phase-schedules`
  for schedule/topology/task-surface refreshes.
- Verify with `make phase-schedule-drift`, `make generate-drift`,
  `make generated-artifact-policy-check`, and `make json-shape-check`.
- If generated metadata still references an old helper path after a move, the
  workstream is incomplete.

## 10. Validation Matrix

| Change type | Required validation |
| --- | --- |
| Tracker-only docs update | `make lint-markdown` |
| Harness docs, schema, or metadata change | `make json-shape-check`; `make generated-artifact-policy-check`; `make harness-contract` |
| Phase map or registry change | `make phase-map-check`; `make phase-test-name-check`; `make phase-ledger-drift`; `make phase-schedule-drift` |
| Planner behavior change | `make phase-slice PHASE=phase4 JSON=1`; `make service-backed-slice PHASE=phase4 JSON=1`; `make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1`; `make service-backed-slice PHASE_NAMESPACE=frontend PHASE=FE-P3 JSON=1` |
| Script or module refactor | `node --check` for changed modules; `make lint-scripts`; `make run-harness-smoke-extended` |
| Frontend retained-evidence audit behavior | `make frontend-evidence-audit` with Make-owned retained roots when available; `make run-harness-smoke-extended`; `make harness-contract` |
| Owner input or backing-script path movement | `make phase-schedules`; `make phase-schedule-drift`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` |
| End-of-run handoff | `make agent-finalize`; broader `make check` only when breadth or risk warrants it |

## 11. Handoff Update Requirements

Every future implementation session using this tracker must update this file or a
successor handoff with:

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

- WS-00 must record the markdown validation result in the session handoff log.
- WS-01 must record baseline command results before code movement.
- WS-02 through WS-07 must mark each touched gap as open, in progress,
  completed, or deferred with validation evidence.
- WS-08 must list owner inputs and generated outputs touched.
- WS-09 must update open risks, binary criteria, and next recommended action.

## 12. Open Risks and Blockers

| ID | Risk or blocker | Why it matters | Resolution path | Status |
| --- | --- | --- | --- | --- |
| OR-001 | Catch-all frontend manifest core remains a high-change module. | Future growth still concentrates unrelated validation concerns. | Complete G-01 through WS-03. | open |
| OR-002 | Remaining hard-coded phase/fixture counts can block future owner data. | `FE-P12+` or `FE-VFIX-22+` should not require tactical range patches. | Complete G-02 through WS-02. | open |
| OR-003 | Row selection duplication can produce planner/accounting drift. | A plan can select a different row set than retained evidence accounting. | Complete G-03 through WS-04. | open |
| OR-004 | Evidence audit logic is coupled to CLI plumbing. | Public input/failure behavior is easier to change accidentally. | Complete G-04 through WS-04. | open |
| OR-005 | Phase-slice planner still coordinates backend/browser/scheduler concerns. | Runner growth and resource policy changes remain brittle. | Complete G-05 through WS-05. | open |
| OR-006 | Phase-slice work-unit schema openness is unresolved. | Public plan shape can drift under a stable schema ID. | Make the G-08 decision through WS-07 before tightening. | open, spec-first |
| OR-007 | Frontend/base phase-number coupling is not fully documented. | Future frontend-only growth may pass registry checks but fail browser/base paths. | Complete G-09 audit before accepting divergent phase data. | open |
| OR-008 | Direct retained-root audit depends on fresh Make-owned browser roots. | Synthetic smoke cannot prove all real retained-root combinations. | Produce or supply retained roots before claiming live browser audit closure. | standing validation requirement |

## 13. Binary Completion Criteria

This tracker update is complete only when all of the following are true:

- Prior remediation is summarized as completed history and no longer presented
  as current work.
- Current public compatibility surfaces are explicitly frozen.
- Out-of-scope product, dependency, and generated-file boundaries are explicit.
- Remaining specification, implementation, test, documentation, metadata, and
  generated-artifact gaps are identified.
- Each gap record includes remediation, affected areas, rationale, expected
  long-term benefit, compatibility or migration impact, risks of leaving it
  unresolved, and validation criteria.
- Workstreams include dependencies, risks, required tracker updates,
  generated-artifact handling, validation commands, and exit criteria.
- Open risks and blockers are listed with resolution paths.
- Validation matrix covers tracker-only, schema/metadata, phase map, planner,
  script/module, retained-audit, generated-artifact, and finalization changes.
- The session handoff log records the `make lint-markdown` result for this
  tracker refresh.

Future implementation slices are complete only when their gap records are marked
completed or intentionally deferred, their public compatibility impact is
recorded, generated-artifact rules were followed, and their validation commands
passed or failures are classified with run roots.

## 14. Session Handoff Log

| Time | Session | Files changed | Commands run | Result | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-07-05T10:37:26-04:00..2026-07-05T10:40:14-04:00 | Codex GPT-5 tracker refresh | `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md` | `git status --short`; inspection of prior tracker, harness/domain docs, frontend guides, phase-accounting implementation notes, schemas, generated-artifact policy, task-surface/topology metadata, and public validation targets; `make lint-markdown`. | Tracker rewritten as a forward-looking next-iteration handoff; starting worktree was clean; `make lint-markdown` passed with no output. | Use WS-01 characterization baseline before any implementation refactor. |
