# Workbook Projections Remediation Tracker and Handoff

## 1. Status and authority

- **Target:** `internal/modules/projections` and its application, owner, contract,
  policy, test, and documentation seams.
- **Status:** complete; all workstreams and Definition-of-Done rows are closed.
- **Behavioral authority:** Core 00 through Core 04 and any explicitly adopted
  owner artifact for its named scope.
- **Tracker effect:** sequencing, evidence, rollback, and handoff control only.
  Runtime code, tests, generators, conformance, and release evidence MUST NOT
  read or otherwise depend on this Markdown file.
- **Compatibility posture:** repository-internal root projection APIs are
  removed directly after their final caller migrates. There is no deprecation
  window.
- **Public behavior posture:** HTTP, WebSocket, authorization, cursor,
  saved-view, telemetry, `view_row_v1`, view-schema, and error behavior is
  retained only where required by its adopted owner.
- **Migration posture:** existing migrations remain in place and are not
  rewritten to express package ownership. Generated files are changed only
  through their owner inputs and Make-owned generation.

This execution request supersedes the tracker-local Authorization A/B gates
from the 2026-08-07 planning revision. Those gates were never Core authority
and are no longer blockers. The prior planning analysis remains historical
evidence; it does not preserve an obsolete implementation surface.

## 2. Initial grounded state

The target contains 29 Go files. The live code-backed policy permits 29 exact
source-file importers of the root `projections` package and approves no adapter
package. The root exports catalog/runtime types, generic query/rebuild services,
generic coordination, and owner-specific row facades. Physical access to the
ten projection tables is distributed across Projections, source-owner
providers, specialized Entities query code, three projection-backed Reporting
providers, and authored Timeline sqlc input.

Planning discovery on 2026-08-09 established that the current narrow projection
rows and backend boundary target pass while certifying that legacy topology:

| Command | Result | Run root | Evidence posture |
| --- | --- | --- | --- |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.boundaries_and_source_ownership_9b8de61f81,module.projections.provider.catalog_validation_684d752cb4,module.projections.query.contract_shape_and_keyset_59528aa56d,module.projections.telemetry.safe_boundary_8e0e774b19` | PASS | `.cartulary/test-results/20260809T071946Z-p1534969` | Planning baseline only; policy is incomplete. |
| `make backend-module-boundary-check` | PASS | `.cartulary/test-results/20260809T071956Z-p1536731` | Planning baseline only; permits the legacy access rules. |

### Adopted owner facts already present

- Core 01 REQ-01-621 and REQ-01-622 own derived-state determinism,
  rebuildability, descriptor v3, manifest v4, source semantics, and physical
  projection storage ownership.
- Core 01 REQ-01-623 owns public query behavior and requires characterization
  before a provider split.
- Core 01 REQ-01-624 and REQ-01-625 own Recovery orchestration, the recovery-owned
  adapter, restore result vocabulary, failure closure, retry, and stale-state
  replacement.
- Core 01 REQ-01-626 and Core 04 AC-473 require package-level production import
  enforcement distinct from test-only permissions.
- Core 04 AC-470 through AC-472 own descriptor, query, and restore evidence.
- `docs/graph_projection_nlspec.md` governs graph projections only and does not
  govern workbook-grid projection tables.

### Completed owner clarification

WS-01 added the caller-owned transaction and exhaustive physical-table access
rule to Core 01, direct acceptance evidence to Core 04, and a narrow
implementation ADR through Core 00. It introduced no workbook projections
NLSpec and changed no public product contract.

## 3. Gap remediation register

| ID | Gap | Required remediation | Areas | Long-term benefit | Compatibility impact | Risk if left open | Binary validation |
| --- | --- | --- | --- | --- | --- | --- | --- |
| GAP-01 | Tracker authority/status drift | Remove tracker-local authorizations and default preservation; use this ledger and mandatory checkpoints. | documentation | Reliable execution and handoff control. | none | Stale non-authority blocks or preserves harmful internals. | WS-00 complete; Markdown lint passes. |
| GAP-02 | Transaction/storage ownership not explicit enough | Add the Core/ADR closure described in WS-01. | specification, tests, documentation | Stable atomicity and physical-storage ownership. | no public change | Future providers can own transactions or projection SQL accidentally. | Adopted text, traceability, and lint pass. |
| GAP-03 | Missing characterization | Implement the exact matrices in section 7 before structural movement. | tests, harness | Owner-backed refactor barrier. | test-only unless a Core defect is found | Query, rollback, restore, or publication drift. | Every matrix passes on the baseline. |
| GAP-04 | File-specific imports and broad root API | Add adapters/contracts/private packages; migrate callers; empty the root package. | implementation, tests, contracts | Compiler containment and narrow dependencies. | breaking internal migration; no deprecation | Boundary permissions grow by incidental filenames. | Empty root imports/API; negative imports fail. |
| GAP-05 | Catalog/construction leakage | Immutable descriptor set; private executable providers/catalog; fail-closed constructor. | implementation, contracts, tests | No partial or mutable runtime composition. | descriptor v3/manifest v4 retained | Duplicate, nil, or inconsistent ports reach runtime. | Constructor and immutability matrix passes. |
| GAP-06 | Inconsistent source-owner seams | Eight typed `workbookprojection` contributions and ports; remove generic mutations/deletion. | implementation, tests | Explicit semantic ownership and compile-time expansion. | direct internal caller migration | Wrong-provider dispatch and accidental feature expansion. | Eight owner slices pass; generic symbols absent. |
| GAP-07 | Distributed projection/Reporting SQL | Move all production projection-table access under Projections; use typed Reporting readers; remove unused Timeline queries through generation. | implementation, tests, contracts | One enforceable physical lifecycle owner without stealing semantics. | no migration or public change | Multi-owner schema edits and divergent derived facts. | Closed production SQL inventory and provider parity. |
| GAP-08 | Raw SQL contract, Links coupling, host/identity divergence | Replace `QuerySurface` with semantic `SurfaceIntent`; private compiled plans; bounded ID hydration. | implementation, tests | One safe query engine and no physical SQL in owner contracts. | host/identity query capability becomes true | Unversioned SQL coupling and paging drift. | No raw SQL contract/Links import; parity and bounds pass. |
| GAP-09 | Recovery/Revisions concrete coupling | Return their consumer-owned ports and recovery state only. | implementation, tests | Consumers cannot depend on Store/Catalog/Coordinator. | internal composition only | False restore readiness or private coupling. | Restore and Revisions matrices pass. |
| GAP-10 | Timeline-backed test composition | Projection-assembly-backed typed `testsupport` capabilities. | implementation, tests, harness | Tests mirror production ownership. | test-only | Test wiring hides production defects. | Caller map migrated; permissions remain test-only. |
| GAP-11 | Weak policy/manifest/schema/harness parity | Ten rules, production/test separation, four-way equality, registry-first contract updates. | contracts, tools, tests | Drift fails closed. | stricter policy | Green checks certify the wrong topology. | Boundary, shape, parity, harness, and drift checks pass. |
| GAP-12 | Residual compatibility and stale guidance | Delete legacy surfaces; reconcile Appendix I/dev guide; complete handoff. | implementation, documentation, tests | One durable architecture. | no compatibility window | Transitional delegation becomes permanent. | No legacy symbols; all DoD rows pass. |

## 4. Target architecture and interfaces

```text
internal/modules/projections/
├── adapters/                    # sole production constructor
├── providercontract/            # immutable descriptors and semantic intent
├── testsupport/                 # typed test-only capability wrappers
└── internal/
    ├── runtime/                 # catalog validation and coordination
    ├── storage/                 # all projection-table SQL
    └── queryengine/             # private query compilation and rows
```

`adapters.New(Dependencies) (Ports, error)` is the only production constructor.
`Dependencies` contains `postgres.DB` plus one required typed contribution from
Timeline, Entities, Indicators, Assessments, Artifacts, Evidence, Parties, and
Tasks/Decisions. `Ports` exposes only an immutable descriptor set, consumer-owned
Workbook/Recovery/Revisions interfaces, recovery state, typed owner ports, and
typed artifact/entity/task-decision derived-fact readers. Projections does not
import Reporting; source owners retain `exportprovider.FieldProvider` semantics.

Owner-facing semantic query intent contains no table name, join, SQL expression,
callback, Store, or query-engine type. Private plans own those physical details.
Writers accept a caller-owned `pgx.Tx` and never begin, commit, roll back,
authorize, publish Collaboration events, or mutate authoritative/history state.

The current ten providers and rebuild order remain:

| Order | Provider | Table | Source owner |
| ---: | --- | --- | --- |
| 1 | timeline | `timeline_grid_projection` | Timeline |
| 2 | host | `host_grid_projection` | Entities |
| 3 | identity | `identity_grid_projection` | Entities |
| 4 | indicator | `indicator_grid_projection` | Indicators |
| 5 | assessment | `assessment_grid_projection` | Assessments |
| 6 | artifact | `artifact_grid_projection` | Artifacts |
| 7 | evidence | `evidence_grid_projection` | Evidence |
| 8 | party | `party_grid_projection` | Parties |
| 9 | task_request | `task_request_grid_projection` | Tasks/Decisions |
| 10 | decision | `decision_grid_projection` | Tasks/Decisions |

## 5. Mandatory checkpoint protocol

Exactly one workstream may be `IN_PROGRESS`. For every workstream:

1. Confirm its dependencies are `DONE` and mark it `IN_PROGRESS` here.
2. Make only that workstream's independently reversible changes.
3. Run its narrow owner/static and service-backed validation.
4. On failure, retain the run root, mark the workstream `BLOCKED`, and stop.
5. On success, record files/behavior, commands/results/run roots, risks, and the
   next eligible workstream; mark it `DONE`.
6. Run `make lint-markdown` and record its run root.
7. Only then mark the next workstream `IN_PROGRESS` and begin it.

Tracker updates are documentation evidence and never substitute for product
validation. Generated outputs are updated only after their owner inputs.

## 6. Workstream ledger

| ID | Workstream | Depends on | Status | Exit condition | Evidence |
| --- | --- | --- | --- | --- | --- |
| WS-00 | Tracker reset | none | DONE | Current controller, gap register, crosswalk, history, and lint pass. | Replaced the stale planning controller; `make lint-markdown` PASS at `.cartulary/test-results/20260809T075110Z-p1544227`. |
| WS-01 | Owner and ADR closure | WS-00 | DONE | Core/ADR agree; domain/graph re-audit and lint pass. | Added Core 00 REQ-00-070, Core 01 REQ-01-658, Core 04 AC-539, and the adopted implementation ADR. Re-audited `docs/domain.md` and the Graph Projection NLSpec: no vocabulary or graph-scope edit is required. `make lint-markdown` PASS at `.cartulary/test-results/20260809T075317Z-p1546437`. |
| WS-02 | Characterization baseline | WS-01 | DONE | All section 7 matrices pass before structural movement. | Added routed restore failure/rollback, typed deletion/transaction, and exact test-capability caller matrices; corrected the no-active-provider restore panic to the Core-permitted `not_applicable` result. Full Projections unit/static PASS at `.cartulary/test-results/20260809T082022Z-p2029093`, service-backed PASS at `.cartulary/test-results/20260809T082035Z-p2031163`, backend boundary PASS at `.cartulary/test-results/20260809T082046Z-p2032525`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T082148Z-p2033195`. |
| WS-03 | Adapter/contract foundation | WS-02 | DONE | Constructor/immutability/negative-boundary tests pass; no consumer migrated. | Added `adapters.New`, canonical immutable descriptor/semantic-intent contracts, private runtime/storage/query-engine packages, and eight typed owner contribution shells. The constructor rejects nil dependencies, duplicate provider/table/view ownership, missing providers, unsupported versions, unresolved intents, ownership mismatch, and rebuild cycles; successful descriptor snapshots are deep-copy immutable. Projections unit/static PASS at `.cartulary/test-results/20260809T083226Z-p2053555`, service-backed PASS at `.cartulary/test-results/20260809T083241Z-p2056852`, boundary PASS at `.cartulary/test-results/20260809T083327Z-p2058950`, JSON shape PASS at `.cartulary/test-results/20260809T083404Z-p2062361`, generation drift PASS at `.cartulary/test-results/20260809T083413Z-p2062821`, artifact policy PASS at `.cartulary/test-results/20260809T083424Z-p2065582`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T083515Z-p2066395`. |
| WS-04 | Projection assembly migration | WS-03 | DONE | Sole adapter importer; catalog/source tests pass. | `internal/app/projectionassembly` is the only production importer of `projections/adapters`; it constructs the eight typed declarative contributions, calls the fail-closed adapter, exposes `Ports` plus an immutable `DescriptorSet`, and derives manifest expectations from that canonical snapshot. Characterized legacy execution fields remain isolated in the assembly bundle for the ordered consumer/source-owner migrations and are not re-exposed through the new adapter contract. Focused assembly/foundation/boundary rows PASS at `.cartulary/test-results/20260809T083836Z-p2072447`; full Projections unit/static PASS at `.cartulary/test-results/20260809T083901Z-p2075286`, service-backed PASS at `.cartulary/test-results/20260809T083917Z-p2078085`, boundary PASS at `.cartulary/test-results/20260809T083927Z-p2079442`, JSON shape PASS at `.cartulary/test-results/20260809T083947Z-p2082277`, generation drift PASS at `.cartulary/test-results/20260809T083958Z-p2082769`, artifact policy PASS at `.cartulary/test-results/20260809T084008Z-p2085574`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T084127Z-p2086415`. |
| WS-05 | Workbook migration | WS-04 | DONE | Workbook root imports removed; route/cursor/auth/query suites pass. | Workbook catalog validation now consumes the immutable canonical `DescriptorSet` and executable queries only through Workbook-owned `QueryProvider` contributions. No Workbook production or test file imports the Projections root. Workbook unit/static PASS at `.cartulary/test-results/20260809T084908Z-p2131371`, service-backed PASS at `.cartulary/test-results/20260809T085124Z-p2180568`, Projections unit/static PASS at `.cartulary/test-results/20260809T085317Z-p2214838`, service-backed PASS at `.cartulary/test-results/20260809T085333Z-p2218414`, backend boundary PASS at `.cartulary/test-results/20260809T085425Z-p2220526`, JSON shape PASS at `.cartulary/test-results/20260809T085441Z-p2220910`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T085525Z-p2221645`. |
| WS-06 | Recovery migration | WS-05 | DONE | Recovery uses only consumer port/state; restore matrix passes. | Added Recovery-owned `restorecontract.ProjectionPorts`, exposed only the rebuilder and recovery-state contribution from projection assembly, bound the restore probe through its Workbook-owned query interface, moved the ten-table recovery fact into the immutable provider contract, and deleted the root recovery-state API. Recovery unit/static PASS at `.cartulary/test-results/20260809T090429Z-p2318750`, service-backed PASS at `.cartulary/test-results/20260809T090524Z-p2360932`, Projections unit/static PASS at `.cartulary/test-results/20260809T090559Z-p2394089`, service-backed PASS at `.cartulary/test-results/20260809T090617Z-p2400044`, boundary PASS at `.cartulary/test-results/20260809T090628Z-p2401391`, JSON shape PASS at `.cartulary/test-results/20260809T090636Z-p2401757`, generation drift PASS at `.cartulary/test-results/20260809T090645Z-p2402188`, artifact policy PASS at `.cartulary/test-results/20260809T090655Z-p2404946`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T090736Z-p2405668`. |
| WS-07 | Revisions migration | WS-06 | DONE | Revisions uses only its consumer port; delete/restore/rollback pass. | Removed projection catalog/descriptor knowledge from revision composition. Record/view routing now validates source-owner contributions against public view facts, while rebuild/support/load enter only through `revisions.ProjectionServices`. Revisions unit/static PASS at `.cartulary/test-results/20260809T091153Z-p2414887`, service-backed PASS at `.cartulary/test-results/20260809T091250Z-p2451076`, Projections unit/static PASS at `.cartulary/test-results/20260809T091444Z-p2476299`, service-backed PASS at `.cartulary/test-results/20260809T091502Z-p2482622`, boundary PASS at `.cartulary/test-results/20260809T091516Z-p2483973`, JSON shape PASS at `.cartulary/test-results/20260809T091523Z-p2484346`, generation drift PASS at `.cartulary/test-results/20260809T091533Z-p2484777`, artifact policy PASS at `.cartulary/test-results/20260809T091543Z-p2487526`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T091626Z-p2488245`. |
| WS-08 | Timeline facade | WS-07 | DONE | Timeline mutation/effects/rebuild/import/Collaboration pass. | `timeline/workbookprojection` now owns the runtime contribution, canonical descriptor and semantic intent, source reader, mutation writer, rebuilder, and typed port aggregate. Timeline production/tests have no Projections-root import; assembly injects typed Timeline and separate temporary Entities ports. Timeline unit/static PASS at `.cartulary/test-results/20260809T092446Z-p2499015`, service-backed PASS at `.cartulary/test-results/20260809T092601Z-p2536475`, Projections unit/static PASS at `.cartulary/test-results/20260809T092702Z-p2563711`, service-backed PASS at `.cartulary/test-results/20260809T092720Z-p2570155`, boundary PASS at `.cartulary/test-results/20260809T092733Z-p2571521`, JSON shape PASS at `.cartulary/test-results/20260809T092741Z-p2571887`, generation drift PASS at `.cartulary/test-results/20260809T092753Z-p2572333`, artifact policy PASS at `.cartulary/test-results/20260809T092804Z-p2575099`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T092853Z-p2575849`. |
| WS-09 | Entities facade | WS-08 | DONE | Host/identity create/patch/import/mention/merge/delete pass. | `entities/workbookprojection` now owns the canonical host/identity contribution plus typed writer/rebuilder ports. Entities unit/static PASS at `.cartulary/test-results/20260809T094517Z-p2693192`, service-backed PASS at `.cartulary/test-results/20260809T094635Z-p2726347`, Projections unit/static PASS at `.cartulary/test-results/20260809T094744Z-p2757073`, service-backed PASS at `.cartulary/test-results/20260809T094759Z-p2761384`, boundary PASS at `.cartulary/test-results/20260809T094805Z-p2762754`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T094855Z-p2763423`. |
| WS-10 | Indicators facade | WS-09 | DONE | Indicator typed refresh/load/delete lifecycle passes. | `indicators/workbookprojection` now owns the canonical contribution and typed row/rebuild ports; executable provider construction is isolated in the Indicator-owned provider package. Indicators unit/static PASS at `.cartulary/test-results/20260809T095913Z-p2796491`, service-backed PASS at `.cartulary/test-results/20260809T095934Z-p2804943`, Projections unit/static PASS at `.cartulary/test-results/20260809T095947Z-p2806415`, service-backed PASS at `.cartulary/test-results/20260809T100004Z-p2812517`, boundary PASS at `.cartulary/test-results/20260809T095907Z-p2796098`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T100113Z-p2817832`. |
| WS-11 | Assessments facade | WS-10 | DONE | Assessment typed source/mutation paths pass. | `assessments/workbookprojection` now owns canonical typed source DTOs, descriptor/intent facts, row/rebuild ports, and their immutability/validation tests. Assessments unit/static PASS at `.cartulary/test-results/20260809T101924Z-p2944973`, service-backed PASS at `.cartulary/test-results/20260809T102025Z-p2975343`, Projections unit/static PASS at `.cartulary/test-results/20260809T102009Z-p2970559`, service-backed PASS at `.cartulary/test-results/20260809T102103Z-p2999055`, boundary PASS at `.cartulary/test-results/20260809T101822Z-p2937600`, JSON shape PASS at `.cartulary/test-results/20260809T102112Z-p3000372`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T102205Z-p3001212`. |
| WS-12 | Artifacts facade | WS-11 | DONE | All artifact surfaces/import/conflict paths pass. | `artifacts/workbookprojection` now owns the eight-surface descriptor/intent contribution and typed row/rebuild aggregate; all mutation, contextual-note, import, and conflict composition receives its rows explicitly. Artifacts unit/static PASS at `.cartulary/test-results/20260809T102948Z-p3011114`, service-backed PASS at `.cartulary/test-results/20260809T103001Z-p3014559`, Projections unit/static PASS at `.cartulary/test-results/20260809T103012Z-p3015919`, service-backed PASS at `.cartulary/test-results/20260809T103028Z-p3020817`, boundary PASS at `.cartulary/test-results/20260809T103035Z-p3022190`, JSON shape PASS at `.cartulary/test-results/20260809T103042Z-p3022582`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T103129Z-p3026459`. |
| WS-13 | Evidence facade | WS-12 | DONE | Evidence attach/import/support/conflict paths pass. | `evidence/workbookprojection` now owns the canonical semantic contribution and typed row/rebuild ports. Evidence unit/static PASS at `.cartulary/test-results/20260809T104214Z-p3072212`, service-backed PASS at `.cartulary/test-results/20260809T104321Z-p3105004`, Projections unit/static PASS at `.cartulary/test-results/20260809T104607Z-p3149617`, service-backed PASS at `.cartulary/test-results/20260809T104619Z-p3153007`, boundary PASS at `.cartulary/test-results/20260809T104551Z-p3145687`, JSON shape PASS at `.cartulary/test-results/20260809T104552Z-p3146005`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T104700Z-p3154612`. |
| WS-14 | Parties facade | WS-13 | DONE | Party create/import/conflict/refresh paths pass. | `parties/workbookprojection` now owns the canonical semantic contribution and typed row/rebuild ports. Parties unit/static PASS at `.cartulary/test-results/20260809T105331Z-p3163865`, service-backed PASS at `.cartulary/test-results/20260809T105409Z-p3187960`, Projections unit/static PASS at `.cartulary/test-results/20260809T105445Z-p3210394`, service-backed PASS at `.cartulary/test-results/20260809T105502Z-p3214893`, boundary PASS at `.cartulary/test-results/20260809T105505Z-p3216196`, JSON shape PASS at `.cartulary/test-results/20260809T105507Z-p3216514`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T105546Z-p3220380`. |
| WS-15 | Tasks/Decisions facade | WS-14 | DONE | Separate typed task/decision paths pass. | `tasksdecisions/workbookprojection` now owns both canonical semantic descriptors, separate typed task/decision row/rebuild methods, and the typed provider source bridge. Tasks/Decisions unit/static PASS at `.cartulary/test-results/20260809T110402Z-p3230452`, service-backed PASS at `.cartulary/test-results/20260809T110440Z-p3255642`, Projections unit/static PASS at `.cartulary/test-results/20260809T110556Z-p3285342`, service-backed PASS at `.cartulary/test-results/20260809T110607Z-p3288006`, boundary PASS at `.cartulary/test-results/20260809T110617Z-p3289368`, JSON shape PASS at `.cartulary/test-results/20260809T110618Z-p3289692`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T110704Z-p3293562`. |
| WS-16 | Timeline storage provider | WS-15 | DONE | Timeline SQL contained; paging/schema parity and generation pass. | Timeline writes/deletes are private storage operations, the compiled query surface is private query-engine data, the exact table rule permits only those packages, and the unused authored/generated sqlc query pair was removed through `make generate`. Timeline unit/static PASS at `.cartulary/test-results/20260809T111317Z-p3335850`, service-backed PASS at `.cartulary/test-results/20260809T111431Z-p3372274`, Projections unit/static PASS at `.cartulary/test-results/20260809T111530Z-p3399494`, service-backed PASS at `.cartulary/test-results/20260809T111548Z-p3403985`, boundary PASS at `.cartulary/test-results/20260809T111556Z-p3405355`, JSON shape PASS at `.cartulary/test-results/20260809T111608Z-p3405765`, generation drift PASS at `.cartulary/test-results/20260809T111614Z-p3406222`, artifact policy PASS at `.cartulary/test-results/20260809T111626Z-p3409026`, migration drift PASS at `.cartulary/test-results/20260809T111629Z-p3409470`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T111707Z-p3412507`. |
| WS-17 | Host storage provider | WS-16 | DONE | Host SQL/reporting contained; differential parity passes. | Entities now supplies typed Host source inputs and exact-ID authoritative hydration; private storage owns writes/deletes, private query code owns bounded keyset selection and Reporting-derived reads, and Host alone is query-capable in descriptor/manifest v4 facts. Entities unit/static PASS at `.cartulary/test-results/20260809T113449Z-p3488300`, service-backed PASS at `.cartulary/test-results/20260809T113609Z-p3520923`, Reporting unit/static PASS at `.cartulary/test-results/20260809T113716Z-p3551698`, service-backed PASS at `.cartulary/test-results/20260809T113731Z-p3556483`, Projections unit/static PASS at `.cartulary/test-results/20260809T113740Z-p3557820`, service-backed PASS at `.cartulary/test-results/20260809T113756Z-p3561884`, boundary PASS at `.cartulary/test-results/20260809T113807Z-p3563265`, JSON shape PASS at `.cartulary/test-results/20260809T113815Z-p3563663`, generation drift PASS at `.cartulary/test-results/20260809T113826Z-p3564132`, artifact policy PASS at `.cartulary/test-results/20260809T113837Z-p3566923`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T113913Z-p3567583`. |
| WS-18 | Identity storage provider | WS-17 | DONE | Identity SQL/reporting contained; differential parity passes. | Identity now shares the typed Entities source/query/derived-fact architecture: private storage owns writes/deletes, private query code owns bounded selection and Reporting reads, exact-ID hydration remains with Entities, and its descriptor/manifest capability is query-enabled. Entities unit/static PASS at `.cartulary/test-results/20260809T114553Z-p3575644`, service-backed PASS at `.cartulary/test-results/20260809T114716Z-p3609296`, Reporting unit/static PASS at `.cartulary/test-results/20260809T114829Z-p3640107`, service-backed PASS at `.cartulary/test-results/20260809T114853Z-p3644388`, Projections unit/static PASS at `.cartulary/test-results/20260809T114911Z-p3645746`, service-backed PASS at `.cartulary/test-results/20260809T114926Z-p3649542`, boundary PASS at `.cartulary/test-results/20260809T114936Z-p3650923`, JSON shape PASS at `.cartulary/test-results/20260809T114942Z-p3651318`, generation drift PASS at `.cartulary/test-results/20260809T114955Z-p3651807`, artifact policy PASS at `.cartulary/test-results/20260809T115007Z-p3654602`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T115039Z-p3655271`. |
| WS-19 | Indicator storage provider | WS-18 | DONE | Indicator SQL contained; tag/lifecycle/delete parity passes. | Indicators now supplies typed paged materialization inputs; private storage owns upsert/delete/rebuild SQL and private query-engine code owns the compiled surface. The generic deletion dispatch and owner-side projection SQL were deleted. Indicators unit/static PASS at `.cartulary/test-results/20260809T115755Z-p3671028`, service-backed PASS at `.cartulary/test-results/20260809T115817Z-p3678501`, Projections unit/static PASS at `.cartulary/test-results/20260809T115932Z-p3689318`, service-backed PASS at `.cartulary/test-results/20260809T115946Z-p3692961`, boundary PASS at `.cartulary/test-results/20260809T115954Z-p3694334`, JSON shape PASS at `.cartulary/test-results/20260809T120001Z-p3694728`, generation drift PASS at `.cartulary/test-results/20260809T120007Z-p3695188`, artifact policy PASS at `.cartulary/test-results/20260809T120020Z-p3697982`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T120105Z-p3698764`. |
| WS-20 | Assessment storage provider | WS-19 | DONE | Assessment SQL contained; mutation/load/rebuild parity passes. | Private storage now owns Assessment upsert/delete/rebuild SQL, private query-engine code owns the compiled surface, and the owner contribution exposes only typed source DTOs and semantic intent. Assessments unit/static PASS at `.cartulary/test-results/20260809T120502Z-p3705622`, service-backed PASS at `.cartulary/test-results/20260809T120541Z-p3732535`, Projections unit/static PASS at `.cartulary/test-results/20260809T120727Z-p3764068`, service-backed PASS at `.cartulary/test-results/20260809T120744Z-p3766720`, boundary PASS at `.cartulary/test-results/20260809T120756Z-p3768122`, JSON shape PASS at `.cartulary/test-results/20260809T120827Z-p3769333`, generation drift PASS at `.cartulary/test-results/20260809T120836Z-p3769796`, artifact policy PASS at `.cartulary/test-results/20260809T120847Z-p3772604`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T120915Z-p3773238`. |
| WS-21 | Artifact storage provider | WS-20 | DONE | Artifact SQL/reporting contained; eight-surface parity passes. | Artifacts now supplies typed paged materialization inputs; private storage owns table writes, private query code owns all eight compiled surfaces and Reporting-derived reads, and server composition injects the source-owner Reporting provider. Artifacts unit/static PASS at `.cartulary/test-results/20260809T122249Z-p3805216`, service-backed PASS at `.cartulary/test-results/20260809T122304Z-p3808117`, Reporting unit/static PASS at `.cartulary/test-results/20260809T122002Z-p3789169`, service-backed PASS at `.cartulary/test-results/20260809T122017Z-p3793789`, Projections unit/static PASS at `.cartulary/test-results/20260809T122315Z-p3809475`, service-backed PASS at `.cartulary/test-results/20260809T122331Z-p3813370`, boundary PASS at `.cartulary/test-results/20260809T122227Z-p3804747`, JSON shape PASS at `.cartulary/test-results/20260809T122409Z-p3817870`, generation drift PASS at `.cartulary/test-results/20260809T122417Z-p3818366`, artifact policy PASS at `.cartulary/test-results/20260809T122427Z-p3821154`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T122511Z-p3821978`. |
| WS-22 | Evidence storage provider | WS-21 | DONE | Evidence SQL contained; attachment/query/rebuild parity passes. | Evidence now supplies typed paged materialization inputs; private storage owns table writes and private query code owns the compiled attachment-state surface. Evidence unit/static PASS at `.cartulary/test-results/20260809T123038Z-p3829915`, service-backed PASS at `.cartulary/test-results/20260809T123142Z-p3864531`, Projections unit/static PASS at `.cartulary/test-results/20260809T123236Z-p3892140`, service-backed PASS at `.cartulary/test-results/20260809T123252Z-p3896548`, boundary PASS at `.cartulary/test-results/20260809T123259Z-p3897905`, JSON shape PASS at `.cartulary/test-results/20260809T123300Z-p3898248`, generation drift PASS at `.cartulary/test-results/20260809T123303Z-p3898662`, artifact policy PASS at `.cartulary/test-results/20260809T123311Z-p3901426`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T123350Z-p3902175`. |
| WS-23 | Party storage provider | WS-22 | DONE | Party SQL contained; query/import/rebuild parity passes. | Parties now supplies typed paged materialization inputs; private storage owns table writes and private query code owns the compiled surface. Parties unit/static PASS at `.cartulary/test-results/20260809T123750Z-p3909082`, service-backed PASS at `.cartulary/test-results/20260809T123828Z-p3933289`, Projections unit/static PASS at `.cartulary/test-results/20260809T123905Z-p3955784`, service-backed PASS at `.cartulary/test-results/20260809T123923Z-p3959640`, boundary PASS at `.cartulary/test-results/20260809T123930Z-p3960988`, JSON shape PASS at `.cartulary/test-results/20260809T123932Z-p3961334`, generation drift PASS at `.cartulary/test-results/20260809T123934Z-p3961751`, artifact policy PASS at `.cartulary/test-results/20260809T123942Z-p3964507`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T124003Z-p3965125`. |
| WS-24 | Task-request storage provider | WS-23 | DONE | Task SQL/reporting contained; queue parity passes. | Tasks/Decisions now supplies typed Task-request materialization and private Reporting reads; private storage owns table writes and private query code owns the compiled surface. Tasks/Decisions unit/static PASS at `.cartulary/test-results/20260809T124726Z-p3973828`, service-backed PASS at `.cartulary/test-results/20260809T124802Z-p3999327`, Reporting unit/static PASS at `.cartulary/test-results/20260809T124839Z-p4021834`, service-backed PASS at `.cartulary/test-results/20260809T124855Z-p4026240`, Projections unit/static PASS at `.cartulary/test-results/20260809T124940Z-p4031723`, service-backed PASS at `.cartulary/test-results/20260809T124956Z-p4035405`, boundary PASS at `.cartulary/test-results/20260809T125034Z-p4037473`, JSON shape PASS at `.cartulary/test-results/20260809T125035Z-p4037819`, generation drift PASS at `.cartulary/test-results/20260809T125038Z-p4038230`, artifact policy PASS at `.cartulary/test-results/20260809T125046Z-p4041016`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T125115Z-p4041657`. |
| WS-25 | Decision storage provider | WS-24 | DONE | Decision SQL/reporting contained; supersession parity passes. | Tasks/Decisions now supplies typed paged Decision materialization and private Reporting reads; private storage owns table writes and private query code owns the compiled supersession surface. Tasks/Decisions unit/static PASS at `.cartulary/test-results/20260809T125613Z-p4048893`, service-backed PASS at `.cartulary/test-results/20260809T125649Z-p4074292`, Reporting unit/static PASS at `.cartulary/test-results/20260809T125721Z-p4096671`, service-backed PASS at `.cartulary/test-results/20260809T125742Z-p4100667`, Projections unit/static PASS at `.cartulary/test-results/20260809T125750Z-p4101981`, service-backed PASS at `.cartulary/test-results/20260809T125806Z-p4105826`, boundary PASS at `.cartulary/test-results/20260809T125835Z-p4107778`, JSON shape PASS at `.cartulary/test-results/20260809T125837Z-p4108124`, generation drift PASS at `.cartulary/test-results/20260809T125840Z-p4108541`, artifact policy PASS at `.cartulary/test-results/20260809T125847Z-p4111301`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T125915Z-p4111959`. |
| WS-26 | Query seam closure | WS-25 | DONE | Raw `QuerySurface` and Links coupling removed; bounds/parity pass. | Deleted the raw provider query contract, moved all plan/engine types into private query code, replaced Links helpers with private semantic SQL compilation, and added exact intent/plan/schema equality. Projections unit/static PASS at `.cartulary/test-results/20260809T130612Z-p4138154`, service-backed PASS at `.cartulary/test-results/20260809T130629Z-p4141826`, Workbook unit/static PASS at `.cartulary/test-results/20260809T130637Z-p4143138`, service-backed PASS at `.cartulary/test-results/20260809T130847Z-p4186052`, boundary PASS at `.cartulary/test-results/20260809T131043Z-p26385`, JSON shape PASS at `.cartulary/test-results/20260809T131044Z-p26731`, generation drift PASS at `.cartulary/test-results/20260809T131047Z-p27142`, artifact policy PASS at `.cartulary/test-results/20260809T131055Z-p29911`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T131124Z-p30569`. |
| WS-27 | Test capability migration | WS-26 | DONE | Timeline bundle exposure removed; caller map passes. | `projections/testsupport` retains only named owner-facade ports; `httptestx` converts the Timeline composition callback immediately and stores no Timeline-owned aggregate. Projections unit/static PASS at `.cartulary/test-results/20260809T131638Z-p41487`, service-backed PASS at `.cartulary/test-results/20260809T131655Z-p43852`, boundary PASS at `.cartulary/test-results/20260809T131708Z-p45556`, JSON shape PASS at `.cartulary/test-results/20260809T131731Z-p49531`, generation drift PASS at `.cartulary/test-results/20260809T131708Z-p45299`, artifact policy PASS at `.cartulary/test-results/20260809T131708Z-p45329`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T131818Z-p50375`. |
| WS-28 | Contract/policy/harness reconciliation | WS-27 | DONE | Ten rules/four-way equality/shape/harness/drift pass. | All ten rules now use exact private production paths and separate test-fixture permissions; active descriptors, table rules, recovery state, and migration-owned schema tables are exactly equal. Projections unit/static PASS at `.cartulary/test-results/20260809T132712Z-p64172`, service-backed PASS at `.cartulary/test-results/20260809T132729Z-p66550`, boundary PASS at `.cartulary/test-results/20260809T132703Z-p63426`, JSON shape PASS at `.cartulary/test-results/20260809T132703Z-p63310`, generation drift PASS at `.cartulary/test-results/20260809T132742Z-p67934`, artifact policy PASS at `.cartulary/test-results/20260809T132742Z-p67951`, migration drift PASS at `.cartulary/test-results/20260809T132742Z-p67981`, harness contract PASS at `.cartulary/test-results/20260809T132753Z-p73969`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T132830Z-p74718`. |
| WS-29 | Legacy/root removal | WS-28 | DONE | Root API/imports empty; legacy scans and `test-fast` pass. | Moved executable behavior private, made `adapters.New` the sole complete constructor, removed every root Go file/import and obsolete compatibility field, and reconciled explanatory guidance. Projections unit/static PASS at `.cartulary/test-results/20260809T134849Z-p100852`, service-backed PASS at `.cartulary/test-results/20260809T134906Z-p103237`, boundary PASS at `.cartulary/test-results/20260809T135317Z-p111690`, JSON shape PASS at `.cartulary/test-results/20260809T135317Z-p111571`, generation drift PASS at `.cartulary/test-results/20260809T135227Z-p106948`, artifact policy PASS at `.cartulary/test-results/20260809T135227Z-p106859`, Network Flow unit/static PASS at `.cartulary/test-results/20260809T135620Z-p182112`, Indicators unit/static PASS at `.cartulary/test-results/20260809T140207Z-p274525`, `test-fast` PASS at `.cartulary/test-results/20260809T140233Z-p284818`, and checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T140614Z-p341139`. |
| WS-30 | Validation and handoff completion | WS-29 | DONE | Section 8 sequence and every DoD row pass. | All changed-owner and Projections slices pass; boundary/shape/drift/artifact/migration/harness/lint/finalize/fast/browser/build/check/release gates pass. `make check` PASS at `.cartulary/test-results/20260809T143806Z-p1309728`; `make release-check` PASS at `.cartulary/test-results/20260809T145712Z-p1709336`; final tracker checkpoint Markdown lint PASS at `.cartulary/test-results/20260809T150727Z-p1878996`. |

### Legacy slice crosswalk

| Superseded planning slice | Execution workstreams |
| --- | --- |
| A-00 | Removed; WS-01 owns adopted authority closure. |
| S-00 | WS-02 |
| S-01 | WS-03 |
| S-02 | WS-04 through WS-07 |
| S-03 | WS-08 through WS-15 |
| S-04 | WS-16 through WS-26 |
| S-05 | WS-27 |
| S-06 | WS-28 |
| S-07 | WS-29 |
| S-08 | WS-30 |

## 7. Characterization matrices

WS-02 MUST preserve owner-required behavior for these exact cases rather than
freezing an implementation merely because it exists.

### Deletion

- Correct incident/type removes exactly the intended derived row after commit.
- Different-incident and already-absent cases follow the current owner result
  without deleting another row.
- Rollback leaves the row behaviorally identical and publishes no success.
- Authoritative source/history is unchanged by projection deletion.
- Publication remains source-coordinator-owned; telemetry remains bounded.
- Only Timeline mutation deletion and typed host/identity/indicator deletion
  exist; there is no unsupported generic fallback.

### Restore

- No providers, nonparticipation, unsupported version/provider, first/middle/
  final failure, commit failure, cancellation, retry, stale rows, invalid source
  reference, and deterministic ten-provider order are covered.
- Every failure rolls back all projection changes and claims no rebuilt resource
  before a successful commit.

### Host/identity differential parity

- Rows/cells, collections, grouping, paging, null order, tie-breaking, filters,
  mutation, merge/delete, load-row values, continuation authorization, and
  bounds are equal across legacy, transition, and private query paths.
- SQL fetches at most `limit+1`; Entities hydrates only the returned ordered IDs
  and does not reorder, filter, or independently requery the page.

### Other required matrices

- All current query surfaces return complete `view_row_v1` fields, group values,
  normalized sorts/filters, stable cursors, and route-owned errors.
- Artifact, host/identity, task, and decision Reporting facts retain paths,
  values, content classes, source families, support references, and order.
- Import, merge, delete/restore, rollback, row snapshots, affected views,
  continuation reauthorization, and telemetry privacy remain owner-correct.
- Every current `httptestx` projection capability caller maps to an explicitly
  named typed replacement.

### WS-02 baseline evidence

- Restore characterization now directly covers no active providers, explicit
  nonparticipation, unsupported providers and descriptor versions, first/middle/
  final provider failure, commit failure, cancellation, retry, stale-row
  replacement, invalid source-state reference, deterministic ordering, rollback,
  and post-commit-only success claims. The focused routed row passed at
  `.cartulary/test-results/20260809T082002Z-p2025905`.
- Typed host and identity deletion now proves caller-owned commit/rollback,
  already-absent behavior, and preservation of authoritative source and history.
  Revisions' existing delete/restore and rollback suites provide the
  incident/type, publication, Collaboration, and affected-view evidence.
- Query, row-shape, null/collection, grouping, cursor, authorization,
  host/identity keyset, telemetry, Reporting, import, merge, rollback, and
  Collaboration behavior is retained by the owner baselines recorded below.
  These tests assert adopted behavior; they do not adopt arbitrary package
  structure as compatibility policy.
- `TestProjectionCapabilityCallerMatrix` records all seven current consumer
  files plus the application-test composition call so WS-27 cannot omit a
  caller. It does not grant Timeline production ownership.

| Owner | Unit/static baseline | Service-backed baseline |
| --- | --- | --- |
| Projections | `.cartulary/test-results/20260809T082022Z-p2029093` | `.cartulary/test-results/20260809T082035Z-p2031163` |
| Workbook | `.cartulary/test-results/20260809T075423Z-p1551512` | `.cartulary/test-results/20260809T075612Z-p1586248` |
| Recovery | `.cartulary/test-results/20260809T075758Z-p1619746` | `.cartulary/test-results/20260809T075830Z-p1653875` |
| Entities | `.cartulary/test-results/20260809T075906Z-p1686677` | `.cartulary/test-results/20260809T080009Z-p1717496` |
| Reporting | `.cartulary/test-results/20260809T080110Z-p1747714` | `.cartulary/test-results/20260809T080114Z-p1749617` |
| Timeline | `.cartulary/test-results/20260809T080117Z-p1750870` | `.cartulary/test-results/20260809T080214Z-p1779851` |
| Indicators | `.cartulary/test-results/20260809T080307Z-p1806736` | `.cartulary/test-results/20260809T080319Z-p1809930` |
| Assessments | `.cartulary/test-results/20260809T080329Z-p1811306` | `.cartulary/test-results/20260809T080401Z-p1834642` |
| Artifacts | `.cartulary/test-results/20260809T080432Z-p1857899` | `.cartulary/test-results/20260809T080435Z-p1859491` |
| Evidence | `.cartulary/test-results/20260809T080438Z-p1860742` | `.cartulary/test-results/20260809T080530Z-p1888829` |
| Parties | `.cartulary/test-results/20260809T080619Z-p1915967` | `.cartulary/test-results/20260809T080647Z-p1938261` |
| Tasks/Decisions | `.cartulary/test-results/20260809T080716Z-p1960347` | `.cartulary/test-results/20260809T080744Z-p1982988` |

The first focused run, `.cartulary/test-results/20260809T081653Z-p2012536`,
failed because a same-package test imported application composition and the
caller inventory scanned its own string literals. The second focused run,
`.cartulary/test-results/20260809T081901Z-p2018853`, failed because the raw test
fixture omitted the authoritative host row required by its foreign key. Both
were test-fixture defects, were corrected without weakening assertions, and
are retained as non-success evidence. Files changed in WS-02 were
`internal/modules/projections/rebuild.go`, the new deletion and restore matrix
tests, `internal/testutil/httptestx/httptestx_test.go`, and
`tools/test_families/module.projections.json`. No generated file or migration
was edited.

### WS-03 foundation evidence

- New authored packages are `internal/modules/projections/adapters`,
  `providercontract`, private `internal/runtime`, `internal/storage`, and
  `internal/queryengine`, plus one `workbookprojection` contribution package per
  source owner. `SurfaceIntent` has no SQL, table, expression, Store, callback,
  or query-engine field. `ProviderDescriptor` retains the Core-required physical
  table identifiers as immutable declarative facts but no executable value.
- The old root implementation remains temporarily unchanged for characterized
  consumers. No production file imports `projections/adapters` yet, and the 29
  approved production root importer files are unchanged. This is the explicit
  WS-03 rollback boundary, not a compatibility commitment.
- The boundary owner now permits only
  `internal/modules/assessments/workbookprojection/**` through Assessments'
  existing peer-implementation prohibition. No general Projections exception
  was added. The new test-family rows were added to the authored manifest and
  `tools/execution_topology_render_index.json` was refreshed by `make generate`;
  no generated output was hand-edited.
- Focused foundation validation first failed at
  `.cartulary/test-results/20260809T083025Z-p2040471` and
  `.cartulary/test-results/20260809T083111Z-p2046833` because the contract test
  incorrectly treated Core-required `ProjectionTableIDs` and the substring in
  `RestoreRebuild` as forbidden SQL storage. The test now asserts the exact
  allowed field sets and passed at
  `.cartulary/test-results/20260809T083206Z-p2051935` without weakening the
  semantic-intent prohibition.
- The first boundary run,
  `.cartulary/test-results/20260809T083250Z-p2058209`, identified the missing
  narrow Assessments contract allowance. The initial JSON-shape run,
  `.cartulary/test-results/20260809T083334Z-p2059332`, then correctly required
  regeneration after the authored test-family change. Both non-success roots
  are retained; the owner-policy and generated-input remediations are green.

### WS-04 projection assembly evidence

- `internal/app/projectionassembly/catalog.go` converts the characterized legacy
  descriptor facts to canonical names, derives semantic field/filter intent
  without carrying SQL, constructs all eight typed contributions, and invokes
  `adapters.New`. This conversion is temporary assembly glue; source owners take
  over their contribution construction in WS-08 through WS-15.
- `ProductionImportPolicy` and the static guard authorize the adapter package
  but reject its use outside the `internal/app/projectionassembly` package.
  A production scan finds exactly one importer. The validation-only manifest
  now records that adapter package, retains descriptor v3/manifest v4, and has no
  wire, cursor, view-schema, database-schema, or migration change.
- The projection manifest parity test now consumes the adapter's immutable
  `DescriptorSet` and separately proves equality with the still-running legacy
  catalog. Mutating either `All` or `Lookup` results cannot affect assembly
  state.
- Authored changes were made to projection assembly, its catalog/manifest tests,
  the adapter import policy, the validation-only provider manifest, and the
  Projections test-family manifest. `tools/execution_topology_render_index.json`
  was refreshed only by `make generate` at
  `.cartulary/test-results/20260809T083935Z-p2079802`; no generated file or
  migration was hand-edited.

### WS-05 Workbook evidence

- Workbook catalog validation now accepts only the deep-copying canonical
  `providercontract.DescriptorSet`; it no longer imports or interprets the
  executable root catalog. Canonical provider IDs, owner modules, source
  record types, active status, and query capabilities remain fail-closed
  against every public workbook surface.
- Workbook continues to execute projection-backed surfaces through its own
  exact-surface `QueryProvider` port. The temporary adapter from the legacy
  query service is confined to `internal/app/workbookassembly` and will be
  deleted as private query plans move provider by provider in WS-16 through
  WS-26. Host and identity remain source-owner query contributions until their
  scheduled migrations.
- The Workbook package, including its tests, has no import of the Projections
  root. Its projection rebuild characterization now exercises the assembled
  rebuild port in an explicit caller-owned transaction. The boundary manifest
  removed Workbook's root-package permission and permits only
  `projections/providercontract`.
- The first full Workbook validation root,
  `.cartulary/test-results/20260809T084807Z-p2094998`, is retained as a
  non-success result: the updated characterization initially called `Begin`
  instead of the narrow `postgres.DB.BeginTx` port. The corrected test uses
  `BeginTx` and the full Workbook unit/static and service-backed slices pass.
  The first boundary root,
  `.cartulary/test-results/20260809T085345Z-p2219767`, is also retained: it
  correctly detected that the authored Workbook allowlist had not yet replaced
  the root permission with the canonical contract permission.

### WS-06 Recovery evidence

- Recovery's consumer-owned `restorecontract` now defines the complete
  `ProjectionPorts` it accepts: `ProjectionRebuilder` plus the projection
  recovery-state contribution. It has no catalog, coordinator, store, provider,
  SQL, or concrete Projections runtime type.
- Projection assembly returns those ports while keeping the characterized
  legacy rebuild implementation private to composition. Timeline assembly no
  longer reaches through the projection bundle to `Catalog.Query` or
  `Rebuild.RestoreRebuilder`; the restore probe is bound through its existing
  Workbook-owned `ProjectionQuery` interface.
- The canonical provider contract owns the copied ten-table rebuildable-state
  contribution. Adapter construction fails if the active descriptor table set
  differs from that recovery set, and its tests prove returned contribution
  mutation cannot affect subsequent results. The old root
  `RecoveryStateContribution` implementation was deleted.
- The structured restore registry reference now names
  `providercontract/descriptor.go#projection_provider_descriptor.v3` instead of
  the transitional root registry file. Descriptor v3, manifest v4, restore
  request/result vocabulary, status/readiness semantics, database schema, and
  migrations are unchanged.
- Registry-first policy cleanup removed the completed Workbook and Recovery
  root permissions from `ProductionImportPolicy`, then synchronized manifest
  v4 and refreshed tool-owned outputs through `make generate` at
  `.cartulary/test-results/20260809T090405Z-p2316260`. No generated file or
  migration was hand-edited.

### WS-07 Revisions evidence

- Revisions' immutable record/view catalog is now compiled from source-owner
  revision contributions and owner-neutral public view facts. It no longer
  imports projection assembly, creates a projection catalog, or treats runtime
  projection descriptors as the authority for record/view semantics.
- Public-surface validation fails closed for unknown or duplicate view IDs,
  empty source-record sets, and empty source-record types before revision
  routes are accepted. Existing missing, unexpected, ambiguous, and duplicate
  route checks remain in force.
- Projection assembly exposes the consumer-owned
  `revisions.ProjectionServices` interface; application composition injects
  that port into the command service instead of naming a concrete coordinator.
  Revisions continues to rebuild and load snapshots inside the caller-owned
  transaction, and the delete, restore, rollback, history snapshot, and
  failure-atomicity suites are green.
- The exact revision-assembly import allowlist removed projection assembly.
  The authored policy was regenerated through `make generate` at
  `.cartulary/test-results/20260809T091133Z-p2412210`; no generated file or
  migration was hand-edited.

### WS-08 Timeline evidence

- `timeline/workbookprojection` now owns Timeline's canonical descriptor v3
  facts, SQL-free `SurfaceIntent`, typed authoritative source reader, mutation
  writer, rebuilder, and ready-checked port aggregate. The production runtime
  contribution is constructed by Timeline assembly and passed into projection
  assembly; projection assembly no longer fabricates Timeline's canonical
  contribution from generic legacy facts.
- Timeline mutation and mention-effect code consume
  `workbookprojection.Writer` directly. Host/identity refresh was split into a
  distinct temporary `EntityProjectionPort`, preventing the Timeline writer
  from accumulating peer-owner methods before WS-09.
- The redundant root `timeline.ProjectionSource` wrapper and duplicated
  `ProjectionPort` interfaces were deleted. Timeline tests use structural query
  ports and typed fake writers, and `httptestx` reaches Timeline rebuild through
  the named Timeline port rather than through `ProjectionCatalog.Rebuild`.
- The descriptor facade package changed from `internal/modules/timeline` to
  `internal/modules/timeline/workbookprojection` without changing descriptor
  v3 or manifest v4 shape. Timeline's module allowlist no longer grants a root
  Projections import. Manifest/tool inputs were regenerated through
  `make generate` at `.cartulary/test-results/20260809T092418Z-p2496310`; no
  generated file or migration was hand-edited.
- Projection assembly alone retains the concrete root `TimelineRows` bridge.
  That independently green bridge is the rollback boundary until WS-16 moves
  Timeline SQL into private Projections storage and deletes it.

### WS-09 Entities evidence

- `entities/workbookprojection` now owns the canonical host and identity
  descriptor v3 facts plus the typed caller-owned-transaction writer and
  incident rebuilder interfaces. It deliberately retains `capabilities.query`
  as false until the WS-17 and WS-18 storage/query migrations establish the
  Projections-owned query engine.
- Host/identity create, patch, import, clipboard paste, mention lifecycle, and
  merge paths call explicit host or identity methods. The prior string-based
  refresh/delete/rebuild dispatchers and their unknown-type no-op behavior were
  deleted rather than preserved.
- Projection assembly retains the characterized concrete `EntityRows` bridge
  privately and exposes a ready-checked Entities port aggregate. Timeline,
  Workbook, Imports, HTTP composition, and test capabilities receive only that
  typed aggregate; source-owner packages no longer construct or import the
  Projections root.
- Entities constructors now fail closed when the required writer is absent.
  The host/identity manifest facade facts moved to
  `internal/modules/entities/workbookprojection`, and the three exact
  source-owner root-import permissions were removed.
- The first Entities validation attempt failed during compilation because one
  local field still named the deleted generic interface. That retained failure
  root is `.cartulary/test-results/20260809T093932Z-p2587044`; the stale type was
  replaced with `workbookprojection.Writer` before the complete green reruns.
- JSON shape PASS at
  `.cartulary/test-results/20260809T094327Z-p2685182`, generation drift PASS at
  `.cartulary/test-results/20260809T094332Z-p2685629`, and generated-artifact
  policy PASS at `.cartulary/test-results/20260809T094344Z-p2688419`. No
  generated file or migration was hand-edited.

### WS-10 Indicators evidence

- `indicators/workbookprojection` now owns Indicator descriptor v3 and semantic
  intent facts plus typed refresh, load, delete, transactional rebuild, and
  incident-rebuild ports. Indicator create, observation, lifecycle, import,
  rollback-test, HTTP, Workbook, and test composition no longer receive the
  generic view-ID projection coordinator as the Indicator row port.
- The former root `ProjectionContribution`, `ProjectionPort`, constructor, and
  source adapter were deleted. The Indicator-owned `projectionprovider`
  package constructs the executable source and characterized query bridge,
  while the public workbook facade remains free of private-provider imports.
- Projection assembly privately retains a concrete `IndicatorRows` bridge and
  exposes only the ready-checked Indicator aggregate. Source-text resolution
  remains its separately typed cross-owner port and does not enlarge the
  Indicator row interface.
- The descriptor facade moved from `internal/modules/indicators` to
  `internal/modules/indicators/workbookprojection`; the root import permission
  was removed, and the private-provider allowlist was updated registry-first to
  the new Indicator-owned construction package.
- Boundary run `.cartulary/test-results/20260809T095645Z-p2790578` rejected the
  first draft because the public facade imported the private provider. After
  separating construction, run
  `.cartulary/test-results/20260809T095835Z-p2795415` rejected the obsolete
  exact provider-constructor path. Both failures are retained; the corrected
  final boundary run passes.
- JSON shape PASS at
  `.cartulary/test-results/20260809T100017Z-p2813860`, generation drift PASS at
  `.cartulary/test-results/20260809T100023Z-p2814295`, and generated-artifact
  policy PASS at `.cartulary/test-results/20260809T100033Z-p2817068`. No
  generated file or migration was hand-edited.

### WS-11 Assessments evidence

- `assessments/workbookprojection` now owns the canonical descriptor v3 and
  semantic intent, validated typed source DTOs, source-reader contract,
  caller-owned-transaction row methods, incident rebuilder, and ready-checked
  port aggregate. Its contract tests prove descriptor/intent defensive copies
  and reject non-canonical mutation values.
- Assessment create, import, merge, Workbook, HTTP, and test composition receive
  only the typed Assessment row port. The assessment application adapters no
  longer import or construct the Projections root, and projection assembly
  retains the concrete `AssessmentRows` bridge privately as the WS-20 rollback
  boundary.
- Executable source construction remains in the Assessment application
  composition package, where Records and Links are adapted to typed source DTOs;
  the source-owner module does not import peer implementations. The boundary
  guard permits this owner/provider relationship at package scope rather than
  by incidental filename.
- The assessment descriptor facade fact moved from
  `internal/modules/assessments` to
  `internal/modules/assessments/workbookprojection`. Four stale root-import
  permissions were removed from the code-backed registry and manifest without
  changing descriptor v3 or manifest v4 shape.
- The first Projections validation retained at
  `.cartulary/test-results/20260809T101337Z-p2881905` exposed an obsolete test
  reach-through and registry/manifest ordering drift. A second retained run at
  `.cartulary/test-results/20260809T101444Z-p2890701` rejected provider
  construction from an unapproved composition package. Boundary run
  `.cartulary/test-results/20260809T101721Z-p2933311` then rejected a draft that
  moved Records/Links composition into the source-owner provider package. The
  final package-scoped composition seam and all validations pass.
- Generation drift PASS at
  `.cartulary/test-results/20260809T101829Z-p2937994` and generated-artifact
  policy PASS at `.cartulary/test-results/20260809T101839Z-p2940789`. No
  generated file or migration was hand-edited. The temporary executable query
  bridge remains isolated for WS-20 and WS-26; no public behavior or schema
  changed.

### WS-12 Artifacts evidence

- `artifacts/workbookprojection` now owns Artifact descriptor v3 facts, eight
  semantic `SurfaceIntent` values derived from adopted view-schema fields and
  canonical artifact discriminators, typed refresh/load/transactional rebuild
  methods, incident rebuild, and a ready-checked port aggregate. Contract tests
  cover the complete eight-surface set and defensive copies of facts and the
  temporary execution bridge.
- Artifact create, patch, conflict, contextual-note, and import paths no longer
  construct or import the Projections root. Workbook, Imports, server, focused
  test composition, and Timeline composition inject the same typed rows;
  constructors fail closed when those rows are absent.
- Projection assembly accepts the source-owner runtime contribution, privately
  retains its concrete `ArtifactRows` bridge, and exposes only the typed port
  aggregate. The legacy `import_projection.go` constructor was deleted, and
  both Artifact root-import permissions were removed registry-first.
- The manifest facade changed from the broad Artifact and Workbook roots to
  `internal/modules/artifacts/workbookprojection` without changing descriptor
  v3 or manifest v4 shape. All eight externally observable surfaces, collection
  fields, conflict behavior, and caller-owned import transactions remain
  covered by the green Artifact tests.
- Generation drift PASS at
  `.cartulary/test-results/20260809T103051Z-p3023045` and generated-artifact
  policy PASS at `.cartulary/test-results/20260809T103101Z-p3025841`. No
  generated file or migration was hand-edited. Artifact projection SQL and
  Reporting reads remain explicit WS-21 work; raw query compilation remains
  explicit WS-26 work.

### WS-13 Evidence evidence

- `evidence/workbookprojection` now owns Evidence descriptor v3 facts, the
  semantic intent derived from adopted view-schema fields, typed caller-owned
  transaction methods for row refresh/load, support refresh, and incident
  rebuild, plus the incident rebuilder and ready-checked port aggregate.
  Contract tests prove descriptor/intent and temporary execution-bridge
  defensive copies.
- Evidence create, attach, lifecycle, conflict, import, Workbook, HTTP, and
  focused test composition receive the typed row port explicitly. Constructors
  fail closed when rows are absent; the default store seam returns explicit
  missing-port errors instead of silently accepting a projection-less mutation.
- Projection assembly privately combines Evidence row storage with the
  cross-view support refresh required by Evidence attachments and exposes only
  the typed Evidence ports. Timeline assembly obtains the immutable contribution
  through projection assembly and no longer imports Evidence's executable
  provider package.
- The manifest facade changed from `internal/modules/evidence` to
  `internal/modules/evidence/workbookprojection`; three exact Evidence root
  import permissions were removed. Descriptor v3 and manifest v4 shapes,
  public behavior, migrations, and database schemas did not change.
- The retained owner run
  `.cartulary/test-results/20260809T104040Z-p3036715` exposed a stale test-double
  rebuild signature. The retained Projections run
  `.cartulary/test-results/20260809T104416Z-p3132623` exposed the missing
  manifest facade update, and boundary run
  `.cartulary/test-results/20260809T104509Z-p3141500` rejected Timeline's draft
  executable-provider import. Each was corrected before the final green runs.
- Generation drift PASS at
  `.cartulary/test-results/20260809T104555Z-p3146400` and generated-artifact
  policy PASS at `.cartulary/test-results/20260809T104602Z-p3149142`. No
  generated file or migration was hand-edited. Evidence projection SQL remains
  explicit WS-22 work; raw query compilation remains explicit WS-26 work.

### WS-14 Parties evidence

- `parties/workbookprojection` now owns Party descriptor v3 facts, the semantic
  intent derived from adopted view-schema fields, typed caller-owned
  transaction methods for refresh/load/rebuild, the incident rebuilder, and a
  ready-checked port aggregate. Contract tests prove fact and temporary
  execution-bridge immutability.
- Party create, reuse, patch, conflict resolution, and import composition now
  require the typed row port. Source-owner code no longer imports the
  Projections root or its executable provider package, and constructors fail
  closed when the projection port is absent.
- Projection assembly accepts the owner contribution, privately retains the
  concrete `PartyRows`, and exposes only Party-owned ports. Timeline composition
  carries those ports to Workbook and Imports, keeping one production wiring
  path for both ordinary and test composition.
- The manifest facade changed from `internal/modules/parties` to
  `internal/modules/parties/workbookprojection`; two exact Party root-import
  permissions were removed. Descriptor v3 and manifest v4 shapes, public
  behavior, migrations, and database schemas did not change.
- Generation drift PASS at
  `.cartulary/test-results/20260809T105510Z-p3216916` and generated-artifact
  policy PASS at `.cartulary/test-results/20260809T105517Z-p3219656`. No
  generated file or migration was hand-edited. Party projection SQL remains
  explicit WS-23 work; raw query compilation remains explicit WS-26 work.

### WS-15 Tasks/Decisions evidence

- `tasksdecisions/workbookprojection` now owns the canonical task-request and
  decision descriptor v3 facts, semantic intents derived from their adopted
  view schemas, separate typed refresh/load/transactional-rebuild methods,
  separate incident rebuild methods, and a ready-checked port aggregate.
  Contract tests cover both surfaces and defensive copies of canonical facts
  and temporary execution bridges.
- Create, patch, import, conflict, and decision supersession paths receive only
  the typed row port. The view-ID-dispatched projection `LoadTx` interface and
  implementation were removed; owner-local routing selects an explicit task or
  decision method before entering the port.
- Executable provider construction moved from the broad owner root to
  `tasksdecisions/projectionprovider`, with physical SQL remaining in the
  existing owner-private provider until WS-24 and WS-25. The obsolete root
  `projection_contribution.go` was deleted.
- Projection assembly accepts the immutable owner contribution, privately
  retains `TaskDecisionRows`, and exposes only the typed port aggregate.
  Workbook and Imports consume the same assembly-provided rows rather than
  constructing Projections directly.
- Both manifest facade facts changed from
  `internal/modules/tasksdecisions` to
  `internal/modules/tasksdecisions/workbookprojection`; five obsolete exact
  root-import permissions were removed across the owner and application
  adapter. Descriptor v3 and manifest v4 shapes, public behavior, migrations,
  and database schemas did not change.
- Projections run `.cartulary/test-results/20260809T110523Z-p3278132` is retained;
  it caught the legacy runtime descriptor's stale facade value before the final
  parity run. Generation drift PASS at
  `.cartulary/test-results/20260809T110621Z-p3290087` and generated-artifact
  policy PASS at `.cartulary/test-results/20260809T110629Z-p3292838`. No
  generated file or migration was hand-edited. Raw query compilation and
  `links/readshape` remain explicit WS-26 work.

### WS-16 Timeline storage evidence

- `projections/internal/storage` now owns Timeline projection upsert, row
  deletion, and incident-clear SQL. The root execution bridge retains semantic
  source paging and mutation validation but delegates every physical write
  through this private storage boundary using the caller-owned transaction.
- `projections/internal/queryengine` now owns the complete Timeline compiled
  query surface, including table/join/expression facts. The Timeline facade
  retains only its immutable descriptor and semantic `SurfaceIntent`; provider
  assembly no longer receives owner-authored SQL.
- The exact `timeline_grid_projection` policy rule scans all production module
  sources and permits reads only in private query-engine code and writes only
  in private storage code. The obsolete owner-package SQL allowlist was
  removed, and the Records read rule now names the private query engine.
- Deleted authored `db/queries/timeline.sql` because both operations were
  unused, then ran `make generate`, which removed generated
  `internal/gen/sql/timeline.sql.go`. No generated file or migration was
  hand-edited; migration drift remains green.
- Initial Timeline run
  `.cartulary/test-results/20260809T111223Z-p3303153` is retained. It exposed
  that the semantic view-schema constant had shared the removed query-plan
  file; the constant was moved into the owner contribution before the final
  green runs. Format/generation roots are
  `.cartulary/test-results/20260809T111311Z-p3332657` and
  `.cartulary/test-results/20260809T111203Z-p3300499`.
- Public query, paging, schema, cursor, authorization, row shape, telemetry,
  Collaboration, and database behavior are unchanged. Host remains the next
  provider in deterministic rebuild order and is eligible after this
  checkpoint lint passes.

### WS-17 Host storage evidence

- Entities' `workbookprojection` facade now defines the typed Host source
  input/page, bounded query-projection row, derived Reporting fact, reader, and
  ready-checked port. The source-owner provider reads only authoritative Host,
  Records, Links, Evidence, and object facts; no projection table or executable
  projection SQL remains there.
- Private Projections storage owns Host upsert, typed deletion, and incident
  clearing. Root coordination reads paged typed source inputs inside the
  caller-owned transaction and delegates each physical mutation without
  committing, authorizing, publishing Collaboration events, or mutating source
  or history state.
- Private query-engine code owns Host filters, sort expressions, nulls-last
  keyset predicates, `limit+1`, projection scanning, and derived Reporting
  reads. Entities receives only the bounded result IDs and derived fields,
  hydrates authoritative Host/Record fields plus aliases and reusable
  identifiers for exactly those IDs, and preserves projection order.
- The source-owner Reporting provider now receives the typed derived-fact
  reader during server composition. Reporting no longer imports that provider
  as a built-in or reads Host projection SQL; it still owns source-family,
  content-class, support-reference, ordering, and output-shape semantics.
- Host descriptor and manifest v4 values now report `capabilities.query=true`
  and carry a semantic intent with the complete adopted Host field set.
  Identity remains false and is the next ordered workstream. Descriptor v3 and
  manifest v4 shapes, public routes, cursors, rows, authorization, telemetry,
  migrations, and database schemas are unchanged.
- The exact Host table rule scans all module production sources and permits
  reads only in private query-engine code and writes only in private storage.
  The production scan finds no other Host projection-table access.
- Retained run `.cartulary/test-results/20260809T113205Z-p3425579` caught the
  temporary legacy Identity import break. Retained run
  `.cartulary/test-results/20260809T113250Z-p3452819` caught generic-query
  validation of the typed Host reader and sentinel-limit preallocation. Format
  run `.cartulary/test-results/20260809T113413Z-p3481559` caught a local brace
  error before the final green format root
  `.cartulary/test-results/20260809T113442Z-p3485076`. All were corrected before
  the final validation set; no generated file or migration was hand-edited.

### WS-18 Identity storage evidence

- Extended the Entities facade with typed Identity materialization inputs,
  bounded query-projection rows, semantic intent, and derived-fact reads. Both
  Host and Identity descriptors are now query-capable, and the contribution
  carries the complete adopted field-key sets without physical SQL details.
- The source-owner provider now reads and maps only authoritative Identity,
  Records, Links, Evidence, and object facts. The prior combined executable
  `projectionprovider/provider.go` was deleted after both providers migrated;
  neither owner provider nor the owner store references a projection table.
- Private storage owns Identity upsert, typed deletion, and incident clearing.
  Private query code owns filters, sorting, nulls-last keyset predicates,
  `limit+1`, scanning, and Reporting-derived facts. Entities hydrates only the
  returned bounded IDs, restores authoritative envelope/identity fields and
  collections, and preserves the private selection order.
- The injected Entities Reporting provider now consumes typed Host and Identity
  derived facts; its final source-family, content-class, support-reference,
  ordering, and output validation remain source-owner semantics. All direct
  projection SQL was removed from the Reporting provider.
- The exact Identity table rule permits production reads only in private query
  code and writes only in private storage. Descriptor v3 and manifest v4 shapes
  are unchanged; only the Identity `capabilities.query` value changed to true.
  Public routes, row/cursor behavior, authorization, telemetry, Collaboration,
  migrations, and database schemas did not change.
- Format PASS at `.cartulary/test-results/20260809T114544Z-p3572390`. No
  generated file or migration was hand-edited. Indicator is the next provider
  in deterministic rebuild order and is eligible after checkpoint lint.

### WS-19 Indicator storage evidence

- `indicators/workbookprojection` now defines the typed Indicator
  materialization input and bounded source page. The owner-private source
  implementation derives those semantic values from Indicators, Records,
  observations, lifecycle intervals, and Links without reading or writing the
  projection table.
- Private Projections storage owns Indicator upsert, typed row deletion, and
  incident clearing. Root coordination loads typed source values within the
  caller-owned transaction, deletes an absent source row, and rebuilds by
  bounded `limit+1`/record-ID pages. It does not begin, commit, authorize,
  publish Collaboration events, or mutate authoritative/history state.
- Private query-engine code owns the compiled Indicator table, join,
  expressions, field kinds, paging, and row scan contract. The owner
  contribution carries only immutable descriptor v3 facts and the semantic
  `SurfaceIntent` field set.
- Removed owner-side executable projection writes and raw query surfaces,
  `Store.DeleteRowTx`, `Coordinator.DeleteRowTx`, and the view-ID-dispatched
  generic delete file. Indicator deletion now enters only through the typed
  `DeleteIndicatorTx` port and private physical storage.
- The exact Indicator table rule scans all production module sources and
  permits reads only in private query-engine code and writes only in private
  storage. Source-ownership coverage and the Records rule name the new source
  and private query files. Public query, row, cursor, authorization, telemetry,
  Collaboration, migration, and database-schema behavior are unchanged.
- Retained run `.cartulary/test-results/20260809T115718Z-p3663144` caught a
  missing root import after the generic delete removal. Retained run
  `.cartulary/test-results/20260809T115833Z-p3679999` caught the private
  Indicator surface missing from the exact generic-query test matrix. Both
  were corrected before the final green validation set. Format PASS at
  `.cartulary/test-results/20260809T115925Z-p3686101`; no generated file or
  migration was hand-edited. Assessment is next in deterministic rebuild
  order and becomes eligible after checkpoint lint.

### WS-20 Assessment storage evidence

- The Assessment contribution now carries only its immutable descriptor,
  semantic `SurfaceIntent`, and typed authoritative source reader. The
  characterized raw query bridge was deleted, while the existing typed
  mutation DTO and paged source enumeration remain unchanged.
- Private Projections storage owns Assessment upsert, typed deletion, and
  incident clearing. The root coordination layer validates typed mutations and
  controls rebuild paging but contains no Assessment table SQL and never takes
  transaction ownership from the caller.
- Private query-engine code owns the Assessment projection join, collection
  representation, confidence-band sort expression, field expressions, and
  compiled SQL. Its semantic record-reference tags are compiled locally, so
  Projections no longer imports `links/readshape` for this surface.
- The SQLite-only Assessment fixture writer remains explicitly named under
  `assessments/testsupport`; production table reads are permitted only in the
  private query engine and production writes only in private storage. The
  fixture path is an explicit non-production write permission pending the
  final policy partition in WS-28.
- Retained run `.cartulary/test-results/20260809T120622Z-p3756240` rejected a
  draft cross-owner `projections/testsupport` import, so the fixture stayed
  with its source-owner test package. Retained JSON-shape run
  `.cartulary/test-results/20260809T120803Z-p3768514` found the resulting empty
  support directory; it was removed before the final green run. Format PASS at
  `.cartulary/test-results/20260809T120717Z-p3760848`; no generated file or
  migration was hand-edited. Artifact is next in deterministic rebuild order
  and becomes eligible after checkpoint lint.

### WS-21 Artifact storage evidence

- `artifacts/workbookprojection` now owns a typed 55-field materialization DTO,
  bounded source pages, derived-fact reader, and ready-checked reader port in
  addition to its existing eight semantic intents and typed row/rebuild ports.
  The DTO retains nullable subtype facts explicitly without exposing a table or
  executable query plan.
- The source-owner provider derives each Artifact row from authoritative
  Artifacts, Records, subtype, and Link facts and decodes the named result into
  the typed DTO. Private storage owns insert, typed row deletion, and incident
  clearing; private coordination performs `limit+1` record-ID rebuild paging
  in the caller-owned transaction.
- Private query-engine code owns all eight compiled surfaces, discriminators,
  enum sorts, tag/record/party/risk collections, and the typed Reporting fact
  read. Its private view/discriminator map is tested for exact equality with
  the facade's semantic source filters; no concrete `surfacecatalog` import
  crosses the owner port.
- The Artifact Reporting provider now receives only the typed derived-fact
  reader during server composition. It retains source-family, content-class,
  path-prefix, support-reference, ordering, and six included Artifact-family
  semantics; Reporting no longer imports Artifact SQL or treats this provider
  as a built-in.
- Added the exact Artifact table rule: production reads are confined to
  private query-engine code and writes to private storage. The raw owner query
  bridge and executable storage provider were deleted. Descriptor v3,
  manifest v4, public rows/cursors/routes, migrations, and schema are
  unchanged.
- The first format invocation stopped before producing a run root because the
  moved eight-surface test still had its old authored package selector. The
  selector was updated registry-first and `make generate` PASS at
  `.cartulary/test-results/20260809T122358Z-p3815417` refreshed only the
  generator-owned topology index. Retained boundary run
  `.cartulary/test-results/20260809T122132Z-p3800485` rejected a concrete
  `surfacecatalog` import, and retained JSON-shape run
  `.cartulary/test-results/20260809T122342Z-p3814707` identified the expected
  pre-generation topology drift. Both were corrected before the final green
  validation set. Format PASS at
  `.cartulary/test-results/20260809T122216Z-p3801484`; no generated file or
  migration was hand-edited. Evidence is next in deterministic rebuild order
  and becomes eligible after checkpoint lint.

### WS-22 Evidence storage evidence

- `evidence/workbookprojection` now owns the typed Evidence materialization
  DTO and bounded source page in addition to its immutable descriptor and
  semantic intent. Nullable attachment, collector, source, and lifecycle facts
  remain explicit without exposing physical storage or executable SQL.
- The source-owner provider derives rows from authoritative Evidence, Records,
  object-blob, and Link facts only. Private Projections storage owns insert,
  typed row deletion, and incident clearing; private coordination performs
  bounded record-ID rebuild paging in the caller-owned transaction.
- Private query-engine code owns the Evidence table join, attachment-state and
  collection expressions, filters, ordering, scan contract, and compiled SQL.
  The generic query matrix now resolves the Evidence surface from this private
  plan rather than an owner-side raw query bridge.
- The exact Evidence table rule permits production reads only in private
  query-engine code and writes only in private storage. The executable owner
  provider and raw query-surface files were deleted, while Evidence remains the
  owner of authoritative mapping and attachment semantics.
- Public routes, row/cursor behavior, authorization, Collaboration,
  descriptor v3, manifest v4, migrations, and database schema are unchanged.
  Format PASS at `.cartulary/test-results/20260809T123027Z-p3826692`; no
  generated file or migration was hand-edited. Party is next in deterministic
  rebuild order and becomes eligible after checkpoint lint.

### WS-23 Party storage evidence

- `parties/workbookprojection` now owns a typed Party materialization DTO,
  bounded source page, and source-reader port alongside its immutable
  descriptor and semantic intent. Nullable Party fields stay explicit without
  exposing a table, query expression, or executable provider.
- The source-owner provider derives rows only from authoritative Parties and
  Records facts. Private storage owns insert, row deletion, and incident
  clearing; private coordination performs `limit+1` record-ID paging while
  preserving caller control of the transaction.
- Private query-engine code owns the Party table join, fields, filters,
  ordering, row scan contract, and compiled plan. Both the owner raw-query
  bridge and its executable projection writer were deleted.
- The exact Party table rule permits production reads only in private query
  code and writes only in private storage. Party import, reuse, patch,
  conflict, reporting, and authoritative/history behavior remain owned by
  Parties and pass the owner suites.
- Public routes, rows/cursors, authorization, Collaboration, descriptor v3,
  manifest v4, migrations, and database schema are unchanged. Format PASS at
  `.cartulary/test-results/20260809T123744Z-p3905885`; no generated file or
  migration was hand-edited. Task request is next in deterministic rebuild
  order and becomes eligible after checkpoint lint.

### WS-24 Task-request storage evidence

- `tasksdecisions/workbookprojection` now defines the typed Task-request
  materialization DTO, bounded source page, source reader, derived fact, and
  Reporting reader port. The shared contribution retains the Decision
  executable bridge only for its immediately following ordered workstream.
- The Tasks/Decisions-private source derives queue rows from authoritative Task
  Request, Records, and Link facts. Private Projections storage owns insert,
  row deletion, and incident clearing; private coordination performs bounded
  record-ID rebuild paging in the caller-owned transaction.
- Private query-engine code owns Task-request fields, collections, filters,
  ordering, scanning, compiled plan, and typed Reporting fact reads. Server
  composition injects the reader into the source-owner Reporting provider,
  which retains source-family, content-class, support-reference, and output
  semantics without reading the projection table directly.
- The exact Task-request table rule permits production reads only in the
  private query engine and writes only in private storage. The old Task-request
  writer and raw surface were removed from the shared Decision provider files.
- Retained Projections run
  `.cartulary/test-results/20260809T124901Z-p4027557` caught the new surface
  missing from the exact generic-query matrix. Retained boundary run
  `.cartulary/test-results/20260809T125003Z-p4036781` caught the new typed
  owner-private imports and authoritative source reader missing from their
  exact allowlists. Both were corrected before the final green validation set.
  Format PASS at `.cartulary/test-results/20260809T124718Z-p3970582`; public
  behavior, descriptor v3, manifest v4, migrations, and schema are unchanged.
  No generated file or migration was hand-edited. Decision is next in rebuild
  order and becomes eligible after checkpoint lint.

### WS-25 Decision storage evidence

- `tasksdecisions/workbookprojection` now defines typed Decision
  materialization inputs, bounded pages, source-reader and derived-fact ports.
  The shared contribution contains only descriptors, semantic intents, and
  typed authoritative readers; its executable and raw-query bridges are gone.
- The Tasks/Decisions-private source derives Decision rows from authoritative
  Decision, Records, and Link facts, including affected counts, deterministic
  supersession target selection, and active incoming supersession state.
  Private storage owns insert, row deletion, and incident clearing.
- Private query-engine code owns Decision fields, support/affected collections,
  filters, ordering, scan contract, compiled plan, and typed Reporting facts.
  The injected source-owner Reporting provider now receives both Task-request
  and Decision derived facts without any projection-table SQL.
- The exact Decision table rule permits production reads only in private query
  code and writes only in private storage. The final owner-side projection
  provider and raw surface files were deleted; rebuilds remain bounded and
  caller-transaction-controlled.
- Retained boundary run
  `.cartulary/test-results/20260809T125812Z-p4107171` caught the typed Decision
  source missing from the exact Records-envelope allowlist; it was corrected
  before the final green validation set. Format PASS at
  `.cartulary/test-results/20260809T125606Z-p4045637`; public behavior,
  descriptor v3, manifest v4, migrations, and schema are unchanged. No
  generated file or migration was hand-edited. All ten providers have now
  crossed the physical-storage boundary; WS-26 is eligible after checkpoint
  lint.

### WS-26 Query seam closure evidence

- Deleted `providercontract/query.go`; provider contracts now contain only
  immutable descriptor facts and semantic `SurfaceIntent` values. No owner
  facade exposes SQL, field expressions, scan kinds, or query-engine types.
- Moved generic query construction, scanning, field kinds, and every compiled
  provider plan into the Go-private `projections/internal/queryengine`
  package. Runtime `ProviderDescriptor` no longer carries physical plans;
  each provider factory binds its private plans directly to its executable
  provider.
- Replaced `links/readshape` with private semantic SQL compilation for active
  Links, active Tags, record refs, party refs, and record-tag refs. Projections
  has no remaining import of the Links helper package, while collection output
  and stable item-ref formats remain unchanged.
- Added exact equality evidence across each generic semantic intent, private
  compiled plan, and registered view-schema field set. Existing Host and
  Identity private tests continue proving projection-owned keyset selection,
  `limit+1`, no `OFFSET`, and exact bounded-ID hydration behavior.
- Retained Projections run
  `.cartulary/test-results/20260809T130340Z-p4119894` exposed that moving the
  generic engine had displaced the foundation intent registry. The registry
  was restored as a separate fail-closed private type before the final green
  set. Format PASS at `.cartulary/test-results/20260809T130609Z-p4134998`;
  public rows, filters, sorting, cursors, saved views, descriptor v3, manifest
  v4, migrations, and schema are unchanged. No generated file or migration
  was hand-edited. WS-27 is eligible after checkpoint lint.

### WS-27 Test capability migration evidence

- Added `projections/testsupport.Capability` as a field-private wrapper over
  only the named Timeline, Entities, Indicators, and Evidence facade ports
  required by current shared application tests. It exposes no production
  adapter, runtime, catalog, coordinator, store, SQL, or assembly type.
- `httptestx` converts the existing application composition callback
  immediately into that capability and no longer stores
  `timelineassembly.Bundle`. Every characterized caller keeps its named typed
  method, so the migration changes composition ownership without granting a
  broader test API.
- The caller-matrix guard now rejects a local aggregate, the old catalog and
  Evidence helper paths, exported capability fields, or omission of the
  Projections-owned constructor. The support inventory marks this package
  owner-local, support-scanned, runtime-excluded, and non-service-starting;
  production boundary policy still rejects imports of it.
- Retained Projections run
  `.cartulary/test-results/20260809T131600Z-p37603` exposed an over-broad source
  assertion that also rejected the required one-shot server callback
  signature. Retained JSON-shape run
  `.cartulary/test-results/20260809T131708Z-p45356` then caught the new support
  inventory row out of lexical order. Both test/policy defects were corrected
  without weakening the capability boundary. Format PASS at
  `.cartulary/test-results/20260809T131550Z-p34367`; production behavior,
  descriptor v3, manifest v4, migrations, and schema are unchanged. No
  generated file or migration was hand-edited. WS-28 is eligible after
  checkpoint lint.

### WS-28 Contract, policy, and harness reconciliation evidence

- Replaced the three remaining package-wide table permissions with exact
  provider-file rules and normalized all ten rules into rebuild order. Reads
  and counts are confined to each provider's private query/storage files;
  writes and locks are confined to its private storage file.
- Extended the authored boundary checker to scan runtime-excluded support roots
  separately for Projections-table access. The production and test-fixture
  allowlists are now distinct: only Timeline assertions may read a projection
  table, only Assessment's SQLite/Postgres-compatible fixture may write one,
  and the other eight test permissions are empty. Built-in positive and
  negative fixtures prove that a test permission cannot authorize production
  access.
- Added routed static equality evidence across active descriptor table IDs,
  the ten boundary rules, the Recovery rebuild-state set, and tables created by
  migrations and classified to the Projections schema owner. The same test
  asserts the complete test-fixture permission set. Manifest v4 parity and all
  query capability values remain code-backed and green.
- Added the explicit storage-ownership verification claim to the Projections
  owner contract and routed the equality test through its existing assembly
  row. Registry inputs were changed first and generated topology was refreshed
  only through `make generate`, PASS at
  `.cartulary/test-results/20260809T132649Z-p60794`.
- Retained boundary run
  `.cartulary/test-results/20260809T132627Z-p59566` caught the initial support
  scan applying projection-specific empty fixture permissions to Records.
  Retained JSON-shape run
  `.cartulary/test-results/20260809T132627Z-p59441` correctly required topology
  regeneration after the verification-owner and family inputs changed. Both
  were corrected through their owner inputs. Format PASS at
  `.cartulary/test-results/20260809T132615Z-p56102`; no generated file or
  migration was hand-edited. WS-29 is eligible after checkpoint lint.

### WS-29 Legacy and root removal evidence

- Moved the characterized executable provider registry, query service, rebuild
  service, coordinator, telemetry, and owner-row implementations into the
  Go-private `projections/internal/runtime` package. The Projections root
  directory now contains no Go file and the manifest root-importer set is
  empty.
- Expanded `adapters.New` into the sole fail-closed production constructor for
  the immutable declarative catalog, semantic-intent registry, private storage,
  executable catalog, consumer-owned Workbook/Recovery/Revisions ports, and all
  eight typed owner facades. Only `internal/app/projectionassembly` imports the
  adapter; application and owner callers receive narrow interfaces or facade
  ports rather than concrete runtime types.
- Deleted the final assembly compatibility delegates and concrete
  Store/Catalog/Coordinator/Query fields. Timeline, Workbook, Imports,
  Recovery, Revisions, restore probing, source-text loading, import rebuilding,
  and shared tests now consume named capabilities. Host and Identity stay on
  their typed Entities query readers and do not leak through the generic
  Workbook adapter.
- Moved the root boundary guard into `projections/adapters`, where it requires
  an empty root, rejects every root import, permits only projection assembly to
  import the adapter, and rejects private-package imports. A single exact policy
  allowance lets the adapter name Workbook's consumer-owned query interface;
  the broader source-owner prohibition remains intact.
- Reconciled Appendix I, the development guide, and projection-manifest
  maintenance guidance with semantic intents, private compiled plans, exact
  source/storage ownership, the empty root, and current test locations. The
  domain vocabulary remained unchanged because the WS-01 audit found no drift.
- Retained Projections run
  `.cartulary/test-results/20260809T134657Z-p92865` caught an assembly test that
  incorrectly required Host/Identity to use the generic Workbook adapter.
  Retained boundary run
  `.cartulary/test-results/20260809T135227Z-p107108` caught the adapter's
  legitimate consumer-interface import under an over-broad source-owner rule.
  Both were corrected without widening owner or private-package access.
- The first retained `test-fast` run
  `.cartulary/test-results/20260809T135326Z-p112484` found a stale Network Flow
  test access to the removed Coordinator field. The second retained run
  `.cartulary/test-results/20260809T135809Z-p216950` found that Projections test
  support had incorrectly absorbed an Indicators-owned source-text port,
  creating a same-package import cycle. The caller now uses the narrow
  `SourceTextRows` capability, and the Indicators source-text capability stays
  separately owner-typed in the shared harness. The final fast suite passes all
  353 units.
- Format PASS at
  `.cartulary/test-results/20260809T140156Z-p271182`; `make generate` was run
  through its owner at `.cartulary/test-results/20260809T134641Z-p90360`. No
  migration was changed, and no generated file was hand-edited. Checkpoint
  Markdown lint PASS at
  `.cartulary/test-results/20260809T140614Z-p341139`; WS-30 is now eligible.

### WS-30 Validation and final handoff evidence

Every changed owner was revalidated from its authored owner route before the
dedicated Projections and broad checks:

| Owner | Unit/static run root | Service-backed run root |
| --- | --- | --- |
| `app.server` | `.cartulary/test-results/20260809T140727Z-p343287` | `.cartulary/test-results/20260809T140953Z-p444407` |
| `module.workbook` | `.cartulary/test-results/20260809T140727Z-p343292` | `.cartulary/test-results/20260809T140953Z-p444409` |
| `module.recovery` | `.cartulary/test-results/20260809T140727Z-p343286` | `.cartulary/test-results/20260809T140953Z-p444415` |
| `module.revisions` | `.cartulary/test-results/20260809T141157Z-p534514` | `.cartulary/test-results/20260809T141339Z-p624535` |
| `module.timeline` | `.cartulary/test-results/20260809T141157Z-p534520` | `.cartulary/test-results/20260809T141339Z-p624537` |
| `module.entities` | `.cartulary/test-results/20260809T141157Z-p534531` | `.cartulary/test-results/20260809T141339Z-p624541` |
| `module.indicators` | `.cartulary/test-results/20260809T141457Z-p707036` | `.cartulary/test-results/20260809T141537Z-p735250` |
| `module.assessments` | `.cartulary/test-results/20260809T141457Z-p707035` | `.cartulary/test-results/20260809T141537Z-p735249` |
| `module.artifacts` | `.cartulary/test-results/20260809T141457Z-p707040` | `.cartulary/test-results/20260809T141537Z-p735256` |
| `module.evidence` | `.cartulary/test-results/20260809T141618Z-p761511` | `.cartulary/test-results/20260809T141727Z-p837051` |
| `module.parties` | `.cartulary/test-results/20260809T141618Z-p761518` | `.cartulary/test-results/20260809T141727Z-p837058` |
| `module.tasksdecisions` | `.cartulary/test-results/20260809T141618Z-p761530` | `.cartulary/test-results/20260809T141727Z-p837073` |
| `module.reporting` | `.cartulary/test-results/20260809T141829Z-p909155` | `.cartulary/test-results/20260809T142031Z-p969781` |
| `module.imports` | `.cartulary/test-results/20260809T141829Z-p909168` | `.cartulary/test-results/20260809T142031Z-p969788` |
| `module.networkflow` | `.cartulary/test-results/20260809T141829Z-p909171` | `.cartulary/test-results/20260809T142031Z-p969798` |
| `module.projections` | `.cartulary/test-results/20260809T142207Z-p1023102` | `.cartulary/test-results/20260809T142217Z-p1024444` |

The exact ordered repository gates produced these retained successful roots:

| Command | Result | Run root |
| --- | --- | --- |
| `make backend-module-boundary-check` | PASS, 3/3 | `.cartulary/test-results/20260809T142231Z-p1025828` |
| `make json-shape-check` | PASS, 3/3 | `.cartulary/test-results/20260809T142239Z-p1026226` |
| `make generate-drift` | PASS, 4/4 | `.cartulary/test-results/20260809T142246Z-p1026741` |
| `make generated-artifact-policy-check` | PASS, 3/3 | `.cartulary/test-results/20260809T142258Z-p1029556` |
| `make migration-drift` | PASS, 5/5 | `.cartulary/test-results/20260809T142308Z-p1030041` |
| `make harness-contract` | PASS, 2/2 | `.cartulary/test-results/20260809T142319Z-p1032891` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260809T142328Z-p1033432` |
| `make agent-finalize` | PASS, 1/1 | `.cartulary/test-results/20260809T142335Z-p1034189` |
| `make test-fast` | PASS, 353/353 | `.cartulary/test-results/20260809T142351Z-p1036933` |
| `make browser-e2e-webserver-backed` | PASS, 62/62 | `.cartulary/test-results/20260809T142415Z-p1037622` |
| `make browser-e2e-stateful` | PASS, 36/36 | `.cartulary/test-results/20260809T142821Z-p1090969` |
| `make build` | PASS, 7/7 | `.cartulary/test-results/20260809T143026Z-p1117327` |
| `make check` | PASS, 746/746 | `.cartulary/test-results/20260809T143806Z-p1309728` |
| `make release-check` | PASS, 917/917 | `.cartulary/test-results/20260809T145712Z-p1709336` |

`make agent-finalize` ran with `RESULTS_DIR` unset because no already successful
full warm-check root existed at that point. Retained-run maintenance was
therefore skipped exactly as the repository procedure requires; no other
required check was skipped.

The first `make check` run at
`.cartulary/test-results/20260809T143050Z-p1148227` correctly found capitalized
error strings introduced by the new typed paths and stale OTEL evidence paths.
The error strings were normalized, the OTEL conformance owner was updated, and
focused `make otel-conformance` and `make lint` passed at
`.cartulary/test-results/20260809T143725Z-p1292149` and
`.cartulary/test-results/20260809T143738Z-p1298216` before the green broad
retry. The first `make release-check` run at
`.cartulary/test-results/20260809T144510Z-p1486302` had one existing Timeline
browser-support row time out at 60 seconds; its owner target passed 20/20 at
`.cartulary/test-results/20260809T145614Z-p1682180`, and the complete release
retry then passed.

Final read-only scans prove the Projections root contains no Go file, no
production file imports the removed root package, only projection assembly
imports the production adapter, and no raw `QuerySurface`, `links/readshape`,
legacy catalog, or legacy coordinator symbol remains in production. The only
`ProjectionCatalog` text is a negative test marker. `git diff --check` passes.

The final change spans adopted Core/ADR text, typed source-owner facades,
private Projections runtime/storage/query code, application composition,
consumer ports, Reporting readers, characterization tests, manifest and
boundary policy, verification routing, harness support, and explanatory
guidance. Descriptor v3, manifest v4, all public route/wire/cursor/row/event
contracts, and the database schema are retained. The unused Timeline authored
query input and its generated output were removed through `make generate`;
existing migrations were not changed, and no generated file was hand-edited.
All identified compatibility, storage-ownership, transaction, query-seam,
restore, Reporting, test-composition, and policy risks are closed. Final
tracker checkpoint Markdown lint PASS at
`.cartulary/test-results/20260809T150727Z-p1878996`.

## 8. Validation order

After each implementation workstream, run the changed owner's `make test-slice`
and `make service-backed-test-slice`, the affected `module.projections` rows,
and `make backend-module-boundary-check`.

WS-30 runs:

1. All changed-owner unit/static and service-backed slices.
2. Full `module.projections` unit/static and service-backed slices.
3. `make backend-module-boundary-check`
4. `make json-shape-check`
5. `make generate-drift`
6. `make generated-artifact-policy-check`
7. `make migration-drift`
8. `make harness-contract`
9. `make lint-markdown`
10. `make agent-finalize`; pass `RESULTS_DIR` only for an already successful
    full warm-check root, otherwise record that it was unset.
11. `make test-fast`
12. `make browser-e2e-webserver-backed`
13. `make browser-e2e-stateful`
14. `make build`
15. `make check`
16. `make release-check`

Stop at the first failure, retain its run root, and return to the most recent
independently green workstream.

## 9. Historical handoff record

The following history is retained from the prior tracker revision.

| Time | Session | Preserved result | Preserved blockers/next action |
| --- | --- | --- | --- |
| 2026-08-07 22:06 EDT | Planning/documentation | Inventoried all 29 target files and discovered file-specific root imports, distributed SQL, peer helper, and Timeline-backed testutil. Markdown lint passed. | Architecture adoption, SQL inventory, characterization, policy closure, and implementation were pending. |
| 2026-08-07 22:55 EDT | NLSpec tracker revision | Specified sole adapter, immutable contract, private packages, eight facades, ten providers, characterization matrices, and retained documentation lint root `.cartulary/test-results/20260808T030403Z-p2181649`. | Tracker-local Authorization A/B and five `RB-*` blockers were recorded; WS-00 now retires those artificial gates while preserving the technical findings. |
| 2026-08-09 07:19 EDT | Remediation planning | Re-audited adopted Core, live imports, SQL locations, provider manifest, boundary policies, test routing, Recovery/Revisions/Reporting/testutil seams, and confirmed the legacy policy passes its current checks. | Execute WS-00 through WS-30 with a tracker checkpoint between each workstream. |
| 2026-08-09 08:20 EDT | WS-02 characterization | Added direct restore atomicity, typed deletion transaction, and test-capability caller evidence; fixed no-provider restore readiness to return `not_applicable`; all affected owner baselines and boundaries are green. | Begin WS-03 adapter/contract foundation without migrating consumers. |
| 2026-08-09 08:34 EDT | WS-03 foundation | Added fail-closed adapter composition, immutable canonical descriptors, semantic intents, private package containment, typed owner contribution shells, and routed constructor tests. No production consumer was migrated. | Migrate projection assembly as the sole adapter importer in WS-04. |
| 2026-08-09 08:40 EDT | WS-04 projection assembly | Made projection assembly the sole adapter importer and the canonical descriptor/manifest source while retaining isolated characterized execution delegates for the ordered migrations. | Migrate Workbook to its query port and immutable descriptors in WS-05. |
| 2026-08-09 08:54 EDT | WS-05 Workbook | Removed all Workbook root imports, switched catalog validation to immutable canonical descriptors, kept execution behind Workbook-owned exact-surface query ports, and tightened the owner allowlist. | Complete the checkpoint lint, then migrate Recovery to consumer-owned rebuild and state ports in WS-06. |
| 2026-08-09 09:07 EDT | WS-06 Recovery | Replaced Recovery's catalog/rebuild reach-through with its typed port aggregate, moved ten-table restore-state facts to the canonical contract with exact adapter validation, updated the restore probe interface seam, and deleted the root recovery-state API. | Complete the checkpoint lint, then inject only Revisions' projection services in WS-07. |
| 2026-08-09 09:16 EDT | WS-07 Revisions | Removed projection catalog knowledge from revision composition, based record/view validation on public view facts and source-owner contributions, and injected only `revisions.ProjectionServices` for transactional rebuild/load behavior. | Complete the checkpoint lint, then migrate Timeline projection behavior behind `timeline/workbookprojection` in WS-08. |
| 2026-08-09 09:28 EDT | WS-08 Timeline | Made `timeline/workbookprojection` the typed owner facade for contribution/source/write/rebuild contracts, split peer-entity refresh from Timeline writes, migrated test helpers, and confined the legacy row bridge to projection assembly. | Complete the checkpoint lint, then migrate host/identity behavior behind `entities/workbookprojection` in WS-09. |
| 2026-08-09 09:49 EDT | WS-09 Entities | Replaced generic host/identity projection dispatch and source-owner root construction with canonical typed contribution, writer, rebuilder, and composition ports; tightened manifest and import policy facts. | Complete the checkpoint lint, then migrate Indicators behind `indicators/workbookprojection` in WS-10. |
| 2026-08-09 10:01 EDT | WS-10 Indicators | Replaced the root Indicator projection contribution and generic view-ID port with a pure workbook facade, typed row/rebuild ports, and an Indicator-owned executable-provider constructor; tightened manifest and boundary facts. | Complete the checkpoint lint, then migrate Assessments behind `assessments/workbookprojection` in WS-11. |
| 2026-08-09 10:21 EDT | WS-11 Assessments | Moved Assessment descriptor/intent/source DTO/mutation/rebuild contracts behind `assessments/workbookprojection`, migrated application consumers to typed ports, and made owner/provider composition package-scoped. | Complete the checkpoint lint, then migrate Artifacts behind `artifacts/workbookprojection` in WS-12. |
| 2026-08-09 10:31 EDT | WS-12 Artifacts | Moved all eight Artifact surfaces behind semantic facade facts and typed row/rebuild ports; injected the port into Workbook, contextual notes, imports, conflicts, and tests; removed source-owner root construction. | Complete the checkpoint lint, then migrate Evidence behind `evidence/workbookprojection` in WS-13. |
| 2026-08-09 10:47 EDT | WS-13 Evidence | Moved Evidence descriptor/intent and refresh/load/support/rebuild behavior behind `evidence/workbookprojection`; injected the typed port into every mutation/import/test composition path; removed root import exceptions and executable-provider leakage. | Complete the checkpoint lint, then migrate Parties behind `parties/workbookprojection` in WS-14. |
| 2026-08-09 10:56 EDT | WS-14 Parties | Moved Party descriptor/intent and typed refresh/load/rebuild behavior behind `parties/workbookprojection`; injected the port into Workbook and Imports; removed source-owner root construction and exact root permissions. | Complete the checkpoint lint, then migrate Tasks/Decisions behind its separate typed methods in WS-15. |
| 2026-08-09 11:08 EDT | WS-15 Tasks/Decisions | Moved both provider descriptors/intents and separate task/decision refresh/load/rebuild methods behind `tasksdecisions/workbookprojection`; removed generic view-ID load dispatch, direct application construction, and the owner-root contribution. | Complete the checkpoint lint, then begin provider-ordered physical storage migration with Timeline in WS-16. |
| 2026-08-09 11:17 EDT | WS-16 Timeline storage | Moved all Timeline projection-table writes and compiled query-plan facts into private Projections packages, tightened the exact table rule, and removed unused authored/generated sqlc operations through the owner generator. | Complete the checkpoint lint, then move Host projection storage, differential query selection, and Reporting reads in WS-17. |
| 2026-08-09 11:39 EDT | WS-17 Host storage | Split Host derivation, physical storage, bounded query selection, exact-ID hydration, and Reporting through typed owner ports; tightened exact SQL ownership and made only Host query-capable. | Complete the checkpoint lint, then apply the same typed storage/query/Reporting architecture to Identity in WS-18. |
| 2026-08-09 11:50 EDT | WS-18 Identity storage | Completed the typed Entities architecture for Identity, deleted the legacy combined projection provider, fully injected Host/Identity Reporting facts, and tightened exact SQL ownership. | Complete the checkpoint lint, then move Indicator projection storage, query plans, lifecycle writes, and typed deletion in WS-19. |
| 2026-08-09 12:00 EDT | WS-19 Indicator storage | Split typed Indicator derivation from private projection storage/query plans, removed owner-side projection SQL and the final generic deletion dispatch, and tightened the exact table rule. | Complete the checkpoint lint, then move Assessment projection storage and query plans in WS-20. |
| 2026-08-09 12:08 EDT | WS-20 Assessment storage | Moved Assessment persistence and compiled query facts into private Projections packages, retained typed source mapping in Assessments, removed its raw query bridge, and isolated the SQLite-only fixture write. | Complete the checkpoint lint, then migrate all eight Artifact surfaces and Reporting facts in WS-21. |
| 2026-08-09 12:24 EDT | WS-21 Artifact storage | Split typed 55-field Artifact derivation from private storage, moved eight compiled surfaces and Reporting reads private, injected the source-owner Reporting provider, and added the exact table rule. | Complete the checkpoint lint, then migrate Evidence attachment-state storage/query behavior in WS-22. |
| 2026-08-09 12:34 EDT | WS-22 Evidence storage | Split typed authoritative Evidence derivation from private storage and moved the attachment-state compiled surface behind the query engine with an exact table rule. | Complete the checkpoint lint, then migrate Party projection storage/query behavior in WS-23. |
| 2026-08-09 12:40 EDT | WS-23 Party storage | Split typed authoritative Party derivation from private storage, moved compiled query behavior private, and tightened the exact table rule. | Complete the checkpoint lint, then migrate task-request projection storage and Reporting facts in WS-24. |
| 2026-08-09 12:51 EDT | WS-24 Task-request storage | Split typed queue derivation from private storage, moved its compiled surface and Reporting facts private, and tightened the exact table rule. | Complete the checkpoint lint, then migrate Decision storage, supersession query behavior, and Reporting facts in WS-25. |
| 2026-08-09 12:59 EDT | WS-25 Decision storage | Split typed supersession derivation from private storage, moved its compiled surface and Reporting facts private, and removed the final owner-side projection SQL. | Complete the checkpoint lint, then close the raw query seam and Links coupling in WS-26. |
| 2026-08-09 13:11 EDT | WS-26 query seam closure | Deleted the raw query contract, moved plan and engine types private, removed Links-helper coupling, and added semantic-intent/plan/schema equality. | Complete the checkpoint lint, then replace Timeline-owned HTTP test composition in WS-27. |
| 2026-08-09 13:18 EDT | WS-27 test capability migration | Replaced the Timeline-bundle-backed HTTP test capability with a Projections-owned wrapper over named typed facade ports and retained the exact caller guard. | Complete the checkpoint lint, then reconcile the final contract, table-policy, manifest, and harness facts in WS-28. |
| 2026-08-09 13:28 EDT | WS-28 contract/policy/harness reconciliation | Established exact ten-table production and test-fixture rules, four-way storage equality, and a routed storage-ownership verification claim. | Complete the checkpoint lint, then remove all root/legacy compatibility surfaces in WS-29. |
| 2026-08-09 14:02 EDT | WS-29 legacy/root removal | Emptied the Projections root, moved executable behavior private, made the adapter the sole complete constructor, removed compatibility fields/imports, separated the misplaced Indicators test capability, and reconciled explanatory guidance. | Complete the checkpoint lint, then execute the ordered validation and final handoff in WS-30. |
| 2026-08-09 15:07 EDT | WS-30 validation and final handoff | Revalidated every changed owner, all Projections rows, boundaries, contracts, generation, migration, harness, fast, browser, build, check, and release gates; repaired lint/OTEL drift, confirmed a transient browser timeout with a green owner and full release retry, and closed the final tracker checkpoint. | Complete; no workstream or Definition-of-Done blocker remains. |

## 10. Implementation Definition of Done

| ID | Pass condition | Status |
| --- | --- | --- |
| DOD-01 | Root production importer/API set is empty; only projection assembly imports adapters; private packages are compiler-contained; production/test permissions differ. | DONE |
| DOD-02 | Exact fail-closed construction returns immutable descriptors and non-nil consumer/owner ports without concrete leakage. | DONE |
| DOD-03 | Only typed mutation/deletion ports remain and every writer preserves caller-owned transaction control. | DONE |
| DOD-04 | Every production access to all ten projection tables executes inside private Projections storage/query code, including Reporting and host/identity. | DONE |
| DOD-05 | Exact equality holds among active descriptor table IDs, ten boundary rules, recovery-state IDs, and schema-owned projection tables; descriptor v3/manifest v4 parity passes. | DONE |
| DOD-06 | Every characterization matrix passes before and after its affected migration. | DONE |
| DOD-07 | Public query/row/cursor/auth/saved-view/WS/telemetry/schema/migration and source/reporting semantics remain owner-conformant. | DONE |
| DOD-08 | Core 00/01/04 and the adopted implementation ADR agree; no tracker-local authorization gate remains. | DONE |
| DOD-09 | Every workstream has an independently green evidence checkpoint and rollback boundary. | DONE |
| DOD-10 | Existing migrations remain; generated files are tool-owned; runtime/evidence does not consume Markdown. | DONE |
| DOD-11 | WS-30 results and run roots are retained; every skipped check has an explicit reason. | DONE |

Completion requires WS-00 through WS-30 and DOD-01 through DOD-11 to be
`DONE`. Planning completeness or a green legacy policy is not implementation
completion.
