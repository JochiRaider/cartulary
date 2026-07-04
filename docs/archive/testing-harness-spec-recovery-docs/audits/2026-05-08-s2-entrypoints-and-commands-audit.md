---
doc_id: THR-S2-AUDIT-2026-05-08
title: S2 Entrypoints and Commands Audit
status: complete
role: recovery-audit
---

# S2 Entrypoints and Commands Audit

## Audit verdict

`pass_with_followups`

Sprint 2 is complete enough to support Sprint 3 fixture, artifact, and cleanup
recovery. The S2 artifacts exist, the current task surface still reconciles to
the S2 command-family map, aggregate and scheduler flows match the execution
manifests, package and CI-adjacent entrypoints are classified separately, and
runtime-only or authority-sensitive gaps are recorded as source limits or
ambiguities.

No Sprint 3 recovery content was created. No harness implementation, test,
fixture, CI, generated, cleanup, service, lockfile, or generated-code file was
modified by this audit.

## Audit scope

| Scope item | Result | Evidence status | Evidence |
|---|---|---|---|
| Audit write surface | Audit artifact created under `docs/testing-harness-spec-recovery-docs/audits/`. | `observed` | This file. |
| S2 sprint status | S2 is marked `complete` with blocker `none`. | `observed` | `03-sprint-plan.md` S2 section. |
| Required S2 outputs | Required S2 output files exist and are linked by the sprint plan or handoff. | `observed` | `entrypoint-command-map.md`, `sequencing-assumption-list.md`, S2 handoff, S2 register updates. |
| Implementation edits | No implementation, test, fixture, CI, cleanup, generated, service, or lockfile paths were modified by this audit. | `runtime_observed` | `git status --short --branch` before audit work and after audit file creation. |
| Sprint 3 work | Not started. | `observed` | Audit reviewed S2 readiness only; no artifact ownership, cleanup lifecycle, service lifecycle, failure-mode, or NLSpec output was created. |

## Evidence reviewed

| Evidence area | Audit check | Result | Evidence status | Notes |
|---|---|---|---|---|
| Recovery controls | Checked recovery process, checklist templates, S1 audit shape, and S2 handoff. | pass | `observed` | S2 uses the required evidence labels and source-limit/ambiguity model. |
| S2 artifacts | Checked entrypoint map, sequencing list, handoff, sprint-plan status, inventory addendum, ambiguity register, and source-limit log. | pass | `observed` | `EP-0001` through `EP-0021`, `SEQ-0001` through `SEQ-0024`, `HI-S2-0001` through `HI-S2-0004`, `AMB-0011` through `AMB-0014`, and `SL-0006` through `SL-0007` are present. |
| Make task surface | Re-ran task-surface report and manifest summaries. | pass | `runtime_observed/observed` | Current surface is still `122` targets: `77` public, `17` check-internal, `28` helper-only, plus `47` logical harness smoke checks. |
| Target-family reconciliation | Compared current targets and representative `explain-target` output against S2 target families. | pass | `runtime_observed` | Families remain `print_help`, `phase_command`, `alias`, `go_target`, `service_backed_target`, `service_backed_schedule`, `check_schedule`, `sequence`, `browser_batch`, `node_tool`, `summary_target`, and `cleanup`. |
| Aggregate flows | Checked `test-fast`, `test`, `ci`, and `release-check` against `tools/task_surface_manifest.json`. | pass | `observed/runtime_observed` | Sequence step counts are `2`, `2`, `2`, and `5`, matching S2. |
| Check scheduler | Checked `tools/check_schedule_manifest.json` and representative `make explain-target TARGET=check`. | pass | `observed/runtime_observed` | `check` still has `96` work units and resource limits for host, suite-service, Postgres, MinIO, process, and browser lanes. |
| Service-backed scheduler | Checked `tools/service_backed_schedule_manifest.json` and service-backed target guidance. | pass | `observed/runtime_observed` | `test-service-backed` and `check-service-backed` have `8` work-unit sources; `test-fast-service-backed` has `4`. |
| Browser batch surface | Checked browser batch manifest, browser target guidance, browser scripts, and Playwright configs. | pass | `observed/runtime_observed` | Stages remain `webserver-backed`, `functional`, `support`, `stateful`, `measurement`, `visual`, `resettable`, and `isolated`. |
| Package scripts | Checked root, app, and package manifests. | pass | `observed` | Root and `apps/web` scripts are alternate entrypoints; `packages/*` manifests currently declare no scripts. |
| CI-adjacent surface | Checked `.github/**` absence and `scripts/ci/**`. | pass_with_followup | `observed/source_limit` | Provider workflow files remain absent; recoverable CI surface is `make ci`, `scripts/ci/verify.sh`, and `scripts/ci/check-deployable-shape.sh`. |
| Hidden and indirect paths | Searched Make, manifests, scripts, docs, and package files for direct script invocations and known uninvoked surfaces. | pass | `observed` | Browser wrappers, CI scripts, package aliases, library-only `render-phase-ledger.mjs`, and uninvoked `scripts/test-run-go-target-fast.sh` are represented by S2 rows or prior uninvoked-source rows. |
| Environment/default surface | Searched Make, scripts, configs, test utilities, and S2 docs for relevant env names and defaults. | pass_with_followup | `observed/source_limit` | Declared env/default surfaces are recorded; exact cross-layer precedence remains `AMB-0012`. |
| Mutating and cleanup commands | Checked `generate`, `format`, baseline refresh, release evidence, `clean`, and `distclean` classification without executing them. | pass_with_followup | `observed/source_limit` | S2 records these as mutating or cleanup entrypoints and defers ownership/idempotency to later sprints. |

## Commands run

All commands below were non-mutating discovery or inspection commands. They did
not run broad gates, start services, run browser E2E, regenerate artifacts,
format files, clean paths, refresh baselines, or begin Sprint 3.

| Command | Result | Evidence status | Notes |
|---|---|---|---|
| `git status --short --branch` | Exit 0; before the audit, only existing S2 recovery-doc changes were listed. | `runtime_observed` | Re-run after writing this audit file showed only recovery-doc changes plus this audit artifact. |
| `date -Is` | Exit 0; printed `2026-05-08T20:37:50-04:00`. | `runtime_observed` | Audit timestamp evidence from the local environment. |
| `git rev-parse HEAD` | Exit 0; printed `9e523d9b110a7433ed08d4c35474f63f8c6c8080`. | `runtime_observed` | Matches S2 handoff revision. |
| `rg --files docs/testing-harness-spec-recovery-docs \| sort` | Exit 0. | `runtime_observed` | Confirmed S2 artifacts and prior audit files exist. |
| `make help` | Exit 0. | `runtime_observed` | Confirmed compact public command surface. |
| `make help-all` | Exit 0. | `runtime_observed` | Confirmed public tiers and command names. |
| `make task-guide` | Exit 0. | `runtime_observed` | Confirmed role guidance and retained-artifact references without running verification. |
| `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` | Exit 0; check passed. | `runtime_observed` | Confirmed `77` public, `17` check-internal, `28` helper-only targets and `47` logical harness smoke checks. |
| `make target-plan` | Exit 0. | `runtime_observed` | Confirmed backend target families and service-backed classification. |
| `make target-plan-json \| jq ...` | Exit 0 after a corrected shape query; reported `195` backend plan rows across five backend targets. | `runtime_observed` | An initial exploratory jq expression assumed an object shape and exited 5; rerun confirmed the array shape. No repository files changed. |
| `make explain-phase PHASE=phase0` through `phase6` | Exit 0 for each phase. | `runtime_observed` | Confirmed active phase manifests and execution dependencies remain phase0 through phase6. |
| Representative `make explain-target TARGET=<target> DETAIL=summary` commands | Exit 0 for all selected representatives. | `runtime_observed` | Targets included `doctor`, `backend-unit`, `backend-store`, `frontend-unit`, `test-service-backed`, `check-service-backed`, `check`, `test-fast`, `test`, `ci`, `release-check`, `browser-e2e-webserver-backed`, `browser-e2e`, `phase-slice`, `task-surface-report`, `check-harness-smoke`, `format`, `generate`, `clean`, and `license-report`. |
| `jq` manifest inspections | Exit 0 for corrected manifest summaries. | `observed/runtime_observed` | Confirmed task counts, sequence step counts, check work-unit count, service-backed source counts, and browser stages. |
| Package manifest `jq` inspections | Exit 0. | `observed/runtime_observed` | Confirmed root scripts, app scripts, and empty `packages/*` scripts. |
| `.github/**` absence check and `find scripts/ci` | Exit 0. | `observed/source_limit` | `.github` produced no files; `scripts/ci/check-deployable-shape.sh` and `scripts/ci/verify.sh` exist. |
| `git check-ignore -v ... \|\| true` | Exit 0. | `runtime_observed` | Confirmed ignored roots for `.cartulary/test-results`, `coverage`, and `tmp`; some queried roots remain policy surfaces rather than matched ignore rows. |
| Targeted `rg`, `sed`, `find`, and `jq` inspections | Exit 0 except exploratory schema checks noted above. | `observed/runtime_observed` | Checked CLI usage strings, environment names, hidden command paths, Playwright/Vitest configs, CI scripts, and known uninvoked/library-only surfaces. |

## Findings

| Finding ID | Finding | Severity | Evidence status | Evidence reference | Disposition |
|---|---|---|---|---|---|
| AUD-S2-0001 | Required S2 artifacts are present and S2 is marked complete with blocker `none`. | none | `observed` | `03-sprint-plan.md`; S2 output files. | Pass. |
| AUD-S2-0002 | Current Make/task-surface counts match S2: `122` total targets, split into `77` public, `17` check-internal, and `28` helper-only. | none | `runtime_observed/observed` | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`; `tools/task_surface_manifest.json`; `entrypoint-command-map.md`. | Pass. |
| AUD-S2-0003 | S2 command-family reconciliation covers every current task-surface target family and gives stable entrypoint IDs for S3 linking. | none | `observed/runtime_observed` | Target reconciliation table; representative `make explain-target` output. | Pass. |
| AUD-S2-0004 | Aggregate command flows and scheduler paths are documented accurately at declaration level. | none | `observed/runtime_observed` | Task-surface sequence manifest, check schedule manifest, service-backed schedule manifest, browser batch manifest. | Pass. |
| AUD-S2-0005 | Package scripts are correctly recorded as alternate entrypoints that may bypass Make result-root and scheduler policy. | low | `observed` | `package.json`; `apps/web/package.json`; `entrypoint-command-map.md` package script table; `AMB-0011`. | Follow-up authority decision; not a blocker to S3. |
| AUD-S2-0006 | Provider CI remains source-limited because `.github/**` is absent, while `make ci` and `scripts/ci/**` are mapped as provider-neutral CI evidence. | low | `observed/source_limit` | `.github` absence check; `scripts/ci/**`; `EP-0017`; `AMB-0001`; `SL-0001`. | Follow-up provider decision; not a blocker to S3. |
| AUD-S2-0007 | Hidden and indirect invocation paths found by targeted search are already represented by S2 rows, uninvoked-surface rows, or source limits. | none | `observed` | Browser wrapper search; `scripts/ci/**`; `render-phase-ledger.mjs` import search; `scripts/test-run-go-target-fast.sh` uninvoked row. | Pass. |
| AUD-S2-0008 | Environment variables and defaults are recorded at S2 command-surface level, and unresolved cross-layer precedence is explicitly ambiguous. | low | `observed/source_limit` | Env search; `EP-0002`, `EP-0007`, `EP-0009`, `EP-0010`, `EP-0016`, `EP-0018`, `EP-0019`, `EP-0020`; `AMB-0012`. | Follow-up in S4/S5/S9; not a blocker to S3. |
| AUD-S2-0009 | Mutating maintenance and cleanup commands are not mislabeled as ordinary validation and were not executed during the audit. | none | `observed/source_limit` | `EP-0020`; `SEQ-0022`; representative `explain-target` output for `format`, `generate`, `clean`, `license-report`. | Pass with later S3/S8 ownership work. |
| AUD-S2-0010 | S2 avoids unsupported runtime claims: broad gates, browser/service targets, cleanup, format, generate, CI/release runtime, and failure scenarios remain source-limited. | none | `observed/source_limit` | `SL-0006`, `SL-0007`, `AMB-0014`; S2 handoff. | Pass. |
| AUD-S2-0011 | The working-tree scope remains recovery-doc-only. | none | `runtime_observed` | `git status --short --branch`. | Pass. |

## Blocking issues

No blocking S2 issues were found.

Sprint 3 may proceed from:

- `docs/testing-harness-spec-recovery-docs/entrypoint-command-map.md`
- `docs/testing-harness-spec-recovery-docs/sequencing-assumption-list.md`
- `docs/testing-harness-spec-recovery-docs/harness-inventory.md`
- `docs/testing-harness-spec-recovery-docs/uninvoked-surface-list.md`
- `docs/testing-harness-spec-recovery-docs/ambiguity-register.md`
- `docs/testing-harness-spec-recovery-docs/source-limit-log.md`
- `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s2-entrypoints-and-commands.md`

## Follow-up issues

| Follow-up ID | Issue | Target sprint | Why non-blocking for S3 |
|---|---|---|---|
| AUD-S2-FU-0001 | Decide whether CI is external, absent, or represented only by `scripts/ci/**` plus `make ci`. | S8/S9 or maintainer decision | S2 records provider CI as source-limited and maps local/provider-neutral CI entrypoints. |
| AUD-S2-FU-0002 | Recover runtime success/failure behavior for broad gates, service-backed targets, browser E2E, cleanup, format, generate, CI, release, and failure scenarios. | S4/S5/S6 | S2 is declaration-level complete and explicitly source-limits runtime-only behavior. |
| AUD-S2-FU-0003 | Decide whether direct package scripts are supported harness contracts or developer convenience aliases. | S8/S9 | S2 records package scripts separately and does not normalize them into Make behavior. |
| AUD-S2-FU-0004 | Define public environment-variable override contracts and precedence across Make, scripts, schedulers, Playwright, Vitest, and service wrappers. | S4/S5/S9 | S2 records env surfaces and defaults, while `AMB-0012` prevents over-specifying precedence. |
| AUD-S2-FU-0005 | Recover artifact schemas, fixture ownership, retained artifact provenance, and cleanup behavior from S2 entrypoint IDs. | S3/S5 | S2 provides stable entrypoint IDs; ownership and lifecycle are intentionally later-sprint work. |
| AUD-S2-FU-0006 | Recover service lifecycle, resource allocation, browser stack lifecycle, reset behavior, and timing hazards. | S4/S6 | S2 maps command paths and scheduler resource lanes without claiming runtime lifecycle proof. |
| AUD-S2-FU-0007 | Decide treatment for planned phase7/phase8 registry entries. | S7 or maintainer decision | Active command evidence remains phase0 through phase6; planned files are source-limited. |
| AUD-S2-FU-0008 | Decide whether `scripts/test-run-go-target-fast.sh` is retired, manual-only, or missing a task-surface row. | S8 or maintainer decision | It is explicitly recorded as uninvoked and does not block S3 artifact mapping. |

## Validation checklist

| Check | Result | Notes |
|---|---|---|
| Record git status before audit. | pass | Initial status showed only S2 recovery-doc changes. |
| Confirm no non-recovery-doc changes are present or created. | pass | Final status remains recovery-doc-only. |
| Confirm S2 artifacts exist and are linked. | pass | Entrypoint map, sequencing list, S2 handoff, and register updates exist. |
| Re-run non-mutating discovery commands. | pass | Help, task guide, task-surface report, target plan, phase explain, and representative target explain commands all completed. |
| Reconcile task-surface counts and target families. | pass | Current counts and families match S2. |
| Cross-check package scripts and CLI surfaces. | pass | Package scripts, CLI usage strings, Playwright/Vitest configs, testservices, and CI scripts inspected. |
| Cross-check aggregate, scheduler, browser, service, and node-tool flows. | pass | Manifests and representative target guidance match S2 flow rows. |
| Search for hidden or indirect command paths. | pass | Findings are already covered by S2 rows, uninvoked rows, or source limits. |
| Check duplicated, deprecated, conflicting, and uninvoked surfaces. | pass | Package-script duplication, provider CI absence, library-only phase ledger renderer, and uninvoked Go-target fast script are recorded. |
| Check env vars, defaults, side effects, and cleanup claims. | pass | S2 records declared surfaces and defers unresolved precedence/lifecycle questions. |
| Verify unsupported or runtime-only claims are source-limited. | pass | `SL-0006`, `SL-0007`, and `AMB-0014` cover the runtime gaps. |
| Separate blockers from follow-ups. | pass | No blockers found; follow-ups listed above. |
| Assign verdict. | pass | Verdict is `pass_with_followups`. |
| State whether S3 may proceed. | pass | S3 may proceed from S2 entrypoint IDs and sequencing rows. |
| Record final git status. | pass | See implementation-change audit below. |

## Implementation-change audit

| Check | Result |
|---|---|
| Harness implementation files modified | `no` |
| Test logic modified | `no` |
| CI behavior modified | `no` |
| Fixture contents modified | `no` |
| Cleanup scripts modified | `no` |
| Services modified | `no` |
| Generated code modified | `no` |
| Generated manifests modified | `no` |
| Lockfiles modified | `no` |
| Sprint 3 work started | `no` |
| Only recovery docs changed | `yes` |

Final `git status --short --branch` after writing this audit file listed only
existing S2 recovery-doc changes plus this new audit artifact under
`docs/testing-harness-spec-recovery-docs/audits/`.

## Final audit note

S2 is ready for Sprint 3. The entrypoint map is row-complete for the current
declared command surface, and later recovery work can safely attach artifacts,
fixtures, cleanup, and generated-output ownership to the stable `EP-*` and
`SEQ-*` identifiers without treating unresolved runtime, CI-provider,
environment-precedence, or authority questions as settled behavior.
