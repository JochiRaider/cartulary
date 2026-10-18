# Testing-harness remediation handoff

This durable handoff preserves the temporary controlling tracker at
`temp/plan-notes.md`. It is a human record, not a runtime requirement source.
The adopted Testing Harness NLSpec and subsystem owners remain authoritative.
No executable tool, generator, conformance check or release evidence may consume
this document or any other Markdown.

## Current state and restart instructions

Snapshot date: 2026-10-18. Branch: `codex/specification-remediation-20261016`.
Starting commit: `f3b36a19c3201b51d07e1aecdb569be357288388`.
Changes are uncommitted. W00–W09 are checkpointed; W10 is in progress;
W11 has not started. The effort is **not complete**.

The user removed harness performance qualification and baseline publication
from this effort and its specifications. The adopted public surface contains
62 commands. Product measurement, descriptive timing, security and complete
WSL2 check/CI/release acceptance remain required.

Work is **stopped at the user's request**, after E-199's full-assurance command exited and graph cleanup completed. No invocation is left running by this task, no follow-up is scheduled, and no further repair or qualification will run without a resume request. W10/R8 remain IN_PROGRESS; W11 remains TODO.

Current source is `sha256:00e3a62e6cb960f32bbd15e2b4c49ac62ac48f1f992823f628f5d4d3c3c045f4`, at HEAD `afd6c699980d4256af980a9abb906b2d79004755` plus the uncommitted change set. Finalizer passed with zero mutations and RESULTS_DIR unset; source-check all and Markdown lint pass. Full assurance `remediation-w10-candidate-full-20261018-03` exited Make 2 / normalized 10: 57/58 parent units passed. Human check passed 800/800; machine check failed 793/800; fast passed 582/582. All graph cleanup records completed, but seven machine Go units retain eleven test-level PostgreSQL cleanup failures. Exact roots, rows and counts are recorded in E-199 below.

The outstanding implementation issue is PostgreSQL forced-drop deadlines under concurrent assurance. Current typed diagnostics show checkpoint waits, usually with checkpointer IO/DataFileSync. Removing the redundant per-clone checkpoint did not resolve the broader failure. The exact physical I/O cause is unknown; do not conflate these observations with the earlier create-database ProcSignalBarrier waits. The extension concurrency timing repair passed its exact row in both checks; the latest extension failure is a different row's cleanup.

On a resume request, begin with E-199's exact retained machine failures and diagnose before another repair or retry. R1–R7/R6a, R8a/R8b/R8c, R9 and R10 implementation checkpoints remain recorded as DONE; R8 acceptance remains open. Any further executable change needs its own checkpoint and appropriate finalization. Then complete applicable full assurance, sequential CI/release and exact audits, checkpoint W10, and perform W11's final reconciliation and handoff. No current-source complete CI or release acceptance exists; final-candidate release has never started.

Preserve failed evidence, baseline snapshot and private recovery proof. No reset, recovery deletion, timeout/capacity increase or pin change was performed. The finalizer's retained-run maintenance was skipped because RESULTS_DIR was unset when no eligible same-source warm check existed. Do not rerun it merely to attach the later passing human child.

Stopping-point documentation validation passed (`make lint-markdown`, outer 0; `git diff --check`, 0); see E-200. W11 remains open.

## Acceptance scope

| Claim | Current state |
| --- | --- |
| Planning and adopted contracts | Complete for W00–W08; subsequent implementation findings are recorded under W09/W10. |
| Focused correctness | Repairs have passing focused evidence with the applicability and outer-exit limits recorded below. |
| Full harness assurance | E-199 failed 57/58, outer Make 2 / normalized 10. Human check passed 800/800; machine check failed 793/800 with PostgreSQL checkpoint-wait cleanup failures. R8 remains open. Older full passes are historical. |
| Final check, CI and release readiness | Human child check passes on current source, but machine check/full assurance fail. Current-source CI, complete release and exact audits remain unrun. W10 and W11 are open; execution stopped at user request. |
| Harness performance qualification and baseline publication | Removed by the user; no qualification or publication is performed or claimed. |
| Product measurement, security and installed-package qualification | Remain required in their owning gates. |
| Core 05 publication, formal CMVP applicability or deployment approval | Not established by this effort's engineering checks. |

## Migration and rollback

The coordinated cutover has no compatibility aliases. Use `make help ALL=1`
for the active surface, `make task-guide OWNER=<owner>` for narrow guidance,
`make test-slice OWNER=<owner> ROWS=<exact-rows>` or a declared suite for focused
verification, and `make browser-e2e SUITE=<suite>` with an explicit suite.
`ci` replaces the removed comprehensive `test` gate. `check` prepares artifacts
for selected consumers; production qualification belongs to CI/release.

Command, manifest, topology, graph and family wire versions changed. Runtime
plans are v5, graphs v6, retained contexts v3; safe cleanup diagnostics use v1.
Current row results are v4; Go helper failure envelopes are v2 with a separate v1 Go
diagnostic bundle; browser group/target results are v8/v7. Historical runs retain
their original identities. Changed contracts and executable fingerprints invalidate
incompatible evidence and caches; do not relabel or translate them into acceptance.

After owned invocations have finished, roll back the coordinated source change
set as a unit, preserving retained evidence, runtime recovery proof, and
pre-existing user changes, including the Git-ignore update commit. There is no
product-data migration. Do not restore retired aliases as a recovery mechanism.

## Complete working-tracker snapshot

# Testing-harness remediation execution tracker

## 1. Scope and Source Posture

- Controlling artifact: `/home/jochi/code/cartulary/temp/plan-notes.md`; sole tracker. It remains temporary and Git-ignored. The durable human handoff and full tracker snapshot are retained at `docs/handoffs/testing-harness-remediation.md`.
- Canonical target: the testing-harness command/selection/execution/evidence seam rooted in `tools`, with relevant `contracts/verification` and `internal/testutil` projections.
- Artifact mode: normal. Planning completeness: complete for the identified gaps. Implementation authorization: explicitly granted by the user on 2026-10-17; earlier planning-only restrictions do not limit this execution.
- Authority: adopted Testing Harness NLSpec; Core 00–04 and OTel owner for product/security obligations; Core 05 only for claim publication. `docs/domain.md` owns vocabulary, not harness behavior.
- Framework: `docs/handoffs/cartulary_modular_refactor_planning_framework.md`; retain the refactor-tracker twelve-section structure.
- User decisions: reevaluate surface; retain justified command consolidation, retiring the two performance-qualification/publication commands; require browser `SUITE`; coordinated hard cutover without aliases.
- Preserve the existing scheduler. No second registry/scheduler. Tests, generators, runtime metadata and release evidence must never consume Markdown.
- Overall implementation status: IN_PROGRESS; work stopped at the user-requested terminal checkpoint on 2026-10-18 (E-199/E-200). W00–W09 are DONE. W10/R8 remain IN_PROGRESS after full assurance failed 57/58; W11 remains TODO. No invocation is left running by this task. Resume only on user instruction; durable handoff and full tracker snapshot are retained.

| Baseline item | Starting evidence | Comparison / limits |
| --- | --- | --- |
| Branch / commit | `codex/specification-remediation-20261016` / `f3b36a19c3201b51d07e1aecdb569be357288388` | Revalidated at execution start. |
| Index, tracked and non-ignored untracked files | Clean status; 7,392 content/identity records and index digest in `/tmp/cartulary-remediation-baseline-6rr9ozw1/baseline.json` | Recompare at handoff; do not revert concurrent changes. |
| Original tracker | SHA-256 `961c4e6c934db0e9df3d36c74fb6d67bf2dc5c4a1df04619725b4ccca66221df`; regular single-link file, no linked parent | Original retained at `/tmp/cartulary-remediation-baseline-6rr9ozw1/plan-notes.original.md`. |
| Ignored operational outputs | Existing `.cartulary`, `.cache`, dependencies and build outputs | Not content-snapshotted; canonical validation may write declared outputs. |
| Qualification | No published `tools/harness_public_target_duration_baselines.v3.json`; no retained graph run names starting source commit | Reference performance requirement removed by user; no historical pass claimed. |

## 2. Current-State Repository Inventory

The inspected seam includes the task-surface/work-graph owners, Make and generated bindings, compiler and browser selection, finalizer, performance readers, OTel conformance consumer, native runtime/fixture catalogs, owner diagnostics, and dev-service lifecycle. This is a targeted seam inventory, not a claim of behavioral inspection of all 7,392 repository files. Additional concrete edit sources must be inspected before mutation.

| Source group | Disposition / responsibility | Owner / risk |
| --- | --- | --- |
| `tools/task_surface_owner.json`, `tools/harness_work_graph_owner.json` | Directly inspected; 190 command entries and graph policy | Harness; public/private identity coupling |
| `tools/execution_topology_manifest.json`, `tools/generated_artifact_policy.json`, `tools/toolchain_pins.json` | Metadata inspected; authored topology, generated ownership, pins | Harness; generated hand-edit prohibition |
| `tools/harness/scheduler/work-graph`, command-surface and generated-artifacts modules | Relevant entrypoints inspected; exact edits revalidate callers | Harness; selection/dispatch/preparation |
| `tools/harness/finalization`, observability and evidence accounting | Relevant entrypoints inspected | Harness; freshness, rollback, false qualification |
| `tools/otel/check-otel-conformance.mjs`, telemetry catalog/verification | Directly inspected selection consumer/catalog | OTel and Harness; required runtime/privacy evidence |
| `tools/test_families/harness.*.json`, relevant `internal/testutil` assertions | Catalog selection inspected; assertions require pre-move characterization | Harness; race/context distinctions |
| Generated Make/catalog/schema/runtime projections | Metadata only until regenerated through Make | Generated, never hand-edit |
| Product UI/domain implementations and unrelated owners | Excluded: no product behavior change requested | No vocabulary or UI redesign |

## 3. Module Boundary Diagnosis

| Responsibility | Decision | Durable owner |
| --- | --- | --- |
| Public names, selectors/defaults, output/side-effect policy | Keep and simplify | Existing task-surface owner |
| Canonical operations, dependencies and aggregate unions | Split from public identity | Existing work-graph owner |
| Assertion selection, tier, fixture/evidence semantics | Keep explicit | Catalog/verification owners |
| Resource and runtime/artifact bindings | Keep | Execution topology |
| Human guidance and migration mappings | Keep outside runtime | This tracker and normative/supporting documentation |

## 4. Public Contract and Behavior Freeze Map

Owner-aligned assurance survives; accidental scope expansion and redundant wrappers do not become compatibility promises. Public rename alone cannot duplicate physical work. Fixtures, runtime binary, race mode, isolation, security policy and evidence class remain part of semantic identity.

| Contract | Adopted owner / change posture | Required result |
| --- | --- | --- |
| C-000 | Human controlling tracker; no runtime authority | Traceable, checkpointed, evidence-backed execution |
| C-001 | TH-HARNESS-REQ-001, 050–058, 062, 073, 082; amendment required for new surface | One coherent normative command/evidence contract; no harness performance qualification or baseline publication |
| C-002 | TH-HARNESS-REQ-801, 802, 807; interface amendment required | Public bindings over context-sensitive canonical private operations |
| C-003 | TH-HARNESS-REQ-056, 800; existing-requirement repair; OTEL-REQ-150 retained | Exact browser row closure and explicit required policy dependencies |
| C-004 | TH-HARNESS-REQ-053, 054, 801; placement/deduplication repair | Three narrow smoke roles; broad full-tier regressions; distinct assurance contexts |
| C-005 | TH-HARNESS-REQ-260, 261, 379; shared-context permission clarified | Immutable fresh validation facts; separate receipts; retained identity and transactional rollback |
| C-006 | TH-HARNESS-REQ-054, 070, 800 and Section 4 family defaults; four-gate amendment | Consumer-driven preparation; preserved blocking checks; advisory reports outside local check |
| C-007 | TH-HARNESS-REQ-068, 069, 071–076, 112, 202, 816; selector/lifecycle amendments | Closed inputs, exact scope, read-only discovery, preserved lifecycle boundaries |
| C-008 | TH-HARNESS-REQ-077–080, 378–379 and Section 11.6; variant amendment | Exact retained selection/diagnostic evidence; no invented historical timings; qualification retired |
| C-009 | TH-HARNESS-REQ-050, 066, 073; hard cutover retained | Owners, generated outputs, active callers and guidance agree without aliases |
| C-010 | Harness correctness, security and WSL2 release clauses | Exact final-source correctness evidence, exact package and release acceptance |
| C-011 | User's approved completion requirements and repository procedure | Final-source evidence and honest handoff; no closure from partial passes |

| Gate | Final row scope | Required policy |
| --- | --- | --- |
| `test-fast` | fast | Only preparation required by selected product/architecture rows and narrow smoke assertions |
| `check` | fast + standard | Blocking source/static/security/evidence integrity and exactly three smoke roles; consumer-driven builds |
| `ci` | fast + standard + full | Inherited blocking policy, production qualification, advisory reports, required harness contracts |
| `release-check` | all tiers | Inherited CI policy; package, compatibility, security and all installed-package scenarios against one package identity |

These are alternative scopes, not a mandatory sequence for ordinary changes.
Removing `test` intentionally removes the comprehensive tests-only aggregate;
focused suites and owner selection remain available.

**The final public surface should contain these 62 commands.**

Names without a changed contract below retain their current purpose and safety boundaries.

| Area | Final public targets |
|---|---|
| Primary gates — 4 | `test-fast`, `check`, `ci`, `release-check` |
| Focused verification — 4 | `test-slice`, `browser-e2e`, `browser-e2e-visual-update`, `harness-contract` |
| Checks and analysis — 10 | `lint`, `lint-markdown`, **`source-check`**, `migration-drift`, `openapi-compatibility-check`, `otel-conformance`, `go-vulncheck`, `go-gosec-targeted`, `go-gosec-audit`, `frontend-fallow-static` |
| Authoring and builds — 5 | `generate`, `agent-finalize`, `format`, `build`, `frontend-lockfile-update` |
| Setup and development — 9 | `doctor`, `bootstrap`, `services-up`, `services-down`, `db-migrate`, `db-reset`, `postgres-baseline-reset`, `object-store-reset`, `dev` |
| Reusable test services — 3 | `test-services-session-up`, `test-services-session-status`, `test-services-session-down` |
| Discovery and observation — 6 | `help`, `task-guide`, `author-test-row-id`, `target-plan`, `explain-run`, `test-run-status` |
| Evidence and observation — 4 | `test-evidence-audit`, `harness-observability-check`, `harness-otel-export`, `benchmark-claim-check` |
| Release qualification — 6 | `package-release`, `package-inspect`, `seaweedfs-compatibility`, `seaweedfs-release-gate`, `cryptographic-policy-assessment`, `credential-capacity-assessment` |
| UI review — 8 | `browser-design-review`, `ui-review`, `ui-review-status`, `ui-browser`, `ui-capture`, `ui-analyze`, `ui-review-report`, `ui-review-stop` |
| Cleanup and recovery — 3 | `clean`, `distclean`, `test-run-recover` |

**Consolidated targets need closed, typed selectors—not arbitrary forwarded arguments.**

| Target | Proposed contract |
|---|---|
| `help` | Optional `ALL=1` and `JSON=1`. Read-only discovery; complete output includes input contracts and side effects. |
| `bootstrap` | Optional `PART=node\|frontend\|browser\|test-images`. Omission retains current full-bootstrap scope. Frontend preparation includes its toolchain; standalone pnpm preparation becomes private. |
| `services-up` | Start local services and establish their required readiness, including object-store initialization, once. Absorb the useful `db-up` behavior without its repeated startup/readiness path. |
| `test-slice` | Exactly one of `OWNER=<id>` or `SUITE=<name>`. `ROWS` is valid only with `OWNER`. `SCOPE=all\|services`, default `all`. Suites: `backend-unit`, `backend-store`, `backend-integration`, `backend-process`, `frontend-unit`. Integration includes its support rows while preserving their support evidence classification. |
| `browser-e2e` | Required `SUITE=all\|functional\|stateful\|measurement\|accessibility\|visual\|support`; omission fails before preparation. Compile the union of exact canonical stage selections. Direct selections are not aggregate-tier filtered. |
| `harness-contract` | Default `SUITE=contract` preserves the current public contract union. Additional presets: `smoke`, `execution`, `regression`, `lifecycle`, `full`, `runtime`, `fixtures`, `runtime-boundary`, `artifact-consumers`, `ui-review`. Each preset has an explicit assertion inventory. |
| `lint` | `CHECK=standard\|go\|types\|frontend\|scripts\|shell\|frontend-boundary\|backend-boundary\|protocol`. Default `standard` preserves the current blocking lint union. Protocol analysis remains a distinct blocking operation. |
| `source-check` | New source/contract validation entry point. `CHECK=all\|generated\|schema\|catalog\|surface\|identity\|toolchain\|migration-input`. Default `all`; includes required configuration integrity checks. May produce scratch and retained diagnostics when executed, but does not update source or run product test suites. |
| `build` | `COMPONENT=production\|server\|server-harness\|migrate\|operator\|web\|web-measurement\|report-renderer`. Default `production` preserves the current server/migrate/operator/web build union. |
| `package-inspect` | `CHECK=archive\|startup\|reference-pack\|recovery\|all`, default `archive`. `all` includes archive inspection and all three installed-package smoke scenarios against the same package identity. |
| `task-guide` | Required `OWNER`; `DETAIL=summary\|rows\|coverage`. Absorb owner explanation. Remove the constant `ROLE=module-author` argument; the command remains module-author guidance. |
| `target-plan` | Required `TARGET`; optional `INPUTS=<JSON-file>`, `DETAIL=summary\|graph\|rows\|artifacts`, `JSON=1`. Resolve inputs through the execution command’s typed contract. Compile only; acquire no resources. |
| `explain-run` | Preserve existing views; absorb fixture reporting and duration observations through `DETAIL=fixtures` and `DETAIL=performance`. Newly consolidated diagnostics require an exact retained invocation. |
| `harness-observability-check` | `CHECK=all\|events\|timing\|evidence\|diagnostics`, default `all`. Strictly read-only validation of selected retained evidence. A partial check cannot claim complete invocation validity. |
| `agent-finalize` | Preserve its public inputs and maintenance-only role. Share underlying validation context while retaining distinct action outcomes, ordering, rollback and retained-run checks. |

Invalid combinations must fail before preparation. In particular, an explicit row outside the selected owner or service scope must be rejected, not silently discarded.

Bare `make browser-e2e` is a usage error. `SUITE=all` explicitly selects every browser stage; no aggregate tier is inherited by direct selection. This supersedes the initial all-by-default proposal.

The `harness-contract` presets must not blur the current distinction between the public contract union and the older smoke tiers. `regression` absorbs the old extended smoke inventory; `full` retains the complete harness assurance inventory. CI’s default contract selection must not silently become the largest preset.

**Retire these 60 public names without compatibility aliases.**

The following table accounts for every removed public target.

| Existing public target(s) | Replacement or disposition |
|---|---|
| `harness-performance-check`, `harness-public-target-duration-baselines` | Retire qualification and publication with their exclusive readers, schemas and gates. |
| `help-all` | `help ALL=1` |
| `bootstrap-node-runtime` | `bootstrap PART=node` |
| `frontend-toolchain`, `frontend-install` | `bootstrap PART=frontend`; toolchain-only preparation remains private |
| `playwright-install` | `bootstrap PART=browser` |
| `db-up`, `object-store-init` | `services-up` |
| `generate-drift`, `generated-artifact-policy-check` | `source-check CHECK=generated`; retain both validations |
| `json-shape-check` | `source-check CHECK=schema` |
| `toolchain-drift` | `source-check CHECK=toolchain` |
| `task-surface-report` | Discovery through `help ALL=1 JSON=1`; validation through `source-check CHECK=surface` |
| `service-backed-test-slice` | `test-slice OWNER=… SCOPE=services` |
| `backend-unit`, `backend-store`, `backend-integration`, `backend-process`, `frontend-unit` | Corresponding `test-slice SUITE=…` |
| `target-plan-json` | `target-plan … JSON=1` |
| `explain-target` | `target-plan … DETAIL=graph` or `DETAIL=artifacts` |
| `fixture-report` | `explain-run … DETAIL=fixtures` |
| `explain-test-owner` | `task-guide OWNER=… DETAIL=coverage` |
| `go-test-duration-baselines`, `go-test-duration-baseline-drift` | `explain-run … DETAIL=performance` |
| `browser-e2e-duration-baselines`, `browser-e2e-duration-baseline-drift` | Same retained-run performance view |
| `service-backed-make-target-duration-baselines`, `service-backed-make-target-duration-baseline-drift` | Same retained-run performance view |
| `harness-smoke-duration-baselines`, `harness-smoke-duration-baseline-drift` | Same retained-run performance view |
| `go-test-duration-baseline-coverage` | Retire qualification-only roster gate; retain ordinary input/observability integrity |
| `scheduler-event-order-drift` | `harness-observability-check … CHECK=events` |
| `scheduler-summary-timing-drift` | `harness-observability-check … CHECK=timing` |
| `frontend-typecheck` | `lint CHECK=types` |
| `frontend-import-boundary-check` | `lint CHECK=frontend-boundary` |
| `backend-module-boundary-check` | `lint CHECK=backend-boundary` |
| `lint-biome`, `lint-scripts`, `lint-shell` | `lint CHECK=frontend`, `scripts`, `shell`, respectively |
| `browser-e2e-webserver-backed` | `browser-e2e SUITE=functional` |
| `browser-e2e-stateful`, `browser-e2e-measurement`, `browser-e2e-a11y`, `browser-e2e-visual` | Corresponding browser suite: `stateful`, `measurement`, `accessibility`, `visual` |
| `test` | `ci` for comprehensive validation; focused selectors for narrower work |
| `release-readiness-evidence` | Private release-evidence assembly/validation inside `release-check`; retained results remain inspectable |
| `seaweedfs-release-evidence` | Private reporting phase of `seaweedfs-release-gate`, including retained failure evidence |
| `build-server`, `build-server-harness`, `build-migrate`, `build-operator`, `build-web`, `build-web-measurement`, `build-report-renderer` | Corresponding `build COMPONENT=…` |
| `standup-package-smoke`, `standup-reference-pack-smoke`, `standup-operational-recovery-smoke` | `package-inspect CHECK=startup`, `reference-pack`, `recovery`, respectively |
| `browser-design-review-smoke` | Private lifecycle assurance reachable through owner selection and `harness-contract SUITE=lifecycle` |

The eight duration “baseline/drift” names deserve special treatment: they already dispatch to observation, and the implementation’s parsed view is not used to distinguish their output. They should be consolidated as diagnostics, not replaced with new baseline writers. The [performance CLI](/home/jochi/code/cartulary/tools/harness/observability/canonical-performance-cli.mjs:29) supports this distinction. The user retired harness performance qualification and baseline publication; these commands and their exclusive validation paths are removed.

**Remove the 69 internal Make wrappers while retaining their necessary operations.**

Simply hiding these names from help would leave the Makefile complexity intact. Their implementations should become private operations selected through the existing graph and typed command bindings. Genuine file-production rules can remain Make rules.

| Existing internal target(s) | Final disposition |
|---|---|
| `go-toolchain-readiness`, `bootstrap-tool-installations`, `codegen-toolchain`, `go-lint-toolchain`, `govulncheck-toolchain`, `gosec-toolchain`, `go-security-toolchain`, `shell-lint-toolchain`, `goose-toolchain`, `syft-toolchain`, `go-analysis-dependencies` | Private readiness operations shared by their consumers |
| `testservices-build`, `test-service-images`, `embedded-web-assets` | Private artifact producers; explicit image preparation through `bootstrap PART=test-images` |
| `services-wait`, `postgres-wait`, `object-store-wait` | Shared private lifecycle functions |
| `generate-artifacts` | Private producer behind `generate` |
| `migration-input-drift`, `test-catalog-check`, `semantic-identity-check` | Private `source-check` operations; share catalog context where applicable |
| `migration-scratch-apply` | Exact migration verification selection behind `migration-drift` |
| `deployable-shape` | Private CI/release validator |
| `backend-integration-support` | Support-class selection within `test-slice SUITE=backend-integration` |
| `run-harness-smoke-fast` | `harness-contract SUITE=smoke`; exactly three smoke roles |
| `run-harness-smoke-extended` | `harness-contract SUITE=regression` |
| `run-harness-smoke-execution`, `run-harness-smoke-lifecycle`, `run-harness-smoke-full` | Corresponding harness presets |
| `protocol-ts-browser-artifact-reachability` | Private artifact assertion with retained catalog ownership |
| `protocol-ts-dead-code-check` | Private operation behind `lint CHECK=protocol` |
| `harness-contract-tests` | Remove aggregate wrapper; retain its distinct full-tier leaf union |
| `harness-command-surface-contract`, `harness-evidence-contract`, `harness-graph-contract`, `harness-boundary-contract`, `harness-scheduler-contract`, `harness-local-session-contract`, `harness-playwright-adapter-contract`, `harness-observation-contract`, `harness-browser-preparation-contract`, `harness-benchmark-claim-contract` | Private canonical assurance operations, selected by owners/presets and existing tier policy |
| `format-go`, `format-frontend` | Private operations behind `format` |
| `browser-e2e-support` | `browser-e2e SUITE=support` |
| `browser-e2e-resettable` | Delete redundant stage union; preserve its canonical member selections |
| `lint-go`, `lint-go-format`, `lint-go-vet`, `lint-go-staticcheck` | One private union of the three checks behind `lint CHECK=go` |
| `check-frontend-install` | Remove empty readiness projection; consumers depend on canonical installed readiness |
| `check-harness-smoke` | Reporting projection over the three canonical smoke units |
| `release-evidence-contract` | Private release-evidence assertion |
| `license-report`, `sbom`, `release-inventory-artifacts` | One shared inventory producer, with separate required license/SBOM validators |
| `canonical-evidence-drift-suite` | Shared validator behind `harness-observability-check CHECK=evidence` |
| `frontend-artifact-consumer-check` | Exact artifact-consumer selection behind `harness-contract SUITE=artifact-consumers` |
| `harness-ui-review-contract`, `harness-ui-review-lifecycle`, `harness-ui-review-artifacts`, `harness-ui-review-presentation`, `harness-ui-review-execution`, `harness-ui-review-workflow`, `harness-ui-review-seeded-default`, `harness-ui-review-seeded-network-flow-claimed` | Private UI assurance operations; retain both seeded runtime profiles |
| `harness-runtime-contract`, `harness-fixture-contract`, `harness-runtime-boundary-contract` | Harness presets `runtime`, `fixtures`, `runtime-boundary`; preserve race-enabled and installed-runtime contexts |

Public command identity and private operation identity should be separate. Renaming a command must not create a new physical execution of otherwise identical work. Conversely, the same test name under different fixtures, race settings, runtime binaries or evidence classes must remain distinct.

### Complete baseline disposition inventory

This is a frozen human migration record, never a runtime alias or machine registry. Counts: 121 public + 37 internal helpers + 32 check-internal targets = 190. After the user amendment, 60 public names retire. Destination: 61 retained public names plus new `source-check` = 62.

| Baseline name | Baseline class | Destination / disposition | Closure slice | Status |
| --- | --- | --- | --- | --- |
| `help` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `help-all` | `public` | `help ALL=1` | W09 | DONE |
| `doctor` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `go-toolchain-readiness` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `bootstrap-tool-installations` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `bootstrap` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `bootstrap-node-runtime` | `public` | `bootstrap PART=node` | W09 | DONE |
| `frontend-toolchain` | `public` | `bootstrap PART=frontend`; toolchain-only preparation remains private | W09 | DONE |
| `frontend-install` | `public` | `bootstrap PART=frontend`; toolchain-only preparation remains private | W09 | DONE |
| `frontend-lockfile-update` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `playwright-install` | `public` | `bootstrap PART=browser` | W09 | DONE |
| `db-up` | `public` | `services-up` | W09 | DONE |
| `db-migrate` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `db-reset` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `postgres-baseline-reset` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `services-up` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `services-down` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `test-services-session-up` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `test-services-session-status` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `test-services-session-down` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `services-wait` | `internal_helper` | Shared private lifecycle functions | W09 | DONE |
| `postgres-wait` | `internal_helper` | Shared private lifecycle functions | W09 | DONE |
| `object-store-wait` | `internal_helper` | Shared private lifecycle functions | W09 | DONE |
| `object-store-init` | `public` | `services-up` | W09 | DONE |
| `object-store-reset` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `dev` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `browser-design-review` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `codegen-toolchain` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `go-lint-toolchain` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `govulncheck-toolchain` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `gosec-toolchain` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `go-security-toolchain` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `shell-lint-toolchain` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `generate` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `generate-artifacts` | `internal_helper` | Private producer behind `generate` | W09 | DONE |
| `generate-drift` | `public` | `source-check CHECK=generated`; retain both validations | W09 | DONE |
| `generated-artifact-policy-check` | `public` | `source-check CHECK=generated`; retain both validations | W09 | DONE |
| `json-shape-check` | `public` | `source-check CHECK=schema` | W09 | DONE |
| `openapi-compatibility-check` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `toolchain-drift` | `public` | `source-check CHECK=toolchain` | W09 | DONE |
| `migration-input-drift` | `internal_helper` | Private `source-check` operations; share catalog context where applicable | W09 | DONE |
| `migration-scratch-apply` | `internal_helper` | Exact migration verification selection behind `migration-drift` | W09 | DONE |
| `migration-drift` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `deployable-shape` | `internal_helper` | Private CI/release validator | W09 | DONE |
| `standup-package-smoke` | `public` | `package-inspect CHECK=startup` | W09 | DONE |
| `standup-reference-pack-smoke` | `public` | `package-inspect CHECK=reference-pack` | W09 | DONE |
| `standup-operational-recovery-smoke` | `public` | `package-inspect CHECK=recovery` | W09 | DONE |
| `test-catalog-check` | `check_internal` | Private `source-check` operations; share catalog context where applicable | W09 | DONE |
| `agent-finalize` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `test-evidence-audit` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `benchmark-claim-check` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `task-surface-report` | `public` | Discovery through `help ALL=1 JSON=1`; validation through `source-check CHECK=surface` | W09 | DONE |
| `task-guide` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `author-test-row-id` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `test-slice` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `service-backed-test-slice` | `public` | `test-slice OWNER=… SCOPE=services` | W09 | DONE |
| `backend-unit` | `public` | `test-slice SUITE=backend-unit` | W09 | DONE |
| `backend-store` | `public` | `test-slice SUITE=backend-store` | W09 | DONE |
| `backend-integration` | `public` | `test-slice SUITE=backend-integration` | W09 | DONE |
| `backend-integration-support` | `check_internal` | Support-class selection within `test-slice SUITE=backend-integration` | W09 | DONE |
| `backend-process` | `public` | `test-slice SUITE=backend-process` | W09 | DONE |
| `otel-conformance` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `target-plan` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `target-plan-json` | `public` | `target-plan … JSON=1` | W09 | DONE |
| `fixture-report` | `public` | `explain-run … DETAIL=fixtures` | W09 | DONE |
| `explain-run` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `harness-observability-check` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `harness-otel-export` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `harness-performance-check` | `public` | Retire harness performance qualification and its exclusive reader/gate implementation. | W09 | DONE |
| `harness-public-target-duration-baselines` | `public` | Retire harness baseline publication; retain historical evidence without translation. | W09 | DONE |
| `explain-test-owner` | `public` | `task-guide OWNER=… DETAIL=coverage` | W09 | DONE |
| `explain-target` | `public` | `target-plan … DETAIL=graph` or `DETAIL=artifacts` | W09 | DONE |
| `go-test-duration-baselines` | `public` | `explain-run … DETAIL=performance` | W09 | DONE |
| `go-test-duration-baseline-coverage` | `public` | Retire qualification-only roster gate; retain ordinary input/observability integrity | W09 | DONE |
| `go-test-duration-baseline-drift` | `public` | `explain-run … DETAIL=performance` | W09 | DONE |
| `browser-e2e-duration-baselines` | `public` | Same retained-run performance view | W09 | DONE |
| `browser-e2e-duration-baseline-drift` | `public` | Same retained-run performance view | W09 | DONE |
| `service-backed-make-target-duration-baselines` | `public` | Same retained-run performance view | W09 | DONE |
| `service-backed-make-target-duration-baseline-drift` | `public` | Same retained-run performance view | W09 | DONE |
| `harness-smoke-duration-baselines` | `public` | Same retained-run performance view | W09 | DONE |
| `harness-smoke-duration-baseline-drift` | `public` | Same retained-run performance view | W09 | DONE |
| `scheduler-event-order-drift` | `public` | `harness-observability-check … CHECK=events` | W09 | DONE |
| `scheduler-summary-timing-drift` | `public` | `harness-observability-check … CHECK=timing` | W09 | DONE |
| `testservices-build` | `internal_helper` | Private artifact producers; explicit image preparation through `bootstrap PART=test-images` | W09 | DONE |
| `test-service-images` | `internal_helper` | Private artifact producers; explicit image preparation through `bootstrap PART=test-images` | W09 | DONE |
| `run-harness-smoke-fast` | `internal_helper` | `harness-contract SUITE=smoke`; exactly three smoke roles | W09 | DONE |
| `run-harness-smoke-extended` | `internal_helper` | `harness-contract SUITE=regression` | W09 | DONE |
| `run-harness-smoke-execution` | `internal_helper` | Corresponding harness presets | W09 | DONE |
| `run-harness-smoke-lifecycle` | `internal_helper` | Corresponding harness presets | W09 | DONE |
| `run-harness-smoke-full` | `internal_helper` | Corresponding harness presets | W09 | DONE |
| `semantic-identity-check` | `check_internal` | Private `source-check` operations; share catalog context where applicable | W09 | DONE |
| `frontend-typecheck` | `public` | `lint CHECK=types` | W09 | DONE |
| `frontend-unit` | `public` | `test-slice SUITE=frontend-unit` | W09 | DONE |
| `frontend-import-boundary-check` | `public` | `lint CHECK=frontend-boundary` | W09 | DONE |
| `protocol-ts-browser-artifact-reachability` | `check_internal` | Private artifact assertion with retained catalog ownership | W09 | DONE |
| `protocol-ts-dead-code-check` | `check_internal` | Private operation behind `lint CHECK=protocol` | W09 | DONE |
| `backend-module-boundary-check` | `public` | `lint CHECK=backend-boundary` | W09 | DONE |
| `frontend-fallow-static` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `lint-biome` | `public` | `lint CHECK=frontend` | W09 | DONE |
| `lint-scripts` | `public` | `lint CHECK=scripts` | W09 | DONE |
| `lint-markdown` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `harness-contract-tests` | `check_internal` | Remove aggregate wrapper; retain its distinct full-tier leaf union | W09 | DONE |
| `harness-command-surface-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `harness-evidence-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `harness-contract` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `lint-shell` | `public` | `lint CHECK=shell` | W09 | DONE |
| `format` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `format-go` | `internal_helper` | Private operations behind `format` | W09 | DONE |
| `format-frontend` | `internal_helper` | Private operations behind `format` | W09 | DONE |
| `browser-e2e` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `browser-e2e-webserver-backed` | `public` | `browser-e2e SUITE=functional` | W09 | DONE |
| `browser-e2e-support` | `internal_helper` | `browser-e2e SUITE=support` | W09 | DONE |
| `browser-e2e-stateful` | `public` | `browser-e2e SUITE=stateful` | W09 | DONE |
| `browser-e2e-resettable` | `internal_helper` | Delete redundant stage union; preserve its canonical member selections | W09 | DONE |
| `browser-e2e-measurement` | `public` | `browser-e2e SUITE=measurement` | W09 | DONE |
| `browser-e2e-a11y` | `public` | `browser-e2e SUITE=accessibility` | W09 | DONE |
| `browser-e2e-visual` | `public` | `browser-e2e SUITE=visual` | W09 | DONE |
| `browser-e2e-visual-update` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `test-fast` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `test` | `public` | `ci` for comprehensive validation; focused selectors for narrower work | W09 | DONE |
| `lint` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `lint-go` | `internal_helper` | One private union of the three checks behind `lint CHECK=go` | W09 | DONE |
| `lint-go-format` | `internal_helper` | One private union of the three checks behind `lint CHECK=go` | W09 | DONE |
| `lint-go-vet` | `internal_helper` | One private union of the three checks behind `lint CHECK=go` | W09 | DONE |
| `lint-go-staticcheck` | `internal_helper` | One private union of the three checks behind `lint CHECK=go` | W09 | DONE |
| `go-vulncheck` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `go-gosec-targeted` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `go-gosec-audit` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `check` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `check-frontend-install` | `check_internal` | Remove empty readiness projection; consumers depend on canonical installed readiness | W09 | DONE |
| `check-harness-smoke` | `check_internal` | Reporting projection over the three canonical smoke units | W09 | DONE |
| `ci` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `release-check` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `release-readiness-evidence` | `public` | Private release-evidence assembly/validation inside `release-check`; retained results remain inspectable | W09 | DONE |
| `release-evidence-contract` | `check_internal` | Private release-evidence assertion | W09 | DONE |
| `license-report` | `internal_helper` | One shared inventory producer, with separate required license/SBOM validators | W09 | DONE |
| `sbom` | `internal_helper` | One shared inventory producer, with separate required license/SBOM validators | W09 | DONE |
| `seaweedfs-compatibility` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `seaweedfs-release-evidence` | `public` | Private reporting phase of `seaweedfs-release-gate`, including retained failure evidence | W09 | DONE |
| `seaweedfs-release-gate` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `build` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `build-server` | `public` | `build COMPONENT=server` | W09 | DONE |
| `build-server-harness` | `public` | `build COMPONENT=server-harness` | W09 | DONE |
| `embedded-web-assets` | `internal_helper` | Private embedded-web producer consumed by server, harness-server and production builds | W09 | DONE |
| `build-migrate` | `public` | `build COMPONENT=migrate` | W09 | DONE |
| `build-operator` | `public` | `build COMPONENT=operator` | W09 | DONE |
| `build-web` | `public` | `build COMPONENT=web` | W09 | DONE |
| `build-web-measurement` | `public` | `build COMPONENT=web-measurement` | W09 | DONE |
| `clean` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `distclean` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `canonical-evidence-drift-suite` | `check_internal` | Shared validator behind `harness-observability-check CHECK=evidence` | W09 | DONE |
| `goose-toolchain` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `release-inventory-artifacts` | `internal_helper` | One shared inventory producer, with separate required license/SBOM validators | W09 | DONE |
| `frontend-artifact-consumer-check` | `check_internal` | Exact artifact-consumer selection behind `harness-contract SUITE=artifact-consumers` | W09 | DONE |
| `browser-design-review-smoke` | `public` | Private lifecycle assurance reachable through owner selection and `harness-contract SUITE=lifecycle` | W09 | DONE |
| `ui-review` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `ui-review-status` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `ui-browser` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `ui-capture` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `ui-analyze` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `ui-review-report` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `ui-review-stop` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `harness-ui-review-contract` | `check_internal` | Private UI assurance operations; retain both seeded runtime profiles | W09 | DONE |
| `harness-ui-review-lifecycle` | `check_internal` | Private UI assurance operations; retain both seeded runtime profiles | W09 | DONE |
| `harness-ui-review-artifacts` | `check_internal` | Private UI assurance operations; retain both seeded runtime profiles | W09 | DONE |
| `harness-ui-review-presentation` | `check_internal` | Private UI assurance operations; retain both seeded runtime profiles | W09 | DONE |
| `harness-ui-review-execution` | `check_internal` | Private UI assurance operations; retain both seeded runtime profiles | W09 | DONE |
| `harness-ui-review-workflow` | `check_internal` | Private UI assurance operations; retain both seeded runtime profiles | W09 | DONE |
| `harness-ui-review-seeded-default` | `check_internal` | Private UI assurance operations; retain both seeded runtime profiles | W09 | DONE |
| `harness-ui-review-seeded-network-flow-claimed` | `check_internal` | Private UI assurance operations; retain both seeded runtime profiles | W09 | DONE |
| `cryptographic-policy-assessment` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `credential-capacity-assessment` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `package-release` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `package-inspect` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `syft-toolchain` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `test-run-status` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |
| `go-analysis-dependencies` | `internal_helper` | Private readiness operations shared by their consumers | W09 | DONE |
| `harness-runtime-contract` | `check_internal` | `harness-contract SUITE=runtime`; preserve its distinct runtime context | W09 | DONE |
| `harness-fixture-contract` | `check_internal` | `harness-contract SUITE=fixtures`; preserve its distinct runtime context | W09 | DONE |
| `harness-runtime-boundary-contract` | `check_internal` | `harness-contract SUITE=runtime-boundary`; preserve its distinct runtime context | W09 | DONE |
| `harness-graph-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `harness-boundary-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `harness-scheduler-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `harness-local-session-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `harness-playwright-adapter-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `harness-observation-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `harness-browser-preparation-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `harness-benchmark-claim-contract` | `check_internal` | Private canonical assurance operations, selected by owners/presets and existing tier policy | W09 | DONE |
| `build-report-renderer` | `public` | `build COMPONENT=report-renderer` | W09 | DONE |
| `test-run-recover` | `public` | Retain public command; apply adopted selectors and policy where specified. | W09 | DONE |

## 5. Coupling and Boundary Findings

| Finding | Gap / risk if unresolved | Contract | Slice / validation | Priority |
| --- | --- | --- | --- | --- |
| F-000 | Establish controlling tracker and baseline: Lost requirements, stale assumptions, missing evidence. | C-000 | W00 / V-000 | must_fix |
| F-001 | Reconcile and adopt specification changes: Conflicting authorities and unsafe interpretation. | C-001 | W01 / V-001 | must_fix |
| F-002 | Typed bindings and canonical private execution: Deleted wrappers break execution or aliases preserve coupling. | C-002 | W02 / V-002 | must_fix |
| F-003 | Exact rows and explicit policy dependencies: Accidental execution expansion or weakened security assurance. | C-003 | W03 / V-003 | must_fix |
| F-004 | Harness tiers and canonical smoke execution: Slow local checks, duplicate assertions, unsafe deduplication. | C-004 | W04 / V-004 | must_fix |
| F-005 | Shared validation and finalizer transactions: Repeated cost or stale validation bypass. | C-005 | W05 / V-005 | must_fix |
| F-006 | Aggregate policy and consumer preparation: Ambiguous guarantees, excess local work, omitted security/artifacts. | C-006 | W06 / V-006 | must_fix |
| F-007 | Consolidated commands and lifecycle: Hidden deleted-target callers, surprising defaults, incorrect completeness claims. | C-007 | W07 / V-007 | must_fix |
| F-008 | Selection-aware evidence and diagnostics: Lost workloads, timing double counting, incompatible evidence accepted. | C-008 | W08 / V-008 | must_fix |
| F-009 | Regeneration, callers and hard cutover: Mixed surface, stale scripts, missing assertions, generated drift. | C-009 | W09 / V-009 | must_fix |
| F-010 | Behavior and release validation: Unsubstantiated acceptance or hidden regression. | C-010 | W10 / V-010 | must_fix |
| F-013 | Native host binding can race a short child exit between liveness checks; resource-conflict stage is not retained. Historical E-074 cause is unproven. | C-002, C-008 | W09 / V-002, V-008 | must_fix |
| F-015 | Concurrent renderer assertion observes no finished Mermaid SVG at PDF capture; readiness cause requires characterization. | Reporting renderer contract; C-010 | W09 / affected owner row, V-010 | must_fix |
| F-014 | Dependency-blocked selected smoke rows have no terminal row artifact, violating REQ-364/367 and preventing exact-run diagnosis. | C-008 | W09 / V-008 | must_fix |
| F-012 | Operation identity omits executable evidence-publication fields: changed success-summary behavior can retain the same dispatch digest and graph identity. | C-002, C-008 | W09 / V-002, V-008 | must_fix |
| F-016 | Race runtime preset exposes inherited public selectors in nested graph preparation; canonical unit environment boundary requires characterization. | C-002, C-007 | W09 / V-002, V-007 | must_fix |
| F-017 | Private-capture cancellation fixture signals child readiness before native process registration completes, so it may exercise acquisition failure instead of its asserted post-acquisition path. | C-008, C-010 | W09 / native producer characterization | must_fix |
| F-018 | Fast gate passes every unit but runtime_close rejects an unanchored process group; private proof remains fenced. | C-010; REQ-853 | W09 / physical settlement and concurrent assurance | must_fix |
| F-019 | Full graph-cutover fixtures omit row bindings and reject before their intended resource-claim boundary. | C-008, C-009 | W09 / graph-cutover negative fixtures | must_fix |
| F-020 | Concurrent installed HTTPS boundary passes its row but fails browser fixture release; backend settlement acknowledgement is missing and physical failure detail is discarded. | C-007, C-011 | W09 / native browser cleanup and diagnostic propagation | must_fix |
| F-021 | Intentional unsafe-browser cleanup fixture rejects removal, then recovery preservation returns an undifferentiated missing-file error; a later focused pass does not establish cause. | C-007, C-011 | W09 / R9 recovery-proof and disposable-purge boundaries | must_fix |
| F-022 | PID-reuse lifecycle fixture exits its leader after 200 ms while bounded rejection takes up to six seconds, so its no-signal assertion observes its own timer. | C-007, C-010; REQ-853 | W10-R1 / exact UI lifecycle row | must_fix |
| F-023 | CI clipboard-fidelity assertion observes a persisted native editor value after Escape; cause and relation to harness changes remain unproven. | Timeline workbook cancellation; C-010 | W10-R2 / clipboard-fidelity boundary | must_fix |
| F-024 | Browser artifact-lifetime helper still invokes retired frontend Make names through a conditional argument array, missed by the W09 caller review. | C-009, C-010 | W10-R3 / independent frontend publication | must_fix |
| F-025 | CI browser assertion passes but process signaling fails; reproduced reaped-leader pidfd EINVAL boundary, original CI exact syscall remains unproven. | C-007, C-010, C-011 | W10-R4 / browser settlement and final-candidate cleanup | must_fix |
| F-026 | Logical-check reporting still duplicates executable argv and advertises a retired private Make command. | C-001, C-002, C-009 | W10-R5 / canonical reporting metadata | must_fix |
| F-027 | Overlapping full assurance and CI exhausted host admission for the lint-output assertion before launch. | C-010, C-011 | W10-R6 / sequential qualification | must_fix |
| F-028 | Sequential full assurance reports two machine-check PostgreSQL forced-drop deadline failures; the physical wait cause is not retained. | C-007, C-010, C-011 | W10-R6a / shared diagnostics and complete qualification; original cause unproven | resolved |
| F-029 | Current-source CI browser group cannot connect to its local renderer endpoint before test launch; group artifacts are absent. | C-007, C-010, C-011 | W10-R7 / renderer lifecycle investigation and correction | must_fix |
| F-030 | Renewed full assurance exposes a forced-drop timeout with IPC/ProcSignalBarrier while machine check drains; delayed backend is unidentified. | C-007, C-010, C-011 | W10-R8 / exact PostgreSQL barrier diagnosis and correction | must_fix |
| F-031 | Retained reconstruction substitutes physical host capacity for inherited Go worker budget, changing exact graph identity. | C-002, C-008 | W10-R9 / exact retained reconstruction and budget rejection | must_fix |
| F-032 | Fragmented Go output can corrupt otherwise valid typed failure envelopes. | C-008 | W10-R8a / bounded per-test framing and failure precedence | must_fix |
| F-033 | Per-clone explicit checkpoints impose unnecessary whole-cluster synchronization; full assurance records checkpoint I/O cleanup timeouts. | C-007, C-010 | W10-R8c / durable WAL publication without extra checkpoint | must_fix |
| F-034 | Extension concurrency test starts its admission budget before fixture preparation. | C-010; Extension Table 9-E | W10-R10 / operation-scoped timing and cancellation-safe coordination | must_fix |
| F-011 | Validation and handoff completion: Incomplete migration misrepresented as complete. | C-011 | W11 / V-011 | must_fix |

## 6. Refactor Workstreams

The following prerequisite column is authoritative. No next workstream begins until the prior exit criteria pass and its tracker checkpoint is saved. Implementation slices are separate workstreams; one coordinated public cutover occurs at W09.

| Workstream | Name | Prerequisite | Exit criterion |
| --- | --- | --- | --- |
| W00 | Establish controlling tracker and baseline | none | Every gap reaches a workstream and validation obligation; baseline identity and evidence state are recorded. |
| W01 | Reconcile and adopt specification changes | W00 + tracker checkpoint | Owner clauses agree; exact 62-name surface and variants are specified; Markdown lint passes. |
| W02 | Typed bindings and canonical private execution | W01 + tracker checkpoint | Equivalent selections produce the same units; unknown operations fail before setup; no retired Make dependency. |
| W03 | Exact rows and explicit policy dependencies | W02 + tracker checkpoint | Full-tier measurement selects current four rows without 29 release rows; direct selection is complete; required OTel dependencies survive. |
| W04 | Harness tiers and canonical smoke execution | W03 + tracker checkpoint | Exactly three check smoke roles; broad regressions reachable in CI/release/direct presets; distinct contexts remain distinct. |
| W05 | Shared validation and finalizer transactions | W04 + tracker checkpoint | Reuse only unchanged inputs; injected failures preserve taxonomy and restore tracked state including prior edits. |
| W06 | Aggregate policy and consumer preparation | W05 + tracker checkpoint | Exact gate closures; required producers present; blocking gosec independent from advisory audit. |
| W07 | Consolidated commands and lifecycle | W06 + tracker checkpoint | All valid/invalid selectors covered; browser SUITE required; discovery has no preparation; package scenarios share identity. |
| W08 | Selection-aware evidence and diagnostics | W07 + tracker checkpoint | No selection collisions, false full-validation claims or mutated retained evidence; malformed/stale evidence rejects. |
| W09 | Regeneration, callers and hard cutover | W08 + tracker checkpoint | Exactly 62 public names; no active predecessor paths; source checks pass; finalizer succeeds before broad verification. |
| W10 | Behavior and release validation | W09 + tracker checkpoint | Required functional, security, artifact and release gates pass; no harness performance qualification or baseline publication is required. |
| W11 | Validation and handoff completion | W10 + tracker checkpoint | All required evidence closes on final source; all workstreams checkpointed; no unresolved required acceptance item. |

The W09 follow-up repairs below are separate implementation slices. Complete and checkpoint each before starting the next. They do not bypass W09 or authorize W10 early.

| Repair slice | Finding / scope | Prerequisite | Status | Exit criterion |
| --- | --- | --- | --- | --- |
| W09-R1 | F-012 complete operation execution identity | E-072 outer exit collected | DONE | Characterized publication-field changes invalidate identity; command spelling remains separate; graph/dispatch contracts pass. |
| W09-R2 | F-013 short-child host binding and safe failure stage | W09-R1 checkpoint | DONE | Deterministic terminal-child transition, exact proof and capacity guards, process-group cleanup and native regressions pass. |
| W09-R3 | F-014 unavailable-row terminal publication | W09-R2 checkpoint | DONE | Validated row bindings; dependency/cancellation records; no overwritten or fabricated passes; reader and native contracts pass. |
| W09-R4 | F-015 renderer asynchronous readiness | W09-R3 checkpoint | DONE | Delayed readiness is characterized; required SVG exists at PDF capture; exact renderer row and integration validation pass without sleep/retry/timeout relaxation. |
| W09-R5 | F-016 typed public-input isolation at private execution | E-093 outer exit collected | DONE | Selection controls cannot leak across canonical operations; exact nested selection and strict public admission remain validated. |
| W09-R6 | F-017 deterministic private-capture cancellation boundary | W09-R5 checkpoint | DONE | Cancellation reaches its declared ownership phase, preserves private-output confinement, child reaping and capture cleanup; acquisition diagnostics retain their own assertions. |
| W09-R7 | F-018 exact process-group settlement at runtime close | W09-R6 checkpoint | DONE | Characterized ownership/quiescence boundary and runtime regressions pass without unrelated signals, hidden retries, deadline changes or deleted proof; full concurrent acceptance remains a W09 exit requirement. |
| W09-R8 | F-019 graph-cutover fixture fidelity | W09-R7 checkpoint | DONE | Synthetic unit evidence is internally valid; negative assertions reach their intended boundary; both fast/matrix assurance paths pass. |
| W09-R9 | F-020/F-021 installed browser cleanup, recovery preservation and durable closed diagnostics | E-114 full owning exit collected | DONE | Identify/characterize the failing physical boundary, retain bounded safe diagnostics, preserve exact ownership and existing cleanup deadlines, and pass applicable focused/concurrent evidence. |

The W10 qualification repairs are separate slices, sequenced after the active CI owning exit. Each completion requires its own tracker checkpoint before its successor.

| Repair slice | Finding / scope | Prerequisite | Status | Exit criterion |
| --- | --- | --- | --- | --- |
| W10-R1 | F-022 ownership fixture lifetime and exact teardown | E-131 CI owning exit collected | DONE | Explicit readiness/release replaces timer; PID-reuse and leaderless no-signal assertions and the complete lifecycle row pass. |
| W10-R2 | F-023 native-paste settlement and cancellation characterization | W10-R1 checkpoint | DONE | Await required paste autosave, characterize separate uncommitted Escape, and retain malformed-representation no-write assertions; exact clipboard row passes. |
| W10-R3 | F-024 independent frontend build caller migration | W10-R2 checkpoint | DONE | Both production and measurement calls use typed build selectors; programmatic Make callers reviewed; exact artifact-lifetime row passes without weakening immutable assets. |
| W10-R4 | F-025 CI browser fixture release | W10-R3 checkpoint | DONE | Diagnose exact physical boundary from retained safe diagnostics/private authority; characterize and repair without weakened ownership/deadlines or hidden retries; focused and final gate cleanup pass. |
| W10-R5 | F-026 logical-check reporting metadata | W10-R4 checkpoint | DONE | Remove duplicate executable argv from logical reporting inventory, reject its reintroduction, regenerate and validate discovery; canonical operations remain sole execution owner. |
| W10-R6 | F-027 qualification admission; F-028 discovered cleanup follow-up | W10-R5 checkpoint and active CI drain | DONE | E-164: instrumented full assurance passed 58/58 with all cleanup after R6a; human/machine check 800/800 each. Historical PostgreSQL timeout causes remain unproven. |
| W10-R6a | F-028 shared PostgreSQL cleanup diagnostics | E-159 owning exit and checkpoint | DONE | E-161: exact failed stage retained; shared physical observer, redaction, adapter/schema rejection and focused rows pass. Historical wait cause remains unproven; complete qualification is still required under R6. |
| W10-R7 | F-029 renderer startup diagnostic/classification boundary | E-167 CI owning exit and checkpoint | DONE | E-170: red/green, exact nine-row context and all-source checks pass; finalizer changed zero files. Original refusal cause is unproven; renewed complete acceptance remains W10-owned. |
| W10-R8 | F-030/F-033 PostgreSQL cleanup under concurrent assurance | E-173 full owning exit and checkpoint; latest E-199 | IN_PROGRESS | Earlier barrier observations and current checkpoint I/O waits remain distinct. Removing the extra clone checkpoint did not close full-load cleanup; diagnose retained failures without relaxing ownership/budgets and renew affected acceptance. Stopped at user request. |
| W10-R8a | F-032 fragmented Go failure envelopes | E-187 owning exit | DONE | Exact logical envelopes survive event fragmentation; incomplete/oversized or cross-test content rejects without changing valid failure precedence. |
| W10-R8b | F-030 create/drop barrier interference | R8a checkpoint | DONE | Template copy durability and failed-acquisition ownership remain explicit; cleanup barriers complete under existing bounds. |
| W10-R9 | F-031 retained worker-budget reconstruction | R8b implementation checkpoint; common broad qualification follows both | DONE | E-191–E-193: observed mismatch repaired; exact graph, malformed-budget, retained-reader and native admission checks pass. |
| W10-R8c | F-033 durable clone publication without redundant checkpoints | E-195 owning exit and checkpoint | DONE | Prove durable WAL creation without checkpoint privilege; preserve failure ownership, isolation and bounded cleanup; renew concurrent acceptance. |
| W10-R10 | F-034 extension admission test timing | R8c checkpoint | DONE | Start admission timing after fixture readiness, preserve existing operation limits and cancellation-safe synchronization; exact row passes. |

## 7. Proposed Refactor Slice Plan

### S-000 / W00 — Establish controlling tracker and baseline

- Areas: Documentation, evidence.
- Remediation: Stable IDs; complete command dispositions; content/index baseline; record pre-change correctness evidence without a performance-window prerequisite.
- Acceptance / long-term outcome: Every gap reaches a workstream and validation obligation; baseline identity and evidence state are recorded.
- Migration / rollback: Tracker only; preserve original mappings and rationale. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Lost requirements, stale assumptions, missing evidence.
- Traceability: F-000 → C-000 → V-000 → completion evidence recorded at checkpoint.
- Entry/readiness: DONE: baseline content/index evidence, complete dispositions and current correctness run recorded; user removed the timing prerequisite.

### S-001 / W01 — Reconcile and adopt specification changes

- Areas: Specification, documentation.
- Remediation: Amend aggregate/surface/selector/operation/evidence contracts and remove REQ-064 baseline mutation conflict, retire REQ-377/380/848/850 qualification contracts.
- Acceptance / long-term outcome: Owner clauses agree; exact 62-name surface and variants are specified; Markdown lint passes.
- Migration / rollback: Breaking command and evidence-contract changes; no aliases. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Conflicting authorities and unsafe interpretation.
- Traceability: F-001 → C-001 → V-001 → completion evidence recorded at checkpoint.
- Entry/readiness: W00 completed and checkpointed; inspect exact edit sources and characterize risky movement before editing.

### S-002 / W02 — Typed bindings and canonical private execution

- Areas: Implementation, contracts, schemas, tests.
- Remediation: Extend task-surface bindings and work-graph operations; migrate private dispatch, preparation, caches and artifact resolution.
- Acceptance / long-term outcome: Equivalent selections produce the same units; unknown operations fail before setup; no retired Make dependency.
- Migration / rollback: Version affected contracts; new public surface stays inactive until W09. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Deleted wrappers break execution or aliases preserve coupling.
- Traceability: F-002 → C-002 → V-002 → completion evidence recorded at checkpoint.
- Entry/readiness: W01 completed and checkpointed; inspect exact edit sources and characterize risky movement before editing.

### S-003 / W03 — Exact rows and explicit policy dependencies

- Areas: Implementation, contracts, tests.
- Remediation: Carry selected browser IDs through batching/runners/evidence; replace owner-wide policy expansion with declared row dependencies shared with OTel validation.
- Acceptance / long-term outcome: Full-tier measurement selects current four rows without 29 release rows; direct selection is complete; required OTel dependencies survive.
- Migration / rollback: Reduced accidental browser scope; justified higher-tier telemetry evidence remains. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Accidental execution expansion or weakened security assurance.
- Traceability: F-003 → C-003 → V-003 → completion evidence recorded at checkpoint.
- Entry/readiness: W02 completed and checkpointed; inspect exact edit sources and characterize risky movement before editing.

### S-004 / W04 — Harness tiers and canonical smoke execution

- Areas: Catalogs, contracts, implementation, tests.
- Remediation: Split narrow scheduler/fixture smoke assertions from full-tier regressions; share equivalent executions across projections.
- Acceptance / long-term outcome: Exactly three check smoke roles; broad regressions reachable in CI/release/direct presets; distinct contexts remain distinct.
- Migration / rollback: Record catalog-row successors; retain race, installed-runtime and seeded-profile contexts. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Slow local checks, duplicate assertions, unsafe deduplication.
- Traceability: F-004 → C-004 → V-004 → completion evidence recorded at checkpoint.
- Entry/readiness: W03 completed and checkpointed; inspect exact edit sources and characterize risky movement before editing.

### S-005 / W05 — Shared validation and finalizer transactions

- Areas: Implementation, contracts, tests.
- Remediation: Invocation-scoped immutable validated context; explicit invalidation after source mutation; independent action receipts and frozen pre-mutation identity.
- Acceptance / long-term outcome: Reuse only unchanged inputs; injected failures preserve taxonomy and restore tracked state including prior edits.
- Migration / rollback: No expansion of finalizer maintenance scope or baseline publication. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Repeated cost or stale validation bypass.
- Traceability: F-005 → C-005 → V-005 → completion evidence recorded at checkpoint.
- Entry/readiness: W04 completed and checkpointed; inspect exact edit sources and characterize risky movement before editing.

### S-006 / W06 — Aggregate policy and consumer preparation

- Areas: Implementation, projections, tests.
- Remediation: Four gates; consumer-driven check builds; move advisory Fallow reports and production qualification to CI; retain local blocking integrity/security.
- Acceptance / long-term outcome: Exact gate closures; required producers present; blocking gosec independent from advisory audit.
- Migration / rollback: ci replaces test; check no longer promises unconditional production qualification. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Ambiguous guarantees, excess local work, omitted security/artifacts.
- Traceability: F-006 → C-006 → V-006 → completion evidence recorded at checkpoint.
- Entry/readiness: W05 completed and checkpointed; inspect exact edit sources and characterize risky movement before editing.

### S-007 / W07 — Consolidated commands and lifecycle

- Areas: Interfaces, implementation, schemas, tests, documentation.
- Remediation: Implement selectors/source-check; idempotent single service startup/readiness; preserve destructive, interactive, export and publication boundaries.
- Acceptance / long-term outcome: All valid/invalid selectors covered; browser SUITE required; discovery has no preparation; package scenarios share identity.
- Migration / rollback: Apply 60 public and 69 private dispositions; retain support evidence/service completion scope. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Hidden deleted-target callers, surprising defaults, incorrect completeness claims.
- Traceability: F-007 → C-007 → V-007 → completion evidence recorded at checkpoint.
- Entry/readiness: W06 completed and checkpointed; inspect exact edit sources and characterize risky movement before editing.

### S-008 / W08 — Selection-aware evidence and diagnostics

- Areas: Specification, schemas, implementation, tests.
- Remediation: Bind command, normalized selectors, exact selection and operation identities; retain descriptive duration/resource views; identify the complete qualification-only removal closure for the coordinated W09 cutover.
- Acceptance / long-term outcome: No selection collisions, false full-validation claims or mutated retained evidence; malformed/stale evidence rejects.
- Migration / rollback: Version changed wire contracts; historical identity retained without translation; no baseline/performance acceptance bindings. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Lost workloads, timing double counting, incompatible evidence accepted.
- Traceability: F-008 → C-008 → V-008 → completion evidence recorded at checkpoint.
- Entry/readiness: W07 completed and E-017 saved; owner guidance obtained; evidence and receipt boundaries characterized.

### S-009 / W09 — Regeneration, callers and hard cutover

- Areas: Projections, generators, tests, specifications, guides, skills.
- Remediation: Regenerate through Make; atomically replace active public registry; remove predecessors/wrappers and the qualification-only schemas, profiles, readers and publishers identified in W08; update active callers and recommendations.
- Acceptance / long-term outcome: Exactly 62 public names; no active predecessor paths; source checks pass; finalizer succeeds before broad verification.
- Migration / rollback: One coordinated merge/cutover; historical references are not compatibility contracts. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Mixed surface, stale scripts, missing assertions, generated drift.
- Traceability: F-009 → C-009 → V-009 → completion evidence recorded at checkpoint.
- Entry/readiness: W08 completed and E-019 checkpoint saved; active cutover validation in progress.

### S-010 / W10 — Behavior and release validation

- Areas: Validation, evidence, documentation.
- Remediation: Final-candidate focused/gate/release validation; descriptive timing may be reported without a performance acceptance claim.
- Acceptance / long-term outcome: Required functional, security, artifact and release gates pass; no harness performance qualification or baseline publication is required.
- Migration / rollback: Never relabel old observations; selector scope and package identity remain explicit. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Unsubstantiated acceptance or hidden regression.
- Traceability: F-010 → C-010 → V-010 → completion evidence recorded at checkpoint.
- Entry/readiness: W09 completed at E-130; W10 R1–R5 repaired and checkpointed at E-150. Renew complete final-candidate acceptance after finalization.

### S-011 / W11 — Validation and handoff completion

- Areas: Validation, documentation.
- Remediation: Reconcile every finding/evidence obligation; complete missing validation; final diff review and durable tracker handoff.
- Acceptance / long-term outcome: All required evidence closes on final source; all workstreams checkpointed; no unresolved required acceptance item.
- Migration / rollback: Document replacements/defaults/cache-evidence consequences; coordinated rollback only. Revert the coordinated authored/projection slice if validation fails; never restore aliases as a substitute.
- Leaving unresolved: Incomplete migration misrepresented as complete.
- Traceability: F-011 → C-011 → V-011 → completion evidence recorded at checkpoint.
- Entry/readiness: blocked pending W10 completion and checkpoint; inspect exact edit sources and characterize risky movement before editing.

## 8. Validation Plan

Run repository work only through public Make targets. Before cutover use `make task-guide ROLE=module-author OWNER=<owner>`; after cutover use `make task-guide OWNER=<owner>`. Owners: harness.command_surface, harness.browser, harness.test_catalog, harness.generated_artifacts, harness.evidence_accounting, harness.release, platform.telemetry.

| Validation | Stage | Commands / scenarios | Acceptance |
| --- | --- | --- | --- |
| V-000 | Baseline | `make help-all`; `make doctor`; pre-change correctness through public Make commands | Pinned environment ready; baseline identity and pre-change correctness outcome recorded |
| V-001 | Specification | `make lint-markdown` | Adopted clauses consistent; no product verification for documentation-only slice |
| V-002–V-008 | Per implementation slice | Task-guide-selected exact owner rows; affected contract presets | Exact selections; fail-before-setup inputs; context-aware deduplication; lifecycle/cleanup/cancellation; freshness and rollback; variant evidence |
| V-009 | Cutover | `make generate`; new `make source-check CHECK=all`; `make agent-finalize` | Generated parity; 62 names; no old dispatch; finalizer changes reviewed |
| V-010 | Qualification | Final-candidate `make check`, `make ci`, `make release-check`; exact retained-run evidence checks | Full functional/release evidence; no harness timing-qualification requirement |
| V-011 | Handoff | Exact owner evidence audits and `make harness-observability-check ... CHECK=all`; final diff/preservation review | All required evidence applies to final source; no partial success represented as completion |

W00 completion checkpoint (2026-10-17): user removed harness performance qualification and baseline publication. Final public surface becomes 62 commands. Product performance tests, diagnostic timings, release/security validation and Core 05 claim controls retain their separate owners. No production source has changed; the initial check continues as correctness evidence.

Each launch records exact command, fresh run ID, native session, result parent/root, outer exit, summary result and cleanup. Cache hits are not fresh timing evidence. Diagnose before rerunning or broadening. Negative fixtures must reach their intended boundary even with warm caches. Assertions must verify behavior rather than mirror implementation.

Final sequence: Markdown lint; all source checks; required focused suites; agent-finalize; check/ci/complete release; exact-run evidence checks; final handoff. Harness performance qualification and baseline publication are removed by user request. Supply finalizer RESULTS_DIR only for an eligible successful full warm check. Otherwise record retained-run maintenance skipped because RESULTS_DIR was unset. Finalizer source changes invalidate affected evidence.

Qualification remains WSL2 on the pinned environment. Separate scope-reduction savings from unchanged-workload improvements. Do not claim formal CMVP applicability, deployment approval, or Core 05 publication from an engineering pass.

## 9. Top-Level Work Tracker

| Work item | Status | Evidence / current state |
| --- | --- | --- |
| W00 — Establish controlling tracker and baseline | DONE | E-000/E-001 and complete 190-command inventory; user removed reference performance prerequisite. |
| W01 — Reconcile and adopt specification changes | DONE | E-004: user-authorized normative changes, reference review and final Markdown lint 0. No implementation pass inferred. |
| W02 — Typed bindings and canonical private execution | DONE | E-005/E-006: canonical dispatch, identity/cache/artifact migrations, generated checks and focused contracts pass; public cutover remains W09. |
| W03 — Exact rows and explicit policy dependencies | DONE | E-007/E-008: exact tier selections, explicit telemetry dependencies, invalid scope rejection, adapter and conformance evidence pass. |
| W04 — Harness tiers and canonical smoke execution | DONE | E-009/E-010: exact smoke rows, full-tier regressions, canonical projections, 23 graph contracts and both real smoke rows pass. |
| W05 — Shared validation and finalizer transactions | DONE | E-011/E-012: immutable fresh facts, separate receipts, transaction rollback/content reporting, evidence/boundary contracts and real finalizer pass. |
| W06 — Aggregate policy and consumer preparation | DONE | E-013: consumer-driven check, explicit CI production/advisory policy, local configuration integrity and closure contracts pass. |
| W07 — Consolidated commands and lifecycle | DONE | E-014–E-017: private selectors, admission/discovery APIs, lifecycle and complete contract validation; public activation remains W09. |
| W08 — Selection-aware evidence and diagnostics | DONE | E-019: exact evidence identity, scoped checks, receipt repair and full retained validation; retirement closure assigned to coordinated W09 cutover. |
| W09 — Regeneration, callers and hard cutover | DONE | E-130: all 190 dispositions implemented; 62-name surface and generated/source/default contracts pass; R1–R9 checkpointed; finalizer zero mutations; full assurance outer 0, 58/58 and all cleanup complete. Historical causal limits retained. |
| W10 — Behavior and release validation | IN_PROGRESS | E-199: stopped at user request after full assurance outer 2 / normalized 10, 57/58. Human check passed 800/800; machine check failed 793/800 with seven PostgreSQL cleanup-error units. All graph cleanup records completed. R8 remains open; final-source successful full assurance, CI, complete release and exact audits remain required. |
| W11 — Validation and handoff completion | TODO | E-199/E-200: durable stopping-point handoff refreshed and Markdown lint passed, with exact failures, source identity, cleanup distinction and restart order. Final reconciliation, preservation/diff review and completion remain gated on W10; this checkpoint does not close W11. |

## 10. Session Handoff Log

### 2026-10-17 — execution bootstrap

- Scope/authority: user explicitly authorized W00–W11 implementation. Existing branch/commit and clean tracked worktree revalidated.
- Backend/frontend boundaries: no product implementation changed; catalog/context invariants preserved as future acceptance conditions.
- Contract/codegen: no authored executable contract or generated artifact changed.
- Tests/harness: `make help-all` passed; `make task-guide ROLE=module-author OWNER=harness.command_surface` passed; `CARTULARY_TEST_RUN_ID=remediation-w00-doctor-20261017-01 make doctor` passed (outer exit 0).
- Security/authorization: no telemetry export, destructive reset, baseline publication or external message sent.
- Active reference command: `CARTULARY_TEST_RUN_ID=remediation-w00-reference-check-cold-20261017-01 CARTULARY_HARNESS_CACHE_MODE=cold make check`; native session 87188; root `.cartulary/test-results/remediation-w00-reference-check-cold-20261017-01`; result pending.
- A historical inventory found no graph manifest bound to the current commit and no performance evidence-roots manifest suitable for this cutover. Historical passes have not been accepted as current evidence.
- Additional inspection found three older-commit graph roots with the identical executable-source digest. The release root `remediation-final-release-20261017-04` failed with `artifact/artifact_error` and three failed cleanup steps; the two build roots passed but do not constitute qualified windows. Current-schema clean windows for other historical source states are not accepted without compatibility validation. The retained reference/comparison portfolios and checked-in baseline are v2 and cannot be translated into eligible v3 evidence.
- `CARTULARY_TEST_RUN_ID=remediation-w00-markdown-20261017-01 make lint-markdown` passed (outer exit 0). The repository configuration excludes `temp/**`, so this does not establish lint coverage for the ignored controlling tracker; its structure and dispositions were inspected directly.
- The cold reference manifest binds the starting clean commit and source digest `sha256:59068a74158462c1b9bb3c42ea86fab124dedbb551f41c35e46b5feaec47e10e`. W00 records this command as pre-change correctness work; its timing is not qualification evidence.
- Sequencing clarification requested: preserve the frozen source/failure evidence and proceed with remediation while retaining final qualification, or repair/qualify the original baseline first. User resolved the sequencing question by removing harness performance qualification from the plan/specifications; no reference collection or publication is required.
- Tracker update: replace proposal-only content with twelve-section execution tracker, preserve all disposition mappings and rationale, record explicit browser-suite decision.

| Evidence | Source / result | Applicability |
| --- | --- | --- |
| E-000 | `/tmp/cartulary-remediation-baseline-6rr9ozw1/baseline.json` | Start-of-execution content/index preservation evidence |
| E-001 | `.cartulary/test-results/remediation-w00-doctor-20261017-01/doctor/tool-run-summary.json` | Pinned tooling readiness only; not qualification |
| E-002 | `.cartulary/test-results/remediation-w00-reference-check-cold-20261017-01` | Running reference candidate; not yet accepted |

### W01 completion checkpoint — 2026-10-17

- Changed `docs/testing-harness-nlspec.md`, Core 04's browser build hook and the OTel specification's generated-check reference and exact runtime dependency policy. `domain.md` needs no new vocabulary. These apply the user's approved normative changes; independent human projection-fidelity review is not claimed by lint or this checkpoint.
- Adopted four gates and 62 public names (55 main registry + 7 UI-review bindings); closed selectors; canonical operations; exact browser/tier/policy selection; distinct contexts; smoke presets; immutable validation context/finalizer transaction; scoped read-only evidence; consumer builds/advisory CI policy.
- Removed harness performance windows, improvement/no-regression portfolios, observer-overhead qualification, qualification readers/schema requirements and baseline publication. REQ-377/380/848/850 and AC-059/079/145 are explicitly retired. Canonical timing, authored estimates, diagnostics, product measurement, security and Core 05 remain required.
- Existing-requirement repairs: exact selection, broad-regression placement and advisory placement. New behavior: command consolidation/required browser selector, private bindings, four gates, shared validation context, evidence variants and user-directed qualification retirement.
- Exact validation: `CARTULARY_TEST_RUN_ID=remediation-w01-markdown-20261017-01 make lint-markdown`, then `...-02`, then `...-03` after intervening document edits: each outer exit 0. Final W01 E-004: `.cartulary/test-results/remediation-w01-markdown-20261017-03/adhoc/lint-markdown/tool-run-summary.json`. `git diff --check`: 0. Temp tracker is lint-excluded and reviewed separately.
- No generation, finalization or product verification started for this documentation-only slice. Active executable surface remains baseline until W09. Breaking migration has no aliases or runtime translation.
- W00 cold check drained: outer Make 2, harness 10; 798/799 units passed, one failed; all 1,080 cleanup steps passed. E-003 root `.cartulary/test-results/remediation-w00-reference-check-cold-20261017-01`, exact clean starting source retained. Failed row `harness.command_surface.unit.native_runtime`, `TestRuntimeSelectedGraphExecution/off`; nested evidence row `harness.evidence_accounting.behavior.current_epoch_evidence` fails `tools/harness/observability/tests/test-wrapper-receipt.mjs:50` with `native receipt rejected`. Nested E-003a: `native-engine/selected-graph-90313-off` beneath E-003. Repair/validate in evidence/runtime work; no blind rerun.
- `make explain-run RESULTS_DIR=<E-003>` and nested summary/target-log views returned 0. A nested `DETAIL=logs` attempt without TARGET returned 2; corrected `TARGET=test-slice` returned 0. Historical release cleanup failure and pre-existing wrapper receipt failure remain open correctness risks.
- Remaining: W02–W09 implementation/projections, W10 final-source validation, W11 review/handoff. Public count proves neither savings nor assurance. Retained-run finalizer maintenance remains ineligible until a successful full warm check exists.

### W02 execution note — IN_PROGRESS (E-005)

- Owner guidance: `make task-guide ROLE=module-author OWNER=harness.command_surface`, and the same command for `harness.browser`, `harness.generated_artifacts`, and `harness.evidence_accounting`; all outer exits 0.
- Characterization: `CARTULARY_TEST_RUN_ID=remediation-w02-graph-before-20261017-01 make harness-graph-contract`, outer 0. Artifacts: `.cartulary/test-results/remediation-w02-graph-before-20261017-01/harness-graph-contract/`.
- Authored changes so far: 142 private execution definitions in the existing work-graph owner; typed command bindings; a single-operation dispatcher without shell expansion or dependency scheduling; operation identities in policy graphs, shell rows, cache bindings and frontend receipts. Genuine file producers remain in Make. Current public names remain active pending W09.
- Contract migrations: task-surface/work-graph owners v3, cache registry v2, frontend-build receipts v2. Incompatible producer receipts reject; operation digests invalidate changed execution/cache identity. Generated Make projections temporarily retain the current surface and are not an additional authority.
- Regeneration: `CARTULARY_TEST_RUN_ID=remediation-w02-generate-20261017-0N make generate` for N=1,2,3,4,5,6; all outer 0. Each summary is under its exact run root at `generate/tool-run-summary.json`. Generated validator numbering changes were inspected and arise from the revised schema graph.
- Schema checks: `CARTULARY_TEST_RUN_ID=remediation-w02-schema-20261017-0N make json-shape-check`: N=1 outer 2 before publication (overbroad private binding projection); N=2 outer 2 (misplaced import); N=3 outer 2 (expected generated-input drift after edits); N=4 and N=5 outer 0. Investigated each failure before correction/rerun. Private binding projection was removed; runtime receives existing exported bindings without widening the plan wire contract.
- Graph checks: `CARTULARY_TEST_RUN_ID=remediation-w02-graph-after-20261017-0N make harness-graph-contract`: N=1 outer 2 (two old-ID/projection assertions); N=2 outer 2 (receipt schema still admitted old producer IDs); N=3 outer 0, 19/19 tests. New assertions cover identity independent of command spelling, executable identity invalidation, unknown operations, forbidden overrides and literal argument transport.
- `CARTULARY_TEST_RUN_ID=remediation-w02-boundary-20261017-01 make harness-boundary-contract`: outer 0, 26/26 tests.
- `CARTULARY_TEST_RUN_ID=remediation-w02-scheduler-20261017-01 make harness-scheduler-contract`: outer 2, 5/6 tests; catalog loader called repeatedly by the new cache binding lookup. Invocation-local catalog reuse restored; rerun pending.
- `CARTULARY_TEST_RUN_ID=remediation-w02-surface-20261017-01 make harness-command-surface-contract`: outer 2, two stale fixtures (old Govulncheck unit ID and unprojected operation bindings); corrected, rerun pending.
- W02 remains open. The later policy/selector/evidence cutover workstreams have not started. No final release or performance claim is made. All W02 failures above are implementation regressions found during this slice, distinct from E-003's pre-existing native receipt failure.

### W02 completion checkpoint — DONE (E-006)

- Completion: canonical policy and shell-row dispatch now invokes the private operation runner. It executes one operation, never schedules dependencies, evaluates shell/Make expressions, or invokes a phony Make target. Make remains available only for actual file production. Go file producers perform their readiness within their recipes, eliminating the private Make-target dependency. Generation and its old wrapper bind one operation. Current generated public/private wrappers are transitional projections scheduled for deletion in W09; they are not compatibility aliases for the future surface.
- Cache policy is keyed by operation IDs or row-runner types; catalog facts load once per cache-plan invocation. Frontend artifacts bind canonical producers under v2. Typed owners reject unknown operation bindings and executable overrides. Public names/reporting views do not alter operation identity; changed executable definitions alter their digest.
- Exact successful follow-up commands (all outer 0):
  - `CARTULARY_TEST_RUN_ID=remediation-w02-graph-after-20261017-04 make harness-graph-contract` — 19/19.
  - `CARTULARY_TEST_RUN_ID=remediation-w02-scheduler-20261017-02 make harness-scheduler-contract` — 6/6.
  - `CARTULARY_TEST_RUN_ID=remediation-w02-surface-20261017-02 make harness-command-surface-contract` — 33/33.
  - `CARTULARY_TEST_RUN_ID=remediation-w02-shell-dispatch-20261017-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity` — graph pass; exact shell row executed through the operation dispatcher. Manifest source digest `sha256:1a65541f6df9e72bac1e2c9191ed99c60473691ca125347fc4ce13c75b1b7f2d`; a later graph-only assertion addition does not change this selected assertion inventory.
  - `CARTULARY_TEST_RUN_ID=remediation-w02-drift-20261017-01 make generate-drift` — graph pass; current source regenerates exactly.
  - `CARTULARY_TEST_RUN_ID=remediation-w02-policy-20261017-01 make generated-artifact-policy-check` — graph pass.
  - `CARTULARY_TEST_RUN_ID=remediation-w02-schema-20261017-06 make json-shape-check` — graph pass.
  - `CARTULARY_TEST_RUN_ID=remediation-w02-catalog-20261017-01 make test-catalog-check` — pass.
  - `CARTULARY_TEST_RUN_ID=remediation-w02-markdown-20261017-01 make lint-markdown` — pass.
  - `git diff --check` — 0.
- Artifact locations: every exact run is under `.cartulary/test-results/<run-id>/`. Graph runs retain `run-manifest.json`, `run-summary.json`, unit logs/results and execution context. Direct contracts retain `<target>/tool-run-summary.json`, `<target>/<target>/stdout.log` and step summaries. Markdown summary is `adhoc/lint-markdown/tool-run-summary.json`.
- Applicability: WSL2/pinned toolchain, slice validation only. The earlier boundary 26/26 pass remains relevant; later changes addressed producer identity, caching and Make file dispatch, validated by the follow-up contracts. No current-candidate check/ci/release result is claimed. W08 still owns E-003 receipt failure diagnosis, and W09–W11 own cutover, finalization, full validation and handoff. Full baseline/performance qualification remains removed by user decision.
- Migration/rollback: owner schemas v3, cache registry v2 and frontend receipts v2 require the coordinated generated outputs. Old producer identities/cache keys are not translated. Roll back the coordinated source/projection changes as a unit. No aliases are a recovery path.
- Next entry: W03 may begin only after this checkpoint is written. Remaining risks are the explicitly tracked W03–W11 contracts, including telemetry security coverage and complete release cleanup.

### W03 execution note — IN_PROGRESS (E-007)

- Entry: E-006 saved before W03. Owner guidance `make task-guide ROLE=module-author OWNER=<owner>` passed for `harness.browser`, `harness.test_catalog`, and `platform.telemetry` (outer 0).
- Changes: aggregate browser selections now pass exact row IDs rather than expanding stages. Primitive direct browser stages use the same compiler path. Preparation is scoped per selected browser stage. OTel conformance and the graph compiler share one explicit 11-row dependency selection, including seven release-tier security/runtime rows with recorded rationales. Service scope rejects duplicate, unknown, foreign-owner and out-of-scope rows before compiling preparation.
- Tests added: four full-tier measurement rows versus 29 excluded release-tier rows; complete direct measurement selection; excluded-row preparation trap with a positive selection control; required telemetry/security preservation, unrelated-row growth, missing/duplicate dependency rejection; invalid service scope before compilation.
- `CARTULARY_TEST_RUN_ID=remediation-w03-generate-20261017-0N make generate` for N=1,2,3: outer 0; each summary under its exact `generate/tool-run-summary.json` run path.
- `CARTULARY_TEST_RUN_ID=remediation-w03-schema-20261017-01 make json-shape-check`: outer 0, graph pass.
- `CARTULARY_TEST_RUN_ID=remediation-w03-otel-20261017-01 make otel-conformance`: outer 0, graph pass; current required telemetry evidence executed and consumed successfully.
- `CARTULARY_TEST_RUN_ID=remediation-w03-graph-20261017-01 make harness-graph-contract`: outer 2, 19/22 passed. Exact-selection tests passed; release graph exposed conflicting support-browser dependency definitions because row compilation attached preparation for every browser stage to each stage. Fixed by scoping preparation to each selected stage and unifying direct/row primitive stage compilation. Follow-up graph/adapter/drift checks pending; W03 remains open.
- All artifacts are retained under `.cartulary/test-results/<exact-run-id>/`, with native graph summaries or `<target>/<target>/stdout.log` for direct contract runs. No full browser or release acceptance is inferred from graph tests; those remain W10 obligations.

### W03 completion checkpoint — DONE (E-008)

- Final changes: exact browser row sets survive aggregate compilation, stage batching and runner filters. Each stage receives only its selected consumers' preparation. Primitive direct stages and row/owner/aggregate paths share the same compiler path, preserving snapshot-update mode. Explicit required row dependencies replace owner-wide expansion; shared resolver serves OTel compilation and evidence checking. Invalid service-scope selections reject before compiling work.
- `CARTULARY_TEST_RUN_ID=remediation-w03-graph-20261017-02 make harness-graph-contract`: outer 0, 22/22 tests; includes four full-tier versus 29 release-tier measurement rows, complete direct suite selection, excluded preparation trap, telemetry closure and invalid scope fixtures.
- `CARTULARY_TEST_RUN_ID=remediation-w03-playwright-20261017-01 make harness-playwright-adapter-contract`: outer 0, 15/15 tests; adapter filters and row accounting remain valid.
- `CARTULARY_TEST_RUN_ID=remediation-w03-drift-20261017-01 make generate-drift`: outer 0, graph pass on regenerated source.
- Earlier W03 `json-shape-check` and `otel-conformance` passes remain applicable to the unchanged schema/dependency inventory and checker. Read-only inspection confirmed 11 retained telemetry rows, all `passed`, including `platform.telemetry.unit.privacy_and_redaction`. Later changes affected browser compilation only.
- `git diff --check`: 0. All exact-run artifacts remain at `.cartulary/test-results/<run-id>/`; graph/adapter logs and summaries use the same conventions recorded in E-006.
- Migration: ordinary aggregate browser scope is intentionally narrower; direct suites remain complete. Required higher-tier telemetry evidence remains in policy closure. Explicit out-of-scope service rows now fail instead of disappearing. No assurance is claimed from reduced selection counts.
- Remaining risks: broad smoke placement/deduplication (W04), finalizer freshness (W05), gate/interface changes (W06–W09), and final-candidate complete browser/release execution (W10–W11). The known pre-remediation receipt/cleanup failures remain open. No harness performance qualification is required.
- W04 entry is authorized only after this checkpoint is saved.

### W04 execution note — IN_PROGRESS (E-009)

- Entry: E-008 saved before W04. Current-grammar owner guidance completed for `harness.command_surface`, `harness.test_catalog` and `harness.browser` (outer 0).
- Authored changes: split scheduler and fixture ownership assertions into fast, exclusively isolated Go rows; move remaining native runtime/fixture assertions to full tier. Work-graph execution contexts fix the two narrow smokes at one CPU and no cache reuse. Ordinary LPT batching and separate race-enabled contracts retain their contexts. Typed row/operation members let the existing compiler select three canonical smoke units and project each role without another scheduler.
- Normative clarification: the narrow Go smoke execution context is selection independent; broad regressions remain full tier. Existing row IDs retain their remaining assertions; new row IDs own the extracted assertions. No retired-row alias is needed.
- `CARTULARY_TEST_RUN_ID=remediation-w04-generate-20261017-01 make generate`: outer 2; verification entries were not ASCII-sorted. Corrected authored ordering before retry.
- `CARTULARY_TEST_RUN_ID=remediation-w04-generate-20261017-02 make generate`: outer 0; projections regenerated through Make.
- Focused graph and exact smoke row runs launched under explicit W04 run IDs; results pending. Artifacts: `.cartulary/test-results/<exact-run-id>/`. No gate or release acceptance is claimed.

### W04 completion checkpoint — DONE (E-010)

- Completed F-004 / C-004 / V-004: three check smoke roles compile to three physical units and truthful role projections. Scheduler and fixture ownership assertions have separate fast rows and fixed isolated non-race execution contexts; remaining native runtime/fixture assertions are full tier. CI includes both narrow and broad rows. Race-enabled packages remain separate operations; installed-runtime and seeded UI contexts were unchanged.
- Row split: `harness.command_surface.unit.native_runtime` retains 18 assertions and yields `harness.command_surface.unit.scheduler_smoke` (`TestRuntimeScheduling`). `harness.browser.unit.native_fixtures` retains seven assertions and yields `harness.browser.unit.fixture_broker_smoke` (`TestFixtureBrokerOwnership`). Existing rows remain active; assertion inventories are conserved.
- Exact successful commands (each outer 0):
  - `CARTULARY_TEST_RUN_ID=remediation-w04-graph-20261017-01 make harness-graph-contract` — 23/23, including equivalent direct/aggregate smoke units at 16 available Go lanes and separate race contexts.
  - `CARTULARY_TEST_RUN_ID=remediation-w04-scheduler-smoke-20261017-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.scheduler_smoke` — graph pass.
  - `CARTULARY_TEST_RUN_ID=remediation-w04-fixture-smoke-20261017-01 make test-slice OWNER=harness.browser ROWS=harness.browser.unit.fixture_broker_smoke` — graph pass.
  - `CARTULARY_TEST_RUN_ID=remediation-w04-schema-20261017-01 make json-shape-check` — graph pass.
  - `CARTULARY_TEST_RUN_ID=remediation-w04-catalog-20261017-01 make test-catalog-check` — pass.
  - `CARTULARY_TEST_RUN_ID=remediation-w04-markdown-20261017-01 make lint-markdown` — pass.
  - `CARTULARY_TEST_RUN_ID=remediation-w04-drift-20261017-01 make generate-drift` — graph pass.
  - `git diff --check` — 0.
- Artifacts: `.cartulary/test-results/<exact-run-id>/`, native `run-summary.json`/row results for selected rows; direct graph-contract summary and assertion log under `harness-graph-contract/`. WSL2 pinned tooling; this is slice evidence, not full release acceptance.
- Compatibility: tier placement and row decomposition intentionally change selected workloads. Narrow smokes cannot reuse a result and no longer execute inside broader non-race catalog rows. Current private Make projections remain transitional until W09; future presets select these canonical units through the compiler. Public-wrapper assertions are retained in the private operation and are covered by final gate execution after cutover.
- Remaining: W05 freshness/transaction work, W06–W09 policy/interface/evidence cutover, W10–W11 complete final-source acceptance. E-003 receipt failure remains assigned to W08. No performance qualification or publication requirement remains.
- W05 starts only after this checkpoint is saved.

### W05 execution note — IN_PROGRESS (E-011)

- Entry: E-010 saved before W05; `make task-guide ROLE=module-author OWNER=harness.generated_artifacts` outer 0. Characterization `CARTULARY_TEST_RUN_ID=remediation-w05-evidence-before-20261017-01 make harness-evidence-contract` outer 0, 99/99. This direct context does not resolve E-003's nested runtime failure.
- Changes: immutable invocation-local validation facts bound to the executable source snapshot; schema and catalog finalizer actions share catalog validation. No facts are serialized as reusable success. Source changes invalidate facts before consumption; generation creates a new context and reloads compiled schema validators. Every action/substep retains its own outcome and source binding. Pre-mutation retained identity remains separate.
- Finalizer transactions now compare content, not Git-status strings, so edits to already-dirty tracked files are reported. Unexpected exceptions also roll back tracked inputs; original failure identity survives; untracked files remain untouched. Added negative fixtures for immutable maps/rows, changed inputs, mutation during validation, and preserved pre-existing edits.
- Summary migration: `cartulary.agent_finalize_summary.v4`; validation-fact substeps carry `validation_source_digest`; retained canonical evidence is named `retained_evidence`. V3 receipts are incompatible. No baseline publication or performance qualification is introduced.
- `CARTULARY_TEST_RUN_ID=remediation-w05-generate-20261017-01 make generate`: outer 0.
- `CARTULARY_TEST_RUN_ID=remediation-w05-schema-20261017-01 make json-shape-check`: outer 0.
- `CARTULARY_TEST_RUN_ID=remediation-w05-evidence-after-20261017-01 make harness-evidence-contract`: outer 2, 98/99; new test fixture object had a syntax error. Corrected before retry.
- `CARTULARY_TEST_RUN_ID=remediation-w05-evidence-after-20261017-02 make harness-evidence-contract`: outer 0, 99/99.
- `CARTULARY_TEST_RUN_ID=remediation-w05-finalize-20261017-01 make agent-finalize`: running. `RESULTS_DIR` intentionally unset; retained-run maintenance must be recorded as skipped. Artifacts under `.cartulary/test-results/<exact-run-id>/`.

### W05 completion checkpoint — DONE (E-012)

- Closed F-005 / C-005 / V-005: schema/catalog consumers share immutable invocation facts, bound to the exact admitted executable snapshot. Every access checks freshness; a validator that changes inputs cannot publish a fact. Generation discards the context, reloads generated validators and validates the refreshed source. Independent action receipts retain source digests. Frozen retained-run identity is not rewritten.
- Rollback now covers both ordinary failure results and thrown exceptions. File-content comparison detects finalizer changes to already-dirty tracked files. Negative fixtures preserve pre-existing tracked content/modes/symlinks and untracked output; primary failure identity and capture cleanup classification remain intact. Finalization's maintenance-only action registry remains closed.
- Exact additional commands (outer 0):
  - `CARTULARY_TEST_RUN_ID=remediation-w05-finalize-20261017-01 make agent-finalize` — graph pass, 29,227 ms. Finalizer caused **zero** generated changes; `updated_files=[]`, rollback `not_needed`. Schema/catalog/tier receipts bind `sha256:cd8221aa3144eed1f1792c46589b9fe064516a2417a2c08d594794b426b860e4`; refreshed-source validation passed after the Make-owned generation transaction. Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
  - `CARTULARY_TEST_RUN_ID=remediation-w05-boundary-20261017-01 make harness-boundary-contract` — pass.
  - `CARTULARY_TEST_RUN_ID=remediation-w05-markdown-20261017-01 make lint-markdown` — pass.
  - `git diff --check` — 0.
- Reuse E-011's 99/99 follow-up evidence-contract and schema-check passes; no redundant generation-drift rerun is needed because finalization completed its Make-owned refresh/validation/comparison transaction and current-source validation without changes.
- Exact finalizer artifact: `.cartulary/test-results/remediation-w05-finalize-20261017-01/unit-artifacts/finalize-summary.json`. Other roots follow `.cartulary/test-results/<exact-run-id>/`; contract logs/summaries remain under their target directories. WSL2/pinned environment, slice evidence only.
- Migration: v4 finalizer receipts replace v3, including `validation_fact`, `validation_source_digest`, and `retained_evidence`. No historical receipt translation. Coordinated source/schema/generated rollback remains the recovery procedure.
- Remaining: complete retained finalizer eligibility is still governed by its existing negative fixtures; this run supplied no eligible full warm check. W09 finalization and W10–W11 complete final-candidate acceptance remain required. E-003's nested receipt and historical release cleanup failures remain open; direct evidence-contract success does not close them.
- W06 starts only after this checkpoint is saved.

### W06 completion checkpoint — DONE (E-013)

- Entry: E-012 saved before W06; `make task-guide ROLE=module-author OWNER=harness.command_surface` outer 0.
- Closed policy/preparation work F-006 / C-006 / V-006: local check no longer roots server/operator/migrate/server-harness builds or Fallow reports. Exact selected consumers retain producer edges. CI explicitly selects production build qualification and Fallow analysis; release inherits CI. Targeted gosec remains independent of advisory gosec. Obsolete Go timing-roster qualification was removed from local policy, per the user's performance retirement.
- Source schema validation now validates complete Fallow configuration semantics through a read-only owner facade; missing owner-declared inputs remain local failures. Graph contracts verify no local advisory selection, retained blocking policy, CI/release build/report qualification, required consumer edges and independence of targeted security from advisory dependencies. The negative Fallow fixture first succeeds, then removes a declared input and rejects; writes are trapped.
- Corrected REQ-070's stale invocation inventory: four concurrent scans are global dead code, dupes, health and bounded package dead code; Markdown is rendered from retained JSON without another scan. This documents the existing scanner behavior and preserves its bounded enforcement.
- Exact validation commands (outer 0):
  - `CARTULARY_TEST_RUN_ID=remediation-w06-generate-20261017-01 make generate`.
  - `CARTULARY_TEST_RUN_ID=remediation-w06-graph-20261017-01 make harness-graph-contract` — 25/25; exact row/policy/producer closure assertions for the four destination gates.
  - `CARTULARY_TEST_RUN_ID=remediation-w06-schema-20261017-01 make json-shape-check`.
  - `CARTULARY_TEST_RUN_ID=remediation-w06-drift-20261017-01 make generate-drift`.
  - `CARTULARY_TEST_RUN_ID=remediation-w06-markdown-20261017-01 make lint-markdown`.
  - `make target-plan TARGET=<gate>` for `test-fast`, `check`, `ci`, `release-check`; summaries retained in `temp/remediation-w06-<gate>-plan.txt`.
  - `git diff --check` — 0.
- Discovery failures, investigated before correction: `make target-plan TARGET=check JSON=1` and CI equivalent returned outer 2 (JSON undeclared); `CARTULARY_OUTPUT_MODE=machine make target-plan TARGET=<gate>` returned outer 2 for all four gates (machine mode unsupported). Current target-plan provides human summaries through Make; full JSON selector/output support belongs to W07. Rejected outputs are not evidence of execution. Exact closures are covered by graph-contract assertions; summary files are planning artifacts only.
- Artifacts for fresh validation remain under `.cartulary/test-results/<exact-run-id>/`. WSL2 pinned tooling. No final check/CI/release execution is claimed.
- Migration: `check` no longer guarantees unconditional production qualification. Fallow reports remain directly selectable and are required in CI/release. Some production/harness artifacts still appear when a selected consumer requires them. Command-name deletion, including the old `test` entry and its transitional empty-policy tier binding, is deliberately reserved for the coordinated W09 cutover; no compatibility alias will survive it.
- Remaining: W07 selectors/lifecycle, W08 evidence/qualification-reader retirement, W09 activation and caller migration, W10–W11 final acceptance. Existing receipt/release cleanup risks remain open.
- W07 starts only after this checkpoint is saved.

### W07 execution note — IN_PROGRESS (E-014)

- Entry: E-013 saved before W07. Owner guidance `make task-guide ROLE=module-author OWNER=harness.command_surface` and `OWNER=harness.release`: outer 0.
- Initial shared input resolver work: typed cross-input constraints (`exactly_one`, `requires`) and a strict explicit-input resolver shared with execution semantics for future target-plan files. Only declared command inputs are accepted; arbitrary/shared/internal forwarding keys and non-string values reject. These APIs are not yet wired to the future public contracts; current registry remains active until W09.
- Current lifecycle inspection: `services-up` already performs one Compose startup and one wait per service. Legacy `db-up` repeats object-store initialization after the same startup; consolidation will remove that repeated path. Existing bootstrap must still work before Node/native harness installation, so its selector admission needs an installation-free entry boundary.
- W07 is not complete. Interim execution evidence and discoveries follow. Next: canonical private selection bindings/unions, selector/discovery adapters, lifecycle consolidation, negative-input and exact-selection validation. Preserve the one-owner/one-compiler design; no additional command registry or scheduler.


### W07 interim execution evidence — IN_PROGRESS (E-015)

- Canonical private unions now cover test/browser suites, build components, lint/source selections, package scenarios and all retained harness presets. Public bindings remain inactive until W09. Strict command-input resolution rejects unknown keys, non-string values, conflicts and invalid rows before compilation. Source-generated comparison is explicitly read-only; finalizer refresh remains a separate private operation.
- `target-plan` now has a shared pure compiler API, strict duplicate-key JSON parsing, public-only lookup and summary/graph/rows/artifacts rendering. Its closed wire schema becomes v5 with command identity, canonical inputs and exact rows. Owner guidance becomes v4 without a role field and includes exact row commands and coverage distributions; the current public ROLE grammar remains until W09.
- Bootstrap's future PART binding uses generated installation-free enum admission followed by pinned Node admission and a canonical private installation operation. Installation operations may produce declared absolute Make file targets and then perform their declared acquisition action; they do not schedule product graphs. Full acquisition remains explicit. Help's future ALL/JSON views are generated from the task owner and require no runtime provisioning.
- Local service startup now avoids the second object-store initialization in the retiring db-up path. Repeated services-up fixtures verify one readiness pass per service, proxy reuse and absence of resets/migrations. Final removal of the obsolete entry point belongs to W09.
- Commands and investigated outcomes:
  - `CARTULARY_TEST_RUN_ID=remediation-w07-generate-20261017-0N make generate`: N=1–5 outer 0. N=6 outer 2: expanding complete help for the still-active 121-command surface exceeded the existing generated Make size budget. Scoped the new help projection to activation of the consolidated help contract, preserving the budget. N=7 outer 0.
  - `CARTULARY_TEST_RUN_ID=remediation-w07-graph-20261017-0N make harness-graph-contract`: N=1/2 outer 2 because private union names/member strings were still constrained to Make-name syntax; corrected the owner schema. N=3 outer 2, 28/29, because the new negative fixture's diagnostic regex omitted "must not be empty"; rejection itself was correct. N=4 outer 0, 32/32, including every suite/build/preset/source/lint selection, package sharing, plan equivalence and installation path rejection.
  - `CARTULARY_TEST_RUN_ID=remediation-w07-surface-20261017-01 make harness-command-surface-contract`: outer 2, 34/35. Bootstrap admission and existing wrapper cases passed. New JSON help output expanded authored Make path templates; escaped literal metadata in its generated projection. Follow-up pending.
- Artifacts: `.cartulary/test-results/<exact-run-id>/`; generation failure diagnostic at `remediation-w07-generate-20261017-06/generate/generate-task-surface-and-execution-topology/stderr.log`; graph/surface logs below their same-named target directories. All results are slice evidence only. W07 remains open pending wrapper revalidation, source/schema/drift checks and lifecycle/diagnostics fixtures. No final-candidate qualification is claimed.

### W07 assertion reachability correction

Inspection found that `test-dev-services-lifecycle.sh` was listed as a backing script but had no executable caller. W07 adds one canonical local-service lifecycle operation to both the default contract union and the lifecycle preset, and updates their normative assertion inventories. The shared operation deduplicates when both views are selected. Its fixture owns only isolated mock processes and preserves the real service state. This repairs missing validation of the required idempotent startup behavior; it does not expand product suites or local gate smoke roles.


### W07 follow-up evidence and boundary review — IN_PROGRESS (E-016)

- `remediation-w07-surface-20261017-02` (same surface command as E-015): outer 2. The literal-dollar escaping needed a replacement callback (string replacement interprets `$$`); corrected. Also exposed the missing schema declaration for the new installed-runtime policy. `remediation-w07-surface-20261017-03`: outer 0, 35/35, including installation-free help and bootstrap admission fixtures.
- Additional generation commands `CARTULARY_TEST_RUN_ID=remediation-w07-generate-20261017-0N make generate`, N=8/9, and `...-10`, `...-11`, `...-12`: outer 0. Generated source-fidelity checks correctly invalidate the render index after transitive generator input changes; regenerate after the final source edits before retrying.
- `CARTULARY_TEST_RUN_ID=remediation-w07-schema-20261017-01 make json-shape-check`: outer 2 before graph execution, undeclared `runtime_readiness` in the operation schema. Read-only diagnosis `make target-plan TARGET=json-shape-check`: outer 2 with the exact schema message. Added the closed `installed_only` schema value.
- Same schema command `...-02`: outer 2, stale generated renderer input; `...-03`: outer 0 after generation. `...-04`: outer 2 after further CLI/facade edits invalidated the transitive render index. These are generated-projection freshness failures, not product-suite failures.
- `CARTULARY_TEST_RUN_ID=remediation-w07-contract-20261017-01 make harness-contract`: outer 2, 8/10 units passed. New local-service lifecycle operation passed. Two old tests assumed every union member was a Make-target string; updated them to resolve canonical typed members. `...-02`: outer 2, 8/10; one stale generation-index fixture and one over-broad worker assertion that omitted the existing separately reserved four-file evidence context. Preserve that context and regenerate before retry.
- `CARTULARY_TEST_RUN_ID=remediation-w07-drift-20261017-01 make generate-drift`: outer 2, exact `tools/execution_topology_render_index.json` drift after CLI/facade edits. No generated output was hand-edited.
- Markdown commands `CARTULARY_TEST_RUN_ID=remediation-w07-markdown-20261017-01 make lint-markdown` and `...-02`: outer 0.
- Read-only discovery: `make target-plan-json TARGET=check > temp/remediation-w07-check-plan.json` and `make task-guide ROLE=module-author OWNER=harness.command_surface JSON=1 > temp/remediation-w07-owner-guide.json`: outer 0. These use the supported pre-cutover grammar and emit the new v5/v4 projections; they are planning artifacts, not execution evidence.
- Additional boundary fixes: target-plan INPUTS now uses the existing bounded, race-aware, no-symlink local-file reader, with documentation rejection before filesystem metadata access. Added negative fixtures for forbidden paths, duplicate keys and symbolic links. Coverage guidance retains family, runtime/resource profile and fixture distributions. New command bindings perform exact selection admission before retained-root creation or file producers; their entrypoint requires installed Node and reports the explicit bootstrap command when absent.
- Remaining W07 work: finish frozen-source checks and record the completion checkpoint. W08 owns retained evidence/observability variants and retirement of qualification readers; W09 owns live public activation, all old-name removal and caller migration. The four W06 human plan text files are currently non-ignored scratch artifacts; move them into retained ignored storage before the W09/W10 candidate freeze so they do not pollute executable source identity.


### W07 completion checkpoint — DONE (E-017)

- Closed the implementation portion of F-007 / C-007 / V-007: typed private selections cover all test/browser suites, build components, lint/source checks, package scenarios and assurance presets. Owner/suite conflicts, invalid/duplicate/foreign rows and service-scope errors reject during read-only admission before producers. New browser binding requires SUITE. Installation-free generated bootstrap admission closes PART and retains full bootstrap on omission. Frontend/browser preparation uses the existing Make file producers and canonical operation definitions.
- Discovery now shares the execution resolver/compiler. INPUTS admits only bounded safe JSON files and declared command fields; documentation and links reject before traversal. Generated complete help exposes selectors/defaults/side effects without a runtime. Owner coverage retains exact rows, families, profiles, fixture capabilities, evidence classes and narrow commands. Discovery requires installed Node instead of installing it. Package scenario union has one canonical package producer. Repeated local service startup performs one readiness pass per service and never resets or migrates.
- Final frozen-slice commands (outer 0):
  - `CARTULARY_TEST_RUN_ID=remediation-w07-generate-20261017-13 make generate` — Make-owned generation complete.
  - `CARTULARY_TEST_RUN_ID=remediation-w07-contract-20261017-03 make harness-contract` — 10/10 units pass, zero failed/skipped/cancelled, 56,717 ms. Includes all eight existing contract groups and the newly reachable local-service lifecycle operation, with shared preparation. Earlier graph and wrapper focused passes are retained in E-015/E-016; the complete run validates their latest source together.
  - `CARTULARY_TEST_RUN_ID=remediation-w07-drift-20261017-02 make generate-drift` — 4/4 units pass, 25,556 ms.
  - `CARTULARY_TEST_RUN_ID=remediation-w07-schema-20261017-05 make json-shape-check` — 3/3 units pass, 7,276 ms.
  - `git diff --check` — 0. Reuse E-016's final Markdown pass; subsequent edits changed executable source and this ignored tracker only.
- Final contract source digest: `sha256:a0b6276b6dccb42b7e6021aa7a7a8cccfa716011822f84c3879e65c70ea41503`. Exact roots are `.cartulary/test-results/<exact-run-id>/`; each graph has run manifest/summary, unit results/logs, target projections and cleanup results. Pinned WSL2 environment; these close the slice, not final release readiness.
- Migration: target-plan wire contract v5 and task-guide v4 replace v4/v3 respectively. New input bindings remain inactive in the live registry until W09's coordinated cutover; current public ROLE and predecessor names have not become compatibility promises. Retained observability variants belong to W08; their public activation and all 190 name dispositions belong to W09. Existing evidence/security/product contexts remain intact.
- Remaining risks: E-003 nested native receipt failure and historical release cleanup failure remain open for W08/W10. Final-candidate source, check/CI/release qualification and finalization are not yet complete. Four non-ignored human planning text files must move to ignored retained storage before candidate freeze. No harness timing qualification or baseline publication is required or claimed.
- W08 starts only after this checkpoint is saved.

### W08 execution note — IN_PROGRESS (E-018)

- Owner guidance: `make task-guide ROLE=module-author OWNER=harness.evidence_accounting` and `OWNER=harness.generated_artifacts`, outer exits 0.
- Changes underway: canonical resolved input recording (including sorted rows/defaults), manifest/runtime-plan v3, scoped read-only observability checks with explicit `full_validity`, exact fixture explanation, stronger complete evidence/context/cleanup/result closure, and native receipt clock parsing without floating-point seconds.
- Receipt diagnosis: native `/proc/uptime` parsing multiplied binary floating-point seconds by 1000; some hundredth-second values produce fractional milliseconds and violate receipt integer fields. Integer hundredths now match the Node reader; focused/native and nested-runtime evidence remain required before closing E-003.
- `CARTULARY_TEST_RUN_ID=remediation-w08-generate-20261017-01 make generate`: outer 0. First evidence/native/schema invocations (`remediation-w08-{evidence,native,schema}-20261017-01`) exited outer 2 because the new preparation import requested an unexported contract facade member. Evidence summary: `.cartulary/test-results/remediation-w08-evidence-20261017-01/harness-evidence-contract/tool-run-summary.json`. No functional pass claimed.
- Added the missing facade export. `CARTULARY_TEST_RUN_ID=remediation-w08-generate-20261017-02 make generate`: outer 0. Corresponding `-02` evidence/native/schema runs are in progress; outcomes must be recorded before checkpoint.
- Compatibility: v2 manifests/plans are historical; no translation or runtime aliases. New public selector activation remains W09. Performance qualification remains removed from acceptance; its exclusive implementation inventory is being retired without removing product measurement or descriptive timing.

- Sequencing refinement: physical deletion of qualification-only implementations moves with their public callers into W09. This preserves the requested single active-registry cutover; deleting those readers in W08 would leave still-active predecessor wrappers calling missing files. W08 must identify the complete removal closure; W09 must remove it, not retain aliases or stubs. Final acceptance is unchanged and qualification remains DROPPED.
- Removal closure for W09: `observability/qualification/**`, `canonical-performance.mjs`, `canonical-performance-cli.mjs`, `tests/test-qualification.mjs`, `tests/test-canonical-performance.mjs`, qualification-only foundation roster test; `tools/harness_instrumentation_qualification_plan.json`; legacy `tools/harness_public_target_duration_baselines.json`; five schema attachments (`harness_instrumentation_qualification.v2`, `harness_instrumentation_qualification_plan.v3`, `harness_instrumentation_qualification_result.v2`, `harness_performance_evidence_roots.v3`, `harness_public_target_duration_baselines.v3`); task-surface measurement profiles/bindings and their validator; retired public/private commands, Node-tool bindings and Fallow reachability entry. Retain `canonical-context.mjs` and its stronger non-qualification evidence coverage, product performance fixtures, instrumentation diagnostics and ordinary Go LPT estimates. Descriptive comparisons remain non-qualifying.

- Follow-up results: `remediation-w08-native-20261017-02` native observation row passed (outer 0); `remediation-w08-schema-20261017-02` passed (outer 0). Evidence `-02` failed because an old synthetic partial bundle expected full observability success; updated to assert full rejection and scoped timing success. Evidence `remediation-w08-evidence-20261017-03` passed 99/99 (outer 0). Graph command typo `make harness-work-graph-contract` (`remediation-w08-graph-20261017-01`) failed before work because no such target exists; corrected supported `make harness-graph-contract` (`-02`) passed (outer 0).
- Nested native row `CARTULARY_TEST_RUN_ID=remediation-w08-nested-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime` exited outer 2 / harness 10. Both preparation parity and selected-graph integration failed before graph execution: new full configuration resolution incorrectly required the enclosing invocation's prepared artifact identity and omitted native selector arguments. Replaced it with the existing input-only resolver plus explicit native selection origins; no global artifact-identity bypass was added. Fresh nested validation remains pending.

### W08 completion checkpoint — DONE (E-019)

- Canonical declared inputs now use the shared typed resolver with sorted unique row selectors and explicit defaults. Native selection arguments resolve independently of enclosing Make artifact identity. Replay removes canonical defaults before resolving command-line tokens. Manifest/runtime-plan v3 replace v2 across authored consumers and generated projections; historical runs retain their original schemas without translation. The immutable graph/context continues binding exact rows, canonical operation IDs and execution semantics separately from command/report identity.
- `CHECK=all|events|timing|evidence|diagnostics` implemented in the retained reader; only `all` returns `full_validity=1`. Complete checks require successful canonical events, context preimages, graph/command projections, cleanup, unit/row results, secret-scan receipt and complete published diagnostics. Partial views do not establish full validity. Exact fixture explanation derives canonical lease intervals; descriptive performance/resource views remain. Exact-run and export-header admission reject documentation paths before filesystem access, reject symlinks/ambiguous JSON, and perform no retained writes or implicit export.
- Fixed native receipt clock precision by parsing integer seconds/hundredths. Native observation coverage and the real nested off/basic/basic-second graph scenarios pass, closing the E-003 receipt defect for this source. New negative fixtures cover missing unit results, failed results, incomplete cleanup, duplicated graph units, changed selector identity, unsafe files and partial bundles. Equivalent row order resolves to the same inputs and graph.
- Final slice source digest: `sha256:544c85484be98fe356bc750dc792f5cc53c07fc8bf9e575c48182fab376e8f11` (dirty working source; base HEAD unchanged).
- Exact final-source commands and outer results:
  - `CARTULARY_TEST_RUN_ID=remediation-w08-generate-20261017-07 make generate`: 0; generated outputs refreshed through Make.
  - `CARTULARY_TEST_RUN_ID=remediation-w08-nested-20261017-02 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime`: 0; 1/1 graph units, 47,195 ms; all selected native assertions including three nested graph contexts passed.
  - `CARTULARY_TEST_RUN_ID=remediation-w08-contract-20261017-01 make harness-contract`: 0; 10/10 units, 103,723 ms, zero failed/skipped/cancelled.
  - `CARTULARY_TEST_RUN_ID=remediation-w08-drift-20261017-01 make generate-drift`: 0; 4/4 units, 21,859 ms.
  - `CARTULARY_TEST_RUN_ID=remediation-w08-schema-20261017-03 make json-shape-check`: 0; 3/3 units, 6,668 ms.
  - `CARTULARY_TEST_RUN_ID=remediation-w08-markdown-20261017-01 make lint-markdown`: 0; 27,140 ms.
  - `make harness-observability-check RESULTS_DIR=.cartulary/test-results/remediation-w08-nested-20261017-02`: 0; `check=all full_validity=1 read_only=1`.
  - `make harness-observability-check RESULTS_DIR=.cartulary/test-results/remediation-w08-nested-20261017-02/native-engine/selected-graph-39746-basic`: 0; complete basic diagnostics and shared outer terminal receipt validated, read-only.
  - `git diff --check`: 0.
- Artifacts: `.cartulary/test-results/<exact-run-id>/`; graph runs retain manifest, summary, unit events/results, execution context and cleanup. Nested contexts reside under the native run's `native-engine/selected-graph-39746-{off,basic,basic-second}`. Markdown summary is under `adhoc/lint-markdown/tool-run-summary.json`; generation under `generate/tool-run-summary.json`.
- Migration and remaining risks: v2 graph evidence no longer satisfies current consumers; new public input grammar remains inactive until W09. Physical deletion of qualification-only code/registries follows the recorded W09 closure to keep the command cutover coordinated. Product measurement and descriptive diagnostics remain. W09 must remove all retired callers and profiles, regenerate and finalize; W10/W11 still require complete final-source check/CI/release evidence and resolution of historical release cleanup risk. No performance qualification/publication claim is required or made.
- W09 starts only after this checkpoint is saved.

### W09 cutover progress — IN_PROGRESS (E-020)

- Pre-edit owner guidance completed with outer 0 for command-surface, generated-artifact, catalog, browser, release and telemetry owners using the old ROLE grammar. The public registry is now generated as 62 names; performance qualification/publisher remain removed per the scope amendment.
- Authored shell catalog selectors and preparation now reference canonical operations; runtime/frontend producer bindings no longer require callable Make names. Family schema v9, graph owner v4, task owner v4, task manifest v16, topology/browser batch v12 record this hard cutover. Existing report labels are reporting partitions only. Union preparation was moved to actual operation/row consumers; the static migration-input check no longer acquires database/build preparation merely by sharing a union.
- Removed the W08 qualification-only reader/publisher/plan/schema closure and measurement profile metadata. Descriptive timing, product measurements and ordinary Go LPT estimates remain. JSON help is a generated read-only projection beside the Make includes, avoiding installation and excessive Make source size.
- `CARTULARY_TEST_RUN_ID=remediation-w09-generate-20261017-01 make generate`: outer 2; catalog rejected a preparation declaration on a shell row. Corrected the static migration dependency rather than broadening the row permission. `...-02`: outer 2; expanded embedded JSON help exceeded the existing generated Make size budget. Moved JSON help to generated files. `...-03` and `...-04`: outer 0; generated through Make, no hand-edited outputs.
- `make help ALL=1 JSON=1 > /tmp/cartulary-w09-help.json`: outer 0, 62 command entries. `make browser-e2e`: outer 2, `usage_error`, required SUITE rejected before preparation. `make target-plan TARGET=check DETAIL=summary`: first outer 2 due to the plan schema's outdated projection-label pattern; repaired canonical operation partition admission; subsequent outer 0, 800 units. Counts are discovery/selection evidence, not acceptance.
- `CARTULARY_TEST_RUN_ID=remediation-w09-contract-20261017-01 make harness-contract`: outer 2 / artifact error; seven shell row adapters referenced a removed public-command lookup. Corrected launch identity inheritance from the native execution context. `...-02` is running; no pass claim yet.
- Exact artifact roots remain `.cartulary/test-results/<exact-run-id>/`. Current source remains under migration: caller/test/guidance updates, retained fixture view compatibility, focused validation and finalizer review must finish before W09 can close. W10/W11 have not started.

### W09 additional migration checkpoint — IN_PROGRESS (E-021)

- Source-check all passed on `remediation-w09-source-20261017-02` (outer 0). Generation 05, 07–09 and 11 passed; 06 was rejected before work because the run ID was incorrectly supplied on the Make command line, and 10 rejected an invalid authored forwarding enum, corrected before 11.
- Contract runs 03–06 and regression runs 01–02 remained outer 2; retained roots are `.cartulary/test-results/<exact-id>/`. Failures exposed stale callers/assertions, private preparation bindings, inherited graph execution flags, and one telemetry checker still using a public lookup key. These are being corrected; passing subsets do not close W09. Public surface remains 62 names.
- Canonical producer paths now come from execution topology; private tool handler identity is separate from report names. The telemetry checker and compiler share canonical explicit dependency selection and reject absent policy identity. Legacy duplicate Node wrapper capture was removed; capture/cleanup assertions moved to the active row adapter. Source-check generation now invokes public `generate` in its scratch checkout.
- Human W06 plan text moved to `.cartulary/remediation-records/`; no executable source depends on those reports. The user directed that the working tracker remain temporary and that a durable handoff live under `docs/handoffs`. The sole controlling tracker remains ignored at the requested path; `docs/handoffs/testing-harness-remediation.md` retains its complete snapshot and restart guidance. No Git ignore exception is retained. Both documents are human records, never runtime inputs.

- Further pre-handoff evidence: generation runs 12 and 13 and `remediation-w09-format-20261017-01` completed with outer 0. Contract 06 had one routing assertion failure; regression 02 had seven failing units. Follow-up repairs have been made but are not yet validated. W09 remains open.


### W09 durable handoff and follow-up validation — IN_PROGRESS (E-022)

- User-directed durable handoff created at `docs/handoffs/testing-harness-remediation.md`, including the complete tracker snapshot. `temp/plan-notes.md` remains ignored and temporary; `.gitignore` has no change. The durable handoff is a human record and has no executable consumers.
- `CARTULARY_TEST_RUN_ID=remediation-w09-markdown-20261017-01 make lint-markdown`: outer 0; summary `.cartulary/test-results/remediation-w09-markdown-20261017-01/adhoc/lint-markdown/tool-run-summary.json`; 34,380 ms.
- `CARTULARY_TEST_RUN_ID=remediation-w09-contract-20261017-07 make harness-contract`: outer 2, 62,783 ms; exact root `.cartulary/test-results/remediation-w09-contract-20261017-07`. Two rows failed: owner-boundary import isolation (new fixture read at import) and current-epoch evidence (new context-tampering test used timing-only validation). Corrected lazy fixture access and exercised the complete retained-evidence boundary; no acceptance rule was weakened.
- `CARTULARY_TEST_RUN_ID=remediation-w09-regression-20261017-03 make harness-contract SUITE=regression`: outer 2, 256,747 ms; exact root `.cartulary/test-results/remediation-w09-regression-20261017-03`. Four units failed: canonical evidence (same test), task-surface report (retired wrapper assertions), work-graph fast and matrix (removed test aggregate). Native runtime boundary including all three physical fixture lifecycle variants passed. Updated affected assertions to the adopted command and canonical-operation contracts. Follow-up validation is pending.
- Added actual Make rejection coverage for all frozen retired command names, with no retained preparation output. This supplements registry absence with executable admission behavior. Generation 14 is in progress; no pass claimed yet. W09 remains open; finalizer and W10/W11 are pending.

- Follow-up: generation 14–16 passed (outer 0). `remediation-w09-contract-20261017-08` passed (outer 0, 63,313 ms); it includes actual retired Make rejection and context-tampering checks. `remediation-w09-regression-20261017-04` ended outer 2 (256,249 ms), with only the task-surface negative fixture failing; native lifecycle passed again. The fixture now models an explicit synthetic non-public target rather than assuming a removed wrapper exists.
- `remediation-w09-lifecycle-20261017-01` ended outer 2 (8,662 ms): browser lifecycle fixture failed; service and development-stack fixtures passed. Diagnosis found inherited runtime identity contamination and a fabricated runtime-owner record. The fixture now restores parent identity after input-isolation checks, exercises valid admission with a native runtime created through its supported API. Borrowing trusts the exact owner-only runtime binding; an extra test expectation of live-process validation at borrow was incorrect and was removed before acceptance. Follow-up smoke, lifecycle and source-check runs are in progress.
- Caller review updated public-wrapper smoke selection files, UI recovery recommendations, Fallow setup guidance and the active UI refactor digest. Removed the unused standalone release-readiness reader and its obsolete public-name roster; the private graph marker retains exact package identity checks, while retained validation belongs to the existing observability reader. No runtime Markdown dependency was introduced.


### W09 admission and lifecycle boundary repairs — IN_PROGRESS (E-023)

- Generation 17–19 passed (outer 0). `remediation-w09-source-20261017-03` passed (outer 0, 16,343 ms) before the subsequent admission fixes. No final-source acceptance is inferred.
- Smoke runs 01–03 failed (outer 2). Diagnosis progressed from inherited Make selectors to an internal manifest override being read before canonical resolution, then raw inherited DETAIL being forwarded after preflight correctly ignored it. Fixed nested fixture input isolation, canonical preflight policy lookup, and Node dispatch forwarding of shared-resolver values. These are existing input-source contract repairs, not relaxed tests. Smoke 04 is pending.
- Lifecycle runs 02–03 failed (outer 2). The first included an incorrect extra test expectation about runtime borrowing, removed after inspecting native ownership boundaries. The remaining failure was the old synthetic artifact chain lacking real runtime detail authority. Split that chain into `tools/harness/browser/tests/web-e2e-artifact-fixture.sh` and run it while its native owner remains alive; its security, identity, publication, negative attachment and cleanup assertions remain. `remediation-w09-lifecycle-20261017-04` passed (outer 0, 7,526 ms).
- `remediation-w09-shell-20261017-01` failed (outer 2, 17,529 ms) on SC2153 for the intentionally sourced fixture's ROOT_DIR. Declared the fixture's sourced-variable context; shell 02 is pending. Finalizer, full contract follow-up and W10/W11 remain open.

- Additional E-023 validation: smoke 04 and 05 failed only task-guide's false JSON default being serialized as a flag. The dispatcher now serializes typed boolean defaults according to their declared input type; direct coverage verifies inherited selector exclusion and false-default omission. Generation 20–21 passed. `remediation-w09-smoke-20261017-06` passed (outer 0, 7,059 ms). `remediation-w09-shell-20261017-02` passed (outer 0, 17,487 ms).
- Current guidance review removed a stale performance publication paragraph and tests-only gate from the developer guide, corrected Fallow placement and consumer-driven build guarantees, and supplied required target-plan selectors. The duration drift taxonomy label is historical/reserved, not an active qualification gate. Historical handoffs remain historical.
- `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-20261017-01 make agent-finalize` started with RESULTS_DIR unset. Retained-run maintenance is intentionally unavailable: no applicable successful full warm check exists. Finalizer mutations/results and full contract follow-up must be reviewed before W09 completion.

- W09 finalizer 01 ended outer 2 (6,457 ms); its schema/catalog actions passed, but generated refresh rejected an inherited parent operation digest. Summary: `.cartulary/test-results/remediation-w09-finalize-20261017-01/unit-artifacts/finalize-summary.json`. It changed zero files, rollback was not needed, and RESULTS_DIR maintenance was skipped. Each maintenance child now binds its own canonical operation digest. Generation 22 passed; finalizer 02 is running with RESULTS_DIR unset.


### W09 finalizer, full-suite diagnosis and private-dispatch cleanup — IN_PROGRESS (E-024)

- `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-20261017-02 make agent-finalize`: outer 0, 25,774 ms. Summary `.cartulary/test-results/remediation-w09-finalize-20261017-02/unit-artifacts/finalize-summary.json`: zero generated/source mutations, rollback not needed. Retained-run maintenance was skipped because RESULTS_DIR was unset. Validated source digest `sha256:c96d0e7a67c950aa6387ffd15ff10850c18026fa4c025257b200f0ea743cd6bf`.
- Same executable-source checks: `remediation-w09-contract-20261017-09 make harness-contract` passed (outer 0, 61,186 ms); `remediation-w09-source-20261017-04 make source-check CHECK=all` passed (outer 0, 23,140 ms). Each prefix above means `CARTULARY_TEST_RUN_ID=<id>`. `remediation-w09-markdown-20261017-02 make lint-markdown` passed (outer 0, 12,830 ms), preceding the latest human-only Fallow policy clarification and handoff refresh.
- `CARTULARY_TEST_RUN_ID=remediation-w09-full-20261017-01 make harness-contract SUITE=full`: outer 2, 283,481 ms; 46/48 units passed, two failed, none skipped/cancelled; all 57 cleanup steps passed. Exact root `.cartulary/test-results/remediation-w09-full-20261017-01`. The make-node-tools fixture omitted the required HOME/cache-root input; its environment is repaired without relaxing production admission. Real target-output coverage failed in nested backend-unit execution at `tmp/op/r1729-backend-unit`: 204/205 units passed, native runtime row failed. Its child graph `native-engine/selected-graph-28316-off` inherited its enclosing operation digest because shell row compilation did not bind its own. Shell units now bind the selected operation digest at compilation; a regression verifies changed executable definitions change physical identity. Dispatch freshness rejection remains intact.
- Removed thirteen unreferenced predecessor operation definitions and the old smoke scheduling CLI; retained finalizer diagnostic operations and genuine file production. Public commands and harness presets use the existing graph compiler exclusively. No retired scheduler or arbitrary task-surface passthrough remains. Generation 23 is in progress; subsequent executable changes require a fresh finalizer/checkpoint.
- Observed HEAD is now `afd6c699980d4256af980a9abb906b2d79004755`, the user's `gitignore update` commit atop the recorded source baseline. Preserve it; this effort has made no commits. The starting commit remains the comparison baseline, not a claim that HEAD is unchanged.
- W09 remains open. W10/W11 have not started. No functional gate, release readiness or performance publication claim is made by these subset results.


### W09 verification routing migration — IN_PROGRESS (E-025)

- Generation 23 and 24 passed. `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-20261017-03 make agent-finalize`: outer 0, 26,075 ms, zero mutations, rollback not needed, RESULTS_DIR unset and retained maintenance skipped. Source digest `sha256:74e706969998a02a0f1454ab9c4915cebdf4c151f8abf4f3f5105b7b5bd05d4c`.
- `CARTULARY_TEST_RUN_ID=remediation-w09-source-20261017-05 make source-check CHECK=all`: outer 0, 20,014 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-markdown-20261017-03 make lint-markdown`: outer 0, 10,830 ms, before subsequent navigation and routing clarification edits.
- `CARTULARY_TEST_RUN_ID=remediation-w09-full-20261017-02 make harness-contract SUITE=full`: outer 2, 289,976 ms, 47/48 passed. Cache-root fixture and compiled shell-operation identity regressions passed; nested off/basic/basic-second executions passed. Remaining real-target output failure at `tmp/op/r16436-backend-unit` exposed a native private-diagnostic fixture still naming retired build-web, preventing its diagnostic from binding an active command. Updated that synthetic target to build; primary failure classification rules are unchanged. Installed native boundary and physical cleanup contexts passed.
- Human caller review found 102 stale public_target references across verification contracts. Migrated each name to its retained command. public_target remains a command reference, not a complete invocation; exact row selections remain catalog-owned and selectors remain task-surface-owned. The catalog now rejects references to retired/private/unknown public names before source selection; negative fixtures exercise that boundary. This repairs reference integrity without creating another selector registry or adding fields to the verification wire schema.
- Updated active UI digest navigation (Markdown, human TSV and human localization metadata) to current command grammar. Executable tooling excludes all docs, including these non-Markdown human files. Historical evidence text remains historical.
- Owner guidance: `make task-guide OWNER=harness.test_catalog DETAIL=rows`, outer 0. Generation 25 is pending. W09 remains open; W10/W11 have not started.


### W09 diagnostic contract closure — IN_PROGRESS (E-026)

- Generation 25 passed (outer 0, 10,800 ms). `CARTULARY_TEST_RUN_ID=remediation-w09-native-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime`: outer 0, 83,963 ms. Native diagnostics, fresh nested graphs and receipt checks passed.
- `CARTULARY_TEST_RUN_ID=remediation-w09-contract-20261017-10 make harness-contract`: outer 2, 137,502 ms, 9/10 units passed. The new public-target integrity negatives passed. One private shell-row diagnostic fixture inherited the enclosing row's operation digest; it now uses its own compiled unit environment, preserving definition freshness and intended artifact-failure assertions. Exact failure root `.cartulary/test-results/remediation-w09-contract-20261017-10`.
- Removed the unused predecessor owner-explanation CLI mode/function/schema and its extension dependency. task-guide v4 already supplies the selected coverage and row views. Reconciled the remaining normative schema/coverage wording before removal. Schema-attachment manifest v2 now identifies the canonical private validation operation `harness.json_shape_check`, separating its identity from public command spelling.
- `CARTULARY_TEST_RUN_ID=remediation-w09-generate-20261017-26 make generate`: outer 0, 10,620 ms. Authored schemas, owner references and generator input declarations were edited; generated validators and registries were regenerated through Make.
- `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-20261017-04 make agent-finalize` is running with RESULTS_DIR unset. Its results/mutations, full/default contract validation, source checks and Markdown lint remain required before W09 closes. W10/W11 have not started.


### W09 final-source contract checkpoint — IN_PROGRESS (E-027)

- `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-20261017-04 make agent-finalize`: outer 0, 25,108 ms; generated/source mutations zero, rollback not needed; retained maintenance skipped because RESULTS_DIR was unset. Source digest `sha256:d2e17b7e50f566d835519d358a6fbe1aa69248ec7ed3ba3dcb468423c87625bf`.
- On that executable source: `CARTULARY_TEST_RUN_ID=remediation-w09-contract-20261017-11 make harness-contract`: outer 0, 62,017 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-source-20261017-06 make source-check CHECK=all`: outer 0, 23,013 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-markdown-20261017-04 make lint-markdown`: outer 0, 12,090 ms. `make help ALL=1 JSON=1` returned 0 and exactly 62 unique public entries; `.gitignore` and `docs/domain.md` have no worktree diff.
- Full contract 03 is still running; independent installed-runtime and real-command output work remains active. No failures observed at the recorded checkpoint; no terminal pass claim. Its source remains frozen while running.
- Review identified unused predecessor target-explanation and fixture-discovery CLI paths, plus missing-tool diagnostics recommending retired Make names. Complete their removal/migration after the active full run drains. Preserve fixture aggregation assertions through the existing private report builder; exact-run public fixture reader coverage remains in observability tests. Do not preserve implicit newest-run fixture selection as compatibility behavior.
- W09 remains open. No W10/W11 execution has begun.


### W09 machine-output boundary and legacy reader cleanup — IN_PROGRESS (E-028)

- `CARTULARY_TEST_RUN_ID=remediation-w09-full-20261017-03 make harness-contract SUITE=full`: outer 2, 339,454 ms; 47/48 units passed. Exact root `.cartulary/test-results/remediation-w09-full-20261017-03`. Native runtime coverage passed; real-target machine output failed because successful Go readiness emitted a human line before the single JSON summary. This is an output-contract defect, not a product test failure.
- Reproduction `CARTULARY_TEST_RUN_ID=remediation-w09-machine-20261017-01 CARTULARY_OUTPUT_MODE=machine make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.embedded_asset_layout` retained two stdout lines in `/tmp/cartulary-w09-machine-01.stdout`, empty stderr in its sibling `.stderr`. The native session was unavailable after context recovery; no independent outer exit claim is made. The first line was `ok go: launcher=go1.27.1 effective=go1.27.2 source=selected`.
- Successful prerequisite readiness is now silent, including installed-only mode; explicit diagnose remains human-readable and resolve retains its executable-path output. Added readiness coverage for these boundaries. Real-target postvalidation now retains stdout/stderr and canonical run roots when it fails.
- Removed unused explain-target/task-guidance and fixture-report CLIs, including implicit newest-run selection. Fixture aggregation and no-double-counting assertions now exercise the retained report builder; current explain-run tests reject a parent directory without an exact run. Removed target-plan's unused private legacy flag. Updated missing-tool and drift diagnostics to current public commands.
- `CARTULARY_TEST_RUN_ID=remediation-w09-generate-20261017-27 make generate`: outer 0, 10,690 ms. Validation of the repairs is pending. W09 remains open; W10/W11 have not started.


### W09 follow-up candidate validation — IN_PROGRESS (E-029)

- Frozen executable source: `sha256:e1966d03f3737d0a24e376d9665024cc4bf8a2778d3f40dfb38bb5dec117df39`.
- `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-20261017-05 make agent-finalize`: outer 0, graph 25,849 ms. Summary `.cartulary/test-results/remediation-w09-finalize-20261017-05/unit-artifacts/finalize-summary.json`: zero mutations, rollback not needed; retained-run maintenance skipped because RESULTS_DIR was unset.
- `CARTULARY_TEST_RUN_ID=remediation-w09-source-20261017-07 make source-check CHECK=all`: outer 0, 23,661 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-markdown-20261017-05 make lint-markdown`: outer 0, 12,200 ms (before this human-only checkpoint). `git diff --check` and `make help ALL=1 JSON=1` both returned 0; help retained at `/tmp/cartulary-remediation-final-help.json`.
- Default contract 12 and full contract 04 are running on this frozen source. Their native sessions are 58701 and 30769 respectively; exact roots are `.cartulary/test-results/remediation-w09-contract-20261017-12` and `.cartulary/test-results/remediation-w09-full-20261017-04`. No pass is claimed before owning commands finish.


### W09 target-local JSON follow-up — IN_PROGRESS (E-030)

- Default contract 12 completed: outer 0, 61,984 ms on E-029 source. `make help ALL=1 JSON=1` returned exactly 62 unique public names.
- `CARTULARY_TEST_RUN_ID=remediation-w09-json-20261017-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.embedded_asset_layout JSON=1`: outer 0, 178,435 ms. Execution passed, but `/tmp/cartulary-w09-json-01.stdout` contained one human GRAPH line rather than JSON; sibling stderr was empty. Run manifest correctly recorded declared JSON=1, so the remaining defect is native public-result rendering rather than selector forwarding or execution selection.
- Planned repair after the active frozen run drains: return the resolved output choice with the native run result, use it at public rendering, and cover target-local JSON through actual Make execution. Do not re-read raw JSON environment values as authority.
- Full contract 04 currently has 47/48 units passed with no failures, and its real-command unit is still active. No final pass claimed. W09 remains open and W10/W11 have not begun.


### W09 resolved output and operation environment repairs — IN_PROGRESS (E-031)

- Full contract 04 completed: outer 2, 492,355 ms, 47/48 units passed; 57/57 cleanup steps passed. Exact root `.cartulary/test-results/remediation-w09-full-20261017-04`. The real-target unit passed backend human and machine output checks, then failed nested lint at `tmp/op/rd827-lint`: 12/13 units passed, frontend typecheck reported `config/configuration_error` with `spawnSync bash ENOENT`.
- Failure evidence is preserved: `tmp/tool-output-real-targets.9bT1zZ` and `tmp/op/rd827-*`. Diagnosis found the sole operation environment still using Make's `$$PATH` escape. Replaced it with the existing named `$(PATH)` binding; no shell evaluation or fallback dispatch was added.
- Native execution now returns published summary and validated target-local JSON presentation separately. Public rendering uses the resolved choice; wire summary schemas and physical operation identities are unchanged. Updated native API consumers and added actual Make JSON=1 coverage to real-target assurance.
- `CARTULARY_TEST_RUN_ID=remediation-w09-generate-20261017-28 make generate`: outer 0, 10,250 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-format-20261017-01 make format` rejected before preparation (outer 2): that retained root already existed. It was preserved. Follow-up formatting uses fresh ID `remediation-w09-format-json-render-20261017-01` and is pending. Finalizer and focused repairs must pass before broader revalidation.


### W09 focused repair confirmation — IN_PROGRESS (E-032)

- `CARTULARY_TEST_RUN_ID=remediation-w09-format-json-render-20261017-01 make format`: outer 0, 7,589 ms. Generated outputs were refreshed through Make generation 28; no hand-edit of generated paths.
- `CARTULARY_TEST_RUN_ID=remediation-w09-types-20261017-01 make lint CHECK=types`: outer 0, 20,608 ms, confirming private operation PATH binding reaches the real frontend toolchain.
- `CARTULARY_TEST_RUN_ID=remediation-w09-json-20261017-02 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.embedded_asset_layout JSON=1`: outer 0. `/tmp/cartulary-w09-json-02.stdout` contains exactly one passing `cartulary.harness_run_summary.v1` JSON object; sibling stderr is empty. Published run evidence lives at `.cartulary/test-results/remediation-w09-json-20261017-02`.
- `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-20261017-06 make agent-finalize`: outer 0, 25,618 ms, zero mutations, rollback not needed. Retained-run maintenance was skipped because RESULTS_DIR was unset. Frozen source: `sha256:f4328ae0b7f15c9a4a887457b77726dea699046999aa54ac52ee30deee974a15`.
- On that source, full contract 05 (session 13151), default contract 13 (95792), source-check all 08 (64994), and Markdown lint 06 (68781) are running. All IDs use prefix `remediation-w09-`, date `20261017`, and the same command grammar as E-029. Exact final outcomes remain pending; W09 is not complete.


### W09 frozen candidate follow-up results — IN_PROGRESS (E-033)

- Source remains `sha256:f4328ae0b7f15c9a4a887457b77726dea699046999aa54ac52ee30deee974a15` (E-032).
- `CARTULARY_TEST_RUN_ID=remediation-w09-contract-20261017-13 make harness-contract`: outer 0, 61,820 ms.
- `CARTULARY_TEST_RUN_ID=remediation-w09-source-20261017-08 make source-check CHECK=all`: outer 0, 22,436 ms.
- `CARTULARY_TEST_RUN_ID=remediation-w09-markdown-20261017-06 make lint-markdown`: outer 0, 12,460 ms, preceding the latest human-only evidence entries.
- `CARTULARY_TEST_RUN_ID=remediation-w09-full-20261017-05 make harness-contract SUITE=full` is active in session 13151. Exact root `.cartulary/test-results/remediation-w09-full-20261017-05`; last observation 45/48 passed, no failures. No terminal outcome or acceptance inferred from progress.

### W09 nested admission budget correction — IN_PROGRESS (E-034)

- Full contract 05 completed: outer Make 2, normalized timing/timeout_failure, 690,207 ms; 46/48 units passed. Native runtime boundary and real-target assurance both terminated with SIGTERM after their existing deadlines. No deadline was increased and no automatic retry was introduced.
- Nested evidence `.cartulary/test-results/remediation-w09-full-20261017-05/selected-service-boundary/pure` reports one failed unit and 298,334 ms. Its plan requested 19 CPU tokens although the enclosing runtime-boundary operation reserved one. The object-store selection also reported waiting for host admission. No row execution artifact was published for the timed-out pure row.
- Diagnosis against TH-HARNESS-REQ-102: the native run environment exported whole-host CPU capacity, while operation units did not replace it with their actual admitted claims. Concurrent parent operations could therefore retain their reservations while nested Go graphs requested whole-host capacity. This is an existing admission-contract repair, not a timeout relaxation.
- Native process dispatch now binds the actual lease and the unit's CPU claim after merging ambient, fixture and command inputs. Removed the whole-host budget export. Added a real subprocess/admission regression proving all three competing input channels cannot override the admitted lease or one-CPU budget. Focused validation is pending before broader revalidation.

### W09 admitted-budget validation candidate — IN_PROGRESS (E-035)

- `CARTULARY_TEST_RUN_ID=remediation-w09-format-admission-20261017-01 make format`: outer 0, 8,658 ms.
- `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-20261017-07 make agent-finalize`: outer 0, 25,513 ms. Summary `.cartulary/test-results/remediation-w09-finalize-20261017-07/unit-artifacts/finalize-summary.json`: zero mutations, rollback not needed. Retained maintenance skipped because RESULTS_DIR was unset.
- Frozen executable source: `sha256:006ecc5e8be73f38c1e0f18f96b1cda17f36beaa7a35a07c45abc32f4f134c9f`. Full contract 05's outer cleanup completed all 57 steps; its timeouts remain failed evidence.
- Active focused commands: `CARTULARY_TEST_RUN_ID=remediation-w09-native-20261017-02 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime` (session 7507); `CARTULARY_TEST_RUN_ID=remediation-w09-runtime-boundary-20261017-01 make harness-contract SUITE=runtime-boundary` (30730); `CARTULARY_TEST_RUN_ID=remediation-w09-source-20261017-09 make source-check CHECK=all` (24354). Exact roots are `.cartulary/test-results/<run-id>`; results are pending.

### W09 installed-runtime budget validation — IN_PROGRESS (E-036)

- `CARTULARY_TEST_RUN_ID=remediation-w09-runtime-boundary-20261017-01 make harness-contract SUITE=runtime-boundary`: outer 0, 241,688 ms. Installed PostgreSQL, HTTPS, pure, object-store and mixed lifecycle selections completed; exact root `.cartulary/test-results/remediation-w09-runtime-boundary-20261017-01`.
- `CARTULARY_TEST_RUN_ID=remediation-w09-source-20261017-09 make source-check CHECK=all`: outer 0, 16,737 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-markdown-20261017-07 make lint-markdown`: outer 0, 10,710 ms, before this human-only checkpoint.
- Native row 02 remains active in session 7507; it waited for the installed runtime's reservation and is now eligible to proceed. `CARTULARY_TEST_RUN_ID=remediation-w09-contract-20261017-14 make harness-contract` also started to validate downstream consumers under the corrected per-operation budget. W09 remains open until the affected assurance passes.

### W09 admitted-budget focused closure — IN_PROGRESS (E-037)

- `CARTULARY_TEST_RUN_ID=remediation-w09-native-20261017-02 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime`: outer 0, 284,169 ms including host-admission waiting. This exercises the new real subprocess/admission override regression and existing runtime assertions without cache reuse.
- E-036's installed-runtime pass is supported by `.cartulary/test-results/remediation-w09-runtime-boundary-20261017-01/selected-service-boundary/pure/execution-context.json`: the selected Go unit claims one CPU, matching its actual parent reservation. The physical host capability snapshot remains unchanged.
- `CARTULARY_TEST_RUN_ID=remediation-w09-full-20261017-06 make harness-contract SUITE=full` started on E-035 source to validate concurrent assurance; native session 45338, exact root `.cartulary/test-results/remediation-w09-full-20261017-06`. `CARTULARY_TEST_RUN_ID=remediation-w09-contract-20261017-14 make harness-contract` passed (outer 0, 62,601 ms). All focused repairs pass; W09 stays open until the remaining full-suite outcome is known.

### W09 real-command assertion isolation — IN_PROGRESS (E-038)

- Full contract 06 completed: outer Make 2, 653,893 ms; 47/48 units passed. The native runtime boundary passed, including the formerly timing-out pure selection and object-store/mixed cases. Only real-command output assurance timed out at its unchanged ten-minute unit deadline.
- Source-matched diagnostic inventory (not selected by recency): `tmp/op/r15b51-*` matches E-035 source and contains passing schema, backend human/machine, owner JSON, lint, production build human/machine and test-fast summaries. Durations respectively 6,602; 58,429; 30,170; 81; 23,743; 6,712; 6,145; 338,046 ms. This inventory supports cumulative-cost diagnosis; the old helper did not record a parent-root link or independently recoverable outer exits for interrupted work. It does not establish current check/release acceptance.
- Replaced the monolithic sequential assertion operation with eleven canonical cases in the existing operation owner and one private union. The full preset retains source-schema, backend human/machine, owner JSON, lint, production build human/machine, test-fast, check human/machine, and invalid-usage assertions. Each case uses the existing unit deadline; no timeout was increased, no assertion removed, and no scheduler or public command added.
- Every case now retains its stdout/stderr and nested run roots beneath its owning unit artifact directory and emits an explicit diagnostic pointer. Interruption no longer leaves an unlinked global temporary directory. Runtime/run retention owns cleanup. Added graph coverage for the exact case set and distinct operation identities. Generation, finalization and validation are pending.

### W09 isolated-case generation and finalizer evidence — IN_PROGRESS (E-039)

- `CARTULARY_TEST_RUN_ID=remediation-w09-generate-20261017-29 make generate`: outer 0, 10,640 ms. `git diff --check`: outer 0. Outputs were regenerated through Make.
- Finalizer 08 has a terminal pass summary, 25,649 ms, source `sha256:588376e3702414314161a03fcdc83690a93de622e93da473da9e92ee08ea7e95`; its native outer exit was unavailable after context truncation. It changed zero files. This is recorded as reported success, not independently observed command success.
- To establish the required owning-command result, `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-20261017-09 make agent-finalize` completed with outer 0, 25,433 ms. RESULTS_DIR was unset; retained-run maintenance was skipped. Review confirms zero generated mutations and no rollback. Exact roots for both finalizers are `.cartulary/test-results/<exact-run-id>`.
- Source all 10 (session 14130) and default contract 15 (34883) are active. Shell lint with ID `remediation-w09-shell-20261017-01` rejected before preparation (outer 2) because that retained root already exists; it was not cleared. Shell lint was launched with a distinct identity, `remediation-w09-shell-isolated-cases-20261017-01`. Source all 10 passed (outer 0, 20,057 ms); default contract 15 passed (outer 0, 57,845 ms); isolated-case shell lint passed (outer 0, 18,728 ms). Markdown lint 08 passed (outer 0, 29,070 ms). Full suite 07 is now active with the eleven isolated cases. W09 remains open.

### W09 human guidance and retirement review — IN_PROGRESS (E-040)

- Current generated registry contains exactly the adopted 62 names. Default contract 15 includes actual GNU Make rejection of the frozen retired public/private names, with empty retained-preparation output. Active README, root guidance, guides and repository skills contain no removed performance commands, `make help-all`, predecessor unit commands or ROLE-based task guidance.
- Human specification review found three stale observer-qualification references in the live-status/foundation/instrumentation clauses. Removed the remaining obligation to measure concurrent reader overhead and the obsolete qualification-evidence cross-reference. Retained the explicit warning that read-only observation can affect execution; diagnostic integrity and privacy remain required. This finishes the user's qualification-removal amendment without changing executable inputs.
- Clarified each package-inspection and build-command destination in the 190-entry migration table. The complete table remains pending W09 closure, distinct from command availability validation. Full suite 07 is active with 58 units; no final result is claimed.

### W09 nested admission investigation — IN_PROGRESS (E-041)

- Full suite 07 remains active; independent work and cleanup are allowed to drain. At revision 174, 37 units passed, three failed (backend human/machine: config/configuration_error; check human: timing/timeout_failure), and remaining work was still running or pending. These are intermediate observations, not the terminal tally. Its exact nested backend human run is `unit-artifacts/operation-harness.assurance_output_backend_unit/real-command-output/runs/real-backend-unit` beneath the parent root. Public status reported host-admission waiting and failed native runtime/session rows; no full-suite terminal result is claimed.
- Published native runtime unit result reports exit 10, `product/test_assertion_failure`. Bounded stderr identifies `TestRuntimeSelectedGraphExecution/off`: its inner graph returned 130 after the existing 90-second context deadline. The exact retained child is `native-engine/selected-graph-6240-off` beneath that backend root. This differs from simply exhausting the outer ten-minute case deadline.
- The second native-session failure is independent: its public-file negative fixture creates mode 0644 but does not chmod after creation. The secure enclosing umask 077 changes the fixture to 0600, so it never reaches the intended unsafe-permission boundary. Fix the fixture explicitly, preserving secure output permissions and production rejection.
- Source inspection found root-wide FIFO blocking for all descendants in the host-admission tree. A queued reservation can therefore bar a deeper child of an admitted resource holder needed for that holder to finish. This is a causal hypothesis pending a deterministic nested-tree regression; no arbiter mutation or timeout increase has been made while the full run is active.
- Obtained current guidance with `make task-guide OWNER=harness.evidence_accounting DETAIL=rows`, outer 0. The exact native admission row is `harness.evidence_accounting.behavior.observation_admission`. Review will preserve weighted parent/child credit, capacity ceilings, quiet-activity fairness and lease ownership.
- Markdown retirement lint 01 passed (outer 0, 13,630 ms). Subsequent documentation-only clarification makes target-plan's required TARGET/INPUTS and exact retained-run examples explicit; final Markdown coverage remains required.

### W09 deterministic admission repair — IN_PROGRESS (E-042)

- Full contract 07 drained with outer Make 2, 1,029,871 ms: 53/58 units passed, five failed. The parent normalized result was config/configuration_error (the output assertion forwarded a child Make exit 2); exact nested evidence retains native product assertions and service preparation failures. Backend human/machine and test-fast assertions failed; human/machine check exceeded their outer deadlines. No passing acceptance is claimed.
- Nested backend human: 203/205 passed, 343,387 ms. Machine: 201/205 passed, 466,560 ms; additional `harness.browser.unit.fixture_tls` and `harness.browser.unit.object_store_fixture_admission` failed preparation with infra/service_readiness_timeout and no runner log. Those readiness outcomes require focused follow-up. Nested test-fast: 581/582 passed, 154,611 ms; only native_session failed its unsafe-permission fixture. These are distinct child contexts, not W10 gate evidence.
- `CARTULARY_TEST_RUN_ID=remediation-w09-admission-characterization-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission`: outer 2, normalized 10/product/test_assertion_failure, 507 ms. Two added deterministic cases fail at the intended boundary: a queued ancestor sibling and a queued request in another admitted branch each block a child needed to drain existing credit. Existing quiet-ordering/state tests pass.
- Fixed FIFO scope to siblings under the same admitted parent; weighted tree capacity, independent root activity ordering, quiet admission and owner lifetime checks remain unchanged. Added multi-level draining, sibling FIFO and capacity exhaustion assertions within the existing cataloged test. Clarified REQ-381's hierarchical entitlement. This is a structural admission repair, not a new scheduler.
- Fixed the unsafe public-file fixture with explicit chmod after secure creation, so umask 077 cannot sanitize the negative case. Output assertions now return assertion failure after retaining a failed child Make outcome, avoiding misclassifying the wrapper's generic 2 as malformed test inputs. No command outcome is hidden or relabeled inside child evidence.
- Historical W00 cold check took 679,148 ms, exceeding the generic 600,000 ms output-assertion envelope. Aggregate-case budget suitability remains under review after the causal admission repair; no deadline has been increased. Formatting and focused validation are pending.

### W09 admission repair focused evidence — IN_PROGRESS (E-043)

- `CARTULARY_TEST_RUN_ID=remediation-w09-format-nested-fifo-20261017-01 make format`: outer 0, 7,543 ms.
- Under explicit `umask 077`, `CARTULARY_TEST_RUN_ID=remediation-w09-admission-session-fix-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission,harness.evidence_accounting.behavior.native_session`: outer 0, 1,183 ms. Both deterministic admission cases and the actual unsafe-permission negative fixture pass with fresh execution.
- `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-nested-fifo-20261017-01 make agent-finalize`: outer 0, 25,060 ms. Source `sha256:4f4ccc309367b53fff42801d12218fc0a84702533f538dc0e650845908f4a315`; zero generated mutations, no rollback. RESULTS_DIR unset, retained-run maintenance skipped. `git diff --check`: outer 0.
- `make explain-run RESULTS_DIR=<full07/backend-unit-machine exact child> DETAIL=fixtures` rejected incomplete historical evidence (outer 2): the failed preparation did not publish fixture_tls row evidence. No reader pass is claimed; no fallback or retained artifact repair was attempted. Both failed service preparations released their host reservations according to cleanup-results.json, and neither launched its Go runner. Source identifies object-store proxy readiness as the typed failure boundary.
- Fresh focused follow-ups started on the finalized source: readiness-followup 01 selects the exact fixture_tls/object_store_fixture_admission rows with cache off (session 9034); native-nested-fifo 01 selects native_runtime with cache off (73153); contract-nested-fifo 01 selects the default contract (2887); source-nested-fifo 01 selects CHECK=all (69186); markdown-nested-fifo 01 runs lint-markdown (5881). All IDs have prefix `remediation-w09-` and date `20261017`; exact roots are `.cartulary/test-results/<exact-run-id>`. Native runtime passed (outer 0, 102,747 ms); readiness follow-up passed (outer 0, 109,914 ms including host waiting); source all passed (outer 0, 21,631 ms); Markdown passed (outer 0, 11,450 ms). Default contract passed (outer 0, 117,982 ms including host waiting). A second exact readiness run, `remediation-w09-readiness-secure-umask-20261017-01`, uses explicit umask 077 and cache off to reproduce the enclosing output-case permission context rather than infer equivalence from the default-umask pass; it passed (outer 0, 7,319 ms).

### W09 repaired concurrent assurance candidate — IN_PROGRESS (E-044)

- All focused E-043 outcomes pass. `CARTULARY_TEST_RUN_ID=remediation-w09-full-nested-fifo-20261017-01 make harness-contract SUITE=full` now validates the isolated cases together on source `sha256:4f4ccc309367b53fff42801d12218fc0a84702533f538dc0e650845908f4a315`. Full-suite outcome remains pending; W09 is open, and W10/W11 have not begun.

### W09 nested fixture follow-up — IN_PROGRESS (E-045)

- Full nested-fifo 01 remains active in native session 15062. Its backend human child now completes `native-engine/selected-graph-44452-off` in 13,607 ms with one passed and one failed unit, rather than expiring its 90-second context. The failed evidence row reports product/test_assertion_failure (87/88 Node assertions passed).
- Bounded child stderr identifies `tools/harness/tests/test-agent-finalize.mjs:87`: mode 448 (0700) differs from expected 488 (0750). The fixture requested creation mode 0750 under umask 077 and never applied it explicitly. Production rollback correctly preserved the starting 0700; the assertion expected an uncreated mode. Repair will chmod the fixture before snapshotting. A targeted source review of permission assertions found this additional explicit-mode dependency; secure output checks expect 0600/0700 and remain valid.
- Executable source remains frozen while the full run drains. No repeated full run is authorized by this observation alone; apply the concrete fixture repair and pass the exact current_epoch_evidence row under umask 077 first. W09 remains open; W10/W11 have not begun.

### W09 aggregate assertion deadline decision — IN_PROGRESS (E-046)

- Full nested-fifo 01's human check assertion reached its 600,000 ms outer deadline while the exact child graph was progressing without reported failures (revision 399 showed 506/800 passed, 18 active, no failed units). This follows the 679,148 ms original cold check and the prior isolated-case timeout. The generic single-operation envelope is unsuitable for an assertion that executes an entire gate; this evidence does not justify changing product or service deadlines.
- After the active run drains, add an optional positive `timeout_ms` to the existing work-graph policy-unit model, compile it into the existing unit timeout and semantic identity, and assign 1,800,000 ms only to the human/machine full-check output assertions. The owner default remains 600,000 ms. Validate declared/default deadlines, invalid values and identity changes before another full run. This adds no scheduler, public selector, retry or disabled assertion.
- Also apply E-045's explicit-mode fixture correction, then run current_epoch_evidence with cache off and umask 077. Generation/finalization and relevant source/contracts must pass before broader validation. W09 stays open.

### W09 aggregate deadline and mode fixture implementation — IN_PROGRESS (E-047)

- Full nested-fifo 01 ended outer 2, normalized product/test_assertion_failure, 963,285 ms: 54/58 passed. Backend human/machine each passed 204/205, with only the finalizer fixture failure (295,064/294,534 ms); their service-readiness rows passed in the concurrent secure context. Both full-check output assertions timed out. Test-fast passed all 582 units in 198,461 ms. Child roots remain under each owning output unit, as in E-038; these are boundary-context observations, not W10 acceptance.
- Applied explicit chmod 0750 to the finalizer fixture before its snapshot. Added optional `timeout_ms` to the existing authored policy-unit schema with the existing canonical unit bounds (1–3,600,000 ms), and resolved it in the existing compiler. Only the two check output assertions declare 1,800,000 ms. Default remains 600,000 ms; child, service and cleanup deadlines are unchanged. Canonical wire units already carry timeout; this is part of the same unpublished v4 owner cutover.
- Added behavioral graph coverage for explicit/default deadlines, unchanged operation/command identity, changed semantic identity, and rejected null/zero/negative/fractional/string/out-of-bounds values. REQ-801 now states ownership and validation of the outer aggregate assertion envelope.
- `CARTULARY_TEST_RUN_ID=remediation-w09-generate-aggregate-deadline-20261017-01 make generate`: outer 0, 10,200 ms. `git diff --check`: 0. `CARTULARY_TEST_RUN_ID=remediation-w09-format-aggregate-deadline-20261017-01 make format`: outer 0, 7,498 ms. Finalizer aggregate-deadline 01 is active with RESULTS_DIR unset. Focused secure-umask evidence, default contracts, source checks and full follow-up remain required before W09 can close.

### W09 finalizer permission preservation repair — IN_PROGRESS (E-048)

- Finalizer aggregate-deadline 01 passed (outer 0, 24,896 ms), zero mutations/no rollback, RESULTS_DIR maintenance skipped. Its source was `sha256:d8c8122ab3e3010d5eb7250fcc4dc9d095859835f182a43e176220300a361329`; subsequent implementation changes invalidate its final-source applicability.
- With the fixture now explicitly 0750, `umask 077` plus `CARTULARY_TEST_RUN_ID=remediation-w09-finalizer-mode-characterization-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence` failed at the intended rollback boundary (outer 2, normalized 10, 10,210 ms; test-agent-finalize.mjs:88, actual 0700 versus saved 0750). E-045 correctly identified the original invalid fixture, but that fixture had also concealed an implementation defect.
- Repair: restore regular files through a newly created O_EXCL/O_NOFOLLOW descriptor, write contents securely, then fchmod that owned inode to the captured mode before closing. This preserves source permissions independently of the caller's umask and does not follow a replacement path. Existing primary-error/rollback-failure handling remains in force. W05's original checkpoint is historical; C-005's new permission-preservation regression is being closed within W09 before progression.
- The same exact secure-umask/cache-off row is active under fresh ID `remediation-w09-finalizer-mode-fix-20261017-01`, native session 43676. No pass claimed yet. Finalizer, source/contracts and full assurance must be validated after the repair.

### W09 permission-preserving candidate — IN_PROGRESS (E-049)

- `umask 077` plus `CARTULARY_TEST_RUN_ID=remediation-w09-finalizer-mode-fix-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence`: outer 0, 10,104 ms. The corrected fixture now verifies actual permission preservation, freshness/rollback behavior and the remaining evidence assertions.
- `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-mode-fix-20261017-01 make agent-finalize`: outer 0, 24,915 ms; source `sha256:8aca9e6eb49e4e43538a887f877867073ec96c2e4eecdc122f61d99ea8244f27`. Zero generated mutations/no rollback. RESULTS_DIR unset; retained-run maintenance skipped. `git diff --check`: outer 0.
- On this candidate, contract-mode-fix 01 (session 28507), source-mode-fix 01 CHECK=all (82458), script-lint-mode-fix 01 CHECK=scripts (99067), and markdown-mode-fix 01 (45311) are active. All have prefix `remediation-w09-`, date `20261017`, and roots `.cartulary/test-results/<exact-run-id>`. Source all passed (outer 0, 22,767 ms); script lint passed (outer 0, 543 ms); Markdown lint passed (outer 0, 26,960 ms); default contract passed (outer 0, 58,504 ms), including explicit deadline identity and rejection coverage. Full concurrent assurance is now justified; W09 remains open.

### W09 finalized full-suite candidate — IN_PROGRESS (E-050)

- `CARTULARY_TEST_RUN_ID=remediation-w09-full-mode-fix-20261017-01 make harness-contract SUITE=full` is active in native session 71574 on E-049 source `sha256:8aca9e6eb49e4e43538a887f877867073ec96c2e4eecdc122f61d99ea8244f27`. All focused/source/lint/finalizer checks pass. At revision 120, the full run has 37 passed, zero failed, 18 pending and three active units. Both backend output cases and the installed native runtime boundary passed; check human/machine and test-fast remain active. These intermediate results do not establish terminal full-suite success. W10/W11 have not begun.

### W09 full-suite failure triage — IN_PROGRESS (E-051)

- Full mode-fix 01 published 55/58 passed, three failed, 1,425,590 ms; primary product/test_assertion_failure. Native session 71574 subsequently completed with outer Make exit 2, normalized exit 10. No W09 acceptance is claimed.
- Exact human check child: `unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check` beneath E-050 parent; 798/800 passed, 1,102,955 ms. All product-suite projections passed. Generated-source row failed with drift; Go vulnerability operation failed config/configuration_error before scanning: `missing operation binding GOVULNCHECK_DB`. Machine check child at the corresponding `operation-harness.assurance_output_check_machine/.../real-check-machine` root passed 797/800 in 999,729 ms, with the same two failures plus the platform.jobs execution-migration/supervised-runner unit. Execution migration passed. The supervised-runner concurrency subcase observed maximum 5 versus expected 8; its fixture uses a 120 ms lease and 30 ms renewal cadence. Cause remains under investigation; do not change production policy or weaken assertions.
- Test-fast child at `unit-artifacts/operation-harness.assurance_output_test_fast/real-command-output/runs/real-test-fast` passed 581/582 in 354,368 ms. The entity-merge planning Vitest unit failed artifact/artifact_error before row evidence: private child capture stage=child_environment, native=transport, code=unclassified. Existing safe diagnostics omit subprocess exit/signal and protocol-stage distinctions, so the cause remains unresolved; do not infer a product assertion failure or retry success from this wrapper outcome.
- Source inspection confirms GOVULNCHECK_DB is optional, stripped from aggregate public inputs, yet redundantly demanded by the operation environment template. The script already owns its empty/default behavior. Remove the redundant binding and characterize direct/default dispatch. Generated drift compares exact modes while several generators create files through umask-filtered modes; explicit generated-source publication semantics require a focused secure-umask regression and repair. Retained artifact permissions must remain owner-only.
- Continue bounded diagnosis and wait for the owning full command before changing executable inputs. W09 remains IN_PROGRESS; W10/W11 have not begun.

### W09 default bindings, source modes and bounded transport diagnostics — IN_PROGRESS (E-052)

- Characterization: `CARTULARY_TEST_RUN_ID=remediation-w09-default-mode-characterization-20261017-01 make harness-contract`, outer 2, normalized 10, 56,024 ms. Exactly the two added boundaries failed: optional GOVULNCHECK_DB wrongly required, and a newly rendered source file published 0600 under umask 077 instead of required 0644.
- Removed the redundant GOVULNCHECK_DB environment echo from the canonical operation. The scanner retains its omitted default and an explicitly admitted direct-command database; aggregate selectors cannot inject the direct-only input. Added default/explicit dispatch assertions.
- Existing generated publication now normalizes its rendered source modes before publication. Make generation normalizes only the declared generated-source inventory to 0644, preserving unchanged inodes and rejecting links/nonregular files and human-document paths. Evidence/scratch files retain their separate secure modes. Added secure-mask, mode-only correction, inode preservation, evidence exclusion and rejection checks.
- Native private capture now publishes only closed transport phase/exit/signal diagnostics in addition to the existing bounded categories. Fault-injection cases cover signal death, empty response, identity mismatch and inconsistent outcome, with no retries and no private payload disclosure. This improves diagnosis; it does not establish the cause or resolution of E-051's original transport failure.
- `CARTULARY_TEST_RUN_ID=remediation-w09-generate-source-modes-20261017-01 make generate`: outer 0, 10,170 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-format-source-modes-20261017-01 make format`: outer 0, 7,602 ms. `git diff --check`: 0. Finalizer-source-modes 01 active, RESULTS_DIR unset. Focused contracts, secure-umask source validation, exact failed-row follow-ups and concurrent assurance remain required.
- Current guidance obtained for harness.generated_artifacts, harness.command_surface, harness.evidence_accounting, platform.jobs and web.workbook, outer 0. No jobs production code or assertion has been changed. W09 stays open.

### W09 source-mode validation and publication review — IN_PROGRESS (E-053)

- Finalizer-source-modes 01 failed before mutation (outer 2, 3,078 ms): a negative fixture's literal human-root path was correctly rejected by executable-input admission. Constructed that rejection-only fixture path, preserving the prohibition on runtime Markdown dependency. Finalizer-source-modes 02 passed (outer 0, 25,200 ms), no generated mutations/rollback, RESULTS_DIR unset. Source was `sha256:fc8891cf2586d3a4ac35950f241cd3c3faa98d3af745cd7343d7911663eeba04`.
- Under umask 077: contract-source-modes 01 passed (outer 0, 58,953 ms); source-secure-modes 01 CHECK=all passed (0, 23,576 ms); capture-followup 01, cache off, exact entity_merge_planning_contract row passed (0, 3,214 ms); jobs-followup 01, cache off, exact execution_migration/supervised_runner rows passed (0, 24,796 ms). IDs have prefix `remediation-w09-` and date `20261017`; roots `.cartulary/test-results/<run-id>`. These focused successes do not explain or erase the earlier concurrent failures.
- `remediation-w09-vuln-default-20261017-01 make go-vulncheck`: outer 0, 4,699 ms. `remediation-w09-lint-source-modes-20261017-01 make lint CHECK=scripts`: 0, 642 ms. Run IDs were supplied via CARTULARY_TEST_RUN_ID. `git diff --check`: 0.
- Publication review adds rejection of nonregular/hard-linked files before opening and symlinked ancestor directories before mode effects; regression fixtures prove private evidence modes remain unchanged. This changes the candidate after the above passes. Finalizer-source-modes 03 then correctly rejected a stale topology input before mutation (outer 2, 4,153 ms); regenerate before finalization. Generate-source-guards 01 is active. Additional source guards and final-candidate assurance remain pending.
- Removed the remaining AC-007 comparative wall-time/peak-reduction qualification language under the user's performance amendment. Required exact ownership, eager per-test release, cleanup, failure and recovery assertions remain. No performance reference or publication evidence is required or claimed.

### W09 publication-guard candidate — IN_PROGRESS (E-054)

- `CARTULARY_TEST_RUN_ID=remediation-w09-generate-source-guards-20261017-01 make generate`: outer 0, 10,100 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-finalize-source-guards-20261017-01 make agent-finalize`: outer 0, 25,199 ms; zero generated mutations/no rollback. RESULTS_DIR unset, retained-run maintenance skipped.
- Candidate source: `sha256:2e127a77d2f2cc0ff81b64ae1c28e1b0367d92f57b07d8d41c730d5e651bd5d1`. Active validation IDs all start `remediation-w09-source-guards-`, end `-20261017-01`, and use `.cartulary/test-results/<run-id>`. `evidence` (session 66317) runs cache-off exact current_epoch_evidence under umask 077; `all` (39388) runs source-check CHECK=all under umask 077; `lint` (18190) runs lint CHECK=scripts; `markdown` (44323) runs lint-markdown. All four owning commands passed: evidence 10,807 ms; source all 17,322 ms; scripts 554 ms; Markdown 23,570 ms; each outer 0.
- The preceding default contract and exact jobs/frontend follow-ups passed on E-053 source. Only source-publication rejection guards changed afterward; their focused evidence and source checks now pass. W09 stays open, with W10/W11 not started.

### W09 full publication-guard assurance — IN_PROGRESS (E-055)

- `CARTULARY_TEST_RUN_ID=remediation-w09-full-source-guards-20261017-01 make harness-contract SUITE=full` started in native session 15632 on E-054 candidate after all focused checks passed. Exact root: `.cartulary/test-results/remediation-w09-full-source-guards-20261017-01`. Terminal outcome is pending. This validates concrete default-binding/source-mode repairs plus the unchanged exact jobs/frontend contexts; previous failed runs remain retained. W09 is not complete and W10/W11 have not begun.

### W09 final runtime cleanup investigation — IN_PROGRESS (E-056)

- Full source-guards 01 published 57/58 passed, one failed, 1,390,665 ms. Parent primary is product/test_assertion_failure because the human-output assertion rejected its nested failed command. Native session 15632 completed: outer Make exit 2, normalized parent exit 10; terminal scanning finished.
- Human check's exact nested run is `unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check` beneath E-055 root. All 800 units passed in 998,813 ms, but the run failed harness/cleanup_error (child normalized exit 12, Make 2). Cleanup publication contains 929 steps: 928 completed; only final runtime_close failed. Fixture providers/services completed. The repaired source/vulnerability checks and jobs row all passed. Machine check passed all 800 units in 966,301 ms with no failure. Test-fast, both backend output contexts and installed-runtime assurance passed.
- `make explain-run RESULTS_DIR=<exact human child> DETAIL=fixtures` returned 0 and described 137 leases. This is successful diagnosis only, not cleanup acceptance. Bounded parent stderr preserves the nested cleanup failure; no product assertion failed inside that check.
- Read-only inspection selected one private runtime by exact repository AND full run_root binding (never by recency or process name), without printing private paths, tokens, credentials or process identities. Its authority is fenced, detail storage remains, the resource journal is empty, and all 1,065 recorded producer identities are currently absent (1,061 group proofs). These observations narrow failure to producer/detail settlement; they do not reveal the original returned error or authorize deletion. No recovery or retained-evidence mutation was attempted.
- Source inspection found that detail cleanup enumerates procfs separately for each historical producer group within one shared 30-second helper deadline, then performs marked-descendant settlement. This is a scalability/diagnostic concern, not yet an established cause. Preserve fail-closed ownership, orphan/leaderless-group handling, deadlines and private-data boundaries while characterizing any repair. Current canonical cleanup evidence loses the native stage/cause, so bounded diagnostics are also needed before another broad attempt. W09 remains open; W10/W11 have not begun.

### W09 bounded native cleanup diagnosis — IN_PROGRESS (E-057)

- E-056 parent finished with Make 2 / normalized 10. The canonical exact-child `explain-run DETAIL=performance JSON=1` returned 0; its 2,842 ms outside selected phases and cleanup_error classification do not support the hypothesized 30-second settlement timeout. This view's pass means successful diagnosis, not a passing check. No deadline or ownership rule was changed.
- Added closed terminal runtime-cleanup stage/cause/errno diagnostics. Native error unwrapping and primary failure classification are preserved; no PID, proof file, private path, token or arbitrary message is exposed. Process discovery, identity, group anchor, signal, detail descendants/removal/publication can now be distinguished. This is an observability repair, not a claimed repair of the unknown E-056 physical failure.
- Corrected two residual normative references: runtime plan v3 and private browser-design-review lifecycle assurance. No public alias was restored.
- `CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-diagnostic-format-20261017-01 make format`: outer 0, 7,538 ms. `...=remediation-w09-cleanup-diagnostic-generate-20261017-01 make generate`: outer 0, 10,360 ms. Roots follow `.cartulary/test-results/<exact run ID>`.
- Under umask 077, `CARTULARY_HARNESS_CACHE_MODE=off CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-diagnostic-session-20261017-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_session`: outer 0, 1,170 ms. The added regression verifies closed output, nested boundary selection, native classification preservation, successful-cleanup silence and rejection of untrusted stage metadata.
- `...=remediation-w09-cleanup-diagnostic-finalize-20261017-01 make agent-finalize`: outer 0, 25,556 ms; zero generated mutations and no rollback. Retained-run maintenance was skipped because RESULTS_DIR was unset. `...=remediation-w09-cleanup-diagnostic-source-20261017-01 make source-check CHECK=all`: outer 0, 15,874 ms.
- `...=remediation-w09-cleanup-diagnostic-markdown-20261017-01 make lint-markdown`: outer 0, 12,580 ms. After subsequent guide/schema-reference repairs, `...=remediation-w09-guidance-review-markdown-20261017-01 make lint-markdown`: outer 0, 49,770 ms. A final guide INPUTS example correction followed; final Markdown validation remains required.
- Migrated the remaining browser design guide smoke commands to lifecycle/UI-review presets, preserving both seeded profile descriptions. Further human guidance review corrected schema-check references in projection/Network Flow specifications, a retired service selector in the UI refactor digest, and duplicate/retired bootstrap service commands. These prose edits do not change executable source identity.
- Deferred until the current frozen run completes: remove the stale agent-finalize side-effect description claiming duration-baseline writes when RESULTS_DIR is supplied. Runtime publication was already retired, but active help still projects this incorrect owner description. This is a concrete remaining W09 caller/discovery repair.
- Started `CARTULARY_TEST_RUN_ID=remediation-w09-full-cleanup-diagnostic-20261017-01 make harness-contract SUITE=full` to exercise the revised diagnostic boundary in the same concurrent assurance context. Native session 52865; source `sha256:eeb7140ea9b4b5582ce7c460980376d9b83a798bb0d5298b949592f11f09a6c5`. Outcome pending. Original failed runtime/evidence remain unchanged. W09 remains IN_PROGRESS; no W10/W11 acceptance is claimed.

- Additional bounded exact-owner inspection found three repeated producer IDs, each with two distinct start identities; one pair also has different sessions. No identity values or private paths were printed. This establishes real PID reuse within the failed invocation, but not the original returned error. The pending diagnostic candidate can distinguish a process-identity/group-anchor rejection from other cleanup causes.
- Baseline/candidate machine-owner comparison confirms 121 public + 69 internal names originally, 62 public currently, 61 retained originals, 129 retired originals, and one new name (`source-check`). Every retired original is present in the negative-dispatch fixture; none remains active. Counts are migration evidence, not execution/cleanup acceptance; disposition closure remains pending W09.

- Historical release root `remediation-final-release-20261017-04` was inspected read-only: `module.imports.measurement.xlsx_discovery` row and browser-group result passed; its graph unit failed on lease-000049 fixture release, followed by measurement-summary artifact failure, provider-close failure and blocked services-close. Bounded row/group/lease/cleanup evidence contains no precise physical cause. This is not a product assertion failure or current release acceptance. Current diagnostic guide wire versions were corrected and deferred observer qualification wording removed.

### W09 large-run assertion-reader failure — IN_PROGRESS (E-058)

- The E-057 full suite's human-output nested check passed all 800 units, 999,677 ms, with 923/923 cleanup steps completed. The parent output assertion then failed product/test_assertion_failure, exit 1: `test-tool-output-real-targets.sh` loads all `unit-events.ndjson` bytes as one UTF-8 string, exceeding Node's string limit (`ERR_STRING_TOO_LONG`). The final parent result is 56 passed / 2 failed, 1,421,708 ms, outer Make 2 / normalized 10; native session 52865 finished. The machine child passed 797/800 units in 963,026 ms, with three failed rows, all from PostgreSQL force-drop cleanup deadline errors within 20:08:33–35 UTC. Their classification remains the published product/test_assertion_failure, although the log cause is fixture cleanup. No retry or broader run has started.
- Exact parent root: `.cartulary/test-results/remediation-w09-full-cleanup-diagnostic-20261017-01`. Bounded logs: `unit-logs/operation-harness.assurance_output_check/{stderr,stdout}.log`; nested run remains below `unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check`.
- The human event stream is 546,824,152 bytes. This is a confirmed assertion-reader scalability defect, distinct from E-056 runtime cleanup. The existing canonical `readCanonicalUnitEvents` owner already streams, validates schema/sequence/time and rejects unsafe paths or concurrent mutation. Route this assertion through that owner instead of whole-file string buffering. No new event reducer or relaxed assertion is needed.
- The E-056 original native cleanup cause is still unknown. Later passing cleanup is not proof of repair. Real repeated producer identities were independently observed; review retirement at the existing verified physical settlement boundary before asserting that the lifecycle concern is closed.
- Machine failed rows: `module.auth.integration.saml_sp_initiated_begin_acs_verifies_determinist_3b5547023b`, `module.database_migrations.integration.cryptographic_state_admission`, `platform.jobs.integration.claim_recovery_and_publication`. Exact nested root is `unit-artifacts/operation-harness.assurance_output_check_machine/real-command-output/runs/real-check-machine`. Relevant unit logs are `go-85702fd10df1d588-81b8c5f8e689`, `go-2f89aed0bdc82556-de8b3c1351b1`, `go-e8eccc6efddf32cc-b3f543dc9303` under `unit-logs/<id>/stderr.log`. All identify `pgtest.go:676` force-drop deadline expiry. Investigate the shared fixture boundary, not three independent product changes.
- W09 remains IN_PROGRESS; W10/W11 have not started. Historical failures and exact private recovery proof remain unchanged.

### W09 streaming reader and catalog-session characterization — IN_PROGRESS (E-059)

- Replaced whole-file event buffering in the real human-output assertion with the existing `readCanonicalUnitEvents` iterator. No second event parser/reducer was introduced; empty, malformed, aliased, changed, noncontiguous and regressing evidence retain the canonical owner's rejection behavior.
- Removed the obsolete finalizer duration-baseline side-effect declaration from the authored task-surface owner and regenerated through Make. `CARTULARY_TEST_RUN_ID=remediation-w09-stream-reader-generate-20261017-01 make generate`: outer 0, 10,230 ms. These changes await focused validation.
- Obtained current task guidance for platform.postgres and harness.browser. The exact catalog cleanup row owns target-exclusion and forced-fallback assertions. Inspection found a concrete coordination defect: target advisory admission is held on a reserved connection while its catalog mutation runs through a second pooled connection. Cancellation can end the mutation's client connection while the independent lock holder releases exclusion before that backend physically settles.
- Added a behavioral assertion to the existing catalog-coordination integration test: the mutation query's PostgreSQL session must itself own advisory admission. The old implementation is being characterized with `umask 077; CARTULARY_HARNESS_CACHE_MODE=off CARTULARY_TEST_RUN_ID=remediation-w09-catalog-session-characterization-20261017-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.postgres_cleanup_target_scoped_coordination` (native 88723; outer Make 2 / normalized 10, 15,013 ms). The expected assertion failed: “catalog operation session does not hold its target admission.” All deadlines remain unchanged.
- Started read-only `make harness-observability-check CHECK=events RESULTS_DIR=.cartulary/test-results/remediation-w09-full-cleanup-diagnostic-20261017-01/unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check` (native 17547): outer 0, `check=events full_validity=0 read_only=1`. It validates the exact 546,824,152-byte stream through the canonical streaming owner; this scoped check cannot establish full invocation validity.
- Implemented the catalog-session repair: `WithMutation` now gives callers its one reserved `sql.Conn`; create/clone/template/drop callers no longer receive a pool capable of selecting a second backend. The same-session contract is explicit in the NLSpec. Cancellation-retired client connections leave server-side session locks in force until that backend settles; subsequent admission still blocks through PostgreSQL. Closed-connection unlock/close is treated as retirement, while other coordination failures remain hard failures. Added cancellation coverage before the existing same-target and different-target concurrency cases. Focused green validation is pending.
- `CARTULARY_TEST_RUN_ID=remediation-w09-catalog-session-format-20261017-01 make format`: outer 0, 8,743 ms. `...=remediation-w09-catalog-session-generate-20261017-01 make generate`: outer 0, 10,050 ms. `git diff --check`: 0. Focused catalog/template/pgtest rows and the default harness contract have started on this repaired candidate; outcomes pending.
- The Go test helper reports eager database cleanup failure through plain testing.Fatal, so these failures currently become product/test_assertion_failure. Preserve that historical classification; assess typed cleanup reporting without replacing a pre-existing product failure. W09 remains open.

### W09 focused catalog evidence and cleanup attribution — IN_PROGRESS (E-060)

- Exact focused run `remediation-w09-catalog-session-focused-20261017-01` published 5/5 passed, 27,449 ms; default contract `remediation-w09-stream-reader-contract-20261017-01` published 10/10 passed, 56,803 ms. Both bind source `sha256:a5fbbf431f85930013163b1f5437336af45f2770e3c25702f6393ff7fdbfdc52`. Native handles were lost at context compaction; outer exits are unavailable, not inferred from these summaries. Exact `make test-run-status RESULTS_DIR=<each exact root> JSON=1` observers returned 0 with terminal summaries and explicit outer/process-not-observed limits. These observations do not replace acceptance.
- The characterization confirms same-session admission and cancellation preservation on the repaired catalog coordinator. It does not establish the cause of every historical cleanup deadline.
- Source review identified a separate existing-requirement repair: eager pgtest cleanup uses plain testing.Fatal, violating cleanup-only `harness/cleanup_error` and potentially obscuring secondary cleanup. Adopt a versioned test-scoped setup/cleanup envelope and an explicit Go diagnostic bundle, preserving earlier product/setup failure and retaining independent normalized cleanup evidence without database identities or transport text. Characterize cleanup-only, combined failure, malformed/conflicting markers, and unrelated tests before broad validation. No deadline or retry change.

### W09 typed cleanup evidence candidate — IN_PROGRESS (E-061)

- Added `cartulary.harness_test_failure.v2` with explicit setup/cleanup phase and pre-cleanup failure state. The pgtest cleanup boundary emits normalized evidence without raw resource/error text and allows other registered cleanups to drain. The Go adapter retains an explicit `cartulary.harness_go_failure_diagnostic.v1` bundle; cleanup alone is harness/cleanup_error (12), while earlier product/setup failures retain their primary outcome. Repeated cleanup cannot manufacture a product assertion. Malformed/conflicting or over-bound evidence fails accounting.
- Row results advance to v4. Browser group results advance to v8 and target results to v7, removing the unused Go-envelope alternative from browser diagnostics. Updated authored producers/readers, cache and telemetry readers, extension contracts, schemas and NLSpec; generated projections through Make. Historical evidence is retained without translation.
- `CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-attribution-format-20261017-01 make format`: outer 0, 7,721 ms. `...=remediation-w09-cleanup-attribution-generate-20261017-01 make generate`: outer 0, 10,540 ms. `git diff --check`: 0. Current evidence owner guidance returned 0.
- Pending native commands: `CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-attribution-contract-20261017-01 make harness-contract` (89624); `CARTULARY_HARNESS_CACHE_MODE=off CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-attribution-helper-20261017-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.unit.testfailure_helper_contract` (96510). Roots use the exact IDs under `.cartulary/test-results`. The helper row passed (native 96510, outer 0, 1,492 ms); the default contract passed (89624, outer 0, 57,981 ms), including cleanup-only/combined/malformed/unrelated marker behavior. Finalizer and broader acceptance have not run on this candidate. W09 remains open.

### W09 finalizer review and primary precedence characterization — IN_PROGRESS (E-062)

- `CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-evidence-format-20261017-01 make format`: outer 0, 7,502 ms. `...=remediation-w09-cleanup-evidence-finalize-20261017-01 make agent-finalize`: native 15756, outer 0, 25,308 ms. Reviewed `unit-artifacts/finalize-summary.json`: no generated mutations, no rollback, schema/catalog/refresh receipts pass; retained-run maintenance skipped because RESULTS_DIR was unset. Finalized source `sha256:7c95e25f1e48f2358ea0975268cf412d0283c8beea71b7883dbd2d49be73fe12`.
- Review found that Go's existing typed-setup shortcut can override a different selected product assertion, contrary to REQ-304. Added separate-selector and sibling-subtest characterization, preserving independent cleanup evidence. `CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-precedence-characterization-20261017-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity` failed at the intended assertion (native 46036, outer Make 2 / normalized 10, 51,357 ms): setup exit 3 incorrectly replaced product exit 10. Repaired attribution considers failed sibling/independent tests before setup selection while retaining both diagnostic kinds. The green command uses ID `remediation-w09-cleanup-precedence-focused-20261017-01` (native 87341). No broad acceptance is claimed.

- Concurrent exact cache-off follow-ups on the repaired catalog candidate: IDs `remediation-w09-catalog-auth-20261017-01`, `remediation-w09-catalog-migrations-20261017-01`, and `remediation-w09-catalog-jobs-20261017-01`, selecting exactly the three E-058 failed rows after current owner guidance. Native handles: auth 40083, migrations 35959, jobs 75319; outcomes pending. These runs do not replace the original concurrent assurance context.

### W09 focused repair closure — IN_PROGRESS (E-063)

- Primary-precedence focused command `remediation-w09-cleanup-precedence-focused-20261017-01` passed (native 87341, outer 0, 55,513 ms), including the exact red-to-green assertion and sibling-subtest regression. Product precedence and independent typed setup/cleanup records now agree.
- Cache-off follow-ups all passed on `sha256:8aea7f05e4dd68d29ab166cc62ef664de12064ec93599d4c931f90acbf9e5bee`: auth (40083, outer 0, 26,207 ms), migrations (35959, outer 0, 30,321 ms), jobs (75319, outer 0, 23,991 ms). Each selected the exact E-058 failed row and completed 3/3 units including fixture preparation and cleanup. These focused contexts do not establish concurrent full-suite closure.
- Active guidance search found no retired Make command uses in current docs, root guidance or repository skills (retained visual-update command excluded as a false prefix match). Historical handoffs remain historical.
- `CARTULARY_TEST_RUN_ID=remediation-w09-catalog-closure-finalize-20261017-01 make agent-finalize` started with RESULTS_DIR unset. Review its generated mutations and source identity before source/Markdown validation and full assurance. W09 remains open; W10/W11 have not started.

### W09 finalizer public failure boundary — IN_PROGRESS (E-064)

- Catalog-closure finalizer 01 (native 48775) failed before mutation: outer Make 2 / normalized graph 10, 4,054 ms. Its own receipt correctly records `harness/child_target_failure` from source validation: the adapter edit made generated topology inputs stale. The graph incorrectly reported product/test_assertion_failure because the CLI returned an untyped child status. No generated changes or rollback occurred. This is a separate W05 boundary repair, not a product assertion failure.
- `CARTULARY_TEST_RUN_ID=remediation-w09-catalog-closure-generate-20261017-01 make generate`: native 18884, outer 0, 10,190 ms. Updated projections through Make.
- Added finalizer publication through the existing private command-failure channel and canonical primary-failure/exit selection. Public failure now uses the same normalized cause as retained action receipts; catch diagnostics are redacted, unclassified errors remain unknown, and cleanup cannot override an earlier failure. Added classification/precedence tests alongside existing transaction and capture tests.
- `remediation-w09-finalizer-attribution-negative-20261017-01 make agent-finalize` (87559) unexpectedly passed for the intended negative scenario: outer 0, 25,457 ms, zero mutations. The finalizer-only edits did not stale topology inputs, so this is a successful finalizer run, not negative-boundary coverage. A precise unavailable-evidence case now uses `CARTULARY_TEST_RUN_ID=remediation-w09-finalizer-invalid-evidence-20261017-01 make agent-finalize RESULTS_DIR=.cartulary/test-results/remediation-missing-finalizer-evidence-20261017-01` (native 68906). Default contract `remediation-w09-finalizer-attribution-contract-20261017-01` runs in native 59012. The unavailable-evidence negative reached the intended preflight and passed its expected rejection: native 68906, outer Make 2 / normalized 2, 805 ms; both retained finalizer receipt and graph report config/configuration_error, with zero mutations and no rollback. This confirms public attribution rather than treating the rejection as finalizer acceptance. Default contract passed (native 59012, outer 0, 56,178 ms). No broader run has started.

- `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-boundaries-finalize-20261017-01 make agent-finalize` is now running with RESULTS_DIR unset; its mutation review precedes source/Markdown validation and full assurance.

### W09 finalized repaired-boundary candidate — IN_PROGRESS (E-065)

- `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-boundaries-finalize-20261017-01 make agent-finalize`: native 36324, outer 0, 25,124 ms. Reviewed finalizer receipt: zero generated mutations, no rollback, successful required actions; retained-run maintenance skipped because RESULTS_DIR was unset.
- Frozen candidate: `sha256:f4d3c723404adb62d228d8e64d79ddb8c415b26ba54d78000dde6629f3aa2375`.
- `...=remediation-w09-repaired-boundaries-source-20261017-01 make source-check CHECK=all`: native 21124, outer 0, 16,821 ms. `...=remediation-w09-repaired-boundaries-markdown-20261017-01 make lint-markdown`: native 62443, outer 0, 10,760 ms. `...=remediation-w09-repaired-boundaries-scripts-20261017-01 make lint CHECK=scripts`: native 36444, outer 0, 585 ms.
- Started `CARTULARY_TEST_RUN_ID=remediation-w09-full-repaired-boundaries-20261017-01 make harness-contract SUITE=full` to exercise the repaired streaming reader, same-session catalog operations, typed cleanup evidence and finalizer public failure boundary in full concurrent assurance. Native session 12464. Preserve executable source until the command drains. Full outcome is pending; no W09 closure or W10/W11 acceptance is claimed.

### W09 frozen-candidate discovery review — IN_PROGRESS (E-066)

- Read-only `make target-plan TARGET=<gate> DETAIL=graph JSON=1` returned outer 0 for test-fast (native 40804), check (70062), ci (1196), and release-check (35691). Outputs: `/tmp/cartulary-remediation-<gate>-plan-20261017.json`. `make help ALL=1 JSON=1` returned 0 and exactly 62 commands; retained output `/tmp/cartulary-remediation-final-help-20261017.json`.
- Current plans: test-fast 1,109 rows / 582 units; check 1,566 / 800; ci 2,293 / 1,202; release-check 2,328 / 1,281. Counts describe selection, not completed assertions or speed improvement. Check has exactly the scheduler and fixture-broker smoke rows plus the public-wrapper smoke operation. Its builds follow consumers; production server build/deployable qualification and advisory Fallow/Gosec audit appear in CI/release. Blocking targeted Gosec and vulnerability checks remain in check/CI/release. Release has one package producer and one inventory producer plus required security/compatibility policy.
- Gate selections are nested. Shared test-fast/check and check/CI unit IDs have identical semantic digests. CI/release differ only for the measurement-summary operation among common IDs, correctly reflecting four versus 33 selected measurement rows and their exact dependencies.
- Human diff review caught earlier schema-name replacements in the historical instrumentation handoff. Restored exactly those owned replacements after confirming no other edits; historical evidence remains at its original identities. Active Core 04 build, OTel generation/conformance and UI-review guidance use current commands. `docs/domain.md`, `.gitignore`, `go.sum` and `pnpm-lock.yaml` have no worktree diff.
- Read-only boundary review confirmed the planner delegates typed inputs to the execution resolver, strict JSON rejects duplicate keys, source selections reject inactive catalog definitions, and retained/input file readers reject documentation paths before metadata access. Existing behavioral contracts cover these boundaries; this review is not a new test run.
- Full assurance remains active in native 12464. Its exact manifest confirms E-065 source and graph `sha256:b2afb0afcc451bf1b567e8c5e9092ce13eb84d2204f3edbda097adb4b9510ab9`. W09 and all required final acceptance remain open.

### W09 full-suite callback migration failure — IN_PROGRESS (E-067)

- E-065 full assurance remains active in native 12464. A live observation reported three failed completed parent cases (backend human output, backend machine output, lint output); both nested checks continue draining. No final outer result is available and executable source remains frozen.
- Exact nested roots under `.cartulary/test-results/remediation-w09-full-repaired-boundaries-20261017-01/unit-artifacts/`: `operation-harness.assurance_output_backend_unit/real-command-output/runs/real-backend-unit` (204/205 passed, infra/preflight_error, 327,199 ms), the corresponding `backend_unit_machine/.../real-backend-unit-machine` (204/205, same classification, 325,251 ms), and `operation-harness.assurance_output_lint/.../real-lint` (11/13, product/test_assertion_failure, 59,394 ms). Parent assurance cases classify their failed expected-success assertions as product/test_assertion_failure.
- Bounded typed-result and log inspection identified one concrete source defect: `internal/testutil/performancefixturelifecycle/lifecycle_integration_test.go:175` and `:184` still pass `func(*sql.DB)` to the now session-bound `withSharedCatalogMutation(func(*sql.Conn))`. The package fails compilation; selected rows `harness.browser.unit.performance_fixture_lifecycle` and `harness.browser.unit.performance_fixture_objects` therefore publish missing_selector_result. Go vet/staticcheck fail on the same compile error. This is an incomplete test-callsite migration from E-059, not a new product assertion or a justification to change deadlines.
- After the owning full command drains, update both test callbacks, search all remaining API callers, and run the exact fixture lifecycle/object rows plus Go lint before finalization and affected assurance. Do not rerun the unchanged failing candidate or mutate it while active.

### W09 nested-check failure reconciliation — IN_PROGRESS (E-068)

- The human nested check from E-065 published 796/800 passed in 994,550 ms. Its four failed units are Go vet, Go staticcheck, go-vulncheck, and the app.server HTTP-harness row. Bounded scanner logs confirm go-vulncheck also fails on the two E-067 fixture-lifecycle callback compile errors; it reports harness/tool_diagnostic_failure rather than a vulnerability finding.
- `app.server.integration.http_harness_restart_storage` failed in `TestHarnessBootsServerAndAssertsEnvelopes` during eager PostgreSQL deletion. The v2 marker and v4 row correctly retain cleanup-only harness/cleanup_error (12), prior_failure=false. The sibling route-mode test passed. The current marker omits the underlying drop/admission cause; the 24.71-second test duration does not prove a timeout or a specific catalog failure. This remains a separate unresolved cleanup failure.
- Exact human root: `.cartulary/test-results/remediation-w09-full-repaired-boundaries-20261017-01/unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check`. Typed unit `go:7262833f646ee1b9:3ae604cc3194` and its referenced stderr are the diagnostic evidence. No raw database identities or service text is added to the tracker.
- The parent full assurance and machine-check case remain active in native 12464; do not infer outer exits from these child summaries. Source remains frozen. Once drained, complete the two callback migrations and retain bounded normalized cleanup cause/stage evidence before any reproduction. Do not increase deadlines, relax ownership, or treat a later isolated pass as a causal explanation.

### W09 full assurance terminal outcome and diagnostic repair — IN_PROGRESS (E-069)

- E-065 full assurance has finished: native 12464, observed outer Make 2 / normalized 10, 1,400,579 ms, 53/58 cases passed. Five expected-success output cases failed: backend human/machine, lint, and check human/machine. All 67 parent cleanup records completed. These are failed acceptance results, not a partial pass.
- Human check: 796/800 passed, 994,550 ms, all 913 graph cleanup records completed. Machine check: 795/800 passed, 934,353 ms, all 686 graph cleanup records completed. Graph cleanup completion does not erase failed eager test cleanup. Both checks contain the three E-067 compiler-driven policy failures. Machine cleanup rows are `module.database_migrations.integration.cryptographic_state_admission` and `module.evidence.integration.evidence_handle_issuance_reports_registered_unav_e4474a6987`; human cleanup is the E-068 HTTP-harness row. All three typed records are cleanup-only with prior_failure=false; their underlying physical causes are unavailable.
- Source freeze ends after the observed outer exit. Next repairs: complete the two fixture-lifecycle callback migrations; add bounded, closed stage/cause cleanup diagnostics through the existing test-failure envelope, preserving errors.Is/As and all primary/secondary classification. Admission connection/lock/release/close and normal/forced drop remain separately identifiable; no raw SQL, database identity, credentials or transport text is retained. Adopt this diagnostic clarification in the NLSpec before projections. No cleanup timing, fallback policy or concurrency limit changes are authorized by these observations alone.
- Validate helper redaction/precedence and both missed unit callers plus snapshot integration, catalog/pgtest contracts, and Go lint before finalization and affected concurrent assurance. W09 remains IN_PROGRESS; W10/W11 have not begun.

### W09 bounded cleanup diagnostics and focused validation — IN_PROGRESS (E-070)

- Completed both `performancefixturelifecycle` integration callbacks' `*sql.Conn` migration; searched all shared catalog API callers. Added typed catalog/drop error stages without changing errors.Is/As, admission, cleanup timing, normal/forced policy or concurrency.
- Adopted optional closed cleanup stage/cause pairs in the existing unshipped v2 test-failure contract. The helper walks bounded error structure, retains joined causes, deduplicates identical diagnostics and marks over-bound/cyclic structure explicitly. SQLSTATE is mapped to closed categories; error text, raw codes, identifiers and transport messages cannot enter retained evidence. The Go adapter preserves the diagnostic and rejects unregistered values. Added behavioral tests for cause retention, primary precedence, redaction and cyclic bounds within existing assertion inventories.
- Current owner guidance for harness.browser, harness.evidence_accounting and app.server returned outer 0; exact JSON copies are under `/tmp/cartulary-remediation-<browser|evidence|server>-guide-20261017-02.json`.
- `CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-cause-format-20261017-01 make format`: native 68839, outer 0, 7,816 ms. `...=remediation-w09-cleanup-cause-generate-20261017-01 make generate`: native 36837, outer 0, 10,100 ms. Generated projections were refreshed through Make; `git diff --check` returned 0.
- `...=remediation-w09-cleanup-cause-helper-20261017-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.unit.testfailure_helper_contract`: native 93376, outer 0, 4,389 ms.
- Active focused commands: cache-off fixture rows with ID `remediation-w09-cleanup-cause-fixtures-20261017-01` (native 38238; exact six rows: performance_fixture_lifecycle, performance_fixture_objects, performance_fixture_snapshot_lifecycle, postgres_cleanup_target_scoped_coordination, postgres_template_preparation, pgtest_helper_contract under their harness.browser unit/integration IDs); default `make harness-contract` ID `remediation-w09-cleanup-cause-contract-20261017-01` (89650); `make lint CHECK=go` ID `remediation-w09-cleanup-cause-lint-20261017-01` (33029). These outcomes remain pending. Full acceptance and physical cleanup causes remain unresolved.

### W09 focused cleanup-boundary outcomes — IN_PROGRESS (E-071)

- All E-070 focused commands passed with observed outer exit 0 on `sha256:2b2af33aed31413acb8755afcc83e56515ad3c9792504d5886636582b76a942b`: six fixture rows (native 38238, 7/7 units, 50,654 ms), Go lint (33029, 6/6, 30,953 ms), default harness contract (89650, 76,142 ms). The helper pass is recorded in E-070. These close the callback compilation defect and the focused diagnostic behavior; they do not close concurrent cleanup acceptance.
- After outer-0 owner guidance, exact cache-off product follow-ups passed: `CARTULARY_HARNESS_CACHE_MODE=off CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-cause-http-20261017-01 make test-slice OWNER=app.server ROWS=app.server.integration.http_harness_restart_storage` (93647, outer 0, 24,136 ms); `...=remediation-w09-cleanup-cause-migrations-20261017-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.cryptographic_state_admission` (88198, outer 0, 28,542 ms); `...=remediation-w09-cleanup-cause-evidence-20261017-01 make test-slice OWNER=module.evidence ROWS=module.evidence.integration.evidence_handle_issuance_reports_registered_unav_e4474a6987` (39060, outer 0, 28,135 ms). The latter two retain the same cache-off prefix. No failed eager cleanup occurred in these contexts, so the historical physical causes remain unknown.
- Started `CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-cause-finalize-20261017-01 make agent-finalize`, with RESULTS_DIR unset. Review its result and generated mutations before source checks and renewed concurrent assurance. W09 remains open.

### W09 diagnostic candidate finalized for concurrent assurance — IN_PROGRESS (E-072)

- `CARTULARY_TEST_RUN_ID=remediation-w09-cleanup-cause-finalize-20261017-01 make agent-finalize`: native 42294, observed outer 0, 25,316 ms. Reviewed v4 receipt: zero generated mutations, rollback not_needed, required actions passed. Retained-run maintenance was skipped because RESULTS_DIR was unset.
- `...=remediation-w09-cleanup-cause-source-20261017-01 make source-check CHECK=all`: native 57147, outer 0, 15,722 ms. `...=remediation-w09-cleanup-cause-markdown-20261017-01 make lint-markdown`: native 52994, outer 0, 27,040 ms. No failed validation is being relabeled by these passes.
- Frozen candidate remains `sha256:2b2af33aed31413acb8755afcc83e56515ad3c9792504d5886636582b76a942b`. Started `CARTULARY_TEST_RUN_ID=remediation-w09-full-cleanup-causes-20261017-01 make harness-contract SUITE=full` to exercise the repaired callbacks and retain exact cleanup causes under concurrent assurance. Native session 24465. Outcome pending; keep executable source frozen and let independent work/finalizers drain. W09 remains open.

### W09 operation-identity review finding — IN_PROGRESS (E-073)

- Read-only review found F-012 in `operationExecutionDigest`: its selected-field hash omits `execution.success_summary` and `execution.test_target`, even though the operation runner uses those fields to decide whether to execute an additional evidence-summary writer. Changing these fields can therefore leave the dispatch guard and compiled operation identity unchanged while changing executable evidence behavior. `step_label` is also passed to the child invocation but omitted. This conflicts with C-002/C-008's executable/evidence identity rule; a global source change is not a substitute for complete operation identity.
- After E-072 drains, add behavioral characterization for evidence-publication changes and public/report identity separation, then hash the complete canonical execution definition rather than a manually maintained subset. The separate public command identity and operation report_key remain outside that execution digest. This removes a future-field omission hazard without creating a new registry or scheduler. Regenerate/finalize and validate the affected graph/dispatch paths; record changed identity/cache applicability.
- No executable source was changed during E-072. Its outcome remains useful for the callback/cleanup candidate, but cannot establish final identity-fix acceptance. W09 remains open; do not begin W10 before this finding and concurrent assurance are reconciled.

### W09 fast-gate resource conflict — IN_PROGRESS (E-074)

- E-072 remains active in native 24465. The `assurance_output_test_fast` expected-success case failed. Its exact nested root is `.cartulary/test-results/remediation-w09-full-cleanup-causes-20261017-01/unit-artifacts/operation-harness.assurance_output_test_fast/real-command-output/runs/real-test-fast`: 579/582 units passed, one failed, two dependency-skipped, 322,201 ms, infra/resource_conflict (nested Make normalized 4). All 137 graph cleanup records completed. No owning full-suite outer exit is available yet.
- Failed unit `operation:harness.toolchain_drift` retains child exit 0 with infra/resource_conflict. Its stderr is empty. The configured unit timeout is 600,000 ms, longer than the entire child run, so a unit-admission timeout is not established by elapsed time. Preserve the typed classification; do not infer scanner or product failure.
- Read-only `make explain-run RESULTS_DIR=<exact nested root> DETAIL=performance JSON=1` returned outer Make 2 / normalized 11 because the dependency-skipped fixture-broker smoke row has no retained row file. The diagnostic did not produce a valid performance view; no event parser or acceptance substitute was introduced.
- Native-source review identifies a possible short-child guard race: Lease.Bind accepts an already-terminal process before entering the admission transaction, but the transaction rejects the same process if it exits between the preliminary liveness check and the mutex-protected bind. The retained result does not prove that this caused E-074; workspace-descriptor or other host-binding failure remains possible without a stage diagnostic. After the current full run drains, characterize this boundary and preserve bounded stage/cause diagnostics before deciding the repair. Do not relax capacity, proof identity, deadlines or process-group cleanup.

### W09 dependency-blocked row accounting review — IN_PROGRESS (E-075)

- E-074's two skipped units declare `rows/harness.browser.unit.fixture_broker_smoke.json` and `rows/harness.command_surface.unit.scheduler_smoke.json` in their evidence outputs, but neither row file exists. Their unit documents have status=skipped, nullable failure and exit fields, and empty missing_outputs. Native publication only writes the unit result for dependency-skipped work. This is F-014: REQ-364/367 and AC-065 require one explicit skipped_dependency terminal record per selected blocked row. Missing evidence is not an authorized skip.
- Repair after E-072 drains: the existing compiler must supply an exact validated row/runner binding to the native publication owner; the scheduler must publish unavailable-row outcomes at its terminal boundary without running product work, inventing passes, overwriting completed rows, or repairing malformed evidence. Characterize dependency failure and cancellation, preserve root failure precedence and all started-unit cleanup, and keep evidence readers fail-closed. Version any changed wire contract and regenerate through Make. No Markdown may enter this projection.
- F-013 separately tracks the host guard-binding race identified in E-074. Characterization must establish the short-child transition without assuming it caused the historical conflict. Retain safe execution-stage diagnostics and exact proof checks; do not broaden admission or loosen process-group ownership.
- Human specification review also found obsolete owner-slice wording in REQ-367's pseudocode. Updated it to the two `test-slice SCOPE` variants; this is documentation fidelity repair only and does not change the frozen executable candidate.

### W09 human-check renderer failure — IN_PROGRESS (E-076)

- E-072 human check published 799/800 passed in 1,100,347 ms, product/test_assertion_failure; all 927 graph cleanup records completed. No PostgreSQL cleanup row failed in this child. Parent full command and machine check remain active; outer full outcome is still unavailable.
- Sole failed row: `module.reporting.integration.renderer_slidev`, unit `vitest_file:d69d729d8f548a5b559862db20ae48445ac9ae76d328bde45d276ba4aeedf245`. Exact root is E-072's `operation-harness.assurance_output_check/real-command-output/runs/real-check`. Typed `vitest-failure-details.json` reports the real assertion at `packages/report-renderer/src/slidev.test.ts:289`: the PDF hook observed zero finished Mermaid SVGs where one was required. The assertion took 15,773 ms against a 60,000 ms test timeout; this is not a reported Vitest timeout. The parser, click-count and literal-text scenarios passed.
- Read-only production review found the existing print readiness predicate waits for every currently mounted Mermaid host but accepts an empty host set; asynchronous component/print readiness therefore needs characterization. This observation alone does not prove the exact rendering race. Preserve the SVG/content and repeatability assertions; do not add sleep, increase timeout, remove the assertion or retry to erase the failure.
- F-015 is required acceptance follow-up under the Reporting owner. No nested AGENTS.md exists under packages. Relevant renderer source/test and Reporting owner specification were inspected; obtain current module.reporting guidance before a focused reproduction. Keep executable source frozen until E-072 drains, then characterize delayed asynchronous diagram readiness and repair the owning boundary if confirmed.

### W09 full diagnostic candidate terminal outcome — IN_PROGRESS (E-077)

- E-072 completed: native 24465, observed outer Make 2 / normalized 10, 1,415,290 ms, 56/58 cases passed. Failed expected-success cases are fast output (E-074) and human check output (E-076). All 67 parent cleanup records completed. The source freeze for this run ends after this observed outer exit.
- Machine check passed 800/800 units in 962,378 ms; all 711 graph cleanup records completed. Its enclosing machine-output assertion passed. Human check remains 799/800 with the one renderer assertion. Both backend output cases and lint now pass, confirming the callback repair in those contexts. No eager PostgreSQL cleanup failure occurred in either check; historical physical causes remain unproven and are not erased by this later pass.
- Exact roots remain under `.cartulary/test-results/remediation-w09-full-cleanup-causes-20261017-01`. This failed full run establishes no overall acceptance. W09-R1 through R4 now sequence the four remaining concrete follow-up slices, each requiring its own tracker checkpoint. W10/W11 remain unstarted.
- Current command-surface guidance returned outer 0 (native 22686), retained at `/tmp/cartulary-remediation-command-guide-20261017-03.json`. Begin W09-R1 with a red characterization of omitted execution-publication fields before changing the digest implementation.

### W09-R1 operation identity characterization — IN_PROGRESS (E-078)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r1-identity-characterization-20261017-01 make harness-contract`: native 94583, observed outer Make 2 / normalized 10, 56,221 ms. The sole failed row was `harness.test_catalog.behavior.routing_contract`. Its exact assertion reports that removing the evidence writer leaves the operation digest unchanged. Other graph assertions passed. This is the intended red boundary, not acceptance.
- Replaced the manually selected digest fields with the complete canonical execution definition. The owner already separates the report key and public command identity; no second identity registry is introduced. Added coverage for evidence-writer toggling, compiled unit identity, run-step argument changes, artifact summary emission and reporting-key separation. Clarified the existing normative identity rule in the NLSpec.
- `...=remediation-w09-r1-identity-format-20261017-01 make format` is running (native 59881). Generation and green graph/dispatch validation follow. Changed operation digests invalidate affected cached plans/evidence; historical identities remain unchanged. W09-R1 remains open and W09-R2 has not started.

### W09-R1 checkpoint — DONE (E-079)

- Format completed (native 59881, outer 0, 7,857 ms). `CARTULARY_TEST_RUN_ID=remediation-w09-r1-identity-generate-20261017-01 make generate`: native 66302, outer 0, 10,380 ms. Projections regenerated through Make; `git diff --check` passed.
- `...=remediation-w09-r1-identity-contract-20261017-01 make harness-contract`: native 55667, outer 0, 57,633 ms, 10/10 units passed. The E-078 characterization is now green; evidence-writer, wrapper-argument, artifact-summary, public spelling and separate report-key assertions pass. `...=remediation-w09-r1-identity-source-20261017-01 make source-check CHECK=all`: native 51199, outer 0, 20,812 ms.
- Validated candidate: `sha256:793c3887894c95ff55a6dbd6e8663cc2f2a47c59ed0de4a9fcf309162cd3973b`. Changed execution digests invalidate affected old plans/cache identities without aliases or historical translation. No product workload, capacity, timeout or retry behavior changed.
- F-012 is closed by red/green behavior and source validation. W09-R1 is DONE and checkpointed before W09-R2 begins. Finalizer and broad final-candidate validation remain due after the remaining repair slices; W09 and the overall effort remain IN_PROGRESS.

### W09-R2 entry — IN_PROGRESS (E-080)

- W09-R1 was checkpointed at E-079 before this slice began. Current command/evidence owner guidance returned outer 0 (native 43020/54087); exact JSON guides are `/tmp/cartulary-remediation-command-guide-20261017-r2.json` and `/tmp/cartulary-remediation-evidence-guide-20261017-r2.json`.
- Characterize a live captured child becoming an unreaped terminal child before the locked bind transaction. The fix must validate the same boot/PID/start proof, preserve the granted reservation and reject absent/mismatched proof, released or observation leases. E-074 historical resource-conflict cause remains unproven.
- Add closed retained execution-stage/cause diagnostics for workspace-descriptor inheritance and host binding without raw error text or private identities. Existing classification, deadlines, capacity and physical group settlement remain unchanged.

### W09-R2 checkpoint — DONE (E-081)

- F-013 repaired: admission now reads process identity and terminal state from one procfs snapshot. The locked transaction accepts a matching pinned terminal child under a valid granted lease without adding a guard or releasing capacity. Mismatched identity, released/unknown leases and observation guards still reject. Removed the earlier liveness shortcut that could bypass lease validation. Physical process-group settlement is unchanged.
- Native execution retains closed `workspace_inherit`/`host_bind` stage and cause diagnostics without raw errors, paths, process IDs or tokens. Publication waits until child stderr has drained; actual child exit and normalized primary failure remain separate. The NLSpec records both boundaries. No wire version or public command migration is required; changed executable inputs invalidate affected caches/evidence.
- Red characterization: `CARTULARY_TEST_RUN_ID=remediation-w09-r2-binding-characterization-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission` (native 92130) outer 2 / normalized 10, 526 ms. The deterministic live-to-unreaped-terminal transition failed at bind as intended. Artifact: `unit-logs/go-881ae025ace037a6-5d60ca7db31e/stderr.log` below that exact run root.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r2-binding-format-20261017-01 make format` (25012) outer 0, 8,025 ms; `...=remediation-w09-r2-binding-generate-20261017-01 make generate` (70541) outer 0, 10,180 ms. Generated outputs were refreshed through Make.
- Green: `CARTULARY_TEST_RUN_ID=remediation-w09-r2-binding-green-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission` (12129), outer 0, 490 ms, 1/1 row. Tests preserve reservation/FIFO capacity and reject mismatched/released bindings.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r2-runtime-20261017-01 CARTULARULARY_UNUSED=none CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime` (78202), outer 0, 45,380 ms, 1/1 row. The accidental unused ambient variable is recorded exactly; it is not a declared harness input. Process lifecycle, group settlement, cancellation, identity and safe diagnostic cases passed.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r2-binding-source-20261017-01 make source-check CHECK=all` (80732), outer 0, 42,202 ms. `git diff --check` outer 0. All run roots are `.cartulary/test-results/<exact-run-id>`; source is `sha256:cbf5a12026645cf402e34b9fcd36def6046758715b92b4e80b9a015cf2b1f824`, same WSL2 environment and repository pins.
- W09-R2 is DONE and checkpointed before R3 begins. E-074's historical physical cause remains unproven; this concrete race repair does not relabel that run. Finalizer and full final-candidate acceptance remain due after remaining repairs. W09 and the overall effort remain IN_PROGRESS.

### W09-R3 entry — IN_PROGRESS (E-082)

- E-081 checkpoint precedes this slice. Catalog guidance `make task-guide OWNER=harness.test_catalog DETAIL=rows JSON=1` (65280) returned outer 0; output `/tmp/cartulary-remediation-catalog-guide-20261017-r3.json`. Command/evidence owner guidance from E-080 remains applicable.
- Project exact catalog row/runner bindings into canonical graph units, including unit identity. Validate unique graph ownership and exact row-output closure before resources. This avoids inferring runners from command strings or introducing a separate registry. Version graph/runtime-plan/retained-context contracts; historical evidence remains unchanged.
- Characterize missing cancelled/dependency row publication. Native publication must retain existing valid completed rows, reject malformed/unsafe/conflicting evidence, synthesize only non-success terminal facts, and preserve root failure precedence and truthful zero-work accounting. Readers must validate retained row bindings without source or Markdown access.

### W09-R3 characterization and first validation (E-083)

- Red: `CARTULARY_TEST_RUN_ID=remediation-w09-r3-row-characterization-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime` (83058), observed outer 2 / normalized 10, 44,689 ms. The cancellation row file was absent at the intended boundary; log `unit-logs/go-15dbae8d0a97629f-503bb18d423f/stderr.log` below that root.
- Authored graph v6 now includes exact row/runner bindings in unit identity, validates output closure and unique row ownership, and retains bindings for native publication and readers. Runtime-plan v4 carries graph v6. Native terminal events publish missing non-success rows before downstream finalizers; existing records are validated and never overwritten. Cache semantic rows must match the bound runner.
- Make generation: IDs `remediation-w09-r3-row-generate-20261017-01` (87916), `...-02` (71964) returned outer 0, 11,060 / 10,630 ms. `...=remediation-w09-r3-row-format-20261017-01 make format` (73977) returned outer 0, 7,785 ms.
- Initial green attempt `CARTULARY_TEST_RUN_ID=remediation-w09-r3-row-runtime-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime` (90795) returned outer 2 / normalized 10, 63,428 ms. Cache test fixtures incorrectly bound their shell-produced results as Go; corrected the fixture binding. Its nested native selected-graph failed the same schema-absence boundary described next. Initial log excerpt reported 8 selected/8 omitted failure groups; no broad causal claim is based on that bounded excerpt.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r3-row-contract-20261017-01 make harness-contract` (95252), outer 2 / normalized 10, 119,263 ms, 7/10 (terminal summary verified at E-085). Owner boundaries rejected reuse of historical retired execution-context v2; current-epoch audit fixtures lacked the new row bindings. The retired v2 identity remains retired; new retained context uses v3 (no prior v3 schema history found), and synthetic audit units now bind their actual catalog runners. Exact logs are `unit-logs/row-harness.command_surface.behavior.owner_boundaries/stderr.log` and `unit-logs/row-harness.evidence_accounting.behavior.current_epoch_evidence/stderr.log` below that run.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r3-row-generate-20261017-03 make generate` (4420) refreshes corrected schema projections. All roots remain `.cartulary/test-results/<exact-id>`. Prior runs are retained; failed subsets establish no acceptance. W09-R3 remains IN_PROGRESS pending corrected focused and source validation.

### W09-R3 second validation investigation (E-084)

- Generation `remediation-w09-r3-row-generate-20261017-03` (4420) returned outer 0, 10,580 ms. Retired-context v2 absence remains asserted; current context is v3.
- Runtime exact-row command with ID `remediation-w09-r3-row-runtime-20261017-02` (74889) returned outer 2 / normalized 10, 20,650 ms. Cache, publisher and terminal-row tests now passed; only its nested selected-graph reader fixture failed. Nested root: `.cartulary/test-results/remediation-w09-r3-row-runtime-20261017-02/native-engine/selected-graph-31465-off`.
- Default contract ID `remediation-w09-r3-row-contract-20261017-02` (51878) returned outer 2 / normalized 10, 77,965 ms, 8/10 (terminal summary verified at E-085). The evidence row failed, and `harness.command_surface.behavior.public_registry_parity` also failed on the invalid synthetic row ID diagnosed at E-085. Its support-projection fixture changes a catalog row from Go to Playwright but had left the newly explicit unit binding at Go. Updated the fixture's binding in that same scenario and added a negative runner-binding mutation. No production guard was relaxed.
- Source check `CARTULARY_TEST_RUN_ID=remediation-w09-r3-row-source-20261017-01 make source-check CHECK=all` (66379) returned outer 0, 24,022 ms. `git diff --check` returned 0. This source evidence precedes the final audit-fixture correction; focused reader verification now runs before native/broader revalidation. R3 remains IN_PROGRESS.

### W09-R3 final fixture diagnosis (E-085)

- Focused reader `CARTULARY_TEST_RUN_ID=remediation-w09-r3-row-reader-20261017-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence` (26011) passed, outer 0, 10,306 ms.
- Native runtime `CARTULARY_TEST_RUN_ID=remediation-w09-r3-row-runtime-20261017-03 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime` (65297) passed, outer 0, 99,292 ms. Cancellation/dependency publication, setup/watchdog outcomes, existing-evidence preservation, unsafe-path rejection, cache runner binding and independent native/Node reader integration all passed.
- Default contract ID `remediation-w09-r3-row-contract-20261017-03` (70891) returned outer 2 / normalized 10, 56,915 ms, 9/10. Sole remaining row `harness.command_surface.behavior.public_registry_parity` failed because the frontend producer fixture appends `_synthetic` to an already near-limit row ID, violating the catalog's 63-character segment bound now enforced in graphs. Replaced only that fixture identity with `package.protocol_ts.boundary_support.synthetic_consumer`; preserved both producer and cycle/error assertions. This same fixture also failed in attempts 01/02. E-083/E-084 counts are corrected from terminal run summaries: attempts 01/02 passed 7/10 and 8/10, respectively. Live snapshots were incomplete and must not establish final counts.
- Source check ID `remediation-w09-r3-row-source-20261017-02` (50937) passed, outer 0, 20,339 ms, before this final test-fixture correction. `git diff --check` passed. Focused registry verification now runs with fresh ID `remediation-w09-r3-row-registry-20261017-01` (47125). No full aggregate acceptance is claimed; R3 remains open until that correction is validated.

### W09-R3 checkpoint — DONE (E-086)

- F-014's concrete missing-row publication is repaired and characterized. Canonical graph v6 carries exact sorted row/runner bindings with unique graph ownership and exact evidence closure; runtime-plan v4 and retained-context v3 project it. Retired context v2 remains unavailable. Node compiler, native plan admission, cache semantic evidence and retained readers validate the same binding. No Markdown input or second command/runner registry was introduced.
- Native publication records missing dependency/cancellation/setup/watchdog outcomes at terminal transitions, before dependent finalizers. It never invents a pass, per-row assertion outcome, or row duration. Existing valid row records remain unchanged; wrong runners, malformed documents and unsafe paths reject. Root product failure stays primary when dependent rows are skipped. Red and green evidence is recorded in E-083–E-085.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r3-row-registry-20261017-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity` (47125) passed, observed outer 0, 54,288 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-r3-row-source-20261017-03 make source-check CHECK=all` (55402) passed, outer 0, 16,058 ms. Final slice source: `sha256:3c3bfae2c9522dbfa7df2130a3612b542f774be517a0d94d7e4461b160fa83f0`. All roots remain `.cartulary/test-results/<exact-id>`, same WSL2/pinned toolchain environment.
- Native runtime, graph identity/closure, reader, cache and corrected registry fixture checks have applicable passing evidence. Last native pass precedes only the independent frontend test-fixture ID correction; source checks and exact registry row cover that correction. The latest whole default-contract invocation is still a failed historical attempt; a complete new default-contract run remains required during W09 final-candidate validation. No whole-gate acceptance is inferred from combined subsets.
- Migration: old graph/runtime-plan/context evidence is incompatible and not translated. Historical runs retain their schema IDs. All units now bind row metadata in identity, invalidating affected caches. Generated schemas and consumers were refreshed through public Make generation. The sole tracker and this durable snapshot retain the failed attempts and final applicability limits.
- R3 is DONE and checkpointed before R4 begins. Finalizer, complete assurance and final check/CI/release acceptance remain due; W09 and the overall effort stay IN_PROGRESS.

### W09-R4 entry — IN_PROGRESS (E-087)

- E-086 checkpoint precedes this slice. Reporting guidance `make task-guide OWNER=module.reporting DETAIL=rows JSON=1` (66844) passed, outer 0; output `/tmp/cartulary-remediation-reporting-guide-20261017-r4.json`. Exact affected row is `module.reporting.integration.renderer_slidev`.
- Read Reporting's normative embedded-Mermaid/print-completion contract and the pinned installed Slidev 53.0.0 component implementation. Characterize asynchronous readiness deterministically before repair. Preserve completed SVG, page count, admitted asset policy and repeatable PDF bytes; do not add retries, sleeps or larger timeouts.

### W09-R4 characterization and implementation (E-088)

- Red: `CARTULARY_TEST_RUN_ID=remediation-w09-r4-renderer-characterization-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reporting ROWS=module.reporting.integration.renderer_slidev` (8932), observed outer 2 / normalized 10, 40,875 ms. The real installed renderer readiness predicate returned `[true,true]` with the expected Mermaid host explicitly removed; the test expected `[false,false]`. This deterministic structural reproduction uses no sleeps/retries and preserves the host for subsequent rendering. Typed artifact: `unit-logs/row-module.reporting.integration.renderer_slidev/vitest-failure-details.json`.
- Renderer now derives diagram counts from parsed source-slide fence tokens, validates each print page's exact source binding/count, waits for finished SVGs, and rechecks after fonts/images/paint. Click copies each close the same source inventory. Empty mounted selections cannot establish readiness. Reporting's owner clause is clarified; this repairs the existing missing-diagram prohibition without changing the admitted render profile or increasing deadlines.
- `...=remediation-w09-r4-renderer-format-20261017-01 make format` (15794) passed outer 0, 7,984 ms; `...=remediation-w09-r4-renderer-generate-20261017-01 make generate` (41186) passed outer 0, 10,050 ms.
- Green exact renderer command with `CARTULARY_TEST_RUN_ID=remediation-w09-r4-renderer-green-20261017-01 CARTULARY_HARNESS_CACHE_MODE=off` (6995) passed outer 0, 41,360 ms. All four selected test cases passed, including installed package SVG inventory/profile, repeatable PDF bytes, click-state diagrams, page-count rejection, bad-resource/import/diagram rejection, and literal text.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r4-renderer-types-20261017-01 make lint CHECK=types` (79611) failed outer 2 / normalized 2, 21,910 ms. TypeScript correctly reported that the direct `markdown-it` import lacked a directly declared type package (TS7016, plus its untyped token callback). Added pinned dev dependency `@types/markdown-it=14.2.0`, the version already resolved transitively in this repository. Do not replace it with an ambient any declaration. The lockfile is refreshed only through `make frontend-lockfile-update`; review its mutations before installation/verification. All roots are `.cartulary/test-results/<exact-id>`; R4 remains IN_PROGRESS.

### W09-R4 dependency maintenance (E-089)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r4-renderer-lockfile-20261017-01 make frontend-lockfile-update` (76100) returned outer 0, 1,570 ms. Reviewed lockfile mutation: exactly three importer lines for `packages/report-renderer`'s pinned `@types/markdown-it=14.2.0`; no resolution/version churn. Package manifest adds one dev-dependency line. The pre-slice lockfile was unchanged from the baseline; this intentional R4 maintenance supersedes that preservation note.
- Type checks and the exact installed-renderer row now run again because package/lock identity changed. No production dependency version changed, no ambient type bypass was introduced, and tool-managed files were not hand-edited. `git diff --check` passed.

### W09-R4 terminal-observation reconciliation (E-090)

- The post-lockfile type, renderer and source invocations from E-089 finished, but their last native tool response was lost during context transition. Their native handles (87676, 93663, 20990) are now closed. Outer exits are unavailable; they are not inferred from pass summaries.
- Exact run IDs `remediation-w09-r4-renderer-types-20261017-02`, `remediation-w09-r4-renderer-green-20261017-02`, and `remediation-w09-r4-renderer-source-20261017-01` report pass (2/2, 4/4 and 9/9 work units; 24,209 / 46,309 / 17,778 ms). Their source is `sha256:2372cb1d08885218056b0b84c6902e3409bd441101edf959a6c94ee74f01f175`. These records establish reported results, not independently observed whole-command success.
- Repeating only the affected type and exact installed-renderer checks with fresh identities is justified by missing outer-exit evidence. `CARTULARY_TEST_RUN_ID=remediation-w09-r4-renderer-types-20261017-03 make lint CHECK=types` (91547) returned observed outer 0, 21,734 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-r4-renderer-green-20261017-03 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reporting ROWS=module.reporting.integration.renderer_slidev` (99188) remains active. Do not mutate executable source before it settles.

### W09-R4 checkpoint — DONE (E-091)

- F-015 is characterized and repaired. Parsed source-slide diagram inventory is required for every print/click copy, including before async hosts mount, and is revalidated after fonts/images/paint. The installed renderer keeps SVG/content, bad-resource and page-count rejection, admitted-profile and repeatable-byte assertions. No retry, sleep or deadline relaxation was introduced.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r4-renderer-green-20261017-03 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reporting ROWS=module.reporting.integration.renderer_slidev` (99188) returned observed outer 0, 47,397 ms; all four selected cases passed. E-090 records the final pinned type check's observed outer 0. Exact roots are `.cartulary/test-results/<run-id>`. Final slice source is `sha256:2372cb1d08885218056b0b84c6902e3409bd441101edf959a6c94ee74f01f175`, WSL2 with repository toolchain pins.
- Migration: one pinned development type dependency is newly direct, with no dependency resolution churn. Generated outputs were refreshed through Make. Renderer semantics repair an existing contract; no render-profile version changed. Changed package/executable identity invalidates affected caches and retained-evidence applicability. Historical E-076 remains failed; this concrete characterization is not proof of every concurrent scheduling detail.
- R4 is DONE and checkpointed before finalizer and broader W09 validation. The last complete default and full harness invocations remain failed historical attempts. Finalizer, source/Markdown checks, complete default/full assurance and W10 final check/CI/release acceptance remain due. W09 and the overall effort remain IN_PROGRESS.

### W09 final-candidate preparation (E-092)

- `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-finalize-20261017-01 make agent-finalize` (8722) returned observed outer 0, 25,535 ms. Reviewed `unit-artifacts/finalize-summary.json`: v4 receipt, zero generated mutations, rollback not_needed, selected actions passed. Retained-run maintenance was skipped because RESULTS_DIR was unset; no eligible successful full warm final-source check was supplied. The retained-secret scan passed. Executable source remains `sha256:2372cb1d08885218056b0b84c6902e3409bd441101edf959a6c94ee74f01f175`.
- Human active-guidance review found four README examples still using retired ROLE input: workbook navigation/commands, grid-adapter and view-contracts. Removed only ROLE=module-author from these examples. There are no nested instruction files under their packages/apps paths. Markdown remains excluded from executable identities; these guidance edits do not invalidate source evidence.
- `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-markdown-20261017-01 make lint-markdown` (66531) returned outer 0, 16,950 ms. `make help ALL=1 JSON=1` returned outer 0; retained discovery copy `/tmp/cartulary-remediation-final-public-help-20261017.json`. `git diff --check` returned 0.
- Active final-candidate commands: `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-contract-20261017-01 make harness-contract` (35374) and `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-source-20261017-01 make source-check CHECK=all` (70921). These results and renewed full assurance remain pending; W09 is open.

### W09 full final-candidate assurance started (E-093)

- Complete default contract `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-contract-20261017-01 make harness-contract` (35374) returned observed outer 0, 59,475 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-source-20261017-01 make source-check CHECK=all` (70921) returned outer 0, 22,399 ms. Both bind the finalized `sha256:2372cb1d08885218056b0b84c6902e3409bd441101edf959a6c94ee74f01f175` candidate. Public help contains exactly 62 commands. These passes do not establish full assurance or final check/CI/release acceptance.
- Started `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-full-20261017-01 make harness-contract SUITE=full` after R1–R4 checkpointing, finalizer review and complete default/source checks. Native session 21740; exact root `.cartulary/test-results/remediation-w09-final-candidate-full-20261017-01`. Keep executable source frozen while its independent work and finalizers drain. W09 remains IN_PROGRESS; W10/W11 have not started.

### W09 distinct-context validation (E-094)

- Review of authored preset selections confirms full assurance covers runtime-boundary and real-output contexts but does not include the separate race-enabled runtime/fixture operations. Their independent execution identities remain intentional. They cannot be replaced by non-race catalog passes.
- Started `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-runtime-20261017-01 make harness-contract SUITE=runtime` (56216) and `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-fixtures-20261017-01 make harness-contract SUITE=fixtures` (29541), both on the same frozen candidate. The R2/R3 native/process/evidence changes and canonical dispatch migration justify these distinct contexts. Results are pending; no retry or source mutation occurred.

### W09 race-runtime failure investigation (E-095)

- Fixture race preset `CARTULARY_TEST_RUN_ID=remediation-w09-final-candidate-fixtures-20261017-01 make harness-contract SUITE=fixtures` (29541) passed, observed outer 0, 59,845 ms. Runtime race preset with ID `remediation-w09-final-candidate-runtime-20261017-01` (56216) failed, outer Make 2 / normalized 10, 94,491 ms, product/test_assertion_failure. Exact roots are `.cartulary/test-results/<run-id>`.
- Inspected run/unit/target summaries and two referenced logs (stdout 7,757 bytes, stderr empty, within the initial three-artifact/80-line/8-KiB bounds). Three failure groups: native selected graph and preparation parity reject at source preparation; private-capture cancellation retains a closed process-registration transport diagnostic (decode/SIGTERM) where the fixture expected empty stderr. Other race packages harnessschema, harnesssession and harnessreview passed. No Go race detector finding is reported.
- Read-only inspection identifies F-016: native runtime removes parent selector variables only for test-slice/browser-e2e, while the consolidated harness-contract SUITE selector reaches step-command children. Nested test preparation inherits the parent selector/source marker. Characterize and repair the canonical private environment boundary using existing typed command metadata; do not weaken public input validation or merely add another test-local name blacklist. No preparation run directory exists for the rejected nested selection, as rejection precedes retained publication.
- F-017: the cancellation fixture publishes readiness from the captured child before the owner completes native registration. Cancellation can kill the synchronous registration transport, yielding a bounded acquisition diagnostic, while parent classification remains interrupted/cancelled_or_interrupted. This is distinct from the fixture's expected post-acquisition capture-drain path. Establish exact readiness/ownership boundaries deterministically, preserve private-output/reaping/cleanup assertions and retain acquisition-failure coverage. Do not remove assertions, add sleeps/retries or extend deadlines.
- E-093 full assurance remains active in native 21740. Executable source remains frozen. R5 then R6 are separate mandatory follow-up slices with checkpoints after the owning full invocation settles. Full/default subset passes cannot close these distinct runtime failures. W09 stays IN_PROGRESS.

### W09 specification fidelity review (E-096)

- Human review of REQ-113 found one obsolete prose inventory still naming ROLE and duration-maintenance inputs, despite their explicit retirement elsewhere. Replaced that duplicate inventory with references to the normative command selector contracts and input matrix. This removes a contradictory projection without expanding the public interface or restoring performance qualification.
- Existing REQ-112/113 already requires undeclared inherited public inputs to be absent from child environments and defines typed child_forwarding. F-016 is therefore an implementation-boundary repair, not permission to weaken public validation. R5 must consume that existing owner metadata, preserve explicit runtime forwarding and keep the private plan independent of Markdown.
- Full assurance remains active in native 21740 on unchanged executable source. The specification-only correction requires later Markdown lint; it does not change source identity or authorize an acceptance claim.

### W09 fast-gate terminal cleanup failure (E-097)

- E-093 full assurance remains active; its fast output assertion failed. Exact nested root: `.cartulary/test-results/remediation-w09-final-candidate-full-20261017-01/unit-artifacts/operation-harness.assurance_output_test_fast/real-command-output/runs/real-test-fast`. The nested summary reports all 582/582 units passed, then harness/cleanup_error in 407,638 ms. Of 209 cleanup records, 208 completed and runtime_close failed. The retained wrapper stderr reports normalized Make Error 12; the enclosing expected-success assertion has exit 1 / product/test_assertion_failure. No owning full-command outer exit is yet available.
- Bounded inspection of the nested stdout/stderr (empty / 406 bytes) retains the new closed diagnostic: `stage=process_group_anchor cause=unresolved errno=0`. This identifies a physical settlement boundary, not a product assertion, unit timeout or evidence-row failure. It differs from E-074's resource-conflict failure.
- Read-only lookup matched exactly one private runtime by repository/run-root identity. Its detail authority remains fenced and detail storage exists. It records 389 producer proofs (388 groups). At the later observation all 389 producer PIDs were absent and no matching recorded group members were observed. These later observations do not establish the membership at the failed close or prove PID reuse as its cause. Private identities and paths are not copied into this tracker; no proof, files or resources were deleted or reset.
- F-018 / W09-R7 must characterize transient versus persistent unanchored group discovery and exact group/producer lifetime. REQ-853 permits signals only under a trustworthy anchor and requires proven absence before release. Consider bounded reobservation within the existing settlement attempt, or owning-lifecycle closure where needed; do not signal ambiguous/reused identities, silently accept an absent leader, start another cleanup attempt or extend its deadline. Preserve independent releases and primary failure precedence.
- Research reference for the supported kernel: [WSL2 6.6.114.1 PID ownership implementation](https://github.com/microsoft/WSL2-Linux-Kernel/blob/linux-msft-wsl-6.6.114.1/kernel/pid.c). This is primary implementation context for reviewing allocation epochs, not evidence of this run's physical cause. Keep the historical E-056 proof separately preserved. R5, R6, then R7 require individual checkpoints after E-093 drains. W09 and overall acceptance remain open.

### W09 human-check pass and graph-fixture failures (E-098)

- The E-093 human-output check case passed. Its nested `real-check` summary reports 800/800 units passed, 1,104,069 ms, with all 887 cleanup records completed. The renderer row is terminal_state=passed, exit 0, 30,782 ms runner duration; its enclosing wall duration is 33,872 ms. These are distinct measures. Exact root is E-093's `unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check`. The parent case's completed success also enforces the child Make's successful exit; the overall full invocation remains active.
- Two additional parent cases failed: `harness.assurance_work_graph_fast` and `harness.assurance_work_graph_matrix`, both exit 1 / product/test_assertion_failure. Their bounded 1,524-byte stderr artifacts report the same assertion at `tools/harness/tests/test-work-graph-cutover.mjs:199`: expected the unbounded resource-claim rejection, but the synthetic unit first failed row-binding/evidence closure. Each stdout is empty. No product row failed in these two synthetic assertions.
- F-019 / R8: the shared generic `unit()` fixture declares a row artifact without a row binding. Give unit-only synthetic work truthful unit-only evidence, preserving the intended resource/service/dependency assertions and the new row-closure guard. Inspect its callers for other preempted negative boundaries; do not broaden accepted schemas or change the expected resource failure to whichever error happens first. Repair and checkpoint after R7, then complete final-candidate assurance.
- Current full observation is 54 passed / three failed / one running of 58; only the machine-output check remains active. This is not a terminal count or whole-invocation acceptance. W09 remains open.

### W09 full terminal outcome and R5 entry (E-099)

- E-093 completed: native 21740, observed outer Make 2 / normalized 10, 1,411,905 ms, 55/58 cases passed. Failed cases are the fast output cleanup assertion (E-097) and two graph-cutover fixtures (E-098). All 67 parent cleanup records completed. Machine check passed 800/800 units in 952,311 ms with all 768 cleanup records completed. Human check evidence remains E-098. No PostgreSQL eager cleanup or renderer assertion failed in either check.
- Executable source freeze ends after this observed owning exit. Historical evidence remains unchanged. R5 begins only after this checkpoint; R6/R7/R8 retain sequential prerequisites. Their focused exit criteria prove each repaired boundary; complete renewed concurrent assurance remains the W09 gate after all repairs, avoiding a requirement that an earlier slice pass a known later-slice fixture defect. No final acceptance requirement is removed.
- R5 implementation intent: project the existing typed input forwarding rules into the private runtime plan, clear inherited public selection controls/source markers at the unit boundary, and preserve only explicitly authorized runtime forwarding plus canonical unit arguments. Avoid target-name branches, another registry and test-local selector blacklists. Version any changed private wire contract and keep its private bindings out of retained context. Characterize future selector metadata and public/owner/suite equivalence.

### W09-R5 typed input boundary implementation and generation (E-100)

- Read current root/tooling/backend instructions and test-operations recipes; no new instruction chain exists in the affected subtrees. Narrow guidance: `make task-guide OWNER=harness.command_surface DETAIL=rows JSON=1` and `make task-guide OWNER=harness.evidence_accounting DETAIL=rows JSON=1`, observed outer 0; outputs `/tmp/cartulary-remediation-command-guide-20261017-r5.json` and `/tmp/cartulary-remediation-evidence-guide-20261017-r5.json`.
- The source-aware compiler now derives a private environment policy from the existing task-surface input contracts: clear all registered target-local inputs and enclosing Make source markers, then admit only selected runtime-forwarded normalized values. Native dispatch applies this immutable projection without another registry or target-name special cases. Canonical unit/fixture arguments remain separately owned. Parent environments remain unchanged; credentials cannot enter plan facts.
- Private runtime plan moves from v4 to previously unused v5; v4 is rejected. Graph v6 and retained context v3 are unchanged. Schema, attachments, owning REQ-854, compiler, native decoder, fixtures and catalog ownership were updated. Behavioral assertions cover future/foreign selectors, admitted forwarding, invalid input, parent preservation and exclusion from retained context. Generated projections are produced only through Make.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-format-20261017-01 make format`: native 94305, outer 2, pre-publication config/configuration_error because new plan version preceded generated validators. No pass claimed.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-generate-20261017-01 make generate`: outer 2, generator rejected a literal regex Unicode escape outside its supported common ECMAScript/RE2 grammar. Authored schema now uses JSON-escaped actual NUL exclusion; generator behavior was not bypassed.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-generate-20261017-02 make generate`: native 33470, outer 2, 10,405 ms tool summary, harness/unknown_failure. Exact root `.cartulary/test-results/remediation-w09-r5-input-generate-20261017-02`; task-surface/topology generator rejected the uncataloged `TestRuntimeUnitInputEnvironment`. Added it to the existing native-runtime row before rerunning. `make explain-run RESULTS_DIR=.cartulary/test-results/remediation-w09-r5-input-generate-20261017-02 TARGET=generate DETAIL=logs` returned outer 0 but no tool logs; bounded referenced stderr supplied the exact catalog failure.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-generate-20261017-03 make generate`: native 49497, observed outer 0, 10,590 ms. Root `.cartulary/test-results/remediation-w09-r5-input-generate-20261017-03`, `generate/tool-run-summary.json`.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-format-20261017-02 make format` is active in native 67979 at this checkpoint. R5 is IN_PROGRESS; focused input/native/schema validation remains required. R6–R8 have not started. No broader acceptance or historical cleanup closure is claimed.

### W09-R5 first focused results (E-101)

- Format `remediation-w09-r5-input-format-20261017-02` completed in native 67979, observed outer 0, 7,701 ms.
- On source `sha256:d65ae72dbd1ba6efa7fcf7466d36f114ae2be62b1dc60b8caac66e2ae3c3ce77`: `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-contract-20261017-01 make harness-contract` (native 97815, outer 0, 61,660 ms), `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-runtime-20261017-01 make harness-contract SUITE=runtime` (native 81065, outer 0, 69,425 ms), and `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-source-20261017-01 make source-check CHECK=all` (native 1521, outer 0, 23,541 ms). Roots are `.cartulary/test-results/<exact-id>`, with owning `run-summary.json` and unit logs. Native selected-graph/preparation and race contracts now pass; this closes the observed selector-leak reproduction. The independently identified R6 readiness race still requires deterministic characterization despite this pass.
- Read-only `make test-run-status RESULTS_DIR=.cartulary/test-results/remediation-w09-r5-input-runtime-20261017-01 JSON=1` returned outer 0; its live state was observation only, not acceptance. Output `/tmp/cartulary-remediation-r5-runtime-status.json`.
- Review found explicit exact-1 boolean forwarding must preserve admitted `1` as well as normalize a default boolean. Corrected the compiler conversion and added enabled/default-disabled assertions; no existing command has credential-valued private forwarding. Default contracts are running again as `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-contract-20261017-02 make harness-contract`, native 96537. R5 remains IN_PROGRESS pending that result; no successor has started.

### W09-R5 boolean characterization and projection refresh (E-102)

- `remediation-w09-r5-input-contract-20261017-02` completed native 96537, observed outer 2 / normalized 10, 56,935 ms. Two failed rows: `harness.command_surface.behavior.owner_boundaries` (generated topology input stale after the compiler edit) and `harness.test_catalog.behavior.routing_contract` (new synthetic RUNTIME_FLAG lacked the source marker required when CARTULARY_MAKE_INPUT_SOURCES is present, so it was correctly ignored). Exact logs: `<root>/unit-logs/row-<row-id>/stderr.log`; bounded excerpts identify both failures. Strict admission and stale-projection guards were preserved.
- Added the synthetic flag's explicit environment-source marker and regenerated: `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-generate-20261017-04 make generate`, native 64102, observed outer 0, 11,070 ms.
- Follow-up `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-contract-20261017-03 make harness-contract` (native 31354) and `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-source-20261017-02 make source-check CHECK=all` (native 9654) are active. No production behavior, timeout, routing assertion or admission rule was weakened. R5 remains IN_PROGRESS until its focused checks close.

### W09-R5 completion checkpoint (E-103)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-contract-20261017-03 make harness-contract`: native 31354, observed outer 0, 57,501 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-r5-input-source-20261017-02 make source-check CHECK=all`: native 9654, observed outer 0, 20,682 ms. Exact roots `.cartulary/test-results/<id>` and `run-summary.json`. Enabled/default-disabled forwarding and future-selector coverage pass; all schema/catalog/generated/surface/identity checks pass.
- E-101's successful race runtime evidence covers native environment application, private plan decoding, selected-graph execution and preparation parity; subsequent changes only corrected the compiler boolean projection and its fixture, covered by E-103. No native implementation changed after that race pass. R5 / F-016 is DONE. Runtime plan v4 is retired; v5 requires the private typed forwarding policy. Retained context v3 excludes that policy, and no Markdown participates in execution.
- R6 may now begin. R6 cancellation readiness, R7 physical settlement, R8 graph fixtures and renewed complete W09 assurance remain mandatory; no W10/final acceptance is claimed.

### W09-R6 entry (E-104)

- Prerequisite E-103 is saved. `make task-guide OWNER=harness.command_surface DETAIL=rows JSON=1` completed native 61149, outer 0; output `/tmp/cartulary-remediation-command-guide-20261017-r6.json`.
- F-017 is a fixture synchronization defect: child payload readiness precedes parent ownership registration and cancellation-handler installation. The post-acquisition cancellation assertion will wait for both payload readiness and the existing onSpawn callback after registration. Preserve the same five-second readiness deadline and all no-output/reaping/private-capture cleanup assertions. Registration-failure cases remain independently asserted in private-child-capture-cases.mjs. No production cancellation behavior changes. R6 is IN_PROGRESS.

### W09-R6 synchronization repair validation (E-105)

- The private-capture cancellation fixture now waits for both CAPTURE_READY (payload/PID) and CAPTURE_REGISTERED from the existing post-registration onSpawn callback. Only then does it cancel. Existing five-second readiness deadline, ten-second workload timeout, kill grace, cancellation classification, zero stdout/stderr, process reaping, retained interrupted capture and runtime-close cleanup assertions are unchanged. No production file changed in R6.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r6-capture-format-20261017-01 make format`: native 27128, observed outer 0, 7,721 ms; `git diff --check` outer 0.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r6-capture-runtime-20261017-01 make harness-contract SUITE=runtime` is active in native 1731. `CARTULARY_TEST_RUN_ID=remediation-w09-r6-capture-boundary-20261017-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries` is active in native 14400; this independently retains registration-rejection/transport failure assertions and output confinement. Roots `.cartulary/test-results/<exact-id>`. R6 remains IN_PROGRESS.

### W09-R6 completion checkpoint (E-106)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r6-capture-runtime-20261017-01 make harness-contract SUITE=runtime`: native 1731, observed outer 0, 61,568 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-r6-capture-boundary-20261017-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries`: native 14400, observed outer 0, 25,281 ms. Exact roots `.cartulary/test-results/<id>`, owning `run-summary.json` and unit logs.
- R6 / F-017 is DONE: the cancellation assertion reaches the declared registered-owner phase and retains all output/reaping/cleanup assertions; acquisition rejection and transport diagnostics still pass independently. Earlier failure remains historical and was not relabeled by a later pass. No wire or public behavior migration in this fixture-only slice.
- R7 may now begin. It must preserve signaling authority, verified absence, existing deadlines and private recovery proof. R8 and complete renewed W09 assurance remain required.

### W09-R7 entry and characterization design (E-107)

- Prerequisite E-106 is saved. `make task-guide OWNER=harness.evidence_accounting DETAIL=rows JSON=1`: native 78958, observed outer 0; output `/tmp/cartulary-remediation-evidence-guide-20261017-r7.json`. Selected owning row is `harness.evidence_accounting.behavior.native_session`; race runtime preset separately covers native/session integration.
- Current settlement rejects the first unanchored membership sweep immediately, even when a bounded owning cleanup attempt can still establish complete group absence. Characterize transient unanchored membership followed by verified empty discovery, persistent ambiguity, and reused identities. Unanchored observations must never authorize signals or enter the known-proof set. A caller without a deadline must still fail closed immediately. Existing trusted-member settlement, pidfd revalidation and process/scope ownership remain intact.
- This is a bounded observation within one physical-settlement attempt, not a retry of failed cleanup, deadline reset or acceptance of missing leader as absence. E-097's later absent processes do not prove its original cause or retroactively close its failed invocation. R7 is IN_PROGRESS; preserved private proof remains untouched.

### W09-R7 characterized bounded settlement repair (E-108)

- Red characterization: `CARTULARY_TEST_RUN_ID=remediation-w09-r7-settlement-characterization-20261017-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_session`, native 41769, observed outer 2 / normalized 10, 1,669 ms. Exact failing subtest `TestSessionRuntimeProof/unanchored_membership_requires_absence_within_existing_attempt` rejected the first sweep instead of reaching the declared empty third sweep. Artifact `<root>/unit-logs/go-48d56ba07b27bfbd-38df029e7abd/stderr.log`; bounded follow-up isolated the intended assertion.
- Native settlement now keeps unanchored discovery pending only when the caller already provides a deadline. It neither signals nor learns those members, preserves its original grace clock/deadline, and succeeds only through the existing complete absence/anchored-settlement conditions. Persistent ambiguity returns process_group_anchor plus the caller deadline/cancellation error; unbounded ambiguity immediately rejects. Existing process identity, pidfd signaling, discovery errors, trusted surviving members and scope cleanup remain unchanged. REQ-853 now explicitly adopts this bounded observation policy. This does not infer E-097's original physical cause.
- Tests cover disappearing unknown membership with zero signals, persistent ambiguity/deadline/stage, unbounded rejection and reused identities; trusted-member and inherited-session cases remain. The transient fixture's one-second safety bound avoids treating test scheduling latency as product behavior; production deadlines are unchanged.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r7-settlement-format-20261017-01 make format`: native 66535, observed outer 0, 8,359 ms. `git diff --check` outer 0.
- Active validation: `CARTULARY_TEST_RUN_ID=remediation-w09-r7-settlement-green-20261017-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_session` (native 10274), `CARTULARY_TEST_RUN_ID=remediation-w09-r7-settlement-runtime-20261017-01 make harness-contract SUITE=runtime` (native 75147), and `CARTULARY_TEST_RUN_ID=remediation-w09-r7-settlement-markdown-20261017-01 make lint-markdown` (native 54863). R7 remains IN_PROGRESS. No preserved proof was removed and no historical failed run was relabeled.

### W09-R7 completion checkpoint (E-109)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r7-settlement-green-20261017-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_session`: native 10274, observed outer 0, 62,868 ms including host-admission waiting. Exact status observation at elapsed 56,005 ms showed the single row waiting for host admission, not failed or cleanup blocked; reader outer 0, `/tmp/cartulary-remediation-r7-green-status.json`.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r7-settlement-runtime-20261017-01 make harness-contract SUITE=runtime`: native 75147, observed outer 0, 62,896 ms. Native runtime/schema/session/review race coverage passes. `CARTULARY_TEST_RUN_ID=remediation-w09-r7-settlement-markdown-20261017-01 make lint-markdown`: native 54863, observed outer 0, 42,580 ms, `adhoc/lint-markdown/tool-run-summary.json`. Roots `.cartulary/test-results/<exact-id>`.
- R7 / F-018 implementation slice is DONE: transient absence, persistent ambiguity, reused identity, unbounded rejection, trusted-member settlement and native race integration pass. Full concurrent assurance remains a W09 exit obligation. Existing failed runtime-close invocations remain failed, their causal limits and private proof preserved. No deadline increase, cleanup retry, unproven signal, or migration of historic evidence occurred.
- R8 may now begin; W09 and overall acceptance remain IN_PROGRESS.

### W09-R8 entry (E-110)

- Prerequisite E-109 is saved. `make task-guide OWNER=harness.command_surface DETAIL=rows JSON=1`: native 29190, observed outer 0; output `/tmp/cartulary-remediation-command-guide-20261017-r8.json`.
- E-098/099 retained both fast and matrix failures at the intended resource-negative fixture, preempted by unbound synthetic row evidence. Inspection confirms the shared helper has only two callers: omitted service dependency and unbounded resources. Replace fabricated row evidence with an explicit empty row binding/output selection; the existing graph builder supplies canonical unit evidence. Add a valid control fixture before preserving both existing negative assertions. No production schema/validator behavior changes. R8 is IN_PROGRESS.

### W09-R8 fixture correction and validation (E-111)

- Generic synthetic units now explicitly select no catalog rows and no row evidence. The graph builder supplies their canonical unit-result artifact. A positive control confirms this fixture is valid before the unchanged missing-service and unbounded-resource negative assertions. Both actual row-binding closure guards and all other compiler assertions remain intact.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r8-fixtures-generate-20261017-01 make generate`: native 76217, observed outer 0, 11,220 ms. Root `.cartulary/test-results/remediation-w09-r8-fixtures-generate-20261017-01`, `generate/tool-run-summary.json`; `git diff --check` outer 0.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r8-fixtures-regression-20261017-01 make harness-contract SUITE=regression` is active in native 58833. This supported preset includes both fast and matrix graph-cutover assertions; a duplicate execution preset is unnecessary if those exact units pass here. R8 remains IN_PROGRESS.

### W09-R8 completion checkpoint (E-112)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r8-fixtures-regression-20261017-01 make harness-contract SUITE=regression`: native 58833, observed outer 0, 264,970 ms. All 42 units passed, including fast/matrix graph cutover and the installed-runtime boundary contract. Exact root `.cartulary/test-results/remediation-w09-r8-fixtures-regression-20261017-01`, `run-summary.json`. `unit-results/operation-harness.assurance_work_graph_fast.json` and `unit-results/operation-harness.assurance_work_graph_matrix.json` each report passed/exit 0.
- Canonical status observations returned outer 0; `/tmp/cartulary-remediation-r8-regression-status.json` recorded progress, not acceptance. Final owning exit above establishes command completion. R8 / F-019 is DONE; intended negative boundaries now execute with internally valid fixtures, without relaxing row evidence or resource/service rules.
- R1–R8 are checkpointed. W09 remains IN_PROGRESS pending renewed finalizer/source/default/full/race assurance on the frozen candidate; W10/W11 remain TODO. Historical failed runs and proof remain preserved.

### W09 renewed candidate finalization and freeze (E-113)

- `env -u RESULTS_DIR CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-finalize-20261017-01 make agent-finalize`: native 40522, observed outer 0, graph duration 25,843 ms. Exact root `.cartulary/test-results/remediation-w09-repaired-candidate-finalize-20261017-01`; reviewed `unit-artifacts/finalize-summary.json` v4. Generated status unchanged, zero updated files, rollback not needed. Schema/catalog/tier/generated validation actions passed. Retained-run maintenance was skipped because RESULTS_DIR was unset; no retained-run validity is claimed from finalization.
- Frozen executable candidate: HEAD `afd6c699980d4256af980a9abb906b2d79004755`, dirty authored changes, source `sha256:6ddf58c01b40468b66f4a9c90fd6727deb304865447cc4a942d4c49598ac662a`. Runtime plan v5; graph v6; retained context v3. `git diff --check` outer 0. No executable edit is planned while this candidate's assurance is active.
- `make help ALL=1 JSON=1` returned observed outer 0; output `/tmp/cartulary-remediation-repaired-public-help-20261017.json`. Authored task surface has 62 public entries including seven UI commands; typed forwarding covers all 49 distinct target-local input names.
- Active commands: `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-source-20261017-01 make source-check CHECK=all` (native 26438), `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-contract-20261017-01 make harness-contract` (native 10858), and `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-markdown-20261017-01 make lint-markdown` (native 93012). Full assurance and final-source runtime/fixture race contexts remain required. W09 and overall effort remain IN_PROGRESS; W10/W11 have not begun.

### W09 frozen candidate focused assurance (E-114)

- On frozen source `sha256:6ddf58c01b40468b66f4a9c90fd6727deb304865447cc4a942d4c49598ac662a`, completed commands all have observed outer 0:
  - `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-source-20261017-01 make source-check CHECK=all`: native 26438, 22,109 ms, 9/9 units.
  - `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-contract-20261017-01 make harness-contract`: native 10858, 60,139 ms, 10/10 units.
  - `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-markdown-20261017-01 make lint-markdown`: native 93012, 12,500 ms; `adhoc/lint-markdown/tool-run-summary.json`.
  - `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-runtime-20261017-01 make harness-contract SUITE=runtime`: native 29846, 60,127 ms.
  - `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-fixtures-20261017-01 make harness-contract SUITE=fixtures`: native 7489, 9,616 ms.
- Roots `.cartulary/test-results/<exact-id>`; graph manifest/summary and unit logs retain selection/source/outcomes. These are work-unit counts, not individual test counts. Help JSON independently confirms exactly 62 names. Finalizer introduced no mutation; retained maintenance skipped because RESULTS_DIR was unset.
- Starting `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-full-20261017-01 make harness-contract SUITE=full`, native 66368. Complete concurrent assurance is required because earlier failures occurred in nested gate output and runtime-close paths; focused successes cannot substitute. Source stays frozen. W09 and overall acceptance remain IN_PROGRESS; W10/W11 have not begun.

### W09 active assurance and migration review (E-115)

- Full invocation native 66368 is active, exact manifest confirms frozen source `sha256:6ddf58c01b40468b66f4a9c90fd6727deb304865447cc4a942d4c49598ac662a`, command `cartulary.harness.command.harness_contract.v3`, SUITE=full, normal cache mode. First canonical observation reported 30/58 units passed, four executing, no failures. Observation output `/tmp/cartulary-remediation-repaired-full-status.json`; reader outer 0 is not acceptance.
- Supplemental source review compared original authored JSON at baseline commit against current task owner: 129 baseline names are removed, `source-check` is the only added name, and the retired-name negative fixture covers all 129 plus six already-retired historical names. Of 69 original internal targets, 50 retain direct private operation report bindings and 19 become explicit unions, catalog support selections, existing readiness/lifecycle functions or removed redundant projections. This review reads authored machine inputs only; canonical source/contract checks remain the validation evidence.
- Editorial review of active README/AGENTS/repository skills/guides and current harness guidance found no matches for the searched retired Make-command forms or task-guide ROLE grammar (rg exit 1 means no matches). Historical handoffs remain historical. Corrected the human disposition table's embedded-web producer wording and expanded grouped replacements into exact lint/browser/backend/native-preset commands. Current NLSpec verification guidance now names `source-check CHECK=catalog` and removes a duplicated generated-check recommendation. These are documentation-only corrections; executable source remains frozen. Markdown lint for the final editorial state remains due.
- No disposition is marked implemented solely from counts. W09 closure still requires the full owning command and cleanup result. W10/W11 have not begun.

### W09 concurrent HTTPS cleanup failure investigation (E-116)

- E-114 full assurance remains active in native 66368; source remains frozen. The installed-runtime boundary unit failed product/test_assertion_failure with unit exit 1; `TestNativeGraphServiceObservations/https` reports nested exit 12 with nil returned error. Parent outer exit is still pending. Independent work drains; no cancellation, retry, executable edit or recovery action was performed.
- Exact nested root `.cartulary/test-results/remediation-w09-repaired-candidate-full-20261017-01/native-boundary/https`: source matches candidate; selected OWNER=harness.browser, ROWS=harness.browser.boundary_support.https_trust, SCOPE=services. Row passed (exit 0, runner 438 ms, wall 4,910 ms). Graph failed harness/cleanup_error, 11/12 units passed, wall 146,338 ms. Browser-group unit exit 12; fixture lease-000001 release failed, provider_close failed and services_close blocked. Startup diagnostics were ready. Unit stderr was empty; browser stderr contained only the NO_COLOR/FORCE_COLOR warning.
- First failure artifact: `<full-root>/unit-logs/operation-harness.harness_runtime_boundary_contract/stdout.log` (434 bytes); owning unit result is `unit-results/operation-harness.harness_runtime_boundary_contract.json`. Nested cleanup evidence is `cleanup-results.json`; exact row and browser-group unit results are retained. Bounded logs were inspected; no raw credentials or process proof were exported.
- Read-only exact-run private authority match found one runtime. Recovery resources retain the acquired browser stack and its managed suite. Browser acquisition has stopping recorded, no terminal settlement receipt, frontend and producer launches closed, backend launch unclosed. All three process proofs remain declared as secure single-link recovery files. A later /proc observation found their leaders and groups absent; that is not evidence of membership at failure and does not establish the original cause. Private path is recorded locally at `/tmp/cartulary-remediation-r9-private-runtime-path.txt`; do not delete, settle or relabel this proof.
- Native physical process cleanup currently returns a bare nonzero status; the Node adapter discards its output and fixture lifecycle reduces failures to class/reason. This loses the stage/cause needed to distinguish authorization, discovery, identity, signaling and physical settlement failures. F-020 / R9 must characterize and repair the owning boundary and propagate closed, bounded, non-sensitive diagnostics without weakening safety or turning a new pass into a causal explanation. Review also identifies broad process discovery as a possible unrelated-process coupling; this is a hypothesis to characterize, not the established cause of this invocation.
- Parent observation: one failed unit, independent human/machine/fast gate assertions still executing. Latest human check observation has 751/800 work units passed, no failures. R9 remains TODO until the full owning exit is collected and this checkpoint is saved. W09 and overall acceptance remain IN_PROGRESS; W10/W11 have not begun.


### W09 full assurance terminal publication, outer exit pending (E-117)

- Frozen source remains `sha256:6ddf58c01b40468b66f4a9c90fd6727deb304865447cc4a942d4c49598ac662a`. Full root `remediation-w09-repaired-candidate-full-20261017-01` published fail, 57/58 work units passed, 1,448,668 ms; all 67 parent cleanup records completed. Native session 66368 remains active, so its outer exit is not yet observed and executable source remains frozen.
- All three real-output assertions published pass/exit 0. Their exact nested roots under `unit-artifacts/operation-harness.assurance_output_<role>/real-command-output/runs/` retain: human `real-check` 800/800, 1,099,260 ms, 926 completed cleanup records; machine `real-check-machine` 800/800, 985,787 ms, 735 completed cleanup records; fast `real-test-fast` 582/582, 352,292 ms, 156 completed cleanup records. Each owning output assertion requires its nested public Make exit to be zero; these are not standalone native exits observed by this session. No acceptance audit has run.
- The sole failed unit remains the E-116 installed-runtime HTTPS boundary. Passing gate children do not explain that failure or the older fenced private-runtime failures. No failed runtime proof has been removed or acknowledged.
- R9 narrow guidance: `make task-guide OWNER=harness.command_surface JSON=1` and `make task-guide OWNER=harness.browser JSON=1`, both outer 0; outputs `/tmp/cartulary-remediation-command-guide-20261017-r9.json` and `/tmp/cartulary-remediation-browser-guide-20261017-r9.json`. Planned diagnostic repair stays within the native physical owner and existing browser adapters; it must carry bounded closed stage/cause/errno, preserve primary classification, and exclude process identities, paths and arbitrary error strings. No physical cause is yet established.


### W09 full owning exit and R9 start (E-118)

- Native session 66368 completed: exact command `CARTULARY_TEST_RUN_ID=remediation-w09-repaired-candidate-full-20261017-01 make harness-contract SUITE=full`, outer Make exit 2 / normalized 10, 57/58 passed. E-117's terminal counts and completed cleanup are confirmed; the source freeze ends here. R9 is IN_PROGRESS; W09 remains open.
- First repair: carry an optional closed cleanup diagnostic through the native process adapter, owned-process adapter, browser lifecycle boundary and native run report. Use one shared typed diagnostic attachment, bounded transport, schema validation and safe projection. Failure status and physical authority remain independent; no retries, timeout increases, process-discovery changes or automatic recovery. Characterize safe propagation and invalid/oversized data before reproducing the exact installed boundary.
- A separate fixture-input mismatch was identified in `TestNativeGraphServiceObservations`: basic diagnostics and cache-off were raw environment values without their input-source markers. Correct the fixture's declared inputs so its retained manifest matches its intended execution; do not treat that mismatch as the HTTPS cleanup cause.


### W09 R9 closed physical diagnostics implemented (E-119)

- Added shared `cartulary.harness_cleanup_diagnostic.v1` with closed stages/causes and errno 0..4095; 4096-byte admission. Native physical process failures emit safe diagnostics; Node consumes bounded stderr after close, validates before forwarding, and keeps failure on missing/malformed/oversized output. Browser process settlement and launch acknowledgement carry distinct stages. The fixture adapter imports only the safe schema; the native run reports its first cleanup diagnostic without changing primary failure classification.
- Characterizations cover live mismatched identity rejection without signals; duplicate/unknown/private fields, oversized frames and arbitrary messages; nested aggregate propagation; adapter classification/cause preservation; private rejection projection without PID/proof leakage. Existing physical settlement/discovery/signaling logic and deadlines are unchanged. Installed-boundary fixture markers now declare basic diagnostics and cache-off correctly.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-diagnostics-generate-20261017-01 make generate`: native 96836 outer 0, 10830 ms; shared schema projections generated through Make. `CARTULARY_TEST_RUN_ID=remediation-w09-r9-diagnostics-format-20261017-01 make format`: native 86721 outer 0, 7629 ms.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-diagnostics-fixtures-20261017-01 make harness-contract SUITE=fixtures`: native 92991 outer 0, 10836 ms. Native runtime and exact owner-boundary assertions remain active in sessions 59460 and 63235. The E-116 physical cause remains unresolved; R9 is not complete.


### W09 R9 native projection omission found (E-120)

- Runtime diagnostic validation `remediation-w09-r9-diagnostics-runtime-20261017-01`, native 59460, outer 2 / normalized 10, 62016 ms: the new physical-failure characterization reached the intended identity rejection and emitted the safe diagnostic, but native decoding rejected it. The new attachment was registered for Node but omitted `foundation_runtime`, so generation excluded it from the native validator closure. The test found a concrete projection omission; physical cleanup behavior was not implicated.
- Corrected the authored attachment's foundation binding. Strengthened the fixture-adapter positive assertion to expect admission independently of the validator being exercised. Regenerate through Make, then rerun the two affected native suites. `remediation-w09-r9-diagnostics-boundaries-20261017-01`, native 63235, outer 0, 25306 ms. `remediation-w09-r9-diagnostics-source-20261017-01`, native 71122, outer 0, 16004 ms; this source pass did not establish native consumer coverage.


### W09 R9 corrected projection validation (E-121)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-diagnostics-generate-20261017-02 make generate`: native 9121 outer 0, 11000 ms. Native schema closure now includes the authored safe diagnostic; no generated file was hand-edited.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-diagnostics-runtime-20261017-02 make harness-contract SUITE=runtime`: native 13395 outer 0, 66785 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-r9-diagnostics-fixtures-20261017-02 make harness-contract SUITE=fixtures`: native 59203 outer 0, 13618 ms. Both independently assert valid native diagnostic admission and preserve unsafe rejection.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-diagnostics-source-20261017-02 make source-check CHECK=all`: native 80881 outer 0, 19234 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-r9-diagnostics-markdown-20261017-01 make lint-markdown`: native 57372 outer 0, 21370 ms.
- Exact installed-boundary reproduction is active: `CARTULARY_TEST_RUN_ID=remediation-w09-r9-installed-boundary-20261017-01 make harness-contract SUITE=runtime-boundary`, native session 22288, source `sha256:28a70f124ed077a4231564cb85b371462383b30a307751b250ecd09af7ea60f8`. Preserve this source until owning exit. No physical cleanup policy has changed; E-116's cause and R9 closure remain open.


### W09 R9 installed-boundary result and preservation gap (E-122)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-installed-boundary-20261017-01 make harness-contract SUITE=runtime-boundary`, native 22288: outer 2 / normalized 10, 258020 ms. HTTPS graph passed 12/12, 73945 ms, all 23 cleanup records completed. Its manifest declares cache-off/basic instrumentation and the exact row correctly. This pass does not explain E-116's concurrent failure.
- The failed assertion is now `TestNativeFixturePhysicalLifecycle/failed_cleanup`, `fixtures_linux_test.go:195`: `exact dependency preservation no such file or directory`. The intentional unsafe browser proof correctly rejected owner cleanup and runtime removal; the later `PreserveRecovery` call returned a missing-file error. The fixture's authorized deferred recovery completed without another reported failure. Exact unit log: `.cartulary/test-results/remediation-w09-r9-installed-boundary-20261017-01/unit-logs/operation-harness.harness_runtime_boundary_contract/stdout.log`.
- PreserveRecovery currently joins protected-proof validation and disposable-file purge errors without a closed stage, so the retained failure cannot distinguish missing required proof from a disappearing disposable entry. R9 remains IN_PROGRESS. Extend the same safe diagnostic projection to those two boundaries and characterize their independent behavior before choosing a repair. Do not ignore missing required proof or infer the cause from the later HTTPS pass. No validation is active; executable source freeze ends after the observed outer exit.


### W09 R9 recovery diagnostic characterization (E-123)

- Recovery preservation now labels required-file validation as `recovery_proof` and disposable traversal/removal as `recovery_purge`; errors.Is/As and fail-closed behavior are preserved. Missing/symlink/directory proof fixtures independently assert that the required-proof boundary remains distinct. The installed negative fixture reports only safe stage/cause/errno and unavailable-proof counts by resource kind, before/after preservation; it exposes no private path or credential.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-preservation-generate-20261017-01 make generate`: native 36061 outer 0, 10910 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-r9-preservation-format-20261017-01 make format`: native 12499 outer 0, 7667 ms.
- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-preservation-runtime-20261017-01 make harness-contract SUITE=runtime`: native 93582 outer 0, 64963 ms. The instrumented installed-boundary run `CARTULARY_TEST_RUN_ID=remediation-w09-r9-installed-boundary-20261017-02 make harness-contract SUITE=runtime-boundary` is active in native session 86076 at `sha256:2693b74c12ced32242aaf61aad9e36d570b6355c39dd08ec5ebc15bf905dacab`. Keep executable source frozen until its outer exit; no preservation or process policy has yet changed.


### W09 R9 installed-boundary pass, causal limits retained (E-124)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-installed-boundary-20261017-02 make harness-contract SUITE=runtime-boundary`, native 86076: outer 0, 252869 ms. Installed physical ownership, failed-cleanup preservation/recovery, exact PostgreSQL/HTTPS graphs, selected-service lifecycle and standalone native session assertions passed at `sha256:2693b74c12ced32242aaf61aad9e36d570b6355c39dd08ec5ebc15bf905dacab`.
- This later pass establishes applicable focused behavior, not the cause of E-116's concurrent HTTPS cleanup failure or E-122's preservation error. Both historical outcomes remain recorded. The safe diagnostics now distinguish physical cleanup stages, launch acknowledgement, required recovery proof and disposable purge. No signal, deadline, discovery, preservation, retry or recovery-authority policy was relaxed.
- Next: finalizer with RESULTS_DIR unset (no applicable full warm check on this changed executable source), inspect every generated mutation, validate source/default contracts, then concurrent full assurance with the new diagnostic boundaries. R9/W09 remain IN_PROGRESS; W10/W11 have not begun. Required concurrent/full/release acceptance is still open.


### W09 R9 candidate finalization (E-125)

- `env -u RESULTS_DIR CARTULARY_TEST_RUN_ID=remediation-w09-r9-candidate-finalize-20261017-01 make agent-finalize`: native 40460 outer 0, graph 26427 ms. Reviewed `unit-artifacts/finalize-summary.json`: generated unchanged, zero updated files, rollback not needed. Retained-run maintenance was skipped because RESULTS_DIR was unset; earlier full warm checks are not applicable to changed source.
- Candidate source `sha256:2693b74c12ced32242aaf61aad9e36d570b6355c39dd08ec5ebc15bf905dacab`. Source/default-contract/Markdown commands are active with IDs `remediation-w09-r9-candidate-{source,contract,markdown}-20261018-01`, native sessions 26559/80352/94318. No broader acceptance claimed.
- Assigned F-021 to E-122's preservation failure so it remains an explicit validation obligation alongside F-020; R9 owns both connected physical-cleanup diagnostic boundaries. New safe diagnostics are validated, but the historical failures are not relabeled as solved by a later focused pass.
- Durable handoff open was queued in Codex. It remains the retained copy of this sole controlling tracker, not a second runtime or planning authority.


### W09 R9 candidate pre-concurrency checks (E-126)

- Current candidate remains `sha256:2693b74c12ced32242aaf61aad9e36d570b6355c39dd08ec5ebc15bf905dacab`. `CARTULARY_TEST_RUN_ID=remediation-w09-r9-candidate-source-20261018-01 make source-check CHECK=all`: native 26559 outer 0, 21991 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-r9-candidate-contract-20261018-01 make harness-contract`: native 80352 outer 0, 59398 ms. `CARTULARY_TEST_RUN_ID=remediation-w09-r9-candidate-markdown-20261018-01 make lint-markdown`: native 94318 outer 0, 12630 ms.
- Launch the full concurrent assurance with fresh identity `remediation-w09-r9-candidate-full-20261018-01`. Keep executable source fixed until its owning exit; preserve independent draining on failure. F-020/F-021 remain open pending applicable evidence and causal assessment; no policy/timeout/retry changes were made to obtain the focused pass.

- Full assurance launch confirmed: native session 28195, exact command `CARTULARY_TEST_RUN_ID=remediation-w09-r9-candidate-full-20261018-01 make harness-contract SUITE=full`; emitted manifest matches the frozen candidate and SUITE=full. Do not edit executable inputs until this session's outer exit is collected.


### W09 R9 concurrent installed-boundary closure observed (E-127)

- Frozen full run `remediation-w09-r9-candidate-full-20261018-01`, native 28195, remains active. Latest canonical snapshot: 37/58 work units passed, no failures; human/machine check-output assertions executing and fast-output assertion waiting. No outer exit or complete acceptance is claimed.
- `unit-results/operation-harness.harness_runtime_boundary_contract.json` published passed/exit 0. Its HTTPS graph passed 12/12, 144873 ms, all 23 cleanup records completed. The owning installed-boundary assertion also covers intentional unsafe cleanup, exact preservation/recovery, selected-service lifecycle and standalone ownership. Backend-unit human/machine and invalid-usage assertions passed.
- This is current-source concurrent evidence for the repaired diagnostic boundaries and unchanged physical safety behavior. E-116 and E-122 remain historical failures with unproven physical causes; they are not retroactively converted into passes. Complete full assurance and final gate/audit requirements remain open.


### W09 R9 real-output assertions drain (E-128)

- Full assurance remains active in native 28195. Human `operation:harness.assurance_output_check` published passed/exit 0: exact nested `unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check`, 800/800 units, 1051924 ms, all 928 cleanup records completed. Fast `operation:harness.assurance_output_test_fast` published passed/exit 0: exact `unit-artifacts/operation-harness.assurance_output_test_fast/real-command-output/runs/real-test-fast`, 582/582, 279523 ms, all 98 cleanup records completed. Both manifests bind the frozen candidate `sha256:2693b74c12ced32242aaf61aad9e36d570b6355c39dd08ec5ebc15bf905dacab`.
- The machine check has a terminal pass summary; its parent assertion and overall native outer exit are still pending. An assertion pass independently requires nested Make success; a summary alone does not establish that exit. Required evidence audits remain unrun. These timings are descriptive and make no performance-qualification claim.


### W09 R9 complete terminal publication, outer exit pending (E-129)

- Machine `operation:harness.assurance_output_check_machine` published passed/exit 0: exact nested `unit-artifacts/operation-harness.assurance_output_check_machine/real-command-output/runs/real-check-machine`, 800/800, 931388 ms, all 733 cleanup records completed. This supersedes E-128's pending machine assertion observation.
- Parent `remediation-w09-r9-candidate-full-20261018-01` published pass, 58/58, 1392557 ms, all 67 cleanup records completed. Native session 28195 is still active, so its outer exit and final wrapper checks remain unobserved. Executable source stays frozen. No workstream is closed from terminal publication alone.


### W09 R9 and coordinated-cutover completion checkpoint (E-130)

- `CARTULARY_TEST_RUN_ID=remediation-w09-r9-candidate-full-20261018-01 make harness-contract SUITE=full`: native 28195 observed outer 0. Parent summary pass, 58/58, 1392557 ms; all 67 cleanup records completed. E-128/E-129 preserve exact successful human/machine/fast roots, counts, durations and cleanup. Full owning completion, including wrapper checks, is now observed.
- Final executable candidate remains `sha256:2693b74c12ced32242aaf61aad9e36d570b6355c39dd08ec5ebc15bf905dacab`, HEAD `afd6c699980d4256af980a9abb906b2d79004755`, uncommitted work. Finalizer changed zero files; retained-run maintenance was skipped because RESULTS_DIR was unset. Current source/default/native runtime/fixture/installed-boundary/concurrent assurance evidence passes; generated projections came only through Make.
- R9 closes the missing-diagnostic boundaries in F-020/F-021: schema-validated bounded physical failure detail survives adapters; acknowledgement, required proof and disposable purge remain distinguishable; unsafe proof still fails and primary classification is preserved. Focused negative characterizations and the concurrent installed-boundary context pass. No physical cleanup policy or deadline was relaxed.
- Causal limits remain explicit: E-116's HTTPS failure and E-122's preservation error are failed historical invocations whose exact physical causes are unproven. The diagnostic repair and new passes do not establish why those older runs failed or claim to have prevented that specific cause. Preserve E-116's private recovery proof. Any recurrence on the final candidate is a required failure to investigate, not an authorized retry. Native harness ownership retains follow-up responsibility for these historical uncertainties.
- All 190 baseline command dispositions are DONE for implemented bindings/selectors/removals, supported by W01–W08 slice evidence, E-115's complete baseline mapping/caller review, current source/default contracts and full assurance. Exactly 62 public names remain; no retired Make runtime aliases. This is interface/migration closure, not release or performance acceptance.
- W09 is DONE. W10 may now begin after this checkpoint. Reuse the exact E-128 current-source successful full check where it proves the required contract, subject to its exact evidence audit; do not repeat it merely to fill a checklist. Launch fresh CI and release identities, retain their native outer exits, then complete required owner/evidence/observability validation. W11 remains TODO; overall remains IN_PROGRESS.


### W10 entry and fresh CI launch (E-131)

- W09 completion checkpoint E-130 was saved before beginning W10. Exact command `CARTULARY_TEST_RUN_ID=remediation-w10-ci-20261018-01 make ci` launched on the frozen final candidate; native session is recorded below. Current-source successful full check from E-128 is retained for applicable check acceptance, with exact audit/observability still pending. Release-check has not started.
- Maintain distinct fresh CI/release run roots. Preserve primary failure and independent draining. On product-measurement publication pause, stop structured observations and avoid sibling status launches until a supported boundary resumes them. No performance qualification, baseline publication, telemetry export or historical evidence translation is authorized or performed.

- CI native session: 71026. Exact root `.cartulary/test-results/remediation-w10-ci-20261018-01`. Executable source remains frozen until its owning exit.


### W10 current-source check evidence validation (E-132)

- CI manifest confirms command `cartulary.harness.command.ci.v4`, 1202 selected units, source `sha256:2693b74c12ced32242aaf61aad9e36d570b6355c39dd08ec5ebc15bf905dacab`; native 71026 remains active. This is a fresh CI graph, not recursive gate execution.
- Exact command `CARTULARY_TEST_RUN_ID=remediation-w10-check-observability-20261018-01 make harness-observability-check RESULTS_DIR=.cartulary/test-results/remediation-w09-r9-candidate-full-20261018-01/unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check CHECK=all`: native 76102 outer 0; terminal output `harness-observability-check PASS check=all full_validity=1 run_id=real-check read_only=1`. This validates complete retained evidence/observability for the exact successful full check without rewriting it. The result is direct reader output; no separate graph-run root is fabricated.
- E-128/E-130 establish check execution/cleanup and the owning output assertion's enforcement of nested Make success; E-132 establishes full retained validity. Reuse this evidence for W10 check acceptance. Required full-owner audits will use a complete applicable gate root; check's tier-limited rows cannot be relabeled as full-owner coverage. CI/release and final owner audits remain pending.


### W10 CI failure investigation checkpoint (E-133)

- CI `remediation-w10-ci-20261018-01`, native 71026, remains active and independently draining on source `2693b74c12ced32242aaf61aad9e36d570b6355c39dd08ec5ebc15bf905dacab`. Two observed failures; no outer exit or final cleanup outcome yet. No executable source has changed.
- F-022: `row:harness.browser.boundary_support.ui_review_lifecycle`, unit normalized 10 / product/test_assertion_failure, row exit 1, wall 33383 ms. The 37-test Node row reports 36 passes and one failure at `tools/harness/browser/tests/test-ui-review-lifecycle.mjs:41`: expected child.exitCode null, received 0. The fixture schedules its leader exit after 200 ms before awaiting forged-start rejection; R7 now preserves unanchored discovery within the existing bounded six-second attempt. W10-R1 must hold the fixture leader until an explicit release, preserve no-signal/PID-reuse/leaderless checks and use exact ownership for teardown. Do not shorten production deadlines or weaken the assertion.
- F-023: `browser_group:webserver-backed:functional-support-default-timeline-grid-entry`, normalized 10 / product/test_assertion_failure. Exactly `module.timeline.browser.clipboard_fidelity` fails; 18 sibling rows pass. `playwright-report.json` in that group's retained artifacts locates `apps/web/e2e/timeline-grid-entry.spec.ts:599`: unchanged-row comparison sees third-row raw text changed from the rich scalar to the prior native-editor paste, row_version 4→5 and edited timestamp advanced. The test presses Escape before the malformed-representation checks. Cause is not established; investigate actual cancellation/autosave and fixture synchronization before remediation. No product source changes existed in this spec at failure.
- Exact logs and row/unit artifacts are retained beneath the CI root, including each unit's stderr and the timeline group's `playwright-report.json`. Trace is retained but has not been externally exported. W10-R1 and W10-R2 will be separate sequential repair slices after the CI owning exit, each checkpointed before its successor. Release-check has not started. W10 and overall remain IN_PROGRESS; W11 TODO.


### W10 clipboard contract diagnosis (E-134)

- Core 03 REQ-03-087 requires grid autosave on paste completion. `TimelineScalarEditor` marks native `insertFromPaste`; `useTimelineScalarRenderers` immediately delegates that value to the save owner. Escape only discards uncommitted work; it cannot revoke the already admitted paste write. The existing clipboard test snapshots rows immediately after Escape without awaiting that paste settlement. CI's row_version 4→5 and exact pasted value match this pending write. This is a fixture synchronization gap, not evidence of malformed data being submitted.
- W10-R2 will await the exact persisted pasted value and accepted local version before the immutable malformed-input baseline, and add a separate native typed draft followed by Escape to preserve the cancellation assertion. Product autosave behavior and invalid-input rejection remain unchanged; no arbitrary sleep, retry or timeout expansion.
- `make task-guide OWNER=harness.browser` and `make task-guide OWNER=module.timeline` completed outer 0 (combined native 58685); selected exact rows are identified in E-133. `git diff --check` passed; `.gitignore`, `docs/domain.md` and `go.sum` remain unchanged against current HEAD. This is an interim review, not W11 closure. CI continues draining on unchanged executable source; no release run has started.


### W10 missed programmatic caller (E-135)

- CI now reports a third failed unit, `browser_group:webserver-backed:functional-support-default-visual-harness`, normalized 10 / product/test_assertion_failure. Exactly `harness.browser.boundary_support.frontend_artifact_lifetime` fails (1571 ms); its five sibling assertions pass. The group's `playwright-report.json` reports `apps/web/e2e/support/runtime/frontendBuild.ts:35`: independent Make exits 2 because `build-web-measurement` has no rule. The conditional argument array also retains retired `build-web`.
- F-024 is a concrete missed W09 caller migration, not an asset-publication implementation failure. E-130's statement that active callers were reconciled was incomplete; its recorded command-disposition inventory and actual passing outcomes are unchanged. Caller closure is reopened under mandatory W10-R3, sequenced after R1/R2 checkpoints. Replace this helper's invocation with `build COMPONENT=web` / `build COMPONENT=web-measurement`; retain independent identity/lease/environment isolation and both real publication assertions. Review programmatic Make callers beyond literal shell commands.
- CI remains active in native 71026 with executable source frozen; its complete outcome/cleanup is pending. The exact failure artifacts remain in `remediation-w10-ci-20261018-01`. No executable edits or blind reruns have occurred. All three follow-up slices must close before renewed final-candidate acceptance; W11 and release remain unstarted.


### W10 terminal browser release failure observed (E-136)

- CI `remediation-w10-ci-20261018-01` is still active in native 71026. Five unit failures now comprise the three prior independent assertions, one new cleanup failure and the webserver-backed target-summary projection. The aggregate summary failure is not an additional root cause. Latest observation: 1195 passed, five failed, one running and one pending; outer exit remains unobserved.
- F-025: `browser_group:webserver-backed:functional-support-default-workbook-generic` is normalized 12 / harness/cleanup_error. Its retained browser-group result passed in 29814 ms, and `module.workbook.browser.the_browser_workbook_exercises_parties_notes_tas_8ca985fff2` passed. The later fixture release changes unit completion to failure. The selected session is `webserver-backed-webserver-backed-default-2-default-browser_functional--allocation-002`; public startup/stack attestations remain retained under `_shared/test-services/bbfd07860ca176043d467410/browser-sessions/`.
- Group stderr contains only the standard NO_COLOR/FORCE_COLOR warning; unit stderr is empty. `cleanup-results.json` and `run-summary.json` were not yet published at initial inspection. Await terminal native cleanup diagnostic and exact retained cleanup records; do not infer a process cause from row success or missing logs. Do not delete private recovery proof or alter source while the owning command drains. W10-R4 is mandatory after R1–R3 checkpoints.
- Programmatic caller review used baseline/current authored task-surface owners to identify retired names, then inspected 23 source files containing a literal Make invocation (31 relevant-name matches), outside-tooling references and shell callers. The confirmed live stale command is the F-024 conditional frontend helper. Other inspected matches are report partitions/private protocol symbols, file-production stamps, synthetic schema fixtures or explicit retired-command rejection tests. Review notes: `/tmp/cartulary-remediation-programmatic-make-review.txt`, `/tmp/cartulary-remediation-retired-caller-review.txt`, `/tmp/cartulary-remediation-shell-caller-review.txt`. These manual review notes are not executable requirements or conformance inputs.


### W10 CI owning outcome and R1 entry (E-137)

- `CARTULARY_TEST_RUN_ID=remediation-w10-ci-20261018-01 make ci`, native 71026, completed with observed outer Make 2 / normalized 10. Exact source is `sha256:2693b74c12ced32242aaf61aad9e36d570b6355c39dd08ec5ebc15bf905dacab`; summary fail, product/test_assertion_failure, 2646162 ms, 1197/1202 units passed and five failed. The five comprise four independent failures and the failed webserver-backed summary described in E-136.
- Cleanup publication contains 1199 completed, three failed and one blocked record. Lease `lease-000445` fails fixture_release in records 1179/1180 for the workbook-generic unit; provider_close fails at 1201 and services_close is blocked at 1202. Preserve cleanup separately from the primary product failure. Native terminal output adds `[RUNTIME-CLEANUP] stage=process_signal cause=io errno=22`. This narrows F-025 to the native signaling path; it does not distinguish pidfd acquisition from signal delivery or prove the physical cause. Do not infer that the historical E-116/E-122 failures share this cause.
- The source freeze ends after this observed owning exit. Begin W10-R1 only: synchronize leader readiness/release and exact owned teardown in the existing UI lifecycle fixture. Retain the forged-start and leaderless rejection assertions and existing production settlement deadlines. Task-guide guidance is E-134. Validate the complete exact row before checkpointing R1 and starting R2. R2/R3/R4 remain TODO; no successful CI/release acceptance is claimed.


### W10-R1 first focused result and readiness repair (E-138)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r1-ui-lifecycle-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_lifecycle`, native 42854, outer 2 / normalized 10; 36/37 Node tests passed. Source `sha256:7624636f7f0b7350204b26a13271b2ce034dc80d25c7ffd79d7dceb70a87a5b3`; graph wall 29499 ms. The changed fixture failed its five-second readiness bound before reaching either ownership assertion.
- Cause is an escaping error introduced by R1 in nested Node eval source: the descendant received a literal newline inside a quoted program string. Replace that construction with `String.fromCharCode(10)` and propagate descendant startup exit/error through the leader; the parent now rejects early leader exit while awaiting readiness. The existing readiness bound is unchanged. Exact owned teardown completed; do not treat this as the original CI failure or as evidence against native ownership.
- Original CI private recovery authority is retained. Its exact locator is saved privately in `/tmp/cartulary-remediation-w10-ci-private-runtime-path.txt` (mode 0600); matching runtime owner binds the original CI run root/run ID. No recovery or deletion was performed. R1 remains IN_PROGRESS; R2–R4 have not started.


### W10-R1 diagnostic assertion correction (E-139)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r1-ui-lifecycle-20261018-02 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_lifecycle`, native 86459, outer 2 / normalized 10, wall 31451 ms; 36/37 tests passed. Source `sha256:83183a70115ddcaafc68b855a916ff2d874418b7bbb79e5839bd55aa58855bdf`. Readiness succeeded and forged-start rejection occurred; the new predicate incorrectly expected the diagnostic directly on the UI wrapper error.
- The UI adapter correctly wraps native cleanup failure as `ReviewFailure(cleanup_failed)` with its native cause. Use the existing bounded `findCleanupDiagnostic` reader to validate the retained native stage through that cause chain. No production adapter, native policy or deadline changes. This corrects the new test's observation boundary, without relaxing its required process-group-anchor rejection. R1 remains open until the complete row passes.


### W10-R1 completion checkpoint (E-140)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r1-ui-lifecycle-20261018-03 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_lifecycle`, native 2834, observed outer 0; complete row passes, graph wall 36519 ms, 1/1 units and all ten cleanup records completed. Source `sha256:96a9c6e8a681b1d69ae95ec202398553e0f9b6daf01f0047a7dd2a4b31f89a71`. The lifecycle inventory retains all 37 tests, including UI/public sessions, native ownership, failed acquisition, orphan recovery and repeated settlement.
- F-022 closes through explicit descendant readiness and leader release, both exact process-group-anchor rejection assertions, proof-based liveness and native cleanup using separately journaled descendant authority. Production behavior and existing grace/kill budgets are unchanged. The earlier R1 fixture mistakes and failed attempts remain in E-138/E-139.
- R1 is DONE and checkpointed before R2 begins. Only the lifecycle test file changed. E-132's successful check remains evidence for its original source; changed executable test input means it cannot be relabeled as a current-source evidence audit. Renew required final-candidate acceptance after R2–R4. Finalizer remains due; RESULTS_DIR must remain unset without an eligible successful current-source full check.


### W10-R2 entry (E-141)

- E-140 was saved before R2 entry. Apply the existing Core 03 REQ-03-087 paste-completion contract and Design §8.6 uncommitted Escape semantics to `apps/web/e2e/timeline-grid-entry.spec.ts`. Await native pasted content in the server query and the matching displayed row version; then type a separate unsaved draft and cancel it. Only after those boundaries take the full immutable baseline for malformed clipboard rejection.
- Keep the existing malformed-representation list, exact no-batch-attempt assertion and full-row equality. Do not change autosave/product implementation, goldens, timing budgets or retry policy. Owner guidance is E-134; validation is exact `module.timeline.browser.clipboard_fidelity`.


### W10-R2 completion checkpoint (E-142)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r2-clipboard-20261018-01 make test-slice OWNER=module.timeline ROWS=module.timeline.browser.clipboard_fidelity`, native 56651, observed outer 0. Source `sha256:ca3aff932469552b284a56da1d5b47732fbf1cada83514c40c53cc1f297e9c84`; graph pass 12/12, 78212 ms, all 23 cleanup records completed. Exact Playwright row passed/exit 0 in 4169 ms (group wall 6854 ms).
- F-023 closes with persisted native-paste content and matching displayed version, a separate typed-draft Escape assertion, and the unchanged complete malformed-input no-write comparison. No product code, timing budget, retry, golden or clipboard expectation was weakened. R2 is DONE and checkpointed before R3 begins.


### W10-R3 entry (E-143)

- R2 checkpoint E-142 is saved. Migrate `apps/web/e2e/support/runtime/frontendBuild.ts` from the conditional retired Make names to the adopted `build COMPONENT=web` and `build COMPONENT=web-measurement` contracts. Preserve the helper's independent environment/identity and both concurrent real build invocations. E-136 records the extended programmatic caller review; E-134 supplies harness.browser owner guidance.
- Validate exact `harness.browser.boundary_support.frontend_artifact_lifetime`: loaded assets must retain their hashes across repeated reloads while both independent production/measurement frontend builds finish successfully. No source/protocol alias or snapshot update is permitted as a repair.


### W10-R3 completion checkpoint (E-144)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r3-frontend-lifetime-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.frontend_artifact_lifetime`, native 28893, observed outer 0; graph pass 12/12, 88468 ms, all 23 cleanup records completed. Exact row passed/exit 0 in 13313 ms (group wall 16057 ms). Source `sha256:7114ad90b3c2f655196ff039dbe2967a71652d56017208b5b48e405a7eebe108`. The exact artifact-lifetime assertion succeeds using both independent typed frontend build variants and unchanged hash/reload checks.
- F-024 closes: `publishIndependentFrontendBuild` now invokes `build COMPONENT=web` or `build COMPONENT=web-measurement`; no compatibility alias was introduced. Its existing environment/identity isolation is unchanged. E-136 records the extended programmatic/shell caller review. `git diff --check` passes for all three W10 repair files. R3 is DONE and checkpointed before R4 entry. Final source/format/finalizer and complete gate acceptance remain due after R4.


### W10-R4 entry (E-145)

- E-144 is saved before R4 entry. Investigate F-025 from the original CI native diagnostic (`process_signal`, `io`, errno 22), exact failed browser-session proof and native process owner. The current stage combines pidfd acquisition, identity revalidation and signal delivery; do not guess which syscall failed or infer original process state from later absence.
- Preserve all existing ownership, PID-reuse, complete-discovery, grace/deadline and primary-failure contracts under REQ-853/854. Establish a meaningful red/green physical-boundary characterization before a production correction. Do not erase the original failed CI evidence, reset services, weaken proof checks or use a blind rerun as diagnosis. Original private locator is retained as described in E-138.
- Current candidate source is E-144. R1–R3 are checkpointed; W10 remains IN_PROGRESS and W11 TODO. No successful final-candidate CI/release result is available.


### W10-R4 physical characterization and correction (E-146)

- `make task-guide OWNER=harness.evidence_accounting` outer 0 (native 21756); existing native-session row selected. `CARTULARY_TEST_RUN_ID=remediation-w10-r4-native-red-20261018-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_session`, native 52553, observed outer 2 / normalized 10, failed 1/1, 1299 ms, all 10 cleanup records completed. Source `sha256:afcd64160f2abf4310730439c732dc6f09ed05413ecb3d96f7a5b794b069956d`. `make explain-run RESULTS_DIR=.cartulary/test-results/remediation-w10-r4-native-red-20261018-01` outer 0; exact status saved at `/tmp/cartulary-remediation-w10-r4-red-status.json`.
- The new real-process characterization explicitly holds a group member alive, releases/reaps its leader, verifies the leader's `/proc` absence, then signals the stale member observation. It failed at `TestSessionRuntimeProof/signal_observes_reaped_leader_while_group_survives`: `reaped member did not become an absence observation invalid argument`. Both owned fixture processes are reaped by test cleanup. This is a deterministic physical failure, not a later CI pass used as diagnosis.
- The matching WSL 6.6.114.1 [kernel process-handle implementation](https://raw.githubusercontent.com/microsoft/WSL2-Linux-Kernel/linux-msft-wsl-6.6.114.1/kernel/fork.c) rejects a PID lacking a thread-group task with EINVAL. The [PID owner](https://raw.githubusercontent.com/microsoft/WSL2-Linux-Kernel/linux-msft-wsl-6.6.114.1/kernel/pid.c) retains a numeric identity while other PID roles remain. This explains the reproduced reaped-leader case. Original CI retained only the combined process_signal stage, so its exact syscall/race remains an inference; neither the original CI nor E-116/E-122 is reclassified as successful or causally proven.
- Corrected pidfd acquisition: only EINVAL followed by a fresh missing-process observation becomes a member-absence observation. Live/reused/unreadable processes retain the acquisition failure. Complete group discovery, original grace/deadline, pidfd-only delivery, exact post-acquisition identity validation and primary failure remain unchanged. Handle acquisition and identity revalidation now retain distinct safe diagnostic stages; delivery remains process_signal. REQ-853 clarifies this boundary. Added an acquired-handle/forged-start rejection and unreconciled EINVAL failure coverage alongside permission denial and complete-discovery checks.
- Generated schema projections are being refreshed through `CARTULARY_TEST_RUN_ID=remediation-w10-r4-generate-20261018-01 make generate` (native 52404); outcome pending. R4 remains IN_PROGRESS pending green physical, runtime/installed, and browser cleanup validation. No resets, deadline extensions, retries, aliases or historical evidence rewrites were introduced.


### W10-R4 focused green results and reporting follow-up (E-147)

- Generation native 52404 completed outer 0, 10630 ms. The new safe stage appears in generated Go schema data and the generated JavaScript cleanup-diagnostic validator; generated through Make only. `CARTULARY_TEST_RUN_ID=remediation-w10-r4-format-20261018-01 make format`, native 73364, observed outer 0, 2/2 units, 7751 ms. `git diff --check` passes.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r4-native-green-20261018-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_session`, native 99920, observed outer 0, 1/1, 5071 ms, all 10 cleanup records completed. The physical red/green, forged-start guard and unreconciled EINVAL/permission failure cases pass.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r4-workbook-cleanup-20261018-01 make test-slice OWNER=module.workbook ROWS=module.workbook.browser.the_browser_workbook_exercises_parties_notes_tas_8ca985fff2`, native 53660, observed outer 0, 12/12, 105074 ms, all 23 cleanup records completed. Both focused runs bind source `sha256:9e6eb1cf096462044c434314b629a9ab2562b2f29603704a0e4f66fcc9adf467`.
- Focused physical-context validation is running: `CARTULARY_TEST_RUN_ID=remediation-w10-r4-runtime-20261018-01 make harness-contract SUITE=runtime` (native 19674) and `CARTULARY_TEST_RUN_ID=remediation-w10-r4-installed-20261018-01 make harness-contract SUITE=runtime-boundary` (native 28057). Source stays frozen. R4 remains IN_PROGRESS until their owning exits are collected; broader final acceptance remains open.
- F-026 / W10-R5 is TODO after the R4 checkpoint. Read-only review found task-surface `harness_checks[].command` duplicates operation argv and still reports `$(MAKE) --no-print-directory harness-runtime-boundary-contract`. These logical-check argv are not dispatched by the current graph, but the task-surface report exposes them as commands. Remove redundant executable guidance from logical inventory/reporting and reject its reintroduction, keeping canonical operations as the execution owner. This is a reporting-migration omission, separate from the R3 live caller. Do not begin its edits before R4 closes.


### W10-R4 completion checkpoint (E-148)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r4-runtime-20261018-01 make harness-contract SUITE=runtime`, native 19674, observed outer 0, 1/1 graph unit, 68071 ms, all 10 cleanup records completed. The operation retains the race-enabled runtime/schema/session/review package context.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r4-installed-20261018-01 make harness-contract SUITE=runtime-boundary`, native 28057, observed outer 0, 1/1 graph unit, 253495 ms, all 10 cleanup records completed. Installed service/browser lifecycle, unsafe-proof rejection and preservation/recovery boundaries pass. Nested HTTPS evidence is retained under `native-boundary/https`; its parent assertion requires a successful native result and independent retained observability validation.
- R4 is DONE after the E-146 red physical characterization and E-147/E-148 green native, race, installed and exact browser validations. F-025's reproducible member-absence defect is corrected without loss of live group settlement, authority, primary classification or deadlines. Original CI syscall attribution remains explicitly unproven; no historical failure is relabeled. Final-candidate concurrent CI/release acceptance still must pass.
- E-148 is saved before R5 begins. R5 will remove stale logical reporting argv; no physical operation, preset inventory or scheduling change is intended. W10/overall remain IN_PROGRESS; W11 TODO.


### W10-R5 entry (E-149)

- E-148 is saved before edits. `make task-guide OWNER=harness.command_surface DETAIL=rows` and `make task-guide OWNER=harness.generated_artifacts DETAIL=rows` completed outer 0 (native 56791/54058). Select public-registry parity for admission/discovery and use all-source validation after regeneration; preserve the existing comprehensive assurance preset for final acceptance.
- Logical checks retain names, backing-script inventory and the three canonical smoke selections. Remove their unused duplicated command arrays and the corresponding report field; private executable argv remains owned by the work graph. Add a negative semantic-admission case so even a plausible Make command cannot be reintroduced in reporting metadata. No command aliases, execution identities, assertion inventories or scheduler changes.


### W10-R5 completion checkpoint (E-150)

- Removed 20 duplicated command arrays from the 45 logical-check records and removed command argv from their machine/human report projection. Names, backing-script inventories and canonical smoke selections remain. Semantic admission rejects command fields, including empty/null values; executable definitions remain only in the operation owner. The report's new cutover schema remains v4; no legacy output-field alias is added.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r5-generate-20261018-01 make generate`, native 81127, observed outer 0, 10820 ms. Generated task-surface projection refreshed through Make.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r5-parity-20261018-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity`, native 18326, observed outer 0, 1/1, 58094 ms. The new negative admission case and existing wrapper/discovery/registry tests pass.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r5-source-20261018-01 make source-check CHECK=all`, native 4635, observed outer 0, 9/9, 16079 ms, all 18 cleanup records completed. Source `sha256:a2835c668d57649585bb36bd5b211fa272ec642d30dbdf763c7dceef95424191`. `git diff --check` passes; `.gitignore`, `docs/domain.md` and `go.sum` remain unchanged against current HEAD.
- F-026/R5 closes. R1–R5 are DONE and checkpointed. Next: finalizer with RESULTS_DIR unset, review any generated mutations, then freeze the source for complete harness/check, CI and release validation. Earlier successful full/check evidence predates the W10 repairs and is not relabeled current. W10/overall IN_PROGRESS; W11 TODO.


### Final candidate finalization and acceptance freeze (E-151)

- `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-finalize-20261018-01 env -u RESULTS_DIR make agent-finalize`, native 52959, observed outer 0, graph 25463 ms. Reviewed `unit-artifacts/finalize-summary.json`: schema v4, pass; generated status unchanged, zero updated files, rollback not needed. Schema/catalog/tier receipts and refreshed validation bind the same source `sha256:a2835c668d57649585bb36bd5b211fa272ec642d30dbdf763c7dceef95424191`.
- Retained-run maintenance was skipped because RESULTS_DIR was unset; no applicable successful full warm check existed for this candidate. No finalizer-caused source changes exist. E-150 all-source/parity results remain applicable; do not rerun them solely to fill a checklist.
- Executable source is now frozen for final acceptance. Renew full harness assurance, including its real human/machine check invocations, and fresh CI under distinct roots. Their physical graphs remain independent and use existing host admission. A successful nested full check may satisfy the check obligation only with its parent assertion/outer exit and exact retained audit; no success is assumed. Complete release follows resolved candidate acceptance, with its own fresh identity and one package identity. No performance qualification/baseline publication is required.


### Frozen candidate gate launches (E-152)

- `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-full-20261018-01 make harness-contract SUITE=full`, native 25581, active. Command `cartulary.harness.command.harness_contract.v3`, graph `sha256:a1ce5b6f71d3b6b3f7c993d15c276f7d3b2f30df36766191c2a421e0552da33b`.
- `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-ci-20261018-01 make ci`, native 23399, active. Command `cartulary.harness.command.ci.v4`, graph `sha256:3c7f7f37c69a3cd68ce65e45c373ab72a7f99153b83713f0f6dbace099825e56`.
- Both emitted manifests bind frozen executable source `sha256:a2835c668d57649585bb36bd5b211fa272ec642d30dbdf763c7dceef95424191`, current user HEAD `afd6c699980d4256af980a9abb906b2d79004755`, normal cache mode. Roots are `.cartulary/test-results/<exact-run-id>`; preserve native outer exits independently. No gate acceptance is claimed yet; independent work and cleanup must drain on failure.
- `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-markdown-20261018-01 make lint-markdown`, native 53706, observed outer 0, 37720 ms. Summary `adhoc/lint-markdown/tool-run-summary.json` under its exact root. Source/parity/finalizer checks are already applicable; no duplicate source-check run is needed without new executable changes.
- Complete release has not started. Required exact owner evidence and full observability audits remain pending, as does W11. No performance qualification or baseline publication remains in the acceptance scope.


### Frozen validation observation and audit input preparation (E-153)

- Latest permitted full-assurance observation: 34/58 units passed, no failed units; the installed-runtime boundary has completed successfully within this concurrent run. Real human/machine checks remain active. The parent owning exit is still pending.
- CI's canonical observation reports `publication_mode=paused_measurement` (last snapshot 62 passed, no failures; sample elapsed 375626 ms). Stop structured polling, including sibling status launches, during this known quiet interval. Observe native sessions 25581/23399 only until an explicit boundary or terminal exit permits further diagnosis. A paused snapshot is not current progress or liveness evidence.
- Prepared, but did not execute, owner audit inputs under `/tmp/cartulary-remediation-final-owner-inputs/<owner>.json` for harness.command_surface, harness.browser, harness.test_catalog, harness.generated_artifacts, harness.evidence_accounting, harness.release and platform.telemetry. Entries are unique ASCII-sorted catalog target partitions and point to the planned exact release root `.cartulary/test-results/remediation-w10-candidate-release-20261018-01`. This root has no qualifying evidence yet; release has not started. The eventual Make audit must validate the actual final source, successful complete run, projections, rows and cleanup. Preparation is not validation.
- Executable source remains frozen at E-151; original failed CI and private recovery proof remain preserved. W10/overall IN_PROGRESS, W11 TODO.


### Full-assurance terminal failure and admission diagnosis (E-154)

- `remediation-w10-candidate-full-20261018-01` native 25581 finished with observed outer Make 2 / normalized 4; graph duration 1687834 ms, 57/58 passed, one failed. All 66 cleanup records completed. Source remains `sha256:a2835c668d57649585bb36bd5b211fa272ec642d30dbdf763c7dceef95424191`.
- The only failure is `operation:harness.assurance_output_lint`, `infra/resource_conflict`, exit 4. Its retained target summary `target-summaries/harness.assurance_output_lint.json` records 600003 ms exclusive time against its declared 600000 ms admission allowance, plus 420973 ms earlier graph-resource queueing. No unit log/artifact directory or real-lint invocation exists. The native worker admits before invoking the runner; `nativeAdmission` applies the unit timeout and maps unfulfilled host admission to this class/reason. This evidence supports host-admission deadline exhaustion before workload launch, not a lint assertion failure. Independent units and cleanup continued to completion.
- Concurrent full assurance/CI was an execution choice in E-151, not an adopted requirement. F-027 / R6 corrects qualification sequencing: after the already-running CI drains, run full assurance and release sequentially. Preserve the current source, admission bounds, ownership and failure normalization. Do not patch product behavior, relax timeouts or relabel the failed invocation. Full assurance remains incomplete until a complete qualifying invocation succeeds. The retry is justified by the evidenced admission exhaustion and changed concurrency condition, not an unexplained repeat.
- Both real check assertions passed: `unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check` is 800/800, 1016566 ms, all 907 cleanup records completed; machine counterpart `operation-harness.assurance_output_check_machine/.../real-check-machine` is 800/800, 906301 ms, all 782 cleanup records completed. Real fast is 582/582, 117387 ms, all 19 cleanup records completed. Their owning output assertions require child Make success and passed; no separate native child outer handle was observed. Exact current-source audits remain required before acceptance; the unrelated parent admission failure does not itself invalidate these successful child invocations.
- `make explain-run RESULTS_DIR=.cartulary/test-results/remediation-w10-candidate-full-20261018-01`, `... DETAIL=performance JSON=1`, and `... DETAIL=accounting` returned outer 0. Saved descriptive reader outputs: `/tmp/cartulary-remediation-w10-full-failure-performance.json` and `/tmp/cartulary-remediation-w10-full-failure-accounting.txt`. The performance view omits 38 units by its documented top-20 bound, so the exact retained target summary supplies the failed unit timing. An attempted `TARGET=harness.assurance_output_lint DETAIL=accounting` was rejected with usage/outer 2 because TARGET accepts declared command names; it launched no work.
- CI native 23399 remains active and executable source stays frozen. Its last known structured state was paused_measurement; continue native observation, with no repeated CI status launches until an explicit boundary or terminal exit. Terminal diagnosis above concerns the completed full run, not CI progress. No new test or release run has started. W10/overall IN_PROGRESS; W11 TODO.


### Frozen-candidate CI completion and sequential qualification entry (E-155)

- `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-ci-20261018-01 make ci`, native 23399, observed outer 0. Graph pass 1202/1202, no failed/skipped/cancelled units, 4055607 ms; all 1228 cleanup records completed. Command ci.v4, graph `sha256:3c7f7f37c69a3cd68ce65e45c373ab72a7f99153b83713f0f6dbace099825e56`, source `sha256:a2835c668d57649585bb36bd5b211fa272ec642d30dbdf763c7dceef95424191`. Exact root `.cartulary/test-results/remediation-w10-candidate-ci-20261018-01`. `make explain-run RESULTS_DIR=<that exact root>` returned outer 0 and confirms complete graph/cleanup. Original failed CI E-137 remains failed historical evidence.
- CI's owning command has exited, ending the observation restriction. No executable source changed during the concurrent validation. The W10 repair paths now have complete CI evidence in addition to their focused checks; complete release remains outstanding.
- R6 enters IN_PROGRESS after E-150/R5 and the completed CI drain. Requalify the full assertion inventory with a fresh invocation and no overlapping CI/release command. The adopted public grammar exposes the real-output assertion union through SUITE=full; it has no private-operation execution input. This run closes the unexecuted mandatory lint-output assertion without introducing a new selector, alias or deadline change. Successful current-source check evidence from E-154 remains reusable subject to its exact audit.
- Planned command: `CARTULARY_TEST_RUN_ID=remediation-w10-r6-full-20261018-01 make harness-contract SUITE=full`. Record its native handle and terminal result before R6 closure and release entry. W10/overall IN_PROGRESS; W11 TODO.


### Sequential full-assurance invocation (E-156)

- Started `CARTULARY_TEST_RUN_ID=remediation-w10-r6-full-20261018-01 make harness-contract SUITE=full`, native 29335, exact root `.cartulary/test-results/remediation-w10-r6-full-20261018-01`. No CI/release invocation overlaps it. Its emitted manifest binds unchanged candidate source `sha256:a2835c668d57649585bb36bd5b211fa272ec642d30dbdf763c7dceef95424191` and the same full graph `sha256:a1ce5b6f71d3b6b3f7c993d15c276f7d3b2f30df36766191c2a421e0552da33b`.
- This is a fresh invocation after the diagnosed E-154 admission exhaustion and completed E-155 CI. Existing admission bounds, private operations, assertion inventory and source are unchanged. Await the owning exit and complete cleanup before closing R6 and starting release. W10/overall remain IN_PROGRESS; W11 TODO.


### Sequential lint-output assertion passes (E-157)

- R6's current structured observation is active, 34/58 units passed, no failures (sample elapsed 269307 ms); the parent native 29335 remains running. Real human/machine checks and installed-runtime assurance are still active.
- `unit-results/operation-harness.assurance_output_lint.json` is now passed, exit 0, no missing outputs, semantic digest `sha256:a618fa2678d93a803c2aa5877c50ea761b4b21ebc627b2431471cd9bab3512fe`, identical to E-154's unlaunched unit. This closes that assertion's execution gap under the corrected concurrency condition without source or timeout changes. R6 remains IN_PROGRESS until the complete owning command and cleanup succeed; no partial full-suite acceptance is claimed.


### Sequential full assurance PostgreSQL cleanup failures (E-158)

- E-156's owning native session 29335 is still active on source `sha256:a2835c668d57649585bb36bd5b211fa272ec642d30dbdf763c7dceef95424191`. Parent observation: 56/58 passed, two real-check assertions running. The lint-output assertion remains passed; no CI/release command overlaps. Do not cancel independent work or change executable source before the owning exit.
- Exact machine-check root: `.cartulary/test-results/remediation-w10-r6-full-20261018-01/unit-artifacts/operation-harness.assurance_output_check_machine/real-command-output/runs/real-check-machine`. Its live observation reports two failed units: `go:08d8744205c2234b:31829442b4d5` and `go:2f89aed0bdc82556:de8b3c1351b1`, both normalized 12 / harness/cleanup_error with complete declared row outputs.
- Exact failed rows are `module.reference_data.integration.canonical_persistence` (156090 ms runner duration) and `module.database_migrations.integration.cryptographic_state_admission` (72850 ms runner duration). Both retain pgtest cleanup-only diagnostics: service postgres, delete, attempt 1, prior_failure=false, stage forced_drop, cause deadline_exceeded. These durations cover their rows, not just cleanup. The normal/forced operation budgets remain 5/15 seconds.
- Bounded unit stderr confirms the migration failure in `TestCryptoAdmissionFreshConcurrentAndReadOnly_Integration`; no assertion failure preceded its cleanup marker. Physical database wait, blocking process, and relation to historical cleanup failures remain unproven. Current stage diagnostics do not distinguish lock, checkpoint, barrier or IO waits; do not guess one or increase deadlines.
- F-028 is a required follow-up after collecting the R6 owning exit. Review the existing exact-command safe PostgreSQL wait observer for reusable diagnostics, preserve exact ownership and cancellation, and obtain current owner guidance before any focused characterization. No retry, service reset, recovery deletion or source mutation has occurred. R6/W10/overall remain IN_PROGRESS; release and W11 have not started.


### Sequential full assurance owning exit and cleanup follow-up (E-159)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r6-full-20261018-01 make harness-contract SUITE=full`: native 29335 completed with observed outer Make 2 / normalized 10, 1168465 ms, 57/58 units passed. Sole failed assertion is machine check output; all 67 parent graph cleanup records completed. `make explain-run RESULTS_DIR=.cartulary/test-results/remediation-w10-r6-full-20261018-01` returned outer 0.
- Human check child passed 800/800 in 1038832 ms with all 462 graph cleanup records completed. Machine child passed 798/800 in 1016419 ms; its two E-158 row cleanup failures remain harness/cleanup_error (12), while all 462 graph-level cleanup records completed. Row database deletion and graph provider cleanup are distinct boundaries. Both children used 621 cache hits and 179 bypasses; do not describe them as fresh execution of every unit. Their owning assertions enforce the nested command exits; no independent native child session was observed.
- F-027's lint admission assertion now passes on unchanged source, but R6's complete-assurance criterion remains open because of F-028. Decompose the remaining R6 work into W10-R6a: exact PostgreSQL cleanup diagnostics/characterization and evidence-backed correction, checkpointed separately before renewed complete qualification. No new timeout, capacity, retry or cleanup policy is adopted by this failed run. The preceding source freeze has ended only after the owning exit.
- Begin with current evidence-accounting, database-migrations and reference-data guidance. Reuse the existing bounded structural PostgreSQL wait observer at its proper common cleanup owner if diagnostics must change; retain closed outputs, original error identity, exact target ownership and bounded teardown. Characterize any physical correction before broad validation. Required finalizer and affected final-source acceptance must follow executable changes. R6/W10/overall remain IN_PROGRESS; release and W11 remain unstarted.


### W10-R6a shared PostgreSQL wait diagnostics (E-160)

- Current owner guidance completed outer 0: evidence-accounting native 21017, database-migrations 9988, reference-data 94554, browser 65514 and command-surface 60546. Exact guidance is retained at `/tmp/cartulary-remediation-w10-r6a-{evidence,migrations,reference,browser,command}-guide.txt`.
- Amended the existing normalized cleanup-diagnostic clause before its projection. Moved the browser-only exact-drop wait observer to `internal/testutil/postgrescleanup`; both normal and forced attempts now share it, and browser cleanup no longer owns a duplicate observer. It opens no diagnostic connection until a drop exceeds one second, bounds probes to one second and the attempt context, and joins the observer before return. No operation deadline, admission, fallback, ownership or retry policy changed.
- The current unreleased v2 test-failure envelope adds an optional closed `wait` object: sampled/unavailable/not_observed and at most first/last structural wait/checkpointer snapshots, bounded blocker and cumulative-counter values. Unknown extension text never escapes; invalid observations become unavailable. The original error remains wrapped for errors.Is/As; bounded stage/cause traversal and primary failure classification remain unchanged. No historical schema is translated and no Markdown is executable input.
- Red command: `CARTULARY_TEST_RUN_ID=remediation-w10-r6a-diagnostic-red-20261018-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.unit.testfailure_helper_contract`, native 99082, outer Make 2 / normalized 12, 696 ms, all 10 graph cleanup records completed. The exact intended assertion was missing wait detail from the marker. Its failure message echoed the synthetic marker, so the runner classified that failed red as cleanup; corrected the assertion message to avoid emitting a fixture marker. This was not an actual database deletion failure.
- `...=remediation-w10-r6a-format-20261018-01 make format`: native 76234, outer 0, 8260 ms. `...=remediation-w10-r6a-generate-20261018-01 make generate`: native 57326, outer 0, 10900 ms; validators regenerated through Make.
- Green envelope/redaction row: `...=remediation-w10-r6a-diagnostic-green-20261018-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.unit.testfailure_helper_contract`, native 11574, outer 0, 1/1, 965 ms. PostgreSQL and lifecycle contracts: `...=remediation-w10-r6a-postgres-contract-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.recovery_drop_observation,harness.browser.integration.postgres_cleanup_target_scoped_coordination,harness.browser.unit.recovery_fixture_server_lifecycle,harness.browser.unit.pgtest_helper_contract`, native 2474, outer 0, 6/6, 20921 ms. The blocked-drop subtest retains the original deadline and exact Lock/object observation; existing successful release, cancellation, fallback, isolation and redaction assertions remain. Source for these passes: `sha256:811766984d2327bc05a64ffe7cc4aa61343ef449e3dac4c823f6c980795f702b`.
- Added Go-adapter positive retention and malformed/private wait rejection. Current source after this test addition: `sha256:c8658cfb43495302529445aebae7382214e929266a4df83c8bb91805af675b88`. Active fresh focused runs: `remediation-w10-r6a-row-adapter-20261018-01` (native 34308; command-surface public_registry_parity row), `remediation-w10-r6a-migrations-20261018-01` (native 67857; cryptographic_state_admission row), and `remediation-w10-r6a-reference-20261018-01` (native 78865; canonical_persistence row), each through `make test-slice OWNER=<recorded owner> ROWS=<exact recorded row>`. No final pass is claimed for active runs.
- The diagnostic gap is characterized; F-028's original physical wait remains unproven. Earlier CI/check evidence predates this executable change. R6a/R6/W10 remain IN_PROGRESS; finalizer and applicable full-source acceptance remain required.


### W10-R6a completion checkpoint (E-161)

- E-160 focused invocations completed outer 0: adapter native 34308, 58953 ms, 2/2 and all 11 cleanup records; migrations 67857, 19816 ms, 3/3 and all 14 cleanup; reference-data 78865, 53018 ms, 3/3 and all 14 cleanup. The original rows pass with shared diagnostics. This does not establish the cause of their earlier timeouts. E-160 green envelope and physical contracts had all 10/19 cleanup records completed respectively.
- Finalized schema attribution: wait details are permitted only for normal_drop/forced_drop, and unknown observation state, missing snapshots, private event strings, negative/oversized counters or unrelated stages fail closed. Added actual Go-adapter preservation/rejection coverage; no second result reader or event reducer. Existing unreleased v2 marker shape remains the current cutover contract with an optional validated diagnostic field.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r6a-format-20261018-02 make format`: native 16447, outer 0, 8155 ms. `...=remediation-w10-r6a-generate-20261018-02 make generate`: native 88667, outer 0, 10580 ms. Authored schema was regenerated through Make; no generated file was hand-edited. `git diff --check` passed.
- Final focused validation, source `sha256:b180dc8026778f0087c935d4b0359d88a45b9e4b87f926561618b22a00097777`: `...=remediation-w10-r6a-diagnostic-green-20261018-02 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.unit.testfailure_helper_contract`, native 71963, outer 0, 1/1, 58398 ms, all 10 cleanup; `...=remediation-w10-r6a-row-adapter-20261018-02 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity`, native 27775, outer 0, 2/2, 58024 ms, all 11 cleanup; `...=remediation-w10-r6a-source-20261018-01 make source-check CHECK=all`, native 92456, outer 0, 9/9, 73102 ms, all 18 cleanup.
- R6a is DONE: its bounded diagnostic gap is repaired and characterized. No physical timeout remedy is claimed; F-028's original wait cause remains unknown. R6 remains IN_PROGRESS until complete assurance closes. Next finalizer must use RESULTS_DIR unset because no eligible full warm check exists on this changed source; inspect mutations, then freeze executable source and requalify the full context sequentially. A recurrence must be investigated through the new exact diagnostics before another attempt. Release and W11 remain unstarted.


### Post-diagnostic finalization and candidate freeze (E-162)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r6a-finalize-20261018-01 env -u RESULTS_DIR make agent-finalize`: native 99815, observed outer 0, graph 25609 ms. Reviewed `unit-artifacts/finalize-summary.json` (v4): pass, generated unchanged, updated_file_count=0, updated_files=[], rollback not needed. Schema, catalog, tier and post-generation validation all retain source `sha256:b180dc8026778f0087c935d4b0359d88a45b9e4b87f926561618b22a00097777`. There are no finalizer-caused source mutations. Retained maintenance was skipped because RESULTS_DIR was unset; no eligible same-source full warm check existed at this finalization.
- `...=remediation-w10-r6a-markdown-20261018-01 make lint-markdown`: native 34955, outer 0, 34180 ms; summary `adhoc/lint-markdown/tool-run-summary.json`. This validates the amended spec and durable handoff. Subsequent human checkpoint edits do not alter executable source identity.
- R6a's completion checkpoint precedes broader validation. Freeze source b180dc80 for full assurance with the characterized shared observer, no overlapping CI/release and unchanged cleanup/admission limits. The new invocation validates the changed diagnostic boundary in the previously failing context; it is not an unexplained repeat. Preserve all earlier failures and physical-cause limits. R6/W10/overall IN_PROGRESS; release and W11 unstarted.


### Instrumented complete-assurance entry (E-163)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r6-diagnostic-full-20261018-01 make harness-contract SUITE=full` is active in native session 85605. Exact root `.cartulary/test-results/remediation-w10-r6-diagnostic-full-20261018-01`; command `cartulary.harness.command.harness_contract.v3`, source `sha256:b180dc8026778f0087c935d4b0359d88a45b9e4b87f926561618b22a00097777`, graph `sha256:a1ce5b6f71d3b6b3f7c993d15c276f7d3b2f30df36766191c2a421e0552da33b`. Manifest confirmed.
- This is the only active qualification command. Keep source frozen, use exact canonical status/native observation, stop structured polling during any declared measurement pause, and collect owning exit after all independent work and cleanup drains. Any repeated PostgreSQL failure must be investigated through the bounded current-row wait observation before another attempt. Full outcome, current-source CI, complete release, exact audits and W11 remain required.


### W10-R6 complete-assurance checkpoint (E-164)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r6-diagnostic-full-20261018-01 make harness-contract SUITE=full`: native 85605 completed with independently observed outer 0, 1496729 ms, 58/58 units passed, no failed/skipped/cancelled units. All 67 graph cleanup records completed. Source remained `sha256:b180dc8026778f0087c935d4b0359d88a45b9e4b87f926561618b22a00097777`; graph `sha256:a1ce5b6f71d3b6b3f7c993d15c276f7d3b2f30df36766191c2a421e0552da33b`. `make explain-run RESULTS_DIR=<exact full root>` returned outer 0.
- Same-source child evidence: human `unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check`, 800/800, 1063734 ms, all 915 cleanup records; machine `operation-harness.assurance_output_check_machine/real-command-output/runs/real-check-machine`, 800/800, 1004184 ms, all 740 cleanup; fast `operation-harness.assurance_output_test_fast/real-command-output/runs/real-test-fast`, 582/582, 314602 ms, all 125 cleanup. Every child root is beneath the exact full root. Owning output assertions passed and enforce nested Make success; no independent native child handle was observed.
- Cache accounting is explicit: full 58 bypasses; human 179 bypass / 168 hit / 453 miss; machine 179 bypass / 343 hit / 278 miss; fast 6 bypass / 466 hit / 110 miss. These are valid gate outcomes, not fresh execution or timing samples for every unit. Reuse the exact successful human check for check acceptance subject to the required full evidence audit; do not run another check solely to populate a list.
- F-027 closes through sequential admission and the passing lint-output assertion. F-028's lost-diagnostic boundary closes through R6a and full-context validation: both previously failed rows now pass in both real checks; exact normal/forced ownership, deadlines, fallback and primary failure remain unchanged. The historical physical timeout cause is still unproven; this pass does not rewrite failed runs or claim a causal timing remedy. Preserve their evidence.
- R6 is DONE and checkpointed before current-source CI entry. All mandatory W10 repair slices are complete. Full assurance/current-source check acceptance now has execution evidence; current-source CI, complete release, exact-run audits/observability and W11 remain open. No harness performance qualification/publication is required or claimed.


### Current-source CI entry (E-165)

- After R6 checkpoint E-164, launched `CARTULARY_TEST_RUN_ID=remediation-w10-r6-candidate-ci-20261018-01 make ci`, native 43658. Intended exact root `.cartulary/test-results/remediation-w10-r6-candidate-ci-20261018-01`. Confirmed manifest: command `cartulary.harness.command.ci.v4`, source `sha256:b180dc8026778f0087c935d4b0359d88a45b9e4b87f926561618b22a00097777`, graph `sha256:3c7f7f37c69a3cd68ce65e45c373ab72a7f99153b83713f0f6dbace099825e56`; 1202 selected work units. Terminal outcome remains pending. No full-assurance or release command overlaps.
- Preserve native outer exit separately from canonical observations; stop structured polling on any declared measurement pause and allow all independent work/cleanup to drain. On success, checkpoint CI before starting fresh complete release `remediation-w10-candidate-release-20261018-01`. Prepared owner-audit input files still name that planned release root and are not validation results. W10/overall remain IN_PROGRESS; W11 has not started.


### Current-source CI measurement pause (E-166)

- E-165 CI is still active in native 43658. The canonical observation at revision 107 declared publication_mode=paused_measurement, sampled elapsed 210077 ms, 184/1202 passed with no failures observed. These counts are the last pre-pause publication, not continuing live progress; zero running in this snapshot is not process-liveness or completion evidence.
- Stop structured polling, including avoidable sibling readers, until an explicit boundary permits resumption or the owning native command exits. Continue native-session observation only. Source b180dc80 remains frozen, no release/audit command overlaps, and no terminal pass is claimed.


### Current-source CI terminal failure and renderer follow-up (E-167)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r6-candidate-ci-20261018-01 make ci`: native 43658 completed with independently observed outer Make 2 / normalized 11, 2638171 ms; 1200/1202 units passed, two failed, no skipped/cancelled units. All 1202 graph cleanup records completed. Source remains `sha256:b180dc8026778f0087c935d4b0359d88a45b9e4b87f926561618b22a00097777`; graph `sha256:3c7f7f37c69a3cd68ce65e45c373ab72a7f99153b83713f0f6dbace099825e56`. `make explain-run RESULTS_DIR=.cartulary/test-results/remediation-w10-r6-candidate-ci-20261018-01` returned outer 0. The owning exit ends the measurement observation restriction.
- Independent failure: `browser_group:webserver-backed:functional-support-default-timeline-find`, artifact/artifact_error, exit 11. Dependent `browser_target_summary:browser-e2e-webserver-backed` also failed. The group directory contains only `playwright-test-list.txt`; browser-group-result and renderer/TLS attestations are absent, despite unit missing_outputs being empty. Nine selected workbook find rows were infrastructure-failed before launch; inspected rows have zero duration.
- Bounded unit stderr reports browser-control WebSocket connection refused at the local endpoint. The private endpoint is deliberately omitted here. No physical cause is yet established; renderer death, lifetime ownership and endpoint readiness require investigation. Do not call this a product assertion failure or silently retry.
- W10-R7 enters IN_PROGRESS after the terminal checkpoint. Obtain fresh browser/workbook guidance, trace exact renderer ownership and retained diagnostics, and characterize an evidence-backed correction before broad validation. No reset, recovery deletion, retry or executable mutation has occurred for this finding. Full assurance E-164 remains passing evidence on the current source, but current-source CI and complete release acceptance remain open. Release has never started. W10/overall IN_PROGRESS; W11 TODO.


### W10-R7 renderer failure characterization (E-168)

- Owner guidance passed outer 0: `make task-guide OWNER=harness.browser` native 42118 and `make task-guide OWNER=module.workbook` native 76441; retained at `/tmp/cartulary-remediation-w10-r7-{browser,workbook}-guide.txt`.
- Inspected renderer acquisition: endpoint discovery verifies container running/log readiness, then the first remote-browser connection fails; cleanup removes the container before preserving its state. The raw exception loses its infrastructure classification at the group CLI fallback and may include the private WebSocket endpoint. No exact container state survives E-167, so its physical failure cause remains unproven. No speculative readiness delay, reconnect or resource-limit change is justified.
- Red characterization: `CARTULARY_TEST_RUN_ID=remediation-w10-r7-renderer-red-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_contract`, native 70724, outer Make 2 / normalized 10, 744 ms; 1/2 units passed, all 11 cleanup records completed. `make explain-run RESULTS_DIR=<exact root>` returned outer 0. The intended assertion reached missing infrastructure classification (`undefined` versus `infra`). Source `sha256:cf8bee84f1967e6db4743c55052c9cffb5033302572924409fb25c8623ecc62c`.
- R7's implementation is limited to the characterized diagnostic/classification boundary: preserve `infra/service_start_error`, perform one bounded exact-owned-container state inspection before removal, retain closed status/running/OOM/exit or unavailable, suppress raw private endpoint/error output, preserve original cause and secondary cleanup. No physical connection remedy, retry or deadline relaxation is claimed. Tests cover unavailable/malformed state and cleanup failure without exposing identities. Applicable focused browser execution, finalization and renewed complete acceptance remain required.
- Moved accumulated E-080 onward checkpoint entries into section 10, restoring the tracker's twelve-section structure without dropping historical evidence. The durable handoff retains the full synchronized snapshot. W10-R7/W10/overall IN_PROGRESS; W11 TODO.

### W10-R7 focused outcomes and generated refresh (E-169)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r7-renderer-green-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_contract`: native 46145, outer 0, 2640 ms, 2/2 passed and all 11 cleanup records completed. Exact owned-container inspection precedes removal; malformed/unavailable state never masks the primary failure, and cleanup failure remains secondary. The diagnostic excludes private endpoint, daemon text and process identity.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r7-find-20261018-01 make test-slice OWNER=module.workbook ROWS=module.workbook.browser.find_accessible_geometry,module.workbook.browser.find_authority,module.workbook.browser.find_edit_departure,module.workbook.browser.find_full_window,module.workbook.browser.find_literal,module.workbook.browser.find_query_creation,module.workbook.browser.find_scope_lifecycle,module.workbook.browser.find_source_editors,module.workbook.browser.find_virtualized`: native 75787, outer 0, 125789 ms. The exact previously unlaunched group completes. This pass does not establish E-167's physical cause. Both focused invocations bind source `sha256:c9ba02d6fe3319a7fbf0773c8fb14615409ce8f60518b52f4ffa86119f41ffd3`.
- `...=remediation-w10-r7-source-20261018-01 make source-check CHECK=all`: native 34930, outer Make 2 / normalized 10, 15941 ms, 7/9 passed, all 18 cleanup records completed. JSON-shape check explicitly reports stale generated topology input for the changed visual-renderer-lease module; generated-source drift row also fails. Source readers correctly do not repair it. After the active focused browser run exited, started `...=remediation-w10-r7-generate-20261018-01 make generate` to refresh through the canonical authoring path. No generated file is hand-edited.
- `make target-plan TARGET=release-check DETAIL=summary`: native 92379, outer 0; 1281 units, graph `sha256:83e02b2d6a527f45eadd0e7acfa0c0cc053c236f85c96e066df8dafd5e529274`, plan retained at `/tmp/cartulary-remediation-w10-r7-release-plan.txt`. Discovery is not release execution or acceptance. R7/W10/overall IN_PROGRESS; W11 TODO.

### W10-R7 diagnostic repair and finalizer checkpoint (E-170)

- Exact nine-row browser invocation E-169 completed 12/12 graph units and all 23 cleanup records. The red/green contract and exact physical group validate the changed failure boundary and unchanged normal path; they do not explain the original CI refusal.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r7-generate-20261018-01 make generate`: native 19621, outer 0, 10800 ms; summary `generate/tool-run-summary.json`. Regenerated topology render-index fingerprint from the authored module through Make. `...=remediation-w10-r7-source-20261018-02 make source-check CHECK=all`: native 53599, outer 0, 18505 ms, 9/9 and all 18 cleanup records. Earlier stale-source failure is preserved.
- New executable source is `sha256:ebf9397d006b49a119c219a7939fd7ab133a9e5becb353549cf15e49130500eb`. `...=remediation-w10-r7-markdown-20261018-01 make lint-markdown`: native 1020, outer 0, 27040 ms, summary `adhoc/lint-markdown/tool-run-summary.json`. `git diff --check` passed.
- `...=remediation-w10-r7-finalize-20261018-01 env -u RESULTS_DIR make agent-finalize`: native 51821, independently observed outer 0, graph 25728 ms. Reviewed `unit-artifacts/finalize-summary.json`: generated unchanged, zero updated files, rollback not needed; schema/catalog/refreshed validation digest matches ebf9397d. Retained maintenance skipped because RESULTS_DIR was unset; no eligible same-source full warm check existed at this finalization. No finalizer-caused source mutations.
- R7's characterized diagnostic/classification repair is DONE and checkpointed. F-029's original physical cause remains unknown, and complete CI acceptance remains unresolved. Freeze source ebf9397d for sequential complete assurance, CI and release. Prior-source E-164 full/check results and E-167 failed CI remain historical; do not relabel them. Requalifying the affected full context with the new bounded failure observation is justified; no timeout, retry, admission or browser connection policy changed. W10/overall IN_PROGRESS; W11 TODO.

### Renewed full-assurance entry and durable handoff retention (E-171)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r7-candidate-full-20261018-01 make harness-contract SUITE=full` is active in native session 14703. Manifest confirms command `cartulary.harness.command.harness_contract.v3`, source `sha256:ebf9397d006b49a119c219a7939fd7ab133a9e5becb353549cf15e49130500eb`, graph `sha256:a1ce5b6f71d3b6b3f7c993d15c276f7d3b2f30df36766191c2a421e0552da33b`, started 2026-10-18T04:54:26.353Z. Root `.cartulary/test-results/remediation-w10-r7-candidate-full-20261018-01`. Preserve source, permit independent work/cleanup to drain, and collect the owning exit before CI/release entry.
- Verified the durable handoff's complete tracker snapshot matches `temp/plan-notes.md`, and the tracker has twelve top-level numbered sections after relocating the accumulated session log. `git status --short -- docs/handoffs/testing-harness-remediation.md` reports the new document as untracked, not ignored; `git check-ignore temp/plan-notes.md docs/handoffs/testing-harness-remediation.md` lists only the temporary tracker. The handoff must be included in any later source handoff/commit; no commit was requested or created.
- This is implementation-readiness and retention evidence, not complete acceptance. W10/overall remain IN_PROGRESS; W11 TODO. No harness performance qualification or baseline publication is required.

### Renewed full assurance exposes PostgreSQL signal-barrier wait (E-172)

- E-171 parent native 14703 is still running on frozen source ebf9397d. Human check reports 800/800 passed with all 907 graph cleanup records completed (1011.53 s rounded diagnostic view). Fast child passed 582/582 with all 178 cleanup records (372.91 s rounded view). `make explain-run RESULTS_DIR=<each exact child root>` returned outer 0. Parent owning exit remains pending; preserve independent work and cleanup.
- Machine check reports one failed unit `go:5b4848a6c0fadceb:8f1f46438669`, harness/cleanup_error, normalized 12, no missing declared outputs. Exact failing row: `module.incidentbundles.integration.incident_bundle_export_uses_the_common_job_resou_a1e51b82d7`, row duration 28540 ms. Its pgtest marker records cleanup-only failure, prior_failure=false, delete attempt 1, forced_drop/deadline_exceeded.
- The new typed wait diagnostic is sampled: first and last both report IPC/ProcSignalBarrier, zero lock blockers, empty checkpointer wait fields, requested/done counters 297/296 then 301/300, buffers 36492 then 36537. This establishes the observed wait category, not which backend delayed acknowledgment or a relationship to any historical timeout. Raw database/process/endpoint identities are absent.
- The physical mechanism now has retained evidence for scoped investigation after the owning full command exits. Source remains frozen; no retry, reset, deadline change, cleanup recovery or implementation edit has occurred. Record F-030 / W10-R8 at the terminal checkpoint, obtain applicable owner guidance, and characterize a structural correction before renewed acceptance. W10/overall IN_PROGRESS; CI/release and W11 remain unstarted on this candidate.

### Full-assurance terminal checkpoint and PostgreSQL barrier investigation (E-173)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r7-candidate-full-20261018-01 make harness-contract SUITE=full`: native 14703 completed with observed outer Make 2 / normalized 10, 1434462 ms; 57/58 units passed. Sole failed parent assertion is machine-check output. All 67 parent graph cleanup records completed. `make explain-run RESULTS_DIR=<exact full root>` returned outer 0. Source remains ebf9397d; the source freeze ends after this owning-exit checkpoint.
- Machine child completed 799/800, normalized harness/cleanup_error, all 720 graph cleanup records completed; human child passed 800/800 with all 907 cleanup, fast child passed 582/582 with all 178 cleanup. Children live under the exact E-171 root at their recorded `unit-artifacts/operation-harness.assurance_output_<check|check_machine|test_fast>/real-command-output/runs/<real-check|real-check-machine|real-test-fast>` paths. Owning output assertions enforce nested Make exits; no independent native child handles were observed.
- Cache counts: full 58 bypasses; human 179 bypass / 176 hit / 445 miss; machine 179 bypass / 363 hit / 258 miss; fast 6 bypass / 413 hit / 163 miss. The E-172 row cleanup failure remains distinct from completed graph provider cleanup. Historical successes do not erase this failed context.
- F-030 / R8 is IN_PROGRESS after R7's checkpoint. The retained wait is IPC/ProcSignalBarrier, not a database lock-blocker inference. [Pinned PostgreSQL 18.6 barrier source](https://raw.githubusercontent.com/postgres/postgres/REL_18_6/src/backend/storage/ipc/procsignal.c) establishes cross-backend acknowledgment and periodic server diagnostics; [18.6 release notes](https://www.postgresql.org/docs/18/release-18-6.html) already include the known initialization-race fix. Do not attribute this failure to that fixed bug or change pins speculatively. The delayed participant and physical cause remain unproven.
- Obtain fresh browser, evidence-accounting and incident-bundle owner guidance before characterization or edits. Preserve exact target admission, eager cleanup, parent cancellation and separate operation bounds. No retry, new timeout, service reset or recovery deletion occurred. R7 diagnostic repair remains complete; current-source complete assurance, CI, release, exact audits and W11 remain required. W10/overall IN_PROGRESS; W11 TODO.

### W10-R8 isolated barrier characterization entry (E-174)

- Fresh owner guidance passed outer 0: browser native 21909, evidence-accounting 88666, incident-bundles 1587; files `/tmp/cartulary-remediation-w10-r8-{browser,evidence,incidentbundles}-guide.txt`. Exact child durations from run summaries are human 1011526 ms, machine 999614 ms, fast 372908 ms.
- Source review shows the failed incident-bundle row completed its job assertions before eager cleanup; the marker has prior_failure=false. PostgreSQL's cross-backend signal acknowledgment makes another participant a possible source of delay; the recorded wait alone does not identify it. The pinned server already includes the known initialization-race fix. No pin or concurrency-limit change is justified by the current evidence.
- Added a focused characterization subcase inside the existing target-coordination contract. It creates a separately owned PostgreSQL cluster, deliberately leaves one bounded-row producer unread in its administrative database until ClientWrite is observed, then checks whether dropping a different owned database reaches the retained ProcSignalBarrier wait. The exact peer is closed before ordinary owned cleanup. A separate cluster prevents this deliberately blocked result from disturbing borrowed suite services. Existing four-lane contract allocation remains sufficient; no production policy, deadline or retry changed.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r8-format-20261018-01 make format`: native 19596, outer 0, 10088 ms. `git diff --check` passed. Started `...=remediation-w10-r8-barrier-characterization-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.postgres_cleanup_target_scoped_coordination`; outcome pending. A passing characterization would establish this mechanism only, not the identity or cause of E-172's delayed participant.
- R8/W10/overall IN_PROGRESS; complete acceptance and W11 remain open. No blind rerun of full assurance or failed product row has occurred.

### W10-R8 isolated hypothesis outcome (E-175)

- `CARTULARY_TEST_RUN_ID=remediation-w10-r8-barrier-characterization-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.postgres_cleanup_target_scoped_coordination`: native 51813, observed outer Make 2 / normalized 10, 18786 ms; 2/3 units passed and all 14 graph cleanup records completed. Source `sha256:c0213e24f102da00192db45bf74e902a6ca1421c56e7a4f9ef9b330b104a9e62`; graph `sha256:77305559be90b3f85f11daaeb1a6f3022995a5141657e74eaa8a668796e5d6f8`. `make explain-run RESULTS_DIR=<exact root>` returned outer 0.
- The exploratory subcase reached ClientWrite, but the different owned database dropped successfully: the expected deadline was absent. This experiment does not reproduce the E-172 barrier timeout. Removed only the exploratory subcase and its pgconn import; retained the existing target/session coordination characterizations. No production correction is justified by this rejected hypothesis. No pins, operation budgets, admission limits or retry behavior changed.
- All physical fixture cleanup completed. Original E-172 delayed participant and cause remain unknown; failed evidence is preserved. R8/W10/overall remain IN_PROGRESS and W11 remains TODO. Further diagnosis must preserve the original failure and identify actionable participant evidence, rather than turning speculative timeout assumptions into permanent tests.

### Durable handoff checkpoint and remaining completion gate (E-176)

- Created and reconciled `docs/handoffs/testing-harness-remediation.md` as the durable human handoff requested by the user. It contains restart instructions, current acceptance limits, command/evidence/cache migration and rollback guidance, and the complete synchronized controlling tracker with all 190 original dispositions. `temp/plan-notes.md` remains the sole temporary controlling tracker. If it is absent in a future checkout, restore it from the embedded snapshot before resuming.
- Admission checks confirm regular single-link tracker/handoff files with no symlinked path components. Git reports the handoff as a new non-ignored file; only the temporary tracker is ignored. The new handoff must be included in any subsequent source commit. No commit or PR was requested or created.
- Corrected stale current-state acceptance claims in the handoff and top-level tracker. E-173 remains failed full assurance; E-175 rejected the exploratory unread-peer expectation. All original failed evidence and private recovery-proof references remain intact. No active native invocation remains after the documentation check.
- `CARTULARY_TEST_RUN_ID=remediation-durable-handoff-markdown-20261018-01 make lint-markdown`: native 14649, independently observed outer 0, 16090 ms; summary `.cartulary/test-results/remediation-durable-handoff-markdown-20261018-01/adhoc/lint-markdown/tool-run-summary.json`. `git diff --check` passed. Manual snapshot comparison confirmed exact equality and all twelve numbered tracker sections. No generation, finalizer or product gate was rerun for this documentation checkpoint; it establishes no additional functional or release acceptance.
- W11 completion status reviewed: TODO, pending W10's required successful applicable evidence and checkpoint. Creating the handoff is complete, but final candidate reconciliation, full diff/preservation review and release readiness are not complete. W10-R8 and W10 remain IN_PROGRESS; overall remains IN_PROGRESS. Resume with exact F-030 barrier-participant diagnosis; do not blindly retry assurance or treat the removed exploratory assertion as a fix. Complete required current-source assurance, CI, release and audits before W11 can close. Harness performance qualification and baseline publication remain removed by the user's decision.

### W10/W11 resumption and barrier diagnostic boundary (E-177)

- User explicitly requested completion of W10 and W11. Revalidated repository branch and HEAD `afd6c699980d4256af980a9abb906b2d79004755`, applicable guidance and current E-176 tracker/handoff. Existing implementation changes remain uncommitted and preserved. No active prior invocation remains.
- Fresh `make task-guide OWNER=harness.browser` (native 1386) and `make task-guide OWNER=harness.evidence_accounting` (native 28358) both returned outer 0; retained at `/tmp/cartulary-remediation-w10-r8-resume-{browser,evidence}-guide.txt`.
- R8 investigation continues from the preserved failure. Pinned PostgreSQL's `WaitForProcSignalBarrier` emits a LOG notice after a five-second wait identifying a delayed participant. Existing observations omit that participant. Add bounded notice capture only to the owned drop session, use its transient identity for read-only activity sampling, and retain only closed backend/state/wait classifications. Raw notices, SQL and process identity must never enter retained evidence. No global catalog exclusion, retries, deadline relaxation or pin change.
- Characterize the diagnostic with a deliberately stopped backend in a separately owned fixture cluster, restoring that backend before owned teardown. Verify redaction, unavailable state and original failure preservation. This establishes the diagnostic mechanism, not the historical failure's cause. Renew applicable acceptance only after focused validation and finalization. R8/W10/overall IN_PROGRESS; W11 TODO.

### W10-R8 participant observation characterization and focused validation (E-178)

- Red: `CARTULARY_TEST_RUN_ID=remediation-w10-r8-participant-red-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.postgres_cleanup_target_scoped_coordination`, native 47626, outer Make 2 / normalized 10, 25960 ms, 2/3 passed, all 14 cleanup steps completed. Source `sha256:4d33c7b895cc39c5ea106a28e464a1dfaff36853e8fcdf0bf545826356e0c98f`. The separately owned paused backend caused the intended ProcSignalBarrier wait; the assertion reached missing redacted participant metadata. The exact backend was resumed before teardown.
- Adopted bounded participant observation under the existing cleanup-diagnostic clause. The owned drop session captures only recognized pinned-server barrier notices; transient identity is used for a bounded read-only activity query. Evidence retains closed status/backend/state/wait values and bounded blockers, never PID, notice, SQL or connection identity. Unrecognized notices are ignored and failed sampling is explicitly unavailable. Config registration is pgx's existing scoped connection mechanism, unregistered after the operation pool closes. Normal/forced admission, budgets, primary errors and fallback policy are unchanged.
- Format native 6155 returned outer 0 (8175 ms). `...=remediation-w10-r8-participant-generate-20261018-01 make generate`: native 26067, outer 0 (10790 ms); generated schema projections refreshed through Make. No generated file was hand-edited.
- Green exact coordination row: `...=remediation-w10-r8-participant-green-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.postgres_cleanup_target_scoped_coordination`, native 57241, outer 0, 28451 ms, 3/3 passed and all 14 cleanup steps complete. Envelope row: `...=remediation-w10-r8-participant-envelope-20261018-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.unit.testfailure_helper_contract`, native 74979, outer 0, 4798 ms, 1/1 and all 10 cleanup complete.
- Fresh command-surface and incident-bundle guidance returned outer 0 (native 88095/86395). Adapter contract: `...=remediation-w10-r8-participant-adapter-20261018-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity`, native 71941, outer 0, 59943 ms. Exact affected product row: `...=remediation-w10-r8-participant-incident-20261018-01 make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_uses_the_common_job_resou_a1e51b82d7`, native 1355, outer 0, 22722 ms. This does not explain the historical participant's delay.
- `...=remediation-w10-r8-participant-source-20261018-01 make source-check CHECK=all`: native 60355, outer 0, 16629 ms, 9/9, all 18 cleanup complete. `...=remediation-w10-r8-participant-markdown-20261018-01 make lint-markdown`: native 75758, outer 0, 32820 ms. `git diff --check` passed. Current source `sha256:07ab7a4ff0fd570002c1613b63d190dca7037d9978d4f5eaefe8d261aa00d48b`.
- Started `...=remediation-w10-r8-participant-finalize-20261018-01 env -u RESULTS_DIR make agent-finalize`. RESULTS_DIR is unset because no eligible successful full warm check exists for this source. Review its result/mutations before renewed full-context acceptance. R8's diagnostic boundary is characterized; historical physical cause remains unknown. R8/W10/overall IN_PROGRESS; W11 TODO.

### W10-R8 finalizer checkpoint and complete assurance entry (E-179)

- Finalizer native 47494 completed outer 0, 25217 ms. Reviewed `.cartulary/test-results/remediation-w10-r8-participant-finalize-20261018-01/unit-artifacts/finalize-summary.json`: generated unchanged, zero mutations, rollback not needed, refreshed source digest `sha256:07ab7a4ff0fd570002c1613b63d190dca7037d9978d4f5eaefe8d261aa00d48b`. Retained maintenance was skipped because RESULTS_DIR was unset; no eligible same-source full warm check existed at finalization.
- E-178 adapter invocation completed 2/2 with all 11 cleanup steps; exact incident-bundle invocation completed 3/3 with all 14 cleanup steps. Canonical explanations returned outer 0. The new participant observation is physically characterized and rejects private/malformed values; this does not establish the original E-172 participant or root cause.
- Began `CARTULARY_TEST_RUN_ID=remediation-w10-r8-candidate-full-20261018-01 make harness-contract SUITE=full` after finalization, native session 52934. Keep source 07ab7a4f frozen through this invocation, then collect the independent owning exit and investigate any failures before CI/release. No CI or release overlaps the full context. R8/W10/overall remain IN_PROGRESS pending applicable complete acceptance; W11 TODO.

### W10-R8 full-context recurrence with participant evidence (E-180)

- E-179 full assurance remains active in native 52934; source 07ab7a4f is frozen. Machine check reported two cleanup-only failures while independent work continues; no owning full exit is yet available. Human check's last observed publication was 735/800 passed with zero failures. Do not start CI/release or edit executable inputs before the owning command drains.
- Exact machine child root: `.cartulary/test-results/remediation-w10-r8-candidate-full-20261018-01/unit-artifacts/operation-harness.assurance_output_check_machine/real-command-output/runs/real-check-machine`. Failed units `go:0eab78ed6e254ecb:b12e48365b78` and `go:102d5fe4ec38831e:8128a5ab32cf` retain harness/cleanup_error, normalized 12, missing_outputs empty.
- Failed rows are `module.recovery.integration.selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a` (80240 ms row duration) and `module.evidence.integration.background_verification_cancellation` (24510 ms). Both retain pgtest/delete/attempt1, prior_failure=false, forced_drop/deadline_exceeded. Recovery's first wait was CheckpointDone with checkpointer IO/DataFileSync; last wait was ProcSignalBarrier. Evidence's first and last waits were ProcSignalBarrier.
- The new diagnostic identifies the delayed participant at the last sample in both cases as a client backend, state active, empty wait type/event and zero lock blockers. The recovery checkpointer was then at Activity/CheckpointerMain. This narrows the observed delay beyond the prior unidentified participant; it does not establish the active statement or physical cause, nor prove that both notices named the same participant. No process identity or SQL is retained.
- Continue bounded diagnosis and permit all cleanup to drain. Preserve the full failed context, obtain affected owner guidance, and investigate the active participant's operation category before choosing a structural correction. No retry, reset, capacity/deadline change or cancellation has occurred. R8/W10/overall IN_PROGRESS; W11 TODO.

### W10-R8 terminal recurrence checkpoint (E-181)

- E-179 full command native 52934 completed with observed outer Make 2 / normalized 10, 1436764 ms. It passed 57/58 parent units; the sole failed parent is machine-check output. This is failed acceptance and is not eligible for CI/release progression.
- Human check passed 800/800 with all 962 graph cleanup steps complete. Machine check completed 798/800, normalized harness/cleanup_error, with all 648 graph cleanup steps complete. Canonical explanations returned outer 0. The two E-180 per-test database failures remain separate from graph cleanup success; source remains 07ab7a4f.
- Source inspection confirms the next missing observation is the delayed client's operation category. Retain a closed SQL operation category only, derived in memory, plus bounded observation context if needed; never retain raw statements, process/database identities or arbitrary application labels. Characterize an active owned backend rather than infer that an empty wait field means CPU or filesystem delay. The existing observation remains primary-error-preserving and no cleanup policy changes are justified yet.
- Historical causal limits remain explicit. Obtain fresh recovery/evidence owner guidance and extend the focused physical/redaction boundary before any further acceptance reproduction. No destructive recovery, timeout adjustment, service reset or blind retry. R8/W10/overall IN_PROGRESS; W11 TODO.

### W10-R8 operation-category diagnostic entry (E-182)

- Fresh recovery and evidence-product guidance returned outer 0 (native 23696/24557); retained at `/tmp/cartulary-remediation-w10-r8-{recovery,evidence-product}-guide.txt`. E-181 is terminal; executable editing resumed only afterward.
- Extended the bounded participant snapshot with a closed statement-operation category, administrative/target/other/none database scope, and statement age clamped to one day. Classification discards statement text in memory. Unknown syntax remains `other`; invalid retained values fail closed. No statement, database name, process identifier, connection string or arbitrary notice is retained.
- Physical characterization now stops an exact separately owned backend only after its active bounded sleep statement is observed; resume and owned cleanup remain mandatory. Added invalid operation/scope/age rejection to the Go envelope and schema adapter assertions. The adopted observation clause and authored v2 schema match; generated validators remain Make-owned.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r8-operation-format-20261018-01 make format`: native 4406, outer 0, 7910 ms. `...=remediation-w10-r8-operation-generate-20261018-01 make generate`: native 8745, outer 0, 10500 ms. Current source `sha256:ec4e473523e14497ccdccfc1d3c0a50985c84c3b2f9e904ff053bfbc3ab02c5d`.
- Focused physical, envelope and adapter checks are active in native 51252/28915/45409 at `.cartulary/test-results/remediation-w10-r8-operation-{physical,envelope,adapter}-20261018-01`. R8 remains IN_PROGRESS; this is diagnostic characterization, not a physical timeout correction. No admission, deadline, retry or pin change. W10/overall IN_PROGRESS; W11 TODO.

### W10-R8 operation diagnostic validation and source freeze (E-183)

- E-182 focused invocations all returned independently observed outer 0: physical native 51252, 85199 ms, 3/3 and all 14 cleanup complete; envelope native 28915, 59674 ms, 1/1 and all 10 cleanup; adapter native 45409, 59130 ms, 2/2 and all 11 cleanup. Physical assertion confirmed the paused active participant is categorized as select/administrative without retaining its statement or identity. Admission wait contributes to these descriptive durations; no performance claim is made.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r8-operation-source-20261018-01 make source-check CHECK=all`: native 7323, outer 0, 17373 ms, 9/9 and all 18 cleanup. `...=remediation-w10-r8-operation-markdown-20261018-01 make lint-markdown`: native 46596, outer 0, 12050 ms. `git diff --check` passed.
- `...=remediation-w10-r8-operation-finalize-20261018-01 env -u RESULTS_DIR make agent-finalize`: native 2893, outer 0, 25954 ms. Reviewed v4 finalizer receipt at that root's `unit-artifacts/finalize-summary.json`: generated unchanged, zero mutations, rollback not needed, source ec4e473523e14497ccdccfc1d3c0a50985c84c3b2f9e904ff053bfbc3ab02c5d. Retained maintenance skipped because RESULTS_DIR was unset; no eligible current-source warm check existed.
- Renew complete assurance with this characterized diagnostic, preserving the failing concurrent wrapper context and all cleanup bounds. Started exact command `CARTULARY_TEST_RUN_ID=remediation-w10-r8-operation-full-20261018-01 make harness-contract SUITE=full`, native 72524. Confirmed source ec4e4735 and graph a1ce5b6f71d3b6b3f7c993d15c276f7d3b2f30df36766191c2a421e0552da33b. Freeze executable source until owning exit. Do not overlap CI/release. This is a new observation boundary, not a claimed physical correction or relabeling of E-181. R8/W10/overall IN_PROGRESS; W11 TODO.

### W10 retained worker-budget review finding (E-184)

- During frozen-source R8 assurance, source review found a separate evidence-reconstruction gap: `prepare-plan.mjs` constrains Go planning through `availableGoWorkerBudget(snapshot.cpu_tokens, process.env)` under an admitted parent, but `retained-plan.mjs` reconstructs from `manifest.capability_snapshot.cpu_tokens`. The current manifest retains physical capacity and omits the resolved worker budget. Current nested human check retains physical CPU 19; its enclosing assurance operation admits CPU 4. No source was edited during this review.
- F-031 is a candidate W10-R9 repair: retain and validate the resolved planning budget separately from physical capability, reconstruct the same graph without reading ambient inherited state, and characterize direct/nested equivalence plus malformed/missing budgets. Confirm the graph mismatch through canonical evidence validation before implementing. Preserve the actual parent lease and physical capability; no fabricated capacity override or relaxed admission.
- R8 assurance remains active in native 72524. Its PostgreSQL result and all finalizers must settle first. R9 is TODO; it does not imply R8 completion or authorize editing frozen executable source. W10/overall IN_PROGRESS; W11 TODO.

### W10 retained-budget investigation audit (E-185)

- Fresh `make task-guide OWNER=platform.telemetry DETAIL=coverage JSON=1` returned outer 0 (native 49607), retained at `/tmp/cartulary-remediation-w10-retained-budget-telemetry-guide.json`.
- The real backend-unit wrapper published passed/exit0 and no missing outputs. Its exact child is `.cartulary/test-results/remediation-w10-r8-operation-full-20261018-01/unit-artifacts/operation-harness.assurance_output_backend_unit/real-command-output/runs/real-backend-unit`.
- `CARTULARY_TEST_RUN_ID=remediation-w10-retained-budget-red-20261018-01 make test-evidence-audit OWNER=platform.telemetry EVIDENCE_ROOTS_FILE=/tmp/cartulary-remediation-w10-retained-budget-telemetry-input.json`: native 57816, observed outer 0, 16445 ms. Despite the exploratory run-ID suffix, this passed: all eleven telemetry rows, including seven security rows, accepted under source ec4e4735 and graph cff59262452c7d6cb8e03e93229c80fd202fb078740793a1b58c756056b2355f. Audit report is at that invocation's `test-evidence-audit/audit.json`.
- This selection did not reproduce E-184's concern. The LPT planner derives per-process CPU 4 both from host19 with four workers and inherited4 with one worker when enough compatibility groups exist. The graph retains those per-process claims, so this particular reconstruction agrees. Single-group or other-capacity cases still need characterization before declaring the general reconstruction contract complete; do not call this passing evidence a failure. No executable change or expanded acceptance claim follows from this audit.
- Full assurance remains active in native 72524, with no observed failures. Earlier commentary overstated machine completion of two affected units; corrected immediately. Only the human check's unit results had passed at that time; the machine units are now executing. R8/W10/overall IN_PROGRESS, R9 candidate TODO, W11 TODO.

### W10-R8 actionable recurrence and framing gap (E-186)

- R8 full native 72524 is still draining; no owning exit yet. Human check completed and its wrapper passed. Machine check reported three failed units: `go:102d5fe4ec38831e:8128a5ab32cf`, `go:2f89aed0bdc82556:de8b3c1351b1`, `go:e8eccc6efddf32cc:b3f543dc9303`. Published classification is harness/scheduler_accounting_error (11), not cleanup_error. Preserve it.
- Failed rows include evidence handle issuance, migration cryptographic-state admission and Jobs transition/progress. Previously affected evidence background cancellation passed in the human and machine row results observed so far. No source mutation, cancellation or overlapping CI/release occurred.
- Bounded excerpts from the exact machine child's `unit-logs/go-2f89aed0bdc82556-de8b3c1351b1/stderr.log` and `unit-logs/go-e8eccc6efddf32cc-b3f543dc9303/stderr.log` retain the original cleanup envelope across Go output events. Both describe forced_drop/deadline_exceeded, IPC/ProcSignalBarrier, and a sampled active client backend executing `create_database` in administrative scope, age 19387/19300 ms. Checkpointer last observed Activity/CheckpointerMain; no lock blockers. This identifies an operation, not a particular SQL statement, PID, or continuous filesystem wait.
- The pinned Go stream splits the long envelope into `error` and `error-continue` Output events before the newline. Current `observeEvent` parses each Output independently, treating the incomplete first fragment as malformed. F-032 requires bounded per-test line reassembly in the existing Go adapter before envelope parsing, retaining original classification and rejecting truncated/oversized/cross-test input. Characterize the physical split and streaming/in-memory parity. No alternate parser or relaxed schema is justified.
- Pinned PostgreSQL FILE_COPY source has an explicit per-file fsync traversal without an interrupt check between those sync operations; CREATE DATABASE participation and the current template-copy policy make copy preparation the next structural investigation. Evaluate WAL_LOG plus an explicit preparation checkpoint, retaining physical copy completion before publication, existing drop budgets, exact target exclusion, and failed-acquisition cleanup. Do not introduce global catalog serialization or claim the historical physical syscall cause is proved.
- After owning exit, split R8 follow-up into R8a (Go envelope framing, first) and R8b (template preparation and cleanup, second), each with focused validation and a tracker checkpoint. R9 retained-budget characterization follows separately if confirmed. Source ec4e4735 remains frozen until the current full invocation ends. W10/overall IN_PROGRESS; W11 TODO.

### W10-R8 terminal checkpoint and R8a entry (E-187)

- Full native 72524 independently exited outer Make 2 / normalized 10, 1450444 ms. Parent 57/58 passed, sole failed output-machine-check assertion, all 67 cleanup complete. Canonical explanation returned outer 0. Source ec4e4735 freeze ends here.
- Exact child summaries: human 800/800, 1020455 ms, all 914 cleanup; machine 797/800, harness/scheduler_accounting_error, 983409 ms, all 716 cleanup; fast 582/582, 205383 ms, all 33 cleanup. Cache counts human 179 bypass/169 hit/452 miss; machine 179/367/254; fast 6/558/18. Every child is under E-183's exact root and previously recorded real-command-output path. Parent wrapper assertions observe nested Make outcomes; no separate native child handles are claimed.
- R8a/F-032 enters IN_PROGRESS after this owning-exit checkpoint. Add red characterization of split marker/JSON, interleaved tests, incomplete and oversized envelopes, and memory/file parser equivalence. Implement bounded newline framing in the existing Go adapter, retaining closed schema checks, primary/secondary failure precedence and no-output disclosure. Complete focused validation and checkpoint before R8b's clone change. No timeout or source pin change.
- R8b/F-030 remains TODO with the E-186 observed create_database participant. R9/F-031 remains a characterization candidate, not an observed failure of the successful audit. Required current-source complete assurance, CI, release and exact audits remain open. W10/overall IN_PROGRESS; W11 TODO.

### W10-R8a framing repair completion checkpoint (E-188)

- Fresh owner guidance returned outer 0 (native 2291/55263), retained at `/tmp/cartulary-remediation-w10-r8a-{surface,evidence}-guide.txt`.
- Red `CARTULARY_TEST_RUN_ID=remediation-w10-r8a-framing-red-20261018-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity`: native 61165, outer Make 2 / normalized 10, 56438 ms, 1/2, all 11 cleanup. The first assertion observed 11 instead of required 12 for a fragmented cleanup envelope; bounded unit stderr records the exact assertion. The source was not frozen through this exploratory red run: an additional pending-line-bound case was authored after launch; no final-source acceptance is claimed for it. The failing first assertion and production adapter were unchanged during that run.
- Implemented one bounded framing owner used by in-memory and streaming adapters. It reconstructs marker/JSON fragments per test and package until newline, retains only a marker-prefix tail before detection, caps envelopes at 64 KiB and unfinished lines at 128, and rejects incomplete/oversized data. Schema validation, redaction and existing primary/secondary failure precedence remain intact. Non-event JSON is ignored consistently; outer Go JSON events share the existing 1 MiB limit.
- Initial green native 55625 returned outer 0, 57268 ms. After adding non-event and cross-package cases, `...=remediation-w10-r8a-framing-format-20261018-01 make format` native 86411 returned outer 0, 7570 ms; `...=remediation-w10-r8a-framing-generate-20261018-01 make generate` native 31752 returned outer 0, 10720 ms. No generated projection was hand-edited.
- Final focused `...=remediation-w10-r8a-framing-verified-20261018-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity`: native 82609, outer 0, 59373 ms, 2/2 with all 11 cleanup. Source `sha256:47f7441d375ed7525710581357775d79783cd1c79979193d8b9edc042b8c6ccf`.
- `...=remediation-w10-r8a-framing-source-20261018-01 make source-check CHECK=all`: native 21425, outer 0, 17141 ms, 9/9 with all 18 cleanup. `...=remediation-w10-r8a-framing-markdown-20261018-01 make lint-markdown`: native 79958, outer 0, 34810 ms. `...=remediation-w10-r8a-framing-lint-20261018-01 make lint CHECK=scripts`: native 25266, outer 0, 440 ms. `git diff --check` passed. Roots follow `.cartulary/test-results/<exact-id>`.
- R8a/F-032 DONE. This repairs the diagnostic transport boundary; E-187 remains failed acceptance and PostgreSQL interference is not fixed. R8b may start after this checkpoint with fresh guidance. Finalization and complete acceptance follow the remaining executable repairs. R8/W10/overall IN_PROGRESS; W11 TODO.

### W10-R8b preparation contract and characterization entry (E-189)

- R8a checkpoint E-188 precedes this slice. Fresh browser, migration and Jobs guidance returned outer 0 (native 63698/53210/30323), retained at `/tmp/cartulary-remediation-w10-r8b-{browser,migrations,jobs}-guide.txt`.
- Adopted WAL_LOG plus an explicit preparation checkpoint under the same target-specific session admission. Physical capability, fsync/full-page writes/synchronous commit, normal/forced cleanup budgets, exact ownership, and unrelated-target admission remain unchanged. This removes dependence on the FILE_COPY per-file sync traversal identified during source investigation; the exact historical syscall remains unproven.
- Creation and preparation success become explicit separate outcomes. The common clone owner will report successful CREATE even if its checkpoint/release later fails. Each existing lifecycle owner must compensate only a proven-created target, preserving the primary preparation error and any secondary cleanup error. A pre-existing target must survive failed creation.
- Added a physical failed-preparation subcase to the existing template-preparation row: a fixture role can create databases but cannot checkpoint/provision roles; failure must preserve its permission cause and remove only the created destination. Existing checkpoint-LSN, clone isolation, existing-target protection and missing-template assertions remain. The existing target-coordination and full concurrent qualification remain required.
- R8b IN_PROGRESS; begin red characterization after formatting. No implementation strategy, timeout, resource capacity, pin or global catalog exclusion has changed yet. R8/W10/overall IN_PROGRESS; W11 TODO.

### W10-R8b preparation implementation checkpoint (E-190)

- First red `...=remediation-w10-r8b-preparation-red-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.postgres_template_preparation`: native 95637, outer Make 2 / normalized 10, 13081 ms, 2/3 with all 14 cleanup. It failed before the intended boundary because the fixture connection option encoded a space as `+`; source bffdfef7. Corrected the fixture option to the attached `-crole=...` form.
- Intended red with fresh ID `remediation-w10-r8b-preparation-red-20261018-02`, same command: native 95049, outer Make 2 / normalized 10, 13656 ms, 2/3 with all 14 cleanup; source bcaf17f4. It reached permission-denied preparation after creation and failed because the database remained present. Bounded unit stderr records `failed preparation leaked its created database`.
- Common clone preparation now uses WAL_LOG then CHECKPOINT on the target-admitted physical session, and returns successful CREATE independently of later failure. Pgtest, product-measurement fixture lifecycle, and recovery-browser callers dispose only proven-created failed acquisitions through the existing bounded cleanup owner, joining preparation and cleanup errors. Pgtest also disposes a freshly created database if role provisioning fails. Failed CREATE never authorizes deleting an existing target. No global serialization, new timeout, retry, pin or weakened durability setting.
- Format `remediation-w10-r8b-preparation-format-20261018-01` native 21376 outer 0, 8158 ms; after implementation `...format-20261018-02` native 80535 outer 0, 8026 ms. `...=remediation-w10-r8b-preparation-generate-20261018-01 make generate`: native 42928, outer 0, 10610 ms. Generated outputs refreshed through Make.
- Green `...=remediation-w10-r8b-preparation-green-20261018-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.postgres_template_preparation,harness.browser.integration.postgres_cleanup_target_scoped_coordination`: native 1224, outer 0, 25412 ms, 4/4 with all 17 cleanup. Existing checkpoint-LSN, clone isolation, existing-target protection, missing-template and target-admission assertions survive; failed preparation now removes its destination and preserves its permission cause.
- Exact product regressions, all outer 0: `...=remediation-w10-r8b-migrations-20261018-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.cryptographic_state_admission` (native 35222, 17253 ms); `...=remediation-w10-r8b-jobs-20261018-01 make test-slice OWNER=platform.jobs ROWS=platform.jobs.integration.transition_progress_contract` (24414, 11945 ms); `...=remediation-w10-r8b-evidence-20261018-01 make test-slice OWNER=module.evidence ROWS=module.evidence.integration.preview_and_download_handle_issuance_accepts_onl_84f55ae52b` (2881, 18438 ms). Migration and Jobs each 3/3 with all 14 cleanup; evidence summary/cleanup remains to be reconciled with the next evidence checkpoint.
- `...=remediation-w10-r8b-preparation-source-20261018-01 make source-check CHECK=all`: native 39598, outer 0, 16332 ms, 9/9 with all 18 cleanup. `...=remediation-w10-r8b-preparation-markdown-20261018-01 make lint-markdown`: native 40247, outer 0, 10820 ms. `git diff --check` passed. Current source `sha256:c12ad519b44f4698050dc27c9576b15cbf131648693f44582dd785a3c87105a6`.
- R8b implementation/focused validation DONE. The original exact filesystem wait remains unproven; concurrent full qualification remains an R8/W10 requirement. R9 characterization may now proceed after this checkpoint, before the common finalizer and renewed broad acceptance, so neither known repair is qualified against unfinished executable inputs. This sequencing changes no final acceptance criterion. R8/W10/overall IN_PROGRESS; W11 TODO.

### W10-R9 retained worker-budget characterization and repair entry (E-191)

- Fresh owner guidance: test_catalog and evidence_accounting outer 0 (native 79897/85585), command_surface outer 0 (76855); retained at `/tmp/cartulary-remediation-w10-r9-{catalog,evidence,surface}-guide.txt`.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r9-budget-red-20261018-01 make harness-contract SUITE=execution`: native 72034, outer Make 2 / normalized 10, 251315 ms, 6/7 units passed. Exact retained reconstruction characterization fails: telemetry owner graph executes with CPU/GOMAXPROCS 4, but retained reconstruction uses 9 from host capacity 19, changing unit and graph digests. Independent runtime-boundary work drained before edits. Failed root remains preserved.
- Add required resolved `go_worker_budget` to the coordinated, unreleased v3 manifest contract, separate from physical capacity. Producer retains it; reconstruction consumes it independently of ambient lease state; native admission and complete retained validation reject invalid/missing/excess budgets. No capacity override, compatibility reader or lease fabrication. Existing v3 development evidence without the field is historical and cannot qualify the new candidate.
- Authored schema, producer/reader, native semantic check, synthetic fixtures, exact graph and malformed-budget tests updated. Generated outputs must be refreshed through Make. First `...=remediation-w10-r9-budget-format-20261018-01 make format` (26874) exited 2/configuration_error before graph publication because the old native generated validator rejects the newly added manifest field. This is a schema regeneration prerequisite, not a product failure; use public `make generate` before renewed format and validation.
- E-190 evidence-row reconciliation: 3/3 units passed in 18438 ms. R8b focused implementation remains DONE; R8 full concurrent acceptance remains open. R9/W10/overall IN_PROGRESS; W11 TODO.

### W10-R9 focused validation diagnosis (E-192)

- Public generation succeeded: `...=remediation-w10-r9-budget-generate-20261018-01 make generate`, native 86758 outer 0, 10660 ms. Format after generation succeeded: `...=remediation-w10-r9-budget-format-20261018-02 make format`, 38433 outer 0, 8441 ms. No generated artifact was hand-edited.
- Frozen intermediate source `sha256:cfb3297830454274208baa4c3d1f3a10ccf74c7237effc524f196e2de821310f`: regression preset (21155) outer 2/normalized 11, 41/42 units in 367388 ms, all 51 cleanup complete; native runtime exact row (62173) outer 2/normalized 10, 0/1 in 279699 ms, all 10 cleanup complete; source-check all (56666) outer 0, 9/9 in 297172 ms, all 18 cleanup complete. These durations include shared admission waits, not qualification timings. Markdown (84135) outer 0, 11320 ms; diff-check outer 0.
- Both failed paths reach the same newly added negative fixture: it computed a semantic digest for an unsafe integer before invoking the retained reader. Native runtime's nested evidence selection therefore failed; native plan budget assertions themselves passed. Independent regression work drained before correction. Graph reconstruction fast/matrix passed, including owner/single-row, direct/nested budgets and ambient-reader isolation.
- Corrected unsafe fixture bypasses impossible semantic hashing and expects rejection at strict JSON decoding. First focused rerun `...=remediation-w10-r9-budget-evidence-20261018-02 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.current_epoch_evidence` (89678) outer 2/10, 10279 ms: intended rejection occurred, but the assertion expected “number” while the diagnostic says “integer is outside the I-JSON interoperable range”. Corrected that exact expectation; no runtime validation was weakened.
- Same exact evidence selection with ID `remediation-w10-r9-budget-evidence-20261018-03` (37045) passed outer 0, 10314 ms. Native exact row renewed with ID `remediation-w10-r9-budget-native-20261018-02` (71827); collect its result before finalization. R9/W10/overall IN_PROGRESS, W11 TODO.
- E-191 cleanup reconciliation: red graph source be16b898, all 16 cleanup records completed. E-190 evidence row all 14 cleanup records completed.

### W10-R9 implementation completion and finalizer entry (E-193)

- Native exact runtime row `remediation-w10-r9-budget-native-20261018-02` (71827) passed outer 0, 1/1 in 44003 ms; all 10 cleanup records completed. Source `sha256:88e1f8998bdb39b0679803e77447721c2a712d805ad1cdcd69a4ce89ff045103`. The nested evidence selection also passes. Together with E-192 graph reconstruction and corrected evidence validation, R9/F-031 implementation is DONE; broad final-source acceptance remains W10-owned.
- Resolved worker budget is retained separately from physical capability; exact reconstruction covers one/multiple groups and direct/nested budgets, ignores reader environment, and rejects malformed/missing/excess budgets. Native schema/semantic rejection and retained validation are characterized. No historical evidence is relabeled.
- Begin `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-finalize-20261018-02 env -u RESULTS_DIR make agent-finalize`. Retained-run maintenance is skipped because RESULTS_DIR is unset: no successful full warm check is eligible for this changed source. Review reported mutations before freezing and beginning full assurance. R8/W10/overall IN_PROGRESS; W11 TODO.

### W10 finalized candidate and full assurance entry (E-194)

- `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-finalize-20261018-02 env -u RESULTS_DIR make agent-finalize`: native 61554, outer 0, graph 25284 ms; all maintenance actions selected by the command passed. Reviewed `unit-artifacts/finalize-summary.json`: generated unchanged, updated_file_count 0, rollback not needed, no failures. Retained maintenance explicitly skipped because RESULTS_DIR was unset.
- Finalizer validation source remains `sha256:88e1f8998bdb39b0679803e77447721c2a712d805ad1cdcd69a4ce89ff045103`. Freeze this source for renewed complete assurance, then sequential CI/release. R8's WAL_LOG/checkpoint preparation and R9's exact retained budget now share one candidate. No harness performance qualification/publication is required.
- Start `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-full-20261018-02 make harness-contract SUITE=full`. Preserve all earlier failed evidence; new success cannot rewrite it. R8/W10/overall IN_PROGRESS pending complete applicable acceptance; W11 TODO.

### W10 complete-assurance failure checkpoint and structural follow-ups (E-195)

- `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-full-20261018-02 make harness-contract SUITE=full`: native 14598, observed outer Make 2 / normalized 10, 1531847 ms, 57/58 parent units passed, all 67 cleanup records completed. Frozen source 88e1f899, graph a1ce5b6f. No CI/release overlapped. Owning exit was collected after retained scanning; source edits may resume now.
- Human child `unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check` passed 800/800 in 1162154 ms, all 851 cleanup records completed (cache 179 bypass/232 hit/389 miss). Machine child `operation-harness.assurance_output_check_machine/.../real-check-machine` failed 795/800 in 1073070 ms, all 767 graph cleanup records completed (179/316/305). Fast child `operation-harness.assurance_output_test_fast/.../real-test-fast` passed 582/582 in 287641 ms, all 107 cleanup records completed (6/484/92). Parent wrapper assertions observe nested Make outcomes; no independent child native handle exists.
- Machine cleanup failures: database_migrations concurrent_apply_locking and cryptographic_state_admission; reference_data canonical_persistence and reference_pack_activation_prior_active_retention_af4f8ee94a; revisions portability.atomicity. Four failed Go units retain harness/cleanup_error 12, complete row outputs and no missing output. R8a correctly retains fragmented envelopes. Inspected revisions/migration/reference_data diagnostics show forced_drop/deadline_exceeded, IPC/CheckpointStart or CheckpointDone, checkpointer IO/DataFileSync, no lock blockers and no observed barrier participant. This is checkpoint I/O evidence, not another proved signal-barrier delay. Exact storage-device cause remains unknown.
- Separate product/test_assertion_failure 10: `module.extensions.integration.state_concurrent_admission` (24110 ms), `state_service_test.go:444`, waiting for first lock attempt. Its test context starts before expensive fixture acquisition; source inspection confirms setup consumes the admission observation budget. Preserve this distinct failure classification and investigate/repair its timing boundary in W10-R10 after the clone checkpoint slice.
- R8c/F-033 reevaluates the added per-clone CHECKPOINT. [Pinned PostgreSQL source](https://raw.githubusercontent.com/postgres/postgres/REL_18_6/src/backend/commands/dbcommands.c) logs every copied block for WAL_LOG and forces synchronous CREATE commit; FILE_COPY alone requires surrounding checkpoints. The harness's extra whole-cluster checkpoint is unnecessary for durable clone publication and couples each acquisition to unrelated dirty data. Replace that requirement with durable successful WAL-backed creation, preserve ownership/compensation and all durability settings, and characterize creation without checkpoint privilege plus WAL flush, isolation and failed provisioning. Do not increase cleanup deadlines or add retries/global serialization.
- Fresh guidance browser/revisions/extensions passed outer 0 (78632/4456/22766), retained `/tmp/cartulary-remediation-w10-r8c-{browser,revisions}-guide.txt` and `/tmp/cartulary-remediation-w10-r10-extensions-guide.txt`. Extension Table 9-E starts migration-lock acquisition timing immediately before the first lock attempt, separately from fixture setup. R8c starts after this checkpoint; R10 follows its checkpoint. W10/overall IN_PROGRESS, W11 TODO.

### W10-R8c durable clone publication checkpoint (E-196)

- Amended clone publication to require successful durable WAL-backed CREATE and required provisioning, removing the unnecessary whole-cluster checkpoint. Retained fsync/full-page writes/synchronous commit/checksums, target admission, creation outcome and compensating cleanup. Original signal-barrier and current storage-device causes remain distinct and are not declared proved.
- Physical characterization first failed at the intended boundary: a CREATEDB role without checkpoint privilege completed CREATE but the extra CHECKPOINT returned 42501. Native 63995 outer 2/10, no leaked owned fixture. Removing the explicit checkpoint passes the same role boundary, WAL-flush progress and enabled durability settings, template isolation, failed provisioning cleanup, existing-target protection, and target-scoped cleanup coordination. No operation limit, retry, capacity or pin changed.
- Exact Make selections: browser rows `postgres_template_preparation,postgres_cleanup_target_scoped_coordination` (32793 outer 0); revisions row `module.revisions.portability.atomicity` (42964 outer 0); migrations rows `concurrent_apply_locking,cryptographic_state_admission` (42913 outer 0); reference_data rows `canonical_persistence,reference_pack_activation_prior_active_retention_af4f8ee94a` (17668 outer 0). Full row IDs retain the owner/family prefixes shown in E-195 and the manifests.
- `remediation-w10-r8c-clone-red-20261018-01`: fail, 2/3, 16890 ms, all 14 cleanup records completed.
- `remediation-w10-r8c-clone-green-20261018-01`: pass, 4/4, 27533 ms, all 17 cleanup records completed.
- `remediation-w10-r8c-revisions-20261018-01`: pass, 3/3, 15369 ms, all 14 cleanup records completed.
- `remediation-w10-r8c-migrations-20261018-01`: pass, 3/3, 28968 ms, all 14 cleanup records completed.
- `remediation-w10-r8c-reference-20261018-01`: pass, 4/4, 49981 ms, all 17 cleanup records completed.
- `remediation-w10-r8c-clone-source-20261018-01`: pass, 9/9, 18385 ms, all 18 cleanup records completed.
- `...=remediation-w10-r8c-clone-format-20261018-01 make format` (48617) outer 0, 8293 ms. `...clone-source... make source-check CHECK=all` (59597) outer 0. `...clone-markdown... make lint-markdown` (33760) outer 0, 22840 ms; diff-check outer 0. Fresh migration/reference guidance 83250/43788 outer 0. No generated projection changed in this slice; next finalizer regenerates/checks the coordinated candidate.
- R8c implementation DONE; its concurrent qualification remains R8/W10-owned. R10 may now begin after this checkpoint. W10/overall IN_PROGRESS; W11 TODO.

### W10-R10 admission-test boundary completion (E-197)

- Existing E-195 failure and source characterization establish the setup/admission boundary defect. Moved the unchanged 15-second concurrent-admission test window after fixture/runtime preparation, keeping LockTimeout at five seconds. The blocking test validator now obeys cancellation; owned test workers are joined during cleanup before the database fixture is released. Product admission code and assertions of two acquisitions/releases with maximum concurrency one are unchanged.
- `CARTULARY_TEST_RUN_ID=remediation-w10-r10-timing-format-20261018-01 make format`: native 38476 outer 0, 8104 ms. `CARTULARY_TEST_RUN_ID=remediation-w10-r10-timing-green-20261018-01 make test-slice OWNER=module.extensions ROWS=module.extensions.integration.state_concurrent_admission`: native 71623 outer 0, 3/3 units, 11592 ms, all 14 cleanup records completed. Source `sha256:00e3a62e6cb960f32bbd15e2b4c49ac62ac48f1f992823f628f5d4d3c3c045f4`. Diff-check outer 0.
- R10/F-034 implementation DONE and checkpointed. R8c's redundant checkpoint removal supersedes E-190's explicit-checkpoint requirement; durable WAL publication and exact cleanup remain required. R8/W10 acceptance stays open until the new candidate passes complete qualification. W11 TODO; overall IN_PROGRESS.
- Begin `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-finalize-20261018-03 env -u RESULTS_DIR make agent-finalize`. No successful full warm check matches this changed source; retained maintenance remains explicitly skipped because RESULTS_DIR is unset. Review every mutation before the next freeze.

### W10 revised candidate freeze and full assurance entry (E-198)

- `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-finalize-20261018-03 env -u RESULTS_DIR make agent-finalize`: native 59118 outer 0, 27020 ms. Reviewed finalize-summary: generated unchanged, zero updated files, rollback not needed, no failures; retained maintenance skipped because RESULTS_DIR was unset. Source is `sha256:00e3a62e6cb960f32bbd15e2b4c49ac62ac48f1f992823f628f5d4d3c3c045f4`.
- Final-source `...=remediation-w10-candidate-source-20261018-03 make source-check CHECK=all` (84684) outer 0, 17119 ms. `...=remediation-w10-candidate-markdown-20261018-03 make lint-markdown` (47931) outer 0, 11670 ms.
- Renew complete assurance with `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-full-20261018-03 make harness-contract SUITE=full`. This is justified by the characterized removal of redundant checkpoint work and the distinct admission-test timing correction; it is not an unexplained retry. Freeze executable source through owning exit and all finalization. CI and release remain sequential successors, followed by exact audits and W11. W10/overall IN_PROGRESS; W11 TODO.

### W10 full-assurance terminal result and user-requested stop (E-199)

- User explicitly requested stopping after the already-running full-assurance command exits and finishes cleanup. No cancellation was sent. The owning native session 36456 exited with Make 2, nested normalized code 10, after retained processing. Command: `CARTULARY_TEST_RUN_ID=remediation-w10-candidate-full-20261018-03 make harness-contract SUITE=full`. Exact root: `.cartulary/test-results/remediation-w10-candidate-full-20261018-03`. Terminal checkpoint observed by 2026-10-18T14:42:38Z.
- Frozen executable source: `sha256:00e3a62e6cb960f32bbd15e2b4c49ac62ac48f1f992823f628f5d4d3c3c045f4`; parent graph `sha256:a1ce5b6f71d3b6b3f7c993d15c276f7d3b2f30df36766191c2a421e0552da33b`. Branch `codex/specification-remediation-20261016`, HEAD `afd6c699980d4256af980a9abb906b2d79004755`. WSL2 pinned environment remains unchanged. No executable source was edited after the stop request.
- Parent `run-summary.json`: fail, product/test_assertion_failure, 57/58 units passed, 1312506 ms graph wall duration, cache 58 bypass/0 hit/0 miss. The failed assertion is `operation:harness.assurance_output_check_machine`, observing its nested Make failure. All 67 parent graph cleanup records completed. This wrapper classification does not replace the child's harness/cleanup_error classification.
- Human child `unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check`: pass, 800/800, 944868 ms, cache 179 bypass/225 hit/396 miss, all 858 graph cleanup records completed. Fast child `operation-harness.assurance_output_test_fast/real-command-output/runs/real-test-fast`: pass, 582/582, 406129 ms, cache 6/368/208, all 223 graph cleanup records completed. Wrapper assertions observe child Make outcomes; no independent child native session exists. These are complete child selections, not separate outer invocations or performance qualification.
- Machine child `unit-artifacts/operation-harness.assurance_output_check_machine/real-command-output/runs/real-check-machine`: fail, harness/cleanup_error, 793/800, 975298 ms, cache 179 bypass/348 hit/273 miss, all 735 graph cleanup records completed. Seven Go units retain normalized code 12 and eleven failed rows; no missing unit evidence outputs. Graph cleanup completion is distinct from these failed test-level PostgreSQL teardown operations and does not erase them.
- All seven failed-unit results and their referenced failed-row JSON were inspected. Typed diagnostics report forced_drop/deadline_exceeded, IPC/CheckpointStart or CheckpointDone, zero observed blockers, and no observed barrier participant. Most first/last checkpointer samples show IO/DataFileSync; imports/upload retains one other/empty sample. Several rows share requested/done snapshots 64/63 or 100/99, with later 65/64 or 101/100. These are bounded observations, not proof of continuous wait or a storage-device root cause. No raw log inspection, reproduction, source repair, retry, reset, timeout change or capacity change followed the stop request.

The following exact unit/row inventory is relative to the machine child above. `unit-results` preserves owning normalization; `rows/<row-id>.json` preserves the typed cleanup diagnostics.

| Failed unit | Failed rows |
| --- | --- |
| `go:0eab78ed6e254ecb:b12e48365b78` | `module.recovery.integration.selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a` |
| `go:102d5fe4ec38831e:8128a5ab32cf` | `module.evidence.integration.a_successful_evidence_attach_publishes_canonical_d63a42e346`; `module.evidence.integration.preview_and_download_handle_issuance_accepts_onl_84f55ae52b` |
| `go:2f89aed0bdc82556:de8b3c1351b1` | `module.database_migrations.integration.cryptographic_state_admission`; `module.database_migrations.integration.production_preflight_state_matrix` |
| `go:85702fd10df1d588:81b8c5f8e689` | `module.auth.integration.admin_password_reset_totp_reset_and_revoke_all_r_eaa09b4d6c`; `module.auth.integration.saml_sp_initiated_begin_acs_verifies_determinist_3b5547023b` |
| `go:968ed2e13745016a:d5b3a26a174c` | `module.extensions.integration.state_admission_matrix_7f6ad80724` |
| `go:abc0959db227a951:10f1bc475143` | `module.imports.integration.apply_transaction_current_state_bc8ad0ac32`; `module.imports.integration.upload_admission_replay_complete_b8684935f2` |
| `go:e8eccc6efddf32cc:b3f543dc9303` | `platform.jobs.integration.expiry_and_compaction` |

- R8c's unnecessary explicit clone checkpoint is removed and its focused durability/privilege/isolation evidence remains valid, but the complete run disproves closure of the broader PostgreSQL cleanup problem under this load. Current checkpoint waits must not be mislabeled as the earlier observed create-database ProcSignalBarrier wait. R8 remains IN_PROGRESS. R10's exact `module.extensions.integration.state_concurrent_admission` row passed in both human and machine checks (3290/1480 ms); the machine extension unit failed a different row, `state_admission_matrix_7f6ad80724`, during cleanup.
- E-198 finalizer passed with zero mutations, and source-check all/Markdown lint passed on this source. Retained-run maintenance was skipped because RESULTS_DIR was unset: no eligible same-source full warm check existed when finalization ran. Do not rerun finalization solely to attach the later human check. No current-source CI or complete release was started; exact owner/evidence and full observability audits remain unrun. Harness performance qualification/publication remain removed by the user.
- Working tree remains uncommitted: 546 modified tracked paths, 66 deleted tracked paths, 62 untracked paths at the checkpoint; no staged entries. The ignored temporary tracker and Git-visible durable handoff are preserved. This is a status inventory, not W11's final diff/index/content reconciliation against E-000. Preserve baseline evidence and private recovery proofs; no cleanup/reset command was run.
- Restart only after the user asks to resume. First inspect this exact retained machine failure and the common checkpoint/fixture lifecycle boundary; do not blindly retry the full suite. Any new repair must receive its own characterized implementation checkpoint and appropriate finalization. Then obtain successful applicable full assurance, sequential CI and complete release, exact evidence/observability audits, and the W10 completion checkpoint before W11's reconciliation and final handoff. Reuse eligible evidence only when source and selection contracts permit it.
- Stopping status: W00–W09 DONE; R8 and W10 IN_PROGRESS; W11 TODO; overall IN_PROGRESS, execution stopped at user request. There is no task-owned invocation still running and no follow-up scheduled. This is a durable stopping-point handoff, not completed release acceptance.

### Stopping-point handoff validation and final checkpoint (E-200)

- `CARTULARY_TEST_RUN_ID=remediation-w10-stop-markdown-20261018-01 make lint-markdown`: native 41100, observed outer 0, 18730 ms; summary `.cartulary/test-results/remediation-w10-stop-markdown-20261018-01/adhoc/lint-markdown/tool-run-summary.json`. This documentation check ran after the full-assurance exit; no product verification, generation or finalization was started.
- `git diff --check` returned 0. Human snapshot consistency check confirmed that the durable document contains the complete controlling tracker verbatim, with E-199 present once. Reviewed the durable introduction, acceptance table, current issue, migration/rollback and restart instructions. This receipt is the only addition after Markdown validation; no executable inputs changed.
- Final execution step for this stopping checkpoint: retain the full tracker snapshot under `docs/handoffs/testing-harness-remediation.md` and preserve `temp/plan-notes.md`. W10/R8 remain IN_PROGRESS, W11 remains TODO, overall remains IN_PROGRESS; execution is stopped at the user's request. Full assurance's outer exit and graph finalization were collected, both native sessions are finished, and no further invocation or follow-up is scheduled. W11's final acceptance reconciliation is deliberately not claimed complete.

## 11. Open Questions and Blockers

| ID | Condition | Affected slice | Resolution | Status |
| --- | --- | --- | --- | --- |
| RB-001 | No usable harness performance reference portfolio | Resolved for this effort | User explicitly removed harness performance qualification from plan/specifications | DROPPED |
| RB-002 | Full output-case assurance after boundary repairs | W09 | Closed by E-130: full assurance outer 0, 58/58, human/machine check and fast assertions passed with complete cleanup. Historical causal limits and private proof remain preserved. | DONE |
| RB-003 | Historical release fixture cleanup and dependent measurement-summary failures lack final-source closure | W10 | XLSX discovery row itself passed; lease-000049 release failed, followed by summary artifact rejection/provider cleanup failure. Harness browser/release owners require complete final-candidate release evidence. | TODO |
| RB-004 | Final-source full assurance fails in seven machine-check Go units with PostgreSQL checkpoint-wait cleanup deadlines | W10/R8 | E-199 retains all failed units/rows; graph finalizers completed, physical I/O cause remains unresolved. Diagnose before another repair or rerun. User requested stopping after this terminal checkpoint. | IN_PROGRESS |

## 12. Binary Completion Criteria

- Planning completeness: complete for the identified remediation scope; implementation discoveries are recorded before changing a decision.
- Implementation authorization: granted by the user; execution in progress.
- Per-slice readiness: W00–W09 complete; W10 repairs R1–R7/R6a, R8a/R8b/R8c, R9 and R10 are checkpointed. E-199 full assurance failed; R8/W10 remain IN_PROGRESS and work is stopped at user request. Successful final-source full assurance, CI, release and evidence audits remain open; W11 requires W10 completion and checkpoint.
- Completion: no required finding, slice or release acceptance remains unresolved; final-source validation and cleanup evidence closes; all 190 original command dispositions reconciled; exact 62-name surface exposed; obsolete aliases/dispatch removed.
- Handoff: migration guide, selectors/defaults, context/assurance mappings, cache/evidence compatibility, exact run roots and outcomes, finalizer edits and coordinated rollback recorded.
- Preservation: compare index and file contents against E-000; record expected authored/generated edits and unexpected/concurrent differences without reverting unseen work.
- Last execution step: update W11 and overall completion only after every required gate passes. Refresh the durable handoff and tracker snapshot at `docs/handoffs/testing-harness-remediation.md`; keep this working tracker temporary and ignored.

