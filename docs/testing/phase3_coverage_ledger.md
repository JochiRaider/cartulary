# Phase 3 Coverage Ledger

This ledger is the authoritative repo-local map for Phase 3 test evidence.

- Scope: Timeline workbook create, patch, query, lifecycle actions, idempotent replay, projection rebuild, and workbook-visible collaboration behavior.
- Normative owners: Core 03 `§6`, `§7`, `§15`; Core 01 `§3.3.5`, `§7.4.1`; Core 04 `AC-043`, `AC-191` through `AC-199`, `AC-329` through `AC-331`.
- Support-only evidence: no separate Phase 3 process-only suite exists in `cmd/server`; authoritative Phase 3 browser evidence lives under `apps/web/e2e`.

## Unit

| Row | Owner sections | Evidence |
| --- | --- | --- |
| `U-3-01` | Core 03 `§7`; Core 04 `AC-191`, `AC-192` | `internal/modules/timeline/phase3_create_request_test.go::TestPhase3_CreateRequestContracts_U_3_01` |
| `U-3-02` | Core 03 `§6`; Core 04 `AC-191` | `internal/modules/timeline/phase3_create_request_test.go::TestPhase3_CreateInitialStateVocabulary_U_3_02` |
| `U-3-03` | Core 03 `§6`; Core 04 `AC-193`, `AC-195` | `internal/modules/timeline/phase3_state_test.go::TestPhase3_CaptureStateTransitions_U_3_03` |
| `U-3-04` | Core 03 `§6`; Core 04 `AC-195`, `AC-197` | `internal/modules/timeline/phase3_state_test.go::TestPhase3_ReviewedRowsDemoteAndSupersededRowsAreTerminal_U_3_04` |
| `U-3-06` | Core 03 `§15`; Core 01 `§7.4.1`; Core 04 `AC-192` | `internal/modules/timeline/phase3_create_request_test.go::TestPhase3_PatchPayloadValidation_U_3_06` |
| `U-3-07` | Core 01 `§3.3.5`; Core 04 `AC-199`, `AC-330` | `internal/modules/timeline/phase3_create_request_test.go::TestPhase3_PatchRequestHashNormalization_U_3_07` |
| `U-3-08` | Core 03 `§15`; Core 01 `§7.4.1` | `internal/modules/timeline/phase3_projection_payload_test.go::TestPhase3_ProjectionRowsKeepStableBindingAndDerivedFields_U_3_08` |
| `U-3-09` | Core 01 `§3.3.5`; Core 03 `§15` | `internal/modules/timeline/phase3_projection_payload_test.go::TestPhase3_TimelinePayloadBuildersExposeStableShapes_U_3_09` |
| `U-3-10` | Core 03 `§6`; Core 04 `AC-329`, `AC-330` | `internal/modules/timeline/phase3_projection_payload_test.go::TestPhase3_SupersedeGuardsAndActionHashes_U_3_10` |

## Integration

| Row | Owner sections | Evidence |
| --- | --- | --- |
| `I-3-01` | Core 03 `§6`, `§7`, `§15`; Core 04 `AC-191` through `AC-199`, `AC-330`, `AC-331` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_01_CreatePatchReplayAndRollback` |
| `I-3-02` | Core 03 `§15`; Core 01 `§7.4.1` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild` |
| `I-3-03` | Core 03 `§6`, `§15`; Core 04 `AC-194`, `AC-196`, `AC-197`, `AC-329`, `AC-330`, `AC-331` | `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions` |

## Browser E2E

| Row | Owner sections | Evidence |
| --- | --- | --- |
| `E-3-01` | Core 03 `§7`; Core 04 `AC-191` | `apps/web/e2e/phase3.spec.ts::E-3-01 creates a Timeline row in-grid and continues editing on the draft row` |
| `E-3-02` | Core 04 `AC-043`; Core 05 measurement predicates `perf.typing_ack.v1`, `perf.timeline_blank_row_create.v1` | `apps/web/e2e/phase3.spec.ts::E-3-02 measures user-visible typing_ack and blank-row-create completion within the Phase 3 envelope` |
| `E-3-03` | Core 03 `§6`; Core 04 `AC-194`, `AC-195`, `AC-196`, `AC-197` | `apps/web/e2e/phase3.spec.ts::E-3-03 drives review, demotion, and supersede through the visible workbook surface` |
| `E-3-04` | Core 01 `§3.3.5`; Core 04 `AC-199`, `AC-330`, `AC-331` | `apps/web/e2e/phase3.spec.ts::E-3-04 uses a disclosed hybrid replay harness to prove replay avoids duplicate history and visible invalidation` |

## Shared Harness Coverage

| Harness | Phase 3 evidence |
| --- | --- |
| Envelope consistency and default query meta | `internal/testutil/httptestx/httptestx.go::RequireSuccessEnvelope`, `RequireErrorEnvelope`, and `internal/testutil/httptestx/crosscutting.go::RequireDefaultQueryMeta`, exercised by `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_01_CreatePatchReplayAndRollback`, `TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild`, and `TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions`. |
| Mutation attribution, revision emission, and coupled-history assertions | `internal/testutil/httptestx/crosscutting.go::RequireMutationAttribution` plus `internal/testutil/timelinetest/timelinetest.go::LookupChangeSet`, `CountChangeSetMutations`, `CountRecordRevisions`, and `CountActiveSupersedesLinks`, exercised by `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_01_CreatePatchReplayAndRollback` and `TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions`. |
| Projection-row substrate inspection | `internal/testutil/timelinetest/timelinetest.go::LookupProjectionRow` exercised by `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_01_CreatePatchReplayAndRollback` and `TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions`. |
| Replay stability and divergent replay rejection | `internal/testutil/httptestx/crosscutting.go::RequireReplayScaffold` exercised by `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_01_CreatePatchReplayAndRollback` and `TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions`. |
| Closed-vocabulary and field-key fail-closed validation | `internal/testutil/httptestx/crosscutting.go::RequireClosedVocabularyRejected` exercised by `internal/modules/timeline/phase3_create_request_test.go::TestPhase3_PatchPayloadValidation_U_3_06` and `internal/modules/timeline/phase3_integration_test.go::TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild`. |
| Real-service integration harness | `internal/testutil/pgtest`, `internal/testutil/s3test`, `internal/testutil/httptestx`, and `internal/testutil/wstest`, exercised across all Phase 3 integration tests under `internal/modules/timeline/phase3_integration_test.go`. |
| Browser timing and disclosed hybrid replay harness | `apps/web/e2e/helpers.ts::measureTypingAck`, `measureBlankRowCreate`, `fetchTimelineRecordSubstrate`, and `fetchTimelineRecordChangeCount`, exercised by `apps/web/e2e/phase3.spec.ts::E-3-02 measures user-visible typing_ack and blank-row-create completion within the Phase 3 envelope` and `E-3-04 uses a disclosed hybrid replay harness to prove replay avoids duplicate history and visible invalidation`. |
