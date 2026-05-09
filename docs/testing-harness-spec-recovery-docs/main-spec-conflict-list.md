---
doc_id: THR-S8-MAIN-SPEC-CONFLICTS
title: Testing Harness Recovery Main Spec Conflict List
status: active
role: main-spec-conflict-list
---

# Testing Harness Recovery Main Spec Conflict List

## Document Role

This artifact lists S6 resource, timing, cleanup, reset, local-dev, generated
artifact, and platform surfaces that could conflict with Core 00 through Core
04 if later NLSpec drafting misclassifies them. No product-spec contradiction
is resolved here.

## Conflict List

| Conflict ID | Surface | Potential conflict | Current handling | Required owner path | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|
| MSC-0001 | Test runtime reset route. | A destructive test-only route in `internal/app` could be mistaken for public product API or production behavior. | Keep as `authority_required` harness/test hook; do not make normative product behavior. | Maintainer decision with product/security owner input. | `internal/app/test_runtime_reset.go`; `AMB-0006`; `HAZ-S4-0007`; `RTR-0014`; `AUTH-0004`; `PRES-0010` | `maintainer_decision_required/source_limit` | Core product routes remain owned by Core 01/Core 04. |
| MSC-0002 | Local test credentials and Compose/dev defaults. | Local credentials could be mistaken for production secret policy. | S4 records local test/dev credentials only. | Harness/local-dev owner plus Core 04 owner where product config keys overlap. | `ENV-0010`; `ENV-0012`; `ENV-0020`; S4 audit; `AUTH-0005`; `AUTH-0006` | `observed/source_limit` | Core 04 security policy remains separate. |
| MSC-0003 | Local-dev Compose topology and `make dev`. | Compose dev services could be mistaken for deployment topology or conformance requirement. | Keep local-dev scope separate from verification/product conformance unless owner adopts it. | Local-dev authority decision. | `SVC-0008`; `SVC-0015`; `HAZ-S4-0008`; `RTR-0019`; `PRES-0012`; `AUTH-0005` | `maintainer_decision_required/source_limit` | Core 01/Core 04 own product/deployment topology. |
| MSC-0004 | Public env contracts versus deployment config. | Harness env precedence could conflict with product deployment config ownership. | Mark precedence unknown and route owner questions. | Harness config owner plus Core 04 owner for product config keys. | `ENV-0013`; `ENV-0020`; `AMB-0012`; `AMB-0026`; `SL-0015`; `PRES-0014`; `AUTH-0006` | `maintainer_decision_required/source_limit` | Do not use S4 env rows as a complete public contract. |
| MSC-0005 | Generated artifacts and generated Make/schedule files. | Generated downstream files could be treated as behavioral owners. | Preserve generated-as-input, not owner. | NLSpec authority language and generated policy owner. | `ART-0008`; `ART-0010`; `HAZ-S3-0003`; `HAZ-S3-0004`; `RTR-0002`; `PRES-0002`; `AUTH-0010` | `observed` | Upstream manifests/contracts/specs own behavior. |
| MSC-0006 | `docs/domain.md` vocabulary context. | Domain reference could be used to add runtime behavior. | Use vocabulary only; Core owner sections govern. | none unless drift found. | `docs/domain.md`; Core 00 | `observed` | Domain docs inform terms and concepts only. |
| MSC-0007 | Browser visual/platform behavior. | Platform-specific browser artifacts could be mistaken for product UI conformance beyond harness validation. | Keep visual platform/update authority open. | Browser/harness owner decision. | `AMB-0022`; `ENV-0024`; Playwright configs; `RTR-0021`; `PRES-0018`; `AUTH-0014` | `maintainer_decision_required/source_limit` | No snapshot update command was executed. |
| MSC-0008 | Stale janitor destructive cleanup. | Generated fixture cleanup could delete DBs, buckets, or containers that are not harness-owned if proof rules are wrong. | Keep janitor destructive bounds source-limited and authority-required. | Harness service owner plus product data safety review where non-harness resources might be implicated. | `SVC-0011`; `RES-0013`; `RES-0014`; `RES-0016`; `AMB-0019`; `AMB-0027`; `RTR-0009`; `PRES-0013`; `AUTH-0007` | `maintainer_decision_required/source_limit` | Later specs must not broaden cleanup beyond generated-name/metadata evidence. |
| MSC-0009 | Retained run artifacts and benchmark/baseline claims. | Stale local artifacts or contaminated timing could be mistaken for current product or harness performance evidence. | Require explicit provenance for normative claims; newest fallback remains diagnostic unless owner adopts it. | Harness reporting/benchmark owner. | `HAZ-S3-0001`; `HAZ-S3-0005`; `AMB-0017`; `RTR-0003`; `RTR-0016`; `PRES-0017`; `AUTH-0013` | `observed/source_limit` | S6 did not validate retained artifact freshness or refresh baselines. |
| MSC-0010 | CI provider workflow and annotations. | Provider-specific workflow behavior could be invented despite absent `.github/**`. | Keep repo-local `make ci`/`scripts/ci/**` separate from provider workflows. | CI owner decision if provider contract is required. | `SL-0001`; `AMB-0001`; `PRES-0019`; `AUTH-0015` | `source_limit` | No workflow upload, annotation, or matrix behavior is recoverable from absent files. |

## Conflict Summary

No confirmed Core 00 through Core 04 contradiction was resolved. S6 only routes
product-spec-adjacent risks for later owner review and NLSpec authority text.
