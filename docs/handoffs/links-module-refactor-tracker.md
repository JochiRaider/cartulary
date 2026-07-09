# Links Module Refactor Tracker: Next Iteration

## 1. Current-State Audit

### Scope and authority

- Target path: `internal/modules/links`
- Output path: `docs/handoffs/links-module-refactor-tracker.md`
- Status: active execution tracker and handoff ledger for the links module remediation sequence.
- This iteration may change production code, tests, harness inputs, migrations, and owner documentation only when required by the workstreams below. Generated files and generated harness/topology outputs remain downstream artifacts and must not be hand-edited.
- Product behavior remains owned by Core 00 through Core 04. Core 05 is only for claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
- Domain vocabulary and owner-boundary interpretation follow `docs/domain.md`.
- Harness mechanics and Make target interpretation follow `docs/testing-harness-nlspec.md`.
- No links-specific adopted subsystem NLSpec was found; links behavior is currently Core-owned plus repository-owned implementation evidence.

### Execution tracking rules

- This file is the controlling remediation artifact for LRG-001 through LRG-007.
- Before starting a workstream, reread this tracker and confirm the previous workstream status.
- After completing a workstream and before starting the next workstream, update the execution ledger with status, substantive changes, compatibility notes, validation commands and results, result roots when available, skipped checks with reasons, and remaining exceptions.
- Temporary raw links-table access exceptions are allowed only while mapped to a later workstream. They must be removed or explicitly superseded before the validation and handoff completion slice.
- Public route payloads, item-ref strings, history refs, incident-bundle filenames, WebSocket payload shapes, generated contracts, and `active_record_links_v1` / `active_record_tags_v1` semantics are frozen unless an owner spec authorizes drift.

Primary source documents and files inspected for this tracker:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/testing-harness-nlspec.md`
- `tools/schema_object_ownership_manifest.json`
- `db/migrations/00025_links_projection_read_contracts.sql`
- `internal/modules/links/**`
- `internal/modules/workbook/**`
- `internal/modules/artifacts/**`
- `internal/modules/entities/merge/**`
- `internal/modules/revisions/**`
- `internal/modules/*/projectionprovider/**`
- Adjacent production direct readers and writers of `record_links`, `record_tags`, `active_record_links_v1`, and `active_record_tags_v1`.

### Boundaries now clean

- `record_links`, `record_tags`, `active_record_links_v1`, and `active_record_tags_v1` are owned by Links in `tools/schema_object_ownership_manifest.json`.
- `handoff_risk_refs` is owned by Artifacts in `tools/schema_object_ownership_manifest.json`; workbook mutation dispatch routes `handoff.open_risk_refs` validation and application to the artifacts store.
- Projection providers generally consume links-owned active read contracts, especially `active_record_links_v1` and `active_record_tags_v1`, instead of raw links source tables.
- Revisions instantiate `internal/modules/links/revisionprovider` for link/tag rollback target operations. Revisions still owns rollback orchestration, route semantics, destructive-operation behavior, and history selection.
- Entity merge no longer makes Links own timeline field-key invalidation mapping. Links returns relationship effects, and `internal/modules/timeline/linkeffects` maps relationship link types to timeline field keys.
- `record_tags` remains inside Links and Tags by ownership decision; tag CRUD is exposed through the narrower `TagStore` facade rather than a new top-level module.
- Links incident-bundle and reporting behavior remains shaped as owner-provider adapters rather than moving bundle or reporting orchestration into Links.

### Residual boundaries and deliberate exceptions

- `links.Store` remains the concrete owner implementation inside Links and private caller adapters, but non-owner module fields added during this remediation depend on narrow local ports.
- Workbook, linked notes, assessments, tasks/decisions, Timeline, entity merge, and mention code now pass caller-shaped link/tag commands or use narrow Links ports instead of depending on Links-owned surface policy switches.
- Non-links production code no longer reads or writes raw `record_links` or `record_tags` source tables for active relationship/tag behavior. Active consumers use `active_record_links_v1`, `active_record_tags_v1`, or Links ports.
- Link/tag rollback and mutation-value loading use Links-owned value codecs. Merge keeps its pre-existing compact mutation value shape behind Links-owned helper functions to avoid history/rollback drift.
- Canonical `record_ref`, `party_ref`, and persisted `record_tag` helpers live under Links. Canonical `risk_ref` helpers live under Artifacts.
- `backend-module-boundary-check` now enforces source-table access for `record_links` and `record_tags`; the final allowlist contains only `internal/modules/links/**`.

### Prior gaps: genuinely closed versus mechanically moved

- RB-001 is closed. `handoff_risk_refs` mutation no longer flows through Links and is artifacts-owned.
- RB-002 is closed. Timeline field-key mapping for merge invalidation lives behind a timeline adapter, not in Links.
- RB-003 is closed for link/tag value ownership. Revisions still orchestrates rollback and changed-field-key policy, while Links owns record-link and record-tag target loading/reconstruction semantics.
- RB-004 is closed for active links/tag reads. Projection, Timeline, support-count, and lifecycle readers use Links-owned active views or Links ports instead of raw source tables.
- RB-005 is closed as an ownership decision. Tags stay under Links and Tags, and Timeline now uses Links helpers for tag mutation values and item refs.

## 2. Gap Inventory and Closure Criteria

The table below preserves the original gap statement and validation criteria. Closure status, result roots, compatibility notes, and remaining exceptions are tracked in the execution ledger.

| Gap ID | Affected files/modules | Current behavior | Proposed durable remediation | Areas affected | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if left unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| LRG-001 | `internal/modules/links/collection_actions.go`, `internal/modules/links/field_refs.go`, `internal/modules/workbook/mutation_store.go`, `internal/modules/artifacts/linkednotes` | Links owns a generic collection payload and hardcoded field-key registry for record refs, party refs, tags, and link-type selection. | Replace with caller-shaped commands where the owning caller passes target policy, public field key, item family, and relation token; Links applies link/tag invariants and source-state mutation only. | implementation, tests, docs | Future workbook fields should not require editing a Links-owned switch for non-links surface semantics. | New collection fields can be added by their surface owner while Links stays small and testable. | No public route or payload change. Internal APIs may break intentionally. No migration expected. | Future phase fields will couple workbook/artifact semantics into Links and make field growth brittle. | Existing Phase 8/9 collection behavior passes; Links no longer owns target-type lookup for non-links surfaces. |
| LRG-002 | `internal/modules/timeline/mentions_collections_store.go`, `internal/modules/timeline/rowsnapshot`, `internal/modules/timeline/query_projection_store.go`, `internal/modules/timeline/auto_resolution.go`, `internal/modules/timeline/lifecycle_store.go` | Timeline duplicates link/tag mutation, item-ref parsing, value loading, attached-evidence link mutation, tag filtering, and raw table reads. | Add links-owned tag/link mutation helpers returning mutation values; use active read views or narrow links query ports for timeline hydration, filtering, and metadata lookup. | implementation, tests | Timeline should focus on capture, mentions, and row lifecycle while Links owns relationship/tag source state. | Timeline changes stop drifting from workbook link/tag semantics; link/tag schema evolution becomes local. | No migration expected unless new read views are added. Public timeline fields and payloads stay stable. | Divergent semantics between timeline and workbook paths; duplicate rollback/history value builders. | `make backend-store`, `make backend-integration`, `make phase-slice PHASE=phase8`; include Phase 3/4 timeline coverage when touched. |
| LRG-003 | `internal/modules/projections/query_sql.go`, `internal/modules/indicators/store.go`, `internal/modules/assessments/store.go`, `internal/modules/evidence/store.go`, `internal/modules/tasksdecisions/mutations.go`, timeline stores | Non-owner production code still queries `record_links` or `record_tags` directly. | Define and enforce an allowlist: raw tables are allowed only inside links owner code and approved owner-provider export/import/reporting adapters; other readers use active views or links query ports. | implementation, tests, harness | The current active read views help projection providers, but convention alone will not keep future code from reaching into source tables. | Links schema changes become reviewable and local; ownership violations fail early. | If a boundary manifest changes, run JSON shape checks. No public contract change. | Raw table coupling will reappear and make future schema or view-version changes expensive. | `make backend-module-boundary-check`; `make json-shape-check` if manifests change; no forbidden raw table reads outside the allowlist. |
| LRG-004 | `internal/modules/links/revisionprovider`, `internal/modules/links/merge_effects.go`, `internal/modules/timeline/mentions_collections_store.go`, `internal/modules/revisions/rollback_store.go`, `internal/modules/revisions/workbook_conflicts.go` | Link/tag revision values are untyped `map[string]any`; changed-field-key derivation remains partly outside Links. | Add links-owned value codecs and item-ref helpers for link/tag mutation values; keep Revisions as rollback orchestrator, but move owner-specific link/tag value and affected-field derivation behind links-owned helpers where practical. | implementation, tests | Future schema fields should update one owner codec, not several JSON builders. | Rollback, merge, timeline mutation entries, and history refs remain consistent as link/tag schema evolves. | Public history refs and rollback payloads must remain byte-shape compatible unless Core changes. No migration expected for helper-only work. | Rollback/history silently drift when link/tag schema changes or when new link families are added. | Phase 7 rollback/history coverage plus `make backend-store` and Phase 8 link/tag rollback coverage. |
| LRG-005 | `internal/modules/links/*`, callers in workbook, timeline, entities, tasksdecisions, artifacts | `links.Store` exposes unrelated operations to all callers. | Split internal links facades into narrow surfaces: relations, tags, collections, merge effects, revision targets, portability, and reporting. Keep the same package or subpackages; do not create a new top-level module. | implementation, tests | A broad store invites opportunistic reuse and hides the reason each caller depends on Links. | Callers depend on small behavior-shaped ports, making expansion and tests easier. | Internal compatibility only; public API unchanged. No migration expected. | New callers keep depending on the broad store and expand accidental coupling. | Callers import or depend on narrow interfaces; backend boundary check passes. |
| LRG-006 | Links item-ref parser, timeline tag helpers, revisions conflict helpers, projection SQL, artifacts risk refs | Public item refs are duplicated string construction. | Centralize owner helpers for links/tag refs and artifacts risk refs; SQL builders and client-conflict helpers should match helper output exactly. | implementation, tests, docs | Public refs are compatibility surface, not implementation detail. | History, rollback, projection chips, and conflict UI stay aligned. | Item-ref strings must remain stable. No migration expected. | Rollback/history/query clients see inconsistent refs or invalid remove actions. | Characterization tests for `record_ref`, `party_ref`, `risk_ref`, and `record_tag` refs pass. |
| LRG-007 | `db/migrations/00025_links_projection_read_contracts.sql`, projection providers, ownership manifest | Active read views exist but this tracker had no explicit versioning policy for them. | Document that breaking active-view shape changes require additive `*_v2` views and manifest updates; v1 must not be silently reinterpreted. | docs, migration, tests | The active views are now owner contracts, so they need version discipline. | Projection/read-side growth can happen without breaking existing consumers. | New views require migration and drift checks. Existing v1 views remain stable. | Future projection changes break consumers without a contract signal. | `make migration-drift`, `make json-shape-check`, `make generated-artifact-policy-check` when views/manifests change. |

## 3. Workstream Plan

| Workstream | Dependencies | Sequencing | Risk level | Exit criteria | Suggested narrow Make validation targets |
| --- | --- | --- | --- | --- | --- |
| WS-1 Links access audit and guardrails | none | first | medium | Raw links-table access allowlist is documented and enforced; approved owner-provider adapters are explicit. | `make backend-module-boundary-check`; `make json-shape-check` if manifests change |
| WS-2 Caller-shaped collection contracts | WS-1 | before timeline cleanup | high | Links no longer owns non-links target-type field registry; workbook/artifacts pass caller-shaped policies into Links. | `make backend-store`; `make backend-integration`; `make phase-slice PHASE=phase8` |
| WS-3 Timeline link/tag convergence | WS-1, WS-2 | after collection contracts | high | Timeline no longer duplicates tag/link mutation or raw read behavior except approved active-view reads or narrow links query ports. | `make backend-store`; `make backend-integration`; `make service-backed-slice PHASE=phase8`; include Phase 3/4 timeline targets when touched |
| WS-4 Link/tag revision value codec | WS-1 | may run parallel with WS-2 | high | One links-owned value codec feeds rollback, merge, and timeline mutation entries; public history refs remain stable. | Phase 7 rollback coverage through Make-owned backend targets; `make backend-store`; `make phase-slice PHASE=phase8` |
| WS-5 Facade simplification | WS-2 through WS-4 | last | medium | Broad `links.Store` is no longer the default caller dependency; callers use narrow interfaces or subfacades. | `make backend-module-boundary-check`; `make backend-integration` |
| WS-6 Contract/versioning documentation | WS-1 | after any view or manifest change | low | Tracker and owner inputs explain active read-view stability and additive versioning. | `make migration-drift`; `make generated-artifact-policy-check`; `make lint-markdown` |

## 3A. Execution Ledger

| Workstream | Status | Changed areas | Compatibility notes | Validation | Remaining exceptions |
| --- | --- | --- | --- | --- | --- |
| WS-0 Tracker and spec cleanup | complete | Tracker execution rules and active-view policy. | No public contract, schema, or generated artifact changes. | Pending end-of-run docs validation. | None. |
| WS-1 Access audit and guardrail staging | complete | Added manifest-driven `source_table_access` scanning to `backend-module-boundary-check`; added the links source-table rule for production Go and authored query SQL. | No public contract change. Temporary raw-table exceptions are explicit and must be retired by WS-3 through WS-5. | `make backend-module-boundary-check` pass, run root `.cartulary/test-results/20260709T162142Z-p14539`; `make json-shape-check` pass, run root `.cartulary/test-results/20260709T162142Z-p14560`. | Temporary exceptions: `db/queries/timeline_phase3.sql`, `internal/modules/assessments/store.go`, `internal/modules/evidence/store.go`, `internal/modules/indicators/store.go`, `internal/modules/projections/query_sql.go`, `internal/modules/tasksdecisions/mutations.go`, `internal/modules/timeline/auto_resolution.go`, `internal/modules/timeline/lifecycle_store.go`, `internal/modules/timeline/mentions_collections_store.go`, `internal/modules/timeline/query_projection_store.go`, and `internal/modules/timeline/rowsnapshot/**`. |
| WS-2 Caller-shaped collection contracts | complete | Links collection APIs now take caller-supplied `CollectionFieldPolicy`; Workbook and linked-notes owners supply field policy for link type, target type, item family, and tag handling. Removed Links-owned `FieldLinkType` and `expectedCollectionTargetType` routing. | Public collection payloads and item-ref strings unchanged. Internal Links APIs intentionally changed. No migration. | `make backend-store` pass, run root `.cartulary/test-results/20260709T162450Z-p27228`; `make backend-integration` pass, run root `.cartulary/test-results/20260709T162450Z-p27279`; `make phase-slice PHASE=phase8` pass, run root `.cartulary/test-results/20260709T162602Z-p58014`. | Raw table temporary exceptions remain for WS-3 through WS-5. |
| WS-3 Timeline link/tag convergence | complete | Timeline tag and attached-evidence collection mutations now route through Links collection mutation helpers with Links-owned mutation values. Timeline collection hydration, filters, row snapshots, auto-resolution metadata, and Phase 3 query SQL now use active read views or Links ports. Supersedes active-link locking moved behind a Links query port. Timeline stale raw-table exceptions were removed from the boundary manifest. | Public Timeline routes, fields, payloads, item refs, history refs, WebSocket shape, and active-view v1 semantics unchanged. Internal Timeline/Links ports changed intentionally. No migration. | `make backend-store` pass, run root `.cartulary/test-results/20260709T164421Z-p78096`; `make backend-integration` pass, run root `.cartulary/test-results/20260709T164513Z-p97519`; `make phase-slice PHASE=phase3` pass, run root `.cartulary/test-results/20260709T164614Z-p13573`; `make phase-slice PHASE=phase4` pass, run root `.cartulary/test-results/20260709T164806Z-p38828`; `make service-backed-slice PHASE=phase8` pass, run root `.cartulary/test-results/20260709T164849Z-p56209`; `make backend-module-boundary-check` pass, run root `.cartulary/test-results/20260709T165051Z-p82799`; `make json-shape-check` pass, run root `.cartulary/test-results/20260709T165051Z-p82820`. | Remaining temporary raw-table exceptions: `internal/modules/assessments/store.go`, `internal/modules/evidence/store.go`, `internal/modules/indicators/store.go`, `internal/modules/projections/query_sql.go`, and `internal/modules/tasksdecisions/mutations.go`. |
| WS-4 Link/tag revision value codec | complete | Added shared Links value codec package for full `record_link` / `record_tag` mutation values and identity parsing; Links store and revision provider now delegate value loading/parsing to it. Merge effects use Links-owned compact mutation value helpers to preserve existing merge value shape. Timeline supersedes mutation entries use the Links value loader. Added canonical Links helpers and tests for `record_ref`, `party_ref`, and `record_tag`; added Artifacts-owned `risk_ref` helpers/tests and routed Workbook/Revisions parsing/formatting through owner helpers. | Public item-ref strings and route payloads unchanged. Rollback/history semantics preserved; merge compact mutation value shape intentionally retained while moving construction behind Links helpers. Internal helper/package APIs changed. No migration. | `make backend-store` pass, run root `.cartulary/test-results/20260709T170144Z-p31962`; `make backend-integration` pass, run root `.cartulary/test-results/20260709T170144Z-p31993`; `make phase-slice PHASE=phase7` pass, run root `.cartulary/test-results/20260709T170300Z-p66691`; `make phase-slice PHASE=phase8` pass, run root `.cartulary/test-results/20260709T170342Z-p80161`. | Remaining temporary raw-table exceptions: `internal/modules/assessments/store.go`, `internal/modules/evidence/store.go`, `internal/modules/indicators/store.go`, `internal/modules/projections/query_sql.go`, and `internal/modules/tasksdecisions/mutations.go`. |
| WS-5 Facade simplification | complete | Replaced remaining non-owner raw `record_links` / `record_tags` production reads with `active_record_links_v1` / `active_record_tags_v1`; reduced boundary source-table allowlist to `internal/modules/links/**` only. Narrowed remaining direct module fields for Workbook, linked notes, assessments, and tasks/decisions to local Links ports; retained concrete `links.Store` only inside constructors/private adapters. | No public contract change. Active read-view v1 semantics used as intended. Internal caller dependencies narrowed; no migration. | `make backend-module-boundary-check` pass, run root `.cartulary/test-results/20260709T171243Z-p39325`; `make json-shape-check` pass, run root `.cartulary/test-results/20260709T171243Z-p39331`; `make backend-integration` pass, run root `.cartulary/test-results/20260709T171243Z-p39398`. | No temporary raw-table exceptions remain. |
| WS-6 Contract/versioning documentation | complete | Refreshed tracker language so the audit describes the implemented owner boundaries, source-table guardrails, item-ref helpers, and active-view versioning posture. Regenerated SQL artifacts after authored Phase 3 query changes. | No active-view v1 shape or migration change. Generated SQL output updated through `make generate`; generated roots were not hand-edited. | `make migration-drift` pass, run root `.cartulary/test-results/20260709T171524Z-p69097`; `make generated-artifact-policy-check` pass, run root `.cartulary/test-results/20260709T171620Z-p78722`; `make json-shape-check` pass, run root `.cartulary/test-results/20260709T171524Z-p69150`; `make lint-markdown` pass; `make generate` pass, run root `.cartulary/test-results/20260709T171604Z-p75909`; `make generate-drift` pass after regeneration, run root `.cartulary/test-results/20260709T171612Z-p77216`. Initial `make generate-drift` before regeneration failed at `.cartulary/test-results/20260709T171525Z-p69230` due expected generated SQL drift. `make generate-sqlc` was not available as a public target, so `make generate` was used. | No temporary raw-table exceptions remain. |
| WS-7 Validation and handoff completion | complete | Final handoff slice closed all LRG gaps, removed dead helpers found by staticcheck, retained the successful full-check evidence, and refreshed generated harness maintenance files through `make agent-finalize RESULTS_DIR=...`. | Public contracts remain unchanged. Internal API changes are intentional: caller-shaped Links collection policy, narrow caller ports, owner item-ref helpers, shared Links value codecs, and Links-owned active read contracts. No data migration. | `make agent-finalize` pass, run root `.cartulary/test-results/20260709T171703Z-p81698`; `make test-fast` pass, run root `.cartulary/test-results/20260709T171719Z-p84037`; initial `make check` failed at `.cartulary/test-results/20260709T172008Z-p48778` on staticcheck dead helpers and was fixed; `make lint-go-staticcheck` pass; `make lint-go` pass; rerun `make check` pass, run root `.cartulary/test-results/20260709T172244Z-p18876`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260709T172244Z-p18876` pass, run root `.cartulary/test-results/20260709T172502Z-p15557`; post-finalizer `make json-shape-check` pass, run root `.cartulary/test-results/20260709T172732Z-p28291`; final `make backend-module-boundary-check` pass, run root `.cartulary/test-results/20260709T172850Z-p36184`; final `make lint-markdown` pass. | No temporary exceptions remain. |

## 4. Future Phase Expansion Review

The current links/tag/projection/revisions/workbook contracts can scale only if future relationship behavior enters through owner-shaped commands and active read contracts.

Design choices that scale:

- Links owns source-state relationship/tag invariants, not every workbook surface rule.
- Workbook and artifact owners translate route DTOs and surface-specific target rules before invoking Links.
- Projection providers consume links-owned active views for active-row and active-endpoint semantics.
- Revisions owns history and rollback orchestration while Links owns link/tag target reconstruction.
- Tags stay under Links and Tags unless a later owner document creates a real separate bounded context.

Choices that would make future phases brittle:

- Growing `FieldLinkType` or `expectedCollectionTargetType` as global Links switch statements.
- Letting non-links modules read or write `record_links` or `record_tags` directly when an active view or narrow links port would work.
- Preserving map-shaped mutation values as the only owner contract between Links, Revisions, Timeline, and merge.
- Treating `links.Store` as the normal dependency for all callers instead of narrow ports.
- Reconstructing item refs by hand in route, projection, conflict, or rollback code.
- Silently changing `active_record_links_v1` or `active_record_tags_v1` instead of adding an additive versioned read contract.
- Keeping internal compatibility shims that preserve the old broad ownership model after public behavior has been protected.

Future phase work should preserve public compatibility only where the public contract requires it. Internal APIs can be removed or broken when they preserve the wrong ownership model.

## 5. Simplification Decisions Applied

- Removed `FieldLinkType` and `expectedCollectionTargetType` as Links-owned global registries. Surface owners now pass caller-owned relation-token and target-type policy into Links commands.
- Consolidated record-link and record-tag value loading through Links-owned codecs used by Timeline, Revisions, rollback helpers, and merge effects.
- Kept linked-note/artifact collection policy outside generic Links field switches.
- Moved direct active tag/link reads in Timeline, projections, indicators, assessments, evidence, and tasks/decisions to Links-owned active views or Links ports.
- Replaced broad non-owner `*links.Store` fields with local narrow ports in the modules touched by this remediation.
- Centralized persisted `record_ref`, `party_ref`, and `record_tag` helper behavior under Links, with Artifacts-owned helpers for `risk_ref`.
- Kept incident bundle and reporting code as thin owner-provider adapters.
- Added guardrails that fail production raw links-table source access outside the Links owner path.

## 5A. Active Read-View Versioning Policy

- `active_record_links_v1` and `active_record_tags_v1` are links-owned read contracts for active link/tag semantics.
- Consumers may rely on the v1 column names, column meanings, column types, active-row predicate, and endpoint-record liveness predicate.
- A breaking change to either active view requires an additive `*_v2` view, a migration, ownership-manifest updates when needed, and consumer migration. V1 must not be silently reinterpreted.
- Non-breaking additions should still prefer additive views when the new meaning would be ambiguous to existing consumers.

## 6. Public Contract Freeze Map

These contracts must remain stable unless a normative spec change authorizes drift.

| Surface | Stable contract | Owner posture | Notes |
| --- | --- | --- | --- |
| Workbook row create | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` | Core 01/Core 03; owner modules apply source-state behavior | Collection payload behavior must stay route-compatible. |
| Workbook row patch | `PATCH /api/v1/records/{record_id}` | Core 01/Core 03/Revisions/source owners | Collection actions, changed field keys, row versions, and conflict behavior are public. |
| Workbook query | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` | Core 01/Projections/source owners | Tag chips, linked-record cells, support refs, supersedes fields, and counts must remain stable. |
| Linked-note create | `POST /api/v1/records/{record_id}/linked-notes` | Workbook route; Artifacts and Links source owners | `references_artifact` link behavior and note tag behavior are public through the route result/history. |
| Supersede route | `POST /api/v1/records/{record_id}/supersede` | Workbook route; Timeline or Decisions semantics depending target | Public relation remains a typed `supersedes` link. |
| History route | `GET /api/v1/records/{record_id}/history` | Revisions | `record_link` and `record_tag` history entries and refs must remain stable. |
| Rollback route | `POST /api/v1/records/{record_id}/rollback` | Revisions orchestration; owner target providers | Link/tag rollback target semantics must stay compatible. |
| Item refs | `record_ref:<record_id>`, `party_ref:<party_id>`, `risk_ref:<risk_ref_id>`, `record_tag:<record_id>:<record_tag_id>` | Core 01/Core 02 plus source owners | Do not change strings or parsing without spec authorization. |
| Incident bundle files | `data/record_links.ndjson`, `data/tags.ndjson`, `data/record_tags.ndjson`, `data/handoff_risk_refs.ndjson` | Links for link/tag files; Artifacts for risk refs | File names and stable row identity are portability contracts. |
| WebSocket payloads | `record_changed.payload.changed_field_keys[]`, `record_changed.payload.affected_views[]` | Core 01/Collaboration plus route owners | Link/tag mutations can affect changed-key lists even when Links does not publish directly. |
| Projection fields | Tags, linked record counts, support refs, supersedes fields, evidence counts, active link/tag semantics | Core 01/Projections/source owners | Active read-view semantics must remain consistent with source state. |
| Generated contracts | `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**` | Downstream generated roots | Do not hand-edit generated files. Update owner inputs and regenerate through Make when authorized. |

## 7. Validation Plan

### Tracker-only documentation edits

Run the narrow documentation and generated-policy checks first:

- `make lint-markdown`
- `make generated-artifact-policy-check`
- `make agent-finalize`

This tracker update does not require backend service-backed tests because it changes no production code, schema, generated contracts, harness maps, or migrations.

### Later implementation slices

Use the narrowest credible matrix before broadening:

- Boundary/refactor slices: `make backend-module-boundary-check`
- Link/tag/store behavior: `make backend-store`
- Route/query/history behavior: `make backend-integration`
- Schema, view, or manifest changes: `make migration-drift`, `make json-shape-check`, `make generated-artifact-policy-check`
- Phase 8 behavior: `make service-backed-slice PHASE=phase8`, then `make phase-slice PHASE=phase8`
- Rollback behavior changes: Phase 7 revision/rollback coverage through Make-owned backend targets, plus `make backend-store`
- Final handoff hygiene: `make agent-finalize`

Broaden to `make test-fast`, `make check`, or browser-specific targets only when the implementation slice affects cross-module route behavior, WebSocket/client behavior, projection rebuild behavior, or generated/phase evidence.

When a command fails, record the failing target, the retained result root or summary artifact when available, and whether the failure appears related to the refactor slice.

## 8. Open Questions and Decisions Needed

No open owner/spec questions block this tracker iteration.

Current architectural recommendation:

- Keep public compatibility stable for routes, item refs, incident bundle file names, WebSocket payloads, history refs, projection fields, and generated contracts.
- Remove or break internal APIs when they preserve broad or wrong ownership.
- Do not introduce a top-level tags module or another top-level abstraction unless a later owner decision shows that it reduces coupling more than narrow Links facades.
- Do not treat `deferred` as a substitute for a clear owner recommendation. The next implementation sequence should start with access guardrails, then caller-shaped collection contracts, then timeline convergence and revision value-codec cleanup.
- Any change to public item-ref strings, bundle file names, route payloads, WebSocket payloads, projection field semantics, or generated contract shapes requires a normative spec change before implementation.

## 9. Handoff Notes

This tracker intentionally does not repeat the closed RB-001 through RB-005 remediation plan as active work. Future agents should use those entries only as audit evidence that the first ownership cleanup happened, then start from LRG-001 through LRG-007.

Implementation agents should inspect live code again before editing. This file is a tracker and decision handoff, not a substitute for reading the current source.
