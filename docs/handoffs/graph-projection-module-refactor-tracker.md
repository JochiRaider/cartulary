# Graph Projection Remediation Tracker and Handoff

## 1. Control Status

This document is the controlling execution ledger for Graph Projection remediation. A workstream may start only after its dependencies are marked `complete` here. After each workstream, update this tracker with substantive edits, validation commands, result roots, deviations, and the next unblocked workstream before beginning further work.

| Item | Current value |
| --- | --- |
| Target module | `internal/modules/graphprojection` |
| Owner specification | `docs/graph_projection_nlspec.md` |
| Domain vocabulary | `docs/domain.md` |
| Harness mechanics | `docs/testing-harness-nlspec.md` |
| Planning baseline | Clean `main` at `c9524cc6db47449d50ffdf634aec88f5b92e8a70` on 2026-07-13 |
| Current workstream | `WS-09` |
| Overall status | Remediation implementation complete; final checks passed |
| Acceptance posture | GP-AC-001 through GP-AC-069 are `implemented` with executable selector and fixture evidence |
| Boundary posture | No public Graph route, Graph-owned authorization, workbook ownership, source mutation, or Core 05 publication claim |

Authority order is the adopted Core 00 through Core 04 document set, the Graph Projection NLSpec for Graph Projection behavior, the Reporting and Network Flow owner NLSpecs for their adapters, `docs/domain.md` for terminology and concept boundaries, and the Testing Harness NLSpec for command and evidence mechanics. Core 05 applies only if a later claim-bearing publication boundary is activated.

## 2. Current Repository Baseline

The Graph Projection module currently contains 37 Go files across the root package and its `fixturetest`, `postgresbinding`, and `postgresstore` subpackages. Its production consumers are Network Flow's ephemeral projection adapter and Reporting's transaction-scoped projection binding reader. There is no production assembly for retained Graph Projection operations or queries.

Current evidence state:

- all 36 registered GP-FIX directories exist;
- every backend-unit and backend-store fixture has an operation-appropriate semantic input and reviewed exact output;
- backend-unit and backend-store fixture tests execute manifests through shared verifier paths and exact comparators;
- the conformance matrix contains 69 `implemented` criteria and no `planned` or `deferred` criteria;
- admission, validation, direct mapping, aggregation, resources, lifecycle, and query responsibilities now have separate implementation files behind the facade;
- the PostgreSQL store requires stable cursor key material through its options constructor; query digest ownership, cursor validation, retained expiry, graph-order traversal, and consumer binding semantics are implemented.

Planning baselines:

| Target | Result | Result root |
| --- | --- | --- |
| `make backend-unit` | PASS | `.cartulary/test-results/20260713T142535Z-p5968` |
| `make backend-store` | PASS | `.cartulary/test-results/20260713T142623Z-p30341` |
| `make check` | PASS | `.cartulary/test-results/20260713T163435Z-p509687` |

These passing targets establish final repository stability for this remediation. They are implementation conformance evidence only and are not a Core 05 claim-publication boundary.

## 3. Gap Ledger

| Gap | Area | Required durable correction | Risk if unresolved | Workstream | Status |
| --- | --- | --- | --- | --- | --- |
| G1 | Control documentation | Replace stale and contradictory tracker content with this current ledger | Execution and handoff can rely on false completion state | WS-00 | complete |
| G2 | Specification and contracts | Close outer-operation errors, query reasons, retryability, serialization, and whole-input-limit behavior | Callers and fixtures cannot classify exact outcomes | WS-01 | complete |
| G3 | Fixtures and conformance mechanics | Make all manifests semantically executable and replace label-only goldens | Passing tests can overstate behavior | WS-02, WS-04 through WS-08 | complete |
| G4 | Module API and cohesion | Narrow the facade, separate lifecycle/query errors and exact DTOs, and split mixed implementation responsibilities | Callers couple to private state and source data | WS-03, WS-04 | complete |
| G5 | Admission and validation | Use a closed issue registry and ordered phase runner; enforce exact identities, limits, and output schema | Invalid issue combinations and unstable IDs remain possible | WS-04 | complete |
| G6 | Derivation and failure behavior | Make stages deterministic and context-aware; close ephemeral and retained failure handling | Cancellation and failures can leak partial or nonterminal state | WS-04, WS-05 | complete |
| G7 | Retained lifecycle | Serialize first-create/idempotency races; atomically publish/fail/invalidate/prune; enforce expiry at reads | Raw database conflicts and stale bindings remain possible | WS-05 | complete |
| G8 | Queries and cursors | Normalize query validation, inject a stable cursor codec, return exact DTOs, and use graph-order BFS | Cursors and traversal are restart- or ordering-unstable | WS-06 | complete |
| G9 | Consumer adapters | Preserve Network Flow errors and Reporting state/digest/expiry rules through the narrowed facade | Consumers erase distinctions or bind ineligible evidence | WS-07 | complete |
| G10 | Conformance and handoff | Audit and promote all 69 criteria only with sentence-complete executable selectors | Completion claims remain unsupported | WS-08, WS-09 | complete |

## 4. Interface and Migration Decisions

- Keep `Service.ProjectEphemeral` as the Network Flow seam.
- Introduce distinct lifecycle and query error types with closed constructors and serializers.
- Return exact projection-run inspection and graph-view list result DTOs; do not expose stored `ProjectionRun` internals.
- Replace random or panicking PostgreSQL-store constructors with one explicit options constructor requiring database, clock, and cursor key material.
- Keep Graph Projection authorization caller-owned and add no public transport surface or deployment configuration key.
- Keep the existing projection schema ID and graph digest identity because the graph wire shape is not being redesigned.
- Use existing database lifecycle columns, indexes, and transaction facilities; no migration is planned.
- Keep all GP-FIX identifiers and paths stable while replacing their semantic contents and hashes.
- Do not retain aliases for incorrect internal errors, constructors, or DTOs after repository callers migrate.

## 5. Ordered Workstreams

| Workstream | Depends on | Deliverable | Exit gate | Status |
| --- | --- | --- | --- | --- |
| WS-00 | none | Current controlling tracker | Markdown, JSON-shape, and generated-artifact policy checks | complete |
| WS-01 | WS-00 | Closed normative and derived contracts | Owner tables agree; documentation and JSON shape pass | complete |
| WS-02 | WS-01 | Manifest-driven fixture executor and conformance validation | Corrupt semantic goldens fail; fixture shapes and harness contract pass | complete |
| WS-03 | WS-02 | Narrow facade, exact errors/DTOs, cohesive internal files | Export/import guards and backend unit pass | complete |
| WS-04 | WS-03 | Admission, validation, derivation, cancellation, and failure remediation | Engine/admission fixture families and backend unit pass | complete |
| WS-05 | WS-04 | Retained lifecycle, persistence, and expiry remediation | Lifecycle fixtures, backend store, and migration drift pass | complete |
| WS-06 | WS-05 | Query, cursor, inspection, list, and traversal remediation | Query fixtures plus backend unit/store pass | complete |
| WS-07 | WS-04 through WS-06 | Reporting and Network Flow adapter migration | Phase 11 and Phase 12 local/service-backed slices pass | complete |
| WS-08 | WS-07 | Criterion-level evidence and 69-row promotion | All selectors resolve and execute; 69/69 implemented | complete |
| WS-09 | WS-08 | Final validation and handoff | Finalize, full check, tracker-sensitive checks, clean intentional diff | complete |

## 6. Workstream Completion Log

### WS-00 — Rebase the controlling tracker

Status: complete.

Substantive edits:

- removed obsolete ten-file inventory and superseded completion/generation sections from the active tracker;
- recorded the current 29-file module shape, production consumers, fixture posture, and matrix posture;
- installed the gap ledger, interface decisions, ordered dependencies, and per-workstream exit gates.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make lint-markdown` | PASS | Target emitted no retained summary |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T144438Z-p44410` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260713T144440Z-p44743` |

Checkpoint transition: WS-01 was unblocked.

### WS-01 — Close normative and derived contracts

Status: complete.

Substantive edits:

- added the closed `invalid_operation_request` family for malformed outer operation envelopes and reserved lifecycle-semantic rejection for `invalid_operation`;
- added query `invalid_value`, closed lifecycle/query envelope ordering and redaction, and corrected `max_input_bytes` to pre-admission `whole_input_limit_exceeded`;
- made fixture comparison scope, exact-artifact mode, run-independent validation posture, retained-state effects, and nonempty acceptance IDs machine-required;
- updated all 36 manifests with explicit comparison and state-effect declarations without changing their artifact hashes.

Compatibility decision: obsolete internal request/query reasons will be removed rather than aliased because there is no public Graph transport or assembled retained caller.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make lint-markdown` | PASS | Target emitted no retained summary |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T144721Z-p47354` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260713T144722Z-p47693` |

Checkpoint transition: WS-02 was unblocked.

### WS-02 — Build executable fixture and conformance mechanics

Status: complete.

Substantive edits:

- added a business-rule-free fixture verifier that loads ordered manifest steps, supplies declared inputs to a real-operation adapter, compares JSON or byte artifacts, and enforces the declared retained-state effect;
- added a mutation test proving that a semantic expected-value change fails verification;
- strengthened JSON-shape validation so every matrix selector resolves, every manifest test symbol exists, and fixture-to-criterion attachments agree in both directions;
- refreshed Make-owned phase schedules after the harness owner changed and aligned all manifest acceptance IDs with the matrix.

Evidence posture at this checkpoint: the executor foundation was complete. The then-remaining label-only input and expected artifacts were non-claiming until replaced by operation-derived authored artifacts in WS-04 through WS-06.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260713T145305Z-p55750` |
| `make backend-unit` | PASS, 231 tests | `.cartulary/test-results/20260713T145308Z-p57912` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T145319Z-p59721` |
| `make harness-contract` | PASS | Target emitted no retained summary |
| `make phase-schedule-drift` | PASS | `.cartulary/test-results/20260713T145331Z-p60785` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260713T145332Z-p60950` |

The first post-owner-change JSON-shape run failed as expected because generated phase schedules were stale; `make phase-schedules` refreshed them at `.cartulary/test-results/20260713T145241Z-p54624`. A second run then exposed and drove correction of the pre-existing GP-FIX-022/GP-AC-050 attachment mismatch.

Checkpoint transition: WS-03 was unblocked.

### WS-03 — Establish the facade, error contracts, and internal structure

Status: complete. The mechanical split of the derivation engine remains coupled to its behavioral remediation and is carried into WS-04; no caller-facing facade work remains.

Substantive edits:

- replaced the shared operation error with distinct `LifecycleError` and `QueryError` types, derived retryability, closed ordered envelope serializers, and rejection of unregistered detail members;
- added exact `ProjectionRunInspection` and `ListGraphViewsResult` DTOs; service inspection no longer exposes requests, nonces, raw graph state, generated timestamps, or private digest inputs;
- replaced random/panicking PostgreSQL store constructors with one explicit options constructor requiring database, clock, and stable cursor key material;
- migrated all repository tests to stable injected cursor keys and added facade leakage, nullability, retryability, ordering, and closed-detail tests;
- removed obsolete `OperationError`, random cursor-key construction, and old store constructors without aliases.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260713T150443Z-p76696` |
| `make phase-schedules` | PASS | `.cartulary/test-results/20260713T150446Z-p78839` |
| `make backend-unit` | PASS, 234 tests | `.cartulary/test-results/20260713T150446Z-p79007` |
| `make backend-store` | PASS, 141 tests | `.cartulary/test-results/20260713T150509Z-p82774` |

An earlier backend-unit run passed its tests but failed artifact accounting because the renamed boundary symbol no longer matched the subsystem owner. The public target reported `.cartulary/test-results/20260713T145945Z-p65114`; restoring the stable selector and adding the two new facade tests to the owner map resolved it.

Checkpoint transition: WS-04 was unblocked.

### WS-04 — Remediate admission, validation, and projection derivation

Status: complete.

Substantive edits:

- replaced scattered validation construction with a closed registry that validates code, severity, target, detail members, and §9.6.2 reason values, while deriving issue identity from required details only;
- removed predictable nonce fallback behavior, required service-owned nonces at retained and ephemeral admission, and mapped pre-admission identity and normalization failures to registered lifecycle reasons;
- separated validation, direct mapping, aggregation/property derivation, output resources, and orchestration into cohesive files within the single module package;
- passed context into direct and aggregate derivation loops, added deterministic cancellation boundaries, and added post-derivation schema/canonical validation with `output_schema_violation` terminal handling;
- converted every unit-layer GP-FIX input and expected artifact from labels or helper assertions to explicit manifest operations and reviewed semantic goldens, including full ephemeral resources and exact error envelopes;
- migrated unit fixture symbols to the common executor so fixture behavior is data-driven; the diagnostic candidate path writes only under test-results and cannot update reviewed artifacts.

Sequencing decision: GP-FIX-004 is a retained failed-run/publication fixture despite its earlier inclusion in the WS-04 list. Its exact durable snapshot requires the PostgreSQL transaction executor, so it remains with the other retained fixtures in WS-05 rather than being weakened into ephemeral-only evidence.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260713T153953Z-p150397` |
| `make backend-unit` | PASS, 236 tests | `.cartulary/test-results/20260713T153955Z-p152553` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T154057Z-p159391` |
| `make harness-contract` | PASS | Target emitted no retained summary |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260713T154109Z-p160347` |

The first tracker-sensitive JSON-shape run failed because the fixture schema's operation enum did not yet include the newly explicit unit operations. Expanding the authored schema to the closed operation set resolved the related failure; no product test failed.

Acceptance posture remains unchanged: all 69 rows are still `planned`. WS-04 establishes executable fixture evidence but promotion remains gated on the sentence-level WS-08 audit.

Checkpoint transition: WS-05 was unblocked.

### WS-05 — Harden retained lifecycle and persistence

Status: complete.

Substantive edits:

- closed the remaining lifecycle/query error wrapping at service boundaries so repository sentinels are converted into `LifecycleError` or `QueryError` envelopes;
- corrected malformed outer `idempotency_key` specification behavior to `invalid_operation_request` and aligned unit tests with the closed taxonomy;
- added row-count checks to retained publication/failure transitions and read-time expiry before projection-run, graph, vertex, edge, and traversal reads;
- tightened invalidation admissibility for non-consumable graph-view states and projection-run targets;
- filtered Reporting transaction-scoped binding lookup to reject expired retained runs;
- removed stale non-context projection and fixture helper code after staticcheck surfaced the obsolete paths.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260713T163345Z-p496469` |
| `make backend-unit` | PASS, 236 tests | `.cartulary/test-results/20260713T163412Z-p506905` |
| `make backend-store` | PASS, 141 tests | `.cartulary/test-results/20260713T163014Z-p363905` |
| `make migration-drift` | PASS | `.cartulary/test-results/20260713T162620Z-p226939` |

### WS-06 — Close query, cursor, traversal, and store fixture evidence

Status: complete.

Substantive edits:

- moved list-query shape digest ownership into the Graph Projection service while leaving visibility-scope digest caller-owned;
- returned fixed-precision timestamp strings in graph-view list DTOs and preserved nullable cursor behavior;
- added scalar validation for traversal kind filters and preserved graph-output-order BFS traversal;
- replaced GP-FIX-004, GP-FIX-018 through GP-FIX-021, GP-FIX-029 through GP-FIX-033, and GP-FIX-035 label artifacts with semantic retained-state goldens;
- added the backend-store fixture executor so mutating expected retained state fails exact fixture verification.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T162611Z-p226515` |
| `make harness-contract` | PASS | Target emitted no retained summary |
| `make backend-store` | PASS, 141 tests | `.cartulary/test-results/20260713T162817Z-p287965` |

### WS-07 — Migrate Reporting and Network Flow consumers

Status: complete.

Substantive edits:

- Reporting binding lookup now rejects expired projection runs in the caller transaction before returning graph projection refs;
- Network Flow maps invalid adapter inputs and Graph Projection fatal validation to `adapter_contract_rejected`;
- Network Flow maps Graph Projection computation failure, query unavailability, and unknown pre-outcome projection errors to `projection_unavailable`;
- added focused Network Flow unit coverage for the consumer error classifier.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make phase-slice PHASE=phase11` | PASS, 55 tests | `.cartulary/test-results/20260713T162637Z-p230812` |
| `make service-backed-slice PHASE=phase11` | PASS, 33 tests | `.cartulary/test-results/20260713T162705Z-p246196` |
| `make phase-slice PHASE=phase12` | PASS, 134 tests | `.cartulary/test-results/20260713T162731Z-p260125` |
| `make service-backed-slice PHASE=phase12` | PASS, 37 tests | `.cartulary/test-results/20260713T162752Z-p274250` |

### WS-08 — Promote criterion evidence

Status: complete.

Substantive edits:

- promoted all 69 Graph Projection acceptance criteria from `planned` to `implemented` after unit, store, consumer, shape, and harness gates passed;
- confirmed the final matrix contains 69 `implemented` criteria and zero `planned` or `deferred` criteria;
- retained the no-Core-05-publication posture because the evidence is implementation conformance evidence only.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T163738Z-p592238` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260713T163738Z-p592281` |
| `make lint-markdown` | PASS | Target emitted no retained summary |

### WS-09 — Final validation and handoff

Status: complete.

Substantive edits:

- updated this tracker with final workstream statuses, result roots, and acceptance posture;
- ran finalization without retained-run reuse because no prior successful full-run `RESULTS_DIR` was supplied;
- ran the broad repository check after staticcheck cleanup of obsolete helper code.

Validation:

| Target | Result | Result root |
| --- | --- | --- |
| `make lint` | PASS | `.cartulary/test-results/20260713T163345Z-p496494` |
| `make agent-finalize` | PASS | `.cartulary/test-results/20260713T162843Z-p292771` |
| `make lint-markdown` | PASS | Target emitted no retained summary |
| `make check` | PASS, 1116 tests, 139/139 work units | `.cartulary/test-results/20260713T163435Z-p509687` |

## 7. Required Verification and Handoff Record

Each workstream entry must record:

1. the changed owner and implementation surfaces;
2. substantive decisions and any deviations from this ledger;
3. public Make targets run, verdicts, and retained result roots when emitted;
4. failures and whether they were related to the workstream;
5. acceptance criteria or fixtures completed without overstating claim status;
6. the next workstream unblocked by the checkpoint.

The final handoff must report the planning summary, inspected and changed files, substantive edits, verification commands and results, skipped checks with reasons, final GP-AC totals, and any remaining risk. Harness readiness and retained test artifacts must not be represented as Core 05 claim-publication evidence.
