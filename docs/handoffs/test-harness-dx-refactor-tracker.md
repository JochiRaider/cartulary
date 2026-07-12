# Test Harness DX Refactor Tracker — Iteration 2

Status: complete  
Started: 2026-07-12, America/New_York  
Baseline branch/commit: `main` at `388421f6`, with intentional uncommitted
DX-008 through DX-020 remediation changes preserved  
Authority: `docs/testing-harness-nlspec.md` owns harness behavior;
`docs/domain.md` classifies the affected vocabulary as implementation support

## 1. Objective and compatibility boundary

This tracker controls DX-021 through DX-030. The primary seam is task-surface
ownership, validation, and generated Make projection. The iteration removes private
legacy compatibility that has no continuing value, separates command ownership from
execution topology, and replaces dense repeated Make output with typed shared runtime
profiles and thin bindings.

Public Make targets, `command_id` values, declared inputs and defaults, output and
artifact contracts, failure/exit behavior, cleanup, result-root layout, and
`cartulary.task_surface_manifest.v15` remain stable. Negative private-path guardrails,
the owner-required SeaweedFS migration window, current recovery commands, and current
schemas remain. Domain vocabulary is unchanged.

## 2. Baseline

- Task surface: 140 targets: 96 public, 11 check-internal, 33 helper-only.
- `tools/task_surface.generated.mk`: 931 lines, 879,944 bytes, 191 lines over
  1,000 characters, maximum line length 7,037.
- The execution topology embeds approximately 229 KiB of task-surface ownership.
- `tools/harness/generated-artifacts/task-surface.mjs` is 2,369 lines and combines
  model access, validation, queries, and Make rendering.
- Standalone task-surface report, harness-contract, generated-policy, and JSON-shape
  checks passed during planning. One combined multi-target planning invocation was
  invalid because target-local `TASK_SURFACE_REPORT_ARGS` reached later targets; it is
  not acceptance evidence.

## 3. Work tracker

Update the active row after its implementation and validation and before starting the
next row. Final statuses are `DONE`, `DROPPED`, or threshold-qualified `BLOCKED`.

| ID | Workstream | Status | Depends on | Evidence / result | Exit condition |
| -- | ---------- | ------ | ---------- | ----------------- | -------------- |
| DX-021 | Specify the new ownership and compatibility contract. | DONE | DX-020 | NLSpec Sections 1, 4/4.1, 5.3, 8, and 17; `make lint-markdown`; standalone task-surface report `check=pass`. | The owner input, topology v4, public invariants, explicit removals, typed profiles, compact input source transport, and generated Make density/growth gates are normative. |
| DX-022 | Separate task-surface ownership from execution topology. | DONE | DX-021 | `task_surface_owner.v1`; topology v4; generation root `p30725`; drift root `p30898`; JSON shape `p91145`; extended smoke `p65268`; exact 140-target owner/projection comparison. | Task metadata has one authored owner; 40 check profiles are topology-owned; generated v15 is exactly the owner projection and public behavior is unchanged. |
| DX-023 | Decompose the generated-artifact implementation. | DONE | DX-022 | 13-line facade, 132-line model/query module, 1,674-line closed validator, and 557-line Make renderer; generation root `p32286`; extended smoke root `p33289`; harness-contract pass. | Live callers use the directory facade; the 2,369-line catch-all file is deleted with no forwarding shim. |
| DX-024 | Replace untyped alias recipes. | DONE | DX-023 | 15 `artifact_binding`, six `aggregate`, one `readiness_projection`; generation root `p99676`; extended smoke root ending `p380`; harness-contract and report parity pass. | All former aliases are typed, profile invariants are fixture-tested, and `alias` is rejected. |
| DX-025 | Replace dense Make rendering. | DONE | DX-024 | Bindings 93,939 bytes; runtime 4,029 bytes; combined 97,968 bytes; maximum line 500 bytes; synthetic growth 147 bytes/target; generation `p69674`; drift `p69824`; generated policy `p69961`; JSON shape `p70103`; generate drift `p70536`; harness-contract and wrapper fixtures pass. | Shared runtime and thin bindings meet every density budget with public behavior preserved. |
| DX-026 | Remove private runtime compatibility payloads. | DONE | DX-025 | Canonical duration/failure/accounting/source fixtures pass; execution smoke, browser batch, Go target, run-phase, public-wrapper, agent-finalize, harness-contract pass; schedules `p30958`; drift `p31131`; generate drift `p31267`. | Current producers use only canonical duration, failure, accounting, and Make-input fields; retired private module fallback is removed. |
| DX-027 | Remove historical schema and baseline compatibility. | DONE | DX-026 | Four obsolete schemas deleted; obsolete explain-run fixture, stateful weight, raw-baseline shape tests pass; schedules `p33609`; JSON shape `p33864`; duration coverage `p40921`; drift `p40953`; generate drift `p41090`; harness-contract pass. | Obsolete schemas, readers, aggregate fallback, and raw legacy-key pruning are absent; current evidence remains valid. |
| DX-028 | Add future-growth and anti-drift guardrails. | DONE | DX-022–DX-027 | 25-target growth 147 bytes/target; public identity digest fixed; topology ownership/redefinition/generated-owner negatives pass; schedules `p44413`; JSON shape `p45207`; policy `p45518`; drift `p45661`; generate drift `p45798`; extended smoke pass `p33690`. | Growth, ownership, density, public identity, retired-path, and reintroduction fixtures pass. |
| DX-029 | Run final acceptance. | DONE | DX-028 | Finalizer `p22069`; final drift roots `p23430`–`p24317`; Phases 4/9/12 and service-backed slices pass; FE-P8 exact/cumulative pass; browser functional/stateful pass; check `p29326`; release `p36616`. | Narrow, generated, representative phase/frontend/browser, check, and release gates pass. |
| DX-030 | Complete tracker and handoff. | DONE | DX-029 | Sections 5.1–5.9; finalizer `p28475`; schedule drift `p29846`; generate drift `p29976`; report and whitespace audit pass. | Decisions, migrations, measurements, commands, roots, failures, and residual risks are recorded with no active row. |

## 4. Workstream notes

| Date/time | ID | Change and ownership | Validation | Compatibility / next action |
| --------- | -- | -------------------- | ---------- | --------------------------- |
| 2026-07-12 | Bootstrap | Created the controlling iteration-2 tracker without modifying or reverting the intentional prior remediation tree. | Planning inspection and standalone task-surface report. | Preserve prior work; start DX-021. |
| 2026-07-12 | DX-021 | Amended the harness owner contract for the authored task surface, topology v4, compact input sources, typed profiles, unsupported historical schemas, and Make density limits. | `make lint-markdown`; `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` (`check=pass`). | Public contracts unchanged; start owner/topology separation. |
| 2026-07-12 | DX-022 | Extracted 140 targets/recipes into schema-validated `task_surface_owner.v1`; moved 40 scheduler profiles to topology v4; added owner digesting and scratch-repo coverage; regenerated projections. | `make json-shape-check` `p91145`; `make run-harness-smoke-extended` `p65268`; `make generate-drift` `p29545`; `make phase-schedules phase-schedule-drift` `p30725`/`p30898`; owner/projection deep equality passed. | The first extended runs exposed expected v3 fixtures and one missing scratch owner copy; both were corrected. Start generator decomposition. |
| 2026-07-12 | DX-023 | Split task-surface model/query, validation, Make rendering, and facade modules; migrated every live import and deleted the old catch-all without a shim. | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`; `make phase-schedules` `p32286`; `make harness-contract`; `make run-harness-smoke-extended` `p33289`. | Public report remains `check=pass`; start typed recipe profiles. |
| 2026-07-12 | DX-024 | Reclassified 22 unvalidated aliases into 15 artifact bindings, six aggregates, and one internal readiness projection; added closed validators and negative legacy/type fixtures. | `make harness-contract`; `make phase-schedules` `p99676`; `make run-harness-smoke-extended` root ending `p380`; standalone task-surface report `check=pass`. | A one-generation transitional validator allowance was removed immediately after owner-derived outputs migrated; no alias remains. Start compact Make rendering. |
| 2026-07-12 | DX-025 | Generated a 4 KiB invariant runtime include and 94 KiB reviewable target bindings; replaced repeated input/origin payloads with the closed compact source vector; bounded physical lines and added generation-time density plus 25-target growth validation. | Report metrics: 93,939/4,029/97,968 bytes, 500-byte maximum line, 147 bytes/target growth; `phase-schedules` `p69674`; drift `p69824`; policy `p69961`; JSON shape `p70103`; generate drift `p70536`; task-surface, execution-topology, Make-sequence, public-wrapper, and harness-contract fixtures pass. | The 879,944-byte baseline fell by 88.9%. Two fixture runs initially failed because they asserted the retired inline preflight and five-output renderer; the fixtures were migrated and passed. Public help, wrapper, prerequisite, sequence, and artifact behavior is unchanged. Start private runtime compatibility removal. |
| 2026-07-12 | DX-026 | Migrated every phase producer to explicit logical/executed durations; removed generic-duration fallback; closed failure ordering to three canonical fields; made duplicated frontend accounting extensions fail; propagated compact Make sources through node tools and finalization; removed the named success-log compatibility test and stale private output-module fallback. | Run-phase, Go-target, browser-batch, public-wrapper, agent-finalize, harness-contract, and execution-smoke fixtures pass; schedules `p30958`; drift `p31131`; generate drift `p31267`. | One Go fixture exposed the stale `tools/harness/core/test-output.mjs` private fallback, which was corrected to the owned output module. One drift invocation correctly failed after that input changed (`p30768`), then passed after regeneration. Public contracts remain stable; start historical schema/baseline removal. |
| 2026-07-12 | DX-027 | Deleted phase-plan v1 and pressure-summary v1–v3 schemas; made explain-run label old retained JSON `unsupported_schema` with its path; removed aggregate stateful-weight division and its unpartitioned baseline key; removed raw aggregate legacy-key pruning and closed raw keys to `target::aggregate::package`. | Schedules `p33609`; JSON shape `p33864`; execution-topology, run-phase/explain-run, and Go-duration fixtures pass; duration coverage `p40921`; drift `p40953`; generate drift `p41090`; harness-contract pass. | Historical JSON remains raw-readable but cannot validate as current evidence. Exact partition weights remain; missing partitions use the declared default. Current schemas and public commands are unchanged. Start growth and anti-drift guardrails. |
| 2026-07-12 | DX-028 | Added generation-time 25-target growth enforcement, public target/command identity digest, explicit topology unknown-target and metadata-redefinition negatives, generated-owner rejection, retired schema/path/transport checks, and owner-only future-target coverage. | Schedules `p44413`; JSON shape `p45207`; policy `p45518`; drift `p45661`; generate drift `p45798`; harness-contract/report pass; extended smoke success `p33690` (120.384 s). | Extended attempts `p48504`, `p64277`, and `p87812` exposed scratch runtime-copy, single-line `.PHONY`, and old node-tool-macro fixture assumptions. Each was corrected; no budget was relaxed. Start final acceptance. |
| 2026-07-12 | DX-029 | Ran finalizer without retained evidence, all structural/drift gates, execution/lifecycle/extended harness, base and service-backed Phases 4/9/12, exact and cumulative FE-P8, functional/stateful browser targets, full check, and release-check. | Finalizer `p22069`; policy `p23430`; JSON shape `p23566`; schedule drift `p23876`; ledger drift `p24006`; generate drift `p24317`; base phase roots `p48911`/`p68885`/`p86770`; service roots `p7711`/`p27112`/`p44205`; FE roots `p57650`/`p59046`; stateful `p10895`; check `p29326`; release `p36616`. | Initial Phase 4 root `p30716` exposed compact source metadata leaking into nested children; the child boundary now consumes it and a regression test passes. Final check: 259/259, 1,074 tests. Release: 14/14, 1,131 tests. Final memory: 16,400 MiB available; swap 5,926 MiB in use (no start-of-pass sample, so no growth claim). Start handoff. |
| 2026-07-12 | DX-030 | Completed this final ownership, compatibility, migration, evidence, failure, and residual-risk handoff without changing domain vocabulary. | Tracker audit: DX-021–DX-030 are all `DONE`; finalizer `p28475`; schedule drift `p29846`; generate drift `p29976`; task report, Markdown lint, and `git diff --check` pass. | Iteration 2 is complete. The recommended next seam is artifact-reference/path typing in retained-run diagnostics, followed by further validator-module decomposition if its change rate warrants it. |

## 5. Final handoff

### 5.1 Outcome

Iteration 2 is complete. The public task surface remains 140 targets: 96 public,
11 check-internal, and 33 helper-only. The ordered public target and `command_id`
inventory has an explicit compatibility digest and did not change. Domain vocabulary
is unchanged; targets, phases, fixtures, schedulers, and generated Make remain
implementation-support concepts.

The current tracked worktree diff, including intentional prior remediation, removes
16,027 lines and adds 2,673 lines before counting new owner, schema, modular
implementation, runtime Make, and tracker files. The large deletion is principally
embedded topology ownership, the 2,369-line catch-all module, obsolete schemas, and
duplicated generated Make text.

### 5.2 Final ownership and generation flow

1. `docs/testing-harness-nlspec.md` owns public harness behavior and compatibility.
2. `tools/task_surface_owner.json` owns targets, help, input contracts, sequences,
   harness tiers/checks, and typed Make recipe profiles under
   `cartulary.task_surface_owner.v1`.
3. `tools/execution_topology_manifest.json` v4 owns execution dependencies, runtime
   binaries, check profiles, browser/service schedules, and generated output paths.
   It references, but cannot redefine, task-surface targets.
4. The generated-artifact facade joins those inputs and emits the v15 task-surface
   projection, scheduler and browser manifests, thin target bindings, the shared Make
   runtime, and the digest-indexed render record.
5. Callers use the task-surface model/query facade. Validation and Make rendering are
   private cohesive modules; the old catch-all module and forwarding paths are absent.

Generated outputs remain downstream only. `tools/task_surface_manifest.json`,
`tools/task_surface.generated.mk`, `tools/task_surface.runtime.generated.mk`,
`tools/scheduler_manifest.json`, the browser batch manifest, and the render index may
not be selected as authored owner inputs.

### 5.3 Compatibility decisions

Retained public behavior:

- all public Make target names and stable `command_id` values;
- declared input/default/output/artifact/failure/exit/cleanup contracts;
- result-root layout and `cartulary.task_surface_manifest.v15`;
- current SeaweedFS migration/recovery support and retired-alias negatives;
- all current schemas, including current schemas whose identifiers end in `.v1`.

Removed private/historical behavior:

- embedded task-surface ownership in execution topology;
- untyped `alias` recipes and the old task-surface catch-all module;
- per-variable Make-origin transport and propagation beyond the public wrapper;
- generic phase-duration fallback and failure-order field aliases;
- duplicate frontend row accounting in target-summary extensions;
- the named legacy success-log compatibility behavior;
- phase-plan v1 and pressure-summary v1–v3 validation attachments;
- unpartitioned stateful weight division and raw Go aggregate legacy-key pruning;
- moved-module fallback paths.

Old local JSON with removed schema IDs remains manually readable. `explain-run`
labels it `unsupported_schema` with its artifact path and does not treat it as current
evidence.

### 5.4 Generated Make measurements

| Measure | Before | After | Gate |
| ------- | ------ | ----- | ---- |
| Binding file bytes | 879,944 | 93,934 | at most 180 KiB |
| Shared runtime bytes | none | 4,120 | at most 64 KiB |
| Combined bytes | 879,944 | 98,054 | at most 220 KiB |
| Binding/runtime lines | 931/none | 1,080/82 | informational |
| Maximum physical line | 7,037 bytes | 500 bytes | at most 512 bytes |
| Lines over 1,000 bytes | 191 | 0 | implied by line gate |
| Twenty-five-target average growth | unmeasured | 147 bytes/target | at most 512 bytes |
| Per-variable origin expansions | repeated per target | 0 | exactly 0 |

Combined generated Make size fell 88.9 percent. The line count increased because
`.PHONY`, help, and recipes now use bounded reviewable physical lines; density is
measured in bytes and maximum line length, not compressed line count.

### 5.5 Migration guide

To add a target:

1. Add or revise the NLSpec public registry only when the public contract changes.
2. Add one owner target row and one closed typed recipe in
   `tools/task_surface_owner.json`.
3. Add a topology check profile only when the target participates in a scheduler.
4. Run `make phase-schedules`, the task-surface report, JSON shape, harness contract,
   and generation drift checks. Do not add renderer branching for an ordinary target.

To add a public input, revise the NLSpec input matrix and the owner input contract
together. The compact source vector is generated from the closed inventory; no new
Make-origin variable or wrapper parser branch is permitted.

Use `artifact_binding`, `aggregate`, or `readiness_projection` only when their closed
invariants fit. A genuinely new side-effect shape requires an explicit NLSpec and
validator/runtime profile change, not reuse of a generic alias. Topology additions may
reference owner targets and execution metadata only; command IDs, help, inputs, output
policy, and lifecycle remain owner fields.

### 5.6 Final acceptance evidence

| Scope | Result and retained root |
| ----- | ------------------------ |
| Finalizer, `RESULTS_DIR` intentionally unset | pass, `.cartulary/test-results/20260712T182009Z-p28475` |
| Generated policy / JSON shape | pass, `p23430` / `p23566` |
| Schedule / ledger / generation drift | pass, `p29846` / `p24006` / `p29976` |
| Harness contract | pass, `.cartulary/test-results/20260712T181624Z-p25246` |
| Execution / lifecycle / extended harness | pass, `p3587` / `p28539` / `p33690` |
| Base Phases 4 / 9 / 12 | pass, `p48911` / `p68885` / `p86770` |
| Service-backed Phases 4 / 9 / 12 | pass, `p7711` / `p27112` / `p44205` |
| Exact / cumulative FE-P8 | pass, `p57650` / `p59046` |
| Browser functional / stateful | pass, `p3566` / `p10895` |
| Full `make check` | 259/259 units, 1,074 tests, root `p29326` |
| Full `make release-check` | 14/14 units, 1,131 tests, root `p36616` |

All final summaries report zero failed, missing, and unmapped evidence. Final memory
observation was 16,400 MiB available, above the 4 GiB gate. Swap had 5,926 MiB in use;
because no start-of-pass sample was retained, this handoff makes no unsupported claim
about swap growth.

### 5.7 Failed, stale, and retried evidence

- DX-025 fixtures initially asserted inline preflight text and five generated outputs;
  both were migrated to the shared runtime/six-output design.
- DX-026 drift root `p30768` correctly rejected a changed generator input; regeneration
  produced passing roots `p30958` and `p31131`.
- DX-028 extended roots `p48504`, `p64277`, and `p87812` exposed a missing scratch
  runtime include, a single-line `.PHONY` assumption, and old node-tool macro text.
  The uncontaminated rerun `p33690` passed without relaxing gates.
- DX-029 Phase 4 root `p30716` exposed parent Make-source metadata in nested scheduler
  children. The public wrapper now consumes that transport; regression coverage and
  rerun `p48911` pass.

No failed, stale, or retry-contaminated root is used as successful acceptance evidence.
No timing or duration baseline was refreshed to conceal a regression.

### 5.8 Residual risks and recommended next seam

The generated bindings still contain 820 ordinary Make variable expansions. They are
target-specific and within all density/growth gates, so replacing them with a universal
dispatcher would reduce reviewability and is not recommended.

During investigation of the failed Phase 4 sample, `explain-run DETAIL=logs` attempted
to read a directory-valued log artifact and returned `EISDIR`. This did not affect test
execution or final evidence, but it is the clearest next seam: close artifact-reference
kind/path semantics and make diagnostics reject or enumerate directory artifacts
explicitly. After that, the 1,674-line task-surface validator is the next candidate for
domain-oriented decomposition if its maintenance/change rate justifies another pass.

### 5.9 Final state

DX-021 through DX-030 are `DONE`. There are no `TODO`, `IN_PROGRESS`, `DEFERRED`,
`BLOCKED`, or dropped workstreams in this iteration. Intentional remediation changes
that predated this iteration remain preserved; this work did not revert unrelated user
changes.
