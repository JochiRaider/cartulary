# Cartulary Frontend Implementation and Testing Guide

| Field | Value |
| --- | --- |
| Status | Derived frontend implementation-planning guide. This guide is executable planning and verification support only; it is not a product-behavior authority. |
| Authoritative sources | Core 00 through Core 04 under `docs/spec/` remain implementation-conformance authority. Core 05 governs only claim-bearing timed or fixture-sensitive publication. |
| Traceability source | Exact owner `REQ-*` and `AC-*` identifiers from Core 01 through Core 04, with Core 05 cited only for publication-boundary separation. |
| Scope | Frontend MVP delivery, package boundaries, phase sequencing, shared browser harnesses, frontend coverage-ledger expectations, and guide-level acceptance criteria. |
| Intended repository path | `docs/guides/cartulary_frontend_implementation_testing_guide.md` |

## 1. How To Read This Guide

### 1.1 Source hierarchy

Use this authority order for frontend implementation planning:

1. Adopted Cartulary NLSpecs explicitly adopted in the repository, including `docs/testing-harness-nlspec.md` for harness mechanics only.
2. Cartulary Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication boundaries.
4. `docs/domain.md` for vocabulary and domain-boundary interpretation.
5. `docs/guides/cartulary_implementation_testing_guide.md` as the structural model for sequencing, test-row mapping, shared harnesses, completion rules, and coverage-ledger shape.
6. `docs/guides/cartulary-dev-guide.md` for repo-local frontend package boundaries, dependency boundaries, generated-artifact policy, Make targets, workspace shape, and frontend implementation baseline.
7. `docs/guides/cartulary-ui-ux-design-guide.md` and `docs/design.md` for design-direction constraints, shell composition, density, visual language, accessibility posture, and UI acceptance criteria.
8. `docs/guides/cartulary_visual_golden_maintenance.md` for visual-regression maintenance procedure.
9. Research reports only as rationale or risk evidence.

When sources differ, follow this hierarchy. When two owner sections appear to conflict, record the contradiction as a corpus defect and do not pick a side. When an owner reference cannot be located, write `TODO: owner lookup required` and exclude the affected row from authoritative completion until resolved.

### 1.2 Conformance posture

This guide must not change requirement ownership, route semantics, data-model rules, security boundaries, or conformance scope.

Core 00 through Core 04 remain the implementation-conformance authority. Core 05 governs only claim-bearing timed or fixture-sensitive publication. UI/UX guide language and `docs/design.md` govern design-direction interpretation and reviewer discipline only. `docs/domain.md` is a vocabulary and concept-boundary aid, not an API reference, schema reference, implementation guide, or test plan. Research reports are evidence and rationale, not runtime behavior owners.

### 1.3 Frontend guide purpose

This guide groups frontend work into executable phases, maps each phase to test rows, identifies shared harnesses, and defines binary completion rules. It is the frontend counterpart to the current progressive implementation and testing guide, but it owns only this guide-local planning shape.

The phase numbers here are local to this frontend guide. They must not be appended to, substituted for, or backfilled into the backend/base implementation guide unless a future owner document explicitly adopts that change.

### 1.4 Relationship to Core 00 through Core 05

Core 00 defines document status, precedence, and the distinction between implementation conformance and publication claims. Core 01 owns application boundaries, public HTTP and WebSocket routes, record/view contracts, saved-view contracts, evidence handles, generated protocol surfaces, and session/auth route behavior. Core 02 owns domain model behavior for incidents, entities, mentions, evidence, parties, rollback history, and vocabulary constraints. Core 03 owns workbook interaction behavior, system views, saved views, startup surface selection, editing, conflicts, coordination surfaces, and grid interaction invariants. Core 04 owns security, authorization, session, CSRF, WebSocket origin, and evidence access security constraints.

Core 05 does not expand frontend conformance. It applies when visual, timing, benchmark, fixture-sensitive, or performance evidence is published as a claim. Visual and responsiveness rows in this guide are implementation-quality evidence unless Core 05 publication predicates are separately satisfied.

### 1.5 Relationship to implementation, development, design, visual, and harness guides

`docs/guides/cartulary_implementation_testing_guide.md` is the structural model for this guide's row discipline, harness discipline, completion checklist, and coverage-ledger expectations.

`docs/guides/cartulary-dev-guide.md` defines current repo-local frontend package boundaries. In particular, `/apps/web` owns workbook shell, controllers, query state, HTTP mutation submission, pending replay, WebSocket refresh behavior, conflicts, inspector, and presence UI. `/apps/web` must consume `/packages/grid-adapter` and must not import `react-data-grid` directly. `/packages/grid-adapter` owns all direct `react-data-grid` imports, stylesheet import, row and cell identity translation, renderer/editor wiring, focus, selection, sorting, paste, fill, grouping, and imperative API containment. `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/test-utils`, and `/packages/protocol-ts` keep the responsibilities defined in the development guide; `/packages/ui` is reserved until a manifest-backed reusable presentational component package is introduced.

`docs/guides/cartulary-ui-ux-design-guide.md` and `docs/design.md` supply design-direction evidence for reviewer discipline, shell composition, density, typography, status patterns, visual affordances, and accessibility posture. Those documents do not create product-conformance behavior unless the same behavior is owned by Core 00 through Core 04 or by an adopted NLSpec.

`docs/guides/cartulary_visual_golden_maintenance.md` supplies the visual-regression maintenance procedure and the current Playwright golden target. The adopted `docs/testing-harness-nlspec.md` owns harness mechanics, but it does not define Core product behavior or Core 05 claim-publication evidence.

### 1.6 Source limits discovered during local inspection

- `docs/testing-harness-nlspec.md` is adopted current authority for harness mechanics only.
- Frontend-local phase readiness MUST use `tools/frontend_phase_registry.json`, `tools/frontend_phase_maps/fe_p0_test_map.json` through `tools/frontend_phase_maps/fe_p11_test_map.json`, and generated ledgers under `docs/testing/frontend_phase_coverage_ledgers/fe_p0_coverage_ledger.md` through `docs/testing/frontend_phase_coverage_ledgers/fe_p11_coverage_ledger.md`. Rows remain blocked from frontend phase completion until the registry, maps, ledgers, and drift checks agree.
- Frontend accessibility readiness MUST use `make browser-e2e-a11y` and the retained `cartulary.frontend_accessibility_summary.v1` artifact. Accessibility rows remain blocked from frontend phase completion until the target exists, emits the summary artifact, and maps every `FE-A11Y-*` row to an executed scenario.
- `/packages/ui` is future-reserved for reusable presentational components and is not an active FE-P0 package surface until a real manifest-backed package is introduced.
- Existing visual coverage under `apps/web/e2e/workbook.visual.spec.ts` does not yet include every frontend fixture required by this guide.
- No repository-local docs index was found that requires this new guide to be listed.

## 2. Test Categories

### 2.1 Test ID grammar

Frontend rows use guide-local IDs that must not collide with existing backend/base implementation rows:

| Prefix | Layer |
| --- | --- |
| `FE-U-P<phase>-NN` | Frontend unit test |
| `FE-I-P<phase>-NN` | Frontend integration test |
| `FE-B-P<phase>-NN` | Browser component or browser integration test |
| `FE-E-P<phase>-NN` | Full browser E2E test |
| `FE-V-P<phase>-NN` | Visual-regression test |
| `FE-A11Y-P<phase>-NN` | Accessibility test |
| `FE-S-P<phase>-NN` | Support, tooling, drift, or manifest test |

Every phase row must use this table shape:

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |

Allowed evidence classes are:

| Evidence class | Meaning |
| --- | --- |
| `product_conformance` | The row verifies behavior owned by Core 00 through Core 04 or by a future adopted NLSpec. |
| `design_direction` | The row verifies UI/UX guide or `docs/design.md` direction only. It must not count as product conformance. |
| `implementation_support` | The row verifies tooling, package boundaries, harnesses, generated artifacts, selectors, visual maintenance, or other implementation support. |
| `claim_publication_boundary` | The row checks that visual, timing, performance, fixture-sensitive, or benchmark evidence is not published as a claim unless Core 05 is separately satisfied. |
| `TODO_owner_lookup` | The row needs owner evidence. It must not count toward authoritative completion. |

#### Manifest row metadata

Frontend phase maps MUST preserve Core ownership separately from design and support ownership. A row object in `tools/frontend_phase_maps/fe_p*_test_map.json` MUST include `core_req_ids[]`, `core_ac_ids[]`, and `support_or_design_ac_ids[]`.

| Evidence class | `core_req_ids[]` | `core_ac_ids[]` | `support_or_design_ac_ids[]` | Completion rule |
| --- | --- | --- | --- | --- |
| `product_conformance` | non-empty | non-empty | optional | Counts only when Core 00 through Core 04 or an adopted NLSpec owns the behavior. |
| `design_direction` | empty | empty | non-empty | Readiness evidence only; MUST NOT be represented as product conformance. |
| `implementation_support` | empty | empty | non-empty | Support evidence only; MUST NOT be represented as product conformance. |
| `claim_publication_boundary` | Core 05 IDs allowed | Core 05 `PC-*` IDs allowed | optional | Active only for claim-publication boundary checks. |
| `TODO_owner_lookup` | empty | empty | optional | MUST use `claim_status="blocked"` and MUST be excluded from completion. |

`TODO_owner_lookup` MUST be used only when owner evidence is genuinely unresolved. A row that intentionally cites UI/UX, design, visual-maintenance, development-guide, or guide-local support criteria MUST use `design_direction` or `implementation_support` and MUST NOT carry generic owner-lookup text.

#### Guide-local support acceptance criteria

The following support criteria are repository-readiness criteria. They do not define product behavior.

| ID | Criterion |
| --- | --- |
| `FE-SUPPORT-AC-001` | `/apps/web` MUST NOT import `react-data-grid` directly; runtime grid use MUST pass through `/packages/grid-adapter`. |
| `FE-SUPPORT-AC-002` | The RDG stylesheet MUST be owned by `/packages/grid-adapter` and imported exactly once before workbook-grid rendering. |
| `FE-SUPPORT-AC-003` | Mutation-capable grid row identity MUST be asserted through `record_id`; presentation rows MUST NOT emit mutation events. |
| `FE-SUPPORT-AC-004` | Frontend phase completion MUST be derived from `tools/frontend_phase_registry.json`, matching frontend phase maps, generated frontend coverage ledgers, and drift checks. |
| `FE-SUPPORT-AC-005` | Bootstrap, route, and error-state selectors MUST be produced from stable selector builders rather than visible text. |
| `FE-SUPPORT-AC-006` | Renderer and editor registry cleanup MUST remove subscriptions, portals, observers, timers, and stale row references. |

### 2.2 Unit

Unit tests must exercise deterministic frontend reducers, adapters, serializers, command translators, registry behavior, and package-boundary rules without requiring a browser server. Unit rows that assert route behavior must use generated or contract-derived types and must not replace public-boundary E2E evidence.

Current repository targets: `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`, `make generated-artifact-policy-check`, `make lint-biome`.

### 2.3 Integration

Integration tests must compose frontend packages at their public boundaries. They may use fixtures and test utilities, but product-conformance evidence for backend/API assumptions must be verified through the public browser-facing boundary in browser or E2E rows.

Current repository targets: `make frontend-unit` for TypeScript integration-style tests where colocated with the frontend unit suite; browser-backed integration should use browser targets.

### 2.4 Browser integration

Browser integration rows cover workbook shell behavior, grid adapter behavior, keyboard choreography, focus, scroll anchors, cell editors, clipboard, drag fill, resize, tree/group rows, and public route/UI transitions in a real browser environment.

Current repository targets: `make browser-e2e-webserver-backed`, `make browser-e2e-stateful`, `make browser-e2e-support`.

### 2.5 E2E

E2E rows must verify browser-visible behavior through the public client/server boundary:

- HTTP+JSON under `/api/v1/`.
- WebSocket under `/ws/v1/`.
- Server-managed session state.
- Stable identifiers: `incident_id`, `view_schema_id`, `record_id`, `base_row_version`, `field_key`, and `client_txn_id`.
- Same-origin evidence preview and download handles.
- Current-role/current-membership authorization behavior observed through public routes.
- Unknown top-level request members rejected where owner route contracts require closed request shape.
- Frontend error rendering based on public error envelopes rather than private server details.

Current repository targets: `make browser-e2e-webserver-backed`, `make browser-e2e-stateful`.

### 2.6 Visual regression

Visual regression rows are implementation-quality evidence. They verify that the frontend continues to meet the intended visual contract for shell density, grid affordances, state presentation, and deterministic workbook composition. They are not claim-bearing benchmark evidence unless Core 05 publication requirements are separately satisfied.

Current repository target: `make browser-e2e-visual`.

### 2.7 Accessibility

Accessibility rows verify keyboard access, visible focus, accessible state communication, ARIA where applicable, icon-only control names, color contrast, and non-color-only state communication. The canonical target is `make browser-e2e-a11y`.

`make browser-e2e-a11y` MUST be a service-backed browser target. It MUST use ordinary authenticated browser sessions, public `/api/v1/` HTTP routes, and public `/ws/v1/` WebSocket boundaries. It MUST NOT disable ordinary security controls. Test-only runtime-control routes MAY be used only under the Core 04 harness-owned boundary.

The target MUST retain exactly one normalized accessibility evidence artifact named `frontend-accessibility-summary.json` with `schema_id="cartulary.frontend_accessibility_summary.v1"`. Raw Playwright output, DOM snapshots, screenshots, traces, and rule-engine output are diagnostic inputs; they MUST NOT replace the normalized summary artifact.

The testing harness NLSpec owns `make browser-e2e-a11y` target mechanics and the normalized `cartulary.frontend_accessibility_summary.v1` schema. This guide maps frontend phase rows to that target; it MUST NOT define an incompatible schema shape. As a derived restatement, the current harness schema includes these arrays:

| Field | Required content |
| --- | --- |
| `phase_rows[]` | Every frontend phase row claimed by the target, including row ID, phase ID, evidence class, claim status, and target mapping. |
| `scenarios[]` | Executed browser scenario titles and row mappings. |
| `keyboard_matrix[]` | Keyboard reachability, focus visibility, entry/exit, focus restoration, and `Esc` priority checks. |
| `state_communication_checks[]` | Accessible-name, ARIA, non-color-only, status, blocked, error, presence, conflict, and evidence-state checks. |
| `contrast_checks[]` | Contrast assertions or retained references for deterministic contrast evaluation. |
| `violations[]` | Blocking accessibility violations with severity, row mapping, and artifact references. |
| `artifact_refs[]` | Retained traces, screenshots, DOM snapshots, logs, and target summaries needed to audit failures. |

### 2.8 Performance and responsiveness

Performance and responsiveness rows may verify implementation support such as sparse patch reference preservation, virtualization behavior, post-scroll settle, and row/cell anchor stability. They must be clearly separated from claim-bearing timed or fixture-sensitive publication. Any published timing claim must satisfy Core 05.

Current repository targets for support evidence: `make browser-e2e-measurement`, `make browser-e2e-support`. Claim publication is not satisfied by these targets alone.

## 3. Current Command Surface

The following targets were found in the repository task surface and may be referenced by rows in this guide:

| Target | Use in this guide |
| --- | --- |
| `make frontend-typecheck` | TypeScript type checking and generated type consumption. |
| `make frontend-unit` | Frontend unit and integration-style package tests. |
| `make frontend-import-boundary-check` | Direct import boundary enforcement, including `/apps/web` direct RDG import prohibition. |
| `make lint-biome` | Frontend formatting/lint support. |
| `make browser-e2e-webserver-backed` | Browser tests that use a webserver-backed public boundary. |
| `make browser-e2e-stateful` | Browser tests that require persistent stateful behavior. |
| `make browser-e2e-support` | Browser support and harness tests. |
| `make browser-e2e-a11y` | Deterministic browser accessibility scenarios and `cartulary.frontend_accessibility_summary.v1` evidence. |
| `make browser-e2e-measurement` | Measurement support evidence, not claim publication by itself. |
| `make browser-e2e-visual` | Playwright visual regression suite. |
| `make generated-artifact-policy-check` | Generated-artifact policy verification. |
| `make generate-drift` | Generated output drift verification. |
| `make phase-ledger-drift` | Existing phase ledger drift verification. |
| `make phase-schedule-drift` | Existing phase schedule drift verification. |
| `make check` | Developer verification gate. |
| `make release-check` | Release verification gate. |

Frontend phase command surface:

- `tools/frontend_phase_registry.json` owns the frontend phase catalog under `phase_namespace="frontend"`.
- `tools/frontend_phase_maps/fe_p*_test_map.json` own frontend row inventories.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p*_coverage_ledger.md` are generated companions and MUST NOT be edited by hand.
- `PHASE=phaseN` remains base-only. Frontend phase selection MUST use `PHASE_NAMESPACE=frontend PHASE=FE-P<N>` and MUST reject ambiguous phase identifiers.

## 4. Frontend Phase Sequence

Each phase must satisfy its own rows, shared harnesses introduced or changed by the phase, and all earlier frontend phases. A phase is not complete when a test name exists; completion requires executed evidence in the intended layer.

### 4.0 Phase FE-P0: Contract/Codegen Baseline And Package Boundaries

| Item | Value |
| --- | --- |
| Scope | Generated protocol facade, view-contract adapters, selector/test-id builders, import-boundary checks, generated drift, and package ownership. |
| Primary owner sections | Core 00 Section 1; Core 01 Sections 3.3.1, 3.3.4, 7.4, 8.5; development guide Section 6; implementation guide Sections 7, 14, 15. |
| Frontend package or app surfaces | `/packages/protocol-ts`, `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/grid-adapter`, `/packages/test-utils`, `/apps/web`; `/packages/ui` is future-reserved and inactive for FE-P0 closure. |
| Introduced user-observable behavior | None by itself. This phase creates the typed and stable-selector foundation used by later user-visible phases. |
| Shared harnesses triggered | Contract-derived view-schema and field-key mapping; stable selector/test-id contracts; generated contract drift; frontend route/API boundary conformance; grid-adapter identity baseline. |
| Completion criteria | Generated protocol types are consumed without hand edits; `/apps/web` has no direct `react-data-grid` imports; RDG stylesheet ownership is enforced through `/packages/grid-adapter`; selector/test-id builders are deterministic; contract drift, generated-artifact policy checks, frontend phase maps, and generated frontend ledgers agree. |
| Out of scope | Workbook shell behavior, data mutation UX, browser visual acceptance, accessibility acceptance, and product claim publication. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P0-01 | Unit | Verify generated protocol exports and frontend contract facades expose stable identifiers without hand-editing generated code. | Core 01 Sections 3.3.1, 7.4; development guide Sections 6.2, 6.6 | `REQ-01-019`, `REQ-01-020`, `REQ-01-022`, `REQ-01-034`, `REQ-01-307`..`REQ-01-311` | `AC-124`, `AC-125`, `AC-127`, `AC-184`, `AC-185`, `AC-231`, `AC-300`..`AC-303`, `AC-366`..`AC-368` | `make frontend-typecheck`; `make frontend-unit` | `product_conformance` |
| FE-U-P0-02 | Unit | Verify view-schema adapters key editable and queryable fields by `field_key`, not labels, indexes, or visible column order. | Core 01 Section 3.3.4; Core 03 Section 14 | `REQ-01-034`, `REQ-01-035`, `REQ-01-036`, `REQ-03-223`..`REQ-03-235` | `AC-124`, `AC-127`, `AC-184`, `AC-185`, `AC-231`, `AC-238`..`AC-243`, `AC-360`, `AC-363`, `AC-364`, `AC-366`..`AC-368`, `AC-372`..`AC-375` | `make frontend-unit` | `product_conformance` |
| FE-U-P0-03 | Unit | Verify stable selector and test-id builders derive identifiers from stable IDs, including row-history `history_item_ref`, registry-backed `view_schema_id` values, and selector-relevant closed vocabularies rather than visible labels. | Core 01 Sections 3.3.4.2 and 7.4; Core 02 Section 18; development guide Section 6.4 | `REQ-01-052`, `REQ-01-053`, `REQ-01-053A`, `REQ-01-054`, `REQ-01-307`..`REQ-01-311`, `REQ-02-222`, `REQ-02-223` | `AC-076`..`AC-084`, `AC-116`..`AC-122`, `AC-137`..`AC-145`, `AC-231`, `AC-252`, `AC-253`, `AC-277`, `AC-284`..`AC-287`, `AC-300`..`AC-303` | `make frontend-unit` | `product_conformance` |
| FE-S-P0-01 | Support | Enforce generated protocol policy and detect generated contract drift. | Core 00 Section 1; development guide Sections 2, 6.6, 7.1 | `N/A: implementation_support; Core owner not claimed` | `FE-SUPPORT-AC-004` | `make generated-artifact-policy-check`; `make generate-drift` | `implementation_support` |
| FE-S-P0-02 | Support | Enforce frontend import boundaries: `/apps/web` must consume `/packages/grid-adapter` and must not import `react-data-grid` directly. | Development guide Sections 6.1, 6.3, 6.8, 6.10; implementation guide Section 14.8 | `N/A: implementation_support; Core owner not claimed` | `FE-SUPPORT-AC-001`, `FE-SUPPORT-AC-002` | `make frontend-import-boundary-check` | `implementation_support` |
| FE-S-P0-03 | Support | Record frontend phase rows in a scheduler-enforced manifest or ledger before claiming phase-enforced completion. | Implementation guide Sections 15 and 16 | `N/A: implementation_support; Core owner not claimed` | `FE-SUPPORT-AC-004`, `FE-GUIDE-AC-018` | `tools/frontend_phase_registry.json`; `tools/frontend_phase_maps/fe_p*_test_map.json`; `docs/testing/frontend_phase_coverage_ledgers/fe_p*_coverage_ledger.md`; `make phase-ledger-drift` | `implementation_support` |

### 4.1 Phase FE-P1: App Shell And Session Bootstrap

| Item | Value |
| --- | --- |
| Scope | `/apps/web` app bootstrap, server-managed session state, login/MFA/admin/incidents entry points, public error envelopes, and initial incident selection through `/api/v1/`. |
| Primary owner sections | Core 01 Sections 2.2, 3.3.1, 3.3.2; Core 04 Sections 2, 3, 5; development guide Section 6.1. |
| Frontend package or app surfaces | `/apps/web`, `/packages/protocol-ts`, `/packages/ui-contracts`, `/packages/test-utils`. |
| Introduced user-observable behavior | App startup routes, authentication/session-visible state, incident entry, browser-visible API errors, and session-revocation handling. |
| Shared harnesses triggered | Frontend route/API boundary conformance; frontend error-state rendering; stable selector/test-id contracts; accessibility names and state communication. |
| Completion criteria | Session and incident bootstrap evidence uses `/api/v1/` public routes and server-managed session state; authorization failures render public envelopes; unknown closed request members are tested where route owners require closure; earlier P0 rows remain green. |
| Out of scope | Workbook shell layout, grid interaction, mutation replay, live collaboration, visual fixture matrix, and claim publication. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P1-01 | Unit | Verify app bootstrap state distinguishes unauthenticated, MFA-required, authenticated, forbidden, revoked, loading, and public error-envelope states. | Core 01 Sections 3.3.1, 3.3.2; Core 04 Sections 2, 3 | `REQ-01-023`..`REQ-01-031`, `REQ-01-522`..`REQ-01-526`, `REQ-04-001`..`REQ-04-017`, `REQ-04-083`, `REQ-04-084` | `AC-123`, `AC-130`, `AC-156`..`AC-163`, `AC-231`, `AC-334`..`AC-342` | `make frontend-unit` | `product_conformance` |
| FE-I-P1-01 | Integration | Verify API client requests stay under `/api/v1/`, preserve server-managed session behavior, and reject private server-detail dependencies in rendered errors. | Core 01 Section 3.3.1; Core 04 Sections 2, 5 | `REQ-01-019`..`REQ-01-022`, `REQ-04-052`, `REQ-04-053`, `REQ-04-110` | `AC-124`..`AC-131`, `AC-135`, `AC-136`, `AC-219`, `AC-220`, `AC-231`, `AC-232`, `AC-234`, `AC-298` | `make frontend-unit` | `product_conformance` |
| FE-E-P1-01 | E2E | Verify login/session bootstrap and incident entry through public routes, including current-role/current-membership authorization effects as browser-observed behavior. | Core 01 Sections 3.3.1, 3.3.2; Core 04 Sections 2, 3 | `REQ-01-023`..`REQ-01-031`, `REQ-04-001`..`REQ-04-017`, `REQ-04-021`..`REQ-04-030`, `REQ-04-085`, `REQ-04-094`, `REQ-04-105`, `REQ-04-106` | `AC-123`, `AC-130`, `AC-149`, `AC-156`..`AC-163`, `AC-178`..`AC-180`, `AC-231`, `AC-254`, `AC-255`, `AC-257`, `AC-260`, `AC-261`, `AC-334`..`AC-342`, `AC-352`, `AC-370`, `AC-371`, `AC-402` | `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-A11Y-P1-01 | Accessibility | Verify session, MFA, incident, forbidden, loading, and error states have keyboard-reachable controls, visible focus, and screen-reader-safe labels. | UI/UX guide Sections 10.5 and 14; `docs/design.md` Accessibility Direction | `N/A: design_direction; Core owner not claimed` | `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |
| FE-S-P1-01 | Support | Verify bootstrap route selectors and error-state selectors use stable test-id builders. | Development guide Sections 6.4 and 7.4; implementation guide Section 14.9A | `N/A: implementation_support; Core owner not claimed` | `FE-SUPPORT-AC-005` | `make frontend-unit`; `make browser-e2e-support` | `implementation_support` |

### 4.2 Phase FE-P2: Workbook Shell And Startup Surface

| Item | Value |
| --- | --- |
| Scope | One continuous workbook shell, built-in tabs, `System views`, current system-view title, saved-view selector, view bar, primary grid surface slot, inspector regions, status strip, presence summary slot, startup fallback, and focus behavior for the `System views` switcher. |
| Primary owner sections | Core 03 Sections 1, 2, 3, 4.1; Core 01 Section 3.3.5.2; UI/UX guide Sections 5, 6, 7, 8, 9, 10; development guide Section 6.1. |
| Frontend package or app surfaces | `/apps/web`, `/packages/ui-contracts`, `/packages/test-utils`, `/packages/ui` where reusable presentational components exist. |
| Introduced user-observable behavior | Workbook navigation shell, startup-surface selection, system-view entry, saved-view placement, status strip slots, and shell-level focus. |
| Shared harnesses triggered | Stable selector/test-id contracts; keyboard and focus traversal; accessibility names and state communication; visual-regression fixtures; frontend route/API boundary conformance. |
| Completion criteria | Shell opens as one continuous workbook surface; built-in tabs are ordered Timeline, Hosts, Identities, Evidence, Notes; required system views are grouped and ordered by the UI/UX guide; startup selection follows explicit launch surface, user preference, incident default, and Timeline fallback; invalid pointers clear and fall through according to owner contracts; earlier phases remain green. |
| Out of scope | RDG adapter internals, inline edit, row mutation replay, evidence handle redemption, and same-field conflict resolver implementation. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P2-01 | Unit | Verify startup-surface resolution order: explicit launch surface, user preference, incident default, Timeline fallback, with invalid pointers clearing and falling through. | Core 03 Section 3; Core 01 Section 3.3.5.2 | `REQ-03-027`..`REQ-03-032`, `REQ-01-138`..`REQ-01-151` | `AC-146`..`AC-153`, `AC-231`, `AC-233` | `make frontend-unit` | `product_conformance` |
| FE-U-P2-02 | Unit | Verify built-in tabs and required system-view groups are registered by stable surface IDs, not visible labels. | Core 03 Sections 2.1, 2.2; Core 01 Sections 7.4, 8.5 | `REQ-03-004`..`REQ-03-011`, `REQ-01-296`..`REQ-01-302`, `REQ-01-499`..`REQ-01-506` | `AC-078`, `AC-085`..`AC-090`, `AC-112`, `AC-116`..`AC-122`, `AC-231`, `AC-277`, `AC-281`..`AC-284`, `AC-300`..`AC-303`, `AC-318`, `AC-410`, `AC-411` | `make frontend-unit` | `product_conformance` |
| FE-B-P2-01 | Browser integration | Verify continuous workbook shell composition: top bar, tab bar, `System views`, view bar, grid slot, inspector tabs, and status strip slots stay in one shell. | Core 03 Sections 1 and 2; UI/UX guide Sections 5, 6, 8, 9 | `REQ-03-001`..`REQ-03-011` | `AC-001`, `AC-002`, `AC-005`, `AC-043`, `AC-078`, `AC-085`..`AC-090`, `AC-112`, `AC-116`, `AC-121`, `AC-122`, `AC-231`, `AC-410`, `AC-411`; `R2-AC-017`..`R2-AC-026` | `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-B-P2-02 | Browser integration | Verify `System views` switcher keyboard entry, roving focus, selection, dismissal, and focus restoration. | Core 03 Sections 2.2, 4.1; UI/UX guide Sections 6, 10 | `REQ-03-005`..`REQ-03-011`, `REQ-03-217`..`REQ-03-222` | `AC-005`, `AC-043`, `AC-078`, `AC-085`..`AC-090`, `AC-121`, `AC-122`, `AC-231`, `AC-354`, `AC-394`..`AC-396`; `R2-AC-027`..`R2-AC-032`, `R2-AC-080`..`R2-AC-086` | `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-E-P2-01 | E2E | Verify saved views appear only under the active surface's view selector and system views open inside the same workbook shell. | Core 03 Sections 2.2 and 3; Core 01 Section 3.3.5.2 | `REQ-03-005`..`REQ-03-026`, `REQ-01-138`..`REQ-01-151` | `AC-078`, `AC-085`..`AC-090`, `AC-121`, `AC-122`, `AC-146`..`AC-153`, `AC-231`, `AC-233`, `AC-410`, `AC-411` | `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-V-P2-01 | Visual regression | Capture default workbook shell with top bar, tabs, `System views`, view bar, primary grid slot, inspector, and status strip using deterministic visual harness settings. | UI/UX guide Sections 5, 6, 8, 9, 13; visual golden guide Sections 2, 3, 5 | `N/A: design_direction; Core owner not claimed` | `R2-AC-017`..`R2-AC-026`, `R2-AC-073`..`R2-AC-079`, `D-AC-001`..`D-AC-006` | `make browser-e2e-visual` | `design_direction` |
| FE-A11Y-P2-01 | Accessibility | Verify shell regions, tabs, switchers, menus, inspector controls, and status strip are keyboard reachable, visibly focused, and named. | UI/UX guide Sections 10.5 and 14; `docs/design.md` Accessibility Direction | `N/A: design_direction; Core owner not claimed` | `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |

### 4.3 Phase FE-P3: Grid Adapter And View-Schema Rendering

| Item | Value |
| --- | --- |
| Scope | `/packages/grid-adapter` RDG containment, stylesheet ownership, `record_id` row identity, `field_key` cells, renderer/editor registry, sort/paste/fill/focus translation, group/tree/presentation rows, sparse patches, lifecycle cleanup, and imperative API containment. |
| Primary owner sections | Core 01 Section 3.3.4; Core 03 Sections 4.1, 14, and 4.13; development guide Sections 6.3, 6.7, 6.8, 6.9, 6.10; implementation guide Sections 14.7, 14.8, 14.9A; R09 research as rationale only. |
| Frontend package or app surfaces | `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/test-utils`, `/apps/web` adapter consumption. |
| Introduced user-observable behavior | Contract-derived grid rendering with stable row/cell identity, read/write affordances, keyboard navigation, sorting commands, clipboard/fill behavior, and presentation-only group/tree rows. |
| Shared harnesses triggered | Contract-derived view-schema and field-key mapping; grid-adapter identity and capability invariants; renderer/editor registry lifecycle cleanup; browser command helpers; keyboard and focus traversal; visual-regression fixtures; accessibility state names and ARIA. |
| Completion criteria | Every rendered data row is keyed by `record_id`; every editable or queryable cell is keyed by `field_key`; vendor coordinates translate to `record_id + field_key`; missing or duplicate `record_id` fails before mutation-capable rendering; group/loading/spacer/presentation rows never emit mutation-capable events; editability derives from explicit editor adapters and contract writeability; `react-data-grid` `editable=true` alone is not sufficient; sort changes compile to view-query `sort[]`; local row reordering is not authoritative sorting; copy/paste/drag fill/editor entry obey contracts; active cell and edit mode restore or intentionally clear by `record_id + field_key`; sparse patches preserve unchanged row references where required; cleanup removes stale subscriptions, portals, observers, timers, and row references; imperative mutation-capable vendor APIs are inaccessible outside the adapter; RDG stylesheet is imported exactly once before workbook-grid rendering; styling uses wrapper classes, CSS variables, documented stable classes, and accessible state attributes. |
| Out of scope | Timeline-specific mutation semantics, pending replay, same-field resolver UI, evidence handle redemption, and saved-view persistence. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P3-01 | Unit | Fail before mutation-capable rendering when data rows have missing or duplicate `record_id`; verify group/loading/spacer/presentation rows cannot emit write events. | Core 01 Section 3.3.4; Core 03 Sections 4.1 and 14; development guide Sections 6.3 and 6.9 | `REQ-01-034`, `REQ-01-037`, `REQ-03-033`..`REQ-03-035`, `REQ-03-217`..`REQ-03-235` | `AC-009`, `AC-013`, `AC-014`, `AC-024`..`AC-026`, `AC-043`, `AC-047`, `AC-124`, `AC-184`, `AC-185`, `AC-231`, `AC-360`, `AC-363`, `AC-364` | `make frontend-unit` | `product_conformance` |
| FE-U-P3-02 | Unit | Translate vendor row/column coordinates to `record_id + field_key` for edit, copy, paste, drag fill, selection, focus, and anchor assertions. | Core 01 Section 3.3.4; Core 03 Sections 4.1 and 4.13; development guide Sections 6.3 and 6.9 | `REQ-01-034`, `REQ-01-035`, `REQ-01-058`, `REQ-03-217`..`REQ-03-222`, `REQ-03-263`..`REQ-03-265` | `AC-003`, `AC-005`, `AC-009`, `AC-040`, `AC-043`, `AC-124`..`AC-127`, `AC-184`, `AC-185`, `AC-231`, `AC-354`, `AC-394`..`AC-396` | `make frontend-unit` | `product_conformance` |
| FE-U-P3-03 | Unit | Verify editability derives from explicit editor adapters and contract writeability; `editable=true` alone never enters edit mode. | Core 01 Sections 3.3.4 and 7.4; Core 03 Section 4.1; development guide Sections 6.3 and 6.7 | `REQ-01-034`, `REQ-01-057`..`REQ-01-070`, `REQ-03-217`..`REQ-03-222` | `AC-003`, `AC-005`, `AC-040`, `AC-043`, `AC-124`..`AC-127`, `AC-181`..`AC-183`, `AC-188`..`AC-190`, `AC-200`..`AC-218`, `AC-221`..`AC-225`, `AC-231`, `AC-299`, `AC-354`, `AC-394`..`AC-396` | `make frontend-unit` | `product_conformance` |
| FE-U-P3-04 | Unit | Verify renderer/editor registry precedence, deterministic fallback, and lifecycle cleanup for subscriptions, portals, observers, timers, and row references. | Development guide Sections 6.7, 6.9, 6.10; implementation guide Section 14.8 | `N/A: implementation_support; Core owner not claimed` | `FE-SUPPORT-AC-006` | `make frontend-unit` | `implementation_support` |
| FE-I-P3-01 | Integration | Verify sparse patches preserve unchanged row object references and intentionally replace changed rows by `record_id`. | Core 01 Section 3.3.4; Core 03 Sections 4.1 and 14; development guide Section 6.9 | `REQ-01-034`, `REQ-01-036`, `REQ-03-033`..`REQ-03-035`, `REQ-03-223`..`REQ-03-235` | `AC-009`, `AC-013`, `AC-047`, `AC-124`, `AC-127`, `AC-184`, `AC-185`, `AC-231`, `AC-238`..`AC-243`, `AC-360`, `AC-361`, `AC-363`, `AC-364`, `AC-366`, `AC-367`, `AC-372`..`AC-374` | `make frontend-unit`; `make browser-e2e-support` | `product_conformance` |
| FE-B-P3-01 | Browser integration | Verify sort, filter, group, resize, paste, drag fill, scroll-to-cell, tree expand/collapse, and anchor assertions through browser command helpers. | Core 03 Sections 4.1, 14, and 4.13; implementation guide Section 14.9A | `REQ-03-217`..`REQ-03-235`, `REQ-03-263`..`REQ-03-265` | `AC-003`, `AC-005`, `AC-013`, `AC-014`, `AC-024`..`AC-026`, `AC-040`, `AC-043`, `AC-044`, `AC-047`, `AC-231`, `AC-354`, `AC-360`, `AC-363`, `AC-364`, `AC-394`..`AC-396` | `make browser-e2e-support`; `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-V-P3-01 | Visual regression | Capture frozen column, resize handle, drag-fill handle, edit cell, tree/group row, grouped result, row gutter presence, and empty successful query fixtures. | UI/UX guide Sections 10 and 13; visual golden guide Sections 2, 3, 5 | `N/A: design_direction; Core owner not claimed` | `R2-AC-033`..`R2-AC-039`, `R2-AC-051`..`R2-AC-054`, `R2-RDG-AC-001`..`R2-RDG-AC-010`, `R2-AC-073`..`R2-AC-079` | `make browser-e2e-visual`; `TODO: add missing fixtures to apps/web/e2e/workbook.visual.spec.ts where absent` | `design_direction` |
| FE-A11Y-P3-01 | Accessibility | Verify grid cells, editors, group/tree rows, active cell, edit mode, disabled/read-only state, and blocked actions are keyboard accessible and announced without color-only signals. | UI/UX guide Sections 10.5 and 14; `docs/design.md` Accessibility Direction | `N/A: design_direction; Core owner not claimed` | `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |
| FE-S-P3-01 | Support | Enforce direct RDG import containment and single stylesheet ownership in `/packages/grid-adapter`. | Development guide Sections 6.3, 6.8, 6.10; R09 Sections 3.4 and 3.8 as rationale | `N/A: implementation_support; Core owner not claimed` | `FE-SUPPORT-AC-001`, `FE-SUPPORT-AC-002` | `make frontend-import-boundary-check`; `make lint-biome` | `implementation_support` |

### 4.4 Phase FE-P4: Timeline Hot Path And Sync Engine

| Item | Value |
| --- | --- |
| Scope | Timeline query, rough row creation, inline edit, paste, save state, pending queue, replay behavior, and public mutation route use. |
| Primary owner sections | Core 01 Sections 3.3.4, 3.3.6; Core 02 Sections 2.3, 5.2; Core 03 Sections 4.1, 4.2, 4.7, 4.9; development guide Section 6.1. |
| Frontend package or app surfaces | `/apps/web` sync engine and controllers, `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/protocol-ts`, `/packages/test-utils`. |
| Introduced user-observable behavior | Timeline query results, low-friction rough row creation, inline edit, paste, queued save state, replay after transient failure, and cell-level validation/errors. |
| Shared harnesses triggered | Sync-engine pending queue and replay; save-state presentation; frontend route/API boundary conformance; grid-adapter identity and capability invariants; frontend error-state rendering; keyboard and focus traversal. |
| Completion criteria | Timeline creates and patches use public route contracts with `client_txn_id`, `record_id`, `base_row_version`, and `field_key`; rough input is preserved; paste and inline edit obey write contracts; save-state strip shows one primary state label and one secondary same-surface message; pending replay is deterministic and idempotent from the browser user's perspective; earlier phases remain green. |
| Out of scope | Entity mention resolution, evidence handles, WebSocket live updates, same-field conflict resolver, saved-view persistence. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P4-01 | Unit | Verify sync-engine pending queue orders creates, patches, paste batches, retries, success, validation failures, and replay by stable mutation identifiers. | Core 01 Section 3.3.6; Core 03 Sections 4.1, 4.7 | `REQ-01-057`..`REQ-01-070`, `REQ-03-217`..`REQ-03-222`, `REQ-03-236`..`REQ-03-241` | `AC-003`, `AC-005`, `AC-040`, `AC-043`, `AC-119`, `AC-120`, `AC-124`..`AC-127`, `AC-181`..`AC-183`, `AC-188`..`AC-193`, `AC-200`..`AC-218`, `AC-221`..`AC-225`, `AC-231`, `AC-299`, `AC-354`, `AC-394`..`AC-396` | `make frontend-unit` | `product_conformance` |
| FE-U-P4-02 | Unit | Verify save-state presentation derives one primary label and one same-surface secondary message from pending, saved, failed, and conflict states. | Core 03 Sections 4.1 and 4.4; UI/UX guide Sections 8 and 10 | `REQ-03-033`..`REQ-03-040`, `REQ-03-077`..`REQ-03-084` | `AC-009`, `AC-013`, `AC-040`, `AC-041`, `AC-047`, `AC-126`, `AC-163`, `AC-231`, `AC-381`; `R2-AC-023`..`R2-AC-026`, `R2-AC-045`..`R2-AC-050` | `make frontend-unit` | `product_conformance` |
| FE-I-P4-01 | Integration | Verify Timeline query response rows render full `view_row_v1` cells and preserve row identity through create, patch, validation error, and refresh. | Core 01 Sections 3.3.4 and 3.3.6; Core 03 Section 4.7 | `REQ-01-034`..`REQ-01-036`, `REQ-01-057`..`REQ-01-070`, `REQ-03-236`..`REQ-03-241` | `AC-119`, `AC-120`, `AC-124`..`AC-127`, `AC-181`..`AC-183`, `AC-188`..`AC-193`, `AC-200`..`AC-218`, `AC-221`..`AC-225`, `AC-231`, `AC-238`..`AC-243`, `AC-299`, `AC-361`, `AC-366`, `AC-367`, `AC-372`..`AC-374` | `make frontend-unit`; `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-E-P4-01 | E2E | Verify rough Timeline row creation, inline edit, paste, pending save, refresh, and replay through `/api/v1/` route contracts. | Core 01 Sections 3.3.4 and 3.3.6; Core 02 Section 2.3; Core 03 Sections 4.1 and 4.7 | `REQ-01-057`..`REQ-01-070`, `REQ-02-024`, `REQ-02-025`, `REQ-03-217`..`REQ-03-222`, `REQ-03-236`..`REQ-03-241` | `AC-003`, `AC-005`, `AC-040`, `AC-043`, `AC-119`, `AC-120`, `AC-124`..`AC-127`, `AC-181`..`AC-183`, `AC-188`..`AC-193`, `AC-200`..`AC-218`, `AC-221`..`AC-225`, `AC-231`, `AC-299`, `AC-354`, `AC-394`..`AC-396`, `AC-406` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | `product_conformance` |
| FE-V-P4-01 | Visual regression | Capture save-state strip, pending replay indication, inline edit cell, and empty successful Timeline query fixtures. | UI/UX guide Sections 8, 10, 13; visual golden guide Sections 2, 3, 5 | `N/A: design_direction; Core owner not claimed` | `R2-AC-023`..`R2-AC-026`, `R2-AC-033`..`R2-AC-039`, `R2-AC-045`..`R2-AC-050`, `R2-AC-073`..`R2-AC-079` | `make browser-e2e-visual`; `TODO: add missing fixtures to apps/web/e2e/workbook.visual.spec.ts where absent` | `design_direction` |
| FE-A11Y-P4-01 | Accessibility | Verify grid navigation, edit entry/exit, paste feedback, validation feedback, save-state communication, and `Esc` priority are keyboard and screen-reader safe. | UI/UX guide Sections 10.1, 10.2, 10.5, 14 | `N/A: design_direction; Core owner not claimed` | `R2-AC-033`..`R2-AC-039`, `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |

### 4.5 Phase FE-P5: Entity And Mention Flows

| Item | Value |
| --- | --- |
| Scope | Hosts, Identities, Notes, entity mention tokens, unresolved token, resolved chip, auto-resolved chip, dismissed mention, manual resolution, and mention provenance visibility. |
| Primary owner sections | Core 01 Sections 3.3.4, 8.5; Core 02 Sections 2.4, 2.5, 3.1, 3.2; Core 03 Sections 4.3, 4.7, 4.11; UI/UX guide Section 10.3. |
| Frontend package or app surfaces | `/apps/web`, `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/protocol-ts`, `/packages/ui-contracts`. |
| Introduced user-observable behavior | Entity sheets and mention chips with unresolved, resolved, auto-resolved, dismissed, and manual-resolution states. |
| Shared harnesses triggered | Contract-derived field-key mapping; renderer/editor registry behavior; frontend route/API boundary conformance; stable selectors; visual-regression fixtures; accessibility state names and ARIA. |
| Completion criteria | Mention and entity flows preserve token/provenance separation; chip state is anchored to stable IDs and field keys; manual resolution and dismissal use public mutation boundaries; auto-resolution is disclosed and undoable where owned; Hosts, Identities, and Notes surfaces obey view-schema contracts; earlier phases remain green. |
| Out of scope | Evidence handle redemption, same-field conflict resolver, WebSocket presence, saved-view persistence, and coordination workbook surfaces beyond entity relationships. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P5-01 | Unit | Verify mention chip view models preserve unresolved, resolved, auto-resolved, dismissed, and manual-resolution state by stable identifiers and field keys. | Core 02 Sections 2.4 and 2.5; Core 03 Section 4.3 | `REQ-02-026`..`REQ-02-044`, `REQ-03-276`..`REQ-03-281` | `AC-006`, `AC-019`..`AC-023`, `AC-028`, `AC-029`, `AC-188`..`AC-190`, `AC-201`, `AC-205`, `AC-221`..`AC-225`, `AC-231`, `AC-388`..`AC-393` | `make frontend-unit` | `product_conformance` |
| FE-I-P5-01 | Integration | Verify Hosts, Identities, and Notes grids render contract-derived columns and preserve mention/entity provenance through edit and refresh. | Core 01 Sections 3.3.4 and 8.5; Core 02 Sections 3.1 and 3.2; Core 03 Sections 4.7 and 4.11 | `REQ-01-034`..`REQ-01-036`, `REQ-01-303`..`REQ-01-311`, `REQ-02-054`..`REQ-02-061`, `REQ-03-242`..`REQ-03-249`, `REQ-03-272` | `AC-006`, `AC-020`, `AC-023`, `AC-045`, `AC-068`..`AC-075`, `AC-097`..`AC-100`, `AC-112`, `AC-116`..`AC-118`, `AC-124`, `AC-127`, `AC-184`..`AC-187`, `AC-196`, `AC-209`, `AC-210`, `AC-231`, `AC-278`, `AC-279`, `AC-300`..`AC-303`, `AC-315`, `AC-318`, `AC-366`..`AC-368`, `AC-396`, `AC-410`, `AC-411` | `make frontend-unit`; `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-E-P5-01 | E2E | Verify manual mention resolution, dismissal, auto-resolution disclosure, and undo through public mutation routes and refreshed rows. | Core 01 Section 3.3.6; Core 02 Sections 2.4, 2.5, 3.1; Core 03 Section 4.3 | `REQ-01-057`..`REQ-01-070`, `REQ-02-026`..`REQ-02-044`, `REQ-02-054`..`REQ-02-061`, `REQ-03-276`..`REQ-03-281` | `AC-006`, `AC-019`..`AC-023`, `AC-028`, `AC-029`, `AC-124`..`AC-127`, `AC-181`..`AC-183`, `AC-186`..`AC-190`, `AC-200`..`AC-218`, `AC-221`..`AC-225`, `AC-231`, `AC-279`, `AC-299`, `AC-388`..`AC-393` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | `product_conformance` |
| FE-V-P5-01 | Visual regression | Capture unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual resolution state fixtures. | UI/UX guide Sections 10.3 and 13; visual golden guide Sections 2, 3, 5 | `N/A: design_direction; Core owner not claimed` | `R2-AC-040`..`R2-AC-044`, `R2-AC-073`..`R2-AC-079` | `make browser-e2e-visual`; `TODO: add missing fixtures to apps/web/e2e/workbook.visual.spec.ts where absent` | `design_direction` |
| FE-A11Y-P5-01 | Accessibility | Verify mention chip states and manual-resolution controls have accessible names, visible focus, and non-color-only distinction. | UI/UX guide Sections 10.3, 10.5, 14 | `N/A: design_direction; Core owner not claimed` | `R2-AC-040`..`R2-AC-044`, `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |

### 4.6 Phase FE-P6: Evidence Lifecycle

| Item | Value |
| --- | --- |
| Scope | Evidence counts, requested/pending/available/blocked/inconsistent states, attach flow, preview handle, download handle, same-origin redemption, current authorization behavior, and blocked/error states. |
| Primary owner sections | Core 01 Sections 3.3.8 and 3.3.9; Core 02 Section 5.1; Core 03 Sections 4.11 and 4.12; Core 04 Sections 3, 5, 6; UI/UX guide Section 12. |
| Frontend package or app surfaces | `/apps/web`, `/packages/protocol-ts`, `/packages/ui-contracts`, `/packages/test-utils`. |
| Introduced user-observable behavior | Evidence affordances, attach status, preview/download handle redemption, blocked preview/download messages, and same-origin evidence access behavior. |
| Shared harnesses triggered | Same-origin evidence preview and download handle behavior; frontend route/API boundary conformance; frontend error-state rendering; visual-regression fixtures; accessibility names and ARIA. |
| Completion criteria | Evidence states are closed and distinguishable; preview/download handles are redeemed through same-origin public routes with current auth re-derived; raw object paths/URLs are not exposed as user-facing access handles; blocked, failed, and inconsistent evidence states are visible without color alone; earlier phases remain green. |
| Out of scope | WebSocket live updates, same-field conflict resolver, saved-view persistence, full coordination surfaces. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P6-01 | Unit | Verify evidence state view models cover requested, pending upload, available, preview blocked, failed, inconsistent, and count-display states. | Core 01 Sections 3.3.8 and 3.3.9; Core 02 Section 5.1; Core 03 Sections 4.11 and 4.12 | `REQ-01-243`..`REQ-01-247`, `REQ-01-458`..`REQ-01-465`, `REQ-02-186`..`REQ-02-201`, `REQ-03-242`..`REQ-03-249`, `REQ-03-272` | `AC-015`, `AC-016`, `AC-053`, `AC-100`, `AC-102`, `AC-103`, `AC-107`..`AC-111`, `AC-128`, `AC-154`, `AC-155`, `AC-231`, `AC-251`..`AC-256`, `AC-278`, `AC-280`, `AC-313`, `AC-321`, `AC-322` | `make frontend-unit` | `product_conformance` |
| FE-I-P6-01 | Integration | Verify attach flow uses generated protocol types, public error envelopes, and stable evidence selectors without raw object URLs or paths. | Core 01 Sections 3.3.8 and 3.3.9; Core 04 Sections 5 and 6 | `REQ-01-243`..`REQ-01-247`, `REQ-01-458`..`REQ-01-465`, `REQ-04-048`, `REQ-04-052`, `REQ-04-053` | `AC-049`..`AC-055`, `AC-102`, `AC-103`, `AC-128`, `AC-130`, `AC-131`, `AC-154`, `AC-155`, `AC-231`, `AC-232`, `AC-234`, `AC-251`..`AC-256`, `AC-321`, `AC-322` | `make frontend-unit`; `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-E-P6-01 | E2E | Verify evidence attach, preview, download, blocked preview, and authorization denial through same-origin public handles. | Core 01 Sections 3.3.8 and 3.3.9; Core 04 Sections 3, 5, 6 | `REQ-01-243`..`REQ-01-247`, `REQ-01-458`..`REQ-01-465`, `REQ-04-021`..`REQ-04-030`, `REQ-04-048`, `REQ-04-052`, `REQ-04-053`, `REQ-04-110` | `AC-049`..`AC-055`, `AC-102`, `AC-103`, `AC-128`, `AC-130`, `AC-131`, `AC-149`, `AC-154`, `AC-155`, `AC-178`..`AC-180`, `AC-231`, `AC-232`, `AC-234`, `AC-251`..`AC-257`, `AC-260`, `AC-261`, `AC-298`, `AC-321`, `AC-322`, `AC-340`..`AC-342`, `AC-352`, `AC-370`, `AC-371`, `AC-402` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | `product_conformance` |
| FE-V-P6-01 | Visual regression | Capture evidence count, affordance, available, requested, pending, blocked, failed, inconsistent, preview, and download-handle state fixtures. | UI/UX guide Sections 12 and 13; visual golden guide Sections 2, 3, 5 | `N/A: design_direction; Core owner not claimed` | `R2-AC-059`..`R2-AC-062`, `R2-AC-073`..`R2-AC-079` | `make browser-e2e-visual`; `TODO: add missing fixtures to apps/web/e2e/workbook.visual.spec.ts where absent` | `design_direction` |
| FE-A11Y-P6-01 | Accessibility | Verify evidence icon buttons, blocked states, error states, preview controls, and download controls have names, focus, contrast, and non-color-only distinctions. | UI/UX guide Sections 12, 14; `docs/design.md` Accessibility Direction | `N/A: design_direction; Core owner not claimed` | `R2-AC-059`..`R2-AC-062`, `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |

### 4.7 Phase FE-P7: Live Collaboration And Conflicts

| Item | Value |
| --- | --- |
| Scope | WebSocket stream handling, event reducer, presence anchoring, live row updates, reset handling, invalidate handling, stale-row requery, same-field conflict display, and resolver. |
| Primary owner sections | Core 01 Section 3.3.10; Core 03 Sections 4.2, 4.4, 4.5, 4.6; Core 04 Sections 2, 5; UI/UX guide Sections 10.4 and 10.5. |
| Frontend package or app surfaces | `/apps/web`, `/packages/grid-adapter`, `/packages/protocol-ts`, `/packages/ui-contracts`, `/packages/test-utils`. |
| Introduced user-observable behavior | Live row refresh, presence hints, reset/invalidate responses, stale-row requery, conflict badges, conflict detail, and resolver actions. |
| Shared harnesses triggered | WebSocket event reducer; same-field conflict anchoring; presence anchoring; save-state presentation; frontend route/API boundary conformance; browser command helpers; visual-regression fixtures; accessibility state names and ARIA. |
| Completion criteria | WebSocket connects only under `/ws/v1/`; reducer handles row update, reset, invalidate, stale requery, and unauthorized/revoked session states; presence anchors by stable row/cell identity; same-field conflicts are cell-local, focus-preserving, and resolved through public mutation contracts; earlier phases remain green. |
| Out of scope | Saved-view persistence, inspector rollback, full coordination surfaces, claim publication. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P7-01 | Unit | Verify WebSocket reducer handles row update, reset, invalidate, stale-row requery request, authorization close, and session revocation states. | Core 01 Section 3.3.10; Core 04 Sections 2 and 5 | `REQ-01-250`..`REQ-01-253`, `REQ-04-001`..`REQ-04-017`, `REQ-04-052`, `REQ-04-053`, `REQ-04-110` | `AC-129`, `AC-131`..`AC-136`, `AC-156`..`AC-163`, `AC-231`, `AC-232`, `AC-233`, `AC-234`, `AC-298`, `AC-334`..`AC-342` | `make frontend-unit` | `product_conformance` |
| FE-U-P7-02 | Unit | Verify same-field conflict anchors, conflict queue, and resolver state use `record_id + field_key + base_row_version` rather than visible indexes. | Core 03 Sections 4.2, 4.4, 4.5, 4.6 | `REQ-03-033`..`REQ-03-084` | `AC-009`, `AC-013`, `AC-037`..`AC-042`, `AC-047`, `AC-126`, `AC-163`, `AC-203`, `AC-204`, `AC-226`..`AC-231`, `AC-381` | `make frontend-unit` | `product_conformance` |
| FE-I-P7-01 | Integration | Verify conflict resolver actions submit public mutations and refresh rows without losing focus or pending queue ordering. | Core 01 Section 3.3.6; Core 03 Sections 4.2 through 4.6 | `REQ-01-057`..`REQ-01-070`, `REQ-03-033`..`REQ-03-084` | `AC-009`, `AC-013`, `AC-037`..`AC-042`, `AC-047`, `AC-124`..`AC-127`, `AC-181`..`AC-183`, `AC-188`..`AC-190`, `AC-200`..`AC-218`, `AC-221`..`AC-231`, `AC-299`, `AC-381` | `make frontend-unit`; `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-E-P7-01 | E2E | Verify multi-client live row update, presence anchoring, reset/invalidate handling, stale-row requery, and same-field conflict resolver through `/ws/v1/` and `/api/v1/`. | Core 01 Sections 3.3.6 and 3.3.10; Core 03 Sections 4.2 through 4.6 | `REQ-01-057`..`REQ-01-070`, `REQ-01-250`..`REQ-01-253`, `REQ-03-033`..`REQ-03-084` | `AC-009`, `AC-013`, `AC-037`..`AC-042`, `AC-047`, `AC-124`..`AC-129`, `AC-131`..`AC-136`, `AC-156`..`AC-163`, `AC-181`..`AC-183`, `AC-188`..`AC-190`, `AC-200`..`AC-218`, `AC-221`..`AC-233`, `AC-299`, `AC-381` | `make browser-e2e-stateful`; `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-V-P7-01 | Visual regression | Capture same-field conflict, row-gutter presence, presence hint, conflict resolver, reset/invalidate notice, and save-state conflict fixtures. | UI/UX guide Sections 8, 10.4, 10.5, 13; visual golden guide Sections 2, 3, 5 | `N/A: design_direction; Core owner not claimed` | `R2-AC-023`..`R2-AC-026`, `R2-AC-045`..`R2-AC-050`, `R2-AC-073`..`R2-AC-079` | `make browser-e2e-visual`; `TODO: add missing fixtures to apps/web/e2e/workbook.visual.spec.ts where absent` | `design_direction` |
| FE-A11Y-P7-01 | Accessibility | Verify conflict state, resolver controls, presence hint, stale-row notice, and save-state conflict communicate state by accessible name/state, not color alone. | UI/UX guide Sections 10.4, 10.5, 14 | `N/A: design_direction; Core owner not claimed` | `R2-AC-045`..`R2-AC-050`, `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |

### 4.8 Phase FE-P8: Saved Views And Layout/Query Controls

| Item | Value |
| --- | --- |
| Scope | Saved views, sorting, filtering, grouping, layout state, active chips, startup/default-surface UI, group/tree rows, and query-control persistence. |
| Primary owner sections | Core 01 Sections 3.3.4 and 3.3.5.2; Core 03 Sections 3 and 14; UI/UX guide Sections 7, 10.5; development guide Sections 6.1 and 6.3. |
| Frontend package or app surfaces | `/apps/web`, `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/protocol-ts`, `/packages/ui-contracts`, `/packages/test-utils`. |
| Introduced user-observable behavior | Saved-view selector, sort/group/filter controls, active filter chips, layout persistence, default/startup surface UI, and non-writable group/tree rows. |
| Shared harnesses triggered | Contract-derived view-schema and field-key mapping; grid-adapter identity and capability invariants; browser command helpers; stable selector/test-id contracts; visual-regression fixtures; accessibility state names and ARIA. |
| Completion criteria | Sort changes compile to view-query `sort[]` keyed by sortable `field_key`; filters and groups use schema fields and public query contracts; group rows are presentation-only; saved views appear under the active surface selector; startup/default choices persist through owner contracts; earlier phases remain green. |
| Out of scope | Timeline create/patch semantics beyond query state, evidence handle redemption, same-field conflict resolver internals, inspector rollback. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P8-01 | Unit | Compile sort, filter, group, layout, and active chips to owner query contracts using schema `field_key` and capability metadata. | Core 01 Section 3.3.4; Core 03 Section 14 | `REQ-01-035`, `REQ-01-038`..`REQ-01-047`, `REQ-03-223`..`REQ-03-235` | `AC-013`, `AC-014`, `AC-024`..`AC-026`, `AC-044`, `AC-047`, `AC-124`, `AC-127`, `AC-184`, `AC-185`, `AC-231`, `AC-238`..`AC-243`, `AC-359`..`AC-361`, `AC-363`, `AC-364`, `AC-372`..`AC-375`, `AC-387` | `make frontend-unit` | `product_conformance` |
| FE-I-P8-01 | Integration | Verify saved-view create/update/select/default UI uses active surface scope and public saved-view/workbook-preference contracts. | Core 01 Section 3.3.5.2; Core 03 Section 3 | `REQ-01-138`..`REQ-01-151`, `REQ-03-012`..`REQ-03-032` | `AC-146`..`AC-153`, `AC-231`, `AC-233`, `AC-360` | `make frontend-unit`; `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-B-P8-01 | Browser integration | Verify browser command helpers for sort, filter, group, active chips, layout persistence, tree/group expand-collapse, and startup/default surface UI. | Core 03 Sections 3 and 14; implementation guide Section 14.9A | `REQ-03-012`..`REQ-03-032`, `REQ-03-223`..`REQ-03-235` | `AC-013`, `AC-014`, `AC-024`..`AC-026`, `AC-044`, `AC-047`, `AC-146`..`AC-153`, `AC-231`, `AC-233`, `AC-360`, `AC-363`, `AC-364` | `make browser-e2e-webserver-backed`; `make browser-e2e-support` | `product_conformance` |
| FE-E-P8-01 | E2E | Verify saved-view persistence, default/startup surface persistence, and query replay through `/api/v1/` after reload. | Core 01 Sections 3.3.4 and 3.3.5.2; Core 03 Sections 3 and 14 | `REQ-01-035`, `REQ-01-038`..`REQ-01-047`, `REQ-01-138`..`REQ-01-151`, `REQ-03-012`..`REQ-03-032`, `REQ-03-223`..`REQ-03-235` | `AC-013`, `AC-014`, `AC-024`..`AC-026`, `AC-044`, `AC-047`, `AC-124`, `AC-127`, `AC-146`..`AC-153`, `AC-184`, `AC-185`, `AC-231`, `AC-233`, `AC-238`..`AC-243`, `AC-359`..`AC-361`, `AC-363`, `AC-364`, `AC-372`..`AC-375`, `AC-387` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | `product_conformance` |
| FE-V-P8-01 | Visual regression | Capture saved-view selector, active chips, grouped result, tree/group row, default/startup state indicator, and empty successful query fixtures. | UI/UX guide Sections 7, 10.5, 13; visual golden guide Sections 2, 3, 5 | `N/A: design_direction; Core owner not claimed` | `R2-AC-027`..`R2-AC-032`, `R2-AC-051`..`R2-AC-054`, `R2-AC-073`..`R2-AC-079`, `R2-RDG-AC-001`..`R2-RDG-AC-010` | `make browser-e2e-visual`; `TODO: add missing fixtures to apps/web/e2e/workbook.visual.spec.ts where absent` | `design_direction` |
| FE-A11Y-P8-01 | Accessibility | Verify sort, filter, group, saved-view menu, active chips, tree/group expand-collapse, and default/startup controls are keyboard reachable and announced. | UI/UX guide Sections 7, 10.5, 14 | `N/A: design_direction; Core owner not claimed` | `R2-AC-027`..`R2-AC-032`, `R2-AC-051`..`R2-AC-054`, `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |

### 4.9 Phase FE-P9: Inspector And Row-Local Actions

| Item | Value |
| --- | --- |
| Scope | Row-local inspector tabs for Details, Relationships, Evidence, History, rollback, and destructive actions. |
| Primary owner sections | Core 01 Sections 3.3.4, 3.3.6, 3.3.7, 3.3.8; Core 02 Sections 5.1, 5.2, 5.4; Core 03 Sections 4.11 and 4.12; UI/UX guide Sections 9, 10, 12. |
| Frontend package or app surfaces | `/apps/web`, `/packages/protocol-ts`, `/packages/ui-contracts`, `/packages/test-utils`, `/packages/ui` where reusable presentational components exist. |
| Introduced user-observable behavior | Inspector tabs, row details, relationship links, evidence panel, history timeline, rollback preview/action, and destructive-action confirmation/errors. |
| Shared harnesses triggered | Frontend route/API boundary conformance; same-origin evidence handle behavior; frontend error-state rendering; stable selector/test-id contracts; keyboard and focus traversal; accessibility names and ARIA. |
| Completion criteria | Inspector is row-local and anchored by `record_id`; relationships and evidence use stable IDs; history and rollback use public route contracts; destructive actions re-check current authorization and render public envelopes; focus returns to the originating row/cell when appropriate; earlier phases remain green. |
| Out of scope | Coordination workbook surfaces beyond relationships, WebSocket live updates, visual claim publication. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P9-01 | Unit | Verify inspector selection, tab state, details, relationships, evidence, and history anchors are `record_id` based and survive row refresh. | Core 01 Sections 3.3.4, 3.3.7, 3.3.8; Core 03 Sections 4.11 and 4.12 | `REQ-01-034`..`REQ-01-036`, `REQ-01-048`..`REQ-01-056`, `REQ-01-243`..`REQ-01-247`, `REQ-03-242`..`REQ-03-249`, `REQ-03-272` | `AC-006`, `AC-015`, `AC-020`, `AC-023`, `AC-045`, `AC-072`..`AC-075`, `AC-097`..`AC-100`, `AC-102`, `AC-103`, `AC-107`..`AC-111`, `AC-124`, `AC-127`, `AC-154`, `AC-155`, `AC-184`, `AC-185`, `AC-209`, `AC-210`, `AC-231`, `AC-278`..`AC-280`, `AC-313`, `AC-315`, `AC-318`, `AC-321`, `AC-366`, `AC-367`, `AC-372`..`AC-374` | `make frontend-unit` | `product_conformance` |
| FE-I-P9-01 | Integration | Verify history and rollback preview/action use public route contracts, preserve retained history, and render public error envelopes. | Core 01 Section 3.3.7; Core 02 Section 5.2; Core 03 Section 4.12 | `REQ-01-048`..`REQ-01-056`, `REQ-01-089`..`REQ-01-108`, `REQ-02-205`..`REQ-02-218`, `REQ-02-238`..`REQ-02-242` | `AC-107`..`AC-111`, `AC-124`..`AC-128`, `AC-154`, `AC-155`, `AC-215`..`AC-218`, `AC-231`, `AC-383`..`AC-386`, `AC-412` | `make frontend-unit`; `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-E-P9-01 | E2E | Verify inspector Details, Relationships, Evidence, History, rollback, and destructive-action authorization through public browser routes. | Core 01 Sections 3.3.6, 3.3.7, 3.3.8; Core 02 Sections 5.1 and 5.2; Core 04 Sections 3 and 5 | `REQ-01-057`..`REQ-01-070`, `REQ-01-089`..`REQ-01-108`, `REQ-01-243`..`REQ-01-247`, `REQ-02-186`..`REQ-02-201`, `REQ-02-205`..`REQ-02-218`, `REQ-04-021`..`REQ-04-030`, `REQ-04-052`, `REQ-04-053` | `AC-049`..`AC-055`, `AC-102`, `AC-103`, `AC-107`..`AC-111`, `AC-124`..`AC-128`, `AC-149`, `AC-154`, `AC-155`, `AC-178`..`AC-180`, `AC-181`..`AC-183`, `AC-188`..`AC-190`, `AC-200`..`AC-218`, `AC-221`..`AC-225`, `AC-231`, `AC-232`, `AC-234`, `AC-251`..`AC-257`, `AC-260`, `AC-261`, `AC-278`, `AC-280`, `AC-299`, `AC-313`, `AC-321`, `AC-340`..`AC-342`, `AC-352`, `AC-370`, `AC-371`, `AC-402` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | `product_conformance` |
| FE-V-P9-01 | Visual regression | Capture inspector Details, Relationships, Evidence, History, rollback preview, destructive confirmation, and public error fixtures. | UI/UX guide Sections 9, 12, 13; visual golden guide Sections 2, 3, 5 | `N/A: design_direction; Core owner not claimed` | `R2-AC-023`..`R2-AC-026`, `R2-AC-059`..`R2-AC-062`, `R2-AC-073`..`R2-AC-079` | `make browser-e2e-visual`; `TODO: add missing fixtures to apps/web/e2e/workbook.visual.spec.ts where absent` | `design_direction` |
| FE-A11Y-P9-01 | Accessibility | Verify inspector tabs, relationship links, evidence controls, history controls, rollback, destructive actions, and errors are keyboard reachable and announced. | UI/UX guide Sections 9, 12, 14 | `N/A: design_direction; Core owner not claimed` | `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |

### 4.10 Phase FE-P10: Remaining Workbook Surfaces And Keyboard Completion

| Item | Value |
| --- | --- |
| Scope | Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, Lesson, full keyboard contract, clipboard behavior, drag fill, column resize, frozen columns, virtual scroll, tree/group rows, focus restoration, and `Esc` behavior. |
| Primary owner sections | Core 01 Sections 8.5 and Appendix F; Core 02 Sections 5.3 and 5.4; Core 03 Sections 2.2, 4.1, 14, 4.9, 4.10, and 4.13; UI/UX guide Sections 6, 10, 11, 13, 14. |
| Frontend package or app surfaces | `/apps/web`, `/packages/grid-adapter`, `/packages/view-contracts`, `/packages/protocol-ts`, `/packages/ui-contracts`, `/packages/test-utils`. |
| Introduced user-observable behavior | Remaining coordination and review surfaces plus complete keyboard, clipboard, and grid manipulation contract. |
| Shared harnesses triggered | Contract-derived view-schema and field-key mapping; grid-adapter identity/capability invariants; browser command helpers; keyboard and focus traversal; stable selectors; visual fixtures; accessibility names and ARIA. |
| Completion criteria | Required system-view surfaces open inside the same workbook shell; coordination and review surfaces use contract-backed fields; all keyboard/clipboard/resize/fill/frozen/scroll/tree/group interactions preserve `record_id + field_key` identity; `Esc` priority is deterministic; earlier phases remain green. |
| Out of scope | New behavior for route semantics, security policy, claim publication, or package ownership changes. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P10-01 | Unit | Verify coordination and review system-view registrations, field mappings, and closed vocabulary options use stable IDs and contract metadata. | Core 01 Section 8.5; Core 02 Sections 5.3 and 5.4; Core 03 Sections 2.2, 4.9, 4.10, 4.13 | `REQ-01-296`..`REQ-01-302`, `REQ-01-499`..`REQ-01-509`, `REQ-02-222`..`REQ-02-232`, `REQ-03-005`..`REQ-03-011`, `REQ-03-250`..`REQ-03-260`, `REQ-03-265`..`REQ-03-274` | `AC-076`..`AC-090`, `AC-116`..`AC-122`, `AC-137`..`AC-145`, `AC-231`, `AC-252`, `AC-253`, `AC-277`..`AC-287`, `AC-300`..`AC-303`, `AC-315`, `AC-318`, `AC-319`, `AC-410`, `AC-411` | `make frontend-unit` | `product_conformance` |
| FE-B-P10-01 | Browser integration | Verify Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson open inside the same workbook shell and retain view controls. | Core 03 Sections 2.2, 4.9, 4.10; UI/UX guide Sections 6 and 11 | `REQ-03-005`..`REQ-03-011`, `REQ-03-250`..`REQ-03-260`, `REQ-03-273` | `AC-078`, `AC-080`..`AC-090`, `AC-121`, `AC-122`, `AC-137`..`AC-145`, `AC-231`, `AC-277`..`AC-287`, `AC-315`, `AC-319`, `AC-410`, `AC-411`; `R2-AC-055`..`R2-AC-058` | `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-B-P10-02 | Browser integration | Verify full keyboard/clipboard contract: copy, paste, drag fill, column resize, frozen columns, virtual scroll, tree/group rows, focus restoration, and `Esc` priority ladder. | Core 03 Sections 4.1, 14, and 4.13; UI/UX guide Sections 10 and 14 | `REQ-03-217`..`REQ-03-235`, `REQ-03-263`..`REQ-03-265` | `AC-003`, `AC-005`, `AC-013`, `AC-014`, `AC-024`..`AC-026`, `AC-040`, `AC-043`, `AC-044`, `AC-047`, `AC-231`, `AC-354`, `AC-360`, `AC-363`, `AC-364`, `AC-394`..`AC-396`; `R2-AC-033`..`R2-AC-039`, `R2-AC-080`..`R2-AC-086` | `make browser-e2e-support`; `make browser-e2e-webserver-backed` | `product_conformance` |
| FE-E-P10-01 | E2E | Verify coordination rows can be queried and edited through public view/row mutation contracts with current-role authorization. | Core 01 Sections 3.3.4, 3.3.6, 8.5; Core 03 Sections 4.9 and 4.10; Core 04 Section 3 | `REQ-01-034`..`REQ-01-036`, `REQ-01-057`..`REQ-01-070`, `REQ-01-296`..`REQ-01-302`, `REQ-01-499`..`REQ-01-506`, `REQ-03-250`..`REQ-03-260`, `REQ-03-273`, `REQ-04-021`..`REQ-04-030` | `AC-078`, `AC-080`..`AC-090`, `AC-121`, `AC-122`, `AC-124`..`AC-127`, `AC-137`..`AC-145`, `AC-149`, `AC-178`..`AC-180`, `AC-181`..`AC-183`, `AC-188`..`AC-190`, `AC-200`..`AC-218`, `AC-221`..`AC-225`, `AC-231`, `AC-238`..`AC-243`, `AC-277`..`AC-284`, `AC-299`, `AC-300`..`AC-303`, `AC-315`, `AC-318`, `AC-319`, `AC-340`..`AC-342`, `AC-352`, `AC-370`, `AC-371`, `AC-402` | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | `product_conformance` |
| FE-V-P10-01 | Visual regression | Capture Task Requests or Decisions, Parties link state, Communications Log, Handoff, Status Review, Lesson, keyboard focus, frozen column, resize handle, and drag-fill fixtures. | UI/UX guide Sections 11, 13, 14; visual golden guide Sections 2, 3, 5 | `N/A: design_direction; Core owner not claimed` | `R2-AC-055`..`R2-AC-058`, `R2-AC-073`..`R2-AC-086`, `R2-RDG-AC-001`..`R2-RDG-AC-010` | `make browser-e2e-visual`; `TODO: add missing fixtures to apps/web/e2e/workbook.visual.spec.ts where absent` | `design_direction` |
| FE-A11Y-P10-01 | Accessibility | Verify coordination surfaces and full keyboard/clipboard controls meet keyboard reachability, focus visibility, accessible-name, ARIA, and non-color-only state expectations. | UI/UX guide Sections 10, 11, 14; `docs/design.md` Accessibility Direction | `N/A: design_direction; Core owner not claimed` | `R2-AC-033`..`R2-AC-039`, `R2-AC-055`..`R2-AC-058`, `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |

### 4.11 Phase FE-P11: Visual, Accessibility, And Readiness Gates

| Item | Value |
| --- | --- |
| Scope | Visual fixture matrix, accessibility coverage, density/theme/typography/status patterns, command-surface integration, measurement support separation, Core 05 publication boundary, and readiness gate composition. |
| Primary owner sections | Core 00 Section 1; Core 05 Sections 1 through 4; UI/UX guide Sections 13 and 14; `docs/design.md`; visual golden guide Sections 2 through 6; implementation guide Sections 15 and 16; development guide Section 7. |
| Frontend package or app surfaces | `/apps/web/e2e`, `/packages/test-utils`, `/packages/ui-contracts`, root Make surface and task manifests. |
| Introduced user-observable behavior | No new product behavior. This phase completes verification discipline for the frontend MVP. |
| Shared harnesses triggered | Visual-regression fixtures; accessibility names and ARIA; browser command helpers; generated contract drift; frontend route/API boundary conformance; claim-publication boundary checks. |
| Completion criteria | Visual fixtures are deterministic and refreshed only for allowed reasons; accessibility coverage applies to all visible phases; readiness gates compose exact declared targets and explicitly blocked missing fixtures; performance/measurement evidence is not presented as Core 05 publication evidence unless Core 05 is satisfied; earlier phases remain green. |
| Out of scope | Creating new product behavior, changing owner specs, or publishing benchmark/fixture-sensitive claims without Core 05 evidence. |

| ID | Layer | Test | Exact owner sections | Exact REQs | Exact ACs | Repository target or TODO | Evidence class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FE-V-P11-01 | Visual regression | Run the owned-stack Playwright visual suite with deterministic seed data, viewport, zoom, fixture ordering, dynamic masks, scroll anchors, focus/editor state, inspector state, and post-scroll settle behavior. | Visual golden guide Sections 2, 3, 4, 5; UI/UX guide Section 13 | `N/A: implementation_support; Core owner not claimed` | `R2-AC-073`..`R2-AC-079`, `D-AC-001`..`D-AC-006`, `D-AC-010`, `D-AC-011` | `make browser-e2e-visual` | `implementation_support` |
| FE-V-P11-02 | Visual regression | Ensure the visual fixture matrix includes default grid shell, unresolved/resolved entity state, same-field conflict, row-gutter presence, evidence affordance, grouped result, Task Requests or Decisions, save-state strip, frozen column, resize handle, drag-fill handle, edit cell, tree/group row, exposed theme states, and empty successful query. | Visual golden guide Sections 2, 3, 5; UI/UX guide Sections 10 through 13 | `N/A: implementation_support; Core owner not claimed` | `R2-AC-040`..`R2-AC-062`, `R2-AC-073`..`R2-AC-079`, `R2-RDG-AC-001`..`R2-RDG-AC-010` | `make browser-e2e-visual`; `TODO: add missing fixtures to apps/web/e2e/workbook.visual.spec.ts where absent` | `implementation_support` |
| FE-A11Y-P11-01 | Accessibility | Verify global accessibility matrix for keyboard access, visible focus, `System views`, grid navigation/edit entry/exit, `Esc`, ARIA states, icon-only labels, contrast, and non-color-only empty/loading/error/blocked states. | UI/UX guide Section 14; `docs/design.md` Accessibility Direction | `N/A: design_direction; Core owner not claimed` | `R2-AC-080`..`R2-AC-086`, `D-AC-009`, `D-AC-012` | `make browser-e2e-a11y` | `design_direction` |
| FE-S-P11-01 | Support | Verify frontend type check, unit tests, import-boundary check, Biome lint, generated drift, generated-artifact policy, phase ledger drift, and phase schedule drift are composed into readiness expectations. | Development guide Section 7; implementation guide Sections 15 and 16 | `N/A: implementation_support; Core owner not claimed` | `FE-SUPPORT-AC-004`, `FE-GUIDE-AC-018`, `FE-GUIDE-AC-022` | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make lint-biome`; `make generated-artifact-policy-check`; `make generate-drift`; `make phase-ledger-drift`; `make phase-schedule-drift`; `make check` | `implementation_support` |
| FE-S-P11-02 | Support | Verify frontend release readiness composes build, check, visual readiness, generated drift, frontend phase drift, and accessibility readiness without representing blocked fixtures or planned phases as complete. | Development guide Section 7; implementation guide Sections 15 and 16 | `N/A: implementation_support; Core owner not claimed` | `FE-GUIDE-AC-018`, `FE-GUIDE-AC-021`, `FE-GUIDE-AC-022` | `make check`; `make release-check`; `make browser-e2e-a11y` | `implementation_support` |
| FE-S-P11-03 | Support | Verify visual, responsiveness, and measurement evidence remains implementation-quality evidence unless Core 05 claim-publication requirements are separately satisfied. | Core 05 Sections 1 through 4; visual golden guide Section 1 | `REQ-05-001`..`REQ-05-013` | `PC-001`..`PC-006` | `make browser-e2e-measurement`; `make browser-e2e-visual`; conditional claim-publication review object from §7.6 | `claim_publication_boundary` |

## 5. Shared Frontend Harnesses

Shared harnesses apply across every phase that introduces or changes the covered surface. A phase must not be marked complete until each triggered harness has run in the intended layer or is explicitly marked TODO and excluded from authoritative completion.

| Harness ID | Harness | Applies to phases | Owner basis | Repository target or TODO | Completion rule |
| --- | --- | --- | --- | --- | --- |
| FE-H-01 | Contract-derived view-schema and field-key mapping | P0, P2 through P10 | Core 01 Section 3.3.4; Core 03 Sections 4.1 and 14 | `make frontend-unit`; `make frontend-typecheck` | Every queryable/editable field maps by `field_key`; labels and visible indexes are never authoritative. |
| FE-H-02 | Grid-adapter identity and capability invariants | P0, P3 through P10 | Core 01 Section 3.3.4; Core 03 Sections 4.1, 14, and 4.13; development guide Section 6.3 | `make frontend-unit`; `make browser-e2e-support` | Row identity is `record_id`; cell identity is `field_key`; presentation rows cannot write; writeability requires explicit editor adapter and contract capability. |
| FE-H-03 | Renderer/editor registry behavior and lifecycle cleanup | P3 through P10 | Development guide Sections 6.7, 6.9, 6.10; R09 rationale | `make frontend-unit`; `make browser-e2e-support` | Registry precedence is deterministic; cleanup removes stale subscriptions, portals, observers, timers, and row references. |
| FE-H-04 | Sync-engine pending queue and replay behavior | P4 through P10 | Core 01 Section 3.3.6; Core 03 Sections 4.1, 4.4, 4.7 | `make frontend-unit`; `make browser-e2e-stateful` | Pending creates, patches, paste batches, retries, failures, and replay are ordered by stable mutation identifiers and public route results. |
| FE-H-05 | WebSocket event reducer behavior | P7 through P10 | Core 01 Section 3.3.10; Core 04 Sections 2 and 5 | `make frontend-unit`; `make browser-e2e-stateful` | Events under `/ws/v1/` reduce to row updates, reset, invalidate, stale requery, auth close, and session revocation states without private protocol assumptions. |
| FE-H-06 | Same-field conflict anchoring | P4, P7 through P10 | Core 03 Sections 4.2 through 4.6 | `make frontend-unit`; `make browser-e2e-stateful` | Conflict state is anchored by `record_id + field_key + base_row_version`; resolver focus returns or intentionally clears by that anchor. |
| FE-H-07 | Presence anchoring | P2, P7 through P10 | Core 01 Section 3.3.10; Core 03 Sections 4.2 and 4.4; UI/UX guide Section 10.4 | `make browser-e2e-stateful`; `make browser-e2e-visual` | Presence UI anchors to stable row/cell/surface IDs and survives row refresh, sort, group, and scroll. |
| FE-H-08 | Save-state presentation | P2, P4, P7 through P10 | Core 03 Sections 4.1, 4.4; UI/UX guide Sections 8 and 10.4 | `make frontend-unit`; `make browser-e2e-webserver-backed`; `make browser-e2e-visual` | Status strip shows one primary save-state label and one secondary same-surface message. |
| FE-H-09 | Browser command helpers | P3 through P10 | Core 03 Sections 4.1, 14, and 4.13; implementation guide Section 14.9A | `make browser-e2e-support`; `make browser-e2e-webserver-backed` | Helpers cover sort, filter, group, resize, paste, drag fill, scroll-to-cell, tree expand/collapse, and anchor assertions. |
| FE-H-10 | Visual-regression fixtures | P2 through P11 | UI/UX guide Section 13; `docs/design.md`; visual golden guide | `make browser-e2e-visual` | Fixtures use deterministic seed data, viewport, browser zoom, fixture order, dynamic masks, scroll anchors, focus/editor state, inspector state, and post-scroll settle behavior. |
| FE-H-11 | Keyboard and focus traversal | P1 through P11 | Core 03 Sections 4.1 and 4.13; UI/UX guide Section 14 | `make browser-e2e-support`; `make browser-e2e-a11y` | Keyboard access, visible focus, focus restoration, edit entry/exit, menus, switchers, inspector controls, conflict resolver, and `Esc` priority are covered. |
| FE-H-12 | Accessibility names, ARIA, and state communication | P1 through P11 | UI/UX guide Section 14; `docs/design.md` Accessibility Direction | `make browser-e2e-a11y` | Active tab/surface/menu/tree/group/conflict/save/presence/evidence states are named and communicated without color alone. |
| FE-H-13 | Stable selector and test-id contracts | P0 through P11 | Development guide Sections 6.4 and 7.4; implementation guide Section 14.9A | `make frontend-unit`; `make browser-e2e-support` | Test IDs are generated from stable IDs, row-history `history_item_ref`, and closed vocabularies; visible labels are not stable selectors. |
| FE-H-14 | Frontend route/API boundary conformance | P1 through P10 | Core 01 Section 3.3.1; Core 04 Sections 2, 3, 5 | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | E2E evidence uses `/api/v1/`, `/ws/v1/`, server sessions, stable IDs, current auth, closed request shapes where owned, and public error envelopes. |
| FE-H-15 | Same-origin evidence preview and download handles | P6, P9 | Core 01 Section 3.3.9; Core 04 Sections 5 and 6 | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Evidence preview/download redeems same-origin handles; raw storage paths are not UI handles; current auth is re-derived. |
| FE-H-16 | Frontend error-state rendering | P1 through P10 | Core 01 Section 3.3.1; Core 04 Sections 2, 3, 5 | `make frontend-unit`; `make browser-e2e-webserver-backed` | Errors render public envelopes and user-visible states without relying on private server details. |
| FE-H-17 | Generated contract drift | P0 through P11 | Core 00 Section 1; development guide Sections 2, 6.6, 7.1 | `make generated-artifact-policy-check`; `make generate-drift` | Generated TypeScript protocol surfaces are not hand-edited and drift checks pass before completion claims. |

### 5.1 Visual fixture matrix

The visual harness must maintain one owned-stack Playwright screenshot suite unless a newer adopted harness source changes that boundary. The current suite location is `apps/web/e2e/workbook.visual.spec.ts`; the current target is `make browser-e2e-visual`.

| Fixture ID | Fixture | Phase owners | Current target |
| --- | --- | --- | --- |
| FE-VFIX-01 | Default grid shell | P2 | `make browser-e2e-visual` |
| FE-VFIX-02 | Unresolved and resolved entity state | P5 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-03 | Same-field conflict | P7 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-04 | Row-gutter presence | P7 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-05 | Evidence affordance | P6 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-06 | Grouped result | P8 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-07 | Task Requests or Decisions | P10 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-08 | Save-state strip | P4, P7 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-09 | Frozen column | P3, P10 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-10 | Resize handle | P3, P10 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-11 | Drag-fill handle | P3, P10 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-12 | Edit cell | P3, P4, P10 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-13 | Tree/group row | P3, P8, P10 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-14 | Exposed theme states | P11 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |
| FE-VFIX-15 | Empty successful query | P3, P4, P8 | `make browser-e2e-visual`; `TODO: add missing fixture where absent` |

Visual fixture refresh is allowed only for intentional UI contract changes, visual harness changes, dependency/browser/platform pin changes, or stale goldens relative to validated behavior. Refreshes must preserve deterministic seed data, deterministic viewport, deterministic browser zoom, deterministic fixture ordering, masked dynamic regions for timestamps, cursors, avatars, generated IDs, and clock-derived labels, explicit capture state for scroll top/left, named scroll anchors, focus/editor state, inspector state, dynamic masks, and post-scroll settle behavior.

Visual-regression results are implementation-quality evidence and are not claim-bearing benchmark evidence unless Core 05 publication requirements are separately satisfied.

## 6. Phase Completion Checklist

Frontend phase completion is binary. A phase is complete only when all applicable rules below are satisfied:

1. Every phase row executes in its intended layer: unit rows in frontend unit/typecheck targets, integration rows in frontend integration-style or browser-backed targets, browser rows in Playwright/browser targets, visual rows in `make browser-e2e-visual`, accessibility rows in `make browser-e2e-a11y`, and support rows in their named support targets.
2. Every shared harness triggered by introduced or changed surfaces has executed or is explicitly marked `claim_status="blocked"` and excluded from authoritative completion.
3. All earlier frontend phases remain green.
4. Backend/API assumptions are verified through public browser-facing boundaries, not frontend-only mocks, whenever the row claims product conformance for route, auth, session, WebSocket, evidence, mutation, saved-view, or history behavior.
5. Public boundary evidence uses `/api/v1/`, `/ws/v1/`, server-managed session state, stable identifiers, same-origin evidence handles, current-role/current-membership authorization, closed request-shape rejection where owned, and public error envelopes.
6. Product conformance, design-direction acceptance, implementation-support evidence, and claim-publication boundaries remain separate in row metadata and summary reporting.
7. Rows with `TODO: owner lookup required` or `claim_status="blocked"` are not counted as authoritative completion criteria.
8. Missing repository artifacts are represented as blocked rows and are not presented as existing targets or manifests.
9. Generated protocol outputs under `/packages/protocol-ts/src/generated/**` remain unedited by hand.
10. `/apps/web` does not import `react-data-grid` directly and consumes `/packages/grid-adapter`.
11. Visual fixtures are deterministic and refreshed only under the visual golden maintenance rules.
12. Accessibility coverage exists for every phase that introduces user-visible UI and is represented in `cartulary.frontend_accessibility_summary.v1`.
13. Core 05 claim-publication checks are required only when publishing claim-bearing timed, benchmark, or fixture-sensitive evidence.

## 7. Coverage Ledger

### 7.1 Row-to-owner mapping

Every frontend row must map to exact owner sections. Product-conformance rows MUST map to exact `REQ-*` identifiers and exact Core `AC-*` identifiers. A row may cite UI/UX guide, `docs/design.md`, visual-golden, development-guide, or guide-local acceptance criteria for design direction or implementation support, but those evidence classes MUST NOT be `product_conformance`.

Rows without exact owner evidence MUST use `TODO: owner lookup required` and evidence class `TODO_owner_lookup`. Rows that intentionally have non-Core ownership MUST declare that Core ownership is not claimed and MUST cite support/design acceptance criteria.

### 7.2 Row-to-repository target mapping

Rows must map to exact existing repository targets or to explicit TODOs. Current exact target names available for frontend guide rows are listed in Section 3. Do not invent target names, manifest names, package names, or scripts.

The existing repository has `tools/phase*_test_map.json`, `docs/testing/phase*_coverage_ledger.md`, `tools/browser_e2e_batch_manifest.json`, `tools/execution_topology_manifest.json`, and `tools/test_accounting_classification.json`. Frontend phase rows use a separate namespace and MUST NOT be appended to base `phaseN` maps.

`tools/frontend_phase_registry.json` MUST declare `schema_id="cartulary.frontend_phase_registry.v1"`, `phase_namespace="frontend"`, `guide_path`, and `phases[]`. Valid phase IDs are exactly `FE-P0` through `FE-P11`. Valid status values are `planned`, `active`, and `retired`. Each phase item MUST include `phase_id`, `status`, `manifest_path`, `ledger_path`, `owner_refs[]`, and `depends_on[]`.

Initial frontend phase registry entries MUST use `status="planned"`. A frontend phase MAY move to `active` only when its frontend map, generated ledger, public targets, target schedule metadata, row evidence, and evidence-class owner metadata are promoted in the same change. An `active` frontend phase MUST NOT contain `claim_status="blocked"` rows. A `planned` frontend phase MAY be explainable by task guidance but MUST NOT be executable by `phase-slice` or `service-backed-slice`.

Each `tools/frontend_phase_maps/fe_p*_test_map.json` file MUST declare `schema_id="cartulary.frontend_phase_test_map.v1"`, `phase_namespace="frontend"`, matching `phase_id`, and `rows[]`. Each row MUST include `id`, `layer`, `evidence_class`, `owner_refs[]`, `core_req_ids[]`, `core_ac_ids[]`, `support_or_design_ac_ids[]`, `targets[]`, `scenario_titles[]`, `claim_status`, `claim`, and `out_of_scope`.

Generated frontend ledgers under `docs/testing/frontend_phase_coverage_ledgers/` MUST render registry metadata, row metadata, target mapping, evidence class, owner refs, Core IDs, support/design AC IDs, blocked status, scenario titles, and source-map claims. The maps are the source of truth; ledgers are rendered companions.

### 7.3 Browser, visual, accessibility, and support classification

Browser rows classify as `product_conformance` only when they verify Core-owned behavior through the public browser-facing boundary. Browser rows that only prove helper choreography, selector stability, timing support, or visual harness mechanics classify as `implementation_support`.

Visual rows classify as `design_direction` when they verify UI/UX or `docs/design.md` direction, and as `implementation_support` when they verify the visual harness itself. Visual rows do not classify as product conformance unless a Core owner or adopted NLSpec owns the exact visual behavior.

Accessibility rows classify as `design_direction` unless a future Core owner or adopted NLSpec creates product-conformance accessibility requirements. Accessibility rows still gate implementation readiness because the UI/UX guide requires reviewer discipline for keyboard access, visible focus, accessible names, contrast, and non-color-only state communication.

Support rows classify as `implementation_support` unless they check Core 05 claim-publication separation, in which case they classify as `claim_publication_boundary`.

### 7.4 Unowned regression tests

An unowned regression test must be promoted by adding row metadata to the relevant existing manifest or future frontend-local manifest, mapping it to exact owner sections, `REQ-*`, `AC-*`, target, and evidence class. After promotion, remove the unowned classification from `tools/test_accounting_classification.json`.

If no exact owner exists, keep the row as `TODO: owner lookup required` and do not count it as authoritative completion.

### 7.5 Drift checks

Frontend guide/test divergence must be checked with existing drift targets where present:

- `make generated-artifact-policy-check`
- `make generate-drift`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `make frontend-import-boundary-check`
- `make browser-e2e-duration-baseline-drift RESULTS_DIR=<dir>` when retained browser timing artifacts are used for duration-maintenance work

Frontend drift coverage MUST fail when a frontend map is missing, a generated frontend ledger is stale, a row has invalid evidence-class metadata, a row targets an unknown public command, a browser-backed row lacks `scenario_titles[]`, or an active frontend phase contains blocked rows.

### 7.6 Conditional claim-publication review

Frontend readiness evidence MUST NOT activate Core 05 by default. Release-support metadata MAY declare `claim_publication_intent` with exactly one of these values:

| Value | Meaning | Default |
| --- | --- | --- |
| `none` | No timed, benchmark, fixture-sensitive, visual, accessibility, or measurement evidence is being published as a claim. | yes |
| `informative_engineering_measurement` | Measurement evidence is retained for engineering analysis only and is not claim-bearing publication. | no |
| `claim_bearing_publication` | Evidence is intended to support a public timed, benchmark, fixture-sensitive, or measurement claim. | no |

When `claim_publication_intent="none"`, no Core 05 gate MUST run. When `claim_publication_intent="informative_engineering_measurement"`, measurement artifacts MAY be retained but MUST NOT be treated as claim-bearing evidence. When `claim_publication_intent="claim_bearing_publication"`, release validation MUST require:

| Field | Requirement |
| --- | --- |
| `claim_text` | Exact public claim text. |
| `criterion_ids[]` | Criteria the claim purports to satisfy. |
| `measurement_predicate_ids[]` | Core 05 measurement predicates used by the claim. |
| `fixture_ids[]` | Fixture identifiers or visual/accessibility fixture rows used by the claim. |
| `benchmark_profile_id` | Core 05 benchmark profile. |
| `benchmark_manifest_ref` | Retained benchmark manifest reference. |
| `artifact_bundle_sha256` | Digest of the retained publication evidence bundle. |
| `security_controls_state` | Security-control state for the claim-bearing run. |
| `pc_results[]` | Core 05 publication-criteria results. |

## 8. Acceptance Criteria For This Guide

This guide is complete enough for frontend implementation planning when all criteria are true:

| ID | Acceptance criterion |
| --- | --- |
| FE-GUIDE-AC-001 | The guide states that Core 00 through Core 04 remain implementation-conformance authority and that this guide is not a behavior owner. |
| FE-GUIDE-AC-002 | The guide states that Core 05 governs only claim-bearing timed or fixture-sensitive publication. |
| FE-GUIDE-AC-003 | The guide states that UI/UX guide and `docs/design.md` language is design-direction evidence only. |
| FE-GUIDE-AC-004 | The guide states that `docs/domain.md` is vocabulary and concept-boundary guidance only. |
| FE-GUIDE-AC-005 | The guide states that research reports are rationale and not runtime behavior owners. |
| FE-GUIDE-AC-006 | Every product-conformance row has exact owner sections, Core `REQ-*` IDs, and Core `AC-*` IDs; every design/support row has explicit non-Core ownership metadata. |
| FE-GUIDE-AC-007 | Frontend package boundaries match the current development guide, including the prohibition on direct `react-data-grid` imports from `/apps/web`. |
| FE-GUIDE-AC-008 | All 12 frontend phases include scope, primary owner sections, frontend surfaces, introduced behavior, test rows, shared harnesses, completion criteria, and out-of-scope boundaries. |
| FE-GUIDE-AC-009 | Grid adapter requirements cover `record_id`, `field_key`, vendor coordinate translation, presentation-only rows, editability, sort compilation, copy/paste/fill, focus restoration, sparse patches, cleanup, imperative API containment, stylesheet ownership, and styling boundaries. |
| FE-GUIDE-AC-010 | Workbook shell requirements cover continuous shell, top bar, built-in tab order, `System views`, saved-view selector, view bar, grid surface, inspector tabs, status strip, startup fallback, invalid pointer fallthrough, and switcher keyboard behavior. |
| FE-GUIDE-AC-011 | Visual regression requirements cover deterministic fixture data, viewport, zoom, ordering, masks, scroll anchors, focus/editor state, inspector state, settle behavior, fixture matrix, and golden refresh rules. |
| FE-GUIDE-AC-012 | Accessibility requirements cover keyboard access, focus visibility, `System views`, grid navigation/editing, `Esc`, ARIA/accessibility state communication, icon-only labels, contrast, and non-color-only states. |
| FE-GUIDE-AC-013 | Runtime/API boundary requirements cover `/api/v1/`, `/ws/v1/`, server-managed session state, stable IDs, same-origin evidence handles, authorization, closed request shapes, and public error envelopes. |
| FE-GUIDE-AC-014 | Command-surface references include `make browser-e2e-a11y`, frontend phase registry/maps/ledgers, and namespace-aware phase selection. |
| FE-GUIDE-AC-015 | Frontend-local manifests, ledgers, accessibility target, and incomplete visual fixtures are represented as named artifacts or blocked rows, not generic TODOs. |
| FE-GUIDE-AC-016 | Coverage-ledger rules define row-to-owner, row-to-target, browser, visual, accessibility, support, unowned regression, metadata validation, and drift-check expectations. |
| FE-GUIDE-AC-017 | The guide contains no broad "frontend complete" claim without row-level acceptance evidence. |
| FE-GUIDE-AC-018 | The frontend manifest schema requires evidence-class-specific Core IDs and support/design IDs. |
| FE-GUIDE-AC-019 | `make browser-e2e-a11y` emits `cartulary.frontend_accessibility_summary.v1` with row, scenario, keyboard, state-communication, contrast, violation, and artifact-ref arrays. |
| FE-GUIDE-AC-020 | Conditional Core 05 review is inactive by default and activates only when `claim_publication_intent="claim_bearing_publication"`. |
| FE-GUIDE-AC-021 | Frontend phase selection uses `PHASE_NAMESPACE=frontend PHASE=FE-P<N>` and rejects ambiguous frontend phase identifiers. |
| FE-GUIDE-AC-022 | `make check` and `make release-check` include frontend map validation, frontend ledger drift, and accessibility evidence after the target exists. |

## 9. Sources

Claims in this guide are derived from the following local sources:

- `docs/spec/00_document_set_status_and_precedence.md`: document status, precedence, implementation-conformance authority, Core 05 publication boundary, `docs/domain.md` role.
- `docs/spec/01_architecture_storage_and_view_contracts.md`: public HTTP and WebSocket boundary, stable identifiers, session routes, view-row/query/mutation contracts, saved-view/workbook-preference contracts, evidence routes and handles, generated/view-schema registry, system-view contracts, acceptance criteria appendix.
- `docs/spec/02_domain_model_schema_and_history.md`: rough input preservation, mention/entity lifecycle, evidence lifecycle, history/rollback retention, closed vocabularies, party model.
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`: workbook interaction model, built-in tabs, system views, saved views, startup/default surface selection, grid editing, sorting/filtering/grouping, conflicts, Timeline, entity/evidence sheets, coordination surfaces, keyboard/grid invariants.
- `docs/spec/04_security_deployment_and_conformance.md`: session/authentication, current role/current membership authorization, CSRF/WebSocket origin behavior, same-origin evidence access, untrusted content handling, acceptance criteria.
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`: claim-publication scope, fixture-sensitive/timed publication boundaries, predicates `PC-001` through `PC-006`.
- `docs/domain.md`: vocabulary and concept-boundary interpretation.
- `docs/guides/cartulary_implementation_testing_guide.md`: structural model for phase sequencing, test categories, shared harnesses, completion checklist, coverage ledger, grid and visual harness discipline.
- `docs/guides/cartulary-dev-guide.md`: frontend package boundaries, RDG adapter containment, generated TypeScript policy, frontend testing baseline, command surface.
- `docs/guides/cartulary-ui-ux-design-guide.md`: design-direction status, workbook shell composition, system-view grouping, startup surface, status strip, grid editing, conflict/presence, saved-view/query controls, coordination surfaces, evidence UX, visual language, accessibility posture, R2 acceptance criteria.
- `docs/design.md`: visual direction, density, accessibility posture, grid identity direction, design acceptance criteria.
- `docs/testing-harness-nlspec.md`: adopted current authority for harness mechanics only.
- `docs/guides/cartulary_visual_golden_maintenance.md`: visual target, snapshot location, deterministic fixture expectations, scroll normalization, golden refresh procedure.
- `apps/web/e2e/workbook.visual.spec.ts`: current visual Playwright suite location.
- `tools/task_surface.generated.mk`: discovered Make targets.
- `tools/execution_topology_manifest.json`: check/release scheduling metadata.
- `tools/browser_e2e_batch_manifest.json`: browser batch target metadata.
- `tools/phase*_test_map.json` and `docs/testing/phase*_coverage_ledger.md`: existing phase row and coverage-ledger convention.
- `tools/test_accounting_classification.json`: unowned/support classification surface.
- `docs/research/R09-react-data-grid-research-report.md`: RDG-specific rationale for row identity, sorting, editability, group rows, browser risk, stylesheet, and adapter containment. This report is not used as product-conformance authority.
