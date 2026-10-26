# Testing Remediation 01: Graph and Scanner Assurance

## 1. Scope and Source Posture

This is the controlling execution tracker for the approved next iteration. The
user authorized implementation on 2026-10-26. Preserve the preceding
[remediation handoff](testing-remediation.md), its completed slices, and its
separate open WF-33 effort. Update this tracker after each workstream and before
beginning the next. Finish with WF-05 validation and handoff completion.

Repository: `/home/jochi/code/cartulary`; canonical target `.`; bounded seam:
graph/planning assurance and Go scanner wrapper assurance. Starting branch is
`main`, HEAD `3d192eea610867bac59f4eaf79ba54f4dd05dd68`, with the preceding
iteration's uncommitted changes already present. The new tracker was absent.
No commit, push, or destructive cleanup is requested.

[Core 00](../spec/00_document_set_status_and_precedence.md) governs precedence;
the adopted [Testing Harness NLSpec](../testing-harness-nlspec.md) owns harness
mechanics. Its REQ-001, REQ-003/004, REQ-082, REQ-800/805/806, Sections 8–10 and
AC-040 govern the affected boundaries. This handoff sequences work; tests and
machine contracts are downstream projections, not requirements authorities.
Executable consumers must never read, stat, hash, or depend on Markdown.

Preserve production features, security thresholds, scanner rule/package coverage,
public commands, artifact schemas, and scheduler semantics. The separately
approved behavioral correction is rejecting failed Staticcheck package discovery
before scanning. Do not introduce a runner, registry, scheduler, compatibility
mode, legacy alias, benchmark campaign, or test-only production API.

## 2. Current-State Repository Inventory

Planning enumerated 7,633 tracked/non-ignored paths, including four already-deleted
files. Enumeration is not a behavioral audit. Primary subjects below were read;
other repository paths remain metadata-only or outside this bounded iteration.
Ignored builds, caches, service data, vendor dependencies and retained results are
not implementation inputs to change. Authored owner inputs precede Make-generated
task-surface/topology outputs; generated outputs must not be edited by hand.

| Area | Inspected responsibility and callers | Disposition / owner |
| --- | --- | --- |
| Graph contracts, tier selection, command selections, composition | Policy and repository wiring, reached by graph assurance | Split responsibilities under work-graph owner |
| `test-work-graph-cutover.mjs` | Identical fast/matrix bodies; grouping, replay, resource and artifact checks | Replace with named owner-local tests |
| Canonical target-plan tests | Planning/execution/replay equivalence, CLI and purity; duplicate regression entry | Keep diagnostics ownership, one registration |
| Four `test-run-go-*.sh` suites and scanner wrappers | Repeated shell fixtures and string assertions | Consolidate static-analysis assurance |
| Command-failure and output diagnostic tests | Make/private-channel transport and summary publication | Keep distinct witnesses; not duplicate wrapper checks |
| Task surface, work graph, catalog and verification inputs | Public selection and canonical operation binding | Update authored inputs and regenerate projections |
| Product/backend/browser implementation | Not proposed for movement | Excluded; no product behavior or fixture redesign |

## 3. Module Boundary Diagnosis

Work-graph tests own selection, composition, grouping, dependency/capacity policy
and identity. Diagnostics tests own input admission, rendering, CLI and replay
equivalence. Execution tests own published-unit admission. Static-analysis tests
own wrapper invocation and artifacts; runtime/output tests retain failure-channel
and publication coverage. Shared private helpers construct fresh data, scratch
and bounded processes; they must not compute expected policy or become a DSL.

## 4. Public Contract and Behavior Freeze Map

| ID | Required guarantee | Observation / disposition |
| --- | --- | --- |
| C-001 | Exact selection, admission before preparation, tiers, dependency closure, deduplication | Independent policy expectations plus real repository wiring; preserve |
| C-002 | Planning/execution/replay equivalence, determinism, capacity and isolation | Separate paths under equal input/capacity; preserve |
| C-003 | Planning purity and rejection of forged/unsafe published selection | Isolated side-effect checks and explicit valid/mutated artifacts; preserve |
| C-004 | Authored package coverage, generated exclusions, targeted blocking, advisory findings and truthful tool failure | Actual wrapper subprocesses; preserve, separately correct discovery failure |
| C-005 | Govulncheck classifications, private artifacts, diagnostic transport, cleanup and publication | Wrapper plus Make/runtime/output boundaries; preserve |

Private paths, incidental log wording, historical row counts and duplicate modes
are replaceable. Independently reviewed security and gate-membership expectations
remain necessary when they detect incorrect authored configuration.

## 5. Coupling and Boundary Findings

| ID | Remediation / areas | Rationale and benefit | Migration / unresolved risk | Acceptance |
| --- | --- | --- | --- | --- |
| F-001 | One canonical graph assurance operation; remove duplicate fast/matrix/target-plan paths. Spec, tests, routing, docs. | Prevent repeated work and divergent inventories. | Retire private identities without aliases; otherwise duplicate execution persists. | One canonical unit per equivalent selection; successor coverage mapped. |
| F-002 | Small fixtures for existing pure policy seams; real unmodified inputs for wiring. Tests. | Distinguish policy defects from configuration defects. | No public API change; otherwise test replacement of configuration masks drift. | Independent expectations, fresh mutable contexts. |
| F-003 | Explicit schema-valid published-run fixtures and owned scratch. Tests. | Remove dependence on enclosing run and case order. | Private setup only; otherwise parent mutation and misleading failures remain. | Positive control plus each identity/corruption/symlink rejection. |
| F-004 | Remove migration counts, stale narration, dead loops and unused scaffolding. Tests/docs/spec as needed. | Permit legitimate suite growth without irrelevant failures. | No compatibility promise for incidental counts; retain exact selected-row closure. | Every retirement is classified, with meaningful coverage retained. |
| F-005 | Four named Node wrapper modules, shared local process fixtures and parsed artifacts; one canonical scanner operation. Tests/routing/docs. | Less boilerplate and more precise assertions. | Delete four shell test entrypoints, no aliases; otherwise substring checks miss malformed/incorrect data. | Unique obligations covered; distinct channel/output witnesses remain. |
| F-006 | Explicit discovery status before filtering/scanning. Implementation/tests. | Prevent incomplete analysis from appearing successful. | Previously masked Staticcheck discovery failures now fail. | Partial-output discovery failure never invokes scanner. |

All findings are `must_fix` within the named slices, not authority for unrelated
cleanup. Each assertion group is retained, merged into a named successor, retired
as incidental, or assigned to the separate normative correction.

## 6. Refactor Workstreams

The prerequisite column is authoritative. Workstreams execute sequentially.

| ID / slice | Prerequisite | Goal and risk | Binary exit criterion |
| --- | --- | --- | --- |
| WF-00 / S-000 | None | Owner cleanup, coverage map and baseline; avoid treating history as requirements | Owner-aligned scope and baseline recorded |
| WF-01 / S-001 | WF-00 | Policy/fixture organization; avoid self-confirming expectations | Focused graph assurance passes with fresh cases |
| WF-02 / S-002 | WF-01 | Graph integration/routing consolidation; preserve distinct contexts | One registration, complete coverage, generated projections agree |
| WF-03 / S-003 | WF-02 | Scanner fixture and routing consolidation; retain transport coverage | All wrapper obligations pass and old paths retired |
| WF-04 / S-004 | WF-03 | Staticcheck fail-closed discovery; preserve successful scan selection | Negative discovery characterization and actual tooling pass |
| WF-05 / S-005 | WF-04 | Finalization, compatible validation and handoff | Section 12 satisfied, evidence and exclusions recorded |

## 7. Proposed Refactor Slice Plan

S-000 cleans up REQ-082's duplicate graph modes while preserving assurance
responsibilities and security contexts. S-001 uses existing input/selection and
operation functions for small fixtures; real compiler mutations stay case-owned.
No shared mutable compiler cache or new production injection API is permitted.

### Assertion successor ledger

| Existing group | Classification / successor | Retirement limit |
| --- | --- | --- |
| Command selector, owner/suite, browser, package, lint/source wiring | C-001; real authored bindings in command-selection integration tests | Remove fixture rebinding, not independent expected selections |
| Selection admission and private union mechanics | C-001; small explicit policy fixtures | No repository compiler required for pure resolution |
| Graph schema, topology rejection, tier selection | C-001/C-002; graph contract tests | Remove unused roots/blocks; keep independent full-tier sentinels |
| Pure Go LPT/cohort arithmetic and argument transport | C-002/C-004; named pure cases | Keep timeout/oversized/isolation and literal-argument negative cases |
| Aggregate determinism, tier closure, blocking policy, projection closure | C-001/C-002; canonical aggregate planning matrix | Remove repeated full compilations after successor assertions are wired |
| Fast/matrix cutover body | C-001/C-002; named replay, preparation, grouping, resources and model cases | Modes are identical; neither is an independent witness |
| App.auth/measurement historical row counts | Incidental; selected membership and isolation remain | Remove exact historical population, not nonempty fixtures/closure |
| Borrowed run-manifest/execution-context selection check | C-003; explicit execution-owned published-unit fixture | No enclosing-run artifact dependency |
| Plan input, documentation exclusion, capacity, purity and CLI | C-001/C-003; diagnostics-owned tests | Keep independent CLI/purity boundary; bound subprocesses |
| Staticcheck shell cases | C-004; named wrapper tests | Keep roots, generated exclusions, checks override, env and tool status |
| Targeted/advisory Gosec shell cases | C-004/C-005; named wrapper tests | Preserve enforcing/advisory contexts and exact rules/coverage |
| Govulncheck shell invocation/artifact cases | C-004/C-005; named wrapper tests | Parse structured artifacts; retain private modes and database forwarding |
| Govulncheck Make/private-channel matrix | C-005; existing command-failure tests retained | Do not replace with direct-process status assertions |
| Govulncheck output summaries | C-005; existing output diagnostics retained | Distinct evidence consumer boundary remains |

S-002 retains one integration matrix for test-fast/check/ci/release-check with
independent planning, execution, replay and determinism paths. Immutable results
may be shared only within that target's group. Preserve exact/service/visual
selection, policy dependencies, tiers, grouping, resources, isolation and evidence
closure. Move published-unit checks to execution tests with explicit fixtures.
Keep `harness.harness_graph_contract` and its existing routing row; replace
duplicate operation references with that canonical identity. Broad assurance
remains full-tier; the three narrow fast-smoke roles remain unchanged.

S-003 directly registers four Node wrapper test modules with scanner-local
fixtures and existing external scratch. Capture argv/environment structurally,
bound subprocesses, isolate parent failure/artifact state, and clean up per case.
Introduce `harness.assurance_go_scanner_wrappers`, full-tier row
`harness.command_surface.behavior.go_scanner_wrappers`, and its verification.
Replace four private operations in regression/full, without fast-tier promotion.
Keep existing Make-wrapped Govulncheck failure and output-publication matrices.

S-004 tests discovery failure with no/partial output, empty successful discovery,
mixed authored/generated packages, and scanner failure; then replaces the unsafe
Staticcheck process-substitution discovery with checked capture and filtering.

S-005 removes stale callers, finalizes/generated-refreshes before broad validation,
freezes source and records evidence. Rollback is slice-scoped authored and generated
changes together; never broad-reset existing user work or restore compatibility
aliases. A failed required check blocks the dependent workstream until diagnosed.

## 8. Validation Plan

| ID | Canonical command / stage | Required outcome |
| --- | --- | --- |
| V-001 | `make help ALL=1`; `make task-guide OWNER=harness.test_catalog`; `make task-guide OWNER=harness.command_surface`; baseline `make harness-contract SUITE=regression` | Discover current bindings; characterize legacy union |
| V-002 | `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | Focused graph coverage after each graph slice |
| V-003 | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.go_scanner_wrappers` | Focused scanner coverage after registration |
| V-004 | `make generate`; `make source-check CHECK=generated`; `make source-check CHECK=schema` | Authored/generated consistency after routing changes |
| V-005 | `make lint CHECK=go` plus V-003 | Actual Staticcheck tooling and controlled failure cases |
| V-006 | `make agent-finalize`, then default `make harness-contract`, V-003, `make source-check CHECK=all`, `make lint CHECK=scripts`, `make lint CHECK=shell`, affected V-005 | Compatible final portfolio, including exact `harness.evidence_accounting.behavior.output_accounting` selection for distinct publication coverage; inspect finalizer changes first |
| V-007 | `make lint-markdown` | Documentation validity |

Use fresh explicit run IDs; record exact roots, outer exit, source identity,
canonical validity, cleanup and retained-secret scan. Diagnose failures before
reruns. Reassess source compatibility after repair; do not claim full-owner closure
from a subset. No blanket CI/browser/UI-review/benchmark run is required. Supply
finalizer RESULTS_DIR only for an existing compatible successful full warm check;
otherwise report retained maintenance skipped. Historical passes are not current
qualification.

## 9. Top-Level Work Tracker

| Workstream | Status | Evidence / next action |
| --- | --- | --- |
| WF-00 | DONE | E-002: baseline passed; owner cleanup and successor ledger recorded |
| WF-01 | DONE | E-003: policy, wiring and diagnostics split; focused assurance passed |
| WF-02 | DONE | E-004: canonical graph matrix, fixtures and routing passed |
| WF-03 | DONE | E-005: scanner, graph and source checks passed |
| WF-04 | DONE | E-006: negative characterization, scanner cases and actual Go lint passed |
| WF-05 | DONE | E-008/E-009: finalization, compatible validation and handoff complete |

## 10. Session Handoff Log

| Evidence | Observation / command / result | Applicability |
| --- | --- | --- |
| E-001 | 2026-10-26 implementation resumed on the unchanged planning HEAD; prior dirty work present. Root/tooling instructions and test-operations skill read. `make help ALL=1` passed; task guides started. | Implementation now authorized; no completed slice yet |
| E-002 | Both task guides passed. `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-baseline-20261026-01 make harness-contract SUITE=regression` passed, outer exit 0, 254010 ms, root `.cartulary/test-results/tr01-baseline-20261026-01`, source `sha256:aa432d498946013285d924d51e7f754a7d1bfc37a40492be0cb5c09e413ca9d0`. All 55 graph units passed. REQ-082 now describes canonical graph/scanner responsibilities and fail-closed discovery. | Pre-move baseline; no executable edits preceded the run. WF-00 complete before WF-01. |
| E-003 | WF-01 moved LPT and plan-input assertions to their owners, split graph registration from inert boundary registration, added small selection fixtures and owned scratch, and removed synthetic command rebinding. Focused run `tr01-wf01-graph-20261026-01` failed (outer 2): an old assertion omitted declared build defaults. Corrected the explicit expectation. Rerun `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf01-graph-20261026-02 make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` passed (outer 0, 107020 ms), source `sha256:d01d0ab8e092d8ad16635b1d6339efa5c9917c53b37743fc2d6ec5c72d33e294`; root `.cartulary/test-results/tr01-wf01-graph-20261026-02`. | All 60 tests passed; no production changes. WF-01 complete before WF-02. |
| E-004 | WF-02 retired `test-work-graph-cutover.mjs` and its fast/matrix operations plus the duplicate target-plan operation. Named compilation cases and execution-owned fixtures preserve unique assertions; aggregate policy/resource/tier checks consume the canonical planning/execution/replay matrix. Retired historical App.auth/measurement counts and narration. Generation first rejected an omitted tier check binding; adding the explicit canonical binding repaired it. `make generate` passed at `tr01-wf02-generate-20261026-02`. `make source-check CHECK=generated` and `CHECK=schema` passed at `tr01-wf02-generated-20261026-01` / `tr01-wf02-schema-20261026-01`. First graph run failed because synthetic row IDs violated the schema; positive controls caught the fixture defect. Corrected IDs; focused V-002 rerun `tr01-wf02-graph-20261026-02` passed (outer 0, 108567 ms, 74 tests), source `sha256:b265ac13694e8743ebe62a79113f24cf1af7561ed63617b5e3fd9ff2d4e04b1e`. | All roots under `.cartulary/test-results/`; graph cleanup and secret scans passed. Private names removed without aliases; public commands and distinct contexts preserved. WF-02 complete before WF-03. |
| E-005 | WF-03 replaced four `test-run-go-*.sh` suites with `test-go-{staticcheck,gosec-targeted,gosec-audit,govulncheck}.mjs` and `go-scanner-fixtures.mjs`. Bounded wrapper subprocesses record structured argv/environment, parse metadata/findings and verify private modes under ambient 022. Added one full-tier scanner row/verification and canonical operation; removed four private operations without aliases. Existing Make diagnostic/output matrices remain. `make generate` passed (`tr01-wf03-generate-20261026-01`). V-003 passed (`tr01-wf03-scanner-20261026-01`, 2246 ms, 16 tests); V-002 passed (`tr01-wf03-graph-20261026-01`, 108631 ms, 74 tests); `make source-check CHECK=all` passed (`tr01-wf03-source-20261026-01`, 19450 ms). All outer exits 0; source `sha256:3c89f898027fa3aea7a9162142d435fbe00c8849ce8854f7edb1b5bfd0206caa`. | Roots under `.cartulary/test-results/`; cleanup completed and secret scans passed. No scanner production behavior changed yet. WF-03 complete before WF-04. |
| E-006 | WF-04 added failed-without-output, failed-with-partial-output, empty-success and generated-only discovery cases. `tr01-wf04-negative-20261026-01` failed exactly the partial-output case (19/20 passed, outer 2): scanner returned success after failed discovery. `go-staticcheck.sh` now checks captured discovery status before filtering and rejects an empty result. V-003 passed at `tr01-wf04-scanner-20261026-01` (20 tests, 4306 ms); `make lint CHECK=go` passed at `tr01-wf04-go-20261026-01` (13003 ms). Both outer exits 0, source `sha256:543bd226a806823e80b4f153878a78d24c60b1f0dbc0c8161643dab0b9fb582a`. | Roots under `.cartulary/test-results/`; cleanup completed and secret scans passed. Intentional migration: incomplete discovery now fails before analysis; successful package/rule selection unchanged. WF-04 complete before WF-05. |
| E-007 | Initial WF-05 finalization passed (`tr01-finalize-20261026-01`, outer 0, 29054 ms), no generated mutations; retained maintenance skipped because RESULTS_DIR unset. Final scanner/source/scripts/shell/Go/output invocations passed. Default contract `tr01-final-contract-20261026-01` failed (outer 2, 114443 ms): the graph umbrella imported a backend-private test module. Removed that cross-owner import and registered the LPT test directly in the canonical operation command. | Structural registration repair, no boundary exemption or production export. Initial portfolio is superseded for final-source acceptance; regenerate/finalize and requalify. |
| E-008 | After E-007, regeneration passed (`tr01-final-generate-20261026-02`); the exact boundary row passed (`tr01-final-boundaries-20261026-02`, outer 0, 17248 ms). `env -u RESULTS_DIR CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-finalize-20261026-02 make agent-finalize` passed (outer 0, 28903 ms), with zero generated changes. The final portfolio below passed on one executable source identity. | Finalization precedes all accepted final invocations. Retained maintenance is explicitly excluded; ordinary canonical validity, cleanup and scans passed separately. |
| E-009 | `CARTULARY_TEST_RUN_ID=tr01-final-markdown-20261026-01 make lint-markdown` passed (outer 0, 7610 ms); summary `.cartulary/test-results/tr01-final-markdown-20261026-01/adhoc/lint-markdown/tool-run-summary.json`. Final whitespace review passed. All twelve sections, assertion successors, migration limits and compatible acceptance evidence are present. | WF-05 complete; all six sequential workstreams are closed. Only completion metadata changed after documentation validation; executable source remains E-008. |

### Final acceptance portfolio

Every command below used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run ID>`. Every
native command exited 0. Roots are `.cartulary/test-results/<run ID>`; each contains
`run-manifest.json`, `run-summary.json`, target/unit results, `cleanup-results.json`
and `retained-secret-scan.json`.

Executable source: `sha256:b296423ca57d61784fda96b379452ee93e05f609ad2fda8d08a584a56b1d8efc`.
Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
Earlier-source passes and failed characterization runs are historical evidence,
not part of this final acceptance union.

| Run ID | Command | Duration ms | Added assurance |
| --- | --- | --- | --- |
| `tr01-final-contract-20261026-02` | `make harness-contract` | 114368 | Default contracts, graph matrix, owner boundaries and real Make/private diagnostic channel; 11/11 units |
| `tr01-final-scanner-20261026-02` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.go_scanner_wrappers` | 14457 | All 20 wrapper cases, including discovery failure before scan |
| `tr01-final-output-20261026-02` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.output_accounting` | 27139 | Distinct step/target/tool publication, including valid and malformed Govulncheck findings |
| `tr01-final-source-20261026-02` | `make source-check CHECK=all` | 35700 | Generated, schema, catalog, identity, surface, toolchain and migration-input consistency |
| `tr01-final-scripts-20261026-02` | `make lint CHECK=scripts` | 625 | Authored script validation |
| `tr01-final-shell-20261026-02` | `make lint CHECK=shell` | 23784 | Shell analysis including the changed Staticcheck wrapper |
| `tr01-final-go-20261026-02` | `make lint CHECK=go` | 14032 | Actual Go formatting, vet and Staticcheck invocation |

For every listed root, `make harness-observability-check RESULTS_DIR=<exact-root>
CHECK=all` passed with outer exit 0 and `full_validity=1`. All cleanup outcomes
were `completed`; retained-secret scans passed; no failed, skipped or cancelled
units remain. This qualifies the selected obligations, not complete acceptance
of every owner or the whole project. No performance claim follows from these
functional timings; optional resource diagnostics were not collected.

Full CI, browser/visual/UI-review, dedicated race-suite execution, benchmarks,
release qualification and complete audits of other owners were not run: no
production behavior in those areas changed. Their selection/isolation guarantees
remain covered by the graph assertions. Finalizer canonical retained-evidence and
scheduler-drift maintenance were skipped because `RESULTS_DIR` was unset; no
compatible full warm check was supplied and no redundant check was created for
maintenance. This does not replace the per-run validity checks recorded above.

### Handoff boundaries and migration

Review this iteration as one source/test/routing change set, including its new
untracked authored files. The preceding frontend/Fallow and type-export changes
remain separate pre-existing work; this iteration preserves them. WF-33 in the
preceding handoff remains separately open.

| Owner | Final organization / responsibility |
| --- | --- |
| Work graph | `test-selection-policy.mjs` uses small policy inputs; `test-graph-contracts.mjs` and `test-compilation.mjs` cover real wiring, grouping and resources; `aggregate-assertions.mjs` consumes the canonical matrix. Boundary/command registration stays inert. |
| Diagnostics | `test-plan-inputs.mjs` owns input admission; `test-canonical-target-plan.mjs` owns independent planning/execution/replay, deterministic results, CLI and purity. |
| Backend | `backend/tests/test-go-lpt-planning.mjs` owns pure packing assertions and is a direct file argument of the canonical graph operation, respecting private import boundaries. |
| Execution | `execution/tests/test-unit-selection.mjs` constructs and validates its own published artifacts before corruption; it never borrows the enclosing run. |
| Static analysis | Four named `test-go-*.mjs` modules and `go-scanner-fixtures.mjs` own wrapper invocation, environment, findings and artifact permissions. Runtime diagnostic transport and output publication remain separate. |
| Routing | `harness_work_graph_owner.json`, `task_surface_owner.json`, the command-surface family and verification owner declare one scanner operation/row and one graph operation. Task-surface manifest and topology render index are Make-generated. |

The retired private graph fast/matrix/target-plan operations and four scanner
operations have no aliases or forwarding wrappers. Use existing public Make
presets, the graph routing row, or the new exact scanner row. No public command,
artifact schema, dependency, product data, golden, or database migration is needed.
Staticcheck now rejects discovery errors even after partial output; callers must
repair package discovery instead of accepting an incomplete analysis. Successful
rule and package selection remains unchanged.

No production API was added for tests. No owner-boundary suppression or import
exception was introduced. Security thresholds, generated exclusions, independent
full-tier sentinels, the three fast-smoke roles, and distinct runtime/race/security
contexts remain verified. Executable consumers do not consume this handoff or
other Markdown.

Rollback must restore this iteration's authored code/tests/routing and generated
projections together, preserving prior unrelated work and historical results.
Rollback requires its own validation and must not reintroduce compatibility
aliases. Changes remain uncommitted; no commit, push or PR was requested.

## 11. Open Questions and Blockers

No unresolved implementation or validation blocker remains in this bounded
iteration. WF-00–WF-05 and F-001–F-006 are complete. Human review of owner/projection fidelity remains required by
the repository procedure; no test result substitutes for that review.

## 12. Binary Completion Criteria

- [x] Every affected assertion has a successor or reasoned retirement.
- [x] Duplicate graph modes and target-plan execution paths are removed without aliases.
- [x] Policy expectations are independent; repository wiring and distinct contexts remain covered.
- [x] Scanner fixtures use bounded real subprocesses and structured assertions.
- [x] Staticcheck never scans after failed or empty discovery.
- [x] Public commands, schemas, security policies and product guarantees remain intact.
- [x] Authored inputs and generated projections agree.
- [x] Required validation passes on compatible source identities, with cleanup and secret scans.
- [x] No executable consumer depends on Markdown.
- [x] Each workstream has its checkpoint; WF-05 records final evidence and exclusions.

All scoped exit criteria are met. No further execution is queued after the final
documentation check. Review and land the coordinated change set; retain the
recorded exclusions and the separate WF-33 status.
