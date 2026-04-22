# Phase 3 Coverage Ledger

This ledger is the human-readable companion to `tools/phase3_test_map.json`.

- Scope: Timeline workbook create, patch, query, lifecycle actions, replay stability, projection rebuild, save-state UI, and browser-visible collaboration behavior.
- Normative owners: Core 03 `§6`, `§7`, `§15`; Core 01 `§3.3.5`, `§7.4.1`; Core 04 `AC-043`, `AC-191` through `AC-199`, `AC-329` through `AC-331`.
- Authority: `tools/phase3_test_map.json` is the enforced Phase 3 traceability source. This ledger summarizes the same surface in prose and does not control the mechanical row inventory.
- Timeline zero-field create traceability: cite Core 01 `REQ-01-057` plus Core 04 `AC-191` and `AC-192` for the owner rule. `contracts/view-schemas/cartulary.view.timeline.v1.json` is derived evidence only and is not the behavior source.
- Authoritative execution:
  - `backend-unit` selects authoritative `U-3-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase3 unit authoritative backend_unit`.
  - `backend-store` selects store-backed authoritative `U-3-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase3 unit authoritative backend_store`.
  - `backend-integration` selects authoritative `I-3-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase3 integration authoritative backend_integration`.
  - `frontend-unit` selects authoritative `U-3-*` workbook rows only through the Phase 3 Vitest manifest for `frontend_unit`.
  - `browser-e2e-webserver-backed` and delegated `browser-e2e-functional` select authoritative functional `E-3-*` rows only through the Phase 3 Playwright manifest for `browser_functional`.
  - `browser-e2e-measurement` selects authoritative measurement `E-3-*` rows only through the Phase 3 Playwright manifest for `browser_measurement`.
- Support-only evidence: `internal/modules/timeline/phase3_support_test.go` and `internal/modules/timeline/phase3_support_integration_test.go` remain regression coverage only. They do not satisfy authoritative `U-3-*`, `I-3-*`, or `E-3-*` rows.
- Support-only execution:
  - `internal/modules/timeline/phase3_support_test.go` runs through `backend-unit` with `TestSupportPhase3Unit_` and is forbidden from claiming `U-3-*` identifiers.
  - `internal/modules/timeline/phase3_support_integration_test.go` runs through `backend-integration-support` with `TestSupportPhase3Integration_` and is forbidden from claiming `I-3-*` identifiers.

## Unit

| Row | Evidence | Task target | Real services | Major assertions | Remaining gap |
| --- | --- | --- | --- | --- | --- |
| `U-3-01` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_CreateCommitsAndAssignsIdentity_U_3_01` | `backend-store` | PostgreSQL | A one-value Timeline create commits a real row, assigns `record_id`, starts at `row_version=1`, and persists the Timeline projection row. | Browser-visible create flow and draft-row continuation remain browser evidence. |
| `U-3-02` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_InitialCreateState_U_3_02` | `backend-store` | PostgreSQL | Real committed create initializes `capture_state='rough'` and rejects obsolete lifecycle vocabulary through the store path rather than helper constants alone. | None for the declared store-owned contract. |
| `U-3-03` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_CaptureStateLifecycle_U_3_03` | `backend-store` | PostgreSQL | The first material patch demotes `rough -> enriched`, and reviewer-authorized `mark-reviewed` persists the explicit `reviewed` transition on real store state. | UI disclosure and browser action affordances remain browser evidence. |
| `U-3-04` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_ReviewedDemotionAndSupersedeTerminality_U_3_04` | `backend-store` | PostgreSQL | A material edit demotes `reviewed -> enriched`, legal supersede persists `superseded`, and ordinary patch attempts fail closed afterwards. | Replacement-target guards and rollback coupling are owned by `U-3-10` and integration rows. |
| `U-3-05` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels` | `frontend-unit` | None | The authoritative workbook component row proves no explicit Save button is required, autosave commits on Enter, Tab, blur, and paste completion, exact `Syncing` / `Saved` / `Conflict` labels remain stable, and the continuation-row helper contract stays bound to the same test. | Focus continuation on the real browser surface remains `E-3-01`. |
| `U-3-06` | `internal/modules/timeline/phase3_decoder_test.go::TestPhase3_PatchPayloadValidation_U_3_06` | `backend-unit` | None | Patch decode fails closed on missing required members, duplicate `field_key` entries, empty `changes[]`, protected field mutations, and unknown top-level members. | Durable patch consequences remain store or integration evidence. |
| `U-3-07` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_PatchReplayStability_U_3_07` | `backend-store` | PostgreSQL | Patch replay returns the original committed result and creates no second mutation row or record revision on the write substrate. | Collaboration suppression remains integration and browser evidence, not store-only evidence. |
| `U-3-08` | `internal/modules/timeline/phase3_projection_contract_test.go::TestPhase3_ProjectionContract_U_3_08` | `backend-unit` | None | Timeline projection rows expose the full grid-owned field contract and keep stable `record_id` / `row_version` binding plus derived field shapes. | Persisted rebuild maintenance remains integration evidence. |
| `U-3-09` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_CreateAndPatchWriteHistory_U_3_09` | `backend-store` | PostgreSQL | Successful Timeline create and patch each write one attributed mutation entry and one new record revision on the real write substrate. | Browser-visible history inspection is deferred to later phases. |
| `U-3-10` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_SupersedeReplayAndRollbackCoupling_U_3_10` | `backend-store` | PostgreSQL + store rollback hooks | Supersede-with-replacement rejects illegal targets, replays idempotently, writes one coupled replacement link, and rolls back both lifecycle and link state together on hook-induced failure. | End-to-end action routing and visibility still depend on integration and browser evidence. |
| `U-3-GRID-01` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-GRID-01 binds Timeline grid columns from the active view_schema and commits writable cells by field_key` | `frontend-unit` | None | The authoritative workbook row proves the rendered Timeline columns and labels come from the active contract default-visible field set, omits hidden contract fields from the grid, and emits a PATCH bound to the authoritative `field_key` instead of the visible label. | Full browser routing remains supplemental because the authoritative owner here is the unit-layer binding contract. |
| `U-3-GRID-02` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-GRID-02 binds saved rows by record_id and row_version instead of visible row index` | `frontend-unit` | None | The authoritative workbook row proves the viewport keeps saved-row identity bound to `record_id` and `row_version`, and the emitted PATCH route and `base_row_version` come from that bound row rather than visible row position. | Real browser reorder regressions remain supplemental coverage. |
| `U-3-GRID-03` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-GRID-03 keeps sorted and filtered local edits bound to the original record_id, base_row_version, and field_key` | `frontend-unit` | None | The authoritative workbook row proves sort and filter controls can reorder the visible viewport without breaking mutation identity, and a local edit still emits the original bound `record_id`, `base_row_version`, and `field_key`. | End-to-end browser interaction remains supplemental because the authoritative owner here is the unit-layer binding path. |

## Integration

| Row | Evidence | Task target | Real services | Major assertions | Remaining gap |
| --- | --- | --- | --- | --- | --- |
| `I-3-01` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_01_CreatePatchReplayAndRollback` | `backend-integration` | PostgreSQL + MinIO + real runtime + real WebSocket boundary | Real HTTP create, patch, review, supersede, replay, mutation history, projection maintenance, WebSocket invalidation, and rollback all stay transactionally coherent. | None for the declared integration contract. |
| `I-3-02` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild` | `backend-integration` | PostgreSQL + MinIO + real runtime | The query route reads projection rows instead of source rows, and deterministic rebuild restores the same Timeline query surface. | None for the declared projection-query contract. |
| `I-3-03` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions` | `backend-integration` | PostgreSQL + MinIO + real runtime + real WebSocket boundary | Incident-role authorization, review or supersede legality, replay semantics, replacement-target guards, and rollback of supersede link or projection state all hold on the live route surface. | Systematic route-matrix coverage is support-only evidence, not authoritative row coverage. |

## Browser E2E

| Row | Evidence | Task target | Real services | Major assertions | Remaining gap |
| --- | --- | --- | --- | --- | --- |
| `E-3-01` | `apps/web/e2e/phase3.spec.ts::E-3-01 creates a Timeline row in-grid and continues editing on the draft row` | `browser-e2e-webserver-backed` | Real browser + Go server + PostgreSQL + MinIO + Vite | A user creates a Timeline row on the real workbook surface and immediately continues editing on the fresh draft row without leaving the grid. | None for the declared browser create-flow contract. |
| `E-3-02` | `apps/web/e2e/measurement/phase3_measurement.spec.ts::E-3-02 measures user-visible typing_ack and blank-row-create completion within the Phase 3 envelope` | `browser-e2e-measurement` | Real browser + Go server + PostgreSQL + MinIO + Vite | The isolated measurement suite proves visible `Syncing -> Saved` save-state transitions once on the real stack and then measures typing acknowledgement and blank-row-create completion against the declared Phase 3 p95 envelope. | `Conflict` labeling remains covered at the component layer rather than the isolated timing suite. |
| `E-3-03` | `apps/web/e2e/phase3.spec.ts::E-3-03 drives review, demotion, and supersede through the visible workbook surface` | `browser-e2e-webserver-backed` | Real browser + Go server + PostgreSQL + MinIO + Vite | The visible workbook surface supports review, reviewed-row demotion after material edit, legal supersede, and post-supersede disabling. | None for the declared browser lifecycle contract. |
| `E-3-04` | `apps/web/e2e/phase3.spec.ts::E-3-04 uses a disclosed hybrid replay harness to prove replay avoids duplicate history and visible invalidation` | `browser-e2e-webserver-backed` | Real browser + Go server + PostgreSQL + MinIO + Vite + disclosed test-only snapshot routes | Real browser replay preserves the original committed result, avoids duplicate history growth, and suppresses duplicate visible invalidation. | The disclosed snapshot harness is intentionally test-only and not part of runtime behavior. |

## Shared Harness Coverage

| Harness | Phase 3 evidence |
| --- | --- |
| Real runtime, store, and socket harness | `internal/testutil/phase3test` now centralizes the Postgres + MinIO runtime boot path, HTTP session helpers, incident or membership seeding, and Timeline WebSocket assertions shared by authoritative and support-only Phase 3 tests. |
| Cross-cutting HTTP and replay helpers | `internal/testutil/httptestx` still owns success or error envelope checks, replay scaffolding, mutation attribution helpers, and closed-vocabulary assertions used across the Phase 3 backend suite. |
| Timeline substrate inspection helpers | `internal/testutil/timelinetest` continues to provide projection-row, change-set, revision-count, and supersede-link inspection used by the Phase 3 store and integration slices. |
| Browser timing and replay helpers | `apps/web/e2e/helpers.ts` provides the Phase 3 timing predicates, substrate snapshot accessors, and browser auth helpers shared by `E-3-02` and `E-3-04`. |

## Support-Only Evidence

- `internal/modules/timeline/phase3_support_test.go` keeps helper-level regression coverage for request-shape helpers, vocabulary helpers, hash normalization, payload builders, and supersede guards. These tests run under `TestSupportPhase3Unit_` and are intentionally forbidden from carrying authoritative Phase 3 IDs.
- `internal/modules/timeline/phase3_support_integration_test.go::TestSupportPhase3Integration_AuthorizationMatrix` table-drives create, query, patch, review, and supersede authorization across no-membership, editor, reviewer, and admin states. It runs through `backend-integration-support`, strengthens route inventory confidence, and does not replace `I-3-03`.
