# FE-P3 Active Completion Handoff

## Status

FE-P3 is active and executable as of 2026-05-31. `tools/frontend_phase_registry.json`
now promotes the contiguous FE-P0 through FE-P3 chain to `active`; FE-P4 and
later remain `planned`.

This file is a handoff record only. Core 00 through Core 04 remain the current
product-conformance authority. `docs/testing-harness-nlspec.md` owns harness
mechanics. Frontend visual and accessibility rows remain `design_direction` or
`implementation_support` where mapped. Core 05 stayed inactive; no
claim-publication predicate was introduced.

## Completed Remediation

| Area | Result |
| --- | --- |
| Frontend phase governance | Added active frontend phase-slice planning and execution under `PHASE_NAMESPACE=frontend`; full FE-P3 slices schedule the target-owned row-accounting targets, and service-backed FE-P3 slices schedule browser-backed targets only. |
| Phase activation | FE-P0, FE-P1, FE-P2, and FE-P3 are active; FE-P4+ remain planned because future rows are still blocked. |
| FE-P0 evidence cleanup | Removed `make generate` from FE-P0 executable evidence; drift and generated-artifact checks remain validation evidence. |
| Browser helper behavior | FE-B-P3 now proves real browser route behavior for sort, filter, group, paste, fill-down, scroll-to-cell, group expand/collapse, and stable mutation anchors. Unsupported drag/tree helper claims were removed. |
| Grid adapter | Group rows expose a non-mutating `aria-expanded` outline toggle while preserving record-mutation gating and stable group labels. |
| Visual evidence | Refreshed `v-3-grid-03-grouped-grid-linux.png` for the intentional group outline affordance. Resize/fill handles remain design/support evidence, not Core product conformance. |
| Documentation and ledgers | Updated frontend maps, guide wording, visual guide wording, and regenerated frontend/base ledgers from the maps. |

## Current Evidence

| Command | Result |
| --- | --- |
| `make phase-slice PHASE_NAMESPACE=frontend PHASE=FE-P3` | pass, `.cartulary/test-results/20260531T140247Z-p2010616` |
| `make service-backed-slice PHASE_NAMESPACE=frontend PHASE=FE-P3` | pass, `.cartulary/test-results/20260531T140949Z-p2035166` |
| `make browser-e2e-support` | pass, retained in service-backed FE-P3 root `.cartulary/test-results/20260531T140949Z-p2035166/browser-e2e-support` |
| `make browser-e2e-webserver-backed` | pass, final full-check root `.cartulary/test-results/20260531T141811Z-p2060579/browser-e2e-webserver-backed` |
| `make browser-e2e-visual` | pass, final full-check root `.cartulary/test-results/20260531T141811Z-p2060579/browser-e2e-visual` |
| `make browser-e2e-a11y` | pass, final full-check root `.cartulary/test-results/20260531T141811Z-p2060579/browser-e2e-a11y` |
| `make phase-ledgers` | pass, `.cartulary/test-results/20260531T140225Z-p2008590` |
| `make phase-ledger-drift` | pass, `.cartulary/test-results/20260531T142813Z-p2117283` |
| `make phase-map-check` | pass, `.cartulary/test-results/20260531T142813Z-p2117290` |
| `make generated-artifact-policy-check` | pass, `.cartulary/test-results/20260531T142908Z-p2118786` |
| `make generate-drift` | pass, `.cartulary/test-results/20260531T142908Z-p2118789` |
| `make agent-finalize` | pass, `.cartulary/test-results/20260531T141709Z-p2058483`; generated maintenance updated one file, retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make check` | pass, `.cartulary/test-results/20260531T141811Z-p2060579`, 161/161 work units, 819 tests |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260531T141811Z-p2060579` | pass, `.cartulary/test-results/20260531T142602Z-p2112935`; retained-run maintenance passed, duration baselines refreshed |

The first `agent-finalize` updated
`tools/execution_topology_render_index.json`. The retained-run finalizer
updated `tools/browser_e2e_duration_baselines.json`,
`tools/go_test_duration_baselines.json`,
`tools/harness_smoke_duration_baselines.json`, `tools/scheduler_manifest.json`,
and `tools/service_backed_make_target_duration_baselines.json`.

## Row Closure Boundary

FE-P3 row closure is through target-owned
`cartulary.frontend_row_accounting.v1` artifacts. The active `phase-slice` root
closes all FE-P3 rows from the intended targets:

- `frontend-unit` closes `FE-U-P3-01` through `FE-U-P3-04` and `FE-I-P3-01`.
- `browser-e2e-support` and `browser-e2e-webserver-backed` close `FE-B-P3-01`.
- `browser-e2e-visual` closes `FE-V-P3-01` as `design_direction`.
- `browser-e2e-a11y` closes `FE-A11Y-P3-01` as `design_direction`.
- `frontend-import-boundary-check` and `lint-biome` close `FE-S-P3-01` as
  `implementation_support`.

Old retained artifacts from the planned-phase period are historical only and do
not support the active phase-completion claim.

## Carry Forward

- Keep frontend active phase execution target-owned; do not add row-local
  execution when existing target row accounting can close the row.
- Keep resize as visual/support evidence until an owner spec adopts product
  conformance semantics.
- Treat group expand/collapse as one-level grouping behavior, not tree rows.
- Preserve Core product, design-direction, implementation-support, and Core 05
  claim-publication boundaries in future maps, ledgers, and handoffs.
