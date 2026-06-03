---
doc_id: THR-S8-PRESERVATION-MATRIX
title: Testing Harness Recovery Preservation Matrix
status: active
role: preservation-matrix
---

# Testing Harness Recovery Preservation Matrix

## Document Role

This artifact classifies harness subsystems for preservation, clarification,
exclusion, redesign, deprecation, compatibility-only handling, or owner review.
S6 refreshed this matrix from the S3 artifact/resource ownership matrix, S4
service/resource lifecycle map, and S5 lifecycle/failure registers.

The matrix is documentation-only. It does not change harness behavior and does
not close maintainer decisions by inference.

## Classification Rules

Allowed classifications:

- `preserve`: behavior is intentional enough for later NLSpec drafting.
- `preserve_with_clarification`: preserve the behavior but clarify limits,
  source status, or authority.
- `compatibility_only`: keep as compatibility support, not a preferred contract.
- `accidental`: observed behavior should not become a contract without redesign.
- `deprecated`: keep only as deprecated behavior if owner confirms.
- `redesign_required`: behavior is too risky or underspecified for preservation.
- `authority_required`: maintainer or owner decision is required before drafting.
- `exclude_from_contract`: document as out of contract or implementation detail.

## Preservation Matrix

| Behavior ID | Subsystem or behavior | Controlling evidence | Hazard/timing links | Main-spec dependency | Classification | Required owner question | Evidence status | S7/NLSpec handoff |
|---|---|---|---|---|---|---|---|---|
| PRES-0001 | Make/task-surface as canonical harness command surface. | S2 entrypoint map; `Makefile`; `tools/task_surface_manifest.json`; `tools/task_surface.generated.mk`; `MD-S7-0001`. | `RTR-0002`, `CONC-0001`, `CONC-0013`; `AUTH-0002`. | none beyond verification scope. | `preserve` | none for S7; Make is canonical. | `observed/maintainer_decision` | Draft Make/public target surface as canonical and state package scripts separately. |
| PRES-0002 | Generated task surface, schedules, ledgers, generated Go/TS, and generated Make fragments as execution inputs. | S3 `ART-0006` through `ART-0011`; `HAZ-S3-0003`, `HAZ-S3-0004`; generated artifact policy. | `RTR-0002`, `CONC-0013`; `AUTH-0010`; `MSC-0005`. | Core specs/contracts/SQL own upstream behavior. | `preserve_with_clarification` | Which generated artifacts are execution inputs but not behavior owners? | `observed` | Draft generated files as downstream inputs and require drift checks for freshness. |
| PRES-0003 | Check and service-backed scheduler lanes, resource claims, and work-unit dependencies. | S4 `RES-0001` through `RES-0010`; `HAZ-S4-0001`; scheduler manifests; `MD-S7-0006`. | `RTR-0001`, `TMR-0020`, `TMR-0021`, `CONC-0002`, `CONC-0003`; `AUTH-0011`. | none. | `preserve_with_clarification` | none for S7; lanes are logical scheduling constraints only. | `observed/maintainer_decision` | Draft logical concurrency limits separately from Docker/Postgres/object-store/browser capacity. |
| PRES-0004 | Go shard construction, shared report directories, and lock-protected report reuse. | S3 `HAZ-S3-0006`; S4 `RES-0020`, `RES-0021`; S5 failure rows for report/lock failures. | `RTR-0018`, `TMR-0018`, `TMR-0019`, `CONC-0004`. | none. | `preserve_with_clarification` | Are shared report lock deadlines part of the public harness contract or implementation detail? | `observed/source_limit` | Draft lock behavior as source-observed unless owner confirms public timeout semantics. |
| PRES-0005 | Suite service lifecycle for Docker/testcontainers, service leases, and managed containers. | S4 `SVC-0001`, `SVC-0007`, `RES-0012`, `RES-0013`; `HAZ-S4-0002`, `HAZ-S4-0003`. | `RTR-0004`, `RTR-0007`, `RTR-0008`; `TMR-0001`, `TMR-0002`, `TMR-0004`, `TMR-0017`, `TMR-0028`; `AUTH-0012`. | none unless cleanup touches product-like resources. | `preserve_with_clarification` | Which cleanup paths are guaranteed, best-effort, or source-limited? | `observed/source_limit` | Draft startup/readiness from source and keep live Docker readiness as source-limited. |
| PRES-0006 | Postgres templates, cloned DBs, package DB reuse, transactions, and reset tables. | S3 `HAZ-S3-0007`, `HAZ-S3-0008`; S4 `SVC-0002`, `SVC-0003`, `RES-0014`; S5 DB failure rows. | `RTR-0005`, `RTR-0010`; `TMR-0002`, `TMR-0003`, `TMR-0006`, `TMR-0022`; `AUTH-0006`, `AUTH-0007`, `AUTH-0012`. | Core 04 only if harness cleanup could affect non-harness DBs. | `preserve_with_clarification` | What DB naming and metadata proof is sufficient before cleanup or stale janitor deletion? | `observed/source_limit` | Draft template/clone/reset behavior with stale cleanup authority separated. |
| PRES-0007 | object-store buckets, package bucket reuse, browser buckets, and object cleanup. | S3 `HAZ-S3-0009`, `HAZ-S3-0010`; S4 `SVC-0004`, `SVC-0005`, `RES-0015`, `RES-0016`. | `RTR-0006`, `RTR-0010`; `TMR-0004`, `TMR-0005`, `TMR-0023`; `AUTH-0007`, `AUTH-0012`. | Core 04 if object-store config is product-facing. | `preserve_with_clarification` | What bucket/prefix proof is sufficient before stale object cleanup? | `observed/source_limit` | Draft bucket scope and cleanup as source-observed, not live-proven. |
| PRES-0008 | Browser E2E dynamic ports, process groups, webserver readiness, runtime roots, and port release waits. | S3 `HAZ-S3-0011`; S4 `SVC-0009`, `SVC-0010`, `RES-0017`, `RES-0018`; S5 browser failure rows. | `RTR-0011`, `RTR-0012`; `TMR-0008`, `TMR-0009`, `TMR-0010`, `TMR-0029`; `AUTH-0008`, `AUTH-0012`. | none unless backend is accidentally non-test product runtime. | `preserve_with_clarification` | Is `ss` required, and what is supported when configured-port checks are unavailable? | `observed/source_limit` | Draft dynamic-port behavior as best-effort allocation with platform caveats. |
| PRES-0009 | Playwright shared state, global setup lock, worker admin manifest, cleanup markers, and worker profiles. | S3 `HAZ-S3-0012`; S4 `SVC-0013`, `RES-0022`; S5 Playwright failure rows. | `RTR-0013`; `TMR-0011`, `TMR-0012`, `TMR-0024`, `TMR-0025`; `CONC-0010`; `AUTH-0008`, `AUTH-0012`. | none. | `preserve_with_clarification` | Which shared-state cleanup behaviors survive abrupt exit, and which are best-effort? | `observed/source_limit` | Draft setup lock/manifest semantics as source-observed and runtime-sensitive. |
| PRES-0010 | Browser reset boundaries and harness-owned test runtime reset route. | S3 `HAZ-S3-0014`; S4 `SVC-0014`, `RES-0026`; S5 reset failure rows; `AMB-0006`; `MD-S7-0004`. | `RTR-0014`, `TMR-0016`, `TMR-0030`; `AUTH-0004`; `MSC-0001`. | Core 01/Core 04 product runtime and security boundaries. | `preserve_with_clarification` | none for ownership; partial reset failure semantics remain source-limited. | `maintainer_decision/source_limit` | Do not draft as product API; reset hook is harness-owned only. |
| PRES-0011 | Direct root and app package scripts. | S3 `ART-0030`, `CLN-0020`, `HAZ-S3-0013`; S4 `HAZ-S4-0009`; S5 package-script failure/output rows; `MD-S7-0001`. | `RTR-0015`, `TMR-0032`, `CONC-0011`; `AUTH-0003`. | none. | `exclude_from_contract` | none for S7; package scripts are developer conveniences unless they re-enter Make wrappers. | `maintainer_decision` | Draft Make as canonical and package scripts separately. |
| PRES-0012 | Local-dev Compose services, `make dev`, fixed ports, persistent volumes, and object-store reset gap. | S3 cleanup matrix; S4 `SVC-0008`, `SVC-0012`, `SVC-0015`, `RES-0019`; `HAZ-S4-0008`; `MD-S7-0005`. | `RTR-0019`; `TMR-0013`, `TMR-0014`, `TMR-0015`; `CONC-0012`; `AUTH-0005`; `MSC-0003`. | Core 04 only by analogy; not deployment conformance. | `preserve_with_clarification` | none for S7; local-dev behavior is local verification, not deployment conformance. | `maintainer_decision/source_limit` | Keep local-dev persistence and reset gaps out of deployment rules. |
| PRES-0013 | Stale janitors for generated DBs, buckets, and managed containers. | S4 `SVC-0011`, `RES-0013`, `RES-0014`, `RES-0016`; `AMB-0019`, `AMB-0027`; `MD-S7-0007`. | `RTR-0007`, `RTR-0009`, `RTR-0010`; `TMR-0007`; `AUTH-0007`; `MSC-0008`. | Core 04/product data safety if bounds are wrong. | `preserve_with_clarification` | none for S7 proof gates; completion guarantees remain source-limited. | `maintainer_decision/source_limit` | Draft generated-name plus metadata/lease plus conservative age/completed-summary proof gates. |
| PRES-0014 | Public environment variables and cross-layer precedence. | S4 environment observations; `AMB-0012`, `AMB-0026`, `SL-0015`; `MD-S7-0010`. | `RTR-0015`, `RTR-0019`; `TMR-0031`, `TMR-0032`; `AUTH-0006`; `MSC-0004`. | Core 04 for product config keys. | `preserve_with_clarification` | Precedence remains owner-required. | `maintainer_decision/source_limit` | Draft known defaults/read sites only; mark precedence unresolved. |
| PRES-0015 | Supported platform/tool profile. | S4 `ENV-0024`; `AMB-0025`, `AMB-0028`; `SL-0013`; `MD-S7-0011`. | `RTR-0004`, `RTR-0011`, `RTR-0012`, `RTR-0019`; `AUTH-0008`. | none unless deployment support is implicated. | `preserve_with_clarification` | Full platform/tool support matrix remains owner-required. | `maintainer_decision/source_limit` | State WSL/Linux as observed primary environment while keeping Linux portability practical. |
| PRES-0016 | External Go caches under `/tmp/cartulary-go-*`. | S3 `ART-0024`, `HAZ-S3-0015`; `AMB-0021`; `MD-S7-0009`. | `RTR-0017`; `AUTH-0009`. | none. | `exclude_from_contract` | none for S7; external Go caches stay outside default cleanup. | `maintainer_decision/source_observed` | Keep tool-managed unless a future owner expands cleanup. |
| PRES-0017 | Retained `.cartulary/test-results/**` artifacts, run selection, newest-run fallback, and failure bundles. | S3 `HAZ-S3-0001`; S5 output/failure registers; `AMB-0017`; `SL-0004`, `SL-0009`, `SL-0010`; `MD-S7-0012`. | `RTR-0016`; `AUTH-0013`; `MSC-0009`. | none. | `preserve_with_clarification` | none for S7 run identity; failure-only bundle schemas remain source-limited. | `maintainer_decision/source_limit` | Draft explicit `RESULTS_DIR`/run-id provenance; keep newest fallback diagnostic-only. |
| PRES-0018 | Browser visual snapshot baselines and update authority. | S3 `AMB-0015`, `AMB-0022`; Playwright visual snapshots and configs; `MD-S7-0013`. | `RTR-0021`; `AUTH-0014`; `MSC-0007`. | Product UI conformance only through test assertions. | `preserve_with_clarification` | Visual snapshot refresh OS/browser/version/command remains owner-required. | `maintainer_decision_required/source_limit` | Visual snapshots are validation-only until owner supplies exact refresh bounds. |
| PRES-0019 | Provider-neutral `make ci` and absent `.github/**` workflows. | `SL-0001`; S2 command map; `scripts/ci/**`; `MD-S7-0015`. | `AUTH-0015`; `MSC-0010`. | release/deployment process only if owner links it. | `preserve_with_clarification` | Provider workflow behavior remains source-limited. | `maintainer_decision/source_limit` | Draft repo-local CI entrypoint; keep provider workflow behavior out of scope. |
| PRES-0020 | Shell scratch directories and temporary roots used by harness scripts. | S3 cleanup matrix; shell lifecycle scripts. | `RTR-0020`, `TMR-0026`, `TMR-0027`. | none. | `preserve_with_clarification` | Which scratch dirs are stable interfaces versus local implementation detail? | `observed/source_limit` | Draft cleanup as best-effort where traps/source support it; avoid public path guarantees unless needed. |

## Coverage Check

- Command surface: `PRES-0001`.
- Generated artifacts: `PRES-0002`.
- Schedulers and lanes: `PRES-0003`.
- Go sharding/report locks: `PRES-0004`.
- Suite services and Docker/testcontainers: `PRES-0005`.
- Postgres resources: `PRES-0006`.
- object-store resources: `PRES-0007`.
- Browser stack/processes/ports: `PRES-0008`.
- Playwright state/workers: `PRES-0009`.
- Reset route: `PRES-0010`.
- Package scripts: `PRES-0011`.
- Local-dev services: `PRES-0012`.
- Stale janitors: `PRES-0013`.
- Environment contracts: `PRES-0014`.
- Platform/tools: `PRES-0015`.
- External Go caches: `PRES-0016`.
- Retained artifacts: `PRES-0017`.
- Visual snapshots: `PRES-0018`.
- CI provider gap: `PRES-0019`.
- Scratch dirs: `PRES-0020`.

Every `authority_required` row has a concrete owner question. Runtime-sensitive
claims remain `source_limit` unless a controlling input already marked runtime
evidence.
