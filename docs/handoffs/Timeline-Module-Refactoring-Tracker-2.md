# Timeline Module Refactoring Tracker 2

## Session Header

- Artifact path: `docs/handoffs/Timeline-Module-Refactoring-Tracker-2.md`
- Target directory: `internal/modules/timeline`
- Branch/commit: initial/S-01 reference `main` at `a6cd41d2dba44094bed875551633bd6a679eb54c`; gap-closure update inspected `main` at `d0afbad53dabfddd8688b6c9f2965f9532106045`
- Dirty-tree state: clean before this gap-closure tracker edit; this doc-only update makes the tree dirty.
- Date/time: initial plan `2026-06-29T17:27:31-04:00`; slice update `2026-06-29T17:58:13-04:00`; gap-closure update `2026-06-29T18:13:45-04:00`
- Agent/mode: Codex, implementation handoff update
- Framework: `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- Source limits: Read `AGENTS.md`, framework, `docs/domain.md`, all Timeline production files, selected Timeline tests, and focused adjacent backend/contract/SQL/harness files. Core 01-04 were searched but not fully read in the initial plan; read exact owner sections before implementation that touches public behavior. Frontend Timeline files and generated artifacts were searched, not fully inspected. Slice update inspected code diffs and retained validation artifacts only. Gap-closure update rechecked this tracker, Core 01/Core 03 grep evidence, authored Timeline view schema, Timeline query sort map, duplicated Entities Timeline row presenter evidence, and R04/R07/R09 research support.

## Slice Update - S-01/S-02 Facade Boundary

Implemented on `2026-06-29T17:58:13-04:00`.

Changed files:

- `internal/modules/timeline/facade.go`
- `internal/modules/timeline/routes.go`
- `internal/modules/timeline/boundary_guard_test.go`
- `internal/modules/workbook/routes.go`
- `internal/modules/imports/routes.go`
- `docs/handoffs/Timeline-Module-Refactoring-Tracker-2.md`

Implemented scope:

- Added Timeline command DTOs and command-style facade methods for create, imported create, patch, conflict resolve, clipboard paste, mark reviewed, and supersede.
- Kept legacy facade methods as compatibility shims that delegate to the new command methods.
- Moved workbook/imports/timeline production route callers onto the command facade for Timeline operations.
- Preserved workbook bulk mutation's existing `BulkMutationRequestHash` override by passing `RequestHash` through `ClipboardPasteCommand`.
- Left workbook decision supersede on the existing decision-store hash path; Timeline supersede now lets the Timeline facade compute its own action hash.
- Added a static boundary guard test blocking inspected production callers from regressing to legacy Timeline facade calls.
- No generated files were edited; `make agent-finalize` reported generated artifacts unchanged.

Post-slice scan:

- `rg -n "timelineStore\\.(CreateTimelineRow|CreateImportedTimelineRow|PatchTimelineRow|ResolveTimelineConflict|ClipboardPaste\\(|Supersede\\()|facade\\.MarkReviewed\\(" internal/modules/workbook internal/modules/imports internal/modules/timeline -g '*.go'` returned no matches.
- `rg -n "Timeline(Create|Patch|ClipboardPaste|ConflictResolve)RequestHash" internal/modules/workbook internal/modules/imports internal/modules/timeline/routes.go -g '*.go'` returned no matches.
- Remaining inspected production `TimelineActionRequestHash` usage is `internal/modules/workbook/routes.go:670`, isolated to workbook decision supersede.

## Target Scope And Non-Scope

In scope: behavior-preserving planning for Timeline capture/mutation, request DTOs, facade, validation/defaulting, idempotency, row-version conflicts, lifecycle actions, clipboard paste, raw capture, time conversion, projection refresh, revision writes, mention/entity hooks, evidence hooks, and WebSocket record-change effects.

Adjacent inspected: `internal/modules/workbook`, `internal/modules/imports`, `internal/modules/entities`, `internal/modules/projections`, `internal/modules/revisions`, `internal/modules/evidence`, `internal/app`, `db/migrations`, `db/queries`, `contracts`, `tools/task_surface.generated.mk`, and Make task guidance.

Non-scope: external behavior changes, route redesign, schema redesign, UI redesign, phase-accounting rewrites, generated-file hand edits, new runtime phase dependencies, or moving files before characterization coverage is frozen.

## Current-State Inventory

| Path | Current responsibility | Target owner | Public/private/generated | Incoming deps | Outgoing deps | External contracts touched | Risk | Notes |
|---|---|---|---|---|---|---|---|---|
| `api.go` | View schema ID, route keys, request decoders, request hashes, row/action payload builders, error helpers | Timeline facade/API boundary | Public | workbook, imports, tests | auth, httpapi, viewquery, viewschema | OpenAPI, view row wire, error wire | High | Public surface is broad. |
| `facade.go` | Public facade; now includes command DTOs/methods delegating to existing store plus legacy shim methods | Timeline public facade | Public | workbook, imports | store, postgres, viewschema, authn | Service API | High | S-01 implemented command shell; store still exported/private cleanup remains TODO. |
| `routes.go` | Timeline-specific HTTP routes, test routes, auth checks, WS publish | Timeline route adapter | Public module entry | app runtime | auth, entities, incidents, httpapi, ws | HTTP, WS | Medium | Transport logic should stay thin. |
| `store.go` | Create/import/patch/conflict persistence, idempotency, revisions, projections | Timeline private application service/store | Public today, private target | facade, tests | sqlc, incidents, authn, postgres, viewschema | storage, row_version, idempotency | High | Main behavior concentration. |
| `ports.go` | Adapters to authn, records, revisions, projections, links, entities | Timeline private ports | Private target | store | sibling modules | revision/projection/entity hooks | High | Good pattern, but still direct sibling-store coupling. |
| `query_projection_store.go` | Timeline query over projection table and collections | Timeline query store | Private target | facade/store | generated sql, viewschema | view query contract | High | Must preserve field-key query semantics. |
| `mentions_collections_store.go` | Host/identity mentions, tags, attached evidence collections | Timeline collection service | Private target | store | entities, authn | mention_origin, evidence collection wire | High | Critical low-friction capture behavior. |
| `auto_resolution.go` | Narrow exact alias auto-resolution | Timeline mention policy | Private | mentions store | fieldnorm | mention semantics | High | Must not broaden into fuzzy/implicit entity creation. |
| `clipboard_paste.go` | Clipboard request decode and paste plan | Timeline clipboard API helper | Public today | workbook, imports tests | auth, tabularingest, viewschema | workbook paste | High | Public callers use DTOs directly. |
| `clipboard_paste_store.go` | Batch create/patch paste transaction | Timeline private application service | Private target | facade/store | incidents, authn | paste/idempotency/raw capture | High | Cross-row conflict semantics. |
| `lifecycle_store.go` | Mark reviewed and supersede | Timeline private lifecycle service | Private target | facade/store/routes | incidents, authn | lifecycle routes | High | Terminal-state and rollback coupling. |
| `time_conversion_store.go` | Incident fixed-offset profile and time-pair generation | Timeline private service | Private target | facade/routes/store | incidents, timecontract, authn | time profile route, projection fields | Medium | Public profile route exists. |
| `state.go` | Capture-state vocabulary and transitions | Timeline domain policy | Private target | store/tests | none | capture state | Medium | Preserve rough/enriched/reviewed/superseded semantics. |
| `timing.go` | Timeline create timing instrumentation | Timeline implementation support | Private | workbook route | context | timing debug header | Low | Do not make product contract unless specified. |
| `hooks.go` | Package-level store test hook | Timeline test seam | Private/test target | tests | none | none | Medium | Prefer injectable store options later. |
| `timecontract/timecontract.go` | UTC/local parse and fixed offset formatting | Shared time utility | Public subpackage | timeline, entities, projections | stdlib | time conversion | Medium | Keep stable for first refactor; reassess ownership later. |
| `*_test.go` in target | Characterization and guards | Timeline tests | Test | Make harness | testutil, modules | phase evidence | Medium | Test names searched; selected bodies inspected. TODO: full body audit before code moves. |
| `internal/modules/workbook/*` | Generic workbook routes dispatch Timeline create/patch/paste/bulk/supersede/conflict | Workbook route owner + Timeline facade caller | Public app code | HTTP clients | timeline | HTTP/API, WS | High | S-02 moved Timeline operations to command facade; decision supersede still uses existing workbook hash path. |
| `internal/modules/imports/routes.go` | Imports apply Timeline creates and raw capture | Imports + Timeline facade caller | Public app code | import routes | timeline | import apply | High | S-02 moved imported Timeline create to command facade after raw capture columns are applied. |
| `internal/modules/entities/*` | Mention lifecycle and duplicate Timeline row projection | Entities + Timeline hook seam | Public app code | mention routes | timecontract | mention resolution, row_version | High | `timeline_projection.go` duplicates row wire. |
| `internal/modules/evidence/store.go` | Attached-evidence refresh emits Timeline changed keys | Evidence hook | Public app code | evidence routes | ws/projections | evidence counts, WS | Medium | Hard-coded Timeline field keys. |
| `internal/modules/revisions/*` | Timeline rollback mapping/projection rebuild | Revisions hook | Public app code | history routes | projections | rollback/revisions | High | Timeline source mapping must not drift. |
| `internal/modules/projections/store.go` | Timeline projection input/upsert/rebuild | Projection module | Public app code | timeline/revisions | timecontract | projection table | High | Projection table is derived state. |
| `contracts/*`, `db/*` | Authored contracts and SQL inputs | Contract/storage owners | Authored | generators/app | generated outputs | OpenAPI/view schema/sqlc | High | Do not hand-edit generated downstream outputs. |
| `internal/gen/**`, generated TS | Generated outputs | Generators | Generated | app/packages | none | derived contracts | Medium | Searched only; never hand-edit. |
| `apps/web/src/workbook/timeline/**` | Frontend Timeline consumer | Frontend owner | Authored | browser users | API/WS | UI behavior | TODO | Searched only; TODO: inspect if wire shape changes. |

## Timeline Module Contract Map

| Surface | Specific contract | Owner source | Repo evidence | Characterization coverage | Drift risk |
|---|---|---|---|---|---|
| HTTP create/patch/query | Timeline v2 rows use `record_id`, `row_version`, `field_key`, strict request shape, zero-field create | Core 01; OpenAPI; view schema | `api.go`, workbook routes, `contracts/openapi` | Phase3 decoder/store/integration tests | High |
| Lifecycle actions | Mark reviewed and supersede require `base_row_version` and `client_txn_id` | Core 01/02/03; OpenAPI | `lifecycle_store.go`, `routes.go` | Phase3 store/integration | High |
| Storage/revisions | Writes `timeline_events`, `records`, `change_sets`, `record_revisions`, links, raw capture | Core 01/02; migrations | `00004`, `00006`, `00015`, `00037`, `store.go` | Phase3/9 store + revisions guards | High |
| Row version/OCC | Stale base rejects or same-field conflict; non-overlap can auto-rebase | Core 01/03 | `store.go`, OpenAPI schemas | Phase3 concurrency/conflict tests | High |
| Idempotency | Route/actor/scope/client_txn_id replay with request hash; divergent hash conflicts | Core 01/Auth | `route_idempotency`, `store.go`, `lifecycle_store.go` | Phase3 idempotency tests | High |
| Projection refresh | Source writes update `timeline_grid_projection`; query reads projection + collections | Core 01 | `projections.Store`, `query_projection_store.go`, sqlc | Phase3 projection/query guards | High |
| Mention/entity hooks | Host/identity refs are `mention_origin`; unresolved raw text preserved; narrow auto-resolution only | Core 02/03; domain | `mentions_collections_store.go`, `auto_resolution.go`, entities hooks | Phase4 tests | High |
| Collaboration/WS | `record_changed` includes row version, client txn, changed keys, affected views, optional patch cells | Core 01/03; WS schema | `platform/ws`, workbook/timeline/entities/evidence routes | Phase3/5/6/8 tests | High |
| Generated contracts | Generated Go/TS derive from authored OpenAPI/view schemas | Generated policy | `tools/generated_artifact_policy.json`, `internal/gen/**`, TS generated roots | TODO: run drift checks when contracts touched | Medium |
| Harness accounting | Use Make-owned phase slices and retained artifacts | `docs/testing-harness-nlspec.md` | `make task-guide`, `make explain-target` | Not validation-run in this session | Medium |

## Boundary And Coupling Scan

| Finding | Evidence | Risk | Classification | Proposed owner | Required action |
|---|---|---|---|---|---|
| Facade is too thin; callers still depend on Timeline DTOs, hashes, and errors | S-01/S-02 added command facade and migrated inspected production Timeline mutation callers; decoders/errors still leak to route adapters | Public API sprawl | should_fix | Timeline facade | Next: close error mapping boundary and decide DTO ownership. |
| `Store` and `NewStore` are exported | tests/testutil and some module tests import them | Internals become contract | should_fix | Timeline | Hide behind facade after tests are converted. |
| Entity mention lifecycle duplicates Timeline row projection | `internal/modules/entities/timeline_projection.go` mirrors `buildRow` and may drift | Wire-shape mismatch | must_fix | Timeline + Entities | Add parity characterization, then consolidate through a Timeline-owned leaf row presenter/contract seam. |
| Direct sibling-store adapters in Timeline | `ports.go` wraps records/revisions/projections/links/entities | Tight module coupling | should_fix | Timeline ports | Keep ports private; move construction/injection behind facade. |
| Timeline field keys hard-coded outside Timeline | evidence, entities, revisions, incidents, frontend, contracts | Field-key drift | should_fix | Contract layer + module hooks | Prefer generated/registry-backed contract lookups or leaf contract support; do not make private Timeline runtime constants the public key authority. |
| Transport/auth logic lives in module routes | `routes.go` imports `httpapi`, `ws`, auth modules | Boundary blur | intentional/defer | Route adapters | Keep thin; do not move in first slices. |
| `timecontract` public subpackage imported by entities/projections | `timecontract` imports found outside Timeline | Ownership ambiguity | defer | Timeline/time utility | Keep stable now; evaluate platform move separately. |
| Package-level test hook | `hooks.go` global hook | Test-only global state | should_fix | Timeline tests | Replace with injectable store option in cleanup slice. |
| Phase-named tests exist | `phase*_test.go`, task maps | Phase language leakage risk | intentional for tests | Harness | Do not introduce runtime phase dependencies. |

## Gap Closure Actions - Core And Implementation

These actions close the current open questions without letting implementation mechanics leak into Core documents. Core 00-04 should describe observable product behavior only. The tracker, appendices, and supporting handoffs should describe Go package boundaries, facade shape, validation sequencing, and refactor mechanics.

| Gap | Concrete closure decision | Core document content | Appendix/supporting content | Implementation closure action | Validation |
|---|---|---|---|---|---|
| `timeline.evidence_count` sortable mismatch | Treat `timeline.evidence_count` as intentionally sortable. Current Core 01 already says field `sortable` must equal `sort_fields` membership, and the authored view schema currently marks the field sortable while omitting it from `sort_fields`. | Core 01 Timeline registry entry must include `timeline.evidence_count` in `sort_fields`; define it as a read-only non-negative integer count defaulting to `0`, hidden by default, not groupable, and not filterable in the current profile. Keep the existing Core 01 invariant that sortable field metadata must be backed by `sort_fields` or a valid `header_sort_field_key`. | Record this as an owner-contract/schema consistency fix, not an implementation guess. Generated artifacts remain downstream only. | Edit authored Core 01 text and `contracts/view-schemas/cartulary.view.timeline.v2.json`; add SQL sort mapping in Timeline query store; regenerate through Make only; do not hand-edit generated roots. | Add schema guard for sortable metadata consistency and query test for ascending/descending `timeline.evidence_count`; run `make json-shape-check`, `make generated-artifact-policy-check`, `make backend-unit`, `make phase-slice PHASE=phase3`, and `make frontend-typecheck` if generated TS changes. |
| Entity mention lifecycle duplicates Timeline row shape | Timeline owns Timeline row wire shape. Entities may trigger Timeline mention effects but must not maintain a parallel Timeline row presenter. | Core 03 should explicitly state observable mention lifecycle effects: raw mention preservation, no implicit entity creation, selected-mention-only resolve/dismiss/restore, exact source collection changes, and `timeline.has_unresolved_mentions` recomputation. Core should not name Go packages. | Add parity characterization as the prerequisite for S-05. The row presenter seam belongs in Timeline supporting material as a module-owned helper/hook with no store or DB dependency. | Introduce a Timeline-owned leaf presenter package or hook with DTO inputs; have Timeline and Entities adapt their projected rows into it, avoiding an Entities import of the Timeline root package. Include `timeline.attached_evidence_ids` in the parity scope. | Add row parity test covering host refs, identity refs, tags, attached evidence ids, `timeline.evidence_count`, `timeline.has_evidence`, and `timeline.has_unresolved_mentions`; run `make phase-slice PHASE=phase4` plus phase3/5 coverage if patch payloads change. |
| Imported mention auto-resolution ambiguity | Imported or machine-applied rows must not run interactive auto-resolution. Interactive Timeline relationship-cell create/patch and user-initiated paste may auto-resolve only exact, unique, unsuppressed host/identity aliases. | Core 03 should define auto-resolution eligibility by mutation source and preserve the Core 01 `mention_token_text_v1` contract. Imports and machine ingestion preserve unresolved mentions unless explicit resolved refs are supplied. | Tracker should keep import auto-resolution as a characterization item until the test exists. R04 supports low-friction partial capture; it is supporting evidence, not normative authority. | Add or locate import coverage proving raw host/identity mention text remains unresolved on import when only an alias-like token is supplied. | Run `make phase-slice PHASE=phase11` for imports and `make phase-slice PHASE=phase4` for mention semantics. |
| Error boundary closure | Core owns public wire errors; implementation owns Go error categories. Route adapters should map closed Timeline error kinds to existing API envelopes. | Core 01/03 should list stable public Timeline mutation error codes/details only where missing or ambiguous: validation, row-version conflict, same-field conflict, client transaction conflict, illegal lifecycle transition, no-effective-change, incident closed, and not-found behavior. | Keep S-03 as implementation-only unless a missing public wire code is found. Do not introduce transport or auth dependencies into private store code. | Add a closed Timeline mutation error classifier such as `TimelineErrorKind`; route/workbook adapters map it to existing `auth.APIError` envelopes without inspecting private store error structs. | Add error mapping tests that prove HTTP envelopes remain unchanged; run `make backend-unit` and `make phase-slice PHASE=phase3`. |
| Store privacy and legacy facade shims | No Core change. This is a module-boundary cleanup after callers and tests move to command facade paths. | None. | Legacy shim methods remain for one compatibility slice only. S-04 removes or narrows them after production callers and tests/testutil stop depending on store internals. | Migrate tests/testutil to facade or test-only helpers; then hide `Store`, `NewStore`, and persistence details where Go package cycles allow. | Run `make backend-unit`, `make phase-slice PHASE=phase3`, `make phase-slice PHASE=phase4`, and `make phase-slice PHASE=phase9`. |
| Phase3 visual golden diffs | Treat as a validation blocker, not a Core contradiction, until triage proves product UI drift. | No Core change unless triage proves an intended observable behavior or layout contract changed. | Use R09 as supporting grid-layout/visual-test evidence and the visual golden maintenance guide. Do not update goldens ad hoc. | Inspect V-3-GRID-01/02/03 actual/diff artifacts. If environmental/subpixel drift, rerun or refresh only through the approved guide with owner approval; if UI drift, fix implementation or update owner docs first. | Re-run `make phase-slice PHASE=phase3` or the narrower visual target recommended by `make explain-run`; do not claim phase3 passed until this clears or is owner-approved. |
| Hard-coded Timeline field keys outside Timeline | Public field keys are view-schema contract keys, not private Timeline runtime constants. | Core 01 remains the field-key authority. Add no Timeline-runtime coupling rule to Core beyond the sortable consistency invariant already noted. | Supporting docs should prefer generated/registry-backed constants or schema lookups where practical and avoid importing Timeline runtime solely for strings. | Do not centralize public field keys inside private store code. If constants are needed, keep them in contract/generated or leaf contract support, not in private Timeline persistence. | Add static guard only if drift recurs; otherwise validate through schema/query tests. |
| `timecontract` ownership | Not a blocker for the Timeline boundary slice. | Core 01 timestamp and string contracts remain normative. | Keep `timecontract` ownership as deferred utility cleanup. R07 supports sortable UTC timeline practice as research context only. | Do not move `timecontract` in this refactor. Reassess in a separate utility-boundary plan if needed. | Existing phase3 support coverage remains the relevant guard. |

Research support note: R04 supports low-friction, partial, evolving capture and cautions against over-structuring first capture; R07 supports evidence-linked Timeline triage and sortable time practice; R09 supports treating grid visual diffs as layout evidence that must be triaged through visual tests. These reports are supporting material only and do not override Core 00-04.

## Characterization-Test Plan

| Behavior | Existing test/evidence | Missing evidence | Required characterization test or blocker | Command | Status |
|---|---|---|---|---|---|
| Low-friction/zero-field rough create | Phase3 decoder/store/integration | Full Core owner line refs | Read Core 01/04 sections before code | `make phase-slice PHASE=phase3` | TODO |
| Strict create/patch validation | Phase3 decoder tests | None known | Preserve unknown/duplicate/trailing JSON rejection | `make phase-slice PHASE=phase3` | TODO |
| Patch OCC and same-field conflict | Phase3 store/integration | None known | Preserve conflict token payload and no idempotency write on conflict | `make phase-slice PHASE=phase3` | TODO |
| Route idempotency | Phase3 integration/store | None known | Preserve route/actor/scope replay semantics | `make phase-slice PHASE=phase3` | TODO |
| Projection query/sort/filter/group | Projection contract/query schema guards | Authored schema/Core consistency gap: `timeline.evidence_count` field metadata is sortable, but current `sort_fields` and query sort map omit it | Add Core/schema/query closure: make `timeline.evidence_count` sortable, add guard that sortable metadata matches `sort_fields` or `header_sort_field_key` | `make phase-slice PHASE=phase3` | TODO |
| Mention extraction and no implicit entity creation | Phase4 unit/integration | Full entity-driven row parity, including attached evidence cells | BLOCKED: add parity test for entity mention lifecycle row payload vs Timeline-owned row presenter before consolidating duplicate row-wire logic | `make phase-slice PHASE=phase4` | BLOCKED |
| Auto-resolution narrowness | Phase4 request/integration | None known | Preserve exact unique alias and suppressor behavior | `make phase-slice PHASE=phase4` | TODO |
| Evidence attachment hooks | Evidence phase5 tests searched | Timeline row patch parity through evidence not fully inspected | Add/locate test for attached evidence row patch includes ids/count/flag | `make phase-slice PHASE=phase5` or TODO discover | TODO |
| Clipboard paste/raw capture | Timeline/workbook phase9 tests; `make phase-slice PHASE=phase9` passed after S-02 | None known | Preserve raw unmapped columns, target conflict behavior, batch conflicts | `make phase-slice PHASE=phase9` | DONE |
| Import apply | Imports integration tests searched | Imported auto-resolution-disabled assertion unclear | Add/locate test proving imported or machine-applied host/identity mention tokens stay unresolved unless explicit resolved refs are supplied | `make phase-slice PHASE=phase11` | TODO |
| Time-conversion profile | Phase3 support integration | Edge cases for generated UTC/local pair | Preserve fixed-offset profile and profile-version conflict | `make phase-slice PHASE=phase3` | TODO |
| WS collaboration effects | Phase3/5/6/8 tests searched; phase3 backend/browser functional/measurement passed in service-backed slice | Full changed-key audit; visual target failing as validation blocker, not Core contradiction | Preserve `record_changed` payload and sparse patch cells; triage V-3-GRID-01/02/03 before claiming phase3 pass | `make phase-slice PHASE=phase3` | BLOCKED |

## Refactor Slice Plan

| Slice | Depends on | Change | Files likely touched | Behavior unchanged | Validation | Rollback note | Status |
|---|---|---|---|---|---|---|---|
| S-00 Characterization freeze | None | Add/locate missing parity tests; no production moves | tests only | All current behavior | phase3/4/9 slices | Drop tests only | DEFERRED |
| S-01 Facade command/result shell | S-00 | Add facade command/result/error wrappers that delegate to current store | `facade.go`, `api.go` | Wire payloads, hashes, errors | phase3 | Revert wrappers | DONE |
| S-02 Caller conversion | S-01 | Convert workbook/imports to facade methods; keep old helpers temporarily | workbook/imports + facade | Routes and responses | phase3/9/11 as needed | Revert caller edits | DONE |
| S-03 Error boundary closure | S-02 | Add Timeline-owned error classification/HTTP mapping seam; remove caller type peeking where possible | facade/api/routes/workbook | Error codes/details | phase3/9 | Revert mapping seam | TODO |
| S-04 Store privacy | S-03 | Make store constructor/internal types private; update tests/testutil through facade or test-only helpers | timeline + testutil | Storage behavior | phase3/4/9 | Re-export temporarily | TODO |
| S-05 Projection row contract | S-00 | Consolidate duplicate row-wire logic through a Timeline-owned leaf row presenter after parity tests | timeline/entities/projections as needed | View row shape, including `timeline.attached_evidence_ids` | phase3/4/5/8 | Restore old builder/presenter adapter | BLOCKED |
| S-06 Hook cleanup | S-04 | Replace global test hook with injectable option | timeline tests/store | Commit timing behavior | phase3 | Restore hook | TODO |
| S-07 Generated/contracts audit | Any contract touch | Update authored contracts only, then generate; never hand-edit generated roots | contracts/tools/generated outputs | Contract derivation | drift targets | Revert owner input | DEFERRED |
| S-08 Broad verification/handoff | All | Run required phase slices, `agent-finalize`, then broader gate | none | All public behavior | `make test-fast`/`make check` | Use last green slice | TODO |

## Backend Facade Target Shape

- Public facade: `TimelineFacade`/`Facade` is the only production entry for create, imported create, patch, conflict resolve, clipboard paste, mark-reviewed, supersede, query, record-incident lookup, snapshot substrate, and time-conversion profile.
- Command/query DTOs: facade-level commands carry actor, incident/record IDs, request IDs, client transaction IDs, base row versions, field-key changes/actions, request hash, and operation time. HTTP JSON structs may remain route-local or facade-owned, but store internals should not leak to callers.
- Validation/defaulting: Timeline owns Timeline v2 view-schema enforcement, zero-field create eligibility, visible-text normalization, collection action validation, rough-capture defaults, max change/action counts, and time-pair defaulting.
- Errors: expose closed Timeline error categories with stable details; route adapters map them to HTTP/API envelopes. Callers should not inspect private store error structs.
- Private store interface: persistence, idempotency, revision, projection, link, mention, and record-version ports are private to the module.
- Hooks: stable internal hook seams for projections, revisions, entities/mentions, evidence, and collaboration changed-key results; no phase-shaped runtime dependency.
- Tests: group by behavior contracts after refactor; phase names may remain for harness accounting, but behavior labels must be clear.

## Workflow Dependency Map

| Workflow | Status | Dependencies | Outputs | Exit criteria |
|---|---|---|---|---|
| WF-00 Session/source bootstrap | DONE | AGENTS/framework/domain/git | Session header/source limits | Branch, commit, mode, limits recorded |
| WF-01 Current-state repository scan | DONE | WF-00 | File/import/contract scan | Target and adjacent surfaces identified |
| WF-02 Module ownership inventory | DONE | WF-01 | Inventory table | Owners/risks recorded with TODO limits |
| WF-03 Public contract freeze | IN_PROGRESS | WF-01 | Contract map plus gap-closure actions | Exact Core 01/03 sections read and edited, or explicitly confirmed unchanged, before contract implementation |
| WF-04 Refactor slice selection | DONE | WF-02/03 | Slices S-00..S-08; S-01/S-02 implemented | Each slice independently reviewable |
| WF-05 Characterization test plan | BLOCKED | WF-03 | Missing-test list plus closure tests | Evidence-count sort test and import auto-resolution test added; mention row parity blocker resolved |
| WF-06 Boundary guardrail plan | DONE | WF-02 | Static production caller guard added in `boundary_guard_test.go` | Guard blocks inspected legacy Timeline facade production calls |
| WF-07 Backend module facade plan | DONE | WF-04 | Facade target shape | DTO/error/store boundaries decided |
| WF-09 Execution checkpoint plan | IN_PROGRESS | WF-05 | S-01/S-02 handoff record added | Handoff record filled after each slice |
| WF-10 Validation/harness accounting plan | DONE | task-guide/explain-target | Validation table | Required Make targets named |
| WF-11 Documentation/generated-artifact plan | DONE | generated policy | No hand-edit rule | Owner inputs and drift targets identified |
| WF-12 Cleanup/anti-drift plan | TODO | S-04/S-05 | Store privacy and duplication cleanup | No broad public internals remain |
| WF-13 Handoff/next-slice bootstrap | TODO | Each slice | Handoff record | Safe restart command and blockers recorded |

## Top-Level Tracker

| ID | Task | Status | Notes |
|---|---|---|---|
| T-001 | Define target module and scope | DONE | Timeline low-friction capture/mutation preserved. |
| T-002 | Inspect current repo state | DONE | Source limits recorded. |
| T-003 | Map owner contracts | IN_PROGRESS | Gap-closure actions identify required Core 01/Core 03 edits; exact sections still must be read before changing owner docs. |
| T-004 | Freeze characterization evidence | BLOCKED | Evidence-count sort and import auto-resolution tests are actionable TODOs; mention row parity still blocks S-05; phase3 visual golden remains validation blocker only. |
| T-005 | Plan boundary guardrails | DONE | Static guard added for inspected production Timeline facade callers. |
| T-006 | Plan behavior-preserving moves | DONE | Facade-first slices defined. |
| T-007 | Plan validation loop | DONE | S-01/S-02 validation recorded; gap-closure validation targets listed for future Core/contract/code edits. |
| T-008 | Update docs/contracts if required | IN_PROGRESS | Core 01/Core 03 and authored Timeline view schema edits are now identified; generated roots remain downstream only. |
| T-009 | Execute or hand off | IN_PROGRESS | S-01/S-02 implemented; gap-closure tracker update added; next handoff should resolve Core/schema evidence-count closure or S-03 error boundary. |

## Validation Plan

| Target | Purpose | Required? | Expected artifact | Failure handling |
|---|---|---|---|---|
| `make task-guide ROLE=feature-dev PHASE=phase3` | Discover Timeline create/patch/query/lifecycle evidence | Discovery only | terminal guidance | Update plan if guidance changes |
| `make phase-slice PHASE=phase3` | Validate core Timeline mutation/projection/collab | Required for most slices | `.cartulary/test-results/<run>/phase-slice/target-summary.json` | Currently blocked by `browser-e2e-visual`; do not treat phase3 as passing until visual diff is resolved or owner-approved |
| `make phase-slice PHASE=phase4` | Validate mention and auto-resolution behavior | Required when mention hooks touched | same | Stop on mention/entity failures |
| `make phase-slice PHASE=phase9` | Validate clipboard/bulk behavior | Required when paste/bulk touched | same | Stop on batch conflict/raw-capture failures |
| `make phase-slice PHASE=phase11` | Validate import apply and machine-ingestion semantics | Required when imported Timeline create or auto-resolution eligibility changes | same | Stop on import/raw-capture/mention failures |
| `make service-backed-slice PHASE=phase3` | Persistence-backed Timeline validation | Required for store/projection moves | `.cartulary/test-results/<run>/service-backed-slice/target-summary.json` | Currently blocked by `browser-e2e-visual`; backend/store/browser functional children passed in latest run |
| `make generated-artifact-policy-check` | Generated root policy | Required if generation/contracts touched | target summary | Fix owner input/policy, not generated file |
| `make json-shape-check` | JSON contract/manifests shape | Required if JSON touched | target summary | Fix authored JSON |
| `make frontend-typecheck` | TS type safety for generated contract or wire-shape changes | Required if authored contracts regenerate TS or API/row shape changes | target summary | Inspect TS errors; fix authored contracts or consumers, not generated files by hand |
| `make frontend-unit` | Frontend Timeline consumer behavior | Required if row/API semantics touched | phase summary | Fix consumer or contract |
| `make test-fast` | Aggregate local gate | Required before handoff after multi-slice refactor | `.cartulary/test-results/<run>/test-fast/target-summary.json` | Report failing target/run root |
| `make agent-finalize` | End-of-run hygiene | Required before broad verification | phase summary | Report retained-run status |
| `make check` | Full local gate with browser stack | Required before review for broad refactor | `.cartulary/test-results/<run>/check/target-summary.json` | Report scheduler child failure |

Initial planning session ran discovery only. S-01/S-02 implementation validation:

| Command | Result | Run root / artifact | Notes |
|---|---|---|---|
| `make backend-unit` | PASS | `.cartulary/test-results/20260629T215407Z-p574754/backend-unit/tool-run-summary.json` | 89 tests, failed 0. |
| `make phase-slice PHASE=phase9` | PASS | `.cartulary/test-results/20260629T215440Z-p578952/phase-slice/tool-run-summary.json` | 69 tests, failed 0; covers clipboard/bulk paths touched by S-02. |
| `make agent-finalize` | PASS | `.cartulary/test-results/20260629T215522Z-p589940/agent-finalize/tool-run-summary.json` | Generated artifacts unchanged; `RESULTS_DIR` unset. |
| `make phase-slice PHASE=phase3` | FAIL | `.cartulary/test-results/20260629T214915Z-p543942/phase-slice/tool-run-summary.json` | Failed only in `browser-e2e-visual`; three Timeline grid screenshot diffs. Backend/store/frontend unit/functional/measurement children passed in the run. |
| `make service-backed-slice PHASE=phase3` | FAIL | `.cartulary/test-results/20260629T215606Z-p590463/service-backed-slice/tool-run-summary.json` | Failed only in `browser-e2e-visual`; backend-store, backend-integration, backend-integration-support, browser webserver-backed, browser support, and browser measurement passed. |

## Handoff Planner

Workstream notes:

- Scope and evidence: keep this tracker tied to `internal/modules/timeline`; do not treat framework or prior handoffs as repository proof.
- Contracts and docs: Core 00-04 own behavior; `docs/domain.md` owns vocabulary; authored OpenAPI/view schema/SQL inputs are contract evidence; generated roots are downstream. Concrete closure now requires Core 01 evidence-count sort consistency and Core 03 auto-resolution/mention-effect wording before related implementation changes.
- Backend modules: S-01/S-02 moved inspected production Timeline mutation callers to the command facade. Next: close error mapping boundary, then hide store internals, then consolidate duplicated projection logic.
- Tests and harness: resolve BLOCKED mention row parity before moving projection or mention lifecycle code; add evidence-count sort and import no-auto-resolution characterization before changing those paths; resolve or owner-approve phase3 visual screenshot diffs before claiming full phase3 validation.
- Generated artifacts: update owner inputs and run generation/drift targets only when contract edits are required; never hand-edit generated outputs.
- Research support: R04 and R07 support low-friction Timeline capture, evidence-linked triage, and sortable time practice; R09 supports treating grid visual diffs as layout evidence. These are supporting resources only, not Core authority.
- Risks and blockers: `timeline.evidence_count` now has a concrete closure decision but still needs Core/schema/query/test edits; duplicated entity Timeline row projection still needs parity characterization; phase3 visual golden diffs require validation triage before phase3 can be claimed passing.

Handoff record template:

| Date/time | Branch/commit | Target module/seam | Current workflow | Completed workflows | Changed files | Commands run | Passing validation | Failing validation | Decisions made | Open questions | Blockers | Next recommended workflow | Safe restart command |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `2026-06-29T17:58:13-04:00` | `main` / `a6cd41d2dba44094bed875551633bd6a679eb54c` | Timeline facade/caller boundary | WF-09 | WF-04, WF-06 partial execution, WF-07 | `internal/modules/timeline/facade.go`; `internal/modules/timeline/routes.go`; `internal/modules/timeline/boundary_guard_test.go`; `internal/modules/workbook/routes.go`; `internal/modules/imports/routes.go`; this tracker | `gofmt -w ...`; `make backend-unit`; `make phase-slice PHASE=phase9`; `make agent-finalize`; `make phase-slice PHASE=phase3`; `make service-backed-slice PHASE=phase3`; focused `rg` scans | `make backend-unit`; `make phase-slice PHASE=phase9`; `make agent-finalize`; phase3 backend/store/functional/measurement children inside failed slice runs | `make phase-slice PHASE=phase3`; `make service-backed-slice PHASE=phase3`, both due to `browser-e2e-visual` Timeline grid screenshot diffs | Command facade is the production caller boundary for inspected Timeline mutations; legacy facade methods remain as one-slice compatibility shims; workbook decision supersede keeps current workbook hash path | Classify phase3 visual diffs before claiming phase3 pass. | Phase3 visual golden validation blocker; S-05 projection parity blocker; mention parity blocker | WF-03/S-03: read exact owner sections, then close Timeline error mapping boundary; separately triage visual golden diffs | `make task-guide ROLE=feature-dev PHASE=phase3` |
| `2026-06-29T18:16:51-04:00` | `main` / `d0afbad53dabfddd8688b6c9f2965f9532106045` | Timeline gap closure planning | WF-03/WF-05 | WF-10 tracker validation updated | this tracker | `git status --short --branch`; `date`; focused `rg`/`sed`; `make lint-markdown` | `make lint-markdown` | none | `timeline.evidence_count` is the chosen sortable closure; imported/machine-applied mentions do not run interactive auto-resolution; Timeline owns row wire presenter seam; visual diffs are validation blockers, not Core contradictions until triaged. | None for tracker update. | Core/schema/query/test edits still required for evidence-count closure; mention row parity still blocks S-05; phase3 visual diffs still block phase3 pass claim. | WF-03: implement Core/schema evidence-count closure, or S-03 error boundary if contracts remain untouched | `make task-guide ROLE=feature-dev PHASE=phase3` |

## Binary Acceptance Criteria

- RF-AC-001: Timeline external HTTP, OpenAPI, WebSocket, view-row, idempotency, and error contracts are unchanged unless an owner spec and authored contract update explicitly require it.
- RF-AC-002: Low-friction Timeline capture remains intact, including `client_txn_id`-only zero-field create where allowed and rough rows with partial/unstructured values.
- RF-AC-003: Rough capture validation does not require normalization, canonical host/identity records, mandatory owner/approver/task fields, or time completeness.
- RF-AC-004: Host/identity mention extraction preserves raw text and separate mention rows; no implicit stub/entity creation is introduced outside explicit owner-defined actions.
- RF-AC-005: Narrow interactive auto-resolution remains exact, unique, suppressible, and limited to Timeline relationship cells.
- RF-AC-006: `row_version`, `base_row_version`, same-field conflict token, stale conflict, replay, and no-effective-change semantics are preserved.
- RF-AC-007: `client_txn_id` idempotency remains route/actor/scope/request-hash scoped, with exact replay and divergent conflict behavior unchanged.
- RF-AC-008: Projection writes, query results, group values, changed field keys, and `record_changed` patch/invalidate behavior remain stable.
- RF-AC-009: Revisions, rollback coupling, supersede links, evidence attachment counts/flags, tags, and raw capture persist with the same source-of-truth semantics.
- RF-AC-010: Generated roots are not hand-edited; any required contract changes start from owner inputs and pass drift/policy checks.
- RF-AC-TL-011: Clipboard paste preserves ordered mutations, raw unmapped columns, cross-incident target rejection, paste-time batch conflicts, and collection action mapping.
- RF-AC-TL-012: Time-conversion profile behavior, profile-version conflicts, generated UTC/local pair flags, and activity sort fields remain unchanged.
- RF-AC-TL-013: Workbook/imports callers use the Timeline facade after refactor; private store and persistence details are not production caller contracts.
- RF-AC-TL-014: No runtime dependency on phase maps, phase ledgers, or harness accounting is introduced.
- RF-AC-TL-015: `timeline.evidence_count` is a read-only non-negative integer projection field, hidden by default, not groupable, not filterable in the current profile, and sortable only through authored Core/view-schema/query support.
- RF-AC-TL-016: Imported or machine-applied Timeline host/identity mention tokens do not run interactive auto-resolution and remain unresolved unless explicit resolved refs are supplied.
- RF-AC-TL-017: Timeline row wire shape has one Timeline-owned presenter/contract seam; entity mention lifecycle code does not hand-maintain a divergent row payload.
