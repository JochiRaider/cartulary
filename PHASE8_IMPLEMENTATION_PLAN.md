# Phase 8 Implementation Plan

## Summary

This file is the execution roadmap and progress marker for Cartulary Phase 8: links, tags, saved views, sorting, filtering, grouping, startup selection, and projection-backed query semantics.

`docs/guides/cartulary_implementation_testing_guide.md` section 10 is the controlling implementation-scope reference for this phase. Normative behavior remains owned by the core documents, especially Core 01 section 3.3.4, Core 01 section 3.3.5.2, Core 01 section 7.4, Core 01 section 8, Core 02 sections 11 and 12, Core 03 sections 2 and 14, and Core 04 section 2.

This planning artifact does not implement Phase 8 behavior. It is intentionally root-level so agents can find it quickly during handoff or interrupted implementation sessions. No README update is required for discoverability.

Current repo status after remediation on 2026-05-13: Phase 8 is listed as `active` in `tools/phase_registry.json`; `tools/phase8_test_map.json` and `docs/testing/phase8_coverage_ledger.md` exist; generated schedule artifacts have been regenerated through the canonical phase schedule target; all Sprint 0 placeholder tests have been replaced by direct row evidence or by an explicit product-gap sentinel where the saved-view browser lifecycle remains unimplemented. `AC-184` and `AC-185` now have direct Phase 8 evidence. `make phase-slice PHASE=phase8`, `make service-backed-slice PHASE=phase8`, and `make agent-finalize` pass after real Phase 8 service-backed duration evidence refreshed the advisory Go duration baselines. A full `make check` run remains blocked by a non-Phase-8 Phase 4 browser assessment ordering failure recorded under `.cartulary/test-results/20260513T015011Z-p36877`.

## Phase Objective

Phase 8 completes workbook configuration and projection-backed navigation. By phase exit, users must be able to manage typed record relationships and incident-scoped tags as structured relationship state, create and manage saved views with exact scope vocabulary, open workbooks through deterministic startup selection, run sort/filter/group queries through stable contract keys only, and page through projection-backed query results with live authorization and exact server-side query semantics.

Phase 8 also closes explicit Phase 7 handoff items: typed tag rollback, `AC-184`, and `AC-185` are Phase 8-owned. Typed tag rollback, `AC-184`, and `AC-185` now have direct Phase 8 evidence; the remaining product gap is saved-view browser lifecycle implementation.

## Implementation Scope

In scope:
- Typed record links and incident-scoped tags stored as structured rows, with closed base relationship vocabulary and query/history/projection consequences.
- `timeline.tags` collection action semantics, including `add_tag`, `remove_tag`, `record_tag:<record_id>:<tag_id>` item refs, normalization, de-duplication, and executable rollback for tag mutation entries.
- Saved-view object lifecycle: list, create, patch, delete, duplicate semantics, exact `private` / `shared` / `system` scope vocabulary, authorization consequences, no-op patch semantics, normalized `query_json`, and canonical `layout_json`.
- Separate workbook preference pointers: caller-owned `home_sheet_ref` and incident-wide `default_sheet_ref`.
- Workbook startup selection fallback: explicit launch pointer, caller home pointer, incident default pointer, then `cartulary.view.timeline.v1`, clearing invalid, hidden, missing, or unauthorized pointers before continuing.
- View-query validation and normalization for `sort[]`, `filters[]`, `group_by`, pagination, `meta.query`, saved-view query persistence, saved-view layout persistence, and cursor binding.
- Timeline grouping whitelist and presentation-only group headers.
- View-schema discovery of `sort_fields`, `header_sort_field_key`, grouping whitelists, null-last ordering, and no client-sortable `record_id`.
- Full-row and sparse-patch wire families preserving hidden writable fields, authoritative nulls, and canonical `changed_field_keys[]` plus `affected_views[]`.
- Browser-visible saved-view, startup, sorting, filtering, grouping, and exact search workflows.
- Phase 8 executable manifest, ledgers, generated schedules, focused tests, public phase slices, service-backed slices, and final gate evidence.

Out of scope unless an owner decision pulls it forward:
- Phase 9 keyboard, clipboard, Notes, Indicators, Parties, Assessments, Task Requests, Decisions, and coordination-surface workflow implementation. Phase 8 may only preserve compatibility through stable view-schema, query, row-patch, and saved-view contracts needed by those later surfaces.
- Treating support-only tests, generated files, visual goldens, or non-normative guide prose as behavior authority.
- Immutable `snapshot_stable` cursor continuation for live workbook routes. Phase 8 live routes must use live-authorized, tamper-proof, contract-stable keyset continuation per the normative core.
- New route families for saved-view duplication unless normative owner text is found. Default implementation choice: duplicate a visible view by using the ordinary saved-view create route with the visible source view's normalized query/layout copied into a new saved-view resource.
- Reworking Phase 7 delete/restore/rollback behavior outside the tag rollback closure required for Phase 8.

## Sprint Checklist

| Done | Sprint | Validation | Blockers | Follow-up Notes |
| --- | --- | --- | --- | --- |
| [x] | 0. Phase 8 ownership manifest and harness setup | [x] manifest, ledger, schedule drift, name check, target plan, and ID audit passed | None for the requested Sprint 0 gates. | Phase 8 is active; Sprint 0 placeholders have been replaced. |
| [x] | 1. Typed links, tags, and rollback handoff | [x] focused Sprint 1 checks passed; aggregate Phase 8 wrappers now pass after later remediation | None for this sprint. | `record_tag` rollback is executable and evidenced. `AC-184`/`AC-185` are evidenced by later Phase 8 query and Notes tests. |
| [x] | 2. Saved-view persistence and create defaults | [x] implemented | Remediated on 2026-05-14; storage/routes, create defaults, list visibility, OpenAPI create-input/resource-output split, route-level negative coverage, ledgers, schedules, and duration baselines are current. | Create/list foundation, private default, exact scope tokens, one `view_schema_id`, canonical query/layout persistence. |
| [x] | 3. Saved-view patch, duplicate, and delete | [x] implemented and audited | None for this sprint. | Patch/delete routes, stale-version conflict, structural no-op behavior, ordinary-create duplicate semantics, system immutability, and delete-only configuration removal are evidenced. |
| [ ] | 4. Workbook preferences and startup selection | [ ] planned | Requires saved-view visibility checks from Sprint 2/3. | Add saved-view sheet refs and fallback clearing behavior. |
| [ ] | 5. Query contract validation and normalization | [ ] planned | Existing query decoder must be audited against saved-view persistence and `meta.query`. | Stable keys only, ceilings, canonical filters/sort/group, duplicate rejection. |
| [ ] | 6. Projection-backed execution, search, and cursor semantics | [ ] planned | Requires Sprint 5 canonical query contract. | Live-authorized keyset cursor, exact-token `full_text`, strict `prefix`, null-last ordering. |
| [ ] | 7. Grouping, discovery, and grid controls | [ ] planned | Requires Sprint 5/6 query semantics. | Timeline whitelist, presentation-only headers, no client-sortable `record_id`, stable frontend keys. |
| [ ] | 8. Sparse patch, hidden writable fields, and browser workflows | [ ] planned | Requires backend query/saved-view/startup behaviors. | Complete browser-functional rows `E-8-01..E-8-04` and sparse-patch row `U-8-10`. |
| [ ] | 9. Phase gate, ledgers, schedules, baselines, and exit | [ ] planned | Depends on all authoritative rows having direct non-placeholder evidence. | Finalize generated artifacts, public wrappers, service-backed slices, and `make check` or recorded non-Phase-8 blocker. |

## Global References

- Controlling guide: `docs/guides/cartulary_implementation_testing_guide.md`, `Phase 8`, `U-8-01..U-8-10`, `U-8-GRID-01`, `I-8-01..I-8-04`, `E-8-01..E-8-04`.
- Phase-owned ACs: `AC-013..AC-014`, `AC-024..AC-026`, `AC-124`, `AC-127`, `AC-146..AC-153`, `AC-184..AC-185`, `AC-200`, `AC-206..AC-208`, `AC-210`, `AC-359..AC-368`, `AC-372..AC-375`, `AC-387`.
- Primary REQs: `REQ-00-027`, `REQ-01-022`, `REQ-01-034..REQ-01-047`, `REQ-01-138..REQ-01-151`, `REQ-01-267`, `REQ-01-286`, `REQ-01-310`, `REQ-01-312`, `REQ-01-323`, `REQ-01-326`, `REQ-01-328`, `REQ-01-329`, `REQ-01-331`, `REQ-01-332`, `REQ-01-336`, `REQ-01-339`, `REQ-01-351..REQ-01-353`, `REQ-01-499`, `REQ-01-503..REQ-01-506`, `REQ-01-554..REQ-01-560`, `REQ-01-565..REQ-01-567`, `REQ-02-010..REQ-02-011`, `REQ-02-163..REQ-02-176`, `REQ-03-012..REQ-03-032`, `REQ-03-097`, `REQ-03-223..REQ-03-235`, `REQ-03-247`.
- Normative owners:
  - Core 01: `docs/spec/01_architecture_storage_and_view_contracts.md`, especially view-shaped reads, saved-view routes, workbook-preference routes, view-schema registry, projection model, cursor continuation, and route query contracts.
  - Core 02: `docs/spec/02_domain_model_schema_and_history.md`, especially saved views, workbook preferences, typed relationships, tags, mutation targets, and history/projection consequences.
  - Core 03: `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, especially workbook surfaces, saved views, startup selection, sorting, filtering, grouping, and presentation-only grouping behavior.
  - Core 04: `docs/spec/04_security_deployment_and_conformance.md`, especially authorization ACs and Phase 8-owned acceptance criteria.
- Domain vocabulary reference: `docs/domain.md`. Use it for terms such as `saved_view`, `view_schema_id`, `sheet_ref`, `field_key`, `record_id`, `record_tags`, and `projection`; do not treat it as a behavior owner.
- Phase 7 handoff: `PHASE7_IMPLEMENTATION_PLAN.md`, `Handoff Requirements for Phase 8`.
- Generated boundaries: do not hand-edit `internal/gen/**` or `packages/protocol-ts/src/generated/**`; if SQL or contract generation is required, edit source contracts/queries and run canonical generation.

## Evidence Layer Matrix

| Row | Evidence layer | Phase 8 claim |
| --- | --- | --- |
| `U-8-01` | `backend_store` | Typed links and tags persist as structured rows and use the closed base relationship vocabulary. |
| `U-8-02` | `backend_store` | Saved-view create defaults omitted scope to `private`, rejects ordinary `system` create, and persists exactly one `view_schema_id`. |
| `U-8-03` | `backend_unit` | Scope vocabulary is exactly `private`, `shared`, `system`; obsolete `team` is rejected. |
| `U-8-04` | `backend_store` | Saved-view patch mutable-field contract, immutable rejection, and structurally valid no-op without version/timestamp change. |
| `U-8-05` | `backend_store` | `home_sheet_ref` and `default_sheet_ref` remain separate and startup fallback clears invalid/hidden pointers. |
| `U-8-06` | `backend_unit` | View-query contract accepts stable `field_key` values only and rejects invalid filters/sort/grouping shapes. |
| `U-8-07` | `backend_unit` | Timeline grouping whitelist and presentation-only group headers. |
| `U-8-08` | `backend_unit` | Sort/filter ceilings, canonical normalization, `meta.query`, default-tail expansion, and grouping omission semantics. |
| `U-8-09` | `backend_unit` | View-schema discovery exposes exact sorting/grouping contracts, null-last ordering, and no client-sortable `record_id`. |
| `U-8-10` | `backend_integration` | Full-row and sparse-patch families preserve hidden writable fields, nulls, and canonical change metadata. |
| `U-8-GRID-01` | `frontend_unit` | Grid sort/filter/group controls send only stable contract keys, never labels or implementation names. |
| `I-8-01` | `backend_integration` | Saved-view create/update/duplicate/delete persist normalized state and authorization consequences in the database. |
| `I-8-02` | `backend_integration` | Startup selection fallback behaves correctly across valid, missing, hidden, and invalid saved-view refs. |
| `I-8-03` | `backend_integration` | Link and tag mutations atomically update projections, history, and view-query results. |
| `I-8-04` | `backend_integration` | Cursor pagination re-derives live authorization, rejects tampering, and returns fetch-time payloads. |
| `E-8-01` | `browser_functional` | Browser saved-view lifecycle remains explicitly blocked until product saved-view routes and browser affordances exist. |
| `E-8-02` | `browser_functional` | Browser startup honors explicit sheet, home, default, and Timeline fallback while clearing invalid pointers. |
| `E-8-03` | `browser_functional` | Browser Timeline sort/filter/group reaches stable viewport and deterministic grouping with no writable header rows. |
| `E-8-04` | `browser_functional` | Browser exact-token `full_text` and strict `prefix` behavior do not degrade to fuzzy or relevance-ranked search. |

Support evidence may be listed in the manifest notes. Every authoritative row above must have direct non-placeholder evidence in its declared layer before Phase 8 exit; `E-8-01` is currently direct evidence of the remaining saved-view lifecycle blocker, not proof that the product lifecycle is complete.

## Dependencies and Prerequisites

- Phase 7 history reads, delete/restore, rollback, destructive locks, and retained-history behavior are available and must not be weakened.
- Tag rollback remains Phase 8-owned. Current Phase 7 history may show `record_tag` entries, but Phase 7 did not advertise them as executable rollback actions and returns `target_not_reversible` when addressed.
- `AC-184` and `AC-185` are Phase 8-owned support references for Phase 7; direct Phase 8 evidence now exists in the query contract and Notes full-text tests.
- Existing workbook query, view-schema registry, projection, collaboration, and row-patch paths are support substrate only until Phase 8 rows make direct assertions.
- Generated artifacts are downstream of owner-driven contract, SQL, and manifest changes.

## Public Interfaces and Deliverables

Expected route and schema deliverables:
- `GET /api/v1/incidents/{incident_id}/saved-views`
- `POST /api/v1/incidents/{incident_id}/saved-views`
- `PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}`
- `DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}`
- Existing workbook preference routes with Phase 8-complete `sheet_ref.kind='saved_view'` validation and clearing behavior.
- Existing `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` with Phase 8-complete query validation, `meta.query`, grouping, search, sort, and pagination semantics.
- OpenAPI schemas for saved-view resources, saved-view create/patch/delete payloads, saved-view list paging, canonical query/layout objects, and query response `meta.query`.

Expected persistence deliverables:
- Saved-view storage with `saved_view_id`, `incident_id`, `view_schema_id`, `scope`, `display_name`, `query_json`, `layout_json`, `owner_user_id`, timestamps, and `saved_view_version`.
- Structured link/tag persistence and mutation target detail sufficient for projection refresh, history, query readback, and tag rollback.
- Workbook preference storage that keeps `home_sheet_ref` and `default_sheet_ref` separate and nullable.
- Live-authorized cursor continuation that binds route contract and continuation position without snapshot-stable authorization caching.

Expected frontend deliverables:
- Saved-view lifecycle affordances sufficient for `E-8-01`; currently missing and tracked by an explicit browser-functional blocker sentinel.
- Startup selection behavior sufficient for `E-8-02`.
- Sort/filter/group controls that send stable contract keys only.
- Rendered grouping that never exposes group headers as editable rows or mutation targets.
- Search UI behavior that reflects server canonical `meta.query`.

Expected harness deliverables:
- `tools/phase8_test_map.json`
- `docs/testing/phase8_coverage_ledger.md`
- Generated service-backed and check schedule updates.
- Backend unit, backend store, backend integration, frontend unit, and browser-functional authoritative tests.

## Sprint 0. Phase 8 Ownership Manifest and Harness Setup

Objective: Establish Phase 8 test ownership before feature work so TDD rows can be selected by repo tooling.

Status: Complete for harness setup on 2026-05-13. Phase 8 is active and selectable; placeholder tests remain intentionally red and are not exit evidence.

Relevant IDs: all `U-8-*`, `U-8-GRID-01`, `I-8-*`, `E-8-*`; `tools/phase8_test_map.json`; `docs/testing/phase8_coverage_ledger.md`.

Evidence layers: `backend_unit`, `backend_store`, `backend_integration`, `frontend_unit`, `browser_functional`.

Grep references:
- `phase8`
- `U-8-`
- `I-8-`
- `E-8-`
- `U-8-GRID`
- `forbidden_id_files`
- `CARTULARY_MANIFEST_PHASE=phase8`

Files and areas:
- `tools/phase_registry.json`
- `tools/phase8_test_map.json`
- `docs/testing/phase8_coverage_ledger.md`
- `tools/scheduler_manifest.json`
- `tools/execution_topology_render_index.json`
- Backend placeholder tests in the owning modules selected by later sprints.
- Frontend placeholder test under `apps/web/src/WorkbookShell.phase8.query.test.tsx`.
- Browser placeholder test under `apps/web/e2e/phase8.workbook.spec.ts`.

Test-first sequence:
1. Add manifest rows for every `U-8-01..U-8-10`, `U-8-GRID-01`, `I-8-01..I-8-04`, and `E-8-01..E-8-04` row before product behavior implementation.
2. Declare one authoritative execution dependency per row and list any support-only tests as support, not row owners.
3. Add initial failing row symbols/titles only where the repo's manifest tooling requires discoverable tests before behavior exists; replace all placeholders before Phase 8 exit.
4. Add `forbidden_id_files` for known support-only carryover files from Phases 3 through 7 so earlier phase tests cannot accidentally claim Phase 8 IDs.
5. Generate ledger and schedule artifacts through canonical commands.

Implementation tasks:
- Create `tools/phase8_test_map.json` from the guide rows and the evidence layer matrix above.
- Keep `tools/phase_registry.json` status `planned` until selectable row symbols exist; move to `active` only when the phase wrappers can select Phase 8 rows.
- Generate `docs/testing/phase8_coverage_ledger.md`.
- Generate service-backed and check schedules after the manifest is active.
- Record Phase 7 handoff boundaries in manifest notes.

Implementation results:
- `tools/phase8_test_map.json` now owns every authoritative `U-8-01..U-8-10`, `U-8-GRID-01`, `I-8-01..I-8-04`, and `E-8-01..E-8-04` row with exactly one authoritative execution dependency per row.
- `forbidden_id_files` records Phase 3 through Phase 7 support-only carryover files so earlier phase tests cannot claim Phase 8 IDs.
- Manifest notes record the Phase 7 handoff boundaries for typed tag rollback, `AC-184`, `AC-185`, and the non-reversible `record_tag` behavior that Phase 8 has since closed with direct evidence.
- `tools/phase_registry.json` was moved to `active` only after row symbols and titles existed.
- `docs/testing/phase8_coverage_ledger.md` was generated with `make phase-ledgers`.
- `tools/scheduler_manifest.json` and `tools/execution_topology_render_index.json` were regenerated with `make phase-schedules`.
- Historical Sprint 0 placeholder owners were originally added at these paths and have since been replaced:
  - `internal/modules/links/phase8_links_tags_test.go` for `U-8-01` and `I-8-03`.
  - `internal/modules/savedviews/phase8_savedviews_test.go` for `U-8-02`, `U-8-03`, `U-8-04`, and `I-8-01`.
  - `internal/modules/incidents/phase8_workbook_startup_test.go` for `U-8-05` and `I-8-02`.
  - `internal/platform/viewquery/phase8_query_contract_test.go` for `U-8-06`, `U-8-07`, and `U-8-08`.
  - `internal/modules/viewschemas/phase8_viewschema_test.go` for `U-8-09`.
  - `internal/modules/workbook/phase8_row_wire_test.go` for `U-8-10` and `I-8-04`.
  - `apps/web/src/WorkbookShell.phase8.query.test.tsx` for `U-8-GRID-01`.
  - `apps/web/e2e/phase8.workbook.spec.ts` for `E-8-01..E-8-04`.
- The Sprint 0 placeholder failure message has been removed from these files. Any future placeholder reintroduction is a Phase 8 exit blocker.

Validation commands:
- `make phase-map-check`
- `make explain-phase PHASE=phase8`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-schedules`
- `make phase-schedule-drift`
- `make phase-test-name-check`
- `make target-plan-json`
- `git diff --check`

Sprint 0 validation results:
- Passed: `make phase-map-check`.
- Passed: `make explain-phase PHASE=phase8`; it reported `tools/phase8_test_map.json`, `docs/testing/phase8_coverage_ledger.md`, execution dependencies `backend_unit`, `backend_store`, `backend_integration`, `frontend_unit`, and `browser_functional`, service requirements, and target coverage for 19 authoritative rows.
- Passed after generation: `make phase-ledgers` and `make phase-ledger-drift`.
- Passed after generation: `make phase-schedules` and `make phase-schedule-drift`.
- Passed: `make phase-test-name-check`.
- Passed: `make target-plan-json`.
- Passed: `git diff --check`.
- Passed ID audit: every `U-8-*`, `I-8-*`, `E-8-*`, `U_8_*`, `I_8_*`, and `E_8_*` hit was in an approved Phase 8 placeholder, `tools/phase8_test_map.json`, generated Phase 8 ledger or scheduler metadata, the testing guide, or this implementation plan.
- Additional maintenance note: the earlier `make agent-finalize` duration-baseline failure was resolved by refreshing `tools/go_test_duration_baselines.json` from successful real Phase 8 service-backed evidence. Do not add synthetic baselines in future maintenance work.

Deliverables:
- `tools/phase8_test_map.json`
- `docs/testing/phase8_coverage_ledger.md`
- Generated schedule updates in `tools/scheduler_manifest.json` and `tools/execution_topology_render_index.json`
- Selectable Phase 8 row inventory with no accidental support-only ownership

Risks and open questions:
- Sprint 0 placeholder tests have been replaced; any new placeholder must be treated as temporary selection scaffolding and removed before exit.
- Service-backed duration baselines are grounded in successful real Phase 8 service-backed evidence, not synthetic placeholder timings.
- Manifest activation has changed public task-surface behavior; later sprints must keep the manifest coherent as placeholders are replaced by direct evidence.

Exit criteria:
- Met for Sprint 0: `make explain-phase PHASE=phase8` reports the manifest, ledger path, execution dependencies, service requirements, and target coverage.
- Met for Sprint 0: `make phase-ledger-drift` and `make phase-schedule-drift` pass after generation.
- Met for Sprint 0: every Phase 8 authoritative row ID appears only in approved Phase 8 tests, manifest metadata, generated Phase 8 metadata, the testing guide, or this implementation plan.

## Sprint 1. Typed Links, Tags, and Rollback Handoff

Objective: Complete typed link and incident-scoped tag semantics as structured relationship state, including projection/history/query consequences and executable tag rollback.

Status: Implemented for Sprint 1 on 2026-05-13. Focused `U-8-01` and `I-8-03` evidence passes; later remediation replaced the remaining Sprint 0 placeholders and Phase 8 aggregate wrappers now pass.

Relevant IDs:
- `U-8-01`, `I-8-03`
- Handoff closure: `AC-184`, `AC-185`
- Related Phase 8 ACs: `AC-200`, `AC-205..AC-210`
- Primary REQs: `REQ-02-011`, `REQ-02-163..REQ-02-176`, `REQ-01-351..REQ-01-353`

Evidence layers: `backend_store` for `U-8-01`; `backend_integration` for `I-8-03`.

Grep references:
- `record_links`
- `record_tags`
- `record_tag:<record_id>:<tag_id>`
- `timeline.tags`
- `add_tag`
- `remove_tag`
- `change_set_mutations`
- `target_not_reversible`
- `available_rollback_actions`
- `refresh`

Files and areas:
- `db/migrations/**`
- `db/queries/**`
- `internal/modules/links`
- `internal/modules/timeline`
- `internal/modules/revisions`
- `internal/modules/workbook`
- `internal/modules/projections`
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/errors/index.json`

Test-first sequence:
1. `U-8-01` asserts active typed links and tags are durable structured rows, not JSON-only cells, and use the closed base relationship vocabulary.
2. `U-8-01` asserts `timeline.tags` accepts `add_tag` and `remove_tag`, rejects obsolete action names, normalizes labels, coalesces duplicate adds, rejects normalized-empty labels, and serializes item refs as `record_tag:<record_id>:<tag_id>`.
3. `U-8-01` asserts tag mutation entries have deterministic target identity and enough before/after detail to support rollback.
4. `I-8-03` asserts link and tag mutations update source rows, projections, history, workbook query results, and emitted change metadata in one committed transaction.
5. Phase 7 handoff tests assert current `record_tag` history entries advertise executable rollback only after Sprint 1 support lands, and rollback of tag add/remove creates a new attributed `change_set` without rewriting prior history.

Implementation tasks:
- Audit existing `record_links` and `record_tags` migrations against Core 02 section 12 and add only owner-driven migrations if required.
- Route `timeline.tags` through structured tag operations, not JSON-only cell replacement.
- Persist tag add/remove mutation entries with deterministic `record_tag:<record_id>:<tag_id>` target identity.
- Make tag rollback executable through the Phase 7 rollback route using new Phase 8 tag reversal logic.
- Refresh affected projections and publish ordinary workbook invalidation/change events.
- Keep whole-row restore behavior unchanged: it must not implicitly recreate or delete relationship-like adjuncts.

Implementation results:
- Replaced `internal/modules/links/phase8_links_tags_test.go` placeholders with direct `U-8-01` backend-store evidence and direct `I-8-03` backend-integration evidence.
- Added shared `fieldnorm.NormalizeTagLabel`, enforcing trim, NFC, control rejection, normalized-empty rejection, a 64 Unicode-scalar ceiling, and locale-independent folded dedupe keys.
- Updated `timeline.tags` and `note.tags` mutation decoding to use `add_tag { tag_name }` and `remove_tag { item_ref }`; obsolete `add_token` is rejected for tag fields and remains available for mention-token fields.
- Persisted tags as structured `record_tags` rows and emitted tag collection items as `item_ref="record_tag:<record_id>:<tag_id>"`, `item_kind="tag"`, `display_text=<tag_name>`, and `tag_id=<record_tag_id>`.
- Persisted tag add/remove mutations as `change_set_mutations.target_kind='record_tag'` with composite target IDs and full before/after tag detail.
- Extended history selection, rollback addressability, rollback changed-field mapping, and direct history-entry rollback for `record_tag`; rollback now appends a new attributed `source='rollback'` change set and does not rewrite prior history.
- Added migration `00013_phase8_relationship_vocabulary.sql` to enforce the closed Core 02 base `record_links.link_type` vocabulary.
- Updated OpenAPI `CollectionActionsV1` with `CollectionAddTagAction` and `CollectionRemoveTagAction`, then regenerated downstream contract artifacts.
- Updated frontend tag mutation payload builders and support tests to emit `add_tag` / `remove_tag`.

Files changed:
- `db/migrations/00013_phase8_relationship_vocabulary.sql`
- `contracts/openapi/cartulary.openapi.yaml`
- `internal/gen/contracts/contracts_gen.go`
- `packages/protocol-ts/src/generated/contracts.ts`
- `internal/platform/fieldnorm/text.go`
- `internal/modules/links/phase8_links_tags_test.go`
- `internal/modules/timeline/api.go`
- `internal/modules/timeline/store.go`
- `internal/modules/timeline/phase3_decoder_test.go`
- `internal/modules/entities/timeline_projection.go`
- `internal/modules/entities/openapi_contract_test.go`
- `internal/modules/revisions/store.go`
- `internal/modules/revisions/rollback_store.go`
- `internal/modules/workbook/store.go`
- `internal/modules/workbook/mutation_api.go`
- `internal/modules/workbook/mutation_store.go`
- `internal/modules/workbook/phase6_conflict_support_test.go`
- `internal/modules/workbook/workbook_mutation_integration_test.go`
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/src/WorkbookShell.surfaces.test.tsx`

Validation commands:
- `go test ./internal/modules/timeline ./internal/modules/links ./internal/modules/revisions -run 'TestPhase8_.*(U_8_01|I_8_03)'`
- `go test ./internal/modules/workbook ./internal/modules/projections -run 'TestPhase8_.*I_8_03'`
- `make backend-store CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=unit CARTULARY_MANIFEST_COVERAGE=authoritative CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=backend_store`
- `make backend-integration CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=integration CARTULARY_MANIFEST_COVERAGE=authoritative`
- `git diff --check`

Validation results:
- Red check before implementation: `go test ./internal/modules/links -run 'TestPhase8_.*(U_8_01|I_8_03)'` failed as expected on closed vocabulary, `add_tag`, obsolete `add_token`, and integration behavior.
- Passed: `go test ./internal/modules/links -run 'TestPhase8_.*(U_8_01|I_8_03)' -count=1`.
- Passed: `go test ./internal/modules/timeline ./internal/modules/links ./internal/modules/revisions -run 'TestPhase8_.*(U_8_01|I_8_03)' -count=1`.
- Passed: `go test ./internal/modules/workbook ./internal/modules/projections ./internal/modules/entities -run 'TestPhase8_.*I_8_03|Test.*OpenAPI|Test.*Collection' -count=1`.
- Historical failure outside Sprint 1: `go test ./internal/modules/timeline ./internal/modules/workbook ./internal/modules/links ./internal/modules/revisions -count=1`; `timeline`, `links`, and `revisions` passed, and `workbook` failed only on then-unimplemented Phase 8 rows `U-8-10` and `I-8-04`. These rows have since been replaced with direct tests.
- Passed: `git diff --check`.
- Passed: `make generate`; retained root `.cartulary/test-results/20260513T004626Z-p4296`.
- Passed: `make generate-drift`; retained root `.cartulary/test-results/20260513T004805Z-p6339`.
- Passed: `make migration-drift`; retained root `.cartulary/test-results/20260513T004811Z-p7066`.
- Passed: `make phase-ledger-drift`; retained root `.cartulary/test-results/20260513T004822Z-p8382`.
- Passed: `make phase-schedule-drift`; retained root `.cartulary/test-results/20260513T004822Z-p8631`.
- Passed: `make frontend-typecheck`; retained root `.cartulary/test-results/20260513T004825Z-p8905`.
- Historical failure outside Sprint 1: `make frontend-unit`; retained root `.cartulary/test-results/20260513T004830Z-p9223`; blocker `U-8-GRID-01` has since been replaced with frontend unit evidence.
- Historical failure outside Sprint 1 after `U-8-01` progressed: `make backend-store CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=unit CARTULARY_MANIFEST_COVERAGE=authoritative CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=backend_store`; retained root `.cartulary/test-results/20260513T005235Z-p15799`; blocker `U-8-10` has since moved to backend integration evidence.
- Historical failure outside Sprint 1 after `I-8-03` progressed: `make backend-integration CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=integration CARTULARY_MANIFEST_COVERAGE=authoritative`; retained root `.cartulary/test-results/20260513T005250Z-p17645`; blocker `I-8-04` has since been replaced with cursor and Notes full-text tests.
- Historical failure outside Sprint 1: `make phase-slice PHASE=phase8`; retained root `.cartulary/test-results/20260513T005312Z-p20519`; later remediation now passes this wrapper.
- Historical failure outside Sprint 1: `make service-backed-slice PHASE=phase8`; retained root `.cartulary/test-results/20260513T005331Z-p25339`; later remediation now passes this wrapper.
- Failed maintenance preflight: `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260513T004805Z-p6339`; retained root `.cartulary/test-results/20260513T005352Z-p28963`; the retained run was not a passing warm `check` run.
- Historical maintenance coverage failure: `make agent-finalize`; retained root `.cartulary/test-results/20260513T005409Z-p29440`; structure ledger refresh was unchanged, retained-run maintenance was skipped because `RESULTS_DIR` was unset, then `go-test-duration-baseline-coverage` failed because Phase 8 service-backed duration baselines were missing. This has since been resolved from real service-backed evidence.

Generated artifacts:
- Changed by `make generate`: `internal/gen/contracts/contracts_gen.go` and `packages/protocol-ts/src/generated/contracts.ts`.
- `make generate-drift` passed after generation.

Deliverables:
- Direct `U-8-01` and `I-8-03` evidence.
- Executable `record_tag` rollback and clear Phase 7 handoff closure.
- Updated API/schema/error docs only where owner-driven behavior changed.

Phase 7 handoff closure:
- Closed for typed tag rollback: `record_tag` history entries are addressable, advertise executable history-entry rollback, and rollback through the existing route appends a new rollback change set while preserving original history and mutation rows.
- Preserved outside the specific Phase 8 tag rollback closure: Phase 7 delete, restore, rollback, retained-history, and whole-row restore behavior were not intentionally reworked.
- `AC-184` and `AC-185` now have direct Phase 8 evidence in the query contract and Notes full-text tests.

Risks and open questions:
- Existing tag behavior may be partially implemented with `add_token`; Sprint 1 must not preserve obsolete public action names if the normative action vocabulary requires `add_tag` and `remove_tag`.
- Relationship changes must be atomic with history/projection refresh; a projection-only assertion is insufficient evidence.

Exit criteria:
- Met for Sprint 1 focused evidence: `record_tag` no longer remains a Phase 7 non-claim.
- Met for Sprint 1 focused evidence: link/tag query readback, history, rollback, projections, workbook query results, and record-change metadata agree in the direct integration test.
- Later remediation achieved full Phase 8 aggregate wrapper passes; remaining work is the saved-view lifecycle product gap tracked by `E-8-01` and the non-Phase-8 `make check` blocker.

## Sprint 2. Saved-View Persistence and Create Defaults

Objective: Add saved-view list/create persistence foundation with exact scope vocabulary, private default, ordinary `system` create rejection, one `view_schema_id`, and canonical stored query/layout state.

Status: Implemented and remediated on 2026-05-14. Product saved-view storage, SQL, public list/create routes, scope vocabulary, create defaults, canonical query/layout persistence, OpenAPI create-input/resource-output split, and route-level negative coverage now have direct evidence. The 2026-05-14 follow-up closed the contract defect where OpenAPI required pre-normalized `query_json.sort`, `query_json.filters`, and full `layout_json` even though the create route accepts omitted/defaulted input.

Relevant IDs:
- `U-8-02`, `U-8-03`
- Foundation for `I-8-01`
- `REQ-03-012..REQ-03-023`, `REQ-01-138..REQ-01-151`
- `AC-146..AC-152`

Evidence layers: `backend_store` for `U-8-02`; `backend_unit` for `U-8-03`; `backend_integration` support for `I-8-01`.

Remediation evidence added on 2026-05-14:
- `TestPhase8_SavedViewCreateDefaults_U_8_02` now explicitly covers `query_json:{}` normalization to `sort:[]` and `filters:[]`, omitted `group_by`, omitted `layout_json` normalization to the schema default, `group_by:null` rejection, and `record_id` / `row_version` rejection in both `query_json` and `layout_json`.
- `TestPhase8_SavedViewOpenAPICreateInputIsLenient_U_8_02` asserts `SavedViewCreateRequest` uses lenient create-input schemas while `SavedViewResource` continues to use strict canonical persisted schemas.
- `tools/phase8_test_map.json`, `docs/testing/phase8_coverage_ledger.md`, generated schedule artifacts, and `tools/go_test_duration_baselines.json` were refreshed after adding the new authoritative U-8-02 test.

Grep references:
- `saved-views`
- `saved_view_id`
- `saved_view_version`
- `view_schema_id`
- `query_json`
- `layout_json`
- `scope`
- `system`
- `team`
- `cartulary.layout.v1`

Files and areas:
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/errors/index.json`
- `db/migrations/**`
- `db/queries/**`
- Saved-view route/service module under `internal/modules/savedviews`
- `internal/app/runtime.go`
- `internal/platform/viewquery`
- `internal/platform/viewschema`
- Generated SQL and protocol outputs via `make generate`

Test-first sequence:
1. Completed: `U-8-03` unit tests assert only `private`, `shared`, and `system` parse; `team` and unknown/null scope tokens fail closed.
2. Completed: `U-8-02` store/route tests assert omitted create scope persists as `private`, explicit `scope='system'` is rejected on the ordinary public create route, and each saved view binds exactly one `view_schema_id`.
3. Completed: Create tests assert `query_json` and `layout_json` normalize through the same stable field-key grammar as workbook query/view-schema contracts.
4. Completed: List tests assert visible saved views only, deterministic `updated_at desc, saved_view_id asc` ordering, and bound cursor paging.
5. Completed: OpenAPI regression tests assert create input accepts omitted `sort` / `filters` and omitted or `{}` `layout_json`, while resource output remains canonical and complete.

Implementation tasks:
- Completed: Saved-view storage and SQL queries exist for create/list.
- Completed: `GET` and `POST /api/v1/incidents/{incident_id}/saved-views` are registered and covered.
- Completed: `display_name`, `scope`, `view_schema_id`, `query_json`, and `layout_json` validate under owner contracts.
- Completed: Omitted `query_json.sort` and `query_json.filters` normalize to `[]`; inactive `group_by` is omitted; `group_by=null` is rejected.
- Completed: Omitted or `{}` `layout_json` normalizes to the schema-derived `cartulary.layout.v1` default.
- Completed: `record_id` and `row_version` are rejected anywhere inside saved-view `query_json` or `layout_json`.
- Completed: OpenAPI now distinguishes lenient create-input schemas (`SavedViewCreateQueryJSON`, `SavedViewCreateLayoutJSON`) from canonical persisted resource schemas (`SavedViewQueryJSON`, `SavedViewLayoutJSON`).

Validation commands:
Passed on 2026-05-14:
- `go test ./internal/modules/savedviews ./internal/platform/viewquery ./internal/platform/viewschema -run 'TestPhase8_.*(U_8_02|U_8_03|U_8_06|U_8_08)'`
- `go test ./internal/modules/entities ./internal/modules/savedviews -run 'Test.*OpenAPI|TestPhase8_SavedViewCreateDefaults_U_8_02'`
- `make generate`
- `make generate-drift`
- `make backend-store CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=backend_store`
- `make phase-slice PHASE=phase8`
- `make service-backed-slice PHASE=phase8`
- `make go-test-duration-baseline-coverage`
- `make phase-ledger-drift`
- `make phase-schedule-drift`
- `git diff --check`
- `make agent-finalize`

Deliverables:
- Delivered: Saved-view storage and create/list route foundation.
- Delivered: Core 03 saved-view normalization clarification.
- Delivered: OpenAPI create-input/resource-output schema split.
- Delivered: Route-level negative tests and empty-query canonical persistence tests.
- Delivered: Generated Go and TypeScript contract artifacts refreshed from source artifacts.
- Delivered: Phase 8 manifest, coverage ledger, schedule artifacts, and Go duration baselines refreshed for the added U-8-02 evidence.

Risks and open questions:
- Closed: `internal/modules/savedviews` is the module owner for saved-view list/create.
- Preserved: `system` saved views may still need seed/admin-only creation outside ordinary create in a later sprint; no such path was added because owner text does not require it here.
- Maintenance note: `agent-finalize` initially exposed the new OpenAPI regression test as missing from `tools/phase8_test_map.json` and then as missing duration baseline coverage. Both were remediated by adding the test to the U-8-02 row and refreshing duration baselines from a successful Phase 8 service-backed run. A pruned baseline refresh was refused because the retained run was partial; the non-pruned refresh succeeded and baseline coverage now passes.

Exit criteria:
- Met: Saved-view resources persist normalized canonical state and enforce exact scope vocabulary.
- Met: The ordinary create route cannot create `system` views.
- Met: OpenAPI create contracts match route behavior for omitted `query_json.sort`, omitted `query_json.filters`, and omitted or `{}` `layout_json`.
- Met: Route-level tests reject `group_by:null` and forbidden technical fields in both saved-view persisted blobs.

## Sprint 3. Saved-View Patch, Duplicate, and Delete

Objective: Complete mutable saved-view lifecycle behavior: patch, no-op semantics, immutable rejection, duplicate semantics, delete, and authorization consequences.

Status: Implemented and audited on 2026-05-15. Product saved-view patch and delete routes are wired through route, service, store, SQL, generated SQL, OpenAPI, and error contracts. `U-8-04` and `I-8-01` have direct, non-placeholder evidence in their declared layers. Browser saved-view lifecycle behavior remains Sprint 8 `E-8-01` scope and was not counted as Sprint 3 completion evidence.

Relevant IDs:
- `U-8-04`, `I-8-01`
- Browser follow-up: `E-8-01`
- `REQ-03-024..REQ-03-026`, `REQ-01-138..REQ-01-151`
- `AC-146..AC-152`

Evidence layers: `backend_store` for `U-8-04`; `backend_integration` for `I-8-01`; `browser_functional` support in Sprint 8 for `E-8-01`.

Grep references:
- `PATCH /api/v1/incidents/{incident_id}/saved-views`
- `DELETE /api/v1/incidents/{incident_id}/saved-views`
- `base_saved_view_version`
- `saved_view_version`
- `updated_at`
- `owner_user_id`
- `private`
- `shared`
- `system`
- `structural equality`

Files and areas:
- Saved-view route/service/store module
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/errors/index.json`
- `db/queries/**`
- `internal/platform/viewquery`
- `internal/platform/viewschema`
- Integration tests for saved-view authorization and persistence

Test-first sequence:
1. `U-8-04` asserts patch accepts `base_saved_view_version` plus mutable fields only: `display_name`, `query_json`, `layout_json`, and permitted `scope`.
2. `U-8-04` asserts attempts to mutate `incident_id`, `saved_view_id`, `view_schema_id`, owner fields, timestamps, or unknown members fail with `invalid_mutation_payload`.
3. `U-8-04` asserts structurally valid no-op patch compares normalized values, returns the current resource, and leaves `saved_view_version` and `updated_at` unchanged.
4. `I-8-01` asserts create, update, duplicate, and delete persist normalized `query_json` / `layout_json`, scope rules, and authorization consequences against a real database.
5. `I-8-01` asserts delete removes only the saved-view configuration object and never underlying records, links, tags, evidence, or workbook rows.

Implementation tasks:
- Completed: added saved-view patch and delete routes at `/api/v1/incidents/{incident_id}/saved-views/{saved_view_id}`.
- Completed: implemented stale `base_saved_view_version` conflict handling with `saved_view_version_conflict`.
- Completed: implemented no-op detection after normalized structural equality, not textual JSON comparison.
- Completed: enforced patch mutable-field limits, immutable/server-managed field rejection, ownership/role authorization, and `system` saved-view immutability.
- Completed: implemented duplicate-visible-view semantics by ordinary create from a visible saved-view resource's normalized `view_schema_id`, `query_json`, and `layout_json`; no narrower duplicate route was required by normative owner text.
- Deferred to Sprint 4 evidence: deleted or hidden saved views become invalid for startup preference resolution; Sprint 3 depends on that behavior but does not claim startup completion.

Implementation results:
- `PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}` accepts required `base_saved_view_version` plus optional `display_name`, `query_json`, `layout_json`, and `scope` limited to ordinary `private`/`shared` values.
- Patch rejects attempted mutation of `incident_id`, `saved_view_id`, `view_schema_id`, `owner_user_id`, `created_at`, `updated_at`, `saved_view_version`, and unknown members with `invalid_mutation_payload`.
- Structurally valid no-op patch returns the current saved-view resource and does not advance `saved_view_version` or `updated_at`.
- Non-owner incident members may see shared and system saved views but cannot mutate or delete them in place; incident admins can mutate/delete ordinary private/shared saved views; `system` saved views remain immutable through ordinary write paths.
- Duplicate from a visible shared or system saved view is represented as ordinary create of a new caller-owned saved view using normalized source query/layout state. Duplicating a `system` saved view does not mutate the `system` row in place.
- Delete removes the `saved_views` configuration row only and preserves underlying records, links, tags, evidence, projections, workbook rows, and workbook preference rows.

Validation commands:
- `go test ./internal/modules/savedviews -run 'TestPhase8_.*(U_8_04|I_8_01)'`
- `make backend-integration CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=integration CARTULARY_MANIFEST_COVERAGE=authoritative`
- `make generate`
- `make generate-drift`
- `git diff --check`

Sprint 3 validation results:
- Passed: `go test ./internal/modules/savedviews -run 'TestPhase8_.*(U_8_04|I_8_01)'`.
- Passed: `make backend-integration CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=integration CARTULARY_MANIFEST_COVERAGE=authoritative`; latest audit run reported 117 tests, 0 failed, run root `.cartulary/test-results/20260515T005016Z-p1215826`.
- Passed: `make generate`; run root `.cartulary/test-results/20260515T005547Z-p1224913`.
- Passed: `make generate-drift`; run root `.cartulary/test-results/20260515T005557Z-p1225682`.
- Passed: `git diff --check`; the audit also checked staged whitespace with `git diff --cached --check`.

Deliverables:
- Patch/delete lifecycle routes.
- Duplicate-visible-view path via ordinary create semantics.
- Direct service-backed saved-view lifecycle evidence.

Risks and open questions:
- Closed: authorization distinguishes private owner visibility, incident admin visibility, shared visibility, non-owner mutation denial, and `system` immutability.
- Closed: duplicate-visible-view behavior uses ordinary create and does not allow in-place mutation of `system` saved views.
- Preserved scope boundary: `E-8-01` may receive browser support later in Sprint 8, but browser lifecycle behavior is not Sprint 3 completion evidence.

Exit criteria:
- Met: `I-8-01` passes against a real database and proves create/update/duplicate/delete consequences.
- Met: saved-view patch no-op cannot advance version or timestamp.

## Sprint 4. Workbook Preferences and Startup Selection

Objective: Complete separate workbook preference pointers, saved-view sheet refs, and deterministic workbook startup fallback.

Status: Partially implemented by remediation on 2026-05-13. `U-8-10`, `E-8-02`, `E-8-03`, and `E-8-04` have direct evidence. `E-8-01` remains an explicit browser-functional blocker sentinel because in-product saved-view lifecycle affordances are still missing; Sprint 3 backend saved-view lifecycle routes now exist and are evidenced separately.

Relevant IDs:
- `U-8-05`, `I-8-02`
- Browser follow-up: `E-8-02`
- `REQ-03-027..REQ-03-032`, `REQ-01-145..REQ-01-151`
- `AC-150`, `AC-153`

Evidence layers: `backend_store` for `U-8-05`; `backend_integration` for `I-8-02`; `browser_functional` support in Sprint 8 for `E-8-02`.

Grep references:
- `home_sheet_ref`
- `default_sheet_ref`
- `sheet_ref`
- `kind='saved_view'`
- `kind='view_schema'`
- `workbook-preferences/me`
- `workbook-preferences/default`
- `unsupported_sheet_ref`
- `startup`

Files and areas:
- `internal/modules/incidents`
- Saved-view visibility helpers from Sprint 2/3
- Workbook open/startup route or shell bootstrap path
- `contracts/openapi/cartulary.openapi.yaml`
- `apps/web/src/WorkbookShell.tsx`
- Frontend startup/sheet-ref state tests

Test-first sequence:
1. `U-8-05` asserts `home_sheet_ref` and `default_sheet_ref` remain separate objects and no route collapses one into the other.
2. `U-8-05` asserts `PUT /workbook-preferences/me` mutates only caller `home_sheet_ref` and valid no-op preserves `updated_at`.
3. `U-8-05` asserts `PUT /workbook-preferences/default` mutates only incident `default_sheet_ref`, is admin-only, and valid no-op preserves `updated_at` and `updated_by_user_id`.
4. `I-8-02` asserts startup fallback order across explicit, home, default, and Timeline for valid, missing, hidden, deleted, and invalid saved-view references.
5. `I-8-02` asserts invalid/hidden pointers are cleared before fallback continues.

Implementation tasks:
- Extend `sheet_ref` validation to support `kind='saved_view'` for saved-view resources, while keeping base surfaces as `kind='view_schema'`.
- Use saved-view visibility and view-schema availability to validate startup pointers.
- Implement fallback clearing transactionally so invalid pointers do not remain sticky after successful fallback.
- Keep ordinary preference route no-op timestamp behavior.
- Surface startup selection state to the frontend without treating saved-view absence as base-surface absence.

Validation commands:
- `go test ./internal/modules/incidents ./internal/modules/savedviews -run 'TestPhase8_.*(U_8_05|I_8_02)'`
- `make backend-store CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=backend_store`
- `make backend-integration CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=integration`
- `make frontend-unit`
- `git diff --check`

Deliverables:
- Saved-view-capable sheet refs.
- Startup fallback and invalid-pointer clearing.
- Backend preference and startup evidence.

Risks and open questions:
- Startup fallback is visible in browser workflows, but backend store/integration tests must own the behavior first.
- Phase 9 base surfaces may be directly addressable as `view_schema` refs later; do not require saved-view resources for base surfaces.

Exit criteria:
- Workbook open never fails solely because a saved-view pointer became hidden, deleted, or invalid; it clears and falls through deterministically.
- `home_sheet_ref` and `default_sheet_ref` semantics remain separate.

## Sprint 5. Query Contract Validation and Normalization

Objective: Complete route-owned query validation and canonical normalization for sort, filter, grouping, saved-view persistence, and response `meta.query`.

Status: Remediation evidence collected on 2026-05-13. Phase ledgers and schedules were regenerated; Phase 8 public wrappers pass; `make agent-finalize` passes after real service-backed baseline refresh. Full `make check` remains blocked by a non-Phase-8 Phase 4 browser assessment ordering failure in `.cartulary/test-results/20260513T015011Z-p36877`.

Relevant IDs:
- `U-8-06`, `U-8-08`
- Supports `U-8-09`, `I-8-04`, `E-8-03`, `E-8-04`
- `REQ-01-035..REQ-01-047`, `REQ-02-010`, `REQ-03-223..REQ-03-233`
- `AC-024..AC-026`, `AC-124`, `AC-184`, `AC-243`, `AC-359..AC-361`

Evidence layers: `backend_unit` for `U-8-06` and `U-8-08`; service-backed support in later sprints.

Grep references:
- `Decode`
- `invalid_view_query`
- `meta.query`
- `sort_count_exceeded`
- `filter_count_exceeded`
- `duplicate field_key`
- `group_by`
- `full_text`
- `prefix`
- `record_id`
- `row_version`

Files and areas:
- `internal/platform/viewquery`
- `internal/modules/workbook`
- `internal/platform/viewschema`
- Saved-view query/layout validators
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/view-schemas/**`
- `apps/web/src/workbookQuery.ts`

Test-first sequence:
1. `U-8-06` asserts only stable `field_key` values are accepted for `filters[]`, `sort[]`, and `group_by`.
2. `U-8-06` asserts invalid operators, invalid argument shapes, duplicate filter keys after normalization, duplicate sort keys, disallowed sort keys, disallowed grouping keys, and malformed `group_by` fail with `invalid_view_query`.
3. `U-8-08` asserts route sort ceiling `8`, filter ceiling `16`, canonical filter ordering, canonical set values, canonical `prefix` value, canonical `full_text` token query, and no truncation.
4. `U-8-08` asserts successful query responses always include `meta.query.sort`, `meta.query.filters`, and omit `meta.query.group_by` when grouping is inactive.
5. `U-8-08` asserts omitted sort means no user sort override, while response `meta.query.sort` includes user override, default-tail expansion, and final `record_id asc` only as server-applied tie-breaker.

Implementation tasks:
- Audit and tighten `internal/platform/viewquery` against Core 01 query rules.
- Remove client-supplied `record_id` as a sortable field while preserving server-added final tie-breaker in effective `meta.query.sort`.
- Bind saved-view `query_json` validation to the same canonical user-query normalizer, while storing only user overrides and not applied effective sort.
- Add OpenAPI schemas for `meta.query` and reason-code details where absent.
- Ensure `Group: None` is represented by omission, never JSON `null`.

Validation commands:
- `go test ./internal/platform/viewquery ./internal/modules/workbook ./internal/modules/savedviews -run 'TestPhase8_.*(U_8_06|U_8_08)'`
- `make backend-unit CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=unit CARTULARY_MANIFEST_COVERAGE=authoritative`
- `make generate`
- `make generate-drift`
- `git diff --check`

Deliverables:
- Canonical query normalizer/validator.
- Query response `meta.query` contract.
- Saved-view query persistence shares the route-owned validator without storing applied default tails.

Risks and open questions:
- Existing tests may allow `record_id` sort as client input; Phase 8 must treat that as non-conformant and preserve only server-applied final tie-breaker behavior.
- Saved-view persistence and route response normalization are similar but not identical; do not store `meta.query.sort` as saved-view `query_json.sort`.

Exit criteria:
- `U-8-06` and `U-8-08` pass in backend unit selection and are not dependent on browser behavior.
- Equivalent normalized filter contracts bind equivalent cursor contracts.

## Sprint 6. Projection-Backed Execution, Search, and Cursor Semantics

Objective: Make projection-backed query execution conform to canonical sort/filter/search semantics and live-authorized cursor continuation.

Status: Planned.

Relevant IDs:
- `I-8-04`, `E-8-04`
- Supports `U-8-08`, `U-8-09`
- `REQ-01-035..REQ-01-036`, `REQ-01-554..REQ-01-560`, `REQ-01-565..REQ-01-567`
- `AC-372..AC-375`, `AC-387`

Evidence layers: `backend_integration` for `I-8-04`; `browser_functional` in Sprint 8 for `E-8-04`.

Grep references:
- `live_authorized_keyset`
- `snapshot_stable`
- `cursor`
- `tamper`
- `full_text`
- `prefix`
- `null-last`
- `fetch-time`
- `authorization`
- `route contract`

Files and areas:
- `internal/platform/pagination`
- `internal/platform/viewquery`
- `internal/modules/workbook`
- Projection query SQL under `db/queries/**`
- `contracts/openapi/cartulary.openapi.yaml`
- Browser search specs in `apps/web/e2e`

Test-first sequence:
1. `I-8-04` asserts cursor continuation re-derives current session, route authorization, incident membership, and route visibility before returning rows.
2. `I-8-04` asserts cursor tokens are opaque and reject tampering or replay against a different bound route/query contract.
3. `I-8-04` asserts continuation returns fetch-time row payloads and fresh first-page queries evaluate current live state from the beginning.
4. `E-8-04` backend support asserts `full_text` matches exact normalized tokens and `prefix` matches only at code-point offset 0 after comparison normalization.
5. `E-8-04` asserts no fuzzy similarity, phrase semantics, wildcard semantics, stemming, transliteration, accent-insensitive matching, or hidden relevance ranking changes result inclusion or ordering.

Implementation tasks:
- Bind workbook query pagination to `live_authorized_keyset`; explicitly avoid `snapshot_stable` behavior for live routes.
- Ensure cursor contract includes normalized route, view schema, filters, sort, grouping, actor/session-visible authorization scope as required by the pagination helper, and continuation position.
- Implement or tighten exact-token `full_text` and strict `prefix` SQL execution against projections.
- Preserve applied sort tuple from `meta.query.sort` even when `full_text` is present; do not inject relevance score ordering.
- Implement null-last sort ordering per view-schema discovery and route execution semantics.

Validation commands:
- `go test ./internal/modules/workbook ./internal/platform/pagination -run 'TestPhase8_.*I_8_04'`
- `go test ./internal/modules/workbook -run 'TestPhase8_.*E_8_04'`
- `make backend-integration CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=integration`
- `make service-backed-slice PHASE=phase8`
- `git diff --check`

Deliverables:
- Direct live-authorized cursor evidence.
- Exact search semantics ready for browser workflow in Sprint 8.
- No snapshot-stable live workbook continuation behavior.

Risks and open questions:
- Pagination tests must avoid relying on stale retained row payloads; payload freshness is part of the claim.
- Search fixtures must include near-miss rows proving what does not match, not only positive matches.

Exit criteria:
- Cursor continuation loses access when the caller loses authorization before page 2.
- Search result ordering remains the applied sort order, not relevance order.

## Sprint 7. Grouping, Discovery, and Grid Controls

Objective: Complete Timeline grouping, view-schema discovery, and frontend query-control contracts so grouping and header controls are stable-key, presentation-only behavior.

Status: Planned.

Relevant IDs:
- `U-8-07`, `U-8-09`, `U-8-GRID-01`
- Browser follow-up: `E-8-03`
- `REQ-01-022`, `REQ-01-286`, `REQ-01-310`, `REQ-01-312`, `REQ-01-323`, `REQ-01-326`, `REQ-01-328`, `REQ-01-329`, `REQ-01-331`, `REQ-01-332`, `REQ-01-336`, `REQ-01-339`, `REQ-01-499`, `REQ-01-503..REQ-01-506`, `REQ-03-223..REQ-03-235`
- `AC-024..AC-026`, `AC-124`, `AC-127`, `AC-184`, `AC-243`, `AC-359..AC-365`

Evidence layers: `backend_unit` for `U-8-07` and `U-8-09`; `frontend_unit` for `U-8-GRID-01`; `browser_functional` support in Sprint 8.

Grep references:
- `grouping_fields`
- `timeline.occurred_day`
- `timeline.recorded_day`
- `timeline.capture_state`
- `timeline.has_evidence`
- `timeline.has_unresolved_mentions`
- `sort_fields`
- `header_sort_field_key`
- `technical_fields`
- `record_id`
- `resolveHeaderSortFieldKey`

Files and areas:
- `contracts/view-schemas/cartulary.view.timeline.v1.json`
- `contracts/view-schemas/index.json`
- `internal/platform/viewschema`
- `internal/platform/viewquery`
- `internal/modules/workbook`
- `apps/web/src/workbookQuery.ts`
- Workbook grid/control components and tests under `apps/web/src`

Test-first sequence:
1. `U-8-07` asserts Timeline grouping permits exactly the normative whitelist keys and rejects all other grouping keys.
2. `U-8-07` asserts group headers are presentation-only and never serialize as writable rows, paste targets, subtotal rows, or `record_id`-bound mutation targets.
3. `U-8-09` asserts discovery exposes exact `sort_fields`, `header_sort_field_key`, grouping whitelists, null-last ordering, and no client-sortable `record_id`.
4. `U-8-09` asserts header sort on non-sortable collection fields such as `timeline.tags` synthesizes no client sort.
5. `U-8-GRID-01` asserts header sort, filter, and group controls send stable `field_key`/group keys only, never visible labels, vendor column indexes, projection table names, or storage table names.

Implementation tasks:
- Audit view-schema contract files and registry output against Phase 8 discovery requirements.
- Ensure technical fields remain discoverable as technical fields while not being client-sortable user fields.
- Add grouping result metadata/rendering support without creating row resources for headers.
- Update frontend query builders and controls to use server-discovered keys.
- Add frontend tests that fail if labels or implementation names appear in query request bodies.

Validation commands:
- `go test ./internal/platform/viewschema ./internal/platform/viewquery ./internal/modules/workbook -run 'TestPhase8_.*(U_8_07|U_8_09)'`
- `make frontend-unit CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=unit CARTULARY_MANIFEST_COVERAGE=authoritative CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=frontend_unit`
- `make frontend-typecheck`
- `make lint-biome`
- `git diff --check`

Deliverables:
- View-schema discovery conforms to Phase 8.
- Frontend grid controls conform to stable-key query contract.
- Group headers remain presentation-only.

Risks and open questions:
- Existing view schemas may include `record_id` in `default_sort`; that is allowed as server-applied tie-breaker but not as a client-sortable field.
- Frontend grouping UX must not introduce editable pseudo-rows that can be selected as mutation targets.

Exit criteria:
- `U-8-GRID-01` has authoritative frontend unit coverage.
- Backend and frontend agree on grouping/sort/filter keys through public discovery, not duplicated labels.

## Sprint 8. Sparse Patch, Hidden Writable Fields, and Browser Workflows

Objective: Complete sparse-patch/full-row wire semantics and browser-functional Phase 8 workflows.

Status: Planned.

Relevant IDs:
- `U-8-10`
- `E-8-01`, `E-8-02`, `E-8-03`, `E-8-04`
- `REQ-00-027`, `REQ-01-034`, `REQ-01-036`, `REQ-01-267`, `REQ-01-310`, `REQ-03-097`, `REQ-03-247`
- Browser REQs/ACs from the relevant E rows
- `AC-014`, `AC-024..AC-026`, `AC-044`, `AC-146..AC-153`, `AC-184..AC-185`, `AC-366..AC-368`, `AC-387`

Evidence layers: `backend_integration` for `U-8-10`; `browser_functional` for `E-8-01..E-8-04`; frontend unit support where useful.

Grep references:
- `changed_field_keys`
- `affected_views`
- `full row`
- `sparse`
- `hidden writable`
- `null`
- `saved view`
- `Group: None`
- `first useful viewport`
- `full_text`
- `prefix`

Files and areas:
- `internal/modules/workbook`
- `internal/modules/timeline`
- `internal/platform/viewquery`
- `internal/platform/viewschema`
- `apps/web/src/WorkbookShell.tsx`
- `apps/web/src/workbookQuery.ts`
- `apps/web/e2e/phase8.*.spec.ts`
- Browser fixtures and seeded data helpers

Test-first sequence:
1. `U-8-10` asserts full-row query and sparse live-patch payloads include hidden writable fields when required by the wire family.
2. `U-8-10` asserts authoritative nulls remain explicit where the wire contract requires them and are not guessed away by partial patch logic.
3. `U-8-10` asserts `changed_field_keys[]` and `affected_views[]` are canonical and stable.
4. `E-8-01` currently asserts the saved-view lifecycle product gap directly by proving saved-view routes and browser affordances are not available yet; replace this sentinel with lifecycle workflow proof when the product behavior lands.
5. `E-8-02` asserts workbook open honors explicit sheet, home pointer, default pointer, then Timeline fallback, clearing invalid pointers along the way.
6. `E-8-03` asserts Timeline sorting, filtering, and grouping produce a stable first useful viewport and deterministic final grouping order without group headers as rows.
7. `E-8-04` asserts exact-token `full_text` and strict `prefix` behavior does not degrade to fuzzy, phrase, wildcard, stemming, transliteration, or relevance-ranked search.

Implementation tasks:
- Tighten row/full-patch serializers and collaboration patch payloads for hidden writable fields and null preservation.
- Canonicalize `changed_field_keys[]` and `affected_views[]` at the server boundary.
- Build or complete saved-view routes and browser UI affordances needed to replace the `E-8-01` blocker sentinel with lifecycle workflow proof.
- Wire startup selection and invalid-pointer clearing through the browser shell for `E-8-02`.
- Add query-control workflows for `E-8-03` and search fixtures for `E-8-04`.
- Keep browser tests focused on Phase 8 workflows; do not claim Phase 9 keyboard/clipboard or later surface workflow behavior.

Validation commands:
- `go test ./internal/modules/workbook ./internal/modules/timeline -run 'TestPhase8_.*(U_8_10|I_8_04|AC185)'`
- `make frontend-unit`
- `make frontend-typecheck`
- `make browser-e2e-webserver-backed CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_SECTION=e2e CARTULARY_MANIFEST_COVERAGE=authoritative`
- `make phase-slice PHASE=phase8`
- `make service-backed-slice PHASE=phase8`
- `git diff --check`

Deliverables:
- `U-8-10` backend integration evidence.
- Browser-functional evidence for `E-8-02..E-8-04` and a direct `E-8-01` product-gap sentinel.
- Frontend and browser fixtures that prove stable row binding under sorting/filtering/grouping.

Risks and open questions:
- Browser rows must not mask missing backend behavior; backend rows must pass independently first.
- Visual goldens may support regression review, but they cannot be the authoritative source for Phase 8 behavior claims.

Exit criteria:
- Browser-functional rows are real assertions, not placeholders. `E-8-01` remains a deliberate blocker sentinel until product saved-view lifecycle behavior exists.
- Sparse-patch and full-row payloads do not drop hidden writable fields or authoritative nulls.

## Sprint 9. Phase Gate, Ledgers, Schedules, Baselines, and Exit

Objective: Prove Phase 8 is complete, resolve generated drift, update ledgers and schedules, refresh baselines when needed, and record handoff notes for Phase 9.

Status: Planned.

Relevant IDs: all Phase 8 authoritative rows and generated artifacts.

Evidence layers: all.

Grep references:
- `phase8_test_map`
- `phase8_coverage_ledger`
- `phase-slice PHASE=phase8`
- `service-backed-slice PHASE=phase8`
- `generate-drift`
- `phase-ledger-drift`
- `phase-schedule-drift`
- `agent-finalize`

Files and areas:
- `PHASE8_IMPLEMENTATION_PLAN.md`
- `tools/phase8_test_map.json`
- `docs/testing/phase8_coverage_ledger.md`
- `tools/service_backed_schedule_manifest.json`
- `tools/check_schedule_manifest.json`
- Duration baseline artifacts when applicable.
- Generated outputs under `internal/gen/**` and `packages/protocol-ts/src/generated/**` produced only by canonical generation.

Test-first sequence:
1. Confirm every authoritative Phase 8 row has direct, non-placeholder evidence in its declared layer.
2. Confirm support-only tests, generated files, visual goldens, and non-normative prose do not own Phase 8 behavior.
3. Confirm all generated artifacts are current.
4. Run public Phase 8 wrappers and broader gates.
5. Record artifact roots and any exact non-Phase-8 blocker.

Implementation tasks:
- Run `make phase-ledgers` and `make phase-schedules` after final manifest/test changes.
- Run generated-contract hygiene after OpenAPI, view-schema, SQL, or protocol source changes.
- Refresh duration baselines only from qualifying successful retained run roots.
- Update this plan's sprint checklist with actual validation status, blockers, and artifact roots.
- Record Phase 9 compatibility notes without broadening Phase 8 scope.

Validation commands:
- `make explain-phase PHASE=phase8`
- `make phase-map-check`
- `make phase-test-name-check`
- `make generate`
- `make generate-drift`
- `make phase-ledgers`
- `make phase-ledger-drift`
- `make phase-schedules`
- `make phase-schedule-drift`
- `make backend-unit CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_COVERAGE=authoritative`
- `make backend-store CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_COVERAGE=authoritative`
- `make backend-integration CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_COVERAGE=authoritative`
- `make frontend-unit CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_COVERAGE=authoritative`
- `make frontend-typecheck`
- `make browser-e2e-webserver-backed CARTULARY_MANIFEST_PHASE=phase8 CARTULARY_MANIFEST_COVERAGE=authoritative`
- `make phase-slice PHASE=phase8`
- `make service-backed-slice PHASE=phase8`
- `make test-fast`
- `make agent-finalize`
- `make check`
- `git diff --check`

Deliverables:
- Current manifest, ledger, schedules, and generated artifacts.
- Passing Phase 8 public phase wrappers.
- Passing `make check`, or exact documented non-Phase-8 blocker with target, artifact root, failing row, and reason.
- Updated Phase 8 plan with final artifact roots and Phase 9 handoff notes.

Risks and open questions:
- If `make check` fails outside Phase 8, do not claim full gate pass; record the exact non-Phase-8 blocker and keep Phase 8 public wrappers clean.
- If generated drift appears, fix source artifacts and regenerate; do not hand-edit generated files.

Exit criteria:
- All Phase 8 rows in `tools/phase8_test_map.json` are direct, passing, and non-placeholder.
- `docs/testing/phase8_coverage_ledger.md` is generated and `make phase-ledger-drift` passes.
- Phase schedules are regenerated and `make phase-schedule-drift` passes.
- `make phase-slice PHASE=phase8` and `make service-backed-slice PHASE=phase8` pass.
- Generated-contract hygiene is clean through `make generate` and `make generate-drift`.
- `make agent-finalize` is run as the first end-of-run maintenance command before broader final verification; record whether it ran unchanged, updated artifacts, skipped retained-run maintenance because `RESULTS_DIR` was unset, failed at a subtarget, or was explicitly skipped.
- `make check` passes, or the plan records an exact non-Phase-8 blocker.

## Phase Validation Criteria

Phase 8 is evidence-complete only when:
- `tools/phase8_test_map.json` exists, is active/selectable, and maps every `U-8-*`, `U-8-GRID-01`, `I-8-*`, and `E-8-*` row from the guide.
- `docs/testing/phase8_coverage_ledger.md` exists and matches the manifest.
- Authoritative evidence is separated by `backend_unit`, `backend_store`, `backend_integration`, `frontend_unit`, and `browser_functional`.
- `U-8-01..U-8-10`, `U-8-GRID-01`, `I-8-01..I-8-04`, and `E-8-01..E-8-04` have direct passing tests with Phase 8 names/titles.
- No authoritative row depends on support-only tests, generated files, visual goldens, or non-normative guide prose.
- Phase 7 handoff boundaries are closed or preserved explicitly: tag rollback is implemented and evidenced in Phase 8, `AC-184` and `AC-185` have Phase 8 evidence, and merge-specific rollback remains outside Phase 8 unless separately owner-driven.
- Phase 9 remains out of scope except for stable compatibility in shared query/view-schema/row-patch contracts.
- Saved-view duplicate behavior is evidenced without inventing an unowned duplicate route.
- Live workbook cursor pagination uses live-authorized keyset semantics and never claims immutable snapshot-stable behavior.

## Phase Exit Criteria

Phase 8 may be marked complete only after all of the following are true:
- Sprint checklist rows are updated with actual validation status, blockers, and follow-up notes.
- `make explain-phase PHASE=phase8` reports expected manifest, ledger, execution dependencies, service requirements, and coverage.
- `make phase-map-check`, `make phase-test-name-check`, `make phase-ledger-drift`, and `make phase-schedule-drift` pass.
- `make generate` and `make generate-drift` pass after all contract, SQL, and view-schema source changes.
- `make backend-unit`, `make backend-store`, `make backend-integration`, `make frontend-unit`, and the Phase 8 browser-functional selection pass for authoritative Phase 8 rows.
- `make phase-slice PHASE=phase8` and `make service-backed-slice PHASE=phase8` pass.
- `make test-fast` passes or any failure is recorded as a non-Phase-8 blocker with exact artifact root and failing target.
- `make agent-finalize` is run and recorded.
- `make check` passes, or an explicitly documented non-Phase-8 blocker remains with exact target, artifact root, failing row or test, and why it is outside Phase 8.
- `git diff --check` passes.

## Handoff Requirements for Phase 9

Phase 9 may rely on these completed Phase 8 contracts after exit:
- Stable view-query validation and canonical `meta.query` semantics for later surfaces.
- Stable view-schema discovery keys for keyboard, clipboard, and remaining workbook surface controls.
- Saved-view lifecycle and startup selection behavior that can address required base surfaces by `sheet_ref.kind='view_schema'` and user-created views by `sheet_ref.kind='saved_view'`.
- Presentation-only group headers that are never editable rows or mutation targets.
- Full-row and sparse-patch contracts that preserve hidden writable fields, authoritative nulls, stable `record_id`, `row_version`, `changed_field_keys[]`, and `affected_views[]`.
- Exact-token and strict-prefix search semantics without relevance ranking.

Phase 9 handoff must also state:
- Phase 8 did not implement Phase 9 keyboard/clipboard workflows.
- Phase 8 did not complete Notes, Indicators, Parties, Assessments, Task Requests, Decisions, or coordination-surface workflow obligations except where those surfaces consume shared Phase 8 query/view-schema/row-patch behavior.
- Any optional or future immutable snapshot/reporting route family remains outside Phase 8 live workbook query behavior.

## Risks and Open Questions

- Saved-view duplicate route shape: no separate duplicate route is assumed. Use ordinary create from a visible source resource unless normative owner text found during implementation explicitly requires a separate route.
- Cursor terminology: earlier planning language may mention snapshot-stable pagination, but Core 01 reserves `snapshot_stable` for explicit immutable snapshot/reporting artifacts. Phase 8 live workbook routes must implement live-authorized keyset continuation.
- `record_id` sort: discovery must expose `record_id` only as a technical field/server tie-breaker, not as a client-sortable field.
- Support evidence: existing Phase 3 through Phase 7 tests may cover parts of the substrate, but Phase 8 claims require direct Phase 8 IDs and authoritative rows.
- Browser workflow scope: browser tests must prove user-visible Phase 8 behavior but must not broaden into Phase 9 keyboard/clipboard or surface-specific workflow obligations.
