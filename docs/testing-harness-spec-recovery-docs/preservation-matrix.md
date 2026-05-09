---
doc_id: THR-S8-PRESERVATION-MATRIX
title: Testing Harness Recovery Preservation Matrix
status: active
role: preservation-matrix
---

# Testing Harness Recovery Preservation Matrix

## Document role

This S8 artifact classifies S4 follow-up behavior as preserve, clarify,
authority-required, or exclude from contract. Maintainer decisions are still
required where noted.

## Preservation matrix

| Behavior ID | Behavior | Current evidence | Main-spec dependency | External dependency | Failure cost | Classification | Required decision | Roadmap target | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| PRES-0001 | Make/task-surface as canonical harness command surface. | S2 command map; Make help/task manifests. | none beyond product validation scope. | Make, shell, Go/Node tools. | high | `preserve_with_clarification` | Confirm direct package scripts remain secondary unless explicitly adopted. | NLSpec command surface | `observed` | S4 follow-ups rely on Make-owned result roots and scheduler policy. |
| PRES-0002 | S4 source-limit preservation for live readiness, runtime, cleanup, platform, and env precedence. | `SL-0012` through `SL-0015`; S4 audit. | none. | Docker, browsers, platform tools. | high | `preserve` | none; keep limits open until evidence. | NLSpec evidence discipline | `observed/source_limit` | Do not rewrite static findings as runtime proof. |
| PRES-0003 | App test runtime reset route. | `SVC-0014`, `RES-0026`, `HAZ-S4-0007`, `AMB-0006`. | Core 01/Core 04 product state and deployment trust boundaries. | running test-enabled backend, Postgres, object store. | high | `authority_decision_required` | Is reset route a harness-owned test hook, and what visibility/security boundary applies? | S8 owner review | `observed/source_limit` | Do not make public product behavior. |
| PRES-0004 | Direct package scripts. | `EP-0016`, `ART-0030`, `HAZ-S4-0009`, `AMB-0020`. | none. | pnpm, Vite, Vitest, Biome, Playwright. | medium | `authority_decision_required` | Are direct package scripts first-class harness contracts or developer conveniences? | S8 owner review | `observed` | Default is Make canonical. |
| PRES-0005 | Local-dev Compose services and `make dev`. | `SVC-0008`, `SVC-0012`, `SVC-0015`, `HAZ-S4-0008`. | Core 04 deployment config only by analogy; not product conformance. | Docker Compose, fixed host ports, local volumes. | medium | `authority_decision_required` | Which local-dev service behavior belongs in harness contract? | S8 owner review | `observed/source_limit` | Persistent volumes and no compose-down target remain source-limited. |
| PRES-0006 | Stale janitor destructive cleanup. | `SVC-0011`, `RES-0013`, `RES-0014`, `RES-0016`, `AMB-0027`. | product DB/object state must not be destroyed outside harness ownership. | Docker/Postgres/MinIO. | high | `authority_decision_required` | What proof is sufficient before deleting stale DBs, buckets, or containers? | S8 owner review and S6 runtime evidence | `observed/source_limit` | Keep generated-name/metadata bounds narrow. |
| PRES-0007 | Public environment-variable contracts and precedence. | `ENV-0001` through `ENV-0024`; `AMB-0012`, `AMB-0026`, `SL-0015`. | Core 04 owns product deployment config where applicable. | Make, schedulers, shell wrappers, package scripts, Go helpers, config files, Playwright. | medium | `authority_decision_required` | Which env vars are public harness API, and what precedence applies? | S8 owner review | `observed/source_limit` | Optional runtime matrix requires separate authorization. |
| PRES-0008 | Supported platform/tool profile. | `ENV-0024`; `AMB-0025`, `AMB-0028`; `SL-0013`. | none unless product deployment support is implicated. | Docker, Compose, `ss`, `curl`, `setsid`, `realpath`, bash, localhost, Node/pnpm, browser runtime. | high | `authority_decision_required` | What host/tool profile is supported, and what behavior is unsupported? | S8 owner review | `observed/source_limit` | Missing-tool behavior is mixed by source. |
| PRES-0009 | Scheduler resource lanes. | `RES-0001` through `RES-0010`; `HAZ-S4-0001`; `RTR-0001`. | none. | host and service capacity. | high | `preserve_with_clarification` | State lanes are scheduling limits, not concrete capacity guarantees. | NLSpec resource rules | `observed/source_limit` | Runtime capacity evidence is separate. |
| PRES-0010 | External Go caches. | `RES-0025`, `ART-0024`, `AMB-0021`. | none. | Go toolchain caches under `/tmp`. | low | `authority_decision_required` | Are external Go caches in cleanup contract? | S8 owner review | `observed/source_limit` | Proposed default: tool-managed, outside `clean`/`distclean`. |
| PRES-0011 | Generated artifacts as execution inputs but not behavior owners. | `ART-0006` through `ART-0011`; S3 hazards. | Core specs/contracts/SQL own upstream behavior. | generators and drift checks. | high | `preserve_with_clarification` | none unless owner changes generated policy. | NLSpec authority language | `observed` | Avoid treating generated Make/schedules as normative owners. |

