---
doc_id: THR-S8-HARNESS-AUTHORITY
title: Testing Harness Recovery Harness Authority Map
status: active
role: harness-authority-map
---

# Testing Harness Recovery Harness Authority Map

## Document role

This S8 artifact records authority routing for S4 audit follow-ups. It does not
settle maintainer-owned decisions; it names them precisely.

## Authority map

| Surface | Current source of behavior | Proposed owner | May drive execution | Conflict precedence | Known conflicts | Required decision | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|
| Product behavior and public API semantics | Core 00 through Core 04. | product spec owners | yes for product conformance | Core owner sections govern. | none newly confirmed. | none for S4 follow-up. | Core 00; `docs/domain.md` status | `observed` | Harness docs must not redefine product behavior. |
| Harness command surface | Make/task-surface, S2 command map. | harness spec owner | yes | adopted harness NLSpec, then recovery evidence, then implementation. | direct package scripts bypass Make. | Confirm Make remains canonical. | `entrypoint-command-map.md`; Make files | `observed` | S8 default: preserve Make canonical surface. |
| Service lifecycle | `tools/testservices`, testutil helpers, browser/dev scripts, S4 maps. | harness service owner | yes, when documented. | harness spec should own once adopted; implementation is current evidence. | live readiness not observed. | Decide which static lifecycle paths become contract after evidence. | `service-lifecycle-map.md` | `observed/source_limit` | Runtime evidence remains S6. |
| Reset route | `internal/app/test_runtime_reset.go`, reset script, browser stack env. | maintainer/harness owner with product-security input | conditional | Core product/security owns product state; harness may own test-only hook only if adopted. | product package contains destructive test hook. | Visibility/security boundary and contract status. | `AMB-0006`; `HAZ-S4-0007` | `maintainer_decision_required` | Do not expose as product API. |
| Package scripts | root/app package manifests. | harness/tooling owner if adopted | currently yes as developer commands, not Make contract | Make remains canonical until decision. | weaker output/result/cleanup guarantees. | First-class contract or convenience aliases. | `AMB-0020`; `HAZ-S4-0009` | `maintainer_decision_required` | Some scripts re-enter harness wrappers. |
| Local-dev services | `docker-compose.dev.yml`, dev scripts, configs. | local-dev/harness owner if adopted | yes for local dev | not product deployment authority. | fixed ports, persistent volumes, MinIO not reset by `db-reset`. | Contract status and cleanup boundaries. | `SVC-0008`, `SVC-0015`, `HAZ-S4-0008` | `maintainer_decision_required` | Keep separate from verification gates. |
| Stale janitor destructive cleanup | `tools/testservices`, metadata/labels, generated names. | harness service owner plus maintainer decision | yes only if adopted with bounds | destructive cleanup must require proven ownership. | source-only generated ownership proof. | Deletion bounds and sufficient evidence. | `AMB-0027`; `RTR-0006` | `maintainer_decision_required` | S6 runtime evidence may inform but not replace authority. |
| Public env contracts | Make, schedulers, shell wrappers, Go helpers, Playwright, config. | harness config owner | yes when declared | Core 04 owns product deployment config keys. | unresolved precedence across layers. | public env set and precedence matrix. | `ENV-0001` through `ENV-0024`; `AMB-0026` | `maintainer_decision_required` | Avoid over-specifying incidental pass-through. |
| Supported platform/tool profile | shell scripts, Docker/testcontainers, Compose, Playwright, toolchain baseline. | harness/platform owner | yes when declared | repository procedure owns baseline tool versions; S8 owns support statement. | missing-tool behavior mixed. | supported OS/shell/tool/network profile. | `ENV-0024`; `AMB-0028` | `maintainer_decision_required` | WSL/Linux was observed, not a full matrix. |
| External Go caches | Make/runner defaults. | harness/tooling owner | no, unless adopted | Go tool behavior owns cache consistency. | repo cleanup does not remove them. | in cleanup contract or tool-managed outside scope. | `RES-0025`; `AMB-0021` | `maintainer_decision_required` | Proposed default: outside cleanup contract. |
| Generated task/schedule artifacts | generated outputs and policy file. | upstream manifest/spec owners; generated policy owner | yes as execution inputs | owner inputs and drift checks govern. | stale generated outputs can alter execution. | none unless policy changes. | `ART-0008`, `ART-0010`, `HAZ-S3-0003`, `HAZ-S3-0004` | `observed` | Clarify downstream authority in NLSpec. |

