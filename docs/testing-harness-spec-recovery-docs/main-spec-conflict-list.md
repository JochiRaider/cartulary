---
doc_id: THR-S8-MAIN-SPEC-CONFLICTS
title: Testing Harness Recovery Main Spec Conflict List
status: active
role: main-spec-conflict-list
---

# Testing Harness Recovery Main Spec Conflict List

## Document role

This S8 artifact lists S4 follow-up surfaces that could conflict with Core 00
through Core 04 if misclassified. No confirmed product-spec contradiction was
closed in this pass.

## Conflict list

| Conflict ID | Surface | Potential conflict | Current handling | Required owner path | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|
| MSC-0001 | Test runtime reset route. | A destructive test-only route in `internal/app` could be mistaken for public product API or production behavior. | Keep as authority-required harness/test hook; do not make normative product behavior. | S8 maintainer decision with product/security owner input. | `internal/app/test_runtime_reset.go`; `AMB-0006`; `HAZ-S4-0007` | `observed/source_limit` | Core product routes remain owned by Core 01/Core 04. |
| MSC-0002 | Local test credentials and Compose/dev defaults. | Local credentials could be mistaken for production secret policy. | S4 records local test/dev credentials only. | S8 platform/local-dev decision. | `ENV-0010`, `ENV-0012`, `ENV-0020`; S4 audit | `observed/source_limit` | Core 04 security policy remains separate. |
| MSC-0003 | Local-dev Compose topology. | Compose dev services could be mistaken for deployment topology or conformance requirement. | Keep local-dev scope separate from verification/product conformance. | S8 local-dev authority decision. | `SVC-0008`, `SVC-0015`, `HAZ-S4-0008` | `observed/source_limit` | Core 01/Core 04 own product/deployment topology. |
| MSC-0004 | Public env contracts versus deployment config. | Harness env precedence could conflict with product deployment config ownership. | Mark precedence unknown and route to S8. | S8 harness config owner plus Core 04 owner for product config keys. | `ENV-0013`, `ENV-0020`, `AMB-0012`, `AMB-0026` | `observed/source_limit` | Do not use S4 env rows as full public contract. |
| MSC-0005 | Generated artifacts and generated Make/schedule files. | Generated downstream files could be treated as behavioral owners. | Preserve generated-as-input, not owner. | NLSpec authority language; generated policy owner. | `ART-0008`, `ART-0010`, `HAZ-S3-0003`, `HAZ-S3-0004` | `observed` | Upstream manifests/contracts own behavior. |
| MSC-0006 | `docs/domain.md` vocabulary context. | Domain reference could be used to add runtime behavior. | Use vocabulary only; Core owner sections govern. | none unless drift found. | `docs/domain.md`; Core 00 | `observed` | No drift claim made. |
| MSC-0007 | Browser visual/platform behavior. | Platform-specific browser artifacts could be mistaken for product UI conformance beyond harness validation. | Keep visual/platform update authority open. | S8/browser owner decision. | `AMB-0022`; `ENV-0024`; Playwright configs | `observed/source_limit` | No snapshot update command executed. |

## Conflict summary

No confirmed Core 00 through Core 04 contradiction was resolved. All listed
items remain routing concerns for owner review or later NLSpec authority text.

