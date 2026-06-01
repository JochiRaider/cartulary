# Frontend Phase 4 Implementation Plan

## Summary

Frontend verification contract. This file is the execution roadmap, progress marker, and FE-P5 handoff aid for `FE-P4: Timeline Hot Path And Sync Engine`.

Frontend verification contract. This plan is not behavior authority and MUST NOT mark any FE-P4 row complete without direct current row-owned evidence from the mapped target and `cartulary.frontend_row_accounting.v2` where the harness requires it.

Source limit. This task creates planning documentation only. It does not implement FE-P4 product behavior, does not activate FE-P4, does not edit generated ledgers, and does not introduce a Core 05 claim-publication predicate.

## Authority Model

- Product-conformance behavior remains owned by Core 00 through Core 04 under `docs/spec/`.
- Core 05 remains claim-publication-only. It is inactive for ordinary FE-P4 engineering, visual, accessibility, or readiness evidence unless a claim-bearing timed, benchmark, or fixture-sensitive publication predicate is explicitly introduced.
- `docs/testing-harness-nlspec.md` owns harness mechanics only: public Make targets, frontend row accounting, generated ledgers, accessibility summaries, visual fixture validation, phase selection, retained artifacts, cleanup, and failure normalization.
- `docs/domain.md` is the vocabulary and concept-boundary reference. For FE-P4, this is especially relevant to Timeline, rough capture, workbook surface, `record_id`, `field_key`, `view_schema_id`, projection versus source state, mention versus entity, and evidence versus blob boundaries.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md` owns the frontend phase shape, FE-P4 row set, row accounting, and frontend completion rules. It does not replace Core behavior authority.
- `docs/guides/cartulary_implementation_testing_guide.md` supplies sequencing, row discipline, shared harness expectations, completion checklist, and coverage-ledger expectations.
- `docs/guides/cartulary-dev-guide.md` supplies repo-local frontend package boundaries, generated-artifact policy, Make task-surface expectations, workspace shape, and implementation baseline.
- `docs/guides/cartulary-ui-ux-design-guide.md`, `docs/design.md`, and `docs/guides/cartulary_visual_golden_maintenance.md` are design-direction, visual, accessibility, and reviewer-discipline context only.
- Research reports remain rationale only.
- Product-conformance, design-direction, implementation-support, and claim-publication-boundary evidence classes MUST remain separate in maps, ledgers, summaries, plans, and handoff text.

## Current Repo Status

Locally verified facts as of 2026-06-01:

- `FRONTEND_PHASE4_IMPLEMENTATION_PLAN.md` did not exist before this plan was created.
- `tools/frontend_phase_registry.json` has `schema_id="cartulary.frontend_phase_registry.v2"`, `schema_version=2`, `phase_namespace="frontend"`, and `guide_path="docs/guides/cartulary_frontend_implementation_testing_guide.md"`.
- FE-P0 through FE-P3 are currently registry-active with `row_rollup_state="active_green"` and no activation blockers.
- `FRONTEND_PHASE3_IMPLEMENTATION_PLAN.md` is an active-completion handoff. It records FE-P3 as active and executable as of 2026-05-31 and states FE-P4 and later remain planned. Its historical run roots are handoff context only unless rerun.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P4` passed and reported FE-P4 as planned, explainable, and non-executable.
- `make help` passed and points to `make help-all` for the exhaustive public target catalog.
- `make explain-target` confirmed local target availability for FE-P4 direct or conditional targets: `frontend-unit`, `browser-e2e-webserver-backed`, `browser-e2e-stateful`, `browser-e2e-visual`, `browser-e2e-a11y`, `browser-e2e-a11y-preflight`, `browser-e2e-support`, and `phase-ledger-drift`.
- `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary` reports `phase_coverage: none`; this preflight target is blocked-row smoke and does not close implemented accessibility rows.
- `make explain-target TARGET=browser-e2e-a11y DETAIL=summary` also reports `phase_coverage: none` in the current task surface. FE-P4 accessibility closure therefore remains blocked until the map, target, and normalized accessibility summary account for an implemented FE-A11Y row.
- `make phase-ledger-drift` passed after this plan was written with run root `.cartulary/test-results/20260601T210509Z-p23551` and summary `phase-ledger-drift/tool-run-summary.json`.
- Registry tuple validation passed with `jq -e`.
- FE-P4 row-inventory validation passed with `jq -e`.

FE-P4 registry tuple:

| Field | Current value |
| --- | --- |
| Phase ID | `FE-P4` |
| Status | `planned` |
| Row rollup state | `no_rows_implemented` |
| Manifest path | `tools/frontend_phase_maps/fe_p4_test_map.json` |
| Ledger path | `docs/testing/frontend_phase_coverage_ledgers/fe_p4_coverage_ledger.md` |
| Depends on | `FE-P0`, `FE-P1`, `FE-P2`, `FE-P3` |
| Activation blocker | `FE-P4-ACTIVATION-BLOCKER-01`, `reason_code="frontend_phase_not_active"` |
| Manifest digest | `dc056454fb03e6a172346cb6890c30509ae4ce2a7fa8f9bb424294e0fe99d298` |
| Ledger digest | `bc6990dc9d18dc5dccc4251ad0fe909633a1f16bd843bb21f5401d852944e42e` |
| Evidence freshness digest | `7bb17dd7fd1d5bbfa1423571b3c85b3ac152aedaf49bb8eeef8fc462daf61857` |

FE-P0 through FE-P4 current activation chain:

| Phase | Status | Row rollup state | Activation blockers | Depends on |
| --- | --- | --- | --- | --- |
| `FE-P0` | `active` | `active_green` | none | none |
| `FE-P1` | `active` | `active_green` | none | `FE-P0` |
| `FE-P2` | `active` | `active_green` | none | `FE-P0`, `FE-P1` |
| `FE-P3` | `active` | `active_green` | none | `FE-P0`, `FE-P1`, `FE-P2` |
| `FE-P4` | `planned` | `no_rows_implemented` | `frontend_phase_not_active` | `FE-P0`, `FE-P1`, `FE-P2`, `FE-P3` |

FE-P4 row inventory:

| Row | Layer | Evidence class | Claim status | Target(s) | Scenario titles | Current blocker |
| --- | --- | --- | --- | --- | --- | --- |
| `FE-U-P4-01` | `unit` | `product_conformance` | `blocked` | `make frontend-unit` | none | `frontend_phase_row_not_implemented` |
| `FE-U-P4-02` | `unit` | `product_conformance` | `blocked` | `make frontend-unit` | none | `frontend_phase_row_not_implemented` |
| `FE-I-P4-01` | `integration` | `product_conformance` | `blocked` | `make frontend-unit`; `make browser-e2e-webserver-backed` | exact row title present | `frontend_phase_row_not_implemented` |
| `FE-E-P4-01` | `e2e` | `product_conformance` | `blocked` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | exact row title present | `frontend_phase_row_not_implemented` |
| `FE-V-P4-01` | `visual` | `design_direction` | `blocked` | `make browser-e2e-visual` | exact row title present | `visual_fixture_not_recaptured_for_frontend_row` |
| `FE-A11Y-P4-01` | `accessibility` | `design_direction` | `blocked` | `make browser-e2e-a11y-preflight` | exact row title present | `frontend_phase_row_not_implemented` |

FE-P4 visual fixture registry facts:

| Fixture ID | Title | Status | FE-P4 row link | Scenario title in registry | Golden filename |
| --- | --- | --- | --- | --- | --- |
| `FE-VFIX-08` | Save-state strip | `current` | `FE-V-P4-01` | `V-3-GRID-02 captures Timeline edit save-state visuals for active cell syncing saved and conflict states` | `apps/web/e2e/workbook.visual.spec.ts-snapshots/v-3-grid-02-saved-strip-linux.png` |
| `FE-VFIX-12` | Edit cell | `current` | `FE-V-P4-01` | `FE-V-P3-01 Capture frozen column, resize handle, fill-down handle, edit cell, group outline row, and empty successful query grid-adapter fixtures.` | `apps/web/e2e/workbook.visual.spec.ts-snapshots/fe-v-p3-01-grid-adapter-fixtures-linux.png` |
| `FE-VFIX-15` | Empty successful query | `current` | `FE-V-P4-01` | `FE-V-P3-01 Capture frozen column, resize handle, fill-down handle, edit cell, group outline row, and empty successful query grid-adapter fixtures.` | `apps/web/e2e/workbook.visual.spec.ts-snapshots/fe-v-p3-01-grid-adapter-fixtures-linux.png` |

Source limits:

- The generated FE-P4 ledger was inspected but not edited. It is downstream of `tools/frontend_phase_maps/fe_p4_test_map.json` and is not behavior authority.
- Current `current` fixture status does not close `FE-V-P4-01`; row-owned frontend visual accounting is still blocked.
- No FE-P4 row-owned target was run in this planning task. Existing latest artifacts reported by `make explain-target` are historical diagnostic context only.
- The current FE-P4 accessibility row is preflight-only in the map. Preflight evidence cannot complete an implemented accessibility row under the current frontend guide and testing harness NLSpec.
- The plan does not inspect every implementation module under `/apps/web`; exact local module filenames must be rechecked in the owning sprint before product edits.

## Phase Objective

Core restatement. FE-P4 must turn the existing workbook shell and grid adapter into the first Timeline mutation hot path through public route contracts.

At FE-P4 exit, a browser user must be able to:

- query the Timeline and render full `view_row_v1` cells by stable row and field identity;
- create a low-friction rough Timeline row without requiring entity mention resolution, evidence handles, or canonical enrichment first;
- edit cells inline;
- paste tabular data into the Timeline hot path;
- see queued, saved, failed, and conflict save-state feedback on the same workbook surface;
- recover from transient write failure through deterministic pending replay;
- see cell-level validation and public error-envelope feedback without private server details;
- refresh without losing stable row identity or retargeting pending work by visible row order.

FE-P4 completion requires current row-owned evidence for all intended FE-P4 rows plus satisfaction or precise blocking for every triggered shared harness.

## Implementation Scope

### In scope

- `/apps/web` sync engine and workbook controllers for Timeline query state, rough creation, inline edit, paste, save-state derivation, pending queue, replay, validation display, and public mutation submission.
- `/packages/grid-adapter` where FE-P4 needs stable `record_id`, `field_key`, edit, paste, focus, and adapter callback integration. Any adapter lifecycle or helper change must preserve FE-P3 guarantees.
- `/packages/view-contracts` for contract-derived view schema, field capabilities, writable fields, and `view_row_v1` interpretation.
- `/packages/protocol-ts` for generated protocol type consumption through handwritten facades. Generated paths remain downstream and must not be hand-edited.
- `/packages/ui-contracts` for stable selector and test-id builders if FE-P4 introduces new selectors that cross runtime, unit, browser, support-test, helper, or option-object boundaries.
- `/packages/test-utils` for browser command helpers if FE-P4 paste, edit, replay, or row-anchor helpers need shared choreography.
- FE-P4 unit, integration, browser E2E, visual, accessibility-preflight, and conditional support validation.
- Authored FE-P4 phase metadata updates only when direct row-owned evidence exists. Metadata changes require generated ledger regeneration through Make targets.

### Out of scope

- Entity mention resolution.
- Evidence handles.
- WebSocket live updates.
- Same-field conflict resolver implementation.
- Saved-view persistence.
- Claim-bearing publication work unless a claim-publication predicate is explicitly introduced.
- New public API shapes that are not already owner-backed by Core 00 through Core 04 or an adopted NLSpec.
- Hand edits to generated roots, generated ledgers, generated schedules, `pnpm-lock.yaml`, `go.sum`, or tool-managed artifacts.
- Adding FE-P4 rows to base `tools/phase*_test_map.json`; frontend rows stay in the frontend namespace.

## Evidence Layer Matrix

| Row | Evidence class | Intended validation layer | Target(s) | Core/product ownership | Claim intent |
| --- | --- | --- | --- | --- | --- |
| `FE-U-P4-01` | `product_conformance` | Unit sync-engine model | `make frontend-unit` | Core REQs: `REQ-01-057`, `REQ-01-070`, `REQ-03-217`, `REQ-03-222`, `REQ-03-236`, `REQ-03-241`; Core ACs: `AC-003`, `AC-005`, `AC-040`, `AC-043`, `AC-119`, `AC-120`, `AC-124`, `AC-127`, `AC-181`, `AC-183`, `AC-188`, `AC-193`, `AC-200`, `AC-218`, `AC-221`, `AC-225`, `AC-231`, `AC-299`, `AC-354`, `AC-394`, `AC-396` | Pending queue ordering, retry, success, validation failure, and replay by stable mutation identifiers |
| `FE-U-P4-02` | `product_conformance` | Unit save-state model | `make frontend-unit` | Core REQs: `REQ-03-033`, `REQ-03-040`, `REQ-03-077`, `REQ-03-084`; Core ACs: `AC-009`, `AC-013`, `AC-040`, `AC-041`, `AC-047`, `AC-126`, `AC-163`, `AC-231`, `AC-381`, `AC-023`, `AC-026`, `AC-045`, `AC-050` | One primary save-state label and one same-surface secondary message from pending, saved, failed, and conflict states |
| `FE-I-P4-01` | `product_conformance` | Integration plus browser-backed query/render identity | `make frontend-unit`; `make browser-e2e-webserver-backed` | Core REQs: `REQ-01-034`, `REQ-01-036`, `REQ-01-057`, `REQ-01-070`, `REQ-03-236`, `REQ-03-241`; Core ACs: `AC-119`, `AC-120`, `AC-124`, `AC-127`, `AC-181`, `AC-183`, `AC-188`, `AC-193`, `AC-200`, `AC-218`, `AC-221`, `AC-225`, `AC-231`, `AC-238`, `AC-243`, `AC-299`, `AC-361`, `AC-366`, `AC-367`, `AC-372`, `AC-374` | Timeline query rows render full `view_row_v1` cells and preserve row identity through create, patch, validation error, and refresh |
| `FE-E-P4-01` | `product_conformance` | Public-route browser E2E | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Core REQs: `REQ-01-057`, `REQ-01-070`, `REQ-02-024`, `REQ-02-025`, `REQ-03-217`, `REQ-03-222`, `REQ-03-236`, `REQ-03-241`; Core ACs: `AC-003`, `AC-005`, `AC-040`, `AC-043`, `AC-119`, `AC-120`, `AC-124`, `AC-127`, `AC-181`, `AC-183`, `AC-188`, `AC-193`, `AC-200`, `AC-218`, `AC-221`, `AC-225`, `AC-231`, `AC-299`, `AC-354`, `AC-394`, `AC-396`, `AC-406` | Rough Timeline row creation, inline edit, paste, pending save, refresh, and replay through `/api/v1/` route contracts |
| `FE-V-P4-01` | `design_direction` | Visual regression | `make browser-e2e-visual` | No Core product-conformance claim; support/design ACs: `R2-AC-023`, `R2-AC-026`, `R2-AC-033`, `R2-AC-039`, `R2-AC-045`, `R2-AC-050`, `R2-AC-073`, `R2-AC-079` | Design-direction capture of save-state strip, pending replay indication, inline edit cell, and empty successful Timeline query fixtures |
| `FE-A11Y-P4-01` | `design_direction` | Accessibility preflight in current map; implemented a11y closure requires current guide/map correction or `make browser-e2e-a11y` row mapping | `make browser-e2e-a11y-preflight` currently | No Core product-conformance claim; support/design ACs: `R2-AC-033`, `R2-AC-039`, `R2-AC-080`, `R2-AC-086`, `D-AC-009`, `D-AC-012` | Keyboard and screen-reader safety for grid navigation, edit entry/exit, paste feedback, validation feedback, save-state communication, and `Esc` priority |

Design-direction rows MUST NOT claim Core product conformance. Product-conformance rows MUST close only from current Core-owned row evidence, not from visual, accessibility, support, generated ledger, broad check, or plan text.

## Dependencies And Prerequisites

- FE-P0 through FE-P3 must remain green. Current registry state is `active_green`, but closure work must rerun relevant earlier-phase checks or record precise blockers before FE-P4 completion.
- FE-P4 is currently `planned` and non-executable as a frontend phase. It may move to `active` only when its map, generated ledger, public targets, target schedule metadata, row evidence, evidence-class owner metadata, and evidence freshness are promoted together.
- Generated protocol outputs under `packages/protocol-ts/src/generated/**`, generated backend outputs under `internal/gen/**`, generated frontend ledgers, generated schedules, `tools/task_surface.generated.mk`, `pnpm-lock.yaml`, and `go.sum` must not be hand-edited.
- Any FE-P4 map or registry metadata change requires `make phase-ledgers` followed by `make phase-ledger-drift`.
- Browser rows require service-backed infrastructure: Postgres, MinIO, owned backend/frontend processes, and browser runtime readiness.
- Public-boundary FE-P4 product-conformance rows must use `/api/v1/`, server-managed sessions, stable IDs, public success envelopes, and public error envelopes. Frontend-only mocks cannot close public-route product rows.
- Same-field conflict work in FE-P4 is limited to anchoring and save-state compatibility. Resolver implementation belongs outside FE-P4.
- Accessibility preflight evidence is diagnostic readiness evidence only. It cannot close an implemented accessibility row under the current harness rules.
- Core 05 remains inactive unless FE-P4 introduces `claim_publication_intent="claim_bearing_publication"` with Core 05 predicate metadata.

### Shared Harness Analysis

| Harness | FE-P4 trigger | Plan response |
| --- | --- | --- |
| `FE-H-01` Contract-derived view-schema and field-key mapping | Timeline query/render, create, patch, paste, validation, and refresh all address cells by `field_key` under `view_schema_id`. | Sprints 2 and 4 must keep field identity contract-derived through `/packages/view-contracts`; run `make frontend-unit` and `make frontend-typecheck`; block if labels, visible indexes, or SQL names drive mutation behavior. |
| `FE-H-02` Grid-adapter identity and capability invariants | Inline edit and paste depend on `record_id`, `field_key`, writeability, presentation-row suppression, and adapter capability gating. | Sprints 4 and 5 must preserve FE-P3 adapter guarantees; run `make frontend-unit` and `make browser-e2e-support` if adapter behavior or helpers change. |
| `FE-H-03` Renderer/editor registry behavior and lifecycle cleanup | FE-P4 may add editor lifecycle states for inline edit, validation, or pending replay display. | Required only if FE-P4 changes editor registry, portals, timers, observers, subscriptions, or adapter cleanup. Then run `make frontend-unit` and `make browser-e2e-support`; block if cleanup coverage is absent. |
| `FE-H-04` Sync-engine pending queue and replay behavior | FE-P4 introduces pending queue, retries, failures, and replay for creates, patches, and paste batches. | Sprint 2 owns the unit model; Sprint 5 owns browser public-route replay; both require stable mutation identifiers and row accounting. |
| `FE-H-06` Same-field conflict anchoring | FE-P4 save-state must represent conflict state and preserve anchors without implementing the resolver. | Sprints 2, 3, and 5 must anchor conflict state by `record_id + field_key + base_row_version`; block any resolver UI claim as out of scope. |
| `FE-H-08` Save-state presentation | FE-P4 introduces queued, saved, failed, conflict, and replay-visible save-state. | Sprint 3 owns unit derivation; Sprint 5 validates browser behavior; Sprint 6 captures design-direction visual/a11y state where available. |
| `FE-H-09` Browser command helpers | Paste, edit, replay, and row-anchor browser helpers may need extension. | If helpers change, run `make browser-e2e-support` and keep helper selectors stable; otherwise document no helper change. |
| `FE-H-10` Visual-regression fixtures | FE-P4 visual row maps to `FE-VFIX-08`, `FE-VFIX-12`, and `FE-VFIX-15`. | Sprint 6 must use exact fixture IDs and exact scenario titles. It must not infer closure from snapshot filenames or base `V-*` rows. |
| `FE-H-11` Keyboard and focus traversal | Inline edit, paste, validation feedback, pending replay, and `Esc` priority affect keyboard behavior. | Sprint 5 must keep browser hot-path focus behavior stable; Sprint 6 records current a11y preflight limitation and blocks implemented-row closure until `browser-e2e-a11y` mapping exists. |
| `FE-H-12` Accessibility names, ARIA, and state communication | Save-state, validation errors, pending replay, edit state, and paste feedback need non-color-only communication. | Sprint 6 must verify or block accessible names, live/status communication, focus, and contrast. Current preflight-only target cannot complete the row. |
| `FE-H-13` Stable selector and test-id contracts | FE-P4 browser scenarios need stable selectors for cells, rows, save state, pending queue, validation, paste, and replay controls. | Use `/packages/ui-contracts` builders for cross-boundary selectors; run `make frontend-unit` and `make browser-e2e-support` if selector contracts change. |
| `FE-H-14` Frontend route/API boundary conformance | FE-P4 product rows require `/api/v1/` query and mutation evidence. | Sprint 5 must use public route contracts, server sessions, stable IDs, and public envelopes. Frontend-only route mocks block product closure. |
| `FE-H-16` Frontend error-state rendering | FE-P4 introduces cell-level validation and public error presentation. | Sprints 2, 4, and 5 must render public errors without private details and keep validation anchored to stable cells. |
| `FE-H-17` Generated contract drift | FE-P4 may consume generated protocol shapes or require contract regeneration. | Run `make generated-artifact-policy-check` and `make generate-drift` when contract/generated surfaces are touched; always run before completion if protocol consumption changed. |

No shared harness is satisfied by this plan alone.

## Public Interfaces And Deliverables

- New root plan: `FRONTEND_PHASE4_IMPLEMENTATION_PLAN.md`.
- Expected authored app surfaces: existing `/apps/web` workbook controllers, Timeline query state, sync-engine state, pending queue, public mutation submission, save-state presentation, error rendering, browser scenarios, and unit tests. Exact local module filenames must be inspected during implementation before edits.
- Expected package surfaces: `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/protocol-ts`, `/packages/ui-contracts`, and `/packages/test-utils` only where FE-P4 behavior crosses those package boundaries.
- Expected tests by layer:
  - unit tests for `FE-U-P4-01` and `FE-U-P4-02`;
  - integration-style unit and webserver-backed browser tests for `FE-I-P4-01`;
  - webserver-backed and stateful browser tests for `FE-E-P4-01`;
  - visual regression scenarios for `FE-V-P4-01` using exact fixture IDs;
  - accessibility preflight now, and implemented accessibility target mapping before any FE-A11Y completion claim.
- Expected selector/test-id updates: only through stable builders when selectors cross runtime, unit, browser, support-test, helper, or option-object boundaries.
- Expected phase metadata handling: update `tools/frontend_phase_maps/fe_p4_test_map.json` only after direct row-owned evidence exists; regenerate `docs/testing/frontend_phase_coverage_ledgers/fe_p4_coverage_ledger.md` through `make phase-ledgers`; validate with `make phase-ledger-drift`.
- FE-P4 should consume existing public route contracts. This plan does not create new public API routes or wire shapes.

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers |
| --- | --- | --- | --- |
| [ ] | 1. Readiness, map, ledger, FE-P3 handoff | `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P4`; registry and row-inventory checks; `make phase-ledger-drift`; `git diff --check` | FE-P4 currently planned, non-executable, and all rows blocked |
| [ ] | 2. Sync-engine pending queue unit model for `FE-U-P4-01` | `make frontend-unit`; `make frontend-typecheck` | Block if queue ordering, retry, validation failure, or replay is not keyed by stable mutation identifiers |
| [ ] | 3. Save-state derivation and status strip unit model for `FE-U-P4-02` | `make frontend-unit`; `make frontend-typecheck` | Block if more than one primary label appears or secondary message is detached from the same surface |
| [ ] | 4. Timeline query/render identity integration for `FE-I-P4-01` | `make frontend-unit`; `make browser-e2e-webserver-backed` | Block if full `view_row_v1` cells or `record_id` identity are not preserved through refresh/error |
| [ ] | 5. Public-route E2E for rough create, edit, paste, pending save, refresh, and replay for `FE-E-P4-01` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Block if product evidence uses frontend-only mocks instead of public `/api/v1/` route contracts |
| [ ] | 6. Visual and accessibility readiness for `FE-V-P4-01` and `FE-A11Y-P4-01` | `make browser-e2e-visual`; current `make browser-e2e-a11y-preflight`; implemented closure requires `make browser-e2e-a11y` mapping | FE-V row is blocked until recaptured; FE-A11Y row is preflight-only and cannot complete from preflight evidence |
| [ ] | 7. Closure, drift, final validation, and FE-P5 handoff | Row-owned targets plus `make frontend-import-boundary-check`, `make generated-artifact-policy-check`, `make generate-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make agent-finalize`, and `make check` when required | Block if any row remains blocked/stale or earlier phases cannot be rerun or precisely justified |

## Sprint 1: Readiness, Map, Ledger, And FE-P3 Handoff

Objective: prove FE-P4 planning metadata is traceable before behavior work begins.

Non-goals: no FE-P4 product behavior, no row promotion, no generated-ledger hand edit, no Core 05 activation, and no FE-P4 phase activation.

Owned rows: none. This sprint supports all FE-P4 rows.

Non-owned rows: all FE-P4 rows remain blocked until their owning sprint collects direct row-owned evidence.

Source constraints:

- The frontend guide controls FE-P4 row set and row accounting.
- The authored map controls row inventory; the generated ledger is downstream.
- The FE-P3 handoff is useful context but historical evidence unless rerun.

Inspection checklist:

- Inspect `docs/guides/cartulary_frontend_implementation_testing_guide.md` Phase FE-P4 and shared harness table.
- Inspect `tools/frontend_phase_registry.json` FE-P0 through FE-P4 entries.
- Inspect `tools/frontend_phase_maps/fe_p4_test_map.json`.
- Inspect `docs/testing/frontend_phase_coverage_ledgers/fe_p4_coverage_ledger.md`.
- Inspect `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md` through `FRONTEND_PHASE3_IMPLEMENTATION_PLAN.md`.
- Inspect `tools/frontend_visual_fixture_registry.json` and `docs/guides/cartulary_visual_golden_maintenance.md` for FE-P4 fixture IDs.
- Inspect the Make target surface through `make help`, `make explain-phase`, and `make explain-target`.

Test-first sequence:

- Run `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P4`.
- Validate the FE-P4 registry tuple with `jq -e`.
- Validate FE-P4 row inventory, duplicate count, and blocked claim statuses with `jq -e`.
- Run `make phase-ledger-drift`.
- Run `git diff --check` after creating or updating this plan.

Validation commands:

- Already passed during plan creation: `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P4`.
- Already passed during plan creation: registry tuple `jq -e` check.
- Already passed during plan creation: row inventory `jq -e` check.
- Already passed after this plan was written: `make phase-ledger-drift`, run root `.cartulary/test-results/20260601T210509Z-p23551`.
- Passed after file write: `git diff --check` with the new file temporarily marked intent-to-add so the untracked Markdown file was included in whitespace validation.

Deliverables:

- This root plan.
- Recorded source limits, exact row inventory, current blockers, target availability, and fixture IDs.

Blocker rules:

- `BLOCKER: FE-P4 map row inventory invalid; expected exactly FE-U-P4-01, FE-U-P4-02, FE-I-P4-01, FE-E-P4-01, FE-V-P4-01, FE-A11Y-P4-01 once each; actual=<ids/counts>.`
- `BLOCKER: FE-P4 registry tuple is not frontend-namespace traceable; expected namespace/path/dependency tuple does not match registry.`
- `BLOCKER: FE-P4 generated ledger is stale relative to map; rerun generator only after confirming the authored map is the intended source.`
- `BLOCKER: FE-P3 handoff is not current enough for FE-P4 closure; minimum_follow_up=<rerun exact FE-P3 regression or record accepted owner rationale>.`

Binary acceptance:

- FE-P4 is explainable in the frontend namespace.
- Registry tuple and row inventory are exact.
- Generated ledger drift passes.
- FE-P4 remains planned and blocked; no row is promoted.
- Plan records source limits and non-claims.

Explicit non-claims:

- Sprint 1 does not prove Timeline query, create, edit, paste, save-state, replay, validation, visual, or accessibility behavior.
- Sprint 1 does not close any FE-P4 row.
- Sprint 1 does not activate Core 05.

## Sprint 2: Sync-Engine Pending Queue Unit Model

Objective: implement the unit-level sync-engine model for `FE-U-P4-01`.

Owned row: `FE-U-P4-01`.

Non-owned rows: `FE-U-P4-02`, `FE-I-P4-01`, `FE-E-P4-01`, `FE-V-P4-01`, and `FE-A11Y-P4-01`.

Source constraints:

- Core 01 Section 3.3.6 owns success/error envelope and mutation route behavior.
- Core 03 Sections 4.1 and 15 own autosave and Timeline read/write interaction behavior.
- The sync engine must be unit-testable without private server behavior.

Test-first sequence:

- Add failing unit tests that cover creates, patches, paste batches, retries, successes, validation failures, and replay.
- Assert ordering by stable mutation identifiers, not visible row order.
- Assert replay idempotency behavior uses stable keys such as `client_txn_id`, `record_id`, `base_row_version`, and `field_key` as applicable.
- Assert non-retryable validation failure stops or surfaces according to public envelope semantics.

Implementation tasks:

- Model queue entries for row create, row patch, and paste batch.
- Preserve admission order and dispatch order deterministically.
- Associate every queue entry with a stable mutation identity and public route shape.
- Distinguish pending, dispatching, retryable failure, validation failure, conflict, and success outcomes.
- Apply successful row refreshes without retargeting pending entries by visible row index.
- Surface validation failures as cell-level state for later rendering work.
- Keep same-field conflict resolver implementation out of scope; preserve only conflict anchoring.

Validation commands:

- `make frontend-unit`
- `make frontend-typecheck`
- `make generated-artifact-policy-check` and `make generate-drift` if generated protocol consumption changes
- `make phase-ledgers` and `make phase-ledger-drift` only after row metadata is intentionally promoted

Evidence requirements:

- `frontend-unit` must emit current frontend row accounting that closes `FE-U-P4-01` when the row is promoted.
- Unit evidence must use public route-shaped success/error results, not private server internals.
- The row must remain `claim_status="blocked"` until direct current evidence exists.

Blocker rules:

- `BLOCKER: FE-U-P4-01 pending queue evidence missing stable mutation identifiers; minimum_follow_up=add unit coverage for create, patch, paste, retry, success, validation failure, and replay identity.`
- `BLOCKER: FE-U-P4-01 product evidence uses visible row order or labels as mutation identity; minimum_follow_up=replace with record_id, field_key, base_row_version, view_schema_id, and client_txn_id as owner contracts require.`
- `BLOCKER: FE-U-P4-01 target passed but frontend row accounting did not close row; target=frontend-unit failure_reason=frontend_row_accounting.`

Binary acceptance:

- `FE-U-P4-01` is implemented only when the row is closed by current `make frontend-unit` evidence and the FE-P4 map/ledger are updated through the proper generator flow.

Explicit non-claims:

- Sprint 2 does not prove browser public-route behavior.
- Sprint 2 does not prove visual or accessibility behavior.
- Sprint 2 does not implement the same-field conflict resolver.

## Sprint 3: Save-State Derivation And Status Strip Unit Model

Objective: implement save-state derivation for `FE-U-P4-02`.

Owned row: `FE-U-P4-02`.

Non-owned rows: `FE-U-P4-01`, `FE-I-P4-01`, `FE-E-P4-01`, `FE-V-P4-01`, and `FE-A11Y-P4-01`.

Source constraints:

- Core 03 Sections 4.1 and 4.4 own autosave and local pending queue behavior.
- UI/UX guide and design docs contribute design-direction only and must not widen product-conformance claims.

Test-first sequence:

- Add unit tests for pending, saved, failed, conflict, queue-overflow, validation-failure, and replay-halted states.
- Assert exactly one primary label.
- Assert exactly one secondary message on the same surface when a secondary message is needed.
- Assert conflict state remains compatible with `record_id + field_key + base_row_version` anchoring.

Implementation tasks:

- Build or harden a pure save-state derivation function.
- Feed save-state derivation from pending queue and conflict/error state rather than DOM or visible labels.
- Preserve status strip capacity limits.
- Keep validation and conflict text non-color-only and ready for a11y work.

Validation commands:

- `make frontend-unit`
- `make frontend-typecheck`
- `make browser-e2e-visual` later in Sprint 6 for design-direction visual capture
- `make phase-ledgers` and `make phase-ledger-drift` only after row metadata is intentionally promoted

Evidence requirements:

- `frontend-unit` must close `FE-U-P4-02` through current frontend row accounting before promotion.
- Design-direction status-strip evidence cannot replace product-conformance unit evidence.

Blocker rules:

- `BLOCKER: FE-U-P4-02 save-state derivation emits more than one primary label; minimum_follow_up=normalize pending/saved/failed/conflict precedence into one primary label.`
- `BLOCKER: FE-U-P4-02 save-state secondary message is detached from the same workbook surface; minimum_follow_up=render or derive secondary message in the status strip surface.`
- `BLOCKER: FE-P4 evidence classes collapsed; design/support/claim-publication-boundary evidence cannot be counted as product_conformance.`

Binary acceptance:

- `FE-U-P4-02` is implemented only when current `make frontend-unit` row accounting closes it and the generated ledger is refreshed from the authored map.

Explicit non-claims:

- Sprint 3 does not prove public-route replay.
- Sprint 3 does not close visual or accessibility rows.
- Sprint 3 does not implement conflict resolution.

## Sprint 4: Timeline Query And Render Identity Integration

Objective: implement query/render identity integration for `FE-I-P4-01`.

Owned row: `FE-I-P4-01`.

Non-owned rows: `FE-U-P4-01`, `FE-U-P4-02`, `FE-E-P4-01`, `FE-V-P4-01`, and `FE-A11Y-P4-01`.

Source constraints:

- Core 01 Section 3.3.4 owns `view_row_v1`, full cell serialization, and query response shape.
- Core 01 Section 3.3.6 owns success and error envelopes.
- Core 03 Section 15 owns Timeline read and write behavior.

Test-first sequence:

- Add integration tests that render full `view_row_v1` cells for Timeline rows.
- Add tests for row identity preservation through create, patch, validation error, and refresh.
- Add browser-backed scenario with the exact current map title: `FE-I-P4-01 Verify Timeline query response rows render full view_row_v1 cells and preserve row identity through create, patch, validation error, and refresh.`
- Assert missing or omitted schema-declared non-technical cells fail or render a public error rather than silently dropping fields.

Implementation tasks:

- Feed Timeline query rows through `/packages/view-contracts` and `/packages/grid-adapter`.
- Preserve `record_id`, `row_version`, `view_schema_id`, and `field_key` through render, refresh, validation, and retry paths.
- Ensure create and patch refreshes merge by `record_id`.
- Ensure validation errors anchor to stable cell identity.
- Keep presentation rows non-mutating.

Validation commands:

- `make frontend-unit`
- `make frontend-typecheck`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-support` if browser helpers or selector contracts change

Evidence requirements:

- `frontend-unit` and `browser-e2e-webserver-backed` must retain current row accounting for `FE-I-P4-01`.
- Browser scenario title must match the phase map exactly.
- Passing unit tests alone do not prove public browser boundary behavior if the browser target remains unmapped or missing row accounting.

Blocker rules:

- `BLOCKER: FE-P4 browser row lacks exact scenario_titles[] for row-owned closure; row=FE-I-P4-01 target=browser-e2e-webserver-backed.`
- `BLOCKER: FE-I-P4-01 Timeline query row omitted full view_row_v1 cells; minimum_follow_up=preserve every schema-declared non-technical field under rows[].cells.`
- `BLOCKER: FE-I-P4-01 row identity retargeted by visible index after refresh; minimum_follow_up=anchor refresh and validation state by record_id and field_key.`

Binary acceptance:

- `FE-I-P4-01` is implemented only when both mapped layers that are required by the current map and harness close with current row accounting.

Explicit non-claims:

- Sprint 4 does not prove replay after transient failure.
- Sprint 4 does not prove visual or accessibility closure.
- Sprint 4 does not implement saved-view persistence or WebSocket live updates.

## Sprint 5: Public-Route E2E For Rough Create, Edit, Paste, Pending Save, Refresh, And Replay

Objective: implement public-route E2E coverage for `FE-E-P4-01`.

Owned row: `FE-E-P4-01`.

Non-owned rows: `FE-U-P4-01`, `FE-U-P4-02`, `FE-I-P4-01`, `FE-V-P4-01`, and `FE-A11Y-P4-01`.

Source constraints:

- Product-conformance evidence must use public `/api/v1/` route contracts and server-managed sessions.
- Core 02 Section 5 owns rough and uncertain input preservation.
- Frontend-only mocks, private server details, snapshot filenames, and generated ledgers cannot close this row.

Test-first sequence:

- Add or extend browser scenario with the exact current map title: `FE-E-P4-01 Verify rough Timeline row creation, inline edit, paste, pending save, refresh, and replay through /api/v1/ route contracts.`
- Start from an authenticated browser session and a real incident.
- Query Timeline through public route-backed UI.
- Create a rough Timeline row with low-friction input.
- Inline edit a writable cell and observe pending then saved state.
- Paste deterministic tabular content and verify ordered creates or patches by stable row/cell anchors.
- Induce a transient failure only through an accepted harness-owned public test control or service-boundary behavior. If only private frontend mocks are available, block the row.
- Refresh and verify stable row identity and replay behavior.
- Verify cell-level validation and public error-envelope rendering.

Implementation tasks:

- Wire FE-P4 browser hot path to `/api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, row create, patch, and clipboard-paste route contracts.
- Preserve `client_txn_id`, `record_id`, `base_row_version`, and `field_key` through replay.
- Keep rough capture valid without mention resolution or evidence handles.
- Prevent replay from applying later units out of order after non-retryable failure.
- Keep validation and conflict state on the same workbook surface.

Validation commands:

- `make browser-e2e-webserver-backed`
- `make browser-e2e-stateful`
- `make browser-e2e-support` if browser helpers or selectors change
- `make frontend-unit` if app controller or helper unit code changes
- `make frontend-typecheck`

Evidence requirements:

- Both mapped browser targets must emit current frontend row accounting for `FE-E-P4-01` when required by the harness.
- Browser evidence must show `/api/v1/` route use, server-managed session behavior, stable IDs, and public error envelopes.
- Any test-only control used to trigger transient failure must be harness-owned and not a production route behavior claim.

Blocker rules:

- `BLOCKER: FE-P4 public-boundary behavior was tested through frontend-only mocks; product-conformance route evidence requires public /api/v1/ browser-facing evidence.`
- `BLOCKER: FE-E-P4-01 replay cannot be induced through a harness-owned public boundary; minimum_follow_up=add accepted test control or record row as blocked.`
- `BLOCKER: FE-E-P4-01 browser target passed but row accounting is missing or stale; target=<target> failure_reason=frontend_row_accounting.`

Binary acceptance:

- `FE-E-P4-01` is implemented only when current browser row accounting closes the row through the mapped webserver-backed and stateful targets, or the current map is corrected with owner-approved target semantics before promotion.

Explicit non-claims:

- Sprint 5 does not close visual or accessibility rows.
- Sprint 5 does not implement WebSocket live updates.
- Sprint 5 does not implement same-field conflict resolver UI.
- Sprint 5 does not claim Core 05 publication evidence.

## Sprint 6: Visual And Accessibility Readiness

Objective: collect or precisely block design-direction FE-P4 visual and accessibility evidence.

Owned rows: `FE-V-P4-01`, `FE-A11Y-P4-01`.

Non-owned rows: `FE-U-P4-01`, `FE-U-P4-02`, `FE-I-P4-01`, and `FE-E-P4-01`.

Source constraints:

- Visual evidence is design-direction only.
- Accessibility evidence is design-direction only.
- Current FE-P4 accessibility target is `make browser-e2e-a11y-preflight`, which cannot complete implemented-row accessibility evidence.
- Visual fixture identity must come from `tools/frontend_visual_fixture_registry.json`.

Test-first sequence:

- Validate exact FE-P4 visual fixture IDs: `FE-VFIX-08`, `FE-VFIX-12`, and `FE-VFIX-15`.
- Verify whether pending replay indication is covered by `FE-VFIX-08` or requires a registry/map update with a precise blocker.
- Add or update exact visual scenario title from the FE-P4 map.
- Run `make browser-e2e-visual`.
- Run current `make browser-e2e-a11y-preflight` only as blocked-row smoke.
- If the row is promoted to implemented accessibility evidence, update the map to `make browser-e2e-a11y` and require `cartulary.frontend_accessibility_summary.v2`.

Implementation tasks:

- Capture save-state strip, pending replay indication, inline edit cell, and empty successful Timeline query in deterministic app-owned browser state.
- Preserve deterministic seed, viewport, zoom, mask, scroll, focus/editor, inspector, and post-scroll settle metadata.
- Add accessibility checks for grid navigation, edit entry/exit, paste feedback, validation feedback, save-state communication, and `Esc` priority only when mapped to the implemented-row target.

Validation commands:

- `make browser-e2e-visual`
- `make browser-e2e-a11y-preflight`
- `make browser-e2e-a11y` only after map and target semantics support implemented FE-P4 accessibility row closure
- `make browser-e2e-support` if selectors or helper choreography change
- `make frontend-unit` if selector builders or state derivation change

Evidence requirements:

- `FE-V-P4-01` must close from exact scenario titles and frontend row accounting, not snapshot filenames or base `V-*` rows.
- `FE-A11Y-P4-01` must remain blocked while mapped only to preflight. Preflight artifacts may support diagnostics but not completion.
- Visual and accessibility evidence must not be promoted into product conformance.

Blocker rules:

- `BLOCKER: FE-V-P4-01 fixture identity missing or ambiguous; expected exact FE-VFIX-08, FE-VFIX-12, and FE-VFIX-15 linkage or precise map/registry correction.`
- `BLOCKER: FE-V-P4-01 visual fixture is current in registry but not recaptured as FE-P4 row-owned evidence; target=browser-e2e-visual minimum_follow_up=run exact FE-P4 scenario with frontend row accounting.`
- `BLOCKER: FE-P4 accessibility row is preflight-only and cannot be counted as implemented accessibility completion; row=FE-A11Y-P4-01 target=browser-e2e-a11y-preflight minimum_follow_up=map implemented scenario to browser-e2e-a11y or keep row blocked.`

Binary acceptance:

- `FE-V-P4-01` is implemented only when direct visual row accounting closes it.
- `FE-A11Y-P4-01` is not implemented under the current preflight-only closure posture.

Explicit non-claims:

- Sprint 6 does not prove product conformance.
- Sprint 6 does not activate Core 05.
- Sprint 6 does not infer closure from snapshot filenames, old retained artifacts, or fixture registry `current` status.

## Sprint 7: Closure, Drift, Final Validation, And FE-P5 Handoff

Objective: close FE-P4 only when all row-owned evidence, shared harnesses, drift checks, and earlier-phase dependencies are current.

Owned rows: none directly. This sprint composes closure evidence for all FE-P4 rows.

Non-owned rows: none; all row claims must be completed by their owning sprints before closure.

Source constraints:

- FE-P4 cannot be represented as complete with any row still `blocked`, `stale`, or missing row accounting.
- A broad `make check` pass alone cannot close FE-P4 rows.
- Generated ledgers must be regenerated, not hand-edited.

Test-first sequence:

- Confirm every FE-P4 product row has direct current row-owned evidence.
- Confirm design-direction rows are either closed by intended current design evidence or explicitly blocked and excluded from completion.
- Confirm every triggered shared harness is satisfied or precisely blocked.
- Rerun earlier-phase checks required by risk and touched surfaces.
- Run drift and policy checks.
- Run `make agent-finalize` before broader end-of-run verification. If using a retained successful full warm check run, pass `RESULTS_DIR=<run-root>`.
- Run `make check` when repository completion rules require the broad developer gate.

Validation commands:

- `make frontend-typecheck`
- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-stateful`
- `make browser-e2e-visual`
- `make browser-e2e-a11y-preflight` for blocked-row smoke only
- `make browser-e2e-a11y` when the row is implemented and mapped
- `make browser-e2e-support` when shared helpers or selectors change
- `make frontend-import-boundary-check`
- `make generated-artifact-policy-check`
- `make generate-drift`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `make agent-finalize`
- `make check` when the completion gate requires it
- `git diff --check`

Evidence requirements:

- Retain run roots and exact summary paths for every row-owned command.
- Update FE-P4 map statuses only when direct evidence exists.
- Regenerate FE-P4 ledger through `make phase-ledgers`.
- Keep evidence classes separate in map, ledger, plan, and handoff.
- Activate FE-P4 only when the registry, map, ledger, freshness digests, and row evidence are promoted together.

Blocker rules:

- `BLOCKER: FE-P4 row remains blocked at closure; row=<row_id> blocker=<reason_code> minimum_follow_up=<specific target or owner patch>.`
- `BLOCKER: FE-P4 generated ledger is stale relative to map; rerun generator only after confirming authored map is intended source.`
- `BLOCKER: FE-P4 earlier active phase regression failed; phase=<FE-P0..FE-P3> target=<target> run_root=<run_root> minimum_follow_up=<fix or owner acceptance>.`
- `BLOCKER: FE-P4 evidence freshness digest stale; minimum_follow_up=rerun freshness/ledger validation after map, registry, fixture, and target evidence are final.`

Binary acceptance:

- All intended FE-P4 rows close through current row-owned evidence.
- Shared harnesses are satisfied or have precise blocking entries that exclude affected rows from completion.
- FE-P0 through FE-P3 remain green or have precise owner-accepted blockers that do not invalidate FE-P4 closure.
- `make phase-ledger-drift`, `make phase-schedule-drift`, generated-artifact policy, generated drift, and whitespace checks pass.
- Core/product, design-direction, implementation-support, and claim-publication evidence classes remain separate.

Explicit non-claims:

- Sprint 7 does not use old retained artifacts, generated ledgers, broad `make check`, visual goldens, support-only tests, accessibility preflight smoke, test names, or this plan as row completion evidence.
- Sprint 7 does not activate Core 05 unless claim-publication metadata is explicit.

## Blocker Recording Rules

Every blocker must record:

- exact command or missing artifact;
- failing target, scheduler unit, or missing manifest path;
- run root when available;
- FE-P4 row ID or shared harness ID;
- failure class and reason when exposed by the harness;
- ownership: product, design, support, harness, generated, fixture, or source-doc;
- minimum follow-up action;
- whether it blocks FE-P4 completion.

Required blocker language:

- `BLOCKER: FE-P4 map row inventory invalid; expected exactly FE-U-P4-01, FE-U-P4-02, FE-I-P4-01, FE-E-P4-01, FE-V-P4-01, FE-A11Y-P4-01 once each; actual=<ids/counts>.`
- `BLOCKER: FE-P4 registry tuple is not frontend-namespace traceable; expected namespace/path/dependency tuple does not match registry.`
- `BLOCKER: FE-P4 generated ledger is stale relative to map; rerun generator only after confirming the authored map is the intended source.`
- `BLOCKER: FE-P4 owner refs unresolved; row=<row_id> owner_ref=<source/section/req/ac> minimum_follow_up=<specific inspection or owner patch>.`
- `BLOCKER: FE-P4 browser row lacks exact scenario_titles[] for row-owned closure; row=<row_id> target=<target>.`
- `BLOCKER: FE-P4 accessibility row is preflight-only and cannot be counted as implemented accessibility completion; row=FE-A11Y-P4-01 target=browser-e2e-a11y-preflight minimum_follow_up=<required implemented-row a11y target or map status correction>.`
- `BLOCKER: FE-P4 evidence classes collapsed; design/support/claim-publication-boundary evidence cannot be counted as product_conformance.`
- `BLOCKER: FE-P4 public-boundary behavior was tested through frontend-only mocks; product-conformance route evidence requires public /api/v1/ browser-facing evidence.`

Strict non-claims:

- Do not claim FE-P4 completion from generated ledgers, old retained artifacts, broad `make check`, test names, visual goldens, support-only tests, accessibility preflight smoke, or this plan.
- Do not hand-edit generated ledgers.
- Do not invent run roots, test counts, row statuses, target availability, current package files, fixture IDs, or registry status.
- Do not promote visual or accessibility evidence into product-conformance evidence.
- Do not activate Core 05 unless a claim-bearing publication predicate is explicit.
- Do not add FE-P4 rows to base `tools/phase*_test_map.json`; frontend rows use the frontend namespace.

## Binary Exit Criteria

Plan creation is complete when:

- this root plan exists;
- current FE-P4 source facts, registry tuple, row inventory, claim statuses, activation blockers, visual fixture IDs, and target availability are recorded from local inspection;
- readiness checks are run or precise blockers are recorded;
- only this plan is changed unless a traceability correction is strictly required;
- whitespace validation passes.

FE-P4 phase completion is allowed only when:

- FE-P4 registry status and row rollup are updated through the proper authored metadata path;
- every intended FE-P4 row closes from direct current row-owned evidence;
- all product-conformance rows have Core 00 through Core 04 or adopted NLSpec ownership with resolved REQ and AC IDs;
- all design-direction rows remain design-direction only;
- every triggered shared harness is satisfied or blocks affected completion claims;
- generated ledgers and schedules are regenerated through Make and drift checks pass;
- generated-artifact policy and generated drift checks pass when generated or contract surfaces are touched;
- earlier active frontend phases remain green or precise owner-accepted blockers are recorded;
- Core/design/support/claim-publication evidence classes remain separate;
- no Core 05 claim-publication evidence is claimed unless explicit claim metadata satisfies Core 05.

## FE-P5 Handoff

If FE-P4 completes, FE-P5 may rely on:

- a route-backed Timeline query surface that renders full `view_row_v1` cells by stable identity;
- rough Timeline row creation through public route contracts;
- inline edit and paste hot paths anchored by `record_id`, `field_key`, `base_row_version`, `view_schema_id`, and `client_txn_id`;
- deterministic pending queue and replay behavior;
- same-surface save-state and validation/error presentation;
- visual and accessibility readiness state recorded without promoting design evidence into product conformance.

Current unresolved blockers at plan creation:

- FE-P4 is `planned`, not active.
- All six FE-P4 rows have `claim_status="blocked"`.
- `FE-V-P4-01` is blocked because direct frontend row-accounting visual evidence has not been recaptured under the closed fixture registry.
- `FE-A11Y-P4-01` is mapped to `make browser-e2e-a11y-preflight`; preflight evidence cannot complete an implemented accessibility row.
- No FE-P4 row-owned implementation target has been run in this planning task.

Non-claims carried to FE-P5:

- No entity mention resolution.
- No evidence handles.
- No WebSocket live updates.
- No same-field conflict resolver implementation.
- No saved-view persistence.
- No Core 05 claim-bearing publication evidence.
