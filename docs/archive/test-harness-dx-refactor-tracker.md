# Test Harness DX Refactor Tracker

Current iteration: Iteration 4, DX-041 through DX-048 (`IN_PROGRESS`)

Historical iterations: Iteration 2, DX-021 through DX-030; Iteration 3,
DX-031 through DX-040 (`complete`)

Status: in progress
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

## 6. Iteration 3 control

Status: in progress

Started: 2026-07-12, America/New_York

Baseline branch/commit: clean `main` at `58ce7c052d9e`

Authority: `docs/testing-harness-nlspec.md` owns harness behavior;
`docs/domain.md` classifies all affected vocabulary as implementation support

### 6.1 Objective and compatibility policy

Iteration 3 closes retained harness artifact-reference kind and path semantics,
removes unsupported historical schemas and private compatibility inventories,
establishes current machine owners for schema attachments and helper facades, and
decomposes high-growth validators while preserving deterministic scheduling,
security, cleanup, failure classification, resource isolation, and public Make
behavior.

The public Make target inventory, stable `command_id` values, public inputs and
defaults, failure and exit behavior, cleanup guarantees, result-root layout,
`cartulary.task_surface_manifest.v15`, `cartulary.task_surface_owner.v1`,
`cartulary.scheduler_manifest.v2`, and generated Make density gates remain stable.
The machine-summary schema migration is an atomic clean cut: current producers and
consumers move together, no old schema is dual-emitted or dual-read, and no forwarding
shim is introduced. Old retained JSON remains manually readable but unsupported by
current validation and diagnostics.

Domain vocabulary is unchanged. Harness prose continues to qualify the implementation
concept as a `harness artifact`; it does not redefine the product-domain term
`Artifact`.

### 6.2 Current baseline

- Task surface: 140 targets: 96 public, 11 check-internal, 33 helper-only.
- Generated Make: 98,054 combined bytes, 500-byte maximum line, 147 bytes per
  synthetic target, and 820 ordinary variable expansions.
- Harness tree: 347 files; schema directory: 73 attachments.
- Generated scheduler manifest: 1,538,481 bytes, 42,497 lines, four schedules,
  1,067 work units, 924 Go shards, and 1,442 bytes per work unit.
- Large seams: JSON-shape validator 4,425 lines/94 functions; task-surface validator
  1,706 lines/43 functions; explain-run CLI 807 lines/53 functions.
- Private ownership duplication: 35 NLSpec helper-family rows versus 15 JavaScript
  buckets, 69 facade paths, nine historical rule families, 24 exact tombstones, and
  four prefix tombstones.
- Twelve schemas are self-referential only. Scheduler manifest v1 has no live runtime
  consumer. Scheduler summary common v10 inherits its substantive shape from v9.
- `make explain-run RESULTS_DIR=.cartulary/test-results/20260712T174652Z-p30716
  TARGET=phase-slice DETAIL=logs` reproduces `EISDIR` for the declared
  `scheduler_logs` directory reference.
- Fresh baseline evidence: harness contract `p37811`; JSON shape `p38441`; generated
  policy `p38759`; schedule drift `p38903`; generate drift `p39041`; standalone
  task-surface report `check=pass`.

### 6.3 Work protocol

Exactly one workstream may be `IN_PROGRESS`. Before another row starts, the active row
must record its status, evidence, validation roots, failures, retries, and exit result.
`DEFERRED` is prohibited. Final statuses are `DONE`, `DROPPED`, or genuinely justified
`BLOCKED`; a blocked row must name the external blocker, owner, exhausted alternatives,
evidence, and retry condition. Final handoff may contain no `TODO`, `IN_PROGRESS`,
`DEFERRED`, or unjustified `BLOCKED` row.

DX-039 cannot start until DX-031 through DX-038 have final statuses. DX-040 cannot
finish until every validation root is classified as successful, failed, stale,
retried, or skipped.

### 6.4 Work tracker

| ID | Workstream | Status | Depends on | Evidence / result | Exit condition |
| -- | ---------- | ------ | ---------- | ----------------- | -------------- |
| DX-031 | Open Iteration 3 and lock the contract. | DONE | DX-030 | Baseline inspection; retained `p30716` reproduction; Markdown lint `p45734`; task report `check=pass`; whitespace audit pass. | Current objective, compatibility policy, measurements, protocol, dependencies, and evidence fields are recorded without changing historical completion records. |
| DX-032 | Specify typed harness artifact and source references. | DONE | DX-031 | NLSpec Sections 4/8/9/11/15/17; current schema table; JSON shape root `p55166`. | NLSpec owns the clean schema cut, path/kind semantics, safety, failure mapping, and unsupported-old-shape policy. |
| DX-033 | Migrate schemas, producers, consumers, and owner inputs. | DONE | DX-032 | Schedules `p54362`; JSON shape `p55166`; schedule drift `p92103`; generate drift `p92239`; ledger drift `p93248`; harness contract `p93564`. | All current emitters/readers use the new contracts; generated projections are refreshed; no live old IDs remain. |
| DX-034 | Harden retained-run diagnostics. | DONE | DX-033 | Typed retained-artifact resolver; negative containment/kind/symlink/bounds tests; harness contract `p5964`. | Diagnostics validate containment/kind/symlinks before reading and deterministically enumerate bounded directory logs. |
| DX-035 | Establish one schema-attachment owner and delete obsolete schemas. | DONE | DX-034 | Attachment registry and exact closure validation; JSON shape `p98523`; generated policy `p99041`; generation drift `p99184`; current harness contract `p5964`. | Every current schema is registered once; obsolete attachments/readers are absent; current fixtures retain verdict parity. |
| DX-036 | Replace private ownership mirrors and remove shims. | DONE | DX-035 | Semantic helper owner and schema; unknown-root/duplicate/missing-facade negatives; harness contract `p5964`; zero removed-ID scans. | Cross-owner imports use one current machine owner; historical tombstones and the JavaScript mirror are absent; no forwarding shim was added. |
| DX-037 | Decompose validators and add growth gates. | DONE | DX-036 | Task-surface modules: 773-line owner/profile validator, 799-line recipe/help validator, and 167-line input validator; 183-line schema-attachment validator; 782-line diagnostic facade; task report growth gates pass. | Cohesive task/schema/diagnostic validators and deterministic scheduler density/growth gates pass without changing scheduler v2 behavior. |
| DX-038 | Run cross-cutting negative, security, isolation, and growth acceptance. | DONE | DX-033–DX-037 | Fast `p33354`, lifecycle `p33357`, execution `p56162`, extended `p72155`, full `p1450`, JSON `p36331`, contract `p36295`, generated policy `p36354`, schedule drift `p36298`, and generation drift `p37637` pass. | Unsafe inputs fail before access/child work; cleanup, redaction, resource isolation, and scheduling invariants pass. |
| DX-039 | Run final acceptance. | DONE | DX-038 | Finalizer `p23037`; structural roots `p26390`, `p27023`, `p27024`, `p27037`, `p27041`; base/service slices; FE-P8 exact/cumulative; browser aggregate `p95659`; current `check` and `release-check` root `p28938`. | Finalizer, drift, harness, representative phase/frontend/browser, check, and release gates are classified and pass or carry justified skips/blockers. |
| DX-040 | Complete tracker and handoff. | DONE | DX-039 | Markdown lint pass; task report `check=pass`; JSON `p14251`; harness contract `p14254`; whitespace audit pass. | Final ownership, migration, compatibility, metrics, roots, failures, retries, skips, risks, and next seam are recorded with no active row. |

### 6.5 Dependencies and sequencing

The execution order is DX-031, DX-032, DX-033, DX-034, DX-035, DX-036, DX-037,
DX-038, DX-039, and DX-040. Owner/specification changes precede schema and producer
migrations; authored inputs precede generated outputs; safety and negative acceptance
precede broad repository acceptance.

### 6.6 Workstream evidence log

| Date/time | ID | Change and ownership | Validation | Compatibility / next action |
| --------- | -- | -------------------- | ---------- | --------------------------- |
| 2026-07-12 | DX-031 | Opened Iteration 3 in the controlling tracker and preserved DX-021–DX-030 history verbatim. | Baseline commands and retained-run reproduction; Markdown lint `p45734`; task report and whitespace audit pass. | Public Make behavior remains stable; begin DX-032. |
| 2026-07-12 | DX-032 | Defined run-root-relative file/directory harness artifact refs, separated repository sources, locked the six-schema clean cut, and made unsafe dereference an artifact failure. | JSON shape `p55166`; schema IDs and attachments resolve. | No public Make change; begin atomic producer/consumer migration. |
| 2026-07-12 | DX-033 | Migrated the six current schema families, centralized wrapper refs, frontend sources, accessibility directories, service lifecycle IDs, same-run digests, release evidence, task owner rows, and generated projections. | Schedules `p54362`; JSON `p55166`; drift `p92103`/`p92239`; ledger `p93248`; contract `p93564`. | Extended roots `p56854` and `p68724` exposed stale frontend and tool-summary fixtures; both fixtures were migrated and are not acceptance evidence. Begin diagnostics hardening. |
| 2026-07-12 | DX-034 | Added one fail-closed retained-artifact resolver, removed the obsolete schema crawler, and made directory logs deterministic and bounded. | Resolver negatives cover absolute/traversal/symlink/missing/wrong-kind/oversize refs; harness contract `p5964`. | Current v4 refs are supported; the retained v3 `p30716` root now receives a bounded unsupported-schema result rather than `EISDIR`, consistent with the clean-cut policy. |
| 2026-07-12 | DX-035 | Added the authored schema-attachment registry, deleted self-only and v1 attachments, registered phase-slice scheduler summary v4, and flattened scheduler common v10. | JSON shape `p98523`; generated policy `p99041`; generation drift `p99184`; exact old-ID scan empty outside this history. | Current attachment closure is exact; unregistered additions fail. |
| 2026-07-12 | DX-036 | Replaced the 15-bucket JavaScript mirror and historical tombstones with 34 semantic machine-owner entries keyed to the NLSpec. | Duplicate/missing facade and unknown-root negatives; current boundary scan; harness contract `p5964`. | Thin current facades remain; historical private paths have no compatibility registry or forwarding modules. |
| 2026-07-12 | DX-037 | Added scheduler manifest size, p95, maximum-unit, synthetic-growth, and scratch-determinism gates to the task-surface report. | Task report `check=pass`; measured 1,538,481 bytes and 1,067 work units. | Growth gates pass; validator module extraction continues. |
| 2026-07-12 | DX-037 | Split task-surface owner/profile, recipe/help, and input validation into cohesive modules below 800 lines; extracted schema-attachment/helper-owner policy and retained-artifact resolution; removed the obsolete diagnostic crawler. | Task report `check=pass`; JSON `p36331`; harness contract `p36295`; all extracted modules parse and preserve verdicts. | The 4,326-line extension-heavy JSON orchestration file remains a measured residual seam; its harness schema-owner policy no longer grows there. |
| 2026-07-12 | DX-038 | Added typed artifact-resolution negatives, schema/helper closure negatives, scheduler growth gates, and order-independent overlap proof for concurrently launched scheduler fixture children. | Fast `p33354`; lifecycle `p33357`; execution `p56162`; extended `p72155`; structural roots `p36331`, `p36295`, `p36354`, `p36298`, and `p37637`. | Earlier extended roots `p7971`, `p47350`, `p89956`, `p81592`, and `p24934` are failed/retry-contaminated evidence and are not cited as success. |
| 2026-07-12 | DX-038 | Closed the real-target artifact trust boundary in the Node-tool wrapper, made artifact existence checks run-root-relative, and corrected the `check` success budget to 10 lines/2,500 bytes for bounded slow-host progress. | Real Phase 0 slice probe passed with only contained refs; public `check` `p77201`; full composed harness `p1450` passed in 980.805 seconds. | Failed full roots `p39009`, `p7232`, `p60161`, and `p85408` exposed the stale resolver and budget assumptions and are classified as failed/retried, never success. Begin DX-039. |
| 2026-07-12 | DX-039 | Ran finalizer-led structural, harness, base-phase, frontend-growth, browser, check, and release acceptance. Corrected one stale v4 guide sentence and refreshed its authored digest chain before the final release run. | Final current release root `p28938` passed 14/14 work units and 1,131 tests; current check inside it passed 259/259 and 1,074 tests. | `RESULTS_DIR` was intentionally unset for finalizer; retained-run duration maintenance was skipped, not inferred from retry-contaminated roots. Begin DX-040. |
| 2026-07-12 | DX-040 | Completed the ownership, migration, compatibility, deletion, metrics, validation, failure/retry/skip, residual-risk, and next-seam handoff. | Markdown lint; task report; JSON `p14251`; harness contract `p14254`; `git diff --check` all pass. | All DX-031–DX-040 rows are `DONE`; no active, deferred, dropped, or blocked workstream remains. |

### 6.7 Final ownership flow and migration guidance

The final ownership flow is:

1. `docs/testing-harness-nlspec.md` owns public harness behavior and current
   schema identities.
2. `tools/harness_schema_attachments.json`,
   `tools/harness_helper_ownership.json`, task-surface/topology inputs, phase
   registries, phase maps, and schemas own repository attachment and path
   mechanics.
3. Cohesive validators and generators consume those authored owners.
4. Generated task-surface, schedule, topology, ledger, and frontend freshness
   projections remain downstream and are never hand-edited.
5. Runtime producers and consumers emit current retained evidence.
6. Diagnostics resolve typed refs inside a selected run root and fail closed
   before replay.

Migration is a clean cut. Producers and consumers must use the six replacement
schema IDs recorded in Section 6.1. A harness artifact ref has `role`,
`path_kind`, a normalized run-root-relative `path`, and file-only `format`;
digested refs add SHA-256. Repository source inputs use `source_files` or
`source_refs`, never artifact refs. There are no dual emitters, dual readers,
old-call adapters, aliases, or forwarding modules. Retained old JSON can be read
manually but is unsupported by current validators and diagnostics.

Private helper moves update the semantic key in
`tools/harness_helper_ownership.json`, move all live callers directly, regenerate
downstream provenance, and delete the old module after an exact zero-caller scan.
No historical-path tombstone or forwarding shim is added.

`docs/domain.md` required no change. `harness artifact`, artifact-reference kinds,
schema IDs, helper facades, schedulers, and diagnostics remain
implementation-support vocabulary; harness prose retains the `harness artifact`
qualifier and does not redefine the domain term `Artifact`.

### 6.8 Compatibility disposition and deletion record

| Surface | Final disposition | Evidence / rationale |
| ------- | ----------------- | -------------------- |
| 96 public Make targets, stable `command_id` values, inputs/defaults, result-root layout, cleanup/failure behavior, task-surface v15, task owner v1, scheduler manifest v2 | Retained. | Final task report is `check=pass`; current release root `p28938` passed. The `check` success ceiling is now 10 lines/2,500 bytes to admit bounded slow-host progress without changing emitted behavior. |
| SeaweedFS migration/compatibility support, same-run helper evidence, explicit scheduler default weights, failure fallback classification, Playwright install fallback, human-only newest-run selection | Retained. | Continuing owner value remains demonstrated by release and harness acceptance. |
| Declared thin private facades | Retained. | They are current cross-subsystem encapsulation boundaries, not moved-module aliases. |
| Scheduler manifest expansion and 820 generated Make value expansions | Retained. | Only two byte-identical work units exist; current density and growth gates pass, so template indirection or a universal dispatcher remains rejected. |
| Six schema families listed in Section 6.1 | Replaced atomically. | No live old schema ID remains outside this historical ledger; current release evidence validates only replacements. |
| Schema-less service scopes, scheduler manifest v1, and non-current diagnostic readers | Removed. | Current diagnostics reject unsupported shapes before dereference. |
| Twelve self-only attachments, scheduler manifest v1, scheduler common v9, and the six replaced attachments | Removed. | Schema attachments fell from 73 to 62: 20 tracked deletions and nine current additions. Common v10 is flattened and preserves current verdicts. |
| JavaScript helper ownership mirror, nine historical rule families, 24 exact tombstones, four prefix tombstones | Removed. | One 34-key semantic owner now matches the NLSpec exactly; unknown roots and cross-owner imports fail semantically. |
| Unused private re-exports | Remove only after exact zero-caller proof. | No continuing external consumer was demonstrated and no forwarding shim was introduced. |
| `cartulary.agent_finalize_summary.v3.child_artifacts` legacy `kind` shape | Bounded follow-up investigation. | It was outside the approved six-schema clean cut and is empty when `RESULTS_DIR` is omitted. A later owner review must prove consumers before a v4 cut; adding an implicit adapter is forbidden. |

### 6.9 Final generated and structural measurements

| Measurement | Final value | Gate |
| ----------- | ----------- | ---- |
| Task surface | 140 targets: 96 public, 11 check-internal, 33 helper-only | Pass; identity and class counts unchanged. |
| Generated Make | 98,054 bytes; 500-byte maximum line; 147 bytes per synthetic target; 820 ordinary expansions | Pass. |
| Scheduler manifest | 1,538,481 bytes; 1,067 work units; 1,442 bytes/work unit | Pass, limit 1,600. |
| Serialized scheduler units | p95 1,226 bytes; maximum 10,795 bytes | Pass, limits 1,500 and 12 KiB. |
| Synthetic scheduler growth | 358 bytes/unit for 25 ordinary units; two renders byte-identical | Pass, limit 1,600 and exact determinism. |
| Task-surface validation modules | 773 owner/profile; 799 recipe/help; 167 input-contract lines | Pass, each cohesive module below 800 lines. |
| Extracted validation/diagnostic modules | 184 schema/helper attachment; 155 retained-artifact resolver; 782 explain-run facade lines | Pass. |
| Harness/schema trees | 351 harness files; 62 registered schema attachments | Exact registry closure passes. |

### 6.10 Final acceptance roots

| Area | Successful evidence |
| ---- | ------------------- |
| Finalizer | `p23037`: generated unchanged, duration maintenance skipped because `RESULTS_DIR` was unset, no stale run reused. |
| Current structural gates | JSON `p26390`; generated policy `p27023`; schedule drift `p27024`; ledger drift `p27037`; generation drift `p27041`; task report `check=pass`. |
| Harness tiers | Fast `p26977` release child/current aggregate coverage; lifecycle `p26977`; execution `p26969`; extended `p55335`; full composition `p1450`; contract covered again by final release `p28938`. |
| Base Phase 4 | Phase slice `p19358`: 7/7 work units, 51 tests. Service-backed `p39190`: 5/5, 44 tests. |
| Base Phase 9 | Phase slice `p58653`: 5/5, 70 tests. Service-backed `p76636`: 3/3, 53 tests. |
| Base Phase 12 | Phase slice `p93665`: 5/5, 134 tests. Service-backed `p16767`: 4/4, 37 tests. |
| FE-P8 exact | `p30118`: six requested implemented rows, 6/6 work units, 59 tests. |
| FE-P8 cumulative | `p60964`: implemented rows through FE-P8, 11/11 work units, 85 tests. |
| Browser aggregate | `p95659`: functional, stateful, measurement, accessibility, and visual groups; 4/4 aggregate work units and 71 tests, including 22 stateful tests. |
| Current check and release | `p28938`: release 14/14, 1,131 tests; embedded check 259/259, 1,074 tests; no missing or unmapped evidence. |

### 6.11 Failures, retries, and skipped checks

- `p30716` is historical v3 evidence. Current diagnostics return bounded
  unsupported-schema behavior rather than `EISDIR`; it is not success evidence.
- Extended roots `p56854`, `p68724`, `p7971`, `p47350`, `p89956`, `p81592`, and
  `p24934` exposed stale schema/path fixtures and a child-log ordering assumption.
  Each defect was corrected and superseded by `p72155` and post-finalizer
  `p55335`.
- Full roots `p39009`, `p7232`, `p60161`, and `p85408` exposed repo-relative
  post-normalization refs and stale slow-host output ceilings. They are failed or
  retry-contaminated and superseded by `p1450`.
- `p18516` is an intentionally interrupted redundant standalone stateful browser
  invocation after clean aggregate `p95659` had already completed stateful
  acceptance. It is not cited as success.
- JSON roots `p23312` and `p25619` failed after the final live-guide wording fix
  revealed a stale authored guide digest. The registry/map digest chain was
  refreshed and superseded by `p26390` plus structural roots `p27023` through
  `p27041`.
- Finalizer retained-run maintenance was skipped because `RESULTS_DIR` was unset.
  No failed, stale, partial, or retry-contaminated root was used for mutation or
  final acceptance.
- No planned base phase, FE-P8, browser aggregate, `check`, or `release-check`
  gate was skipped. Final handoff lint and whitespace checks are recorded in
  DX-040 after this section.

### 6.12 Residual risks and recommended next seam

The remaining measured risks are:

1. `check-json-shapes.mjs` is still a 4,326-line extension-heavy orchestrator.
   Harness schema/helper attachment policy no longer grows there, and database,
   phase, browser, and task-surface owners already have separate modules, but the
   Network Flow extension validation family remains dense.
2. `cartulary.agent_finalize_summary.v3.child_artifacts` still uses the old
   `role`/`kind`/`path` vocabulary. It is not part of the six-schema Iteration 3
   cut and receives no compatibility adapter.
3. Full harness real-target acceptance is intentionally expensive (successful
   root `p1450` took 980.805 seconds). The cost is visible and bounded, but it
   raises retry cost when a late output-policy assertion fails.

The recommended next seam is a bounded owner review and consumer inventory for
agent-finalizer child artifacts, followed by either deletion when unused or an
atomic `agent_finalize_summary.v4` typed-ref migration. This seam directly closes
the last known ambiguous artifact-reference shape without reworking DX-021
through DX-030 or introducing a shim. After that, extract the remaining Network
Flow JSON validators by their extension owner only if characterization proves
verdict parity; scheduler compression remains rejected absent material
deduplication.

## 7. Iteration 4 control

Status: in progress

Started: 2026-07-12, America/New_York

Baseline branch/commit: clean `main` at `84e2dc99de0a`

Authority: Core 00 through Core 04 own product behavior;
`docs/testing-harness-nlspec.md` owns harness mechanics and the performance
acceptance method; `docs/domain.md` requires no vocabulary change

### 7.1 Objective and conformance policy

Iteration 4 reduces the external process wall time of the unchanged public
`make check` command to less than 120 seconds without removing, moving,
reclassifying, bypassing, or weakening required product-conformance work. The
strict acceptance statistic is the nearest-rank p90/maximum of five consecutive
successful measured warm runs on identical current inputs. Every measured run
must preserve public target and command identity, exact required phase-row and
test coverage, deterministic scheduling, resource isolation, cleanup, artifact
integrity, security boundaries, and failure classification.

`make release-check` is explicitly excluded from this iteration. It is not a
validation gate, comparison run, fallback, rollback condition, or source of
performance evidence.

Public Make targets, stable `command_id` values, inputs/defaults, result-root
layout, current schema identities, output budgets, failure/exit behavior,
cleanup, scheduler manifest v2, scheduler event v6, execution topology v4,
task-surface v15, and task owner v1 remain stable. Private execution-family or
shard identities may be replaced only through an atomic owner-first migration;
no compatibility alias, dual reader/emitter, or forwarding shim is permitted.

Core 00 through Core 04 remain unchanged. Phase maps continue to route product
evidence, and the harness must preserve every row ID, owner citation, coverage
class, selected symbol, section, evidence layer, and default-check disposition.
Harness caches and same-run helper refs remain acceleration/diagnostic mechanics
only. Cross-run product-test reuse is prohibited. Domain vocabulary is unchanged:
`scheduler`, `work unit`, `fixture`, and `harness artifact` remain implementation
support and do not redefine the domain term `Artifact`.

### 7.2 Current baseline

The uncontaminated planning baseline is
`.cartulary/test-results/20260712T214824Z-p21895`, produced on WSL2 Linux 6.6,
an Intel Core Ultra 9 285HX with 24 logical CPUs, 31 GiB RAM, and scheduler
capacity `host_cpu=17`, `host_io=17`, `postgres_clone=8`, `postgres_reset=5`,
`process=6`, and `browser_stack=4`.

| Measurement | Current value |
| ----------- | ------------- |
| External / scheduler wall | 250.54 s / 248.132 s |
| Coverage | 259/259 work units; 1,074 tests; 427 authoritative, 2 support, and 1 supplemental phase-plan rows; zero failed, missing, or unmapped evidence |
| Logical / executed / reused / derived | 1,457.890 s / 1,457.890 s / 0 / 24.576 s |
| Setup / target teardown / service cleanup | 9.074 s / 0.592 s / 4.248 s |
| Backend aggregate finalizers | 4.837 s integration; 5.242 s store; 1.144 s process |
| Fixture work | 883 operations; 90.448 s aggregate; 40.199 s migration-scratch drop; 33.371 s across 241 isolated clones |
| Critical path | frontend install 1.777 s; web build 6.477 s; embedded assets 8.066 s; backend-unit 227.398 s |
| Service completion without backend-unit | 142.705 s; integration/store/process lane totals 427.349/363.420/66.597 s |
| Other long work | frontend unit 65.828 s; frontend typecheck 24.893 s; browser shard max 60.601 s; stateful lane sum 83.857 s; Go lint 18.432 s; vulnerability/Gosec 14.347/14.581 s |
| Scheduler utilization | maximum 17; average 7.53; 44.3% of declared capacity; dominant blockers `host_io`, `host_cpu`, service readiness, and `postgres_clone` |
| Artifacts | 31.5 MB; 6,747 files; principal run/check summaries and events approximately 16.5 MB |
| Structural size | 140 targets; 98,054 generated-Make bytes; 1,067 scheduler units in 1,538,481 bytes; 147 bytes/synthetic target; 358 bytes/synthetic scheduler unit; byte-stable scratch renders |

`backend-unit` executes 128 actual private families across 24 packages for 201
plan rows: 158 authoritative, 42 support, and one raw. Phase 12 contributes 97
authoritative Network Flow rows. The service schedule has 204 Go shards; 133
`clone_heavy` shards claim both two broad I/O slots and the precise
`postgres_clone` resource. Frontend, browser, static analysis, generation,
readiness caches, and lifecycle work are required but are not the initial
critical-path owners.

### 7.3 Work protocol and tracker

Exactly one row may be `IN_PROGRESS`. `QUEUED` is an intermediate state only.
Before another row starts, the active row must record its status, changes,
timing, validation roots, failures, contaminated/stale evidence, retries,
compatibility disposition, rollback decision, and exit result. Final statuses
are `DONE`, `DROPPED`, or genuinely justified `BLOCKED`; a blocked row must name
the external blocker, owner, exhausted alternatives, evidence, and retry
condition. Final handoff may contain no `QUEUED`, `TODO`, `IN_PROGRESS`,
`DEFERRED`, or unjustified `BLOCKED` row.

| ID | Workstream | Status | Depends on | Expected contribution | Exit condition |
| -- | ---------- | ------ | ---------- | --------------------- | -------------- |
| DX-041 | Establish the current baseline and lock the performance/conformance contract. | DONE | DX-040 | Measurement only. | Tracker and NLSpec own the objective, strict five-run statistic, compatibility boundary, baseline, and evidence fields; structural validation passes. |
| DX-042 | Consolidate compatible pure backend execution families. | DONE | DX-041 | Reduce backend-unit from 227 s to at most 40 s. | Exact 201-row/test/owner parity; no over-selection; median warm backend-unit improvement at least 70%; retired private IDs have no shims. |
| DX-043 | Recalibrate service resource claims and scarce-resource ordering through a bounded experiment. | DROPPED | DX-042 | Reduce service completion by 25–40 s. | Adopt only with paired-run thresholds, fixture/lifecycle parity, and no capacity increase; otherwise remove experiment and record `DROPPED`. |
| DX-044 | Conditionally add deterministic compatible service Go batching. | DONE | DX-043 | Additional 10–30 s only if required. | Trigger only when post-DX-043 probes fail the 117 s margin; adopt with exact scenario/resource parity and thresholds or delete and mark `DROPPED`. |
| DX-045 | Harden cache, artifact, lifecycle, and synthetic-growth invariants. | DONE | DX-042–DX-044 | 0–5 s; prevents regression. | Stale/mismatched reuse fails closed; lifecycle/artifacts remain exact; compatible-row growth is bounded and deterministic. |
| DX-046 | Run cross-cutting conformance and representative validation. | DONE | DX-045 | Validation only. | Structural/drift, harness, representative base/frontend/browser, security, isolation, and cleanup gates pass with every root classified. |
| DX-047 | Run final five-run timing acceptance. | IN_PROGRESS | DX-046 | Prove the target. | One warm run plus five consecutive current-input measured runs; maximum external wall time strictly below 120 s; exact coverage and cleanup parity. |
| DX-048 | Complete the final tracker handoff. | QUEUED | DX-047 | Handoff only. | All rows have final status; before/after, compatibility, roots, failures/retries/skips, residual risks, and next measured seam are complete. |

DX-043 compares three control and three candidate warm checks. The candidate may
reduce `clone_heavy` broad I/O from two to one only while retaining
`postgres_clone=1` and its limit of eight, and may prioritize process-constrained
backend work without raising any capacity. Adoption requires median service
completion at most 112 seconds, at least 15 percent improvement, candidate check
median at most 112 seconds and maximum at most 117 seconds, clone/reset fixture
p95 regression no greater than 10 percent, successful cleanup, and per-run swap
growth below 512 MiB.

DX-044 triggers only if DX-043 is dropped or three post-DX-043 checks are not all
below 117 seconds. Its fixed candidate groups compatible items by target,
package, fixture policy/budget, runtime binaries, isolation class, and resource
profile, then uses stable longest-estimate-first packing into 12-second batches
of at most eight symbols with row-ID tie breaking. Incomplete batch evidence
fails closed. Representative Phase 9 store and Phase 11 integration targets must
improve at least 20 percent before full-schedule adoption.

### 7.4 Ownership and compatibility disposition

1. Core 00 through Core 04 own product behavior.
2. Phase registries and phase maps own exact evidence routing.
3. The harness NLSpec owns batching, scheduling, logical resources, lifecycle,
   artifacts, failure behavior, and performance acceptance.
4. Authored phase maps, execution topology, resource registry, and duration
   inputs drive generated schedules and ledgers.
5. Runners execute every selected symbol and emit current-run artifacts.
6. Summaries and diagnostics prove coverage, timing, cleanup, and failure parity.

| Surface | Iteration 4 disposition |
| ------- | ----------------------- |
| Public Make targets, command IDs, inputs, schemas, output/failure/cleanup contracts | Retain; continuing public value and exact identity parity are acceptance gates. |
| Core requirements, acceptance criteria, phase row IDs, symbols, evidence classes and owners | Retain exactly; optimizations change execution grouping only. |
| Private fragmented backend execution families and, if DX-044 triggers, private shard IDs | Replace atomically; delete old paths and keys without shims or dual readers. |
| Input-addressed build/static caches and same-run helper refs | Retain; validate digests and fail closed. |
| Cross-run product-test result reuse | Remove/reject; it has no accepted conformance value. |
| Scheduler manifest/event/topology/task-surface schemas | Retain unless an implementation defect proves a clean version cut necessary. |
| SeaweedFS support, Playwright fallback, browser isolation, thin facades, agent-finalizer v3 child artifacts | Retain; current consumers remain and these surfaces are not measured critical-path defects. |
| Release-check | Excluded by scope; never used as Iteration 4 evidence or rollback. |

### 7.5 Per-workstream evidence record

Every workstream log entry must populate: start/end time; commit and authored-
input digest; problem and current evidence; affected Core/phase routes; changed
specifications, owner inputs, implementation, tests, schemas, generated outputs,
diagnostics, and documentation; before/after wall and critical-path time;
logical/executed/reused/fixture/setup/teardown/finalization time; host and
scheduler capacity; expected and actual benefit; public/private compatibility;
migration/deletion; validation commands and roots; successful, failed,
contaminated, stale, retried, and skipped evidence; rollback decision; exit
criteria; status; and residual risk.

| Date/time | ID | Change and ownership | Timing and validation | Failures, retries, compatibility, next action |
| --------- | -- | -------------------- | --------------------- | -------------------------------------------- |
| 2026-07-12 | DX-041 | Opened Iteration 4 in the existing tracker; locked the sub-120-second method and pure-Go family-consolidation contract without changing Core or domain vocabulary. | Baseline `p21895`; JSON shape `p32200`; Markdown lint `p32201`; task-surface report `check=pass`; target plan remains 201 rows/128 families/24 packages; whitespace audit passes. | Historical DX-021–DX-040 records are preserved. Public identity remains 96 targets and generated/growth measurements are unchanged. Begin DX-042. |
| 2026-07-12 | DX-042 | Consolidated 97 Phase 12 Network Flow, 19 Phase 11 incident-bundle/reference-pack/reporting, and their section labels into compatible owner-derived families; made shared-report emission always use its exact owner row IDs. | Families 128→16; final schedule generation `p48101`; backend-unit roots `p48809`, `p50778`, `p52438` pass with 204 tests, 158 authoritative, 42 support, four raw, 24 packages, and zero missing/unmapped; external times 15.62/14.31/14.17 s, median 14.31 s (93.7% below the 227.398 s scheduled baseline); JSON `p54130`, schedule drift `p54129`, ledger drift `p54142`, generation drift `p54134`, and execution smoke pass. | JSON `p35292` correctly rejected stale projections. `p36620` overcounted shared reports; `p39207` failed closed on unexpected row IDs; `p42134`, `p44204`, and `p45806` exposed a collapsed unit/integration label. All are failed or parity-contaminated evidence, superseded by exact roots. Old private family IDs are removed without shims. Begin DX-043. |
| 2026-07-12 | DX-043 | Compared the current two-I/O clone profile against a candidate one-I/O profile retaining `postgres_clone=1`, plus deterministic process-target priority, without raising capacity or changing selected work. | Control roots `p25441`, `p21574`, `p17551`: external 126.43/129.19/129.33 s, service 115.258/116.734/115.240 s. Candidate roots `p15332`, `p11471`, `p7547`: external 119.85/125.79/119.56 s, service 108.016/114.385/107.583 s; all 259/259 and 1,074 tests. | Candidate median service improvement was only 6.3%, below 15%, and maximum external time 125.79 s exceeded 117 s. Experimental spec, policy, topology, tests, and generated changes were removed completely. Failed `p80869` exposed over-broad service emission; `p81436` rejected stale projections; both are non-performance evidence. DX-044 trigger is satisfied. |
| 2026-07-12 | DX-044 | Adopted NLSpec-owned deterministic exact-symbol service batching: compatible items share 12 s/8-symbol shards; package, runtime-binary, complete fixture budget, isolation, evidence class, and scheduler-profile boundaries remain closed. Phase 11 integration families were atomically replaced by seven package-aligned private families with no aliases. | Service Go shards 204→110 for the same 332 executable items; representative Phase 9 store processes/executed time 25/65.484 s→9/27.795 s (57.6% lower) and Phase 11 integration 30/72.948 s→7/38.298 s (47.5% lower). Phase 9/11 slice roots `p6990`/`p21860` pass 53/33 tests. Full roots `p35001`, `p13260`, `p91146` pass 133/133 units and 1,074 tests with zero missing/unmapped; external 97.97/97.57/98.29 s and service 85.694/86.406/86.202 s, all below adoption thresholds. Schedule `p5949`, JSON `p6094`, schedule drift `p6405`, ledger drift `p6547`, and harness contract `p5202` pass. | Public targets, command IDs, row/scenario/owner routing, fixture/resource policies, schemas, artifacts, failures, and cleanup remain stable. Obsolete private Phase 11 family/shard IDs are deleted without readers or shims. The experiment exceeded the 20% representative and full-check thresholds, so rollback was not triggered. Begin DX-045 hardening. |
| 2026-07-12 | DX-045 | Added compatible/incompatible/isolated/oversized exact-batch and deterministic 25/100-row growth fixtures; registered the test with the authored task-surface owner. Made scheduler growth gates composition-invariant by retaining overall metrics while separately gating ordinary and structurally-wide p95 populations. Existing cache, same-run helper, partial/stale evidence, lifecycle, failure, and cleanup fixtures remain authoritative; no cross-run test-result reuse was added. | Growth fixture: 25 rows→4 shards and 100→13, max eight, byte-identical under reversed input. Current scheduler projection is 755,759 bytes/519 units, 1,457 bytes/unit, overall p95 2,014, ordinary p95 1,043/1,500, wide p95 4,534/5,000, max 4,716/12 KiB, synthetic 358 bytes/unit, byte-stable. Post-batch full root `p35001` is 23.28 MB/3,768 files versus 31.5 MB/6,747 baseline. Non-cleanup collation is not a new workstream: the longest finalizer is 3.553 s (3.7% wall), below 5 s and 5%; existing emission remains. Harness `p74076`, artifact policy `p74690`, JSON `p74833`, schedule drift `p75144`, ledger drift `p75281`, generation drift `p75600`, and task report pass. | No public/schema compatibility cut. The former single global p95 could fail merely because compact Go units were removed; the new kind-stratified p95 plus unchanged bytes/unit and max gates closes that denominator defect rather than relaxing growth limits. Stale/missing/failed cache and helper evidence continues to fail closed. Begin DX-046 broad validation. |
| 2026-07-12 | DX-046 | Ran finalizer-led structural, negative, lifecycle, execution, base/service phase, frontend namespace, and browser lifecycle validation. Updated three private smoke fixtures to consume shard-plan-owned batches and exact embedded scenario items instead of retired scenario-named shard paths. | Finalizer `p77171` passed with generated output unchanged; retained duration maintenance was skipped because no current-input full root was supplied. Base Phase 4/9/12 roots `p78560`/`p96822`/`p12549` pass 51/70/134 tests; service roots `p25899`/`p43748`/`p58660` pass 44/53/37. FE-P8 exact `p71686` passes 6/6 and 59 tests; cumulative `p2779` passes 11/11 and 85. Browser webserver-backed `p37375` passes 97 tests; stateful `p54515` passes 22. Fast/lifecycle/execution tiers `p72846`/`p73852`/`p76045` pass; extended `p39714` passes 48/48. Current harness/structural roots are `p74076`, `p74690`, `p74833`, `p75144`, `p75281`, and `p75600`; task report and whitespace audit pass. | Extended roots `p1559`, `p14857`, `p55151`, and `p95137` failed closed on stale scenario-shard, clone-user-count, and Go-runner assumptions; none is success or timing evidence. The final fixtures retain all scenario IDs, symbols, owners, runtime binaries, clone claims, and exact batch selectors. No required gate is skipped; begin DX-047 timing acceptance. |
| 2026-07-12 | DX-047 | The first acceptance attempt was invalidated after its second measured run exposed migration-scratch teardown variance. The attached-suite migration helper now retains each fresh scratch database for scheduler-owned stack teardown; standalone invocations still drop during test cleanup. The Phase 0 migration-evidence row uses that owner helper instead of a manual critical-path drop. | Readiness root `p4762` passed at 110.479 s scheduler. Measured root `p84638` passed at 106.38 s external/104.528 s scheduler, but `p63408` passed coverage at 122.76/120.990 s and therefore fails performance acceptance. Its migration-evidence shard was 97.059 s, including a contended forced drop. After the lifecycle change, Phase 0 service root `p45402` passes 18 tests; the same exact row is 4.59 s and its shard 6.108 s. | `p84638` and `p63408` are an abandoned partial window, never acceptance evidence. The change preserves fresh identity, exact migration/product assertions, fixture attribution, standalone cleanup, attached owned-stack cleanup, and failure semantics; it adds no cache or reuse. A new readiness run and all five measurements are required after broad validation. |

### 7.6 Final acceptance method

After final inputs and generated outputs stabilize, run one successful unmeasured
warm `make check`, then five consecutive measured canonical runs. External
process wall time from `/usr/bin/time -p` controls; scheduler wall/critical-path
time remains required comparison evidence. The nearest-rank p90 is the maximum
of five and must be strictly below 120.00 seconds.

A qualifying run has the same authored-input and public identity digests,
documented capacity, exact target/row/test coverage, zero missing/unmapped
evidence, scheduler timing validity, byte-stable generated output, no stale or
cross-run product-test reuse, valid artifacts, and successful cleanup. Failed,
interrupted, input/capacity-mismatched, stale, retried, or otherwise contaminated
roots remain recorded but cannot enter the window. Any input change restarts the
warm run and all five measurements.

Final validation uses `agent-finalize`, structural/drift gates, harness contract
and relevant smoke tiers, representative base and service-backed Phases 4/9/12,
exact and cumulative FE-P8, and affected browser lifecycle targets. Duration
maintenance may use only a successful full warm current-input check root and
forces regeneration plus repetition of affected gates before timing acceptance.

### 7.7 Residual-risk and next-seam rule

The final next seam must come from measured post-acceptance evidence: remaining
service batching, frontend/browser critical path, non-cleanup artifact collation,
or the retained agent-finalizer child-artifact owner review. An unmeasured seam
must not be recommended. DX-048 must record before/after timing, coverage parity,
all validation roots and classified failures/retries/skips, final compatibility
dispositions, and the selected measured seam.
