---
doc_id: THR-S8-HARNESS-AUTHORITY
title: Testing Harness Recovery Harness Authority Map
status: active
role: harness-authority-map
---

# Testing Harness Recovery Harness Authority Map

## Document role

This artifact records authority routing for S6 hazards and earlier S4/S8
follow-ups. It does not settle maintainer-owned decisions; it names them
precisely so later NLSpec, acceptance, and roadmap work do not silently convert
implementation behavior into intended contract.

## Authority map

| Authority ID | Surface | Current source of behavior | Proposed owner | May drive execution | Conflict precedence | Known conflicts or risks | Required decision | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| AUTH-0001 | Product behavior and public API semantics | Core 00 through Core 04 | product spec owners | yes for product conformance | Core owner sections govern product behavior. | Harness docs could accidentally redefine product behavior. | none for S6; preserve Core authority. | Core 00; `docs/domain.md` status | `observed` | Harness specs may define validation mechanics only. |
| AUTH-0002 | Harness command surface | Make/task-surface, S2 command map | harness spec owner | yes | adopted harness NLSpec, then recovery evidence, then implementation. | direct package scripts bypass Make result-root and scheduler policy. | none for S7; maintainer decided Make is the sole canonical harness command surface. | `entrypoint-command-map.md`; Make/task manifests; `RTR-0015`; `MD-S7-0001` | `observed/maintainer_decision` | Preserve Make canonical surface. |
| AUTH-0003 | Direct package scripts | root/app package manifests | developer convenience unless adopted later | no as canonical harness surface | Make remains canonical. | weaker output, cleanup, worker, env, and resource guarantees. | none for S7; direct package scripts are developer conveniences unless they re-enter Make-owned wrappers. | `EP-0016`; `ART-0030`; `HAZ-S4-0009`; `RTR-0015`; `MD-S7-0001` | `maintainer_decision` | Some scripts re-enter harness wrappers, but direct tool output remains separate. |
| AUTH-0004 | Test runtime reset route | `internal/testutil/testruntime/reset.go`, reset script, browser stack env | harness owner with product-security boundary preserved | conditional | Core product/security owns product state; harness owns the test-only hook only. | destructive test hook must not be mistaken for product API. | none for ownership; partial failure semantics remain source-limited. | `AMB-0006`; `SVC-0014`; `RES-0026`; `RTR-0014`; `MSC-0001`; `MD-S7-0004` | `maintainer_decision/source_limit` | Do not expose as product behavior. |
| AUTH-0005 | Local-dev services and Compose state | `docker-compose.dev.yml`, dev scripts, configs | local-dev/harness owner | yes for local verification | not product deployment authority. | fixed ports, persistent volumes, no observed compose-down target, object-store not reset by `db-reset`. | none for S7; local-dev Compose and `make dev` are local verification behavior, not deployment conformance. | `SVC-0008`; `SVC-0012`; `SVC-0015`; `RTR-0019`; `MD-S7-0005` | `maintainer_decision/source_observed` | Keep deployment claims out of the harness contract. |
| AUTH-0006 | Public environment-variable contracts and precedence | Make, schedulers, shell wrappers, Go helpers, Playwright, config files | harness config owner | yes when declared | Core 04 owns product deployment config keys. | unresolved precedence across Make, package scripts, schedulers, shell wrappers, Playwright, Go helpers, and config. | Which env vars are public harness API, and what precedence applies? | `ENV-0001` through `ENV-0024`; `AMB-0026`; `RTR-0015` | `maintainer_decision_required` | Avoid over-specifying incidental pass-through. |
| AUTH-0007 | Stale janitor destructive cleanup | `tools/testservices`, metadata, Docker labels, generated names | harness service owner within decided bounds | yes only with proof gates | destructive cleanup requires proven harness ownership. | source-only generated-name checks must be paired with metadata or lease proof and conservative age or completed-summary checks. | none for S7 proof rule; exact unobserved cleanup completion remains source-limited. | `AMB-0027`; `RTR-0009`; `HAZ-S4-0002`; `HAZ-S4-0004`; `MD-S7-0007` | `maintainer_decision/source_limit` | Runtime evidence may inform but cannot broaden destructive authority. |
| AUTH-0008 | Supported platform/tool profile | shell scripts, Docker/testcontainers, Compose, Playwright, toolchain baseline | harness/platform owner | yes when declared | repository procedure owns baseline tool versions; platform owner must define support profile. | missing-tool behavior is mixed for Docker, Compose, `ss`, `curl`, `setsid`, `realpath`, shell, localhost, Node/pnpm, browsers. | What host/tool/network profile is supported and what is unsupported? | `ENV-0024`; `AMB-0025`; `AMB-0028`; `RTR-0004`; `RTR-0011` | `maintainer_decision_required` | WSL/Linux was observed, not a platform matrix. |
| AUTH-0009 | External Go caches | Make and runner defaults | Go toolchain and local tooling | no by default | Go tool behavior owns cache consistency. | repo cleanup does not remove `/tmp/cartulary-go-*`. | none for S7; external Go caches are outside default `clean` and `distclean` scope. | `RES-0025`; `ART-0024`; `AMB-0021`; `RTR-0017`; `MD-S7-0009` | `maintainer_decision/source_observed` | Keep outside default cleanup unless a future owner expands scope. |
| AUTH-0010 | Generated task/schedule/code artifacts | generated outputs and generated-artifact policy | upstream manifest/spec/SQL owners; generated policy owner | yes as execution inputs | owner inputs and drift checks govern. | stale generated outputs can alter execution but must not become behavioral owners. | none unless generated policy changes. | `ART-0006` through `ART-0011`; `RTR-0002`; `MSC-0005` | `observed` | Clarify downstream authority in NLSpec. |
| AUTH-0011 | Scheduler lanes and resource claims | scheduler manifests, registry, scheduler code | harness scheduler owner | yes | harness spec may own scheduling policy; concrete host/service capacity needs separate evidence. | lanes could be mistaken for OS/Docker/Postgres/object-store capacity guarantees. | none for S7; lanes are logical scheduling constraints only, not concrete capacity guarantees. | `RES-0001` through `RES-0010`; `RTR-0001`; `MD-S7-0006` | `observed/maintainer_decision` | Preserve with clarification. |
| AUTH-0012 | Service lifecycle and cleanup guarantees | `tools/testservices`, testutil helpers, browser/dev scripts, S4/S6 maps | harness service owner | yes when documented | harness spec should own after adoption; implementation is evidence. | live readiness, timeout, interrupt, detached cleanup, and cleanup completion are not runtime-observed. | Which static lifecycle paths become contract after evidence, and which remain best-effort? | `SVC-*`; `CLN-*`; `RTR-0004`, `RTR-0007`, `RTR-0008`, `RTR-0012` | `observed/source_limit` | Runtime evidence remains future authorized work. |
| AUTH-0013 | Retained artifact run selection | artifact discovery, retained result roots, explain/fixture/drift/baseline tools | harness reporting owner | yes for investigation and drift/baseline flows | selected run identity governs durable claims. | newest-run fallback can select stale or unrelated local evidence. | none for S7; durable claims require explicit run identity and newest-run fallback is human-investigation only. | `ART-0013` through `ART-0017`; `RTR-0016`; `AMB-0017`; `MD-S7-0012` | `maintainer_decision/source_observed` | Human investigation fallback must print selected run identity if retained. |
| AUTH-0014 | Visual snapshot platform/update authority | committed Playwright PNG baselines and visual target map | browser/harness owner | yes for visual validation only | snapshot files are canonical validation evidence; refresh authority remains unresolved. | platform/browser drift and missing refresh command can make snapshot mutation ambiguous. | Which OS, browser version, and command may refresh visual snapshots? | `ART-0003`; `AMB-0022`; `RTR-0021`; `MSC-0007`; `MD-S7-0013` | `maintainer_decision_required/source_limit` | Validation-only targets must not update snapshots. |
| AUTH-0015 | CI provider workflow and annotations | `scripts/ci/**`; absent `.github/**` | CI/harness owner for provider-neutral entrypoint only | yes for `make ci`; no for absent provider workflow | provider-neutral `make ci` is recoverable; provider-specific workflow source is absent. | CI annotations and upload behavior are unavailable. | none for provider-neutral CI; provider-specific behavior remains source-limited. | `SL-0001`; `AMB-0001`; `FAIL-0028`; `MD-S7-0015` | `maintainer_decision/source_limit` | Do not invent provider behavior. |

## S7 authority clarification notes

The 2026-05-09 S7 evidence pass records these authority clarifications for
later NLSpec and acceptance-matrix work:

- Retained artifacts are strong evidence only when selected by explicit result
  directory and run ID. Ambient newest-run fallback is human-investigation only
  and cannot support normative claims.
- Generated artifacts are downstream execution inputs only. Upstream sources
  such as specs, manifests, SQL, Make/task definitions, and the adopted harness
  NLSpec own behavior. Drift checks prove freshness; generated files do not
  define behavior.
- `internal/gen/**` remains downstream of `contracts/**`, `db/queries/**`,
  generator source, and authored SQL inputs. Freshness checks include
  `make generate-drift`, `make generated-artifact-policy-check`, and
  `make json-shape-check`.
- `packages/protocol-ts/src/generated/**` remains downstream protocol code.
  Freshness checks include generated-artifact policy checks plus
  `make frontend-import-boundary-check` when import-boundary claims are made.
- Generated task and schedule manifests may drive execution when fresh, but
  concrete host/Docker/Postgres/object-store/browser capacity claims still need
  selected runtime evidence or maintainer decision.
- Duration baselines are planning weights only. The S7 pass ran drift readers
  against selected runs; no baseline refresh target was run.
