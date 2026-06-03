# Frontend Phase 4 Implementation Plan

## Summary

Frontend verification contract. This file is the execution roadmap, progress marker, and FE-P5 handoff aid for `FE-P4: Timeline Hot Path And Sync Engine`.

Frontend verification contract. This plan is not behavior authority and MUST NOT mark any FE-P4 row complete without direct current row-owned evidence from the mapped target and `cartulary.frontend_row_accounting.v2` where the harness requires it.

Source limit. This plan is not behavior authority. Sprint execution records may summarize row-owned implementation outcomes, but they do not activate FE-P4, do not replace generated ledgers, and do not introduce a Core 05 claim-publication predicate.

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

Locally verified facts as of 2026-06-01, with Sprint 2, Sprint 3, and Sprint 4 updates noted below:

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
| Row rollup state | `partially_implemented` |
| Manifest path | `tools/frontend_phase_maps/fe_p4_test_map.json` |
| Ledger path | `docs/testing/frontend_phase_coverage_ledgers/fe_p4_coverage_ledger.md` |
| Depends on | `FE-P0`, `FE-P1`, `FE-P2`, `FE-P3` |
| Activation blocker | `FE-P4-ACTIVATION-BLOCKER-01`, `reason_code="frontend_phase_not_active"` |
| Manifest digest | `573d699ee59e3e9c613acd57f650991756a71f3b6b3c118744818f219d92cf47` |
| Ledger digest | `fa39e019598bcc765a94271e46ad197e2714fa0ab4147c0fb1fe7fc449b67212` |
| Evidence freshness digest | `2280965556d8bcdb20e6dfdc340fe9184b41a46bb6f0f2ed89e4773a1ecdf6ea` |

FE-P0 through FE-P4 current activation chain:

| Phase | Status | Row rollup state | Activation blockers | Depends on |
| --- | --- | --- | --- | --- |
| `FE-P0` | `active` | `active_green` | none | none |
| `FE-P1` | `active` | `active_green` | none | `FE-P0` |
| `FE-P2` | `active` | `active_green` | none | `FE-P0`, `FE-P1` |
| `FE-P3` | `active` | `active_green` | none | `FE-P0`, `FE-P1`, `FE-P2` |
| `FE-P4` | `planned` | `partially_implemented` | `frontend_phase_not_active` | `FE-P0`, `FE-P1`, `FE-P2`, `FE-P3` |

FE-P4 row inventory:

| Row | Layer | Evidence class | Claim status | Target(s) | Scenario titles | Current blocker |
| --- | --- | --- | --- | --- | --- | --- |
| `FE-U-P4-01` | `unit` | `product_conformance` | `implemented` | `make frontend-unit` | 9 exact row-owned titles | none |
| `FE-U-P4-02` | `unit` | `product_conformance` | `implemented` | `make frontend-unit` | 4 exact row-owned titles | none |
| `FE-I-P4-01` | `integration` | `product_conformance` | `implemented` | `make frontend-unit`; `make browser-e2e-webserver-backed` | exact row title present | none |
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
- Current row-owned `frontend-unit` evidence closes `FE-U-P4-01`, `FE-U-P4-02`, and `FE-I-P4-01`; current row-owned `browser-e2e-webserver-backed` evidence also closes `FE-I-P4-01`. Existing latest artifacts for other FE-P4 rows remain historical diagnostic context until their owning targets rerun.
- The current FE-P4 accessibility row is preflight-only in the map. Preflight evidence cannot complete an implemented accessibility row under the current frontend guide and testing harness NLSpec.
- The plan does not inspect every implementation module under `/apps/web`; exact local module filenames must be rechecked in the owning sprint before product edits.

Sprint 2 remediation update as of 2026-06-02:

- `FE-U-P4-01` is implemented from current `frontend-unit` row accounting after expanded model coverage and `WorkbookShell` runtime convergence on the shared pending-queue model.
- `WorkbookShell` pending replay uses the shared queue semantics for admission, replay, retry, overflow, auth pause, same-field conflict, halt, success settlement, and save-state derivation.
- Paste-derived queue-unit evidence MUST NOT be used as full clipboard planning evidence. Actual paste closure belongs to public-route browser evidence for `/api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste`.
- FE-P4 remains `planned`; `FE-I-P4-01`, `FE-E-P4-01`, `FE-V-P4-01`, and `FE-A11Y-P4-01` remain blocked.

Sprint 3 update as of 2026-06-02:

- `FE-U-P4-02` is implemented from current `frontend-unit` row accounting after save-state derivation and status-strip unit coverage.
- `deriveWorkbookSaveState` owns Table C primary-label precedence and same-surface secondary message derivation from pending queue, conflict, overflow, halted replay, paused replay, in-flight mutation, and saved inputs.
- `WorkbookShell` publishes the derived save-state presentation into the status strip; ambient presence-only updates do not move the primary label away from `Saved`.
- FE-P4 remains `planned`; `FE-I-P4-01`, `FE-E-P4-01`, `FE-V-P4-01`, `FE-A11Y-P4-01`, and phase activation remain blocked.

Sprint 4 update as of 2026-06-02:

- `FE-I-P4-01` is implemented from current row-owned `frontend-unit` and `browser-e2e-webserver-backed` evidence using the exact scenario title `FE-I-P4-01 Verify Timeline query response rows render full view_row_v1 cells and preserve row identity through create, patch, validation error, and refresh.`
- Timeline query, create-success, patch-success, refresh, and record-changed patch rows now route through the view-contract `view_row_v1` boundary before grid rendering or reconciliation.
- Full Timeline row coverage requires schema-declared non-technical cells, including hidden and default-hidden fields, treats `{ "value": null }` as authoritative null, rejects omitted declared cells, rejects technical or unknown cells, and preserves top-level `record_id`, `row_version`, and `view_schema_id`.
- Browser-backed coverage exercises omitted and empty `sort` and `filters`, omitted `group_by`, rejected `group_by: null`, required `meta.query`, default-tail sort behavior, public error rendering for invalid full rows, and stable identity through create, patch, validation error, and refresh.
- FE-P4 remains `planned`; `FE-E-P4-01`, `FE-V-P4-01`, `FE-A11Y-P4-01`, and phase activation remain blocked.

## Phase Objective

Core restatement. FE-P4 must turn the existing workbook shell and grid adapter into the first Timeline mutation hot path through public route contracts.

At FE-P4 exit, a browser user must be able to:

- query the Timeline and render full `view_row_v1` cells by stable row and field identity;
- create a low-friction rough Timeline row without requiring entity mention resolution, evidence handles, or canonical enrichment first;
- edit cells inline;
- paste tabular data into the Timeline hot path;
- see `Syncing`, `Saved`, and `Conflict` save-state feedback, plus same-surface secondary failure, overflow, and replay messages;
- recover from transient write failure through deterministic pending replay;
- see cell-level validation and public error-envelope feedback without private server details;
- refresh without losing stable row identity or retargeting pending work by visible row order.

FE-P4 phase completion requires current row-owned evidence for all six FE-P4 rows plus satisfaction or precise blocking for every triggered shared harness. Product hot-path closure is a narrower handoff claim and does not equal FE-P4 phase completion.

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
| `FE-U-P4-02` | `product_conformance` | Unit save-state model | `make frontend-unit` | Core REQs: `REQ-03-033`, `REQ-03-040`, `REQ-03-077`, `REQ-03-084`; Core ACs: `AC-009`, `AC-013`, `AC-040`, `AC-041`, `AC-047`, `AC-126`, `AC-163`, `AC-231`, `AC-381`, `AC-023`, `AC-026`, `AC-045`, `AC-050` | Exactly one primary save-state label from `Syncing`, `Saved`, or `Conflict`, plus one same-surface secondary message when failure, overflow, validation, or replay-halted detail is required |
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
| `FE-H-08` Save-state presentation | FE-P4 introduces `Syncing`, `Saved`, and `Conflict` save-state labels plus replay-visible same-surface secondary detail. | Sprint 3 owns unit derivation; Sprint 5 validates browser behavior; Sprint 6 captures design-direction visual/a11y state where available. |
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

## NLSpec Boundary Closure Contract

This section is the closed FE-P4 implementation contract for readers of this plan. It imports the Core-owned and harness-owned boundary values that FE-P4 implementation work needs in order to avoid drift. Core 00 through Core 04 remain the normative product owners, and `docs/testing-harness-nlspec.md` remains the normative harness owner. When this plan restates a Core or harness rule, the restatement MUST preserve the exact owner semantics. If a future edit discovers a conflict, the owner document wins and this plan MUST be corrected before FE-P4 row promotion or phase-completion claims.

Normative terms:

- `MUST`, `MUST NOT`, `SHOULD`, and `MAY` have their ordinary requirement force in this plan.
- `omitted` means the JSON member, map member, table row, fixture link, scenario title, or artifact is absent.
- `explicit null` means the JSON member is present with value `null`; explicit null is distinct from omission unless the owner contract says they compare equal.
- `blocked` means the row, target, fixture, or claim lacks current direct evidence or has a recorded reason that prevents the affected closure claim.
- `closed` means the row or closure claim has current direct row-owned evidence from the mapped target and required row accounting.
- `product hot-path closed` is a narrow FE-P4 handoff claim for product-conformance rows only. It is not FE-P4 phase completion.

### Completion State Model

Table A. FE-P4 Completion Claims:

| Claim name | Included rows | Required evidence | Allowed blockers | Forbidden wording |
| --- | --- | --- | --- | --- |
| `product hot-path closed` | `FE-U-P4-01`, `FE-U-P4-02`, `FE-I-P4-01`, `FE-E-P4-01` | Current row-owned evidence from every mapped product-conformance target, current frontend row accounting where required, exact scenario titles where required, and generated ledger freshness after map promotion. | Design-direction rows may remain blocked only if the handoff explicitly says `FE-P4 product hot-path closed; FE-P4 phase completion blocked by <row/blocker>`. Product rows may not remain blocked. | `FE-P4 complete`, `phase complete`, `all FE-P4 rows closed`, or any wording that implies visual or accessibility closure. |
| `visual readiness closed` | `FE-V-P4-01` | Current `browser-e2e-visual` row-owned evidence, exact FE-P4 scenario title, and exact fixture linkage for `FE-VFIX-08`, `FE-VFIX-12`, and `FE-VFIX-15` or an owner-approved map/registry correction. | None for this claim. Missing fixture identity, stale fixture evidence, or snapshot-only evidence blocks the claim. | `product conformance`, `Core conformance`, or closure from snapshot filenames alone. |
| `accessibility readiness closed` | `FE-A11Y-P4-01` | Current implemented accessibility row evidence from `make browser-e2e-a11y`, exact FE-P4 row mapping, and `cartulary.frontend_accessibility_summary.v2`. | None for this claim. Current preflight-only mapping blocks the claim. | `implemented a11y closed` from `browser-e2e-a11y-preflight`, or product-conformance wording. |
| `FE-P4 phase complete` | All six FE-P4 rows: `FE-U-P4-01`, `FE-U-P4-02`, `FE-I-P4-01`, `FE-E-P4-01`, `FE-V-P4-01`, `FE-A11Y-P4-01` | Product hot-path closed, visual readiness closed, accessibility readiness closed, all triggered shared harnesses satisfied, generated drift checks passing, FE-P0 through FE-P3 still green or precisely owner-accepted, and registry/map/ledger/freshness promotion performed together. | None. A blocked row may be documented, but it prevents this claim. | Any wording that excludes blocked design rows from full FE-P4 phase completion. |

### Table B. Pending Queue Boundary

| Boundary | Closed FE-P4 rule |
| --- | --- |
| Queue unit | One autosave-originated workbook hot-path mutation: one row-create intent or one row-patch intent. Paste batches are represented by their route-shaped create and patch replay units. |
| Scope | The local pending queue is memory-local, incident-scoped, and client-instance-scoped by `(incident_id, client_instance_id)`. |
| Capacity | Exactly `64` replay units per `(incident_id, client_instance_id)`. The 65th non-coalescible unit MUST be refused. |
| Ordering | Admission order and replay order are FIFO. The client MUST NOT reorder queued writes by visible row order, sort order, record type, labels, or other presentation-derived state. |
| Survival | The queue MUST survive transient transport failure, HTTP auth failure on a queued write, and `session_revoked` within the same browser runtime. |
| Non-survival | The base profile MUST NOT rely on queue survival across full page reload, tab close, browser restart, cross-tab transfer, or tab crash. |
| Cross-tab behavior | Queue contents MUST NOT be shared across tabs or client instances in the base profile. |
| Overflow behavior | On capacity overflow, the client MUST keep all already queued units, refuse admission of the new unit, preserve the current visible edit as unsaved local work, set save state to `Conflict`, and show a same-surface non-modal overflow message. It MUST NOT silently evict queued units or reorder to make room. |
| Allowed coalescing | A still-uncommitted local row may fold one queued create plus later unsent edits to that same local row into one create unit until the first authoritative create succeeds. Existing-row unsent patch units for the same `record_id` MAY coalesce only within one contiguous same-record run; the coalesced unit preserves final direct-write value per `field_key` and declared order of any `collection_actions_v1.actions[]`. |
| Forbidden coalescing | The client MUST NOT coalesce across record boundaries, intervening queued units for another record, destructive actions, conflict-resolution actions, or non-hot-path operations. An interleaving such as `A1, B1, A2` MUST NOT replay as reordered `A1+A2, B1`. |
| Non-retryable halt | Replay MUST stop at the first non-retryable failure that requires analyst action. Later queued units remain queued and unapplied behind the blocked unit. |

### Table C. Save-State Derivation

| Input condition | Primary label | Secondary same-surface message | Closure rule |
| --- | --- | --- | --- |
| Any unresolved same-field conflict exists. | `Conflict` | Conflict detail anchored by `record_id:field_key`; resolver UI remains out of FE-P4 scope. | `Conflict` wins over `Syncing` and `Saved`. |
| Queue overflow refused admission of a replay unit. | `Conflict` | Overflow message explaining the current visible edit remains unsaved local work. | Must reference Table B overflow behavior. |
| Replay halted on a non-retryable failure that requires analyst action. | `Conflict` | Public failure detail from the error envelope without private server details. | Later queued units remain queued and unapplied. |
| At least one workbook mutation is in flight or the local pending queue is non-empty, including paused replay waiting for connectivity recovery, re-authentication, or HTTP re-query. | `Syncing` | Optional queue/replay detail when needed for browser evidence or accessibility communication. | Ambient presence or collaboration state MUST NOT change this label. |
| No workbook mutation is in flight, local pending queue is empty, and no unresolved same-field local drafts exist. | `Saved` | Optional recent-success detail only if it remains on the same workbook surface. | Presence updates alone MUST NOT move the label away from `Saved`. |

Primary labels are exactly `Syncing`, `Saved`, and `Conflict`. `Failed`, `Queued`, `Pending`, `Retrying`, and `Replay halted` MAY appear only as secondary same-surface detail or test/internal state names; they MUST NOT be primary save-state labels.

### Table D. Public Mutation Route Matrix

| Operation | Route | Required request members | Idempotency key | First success | Exact replay | Same key, different normalized request | Stale version behavior | Returned row shape |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Row create | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` | JSON object with required `client_txn_id`; additional top-level members only when each member name is a writable `field_key` allowed for create by the addressed view contract. | `(actor_user_id, incident_id, view_schema_id, client_txn_id)` | `201 Created` with `data.view_schema_id`, `data.change_set_id`, and `data.row`. | `200 OK` with the originally committed create result, not current mutable row state. | `409` with `error.code = client_txn_conflict` and `error.details.client_txn_id` at minimum. | Not applicable to create. | `data.row` is exactly full `view_row_v1`; `record_id` and `row_version` are carried only inside `data.row`. |
| Patch | `PATCH /api/v1/records/{record_id}` | JSON object with required `view_schema_id`, `base_row_version`, `client_txn_id`, and non-empty `changes[]`. Each `changes[]` entry contains `field_key` and exactly one of `value` or `action_payload`. | `(actor_user_id, record_id, client_txn_id)`; `view_schema_id`, `base_row_version`, and canonical `changes[]` participate in normalized replay comparison. | `200 OK` with `data.view_schema_id`, `data.change_set_id`, and `data.row`. | `200 OK` with the originally committed patch result before fresh optimistic-concurrency evaluation. | `409` with `error.code = client_txn_conflict`; this wins before fresh optimistic-concurrency evaluation. | With no prior committed idempotency hit, non-overlapping writable fields auto-rebase, overlapping writable fields fail with `same_field_conflict`, and missing or unusable revision history fails closed with `row_version_conflict`. | `data.row` is exactly full `view_row_v1`; `record_id` and `row_version` are carried only inside `data.row`. |

Patch limits and canonicalization:

- `changes[]` is required and non-empty; `changes[]: []` and explicit null are invalid mutation payload.
- Raw parsed `changes[]` length MUST be at most `32`.
- Duplicate `field_key` entries are invalid and MUST NOT be normalized away.
- Outer `changes[]` order is non-semantic and canonicalized by `field_key asc`.
- `collection_actions_v1.actions[]` is ordered inside its field payload; empty `actions[]` is invalid and raw parsed length MUST be at most `64`.
- Count-limit failures are evaluated before idempotency replay comparison or write execution.

### Table E. Query And `view_row_v1` Omission Semantics

| Surface | Omission or boundary rule |
| --- | --- |
| `sort` | Omitted `sort` or `sort: []` means no user sort override. When present, `sort[]` contains at most `8` raw entries. Duplicate normalized `field_key` entries are invalid. Effective applied sort appends the remaining schema default-sort tail and `record_id asc` when absent. |
| `filters` | Omitted `filters` or `filters: []` means no filters. Request order is non-semantic. Raw `filters[]` length is at most `16`. |
| `group_by` | Omitted `group_by` means grouping inactive. `group_by: null` is invalid. The current profile allows at most one active grouping key. |
| Pagination members | `limit` and `cursor_token` appear only as JSON-body members for the view-query route, not query parameters. `limit` counts serialized `rows[]` entries only. |
| `meta.query` | Successful view-query responses MUST include `meta.query`. `meta.query.sort[]` is the effective applied sort after default-tail expansion. `meta.query.group_by` is omitted when grouping is inactive and never serialized as JSON null. |
| Full row cells | Every `rows[]` entry is full `view_row_v1`. `rows[].cells` MUST include every schema-declared non-technical field for the active `view_schema_id`, regardless of visibility, default-hidden state, writability, or read-only state. |
| Authoritative null | A schema-declared non-technical field serialized as `{ "value": null }` means authoritative null when the field admits null. Omission of that field is invalid. |
| Technical fields | `record_id` and `row_version` are top-level technical identifiers and MUST NOT be duplicated inside `cells`. |
| Group values | If the schema declares grouping keys, the full row includes the full current `group_values` object. If the schema declares no grouping keys, `group_values` is omitted. |
| Unknown additive members | Clients MAY ignore unknown additive members inside row or cell objects only where Core allows additive unknown members. Missing required members MUST fail or render a public error; they MUST NOT be silently replaced by blanks. |

### Table F. Clipboard Paste Boundary

| Paste boundary | Closed FE-P4 rule |
| --- | --- |
| Base-profile scope | Clipboard paste remains base hot-path ingest. It MUST NOT be used to claim file-based structured import support. |
| Default Ctrl+V dispatch | Interactive tabular dispatch requires an unambiguous tabular signal: tab, newline, carriage return, or a future explicit paste-as-table command. |
| Single-line comma text | A single-line comma-only `text/plain` payload such as `Hello, world` is scalar text by default, not default tabular CSV. |
| Planning identity | Base-profile clipboard paste derives its row plan from active `view_schema_id`, stable `field_key` columns, source-column ordinals, and declared `entity_binding_mode`. It MUST NOT use surface-local header heuristics as authoritative identity. |
| Existing-row targets | Every record target must belong to the addressed incident and active workbook surface. Target ownership and visibility are validated before row-version comparison, conflict construction, batch commit, or response row serialization. |
| Rejected batch targets | A paste containing a missing, foreign-incident, wrong-surface, wrong-type, or deleted record target MUST fail closed as one rejected batch rather than partially committing other targets. |
| Successful non-conflicting writes | Successful non-conflicting writes from one paste action appear as one visible `change_set`, ordered mutation entries, and one row revision per affected record. |
| Same-field conflicts | Same-field conflicts from the paste remain outside the committed non-conflicting batch until explicit resolution. Each later same-field conflict resolution creates its own attributed `change_set`. |

### Table G. Frontend Error-State Mapping

| Public condition | Frontend anchor | Queue behavior | Required user-visible behavior |
| --- | --- | --- | --- |
| `invalid_view_query` | Query/surface level, with `error.details` when available. | No mutation replay. | Render same-surface public error and keep private server details hidden. |
| `invalid_mutation_payload` | Cell-level when `error.details.field`, `field_key`, or equivalent member identifies a cell; otherwise row or mutation level. | Non-retryable; halt the affected replay unit when it is queued. | Preserve local unsaved work when possible and show same-surface validation detail. |
| `client_txn_conflict` | Mutation level keyed by `client_txn_id` plus the route scope. | Non-retryable; keep later units queued behind the blocked unit. | Show same-surface conflict/error detail without treating it as successful replay. |
| `same_field_conflict` | Cell conflict keyed by `record_id:field_key`. | Blocked unit leaves the local pending queue and enters the client-local same-field conflict queue; later units remain queued and unapplied. | Primary save-state label is `Conflict`; resolver implementation remains outside FE-P4. |
| `row_version_conflict` | Record or cell level according to `error.details`. | Non-retryable unless Core field-level rules auto-rebase the request before failure; otherwise later units remain queued behind the blocked unit. | Show same-surface conflict/error detail and do not retarget by visible row index. |
| Auth failure, `session_revoked`, or transient transport disconnect on queued write | Queue/runtime level scoped to the same browser runtime. | Retryable after re-authentication, current incident authorization re-derivation, and any required HTTP re-query. | Preserve queued writes and unresolved same-field local drafts within the same browser runtime. |
| Unknown public error code | Surface, row, or mutation level using any stable public details available. | Non-retryable unless public `error.retryable=true`; retryable unknown errors still obey FIFO queue order. | Render a same-surface non-private public error; never silently discard the pending unit. |

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers |
| --- | --- | --- | --- |
| [x] | 1. Readiness, map, ledger, FE-P3 handoff | `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P4`; registry and row-inventory checks; `make phase-ledger-drift`; `git diff --check` | Complete as readiness evidence only; at Sprint 1 handoff FE-P4 remained planned, non-executable, and all rows were blocked |
| [x] | 2. Sync-engine pending queue unit model for `FE-U-P4-01` | `make frontend-unit`; `make frontend-typecheck` | Complete as Sprint 2 only; does not close save-state, browser public-route, visual, accessibility, or full FE-P4 phase readiness |
| [x] | 3. Save-state derivation and status strip unit model for `FE-U-P4-02` | `make frontend-unit`; `make frontend-typecheck` | Complete as Sprint 3 only; does not close public-route replay, visual, accessibility, conflict resolution, or full FE-P4 phase readiness |
| [x] | 4. Timeline query/render identity integration for `FE-I-P4-01` | `make frontend-unit`; `make browser-e2e-webserver-backed` | Complete as Sprint 4 only; does not close public-route replay, rough-create/paste hot path, visual, accessibility, saved-view persistence, WebSocket live updates, or full FE-P4 phase readiness |
| [ ] | 5. Public-route E2E for rough create, edit, paste, pending save, refresh, and replay for `FE-E-P4-01` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Block if rough create, inline edit, paste, replay, or validation evidence bypasses Tables D, F, and G public-boundary behavior |
| [ ] | 6. Visual and accessibility readiness for `FE-V-P4-01` and `FE-A11Y-P4-01` | `make browser-e2e-visual`; current `make browser-e2e-a11y-preflight`; implemented closure requires `make browser-e2e-a11y` mapping | FE-V row is blocked until recaptured; FE-A11Y row is preflight-only and blocks full FE-P4 phase completion until implemented evidence exists |
| [ ] | 7. Closure, drift, final validation, and FE-P5 handoff | Row-owned targets plus `make frontend-import-boundary-check`, `make generated-artifact-policy-check`, `make generate-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`, `make agent-finalize`, and `make check` when required | Block if any row remains blocked/stale for full phase completion; product hot-path closure must be named separately when design rows remain blocked |

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

Sprint 1 execution record, 2026-06-01:

- Passed: `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P4`; FE-P4 reported as planned, explainable, and non-executable.
- Passed: `make task-guide ROLE=feature-dev PHASE_NAMESPACE=frontend PHASE=FE-P4`; the planned-phase guide recommended `explain-phase` and `phase-ledger-drift`.
- Passed: registry tuple inspection; FE-P4 is in the frontend namespace with the expected manifest path, ledger path, dependency chain, activation blocker, and freshness digests.
- Passed: Sprint 1 row inventory inspection; FE-P4 had exactly `FE-U-P4-01`, `FE-U-P4-02`, `FE-I-P4-01`, `FE-E-P4-01`, `FE-V-P4-01`, and `FE-A11Y-P4-01` once each, all `claim_status="blocked"` before later row promotions.
- Passed: `make json-shape-check`, run root `.cartulary/test-results/20260601T214448Z-p61340`, summary `json-shape-check/tool-run-summary.json`.
- Passed: `make generated-artifact-policy-check`, run root `.cartulary/test-results/20260601T214453Z-p61529`, summary `generated-artifact-policy-check/tool-run-summary.json`.
- Passed: `make phase-ledger-drift`, run root `.cartulary/test-results/20260601T214457Z-p62021`, summary `phase-ledger-drift/tool-run-summary.json`.
- Passed: `make explain-run RESULTS_DIR=.cartulary/test-results/20260601T214457Z-p62021 TARGET=phase-ledger-drift`.
- Inspected target surface with `make help`, `make help-all`, and `make explain-target TARGET=<target> DETAIL=summary` for `frontend-unit`, `browser-e2e-webserver-backed`, `browser-e2e-stateful`, `browser-e2e-visual`, `browser-e2e-a11y-preflight`, `browser-e2e-a11y`, `browser-e2e-support`, and `phase-ledger-drift`.

Sprint 1 handoff state:

- At Sprint 1 handoff, FE-P4 remained `planned`, non-executable, and all six FE-P4 rows remained `blocked`.
- FE-P3 handoff context and retained roots are historical only unless rerun under the current target and row-accounting rules.
- `FE-A11Y-P4-01` remains blocked because the current `browser-e2e-a11y-preflight` and `browser-e2e-a11y` target summaries report `phase_coverage: none`.
- `FE-V-P4-01` remains blocked because fixture registry `current` status does not equal FE-P4 row-owned recapture evidence.
- No Core 05 claim-publication predicate was introduced.
- No FE-P4 product behavior, row promotion, generated-ledger edit, registry/map metadata change, or phase activation was performed.

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
- Core 03 Sections 4.1, 4.4, and 15 own autosave, local pending queue, and Timeline read/write interaction behavior.
- The sync engine must be unit-testable without private server behavior.

Test-first sequence:

- Add failing unit tests that cover row-create units, row-patch units, paste-derived create and patch replay units, retryable public failures, success replay, validation failures, and non-retryable replay halt.
- Assert Table B capacity exactly: 64 non-coalescible units are admitted for one `(incident_id, client_instance_id)`, and a 65th non-coalescible unit is refused without evicting or reordering the first 64.
- Assert overflow preserves the current visible edit as unsaved local work, sets primary save-state label input to `Conflict`, and emits same-surface overflow detail.
- Assert replay is FIFO by original enqueue order and never reorders by visible row order, sort order, record type, labels, or presentation position.
- Assert base-profile non-survival boundaries: a full reload or recreated page instance does not restore or silently replay the local pending queue.
- Assert allowed coalescing for a still-uncommitted local row and one contiguous same-record patch run, and forbidden coalescing across an interleaving such as `A1, B1, A2`.
- Assert create identity uses `client_txn_id` plus create route scope; assert patch identity uses `record_id`, `client_txn_id`, `view_schema_id`, `base_row_version`, and canonical `changes[]`.
- Assert non-retryable validation, `client_txn_conflict`, `same_field_conflict`, and terminal `row_version_conflict` follow Table G queue behavior.

Implementation tasks:

- Model queue entries for row create, row patch, and paste-derived create or patch units.
- Preserve admission order and dispatch order deterministically.
- Associate every queue entry with the Table D route shape and route-scoped idempotency identity.
- Distinguish pending, dispatching, retryable failure, validation failure, conflict, overflow, replay-halted, and success outcomes without turning non-Core state names into primary save-state labels.
- Apply successful row refreshes without retargeting pending entries by visible row index.
- Surface validation failures as cell-level state for later rendering work.
- Preserve Table B same-runtime survival and base-profile non-survival boundaries.
- Keep same-field conflict resolver implementation out of scope; preserve only conflict anchoring.

Validation commands:

- `make frontend-unit`
- `make frontend-typecheck`
- `make generated-artifact-policy-check` and `make generate-drift` if generated protocol consumption changes
- `make phase-ledgers` and `make phase-ledger-drift` only after row metadata is intentionally promoted

Evidence requirements:

- `frontend-unit` must emit current frontend row accounting that closes `FE-U-P4-01` when the row is promoted.
- Unit evidence must use public route-shaped success/error results, not private server internals.
- Before promotion, the row must remain `claim_status="blocked"` until direct current evidence exists.

Sprint 2 execution record, 2026-06-02:

- Implemented the shared pending-queue semantic owner in `workbookPendingQueue.ts` with peek, mark-dispatched, settlement, auth pause/resume, same-field conflict clearing, and public pending-error parsing.
- Converted `WorkbookShell` pending replay to a model-backed runtime adapter with shell-only replay metadata keyed by queue unit ID.
- Added row-owned model and shell runtime scenarios for the nine mapped `FE-U-P4-01` titles.
- Promoted only `FE-U-P4-01` to `claim_status="implemented"` and regenerated the FE-P4 coverage ledger with `make phase-ledgers`.
- Current `frontend-unit` row accounting closes `FE-U-P4-01`; all other FE-P4 rows remain blocked or outside Sprint 2 scope.

Blocker rules:

- `BLOCKER: FE-U-P4-01 pending queue evidence missing Table B boundary coverage; missing=<capacity|overflow|FIFO|coalescing|non_survival|non_retryable_halt> minimum_follow_up=add unit coverage for the missing closed boundary.`
- `BLOCKER: FE-U-P4-01 pending queue evidence missing route-scoped mutation identity; minimum_follow_up=add unit coverage for create identity client_txn_id plus route scope and patch identity record_id, client_txn_id, view_schema_id, base_row_version, and canonical changes[].`
- `BLOCKER: FE-U-P4-01 product evidence uses visible row order or labels as mutation identity; minimum_follow_up=replace with record_id, field_key, base_row_version, view_schema_id, and client_txn_id as owner contracts require.`
- `BLOCKER: FE-U-P4-01 target passed but frontend row accounting did not close row; target=frontend-unit failure_reason=frontend_row_accounting.`

Binary acceptance:

- `FE-U-P4-01` is implemented only when current `make frontend-unit` evidence closes the row, every Table B and Table G queue boundary required by Sprint 2 is covered, and the FE-P4 map/ledger are updated through the proper generator flow.

Explicit non-claims:

- Sprint 2 does not prove browser public-route behavior.
- Sprint 2 does not prove visual or accessibility behavior.
- Sprint 2 does not implement the same-field conflict resolver.

## Sprint 3: Save-State Derivation And Status Strip Unit Model

Objective: implement save-state derivation for `FE-U-P4-02`.

Owned row: `FE-U-P4-02`.

Non-owned rows: `FE-U-P4-01`, `FE-I-P4-01`, `FE-E-P4-01`, `FE-V-P4-01`, and `FE-A11Y-P4-01`.

Source constraints:

- Core 03 Sections 4.1, 4.2, and 4.4 own autosave, save-state presentation, and local pending queue behavior.
- UI/UX guide and design docs contribute design-direction only and must not widen product-conformance claims.

Test-first sequence:

- Add unit tests for every Table C input condition: unresolved same-field conflict, queue overflow, replay halted on non-retryable failure, mutation in flight, non-empty local pending queue, paused replay, fully saved state, and ambient presence-only updates.
- Assert exactly one primary label and assert the label is always one of `Syncing`, `Saved`, or `Conflict`.
- Assert `Failed`, `Queued`, `Pending`, `Retrying`, and `Replay halted` never appear as primary save-state labels.
- Assert exactly one secondary message on the same surface when a secondary message is needed.
- Assert conflict state remains compatible with `record_id + field_key + base_row_version` anchoring.

Implementation tasks:

- Build or harden a pure save-state derivation function.
- Feed save-state derivation from pending queue and conflict/error state rather than DOM or visible labels.
- Preserve Table C precedence: `Conflict` wins over `Syncing`, and `Saved` is emitted only when no mutation is in flight, the queue is empty, and no unresolved same-field local drafts exist.
- Treat failed, overflow, validation, and replay-halted detail as secondary same-surface messages, not primary labels.
- Keep validation and conflict text non-color-only and ready for a11y work.

Validation commands:

- `make frontend-unit`
- `make frontend-typecheck`
- `make browser-e2e-visual` later in Sprint 6 for design-direction visual capture
- `make phase-ledgers` and `make phase-ledger-drift` only after row metadata is intentionally promoted

Evidence requirements:

- `frontend-unit` must close `FE-U-P4-02` through current frontend row accounting before promotion.
- Design-direction status-strip evidence cannot replace product-conformance unit evidence.

Sprint 3 execution record, 2026-06-02:

- Implemented the pure save-state derivation model in `workbookPendingQueue.ts` with exact primary labels `Syncing`, `Saved`, and `Conflict`, Table C precedence, conflict anchors, and same-surface secondary detail.
- Wired `WorkbookShell` status-strip presentation to the derived model rather than DOM labels or visible-row state.
- Added row-owned model coverage for every Table C input condition and status-strip rendering coverage for same-surface secondary detail.
- Promoted only `FE-U-P4-02` to `claim_status="implemented"` and kept generated ledger freshness aligned through the authored map/generator flow.
- Current `frontend-unit` row accounting closes `FE-U-P4-02`; `FE-I-P4-01`, `FE-E-P4-01`, `FE-V-P4-01`, `FE-A11Y-P4-01`, and FE-P4 phase activation remain blocked.

Blocker rules:

- `BLOCKER: FE-U-P4-02 save-state derivation emits more than one primary label or a non-Core primary label; minimum_follow_up=normalize Table C precedence into exactly one of Syncing, Saved, or Conflict.`
- `BLOCKER: FE-U-P4-02 save-state secondary message is detached from the same workbook surface; minimum_follow_up=render or derive secondary message in the status strip surface.`
- `BLOCKER: FE-P4 evidence classes collapsed; design/support/claim-publication-boundary evidence cannot be counted as product_conformance.`

Binary acceptance:

- `FE-U-P4-02` is implemented only when current `make frontend-unit` row accounting closes it, every Table C derivation row is covered, no non-Core primary save-state label is normative, and the generated ledger is refreshed from the authored map.

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
- Assert Table E query omission semantics: omitted `sort` and `sort: []`, omitted `filters` and `filters: []`, omitted `group_by`, invalid `group_by: null`, required `meta.query`, and effective sort/default-tail behavior.
- Assert hidden and default-hidden schema-declared non-technical fields remain present in full `view_row_v1.cells`.
- Assert unknown additive row or cell members are ignored only where Core allows additive unknown members; missing required members fail or render a public error.

Implementation tasks:

- Feed Timeline query rows through `/packages/view-contracts` and `/packages/grid-adapter`.
- Preserve `record_id`, `row_version`, `view_schema_id`, and `field_key` through render, refresh, validation, and retry paths.
- Ensure create and patch refreshes merge by `record_id`.
- Ensure validation errors anchor to stable cell identity.
- Keep presentation rows non-mutating.
- Treat `record_id` and `row_version` as top-level technical identifiers, never as editable or queryable `cells` entries.
- Treat `{ "value": null }` as authoritative null and omitted schema-declared non-technical cells as invalid.

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
- `BLOCKER: FE-I-P4-01 query omission semantics incomplete; missing=<sort|filters|group_by|meta.query|hidden_fields|null_vs_omission> minimum_follow_up=add Table E coverage.`
- `BLOCKER: FE-I-P4-01 row identity retargeted by visible index after refresh; minimum_follow_up=anchor refresh and validation state by record_id and field_key.`

Binary acceptance:

- `FE-I-P4-01` is implemented only when both mapped layers required by the current map and harness close with current row accounting and Table E query and full-row boundaries are covered.

Sprint 4 execution record, 2026-06-02:

- Started with discovery of the current FE-P4 map, row inventory, exact scenario title, row accounting, and target mapping for `FE-I-P4-01`; both `frontend-unit` and `browser-e2e-webserver-backed` were confirmed required for row closure.
- Added failing test-first coverage before product edits. Initial `make frontend-unit` failed at run root `.cartulary/test-results/20260602T224252Z-p61355` because the Timeline row fixture omitted a schema-declared non-technical cell.
- Added unit/integration coverage in `apps/web/src/WorkbookShell.phase4.timelineQuery.test.tsx` with the exact scenario title `FE-I-P4-01 Verify Timeline query response rows render full view_row_v1 cells and preserve row identity through create, patch, validation error, and refresh.`
- Added browser-backed coverage in `apps/web/e2e/frontend.phase4.timeline-query.spec.ts` with the same exact scenario title.
- Implemented `view_row_v1` interpretation in `packages/view-contracts/src/index.ts` for full rows and sparse patch rows, preserving `record_id`, `row_version`, `view_schema_id`, `field_key`, authoritative null values, and additive-member behavior only where allowed.
- Routed Timeline query rows, mutation-success rows, refresh rows, preview/support rows, and record-changed patch rows in `apps/web/src/WorkbookShell.tsx` through the view-contract boundary before rendering or reconciliation.
- Reconciled refreshed Timeline rows by `record_id`, preserved high-water `row_version`, retained validation and pending local state by `record_id` plus `field_key`, and replaced draft-created rows with server-returned `record_id`.
- Updated authored FE-P4 map metadata for `FE-I-P4-01`, regenerated the FE-P4 coverage ledger through `make phase-ledgers`, and refreshed FE-P4 registry digests required by `json-shape-check`.

Sprint 4 validation record, 2026-06-02:

- Passed: `make frontend-typecheck`, run root `.cartulary/test-results/20260602T230401Z-p132932`, summary `frontend-typecheck/tool-run-summary.json`.
- Passed: `make frontend-unit`, run root `.cartulary/test-results/20260602T230410Z-p133356`, summary `frontend-unit/tool-run-summary.json`.
- Passed: `make browser-e2e-webserver-backed`, run root `.cartulary/test-results/20260602T230428Z-p134964`, summary `browser-e2e-webserver-backed/tool-run-summary.json`.
- Passed: `make phase-ledgers`, run root `.cartulary/test-results/20260602T225425Z-p95709`, summary `phase-ledgers/tool-run-summary.json`.
- Passed: `make phase-ledger-drift`, run root `.cartulary/test-results/20260602T230555Z-p143317`, summary `phase-ledger-drift/tool-run-summary.json`.
- Passed: `make json-shape-check`, run root `.cartulary/test-results/20260602T230601Z-p143632`, summary `json-shape-check/tool-run-summary.json`.
- Passed: `make agent-finalize`, run root `.cartulary/test-results/20260602T230619Z-p144653`, summary `agent-finalize/tool-run-summary.json`; retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- Skipped: `make browser-e2e-support`; Sprint 4 added a browser spec but did not change shared browser helpers, selector contracts, or shared choreography.
- Diagnostic only: a scoped Biome check for Sprint 4 touched files passed; full `make lint-biome` still failed on existing untouched import-order/format findings in `apps/web/src/WorkbookShell.phase4.saveState.test.tsx` and `apps/web/src/workbookPendingQueue.test.ts`.

Sprint 4 row-accounting outcome, 2026-06-02:

- `FE-I-P4-01` is closed for `frontend-unit` in `.cartulary/test-results/20260602T230410Z-p133356/frontend-unit/frontend-row-accounting.json` with the exact row-owned scenario title.
- `FE-I-P4-01` is closed for `browser-e2e-webserver-backed` in `.cartulary/test-results/20260602T230428Z-p134964/browser-e2e-webserver-backed/frontend-row-accounting.json` with the exact row-owned scenario title.
- `FE-I-P4-01` is implemented from current row-owned evidence.

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
- Rough create requirement: create a rough Timeline row through `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`, include required `client_txn_id`, use only create-time `field_key` members allowed by the active view contract, preserve rough input, and verify forbidden, read-only, or server-managed fields fail closed through the public envelope.
- Inline edit requirement: patch a writable cell through `PATCH /api/v1/records/{record_id}` with non-empty `changes[]`, required `view_schema_id`, required `base_row_version`, required `client_txn_id`, max-`32` change boundary coverage, duplicate-`field_key` rejection, and original committed replay response for exact idempotent replay.
- Save-state requirement: observe Table C primary labels through public-route behavior, including `Syncing` while work is in flight or queued and `Saved` only when the queue is empty and no unresolved local drafts remain.
- Paste requirement: paste deterministic scalar and tabular content to prove Table F scalar-vs-tabular dispatch, target validation before commit, non-conflicting commit behavior, same-field conflict grouping, and public-envelope rendering.
- Induce a transient failure only through an accepted harness-owned public test control or service-boundary behavior. If only private frontend mocks are available, block the row.
- Replay requirement: verify transient transport/auth/session failure preserves queued work within the same browser runtime, re-auth/re-query occurs when required, replay remains FIFO, and a full reload or recreated page instance does not silently restore or replay the base-profile queue.
- Refresh requirement: refresh and verify stable row identity by `record_id` plus cell identity by `field_key`; pending work must not retarget by visible row order.
- Validation requirement: verify Table G public error-state mapping for cell-level validation, mutation-level conflict, same-field conflict, row-version conflict, and unknown public error fallback without private server details.

Implementation tasks:

- Wire FE-P4 browser hot path to `/api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, row create, patch, and clipboard-paste route contracts.
- Preserve `client_txn_id`, `record_id`, `base_row_version`, and `field_key` through replay.
- Keep rough capture valid without mention resolution or evidence handles.
- Prevent replay from applying later units out of order after non-retryable failure.
- Keep validation and conflict state on the same workbook surface.
- Keep create and patch request construction within Table D limits and omission semantics.
- Keep paste dispatch and batch target validation within Table F.

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
- Browser evidence must separately exercise rough create, inline edit, paste, replay, refresh, and validation requirements, even if they are implemented in one scenario file.

Blocker rules:

- `BLOCKER: FE-P4 public-boundary behavior was tested through frontend-only mocks; product-conformance route evidence requires public /api/v1/ browser-facing evidence.`
- `BLOCKER: FE-E-P4-01 replay cannot be induced through a harness-owned public boundary; minimum_follow_up=add accepted test control or record row as blocked.`
- `BLOCKER: FE-E-P4-01 rough create route contract incomplete; missing=<client_txn_id|allowed_field_keys|forbidden_field_rejection|rough_input_preservation> minimum_follow_up=add Table D public-create coverage.`
- `BLOCKER: FE-E-P4-01 inline edit route contract incomplete; missing=<non_empty_changes|max_32|duplicate_field_key|original_committed_replay> minimum_follow_up=add Table D public-patch coverage.`
- `BLOCKER: FE-E-P4-01 clipboard paste boundary incomplete; missing=<scalar_vs_tabular|target_validation|non_conflicting_commit|same_field_conflict_grouping> minimum_follow_up=add Table F browser coverage.`
- `BLOCKER: FE-E-P4-01 public error-state mapping incomplete; missing=<invalid_mutation_payload|client_txn_conflict|same_field_conflict|row_version_conflict|unknown_public_error> minimum_follow_up=add Table G browser coverage.`
- `BLOCKER: FE-E-P4-01 browser target passed but row accounting is missing or stale; target=<target> failure_reason=frontend_row_accounting.`

Binary acceptance:

- `FE-E-P4-01` is implemented only when current browser row accounting closes the row through the mapped webserver-backed and stateful targets, every Sprint 5 requirement above is covered against Tables C, D, F, and G, or the current map is corrected with owner-approved target semantics before promotion.

Execution note (2026-06-02):

- Added focused exact-title browser coverage in `apps/web/e2e/frontend.phase4.public-route.spec.ts` for the real public-route create, patch, paste, save-state, refresh, and replay paths that can be induced through the current harness.
- Kept `FE-E-P4-01` blocked because the unknown public error fallback cannot currently be induced through a real public route, accepted harness-owned control, or service-boundary behavior without private frontend mocks:
  `BLOCKER: FE-E-P4-01 public error-state mapping incomplete; missing=unknown_public_error minimum_follow_up=add Table G browser coverage.`
- No FE-P4 phase, visual, accessibility, WebSocket, same-field resolver UI, Core 05, or product hot-path closure is claimed.

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
- Resolve pending replay visual coverage by exact fixture linkage to `FE-VFIX-08`, `FE-VFIX-12`, or `FE-VFIX-15`; if no exact linkage exists, record a fixture-registry or phase-map blocker before any visual readiness claim.
- Add or update exact visual scenario title from the FE-P4 map.
- Run `make browser-e2e-visual`.
- Run current `make browser-e2e-a11y-preflight` only as blocked-row smoke.
- If the row is promoted to implemented accessibility evidence, update the map to `make browser-e2e-a11y` and require `cartulary.frontend_accessibility_summary.v2`.
- If `FE-A11Y-P4-01` remains mapped only to `browser-e2e-a11y-preflight`, record it as a blocker to `FE-P4 phase complete` under Table A. Product hot-path closure may still be reported separately when product rows close.

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
- `FE-A11Y-P4-01` must remain blocked while mapped only to preflight. Preflight artifacts may support diagnostics but not accessibility readiness closure or FE-P4 phase completion.
- Visual and accessibility evidence must not be promoted into product conformance.
- Visual readiness closure and accessibility readiness closure are separate Table A claims and must be named separately in handoff text.

Blocker rules:

- `BLOCKER: FE-V-P4-01 fixture identity missing or ambiguous; expected exact FE-VFIX-08, FE-VFIX-12, and FE-VFIX-15 linkage or precise map/registry correction.`
- `BLOCKER: FE-V-P4-01 visual fixture is current in registry but not recaptured as FE-P4 row-owned evidence; target=browser-e2e-visual minimum_follow_up=run exact FE-P4 scenario with frontend row accounting.`
- `BLOCKER: FE-P4 accessibility row is preflight-only and cannot be counted as accessibility readiness closure or FE-P4 phase completion; row=FE-A11Y-P4-01 target=browser-e2e-a11y-preflight minimum_follow_up=map implemented scenario to browser-e2e-a11y with cartulary.frontend_accessibility_summary.v2 or keep row blocked.`
- `BLOCKER: FE-V-P4-01 pending replay visual coverage unresolved; expected exact fixture linkage or owner-approved map/registry correction; minimum_follow_up=resolve pending replay fixture identity before visual readiness claim.`

Binary acceptance:

- `FE-V-P4-01` is implemented only when direct visual row accounting closes it.
- `FE-A11Y-P4-01` is not implemented under the current preflight-only closure posture, and full `FE-P4 phase complete` remains blocked until implemented accessibility evidence exists.

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
- Confirm design-direction rows are closed by intended current design evidence before any `FE-P4 phase complete` claim.
- If a design-direction row remains blocked, record only the narrower handoff claim `FE-P4 product hot-path closed; FE-P4 phase completion blocked by <row/blocker>`.
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

- `BLOCKER: FE-P4 row remains blocked for phase completion; row=<row_id> blocker=<reason_code> closure_claim=<product_hot_path|visual_readiness|accessibility_readiness|phase_complete> minimum_follow_up=<specific target or owner patch>.`
- `BLOCKER: FE-P4 generated ledger is stale relative to map; rerun generator only after confirming authored map is intended source.`
- `BLOCKER: FE-P4 earlier active phase regression failed; phase=<FE-P0..FE-P3> target=<target> run_root=<run_root> minimum_follow_up=<fix or owner acceptance>.`
- `BLOCKER: FE-P4 evidence freshness digest stale; minimum_follow_up=rerun freshness/ledger validation after map, registry, fixture, and target evidence are final.`
- `BLOCKER: FE-P4 product hot-path closure cannot be reported as phase completion; blocked_design_rows=<rows> minimum_follow_up=close design readiness rows or use the exact product-hot-path handoff wording.`

Binary acceptance:

- `product hot-path closed` is allowed only when `FE-U-P4-01`, `FE-U-P4-02`, `FE-I-P4-01`, and `FE-E-P4-01` close through current row-owned evidence.
- `FE-P4 phase complete` is allowed only when all six FE-P4 rows close through current row-owned evidence.
- Shared harnesses are satisfied or have precise blocking entries that prevent the affected Table A closure claim.
- FE-P0 through FE-P3 remain green or have precise owner-accepted blockers that do not invalidate the specific Table A closure claim being made.
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
- the exact Table A closure claim it blocks.

Required blocker language:

- `BLOCKER: FE-P4 map row inventory invalid; expected exactly FE-U-P4-01, FE-U-P4-02, FE-I-P4-01, FE-E-P4-01, FE-V-P4-01, FE-A11Y-P4-01 once each; actual=<ids/counts>.`
- `BLOCKER: FE-P4 registry tuple is not frontend-namespace traceable; expected namespace/path/dependency tuple does not match registry.`
- `BLOCKER: FE-P4 generated ledger is stale relative to map; rerun generator only after confirming the authored map is the intended source.`
- `BLOCKER: FE-P4 owner refs unresolved; row=<row_id> owner_ref=<source/section/req/ac> minimum_follow_up=<specific inspection or owner patch>.`
- `BLOCKER: FE-P4 browser row lacks exact scenario_titles[] for row-owned closure; row=<row_id> target=<target>.`
- `BLOCKER: FE-P4 accessibility row is preflight-only and cannot be counted as accessibility readiness closure or FE-P4 phase completion; row=FE-A11Y-P4-01 target=browser-e2e-a11y-preflight minimum_follow_up=<required implemented-row a11y target or map status correction>.`
- `BLOCKER: FE-P4 evidence classes collapsed; design/support/claim-publication-boundary evidence cannot be counted as product_conformance.`
- `BLOCKER: FE-P4 public-boundary behavior was tested through frontend-only mocks; product-conformance route evidence requires public /api/v1/ browser-facing evidence.`

Strict non-claims:

- Do not claim FE-P4 phase completion from generated ledgers, old retained artifacts, broad `make check`, test names, visual goldens, support-only tests, accessibility preflight smoke, or this plan.
- Do not claim FE-P4 phase completion when any of the six FE-P4 rows remains `blocked`, `stale`, or missing row accounting. Use `FE-P4 product hot-path closed; FE-P4 phase completion blocked by <row/blocker>` only when the product rows close and design rows remain blocked.
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
- all six FE-P4 rows close from direct current row-owned evidence;
- all product-conformance rows have Core 00 through Core 04 or adopted NLSpec ownership with resolved REQ and AC IDs;
- all design-direction rows remain design-direction only;
- every triggered shared harness is satisfied or blocks the affected Table A closure claim;
- generated ledgers and schedules are regenerated through Make and drift checks pass;
- generated-artifact policy and generated drift checks pass when generated or contract surfaces are touched;
- earlier active frontend phases remain green or precise owner-accepted blockers are recorded;
- Core/design/support/claim-publication evidence classes remain separate;
- no Core 05 claim-publication evidence is claimed unless explicit claim metadata satisfies Core 05.

Product hot-path handoff is allowed only when:

- `FE-U-P4-01`, `FE-U-P4-02`, `FE-I-P4-01`, and `FE-E-P4-01` close from direct current row-owned evidence;
- the handoff uses the exact form `FE-P4 product hot-path closed; FE-P4 phase completion blocked by <row/blocker>` when any design-direction row remains blocked;
- no product hot-path handoff text claims visual readiness, accessibility readiness, full FE-P4 phase completion, or Core 05 publication evidence.

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
