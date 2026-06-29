# Timeline Module Refactoring Tracker and Handoff Planner

## 1. Session Header

| Field | Value |
| --- | --- |
| Target directory | `internal/modules/timeline` |
| Branch / commit | `main` / `8beaff1f4919fc3e0c7840dd43ce86ac421db6f4` |
| Dirty tree | Dirty at characterization update time: added `docs/handoffs/Timeline-Module-Refactoring-Tracker.md`, added `internal/modules/timeline/facade.go`, modified `internal/modules/imports/routes.go`, `internal/modules/timeline/routes.go`, `internal/modules/timeline/phase3_decoder_test.go`, `internal/modules/timeline/phase3_integration_test.go`, `internal/modules/workbook/routes.go`, `internal/modules/workbook/store.go` |
| Date/time | Initial planning: `2026-06-28 20:48:32 EDT -0400`; facade update: `2026-06-28 21:02:43 EDT -0400`; characterization update: `2026-06-28 21:15:58 EDT -0400` |
| Agent | Codex |
| Mode | Implementation update after planning; no generated-file edits, no route/schema/UI redesign, no test rewrites |
| Framework | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Source limits | `AGENTS.md` and framework read. `internal/modules/timeline` implementation files inspected for facade scope. Characterization audit directly searched and targeted-read Timeline phase3/phase4/phase9 tests, Workbook phase6/phase9 tests, Revisions phase7 rollback/history tests, Core 01/04 time-conversion requirements, OpenAPI/view schema/WS/migration/sqlc references, and adjacent Workbook/Imports/Revisions seams. Frontend and generated outputs were searched only; inspect them before frontend or contract changes. |

## Summary

This tracker plans a behavior-preserving refactor of `internal/modules/timeline` toward a cleaner backend module boundary. The Timeline module’s target responsibility is low-friction timeline capture and mutation: rough capture, Timeline v2 DTO validation/defaulting, row-version/idempotent mutation handling, mention/entity action hooks, projection/revision/collaboration effects, and stable public route-facing results. Current repo evidence shows much of that behavior lives in `Store`, while route/auth/WebSocket handling and caller-specific error mapping are spread across Timeline, Workbook, Imports, and Revisions seams.

No owner contradiction was found in inspected sources. If later inspection finds conflict between Core 00-04, adopted NLSpecs, generated contracts, or code, mark `BLOCKED: owner contradiction`.

Implementation update: completed the facade-first runtime slice. `timeline.Facade` now wraps existing `Store` behavior, Timeline HTTP handlers use the facade, Workbook holds a Timeline facade for Timeline query/create/patch/conflict paths, and Imports uses an explicit imported-create facade method. No persistence internals, contracts, migrations, generated files, tests, route paths, envelopes, or WebSocket payload construction were changed.

## 2. Target Scope And Non-Scope

In scope:
- `internal/modules/timeline/{api.go,routes.go,state.go,store.go,auto_resolution.go,clipboard_paste.go,clipboard_paste_store.go,hooks.go,timing.go}`.
- Timeline tests under `internal/modules/timeline/*_test.go`.
- Adjacent callers and seams inspected for boundary impact: Workbook routes/store/bulk/telemetry, Imports apply path, Revisions rollback mapping, App route registration, DB migrations/queries, OpenAPI, WS schema, view schema, generated references, harness docs.

Non-scope unless proven inseparable by implementation:
- Route redesign, schema redesign, UI redesign, phase-accounting rewrites, generated-file hand edits.
- Product behavior changes to Timeline create/patch/paste/review/supersede/conflict semantics.
- Replacing Core-owned route envelopes, view schemas, WebSocket payloads, migrations, sqlc query contracts, or harness accounting.
- Introducing phase-shaped runtime dependencies.

## 3. Current-State Inventory

| Path | Current responsibility | Target owner | Public/private/generated | Incoming deps | Outgoing deps | External contracts touched | Risk | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/timeline/api.go` | Timeline DTOs, decode/validate/hash, row/action payload builders, errors | Timeline facade/API boundary | Public within Go module | Workbook, Imports, tests | `viewschema`, `httpapi`, `fieldnorm` | OpenAPI, Core 01/03 mutation contracts | High | Public surface is broad; keep DTO/error compatibility first. |
| `internal/modules/timeline/facade.go` | Public backend facade delegating to existing `Store` for Timeline commands/queries/profile/substrate behavior | Timeline facade/API boundary | Public within Go module | Timeline routes, Workbook, Imports | existing `Store` | No public wire contract; wraps existing behavior | Medium | Added in implementation slice; intentionally thin compatibility layer. |
| `internal/modules/timeline/routes.go` | Timeline HTTP routes for time-conversion profile, mark-reviewed, test WS hooks; auth/session/role/WS publish | Transport adapter over Timeline facade | Public route registration plus private handlers | App runtime | `timeline.Facade`, platform auth/http/ws, incidents/auth/entities | Core 01 routes, Core 04 auth/CSRF, WS schema | High | Now calls facade for Timeline store behavior; platform logic still remains in route adapter. |
| `internal/modules/timeline/store.go` | Main transactional implementation: query, create, patch, conflict, actions, revisions, projection, links, mentions, idempotency | Private Timeline application/persistence layer behind facade | Exported `Store` still used by lower-level tests and wrapped by `Facade` | Facade, tests | authn, records, revisions, projections, links, entities, SQL | DB state, revision history, WS payload inputs | Very high | Store internals not moved in this slice. |
| `internal/modules/timeline/state.go` | Capture-state lifecycle helpers | Timeline domain/application | Private-ish helpers | Store/tests | none significant | Core 03 lifecycle | High | Must preserve rough/enriched/reviewed/superseded behavior. |
| `internal/modules/timeline/auto_resolution.go` | Mention auto-resolution eligibility and alias lookup | Timeline mention hook / entity port | Private | Store/tests | entities tables, normalization | Core 02/03 mention auto-resolution | High | Keep unresolved fallback and no-stub guarantee. |
| `internal/modules/timeline/clipboard_paste.go` | Paste request decode, plan, header mapping, field conversion | Timeline ingest adapter | Public within module | Workbook bulk/paste routes, tests | `imports/tabularingest`, viewschema | Core 01/03 clipboard contract | High | Shared tabular-ingest dependency is intentional; guard against local parser drift. |
| `internal/modules/timeline/clipboard_paste_store.go` | Batch paste transaction, conflicts, idempotency, projections | Timeline application/persistence | Exported method on `Store` | Workbook routes/bulk/tests | records/revisions/projections/mentions/links | Core 03 paste conflict semantics | Very high | Needs separate paste characterization before split. |
| `internal/modules/timeline/hooks.go` | Test-only pre-commit hook | Test seam | Public test hook in production package | Tests | none | Harness/test behavior | Medium | Keep private or build-tagged only after tests prove no external use. |
| `internal/modules/timeline/timing.go` | Create timing instrumentation interface | Adapter/telemetry seam | Public within module | Workbook create route | timing recorder only | Implementation support | Low | Keep out of domain facade if possible. |
| `internal/modules/timeline/phase3_*_test.go` | Timeline create/patch/lifecycle/projection/HTTP/WS/time-conversion tests | Behavior tests | Test only | Go test runner | testutil | Core 01/03/04 behavior evidence | High | Characterization audit covered named tests in Section 6; full-suite rewrite remains out of scope. |
| `internal/modules/timeline/phase4_*_test.go` | Mention/entity/evidence relationship tests | Behavior tests | Test only | Go test runner | entities/evidence fixtures | Core 02/03 evidence | High | Characterization audit covered named import and auto-resolution tests in Section 6. |
| `internal/modules/timeline/phase9_clipboard_paste_test.go` | Paste parsing/header mapping support tests | Behavior tests | Test only | Go test runner | paste helpers | Core 03 ingest | High | Covered parser/header/cross-incident target helpers; batch persistence evidence lives in Workbook phase9 integration tests. |
| `internal/modules/workbook/{routes.go,store.go,bulk_mutation_api.go,telemetry.go}` | Generic workbook route/store layer dispatching to Timeline | Workbook caller over Timeline facade | Public internal module | App/web/API | `timeline.Facade`, DTOs/errors | Row create/query/patch/paste/bulk/supersede | Very high | Runtime Timeline store calls now go through facade; DTO/error imports remain. |
| `internal/modules/imports/routes.go` | Import apply path maps approved rows to Timeline create | Imports caller over Timeline import facade | Public internal module | Jobs/import API | `timeline.Facade`, `timeline.DecodeTimelineCreateRequest` | Import extension, mention_origin ingest | High | Uses explicit `CreateImportedTimelineRow`; preserves imported-create no-interactive-auto-resolution path. |
| `internal/modules/revisions/rollback_store.go` | Rollback changed-field mapping for Timeline links/tags/source fields | Revisions owner with Timeline port/table map | Public internal module | Revisions routes/tests | Timeline field keys/record types | History/rollback | High | Timeline-specific mapping outside Timeline; define explicit port/contract before moving. |
| `internal/app/runtime.go` | Registers Timeline routes | App composition root | Public app assembly | Server binary | timeline service | Route inventory | Medium | Should depend on facade/service construction only. |
| `internal/app/recovery_probe.go` | Searched only; appears to query Timeline schema/workbook rows | App/recovery | Not inspected for S-00 | Recovery probe | workbook/timeline refs | Recovery readiness | Medium | Out of scope for characterization; inspect before changing recovery query/facade construction. |
| `contracts/view-schemas/cartulary.view.timeline.v2.json` | Timeline v2 view schema | Contract owner input | Authored contract input, not generated | generators/tests | none | Core 01 view schema | Very high | Do not edit for refactor unless owner contract change is explicitly required. |
| `contracts/openapi/cartulary.openapi.yaml` | Public HTTP schemas/routes | Contract owner input | Authored contract input | generators/tests | none | API compatibility | Very high | Refactor should not change it. |
| `contracts/ws/index.schema.json` | WS message schema including `record_changed` | Contract owner input | Authored contract input | generators/tests | none | Collaboration payload | High | Sparse patch behavior must remain. |
| `db/migrations/00004_*.sql`, `00037_*.sql`; `db/queries/timeline_phase3.sql` | Timeline tables/projection/sqlc query inputs | Storage/query owners | Authored SQL inputs | sqlc/store | Postgres | DB compatibility | High | No schema redesign in this plan. |
| `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**` | Generated contracts | Generated owners | Generated; no hand edits | Go/TS callers | generated from contracts | API/view schemas | High | Searched for references; never hand-edit. |
| `apps/web/src/workbook/timeline/**` and related frontend tests | UI callers and behavior tests | Web app | Frontend source | API/WS | generated TS/contracts | UX/workbook behavior | Medium | Searched only; inspect if a later route/DTO/WS behavior may drift. |

## 4. Timeline Module Contract Map

| Surface | Specific contract | Owner source | Repo evidence | Characterization coverage | Drift risk |
| --- | --- | --- | --- | --- | --- |
| HTTP create/query/patch | Stable `view_schema_id`, `record_id`, `field_key`, `base_row_version`, `client_txn_id`; common envelopes | Core 01 §3.3.5, Core 03 §15 | OpenAPI, `api.go`, Workbook routes/store | Covered by `TestPhase3_I_3_01_CreatePatchReplayAndRollback`, `TestPhase3_RouteEnvelopeMatrix_I_3_06`, `TestPhase6_GridWriteConcurrencyRoute_U_6_01`, `TestPhase6_SameFieldConflictHTTP_U_6_02` | High |
| Low-friction create | Zero-field create allowed; ten visible cells default `null`; first state `rough` | Core 03 lifecycle, Core 04 AC-191, view schema `inline_create` | `DecodeTimelineCreateRequest`, `CreateRow`, view schema | Covered by `TestSupportPhase3Unit_CreateRequestCoverage`, `TestPhase3_I_3_01_CreatePatchReplayAndRollback`, `TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild` | High |
| Visible text fields | `timeline_visible_text_v1`: nullable string, exact decoded text, empty string distinct, max 32768, no NUL/control except tab/LF/CR | Core 01 §18, Core 04 AC-445/447/448 | `validTimelineVisibleText`, v2 schema | Covered by new `TestPhase3_TimelineVisibleTextContract_U_3_12`, plus paste/import paths through shared Timeline decoder/plan tests | High |
| Row version / optimistic concurrency | Exact replay wins; stale base conflicts; non-overlapping stale patch can proceed; same-field conflicts explicit | Core 01 §3.3.5, Core 03 §3.2-3.3 | `PatchRow`, conflict token logic | Covered by `TestPhase3_PatchFieldLevelConcurrency_U_3_11`, `TestPhase3_PatchSameFieldConflictEnvelope_I_3_04`, `TestPhase6_TextCompareMergeDurability_U_6_03`, `TestPhase6_ConflictResolveDurability_U_6_06` | Very high |
| `client_txn_id` idempotency | Route-scoped keys for create, patch, paste, conflict resolve, review, supersede, imports, bulk; profile PUT has no idempotency | Core 01 mutation table | `authn` idempotency calls in store; OpenAPI | Covered by `TestPhase3_I_3_01_CreatePatchReplayAndRollback`, `TestSupportPhase3Integration_RouteIdempotencyIsActorScoped`, `TestPhase9_I_9_01_TimelineClipboardPastePersistsOrderedMutationsAndConflicts`, `TestPhase4_BindingMode_U_4_01`, `TestPhase3_TimelineTimeConversionProfile_I_3_08` | Very high |
| Storage/revision history | Mutations append change_set/mutations/revisions and preserve source state | Core 02 history/source model | `store.go`, migrations, revisions rollback | Covered by `TestPhase3_CreateAndPatchWriteHistory_U_3_09`, `TestPhase3_I_3_01_CreatePatchReplayAndRollback`, `TestPhase7_RollbackSelectorUnion_U_7_05` | Very high |
| Projection refresh | Timeline reads from `timeline_grid_projection`; mutations update projection before commit response/live event | Core 01/03 | `projectionInput`, `UpsertTimelineRowTx`, query SQL | Covered by `TestPhase3_ProjectionContract_U_3_08`, `TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild`, `TestPhase3_I_3_01_CreatePatchReplayAndRollback` | High |
| Mention/entity hooks | `mention_origin` creates `entity_mentions`, no implicit stub; auto-resolution only on eligible interactive path; imports preserve mentions | Core 02 §6, Core 03 §12 | mention action functions, `auto_resolution.go`, Imports apply path | Covered by `TestPhase4_BindingMode_U_4_01`, `TestPhase4_AutoResolutionEligibility_U_4_08`, `TestPhase4_AutoResolutionEligibility_I_4_08`, `TestPhase4_ManualTimelineConfidenceNull_I_4_09` | Very high |
| Collaboration/WS | `record_changed` has `row_version`, `client_txn_id`, sorted `changed_field_keys`, sparse affected view patch | Core 03 live update, Core 04 AC-368, WS schema | `publishRecordChange`, `contracts/ws/index.schema.json` | Covered by `TestPhase3_I_3_01_CreatePatchReplayAndRollback`, `TestPhase3_CanonicalIncidentWebSocket_I_3_05`, frontend socket model tests searched for consumer shape | High |
| Supersede/review | Reviewer/admin actions, row_version increment, history, replacement `supersedes` link direction, terminal superseded state | Core 03 §6, Core 04 AC-194/196/197/329-331 | `MarkReviewed`, `Supersede`, `ValidateSupersedeReplacement` | Covered by `TestPhase3_CaptureStateLifecycle_U_3_03`, `TestPhase3_ReviewedDemotionAndSupersedeTerminality_U_3_04`, `TestPhase3_SupersedeReplayAndRollbackCoupling_U_3_10`, `TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions` | Very high |
| Clipboard paste/bulk | Batch result, conflicts-only success, target validation before version/conflict, shared tabular ingest | Core 01 §2.1/§3.3.5, Core 03 §13 | paste decode/store, Workbook bulk mutation | Covered by `TestSupportPhase9_ClipboardPasteParsingMappingRawCaptureAndBinding`, `TestPhase9_I_9_01_TimelineClipboardPastePersistsOrderedMutationsAndConflicts`, `TestPhase9_I_9_01_ClipboardPasteAndBulkRejectCrossIncidentTargets`, `TestPhase9_I_9_01_BulkMutationsPersistOneVisibleBatch` | Very high |
| Time conversion profile | GET default disabled version 1; PUT admin-only, profile_version concurrency, generated paired time semantics | Core 01 REQ-01-611..613, Core 04 AC-449/451 | `routes.go`, `PutTimeConversionProfile`, conversion helper | Covered by new `TestPhase3_TimelineTimeConversionProfile_I_3_08` | High |
| Generated contracts | Generated outputs downstream of contracts only | AGENTS, generated artifact policy | policy JSON, generated refs search | No generated files edited; `make agent-finalize` required before final handoff | High |
| Harness accounting | Make targets and retained evidence rules | `docs/testing-harness-nlspec.md`, AGENTS | targeted doc read | Validation commands and run roots recorded in Section 11 | Medium |

## 5. Boundary And Coupling Scan

| Finding | Evidence | Risk | Classification | Proposed owner | Required action |
| --- | --- | --- | --- | --- | --- |
| Workbook imports Timeline DTOs/errors and now uses `timeline.Facade` for runtime calls | `workbook.Store` now has `timelineStore *timeline.Facade`; routes still map Timeline errors | DTO/error coupling remains but direct store coupling is reduced | should_fix | Timeline facade + Workbook adapter | DONE for facade-first runtime calls; later cleanup can narrow DTO/error mapping. |
| Timeline route handlers include auth/session/CSRF/role/WS transport logic | `routes.go` handles auth and publishes WS | Module boundary mixes platform with domain mutation | should_fix | Platform HTTP adapter calling Timeline service | Extract route-error/result mapping only after tests freeze. |
| Timeline store constructs/uses peer module stores directly | `Store` owns records/revisions/projections/links/entities/authn stores | Tight coupling and hard-to-test transactions | must_fix | Private ports with existing implementations | Define private interfaces matching current calls, then adapt. |
| Entity mention lifecycle logic crosses Timeline and Entities | `insertMentionActionsTx`, `entities.NewStore(nil)`, projection rebuild hooks | Risk of duplicate validation or divergent mention semantics | should_fix | Timeline mention port + Entities owner | Inspect Entities module before any move; add characterization. |
| Revisions has Timeline field-key mapping | `rollback_store.go` maps Timeline source/cell/link/tag fields | Rollback can drift from Timeline DTO/schema | should_fix | Revisions owner with Timeline mapping contract | Expose small Timeline field map or shared contract; do not move blindly. |
| Imports calls `DecodeTimelineCreateRequest` and imported-create facade | `imports/routes.go` apply path now calls `CreateImportedTimelineRow` | Import behavior could still drift if facade implementation changes | should_fix | Timeline import facade | DONE for explicit imported-create command; preserve no-interactive-auto-resolution in later internals. |
| Projection/query SQL is partly manual | `QueryRows`, filter/sort maps, sqlc query inputs | View schema drift can break sorting/filter/grouping | should_fix | Timeline query/store + view schema guard tests | Add guardrail tests against v2 schema before refactor. |
| Clipboard paste parser/header mapping lives in Timeline | `clipboard_paste.go` uses `tabularingest` plus Timeline header tables | Could diverge from shared ingest contract | intentional | Timeline ingest adapter | Keep dependency to shared tabular-ingest; prevent XLSX/OpenXML deps. |
| Generated artifacts are dependencies | generated Go/TS references found | Hand edits would violate repo policy | intentional | Contract/generator owners | Never edit generated roots; update owner inputs only if necessary. |
| Tests are phase-shaped | `phase3_*`, `phase4_*`, `phase9_*` | Behavior grouping is hard to reason about during refactor | should_fix | Test suites by behavior | Add new characterization grouped by behavior; do not rewrite existing tests first. |

## Characterization Audit Summary

Audited files:
- `internal/modules/timeline/phase3_decoder_test.go`, `phase3_support_test.go`, `phase3_store_test.go`, `phase3_projection_contract_test.go`, `phase3_integration_test.go`
- `internal/modules/timeline/phase4_unit_test.go`, `phase4_request_test.go`, `phase4_integration_test.go`
- `internal/modules/timeline/phase9_clipboard_paste_test.go`
- `internal/modules/workbook/phase6_conflict_route_test.go`, `phase9_clipboard_paste_integration_test.go`, `phase9_clipboard_paste_unit_test.go`
- `internal/modules/revisions/phase7_rollback_test.go`, `phase7_integration_test.go`, `phase7_history_test.go`
- Core owner text for time conversion in `docs/spec/01_architecture_storage_and_view_contracts.md` and `docs/spec/04_security_deployment_and_conformance.md`

Existing coverage accepted:
- Zero-field rough capture, create/replay/projection/history, same-field and stale conflict handling, conflict resolution, paste partial conflict behavior, cross-incident target rejection, imported mention no-auto-resolution, interactive auto-resolution, review/supersede lifecycle, projection rebuild, WS record changes, and rollback families.

New tests added:
- `TestPhase3_TimelineVisibleTextContract_U_3_12` covers nullable, empty-string, exact whitespace, source-like/formula/HTML/Markdown text preservation, max length, NUL/C0/C1 controls, and oversize rejection.
- `TestPhase3_TimelineTimeConversionProfile_I_3_08` covers default GET, admin-only PUT, disabled profile with null offset, stale `base_profile_version`, fixed-offset generated UTC, user-paired preservation, and unparseable text preservation.

Behaviors intentionally not preserved:
- None identified in this audit. Future refactor slices may explicitly drop accidental behavior only after owner-source review and replacement tests.

Remaining follow-up:
- Add a backend import-boundary/static guard before S-07 if a suitable existing pattern is found.
- Inspect frontend Timeline files before any frontend, generated TypeScript, or browser-facing route/WS change.

## 6. Characterization-Test Plan

| Behavior | Existing test/evidence | Missing evidence | Required characterization test or blocker | Command | Status |
| --- | --- | --- | --- | --- | --- |
| Zero-field and rough capture create | `TestSupportPhase3Unit_CreateRequestCoverage`; `TestPhase3_I_3_01_CreatePatchReplayAndRollback`; `TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild` | None for S-00 | Keep current tests; add future focused route test only if route envelope is split again | `make backend-unit`; `make backend-integration`; `make test-fast` | DONE: covered |
| Visible text validation | New `TestPhase3_TimelineVisibleTextContract_U_3_12`; `TestPhase3_RoughUncertainCapturePreservation_I_3_07`; paste/import call shared Timeline decode/planning paths | None for decoder contract; browser rendering remains separate UI evidence | Keep decoder test; run browser checks only for frontend rendering edits | `make backend-unit`; `make test-fast` | DONE: covered |
| Patch concurrency and same-field conflict | `TestPhase3_PatchFieldLevelConcurrency_U_3_11`; `TestPhase3_PatchSameFieldConflictEnvelope_I_3_04`; `TestPhase6_GridWriteConcurrencyRoute_U_6_01`; `TestPhase6_TextCompareMergeDurability_U_6_03`; `TestPhase6_ConflictResolveDurability_U_6_06` | None for S-00 | Keep route/store conflict tests before moving mutation internals | `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Paste batch conflicts | `TestSupportPhase9_ClipboardPasteParsingMappingRawCaptureAndBinding`; `TestPhase9_I_9_01_TimelineClipboardPastePersistsOrderedMutationsAndConflicts`; `TestPhase9_I_9_01_ClipboardPasteAndBulkRejectCrossIncidentTargets`; `TestPhase9_I_9_01_BulkMutationsPersistOneVisibleBatch` | None for S-00 | Keep Workbook phase9 integration tests as paste facade contract while paste internals move | `make backend-unit`; `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Imported Timeline create | `TestPhase4_BindingMode_U_4_01`; Imports apply path inspected and now calls `CreateImportedTimelineRow` facade | None for S-00 | Add route/job-level import apply test only if later Import/Timeline port changes alter apply orchestration | `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Auto-resolution interactive create | `TestPhase4_AutoResolutionEligibility_U_4_08`; `TestPhase4_AutoResolutionEligibility_I_4_08`; `TestPhase4_ManualTimelineConfidenceNull_U_4_09`; `TestPhase4_ManualTimelineConfidenceNull_I_4_09` | None for S-00 | Keep exact-match/suppressor/competing-alias/rollback/rebuild cases before entity port extraction | `make backend-unit`; `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Projection refresh and WS sparse patch | `TestPhase3_ProjectionContract_U_3_08`; `TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild`; `TestPhase3_I_3_01_CreatePatchReplayAndRollback`; `TestPhase3_CanonicalIncidentWebSocket_I_3_05` | Browser socket consumer evidence not rerun here | Run browser stateful/webserver-backed before merging route/WS/frontend-facing slices | `make backend-unit`; `make backend-integration`; `make test-fast` | DONE: backend covered |
| Review/supersede lifecycle | `TestPhase3_CaptureStateLifecycle_U_3_03`; `TestPhase3_ReviewedDemotionAndSupersedeTerminality_U_3_04`; `TestPhase3_SupersedeReplayAndRollbackCoupling_U_3_10`; `TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions`; `TestPhase7_RollbackSelectorUnion_U_7_05` | None for S-00 | Keep lifecycle and rollback cases before splitting action internals | `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Time conversion profile | New `TestPhase3_TimelineTimeConversionProfile_I_3_08`; Core 01 REQ-01-611..613 and Core 04 AC-449/451 audited | None for S-00 | Add patch/paste time-conversion cases if future work changes patch or paste conversion application paths | `make backend-integration`; `make test-fast` | DONE: covered |
| Revisions rollback mapping | `TestPhase7_RollbackSelectorUnion_U_7_05` subcases for timeline record patch, entity mention patch/dismiss/restore, supersede link rollback, and record tag unsupported; attached-evidence helpers covered in same file | Exact changed-field-key payload assertions remain a useful future guardrail before changing rollback event emission | Add changed-field-key assertions in S-05 if rollback port extraction changes live patch materialization | `make backend-store`; `make backend-integration`; `make test-fast` | DONE: sufficient for S-00; guardrail recommended |

## 7. Refactor Slice Plan

| Slice | Depends on | Change | Files likely touched | Behavior expected unchanged | Validation | Rollback note | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 Characterization freeze | None | Audited high-risk contracts and added missing visible-text/time-conversion tests | Timeline/workbook/imports/revisions tests | External behavior preserved; future moves must preserve characterized behavior only where owner/future value remains clear | Task guides, `make backend-unit`, `make backend-store`, `make backend-integration`, `make agent-finalize`, `make test-fast` | Revert only the two characterization tests if proven incorrect | DONE |
| S-01 Define facade contract | S-00 | Add facade methods that wrap current `Store` methods and return current result/errors | `internal/modules/timeline/facade.go` | DTOs, errors, status codes, payloads | `make backend-unit`, `make backend-store`, `make backend-integration`, `make test-fast` | Revert facade wrapper | DONE |
| S-02 Route adapter through facade | S-01 | Make Timeline routes call facade, keep auth/session/WS/error mapping unchanged | `internal/modules/timeline/routes.go` | HTTP behavior and WS publish | `make backend-integration`, `make test-fast` | Revert handler call indirection | DONE |
| S-03 Convert Workbook callers | S-01 | Change Workbook store/routes to depend on facade instead of `*timeline.Store` internals | `internal/modules/workbook/store.go`, `internal/modules/workbook/routes.go` | Public workbook routes unchanged | `make backend-store`, `make backend-integration`, `make test-fast` | Revert caller adapter | DONE for runtime store calls; DTO/error imports remain. |
| S-04 Add import-specific command | S-01 | Expose explicit imported-create facade preserving current no-interactive-auto-resolution path | `internal/modules/timeline/facade.go`, `internal/modules/imports/routes.go` | Import apply results and mention provenance unchanged | `make backend-integration`, `make test-fast` | Revert to direct store call | DONE |
| S-05 Private ports for peer modules | S-00..S-04 | Introduce private interfaces for revisions/projections/records/links/entities/idempotency calls using existing implementations | Timeline internals | Transactions, row_version, revisions, projections unchanged | Timeline + revisions/entity tests | Revert port extraction | TODO - unblocked for characterized behaviors |
| S-06 Split private implementation files | S-05 | Move internal code by behavior within Timeline package: validation, mutations, mentions, projections, paste | Timeline only | No package/import behavior change | `make test-fast` | Revert file split | TODO |
| S-07 Boundary guardrails | S-03..S-06 | Add import-boundary/static tests preventing platform/generated/peer internals from leaking back | Tests/tools if existing pattern found | Runtime unchanged | `make frontend-import-boundary-check` if relevant; discover backend boundary target before implementing | Revert guardrail only | TODO |
| S-08 Cleanup exported internals | S-07 | Narrow exported store/hooks/errors only where no caller remains; preserve public DTOs needed by callers | Timeline + callers | API/contracts unchanged | Full targeted + `make check` if broad | Restore exported symbols | TODO |

## 8. Backend Facade Target Shape

Target shape to build; do not claim this exists today:
- Public service/facade: `CreateTimelineRow`, `CreateImportedTimelineRow`, `QueryTimelineRows`, `PatchTimelineRow`, `ResolveTimelineConflict`, `ClipboardPaste`, `MarkReviewed`, `Supersede`, `Get/PutTimeConversionProfile`.
- Command/query DTOs: retain current JSON contract structs or compatibility wrappers; keep `TimelineViewSchemaID` and route keys stable.
- Validation/defaulting: module-owned validation for Timeline v2 visible text, collection actions, capture-state rules, reason normalization, zero-field create, and imported-create differences.
- Errors: closed Timeline error family plus route-error mapping boundary; callers must not inspect persistence details.
- Private store interface: transaction/persistence methods behind facade; persistence structs unexported where possible.
- Private persistence implementation: DB logic split by behavior but still one transaction owner for each mutation.
- Stable hooks: explicit private ports for projections, revisions, records, links, entity mentions/auto-resolution, idempotency, and collaboration event materialization.
- Tests: add new behavior-focused characterization tests while retaining phase tests until coverage is safely redundant.

## 9. Workflow Dependency Map

| Workflow | Status | Dependencies | Outputs | Exit criteria |
| --- | --- | --- | --- | --- |
| WF-00 Session/source bootstrap | DONE | AGENTS, framework, git/date | Header, authority posture, source limits | Branch/commit/dirty state recorded |
| WF-01 Current-state repository scan | DONE | WF-00 | Timeline and adjacent inventory | All claims tied to inspected/searched files |
| WF-02 Module ownership inventory | DONE | WF-01 | Current owner/coupling map | In-scope/non-scope and target owners recorded |
| WF-03 Public contract freeze | DONE for S-00 | WF-01, Core docs, characterization audit | Contract map plus exact test evidence | Section 4 and Section 6 name owner source and test evidence for high-risk behavior |
| WF-04 Refactor slice selection | DONE for facade slice | WF-03 | S-01 through S-04 selected and executed | Facade slice has passing backend validation and rollback note |
| WF-05 Characterization test plan | DONE | WF-03 | Behavior/test matrix plus two added characterization tests | No S-00 characterization blockers remain |
| WF-06 Boundary guardrail plan | TODO | WF-02, WF-05 | Guardrail tests/import rules | Existing boundary target discovered |
| WF-07 Backend module facade plan | DONE | WF-02 | `timeline.Facade` wrapping existing `Store` | Facade compiles and targeted/backend aggregate checks passed |
| WF-09 Execution checkpoint plan | DONE for facade slice | WF-04 | Per-slice checkpoint/handoff | S-01 through S-04 independently reviewable |
| WF-10 Validation/harness accounting plan | DONE | AGENTS, harness docs | Validation table with run roots | Commands named and current run roots recorded |
| WF-11 Documentation/generated-artifact plan | DONE for planning | Generated policy | No-hand-edit rule | Generated roots excluded from manual edits |
| WF-12 Cleanup/anti-drift plan | TODO | WF-06..S-07 | Export cleanup and drift checks | No caller depends on removed internals |
| WF-13 Handoff/next-slice bootstrap | DONE for characterization slice | All planning WFs | Handoff record below | Next agent can restart at S-05 with Section 6 evidence in hand |

## 10. Top-Level Tracker

| ID | Item | Status |
| --- | --- | --- |
| T-001 | Define target module and scope | DONE |
| T-002 | Inspect current repo state | DONE, with source limits |
| T-003 | Map owner contracts | DONE, exact evidence recorded in Sections 4 and 6 |
| T-004 | Freeze characterization evidence | DONE |
| T-005 | Plan boundary guardrails | TODO |
| T-006 | Plan behavior-preserving moves | DONE for S-01 through S-04; S-05 through S-08 remain TODO |
| T-007 | Plan validation loop | DONE; targeted backend checks and `make test-fast` passed |
| T-008 | Update docs/contracts if required | DEFERRED; refactor should not require contract changes |
| T-009 | Execute or hand off | DONE for S-00 through S-04; next handoff starts at S-05 |

## 11. Validation Plan

Validation run record for implemented facade plus characterization slice:

| Command | Result | Run root / artifact |
| --- | --- | --- |
| `make task-guide ROLE=feature-dev PHASE=phase3` | PASS; guidance emitted | Latest relevant artifacts listed by command |
| `make task-guide ROLE=feature-dev PHASE=phase4` | PASS; guidance emitted | Latest relevant artifacts listed by command |
| `make task-guide ROLE=feature-dev PHASE=phase7` | PASS; guidance emitted | Latest relevant artifacts listed by command |
| `make task-guide ROLE=feature-dev PHASE=phase9` | PASS; guidance emitted | Latest relevant artifacts listed by command |
| `make backend-unit` | PASS; 87 tests, 0 failed | `.cartulary/test-results/20260629T011453Z-p3728208` |
| `make backend-store` | PASS; 128 tests, 0 failed | `.cartulary/test-results/20260629T011453Z-p3728207` |
| `make backend-integration` | PASS; 160 tests, 0 failed | `.cartulary/test-results/20260629T011453Z-p3728228` |
| `make agent-finalize` | PASS; generated unchanged, retained `RESULTS_DIR` unset | `.cartulary/test-results/20260629T012005Z-p3743171` |
| `make test-fast` | PASS; 963 tests, 0 failed | `.cartulary/test-results/20260629T012009Z-p3743365` |
| `git diff --check` | PASS | no artifact |

Browser E2E was skipped for this slice because public routes, generated contracts, frontend code, schema inputs, and WS payload construction were unchanged. Run browser targets before merging any later route/WS/frontend-facing slice.

| Target | Purpose | Required? | Expected artifact | Failure handling |
| --- | --- | --- | --- | --- |
| `make task-guide ROLE=feature-dev PHASE=phase3` | Discover narrow Timeline create/patch/lifecycle validation | Required before S-00/S-01 | Task-guide output | Use recommended narrow target; record if absent |
| `make task-guide ROLE=feature-dev PHASE=phase4` | Discover mention/entity/evidence validation | Required before mention/port work | Task-guide output | Block mention refactor if no suitable target/test |
| `make task-guide ROLE=feature-dev PHASE=phase7` | Discover rollback/history validation | Required before revisions seam changes | Task-guide output | Block revisions seam changes if no coverage |
| `make task-guide ROLE=feature-dev PHASE=phase9` | Discover paste validation | Required before paste refactor | Task-guide output | Block paste refactor if no coverage |
| `make test-fast` | Cheapest broad backend regression named in AGENTS | Required after each code slice | Command output / run root if emitted | Report failing package/target and relation to slice |
| `make generated-artifact-policy-check` | Ensure generated roots not hand-edited | Required if contracts/generated refs touched; optional otherwise | Command output | Revert generated hand edits; update owner inputs only |
| `make json-shape-check` | Contract JSON/schema shape guard | Required if view schema/contracts touched | Command output | Treat failure as contract drift |
| `make migration-drift` | SQL migration/query drift guard | Required only if SQL inputs change | Command output | Do not continue with schema drift |
| `make frontend-import-boundary-check` | Frontend import boundary guard | Required only if frontend/generated TS imports change | Command output | Fix caller boundary or revert |
| `make frontend-unit` | Frontend caller compatibility | Required if public API/generated TS behavior changes | Command output | Treat failures as public contract risk |
| `make browser-e2e-stateful` or `make browser-e2e-webserver-backed` | Stateful workbook/WS/projection behavior | Required before merging route/WS-affecting slices | Harness result root | Report run root and failing summary |
| `make agent-finalize` | End-of-run repo finalization per AGENTS | Required before broad final verification | Command output; `RESULTS_DIR` if retained | If `RESULTS_DIR` unset, state retained-run maintenance skipped |
| `make check` | Broad confidence for shared-boundary changes | Required only for high-blast-radius slices | Command output / run root | Report target and summary artifact |

## 12. Handoff Planner

Workstream notes:
- Scope and evidence: treat this artifact as planning evidence plus S-00 characterization record; re-open any out-of-scope or searched-only file before editing. Current repo state, not the framework, owns implementation facts.
- Contracts and docs: Core 00-04 own behavior; `docs/domain.md` owns vocabulary; generated contracts are downstream. Do not hand-edit generated roots.
- Backend modules: facade-first. Convert callers to facade before splitting store internals or narrowing exports.
- Tests and harness: S-00 characterization is complete for known high-risk Timeline behavior. Prefer behavior tests over new phase-shaped tests, but do not rewrite existing phase tests just for naming.
- Generated artifacts: if a public contract change becomes unavoidable, update owner inputs and run Make generation/drift targets; otherwise no contract/generation changes.
- Risks and blockers: row_version/idempotency, paste conflicts, imported mention behavior, projection/WS sparse patching, revisions rollback, and time-conversion profile are now characterized for S-00; add narrower guardrails before changing their internals.

Handoff record template:

| Date/time | Branch/commit | Target module/seam | Current workflow | Completed workflows | Changed files | Commands run | Passing validation | Failing validation | Decisions made | Open questions | Blockers | Next recommended workflow | Safe restart command |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-06-28 21:02:43 EDT -0400` | `main` / `8beaff1f4919fc3e0c7840dd43ce86ac421db6f4` | `internal/modules/timeline` / facade-first boundary | WF-09 execution checkpoint | WF-00, WF-01, WF-02, WF-04 for facade slice, WF-07, WF-09, WF-10, WF-13 | `internal/modules/timeline/facade.go`, `internal/modules/timeline/routes.go`, `internal/modules/workbook/store.go`, `internal/modules/workbook/routes.go`, `internal/modules/imports/routes.go`, `temp/current.md` | task guides for phases 3/4/7/9; `make backend-unit`; `make backend-store`; `make backend-integration`; `make agent-finalize`; `make test-fast`; `git diff --check` | backend-unit, backend-store, backend-integration, agent-finalize, test-fast, diff check | None | Added thin `timeline.Facade`; routed Timeline HTTP, Workbook runtime calls, and Imports imported-create through it; left `Store` internals and lower-level tests unchanged | Should S-05 private ports start with revisions/projections or entity mention port first? | Characterization blocker later cleared in the `2026-06-28 21:15:58 EDT -0400` row | S-05 private ports | `make task-guide ROLE=feature-dev PHASE=phase3` |
| `2026-06-28 21:15:58 EDT -0400` | `main` / `8beaff1f4919fc3e0c7840dd43ce86ac421db6f4` | `internal/modules/timeline` / characterization freeze | WF-13 handoff bootstrap | WF-03, WF-05, WF-10, WF-13 for S-00 | `docs/handoffs/Timeline-Module-Refactoring-Tracker.md`, `internal/modules/timeline/phase3_decoder_test.go`, `internal/modules/timeline/phase3_integration_test.go` | `make task-guide ROLE=feature-dev PHASE=phase3`; `make task-guide ROLE=feature-dev PHASE=phase4`; `make task-guide ROLE=feature-dev PHASE=phase7`; `make task-guide ROLE=feature-dev PHASE=phase9`; `make backend-unit`; `make backend-store`; `make backend-integration`; `make agent-finalize`; `make test-fast`; `git diff --check` | backend-unit, backend-store, backend-integration, agent-finalize, test-fast, diff check | None | Added visible-text and time-conversion characterization, audited and cited existing coverage, set S-00/T-004 DONE | Exact changed-field-key rollback assertions are recommended before rollback event-emission refactors | None for S-05 start; add per-port guardrails before each extraction | S-05 private ports, starting with projection/revisions/idempotency ports | `make task-guide ROLE=feature-dev PHASE=phase3` |

## 13. Binary Acceptance Criteria

| ID | Acceptance criterion |
| --- | --- |
| RF-AC-001 | Inventory is complete enough that every edited file in `internal/modules/timeline` or adjacent callers was directly inspected before modification. |
| RF-AC-002 | No external route path, method, request schema, response envelope, error code, status code, or OpenAPI contract changes unless an owner contract change is explicitly approved. |
| RF-AC-003 | Timeline create still supports low-friction rough capture, including zero-field create when addressed to `cartulary.view.timeline.v2`, with initial `capture_state='rough'`. |
| RF-AC-004 | Timeline v2 visible fields preserve `timeline_visible_text_v1` semantics, including exact string preservation, nullable cells, empty-string distinction, and control/length validation. |
| RF-AC-005 | `row_version`, `base_row_version`, same-field conflict, conflict-token, and stale replay behavior remain byte-for-byte compatible at public surfaces. |
| RF-AC-006 | `client_txn_id` route-scoped idempotency remains unchanged for create, patch, paste, conflict resolve, mark-reviewed, supersede, imports, and bulk paths; time-conversion PUT remains non-idempotent optimistic concurrency. |
| RF-AC-007 | Mention extraction and auto-resolution preserve `mention_origin`: no implicit stub creation, raw mention preservation, imported-create no interactive auto-resolution, and unresolved fallback. |
| RF-AC-008 | Projection refresh and revision history remain in the same transaction boundary as successful mutations, with no misleading projection row or partial history on failure. |
| RF-AC-009 | Collaboration effects preserve `record_changed` payload shape, authoritative `row_version`, canonical changed-field keys, affected view identity, and sparse patch semantics. |
| RF-AC-010 | Generated files under declared generated roots are not hand-edited; any generated drift is produced only from owner inputs and Make generation targets. |
| RF-AC-011 | Review/supersede lifecycle remains reviewer/admin-gated, idempotent, row-versioned, history-backed, and terminal for ordinary edits after `superseded`. |
| RF-AC-012 | Clipboard paste and bulk mutations preserve target validation ordering, conflicts-only batch success shape, shared tabular-ingest mapping, and change_set grouping. |
| RF-AC-013 | Time-conversion profile default/read/update semantics, profile_version conflict behavior, and generated pair-state effects remain unchanged. |
| RF-AC-014 | Revisions rollback continues to restore Timeline source fields, tags, evidence, mentions, and supersede links with correct changed field keys. |
| RF-AC-015 | No new phase-shaped runtime dependency, route redesign, schema redesign, UI redesign, or phase-accounting rewrite is introduced by the refactor. |
