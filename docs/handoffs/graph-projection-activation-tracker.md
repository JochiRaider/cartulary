# Graph Projection NLSpec Activation Tracker

## 1. Authority and Scope

This tracker controls the activation remediation for `docs/graph_projection_nlspec.md`.

Authority posture at session start:

| Source | Status |
| --- | --- |
| `docs/graph_projection_nlspec.md` | Draft at start of activation; intended target is `status: adopted/current`. |
| Core 00 REQ-00-062 | Requires explicit adopted status and project taxonomy listing before a projections-specific NLSpec is authoritative. |
| Appendix I | Records the pre-activation evidence posture and must be updated when activation completes. |
| `docs/domain.md` | Records graph projection as draft-only at start and must be updated when activation completes. |

Activation constraints:

- Activate graph projection as a bounded subsystem NLSpec.
- Add a first internal runtime implementation under `internal/modules/graphprojection`.
- Do not add public HTTP routes, WebSocket messages, generated protocol changes, workbook `view_row_v1` changes, saved-view behavior, import behavior, or restore-rebuild behavior in this slice.
- Keep graph projection separate from workbook projection read models owned by Core 01/Core 03.

If any owner document contradiction is found, stop the current workstream and record `BLOCKED: owner contradiction`.

## 2. Workstream Status

| ID | Workstream | Status | Dependencies | Exit Criteria | Evidence |
| --- | --- | --- | --- | --- | --- |
| WS-00 | Authority gate and tracker setup | completed | none | Tracker records authority posture, assumptions, validation policy, and handoff protocol. | Created this tracker on 2026-07-01T14:27:40-04:00. |
| WS-01 | Specification activation | completed | WS-00 | Spec and authority references agree; `make lint-markdown` passes or blocker is recorded. | Activated status in spec/Core/Appendix/domain/projection trackers; `make lint-markdown` passed. |
| WS-02 | Conformance matrix and fixtures | completed | WS-01 | Matrix maps every `GP-AC-*`; shape validation passes or blocker is recorded. | Added matrix/schema/fixture corpus; `make json-shape-check` passed. |
| WS-03 | Engine foundation | completed | WS-02 | Internal package covers admission, duplicate keys, canonicalization, digests, IDs, and normalization. | Added graphprojection admission/canonical/digest foundations; backend unit passed. |
| WS-04 | Projection semantics | completed | WS-03 | Direct/reverse/aggregation/validation fixture tests pass. | Added direct, reverse, aggregation, filter, schema, and validation behavior; backend unit passed. |
| WS-05 | Persistence, lifecycle, and retention | completed | WS-04 | Migration and store lifecycle tests pass. | Added migration, migration manifest entry, store lifecycle/idempotency/invalidation/retention tests; `make migration-drift` and `make backend-store` passed. |
| WS-06 | Internal consumer queries | completed | WS-05 | List/get/traverse and cursor tests pass. | Added lifecycle-state, cursor mutation/expiry, idempotency, and traversal determinism coverage; `make backend-store` passed. |
| WS-07 | Boundary, security, and integration guards | completed | WS-06 | Import-boundary and no-public-contract-change tests pass. | Added import/public-contract/source-table/redaction guards; `make backend-unit`, `make lint-go`, `make go-gosec-targeted`, and `make generated-artifact-policy-check` passed. |
| WS-08 | Final validation and handoff | completed | WS-07 | Final validation and handoff evidence recorded. | Final narrow and fast gates passed. Broad `make check` rerun failed in unrelated Phase 2 browser assertion; blocker and run root recorded below. |

## 3. Validation Policy

Narrow validation runs before broad validation:

| Change Family | Required Validation |
| --- | --- |
| Spec/docs | `make lint-markdown` |
| JSON contracts | `make json-shape-check`, `make generated-artifact-policy-check` |
| SQL migrations | `make migration-drift` |
| Go implementation | `make backend-unit`, `make backend-store`, `make lint-go` |
| Security/boundary | `make go-gosec-targeted`, import-boundary tests |
| Final handoff | `make agent-finalize`, `make test-fast`, risk-appropriate `make check` |

Skipped checks must be recorded with a concrete reason and any available run root.

## 4. Handoff Log

| Time | Agent | Workstream | Files / Areas | Commands | Result | Blockers / Residual Risk | Next Action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-01T14:27:40-04:00 | Codex | WS-00 | `docs/handoffs/graph-projection-activation-tracker.md` | `git status --short`, `date -Iseconds` | Tracker created and WS-00 marked complete. | None. | Start WS-01 specification activation. |
| 2026-07-01T14:29:37-04:00 | Codex | WS-01 | `docs/graph_projection_nlspec.md`, Core 00, Core 04, Appendix I, `docs/domain.md`, projection trackers. | `rg`, `make lint-markdown`, `date -Iseconds` | Graph Projection NLSpec is now `adopted/current` for graph projection only; authority docs and active tracker rows agree; markdown lint passed. | Historical tracker log rows still describe prior draft status at the time they were written. | Start WS-02 conformance matrix and fixtures. |
| 2026-07-01T14:33:51-04:00 | Codex | WS-02 | `contracts/graph-projection/conformance_matrix.v1.json`, `contracts/graph-projection/fixtures/corpus.v1.json`, `tools/schemas/cartulary.graph_projection_conformance_matrix.v1.schema.json`, `scripts/check-json-shapes.mjs`. | `make json-shape-check`, `date -Iseconds` | Matrix maps `GP-AC-001` through `GP-AC-068`; fixture registry maps `GP-FIX-001` through `GP-FIX-023`; JSON shape check passed with run root `.cartulary/test-results/20260701T183345Z-p2092726`. Later WS-08 reconciled matrix selectors to implemented test names. | None. | Start WS-03 engine foundation. |
| 2026-07-01T14:44:45-04:00 | Codex | WS-03 / WS-04 | `internal/modules/graphprojection/*`. | `gofmt`, non-canonical local probe `go test ./internal/modules/graphprojection`, `make backend-unit`, `date -Iseconds` | Added internal graphprojection runtime foundations and pure projection semantics. Package tests cover admission, duplicate members, graph-view IDs, direct/reverse edge output, aggregation, filters, canonical JSON, scalar/field validation, and failed-run shape. `make backend-unit` passed with run root `.cartulary/test-results/20260701T184428Z-p2096119`. | Pure engine tests are package-local; persistence and query APIs still pending. | Start WS-05 persistence, lifecycle, and retention. |
| 2026-07-01T14:51:29-04:00 | Codex | WS-05 | `db/migrations/00043_graph_projection_v1.sql`, `tools/migration_history_manifest.json`, `internal/modules/graphprojection/store.go`, `internal/modules/graphprojection/store_test.go`. | `gofmt`, non-canonical local probe `go test ./internal/modules/graphprojection`, `sha256sum db/migrations/00043_graph_projection_v1.sql`, `make migration-drift`, `make backend-store`, `date -Iseconds` | Added durable graph projection view/run/vertex/edge/idempotency tables and store lifecycle behavior. `make migration-drift` passed with run root `.cartulary/test-results/20260701T185031Z-p2100422`; `make backend-store` passed with run root `.cartulary/test-results/20260701T185047Z-p2102581`. | Store APIs are internal-only; broad backend/unit and boundary/security validation still pending. | Start WS-06 internal consumer query validation. |
| 2026-07-01T14:54:15-04:00 | Codex | WS-06 | `internal/modules/graphprojection/store.go`, `internal/modules/graphprojection/store_test.go`. | `gofmt`, non-canonical local probe `go test ./internal/modules/graphprojection`, `make backend-store`, `date -Iseconds` | Internal query APIs now cover graph-view retrieval, run inspection, listing with cursor expiry and deleted-anchor behavior, and deterministic traversal over self-loops, multi-edges, unknown seeds, filters, and closed metadata. `make backend-store` passed with run root `.cartulary/test-results/20260701T185335Z-p2113595`. | Query APIs remain internal-only; boundary/security validation still pending. | Start WS-07 boundary, security, and integration guards. |
| 2026-07-01T14:57:19-04:00 | Codex | WS-07 | `internal/modules/graphprojection/boundary_guard_test.go`. | `gofmt`, non-canonical local probe `go test ./internal/modules/graphprojection`, `make backend-unit`, `make lint-go`, `make go-gosec-targeted`, `make generated-artifact-policy-check`, `date -Iseconds` | Boundary guards prove graphprojection production code does not import sibling source/workbook modules, workbook/projections/public route roots do not import graphprojection, public OpenAPI/WS/generated protocol artifacts omit graph projection terms, store SQL only targets graph-derived tables, and operation errors do not echo source-authored values. `make backend-unit` passed with run root `.cartulary/test-results/20260701T185630Z-p2123829`; `make lint-go` passed; `make go-gosec-targeted` passed with run root `.cartulary/test-results/20260701T185709Z-p2129826`; `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260701T185716Z-p2137837`. | Broad final validation still pending. | Start WS-08 final validation and handoff completion. |
| 2026-07-01T15:10:01-04:00 | Codex | WS-08 | `contracts/graph-projection/conformance_matrix.v1.json`, `tools/schema_object_ownership_manifest.json`, generated `internal/gen/sql/models.go`, all graphprojection activation changes. | Selector existence probe, `make json-shape-check`, `make lint-markdown`, `make agent-finalize`, `make test-fast`, `make check`, `make generate`, `make generate-drift`, post-generation `make test-fast`, post-generation `make generated-artifact-policy-check`, `date -Iseconds` | Reconciled conformance matrix selectors to implemented tests. Added graphprojection schema-object ownership after `json-shape-check` initially failed for the new migration objects; rerun passed with run root `.cartulary/test-results/20260701T190031Z-p2139135`. `make lint-markdown` passed, including the final tracker edit. `make agent-finalize` passed with run root `.cartulary/test-results/20260701T190049Z-p2139897` and retained-run maintenance skipped because `RESULTS_DIR` was unset. First `make test-fast` passed with run root `.cartulary/test-results/20260701T190100Z-p2141291`. First `make check` failed with generated sqlc model drift at `.cartulary/test-results/20260701T190309Z-p2188724`; `make generate` passed with run root `.cartulary/test-results/20260701T190425Z-p2222420`, and `make generate-drift` then passed with run root `.cartulary/test-results/20260701T190434Z-p2223279`. Post-generation `make test-fast` passed with run root `.cartulary/test-results/20260701T190737Z-p2292098`; post-generation `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260701T191107Z-p2340295`. | Second `make check` failed at `.cartulary/test-results/20260701T190440Z-p2224270` in `browser-e2e-webserver-backed/browser-functional-shard-04`, test `phase2.spec.ts` E-2-02: expected `incident-admin-status` to contain `synced`, received `Saved promoted incident fields.` This is a pre-existing/public incident-shell browser assertion path and does not import or exercise graphprojection. | Handoff complete; next action is owner/frontend follow-up on the Phase 2 browser assertion if broad `make check` must be made fully green. |
