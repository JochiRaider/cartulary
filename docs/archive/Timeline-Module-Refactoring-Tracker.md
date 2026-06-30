# Timeline Module Refactoring Tracker and Handoff Planner

## 1. Session Header

| Field | Value |
| --- | --- |
| Target directory | `internal/modules/timeline` |
| Branch / commit | `main` / `8beaff1f4919fc3e0c7840dd43ce86ac421db6f4` |
| Dirty tree | Dirty at characterization update time: added `docs/handoffs/Timeline-Module-Refactoring-Tracker.md`, added `internal/modules/timeline/facade.go`, modified `internal/modules/imports/routes.go`, `internal/modules/timeline/routes.go`, `internal/modules/timeline/phase3_decoder_test.go`, `internal/modules/timeline/phase3_integration_test.go`, `internal/modules/workbook/routes.go`, `internal/modules/workbook/store.go` |
| Date/time | Initial planning: `2026-06-28 20:48:32 EDT -0400`; facade update: `2026-06-28 21:02:43 EDT -0400`; characterization update: `2026-06-28 21:15:58 EDT -0400`; remediation register update: `2026-06-28 21:31:50 EDT -0400`; pre-port guardrails: `2026-06-28 21:38:26 EDT -0400`; private ports: `2026-06-28 21:46:49 EDT -0400`; internal file split: `2026-06-28 21:56:45 EDT -0400`; boundary guardrails: `2026-06-28 21:59:38 EDT -0400` |
| Agent | Codex |
| Mode | Implementation update after planning; no generated-file edits, no route/schema/UI redesign, no test rewrites |
| Framework | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Source limits | `AGENTS.md` and framework read. `internal/modules/timeline` implementation files inspected for facade scope. Characterization audit directly searched and targeted-read Timeline phase3/phase4/phase9 tests, Workbook phase6/phase9 tests, Revisions phase7 rollback/history tests, Core 01/04 time-conversion requirements, OpenAPI/view schema/WS/migration/sqlc references, and adjacent Workbook/Imports/Revisions seams. Frontend and generated outputs were searched only; inspect them before frontend or contract changes. |

## Summary

This tracker plans a behavior-preserving refactor of `internal/modules/timeline` toward a cleaner backend module boundary. The Timeline module’s target responsibility is low-friction timeline capture and mutation: rough capture, Timeline v2 DTO validation/defaulting, row-version/idempotent mutation handling, mention/entity action hooks, projection/revision/collaboration effects, and stable public route-facing results. Current repo evidence shows much of that behavior lives in `Store`, while route/auth/WebSocket handling and caller-specific error mapping are spread across Timeline, Workbook, Imports, and Revisions seams.

No owner contradiction was found in inspected sources. If later inspection finds conflict between Core 00-04, adopted NLSpecs, generated contracts, or code, mark `BLOCKED: owner contradiction`.

Implementation update: completed the facade-first runtime slice. `timeline.Facade` now wraps existing `Store` behavior, Timeline HTTP handlers use the facade, Workbook holds a Timeline facade for Timeline query/create/patch/conflict paths, and Imports uses an explicit imported-create facade method. No persistence internals, contracts, migrations, generated files, tests, route paths, envelopes, or WebSocket payload construction were changed.

Remediation update: this artifact is the controlling plan for S-05 through final handoff. Core 00-04 behavior, OpenAPI, view schemas, WebSocket payloads, SQL contracts, and generated files remain unchanged unless an owner contradiction is found. If later implementation uncovers such a contradiction, stop the slice and mark this tracker `BLOCKED: owner contradiction` before editing owner specs or contracts.

Pre-port guardrail update: added query/schema and rollback mapping guard tests before S-05 ports. The guard found one non-owner-backed implementation allowance: Timeline SQL accepted sorting by `timeline.evidence_count` even though Core 01 and `cartulary.view.timeline.v2` do not advertise it in `sort_fields`. The runtime mapping was removed to align implementation with the owner schema; no Core/spec/generated files were edited.

Private ports update: S-05 is complete. `Store` now depends on private Timeline ports for idempotency, record versioning, revisions, projections, links, and mention lifecycle/resolution. Production adapters in `ports.go` wrap the existing peer stores/functions and keep calls inside the existing transaction owner. Direct concrete peer-store construction in Timeline mutation paths was removed; a temporary `entities` error import remains in `store.go` to preserve current Workbook error mapping until WS-5 adapter narrowing.

Internal file split update: S-06 is complete. Private Timeline behavior was split into `time_conversion_store.go`, `query_projection_store.go`, `lifecycle_store.go`, and `mentions_collections_store.go` without public symbol changes. The first `make test-fast` run failed during object-store testcontainer readiness in unrelated Phase 10 operator fixture setup; a rerun passed.

Boundary guardrail update: S-07 is complete. `boundary_guard_test.go` runs under the existing `make backend-unit` path and fails closed for unapproved production sibling-module imports, route-only adapter imports outside `routes.go`, concrete peer-store imports outside `ports.go`, `tabularingest` outside paste decoding, and XLSX/OpenXML parser imports in Timeline production files. No generated task-surface files were edited.

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
| Visible text fields | `timeline_visible_text_v1`: nullable string, exact decoded text, empty string distinct, max 32768, no NUL/control except tab/LF/CR | Core 01 §18, Core 04 AC-445/447/448 | `validTimelineVisibleText`, v2 schema | Covered by support guard `TestSupportPhase3Unit_TimelineVisibleTextContract`, plus paste/import paths through shared Timeline decoder/plan tests | High |
| Row version / optimistic concurrency | Exact replay wins; stale base conflicts; non-overlapping stale patch can proceed; same-field conflicts explicit | Core 01 §3.3.5, Core 03 §3.2-3.3 | `PatchRow`, conflict token logic | Covered by `TestPhase3_PatchFieldLevelConcurrency_U_3_11`, `TestPhase3_PatchSameFieldConflictEnvelope_I_3_04`, `TestPhase6_TextCompareMergeDurability_U_6_03`, `TestPhase6_ConflictResolveDurability_U_6_06` | Very high |
| `client_txn_id` idempotency | Route-scoped keys for create, patch, paste, conflict resolve, review, supersede, imports, bulk; profile PUT has no idempotency | Core 01 mutation table | `authn` idempotency calls in store; OpenAPI | Covered by `TestPhase3_I_3_01_CreatePatchReplayAndRollback`, `TestSupportPhase3Integration_RouteIdempotencyIsActorScoped`, `TestPhase9_I_9_01_TimelineClipboardPastePersistsOrderedMutationsAndConflicts`, `TestPhase4_BindingMode_U_4_01`, `TestSupportPhase3Integration_TimelineTimeConversionProfile` | Very high |
| Storage/revision history | Mutations append change_set/mutations/revisions and preserve source state | Core 02 history/source model | `store.go`, migrations, revisions rollback | Covered by `TestPhase3_CreateAndPatchWriteHistory_U_3_09`, `TestPhase3_I_3_01_CreatePatchReplayAndRollback`, `TestPhase7_RollbackSelectorUnion_U_7_05` | Very high |
| Projection refresh | Timeline reads from `timeline_grid_projection`; mutations update projection before commit response/live event | Core 01/03 | `projectionInput`, `UpsertTimelineRowTx`, query SQL | Covered by `TestPhase3_ProjectionContract_U_3_08`, `TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild`, `TestPhase3_I_3_01_CreatePatchReplayAndRollback` | High |
| Mention/entity hooks | `mention_origin` creates `entity_mentions`, no implicit stub; auto-resolution only on eligible interactive path; imports preserve mentions | Core 02 §6, Core 03 §12 | mention action functions, `auto_resolution.go`, Imports apply path | Covered by `TestPhase4_BindingMode_U_4_01`, `TestPhase4_AutoResolutionEligibility_U_4_08`, `TestPhase4_AutoResolutionEligibility_I_4_08`, `TestPhase4_ManualTimelineConfidenceNull_I_4_09` | Very high |
| Collaboration/WS | `record_changed` has `row_version`, `client_txn_id`, sorted `changed_field_keys`, sparse affected view patch | Core 03 live update, Core 04 AC-368, WS schema | `publishRecordChange`, `contracts/ws/index.schema.json` | Covered by `TestPhase3_I_3_01_CreatePatchReplayAndRollback`, `TestPhase3_CanonicalIncidentWebSocket_I_3_05`, frontend socket model tests searched for consumer shape | High |
| Supersede/review | Reviewer/admin actions, row_version increment, history, replacement `supersedes` link direction, terminal superseded state | Core 03 §6, Core 04 AC-194/196/197/329-331 | `MarkReviewed`, `Supersede`, `ValidateSupersedeReplacement` | Covered by `TestPhase3_CaptureStateLifecycle_U_3_03`, `TestPhase3_ReviewedDemotionAndSupersedeTerminality_U_3_04`, `TestPhase3_SupersedeReplayAndRollbackCoupling_U_3_10`, `TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions` | Very high |
| Clipboard paste/bulk | Batch result, conflicts-only success, target validation before version/conflict, shared tabular ingest | Core 01 §2.1/§3.3.5, Core 03 §13 | paste decode/store, Workbook bulk mutation | Covered by `TestSupportPhase9_ClipboardPasteParsingMappingRawCaptureAndBinding`, `TestPhase9_I_9_01_TimelineClipboardPastePersistsOrderedMutationsAndConflicts`, `TestPhase9_I_9_01_ClipboardPasteAndBulkRejectCrossIncidentTargets`, `TestPhase9_I_9_01_BulkMutationsPersistOneVisibleBatch` | Very high |
| Time conversion profile | GET default disabled version 1; PUT admin-only, profile_version concurrency, generated paired time semantics | Core 01 REQ-01-611..613, Core 04 AC-449/451 | `routes.go`, `PutTimeConversionProfile`, conversion helper | Covered by support guard `TestSupportPhase3Integration_TimelineTimeConversionProfile` | High |
| Generated contracts | Generated outputs downstream of contracts only | AGENTS, generated artifact policy | policy JSON, generated refs search | No generated files edited; `make agent-finalize` required before final handoff | High |
| Harness accounting | Make targets and retained evidence rules | `docs/testing-harness-nlspec.md`, AGENTS | targeted doc read | Validation commands and run roots recorded in Section 11 | Medium |

## 5. Boundary And Coupling Scan

| Finding | Evidence | Risk | Classification | Proposed owner | Required action |
| --- | --- | --- | --- | --- | --- |
| Workbook imports Timeline DTOs/errors and now uses `timeline.Facade` for runtime calls | `workbook.Store` now has `timelineStore *timeline.Facade`; routes still map Timeline errors | DTO/error coupling remains but direct store coupling is reduced | should_fix | Timeline facade + Workbook adapter | DONE for facade-first runtime calls; later cleanup can narrow DTO/error mapping. |
| Timeline route handlers include auth/session/CSRF/role/WS transport logic | `routes.go` handles auth and publishes WS | Module boundary mixes platform with domain mutation | should_fix | Platform HTTP adapter calling Timeline service | Extract route-error/result mapping only after tests freeze. |
| Timeline store constructs/uses peer module stores directly | `Store` now owns private port interfaces; concrete records/revisions/projections/links/entities/authn adapters live in `ports.go` | Concrete store construction in mutation paths was tight coupling | must_fix | Private ports with existing implementations | DONE for S-05; preserve transaction order through adapters. |
| Entity mention lifecycle logic crosses Timeline and Entities | Timeline now calls a private mention port for lifecycle/resolution; insertion still owns Timeline mention observation and link materialization through private link port | Risk of duplicate validation or divergent mention semantics | should_fix | Timeline mention port + Entities owner | DONE for private port seam; WS-5 should narrow exported error coupling. |
| Revisions has Timeline field-key mapping | `rollback_store.go` maps Timeline source/cell/link/tag fields | Rollback can drift from Timeline DTO/schema | should_fix | Revisions owner with Timeline mapping contract | Expose small Timeline field map or shared contract; do not move blindly. |
| Imports calls `DecodeTimelineCreateRequest` and imported-create facade | `imports/routes.go` apply path now calls `CreateImportedTimelineRow` | Import behavior could still drift if facade implementation changes | should_fix | Timeline import facade | DONE for explicit imported-create command; preserve no-interactive-auto-resolution in later internals. |
| Projection/query SQL is partly manual | `QueryRows`, filter/sort maps, sqlc query inputs | View schema drift can break sorting/filter/grouping | should_fix | Timeline query/store + view schema guard tests | Add guardrail tests against v2 schema before refactor. |
| Clipboard paste parser/header mapping lives in Timeline | `clipboard_paste.go` uses `tabularingest` plus Timeline header tables | Could diverge from shared ingest contract | intentional | Timeline ingest adapter | Keep dependency to shared tabular-ingest; prevent XLSX/OpenXML deps. |
| Generated artifacts are dependencies | generated Go/TS references found | Hand edits would violate repo policy | intentional | Contract/generator owners | Never edit generated roots; update owner inputs only if necessary. |
| Tests are phase-shaped | `phase3_*`, `phase4_*`, `phase9_*` | Behavior grouping is hard to reason about during refactor | should_fix | Test suites by behavior | Add new characterization grouped by behavior; do not rewrite existing tests first. |
| Exported Timeline internals remain broad | `Store`, store hooks, and helper builders remain exported for tests and adjacent lower-level callers | Accidental external dependency on persistence internals can become a compatibility burden | should_fix | Timeline facade/API boundary | After ports and adapters land, unexport or narrow any production-unused internals and update tests to use stable seams where practical. |

## 5A. Remediation Gap Register

This register governs S-05 through final validation. It intentionally favors clean module boundaries and future phase growth over preserving accidental internals. Public behavior remains stable unless an owner document is deliberately changed.

| Gap | Areas | Remediation | Rationale and long-term benefit | Compatibility / migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- |
| Workbook still imports Timeline DTOs/errors | Implementation, tests | Add or retain a narrow Workbook-facing adapter over `timeline.Facade`; centralize Timeline error mapping outside ad hoc Workbook route branches; keep DTO use only where request decoding still needs the existing wire contract. | Reduces cross-module coupling while avoiding a premature second wire contract. Future Workbook changes can target a stable facade seam instead of Timeline store details. | Internal Go API churn only; no HTTP, DB, OpenAPI, or generated contract change. | Workbook stays sensitive to Timeline internal error and DTO churn. | Workbook phase6/phase9 tests and Timeline integration tests pass; `rg` shows Workbook runtime mutations go through `timeline.Facade`. |
| Timeline routes mix auth/session/role/WS adapter code with mutation logic | Implementation, tests | Keep `timeline/routes.go` as the transport adapter, but constrain it to auth, authorization, request decode, route error mapping, facade calls, and WS publishing; move reusable mapping into small adapter helpers. | Keeps transport concerns out of application logic without a risky route redesign. | No route path, status, envelope, CSRF, or WS payload change. | Business rules can leak back into handlers and make future transport/API expansion brittle. | Route envelope, auth, mark-reviewed, time-profile, and WS tests pass. |
| `Store` owns concrete peer stores | Implementation, tests | Introduce private Timeline ports for idempotency, records, revisions, projections, links, and entity mentions; production adapters wrap current stores/functions and preserve transaction boundaries. | Makes transaction collaborators explicit, easier to test, and ready for future module growth. | Internal only; no schema or data migration. | Tight coupling continues to block safe refactors and focused tests. | Timeline create/patch/paste/review/supersede, revisions, entity, and projection tests pass. |
| Entity mention lifecycle crosses Timeline and Entities | Implementation, tests | Extract a Timeline-owned mention port that delegates to Entities for resolution/upsert and Links for relationship materialization; preserve no-stub and import no-auto-resolution behavior. | Clarifies that Timeline observes mention text while Entities owns entity semantics. | Internal only. | Mention semantics can drift between modules or accidentally re-enable stub creation. | Phase4 auto-resolution/import mention tests and rollback mention tests pass. |
| Revisions hard-codes Timeline field mappings | Implementation, tests, docs | Add a small Timeline-owned or neutral internal contract for rollback source-field mapping and changed-field-key derivation; Revisions consumes that contract without importing the parent Timeline package. | Prevents rollback drift while avoiding Go import cycles and making field ownership explicit. | Internal only; no public contract change. | Rollback can restore the wrong fields or emit incomplete changed-field keys. | Changed-field-key assertions exist; phase7 rollback/history tests pass. |
| Projection/query SQL can drift from view schema | Tests, implementation | Add guard tests comparing Timeline query sort/filter/group support to `cartulary.view.timeline.v2`; keep SQL mapping private and explicit. | Catches schema/query divergence before runtime and documents supported query semantics near the owner contract. | No behavior change unless the guard reveals real drift. | Sorting, filtering, or grouping can break silently. | New guard test plus phase3 projection/query tests pass. |
| Clipboard paste parser/header mapping is Timeline-local | Tests, docs | Preserve Timeline header mapping and shared `imports/tabularingest` dependency; add a guard that Timeline paste does not import XLSX/OpenXML parser code. | Keeps paste cohesive while preventing parser duplication and import-ingest drift. | None. | Clipboard behavior can diverge from shared import ingest rules. | Phase9 paste tests and import-boundary guard pass. |
| Imports decodes Timeline create payload directly | Implementation, tests | Keep current decode until a higher-level import command exists; document it as a compatibility seam and route all persistence through `CreateImportedTimelineRow`. | Avoids moving coupling prematurely while protecting imported-create semantics. | Internal only. | Imports can re-enable interactive auto-resolution or bypass Timeline defaults. | Import apply and phase4 binding-mode tests pass. |
| Generated artifacts are dependencies | Docs, validation | Never hand-edit generated roots; if a contract change becomes unavoidable, update owner inputs and run Make generation/drift targets. | Preserves source-of-truth discipline and reproducible contracts. | Generated drift only through approved generators. | Hand edits create non-reproducible contracts and downstream type drift. | `make generated-artifact-policy-check`; `make agent-finalize`. |
| Tests are phase-shaped | Tests, docs | Add behavior-focused tests for new seams; do not rename or rewrite existing phase tests unless coverage is redundant and stable. | Improves maintainability without destabilizing evidence accounting. | Test-only. | Refactors remain hard to reason about and failures stay too broad. | New tests run under existing Make targets; tracker records coverage. |
| Exported internals and test hooks remain broad | Implementation, tests | After callers use `Facade` and private ports, unexport or narrow `Store`, hooks, and helper builders where no external production caller remains; preserve public DTOs/constants still needed by callers. | Shrinks module surface and future compatibility burden. | Internal Go API and test updates only. | Accidental external dependency on internals persists. | `rg` confirms no production internal callers; targeted backend checks and broad validation pass. |

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

New support tests added:
- `TestSupportPhase3Unit_TimelineVisibleTextContract` covers nullable, empty-string, exact whitespace, source-like/formula/HTML/Markdown text preservation, max length, NUL/C0/C1 controls, and oversize rejection without replacing workbook-owned `U-3-12`.
- `TestSupportPhase3Integration_TimelineTimeConversionProfile` covers default GET, admin-only PUT, disabled profile with null offset, stale `base_profile_version`, fixed-offset generated UTC, user-paired preservation, and unparseable text preservation.

Behaviors intentionally not preserved:
- None identified in this audit. Future refactor slices may explicitly drop accidental behavior only after owner-source review and replacement tests.

Remaining follow-up:
- Add a backend import-boundary/static guard before S-07 if a suitable existing pattern is found.
- Inspect frontend Timeline files before any frontend, generated TypeScript, or browser-facing route/WS change.

## 6. Characterization-Test Plan

| Behavior | Existing test/evidence | Missing evidence | Required characterization test or blocker | Command | Status |
| --- | --- | --- | --- | --- | --- |
| Zero-field and rough capture create | `TestSupportPhase3Unit_CreateRequestCoverage`; `TestPhase3_I_3_01_CreatePatchReplayAndRollback`; `TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild` | None for S-00 | Keep current tests; add future focused route test only if route envelope is split again | `make backend-unit`; `make backend-integration`; `make test-fast` | DONE: covered |
| Visible text validation | New support guard `TestSupportPhase3Unit_TimelineVisibleTextContract`; `TestPhase3_RoughUncertainCapturePreservation_I_3_07`; paste/import call shared Timeline decode/planning paths | None for decoder contract; browser rendering remains separate UI evidence | Keep decoder test; run browser checks only for frontend rendering edits | `make backend-unit`; `make test-fast` | DONE: covered |
| Patch concurrency and same-field conflict | `TestPhase3_PatchFieldLevelConcurrency_U_3_11`; `TestPhase3_PatchSameFieldConflictEnvelope_I_3_04`; `TestPhase6_GridWriteConcurrencyRoute_U_6_01`; `TestPhase6_TextCompareMergeDurability_U_6_03`; `TestPhase6_ConflictResolveDurability_U_6_06` | None for S-00 | Keep route/store conflict tests before moving mutation internals | `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Paste batch conflicts | `TestSupportPhase9_ClipboardPasteParsingMappingRawCaptureAndBinding`; `TestPhase9_I_9_01_TimelineClipboardPastePersistsOrderedMutationsAndConflicts`; `TestPhase9_I_9_01_ClipboardPasteAndBulkRejectCrossIncidentTargets`; `TestPhase9_I_9_01_BulkMutationsPersistOneVisibleBatch` | None for S-00 | Keep Workbook phase9 integration tests as paste facade contract while paste internals move | `make backend-unit`; `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Imported Timeline create | `TestPhase4_BindingMode_U_4_01`; Imports apply path inspected and now calls `CreateImportedTimelineRow` facade | None for S-00 | Add route/job-level import apply test only if later Import/Timeline port changes alter apply orchestration | `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Auto-resolution interactive create | `TestPhase4_AutoResolutionEligibility_U_4_08`; `TestPhase4_AutoResolutionEligibility_I_4_08`; `TestPhase4_ManualTimelineConfidenceNull_U_4_09`; `TestPhase4_ManualTimelineConfidenceNull_I_4_09` | None for S-00 | Keep exact-match/suppressor/competing-alias/rollback/rebuild cases before entity port extraction | `make backend-unit`; `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Projection refresh, query schema, and WS sparse patch | `TestPhase3_ProjectionContract_U_3_08`; new support guard `TestSupportPhase3Unit_TimelineQuerySchemaMappingGuard`; `TestPhase3_I_3_02_ProjectionQueryUsesDeterministicRebuild`; `TestPhase3_I_3_01_CreatePatchReplayAndRollback`; `TestPhase3_CanonicalIncidentWebSocket_I_3_05` | Browser socket consumer evidence not rerun here | Run browser stateful/webserver-backed before merging route/WS/frontend-facing slices | `make backend-unit`; `make backend-integration`; `make test-fast` | DONE: backend and query/schema guard covered |
| Review/supersede lifecycle | `TestPhase3_CaptureStateLifecycle_U_3_03`; `TestPhase3_ReviewedDemotionAndSupersedeTerminality_U_3_04`; `TestPhase3_SupersedeReplayAndRollbackCoupling_U_3_10`; `TestPhase3_I_3_03_AuthorizationLifecycleAndSupersedeTransitions`; `TestPhase7_RollbackSelectorUnion_U_7_05` | None for S-00 | Keep lifecycle and rollback cases before splitting action internals | `make backend-store`; `make backend-integration`; `make test-fast` | DONE: covered |
| Time conversion profile | New support guard `TestSupportPhase3Integration_TimelineTimeConversionProfile`; Core 01 REQ-01-611..613 and Core 04 AC-449/451 audited | None for S-00 | Add patch/paste time-conversion cases if future work changes patch or paste conversion application paths | `make backend-integration`; `make test-fast` | DONE: covered |
| Revisions rollback mapping | `TestPhase7_RollbackSelectorUnion_U_7_05` subcases for timeline record patch, entity mention patch/dismiss/restore, supersede link rollback, and record tag unsupported; attached-evidence helpers covered in same file; new rollback mapping guards cover Timeline changed-field keys, mention fallback keys, and Timeline source-field mapping | None for WS-1 | Keep guard tests in place before introducing a shared rollback mapping contract in later cleanup | `make backend-store`; `make backend-integration`; `make test-fast` | DONE: guardrail added |

## 7. Refactor Slice Plan

| Slice | Depends on | Change | Files likely touched | Behavior expected unchanged | Validation | Rollback note | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 Characterization freeze | None | Audited high-risk contracts and added missing visible-text/time-conversion tests | Timeline/workbook/imports/revisions tests | External behavior preserved; future moves must preserve characterized behavior only where owner/future value remains clear | Task guides, `make backend-unit`, `make backend-store`, `make backend-integration`, `make agent-finalize`, `make test-fast` | Revert only the two characterization tests if proven incorrect | DONE |
| S-01 Define facade contract | S-00 | Add facade methods that wrap current `Store` methods and return current result/errors | `internal/modules/timeline/facade.go` | DTOs, errors, status codes, payloads | `make backend-unit`, `make backend-store`, `make backend-integration`, `make test-fast` | Revert facade wrapper | DONE |
| S-02 Route adapter through facade | S-01 | Make Timeline routes call facade, keep auth/session/WS/error mapping unchanged | `internal/modules/timeline/routes.go` | HTTP behavior and WS publish | `make backend-integration`, `make test-fast` | Revert handler call indirection | DONE |
| S-03 Convert Workbook callers | S-01 | Change Workbook store/routes to depend on facade instead of `*timeline.Store` internals | `internal/modules/workbook/store.go`, `internal/modules/workbook/routes.go` | Public workbook routes unchanged | `make backend-store`, `make backend-integration`, `make test-fast` | Revert caller adapter | DONE for runtime store calls; DTO/error imports remain. |
| S-04 Add import-specific command | S-01 | Expose explicit imported-create facade preserving current no-interactive-auto-resolution path | `internal/modules/timeline/facade.go`, `internal/modules/imports/routes.go` | Import apply results and mention provenance unchanged | `make backend-integration`, `make test-fast` | Revert to direct store call | DONE |
| S-05 Private ports for peer modules | S-00..S-04, WS-1 | Introduce private interfaces for revisions/projections/records/links/entities/idempotency calls using existing implementations | Timeline internals plus `ports.go` | Transactions, row_version, revisions, projections unchanged | Timeline + revisions/entity tests | Revert port extraction | DONE |
| S-06 Split private implementation files | S-05 | Move internal code by behavior within Timeline package: validation, query/projection, create/patch/conflict, lifecycle actions, mentions/collections, paste, and time conversion | Timeline only | No package/import behavior change or public symbol changes in this slice | `make test-fast`; `git diff --check` | Revert file split | DONE |
| S-07 Boundary guardrails | S-03..S-06 | Add backend import-boundary/static tests at file-group level: route adapter imports only in route files, peer-module imports only in production adapter files, `tabularingest` only for paste | `internal/modules/timeline/boundary_guard_test.go` | Runtime unchanged | `make backend-unit` | Revert guardrail only | DONE |
| S-08 Cleanup exported internals | S-07 | Narrow exported `Store`, hooks, helper builders, Workbook/Imports error/DTO coupling only where no external production caller remains; preserve wire-compatible DTOs still needed for decoding | Timeline + callers | API/contracts unchanged | Targeted backend checks; `make check` if blast radius is high | Restore exported symbols | DONE with retained `Store`/`NewStore`, explicit test hook, and wire DTOs for active test/caller seams |

## 7A. Remediation Workstream Plan

| Workstream | Depends on | Required work | Key risk | Exit criteria | Status |
| --- | --- | --- | --- | --- | --- |
| WS-0 Spec and tracker cleanup | S-00..S-04 | Add this remediation/gap register; record that Core specs stay unchanged unless public behavior changes; expand S-05/S-07/S-08 notes with implementation sequencing. | Treating implementation-support docs as product spec. | Tracker updated; no Core edits unless owner conflict is found. | DONE |
| WS-1 Pre-port guardrails | WS-0 | Add behavior guard tests for Timeline query/schema mapping and rollback changed-field keys; confirm task-guide targets for phases 3/4/7/9. | Guard tests accidentally encode implementation quirks. | `make backend-unit`, `make backend-store`, and `make backend-integration` pass; tracker updated. | DONE |
| WS-2 S-05 private ports | WS-1 | Add private port interfaces and production adapters; replace direct peer-store construction in mutation paths; keep all port calls inside existing transactions. | Breaking idempotency, revision, projection, or mention transaction order. | Timeline, Workbook, Imports, Revisions, and Entities targeted tests pass; tracker updated. | DONE |
| WS-3 S-06 internal file split | WS-2 | Split Timeline internals by behavior without public symbol changes. | Mechanical move hides behavior drift. | `make test-fast` and `git diff --check` pass; tracker updated. | DONE |
| WS-4 S-07 boundary guardrails | WS-3 | Add backend import-boundary guardrails at file-group level. | Over-broad guardrail blocks intentional seams. | Boundary check runs under an existing Make-owned validation path or a documented new one; tracker updated. | DONE |
| WS-5 S-08 exported cleanup and adapter narrowing | WS-4 | Remove or unexport unused internals; narrow Workbook/Imports coupling through facade or adapter helpers where safe. | Removing symbols still used by tests or adjacent modules. | `rg` proves no production internal callers; targeted backend checks and `make check` if blast radius is high; tracker updated. | DONE |
| WS-6 Validation and handoff completion | WS-5 | Run final verification, `make agent-finalize`, generated/drift checks as applicable, and browser E2E only if route/WS/frontend behavior changed; complete final handoff row. | Skipping retained-run or generated-artifact accounting. | Tracker shows all slices DONE or explicitly DEFERRED/DROPPED with rationale, commands, artifacts, failures, and next safe restart command. | DONE |

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
| WF-06 Boundary guardrail plan | DONE for remediation scope | WF-02, WF-05, WS-0 | Guardrail tests/import rules | Guardrail scope recorded in Section 5A and WS-4 |
| WF-07 Backend module facade plan | DONE | WF-02 | `timeline.Facade` wrapping existing `Store` | Facade compiles and targeted/backend aggregate checks passed |
| WF-09 Execution checkpoint plan | DONE for facade slice | WF-04 | Per-slice checkpoint/handoff | S-01 through S-04 independently reviewable |
| WF-10 Validation/harness accounting plan | DONE | AGENTS, harness docs | Validation table with run roots | Commands named and current run roots recorded |
| WF-11 Documentation/generated-artifact plan | DONE for planning | Generated policy | No-hand-edit rule | Generated roots excluded from manual edits |
| WF-12 Cleanup/anti-drift plan | DONE for remediation scope | WF-06..S-07 | Export cleanup and drift checks | S-08 gated by caller audit and targeted validation |
| WF-13 Handoff/next-slice bootstrap | DONE for WS-0 | All planning WFs | Handoff record below | Next workstream starts at WS-1 pre-port guardrails |

## 10. Top-Level Tracker

| ID | Item | Status |
| --- | --- | --- |
| T-001 | Define target module and scope | DONE |
| T-002 | Inspect current repo state | DONE, with source limits |
| T-003 | Map owner contracts | DONE, exact evidence recorded in Sections 4 and 6 |
| T-004 | Freeze characterization evidence | DONE |
| T-005 | Plan boundary guardrails | DONE for remediation scope; implementation pending WS-4 |
| T-006 | Plan behavior-preserving moves | DONE for S-01 through S-08 sequencing; implementation pending WS-1 through WS-6 |
| T-007 | Plan validation loop | DONE; targeted backend checks and `make test-fast` passed |
| T-008 | Update docs/contracts if required | DONE for tracker-only remediation register; Core specs/contracts unchanged because no owner contradiction was found |
| T-009 | Execute or hand off | DONE; WS-6 final validation and handoff complete |

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

Remediation run record:

| Workstream | Command | Result | Run root / artifact |
| --- | --- | --- | --- |
| WS-0 | `git status --short --branch` | PASS; clean tree before tracker edit | no artifact |
| WS-0 | `date`; `git rev-parse --abbrev-ref HEAD`; `git rev-parse --short HEAD` | PASS; `2026-06-28 21:31:50 EDT`, `main`, `6662cfa` | no artifact |
| WS-0 | Tracker inspection with `sed`/`nl` | PASS; S-05 through S-08 pending and ready for remediation register | no artifact |
| WS-0 | Core spec edits | SKIPPED; no owner contradiction or public behavior change found | no artifact |
| WS-1 | `make task-guide ROLE=feature-dev PHASE=phase3` | PASS; phase3 backend-unit/store/integration remain relevant | Latest relevant artifacts listed by command |
| WS-1 | `make task-guide ROLE=feature-dev PHASE=phase4` | PASS; phase4 backend-unit/store/integration remain relevant | Latest relevant artifacts listed by command |
| WS-1 | `make task-guide ROLE=feature-dev PHASE=phase7` | PASS; phase7 backend-store/integration remain relevant | Latest relevant artifacts listed by command |
| WS-1 | `make task-guide ROLE=feature-dev PHASE=phase9` | PASS; phase9 backend-unit/store/integration remain relevant | Latest relevant artifacts listed by command |
| WS-1 | `make format` | PASS | `.cartulary/test-results/20260629T013702Z-p3779761` |
| WS-1 | `make backend-unit` | PASS; 87 tests, 0 failed | `.cartulary/test-results/20260629T013723Z-p3781348` |
| WS-1 | `make backend-store` | PASS; 128 tests, 0 failed | `.cartulary/test-results/20260629T013723Z-p3781366` |
| WS-1 | `make backend-integration` | PASS; 160 tests, 0 failed | `.cartulary/test-results/20260629T013723Z-p3781385` |
| WS-1 | `git diff --check` | PASS | no artifact |
| WS-2 | `make format` | PASS | `.cartulary/test-results/20260629T014440Z-p3800599` |
| WS-2 | `make backend-unit` | PASS; 87 tests, 0 failed | `.cartulary/test-results/20260629T014450Z-p3802031` |
| WS-2 | `make backend-store` | PASS; 128 tests, 0 failed | `.cartulary/test-results/20260629T014547Z-p3807727` |
| WS-2 | `make backend-integration` | PASS; 160 tests, 0 failed | `.cartulary/test-results/20260629T014547Z-p3807755` |
| WS-2 | `git diff --check` | PASS | no artifact |
| WS-3 | `make format` | PASS | `.cartulary/test-results/20260629T015028Z-p3824087` |
| WS-3 | `make backend-unit` | PASS; 87 tests, 0 failed | `.cartulary/test-results/20260629T015034Z-p3825514` |
| WS-3 | `git diff --check` | PASS | no artifact |
| WS-3 | `make test-fast` | FAIL; object-store testcontainer readiness timeout in Phase 10 operator fixture setup | `.cartulary/test-results/20260629T015116Z-p3831591` |
| WS-3 | `make test-fast` rerun | PASS; 963 tests, 0 failed | `.cartulary/test-results/20260629T015452Z-p3863491` |
| WS-4 | `make format` | PASS | `.cartulary/test-results/20260629T015905Z-p3894082` |
| WS-4 | `make backend-unit` | PASS; 87 tests, 0 failed | `.cartulary/test-results/20260629T015913Z-p3895519` |
| WS-4 | `git diff --check` | PASS | no artifact |
| WS-5 | `make format` | PASS | `.cartulary/test-results/20260629T020257Z-p3899338`; rerun after support-test rename `.cartulary/test-results/20260629T020904Z-p3931868` |
| WS-5 | `make phase-map-check` | PASS after support-test manifest repair | no output artifact |
| WS-5 | `make backend-unit` | PASS; 89 tests, 0 failed | `.cartulary/test-results/20260629T020916Z-p3933836` |
| WS-5 | `make backend-store` | PASS; 128 tests, 0 failed | `.cartulary/test-results/20260629T020916Z-p3933840` |
| WS-5 | `make backend-integration` | PASS; 160 tests, 0 failed | `.cartulary/test-results/20260629T020916Z-p3933889` |
| WS-5 | `make backend-integration-support` | PASS; quiet success | `.cartulary/test-results/20260629T021205Z-p3972274` |
| WS-5 | `make go-test-duration-baselines RESULTS_DIR=.cartulary/test-results/20260629T020916Z-p3933889 PRUNE_OBSERVED_PACKAGES=1` | FAIL; partial service-backed evidence cannot prune baselines | `.cartulary/test-results/20260629T021121Z-p3970820` |
| WS-5 | `make go-test-duration-baselines RESULTS_DIR=.cartulary/test-results/20260629T021205Z-p3972274` | PASS; added support integration test timing | `.cartulary/test-results/20260629T021246Z-p3974474` |
| WS-5 | `make go-test-duration-baseline-coverage` | PASS after baseline refresh | `.cartulary/test-results/20260629T021256Z-p3974567` |
| WS-5 | `make go-test-duration-baseline-drift` | PASS | `.cartulary/test-results/20260629T021256Z-p3974590` |
| WS-5 | `make phase-ledgers` | PASS; generated ledger refreshed from owner map | `.cartulary/test-results/20260629T021303Z-p3974728` |
| WS-5 | `make phase-schedules` | PASS; generated schedules refreshed from owner map/baseline | `.cartulary/test-results/20260629T021304Z-p3974751` |
| WS-5 | `make json-shape-check` | PASS after phase schedule refresh | `.cartulary/test-results/20260629T021315Z-p3975271` |
| WS-5 | `make phase-ledger-drift` | PASS | `.cartulary/test-results/20260629T021315Z-p3975273` |
| WS-5 | `make phase-schedule-drift` | PASS | `.cartulary/test-results/20260629T021315Z-p3975310` |
| WS-5 | `make check` | First FAIL on phase3 map/baseline/schedule accounting, then PASS; 969 tests, 0 failed | fail roots `.cartulary/test-results/20260629T020418Z-p3917348`, `.cartulary/test-results/20260629T021012Z-p3947634`; pass root `.cartulary/test-results/20260629T021323Z-p3976097` |
| WS-5 | `git diff --check` | PASS | no artifact |
| WS-6 | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260629T021323Z-p3976097` | PASS; retained run accepted, generated updated 3, duration refreshed, run checks pass | `.cartulary/test-results/20260629T021850Z-p4037876` |
| WS-6 | `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260629T021938Z-p4040119` |
| WS-6 | `make json-shape-check` | PASS | `.cartulary/test-results/20260629T021938Z-p4040141` |
| WS-6 | `make lint-markdown` | PASS | no output artifact |
| WS-6 | `make go-test-duration-baseline-coverage` | PASS | `.cartulary/test-results/20260629T022322Z-p4042878` |
| WS-6 | `make go-test-duration-baseline-drift` | PASS | `.cartulary/test-results/20260629T022322Z-p4042882` |
| WS-6 | `make phase-ledger-drift` | PASS | `.cartulary/test-results/20260629T022322Z-p4042915` |
| WS-6 | `make phase-schedule-drift` | PASS | `.cartulary/test-results/20260629T022322Z-p4043176` |
| WS-6 | `git diff --check` | PASS | no artifact |

Standalone browser E2E was initially skipped for this slice because public routes, generated contracts, frontend code, schema inputs, and WS payload construction were unchanged; `make check` later ran the browser-backed check shards as part of broad S-08 validation. Run browser targets before merging any later route/WS/frontend-facing slice.

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
| `2026-06-28 21:31:50 EDT -0400` | `main` / `6662cfa` | `internal/modules/timeline` / remediation register | WS-0 complete; WS-1 next | WF-06 and WF-12 planning scope for S-05 through S-08 | `docs/handoffs/Timeline-Module-Refactoring-Tracker.md` | `git status --short --branch`; `date`; `git rev-parse --abbrev-ref HEAD`; `git rev-parse --short HEAD`; tracker inspection with `sed`/`nl` | Tracker updated; Core specs/contracts unchanged | None | Added remediation gap register, workstream sequencing, S-05/S-07/S-08 implementation notes, validation accounting, and Core owner-contradiction stop rule | None | None | WS-1 pre-port guardrail tests | `make task-guide ROLE=feature-dev PHASE=phase3` |
| `2026-06-28 21:38:26 EDT -0400` | `main` / `6662cfa` | `internal/modules/timeline` and `internal/modules/revisions` / pre-port guardrails | WS-1 complete; WS-2 next | WS-1 | `docs/handoffs/Timeline-Module-Refactoring-Tracker.md`, `internal/modules/timeline/store.go`, `internal/modules/timeline/phase3_query_schema_guard_test.go`, `internal/modules/revisions/rollback_mapping_guard_test.go` | task guides for phases 3/4/7/9; `make format`; `make backend-unit`; `make backend-store`; `make backend-integration`; `git diff --check` | backend-unit, backend-store, backend-integration, diff check | None | Added Timeline query/schema mapping guard, rollback mapping guard, and removed unsupported `timeline.evidence_count` sort mapping to match Core 01/view schema | None | None | WS-2 S-05 private ports | `make backend-unit` |
| `2026-06-28 21:46:49 EDT -0400` | `main` / `6662cfa` | `internal/modules/timeline` / private peer ports | WS-2 complete; WS-3 next | S-05, WS-2 | `docs/handoffs/Timeline-Module-Refactoring-Tracker.md`, `internal/modules/timeline/ports.go`, `internal/modules/timeline/store.go`, `internal/modules/timeline/clipboard_paste_store.go` | `make format`; `make backend-unit`; `make backend-store`; `make backend-integration`; `git diff --check` | backend-unit, backend-store, backend-integration, diff check | None | Added private ports and production adapters for idempotency, records, revisions, projections, links, and mentions; removed direct peer-store construction from Timeline mutation paths while preserving transaction order | Should WS-5 replace the residual `entities` error import in `store.go` with a Timeline error adapter? | None | WS-3 S-06 internal file split | `make test-fast` |
| `2026-06-28 21:56:45 EDT -0400` | `main` / `6662cfa` | `internal/modules/timeline` / internal file split | WS-3 complete; WS-4 next | S-06, WS-3 | `docs/handoffs/Timeline-Module-Refactoring-Tracker.md`, `internal/modules/timeline/store.go`, `internal/modules/timeline/time_conversion_store.go`, `internal/modules/timeline/query_projection_store.go`, `internal/modules/timeline/lifecycle_store.go`, `internal/modules/timeline/mentions_collections_store.go` | `make format`; `make backend-unit`; `git diff --check`; `make test-fast`; `make test-fast` rerun | backend-unit, diff check, second test-fast | First test-fast failed on object-store readiness before passing on rerun | Split time conversion, query/projection, lifecycle actions, and mention/collection behavior into dedicated same-package files; left public symbols unchanged | None | None | WS-4 S-07 boundary guardrails | `make backend-unit` |
| `2026-06-28 21:59:38 EDT -0400` | `main` / `6662cfa` | `internal/modules/timeline` / import-boundary guardrails | WS-4 complete; WS-5 next | S-07, WS-4 | `docs/handoffs/Timeline-Module-Refactoring-Tracker.md`, `internal/modules/timeline/boundary_guard_test.go` | `make format`; `make backend-unit`; `git diff --check` | backend-unit, diff check | None | Added a Go import-boundary guard that runs under `backend-unit`; no generated Make target was edited | WS-5 should tighten the temporary `entities` allowance after error mapping is narrowed | None | WS-5 S-08 cleanup and adapter narrowing | `rg -n \"timeline\\.NewStore|\\*timeline\\.Store|SetStoreHooksForTesting|BuildRow\" internal cmd` |
| `2026-06-28 22:16:08 EDT -0400` | `main` / `6662cfa` | `internal/modules/timeline`, Workbook, Imports, phase accounting / export cleanup | WS-5 complete; WS-6 next | S-08, WS-5 | `docs/handoffs/Timeline-Module-Refactoring-Tracker.md`, `internal/modules/imports/routes.go`, `internal/modules/timeline/{api.go,facade.go,routes.go,store.go,phase3_decoder_test.go,phase3_integration_test.go,phase3_projection_contract_test.go}`, `internal/modules/workbook/routes.go`, `tools/phase3_test_map.json`, `tools/go_test_duration_baselines.json`, generated phase ledger/schedule outputs | `rg` caller audits; `make format`; `make phase-map-check`; `make backend-unit`; `make backend-store`; `make backend-integration`; `make backend-integration-support`; baseline refresh/drift targets; `make phase-ledgers`; `make phase-schedules`; `make json-shape-check`; `make phase-ledger-drift`; `make phase-schedule-drift`; `make check`; `git diff --check` | backend-unit, backend-store, backend-integration, backend-integration-support, phase-map-check, baseline coverage/drift, json-shape-check, phase-ledger/schedule drift, check, diff check | Initial `make check` failed on phase3 map accounting and then duration-baseline/schedule drift; resolved by support-test renames, phase map updates, Make-owned baseline refresh, and phase ledger/schedule generation | Unexported low-risk helpers (`BuildRow`, facade/store hook builders), centralized Workbook mention mutation error predicate, documented Imports decode seam, retained `Store`/`NewStore`, `SetStoreHooksForTesting`, and wire DTOs where active tests/callers still need them | Future cleanup can tighten temporary `entities` route allowance after a Timeline error contract replaces direct entity errors | None | WS-6 final validation and handoff completion | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260629T021323Z-p3976097` |
| `2026-06-28 22:20:02 EDT -0400` | `main` / `6662cfa` | Timeline remediation / final validation and handoff | COMPLETE | WS-0 through WS-6; S-00 through S-08 | `docs/handoffs/Timeline-Module-Refactoring-Tracker.md`, Timeline split/port/guard files, Workbook/Imports adapter touchpoints, Phase 3 map/ledger/schedule/duration outputs | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260629T021323Z-p3976097`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-markdown`; duration baseline coverage/drift; phase ledger/schedule drift; `git diff --check` | agent-finalize, generated-artifact-policy-check, json-shape-check, lint-markdown, duration baseline coverage/drift, phase ledger/schedule drift, diff check | None in final slice | Retained successful `make check` run for finalizer; finalizer refreshed duration baselines and generated phase artifacts through Make | None | None | Review/commit the completed remediation set | `make check` |

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
