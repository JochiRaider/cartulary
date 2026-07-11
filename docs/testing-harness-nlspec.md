---
doc_id: cartulary.testing_harness.v1
title: Testing Harness NLSpec
conformance_profile_id: cartulary.testing_harness.current.v1
doc_type: nlspec
status: adopted/current
authority_boundary: Harness mechanics only; command invocation, target selection, scheduling, fixture lifecycle, service ownership, artifact emission, summary emission, cleanup, and harness verification gates.
---

## 1. Status, Scope, and Authority

This NLSpec defines the Cartulary testing harness subsystem. It is adopted current authority for the harness mechanics identified in `authority_boundary`. Adoption does not make harness readiness evidence product conformance or Core 05 claim-publication evidence.

**TH-HARNESS-REQ-001**
This NLSpec owns only harness mechanics: command invocation, target selection, scheduling, fixture lifecycle, service ownership, artifact emission, summary emission, cleanup, and harness verification gates. It MUST NOT define product behavior owned by Core 00 through Core 04. It MUST NOT define claim-publication or benchmark-publication behavior owned by Core 05.

Frontend readiness mechanics introduced by `browser-e2e-visual`, `browser-e2e-a11y`, `frontend-evidence-audit`, `tools/frontend_phase_registry.json`, `tools/frontend_phase_maps/*.json`, `tools/frontend_visual_fixture_registry.json`, `docs/testing/frontend_phase_coverage_ledgers/*.md`, `cartulary.frontend_phase_registry.v5`, `cartulary.frontend_phase_test_map.v4`, `cartulary.frontend_row_accounting.v4`, `cartulary.frontend_visual_fixture_registry.v3`, `cartulary.frontend_accessibility_summary.v2`, and `cartulary.frontend_claim_publication_review.v1` are harness and implementation-readiness mechanics only. They MUST NOT define Core product behavior, MUST NOT promote visual or accessibility evidence into product-conformance evidence, and MUST NOT activate Core 05 claim-publication review unless a claim-bearing publication predicate is active.

Where repository planning, handoffs, or command descriptions use the phrase "production evidence", the current harness interpretation is release-readiness evidence. Release-readiness evidence is a release gate input, not product conformance by itself and not Core 05 claim-publication evidence unless an owner document later promotes a narrower claim-bearing publication boundary.

Harness-local cache mechanics introduced by `cartulary.cache.readiness.v1`, `cartulary.cache.build_artifact.v1`, `cartulary.cache.static_analysis.v1`, `cartulary.agent_finalize_action_cache_record.v1`, and `cartulary.execution_topology_render_cache.v1` are local acceleration mechanics only. They MUST NOT define product behavior, MUST NOT weaken public-target summary emission, MUST NOT replace drift, security, service-readiness, cleanup, runtime-reset, or generated-artifact verdicts, and MUST NOT be cited as product-conformance, release, benchmark, or Core 05 publication evidence.

Same-run helper artifact references introduced by `cartulary.same_run_helper_artifact_ref.v1` are retained harness diagnostics only. They MAY explain that an aggregate consumed helper/setup artifacts produced earlier in the same run, but they MUST NOT mark scheduler work as `reused`, MUST NOT reference old retained runs, MUST NOT skip selected conformance rows, and MUST NOT be cited as product-conformance, release, benchmark, or Core 05 publication evidence.

Fallow static-analysis mechanics introduced by `frontend-fallow-static`, `.fallowrc.json`, `tools/fallow/*`, and `cartulary.fallow_static_summary.v1` are harness and implementation-support mechanics only. They MUST NOT define product behavior, MUST NOT replace TypeScript, Biome, frontend import-boundary checks, tests, security scans, generated-artifact drift checks, or harness gates, MUST NOT activate Fallow Runtime behavior, and MUST NOT be cited as Core 05 publication evidence.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-016, TH-HARNESS-AC-022, TH-HARNESS-AC-026, TH-HARNESS-AC-028

**TH-HARNESS-REQ-002**
A harness conformance claim MUST identify this NLSpec version, the exact public Make target or target set under evaluation, the conformance environment from Section 14, and the retained result root/run ID/run root when retained harness artifacts are used as evidence.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-016

**TH-HARNESS-REQ-003**
The current canonical public invocation surface is Make. In the current profile, a command invocation is canonical only when invoked as `make <target>` from the repository root or through a Make-owned wrapper that preserves the target identity.

Each public target MUST also declare one stable `command_id` using the form `cartulary.harness.command.<name>.v1`. The `command_id` identifies the command's semantic contract. The Make target name is the current invocation binding for that semantic command. A later adopted NLSpec MAY add additional invocation bindings only when they preserve the same `command_id`, configuration contract, output contract, artifact contract, failure mapping, and cleanup behavior.

Public behavior change gate. Any change to public Make target identity, stable `command_id`, declared schema IDs, retained artifact paths, output shape, failure mapping, or task-surface/topology metadata MUST be specified here before implementation and before generated task-surface or topology outputs are refreshed. Private module movement that preserves those public contracts MAY proceed without changing this gate.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005

**TH-HARNESS-REQ-004**
Generated files under `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, generated task/schedule artifacts, and generated Make includes are downstream generated artifacts. They MUST NOT be hand-edited and MUST NOT become behavior owners unless a later adopted NLSpec explicitly promotes one of them.

Standalone generated-artifact renderers that produce service-backed schedule source files MUST write only to an explicit caller-supplied output path. `tools/harness/generated-artifacts/render-service-backed-schedule-manifest.mjs` MUST NOT create an implicit repo-local `tools/scheduler_service_sources.json`; Make-owned phase-schedule generation is the only current owner for checked-in scheduler and topology artifacts derived from service-backed schedule sources.

When a browser harness helper path change affects task-surface, topology, or schedule outputs, the owner input MUST be updated first and downstream artifacts MUST be refreshed through Make-owned generation. The current refresh ladder is: run `make phase-schedules` when task-surface/topology/schedule generated outputs change, then verify with `make phase-schedule-drift`, `make generate-drift`, `make generated-artifact-policy-check`, and `make json-shape-check`. Generated files MUST NOT be edited as the source of truth for helper-path migration.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

**TH-HARNESS-REQ-005**
Direct package scripts, raw scripts, raw Go/Vitest/Playwright/Biome/Vite/pnpm commands, and tool-specific reports are developer conveniences or child commands unless a public Make target invokes them. Direct invocation of those surfaces MUST NOT be treated as equivalent to a canonical harness run.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-005

## 2. Purpose, Non-Goals, and Conformance Boundary

The testing harness exists to provide a reproducible repository command surface for local developers, CI entrypoints, coding agents, and release verification. It provides deterministic target selection, bounded output, structured artifacts, explicit service ownership, controlled fixture lifecycle, stable failure classification, and destructive cleanup gates.

**TH-HARNESS-REQ-006**
The harness MUST provide all of the following for public Make targets:

- deterministic target-class and target-selection metadata;
- declared configuration resolution;
- output-mode behavior;
- exit-code mapping;
- retained artifact identity when a target declares artifacts;
- failure classification that separates product assertion failures from harness operational failures;
- cleanup predicates for every destructive operation.
Verified by: TH-HARNESS-AC-001..TH-HARNESS-AC-017

**TH-HARNESS-REQ-007**
The harness MUST NOT claim provider-specific hosted CI behavior, benchmark publication, release publication readiness, visual-snapshot refresh authority, macOS support, Windows-native support, Podman support, or Playwright artifact schema stability unless those areas are explicitly included in this NLSpec's current conformance profile or in a later adopted NLSpec.
Verified by: TH-HARNESS-AC-012, TH-HARNESS-AC-016

**TH-HARNESS-REQ-008**
Logical scheduler resources are execution constraints inside the harness. They MUST NOT be represented as guarantees about physical CPU, I/O, Docker, database, object-store, browser, or network capacity.
Verified by: TH-HARNESS-AC-006

**TH-HARNESS-REQ-009**
Adopted Sections 1 through 17 MUST close every current-conformance harness behavior they name. In those sections, open delegation phrases such as `where declared`, `target-specific`, `target-owned`, `producer contract`, `tool-defined`, `where applicable`, or `bounded` are valid only when the same sentence or immediately surrounding requirement cites a closed table, schema attachment, algorithm, or explicitly non-normative diagnostic boundary. Generated manifests and generated Make includes MAY mirror a closed contract, but they MUST NOT be the only current-conformance owner for a public harness behavior.

In adopted Sections 1 through 17, `MAY` means true implementation freedom whose divergent realizations remain interchangeable to callers. `SHOULD` is advisory only and MUST NOT define behavior required by Section 17 acceptance criteria. Any behavior used by an acceptance criterion MUST be expressed with `MUST` or with an explicitly closed equivalent.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-029

## 3. Terminology

| Term                     | Meaning                                                                                                                                                    |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| harness run              | One invocation of a canonical harness command or one child invocation explicitly tied to a result root and run ID.                                         |
| target                   | A named Make target or scheduler target selected by the harness.                                                                                           |
| public target            | A target classified as public in the public command registry and canonical only through Make.                                                              |
| child target             | A target invoked by an aggregate, sequence, scheduler, or wrapper target.                                                                                  |
| work unit                | One scheduler-visible executable unit with an identity, dependencies, resource claims, logs, status, and optional completion keys.                         |
| scheduler                | A harness runner that executes a manifest-defined DAG using logical resource claims and emits scheduler events and summaries.                              |
| lifecycle machine        | A normative finite state contract for one harness lifecycle, including its closed states, closed events, allowed transitions, failure mapping, and evidence. |
| representational lifecycle diagram | A non-normative diagram or list that explains an existing lifecycle without adding requirements.                                                   |
| state                    | One named lifecycle condition inside a lifecycle machine.                                                                                                  |
| event                    | One named input signal that can be presented to a lifecycle machine.                                                                                       |
| transition               | One allowed movement from a source state to a destination state for a specific event and guard.                                                            |
| terminal state           | A state that ends a lifecycle machine instance. No later transition is allowed from a terminal state.                                                      |
| result root              | The root directory that contains run artifacts. The default is `.cartulary/test-results`.                                                                  |
| run ID                   | The run directory name under the result root. The default format is defined in Section 6.                                                                  |
| run root                 | The directory `normalize_result_root(CARTULARY_TEST_RESULTS_DIR) / normalize_run_id(CARTULARY_TEST_RUN_ID)`.                                               |
| harness artifact         | A file or directory produced by a harness run, child command, service, scheduler, test runner, or diagnostic tool.                                         |
| retained artifact        | An artifact preserved after command exit for a specific result root, run ID, and target.                                                                   |
| release-readiness evidence | A retained harness artifact that aggregates product, frontend, visual, accessibility, harness, release, and support evidence for release gating while preserving each record's owner-derived semantic effect. |
| generated artifact       | A file produced from owner inputs by a generator and checked for drift.                                                                                    |
| cache record             | A repo-local JSON record that proves a cache key, input digest, output digest, and profile-specific output contract for a local acceleration profile.      |
| cacheable output         | A deterministic file or directory output whose digest is declared in a cache record and may be reused only when all profile validation succeeds.           |
| non-cacheable side effect | Target behavior that must still execute or be emitted for the current run, such as summaries, failure classification, cleanup, service readiness, or drift/security verdicts. |
| same-run helper artifact ref | A schema-owned retained reference that links an aggregate consumer to helper/setup artifacts produced earlier under the same run root, with declared inputs, producer artifact digests, consumer refs, and fail-closed scope. |
| fixture                  | Test setup state created for a test, package, target, scheduler group, browser stack, or service suite.                                                    |
| service-backed fixture   | A fixture that uses Postgres, object-store services, Docker/testcontainers, browser processes, or Compose-backed services.                                  |
| backing services         | Postgres, object-store services, Docker/testcontainers, Compose services, backend processes, frontend processes, and browser runtime dependencies used by harness targets. |
| output mode              | The resolved mode from Section 7 that controls stdout, stderr, and artifact summary behavior.                                                              |
| machine output           | The `machine` output mode defined in Section 7. For public Make targets that accept it, stdout is exactly one UTF-8 JSON object followed by LF.            |
| failure class            | A coarse normalized grouping for failed harness commands: `product`, `config`, `infra`, `harness`, `artifact`, `timing`, `interrupted`, or `unknown`.      |
| failure reason           | A detailed snake-case reason code used for diagnostics, exit-code mapping, automation, and handoff.                                                        |
| cleanup tier             | A named cleanup scope such as repo-local clean, repo-local distclean, service-suite cleanup, browser-stack cleanup, or stale janitor cleanup.              |
| stale janitor            | A cleanup routine that removes previously generated DBs, buckets, containers, or browser fixtures only when proof predicates match.                        |
| diagnostic-only artifact | An artifact retained for human investigation whose internal shape is not a machine-readable harness conformance contract.                                  |

Domain and product terms keep their meanings from the product specs and `docs/domain.md`.

## 4. Public Command Surface

**TH-HARNESS-REQ-050**
The public command registry MUST be owned by this NLSpec and mirrored by `tools/task_surface_manifest.json` with `target_class="public"`. The implementation MUST provide exactly the public targets listed in the target registry below unless the manifest and this NLSpec are revised together.

`tools/execution_topology_manifest.json` MAY provide scheduler topology, child-work topology, generated schedule inputs, or resource-profile inputs. It MUST NOT independently add, remove, rename, reclassify, or change the output class, artifact policy, schema policy, side-effect declaration, command identity, or public lifecycle state of a public target.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-027

**TH-HARNESS-REQ-051**
Every public target MUST declare exactly one output class, exactly one stable-summary schema policy, and exactly one artifact policy. The output-class behavior is owned by Section 7. The schema policy is owned by Section 8.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005, TH-HARNESS-AC-023

**TH-HARNESS-REQ-052**
A Make-owned wrapper MAY invoke package scripts, raw scripts, or external tools as implementation mechanisms. The wrapper remains responsible for the public target's configuration, output, artifact, failure, exit-code, and cleanup contract.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004

**TH-HARNESS-REQ-058**
A public target MUST provide a semantic harness operation, not merely an alias for one or more child commands. A target qualifies as semantic only when it owns at least one observable behavior from the table below in addition to invoking child work.

| Semantic behavior | Observable requirement |
| --- | --- |
| `configuration_resolution` | Resolves and validates declared harness configuration before child work. |
| `evidence_normalization` | Emits or validates retained artifacts under a stable schema. |
| `failure_normalization` | Maps child or harness failures to Section 9 `failure_class`, `failure_reason`, and public exit code. |
| `service_lifecycle` | Owns service startup, readiness, fixture lifecycle, lease, or cleanup proof. |
| `scheduler_orchestration` | Selects, orders, and executes work units using the scheduler contract. |
| `destructive_safety` | Applies cleanup or reset proof predicates before mutation. |
| `security_boundary` | Applies redaction, token gating, artifact-safety, or secret-handling behavior. |
| `diagnostic_synthesis` | Converts retained evidence into a bounded human or machine diagnostic that cannot be obtained from raw child output alone. |

A target that provides none of these behaviors MUST be private child work or a developer convenience outside the public command registry.
Verified by: TH-HARNESS-AC-020

**TH-HARNESS-REQ-053**
The default `check-harness-smoke` gate MUST remain a small semantic smoke surface rather than a broad harness regression suite. Its fast tier MUST contain exactly one check for each gate role: public Make/wrapper projection, check-scheduler semantics, and service-backed scheduler semantics. Broader field-shape, topology-rendering, and sequence-detail checks MUST live in owner-aligned validation such as `json-shape-check`, generated drift checks, the explicit `harness-contract` extended target, or non-default diagnostic smoke tiers. The `harness-contract` target MUST be selected by CI and release gates, not by default local `check`.

Fast smoke fixtures that create disposable repo-shaped files, directories, fake Make surfaces, manifests, or child-run workspaces MUST create them through `tools/harness/test-support/harness-scratch.sh` outside the repository checkout. `CARTULARY_HARNESS_SCRATCH_ROOT` MAY redirect that scratch root only when it still resolves outside the repository. Repo-local `tmp/` remains reserved for durable tool caches, retained run artifacts, and operator-inspectable local outputs; fast smoke fixtures MUST NOT place transient package-shaped or source-shaped scratch trees there, so concurrent source traversal such as `go list ./...` cannot observe disappearing non-package directories.

Non-default harness smoke tiers MAY carry owner-specific regression checks for harness maintenance surfaces such as finalization, evidence audit, topology generation, scheduler behavior, and wrapper compatibility. A retained harness self-test MUST be reachable from an owner-controlled tier, merged into another active owner test, or deleted with named replacement coverage; manual-only harness self-test files outside owner-controlled tiers MUST NOT be treated as active coverage.

The current owner-controlled harness smoke tiers are `fast`, `execution`, `extended`, `lifecycle`, and `full`. Make helper targets MUST expose the active tiers as `run-harness-smoke-fast`, `run-harness-smoke-execution`, `run-harness-smoke-extended`, `run-harness-smoke-lifecycle`, and `run-harness-smoke-full`. The execution helper target is the narrow validation surface for shell phase execution, Vitest execution wrappers, frontend-unit execution wrapping, Make-node dispatch, service-backed execution runner delegation, and fast Make-sequence wrapper behavior. The lifecycle helper target is the narrow validation surface for browser/dev stack lifecycle, reset, readiness, and teardown harness changes. The execution and lifecycle helper targets MUST remain helper-only and MUST NOT become default local `check` work by themselves.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-006

**TH-HARNESS-REQ-054**
The default local `check` gate MUST prioritize correctness evidence and MUST NOT enforce duration-baseline drift. Duration-baseline coverage MAY remain in `check` because it validates scheduler input completeness. Duration-baseline drift MUST remain available through explicit duration drift targets and MUST be enforced by CI, `agent-finalize RESULTS_DIR=<dir>`, or another timing-maintenance surface.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-006, TH-HARNESS-AC-027

Default `check` placement MUST be selected by evidence ownership rather than target history. Base phase-map rows and frontend phase-map rows MUST use the current closed default-check metadata set: `default_check_required`, `default_check_kind`, `default_check_reason_code`, `primary_evidence_owner`, `duplicate_of`, `evidence_delta`, and `warm_local_cost_class`. Rows whose default inclusion is not self-evident MUST also declare `default_check_reason`. `default_check_required=true` is only valid for implemented, executable rows in the current harness.

| Evidence class | Default `check` placement | Required owner behavior |
| --- | --- | --- |
| `product_conformance` | Include when the row is authoritative, executable, and not a duplicate or explicit-only browser/readiness projection. | Preserve local correctness evidence; use the cheapest layer that owns the behavior. |
| `implementation_support` | Exclude by default unless a row declares `default_check_required=true` and a reason. | Keep support evidence reachable through direct targets, slices, `test`, or CI. A promoted support row MUST retain an evidence owner and the reason MUST state the product or harness risk that justifies local-check promotion. |
| `harness_policy` | Include only for narrow policy gates required to keep the local harness trustworthy. | Prefer dedicated policy targets over broad self-test suites. |
| `readiness` | Include only when downstream default work consumes the readiness result. | Model warm/cold readiness separately from product evidence. Content-addressed readiness caches MAY accelerate provisioning only when the target still validates installed state and emits normal summaries. |
| `visual_readiness` and `accessibility_readiness` | Exclude full suites by default; bounded smoke projections MAY enter with a reason. | Keep direct browser targets full-fidelity. |
| `measurement` | Exclude from default local `check`. | Keep measurement evidence on explicit targets and timing-maintenance surfaces. |
| `security` | Include when the security target is part of local correctness, but never stamp or cache it. | Keep vulnerability and security scans uncached. |
| `release_artifact` | Exclude from default local `check`; include in release/CI gates. | Keep deployable/release evidence separate from local correctness unless explicitly promoted. |
| `duplicate_regression` and `diagnostic` | Exclude by default. | Retain through explicit diagnostic, test, or CI targets when useful. |

The current closed backend phase-map `layer` values are `backend_unit`, `backend_store`, `backend_integration`, `backend_integration_support`, `backend_process`, `browser_functional`, `browser_stateful`, `browser_measurement`, `browser_support`, `browser_visual`, and `frontend_unit`. A new layer token MUST be added to the schema and this NLSpec before phase maps may use it.

The current closed frontend phase-map contract is `cartulary.frontend_phase_test_map.v4`. Frontend `layer` values are `unit`, `integration`, `browser_integration`, `e2e`, `visual`, `accessibility`, and `support`. Frontend `evidence_class` values are `product_conformance`, `design_direction`, `implementation_support`, and `claim_publication_boundary`. Frontend current row `claim_status` MUST be `implemented`. Frontend `target.evidence_role` values are `primary`, `supporting`, `drift`, and `diagnostic_only`. Frontend `claim.claim_publication_intent` values are `none`, `informative_engineering_measurement`, and `claim_bearing_publication`. Frontend `claim.closure_scope` values are `scenario` and `target_level`.

The current closed frontend registry contract is `cartulary.frontend_phase_registry.v5`. Frontend registry `status` MUST be `active`, and `row_rollup_state` MUST be `active_green`. The registry MUST contain the current contiguous phase range beginning at `FE-P0` and ending at the highest declared `FE-P<N>`; the current repository data declares `FE-P0` through `FE-P11`. Each frontend phase item MUST declare `base_phase_join` as either a base `phase<N>` identifier from `tools/phase_registry.json` or `null`. A non-null `base_phase_join` is the only current owner for attaching frontend readiness rows to base-phase browser selection, browser duration, and base Playwright selection surfaces; it is not namespace equivalence. The current repository data declares `FE-P0` through `FE-P11` joined to `phase0` through `phase11`, respectively. A `null` join means the frontend phase has no base-phase attachment and MUST NOT be inferred from its numeric suffix. `phase-slice`, `service-backed-slice`, and direct standalone frontend-aware targets MUST execute only active frontend phases and implemented active frontend rows. Planned, stale, retired, blocked, not-implemented, future, or arbitrary historical frontend phase replay is not current conformance.

Frontend registry freshness MUST validate `guide_digest`, each `manifest_digest`, each `ledger_digest`, and each `evidence_freshness_digest` against the current guide, frontend phase maps, generated ledgers, frontend visual fixture registry, and frontend registry/map/row-accounting schema attachments. A stale freshness digest MUST fail frontend phase validation and MUST NOT permit an `active_green` phase claim.

The current closed `default_check_kind` values are `primary_local_evidence`, `default_local_cross_stack_conformance`, `full_target_equivalent`, `bounded_readiness`, `explicit_only`, and `duplicate_regression`. The current closed `default_check_reason_code` values are `cheapest_authoritative_layer`, `lower_layer_gap`, `full_target_equivalent_stateful`, `bounded_readiness`, `explicit_full_target`, `explicit_readiness`, `explicit_measurement`, `implementation_support_explicit_only`, `design_direction_explicit_only`, `claim_publication_boundary`, and `duplicate_of_primary_owner`. The current closed `warm_local_cost_class` values are `none`, `low`, `medium`, `service_backed`, `browser`, and `explicit_heavy`.

Task-surface default inclusion MUST agree with scheduler reachability. `default_inclusion_sets` describes direct full-target membership only: a target that lists `check` there MUST run as direct `check` work, be an explicit public aggregate summarized by `check`, or be otherwise full-target equivalent. Bounded check evidence for a full public target MUST instead declare `check_projection` metadata with `mode=projection`, `schedule`, `stage`, `evidence`, `evidence_class`, `reason_code`, `full_target`, and `full_target_equivalent=false`; projection mode MUST NOT also advertise ordinary direct `check` membership. A public browser target whose default `check` evidence is full-target-equivalent scheduler work MUST declare `check_projection` metadata with `mode=direct`, the same schedule/stage/evidence fields, and `full_target_equivalent=true`. `explain-target`, task-surface diagnostics, generated task-surface manifests, and scheduler manifests MUST render the same projection or direct metadata. Support-only internal targets MUST NOT advertise `check` membership unless the check scheduler actually selects them.

**TH-HARNESS-REQ-055**
The check scheduler MUST NOT skip required correctness work through digest-only input stamps in the current local profile. Default local `make check` MUST execute every selected static, drift, security, product, service-backed, browser, service-resource mutation, runtime reset, and scratch database apply work unit unless a future NLSpec revision defines a reusable artifact cache with complete retained provenance.

`local_input_stamp`, `CARTULARY_CHECK_DISABLE_INPUT_STAMPS`, `tmp/check-stamps/`, and `cartulary.check_input_stamp.*` records are retired in the current profile. Scheduler manifests and execution-topology owner inputs MUST reject `local_input_stamp` rather than treating it as diagnostic metadata. A future reusable artifact cache MAY classify work as `reused` only when the cache record validates the relevant tracked and untracked inputs, tool versions or binary digests, configuration inputs, expected outputs, summary schema, and artifact references needed to diagnose the reused work from a single retained run.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-006, TH-HARNESS-AC-028

**TH-HARNESS-REQ-060**
The current profile permits only these local acceleration cache families:

- `cartulary.cache.readiness.v1` for toolchain and install readiness, including pinned Node/pnpm readiness, frontend install readiness, Playwright install readiness, pinned Go helper binaries, ShellCheck, test-service image readiness, and scheduler readiness units that only validate provisioning state.
- `cartulary.cache.build_artifact.v1` for deterministic build artifacts, including `build-server`, `build-migrate`, `build-operator`, `build-web`, `testservices-build`, and embedded web asset preparation.
- `cartulary.cache.static_analysis.v1` for deterministic non-security static analysis success stamps, currently limited to `lint-markdown` and strict `lint-shell`.
- `cartulary.agent_finalize_action_cache_record.v1` for the closed `agent-finalize` action IDs listed in Section 8.2.
- `cartulary.execution_topology_render_cache.v1` for internal deterministic execution-topology render content only.

Embedded web asset preparation is a build-artifact producer in the current profile. When the embedded output is consumed by Go `//go:embed`, the publisher MUST update the embedded source-tree artifact atomically from Go embed's point of view. It MUST NOT delete or rewrite a directory of hashed frontend assets in place while concurrently scheduled Go compilation can traverse that directory. Harness-only readiness stamps, cache records, and other operational metadata MUST remain outside the embedded content root unless the owning product spec explicitly makes that file served application content.

All permitted cache families MUST be content-addressed. A valid key MUST include the cache schema ID, cache scope, profile ID, platform identity where relevant, declared tool or runtime versions, declared command/profile inputs, helper implementation digests, and every declared output contract needed by the profile. Broad timestamp-only caching is not a valid harness cache mechanism.

Readiness, build-artifact, and static-analysis cache hits MAY skip only the deterministic provisioning, build, or analyzer command guarded by the cache profile. Static-analysis cache hits MUST be limited to closed non-security profiles whose keys include tool version or binary digest, configuration digests, selected input digests, helper implementation digests, analyzer arguments, and output stamp digests. They MUST NOT skip the public target wrapper, public target summary emission, failure classification, output validation, drift comparison, security scan execution, service lifecycle, service readiness, fixture cleanup, runtime reset, scratch database apply, object-store mutation, destructive-operation guard, or aggregate success/failure computation. Build-artifact and static-analysis cache hits inside `make check`, `make ci`, or `make release-check` MUST NOT be reported as scheduler `reused` work in the current profile.

Test-service image readiness is a readiness cache profile only. Its cache key MUST include the testservices binary digest, image-owner source digests, helper implementation digests, and toolchain pins. A cache hit MUST still prove that every pinned service image named by the testservices helper is locally present before accepting the readiness stamp; missing images invalidate the stamp and force the ordinary image warmup command. This cache profile MUST NOT replace service startup readiness, test service lifecycle evidence, browser reset, cleanup, fixture preparation, or product-conformance evidence.

Cache records that are missing, disabled, forced, invalid, corrupt, or whose declared outputs are missing or digest-mismatched MUST NOT produce success by reuse. They MUST either execute the underlying command and emit a miss/disabled/invalid-cache artifact, or fail as `configuration_error` or `artifact_error` when the target cannot prove the output contract. Security targets, drift verdicts, generated-artifact drift detection, service-backed/browser/live-state tests, cleanup targets, destructive reset targets, and aggregate `check`/`ci`/`release-check` success MUST remain uncached.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-006, TH-HARNESS-AC-018, TH-HARNESS-AC-019, TH-HARNESS-AC-024, TH-HARNESS-AC-028

**TH-HARNESS-REQ-056**
The default local `check` gate MUST keep ordinary browser measurement evidence out of the warm `check-service-backed` critical path. `browser-e2e-measurement` MUST remain available as an explicit public target and MAY remain required by CI, release, or explicit browser aggregate targets, but default local `make check` MUST NOT schedule the measurement browser stage as a `check-service-backed` child.

Default local `make check` MUST NOT schedule current full visual or accessibility browser work. Direct `make browser-e2e-visual`, direct `make browser-e2e-a11y`, explicit browser aggregates, and CI/release profiles that select full visual or accessibility evidence MUST retain their full-fidelity target behavior. A later bounded visual or accessibility readiness projection MAY enter default `check` only after this NLSpec and the task-surface metadata declare a separate bounded-readiness row with `check_projection.mode=projection`, `full_target_equivalent=false`, and a reason that identifies the downstream default-check consumer.

Base `phase-slice` and `service-backed-slice` MAY select visual browser work only when the selected base phase has executable `browser_visual` rows. That work MUST run the ordinary `browser-e2e-visual` visual stage with the selected base phase propagated to manifest selection, so retained Playwright selection artifacts identify only the requested phase's current-profile `V-*` rows. Base phase slices MUST NOT run or enforce unrelated frontend `FE-*` visual readiness rows. Frontend namespace slices MAY run `browser-e2e-visual` for selected frontend rows, but the frontend readiness selection MUST be constrained to the slice's selected frontend row accounting scope.

Browser rows selected by default `check` MUST declare `default_check_required=true` plus a reason that identifies the cross-stack browser value that lower layers cannot provide, or must be full-target-equivalent stateful evidence with `default_check_reason_code=full_target_equivalent_stateful`. Default-check browser rows MUST be classified as `default_local_cross_stack_conformance`, `full_target_equivalent`, or `duplicate_regression`; default local `check` MAY select only implemented executable rows with lower-layer-gap reasons, full-target-equivalent stateful reasons, or separately declared bounded-readiness reasons. Direct public browser targets MUST remain full-fidelity for their declared target, including explicit measurement, full visual, accessibility, and isolated stack behavior where the target owns it. A default-check browser projection MUST NOT be treated as proof that the direct target's full evidence ran.

Default browser schedule generation MUST apply row-level admissibility after stage selection and before work-unit emission. A row is admissible for default local `check` only when its owner metadata permits default selection under the rules above. A browser group with explicit `selected_row_ids[]` MAY be projected by removing inadmissible row IDs; if no row IDs remain, the group MUST be omitted from the default schedule. A mixed base/frontend group MUST retain only the admissible IDs and MUST keep reset, taint, browser-session, teardown, and target-summary behavior for the retained group. A generated default schedule MUST fail closed when an emitted selected row ID is unknown, not implemented/executable, marked `default_check_required=false` with `default_check_kind=explicit_only`, marked with an explicit-only reason such as `explicit_full_target`, `explicit_measurement`, `design_direction_explicit_only`, `claim_publication_boundary`, or has `warm_local_cost_class=explicit_heavy` without a deliberate owner promotion. Direct public browser targets, explicit `test`, `ci`, `release-check`, browser aggregate, phase-slice, and frontend namespace slice schedules MUST NOT use this default-check filtering unless their owner metadata explicitly declares a projection.

Default browser projection artifacts MUST expose the selected scope. For frontend rows selected by a default browser projection, retained `cartulary.frontend_row_accounting.v4` artifacts MUST use `accounting_scope.mode="selected_rows"` with the exact selected frontend row IDs. If a default browser group retains only base rows and no frontend rows, frontend row accounting MUST be disabled for that group rather than closing broad active frontend target rows. Standalone direct browser targets MUST keep broad `active_target` accounting when they map frontend rows outside an explicit slice or projection.

The `browser-e2e-stateful` public target MAY be implemented as generated `stateful_partition` browser groups when every partition declares a selected base phase, explicit selected row IDs, and an isolated browser session group. Partitioning MUST preserve the same stateful base row IDs and frontend row-accounting IDs selected by the unpartitioned public target. Frontend-only partitions MUST skip empty base manifest execution rather than treating absence of base rows as success for product rows; mixed partitions MUST execute the selected base phase and the selected frontend rows in the same isolated browser session. Direct `browser-e2e-stateful` execution MUST reset between stateful partitions, while scheduler execution MUST start and finalize each partition session independently. Stateful partitioning MUST NOT remove reset, taint, teardown, route-token, runtime identity, frontend row-accounting, or target-summary evidence.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-018

**TH-HARNESS-REQ-057**
Warm steady-state `check-service-backed` timing is a harness health contract, not product performance evidence. For the supported WSL2 compatibility profile, a successful warm `check` run accepted for harness maintenance MUST keep `check-service-backed` wall time at or below the hard compatibility cap of `155000ms` unless the caller explicitly supplies a different maintenance budget to the timing checker. Backend and browser scheduler lanes SHOULD remain duration-balanced: no non-isolated peer lane should materially exceed `125%` of its peer median. A lane MAY be excluded from peer balance only when it is explicitly isolated by the shard plan or is the only lane in its peer group, and the checker MAY apply the default `5000ms` bounded materiality floor so normal fixture jitter does not fail otherwise healthy retained runs.

Timing-accepted runs MUST be warm-ready runs. Cold provisioning, pinned tool installation, Go build-cache population, frontend install, Playwright install, service image build or warmup, and hidden helper builds are valid correctness evidence but MUST be attributed as readiness or provisioning work rather than accepted as warm scheduler-health evidence. At minimum, the service test binary build MUST be modeled as first-class scheduler readiness work separate from service image warmup.

Retained-run finalizer duration and finalizer-maintained baseline refresh effects are not part of the evaluated `make check` wall time. When `agent-finalize RESULTS_DIR=<run-root>` updates duration baselines, schedule inputs, generated schedule outputs, or topology render artifacts that can affect later warm scheduling, the updated files MUST be treated as affecting the next warm baseline only. A timing improvement claim after such a refresh MUST be based on a later warm-ready `make check` run produced after the refresh, not on the pre-refresh run or on finalizer duration.
Verified by: TH-HARNESS-AC-018

### 4.1 Mechanism Boundary

Harness implementation packages are owner-local realization details unless a row in this NLSpec explicitly promotes a specific path. Module names and directory names under `tools/harness/**` MUST NOT be treated as public harness invocation bindings, product conformance surfaces, or durable compatibility promises. Implementation code SHOULD depend on owner facades for contract, output, execution, finalization, diagnostics, smoke, frontend, browser, backend, scheduler, phase-accounting, and generated-artifact behavior instead of importing private mixed-responsibility helper paths. The harness import-boundary check MUST reject private catch-all package imports once an owner-specific facade exists for the behavior.

Phase manifest shape, registry semantics, evidence naming, fixture policy, row selection, frontend fixture references, and Go/Vitest/Playwright run verification are phase-accounting behavior. Task guidance, target explanation, phase explanation, retained-run explanation, target-plan display, and fixture diagnostics are diagnostics behavior. Backend target/shard row discovery belongs to backend harness helpers; browser duration/readiness discovery belongs to browser helpers; frontend row metadata belongs to frontend phase-accounting helpers; test-output manifest indexing and target-start statistics belong to test-output helpers; summary topology belongs to execution helpers; task-surface validation and generated Make rendering belong to generated-artifact helpers. Public Make target names, stable `command_id` values, schema IDs, output/failure contracts, and retained artifact paths remain the compatibility surface. A diagnostic surface that combines backend, browser, frontend, task-surface, and scheduler reachability data MUST depend on the owner facades for each source family rather than importing private owner-local helpers.

### 4.1A Private Helper Ownership, Facades, and Compatibility

**TH-HARNESS-REQ-061**
Private helper paths under `tools/harness/**` MUST be classified by semantic ownership, not by historical directory name. A helper path MAY move, be renamed, or be deleted when public Make target names, stable `command_id` values, schema IDs, output shape, retained artifact paths, failure mapping, cleanup behavior, and declared public input contracts remain unchanged.

The following compatibility statuses are closed:

| Status | Definition |
| --- | --- |
| `public_contract` | A Make target, `command_id`, schema ID, output/failure contract, retained artifact path, or declared public input contract owned by this NLSpec. |
| `owner_facade` | A declared semantic import boundary for an implementation helper family. Non-owner callers MUST import this facade when the behavior is needed from another harness subsystem. |
| `unsupported_private` | A helper path that is not a public compatibility path and may move, be renamed, or be deleted without public contract revision when public harness behavior is preserved. |

The following helper ownership registry is closed for the current profile:

The scheduler helper rows below own harness orchestration only. They MUST NOT define product HTTP routes, WebSocket behavior, workbook mutation/query behavior, saved-view behavior, product transport adapters, frontend shell state, or grid-vendor integration; those questions remain owned by Core 00 through Core 04 and adopted subsystem owner documents.

| Helper family | Facade key | Owner boundary | Compatibility status | Observable behavior to preserve |
| --- | --- | --- | --- | --- |
| Backend Go target row discovery and backend target-plan DTOs. | `backend_target_plan` | Backend harness target planning. | `owner_facade` | Backend row identity, ordering, target selection, backend field values consumed by `target-plan-json`, and Go fixture-policy details. |
| Backend Go shard planning and shard-plan DTOs. | `backend_shard_plan` | Backend harness shard planning. | `owner_facade` | Shard names, ordering, fixture policy, target mapping, and shard-plan JSON. |
| Backend Go target execution, capture, finalization, and report emission. | `backend_target_execution` | Backend harness target execution. | `owner_facade` | CLI command set, command text, Go test environment, shared report locking and reuse, runtime-binary validation, phase summaries, target summaries, timing spans, finalizer diagnostics, and failure classification. |
| Go duration baseline coverage, drift, update inputs, and retained Go shard observations. | `backend_duration_accounting` | Backend duration accounting. | `owner_facade` | Baseline schema identity, coverage semantics, read-only drift, update inputs, and default weights. |
| Migration-history and schema-object ownership validation. | `database_contract_drift` | Generated-drift `database_contract_drift` sub-boundary. | `owner_facade` | Manifest schema IDs, current-line lineage validation, deterministic diagnostics when present, fresh and penultimate scratch migration behavior for `migration-drift`, and `json-shape-check` compatibility. |
| Govulncheck findings normalization and redaction. | `static_analysis_security_findings` | Static-analysis/security sub-boundary. | `owner_facade` | Redaction behavior, blocking/non-blocking classification, exit-code mapping, `cartulary.govulncheck_findings.v1`, and `GOVULNCHECK_DB` as the current public vulnerability database override. |
| Task-surface, topology, generated drift, schema attachment validation, and generated Make rendering. | `generated_artifact_surface` | Generated-artifact helpers. | `owner_facade` | Generated artifact source-of-truth parity, schema validation, drift failure mapping, and no hand-editing of generated outputs. |
| Public Make Node-tool command registry, input filtering, child environment construction, and invocation argument synthesis. | `command_surface_node_tool_dispatch` | Command-surface helpers; current facade `tools/harness/command-surface/make-node-tools.mjs`. | `owner_facade` | Public Make node-tool input matrix parity, inherited-environment stripping, runtime-env forwarding, usage diagnostics, and generated Make dispatch behavior. |
| Shared shell phase execution runtime, timing, redaction, artifact directory policy, output mode resolution, shell phase summaries, target summaries, and Vitest watchdog substrate. | `phase_execution_runtime` | Execution runtime helper boundary; current sourceable facade `tools/harness/execution/phase-runtime.sh`. | `owner_facade` | Phase and target timing spans, redacted stdout/stderr handling, retained artifact directories, output mode behavior, shell phase summary emission, public exit-code propagation, and watchdog sidecar behavior. |
| Frontend phase maps, row IDs, fixture references, guide restatements, and slice planning. | `frontend_phase_accounting` | Phase-accounting frontend sub-boundary. | `owner_facade` | Frontend registry and phase-map schema IDs, validation diagnostics, scenario-title grep behavior, ledger rendering inputs, and phase-slice plan semantics. |
| Frontend row-accounting artifact generation and validation. | `frontend_row_accounting` | Phase-accounting frontend row-accounting sub-boundary. | `owner_facade` | `cartulary.frontend_row_accounting.v4`, row closure semantics, selected-row scope validation, retained artifact path `frontend-row-accounting.json`, and target-summary failure injection. |
| Frontend retained evidence audit. | `frontend_evidence_audit` | Phase-accounting frontend audit boundary. | `owner_facade` | `frontend-evidence-audit` retained-root inputs, digest validation, closure audit semantics, and `cartulary.frontend_evidence_audit_summary.v1`. |
| Test-output title and artifact indexing for frontend row evidence. | `test_output_frontend_indexing` | Test-output helper boundary. | `owner_facade` | Vitest and Playwright title indexes, normalized frontend row evidence observations, runner artifact parsing, and the test-output DTO consumed by frontend row accounting. |
| Frontend unit target execution. | `frontend_target_execution` | Execution helper boundary. | `owner_facade` | `frontend-unit` command behavior, Vitest invocation shape, phase summaries, target summaries, selected-row filtering, and public failure mapping. |
| Vitest execution diagnostics and sidecars. | `vitest_execution_diagnostics` | Execution, test-output, and diagnostics helper boundaries. | `owner_facade` | Vitest phase wrappers, manifest-phase filtering, watchdog handling, `cartulary.vitest_failure_details.v1`, and retained `vitest-failure-details.json` sidecar paths. |
| Frontend toolchain and dependency-install readiness. | `frontend_toolchain_readiness` | Readiness helper boundary. | `owner_facade` | Pinned repo-local Node/pnpm preparation, frozen-lockfile install behavior, readiness cache keys, stamp content, and configuration-error exit mapping. |
| Web build and embedded web asset artifacts. | `web_build_artifact` | Readiness/build-artifact helper boundary. | `owner_facade` | `build-web` and embedded asset cache behavior, Vite build invocation, embed archive/stamp atomicity, and public build target behavior. |
| Design-token generation. | `design_token_generation` | Generated-artifact design-token sub-boundary. | `owner_facade` | `docs/design.md` token parsing, generated token TypeScript content, generated provenance identity, and generated-artifact drift behavior. |
| Font asset validation. | `font_asset_validation` | Static-analysis helper boundary. | `owner_facade` | Font manifest validation, vendored font checksum/license checks, local CSS activation checks, and remote-font ban diagnostics. |
| Browser batch manifest loading, normalization, and target/stage metadata. | `browser_batch_manifest` | Browser manifest helper boundary; current canonical path `tools/harness/browser/browser-batch-manifest.mjs`. | `owner_facade` | `cartulary.browser_e2e_batch_manifest.v5`, deterministic stage/group normalization, group selection, target/stage metadata, and diagnostics used by generated-artifact and execution helpers. |
| Browser scheduler adapter command/dependency projection. | `browser_scheduler_adapter` | Scheduler browser adapter boundary; current facade path `tools/harness/scheduler/adapters/browser.mjs`. | `owner_facade` | Browser work-unit command paths, stage/session completion keys, group dependency keys, worker slot ranges, worker environment variables, and scheduler expansion semantics. |
| Scheduler DAG execution, events, finalizers, summaries, and failure mapping. | `scheduler_execution_core` | Scheduler execution helper boundary; current facade `tools/harness/scheduler/scheduler-runner.mjs`. | `owner_facade` | Work-unit ordering, logical resource scheduling, scheduler event stream, scheduler summary emission, finalizer handling, stable failure classes/reasons, public Make target behavior, and retained scheduler artifact paths. |
| Scheduler manifest, resource, family, and reporting helpers. | `scheduler_contract_helpers` | Scheduler contract helper boundary; current facades `tools/harness/scheduler/scheduler-family-contract.mjs`, `tools/harness/scheduler/scheduler-manifest.mjs`, `tools/harness/scheduler/scheduler-resources.mjs`, and `tools/harness/scheduler/scheduler-reporting.mjs`. | `owner_facade` | `cartulary.scheduler_manifest.v2`, scheduler family tokens, scheduler resource registry behavior, resource default/auto policies, retained scheduler paths, bounded reporting lines, and artifact references. |
| Scheduler child process execution and log handling. | `scheduler_process_execution` | Scheduler process adapter boundary; current facade `tools/harness/scheduler/process-executor.mjs`. | `owner_facade` | Child environment construction, sensitive environment stripping, dry-run detection, log-name sanitization, log replay, and child exit/failure propagation. |
| Phase-slice plan construction and selected work-unit accounting. | `phase_slice_planning` | Phase-accounting helper boundary; current facade `tools/harness/phase-accounting/phase-slice-plan.mjs`. | `owner_facade` | Phase namespace/row selection, phase-slice JSON plan shape, selected work-unit ordering, phase-map evidence accounting, and scheduler-runner coordination inputs. |
| Service-backed schedule manifest expansion and topology planning. | `service_backed_schedule_planning` | Execution/service-backed planning boundary; current facade `tools/harness/execution/service-backed/schedule-planning.mjs`. | `owner_facade` | Service-backed schedule source validation, expansion, browser group/session mapping, resource mapping, topology checks, and generated scheduler manifest behavior. |
| Scheduler-retained duration baseline and drift accounting. | `scheduler_duration_accounting` | Duration-accounting helper boundary; current facade `tools/harness/duration-accounting/index.mjs`. | `owner_facade` | Shared duration thresholding, retained-run contamination detection, service-backed target duration baselines, harness-smoke duration baselines, read-only drift behavior, and retained-run validation before baseline mutation. |
| Scheduler retained event/timing diagnostics. | `scheduler_evidence_drift` | Diagnostics helper boundary; current entrypoints `tools/harness/diagnostics/scheduler-event-order-drift-cli.mjs` and `tools/harness/diagnostics/scheduler-summary-timing-drift-cli.mjs`. | `owner_facade` | Event-order drift, summary timing drift, warm-run eligibility checks, lane/session accounting diagnostics, bounded output, and retained-run evidence requirements. |
| Browser target execution wrappers and stage dispatch. | `browser_target_execution` | Browser target execution helper boundary; current canonical private runner `tools/harness/browser/run-browser-e2e-target.sh`. | `owner_facade` | Browser target wrapper entrypoint behavior, batch stage dispatch, target/stage summaries, stack session wrapping, reset boundaries, artifact paths, and public failure mapping. |
| Browser Playwright phase, manifest phase, webserver-batch execution, report parsing, and selection artifacts. | `browser_playwright_execution` | Browser Playwright execution plus test-output adapter boundary; current stable adapter `tools/harness/output/test-output/playwright-artifacts.mjs`. | `owner_facade` | `cartulary.playwright_manifest_selection.v1`, Playwright runner report interpretation, selected-test title/file indexing, merged report behavior, phase summaries, stdout/stderr/output artifact paths, and failure normalization. |
| Browser duration accounting, shard planning, drift, and baseline refresh inputs. | `browser_duration_accounting` | Browser duration accounting boundary; current canonical facade `tools/harness/browser/browser-duration-accounting.mjs`. | `owner_facade` | `cartulary.browser_e2e_duration_baselines.v3`, `cartulary.browser_e2e_shard_plan.v2`, manifest row ID joining, frontend readiness row inclusion, read-only drift semantics, explicit retained-run validation before mutation, and default weights. |
| Browser owned-stack lifecycle, runtime identity proof, and runtime reset adapter. | `browser_lifecycle_adapter` | Browser lifecycle/test-route adapter boundary; current entrypoints `tools/harness/browser/start-web-e2e.sh` and `tools/harness/browser/reset-web-e2e-stack.sh`. | `owner_facade` | `cartulary.web_e2e_stack.v3`, preview-mode startup, port ownership, runtime root/session files, process-group cleanup, runtime identity proof, reset route token/origin/host predicates, reset taint handling, and Playwright state cleanup. |
| Browser accessibility evidence summaries. | `browser_accessibility_evidence` | Browser helper boundary; current canonical path `tools/harness/browser/accessibility-summary-cli.mjs`. | `owner_facade` | Accessibility summary schema, contrast record handling, retained Playwright runner references, and browser a11y target artifact paths. |
| Browser visual snapshot update helper. | `browser_visual_update_helper` | Browser visual maintenance helper boundary; current entrypoint `tools/harness/browser/run-browser-e2e-visual-update.sh`. | `owner_facade` | Helper-only visual update target posture, snapshot-update mode propagation, authorized authored snapshot write path, retained browser evidence, and exclusion from default `check`, `test`, `ci`, and release gates unless separately declared. |

Migration-history and schema-object ownership validators are implementation-support evidence for database-contract drift. They MUST NOT become Core product behavior owners and MUST NOT be imported by backend target execution code except through a declared `database_contract_drift` facade.

Govulncheck findings normalization belongs to `static_analysis_security_findings`. Legacy backend helper paths for that behavior are `unsupported_private` and are invalid as current compatibility paths.

Verified by: TH-HARNESS-AC-038, TH-HARNESS-AC-040, TH-HARNESS-AC-041

**TH-HARNESS-REQ-062**
Private helper compatibility MUST NOT be preserved by default. Archive references, historical handoffs, raw script paths, and old implementation imports do not establish compatibility support.

The current profile rejects `unsupported_private` helper imports by semantic
rule family rather than by preserving every historical helper path as an
exact compatibility registry entry. The following rule families are closed:

| Rule family | Matches | Owner facade or successor |
| --- | --- | --- |
| `legacy_backend_database_contract_drift` | `tools/harness/backend/drift/**`, legacy backend migration-history helpers, and legacy backend schema-object ownership helpers. | `database_contract_drift` |
| `legacy_backend_duration_and_shard_helpers` | `tools/harness/backend/duration/**` and `tools/harness/backend/runner/**`. | `backend_duration_accounting`, `backend_shard_plan` |
| `legacy_backend_security_findings_helper` | Legacy backend `govulncheck-findings` helper path. | `static_analysis_security_findings` |
| `legacy_frontend_catch_all_directory` | Any helper under the retired `tools/harness/frontend/**` catch-all tree. | Frontend phase-accounting, row-accounting, evidence-audit, execution, readiness, browser, generated-artifact, and static-analysis owner facades. |
| `legacy_scheduler_backend_adapters` | Legacy scheduler backend adapter and schedule-context helper paths. | `backend_shard_plan`, execution summary topology, and backend target-plan facades. |
| `legacy_scheduler_phase_slice_and_service_backed_helpers` | Legacy scheduler phase-slice, service-backed expansion/topology, and execution-dependency helper paths. | `phase_slice_planning`, `service_backed_schedule_planning`, and execution topology helper facades. |
| `legacy_scheduler_duration_helpers` | Legacy scheduler duration-baseline, duration-drift, harness-smoke duration, service-backed duration, and target-duration helper paths. | `scheduler_duration_accounting` |
| `legacy_scheduler_process_and_evidence_drift_helpers` | Legacy nested scheduler process-executor and scheduler event/timing drift CLI helper paths. | `scheduler_process_execution`, `scheduler_evidence_drift` |
| `legacy_execution_phase_runtime_and_node_registry` | Legacy catch-all execution helper paths `tools/harness/execution/run-phase-common.sh` and `tools/harness/execution/make-node-tools.mjs`. | `phase_execution_runtime`, `command_surface_node_tool_dispatch` |

Deletion or contraction of an `unsupported_private` family is allowed only
after current live callers are moved to declared owner facades,
generated/task-surface metadata does not reference the old path family,
import-boundary tests reject reintroduction, and characterization tests for
the relevant public targets pass. The import-boundary report MAY include
diagnostic example patterns for these rule families, but exact historical
helper-path inventories are not compatibility surfaces.

Verified by: TH-HARNESS-AC-039

**TH-HARNESS-REQ-063**
Harness implementation code MUST import declared owner facades rather than arbitrary private backend or frontend helper paths once the facade for that behavior exists. The harness import-boundary check MUST reject new non-owner imports from private backend implementation helpers, imports of `unsupported_private` helper paths, and private catch-all imports where an owner facade is declared.

Tests MAY import private implementation fixtures only from declared test-support paths. Generated files MAY mirror declared owner paths but MUST NOT independently widen import allowances. Non-owner harness code MUST NOT import private browser implementation helpers directly once a browser owner facade exists; scheduler code MUST use the scheduler browser adapter instead of direct browser helper imports. Import-boundary failures are harness failures, not product failures.

Verified by: TH-HARNESS-AC-038, TH-HARNESS-AC-039

**TH-HARNESS-REQ-064**
Go duration-baseline maintenance MUST use target-specific retained-run eligibility. `duration_retained_run` means retained run evidence that contains the target summaries, scheduler summaries, scheduler events, and Go shard/duration artifacts required to bind observed durations to current target and shard identities.

| Command family | Retained-run rule |
| --- | --- |
| `go-test-duration-baseline-coverage` | Does not require retained-run evidence. It verifies planned baseline coverage only, is read-only, and may remain ordinary `check` evidence. |
| `go-test-duration-baseline-drift` | MAY use `RESULTS_DIR` explicitly supplied by the caller or the current retained run when invoked inside a retained-run context. It MUST be read-only and MUST reject missing, failed, incomplete, or artifact-insufficient evidence before producing a drift verdict. |
| `go-test-duration-baselines` | MUST require explicit `RESULTS_DIR`. It MUST reject ambiguous result roots unless exactly one retained run is resolved by Sections 5 and 6. It MAY mutate baseline files only after retained-run validation succeeds. |
| `agent-finalize duration_baseline_refresh` | MUST use an existing retained full warm `make check` run root. Service-backed-only, phase-slice, browser-only, and other partial roots are invalid. |
| `agent-finalize duration_baseline_drift_validation` | MUST use the same retained-run requirement as `duration_baseline_refresh`, but remains read-only. |

Duration retained-run evidence MUST be rejected if it is failed, incomplete, contaminated, non-warm where warm evidence is required, missing full-check markers where full-check evidence is required, artifact-insufficient, or older than the latest sibling retained check run without the Section 8.2 older-run override where that override applies. A mutating baseline update MUST fail before the first mutation when retained-run validation fails.

Verified by: TH-HARNESS-AC-042

**TH-HARNESS-REQ-065**
Migration-history evidence capture is database-contract or migration-evidence evidence unless a later adopted owner explicitly promotes another boundary. Harness rows for migration manifest audit, embedded SQL source audit, goose ledger inspection, schema-object ownership drift, and migration-history diagnostics MUST route through database-contract or migration-evidence ownership. They MUST NOT be classified as operator-recovery conformance evidence merely because an implementation exposes a deployment-local operator wrapper.

A deployment-local wrapper for migration evidence MAY exist as an implementation mechanism. Omission behavior: if the wrapper is absent, harness conformance MAY still be satisfied through owner-backed database-contract drift or migration-evidence targets, provided the phase rows cite the correct owner and retained artifacts satisfy the declared schema.

Verified by: TH-HARNESS-AC-045

**TH-HARNESS-REQ-066**
When production code or tests move across packages, harness accounting source-of-truth inputs MUST be updated before generated artifacts. The required order is:

1. Update the owner phase-map row, target-map row, runtime-binary declaration, helper ownership row, task-surface input, topology input, or schedule input that owns the changed path or selection.
2. Regenerate downstream task-surface, schedule, topology, phase-ledger, coverage-ledger, and row-accounting artifacts only through Make-owned generation.
3. Run drift and schema checks before treating the move as complete.
4. Treat any generated-ledger or generated-manifest hand edit as non-conformant.

A path-only test move that preserves row ID, title, target, owner references, evidence class, runtime-binary use, and retained artifact shape MAY remain behavior-preserving. A move that changes command grammar, result schema, authorization outcome, runtime-binary use, public target membership, default-check membership, scheduler topology, fixture lifecycle, or retained artifact shape is a public harness behavior change and MUST be specified in this NLSpec before implementation.

Required validation after moved-test accounting changes is:

| Change class | Required validation |
| --- | --- |
| Phase-map or row ownership only | `make json-shape-check` plus the affected target. |
| Task-surface or public-target metadata | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` and public-target parity checks. |
| Scheduler topology or phase schedule | `make phase-schedules`, `make phase-schedule-drift`, and the affected scheduler target. |
| Generated artifacts | `make generate-drift` and `make generated-artifact-policy-check`. |
| Operator runtime-binary rows | `make build-operator` and the affected scheduler-selected operator work. |

Verified by: TH-HARNESS-AC-046

**TH-HARNESS-REQ-067**
Private backend test-support packages MUST be named for their semantic owner
and responsibility. Phase labels remain admissible as evidence-accounting
identifiers, test IDs, and phase-map selectors, but a new or retained helper
package, exported helper type, or exported helper symbol MUST NOT use a
`phaseNtest`, `phaseNstoretest`, or equivalent rollout-history name.

The repository MUST maintain one schema-validated test-support inventory that
classifies every shared and owner-local Go support root and every committed
shared fixture or golden root. The inventory MUST record semantic owner,
shared or owner-local posture, runtime-scan treatment, support-scan treatment,
and whether a Go root exposes service-starting `Start*` entrypoints. A support
root that is absent, duplicated, missing on disk, excluded from both runtime
and support security profiles, or inconsistent with its service-starting
classification MUST fail harness validation. Private compatibility packages
MUST NOT be retained solely to preserve an old helper import.

Verified by: TH-HARNESS-AC-056

The current canonical private runners for phase-slice child work are `tools/harness/execution/run-frontend-unit.sh` and `tools/harness/browser/run-browser-e2e-target.sh`. Legacy root `scripts/run-frontend-unit.sh`, legacy root `scripts/run-browser-e2e-target.sh`, legacy frontend catch-all runners under `tools/harness/frontend/**`, and legacy `tools/harness/core/explain-run-cli.mjs` shims MUST NOT be recreated as compatibility paths; callers MUST use the owning execution, browser, or diagnostics helper path through Make-owned invocation surfaces.

`tools/harness/execution/cartulary-runner-cli.mjs` private direct use MUST select an explicit runner subcommand. Backend Go target execution is available only through `go-target <target-or-command> [...]`; direct aliases such as `backend-unit` or `backend-store` are unsupported private compatibility and MUST fail with usage status `2`. Quiet successful child logs MUST remain suppressed for public summaries; legacy opt-in replay through `CARTULARY_ENABLE_LEGACY_SUCCESS_LOG` is unsupported and MUST NOT emit child stdout or stderr on successful quiet runs.

Harness import-boundary validation MUST cover statically resolvable shell `source` and `.` references to repo-local `tools/harness/...` paths in addition to JavaScript imports. Unsupported private helper rules apply equally to those shell-source edges. Dynamic shell source expressions that cannot be resolved to a repo-local static path remain outside this static rule and are covered by shell lint plus target-level tests.

| Surface                                                  |                                  Normative? | Required contract                                                                      |
| -------------------------------------------------------- | ------------------------------------------: | -------------------------------------------------------------------------------------- |
| Public Make target name                                  |                                         yes | Stable command surface invoked as `make <target>` from the repository root.            |
| `tools/task_surface_manifest.json` public target_class |                                         yes | Required machine-readable mirror of the public target registry.                         |
| Root/package `pnpm` scripts                              |                                          no | Developer convenience unless invoked by a Make-owned public target. Successful raw package-script output MUST NOT be reported as completion evidence for public harness targets. |
| Raw owner helper scripts and child CLIs                   | no | May change when public Make behavior remains unchanged.                                |
| Make-owned harness contract helper implementation path    | no as an additional invocation binding; yes as the owner implementation path used by Make wrappers | Make-owned public wrappers invoke this owner CLI for preflight, cleanup, schema validation, and related harness-contract mechanics. Direct CLI invocation remains implementation support unless a public Make target adopts it. |
| Make-owned test-output helper implementation path         | yes only for Make-owned command behavior; no for JavaScript module boundaries | Make-owned wrappers may invoke these helpers to emit lifecycle lines, machine output, summaries, and retained artifacts. Command names, arguments, output modes, schema IDs, failure taxonomy, exit codes, and retained artifact paths are the contract; helper filenames, exports, imports, and private module locations are implementation details unless explicitly promoted by this NLSpec. |
| `tools/testservices` binary path                         |                                          no | Service lifecycle behavior is normative; binary path is an implementation realization. |
| Public output classes and schema IDs listed in Section 8 |                                         yes | Required machine-output and artifact validation contracts.                              |
| Docker image tag for Postgres or object-store services   | no unless declared in a service fixture row | Exact tag is not normative unless it defines fixture semantics in Section 11.          |
| Generated Make include names, helper binaries, helper target classes, priority-band names, and generator constants | no | Implementation detail unless promoted by an explicit requirement.                      |

### 4.2 Command Family Defaults

Target membership for each family is defined only by `### 4.3 Public Target Registry`. The family-default table owns shared behavior for targets whose registry row declares that family. If a target appears in a prose command list and not in the registry, the registry governs and the prose is editorial drift.

| Family | Family ID | Required inputs | Optional inputs and defaults | Output class family | Scheduler use | Backing services | Artifact behavior | Failure contract |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| help and discovery | `help_discovery` | Target-local inputs declared by the Section 5.3 per-target input registry. | Omitted optional inputs select documented summary views according to the target's `input_contract`. | The target row's Section 4.3 output class, constrained by Section 7.2. | None | None | Does not create central run evidence unless the target row declares a summary schema. | Usage/config errors use Section 9. |
| bootstrap and toolchain | `bootstrap_toolchain` | Required local tools according to target | Tool paths default from Section 5. | `summary_with_artifacts` | None | May download/install repo-local tools. | Tool-run summary required; readiness cache artifacts MAY be retained when a cache profile is active. | Tool/config failures are `configuration_error` or `preflight_error`. |
| local services and dev | `local_services_dev` | Docker Compose and local config required by the target row. | `CONFIG_FILE=configs/dev/config.toml`, `OBJECT_STORE_BUCKET=cartulary` for rows that read those variables. | `service_summary` or `interactive_raw` | None | Compose Postgres/SeaweedFS S3 and local processes. | `service_summary` rows emit service summaries; `dev` has no verification artifact contract. | Startup/readiness/config failures are harness operational failures. |
| generated and drift | `generated_drift` | Owner inputs and manifests | `RESULTS_DIR` rows in the Section 5.3 input matrix select retained evidence. | `summary_with_artifacts` | Child work is scheduled only through Section 10 normalized scheduler work units. | Migration drift may use scratch Postgres when scheduled. | Tool summary and command-specific drift/finalizer files. | Drift mismatch is `artifact_error` or `scheduler_accounting_error`; unsafe retained-run finalization evidence is `artifact_error` or `configuration_error`. |
| phase and service slices | `phase_service_slices` | `PHASE` and any namespace/output selector declared by the Section 5.3 per-target input registry. | `PHASE_NAMESPACE=base`; `JSON` only for targets listed with `JSON` in the Section 5.3 input matrix. | `scheduler_summary_with_artifacts` | Uses phase selection and service-backed scheduler. | Service-backed slice requires backing services when phase work does. | Target, run, scheduler, and phase artifacts. | Missing/invalid/ambiguous phase is `usage_error`; child failures retain child class. |
| backend and frontend leaf tests | `backend_frontend_leaf_tests` | Toolchain, manifests, package inputs | Parallelism and worker variables from Section 5. | `summary_with_artifacts` | Service-backed targets may use a scheduler or testservices. | Store/integration/process targets require Postgres/object-store services when service-backed. | Phase, target, tool, logs, reports. | Product assertion failures are `test_assertion_failure`; setup failures are operational. |
| browser E2E | `browser_e2e` | Node/pnpm, Playwright browser, backend/migrate/server support, services | `PLAYWRIGHT_WORKERS=3`, `BROWSER_E2E_FUNCTIONAL_SHARDS=auto` unless overridden by Section 5. | `summary_with_artifacts` | Uses browser batch and service-backed scheduler only for rows closed by Section 10. | Postgres, object-store service, backend, frontend, browser runtime. | Browser stack, Playwright, reset, target, scheduler artifacts. | Product assertions are product failures; stack/readiness/reset failures are operational. |
| aggregates and gates | `aggregates_gates` | Toolchain and child inputs | `summary` output mode by default; `ci` defaults to `ci` mode through Make-owned CI target environment. | aggregate or scheduler output classes | `check` uses the check scheduler. | Service-backed and browser children require backing services. | Aggregate run, child target, scheduler, and tool summaries. | Exit nonzero if any required child fails or artifact validation fails. |
| static analysis and security | `static_analysis_security` | Toolchain and source roots | Rule, flag, package, and Fallow static profiles named by public target rows and Section 5.3 inputs. Shell lint is blocking for public Make targets. `GOVULNCHECK_DB` is the only public security-profile override in the current profile. Fallow Runtime and Fallow security-scan commands are not selected by this profile. | `summary_with_artifacts` | Scheduled inside `check` only through Section 10 normalized scheduler work units. | None | Tool summary and logs; security scans are uncached. | Findings are gate failures for scheduled local correctness targets. Advisory targets MUST be explicitly selected outside local `check`. |
| builds | `builds` | Build inputs and toolchain | Output paths from Make variables. | `summary_with_artifacts` | Scheduled as readiness work only through Section 10 normalized scheduler work units. | None | Tool summary and build logs; build cache artifacts MAY be retained when a cache profile is active. | Build failures are gate failures. |
| cleanup | `cleanup` | None | Uses Make path registries. | `destructive_human` | None | Does not stop Docker Compose services. | No central summary contract. | Unsafe path guard failure exits nonzero; missing paths are not failures. |
| formatting | `formatting` | Toolchain | None | `summary_with_artifacts` | None | None | Tool summary and formatter logs. | Formatter failure is operational; formatter rewrites are mutating. |

### 4.3 Public Target Registry

Every command below inherits the matching family defaults. `Default inclusion sets` lists direct full-target default memberships only; bounded default-check evidence for a full target is described by `check_projection` metadata instead of ordinary `check` membership. `helper_only` means the target is public and directly invocable, but is not selected by default by `test`, `check`, `ci`, or `release-check` unless another registry row explicitly includes it. `helper_only` MUST NOT mean private, uncontracted, or exempt from public-target output, configuration, failure, and cleanup contracts.

`Command ID` is the stable semantic command contract; the Make target is the current invocation binding. `Family ID` binds the target to Section 4.2 family defaults. `Semantic behaviors` declares the observable harness operation required by TH-HARNESS-REQ-058. `Side effects` declares the target's intentional mutation and resource contract from TH-HARNESS-REQ-059. The visible `Side effects` column MUST match the `side_effects[].class` list in `tools/task_surface_manifest.json`. `Lifecycle state` is defined by Section 4.6.

Public aggregate targets MAY be represented in scheduler manifests by typed internal helper work when the stable public command contains distinct resource, policy, or lifecycle boundaries. `migration-drift` is the current example: direct `make migration-drift` MUST remain the public aggregate that runs static migration-input validation plus scratch database apply evidence, while default `check` MUST schedule `migration-input-drift` and `migration-scratch-apply` as separate internal helper work units so static policy validation and scratch Postgres evidence keep separate resource and artifact identities.

| Target | Command ID | Family ID | Default inclusion sets | Output class | Stable summary schema | Semantic behaviors | Side effects | Lifecycle state | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `help` | `cartulary.harness.command.help.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `help-all` | `cartulary.harness.command.help_all.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `doctor` | `cartulary.harness.command.doctor.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `bootstrap` | `cartulary.harness.command.bootstrap.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `bootstrap-node-runtime` | `cartulary.harness.command.bootstrap_node_runtime.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `frontend-toolchain` | `cartulary.harness.command.frontend_toolchain.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `frontend-install` | `cartulary.harness.command.frontend_install.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `playwright-install` | `cartulary.harness.command.playwright_install.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `db-up` | `cartulary.harness.command.db_up.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start` | `public_active` |  |
| `db-reset` | `cartulary.harness.command.db_reset.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v3` | `destructive_safety` (Section 13), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `destructive_cleanup` | `public_active` | Requires `CARTULARY_DESTRUCTIVE_CONFIRM=db-reset` unless `CARTULARY_CLEANUP_DRY_RUN=1`; resets only the local database and does not reset object storage. |
| `services-up` | `cartulary.harness.command.services_up.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start` | `public_active` |  |
| `services-down` | `cartulary.harness.command.services_down.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `destructive_safety` (Section 13), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `destructive_cleanup` | `public_active` | Stops local Compose services with named volumes preserved. |
| `object-store-init` | `cartulary.harness.command.object_store_init.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` | Starts the local object-store service if needed and initializes the configured object-store bucket without requiring Postgres. |
| `object-store-reset` | `cartulary.harness.command.object_store_reset.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v3` | `destructive_safety` (Section 13), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `destructive_cleanup` | `public_active` | Requires `CARTULARY_DESTRUCTIVE_CONFIRM=object-store-reset` unless `CARTULARY_CLEANUP_DRY_RUN=1`; clears objects only from the configured local object-store bucket. |
| `dev` | `cartulary.harness.command.dev.v1` | `local_services_dev` | `helper_only` | `interactive_raw` | none | `service_lifecycle` (Section 11) | `service_start` | `public_active` |  |
| `generate` | `cartulary.harness.command.generate.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `generate-drift` | `cartulary.harness.command.generate_drift.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `generated-artifact-policy-check` | `cartulary.harness.command.generated_artifact_policy_check.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `json-shape-check` | `cartulary.harness.command.json_shape_check.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `toolchain-drift` | `cartulary.harness.command.toolchain_drift.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `migration-drift` | `cartulary.harness.command.migration_drift.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_resource_mutation` | `public_active` |  |
| `phase-ledgers` | `cartulary.harness.command.phase_ledgers.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `phase-ledger-drift` | `cartulary.harness.command.phase_ledger_drift.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `phase-schedules` | `cartulary.harness.command.phase_schedules.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `phase-schedule-drift` | `cartulary.harness.command.phase_schedule_drift.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `agent-finalize` | `cartulary.harness.command.agent_finalize.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `frontend-evidence-audit` | `cartulary.harness.command.frontend_evidence_audit.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Audits frontend phase closure across retained broad `check`, support, visual, accessibility, and measurement roots. |
| `benchmark-claim-check` | `cartulary.harness.command.benchmark_claim_check.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Validates retained Core 05 benchmark claim artifacts when the default benchmark manifest exists; absence of the default manifest is a no-claim pass, while an explicitly configured non-default missing manifest remains a harness failure. |
| `task-surface-report` | `cartulary.harness.command.task_surface_report.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `task-guide` | `cartulary.harness.command.task_guide.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `phase-slice` | `cartulary.harness.command.phase_slice.v1` | `phase_service_slices` | `helper_only` | `scheduler_summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `scheduler_orchestration` (Section 10), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `service-backed-slice` | `cartulary.harness.command.service_backed_slice.v1` | `phase_service_slices` | `helper_only` | `scheduler_summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `scheduler_orchestration` (Section 10), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `backend-unit` | `cartulary.harness.command.backend_unit.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `backend-store` | `cartulary.harness.command.backend_store.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` |  |
| `backend-integration` | `cartulary.harness.command.backend_integration.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` |  |
| `backend-process` | `cartulary.harness.command.backend_process.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` |  |
| `otel-conformance` | `cartulary.harness.command.otel_conformance.v1` | `backend_frontend_leaf_tests` | `check`, `ci` | `summary_with_artifacts` | `cartulary.otel_conformance_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `security_boundary` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Validates source snapshot, generated constants evidence, emitted telemetry goldens, browser non-export, retained raw capture policy, and telemetry security boundaries. |
| `target-plan` | `cartulary.harness.command.target_plan.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `target-plan-json` | `cartulary.harness.command.target_plan_json.v1` | `help_discovery` | `helper_only` | `machine_stdout_json` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `fixture-report` | `cartulary.harness.command.fixture_report.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `explain-run` | `cartulary.harness.command.explain_run.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `explain-phase` | `cartulary.harness.command.explain_phase.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `explain-target` | `cartulary.harness.command.explain_target.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `go-test-duration-baselines` | `cartulary.harness.command.go_test_duration_baselines.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `go-test-duration-baseline-coverage` | `cartulary.harness.command.go_test_duration_baseline_coverage.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `go-test-duration-baseline-drift` | `cartulary.harness.command.go_test_duration_baseline_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `browser-e2e-duration-baselines` | `cartulary.harness.command.browser_e2e_duration_baselines.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `browser-e2e-duration-baseline-drift` | `cartulary.harness.command.browser_e2e_duration_baseline_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `service-backed-make-target-duration-baselines` | `cartulary.harness.command.service_backed_make_target_duration_baselines.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `service-backed-make-target-duration-baseline-drift` | `cartulary.harness.command.service_backed_make_target_duration_baseline_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `harness-smoke-duration-baselines` | `cartulary.harness.command.harness_smoke_duration_baselines.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `harness-smoke-duration-baseline-drift` | `cartulary.harness.command.harness_smoke_duration_baseline_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `scheduler-event-order-drift` | `cartulary.harness.command.scheduler_event_order_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `scheduler-summary-timing-drift` | `cartulary.harness.command.scheduler_summary_timing_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `frontend-typecheck` | `cartulary.harness.command.frontend_typecheck.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `frontend-unit` | `cartulary.harness.command.frontend_unit.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `frontend-import-boundary-check` | `cartulary.harness.command.frontend_import_boundary_check.v1` | `backend_frontend_leaf_tests` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `backend-module-boundary-check` | `cartulary.harness.command.backend_module_boundary_check.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `security_boundary` (Section 8), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Enforces backend module ownership boundaries from `tools/backend_module_boundaries.json` and emits `cartulary.backend_module_boundary_summary.v1`. |
| `frontend-fallow-static` | `cartulary.harness.command.frontend_fallow_static.v1` | `static_analysis_security` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Current helper-only Fallow static profile. Emits `cartulary.fallow_static_summary.v1`; Fallow Runtime and Fallow security scans are not selected. |
| `lint-biome` | `cartulary.harness.command.lint_biome.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `lint-scripts` | `cartulary.harness.command.lint_scripts.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `lint-markdown` | `cartulary.harness.command.lint_markdown.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Structural Markdown lint over authored active docs; generated ledgers remain generator-owned. |
| `lint-shell` | `cartulary.harness.command.lint_shell.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `format` | `cartulary.harness.command.format.v1` | `formatting` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `authored_source_write` | `public_active` |  |
| `browser-e2e` | `cartulary.harness.command.browser_e2e.v1` | `browser_e2e` | `test`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Direct aggregate remains full browser evidence and is not a default local `check` member. |
| `browser-e2e-webserver-backed` | `cartulary.harness.command.browser_e2e_webserver_backed.v1` | `browser_e2e` | `test`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Default `check` uses a `check-service-backed` `webserver-backed` projection classified as default-local cross-stack conformance with `full_target_equivalent=false`; direct target evidence remains full-fidelity. |
| `browser-e2e-stateful` | `cartulary.harness.command.browser_e2e_stateful.v1` | `browser_e2e` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Default `check` uses `check_projection.mode=direct` for `check-service-backed` `stateful` evidence with `full_target_equivalent=true`. |
| `browser-e2e-measurement` | `cartulary.harness.command.browser_e2e_measurement.v1` | `browser_e2e` | `test`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `browser-e2e-a11y` | `cartulary.harness.command.browser_e2e_a11y.v1` | `browser_e2e` | `test`, `ci`, `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Emits `cartulary.frontend_accessibility_summary.v2` for implemented accessibility rows only. Direct target evidence remains full-fidelity and default local `check` does not schedule current accessibility work. Release-check uses it as release-readiness evidence only. |
| `browser-e2e-visual` | `cartulary.harness.command.browser_e2e_visual.v1` | `browser_e2e` | `test`, `ci`, `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Direct target evidence remains full visual evidence: current-profile `V-*` visual manifest rows plus frontend visual readiness rows selected from frontend phase-map titles. Base phase slices use the same visual stage only for phase-scoped base `V-*` rows, and default local `check` does not schedule current visual work. Release-check uses it as release-readiness evidence only. |
| `browser-e2e-visual-update` | `cartulary.harness.command.browser_e2e_visual_update.v1` | `browser_e2e` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `authored_source_write`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Helper-only refresh command for committed Playwright visual goldens. It runs the same visual stack and row selection as `browser-e2e-visual` with snapshot-update mode, mutates only authored snapshot PNGs under the visual spec snapshot directory, and MUST NOT be selected by `check`, `test`, `ci`, or release gates. |
| `harness-contract` | `cartulary.harness.command.harness_contract.v1` | `static_analysis_security` | `ci`, `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Runs extended harness topology, schema, and field-shape contract checks outside default local `check`. |
| `test-fast` | `cartulary.harness.command.test_fast.v1` | `aggregates_gates` | `test`, `check`, `ci` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `test` | `cartulary.harness.command.test.v1` | `aggregates_gates` | `test`, `check`, `ci` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `lint` | `cartulary.harness.command.lint.v1` | `aggregates_gates` | `helper_only` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Sequence aggregate that emits run and target summaries for blocking lint/typecheck children. |
| `go-vulncheck` | `cartulary.harness.command.go_vulncheck.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Runs Govulncheck with structured JSON output and retains `govulncheck-findings.json` with schema ID `cartulary.govulncheck_findings.v1`. Symbol-reachable vulnerability findings are blocking `security_finding` failures; package-only and module-only findings are retained as diagnostic security evidence unless a later profile promotes them. |
| `go-gosec-targeted` | `cartulary.harness.command.go_gosec_targeted.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `go-gosec-audit` | `cartulary.harness.command.go_gosec_audit.v1` | `static_analysis_security` | `ci`, `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Advisory no-fail audit evidence. It MUST NOT be selected by default local `check`. |
| `seaweedfs-compatibility` | `cartulary.harness.command.seaweedfs_compatibility.v1` | `local_services_dev` | `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` | Runs the dedicated SeaweedFS S3 compatibility profile and emits the full `SWFS-COMP-*` report outside `services-up` as a command-specific retained artifact. |
| `standup-package-smoke` | `cartulary.harness.command.standup_package_smoke.v1` | `local_services_dev` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs`, `service_start`, `service_resource_mutation` | `public_active` | Builds and smokes the MVP on-prem stand-up package with one app image plus local Postgres and SeaweedFS S3, verifies embedded browser assets, package-local object-store init, readiness, persistent Docker-volume roots, and WebSocket Origin behavior. It is package smoke evidence only and MUST NOT be represented as disconnected-profile or backup/restore conformance evidence. |
| `standup-operational-recovery-smoke` | `cartulary.harness.command.standup_operational_recovery_smoke.v1` | `local_services_dev` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs`, `service_start`, `service_resource_mutation` | `public_active` | Builds and smokes the MVP on-prem operational recovery workflow, creates a backup through the canonical operator recovery result schema, inspects the latest retained backup through the canonical inspect command, runs due restore verification against an isolated target, proves public recovery route-family absence, and retains command-specific recovery artifacts. It is not disconnected-profile evidence and does not reclassify `standup-package-smoke` as backup/restore conformance evidence. |
| `seaweedfs-release-evidence` | `cartulary.harness.command.seaweedfs_release_evidence.v1` | `static_analysis_security` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `security_boundary` (Section 15) | `retained_artifacts` | `public_active` | Runs current SeaweedFS compatibility as a prerequisite, emits SeaweedFS release evidence, and emits a non-enforcing release-gate summary as command-specific retained artifacts; missing strict child evidence is reported as blocked evidence rather than hidden. |
| `seaweedfs-release-gate` | `cartulary.harness.command.seaweedfs_release_gate.v1` | `static_analysis_security` | `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `security_boundary` (Section 15) | `retained_artifacts` | `public_active` | Runs current SeaweedFS compatibility as a prerequisite and enforces the strict SeaweedFS release gate from current compatibility, migration-preservation, storage-ref owner, security, license, and occurrence evidence; redaction scan inputs derive from the current release artifacts and current backend-process Phase E/F artifact roots. The release-gate summary is a command-specific retained artifact. |
| `check` | `cartulary.harness.command.check.v1` | `aggregates_gates` | `check` | `scheduler_summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `scheduler_orchestration` (Section 10), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `ci` | `cartulary.harness.command.ci.v1` | `aggregates_gates` | `ci` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `release-check` | `cartulary.harness.command.release_check.v1` | `aggregates_gates` | `release-check` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Runs release child helpers for extended harness contract checks, advisory security audit evidence, SBOM/license evidence, SeaweedFS S3 release-gate evidence, builds, deployable-shape evidence, frontend support/visual/accessibility readiness children, and final release-readiness aggregation. |
| `release-readiness-evidence` | `cartulary.harness.command.release_readiness_evidence.v1` | `aggregates_gates` | `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Aggregates current-run target summaries, frontend row-accounting artifacts, visual/accessibility/support readiness artifacts, release artifacts, and harness-contract outputs into `cartulary.release_readiness_evidence.v1`. It records release-gate effects without promoting frontend/design/support evidence to product conformance or Core 05 publication evidence. |
| `build` | `cartulary.harness.command.build.v1` | `builds` | `helper_only` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` |  |
| `build-server` | `cartulary.harness.command.build_server.v1` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | In default local `check`, build evidence is readiness for downstream service-backed work, not release deployable-shape evidence. |
| `build-migrate` | `cartulary.harness.command.build_migrate.v1` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | In default local `check`, build evidence is readiness for migration and service-backed work, not release deployable-shape evidence. |
| `build-operator` | `cartulary.harness.command.build_operator.v1` | `builds` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | Selected through `build`, CI, release-shaped gates, and scheduler-visible operator runtime-binary readiness. Default local `check` builds the operator only when selected runtime-binary work declares it. |
| `build-web` | `cartulary.harness.command.build_web.v1` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v3` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | In default local `check`, build evidence is readiness for browser preview work, not release deployable-shape evidence. |
| `clean` | `cartulary.harness.command.clean.v1` | `cleanup` | `helper_only` | `destructive_human` | none | `destructive_safety` (Section 13), `failure_normalization` (Section 9) | `destructive_cleanup` | `public_active` |  |
| `distclean` | `cartulary.harness.command.distclean.v1` | `cleanup` | `helper_only` | `destructive_human` | none | `destructive_safety` (Section 13), `failure_normalization` (Section 9) | `destructive_cleanup` | `public_active` |  |

**TH-HARNESS-REQ-059**
Every public target MUST declare one or more side-effect classes in the public target registry source. The declaration MUST be represented as `side_effects[]`, where each entry is an object with `class`, `owner_section`, and the class-specific details required by the table below. A target that performs an undeclared side effect is non-conformant. `none` is mutually exclusive with every other side-effect class.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-020, TH-HARNESS-AC-023

| Side-effect class | Meaning | Required declaration |
| --- | --- | --- |
| `none` | No intentional file, service, or resource mutation outside ordinary terminal output. | Target row declares only `side_effects[].class=none`. |
| `retained_artifacts` | Writes retained run-root artifacts. | Artifact policy declares retained artifact families or paths. |
| `generated_artifacts` | Mutates checked-in generated or maintenance artifacts. | Target row declares exact generated file families. |
| `authored_source_write` | Mutates authored source files. | Target row declares source families or paths. |
| `build_outputs` | Writes reproducible build outputs. | Target row declares output roots or artifact families. |
| `tool_install` | Installs or updates repo-local tools or dependencies. | Target row declares install root and cleanup behavior. |
| `service_start` | Starts local or harness-owned services or runtime processes. | Target row declares ownership mode and lifecycle machine. |
| `service_resource_mutation` | Creates, modifies, or deletes service resources such as scratch databases, buckets, fixture resources, or local service bootstrap resources. | Target row declares ownership mode, resource families, and lifecycle machine. |
| `destructive_cleanup` | Deletes files, directories, services, databases, buckets, or other resources. | Target row cites Section 13 predicates. |
| `runtime_reset` | Mutates test runtime state through the test-only reset boundary. | Target row cites Section 12 predicates. |

### 4.4 Direct Script and Package Boundary

| Surface                                                  | Classification                           | Contract                                                                                                  |
| -------------------------------------------------------- | ---------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Root `package.json` scripts `build`, `test`, `typecheck` | Developer convenience                    | They do not promise Make result roots, run IDs, scheduler summaries, cleanup, or machine output.          |
| `apps/web/package.json` scripts                          | Developer convenience or child command   | Browser and unit scripts become harness child work only when invoked through Make wrappers.               |
| Raw owner-local helper scripts and CLIs                  | Tool-owned diagnostics or child commands | Their direct usage and exit codes are not public harness contracts unless a Make target adopts them.      |
| Raw Go, Vitest, Playwright, Biome, Vite, pnpm commands   | Tool-owned                               | Tool output schemas remain external or diagnostic unless consumed and normalized by a Make-owned wrapper. |

### 4.5 Public Wrapper Lifecycle

Every Make-owned public wrapper that is not `interactive_raw` MUST execute this observable lifecycle:

1. establish wrapper identity and target identity;
2. resolve output mode;
3. resolve and validate harness configuration;
4. compute result-root and run-id identity if the output class declares retained artifacts;
5. initialize redaction before capturing child output;
6. run the target's semantic behavior;
7. validate required schema-owned artifacts before success;
8. select the primary failure using Section 9.1;
9. run required cleanup or finalizers;
10. emit the target's public output according to Section 7;
11. expose the normalized public exit code through retained summaries and failure output, and exit nonzero on failure.

A target MAY skip a step only when its output class or target row explicitly declares that the step does not apply. A skipped step MUST NOT be implemented as an implicit child-command side effect.

When a public Make wrapper recipe directly invokes repo-owned Node tooling before delegating to child or scheduled work, the wrapper MUST make pinned repo-local Node readiness an explicit precondition before semantic behavior begins. The current Make binding satisfies this with a `$(NODE_BIN)` prerequisite rendered from owner inputs. If the pinned Node runtime cannot be resolved, installed, downloaded, verified, or executed, the wrapper MUST fail before semantic work with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2`; it MUST NOT surface a raw shell, raw `curl`, or `env` executable-not-found failure as the public target result. Node runtime bootstrap MUST serialize mutation of the repo-local runtime and archive paths, download into temporary files, publish an archive only after checksum verification, remove corrupt or partial archive candidates, and use bounded retry for transient download failures.

### 4.6 Public Target Lifecycle

A target has one of these public-lifecycle states:

| State | Meaning | Invocation behavior |
| --- | --- | --- |
| `candidate_child` | Internal or generated child work, not a public command. | MUST NOT be required for public conformance by direct invocation. |
| `public_active` | Current public command. | MUST satisfy all public target contracts. |
| `removed` | No longer public. | MUST NOT appear in the public registry. |

A target may move to `public_active` only when it passes the semantic-value test from TH-HARNESS-REQ-058. A pre-release target may move to `removed` by revising the registry and generated mirrors in one change set. `removed` is represented by absence from the public registry, not by a retained registry row.

## 5. Configuration Resolution Contract

**TH-HARNESS-REQ-100**
Every public Make target MUST resolve harness configuration through `resolve_harness_config()` before child work begins. A target that cannot resolve or validate configuration MUST fail with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2`.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-003, TH-HARNESS-AC-014

`resolve_harness_config()` is the normative configuration-resolution contract. Repository implementation entrypoints such as preflight helpers MAY wrap this resolver, but MUST NOT define a narrower public-target configuration contract.

**TH-HARNESS-REQ-101**
Generated manifests are execution inputs, not caller configuration. A caller-supplied variable that attempts to override a non-overridable manifest field MUST fail with `configuration_error` before child work.
Verified by: TH-HARNESS-AC-002

**TH-HARNESS-REQ-102**
When a scheduler work unit invokes a child runner that starts its own worker pool, the scheduler input MUST either declare logical resource claims equal to the child worker budget or constrain that child worker budget through scheduler-owned environment. The scheduler-owned value wins for that scheduled work unit even when the same variable has a different direct public-target default. In the current check profile, scheduled `frontend-unit` MUST run Vitest with `VITEST_MAX_WORKERS=2` and MUST claim `host_cpu=2`; direct `make frontend-unit` MAY keep a faster developer default outside the check scheduler.

Auto-derived scheduler capacity MUST NOT resolve below the largest declared claim for that resource in the normalized work-unit set. Caller overrides MAY still choose lower limits, but such overrides are configuration errors when they cannot satisfy a declared work unit.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-021

**TH-HARNESS-REQ-103**
Frontend unit harness tests that depend on asynchronous jsdom rendering, workbook row hydration, controlled input replacement, row-history rendering, or virtualized grid mounting MUST use shared bounded wait helpers and stable selector builders with actionable diagnostics. The default wait budget MUST be finite and configuration-backed. Tests SHOULD wait for stable workbook row or row-history item identity when identity matters, not only for a visible row count or visible text. Helper diagnostics SHOULD identify expected row IDs, mounted row IDs, row-history item refs when the helper receives them, surface, and the failing selector class without reclassifying ordinary test assertions away from `failure_class=product`.
Verified by: TH-HARNESS-AC-021

**TH-HARNESS-REQ-104**
Authoritative phase-manifest evidence names and titles MUST include the exact row identity they claim, using either the public hyphenated form such as `U-8-10` or the runner-safe underscore form such as `U_8_10`. This applies to `symbol`, every member of `symbols[]`, `title`, and every member of `titles[]` when those fields are used by an authoritative row. Supplemental and support-only evidence MAY omit authoritative row identity only when it does not claim authoritative coverage. Rendered ledgers, schedules, and other generated companions remain downstream of the manifest and MUST NOT become alternate traceability owners.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

**TH-HARNESS-REQ-105**
Vitest is an executable runner, not an evidence tier. Frontend-unit phase selection MUST select every phase-manifest row whose `runner` is `vitest`, whose `coverage` and `execution_dependency` match the selected slice, and whose phase matches the selected manifest. Selection, derived summaries, manifest-aware classification, and residual exclusion MUST NOT narrow Vitest rows to the manifest `unit` section. A Vitest row in `integration`, `e2e`, or a later evidence section remains selected by the same frontend-unit contract when its execution dependency is `frontend_unit`.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

Selected `frontend-unit` phase and target summaries MUST derive pass/fail
status from selected Vitest scenarios, manifest verification, row accounting,
and artifact validation. When Vitest emits a runner JSON report with
`success=true`, `numFailedTests=0`, and `numFailedTestSuites=0`, and selected
manifest/accounting validation finds no failures, the selected `frontend-unit`
summary MUST pass even when the raw runner process status is nonzero or the
runner report also contains unselected scenarios. A selected `frontend-unit`
summary MAY fail only when retained evidence identifies a selected test failure,
missing runner report, watchdog timeout, manifest mismatch, frontend row
accounting failure, test-accounting failure, or artifact validation failure.

Vitest wrappers that receive SIGINT or SIGTERM while an active child process is running MUST forward the signal to the active child process group, wait through the bounded watchdog grace interval, kill any remaining child process group after that grace, redact and retain stdout/stderr logs, and return the conventional interrupted exit code: `130` for SIGINT and `143` for SIGTERM. If interruption prevents `runner.json` from being written, test-output MUST classify the phase as `failure_class=interrupted` and `failure_reason=cancelled_or_interrupted`; absence of `runner.json` in that case MUST NOT be reclassified as `artifact_error`.

**TH-HARNESS-REQ-106**
Playwright authoritative phase-manifest rows MAY declare either `title` or `titles[]`. When `titles[]` is used, each listed title is an authoritative executable scenario for the same row ID. Playwright selection, grep generation, list verification, run verification, manifest-aware accounting, browser shard planning, and ledger rendering MUST flatten those titles as independently required scenarios while retaining the row ID as the ownership unit.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

Browser shard manifest verification MUST carry exact selected scenario identity for each selected Playwright shard entry. The shard-selected artifact MUST use `cartulary.playwright_manifest_selection.v1` and MUST identify every selected executable scenario by row ID, normalized Playwright file, and exact title. Per-shard run verification MUST compare only the selected scenario titles for that shard and phase. Browser shard verification MUST use the selected-tests artifact when present; `CARTULARY_MANIFEST_SELECTED_IDS` is not a current browser-shard verification fallback, and row ID alone MUST NOT narrow browser verification. Row ID remains the evidence owner, and frontend row accounting remains aggregate/title-based under `cartulary.frontend_row_accounting.v4`.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

**TH-HARNESS-REQ-107**
Authoritative phase-manifest rows MAY declare `claim_status` with exactly one of `implemented`, `blocked`, or `not_applicable`. Generated ledgers and phase-slice summaries MUST expose claim status separately from execution pass/fail status. A phase with any authoritative `blocked` row MUST report an incomplete aggregate claim status even when all selected harness work exits successfully.

A `blocked` phase-manifest row is a traceability placeholder for owner-known future evidence, not executable conformance. The row MUST still declare closed metadata, claim text, out-of-scope text, evidence class, layer, target routing, and a stable row ID, but phase-map validation MUST NOT require the referenced source file, Go symbol, Vitest title, or Playwright title to exist until the row is promoted to `implemented`. Phase slices MUST exclude `blocked` rows from runnable work units and MUST retain the incomplete aggregate claim status.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

Executable frontend phase-slice commands MUST select active current frontend phases and implemented frontend rows only. A selected frontend row that is unknown, belongs to a later phase, is blocked, is stale, is not implemented, is retired, or comes from a non-active phase MUST fail as harness usage, not as a product failure.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

**TH-HARNESS-REQ-108**
Browser E2E helpers that perform a mutating UI action and then drive another action that depends on the committed result MUST wait for the server success response and for the rendered workbook projection to converge on the response's stable row identity before continuing. When the dependent action relies on optimistic concurrency, convergence MUST include the returned `row_version` rendered under the stable row identifier. A visible global save-state label such as `Saved` MAY be asserted after convergence, but it MUST NOT be the only completion predicate for a dependent mutation sequence.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-021

Browser E2E tests that need the authenticated incident directory MUST navigate to that surface explicitly before asserting directory UI. Tests MUST NOT assume raw authenticated `/` always renders the incident directory, because Core 01 `REQ-01-580` makes raw root cardinality-sensitive and requires the sole visible incident to auto-open.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-021

**TH-HARNESS-REQ-115**
Browser E2E helpers that select a workbook surface through a menu, popover, selector, or tab strip MUST use stable selector builders rooted in canonical `view_schema_id` or `sheet_ref` identity. Such a helper MUST use bounded retries, reacquire locators after every render-sensitive action, require a single target option before selecting, and after selection converge on the active workbook shell surface signal, the canonical direct `view_schema_id` URL representation when selecting a base surface, and the target grid shell before returning. Final helper diagnostics MUST include the requested target identity, current URL, active shell surface, menu-open state, visible candidate target identities, and final retry error. This requirement governs browser-test synchronization only and MUST NOT define product behavior beyond the owning product specifications.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-021

**TH-HARNESS-REQ-109**
Frontend readiness rows whose identifiers use the `FE-*` namespace MUST be claimed through `tools/frontend_phase_maps/*.json` files that validate as `cartulary.frontend_phase_test_map.v4`. `frontend-unit`, browser E2E, visual, and accessibility targets MUST NOT claim `FE-*` row coverage by filename inference, by Playwright/Vitest title text alone, or by generated ledger text. Browser-backed frontend rows MUST declare `scenario_titles[]` in the frontend phase map, and implemented browser-backed rows that require scenario-title closure MUST use FE-owned Playwright titles prefixed by the owning row ID, MUST NOT reuse authoritative base phase Playwright titles, and MUST NOT share the same Playwright title with another frontend row. Frontend-aware targets MUST retain the target artifact `frontend-row-accounting.json` with schema ID `cartulary.frontend_row_accounting.v4` whenever the target maps frontend rows or an explicit slice scope disables frontend row accounting. The artifact MUST include `accounting_scope` with mode `active_target`, `selected_rows`, or `disabled`. Standalone broad frontend-aware targets MUST use `active_target` and enforce implemented rows from active frontend phases only. Whole `PHASE_NAMESPACE=frontend PHASE=FE-P<N>` and explicit `ROWS=<FE-row-id,...>` frontend slices MUST use `selected_rows` and enforce only selected implemented frontend row IDs from active phases through `FE-P<N>`. Base `PHASE=phaseN` slices MUST use `disabled` for frontend-aware child target summaries and MUST NOT fail on non-selected `FE-*` rows. Target and tool-run summaries MUST reference `frontend-row-accounting.json` through artifact refs rather than duplicating row accounting under `extensions["cartulary.frontend_row_accounting"]`. Base `manifest-selected-tests.json` artifacts remain base phase inventory and MUST NOT be treated as frontend row closure evidence. Generated frontend ledgers are downstream renderings of the `cartulary.frontend_phase_registry.v5` registry and `cartulary.frontend_phase_test_map.v4` maps; they MUST NOT become alternate row owners. Historical implementation plans and handoffs are non-closure evidence when they conflict with the live registry, maps, ledgers, or retained row-accounting artifacts. FE visual and accessibility rows MUST NOT use `product_conformance`; they remain design-direction or implementation-support evidence unless an explicit Core 05 `claim_publication_boundary` route is declared. If a target exits successfully while an implemented, scenario-backed frontend row inside its accounting scope is not closed, target summary generation MUST fail the target with `failure_class="harness"` and `failure_reason="frontend_row_accounting"`. Retained `cartulary.frontend_row_accounting.v1`, `cartulary.frontend_row_accounting.v2`, and `cartulary.frontend_row_accounting.v3` artifacts are diagnostic-only for old runs and MUST NOT close current `frontend_phase_test_map.v4` rows or satisfy `release-readiness-evidence`.

`frontend-evidence-audit` MUST audit an explicit frontend phase by reading the live frontend registry, the selected frontend phase map, and retained `cartulary.frontend_row_accounting.v4` artifacts from the supplied broad `check`, support, visual, accessibility, and measurement roots. It MUST verify current registry, guide, and phase-map digests before accepting row closure. It MUST fail when any implemented selected-phase row is missing closure in a target whose frontend phase-map target object declares `required_for_closure=true`. A passing broad `check` root MUST NOT be treated as proof that `browser-e2e-support`, `browser-e2e-visual`, `browser-e2e-a11y`, or `browser-e2e-measurement` ran; those explicit targets require their own retained roots when the selected frontend phase requires them.

Retained roots for `frontend-evidence-audit` MUST be acquired before the audit is reported as complete. A caller MAY provide existing retained roots only when Make-owned retained-run inspection can identify the root, target, status, and relevant row-accounting artifacts. Otherwise the caller MUST produce fresh roots through Make-owned public targets, retain the broad `check` root separately from the support, visual, accessibility, and measurement roots, and pass those exact roots with `CHECK_RESULTS_DIR`, `BROWSER_SUPPORT_RESULTS_DIR`, `BROWSER_VISUAL_RESULTS_DIR`, `BROWSER_A11Y_RESULTS_DIR`, and `BROWSER_MEASUREMENT_RESULTS_DIR` when the selected frontend phase has implemented required rows for those targets. Missing retained roots are a validation blocker, not permission to infer closure from phase names, task-surface rows, generated ledgers, or historical evidence.

`release-readiness-evidence` MUST aggregate evidence from the current run root only. It MUST read current target summaries, phase summaries, `cartulary.frontend_row_accounting.v4` artifacts, accessibility and visual artifacts, release artifacts, deployable-shape evidence, and harness-contract outputs, then retain `<run-root>/release-readiness-evidence/release-readiness-evidence.json` with schema ID `cartulary.release_readiness_evidence.v1`. Each evidence record MUST include `evidence_id`, `source_target`, `schema_id`, `owner_refs[]`, `evidence_class`, `conformance_effect`, `claim_publication_effect`, `release_gate_effect`, `run_root`, `artifact_refs[]`, and `status`. `release-check` MUST run the mapped frontend readiness children `browser-e2e-support`, `browser-e2e-visual`, and `browser-e2e-a11y` before `release-readiness-evidence`. The aggregate MUST fail when required current-run release-readiness evidence is missing, stale, blocked, failed, or uses unsupported legacy frontend row-accounting shapes.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-022

**TH-HARNESS-REQ-110**
Authoritative phase manifests MAY declare `profile_claims[]` metadata for extension-profile evidence routing. Each row MUST declare `profile_id`, `claimed`, `claim_ac_id`, `required_ac_ids[]`, `direct_evidence_ids[]`, and `aggregate_ac_ids[]`. When `claimed=true`, every `direct_evidence_ids[]` value MUST name an authoritative phase row whose `claim_status` is `implemented`; aggregate claim ACs alone MUST NOT satisfy a claim. This metadata routes evidence only and MUST NOT redefine Core profile behavior.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-013, TH-HARNESS-AC-016

**TH-HARNESS-REQ-111**
Make-owned frontend dependency installation MUST use the pinned repo-local Node and pnpm toolchain, MUST bind pnpm's content-addressable store to the repo-local `.pnpm-store` path through project configuration, MUST run without requiring a TTY or interactive confirmation, and MUST use a frozen lockfile. `frontend-install` is an install/readiness target, not a dependency-update target; if `pnpm-lock.yaml` is out of sync with workspace manifests, the target MUST fail with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2` rather than mutating the lockfile. A package-manager repair that purges and recreates repo-local `node_modules` is allowed only as part of this non-interactive install contract and only for repo-local workspace dependency roots.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-014, TH-HARNESS-AC-023

### 5.1 Precedence

| Precedence | Source                                 | Rule                                                                                                   |
| ---------: | -------------------------------------- | ------------------------------------------------------------------------------------------------------ |
|          1 | Make-owned wrapper CLI flags           | Highest priority only for flags explicitly declared by the target.                                     |
|          2 | Make command-line variables            | `VAR=value make target` overrides inherited environment for the same canonical variable.               |
|          3 | Exported environment inherited by Make | Accepted only for variables declared in the configuration table.                                       |
|          4 | Target manifest values                 | Source inputs for scheduler and target behavior, not caller overrides.                                 |
|          5 | Config files                           | Apply only to the application-under-test runtime unless the variable table declares a harness binding. |
|          6 | Hardcoded harness defaults             | Used only when all higher layers omit the value.                                                       |

### 5.2 Configuration Algorithm

```text
resolve_harness_config(target, raw_make_vars, raw_env, wrapper_cli_args):
  assert target is a public Make target or a Make-owned wrapper target
  declared = global_configuration_table + per_target_input_registry[target]
  resolved = empty map

  reject undeclared wrapper CLI flags
  reject caller overrides of manifest/internal fields
  reject undeclared public harness Make variables supplied on the Make command line

  for each declared variable in stable table order:
    candidates = [
      wrapper_cli_args value if variable has declared CLI binding,
      raw_make_vars value if supplied on the Make command line,
      raw_env value if exported before Make invocation,
      manifest value if variable has declared manifest binding,
      config_file value if variable has declared config-file binding,
      hardcoded default
    ]

    select the first candidate whose layer is allowed for the variable
    apply the variable's empty-string rule
    normalize the selected value
    validate the normalized value
    if validation fails:
      emit configuration_error summary when the target has a summary layer
      fail before child work with exit code 2
    record selected value, source layer, and normalized value

  ignore undeclared inherited environment variables
  strip undeclared public harness variables from child process environments
  emit resolved values required by Section 8 summaries
```

### 5.3 Per-Target Input Registry

**TH-HARNESS-REQ-112**
Every public target MUST declare a closed per-target input contract. This NLSpec owns the current public target-local input contract. `tools/execution_topology_manifest.json` and the generated `tools/task_surface_manifest.json` mirror MUST use schema `cartulary.task_surface_manifest.v15` or a later adopted schema and MUST contain `input_contract` for every row with `target_class="public"`, but those manifests are mirrors of the closed contract below and MUST NOT independently widen, narrow, or reinterpret a current public target's accepted inputs.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-002, TH-HARNESS-AC-027

Each `input_contract` MUST contain:

| Field | Required value |
| --- | --- |
| `undeclared_make_command_line` | `usage_error` |
| `undeclared_inherited_env` | `ignore` |
| `inputs[]` | Stable ordered array of accepted target-local inputs. Empty array means the public target accepts no target-local Make variables. |

Each `inputs[]` row MUST contain `name`, `binding`, `allowed_sources`, `required`, `type`, `default`, `empty_string`, `normalization`, `invalid_reason`, `summary_emission`, and `child_forwarding`. Rows MAY additionally contain bounded type metadata such as `values`, `min`, or `max`.

| Row field | Meaning |
| --- | --- |
| `name` | Uppercase Make variable name accepted by this target. |
| `binding` | Public invocation binding. The current profile accepts `make_variable`; a later profile MAY add wrapper CLI bindings only when Section 5.2 precedence remains preserved. |
| `allowed_sources` | Subset of `make_command_line`, `environment`, `makefile_default`, `internal_default`, and `manifest`. A source not listed for the row MUST NOT supply the value. |
| `required` | Whether omission after all allowed sources is a usage/configuration failure. |
| `type` | One of the Section 5.3 type tokens. |
| `default` | Default value or `null`; defaults are valid only when their source is declared. |
| `empty_string` | One of `invalid`, `omitted`, or `false`. |
| `normalization` | One of `none`, `trim`, `trim_lowercase`, or `path_token`. |
| `invalid_reason` | `usage_error` for caller selection mistakes or `configuration_error` for invalid paths, retained evidence, internal state, or manifest-derived configuration. |
| `summary_emission` | One of `none`, `value`, `redacted_value`, or `source_and_value`. |
| `child_forwarding` | One of `none`, `argv`, `runtime_env`, or `argv_and_runtime_env`; undeclared public harness inputs MUST NOT reach child environments. |

The closed target-local public input set in the current profile consists only of documented uses of `ROLE`, `PHASE`, `PHASE_NAMESPACE`, `TARGET`, `RESULTS_DIR`, frontend retained evidence root selectors, `ALLOW_OLDER_RESULTS_DIR`, `RUN_ID`, `DETAIL`, `JSON`, fixture report limits, duration-maintenance knobs, scheduler timing knobs, destructive-safety controls, and the explicit Govulncheck database override. A public target accepts one of these names only when it appears in the normative input matrix below.

`frontend-fallow-static` accepts no target-local Make variables in the current Fallow static profile. A future changed-code audit base such as `FALLOW_CHANGED_SINCE` MUST be added to this registry before it becomes public input.

**TH-HARNESS-REQ-114**
Every public target's accepted target-local input set is closed by the normative input matrix. A public target that is not listed in the matrix accepts no target-local Make variables beyond the global variables in Section 5.5. A grouped row is valid only when every listed target has identical `type`, `default`, `allowed_sources`, `required`, `empty_string`, `normalization`, `values` or `min`/`max` bounds, `invalid_reason`, `summary_emission`, and `child_forwarding`.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-029

| Target(s) | Input | Type | Required | Allowed sources | Default | Omission behavior | Empty-string behavior | Normalization | Values/bounds | Invalid behavior | Summary emission | Child forwarding |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `db-reset`, `services-down`, `object-store-reset`, `clean`, `distclean` | `CARTULARY_CLEANUP_DRY_RUN` | `exact_1_bool` | no | Make command line, environment, Makefile default | `false` | false | false | `trim` | exact `1` means true | `usage_error`, exit `2` | value | runtime env |
| `db-reset` | `CARTULARY_DESTRUCTIVE_CONFIRM` | `enum` | no | Make command line only | none | omitted allowed only for dry-run | invalid | `trim` | `db-reset` | `usage_error`, exit `2` | value | none |
| `object-store-reset` | `CARTULARY_DESTRUCTIVE_CONFIRM` | `enum` | no | Make command line only | none | omitted allowed only for dry-run | invalid | `trim` | `object-store-reset` | `usage_error`, exit `2` | value | none |
| `agent-finalize` | `ALLOW_OLDER_RESULTS_DIR` | `exact_1_bool` | no | Make command line, environment, Makefile default | none | older retained root rejected | false | `trim` | exact `1` means true | `usage_error`, exit `2` | value | runtime env |
| `agent-finalize` | `RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | none | actions that require retained evidence are not selected | omitted | `path_token` | existing retained full warm `check` run root when supplied | `usage_error`, exit `2` | value | runtime env |
| `frontend-evidence-audit` | `PHASE_NAMESPACE` | `phase_namespace` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `trim` | `frontend` | `usage_error`, exit `2` | value | runtime env |
| `frontend-evidence-audit` | `PHASE` | `phase_id` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `trim` | frontend phase ID `FE-P<N>` | `usage_error`, exit `2` | value | runtime env |
| `frontend-evidence-audit` | `CHECK_RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | none | omitted unless selected phase requires broad `check` closure | omitted | `path_token` | existing retained full warm `check` run root | `usage_error`, exit `2` when required and missing | value | runtime env |
| `frontend-evidence-audit` | `BROWSER_SUPPORT_RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | none | omitted unless selected phase requires support closure | omitted | `path_token` | existing retained `browser-e2e-support` run root | `usage_error`, exit `2` when required and missing | value | runtime env |
| `frontend-evidence-audit` | `BROWSER_VISUAL_RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | none | omitted unless selected phase requires visual closure | omitted | `path_token` | existing retained `browser-e2e-visual` run root | `usage_error`, exit `2` when required and missing | value | runtime env |
| `frontend-evidence-audit` | `BROWSER_A11Y_RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | none | omitted unless selected phase requires accessibility closure | omitted | `path_token` | existing retained `browser-e2e-a11y` run root | `usage_error`, exit `2` when required and missing | value | runtime env |
| `frontend-evidence-audit` | `BROWSER_MEASUREMENT_RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | none | omitted unless selected phase requires measurement closure | omitted | `path_token` | existing retained `browser-e2e-measurement` run root | `usage_error`, exit `2` when required and missing | value | runtime env |
| `task-surface-report` | `TASK_SURFACE_REPORT_ARGS` | `task_surface_report_args` | no | Make command line, environment, Makefile default | none | compact default report | omitted | `trim` | empty, `--all`, `--check`, `--check --all`, `--all --check` | `usage_error`, exit `2` | value | argv |
| `task-guide` | `ROLE` | `enum` | no | Make command line, environment, Makefile default | none | default role/phase overview selected by `task-guide` | omitted | `trim` | `local-dev`, `feature-dev`, `phase-author`, `ci-investigator`, `release` | `usage_error`, exit `2` | value | argv |
| `task-guide` | `PHASE` | `phase_id` | no | Make command line, environment, Makefile default | none | all applicable phases for selected role | omitted | `trim` | Section 10.1 phase grammar | `usage_error`, exit `2` | value | argv |
| `task-guide`, `phase-slice`, `service-backed-slice`, `explain-phase` | `PHASE_NAMESPACE` | `phase_namespace` | no | Make command line, environment, Makefile default | `base` | `base` | omitted | `trim` | `base`, `frontend` | `usage_error`, exit `2` | value | argv |
| `phase-slice`, `service-backed-slice` | `ROWS` | `frontend_row_ids` | no | Make command line, environment, Makefile default | none | omitted selects whole executable phase behavior | omitted | `trim` | comma-separated `FE-*` row IDs from frontend phase maps | `usage_error`, exit `2` | value | argv |
| `task-guide`, `phase-slice`, `service-backed-slice`, `fixture-report`, `explain-phase`, `explain-target` | `JSON` | `exact_1_bool` | no | Make command line, environment, Makefile default | `false` | human output | false | `trim` | exact `1` means JSON target-local output | `usage_error`, exit `2` | value | argv |
| `frontend-unit` | `VITEST_MAX_WORKERS` | `positive_integer` | no | Make command line, environment, Makefile default | `4` | `4` | invalid | `trim` | `1..16` | `usage_error`, exit `2` | value | runtime env |
| `phase-slice`, `service-backed-slice`, `explain-phase` | `PHASE` | `phase_id` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `trim` | Section 10.1 phase grammar | `usage_error`, exit `2` | value | argv |
| `target-plan`, `target-plan-json`, `fixture-report`, `explain-run`, `scheduler-event-order-drift`, `scheduler-summary-timing-drift` | `TARGET` | `target_name` | no | Make command line, environment, Makefile default | none | all target rows accepted by the target | omitted | `trim` | public or scheduler target name present in the task-surface manifest | `usage_error`, exit `2` | value | argv |
| `fixture-report` | `RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | `.cartulary/test-results` | default result root | omitted | `path_token` | existing result root or retained run root | `usage_error`, exit `2` | value | argv |
| `fixture-report`, `explain-run` | `RUN_ID` | `run_id` | no | Make command line, environment, Makefile default | none | latest retained run under selected result root for human investigation only | omitted | `trim` | Section 6.2 `run_id_v1` | `usage_error`, exit `2` | value | argv |
| `fixture-report` | `FIXTURE_THRESHOLD_MS` | `positive_integer` | no | Make command line, environment, Makefile default | `30000` | `30000` | omitted | `trim` | `1..999999999` | `usage_error`, exit `2` | value | argv |
| `fixture-report` | `FIXTURE_TOP` | `positive_integer` | no | Make command line, environment, Makefile default | `5` | `5` | omitted | `trim` | `1..999999999` | `usage_error`, exit `2` | value | argv |
| `explain-run`, `go-test-duration-baselines`, `browser-e2e-duration-baselines`, `service-backed-make-target-duration-baselines`, `harness-smoke-duration-baselines` | `RESULTS_DIR` | `result_selector` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `path_token` | existing result root or retained run root | `usage_error`, exit `2` | value | argv |
| `explain-run` | `DETAIL` | `enum` | no | Make command line, environment, Makefile default | `summary` | `summary` | omitted | `trim` | `summary`, `children`, `logs`, `progress`, `accounting` | `usage_error`, exit `2` | value | argv |
| `explain-target` | `TARGET` | `target_name` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `trim` | public or scheduler target name present in the task-surface manifest | `usage_error`, exit `2` | value | argv |
| `explain-target` | `DETAIL` | `enum` | no | Make command line, environment, Makefile default | `summary` | `summary` | omitted | `trim` | `summary`, `rows`, `artifacts` | `usage_error`, exit `2` | value | argv |
| `go-test-duration-baselines` | `PRUNE_OBSERVED_PACKAGES` | `exact_1_bool` | no | Make command line, environment, Makefile default | `false` | false | false | `trim` | exact `1` means true | `usage_error`, exit `2` | value | argv |
| `go-test-duration-baselines` | `ALLOW_COMMAND_OVERHEAD_DECREASE` | `exact_1_bool` | no | Make command line, environment, Makefile default | `false` | false | false | `trim` | exact `1` means true | `usage_error`, exit `2` | value | argv |
| `go-test-duration-baselines`, `go-test-duration-baseline-coverage`, `go-test-duration-baseline-drift` | `GO_TEST_DURATION_BASELINE` | `path` | no | Make command line, environment, Makefile default | none | omitted caller input lets the Makefile default source supply `tools/go_test_duration_baselines.json` | omitted | `path_token` | filesystem path token | `configuration_error`, exit `2` | value | argv |
| `go-test-duration-baseline-drift`, `browser-e2e-duration-baseline-drift`, `service-backed-make-target-duration-baseline-drift`, `harness-smoke-duration-baseline-drift`, `scheduler-event-order-drift`, `scheduler-summary-timing-drift` | `RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | `current-run` | current retained run when available | omitted | `path_token` | existing result root or retained run root | `usage_error`, exit `2` | value | argv |
| `browser-e2e-duration-baselines`, `browser-e2e-duration-baseline-drift` | `BROWSER_E2E_DURATION_BASELINE` | `path` | no | Make command line, environment, Makefile default | none | omitted caller input lets the Makefile default source supply `tools/browser_e2e_duration_baselines.json` | omitted | `path_token` | filesystem path token | `configuration_error`, exit `2` | value | argv |
| `service-backed-make-target-duration-baselines`, `service-backed-make-target-duration-baseline-drift` | `SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE` | `path` | no | Make command line, environment, Makefile default | none | omitted caller input lets the Makefile default source supply `tools/service_backed_make_target_duration_baselines.json` | omitted | `path_token` | filesystem path token | `configuration_error`, exit `2` | value | argv |
| `harness-smoke-duration-baselines`, `harness-smoke-duration-baseline-drift` | `HARNESS_SMOKE_DURATION_BASELINE` | `path` | no | Make command line, environment, Makefile default | none | omitted caller input lets the Makefile default source supply `tools/harness_smoke_duration_baselines.json` | omitted | `path_token` | filesystem path token | `configuration_error`, exit `2` | value | argv |
| `scheduler-summary-timing-drift` | `SCHEDULER_WARM_CHECK_BUDGET_MS` | `positive_integer` | no | Make command line, environment, Makefile default | `155000` | `155000` | omitted | `trim` | `1..999999999` | `usage_error`, exit `2` | value | argv |
| `scheduler-summary-timing-drift` | `SCHEDULER_WARM_CHECK_BALANCE_RATIO` | `positive_decimal` | no | Make command line, environment, Makefile default | `1.25` | `1.25` | omitted | `trim` | `>=1` | `usage_error`, exit `2` | value | argv |
| `build-operator` | `OPERATOR_BIN` | `path` | no | Make command line, environment, Makefile default | `$(CURDIR)/operator` | build writes the default repo-local operator binary | invalid | `path_token` | concrete build-output file path; not `/`, `.`, `..`, not under protected repo roots, no NUL, no POSIX backslash, not an existing directory, and not a symlink path | `configuration_error`, exit `2` | source_and_value | none |
| `go-vulncheck` | `GOVULNCHECK_DB` | `path` | no | Make command line, environment, Makefile default | none | Govulncheck default DB | omitted | `path_token` | optional Govulncheck vulnerability DB path or endpoint token | `usage_error`, exit `2` | value | runtime env |

`fixture-report` remains a `human_summary` target by default. `JSON=1` selects the target-local diagnostic JSON path and is not equivalent to `CARTULARY_OUTPUT_MODE=machine`. When `JSON=1`, stdout MUST be exactly one `cartulary.fixture_report.v1` JSON object followed by one LF, and stderr follows the Section 7 failure budget for `human_summary` targets. `CARTULARY_OUTPUT_MODE=machine make fixture-report` MUST continue to fail before child work under Section 7.2 unless a later adopted registry row changes the target's output class.

**TH-HARNESS-REQ-113**
Undeclared public harness inputs MUST have one shared result:

| Caller input class | Required behavior |
| --- | --- |
| Undeclared wrapper CLI flag | Reject before child work with `failure_reason=usage_error`, exit `2`. |
| Undeclared public harness Make variable supplied on the Make command line | Reject before child work with `failure_reason=usage_error`, exit `2`. |
| Undeclared inherited environment variable | Ignore for resolution and strip from child process environments. |
| Caller override of manifest/internal fields | Reject before child work with `failure_reason=configuration_error`, exit `2`. |

Manifest and internal fields include at least `TASK_SURFACE_MANIFEST`, `CARTULARY_TASK_SURFACE_MANIFEST`, `EXECUTION_TOPOLOGY_MANIFEST`, `CARTULARY_EXECUTION_TOPOLOGY_MANIFEST`, `SCHEDULER_MANIFEST`, and `CARTULARY_OPERATOR_BIN` when supplied through public Make command-line variables. Script-level environment fallbacks such as broad manifest-path overrides, broad passthrough argument strings such as `VITEST_FLAGS`, or unbounded threshold variables are non-canonical implementation inputs unless a public target row declares a bounded `input_contract` entry.

For public `frontend-unit` evidence, `VITEST_MAX_WORKERS` is the only current public Vitest input. `VITEST_FLAGS` MUST NOT be accepted from the Make command line for `frontend-unit`; inherited `VITEST_FLAGS` from the caller environment MUST be stripped before child Vitest execution and MUST NOT narrow the canonical runner report. Filtered Vitest diagnostics, when needed, are private developer commands outside public harness evidence.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-002, TH-HARNESS-AC-003

| Type token | Valid values |
| --- | --- |
| `enum` | One of the row's `values[]` tokens after normalization. |
| `exact_1_bool` | Exact `1` when true; empty string is false only when the row says `empty_string=false`. |
| `phase_id` | `phaseN` for the base namespace or `FE-P<N>` for the frontend namespace, subject to Section 10. |
| `phase_namespace` | `base` or `frontend`. |
| `target_name` | A target name present in the task-surface manifest. |
| `run_id` | `run_id_v1` from Section 6.2. |
| `result_selector` | Existing result root or retained run-root path accepted by the target. |
| `path` | Filesystem path token; path existence and safety follow the Section 5.3 matrix row's `Values/bounds` and `Invalid behavior` cells unless a later adopted row narrows them. |
| `positive_integer` | Decimal integer greater than zero and inside row bounds when declared. |
| `positive_decimal` | Decimal number greater than zero and inside row bounds when declared. |
| `task_surface_report_args` | Empty string, `--all`, `--check`, `--check --all`, or `--all --check`. |

### 5.4 Empty-String Rules

| Variable family                              | Empty string behavior                                                                |
| -------------------------------------------- | ------------------------------------------------------------------------------------ |
| Output mode                                  | Treated as omitted; default resolution applies.                                      |
| Result root                                  | Invalid.                                                                             |
| Run ID                                       | Invalid.                                                                             |
| Boolean exact-`1` flags                      | Empty string is false.                                                               |
| Integer limits                               | Treated as omitted; default applies.                                                 |
| Required DSN, endpoint, credential, or token | Invalid.                                                                             |
| Optional config path                         | Treated as omitted.                                                                  |
| Comma-separated lists                        | Empty string is an empty list only when the variable row says so; otherwise invalid. |

### 5.5 Configuration Variable Table

| Name or family                                                                                  | Scope                   | Type and valid values                                                                                                 | Default                                                                                   | Allowed sources                                 | Empty-string behavior                                   | Normalization                                                                                                 | Invalid behavior                                                                   | Summary emission                                   |
| ----------------------------------------------------------------------------------------------- | ----------------------- | --------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | -------------------------------------------------- |
| `CARTULARY_TEST_RESULTS_DIR`                                                                    | global                  | `result_root_path_v1` from Section 6                                                                                  | `.cartulary/test-results`                                                                 | Make variable, env, default                     | invalid                                                 | Path normalized by Section 6                                                                                  | `configuration_error`, exit `2`                                                    | normalized path and cleanup scope                  |
| `CARTULARY_TEST_RUN_ID`                                                                         | global                  | `run_id_v1` from Section 6                                                                                            | generated by Section 6                                                                    | Make variable, env, default                     | invalid                                                 | Grammar validation only                                                                                       | `configuration_error`, exit `2`                                                    | run ID and whether generated                       |
| `CARTULARY_OUTPUT_MODE`                                                                         | global                  | `quiet`, `summary`, `ci`, `verbose`, `debug`, `machine`                                                               | resolved by Section 7                                                                     | Make variable, env, default                     | omitted                                                 | lower-case exact token                                                                                        | `configuration_error`, exit `2`                                                    | resolved output mode and source                    |
| `VERBOSE`                                                                                       | global                  | exact `1` means verbose request; any other value false                                                                | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` ignored as false                                                           | boolean source when true                           |
| `CI_VERBOSE`                                                                                    | global                  | exact `1` means CI-output request; any other value false                                                              | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` ignored as false                                                           | boolean source when true                           |
| `CI`                                                                                            | global                  | exact `1` marks CI environment                                                                                        | false                                                                                     | env, default                                    | false                                                   | exact string compare                                                                                          | non-`1` ignored as false                                                           | boolean source when true                           |
| Scheduler resource limits                                                                       | scheduler               | positive integer `1..256` unless resource row declares narrower bound                                                 | resource registry default                                                                 | CLI flag, Make variable, env, manifest, default | omitted                                                 | decimal parse with no separators                                                                              | `configuration_error`, exit `2`                                                    | normalized limit and source                        |
| `--resource-limit name=value`                                                                   | scheduler               | declared resource name and positive integer value                                                                     | none                                                                                      | CLI flag only                                   | invalid                                                 | name exact; value decimal                                                                                     | `usage_error` for malformed flag, `configuration_error` for invalid declared value | normalized override                                |
| `GO`, `GO_CACHE_DIR`, `GO_MOD_CACHE_DIR`, `GOCACHE`, `GOMODCACHE`                               | toolchain               | executable path or filesystem path                                                                                    | Go auto-discovered; external caches `/tmp/cartulary-go-build` and `/tmp/cartulary-go-mod` | Make variable, env, default                     | invalid for paths, omitted for executable               | path normalization by target helper                                                                           | `configuration_error`, exit `2`                                                    | executable path and cache path unless redacted     |
| `NODE_VERSION`, `PNPM_VERSION`, `NODE_RUNTIME_DIR`, `NODE_BIN`, `PNPM`, `COREPACK_HOME`, `PATH` | toolchain               | version token or filesystem path                                                                                      | Node `24.15.0`, pnpm `10.33.0`, repo-local `tmp/node-runtime`                             | Make variable, env, default                     | invalid for paths/versions unless row-specific optional | exact version token; path normalization                                                                       | `configuration_error`, exit `2`                                                    | version and runtime path                           |
| `CARTULARY_READINESS_CACHE_DIR`, `CARTULARY_READINESS_DISABLE_CACHE`, `CARTULARY_FORCE_REINSTALL` | harness cache           | repo-local path for cache dir; exact `1` for disable or force reinstall                                                | `.cache/cartulary/readiness`; false; false                                                 | Make variable, env, default                     | invalid for path; false for boolean flags                | path normalization; exact string compare for flags                                                           | invalid path is `configuration_error`; non-`1` flags are false                     | cache state and record path only                   |
| `CARTULARY_BUILD_CACHE_DIR`, `CARTULARY_BUILD_CACHE_DISABLE`, `CARTULARY_FORCE_REBUILD`          | harness cache           | repo-local path for cache dir; exact `1` for disable or force rebuild                                                  | `.cache/cartulary/build-artifacts`; false; false                                           | Make variable, env, default                     | invalid for path; false for boolean flags                | path normalization; exact string compare for flags                                                           | invalid path is `configuration_error`; non-`1` flags are false                     | cache state and record path only                   |
| `CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR`, `CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE`      | harness cache           | repo-local path for cache dir; exact `1` disables action cache                                                         | `.cache/cartulary/agent-finalize-action-cache`; false                                      | env, default                                    | invalid for path; false for boolean flag                 | path normalization; exact string compare for flag                                                            | invalid path is `configuration_error`; non-`1` flag is false; `CI=1` disables cache | action cache state in finalizer summary            |
| `CARTULARY_EXECUTION_TOPOLOGY_RENDER_CACHE_DIR`, `CARTULARY_EXECUTION_TOPOLOGY_RENDER_DISABLE_CACHE` | harness cache        | repo-local path for cache dir; exact `1` disables render cache                                                         | `.cache/cartulary/execution-topology-render`; false                                        | env, default                                    | invalid for path; false for boolean flag                 | path normalization; exact string compare for flag                                                            | invalid path is `configuration_error`; non-`1` flag is false                       | debug/cache diagnostics only                       |
| `CONFIG_FILE`, `CARTULARY_CONFIG_FILE`                                                          | app runtime             | config file path                                                                                                      | `configs/dev/config.toml` for local/dev/browser targets                                   | Make variable, env, config binding, default     | omitted                                                 | path normalization; `CARTULARY_CONFIG_FILE` wins only inside application runtime when both are passed through | harness invalid path: `configuration_error`; app invalid config: target failure    | path, not file contents                            |
| `TEST_SERVICES_BIN`, `CARTULARY_TEST_SERVICES_BIN`                                              | service suite           | executable path                                                                                                       | `tmp/toolbin/cartulary-test-services`                                                     | Make variable, env, default                     | invalid                                                 | path normalization                                                                                            | `configuration_error`, exit `2`                                                    | normalized path                                    |
| `CARTULARY_OPERATOR_BIN`                                                                         | runtime binary          | scheduler-owned executable path for operator scenario Go tests                                                        | produced by `build-operator` from `OPERATOR_BIN`; current default `operator`              | scheduler/runtime wiring only; not public Make command line | invalid for canonical scheduler-selected operator scenario work | path normalization; existing regular executable file; symlinks rejected | missing, empty, non-regular, non-executable, or caller command-line override is `configuration_error`, exit `2`; build-artifact digest/provenance mismatch is `artifact_error`, exit `11` | source, normalized path, producer target, file digest, build-artifact reference |
| `CARTULARY_TEST_SERVICES_ACTIVE`                                                                | service suite           | exact `1` selects attach mode                                                                                         | owned mode                                                                                | env, Make variable, default                     | false                                                   | exact string compare                                                                                          | non-`1` selects owned mode                                                         | mode only                                          |
| `CARTULARY_TEST_SUITE_ID`, `CARTULARY_TEST_TARGET`                                              | service suite           | non-empty ASCII token; suite ID is 24 lowercase hex in owned mode                                                     | generated in owned mode                                                                   | service manifest, env in attach mode            | invalid in attach mode                                  | exact grammar validation                                                                                      | `configuration_error`, exit `2`                                                    | suite ID, target                                   |
| Postgres attach set                                                                             | service suite           | `CARTULARY_PGTEST_ADMIN_DSN`, `CARTULARY_PGTEST_DSN_TEMPLATE` containing `{database}`, `CARTULARY_PGTEST_TEMPLATE_DB`, optional `CARTULARY_PGTEST_SCHEMA_HASH` | none                                                                                      | env, Make variable                              | invalid                                                 | DSN redacted; template exact placeholder validation; schema hash exact match when supplied                     | partial or malformed set or schema-hash mismatch: `configuration_error`, exit `2`  | redacted DSN, attach mode, schema hash             |
| Postgres fixture policy                                                                         | service suite           | `template_clone`, `transaction`, `package_reset`, `group_clone`, `migration_scratch`; table lists use comma-separated SQL identifiers | target-plan policy; direct helper fallback `template_clone`                                | Make variable, env, manifest, default           | omitted                                                 | lower-case exact token; table identifiers sorted for summaries                                                | unknown policy or bad identifier: `configuration_error`, exit `2`                  | policy, fixture class, schema hash, table count    |
| Object-store S3 attach set                                                                      | service suite           | endpoint, access key, secret key, secure bool through `CARTULARY_S3TEST_*`                                            | none                                                                                      | env, Make variable                              | invalid for required members                            | endpoint normalized; credentials redacted; secure bool exact `true`/`false` or `1`/`0`                        | partial set or invalid bool: `configuration_error`, exit `2`                       | endpoint, secure flag, credential redaction marker |
| `CARTULARY_TEST_SERVICES_WEB_E2E_CLEANUP_WORKERS`                                               | browser/service cleanup | integer `1..16`                                                                                                       | `4`                                                                                       | Make variable, env, default                     | omitted                                                 | decimal parse                                                                                                 | invalid value falls back to `4` and records warning                                | resolved value and warning when fallback used      |
| Compose env                                                                                     | local services          | `CARTULARY_COMPOSE_FILE`, ready timeouts, `OBJECT_STORE_BUCKET`                                                       | `docker-compose.dev.yml`, Postgres `180s`, object-store `120s`, bucket `cartulary`        | Make variable, env, default                     | omitted for optional values                             | path and duration normalization                                                                               | missing Docker/Compose: Section 9 class                                            | non-secret values                                  |
| Browser owned-stack env                                                                         | browser                 | runtime roots, origins, backend/frontend port overrides, built frontend preview artifact                               | dynamic ports; frontend served from `apps/web/dist` by non-watching preview; `build-web` is a first-class prerequisite | Make variable, env, manifest, default           | invalid for required values or missing built frontend artifact | origin values lower-case scheme and host; ports decimal; backend readiness must prove owned process identity through the token-protected test runtime identity route; frontend readiness must report `frontend_mode="preview"` and `frontend_command_kind="vite-preview"`; service-backed frontend auto-allocation uses stage-owned CORS windows, with stateful browser work in `39100-39199` and current non-stateful browser work in `39000-39099`; frontend port selection MUST use `CARTULARY_BROWSER_STAGE` or scheduler session metadata rather than target-name substring matching | config or port collision: `resource_conflict`; missing preview artifact or invalid config: `configuration_error` | origins, ports, runtime root, ownership proof, frontend mode |
| `PLAYWRIGHT_WORKERS`, worker count/index/offset envs                                            | browser                 | positive integers; worker offset `0..1024`                                                                            | Make `3`; shared config fallback `2`; direct isolated offset `0`; scheduled browser groups require scheduler-owned count and offset | Make variable, env, default, scheduler manifest | omitted only for direct isolated browser invocation      | decimal parse                                                                                                 | `configuration_error`, exit `2`                                                    | worker counts and scheduled worker slot range      |
| `VITEST_MAX_WORKERS`                                                                            | frontend unit           | positive integer `1..16`                                                                                              | direct public `frontend-unit` default `4`; scheduled `frontend-unit` default-check override `2` | Make variable, env, default, scheduler manifest | invalid                                                 | decimal parse                                                                                                 | `usage_error`, exit `2` for public target input                                     | worker count and scheduler-owned source            |
| Webserver-backed shard env                                                                      | browser                 | required grep/file values declared by target, plus selected-test artifact path for manifest-aware shard verification   | none                                                                                      | Make variable, manifest                         | invalid                                                 | exact string after JSON/shell decoding; selected tests validated as `cartulary.playwright_manifest_selection.v1` | missing required value: `configuration_error`, exit `2`                            | declared shard IDs as compatibility fallback; selected `(row_id,file,title)` entries when artifact-backed |
| `CARTULARY_ENABLE_TEST_ROUTES`                                                                  | reset/browser           | exact `1` enables test routes                                                                                         | disabled                                                                                  | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` means disabled                                                             | enabled boolean                                    |
| `CARTULARY_TEST_ROUTE_TOKEN`                                                                    | reset/browser           | non-empty opaque string with at least 128 bits entropy                                                                | generated by harness stack when reset route enabled                                       | harness generated, env for attach mode          | invalid                                                 | not normalized                                                                                                | missing/low-entropy token: `configuration_error`, exit `2` before stack use        | redaction token only                               |
| Object-store runtime env                                                                        | app runtime             | `CARTULARY_S3_OBJECT_PRIMARY_*` endpoint, credentials, secure bool, bucket                                            | browser/dev SeaweedFS S3 local values                                                     | Make variable, env, config binding, default     | invalid for required members                            | endpoint normalized; credentials redacted                                                                     | app startup/reset failure according to Section 12                                  | redacted credential fields                         |
| Runtime root envs                                                                               | app runtime             | `CARTULARY__ROOTS__*__PATH` filesystem paths                                                                          | browser stack creates under runtime root                                                  | Make variable, env, config binding, default     | invalid                                                 | path normalization                                                                                            | invalid/unwritable path: `configuration_error` or app startup failure              | normalized path                                    |
| `CARTULARY_HARNESS_REPO_ROOT`, `CARTULARY_HARNESS_SCRATCH_ROOT`, `TMPDIR`                       | harness scratch         | filesystem path                                                                                                       | `${TMPDIR:-/tmp}/cartulary-harness-scratch`                                               | env, default                                    | invalid for explicit scratch                            | path normalization; scratch root must be outside repo                                                         | in-repo scratch root: `configuration_error`, exit `2`                              | normalized scratch root                            |
| `CARTULARY_CLEANUP_DRY_RUN`                                                                     | cleanup                 | exact `1`                                                                                                             | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` false                                                                      | dry-run boolean                                    |
| `CARTULARY_DESTRUCTIVE_CONFIRM`                                                                 | destructive local reset | enum equal to the target name, currently `db-reset` or `object-store-reset`                                            | none                                                                                      | Make command line only                          | invalid when supplied empty; omitted allowed only for dry-run | trim exact token                                                                                              | wrong Make command-line token is `usage_error`; inherited-env-only confirmation is ignored and cannot satisfy reset confirmation; missing token on real reset fails before mutation | selected target token                            |
| `LINT_SHELL_STRICT`                                                                             | lint                    | exact `1`; public Make lint targets force strict blocking behavior                                                    | `1` for public Make targets; raw script fallback may default false                       | Make recipe                                      | false outside public Make                            | exact string compare                                                                                          | public Make target overrides ignored by recipe-owned strict value                   | boolean when true                                  |
| `STATICCHECK_CHECKS`                                                                             | static analysis          | closed; not a public Make target input                                                                                | Staticcheck default fixed by the public target row and wrapper                           | raw script only                                  | not applicable                                           | none                                                                                                          | Make command-line use is `usage_error`, exit `2`                                  | none                                               |
| Gosec rule, flag, and pattern variables (`GOSEC_*`, `GOSEC_TARGETED_*`, `GOSEC_AUDIT_*`)         | security                | closed; not public Make target inputs                                                                                 | curated profiles from this NLSpec and task-surface owner inputs                          | raw script only                                  | not applicable                                           | none                                                                                                          | Make command-line use is `usage_error`, exit `2`                                  | retained security profile metadata                 |
| `GOVULNCHECK_DB`                                                                                 | security                | optional Govulncheck vulnerability DB path or endpoint token                                                          | omitted; Govulncheck default DB                                                           | Make variable, env, default                     | omitted                                                 | path-token validation                                                                                         | invalid value is `usage_error`, exit `2`                                           | value                                              |
| `GOVULNCHECK_FLAGS`, `GOVULNCHECK_PATTERNS`                                                      | security                | closed; not public Make target inputs                                                                                 | `-test -json` flags and authored package roots fixed by the public target row and wrapper | raw script only                                  | not applicable                                           | none                                                                                                          | Make command-line use is `usage_error`, exit `2`                                  | none                                               |

## 6. Result Roots, Run IDs, and Artifact Identity

**TH-HARNESS-REQ-150**
A public Make target that emits retained artifacts MUST compute artifact identity as:

```text
run_root = normalize_result_root(CARTULARY_TEST_RESULTS_DIR) / normalize_run_id(CARTULARY_TEST_RUN_ID)
```
Verified by: TH-HARNESS-AC-003, TH-HARNESS-AC-015

Retained raw telemetry captures from `otel-conformance` MUST be retained artifacts owned by the `otel-conformance` target below the normalized run root. They MUST NOT be written below committed golden directories such as `internal/testutil/golden/otel/`.

### 6.1 Result Root Normalization

```text
normalize_result_root(input):
  if input is omitted:
    input = ".cartulary/test-results"
  reject empty string
  reject NUL
  reject path equal to "/" after lexical normalization
  reject path equal to "." after lexical normalization
  reject any caller-supplied segment equal to ".."
  reject backslash on POSIX conformance hosts
  if relative:
    resolve against repository root
  if absolute:
    allow for artifact writing and set cleanup_scope = "external_or_custom"
  create the directory if missing
  create retained run roots and target artifact directories with owner-only permissions on POSIX conformance hosts
  fail with configuration_error if parent is not writable
  fail with configuration_error if a caller-supplied custom result root is world-writable without sticky bit
```

### 6.2 Run-ID Grammar

```text
run_id = 1*96(ALPHA / DIGIT / "-" / "_" / ".")
run_id MUST NOT equal "." or ".."
run_id MUST NOT contain "/"
run_id MUST NOT contain "\\"
run_id MUST NOT contain whitespace
```

When `CARTULARY_TEST_RUN_ID` is omitted, the wrapper MUST generate:

```text
YYYYMMDDTHHMMSSZ-p<PID>
```

`YYYYMMDDTHHMMSSZ` is the UTC wrapper start time. `<PID>` is the decimal process ID of the Make-owned top-level wrapper.

### 6.3 Collision Rules

| Case                                                | Required behavior                                                                        |
| --------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Omitted run ID                                      | Generate a default run ID.                                                               |
| Caller-supplied run ID path does not exist          | Create it.                                                                               |
| Caller-supplied run ID path exists and is empty     | Reuse it.                                                                                |
| Caller-supplied run ID path exists and is non-empty | Fail before child work with `configuration_error`, exit `2`.                             |
| Generated default run ID collides                   | Append `-n<N>` with the smallest positive decimal `N` that produces a non-existing path. |

### 6.4 Artifact-Proof Rule

**TH-HARNESS-REQ-151**
Retained artifacts prove only the explicit `{result_root, run_id, target}` triple. A newest-run fallback MAY be used for human investigation, but MUST NOT satisfy harness conformance evidence.
Verified by: TH-HARNESS-AC-015

## 7. Output Modes and Machine Output

**TH-HARNESS-REQ-200**
Public Make targets MUST recognize exactly these output-mode tokens: `quiet`, `summary`, `ci`, `verbose`, `debug`, and `machine`. Unknown output modes MUST fail with `configuration_error` and exit `2` before child work. A recognized mode is accepted only when the target's output class allows it. When a recognized mode is not accepted for that output class, the target MUST fail before child work with the Section 7.2 output-class rejection behavior.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-004, TH-HARNESS-AC-005

### 7.1 Output-Mode Resolution

```text
resolve_output_mode(CARTULARY_OUTPUT_MODE, VERBOSE, CI_VERBOSE, CI, target):
  if CARTULARY_OUTPUT_MODE is present and non-empty:
    return exact token after lower-case validation
  if VERBOSE == "1":
    return "verbose"
  if CI_VERBOSE == "1":
    return "ci"
  if target == "ci" or CI == "1":
    return "ci"
  return "summary"
```

`quiet`, `debug`, and `machine` are selected only by `CARTULARY_OUTPUT_MODE`.

### 7.2 Output Class Matrix

| Output class                       | Public targets                                                                                    | `machine` accepted? | `machine` stdout                                        | `machine` stderr                                                                      | Success artifacts                                         | Failure behavior                                                       |
| ---------------------------------- | ------------------------------------------------------------------------------------------------- | ------------------: | ------------------------------------------------------- | ------------------------------------------------------------------------------------- | --------------------------------------------------------- | ---------------------------------------------------------------------- |
| `summary_with_artifacts`           | Leaf, toolchain, non-aggregate build/lint child, drift, browser-stage, and formatting targets with wrapper summaries |                 yes | One `cartulary.tool_run_summary.v3` JSON object plus LF | Empty after wrapper starts; pre-wrapper diagnostics allowed only before JSON emission | Tool-run summary required                                 | Same schema with failure fields and nonzero exit                       |
| `service_summary`                  | `db-up`, `db-reset`, `services-up`, `services-down`, `object-store-init`, `object-store-reset` |                 yes | One `cartulary.tool_run_summary.v3` JSON object plus LF | Empty after wrapper starts                                                            | Tool-run summary and service diagnostics for service-owning rows | Same schema with service failure fields                                |
| `aggregate_summary_with_artifacts` | `test-fast`, `test`, `lint`, `ci`, `release-check`, `build`                                       |                 yes | One `cartulary.tool_run_summary.v3` JSON object plus LF | Empty after wrapper starts                                                            | Aggregate summary plus child references                   | Same schema with primary failure                                       |
| `scheduler_summary_with_artifacts` | `check`, `phase-slice`, `service-backed-slice`                                                    |                 yes | One `cartulary.tool_run_summary.v3` JSON object plus LF | Empty after wrapper starts; no scheduler progress prose                               | Scheduler summary/events and run summary                  | Same schema with scheduler or child failure                            |
| `machine_stdout_json`              | `target-plan-json` and other explicitly declared JSON discovery targets                           |                 yes | One closed target JSON value plus LF                    | Empty on success                                                                      | None unless target declares artifacts                     | Invalid input exits `2`; error JSON only when Section 7.4 declares it |
| `human_summary`                    | `help`, `help-all`, text discovery/explanation targets                                            |                  no | Empty                                                   | Bounded diagnostic allowed on failure                                                 | None unless target row declares diagnostic artifacts      | `machine` rejected as `usage_error`, exit `2`                          |
| `interactive_raw`                  | `dev`                                                                                             |                  no | Empty when `machine` requested                          | Diagnostic allowed                                                                    | None                                                      | `machine` rejected as `usage_error`, exit `2`                          |
| `destructive_human`                | `clean`, `distclean`                                                                              |                  no | Empty when `machine` requested                          | Diagnostic allowed                                                                    | None                                                      | `machine` rejected as `usage_error`, exit `2`                          |

Successful `frontend-fallow-static` `summary` and `quiet` output MUST expose only bounded result and artifact references. Raw Fallow JSON, SARIF, Markdown, stdout, and stderr diagnostics MUST be retained as artifacts rather than emitted in successful summary stdout; `verbose` and `debug` modes MAY expose raw child diagnostics.

### 7.3 Human Output Budgets

| Mode      |             Success stdout budget | Success stderr                                              | Child logs                                       |
| --------- | --------------------------------: | ----------------------------------------------------------- | ------------------------------------------------ |
| `quiet`   |                    At most 1 line | Empty unless failure                                        | No child logs.                                   |
| `summary` |   At most 30 lines and 8192 bytes | Empty                                                       | Retained in artifacts only.                      |
| `ci`      | At most 120 lines and 32768 bytes | Empty unless CI wrapper failure occurs before summary layer | Retained in artifacts; bounded progress allowed. |
| `verbose` |              No fixed line budget | Tool-dependent                                              | May stream child logs.                           |
| `debug`   |              No fixed line budget | Tool-dependent                                              | May stream wrapper telemetry.                    |

### 7.4 Machine Output

**TH-HARNESS-REQ-201**
For every public target whose output class accepts `machine`, stdout MUST be exactly one UTF-8 JSON object followed by one LF, except that `machine_stdout_json` discovery targets MAY emit one closed target JSON value followed by one LF when Section 7.4 defines that value. The JSON payload MAY contain artifact pointers. Stdout MUST NOT be a pointer-only payload, multiple JSON values, scheduler progress prose, child logs, or human summary text.
Verified by: TH-HARNESS-AC-004

**TH-HARNESS-REQ-202**
For every public target whose output class rejects `machine`, setting `CARTULARY_OUTPUT_MODE=machine` MUST fail before child work with `failure_class=config`, `failure_reason=usage_error`, public exit code `2`, empty stdout, and bounded stderr diagnostic.
Verified by: TH-HARNESS-AC-005

For `target-plan-json`, the current closed JSON contract is a UTF-8 JSON array followed by one LF. Each array item MUST be an object and MUST declare `source_family` with exactly one of `backend`, `browser`, or `frontend`.

Backend rows MUST preserve the existing target-plan row fields emitted by the backend target-plan collector: `target`, `service_backed`, `runner_family`, `check_heavy_safe`, `check_service_backed_safe`, `check_isolated_safe`, `canonical_authoritative`, `sharding`, `go_test_parallelism`, `id`, `manifest_phase`, `section`, `coverage`, `execution_dependency`, `evidence_class`, `layer`, `default_check_required`, `default_check_kind`, `default_check_reason_code`, `default_check_reason`, `primary_evidence_owner`, `duplicate_of`, `evidence_delta`, `warm_local_cost_class`, `execution_family`, `execution_label`, `packages[]`, `support_only`, `support_selector`, `raw_selector`, `file`, `package`, `symbols[]`, `scenario_symbols`, `shard_isolation`, `evidence_layer`, `fixture_policy`, `fixture_budget`, and `symbol_fixture_details` when those fields apply to the selected row. `fixture_policy` and `fixture_budget` describe the row-level default; `symbol_fixture_details` is a closed object keyed by declared Go symbol and records the effective `fixture_policy`, effective `fixture_budget`, `reason`, `reason_code`, `proof_kind`, `proof_status`, `proof_ref`, `proof_reason`, and `proof_dirty_tables` for that symbol when present.

Browser rows MUST describe current executable base phase-map Playwright rows and MUST include `target`, `id`, `manifest_phase`, `section`, `coverage`, `execution_dependency`, `evidence_class`, `layer`, `default_check_required`, `default_check_kind`, `default_check_reason_code`, optional `default_check_reason`, `primary_evidence_owner`, `duplicate_of`, `evidence_delta`, `warm_local_cost_class`, `runner_family`, `file`, `titles[]`, `browser_stage`, `browser_group`, `browser_group_kind`, `check_projection_mode`, `check_projection_full_target`, `check_projection_full_target_equivalent`, `scheduled_by_default_check`, `default_check_schedule`, and `default_check_selected_row_ids[]`.

Frontend rows MUST describe current implemented active frontend phase-map rows and MUST include `target`, `id`, `phase_id`, `phase_namespace`, `base_phase_join`, `layer`, `evidence_class`, `claim_status`, `target_evidence_role`, `target_required_for_closure`, `frontend_row_accounting_required`, `default_check_required`, `default_check_kind`, `default_check_reason_code`, optional `default_check_reason`, `primary_evidence_owner`, `duplicate_of`, `evidence_delta`, `warm_local_cost_class`, `scenario_titles[]`, `check_projection_mode`, `check_projection_full_target`, `check_projection_full_target_equivalent`, `scheduled_by_default_check`, `default_check_schedule`, `browser_stage`, `browser_group`, and `default_check_selected_row_ids[]`.

Rows MUST be ordered by `source_family`, target, phase, stage/group where present, and row ID. Unknown `TARGET` input MUST fail with `usage_error`, exit `2`, empty stdout, and bounded stderr; no partial JSON is emitted. Callers that consume `target-plan-json` MUST branch on `source_family`; backend-only consumers MUST ignore non-backend rows rather than assuming every item has backend Go fields.

## 8. Artifact and Schema Contract

**TH-HARNESS-REQ-250**
A public Make-owned command that declares a stable schema ID MUST emit JSON that validates against the matching normative schema attachment before command success. If required artifact validation fails, the public target MUST fail with `artifact_error` or `scheduler_accounting_error` according to Section 9.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-004, TH-HARNESS-AC-025

The following schema IDs are public contracts. Schema file paths are repository attachments, not behavioral owners. A missing file in the current repository snapshot MUST be marked as a future attachment here rather than implied by prose.

| Schema ID                                       | Repository attachment path                                               | Status            | Producer class           | Required validation point                 |
| ----------------------------------------------- | ------------------------------------------------------------------------- | ----------------- | ------------------------ | ----------------------------------------- |
| `cartulary.tool_run_summary.v3`                 | `tools/schemas/cartulary.tool_run_summary.v3.schema.json`                 | present           | Centralized wrappers     | Before wrapper exits.                     |
| `cartulary.fallow_static_summary.v1`            | `tools/schemas/cartulary.fallow_static_summary.v1.schema.json`            | present           | Fallow static target     | Before `frontend-fallow-static` success.  |
| `cartulary.test_phase_summary.v3`               | `tools/schemas/cartulary.test_phase_summary.v3.schema.json`               | present           | Phase handlers           | Before target summary consumes it.        |
| `cartulary.vitest_failure_details.v1`           | `tools/schemas/cartulary.vitest_failure_details.v1.schema.json`           | present           | Vitest wrappers          | Before Vitest phase summaries consume failure diagnostics. |
| `cartulary.test_target_summary.v4`              | `tools/schemas/cartulary.test_target_summary.v4.schema.json`              | present           | Target summary generator | Before aggregate/run summary consumes it. |
| `cartulary.test_run_summary.v6`                 | `tools/schemas/cartulary.test_run_summary.v6.schema.json`                 | present           | Run summary generator    | Before public aggregate success.          |
| `cartulary.same_run_helper_artifact_ref.v1`     | `tools/schemas/cartulary.same_run_helper_artifact_ref.v1.schema.json`     | present           | Run summary generator    | Before an aggregate reports same-run helper artifact reuse. |
| `cartulary.test_accounting_classification.v2`   | `tools/schemas/cartulary.test_accounting_classification.v2.schema.json`   | present           | Test accounting classification validation | During JSON shape checks and before target summary classification consumes the manifest. |
| `cartulary.contract_family_registry.v1`         | `tools/schemas/cartulary.contract_family_registry.v1.schema.json`         | present           | Contract generator family registry | During JSON shape checks and before `tools/contractgen` emits generated contract roots. |
| `cartulary.check_scheduler_summary.v10`          | `tools/schemas/cartulary.check_scheduler_summary.v10.schema.json`          | present           | Check scheduler          | Before scheduler target success.          |
| `cartulary.service_backed_scheduler_summary.v10` | `tools/schemas/cartulary.service_backed_scheduler_summary.v10.schema.json` | present           | Service-backed scheduler | Before scheduler target success.          |
| `cartulary.scheduler_event.v6`                  | `tools/schemas/cartulary.scheduler_event.v6.schema.json`                  | present           | Scheduler                | During scheduler JSONL validation.        |
| `cartulary.scheduler_pressure_summary.v1`       | `tools/schemas/cartulary.scheduler_pressure_summary.v1.schema.json`       | historical_diagnostic | Scheduler reporter retained-run readers | During investigation of retained runs that predate v2. |
| `cartulary.scheduler_pressure_summary.v2`       | `tools/schemas/cartulary.scheduler_pressure_summary.v2.schema.json`       | historical_diagnostic | Scheduler reporter retained-run readers | During investigation of retained runs that predate v3. |
| `cartulary.scheduler_pressure_summary.v3`       | `tools/schemas/cartulary.scheduler_pressure_summary.v3.schema.json`       | historical_diagnostic | Scheduler reporter retained-run readers | During investigation of retained runs that predate v4. |
| `cartulary.scheduler_pressure_summary.v4`       | `tools/schemas/cartulary.scheduler_pressure_summary.v4.schema.json`       | present           | Scheduler reporter       | Before scheduler target success.          |
| `cartulary.fixture_tier_proof.v1`               | `tools/schemas/cartulary.fixture_tier_proof.v1.schema.json`               | present           | Scheduler reporter and fixture-proof validators | Before a retained fixture-tier proof artifact is accepted. |
| `cartulary.phase_slice_plan.v1`                 | `tools/schemas/cartulary.phase_slice_plan.v1.schema.json`                 | present           | Phase-slice planner      | Before phase-slice JSON plan output is accepted and during JSON shape checks. |
| `cartulary.govulncheck_findings.v1`             | `tools/schemas/cartulary.govulncheck_findings.v1.schema.json`             | present           | Govulncheck wrapper      | Before failure classification or target-summary security rollup consumes findings. |
| `cartulary.test_services.lease.v1`              | `tools/schemas/cartulary.test_services.lease.v1.schema.json`              | present           | Service suite            | Before attach or cleanup relies on lease. |
| `cartulary.test_services.lifecycle.v1`          | `tools/schemas/cartulary.test_services.lifecycle.v1.schema.json`          | present           | Service suite            | During service lifecycle JSONL validation. |
| `cartulary.web_e2e_stack.v3`                    | `tools/schemas/cartulary.web_e2e_stack.v3.schema.json`                    | present           | Browser stack            | Before browser target starts Playwright.  |
| `cartulary.browser_startup_diagnostics.v1`      | `tools/schemas/cartulary.browser_startup_diagnostics.v1.schema.json`      | present           | Browser stack            | On browser frontend artifact validation or readiness completion/failure. |
| `cartulary.test.runtime_identity.v1`             | `tools/schemas/cartulary.test.runtime_identity.v1.schema.json`             | present           | Browser stack            | During backend identity readiness probing. |
| `cartulary.test.runtime_reset.v1`               | `tools/schemas/cartulary.test.runtime_reset.v1.schema.json`               | present           | Reset route/wrapper      | Before browser reset success is accepted. |
| `cartulary.test.clock_control.v1`               | `tools/schemas/cartulary.test.clock_control.v1.schema.json`               | present           | Test clock route         | Before a fixed, offset, reset, or state clock-control response is accepted. |
| `cartulary.test.public_error_fault.v1`          | `tools/schemas/cartulary.test.public_error_fault.v1.schema.json`          | present           | Browser stack            | Before an armed public-error fault is accepted. |
| `cartulary.test.network_flow_fault_control.v1`  | `tools/schemas/cartulary.test.network_flow_fault_control.v1.schema.json`  | present           | Network Flow fault-control route | Before an armed Network Flow commit or worker fault is accepted. |
| `cartulary.test.network_flow_randomness_control.v1` | `tools/schemas/cartulary.test.network_flow_randomness_control.v1.schema.json` | present       | Network Flow randomness-control route | Before an armed deterministic Network Flow random stream is accepted. |
| `cartulary.test.network_flow_auth_transition_control.v1` | `tools/schemas/cartulary.test.network_flow_auth_transition_control.v1.schema.json` | present | Network Flow auth-transition control route | Before an armed Network Flow auth-transition control is accepted. |
| `cartulary.test.network_flow_audit_assertion_control.v1` | `tools/schemas/cartulary.test.network_flow_audit_assertion_control.v1.schema.json` | present | Network Flow audit-assertion control route | Before an armed Network Flow audit-count or replay assertion is accepted. |
| `cartulary.fixture_report.v1`                   | `tools/schemas/cartulary.fixture_report.v1.schema.json`                   | present           | Fixture report target    | Before machine JSON is emitted.           |
| `cartulary.network_flow_fixture_manifest.v1`    | `tools/schemas/cartulary.network_flow_fixture_manifest.v1.schema.json`    | present           | Network Flow fixture manifest validator | Before a Network Flow fixture manifest is selected for conformance execution. |
| `cartulary.network_flow_activity_accounting.v1` | `tools/schemas/cartulary.network_flow_activity_accounting.v1.schema.json` | present           | Network Flow structural-accounting validator | During JSON shape checks and before Network Flow generated-contract or Phase 12 drift evidence is accepted. |
| `cartulary.network_flow_timezone_ruleset_provenance.v1` | `tools/schemas/cartulary.network_flow_timezone_ruleset_provenance.v1.schema.json` | present | Network Flow timezone provenance validator | During JSON shape checks and before timestamp fixtures are accepted. |
| `cartulary.agent_finalize_summary.v3`           | `tools/schemas/cartulary.agent_finalize_summary.v3.schema.json`           | present           | Agent finalizer          | Before `agent-finalize` exits.            |
| `cartulary.cache.readiness.v1`                  | `tools/schemas/cartulary.cache.readiness.v1.schema.json`                  | present           | Readiness cache helper   | Before a readiness cache record or retained cache artifact is accepted. |
| `cartulary.cache.build_artifact.v1`             | `tools/schemas/cartulary.cache.build_artifact.v1.schema.json`             | present           | Build-artifact cache helper | Before a build cache record or retained cache artifact is accepted. |
| `cartulary.cache.static_analysis.v1`            | `tools/schemas/cartulary.cache.static_analysis.v1.schema.json`            | present           | Static-analysis cache helper | Before a static-analysis cache record or retained cache artifact is accepted. |
| `cartulary.agent_finalize_action_cache_record.v1` | `tools/schemas/cartulary.agent_finalize_action_cache_record.v1.schema.json` | present         | Agent finalizer action cache | Before an `agent-finalize` action-cache hit is accepted. |
| `cartulary.execution_topology_render_cache.v1`  | `tools/schemas/cartulary.execution_topology_render_cache.v1.schema.json`  | present           | Execution-topology renderer | Before cached render content is reused.   |
| `cartulary.frontend_accessibility_summary.v2`   | `tools/schemas/cartulary.frontend_accessibility_summary.v2.schema.json`   | present           | Browser accessibility target | Before `browser-e2e-a11y` success.    |
| `cartulary.frontend_phase_registry.v5`          | `tools/schemas/cartulary.frontend_phase_registry.v5.schema.json`          | present           | Frontend phase registry validation | During JSON shape checks, phase-map checks, phase-schedule generation, and frontend freshness validation. |
| `cartulary.frontend_phase_test_map.v4`          | `tools/schemas/cartulary.frontend_phase_test_map.v4.schema.json`          | present           | Frontend phase-map validation | During JSON shape checks, phase-map checks, frontend ledger generation, phase-schedule generation, and frontend row accounting. |
| `cartulary.frontend_row_accounting.v4`          | `tools/schemas/cartulary.frontend_row_accounting.v4.schema.json`          | present           | Frontend-aware target summaries | Before target-summary success when frontend rows map to the target or explicit slice scope disables frontend row accounting. |
| `cartulary.frontend_evidence_audit_summary.v1`  | `tools/schemas/cartulary.frontend_evidence_audit_summary.v1.schema.json`  | present           | Frontend retained-evidence audit | Before `frontend-evidence-audit` writes its command-specific summary. |
| `cartulary.release_readiness_evidence.v1`       | `tools/schemas/cartulary.release_readiness_evidence.v1.schema.json`       | present           | Release-readiness aggregation | Before `release-readiness-evidence` success. |
| `cartulary.frontend_visual_fixture_registry.v3` | `tools/schemas/cartulary.frontend_visual_fixture_registry.v3.schema.json` | present           | Frontend visual fixture registry validation | During JSON shape checks, phase-map checks, and visual readiness validation. |
| `cartulary.frontend_claim_publication_review.v1` | `tools/schemas/cartulary.frontend_claim_publication_review.v1.schema.json` | present          | Conditional frontend claim-publication review metadata; no default target emits it | Before any future or explicit frontend claim-review artifact is accepted as Core 05-routed release evidence. |
| `cartulary.otel_conformance_summary.v1`         | `tools/schemas/cartulary.otel_conformance_summary.v1.schema.json`         | present           | OpenTelemetry conformance target | Before `otel-conformance` success. |

Adoption and continued conformance for `cartulary.testing_harness.current.v1` require live repository verification of every row whose `Status` is `present`. Each declared attachment path MUST exist, parse as a JSON schema, reject unknown top-level fields unless the schema declares an explicit extension container, and validate at least one positive fixture and one negative fixture. If any declared path is missing or malformed, the repository MUST either add the attachment and validation fixture or change the row status to `future_attachment`. Historical schemas retained in `tools/schemas` but absent from this table, including older browser stack and accessibility summary schemas, are best-effort diagnostic aids only and MUST NOT be treated as Section 8 current-conformance attachments.

`cartulary.phase_slice_plan.v1` validation MUST include schema validation plus phase-slice semantic validation for stable emitted plan fields: target/mode consistency, work-unit identity uniqueness, dependency completion-key closure, scheduler resource names and claim membership, and total/finalizer counts. Because v1 work-unit objects were introduced with open object shape, that openness is retained as an explicit private extension zone only. Unknown work-unit fields are planner internals, MUST NOT be treated as public contract, and MUST NOT be required by downstream callers unless this NLSpec or a later schema revision explicitly promotes them. Closing the work-unit object shape or making work-unit internals public requires a new owner-approved schema revision rather than a same-ID v1 tightening.

Current conformance MUST be proven from current schema-owned artifacts. Historical artifact inspection is non-conformance troubleshooting and is not an acceptance criterion for Sections 1-17.

**TH-HARNESS-REQ-251**
Schema-owned artifacts MUST be closed by default. Unknown top-level fields are invalid unless the schema declares an explicit extension container.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-025

**TH-HARNESS-REQ-252**
Every retained summary artifact MUST include normalized `result_root`, `run_id`, `run_root`, `target`, `output_mode`, public `exit_code`, primary `failure_class`, primary `failure_reason`, `started_at`, and `completed_at`. Timestamps MUST be RFC3339 UTC strings with non-null values.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-015

Public and machine summaries SHOULD print or serialize `run_root` once per summary and SHOULD express retained artifact references relative to `run_root` when the referenced path is under that directory. Absolute paths remain allowed only for external diagnostics that are not under the retained run root.

**TH-HARNESS-REQ-253**
Every schema-owned artifact MUST include `schema_id`. A schema-owned artifact MAY include `extensions` only when its schema declares that field. When present, `extensions` MUST be an object keyed by reverse-DNS or `cartulary.*` extension keys. Consumers MUST ignore unknown extension keys and MUST NOT derive required behavior from an unknown extension key. Adding a new required top-level member or changing the meaning of an existing member requires a new schema ID. Extension data is supplemental only; any value required for conformance, drift, timing, cleanup, scheduling, or failure classification MUST be a declared schema member.

Supplemental service-backed extension data under `extensions["cartulary.service_backed"]` MUST normalize extension-level `readiness_status`, `teardown_status`, and `leak_status` to `pass`, `fail`, or `unknown`. These extension rollups MUST be derived from canonical service lifecycle artifacts and MUST NOT expose raw lifecycle tokens such as `succeeded`, `cleanup_failed`, or `skipped_no_lease` as pass/fail status fields. The schema-owned scheduler and service artifacts remain authoritative when extension data is absent or `unknown`.

Supplemental security extension data under `extensions["cartulary.security"]` MAY summarize scanner artifacts retained by the current target. For Govulncheck, target and tool-run summaries SHOULD expose `govulncheck.status`, `govulncheck.finding_count`, `govulncheck.blocking_count`, and `govulncheck.blocking_vulnerability_ids` derived only from the current target's schema-valid `cartulary.govulncheck_findings.v1` artifacts. This extension is a convenience rollup; `cartulary.govulncheck_findings.v1` and normalized failure fields remain authoritative for security classification.
Verified by: TH-HARNESS-AC-000

**TH-HARNESS-REQ-254**
Generated-drift replay MUST be driven by a declared scratch-input manifest. The drift checker MUST copy every declared generator runtime input into its scratch tree before invoking generation, including shared harness helper scripts used by generator wrappers. A missing declared scratch input MUST fail as an artifact error with a diagnostic naming the missing path, not as a tool-specific module, import, or Make lookup failure.
Verified by: TH-HARNESS-AC-000

**TH-HARNESS-REQ-255**
`browser-e2e-a11y` MUST emit exactly one retained `cartulary.frontend_accessibility_summary.v2` artifact for a completed accessibility target attempt. The artifact MUST contain implemented `FE-A11Y-*` rows only and MUST be a JSON object with `schema_id`, `phase_rows[]`, `scenarios[]`, `keyboard_matrix[]`, `state_communication_checks[]`, `contrast_checks[]`, `violations[]`, and `artifact_refs[]`. Scenario status fields MUST use only `pass`, `fail`, `missing`, or `skipped`; check result fields MUST use only `pass` or `fail`; nested objects MUST be schema-closed. Blocked frontend accessibility rows MUST NOT appear in this row-evidence artifact.

`browser-e2e-visual` MUST keep current-profile `V-*` visual manifest selection separate from frontend `FE-V-*` readiness selection. Direct `browser-e2e-visual` runs both current-profile base visual rows and frontend visual readiness rows selected from frontend phase-map titles. Base `PHASE=phaseN` slices that invoke `browser-e2e-visual` MUST run only the selected base phase's `V-*` rows and MUST disable frontend visual readiness selection. Frontend namespace slices that invoke `browser-e2e-visual` MUST constrain frontend visual readiness to the selected frontend row accounting scope. Frontend visual readiness selection MUST derive exact Playwright title patterns from `tools/frontend_phase_maps/*.json` rows whose structured target object declares `target_name="browser-e2e-visual"`. Matching screenshots remain implementation-readiness evidence and MUST NOT be inferred from base `V-*` manifest rows, snapshot filenames, generated ledgers, or visual fixture registry text alone.

`browser-e2e-visual-update` MUST use the same visual row selection, service lifecycle, and frontend readiness selection mechanics as direct `browser-e2e-visual`, with Playwright snapshot update mode enabled. The target MUST remain helper-only, MUST NOT be selected by `check`, `test`, `ci`, release gates, phase slices, or service-backed slices, and MUST NOT satisfy validation by itself. Its authored writes are limited to committed Playwright visual goldens under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`. A refresh is complete only after a later `browser-e2e-visual` run passes with screenshot comparisons active.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-022

Frontend visual fixture identity MUST come from `tools/frontend_visual_fixture_registry.json` and schema `cartulary.frontend_visual_fixture_registry.v3`. Current entries MUST declare `capture_scope` as `full_viewport`, `selector`, or `region`. Visual fixture IDs MUST form a contiguous append-only range beginning at `FE-VFIX-01`; the current repository data declares `FE-VFIX-01` through `FE-VFIX-21`. Selector captures MUST use a stable data-attribute selector. Selector-only non-grid specimens MAY declare `scroll_normalization.kind="not_applicable"` with a reason instead of a workbook-grid anchor. A fixture with `no_dynamic_regions=true` MUST keep `dynamic_masks=[]`. Missing or ambiguous frontend visual fixtures MUST block the affected frontend visual rows rather than closing them from generic TODO text, inferred snapshot names, or ad hoc scenario titles.

The current `FE-VFIX-01` frontend visual fixture is the Default Timeline workbook shell. Its closure evidence MUST come from the exact `FE-V-P2-01` scenario title declared in the frontend phase map, and the retained screenshot MUST be Playwright output from the running app-owned workbook shell with browser and operating-system chrome excluded. External concept images, generated mockups, browser-chrome screenshots, or design-reference bitmaps MAY inform review, but they MUST NOT be committed as the frontend golden, cited as row-accounting evidence, or used as an alternate visual fixture source.

Frontend visual support specimens MUST remain separate from default shell evidence. FE-P3 grid-adapter screenshots and FE-P11 token/theme screenshots are implementation-support or design-support specimens only; they MUST NOT close `FE-VFIX-01`, substitute for the Default Timeline workbook shell fixture, or be treated as product-conformance evidence. Current-profile `V-*` rows, including the Timeline default viewport row, remain base visual manifest evidence and MUST NOT close `FE-*` rows unless the frontend phase map explicitly maps the exact scenario title to the frontend row.

Frontend row accounting MUST keep product-conformance and design-direction rows scenario-backed: implemented rows in those evidence classes are closed only by exact row `scenario_titles[]` in the mapped target artifact. An implemented `implementation_support` row MAY close from a passing mapped target with empty `scenario_titles[]` only when the support assertion is target-level; failing or incomplete mapped targets MUST leave that support row blocked by target status rather than closed.

Public Make-owned verification targets MUST treat `unmapped` as unexpected executed test inventory. If a target exits successfully while its own accounting section has `counts.unmapped > 0`, target-summary generation MUST fail the target with `failure_class="harness"` and `failure_reason="test_accounting_unmapped"`. The retained `unmapped` and `unmapped_failed` count fields remain diagnostic fields, but successful public targets require `unmapped=0`. Canonical conformance evidence MUST be mapped through the relevant base phase map or frontend phase map by exact title or symbol. Intentional residual coverage MUST be declared in `tools/test_accounting_classification.json` using schema `cartulary.test_accounting_classification.v2`, a Make target scope, bounded file/title matching, a non-`unmapped` coverage bucket, and an explicit reason. Tests whose titles contain base row IDs (`U-*`, `I-*`, or `E-*`) or frontend row IDs (`FE-*`) MUST either match a live mapped scenario/title exactly or remove the row-like prefix before residual classification.

Frontend claim-publication review metadata is inactive by default, and no default harness target emits `cartulary.frontend_claim_publication_review.v1`. A frontend claim review with `claim_publication_intent="none"` MUST NOT require Core 05 publication predicates. `claim_publication_intent="informative_engineering_measurement"` MAY retain engineering measurement evidence but MUST NOT satisfy claim-bearing publication. `claim_publication_intent="claim_bearing_publication"` MUST route through Core 05-compatible predicate and artifact-bundle metadata before any frontend visual, timing, responsiveness, or fixture-sensitive evidence is accepted as publication-grade. The existing `benchmark-claim-check` target remains the Core 05 benchmark-manifest validator; it MUST NOT run merely because ordinary frontend engineering evidence exists. When invoked with the default benchmark manifest path absent, it reports a no-claim pass and MUST NOT be cited as Core 05 publication evidence.

**TH-HARNESS-REQ-256**
`explain-run` MUST diagnose retained aggregate run roots that contain `run-summary.json` and retained public tool-run roots that contain at least one `<target>/tool-run-summary.json`. Tool-run diagnostics MUST NOT require a synthetic aggregate `run-summary.json`. When a tool-run target also emits a command-specific summary artifact, such as `agent-finalize/finalize-summary.json`, `explain-run` MUST surface a bounded human summary of that artifact and retain `DETAIL=logs` access to target and child logs when `TARGET=<target>` is supplied. `DETAIL=accounting` MUST group retained test inventory by target, coverage bucket, file, row-like title count, and inferred owner source so future phase authors can triage unmapped or residual coverage without scraping raw summary JSON.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-019

**TH-HARNESS-REQ-257**
Cache records and retained cache-state artifacts are schema-owned harness artifacts only for the cache profile that produced them. A readiness or build cache record MUST identify the cache schema, scope, profile ID, state, reason code, key digest, input digest, output digest, record path, cacheable outputs, non-cacheable side effects, and timestamp. An `agent-finalize` action-cache record MUST identify the action ID, command ID, action contract version, input profile, key digest, cache schema, input/output digests, and output paths. An execution-topology render cache record MUST identify the renderer, generator version, Node version, input digest, and rendered content digests.

Cache records MAY be retained in `.cache/cartulary/*` and MAY be referenced by compact retained run-root cache artifacts such as `<target>/*-cache-*.json`. A run-root cache artifact proves only cache behavior for the current target attempt; it MUST NOT substitute for the target's required summary, child log, generated-drift verdict, security-scan output, service lifecycle evidence, or scheduler summary. Investigation tools MAY display cache state, but conformance evidence MUST cite the public target summary and required target artifacts rather than the local cache record.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-015, TH-HARNESS-AC-028

Operator runtime-binary consumers MUST retain a bounded provenance artifact for scheduler-selected operator scenario work. That artifact MUST name `CARTULARY_OPERATOR_BIN` as scheduler-produced, identify the `build-operator` producer target, record the normalized binary path, record the executable file digest, and reference the retained `cartulary.cache.build_artifact.v1` artifact and output digest from the producing target. A valid build-artifact cache hit MAY satisfy the producer target under this section, but the binary consumer MUST NOT be marked as scheduler `reused` merely because the binary was already built.
Verified by: TH-HARNESS-AC-037

### 8.1 Artifact Families

| Artifact family                                      | Producer                                        | Path under run root                                             | Schema policy                                                 | Ordering and nullability                                                              | Retention and cleanup                                        |
| ---------------------------------------------------- | ----------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| Tool-run summary                                     | Centralized wrappers                            | `<target>/tool-run-summary.json` or the summary directory closed by the target's Section 8.1 row | `cartulary.tool_run_summary.v3`                               | Required non-null timestamps, target, exit code, output mode, artifact refs, failures | Retained; removed by cleanup only under default result root. |
| Fallow static reports                                | `frontend-fallow-static`                        | `frontend-fallow-static/fallow/*` and `frontend-fallow-static/fallow-static-summary.json` | `cartulary.fallow_static_summary.v1` for the normalized summary; raw JSON, SARIF, Markdown, stdout, and stderr files are diagnostic-only | Report names, statuses, issue counts, artifact refs, baseline state, and enforcement state in schema-defined order | Retained as run-root artifacts; generated/source roots and Fallow config or baseline inputs are not cleanup candidates. |
| Phase summary                                        | Phase handlers                                  | target phase dirs                                               | `cartulary.test_phase_summary.v3`                             | Stable phase/target/runner/status/count fields                                        | Retained.                                                    |
| Vitest failure details                               | Vitest wrappers                                 | `<target>/raw/<phase>/vitest-failure-details.json`              | `cartulary.vitest_failure_details.v1`                         | Runner JSON reference, stdout/stderr refs, failed owner path, title, message, source, raw messages, diagnostic tags, and first app frame | Retained when a Vitest runner report exists; absent sidecar uses phase-summary fallback. |
| Target summary                                       | Target summary generator                        | `<target>/target-summary.json`                                  | `cartulary.test_target_summary.v4`                            | Child/totals rollups ordered by registry order                                        | Retained.                                                    |
| Run summary                                          | Run summary generator                           | `run-summary.json` or aggregate dir                             | `cartulary.test_run_summary.v6`                               | Work units and artifact dirs ordered deterministically                                | Retained.                                                    |
| Same-run helper artifact reference                   | Run summary generator                           | `_shared/same-run-helper-artifacts/*.json`                      | `cartulary.same_run_helper_artifact_ref.v1`                    | Helper target, producer work-unit ID, declared inputs, producer artifact digests, consumer refs, same-run scope, and non-scheduler-reuse accounting | Retained with the run root; diagnostic only.                 |
| Scheduler summary                                    | Scheduler                                       | `<target>/scheduler-summary.json`                               | Scheduler summary schema by scheduler type                    | Work units by manifest ordinal; resources by registry order                           | Retained.                                                    |
| Scheduler event stream                               | Scheduler                                       | `<target>/scheduler-events.jsonl`                               | `cartulary.scheduler_event.v6`                                | `seq` strictly increases with no gaps                                                 | Retained.                                                    |
| Scheduler progress summary                           | Scheduler reporter                              | `<target>/progress-summary.log`                                 | diagnostic-only                                               | Bounded progress snapshots                                                            | Retained.                                                    |
| Scheduler pressure summary                           | Scheduler reporter                              | `<target>/pressure-summary.json`                                | `cartulary.scheduler_pressure_summary.v4`                      | Closed current-profile fields are defined below; ordering is target, lane, resource, fixture-class, row fixture pressure, execution-family fixture pressure, fixture proofs, fixture-tier proof artifacts, readiness attribution, and slowest-work lexical order after producer timestamps are normalized | Retained.                                                    |
| Govulncheck findings                                 | Govulncheck wrapper                             | `<target>/<phase>/govulncheck-findings.json`                     | `cartulary.govulncheck_findings.v1`                            | Finding IDs and counts in deterministic order; symbol-reachable findings are blocking | Retained with the phase; promoted by target summary as artifact refs and supplemental security extension data. |
| Agent finalizer summary                              | Agent finalizer                                 | `agent-finalize/finalize-summary.json`                          | `cartulary.agent_finalize_summary.v3`                         | Ordered actions, private substeps, skipped work, cache state, updated files, `RESULTS_DIR`, child artifact refs | Retained.                                                    |
| Cache state artifacts                                | Cache helpers and agent finalizer               | `<target>/*-cache-*.json` when emitted; records under `.cache/cartulary/*` | `cartulary.cache.readiness.v1`, `cartulary.cache.build_artifact.v1`, `cartulary.cache.static_analysis.v1`, `cartulary.agent_finalize_action_cache_record.v1`, or `cartulary.execution_topology_render_cache.v1` | Profile ID, cache state/reason, key digest, input digest, output digest, output paths, and record path in schema-defined order | Run-root artifacts retained; default cache records removed only by `make distclean`. |
| Runtime binary provenance                            | Go target runner                                | `_shared/<execution-family>/runtime-binaries.json` when an aggregate declares runtime binaries | diagnostic-only                                               | Runtime binary ID, scheduler-owned consumer env, producer target, normalized path, file digest, build-artifact ref, and output digest | Retained with the shared Go report. |
| Service scope artifacts                              | Service suite                                   | `_shared/test-services/<suite-id>/...`                          | lease and lifecycle schemas required; other service logs diagnostic-only | Lease fields and lifecycle events closed by Section 11                                  | Retained; cleanup may append diagnostics.                    |
| Service lifecycle event stream                       | Service suite                                   | `_shared/test-services/<suite-id>/lifecycle-events.jsonl`        | `cartulary.test_services.lifecycle.v1`                        | `seq` strictly increases; transitions match Section 11.2                               | Retained; not cleanup proof.                                |
| Browser stack metadata                               | Browser stack                                   | browser target support dir                                      | `cartulary.web_e2e_stack.v3`                                  | Origins, ports, frontend preview mode and command kind, runtime root, log paths, process group IDs, readiness timestamps, backend identity proof, frontend ownership proof, startup diagnostic reference, and selected fixture identity where known | Retained for browser target.                                 |
| Browser startup diagnostics                          | Browser stack                                   | browser target support dir                                      | `cartulary.browser_startup_diagnostics.v1`                    | Frontend mode, command kind, selected origins and ports, readiness phase, normalized failure cause when failed, log references, and inotify diagnostics when an ENOSPC watcher failure is detected | Retained for browser target startup pass/fail diagnostics.    |
| Reset response/status/state                          | Reset route/wrapper                             | `reset-boundary/*.json`, `*.status`, `*.state-reset`            | `cartulary.test.runtime_reset.v1` for reset data              | Reset ID, table list, migration/admin flags, object count required                    | Retained for browser target.                                 |
| Test clock-control response                          | Test clock route                               | clock-control transcript or target-owned clock-control dir       | `cartulary.test.clock_control.v1`                             | Clock mode, current RFC3339 timestamp, offset seconds, and fixed timestamp when mode is fixed | Retained only by the target or fixture transcript that controls the clock; never production API evidence. |
| Network Flow fault-control response                  | Network Flow fault-control route                | Network Flow fixture transcript or target-owned fault-control dir | `cartulary.test.network_flow_fault_control.v1`                | Fault ID, exact boundary token, fault kind, optional safe error code, optional correlation key, and `consume_once=true` | Retained only by the target or fixture transcript that arms the fault; never production API evidence. |
| Network Flow randomness-control response             | Network Flow randomness-control route           | Network Flow fixture transcript or target-owned randomness-control dir | `cartulary.test.network_flow_randomness_control.v1`           | Control ID, exact stream token, value kind, value count, remaining count, `consume_once=true`, and `exhaustion="fail_closed"` | Retained only by the target or fixture transcript that arms deterministic fixture randomness; never production API evidence. |
| Network Flow auth-transition-control response        | Network Flow auth-transition-control route      | Network Flow fixture transcript or target-owned auth-transition-control dir | `cartulary.test.network_flow_auth_transition_control.v1`      | Control ID, exact boundary token, transition kind, actor ref, incident ref, resource kind/ref, hidden response kind, optional correlation key, `must_not_disclose_resource=true`, and `consume_once=true` | Retained only by the target or fixture transcript that arms route-time authorization or hidden-resource assertions; never production API evidence. |
| Network Flow audit-assertion-control response        | Network Flow audit-assertion-control route      | Network Flow fixture transcript or target-owned audit-assertion-control dir | `cartulary.test.network_flow_audit_assertion_control.v1`     | Assertion ID, assertion kind, event code, operation ref, actor ref, incident ref, resource kind/ref, baseline count, expected final count, expected replay increment, optional correlation key, and `consume_once=true` | Retained only by the target or fixture transcript that arms exact-count or replay-silence assertions; never product audit evidence by itself. |
| Frontend accessibility summary                       | Browser accessibility target                    | `browser-e2e-a11y/accessibility/frontend-accessibility-summary.json` | `cartulary.frontend_accessibility_summary.v2`                  | Implemented `phase_rows[]`, `scenarios[]`, `keyboard_matrix[]`, `state_communication_checks[]`, `contrast_checks[]`, `violations[]`, and `artifact_refs[]` in schema-defined order | Retained for browser target.                                 |
| Frontend row accounting                              | Frontend-aware target summaries                 | `<target>/frontend-row-accounting.json`                             | `cartulary.frontend_row_accounting.v4`                         | Accounting scope, command ID, map and registry digests, run root, scenario results, row results, and required-target closure | Retained for target; target/tool-run summaries reference this artifact instead of duplicating row details. |
| Release-readiness evidence                           | Release-readiness aggregation                   | `release-readiness-evidence/release-readiness-evidence.json`        | `cartulary.release_readiness_evidence.v1`                      | Evidence records with explicit owner refs, evidence class, product conformance effect, Core 05 publication effect, release-gate effect, run root, artifact refs, and status | Retained for release-readiness target; target/tool-run summaries reference the artifact. |
| Network Flow fixture manifest                        | Network Flow fixture manifest validator         | `fixtures/network-flow/<fixture_id>/manifest.json`                  | `cartulary.network_flow_fixture_manifest.v1`                   | Fixture IDs, owner refs, source files, expected artifacts, transcript files, selectors, per-file SHA-256 values, and aggregate bundle hashes in canonical sorted order | Source fixture roots are committed and immutable after freeze; run-local materializations are retained under the selected target's run root. |
| Generated manifest summaries                         | Generation/drift scripts                        | tool-specific target dirs                                       | JSON schemas declared by generated artifacts                  | Unknown fields rejected where shape tools enforce closure                             | Generated files remain checked in; summaries retained.       |
| Logs                                                 | Shell, Go, scheduler, browser, service wrappers | target log dirs                                                 | diagnostic-only unless producer declares schema               | Logs are text after redaction; empty logs may be omitted                              | Retained unless cleanup removes result root.                 |
| Coverage reports                                     | Go/frontend/test tools                          | tool-specific coverage paths                                    | diagnostic-only                                               | No current schema-owned field contract; retained only as tool diagnostic output       | Removed by `make clean` when under registered paths.         |
| Playwright screenshots, videos, traces, HTML reports | Playwright                                      | Playwright report/test-results dirs                             | diagnostic-only secret-bearing                                | No current schema-owned field contract; retained only as Playwright diagnostic output | Removed by `make clean` when under registered paths.         |
| Visual snapshots and goldens                         | Browser/fixture tools                           | source and tool-specific dirs                                   | validation-only; helper refresh is owned by `browser-e2e-visual-update` | No current schema-owned diagnostic schema contract; helper refresh writes tool-specific committed PNGs only | Refresh is helper-only and is not validation evidence until `browser-e2e-visual` passes. |

**TH-HARNESS-REQ-258**
Every target named by a sequence step's `produces_summary_targets[]` MUST retain `<target>/target-summary.json` in the selected run root before the sequence aggregate emits its run summary or aggregate target summary. A target's `<target>/tool-run-summary.json` remains the wrapper-owned tool-run summary. Command-specific reports retained by the target, such as SeaweedFS compatibility or release-gate reports, MUST NOT substitute for `target-summary.json` when the target is a sequence-produced summary target.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-023, TH-HARNESS-AC-027

Nested scheduler targets MUST expose their scheduler artifacts under their own target directory even when a parent aggregate also references them. `check-service-backed` MUST retain first-class `check-service-backed/scheduler-summary.json`, `check-service-backed/scheduler-events.jsonl`, and `check-service-backed/pressure-summary.json`; the pressure summary MUST report backend/browser lane timing, fixture class counts, resource-claim counts, planned child totals, executed child totals, and slowest child work. Parent `check` artifacts MAY link to those nested artifacts, but investigation tools MUST NOT require callers to mine a large parent scheduler summary to diagnose `check-service-backed`.

In the current profile, `pressure-summary.json` is a required retained diagnostic artifact with schema-owned field closure. It MUST be a JSON object that validates against `cartulary.scheduler_pressure_summary.v4` before scheduler target success. Diagnostics MAY read `cartulary.scheduler_pressure_summary.v1`, `cartulary.scheduler_pressure_summary.v2`, and `cartulary.scheduler_pressure_summary.v3` artifacts from older retained runs for investigation only; new scheduler producers MUST emit v4. Current-profile scheduler work reuse is not adopted: `reused_accounting_counts.reused` MUST be `0`, and the scheduler MUST emit truthful nonnegative `executed` and `skipped` counts where those values are derivable from scheduler records. `readiness_attribution_counts`, `readiness_attribution_duration_ms`, and `readiness_attribution_units` MUST be derived only from scheduler-readable `work_units[].readiness_attribution` metadata. Empty readiness attribution means the selected schedule contains no readiness source metadata to attribute, not that readiness was free. The schema validates harness diagnostics only; the artifact MUST NOT be cited as product-conformance or Core 05 claim-publication evidence.

| Field | Required type | Meaning | Omission/null rule |
| --- | --- | --- | --- |
| `schema_id` | string | Schema-owned producer marker; value is `cartulary.scheduler_pressure_summary.v4`. | MUST be present and non-null. |
| `target` | string | Public or scheduler target whose retained directory contains the artifact. | MUST be present and non-empty. |
| `scheduler_kind` | string | Scheduler family that produced the artifact, using the Section 10.1 closed scheduler-kind values. | MUST be present and non-empty. |
| `status` | string | Final scheduler status, using the same status token as the scheduler summary. | MUST be present and non-empty. |
| `total_work_units` | nonnegative integer | Planned child total after schedule normalization. | MUST be present; zero is allowed only for an empty diagnostic fixture schedule. |
| `completed_work_units` | nonnegative integer | Executed child total observed by the scheduler reporter. | MUST be present and MUST NOT exceed `total_work_units`. |
| `scheduler_total_duration_ms` | nonnegative integer | Scheduler wall-clock duration in milliseconds. | MUST be present. |
| `target_counts` | object | Counts by normalized child target or lane key; keys are non-empty strings and values are nonnegative integers. | MUST be present; an empty object means no child count was attributable. |
| `lane_duration_ms` | object | Backend and browser lane timing by normalized lane key; keys are non-empty strings and values are nonnegative integer milliseconds. | MUST be present; an empty object means no lane timing was attributable. |
| `resource_claim_counts` | object | Aggregate logical resource-claim counts by Section 10.2 resource name. | MUST be present; an empty object means no resource claim was observed. |
| `fixture_class_counts` | object | Counts by closed fixture class: `migration_scratch`, `template_clone`, `package_reset`, `transaction_or_shared_postgres`, or `none`. | MUST be present; absent classes have count `0`. |
| `row_fixture_pressure` | array | Row-level fixture pressure aggregates keyed by `target`, `row_id`, `execution_family`, and effective fixture class, with `work_unit_count` and `duration_ms`. | MUST be present; an empty array means no scheduler work had row attribution. |
| `execution_family_fixture_pressure` | array | Execution-family fixture pressure aggregates keyed by `target`, `execution_family`, and effective fixture class, with `work_unit_count` and `duration_ms`. | MUST be present; an empty array means no scheduler work had execution-family attribution. |
| `fixture_proof_records` | array | Retained fixture-proof diagnostics for candidate row or symbol decisions. Each record MUST name `target`, `row_id`, `execution_family`, optional `symbol`, effective `fixture_policy`, `proof_kind`, `proof_status`, optional `proof_ref`, `reason`, and optional `dirty_tables[]`. | MUST be present; an empty array means no candidate proof diagnostics were emitted for the target. |
| `fixture_tier_proofs` | array | Retained `cartulary.fixture_tier_proof.v1` objects for schema-owned fixture-tier proof decisions accepted by the current scheduler reporter. Each object MUST identify target, phase, row ID, execution family, optional symbol, effective fixture policy, proof kind/status, execution boundary, observed surfaces, reset surface, and final verdict. | MUST be present; an empty array means no retained fixture-tier proof artifacts were emitted for the target. |
| `slowest_work_units` | array | Slowest work-unit diagnostics ordered by descending `duration_ms`, then `label` lexical order. Each item MUST include at least `id`, `label`, `status`, and `duration_ms`. | MUST be present; an empty array is allowed only when no work unit executed. |
| `reused_accounting_counts` | object | Reuse accounting counts; current closed keys are `executed`, `reused`, and `skipped`, each a nonnegative integer. | MUST be present with all three keys; `reused` MUST be `0` until scheduler work reuse is owner-adopted. |
| `readiness_attribution_counts` | object | Readiness attribution counts by scheduler-declared readiness class. Keys are non-empty strings and values are nonnegative integers. | MUST be present; an empty object means the selected schedule has no readiness source metadata to attribute. |
| `readiness_attribution_duration_ms` | object | Duration totals by scheduler-declared readiness class. Keys are non-empty strings and values are nonnegative integer milliseconds. | MUST be present; an empty object means the selected schedule has no readiness source metadata to attribute. |
| `readiness_attribution_units` | array | Unit-level readiness/provisioning records derived from manifest metadata. Each item MUST name `id`, `label`, `timing_role`, `readiness_class`, `duration_ms`, `warm_threshold_ms`, and `warm_status`; `reason` SHOULD be present when declared by the manifest. | MUST be present; empty only when no completed work unit has scheduler-readable readiness metadata. |
| `generated_at` | RFC 3339 UTC string | Time the pressure summary object was generated. | MUST be present and non-empty. |

**TH-HARNESS-REQ-259**
Same-run helper artifact reuse is adopted only for helper/setup artifacts produced earlier in the same retained run root. A consumer MUST retain a `cartulary.same_run_helper_artifact_ref.v1` object before reporting helper reuse. The object MUST name the current `run_id` and `run_root`, the helper target, the producer work-unit ID, declared input artifact refs, producer artifact refs with SHA-256 digests, consumer refs, an output digest, `reuse_scope="same_run_only"`, `accounting_mode` of `derived` or `helper_reused`, and `scheduler_reused=false`.

The producer artifacts referenced by a same-run helper artifact ref MUST resolve under the current run root and MUST exist before the consumer aggregate succeeds. Missing artifacts, paths outside the current run root, malformed refs, digest mismatches, or retained-run references from any prior run MUST fail closed as artifact/configuration errors rather than silently falling back to old evidence. A same-run helper artifact ref MAY reduce duplicate helper/setup work inside one aggregate, but it MUST NOT skip any selected production-conformance row, product test, browser/live-state row, service-backed row, drift verdict, security verdict, cleanup, runtime reset, scratch migration apply, object-store mutation, or destructive-operation safeguard.

Same-run helper artifact refs are not cache records and are not scheduler work reuse. Aggregates MAY expose them in run summaries and `explain-run`, but scheduler summaries and pressure summaries MUST continue to report current-profile scheduler `reused` counts as `0` unless a later NLSpec revision adopts scheduler work reuse separately.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-025, TH-HARNESS-AC-028, TH-HARNESS-AC-048

### 8.2 Network Flow Fixture Manifest Contract

**TH-HARNESS-REQ-262**
Network Flow conformance fixtures MUST use a directory-scoped manifest at `fixtures/network-flow/<fixture_id>/manifest.json`. The `<fixture_id>` directory name and manifest `fixture_id` MUST be identical, MUST use the full `NF-FIX-###-slug` identifier, and MUST NOT be inferred from source filenames, display names, route labels, generated output names, or test titles. Legacy single-file fixture locators are not canonical manifest identity.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-263**
A Network Flow fixture manifest MUST validate as `cartulary.network_flow_fixture_manifest.v1`, MUST be schema-closed at every object boundary, and MUST declare `profile_id="network_flow_activity"`, `manifest_version`, `freeze.status`, `freeze.revision`, `owner_refs[]`, `source_files[]`, `expected_artifacts[]`, `transcript_files[]`, `acceptance_ids[]`, `execution_selectors[]`, `source_bundle_sha256`, and `expected_bundle_sha256`. `source_files[]`, `expected_artifacts[]`, and `transcript_files[]` MUST be ordered by `logical_path` ascending by Unicode code point, and each listed file MUST declare exact byte `size_bytes` and lowercase hex SHA-256 of the committed file bytes.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-264**
The Network Flow source bundle digest algorithm is `network_flow_fixture_bundle_hash_v1`. For each `source_files[]` entry in manifest order, the validator hashes the UTF-8 frame `logical_path`, a NUL byte, lowercase `sha256`, a NUL byte, decimal `size_bytes`, and LF. `source_bundle_sha256` MUST equal the SHA-256 of the concatenated frames. `expected_bundle_sha256` uses the same frame algorithm over `expected_artifacts[]` followed by `transcript_files[]`, preserving each array's manifest order. A missing file, extra unlisted file, digest mismatch, size mismatch, unsorted list, duplicate path, absolute path, symlink, or traversal path MUST fail before product code starts.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-265**
Only `freeze.status="frozen"` Network Flow manifests may close Network Flow conformance evidence. A frozen manifest is append-only by revision: any byte change to source files, expected artifacts, transcript files, owner refs, selectors, or aggregate digests requires a new `freeze.revision` and a tracker entry that names the changed fixture. Draft manifests MAY exist during fixture authoring, but public conformance targets MUST report them as blocked rather than treating draft bytes as current evidence.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-266**
Network Flow fixture execution MUST materialize manifest-listed files into a run-local read-only input workspace under the selected result root. Product tests MUST read the run-local materialization, not mutate the committed fixture directory. Expected artifacts and transcript files are read-only comparison inputs. The runner MUST retain a bounded execution summary that names selected fixture IDs, manifest file SHA-256, source and expected bundle SHA-256 values, materialized input root, produced artifact refs, and comparison status. This summary is harness evidence only and does not define Network Flow product behavior.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-267**
Network Flow fixture manifests route evidence; they do not own domain semantics. `owner_refs[]`, `acceptance_ids[]`, and `execution_selectors[]` MAY cite Network Flow, Core, Graph Projection, or harness rows, but the manifest MUST NOT redefine import parsing, row identity, cursor behavior, graph behavior, indicator binding, authorization, retention, audit occurrence rules, or generated contract shape. If no adopted product owner defines the expected behavior named by a manifest selector, the harness MUST classify the row as blocked or unsupported.
Verified by: TH-HARNESS-AC-049

### 8.3 Agent Finalizer

**TH-HARNESS-REQ-260**
`agent-finalize` is a harness-maintenance finalizer. It refreshes and validates deterministic harness-maintenance artifacts before a caller runs explicit verification. It MUST NOT be described or implemented as a verification gate, test runner, cleanup target, code-generation workflow, migration workflow, release gate, security gate, build gate, browser E2E surface, or benchmark-claim surface.
Verified by: TH-HARNESS-AC-019

**TH-HARNESS-REQ-261**
`agent-finalize` exposes exactly one semantic operation: finalize harness-maintenance artifacts before explicit verification. Its public input surface is `make agent-finalize` with optional `RESULTS_DIR=<run_root>` and the ordinary output-mode controls. Callers MUST NOT select finalizer substeps by child target name.

`agent-finalize` MUST derive its execution plan from the closed finalizer action registry below.

| Action ID | Requires `RESULTS_DIR` | Mutating | Cache eligible | Input profile ID | Action contract version | Required behavior | Allowed output |
| --- | ---: | ---: | ---: | --- | --- | --- | --- |
| `scheduler_drift_validation` | yes | no | yes | `agent_finalize.scheduler_drift_validation.v1` | `v1` | Validate scheduler event ordering and warm-check timing health against the retained run before retained-run mutations are allowed. | Finalizer summary and child summaries. |
| `structure_ledger_refresh` | no | yes | yes | `agent_finalize.structure_ledger_refresh.v1` | `v1` | Refresh phase-ledger and phase-schedule generated artifacts, then verify no unsupported drift remains. | Finalizer summary, child summaries, updated-file list. |
| `schema_shape_validation` | no | no | yes | `agent_finalize.schema_shape_validation.v1` | `v1` | Validate harness-owned JSON shape and schema attachments needed by the finalizer path. | Finalizer summary and child summaries. |
| `duration_baseline_refresh` | yes | yes | yes | `agent_finalize.duration_baseline_refresh.v1` | `v1` | Refresh only advisory harness duration-baseline artifacts from a successful, uncontaminated retained run. | Finalizer summary, child summaries, updated-file list. |
| `duration_baseline_coverage` | no | no | yes | `agent_finalize.duration_baseline_coverage.v1` | `v1` | Verify that required advisory duration-baseline entries exist or are explicitly defaulted. | Finalizer summary and child summaries. |
| `duration_baseline_drift_validation` | yes | no | yes | `agent_finalize.duration_baseline_drift_validation.v1` | `v1` | Validate advisory duration-baseline freshness against the retained run without promoting the baselines to benchmark claims or product performance conformance evidence. | Finalizer summary and child summaries. |

The implementation MAY realize an action by invoking one or more Make targets or scripts. Child target names are not part of the `agent-finalize` public contract unless this NLSpec explicitly promotes them.
Verified by: TH-HARNESS-AC-019, TH-HARNESS-AC-020

`agent-finalize` MAY reuse a prior result for a cache-eligible finalizer action only at the finalizer `action_id` boundary. It MUST NOT cache or identify work by child Make target name, script path, package command, or raw tool invocation. A cache hit MUST prove that the action is cache-eligible, the action's input profile is closed, declared repo inputs and implementation bytes match, relevant tool identity and environment values match, declared output artifacts already exist and match the cached output digest, the cache record validates against the current cache schema, and any required retained-run evidence digest matches after retained-run validation succeeds.

Action-cache use MUST be disabled when `CI=1` unless a later adopted harness revision defines CI cache semantics. It MUST be disabled locally when `CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1`. Cache records MUST live under `.cache/cartulary/agent-finalize-action-cache` by default, MUST validate against `cartulary.agent_finalize_action_cache_record.v1` before reuse, and MUST be treated as local acceleration evidence only, not product verification evidence.

A cache hit MUST mark the action `execution_state="reused"` in `agent-finalize/finalize-summary.json`. It MUST NOT report reused work as zero-duration executed child work. Substeps skipped because of an action-cache hit MUST be marked skipped with a cache-hit skipped reason and MUST NOT require child logs from the current run. The action cache state MUST be reported with closed states `hit`, `miss`, `bypass`, `disabled`, `corrupt`, or `ineligible`, plus a closed reason code and the cache key, input profile, input digest, output digest, action contract version, cache schema ID, and record path when a key was computed.

Cache records that are missing or whose inputs changed MUST miss and run normally. Cache records that are malformed, fail schema/key validation, or disagree with current digests MUST be reported as `corrupt`; they may run normally only when ordinary action execution is safe, and they MUST never produce success by reuse. Missing or changed required outputs MUST miss and run normal refresh or validation behavior.

When `RESULTS_DIR` is set, `agent-finalize` MUST validate the supplied retained run and run retained-run scheduler health before any mutating generated-artifact or duration-baseline refresh. A valid finalizer `RESULTS_DIR` is an existing retained full warm `make check` run root that identifies a successful, uncontaminated run and contains the artifact families required by the selected refresh and warm scheduler checks. The retained root SHOULD be the latest successful full warm `make check` retained run under the same retained-results parent. If a caller intentionally supplies an older valid root, `ALLOW_OLDER_RESULTS_DIR=1` MUST be supplied and the finalizer summary MUST record the override. Without that override, an older retained root MUST fail before mutating refresh work. The retained root MUST contain a passing `check/tool-run-summary.json`, `check/scheduler-summary.json`, and `check/scheduler-events.jsonl`; service-backed-only, phase-slice, browser-only, and other partial run roots MUST NOT be accepted as finalizer retained-run maintenance input. The finalizer MUST reject missing paths, non-directory paths, missing full-check markers, failed retained target summaries, missing required scheduler/target/phase summary families, contaminated service timing evidence, older retained roots without explicit override, and retained evidence that cannot support `scheduler-summary-timing-drift TARGET=check`. Rejection before a mutating refresh MUST fail with `configuration_error` for invalid caller input or `artifact_error` for unsafe retained evidence.
Verified by: TH-HARNESS-AC-019

Duration-baseline refreshes remain advisory harness planning data and MUST NOT become benchmark claims or product performance conformance evidence.
Verified by: TH-HARNESS-AC-019

Before any selected finalizer action can mutate tracked generated or baseline files, `agent-finalize` MUST snapshot or stage the tracked worktree state it is allowed to change. If any later selected action fails, the finalizer MUST restore tracked files to the start-of-run state before writing `finalize-summary.json`; pre-existing tracked user changes MUST be preserved, and untracked files MUST NOT be deleted as part of rollback. The finalizer summary MUST report rollback status with `mutation_rollback.status`, `restored_file_count`, `restored_files[]`, and `error`.
Verified by: TH-HARNESS-AC-019

When `agent-finalize` mutates tracked generated or baseline artifacts and completes successfully, those promoted mutations MUST be explicit in `finalize-summary.json` through `generated.updated_file_count` and `updated_files[]`. Audit or handoff records that cite an `agent-finalize` run MUST distinguish pre-existing worktree changes from finalizer-caused updates. A finalizer update MUST NOT be treated as silent remediation merely because the command succeeded.
Verified by: TH-HARNESS-AC-019

`agent-finalize` MUST NOT invoke `format`, `generate`, `generate-drift`, `migration-drift`, `test-fast`, `test`, `check`, `ci`, `release-check`, browser E2E targets, security scan targets, build targets, `clean`, `distclean`, or `benchmark-claim-check`.
Verified by: TH-HARNESS-AC-019

`agent-finalize` MUST be fail-fast and resumable. It MUST stop at the first failed substep, preserve completed substeps, mark later selected substeps as skipped-after-failure, retain child logs and child summaries when available, and propagate the normalized child failure class and reason when a failed child summary is readable. Summary-write or cleanup-reporting failures MUST be reported without masking an earlier primary child failure.
Verified by: TH-HARNESS-AC-019

`agent-finalize` MUST retain `agent-finalize/finalize-summary.json` with schema ID `cartulary.agent_finalize_summary.v3`. The summary MUST include selected and not-selected actions, private substeps, skipped work, per-action cache state, updated files, `RESULTS_DIR`, failure records, and child artifact references. Human summary output MUST include one compact `[FINALIZE]` line before the ordinary `[RESULT]` line and SHOULD expose reused-action/cache-hit counts when cache evidence exists. Machine output MUST remain exactly one `cartulary.tool_run_summary.v3` JSON object; callers MUST use the `finalize_summary` artifact reference to read command-specific finalizer details.
Verified by: TH-HARNESS-AC-004, TH-HARNESS-AC-019

Check scheduler summaries that include service sessions MUST report service-suite setup timing separately from child test timing. Each `service_sessions[]` entry MUST include `setup_duration_ms`, `ready_at_monotonic_ms`, `child_work_started_at_monotonic_ms`, and `cleanup_duration_ms`; fields MAY be `null` only when the corresponding lifecycle segment did not run or did not reach readiness. Duration baselines for backend, browser, and service-backed child work MUST be derived from child work-unit timings, not from service-suite setup or cleanup timing.

## 9. Failure Classes and Exit Codes

**TH-HARNESS-REQ-300**
Public Make-owned wrappers MUST expose exact normalized public exit codes according to the failure-reason table below in retained summaries and compact failure output. Raw child process exit codes MAY be preserved in summaries but MUST NOT define the normalized public exit code except where `child_target_failure` explicitly delegates to a normalized child failure class.
Verified by: TH-HARNESS-AC-014

Public exit-code selection is reason-based. Wrappers MUST derive the
normalized public exit code from the normalized `failure_reason` and primary-failure
rules in this section, not from the raw process status of a child command. The current GNU Make invocation binding may return GNU Make's executor failure status for a failed recipe; callers that require reason-specific failure codes MUST read the retained `tool-run-summary.json` or compact failure line rather than treating the outer `make` process status as the normalized public exit code.

Scheduler summaries MUST propagate normalized failures from every completed failed work unit whose child target summary is available. The scheduler's own fallback classification is used only when no child summary exists, the child summary is unreadable, or the failure belongs to scheduler orchestration rather than completed child target work. The summary's primary failure still follows Section 9.1 ordering, but `failures[]`, `failure_classes`, and `failure_reasons` MUST retain all completed failed work units. Dependency-skipped completion rows MUST NOT be counted as independent failures when the underlying failed work unit is already represented. A child target assertion failure therefore remains `failure_class=product` and `failure_reason=test_assertion_failure` at the scheduler summary layer, while a concurrent config, harness, artifact, timing, infra, or security failure remains visible in the same scheduler and aggregate summaries.

Every failed retained summary that carries the standard failure fields MUST expose both a non-null `failure_class` and a non-null `failure_reason`. Passing summaries MUST expose no primary failure. A generic shell-wrapper exit such as `command exited with status 1` is diagnostic wrapper evidence when a tool runner has already emitted a classified failure for the same target; it MUST NOT become the primary failure and SHOULD NOT be counted as an independent primary harness failure.

Post-summary scheduler validation failures, including scheduler event, timing, critical-path, summary, or accounting drift detected after child work has completed, MUST be normalized as `failure_class=harness`, `failure_reason=scheduler_accounting_error`, and public exit code `11`. They MUST NOT fall through as caller `configuration_error` merely because they are detected by the scheduler runner.

Failure classification uses two layers:

- `failure_class`: coarse stable grouping for humans and automation.
- `failure_reason`: detailed snake-case reason for diagnosis, exit-code mapping, and handoff.

| Failure class | Meaning                                                                                 |
| ------------- | --------------------------------------------------------------------------------------- |
| `product`     | The product behavior under test failed after harness setup completed.                   |
| `security`    | A blocking security scanner finding was reported after harness setup completed.         |
| `config`      | Caller input, environment, manifest, or local tool configuration was invalid or missing. |
| `infra`       | Required backing infrastructure failed preflight, startup, readiness, or capacity.      |
| `harness`     | Harness orchestration, fixture, scheduler, child aggregation, or cleanup failed.        |
| `artifact`    | Required retained evidence was missing, malformed, invalid, or unsafe.                 |
| `timing`      | A deadline, timeout, or timing-accounting guard failed.                                |
| `interrupted` | The command was cancelled or interrupted.                                               |
| `unknown`     | The wrapper could not classify the failure.                                             |

| Failure reason                | Default class | Trigger                                                            |                                    Public exit code |
| ----------------------------- | ------------- | ------------------------------------------------------------------ | --------------------------------------------------: |
| success                       | none          | No failure                                                         |                                                 `0` |
| `usage_error`                 | `config`      | Invalid arguments, missing required flags, unsupported output mode |                                                 `2` |
| `configuration_error`         | `config`      | Missing/invalid tool, path, env, config, manifest, resource limit  |                                                 `2` |
| `preflight_error`             | `infra`       | Docker/platform/tool preflight fails before managed services       |                                                 `3` |
| `service_start_error`         | `infra`       | Backing service or browser process fails to start                  |                                                 `3` |
| `service_readiness_timeout`   | `infra`       | Started service fails readiness before deadline                    |                                                 `3` |
| `fixture_error`               | `harness`     | DB/bucket/template/reset/janitor/fixture operation or shape validation fails |                                      `3` |
| `resource_conflict`           | `infra`       | Logical resource, port, lock, DB/bucket name, or host conflict     |                                                 `4` |
| `test_assertion_failure`      | `product`     | Test runner assertion fails after harness setup                    |                                                `10` |
| `security_finding`            | `security`    | Blocking Govulncheck vulnerability or enforcing Gosec finding after scanner setup; Fallow uses this reason only if a later adopted security-scan profile selects it |                         `1` |
| `child_target_failure`        | `harness`     | Aggregate child exits nonzero                                      |                         normalized child class exit |
| `tool_diagnostic_failure`     | `harness`     | ShellCheck/Biome/Fallow-style static-analysis, formatter, linter, or tool diagnostic failure after setup |                         `1` |
| `scheduler_accounting_error`  | `harness`     | Manifest, summary, timing, event, or accounting mismatch           |                                                `11` |
| `frontend_row_accounting`     | `harness`     | Implemented frontend row mapped to a target is missing, stale, failed, or otherwise not closed by required row-accounting evidence |                    `11` |
| `test_accounting_unmapped`    | `harness`     | Public target completed with executed tests that were neither mapped nor intentionally classified |                  `11` |
| `artifact_error`              | `artifact`    | Required artifact missing, invalid, unredacted, or schema-invalid  |                                                `11` |
| `cleanup_error`               | `harness`     | Cleanup command/finalizer/leak check/reaper scheduling fails       |         `12` when no earlier primary failure exists |
| `duration_baseline_drift`     | `timing`      | Explicit duration-baseline or warm scheduler timing drift check fails |                                             `13` |
| `timeout_failure`             | `timing`      | Command, readiness, watchdog, cleanup, or lock exceeds deadline    |                                                `13` |
| `cancelled_or_interrupted`    | `interrupted` | Signal, cancellation, abort                                        | `130` for SIGINT, `143` for SIGTERM, otherwise `15` |
| `unknown_failure`             | `unknown`     | Failure cannot be classified                                       |                                                 `1` |

Default human output MUST expose bounded failure fields for a failed public target before GNU Make's generic recipe failure line can be the only visible diagnostic. It SHOULD avoid full failure records unless verbose output is requested. The canonical compact shape is:

```text
failure_class=infra reason=service_readiness_timeout failed=<unit>
```

Full failure records belong in retained JSON summaries and investigation commands.

For `frontend-fallow-static`, missing Fallow tools, missing install state, or invalid `.fallowrc.json` MUST map to `configuration_error`; enforcing static findings MUST map to `tool_diagnostic_failure`; malformed raw reports or invalid normalized summaries MUST map to `artifact_error`. Non-blocking Fallow findings MAY be retained as warning evidence without failing the target.

### 9.1 Primary Failure Selection

```text
select_primary_failure(failures):
  if failures is empty:
    return success
  if any non-cleanup failure exists:
    return non-cleanup failure by:
      1. failure-class precedence from TH-HARNESS-REQ-304,
      2. top-level command lifecycle order from TH-HARNESS-REQ-304,
      3. scheduler event sequence if scheduler-owned,
      4. child target registry order if aggregate-owned,
      5. artifact path lexical order,
      6. failure reason lexical order
  return earliest cleanup failure by cleanup step order
```

**TH-HARNESS-REQ-301**
Cleanup failure after an earlier product or operational failure MUST be recorded but MUST NOT override the public exit code selected for the earlier primary failure.
Verified by: TH-HARNESS-AC-014

**TH-HARNESS-REQ-302**
Harness setup, readiness, fixture, artifact, scheduler, timeout, and cleanup failures MUST NOT use `failure_class=product`. A failing assertion after successful harness setup MUST be classified with `failure_class=product` and `failure_reason=test_assertion_failure`.

Vitest and Playwright per-test timeouts after the test runner has reached product execution are product test failures and MUST be classified as `failure_class=product` with `failure_reason=test_assertion_failure`. Harness-owned watchdogs, command deadlines, lock deadlines, service readiness deadlines, and cleanup deadlines remain operational failures and MUST use `failure_reason=timeout_failure` or `failure_reason=service_readiness_timeout` according to the failure-reason table.

Vitest assertion summaries that contain reporter stack-formatting markers such as `STACK_TRACE_ERROR` MUST preserve actionable assertion context when the runner report provides it. The retained diagnostic MUST keep the assertion title, owner path, raw runner-report reference, reproduce command, and a diagnostic tag such as `vitest_stack_trace_error`; the stack-formatting marker MUST NOT replace an available assertion message as the primary human diagnostic.

Vitest wrappers MAY retain `cartulary.vitest_failure_details.v1` sidecars under their raw artifact directory. When present and schema-valid, phase summaries MUST prefer a matching sidecar assertion message over the runner JSON's stack-formatting marker and MUST retain both the runner JSON and sidecar references in failure details. Missing sidecars remain compatible; summaries MUST continue to use the runner JSON diagnostic fallback when no sidecar entry is available.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-014, TH-HARNESS-AC-021

**TH-HARNESS-REQ-303**
A scheduler MUST preserve the first failed work unit and its retained detail as `failed_work_unit` and `failed_work_unit_detail` even when later sibling work drains and also fails. Scheduler summaries MUST include a bounded `observed_failed_work_units[]` array containing completed nonzero-exit work units in finish order, including later drained sibling failures; that array is diagnostic and MUST NOT rewrite the first failed work unit. Human failure output, scheduler summaries, target summaries, and tool-run summaries MUST choose a primary headline and public exit code from the primary-failure rules without contradicting `failed_work_unit` when the failed work unit has retained classified child evidence.

Scheduler `critical_path_wall_duration_ms` is the scheduler timing envelope and MUST equal the scheduler total duration for every emitted scheduler summary, including failed schedules. When a failed schedule completes no successful work unit, `critical_path_units[]` MUST be empty and `critical_path_terminal_unit` MUST be `null`; the failed path remains represented by `failed_work_unit_detail` and `observed_failed_work_units[]`.
Verified by: TH-HARNESS-AC-014, TH-HARNESS-AC-024

**TH-HARNESS-REQ-304**
Primary-failure precedence is closed. Failure-class precedence is exactly: `product`, `security`, `config`, `infra`, `harness`, `artifact`, `timing`, `interrupted`, `unknown`. Top-level command lifecycle order is exactly: wrapper identity, output-mode resolution, configuration resolution, result-root/run-ID resolution, redaction initialization, semantic target behavior, artifact validation, cleanup or finalizers, public output emission.

When class and lifecycle step tie, scheduler-owned failures order by scheduler event sequence; aggregate-owned child failures order by public child target registry order; artifact failures order by normalized artifact path lexical order; remaining ties order by `failure_reason` lexical order. A cleanup or finalizer failure MUST NOT override an earlier non-cleanup primary failure.
Verified by: TH-HARNESS-AC-014, TH-HARNESS-AC-032

## 10. Scheduler Contract

**TH-HARNESS-REQ-350**
Scheduler manifests are normative scheduler inputs. A scheduler target MUST validate manifest schema, work-unit IDs, dependencies, resource claims, finalizers, output schemas, and timing settings before starting child work.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-024

**TH-HARNESS-REQ-351**
Phase selection MUST include a `phase_namespace` value. The default namespace is `base`. `PHASE=phaseN` without an explicit namespace MUST remain base-only and MUST resolve only through `tools/phase_registry.json`. Frontend phase selection MUST use `PHASE_NAMESPACE=frontend PHASE=FE-P<N>` and MUST resolve only through the current active contiguous phase IDs in `tools/frontend_phase_registry.json`. `task-guide`, `explain-phase`, `phase-slice`, and `service-backed-slice` MUST reject ambiguous or cross-namespace phase identifiers with a bounded usage diagnostic instead of guessing. Executable slice commands MUST reject any non-active frontend phase and any selected frontend row that is not implemented current evidence.

Current browser readiness, browser duration, and base Playwright selection joins MUST use the frontend registry `base_phase_join` field to attach frontend readiness rows to existing base-phase browser selection surfaces. Implementations MUST NOT derive this join from the numeric suffix of `FE-P<N>`. A non-null join attaches that frontend phase to the named base phase; a `null` join means no base-phase attachment. Registry validation MUST reject missing or malformed `base_phase_join` values before browser selection or duration accounting consumes them.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-022

**TH-HARNESS-REQ-352**
Check-scheduler dependency declarations MUST account for readiness work once, at the scheduler layer. A scheduled work unit whose actual child path or retained-summary emission requires pnpm-managed workspace Node packages, including root Node tool dependencies, harness schema-validation dependencies, or frontend dependencies, MUST depend on `check-frontend-install`; a scheduled work unit whose actual child path requires a build artifact or service image MUST depend on the scheduler-modeled readiness unit that produces that artifact or image. A scheduled work unit whose selected behavior and retained-summary emission do not require installed pnpm-managed workspace packages MUST NOT depend directly or indirectly on `FRONTEND_INSTALL_STAMP`. The default fast `check-harness-smoke` child behavior MUST require only the Node runtime and harness source inputs needed by the selected fast smoke checks, though its retained summary emission MAY rely on scheduler-modeled pnpm package readiness.
Verified by: TH-HARNESS-AC-006

If a later profile promotes `frontend-fallow-static` into default `check`, it MUST be a direct scheduler work unit with `check-frontend-install` readiness and the static Node resource shape, not a hidden prerequisite of another frontend target.

The public `check` wrapper MUST NOT run substantial frontend install, build, service-image, or browser readiness work outside scheduler accounting. It MAY perform only minimal runner bootstrap needed to start the scheduler process, including pinned repo-local Node readiness for `run-check-schedule.mjs`, plus fail-fast configuration validation that does not provision dependencies or build artifacts. When scheduler event or summary schema validation depends on pnpm-managed Node packages, the check scheduler MAY defer its own early event validation until the scheduler-visible package readiness unit completes, but it MUST validate deferred records before final summary emission. Frontend install, backend build, migration build, service-image build, service-image warmup, and browser readiness that are required by default `check` work MUST appear as scheduler-visible units in retained scheduler summaries.

### 10.1 Scheduler Manifest Fields

The canonical scheduler input schema is `cartulary.scheduler_manifest.v2`.
`tools/scheduler_manifest.json` is the committed generated scheduler input for
check, service-backed, phase-slice, and future scheduler families. Family-
specific source forms such as check schedule metadata, service-backed
`work_unit_sources[]`, Go shard expansion, and browser group expansion MAY exist
only as upstream authoring inputs. Scheduler runners MUST NOT accept those
family-specific source forms as runtime scheduler manifests. Historical
`cartulary.scheduler_manifest.v1` inputs MAY be read only for retained-run or
diagnostic compatibility; current producers MUST emit v2.

| Field                              | Type                 | Required | Default                      | Rule                                                        |
| ---------------------------------- | -------------------- | -------: | ---------------------------- | ----------------------------------------------------------- |
| `schema_id`                        | string               |      yes | none                         | Must be `cartulary.scheduler_manifest.v2`.                  |
| `generated`                        | object               |      yes | none                         | Generator and authoring-input provenance.                   |
| `schedules[]`                      | array                |      yes | none                         | Normalized scheduler inputs.                                |
| `schedules[].target`               | string               |      yes | none                         | Public target or scheduler target identity.                 |
| `schedules[].scheduler_kind`       | string               |      yes | none                         | `check`, `service_backed`, or `phase_slice`; future families require a later adopted schema/spec revision. |
| `schedules[].capacity_profile`     | string               |      yes | none                         | Registry-backed capacity profile name.                      |
| `schedules[].resource_limits`      | object               |      yes | none                         | Logical resource limits or `auto` policies.                 |
| `schedules[].stop_on_first_failure` | boolean             |      yes | none                         | Check scheduler: `true`; service-backed scheduler: `false`; phase-slice uses its selected scheduler family. |
| `schedules[].progress_tick_seconds` | integer             |      yes | `30`                         | Must be `5..300`; affects reporting only.                   |
| `schedules[].validate_timing`      | boolean              |      yes | `true`                       | Must be `true` for conformance runs.                        |
| `schedules[].summary_groups`       | array                |       no | `[]`                         | Summary grouping policy for scheduler output.               |
| `schedules[].work_units[]`         | array                |      yes | none                         | Ordered by normalized manifest ordinal.                     |
| `work_units[].id`                  | string               |       no | `target`                     | Unique within schedule when defaulted.                      |
| `work_units[].command`             | object               |      yes | none                         | Structured descriptor resolved by the scheduler runner.     |
| `work_units[].priority`            | integer              |       no | `0`                          | Higher integer wins among ready work.                       |
| `work_units[].weight_ms`           | positive integer     |      yes | none                         | Advisory duration estimate only.                            |
| `work_units[].needs[]`             | string array         |       no | `[]`                         | Completion keys required before start.                      |
| `work_units[].completion_keys[]`   | string array         |       no | `[work_unit.id]`             | Added on success.                                           |
| `work_units[].failure_keys[]`      | string array         |       no | completion keys              | Added on failure.                                           |
| `work_units[].complete_on_failure` | boolean              |       no | `false`                      | Adds completion keys even when command fails.               |
| `work_units[].make_prerequisite_policy` | enum `run` or `skip` | required for `make_target` | none | Controls whether scheduler-owned `make_target` execution runs or suppresses the target's recursive Make prerequisites; omitted policy is invalid. |
| `work_units[].resource_claims`     | object               |      yes | `{}`                         | Logical claims only.                                        |
| `work_units[].env`                 | object               |       no | `{}`                         | Scheduler-owned child environment values; MUST NOT override scheduler-owned harness identity variables. |
| `work_units[].readiness_attribution` | object             |       no | none                         | Scheduler-readable readiness/provisioning attribution metadata. When present it MUST declare `timing_role`, `readiness_class`, `warm_threshold_ms`, and `reason`; producers MUST NOT infer readiness attribution from work-unit names. |
| `work_units[].browser_session_group` | string             |       no | stage target                  | Browser stack/session identity shared by compatible browser work units. |
| `work_units[].browser_session_isolation_reason` | string |       no | none                          | Required explanation when a default-check browser stage uses an extra session instead of the shared default-check session. |
| `work_units[].browser_session_finalizer` | boolean        |       no | `true`                        | Whether a browser stage completion unit stops its session. Shared projection sessions MUST use a separate `browser_session_finalizer` work unit instead of coupling one target's summary to all groups. |
| `work_units[].shard_names[]`     | string array         | required for `go_shard_finalize` | none               | Selected Go shard names that the aggregate finalizer is allowed to require. For `go_shard_finalize`, this list MUST be non-empty and MUST match the finalizer's `go_shard:<name>` needs after prefix removal. |
| `work_units[].retained_resource_claims` | object         |       no | `{}`                         | Claims kept after work-unit exit until explicit release.    |
| `work_units[].release_retained_resource_claims` | object |       no | `{}`                         | Retained claims to release after work-unit exit.            |
| `schedules[].finalizers[]`         | array                |      yes | `[]`                         | Always run after scheduler drains or stops.                 |

Supported `work_units[].command.type` values are `make_target`,
`service_session_start`, `browser_stage_session_start`, `browser_group`,
`browser_stage_complete`, `browser_session_finalizer`, `go_shard`, `go_shard_finalize`, and
`service_complete`. Dependency-gated aggregate work, including Go shard
aggregation, MUST remain in `work_units[]` and MUST NOT be modeled as an
unconditional scheduler finalizer.

For every command type, fields not listed for that type are forbidden. Scheduler manifest validation MUST reject unknown command types, missing required fields, forbidden fields, wrong field types, omitted `make_prerequisite_policy` on `make_target` work units, and `make_prerequisite_policy` values outside `run` or `skip` before starting child work.

| `command.type` | Required command fields | Optional fields and defaults | Forbidden command fields | Start and success behavior |
| --- | --- | --- | --- | --- |
| `make_target` | `target` | `service_target` only when joining a service session. | `shard`, `browser_stage`, `group_id` | Starts `make <target>` under scheduler-owned environment. Success requires the target's declared summary and artifact policy. When `make_prerequisite_policy=run`, recursive Make prerequisites run normally; when `skip`, the scheduler injects the prerequisite-skip environment only after declared scheduler dependencies have modeled the required readiness work. |
| `service_session_start` | `service_target` | none | `target`, `shard`, `browser_stage`, `group_id` | Starts the owned service session, emits lease and lifecycle evidence, and retains service-stack claims until service completion or finalizers release them. |
| `browser_stage_session_start` | `service_target`, `browser_stage` | none | `target`, `shard`, `group_id` | Starts the browser session group for an existing service session and retains declared browser/process claims until the session finalizer releases them. |
| `browser_group` | `service_target`, `browser_stage`, `group_id` | none | `target`, `shard` | Runs the selected browser group. Success emits group evidence and the work unit's completion key. |
| `browser_stage_complete` | `service_target`, `browser_stage` | none | `target`, `shard`, `group_id` | Aggregates completed browser groups and emits the stage target summary. For shared projection sessions it MUST depend only on its own browser groups and MUST NOT stop or release the shared session. |
| `browser_session_finalizer` | `service_target`, `browser_session_group` | none | `target`, `shard`, `group_id`, `browser_stage` | Stops a shared browser session and releases its retained browser/process claims after every group in the session has finished. |
| `go_shard` | `target`, `shard`, `service_target` | `complete_on_failure=false` unless explicitly declared on the work unit. | `browser_stage`, `group_id` | Runs one Go shard under the service session. Product assertion failures map as product failures; setup and runtime failures map through Section 9. |
| `go_shard_finalize` | `target`, `service_target` | top-level `work_units[].shard_names[]` is required and is not a command field. | `shard`, `browser_stage`, `group_id` | Aggregates summaries for scheduler-selected shards and emits the target summary. Missing or inconsistent evidence for a selected shard is an artifact or scheduler-accounting failure; shards omitted by the scheduler selection MUST NOT be required by this finalizer. |
| `service_complete` | `service_target` | none | `target`, `shard`, `browser_stage`, `group_id` | Finalizes service-backed aggregate evidence after dependent work completes. Cleanup remains owned by scheduler finalizers and lifecycle rules. |

`weight_ms` is an advisory scheduling estimate. It MUST NOT be treated as a logical resource claim, timeout, benchmark claim, pass/fail threshold, or product performance conformance statement.

Nested child-runner concurrency is not advisory. A work unit whose command launches a worker pool MUST keep its declared resource claims and scheduled child worker budget aligned according to Section 5. Direct public targets may expose different developer-loop defaults only when the scheduler-owned invocation remains deterministic and resource-accounted.

Retained resource claims represent continuing logical capacity pressure after a work unit exits. They MUST NOT be used to preserve historical ownership for resources that no longer constrain future work. A browser stage session MAY retain browser stack, stage-lane, and process claims while the stage remains live, but it MUST NOT retain broad database or object-store capacity solely because the stage used those services during readiness.

Every `go_shard` scheduler work unit MUST be executable by its declared `target` through the shared Go shard-plan contract. Scheduler generation MUST fail before writing a manifest when a work unit assigns an authoritative/raw shard to `backend-integration-support`, a support shard to `backend-integration`, or any shard name that the target runner cannot resolve for the same target.

Phase-map Go rows that carry multiple Go symbols MAY declare `scenario_symbols` as a closed mapping from stable scenario IDs to the row's declared symbols. When a default-check-selected row declares `scenario_symbols`, scheduler generation MUST emit one scheduler-visible exact-symbol Go work unit per scenario, the work-unit shard identity MUST include the scenario ID, and each generated unit MUST retain the original row ID, evidence class, and primary evidence owner. Harness smoke assertions that enumerate scenario-scoped shards MUST derive the expected shard names from phase-map `scenario_symbols` or the shared Go shard-plan contract rather than maintaining phase-specific copies. Scenario-scoped splitting is a scheduling and evidence-accounting refinement only; it MUST NOT convert product-conformance evidence into a projection, drop default-check selection, or require aggregate work that reruns every scenario under the same row. Rows that omit `scenario_symbols` retain the ordinary package, fixture, and duration-aware shard packing rules.

For the `check` scheduler, a service-backed suite session SHOULD start as soon as the suite's own readiness prerequisites are satisfied. Browser build artifacts such as server and migration binaries MUST be modeled as dependencies of browser stage sessions that require them, not as prerequisites of the shared service-suite readiness unit. Backend service-backed shards that use the suite template database MAY depend only on the service-session readiness completion key when they do not require those browser build artifacts.

For the `check` scheduler, readiness work such as frontend install, pinned tool bootstrap, binary builds, and service-image warmup SHOULD be represented as first-class scheduler work units with completion keys. Scheduler-invoked child targets MAY suppress recursive Make prerequisite setup only when `make_prerequisite_policy=skip` and their declared readiness keys are already satisfied. Direct public Make target invocation MUST continue to run its normal prerequisites.

A scheduled work unit that inspects, embeds, serves, signs, packages, or otherwise consumes a generated build artifact MUST depend on the scheduler-visible work unit that produces or proves that artifact. This includes static/security conformance scanners that inspect built bundle roots. A build-artifact cache hit MAY satisfy the producer work unit only under the Section 8 cache contract; it MUST NOT allow an artifact consumer to start without the producer completion key. Cache-profile smoke fixtures MUST create every declared file and directory output for the profile they exercise; partial-output fixtures MUST model a rebuild or fail-closed path and MUST NOT accept success by incomplete output.

The current check schedule MUST model embedded web asset preparation as first-class build-artifact readiness. The `embedded-web-assets` producer work unit MUST depend on `build-web`, `build-server` MUST depend on `embedded-web-assets`, and scheduled `backend-unit` MUST depend on `embedded-web-assets` while its selected Go package graph can compile `internal/platform/httpapi/webassets`. Direct public target invocation MAY keep ordinary Make prerequisites and fallback embedded assets, but scheduler-selected work MUST use producer completion keys instead of ambient source-tree stability.

Readiness work that materially affects timing MUST NOT be hidden behind another scheduler unit's Make prerequisites. The current profile MUST model `testservices-build` separately from `test-service-images`; future frontend or web asset builds that materially contribute to a downstream readiness unit MUST receive the same first-class treatment. Warm eligibility checks MAY reject retained timing evidence when readiness units show provisioning-heavy work above their warm thresholds.

Runtime binary injection is scheduler-owned. The current runtime-binary registry is closed and data-driven. Each entry MUST declare the runtime binary ID, producer target, producer output variable, consumer environment variable, and repo-relative default output path used by scheduler-owned consumer env wiring. The current registry contains exactly one entry:

| ID | Producer target | Producer output variable | Consumer env | Default output path |
| --- | --- | --- | --- | --- |
| `operator` | `build-operator` | `OPERATOR_BIN` | `CARTULARY_OPERATOR_BIN` | `operator` |

Phase-map Go rows MAY declare `runtime_binaries=["operator"]`. A scheduler-selected Go shard or process unit that includes such a row MUST depend on the registry producer target before starting the consumer process, and MUST pass only the registry consumer env to that Go process with the registry default output path unless a later adopted registry rule declares another scheduler-owned source. The public producer input `OPERATOR_BIN` MUST NOT be forwarded to arbitrary child processes. `CARTULARY_OPERATOR_BIN` MUST be forwarded only to Go shard/process units selected from rows that declare `runtime_binaries=["operator"]`.

Declaring or consuming a runtime binary is not itself a resource-isolation claim. Scheduler-expanded Go shards MUST claim only their real logical contention, such as one `process` slot plus database, object-store, CPU, and I/O fixture claims when applicable. A runtime-binary consumer MUST NOT claim the full `process` capacity merely because it consumes `operator`; any future full-lane isolation requirement needs an adopted owner rule that names the isolation reason or a dedicated resource.

Omitted `OPERATOR_BIN` on `build-operator` resolves to the Make default `$(CURDIR)/operator`. Omitted `CARTULARY_OPERATOR_BIN` is invalid for canonical scheduler-selected operator scenario work. Raw direct `go test ./cmd/operator` MAY retain a non-conformance fallback that builds the operator locally, but retained harness evidence MUST NOT rely on hidden nested `make build-operator` calls from operator scenario tests. Public `check` and unrelated targets MUST reject command-line `OPERATOR_BIN` or `CARTULARY_OPERATOR_BIN` before child work; inherited undeclared values MUST be stripped unless the selected runtime-binary registry entry reintroduces the scheduler-owned consumer env.

The runtime consumer MUST verify that `CARTULARY_OPERATOR_BIN` is non-empty, normalized as a filesystem path, an existing regular executable file, and not a symlink path. Missing, non-regular, non-executable, or caller-supplied public-command-line values are `configuration_error`, exit `2`, before product assertions. A mismatch between the consumed binary and the `build-operator` build-artifact reference or output digest is `artifact_error`, exit `11`.
Verified by: TH-HARNESS-AC-024, TH-HARNESS-AC-037

**TH-HARNESS-REQ-398**
Default `check-service-backed` browser work MUST model browser stack sharing explicitly through scheduler-owned browser session groups. Check-selected browser stages SHOULD use one shared default-check browser session when they can safely share backend/frontend/runtime state; any extra session MUST declare `browser_session_isolation_reason`. Direct public browser leaf targets MUST retain their ordinary isolated stack behavior. A shared browser session MUST preserve per-target summaries, per-target completion keys, cleanup on failure, and redaction-safe session artifacts. Reset boundaries MUST be explicit scheduler work or an explicit isolation reason; target-name conventions MUST NOT be used as the sharing contract.

Service-backed browser stage dependencies generated by the schedule renderer MUST be owned by `tools/execution_topology_manifest.json`, not by hard-coded renderer or validator lists. The current ordinary measurement policy is `service_backed_schedules.defaults.browser_stage_generated_needs.measurement.selected_peer_stages=["webserver-backed","stateful","visual","a11y"]`; the renderer and topology validator MUST consume that same owner input and fail closed when a selected peer browser stage is not explicitly listed with an owner reason.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-018

**TH-HARNESS-REQ-399**
Every scheduled `browser_group` that runs inside one service-backed scheduler runtime MUST receive a scheduler-owned worker-admin slot range. The ranges for all concurrently schedulable browser groups in that service-backed runtime MUST be non-overlapping and contiguous from offset `0`, and every group MUST receive `CARTULARY_PLAYWRIGHT_WORKER_COUNT` equal to the total slot count plus `CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET` equal to the start of its range. A group that launches more than one Playwright worker MUST own a range with at least that many slots. Direct public browser leaf targets MAY keep the direct isolated offset default from Section 5, but scheduled browser groups MUST fail before Playwright product assertions when the count, offset, or range is missing or invalid.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-018

### 10.2 Logical Resource Registry

The current-conformance logical resource registry is closed by TH-HARNESS-REQ-353 below. Earlier summary tables MUST NOT be used as resource-bound authority.

Scheduler resource claims MUST distinguish reset-heavy, clone-heavy, transaction-heavy, fixture-start-heavy, browser, build, and static work well enough that one expensive class cannot monopolize all broadly shared I/O capacity. In the default check profile, reset-heavy service-backed shards MUST be capped below total `host_io`/`go_io` capacity while `postgres_reset` remains a separate bottleneck, so static work and clone or transaction shards can backfill while resets run. Owner topology inputs MAY declare additional per-execution-family Go shard claims for fixture-sensitive families when measured scheduler pressure shows that a family needs narrower contention control than its shared scheduler profile. Fixture-specific claims SHOULD use the narrowest resource that models the actual contention and SHOULD NOT inflate broad `host_io`, `go_io`, or `process` claims when a closed fixture resource is available. A caller override MAY reduce total capacity only when the normalized schedule still has a feasible non-deadlocking resource assignment.

**TH-HARNESS-REQ-353**
The logical resource registry is closed by the table below. `tools/scheduler_resource_registry.json` is a mirror of this current-conformance registry and MUST NOT independently add resources, change bounds, change override inputs, or redefine auto policies.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-030

| Resource | Schedulers | Default limit | Auto policy | Override input | Min | Max | Display/order rule | Omission behavior |
| --- | --- | --- | --- | --- | ---: | ---: | --- | --- |
| `host_cpu` | `check` | none | `check_host_cpu` | `CHECK_HOST_CPU_JOBS` | 1 | 256 | display order `10` | resolve by auto policy |
| `host_io` | `check` | none | `check_host_io` | `CHECK_HOST_IO_JOBS` | 1 | 256 | display order `20` | resolve by auto policy |
| `suite_service_stack` | `check` | `1` | none | none | 1 | 256 | display order `30` | use default `1` |
| `migration_scratch_postgres` | `check` | `1` | none | none | 1 | 256 | display order `40` | use default `1` |
| `go_cpu` | `service_backed`, `phase_slice` | none | `service_backed_go_cpu` | `CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT` | 1 | 256 | display order `110` | resolve by auto policy |
| `go_io` | `service_backed`, `phase_slice` | none | `service_backed_go_io` | `CARTULARY_SERVICE_BACKED_GO_IO_LIMIT` | 1 | 256 | display order `120` | resolve by auto policy |
| `browser_stack` | `check`, `service_backed`, `phase_slice` | none | `service_backed_browser_stack` | `CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT` | 1 | 256 | display order `130` | resolve by auto policy |
| `object_store` | `check`, `service_backed`, `phase_slice` | `32` | none | none | 1 | 256 | display order `140` | use default `32` |
| `seaweedfs_fixture` | `check`, `service_backed`, `phase_slice` | `2` | none | none | 1 | 8 | display order `145` | use default `2` |
| `postgres` | `check`, `service_backed`, `phase_slice` | `32` | none | none | 1 | 256 | display order `150` | use default `32` |
| `process` | `check`, `service_backed`, `phase_slice` | `6` | none | none | 1 | 256 | display order `160` | use default `6` |
| `postgres_reset` | `check`, `service_backed`, `phase_slice` | none | `service_backed_postgres_reset` | `CARTULARY_SERVICE_BACKED_POSTGRES_RESET_LIMIT` | 1 | 8 | display order `170` | resolve by auto policy |
| `postgres_clone` | `check`, `service_backed`, `phase_slice` | none | `service_backed_postgres_clone` | `CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT` | 1 | 8 | display order `175` | resolve by auto policy |
| `browser_stage_*` | `check`, `service_backed`, `phase_slice` | `1` | none | manifest positive integer only | 1 | 8 | display order `135`, then resource name lexical order | use generated default `1` unless manifest declares another positive value; current default-check stateful partitions declare `browser_stage_stateful=3` |

Resource override inputs accept positive decimal integers only. An override below the largest declared claim for that resource, above the resource maximum, or incompatible with a feasible non-deadlocking schedule is a `configuration_error`, exit `2`, before child work.

Current capacity profiles are `check_default` for the `check` scheduler, `service_backed_full` and `service_backed_backend` for the `service_backed` scheduler, and `phase_slice_default` for the `phase_slice` scheduler. Capacity profiles are registry-owned shortcuts only; they MUST NOT authorize resources outside the scheduler family named by the profile. Owner-input validation MUST reject unknown profile names, profile names attached to the wrong scheduler family, and auto-policy names outside the current closed set before schedule execution begins.

Go shard scheduler-profile resource claims are a closed shared policy. For the `check` scheduler, CPU and I/O claims use `host_cpu` and `host_io`; for `service_backed` and `phase_slice`, they use `go_cpu` and `go_io`. Profile claims are: default or `balanced` = CPU `1`, I/O `1`; `cpu_heavy` = CPU `2`, I/O `1`; `io_heavy` = CPU `1`, I/O `2`; `transaction_heavy` = CPU `1`, I/O `1`; `reset_heavy` = CPU `1`, I/O `2`, `postgres_reset` `1`; and `clone_heavy` = CPU `1`, I/O `2`, `postgres_clone` `1`. The `phase_slice` mapping intentionally uses `go_io:2` for `reset_heavy`; historical `go_io:3` phase-slice behavior is not current conformance.

**TH-HARNESS-REQ-354**
Auto resource policies are closed by the following algorithms. `available_parallelism` is `os.availableParallelism()` when the runtime exposes it and otherwise the host CPU count; in either case it is floored at `1`. `clamp(value, min, max)` returns `min(max(value, min), max)`.

| Auto policy | Required algorithm |
| --- | --- |
| `check_host_cpu` | `clamp(ceil(available_parallelism * 0.7), 1, 256)`. |
| `check_host_io` | `max(check_host_cpu, largest declared host_io claim in the normalized provisional work-unit set)`. |
| `service_backed_go_cpu` | If no Go shard units exist, return `1`. Otherwise compute `total_weight=sum(max(1, weight_ms))`, `max_weight=max(max(1, weight_ms))`, `weighted_concurrency=ceil(total_weight / max(30000, max_weight))`, `host_concurrency=max(2, available_parallelism - 1)` when `available_parallelism <= 4` and `floor(available_parallelism * 0.75)` otherwise; return `clamp(max(4, min(host_concurrency, weighted_concurrency)), 4, 16)`. |
| `service_backed_go_io` | If no Go shard units exist, return `1`. Otherwise count Go shard scheduler profiles and compute `profile_concurrency=balanced + transaction_heavy + 2*io_heavy + 2*clone_heavy + 2*reset_heavy + ceil(cpu_heavy/2)`; return `clamp(max(6, go_cpu + 2, profile_concurrency), 6, 24)`. |
| `service_backed_browser_stack` | Count distinct `browser_stage_session` work units that claim `browser_stack`, keyed by `browser_session_group`; if no retained session starters exist, count distinct `browser_stage_*` resource lanes in the normalized provisional work-unit set for legacy flat schedules. If the resulting demand count is `0`, return `1`. Otherwise return `max(1, min(demand_count, stack_claiming_unit_count when nonzero, process limit when set, max selected CPU limit when set))`, where the selected CPU limit is `host_cpu` for `check` and `go_cpu` for `service_backed` or `phase_slice`. |
| `service_backed_postgres_clone` | Let CPU and I/O bounds be the selected scheduler CPU and I/O resource limits. If neither bound is positive, return default `6`. Otherwise return `max(1, min(8, max(6, floor(min(positive CPU/I/O bounds)/2))))`. |
| `service_backed_postgres_reset` | Let the I/O bound be `host_io` for `check` and `go_io` for `service_backed` or `phase_slice`. If the bound is absent or non-positive, return default `4`. Otherwise return `max(1, min(8, floor(io_bound/3)))`. |

Current scheduler manifests MUST NOT declare `work_units[].timeout_seconds`; scheduler-owned work-unit watchdogs are deferred until a later adopted revision specifies schema support, runner cancellation semantics, finalizer interaction, event emission, failure mapping, and fixtures end to end. Service readiness, browser readiness, runtime reset, cleanup, Playwright, Vitest, and child Make target deadlines remain owned by their specific sections or tools.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-030

### 10.3 Scheduling Algorithm

```text
pending = work_units in manifest order
running = empty map
completed_keys = empty set
failed_keys = empty set
primary_failure = null
scheduler_stopped = false
emit scheduler_started

while pending is not empty or running is not empty:
  ready = pending units whose dependencies are all in completed_keys
          and whose dependencies are not in failed_keys

  for unit in ready by priority DESC, weight_ms DESC, manifest ordinal ASC, id ASC:
    if scheduler_stopped:
      break
    if any earlier ready unit is resource-blocked and overlaps unit.resource_claims:
      continue
    if resources_available(unit.resource_claims):
      start unit
      remove unit from pending
      add unit to running
      emit unit_started

  if running is empty:
    if pending contains units with failed dependencies:
      mark them skipped in manifest order
      remove them from pending
      continue
    fail with scheduler_accounting_error for deadlock or impossible resources

  wait until one or more running units finish or the progress tick fires

  if progress tick fires:
    emit progress event
    continue

  process finished units by:
    1. observed_monotonic_finished_at ascending,
    2. manifest ordinal ascending

  for each finished unit:
    release non-retained resources
    record status, logs, duration, completion keys, failure keys
    if success or unit.complete_on_failure:
      add completion_keys
    if failure:
      add failure_keys
      set primary_failure if null
      if stop_on_first_failure:
        scheduler_stopped = true

run finalizers in manifest order after all running units drain
release retained resource claims
emit scheduler summary, progress summary, and scheduler_complete
validate timing when validate_timing is true; validation failures are `scheduler_accounting_error`
exit with selected primary failure
```

Finalizer failure becomes primary only when no earlier non-finalizer failure exists.

Work units skipped because `stop_on_first_failure` has selected a primary failed work unit, or because their dependencies include a failed key, are propagation records. They MUST be retained in scheduler and target summaries with their skipped reason and failed dependency, but they MUST NOT be reported as additional root failures.

Dependencies outrank priority: a work unit is not ready until its dependencies are satisfied. Priority affects only ready work and MUST NOT preempt work that is already running.

For the `check` scheduler, priority assignments MUST preserve the service-backed critical path once service readiness exists. A ready `check-service-backed` service session and its expanded browser stage, browser group, Go shard, backend make target, aggregate finalizer, and service-complete child work MUST have higher priority than post-build local evidence, static validation, phase validation, and drift validation work. Build readiness and service-image readiness MAY remain above service-backed child work when those dependencies are still required to create valid service-backed evidence. Lower-priority ready work MAY start only when it does not overlap the resource claims of an earlier ready service-backed child that is resource-blocked.

### 10.4 Event Ordering

| Event field             | Rule                                                                         |
| ----------------------- | ---------------------------------------------------------------------------- |
| `schema_id`             | `cartulary.scheduler_event.v6`.                                              |
| `target`                | Public target or scheduler target identity.                                  |
| `scheduler_kind`        | Scheduler family such as `check`, `service_backed`, or `phase_slice`.        |
| `seq`                   | Starts at `1`, increments by `1`, no gaps.                                   |
| `event`                 | Compact event token such as `scheduler-started`, `unit-started`, or `progress`. |
| `monotonic_ms`          | Non-decreasing scheduler-relative monotonic time.                            |
| `emitted_at`            | RFC3339 UTC. Wall-clock regressions require a `clock-skew` marker event.     |
| Work-unit ordering      | Manifest ordinal unless completion tie rule applies.                         |
| Completion tie          | `observed_monotonic_finished_at` ascending, then manifest ordinal ascending. |
| Artifact ordering       | Lexicographic by normalized artifact path.                                   |
| Resource ordering       | Registry display order, then lexicographic fallback.                         |

## 11. Service and Fixture Lifecycle

**TH-HARNESS-REQ-400**
Service-backed targets MUST run in exactly one service mode: `owned` or `attach`.
Verified by: TH-HARNESS-AC-007

| Mode     | Selection rule                                       | Missing variables                                                      | Ownership                                                                     |
| -------- | ---------------------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `owned`  | `CARTULARY_TEST_SERVICES_ACTIVE` omitted or not `1`. | Not applicable.                                                        | Harness starts and cleans suite resources.                                    |
| `attach` | `CARTULARY_TEST_SERVICES_ACTIVE=1`.                  | Any missing required attach variable fails with `configuration_error`. | Harness uses supplied services and MUST NOT delete container-level resources. |

### 11.0 Postgres Fixture Model

**TH-HARNESS-REQ-405**
Postgres fixture selection MUST be intent-based so future phase growth does not turn fixture setup into hidden critical-path cost. Service-backed Go manifests are the sole owner for target fixture policy, fixture budget, and clone/reset exception reason. Helper code MUST fail closed when service-backed rows omit explicit Postgres fixture policy or budget instead of applying defaults.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-018

| Fixture class       | Policy tokens                         | Intended use                                                                                 | Guardrails                                                                 |
| ------------------- | ------------------------------------- | -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `transaction`       | `transaction`                         | Store-layer tests that use the narrow `postgres.DB` interface and can roll back all writes.  | MUST NOT use package reset; records transaction events instead of resets.   |
| `reusable_database` | `package_reset`, `group_clone`        | Route/runtime tests proven safe to reuse a package or group DB with a closed reset surface.  | Package resets MUST declare a closed `dirty_tables` set when targeted, or an explicit reset reason when broad reset is intentional; grouped clones MUST declare why shared committed state is safe. |
| `isolated_clone`    | `template_clone`                      | Tests needing committed route/runtime state, DB identity, schema mutation, process/runtime isolation, or unsafe reset state. | Explicit clone rows MUST declare why clone isolation is required in the phase manifest. |
| `migration_scratch` | `migration_scratch`, migration helper | Migration runner, upgrade, downgrade, and backfill-path assertions only.                     | Current-head schema assertions MUST use the schema template instead.        |

The suite template database MUST be named from the suite ID and a schema hash derived from sorted migration SQL inputs plus the migration runner identity. Fixture events and summaries MUST include `schema_hash`, `fixture_class`, and `reuse_group` when a package or group reuse key exists. Attach mode MUST fail before child execution when an advertised `CARTULARY_PGTEST_SCHEMA_HASH` does not match the local migration schema hash.

Service-backed Go rows MUST explicitly declare `fixture_policy.postgres` and `fixture_budget.postgres`. The current fixture-budget fields are `max_template_clones`, `max_group_clones`, `max_package_resets`, `max_transactions`, `max_migration_scratch`, `dirty_tables`, and `reset_conformance`. Fixture policy selection MUST prefer `transaction` for rollback-safe tests, `package_reset` for committed-state tests with a known bounded reset surface, `group_clone` for shared seeded packages, and `template_clone` only for schema mutation, process lifecycle, migrations, destructive residue, or isolation-sensitive tests. Multi-symbol Go rows MAY declare `symbol_fixture_overrides` as a closed object keyed by symbols that appear in the same row's `symbols[]`; each override uses the same fixture-policy, fixture-budget, reason, reason-code, and proof fields as the row-level fixture contract, and the override applies only to that symbol. Overrides MUST NOT introduce new symbols, widen selected product-conformance evidence, or change target selection. Explicit `template_clone`, explicit `group_clone`, explicit broad `package_reset`, and `migration_scratch` rows or overrides MUST carry both a human reason field and a closed reason-code field: `template_clone_reason_code`, `group_clone_reason_code`, `package_reset_reason_code`, or `migration_scratch_reason_code`. The current closed Postgres fixture reason codes are `committed_cross_connection_visibility`, `database_identity`, `process_lifecycle`, `schema_mutation`, `destructive_residue`, `shared_seeded_state`, `bounded_reset_surface`, and `migration_scratch`. Fixture reason fields and reason-code fields are admissible only for the explicitly selected policy; stale reason fields from another policy, or any fixture reason field on a row or override without `fixture_policy.postgres`, MUST fail phase-map validation.

| Policy token | Required reason field | Admissible reason-code values |
| --- | --- | --- |
| `template_clone` | `template_clone_reason`, `template_clone_reason_code` | `committed_cross_connection_visibility`, `database_identity`, `process_lifecycle`, `schema_mutation`, `destructive_residue` |
| `group_clone` | `group_clone_reason`, `group_clone_reason_code` | `shared_seeded_state` |
| `package_reset` | `package_reset_reason`, `package_reset_reason_code` when explicit broad or targeted reset evidence is declared | `bounded_reset_surface` |
| `migration_scratch` | `migration_scratch_reason`, `migration_scratch_reason_code` | `migration_scratch` |
| `transaction` | none | none |

`migration_scratch_reason` MUST justify migration, boundary, replay, upgrade, or backfill coverage. Store-layer rows that are transaction-safe MUST use `transaction` and MUST NOT use package reset. `package_reset` is a compatibility path only: new reusable reset rows SHOULD declare a targeted table set closed over relevant foreign-key dependencies when the touched surface is small and stable, and rows that intentionally depend on broad mutable reset MUST declare an explicit reset reason. Backend-store `package_reset` remains unsupported; committed route/runtime behavior belongs in `backend_integration` or `backend_process` before a package reset can be considered. Safe committed route groups SHOULD move from per-test clones to `group_clone` when grouping reduces cost without reducing fixture clarity. `max_group_clones` counts physical `group-reused` database creations, not top-level manifest rows. Nested subtests, child HTTP runtimes, or adapter-variation cases under one `group_clone` row MUST either reuse a parent-scoped group database or be represented by separately budgeted manifest evidence; they MUST NOT multiply group databases behind a single shared-state reason.

Fixture-tier lowering MUST be backed by a closed proof model before selected evidence is relabeled. A retained proof artifact MUST validate as `cartulary.fixture_tier_proof.v1` and MUST identify `target`, `phase`, `row_id`, optional `symbol`, `execution_family`, `effective_fixture_policy`, `proof_kind`, `proof_status`, optional `proof_ref`, human `reason`, `execution_boundary`, `observed_surfaces`, `reset_surface`, and `final_verdict`. Current proof statuses are `accepted`, `retained`, and `blocked`; current final verdict tokens are the same status values. Proof artifacts are retained harness evidence only; they MUST NOT reduce selected product rows, enable scheduler reuse for product tests, make old retained artifacts satisfy current validation, or change the fixture policy tokens accepted by manifests.

The current proof admission rules are closed:

| Policy token | Required execution boundary | Required proof status before lower-tier admission | Required reset/surface closure |
| --- | --- | --- | --- |
| `transaction` | `rollback_transaction` | `accepted` for row-specific lowering or symbol override proof. | The proof MUST show a rollback-scoped store fixture or equivalent transaction boundary, no committed cross-connection visibility requirement, no WebSocket observer residue, and no package reset event. |
| `package_reset` | `committed_package_reset` | `accepted` before any product row may use package reset; raw reset-conformance helper rows are implementation-support only. | The proof MUST include FK-closed `dirty_tables[]` matching `fixture_budget.postgres.dirty_tables`, preserved `goose_db_version`, captured before/after row counts, route-idempotency cleanup or non-applicability, no schema mutation, no process restart dependency, and `object_store=none` or proven object cleanup. Backend-store package reset remains unsupported. |
| `group_clone` | `shared_group_database` | `accepted` when grouped committed state is intentional. | The proof MUST name the shared seeded state, the grouping key, and why all grouped symbols can share committed state without hiding per-test isolation requirements. Nested child runtimes MUST reuse the parent group database or be separately budgeted. |
| `template_clone` | `isolated_template_clone` | `retained` when clone isolation is intentionally preserved. | The proof MUST name the clone-only boundary being preserved, such as process lifecycle, schema mutation, database identity, destructive residue, WebSocket/cross-connection observer semantics, unsafe object-store effects, or unproven cleanup. |
| `migration_scratch` | `migration_scratch_database` | `retained` for migration-path evidence. | The proof MUST stay scoped to migration, replay, upgrade, downgrade, boundary, or backfill coverage and MUST NOT be used for current-head product runtime rows. |

Proof must account for the whole harness execution boundary, not only the product row's narrow assertion. Route/runtime proof MUST include auth/session/bootstrap side effects, `route_idempotency`, jobs, object-store buckets or prefixes, process lifecycle, WebSocket observers, cross-connection observers, and any schema or migration side effects when those surfaces are observed or required. A "no durable rows" product assertion is insufficient for reset admission unless every harness-created durable side effect in the same execution boundary is captured and reset-safe. WebSocket, cross-connection observer, process lifecycle, recovery, operator-binary, and SeaweedFS/object-store migration rows are clone-only unless proof records committed observer setup, teardown, no pending events, no cross-test connection residue, and no object/job residue.

Service-backed fixture shape validation MUST fail unplanned `template_clone` use, unplanned `group_clone` use, unplanned transaction use, forbidden package resets, migration scratch overuse, and declared structural fixture-count overuse. Structural overuse diagnostics for template clones, group clones, and transactions MUST identify bounded actual source details, including top-level test, full test name, caller package or file when available, reuse group when present, actual count, declared budget, and planned manifest symbols. Explicit clone rows MUST carry a clone reason, package resets MUST stay below the current warm target of `30` reset operations and `60000ms` executed reset time, and retained fixture reports MUST distinguish transaction, package reset, group clone, and template clone events so warm scheduler health can identify fixture pressure instead of hiding it inside shard time. Fixture duration evidence is diagnostic and advisory in the default local `check` path; explicit duration drift targets, warm scheduler timing checks, and `agent-finalize RESULTS_DIR=<dir>` own timing freshness. Raising `postgres_clone` capacity is valid only as a measured capacity calibration; it MUST NOT substitute for converting safe tests to transaction, template clone, or group clone policy.

### 11.1 Lifecycle Machine Contract

**TH-HARNESS-REQ-403**
A lifecycle machine is normative only when this NLSpec explicitly labels it normative. A representational lifecycle diagram MUST be labeled non-normative, MUST cite its owning requirements, and MUST NOT add behavior. A normative harness lifecycle machine MUST define scope, instance key, closed state set, closed event set, terminal states, transition table, guard precedence, failure mapping, authoritative state derivation, observable evidence, and conformance criteria. Illegal transitions MUST NOT mutate state, MUST fail closed with Section 9 failure classification, and MUST emit retained evidence. State-advancing artifact writes MUST be atomic and idempotent, or the machine MUST define guardrails that prevent unsafe re-execution. Parent lifecycle logic MUST depend only on child terminal status and retained artifacts, not on child in-memory state.
Verified by: TH-HARNESS-AC-017

Implementations MAY realize a normative lifecycle machine with ordinary control flow, tables, generated code, or a state-machine library. The runtime mechanism is not normative. The closed states, events, transitions, failure mapping, and observable evidence are normative.

Normative lifecycle-machine state and event names MUST be ASCII `lower_snake_case`. A transition table is closed by default: any `(state, event)` pair not listed by the owning machine is illegal.

### 11.2 Normative Service Suite Lifecycle Machine

**TH-HARNESS-REQ-404**
The service suite lifecycle machine is normative for every service-backed suite in `owned` or `attach` mode. The machine ID is `test_services_suite_lifecycle_v1`. The machine instance key is `suite_id`. The authoritative transition record is `_shared/test-services/<suite-id>/lifecycle-events.jsonl`, where every line MUST validate as `cartulary.test_services.lifecycle.v1`. The current state is `requested` before the first lifecycle event and otherwise the `to_state` of the last valid lifecycle event. A missing, malformed, non-sequential, or transition-invalid lifecycle event stream after a suite directory or lease exists MUST fail closed with `failure_class=artifact` and `failure_reason=artifact_error`. The service lease remains cleanup-proof evidence and MUST NOT be interpreted as a transition log.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-017

Lifecycle event `seq` starts at `1`, increments by `1`, and has no gaps. Events MUST be processed in emitted sequence order. When competing conditions are observed before the next event is emitted, guard precedence is:

| State           | Precedence rule                                      |
| --------------- | ---------------------------------------------------- |
| `starting`      | `startup_failed` before `readiness_passed`.          |
| `running_child` | `interrupt_received` before `child_started` before `child_finished` when multiple child signals are observed before the next event is emitted. |
| `cleaning`      | `cleanup_failed` before `cleanup_succeeded`.         |
| all others      | The transition table has at most one allowed event.  |

#### States

| State           | Kind         | Invariants                                                                 | Observable signals                                                                 |
| --------------- | ------------ | -------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `requested`     | initial      | Suite setup has been requested but no lifecycle event has been emitted.    | No lifecycle event stream exists for the selected `suite_id`.                      |
| `starting`      | intermediate | Suite setup or attach-mode validation is in progress; child work has not started. | Latest lifecycle event has `to_state=starting`; lease or startup diagnostics exist when the suite writes them. |
| `ready`         | intermediate | Required services or supplied attach endpoints have passed readiness.      | Latest lifecycle event has `to_state=ready`; readiness diagnostics are retained when produced. |
| `running_child` | intermediate | One or more child work units are executing under the suite.                 | Latest lifecycle event has `to_state=running_child`; the event references child logs or target artifacts when known and reports `active_child_count`. |
| `interrupted`   | intermediate | Cancellation or interruption was observed while child work was active.     | Latest lifecycle event has `to_state=interrupted` and `failure_reason=cancelled_or_interrupted`. |
| `cleaning`      | intermediate | Owned teardown or attach-mode diagnostic finalization is in progress.      | Latest lifecycle event has `to_state=cleaning`; lease `cleanup_state=in_progress` when a lease exists. |
| `cleaned`       | terminal     | Required cleanup or attach-mode finalization completed.                   | Latest lifecycle event has `to_state=cleaned`; lease `cleanup_state=completed` or `deferred`. |
| `failed_start`  | terminal     | Startup, attach validation, preflight, or readiness failed before child work started. | Latest lifecycle event has `to_state=failed_start`; failure summary records the Section 9 reason. |
| `cleanup_failed` | terminal   | Cleanup or finalization failed and retained proof remains for investigation or stale janitor handling. | Latest lifecycle event has `to_state=cleanup_failed`; lease `cleanup_state=failed` when a lease exists. |

#### Events

| Event                | Definition                                                                 |
| -------------------- | -------------------------------------------------------------------------- |
| `start_services`     | Begin owned suite startup or attach-mode suite validation.                 |
| `readiness_passed`   | All readiness predicates required by Section 11.4 passed before deadline.  |
| `startup_failed`     | Startup, attach validation, preflight, fixture preparation, or readiness failed before child work started. |
| `child_started`      | A child target, child command, or scheduler work unit started under a ready suite. |
| `child_finished`     | An active child target, child command, or scheduler work unit exited and its status was recorded. |
| `interrupt_received` | The wrapper observed cancellation or process interruption while child work was active. |
| `cleanup_started`    | Teardown, cleanup, or attach-mode diagnostic finalization started.         |
| `cleanup_succeeded`  | Teardown, cleanup, or attach-mode diagnostic finalization completed.       |
| `cleanup_failed`     | Teardown, cleanup, or attach-mode diagnostic finalization failed.          |

#### Transition Rules

| From state      | Event                | Guard                                      | To state         | Required actions                                                              | Failure mapping                                                                 | Observable evidence |
| --------------- | -------------------- | ------------------------------------------ | ---------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------- | ------------------- |
| `requested`     | `start_services`     | Configuration resolved; `suite_id` allocated; no prior lifecycle event exists. | `starting`       | Create suite directory as needed; write initial lease or attach diagnostic when applicable; append lifecycle event. | If setup cannot begin, fail before mutation with `configuration_error`, `preflight_error`, or `fixture_error` according to the failed predicate. | Lifecycle event `seq=1`; lease or diagnostic artifact refs when present. |
| `starting`      | `readiness_passed`   | All required readiness predicates passed.  | `ready`          | Append lifecycle event and retain readiness diagnostics when produced.         | none                                                                            | Lifecycle event with readiness artifact refs. |
| `starting`      | `startup_failed`     | Any startup, attach validation, preflight, fixture, or readiness predicate failed before child start. | `failed_start`   | Append lifecycle event; record failure summary; terminate known partial resources or leave proof for stale janitor. | `preflight_error`, `service_start_error`, `service_readiness_timeout`, `fixture_error`, or `configuration_error` according to Sections 9 and 11.4. | Lifecycle event with failure fields and proof artifact refs. |
| `ready`         | `child_started`      | Child key is non-empty and not already active. | `running_child`  | Append lifecycle event; set `active_child_count=1`; retain child log or target artifact refs when known. | Child start failure before process launch is `child_target_failure` or `fixture_error` according to the wrapper boundary. | Lifecycle event with child key, `active_child_count=1`, and child artifact refs when known. |
| `running_child` | `child_started`      | Child key is non-empty and not already active. | `running_child`  | Append lifecycle event; increment `active_child_count`; retain child log or target artifact refs when known. | Duplicate or missing child key is an illegal transition. | Lifecycle event with child key and incremented `active_child_count`. |
| `running_child` | `child_finished`     | Active child key is known, child status has been recorded, at least two children remain active before this event, and no interruption wins by guard precedence. | `running_child` | Append lifecycle event; decrement `active_child_count`; retain child status and artifacts. | Unknown child key or negative active count is an illegal transition. Child failure is recorded for primary failure selection by Section 9.1; the lifecycle state itself is not terminal. | Lifecycle event with child key, child status artifact refs, and decremented `active_child_count`. |
| `running_child` | `child_finished`     | Active child key is known, child status has been recorded, exactly one child remains active before this event, and no interruption wins by guard precedence. | `ready` | Append lifecycle event; set `active_child_count=0`; retain child status and artifacts. | Unknown child key or negative active count is an illegal transition. Child failure is recorded for primary failure selection by Section 9.1; the lifecycle state itself is not terminal. | Lifecycle event with child key, child status artifact refs, and `active_child_count=0`. |
| `running_child` | `interrupt_received` | Cancellation or signal wins by guard precedence. | `interrupted`    | Append lifecycle event and preserve child/interruption diagnostics when available; report current `active_child_count`. | `failure_class=interrupted`, `failure_reason=cancelled_or_interrupted`.         | Lifecycle event with interruption fields and `active_child_count`. |
| `ready`         | `cleanup_started`    | No child is running and cleanup/finalization is required. | `cleaning`       | Set lease `cleanup_state=in_progress` when a lease exists; append lifecycle event. | Cleanup start failure is recorded as `cleanup_error` without deleting unproven resources. | Lifecycle event and updated lease. |
| `interrupted`   | `cleanup_started`    | Interruption has been recorded and cleanup/finalization is required. | `cleaning`       | Set lease `cleanup_state=in_progress` when a lease exists; append lifecycle event. | Primary interruption failure is preserved by Section 9.1.                       | Lifecycle event and updated lease. |
| `cleaning`      | `cleanup_succeeded`  | Required cleanup or finalization completed. | `cleaned`        | Set lease `cleanup_state=completed` or `deferred`; append lifecycle event.     | Success unless an earlier primary failure exists.                               | Lifecycle event and final lease. |
| `cleaning`      | `cleanup_failed`     | Cleanup or finalization failed.            | `cleanup_failed` | Set lease `cleanup_state=failed`; append lifecycle event; retain proof for stale janitor. | `cleanup_error`; Section 9.1 decides whether it becomes the public exit-code driver. | Lifecycle event, final lease, and cleanup diagnostics. |

Any listed event presented in an unlisted state MUST append a lifecycle event with `transition_status=illegal`, `from_state` equal to `to_state`, `failure_class=harness`, and `failure_reason=scheduler_accounting_error`, then fail without mutating suite state. An unrecognized event token MUST be rejected before lifecycle mutation and MUST NOT be appended to the schema-valid lifecycle event stream. A terminal state MUST reject every later event as illegal.

**TH-HARNESS-REQ-406**
The service-suite lifecycle active-child counter is normative. `child_started`, `child_finished`, and `interrupt_received` lifecycle events MUST include `active_child_count`. `ready + child_started` sets the count to `1`; `running_child + child_started` increments it; `running_child + child_finished` decrements it and remains in `running_child` while the count is greater than `0`; `running_child + child_finished` transitions to `ready` when the count becomes `0`. Negative active counts, missing child identity, duplicate `child_started` for the same active child key, and `child_finished` for an unknown active child key are illegal transitions under Section 11.2.
Verified by: TH-HARNESS-AC-017, TH-HARNESS-AC-033

**TH-HARNESS-REQ-410**
An authoritative service-backed Go row MUST make its PostgreSQL fixture policy
explicit both in its phase-map fixture contract and at the helper call site.
The helper call-site policy MUST select exactly one of transaction, package
reset, group clone, isolated template clone, or migration scratch and MUST
agree with the scheduler-resolved phase-map policy. Missing or conflicting
policy MUST fail before database preparation.

Package reset MUST be exposed only through an explicitly named reset helper
and remains admissible only with the closed reset proof, dirty-table surface,
reason, and budget required by TH-HARNESS-REQ-405. A generic helper MUST NOT
silently fall back to package reset, group clone, or template clone. Non-row
implementation-support tests MAY use an explicitly selected isolated clone,
transaction, or migration scratch without phase-map attribution, but the
selected intent must still be visible at the call site.

Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-056

### 11.3 Lease Fields

Lease files MUST be written before child work starts, MUST be redacted before retention, and MUST be written atomically as a complete JSON file. A lease is evidence for cleanup only when its resource proof matches the actual resource state; cleanup MUST verify labels, prefixes, generated names, or equivalent proof and MUST NOT trust the lease path alone.

| Field              | Type                                                               |                        Required |
| ------------------ | ------------------------------------------------------------------ | ------------------------------: |
| `schema_id`        | string, `cartulary.test_services.lease.v1`                         |                             yes |
| `lease_id`         | non-empty opaque lease identifier                                  |                             yes |
| `suite_id`         | 24 lowercase hex chars                                             |                             yes |
| `target`           | string                                                             |                             yes |
| `mode`             | `owned` or `attach`                                                |                             yes |
| `ownership_mode`   | `owned` or `attach`                                                |                             yes |
| `result_root`      | normalized path                                                    |                             yes |
| `run_id`           | normalized run ID                                                  |                             yes |
| `run_root`         | normalized run-root path                                           |                             yes |
| `owner_pid`        | integer process ID for the owning wrapper                          |                             yes |
| `created_at`       | RFC3339 UTC                                                        |                             yes |
| `heartbeat_at`     | RFC3339 UTC                                                        |                              no |
| `expires_at`       | RFC3339 UTC                                                        |                              no |
| `resources[]`      | redacted resource records with service kind, logical ID, and proof | yes in owned mode, may be empty |
| `proof_labels`     | object of required labels used to prove container ownership        |           yes for container use |
| `proof_prefixes`   | object of generated DB/bucket/path prefixes used to prove ownership | yes for DB, bucket, or path use |
| `cleanup_state`    | `not_started`, `in_progress`, `completed`, `failed`, or `deferred` |                             yes |

### 11.4 Readiness Deadlines

Browser owned-stack readiness is an ownership predicate, not only an HTTP availability predicate. The backend is ready only when all of the following hold before the deadline:

- the wrapper-started backend process group is still alive;
- the selected backend port is owned by a process in that process group, when the platform exposes listener process metadata;
- the token-protected `GET /api/v1/test/runtime/identity` route returns `cartulary.test.runtime_identity.v1`, `runtime_marker="harness-owned"`, `test_routes_enabled=true`, and a server process ID;
- the backend process group remains alive after the identity probe.

The canonical browser E2E frontend startup mode is a built preview: `build-web` MUST complete before browser stack startup, `apps/web/dist/index.html` MUST exist before the frontend process starts, and the wrapper MUST launch a non-watching preview command rather than the Vite dev server. Missing built frontend artifacts are `configuration_error`, exit `2`; they are not service-readiness failures.

The frontend is ready only when the wrapper-started preview process group is still alive, the selected frontend port is owned by a process in that process group when listener process metadata is available, the frontend HTTP probe succeeds, the process group remains alive after the probe, and stack metadata records `frontend_mode="preview"` and `frontend_command_kind="vite-preview"`. A stale or unrelated listener MUST NOT satisfy browser readiness.

| Resource                     | Deadline | Poll interval | Failure reason                       |
| ---------------------------- | -------: | ------------: | ------------------------------------ |
| Docker preflight             |    `15s` |          `1s` | `preflight_error`                    |
| Postgres container readiness |   `180s` |       `500ms` | `service_readiness_timeout`          |
| local object-store readiness |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Template DB migration        |   `180s` |           n/a | `fixture_error`                      |
| Browser backend readiness    |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Browser frontend readiness   |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Reset route success          |    `30s` |           n/a | `fixture_error` or `timeout_failure` |

### 11.5 Retry and Teardown Rules

**TH-HARNESS-REQ-401**
No hidden startup retry is allowed. Retry is allowed only when a resource row declares `max_attempts`, bounded backoff, retryable failure reasons, and an overall deadline. Readiness polling within a deadline is not a retry.
Verified by: TH-HARNESS-AC-007

**TH-HARNESS-REQ-402**
Owned teardown order MUST be: browser child processes, browser fixtures, reset-tainted runtime roots, test databases, object buckets or prefixes, service containers, lease finalization. Attach mode MUST record diagnostics but MUST NOT delete container-level resources or external services.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-010

Destructive reset, cleanup, attach-mode service mutation, and non-idempotent operations MUST NOT be retried unless a resource row explicitly declares the operation safe to retry.

| Resource operation          | `max_attempts` | Backoff | Retryable failure reasons                                  | Overall deadline | Safe retry scope                                  |
| --------------------------- | -------------: | ------- | ---------------------------------------------------------- | ---------------- | ------------------------------------------------- |
| Docker preflight            |            `1` | none    | none                                                       | `15s`            | none                                              |
| Postgres owned startup      |            `3` | `500ms` | transient Docker startup or transport failure before readiness polling begins | attempt startup only; readiness deadline is Section 11.4 `180s` | Failed attempt container is terminated first.     |
| object-store owned startup  |            `2` | `250ms` | transient Docker startup or transport failure before readiness polling begins | attempt startup only; readiness deadline is Section 11.4 `120s` | Failed attempt container is terminated first.     |
| Template DB migration       |            `1` | none    | none                                                       | `180s`           | none                                              |
| Browser backend startup     |            `1` | none    | none                                                       | `120s`           | readiness polling only                            |
| Browser frontend startup    | `1`; auto-selected service-backed frontend bind conflicts MAY reselect a port up to `5` times before service, database, or backend startup begins | none | strict-port bind or listener conflict only for auto-selected service-backed frontend ports | `120s` per selected port | Retry scope is frontend preview process start/readiness only; explicit frontend port overrides do not retry. |
| Runtime reset route         |            `1` | none    | none                                                       | `30s`            | none                                              |
| Owned teardown and cleanup  |            `1` | none    | none                                                       | cleanup-specific | cleanup records failure and leaves proof for janitor |

Stale janitor cleanup of previously owned service containers is a proof-gated startup preflight maintenance step, not authoritative product evidence. Once ownership proof and current-suite exclusion pass, Docker `not found` and Docker "removal already in progress" results MUST be accepted as idempotent cleanup outcomes and MUST NOT fail the new suite. Concurrent removal MUST be retained as deferred cleanup diagnostics. Docker daemon/list failures, unsafe ownership proof, and non-idempotent removal failures remain blocking startup-preflight failures.

Attach mode MAY write diagnostic records and lease observations. It MUST NOT delete externally supplied services, containers, databases, buckets, or object prefixes.

For browser owned stacks, a listener conflict detected before process startup or during backend/frontend process bind/startup maps to `resource_conflict`. Backend or frontend process exit before readiness maps to `service_start_error` only when retained startup diagnostics and logs do not identify a listener, port, lock, or other resource conflict. A live owned process that does not satisfy its readiness predicates before the deadline maps to `service_readiness_timeout`. Suite-admin login failures after owned readiness has been proven are no longer treated as readiness failures.

Startup retry windows and readiness deadlines are separate. If a Postgres or object-store startup attempt reaches readiness polling and then the Section 11.4 readiness deadline expires, the operation MUST NOT retry that service. The failure MUST be `failure_class=infra`, `failure_reason=service_readiness_timeout`, and public exit `3`. Browser backend startup, runtime reset, and cleanup have `max_attempts=1`; their polling or operation deadlines do not create retry attempts. Browser frontend startup may reselect and retry only auto-selected service-backed frontend ports after a strict-port or listener conflict, and that retry window MUST complete before service, database, or backend startup begins. A selected frontend process that reaches readiness MUST be rechecked after backend readiness without retry before browser child work begins.

### 11.6 Duration Baselines

Duration baselines are advisory scheduler planning data only. They MUST NOT become benchmark claims, product performance conformance, timeout policy, or evidence that product behavior is fast enough.

Baseline values MUST be positive integer `weight_ms` values derived only from successful, uncontaminated retained runs. Missing entries MUST use explicit default weights and MUST be reported as defaulted, not silently ignored.

Baseline refresh MUST reject contaminated evidence, including failed scheduler runs, service startup retries, service failures, reset taint, missing timing events, or interrupted runs.

For row-keyed browser E2E duration baselines, refresh MUST join retained timing evidence to the active phase manifest by manifest row ID. Refresh MAY replace stale stored file/title metadata with the active manifest file/title for that row ID. Planning and drift validation MUST remain strict and reject stale file/title metadata instead of silently using it.

Mutating `browser-e2e-duration-baselines` refresh MUST require an explicit, existing `RESULTS_DIR` retained run selector and MUST reject omitted or ambiguous retained evidence. Read-only `browser-e2e-duration-baseline-drift` MAY use the current retained run default only where the public input matrix declares that default; any caller-supplied `RESULTS_DIR` MUST still resolve to an existing retained root or run root.

Raw Go duration baselines for package-level harness suites MUST be stored and checked by raw package baseline key. Current shard planning, coverage, and drift validation MUST NOT use an aggregate raw-suite duration as a fallback for missing package baselines, and current shard plans MUST NOT emit legacy aggregate baseline keys.

`agent-finalize RESULTS_DIR=<dir>` MUST perform retained-run validation before any duration-baseline refresh action mutates committed baseline artifacts. That validation MUST reject failed, incomplete, contaminated, or non-warm retained evidence before the first mutating refresh substep starts.

Duration-baseline drift checks MAY fail only for severe stale planning. Compact drift diagnostics MUST include `subject`, `planned_ms`, `actual_ms`, `ratio`, and `kind`.

Warm scheduler health checks MAY consume retained timing artifacts from a successful warm-ready run. Such checks MUST remain harness-maintenance evidence and MUST NOT be described as claim-bearing product benchmark evidence. When a warm `check` artifact is evaluated, the check MUST fail if default local `check-service-backed` includes ordinary browser measurement work, if hidden provisioning prevents warm eligibility, if `check-service-backed` exceeds the configured warm budget, or if non-isolated backend/browser peer lanes exceed the configured balance ratio by more than the bounded materiality floor. Unless the caller supplies a different value, the supported WSL2 hard warm-maintenance budget for `check-service-backed` is `155000ms`, the balance ratio is `1.25`, and the materiality floor is `5000ms`.

### 11.7 Network Flow Fixture Materialization Lifecycle

**TH-HARNESS-REQ-407**
Network Flow fixture source roots are committed fixture inputs, not service-owned mutable state. A harness runner MUST NOT write into `fixtures/network-flow/**` during validation, preview, apply, graph, cursor, indicator-link, or transcript comparison work. Any generated, staged, normalized, or copied fixture material MUST live under the current run root and MUST be removed only by ordinary result-root cleanup. `make clean`, service teardown, stale janitors, database reset, and object-store reset MUST NOT delete committed Network Flow fixture roots.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-408**
Before a Network Flow fixture is materialized, the runner MUST validate the manifest schema, path safety, per-file byte hashes, aggregate bundle hashes, frozen status when selected for conformance, and owner routing. Materialization MUST use only manifest-listed files, MUST reject symlinks and traversal paths, and MUST make product execution observe a read-only run-local copy. A failed pre-materialization check is `failure_class=artifact`, `failure_reason=artifact_error`; product code MUST NOT start for that fixture.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-409**
Network Flow fixture materialization participates in the existing service lifecycle only as fixture preparation. If materialization fails before child work starts, the service-suite lifecycle records `startup_failed` with `fixture_error` or `artifact_error` according to Section 9 ownership, preserves diagnostics, and performs ordinary owned teardown. If child product work has started, later comparison failures are product or artifact failures according to the owning assertion, but the committed fixture root remains immutable.
Verified by: TH-HARNESS-AC-049

## 12. Test-Only Harness Routes

**TH-HARNESS-REQ-450**
Test-only harness routes are harness routes. Runtime-control routes include `POST /api/v1/test/runtime/reset`, `GET /api/v1/test/runtime/identity`, `POST /api/v1/test/clock/set`, `POST /api/v1/test/clock/reset`, `GET /api/v1/test/clock/state`, `POST /api/v1/test/runtime/public-error-faults`, `POST /api/v1/test/runtime/network-flow-faults`, `POST /api/v1/test/runtime/network-flow-randomness`, `POST /api/v1/test/runtime/network-flow-auth-transitions`, and `POST /api/v1/test/runtime/network-flow-audit-assertions`. Fixture routes include `POST /api/v1/test/incidents/{incident_id}/saved-views/system`. Any future `/api/v1/test/*` or `/ws/v1/test/*` route that observes or mutates harness runtime state or fixture state is also a test-only harness route. These routes MUST be unavailable unless every enablement predicate below is satisfied. They MUST NOT be documented as production API behavior.

Runtime-linked implementations of Section 12 routes that are registered by application binaries are owned by the platform harness runtime adapter boundary. Private implementation package paths are not public route contracts. Non-test runtime code MUST NOT import broad test-helper packages to register or execute Section 12 routes once a platform harness runtime adapter exists for that route family.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-013

### 12.1 Enablement

**TH-HARNESS-REQ-480**
An in-process application test server MUST receive an explicit test-route
mode. The closed modes are disabled, harness-owned, and custom environment.
An omitted, empty, or unknown mode MUST fail before runtime construction.
Harness-owned mode is admissible only when the test exercises a guarded
test-route contract or registers an owner test-route contribution; ordinary
product-route tests MUST use disabled mode. Custom-environment mode is limited
to negative configuration, authorization, host, origin, token, and process
composition tests that supply the complete environment under test.

This requirement changes private test-helper setup only. It does not change
the production enablement predicates below.

Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-056

| Predicate                      | Required value                                                                      |
| ------------------------------ | ----------------------------------------------------------------------------------- |
| `CARTULARY_ENABLE_TEST_ROUTES` | Exact `1`.                                                                          |
| `CARTULARY_TEST_ROUTE_TOKEN`   | Non-empty string with at least 128 bits of entropy, generated by the harness stack. |
| Runtime ownership              | Server started by a Make-owned browser or test harness stack.                       |
| Production/default runtime     | Test-only harness routes are not registered.                                         |

**TH-HARNESS-REQ-453**
The test-route authorization header name is exactly `X-Cartulary-Test-Route-Token`. Harness-generated route tokens MUST be 32 bytes from a cryptographically secure pseudorandom generator encoded as unpadded base64url, producing exactly 43 ASCII characters. Attach-mode supplied tokens MUST be ASCII visible characters, MUST be length `43..512`, MUST contain no whitespace, MUST NOT equal `test`, `token`, `secret`, `password`, or `changeme`, and MUST NOT be a repeated single-character string. Missing, malformed, or weak attach-mode tokens MUST fail before test-route registration with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2`.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-035

### 12.2 Authorization

| Request condition                                     | Behavior                                                     |
| ----------------------------------------------------- | ------------------------------------------------------------ |
| Test-only harness route not enabled                   | Return ordinary not-found behavior.                          |
| Test-only harness route enabled, missing token header | `403`, `error.code=test_route_forbidden`.                    |
| Test-only harness route enabled, wrong token header   | `403`, `error.code=test_route_forbidden`.                    |
| Test-only harness route enabled, correct token header | Evaluate request after host/origin boundary checks.          |
| Cookie-authenticated request without token            | Forbidden; session auth does not authorize test-only routes. |
| Bearer/session/bootstrap-token request without token  | Forbidden; product auth does not authorize test-only routes. |

CSRF does not apply because cookie authentication is not accepted as authorization for test-only harness routes. Incident roles, session cookies, bearer sessions, bootstrap tokens, and `deployment_admin` authority do not bypass the test-route token requirement.

**TH-HARNESS-REQ-451**
When test routes are enabled and a harness-owned API or browser origin is configured, test-only harness routes MUST reject requests whose request origin is not the harness-declared browser origin or harness-declared API origin, or whose request host does not match the harness-owned API origin. Same-process health and readiness probes MUST use explicitly declared non-destructive health endpoints rather than test-only harness routes. Test-only harness routes MUST NOT enable permissive CORS. A rejected origin or host MUST fail before any runtime-control mutation or fixture mutation with `403`, `error.code=test_route_forbidden`.
Verified by: TH-HARNESS-AC-008

### 12.2.1 Runtime Identity

`GET /api/v1/test/runtime/identity` is a harness test route with the same enablement and authorization predicates as the reset route. It returns `cartulary.test.runtime_identity.v1`, the harness runtime marker, `test_routes_enabled=true`, and the server process ID. It MUST NOT return the route token or any database, object-store, credential, or session secret. Browser wrappers and Playwright global setup use this route to prove that the selected API origin belongs to the current harness-owned backend before destructive reset or suite-admin login work begins.

### 12.2.2 Test Clock

**TH-HARNESS-REQ-460**
The test clock routes are harness test routes with the same enablement, host/origin, and token authorization predicates as the reset route. The route family consists of `POST /api/v1/test/clock/set`, `POST /api/v1/test/clock/reset`, and `GET /api/v1/test/clock/state`. They MAY control or observe the harness-owned runtime clock for scenarios that need deterministic authentication, cursor TTL, timestamp, timezone, source-retention, or session-expiry timing. Because that clock can feed security-sensitive product decisions in harness-owned runtimes, missing token, wrong token, product session credentials alone, wrong host, missing origin when origins are configured, malformed origin, or unapproved origin MUST fail before clock mutation or state disclosure with `403`, `error.code=test_route_forbidden`.
Verified by: TH-HARNESS-AC-035, TH-HARNESS-AC-051

**TH-HARNESS-REQ-461**
`POST /api/v1/test/clock/set` accepts a JSON object with exactly one of `fixed_now` or `offset_seconds`. `fixed_now` MUST be an RFC3339/RFC3339Nano timestamp and is normalized to UTC. `offset_seconds` MUST set the clock to wall time plus that offset and MUST clear any fixed clock. Unknown members, invalid JSON, non-object JSON, trailing JSON, missing both command fields, or specifying both command fields MUST fail with `400`, `error.code=invalid_mutation_payload`.
Verified by: TH-HARNESS-AC-035, TH-HARNESS-AC-051

**TH-HARNESS-REQ-462**
`POST /api/v1/test/clock/reset` accepts no body or `{}` and restores wall-clock mode by clearing both fixed time and offset. Unknown members, invalid JSON, non-object JSON, or trailing JSON MUST fail with `400`, `error.code=invalid_mutation_payload`. `GET /api/v1/test/clock/state` MUST return the current test-clock state without mutation.
Verified by: TH-HARNESS-AC-051

**TH-HARNESS-REQ-463**
Successful test-clock set, reset, and state responses MUST return `cartulary.test.clock_control.v1` in the standard success envelope. The response MUST include `schema_id`, `mode`, `now`, and `offset_seconds`; `mode` is one of `wall`, `offset`, or `fixed`; `now` is the current harness clock in RFC3339Nano UTC form; `fixed_now` appears only when `mode="fixed"`. The response MUST NOT include the test-route token, cookies, product session credentials, database credentials, object-store credentials, host filesystem paths, or private runtime state.
Verified by: TH-HARNESS-AC-051

**TH-HARNESS-REQ-464**
Runtime reset MUST clear any registered in-memory test clock before reset success is accepted, restoring `mode="wall"` and `offset_seconds=0`. A fixture or target that changes the test clock MUST either run in an owned runtime that is torn down afterward or restore wall mode before subsequent unrelated product work starts. Network Flow fixtures that rely on cursor TTL, safe-digest rotation windows, source-retention expiry, soft-delete timing, timezone fold/gap interpretation, uptime-derived timestamps, or timestamp ordinal boundaries MUST record the selected clock-control response in their transcript and MUST cite the adopted product owner requirement that defines the expected time behavior.
Verified by: TH-HARNESS-AC-051

### 12.2.3 Saved-View System Fixture

`POST /api/v1/test/incidents/{incident_id}/saved-views/system` is a harness fixture route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY seed one incident-bound `scope='system'` saved-view fixture per successful request for browser scenarios that must distinguish implementation-owned saved-view configurations from contract-backed system views. It MUST NOT be exposed as production API behavior, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or creating any saved-view row.

This fixture route MUST create only saved-view fixture rows through the saved-view store path. It MUST NOT expose arbitrary SQL execution, projection mutation, generic fixture mutation, or caller-supplied saved-view identity, owner, timestamps, version, or scope. The route fixes `scope='system'`, fixes `owner_user_id=null`, derives `incident_id` from the path, and returns the normal saved-view resource in the standard success envelope with HTTP `201`.

### 12.2.4 Public-Error Fault Control

`POST /api/v1/test/runtime/public-error-faults` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm a one-shot public error envelope for the next exact ordinary `/api/v1/` request whose method and path match the request body. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any fault.

The armed fault is in-memory harness runtime state. It is consumed at the service boundary before the ordinary route handler runs. Matching is exact on uppercase HTTP method and request path; query strings and fragments are not part of the accepted path. Faults MUST apply only to paths that start with `/api/v1/` and MUST NOT apply to paths that start with `/api/v1/test/`. Test-control routes therefore cannot fault themselves or other test-only harness controls.

The request body MUST be a JSON object with exactly the fields below.

| Field          | Required | Behavior                                                                                             |
| -------------- | -------- | ---------------------------------------------------------------------------------------------------- |
| `method`       | yes      | Non-empty HTTP method; normalized to uppercase for exact matching.                                   |
| `path`         | yes      | Exact ordinary public route path beginning `/api/v1/` and not beginning `/api/v1/test/`; no query.   |
| `status`       | yes      | Public error status from `400` through `599`.                                                        |
| `code`         | yes      | Non-empty public error code.                                                                         |
| `message`      | no       | Public error message to place in the public envelope.                                                |
| `retryable`    | no       | Public retryability flag; omitted values default to `false`.                                         |
| `details`      | no       | Public error details object. Consumers MUST render only details keys allowlisted by product UI code. |
| `consume_once` | yes      | Must be `true`; persistent or multi-consume faults are not accepted.                                 |

Unknown members, missing required fields, non-object JSON, invalid JSON, status outside `400..599`, empty `code`, a path outside ordinary `/api/v1/`, a path under `/api/v1/test/`, a path with query or fragment, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_public_error_fault_request`.

Successful arming MUST return HTTP `201` with `cartulary.test.public_error_fault.v1` in the standard success envelope. The response MUST include a generated `fault_id`, normalized `method`, exact `path`, `status`, `code`, `retryable`, and `consume_once=true`. The response MUST NOT include the test-route token, configured origins, cookies, product session credentials, database credentials, object-store credentials, or private runtime state.

The next exact ordinary public-route match MUST return a standard public error envelope using the armed `status`, `code`, `message`, `retryable`, and `details`, with the request's public `request_id`. After that response, the fault MUST be consumed and the same request match MUST reach the ordinary route handler unless another fault is armed.

**TH-HARNESS-REQ-454**
At most one public-error fault may be armed per harness-owned runtime. A request to arm a second fault while one is pending MUST fail before replacing the pending fault with HTTP `409`, `error.code=test_public_error_fault_already_armed`. Runtime reset MUST clear any pending fault before reset success is accepted. A consumed fault MUST be removed before the fault response is written, so a retry of the same ordinary request reaches the ordinary route handler unless another fault has been armed.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-035

### 12.2.5 Network Flow Fault Control

**TH-HARNESS-REQ-455**
`POST /api/v1/test/runtime/network-flow-faults` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm a one-shot Network Flow commit or worker fault for a named harness boundary consumed by Network Flow tests. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any fault.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-456**
Network Flow fault boundaries are closed tokens owned by the harness. They identify where an opted-in Network Flow test implementation checks for an armed fault; they do not define product state-machine semantics. The supported boundary tokens are exactly:

| Boundary token                                                     | Harness use |
| ------------------------------------------------------------------ | ----------- |
| `network_flow.import.before_owner_prepare`                         | Fault before Network Flow table or diagnostic state is prepared for the shared unit of work. |
| `network_flow.import.after_owner_prepare`                          | Fault after Network Flow owner state has been prepared but before later participants are prepared. |
| `network_flow.import.after_indicator_prepare`                      | Fault after indicator create/dedupe or binding participants have prepared their writes. |
| `network_flow.import.after_audit_prepare`                          | Fault after transactional audit occurrences have been prepared. |
| `network_flow.import.after_idempotency_prepare`                    | Fault after idempotency-success state has been prepared. |
| `network_flow.import.after_terminal_publication_prepare`           | Fault after terminal import-result publication has been prepared. |
| `network_flow.import.before_transaction_commit`                    | Fault immediately before the shared transaction commit. |
| `network_flow.import.after_transaction_commit_before_reply`        | Fault after the shared transaction commits but before the apply caller receives the terminal response. |
| `network_flow.worker.before_handler_start`                         | Fault before a durable Network Flow worker handler starts owner work. |
| `network_flow.worker.before_apply_start`                           | Fault before the worker starts an apply attempt. |
| `network_flow.worker.before_cancellation_check`                    | Fault before the worker observes a cancellation gate. |
| `network_flow.worker.before_final_commit`                          | Fault before the worker reaches the final shared transaction commit. |
| `network_flow.worker.after_final_commit_before_terminal_publication` | Fault after final commit and before terminal-result publication or recovery reconciliation. |
| `network_flow.worker.after_terminal_publication_before_ack`        | Fault after terminal publication and before the worker acknowledges durable completion. |
| `network_flow.worker.before_replay_reconciliation`                 | Fault before worker recovery reconciles an already-committed operation with terminal publication. |

Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-457**
The request body MUST be a JSON object with exactly the fields below.

| Field             | Required | Behavior |
| ----------------- | -------- | -------- |
| `boundary`        | yes      | One of the closed Network Flow boundary tokens in TH-HARNESS-REQ-456. |
| `fault_kind`      | yes      | One of `return_error`, `panic`, `cancel_context`, `worker_crash`, or `worker_cancel`; worker-only kinds are accepted only with `network_flow.worker.*` boundaries. |
| `error_code`      | conditional | Required only for `fault_kind="return_error"`; a lowercase safe diagnostic token matching `^[a-z][a-z0-9_]{1,127}$`. |
| `correlation_key` | no       | Optional ASCII token matching `^[A-Za-z0-9._:-]{1,128}$`; when supplied, consumption requires the same key. |
| `consume_once`    | yes      | Must be `true`; persistent or multi-consume faults are not accepted. |

Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported boundary, unsupported fault kind, worker-only fault kind on an import boundary, invalid or misplaced `error_code`, invalid `correlation_key`, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_network_flow_fault_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_fault_control.v1` in the standard success envelope. The response MUST include a generated `fault_id`, exact `boundary`, exact `fault_kind`, optional `error_code`, optional `correlation_key`, and `consume_once=true`. The response MUST NOT include the test-route token, configured origins, cookies, product session credentials, database credentials, object-store credentials, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-458**
The armed Network Flow fault is in-memory harness runtime state and is consumed only by an opted-in Network Flow test implementation at the exact named boundary. A boundary mismatch MUST leave the fault pending. If the fault has a `correlation_key`, a missing or different consumer key MUST leave the fault pending. A consumed fault MUST be removed before applying its effect so retry, replay, or recovery reaches ordinary behavior unless another fault is armed. If no Network Flow fault registry is registered in the harness-owned runtime, boundary checks MUST be no-ops.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-459**
At most one Network Flow fault may be armed per harness-owned runtime. A request to arm a second fault while one is pending MUST fail before replacing the pending fault with HTTP `409`, `error.code=test_network_flow_fault_already_armed`. Runtime reset MUST clear any pending registered Network Flow fault before reset success is accepted. Network Flow fault controls MAY be used to prove all-or-nothing final commit, worker crash, cancellation, terminal-publication replay, and recovery behavior only when the executed fixture also cites the adopted product owner requirement that defines the expected state.
Verified by: TH-HARNESS-AC-050

### 12.2.6 Network Flow Deterministic Randomness Control

**TH-HARNESS-REQ-465**
`POST /api/v1/test/runtime/network-flow-randomness` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm one deterministic random stream for opted-in Network Flow fixture code that needs repeatable IDs, nonces, key IDs, digest salts, or intentional collision values. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any stream.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-466**
Network Flow deterministic-randomness streams are closed harness tokens. They identify fixture-only injection points and do not define product identity semantics, digest algorithms, cursor algorithms, Graph Projection semantics, or public API compatibility. The supported stream tokens are exactly:

| Stream token                             | Harness use |
| ---------------------------------------- | ----------- |
| `network_flow.table_id`                  | Deterministic table identity and table-name collision fixtures. |
| `network_flow.row_id`                    | Deterministic row identity fixtures. |
| `network_flow.diagnostic_id`             | Deterministic diagnostic identity and diagnostic-order fixtures. |
| `network_flow.import_job_id`             | Deterministic import/apply job identity fixtures. |
| `network_flow.import_source_ref`         | Deterministic opaque import-source reference fixtures. |
| `network_flow.cursor_nonce`              | Deterministic cursor nonce, replay, TTL, and rotation fixtures. |
| `network_flow.safe_digest_nonce`         | Deterministic safe-digest salt or key-bound comparison fixtures without exposing production secrets. |
| `network_flow.graph_invocation_id`       | Deterministic ephemeral Graph Projection invocation fixtures. |

Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-467**
The request body MUST be a JSON object with exactly the fields below.

| Field          | Required | Behavior |
| -------------- | -------- | -------- |
| `stream`       | yes      | One of the closed Network Flow stream tokens in TH-HARNESS-REQ-466. |
| `value_kind`   | yes      | One of `uuid`, `token`, or `hex_bytes`. |
| `values`       | yes      | Ordered deterministic values, length `1..256`; duplicate values are allowed only to exercise collision behavior. |
| `consume_once` | yes      | Must be `true`; persistent or multi-consume values are not accepted. |
| `exhaustion`   | yes      | Must be `fail_closed`; an armed stream exhausted by fixture code MUST fail the fixture rather than silently falling back to production randomness. |

For `value_kind="uuid"`, each value MUST be canonical lowercase UUID text. For `value_kind="token"`, each value MUST match `^[A-Za-z0-9._:-]{1,128}$`. For `value_kind="hex_bytes"`, each value MUST be lowercase, even-length hex text no longer than 512 characters.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-468**
Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported stream, unsupported value kind, an empty or oversized `values` array, a value that does not match the selected `value_kind`, `consume_once` other than `true`, or `exhaustion` other than `fail_closed` MUST fail with `400`, `error.code=invalid_network_flow_randomness_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_randomness_control.v1` in the standard success envelope. The response MUST include a generated `control_id`, exact `stream`, exact `value_kind`, `value_count`, `remaining_count`, `consume_once=true`, and `exhaustion="fail_closed"`. The response MUST NOT include deterministic values, the test-route token, configured origins, cookies, product session credentials, database credentials, object-store credentials, production secret material, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-469**
At most one deterministic-randomness sequence may be armed for a stream in a harness-owned runtime. A request to arm a second sequence for the same stream while the stream is registered, including after all values have been consumed, MUST fail before replacing the registered sequence with HTTP `409`, `error.code=test_network_flow_random_stream_already_armed`. Consuming a stream MUST return values in request order exactly once. A missing stream MUST be a no-op for opted-in consumers, but an armed stream that is exhausted or consumed with the wrong `value_kind` MUST fail closed. Runtime reset MUST clear any registered Network Flow deterministic-randomness streams before reset success is accepted.
Verified by: TH-HARNESS-AC-052

### 12.2.7 Network Flow Authorization-Transition Control

**TH-HARNESS-REQ-470**
`POST /api/v1/test/runtime/network-flow-auth-transitions` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm one fixture-only authorization-transition control for opted-in Network Flow tests that need a route-time authorization change, hidden-resource assertion, cursor authorization recheck, or extension-resource invalidation trigger. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any transition.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-471**
Network Flow authorization-transition boundaries, transition kinds, resource kinds, and hidden-response kinds are closed harness tokens. They identify fixture-only injection and assertion points and do not define product authorization semantics, incident membership rules, hidden-resource status codes, cursor algorithms, WebSocket contracts, or public API compatibility. The supported boundary tokens are exactly `network_flow.route.before_authorization`, `network_flow.route.after_authorization_before_lookup`, `network_flow.route.after_lookup_before_response`, `network_flow.cursor.before_authorization_recheck`, `network_flow.websocket.before_invalidation_publish`, and `network_flow.fixture.after_transition`. The supported `transition_kind` values are exactly `incident_membership_revoked`, `incident_membership_restored`, `incident_soft_deleted`, `network_flow_table_soft_deleted`, `network_flow_table_renamed`, `session_revoked`, and `extension_claim_removed`. The supported `resource_kind` values are exactly `incident`, `network_flow_table`, `network_flow_cursor`, `network_flow_graph`, `network_flow_contributors`, and `network_flow_workspace`. The supported `hidden_response_kind` values are exactly `not_found`, `forbidden_without_resource`, `empty_collection`, `cursor_rejected`, `extension_profile_not_claimed`, and `invalidation_event`.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-472**
The request body MUST be a JSON object with exactly the fields below.

| Field                        | Required | Behavior |
| ---------------------------- | -------- | -------- |
| `boundary`                   | yes      | One of the closed Network Flow auth-transition boundary tokens in TH-HARNESS-REQ-471. |
| `transition_kind`            | yes      | One of the closed transition kinds in TH-HARNESS-REQ-471. |
| `actor_ref`                  | yes      | Safe fixture actor reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `incident_ref`               | yes      | Safe fixture incident reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `resource_kind`              | yes      | One of the closed resource kinds in TH-HARNESS-REQ-471. |
| `resource_ref`               | yes      | Safe fixture resource reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `hidden_response_kind`       | yes      | One of the closed hidden-response kinds in TH-HARNESS-REQ-471. |
| `must_not_disclose_resource` | yes      | Must be `true`; controls that allow resource disclosure are not accepted. |
| `correlation_key`            | no       | Optional safe fixture correlation key matching `^[A-Za-z0-9._:-]{1,128}$`; when supplied, consumption requires the same key. |
| `consume_once`               | yes      | Must be `true`; persistent or multi-consume transitions are not accepted. |

Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-473**
Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported boundary, unsupported transition kind, unsupported resource kind, unsupported hidden-response kind, unsafe refs, `must_not_disclose_resource` other than `true`, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_network_flow_auth_transition_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_auth_transition_control.v1` in the standard success envelope. The response MUST include a generated `control_id`, exact boundary, transition kind, actor ref, incident ref, resource kind/ref, hidden response kind, optional correlation key, `must_not_disclose_resource=true`, and `consume_once=true`. The response MUST NOT include product session credentials, membership row IDs, role grants, route-token material, raw hidden resource details, database credentials, object-store credentials, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-474**
An armed Network Flow auth-transition control is in-memory harness runtime state and is consumed only by an opted-in Network Flow test implementation at the exact boundary, actor ref, incident ref, resource ref, and optional correlation key. A mismatch MUST leave the control pending. A consumed control MUST be removed before the fixture applies the transition or hidden-resource assertion, so retry/replay reaches ordinary behavior unless another control has been armed. A request to arm a duplicate exact boundary/actor/incident/resource tuple while one is pending MUST fail before replacement with HTTP `409`, `error.code=test_network_flow_auth_transition_already_armed`; independent tuples MAY be armed concurrently. Runtime reset MUST clear any registered Network Flow auth-transition controls before reset success is accepted.
Verified by: TH-HARNESS-AC-053

### 12.2.8 Network Flow Audit Assertion Control

**TH-HARNESS-REQ-475**
`POST /api/v1/test/runtime/network-flow-audit-assertions` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm one fixture-only audit assertion for opted-in Network Flow tests that need exact domain audit occurrence counts, zero-occurrence cases, or no-additional-occurrence replay checks. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any assertion.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-476**
Network Flow audit assertion kinds, event codes, and resource kinds are closed harness tokens. They identify fixture-only assertion points and do not define product audit semantics, audit storage, operation authorization, event payload shape, or public API compatibility. The supported `assertion_kind` values are exactly `exact_count`, `zero_occurrences`, and `no_audit_replay`. The supported `event_code` values are exactly `network_flow_table_created`, `network_flow_table_renamed`, `network_flow_table_soft_deleted`, `network_flow_graph_query_executed`, `network_flow_indicator_binding_created`, and `network_flow_indicator_binding_reused`. The supported `resource_kind` values are exactly `network_flow_table`, `network_flow_graph`, `network_flow_indicator_binding`, and `network_flow_import`.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-477**
The request body MUST be a JSON object with exactly the fields below.

| Field                       | Required | Behavior |
| --------------------------- | -------- | -------- |
| `assertion_kind`            | yes      | One of the closed assertion kinds in TH-HARNESS-REQ-476. |
| `event_code`                | yes      | One of the closed Network Flow audit event codes in TH-HARNESS-REQ-476. |
| `operation_ref`             | yes      | Safe fixture operation reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `actor_ref`                 | yes      | Safe fixture actor reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `incident_ref`              | yes      | Safe fixture incident reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `resource_kind`             | yes      | One of the closed resource kinds in TH-HARNESS-REQ-476. |
| `resource_ref`              | yes      | Safe fixture resource reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `baseline_count`            | yes      | Non-negative count before the exercised product operation, maximum `1000000`. |
| `expected_final_count`      | yes      | Exact count expected after the exercised product operation, maximum `1000000`. |
| `expected_replay_increment` | yes      | Exact additional occurrence count expected from committed replay, maximum `1000000`; no-audit replay assertions require `0`. |
| `correlation_key`           | no       | Optional safe fixture correlation key matching `^[A-Za-z0-9._:-]{1,128}$`; when supplied, consumption requires the same key. |
| `consume_once`              | yes      | Must be `true`; persistent or multi-consume assertions are not accepted. |

`expected_final_count` MUST be greater than or equal to `baseline_count`. For `assertion_kind="zero_occurrences"`, `baseline_count`, `expected_final_count`, and `expected_replay_increment` MUST all be `0`. For `assertion_kind="no_audit_replay"`, `expected_replay_increment` MUST be `0`.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-478**
Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported assertion kind, unsupported event code, unsupported resource kind, unsafe refs, negative or oversized counts, `expected_final_count < baseline_count`, assertion-kind count-rule violations, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_network_flow_audit_assertion_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_audit_assertion_control.v1` in the standard success envelope. The response MUST include a generated `assertion_id`, exact assertion kind, event code, operation ref, actor ref, incident ref, resource kind/ref, baseline count, expected final count, expected replay increment, optional correlation key, and `consume_once=true`. The response MUST NOT include product session credentials, raw audit payloads, raw source data, safe-digest key material, cursor tokens, membership row IDs, role grants, database credentials, object-store credentials, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-479**
An armed Network Flow audit assertion is in-memory harness runtime state and is consumed only by an opted-in Network Flow test implementation at the exact event code, operation ref, resource kind/ref, and optional correlation key. A mismatch MUST leave the assertion pending. A consumed assertion MUST be removed before the fixture compares the observed product audit counts, so retry/replay reaches ordinary behavior unless another assertion has been armed. A request to arm a duplicate exact event/operation/resource tuple while one is pending MUST fail before replacement with HTTP `409`, `error.code=test_network_flow_audit_assertion_already_armed`; independent tuples MAY be armed concurrently. Runtime reset MUST clear any registered Network Flow audit assertions before reset success is accepted.
Verified by: TH-HARNESS-AC-054

### 12.3 Runtime Reset Request Body

| Body                                | Behavior                                        |
| ----------------------------------- | ----------------------------------------------- |
| No body                             | Accepted.                                       |
| `{}` JSON object                    | Accepted.                                       |
| JSON object with any members        | `400`, `error.code=invalid_test_reset_request`. |
| Non-object JSON                     | `400`, `error.code=invalid_test_reset_request`. |
| Invalid JSON with JSON content type | `400`, `error.code=invalid_test_reset_request`. |

### 12.4 Saved-View System Fixture Request Body

The saved-view system fixture request body MUST be a JSON object with exactly the fixture fields below.

| Field            | Required | Behavior                                                                                                                                 |
| ---------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `view_schema_id` | yes      | Stable workbook `view_schema_id`; unknown or empty values fail with `400`, `error.code=invalid_mutation_payload`.                       |
| `display_name`   | yes      | Saved-view display name; normalized with ordinary saved-view display-name normalization.                                                   |
| `query_json`     | yes      | Persisted saved-view query JSON; normalized with ordinary saved-view persisted-query normalization for the selected `view_schema_id`.      |
| `layout_json`    | no       | Saved-view layout JSON; omitted values receive the ordinary normalized default layout for the selected `view_schema_id`.                  |

Unknown members, including `scope`, `owner_user_id`, `saved_view_id`, `incident_id`, `created_at`, `updated_at`, and `saved_view_version`, MUST fail with `400`, `error.code=invalid_mutation_payload`. Missing required fields, non-object JSON, invalid JSON, invalid `view_schema_id`, invalid `display_name`, invalid persisted query shape, or invalid layout shape MUST fail through the ordinary saved-view mutation error envelope with `400`, `error.code=invalid_mutation_payload`, and field/reason details when available.

Successful fixture creation MUST return a saved-view resource with `scope='system'`, `owner_user_id=null`, the path `incident_id`, a generated `saved_view_id`, normalized `display_name`, normalized `query_json`, normalized `layout_json`, and store-managed timestamps/version. The returned resource MAY be visible through ordinary saved-view list behavior because `scope='system'` is visible fixture data; that visibility does not make the fixture route a production create API and does not change the ordinary saved-view create rule that rejects `scope='system'`.

### 12.5 Runtime Reset Concurrency and Timeout

| Condition            | Behavior                                                                                    |
| -------------------- | ------------------------------------------------------------------------------------------- |
| No reset active      | Acquire reset lock and run reset.                                                           |
| Reset already active | `409`, `error.code=test_runtime_reset_in_progress`.                                         |
| Reset exceeds `30s`  | `503`, `error.code=test_runtime_reset_timeout`; response includes failed action when known. |

### 12.6 Runtime Reset Algorithm and Partial Failure

**TH-HARNESS-REQ-452**
The reset route MUST preserve migration metadata, restore the active deployment admin, truncate mutable public-schema runtime state, clear route idempotency state, clear registered in-memory test-clock state, clear in-memory public-error fault state, clear registered in-memory Network Flow fault state, clear registered in-memory Network Flow deterministic-randomness state, clear registered in-memory Network Flow auth-transition state, clear registered in-memory Network Flow audit-assertion state, and clear the configured object store bucket or prefix for the harness-owned runtime.

The database reset table set is selected by this algorithm:

```text
select_reset_tables(database):
  query information_schema.tables
  keep rows where table_schema = "public"
  keep rows where table_type = "BASE TABLE"
  reject table_name = "goose_db_version"
  order table_name ascending
  return table_name list
```

The reset MUST execute `TRUNCATE TABLE public.<table> ... RESTART IDENTITY CASCADE` for the selected table list inside one database transaction, using identifier-safe table quoting. If the selected table list is empty, the truncate step is a successful no-op and the bootstrap restoration step still runs. After truncate, bootstrap restoration MUST restore exactly one active deployment admin and exactly one bootstrap marker before commit. The `goose_db_version` row count before reset MUST equal the count after reset and MUST be nonzero. The `route_idempotency` row count after reset MUST be `0`.

Database table truncation and bootstrap restoration MUST execute in one database transaction when the database supports that transaction shape. Object-store deletion occurs after the database transaction commits. The route MUST NOT claim rollback across object-store deletion.

The reset response MUST include `tables_reset` in sorted order, post-reset counts, `object_count_after`, `partial_failure`, and `failed_action` when a failed action exists. `tables_reset` MUST be the exact output of `select_reset_tables(database)` for the reset attempt.

| Surface | Selection rule | Mutation | Success proof | Failure status/code |
| --- | --- | --- | --- | --- |
| Migration metadata | `public.goose_db_version` only | Preserve; never truncate | before and after counts equal and nonzero | `500`, `error.code=test_runtime_reset_failed` |
| Public mutable DB tables | `select_reset_tables(database)` | Truncate with `RESTART IDENTITY CASCADE` inside the reset transaction | `tables_reset` sorted and post-reset mutable counts are zero except bootstrap-restored rows | `500`, `error.code=test_runtime_reset_failed` |
| Bootstrap admin state | Ordinary bootstrap preflight inside the reset transaction | Restore active deployment admin and bootstrap marker | exactly one active deployment admin and exactly one bootstrap marker | `500`, `error.code=test_runtime_reset_failed` |
| Route idempotency | `public.route_idempotency` when selected by the table algorithm | Truncate with other mutable tables | post-reset `route_idempotency` count equals `0` | `500`, `error.code=test_runtime_reset_failed` |
| Public-error fault state | In-memory harness runtime fault slot | Clear pending fault before success is accepted | no pending fault remains after reset | `500`, `error.code=test_runtime_reset_failed` when clear hook fails |
| Configured object-store bucket or prefix | Harness-owned object-store configuration for the runtime | Delete every object after DB transaction commit | `object_count_after=0` | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true` |

| Failure point                                      | Required response                                                                                                  |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Before DB transaction commit                       | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=false` unless prior mutation occurred.             |
| After DB commit and before object cleanup complete | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true`.                                             |
| Object cleanup partial deletion                    | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true`, include `object_count_after` if measurable. |
| Bootstrap admin not restored                       | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true`.                                             |

A browser wrapper receiving `partial_failure=true` MUST mark the owned stack tainted and restart it before further browser child work.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-034

### 12.7 Runtime Reset Success Readiness

Success requires all of the following:

- migration metadata preserved;
- active deployment admin restored;
- mutable incident/product tables empty;
- route idempotency table empty;
- object count after reset equals `0`;
- response validates against `cartulary.test.runtime_reset.v1`.

## 13. Cleanup and Destructive Safety

**TH-HARNESS-REQ-500**
Cleanup is destructive. Cleanup commands MUST delete only paths or resources satisfying the exact ownership predicates in this section.
Verified by: TH-HARNESS-AC-009, TH-HARNESS-AC-010

`make clean` and `make distclean` are repo-local cleanup commands. They MUST NOT remove caller-supplied external result roots and MUST NOT stop local Compose services. Local service teardown belongs to `make services-down`, not to repo-local cleanup.
Frontend dependency install state is a coupled repo-local artifact set. `make clean` MUST preserve installed dependency roots for local loop speed, while `make distclean` MUST remove the repo-local pnpm store, frontend install stamps, and root/workspace `node_modules` directories together so stale package-manager metadata cannot survive without its store.
Harness cache state is repo-local acceleration state. `make clean` MUST preserve default `.cache/cartulary/*` cache roots so ordinary cleanup does not erase warm local readiness. `make distclean` MUST remove default cache roots under `.cache/cartulary/`, including readiness, build-artifact, agent-finalize action-cache, and execution-topology render cache roots. Neither cleanup target may remove caller-supplied cache directories outside the repository.
Fallow configs, rule packs, schemas, and any future reviewed baselines under repository source/tool roots are harness inputs, not cleanup-owned artifacts. Fallow run-root outputs are ordinary retained artifacts and may be removed only through result-root cleanup predicates.

### 13.1 Path Algorithm

```text
normalize_cleanup_candidate(path):
  reject empty string
  reject NUL
  reject "/"
  reject "."
  reject ".."
  reject any caller-supplied segment equal to ".."
  reject backslash on POSIX conformance hosts
  resolve relative paths against repository root
  reject absolute paths outside repository root
  reject protected repository roots named in the table below when they are named as cleanup candidates
  lstat path
  if path is symlink:
    unlink symlink object only
    MUST NOT follow target
  if path is directory:
    remove directory tree only after every traversed entry remains under the candidate root by lexical path and lstat traversal
```

The protected repository root set is closed in the current profile:

| Protected root | Protection rule |
| --- | --- |
| `.git` | Reject when named directly as a cleanup candidate. |
| `docs` | Reject when named directly as a cleanup candidate. |
| `cmd` | Reject when named directly as a cleanup candidate. |
| `internal` | Reject when named directly as a cleanup candidate. |
| `apps` | Reject when named directly as a cleanup candidate. |
| `packages` | Reject when named directly as a cleanup candidate. |
| `contracts` | Reject when named directly as a cleanup candidate. |
| `db/migrations` | Reject when named directly as a cleanup candidate. |
| `db/queries` | Reject when named directly as a cleanup candidate. |
| `configs` | Reject when named directly as a cleanup candidate. |
| `scripts` | Reject when named directly as a cleanup candidate. |
| `tools` | Reject when named directly as a cleanup candidate. |
| `go.mod` | Reject when named directly as a cleanup candidate. |
| `go.sum` | Reject when named directly as a cleanup candidate. |
| `package.json` | Reject when named directly as a cleanup candidate. |
| `pnpm-lock.yaml` | Reject when named directly as a cleanup candidate. |
| `pnpm-workspace.yaml` | Reject when named directly as a cleanup candidate. |

A child path under a protected root MAY be removed only when Section 13.2 or another adopted cleanup table explicitly lists that exact path or path family as cleanup-owned. Missing cleanup-owned paths are successful no-ops. A path that is both protected and cleanup-owned MUST use the narrower cleanup-owned row; broad ancestor deletion remains rejected.

### 13.2 Cleanup Scope

| Command               |      Removes default result root? | Removes custom `CARTULARY_TEST_RESULTS_DIR`? | Removes default `.cache/cartulary` cache roots? | Removes external Go caches? | Stops Docker/Compose globally? |
| --------------------- | --------------------------------: | -------------------------------------------: | ---------------------------------------------: | --------------------------: | -----------------------------: |
| `make clean`          | yes, only default registered path |                                           no |                                             no |                          no |                             no |
| `make distclean`      | yes, only default registered path |                                           no |                                            yes |                          no |                             no |
| `make services-down`  |                                no |                                           no |                                             no |                          no | no; stops only this repo's local Compose services and preserves named volumes |
| Service-suite cleanup |        only suite-owned artifacts |                                           no |                                             no |                          no |                             no |
| Stale janitor         |        proof-gated resources only |                                           no |                                             no |                          no |                             no |

`make distclean` owns removal of `.pnpm-store`, the repository-root `node_modules` directory, workspace package `node_modules` directories under `apps/web` and `packages/*`, and default repo-local cache roots under `.cache/cartulary/`. Missing workspace dependency roots or cache roots are not cleanup failures.

### 13.2.1 Local Service And Data Reset Scope

`make services-down` MUST stop only the local Compose services declared for repository development and MUST preserve named volumes. It MUST NOT pass a Compose volume-removal flag. `db-down` is not a current public command binding; new and existing automation MUST use `services-down`.

`make db-reset` MUST recreate only the local development database and rerun migrations. It MAY start Postgres to perform the reset, but it MUST NOT reset, delete, or inspect object storage. A real `db-reset` MUST reject before Compose, database, migration, or object-store commands unless `CARTULARY_DESTRUCTIVE_CONFIRM=db-reset` was supplied on the Make command line.

`make object-store-reset` MUST clear only objects in the configured local object-store bucket and MUST leave the bucket present afterward. In the current implementation profile the local object store is SeaweedFS S3, and the public command and command ID are provider-neutral. A real `object-store-reset` MUST reject before Compose or object-store commands unless `CARTULARY_DESTRUCTIVE_CONFIRM=object-store-reset` was supplied on the Make command line.

### 13.3 Stale Janitor Thresholds

| Resource        | Completed-run predicate                                         | Uncompleted stale predicate                | Active-resource rule                                                                     |
| --------------- | --------------------------------------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------- |
| Database        | Completed summary or lease cleanup state older than 15 minutes. | Lease or metadata older than 24 hours.     | Active connections may be terminated only after proof predicate passes.                  |
| Bucket          | Completed summary or lease cleanup state older than 15 minutes. | Lease or metadata older than 24 hours.     | Delete only generated bucket/prefix with proof metadata.                                 |
| Container       | Completed summary or lease cleanup state older than 15 minutes. | Harness Docker labels older than 24 hours. | Running container may be stopped only if proof predicate passes and label owner matches. |
| Browser fixture | Completed target summary older than 15 minutes.                 | Fixture metadata older than 6 hours.       | Delete only generated fixture directory with ownership metadata.                         |
| Browser process/session | Completed session lease older than 15 minutes.          | Session lease older than 6 hours with matching runtime-root marker and process command/env proof. | Running processes may be stopped only when PGID, runtime root, command/env proof, and lease identity all match; a port listener alone is never sufficient proof. |

For container cleanup, an already-deleting Docker resource is treated as deferred successful cleanup only after the same proof predicates pass. This compatibility rule exists to make repeated service-backed public targets reproducible under Docker's asynchronous removal lifecycle; it MUST NOT broaden cleanup authority to unlabelled, current-suite, or externally owned containers.

### 13.4 Dry-Run Contract

| Setting                                        | Behavior                                                                                                                    |
| ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `CARTULARY_CLEANUP_DRY_RUN` omitted or not `1` | Cleanup or reset may delete resources satisfying predicates and confirmation rules.                                         |
| `CARTULARY_CLEANUP_DRY_RUN=1`                  | Cleanup MUST emit deletion candidates and reasons and MUST NOT start services, stop services, delete files, delete DBs, delete bucket objects, delete buckets, delete containers, or delete browser fixtures. |

Dry-run output MUST include normalized path or resource identity, proof predicate, action that would be taken, and rejection reason for retained candidates. For human destructive targets, the dry-run line format is:

```text
DRY-RUN <action> <normalized-identity> <proof-or-rejection-reason>
```

`CARTULARY_DESTRUCTIVE_CONFIRM` is ignored for dry-runs. Inherited environment values MUST NOT satisfy the confirmation predicate for public Make targets; only Make command-line values are valid confirmation sources.

### 13.5 Parent-Death and Reaper Rule

Immediate cleanup after parent death is not guaranteed. The conformance guarantee is:

- owned resources carry enough lease or proof metadata for later stale janitor evaluation;
- detached reaper scheduling is optional unless the command declares it;
- if a detached reaper is scheduled, it writes `reaper-scheduled.json` with lease ID, started-at timestamp, target resources, and timeout.

## 14. Platform and CI Support

**TH-HARNESS-REQ-550**
The current conformance support matrix is closed by this section. A target may be run elsewhere, but unsupported environments MUST NOT be used for current harness conformance claims.
Verified by: TH-HARNESS-AC-012

| Environment/tool                                   | Current conformance status                 | Required evidence                                                                                                                             |
| -------------------------------------------------- | ------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Linux x86_64 with Docker Engine and Docker Compose | required                                   | Full acceptance matrix.                                                                                                                       |
| WSL2 Ubuntu with Docker Desktop integration        | supported compatibility profile            | `doctor`, `test-fast`, `check`, browser E2E smoke.                                                                                            |
| macOS                                              | unsupported for current conformance        | None.                                                                                                                                         |
| Windows native                                     | unsupported                                | None; use WSL2 profile.                                                                                                                       |
| Hosted CI provider                                 | provider-neutral only                      | `make ci`; no annotation/upload claims.                                                                                                       |
| Podman/Podman Compose                              | unsupported                                | None.                                                                                                                                         |
| Docker                                             | required for service-backed targets        | Missing Docker yields `preflight_error`.                                                                                                      |
| Docker Compose                                     | required for local Compose targets         | Missing Compose yields `preflight_error` when Compose is absent and `service_readiness_timeout` when Compose-started service readiness fails. |
| Go/Node/pnpm/Playwright/Bash utilities             | required as pinned by repository procedure | Version mismatch yields `configuration_error`.                                                                                                |
| Linux inotify capacity                             | required only for Vite dev surfaces        | Low watcher limits or exhausted watcher usage MUST be diagnosed before Vite dev startup; release/browser E2E preview paths MUST NOT require this preflight. |

Harness diagnostics MAY report Linux inotify `max_user_watches`, `max_user_instances`, best-effort current watcher usage, and bounded operator diagnostics. The harness MUST NOT mutate host sysctl settings.

## 15. Security and Redaction

**TH-HARNESS-REQ-600**
Centralized summaries, machine output, and retained logs captured by harness wrappers MUST be redacted before retention and before stdout emission.
Verified by: TH-HARNESS-AC-011

**TH-HARNESS-REQ-601**
Redaction MUST be applied to captured stdout, stderr, wrapper diagnostics, machine JSON, retained logs, service env dumps, and summary artifacts before those bytes are written outside a private runtime working file or emitted to stdout/stderr. A redaction failure MUST fail the public target with `failure_class=artifact`, `failure_reason=artifact_error`, and public exit code `11` unless an earlier primary failure is preserved by Section 9.1.
Verified by: TH-HARNESS-AC-011

**TH-HARNESS-REQ-602**
The redaction algorithm MUST apply to both keys and values after decoding structured JSON where possible and to raw text otherwise. Matching MUST be case-insensitive for key names and header names. At minimum, the algorithm MUST redact:

- variables, JSON keys, HTTP headers, and CLI arguments whose names match the secret pattern table;
- URL userinfo and DSN password segments;
- bearer-token, session-cookie, JWT, private-key, and object-store credential forms in raw text;
- `Authorization`, `Cookie`, `Set-Cookie`, and `X-Cartulary-Test-Route-Token` values.
Verified by: TH-HARNESS-AC-011

Structured redaction MUST preserve schema-owned container shapes and scalar types unless the value itself is secret. Object and array fields such as `service_sessions`, `browser_stage_sessions`, `session_target`, `cleanup_status`, `lease_file`, and timing fields MUST NOT be replaced merely because their names contain a secret-related substring. Secret key matching MUST use exact or anchored credential-name patterns rather than broad substring matching that can redact structural diagnostics.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-011

**TH-HARNESS-REQ-604**
Structured secret-key matching is closed. Before comparing a structured key name, the redactor MUST uppercase it and replace every non-ASCII-alphanumeric run with one `_`, then trim leading and trailing `_`. The resulting token is secret-bearing only when it equals or has an anchored credential suffix/prefix equivalent to one of: `PASSWORD`, `PASS`, `PWD`, `TOKEN`, `JWT`, `BEARER`, `API_KEY`, `ACCESS_KEY`, `SECRET_KEY`, `AUTHORIZATION`, `COOKIE`, `SET_COOKIE`, or `X_CARTULARY_TEST_ROUTE_TOKEN`. Substring-only matches such as `session_target` containing `token` across token boundaries MUST NOT redact the field.

Raw-text redaction MUST apply after structured redaction to these closed families: URL userinfo (`scheme://userinfo@host`), PostgreSQL-style DSN password segments (`password=` or `:password@` credential forms), bearer authorization headers, JWT-like three-part base64url tokens, PEM private-key blocks, and S3-compatible access-key or secret-key assignments. Structured redaction MUST preserve object and array shapes and preserve numeric, boolean, and null scalar types unless that scalar value itself is secret. A redaction write or validation failure maps to `failure_class=artifact`, `failure_reason=artifact_error`, and public exit `11` unless Section 9.1 preserves an earlier primary failure.
Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-036

**TH-HARNESS-REQ-605**
Runtime and support security scans MUST derive their default support-root
exclusions and support package patterns from the schema-validated test-support
inventory required by TH-HARNESS-REQ-067. Every registered Go support root
MUST be included in the support scan. A root MAY be excluded from the runtime
profile only when the inventory classifies it as test support; a package that
is compiled into a production binary, including a module-owned harness-control
package, MUST remain in the runtime profile.

An unknown `internal/**/testsupport` root, a registered root missing from the
support profile, a duplicate path, or a path excluded from both profiles MUST
fail before the security tool starts. Publicly declared security overrides
remain governed by the Section 5 input registry and MUST NOT silently replace
inventory validation.

Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-056

SeaweedFS strict release evidence MUST derive its redaction scan input set from the current release evidence run, the current `seaweedfs-compatibility` target run-root compatibility report, and the current backend-process Phase E backup/restore and Phase F migration artifact roots selected by the release-gate invocation. The strict compatibility input MUST be `CARTULARY_TEST_RESULTS_DIR/CARTULARY_TEST_RUN_ID/seaweedfs-compatibility/object-store-compatibility-report.json` or an equivalent caller-supplied path under the same current run root and target directory, and the sibling `seaweedfs-compatibility/tool-run-summary.json` MUST report a passing `seaweedfs-compatibility` target. Strict release targets MUST run the current `seaweedfs-compatibility` prerequisite; an invocation that explicitly skips prerequisites MUST NOT claim compatibility evidence. Stable copied release-artifact compatibility reports, fixed retained artifact path lists, newest-run fallback evidence, and retained `services-up` compatibility reports MUST NOT satisfy the strict release gate. `SEAWEEDFS_MIGRATION_PASS_DIR`, when explicitly supplied by the caller or release orchestration, selects the current Phase F pass directory; otherwise the gate derives the Phase E/F roots from `CARTULARY_TEST_RESULTS_DIR/CARTULARY_TEST_RUN_ID`. Missing selected child artifacts MUST be reported as blocking artifact findings rather than replaced with fallback evidence.

Phase 10 SeaweedFS migration-preservation support rows are release-support
evidence only. They MUST remain reachable through explicit/release surfaces and
MUST NOT be selected by default local `check` as bounded readiness or product
recovery conformance evidence.
Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-015

**TH-HARNESS-REQ-603**
Retained run roots and target artifact directories MUST be created with owner-only permissions on POSIX conformance hosts unless the caller explicitly supplied a custom result root whose permissions cannot be narrowed without changing ownership. Required summary artifacts and retained logs MUST be written with owner-read/write permissions. A custom result root that is world-writable without the sticky bit, or that cannot protect newly created files from other users on the host, MUST fail preflight with `configuration_error`.
Verified by: TH-HARNESS-AC-003, TH-HARNESS-AC-011, TH-HARNESS-AC-015

Screenshots, videos, traces, visual geometry diagnostics, and Playwright HTML reports are diagnostic secret-bearing artifacts. They MUST NOT be described as safe to upload or publish without separate review. Browser visual targets MAY retain compact geometry diagnostics for workbook screenshot failures, including scroll metrics, visible field keys, required element rectangles, active element identity, and inspector state. Those diagnostics are harness mechanics only; they MUST NOT define product UI behavior or visual-snapshot refresh authority.

Workbook visual regression tests that capture an outer grid shell while driving an inner grid scrollport MUST normalize and verify both layers before assertion. The screenshot-target shell MUST be reset to `scrollLeft=0` and `scrollTop=0` for left/default viewport captures unless a test explicitly declares a different shell-scroll contract, while the owned grid scrollport MUST be normalized to the test's declared scroll or anchor state. Anchor-based captures that intentionally frame off-screen workbook columns are explicit shell-scroll contracts and MUST still reset stale shell state before computing their deterministic offset. Visual diagnostics SHOULD identify the screenshot target and report both shell and scrollport scroll metrics. This normalization is harness mechanics only; it does not promote visual-snapshot refresh authority into the current conformance profile.

### 15.1 Secret Pattern Table

| Secret class               | Match rule                                                                          | Redaction token                      |
| -------------------------- | ----------------------------------------------------------------------------------- | ------------------------------------ |
| Passwords                  | Exact or anchored variable/key names for `PASSWORD`, `PASS`, `PWD`, or equivalent credential suffixes | `[REDACTED:password]`                |
| Tokens                     | Exact or anchored variable/key names for `TOKEN`, `JWT`, `BEARER`, cookie headers, or equivalent credential suffixes | `[REDACTED:token]`                   |
| API or access keys         | Exact or anchored `API_KEY`, `ACCESS_KEY`, `SECRET_KEY`, or credential-context key names | `[REDACTED:key]`                     |
| DSNs/URLs with credentials | URL userinfo or DSN password segment present                                        | `[REDACTED:dsn]`                     |
| Object-store credentials   | S3-compatible access key or secret key variables                                    | `[REDACTED:object-store-credential]` |
| Private keys               | PEM private-key block markers                                                       | `[REDACTED:private-key]`             |

### 15.2 Artifact Redaction Table

| Artifact class                      | Redaction requirement                                                                                                      |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Tool/run/target/scheduler summaries | Redact before write.                                                                                                       |
| Machine stdout JSON                 | Redact before stdout.                                                                                                      |
| Captured child stdout/stderr logs   | Redact before retention.                                                                                                   |
| Service env files and env dumps     | Store only redacted credential values unless the file is required for child execution and kept under private runtime root. |
| Browser screenshots/videos/traces   | Diagnostic secret-bearing; not safe for publication.                                                                       |
| Visual geometry diagnostics         | Diagnostic secret-bearing; not safe for publication.                                                                       |
| Playwright HTML reports             | Diagnostic secret-bearing; not safe for publication.                                                                       |
| CI logs                             | Redact using the same token rules before harness-controlled emission.                                                      |

## 16. Integration with Product Specifications

**TH-HARNESS-REQ-650**
Harness tests MAY reference product requirement IDs, phase ledgers, coverage ledgers, phase maps, and acceptance criteria. Those references provide evidence routing only. They MUST NOT redefine product behavior under test.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-016

**TH-HARNESS-REQ-651**
Support-only tests, helper phases, raw aggregate suites, and direct package scripts MUST NOT be counted as authoritative product conformance evidence unless a canonical Make target adopts them and emits required evidence artifacts.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-652**
Load, login bursts, service resets, artificial stress margins, and browser harness stress tests are harness-only unless product specs explicitly adopt them. Browser login bursts MUST NOT be used as the sole evidence for Core 04 session-cap semantics when backend or integration evidence can prove victim selection and revocation delivery directly.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-653**
Timing-sensitive browser evidence for asynchronous socket behavior MUST prove the relevant sender readiness, receiver readiness, event identity, and diagnostic capture boundary before starting the measured interaction. A timed assertion MUST measure the product event under test, not page navigation, socket establishment, route cleanup, or waiter attachment.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-654**
Harness recovery evidence MUST be product-owner-subordinate. A harness target, phase row, retained artifact, task-surface entry, topology entry, scheduler work unit, or runtime-binary injection rule MUST NOT redefine recovery CLI authorization, logical command grammar, result schema, progress schema, timeout defaults, exit-code mapping, safe-output rules, restore-target preflight, journal behavior, or public route absence.

Any row claiming Core 04 AC-402, AC-427, or AC-428 MUST cite the Core-owned recovery requirements it verifies and MUST treat the recovery CLI as `deployment_admin`-irrelevant. A test that requires `deployment_admin`, session cookies, bearer tokens, CSRF, browser Origin, incident role, common-job authorization, WebSocket authorization, or public HTTP route access for recovery CLI invocation MUST NOT count as recovery conformance evidence.

Legacy tests that characterize a repository implementation requiring `deployment_admin` for recovery commands are negative evidence only. They MUST NOT close product-conformance rows, MUST NOT use blocked or stale status as current closure, and MUST NOT verify AC-402, AC-427, or AC-428.

Verified by: TH-HARNESS-AC-043

**TH-HARNESS-REQ-655**
Operator recovery conformance rows MUST map implementation-owned executable or wrapper behavior to exactly one Core logical command from Core 01 §12.2.1 before product assertions execute.

The current valid conformance command set is closed:

| Logical command | Operation token |
| --- | --- |
| `operator backup inspect latest` | `backup_inspect_latest` |
| `operator backup create` | `backup_create` |
| `operator restore latest` | `restore_latest` |
| `operator restore-verify latest` | `restore_verify_latest` |
| `operator restore-verify due` | `restore_verify_due` |

Compatibility aliases are not a second public recovery contract. A recovery invocation that is not one of the five canonical logical commands is negative evidence only and MUST fail current conformance routing.

Verified by: TH-HARNESS-AC-044

**TH-HARNESS-REQ-656**
Restore-workbook-probe evidence MUST cite a product owner for the probe behavior. Phase 10 recovery evidence MAY use a restore-workbook-probe fixture only when the probe behavior is owned by Core 01 or an adopted recovery NLSpec and the test row cites that owner.

The harness MAY route, schedule, and retain evidence for the probe, but it MUST NOT define workbook query semantics, selected workbook surfaces, source-row eligibility, pass/fail reason codes, or operator error mapping by itself. If no owner-defined probe contract exists for a claimed row, the harness MUST report the row as blocked or unsupported rather than inventing semantics from test fixtures, filenames, or package names.

Verified by: TH-HARNESS-AC-047

**TH-HARNESS-REQ-657**
Network Flow fixture manifests are evidence-routing and byte-freeze artifacts. They MAY cite adopted Network Flow, Core, Graph Projection, and harness requirements, but they MUST NOT define product behavior independently of those owners. A Network Flow conformance row that cites a manifest selector MUST also cite the adopted owner requirement or acceptance criterion that defines the expected route, import, lifecycle, cursor, graph, indicator, authorization, audit, redaction, or retention behavior. Missing owner citation, draft owner status, or selector-only semantics MUST leave the affected row blocked or unsupported.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-658**
Network Flow fault controls are harness mechanics for exercising adopted product-owned commit, worker, cancellation, replay, and recovery requirements. A fixture or target MAY use `cartulary.test.network_flow_fault_control.v1` evidence only when the same row cites the adopted Network Flow, Core, or Graph Projection owner requirement that defines the expected state after the injected fault. Fault-control boundary names, fault kinds, correlation keys, and route responses MUST NOT be cited as independent product semantics, public API compatibility, Core 05 publication evidence, or performance evidence.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-659**
Network Flow clock-control evidence is harness mechanics for exercising adopted product-owned time behavior. A fixture or target MAY use `cartulary.test.clock_control.v1` evidence only when the same row cites the adopted Network Flow, Core, or Graph Projection owner requirement that defines the expected timestamp, cursor TTL, retention, rotation, or timezone result. Clock route responses, wall-clock mode names, offsets, and fixed timestamps MUST NOT be cited as independent product semantics, public API compatibility, Core 05 publication evidence, or performance evidence.
Verified by: TH-HARNESS-AC-051

**TH-HARNESS-REQ-660**
Network Flow deterministic-randomness evidence is harness mechanics for exercising adopted product-owned identity, nonce, digest, collision, ordering, and replay behavior. A fixture or target MAY use `cartulary.test.network_flow_randomness_control.v1` evidence only when the same row cites the adopted Network Flow, Core, Graph Projection, or security owner requirement that defines the expected product result. Stream names, value kinds, collision values, response counts, and fail-closed exhaustion behavior MUST NOT be cited as independent product semantics, public API compatibility, Core 05 publication evidence, production secret management, or performance evidence.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-661**
Network Flow auth-transition evidence is harness mechanics for exercising adopted product-owned route authorization, hidden-resource, cursor recheck, and extension-resource invalidation behavior. A fixture or target MAY use `cartulary.test.network_flow_auth_transition_control.v1` evidence only when the same row cites the adopted Network Flow, Core 03, Core 04, Graph Projection, or security owner requirement that defines the expected product result. Boundary names, transition kinds, safe fixture refs, hidden-response kinds, correlation keys, and route responses MUST NOT be cited as independent product semantics, public API compatibility, production authorization policy, Core 05 publication evidence, or performance evidence.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-662**
Network Flow audit-assertion evidence is harness mechanics for exercising adopted product-owned domain audit occurrence counts, transactional audit boundaries, exact idempotency replay behavior, and no-audit failure cases. A fixture or target MAY use `cartulary.test.network_flow_audit_assertion_control.v1` evidence only when the same row cites the adopted Network Flow, Core 04, Core 01, or security owner requirement that defines the expected product result. Assertion kinds, event codes, safe fixture refs, baseline counts, expected final counts, replay increments, correlation keys, and route responses MUST NOT be cited as independent product semantics, public API compatibility, audit storage design, Core 05 publication evidence, or performance evidence.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-663**
Network Flow structural-accounting evidence is harness mechanics for failing
closed before generated contracts, fixture selection, or Phase 12 row evidence
are accepted. The current accounting manifest MUST validate all of the
following: Network Flow fixture and acceptance IDs are contiguous and present
in the owner spec and adoption tracker; the owner spec's Table 1-B dependency
registry has exactly the expected dependency rows and no `TODO:` locator cells;
the `network-flow` contract family is present in the contract-family registry;
planned contract status has the declared activation dependencies and no
generated Network Flow symbols; active contract status has generated Network
Flow symbols and no remaining activation dependencies; generated-drift scratch
inputs include the accounting and contract owner inputs; and required public
Make targets exist. This accounting MUST NOT define Network Flow product
behavior, activate generated outputs, invent fixture semantics, or close a
product conformance row without the product owner citations required by
TH-HARNESS-REQ-657 through TH-HARNESS-REQ-662.
Verified by: TH-HARNESS-AC-055

**TH-HARNESS-REQ-664**
Platform harness runtime owns generic guarded-route authorization, control
contribution registration, reset-hook orchestration, and centralized
redaction. Product- or extension-specific control registries, request
validation, pending state, consume semantics, and product dependency adapters
MUST be owned by the corresponding module in a runtime-scanned package.

A module control contribution MUST provide only its guarded route registrars,
reset hook, and explicitly typed dependency adapters to the generic platform
boundary. Binary composition MAY register that contribution only when test
routes are enabled. Moving a control between implementation packages MUST
preserve its route path, schema ID, guard ordering, disclosure behavior,
redaction, pending-state conflict rules, one-shot consumption, and runtime
reset behavior. Test-support scan exclusions MUST NOT hide a package compiled
into the server binary.

Verified by: TH-HARNESS-AC-050, TH-HARNESS-AC-052, TH-HARNESS-AC-053,
TH-HARNESS-AC-054, TH-HARNESS-AC-056

## 17. Acceptance Criteria / Definition of Done

The acceptance matrix is the harness Definition of Done. Each row is binary. A row passes only when its setup, invocation, exit/status, stdout/stderr, artifact, and cleanup expectations all match.

| ID                | Requirement owner  | Scope                            | Setup fixture                                                                | Invocation                                                                                | Expected exit/status                                           | Stdout                                                 | Stderr                                                       | Required artifacts                                                                                 | Negative case                                                                | Cleanup expectation                                     |
| ----------------- | ------------------ | -------------------------------- | ---------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------- |
| TH-HARNESS-AC-000 | Section 8          | Schema validation                | Any public target that emits required JSON                                   | Target named by the fixture                                                               | Success only if JSON validates                                 | Per Section 7                                          | Per Section 7                                                | Every emitted required JSON artifact validates against Section 8 schema attachments                | Inject schema-invalid required summary                                       | No extra cleanup beyond target contract                 |
| TH-HARNESS-AC-001 | Sections 1, 4      | Command registry                 | Current tree                                                                 | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` plus registry parity checker    | `0` when registry matches exactly                              | Bounded report                                         | Empty on success                                             | Public target registry parity report                                                               | Extra/missing public target fails                                            | none                                                    |
| TH-HARNESS-AC-002 | Section 5          | Config precedence                | Fixture target with CLI, Make var, env, manifest, config, default candidates | Dedicated config resolver test target or unit harness                                     | `0`                                                            | Machine or bounded summary                             | Empty on success                                             | Resolver summary showing CLI > Make var > env > manifest > config file > default                   | Non-positive scheduler limit exits `2` with `configuration_error`            | no child work                                           |
| TH-HARNESS-AC-003 | Sections 5, 6      | Result root and run ID           | No child work required                                                       | Invalid result root, invalid run ID, and unsafe custom result root fixtures                | `2`                                                            | Empty or failure JSON according to target output class | Bounded config diagnostic                                    | Failure summary when wrapper starts; retained root preflight rejects unsafe custom permissions      | Slash, backslash, whitespace, `.`, `..`, existing non-empty run dir, world-writable custom root all fail | no child work                                           |
| TH-HARNESS-AC-004 | Sections 7, 8      | Machine output accepted          | Toolchain ready; explicit result root/run ID                                 | `CARTULARY_OUTPUT_MODE=machine make backend-unit`; `... make test-fast`; `... make check` | Target status                                                  | Exactly one JSON object plus LF                        | Empty after wrapper starts                                   | `cartulary.tool_run_summary.v3` and target artifacts                                               | Progress prose or duplicate JSON fails                                       | normal target cleanup                                   |
| TH-HARNESS-AC-005 | Section 7          | Machine output rejected          | No child work                                                                | `CARTULARY_OUTPUT_MODE=machine make clean`; `... make dev`; `... make help`               | `2`                                                            | Empty                                                  | Bounded `usage_error` diagnostic                             | None required                                                                                      | Child work starts despite rejection                                          | no deletion or service start                            |
| TH-HARNESS-AC-006 | Section 10         | Scheduler determinism            | Controlled manifest with simultaneous child completions and scheduled browser groups | Run scheduler fixture twice with same manifest; validate generated browser worker-admin slot ranges | `0`                                                            | Bounded summary or machine object                      | Empty on success                                             | Byte-identical scheduler events after dynamic timestamp normalization allowed only by schema rules; browser group worker slots are explicit, contiguous, and non-overlapping | Event sequence differs; browser worker slot env is missing or overlaps       | finalizers run                                          |
| TH-HARNESS-AC-007 | Section 11         | Service modes                    | Owned and attach fixtures                                                    | Owned service target; attach target missing one required var                              | owned success; attach failure `2`                              | Bounded summary                                        | Empty on owned success; config diagnostic for attach failure | Owned lease before child work; attach failure summary                                              | Attach mode deletes container-level resource                                 | owned teardown recorded                                 |
| TH-HARNESS-AC-008 | Section 12         | Test-only harness routes         | Browser test runtime with test route token and saved-view fixture inputs     | Reset route success, saved-view system fixture success, auth rejection, origin/host rejection, concurrent reset, timeout, partial failure fixtures | Expected HTTP statuses from Section 12                         | HTTP JSON response                                     | n/a                                                          | Reset response validates schema; saved-view fixture response is a normal saved-view resource with `scope='system'`; tainted stack marker on partial failure; no permissive CORS headers | Default runtime exposes any test route, wrong host/origin reaches mutation, product auth bypasses the test token, saved-view fixture accepts caller-supplied scope/owner/identity, or wildcard CORS is emitted | tainted stack restarted before further work             |
| TH-HARNESS-AC-009 | Section 13         | Cleanup and destructive reset guard | Synthetic registry with safe and unsafe paths; fake Compose, database, migration, and object-store commands | Cleanup guard unit; `CARTULARY_CLEANUP_DRY_RUN=1 make clean`; dry-run and missing-confirmation invocations for `services-down`, `db-reset`, and `object-store-reset` | `0` for safe dry-run; nonzero for unsafe synthetic path or missing destructive confirmation | Dry-run lines match format                             | Bounded guard or confirmation diagnostic before mutation      | Candidate list, guard evidence, and command-shape evidence for confirmed local resets                              | Empty path, `/`, `.`, `..`, traversal, protected root, outside-repo path, symlink-following, inherited-env-only destructive confirmation, object-store reset touching another bucket, or `services-down` removing volumes accepted | no deletion, service start, or service stop in dry-run  |
| TH-HARNESS-AC-010 | Section 13         | Stale janitor proof gates        | Fake DB, bucket, container, and browser fixtures with/without proof          | Focused stale-janitor tests                                                               | `0`                                                            | Bounded summary                                        | Empty on success                                             | Evidence that unproven resources retained and proven stale fixtures deleted only outside dry-run   | Resource lacking generated name/proof deleted                                | unproven resources retained                             |
| TH-HARNESS-AC-011 | Section 15         | Redaction                        | Fake DSN, object-store secret, token, header, cookie, CLI arg, nested JSON, structural session fields, and private-key fixtures | Redaction unit plus one wrapper log capture                                               | `0`; redaction/write failure exits `11` unless Section 9.1 preserves an earlier primary failure | No unredacted secret in machine JSON                   | No unredacted secret in captured stderr                      | Summaries/logs contain required redaction tokens, preserve schema-owned object/array fields, and use owner-read/write file modes | Any secret pattern appears unredacted, required structural fields are replaced by redaction tokens, or required retained log is group/world-readable | none                                                    |
| TH-HARNESS-AC-012 | Section 14         | Platform matrix                  | Platform claim checker fixture                                               | Platform matrix checker                                                                   | `0` for allowed profiles; nonzero for unsupported claim        | Bounded summary                                        | Diagnostic on unsupported claim                              | Matrix report                                                                                      | macOS/Windows-native/Podman claimed as current conformance                   | none                                                    |
| TH-HARNESS-AC-013 | Sections 9, 16     | Product versus harness failure   | One known failing assertion, one harness setup failure, and one browser strict-port conflict fixture | Canonical test target under each fixture                                                  | Product failure exits `10`; setup failure exits Section 9 code; strict-port conflict exits `4` | Failure headline names class and reason                | Bounded diagnostic                                           | Target/tool summary with failure class and reason; strict-port conflict startup diagnostics, phase summary, target summary, scheduler summary, and public exit agree on `resource_conflict` | Setup failure classified as product, or strict-port conflict classified as `service_start_error` | harness cleanup attempted                               |
| TH-HARNESS-AC-014 | Section 9          | Exit-code matrix                 | Controlled failure fixtures                                                  | Exit matrix test target                                                                   | Exact Section 9 code for every class                           | Per output mode                                        | Per output mode                                              | Failure summaries with primary failure selection                                                   | Cleanup failure overrides earlier product failure                            | cleanup failure recorded but primary exit preserved     |
| TH-HARNESS-AC-015 | Sections 6, 8      | Retained artifact identity       | Explicit result root/run ID                                                  | `CARTULARY_TEST_RESULTS_DIR=<dir> CARTULARY_TEST_RUN_ID=<id> make backend-unit`           | `0`                                                            | Summary names run root                                 | Empty                                                        | Artifacts under `<dir>/<id>` with target, run ID, run root; retained run roots and target dirs are owner-only on POSIX hosts | Newest-run fallback accepted as proof or retained directories are group/world-accessible | custom absolute result root not removed by `make clean` |
| TH-HARNESS-AC-016 | Sections 1, 2, 18, 19 | Editorial and boundary closure | Revised document                                                             | Editorial lint, open-delegation scanner, and future-decision scanner                       | `0`                                                            | Bounded summary                                        | Empty on success                                             | No prohibited evidence markers in Sections 1-17; no adopted-section open delegation phrase without a cited closed table, schema attachment, algorithm, or diagnostic boundary; no current blockers in Section 19 | Current-profile blocker appears in future section, or an adopted section uses open delegation without a cited closure owner | none                                                    |
| TH-HARNESS-AC-017 | Section 11         | Lifecycle-machine conformance    | Service-suite fixtures for happy path, startup failure, interrupted child, cleanup failure, illegal transition, and crash/rerun | Lifecycle-machine conformance target or unit harness                                      | Happy path `0`; failure fixtures use exact Section 9 code      | Bounded summary or machine object                      | Empty on happy path; bounded diagnostic on failure fixture | `cartulary.test_services.lifecycle.v1` stream with sequential events, valid transitions, terminal state, Section 9 failure mapping, and cleanup proof behavior | Unlisted `(state,event)` mutates state, terminal state accepts later event, or lifecycle stream validates with a sequence gap | normal suite cleanup; unproven resources retained       |
| TH-HARNESS-AC-018 | Sections 4, 10, 11 | Warm scheduler health            | Retained warm-ready `check` fixture plus over-budget, cold-provisioning, measurement-in-default-check, skewed-lane, shared-browser-session, unexpected-reuse, and fixture-budget fixtures | `make scheduler-summary-timing-drift RESULTS_DIR=<dir> TARGET=check SCHEDULER_WARM_CHECK_BUDGET_MS=60000 SCHEDULER_WARM_CHECK_BALANCE_RATIO=1.25` | Success only for warm-eligible, in-budget, balanced fixtures                    | Bounded summary                                        | Bounded diagnostic on failure fixture                  | Scheduler and target summaries identify `check-service-backed` wall time, evaluated lanes, browser session groups, unexpected reused accounting, readiness attribution, and fixture class counts | Measurement stage, undeclared extra browser session, hidden provisioning, unexplained reused work, unplanned clone/reset, or skewed non-isolated lane passes unnoticed | none                                                    |
| TH-HARNESS-AC-019 | Section 8.2        | Agent finalizer                  | Fake Make fixture plus valid, missing, failed, incomplete, contaminated, non-warm retained run, action-cache hit/miss/disabled/corrupt/input-change/output-change fixtures | `make agent-finalize`; `RESULTS_DIR=<dir> make agent-finalize`; `CARTULARY_OUTPUT_MODE=machine make agent-finalize` | Success for coherent maintenance inputs; fail-fast for first failed action substep or invalid retained run; cache hits only for eligible closed-profile actions | One `[FINALIZE]` line then bounded result/artifact lines; machine emits one JSON object | Bounded failure diagnostic naming failed action/substep | `agent-finalize/finalize-summary.json`, per-action `execution_state` and cache state, child summaries/logs when executed, and `finalize_summary` artifact ref | Excluded targets run, mutation starts after invalid `RESULTS_DIR`, cache hit bypasses retained-run validation, corrupt cache produces success, machine output requires log parsing, semantic action IDs are absent, or skipped-after-failure work is absent | No cleanup or destructive command is run               |
| TH-HARNESS-AC-020 | Section 4          | Public target semantic value     | Current target registry plus one synthetic shallow-wrapper fixture           | Registry semantic-value checker                                                           | Success only when every public target declares at least one semantic behavior and every declared behavior has an owner section | Bounded report                                         | Empty on success                                             | Semantic-value parity report                                                                    | Target with only child command aliases and no semantic behavior passes        | none                                                    |
| TH-HARNESS-AC-021 | Sections 5, 8, 10   | Frontend-unit harness stability  | Current check topology plus delayed jsdom workbook-row, row-history selector, controlled-input helper fixtures, and Vitest failure sidecar fixtures | Topology contract tests; `make frontend-unit`; constrained `CHECK_HOST_CPU_JOBS=2 make check` | Success when scheduled Vitest workers and resource claims match, shared helpers tolerate bounded async hydration, and sidecar diagnostics are preferred when present | Bounded summary                                        | Empty on success                                             | Scheduler manifest shows `frontend-unit` `host_cpu=2` and `VITEST_MAX_WORKERS=2`; frontend helper and selector-policy tests pass; `frontend-unit` retains `cartulary.vitest_failure_details.v1` when `runner.json` exists; synthetic `STACK_TRACE_ERROR` plus sidecar reports the sidecar assertion message | `frontend-unit` can run more workers than it claims, row waits use unbounded/ad hoc selectors, row-history waits use display labels instead of `history_item_ref`, helper diagnostics omit mounted row identity, sidecar schema drift is accepted, or `STACK_TRACE_ERROR` replaces an available sidecar assertion message | none                                                    |
| TH-HARNESS-AC-022 | Sections 1, 4, 5, 8, 10 | Frontend readiness harness metadata | Current frontend registry/maps plus malformed namespace, missing a11y target, raw-only a11y artifact, missing visual fixture, stale row-accounting fixtures, support-specimen substitution, concept-image substitution, explicit base-phase join fixtures, and removed-preflight target fixtures | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`; JSON shape checks; `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P3`; ambiguity fixture for `PHASE=FE-P3`; a11y, visual fixture, row-accounting schema, and frontend registry join validation | Success only when the public target, frontend namespace, frontend maps, frontend registry base-phase joins, visual fixture registry, normalized a11y summary, and scoped row-accounting closure are all valid | Bounded report or explain output | Empty on success; bounded usage diagnostic for ambiguous namespace fixture | Public registry includes `browser-e2e-a11y` and excludes `browser-e2e-a11y-preflight`; `cartulary.frontend_phase_registry.v5` admits active current phases only with `row_rollup_state=active_green`; `cartulary.frontend_phase_test_map.v4` owns structured implemented `FE-*` rows; `cartulary.frontend_visual_fixture_registry.v3` owns append-only fixture identity and screenshot capture scope; `FE-VFIX-01` closes only from the exact default Timeline workbook-shell scenario and Playwright app-owned screenshot; FE-P3/FE-P11 support specimens and external concept images cannot close default-shell evidence; `cartulary.frontend_row_accounting.v4` closes mapped rows only within explicit accounting scope; `cartulary.frontend_accessibility_summary.v2` validates for implemented rows; frontend ledger artifacts remain generated | Filename-only or title-only `FE-*` row claims pass, `PHASE=FE-P3` without namespace is accepted, blocked rows appear as completed row evidence, missing visual fixture TODOs pass, stale row accounting closes a row, base `PHASE=phaseN` slices enforce unrelated `FE-*` rows, frontend/base phase joins are inferred from numeric suffixes, support specimens or concept images close default-shell evidence, raw third-party a11y output substitutes for the summary, or removed preflight target references pass | no child browser work required for metadata fixtures |
| TH-HARNESS-AC-023 | Sections 4, 7      | Registry output-class and side-effect parity | Current NLSpec registry, `tools/task_surface_manifest.json`, rendered task-surface report, and output-class matrix | Registry parity checker and malformed registry fixtures | Success only when target membership, output class, artifact policy, schema policy, and side effects agree exactly | Bounded parity report | Empty on success; bounded drift diagnostic on failure | Public registry parity report with side-effect and output-class comparison | Missing `side_effects[]`, `none` plus another class, undeclared side effect, or output-class drift passes | none |
| TH-HARNESS-AC-024 | Section 10         | Scheduler command-shape closure | One live generated scheduler fixture for each required command type plus malformed command descriptors, prerequisite-policy fixtures, and artifact-consumer dependency fixtures, with optional command types validated when present | Scheduler manifest schema and shape checks | Success only for closed command shapes, required and closed `make_prerequisite_policy` values for `make_target`, and declared artifact producer/consumer DAG edges | Bounded summary | Empty on success; bounded shape diagnostic on failure | Shape-check evidence naming command type, required fields, forbidden fields, wrong-type failures, prerequisite policy, and artifact dependency parity | Missing required field, forbidden field, unknown command type, wrong field type, omitted or invalid prerequisite policy, or missing build-artifact producer edge passes | no child work |
| TH-HARNESS-AC-025 | Section 8          | Schema attachment closure | Every Section 8 `present` schema attachment plus positive and negative fixtures | Schema attachment policy check and schema fixture validation | Success only when every present attachment exists, parses, is top-level closed or extension-container closed, and validates fixtures | Bounded summary | Empty on success; bounded schema diagnostic on failure | Schema attachment report | Missing schema, malformed schema, open top-level schema without extension container, or fixture-blind schema passes | none |
| TH-HARNESS-AC-026 | Sections 1, 16     | Core test-route traceability | Core 04, Appendix F, frontend phase maps, and generated ledgers | Requirement-ID uniqueness check and `REQ-04-109` citation classifier | Success only when `REQ-04-109` means test-only runtime-control route security and public-origin behavior cites `REQ-04-110` | Bounded traceability report | Empty on success; bounded traceability diagnostic on failure | Duplicate-ID report and citation classification report | Public route, WebSocket origin, evidence-handle, or deployment-origin row cites `REQ-04-109` | none |
| TH-HARNESS-AC-027 | Section 4          | Public registry source-of-truth parity | Current public Make surface, task-surface manifest, execution topology, generated Make include, scheduler reachability, projection metadata, and prose registry | Registry source-of-truth parity checker | Success only when every public target exists in the NLSpec registry and `tools/task_surface_manifest.json`, no removed target appears in public generated mirrors, no public target is introduced only by topology, generated Make, or prose, every advertised `check` inclusion is reachable as direct work/aggregate/projection, and support-only internal targets do not advertise default `check` unless scheduled | Bounded parity report | Empty on success; bounded drift diagnostic on failure | Public-source parity report with default-check projection labels and removed-target exclusions | Public target appears only in execution topology, generated Make includes, or prose; removed target remains public; full direct browser target is counted when only a smoke projection ran; support-only target advertises unselected `check` membership | none |
| TH-HARNESS-AC-028 | Sections 4, 5, 8, 13 | Local cache profiles | Cache helper fixture plus representative readiness, build, finalizer, and render-cache fixtures | Cache-helper smoke tests; cold/hot readiness and build target runs; `make agent-finalize`; render drift checks; cleanup fixture | Success only when first run misses, second run hits, disable/force/corrupt/missing-output cases execute or fail safely, summaries remain emitted, and scheduler accounting reports no undeclared reuse | Bounded summary | Empty on success; bounded diagnostic on invalid cache state | Cache records validate against Section 8 schemas; run-root cache artifacts show state/reason/record path; public summaries remain present | Security, drift, service readiness, runtime reset, cleanup, destructive guard, browser/service-backed live-state work, or aggregate success is accepted by cache reuse; missing output succeeds by reuse; cache artifact is cited as product evidence | `make clean` preserves default cache roots; `make distclean` removes default cache roots |
| TH-HARNESS-AC-029 | Sections 1, 5     | Public input matrix closure | Current NLSpec input matrix and task-surface metadata | Input-contract parity checker | Success only when every public target input in task-surface metadata appears in the NLSpec matrix and every NLSpec matrix row mirrors metadata or an explicitly documented NLSpec default override | Bounded parity report | Empty on success; bounded drift diagnostic on mismatch | Public input matrix parity report | Public target accepts an input absent from the NLSpec matrix, or NLSpec row omits default, bound, empty-string, invalid, summary, or forwarding behavior | none |
| TH-HARNESS-AC-030 | Section 10        | Scheduler defaults and auto policies | Fixture schedules for each resource and auto policy, plus override and impossible-resource fixtures | Scheduler resource-resolution tests | Success only when every fixed default, override bound, auto formula, omission rule, and impossible-resource failure matches Section 10 | Bounded summary | Empty on success; bounded diagnostic on mismatch | Resource-limit source and resolved-limit evidence in scheduler summaries | `auto` resolves differently, override above max passes, omission lacks default/auto behavior, or impossible resources start child work | no child work beyond fixture scheduler |
| TH-HARNESS-AC-031 | Section 8         | Scheduler diagnostic artifact closure | Current scheduler artifact families plus positive and negative pressure-summary and fixture-tier proof fixtures | Schema/artifact policy checker and schema fixture validation | Success only when `pressure-summary.json` is schema-owned by present `cartulary.scheduler_pressure_summary.v4`, embeds retained proof objects that validate as `cartulary.fixture_tier_proof.v1`, validates before scheduler success, and rejects unknown or missing required fields | Bounded report | Empty on success; bounded diagnostic on mismatch | Artifact policy report naming pressure-summary status and schema fixture results | Missing `cartulary.scheduler_pressure_summary.v4` or `cartulary.fixture_tier_proof.v1` schema, open top-level pressure-summary fields, open fixture proof fields, or unvalidated scheduler success passes | none |
| TH-HARNESS-AC-032 | Section 9         | Primary failure determinism | Simultaneous failure fixtures covering class, lifecycle, event, target, path, and reason ties | Exit matrix and primary-failure unit tests | Success only when selected primary failure and public exit follow Section 9.1 and TH-HARNESS-REQ-304 exactly | Bounded summary | Bounded diagnostic on mismatch | Failure summary with ordered candidate failures and selected primary | Cleanup overrides earlier non-cleanup failure, or tie order differs across runs | cleanup failure recorded when fixture creates one |
| TH-HARNESS-AC-033 | Section 11        | Concurrent lifecycle | Service-suite lifecycle fixtures with overlapping child work, duplicate child start, unknown child finish, and interruption | Lifecycle-machine conformance target or unit harness | Success only when active-child counts transition legally and illegal duplicate/unknown events fail closed | Bounded summary or machine object | Empty on happy path; bounded diagnostic for illegal fixtures | Lifecycle stream with `active_child_count`, child keys, legal transitions, and terminal state | Concurrent child start is rejected, unknown finish mutates state, duplicate start passes, or active count becomes negative | normal suite cleanup; unproven resources retained |
| TH-HARNESS-AC-034 | Section 12        | Runtime reset closure | Reset fixtures with public mutable tables, migration metadata, bootstrap admin, route idempotency, object-store objects, partial failure, and pending public-error fault | Test-runtime reset route tests | Success only when selected table ordering, migration metadata preservation, bootstrap restoration, route-idempotency clearing, object cleanup, partial-failure response, and fault clearing match Section 12 | HTTP JSON response | n/a | Reset response validates schema and includes sorted `tables_reset`, post-reset counts, object count, partial failure fields when applicable | Unsorted or incomplete table set, migration metadata truncation, missing bootstrap admin, route idempotency residue, uncleared fault, or object residue passes | tainted stack restarted before further work |
| TH-HARNESS-AC-035 | Section 12        | Test-route edge closure | Weak token, malformed token, missing/wrong header, pending-fault conflict, consumed-fault retry, and reset-clears-fault fixtures | Test-only route unit/integration tests | Expected HTTP status or startup/config failure for every Section 12 token and fault edge case | HTTP JSON response where route starts | n/a | Error envelopes use exact `test_route_forbidden`, `test_public_error_fault_already_armed`, or startup `configuration_error` as applicable | Weak token starts, product auth bypasses token, second fault replaces pending fault, consumed fault remains armed, or reset does not clear fault | runtime reset clears fault state |
| TH-HARNESS-AC-036 | Sections 13, 15   | Cleanup and redaction closure | Protected-root cleanup fixtures, cleanup-owned child paths, structured secret keys, raw secret text, and structural field names | Cleanup guard and redaction tests | Success only when protected-root attempts fail, cleanup-owned paths succeed or no-op when missing, exact key/raw-text redaction applies, and schema-owned structures are preserved | Bounded summary | Empty on success; bounded diagnostic on mismatch | Cleanup guard report and redacted summary/log fixtures | Protected root deletion passes, missing cleanup-owned path fails, secret leaks, or structural fields are over-redacted | no deletion outside cleanup-owned fixtures |
| TH-HARNESS-AC-037 | Sections 5, 8, 10 | Operator runtime binary injection | Current topology plus missing, non-executable, digest-mismatch, undeclared-input, and raw-Go fallback fixtures | `make build-operator`; scheduler-selected operator Go work; `make check OPERATOR_BIN=/tmp/x`; `make check CARTULARY_OPERATOR_BIN=/tmp/x` | Producer succeeds with declared output path; consumer fails before product assertions for invalid injection; undeclared public inputs exit `2`; digest mismatch exits `11` | Bounded summary | Empty on success; bounded config/artifact diagnostic on mismatch | `build-operator` tool summary and build-artifact cache artifact; operator aggregate `runtime-binaries.json`; Go runner logs contain no nested `make build-operator` for scheduler-selected operator work | Hidden nested operator builds, arbitrary child forwarding, missing producer dependency, missing runtime-binary provenance, or binary cache hit marked as scheduler `reused` passes | no extra cleanup beyond build output contract |
| TH-HARNESS-AC-038 | Section 4.1A      | Helper ownership registry | Current helper registry plus allowed facade and forbidden backend/browser/scheduler-internal import fixtures | Harness import-boundary contract tests | Success only when every helper family listed in the helper registry is classified exactly once and no unclassified private backend, browser, scheduler, duration-accounting, phase-accounting, execution, diagnostics, generated-artifact, or finalization helper is imported by non-owner code | Bounded report | Empty on success; bounded import-boundary diagnostic on mismatch | Import-boundary report naming helper family, facade key, source, and target | Non-owner code imports private implementation helpers after a facade exists, or a helper family is unclassified or duplicated | none |
| TH-HARNESS-AC-039 | Section 4.1A      | Compatibility paths | Current tree plus unsupported legacy path fixtures | Caller inventory, import-boundary contract tests, and relevant public target characterization | Success only when unsupported legacy backend, frontend, and scheduler catch-all paths are absent as compatibility paths, no temporary redirects are accepted, no new caller imports them, generated/task-surface metadata does not reference them, and public Make behavior remains unchanged | Bounded report | Empty on success; bounded unsupported-private diagnostic on mismatch | Caller inventory and import-boundary report for unsupported private paths | New caller imports an unsupported private path, generated metadata points at a legacy scheduler path, temporary redirects are accepted, or public behavior changes after deletion | none |
| TH-HARNESS-AC-040 | Section 4.1A      | Govulncheck findings ownership | Govulncheck JSON stream fixtures covering no findings, package/module findings, symbol findings, redaction, and malformed JSON | Static-analysis security findings tests and `go-vulncheck` when toolchain is ready | Success only when static-analysis/security ownership proves identical normalized findings, redaction, exit mapping, and artifact behavior after helper movement | Bounded summary | Empty on success; bounded parse/security diagnostic on failure | `govulncheck-findings.json` validates `cartulary.govulncheck_findings.v1` with deterministic finding order | Backend path remains a supported import, redaction changes, symbol findings stop blocking, or exit mapping drifts | temp files removed |
| TH-HARNESS-AC-041 | Section 4.1A      | Migration/schema validator ownership | Migration history, schema object ownership, scratch migration, and JSON-shape fixtures | `json-shape-check`, `migration-drift`, and database-contract drift tests | Success only when `json-shape-check` and `migration-drift` retain manifest schema validation, scratch apply behavior, diagnostics, and failure classification after helper movement | Bounded summary | Empty on success; bounded drift diagnostic on failure | Migration/schema manifest validation summaries and migration drift retained artifacts | Schema IDs change, diagnostics drift unexpectedly, scratch DB cleanup changes, or validators become backend execution behavior owners | scratch DB cleanup per migration-drift contract |
| TH-HARNESS-AC-042 | Section 4.1A      | Duration retained-run safety | Coverage, drift, update, and finalizer retained-run fixtures including failed, partial, missing-artifact, stale, contaminated, ambiguous, valid drift, and valid full warm check roots | Duration baseline and finalizer tests | Success only when coverage is read-only, drift is read-only, update rejects invalid retained evidence before mutation, and `agent-finalize` accepts only valid retained full warm `make check` roots for mutating duration refresh | Bounded summary | Empty on success; bounded retained-run diagnostic on failure | Baseline files unchanged for invalid evidence; finalizer summary records retained-run validation before mutation | Mutating update starts before retained-run validation, partial run is accepted for finalizer refresh, invalid evidence writes baselines, or drift mutates files | invalid fixtures leave tracked baselines unchanged |
| TH-HARNESS-AC-043 | Section 16        | Recovery evidence routing | Core recovery rows plus legacy `deployment_admin` recovery fixtures | Phase-map/evidence-routing validation and affected recovery target planning | Success only when recovery CLI invocation is classified as `deployment_admin`-irrelevant and legacy `deployment_admin` recovery tests are negative-only and cannot close AC-402, AC-427, or AC-428 | Bounded report | Empty on success; bounded routing diagnostic on mismatch | Evidence-routing report names row IDs, owner citations, evidence class, and negative legacy status where applicable | A `deployment_admin`-gated recovery CLI test counts as product conformance evidence or blocked/stale legacy status is accepted as current closure | none |
| TH-HARNESS-AC-044 | Section 16        | Canonical operator command mapping | Operator scenario rows, runtime-binary rows, and compatibility-alias fixtures | `make build-operator` plus scheduler-selected operator work or command-mapping validation | Success only when implementation-owned executable behavior maps to exactly the five Core logical commands, compatibility aliases are negative-only, and Core-owned final stdout, optional progress, timeout/default, target-config, and exit-code behavior validate | Bounded summary or machine result | Empty on success; bounded command-mapping diagnostic on mismatch | Operator runtime-binary provenance and retained operator result/progress artifacts for selected work | Old command names, compatibility aliases, or alternate JSON envelopes pass as a second public contract | no extra cleanup beyond build output contract |
| TH-HARNESS-AC-045 | Section 4.1A      | Migration-evidence classification | Migration manifest, embedded SQL source, goose ledger, schema-object ownership, and recovery phase-map fixtures | `json-shape-check`, database-contract drift validation, and phase-map classification checks | Success only when migration-history evidence routes through database-contract or migration-evidence ownership and cannot close operator-recovery conformance evidence | Bounded report | Empty on success; bounded classification diagnostic on mismatch | Migration/schema validation summaries and phase-row owner classification report | A migration-evidence operator wrapper is counted as AC-428 recovery evidence | none |
| TH-HARNESS-AC-046 | Section 4.1A      | Moved-test accounting | Test-path movement fixtures covering phase maps, task-surface inputs, topology inputs, generated ledgers, and runtime-binary rows | `make json-shape-check`; `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`; schedule or generated-drift checks when owner inputs change | Success only when owner inputs change before generated artifacts, generated outputs are regenerated through Make, and generated ledgers/manifests are not hand-edited | Bounded report | Empty on success; bounded accounting diagnostic on mismatch | Owner-input diff plus drift/schema summaries for the changed accounting surface | A moved test completes by hand-editing generated outputs or by changing runtime-binary/default-check/scheduler behavior without NLSpec coverage | none |
| TH-HARNESS-AC-047 | Section 16        | Restore-workbook-probe owner routing | Phase 10 restore-probe rows with owner-cited and owner-missing fixtures | Recovery evidence-routing validation and affected phase10 target planning | Success only when restore-workbook-probe evidence cites Core 01 or an adopted recovery owner and harness fixtures do not define probe semantics by themselves | Bounded report | Empty on success; bounded owner-routing diagnostic on mismatch | Probe evidence-routing report with owner citation or blocked/unsupported status | Harness-only probe semantics from fixtures, filenames, or package names close recovery evidence | none |
| TH-HARNESS-AC-048 | Section 8         | Same-run helper artifact closure | Helper aggregate fixtures plus valid, scheduler-reused, missing-digest, and old-run-scope schema fixtures | Run-summary helper fixture, `explain-run`, schema validation, and `make json-shape-check` | Success only when helper refs validate, resolve under the current run root, expose producer artifact digests and consumer refs, report helper reuse without scheduler `reused`, and fail closed for old-run or malformed refs | Bounded run summary and `explain-run` lines | Bounded schema/artifact diagnostic on mismatch | `cartulary.same_run_helper_artifact_ref.v1` retained under `_shared/same-run-helper-artifacts`; run summary links helper refs; pressure/scheduler reused counts remain current-profile zero | Old retained helper artifacts are accepted as fresh, helper reuse is reported as scheduler `reused`, malformed refs pass, or selected conformance rows are skipped | none |
| TH-HARNESS-AC-049 | Sections 8, 11, 16 | Network Flow fixture manifests | Positive frozen fixture manifest with committed source, expected, and transcript bytes plus negative fixtures for unsorted paths, missing files, symlink/traversal paths, digest mismatch, draft selection, and ownerless selector | Network Flow fixture-manifest validator, schema validation, `make json-shape-check`, and the Network Flow conformance target that selects the fixture | Success only when every selected Network Flow fixture manifest validates, hashes match exact bytes, committed fixture roots are not mutated, run-local materialization is used, frozen-only conformance selection is enforced, and owner refs route behavior to product owners | Bounded summary naming fixture IDs and bundle digests | Bounded schema/artifact diagnostic on mismatch | `cartulary.network_flow_fixture_manifest.v1` manifest validation evidence and retained execution summary with manifest SHA-256, source and expected bundle digests, materialized input root, produced artifact refs, comparison status, and blocked/unsupported rows for ownerless selectors | Fixture identity inferred from filename, unlisted file accepted, committed bytes mutated, missing owner semantics closes evidence, draft fixture selected as current evidence, or digest/order/path mismatch reaches product code | Run-local materialization removed by result-root cleanup; committed fixture bytes retained |
| TH-HARNESS-AC-050 | Sections 12, 16    | Network Flow fault controls | Disabled route, token/host/origin edge cases, invalid boundary/kind/correlation/error-code bodies, pending conflict, exact boundary/correlation consume, reset-clears-fault, and ownerless selector fixtures | Network Flow fault-control route tests, schema validation, and Network Flow fixture targets that select commit or worker fault controls | Success only when the test route is unavailable by default, guard failures happen before mutation, request validation is closed, one pending fault is consumed once by exact boundary and optional correlation key, reset clears pending faults, and fault evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.network_flow_fault_control.v1` route response plus fixture transcript naming boundary, fault kind, correlation scope, owner refs, pre/post state counts, and replay or recovery result where product work executed | Product auth bypasses the token, a second fault replaces a pending fault, wrong boundary or correlation consumes a fault, reset leaves a pending fault, or fault-control-only evidence closes a product row | runtime reset clears pending fault state |
| TH-HARNESS-AC-051 | Sections 12, 16    | Test clock controls | Disabled route, token/host/origin edge cases, fixed clock, offset clock, reset clock, read-only state, invalid payloads, runtime-reset-clears-clock, and ownerless Network Flow time selector fixtures | Test clock route tests, reset route tests, schema validation, and Network Flow fixture targets that select time-sensitive rows | Success only when test clock routes are unavailable by default, guard failures happen before mutation/disclosure, set accepts exactly one command, reset restores wall mode, state is non-mutating, runtime reset clears registered clock state, responses validate as `cartulary.test.clock_control.v1`, and clock evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.clock_control.v1` route response plus fixture transcript naming mode, fixed/offset selection, owner refs, pre/post state, and product time result where product work executed | Product auth bypasses the token, fixed time survives runtime reset, state mutates the clock, invalid payload changes clock state, or clock-control-only evidence closes a product row | runtime reset or explicit clock reset restores wall mode |
| TH-HARNESS-AC-052 | Sections 12, 16    | Network Flow deterministic randomness controls | Disabled route, token/host/origin edge cases, invalid stream/kind/value/exhaustion bodies, duplicate value collision fixture, same-stream pending conflict, fail-closed exhaustion, reset-clears-randomness, and ownerless selector fixtures | Network Flow randomness-control route tests, schema validation, and Network Flow fixture targets that select deterministic ID, nonce, digest, collision, ordering, or replay controls | Success only when the test route is unavailable by default, guard failures happen before mutation/disclosure, request validation is closed, duplicate fixture values are preserved in order, response data never echoes values, same-stream replacement is rejected, exhausted or wrong-kind streams fail closed, reset clears registered streams, and randomness evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.network_flow_randomness_control.v1` route response plus fixture transcript naming stream, value kind, value count, owner refs, pre/post state counts, and product result where product work executed | Product auth bypasses the token, deterministic values leak in responses/logs, a second sequence replaces a pending stream, duplicate values are deduplicated, exhaustion silently falls back to production randomness, reset leaves registered stream state, or randomness-control-only evidence closes a product row | runtime reset clears registered deterministic-randomness streams |
| TH-HARNESS-AC-053 | Sections 12, 16    | Network Flow auth-transition controls | Disabled route, token/host/origin edge cases, invalid boundary/transition/resource/hidden-response/ref bodies, exact transition consume, correlation-scoped consume, independent tuple arming, duplicate tuple conflict, reset-clears-transitions, and ownerless selector fixtures | Network Flow auth-transition route tests, schema validation, and Network Flow fixture targets that select route authorization, cursor recheck, hidden-resource, or invalidation controls | Success only when the test route is unavailable by default, guard failures happen before mutation/disclosure, request validation is closed, exact boundary/actor/incident/resource/correlation matching consumes once, mismatches leave pending state, independent tuples can coexist, duplicate tuple replacement is rejected, reset clears registered transitions, responses never disclose hidden resource details, and auth-transition evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.network_flow_auth_transition_control.v1` route response plus fixture transcript naming boundary, transition kind, actor ref, incident ref, resource kind/ref, hidden response kind, owner refs, pre/post auth state, and product result where product work executed | Product auth bypasses the token, a mismatch consumes a transition, a duplicate replaces pending state, hidden resource identifiers leak in route output/logs, reset leaves registered transition state, or auth-transition-only evidence closes a product row | runtime reset clears registered auth-transition controls |
| TH-HARNESS-AC-054 | Sections 12, 16    | Network Flow audit assertion controls | Disabled route, token/host/origin edge cases, invalid assertion/event/resource/ref/count bodies, exact-count assertion, zero-occurrence assertion, no-audit replay assertion, exact consume, correlation-scoped consume, independent tuple arming, duplicate tuple conflict, reset-clears-assertions, and ownerless selector fixtures | Network Flow audit-assertion route tests, schema validation, and Network Flow fixture targets that select transactional audit, replay, failure, graph-success, or binding-count controls | Success only when the test route is unavailable by default, guard failures happen before mutation/disclosure, request validation is closed, exact event/operation/resource/correlation matching consumes once, mismatches leave pending state, independent tuples can coexist, duplicate tuple replacement is rejected, count rules fail closed, reset clears registered assertions, responses never disclose raw audit payloads or secret material, and audit-assertion evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.network_flow_audit_assertion_control.v1` route response plus fixture transcript naming assertion kind, event code, operation ref, actor ref, incident ref, resource kind/ref, baseline count, expected final count, expected replay increment, owner refs, observed product counts, and product result where product work executed | Product auth bypasses the token, a mismatch consumes an assertion, a duplicate replaces pending state, invalid count rules arm a control, raw audit payloads or secret material leak, reset leaves registered assertion state, or audit-assertion-only evidence closes a product row | runtime reset clears registered audit assertions |
| TH-HARNESS-AC-055 | Sections 8, 16     | Network Flow structural accounting | Current Network Flow owner spec, adoption tracker, contract-family registry, generated outputs, generated-drift scratch manifest, and task surface, plus negative fixtures for missing IDs, unresolved dependency locators, premature generated symbols, missing scratch inputs, and missing public targets | Network Flow structural-accounting validator, schema validation, `make json-shape-check`, `make harness-contract`, and `make generate-drift` | Success only when fixture and acceptance IDs are contiguous and tracker-mapped, Table 1-B has exactly the expected dependency rows and no `TODO:` locator cells, planned contract status has activation dependencies with no generated Network Flow symbols, active status has generated symbols with no dependencies, scratch replay copies required inputs, and required public Make targets exist | Bounded summary naming current counts, locator status, and contract status | Bounded schema/artifact diagnostic on mismatch | `cartulary.network_flow_activity_accounting.v1` manifest validation plus retained JSON-shape, harness-contract, and generate-drift summaries | Missing owner IDs, unresolved dependency locators, hand-edited generated Network Flow symbols while planned, active registry without generated symbols, missing scratch input, or private raw target closes drift evidence | no child work beyond shape and drift checks |

| TH-HARNESS-AC-056 | Sections 4, 11, 12, 15, 16 | Test-support ownership and explicit policy | Registered shared and owner-local support roots, unknown-root fixtures, in-process server mode fixtures, fixture-policy agreement fixtures, security-profile rendering fixtures, and module control contributions | Support-inventory validator, service-backed guard tests, server-helper tests, PostgreSQL helper tests, static-analysis wrapper tests, backend boundary check, and affected product-control route tests | Success only when support roots are owner-named and registered exactly once, phase-shaped active helpers are absent, route mode and database policy are explicit, manifest and call-site fixture policy agree, every support root is security-scanned, runtime-compiled controls stay in runtime scans, and module contributions preserve guarded behavior | Bounded ownership/policy summary | Bounded configuration, boundary, or security diagnostic | Validated test-support inventory plus ordinary target summaries from the affected checks | Unknown/duplicate support root, zero-value route mode, fixture-policy mismatch, unscanned support root, phase-shaped active helper, or runtime control hidden by support exclusions passes | no cleanup beyond the selected test/service contract |

### 17.1 Requirement-to-Acceptance Traceability

| Requirement range         | Owner section                      | Acceptance criteria                                     |
| ------------------------- | ---------------------------------- | ------------------------------------------------------- |
| `TH-HARNESS-REQ-001..049` | Status, scope, authority, purpose  | TH-HARNESS-AC-013, TH-HARNESS-AC-015, TH-HARNESS-AC-016, TH-HARNESS-AC-022, TH-HARNESS-AC-026, TH-HARNESS-AC-029 |
| `TH-HARNESS-REQ-050..099` | Public command surface             | TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005, TH-HARNESS-AC-018, TH-HARNESS-AC-020, TH-HARNESS-AC-022, TH-HARNESS-AC-023, TH-HARNESS-AC-027, TH-HARNESS-AC-028, TH-HARNESS-AC-038, TH-HARNESS-AC-039, TH-HARNESS-AC-040, TH-HARNESS-AC-041, TH-HARNESS-AC-042, TH-HARNESS-AC-045, TH-HARNESS-AC-046, TH-HARNESS-AC-056 |
| `TH-HARNESS-REQ-100..149` | Configuration                      | TH-HARNESS-AC-002, TH-HARNESS-AC-003, TH-HARNESS-AC-021, TH-HARNESS-AC-022, TH-HARNESS-AC-028, TH-HARNESS-AC-029 |
| `TH-HARNESS-REQ-150..199` | Result roots and artifact identity | TH-HARNESS-AC-003, TH-HARNESS-AC-015                    |
| `TH-HARNESS-REQ-200..249` | Output modes                       | TH-HARNESS-AC-004, TH-HARNESS-AC-005, TH-HARNESS-AC-023 |
| `TH-HARNESS-REQ-250..299` | Artifacts and schemas              | TH-HARNESS-AC-000, TH-HARNESS-AC-004, TH-HARNESS-AC-015, TH-HARNESS-AC-019, TH-HARNESS-AC-022, TH-HARNESS-AC-025, TH-HARNESS-AC-028, TH-HARNESS-AC-031, TH-HARNESS-AC-048, TH-HARNESS-AC-049 |
| `TH-HARNESS-REQ-300..349` | Failure and exit codes             | TH-HARNESS-AC-013, TH-HARNESS-AC-014, TH-HARNESS-AC-032 |
| `TH-HARNESS-REQ-350..399` | Scheduler                          | TH-HARNESS-AC-006, TH-HARNESS-AC-018, TH-HARNESS-AC-021, TH-HARNESS-AC-022, TH-HARNESS-AC-024, TH-HARNESS-AC-030 |
| `TH-HARNESS-REQ-400..449` | Services                           | TH-HARNESS-AC-007, TH-HARNESS-AC-010, TH-HARNESS-AC-017, TH-HARNESS-AC-033, TH-HARNESS-AC-049, TH-HARNESS-AC-056 |
| `TH-HARNESS-REQ-450..499` | Reset route                        | TH-HARNESS-AC-008, TH-HARNESS-AC-034, TH-HARNESS-AC-035, TH-HARNESS-AC-050, TH-HARNESS-AC-051, TH-HARNESS-AC-052, TH-HARNESS-AC-053, TH-HARNESS-AC-054, TH-HARNESS-AC-056 |
| `TH-HARNESS-REQ-500..549` | Cleanup                            | TH-HARNESS-AC-009, TH-HARNESS-AC-010, TH-HARNESS-AC-028, TH-HARNESS-AC-036 |
| `TH-HARNESS-REQ-550..599` | Platform                           | TH-HARNESS-AC-012                                       |
| `TH-HARNESS-REQ-600..649` | Security and redaction             | TH-HARNESS-AC-003, TH-HARNESS-AC-011, TH-HARNESS-AC-015, TH-HARNESS-AC-036, TH-HARNESS-AC-056 |
| `TH-HARNESS-REQ-650..699` | Product integration                | TH-HARNESS-AC-013, TH-HARNESS-AC-016, TH-HARNESS-AC-026, TH-HARNESS-AC-043, TH-HARNESS-AC-044, TH-HARNESS-AC-047, TH-HARNESS-AC-049, TH-HARNESS-AC-050, TH-HARNESS-AC-051, TH-HARNESS-AC-052, TH-HARNESS-AC-053, TH-HARNESS-AC-054, TH-HARNESS-AC-055, TH-HARNESS-AC-056 |

## 18. Sources and Evidence Limits

This section is traceability and evidence posture. It does not add current conformance behavior.

Primary repository evidence used to shape this NLSpec includes:

- `testing-harness-nlspec.md`, prior draft;
- `nlspec-spec.md`, NLSpec standard;
- Core 00 through Core 04 for product-conformance authority;
- Core 05 for claim-publication separation;
- `docs/domain.md` for vocabulary and owner navigation;
- implementation and testing guides for repository command-surface and harness context;
- `Makefile`, `tools/task_surface_manifest.json`, generated task-surface includes, scheduler manifests, and schema files when present in the repository.

The following evidence categories remain non-normative in this document unless promoted by a requirement above:

| Evidence category                                                | Current role                                                                      |
| ---------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| Recovery docs under `docs/testing-harness-spec-recovery-docs/**` | Historical traceability and diagnostic context.                                   |
| Raw package scripts and tool output                              | Developer convenience or child-command evidence.                                  |
| Script paths, generated Make include names, helper binaries, `internal_helper`, `check_internal`, priority-band names, and generator-only constants | Implementation details unless a requirement above explicitly promotes one. |
| Playwright screenshots, videos, traces, HTML reports             | Diagnostic secret-bearing artifacts.                                              |
| Hosted CI provider workflows                                     | Outside current conformance unless provider source is supplied and later adopted. |
| Visual snapshot refresh process                                  | Future-only unless refresh authority is later adopted.                            |

Exact numeric constants are normative only when they protect security, cleanup safety, bounded output, or deterministic scheduling. Other numeric values in generated manifests, helper names, priority bands, and generator-only constants are implementation details unless this NLSpec gives them a requirement.

The editorial lint for TH-HARNESS-AC-016 rejects the forbidden evidence markers listed in this non-normative section when they appear in Sections 1 through 17. The forbidden markers are: `TODO`, `source_limited`, `source-limited`, `source-observed`, `current code`, `selected evidence`, `recovery evidence`, and `maintainer_decision_required`.

## 19. Future Decisions Outside Current Conformance

The items below are explicitly outside the current conformance profile. They do not block implementation of the current harness contract.

Adoption of `cartulary.testing_harness.current.v1` adopts only Sections 1 through 17 and the current conformance rules explicitly listed there. It MUST NOT adopt any Section 19 future area as current harness conformance, product conformance, visual refresh authority, provider-specific hosted CI behavior, Playwright diagnostic schema stability, or Core 05 claim-publication evidence.

| Future area                                                 | Current treatment                    | Future adoption requirement                                                                  |
| ----------------------------------------------------------- | ------------------------------------ | -------------------------------------------------------------------------------------------- |
| macOS certification                                         | Unsupported for current conformance. | Add platform profile, exact toolchain matrix, and acceptance evidence.                       |
| Windows-native support                                      | Unsupported for current conformance. | Add platform profile separate from WSL2.                                                     |
| Podman/Podman Compose                                       | Unsupported for current conformance. | Add service fixture compatibility profile and cleanup proof.                                 |
| Hosted CI annotations/uploads/artifact-retention dashboards | Provider-neutral `make ci` only.     | Add provider workflow source and provider-specific contract.                                 |
| Visual snapshot refresh authority                           | Validation-only.                     | Add exact refresh command, platform/browser bounds, review path, and golden update criteria. |
| Playwright report/trace/video/screenshot and visual-geometry diagnostic schemas | Diagnostic-only.                     | Adopt exact Playwright version/schema family or wrapper schema.                              |
| Benchmark-publication harness integration                   | Not part of harness conformance.     | Add Core 05-compatible benchmark manifest and claim-publication profile.                     |
| Scheduler, cross-run helper, or product test-result artifact reuse | Not adopted; selected work still executes in local `check`. Current same-run helper refs are limited to helper/setup artifacts produced under the same run root and do not count as scheduler reuse. | Define reusable artifact cache provenance, allowed work-unit classes, retained artifact schema, reused accounting, bypass controls, CI policy, revised TH-HARNESS-AC-018 behavior, and explicit exclusions for security, drift, services, cleanup, destructive safeguards, browser/live-state work, and runtime reset. |
