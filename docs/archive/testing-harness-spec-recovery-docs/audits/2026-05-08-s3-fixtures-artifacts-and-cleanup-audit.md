---
doc_id: THR-S3-AUDIT-2026-05-08
title: S3 Fixtures Artifacts and Cleanup Audit
status: complete
role: recovery-audit
---

# S3 Fixtures Artifacts and Cleanup Audit

## Audit verdict

`pass_with_followups`

Sprint 3 is complete enough to support Sprint 4 service, environment, and
resource recovery. The required S3 outputs exist, artifact classes are separated
by authority and persistence, cleanup behavior is linked to entrypoints and
source evidence, external state hazards are represented, and unsupported runtime
claims are carried as ambiguities or source limits instead of being guessed.

No Sprint 4 work was started. No harness implementation, test, fixture,
snapshot, generated artifact, cleanup script, service, lockfile, or S4 output
was modified by this audit.

## Audit scope

| Scope item | Result | Evidence status | Evidence |
|---|---|---|---|
| Audit write surface | Audit artifact created under `docs/testing-harness-spec-recovery-docs/audits/`. | `observed` | This file. |
| S3 sprint status | S3 is marked `complete` with blocker `none`; S4 remains `not_started`. | `observed` | `03-sprint-plan.md` S3 and S4 sections. |
| Required S3 outputs | Required S3 output files exist. | `observed` | `artifact-ownership-matrix.md`, `cleanup-lifecycle-matrix.md`, `shared-state-hazard-list.md`, register updates, and S3 handoff. |
| Implementation edits | No implementation, fixture, snapshot, generated, cleanup, service, lockfile, or S4 paths were modified by this audit. | `runtime_observed` | `git status --short --branch` before and after audit file creation. |
| Audit command boundary | Only non-mutating inspection commands were run. | `runtime_observed` | Commands listed below. |

## Evidence reviewed

| Evidence area | Audit check | Result | Evidence status | Notes |
|---|---|---|---|---|
| Recovery controls | Checked charter, process rules, register templates, S0-S2 outputs, S3 handoff, and authority docs. | pass | `observed` | S3 follows the recovery-doc-only write boundary and evidence-label model. |
| S3 output presence | Checked all required S3 files listed in the sprint plan and audit plan. | pass | `observed` | All required files were present at audit start. |
| Artifact row coverage | Counted and sampled artifact rows. | pass_with_followup | `observed` | `ART-0001` through `ART-0030` cover fixtures, goldens, snapshots, generated outputs, reports, logs, caches, build outputs, browser state, external state, and package-script artifacts. |
| Cleanup row coverage | Counted and sampled cleanup lifecycle rows. | pass | `observed` | `CLN-0001` through `CLN-0020` cover Make cleanup, shell traps, Go cleanup, Postgres, S3, testservices, browser stack, Playwright, reset route, and package-script gaps. |
| Shared-state hazards | Counted and sampled S3 hazard rows. | pass | `observed` | `HAZ-S3-0001` through `HAZ-S3-0015` cover retained results, generated execution inputs, schedules, baselines, Go reports, Postgres, MinIO, browser state, package scripts, reset route, ports/processes, and external caches. |
| Ambiguity and source-limit updates | Checked S3 register additions. | pass | `observed/source_limit` | `AMB-0015` through `AMB-0022` and `SL-0008` through `SL-0011` cover S3-specific unknowns and non-executed runtime behavior. |
| Canonical fixtures versus derived reports | Compared S3 classifications with fixture/golden/snapshot paths and S1/S2 inputs. | pass | `observed/source_limit` | Canonical fixtures, goldens, and visual snapshots are not treated as disposable diagnostics; unknown update authority is recorded. |
| Generated artifacts | Checked generated artifact policy, generation scripts, drift check script, and S3 generated rows. | pass | `observed` | Generated roots and generated Make surface are downstream of policy, SQL/contracts/manifests, and generator commands; execution-driving generated artifacts are called out. |
| Runtime and retained artifacts | Checked result artifact discovery, scheduler/report paths, fixture-reporting, runner logs, and retained artifact source limits. | pass | `observed/source_limit` | Retained result provenance is not treated as current-run truth; freshness and schemas are deferred. |
| Cleanup behavior | Checked Make cleanup macros, generated recipes, shell traps, Go test cleanup helpers, service cleanup, browser lifecycle, Playwright teardown, and reset script evidence. | pass | `observed/source_limit` | Timeout, interrupt, live janitor, and external cleanup behavior remain source-limited where runtime evidence was not gathered. |
| External state | Checked Postgres, MinIO, browser runtime roots, Playwright state, process/port handling, reset-route behavior, and external Go caches. | pass | `observed/source_limit` | External state is represented in artifact rows, cleanup rows, and hazard rows with S4/S6 handoff. |
| Evidence discipline | Checked that unsupported S3 claims were either cited, marked `TODO:`, or carried in ambiguity/source-limit rows. | pass_with_followup | `observed/source_limit` | One cross-reference quality issue is recorded below; it does not change S3 readiness. |

## Commands run

All commands below were non-mutating inspection commands. The audit did not run
generation, formatting, cleanup, service-backed targets, browser E2E, reset
commands, snapshot refresh, baseline refresh, S4 work, or broad verification.

| Command | Result | Evidence status | Notes |
|---|---|---|---|
| `date -Is` | Exit 0; printed `2026-05-08T21:17:25-04:00`. | `runtime_observed` | Local audit timestamp evidence. |
| `git status --short --branch` | Exit 0; listed existing S3 recovery-doc changes before this audit file was added. | `runtime_observed` | Existing S3 files were uncommitted/untracked recovery docs under the permitted tree. |
| `git rev-parse HEAD` | Exit 0; printed `c4b2af30d0f39de6aeef10a7d5c7862fff19987f`. | `runtime_observed` | Matches the S3 handoff revision. |
| `ls docs/testing-harness-spec-recovery-docs/audits` | Exit 0. | `runtime_observed` | Confirmed prior S0-S2 audit files. |
| `rg --files docs/testing-harness-spec-recovery-docs \| sort` | Exit 0. | `runtime_observed` | Confirmed recovery output files. |
| Required-file `test -f` loop for S3 outputs | Exit 0 for all required paths. | `runtime_observed` | Confirmed all S3 outputs exist. |
| Targeted `rg` over S3 docs for `ART-*`, `CLN-*`, `HAZ-S3-*`, S3 `AMB-*`, S3 `SL-*`, `TODO:`, and `source_limit` | Exit 0. | `observed/runtime_observed` | Confirmed row IDs, register rows, and open gaps are represented. |
| Row-count `rg -c` commands | Exit 0; found `30` artifact rows, `20` cleanup rows, `15` S3 hazard rows, `8` S3 ambiguity rows, and `4` S3 source-limit rows. | `runtime_observed` | Counted table coverage. |
| Targeted `rg` over `Makefile`, `.gitignore`, and `tools/task_surface.generated.mk` | Exit 0. | `observed/runtime_observed` | Checked result roots, clean/distclean paths, ignored report/cache/build roots, generated recipes, and external Go cache defaults. |
| `sed` over `tools/generated_artifact_policy.json`, `scripts/generate-artifacts.sh`, and `scripts/check-generate-drift.sh` | Exit 0. | `observed` | Checked generated roots, generated markers, generator commands, scratch cleanup, and drift inputs. |
| Targeted `rg` over Playwright configs and E2E harness support files | Exit 0. | `observed` | Checked trace retention, setup/teardown, worker admin manifest, lock/state behavior, and session cleanup. |
| Targeted `rg` over `pgtest`, `s3test`, `suiteservices`, and `tools/testservices` | Exit 0. | `observed/source_limit` | Checked fixture DB/bucket cleanup, retained suite state, janitor/reclaim behavior, event summaries, and runtime source limits. |
| Targeted `rg` over browser lifecycle and reset scripts | Exit 0. | `observed/source_limit` | Checked runtime roots, ports, process groups, traps, cleanup status, standalone DB cleanup, and reset-route validation. |
| Targeted `rg` over artifact discovery, result artifacts, scheduler, fixture reporting, and Go runner code | Exit 0. | `observed/source_limit` | Checked machine-consumed summaries, logs, scheduler events, fixture report schema, and newest-retained artifact selection. |
| Targeted `rg` over visual snapshot, fixture, golden, and S2 entrypoint sources | Exit 0. | `observed/source_limit` | Checked visual snapshot assertions and fixture/golden helper use; no repo-owned refresh command was proven. |
| `sed` samples of S3 artifact, cleanup, and hazard rows | Exit 0. | `observed` | Checked detailed external-state rows and source-limit handoffs. |

`make agent-finalize` was not run because this audit's non-scope prohibited
mutating maintenance commands and generated-artifact refresh paths.

## Findings

| Finding ID | Finding | Severity | Evidence status | Evidence reference | Disposition |
|---|---|---|---|---|---|
| AUD-S3-0001 | Required S3 outputs are present, S3 is marked complete, and S4 remains not started. | none | `observed` | `03-sprint-plan.md`; S3 output files; S3 handoff. | Pass. |
| AUD-S3-0002 | The artifact matrix is broad enough for S4: it covers committed fixtures, goldens, snapshots, generated outputs, derived reports, retained reports, failure-only artifacts, caches, build outputs, browser state, package-script artifacts, and external state. | none | `observed/source_limit` | `artifact-ownership-matrix.md` `ART-0001` through `ART-0030`; S1 `HI-*`; S2 `EP-*`. | Pass. |
| AUD-S3-0003 | Canonical fixtures, goldens, and visual snapshots are not misclassified as derived reports; unresolved update authority is recorded. | none | `observed/source_limit` | `ART-0001` through `ART-0003`, `ART-0005`; `AMB-0015`; `AMB-0022`; `SL-0008`. | Pass. |
| AUD-S3-0004 | Generated artifacts distinguish upstream owners from downstream outputs, including generated Go/TS, generated Make surface, ledgers, schedules, and duration baselines. | none | `observed` | `ART-0006` through `ART-0011`; `tools/generated_artifact_policy.json`; `scripts/generate-artifacts.sh`; `scripts/check-generate-drift.sh`. | Pass. |
| AUD-S3-0005 | Derived reports that are consumed by execution, drift, scheduling, investigation, or fixture reporting are called out rather than treated as human-only diagnostics. | none | `observed/source_limit` | `ART-0009`, `ART-0010`, `ART-0011`, `ART-0014`, `ART-0015`, `ART-0017`, `ART-0029`; `AMB-0016`. | Pass. |
| AUD-S3-0006 | Retained run artifacts and failure-only artifacts are classified with freshness and provenance limits. | none | `observed/source_limit` | `ART-0013` through `ART-0018`; `HAZ-S3-0001`; `SL-0009`; `SL-0010`. | Pass. |
| AUD-S3-0007 | Cleanup coverage is complete enough at static-evidence level and avoids overclaiming timeout, interrupt, and live janitor behavior. | none | `observed/source_limit` | `cleanup-lifecycle-matrix.md` `CLN-0001` through `CLN-0020`; `AMB-0018`; `AMB-0019`; `SL-0011`. | Pass. |
| AUD-S3-0008 | External state dependencies are represented and handed to S4/S6: Postgres, MinIO, browser runtime roots, Playwright shared state, reset route, ports, process groups, containers, stale fixtures, and external caches. | none | `observed/source_limit` | `ART-0024` through `ART-0028`; `CLN-0005` through `CLN-0019`; `HAZ-S3-0007` through `HAZ-S3-0015`. | Pass. |
| AUD-S3-0009 | Package-script artifacts and cleanup remain authority-ambiguous but are explicitly represented instead of normalized into Make behavior. | low | `observed` | `ART-0030`; `CLN-0020`; `HAZ-S3-0013`; `AMB-0020`; S2 `EP-0016`, `EP-0018`, `EP-0019`. | Follow-up authority decision; not a blocker to S4. |
| AUD-S3-0010 | Artifact rows `ART-0025` and `ART-0026` included a future S4 hazard placeholder in the inventory-ID column, but no S4 hazard ID existed yet. | low | `observed` | `artifact-ownership-matrix.md` `ART-0025`, `ART-0026`; `shared-state-hazard-list.md` current IDs are `HAZ-S3-*`. | Follow-up replaced the placeholder with existing S3 hazard IDs; the rows otherwise cite valid `HI-0028`, `HI-0029`, source files, cleanup rows, and S3 hazards. |
| AUD-S3-0011 | S3 does not claim runtime proof for commands or cleanup paths it intentionally did not execute. | none | `observed/source_limit` | `source-limit-log.md` `SL-0008` through `SL-0011`; S3 handoff commands and non-scope statements. | Pass. |
| AUD-S3-0012 | The audit itself stayed inside recovery documentation and did not begin S4. | none | `runtime_observed` | `git status --short --branch`; this file. | Pass. |

## Blocking issues

No blocking S3 issues were found.

S4 may proceed from the S3 outputs, especially:

- `artifact-ownership-matrix.md` rows `ART-0025` through `ART-0028`
- `cleanup-lifecycle-matrix.md` rows `CLN-0011` through `CLN-0019`
- `shared-state-hazard-list.md` rows `HAZ-S3-0007` through `HAZ-S3-0015`
- `ambiguity-register.md` rows `AMB-0018` through `AMB-0019`
- `source-limit-log.md` row `SL-0011`

S4 should not reinterpret S3's source-limited runtime cleanup statements as
runtime-observed behavior.

## Follow-up issues

| Follow-up ID | Issue | Target sprint | Why non-blocking for S4 |
|---|---|---|---|
| AUD-S3-FU-0001 | Replace or clarify the future S4 hazard placeholder in `ART-0025` and `ART-0026` because S4 has not created that hazard ID. | S3 follow-up complete; future S4 rows may create their own IDs later. | The rows now contain valid inventory, evidence, cleanup, S3 hazard, and S4 handoff references. |
| AUD-S3-FU-0002 | Decide update authority for committed fixtures, goldens, and visual snapshots. | S8 or owner decision | `AMB-0015`, `AMB-0022`, and `TODO: update_rule_unknown` preserve the gap. |
| AUD-S3-FU-0003 | Define retained artifact freshness and run-selection rules for explain, fixture, drift, and baseline tooling. | S5/S8 | `ART-0013` through `ART-0017`, `HAZ-S3-0001`, `AMB-0017`, and `SL-0009` preserve the gap. |
| AUD-S3-FU-0004 | Recover failure-only artifact schemas for runner logs, watchdog JSON, Playwright traces, screenshots, videos, and reports. | S5 | S3 classifies the artifacts and records `SL-0010` without claiming schema completeness. |
| AUD-S3-FU-0005 | Recover live external-state cleanup, stale fixture janitor, active connection, port release, timeout, and interrupt behavior. | S4/S6 | S3 maps the surfaces and records `SL-0011`, `AMB-0018`, and `AMB-0019`. |
| AUD-S3-FU-0006 | Decide whether direct package-script artifacts and cleanup behavior are first-class harness contracts or developer conveniences. | S8 | `ART-0030`, `CLN-0020`, `HAZ-S3-0013`, and `AMB-0020` preserve the authority gap. |
| AUD-S3-FU-0007 | Decide whether external Go cache paths are in the harness contract or only tool-managed defaults. | S8 | `ART-0024`, `HAZ-S3-0015`, and `AMB-0021` preserve the low-severity gap. |

## Source-limit and ambiguity updates

No mandatory register updates were found during the audit. Existing S3 rows
cover the unsupported assumptions surfaced by this audit:

- update authority: `AMB-0015`, `AMB-0022`
- derived report stability: `AMB-0016`
- retained artifact freshness: `AMB-0017`, `SL-0009`
- timeout and interrupt cleanup: `AMB-0018`, `SL-0011`
- destructive janitor boundaries: `AMB-0019`, `SL-0011`
- package-script authority: `AMB-0020`
- external Go cache cleanup: `AMB-0021`
- non-executed mutating/runtime commands: `SL-0008` through `SL-0011`

The future S4 hazard placeholder issue is a row-quality follow-up, not a new
source-limit or ambiguity.

## Validation checklist

| Check | Result | Notes |
|---|---|---|
| S3 output files exist and are internally consistent. | pass_with_followup | Required files exist; the future-looking cross-reference in `ART-0025` and `ART-0026` has been replaced by existing S3 hazard IDs. |
| Artifact ownership matrix covers all S1/S2 fixture, artifact, report, cache, generated, ignored, runtime, and external-state classes. | pass | `ART-0001` through `ART-0030`. |
| Cleanup lifecycle matrix covers Make cleanup, shell traps, Go cleanup, Postgres/S3 cleanup, testservices cleanup, browser stack cleanup, Playwright cleanup, reset route, and package-script gaps. | pass | `CLN-0001` through `CLN-0020`. |
| Shared-state hazard list covers retained artifacts, generated execution inputs, schedules, baselines, Go reports, Postgres, MinIO, browser state, package scripts, reset route, ports/processes, and external caches. | pass | `HAZ-S3-0001` through `HAZ-S3-0015`. |
| Ambiguity and source-limit updates cover unknown update authority, retained artifact freshness, timeout/interrupt cleanup, janitor boundaries, package-script authority, and runtime-only behavior. | pass | `AMB-0015` through `AMB-0022`; `SL-0008` through `SL-0011`. |
| Canonical fixtures are not treated as derived diagnostics. | pass | `ART-0001` through `ART-0005`; `ART-0012`. |
| Derived reports consumed by execution are identified as machine-consumed downstream artifacts. | pass | `ART-0009` through `ART-0011`, `ART-0014`, `ART-0015`, `ART-0017`, `ART-0029`. |
| Reused-across-runs artifacts are separated from temporary and failure-only artifacts. | pass | `ART-0011`, `ART-0013`, `ART-0022`, `ART-0024` are distinct from `ART-0016`, `ART-0018`, `ART-0023`, and `ART-0027`. |
| External service state has owner/reset/isolation notes or S4 handoff. | pass | `ART-0025` through `ART-0028`; `CLN-0005` through `CLN-0019`; `HAZ-S3-0007` through `HAZ-S3-0015`. |
| No unsupported assumptions remain unregistered. | pass_with_followup | No blocker; one cross-reference quality follow-up is listed above. |
| No prohibited mutations occurred. | pass | Audit added this recovery-doc file only. |
| S4 readiness conclusion is explicit. | pass | Verdict is `pass_with_followups`; no blockers to S4. |

## Implementation-change audit

| Check | Result |
|---|---|
| Harness implementation files modified | `no` |
| Test logic modified | `no` |
| CI behavior modified | `no` |
| Fixture contents modified | `no` |
| Golden or snapshot contents modified | `no` |
| Cleanup scripts modified | `no` |
| Services modified | `no` |
| Generated code modified | `no` |
| Generated manifests modified | `no` |
| Duration baselines modified | `no` |
| Phase ledgers or schedules modified | `no` |
| Lockfiles modified | `no` |
| Sprint 4 work started | `no` |
| Only recovery docs changed | `yes` |

## Final audit note

S3 is ready for S4 with follow-ups. S4 should build service lifecycle and
resource allocation maps from the external-state rows and hazards without
rediscovering fixture authority or cleanup surfaces, and it should preserve
S3's evidence boundary: static cleanup/source inspection is not runtime proof
for timeout, interrupt, stale janitor, active connection, or failure-bundle
behavior.
