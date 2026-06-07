# Frontend Phase 6 Implementation Plan

## Summary

`FE-P6: Evidence Lifecycle` is a frontend planning and verification phase for evidence lifecycle user experience, route usage, selectors, and evidence-specific frontend readiness. This document is an execution roadmap, progress marker, validation guide, blocker register, and FE-P7 handoff aid. It is not product behavior authority; row promotion and closure come from authored maps, mapped target evidence, and row-accounting artifacts.

Current repository state shows FE-P6 is planned, inactive for whole-phase execution, and partially implemented. Sprint 1 readiness validation completed on 2026-06-06 with the binary recommendation `ready with blockers`. Sprint 2 promoted and closed `FE-U-P6-01` from current mapped unit evidence. Sprint 3 row-owned evidence audit completed on 2026-06-07 with the binary recommendation `promote FE-I-P6-01`. Sprint 4 promoted and closed `FE-E-P6-01` from current mapped browser E2E evidence through public same-origin evidence handles, and the Sprint 4 row-owned audit completed on 2026-06-07 with the binary recommendation `PROMOTE FE-E-P6-01`. Remaining FE-P6 blockers are `FE-V-P6-01` and `FE-A11Y-P6-01`. The implementation path must keep product-conformance, design-direction, support, and claim-publication evidence separate. Generated ledgers and retained artifacts are downstream status aids only.

The phase focuses on evidence counts, evidence states, attach flow, preview and download handle flow, same-origin handle redemption, current authorization behavior, public success and error envelopes, and prevention of raw object storage details in browser-facing access handles.

## Authority Model

Product behavior authority remains with adopted Cartulary NLSpecs and the normative core set:

- Core 00 through Core 04 own product implementation-conformance behavior.
- Core 05 is inactive for FE-P6 unless explicit claim-bearing timed, benchmark, visual, fixture-sensitive, or publication metadata exists and satisfies Core 05.
- `docs/testing-harness-nlspec.md` owns harness mechanics only: target selection, invocation, scheduling, artifacts, cleanup, row accounting, fixture lifecycle, and verification gates.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md` owns frontend phase mechanics, FE-P6 row mapping, evidence classes, and row-owner mapping as implementation-support input.
- Generated phase ledgers are downstream generated artifacts. They are never row owners and never closure evidence by themselves.
- `docs/domain.md` supplies vocabulary and concept boundaries. It does not override Core behavior.
- `docs/guides/cartulary-dev-guide.md` supplies package, import, generated-artifact, and workspace boundaries.
- `docs/guides/cartulary-ui-ux-design-guide.md`, `docs/design.md`, and `docs/guides/cartulary_visual_golden_maintenance.md` are design-direction, visual, and accessibility readiness inputs only.
- Research reports may explain rationale. They are not product behavior authority.

If a live source, map, owner reference, guide, or retained artifact conflicts with another source, FE-P6 must record a blocker instead of choosing a side.

## Current Repo Status

The following live inputs were inspected before this plan was authored and during Sprint close-outs:

- `tools/frontend_phase_registry.json`
- `tools/frontend_phase_maps/fe_p6_test_map.json`
- `docs/testing/frontend_phase_coverage_ledgers/fe_p6_coverage_ledger.md`
- `tools/frontend_visual_fixture_registry.json`
- `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md` through `FRONTEND_PHASE5_IMPLEMENTATION_PLAN.md`
- `docs/guides/cartulary_frontend_implementation_testing_guide.md`
- `docs/testing-harness-nlspec.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_base_profile.md`
- `docs/spec/02_event_model.md`
- `docs/spec/03_workbook_model.md`
- `docs/spec/04_authorization_trust_model.md`
- `docs/spec/05_claim_publication_profile.md`
- `docs/domain.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/guides/cartulary-ui-ux-design-guide.md`
- `docs/design.md`
- `docs/guides/cartulary_visual_golden_maintenance.md`
- Existing evidence lifecycle app code under `apps/web`
- Generated protocol facade and generated evidence contracts under `packages/protocol-ts`
- Selector and test-id builders under `packages/ui-contracts`
- Test utilities and browser helpers under `packages/test-utils` and `apps/web/e2e`
- Current retained visual, accessibility, preflight, check, and agent-finalize artifacts under `.cartulary/test-results`
- Make target explanations for FE-P6 and every validation target listed in this plan

Verified current status:

- `tools/frontend_phase_registry.json` marks `FE-P0` through `FE-P5` as `active` and `active_green`.
- `tools/frontend_phase_registry.json` marks `FE-P6` as `planned` with `row_rollup_state=partially_implemented` and activation blocker `frontend_phase_not_active`.
- `tools/frontend_phase_maps/fe_p6_test_map.json` contains exactly five FE-P6 rows. `FE-U-P6-01`, `FE-I-P6-01`, and `FE-E-P6-01` have `claim_status=implemented`; `FE-V-P6-01` and `FE-A11Y-P6-01` remain blocked.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p6_coverage_ledger.md` is generated from the FE-P6 map and must remain downstream only.
- `tools/frontend_visual_fixture_registry.json` contains `FE-VFIX-05`, title `Evidence affordance`, status `current`, owner row `FE-V-P6-01`. Registry status is not row closure evidence.
- Existing evidence code and legacy tests exist in `apps/web`, `packages/protocol-ts`, `packages/ui-contracts`, and E2E, visual, and accessibility files.
- Existing legacy scenario titles do not close FE-P6 rows. FE-P6 closure requires direct current row-owned evidence with the mapped FE-P6 row and scenario-title expectations.
- Latest accessibility preflight retained evidence lists `FE-A11Y-P6-01` as blocked-row smoke with `required_for_closure=false`. It cannot close FE-P6 accessibility readiness.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P6` passed and reports FE-P6 as planned and not executable as a whole phase.
- `make explain-target TARGET=<target> DETAIL=summary` passed for the requested frontend, browser, drift, finalization, and broad check targets.

## Source Limits

FE-P6 planning is bounded by source authority and current repository state:

- This plan cannot become product behavior authority.
- The FE-P6 generated ledger cannot become an owner, implementation source, or closure source.
- The visual fixture registry and visual golden filenames cannot close `FE-V-P6-01`.
- Retained successful runs can provide context only when row ownership, artifact freshness, and target mapping are current and directly applicable.
- Existing code and legacy tests can be implementation context, but they do not close FE-P6 unless the phase map, row accounting, direct evidence, and scenario-title expectations align.
- `browser-e2e-a11y-preflight` is blocked-row smoke for `FE-A11Y-P6-01` unless the accessibility row is promoted or remapped to implemented accessibility evidence.
- Broad `make check` can satisfy a repository completion rule, but it cannot close FE-P6 rows.
- Support-only tests cannot close product-conformance rows.
- Frontend-only mocks cannot close public-route product-conformance rows.
- Product-conformance rows close only from direct current row-owned evidence in mapped targets.
- Design-direction rows remain design readiness only and cannot be represented as product conformance.
- Claim-publication evidence remains absent unless explicit claim-bearing metadata exists and satisfies Core 05.

No live guide/map contradiction was found for the five current FE-P6 row identities. The live map and guide both define the same five row IDs. Any future difference between the guide and `tools/frontend_phase_maps/fe_p6_test_map.json` must be recorded as a blocker instead of silently reconciled.

## FE-P5 Handoff Inputs

FE-P5 is currently recorded as complete and active/green. FE-P6 may inherit only dependency context from FE-P5, not row closure.

Current FE-P5 handoff inputs:

- FE-P5 product flow is closed for its own phase.
- FE-P5 row states are recorded as active/green through the frontend phase registry.
- FE-P5 covers hosts, identities, notes contract-derived grid rendering, mention chip state modeling, provenance preservation, manual resolution, dismissal, auto-resolution, correction, undo, visual readiness, and accessibility readiness for FE-P5 scope.
- FE-P5 retained current evidence status is useful only for dependency validation and regression context.

FE-P6 must not inherit:

- FE-P5 generated ledger status as FE-P6 row evidence.
- FE-P5 visual or accessibility evidence as FE-P6 product conformance.
- FE-P5 public-route evidence as FE-P6 preview/download handle evidence unless direct FE-P6 row ownership and mapped target evidence exist.
- FE-P4 or FE-P5 retained artifacts as FE-P6 row closure.
- Core 05 claim-publication readiness.

If FE-P5 handoff evidence becomes stale or cannot be re-validated, record:

`BLOCKER: FE-P5 handoff validation missing or stale; minimum_follow_up=<rerun dependency validation or record owner-accepted rationale>.`

## Phase Objective

FE-P6 implements and verifies evidence lifecycle behavior in the frontend through current row-owned evidence. The objective is to move each FE-P6 row from blocked to closed only after direct mapped evidence demonstrates the required behavior without collapsing evidence classes.

Product-conformance rows must show evidence lifecycle behavior through current source-owned contracts and public browser-facing surfaces. Browser-facing product evidence must use public `/api/v1/` routes, same-origin evidence-handle redemption, server-managed sessions, stable identifiers, public success envelopes, and public error envelopes.

Design-direction rows must show visual and accessibility readiness for evidence affordances without promoting visual or accessibility artifacts into product conformance.

## Implementation Scope

FE-P6 implementation scope covers:

- Evidence counts and count-display states.
- Requested, pending upload, available, preview blocked, failed, inconsistent, blocked, and public error states.
- Evidence attach flow from frontend affordance through generated protocol and public error handling.
- Preview handle flow through opaque same-origin handle issuance and redemption.
- Download handle flow through opaque same-origin handle issuance and redemption.
- Same-origin preview and download handle redemption through public `/api/v1/` routes.
- Current authorization behavior re-derived at issuance and redemption time with server-managed sessions.
- Public success envelopes and public error envelopes.
- Stable evidence selectors and test-id builders for attach, preview, download, blocked/error states, and access messages.
- Prevention of raw object paths, raw object URLs, raw object-store keys, bucket names, backend paths, storage backend identifiers, object-store implementation details, and raw backend storage identifiers in user-facing preview or download access handles.
- Visual readiness for evidence count, affordance, available, requested, pending, blocked, failed, inconsistent, preview, and download-handle state fixtures.
- Accessibility readiness for evidence icon buttons, blocked states, error states, preview controls, and download controls.

Core owner interpretation for FE-P6:

- Core 01 owns object blob creation and evidence handle issuance/redemption behavior, including opaque same-origin upload targets, preview/download handles, public envelopes, and no raw storage details.
- Core 02 owns evidence lifecycle state vocabulary and object blob versus evidence record separation.
- Core 03 owns workbook-facing evidence surfaces, attach behavior, blocked states, preview/download affordances, and workbook continuity.
- Core 04 owns current authorization, trust boundaries, fail-closed preview/download behavior, and raw storage exposure prevention.

## Out of Scope

FE-P6 is not responsible for:

- WebSocket live updates.
- Same-field conflict resolver implementation.
- Saved-view persistence.
- Full coordination surfaces.
- Core 05 claim-publication readiness unless explicit claim-bearing publication metadata exists and satisfies Core 05.
- Backend storage implementation changes beyond what is required to exercise public frontend routes in mapped evidence.
- Promoting visual or accessibility artifacts into product conformance.
- Closing FE-P6 rows from this plan, generated ledgers, retained artifacts, fixture registry status, support-only tests, broad `make check`, or scenario titles alone.
- Raw object-store access as public evidence access behavior.

## Row Inventory

Current FE-P6 row status is partially implemented. The table records the live map status, closure posture, and direct evidence still needed before remaining promotions can be considered.

| Row | Status | Evidence | Targets | Owner references | Scenario expectation | Blocker | Direct evidence requirement |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `FE-U-P6-01` | `implemented`; closed by Sprint 2 mapped unit evidence | `product_conformance` | `make frontend-unit` | Core 01 Sections 3.3.8 and 3.3.9; Core 02 Section 13; Core 03 Sections 8 and 16; Core 04 Sections 3, 5, and 6 | Sprint 2 row-owned unit scenarios | none | current unit row-owned evidence for evidence state view models, evidence counts, requested, pending upload, available, preview blocked, failed, inconsistent, and count-display states |
| `FE-I-P6-01` | `implemented`; promoted by Sprint 3 row-owned evidence audit | `product_conformance` | `make frontend-unit`; `make browser-e2e-webserver-backed` | Core 01 Sections 3.3.8 and 3.3.9; Core 02 Section 13; Core 03 Sections 8 and 16; Core 04 Sections 3, 5, and 6; generated protocol and selector contracts | `FE-I-P6-01 Verify attach flow uses generated protocol types, public error envelopes, and stable evidence selectors without raw object URLs or paths.` | none | current generated protocol/public envelope/selector evidence, attach flow evidence, public error rendering, and absence of raw object URLs, raw paths, raw object keys, bucket names, backend paths, or storage identifiers |
| `FE-E-P6-01` | `implemented`; closed by Sprint 4 mapped browser E2E evidence | `product_conformance` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Core 01 Sections 3.3.8 and 3.3.9; Core 02 Section 13; Core 03 Sections 8 and 16; Core 04 Sections 3, 5, and 6 | `FE-E-P6-01 Verify evidence attach, preview, download, blocked preview, and authorization denial through same-origin public handles.` | none | current public `/api/v1` same-origin handle evidence for attach, preview, download, blocked preview, current authorization denial, server-managed session behavior, stable identifiers, public envelopes, and raw-handle negative checks |
| `FE-V-P6-01` | `blocked` | `design_direction` | `make browser-e2e-visual` | UI/UX design guide evidence affordance direction; `docs/design.md`; visual golden maintenance guide; `FE-VFIX-05` fixture identity | `FE-V-P6-01 Capture evidence count, affordance, available, requested, pending, blocked, failed, inconsistent, preview, and download-handle state fixtures.` | `visual_fixture_not_recaptured_for_frontend_row` | row-owned visual accounting tied to `FE-VFIX-05`, not fixture registry status or visual golden filename alone |
| `FE-A11Y-P6-01` | `blocked` | `design_direction` | `make browser-e2e-a11y-preflight` | UI/UX design guide accessibility direction; `docs/design.md`; frontend accessibility harness mechanics | `FE-A11Y-P6-01 Verify evidence icon buttons, blocked states, error states, preview controls, and download controls have names, focus, contrast, and non-color-only distinctions.` | `frontend_phase_row_not_implemented` | preflight smoke only unless row is promoted or remapped to implemented accessibility target; closure requires mapped row-owned accessibility evidence |

Owner note: Core 01 Sections 3.3.8 and 3.3.9 cover object blob and evidence handle public route behavior. Core 01 Section 16 covers blob and evidence handle schema and error semantics. Core 02 Section 13 covers evidence lifecycle state. Core 03 Sections 8 and 16 cover workbook and evidence surface behavior. Core 04 Sections 3, 5, and 6 cover authorization, trust, and raw storage exposure prevention. UI/design/visual guides apply only to design readiness rows.

## Evidence Layer Matrix

| Evidence layer | FE-P6 rows | Closure source | Cannot close from |
| --- | --- | --- | --- |
| Product conformance | `FE-U-P6-01`, `FE-I-P6-01`, `FE-E-P6-01` | Direct current row-owned evidence in mapped targets, with resolved Core or adopted NLSpec owner references | This plan, generated ledgers, retained artifacts alone, support-only tests, visual evidence, accessibility evidence, broad `make check`, frontend-only mocks for public-route behavior |
| Design direction | `FE-V-P6-01`, `FE-A11Y-P6-01` | Direct current row-owned visual or accessibility readiness evidence in mapped targets | Product conformance claims, Core 05 claims, fixture registry `current` status, golden filenames, preflight smoke for implemented accessibility closure |
| Support evidence | Shared selectors, helpers, harness utilities, import-boundary checks | Support targets and drift checks when shared surfaces change | Product-conformance row closure unless mapped row-owned product evidence also exists |
| Claim publication | none current | Explicit Core 05 claim-bearing metadata and required Core 05 evidence, if introduced | FE-P6 implementation, visual readiness, accessibility readiness, fixture status, retained artifacts, or broad checks without Core 05 metadata |

Evidence classes must never be collapsed. If product, design, support, or claim-publication evidence is counted across classes, record:

`BLOCKER: FE-P6 evidence classes collapsed; product/design/support/claim-publication evidence cannot be counted across classes.`

## Dependencies And Prerequisites

Before FE-P6 implementation work promotes any row, the phase owner should verify:

- `tools/frontend_phase_registry.json` still marks FE-P0 through FE-P5 as green or records owner-accepted blockers that do not invalidate FE-P6.
- FE-P6 is either activated by the appropriate owner workflow or rows are implemented without claiming phase completion.
- `tools/frontend_phase_maps/fe_p6_test_map.json` still contains exactly the expected FE-P6 row set or records any guide/map drift as a blocker.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p6_coverage_ledger.md` has not been hand edited.
- `tools/frontend_visual_fixture_registry.json` still has unambiguous `FE-VFIX-05` ownership for `FE-V-P6-01` if visual readiness work is attempted.
- Existing evidence code in `apps/web` is reconciled with current FE-P6 scenario-title expectations instead of relying on legacy titles.
- Generated protocol consumption uses the `packages/protocol-ts` facade and does not hand edit generated protocol files.
- Selector/test-id additions route through `packages/ui-contracts` authored inputs and generated outputs are regenerated by Make if needed.
- Shared test utilities remain inside package boundaries and import boundaries.
- Any grid usage continues to route through `/packages/grid-adapter`; direct `react-data-grid` imports outside that package are blockers.
- Browser product evidence is server-backed and public-route based.
- Authorization denial evidence uses current server-managed sessions and request-time authorization derivation.

## Shared Harness Analysis

FE-P6 touches shared harnesses because evidence lifecycle behavior spans unit state modeling, protocol integration, browser route behavior, visual readiness, and accessibility readiness.

Harness mechanics:

- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P6` is the phase discovery source. It currently passes and reports FE-P6 as planned and not executable as a whole phase.
- `make explain-target TARGET=<target> DETAIL=summary` verified target availability for all validation commands listed below.
- Product row accounting must be produced by the mapped target for the implemented row. Row accounting from old retained artifacts is diagnostic only.
- Scenario title expectations from the FE-P6 map are required for rows that declare them, but title text alone is never evidence.
- `browser-e2e-a11y-preflight` is preflight smoke. It can demonstrate that blocked future rows are not required for closure, but it cannot close `FE-A11Y-P6-01`.
- Visual fixture lifecycle mechanics are owned by the harness. `FE-VFIX-05` fixture identity must be unambiguous before visual readiness is claimed.
- Generated ledgers must be regenerated through `make phase-ledgers` only after authored phase-map changes. The ledger itself must not be hand edited.
- Drift targets detect stale generated outputs and schedules; they do not themselves create row evidence.

Public-route product rows:

- `FE-I-P6-01` may combine unit and server-backed browser evidence only when the browser portion uses public `/api/v1/` routes and public envelopes.
- `FE-E-P6-01` requires browser evidence through server-backed public routes, same-origin handle redemption, stable identifiers, current authorization, and fail-closed public error behavior.
- Frontend mocks can support view states and helper coverage, but cannot close public-route product-conformance rows.

## Public Interfaces And Deliverables

Expected implementation deliverables for FE-P6, when implementation work begins:

- Evidence lifecycle view-model helpers for evidence counts and requested, pending upload, available, preview blocked, failed, inconsistent, blocked, and public error states.
- Attach flow integration using generated protocol types and public error envelopes.
- Stable selectors/test IDs for evidence attach, preview, download, access messages, preview frames, blocked states, and count/status affordances.
- Browser-facing preview and download actions that request opaque same-origin handles and redeem them through public `/api/v1/` routes.
- Negative checks that user-facing handles do not expose raw object URLs, raw object-store keys, bucket names, backend paths, backend storage identifiers, or implementation details.
- Public error rendering for blocked preview/download, failed upload, inconsistent lifecycle state, authorization denial, stale/expired handle, and unavailable evidence.
- Server-backed E2E coverage for attach, preview, download, blocked preview, and authorization denial with current authorization re-derived at request time.
- Visual fixture coverage tied to `FE-VFIX-05` for evidence count and evidence affordance states.
- Accessibility readiness coverage for evidence icon buttons, blocked/error states, preview controls, download controls, focus, names, contrast, and non-color-only distinctions.
- FE-P7 handoff notes that preserve row evidence status, unresolved blockers, and non-claims.

Generated protocol files, generated selector outputs, generated ledgers, generated schedules, and lockfiles are not deliverables for this plan-only task. If implementation later requires generated or contract-surface changes, update authored owner inputs and run the repository-supported generator targets.

## Sprint Checklist

- [x] Sprint 1: Validate guide, map, ledger, registry, fixture, retained evidence, target availability, and FE-P5 handoff; record all FE-P6 rows as blocked.
- [x] Sprint 2: Implement evidence lifecycle view-model vocabulary and count-display coverage for `FE-U-P6-01`.
- [x] Sprint 3: Implement attach flow, generated protocol and public error envelope handling, stable selectors, and raw-handle prevention for `FE-I-P6-01`.
- [x] Sprint 4: Implement public-route browser E2E for attach, preview, download, blocked preview, and current authorization denial for `FE-E-P6-01`.
- [ ] Sprint 5: Implement visual readiness for `FE-V-P6-01` tied to `FE-VFIX-05`, keeping design evidence separate.
- [ ] Sprint 6: Implement accessibility readiness for `FE-A11Y-P6-01`, treating preflight as blocked smoke unless the row is promoted or remapped.
- [ ] Sprint 7: Run closure, drift, final validation, dependency verification for FE-P0 through FE-P5, and FE-P7 handoff.

## Sprint-by-Sprint Execution Plan

### Sprint 1: Readiness

Goal: confirm current source truth before code or map edits.

Tasks:

- Re-run `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P6`.
- Re-read `tools/frontend_phase_registry.json`, `tools/frontend_phase_maps/fe_p6_test_map.json`, FE-P6 generated ledger, visual fixture registry, FE-P5 handoff, and retained current evidence.
- Verify FE-P6 still has exactly five rows and all are blocked, or record exact drift.
- Verify FE-P0 through FE-P5 remain active/green or record owner-accepted blockers.
- Verify target availability with `make explain-target TARGET=<target> DETAIL=summary` before adding required commands to implementation checklists.
- Confirm FE-P6 generated ledger is downstream and untouched.
- Confirm `FE-VFIX-05` fixture identity remains unambiguous.
- Confirm latest `FE-A11Y-P6-01` preflight evidence is blocked-row smoke only.

Exit: readiness notes identify row status, owner references, required direct evidence, and exact blockers. No row is promoted during readiness.

Sprint 1 close-out, 2026-06-06:

- Binary recommendation: `ready with blockers` for Sprint 2 implementation work. This is not FE-P6 completion, row closure, registry activation, Core 05 readiness, or claim-publication readiness.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P6` passed and reported `status=planned`, `row_rollup_state=no_rows_implemented`, dependency set `FE-P0` through `FE-P5`, five blocked rows, and whole-phase execution unavailable while planned.
- `make explain-target TARGET=<target> DETAIL=summary` passed for every validation target referenced by this plan: `frontend-typecheck`, `frontend-unit`, `browser-e2e-webserver-backed`, `browser-e2e-stateful`, `browser-e2e-support`, `frontend-import-boundary-check`, `browser-e2e-visual`, `browser-e2e-a11y-preflight`, `browser-e2e-a11y`, `generated-artifact-policy-check`, `generate-drift`, `phase-ledgers`, `phase-ledger-drift`, `phase-schedule-drift`, `json-shape-check`, `agent-finalize`, and `check`.
- `make phase-ledger-drift` passed with run root `.cartulary/test-results/20260606T232930Z-p1968618`.
- `git diff --check` passed with no output.
- `tools/frontend_phase_registry.json` still records `FE-P6` as `planned`, `no_rows_implemented`, with activation blocker `frontend_phase_not_active`.
- `tools/frontend_phase_registry.json` still records `FE-P0` through `FE-P5` as `active` and `active_green`, with no activation blockers.
- FE-P5 handoff validation passed as dependency context only: the FE-P5 map has all five rows implemented with no blockers, the FE-P5 ledger is generated from the FE-P5 map, and the FE-P5 handoff says `FE-P5 phase complete`.
- The frontend guide, FE-P6 map, and FE-P6 generated ledger all enumerate exactly `FE-U-P6-01`, `FE-I-P6-01`, `FE-E-P6-01`, `FE-V-P6-01`, and `FE-A11Y-P6-01`.
- FE-P6 owner source paths exist. Product row `REQ-*` IDs and Core AC IDs resolve in the cited Core/spec corpus; design/support AC IDs resolve in `docs/design.md` or `docs/guides/`.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p6_coverage_ledger.md` remains a generated downstream ledger and is not closure evidence. No FE-P6 ledger hand edit was detected by `git diff -- docs/testing/frontend_phase_coverage_ledgers/fe_p6_coverage_ledger.md`, and ledger drift passed.
- `tools/frontend_visual_fixture_registry.json` contains a single `FE-VFIX-05` fixture with title `Evidence affordance`, status `current`, owner phase `FE-P6`, and owner row `FE-V-P6-01`. Registry status and golden filenames remain non-closure evidence.
- Latest retained target artifacts from 2026-06-06 for `frontend-unit`, `browser-e2e-webserver-backed`, `browser-e2e-stateful`, and `browser-e2e-visual` passed but did not include FE-P6 row results. Older FE-P6 retained row-accounting artifacts are stale or blocked-context only and must not close rows.
- Latest retained `browser-e2e-a11y-preflight` summary `.cartulary/test-results/20260606T164059Z-p1264123/browser-e2e-a11y-preflight/accessibility-preflight/frontend-accessibility-preflight-summary.json` passed and listed `FE-A11Y-P6-01`, but the row remains blocked and `required_for_closure=false`. This is blocked-row smoke only.
- Latest retained broad `make check` artifact exists and passed, but broad `make check` remains repository-gate context only and cannot close FE-P6 rows.

Sprint 1 row inventory and blocker status:

| Row | Evidence class | Claim status | Target mapping | Sprint 1 blocker |
| --- | --- | --- | --- | --- |
| `FE-U-P6-01` | `product_conformance` | `blocked` | `make frontend-unit` | `frontend_phase_row_not_implemented` |
| `FE-I-P6-01` | `product_conformance` | `blocked` | `make frontend-unit`; `make browser-e2e-webserver-backed` | `frontend_phase_row_not_implemented` |
| `FE-E-P6-01` | `product_conformance` | `blocked` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | `frontend_phase_row_not_implemented` |
| `FE-V-P6-01` | `design_direction` | `blocked` | `make browser-e2e-visual` | `visual_fixture_not_recaptured_for_frontend_row` |
| `FE-A11Y-P6-01` | `design_direction` | `blocked` | `make browser-e2e-a11y-preflight` | `frontend_phase_row_not_implemented`; preflight-only evidence is not closure |

Sprint 1 exact blockers:

`BLOCKER: FE-P6 row remains blocked; row=FE-U-P6-01 target=frontend-unit reason=frontend_phase_row_not_implemented minimum_follow_up=implement current row-owned evidence lifecycle view-model and count-display unit coverage.`

`BLOCKER: FE-P6 row remains blocked; row=FE-I-P6-01 target=frontend-unit,browser-e2e-webserver-backed reason=frontend_phase_row_not_implemented minimum_follow_up=implement attach flow, generated protocol/public envelope handling, stable selectors, and raw-handle prevention evidence.`

`BLOCKER: FE-P6 row remains blocked; row=FE-E-P6-01 target=browser-e2e-webserver-backed,browser-e2e-stateful reason=frontend_phase_row_not_implemented minimum_follow_up=collect server-backed public /api/v1 same-origin preview/download handle and authorization-denial evidence.`

`BLOCKER: FE-P6 row remains blocked; row=FE-V-P6-01 target=browser-e2e-visual reason=visual_fixture_not_recaptured_for_frontend_row minimum_follow_up=recapture row-owned FE-VFIX-05 visual readiness evidence; do not use registry status or golden filename alone.`

`BLOCKER: FE-P6 row remains blocked; row=FE-A11Y-P6-01 target=browser-e2e-a11y-preflight reason=frontend_phase_row_not_implemented minimum_follow_up=map and run implemented-row accessibility evidence if row is promoted.`

`BLOCKER: FE-P6 accessibility evidence is preflight-only; row=FE-A11Y-P6-01 minimum_follow_up=map and run implemented-row accessibility target if row is promoted.`

### Sprint 2: Unit

Goal: create direct current unit evidence for `FE-U-P6-01`.

Tasks:

- Implement or refine evidence lifecycle view-model vocabulary for count-display, requested, pending upload, available, preview blocked, failed, inconsistent, blocked, and public error states.
- Keep object blob state and evidence record state separate according to Core 02 Section 13.
- Add unit tests that assert state distinctions, count behavior, and blocked/failed/inconsistent rendering inputs.
- Use stable selectors or view-model keys that can be reused by integration and browser tests.
- Avoid claims about public route behavior from unit tests.

Validation:

- `make frontend-unit`
- `make frontend-typecheck` if typed surfaces changed
- `make frontend-import-boundary-check` if shared imports changed

Exit: `FE-U-P6-01` can be considered for promotion only if row accounting and current mapped unit evidence directly own the row.

Sprint 2 close-out, 2026-06-06:

- Binary result: `FE-U-P6-01` is promoted to row-accounting-owned unit evidence and closed by current mapped `frontend-unit` scenarios. This is not FE-P6 completion, registry activation, Core 05 readiness, public-route conformance, visual readiness, accessibility readiness, or claim-publication readiness.
- Implemented authored app-local evidence lifecycle view-model vocabulary in `apps/web/src/evidenceLifecycleViewModel.ts`, including separate `EvidenceRecordLifecycleState`, `ObjectBlobUploadState`, `EvidenceLifecycleViewStateKey`, and `EvidenceCountDisplayStateKey` vocabularies.
- Added stable helper surfaces: `buildEvidenceLifecycleViewModel`, `buildEvidenceCountDisplayViewModel`, and `summarizeEvidenceLifecycleCounts`.
- Updated `apps/web/src/WorkbookShell.tsx` to consume the new helpers for Evidence-surface access controls and Timeline evidence count display, while preserving existing stable selector builders and adding reusable `data-evidence-state-key` and `data-evidence-count-state` attributes.
- Added `apps/web/src/evidenceLifecycleViewModel.test.ts` with current row-named unit scenarios for `FE-U-P6-01`: requested/pending upload/available/preview-blocked distinctions, failed/blocked/inconsistent/public-error rendering inputs, count contribution behavior, and count-display projection consistency.
- Promoted `FE-U-P6-01` in `tools/frontend_phase_maps/fe_p6_test_map.json` to `claim_status=implemented`, required scenario-backed `frontend-unit` closure, and the four exact Vitest scenario titles.
- Regenerated `docs/testing/frontend_phase_coverage_ledgers/fe_p6_coverage_ledger.md` through `make phase-ledgers`; `tools/frontend_phase_registry.json` now keeps `FE-P6` planned with activation blockers while moving `row_rollup_state` to `partially_implemented`.
- Fresh `make frontend-unit` passed with run root `.cartulary/test-results/20260607T001424Z-p2011519`; `frontend-unit/frontend-row-accounting.json` records `FE-U-P6-01` with `claim_status_at_run=implemented`, `target_mapping_status=mapped`, and `closure_status=closed`.
- `make json-shape-check`, `make phase-ledger-drift`, and `make phase-schedule-drift` passed after promotion; schedule regeneration was not required.
- `make frontend-typecheck` was skipped for the promotion because no typed frontend source changed.
- `make frontend-import-boundary-check` was skipped because no shared package imports, selector package boundaries, or package-boundary rules changed.

Sprint 2 promotion result:

`CLOSED: FE-U-P6-01 unit row is promoted and closed by current mapped unit evidence. Remaining FE-P6 blockers are FE-I-P6-01, FE-E-P6-01, FE-V-P6-01, and FE-A11Y-P6-01.`

### Sprint 3: Integration

Goal: create direct current integration evidence for `FE-I-P6-01`.

Tasks:

- Wire evidence attach flow to generated protocol types through `packages/protocol-ts`.
- Render public success and public error envelopes without raw backend details.
- Use selectors from `packages/ui-contracts`; update authored selector sources only when needed.
- Add negative checks for raw object URLs, raw paths, raw object keys, bucket names, backend paths, storage backend identifiers, and object-store implementation details in user-facing access handles.
- Verify attach, blocked, failed, inconsistent, preview, and download controls keep stable identifiers.
- Ensure frontend-only mocks are used only for support or view-state coverage, not public-route product closure.

Validation:

- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-support` when shared helpers or selectors change
- `make frontend-import-boundary-check`
- `make generated-artifact-policy-check` and `make generate-drift` when generated or contract surfaces are touched

Exit: `FE-I-P6-01` can be considered for promotion only from direct current mapped evidence with the exact FE-I-P6 scenario expectation and raw-handle prevention checks.

Sprint 3 close-out, 2026-06-07:

- Binary result: `FE-I-P6-01` is promoted by current row-owned evidence from mapped `frontend-unit` and `browser-e2e-webserver-backed` targets. This is not full FE-P6 completion, registry activation, Core 05 readiness, `FE-E-P6-01` same-origin authorization-denial closure, visual readiness, accessibility readiness, or claim-publication readiness.
- `apps/web/src/WorkbookShell.tsx` consumes the `@cartulary/protocol-ts` facade for evidence object-blob create, attach, preview-handle, and download-handle request/envelope typing. The attach and handle request bodies use `satisfies ObjectBlobCreateRequest`, `satisfies EvidenceAttachBlobRequest`, and `satisfies EvidenceHandleIssueRequest`.
- `packages/protocol-ts/src/index.ts` exposes authored facade types for `ObjectBlobCreateRequest`, `ObjectBlobCreateEnvelope`, `EvidenceAttachBlobRequest`, `EvidenceAttachBlobEnvelope`, `EvidenceHandleIssueRequest`, `EvidenceHandleEnvelope`, and `ErrorEnvelope`; `packages/protocol-ts/src/index.test.ts` anchors those facade names to generated OpenAPI schema names.
- `apps/web/src/WorkbookShell.tsx` renders evidence public error output through `evidencePublicErrorMessage`, filtering raw evidence storage details from public text before display and falling back to safe public code, reason, status, or generic text.
- Evidence preview and download handles are accepted only as same-origin `/api/v1/evidence-handles/...` paths through `resolvePublicEvidenceHandleHref`; raw URLs or backend/storage-looking values are rejected before reaching iframe or download anchors.
- Evidence attach, blocked, failed, inconsistent, preview, download, access-message, and preview-frame controls use stable selectors from `packages/ui-contracts`, including `evidenceAttachFileInputTestId`, `evidencePreviewButtonTestId`, `evidenceDownloadButtonTestId`, `evidenceAccessMessageTestId`, and `evidencePreviewFrameTestId`.
- `apps/web/src/WorkbookShell.surfaces.test.tsx` provides frontend view-state coverage for attach, blocked, failed, inconsistent, raw-handle rejection, public-error rendering, preview, and download controls. Its frontend-only mocks are support/view-state evidence only and are not used as public-route closure by themselves.
- `apps/web/e2e/phase6.evidence-integration.spec.ts` provides server-backed browser evidence for the mapped FE-I scenario through public `/api/v1/object-blobs`, `/api/v1/object-uploads/...`, `/api/v1/evidence-records/{record_id}/attach-blob`, `/api/v1/evidence-records/{record_id}/preview-handle`, `/api/v1/evidence-records/{record_id}/download-handle`, and `/api/v1/evidence-handles/...` routes.
- Raw-handle prevention checks covered raw object URLs, raw paths, raw object keys, object-store keys, bucket names, backend paths, storage backend identifiers, and object-store implementation details across rendered body text, access messages, preview frame source, download URL, public handle envelope hrefs, and observed evidence-route request URLs.
- Fresh `make frontend-unit` passed with run root `.cartulary/test-results/20260607T012408Z-p2130739`; `frontend-unit/frontend-row-accounting.json` records `FE-I-P6-01` with `claim_status=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the exact mapped FE-I scenario.
- Fresh `make browser-e2e-webserver-backed` passed with run root `.cartulary/test-results/20260607T012432Z-p2133593`; `browser-e2e-webserver-backed/frontend-row-accounting.json` records `FE-I-P6-01` with `claim_status=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the exact mapped FE-I scenario.
- Fresh `make browser-e2e-support` passed with run root `.cartulary/test-results/20260607T013022Z-p2153888`; this is support-only helper/selector evidence and is not product-conformance closure evidence.
- Fresh `make frontend-import-boundary-check` passed with run root `.cartulary/test-results/20260607T012408Z-p2130738`.
- Fresh `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260607T012408Z-p2130754`.
- Fresh `make generate-drift` passed with run root `.cartulary/test-results/20260607T012408Z-p2130784`; no generated artifact was regenerated for Sprint 3 audit closure.
- `make agent-finalize` was skipped for the Sprint 3 audit because it can mutate phase ledgers and schedules and the audit was constrained to avoid product code, generated files, phase maps, ledgers, selectors, tests, and documentation edits unless stale generated artifacts were found. Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

Sprint 3 promotion result:

`CLOSED: FE-I-P6-01 integration row is promoted and closed by current mapped unit and server-backed browser evidence. Remaining FE-P6 blockers are FE-E-P6-01, FE-V-P6-01, and FE-A11Y-P6-01.`

### Sprint 4: Browser E2E

Goal: create direct current browser evidence for `FE-E-P6-01`.

Tasks:

- Exercise evidence attach through public `/api/v1/` routes.
- Exercise preview handle issuance and same-origin redemption through public routes.
- Exercise download handle issuance and same-origin redemption through public routes.
- Exercise blocked preview through public error envelopes.
- Exercise authorization denial with current server-managed sessions and authorization re-derived at request time.
- Verify handles are opaque same-origin browser-facing hrefs, not raw object-store URLs, bucket names, raw keys, backend paths, or storage identifiers.
- Verify preview/download controls do not navigate away from the active workbook surface unless the route contract explicitly requires it.
- Use stable identifiers and exact FE-E-P6 scenario title expectations from the map.

Validation:

- `make browser-e2e-webserver-backed`
- `make browser-e2e-stateful`
- `make frontend-typecheck` when route-facing frontend types changed
- `make agent-finalize` near closure

Exit: `FE-E-P6-01` can be considered for promotion only when public-route, same-origin handle, current authorization, public envelope, stable identifier, and raw-handle negative evidence all exist in mapped row-owned artifacts.

Sprint 4 close-out, 2026-06-07:

- Binary result: `FE-E-P6-01` is promoted and closed by current row-owned browser evidence from mapped `browser-e2e-webserver-backed` and `browser-e2e-stateful` targets. This is not full FE-P6 completion, registry activation, Core 05 readiness, visual readiness, accessibility readiness, or claim-publication readiness.
- `apps/web/e2e/phase6.evidence-integration.spec.ts` now contains the exact mapped scenario `FE-E-P6-01 Verify evidence attach, preview, download, blocked preview, and authorization denial through same-origin public handles.`
- The scenario exercises the public browser-facing route path through `/api/v1/object-blobs`, `/api/v1/object-uploads/...`, `/api/v1/evidence-records/{record_id}/attach-blob`, `/api/v1/evidence-records/{record_id}/preview-handle`, `/api/v1/evidence-records/{record_id}/download-handle`, and `/api/v1/evidence-handles/...`.
- Safe preview evidence is redeemed through a same-origin preview handle in an iframe; download evidence is redeemed through a same-origin download handle while the active Evidence workbook surface remains selected.
- Blocked preview evidence uses public `409 evidence_access_unavailable` with `details.reason_code=unsupported_preview`, not a frontend-only mock.
- Current authorization denial is verified with a live server-managed member session after incident membership removal: post-removal redemption of a previously issued preview handle returns public `404 handle_not_found_or_revoked`, and post-removal handle issuance returns public `404 evidence_record_not_found`.
- Raw-handle negative checks cover browser-facing upload targets, preview handles, download handles, iframe sources, download URLs, public error bodies, visible access messages, and observed route URLs, ensuring no raw object-store URLs, bucket names, raw keys, backend paths, storage backend identifiers, or object-store implementation details are exposed as access handles.
- Stable selector coverage uses existing `packages/ui-contracts` evidence selectors for attach input, preview button, download button, access message, preview frame, and Evidence grid shell. No authored selector inputs or generated selector outputs changed.
- `tools/frontend_phase_maps/fe_p6_test_map.json` now records `FE-E-P6-01` as `claim_status=implemented`, clears the prior `frontend_phase_row_not_implemented` blocker, and marks both mapped browser targets `required_for_closure=true`.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p6_coverage_ledger.md` was regenerated through `make phase-ledgers`; `tools/frontend_phase_registry.json` FE-P6 manifest, ledger, and evidence-freshness digests were updated while FE-P6 remains planned and `partially_implemented`.
- Fresh `make browser-e2e-webserver-backed` passed with run root `.cartulary/test-results/20260607T022152Z-p2218718`; `browser-e2e-webserver-backed/frontend-row-accounting.json` records `FE-E-P6-01` with `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the exact mapped FE-E scenario.
- Fresh `make browser-e2e-stateful` passed with run root `.cartulary/test-results/20260607T022733Z-p2234323`; `browser-e2e-stateful/frontend-row-accounting.json` records `FE-E-P6-01` with `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the exact mapped FE-E scenario.
- Fresh `make json-shape-check`, `make phase-ledger-drift`, and `git diff --check` passed after map, registry, and ledger updates.
- Fresh `make agent-finalize` passed with run root `.cartulary/test-results/20260607T023351Z-p2248130`; retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- `make frontend-typecheck` was skipped because no route-facing frontend types changed. Selector and contract drift checks were skipped because no authored selector or contract inputs changed.

Sprint 4 row-owned audit, 2026-06-07:

- Binary recommendation: `PROMOTE FE-E-P6-01`.
- Audit scope was limited to `FE-E-P6-01` Sprint 4 product-conformance evidence. The audit did not count FE-P6 generated ledger posture, support checks, visual readiness, accessibility readiness, broad `make check`, or Core 05 claim-publication evidence as product-conformance closure.
- The FE-P6 phase map still contains exactly five rows and records `FE-E-P6-01` as `claim_status=implemented`, `evidence_class=product_conformance`, with no blockers, exact scenario title `FE-E-P6-01 Verify evidence attach, preview, download, blocked preview, and authorization denial through same-origin public handles.`, and required mapped targets `make browser-e2e-webserver-backed` and `make browser-e2e-stateful`.
- The generated FE-P6 ledger reflects the map as downstream posture only. The row-owned browser artifacts, not the ledger, are the closure evidence.
- The audited scenario in `apps/web/e2e/phase6.evidence-integration.spec.ts` verifies browser execution through public `/api/v1/` routes for object-blob creation, upload, evidence attach, preview-handle issuance, download-handle issuance, blocked preview, evidence-handle redemption, and authorization denial.
- Same-origin opaque handle evidence covers preview iframe sources, download hrefs, public handle envelope hrefs, and observed route URLs. The audit found no evidence of raw object-store URLs, bucket names, raw keys, backend paths, storage backend identifiers, or object-store implementation details in the row-owned negative checks.
- Blocked preview evidence remains fail-closed through the public `409 evidence_access_unavailable` envelope with `details.reason_code=unsupported_preview` and no silent fallback to download.
- Current authorization denial is covered by a live server-managed member session after incident membership removal: previously issued preview-handle redemption returns public `404 handle_not_found_or_revoked`, and post-removal handle issuance returns public `404 evidence_record_not_found`.
- Fresh `make browser-e2e-webserver-backed` passed with run root `.cartulary/test-results/20260607T140632Z-p2661394`; `browser-e2e-webserver-backed/frontend-row-accounting.json` records `FE-E-P6-01` with `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the exact mapped FE-E scenario.
- Fresh `make browser-e2e-stateful` passed with run root `.cartulary/test-results/20260607T141212Z-p2676334`; `browser-e2e-stateful/frontend-row-accounting.json` records `FE-E-P6-01` with `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the exact mapped FE-E scenario.
- Fresh support/posture checks passed after the audit: `make generated-artifact-policy-check` with run root `.cartulary/test-results/20260607T141900Z-p2689873`, `make json-shape-check` with run root `.cartulary/test-results/20260607T141900Z-p2689869`, `make phase-ledger-drift` with run root `.cartulary/test-results/20260607T141900Z-p2689884`, and `make phase-schedule-drift` with run root `.cartulary/test-results/20260607T141900Z-p2689907`.
- Fresh `make agent-finalize` passed with run root `.cartulary/test-results/20260607T141908Z-p2690688`; generated outputs were unchanged. Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- `make frontend-typecheck` was not run for the audit because no route-facing frontend types changed. Broad `make check` was not run because the audit used the two mapped row-owned browser targets and narrow generated/posture checks.

Sprint 4 promotion result:

`CLOSED: FE-E-P6-01 browser E2E row is promoted and closed by current mapped server-backed public-route evidence. Remaining FE-P6 blockers are FE-V-P6-01 and FE-A11Y-P6-01.`

### Sprint 5: Visual

Goal: create design-direction readiness evidence for `FE-V-P6-01`.

Tasks:

- Verify `FE-VFIX-05` fixture identity and owner row remain unambiguous.
- Capture evidence count, affordance, available, requested, pending, blocked, failed, inconsistent, preview, and download-handle state fixtures.
- Keep visual readiness separate from product conformance.
- Record fixture identity and row accounting, not just golden filename or registry status.
- Follow visual golden maintenance rules for any recapture.

Validation:

- `make browser-e2e-visual`
- `make json-shape-check` when fixture registries or schema-shaped artifacts change
- `make phase-ledger-drift` after map/ledger-affecting changes

Exit: `FE-V-P6-01` can be considered for design-readiness promotion only from row-owned visual accounting tied to `FE-VFIX-05`; fixture registry `current` and golden files alone are insufficient.

### Sprint 6: Accessibility

Goal: create accessibility readiness evidence for `FE-A11Y-P6-01` without confusing preflight smoke with closure.

Tasks:

- Verify evidence icon buttons, blocked states, error states, preview controls, and download controls have accessible names.
- Verify focus behavior, contrast readiness, and non-color-only distinctions.
- Treat `make browser-e2e-a11y-preflight` as blocked-row smoke only while the row remains mapped to preflight and `required_for_closure=false`.
- If the row is promoted or remapped for implemented accessibility closure, run the implemented accessibility target and require row-owned summary evidence.
- Keep accessibility readiness separate from product conformance.

Validation:

- `make browser-e2e-a11y-preflight` for blocked-row smoke only
- `make browser-e2e-a11y` only if the implemented accessibility row is mapped or promoted
- `make frontend-unit` or `make browser-e2e-support` when helper coverage changes

Exit: `FE-A11Y-P6-01` can be considered for design-readiness promotion only if mapped implemented accessibility evidence exists. Preflight smoke alone is not closure.

### Sprint 7: Closure

Goal: close FE-P6 only after direct row-owned evidence, drift checks, dependency validation, and handoff notes are complete.

Tasks:

- Verify every FE-P6 row has direct current row-owned evidence in mapped targets.
- Verify product rows have resolved Core or adopted NLSpec owner references.
- Verify browser product rows use public `/api/v1/` routes, same-origin handle redemption, stable identifiers, current authorization, public success envelopes, and public error envelopes.
- Verify raw object paths, URLs, keys, bucket names, backend paths, storage backend identifiers, and object-store implementation details are absent from user-facing handles.
- Verify visual and accessibility rows remain design-direction only.
- Run `make phase-ledgers` only after authored phase-map changes, then run `make phase-ledger-drift`.
- Run `make phase-schedule-drift` only when schedules are affected.
- Run `make generated-artifact-policy-check` and `make generate-drift` when generated or contract surfaces are touched.
- Verify FE-P0 through FE-P5 remain green or have exact owner-accepted blockers that do not invalidate FE-P6.
- Run `make agent-finalize`; run broad `make check` only when repository completion rules require it.
- Write FE-P7 handoff with exact row states, evidence roots, blockers, and non-claims.

Exit: FE-P6 completion can be claimed only if every binary completion criterion below is satisfied.

## Validation Commands

Target availability was verified with `make explain-target TARGET=<target> DETAIL=summary` for each listed target. Planned FE-P6 phase status was verified with `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P6`.

Plan-only validation:

- `git diff --check`
- `make phase-ledger-drift`

Readiness and discovery:

- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P6`
- `make explain-target TARGET=<target> DETAIL=summary`

Product implementation validation:

- `make frontend-typecheck`
- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-stateful`
- `make browser-e2e-support` when shared helpers or selectors change
- `make frontend-import-boundary-check`

Design readiness validation:

- `make browser-e2e-visual`
- `make browser-e2e-a11y-preflight` for blocked-row smoke only
- `make browser-e2e-a11y` only if the implemented accessibility row is mapped or promoted to that target

Generated, map, and drift validation:

- `make generated-artifact-policy-check` when generated or contract surfaces are touched
- `make generate-drift` when generated or contract surfaces are touched
- `make phase-ledgers` after authored phase-map changes
- `make phase-ledger-drift`
- `make phase-schedule-drift` when schedules are affected
- `make json-shape-check` when manifests, registries, or schema-shaped artifacts change

End-of-run validation:

- `make agent-finalize`
- `make check` only when repository completion rules require the broad developer gate
- `git diff --check`

Commands not required solely for this authored plan:

- `make phase-ledgers`, because this task does not change authored phase maps.
- `make generate-drift`, because this task does not touch generated or contract surfaces.
- `make generated-artifact-policy-check`, because this task does not touch generated roots or generated policy inputs.
- `make json-shape-check`, because this task does not edit manifests, registries, or schema-shaped artifacts.
- `make phase-schedule-drift`, because this task does not affect schedules.
- Row-owned product, visual, or accessibility targets, because this task creates a plan only and does not implement or promote FE-P6 rows.

## Evidence Requirements

General requirements:

- Evidence must be current, direct, and row-owned.
- Evidence must come from mapped targets in `tools/frontend_phase_maps/fe_p6_test_map.json`.
- Scenario-title expectations must match the map for rows that declare them.
- Owner references must resolve to Core or adopted NLSpec sources for product-conformance rows.
- Generated ledgers must be regenerated through Make after map changes and must not be hand edited.
- Support-only targets can support implementation quality but cannot close product rows.
- Broad `make check` can satisfy a repository gate but cannot close FE-P6 rows.

Product row requirements:

- `FE-U-P6-01` needs current unit evidence for evidence lifecycle view models and count-display states.
- `FE-I-P6-01` needs generated protocol, public envelope, selector, attach flow, and raw-handle prevention evidence.
- `FE-E-P6-01` has Sprint 4 browser evidence through public `/api/v1/` routes, same-origin handle redemption, current authorization denial, stable identifiers, public envelopes, and raw-handle negative checks.
- Browser-facing product evidence must use server-managed sessions and current authorization re-derived at request time.
- Frontend-only mocks cannot close `FE-I-P6-01` or `FE-E-P6-01` public-route behavior.

Design row requirements:

- `FE-V-P6-01` needs row-owned visual accounting tied to `FE-VFIX-05`; registry status and golden filenames are not enough.
- `FE-A11Y-P6-01` needs mapped row-owned accessibility evidence if promoted; preflight smoke alone is not enough.
- Visual and accessibility readiness remain design-direction evidence only.

Raw handle prevention requirements:

- User-facing preview and download access handles must not expose raw object URLs.
- User-facing preview and download access handles must not expose raw object-store keys.
- User-facing preview and download access handles must not expose bucket names.
- User-facing preview and download access handles must not expose backend paths.
- User-facing preview and download access handles must not expose storage backend identifiers.
- User-facing preview and download access handles must not expose object-store implementation details.
- Preview/download must route through same-origin opaque handle redemption.

## Blocker Rules

Use blockers when a row, source, owner, target, fixture, generated artifact, retained artifact, or evidence class cannot support a valid FE-P6 claim.

Required blocker templates:

`BLOCKER: FE-P6 row remains blocked; row=<row_id> target=<target> reason=<reason_code> minimum_follow_up=<specific rerun, implementation, owner correction, or map correction>.`

`BLOCKER: FE-P6 product row lacks public /api/v1 same-origin evidence handle coverage; row=<row_id> minimum_follow_up=collect server-backed browser evidence through public routes.`

`BLOCKER: FE-P6 evidence handle exposes raw object URL, raw object key, bucket name, backend path, or storage backend identifier; minimum_follow_up=route preview/download through same-origin opaque handle redemption.`

`BLOCKER: FE-P6 authorization denial tested only through frontend mock; row=<row_id> minimum_follow_up=rerun with current server-managed session and current authorization re-derived at request time.`

`BLOCKER: FE-P6 visual fixture identity missing or ambiguous; row=FE-V-P6-01 expected=<fixture_id_from_registry> actual=<fixture_ids> minimum_follow_up=<registry or map correction>.`

`BLOCKER: FE-P6 accessibility evidence is preflight-only; row=FE-A11Y-P6-01 minimum_follow_up=map and run implemented-row accessibility target if row is promoted.`

`BLOCKER: FE-P6 generated ledger hand edit detected; path=docs/testing/frontend_phase_coverage_ledgers/fe_p6_coverage_ledger.md minimum_follow_up=revert hand edit, update owner map, run make phase-ledgers.`

`BLOCKER: FE-P6 direct react-data-grid import outside /packages/grid-adapter; path=<path> minimum_follow_up=route grid usage through adapter and rerun frontend-import-boundary-check.`

`BLOCKER: FE-P5 handoff validation missing or stale; minimum_follow_up=<rerun dependency validation or record owner-accepted rationale>.`

`BLOCKER: FE-P6 evidence classes collapsed; product/design/support/claim-publication evidence cannot be counted across classes.`

Additional FE-P6 blocker cases:

- `BLOCKER: FE-P6 guide/map row mismatch detected; guide_rows=<rows> map_rows=<rows> minimum_follow_up=correct owner source or phase map before implementation.`
- `BLOCKER: FE-P6 owner reference cannot be verified; row=<row_id> owner=<owner_ref> minimum_follow_up=correct row owner reference or cite adopted owner source.`
- `BLOCKER: FE-P6 scenario title missing or stale; row=<row_id> expected=<title_from_map> actual=<observed_title> minimum_follow_up=align mapped scenario evidence or correct phase map.`
- `BLOCKER: FE-P6 retained artifact is stale or not row-owned; row=<row_id> artifact=<path> minimum_follow_up=rerun mapped target and collect current row-owned evidence.`
- `BLOCKER: FE-P6 frontend-only mock used as product-route evidence; row=<row_id> minimum_follow_up=collect server-backed public-route evidence.`
- `BLOCKER: FE-P6 Core 05 claim implied without claim-bearing metadata; minimum_follow_up=remove claim-publication language or satisfy Core 05 explicitly.`
- `BLOCKER: FE-P6 generated artifact stale after authored input change; input=<path> minimum_follow_up=run supported generator and drift target.`

## Strict Non-Claims

This plan does not claim:

- FE-P6 row closure from this plan.
- FE-P6 row closure from generated ledgers.
- FE-P6 row closure from old retained artifacts.
- FE-P6 row closure from broad `make check`.
- FE-P6 row closure from test names or scenario title text alone.
- FE-P6 row closure from support-only tests.
- FE-P6 row closure from visual golden files or fixture registry `current` status.
- FE-P6 accessibility readiness from `browser-e2e-a11y-preflight` blocked-row smoke.
- Product conformance from visual or accessibility evidence.
- Core 05 claim-publication readiness unless explicit claim-bearing publication metadata exists and satisfies Core 05.
- WebSocket live updates.
- Same-field conflict resolver implementation.
- Saved-view persistence.
- Full coordination surfaces.
- Raw object-store access as public evidence access behavior.
- That existing legacy evidence lifecycle code or legacy scenario titles close FE-P6 rows.
- That frontend-only mocks close public-route product-conformance rows.
- That current `FE-VFIX-05` registry status closes `FE-V-P6-01`.
- That this document is itself authority for FE-P6 activation, registry state, map status, generated ledger contents, or any row `claim_status`.

## Binary Exit Criteria

### 1. Initial Plan Creation

Initial plan creation was complete only if:

- `FRONTEND_PHASE6_IMPLEMENTATION_PLAN.md` exists at the repository root.
- The file contains the required stable headings.
- The file records current FE-P6 row status without using the document itself to promote any row.
- The file records source authority, source limits, non-claims, blocker rules, row inventory, validation commands, and FE-P7 handoff guidance.
- No generated ledger, generated protocol file, generated schedule, lockfile, or tool-managed artifact is hand edited.
- `git diff --check` passes.
- `make phase-ledger-drift` passes or any unrelated failure is recorded precisely.

### 2. Product Evidence Lifecycle Closure

Product evidence lifecycle closure is complete only if:

- `FE-U-P6-01`, `FE-I-P6-01`, and `FE-E-P6-01` close from direct current row-owned evidence in mapped targets.
- Product-conformance rows have resolved Core or adopted NLSpec owner references.
- Browser-facing product rows use public `/api/v1/` route evidence.
- Preview and download access use same-origin opaque handle redemption.
- Server-managed sessions are used.
- Authorization is current and re-derived at request time.
- Stable identifiers, public success envelopes, and public error envelopes are present.
- Raw object paths, raw object URLs, raw object-store keys, bucket names, backend paths, storage backend identifiers, and object-store implementation details are not exposed as user-facing access handles.
- Frontend-only mocks are not used to close public-route behavior.

### 3. Visual Readiness

Visual readiness is complete only if:

- `FE-V-P6-01` closes from direct current row-owned visual evidence in `make browser-e2e-visual`.
- Visual row accounting ties the evidence to `FE-VFIX-05`.
- Evidence covers evidence count, affordance, available, requested, pending, blocked, failed, inconsistent, preview, and download-handle state fixtures.
- Fixture identity is unambiguous.
- Visual evidence remains design-direction only.
- Fixture registry `current` status and golden filenames alone are not used as closure.

### 4. Accessibility Readiness

Accessibility readiness is complete only if:

- `FE-A11Y-P6-01` is mapped to implemented accessibility evidence or remains blocked with preflight smoke only.
- If promoted or remapped, row-owned evidence from the implemented accessibility target exists.
- Evidence covers evidence icon buttons, blocked states, error states, preview controls, download controls, accessible names, focus, contrast, and non-color-only distinctions.
- Accessibility evidence remains design-direction only.
- `browser-e2e-a11y-preflight` blocked-row smoke is not used as closure.

### 5. Full FE-P6 Phase Completion

Full FE-P6 phase completion is complete only if:

- Every FE-P6 row closes from direct current row-owned evidence.
- Product-conformance rows have resolved Core or adopted NLSpec owner references.
- Browser-facing product rows use public route evidence, same-origin handle redemption, stable identifiers, current authorization, public success envelopes, and public error envelopes.
- Raw object paths, raw object URLs, raw object-store keys, bucket names, backend paths, storage backend identifiers, and object-store implementation details are not exposed as user-facing access handles.
- Visual readiness and accessibility readiness remain design-direction only.
- Generated ledgers are regenerated through Make after map changes.
- Generated ledger drift checks pass.
- Frontend namespace, row accounting, and scenario-title requirements are satisfied.
- All triggered shared harnesses are satisfied or precisely blocked without invalidating the completion claim.
- FE-P0 through FE-P5 remain green or have exact owner-accepted blockers that do not invalidate FE-P6 completion.
- Generated-artifact policy and generated drift pass when generated or contract surfaces are touched.
- Core, design, support, and claim-publication evidence classes remain separate.
- No Core 05 claim-publication evidence is implied.

### 6. FE-P7 Handoff Readiness

FE-P7 handoff readiness is complete only if:

- FE-P6 final row statuses are recorded with exact target evidence or exact blockers.
- FE-P6 public-route evidence handle status is recorded.
- Raw-handle exposure negative checks are summarized.
- Visual readiness status and `FE-VFIX-05` identity are recorded without product-conformance claims.
- Accessibility readiness status is recorded, including whether preflight remained smoke-only or the row was promoted/remapped.
- FE-P0 through FE-P5 dependency state is recorded.
- Generated artifact, ledger, schedule, and drift status is recorded.
- Remaining blockers include exact minimum follow-up actions.
- No FE-P7 scope is implemented or claimed as part of FE-P6 closure.

## FE-P7 Handoff

The FE-P7 handoff should preserve enough evidence context for the next phase without turning FE-P6 artifacts into authority for FE-P7.

Required handoff contents:

- Final FE-P6 registry status and row inventory.
- Direct evidence roots for each closed FE-P6 row, if any.
- Exact blockers for every unclosed FE-P6 row.
- Confirmation that product-conformance evidence used public `/api/v1/` routes, same-origin handle redemption, server-managed sessions, current authorization, stable identifiers, and public envelopes.
- Confirmation that user-facing access handles did not expose raw object URLs, raw object keys, bucket names, backend paths, storage backend identifiers, or object-store implementation details.
- Confirmation that visual and accessibility evidence remained design-direction only.
- Confirmation that no Core 05 claim-publication readiness was implied.
- FE-P0 through FE-P6 dependency state.
- Drift and finalization command outcomes.
- Any owner-accepted blockers that FE-P7 must respect.

Handoff non-claims:

- FE-P7 cannot inherit FE-P6 row closure unless the FE-P7 map directly owns and reuses current evidence under its own row rules.
- FE-P7 cannot inherit FE-P6 generated ledger text as owner evidence.
- FE-P7 cannot inherit FE-P6 visual or accessibility readiness as product conformance.
- FE-P7 cannot inherit FE-P6 public evidence handle behavior for new surfaces without direct mapped evidence.
