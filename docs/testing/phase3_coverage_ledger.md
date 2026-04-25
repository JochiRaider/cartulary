# Phase 3 Coverage Ledger

This ledger is generated from `tools/phase3_test_map.json`. Update the manifest row metadata first, then regenerate this file.

- Scope: Timeline workbook create, patch, query, lifecycle actions, replay stability, projection rebuild, save-state UI, and browser-visible collaboration behavior.
- Normative owners: Core 03 `§6`, `§7`, `§15`; Core 01 `§3.3.5`, `§7.4.1`; Core 04 `AC-043`, `AC-191` through `AC-199`, `AC-329` through `AC-331`.
- Authority: `tools/phase3_test_map.json` is the enforced Phase 3 traceability source. This ledger is a rendered companion and does not control the mechanical row inventory.
- Timeline zero-field create traceability: cite Core 01 `REQ-01-057` plus Core 04 `AC-191` and `AC-192` for the owner rule. `contracts/view-schemas/cartulary.view.timeline.v1.json` is derived evidence only and is not the behavior source.
- Grid note: the workbook currently renders through the RDG-backed `@cartulary/grid-adapter`. `U-3-GRID-01/02/03` own workbook binding behavior, while vendor-specific RDG semantics stay with adapter tests.

## Authoritative Execution

- `backend-unit` selects authoritative `U-3-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase3 unit authoritative backend_unit`.
- `backend-store` selects store-backed authoritative `U-3-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase3 unit authoritative backend_store`.
- `backend-integration` selects authoritative `I-3-*` rows only through `RUN_GO_MANIFEST_PHASE ... phase3 integration authoritative backend_integration`.
- `frontend-unit` selects authoritative `U-3-*` workbook rows only through the Phase 3 Vitest manifest for `frontend_unit`.
- `browser-e2e-webserver-backed` and helper `browser-e2e-functional` select authoritative functional `E-3-*` rows through the batched Playwright manifest selection for `browser_functional`.
- `browser-e2e-measurement` selects authoritative measurement `E-3-*` rows only through the Phase 3 Playwright manifest for `browser_measurement`.

## Support-Only Execution

- `internal/modules/timeline/phase3_support_test.go` runs through `backend-unit` with `TestSupportPhase3Unit_` and is forbidden from claiming `U-3-*` identifiers.
- `internal/modules/timeline/phase3_support_integration_test.go` runs through `backend-integration-support` with `TestSupportPhase3Integration_` and is forbidden from claiming `I-3-*` identifiers.

## Unit

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `U-3-01` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_CreateCommitsAndAssignsIdentity_U_3_01` | `backend_store` | A one-value Timeline create commits a real row, assigns `record_id`, starts at `row_version=1`, and persists the Timeline projection row. | Browser-visible create flow and draft-row continuation remain browser evidence. |
| `U-3-02` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_InitialCreateState_U_3_02` | `backend_store` | A committed Timeline create initializes `capture_state='rough'` on the persisted row and projection state. | Obsolete lifecycle-token rejection remains supplemental helper coverage until a public executed submission path owns that proof. |
| `U-3-03` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_CaptureStateLifecycle_U_3_03` | `backend_store` | The first material patch transitions `rough -> enriched`, and store-driven `mark-reviewed` persists an explicit `reviewed` transition on real store state. | Reviewer or admin authorization is owned by `I-3-03`, not this store-only row. |
| `U-3-04` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_ReviewedDemotionAndSupersedeTerminality_U_3_04` | `backend_store` | A material edit demotes `reviewed -> enriched`, legal supersede persists `superseded`, and ordinary patch attempts fail closed afterwards. | Replacement-target guards and rollback coupling are owned by `U-3-10` and integration rows. |
| `U-3-05` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels` | `frontend_unit` | The workbook component proves autosave commits on Enter, Tab, blur, and paste completion without a Save button, with exact `Syncing`, `Saved`, and `Conflict` labels. | Focus continuation on the real browser surface remains `E-3-01`. |
| `U-3-06` | `internal/modules/timeline/phase3_decoder_test.go::TestPhase3_PatchPayloadValidation_U_3_06` | `backend_unit` | Patch decode fails closed on missing required members, duplicate `field_key` entries, empty `changes[]`, protected field mutations, and unknown top-level members. | Durable patch consequences remain store or integration evidence. |
| `U-3-07` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_PatchReplayStability_U_3_07` | `backend_store` | Patch replay returns the original committed result and creates no second mutation row or record revision on the write substrate. | Collaboration suppression remains integration and browser evidence. |
| `U-3-08` | `internal/modules/timeline/phase3_projection_contract_test.go::TestPhase3_ProjectionContract_U_3_08` | `backend_unit` | Timeline projection rows expose the full grid-owned field contract and keep stable `record_id` and `row_version` binding plus derived field shapes. | Persisted rebuild maintenance remains integration evidence. |
| `U-3-09` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_CreateAndPatchWriteHistory_U_3_09` | `backend_store` | Successful Timeline create and patch each write one attributed mutation entry and one new row revision on the real write substrate. | Browser-visible history inspection is deferred to later phases. |
| `U-3-10` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_SupersedeReplayAndRollbackCoupling_U_3_10` | `backend_store` | Supersede-with-replacement rejects illegal targets, replays idempotently, writes one coupled `timeline_record` plus one `record_link` mutation in one change set, and rollback removes both lifecycle and link state together. | End-to-end action routing and visible browser disclosure remain integration and browser evidence. |
| `U-3-11` | `internal/modules/timeline/phase3_store_test.go::TestPhase3_PatchFieldLevelConcurrency_U_3_11` | `backend_store` | Stale Timeline patches derive committed writable field changes since the client base row version. Different-field edits rebase, same-field direct edits return `same_field_conflict`, collection fields use `collection_value_v1`, lifecycle-only stale edits apply against current state, and replay precedence remains idempotent. | HTTP conflict envelope transport and browser-visible concurrent editing behavior remain integration or browser evidence. |
| `U-3-12` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-12 builds zero-field Timeline create payloads only for explicit blank-row creation` | `frontend_unit` | The workbook create payload builder omits zero-field creates during ordinary draft autosave, but emits `client_txn_id`-only payloads for explicit blank-row creation. | Server-side zero-field validation and durable initial row state remain backend store or integration evidence. |
| `U-3-13` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-13 creates an explicit blank Timeline row with only client_txn_id and suppresses duplicate pending submits` | `frontend_unit` | The workbook blank-row action submits exactly one pending `client_txn_id`-only create for a duplicate click burst and renders the committed rough Timeline row plus a fresh draft row. | Real browser create continuation remains `E-3-01`; backend replay and projection durability remain integration evidence. |
| `U-3-GRID-01` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-GRID-01 binds Timeline grid columns from the active view_schema and commits writable cells by field_key` | `frontend_unit` | The RDG-backed workbook grid binds Timeline columns from the active `view_schema` and commits writable cells by authoritative `field_key`. | Vendor-specific RDG behavior stays with adapter tests; this row owns workbook binding only. |
| `U-3-GRID-02` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-GRID-02 binds saved rows by record_id and row_version instead of visible row index` | `frontend_unit` | The RDG-backed workbook grid keeps saved-row identity bound to `record_id` and `row_version` rather than visible row position. | Real browser reorder regressions remain supplemental coverage. |
| `U-3-GRID-03` | `apps/web/src/WorkbookShell.phase3.test.tsx::Phase 3 U-3-GRID-03 keeps sorted and filtered local edits bound to the original record_id, base_row_version, and field_key` | `frontend_unit` | The RDG-backed workbook grid keeps sorted and filtered local edits bound to the original `record_id`, `base_row_version`, and `field_key`. | End-to-end browser interaction remains supplemental because this row owns workbook binding semantics. |

## Integration

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `I-3-01` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_01_CreatePatchReplayAndRollback` | `backend_integration` | Real HTTP create, patch, review, supersede, replay, mutation history, projection maintenance, WebSocket invalidation, and rollback stay transactionally coherent. | None. |
| `I-3-02` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild` | `backend_integration` | The query route reads projection rows instead of source rows, and deterministic rebuild restores the same Timeline query surface. | None. |
| `I-3-03` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions` | `backend_integration` | Incident-role authorization, review and supersede legality, replay semantics, replacement-target guards, and supersede change-set shape all hold on the live route surface. | Systematic route-matrix inventory remains support-only evidence. |
| `I-3-04` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_PatchSameFieldConflictEnvelope_I_3_04` | `backend_integration` | Same-field Timeline patch conflicts are transported as `409 same_field_conflict` with `error.conflict` and no mutation or idempotency write. | Field-level rebase and conflict payload construction remain store-domain evidence; browser conflict resolution remains Phase 6 evidence. |

## Browser E2E

| Row | Evidence | Execution | Claim | Out of scope |
| --- | --- | --- | --- | --- |
| `E-3-01` | `apps/web/e2e/phase3.spec.ts::E-3-01 creates a Timeline row in-grid and continues editing on the draft row` | `browser_functional` | A user creates a Timeline row on the real workbook surface and immediately continues editing on the fresh draft row without leaving the grid. | None. |
| `E-3-02` | `apps/web/e2e/measurement/phase3_measurement.spec.ts::E-3-02 measures user-visible typing_ack and blank-row-create completion within the Phase 3 envelope` | `browser_measurement` | The isolated measurement suite proves one real-stack visible `Syncing -> Saved` transition under controlled latency, then measures typing acknowledgement and blank-row-create completion against the Phase 3 envelope. | `Conflict` labeling remains component-layer evidence. |
| `E-3-03` | `apps/web/e2e/phase3.spec.ts::E-3-03 drives review, demotion, and supersede through the visible workbook surface` | `browser_functional` | A reviewer-session browser flow visibly performs review, material-edit demotion, legal supersede, and post-supersede disabling on the workbook surface. | Durable substrate counts and replay internals remain backend or disclosed hybrid evidence. |
| `E-3-04` | `apps/web/e2e/phase3.spec.ts::E-3-04 uses a disclosed hybrid replay harness to prove replay avoids duplicate history and visible invalidation` | `browser_functional` | Real browser replay preserves the original committed result, avoids duplicate history growth, and suppresses duplicate visible invalidation. | The disclosed snapshot harness is intentionally test-only and not part of runtime behavior. |

## Shared Harness Coverage

| Harness | Phase 3 evidence |
| --- | --- |
| Real runtime, store, and socket harness | `internal/testutil/phase3test` centralizes the Postgres + MinIO runtime boot path, HTTP session helpers, incident or membership seeding, and Timeline WebSocket assertions shared by authoritative and support-only Phase 3 tests. |
| Cross-cutting HTTP and replay helpers | `internal/testutil/httptestx` owns success or error envelope checks, replay scaffolding, mutation attribution helpers, and closed-vocabulary assertions used across the Phase 3 backend suite. |
| Timeline substrate inspection helpers | `internal/testutil/timelinetest` owns projection-row, change-set, mutation-row, revision-count, and supersede-link inspection used by the Phase 3 store and integration slices. |
| Browser timing and replay helpers | `apps/web/e2e/helpers.ts` provides the Phase 3 timing predicates, tracked-user browser auth helpers, and substrate snapshot accessors shared by `E-3-02`, `E-3-03`, and `E-3-04`. |

## Support-Only Evidence

- `internal/modules/timeline/phase3_support_test.go` keeps helper-level regression coverage for request-shape helpers, vocabulary helpers, hash normalization, payload builders, and supersede guards. These tests run under `TestSupportPhase3Unit_` and are intentionally forbidden from carrying authoritative Phase 3 IDs.
- `internal/modules/timeline/phase3_support_integration_test.go::TestSupportPhase3Integration_AuthorizationMatrix` table-drives create, query, patch, review, and supersede authorization across no-membership, editor, reviewer, and admin states. It strengthens route inventory confidence and does not replace `I-3-03`.
