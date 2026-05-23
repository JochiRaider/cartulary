# FE-P0 Frontend Coverage Ledger

This ledger is generated from `tools/frontend_phase_maps/fe_p0_test_map.json`. Update the frontend phase map first, then regenerate this file.

- Namespace: `frontend`
- Status: `planned`
- Owner refs: `docs/guides/cartulary_frontend_implementation_testing_guide.md`
- Depends on: `none`
- Authority: frontend phase maps are implementation-readiness inputs. This rendered ledger does not own product behavior.

## Rows

| Row | Layer | Evidence class | Claim status | Targets | Owner refs | Core REQs | Core ACs | Support/design ACs | Scenario titles | Claim | Out of scope |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `FE-U-P0-01` | `unit` | `product_conformance` | `blocked` | `make frontend-typecheck`<br>`make frontend-unit` | `Core 01 Sections 3.3.1, 7.4`<br>`development guide Sections 6.2, 6.6` | `REQ-01-019`, `REQ-01-020`, `REQ-01-022`, `REQ-01-034`, `REQ-01-307`, `REQ-01-311` | `AC-124`, `AC-125`, `AC-127`, `AC-184`, `AC-185`, `AC-231`, `AC-300`, `AC-303`, `AC-366`, `AC-368` | `none` | `none` | Verify generated protocol exports and frontend contract facades expose stable identifiers without hand-editing generated code. | This frontend readiness row does not change Core product behavior or Core 05 publication authority. |
| `FE-U-P0-02` | `unit` | `product_conformance` | `blocked` | `make frontend-unit` | `Core 01 Section 3.3.4`<br>`Core 03 Sections 4.1, 4.8` | `REQ-01-034`, `REQ-01-035`, `REQ-01-036`, `REQ-03-223`, `REQ-03-235` | `AC-124`, `AC-127`, `AC-184`, `AC-185`, `AC-231`, `AC-238`, `AC-243`, `AC-360`, `AC-363`, `AC-364`, `AC-366`, `AC-368`, `AC-372`, `AC-375` | `none` | `none` | Verify view-schema adapters key editable and queryable fields by field_key, not labels, indexes, or visible column order. | This frontend readiness row does not change Core product behavior or Core 05 publication authority. |
| `FE-U-P0-03` | `unit` | `product_conformance` | `blocked` | `make frontend-unit` | `Core 01 Section 7.4`<br>`Core 02 Section 5.3`<br>`development guide Section 6.4` | `REQ-01-307`, `REQ-01-311`, `REQ-02-222`, `REQ-02-223` | `AC-076`, `AC-084`, `AC-116`, `AC-122`, `AC-137`, `AC-145`, `AC-231`, `AC-252`, `AC-253`, `AC-277`, `AC-284`, `AC-287`, `AC-300`, `AC-303` | `none` | `none` | Verify stable selector and test-id builders derive identifiers from stable IDs and closed vocabularies rather than visible labels. | This frontend readiness row does not change Core product behavior or Core 05 publication authority. |
| `FE-S-P0-01` | `support` | `implementation_support` | `blocked` | `make generated-artifact-policy-check`<br>`make generate-drift` | `Core 00 Section 1`<br>`development guide Sections 2, 6.6, 7.1` | `none` | `none` | `FE-SUPPORT-AC-004` | `none` | Enforce generated protocol policy and detect generated contract drift. | This frontend readiness row does not change Core product behavior or Core 05 publication authority. |
| `FE-S-P0-02` | `support` | `implementation_support` | `blocked` | `make frontend-import-boundary-check` | `Development guide Sections 6.1, 6.3, 6.8, 6.10`<br>`implementation guide Section 14.8` | `none` | `none` | `FE-SUPPORT-AC-001`, `FE-SUPPORT-AC-002` | `none` | Enforce frontend import boundaries: /apps/web must consume /packages/grid-adapter and must not import react-data-grid directly. | This frontend readiness row does not change Core product behavior or Core 05 publication authority. |
| `FE-S-P0-03` | `support` | `implementation_support` | `blocked` | `make phase-ledger-drift` | `Implementation guide Sections 15 and 16` | `none` | `none` | `FE-GUIDE-AC-018` | `none` | Record frontend phase rows in a scheduler-enforced manifest or ledger before claiming phase-enforced completion. | This frontend readiness row does not change Core product behavior or Core 05 publication authority. |
