---
doc_id: THR-S6-HANDOFF-2026-05-08
title: S6 Hazards Resources Timing Handoff
status: active
role: sprint-handoff
---

# S6 Hazards, Resources, and Timing Handoff

## Session Metadata

| Field | Value |
|---|---|
| Sprint | S6: Hazards, resources, and timing |
| Status | `complete` |
| Repository root | `/home/askahn/code/cartulary` |
| Branch | `main...origin/main` |
| HEAD revision | `900d6d858982e8db00b0b15d950d708b43e9e7c9` |
| Timestamp recorded | `2026-05-08T23:50:35-04:00` |
| Platform | `Linux DeskRip 6.6.114.1-microsoft-standard-WSL2 x86_64 GNU/Linux` |
| Recovery write boundary | `docs/testing-harness-spec-recovery-docs/**` |
| Evidence status | `observed/source_limit` unless a source row says otherwise |

## Commands Inspected

S6 used static inspection only: `git status --short --branch`,
`git rev-parse HEAD`, `date --iso-8601=seconds`, `uname -a`, `rg` searches,
`sed` reads, and documentation diffs. Source searches covered the required
terms: `sleep`, `timeout`, `deadline`, `retry`, `backoff`, `poll`, `wait`,
`ready`, `readyz`, `healthcheck`, `debounce`, `setTimeout`, `setInterval`,
`worker`, `parallel`, `shard`, `resource`, `lane`, `lock`, `lease`, `reaper`,
`janitor`, `port`, `cleanup`, `signal`, and `interrupt`.

No service-backed, browser, Docker, Compose, reset, cleanup, formatter,
generator, baseline refresh, snapshot update, lockfile, or broad verification
command was run.

## Outputs Updated

| Output | S6 result |
|---|---|
| `race-timing-resource-register.md` | Added resource coverage, hazard rows `RTR-0001` through `RTR-0021`, failure/partial-completion linkage, and per-resource classification. |
| `concurrency-model-notes.md` | Added concurrency notes `CONC-0001` through `CONC-0014` covering Make, schedulers, services, DBs, buckets, browser stacks, Playwright workers, package scripts, local-dev services, generated inputs, and caches. |
| `timeout-retry-register.md` | Added timing rows `TMR-0001` through `TMR-0033` for fixed sleeps, polling loops, retries, timeouts, readiness checks, cleanup waits, signal waits, locks, and debounce/watch search results. |
| `shared-state-hazard-list.md` | Added S6 routing rows `HAZ-FU-0001` through `HAZ-FU-0015` tying S3/S4 hazards to S6 registers and authority/preservation rows. |
| `preservation-matrix.md` | Added preservation rows `PRES-0001` through `PRES-0020` covering every major subsystem and required owner questions. |
| `harness-authority-map.md` | Added authority rows `AUTH-0001` through `AUTH-0015` for reset route, package scripts, local-dev, stale janitors, env, platform, caches, generated artifacts, scheduler lanes, retained artifacts, visual snapshots, and CI gaps. |
| `main-spec-conflict-list.md` | Added conflict rows `MSC-0001` through `MSC-0010` for product-spec-adjacent risks without resolving them. |
| `ambiguity-register.md` | Added follow-up rows `AMB-FU-0007` through `AMB-FU-0010` while preserving prior open ambiguity rows. |
| `source-limit-log.md` | Added `SL-FU-0006` preserving S6 static-only scope and runtime source limits. |
| `audits/2026-05-08-s6-hazards-resources-timing-audit.md` | Refreshed audit verdict, coverage checks, and no-change audit. |
| `03-sprint-plan.md` | Marked S6 documentation tasks and exit criteria complete. |

## Status Checklist

| Field | Content |
|---|---|
| Status | `complete` |
| Blockers | Runtime service, browser, cleanup, interrupt, reset, stale janitor, platform, env precedence, retained artifact freshness, and owner-decision gaps remain source-limited or maintainer-decision-required. |
| Findings | Scheduler lanes are logical scheduling constraints only; destructive cleanup and reset behavior require owner authority; live readiness/cleanup strength was not proven; direct package scripts bypass Make-owned result roots and scheduler policy; generated files are downstream inputs but not behavior owners. |
| Handoff notes | S7 can consume `RTR-*`, `TMR-*`, `CONC-*`, `PRES-*`, `AUTH-*`, and `MSC-*` rows directly for NLSpec resource/timing rules, acceptance matrix inputs, and roadmap classifications. |
| Evidence status | `observed`, `observed/source_limit`, `source_limit`, or `maintainer_decision_required` at row level. |
| No-change audit | No harness implementation, timing, retry, sharding, allocation, test, fixture, cleanup, generated, lockfile, service, browser, or package behavior was changed. |

## S7-Ready Inputs

- `race-timing-resource-register.md` can drive shared-resource and hazard
  requirements.
- `timeout-retry-register.md` can drive timing, retry, timeout, readiness, and
  cleanup-wait requirements.
- `concurrency-model-notes.md` distinguishes model, guarantee, and source
  limit for schedulers, services, browser, Playwright, package scripts, and
  local-dev flows.
- `preservation-matrix.md` gives each major subsystem a preservation
  classification.
- `harness-authority-map.md` gives every authority-required behavior a concrete
  owner prompt.
- `main-spec-conflict-list.md` records product-spec-adjacent risks without
  resolving them by inference.

## Remaining Source Limits

- Live Docker/testcontainers, Postgres, MinIO, Compose, browser, reset-route,
  and service-backed runtime behavior were not run.
- Cleanup on timeout, interrupt, parent death, active DB connections, detached
  reaper completion, stale janitor execution, and port release was not proven.
- Browser E2E runtime, Playwright failure artifacts, visual snapshot refresh,
  and direct package-script cleanup/output behavior were not runtime-observed.
- Environment precedence across Make, generated recipes, schedulers, scripts,
  package scripts, Playwright, config files, and Go helpers was not executed.
- Retained artifact freshness and newest-run fallback remain source-limited.

## Owner Questions Preserved

- Is `internal/app/test_runtime_reset.go` a harness-owned test hook, and what
  visibility/security boundary applies?
- Are direct package scripts first-class harness contracts or developer
  conveniences?
- Which local-dev Compose and `make dev` behaviors belong in the harness
  contract?
- What proof is sufficient before stale janitors may delete DBs, buckets, or
  containers?
- Which environment variables are public harness API, and what precedence
  applies?
- What platform/tool profile is supported?
- Are external Go caches under `/tmp/cartulary-go-*` in cleanup scope?
- Which visual snapshot platform/browser/update command is authoritative?
- Which generated artifacts are execution inputs but not owners?
