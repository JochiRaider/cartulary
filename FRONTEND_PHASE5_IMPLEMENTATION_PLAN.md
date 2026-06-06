# Frontend Phase 5 Implementation Plan

## Summary

Frontend verification contract. This file is the execution roadmap, progress marker, and FE-P6 handoff aid for `FE-P5: Entity And Mention Flows`.

This plan is not product behavior authority. It does not amend Core requirements, does not define harness mechanics, does not create visual or accessibility product conformance, and does not close any FE-P5 row. No FE-P5 row may be marked complete from this plan, generated ledger text, old retained artifacts, broad target success, visual golden existence, test names, or support-only checks. FE-P5 rows close only from direct current row-owned evidence in their mapped targets.

Current FE-P5 facts from local inspection after Sprints 5-7:

- `tools/frontend_phase_registry.json` lists FE-P5 as `status="active"`, `row_rollup_state="active_green"`, with no activation blockers.
- `tools/frontend_phase_maps/fe_p5_test_map.json` contains exactly five FE-P5 rows: `FE-U-P5-01`, `FE-I-P5-01`, `FE-E-P5-01`, `FE-V-P5-01`, and `FE-A11Y-P5-01`; all five rows have `claim_status="implemented"`.
- `FE-U-P5-01`, `FE-I-P5-01`, and `FE-E-P5-01` remain product-conformance rows and close only from their mapped row-owned frontend unit, webserver-backed, and stateful evidence.
- `FE-V-P5-01` closes only from current `make browser-e2e-visual` row-owned evidence at `.cartulary/test-results/20260606T174416Z-p1391561/browser-e2e-visual/frontend-row-accounting.json`, where the row is `claim_status="implemented"`, `target_status="pass"`, and `closure_status="closed"`.
- `FE-A11Y-P5-01` closes only from current `make browser-e2e-a11y` implemented-row evidence at `.cartulary/test-results/20260606T182938Z-p1612967/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json`, where `schema_id="cartulary.frontend_accessibility_summary.v2"`, `status="pass"`, and the FE-P5 row has keyboard, state-communication, and contrast checks with zero violations.
- `FE-V-P5-01` and `FE-A11Y-P5-01` remain design-direction evidence only. Their closure does not create product conformance and does not activate or imply Core 05.
- `make check` passed with run root `.cartulary/test-results/20260606T181959Z-p1547776`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260606T181959Z-p1547776` passed with run root `.cartulary/test-results/20260606T182755Z-p1608394` and refreshed harness duration/scheduler artifacts from the successful retained run.

## Authority Model

Core 00 through Core 04 remain the only product-conformance authority for FE-P5 product rows:

- Core 00 owns document status, precedence, profile interpretation, and implementation-conformance boundaries.
- Core 01 owns view-shaped read contracts, public `/api/v1/` route envelopes, hot-path retrieval, mutation envelopes, and generated protocol/view contracts.
- Core 02 owns record model behavior for entity mentions, stub entities, host and identity records, mention provenance, entity provenance, exact-match precedence, and entity-origin semantics.
- Core 03 owns workbook interaction behavior, Timeline read/write behavior, entity sheets, presence/workflow interaction boundaries, and stable workbook-surface behavior.
- Core 04 remains applicable for auth, session, authorization, CSRF, same-origin, and trust-boundary behavior when public browser evidence exercises those surfaces.

Core 05 remains inactive for FE-P5 unless an explicit claim-bearing publication predicate exists. Visual, accessibility, fixture, and timing evidence collected for FE-P5 is implementation-quality or design-direction evidence unless Core 05 publication requirements are separately satisfied.

Supporting-source boundaries:

- `docs/testing-harness-nlspec.md` owns harness mechanics: command invocation, target selection, row accounting artifacts, artifact emission, scheduling, cleanup, and verification gates. It does not define Core product behavior.
- `docs/domain.md` is vocabulary and concept-boundary support only. It helps keep `entity mention`, `stub entity`, `host`, `identity`, `resolved`, `dismissed`, and `auto-resolution` usage precise.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md` controls FE-P5 row planning, row-to-owner mapping, row-to-target mapping, evidence classes, phase completion, frontend namespace rules, and claim-publication separation.
- `docs/guides/cartulary-ui-ux-design-guide.md`, `docs/design.md`, and `docs/guides/cartulary_visual_golden_maintenance.md` provide design-direction and visual-maintenance inputs only.
- Generated ledgers under `docs/testing/frontend_phase_coverage_ledgers/` are downstream generated artifacts. They must not be hand-edited and must not become row owners.

## Current Repo Status

Frontend phase registry status after Sprints 5-7:

| Phase | Status | Row rollup state | Activation blocker |
| --- | --- | --- | --- |
| `FE-P0` | `active` | `active_green` | none |
| `FE-P1` | `active` | `active_green` | none |
| `FE-P2` | `active` | `active_green` | none |
| `FE-P3` | `active` | `active_green` | none |
| `FE-P4` | `active` | `active_green` | none |
| `FE-P5` | `active` | `active_green` | none |
| `FE-P6` | `planned` | `no_rows_implemented` | `frontend_phase_not_active` |

FE-P5 map and ledger status:

- `tools/frontend_phase_maps/fe_p5_test_map.json` exists with `schema_id="cartulary.frontend_phase_test_map.v3"`, `schema_version=3`, `phase_namespace="frontend"`, and `phase_id="FE-P5"`.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p5_coverage_ledger.md` exists and states it is generated from the FE-P5 map.
- The FE-P5 generated ledger is a rendered companion only; map metadata is the source of truth.
- `make phase-ledger-drift` passed after Sprint 7 map/registry promotion. Any later map, registry, guide, or generated-ledger edit must rerun the appropriate generator and drift checks.

FE-P5 visual fixture status:

- `tools/frontend_visual_fixture_registry.json` maps `FE-VFIX-02` to `FE-P5` and `FE-V-P5-01`.
- `FE-VFIX-02` has title `Mention chip state matrix` and status `current`.
- The mapped visual scenario now captures unresolved, resolved, auto-resolved, dismissed, and manual-resolution states together under `FE-V-P5-01` row accounting.
- Registry status and golden-file presence are not row closure evidence. `FE-V-P5-01` is closed only by current row-owned `browser-e2e-visual/frontend-row-accounting.json`, not by the registry or snapshot files.

Known source-validation status:

- The FE-P5 accessibility owner ambiguity was resolved by removing stale `D-AC-009`/`D-AC-012` FE-P5 owner claims and mapping `FE-A11Y-P5-01` to `docs/design.md` accessibility criteria `D-AC-050` and `D-AC-051`. Other frontend phases may still reference older design IDs and are outside this FE-P5 remediation.

## Source Limits

FE-P5 source limits:

- The plan may reference owner docs, maps, ledgers, registries, and prior phase plans, but it may not promote itself into evidence.
- Product-conformance rows require Core-owned behavior and current row-owned target evidence. Product route, refresh, mutation, or persistence claims require public browser-facing `/api/v1/` evidence, not frontend-only mocks.
- Visual readiness remains `design_direction`; it cannot close product rows and cannot activate Core 05.
- Accessibility readiness remains `design_direction`; `make browser-e2e-a11y-preflight` is blocked-row smoke only. Implemented accessibility row closure requires `make browser-e2e-a11y` and `cartulary.frontend_accessibility_summary.v2` after the row is implemented and mapped.
- Generated ledgers, generated protocol outputs, lockfiles, and tool-managed artifacts must not be hand-edited.
- FE-P5 rows must remain in the frontend namespace and must not be appended to base `tools/phase*_test_map.json`.
- Direct `react-data-grid` imports outside `/packages/grid-adapter` remain out of bounds for frontend implementation work.
- Old retained FE-P4 artifacts are handoff context only unless rerun under current FE-P5 target and row-accounting rules.

## FE-P4 Handoff Inputs

FE-P5 may rely on FE-P4 only as dependency context, not FE-P5 evidence. FE-P4 handoff states that FE-P5 may rely on:

- route-backed Timeline query surfaces that render full `view_row_v1` cells by stable identity;
- rough Timeline row creation through public route contracts;
- inline edit and paste hot paths anchored by `record_id`, `field_key`, `base_row_version`, `view_schema_id`, and `client_txn_id`;
- deterministic pending queue and replay behavior;
- same-surface save-state and validation/error presentation;
- visual and accessibility readiness state recorded without promoting design evidence into product conformance.

Inherited constraints after FE-P4:

- FE-P5 did not inherit FE-P4 row closure as FE-P5 row evidence; FE-P5 rows closed only after their own direct row-owned evidence.
- FE-P5 depends on active FE-P4; local registry inspection shows FE-P4 is currently `active_green`.
- FE-P5 phase activation occurred only after all five FE-P5 rows had direct row-owned evidence and the required drift/freshness checks passed.
- FE-P4 non-claims remain non-claims for FE-P5 unless FE-P5 closed them through its own rows. FE-P5 still does not claim evidence handles, WebSocket live updates, same-field conflict resolver implementation, saved-view persistence, or Core 05 claim-bearing publication evidence.

## Phase Objective

FE-P5 closes entity and mention flows only when the frontend can show and mutate entity mentions and host/identity/Notes surfaces through owner-backed contracts while preserving token/provenance separation.

The phase objective is to implement and verify:

- Hosts, Identities, and Notes grid rendering from contract-derived columns.
- Entity mention tokens as source-bound observations, not weak entities.
- Unresolved token state, resolved chip state, auto-resolved chip state, dismissed mention state, and manual resolution state.
- Mention provenance visibility and preservation through edit and refresh.
- Manual resolution, dismissal, auto-resolution disclosure, and undo through public mutation routes and refreshed rows.
- Visual readiness for the FE-P5 chip-state fixture matrix.
- Accessibility readiness for names, focus, and non-color-only state distinctions.

## Implementation Scope

In scope:

- View-model support for mention chip states by stable identifiers and `field_key`.
- Rendering and refresh behavior for Hosts, Identities, and Notes grids.
- Preservation of raw mention/provenance data after resolution and refresh.
- Manual resolution and dismissal controls for selected mentions where owner contracts expose the action.
- Auto-resolution disclosure where owner contracts allow deterministic exact-match reuse.
- Undo or revert behavior where owned by the current mutation/history contract and visible through refreshed rows.
- Public browser-facing evidence for product rows that touch routes, persistence, mutations, or refresh.
- Stable selectors and frontend row accounting for mapped scenarios.
- Visual fixture capture for unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual resolution state.
- Accessibility readiness for mention chip names, focus, and non-color-only distinction.

## Out of Scope

Out of scope for FE-P5:

- Evidence handle redemption and evidence lifecycle behavior; FE-P6 owns that scope.
- Same-field conflict resolver implementation and WebSocket live updates.
- Saved-view persistence.
- Coordination workbook surfaces beyond entity relationships.
- General fuzzy matching or automatic entity creation from every mention.
- Claim-bearing timed, benchmark, visual, fixture-sensitive, accessibility, or measurement publication without Core 05 activation.
- Changing Core owner requirements, generated protocol files by hand, or generated ledgers by hand.
- Treating visual or accessibility evidence as product conformance.

## Row Inventory

The FE-P5 row inventory is derived from the current frontend guide, FE-P5 map, and generated ledger. All five FE-P5 rows are implemented and closed by current mapped row-owned evidence. This plan records those artifacts for handoff only; it is not row closure evidence.

| Row | Layer | Evidence class | Current claim status | Mapped targets | Scenario title status |
| --- | --- | --- | --- | --- | --- |
| `FE-U-P5-01` | `unit` | `product_conformance` | `implemented` | `make frontend-unit` | `FE-U-P5-01 preserves closed mention chip states by stable identifiers and field keys` |
| `FE-I-P5-01` | `integration` | `product_conformance` | `implemented` | `make frontend-unit`; `make browser-e2e-webserver-backed` | `FE-I-P5-01 Verify Hosts, Identities, and Notes grids render contract-derived columns and preserve mention/entity provenance through edit and refresh.` |
| `FE-E-P5-01` | `e2e` | `product_conformance` | `implemented` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | `FE-E-P5-01 Verify manual mention resolution, dismissal, auto-resolution disclosure, and undo through public mutation routes and refreshed rows.` |
| `FE-V-P5-01` | `visual` | `design_direction` | `implemented` | `make browser-e2e-visual` | `FE-V-P5-01 Capture unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual resolution state fixtures.` |
| `FE-A11Y-P5-01` | `accessibility` | `design_direction` | `implemented` | `make browser-e2e-a11y` | `FE-A11Y-P5-01 Verify mention chip states and manual-resolution controls have accessible names, visible focus, and non-color-only distinction.` |

Product-conformance row owners:

- `FE-U-P5-01`: Core 02 Sections 6 and 7.1; Core 03 Section 4.3.
- `FE-I-P5-01`: Core 01 Sections 3.3.4 and 8.5; Core 02 Sections 7.3 and 8.2; Core 03 Sections 15 and 16.1.
- `FE-E-P5-01`: Core 01 Section 3.3.6; Core 02 Sections 6, 7.1, and 7.3; Core 03 Section 4.3.

Design-direction row owners:

- `FE-V-P5-01`: UI/UX guide Sections 10.3 and 13; visual golden guide Sections 2, 3, and 5.
- `FE-A11Y-P5-01`: UI/UX guide Sections 10.3, 10.5, and 14.

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers |
| --- | --- | --- | --- |
| [x] | 1. Readiness, map, ledger, and FE-P4 handoff validation | `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P5`; row-inventory checks; `make phase-ledger-drift`; `git diff --check` | Historical readiness gate; FE-P5 is now active after all five rows obtained row-owned evidence |
| [x] | 2. Mention chip state model for `FE-U-P5-01` | `make frontend-unit`; `make frontend-typecheck` | Complete for the unit row; later sprints closed the remaining rows |
| [x] | 3. Hosts, Identities, and Notes grid/provenance integration for `FE-I-P5-01` | `make frontend-unit`; `make browser-e2e-webserver-backed`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make json-shape-check`; `make phase-ledger-drift` | Complete for the integration row; later sprints closed product E2E, visual, and accessibility rows |
| [x] | 4. Manual resolution, dismissal, auto-resolution disclosure, and undo E2E for `FE-E-P5-01` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make frontend-typecheck`; `make frontend-unit`; `make phase-ledgers`; `make phase-ledger-drift` | Complete for the product E2E row; later sprints closed visual and accessibility rows |
| [x] | 5. Visual readiness for `FE-V-P5-01` | `make browser-e2e-visual` | Complete from row-owned visual evidence; `FE-VFIX-02` registry status alone did not close the row |
| [x] | 6. Accessibility readiness for `FE-A11Y-P5-01` | `make browser-e2e-a11y-preflight` for blocked smoke; `make browser-e2e-a11y` after implemented-row mapping | Complete from `cartulary.frontend_accessibility_summary.v2`; preflight smoke was not used as closure |
| [x] | 7. Closure, drift, final validation, and FE-P6 handoff | Row-owned targets plus `make frontend-import-boundary-check`, drift checks, `make agent-finalize`, and `make check` when required | Complete; FE-P5 is active/active_green and FE-P6 remains planned/no_rows_implemented |

## Sprint-by-Sprint Execution Plan

### Sprint 1: Readiness, Map, Ledger, And FE-P4 Handoff Validation

Objective: prove FE-P5 planning metadata is traceable before FE-P5 behavior work begins.

Owned rows: none. This sprint supports all FE-P5 rows.

Non-owned rows at Sprint 1 scope: `FE-I-P5-01`, `FE-E-P5-01`, `FE-V-P5-01`, and `FE-A11Y-P5-01`; each requires owning-sprint evidence before closure.

Non-goals:

- Do not implement FE-P5 behavior.
- Do not promote any row.
- Do not edit generated ledgers.
- Do not activate FE-P5.
- Do not activate Core 05.

Source constraints:

- The frontend guide controls the FE-P5 row set and evidence-class boundaries.
- The FE-P5 authored map controls row inventory; the FE-P5 generated ledger is downstream.
- FE-P4 handoff is dependency context only, not FE-P5 closure evidence.

Inspection checklist:

- Inspect the FE-P5 section of `docs/guides/cartulary_frontend_implementation_testing_guide.md`.
- Inspect `tools/frontend_phase_registry.json` for FE-P0 through FE-P6.
- Inspect `tools/frontend_phase_maps/fe_p5_test_map.json`.
- Inspect `docs/testing/frontend_phase_coverage_ledgers/fe_p5_coverage_ledger.md`.
- Inspect `FRONTEND_PHASE4_IMPLEMENTATION_PLAN.md` FE-P5 handoff.
- Inspect FE-P0 through FE-P4 handoff and closure sections for evidence-recording style only.
- Inspect `tools/frontend_visual_fixture_registry.json` for `FE-V-P5-01` fixture identity.
- Inspect `docs/testing-harness-nlspec.md` frontend namespace, row-accounting, visual, and accessibility mechanics.

Test-first sequence:

- Run `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P5`.
- Validate the FE-P5 registry tuple with `jq -e`.
- Validate FE-P5 row inventory, duplicate count, implemented `FE-U-P5-01` status, and remaining blocked claim statuses with `jq -e`.
- Run `make phase-ledger-drift`.
- Run `git diff --check` after creating or updating this plan.

Implementation tasks:

- Record exact current FE-P5 registry status, row rollup state, map path, ledger path, dependency status, and activation blocker.
- Record exact row inventory and targets from the FE-P5 map.
- Record FE-P4 handoff inputs as context only.
- Record `FE-VFIX-02` fixture metadata as registry metadata only.
- Record the corrected FE-P5 accessibility design IDs and keep unresolved legacy design IDs out of FE-P5 closure criteria.

Validation commands:

- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P5`
- `make phase-ledger-drift`
- `git diff --check`
- `make json-shape-check` only if schema-shaped manifests change

Evidence requirements:

- Retain command outcome, run root, and summary path only for commands actually run.
- Keep this plan's readiness evidence separate from row closure evidence.
- Do not update FE-P5 map status or registry activation metadata in Sprint 1.

Blocker rules:

- `BLOCKER: FE-P5 map row inventory invalid; expected exactly FE-U-P5-01, FE-I-P5-01, FE-E-P5-01, FE-V-P5-01, FE-A11Y-P5-01 once each; actual=<ids/counts>.`
- `BLOCKER: FE-P5 registry tuple is not frontend-namespace traceable; expected namespace/path/dependency tuple does not match registry.`
- `BLOCKER: FE-P5 generated ledger is stale relative to map; rerun generator only after confirming the authored map is the intended source.`
- `BLOCKER: FE-P4 handoff validation missing or stale for FE-P5 dependency; minimum_follow_up=<rerun exact dependency check or record owner-accepted rationale>.`
- `BLOCKER: FE-P5 design acceptance mapping stale; row=FE-A11Y-P5-01 expected=D-AC-050,D-AC-051 actual=<ids> minimum_follow_up=<owner-approved map correction>.`

Binary acceptance:

- FE-P5 row inventory, registry status, generated ledger status, FE-P4 dependency context, and fixture identity are recorded.
- At Sprint 1 close-out, FE-P5 remained planned, `FE-U-P5-01` was implemented, and the other FE-P5 rows still required owning-sprint evidence before closure.
- No generated artifact is hand-edited.

Explicit non-claims:

- Sprint 1 does not close FE-P5 rows.
- Sprint 1 does not claim product mention/entity flow closure.
- Sprint 1 does not claim visual readiness, accessibility readiness, full phase completion, or Core 05 publication evidence.

### Sprint 2: Mention Chip State Model For `FE-U-P5-01`

Objective: implement and unit-test mention chip view models for unresolved, resolved, auto-resolved, dismissed, and manual-resolution states by stable identifiers and field keys.

Owned rows: `FE-U-P5-01`.

Non-owned rows: `FE-I-P5-01`, `FE-E-P5-01`, `FE-V-P5-01`, and `FE-A11Y-P5-01`.

Sprint 2 close-out status before Sprint 3:

- Sprint 2 model support is implemented in `apps/web/src/workbookMentionChips.ts` and consumed by the workbook shell renderer.
- The row-owned unit coverage lives in `apps/web/src/WorkbookShell.phase5.mentionChips.test.ts` under the exact mapped title `FE-U-P5-01 preserves closed mention chip states by stable identifiers and field keys`.
- `make frontend-unit` passed with run root `.cartulary/test-results/20260606T003348Z-p56180`.
- `frontend-unit/frontend-row-accounting.json` includes `FE-U-P5-01` as the only planned FE-P5 row in active-target accounting; its `row_results` entry has `claim_status_at_run="implemented"` and `closure_status="closed"`.
- `make frontend-typecheck` passed with run root `.cartulary/test-results/20260606T003348Z-p56185`.
- No selector helper, generated contract, generated protocol, public route, grid integration, visual readiness, accessibility readiness, FE-P5 activation, or FE-P5 completion update was made by this sprint.

Non-goals:

- Do not claim public-route mutation behavior.
- Do not claim Hosts, Identities, or Notes grid integration.
- Do not claim visual or accessibility readiness.
- Do not activate FE-P5.

Source constraints:

- Core 02 owns mention, stub, and provenance semantics.
- Core 03 owner references in the row must remain resolved before promotion.
- View models must preserve source-bound mention identity and raw mention data; unresolved mentions must not become weak entities.

Inspection checklist:

- Inspect generated protocol and view-contract consumers for mention/entity field shapes.
- Inspect current frontend view-model and renderer/editor registry patterns.
- Inspect stable selector/test-id builders before adding chip selectors.
- Inspect existing FE-P3 and FE-P4 unit test patterns for row-owned frontend accounting.

Test-first sequence:

- Add or update focused unit tests for mention chip state derivation before implementation changes.
- Cover unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual-resolution state.
- Cover stable `record_id`, `field_key`, mention identifier, and target entity identifier anchoring where available.
- Run `make frontend-unit`.
- Run `make frontend-typecheck`.

Implementation tasks:

- Implement the minimal state model needed to render chip states without losing raw mention/provenance fields.
- Keep dismissed mentions inspectable where displayed and excluded from active relationship values where owner behavior requires it.
- Distinguish auto-resolved from manually resolved state in the model.
- Add stable selector/test-id coverage only through existing frontend selector helpers.

Validation commands:

- `make frontend-unit`
- `make frontend-typecheck`
- `make browser-e2e-support` when shared helpers or selectors change
- `make frontend-import-boundary-check` when package boundaries are touched

Evidence requirements:

- `frontend-unit/frontend-row-accounting.json` must close `FE-U-P5-01` before row promotion.
- Unit evidence must be current, row-owned, and mapped to the FE-P5 map digest at run time.
- Unit evidence closes only `FE-U-P5-01`, not integration, browser E2E, visual, accessibility, or phase completion.
- Current implementation evidence is row-owned unit evidence for `FE-U-P5-01` only.

Blocker rules:

- `BLOCKER: FE-U-P5-01 unit row accounting missing or stale; target=frontend-unit run_root=<run_root> minimum_follow_up=rerun after map and implementation are current.`
- `BLOCKER: FE-U-P5-01 mention chip state model collapses unresolved/resolved/auto_resolved/dismissed/manual states; minimum_follow_up=restore closed state vocabulary and stable identifiers.`
- `BLOCKER: FE-U-P5-01 view model treats unresolved mention as host or identity entity; minimum_follow_up=preserve source-bound mention semantics from Core 02.`

Binary acceptance:

- Sprint 2 is complete when `make frontend-unit` passes the focused `FE-U-P5-01` unit coverage, `make frontend-typecheck` passes, and current `frontend-unit/frontend-row-accounting.json` closes `FE-U-P5-01`.
- Evidence classes remain separated.

Current blocker status:

- No current `FE-U-P5-01` blocker remains. At Sprint 2 close-out, `FE-I-P5-01`, `FE-E-P5-01`, `FE-V-P5-01`, and `FE-A11Y-P5-01` still required owning-sprint evidence; those rows are now implemented and closed by later-sprint row-owned evidence.

Explicit non-claims:

- Sprint 2 does not close public-route product behavior.
- Sprint 2 does not close grid integration.
- Sprint 2 does not close visual or accessibility readiness.
- Sprint 2 does not activate FE-P5.
- Sprint 2 does not complete FE-P5.

### Sprint 3: Hosts, Identities, And Notes Grid/Provenance Integration For `FE-I-P5-01`

Objective: verify Hosts, Identities, and Notes grids render contract-derived columns and preserve mention/entity provenance through edit and refresh.

Owned rows: `FE-I-P5-01`.

Non-owned rows: `FE-E-P5-01`, `FE-V-P5-01`, and `FE-A11Y-P5-01`. `FE-U-P5-01` may be dependency evidence only if Sprint 2 has current row closure.

Sprint 3 close-out status before Sprint 4:

- Sprint 3 contract row and column support is implemented in `apps/web/src/workbookContractRows.ts`.
- `apps/web/src/WorkbookShell.tsx` renders Hosts, Identities, and Notes grid columns from the applicable `cartulary.view.*.v1` contracts instead of handwritten Phase 4-era entity column lists.
- Hosts and Identities now expose direct-value edit controls through the existing generic edit selector surface and submit public `PATCH /api/v1/records/{record_id}` mutations with `view_schema_id`, `base_row_version`, `client_txn_id`, and field-keyed `changes`.
- `internal/modules/entities/patch_store.go` and the workbook mutation route bridge add narrow Host/Identity direct-value PATCH support for contract-writable scalar fields. No database migration was required.
- Collection-like cells, including aliases and tags, render stable readable item text where available instead of collapsing contract data to opaque counts.
- Notes remain on the generic workbook surface path but are normalized before rendering and verified against contract-visible columns.
- Mention/entity provenance preservation uses the FE-U-P5-01 `apps/web/src/workbookMentionChips.ts` model as dependency context. Sprint 3 did not move FE-P5 behavior into Phase 4 helper surfaces.
- Unit coverage lives in `apps/web/src/WorkbookShell.phase5.gridProvenance.test.tsx` under exact mapped title `FE-I-P5-01 Verify Hosts, Identities, and Notes grids render contract-derived columns and preserve mention/entity provenance through edit and refresh.`
- Browser-backed product-conformance coverage lives in `apps/web/e2e/frontend.phase5.grid-provenance.spec.ts` under the same exact mapped title and seeds/queries/edits through public browser-facing routes.
- During Sprint 3, `tools/frontend_phase_maps/fe_p5_test_map.json` promoted only `FE-I-P5-01` to `claim_status="implemented"`, cleared its blockers, set `closure_scope="scenario"`, and made both `frontend-unit` and `browser-e2e-webserver-backed` required for closure with scenario-title row accounting. Sprint 4 map status for `FE-E-P5-01` is recorded below.
- `tools/frontend_phase_registry.json` kept `FE-P5` at `status="planned"` and `row_rollup_state="partially_implemented"` after Sprint 3 and Sprint 4; Sprints 5-7 later promoted FE-P5 to `active/active_green`.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p5_coverage_ledger.md` was regenerated through `make phase-ledgers`; it is downstream evidence routing text only, not a row owner.
- Sprint 3 did not close `FE-E-P5-01`, `FE-V-P5-01`, or `FE-A11Y-P5-01`; Sprint 4 status for `FE-E-P5-01` and Sprints 5-7 status for visual/accessibility/phase activation are recorded below.

Non-goals:

- Do not claim manual resolution, dismissal, auto-resolution undo, or stateful mutation closure.
- Do not claim visual or accessibility readiness.
- Do not implement FE-P6 evidence lifecycle behavior.

Source constraints:

- Core 01 owns view-shaped read contracts and hot-path retrieval/evidence boundary semantics.
- Core 02 owns entity provenance and exact-match precedence.
- Core 03 owns Timeline read/write contract and entity/evidence sheet behavior.
- Product-conformance evidence for browser-facing read and refresh behavior must use public route boundaries, not frontend-only mocks.

Inspection checklist:

- Inspect Hosts, Identities, and Notes view-schema definitions and generated field-key contracts.
- Inspect current grid adapter consumption and renderer/editor registry behavior.
- Inspect API client route helpers for public `/api/v1/` use.
- Inspect browser-backed FE-P4 query/render evidence patterns for stable row identity and refresh.

Test-first sequence:

- Add or update unit coverage for contract-derived columns and provenance-preserving cell models.
- Add or update browser-backed scenario coverage matching the exact map scenario title.
- Exercise edit and refresh while preserving raw mention/provenance data.
- Run `make frontend-unit`.
- Run `make browser-e2e-webserver-backed`.

Implementation tasks:

- Wire Hosts, Identities, and Notes grid surfaces to contract-derived columns.
- Preserve mention/entity provenance across edit, refresh, and row render.
- Keep row anchoring by stable record identifiers and field keys.
- Use existing grid-adapter boundaries; do not import `react-data-grid` directly outside `/packages/grid-adapter`.

Validation commands:

- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make frontend-typecheck`
- `make frontend-import-boundary-check`
- `make json-shape-check`
- `make phase-ledgers` after the FE-P5 map promotion
- `make phase-ledger-drift`
- `make browser-e2e-support` when helpers/selectors change

Evidence requirements:

- `FE-I-P5-01` must close in both mapped targets when required by the promoted map state.
- Browser row accounting must include the exact scenario title from the FE-P5 map.
- Public route evidence must use current server-managed data and refreshed rows.
- Current `frontend-unit/frontend-row-accounting.json` at `.cartulary/test-results/20260606T051531Z-p544526/frontend-unit/frontend-row-accounting.json` closes `FE-I-P5-01`.
- Current `browser-e2e-webserver-backed/frontend-row-accounting.json` at `.cartulary/test-results/20260606T051553Z-p546409/browser-e2e-webserver-backed/frontend-row-accounting.json` closes `FE-I-P5-01`.

Blocker rules:

- `BLOCKER: FE-I-P5-01 browser row lacks exact scenario_titles[] for row-owned closure; target=browser-e2e-webserver-backed.`
- `BLOCKER: FE-I-P5-01 product evidence used frontend-only mocks instead of public /api/v1/ browser-facing evidence; minimum_follow_up=rerun with server-backed route evidence.`
- `BLOCKER: FE-I-P5-01 provenance lost after edit or refresh; minimum_follow_up=preserve mention/entity provenance fields through row model and render path.`
- `BLOCKER: FE-P5 direct react-data-grid import detected outside /packages/grid-adapter; path=<path> minimum_follow_up=route through grid adapter.`

Binary acceptance:

- Current mapped `frontend-unit` and `browser-e2e-webserver-backed` row accounting closes `FE-I-P5-01`.
- Grid integration does not collapse product-conformance evidence with visual or accessibility evidence.
- No direct RDG import boundary regression exists.
- `FE-I-P5-01` has no remaining row blocker in `tools/frontend_phase_maps/fe_p5_test_map.json`.
- Sprint 3 left `FE-P5` planned and partially implemented; after Sprint 7, full phase completion is no longer blocked.

Explicit non-claims:

- Sprint 3 does not close manual resolution, dismissal, auto-resolution undo, visual readiness, accessibility readiness, or FE-P5 phase completion.

### Sprint 4: Manual Resolution, Dismissal, Auto-Resolution Disclosure, And Undo E2E For `FE-E-P5-01`

Objective: verify manual mention resolution, dismissal, auto-resolution disclosure, and undo through public mutation routes and refreshed rows.

Owned rows: `FE-E-P5-01`.

Non-owned rows: `FE-V-P5-01` and `FE-A11Y-P5-01`. Earlier product rows may be dependency context only if closed from current evidence.

Current implementation status as of 2026-06-06:

- Timeline relationship-item projections expose mention concurrency metadata as `mention_row_version`; the frontend mention chip model reads it as `mentionRowVersion`.
- Inspector explicit mention actions for resolving to an existing target, dismissing, reverting from dismissed/resolved state, correcting a resolved target, and undoing an auto-resolved target call `POST /api/v1/entity-mentions/{entity_mention_id}/resolve`.
- Mention-route request payloads use `base_mention_row_version`, `client_txn_id`, `action`, and `resolved_record_id` only for `resolve_item`. Create-from-mention remains on the existing record PATCH path because the mention-scoped route does not create entities.
- Rows refresh after each successful public mention mutation through the existing row loader. Dismissed mentions remain inspectable in the inspector response and refreshed row state, but do not contribute to active relationship values as if resolved.
- Resolved-state correction reuses the existing target select/action control, excluding create-from-mention in resolved state.
- Auto-resolved and manually resolved chips remain visibly distinguishable, with auto-resolution disclosure preserved until successful correction or revert. `Review` selects the mention without dismissing the disclosure, and failed correction or revert attempts leave the disclosure visible.
- Browser-backed coverage lives in `apps/web/e2e/frontend.phase5.mention-lifecycle.spec.ts` under exact mapped title `FE-E-P5-01 Verify manual mention resolution, dismissal, auto-resolution disclosure, and undo through public mutation routes and refreshed rows.`
- Existing Phase 4 mention and auto-resolution browser tests plus frontend support/unit mocks were updated to expect mention-route responses and refreshed row queries where inspector actions now use the mention route.
- Sprint 4 promotes only `FE-E-P5-01` in `tools/frontend_phase_maps/fe_p5_test_map.json`, clearing its row blockers, setting `closure_scope="scenario"`, and requiring both browser targets for closure.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p5_coverage_ledger.md` was regenerated through `make phase-ledgers`; it remains downstream generated evidence-routing text only.
- `tools/frontend_phase_registry.json` kept `FE-P5` at `status="planned"` and `row_rollup_state="partially_implemented"` at Sprint 4 close-out. Sprints 5-7 later promoted `FE-V-P5-01` and `FE-A11Y-P5-01` and activated FE-P5.

Non-goals:

- Do not use frontend-only mocks for product closure.
- Do not implement saved-view persistence, conflict resolver UI, WebSocket live updates, or FE-P6 evidence lifecycle behavior.
- Do not claim visual or accessibility readiness.

Source constraints:

- Core 01 owns success and error envelopes.
- Core 02 owns mention resolution, mention provenance, entity provenance, exact-match reuse, and omission behavior when owner conditions fail.
- Core 03 owns relevant workbook interaction behavior.
- Public mutation evidence must use `/api/v1/` browser-facing routes, server-managed session state, stable identifiers, and refreshed rows.

Inspection checklist:

- Inspect available public mutation route contracts for mention resolution, dismissal, and undo/revert surfaces.
- Inspect existing FE-P4 pending queue, replay, refresh, and public-envelope handling.
- Inspect row-accounting expectations for `browser-e2e-webserver-backed` and `browser-e2e-stateful`.
- Inspect seeded fixture/data helpers only through harness-owned public boundaries.

Test-first sequence:

- Add or update browser-backed coverage using the exact `FE-E-P5-01` scenario title from the map.
- Exercise manual resolution and confirm the raw mention remains inspectable after refresh.
- Exercise dismissal and confirm active relationship values do not treat dismissed mention as resolved.
- Exercise auto-resolution disclosure and correction/undo where owned.
- Run `make browser-e2e-webserver-backed`.
- Run `make browser-e2e-stateful`.

Implementation tasks:

- Implement manual resolution controls and route calls using stable mention and target identifiers.
- Implement dismissal controls with refreshed row state.
- Surface auto-resolution disclosure without turning suggestions into mutations unless owner predicates are satisfied.
- Implement undo/revert through the owned mutation/history route if available; if not available, record a precise blocker rather than inventing behavior.

Validation commands:

- `make browser-e2e-webserver-backed`
- `make browser-e2e-stateful`
- `make frontend-unit` when shared model logic changes
- `make frontend-typecheck`
- `make phase-ledgers` after the FE-P5 map promotion
- `make phase-ledger-drift`
- `make browser-e2e-support` when helpers/selectors change

Evidence requirements:

- Both mapped browser targets must produce current `frontend-row-accounting.json` closure for `FE-E-P5-01`.
- Scenario title must exactly match the FE-P5 map.
- Product row evidence must prove mutation through public routes and refreshed rows.
- Error and conflict cases must use public envelopes where exercised.
- Current `browser-e2e-webserver-backed/frontend-row-accounting.json` at `.cartulary/test-results/20260606T144938Z-p936158/browser-e2e-webserver-backed/frontend-row-accounting.json` closes `FE-E-P5-01`.
- Current `browser-e2e-stateful/frontend-row-accounting.json` at `.cartulary/test-results/20260606T145520Z-p952183/browser-e2e-stateful/frontend-row-accounting.json` closes `FE-E-P5-01`.

Blocker rules:

- `BLOCKER: FE-E-P5-01 manual resolution cannot be exercised through an owner-backed public mutation route; minimum_follow_up=<owner route implementation or map blocker>.`
- `BLOCKER: FE-E-P5-01 browser target passed but row accounting is missing or stale; target=<target> run_root=<run_root> failure_reason=frontend_row_accounting.`
- `BLOCKER: FE-E-P5-01 product-conformance evidence gathered through frontend-only mocks; minimum_follow_up=rerun through public /api/v1/ browser-facing evidence.`
- `BLOCKER: FE-E-P5-01 auto-resolution disclosure is not inspectable or undoable where owned; minimum_follow_up=<add disclosure/revert evidence or record owner limitation>.`

Binary acceptance:

- `FE-E-P5-01` closes in `browser-e2e-webserver-backed` and `browser-e2e-stateful` from current row-owned evidence.
- Product-flow handoff wording remains separate from visual and accessibility readiness.
- No Core 05 publication claim is made.

Current blocker status:

- No current `FE-E-P5-01` blocker remains. The remaining FE-P5 blockers at Sprint 4 close-out were `FE-V-P5-01` and `FE-A11Y-P5-01`; no FE-P5 blockers remain after Sprint 7.

Explicit non-claims:

- Sprint 4 did not close visual readiness, accessibility readiness, FE-P5 phase completion, or FE-P6 evidence lifecycle. Sprints 5-7 closed visual and accessibility readiness and completed FE-P5; FE-P6 evidence lifecycle remains out of FE-P5 scope.

### Sprint 5: Visual Readiness For `FE-V-P5-01`

Objective: capture design-direction visual readiness for unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual resolution state fixtures.

Owned rows: `FE-V-P5-01`.

Non-owned rows: product rows and `FE-A11Y-P5-01`.

Sprint 5 close-out status after Sprints 5-7:

- The FE-P5 visual scenario in `apps/web/e2e/workbook.visual.spec.ts` captures unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual-resolution states in one deterministic fixture.
- `FE-VFIX-02` in `tools/frontend_visual_fixture_registry.json` is current with title `Mention chip state matrix`.
- Dynamic identifiers remain masked for the FE-P5 visual fixture.
- Visual goldens were refreshed through the visual golden maintenance path, including `apps/web/e2e/workbook.visual.spec.ts-snapshots/fe-v-p5-01-mention-chip-states-linux.png`.
- `make browser-e2e-visual` passed with run root `.cartulary/test-results/20260606T174416Z-p1391561`.
- `.cartulary/test-results/20260606T174416Z-p1391561/browser-e2e-visual/frontend-row-accounting.json` closes `FE-V-P5-01` with `claim_status="implemented"`, `target_status="pass"`, and `closure_status="closed"`.
- Visual evidence remains `design_direction`; it does not close product conformance and does not activate Core 05.

Non-goals:

- Do not claim product conformance from screenshots or visual goldens.
- Do not infer closure from fixture registry status, snapshot filenames, or old retained artifacts.
- Do not activate Core 05.

Source constraints:

- UI/UX guide Section 10.3 owns design direction for unresolved, resolved, auto-resolved, dismissed, and manual resolution chip presentation.
- Visual golden guide owns maintenance procedure.
- `FE-VFIX-02` is the fixture registry identity for `FE-V-P5-01`; its status is `current`, but it is not closure evidence by itself.

Inspection checklist:

- Inspect `tools/frontend_visual_fixture_registry.json` entry `FE-VFIX-02`.
- Inspect `docs/guides/cartulary_visual_golden_maintenance.md`.
- Inspect the Playwright visual scenario mapped to `FE-V-P5-01`.
- Inspect dynamic masks, viewport, theme, density, focus, editor, and inspector state before capture.

Test-first sequence:

- Confirm the visual scenario title maps to `FE-V-P5-01` row accounting.
- Capture unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual resolution state in the intended fixture.
- Run `make browser-e2e-visual`.

Implementation tasks:

- Add or adjust only the visual fixture scenario and app state needed to expose FE-P5 chip states.
- Keep dynamic identifiers masked according to the visual fixture registry.
- Update goldens only through the visual golden maintenance procedure and only when the UI change is intentional.

Validation commands:

- `make browser-e2e-visual`
- `make frontend-typecheck` when app code changes
- `make frontend-import-boundary-check` when package boundaries are touched

Evidence requirements:

- `browser-e2e-visual/frontend-row-accounting.json` must close `FE-V-P5-01`.
- Fixture evidence must use `FE-VFIX-02` or an owner-approved registry/map correction.
- Visual evidence remains `design_direction` only.

Blocker rules:

- `BLOCKER: FE-P5 visual fixture identity missing or ambiguous; row=FE-V-P5-01 expected=FE-VFIX-02 actual=<fixture_ids> minimum_follow_up=<registry or map correction>.`
- `BLOCKER: FE-V-P5-01 visual fixture not recaptured for frontend row; target=browser-e2e-visual minimum_follow_up=rerun mapped visual scenario with current row accounting.`
- `BLOCKER: FE-V-P5-01 visual fixture registry remains missing; fixture=FE-VFIX-02 minimum_follow_up=recapture all five chip states through the mapped visual scenario and refresh the golden through the visual golden guide.`
- `BLOCKER: FE-V-P5-01 closure attempted from snapshot filename or registry current status; minimum_follow_up=collect direct row-owned visual evidence.`

Binary acceptance:

- Current visual target row accounting closes `FE-V-P5-01`.
- Visual readiness is named separately from product flow closure and accessibility readiness.
- No Core 05 claim-publication predicate is introduced.

Explicit non-claims:

- Sprint 5 does not close product conformance.
- Sprint 5 does not close accessibility readiness.
- Sprint 5 does not complete FE-P5 by itself.

### Sprint 6: Accessibility Readiness For `FE-A11Y-P5-01`

Objective: verify mention chip states and manual-resolution controls have accessible names, visible focus, and non-color-only distinction.

Owned rows: `FE-A11Y-P5-01`.

Non-owned rows: product rows and `FE-V-P5-01`.

Sprint 6 close-out status after Sprints 5-7:

- Mention chips now expose accessible labels for unresolved, resolved, auto-resolved, manual-resolution, and dismissed states.
- Chip states include visible non-color text markers, including `Unresolved`, `Resolved`, `Auto`, `Manual`, and `Dismissed`.
- Manual-resolution controls and mention rows are keyboard reachable and have visible focus treatment.
- `FE-A11Y-P5-01` was promoted from blocked-row preflight smoke to implemented-row `make browser-e2e-a11y` mapping before closure.
- `make browser-e2e-a11y-preflight` passed as blocked-row smoke only with run root `.cartulary/test-results/20260606T164059Z-p1264123`; it was not used as closure evidence.
- `make browser-e2e-a11y` passed with run root `.cartulary/test-results/20260606T182938Z-p1612967`.
- `.cartulary/test-results/20260606T182938Z-p1612967/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json` has `schema_id="cartulary.frontend_accessibility_summary.v2"`, `status="pass"`, zero violations, and mapped `FE-A11Y-P5-01` keyboard, state-communication, and contrast checks.
- Accessibility evidence remains `design_direction`; it does not close product conformance and does not activate Core 05.

Non-goals:

- Do not close accessibility readiness from preflight smoke.
- Do not promote accessibility evidence into product conformance.
- Do not activate Core 05.

Source constraints:

- The current FE-P5 map targets `make browser-e2e-a11y` for `FE-A11Y-P5-01`.
- Blocked-row preflight smoke remains separate from implemented accessibility readiness.
- Implemented accessibility readiness moved through owner-approved map changes to `make browser-e2e-a11y` and the normalized `cartulary.frontend_accessibility_summary.v2` artifact.
- Accessibility evidence remains `design_direction`.
- FE-P5 accessibility design references are mapped to `D-AC-050` and `D-AC-051`; do not reintroduce unresolved `D-AC-009` as a FE-P5 closure dependency.

Inspection checklist:

- Inspect `docs/testing-harness-nlspec.md` accessibility summary and preflight summary rules.
- Inspect the current FE-P5 map target for `FE-A11Y-P5-01`.
- Inspect UI/UX guide Sections 10.3, 10.5, and 14.
- Inspect `docs/design.md` accessibility criteria `D-AC-050` and `D-AC-051`.

Test-first sequence:

- Before implemented-row promotion, run `make browser-e2e-a11y-preflight` only as blocked-row smoke.
- When implementation and map promotion are ready, add implemented-row accessibility coverage.
- Run `make browser-e2e-a11y` only when `FE-A11Y-P5-01` is implemented and mapped to the implemented-row accessibility target.
- Verify the normalized accessibility summary maps `FE-A11Y-P5-01`, scenarios, keyboard checks, state communication checks, contrast checks, violations, and artifact refs.

Implementation tasks:

- Add accessible names matching the chip-state vocabulary for unresolved, resolved, auto-resolved, and dismissed mention states.
- Ensure manual-resolution controls are keyboard reachable and visibly focused.
- Ensure chip states differ by text, marker, shape, accessible name, or other non-color cues.
- Keep preflight and implemented-row accessibility evidence separate.

Validation commands:

- `make browser-e2e-a11y-preflight` for blocked-row smoke only
- `make browser-e2e-a11y` when `FE-A11Y-P5-01` is implemented and mapped
- `make frontend-typecheck` when app code changes
- `make browser-e2e-support` when helpers/selectors change

Evidence requirements:

- Blocked-row smoke may produce `cartulary.frontend_accessibility_preflight_summary.v1`, but that artifact cannot close the row.
- Implemented accessibility closure requires `cartulary.frontend_accessibility_summary.v2` from `make browser-e2e-a11y`.
- The normalized summary must include `FE-A11Y-P5-01` as an implemented phase row with mapped scenarios.

Blocker rules:

- `BLOCKER: FE-P5 accessibility summary missing, stale, or not mapped to FE-A11Y-P5-01; target=browser-e2e-a11y minimum_follow_up=rerun implemented-row accessibility target after map promotion.`
- `BLOCKER: FE-A11Y-P5-01 remains mapped only to browser-e2e-a11y-preflight; preflight smoke cannot close implemented accessibility readiness.`
- `BLOCKER: FE-A11Y-P5-01 chip states distinguished by color only; minimum_follow_up=add visible/accessibility state markers.`

Binary acceptance:

- `FE-A11Y-P5-01` closes only after implemented-row accessibility evidence exists in `cartulary.frontend_accessibility_summary.v2`.
- Preflight evidence is either blocked-row smoke only or absent from closure claims.
- Accessibility readiness remains design-direction only.

Explicit non-claims:

- Sprint 6 does not close product conformance.
- Sprint 6 does not close visual readiness unless Sprint 5 has current row-owned evidence.
- Sprint 6 does not complete FE-P5 by itself.

### Sprint 7: Closure, Drift, Final Validation, And FE-P6 Handoff

Objective: close FE-P5 only when every FE-P5 row has direct current row-owned evidence, shared harnesses are satisfied or precisely blocked, generated artifacts are fresh, and earlier active frontend phases remain green.

Owned rows: none directly. This sprint composes closure evidence from all FE-P5 rows.

Non-owned rows: none; all row claims must be completed by their owning sprint before full phase completion.

Sprint 7 close-out status:

- All five FE-P5 rows have direct current row-owned evidence in their mapped targets.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P5` reports FE-P5 as `status: active`; all five rows are implemented, and `FE-A11Y-P5-01` targets `make browser-e2e-a11y`.
- `tools/frontend_phase_registry.json` reports FE-P0 through FE-P5 as `active/active_green`; FE-P6 remains `planned/no_rows_implemented` with activation blocker `frontend_phase_not_active`.
- Drift and policy checks passed after map, guide, fixture registry, generated ledger, duration-baseline, and scheduler maintenance updates.
- `make check` passed with run root `.cartulary/test-results/20260606T181959Z-p1547776`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260606T181959Z-p1547776` passed with run root `.cartulary/test-results/20260606T182755Z-p1608394` and refreshed harness duration/scheduler artifacts.
- `FE-P5 phase complete` is allowed under this plan's exit criteria because all five rows close from row-owned evidence, evidence classes remain separated, drift/final validation passed, FE-P0 through FE-P4 remain active green, and no Core 05 claim is implied.

Non-goals:

- Do not use broad `make check` alone as row evidence.
- Do not use generated ledgers, old retained artifacts, support-only tests, visual goldens, or this plan as closure evidence.
- Do not activate Core 05 without explicit claim-publication metadata.

Source constraints:

- FE-P5 cannot be represented as complete while any row is `blocked`, `stale`, missing current row accounting, or missing required target evidence.
- Product-flow closure is narrower than full FE-P5 phase completion.
- Visual readiness and accessibility readiness must remain separately named design-direction closure claims.
- Registry, map, ledger, freshness, and row evidence must be promoted together for phase activation.

Inspection checklist:

- Inspect every FE-P5 mapped target's current row accounting.
- Inspect generated FE-P5 ledger after any map promotion.
- Inspect FE-P0 through FE-P4 registry status and any touched regression target evidence.
- Inspect frontend import boundary status.
- Inspect generated-artifact policy and generated drift status when generated or contract surfaces are touched.
- Inspect accessibility summary and visual fixture row accounting separately.

Test-first sequence:

- Confirm every product row has current row-owned evidence.
- Confirm visual readiness has current row-owned visual evidence or record it as blocked.
- Confirm accessibility readiness has current implemented-row accessibility evidence or record it as blocked.
- Run required drift and support checks.
- Run `make agent-finalize` before broader end-of-run verification.
- Run `make check` when repository completion rules require the broad developer gate.

Implementation tasks:

- Promote FE-P5 row statuses only after each row's direct current evidence exists.
- Regenerate frontend ledgers through `make phase-ledgers` after authored phase-map changes.
- Update registry activation metadata only when all FE-P5 rows and required shared harnesses are closed or precisely blocked according to allowed closure wording.
- Prepare FE-P6 handoff that distinguishes product-flow closure, visual readiness, accessibility readiness, phase completion, and remaining blockers.

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
- `make phase-ledgers` after authored phase-map changes
- `make phase-ledger-drift`
- `make phase-schedule-drift` when schedules are affected
- `make json-shape-check` when manifests or schema-shaped artifacts change
- `make agent-finalize`
- `make check` when repository completion rules require the broad developer gate
- `git diff --check`

Evidence requirements:

- Retain run roots and exact summary paths for every command actually run.
- Use mapped `frontend-row-accounting.json` artifacts for row closure.
- Use exact scenario titles from the FE-P5 map where required.
- Keep product-conformance, design-direction, implementation-support, and claim-publication-boundary evidence separate in maps, ledgers, summaries, this plan, and handoff text.
- Record skipped checks with explicit reason.

Blocker rules:

- `BLOCKER: FE-P5 row remains blocked for phase completion; row=<row_id> blocker=<reason_code> closure_claim=<product_mention_entity_flow|visual_readiness|accessibility_readiness|phase_complete> minimum_follow_up=<specific target or owner patch>.`
- `BLOCKER: FE-P5 generated ledger is stale relative to map; rerun generator only after confirming the authored map is the intended source.`
- `BLOCKER: FE-P5 earlier active phase regression failed; phase=<FE-P0..FE-P4> target=<target> run_root=<run_root> minimum_follow_up=<fix or owner acceptance>.`
- `BLOCKER: FE-P5 evidence freshness digest stale; minimum_follow_up=rerun freshness/ledger validation after map, registry, fixture, and target evidence are final.`
- `BLOCKER: FE-P5 evidence classes collapsed; product/design/support/claim-publication evidence cannot be counted across classes.`

Binary acceptance:

- `FE-P5 product mention/entity flow closed` is allowed only when `FE-U-P5-01`, `FE-I-P5-01`, and `FE-E-P5-01` close from direct current row-owned evidence.
- Use `FE-P5 product mention/entity flow closed; FE-P5 phase completion blocked by <row/blocker>` when product rows close but design-direction rows remain blocked.
- `FE-P5 phase complete` is allowed only when all five FE-P5 rows close from direct current row-owned evidence, shared harnesses are satisfied or precisely blocked without invalidating completion, generated ledgers and drift checks pass, earlier active frontend phases remain green, and no Core 05 claim is implied.

Explicit non-claims:

- Sprint 7 does not close any row from old retained artifacts, generated ledgers, broad `make check`, visual goldens, support-only tests, accessibility preflight smoke, test names, or this plan.
- Sprint 7 does not imply Core 05 claim-publication evidence unless explicit claim metadata satisfies Core 05.

## Validation Commands

Commands to run when applicable:

- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P5`
- `make frontend-typecheck`
- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-stateful`
- `make browser-e2e-visual`
- `make browser-e2e-a11y-preflight` for blocked-row smoke only
- `make browser-e2e-a11y` when `FE-A11Y-P5-01` is implemented and mapped to the implemented accessibility target
- `make browser-e2e-support` when shared helpers or selectors change
- `make frontend-import-boundary-check`
- `make generated-artifact-policy-check`
- `make generate-drift`
- `make phase-ledgers` after authored phase-map changes
- `make phase-ledger-drift`
- `make phase-schedule-drift` when schedules are affected
- `make json-shape-check` when manifests or schema-shaped artifacts change
- `make agent-finalize`
- `make check` when repository completion rules require the broad developer gate
- `git diff --check`

Plan-only creation validation:

- For this document-only creation, the smallest required validation set is `git diff --check` and `make phase-ledger-drift`.
- Do not run `make phase-ledgers`, `make generated-artifact-policy-check`, `make generate-drift`, `make json-shape-check`, or `make phase-schedule-drift` solely for this authored plan unless a schema-shaped file, authored phase map, registry, schedule input, generated ledger, or generated artifact is changed.

Final validation after Sprints 5-7:

| Command | Status | Run root | Notes |
| --- | --- | --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P5` | pass | n/a | Reports FE-P5 as `active` with all five rows implemented; `FE-A11Y-P5-01` targets `make browser-e2e-a11y`. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260606T173144Z-p1359550` | TypeScript check passed after FE-P5 chip state and scenario changes. |
| `make frontend-unit` | pass | `.cartulary/test-results/20260606T173144Z-p1359564` | Product unit coverage remained current after chip state accessibility and manual-resolution classification changes. |
| `make browser-e2e-webserver-backed` | pass | `.cartulary/test-results/20260606T181417Z-p1532949` | Product browser target passed after final rerun; earlier transient Phase 9 shard failure did not reproduce and the final broad `make check` passed. |
| `make browser-e2e-stateful` | pass | `.cartulary/test-results/20260606T173832Z-p1378671` | Stateful product target passed after FE-P5 app-state changes. |
| `make browser-e2e-visual` | pass | `.cartulary/test-results/20260606T174416Z-p1391561` | `FE-V-P5-01` closed in `browser-e2e-visual/frontend-row-accounting.json`; visual evidence remains design-direction only. |
| `make browser-e2e-a11y-preflight` | pass | `.cartulary/test-results/20260606T164059Z-p1264123` | Blocked-row smoke only; not used for row closure. |
| `make browser-e2e-a11y` | pass | `.cartulary/test-results/20260606T182938Z-p1612967` | `cartulary.frontend_accessibility_summary.v2` passed with `FE-A11Y-P5-01` mapped keyboard/state/contrast checks and zero violations. |
| `make frontend-import-boundary-check` | pass | `.cartulary/test-results/20260606T173144Z-p1359878` | No direct `react-data-grid` boundary regression. |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260606T182850Z-p1610770` | Generated-artifact policy check passed after finalizer refresh. |
| `make generate-drift` | pass | `.cartulary/test-results/20260606T182850Z-p1610753` | Generated outputs are not drifting after final updates. |
| `make phase-ledgers` | pass | `.cartulary/test-results/20260606T182755Z-p1608394` | Run through `make agent-finalize`; regenerated generated ledgers after FE-P5 map changes. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260606T182850Z-p1610752` | Generated frontend phase ledgers match current maps. |
| `make phase-schedule-drift` | pass | `.cartulary/test-results/20260606T182850Z-p1610798` | Schedule outputs are current after finalizer duration/scheduler refresh. |
| `make json-shape-check` | pass | `.cartulary/test-results/20260606T182850Z-p1610791` | Schema-shaped artifacts validate after FE-P5 map, registry, and finalizer changes. |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260606T181959Z-p1547776` | pass | `.cartulary/test-results/20260606T182755Z-p1608394` | Retained-run maintenance used the successful full check run; six harness duration/scheduler artifacts were refreshed. |
| `make check` | pass | `.cartulary/test-results/20260606T181959Z-p1547776` | Broad repo gate passed with 142/142 work units and 752 tests; this supports final validation but is not row closure evidence. |
| `git diff --check` | pass | n/a | No whitespace errors after final edits. |
| `make browser-e2e-support` | skipped | n/a | Not required as a standalone target because shared selector/helper source was not changed; support coverage also ran where scheduled by webserver-backed/check. |

## Evidence Requirements

Every FE-P5 closure record must include:

- exact command;
- command outcome;
- run root when the command emits one;
- summary path when the command emits one;
- target-level row-accounting artifact path for frontend-aware targets;
- affected FE-P5 row ID;
- exact scenario title when the map requires one;
- evidence class;
- owner reference status;
- claim-publication intent;
- skipped-check reason when a command is not run.

Evidence separation rules:

- Product-conformance evidence closes only product-conformance rows and must use Core-owned behavior plus current mapped target evidence.
- Design-direction evidence closes only design-direction readiness rows.
- Implementation-support evidence supports harness, selector, import-boundary, generated-artifact, or drift readiness but cannot close product behavior.
- Claim-publication-boundary evidence is inactive unless a claim-bearing publication predicate exists.
- Visual and accessibility evidence must never be promoted into product conformance.
- Broad `make check` may support final readiness but cannot replace row-owned target evidence.

## Blocker Recording Rules

Every blocker must record:

- exact missing artifact or failed command;
- failing target, scheduler unit, row, fixture, or manifest path;
- run root and summary path when available;
- FE-P5 row ID or shared harness ID;
- failure class and failure reason when exposed;
- ownership classification: product, design, support, harness, generated, fixture, source-doc, dependency, or external;
- minimum follow-up action;
- the exact closure claim it blocks.

Required blocker templates:

- `BLOCKER: FE-P5 phase map missing, stale, or invalid; path=tools/frontend_phase_maps/fe_p5_test_map.json issue=<missing|stale|schema|namespace|phase_id> minimum_follow_up=<restore authored map or rerun validation>.`
- `BLOCKER: FE-P5 map row inventory invalid; expected exactly FE-U-P5-01, FE-I-P5-01, FE-E-P5-01, FE-V-P5-01, FE-A11Y-P5-01 once each; actual=<ids/counts>.`
- `BLOCKER: FE-P5 duplicate row inventory detected; duplicate_rows=<rows> minimum_follow_up=deduplicate authored FE-P5 map before closure.`
- `BLOCKER: FE-P5 owner refs unresolved; row=<row_id> owner_ref=<source/section/req/ac> minimum_follow_up=<specific inspection or owner patch>.`
- `BLOCKER: FE-P5 row remains blocked or stale; row=<row_id> claim_status=<blocked|stale> blocker=<reason_code> minimum_follow_up=<specific target or map/owner fix>.`
- `BLOCKER: FE-P5 frontend row-accounting artifact missing or stale; row=<row_id> target=<target> expected=<target>/frontend-row-accounting.json run_root=<run_root> minimum_follow_up=rerun mapped target after map and registry are current.`
- `BLOCKER: FE-P5 visual fixture identity missing or ambiguous; row=FE-V-P5-01 expected=FE-VFIX-02 actual=<fixture_ids> minimum_follow_up=<registry or map correction>.`
- `BLOCKER: FE-P5 accessibility summary missing, stale, or not mapped to FE-A11Y-P5-01; target=browser-e2e-a11y minimum_follow_up=rerun implemented-row accessibility target after map promotion.`
- `BLOCKER: FE-P5 product-conformance evidence gathered through frontend-only mocks; row=<row_id> minimum_follow_up=collect public /api/v1/ browser-facing evidence.`
- `BLOCKER: FE-P5 evidence classes collapsed; product/design/support/claim-publication evidence cannot be counted across classes.`
- `BLOCKER: FE-P5 generated ledger hand edit detected; path=docs/testing/frontend_phase_coverage_ledgers/fe_p5_coverage_ledger.md minimum_follow_up=revert hand edit, update owner map, run make phase-ledgers.`
- `BLOCKER: FE-P5 direct react-data-grid import outside /packages/grid-adapter; path=<path> minimum_follow_up=route grid usage through adapter and rerun frontend-import-boundary-check.`
- `BLOCKER: FE-P4 handoff validation missing or stale; minimum_follow_up=<rerun dependency validation or record owner-accepted rationale>.`
- `BLOCKER: FE-P5 earlier-phase dependency evidence stale; phase=<FE-P0..FE-P4> target=<target> minimum_follow_up=<rerun regression or record accepted owner rationale>.`
- `BLOCKER: FE-A11Y-P5-01 design acceptance mapping stale; expected=D-AC-050,D-AC-051 actual=<ids> minimum_follow_up=<owner-approved map correction>.`

## Strict Non-Claims

Do not claim:

- FE-P5 row closure from this plan.
- FE-P5 row closure from generated ledgers.
- FE-P5 row closure from old retained artifacts.
- FE-P5 row closure from broad `make check`.
- FE-P5 row closure from test names or scenario title text alone.
- FE-P5 row closure from support-only tests.
- FE-P5 row closure from visual golden files or fixture registry `current` status.
- FE-P5 accessibility readiness from `browser-e2e-a11y-preflight` blocked-row smoke.
- FE-P5 phase completion while any row remains `blocked`, `stale`, missing current row accounting, or missing required target evidence.
- Product conformance from visual or accessibility evidence.
- Core 05 claim-publication readiness unless explicit claim-bearing publication metadata exists and satisfies Core 05.
- FE-P5 behavior from FE-P4 handoff evidence.
- FE-P5 frontend row coverage from base `tools/phase*_test_map.json` or base browser manifest-selected tests.

## Binary Exit Criteria

Plan-creation or documentation-only completion is allowed when:

- `FRONTEND_PHASE5_IMPLEMENTATION_PLAN.md` exists and states it is not product behavior authority.
- Current FE-P5 registry status, row inventory, claim statuses, mapped targets, FE-P4 handoff inputs, visual fixture metadata, source limits, validation commands, blocker templates, and FE-P6 handoff requirements are recorded.
- No FE-P5 row is marked complete by this plan.
- Only this authored plan changes, unless a sprint implementation or traceability correction is explicitly required.
- `git diff --check` and `make phase-ledger-drift` pass after file creation, or precise blockers are recorded.

FE-P5 product-flow closure is allowed when:

- `FE-U-P5-01`, `FE-I-P5-01`, and `FE-E-P5-01` close from direct current row-owned evidence in their mapped targets.
- Product evidence uses Core-owned behavior and public `/api/v1/` browser-facing evidence wherever routes, mutations, refresh, or persistence are claimed.
- Product closure text uses `FE-P5 product mention/entity flow closed; FE-P5 phase completion blocked by <row/blocker>` when either design-direction row remains blocked.

Current product-flow status after Sprint 7:

- `FE-P5 product mention/entity flow closed.`
- Product-flow closure remains narrower than visual readiness, accessibility readiness, full FE-P5 phase completion, and Core 05 claim-publication readiness.

FE-P5 visual readiness is allowed when:

- `FE-V-P5-01` closes from current `make browser-e2e-visual` row-owned evidence.
- Fixture identity is unambiguous and tied to `FE-VFIX-02` or an owner-approved registry/map correction.
- Visual readiness remains design-direction only and does not imply product conformance or Core 05.

FE-P5 accessibility readiness is allowed when:

- `FE-A11Y-P5-01` closes from implemented-row `make browser-e2e-a11y` evidence and `cartulary.frontend_accessibility_summary.v2`.
- Blocked-row preflight smoke is not used as closure evidence.
- The row remains mapped to resolved design references such as `D-AC-050` and `D-AC-051`, or any replacement owner references are updated through the FE-P5 map.
- Accessibility readiness remains design-direction only.

Full FE-P5 phase completion is allowed only when:

- every FE-P5 row closes from direct current row-owned evidence;
- all shared harnesses triggered by FE-P5 surfaces are satisfied or precisely blocked without invalidating the completion claim;
- generated ledgers are regenerated through Make after map changes and drift checks pass;
- frontend namespace, row-accounting, and scenario-title requirements are satisfied;
- FE-P0 through FE-P4 remain green or have precise owner-accepted blockers that do not invalidate FE-P5 completion;
- product-conformance, design-direction, implementation-support, and claim-publication-boundary evidence remain separate;
- no Core 05 claim is implied.

Use `FE-P5 phase complete` only under the full completion conditions above.

Current full phase status after Sprint 7:

- `FE-P5 phase complete.`
- FE-P5 is `active/active_green`.
- FE-P0 through FE-P4 remain `active/active_green`.
- FE-P6 remains `planned/no_rows_implemented`.
- No FE-P5 blockers remain.
- No Core 05 claim-publication evidence is implied.

FE-P6 handoff readiness is allowed when:

- FE-P5 product-flow, visual-readiness, accessibility-readiness, and phase-completion status are each named separately.
- Any blocked or stale row is listed with exact blocker, target, owner, and minimum follow-up.
- FE-P6 receives only current FE-P5 evidence, not stale retained artifacts or generated-ledger-only claims.

## FE-P6 Handoff

FE-P6 may rely on FE-P5 only after the relevant FE-P5 closure claim is true and evidence is current.

If FE-P5 product mention/entity flow closes before full phase completion in a future replay, use this exact interim handoff form. This is not the current FE-P5 handoff state:

`FE-P5 product mention/entity flow closed; FE-P5 phase completion blocked by <row/blocker>.`

Current FE-P6 handoff status after Sprints 5-7:

- Hosts, Identities, and Notes contract-derived grid rendering is implemented and row-closed by `FE-I-P5-01`.
- Mention chip state modeling is implemented and row-closed by `FE-U-P5-01`.
- Mention/entity provenance preservation through Host, Identity, and Note edit plus refresh is implemented and row-closed by `FE-I-P5-01`.
- Manual resolution, dismissal, auto-resolution disclosure, correction, and undo/revert through public mention mutation routes and refreshed rows are implemented and row-closed by `FE-E-P5-01`; disclosure remains visible after `Review` and after failed correction/revert attempts until a successful correction or revert is reflected by refreshed row state.
- Public product-flow mutation evidence uses `POST /api/v1/entity-mentions/{entity_mention_id}/resolve` for resolve-to-existing, dismiss, correction, and revert/undo, with refreshed rows preserving raw mention inspectability.
- Visual readiness is implemented and row-closed by `FE-V-P5-01` from current `make browser-e2e-visual` row accounting. It remains design-direction only.
- Accessibility readiness is implemented and row-closed by `FE-A11Y-P5-01` from current `make browser-e2e-a11y` and `cartulary.frontend_accessibility_summary.v2`. It remains design-direction only.
- `FE-P5 product mention/entity flow closed.`
- `FE-P5 phase complete.`
- FE-P5 is `active/active_green`; FE-P6 remains `planned/no_rows_implemented`.
- FE-P5 has no remaining blockers.

FE-P6 handoff must include:

- Hosts, Identities, and Notes rendering status.
- Mention chip state model status.
- Manual resolution, dismissal, auto-resolution disclosure, and undo status.
- Mention provenance preservation status.
- Public route evidence status for product rows.
- Visual readiness status for `FE-V-P5-01`.
- Accessibility readiness status for `FE-A11Y-P5-01`.
- FE-P5 map, registry, generated ledger, and drift status.
- Shared harness status.
- Earlier active frontend phase regression status.
- Any remaining blockers with exact follow-up.

FE-P6 must not inherit:

- FE-P5 visual evidence as product conformance.
- FE-P5 accessibility evidence as product conformance.
- FE-P5 preflight smoke as implemented accessibility closure.
- FE-P5 generated ledger text as row-owned evidence.
- FE-P4 retained artifacts as FE-P5 evidence.
- Any Core 05 claim-publication evidence unless explicit Core 05 metadata exists.
