# graph-projection Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current planning posture |
| --- | --- |
| Target path | `internal/modules/graphprojection` |
| Target label | `graph-projection` (normalized lowercase kebab case; no spaces, path separators, shell metacharacters, or unsafe filename characters) |
| Output path | `docs/handoffs/graph-projection-module-refactor-tracker.md` |
| Baseline | Clean `main` at `aa0f9dad6dea184ba84abd45d25029e7c0e7638b` on 2026-07-12. |
| Status | Implementation remediation complete for the GP-FIX-001 through 013, 018 through 021, and 024 through 035 gaps under the 2026-07-12 execution authorization; GP-AC conformance publication remains intentionally unpromoted. |
| Allowed changes | Graph Projection implementation, owner inputs, tests, harness accounting, migrations, approved Reporting/Network Flow adapters, derived generated outputs, and this handoff. |
| Non-goals | Authorization remains caller-owned; Graph Projection does not mutate source/workbook state and gains no public route. |
| Default behavior posture | Adopted owner behavior takes precedence over known nonconforming implementation behavior. |
| Later authorization | No longer required for the remediation slices authorized by the execution request. Conformance publication remains prohibited until executable evidence supports each claim. |

Source hierarchy used for this tracker:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. `docs/domain.md` and implementation-support guides for terminology, package boundaries, and harness mechanics.
5. Current repository code and tests for current implementation state.
6. Prior plans, handoffs, appendices, and the planning framework as evidence only.

Owner and supporting documents inspected:

- `docs/graph_projection_nlspec.md`, including scope, inputs, outputs, identity, validation, lifecycle, ephemeral projection, query contracts, and GP-AC-001 through GP-AC-069;
- `docs/network-flow-activity-nlspec.md` §14.4 through §14.6 for the Network Flow ephemeral adapter and response contract;
- `docs/reporting-subsystem-nlspec.md` for `graph_projection_refs[]`, completed projection binding, diagram selection, and Reporting reason codes;
- `docs/spec/00_document_set_status_and_precedence.md` through `docs/spec/04_security_deployment_and_conformance.md` for authority, modular-monolith, derived-state, workbook, security, and conformance boundaries;
- `docs/domain.md` for the workbook-projection versus graph-projection distinction;
- `docs/testing-harness-nlspec.md` for Make-owned execution and evidence accounting;
- `docs/spec/I_projection_authority_boundary_and_characterization.md` as non-normative boundary evidence;
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` as the planning template and doctrine, not repository-state proof.

Core 05 was not used for conformance publication because all GP acceptance rows remain nonclaiming `planned` rows. No owner-document contradiction was found. The mismatches below are repository implementation or evidence-accounting findings, not `BLOCKED: owner contradiction` findings.

Repository evidence inspected includes every file under `internal/modules/graphprojection`; the Network Flow graph adapter, handler, route integration test, frontend client shape, and derived contracts; the Reporting release decoder, application authorization, store validation path, routes, integration shape, and OpenAPI surface; graph-projection migrations and ownership manifests; graph-projection conformance and fixture contracts; generated-artifact policy; backend boundary tooling; phase maps; execution topology; and the public Make target surface.

The planning-era implementation authorization boundary is superseded by Section 13. The completed remediation still does not publish Graph Projection conformance claims.

## 2. Current-State Repository Inventory

Every file currently present in `internal/modules/graphprojection` is inventoried below. No target file is out of scope.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/graphprojection/boundary_guard_test.go` | AST and source guards for sibling-module imports, workbook/public-route isolation, graph-table SQL scope, and safe error details. | Test-only guard helpers. | Go test runner; conformance matrix selectors indirectly name target tests. | Go parser/filesystem; OpenAPI and WebSocket contracts; target source files. | Self. | `contracts/openapi/cartulary.openapi.yaml`; `contracts/ws/index.schema.json`; boundary/evidence accounting. | `graphprojection` boundary tests | critical | Guards outbound sibling imports and public exposure, but does not restrict which sibling modules may consume the root facade. |
| `internal/modules/graphprojection/canonical.go` | Canonical JSON, tuple bytes, SHA-256 digests, generated IDs, and canonical value keys. | Private helpers only. | Admission and projection engine; target tests. | Go standard library only. | `engine_test.go`; indirectly all identity/output tests. | GP canonical bytes, ID, digest, and fixture claims. | `graphprojection` private engine | critical | Generic struct fallback uses ordinary `json.Marshal`; current structs have no JSON tags establishing NLSpec member names or order. |
| `internal/modules/graphprojection/engine_test.go` | Unit coverage for admission, IDs, normalization, direct/reverse/aggregate projection, filters, canonical JSON, scalar paths, and failed-run shape. | Test-only fixtures and helpers. | Go test runner; many conformance-matrix selectors. | Root package only. | Self. | `contracts/graph-projection/conformance_matrix.v1.json`; fixture IDs cited by that matrix. | `graphprojection` characterization tests | critical | A small number of broad tests are cited for many materially wider GP acceptance criteria. |
| `internal/modules/graphprojection/input.go` | UTF-8/JSON admission, duplicate-member rejection, top-level parsing, defaults, normalization, digest inputs, scalar validation, and graph/run ID derivation. | `AdmitOptions`, `AdmitProjectionInput`, `DeriveGraphViewID`. | `Project`; Network Flow directly calls `DeriveGraphViewID`; target tests. | Go standard library and private canonical/object helpers. | `engine_test.go`; `boundary_guard_test.go`; store tests through `Project`. | Graph Projection input, normalization, identity, error, and fixture contracts. | `graphprojection` private engine behind a facade | critical | Public low-level admission API exposes implementation-shaped `ProjectionRun`; many nested closed-schema and resource-limit owner rules are not characterized. |
| `internal/modules/graphprojection/objects.go` | Converts parsed mapping, aggregation, property, source, and filter values into digest/canonicalization objects. | Private helpers only. | `input.go`. | Graph Projection DTOs only. | Indirectly `engine_test.go`. | Digest transcripts and normalized-input contract. | `graphprojection` private engine | high | Duplicates observable wire-member knowledge outside a single contract serializer seam. |
| `internal/modules/graphprojection/project.go` | Synchronous projection orchestration, validation, direct and reverse mapping, aggregation, property/metadata derivation, schema registry, ordering, traversal helpers, and output digest. | `ProjectOptions`, `Project`. | Network Flow graph adapter; `Store.CreateProjection`; target tests. | Go standard library and root DTO/canonical/admission helpers. | `engine_test.go`; `store_test.go`; Network Flow route integration indirectly. | Graph Projection output, algorithms, validation, identity, and Network Flow ephemeral response. | `graphprojection` private derivation engine behind a facade | critical | A 1,149-line mixed algorithm file; `Project` returns retained-run-shaped state and is used as a substitute for the missing `project_ephemeral` interface. |
| `internal/modules/graphprojection/reporting_refs.go` | Validates Reporting release tuple projection references against retained run rows in a caller transaction and maps four Reporting reason codes. | Reporting reason constants; `ReportingProjectionRef`; `ReportingProjectionRefError`; `ValidateReportingProjectionRefsTx`. | `internal/modules/reporting/store.go`. | `pgx`, `pgtype`, direct SQL over `graph_projection_runs`. | `store_test.go`; Reporting release flows indirectly. | Reporting `source_projection_ref.v1`, release errors, and graph binding contract. | Generic retained projection lookup in `graphprojection`; Reporting-specific mapping in `reporting` | critical | Public Graph package API exposes `pgx.Tx` and Reporting vocabulary; same-transaction consistency must be preserved before changing the seam. |
| `internal/modules/graphprojection/store.go` | Retained create/refresh, persistence, replacement, retention pruning, run/view reads, listing/cursors, traversal, invalidation, and idempotency. | Store errors; `Store`; lifecycle/query option and result types; `NewStore`; create/refresh/get/list/traverse/invalidate methods. | No production `NewStore` assembly was found; target store tests exercise it directly. Reporting reads the same tables through `reporting_refs.go`. | `internal/platform/postgres`, `pgx`, direct SQL, root engine and DTOs. | `store_test.go`; boundary SQL guard. | Migration `00020_graph_projection.sql`; schema ownership; lifecycle, query, retention, and conformance contracts. | `graphprojection` application facade plus private PostgreSQL adapter | critical | Legitimate derived-state ownership, but application coordination and persistence are combined. Refresh aliases create, and current lifecycle/query behavior differs materially from the adopted NLSpec. |
| `internal/modules/graphprojection/store_test.go` | PostgreSQL-backed lifecycle, pre-admission no-touch, state reads, Reporting ref validation, idempotency/list pagination, traversal, and invalidation coverage. | Test-only fixtures and helpers. | Go test runner; conformance-matrix selectors. | `internal/testutil/pgtest`, `pgx`, root package, migration-created graph tables. | Self. | Migration and many GP lifecycle/query acceptance claims; Reporting ref behavior. | `graphprojection` store/integration tests | critical | Current Make target plans do not select this package, so these tests are not Make-owned evidence today. |
| `internal/modules/graphprojection/types.go` | Shared request, config, mapping, lifecycle, graph, schema registry, vertex/edge, validation, capability, and invalidation data types. | `ProjectionSchemaID`; run/view states; errors; all request, mapping, graph, validation, query, and invalidation DTOs. | All target production files; Network Flow consumes selected constants/run fields through `Project`. | Go `time` only. | All target tests. | Graph Projection wire and state vocabulary; Network Flow adapter surface. | `graphprojection` facade DTOs plus private engine model | critical | One broad exported type surface exposes internal representation; types lack explicit JSON tags and mix retained, ephemeral-adjacent, query, and persistence concerns. |

Only two inbound production imports of the root package were found: `internal/modules/networkflow/graph.go` and `internal/modules/reporting/store.go`. No application assembly or public graph-projection route registers `Store` directly.

## 3. Module Boundary Diagnosis

Architectural finding: `internal/modules/graphprojection` is a legitimate graph-oriented derivation subsystem under the adopted Graph Projection NLSpec. It is not the workbook-grid `projections` module and must not absorb Timeline, entity, indicator, evidence, link, import, saved-view, workbook, frontend-shell, or grid-vendor responsibilities. Internally it is a mixed-responsibility package: a view/projection orchestration layer, mutation-free derivation engine, retained lifecycle coordinator, persistence-adjacent adapter, query implementation, and Reporting-specific transaction adapter.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Graph input admission, normalization, canonical identity, and deterministic derivation | `input.go`, `canonical.go`, `objects.go`, `project.go` | `graphprojection` private engine behind a root facade | split | Graph Projection NLSpec §§4-9; Network Flow consumes the engine | Keep graph semantics in this module; reduce the exported low-level surface. |
| Retained graph-view lifecycle and query orchestration | `store.go` | `graphprojection` application/service facade | keep, then split internally | Graph Projection NLSpec §§10-11 | Owner-valid responsibility, but current implementation is not a complete owner-aligned facade. |
| PostgreSQL graph table persistence | `store.go`, `reporting_refs.go` | Private `graphprojection` persistence adapter | split | Migration `00020`; schema ownership manifest | Direct SQL is limited to graph-owned derived tables, which is intentional. |
| Reporting release-tuple reason mapping | `reporting_refs.go` | `reporting`, consuming a generic Graph Projection lookup/validation port | move, after characterization | Reporting NLSpec; `internal/modules/reporting/store.go` | Preserve same-transaction visibility and exact public Reporting reasons. |
| Network Flow ephemeral invocation | `internal/modules/networkflow/graph.go` calling root `Project` | Graph Projection ephemeral facade; Network Flow owns input and output mapping | defer | Graph Projection §10.9; Network Flow §14.4 | Correcting the full result changes observable HTTP data and requires later authorization. |
| Authorization and incident visibility | Network Flow and Reporting handlers/application services | Calling owner modules | keep outside | Graph Projection excludes authorization; callers enforce membership/role | The Graph module must not infer incident or release authorization. |
| Source entity, relationship, Timeline, import, evidence, link, or indicator mutation | Not present | Respective source-owner modules | intentional/no action | Boundary tests; NLSpec source-data non-scope | Graph Projection consumes explicit source input and must remain read-derived. |
| Workbook projections, saved views, view schemas, and revision/change-set refresh | Not present | `projections`, `workbook`, `savedviews`, `revisions`, and source owners | intentional/no action | Domain vocabulary; Graph Projection scope exclusion | Do not merge graph and workbook projection rules. |
| Frontend shell/controller state and grid-vendor integration | Not present | `apps/web` and `packages/grid-adapter` | intentional/no action | No target imports or files; only indirect Network Flow response typing | No frontend or grid movement is supported by target evidence. |
| WebSocket collaboration events | Not present | `collaboration` and platform WebSocket transport | intentional/no action | No Graph Projection term in WebSocket contract or collaboration code | The adopted Graph Projection NLSpec adds no WebSocket messages. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Graph Projection input, normalization, defaults, validation, IDs, digests, mapping, aggregation, ordering, and graph output | Graph Projection NLSpec | `input.go`, `canonical.go`, `objects.go`, `project.go`, `types.go`; GP contract matrix/fixtures | `engine_test.go` | Realize each cited fixture and acceptance family independently; assert canonical wire member names/bytes, closed schemas, limits, issue identity, and output parity before movement | critical | Current tests are useful implementation evidence but insufficient for the matrix's broad implemented claims. |
| Retained create/refresh lifecycle, operation outcomes, idempotency, concurrency, replacement, invalidation, and retention | Graph Projection NLSpec §10 | `store.go`; migration `00020`; store tests | Lifecycle, no-touch, state-matrix, pagination, invalidation tests | Full state-by-operation matrix, accepted response versus terminal inspection, run-specific invalidation, active-run races, duration/count expiry, failed refresh, and invalidated in-flight completion | critical | Current create/refresh behavior and envelopes differ from the owner contract. Preserve current behavior in refactor-only slices; correct separately. |
| Graph read queries: get view/run/vertex/edge, traverse, list and cursor semantics | Graph Projection NLSpec §11 | `Store` methods and query DTOs | State matrix, traversal, list pagination | Missing direct vertex/edge queries; closed error envelopes; all lifecycle states; seed/filter/limit bounds; cursor mutation and precedence fixtures | critical | The current API implements only a subset of owner query shapes. |
| Network Flow graph query HTTP route and exact ephemeral projection result | Network Flow NLSpec §14; Graph Projection §10.9 for embedded result | `internal/modules/networkflow/graph.go`; Network Flow route/schema/error contracts; frontend client | `TestNetworkFlowGraphContributorsAndIndicatorLinkRoutes`; Phase 12 selectors | Exact `project_ephemeral` request/outcome mapping, full ephemeral result, no retained state, cancellation/timeout/unavailable mapping, authorization, and client compatibility | critical | Current adapter calls `Project` and returns a reduced result with a retained run ID used as `ephemeral_projection_id`. |
| Network Flow graph route authorization and audit | Network Flow/Core security owners | `handleGraphQuery` authentication, membership check, and audit call | Route integration and Phase 12 security evidence | Preserve member denial, audit ordering, no partial response, and safe error details when changing the facade | high | Authorization remains outside Graph Projection. |
| Reporting release `graph_projection_refs[]` shape, sorting, uniqueness, completed/stale/digest binding, and reason mapping | Reporting NLSpec | Reporting API/store/application service; OpenAPI; `reporting_refs.go` | Reporting decoder/redaction/resource tests; Graph store reason matrix; Phase 11 integration | Same-transaction committed/uncommitted visibility, no/multiple binding, available/replaced/computing/failed/invalidated states, every digest field, authorization, and public error mapping | critical | Generic Graph lookup and Reporting-specific policy are currently fused. |
| Graph Projection storage semantics | Graph Projection lifecycle owner; physical layout implementation-private | Migration `00020`; `store.go`; schema ownership manifest | Store lifecycle tests and SQL boundary guard | Transactional publish/replacement, rollback, foreign-key cleanup, idempotency scope/expiry, and retention bucket parity | critical | Tables are graph-owned derived state; no migration is planned for the refactor. |
| Generated and derived contract surfaces | Owning NLSpecs and authored contracts | Graph conformance matrix/corpus; Network Flow schemas/routes/errors; OpenAPI; generated protocol TypeScript | `json-shape-check`; owner contract tests | Drift and semantic parity when owner inputs change; no generated hand edits | high | Graph Projection itself adds no generated protocol or public transport contract, but its callers expose derived contracts. |
| Harness/test accounting | Testing Harness NLSpec | Make target plans, phase maps, execution topology, conformance selectors | Phase 11/12 rows exist for callers | Add Make-owned Graph engine/boundary rows to unit evidence and store rows to service-backed evidence before relying on them | critical | Current `backend-unit`, `backend-store`, `backend-integration`, and `backend-process` target plans contain no graphprojection row. |
| WebSocket events | Collaboration/Core owners | Search of WebSocket contract and collaboration/platform code found no Graph Projection surface | Target guard forbids public Graph routes/WS terms | None unless a later owner change adds a socket contract | low | Not applicable to this refactor. |
| Entities rows/queries/mutations, saved views/view schemas, revisions/change sets, projection refresh for workbook rows, and test-util mutation behavior | Their respective Core/module owners | Explicit Graph Projection NLSpec exclusions and absence from target | Existing owner suites outside target | None for graph-only internal slices | low | Not applicable; do not import these semantics into Graph Projection. |
| Frontend shell, UI selectors, and grid adapter | Web app and package owners | Only indirect Network Flow response typing was found | Network Flow frontend tests | Phase 12/frontend evidence only if the public embedded result changes | medium | No direct target ownership or vendor integration. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| The root package is a legitimate subsystem but combines engine, application lifecycle, persistence, query, and consumer-specific adapter concerns. | All ten target files; Graph Projection NLSpec scope | critical | `must_fix` | `graphprojection`, split behind one facade | Characterize first, then separate private engine and persistence seams without changing outputs. |
| Network Flow consumes low-level engine/run types instead of a closed ephemeral facade. | `internal/modules/networkflow/graph.go` calls `DeriveGraphViewID` and `Project` and inspects `ProjectionRun`. | critical | `must_fix` | Graph ephemeral facade plus Network Flow adapter | Preserve the current route during refactor; schedule owner-aligned result correction separately under S-06. |
| Reporting-specific reason codes and a public `pgx.Tx` parameter live in Graph Projection. | `reporting_refs.go`; Reporting caller | critical | `should_fix` | Generic Graph retained-ref port; Reporting mapping | Block movement until same-transaction characterization defines a safe seam. |
| Application lifecycle and direct SQL share `store.go`. | `Store.CreateProjection`, persistence helpers, queries, cursor, traversal, invalidation | high | `should_fix` | Root service plus private PostgreSQL adapter | Extract persistence only after store characterization; keep graph table ownership. |
| Direct SQL accesses only graph-owned derived tables. | `store.go`, `reporting_refs.go`, SQL guard, migration, ownership manifest | medium | `intentional/no_action` | `graphprojection` persistence | Preserve table scope; do not introduce source/workbook SQL. |
| Root production code imports platform PostgreSQL abstractions. | `store.go` imports `internal/platform/postgres`. | medium | `should_fix` | Private persistence adapter | Keep platform dependency below the application facade. |
| Target production code imports no sibling domain module and consumes explicit input bytes. | Production import guard and import scan | low | `intentional/no_action` | `graphprojection` | Preserve the no-sibling-import direction. |
| Current boundary guards do not constrain approved inbound root importers. | Only Network Flow and Reporting import root; guard scans other selected roots only for forbidden imports. | high | `should_fix` | Graph boundary policy | Add an exact facade/importer policy when facade paths stabilize. |
| Current Go models and serializer fallback do not prove canonical snake-case wire output. | `types.go` lacks JSON tags; `canonical.go` marshals unknown structs with `json.Marshal`. | critical | `must_fix` | Graph contract serializer | Add characterization before extraction; correct observable wire behavior only under S-06 authorization. |
| The adopted ephemeral operation and full result are absent. | Graph Projection §§5.1.1/10.9 versus `Project`; Network Flow reduced result | critical | `must_fix` | Graph Projection facade | Treat as conformance remediation, not a behavior-preserving refactor detail. |
| Retained lifecycle/query behavior is partial and diverges from the adopted owner. | `store.go` compared with NLSpec §§10-11 | critical | `must_fix` | Graph Projection service | Record current behavior, avoid accidental change during structural slices, and isolate corrections in S-06. |
| Conformance matrix coverage statuses overstate executable evidence. | Most GP criteria are `implemented`; selectors reuse a few broad tests; GP-AC-069 alone is `planned`. | critical | `must_fix` | Graph contract/evidence owner | Audit every status and selector against executable fixtures before claiming conformance. |
| Make-owned target plans omit target tests. | `make explain-target` and `make target-plan-json` searches found no graphprojection rows. | critical | `must_fix` | Harness owner with Graph test owner | Resolve RB-002 before implementation validation claims. |
| Generated-file hand-edit risk is indirect through Network Flow/OpenAPI outputs. | Generated artifact policy and generated protocol surface | high | `intentional/no_action` | Contract generators | Change owner inputs and run generators; never edit generated roots. |
| No direct grid-vendor, view-schema, saved-view, collaboration, or test-only production leakage was found. | Target imports and repository searches | low | `intentional/no_action` | Existing owners | Keep these concerns out of scope. |

## 6. Refactor Workstreams

These workstreams describe planning completion. They do not authorize implementation.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Fix target, authority, baseline, constraints, and session history. | Tracker only | Git baseline and source list recorded | Scope, sole-write rule, and later-authorization statement are explicit. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every target file, symbol surface, caller, dependency, test, contract, and table. | All ten target files | File count and import/caller searches | Every target file has one inventory row. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map Graph, Network Flow, Reporting, Core, domain, generated, and storage owners. | Owner docs; caller routes/contracts | Owner sections cited; contradictions checked | Each observable contract has one owner and posture. |
| WF-03 | Characterization test gap analysis | parallel | WF-02 | WF-05, WF-07 | Compare tests and matrix selectors with actual owner criteria. | Target tests; caller tests; conformance matrix/corpus | Target-plan and phase-map inspection | RB-001/RB-002 and required pre-move tests are explicit. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Classify imports, SQL, platform, consumer-specific, generated, and unrelated-module coupling. | Target; Network Flow; Reporting; boundary tooling | Import and SQL searches | Every material finding has a classification and proposed owner. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Define root facade, private engine/store seams, caller ports, and deferrals. | Target package; two callers | Contract parity and boundary plan | Behavior-preserving versus behavior-changing work is separated. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order the smallest safe moves with rollback and stop conditions. | Tracker only at planning stage | Section 7 completeness review | S-00 through S-07 have dependencies and completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Define Make-owned evidence needed before and after each slice. | Future harness owner inputs; tests/contracts | Make target discovery and Phase 11/12 guidance | Missing commands are marked TODO rather than invented. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Preserve a restartable implementation handoff with risks and commands. | Tracker only | Tracker-only checks; future implementation gates | Top tracker, blockers, session logs, and next actions are current. |

WF-00 through WF-08 are `DONE` for this planning session. Future code, test, contract, and harness implementation remains pending or blocked as recorded below.

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Add owner-aligned characterization and Make-owned accounting before moving production code. | Target tests; conformance contract; harness owner inputs to be identified | Evidence changes must not be represented as runtime architecture or conformance without execution. | Split engine/boundary and store/lifecycle evidence; realize fixture expectations rather than citing labels alone. | `TODO: harness owner must map Graph tests`; then `make backend-unit`, `make backend-store`, `make json-shape-check` | Revert accounting/status changes together if target selection or evidence meaning drifts. | RB-001 and RB-002 are resolved; target tests execute through public Make targets with honest matrix statuses. |
| S-01 | S-00 | Extract canonicalization, admission, normalization, and derivation behind a private engine seam while retaining root compatibility wrappers and current byte/state behavior. | `canonical.go`, `input.go`, `objects.go`, `project.go`, `types.go`; future private engine package | IDs, digests, validation order, struct serialization, output ordering, and Network Flow caller behavior | Golden/current byte parity; IDs/digests; validation issue ordering; Network Flow adapter parity | `make backend-unit` after S-00 accounting | Restore the root implementation and wrappers as one slice; no caller change is required to roll back. | Root wrappers return exactly the pre-slice current results and no new sibling/platform import enters the engine. |
| S-02 | S-00, S-01 | Separate retained lifecycle coordination from PostgreSQL persistence behind a private store interface without changing tables, transactions, errors, cursors, or lifecycle outcomes. | `store.go`, `types.go`; future private persistence package | Atomic publish, replacement, retention, idempotency, cursor encoding, traversal order, and rollback | Existing store tests plus transaction/publish and retention parity characterization | `make backend-store` after S-00 accounting; `make migration-drift` only if migration ownership is implicated | Keep schema unchanged; revert service/store split together if transaction behavior drifts. | Root lifecycle API behavior and SQL table set are unchanged; platform/pgx dependencies sit below the facade. |
| S-03 | S-00, S-02 | Replace Reporting-specific Graph coupling with a generic completed-projection lookup/validation port; keep Reporting-owned reason mapping and same-transaction visibility. | `reporting_refs.go`; `internal/modules/reporting/store.go`; application assembly if injection is needed | Release admission atomicity, available/replaced eligibility, snapshot and digest matching, public reason codes | Same-transaction uncommitted visibility and full Reporting reason matrix | `make phase-slice PHASE=phase11`; `make backend-module-boundary-check` | Retain the existing Tx validator until the new port passes parity; rollback by restoring the old adapter call. | RB-003 is resolved and Reporting public behavior is byte/reason identical with no public `pgx.Tx` Graph facade. |
| S-04 | S-01, S-02, S-03 | Harden inbound/outbound boundary guards so callers use only the intended facade and workbook/public transport packages remain isolated. | `boundary_guard_test.go`; backend boundary owner inputs if approved | Overbroad allowlists could legitimize unwanted coupling. | Exact allowed importer/facade tests; existing no-route, no-sibling, graph-table, and safe-error guards | `make backend-module-boundary-check`; `make backend-unit` after S-00 | Revert only the policy change if it blocks legitimate existing callers; do not add wildcard exceptions. | Only approved Network Flow/Reporting or composition adapters can consume the facade; engine/store internals have no external imports. |
| S-05 | S-00 through S-04 | Reconcile conformance matrix selectors, executable fixture evidence, and harness rows after structural moves. | Graph contracts, target tests, owner-approved harness inputs | False implemented claims, stale selectors, generated drift, and phase-shaped runtime coupling | Every retained implemented status points to executable owner-aligned evidence. | `make json-shape-check`; `make generated-artifact-policy-check`; `make generate-drift` if owner inputs change | Revert status/selector changes with their corresponding tests; never hand-edit generated outputs. | Matrix, fixtures, Make selection, and code paths agree without treating phase rows as architecture. |
| S-06 | S-00, S-01, S-02, S-05 | `requires later authorization`: implement full `project_ephemeral`, owner lifecycle envelopes/admissibility, missing queries, exact retention/invalidation, canonical wire output, and Network Flow response expansion. | Graph facade/engine/store/types; Network Flow adapter/contracts/client; tests and owner-derived contracts | Intentional observable changes across internal Graph APIs and public Network Flow HTTP data | Full GP fixture corpus; lifecycle/query matrices; Network Flow outcome mapping; Phase 11/12 compatibility as applicable | `make phase-slice PHASE=phase12`; `make phase-slice PHASE=phase11` when Reporting is affected; all S-00 targets | Each behavior correction must be independently revertible behind the characterized facade; no mixed structural/behavior mega-change. | Explicit behavior authorization exists and owner-aligned acceptance tests pass without retained ephemeral state or compatibility ambiguity. |
| S-07 | S-04, S-05; S-06 only if authorized | Remove obsolete compatibility wrappers after all internal callers migrate; finalize documentation and validation. | Root facade, private packages, callers, tracker/handoff | Premature removal can break internal consumers or evidence selectors. | Import scan; all affected characterization and phase evidence | `make agent-finalize` then `make check` | Keep wrappers until the last verified caller is gone; restore wrapper-only compatibility if cleanup fails. | No orphaned old path, stale selector, out-of-scope diff, or undocumented deferral remains. |

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-unit` | Pure Graph engine, admission, canonicalization, DTO, and boundary evidence after S-00 accounting | yes | Current target plan does not select `internal/modules/graphprojection`; `TODO: harness accounting required` before this command proves the target. |
| integration | `make backend-store`; `make phase-slice PHASE=phase11`; `make phase-slice PHASE=phase12` | Graph PostgreSQL behavior, Reporting release binding, and Network Flow public adapter behavior | yes for the affected slice | Use `backend-store` after S-00; use Phase 11 only for Reporting seams and Phase 12 only for Network Flow seams. |
| e2e/browser | `make phase-slice PHASE=phase11` or `make phase-slice PHASE=phase12` | Owner-aligned browser/public-route evidence selected by the affected caller phase | no for internal-only slices | Required when a caller's public response, route behavior, or client contract changes. Do not run browser suites merely for private file movement. |
| generated drift | `make json-shape-check`; `make generated-artifact-policy-check`; conditionally `make generate-drift` | Authored contract shapes, generated-file policy, and generated outputs | yes when contracts/evidence inputs change | Update owner inputs and run generators; never hand-edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check` | Backend ownership and facade dependency direction | yes | Target-local guards must also be selected through S-00 accounting. |
| full check | `make agent-finalize` then `make check` | Final maintenance and broad repository verification | no before implementation; yes before final implementation handoff | Supply `RESULTS_DIR` only for a valid retained successful warm-check root; otherwise report retained-run maintenance skipped. |

Tracker-only validation is `make lint-markdown`, `make generated-artifact-policy-check`, and `make json-shape-check`. No production test suite is required for this documentation-only change.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| GPRT-001 | Fix graph-projection scope, authority, sole-write rule, and baseline | scope | DONE | none | Section 1 | Target, label, output, baseline, and later-authorization boundary are explicit. |
| GPRT-002 | Inventory every target file and direct caller | discovery | DONE | GPRT-001 | Section 2; ten inventory rows | Every target file appears once and both production importers are named. |
| GPRT-003 | Diagnose the module boundary | architecture | DONE | GPRT-002 | Sections 3 and 5 | Legitimate owner responsibilities and misplaced/mixed seams are classified. |
| GPRT-004 | Map observable contracts and owners | contracts | DONE | GPRT-002 | Section 4 | Every applicable contract has an owner, evidence, test posture, and risk. |
| GPRT-005 | Record repository/owner and evidence mismatches | conformance planning | DONE | GPRT-004 | RB-001, RB-002, RB-004 | Mismatches are not mislabeled as owner contradictions or silently resolved. |
| GPRT-006 | Establish owner-aligned characterization and Make accounting | tests/harness | DONE | GPRT-005 | Phase-neutral subsystem manifests; RB-001/RB-002 | Graph tests execute through public Make targets and matrix claims do not overstate evidence. |
| GPRT-007 | Extract the private derivation engine | backend refactor | TODO | GPRT-006 | S-01 | Current IDs, bytes, validation, ordering, and callers remain unchanged. |
| GPRT-008 | Split retained service and PostgreSQL adapter | backend refactor | DONE | GPRT-006, GPRT-007 | GP-W0 repository port and `postgresstore` adapter | The root facade no longer imports PostgreSQL/pgx or exposes `Store`; lifecycle/query behavior remains Make-verified. |
| GPRT-009 | Redesign Reporting projection-reference seam | cross-module boundary | DONE | GPRT-006, GPRT-008 | Graph binding port and PostgreSQL transaction adapter | Same-transaction behavior and public Reporting reasons are characterized and preserved. |
| GPRT-010 | Harden facade/import guardrails | architecture | TODO | GPRT-007, GPRT-008, GPRT-009 | S-04 | Only approved facade imports remain and private seams cannot leak. |
| GPRT-011 | Reconcile matrix, fixtures, selectors, and harness evidence | contracts/harness | IN PROGRESS | GPRT-006, GPRT-010 | All 69 rows downgraded to `planned`; phase-neutral Make rows active; executable generation plan in Section 14 | Evidence claims are executable, owner-aligned, and Make-owned. |
| GPRT-012 | Correct owner-nonconformant observable behavior | behavior remediation | IN PROGRESS | GPRT-011 | Service facade, migration 32, lifecycle/query correction, full ephemeral response | Adopted Graph/Network Flow acceptance behavior passes with no retained ephemeral state. |
| GPRT-013 | Remove compatibility paths and complete implementation handoff | cleanup/validation | TODO | GPRT-010, GPRT-011; GPRT-012 if authorized | S-07 | No stale caller, selector, wrapper, or out-of-scope change remains; final gates pass. |
| GPRT-014 | Complete planning tracker and handoff | planning | DONE | GPRT-001 through GPRT-005 | Sections 1-12 | Another agent can begin S-00 without rediscovering scope, owners, risks, or commands. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-12 20:25 EDT | Codex planning/documentation session | Target exists; tracker created from a clean baseline; implementation remains unauthorized. | Inspected framework, owner NLSpecs, Core 00-04, domain, harness, Appendix I, target tree; touched only this tracker. | `sed`, `rg`, `find`, `wc`, `git status --short`, `git branch --show-current`, `git rev-parse HEAD`, `date`. | Authority and source posture recorded; no owner contradiction found. | RB-001 through RB-004 affect later implementation, not tracker creation. | Preserve the sole-write boundary and begin S-00 only under later authorization. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-12 20:25 EDT | Codex planning/documentation session | Legitimate graph subsystem with mixed engine, lifecycle, persistence, query, and Reporting adapter responsibilities. | All ten target files; Network Flow and Reporting callers; migration and schema ownership; tracker touched. | `find`, `sed`, `rg`, `wc`. | Two production callers found; no sibling-domain import from target; graph SQL remains graph-table scoped. | RB-001, RB-002, RB-003. | Begin S-00 only under a later authorized implementation task. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-12 20:25 EDT | Codex planning/documentation session | No frontend shell, selector, view-schema, saved-view, or grid-vendor responsibility exists in the target; Network Flow response typing is an indirect consumer. | Network Flow frontend client/test references; target imports; tracker touched. | `rg`, `sed`. | Frontend movement is unsupported; Phase 12 frontend evidence applies only to an authorized public response correction. | RB-004 for result expansion. | Keep internal refactor slices frontend-neutral. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-12 20:25 EDT | Codex planning/documentation session | Graph conformance artifacts are authored evidence inputs; Network Flow and Reporting expose derived/public contracts; generated roots are protected. | Graph conformance matrix/corpus, Network Flow contracts, OpenAPI, generated policy, generated protocol references; tracker touched. | `sed`, `rg`, `find`. | Matrix selectors/statuses require audit; no generated hand edit is planned. | RB-001, RB-004. | Resolve evidence accuracy in S-00/S-05 before any conformance claim. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-12 20:25 EDT | Codex planning/documentation session | Target tests exist but are absent from current public backend target plans; caller evidence is phase-accounted. | Target tests, conformance selectors, phase maps, task surface, execution topology; tracker touched. | `make help-all`; `make explain-target TARGET=backend-unit DETAIL=rows`; corresponding store/integration/process explains; `make target-plan-json TARGET=backend-unit`; `make task-guide ROLE=feature-dev PHASE=phase11`; same for phase12; `make explain-phase PHASE=phase11`; same for phase12; `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`. | No Graph Projection row found in backend target plans. No production tests ran. Markdown lint passed. Generated-artifact policy passed at `.cartulary/test-results/20260713T002956Z-p57241`. JSON shape failed at `.cartulary/test-results/20260713T002959Z-p57445` because unrelated file `docs/handoffs/network-flow-activity-adoption-handoff-tracker.md` is missing. | RB-001, RB-002; unrelated JSON-shape harness prerequisite is absent. | Obtain harness-owner mapping in S-00; resolve the missing external handoff in its owning task before rerunning JSON shape. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-12 20:25 EDT | Codex planning/documentation session | Graph Projection defines no source authorization; Network Flow membership/audit and Reporting role/visibility checks remain caller-owned. | Network Flow graph handler; Reporting application service/routes; Core 04; tracker touched. | `sed`, `rg`. | No authorization decision or source-data mutation belongs in Graph Projection. | None for private structural slices; RB-004 for observable corrections. | Preserve caller authorization and safe error evidence around any facade migration. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-12 20:25 EDT | Codex planning/documentation session | Planning is complete; no production refactor, test edit, contract edit, migration, generated edit, or harness edit was performed. | This tracker touched; all evidence paths above inspected. | Discovery and Make command-surface inspection listed above; tracker-only validation listed under Tests and harness. | Workflows, slices, validation, rollback, and blockers are decision-complete. Two tracker checks passed; JSON shape is blocked by a missing unrelated handoff file and was not repaired because of the sole-write constraint. | RB-001 through RB-004; unrelated missing handoff prerequisite for JSON shape. | Later authorize S-00; do not begin structural movement before evidence accounting is resolved. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | The old matrix overstated coverage and the 36-fixture corpus remains descriptive rather than fully executable. | Publication must not rely on metadata-only evidence. | All existing `implemented` claims were downgraded to `planned`; executable per-fixture realization remains required before re-promotion. | RESOLVED for honest status and structural work; OPEN for future conformance publication. |
| RB-002 | Graph tests were absent from Make-owned accounting. | Unselected tests could not support validation. | Phase-neutral subsystem registry and Graph unit/store/migration/binding maps now feed target planning. | RESOLVED; `backend-unit` and `backend-store` select and pass the rows. |
| RB-003 | Reporting needed same-transaction visibility without Graph-owned Reporting reasons or a root `pgx.Tx` facade. | Independent reads could observe stale state. | Immutable `ProjectionBinding`, narrow reader port, transaction-bound PostgreSQL adapter, and Reporting-owned reason mapping. | RESOLVED; Phase 11 and same-transaction tests pass. |
| RB-004 | Owner-aligned Graph/Network Flow corrections intentionally change observable behavior. | The original tracker did not authorize them. | The 2026-07-12 execution request explicitly authorized implementation of the remediation plan. | RESOLVED as an authority blocker; validation status is recorded in Section 13. |

## 12. Binary Completion Criteria

The planning tracker is complete only when all of the following are true:

- [x] Every file in `internal/modules/graphprojection` is inventoried; all ten current files have individual rows.
- [x] Every discovered public contract risk has an owner and test posture.
- [x] Every proposed workflow has dependencies, validation, and a handoff checkpoint.
- [x] Every structural implementation slice is behavior-preserving; S-06 is explicitly marked `requires later authorization`.
- [x] Validation commands are discovered, and the missing Make-owned target coverage is marked `TODO` with the searches performed.
- [x] No owner-document contradiction was found or silently resolved.
- [x] Repository/owner and repository/framework mismatches are recorded as planning findings.
- [x] Phase maps and test rows are treated as verification/evidence accounting, not runtime architecture.
- [x] No generated-file hand edit, migration, route change, or production patch is proposed as part of this tracker task.
- [x] WebSocket, workbook/test-util mutation, saved-view/view-schema, revisions, frontend shell, selector, and grid-adapter surfaces are explicitly classified as not applicable except for indirect caller compatibility.
- [x] RB-001 through RB-004 identify the evidence, harness, transaction, and authorization conditions required before later work.
- [x] The handoff records the baseline, inspected surfaces, commands, current state, blockers, rollback posture, and next action.

The original planning session changed only this tracker. Section 13 records the later authorized production, fixture, and validation work.

## 13. Implementation Remediation Update

This section supersedes planning-era statements that implementation was unauthorized or that only this tracker changed. The execution request authorized the owner-aligned corrections, including intentional internal API and Network Flow v1 response changes.

### Completed remediation

- Added phase-neutral subsystem test manifests and selected Graph engine/boundary evidence through `backend-unit`, plus lifecycle, migration, and transaction-binding evidence through `backend-store`. Target planning reports the rows as `subsystem:graphprojection`, not as a runtime phase.
- Downgraded every GP acceptance row to `planned`. No Graph Projection implementation-conformance claim remains while the fixture corpus is descriptive rather than executable.
- Added the typed `graphprojection.Service` entry point, explicit snake-case resource projection, closed canonical-value support, injected clock and nonce sources, full non-retained `ProjectEphemeral`, retained lifecycle methods, and the complete query method family.
- Added authenticated AEAD cursors bound to operation, query shape, visibility scope, and issuance time. Oversize, malformed, wrong-shape, and expired outcomes have distinct reason codes in normative precedence order.
- Added separately committed accepted and computing lifecycle transitions, atomic terminal publication, explicit selected-run tracking, create/refresh admissibility, replacement timestamps, independent retention buckets, graph/run invalidation, and invalidation dominance over in-flight publication.
- Added migration 32. It blocks when Reporting releases reference existing Graph runs, otherwise resets unverifiable derived Graph state and installs lifecycle columns, constraints, idempotency scope, active-run exclusion, selected-run identity, and direct query/retention indexes.
- Replaced the Reporting-specific Graph validator with immutable `ProjectionBinding` data and a transaction-bound PostgreSQL reader. Reporting owns eligibility, tuple comparison, and public reason mapping while retaining release-transaction visibility.
- Switched Network Flow to `ProjectEphemeral`, expanded its authored v1 contract to the full closed Graph result, regenerated Go/TypeScript contracts, and updated the web client and fixtures atomically.
- Added exact inbound boundary guards: Network Flow consumes the Graph service; Reporting consumes only the binding port/adapter; no public Graph route or source/workbook mutation was introduced.

### Deliberately unclaimed conformance

All 36 fixture registry entries now have materialized directories and focused test symbols selected through `backend-unit` or `backend-store`. The newly added GP-FIX-001 through GP-FIX-013, GP-FIX-018 through GP-FIX-021, and GP-FIX-024 through GP-FIX-035 artifacts are remediation fixtures: they load hash-checked manifests and assert the identified implementation gaps, but they do not yet contain the complete normalized forms, canonical transcripts, expected envelopes, and state snapshots required for publication-grade GP-AC evidence.

The GP-AC-001 through GP-AC-069 claim audit therefore remains conservative: all 69 rows have `coverage_status: planned`, and no nonplanned row cites documents, corpus entries, or matrix metadata as implementation evidence. Re-promoting a row requires a focused Make-selected assertion that proves its full acceptance sentence and every cited fixture. Canonical-byte criteria must compare the transcript before comparing its digest.

### Current validation evidence

| Target | Result | Evidence root |
| --- | --- | --- |
| `make backend-unit` | PASS, 231 tests | `.cartulary/test-results/20260713T132759Z-p76141` |
| `make backend-store` | PASS, 141 tests | `.cartulary/test-results/20260713T132819Z-p78209` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T131543Z-p49334` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260713T131549Z-p49685` |
| `make backend-module-boundary-check` | PASS | `.cartulary/test-results/20260713T131556Z-p49884` |
| `make migration-drift` | PASS | `.cartulary/test-results/20260713T131602Z-p50081` |
| `make target-plan-json` | PASS, emitted plan JSON | command output only |
| `make harness-contract` | PASS | command output only |
| `make generate-drift` | PASS with repo-local `TMPDIR`/Go cache | `.cartulary/test-results/20260713T131652Z-p56481` |
| `make lint-go-staticcheck` | PASS after removing dead helpers | command output only |
| `make go-gosec-targeted` | PASS after root-scoping fixture artifact reads | `.cartulary/test-results/20260713T132641Z-p54405` |
| `make agent-finalize` | PASS; generated outputs unchanged; `RESULTS_DIR` unset | `.cartulary/test-results/20260713T132851Z-p82542` |
| `make check` | PASS, 139/139 work units and 1,111 tests | `.cartulary/test-results/20260713T133417Z-p84457` |

### GP-W0 repository boundary completion

- Moved the retained PostgreSQL implementation and cursor codec to `internal/modules/graphprojection/postgresstore`; root `graphprojection` now contains the `RetainedRepository` port, domain query DTOs/errors, and the `Service` facade only.
- Removed the root `Store`, `NewStore*`, and `StoreProjectionOptions` surface. `ServiceOptions.Repository` is the single retained-state dependency. The transaction-bound Reporting binding adapter remains separate in `postgresbinding`.
- Converted the PostgreSQL lifecycle tests to an external package that constructs the adapter directly, so the root package cannot regain an implicit SQL dependency through tests.
- Added a root-facade import guard and kept the graph-table-only SQL guard against the child repository.

| Target | Result | Evidence root |
| --- | --- | --- |
| `make backend-store` | PASS, 141 tests | `.cartulary/test-results/20260713T024539Z-p94588` |
| `make backend-module-boundary-check` | PASS | `.cartulary/test-results/20260713T024636Z-p1478` |
| `make backend-unit` | PASS, 221 tests | `.cartulary/test-results/20260713T024637Z-p1632` |

### GP-W1 fixture contract and verifier foundation

- Added the closed `cartulary.graph_projection_fixture_manifest.v1` schema, registered it with the harness schema-attachment owner, and added a JSON-shape validator for every materialized Graph fixture directory.
- Added `internal/modules/graphprojection/fixturetest`, a test-only loader that rejects unknown fields, unsafe paths, symlinks, duplicate artifacts, and SHA-256 mismatches. It does not calculate expected Graph behavior.
- Added the explicit developer-only `make graph-projection-fixture-candidate FIXTURE=GP-FIX-NNN` wrapper. Candidate output is written only under the Make test-results root and cannot modify `contracts/`.
- Added the first two owner-derived digest manifests (GP-FIX-022 and GP-FIX-023); their behavior evidence is tracked in the next workstream.

| Target | Result | Evidence root |
| --- | --- | --- |
| `make phase-schedules` | PASS | `.cartulary/test-results/20260713T025232Z-p8709` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T025241Z-p8912` |
| `make backend-unit` | PASS, 224 tests | `.cartulary/test-results/20260713T025008Z-p6091` |
| `make graph-projection-fixture-candidate FIXTURE=GP-FIX-022` | PASS; disposable candidate only | command output only; no contract mutation |

### GP-W2 canonical and scalar core completion

- Materialized GP-FIX-014 through GP-FIX-017, GP-FIX-022, GP-FIX-023, and GP-FIX-036 as hash-checked fixture directories with focused Make-selected symbols.
- GP-FIX-022 and GP-FIX-023 compare the owner-published canonical tuple bytes before comparing the resulting SHA-256 digests.
- Fixture execution exposed a real conformance defect: digest objects were serialized through lexically sorted maps even where §5.8 requires schema-declared member order. Added the internal ordered canonical-object representation and applied it to config/source digest envelopes and their nested mapped objects. Both NLSpec digest goldens now pass.
- Scalar and canonical string fixtures remain non-claim-bearing until the GP-AC audit verifies each complete acceptance sentence.

| Target | Result | Evidence root |
| --- | --- | --- |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T030741Z-p50958` |
| `make backend-unit` | PASS, 231 tests | `.cartulary/test-results/20260713T030743Z-p51287` |

### Remaining handoff conditions

- GP-W3 through GP-W5 implementation remediation is complete at the assertion level: GP-FIX-001 through GP-FIX-013, GP-FIX-018 through GP-FIX-021, and GP-FIX-024 through GP-FIX-035 now have manifest directories, hash-checked artifacts, and focused Go test symbols.
- GP-FIX-034 is implemented through the closed §4.12 resource-limit registry and one focused overflow assertion per implemented registry row. Do not promote GP-AC resource-limit rows until publication-grade fixture artifacts include exact limit/observed/severity envelopes.
- The GP-AC-001 through GP-AC-069 claim posture audit found no promoted rows. All rows remain `planned`, which is intentional until full owner-derived expected artifacts and sentence-complete selectors exist.
- The earlier `generate-drift` `/tmp` failure is resolved for this workspace by using repo-local `TMPDIR`, `GOTMPDIR`, and Go cache directories. `make check` also required `GO_MOD_CACHE_DIR`/`GOMODCACHE` under `.cartulary/go-mod` and an outside-repository `CARTULARY_HARNESS_SCRATCH_ROOT` because harness smoke fixtures may not live inside the checkout.
- An intermediate `make check` run at `.cartulary/test-results/20260713T132858Z-p82733` timed out in unrelated Phase 1 browser stateful evidence while clicking the account-navigation menu; the dedicated `make browser-e2e-stateful` rerun passed at `.cartulary/test-results/20260713T133117Z-p65219`, and the final full `make check` passed at `.cartulary/test-results/20260713T133417Z-p84457`.

## 14. GP-FIX-001 Through GP-FIX-036 Generation Plan

### 14.1 Objective and current state

The fixture registry currently contains 36 stable IDs and descriptions but no executable payloads. This workstream will turn each ID into a self-contained, Make-selected conformance case without treating current implementation output as normative truth. Until an individual fixture and every criterion that cites it satisfy the gates below, the fixture remains non-claim-bearing and the associated matrix rows remain `planned`.

This plan changes evidence quality, not normative behavior. `docs/graph_projection_nlspec.md` remains the owner. If fixture construction exposes an ambiguity or contradiction, stop that fixture, record the exact owner question, and leave its criteria `planned`; do not resolve uncertainty by copying current behavior into the golden.

### 14.2 Authored fixture layout

Keep `contracts/graph-projection/fixtures/corpus.v1.json` as the stable index. Add one directory per fixture:

```text
contracts/graph-projection/fixtures/
  GP-FIX-001/
    fixture.json
    input.raw.json
    expected.error.json
  GP-FIX-022/
    fixture.json
    input.json
    expected.normalized.json
    expected.config-transcript.txt
    expected.source-transcript.txt
    expected.output-transcript.txt
    expected.digests.json
    expected.response.json
```

Only artifacts relevant to a fixture are required. `fixture.json` must use a closed schema and contain:

- `schema_id`, `fixture_id`, `input_kind`, and `owner_sections`;
- `execution_layer`: `backend_unit` or `backend_store`;
- deterministic `clock`, nonce, cursor-key, and identity inputs when applicable;
- ordered setup, operation/query, clock-advance, and assertion steps;
- relative paths to every input and expected artifact;
- expected state effects: graph views, runs, vertices, edges, idempotency records, or explicit `no_state_change`;
- exact acceptance-criterion IDs exercised by the fixture;
- golden provenance and reviewer status.

Raw malformed or duplicate-member input must be stored as bytes and must never be parsed and rewritten by a JSON formatter. Canonical transcripts use UTF-8 text files whose final-newline policy is explicit in `fixture.json`. Expected JSON objects must be closed owner-contract objects with exact nullability and member names. Time-bearing fixtures use a fake clock with explicit microsecond timestamps; nonce and cursor-key material is fixed fixture input, never random test state.

### 14.3 Golden-generation rule

The fixture runner has two strictly separated modes:

1. `candidate` mode may run the implementation and write candidate artifacts under a disposable result directory. It must never update `contracts/`.
2. `verify` mode is read-only. It loads committed authored artifacts, executes the operation, and compares actual results to them.

Candidate output is diagnostic input for review, not an oracle. Before committing a golden, the author must derive or independently verify it from the NLSpec. For digest fixtures, review canonical transcript bytes first, calculate SHA-256 from those committed bytes, and only then record the digest. A test must compare actual transcript bytes before comparing the digest so matching hashes cannot conceal the wrong serialization seam.

### 14.4 Runner and selector design

Add a shared test-only loader and assertion library under `internal/modules/graphprojection/fixturetest`. It owns closed manifest decoding, repository-root-safe path resolution, fake clock/nonce/cursor injection, state snapshots, exact JSON comparison, transcript comparison, and useful first-difference diagnostics. It must not contain Graph business rules or regenerate expected values during verification.

Use focused top-level tests rather than one broad corpus test. Each fixture receives one stable symbol, for example `TestGPFIX001MalformedJSON` and `TestGPFIX022MinimalGraphDigestTranscript`, delegating to the shared runner. Pure admission, derivation, scalar, canonicalization, and ephemeral cases live in a unit fixture test file. PostgreSQL lifecycle, retention, invalidation, and query sequences live in a store fixture test file.

Add every symbol to the phase-neutral Graph subsystem manifest. `backend-unit` selects pure fixtures; `backend-store` selects PostgreSQL fixtures. The conformance matrix cites these exact symbols and fixture paths. A fixture path, registry entry, documentation anchor, or matrix row is never sufficient implementation evidence by itself.

### 14.5 Fixture-by-fixture artifact plan

| Fixture | Layer | Authored inputs | Required expected artifacts and primary assertions |
| --- | --- | --- | --- |
| GP-FIX-001 | unit | Malformed raw UTF-8 JSON bytes | Exact pre-admission error; no run, digest, validation issue, or state effect. |
| GP-FIX-002 | unit | Raw JSON containing a duplicate member at the specified depth | Exact duplicate-member pre-admission error and path; byte-preserving input; no state effect. |
| GP-FIX-003 | unit | Otherwise valid projection input with a mismatched supplied `graph_view_id` | Derived ID transcript, exact invalid-ID error, and no admitted run or digest exposure. |
| GP-FIX-004 | store | Validly admitted input with an invalid mapping | Accepted/computing/failed sequence, exact failed-run envelope and validation summary, no consumable graph. |
| GP-FIX-005 | unit | Input producing more than the validation issue cap | Discovery-order candidate list, selected `N-1` issues, final cap issue, counts, ordering, and failed outcome. |
| GP-FIX-006 | unit | Aggregation contributors covering missing, defaulted, explicit-null, and present values | Normalized input, candidate table, emitted aggregate properties, validation issues, and canonical output. |
| GP-FIX-007 | unit | Contributors that conflict under `single_value` after defaults | Candidate order, exact merge-conflict issue, failed outcome, and no partial consumable output. |
| GP-FIX-008 | unit | `count` aggregation with an unusable `source_field_path` | Exact count showing candidate field evaluation is ignored; registry and output transcript. |
| GP-FIX-009 | unit | Relationship aggregation with a missing endpoint grouping key and `error` policy | Endpoint grouping transcript, exact validation issue, and failed/no-edge result. |
| GP-FIX-010 | unit | Same endpoint absence with `exclude` policy | Deterministic exclusion, no issue beyond the owner rule, and exact remaining graph output. |
| GP-FIX-011 | unit | Endpoint grouping digest that has no matching aggregate vertex | Grouping tuple/transcript, derived digest, exact unmatched-endpoint issue, and no dangling edge. |
| GP-FIX-012 | unit | Direct and aggregate mappings with defaults, mapping labels, and source labels | Exact schema registry, label arrays/order, and `source_labels_preserved` values. |
| GP-FIX-013 | unit | Wildcard property and metadata definitions across multiple kinds | Expanded registry entries, deterministic ordering, mapped values, and closed output members. |
| GP-FIX-014 | unit | Strings covering quote, reverse solidus, controls, and required Unicode cases | Exact canonical UTF-8 transcript and digest; no implementation-produced golden. |
| GP-FIX-015 | unit | Table of accepted and rejected integer lexical forms at boundaries | Per-case normalized scalar or exact rejection code/path; no float coercion. |
| GP-FIX-016 | unit | Valid/invalid calendar timestamps including leap and offset boundaries | Per-case normalized timestamp or exact rejection; calendar validity independent of lexical plausibility. |
| GP-FIX-017 | unit | Identifier cases containing each relevant Unicode whitespace boundary | Per-case acceptance/rejection and exact field attribution. |
| GP-FIX-018 | store | Create/refresh/invalidation idempotency sequence with fixed clock and keys | Original envelopes, exact replay envelopes, conflict errors, expiry-at-equality behavior, and scoped rows. |
| GP-FIX-019 | store | Available plus replaced retained runs followed by graph-view invalidation | Exact invalidation summary, sorted changed-run IDs, cascade states, metadata copy, and retention timestamps. |
| GP-FIX-020 | store | Setup/transition sequence reaching every graph-view summary state | Exact list summary for creating, available, refreshing, failed, and invalidated states and timestamps. |
| GP-FIX-021 | store | Replacement history where count expires a run before duration | Before/equal/after reads, rank transcript, physical/addressable state, and selected-run preservation. |
| GP-FIX-022 | unit | Minimal empty graph input with fixed nonce and timestamps | Normalized input, config/source/output canonical transcripts, all IDs/digests, exact envelope, and empty state arrays. |
| GP-FIX-023 | unit | One-host property graph input with fixed nonce and timestamps | Same full transcript set as 022 plus vertex/property/registry identities and output ordering. |
| GP-FIX-024 | unit | Valid outer input with a nested unknown member | Exact pre-admission error path; explicit proof of no run, digest, or validation summary. |
| GP-FIX-025 | unit | Admitted source item with invalid identifier content | Fixed config/source digests, one itemized issue, failed envelope, normalized source ordering, and no graph output. |
| GP-FIX-026 | unit | Duplicate valid source IDs in intentionally permuted input order | Normalized ordering transcript, full digests, exact fatal duplicate issue, and permutation-invariant result. |
| GP-FIX-027 | unit | Aggregate grouping key sourced from mapped metadata | Mapped-metadata intermediate, grouping tuple/transcript, grouping digest, aggregate identity, and successful output. |
| GP-FIX-028 | unit | Definition using `projected.metadata.mapping_rule_id` | Exact `invalid_field_path` issue, target/path/details, and failed outcome. |
| GP-FIX-029 | store | Refresh paused in computing, then graph-view invalidation, then completion | Observable accepted/computing states, invalidation commit, terminal invalidated run, copied invalidation metadata, and no consumable interval. |
| GP-FIX-030 | store | Invalidated view followed by a refresh that fails | Refreshing transition, retained failed run, return to invalidated, unchanged prior invalidation object, and selected-run behavior. |
| GP-FIX-031 | store | Invalidated view followed by create | Exact `invalid_operation/graph_view_already_exists`; no new run or idempotency/state mutation. |
| GP-FIX-032 | store | Replaced history, invalidation, and count pressure before duration | Bucket migration transcript, invalidated ordering/rank, exact expiry boundaries, and always-retained selected run. |
| GP-FIX-033 | store | Failed initial create followed by refresh | Exact `invalid_operation/no_consumable_prior_run`; unchanged failed view/run state. |
| GP-FIX-034 | unit | One minimal overflow case for every added limit family | Per-case limit key, limit/observed values, severity/scope, exclusion versus fatal behavior, and no oversized allocation. |
| GP-FIX-035 | unit | Authenticated cursor made simultaneously oversized, malformed, wrong-shape, and expired | Exact precedence chooses `cursor_token_too_long`; subsidiary cases prove malformed, wrong-shape, and expiry ordering. |
| GP-FIX-036 | unit | `distinct_sorted_array` values containing quotes and reverse soliduses | Candidate-value transcripts, canonical-byte sort order, final array transcript, and digest. |

### 14.6 Sequencing and dependencies

| Stage | Work | Depends on | Exit criteria |
| --- | --- | --- | --- |
| F0 — Fixture contract | Define the closed fixture schema, directory/path rules, provenance fields, candidate/verify separation, and shape validation. | Existing honest `planned` matrix | Invalid fixture shapes, unsafe paths, missing artifacts, duplicate IDs, and unregistered directories fail `json-shape-check`. |
| F1 — Runner foundation | Implement the shared read-only verifier, deterministic injection, transcript diffing, state snapshots, and focused test wrappers. | F0 | A deliberately passing sample and deliberately corrupted golden prove success and useful failure behavior through Make. |
| F2 — Canonical/scalar core | Realize 014–017, 022, 023, and 036 first. | F1 | Canonical bytes, scalar boundaries, IDs, and digests are independently reviewed and stable; these become dependencies for later expected artifacts. |
| F3 — Admission and validation | Realize 001–005, 024–026, 028, and 034. | F2 | Pre/post-admission boundary, normalization, issue construction/order/cap, and limit severity fixtures pass. |
| F4 — Derivation and registry | Realize 006–013 and 027. | F2, F3 | Aggregation, endpoints, mapped metadata, labels, registry expansion, ordering, and output transcripts pass. |
| F5 — Lifecycle and queries | Realize 018–021 and 029–035. | F1; F2 identity rules; conforming lifecycle repository | PostgreSQL sequences pass at exact time boundaries, including concurrency/invalidation dominance, retention buckets, replay, state summaries, and cursor precedence. |
| F6 — Criterion audit | Audit GP-AC-001–069 against exact focused selectors and fixture executions. | F2–F5 | Each promoted row proves its complete pass condition; partial rows stay `planned`; no row cites itself or metadata-only evidence. |
| F7 — Publication handoff | Reconcile registry, selectors, manifests, commands, and tracker; run final validation. | F6 | All 36 fixtures execute through Make, drift is clean, and any remaining planned criterion has a specific documented missing assertion. |

F2 precedes the larger fixture batches because identity, normalization, and canonical-byte goldens are inputs to trustworthy derivation and lifecycle expectations. F5 may be developed in parallel with F3/F4 after F1, but its final golden review depends on the F2 identity rules.

### 14.7 Review, promotion, and validation gates

A fixture is complete only when all of the following are true:

- its manifest passes the closed fixture schema and references only files inside its own directory;
- every input and expected artifact is committed and human-readable, except intentionally raw byte inputs;
- golden provenance names the NLSpec sections and records independent review;
- `verify` mode passes through `backend-unit` or `backend-store` and candidate mode is absent from normal test targets;
- mutating one expected field or transcript byte makes the focused test fail;
- stateful fixtures prove both expected rows and the absence of forbidden extra state;
- exact-boundary fixtures execute before, equal, and after cases where the NLSpec requires them;
- its subsystem manifest row and conformance selectors name the focused test symbol;
- no generated file was hand-edited.

An acceptance criterion may change from `planned` to `implemented` only after an owner audit confirms that its full sentence—not merely its cited fixtures—has Make-owned executable evidence. Criteria spanning multiple fixtures require all of them. GP-AC-050 and other corpus-level criteria cannot be promoted until all required transcript fixtures execute. GP-AC-068 and GP-AC-069 require focused implementation selectors; their current document/corpus selectors are not sufficient.

Validation for each stage uses the narrowest applicable public targets: `make json-shape-check`, `make backend-unit`, `make backend-store`, `make target-plan-json`, and `make harness-contract`. Run `make generated-artifact-policy-check`, `make generate-drift`, `make agent-finalize`, and `make check` at F7. Use `make phase-slice PHASE=phase11` or `phase12` only if fixture-driven corrections touch Reporting or Network Flow behavior.

### 14.8 Primary risks and controls

| Risk | Control |
| --- | --- |
| Golden copied from current behavior | Candidate output is disposable; committed expectations require NLSpec derivation and independent review. |
| Canonical digest hides wrong bytes | Compare committed transcript bytes before computing or comparing SHA-256. |
| Broad test symbol overstates coverage | One stable top-level symbol per fixture; criteria cite exact focused selectors. |
| Fixture runner becomes a second implementation | Loader/assertion code contains no Graph business rules and never calculates expected semantic outcomes. |
| Time or randomness makes fixtures flaky | Inject clock, nonce, cursor key, and executor scheduling; record every transition instant. |
| Store fixtures leak state or depend on order | Isolated PostgreSQL transaction/scratch fixture per case and explicit before/after snapshots. |
| Large limit fixtures exhaust CI resources | Use the minimum overflowing representation and preallocation guards; assert rejection occurs before expensive derivation. |
| Matrix promotion outruns evidence | Default remains `planned`; promotion is a separate criterion-by-criterion reviewed change. |
