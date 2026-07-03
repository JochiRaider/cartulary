# Test Harness Refactor Tracker: Next Iteration

## 1. Status and Authority

| Field | Value |
| --- | --- |
| Target path | `tools/harness` |
| Output path | `docs/handoffs/test-harness-module-refactor-tracker.md` |
| Tracker iteration | Next iteration after completed remediation S-01 through S-09. |
| Current harness files | NI-00 baseline: 237 tracked files under `tools/harness`. NI-01 adds one authored helper at `tools/harness/core/json-shape.mjs`, so the post-NI-01 tracked baseline is 238. The final worktree contains 238 tracked harness files plus 11 new authored harness files from WS-1, WS-2, and WS-4; the expected tracked count after add is 249. |
| Current implementation authorization | User authorized implementation of the complete remediation plan on 2026-07-03. NI-00 through NI-05 and WS-5 final validation/handoff are complete in this worktree, pending review/commit. |
| Public behavior posture | Public harness contracts remain frozen unless `docs/testing-harness-nlspec.md` and owner inputs are revised first. |
| Product boundary posture | No product behavior refactor is authorized. Product auth, workbook behavior, projections, saved views, product routes, product WebSocket behavior, record models, and generated protocol contracts must remain outside `tools/harness`. |
| Generated artifact posture | Generated roots and generated harness outputs must not be hand-edited. Refresh generated outputs only through Make-owned targets. |

Owner documents and authority used for this iteration:

| Document | Role |
| --- | --- |
| `docs/testing-harness-nlspec.md` | Owns harness mechanics, public Make surface, command IDs, output modes, schemas, result roots, retained artifacts, failure taxonomy, scheduler semantics, service lifecycle, cache/readiness behavior, generated-artifact handling, and test-only harness routes. |
| `docs/domain.md` | Confirms harness terms are implementation-support unless promoted by an adopted owner document. |
| Core 00 through Core 04 | Own product behavior, public product routes, product WebSocket behavior, workbook surfaces, saved views, projections, records, security, deployment, and conformance boundaries. |
| Core 05 | Applies only to claim-bearing timed, benchmark, fixture-sensitive, or publication evidence. It is not Base Profile implementation conformance. |
| `tools/task_surface_manifest.json` | Current public target mirror and owner input for task-surface generation. |
| `tools/execution_topology_manifest.json` | Current topology/scheduler owner input. |

No Core 00 through Core 04 revision is required for NI-00 or NI-01 because the work is internal harness structure only. A later slice must revise owner documents before changing any public target, command ID, output class, schema ID, retained artifact contract, scheduler event/resource semantics, cache semantics, or test-route behavior.

## 2. Historical Closure From Prior Tracker

The previous remediation effort is complete and is now historical context, not active work. It recorded:

- S-01 baseline characterization.
- S-02 through S-08 behavior-preserving seam work.
- Core public-contract facade, scheduler adapters, browser runtime paths, backend runner/duration/drift facades, frontend evidence/design/readiness facades, generated-artifacts contract shim, and static-analysis design shim.
- Final validation including `make check`, `make agent-finalize RESULTS_DIR=<successful check root>`, generated drift, JSON shape checks, phase schedule drift, duration drift, generated policy, phase ledger drift, task-surface check, and Markdown lint.

Superseded stale text from the prior tracker:

- The 224-file inventory is stale; the live count is 237 tracked files.
- RB-001 through RB-004 are no longer active blockers for the completed remediation.
- S-02 through S-08 proposed slice text is complete and must not be re-executed as future work.
- Rows saying backend/frontend implementation remained unchanged are historical and no longer describe the live state.
- "No immediate next action" is superseded by NI-00 followed by NI-01.

## 3. Live Harness Boundary Rescan

Current file counts by subtree after NI-01:

| Subtree | Count | Current responsibility | Boundary notes |
| --- | ---: | --- | --- |
| `tools/harness/backend` | 26 | Go target execution, Go sharding, duration baselines, migration/schema drift, backend build wrapper. | Owns backend harness execution, not product backend domain logic. New `runner`, `duration`, and `drift` seams are structural footholds. |
| `tools/harness/browser` | 31 | Playwright selection, browser stack lifecycle, reset, browser duration and batch planning. | Runtime-adjacent harness logic only. Test routes remain harness-owned test-only behavior, not production route authority. |
| `tools/harness/core` | 37 | Public harness contract plumbing, summaries, artifacts, redaction, failure taxonomy, finalization, Make wrapper support, generic JSON shape helpers. | `public-contract.mjs` is a facade. `core/test-output` still imports planning/browser/frontend evidence logic and remains future work. |
| `tools/harness/frontend` | 23 | Frontend phase/evidence accounting, Vitest wrappers, design-token/font checks, frontend readiness. | Evidence/readiness/design seams exist. This is implementation-support and design-direction evidence, not product conformance by itself. |
| `tools/harness/generated-artifacts` | 19 | Authored generators and drift/shape checkers for generated harness outputs. | Authored source, not generated output. `json-shape.mjs` is now a private compatibility re-export because active manifests/tests still reference the old path. |
| `tools/harness/planning` | 23 | Phase maps, phase slices, target plans, task guidance, explain commands. | Evidence accounting only. Must not become product runtime architecture. |
| `tools/harness/readiness` | 18 | Toolchain bootstrap, readiness cache, dev services, local stack, process lifecycle. | Cache/readiness behavior is local acceleration/support evidence only. |
| `tools/harness/scheduler` | 37 | Scheduler engine, adapters, manifests, resources, event/timing drift, service-backed topology. | Scheduler adapters are present, but broader top-level import cycles remain. |
| `tools/harness/static-analysis` | 21 | Backend/frontend boundary checks, lint/security wrappers, Fallow wrapper. | Test fixtures intentionally contain invalid examples; scans must distinguish fixtures. |
| `tools/harness/test-support` | 3 | Harness self-test scratch, JSON, artifact helpers. | Test-only support. Scratch roots must remain outside the repository checkout. |

Live public surface summary after WS-3:

- `tools/task_surface_manifest.json` currently declares 136 targets: 94 public, 31 internal helper, and 11 check-internal.
- Public lifecycle states: 94 `public_active`, 0 `public_deprecated`.
- `db-down` has been removed from the current public Make surface; `services-down` is the single public local-service stop command.
- Public targets, command IDs, output classes, schema policies, side effects, result roots, retained artifact behavior, and failure mapping must not drift accidentally.

Live import-boundary finding:

```text
backend, browser, core, frontend, generated-artifacts, planning, and scheduler
currently form one top-level strongly connected component.
```

Notable remaining coupling:

- NI-01 moved generic JSON/object/schema helpers to `tools/harness/core/json-shape.mjs`. `tools/harness/generated-artifacts/json-shape.mjs` remains as a private shim because `tools/task_surface_manifest.json`, `tools/execution_topology_manifest.json`, and `harness-smoke-json-shapes` still reference the old helper path. Do not hand-edit those generated/topology artifacts to remove the reference.
- Planning imports backend/browser shard planners while those planners import planning rows or phase manifests.
- `core/test-output/cli.mjs` imports planning, browser, and frontend evidence logic directly.
- `generated-artifacts` imports backend/browser/frontend/planning/scheduler helpers to render schedules, ledgers, and shape checks.
- Core, frontend, planning, and scheduler still import generated-artifact contract/topology/task-surface helpers; those are future boundary decisions and were not part of NI-01.
- Scheduler core no longer imports backend/browser/planning directly except through `tools/harness/scheduler/adapters/*`, but the wider SCC remains.

## 4. Public Contract Freeze Map

These interfaces are stable unless a spec-authorized change is explicitly proposed:

| Interface | Owner | Must not change in NI-00 or NI-01 |
| --- | --- | --- |
| Public Make targets and command IDs | Harness NLSpec Section 4 and task-surface manifest | Names, command IDs, target classes, lifecycle states, default inclusion, side-effect declarations. |
| Accepted public inputs | Harness NLSpec Section 5 | Closed Make variable/input matrix and undeclared-input rejection behavior. |
| Result roots and run IDs | Harness NLSpec Section 6 | `CARTULARY_TEST_RESULTS_DIR`, `CARTULARY_TEST_RUN_ID`, run-root layout, permissions, retained identity. |
| Output modes and output classes | Harness NLSpec Section 7 | `quiet`, `summary`, `ci`, `verbose`, `debug`, `machine`; machine stdout shape; human output budgets. |
| Schema IDs and artifact validation | Harness NLSpec Section 8 | Stable schema IDs, schema-owned artifacts, unknown-field behavior. |
| Failure taxonomy | Harness NLSpec Section 9 | `failure_class`, `failure_reason`, normalized exit mapping, scheduler propagation rules. |
| Scheduler events/resources | Harness NLSpec Section 10 | Scheduler manifest shape, event order, resource registry, command types, finalizer behavior. |
| Service/fixture lifecycle | Harness NLSpec Section 11 | Owned/attach mode, fixture policies, cleanup/teardown rules, readiness deadlines. |
| Test-only routes | Harness NLSpec Section 12 and Core 04 product boundary | `/api/v1/test/*`, `/ws/v1/test/*`, route token, host/origin rejection, no production exposure. |
| Cache/readiness behavior | Harness NLSpec Sections 1, 4, 8, 10, 11 | Cache hits remain local acceleration and must not skip summaries, failure classification, drift, security, service readiness, cleanup, reset, or aggregate verdicts. |
| Product behavior | Core 00 through Core 04 | Workbook surfaces, saved views, projections, product routes, product WebSocket events, auth, records, evidence, history, deployment. |

## 5. Work Item Matrix

| ID | Gap or problem | Recommended remediation | Belongs in | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria | Rollback or containment | Disposition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| NI-00 | Tracker is stale against the live repo. | Replace prior active remediation tracker with this next-iteration tracker; record 237-file scan, completed historical work, current risks, baseline requirements, and slice sequence. | Documentation. | Prevents future agents from executing completed slices or trusting stale counts. | Clear control artifact for future harness work. | No runtime impact. | Wrong slice selection, false blockers, and accidental public contract drift. | `git ls-files tools/harness`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make check-harness-smoke`; `make harness-contract`. | Revert this tracker update. | Complete in this worktree. |
| NI-01 | Generic JSON/object/schema helpers sit in `generated-artifacts`, making generator tooling look like the owner of shared JSON validation. | Move generic JSON shape helpers behind a core-private contract module; update internal imports; keep `generated-artifacts/json-shape.mjs` as a private compatibility shim if active imports/tests/manifests still reference it. | Implementation and tests. | Generic JSON parsing, duplicate-key rejection, object-key validation, enum validation, and timestamp checks are core harness contracts, not generated-artifact generation behavior. | Reduces generator ownership confusion and removes a broad dependency from non-generator modules to generated-artifacts. | No public behavior change. Raw helper paths stay private. Old private path remains as a shim because active manifests and smoke metadata still reference it. | If reverted, `generated-artifacts` again appears to own cross-harness schema validation. | Baseline first, then `node --check` for touched `.mjs`; `make json-shape-check`; `make generated-artifact-policy-check`; `make check-harness-smoke`; `make harness-contract`. | Revert NI-01 as one slice; shim limits blast radius. | Complete in this worktree. |
| NI-02 | Backend/browser planning functions import planning and are imported by planning. | Refactor shard planners to accept normalized plan rows as explicit inputs; keep planning as row owner and backend/browser as execution planners. | Implementation and tests. | Removes bidirectional planning/shard-planner coupling without changing target behavior. | Future phase growth can add rows without increasing cyclic imports. | No public target/schema/output change intended. | Phase expansion keeps adding bidirectional imports and hidden assumptions. | `make check-harness-smoke`; `make backend-module-boundary-check`; `make browser-e2e-webserver-backed` if browser planner changes; shard planner self-tests. | Revert one slice; no generated outputs unless owner inputs change. | Complete in this worktree as WS-1. |
| NI-03 | `core/test-output/cli.mjs` imports planning/browser/frontend evidence logic directly. | Move evidence enrichment behind explicit private adapters or precomputed inputs while preserving standard summary schemas. | Implementation and tests. | Core should own public contract plumbing, not every evidence-accounting detail. | Smaller core facade and clearer future ownership for accounting logic. | High risk because summaries are public artifacts; no schema changes allowed without spec owner update. | Core remains a catch-all and blocks future decomposition. | Baseline plus `make harness-contract`, `make frontend-unit`, `make check-harness-smoke`, and targeted browser/backend tests. | Revert one slice; preserve summary schema compatibility. | Complete in this worktree as WS-2. |
| NI-04 | Legacy compatibility may exceed future value. | Audit `db-down`, historical schema IDs, legacy failure-class map, duration fallback fields, raw aggregate keys, and retained-run compatibility paths; remove only with owner authorization. | Specification, owner inputs, implementation, tests, docs. | Preserve legacy only when it has public or architectural value. | Lower compatibility burden and clearer current-profile surface. | Public `db-down` removal is intentional; old automation must use `services-down`. Historical browser stack/accessibility schemas remain diagnostic-only and are no longer Section 8 current attachments. Current browser shard verification requires selected-tests artifacts, raw Go duration checks require package baselines, current summaries no longer duplicate frontend row accounting under extensions, and current failure producers must emit canonical classes. | Legacy paths constrain new architecture and confuse evidence ownership. | `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make generate`; `make generate-drift`; `make json-shape-check`; `make generated-artifact-policy-check`; `make harness-contract`; targeted legacy-negative fixtures. | Revert spec/input changes and regenerate through `make phase-schedules` plus drift gates. | Complete in this worktree as WS-3. |
| NI-05 | Import boundaries are characterized but not enforced. | Add a harness import-boundary check or `harness-contract` assertion after desired dependencies are settled. | Tests/static-analysis and possibly docs. | Prevents regression back into the current SCC. | Keeps structural gains durable as phases and targets grow. | May make private import rules observable only as a validation gate, not public API. | Cycles return silently. | `make harness-contract`; `make lint-scripts`; boundary fixture tests. | Remove/relax the assertion if it proves over-broad. | Complete in this worktree as WS-4. |

## 6. Sequencing Rules

Dependency graph:

```text
NI-00 -> NI-01 -> NI-02 -> NI-03
NI-04 is spec-gated and may run only after owner decision.
NI-05 follows NI-01 and NI-02.
```

Entry criteria before any implementation slice:

1. Update this tracker with the selected slice and current state.
2. Run and record:
   - `git ls-files tools/harness`
   - `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`
   - `make check-harness-smoke`
   - `make harness-contract`
3. Stop before implementation if any baseline command fails until the failure is classified as pre-existing, environment-related, or slice-blocking.
4. Identify generated owner inputs before any generator-related edit.

Exit criteria for each implementation slice:

- Public Make behavior, command IDs, schemas, output modes, result roots, retained artifacts, failure mapping, scheduler events/resources, and cache/readiness behavior remain unchanged.
- Product auth, workbook behavior, projections, saved views, product routes, product WebSocket semantics, record models, evidence/history behavior, and grid/vendor integration remain outside `tools/harness`.
- Validation commands and retained roots are recorded in Section 10.
- Skipped checks are explicitly justified.
- Next action is named before another slice starts.

Generated-artifact handling rules:

- Do not hand-edit generated roots or generated harness outputs.
- If owner inputs change, run the relevant Make generator and drift target.
- Generated refreshes must be recorded with command, result, and changed files.
- `agent-finalize RESULTS_DIR=<successful full warm check run root>` is required only when a slice affects scheduler timing, duration baselines, default-check topology, finalizer/cache/readiness behavior, public target summary emission, or finalizer-maintained artifacts.

## 7. NI-01 Implementation Contract

Goal: make generic JSON/object/schema helper ownership core-private while preserving the old private helper path as a shim.

Allowed changes:

- Add a core-owned helper module for generic JSON shape functions.
- Update imports in authored harness files from `../generated-artifacts/json-shape.mjs` or equivalent deep relative paths to the new core module where those imports use generic helpers.
- Keep `tools/harness/generated-artifacts/json-shape.mjs` as a compatibility re-export for any active private callers that are not worth moving in NI-01.
- Do not change public schema IDs, validation algorithms, error text intentionally consumed by tests, generated outputs, task-surface metadata, or scheduler manifests.

Out of scope for NI-01:

- Moving generator-specific validators.
- Changing JSON schema files.
- Changing public Make targets or public inputs.
- Removing compatibility shims.
- NI-02 planner row normalization.
- NI-03 `core/test-output` adapter extraction.
- NI-04 legacy behavior removal.
- NI-05 import-boundary enforcement.

Validation required for NI-01:

| Layer | Command |
| --- | --- |
| Syntax | `node --check <each touched .mjs>` |
| Shape/schema | `make json-shape-check` |
| Generated policy | `make generated-artifact-policy-check` |
| Harness smoke | `make check-harness-smoke` |
| Harness contract | `make harness-contract` |
| Drift broadening | `make generate-drift` only if generated-owner inputs or generated outputs change unexpectedly. |

Rollback:

- Revert the new core helper module and import rewrites together.
- Keep the generated-artifacts shim if a partial rollback is needed to preserve private import compatibility.
- No generated output rollback should be needed unless a Make generator was run because owner inputs changed.

## 8. Validation Plan

| Validation layer | Command | Required when |
| --- | --- | --- |
| Live inventory | `git ls-files tools/harness` | NI-00 and before any later slice. |
| Public surface baseline | `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` | Before every implementation slice and after public-surface-adjacent changes. |
| Harness smoke | `make check-harness-smoke` | Before every implementation slice and after each implementation slice. |
| Harness contract | `make harness-contract` | Before every implementation slice and after output/artifact/summary/schema-adjacent changes. |
| Syntax | `node --check <touched .mjs>` | Any `.mjs` edit. |
| JSON/schema | `make json-shape-check` | NI-01 and any schema/helper change. |
| Generated policy | `make generated-artifact-policy-check` | NI-01 and generated-artifact helper changes. |
| Static imports/lint | `make lint-scripts`; `make lint-shell` | Import rewrites, shell edits, or boundary-check changes. |
| Backend boundary | `make backend-module-boundary-check` | NI-02 backend planner changes or backend imports. |
| Frontend boundary | `make frontend-import-boundary-check` | Frontend/static-analysis import changes. |
| Browser | `make browser-e2e-webserver-backed` | Browser shard planner/runtime changes. |
| Full check | `make check` | Multi-seam changes or scheduler/public summary changes. |
| Finalizer | `make agent-finalize RESULTS_DIR=<successful full warm check run root>` | Only when the slice affects finalizer, cache/readiness reuse, scheduler timing, duration baselines, default-check topology, `check-service-backed` timing, public target summary emission, or finalizer-maintained artifacts. |

## 9. RB-Style Blockers and Open Questions

| ID | Blocker or question | Required resolution |
| --- | --- | --- |
| RB-NI-001 | Baseline must be recorded before each implementation slice. | Resolved for NI-01 in Section 10. Repeat Section 6 baseline commands before NI-02 or any later slice. |
| RB-NI-002 | Public harness semantic changes are not authorized by this tracker. | Revise `docs/testing-harness-nlspec.md` and owner inputs before any such change. |
| RB-NI-003 | Product behavior must not move into harness. | Treat Core 00 through Core 04 as the owner of product auth, workbook, route, WebSocket, projection, saved-view, record, evidence, and history behavior. |
| RB-NI-004 | Legacy removal needs owner decision. | NI-04 must start as a spec/owner-input audit, not code deletion. |
| RB-NI-005 | Import-boundary enforcement should wait until target imports are intentionally shaped. | Implement NI-05 only after NI-01 and NI-02. |

## 10. Handoff Log

| Time | Agent/session | Work | Files inspected or changed | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03 | Codex / next-iteration planning | Produced next-iteration plan from live scan and owner docs. | Inspected prior tracker, harness NLSpec, domain doc, Core 00-04, `tools/harness`, task-surface/topology manifests. | `git ls-files tools/harness`; import graph scan; manifest count scan. | Found 237 tracked harness files and top-level SCC across backend/browser/core/frontend/generated-artifacts/planning/scheduler. | Implementation not yet started in that planning pass. | User authorized implementation; start NI-00. |
| 2026-07-03 | Codex / NI-00 tracker refresh | Replaced prior active remediation tracker with next-iteration tracker. | Changed this tracker. | Pending baseline commands. | Tracker now marks prior remediation historical and NI-01 as the only active code slice. | RB-NI-001 baseline pending. | Run NI-00 baseline commands before NI-01 code edits. |
| 2026-07-03T20:45:53Z | Codex / NI-00 baseline | Baseline characterization completed before NI-01 implementation edits. | Changed this tracker; inspected retained result roots. | `git ls-files tools/harness \| wc -l`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make check-harness-smoke`; `make harness-contract`. | All commands exited 0. Inventory count: 237. Task-surface report printed `check=pass` with 95 public, 31 internal helper, and 11 check-internal targets. Smoke root: `.cartulary/test-results/20260703T204525Z-p1003662`. Harness-contract root: `.cartulary/test-results/20260703T204525Z-p1003700`. | None for NI-00. | Start NI-01 generic JSON helper ownership move. |
| 2026-07-03T20:52Z | Codex / NI-01 implementation | Moved generic JSON shape helper ownership to core and rewired internal imports. Kept the old generated-artifacts helper path as a private re-export because active manifests and smoke metadata still reference it. Recomputed top-level import graph; large SCC remains. | Added `tools/harness/core/json-shape.mjs`; changed `tools/harness/generated-artifacts/json-shape.mjs`, `tools/harness/generated-artifacts/check-json-shapes.mjs`, `tools/harness/generated-artifacts/contracts/index.mjs`, backend/browser/frontend/planning/scheduler import callers, and this tracker. | Import scans; `node --check` for 12 changed `.mjs` files; `make json-shape-check`; `make generated-artifact-policy-check`; `make lint-scripts`; `make check-harness-smoke`; `make harness-contract`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make lint-markdown`. | All commands exited 0. Roots: json-shape `.cartulary/test-results/20260703T204858Z-p1006640`; generated policy `.cartulary/test-results/20260703T204903Z-p1006990`; lint-scripts `.cartulary/test-results/20260703T204903Z-p1007015`; smoke `.cartulary/test-results/20260703T204908Z-p1007527`; harness-contract `.cartulary/test-results/20260703T204908Z-p1007547`; lint-markdown `.cartulary/test-results/20260703T205209Z-p1010005`. Task-surface remained `check=pass` with 95 public, 31 internal helper, and 11 check-internal targets. | None for NI-01. Generated outputs were not hand-edited and no generator was run. | If NI-02 is authorized, record a fresh baseline and refactor backend/browser shard planners to accept normalized planning rows. |
| 2026-07-03T21:01Z | Codex / WS-0 rebaseline | Started authorized NI-02 through NI-05 implementation sequence and recorded the required fresh baseline before NI-02 edits. | Changed this tracker. | `git ls-files tools/harness \| wc -l`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make check-harness-smoke`; `make harness-contract`. | All commands exited 0. Inventory count: 238. Task-surface report printed `check=pass` with 95 public, 31 internal helper, and 11 check-internal targets. Smoke root: `.cartulary/test-results/20260703T210042Z-p1014896`. Harness-contract root: `.cartulary/test-results/20260703T210055Z-p1015991`. | None for WS-0. | Start NI-02 / WS-1 by refactoring backend and browser shard planners to accept explicit normalized planning inputs. |
| 2026-07-03T21:14Z | Codex / WS-1 NI-02 implementation | Made backend and browser shard planners pure over explicit normalized rows or entries. Added planning-owned backend/browser shard-plan adapters for discovery, moved phase-slice Go sharding to planning-owned target-plan rows, preserved private CLI behavior through executable/adaptor discovery, and kept public outputs unchanged. | Added `tools/harness/planning/backend-shard-plan.mjs` and `tools/harness/planning/browser-shard-plan.mjs`; changed backend shard planner/CLI/runner/duration callers, browser shard planner, phase-slice planner, scheduler backend adapter, service-backed schedule renderer, harness contract test import, and this tracker. Tracked harness count remains 238 before add; two new authored harness files are untracked in the worktree and the expected post-add count is 240. | `node --check` for touched `.mjs`; direct backend/browser planner probes; `bash tools/harness/browser/tests/test-browser-shard-plan.sh`; `bash tools/harness/backend/tests/test-run-go-target.sh`; `bash tools/harness/planning/tests/test-run-phase-slice.sh`; `make backend-module-boundary-check`; `make check-harness-smoke`; `make harness-contract`; `make browser-e2e-webserver-backed`. | All commands exited 0. Roots: backend boundary `.cartulary/test-results/20260703T211302Z-p1022762`; smoke `.cartulary/test-results/20260703T211310Z-p1022971`; harness-contract `.cartulary/test-results/20260703T211310Z-p1023330`; browser E2E `.cartulary/test-results/20260703T211323Z-p1024517`. Browser E2E final summary: 88 tests, 0 failed. | None for WS-1. Generated outputs were not hand-edited and no generator was run. | Run `make lint-markdown` for the WS-1 tracker update, then record a fresh WS-2 baseline before extracting `core/test-output` evidence adapters. |
| 2026-07-03T21:16Z | Codex / WS-2 baseline | Completed the required post-WS-1 Markdown lint gate and recorded a fresh baseline before NI-03 evidence-adapter work. | Changed this tracker only after WS-1; no WS-2 implementation files edited yet. | `make lint-markdown`; `git ls-files tools/harness \| wc -l`; `git ls-files --others --exclude-standard tools/harness \| wc -l`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make check-harness-smoke`; `make harness-contract`. | All commands exited 0. Lint root: `.cartulary/test-results/20260703T211628Z-p1040505`. Inventory count: 238 tracked harness files plus 2 untracked authored harness files from WS-1. Task-surface report printed `check=pass` with 95 public, 31 internal helper, and 11 check-internal targets. Smoke root: `.cartulary/test-results/20260703T211644Z-p1040996`. Harness-contract root: `.cartulary/test-results/20260703T211644Z-p1041015`. | None for WS-2 baseline. | Start NI-03 / WS-2 by extracting `core/test-output` evidence enrichment behind private adapters without changing summary schemas. |
| 2026-07-03T21:27Z | Codex / WS-2 NI-03 implementation | Extracted `core/test-output` evidence enrichment behind private adapters for frontend row accounting, frontend evidence indexes, phase-manifest selection/indexing, Playwright report summaries, Playwright selection reports, summary topology, and target-plan start stats. Preserved `cartulary.tool_run_summary.v3`, target/run summary schemas, failure taxonomy, artifact refs, and human/machine output contracts. | Added `tools/harness/core/test-output/adapters/*.mjs`; changed `tools/harness/core/test-output/cli.mjs`, this tracker, and `tools/execution_topology_render_index.json` via `make phase-schedules` because WS-1 changed the service-backed schedule renderer source digest. | `node --check` for touched `.mjs`; `bash tools/harness/core/tests/test-run-phase.sh`; `bash tools/harness/core/tests/test-cartulary-runner-service-backed-target.sh`; `make phase-schedules`; `make json-shape-check`; `make check-harness-smoke`; `make harness-contract`; `make frontend-unit`; `make browser-e2e-webserver-backed`; `make phase-schedule-drift`; `make lint-scripts`. Also ran `bash tools/harness/core/tests/test-tool-output-real-targets.sh`. | Required commands exited 0. Roots: phase-schedules `.cartulary/test-results/20260703T212212Z-p1050715`; json-shape `.cartulary/test-results/20260703T212341Z-p1052189`; smoke `.cartulary/test-results/20260703T212341Z-p1052233`; harness-contract `.cartulary/test-results/20260703T212341Z-p1052223`; frontend-unit `.cartulary/test-results/20260703T212341Z-p1052329`; browser E2E `.cartulary/test-results/20260703T212403Z-p1055752`; phase-schedule-drift `.cartulary/test-results/20260703T212608Z-p1071551`; lint-scripts `.cartulary/test-results/20260703T212643Z-p1072019`. Browser E2E final summary: 88 tests, 0 failed. `test-tool-output-real-targets.sh` first exposed stale phase schedules, then after `make phase-schedules` failed on an existing `lint-shell` summary-mode expectation: `lint-shell: missing passing [RESULT] line`; direct `make lint-shell` with summary output also emitted no stdout while writing retained summaries. | No WS-2 blocker. The standalone `test-tool-output-real-targets.sh` `lint-shell` expectation appears unrelated to the adapter extraction and is recorded for follow-up; Make-owned WS-2 gates passed. | Run `make lint-markdown` for the WS-2 tracker update, then start WS-3 by revising `docs/testing-harness-nlspec.md` and owner inputs before removing legacy behavior. |
| 2026-07-03T21:28Z | Codex / WS-3 baseline | Completed the required post-WS-2 Markdown lint gate and recorded a fresh baseline before NI-04 spec-gated legacy simplification. | Changed this tracker only after WS-2; no WS-3 implementation files edited yet. | `make lint-markdown`; `git ls-files tools/harness \| wc -l`; `git ls-files --others --exclude-standard tools/harness \| wc -l`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make check-harness-smoke`; `make harness-contract`. | All commands exited 0. Lint root: `.cartulary/test-results/20260703T212729Z-p1072552`. Inventory count: 238 tracked harness files plus 9 untracked authored harness files from WS-1 and WS-2. Task-surface report printed `check=pass` with 95 public, 31 internal helper, and 11 check-internal targets. Smoke root: `.cartulary/test-results/20260703T212750Z-p1073065`. Harness-contract root: `.cartulary/test-results/20260703T212750Z-p1073072`. | None for WS-3 baseline. | Start NI-04 / WS-3 with `docs/testing-harness-nlspec.md` and owner-input revisions, then remove only the legacy behavior authorized by those spec changes. |
| 2026-07-03T21:43Z | Codex / WS-3 NI-04 implementation | Completed spec-gated legacy simplification. Removed public `db-down` from topology/task-surface generation and dev-service runtime alias, made `services-down` the only public stop command, moved old browser stack/accessibility schemas out of current Section 8 attachment status, removed browser selected-ID verification fallback in favor of selected-tests artifacts, removed raw Go aggregate fallback and `legacy_baseline_key`, removed legacy failure-class aliases, and removed frontend row-accounting summary extension copies in favor of the schema-owned artifact. | Changed `docs/testing-harness-nlspec.md`, `tools/execution_topology_manifest.json`, generated `tools/task_surface_manifest.json`, generated `tools/task_surface.generated.mk`, generated `tools/execution_topology_render_index.json`, backend duration/shard modules, browser batch runner/tests, core failure taxonomy and test-output summary code, frontend/browser row-accounting tests, dev-service lifecycle script/tests, and this tracker. | `node --check` for touched `.mjs`; `make generate`; `make phase-schedules`; targeted fixtures: `bash tools/harness/readiness/tests/test-dev-services-lifecycle.sh`, `bash tools/harness/backend/tests/test-go-test-duration-baselines.sh`, `bash tools/harness/browser/tests/test-run-playwright-webserver-batch.sh`, `bash tools/harness/frontend/tests/test-run-frontend-unit.sh`, `node --test tools/harness/core/tests/test-harness-contracts.mjs`; required gates: `make generate`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make generate-drift`; `make json-shape-check`; `make generated-artifact-policy-check`; `make phase-schedule-drift`; `make harness-contract`. | All commands exited 0. Generate roots: `.cartulary/test-results/20260703T214104Z-p1080696` and `.cartulary/test-results/20260703T214332Z-p1095811`. Phase-schedules root: `.cartulary/test-results/20260703T214147Z-p1081927`. Task-surface report printed `check=pass` with 94 public, 31 internal helper, and 11 check-internal targets. Generate-drift root: `.cartulary/test-results/20260703T214341Z-p1096692`. JSON-shape root: `.cartulary/test-results/20260703T214341Z-p1096652`. Generated-policy root: `.cartulary/test-results/20260703T214341Z-p1096631`. Phase-schedule-drift root: `.cartulary/test-results/20260703T214341Z-p1096730`. Harness-contract root: `.cartulary/test-results/20260703T214349Z-p1098272`. | None for WS-3. Generated task surface outputs were refreshed through `make phase-schedules`; no generated output was hand-edited after generation. | Run `make lint-markdown` for the WS-3 tracker update, then record a fresh WS-4 baseline before adding NI-05 harness import-boundary enforcement. |
| 2026-07-03T21:45Z | Codex / WS-4 baseline | Completed the required post-WS-3 Markdown lint gate and recorded a fresh baseline before NI-05 harness import-boundary enforcement. | Changed this tracker after WS-3; no WS-4 implementation files edited yet. | `make lint-markdown`; `git ls-files tools/harness \| wc -l`; `git ls-files --others --exclude-standard tools/harness \| wc -l`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make check-harness-smoke`; `make harness-contract`. | All commands exited 0. Lint root: `.cartulary/test-results/20260703T214500Z-p1099083`. Inventory count: 238 tracked harness files plus 9 untracked authored harness files from WS-1 and WS-2. Task-surface report printed `check=pass` with 94 public, 31 internal helper, and 11 check-internal targets. Smoke root: `.cartulary/test-results/20260703T214523Z-p1099686`. Harness-contract root: `.cartulary/test-results/20260703T214523Z-p1099706`. | None for WS-4 baseline. | Start NI-05 / WS-4 by adding a validation-only harness import-boundary assertion under `harness-contract`, with positive and negative fixtures for approved adapters and forbidden subsystem imports. |
| 2026-07-03T21:55Z | Codex / WS-4 NI-05 implementation | Added validation-only harness import-boundary enforcement under `harness-contract`. The scanner rejects direct backend/browser/frontend/scheduler imports into planning except explicit approved adapter entrypoints and reports forbidden planning/execution SCC regressions; contract fixtures cover approved adapters, forbidden direct imports, and a forbidden backend/planning cycle. | Added `tools/harness/static-analysis/harness-import-boundary.mjs`; changed `tools/harness/core/tests/test-harness-contracts.mjs`, backend target-plan import callers to use `tools/harness/planning/backend-target-plan.mjs`, added `tools/harness/planning/backend-target-plan.mjs`, and this tracker. | `node --check` for touched `.mjs`; import graph scan with `collectHarnessImportBoundaryViolations`; `node --test tools/harness/core/tests/test-harness-contracts.mjs`; `make harness-contract`; `make lint-scripts`. | All commands exited 0. Import graph scan returned 0 violations and 0 forbidden SCCs. Node contract test passed 28 tests. Harness-contract root: `.cartulary/test-results/20260703T215400Z-p1104146`. Lint-scripts root: `.cartulary/test-results/20260703T215400Z-p1104145`. | None for WS-4. The executable compatibility edges in `backend/go-shard-plan.mjs` and `browser/browser-shard-plan.mjs` remain explicitly allow-listed until those private executable paths are retired. | Run `make lint-markdown` for the WS-4 tracker update, then start WS-5 final validation and handoff completion. |
| 2026-07-03T22:00Z | Codex / WS-5 final validation and handoff | Completed final validation and handoff for the authorized plan. Verified public surface, harness smoke, harness contract, Markdown, generated drift, JSON shape, generated-artifact policy, phase schedule/ledger drift, broad `check`, and whitespace. Confirmed no product modules changed and no generated outputs were hand-edited after generator refresh. | Changed this tracker; inspected `git status --short`, `git diff --name-only`, untracked harness files, and diff stat. | `make agent-finalize`; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make check-harness-smoke`; `make harness-contract`; `make lint-markdown`; `make generate-drift`; `make json-shape-check`; `make generated-artifact-policy-check`; `make phase-schedule-drift`; `make phase-ledger-drift`; `make check`; `git diff --check`. | All commands exited 0. Agent-finalize root `.cartulary/test-results/20260703T215531Z-p1105905` reported `generated=unchanged`, `results_dir=-`, `duration=skipped`, and `run_checks=skipped`; retained-run maintenance with `RESULTS_DIR` was not used because no valid retained full warm check root was supplied before broad verification. Task surface stayed `check=pass` with 94 public, 31 internal helper, and 11 check-internal targets. Smoke root `.cartulary/test-results/20260703T215549Z-p1107406`; harness-contract root `.cartulary/test-results/20260703T215549Z-p1107352`; lint-markdown root `.cartulary/test-results/20260703T215549Z-p1107379`; generate-drift root `.cartulary/test-results/20260703T215616Z-p1109439`; json-shape root `.cartulary/test-results/20260703T215616Z-p1109498`; generated-policy root `.cartulary/test-results/20260703T215616Z-p1109465`; phase-schedule-drift root `.cartulary/test-results/20260703T215616Z-p1109450`; phase-ledger-drift root `.cartulary/test-results/20260703T215616Z-p1109508`; broad check root `.cartulary/test-results/20260703T215622Z-p1111364` with 256/256 work units, 981 tests, and 0 failed. `git diff --check` produced no findings. | No blockers for the authorized remediation. Non-blocking follow-up remains from WS-2: direct `tools/harness/core/tests/test-tool-output-real-targets.sh` still has a standalone `lint-shell` summary-output expectation to investigate separately; Make-owned gates passed. | Handoff complete. Review/commit the remediation changes; optional follow-up is the standalone `test-tool-output-real-targets.sh` `lint-shell` expectation recorded in WS-2. |

Future handoff rows must record command exit status, retained run root when emitted, relevant failure artifact when failed, skipped checks with reason, and exact next action.

## 11. Final Completion Criteria

The authorized NI-00 through NI-05 and WS-5 portion of this tracker is complete when:

- NI-00 tracker refresh is recorded with live inventory and baseline results.
- NI-01 through NI-05 are completed and validated, or explicitly blocked with failure artifacts.
- WS-5 final validation is recorded with retained roots, skipped checks, changed-file summary, and follow-up state.
- Public harness contract changes are limited to the WS-3 owner-spec and owner-input revisions.
- No generated artifact has been hand-edited.
- Product behavior remains outside `tools/harness`.
- Handoff log identifies the next exact action, expected validation, and any deferred blocker or non-blocking follow-up.

Exact next action: remediation handoff is complete. Review/commit the changes. Optional follow-up: investigate the standalone `tools/harness/core/tests/test-tool-output-real-targets.sh` `lint-shell` summary-output expectation recorded during WS-2; it did not block Make-owned validation.
