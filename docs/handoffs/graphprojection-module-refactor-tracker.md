# graphprojection Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

**GP-15 rollback follow-up complete; ready to commit.** E-156–E-159 record the owner clarification, failing witnesses, shared rollback repair, 43 passed rows, eleven-row Graph audit and final handoff gates. T-041 closes last on the final-text Markdown receipt. S-026 and S-027 remain closed, and the replay-authorization repair remains complete. E-151–E-155 remain historical evidence for their recorded source and cases.

E-122–E-126 retain the [GP-11 qualification](#gp-11-result-and-input-admission). E-117–E-121 retain GP-10.

S-022 through S-025 remain closed. E-082–E-092 retain their original results.
[Repair checkpoints](#gp-09-repair-checkpoints) · [Execution checkpoints](#gp-09-execution-checkpoints) · [Completion criteria](#gp-09-completion-criteria)

Planning began on a clean `main` at `9203107cd76ba2849bf02dc466309b6697be3e93`. Its Git tree, `7e206ded2ca9fe6b7c4e92d2088fedf28193ba8f`, matches the retained [GP-08 candidate receipt](../../.cartulary/evidence-manifests/graph-projection-gp08/candidate-tree.txt). GP-08's F-022–F-028 and §10 qualification apply to that commit. GP-09 planning is complete for the selected seams. E-083–E-088 record execution of the six slices.

### GP-08 historical baseline

The seven findings reopened GP-07 acceptance and were reproduced, repaired and requalified on the source recorded in §10. GP-07 evidence E-070–E-075 applies to its reviewed source. GP-08 began at `2dc1519246657929dc7194de9db22987597f9a92` with 34 staged paths (29 modified, five added), no unstaged changes, and staged patch SHA-256 `b32072c54ca6a2aaba55bed059dcae59775d3426d408fb58f404511eccf0fb2e`.

### GP-08 remediation and acceptance gates

| Finding | Graph owner | Compatibility and required witness |
| --- | --- | --- |
| F-022 | §§3.6, 5 | Independent forward/reverse field sets, wildcard, required reverse-only rejection |
| F-023 | §§6.3, 8–9 | Valid linked reverse publication, exact reads, atomic reconstruction and invalid family/link rejection |
| F-024 | §§4–5, 8 | Historical numeric envelope, binding and child bytes exact on read and reprojection; numeric grouping equivalence |
| F-025 | §7 | One bounded issue budget, ordinary reporting at capacity, resource error at first excess across rules |
| F-026 | §§3.6, 5 | One endpoint pair per group, both policies, permutations, missing and invalid later contributors |
| F-027 | §§5, 7 | Terminal callback/parent/deadline cancellation at public entry, no retry or partial result |
| F-028 | §§3.6, 7–8 | Bounded list/set in properties and metadata; exact-limit completion and first-excess rejection |

The user reviewed and confirmed the GP-08 owner wording after the Phase 1 edits. The accepted design preserves v2 storage bytes and rejects endpoint conflicts. GP-08 required reviewed owner wording, red witnesses on GP-07, focused fixes, then finalization, frozen source, complete Graph audit, affected peer/static checks and final-text Markdown lint. An existing selected binding must reproduce exactly or reconstruction rolls back. An affected generic declaration requires source-owner rematerialization before rollout. Upgrade every reader before enabling reverse publication. After publication, rollback requires a reader that supports reverse edges.

- Repository root: `/home/jochi/code/cartulary`.
- Supplied and canonical repository-relative target: `internal/modules/graphprojection`; existing directory, resolved inside the repository.
- Target label: `graphprojection`, derived from the final path component and normalized to lowercase kebab case.
- Output: `/home/jochi/code/cartulary/docs/handoffs/graphprojection-module-refactor-tracker.md`.
- Artifact mode: `normal`. GP-05 implements the accepted Graph/direct-consumer scope, not exhaustive behavioral certification.
- GP-05 authorized writes: the scoped implementation, tests, specification correction, authored routing, this tracker and normal managed Make artifacts. Generated outputs are refreshed only through Make.
- Historical GP-05 authorization covered S-006–S-010. The original GP-09 planning scope was tracker-only. E-082 records the subsequent implementation authorization.
- Non-goals: redesigning module topology or valid Graph semantics, reintroducing leases or compatibility paths, migration/reset, deployment, UI/browser work, performance qualification, full semantic audit and release certification. Newly discovered issues outside the selected seams require separate findings and scope decisions.
- Framework: [Cartulary Modular Refactor Planning Framework](cartulary_modular_refactor_planning_framework.md), read before the initial analysis and re-read before this write. It supplies doctrine, not repository facts.

### GP-07 authorized remediation baseline and checkpoints

E-060: on 2026-11-03, HEAD remains `2dc1519246657929dc7194de9db22987597f9a92` on `main`. The 28 staged paths, empty unstaged diff and 8,064-path content/mode inventory are retained at `/tmp/cartulary-graph-gp07-0n05g89d`, with binary staged/unstaged patches, original changed files and index. Starting index SHA-256: `3c1798136b7992dd197f62306b172533831eaf7a15c0a40da975d5a07d416ea3`. Root/backend/tooling instructions and test-operations guidance remain applicable. The accepted implementation request supersedes the planning-only restriction.

The authoritative chain is WF-24/S-016 → WF-25/S-017 → WF-26/S-018 → WF-27/S-019 → WF-28/S-020 → WF-29/S-021. Revalidate source and prerequisites, enter `IN_PROGRESS`, record the pre-fix reproduction, changes, exact commands, native exits, artifacts, compatibility and risks, then complete the tracker checkpoint before starting the successor. No failed or missing gate is silently closed. No product test has run in GP-07 at entry.

Definitions must support every configured producer input context, independent of emitted data; aggregates use immediate provenance for contributor ordering. Graph v2 public shapes/framing and restore/result v5, registry v5, binding v7 remain current. Corrected generic inputs may reject or change identities, while unaffected Graph fixtures and both NF production bindings remain exact. Never rewrite stored bindings or introduce legacy execution modes. GP-06 passes do not cover these findings.

### GP-06 authorized follow-up baseline

On 2026-11-03, `main` remains at `2dc1519246657929dc7194de9db22987597f9a92`. All 20 existing modified files are preserved as the starting diff, with the 8,061-path baseline, original index and file copies at `/tmp/cartulary-graph-gp06-nsfrn4_z`. The user's explicit S-011–S-015 implementation request supersedes the prior planning-only scope. No commit, deployment, reset, compatibility mode or migration is authorized or needed.

The controlling sequence is S-011 → S-012 → S-013 → S-014 → S-015. Revalidate prerequisites, enter `IN_PROGRESS`, and record completion and evidence before entering each successor. Graph §§3-J/4/5 govern the fixes. Bare `kind` lookup is removed without projected-kind replacements; unrelated discriminators and property-key reservations remain unchanged. Corrections may change identities of previously mishandled generic inputs; unaffected Graph goldens and current Network Flow configurations must remain exact. Historical GP-05 passes establish their recorded assertions, not closure of these new findings.

### GP-05 execution baseline and checkpoints

On 2026-11-03, implementation starts on `main` at `2dc1519246657929dc7194de9db22987597f9a92`. This tracker is the sole pre-existing modified file; the index is unchanged from GP-04. The 8,061-path content/identity baseline, original tracker and index are retained outside the repository at `/tmp/cartulary-graph-gp05-6j7yob3s`. Tracker SHA-256 is `8d54f43b4e150e8ff8c4d346dd0d34bce36120ba30df14e5a2f6d97cb15700ba`; index SHA-256 is `3487f6be7046d9c42a579585b1a1341b309c9bc41d086db82921b8bac99648cc`.

The accepted plan in §7 governed this execution: WF-14/S-006 → WF-15/S-007 → WF-16/S-008 → WF-17/S-009 → WF-18/S-010. Owner requirements were reconciled before each change; each slice entered `IN_PROGRESS`, then this tracker recorded checks, artifacts, risks and completion **before its successor started**. Completed GP-03 history and the GP-04 planning changes were preserved. Graph §§3-J/4 already require duplicate-safe acyclic admission; §9 already requires exact current definitions. Reporting §15.2's obsolete lifetime wording was corrected before S-009 code changes under REQ-RPT-099g. No owner amendment, migration, compatibility mode, reset or deployment was required.

E-030: root/backend/tooling instructions, test-operations guidance, affected source and adopted Graph clauses revalidated. Source baseline matches the accepted plan. S-006 enters characterization; no runtime reproduction or implementation pass is yet claimed. Exact validation commands remain in §8.

### GP-04 planning baseline and execution contract

GP-04 starts on clean `main` at `2dc1519246657929dc7194de9db22987597f9a92`, which commits the eleven GP-03 changed/new files. The tracked tracker resolves to the declared target, is a regular single-link file, and has no symlinked root/docs/handoffs ancestors. Initial tracker SHA-256: `ba6868ec056d36ff351620b4af712250557b56c6a2bff57ad5baac175e8ea87d`; index SHA-256: `3487f6be7046d9c42a579585b1a1341b309c9bc41d086db82921b8bac99648cc`. There are 8,061 tracked/non-ignored paths and no pre-existing worktree changes. The per-path content/identity baseline, index metadata and original tracker are retained at `/tmp/cartulary-graph-gp04-jug0qsnf` (E-022).

The user selected **module and direct consumers** rather than full subsystem qualification or cleanup alone. Read-only planning revalidated the relevant source, adopted owners and exact selectors at this commit. Findings F-011/F-012 are inspection-based correctness exposures; no new reproductions or product checks have run. GP-03 results remain historical and were not re-audited as current acceptance evidence.

After separate implementation authorization, execute WF-14 → WF-15 → WF-16 → WF-17 → WF-18, corresponding to S-006 → S-007 → S-008 → S-009 → S-010. Revalidate affected evidence before starting a slice, mark it `IN_PROGRESS`, and record changes, exact checks, artifacts, risks and status **before starting its successor**. Failed exit criteria remain explicit blockers. Every future slice is `TODO` in GP-04; writing this plan completes none of them. The authoritative prerequisites are in §6.

### GP-03 historical execution baseline and checkpoint rule

Starting baseline: `main` at `f8ae96be698fb7f78c3f403e62cf5099752e6dfe`; index and tracked worktree clean; this tracker was the sole pre-existing untracked file, SHA-256 `7dadce0a7c5b74e390f646b1a83e5dab25ac3f5ab0574a79b3274729c192fd2a`. Content/identity baseline for 8,060 paths and the original tracker are retained outside the repository at `/tmp/cartulary-graph-remediation-j9cs9xwz`. No prior passing product evidence is claimed.

GP-03 executed WF-09 → WF-10 → WF-11 → WF-12 → WF-13. Its checkpoint rule required evidence revalidation, `IN_PROGRESS` on entry, and results recorded before the successor. Normal managed Make artifacts and setup were authorized for that implementation. The GP-02 tracker-only lint restriction was superseded, not retroactively erased; GP-04 likewise explicitly permits the repository-required documentation lint artifacts.

### Authority and inspected owner sources

Apply adopted subsystem NLSpecs only within their named scope; Core 00–04 own their allocated implementation-conformance behavior. Core 05 applies only to claim-bearing timed or fixture-sensitive publication and is not invoked here. Domain vocabulary and implementation guides support navigation; code/tests describe current implementation; prior plans and the framework are evidence only. Typed projections and test routing do not establish requirements or specification completeness.

| Evidence | Inspected owner or supporting source | Applicable authority and clauses | Posture |
| --- | --- | --- | --- |
| E-001a | [Core 00](../spec/00_document_set_status_and_precedence.md) | REQ-00-062, REQ-00-076; coordinated adoption and later frozen-input amendment | Adopted owner allocation; architecture decision subordinate to behavioral owners. Earlier lease wording is not current authority after explicit supersession. |
| E-001b | [Graph Projection NLSpec](../graph_projection_nlspec.md) | Adopted/current 2.2.1, §§1, 3–10; especially §§8–9 | Pure v2 engine, completed results, borrowed persistence capabilities, no live Graph leases, reconstruction v5, registry v5 and binding v7. |
| E-001c | [Graph Projection boundary decision](../decisions/graphprojection-module-boundary.md) | Status, Decision, Reconstruction ownership, Ports and exports, Compatibility, Verification | Adopted under REQ-00-076; positive topology and narrow construction, not a separate domain bounded context. |
| E-001d | [Network Flow Activity NLSpec](../network-flow-activity-nlspec.md) | Adopted/current 8.0.2; §1, §19.1 NF-REQ-170a–b | Source declarations, graph-view identity, public graph/saved-graph resources, materialization and authorization; public major 8, state/minimum 5. |
| E-001e | [Reporting Subsystem NLSpec](../reporting-subsystem-nlspec.md) | Adopted/current 1.4.1; §1, §15.1 REQ-RPT-076–078, REQ-RPT-099g | Exact source-owner admission and privately retained immutable input; subsequent execution independent of live Graph availability. |
| E-001f | [Core 01](../spec/01_architecture_storage_and_view_contracts.md) | REQ-01-424, REQ-01-624–625A | Structural Recovery versus post-claim extension reconstruction; workbook rebuilder is a separate interface. |
| E-001g | [Core 02](../spec/02_domain_model_schema_and_history.md), [Core 03](../spec/03_workbook_interaction_collaboration_and_workflows.md), [Core 04](../spec/04_security_deployment_and_conformance.md) | Core 02 §§1–2; Core 03 §§1–2; Core 04 saved-graph security restatement following REQ-04-130 | Inspected for source/projection, workbook and authorization boundaries. Current Core 04 AC-568 already requires three-table reconstruction and no live Graph leases. The GP-02 suggestion that current Core 04 retained major-6/lease wording was unsupported and is superseded by GP-03 inspection. |
| E-002 | [Domain vocabulary](../domain.md); [NLSpec research](../research/nlspec-spec.md); framework above | Domain owner/context tables and graph/declaration/job/diagram distinctions; research introduction and five specification properties | Supporting vocabulary/doctrine. Neither research nor framework overrides adopted owners. Presentation design is not being changed. |

Root `AGENTS.md`, `internal/AGENTS.md`, and, for inspecting test-support/routing ownership, `tools/AGENTS.md` were read. No applicable nested `AGENTS.md` was found under the target or `docs/handoffs`. The refactor-tracker skill and its tracker-format reference, and the repository test-operations skill/command recipes, supply the planning procedure; the requested table columns are retained here.

### GP-01/GP-02 baseline and inspection limits

Session GP-01 performed read-only planning. Session GP-02 was authorized to create this tracker. Both observed `main` at `f8ae96be698fb7f78c3f403e62cf5099752e6dfe`, with no staged, unstaged or non-ignored untracked changes before this document. The absent tracker was rechecked; there was no on-disk handoff history to replace.

The universe is tracked plus non-ignored untracked files under the literal target. It contains **36 tracked regular authored Go files: 22 production and 14 test files, in five packages**. No target symlink, submodule, generated/vendor file, ignored file or untracked file was found. During GP-01/GP-02 all 36 were directly inspected at inventory depth: imports, declarations, relevant function bodies and test assertions. Selected entrypoints, adapters and boundary tests were read more deeply. This is not a claim that every line or every consumer behavior was exhaustively audited.

Repository files inspected are enumerated in §2 and the E-003–E-010 ledger in §10. Searches expanded only to direct callers, owner contracts, test support, generated projections and command definitions. Frontend trees and complete adjacent modules were not audited. Candidate files found by search alone are identified as such in §10; their assertions are not claimed as inspected.

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits or blocker IDs |
| --- | --- | --- | --- |
| Branch and commit | `main`; `f8ae96be698fb7f78c3f403e62cf5099752e6dfe` | Unchanged in final comparison, E-012. | Revalidate on the next session. |
| Index | Clean; content SHA-256 `08442dc20d4fe650a9e48b24829c76adb75ab25d4e0a061e29ad5f6e44512292`; file identity captured in session memory | Content and identity unchanged, E-012; no staging. | Git optional writes disabled for baseline/final inspection. |
| Tracked/non-ignored working-tree content and identity | 8,059 entries excluding the output; aggregate SHA-256 `2b0f928d3f16235381bf224d93c9fcd8a600d88886d7defdf80c4b87b52fa04a` | Count and aggregate unchanged, E-012; Git reports only the new tracker. | Digest includes path, file identity/mode/size/mtime and content SHA-256 or symlink target; no file contents are copied into this tracker. |
| Target content and identity | 36 entries; aggregate SHA-256 `241f63f32a584902b2153b0606b2581f7ec10acfa47aafa310795218ca9187ae` | Count and aggregate unchanged, E-012. | Boundary inventory, not executable conformance evidence. |
| Relevant ignored outputs | Static-analysis cache: 1,069 entries, SHA-256 `8e51741cdd8da6bf6d0740a4ce179fab3a5c492fdeace3a96f612991e288ea91`; test-results root identity/mtime captured | Cache count/aggregate and test-results root identity/mtime unchanged, E-012. | Test-result contents and unrelated ignored caches are not exhaustively fingerprinted. No test or artifact-producing command ran. |
| Tracker and parent | Output absent; existing repository/docs/handoffs directories resolved without symlinks; parent identity captured | Created a regular, single-link tracker at the admitted path; parent remains the same directory. | Parent directory mtime changed from adding the tracker. Output absence and ancestor identity were rechecked immediately before writing. |

## 2. Current-State Repository Inventory

### GP-09 inventory at `9203107cd`

The target contains **45 tracked files** across the same five Go packages, including the historical JSON fixture in PostgreSQL test data. The inventory comprises 24 production Go files, 20 Go test files and the fixture. All are regular authored files. Enumeration included non-ignored untracked files and checked for ignored entries under the target. Neither search added a path. Names below are relative to `internal/modules/graphprojection`.

| Inspection in GP-09 | Files | What the evidence covers |
| --- | --- | --- |
| Direct, encoding and public result construction (8) | `canonical.go`, `objects.go`, `input.go`, `engine_v2.go`, `project.go`, `resources.go`, `types.go`, `result_v2_ports.go` | Input builders through normalized digests, both result-content builders, envelope and child row construction. |
| Direct, bounds and provenance (7) | `validation.go`, `internal/semanticlimits/registry.go`, `execution_state.go`, `direct_mapping.go`, `aggregation.go`, `schema.go`, `result_v2_validation.go` | Order of collection guards, emission, production of each object family and admission of existing content. |
| Direct, consumers within Graph (3) | `postgresresult/store.go`, `restore/service.go`, `postgresrestore/writer.go` | Validation before insert and on content reads, metadata-only locking, reconstruction and binding comparison. |
| Direct, selected assertions and fixture (11) | `engine_v2_test.go`, `aggregation_v2_test.go`, `semantics_v2_test.go`, `execution_bounds_v2_test.go`, `edge_semantics_v2_test.go`, `result_v2_validation_test.go`, `v2_contract_projection_test.go`, `postgresresult/store_test.go`, `postgresresult/result_proof_test.go`, `postgresresult/historical_numeric_test.go`, `postgresresult/testdata/direct-numeric-v2.json` | Assertions, their parent tests and the pre-GP-07 numeric fixture. Inspection covers the indicated paths through these files. |
| Prior inventory carried forward (16) | `field_paths.go`, `limits.go`, `semantics_v2.go`, `validation_registry.go`, `restore/contract.go`, `restore/recovery_state.go`, `boundary_guard_test.go`, `contributor_order_v2_test.go`, `field_paths_v2_test.go`, `graph_projection_migration_test.go`, `postgresrestore/writer_test.go`, `postgresresult/cleanup_test.go`, `postgresresult/contract_test.go`, `restore/boundary_test.go`, `restore/contract_test.go`, `restore/service_test.go` | Membership and identity checked against the committed tree. Earlier responsibility and test descriptions below supply context. Reinspect any of these files before extending a slice into it. |

The completed GP-09 target adds three authored files to this 45-file baseline:
`input_counts_v2_test.go` (admission precedence and resource boundaries),
`result_vectors_v2_test.go` (fixed-byte producer and pure proof assertions)
and `postgresresult/testdata/result-vectors-v2.json` (six captured bindings,
envelopes and ordered child-row sets shared with the database tests).
The final inventory has 48 files in the existing five Go packages.

Outside the target, the provenance trace reaches the authored v2 schema, Network Flow saved-graph and Reporting admission reads, and the reconstruction writer. The owner and row manifests supply validation routing (E-079/E-080). The inventory and earlier inspections below retain their original baselines.

The 36-file inventory remains the GP-01/GP-02 baseline, with GP-03 affected entries updated below. GP-04 re-enumerated the same 36 tracked authored Go files and re-inspected the affected seams identified in the additional table below; other entries retain their explicitly historical inspection depth. All paths in the first table are relative to `internal/modules/graphprojection`. `GP` means that root package, `NF` means `internal/modules/networkflow`, and `GA` means `internal/app/graphassembly`. `STD` means Go standard-library imports. E-003–E-010 resolve the earlier inspected files. Risk estimates describe the cost of changing the file, not a current passing test or exhaustive audit.

Test shorthand: **engine** = `engine_v2_test.go`; **semantics** = `semantics_v2_test.go`; **proof** = `result_v2_validation_test.go` through the engine parent test; **contract** = `v2_contract_projection_test.go`. Private files have package-local callers, rather than an invented external facade.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `aggregation.go` | Deterministic contributor grouping, aggregate vertices/edges, property and metadata derivation | Private `emitAggregations`, candidate/group/merge helpers | `project.go`, direct mapping | STD; private GP types/semantics/canonical helpers | engine, semantics | v2 semantic matrix and non-empty golden, indirectly | Graph Projection | high | E-003; no source-owner querying or mutation. |
| `boundary_guard_test.go` | Positive root import, package-topology, exported-function and exported-value guards | Four `TestGraphProjection*Allowlist/Boundary_Unit` tests, including the S-009 exported-value guard | Go test catalog | STD AST/parser/filesystem | The three tests themselves | Machine test-row mapping only | Graph Projection tests | medium | E-003/E-010; reads source, not Markdown. Tests are downstream evidence. |
| `canonical.go` | Canonical encoding, framed tuples, hashes and object IDs | Private canonical/tuple/hash helpers | GP engine, objects, projection, validation | STD | engine, semantics, proof | Exact fixture bytes/digests, indirectly | Graph Projection | high | E-003; identity helpers remain private. |
| `direct_mapping.go` | Source-to-vertex/edge mapping, predicates and direction handling | Private `emitDirectVertices`, `emitDirectEdges` | `projectSemanticGraph` | STD context; GP validation, properties and identity | engine, semantics | Semantic input/result v2, indirectly | Graph Projection | high | E-003; source entities are graph inputs, not the Entities module. |
| `engine_v2.go` | Closed JSON admission, trusted invocation, normalization, execution, result identity/resource conversion, cancellation | `ProjectV2`, `InvocationContextV2`, `ProjectionResultV2`, `ProjectionErrorV2`; result/binding/resource methods | NF adapter; `restore`; test fixture | STD; private GP implementation | engine, semantics, proof, root boundary | `graph_projection.v2`, result v2, golden fixtures | Graph Projection | high | E-003; sole engine entrypoint; no constructor, HTTP, SQL or jobs. |
| `engine_v2_test.go` | Golden determinism, rejected members, trusted identity separation, ordering, cancellation and maximum bounds | Six `TestProjectV2*` tests; private fixture helpers | Engine rows in E-010 | STD; generated `contractgraphprojection` | Invokes proof helper as a subtest | Generated fixture index | Graph Projection tests | high | E-003; tests inspected, not run. |
| `graph_projection_migration_test.go` | Current-head table membership, no declaration-to-result FK, cleanup index | `TestGraphProjectionV2HeadSchemaContract_Integration` | Graph storage row | STD; platform postgres; `testutil/pgtest` | Its integration assertions | Current migrated schema; no SQL edits | Graph Projection storage tests | high | E-004; count four includes the NF declaration plus three Graph tables; it is not four Graph-derived tables. |
| `input.go` | Duplicate-member rejection, typed input parsing, defaults, scalar and collection normalization | Private parser/scalar helpers | `engine_v2.go`, mapping/validation | STD | engine, semantics | Input v2, indirectly | Graph Projection | high | E-003; time is parsed data, not a clock capability. |
| `internal/semanticlimits/registry.go` | One private executable semantic/traversal limit registry | `Registry`, `CurrentV2`, maximum result constants within Go internal boundary | GP `limits.go`, `result_v2_ports.go`; `postgresresult` | None | contract; engine bounds; storage traversal | Authored semantic registry values via tests | Graph Projection private implementation | high | E-003; shared semantic owner, not a generic utility package. |
| `limits.go` | Root-local copy of current semantic limits | Private `graphProjectionLimits` | GP parser, engine, projection and validation | Private semanticlimits | contract, maximum-bounds tests | Semantic registry parity | Graph Projection | high | E-003; no public mutable options. |
| `objects.go` | Ordered canonical input objects for mappings, definitions, sources and filters | Private object encoders | GP normalization and identity paths | GP canonical/types helpers | engine, semantics | Configuration/source digests, indirectly | Graph Projection | high | E-003; ordered encoding must remain owner-exact. |
| `postgresrestore/writer.go` | Clear, immutable publication, reconciliation, commit and committed-state proof | `Writer`, `New`, `Clear`, `ReplaceAll`, `DerivedStateReconciler` | GA composition; restore publisher port | STD; pgx; GP, `postgresresult`, `restore` | `writer_test.go`; GA integration | Three derived tables; restore result/binding | Graph Projection persistence adapter | high | E-005/E-008; borrowed transaction; mandatory reconciler; Clear comment corrected by S-002/E-016. |
| `postgresrestore/writer_test.go` | Construction, clear/publication rollback, reconciliation and indeterminate outcomes | Seven top-level tests including constructor test | `storage.restore_publication` row | STD; uuid/pgx; GP/restore; platform postgres; graphfixture/pgtest | Writer integration cases | Restore v5 and current tables | Graph Projection adapter tests | high | E-005; constructor test has `_Unit` suffix but uses PostgreSQL; routed with integration evidence. |
| `postgresresult/cleanup_test.go` | Exact-owner, one-candidate, skip-locked cleanup and rollback assertions | `TestResultV2CleanupBoundIsolationLockingAndRollback_Integration` | Cleanup-locking row | STD; pgx/pool; postgresresult; graphfixture/pgtest | Concurrent borrowed transactions | Storage-maintenance v2 | Graph Projection storage tests | high | E-004; separate from NF selected-declaration race evidence. |
| `postgresresult/contract_test.go` | Cleanup operation contract parity | `TestStorageMaintenanceProjectionMatchesRuntime_Unit` | Service-free contract row | STD; generated contractgraphprojection | Two operations, borrowed transaction, maximum one row | `storage-maintenance.v2.json` | Graph Projection storage tests | medium | E-004/E-009. |
| `postgresresult/result_proof_test.go` | Producer-backed tamper rejection on publication and every content-returning read | Private `checkCurrentResultProofs` and fixture helpers | Parent integration test in `store_test.go` | STD; GP/postgresresult; pgx | Parent subtest checks binding, canonical content and row fidelity | Completed-result v2 | Graph Projection storage tests | high | E-004; helper is not a separately selected test. |
| `postgresresult/store.go` | Immutable result publication/replay, exact reads, metadata locking, ordered objects, bounded traversal and cleanup capabilities | `Publisher`, `Reader`, `Cleaner`; fallible constructors and narrow methods | NF storage/admission/cleanup; postgresrestore | STD; pgx; GP; private semanticlimits | Store/proof/cleanup/contract tests; NF callers | Result tables; storage-maintenance v2 | Graph Projection persistence adapter | high | E-004; no peer declaration query, retention worker or hidden transaction ownership. |
| `postgresresult/store_test.go` | Producer-backed publication/read/traversal/cleanup integration | `TestResultV2PublicationReadTraversalAndCleanup_Integration` | Managed-service storage row | STD; pgx; GP/postgresresult; graphfixture/pgtest | Includes current-result proof helper | Completed-result v2 | Graph Projection storage tests | high | E-004; outgoing/incoming, depth-zero and over-bound assertions inspected. |
| `project.go` | Pure projection pipeline, projected graph validation, schema registry, ordering and filter helpers | Private `projectionWork`, `projectSemanticGraph`, supporting helpers | `ProjectV2` and GP internals | STD; private GP files | engine, semantics | Result schema registry and capabilities | Graph Projection | high | E-003; Graph schema registry is not workbook view-schema logic. |
| `resources.go` | Vertex/edge/schema/validation resource serialization | Private `vertexResource`, `edgeResource`, `schemaRegistryResource`, related helpers | Engine result/resource/completed-result paths | GP types | engine, semantics, proof | Result v2 JSON | Graph Projection | high | E-003; no HTTP envelope ownership. |
| `restore/boundary_test.go` | Positive restore imports/files/exports and recovery contribution guards | Three restore boundary/contribution tests | Deterministic current-restore row | STD AST/parser/filesystem; restore package | Current three-table v5 contribution | Generated Recovery bindings indirectly | Graph Projection restore tests | medium | E-005; platform canonicaljson/recoverystate dependencies are explicitly guarded. |
| `restore/contract.go` | Current source registry, closed request/result/error and implementation binding admission | `Restore*` models/ports/errors; current/frozen binding and registry functions; validation/readiness methods | Restore service; GA/server; NF source registration; Recovery | STD; uuid; generated contractrecovery; GP; platform canonicaljson | Restore contract/service/boundary tests | Rebuild/result v5, registry v5, binding v7 | Graph Projection restore | high | E-005; exact current definitions, not a compatibility dispatcher. |
| `restore/contract_test.go` | Exact registry/binding, strict definition bytes, closed results/errors | `TestGraphRestoreCurrentRegistryAndBindingsAreExact`, `TestGraphRestoreResultContractClosesStatesAndSafeErrors` | Current-restore row | STD; uuid; generated contractrecovery; restore | Duplicate/alias rejection and current-only definitions | Recovery fixtures/schemas and generation registry | Graph Projection restore tests | high | E-005/E-009; does not establish adoption. |
| `restore/recovery_state.go` | Pure three-table rebuildable-state contribution | `RecoveryStateContribution` | Application recovery catalog assembly | Platform recoverystate | Restore contribution boundary test | Frozen Recovery catalog contribution | Graph Projection restore | high | E-005; narrowly allocated platform representation, not database access. |
| `restore/service.go` | Validate current request, clear through port, enumerate/project candidates, enforce operation bounds, publish and validate proof | `RestoreService`, `NewRestoreService`, `Rebuild`; publisher/plan/proof/error types | GA; restore tests | STD; uuid; generated contractrecovery; GP | Restore service tests; GA integration | Exact current result and postcondition digests | Graph Projection restore | high | E-005; ordering is metadata admission → clear → source preflight → publication in caller transaction. S-003/S-004 closed F-005; S-007 adds metadata isolation and binding admission evidence. |
| `restore/service_test.go` | Current candidates, failure codes, bounds, repeatability, proof/readiness and constructor assertions | Nine restore service tests; recording publisher and fixtures | Current-restore row | STD; uuid; GP/restore | Shared clear/enumerate/publish event log; closed failures and cancellation | Current restore definitions | Graph Projection restore tests | high | E-017; phase assertions pass; database rollback is separately verified in S-004. |
| `result_v2_ports.go` | Immutable binding/completed-row/traversal/cleanup value models and errors | `ResultBindingV2.ValidateIdentity`, completed/result row types, traversal types, errors/constants | NF, Reporting, result/restore adapters; engine | STD; private semanticlimits | Engine/proof/contract; storage tests | Result v2 and source projection references | Graph Projection | high | E-003; value models do not create a database facade. Publication time is operational. |
| `result_v2_validation.go` | Binding, canonical envelope, row fidelity, order, counts and endpoint validation | `CompletedResultV2.Validate` | Postgres publication/content reads; engine consumers/tests | STD; GP canonical/admission helpers | proof; result proof integration | Completed-result v2 | Graph Projection | high | E-003/E-004; validation does not re-enter the source owner. |
| `result_v2_validation_test.go` | Canonical bytes, identity, tampered rows, structural invalidity and cancellation characterization | Private `checkCompletedResultProofsV2` | `TestProjectV2GoldenDeterminismAndCompletedResult_Unit` | STD; GP internals | Parent engine subtest | Current producer/model | Graph Projection tests | high | E-003; helper-to-parent linkage explicitly inspected. |
| `schema.go` | Private JSON member types, kind matching, input member paths and definition member rules | Private `jsonKind`, `memberSpec`, helpers | Input/result admission | STD | Engine rejection/proof/semantics tests | Closed Graph JSON schemas | Graph Projection | high | E-003; not saved views or generated workbook view contracts. |
| `semantics_v2.go` | Closed type/merge matrix, scalar/array normalization and exact merge operations | Private semantic helpers | GP validation, candidate evaluation, aggregation | STD incl. math/big | semantics; engine semantic golden | Semantic registry merge matrix | Graph Projection | high | E-003; stays behind the pure engine. |
| `semantics_v2_test.go` | Every admitted/rejected merge pair, golden non-empty result, scalar boundaries, null/default/wildcard behavior | Six `TestProjectV2*` semantic tests | Semantic-matrix row | STD; generated contractgraphprojection | Executes merge pairs and golden projection | Semantic registry and non-empty fixture | Graph Projection tests | high | E-003; fixture parity is downstream evidence. |
| `types.go` | Private semantic input/configuration plus public graph/schema/result value vocabulary | `Vertex`, `Edge`, schema/property/metadata/reference/validation/capability models | GP files; typed consumers of engine results | None | engine, semantics, proof | Result schema/value vocabulary | Graph Projection | high | E-003; source input types are private; no canonical entity lifecycle. |
| `v2_contract_projection_test.go` | Packaged contract identity, limits, closed shapes, fixtures and cleanup contract parity | `TestGraphProjectionV2ContractProjection_Unit`; decode helper | Contract-projection row; other GP tests use helper | STD; generated contractgraphprojection | Contract/runtime parity assertions | Graph index, schemas, semantic registry, fixtures, maintenance v2 | Graph Projection tests | high | E-003/E-009; no Markdown ingestion. |
| `validation.go` | Admitted-source/configuration/filter/mapping/aggregation validation and resource bounds | Private validation helpers | Projection admission/pipeline, mapping | STD; GP semantic/type/issue helpers | engine, semantics, maximum bounds | Input semantics and limits indirectly | Graph Projection | high | E-003; validates graph input, not caller authorization. |
| `validation_registry.go` | Closed internal issue vocabulary, safe details, deterministic summary and truncation | Private issue contracts and projectionWork methods | GP validation/derivation | STD; GP identity/canonical helpers | Engine safe errors; semantic failure cases | Public projection errors derived from internal summaries | Graph Projection | high | E-003; no telemetry adapter or source-owner diagnostics ownership. |

### GP-03 adjacent changed surfaces

| Path / kind | Inspection and responsibility | Owner / validation |
| --- | --- | --- |
| `internal/app/graphassembly/rebuild_test.go` / authored test | Directly inspected and extended: nonempty producer/publisher seed, actual transaction witness and exact committed-state snapshots | Graph assembly evidence; V-012 |
| `internal/modules/networkflow/graph_restore_failure_integration_test.go` / new authored test | Directly inspected: real materialization, quiescent binding mismatch and exact Graph/declaration/job/input rollback snapshot | Network Flow source-owner evidence; V-014 |
| `tools/test_families/module.graphprojection.json`, `module.networkflow.json` / authored routing | Exact renamed/added test selectors; no duplicate aliases | Existing owner rows; V-011 |
| `tools/execution_topology_render_index.json` / generated projection | Make regenerated input hashes/digest; no hand-edit | V-003 |

### GP-04 affected-source reinspection

The remaining unchanged inventory entries are carried forward from their earlier direct inspection, not claimed as freshly audited line by line. Repository-wide symbol searches establish the identified callers; they do not prove absence of every possible dynamic consumer. No generated, vendor or ignored target code was added to scope.

| Source / kind | GP-04 inspection and observation | Owner / finding / evidence |
| --- | --- | --- |
| `engine_v2.go`, `validation.go`, `aggregation.go` / authored production | Directly inspected normalization/order/validation flow: ID-keyed ordering can erase duplicates, and dependency stalls use a fallback before later validation | Graph semantic owner; F-011; E-023 |
| `canonical.go`, `input.go`, `project.go`, `types.go` and their callers / authored production | Directly inspected strict-only mode calls, unused work fields/parameters, overwritten initializers, required-version admission and duplicate sorts | Graph private engine; F-013; E-025 |
| `restore/contract.go`, `restore/service.go` / authored production | Directly inspected shallow nested-slice ownership, current constructors/accessors and generation-loop binding validation | Graph reconstruction; F-012; E-024 |
| `result_v2_ports.go`; NF `reporting_graph_source.go`; Reporting `graph_source.go`, `graphsourcecontract/contract.go` / authored production | Directly inspected both selection sentinels, all discovered production consumers, error mapping and existing destination contract | Source-owner/Reporting handoff; F-014; E-026 |
| Root and restore boundary tests, `engine_v2_test.go`, `semantics_v2_test.go`, `restore/contract_test.go`, Reporting `graph_projection_refs_test.go`, NF `reporting_graph_source_integration_test.go` / authored tests | Directly inspected applicable assertions and test seams; new duplicate/cycle and registry-aliasing cases are future work | Characterization and narrow regression; E-023–E-026 |
| `contracts/graph-projection/schemas.v2.json`; Graph/Reporting NLSpecs and boundary decision / authored projections and owner documents | Inspected current generation/authority, dependency and error requirements, current read/traversal capabilities and stale Reporting §15.2 wording; schemas remain downstream | C-001/C-002/C-006/C-010; E-023–E-026 |
| Graph, NF, Reporting, Recovery and server test-family manifests; Make surface and diagnostic/audit implementations / authored routing and tooling | Inspected exact rows and public command side effects; no harness/routing change in GP-04 | F-015; E-027 |

### GP-05 changed-source reinspection

The target remains 36 authored Go files in five packages (22 production, 14 test); no target file was added, removed or moved. All changed bodies and their direct callers/assertions were reinspected. The current catalog has 11 Graph rows and 43 top-level selectors; nested regression cases execute under their existing parents.

| Changed seam | Current disposition / evidence |
| --- | --- |
| `engine_v2.go`, `engine_v2_test.go` | Fallible duplicate/cycle admission and public-entry regression cases; E-031/E-032 |
| `restore/contract.go`, `restore/service.go`, their tests | Owned nested metadata, exact current binding, mutation isolation/reconstruction and pre-clear rejection; E-033/E-034 |
| `canonical.go`, `input.go`, `project.go` and private work initialization | Removed inactive mode, unused state/arguments/defaults and redundant sorting; unchanged valid semantics/goldens; E-035 |
| `result_v2_ports.go`, `boundary_guard_test.go`; NF/Reporting source adapters and tests; existing Reporting `graphsourcecontract` | Consumer-owned sentinels without aliases, positive value guard, wrapped mapping and real provider cases; E-036 |
| Reporting §15.2, authored Graph test-family selector, Make-generated topology index | Adopted retained-input vocabulary; one additional top-level selector; generated change is only its input hash/digest; E-036/E-037 |

### GP-06 affected-source inventory

Graph now has 39 authored Go files in the same five packages: 23 production and 16 test files. The 36-file tables above remain historical baselines. New cases are subtests under existing selected parents, so Graph routing remains 11 rows/43 top-level selectors; no new routing projection is needed for GP-06.

| Path / seam | Responsibility, ownership and coverage |
| --- | --- |
| `field_paths.go` (new, private Graph root) | Shared lexical classifier and context-specific field admission; used by validation and source/projected resolution; standard library only; F-018 |
| `aggregation_v2_test.go` (new, Graph root test helper) | Scope collision/admission, mixed execution, endpoints, merge/reference/identity/permutation and failure cases under the semantic-ordering parent; F-016/F-017 |
| `field_paths_v2_test.go` (new, Graph root test helper) | Canonical source names and prohibited/wrong-context paths across filters, inclusion, definitions, grouping and endpoints; F-018 |
| `engine_v2.go`, `aggregation.go`, `validation.go`, `project.go`, `engine_v2_test.go` | Reinspected and changed scoped dependency construction, once-only execution, classifier callers and positive canonical test paths; high semantic/identity risk covered by V-001/V-017–V-019 |
| `internal/modules/networkflow/graph_streaming_test.go`, `graph_temporal_test.go` | Complete immutable binding fixtures captured from pre-GP-06 source for both production forms; V-023/V-027 |
| `restore/service_test.go` | Exact mixed-chain reconstruction and incompatible binding rejection under existing restore parent; V-010 |
| Graph §§3-J/4/5 and this tracker | Governing clarification, compatibility, sequential checkpoint and acceptance accounting; V-007 |

### GP-07 inventory changes

The only new files are Graph's private `execution_state.go` (one owner for emitted objects, indexes and cumulative limits) and `contributor_order_v2_test.go` (immediate-provenance expectations). Existing engine, aggregation, direct mapping, canonical writer, field paths and validation now implement the clarified contract. Cases live under existing pure, semantic, bounds and restore parents. The authored semantic fixture and its Make-generated embedded copy carry the one explained golden correction. No routing input, export, SQL table or peer production file changed in GP-07.

## 3. Module Boundary Diagnosis

GP-09 keeps the existing Graph boundary. The root owns canonical input, semantic result content and the shared completed-result proof. PostgreSQL adapters call that proof before publication and on content reads. Network Flow owns selected bindings and rematerialization, and Reporting owns its retained input. All proposed cleanup stays private to Graph. The admission correction belongs in the root proof and its authored schema projection. See F-029–F-032.

The target is a cohesive supporting subsystem: a **deep pure engine with a small entrypoint**, a **Graph reconstruction orchestration layer**, and **Graph-owned persistence adapters**. The restore writer is a mutation coordinator for derived Graph publication only. It is not a thin all-purpose application facade, accidental catch-all, transport adapter, frontend controller or grid-vendor integration layer. Its multiple packages implement one owner boundary rather than demonstrate unrelated mixed responsibility.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Pure admission, normalization, derivation, canonical identity and result proof | GP root | Graph Projection | keep | E-001b/c, E-003 | F-002; C-001–003. A generic service/store facade would enlarge the expressly narrow root. |
| Shared exact bounds | `internal/modules/graphprojection/internal/semanticlimits` | Graph Projection private implementation | keep | E-003; contract parity tests | F-002; C-001/C-004. No general test-util or platform owner is needed. |
| Publication, exact reads and traversal | `postgresresult` | Graph Projection persistence adapter | keep | E-001b §8, E-004 | F-003; C-003/C-004. SQL belongs to the cohesive operational adapter. |
| Bounded cleanup capabilities | `postgresresult.Cleaner`; policy/recheck in NF | Graph for locked result operations; Network Flow for declaration reachability and sweep policy | keep | E-004/E-006 | F-003; C-005. Result lock precedes owner declaration lock; no global reachable-ID list. |
| Current-only Graph rebuild and typed source registry | `restore` | Graph Projection | keep | E-001b §9, E-005 | F-006; C-006. Concrete source enumeration remains Network Flow-owned. |
| Clear/publish/reconcile/commit/postcommit proof | `postgresrestore`; GA composition | Graph adapter and application assembly | keep | E-005/E-008 | F-003/F-006; C-006/C-007. Reconciler mandatory; Graph does not interpret Jobs payloads. |
| Graph-view identity, saved declarations, routes, refresh and authorization | NF adapter/application/routes/source registration | Network Flow | keep | E-001d, E-006 | F-007; C-008/C-009. These responsibilities are already outside this target. |
| Exact admission and retained report input | NF contribution; Reporting registry/retention | Network Flow source semantics; Reporting consumer lifetime | keep | E-001e, E-006/E-007 | F-001/F-007; C-010. Graph leases are not a permanent boundary. |
| Structural Recovery and successful-claim activation | Recovery contract owner; server/GA wiring | Recovery and application assembly | keep | E-001f, E-008 | F-006; C-007. Workbook rebuild, structural completion and Graph availability are distinct. |
| Restore clear scope and transaction comment | `postgresrestore/writer.go`, `Clear` comment | Graph adapter maintenance | keep | E-016 | F-004 corrected; no executable change. |
| Pre-publication failure characterization | Restore service, GA assembly and NF integration tests | Graph restore/application assembly/source-owner tests | keep | E-017/E-018 | F-005/RB-002: service order is observable; real rollback evidence passed; no sequencing change was needed. |
| Aggregation dependency admission and ordering | GP private normalization and validation | Graph Projection | keep | E-023 | F-011; reject malformed graphs before destructive normalization; no new package or generic graph framework. |
| Restore definition ownership and exact current binding admission | `restore` constructors/accessors and service validation | Graph Projection restore | keep | E-024 | F-012; own nested slices and validate one packaged current definition. |
| Unused canonical modes, state and parser/normalization plumbing | GP private engine | Graph Projection | keep | E-025 | F-013; remove dead branches and redundant work within the current boundary. |
| Source selection and snapshot-staleness diagnostics | GP root sentinels used only by NF/Reporting | Existing Reporting `graphsourcecontract` | move | E-026 | F-014; internal symbol relocation, no alias or public reason-code change. |

## 4. Public Contract and Behavior Freeze Map

### GP-09 contracts

| Contract | Owner and disposition | Witness required |
| --- | --- | --- |
| C-001/C-002, input, limits and identity | Preserve Graph §§3–7. Presence of optional members, lexical object keys, semantic array order, numeric spelling and the first applicable failure determine observable behavior. | Exact configuration/source digests, IDs, canonical output and public errors through `ProjectV2`. Include Network Flow default and time-bucket bindings. |
| C-002/C-003, result content | Preserve Graph §§6–8. The same seven content fields feed the digest and envelope. Projection and admission of a stored result each retain their own integrity checks. | Exact envelope and child bytes for empty, direct, reverse and aggregated output, including historical numeric rows. |
| C-003/C-004/C-010/C-011, family provenance | Separate correction under Graph §§5, 6.3 and 8. A direct object has one typed source reference. A reverse edge shares its forward edge's source and mapping. Aggregates carry ordered contributor references. | Rebind malformed objects with matching bytes and digest. Prove rejection by validation, publication and content reads, while the identity-only lock and valid stored rows keep their existing behavior. |

Preserve owner-aligned behavior, not accidental acceptance of invalid input. Each row separates observed implementation from owner requirements; test presence is not a current pass. GP-04 separately plans the owner-required aggregation correction in S-006 and registry hardening in S-007, alongside structural cleanup. These are authorized in GP-05 and do not change adopted valid-input semantics. No contradiction between applicable primary owners was found; an actual new contradiction must block its dependent decision explicitly.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| C-001: closed v2 invocation/input, semantic matrix, bounds and cancellation | Graph NLSpec §§3–5, 7, 8.1 | E-003/E-023/E-025: `ProjectV2(ctx, InvocationContextV2, []byte)`; trusted identity separation; normalization before semantic validation | Engine rejection, trusted-context, ordering, cancellation, maximum-bound and semantic tests | S-006 adds duplicate/cycle failure cases; S-008 preserves rejection, numeric/default/null/wildcard and capacity behavior while removing dead plumbing. | high | Preserve valid semantics; correct F-011 separately; F-013 is private cleanup; V-001/V-008/V-017–V-019. |
| C-002: result/object IDs, canonical bytes, output order and digests | Graph NLSpec §§4–6 | E-003: private encoding/framing, sorted output, result binding and resource conversion | Empty/non-empty golden identity tests; semantic merge/order tests | Preserve exact-byte/digest vectors, not merely graph-equivalence assertions. | high | Preserve; F-002; V-001/V-008. Publication time never enters semantic identity. |
| C-003: completed-result integrity and atomic publication/replay | Graph NLSpec §§6, 8 | E-003/E-004: shared value validation, full-binding reads, row/envelope fidelity, borrowed transaction, same-ID exact replay | Engine proof helper and storage proof helper; parent golden/integration tests; tampered bytes/rows and rollback cases | No additional tests required without movement. Any later publication refactor must retain invalid-input-before-insert and exact stored-content assertions. | high | Preserve; F-003; V-002/V-008. Metadata-only locking is distinct from content admission. |
| C-004: ordered reads and bounded breadth-first traversal | Graph NLSpec §§7–8 | E-004: verified result reads; sorted seeds/kinds, direction/depth/count limits and output-order traversal | `TestResultV2PublicationReadTraversalAndCleanup_Integration`: outgoing, incoming, depth-zero filtered and over-bound cases | Preserve these assertions; expand only for the actual traversal seam selected later. | high | Preserve; F-003; V-002. No traversal API redesign or frontend selector change is planned. |
| C-005: cleanup, admission/publication locking and selected-result protection | Graph NLSpec §8/§8.1; NF declaration owner | E-004/E-006: exact source owner, oldest `(published_at,result_id)`, one skip-locked envelope, same-transaction NF declaration recheck | Cleanup locking/rollback test; NF retained-source integration admission-lock assertion; NF race test discovered, not fully inspected | Before any lock-order move, inspect and run the exact NF race row as well as Graph storage rows; Graph-only accounting cannot prove peer selection policy. | high | Preserve; F-003/F-007; V-002/V-009. No persistent Graph lease or private Reporting-input deletion. |
| C-006: Graph rebuild v5, registry v5, binding v7, three derived tables, closed failure/result proof | Graph NLSpec §9; adopted boundary decision | E-005: exact current definitions; metadata admission before clear; source preflight follows clear within caller transaction; mandatory reconciliation and committed-state proof | Restore contract/service/writer suites; contribution guards | F-005/RB-002: S-003 observes clear/enumerate/publish and closed failure results; S-004 proves actual rollback after enumeration/cancellation/binding failure. E-017 establishes service evidence; E-018 establishes managed-service rollback and successful lifecycle evidence. | high | Preserve; F-004/F-005/F-006; V-002/V-008. Failure claims no committed cleared/rebuilt facts; current definitions are not fallback codecs. |
| C-007: structural Recovery completion versus Graph readiness after successful claim | Core 01 REQ-01-424/625A; Graph §9; Core security/fatal lifecycle | E-008: GA repeatable-read transaction; server uses original admitted identities, one activation attempt, readiness gate and fatal callback for indeterminate outcome | `TestClaimedGraphReconstructionUsesOneTransaction_Integration`; `TestClaimedGraphActivation_Unit` | Preserve failed/unclaimed gates, original identities, same-transaction enumeration/publication and indeterminate handling; V-012/V-015 are required final GP-03 checks. | high | Preserve; F-006; V-012/V-015. Structural Recovery must not reconstruct inactive extension state or become the workbook provider path. |
| C-008: graph-query/saved-graph HTTP paths, statuses and envelopes | Network Flow §19.1 and public route owner; Core HTTP envelope | E-006/E-009: generated route catalog wired through NF handlers; route table below | NF saved-graph/time-bucket lifecycle and contributor route tests discovered; inspected saved-read failure assertions | Preserve owner route tests before any consumer-facing change; no HTTP change is selected here. | high | Preserve; F-007; V-009. Graph root owns no routes, HTTP envelope or public job resource. |
| C-009: authorization, saved-graph version/refresh and error precedence | NF NF-REQ-170a–b; Core authentication/incident admission | E-006: operation admission checks membership/roles before resource handling; direct reads re-admit; refresh delegates `base_graph_view_version`/`client_txn_id` to NF commands | `TestSavedGraphReadFailures_Unit` checks authorization before derived availability; lifecycle route tests are identified separately | Preserve roles, CSRF/auth boundary, optimistic version, replay and availability ordering. Detailed mutation transaction assertions require NF reinspection before changing that owner. | high | Preserve; F-007; V-009. Declaration `graph_view_version` and `materialization_generation` are not Graph result lifecycle or generic Revisions change sets. |
| C-010: exact Reporting admission and frozen input lifetime | Reporting REQ-RPT-076–078/099g; Graph §8 | E-006/E-007: result lock then declaration validation, complete contribution freeze, operational timestamp removal, independent admission digest | NF `TestReportingGraphSourceRetainsSelectedInput_Integration`; Reporting `testFrozenGraphSources` under `TestGraphSourceRegistryReadsExactResultAndOwnsReasonMapping` | Preserve retained-input independence after source unavailability, rejection of changed admitted input and no later source callback. | high | Preserve; F-001/F-007; V-009. Retired lease vocabulary must not become a compatibility requirement. |
| C-011: Graph/Recovery/generated route contracts and schema membership | Applicable adopted owners; authored machine projections downstream | E-009: Graph index/schema/semantic/maintenance inputs; generated Graph/Recovery/route packages; result table test | Graph contract projection, restore contract/current contribution, storage-maintenance and schema integration tests | Preserve exact current generations and generated-output policy. Any future input change regenerates through Make; never edit generated outputs. | high | Preserve; F-008; V-003. S-003/E-017 generated drift check passed; product contract inputs are unchanged. |
| C-012: semantic test rows, owner coverage and helper reachability | Testing Harness procedure; machine catalog as evidence accounting | E-010 historically 11 owner rows/41 selectors; E-017 updates this to 11 rows/42 selectors; helper parents below | Root/restore boundary tests; owner rows across GP and GA | Keep complete required owner coverage distinct from focused selections; do not infer test evidence class from names. | medium | Preserve; F-009; V-008. Rows/phases do not define runtime architecture or conformance completeness. |
| C-013: workbook entities row/query/mutation, save state, conflicts, inspector, revisions, saved views and view schemas | Core and corresponding workbook/source owners; outside Graph NLSpec §1 | E-001b/f/g, E-003: Graph schemas/types concern pure graph values; no workbook provider or row mutation facade in target | No target-owned workbook tests selected | Not applicable to the retained boundary. Reopen owner discovery only if a future task proposes crossing this separation. | no direct change | F-010; no slice/validation needed. No claim of an exhaustive workbook audit. |
| C-014: direct WebSocket paths/events, UI selectors, shell/controller state and grid-vendor integration | Collaboration, browser and package owners; outside this target | E-003–E-006: root/adapters and NF HTTP consumers inspected; no target-owned WebSocket/UI implementation | No direct target browser/selector evidence required | Not applicable to this no-change backend diagnosis. Indirect user-visible graph behavior remains NF-owned. | no direct change | F-010; V-004 records non-applicability. No UI redesign or visual-golden work. |

### GP-04 affected contract details

| Contract | Adopted owner and required behavior | Observed exposure / evidence | Planned disposition and validation |
| --- | --- | --- | --- |
| C-015: duplicate-safe acyclic aggregation admission | Graph §§3-J, 4–6: duplicate detection, acyclic projected dependencies, deterministic normalized identity and no partial result | GP-04 observed duplicate overwrite and dependency-stall fallback (E-023); S-006 now rejects both before normalized identity (E-031/E-032) | Separate owner-required correction S-006/F-011; first record reproductions; valid ready-batch ordering and goldens stay exact; V-001/V-017–V-019 |
| C-016: stable admitted restore definition and current-only binding | Graph §9 and adopted boundary decision: exact selected canonical definition/digest, actual typed algorithm/schema and rejection before mutation | GP-04 observed nested aliases and generation iteration (E-024); S-007 now isolates metadata and admits only the packaged current binding (E-034) | S-007/F-012: private defensive copies and exact current-definition admission; no digest-only shortcut; V-010/V-012/V-013/V-020 |
| C-017: source-owner selection failures and Reporting reason mapping | Graph §8.1 root allocation; Reporting REQ-RPT-076–078/099g: source-owner admission and unchanged public safe reasons | GP-04 observed misplaced selection sentinels and obsolete “leased” prose (E-026); S-009 relocates errors to the consumer contract and corrects retained-input wording (E-036) | S-009/F-014: move to existing consumer contract without aliases; correct superseded wording; direct/wrapped error and real provider evidence; V-001/V-021–V-024 |

### GP-06 clarified contract details

| Contract | Owner / required behavior | Implementation and evidence |
| --- | --- | --- |
| C-018: scoped dependencies and one execution order | Graph §§3-J/4/5: producer lookup uses scope and kind, endpoint references name vertex rules, dependencies deduplicate, own output is excluded, complete ready batches sort by ID | S-012/S-013; F-016/F-017; E-046–E-049 and E-053. One admitted order drives one pass with per-rule contributor snapshots. Genuine cycles/duplicates reject without a result. |
| C-019: canonical field vocabulary and contextual resolution | Graph §4: exact source-kind names read the corresponding source; kind and projected-kind substitutes are forbidden; defaults never legalize invalid paths | S-014/F-018; E-050–E-053. Shared classifier plus context-specific admission, including inclusion predicates, and source/projected readers. |
| C-020: compatibility boundary for corrected computation | Graph §§4–6/9; NF producer definitions: preserve v2 interfaces, identity framing and restore v5/registry v5/binding v7; reconstruction must match the stored expected binding | Corrected generic normalization/output may change identities; aliases now reject. Unaffected Graph goldens and both complete NF producer binding tuples remain exact. No automatic rewrite, historical execution mode or migration. V-008/V-014/V-023/V-027 plus exact restore rejection. |

### GP-07 contract decisions

C-001/C-019: admit definitions against every configured input context before defaults; bound the first attempted excess direct, reverse or aggregated emission. C-002/C-020: group and match endpoints with the same fallible canonical scalar encoding; compare and expose immediate contributor provenance in the specified tuple order. Generic outputs and expected bindings affected by those corrections can change. C-003/C-006: reconstruct exactly against the existing binding or fail atomically; never rewrite a mismatch. C-011/C-012: current generations, schemas and routing remain stable; the changed aggregate fixture is proved separately in E-068/E-069. Tests are evidence for these owner decisions, not their authority.

### Indirect HTTP surface

All paths below are relative to `/api/v1/incidents/{incident_id}/network-flow`. These are Network Flow contracts, not Graph-owned transport. E-006 and the directly inspected generated catalog E-009 establish current registration. Statuses below are normal success statuses, not an exhaustive replay/error matrix.

| Method | Path suffix | Route identity | Normal success / envelope |
| --- | --- | --- | --- |
| POST | `/graphs/query` | `nf.graphs.query` | 200; Core success envelope with NF graph-query result. |
| POST | `/graphs/contributors/query` | `nf.graphs.contributors.query` | 200; NF contributor result in Core success envelope. |
| GET / POST | `/graph-views` | `nf.graph_views.list` / `nf.graph_views.create` | 200 list v4 / 202 accepted v4 with declaration and job status route. |
| GET / PATCH / DELETE | `/graph-views/{graph_view_id}` | `nf.graph_views.get` / `.patch` / `.delete` | 200 get v4 / 200 mutation / 204 retirement. |
| POST | `/graph-views/{graph_view_id}/refresh` | `nf.graph_views.refresh` | 202 accepted v4; refresh request v1 with version and transaction identity. |
| GET | `/graph-views/{graph_view_id}/result` | `nf.graph_views.result` | 200; exact selected result through NF presentation. |
| POST | `/graph-views/{graph_view_id}/contributors/query` | `nf.graph_views.contributors.query` | 200; selected-result contributor query. |

Any current incident member may read; editor/admin may create, rename or refresh; reviewer/admin may retire. NF handlers retain Core authentication/CSRF and incident admission. Ordinary get/list/result/contributor exposure excludes retired declarations. NF owns declaration version conflicts, replay, source-snapshot checks, materialization generation and missing-result classification; Graph result identity contains none of these mutable attempt facts.

### Characterization and accounting posture

The authoritative routing input inspected is `tools/test_families/module.graphprojection.json`, linked to `module.graphprojection.verification.engine_and_storage` in `contracts/verification/owners/module.graphprojection.json`. All row IDs below are relative to `module.graphprojection.`. Current GP-03 counts are selectors in this manifest, not pass counts or assertion counts. GP-02 had 41 selectors; S-003 adds one for 42 total.

| Row suffix | Package relative to repository root | Evidence class / selected tests | Assertion posture |
| --- | --- | --- | --- |
| `engine.v2_deterministic_current_restore_rebuild` | `./internal/modules/graphprojection/restore` | unit / 14 | Current contracts, construction, phase order, source/candidate/clear failures, bounds, cancellation, determinism, readiness and positive boundaries; E-017 passed this row. |
| `engine.v2_maximum_semantic_bounds` | `./internal/modules/graphprojection` | unit / 1 | Maximum semantic resources; catalog resource profile is io_heavy, not a runtime architecture label. |
| `engine.v2_pure_determinism` | `./internal/modules/graphprojection` | unit / 8 | Root boundaries, golden proof, rejected members, ordering, trusted identity and cancellation. |
| `engine.v2_semantic_matrix` | `./internal/modules/graphprojection` | unit / 6 | Closed merge matrix, non-empty golden, value/text bounds and default/null/failure behavior. |
| `storage.claimed_graph_reconstruction_enumerates_and_publ_c18fddd44a` | `./internal/app/graphassembly` | integration / 1 | Actual repeatable-read/runtime-role clear and source enumeration; S-004 adds nonempty rollback and cancellation subtests. |
| `storage.graph_projection_current_head_schema_preserves_r_cf74406ebd` | `./internal/modules/graphprojection` | integration / 1 | Current tables, source/result FK separation and cleanup index. |
| `storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` | `./internal/modules/graphprojection/postgresresult` | integration / 1 | Publication, exact reads, traversal, cleanup and producer-backed tamper subtests. |
| `storage.restore_publication` | `./internal/modules/graphprojection/postgresrestore` | integration / 7 | Includes the constructor test named `_Unit`, which opens PostgreSQL. |
| `storage.result_v2_cleanup_locking` | `./internal/modules/graphprojection/postgresresult` | integration / 1 | Owner isolation, skip-locked candidate exclusion and rollback. |
| `storage.service_free_contract_assertions_for_result_v2_a_4661600b77` | `./internal/modules/graphprojection/postgresresult` | unit / 1 | Maintenance projection parity without service setup. |
| `storage.v2_contract_projection` | `./internal/modules/graphprojection` | unit / 1 | Packaged Graph contracts/limits/fixtures/current cleanup definitions. |

`checkCompletedResultProofsV2` runs under `TestProjectV2GoldenDeterminismAndCompletedResult_Unit`. `checkCurrentResultProofs` runs under `TestResultV2PublicationReadTraversalAndCleanup_Integration`. Reporting's `testFrozenGraphSources` runs under `TestGraphSourceRegistryReadsExactResultAndOwnsReasonMapping`, outside the Graph owner union. These are not independently selectable top-level tests.

## 5. Coupling and Boundary Findings

### GP-09 findings

| Open finding | Required correction | Slice |
| --- | --- | --- |
| F-039 | Include all registered producers of a consumed projected kind in the dependency proof. | S-034 |
| F-040 | Recompute a direct vertex's ID and sort key from its immediate source and mapping. | S-035 |
| F-041 | Direct and reverse edge endpoints must both be direct vertices. | S-036 |
| F-042 | Require one input scope and kind, disjoint contributors and one ordered endpoint-producer pair across groups of one rule. | S-033 |
| F-043 | Close registry row shapes, cross-check nested references, and align representable identifier domains. | S-038 |
| F-044 | Finish shape and scalar admission request-wide before uniqueness. Preserve specific public errors. | S-039 |
| F-045 | Compare property and metadata declaration keys by tuple components. Affected existing bindings require rematerialization. | S-037 |

| Finding | Evidence and classification | Action |
| --- | --- | --- |
| F-029, discarded member order | E-077, `should_fix`. Input builders make `canonicalObject` lists. The v2 digest path converts them back to maps before the writer sorts their keys. The ordered writer branch implies a second encoding. | S-023 removes the list representation and conversion path. C-001/C-002, V-001/V-017/V-028. |
| F-030, unreachable collection checks | E-078, `should_fix`. Nine typed collection counts repeat earlier raw-array checks. Some later checks even use a different limit key. | S-024 removes only the nine dominated checks identified in §7. C-001, V-018. |
| F-031, duplicate result-content builders | E-079, `should_fix`. `projectedGraphResource` and `ProjectionResultV2.outputResource` separately enumerate the same content. A future field could enter producer validation but be omitted from the output digest. | S-025 gives both callers one private builder. C-002/C-003, V-001/V-017/V-028. |
| F-032, incomplete family-specific admission | E-079, `must_fix` for S-026. `validResultObjectV2` checks a typed reference only when non-null. The schema also permits null on a direct vertex or edge. The reverse-link proof does not exercise an ordinary direct edge without a reverse partner. | S-022 must first isolate the gap at public boundaries. S-026 brings the shared proof and authored schema into agreement with Graph §6.3. C-003/C-004/C-011, V-001/V-019/V-028. |
| F-033, source duplicates masked by input ceilings | Graph §4 puts source-ID uniqueness before collection admission. E-078 and E-082 locate the raw ceiling checks before the current duplicate scan. | S-022 captures the competing failures. S-027 corrects the public reason for multiply invalid inputs under C-001, V-001/V-018. |
| F-034, registry membership | A rebound result could claim an invented producer or an unlisted source kind. | S-028 / E-094, registry admission. |
| F-035, endpoint prerequisites | Existence-only endpoint checks missed cycles and direct aggregate endpoints. | S-029 / E-095, combined provenance proof. |
| F-036, contributor key shape | The packaged schema admitted raw keys on projected references and vice versa. | S-030 / E-096, family-local contract. |
| F-037, source scalars and reference timing | Uniqueness or resource errors masked source scalar failures. A source filter could mask a missing endpoint. | S-031 / E-097, scalar admission and submitted-reference validation. |
| F-038, cancellation during uniqueness | An oversized input was sorted before the next callback. | S-032 / E-098, cancellable scan before sorting. |

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| F-001: current Graph glossary/domain/framework lease vocabulary corrected in S-001 | E-002 framework §4; E-001b §1/§8 and E-001e REQ-RPT-099g explicitly supersede live leases; current code E-004/E-007 | high if old prose drives a new API | should_fix | Framework/supporting-document maintenance; current Graph/Reporting owners govern behavior | S-001 aligned the glossary/domain/framework with C-005/C-010/C-011. Current Core 04 already agrees; historical records remain intact. E-015 records lint and semantic review; no owner amendment or migration required. |
| F-002: pure root and private semantic owner already form a deep boundary | E-001b/c; E-003 root imports, single engine entrypoint, private helpers and positive guards | high if a generic facade expands the root | intentional/no_action | Graph Projection | Keep C-001/C-002; preserve V-001/V-005/V-008 if later touched. No new service, helper package, graph-view identity API or mutable engine option is justified. |
| F-003: SQL/traversal/publication and derived transaction coordination are cohesive adapter responsibilities | E-004/E-005; Graph §8; NF cleanup recheck E-006; mandatory GA reconciler E-008 | high for lock order, byte fidelity and rollback | intentional/no_action | Graph persistence adapters; consumer-owned policy and application assembly | Keep C-003–007. Preserve borrowed handles, one-candidate cleanup and narrow ports. V-002/V-009 apply to later changes. No storage-to-platform relocation or extra facade is proposed. |
| F-004: stale four-table `Writer.Clear` comment corrected in S-002 | E-005 historical comment versus clear SQL, restore contribution, E-001b §9; E-016 correction | low documentation ambiguity | should_fix | Graph adapter source-comment maintenance | S-002 corrected the comment against SQL, Recovery contribution and Graph §9. E-016 confirms no executable diff; C-006 and migration history unchanged. |
| F-005: missing clear-order and prepublication rollback evidence | GP-02 E-005 identified a no-op Clear fake; S-003/E-017 replaced it with phase assertions. S-004 adds actual assembly and NF failure evidence | high for undetected partial mutation or false readiness | must_fix | Graph restore, application assembly and NF tests | C-006/C-007/C-012; close only after S-003 and S-004 pass V-010/V-012–V-014. S-003/E-017 and S-004/E-018 passed; F-005 and RB-002 are closed. Final-source qualification follows in S-005. |
| F-006: structural Recovery and claimed Graph reconstruction already have separate ownership | E-001f; E-005 restore contribution; E-008 server/GA wiring | high for readiness and current-generation admission | intentional/no_action | Recovery; Graph restore; application assembly | Preserve C-006/C-007 and V-002/V-009. Do not merge Graph into workbook providers, introduce inactive-source callbacks, or alias older bindings. |
| F-007: declarations, graph-view identity, HTTP authorization, jobs and retained consumer input remain with their semantic owners | E-006/E-007; E-001d/e; direct NF graph adapter and Reporting provider port | high for public behavior/lifetime | intentional/no_action | Network Flow; Reporting | Preserve C-008–010. No Graph ownership transfer for timeline, projections, entities/indicators, evidence, links, imports or saved views is supported. Any future peer change must load its own scope and evidence. |
| F-008: generated contracts and typed value surfaces are downstream; manual edits would create drift | E-009 generated-artifact policy and code markers; E-003/E-005 contract tests | high for current identities and machine projection fidelity | intentional/no_action | Authored Graph/Recovery/route contract inputs and their Make generators | C-001/C-011; product contracts unchanged. S-003/S-004 update authored test selectors and regenerate the topology index through `make generate`; V-003 checks drift. No generator/test reads this tracker or any Markdown. |
| F-009: test helpers and suffixes are not execution accounting | E-010 historically 11 rows/41 selectors; GP-03 11 rows/42 selectors; E-003/E-004 helper parents; writer constructor uses pgtest despite `_Unit` suffix | medium for under-selection and false pass claims | intentional/no_action | Graph test owner; Testing Harness evidence accounting | Preserve C-012; V-008 for complete Graph owner selection, V-009 for affected peers. Only authored selectors and their generated index change; no harness mechanism change is required. |
| F-010: no direct workbook, transport, frontend or vendor responsibility was found inside the target | E-001b scope; E-003–006 imports/interfaces and inspected consumers | medium if generic target-name heuristics force a move | intentional/no_action | Existing Graph boundary; separate workbook/browser owners | C-013/C-014 are explicitly inapplicable to this target diagnosis. No duplicated workbook row/view-schema or grid-vendor logic was observed. No UI-review, browser, selector or visual-golden slice is justified. |

F-001–F-010 retain the earlier bounded findings and dispositions. GP-03 exposed no runtime defect while closing F-005; that historical result does not settle the additional GP-04 exposures below. No production dependency on test utilities or Markdown, source-owner SQL in the root, hidden authorization facade, or generic Revisions mutation side effect was found in the inspected boundary. These are scoped inspection findings, not a repository-wide absence proof.

### GP-04 cleanup and hardening findings

| Finding | Evidence and affected contracts | Classification / priority | Remediation owner and areas | Required action / risk if unresolved | Slice / validation |
| --- | --- | --- | --- | --- | --- |
| F-011: aggregation normalization can discard duplicate IDs and continue after dependency ordering stalls | E-023; C-001/C-002/C-015 | must_fix for S-006; highest | Graph engine; implementation and tests | Reproduce first, reject duplicates before keyed insertion, and make ordering fail closed on cycles. Silent normalization can produce incorrect successful results or reconstruction. DONE E-031/E-032. | S-006; V-001/V-017–V-019 |
| F-012: restore definitions retain nested aliases and indirect generation-selection plumbing | E-024; C-006/C-016 | must_fix for S-007; high | Graph restore; implementation and tests | Own nested slices at construction/access boundaries and validate exactly the current packaged binding. Aliasing can separate definition contents from their recorded digest; generation iteration invites unnecessary compatibility handling. DONE E-033/E-034. | S-007; V-010/V-012/V-013/V-020 |
| F-013: unused canonical mode, work state, parameters, initializers and repeated normalization | E-025; C-001/C-002 | should_fix; bounded mechanical cleanup | Graph private engine; implementation and regression tests | Delete only evidenced remnants and give each mapping/definition collection one sorting stage. Leaving them implies unsupported modes/dependencies and obscures admission guarantees. DONE E-035. | S-008; V-001/V-017–V-019/V-025 |
| F-014: source-selection policy errors were exported by the pure result owner; related consumer wording was stale | E-026; C-008/C-010/C-017 | should_fix; owner-boundary cleanup | Reporting graph source contract, NF provider, Graph root; implementation/tests/documentation/routing | Move two sentinels without aliases; preserve public reason mapping; add a positive root-value guard and align Reporting §15.2 with retained input. Otherwise source-specific policy can spread into the root or lease semantics reappear. DONE E-036. | S-009; V-001/V-003/V-007/V-011/V-021–V-024 |
| F-015: future runtime changes need current compatible evidence beyond GP-03's bounded portfolio | E-020 historical scope; E-027 current commands/rows; C-007/C-011/C-012 | must_fix for S-010 acceptance | Graph and affected peer verification; validation and controlling documentation | Finalize/freeze, execute the selected portfolio, audit the Graph owner union, and retain evidence. Old passing summaries cannot prove new changes or release readiness. DONE E-037–E-042. | S-010; V-003/V-005/V-007/V-008/V-011/V-014–V-026 |

Retain current read/traversal capabilities: Graph §8 and the emitted `consumer_capabilities` advertise them even where production call searches find only tests. Retain caller cancellation callbacks, which NF and restore actually use. Retain immutable SQL, retirement rejection tests, and `GP…1` canonical identity domains; their names do not make them unsupported dispatch. No feature removal, new facade, generic cloning/transaction framework or topology move is justified by these findings.

### GP-06 follow-up findings

| Finding | Owner / areas | Remediation and rationale | Compatibility / unresolved risk | Validation / status |
| --- | --- | --- | --- | --- |
| F-016: kind-only dependencies conflate vertex and edge namespaces | Graph §§3-J/4; implementation/tests/specification/tracker | Index producers by scope and kind; deduplicate actual dependencies; preserve sorted ready batches and closed rejection. A graph of actual data flow avoids false coupling. | Affected generic normalization/identities can change; leaving this rejects valid inputs. No legacy fallback. | S-012; V-001/V-017–V-019; DONE E-047 |
| F-017: vertex-first execution ignores admitted mixed-scope dependencies | Graph §5; implementation/engine and restore tests/specification/tracker | One ordered traversal with private rule emitters and shared available-object/endpoint state. Admission and execution must describe one computation. | Previously incomplete generic output changes; stored bindings never rewrite automatically. Leaving this permits silent data loss and false success. | S-013; V-001/V-010/V-012/V-013/V-017–V-019; DONE E-049 |
| F-018: forbidden kind alias accepted while canonical source-kind paths reject | Graph §4; implementation/tests/specification/tracker | One private path classifier shared by validation/resolution; canonical source names in their contexts; no projected-kind replacement. Removes ambiguous lookup and validator drift. | Alias inputs reject; canonical source names work. No conversion or data migration. Leaving this preserves a false contract in tests. | S-014; V-001/V-010/V-014/V-017–V-019/V-023/V-027; DONE E-053 |

F-011's duplicate/cycle corrections and F-012–F-014 remain implemented, but F-016 invalidates the broad valid-input preservation claim. F-015/S-010's overall acceptance is superseded by S-015. Preserve their historical results; E-055–E-059 now close the follow-up acceptance gate with current compatible evidence.

### GP-07 resolved findings

| Finding | Status / owner | Gap and required remediation | Compatibility / unresolved risk | Validation |
| --- | --- | --- | --- | --- |
| F-018 (reopened) | DONE; Graph Tables 3-H/3-I, §4 | Definitions use output scope instead of actual producer read contexts; S-018 requires every applicable context before defaults | Cross-scope definitions become usable; wrong-context/default masking rejects; otherwise invalid admission persists | V-001/V-010/V-017–V-019/V-023/V-027 |
| F-019 | DONE; Graph §§4–5 | Shared fallible numeric scalar grouping and endpoint encoding; never hash failed encoding | Correct generic groups/endpoints and identities; otherwise decimal groups silently collapse | S-017; V-001/V-010/V-017–V-019 |
| F-020 | DONE; Graph §§5/7 | Cumulative checked emission across direct, reverse and aggregate output; preserve resource errors | Earlier closed limit rejection; otherwise admitted expansion may exhaust memory | S-019; V-001/V-010/V-012/V-013/V-017–V-019 |
| F-021 | DONE; Graph §§3-H/5/6 | Explicit immediate-provenance contributor tuple ordering, matching reference order and sortable key | Correct order-sensitive merges and generic identities; otherwise hash order contradicts owner | S-020; V-001/V-010/V-017–V-019/V-023/V-027 |

## 6. Refactor Workstreams

### GP-09 workstreams

| Closure slice | Predecessor | Gate |
| --- | --- | --- |
| S-033 / F-042 | E-104 | Cross-group input, contributor and endpoint proof. V-001/017/018/028. |
| S-034 / F-039 | S-033 | Scope-qualified inferred dependencies, unused declarations and cycles. V-001/017/028. |
| S-035 / F-040 | S-034 | Producer and reader agree on fixed direct-vertex identity and ordering. V-001/017/028. |
| S-036 / F-041 | S-035 | Endpoint family in both positions for direct and linked reverse edges. V-001/017/028. |
| S-037 / F-045 | S-036 | Independent before/after bindings, rematerialization and restore rollback. V-001/010/012/013/017/028. |
| S-038 / F-043 | S-037 | Registry schema, runtime proof, cross-references and identifier boundaries. V-001/003/011/019/028. Close S-026. |
| S-039 / F-044 | S-038 | Request-wide phase precedence, specific error and budget, cancellation, endpoint existence. V-001/017/018. Close S-027. |
| WF-35 / T-041 | S-039 | Finalize, freeze, qualify, explicitly audit, review the diff and lint the final tracker. |

Each P1 gate requires a fully rebound result, no publication insert attempt or stored row, independent malformed seeds rejected by ordinary and locked exact reads, vertex and edge reads and traversal, and a successful metadata-only binding lock.

| Repair slice | Finding and parent | Predecessor | Acceptance |
| --- | --- | --- | --- |
| S-028 | F-034, registry provenance under S-026 | E-093 reviewed entry | Scope-qualified membership, valid shared kinds and historical bytes. V-001/017/019/028/003/011. |
| S-029 | F-035, aggregate endpoints under S-026 | S-028 | Combined endpoint/contributor cycles and direct endpoints reject in the shared proof and stored reads. V-001/017/028. |
| S-030 | F-036, contributor-key schema under S-026 | S-029 | Packaged schema discriminates raw and projected keys. V-001/019/003/011. |
| S-031 | F-037, scalar precedence under S-027 | S-030 | Shape, scalar, duplicate, resource and reference errors retain the specified priority and safe counts. V-001/017/018. |
| S-032 | F-038, cancellable uniqueness under S-027 | S-031 | Cancellation during uniqueness precedes sorting and all nine collection verdicts remain bounded. V-001/018. |
| WF-35 / T-041 | Combined candidate | S-032 | Fresh frozen-source §8 portfolio, complete-owner audit, preservation review and final-text lint. |

The table defines the prerequisite chain. Each checkpoint records the exact selection, native exit, artifact root and any unexplained difference before the next workstream begins.

| Workflow / task | Prerequisite | Exit checkpoint |
| --- | --- | --- |
| WF-30 / T-036, characterize (S-022) | Reviewed owner mapping, passing entry lint, E-082 checkpoint and revalidated E-076–E-080 | Independent valid baseline and isolated malformed-result observations available to all later slices. |
| WF-31 / T-037, input encoding (S-023) | WF-30 | Canonical bytes and identities match the characterized baseline. |
| WF-32 / T-038, admission counts (S-024) | WF-31 | All nine boundary pairs and competing-failure cases match. |
| WF-36 / T-042, source-ID precedence (S-027) | WF-32 | A duplicate takes precedence over any of the nine collection ceilings. Unique overflow, earlier failures, issue budget and cancellation retain their contracts. |
| WF-33 / T-039, result content (S-025) | WF-36 | One content builder serves both paths and stored-row proof stays independent. |
| WF-34 / T-040, provenance (S-026) | WF-33 and S-022's malformed-result evidence | Schema and runtime enforce the family rules. Positive and negative storage witnesses pass. |
| WF-35 / T-041, qualification | WF-34 | §8 final portfolio, explicit complete-owner audit and §12 handoff pass on frozen material. |

| GP-08 phase | Dependency | Completed exit gate |
| --- | --- | --- |
| 1. Reopen and specify | Reviewed GP-07 index | User confirmed revised Graph §§3–8 and 11. Each finding has a Graph owner, compatibility disposition and witness in §1. |
| 2. Establish regressions | Confirmed owner semantics | §10 records pre-fix failures for all seven findings and the independent numeric storage fixture. |
| 3. Repair and validate | Reproducible failures | Focused Make checks cover shared encoding and limits, kind-specific edges, group endpoints, public cancellation and atomic restore. |
| 4. Qualify and hand off | Generated and finalized combined source | The complete Graph owner audit, named peer and static checks, final-text lint and staged source comparison are recorded in §§10 and 12. |

The table below preserves completed GP-01/GP-02 planning workflows. GP-03 implementation follows the additional authoritative chain below. The dependencies below are a directed acyclic graph; each checkpoint is an exit criterion. WF-05 facade redesign and WF-06 implementation sequencing are omitted because the adopted boundary and inspected code do not justify a structural move.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap | root | None | WF-01 | Establish literal target, safe output, authority, allowed writes and baseline | Framework, owner documents E-001/E-002, instruction files, tracker | Read instructions and owner status; Git/path identity inspection E-012 | Target, output, authority and baseline recorded in §1. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-04 | Account for all five packages and every target file without treating names as architecture | All 36 files in §2; direct callers E-006–E-008 | Reconcile live enumeration with §2; inspect declarations and selected bodies | All 36 files have an evidence-backed responsibility, caller, dependency, test posture and owner disposition. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04 | Freeze the direct and indirect behavior affected by a hypothetical move | E-001 owners; root/adapter ports; NF and Reporting consumers; generated projections E-009 | C-001–C-014 owner/test review | Every applicable contract has a governing owner; inapplicable browser/workbook surfaces are explained. |
| WF-03 | Characterization analysis | parallel | WF-02 | WF-07 | Separate inspected assertions, helper parents, catalog membership and unproved behavior | Target tests; GA/server tests; selected NF/Reporting assertions; E-010 catalog | Static assertion review only; record F-005 and RB-002 | Existing assertions and bounded gaps are recorded; none is mislabeled a passing test. |
| WF-04 | Boundary diagnosis | parallel | WF-01, WF-02 | WF-07 | Test whether the current Graph boundary merits retention, relocation or redesign | Root, semanticlimits, result/restore adapters, application assembly and consumers | Responsibility table §3 and classified findings §5 | Every found responsibility has a keep/move/split/defer disposition; retention conclusion follows adopted owners and inspected topology. |
| WF-07 | Validation and accounting plan | chain | WF-03, WF-04 | WF-08 | Preserve row/test mappings and discover commands without producing forbidden artifacts | Make definitions, owner catalog, test-ops instructions and tracker | E-011 discovery; §8 commands and restrictions | Eleven rows/41 selectors, helper parents, future command scopes and current execution restrictions are explicit. |
| WF-08 | Final handoff | chain | WF-07 | None | Reconcile all twelve sections and the allowed diff | This tracker only | Readback, document-structure/link inspection, baseline comparison E-012/E-013 | Binary planning criteria in §12 are reconciled; only the tracker is created; skipped lint and implementation authorization are reported separately. |

WF-03 and WF-04 may be performed independently once their prerequisites are satisfied. They are not a request to create parallel agents. No workflow depends on an omitted WF-05 or WF-06.

### Completed GP-03 remediation workstreams

| Workflow | Prerequisite | Goal / affected areas | Validation | Binary exit criterion |
| --- | --- | --- | --- | --- |
| WF-09 / S-001 | GP-03 bootstrap | Reconcile F-001 glossary, domain and framework prose with adopted frozen input and transaction locks | V-007; owner review | Current prose agrees; Markdown lint passes; tracker checkpoint recorded |
| WF-10 / S-002 | WF-09 | Correct F-004 writer comment and transaction responsibilities | Executable-diff inspection | Comment agrees with SQL/Graph §9; no executable change; tracker checkpoint recorded |
| WF-11 / S-003 | WF-10 | Close service portion of F-005 with observable phase order, failure states and exact selectors | V-010, V-003, V-011 | Required service cases pass; generated/schema checks pass; RB-002 remains open |
| WF-12 / S-004 | WF-11 | Close transaction portion of F-005 using real assembly and Network Flow binding failure | V-012–V-014 | Real clear followed by failure restores exact persisted state; successful restore remains valid; RB-002 closed |
| WF-13 / S-005 | WF-09, WF-10, WF-11, WF-12 | Finalize, validate frozen executable source and finish handoff | V-003, V-005, V-007, V-010–V-016 | Required bounded evidence recorded, preservation checked, all findings reconciled |

### GP-04 accepted workstreams, executed in GP-05

This prerequisite column is authoritative. GP-05 authorization supersedes the planning-only entry restriction. Current statuses and completion evidence are updated at each sequential checkpoint; correctness precedes cleanup, and no successor starts before its predecessor checkpoint.

| Workflow / slice | Status | Class | Authoritative prerequisites | Goal / affected seam | Validation | Risk / binary exit and checkpoint |
| --- | --- | --- | --- | --- | --- | --- |
| WF-14 / S-006 | DONE | root | Accepted GP-04 plan; GP-05 implementation authorization and E-030 revalidation | Fail-closed aggregation dependency admission | V-001/V-017–V-019 | Canonical drift or hiding invalid input; record initial reproductions, then passing duplicate/cycle/permutation cases and unchanged valid goldens before WF-15 |
| WF-15 / S-007 | DONE | chain | WF-14 | Stable registry ownership and exact current binding admission | V-010/V-012/V-013/V-020 | Digest/content divergence or weakened pre-clear admission; copy/malformed-binding cases and restore rollback pass before WF-16 |
| WF-16 / S-008 | DONE | chain | WF-15 | Remove unused private engine plumbing | V-001/V-017–V-019/V-025 | Accidentally removing a real default or changing order/identity; exact semantic regression passes before WF-17 |
| WF-17 / S-009 | DONE | chain | WF-16 | Move selection errors to existing consumer contract and protect root exports | V-001/V-003/V-005/V-007/V-011/V-014/V-021–V-024 | Wrong public reason mapping or weakened lock/admission order; producer/consumer/routing updates and targeted regressions pass before WF-18 |
| WF-18 / S-010 | DONE | chain | WF-17 (transitively WF-14–WF-16) | Finalize, freeze, validate and hand off bounded hardening | V-003/V-005/V-007/V-008/V-011/V-014–V-026 | E-037–E-042: all required outcomes/artifacts and limitations recorded; Graph owner audit passes; final tracker/handoff complete |

### GP-06 sequential workstreams

| Workstream / slice | Status | Prerequisite | Goal / implementation risks | Exit and checkpoint |
| --- | --- | --- | --- | --- |
| WF-19 / S-011 | DONE | Accepted follow-up and baseline | Reconcile owner clauses and superseded completion; avoid freezing defects as compatibility | Findings/compatibility/obligations recorded; V-007 passes |
| WF-20 / S-012 | DONE | WF-19 | Scope-aware deterministic dependencies; avoid false cycles, duplicate edges and ordering drift | Initial reproduction recorded; V-001/V-017–V-019 pass; tracker checkpoint |
| WF-21 / S-013 | DONE | WF-20 | One dependency-ordered execution pass; preserve contributor/endpoint/cancellation behavior | Nonempty mixed chains and restore proof; V-001/V-010/V-012/V-013/V-017–V-019 pass; checkpoint |
| WF-22 / S-014 | DONE | WF-21 | Shared field-path classification and canonical source lookup; prevent cross-context/default masking | Alias rejection/canonical positive matrix; Graph and NF checks pass; checkpoint |
| WF-23 / S-015 | DONE | WF-22 | Finalize, freeze and requalify entire combined diff; reject stale/incomplete evidence | §8 portfolio and Graph-owner audit pass; final preservation, lint and handoff recorded last |

### GP-07 sequential workstreams

| Workflow / slice / task | Status | Prerequisite | Areas, rationale and exit criterion |
| --- | --- | --- | --- |
| WF-24 / S-016 / T-029 | DONE | Authorized request; E-060 baseline | Specification/tracker: precise owner contract and reopened gate; all findings mapped; V-007 passes |
| WF-25 / S-017 / T-030 | DONE | S-016 checkpoint | Implementation/tests: one numeric tuple encoder prevents silent group loss; reproduced decimal/endpoint failures, exact outputs and restore; V-001/V-010/V-017–V-019 pass |
| WF-26 / S-018 / T-031 | DONE | S-017 checkpoint | Specification/implementation/tests: configured producer-context admission prevents data-dependent validity/default masking; cross-scope and all-context cases plus NF identities pass |
| WF-27 / S-019 / T-032 | DONE | S-018 checkpoint | Implementation/tests: one execution state enforces cumulative output ceilings before construction; bounded reproduction, exact/over limits, cancellation and rollback pass |
| WF-28 / S-020 / T-033 | DONE | S-019 checkpoint | Specification/implementation/tests: local provenance tuple replaces hash comparison; independent merge/reference/key expectations, restore and NF identities pass |
| WF-29 / S-021 / T-034 | DONE | S-020 checkpoint; S-016–S-019 DONE | Finalize/freeze combined diff, complete §8 portfolio including V-027 and explicit V-026 audit; preserve index/prior work, lint final bookkeeping, record handoff last |

## 7. Proposed Refactor Slice Plan

### GP-09 slices and entry gates

**S-022 — Characterize the baseline.** Begin after the specification entry gate. Start with `make task-guide OWNER=module.graphprojection`. Use `ProjectV2` and `CompletedResult` to capture valid bindings, canonical envelopes and ordered child bytes for empty, direct, linked reverse and aggregated results. Include the checked-in historical numeric bytes and both Network Flow producer bindings. Capture outputs independently of the planned replacement builders. For F-032, start with valid output and create two fixtures. Set one direct vertex's entity reference to null in the first. In the second, set an ordinary direct edge's relationship reference to null and use a graph with no reverse partner. Recompute the output digest, binding, envelope and child bytes and retain all other proof fields. Record what `Validate`, borrowed publication and exact reads actually do on the baseline. Also retain duplicate-only, overflow-only and combined outcomes through ProjectV2. Keep these reproductions separate from the passing structural regression portfolio. Exit when the isolated observations and compatibility vectors are retrievable (V-001/V-017/V-018/V-023/V-027/V-028).

**S-023 — Remove the ordered-object encoding path.** Structural, depends on S-022 (E-083). Have the v2 input builders construct ordinary maps. Delete `canonicalObject`, its member-list construction and map-conversion helpers and the writer branch that accepts it. Keep the public v2 canonical number and tuple encodings. Exercise omitted versus explicit-null defaults, optional reverse-edge kinds, nested inclusion filters, input permutations and exponent spelling. Exit on byte-identical valid configuration/source digests, object/result IDs, envelopes and child rows (F-029, V-001/V-017/V-028).

**S-024 — Remove unreachable input-count checks.** Structural, preceded by S-023. These are the exact `admittedResourceLimitIssues` counts dominated by `validateV2ResourceBounds`. Closed-shape checks run first. The raw-array check precedes typed parsing, which cannot grow any of these collections.

| Raw input array(s) | Later typed collection(s) to drop | Public first-excess reason |
| --- | --- | --- |
| `source_entities`, `source_relationships` | `SourceEntities`, `SourceRelationships` | `maximum_source_entities`, `maximum_source_relationships` |
| `property_definitions` | `PropertyDefinitions` | `maximum_property_definitions` |
| `projection_config.entity_mappings`, `projection_config.relationship_mappings` | `EntityMappings`, `RelationshipMappings` | `maximum_entity_mappings`, `maximum_relationship_mappings` |
| `projection_config.metadata_mappings`, `projection_config.aggregation_rules` | `MetadataMappings`, `AggregationRules` | `maximum_metadata_mappings`, `maximum_aggregation_rules` |
| `filters.entity_filters`, `filters.relationship_filters` | `EntityFilters`, `RelationshipFilters` | `maximum_entity_filters`, `maximum_relationship_filters` |

Keep declared-kind counts even though they use the mapping limits. Keep label, property-key, per-object, issue-budget, emission and final output checks. At each ceiling, require the baseline public outcome. At first excess, require the same code, reason/limit key, detail counts at limit and limit + 1, retry action and zero result. Include a malformed shape and two simultaneous overflows to fix error precedence. Exit when all nine pairs pass (F-030, V-018).

**S-027 — Correct source-ID failure precedence.** After S-024, give entity and relationship IDs one uniqueness check after shape and scalar admission and before the nine collection ceilings. Preserve the input-byte limit, shared issue budget, cancellation and closed error details. A duplicate in the last item or in a different collection from the overflow must report the same duplicate error as an isolated duplicate. With distinct IDs the original limit error must win. This changes the public reason only for competing violations and requires no data migration. Checkpoint F-033 with V-001/V-018 before S-025.

**S-025 — Give result content one builder.** Structural, preceded by S-027. Share the construction of `properties`, `mapped_metadata`, `schema_registry`, `vertices`, `edges`, `validation_summary` and `consumer_capabilities`. Add identity fields when constructing the envelope. Keep the producer's validation of the projected graph, the normalization step, per-row serialization and the independent completed-result proof. Exit on exact bytes and output order for all four S-022 result cases and stable error/cancellation behavior (F-031, V-001/V-017/V-018/V-028).

**S-026 — Correct completed-result provenance admission.** Behavior-changing, preceded by S-025 and the S-022 reproduction (E-083). Apply the family rules to `CompletedResultV2.Validate` and the authored `contracts/graph-projection/schemas.v2.json`. Direct vertices and edges require their typed source reference. Preserve the directed reverse-edge linkage and shared source/mapping proof. For aggregated objects validate the ordered contributor references against the §5 immediate-provenance rule. Use schema conditionals for the family-specific shapes. Verify contributor order and relationships between objects in the shared Graph proof. Regenerate through `make generate`. Exit when both isolated null-reference cases yield `ErrResultV2Invalid` before any insert and on content reads of seeded rows. Include exact and locked reads, vertex/edge reads and traversal. Verify that the metadata-only envelope lock still returns the binding, valid historical rows reproduce exactly, and validation rejects malformed or out-of-order aggregate provenance (F-032, V-001/V-019/V-028).

Revert a structural slice coherently if a characterized byte, ID or public error drifts. For the correction, an invalid stored result fails admission. Its source owner can rematerialize from authoritative input. Reverting the correction would reopen malformed admission. The accepted transition requires no stored-data rewrite, compatibility reader, Go API change or restore-version bump.

Retain the module topology. GP-02's no-topology-move conclusion remains valid. GP-03's completed slices below changed prose and evidence; the GP-04 plan adds bounded runtime correction and cleanup, explicitly authorized for GP-05 execution.

| Slice / kind | Findings / contracts | Depends on | Intended change and required scenarios | Readiness / status | Compatibility / rollback | Completion evidence |
| --- | --- | --- | --- | --- | --- | --- |
| S-001 / specification and documentation | F-001; C-005/C-010/C-011 | Bootstrap | Replace current Lease glossary with retained input and transaction admission locks; correct framework/domain rows; preserve historical and unrelated leases | ready / DONE | Editorial; no wire/generation migration; revert coherent prose if needed | E-015: owner review and V-007 passed |
| S-002 / implementation comment | F-004; C-006 | S-001 | Correct derived-table scope; metadata before clear, enumeration after clear, caller rollback before publication and writer commit after postconditions | ready / DONE | No executable change; no migration | E-016: comment-only diff reviewed |
| S-003 / service evidence and routing | F-005; C-006/C-012 | S-002 | Observe clear/enumerate/publish; metadata rejects before clear; clear failure prevents enumeration; source/semantic/binding/resource/cancellation failures skip publication; every failure is closed failed/incomplete with no committed facts; rename misleading tests | ready / DONE | Test selectors change atomically; no aliases or production hooks | E-017: V-010/V-003/V-011 passed |
| S-004 / transaction evidence and routing | F-005/RB-002; C-006/C-007/C-010/C-012 | S-003 | Nonempty real producer/publisher seed; enumeration failure/cancellation after actual clear; real NF binding mismatch with workers stopped; independent pre-cleanup exact snapshots prove rollback; runtime role, repeatable read, one transaction, no commit, usable connections | ready / DONE | Isolated fixtures only; keep existing success, writer failure and fatal/readiness assertions | E-018: V-012–V-014 passed |
| S-005 / validation and handoff | RB-001/T-009; all affected contracts | S-001–S-004 | Finalize, freeze source, bounded portfolio, final Markdown lint and evidence/baseline reconciliation | ready / DONE | No deployment/data migration; revert authored changes and regenerate projections | E-019–E-021: V-003/V-005/V-007/V-010–V-016 passed; handoff complete |

F-001 rationale: a single consumer-lifetime model prevents lease persistence/renewal compatibility from returning. F-004 rationale: clear scope and transaction ownership prevent destructive changes to authoritative declarations. F-005 rationale: meaningful call-order and committed-state evidence protects atomicity and readiness without new production abstractions. Leaving these gaps open risks incompatible APIs, incorrect clear scope, silent partial state, and false test assurance. S-001 review is semantic, not a global replacement of lease vocabulary. S-004 must avoid empty/count-only fixtures, assertions after cleanup, and live-worker races.

No executable runtime defect was exposed by S-003/S-004. That remediation closed specification/comment drift and missing evidence without a new production abstraction. The next iteration also preserves protocol generations, valid canonical identities, migration history and deployment state.

### GP-04 accepted slices and GP-05 readiness

Readiness is separate from status. GP-05 is authorized; E-031 recorded S-006 reproductions before correction. Each successor requires the preceding completion checkpoint. Current readiness/status is shown below; completed slice outcomes are in §10.

| Slice / kind | Finding / contracts | Depends on | Readiness / status | Areas and intended outcome | Completion gate |
| --- | --- | --- | --- | --- | --- |
| S-006 / owner-required correctness correction; authorized GP-05 | F-011; C-001/C-002/C-015 | GP-04 plan and authorization | ready / DONE | Implementation/tests: reject duplicates before keyed normalization and reject dependency cycles | Initial reproductions recorded; no partial result; valid ordering/digests/goldens unchanged; V-001/V-017–V-019 pass |
| S-007 / definition-integrity hardening; authorized GP-05 | F-012; C-006/C-016 | S-006 | ready / DONE | Implementation/tests: own registry slices and validate exactly one current binding | Mutation isolation and strict binding admission before clear proven; V-010/V-012/V-013/V-020 pass |
| S-008 / structural preservation | F-013; C-001/C-002 | S-007 | ready / DONE | Implementation/regression tests: delete unused private branches, values and redundant normalization | Semantic, identity, limit and cancellation checks unchanged and passing; V-001/V-017–V-019/V-025 |
| S-009 / consumer-boundary cleanup | F-014; C-008/C-010/C-017 | S-008 | ready / DONE | Implementation/tests/documentation/routing: relocate two Go errors, preserve public reasons, guard root values, correct retained-input wording | Direct/wrapped mappings, real provider failures, retained-input and cleanup-race evidence pass; routing/generated checks agree |
| S-010 / validation and handoff | F-015; C-007/C-011/C-012 | S-009; all prior slices recorded | ready / DONE | Finalize and freeze source, execute/audit bounded portfolio, reconcile this tracker | E-037–E-042: all required checks pass; no skipped gate is silent; compatible Graph owner evidence and final handoff recorded |

#### S-006 remediation, benefit and compatibility

Reject identical and conflicting duplicate `aggregation_rule_id` values before insertion into the ordering map; do not deduplicate them. Make dependency ordering fallible and remove the stalled-dependency fallback. Propagate a closed, non-retryable projection validation error with no result rather than an infrastructure/retry error. Preserve the existing ordering of valid ready batches, input permutation determinism and canonical identities; do not substitute a different valid topological ordering that changes digests.

Add characterization for duplicate IDs, multi-rule cycles, valid dependency chains and reordered equivalent input before correction, then retain the same cases as regression evidence. Verify no partial graph escapes failures. Initial failures must be observed, not invented from this document; if a case does not reproduce, re-inspect that finding before changing runtime behavior. Graph §§3-J/4 govern the correction; no owner amendment is proposed.

This prevents malformed configurations from silently becoming different successful computations or reconstructed results. Leaving the fallback risks incorrect output and false success. The compatibility impact is intentional rejection of previously accepted malformed input; do not retain aliases, rewrite declarations, reset environments or add conversion paths. Valid inputs and data formats do not migrate. An unpublished rollback is a coherent code/test revert; it must not be described as preserving the invalid behavior as a permanent contract.

#### S-007 remediation, benefit and compatibility

Deep-copy `SemanticQuerySchemaIDs` in every registry entry on construction and when returning `Document()` or `Registrations()`. Keep the clone specific and private; preserve the selected enumerator function values. The service may retain its registry pointer once all exported access paths are isolated. No serialization-based clone or generic registry framework is needed.

Replace the loop over `RecoveryGenerations` in binding admission with exact comparison to the packaged current definition and digest. Retain strict decoding, canonical content/digest verification, independently supported algorithm/schema checks, and all metadata rejection before clear. Matching only the caller-supplied digest is insufficient. Do not remove Core's generation registry or the frozen-binding entrypoint used by real composition.

Tests mutate constructor inputs, returned documents and returned registrations, then verify canonical contents, stored digest, callbacks and reconstruction behavior remain unchanged. Altered definitions with the original digest, rehashed altered definitions, malformed encodings and retired bindings must fail before clear. Existing successful reconstruction and transaction rollback remain required.

Owning admitted metadata prevents contents and digest from diverging; direct current-definition checks remove accidental compatibility pressure. The risks are alias-driven inconsistent admission and future acceptance of unintended generations. Public restore signatures and algorithm/result v5, registry v5 and binding v7 stay unchanged. No schema/data migration; rollback is the coherent implementation/test revert.

#### S-008 remediation, benefit and compatibility

Remove `allowAdmittedNumber` and its forwarding modes in canonical JSON/tuple helpers, retaining strict numeric handling. Remove only `projectionWork.NormalizedConfigurationSHA256` and `.NormalizedSourceSHA256` plus their assignments, not the live result/binding fields with those names. Remove ignored diagnostic arguments from private parsing helpers, the unused `run` argument to `validateProjectedGraph`, and the unused vertex/edge arguments to `buildSchemaRegistry`.

Read required `projection_version` after strict schema admission instead of retaining its unreachable `"1"` default. Delete inclusion-predicate initializations overwritten by `parseInclusionPredicate`. Consolidate duplicate entity/relationship mapping, metadata mapping and property-definition sorting under their normalization stage; retain source-object ordering, real defaults and dependency ordering corrected by S-006.

This reduces unsupported modes and implied dependencies without adding abstractions. The main risk is accidentally changing canonical order, legitimate optional defaults or numeric handling; preserve byte/identity goldens, finite-number/integer boundaries, failure atomicity, limits and cancellation. Existing `GP…1` framing strings remain current identity inputs. No Go public API, wire, schema or data migration; revert the private cleanup coherently if necessary. Do not write deletion-name tests or refresh expected golden values to hide drift.

#### S-009 remediation, benefit and compatibility

Move `ErrResultV2NotSelected` and `ErrResultV2SourceStale` from Graph's result ports to the existing `reporting/graphsourcecontract` package as `ErrNotSelected` and `ErrSourceStale`. Update the NF provider, Reporting mapper and tests atomically. Keep Graph's result-integrity/storage errors at their current owner. No aliases, dual error identities or string-matching compatibility logic remain.

Preserve Reporting's public `graph_projection_not_completed` and `graph_projection_stale` reasons, including wrapped errors through `errors.Is`. Add a positive Graph-root exported-value allowlist, not a growing list of retired names. Use existing owner rows for updated/new selectors and regenerate only through Make. Correct Reporting §15.2's “leased tuple-bound result” phrase to describe the exact admitted, retained tuple-bound result under REQ-RPT-099g; its later adopted amendment already resolves the obsolete lifetime wording.

This removes declaration policy from the pure result owner and makes the consumer handoff the explicit shared contract. Unresolved coupling encourages additional source-specific policy in Graph; careless movement risks incorrect public errors or admission order. Test real provider selection/staleness rejection alongside direct/wrapped mapper cases; preserve result-before-declaration locking, same-transaction validation, private retained input and cleanup-race behavior.

Two internal Go names move; public HTTP codes, semantics, result identity, storage and lifetime do not change. No data migration. Rollback must revert producers, consumers, tests, prose and routing together, then regenerate projections.

#### S-010 remediation, benefit and compatibility

After focused development checks, run `make agent-finalize`, inspect its outputs and freeze executable source/routing. Execute the final portfolio in §8, inspect owning process exits and same-run cleanup, and audit the complete required Graph owner union against explicit compatible roots. Keep peer results scoped to the selected rows. No earlier pass becomes current evidence merely because a selector name is unchanged.

Record commands, fresh run IDs, source identity, environment, outer exit codes, summaries, artifacts, cleanup results, skipped checks/reasons and remaining risks. Reconcile inventory, finding/contract links, slices, selectors, blockers and final criteria. Compare against the implementation session's own baseline while preserving concurrent changes. Complete this handoff last and describe bounded module hardening, not release certification.

The benefit is a reproducible readiness record instead of stale pass assumptions. Missing/incompatible evidence or any failed required gate leaves the relevant acceptance condition blocked. No deployment, migration, reset or release qualification is required; rollback consists of the coherent authored changes plus Make regeneration where routing changed.

### GP-06 accepted implementation details

S-011 clarifies scope-qualified producer identity, once-only dependency execution, and canonical source paths in the adopted specification before code changes. S-012 uses a private scope/kind index, validates unique IDs and vertex endpoint references, deduplicates dependencies, excludes own output and processes complete sorted ready batches; unresolved cycles return `projection_validation_failed/invalid_aggregation_rule/do_not_retry` with no result. S-013 gathers each rule's contributors before emitting, exposes completed output only to subsequent rules, and keeps available vertices/edges and endpoint lookup in one private state. Preserve grouping/identity framing, contributor/merge order, final sorting, limits and cancellation. S-014 centralizes recognition and context admission, implements exact source-kind names, and removes bare `kind` without aliases, translation or new projected-kind paths. Keep existing projected ID/property/mapped-metadata paths and unrelated kind discriminators/reservations.

For each correctness slice, first add canonical-path reproductions and record observed failures before repairing production. Add behavioral assertions for scope collisions, mixed chains, exact output/identities, endpoints, permutation equivalence, invalid references, atomic failure and reconstruction. S-014 covers every path-bearing position, canonical positive and wrong-context negative cases, and rejection before defaults can hide invalid paths. Existing goldens are not refreshed to conceal unexplained drift. Public Go/wire/storage signatures and restore v5/registry v5/binding v7 stay fixed. Rollback uses coherent authored reverts plus Make regeneration, preserving the starting diff; never rewrite stored bindings or reset data.

### GP-07 accepted implementation decisions

S-017 makes canonical encoding normalize `json.Number` through existing signed-int64/finite-number rules without relaxing typed integer admission. Grouping and endpoints share one fallible scalar tuple encoder; missing/null policies remain distinct, containers reject, safe errors propagate without hashing failed bytes. S-018 indexes configured producer input scopes by output scope/kind, including reverse outputs, and requires every applicable context for definitions/wildcards. Inert definitions retain lexical source/projected validation; graph-view reads remain graph-source-only. No aliases, selectors or translation.

S-019 owns direct/reverse/aggregate outputs and lookup indexes in one private state; capacity is checked before constructing the next emitted object, excludes do not consume capacity, and first-over-limit returns a closed non-retryable resource error with observed limit+1 and no result. Preserve cancellation during collection/emission and final defensive integrity checks. This is object-cardinality enforcement, not heap/performance certification.

S-020 compares four byte-string components: raw source (kind, ID, empty, empty); direct/reverse projected (referenced source kind, source ID, mapping ID, projected ID); aggregate projected (own kind, own ID, aggregation ID, own ID). No recursive ancestry. Merges and source refs use that order; `contributor_sort_key` is `csk_` plus lowercase UTF-8 hex components joined with `.`. Remove obsolete contributor hashing; retain object/result framing. Update only independently justified affected generic goldens.

All corrections first retain observed reproductions, then regressions for exact objects/bytes, permutations, errors and reconstruction. Preserve current NF default/time-bucket tuples. S-021 runs finalization then freezes executable source/routing; final V-008 covers focused Graph rows, with V-014/V-015/V-020–V-027 and V-003/V-005/V-007/V-011 as applicable. Generated outputs come only from Make. Optional retained-run maintenance is skipped without a compatible warm-check root. No Markdown enters executable tests/generation/evidence. Rollback is coherent authored reverts plus Make regeneration, preserving original staged work.

## 8. Validation Plan

### GP-09 verification

The planning-only tracker revision used V-007. The authorized implementation adds the following gates. E-082 onward records their results.

| Gate | Make selections from the repository root | Required evidence |
| --- | --- | --- |
| Baseline and structural comparisons | V-009 for owner discovery. V-001 pure determinism, V-017 semantic matrix, V-018 maximum bounds and V-019 contract projection use the exact row commands below. V-028 is `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340`. | Extend the existing pure parent `TestProjectV2GoldenDeterminismAndCompletedResult_Unit` and PostgreSQL parent `TestResultV2PublicationReadTraversalAndCleanup_Integration` for the provenance cases. Preserve the valid byte fixtures and record the malformed baseline separately. Check the numeric fixture through ordinary and locked reads and reprojection. |
| S-026 correction | V-001/V-017 for all families, V-019 for the authored schema, V-028 for publication and reads. Run `make generate` from authored changes, then V-003/V-011 for generated and schema drift. | Publish rejects before any envelope or child insert. Seed malformed rows independently of the corrected publisher so the read rejection actually exercises stored data. Valid replay, metadata locking and restore publication still work. |
| Final qualification | V-016 `make agent-finalize`, review its output and freeze source. Run V-008 `make test-slice OWNER=module.graphprojection`, then V-026 `make test-evidence-audit OWNER=module.graphprojection EVIDENCE_ROOTS_FILE=<explicit manifest>` against the compatible retained roots. Include V-014/V-023/V-027 for Network Flow, V-021/V-022 for Reporting and V-003/V-005/V-011/V-025 for static checks. | Confirm restore rows V-010/V-012/V-013 in the complete Graph union. Check both Network Flow bindings, exact Reporting admission and restore rollback/reconstruction. Capture the native exits, run roots and audit result on the final source. |

Use fresh run IDs and a non-Markdown evidence-roots manifest as described below. Use the full Graph selection for the final restore and storage obligations and record any uncovered selections separately. If an expected row or owner changes, use `make task-guide` to reconcile its selector before claiming coverage.

GP-06 reuses and extends the implementation obligations below; GP-05 outcomes remain historical. GP-04 ran only documentation lint (E-028); its restrictions and GP-03 results remain historical. Read-only Make discovery is not a product pass.

During implementation, run public Make commands from the repository root with fresh explicit `CARTULARY_TEST_RUN_ID` values and retain exact run roots. Revalidate row membership when adding or renaming tests. No product check, generator, runtime metadata, conformance or release evidence may read, stat, hash or otherwise depend on Markdown. Human review assesses specification alignment; selectors establish test selection, not specification completeness.

| Validation | Command | Scope / execution gate |
| --- | --- | --- |
| V-001 | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` | Required for S-006/S-008 and the S-009 root guard; GP-03 omitted this because its pure engine was unchanged |
| V-002 | `make test-slice OWNER=module.graphprojection SCOPE=services` | Broad storage selection not required; choose V-012/V-013 for affected seams |
| V-003 | `make source-check CHECK=generated` | Required after authored selector changes and in S-010; GP-03 passes remain E-017/E-020 history |
| V-004 | Not applicable | No rendered/browser behavior or selector changes |
| V-005 | `make lint CHECK=backend-boundary` | Required for S-009 boundary changes and S-010 |
| V-006 | `make check` | Not required: bounded remediation, no full-repository qualification claim |
| V-007 | `make lint-markdown` | Sole verification target for this document update; also required for later authored documentation and final handoff. GP-03 passes remain historical |
| V-008 | `make test-slice OWNER=module.graphprojection` | Required final Graph owner portfolio for S-010; current routing has 11 rows/43 selectors (one positive exported-value guard added). Not run or required in GP-03 |
| V-009 | `make task-guide OWNER=module.graphprojection DETAIL=rows`; `make task-guide OWNER=module.networkflow`; `make task-guide OWNER=module.reporting`; `make task-guide OWNER=app.server` | Read-only owner discovery; E-027 records which commands ran in GP-04, E-014/E-011 and prior inspection record historical discovery |
| V-010 | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_deterministic_current_restore_rebuild` | Service order, closed failures and boundary guards; 14 top-level selectors |
| V-011 | `make source-check CHECK=schema` | Authored selector validity |
| V-012 | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.claimed_graph_reconstruction_enumerates_and_publ_c18fddd44a` | Assembled clear, source failure/cancellation rollback and successful empty-source clear; one selector |
| V-013 | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.restore_publication` | Existing writer publication/reconciliation/postcondition/indeterminate coverage; seven selectors |
| V-014 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2` | Real binding failure, successful saved-graph lifecycle, retained Reporting input and authority regression; three selectors |
| V-015 | `make test-slice OWNER=app.server ROWS=app.server.unit.post_claim_graph_activation_preserves_structural_f41e41a052` | Readiness/fatal outcomes; one selector |
| V-016 | `make agent-finalize` | Required before final source qualification; review outputs and freeze executable source |
| V-017 | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_semantic_matrix` | S-006/S-008 aggregation validity, ordering and semantic regression |
| V-018 | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_maximum_semantic_bounds` | S-006/S-008 resource bounds and cancellation regression |
| V-019 | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.v2_contract_projection` | S-006/S-008 canonical contract/fixture identity regression |
| V-020 | `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c` | S-007 and final current-generation restore/binding integration |
| V-021 | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.graph_v2_exact_selection_and_rendering_f54d8b1a62` | S-009 and final direct/wrapped sentinel mapping and retained-input unit coverage |
| V-022 | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.graph_input_admission_reads_the_exact_selected_c_4c42af0393` | S-009 and final exact selected input admission |
| V-023 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.bounded_graph_aggregation_and_selectors` | S-009 and final consumer aggregation/selection regression |
| V-024 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.graph_result_cleanup_races` | S-009 and final result cleanup/publication/admission races |
| V-025 | `make lint CHECK=go` | Focused implementation hygiene and S-010 final Go validation |
| V-026 | `make test-evidence-audit OWNER=module.graphprojection EVIDENCE_ROOTS_FILE=<explicit manifest>` | S-010: audit exact compatible Graph owner roots supplied through the non-Markdown manifest described below |
| V-027 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.time_bucket_graph_backend,module.networkflow.integration.time_bucket_saved_graph_lifecycle` | GP-06: time-bucket identity fixtures, projection/selector binding and saved-result lifecycle; required in S-014/S-015 because V-014/V-023 do not select these tests |

Development checkpoints use V-001/V-017–V-019 for S-006/S-008; V-010/V-012/V-013/V-020 for S-007; and V-001/V-014/V-021–V-024 plus affected static/routing checks for S-009. Update authored selectors atomically when cases are added or renamed. Use `make generate` after authored routing changes; never hand-edit projections. Use repository Make formatting/finalization only during authorized implementation.

For S-010 and S-015, run V-016, review its output, then freeze executable sources/routing. S-015 additionally requires V-027 for time-bucket bindings and lifecycle. Obtain compatible final results for V-008, V-014/V-015, V-020–V-025, V-003/V-005/V-011 and V-007. The full Graph owner run includes the focused Graph rows; do not rerun those solely to duplicate evidence. Earlier results count only where the harness establishes compatibility with the frozen source and complete required row/selector coverage. Diagnose failures before rerunning or broadening. Record outer exits separately from summaries and inspect same-run cleanup and missing outputs.

Supply V-026 an explicit JSON evidence-root manifest in a non-source location, using exact retained roots from successful compatible invocations. Its shape is `{"schema_id":"cartulary.harness_evidence_root_manifest.v1","owner_id":"module.graphprojection","entries":[{"target":"test-slice","run_root":"/absolute/exact/final-graph-run-root"}]}`. Replace the illustrative root with observed evidence; add entries only when required to complete the compatible union. Do not discover or synthesize roots from this Markdown tracker. Audit success establishes the required Graph owner evidence union, not semantic completeness or release certification.

GP-03's S-005 portfolio covered 26 top-level tests across five rows; its exclusion of a complete-owner audit applies to that historical scope. GP-05's S-010 explicitly requires V-008/V-026. Browser, visual, performance, deployment, full-repository and release qualification remain outside both bounded iterations. Do not run a broad warm check solely to obtain a `RESULTS_DIR` for optional finalization maintenance.

For S-021, V-016 precedes the freeze. The final portfolio is V-008; V-014/V-023/V-024/V-027 (NF); V-021/V-022 (Reporting); V-020 (Recovery); V-015 (server); V-003/V-011/V-025/V-005; the explicit V-026 audit; and V-007 on the final tracker text. Do not duplicate the focused Graph rows covered by V-008. Exact commands and results are recorded in the GP-07 handoff.

## 9. Top-Level Work Tracker

**GP-09 status.** T-035 is the completed planning checkpoint. The entry gate is complete (E-082). Section 12 records the current status and evidence for each workstream. E-083–E-092 qualified the original candidate. E-093–E-103 qualified the second candidate. E-104–E-116 qualify the third candidate. GP-11 checkpoints and §12 govern current readiness.

Historical T-001–T-008 remain planning completions; T-009–T-017 retain GP-03 implementation evidence. T-018 was the GP-04 document update. T-019–T-023 retain GP-05 completion history; T-024–T-028 track GP-06.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Normalize graphprojection, establish output admission and authority order | WF-00 | DONE | None | §1, E-001/E-002/E-012 | Correct literal target/output and allowed-write scope recorded. |
| T-002 | Inventory 36 authored files across five Graph packages | WF-01 | DONE | T-001 | §2, E-003–E-005 | Each target file has one inventory row and an explicit owner/risk posture. |
| T-003 | Map v2 identities, completed-result proof, storage and reconstruction contracts | WF-02 | DONE | T-002 | C-001–C-007/C-011 | Governing owners, current code and existing tests distinguished. |
| T-004 | Map indirect NF routes, authorization and Reporting retained input; exclude direct workbook/browser surfaces | WF-02 | DONE | T-002 | C-008–C-010/C-013/C-014, E-006–E-008 | Indirect behavior and applicability limits are explicit. |
| T-005 | Account for 11 Graph rows/41 selectors and helper-parent test relationships | WF-03 | DONE | T-003, T-004 | C-012 and §4 accounting | Catalog totals and assertion limits reconcile; no passing result invented. |
| T-006 | Diagnose root/adapters/source-owner/application boundaries | WF-04 | DONE | T-002, T-003, T-004 | §3, F-001–F-010 | Retention finding grounded in adopted boundary and live implementation. |
| T-007 | Discover public validation commands and constrain artifact-producing execution | WF-07 | DONE | T-005, T-006 | §8, E-010/E-011, RB-001 | Commands and execution posture documented without unauthorized side effects. |
| T-008 | Complete tracker-only handoff and integrity review | WF-08 | DONE | T-007 | E-012/E-013, §§10–12 | All twelve sections reconciled and final non-output baseline matches. |
| T-009 | Run canonical documentation lint | WF-13 | DONE | T-007 | V-007, RB-001, E-020 | GP-03 final V-007 passed. S-001 passed E-015; GP-02 skip remains historical. |
| T-010 | Reconcile current lease vocabulary with adopted owners | WF-09 | DONE | T-006 | F-001 | S-001/E-015 corrected supporting prose without changing owner semantics. |
| T-011 | Correct stale four-table Writer.Clear comment | WF-10 | DONE | T-006 | F-004 | S-002/E-016 aligned the comment with SQL and the current contract. |
| T-012 | Add observable Clear/enumeration/rollback characterization | WF-11/WF-12 | DONE | T-005 | F-005, RB-002 | S-003/S-004 demonstrate clear/order/failure rollback; close with both service and database evidence. |
| T-013 | S-001 specification/vocabulary reconciliation | WF-09 | DONE | Bootstrap; F-001 | E-015 | V-007 and owner review pass |
| T-014 | S-002 restore comment | WF-10 | DONE | S-001; F-004 | E-016 | Executable diff empty |
| T-015 | S-003 service sequencing evidence | WF-11 | DONE | S-002; F-005 | E-017 | V-010/V-003/V-011 pass |
| T-016 | S-004 real assembly rollback evidence | WF-12 | DONE | S-003; F-005/RB-002 | E-018 | V-012–V-014 pass |
| T-017 | S-005 final validation and handoff | WF-13 | DONE | S-001–S-004 | E-019–E-021 | Final portfolio and handoff complete |
| T-018 | Record GP-04 cleanup/hardening plan and document-only validation | Planning entry for WF-14–WF-18 | DONE | Clean baseline E-022; accepted plan | E-022–E-029; §§1–12 | Twelve sections/history preserved; every new finding mapped; lint passes; only tracker content changes |
| T-019 | S-006 duplicate-safe, acyclic aggregation admission | WF-14 | DONE | T-018; GP-05 authorization | F-011; C-015; planned §7 checks | Initial reproductions then strict rejection, atomic failures and unchanged valid canonical identities |
| T-020 | S-007 registry ownership and current-only binding admission | WF-15 | DONE | T-019 | F-012; C-016 | Alias isolation, strict actual-definition validation before clear and restore/Recovery regressions pass |
| T-021 | S-008 private engine plumbing cleanup | WF-16 | DONE | T-020 | F-013; C-001/C-002 | Evidenced remnants removed; numeric/default/order/identity/bounds/cancellation behavior preserved |
| T-022 | S-009 consumer-owned source errors and retained-input wording | WF-17 | DONE | T-021 | F-014; C-017 | No Graph aliases; root guard, reason mapping, provider/retention/race checks and generated routing pass |
| T-023 | S-010 final validation, evidence audit and handoff | WF-18 | DONE | T-019–T-022 | F-015; §8 final portfolio; E-037–E-042 | Final source frozen; compatible checks/audit pass; artifacts, cleanup, scope limits and final statuses recorded |
| T-024 | S-011 owner and acceptance reconciliation | WF-19 | DONE | Follow-up authorization | F-016–F-018; E-043/E-044 | Owner clauses, baseline, compatibility and reopened gate recorded; lint passes |
| T-025 | S-012 scope-aware dependencies | WF-20 | DONE | T-024 | F-016; E-045–E-047 | Collision reproduction and deterministic dependency/admission regressions pass |
| T-026 | S-013 dependency-ordered execution | WF-21 | DONE | T-025 | F-017; E-048/E-049 | Nonempty mixed chains, exact reconstruction and rollback checks pass |
| T-027 | S-014 canonical field paths | WF-22 | DONE | T-026 | F-018; E-050 onward | Canonical-context/alias matrix and unchanged production bindings pass |
| T-028 | S-015 final qualification and handoff | WF-23 | DONE | T-024–T-027 | E-054–E-059; final portfolio and explicit owner audit | Combined source qualified, preserved changes verified, tracker completion last |

GP-07 top-level tasks T-029–T-034 use the combined workflow/slice/task table in §6 as their status of record. The per-slice exit evidence is indexed in §12.

## 10. Session Handoff Log

### GP-09 execution checkpoints

**E-082 — Specification entry, DONE.** Execution began from `main` at
`9203107cd76ba2849bf02dc466309b6697be3e93`, with a clean index and one
modified file, this tracker (initial SHA-256
`d7706ca977978f30cb932f0f2024614acd5bf3736128df4f2bedc348db23a009`).
The user selected duplicate-first admission and authorized the sequential plan.
The owner change is in `docs/graph_projection_nlspec.md`; this tracker records
the gates. The production tree at entry is unchanged from GP-08.

| Owner clause | Projection and validation boundary |
| --- | --- |
| Graph §4, C-001 / F-033 | Graph input admission returns the duplicate error before a collection ceiling, with separate ID namespaces, safe details and the shared issue budget. V-001/V-018 exercise the public result. |
| Graph §§5–6.3, C-002/C-003 / F-032 | The authored v2 JSON Schema projects family-local shapes. The Graph proof checks mapping equality, reverse linkage, immediate contributor identity and order. V-001/V-017/V-019 exercise both layers. |
| Graph §8, C-004/C-010/C-011 | Borrowed publication and every content read use that proof. The metadata-only lock returns a binding. Network Flow retains rematerialization ownership and Reporting owns frozen inputs. V-028 and the §8 peer rows cover these boundaries. |

The user reviewed and accepted this written mapping. `make task-guide OWNER=module.graphprojection` exited 0. `CARTULARY_TEST_RUN_ID=gp09-owner-entry-01 make lint-markdown` exited 0 with a passing [receipt](../../.cartulary/test-results/gp09-owner-entry-01/adhoc/lint-markdown/tool-run-summary.json). The final entry-text run `gp09-owner-entry-02` also exited 0 with a passing [receipt](../../.cartulary/test-results/gp09-owner-entry-02/adhoc/lint-markdown/tool-run-summary.json).

**E-083 — S-022 DONE.** On unchanged production source, both fully rebound null-reference vectors passed pure validation, borrowed publication and exact reads. Duplicate-only input reported duplicate_identifier, but adding a filter overflow made maximum_entity_filters win for either source collection. Both source and seven other collection ceiling/first-excess witnesses passed. The four valid vectors pin bindings, canonical envelopes and ordered child rows. The historical numeric and both Network Flow binding witnesses passed. RB-003 is closed. Changed files include Graph root engine_v2_test.go, input_counts_v2_test.go, result_vectors_v2_test.go, postgresresult/result_proof_test.go and postgresresult/testdata/result-vectors-v2.json.

  - `CARTULARY_TEST_RUN_ID=gp09-s022-capture-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 0, pass. Source `a34d1e5d30bb2402c0cc0a2c99fb4c2381259ea18cc6eb7fd31543f08f371e4e`. [Retained run](../../.cartulary/test-results/gp09-s022-capture-01/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s022-graph-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_maximum_semantic_bounds,module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0, pass. Source `2cc2b6d0fc75e2ea514fdbb02651af7979bbaa7151128a7b15653471b3662c85`. [Retained run](../../.cartulary/test-results/gp09-s022-graph-01/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s022-nf-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.time_bucket_saved_graph_lifecycle,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend` exited 0, pass. Source `2cc2b6d0fc75e2ea514fdbb02651af7979bbaa7151128a7b15653471b3662c85`. [Retained run](../../.cartulary/test-results/gp09-s022-nf-01/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s022-options-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 0, pass. Source `be521ec5b85cb2d2050bf927cdfb0b5dc72256fecefb1610e2beb8849ad74134`. [Retained run](../../.cartulary/test-results/gp09-s022-options-01/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s022-fixed-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 0, pass. Source `61a9edd257f59315c38ff037dc863129b399372a45e4797e389e487d41d68488`. [Retained run](../../.cartulary/test-results/gp09-s022-fixed-01/run-manifest.json)

**E-084 — S-023 DONE.** The digest path now receives ordinary maps and uses the single sorted-key writer. Optional default presence, nested predicate, reverse kind, number spelling, valid result bytes and the stored historical fixture reproduced E-083. The structural change preserved both malformed baseline observations. Changed files include Graph root canonical.go, objects.go and engine_v2.go; Make formatting also aligned S-022 test declarations.

  - `CARTULARY_TEST_RUN_ID=gp09-s023-format-01 make format` exited 0, pass. Source `1a2eabe2fa8ba84ce0f1dede53d7aa3121c8f69d04f5c12608aba09136b5ad78`. [Retained run](../../.cartulary/test-results/gp09-s023-format-01/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s023-graph-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0, pass. Source `fa3fef5bf4c789a64c44e6920fb4eb71d13ded8f8b6521d743297033ca530262`. [Retained run](../../.cartulary/test-results/gp09-s023-graph-01/run-manifest.json)

**E-085 — S-024 DONE.** Exactly nine dominated typed counts were removed. Real ceilings, first excess, malformed shape, two simultaneous overflows and the existing duplicate/overflow behavior matched E-083. Declared-kind, object, issue-budget, emission and final result guards remain in their existing phases. Changed files include Graph root validation.go.

  - `CARTULARY_TEST_RUN_ID=gp09-s024-graph-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_maximum_semantic_bounds,module.graphprojection.engine.v2_pure_determinism` exited 0, pass. Source `85b4addaae344f9b10e3c4b34106059fb47238e451db950ca2fbe86c1982a11c`. [Retained run](../../.cartulary/test-results/gp09-s024-graph-01/run-manifest.json)

**E-086 — S-027 DONE.** Before the fix, CARTULARY_TEST_RUN_ID=gp09-s027-red-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism exited 2 (product/test_assertion_failure, [summary](../../.cartulary/test-results/gp09-s027-red-01/run-summary.json)). Source admission now checks scalars and counts duplicate IDs before collection verdicts. In-place source ordering avoids an extra ID index. Tests cover both final-position duplicates above their own ceilings, both cross-collection cases, distinct overflows, independent namespaces, scalar precedence, cancellation and the exact combined issue budget. Inputs with several invalid phases now report the earlier identity phase before downstream diagnostics. Changed files include Graph root input.go, engine_v2.go, validation.go, validation_registry.go, project.go, input_counts_v2_test.go and the private parser calls in engine_v2_test.go, aggregation_v2_test.go and execution_bounds_v2_test.go.

  - `CARTULARY_TEST_RUN_ID=gp09-s027-graph-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_maximum_semantic_bounds,module.graphprojection.engine.v2_pure_determinism` exited 0, pass. Source `084ff12fc2b2f8195454e3cc945a56f5149abbca145b6194e4a03f2f8b83a936`. [Retained run](../../.cartulary/test-results/gp09-s027-graph-01/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s027-format-01 make format` exited 0, pass. Source `084ff12fc2b2f8195454e3cc945a56f5149abbca145b6194e4a03f2f8b83a936`. [Retained run](../../.cartulary/test-results/gp09-s027-format-01/run-manifest.json)

**E-087 — S-025 DONE.** Semantic derivation constructs the existing result content once. Producer validation serializes the same seven fields as the later normalized digest and envelope. All S-022 byte vectors, cancellation, stored-row corruption, historical reads and maximum-result checks passed. Changed files include Graph root project.go and engine_v2.go.

  - `CARTULARY_TEST_RUN_ID=gp09-s025-graph-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_maximum_semantic_bounds,module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0, pass. Source `e472b1bb525f895db39ed8b04464888e8b8006ef7da8c0887de409a92f9f255c`. [Retained run](../../.cartulary/test-results/gp09-s025-graph-01/run-manifest.json)

**E-088 — S-026 DONE.** Family-local schema and Graph validation reject both rebound direct cases. The shared proof verifies referenced contributor identities, rule linkage and §5 key order after indexing all object rows. A bounded dependency graph shares the existing ready-batch algorithm with input admission, rejecting circular producer claims without reconstructing ancestry. Borrowed transactions contained zero matching rows in all three tables after rejection. Separately seeded malformed rows failed exact, locked exact, vertex, edge and traversal reads while the metadata lock returned the binding. Historical read/reprojection and raw, projected, reverse and chained positive cases passed. Changed files include contracts/graph-projection/schemas.v2.json, Make-generated internal/gen/contractgraphprojection/artifacts_gen.go, Graph root result_v2_validation.go, engine_v2.go and the engine, contributor-order, result-validation, result-vector, contract-projection and PostgreSQL proof/historical tests.

  - `CARTULARY_TEST_RUN_ID=gp09-s026-format-01 make format` exited 0, pass. Source `a0557b5100a9b47d792c1c9ff16e8071051ce56bc86d20712ceed7d655069fb5`. [Retained run](../../.cartulary/test-results/gp09-s026-format-01/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s026-graph-02 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340,module.graphprojection.storage.restore_publication,module.graphprojection.storage.v2_contract_projection` exited 0, pass. Source `654be733ee2ff619f29f712ec8e1e6128ab0bc23e96fcf5d8980191a93bb6c04`. [Retained run](../../.cartulary/test-results/gp09-s026-graph-02/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s026-generated-01 make source-check CHECK=generated` exited 0, pass. Source `1cff5f44239f114df9d6a1634c28c1c8035bbc2f98ec01d86442c3cf0bdb740c`. [Retained run](../../.cartulary/test-results/gp09-s026-generated-01/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s026-schema-01 make source-check CHECK=schema` exited 0, pass. Source `1cff5f44239f114df9d6a1634c28c1c8035bbc2f98ec01d86442c3cf0bdb740c`. [Retained run](../../.cartulary/test-results/gp09-s026-schema-01/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s026-format-02 make format` exited 0, pass. Source `20ba97bd34fe814fd936a42ce10ff6d83be201048d6e9339a61e9932e943f541`. [Retained run](../../.cartulary/test-results/gp09-s026-format-02/run-manifest.json)
  - `CARTULARY_TEST_RUN_ID=gp09-s026-graph-03 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340,module.graphprojection.storage.restore_publication,module.graphprojection.storage.v2_contract_projection` exited 0, pass. Source `7201d1597df7d491bb447600a5d923ddf91339e2dfc33e0d1f4f1fb98c9bbccb`. [Retained run](../../.cartulary/test-results/gp09-s026-graph-03/run-manifest.json)

S-026 reproduction and repair history (all paths under `.cartulary/test-results`):

- `gp09-s026-red-01`, cache off, V-001/V-019, exit 2, product assertion failure. The runtime accepted rebound direct and aggregate defects. The first schema test used the Go default regular-expression engine, which could not compile the existing ECMAScript identifier pattern. The test now uses the repository-pinned ECMAScript engine with a timeout.
- `gp09-s026-schema-red-01`, cache off, V-019, exit 2, product assertion failure. The original packaged schema accepted both null-reference fixtures and an unknown aggregate contributor type.
- `CARTULARY_TEST_RUN_ID=gp09-s026-generate-01 make generate` exited 0. Only the packaged Graph artifact changed. [Generation receipt](../../.cartulary/test-results/gp09-s026-generate-01/generate/tool-run-summary.json)
- `gp09-s026-graph-01`, same selection as `gp09-s026-graph-02`, exit 2, harness `infra/preflight_error`. A Go local variable redeclaration prevented package compilation and service preparation. E-088 records the passing selection after the naming correction. The failed diagnostic request with `TARGET=backend-unit` was corrected to the public `TARGET=test-slice`.
- `gp09-s026-cycle-red-01`, cache off, V-001, exit 2, product assertion failure. Three locally consistent aggregate contributor links formed a cycle with no possible producer order. The same witness and the valid multistage result pass in `gp09-s026-graph-03`.

**E-089 — Finalization and executable-source freeze.**
`CARTULARY_TEST_RUN_ID=gp09-finalize-01 make agent-finalize` exited 0. Schema,
tier coverage and generated refresh passed. The optional warm-check evidence
and scheduler actions were skipped because no compatible full warm-check
root was supplied. Before and after inventories matched for all 7,655
tracked or non-ignored non-document source entries, including file modes
and symlink targets. The manifest SHA-256 is
`109004d9eeade960f9128b5ef7d341d65043f43289f90ea70d5a79197b04b454`.
[Finalize actions](../../.cartulary/test-results/gp09-finalize-01/unit-artifacts/finalize-summary.json) · [Frozen source](../../.cartulary/evidence-manifests/graph-projection-gp09/frozen-source.json)

**E-090 — Frozen-source qualification.** All commands below ran from the
repository root on harness source `7201d1597df7d491bb447600a5d923ddf91339e2dfc33e0d1f4f1fb98c9bbccb`.
Every native command exited 0. All reported cleanup steps completed and all
same-run retained-secret scans passed. The final Graph selection covered
eleven rows, including the real borrowed-publication, restore writer and
assembled rollback/reconstruction rows. NF covered both production bindings,
saved-graph admission and lifecycle. Reporting covered exact admission and
its frozen execution input.

| Obligation | Command | Native exit and result | Retained root |
| --- | --- | --- | --- |
| V-008, including V-010/V-012/V-013/V-028 | `CARTULARY_TEST_RUN_ID=gp09-final-graph-01 make test-slice OWNER=module.graphprojection` | 0/pass, 11 work units, 29 cleanup steps | [Run](../../.cartulary/test-results/gp09-final-graph-01/run-summary.json) |
| V-014/V-023/V-027 | `CARTULARY_TEST_RUN_ID=gp09-final-nf-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.time_bucket_saved_graph_lifecycle,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend` | 0/pass, 4 work units, 15 cleanup steps | [Run](../../.cartulary/test-results/gp09-final-nf-01/run-summary.json) |
| V-021/V-022 | `CARTULARY_TEST_RUN_ID=gp09-final-reporting-01 make test-slice OWNER=module.reporting ROWS=module.reporting.integration.graph_input_admission_reads_the_exact_selected_c_4c42af0393,module.reporting.unit.graph_v2_exact_selection_and_rendering_f54d8b1a62` | 0/pass, 4 work units, 15 cleanup steps | [Run](../../.cartulary/test-results/gp09-final-reporting-01/run-summary.json) |
| V-003 | `CARTULARY_TEST_RUN_ID=gp09-final-generated-01 make source-check CHECK=generated` | 0/pass, 5 work units, 14 cleanup steps | [Run](../../.cartulary/test-results/gp09-final-generated-01/run-summary.json) |
| V-011 | `CARTULARY_TEST_RUN_ID=gp09-final-schema-01 make source-check CHECK=schema` | 0/pass, 3 work units, 12 cleanup steps | [Run](../../.cartulary/test-results/gp09-final-schema-01/run-summary.json) |
| V-025 | `CARTULARY_TEST_RUN_ID=gp09-final-go-01 make lint CHECK=go` | 0/pass, 6 work units, 15 cleanup steps | [Run](../../.cartulary/test-results/gp09-final-go-01/run-summary.json) |
| V-005 | `CARTULARY_TEST_RUN_ID=gp09-final-boundary-01 make lint CHECK=backend-boundary` | 0/pass, 4 work units, 13 cleanup steps | [Run](../../.cartulary/test-results/gp09-final-boundary-01/run-summary.json) |
| V-026 | `CARTULARY_TEST_RUN_ID=gp09-final-audit-01 make test-evidence-audit OWNER=module.graphprojection EVIDENCE_ROOTS_FILE=.cartulary/evidence-manifests/graph-projection-gp09/graph-owner-roots.json` | 0/pass, 1 work units, 10 cleanup steps | [Run](../../.cartulary/test-results/gp09-final-audit-01/run-summary.json) |

**E-091 — Complete Graph owner audit.** The audit accepted all eleven required
rows on one compatible partition from the explicit final Graph run. Rejected
and unused input lists were empty. [Audit](../../.cartulary/test-results/gp09-final-audit-01/test-evidence-audit/audit.json) · [Root selection](../../.cartulary/evidence-manifests/graph-projection-gp09/graph-owner-roots.json) · [Native exits](../../.cartulary/evidence-manifests/graph-projection-gp09/native-exits.json)

**E-092 — Compatibility and handoff.** The structural slices preserve current
v2 bytes, historical numeric rows and both Network Flow bindings. The two
intentional corrections affect invalid inputs. Source-ID errors precede a
collection overflow, and malformed stored provenance now fails closed with
`ErrResultV2Invalid`. This includes obsolete aggregate contributor encodings
that cannot satisfy the current §5 key rule. Network Flow can rematerialize
from authoritative declarations. Reconstruction requires an exact selected
binding and rolls back on mismatch. Existing Reporting admissions use their
retained immutable input.

Rollback the structural slices independently by restoring the private input
builders (S-023), the nine typed count guards (S-024), or the producer's separate
content builder (S-025). Reverting S-027 changes error precedence. Reverting
S-026 reopens malformed admission and requires regeneration of the Graph
package from the matching authored schema. No data rewrite, API migration
or Recovery version change is required. The result qualifies Graph and its
named consumers. Deployment and release require separate evidence.

WF-35 / T-041 is the last checkpoint. The starting tracker edits and all
twelve sections remain in place. Git reports 20 modified and three new paths
against HEAD, all in the approved owner, tracker, Graph package, fixture or
contract surfaces. The index matches HEAD. The final source inventory
matches E-089. `git diff --check` passed. The handoff draft lint exited 0.
The final-text command is `CARTULARY_TEST_RUN_ID=gp09-handoff-final-01 make lint-markdown`.
[Draft lint](../../.cartulary/test-results/gp09-handoff-lint-01/adhoc/lint-markdown/tool-run-summary.json) · [Final-text receipt](../../.cartulary/test-results/gp09-handoff-final-01/adhoc/lint-markdown/tool-run-summary.json)

### GP-09 repair checkpoints

**E-093 — Repair entry.** At `9203107cd` the index tree was `7e206ded2ca9fe6b7c4e92d2088fedf28193ba8f`. The 20 modified and three untracked paths matched the reviewed candidate. The [entry inventory](../../.cartulary/evidence-manifests/graph-projection-gp09-repair/entry-source.json) matched all 7,655 non-document entries of the E-089 freeze. SHA-256 of the tracked binary diff was `fdcd3a655884b13ab854e9a0a9ce28571a2f5ca593a3e937bf372cd3da54111e`, of the tracker `f4d411a3aee58d137440f0e520b65a7a013a624e6db0410d2609f5a59bd0390c` and of the Graph owner `807fa04b993d11e6f0974be42e1db8fd59cbc5762a34a4d5d0a6ef038089f985`. The historical numeric fixture was `61ab450789d91a8e14cba9e30cf709e378ff8edc019c1630b7f01abe6793b429`, the four-family vector file `16b7553bd6a276c16893a1d23172a93fcffcecde353133e5325228bebee207b6`.

| Adopted owner | Projection and acceptance boundary |
| --- | --- |
| §§3-J, 5, 6.3 and 8 | Graph owns local registry and contributor shapes in its authored schema. Its pure result proof checks scope-qualified membership, endpoint producers, contributor order and acyclicity at publication and content reads. The metadata-only lock returns the binding. |
| §§3–4 and 7 | Shape precedes source scalar validation. Invalid source IDs or labels win over direction errors. A lone invalid direction wins over duplicate IDs and count excess with its existing direction-specific reason and a count of one. Endpoint existence is a later check. |
| §§4–5 | Source uniqueness observes the item cancellation cadence, including on oversized input. Raw keys have two trailing empty components and projected keys have four nonempty components. Source owners retain rematerialization and Reporting retains admitted inputs. |

Entry DONE. The user reviewed and accepted the amended owner-to-contract mapping. Changed paths were the Graph NLSpec and this tracker. `CARTULARY_TEST_RUN_ID=gp09r-entry-guide make task-guide OWNER=module.graphprojection` exited 0 and returned the Graph owner selector (discovery has no run root). `CARTULARY_TEST_RUN_ID=gp09r-entry-lint make lint-markdown` exited 0, [entry lint receipt](../../.cartulary/test-results/gp09r-entry-lint/adhoc/lint-markdown/tool-run-summary.json). No entry blockers.

**E-094 — S-028 / F-034 DONE.**

The pre-fix Graph proof accepted a rebound invented leaf producer, unlisted direct and reverse source kinds, duplicate declaration ownership and unsorted source kinds. The packaged schema accepted a missing producer list. The scope-qualified registry now rejects these cases, including an otherwise unused rule from a different scope. Empty-output declarations, shared output kinds and the fixed valid bytes passed. Changed Graph root result_v2_validation.go, result_v2_validation_test.go, engine_v2_test.go and v2_contract_projection_test.go, authored contracts/graph-projection/schemas.v2.json and Make-generated internal/gen/contractgraphprojection/artifacts_gen.go. No remaining S-028 blockers. The independent seeded storage witnesses for both proof changes follow in S-029.

- `CARTULARY_TEST_RUN_ID=gp09r-028-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.storage.v2_contract_projection` exited 2 (fail). Source `951bd432621d07632609ec2ad6ba040cb60b3d77ed814cc423ad4a7db4d4b3d4`. [Receipt](../../.cartulary/test-results/gp09r-028-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-028-generate make generate` exited 0 (pass). [Receipt](../../.cartulary/test-results/gp09r-028-generate/generate/tool-run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-028-format make format` exited 0 (pass). Source `fdc59680bd6eed7b11c9a03fc949dee5b64953d450ff0bb4494493d1dd064017`. [Receipt](../../.cartulary/test-results/gp09r-028-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-028-graph make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.v2_contract_projection,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0 (pass). Source `aac3ec7748d1064c910efb327622b85932a04992d00cb82a3a73b9820613a9cd`. [Receipt](../../.cartulary/test-results/gp09r-028-graph/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-028-generated make source-check CHECK=generated` exited 0 (pass). Source `aac3ec7748d1064c910efb327622b85932a04992d00cb82a3a73b9820613a9cd`. [Receipt](../../.cartulary/test-results/gp09r-028-generated/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-028-schema make source-check CHECK=schema` exited 0 (pass). Source `aac3ec7748d1064c910efb327622b85932a04992d00cb82a3a73b9820613a9cd`. [Receipt](../../.cartulary/test-results/gp09r-028-schema/run-summary.json)

**E-095 — S-029 / F-035 DONE.**

Before repair all four direct/cyclic endpoint mutations passed the pure proof. The borrowed publisher also admitted the rebound bridge-to-final cycle. After repair both endpoint producers enter the existing dependency graph. Pure, semantic and PostgreSQL gates passed, including the fixed acyclic chain with a repeated endpoint. The invented tail producer and endpoint cycle wrote zero envelope or child rows. Independently seeded bytes failed ordinary and locked exact, vertex, edge and traversal reads. The metadata-only lock returned each binding. Changed Graph root result_v2_validation.go, result_v2_validation_test.go, engine_v2_test.go and postgresresult/result_proof_test.go. No S-029 blockers.

- `CARTULARY_TEST_RUN_ID=gp09r-029-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 2 (fail). Source `6b4c44120472e6be3bcbb8ee8d09ffd2ee15d31c01169eaab24bb9240e5622ee`. [Receipt](../../.cartulary/test-results/gp09r-029-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-029-format make format` exited 0 (pass). Source `417a98633227348542e08779b17d107467a1bd970f090d284795ae8eef755fb6`. [Receipt](../../.cartulary/test-results/gp09r-029-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-029-graph make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0 (pass). Source `03a48f4638601b7737895df1522e62fcdc3a6bdd5fc02e2b6de5eea365df5439`. [Receipt](../../.cartulary/test-results/gp09r-029-graph/run-summary.json)

**E-096 — S-030 / F-036 and reopened S-026 DONE.**

The packaged schema admitted empty or partial producer/prior IDs on projected references and populated trailing fields on raw references. It now selects the byte-pair pattern by scope. All four reference types, trailing line endings, uppercase and odd hex, valid raw/direct/reverse/chained results and the runtime rejection of another contributor’s well-shaped key passed. Changed authored contracts/graph-projection/schemas.v2.json, generated internal/gen/contractgraphprojection/artifacts_gen.go and Graph root v2_contract_projection_test.go, contributor_order_v2_test.go and result_v2_validation_test.go. No S-026 blockers.

- `CARTULARY_TEST_RUN_ID=gp09r-030-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.v2_contract_projection` exited 2 (fail). Source `a48590e21c8e670478de17d4650c35028a9fc44d2550c1c8a0de10231bfc1097`. [Receipt](../../.cartulary/test-results/gp09r-030-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-030-generate make generate` exited 0 (pass). [Receipt](../../.cartulary/test-results/gp09r-030-generate/generate/tool-run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-030-format make format` exited 0 (pass). Source `0fe7f09af4411920bdb5ee97cc6e6af884a7b782ee8fba38d9aa67b1471c4577`. [Receipt](../../.cartulary/test-results/gp09r-030-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-030-graph make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.storage.v2_contract_projection,module.graphprojection.engine.v2_semantic_matrix` exited 0 (pass). Source `56ee047f7a7f044ee22ecaabbff3baf4371b6f30e0a371e14d4dc4cbbe391ce7`. [Receipt](../../.cartulary/test-results/gp09r-030-graph/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-030-generated make source-check CHECK=generated` exited 0 (pass). Source `56ee047f7a7f044ee22ecaabbff3baf4371b6f30e0a371e14d4dc4cbbe391ce7`. [Receipt](../../.cartulary/test-results/gp09r-030-generated/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-030-schema make source-check CHECK=schema` exited 0 (pass). Source `56ee047f7a7f044ee22ecaabbff3baf4371b6f30e0a371e14d4dc4cbbe391ce7`. [Receipt](../../.cartulary/test-results/gp09r-030-schema/run-summary.json)

**E-097 — S-031 / F-037 DONE.**

Source scalar admission now owns IDs (including empty endpoint IDs), directions and source labels. The retained red returned a collection limit ahead of direction. All four valid directions, both duplicate namespaces and permutations, the closed direction-specific reason, mixed invalid shapes and direction, and exact scalar issue limits passed. Removing redundant scalar checks from direct and aggregate derivation preserves valid outputs. The Table 3-D reference witness also exposed acceptance when a filter or mapping predicate excluded a relationship with an absent endpoint. The first reference reproduction removed the mapping entirely and correctly saw a mapping error instead. The corrected reproduction isolated both exclusions. After configuration admission, both endpoint IDs must resolve to submitted entities before eligibility. A well-formed absent ID returns missing_relationship_endpoint. An existing source entity without its selected vertex retains relationship_endpoint_not_projected. This implements the submitted-entity requirement in the reviewed owner. Changed docs/graph_projection_nlspec.md and Graph root input.go, engine_v2.go, validation.go, validation_registry.go, project.go, direct_mapping.go, aggregation.go, engine_v2_test.go and input_counts_v2_test.go. No S-031 blockers.

- `CARTULARY_TEST_RUN_ID=gp09r-031-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 2 (fail). Source `7dd913e7679f0a9b1c8480405f14e4373fca2cdc2340a93cf4bd2e748feebbdb`. [Receipt](../../.cartulary/test-results/gp09r-031-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-031-format make format` exited 0 (pass). Source `32e81d96b9406595dc33e722fff13ffd4ab7d3135ee535c8854f01c8de6b49a9`. [Receipt](../../.cartulary/test-results/gp09r-031-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-031-scalars make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.engine.v2_maximum_semantic_bounds` exited 0 (pass). Source `e794be1ab598f54164eb2160bb0469212f2d3df92560cb0833127b475a1be009`. [Receipt](../../.cartulary/test-results/gp09r-031-scalars/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-031-reference-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 2 (fail). Source `ed443810c7080615b6ae7001fd469930b43ec6e2779cd7c880a1afad35af6024`. [Receipt](../../.cartulary/test-results/gp09r-031-reference-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-031-reference-red2 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 2 (fail). Source `6d9be33b546f5f2687618dbdc856323bb92b221811835a7573c446fd9825889e`. [Receipt](../../.cartulary/test-results/gp09r-031-reference-red2/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-031-format2 make format` exited 0 (pass). Source `40f76534d3a680dc7f127eedcbd3df62be28f1f6c544d3e762bbdab588372604`. [Receipt](../../.cartulary/test-results/gp09r-031-format2/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-031-graph make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.engine.v2_maximum_semantic_bounds` exited 0 (pass). Source `5bead77e301e931ce97efc4ba1c8977686cd2c542a070282466f1756aabb57b8`. [Receipt](../../.cartulary/test-results/gp09r-031-graph/run-summary.json)

**E-098 — S-032 / F-038 and reopened S-027 DONE.**

Before repair the public callback returned cancellation but the decoded 100,001-row collection had already been reordered. After repair the first uniqueness checkpoint sees the oversized input in its submitted order. The cancellable scan counts repeats beyond the first in each namespace with one temporary set at a time and no untrusted size hint. Only admitted typed rows are sorted. The bound and pure rows passed the nine ceilings, final-position and cross-collection duplicates, issue budget, terminal cancellation, permutations and fixed envelope/child bytes. Changed Graph root input.go, engine_v2_test.go and input_counts_v2_test.go. No S-027 blockers.

- `CARTULARY_TEST_RUN_ID=gp09r-032-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_maximum_semantic_bounds` exited 2 (fail). Source `ba74fe2285c4e44ca505f8aa067444695ff53694c624fe4aa8c9d83858f89f6e`. [Receipt](../../.cartulary/test-results/gp09r-032-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-032-format make format` exited 0 (pass). Source `6d8f89c3e3d720f36bdd63a4539bf155c13d6fdd081e69a38bc6e0b40a1bfbf2`. [Receipt](../../.cartulary/test-results/gp09r-032-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-032-graph make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_maximum_semantic_bounds` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-032-graph/run-summary.json)

**E-099 — Finalization and source freeze.**

Finalization reported zero generated changes. Schema, tier coverage and generated refresh passed. Optional warm-run evidence and scheduler maintenance were skipped because no compatible full warm-check root was supplied. The [frozen inventory](../../.cartulary/evidence-manifests/graph-projection-gp09-repair/frozen-source.json) matched its pre-finalization inventory byte for byte. SHA-256 was `9fe238055b266102910c820adcd3b2bfaacf4b9f0c16b10fe13539a238f79b2b` across 7,655 non-document paths. All entry paths remained present. The 16 changed executable or contract entries are the scoped Graph repair paths from E-094–E-098.

- `CARTULARY_TEST_RUN_ID=gp09r-finalize make agent-finalize` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-finalize/run-summary.json)

**E-100 — Frozen-source portfolio.**

All seven final selections passed on harness source fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007. The Graph owner run covered eleven rows, including storage, reconstruction, restore publication and atomic rollback. Network Flow passed four rows and Reporting passed two. The generated, schema, Go lint and backend-boundary selections passed. In command order, the runs recorded 29, 15, 15, 14, 12, 15 and 13 completed cleanup operations, with no failures. Every same-run retained-secret scan passed. No required selection was missing.

- `CARTULARY_TEST_RUN_ID=gp09r-final-graph make test-slice OWNER=module.graphprojection` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-final-graph/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-final-nf make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.time_bucket_saved_graph_lifecycle,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-final-nf/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-final-reporting make test-slice OWNER=module.reporting ROWS=module.reporting.integration.graph_input_admission_reads_the_exact_selected_c_4c42af0393,module.reporting.unit.graph_v2_exact_selection_and_rendering_f54d8b1a62` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-final-reporting/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-final-generated make source-check CHECK=generated` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-final-generated/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-final-schema make source-check CHECK=schema` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-final-schema/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-final-go make lint CHECK=go` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-final-go/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09r-final-boundary make lint CHECK=backend-boundary` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-final-boundary/run-summary.json)

**E-101 — Complete Graph owner audit.**

The explicit manifest selected the single fresh Graph root from E-100. The owner audit accepted every required row (11 of 11) in one partition, with zero rejected or unused inputs. Its source matches E-100, all ten cleanup operations completed, and the retained-secret scan passed. [Input manifest](../../.cartulary/evidence-manifests/graph-projection-gp09-repair/graph-owner-roots.json) · [Audit result](../../.cartulary/test-results/gp09r-final-audit/test-evidence-audit/audit.json)

- `CARTULARY_TEST_RUN_ID=gp09r-final-audit make test-evidence-audit OWNER=module.graphprojection EVIDENCE_ROOTS_FILE=.cartulary/evidence-manifests/graph-projection-gp09-repair/graph-owner-roots.json` exited 0 (pass). Source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`. [Receipt](../../.cartulary/test-results/gp09r-final-audit/run-summary.json)

**E-102 — Compatibility, rollback and handoff.**

| Finding | Reproduction and repair | Frozen acceptance |
| --- | --- | --- |
| F-034 | E-094, invented or wrong producer and unlisted source kind | V-001/V-017/V-019/V-028. Borrowed publication and seeded content reads rejected the invented leaf. |
| F-035 | E-095, endpoint-to-contributor cycle and direct endpoint | V-001/V-017/V-028. Both endpoint positions rejected the cycle, and the valid repeated producer and mixed chain passed. |
| F-036 | E-096, raw and projected contributor keys with the wrong component counts | V-001/V-019/V-003/V-011. The packaged schema checked each reference kind, and Graph checked the decoded key against its contributor. |
| F-037 | E-097, invalid direction competing with duplicates and overflow, plus a dangling submitted endpoint | V-001/V-017/V-018. Permutations preserved the public reason and safe count. The issue ceiling, earlier mapping error and later endpoint verdict passed. |
| F-038 | E-098, an oversized collection sorted before cancellation | V-001/V-018. The cancellation witness checked the submitted order. Accepted inputs reproduced the S-022 byte vectors. |

The four fixed Graph families retained their exact bindings, canonical envelopes and ordered child bytes. Ordinary and locked reads of the historical numeric result and its reprojection matched the original bytes. The Network Flow default and time-bucket producers retained their complete fixed binding tuples. Reporting covered exact selection and admission on the repaired source.

A stored selection with invalid registry membership or cyclic producer dependencies fails content reads until its source owner rematerializes it. Metadata-only locking still returns the binding. Reporting keeps previously admitted inputs. A mismatched restore binding rolls the transaction back atomically. The repairs require no stored-row rewrite, public Go API change or Recovery version bump. On new requests, source scalar errors precede duplicate and overflow errors. An absent submitted endpoint returns `missing_relationship_endpoint` even on a filtered or predicate-excluded relationship. A present source entity without its selected vertex retains the projected-endpoint verdict.

For rollback of both admission repairs, revert S-032 before S-031. Revert S-029 before S-028 when removing their shared proof. Regenerate with `make generate` from the matching authored schema for either S-028 or S-030. Reverting the read-side checks can readmit malformed content. Rerun the owner and peer portfolio on the rollback candidate. The single-namespace uniqueness set scales with distinct submitted IDs under the input-byte guard. Deployment and release qualification remain separate.

The index still matches `9203107cd`. The working tree has 22 modified and three new paths relative to HEAD. Comparing it with E-093 preserved all starting paths, the fixed fixtures, the twelve tracker sections and the 19,469-byte E-082–E-092 checkpoint block. The [ending source inventory](../../.cartulary/evidence-manifests/graph-projection-gp09-repair/end-source.json) matches E-099, including modes. `git diff --check` passed. The [native exits](../../.cartulary/evidence-manifests/graph-projection-gp09-repair/native-exits.json) record the final portfolio and handoff separately from the original candidate.

**E-103 — WF-35 / T-041 final-text gate.**

The draft lint exited 0. The final gate requires exit 0 from `CARTULARY_TEST_RUN_ID=gp09r-handoff-final make lint-markdown` against this completed tracker, with its receipt below. E-099–E-102 record the qualified source, the finding-to-evidence map and the scope review. This closes T-041 last for Graph and its named consumers. The changes remain unstaged and uncommitted. [Draft lint](../../.cartulary/test-results/gp09r-handoff-draft/adhoc/lint-markdown/tool-run-summary.json) · [Final-text lint](../../.cartulary/test-results/gp09r-handoff-final/adhoc/lint-markdown/tool-run-summary.json)

### GP-15 rollback interruption follow-up

**E-156 — Entry and owner review.** All 7,671 non-Markdown entries and modes match the previous final freeze. Entry contains 115 modified and nineteen untracked paths, with the index matching HEAD. Core 00 and Domain ownership were rechecked. Extensions EXT-REQ-215 owns the deadline interval and first-interruption semantics; EXT-REQ-219 now explicitly applies those semantics after proven rollback. The coordinator owns verdict selection, and consumers retain their existing transport and worker policies. The owner clarification precedes implementation. F-087 is the remaining P2 defect: a completion interruption can lose to an ordinary failure or database conflict, including an already classified conflict. T-041 remains open through repaired-source qualification. [Entry inventory](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/entry-source.json) · [Starting tree](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/entry-status.txt)

**E-157 — F-087 witnesses and shared repair.** Before the production edit, `gp15rb-red-coordinator` and `gp15rb-red-nf` exited 2 with product assertion failures: rollback-time cancellation/expiry returned conflicts, and both PostgreSQL SQLSTATE cases returned 409 instead of the required 503. The shared rollback boundary now preserves completion fatality and unproven-absence fatality, then previously selected fatal/cancellation/timeout outcomes, then the first recorded completion interruption, and only then existing or newly classified conflicts. No new public interface or consumer classification was added. The coordinator matrix covers raw/wrapped `40P01` and `40001` at preconditions, writes and finalization; exact-deadline/cancellation ties; earlier cancellation followed by expiry; typed conflicts; ordinary owner and registered validation failures; selected terminal outcomes; uncertain rollback; and grace exhaustion without duplicate fatal notification. Every applicable case asserts one attempt, one rollback, no commit and no partial result. Network Flow advances the injected clock inside real PostgreSQL rollback with a live outer context, checks both retryable timeout 503 verdicts, retains the no-interruption 409 controls and replay matrix, and compares receipt, binding, indicator, observation and audit snapshots. Both focused Make runs passed: the complete coordinator owner (unit and PostgreSQL atomicity) and the Network Flow contributor/indicator-link route row. Finalizer review retained the original validation-triggered grace-exhaustion control alongside the new SQLSTATE write controls; the focused coordinator unit row passed again at `gp15rb-focused-lifetime`, before the final `agent-finalize` run. [Coordinator baseline](../../.cartulary/test-results/gp15rb-red-coordinator/run-summary.json) · [Network Flow baseline](../../.cartulary/test-results/gp15rb-red-nf/run-summary.json) · [Coordinator repair](../../.cartulary/test-results/gp15rb-focused-coordinator/run-summary.json) · [Network Flow repair](../../.cartulary/test-results/gp15rb-focused-nf/run-summary.json)

**E-158 — Final frozen-source qualification.** The final `make agent-finalize` run passed. Its review found no generated changes beyond the entry tree. The freeze contains 7,671 non-Markdown entries and modes, inventory SHA-256 `ba8cd54ad3a6e0f795ac36c4ff2b4852177ef935025e503ae1604f80a2be55cf`. The harness source is `sha256:33177aa63b3308ef59b00cafd16872f3c21f48656eb4bde860b5c855b0fed198`. All 43 selected rows passed against that source. The explicit Graph audit accepts eleven required rows from one compatible root, with zero rejected or unused inputs. All final native exits are 0; graph-backed cleanup completed and retained-secret scans passed. The baseline failures remain retained. [Finalizer](../../.cartulary/test-results/gp15rb-finalize-final/run-summary.json) · [Frozen inventory](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/frozen-source.json) · [Graph audit](../../.cartulary/test-results/gp15rb-final-audit/test-evidence-audit/audit.json) · [Commands, sessions, exits, cleanup and scans](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/native-exits.json)

| Final Make selection | Passed rows | Receipt |
| --- | --- | --- |
| `test-slice OWNER=module.graphprojection` | 11 | [Graph](../../.cartulary/test-results/gp15rb-final-graph/run-summary.json) |
| Network Flow saved-graph/default/time-bucket selections | 5 | [Selectors and lifecycle](../../.cartulary/test-results/gp15rb-final-nf/run-summary.json) |
| Network Flow contributor/indicator-link route and fixture corpus | 16 | [Routes and fixtures](../../.cartulary/test-results/gp15rb-final-corpus/run-summary.json) |
| `test-slice OWNER=module.crossownertransaction` | 2 | [Coordinator and PostgreSQL atomicity](../../.cartulary/test-results/gp15rb-final-coordinator/run-summary.json) |
| Reporting exact-selection admission and rendering | 2 | [Reporting](../../.cartulary/test-results/gp15rb-final-reporting/run-summary.json) |
| Web saved-graph receipt transport and exploration focus | 2 | [Web](../../.cartulary/test-results/gp15rb-final-web/run-summary.json) |
| Imports transaction-current admission, atomic publication and owner-error translation | 3 | [Imports](../../.cartulary/test-results/gp15rb-final-imports/run-summary.json) |
| Incident Bundle final-publication recheck and worker terminal outcomes | 2 | [Incident Bundles](../../.cartulary/test-results/gp15rb-final-bundles/run-summary.json) |

`make source-check CHECK=generated`, `make source-check CHECK=schema`, `make lint CHECK=go`, `make lint CHECK=backend-boundary`, and `make lint CHECK=types` all passed against the same frozen source. `git diff --check` passed. No required product or static gate was skipped. Markdown remains outside executable contracts and frozen-source inputs. [Generated](../../.cartulary/test-results/gp15rb-final-generated/run-summary.json) · [Schema](../../.cartulary/test-results/gp15rb-final-schema/run-summary.json) · [Go](../../.cartulary/test-results/gp15rb-final-go/run-summary.json) · [Boundary](../../.cartulary/test-results/gp15rb-final-boundary/run-summary.json) · [TypeScript](../../.cartulary/test-results/gp15rb-final-types/run-summary.json) · [Whitespace](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/diff-check.json)

**E-159 — Handoff and closure gate.** F-087 maps to Extensions EXT-REQ-215/219, the E-157 failing witnesses, the shared rollback repair and the E-158 passing receipts. Review of the follow-up diff confirms one production change in `crossownertransaction/coordinator.go` and regression extensions in its existing unit row and Network Flow's existing route row. The only documentation changes are the Extensions clarification and this tracker. All other non-Markdown source bytes and modes match entry, including Graph bindings, selector expectations, fixture revisions and replay implementation. Authorized replay payloads remain exact under both entry paths; no application transport or worker policy changed. The final tree retains 115 modified and nineteen untracked paths, and the index still matches `9203107cd76ba2849bf02dc466309b6697be3e93`. Nothing is staged or committed. E-151–E-155 and all preceding qualification history remain intact. [Follow-up source diff](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/followup-source.diff) · [Complete tracked source diff](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/complete-tracked-source.diff) · [Complete untracked source diff](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/complete-untracked-source.diff) · [GP-15 binding reference](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/bindings-reference.json)

The compatibility change is limited to proven-aborted operations interrupted before rollback completion: they return the existing cancellation or timeout outcome instead of an ordinary failure or conflict. Fatal outcomes and proven commits retain precedence. This remains a single pre-production cutover with no API, error type, receipt format, Graph binding, schema, database conversion or legacy path change. Ship this coordinator correction with the existing GP-15 application repairs. Rollback restores the preceding matched source and requires requalification.

The completed-handoff Markdown check passed, and the final source comparison matches every frozen entry and mode. T-041 closes last on the final-text Markdown receipt below. S-026 and S-027 remain closed. No required gate was skipped; the changes are ready to commit within this qualification scope and remain unstaged and uncommitted. [Draft handoff lint](../../.cartulary/test-results/gp15rb-handoff-draft/adhoc/lint-markdown/tool-run-summary.json) · [Final-text lint](../../.cartulary/test-results/gp15rb-handoff-final/adhoc/lint-markdown/tool-run-summary.json) · [Final source](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/end-source.json) · [Final review](../../.cartulary/evidence-manifests/graph-projection-gp15-rollback/final-review.json)

### GP-15 replay and conflict follow-up

**E-151 — Entry and owner review.** All 7,669 non-Markdown entries and modes match the GP-15 frozen inventory. Entry has 110 modified and seventeen untracked paths, with the index matching HEAD. Core 00 and Domain ownership were rechecked. NF-REQ-026 and §29.1 now assign both replay paths to application-owned post-lookup admission. Incidents owns the editor/admin and open-incident decision, and Indicators owns target visibility. Extensions assigns conflict classification to the coordinator after rollback proves absence, preserving selected terminal outcomes and validation-result classification. This owner review precedes implementation. T-041 remains open through final qualification. [Entry inventory](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/entry-source.json) · [Starting tree](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/entry-status.txt)

**E-152 — Baseline witnesses and focused repairs.** The new coordinator phase matrix failed for raw and wrapped `40P01` and `40001` from preconditions, writes and finalization. The completed Network Flow baseline returned receipts after membership revocation, role downgrade and closure on both lookup paths; hash mismatch and corrupt receipt also outranked admission. Both database conflicts returned HTTP 500. The initial two Network Flow test attempts also exposed test-composition defects (decoded-number comparison and a missing Network Flow validation-policy dependency); those attempts remain historical and are superseded by the completed baseline. [Coordinator baseline](../../.cartulary/test-results/gp15f-red-coordinator/run-summary.json) · [Completed replay baseline](../../.cartulary/test-results/gp15f-red-nf-complete/run-summary.json)

| Finding | Owner and repair | Focused acceptance and compatibility |
| --- | --- | --- |
| F-085 · Replay admission | NF-REQ-026, Table 5-B and §29.1. One private application replay method loads the receipt, rechecks current editor/admin and open-incident admission, compares the request, decodes the receipt, then checks target visibility. Storage no longer orchestrates policy. | Both lookup paths reject membership revocation, role downgrade and closure before receipt-dependent verdicts, including mismatched hashes and malformed receipts. Authorized replay and source-table deletion retain status and payload; hidden targets remain denied. Snapshots prove unchanged receipt bytes, bindings, indicators, observations and audit rows. Test-owned dependency wrappers control ordering, with real Core admission and PostgreSQL rollback. No timing sleeps or production hooks. |
| F-086 · Conflict outcomes | Extensions EXT-REQ-219 transaction protocol. The common rollback boundary classifies database failures only after `CommitAbsent`. Selected fatal, cancellation, timeout and conflict outcomes retain precedence; participant-validation failures keep their registered classifier. | Raw and wrapped SQLSTATEs across all three phases retain their causes and become typed conflicts. Unknown rollback remains fatal, ordinary owner failures retain identity, and committed outcomes remain authoritative. Network Flow maps coordinator-produced conflicts to retryable HTTP 409 with exact reasons, one attempt and zero durable effects. No receipt or wire conversion. |

`gp15f-focused-coordinator` passes both coordinator rows, and `gp15f-focused-nf` passes the route row including the existing locked-admission races and the new replay/conflict matrix. These focused receipts precede final source freezing and do not close T-041. [Coordinator](../../.cartulary/test-results/gp15f-focused-coordinator/run-summary.json) · [Network Flow](../../.cartulary/test-results/gp15f-focused-nf/run-summary.json)

**E-153 — Qualification correction.** The first frozen candidate passed all 43 selected test rows, generated/schema checks and Go/TypeScript lint. Backend-boundary lint rejected the test's direct Extensions owner import. Application test composition now lives in `internal/testutil/appsupport`, using the production generated registry and validation policy. The failure and first-source receipts are retained; final qualification uses a new source freeze and fresh identities. No production behavior, Graph binding or fixture expectation changed during this correction. [Boundary diagnosis](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/boundary-diagnosis.txt) · [First freeze](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/first-frozen-source.json)

**E-154 — Final frozen-source qualification.** `make agent-finalize` passed after the test-composition correction. The final inventory contains 7,671 non-Markdown entries and modes, SHA-256 `f5ab8bfc3c379689c8b8395df16e883eb147566523d714d80ebade8280932da2`. The harness source is `sha256:e717334edaa96620606be43d2521f4b3bb2309f008b61e944d0d100b6505d2e9`. All 43 selected product rows pass against this source. The explicit complete-owner Graph audit accepts all eleven rows with zero rejected or unused inputs. Native exits are 0; graph-backed cleanup completes and retained-secret scans pass. [Finalizer](../../.cartulary/test-results/gp15f-refinalize/run-summary.json) · [Frozen inventory](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/frozen-source.json) · [Explicit Graph audit](../../.cartulary/test-results/gp15f-final-audit/test-evidence-audit/audit.json) · [Native commands, exits and cleanup](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/native-exits.json)

| Final Make selection | Passed rows | Receipt |
| --- | --- | --- |
| `test-slice OWNER=module.graphprojection` | 11 | [Graph](../../.cartulary/test-results/gp15f-final-graph/run-summary.json) |
| GP-15 Network Flow saved-graph/default/time-bucket selections | 5 | [Selectors and lifecycle](../../.cartulary/test-results/gp15f-final-nf/run-summary.json) |
| GP-15 Network Flow route and fixture corpus | 16 | [Routes and fixtures](../../.cartulary/test-results/gp15f-final-corpus/run-summary.json) |
| `test-slice OWNER=module.crossownertransaction` | 2 | [Coordinator](../../.cartulary/test-results/gp15f-final-coordinator/run-summary.json) |
| GP-15 Reporting exact-selection/admission/rendering selections | 2 | [Reporting](../../.cartulary/test-results/gp15f-final-reporting/run-summary.json) |
| GP-15 web receipt-transport and exploration selections | 2 | [Web](../../.cartulary/test-results/gp15f-final-web/run-summary.json) |
| Imports transaction-current admission, Network Flow atomic publication and safe owner-error translation | 3 | [Imports](../../.cartulary/test-results/gp15f-final-imports/run-summary.json) |
| Incident Bundle final-publication recheck and worker terminal outcome | 2 | [Incident Bundles](../../.cartulary/test-results/gp15f-final-bundles/run-summary.json) |

`make source-check CHECK=generated`, `make source-check CHECK=schema`, `make lint CHECK=go`, `make lint CHECK=backend-boundary` and `make lint CHECK=types` pass under the matching `gp15f-final-*` identities. The final boundary check validates the corrected test composition. The detailed row selections and exact commands are retained in the native receipt ledger.

**E-155 — Handoff and closure gate.** The follow-up changes six existing executable-source paths and adds two test-support paths. Each reconstructed entry file in the complete follow-up source diff matches its retained entry hash. All other entry source bytes and modes are unchanged. Graph implementation, selector expectations, fixture revisions and complete GP-15 bindings remain exact; the qualified Graph and selector rows confirm the same pins. Authorized replay retains its recorded status and payload. Rejected replay preserves stored receipts and durable owner rows. No data conversion, legacy reader or database migration is introduced. [Source delta](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/source-changes.json) · [Complete follow-up source diff](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/followup-source.patch) · [Unchanged GP-15 bindings](../../.cartulary/evidence-manifests/graph-projection-gp15/bindings.json)

The combined working tree has 115 modified and nineteen untracked paths. The index still matches HEAD. The complete tracked and untracked diffs preserve the starting GP-15 work. End inventory matches the final freeze, and Git whitespace review passes. Shipping and rollback keep the application and coordinator repairs together with the GP-15 writer, readers and selector bindings; rollback requires requalification. [End inventory](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/end-source.json) · [Final working tree](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/end-status.txt) · [Combined tracked diff](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/combined-tracked.patch) · [Combined untracked diff](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/combined-untracked.patch)

The pre-closure handoff Markdown check passed, followed by an exact frozen-source comparison. T-041 closes last on the final-text Markdown receipt below. S-026 and S-027 remain closed. No required gate was skipped. The repairs are ready to commit within this qualification scope; the changes remain unstaged and uncommitted. [Pre-closure Markdown](../../.cartulary/test-results/gp15f-handoff-preclose/adhoc/lint-markdown/tool-run-summary.json) · [Final-text Markdown](../../.cartulary/test-results/gp15f-handoff-final/adhoc/lint-markdown/tool-run-summary.json) · [Final review](../../.cartulary/evidence-manifests/graph-projection-gp15-followup/final-review.json)

### GP-15 clean cutover

**E-142 — Entry and owner review.** The 7,668 non-Markdown source entries and modes match GP-14 at `f6539474c6d82c8c8b01be5467c67dff62bb9aa42440f937a034f64a309cfb40`. The index matches HEAD. Entry includes 36 modified and sixteen untracked paths. The user selected separate metadata emission, retirement of graph-level properties and chronological extrema with first-contributor ties. Graph §§3–8 own these semantics. Core 00, Domain, Network Flow’s two adapter configurations, Reporting’s exact-selection port, Recovery v5 reconstruction and the pure Graph package boundary were reviewed. No other source or database selection needs migration. The envelope change requires new complete bindings for both selectors.

**E-143 — Baseline witnesses and cutover proofs.** The original GP-14 writer and validator accepted both independently rebound reverse witnesses, surrogate replacement in keys and values, and metadata as an ordinary property source. The two malformed reverse envelopes are retained in both shapes. The current-shape cases exercise the semantic proof after the new envelope has passed. [Original red run](../../.cartulary/test-results/gp15-red/run-summary.json) · [Witness envelopes](../../.cartulary/evidence-manifests/graph-projection-gp15/witnesses)

The merge reproduction replays the GP-14 comparison and contributor loops with counters in the current calling convention. The retained GP-14 source files match the entry inventory. Four and 1,000 repeated default selections after a first operand caused eight and 2,000 full encodings. A 4,096-contributor group made exactly as many checkpoints as its control without a merge. Both namespaces returned success when asked to cancel inside the group. The original passing microsecond-maximum test is retained. [Merge reproduction](../../.cartulary/test-results/gp15-merge-replay/run-summary.json) · [Baseline comparison](../../.cartulary/evidence-manifests/graph-projection-gp15/gp14-semantics_v2.go.txt) · [Baseline contributor loops](../../.cartulary/evidence-manifests/graph-projection-gp15/gp14-aggregation.go.txt)

| Finding | Adopted Graph clause and repair | Admission and compatibility witness |
| --- | --- | --- |
| F-079 · Paired fields | §§5, 6.3, 8. Same-kind linked edges agree independently in both namespaces, including number spelling, missing members, nulls and array order. | The rebound contradiction is `gpres_cd8fdf8804cb1a183d81dbefbf9c6c9c1d8fb695da8c9df203a079a20b9540ea`. Distinct-kind pairs keep their own declarations. |
| F-080 · Repeated defaults | §§3-H, 5. A merge retains its first operand and its comparison with the definition’s default. Both contributor loops honor the item checkpoint. | After the first operand, repeated defaults require zero encodings when the first was defaulted, or two total when it was explicit. Equal explicit operands, a late conflict, cancellation and zero partial results are covered. |
| F-081 · Preservation flags | §§6.1, 6.3, 8. All outputs of a relationship mapping share one preservation choice. Every preserving kind needs an eligible mapping, including with empty output. | The rebound single-bit contradiction is `gpres_822e9620e75545111344ff03faecdec6f8707d423d25fd3706e1f2cd7d0f02c8`. A different contributing mapping can supply a shared kind. |
| F-082 · Unicode admission | §§4, 7. Graph checks original string tokens before consuming their decoded identities. | Isolated high and low surrogates return `invalid_json_syntax`. Valid pairs, U+FFFD, escaped backslashes, and both orders against a duplicate or depth 129 are exercised. |
| F-083 · Metadata emission | Tables 3-H and 3-I, §§4–8. Scoped metadata remains available for selection and grouping. Metadata emission requires an explicit mapping. The result envelope has six output members. | Retired graph property declarations and the old envelope fail. Inert, wildcard, defaulted and projected paths are covered. The 512 MiB complete-envelope ceiling includes every current member. Both selectors switch bindings. |
| F-084 · Timestamp extrema | Table 3-H and §§5–6. Compare instants chronologically. Keep the first canonical contributor's submitted spelling on a tie. | Microsecond extrema and three equivalent spellings are covered. A different submitted order reconstructs the original binding, envelope and children. |

**E-144 — Selector and Recovery cutover.** For the five controls and both selectors the only binding changes are the canonical output digest and result ID. Independent framing reproduces the old and new control pins, including numeric token spelling. Separate Network Flow runs captured both versions, and every child was compared in order, including its ID and sort key. The semantic golden and disjoint-declaration restore fixture carried additional old pins. Restoring the retired member reproduced each original digest and binding before expectations were updated. The stored-array bounds fixture now uses an ordinary property path to reach its intended assertion. [Original selectors](../../.cartulary/test-results/gp15-selector-baseline/run-summary.json) · [Repaired selectors](../../.cartulary/test-results/gp15-selector-bytes/run-summary.json) · [Full selector bindings and ordered child digests](../../.cartulary/evidence-manifests/graph-projection-gp15/selectors.json) · [Control bindings](../../.cartulary/evidence-manifests/graph-projection-gp15/bindings.json)

| Selector | GP-14 result ID | GP-15 result ID |
| --- | --- | --- |
| Default | `gpres_775aade676fd013e8c86b573093f9f4362fa95f375a069ca3fd90af74b7c1a8b` | `gpres_159b0f5bb300271563672babbead7428ef2ec5f6aea63fa1e7f0fabe774a44e9` |
| Time-bucket | `gpres_d52be3bc28fa17b5aa610b184c394e2797c11172923c41adcb7897ca26c49eea` | `gpres_d9aa399320f0ebd95c8affdfd2d271fff0868d44506a284f1fddef252638b0e3` |

Ship the writer, readers, and both source-owner selectors together on fresh pre-production state. Recovery v5 rejects a prior expected binding and rolls the borrowed transaction back. The current binding reconstructs, publishes a second time idempotently, and reads back exactly. Rollback restores the preceding writer, readers and selector expectations as one unit, then requalifies that source.

**E-145 — S-026 and S-027 closed.** Six focused rows passed on the repaired candidate. S-026 proves both current-shape witnesses fail pure admission, before any publication SQL, and across independently seeded exact, locked exact, ordered vertex, ordered edge and traversal reads. All three publication tables stay empty for the rejected IDs. Binding-only locking succeeds. S-027 covers public Unicode and metadata admission, timestamp extrema, merge work, cancellation and result size. The complete-envelope counter passes at 536,870,912 bytes and rejects 536,870,913. The producer and PostgreSQL reader also use exact and one-byte-over limits on small envelopes. The semantic, compiled-schema and transactional reconstruction rows pass. [Focused Graph receipts](../../.cartulary/test-results/gp15-green/run-summary.json) · [Network Flow selector receipts](../../.cartulary/test-results/gp15-selectors/run-summary.json)

**E-146 — Public consumer closure.** A downstream review found the retired envelope member in the authored Network Flow response schema, generated browser validator, route assertions and three browser fixture builders. The owner’s Table 14-H now follows the Graph output members explicitly. The authored schema restricts property declarations to vertex and edge scope. Make regeneration carries that shape into compiled contracts and TypeScript. A browser rejection test covers the retired member, and the compiled-contract test compares Network Flow’s required envelope members against the Graph projection. Both public-route and frozen-fixture witnesses failed before repair. [Public-route red receipt](../../.cartulary/test-results/gp15-route-red/run-summary.json) · [Focused browser receipt](../../.cartulary/test-results/gp15-browser-decoder/run-summary.json)

Fourteen frozen Network Flow fixture revisions contain 24 Graph envelopes in each of the expected and transcript sets. Their existing full bindings were independently reproduced before removing the envelope member and reframing the new IDs. Source inputs, normalized input digests and child bytes are unchanged. The expected and transcript files, manifests, revision selectors and browser fixture references moved together. The privacy comparison separately identified a zero-valued `DeletedLeases` telemetry expectation retired before GP-15. Its current source has no such field, and removing that expectation resolved the sole privacy mismatch. The cutover ledger records this additional change. [Fixture binding ledger](../../.cartulary/evidence-manifests/graph-projection-gp15/fixture-cutover.json) · [Privacy mismatch witness](../../.cartulary/test-results/gp15-qualified-corpus/NF-FIX-016-redaction-2819409940/comparison-summary.json)

**E-147 — Authorized indicator-link admission repair.** Once the public envelope passed, the same route row reached three transaction races that returned HTTP 500. The user explicitly extended GP-15 to repair indicator-link admission. Network Flow §29.1 and Extensions’ ordered transaction protocol were reviewed before the repair. The initiating owner now uses the existing locked precondition phase for incident lifecycle, authorization, source and existing-target checks. The application assembly exposes the admitted participant’s scoped read port. Participant validation requires a successful admission check. A current incident denial precedes a concurrent receipt. Owner failures retain their route verdicts, and coordinator extension-result validation remains centralized. [Three-race failure receipt](../../.cartulary/test-results/gp15-consumer-focus/run-summary.json)

The locked closure, role and source races now return 409, 403 and 409. The route also covers a concurrent receipt under revoked authorization, target deletion, successful reuse and audit failure. Rejected attempts leave the binding count and Core indicators unchanged, preserve the preexisting receipt count, and create no audit. Both coordinator rows pass. [Completed admission receipt](../../.cartulary/test-results/gp15-admission-complete/run-summary.json) · [Coordinator receipt](../../.cartulary/test-results/gp15-coordinator-focus/run-summary.json)

**E-148 — Final frozen source.** Finalization passed and changed zero files. Its optional warm-check maintenance skipped because no compatible complete warm-check root was supplied. The independently required Graph audit has its own explicit source-compatible root manifest. The final inventory contains 7,669 non-Markdown entries at `28555ecbe6010e8fc7d0c224062b8e70d6a20d79be07bba28527815679704384`. Compared with entry, 101 executable paths changed, including one new path. Existing file modes are preserved. The harness identifies this source as `sha256:9fc452fa6d27f9213b6261c10df8210a7a058e2e633f61aca895e6d6ab68c791`. Earlier GP-15 qualification runs retain their source and results. The final portfolio uses this repaired source. [Finalization](../../.cartulary/test-results/gp15-cutover-finalize/unit-artifacts/finalize-summary.json) · [Frozen inventory](../../.cartulary/evidence-manifests/graph-projection-gp15/frozen-source.json) · [Source change ledger](../../.cartulary/evidence-manifests/graph-projection-gp15/source-changes.json)

**E-149 — Qualified portfolio.** The explicit Graph audit accepted all eleven required rows from the final Graph run with zero rejected or unused inputs. The five Network Flow lifecycle and selector rows, sixteen downstream fixture and route rows, two Reporting rows, two browser rows and two coordinator rows pass on the E-148 source. The final Network Flow route row includes the admission races from E-147. Generated and schema source checks, Go lint, backend-boundary lint and TypeScript lint all pass. Every final graph-backed invocation has native exit zero, completed cleanup and a passing retained-secret scan. The native ledger preserves the preceding failures and focused fixes as well. [Graph audit](../../.cartulary/test-results/gp15-final-audit/test-evidence-audit/audit.json) · [Network Flow lifecycle](../../.cartulary/test-results/gp15-cutover-nf/run-summary.json) · [Downstream fixtures and routes](../../.cartulary/test-results/gp15-cutover-corpus/run-summary.json) · [Reporting](../../.cartulary/test-results/gp15-cutover-reporting/run-summary.json) · [Browser](../../.cartulary/test-results/gp15-cutover-web/run-summary.json) · [Coordinator](../../.cartulary/test-results/gp15-cutover-coordinator/run-summary.json) · [Native receipts and static gates](../../.cartulary/evidence-manifests/graph-projection-gp15/native-exits.json)

**E-150 — Final handoff and T-041 closure.** The complete diff preserves the entry changes. The final working tree contains 110 modified and seventeen untracked paths. The index still matches HEAD. All 7,669 non-Markdown entries and existing modes match E-148, and Git whitespace review passes. The six GP-15 findings map to owner clauses and witnesses in E-143, S-026 and S-027 exits in E-145, and full frozen-source qualification in E-149. E-144 records complete old and new bindings, child-byte comparison and atomic reconstruction. E-146 carries the public-schema and frozen-fixture cutover. E-147 records the explicitly authorized indicator-link repair. Shipping and rollback keep the writer, readers, generated browser contract and both selector bindings together. [End inventory](../../.cartulary/evidence-manifests/graph-projection-gp15/end-source.json) · [Final working tree](../../.cartulary/evidence-manifests/graph-projection-gp15/end-status.txt)

The final-text gate is `CARTULARY_TEST_RUN_ID=gp15-handoff-final make lint-markdown`. Its successful native receipt closes T-041 last. All required GP-15 product and static gates ran. Full-repository, deployment and release qualification remain outside this module/direct-consumer task. The changes remain unstaged and uncommitted. [Final Markdown receipt](../../.cartulary/evidence-manifests/graph-projection-gp15/gp15-handoff-final.log)

### GP-14 result admission and resource bounds

E-137 records the GP-13 starting tree. All 7,664 source entries and modes match its frozen inventory. The index matches HEAD at `9203107cd`. The working tree has 31 modified and twelve untracked paths. The [entry inventory](../../.cartulary/evidence-manifests/graph-projection-gp14/entry-source.json) and [Git status](../../.cartulary/evidence-manifests/graph-projection-gp14/entry-status.txt) preserve the baseline. The two repository selectors constitute the full pre-production population.

| Finding | Owner and required evidence |
| --- | --- |
| F-073 | Graph §§6–8. Reject an undeclared raw source kind at each content boundary. |
| F-074 | Graph §§6–8. Reconcile observed scope defaults including dormant aggregation-only kinds. |
| F-075 | Graph §§5–8. Bound the complete canonical envelope at 512 MiB. |
| F-076 | Graph §§4, 7. Bound diagnostic path construction and JSON container depth at 128. |
| F-077 | Graph §§3–5. Separate opaque metadata names, property keys and canonical dotted-key selection. |
| F-078 | Network Flow §14.4. Use the submitted v2 adapter shape and existing resource-error mapping. |

**E-138 — Red witnesses.** The [two-row baseline run](../../.cartulary/test-results/gp14-red/run-summary.json) exited 2. The pure proof admitted undeclared raw kinds. Independently rebound and seeded content for F-073 and F-074 passed the old storage boundary. Their exact bindings are `gpres_8c8f20ad2d0a6733d9043d0f78e898b44f5e2f431fd3b9725cb0b0eab394be54` and `gpres_334743aa2a49831c701bf5829055ad4003d47dc7e694768b7fc7d96dc8a9114e`. Review characterized the original default fan-out and quadratic diagnostic paths by construction. The resource regressions exercise reduced budgets and deep inputs without exhausting host memory.

**E-139 — Focused repair and clean cutover.** The [six Graph rows](../../.cartulary/test-results/gp14-focus/run-summary.json) passed, including the full 100,000-vertex, 250,000-edge result and atomic reverse reconstruction. F-073 and F-074 now fail before SQL and at exact, locked exact, ordered and traversal reads. F-075 charges canonical envelope bytes across emitted fields, references and shells. The [storage boundary](../../.cartulary/test-results/gp14-store-boundary/run-summary.json) passed at capacity and rejected excess bytes. F-076 stops at container 129 and renders bounded diagnostic paths only on failure. F-077 distinguishes opaque metadata, property keys and a bracket-quoted dotted terminal, including in aggregation and endpoint grouping. The [packaged contract and storage rows](../../.cartulary/test-results/gp14-contract-store/run-summary.json) passed. F-078 passed both [Network Flow adapter rows](../../.cartulary/test-results/gp14-focus-nf/run-summary.json).

The semantic matrix initially exposed an obsolete assumption that caller-owned metadata names had the property-key limit. The revised test exercises the limit on an actual property and admits the opaque metadata name. A later storage test seeded a zero publication timestamp and rejected before exercising exact capacity. Correcting the fixture timestamp exposed the intended SQL boundary. The first full Go lint passed vet and staticcheck and found a formatting difference in the contract test. The receipt manifest retains these earlier runs and their native exits.

The [full bindings](../../.cartulary/evidence-manifests/graph-projection-gp14/bindings.json) retain the empty, direct, reverse, aggregate and numeric controls. The complete Network Flow population consists of default `gpres_775aade676fd013e8c86b573093f9f4362fa95f375a069ca3fd90af74b7c1a8b` and time-bucket `gpres_d52be3bc28fa17b5aa610b184c394e2797c11172923c41adcb7897ca26c49eea`. Both retain the complete GP-13 identity tuple. Fresh inputs can select literal dotted keys. Oversized output, excessive nesting and inconsistent content fail closed. Deploy the current writer and selectors together. Rollback restores the preceding qualified writer and contracts together. The pre-production cutover has no database migration.

**E-140 — Frozen source.** After fixing the Go lint finding with `make format`, [finalization](../../.cartulary/test-results/gp14-finalize2/unit-artifacts/finalize-summary.json) changed zero files. Its optional retained-run maintenance skipped because no compatible warm-check root was supplied. The [7,668-entry inventory](../../.cartulary/evidence-manifests/graph-projection-gp14/frozen-source.json) has SHA-256 `f6539474c6d82c8c8b01be5467c67dff62bb9aa42440f937a034f64a309cfb40`. [Changes against entry](../../.cartulary/evidence-manifests/graph-projection-gp14/source-changes.json) comprise 23 modified and four new source files. All entry paths and modes are present. Formatting changed only the contract test between the first and final freezes.

The [complete Graph run](../../.cartulary/test-results/gp14-qualified-graph/run-summary.json) passed all eleven rows. The [explicit Graph audit](../../.cartulary/test-results/gp14-final-audit/test-evidence-audit/audit.json) accepted that single root with zero rejected or unused inputs at source `sha256:ec40d4859ac032e6bbb4c03c09d529cfc8b2952e89f7bc6ad82c2fe72dbe6321`. Both [Reporting rows](../../.cartulary/test-results/gp14-qualified-reporting/run-summary.json) passed exact selection, admission and rendering. The generated, schema, Go lint and backend-boundary checks each exited zero.

The [five Network Flow rows](../../.cartulary/test-results/gp14-qualified-nf/run-summary.json) passed the default and time-bucket saved-graph lifecycles, their backend bindings and adapter admission. Each graph-backed command on the qualified source completed cleanup, passed the retained-secret scan and exited zero. The [native receipts](../../.cartulary/evidence-manifests/graph-projection-gp14/native-exits.json) identify each command, source, run root and outcome.

**E-141 — Final handoff.** The [end inventory](../../.cartulary/evidence-manifests/graph-projection-gp14/end-source.json) matches E-140 byte for byte. The index matches HEAD. The [combined working tree](../../.cartulary/evidence-manifests/graph-projection-gp14/end-status.txt) includes the starting changes and totals 36 modified and sixteen untracked paths. Git whitespace review passed. Every required gate has a receipt. T-041 closes on exit zero from [final Markdown lint](../../.cartulary/test-results/gp14-handoff-final/adhoc/lint-markdown/tool-run-summary.json) against this text.

### GP-13 current v2 admission

**E-132 — Entry and owner review.** HEAD and index match `9203107cd`. All 7,663 non-Markdown paths and modes match the GP-12 frozen source. The [entry status](../../.cartulary/evidence-manifests/graph-projection-gp13/entry-status.txt) records 25 modified and eleven untracked paths. E-127–E-131 remain historical. Graph §§4–8 own the input verdicts, derived registry limits, identity and content proof. Core 00, Domain, Network Flow, Reporting and Recovery retain their existing authority and ports. The two repository Network Flow configurations are the entire pre-production source population.

**E-133 — Reproduction.** The [initial three-row run](../../.cartulary/test-results/gp13-red/run-summary.json) exited 2. The pure result proof admitted all four independently rebound envelopes. ProjectV2 admitted the unused oversized source array and selected the later configuration member on a type tie. The first expansion fixture duplicated metadata IDs. After correcting only that fixture, the [bounds run](../../.cartulary/test-results/gp13-red-expansion/run-summary.json) exited 2 and the baseline issued `gpres_40e5e246280bbe00d8b651e7bba3078f4a1be532ee1c39aad5b2fa1ae7e963eb` for 100,001 expanded declarations.

| Finding | Owner and witness | Current cutover |
| --- | --- | --- |
| F-066 · Edge capability union | §§5, 6.1, 6.3, 8. Old reverse `gpres_57e5a323f6e8c51160008243fd83be74cdaf1d592021d459cb384981d013510c`. | All registered producer contributions must admit the declared union. The current reverse fixture produces `gpres_b356854df17070a45dbf005adfc715170cab59b2e5badb022dea0dfe1854a586`. |
| F-067 · Raw source provenance | Tables 3-C/D and §§6.3, 8. `gpres_d0aaeefb44f5671b0eddec26159043039debf70bac1eeea936d45098294eaa0c`. | Repeated source ID claims within a namespace must agree on kind across all direct and aggregated rows. |
| F-068 · Object sort keys | §§5, 6.3, 8. `gpres_1594a0d95ce665ffaca73f1398e34be4640071ebe9fa6b37cb692d300f5d880e`. | Reject the reordered aggregate edge despite exact child bytes and a self-consistent binding. |
| F-069 · Aggregate label capability | §§6.1, 6.3, 8. `gpres_dc0e89db1c2d24ac6a8f6ab5fbfdd04d3f235f54a6c66ad46127ad60ff922524`. | A kind with only aggregation rules cannot preserve source labels, including in empty output. |
| F-070 · Expansion budget | §§4, 6.1, 7, 8. Enabled definitions over 1,000 configured kinds. | At most 100,000 combined top-level declarations. Nested property references fit their declarations. Preflight rejects the 100-million-entry product before expansion. |
| F-071 · Input array ceiling | §§3, 4, 7. An unused array of 100,001 booleans returned a binding. | Typed input positions and consumed mapped arrays share the resource ceiling. Above capacity the caller receives the bounded resource verdict. |
| F-072 · Competing shape errors | Tables 3-A–J and §4. Numeric version and string empty-registry flag. | Declared member order and immediate descent select the version error in either JSON order. |

**E-134 — S-026 and S-027 complete.** The [six focused rows](../../.cartulary/test-results/gp13-focus2/run-summary.json) passed after the current-sentinel rollback checks. F-066–F-069 reject in pure validation, before publication SQL, and through independently seeded exact, locked, ordered and traversal reads. All three tables remain empty after rejected publication. The binding-only lock succeeds. The old expected reverse binding fails reconstruction atomically, and the corrected selection rebuilds and republishes idempotently. F-070–F-072 pass through public admission. The direct preflight test submits 10,000 kinds and 10,000 definitions, and the public exact-capacity case materializes 100,000 declarations. The [full bindings](../../.cartulary/evidence-manifests/graph-projection-gp13/bindings.json) include all four negative witnesses and the current Graph, numeric and Network Flow controls.

**E-135 — Frozen qualification.** [Finalization](../../.cartulary/test-results/gp13-finalize/unit-artifacts/finalize-summary.json) passed and reported zero generated changes. Its optional retained-run maintenance skipped because no compatible full warm-check root was supplied. The [7,664-entry inventory](../../.cartulary/evidence-manifests/graph-projection-gp13/frozen-source.json) hashes to `84d7d862e6ccfcc32d61d890f635bdf433e507d993324b89d2614d40ffbb00fd`. [Source changes against entry](../../.cartulary/evidence-manifests/graph-projection-gp13/source-changes.json) comprise 24 modified files and the new ordered input-shape validator. Every starting path and file mode remains present. The [Graph run](../../.cartulary/test-results/gp13-final-graph/run-summary.json) passed eleven rows. The [explicit owner audit](../../.cartulary/test-results/gp13-final-audit/test-evidence-audit/audit.json) accepts one compatible root with zero rejected or unused inputs at source `sha256:7ad8a858ca40af16a8867f1b7b6e44437009319ec878a13ded21bc547e54cb04`.

The [four Network Flow rows](../../.cartulary/test-results/gp13-final-nf/run-summary.json) reproduced both selector bindings. Default expands to 24 declarations and time-bucket to 26. Both [Reporting rows](../../.cartulary/test-results/gp13-final-reporting/run-summary.json) passed. Generated and schema checks, Go lint, backend-boundary lint, cleanup and retained-secret scans passed on the same source. Existing valid fixture bytes stayed exact. The pre-production cutover requires only the current writer and current source-owner inputs. Rollback selects a writer and bindings that reproduce one another, followed by requalification.

**E-136 — Final handoff.** The end inventory matches the frozen inventory byte for byte. Git whitespace review passed, and the index matches HEAD. The candidate contains 31 modified and twelve untracked paths including the starting changes. [Native receipts](../../.cartulary/evidence-manifests/graph-projection-gp13/native-exits.json) record the commands, run roots, exits, cleanup and retained-secret scans. T-041 closes on native exit zero from [final Markdown lint](../../.cartulary/test-results/gp13-handoff-final/adhoc/lint-markdown/tool-run-summary.json) against this text. All required gates have a receipt.

### GP-12 producer and content admission

**E-127 — Entry.** HEAD and index match `9203107cd`. All 7,660 source
entries and modes match GP-11. The 25 modified and eight untracked paths are
recorded in the [entry inventory](../../.cartulary/evidence-manifests/graph-projection-gp12/entry-source.json)
and [Git status](../../.cartulary/evidence-manifests/graph-projection-gp12/entry-status.txt).
The Graph owner task guide passed. GP-11 evidence remains historical.

**E-128 — Owner and baseline witnesses.** Graph §§3–8 own the eight rules
below. Domain ownership, Network Flow §19, Reporting §15 and Recovery retain
their existing boundaries. The [baseline run](../../.cartulary/test-results/gp12-red/run-summary.json)
exited 2 on the intended assertions before the writer or proof changed.

| Finding | Owner clause and failing witness | Repaired behavior |
| --- | --- | --- |
| F-058 · Output label ceiling | §§5–7. Each input list fit, yet the writer returned 257 labels. | Distinct defaults, mapping and source labels count at emission. The 257th yields a terminal resource error before retaining the object. |
| F-059 · Mapped edge identity | §§5, 6.3, 8. An all-zero reverse digest was publishable. | The writer and proof share edge transcripts. One direction policy must explain all rows of a mapping and fit its forward kind registry. |
| F-060 · Paired labels | §§5, 6.3, 8. Removing only the reverse edge’s preserved label passed. | Resolve the complete pair and compare its ordered labels. Properties and metadata still use each output’s own kind. |
| F-061 · Reachable nulls | §§3-H/I, 6.1, 8. A graph integer with two error policies admitted null. | Whole null requires nullable output and a possible null-producing policy. The proof uses only the stored declaration. |
| F-062 · Registry producers | §§6.1, 8. An empty result could advertise a ghost kind. | The kind requires a source kind or aggregation rule. A configured producer can emit zero objects. |
| F-063 · Input resource precedence | §§4, 7. An absent endpoint rule masked 257 default labels. | Finish resource admission before ordering aggregation rules. Isolated input violations retain their public verdict. |
| F-064 · Undeclared mapping source | Tables 3-F/G. An unused mapping could advertise an undeclared source. | Scope-specific declared sets apply regardless of source rows and filters. |
| F-065 · Required and optional keys | Tables 3-F/G. A mapping with intersecting key lists received a result binding. | Detect one overlap per mapping and reject through the existing mapping verdict. |

Four independently rebound IDs pin the pure proof and PostgreSQL gates.

| Malformation | Result ID |
| --- | --- |
| Fabricated reverse ID | `gpres_1fa8889a387f3a99b7e15ce500e742d0c2e806a436b69d9dc09fe42b2d51ef01` |
| Divergent reverse labels | `gpres_2d5f4060bfb3e09045e8d50d6c7d881be57c1ebe51d62123e9d02e77f4b0fa9d` |
| Unreachable graph null | `gpres_32b1a2f62a434b0431803218c1ca0ee0f596054ac9f7a298023c0af5f5643610` |
| Producerless kind | `gpres_79311a8ffb73e858880734ee586aaa6f7a766098b9041a57d3ea55e4e1ce38f1` |

**E-129 — Source-owner compatibility.** The registered Network Flow default
and time-bucket selectors use declared mapping kinds, empty optional lists,
mapping-only labels, empty defaults and forward-only preserve mappings.
Neither needs rematerialization for GP-12. GP-12 recorded a conditional inventory for additional owners. E-132 supersedes that condition for the confirmed pre-production scope. Reporting keeps its previously admitted immutable contributions.

The baseline writer reproduced the old binding for a relationship mapping
that placed `group` in both key lists. The owner correction retains it only in
the required list. Both fixture bindings share view `mixed-graph`, owner
`network_flow_activity`, snapshot `snapshot-ordering`, schema
`graph_projection.v2`, version `test-v1`, source digest
`42fcd3f68e654a7fa33a6c044f721e21ce8bddec9073fbce5786ed4f5b11348c`
and output digest `274db2d52afbe43c4df1dd1d19d08b1430b3e53ba297da83b6ba8eae447ec055`.

| Selection | Normalized configuration SHA-256 | Result ID |
| --- | --- | --- |
| Overlapping lists | `8938bebc1b56d4ac4203d837875fe07bb10aea978bdf0c9bf008ba6a830a22c7` | `gpres_3d981c993fec303fb1e1cb1e34fb680cfe77251d96334001c988308a1aae455b` |
| Disjoint lists | `1f8cd704d029b626fd59926aa2ea6d7fecf164e50985b65d490b205faebaf60c` | `gpres_ad27d03d1882f46d9da60b6c711c9596e1778481164a01def4100ce7005444b3` |

The [focused run](../../.cartulary/test-results/gp12-focus/run-summary.json)
passed the pure, bounds, semantic, schema, PostgreSQL and restore rows. The
old selection remained readable after the failed rebuild rolled back its
clear. No replacement was visible. The corrected selection rebuilt, published again
idempotently and read byte-exactly. Valid GP-11 vectors and the earlier
direction, ordering and label migrations retain their pinned bytes. A
deployment owner must correct invalid input, publish from authority and update
the selected binding through its lifecycle. Rollback requires a matching
writer and selection followed by requalification.

The [final focused run](../../.cartulary/test-results/gp12-focus2/run-summary.json)
passed all six rows after consolidating the direction policy capabilities and
adding filtered input and oversized registry-union controls. S-026 and S-027
close here. The four malformed results failed the pure proof and publication
before any SQL execution. All three tables stayed empty. Independently seeded
content failed exact, locked, ordered and traversal reads, and the binding-only
lock returned its binding. A 256-label result published and read exactly.

**E-130 — Frozen-source qualification.** [Finalization](../../.cartulary/test-results/gp12-finalize/unit-artifacts/finalize-summary.json)
refreshed zero generated files. Optional warm-run evidence maintenance skipped
because no compatible full warm-check root was supplied. The
[frozen inventory](../../.cartulary/evidence-manifests/graph-projection-gp12/frozen-source.json)
matches the pre-finalize and [ending inventories](../../.cartulary/evidence-manifests/graph-projection-gp12/end-source.json).
SHA-256 is `be0df2c15bee282873f331b0846dca6652f514c244decc5b3bf37a417b9187fe`
across 7,663 entries. The 7,660 entry paths and modes remain. GP-12 modified
18 existing source or contract paths and added two Graph implementation files
and a storage test input. The combined tree has 25 modified and eleven
untracked paths against HEAD, and the index still matches HEAD.

The [complete Graph selection](../../.cartulary/test-results/gp12-final-graph/run-summary.json)
passed eleven rows. Its [explicit audit](../../.cartulary/test-results/gp12-final-audit/test-evidence-audit/audit.json)
accepted one root covering all eleven, with zero rejected or unused inputs.
The final harness source digest is
`sha256:ecab06580ba6feb1fa917f4192b5bac47a72bc8ad843f58b29b04d3e3aa27895`.
Generated and schema source checks, Go and backend-boundary lint, and
`git diff --check` passed. The [native receipts](../../.cartulary/evidence-manifests/graph-projection-gp12/native-exits.json)
retain exact commands, run roots, native exits, cleanup and secret-scan status.

The four [Network Flow rows](../../.cartulary/test-results/gp12-final-nf/run-summary.json)
and both [Reporting rows](../../.cartulary/test-results/gp12-final-reporting/run-summary.json)
passed on the same source. Native exits, completed cleanup and retained-secret
scans passed across the final Graph-backed runs. No required repository gate
was skipped. No deployment database was inspected.

**E-131 — Final handoff.** T-041 closes last on native exit zero from
[final Markdown lint](../../.cartulary/test-results/gp12-handoff-final/adhoc/lint-markdown/tool-run-summary.json).
The [native receipts](../../.cartulary/evidence-manifests/graph-projection-gp12/native-exits.json)
include that exit separately. E-132 supersedes the historical deployment-inventory condition.

### GP-11 result and input admission

**E-122 — Entry.** At `9203107cd` the index matched HEAD. The entry had
25 modified and six untracked paths. All 7,658 non-Markdown source paths,
file modes and digests matched the GP-10 freeze. The Graph task guide passed.
[Entry status](../../.cartulary/evidence-manifests/graph-projection-gp11/entry-status.txt) ·
[Source inventory](../../.cartulary/evidence-manifests/graph-projection-gp11/entry-source.json)

**E-123 — Owner review and reproduction.** The independent owner review checked
Graph §§3–8, the Network Flow modes, and the Graph/Reporting/Recovery ownership
boundaries. The final wording makes metadata member positions explicit, applies
ordinary request-error precedence before exhaustion, and identifies when distinct
registered edge kinds prove reverse emission. The public storage boundary and
Recovery v5 interfaces retain their existing shapes.

| Finding | Owner and implementation | Failing witness and expected result |
| --- | --- | --- |
| F-052 · Mandatory output | Tables 3-H/I and §§6.1, 8. Count present mandatory fields per kind and namespace from missing and source-null policies. | Removing the first direct vertex's score from the aggregate fixture must reject. Defaults and permitted nulls count as present. A policy that can omit still permits absence. |
| F-053 · Relationship cardinality | Tables 3-D/G and §§6.3, 8. Index direct IDs and mapping claims, then verify unique reverse links and complete pairs where reverse emission is proved. | The independently ordered second direct edge for relation-1 must reject. Different sources may share a mapping, including self-loops and coincident output kinds. |
| F-054 · Reverse label capabilities | §§5, 6.1. Union each mapping's own contribution into its configured outputs. Defaults apply to all kinds and preservation is cumulative. | Shared forward output leaked an unrelated label. Shared reverse output lost a preservation flag. Both historic writer outputs matched independently pinned bytes. |
| F-055 · Object labels | §§5, 6.1, 6.3, 8. Enforce nonempty labels, the UTF-8 byte ceiling, object count, strict byte order and own-kind membership. | An empty label on the first direct-fixture vertex must reject. Arbitrary source labels require a direct or reverse object and a source-preserving kind. Aggregate labels always come from the registry. |
| F-056 · Scalar order | Tables 3-A–J and §§4, 7. Visit inherited mapping members, source relationship scalars and aggregation fields in owner order. | At 100,000 prior invalid labels an invalid target scope exhausts the budget before an invalid input kind. At 99,999 prior issues the invalid identifier wins. |
| F-057 · Metadata ID ties | Tables 3-B/I and §4. Admit metadata mapping IDs in their own namespace with the existing shared duplicate budget before collection ceilings and digesting. | Two mappings with one ID and different keys must return one duplicate in either order. Distinct IDs with colliding expanded keys retain their mapping error. |

The [focused red run](../../.cartulary/test-results/gp11-red/run-summary.json)
reproduced all six gaps. The [historical writer run](../../.cartulary/test-results/gp11-migration-red/run-summary.json)
reproduced the two old migration envelopes. The three malformed result IDs are
`gpres_8dc9378b8352ec844a06a7b1525f9db94d388620295773fa3755d46b4952ede4`
(missing score), `gpres_13542272cb46c0fa083a7e9a5b2ee1cd20d0f109629eee19cd0de23afc7c5b2c`
(duplicate relationship), and `gpres_5e2830942681e889299c2a8d001b754edbd93b76f116110070beb8bced6b7bab`
(empty label). The storage gate checks early rejection before any SQL execution,
all three table counts, independently seeded exact and locked reads, ordered
vertex and edge reads, traversal, and binding-only locking.

**E-124 — Migration inventory.** Network Flow Activity is the sole registered
production owner in this source tree. Its default and time-bucket selectors
each use six distinct metadata mapping IDs, mapping-only labels, and reverse
emission disabled. Neither selector can produce either shared-kind migration.
No deployment database was inspected. The generic shared-kind fixtures exercise
both affected selections. Their normalized configuration and source digests,
ordered vertex and edge bytes, object IDs, and sort keys remain exact.
[Full old and new bindings, canonical envelopes, ordered rows, and inputs](../../internal/app/graphassembly/testdata/label-capabilities.json)

| Selection | GP-10 result ID | GP-11 result ID | Reader and owner impact |
| --- | --- | --- | --- |
| Shared forward kind, zero emitted edges | `gpres_27fde7331df25273564b6223504c9de5e4d8947d11939b1e603349aeb5d89626` | `gpres_442d228ef27255de0e4fcc2033bc8e3fbb55ee10c5684cea1f19635cecc410df` | The reverse kind drops the unrelated label. Both envelopes pass content admission. Reprojection changes selection identity. |
| Shared reverse kind, emitted source label | `gpres_122b07f4110445ac9c34ed8b44f716b5399b86931e3bf54a6d228e73623fce39` | `gpres_a82a881ec30c0950d18dfdfcb2b559831587678cdd2aa036f536a4c69a3a3670` | The reverse kind regains source-label preservation. The old envelope contradicts its own registry and fails content admission. |
| Existing single-mapping reverse fixture | `gpres_b356854df17070a45dbf005adfc715170cab59b2e5badb022dea0dfe1854a586` | Same | Retained GP-10 direction correction remains byte-exact. |

An additional deployed source owner must inventory selected configurations and
authoritative inputs before rollout. Correct duplicate metadata mapping IDs at
the source. Reproject, publish and update each changed expected binding through
its owning lifecycle before enabling stricter reads for that selection. Missing
authoritative input blocks the affected selection. Stored Graph rows are
immutable. Reporting retains already admitted contributions. Rollback requires a
writer and owner-selected bindings that reproduce each other and a fresh qualification.

**E-125 — Focused repair and freeze.** The [first focused run](../../.cartulary/test-results/gp11-focus/run-summary.json)
passed the PostgreSQL, reconstruction, semantic, contract and pure rows. Its
bounds fixture repeated the same metadata ID at capacity and correctly returned
9,999 duplicates before collection admission. Giving the capacity fixture
distinct IDs preserves its intended assertion. The duplicate-precedence case
remains separate. The [second focused run](../../.cartulary/test-results/gp11-focus2/run-summary.json)
passed all six rows. S-027 closes on its exact public verdicts. S-026 closes
on the three storage witnesses, producer capability cases and old/new restore
pairs. Failed rebuilds roll back the clear and publish no replacement.
Successful rebuilds pass idempotent publication and byte-exact reads.

[Finalization](../../.cartulary/test-results/gp11-finalize/unit-artifacts/finalize-summary.json)
exited zero. Its optional retained-run maintenance skipped with
`results-dir-not-provided` because no compatible full warm-check root was
supplied. The [frozen inventory](../../.cartulary/evidence-manifests/graph-projection-gp11/frozen-source.json)
matches the pre-finalize inventory. Its SHA-256 is
`674ae2f78229599b5925ead670dd788263ed3409acd2880837e022132f13a943`.
It contains 7,660 paths. All 7,658 entry paths and modes remain. GP-11
changed 13 existing non-Markdown files and added the two test files for
completed-result claims and paired label migration fixtures. The Graph NLSpec
and this tracker are the documentation changes.

**E-126 — Frozen-source qualification and final handoff.** The final runs
share harness source digest
`sha256:4e50b943e4167d2f0e29694be734e4e0d81444d98f998f90f33a0a3d450d6c02`.
The [ending inventory](../../.cartulary/evidence-manifests/graph-projection-gp11/end-source.json)
matches the freeze exactly. The working tree has 25 modified tracked paths
and eight untracked paths, including the two GP-11 test additions. The index
matches HEAD. The combined diff and untracked paths were reviewed against
entry. `git diff --check` passed.

| Frozen-source gate | Receipt |
| --- | --- |
| Graph complete owner, eleven passed rows | [Graph selection](../../.cartulary/test-results/gp11-final-graph/run-summary.json) |
| Explicit owner audit, one accepted root, eleven required rows, zero rejected or unused | [Manifest](../../.cartulary/evidence-manifests/graph-projection-gp11/graph-owner-roots.json) · [Audit](../../.cartulary/test-results/gp11-final-audit/test-evidence-audit/audit.json) |
| Network Flow default and time-bucket bindings, saved lifecycle and admission | [Four owner rows](../../.cartulary/test-results/gp11-final-nf/run-summary.json) |
| Reporting exact selection, admission and rendering | [Two owner rows](../../.cartulary/test-results/gp11-final-reporting/run-summary.json) |
| Generated and schema checks | [Generated](../../.cartulary/test-results/gp11-final-generated/run-summary.json) · [Schema](../../.cartulary/test-results/gp11-final-schema/run-summary.json) |
| Go and backend-boundary lint | [Go](../../.cartulary/test-results/gp11-final-go/run-summary.json) · [Boundary](../../.cartulary/test-results/gp11-final-boundary/run-summary.json) |
| Final tracker text | [Markdown lint](../../.cartulary/test-results/gp11-handoff-final/adhoc/lint-markdown/tool-run-summary.json) |

The [native exits and run roots](../../.cartulary/evidence-manifests/graph-projection-gp11/native-exits.json)
record each command, including failing development witnesses. All final
graph-backed runs completed cleanup and passed retained-secret scanning.
Historical numeric output and all four retained graph families passed.
Both production selector modes retain their bindings. No required gate was
skipped. T-041 closes last on exit zero from
`CARTULARY_TEST_RUN_ID=gp11-handoff-final make lint-markdown` on this text.

### GP-10 admission and direction registry

**E-117 — Entry.** On main at `9203107cd`, the index matches HEAD.
The [entry inventory](../../.cartulary/evidence-manifests/graph-projection-gp10/entry-source.json)
matched all 7,658 retained source paths and modes. The [status](../../.cartulary/evidence-manifests/graph-projection-gp10/entry-status.txt)
records 24 modified and six untracked paths. The entry tracked diff SHA-256 was
`9c08dab6b92531834de76d89fcfb634f600a04cd77ca41d1ae4635f127d16e9b`.
The Graph task guide exited zero. The authorized corrections cover §5 identity
framing, S-026 declarations, direction and mapping claims, S-027 scalar
exhaustion and definition uniqueness. T-041 closes after requalification.

**E-118 — Owner review and reproduction.** The independent review reproduced
the §5 example and checked Tables 3-A/C/F–J, Core 00/01/04, the Network Flow
source registration, Reporting §15 and Graph §9. The reviewed text keeps the
first-excess allocation guard and distinguishes reader membership from the
producer's exact configured union. Network Flow's two configured modes both use
`preserve`. The registered source owner is Network Flow Activity.

The [six-row red selection](../../.cartulary/test-results/gp10-red/run-summary.json)
exited 2 with product assertion failures. Both rebound P1 results and the
per-kind direction contradiction reached publication. Duplicate definition IDs
projected, the 100,001st label lost to a later request error, the schema allowed
empty directions, and the normalized writer and restore still produced the old
binding. The semantic-matrix row passed its own cases. Separate focused witnesses
exercise the policy table, mapping consistency and declaration namespaces.

**E-119 — Focused repairs and recovery.** The [first focused run](../../.cartulary/test-results/gp10-focus/run-summary.json)
exited 2. Five of six rows passed. A successful preceding restore had replaced
the test's original graph, so its rollback assertion looked for the wrong
selection. The migration test now publishes the old direction binding,
asserts its exact bytes before and after rollback and checks that the
replacement is absent. The corrected binding reconstructs, reads byte-exactly
and republishes idempotently. The [focused follow-up](../../.cartulary/test-results/gp10-focus2/run-summary.json)
exited zero on the pure, PostgreSQL and claimed-reconstruction rows.

| Finding | Owner and repair | Qualification obligation |
| --- | --- | --- |
| F-046 · direct identity | Graph §5 defines each byte of the mapping, ID and sort transcripts. Independent ASCII and multibyte pins agree with the fixed producer. | V-001 and restore |
| F-047 · emitted fields | Graph §§5/6/8. Separate indexed property and metadata declarations authorize each emitted value, including nullable output and typed array elements. | V-001/017/018/028, exact historical numeric bytes |
| F-048 · direct provenance | Graph §§3-F/6.3/8. One direct source ID per result, one output kind per source kind, consistent observed mappings and disjoint entity/relationship mapping IDs. | V-001/028, zero pre-rejection writes and seeded content reads |
| F-049 · directions | Graph §§3-G/6/8. The writer unions configured directions even on empty output. The reader checks own-kind membership. Packaged schema requires nonempty sets. | V-001/017/019/028, exact binding change and restore rollback |
| F-050 · issue ceiling | Graph §§4/7. Each scalar recorder returns the existing limit error on the first excess. The 100,000th invalid label still permits a later request verdict. | V-001/018, exact safe errors and zero result |
| F-051 · definition IDs | Graph Tables 3-A/3-H and §4. The cancellable uniqueness pass counts repetitions within each independent namespace before ceilings, expansion or digests. | V-001/018, order permutations and competing failures |

The independently rebound undeclared-metadata and duplicate-direct witnesses
are `gpres_7056a8dc64a2d291238301de31382886b78492f51a12b22f134f071dda451d71`
and `gpres_08ab26884128e22e5e28f64b0797e7e66a9d5787bfe0b422f47d86a8cdba9bd9`.
Together with the own-kind direction contradiction, they fail prior to the first
SQL execution with zero rows in the three tables. Seeded ordinary and locked
exact reads, both ordered reads and traversal reject. Binding-only locking succeeds.

**Migration inventory.** The sole registered production source owner, Network
Flow Activity, constructs both the default and time-bucket configurations with
fixed definition IDs, `preserve` and reverse emission disabled. These selectors
cannot produce an affected normalized-direction selection. The generic reverse
fixture is affected. Its output digest changes from
`0b68a8d17031188653464ad60e36eb63575cd2d7495c52c03ce1ecefbd79f98a` to
`ee71acff104428c64d31a48613a545900713ac9dbce854f0f373f689158f1b2f`.
Its result ID changes from
`gpres_57e5a323f6e8c51160008243fd83be74cdaf1d592021d459cb384981d013510c` to
`gpres_b356854df17070a45dbf005adfc715170cab59b2e5badb022dea0dfe1854a586`.
The normalized input digests and ordered child bytes match. Configurations
with several producers change only if their combined direction set changes.

Any additional source owner must inventory its authoritative inputs, correct
duplicate definition IDs, rematerialize changed selections and update expected
bindings through its owning lifecycle. Missing authoritative input blocks that
selection's rollout. Reporting retains already admitted contributions.
Rollback requires matching producer and owner-expected bindings and fresh
qualification. Graph v2 immutable storage and the Recovery v5 contract retain
their existing interfaces.

**E-120 — Frozen-source qualification.** `make agent-finalize` exited zero.
The [freeze](../../.cartulary/evidence-manifests/graph-projection-gp10/frozen-source.json)
matched the pre-finalize inventory and the [ending source](../../.cartulary/evidence-manifests/graph-projection-gp10/end-source.json)
entry for entry, including modes. All 7,658 entry paths remain. GP-10 changed
18 non-Markdown files in the approved Graph, schema, generated package and
reconstruction scope, plus the Graph owner and this tracker. The index matches
HEAD and the combined diff passed `git diff --check`.

Each final graph-backed invocation used cache mode `off` on harness source
`a007ca52c97e686568326c722c680622d1bc8c2505bf32f7f5dd43130ac75133`.
All units passed, cleanup completed and retained-secret scans passed.
[Exact commands, native exits, cleanup, scans and selected rows](../../.cartulary/evidence-manifests/graph-projection-gp10/native-exits.json)

| Gate | Receipt |
| --- | --- |
| Graph owner, eleven rows | [Full selection](../../.cartulary/test-results/gp10-final-graph/run-summary.json) |
| Complete Graph audit, one accepted root, eleven required rows, zero rejected or unused inputs | [Explicit manifest](../../.cartulary/evidence-manifests/graph-projection-gp10/graph-owner-roots.json) · [Audit](../../.cartulary/test-results/gp10-final-audit/test-evidence-audit/audit.json) |
| Network Flow fixed default and time-bucket bindings and saved lifecycle | [Four owner rows](../../.cartulary/test-results/gp10-final-nf/run-summary.json) |
| Reporting exact selection, admission and rendering | [Peer receipts](../../.cartulary/test-results/gp10-final-reporting/run-summary.json) |
| Generated and schema checks | [Generated](../../.cartulary/test-results/gp10-final-generated/run-summary.json) · [Schema](../../.cartulary/test-results/gp10-final-schema/run-summary.json) |
| Go and backend boundary lint | [Go](../../.cartulary/test-results/gp10-final-go/run-summary.json) · [Boundary](../../.cartulary/test-results/gp10-final-boundary/run-summary.json) |

The four Graph fixture families and historical numeric result passed on this
source with the single reverse-vector change recorded at E-119. Both Network
Flow binding controls remain exact. The peer review also recorded an independent
documentation discrepancy for its owner. Network Flow Table 14-G1 lists a
reverse kind with reverse emission disabled. The Network Flow producer omits
that member as required by Graph Table 3-G.

**E-121 — Final-text receipt.** T-041 closes last on native exit zero from
`CARTULARY_TEST_RUN_ID=gp10-handoff-final make lint-markdown` on this text.
[Final Markdown receipt](../../.cartulary/test-results/gp10-handoff-final/adhoc/lint-markdown/tool-run-summary.json)

### GP-09 closure checkpoints

**E-104 — Specification and tracker entry.** The reviewed working tree on main at `9203107cd76ba2849bf02dc466309b6697be3e93` contained 22 modified and three untracked files. The index matched the HEAD tree `7e206ded2ca9fe6b7c4e92d2088fedf28193ba8f`. The tracked patch SHA-256 was `3790bf0571f1df73a2f0d8405d660daff504f0febfd30e3d280fa7b8d4e3852a`, the Graph owner was `e923fe840381af08335b60f753456ff818e113491bd26d2e711f21a92c98c0cb` and the tracker was `2ecc35cef9d3e52fbcdd534994e773476c58511e34c6954799391eee71cbebae`. The historical numeric fixture and fixed vectors matched E-093. The 7,655-entry [entry inventory](../../.cartulary/evidence-manifests/graph-projection-gp09-closure/entry-source.json) matched E-099 byte for byte. The [prior audit](../../.cartulary/test-results/gp09r-final-audit/test-evidence-audit/audit.json) retains its eleven accepted rows on harness source `fb2607712277ffdd9d66342a94e60735c0b052faf64198803d456de367690007`.

| Owner | Contract boundary for review |
| --- | --- |
| §§3-J, 5, 6.3, 8 | The shared pure Graph proof recomputes direct-vertex identity, checks edge endpoint families, verifies each rule's groups and builds dependencies using all registered producers. Unobserved inputs stay with the source owner. |
| §§5–6 | The authored schema carries the closed local registry shapes and identifier repertoire. Graph proves byte ceilings, ordering and membership from stored content. A component-wise ordering correction can require source-owner rematerialization and atomic rejection of an old restore binding. |
| §§3–4, 7 | Shape precedes scalar admission throughout the request. Specific request errors keep their details and fixed priority. Counted errors retain severity, reason and shared budget. Endpoint existence precedes eligibility after earlier admission gates succeed. |

Entry DONE. The user instructed that the amendments apply and removed the separate approval gate. `CARTULARY_TEST_RUN_ID=gp09c-entry-guide make task-guide OWNER=module.graphprojection` exited 0 and selected the owner slice. Discovery has no retained run root. `CARTULARY_TEST_RUN_ID=gp09c-entry-lint make lint-markdown` exited 0 ([entry lint](../../.cartulary/test-results/gp09c-entry-lint/adhoc/lint-markdown/tool-run-summary.json)). Changed paths are the Graph owner and this tracker. Reproductions and qualified evidence will be recorded separately before each successor.

**E-105 — S-033 / F-042 DONE.**

Before repair both Validate and the borrowed publisher accepted overlapping groups. Graph now checks input scope, kind, distinct immediate IDs and the ordered endpoint-producer pair across one rule at a time. Changed result_v2_validation.go, its unit tests, engine_v2_test.go and postgresresult/result_proof_test.go. The managed witness checked zero SQL executions and rows, all five seeded content reads, and the successful envelope lock. Valid cross-rule reuse and conflicting group variants passed. No S-033 blockers.

- `CARTULARY_TEST_RUN_ID=gp09c-033-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 2 (fail). Source `ef9cb5dc0340cc99dfcf5e0c2d65d255f2108671a9fec039f05aca18b6baddb3`. [Receipt](../../.cartulary/test-results/gp09c-033-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-033-format make format` exited 0 (pass). Source `af2451da6f694dbfeb4679b0eb502758b5667f827f35feb17f9670d942f4b1b6`. [Receipt](../../.cartulary/test-results/gp09c-033-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-033-green make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.engine.v2_maximum_semantic_bounds,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0 (pass). Source `d665692d44a2eddaf7910d0b309803617b3d7a2c8551de9a1a400a7b62579bc7`. [Receipt](../../.cartulary/test-results/gp09c-033-green/run-summary.json)

**E-106 — S-034 / F-039 DONE.**

A rebound tail producing bridge passed the old content proof and publication even though final consumes that kind and tail consumes final. Both proof boundaries now use the same other-producers rule. All declared rules enter the bounded topological check, including declarations with no output. Changed engine_v2.go, result_v2_validation.go, its tests and postgresresult/result_proof_test.go. The directed cycle failed before SQL and on seeded reads. Direct input of the consuming rule’s own kind, shared kinds and unused declarations passed. No blockers.

- `CARTULARY_TEST_RUN_ID=gp09c-034-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 2 (fail). Source `6111b0374f4be4c0ee12a6e96cabd94911c5d856e1f1c90efa0ec18e911872e4`. [Receipt](../../.cartulary/test-results/gp09c-034-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-034-format make format` exited 0 (pass). Source `43c784db232655f35f956b45c75e26aa0f3fbe834fcc5867b2917d84f257fc74`. [Receipt](../../.cartulary/test-results/gp09c-034-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-034-green make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0 (pass). Source `e8a28635216d10c6d4224a780a1852d7a95529f93d023c3f5cafb48940ec21f9`. [Receipt](../../.cartulary/test-results/gp09c-034-green/run-summary.json)

**E-107 — S-035 / F-040 DONE.**

Changing a direct vertex source or sort key originally passed both proof and publication. The shared private direct-vertex tuples now prove its ID and sort key. The original entity-a and substituted other-entity reproduce their independently fixed IDs. Changed input.go, direct_mapping.go, result_v2_validation.go, result_vectors_v2_test.go and postgresresult/result_proof_test.go. Both malformed cases reject before SQL and on every seeded content read. The metadata lock and valid historical bytes passed. No blockers.

- `CARTULARY_TEST_RUN_ID=gp09c-035-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 2 (fail). Source `618bff6216637417fb8fd7a7690cd7414c12a75d95d0228511ac14f680892bbe`. [Receipt](../../.cartulary/test-results/gp09c-035-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-035-format make format` exited 0 (pass). Source `580d721148fb606d6219af9f882afdef7669078029c5cd8d72b29bd940385794`. [Receipt](../../.cartulary/test-results/gp09c-035-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-035-green make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0 (pass). Source `db7d4d0fa2ac5fa0f68e2174545c644c40cc1e44afb6ffb205b154cf572eb3dc`. [Receipt](../../.cartulary/test-results/gp09c-035-green/run-summary.json)

**E-108 — S-036 / F-041 DONE.**

The direct source and destination mutations and both correctly exchanged reverse-pair mutations originally passed publication. Content proof now resolves both endpoints to the required vertex family before verifying partners and groups. Changed result_v2_validation.go and postgresresult/result_proof_test.go. All four cases rejected before SQL and on every seeded content read; valid mixed fixtures and self-loops passed. No blockers.

- `CARTULARY_TEST_RUN_ID=gp09c-036-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 2 (fail). Source `3ded7a062e428894cde2ffd11fadd57868bdb86bf00fe27dec9e1954e647fb8e`. [Receipt](../../.cartulary/test-results/gp09c-036-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-036-format make format` exited 0 (pass). Source `193e614d88d0320922fffb347ba5003bb1c006afe4b49c308d6cedd1b4480525`. [Receipt](../../.cartulary/test-results/gp09c-036-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-036-green make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0 (pass). Source `2f2f7b549e8d02db85850f26dfb2977425df0fb56c82872ed504192297eb16a6`. [Receipt](../../.cartulary/test-results/gp09c-036-green/run-summary.json)

**E-109 — S-037 / F-045 DONE.**

The old engine reproduced the independent joined-key envelope and binding gpres_f46305cbb73d61e56d08c2bd59e5024f82ae9815aa6f08fdf0169439dcac9b51 exactly and accepted its stored order. Component comparison yields gpres_465bf6cbbb9cd6159f269cbd4ab3466d455c0874c5f0daa09656e39b50cedd8c with the same normalized input digests. The fixture covers a, aa and a literal delimiter collision in both arrays. Reordered input reproduced the corrected binding, restore with the historical expectation rolled back to exact prior state, and the corrected expectation rematerialized. Changed project.go, result_v2_validation.go, graphassembly/reverse_reconstruction_test.go and its testdata/registry-order.json. No blockers.

- `CARTULARY_TEST_RUN_ID=gp09c-037-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.claimed_graph_reconstruction_enumerates_and_publ_c18fddd44a` exited 2 (fail). Source `8324fcae8267dc2c53bb53a5274ba03b710b9b11c6d6781b83b7b4dd033d3c1d`. [Receipt](../../.cartulary/test-results/gp09c-037-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-037-format make format` exited 0 (pass). Source `c099f172b42dcfa6b28a63a98529e79fb382945818dd598e7d2db26ebd1ed8fb`. [Receipt](../../.cartulary/test-results/gp09c-037-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-037-green make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_deterministic_current_restore_rebuild,module.graphprojection.storage.claimed_graph_reconstruction_enumerates_and_publ_c18fddd44a,module.graphprojection.storage.restore_publication,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0 (pass). Source `7a31775712f02e3faff367cb19849951a3264ca332d38f86a01154483370b34f`. [Receipt](../../.cartulary/test-results/gp09c-037-green/run-summary.json)

**E-110 — S-038 / F-043 DONE. S-026 closed.**

The red packaged-schema and borrowed PostgreSQL selections admitted malformed registry declarations, including untyped nested entries and invalid identifiers. The authored schema, generated Graph projection, result_registry_v2.go, result_v2_validation.go, project.go and their contract/storage fixtures now require closed typed rows, ordered unique field references and an exact correspondence with expanded declarations. The pure proof enforces UTF-8 byte limits. Valid producer fixtures passed on the same selection, including declarations for empty output. Both static contract checks passed. No S-026 blocker remains.

- `CARTULARY_TEST_RUN_ID=gp09c-038-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.v2_contract_projection,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 2 (fail). Source `6a760d43f9b4f1e7d609afae691703f1d3e3862753d6396e7ed04563d3bd67d6`. [Receipt](../../.cartulary/test-results/gp09c-038-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-038-generate make generate` exited 0 (pass). [Receipt](../../.cartulary/test-results/gp09c-038-generate/generate/tool-run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-038-format make format` exited 0 (pass). Source `e5990322209a275ccc7deeb8047d0c01dca2fe66accf27fd6a58c8fcb5f9cbdf`. [Receipt](../../.cartulary/test-results/gp09c-038-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-038-green make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.storage.v2_contract_projection,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0 (pass). Source `38221ca80e68f9ccfeccf751ecdc9498e4abe6f2041cf40e57d49f02e93c6343`. [Receipt](../../.cartulary/test-results/gp09c-038-green/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-038-generated make source-check CHECK=generated` exited 0 (pass). Source `38221ca80e68f9ccfeccf751ecdc9498e4abe6f2041cf40e57d49f02e93c6343`. [Receipt](../../.cartulary/test-results/gp09c-038-generated/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-038-schema make source-check CHECK=schema` exited 0 (pass). Source `38221ca80e68f9ccfeccf751ecdc9498e4abe6f2041cf40e57d49f02e93c6343`. [Receipt](../../.cartulary/test-results/gp09c-038-schema/run-summary.json)

**E-111 — S-039 / F-044 DONE. S-027 closed.**

The V-001 red selection reported invalid_identifier before a malformed direction type, and duplicate_identifier before seven separate local configuration violations. engine_v2.go now validates the closed shape throughout the request before input_scalars_v2.go admits local values, using fixed member order, numeric row indices, sorted dynamic keys and bounded counted diagnostics. input.go, validation.go and result_registry_v2.go share the admitted domains. The pure and maximum-bound cases in engine_v2_test.go and input_counts_v2_test.go confirm both §4 examples, nested predicates and defaults, cross-request counts at 100,000 and first excess, cancellation, submitted endpoint existence, and unchanged accepted vectors. The semantic matrix passed. No S-027 blocker remains. Before the freeze, reconcile the S-038 reader with the already accepted inert definitions in Table 3-H.

- `CARTULARY_TEST_RUN_ID=gp09c-039-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 2 (fail). Source `1df2aeab6d5f9e56084f4faf1a9fc833e3bf4335ab0c00abd3cb5f4bf4f7da66`. [Receipt](../../.cartulary/test-results/gp09c-039-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-039-format make format` exited 0 (pass). Source `f5839a452caae942a36724ab383169393db4832ed0e6167840444a17b635ad87`. [Receipt](../../.cartulary/test-results/gp09c-039-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-039-green make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.engine.v2_maximum_semantic_bounds` exited 0 (pass). Source `488585e68ddf0ba8fc02bdf1fffca45c84624ff9ee4bce976d59f2fcff48cfdd`. [Receipt](../../.cartulary/test-results/gp09c-039-green/run-summary.json)

**E-112 — S-038 inert-declaration compatibility, DONE.**

A pre-freeze comparison of Table 3-H with the registry writer revealed an explicit target kind with no configured producer. The red public projection succeeded, but the new content proof rejected both completed property and metadata fixtures. Graph §6.1 now specifies the retained top-level declaration. result_registry_v2.go matches nested references for configured kinds and validates the complete local declaration for inert explicit targets. field_paths_v2_test.go exercises the previously accepted producer bytes through the completed-result proof. The green Graph, semantic, storage and contract selections passed. No additional identity migration or S-026 blocker remains.

- `CARTULARY_TEST_RUN_ID=gp09c-038-inert-red CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix` exited 2 (fail). Source `ae4a57dbbd113d48784e179a3d770e45f303a475f220216be442cdf86875bddb`. [Receipt](../../.cartulary/test-results/gp09c-038-inert-red/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-038-inert-format make format` exited 0 (pass). Source `48b47390b82173e744623ac39eb0064b1d4554fdfcd54354934c84b23c267f70`. [Receipt](../../.cartulary/test-results/gp09c-038-inert-format/run-summary.json)
- `CARTULARY_TEST_RUN_ID=gp09c-038-inert-green make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.storage.v2_contract_projection,module.graphprojection.storage.managed_service_assertions_for_result_v2_atomici_ff9eae0340` exited 0 (pass). Source `9b2b6977e6c9cc954613e15d23d13c9f8cb47202095a3076e3e40e8c39a9dbbb`. [Receipt](../../.cartulary/test-results/gp09c-038-inert-green/run-summary.json)

**E-113 — Combined source freeze.**

Finalization reported unchanged generated topology and zero generated changes. The 7,658-entry non-Markdown inventory, including modes, matched immediately before and after finalization (SHA-256 0c2a522cf857a96a21cb45b145ea6fa802de5e804df4098ac72dde363882b0be). The candidate harness source is 9b2b6977e6c9cc954613e15d23d13c9f8cb47202095a3076e3e40e8c39a9dbbb. No entry path was lost. The three additional source paths are the independent registry-order vector, the registry proof, and the request scalar walk. Retained source identities exclude Markdown. [Frozen inventory](../../.cartulary/evidence-manifests/graph-projection-gp09-closure/frozen-source.json)

- `CARTULARY_TEST_RUN_ID=gp09c-finalize make agent-finalize` exited 0 (pass). Source `9b2b6977e6c9cc954613e15d23d13c9f8cb47202095a3076e3e40e8c39a9dbbb`. [Receipt](../../.cartulary/test-results/gp09c-finalize/run-summary.json)

**E-114 — Final-source portfolio and explicit Graph audit, DONE.**

All selections below used the E-113 freeze. For each graph-backed invocation the native process exited zero, every unit passed, cleanup completed and the retained-secret scan passed. The explicit manifest contains only the new full Graph owner run. The audit accepted all eleven rows in one partition with zero rejected or unused inputs. Storage, restore publication, deterministic reconstruction and rollback are included. [Manifest](../../.cartulary/evidence-manifests/graph-projection-gp09-closure/graph-owner-roots.json) · [Audit](../../.cartulary/test-results/gp09c-final-audit/test-evidence-audit/audit.json) · [Native exits, rows and scan counts](../../.cartulary/evidence-manifests/graph-projection-gp09-closure/native-exits.json)

| Final selection | Native exit | Units passed | Cleanup completed | Evidence |
| --- | --- | --- | --- | --- |
| `CARTULARY_TEST_RUN_ID=gp09c-final-graph make test-slice OWNER=module.graphprojection` | 0 | 11 | 29 | [Receipt](../../.cartulary/test-results/gp09c-final-graph/run-summary.json) |
| `CARTULARY_TEST_RUN_ID=gp09c-final-nf make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.time_bucket_saved_graph_lifecycle,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend` | 0 | 4 | 15 | [Receipt](../../.cartulary/test-results/gp09c-final-nf/run-summary.json) |
| `CARTULARY_TEST_RUN_ID=gp09c-final-reporting make test-slice OWNER=module.reporting ROWS=module.reporting.integration.graph_input_admission_reads_the_exact_selected_c_4c42af0393,module.reporting.unit.graph_v2_exact_selection_and_rendering_f54d8b1a62` | 0 | 4 | 15 | [Receipt](../../.cartulary/test-results/gp09c-final-reporting/run-summary.json) |
| `CARTULARY_TEST_RUN_ID=gp09c-final-generated make source-check CHECK=generated` | 0 | 5 | 14 | [Receipt](../../.cartulary/test-results/gp09c-final-generated/run-summary.json) |
| `CARTULARY_TEST_RUN_ID=gp09c-final-schema make source-check CHECK=schema` | 0 | 3 | 12 | [Receipt](../../.cartulary/test-results/gp09c-final-schema/run-summary.json) |
| `CARTULARY_TEST_RUN_ID=gp09c-final-go make lint CHECK=go` | 0 | 6 | 15 | [Receipt](../../.cartulary/test-results/gp09c-final-go/run-summary.json) |
| `CARTULARY_TEST_RUN_ID=gp09c-final-boundary make lint CHECK=backend-boundary` | 0 | 4 | 13 | [Receipt](../../.cartulary/test-results/gp09c-final-boundary/run-summary.json) |
| `CARTULARY_TEST_RUN_ID=gp09c-final-audit make test-evidence-audit OWNER=module.graphprojection EVIDENCE_ROOTS_FILE=.cartulary/evidence-manifests/graph-projection-gp09-closure/graph-owner-roots.json` | 0 | 1 | 10 | [Receipt](../../.cartulary/test-results/gp09c-final-audit/run-summary.json) |

**E-115 — Finding-to-evidence map, compatibility and rollback.**

| Finding | Accepted behavior and compatibility | Evidence |
| --- | --- | --- |
| F-039 | A projected input depends on every other registered producer of its kind in that scope. Shared output kinds and declarations with no rows remain usable. The cycle that passed when only one contributor's producer was followed now fails. | E-106, V-001/017/028 |
| F-040 | Direct vertex ID and sort key derive from the same mapping and source identity during emission and content admission. Rebinding an envelope cannot authenticate a substituted source. | E-107, V-001/017/028 |
| F-041 | Both endpoints of a direct or reverse edge must be direct vertices. The shared proof rejects either redirected endpoint even when the reverse pair agrees. | E-108, V-001/017/028 |
| F-042 | Groups from one rule agree on input scope, kind and ordered endpoint producers. A contributor can belong to one group per rule. Reuse by another rule remains valid. | E-105, V-001/017/018/028 |
| F-043 | Closed, ordered registry declarations and their nested property references agree. Schema consumers reject malformed local shapes and identifier characters. Graph verifies byte lengths and cross-object facts. A concrete definition for an unconfigured kind retains its inert declaration. | E-110/E-112, V-001/003/011/017/019/028 |
| F-044 | Wrong JSON types and conditional presence are decided across the request before scalar admission. Request errors precede counted scalar errors, then source-ID uniqueness and collection admission follow. Endpoint existence is checked before filtering. Competing invalid requests may now return an earlier reason. | E-111, V-001/017/018 |
| F-045 | Compare registry tuples component by component in both writer and reader. The independently pinned old and corrected bindings cover prefix order and literal delimiters. A mismatched reconstruction restores the previous selection atomically. | E-109, V-001/010/012/013/017/028 |

The four fixed Graph result families retained their exact bindings, envelopes and ordered child bytes. Historical numeric rows passed ordinary and locked exact reads and reprojection. The Network Flow default and time-bucket binding tuples remained exact. E-109 records the intentional compatibility exception for inputs whose old registry order differs from tuple order. Their configuration and source digests stay the same, but their result bytes and binding change.

The source owner must rematerialize an incompatible selected result and update its expected binding through the owning lifecycle. Graph content admission fails closed in the interim. A binding-only lock can still acquire the selection. Reporting retains its already admitted inputs. The restore operation rolls back if the rebuilt binding differs. There is no stored-row rewrite, compatibility reader, public API change or Recovery version change.

For a rollback, take S-039 out independently. Revert the S-038 proof and local schema together, including its E-112 clarification. Regenerate its matching authored contract with `make generate`. Revert the S-037 writer and reader tuple comparison together and restore the owner's matching expected binding. If undoing the result proof, revert S-036 and S-035 independently, then remove S-034 before S-033 because kind-wide dependencies use the per-rule observations. Keep E-082–E-103 as the prior evidence. Requalify the rollback source before using it. A permissive reader would readmit the malformed rows captured by the red witnesses.

The result carries enough information to verify the recorded provenance. A declaration that emitted no groups supplies no observed input kind, grouping tuple or raw source values. Reconstructing those facts requires authoritative input. The temporary contributor set is bounded per rule by its input namespace, and scalar admission counts issues without retaining submitted values. Deployment and release qualification remain separate.

The ending source inventory matches E-113 entry for entry, including modes. Against the entry tree all 7,655 initial source paths were retained. The prior fixed-vector and historical numeric fixture bytes are unchanged. The twelve tracker sections and the 42,361-byte E-082–E-103 checkpoint history were preserved. The index matches HEAD on main at `9203107cd`. The combined diff has 24 modified paths and six new paths. `git diff --check` passed. [Ending inventory](../../.cartulary/evidence-manifests/graph-projection-gp09-closure/end-source.json)

**E-116 — WF-35 / T-041 final-text gate.**

The handoff draft passed `CARTULARY_TEST_RUN_ID=gp09c-handoff-draft make lint-markdown` with native exit 0. The completion condition for this final text is native exit 0 from `CARTULARY_TEST_RUN_ID=gp09c-handoff-final make lint-markdown` and a passing linked receipt. This closes T-041 last. E-113–E-115 record the frozen portfolio, the compatibility exception, rollback and the final diff review. No commit or staging was performed. [Draft receipt](../../.cartulary/test-results/gp09c-handoff-draft/adhoc/lint-markdown/tool-run-summary.json) · [Final-text receipt](../../.cartulary/test-results/gp09c-handoff-final/adhoc/lint-markdown/tool-run-summary.json)

### GP-09 planning handoff

| Evidence | Source or command | Observation and applicability |
| --- | --- | --- |
| E-076, session baseline | `git --no-optional-locks status`, `rev-parse`, `ls-files`, file identity and content comparison | On 2026-11-03, `main` and the GP-08 receipt matched the tree in §1. Index, working tree and non-ignored untracked state were clean. Tracker SHA-256 before writing was `dc62348d5fcbd765e5d7cbb0dc72adec09a3bf55a9625d24fe226a80a2e3f944`, a regular single-link file at the declared path. The pre-write recheck matched the fingerprint and file identity. Comparison material for 8,073 paths is in `/tmp/cartulary-gp09-tracker-vWvvl5gF`. |
| E-077, input encoding | `objects.go` builders, `parseProjectionInputV2`, `normalizeCanonicalValueV2`, `canonicalObjectMapV2` and `writeCanonicalJSON` | The member-list intermediate is discarded before both public digest computations. Optional-member presence, arrays and numeric spelling still affect identity. |
| E-078, collection counts | `validateProjectionInputSchemaV2` → `validateV2ResourceBounds` → typed parsing → `validateAdmittedRequest`. Compared both count lists with the private semantic limit registry. | §7 maps the nine duplicate checks. Later checks on different input collections and object emission have distinct reachability. |
| E-079, result content/admission | `projectedGraphResource`, `outputResource`, `Resource`, `CompletedResult`, `validResultObjectV2`, authored v2 schema and Graph §§5–8. Traced `postgresresult` publication/reads, Network Flow saved and Reporting reads, and Graph restore publication. | Both builders emit the same field set. Current direct-object proof validates a present reference and allows null in the member shape. The historical numeric fixture exercises valid stored bytes. These are source/fixture observations. GP-09 has no execution result for malformed content. |
| E-080, historical scope and future selection | GP-08 candidate, native exits and owner audit linked below. Inspected Graph, Network Flow and Reporting family rows and the Make task surface. | GP-08's full Graph audit covered eleven rows on its recorded source. GP-09 reuses the applicable selectors for the specific obligations in §8. |
| E-081, document verification | `CARTULARY_TEST_RUN_ID=gp09-tracker-lint-01 make lint-markdown`, installed prerequisites, repository root | Outer exit 0 and retained status `pass`. Content and file identity matched the baseline outside this tracker. The index SHA-256 remained `d013a9ff86cf69bde33f19376cce6173067b78ad354382750c485d3008917027`. Final-text run ID and receipt are in §12. [Draft lint receipt](../../.cartulary/test-results/gp09-tracker-lint-01/adhoc/lint-markdown/tool-run-summary.json) |

At planning handoff, S-022 required revalidation of HEAD, the target tree, schema and routed tests against E-076–E-080. E-082–E-083 record the execution entry and characterization. Maintain the typed contract from the adopted owner and keep Markdown out of executable evidence.

### GP-08 checkpoints and requalification

**Entry and authority.** The reviewed staged tree was `d3e2aeb03cac35082e15c815022d49803ebdfbea`. The user confirmed the revised Graph owner clauses before implementation. GP-07's frozen files, 13 process exits and 11-row audit remain historical. The GP-08 [checkpoint ledger](../../.cartulary/evidence-manifests/graph-projection-gp08/checkpoint-exits.json) records the exact Make arguments, result parents and observed outer exits. Commands use `CARTULARY_TEST_RUN_ID` equal to the recorded run ID. No test consumes this document.

| Finding | Failure on the reviewed behavior | Repair and focused evidence |
| --- | --- | --- |
| F-022 | `gp08-red-isolated-01` observed forward-only fields on the reverse edge and accepted its missing required field | Kind-specific property and metadata derivation after capacity admission; positive and negative field sets, source link and registry in `gp08-endpoints-focus-01` |
| F-023 | `gp08-repro-01` rejected the producer's valid reverse family | Order-independent linkage proof, corrected edge schema and vertex/edge admission; `gp08-endpoints-focus-01` and real publication, republication, exact read, rebuild and rollback in `gp08-rebuild-focus-02` |
| F-024 | Both exact reads of the historic envelope failed in `gp08-repro-01`, and reprojection changed its binding/bytes | Historical writer fixture and strict reads/reprojection in `gp08-numeric-storage-01`; grouping permutations and numeric boundaries in `gp08-endpoints-focus-01` |
| F-025 | `gp08-red-isolated-01` returned ordinary missing-key validation after issue exhaustion across rules | One invocation collector and fallible object admission; exact count, capped retained entries, zero objects, and unexecuted later rule in `gp08-arrays-focus-01` |
| F-026 | `gp08-repro-01` accepted conflicting later endpoint values and hid invalid or missing contributors | Complete ordered pairs, both policies, both sides and permutations in `gp08-endpoints-focus-01`; captured historical ambiguous binding rejects and rolls back in `gp08-rebuild-focus-02` |
| F-027 | `gp08-repro-01` returned retryable computation failure on one-shot callback cancellation | Public checkpoint, parent cancellation/deadline, stable cancellation and uninterrupted success in `gp08-combined-focus-01`; real rollback in `gp08-rebuild-focus-02` |
| F-028 | `gp08-repro-01` produced oversized lists/sets and admitted oversized child mapped metadata | Streaming array merges with separate list and distinct-set counting; exact, over, duplicate-heavy and forged-envelope coverage in `gp08-arrays-focus-01` |

**Fixture provenance.** The direct-only fixture was captured with the writer from the GP-07 **entry** index in an isolated repository-root Make run, `gp08-historical-capture-03`. Its SHA-256 is `61ab450789d91a8e14cba9e30cf709e378ff8edc019c1630b7f01abe6793b429`. The first two capture launches stopped in preparation for missing Node and TypeScript dependencies. The third used the installed pinned dependencies. A separate checkout of the reviewed GP-07 **final** source produced the two focused red witnesses above and captured the prior first-contributor binding in `gp08-restore-history-01`. Neither isolated checkout supplies final acceptance evidence.

**Intermediate failures.** `gp08-execution-focus-01` found a misplaced validator guard at compilation. `gp08-execution-focus-02` and `gp08-reverse-focus-01` selected rows containing other still-open findings. `gp08-rebuild-focus-01` rejected a rollback fixture under the assembly row's isolated-database policy. The corrected test uses the row's declared fixture. The combined semantic, bounds, contract and assembled reconstruction selection passed in `gp08-combined-focus-01`. The first final Go lint found six files needing formatting after later edits. `gp08-format-02` repaired them; `gp08-finalize-02` preceded the final freeze. These intermediate roots retain their own outcomes.

**Final qualification.** The 7,652 tracked non-document entries match their frozen content and modes. The inventory SHA-256 is `cd42ae3c0e3caf8dab0299d7fe395fb2da2dff5c18384da5cddd504dc0f73840`. All eleven managed runs below report source digest `sha256:ce3cb2e9e149e9763dfefdc16876d22b7e2219fc75957602e08b5a9066b7ac16`. Across these runs all 42 work units passed, all 157 cleanup steps completed and every retained-secret scan passed. The native ledger records the exact Make command and observed outer exit for each run, including final-text lint. Run roots are under `.cartulary/test-results/<run-id>`. [Frozen inventory](../../.cartulary/evidence-manifests/graph-projection-gp08/frozen-source.json) · [Native exits](../../.cartulary/evidence-manifests/graph-projection-gp08/native-exits.json)

| Qualification | Final run ID | Result |
| --- | --- | --- |
| Agent finalization and Go lint | `gp08-finalize-02`, `gp08-final-go-02` | Passed. The finalizer skipped optional retained-run maintenance without a matching full warm-check root. |
| Complete Graph slice and explicit owner audit | `gp08-final-graph-01`, `gp08-final-audit-01` | All 11 required rows accepted from the new Graph run, with zero rejected or unused inputs. [Explicit roots](../../.cartulary/evidence-manifests/graph-projection-gp08/graph-owner-roots.json) · [Owner audit](../../.cartulary/test-results/gp08-final-audit-01/test-evidence-audit/audit.json) |
| Network Flow | `gp08-final-nf-01` | Default and time-bucket binding fixtures reproduced exactly. Bounded selectors, cleanup races and both saved graph lifecycles passed. |
| Reporting, Recovery and server | `gp08-final-reporting-01`, `gp08-final-recovery-01`, `gp08-final-server-01` | Exact selection and rendering, managed admission, restore codecs and post-claim readiness passed. |
| Generated outputs, schema and backend boundaries | `gp08-final-generated-01`, `gp08-final-schema-01`, `gp08-final-boundary-01` | All selected static checks passed. |
| Final tracker and candidate | `gp08-final-markdown-01` | Final-text Markdown lint, staged diff inspection and comparison with frozen source. [Lint summary](../../.cartulary/test-results/gp08-final-markdown-01/adhoc/lint-markdown/tool-run-summary.json) · [Candidate tree](../../.cartulary/evidence-manifests/graph-projection-gp08/candidate-tree.txt) |

The qualification covers the seven corrections above and their named consumers. Storage and object-ID framing stay at v2, rebuild/result and source registry at v5, and implementation binding at v7. Corrected reverse fields can change result identities. An ambiguous generic aggregation needs distinct grouping keys when its source owner intends multiple endpoint pairs. The historical conflict fixture confirms that reconstruction fails atomically on an old binding.

### GP-07 execution evidence

E-070 / S-021 freeze: `CARTULARY_TEST_RUN_ID=gp07-s021-finalize-01 make agent-finalize` published pass with no generated changes; the outer status was not retained. After reviewing its terminal summary, repeated V-016 specifically to capture the native exit: `gp07-s021-finalize-02`, observed outer 0/pass, zero generated changes, all selected actions passed. No matching full warm-check root exists; optional canonical-evidence/scheduler maintenance was skipped (`results-dir-not-provided`), without launching a full check. Freeze: `/tmp/cartulary-graph-gp07-0n05g89d/frozen-source.json`, 7,645 non-document path/content/mode entries, ledger SHA-256 `a441383b9c5d21b9f96a43c9589c51792d6e788ea7257caf856d37599bcdbc0f`. The pre-existing staged binary patch still matches E-060 byte for byte.

S-021 entry: S-016–S-020 checkpointed. Revalidated combined scope, read-only Make row discovery and final obligations. No successful retained full `check` root matches the S-020 source/toolchain/system identity; omit optional finalizer retained-run maintenance rather than launch a redundant check. Finalize now, freeze executable source/routing, then qualify the combined diff.

E-069 / S-020 DONE: explicit four-component immediate provenance now orders candidates and references; lowercase-hex `csk_` strings preserve that order without a legacy hash comparator. Verified raw sources, direct/reverse projected inputs, shared/multi-stage aggregates, Unicode/prefixes, first/last/list/metadata results, exact key strings, permutations and exact restore/binding rejection. F-019 supplemental integral-number encoding now unifies scientific/decimal forms across signed-int64-representable binary64, retains full int64 precision and distinguishes the upper exclusive bound. `gp07-s020-format-01/02` passed. E-063 five-row Graph command at `gp07-s020-check-01` and E-065 three-row NF command at `gp07-s020-nf-01` (matching CARTULARY_TEST_RUN_ID) exited 0/pass on `1ca3f4453b01fc34df1d92cd8df2a2108926c3324cd247c217916893298d2cea`; 3/3 and 4/4 units, 12/15 cleanup steps, zero failed/blocked, secret scans pass. GP2-FIX-002 changes only the explained output digest/result ID; all other goldens and NF bindings remain exact. Checkpoint complete before S-021.

F-019 supplemental reproduction before S-020 closure: `gp07-s020-number-boundary-01` (cache off, V-001) exited 2; exact-run diagnostics confirm equivalent `1000000`/`1000000.0`/`1e6` and negative forms diverge because integral binary64 values switch to scientific display while int64 uses decimal. E-063 remains its observed pass, but final numeric acceptance requires closing this boundary and rerunning its obligations. Normalize only exactly integral finite binary64 values inside the signed-int64 domain during canonical writing; never round int64 inputs through float64 or relax typed integer admission.

S-020 golden accounting: `gp07-s020-semantic-01` pure/restore rows passed; only semantic fixture and its permutation failed (outer exit 2). The source/config digests, object counts and merged properties were exact. `gp07-s020-golden-proof-01` confirmed that replacing just two contributor-key byte strings recovers the originally frozen output digest; it still failed against the unregenerated packaged fixture. `gp07-s020-generate-01 make generate` exited 0, changing only the generated Graph artifact package; `gp07-s020-golden-proof-02` reran V-017 with cache off and passed (outer 0). Removed the one-time historical-byte proof; permanent assertions pin the new exact reference keys. Only GP2-FIX-002 output digest and dependent result ID changed in the authored fixture. Combined review additionally checks numeric equivalence at the fixed/scientific display transition before accepting S-020.

E-068 / S-020 reproduction: `CARTULARY_TEST_RUN_ID=gp07-s020-repro-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 2 (product assertion failure). Exact-run diagnostics show raw entity order `[aa b é a]`, relationship `[é aa b a]`, projected vertex `[b a aa é]`, projected edge `[aa é a b]` and reversed-edge variants, contrary to required `[a aa b é]`; first/last/ordered_list and mapped metadata all reflect hash order. Production source unchanged before reproduction.

S-020 entry: S-019/E-067 complete; contributor construction, hash sorting, source-reference serialization and merge consumers revalidated. The nonempty semantic contract golden contains one aggregate: only independently explained ordering/key-derived output identities may change; config/source digests and unaffected fixtures remain exact.

E-067 / S-019 DONE: `execution_state.go` owns all outputs/indexes and capacity-checks builders before direct/reverse/aggregate construction; projected resource errors propagate unchanged and excluded contributor scans check cancellation. Bounded emitter tests verify capacity sharing, no over-limit builder, immediate stop before later rules, exclusions and cancellation. Public tests verify actual maxima, direct+aggregate and reverse-edge overflow with no returned result. `gp07-s019-format-01/02` passed; `gp07-s019-check-01` runs E-063 five-row command and `gp07-s019-rollback-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.claimed_graph_reconstruction_enumerates_and_publ_c18fddd44a,module.graphprojection.storage.restore_publication` (matching CARTULARY_TEST_RUN_ID) exited 0/pass. Source `809504ef8d1ee3a4d117074fde0d3a72fc3ac1d4aa6ecc0156e890c4dbc1e400`; 3/3 and 4/4 units; 12/17 cleanup steps, zero failures/blocks; secret scans pass. In-bound goldens unchanged; object limits are not heap/performance certification. Checkpoint complete before S-020.

E-066 / S-019 reproduction: `CARTULARY_TEST_RUN_ID=gp07-s019-repro-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 2 (unit product assertion failure). Bounded emitter tests seed counts at limit−1 and attempt two outputs: vertex and edge emitters incorrectly return nil beyond capacity. A 2,048-item excluded contributor scan also misses the cancellation checkpoint. Exact-run `explain-run TARGET=test-slice DETAIL=logs` retains all three failures. No production change preceded this reproduction.

S-019 entry: S-018/E-065 complete; direct/reverse/aggregate append sites, late projected-output checks, error conversion and skipped-contributor cancellation inspected. Characterize first-over-limit emission without intentionally exhausting memory.

E-065 / S-018 DONE: one configured producer-context index now requires every actual read context for property/metadata definitions, including wildcard/reverse outputs. Cross-scope, default masking, shared producers, filtered/empty inputs and inert definitions covered. `gp07-s018-format-02` passed; `gp07-s018-check-02` repeats the E-063 five-row Graph command, and `gp07-s018-nf-02 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.integration.time_bucket_saved_graph_lifecycle` (matching CARTULARY_TEST_RUN_ID) both exited 0/pass. Source `e6931587a3115929ca6bd07198614fe4ee98b4523bee29ca4e9d1f494348e0f6`; Graph 3/3 units/12 cleanup steps, NF 4/4 units/15 steps; zero failed/blocked cleanup, secret scans pass. Unaffected goldens/default/time-bucket tuples unchanged. Checkpoint complete before S-019.

S-018 development diagnosis: `gp07-s018-check-01` exited 2 because new test fixtures used unsupported filter operator `eq` instead of `equals` and expected semantic validation for a lexical alias rejected earlier by strict decoding. Corrected those assertions/fixtures only; no weakening of production admission. V-023/V-027 passed at `gp07-s018-nf-01` (exit 0, 15 completed cleanup steps). Requalify changed test source before checkpoint.

E-064 / S-018 reproduction: `CARTULARY_TEST_RUN_ID=gp07-s018-repro-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 2, product assertion failure. Exact-run `explain-run TARGET=test-slice DETAIL=logs` confirms canonical entity→edge and relationship→vertex property/metadata paths reject, while wrong contexts succeed through defaults; additional projected-context/all-producer cases fail. No production change preceded reproduction.

S-018 entry: S-017/E-063 complete; definition validators, actual contributor resolvers, producer/reverse output kinds and wildcard expansion revalidated.

E-063 / S-017 DONE: shared fallible scalar tuple encoder and canonical numeric normalization implemented; typed integer admission unchanged. Added decimal/boundary groups, equivalent lexemes, null/missing distinction, invalid array endpoint/group rejection, exact endpoint/contributor/permutation assertions and numeric mixed restore with binding rejection. `gp07-s017-format-01 make format` and `gp07-s017-check-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_deterministic_current_restore_rebuild,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.engine.v2_maximum_semantic_bounds,module.graphprojection.storage.v2_contract_projection` (each with matching CARTULARY_TEST_RUN_ID) exited 0/pass. Check source `fe6a25c4d61cfa936b0a1b3caf0816cf4ff6f52273bb176a6cc8fd64d906b8cc`; 3/3 work units, 12 cleanup steps, zero failed/blocked, retained-secret scan pass. Canonical goldens unchanged; corrected generic numeric identities intentionally change. Checkpoint complete before S-018.

E-062 / S-017 reproduction: `CARTULARY_TEST_RUN_ID=gp07-s017-repro-01 CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` exited 2 (unit exit 10, product assertion failure). `make explain-run RESULTS_DIR=.cartulary/test-results/gp07-s017-repro-01 TARGET=test-slice DETAIL=logs` shows distinct 1.5/2.5 and finite-boundary groups returning 3 vertices/1 edge instead of 4/2; equivalent numeric lexemes change canonical results; array grouping/endpoint keys succeed incorrectly. Cleanup 10 completed, zero failed/blocked, no missing unit outputs. Initial diagnostic attempts using no target or backend-unit were usage errors, corrected without rerunning tests. Production source was unchanged before this observation.

S-017 entry: S-016/E-061 complete; raw source numbers, strict canonical writer, grouping/endpoint digest callers and restore fixture revalidated. Add regressions before production correction.

E-061 / S-016 DONE: adopted clauses clarified and all four findings mapped. `CARTULARY_TEST_RUN_ID=gp07-s016-markdown-01 make lint-markdown` exited 0; `adhoc/lint-markdown/tool-run-summary.json` reports pass. The 28-path staged index is preserved. S-016 checkpoint precedes S-017 entry.

E-060 / S-016 entry: baseline and accepted plan recorded before any implementation change. Source review identifies four gaps; reproductions remain unexecuted. Read-only Make discovery in the planning turn confirmed Graph and peer rows; revalidate affected selections during execution.


### GP-07 final-source validation and handoff

E-071 / final portfolio: repository-root commands used Bash on Linux 6.6.114.1-microsoft-standard-WSL2, pinned Go 1.27.3, Node 24.15.0 and pnpm 10.33.0, with harness-managed services where selected. HEAD is E-060 (dirty); all graph-backed final commands report source `sha256:1ca3f4453b01fc34df1d92cd8df2a2108926c3324cd247c217916893298d2cea`, toolchain `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f` and system `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`. Prefix each command below with `CARTULARY_TEST_RUN_ID=<run ID>`. Roots are under `.cartulary/test-results/`. Counts are passed work units, completed cleanup steps, and cache bypass/hit/miss (not numbers of tests). Native process exits were collected separately from terminal summaries.

| Obligation | Exact command | Run ID | Exit / summary | Units / cleanup | Cache |
| --- | --- | --- | --- | --- | --- |
| V-016 | `make agent-finalize` | `gp07-s021-finalize-02` | 0/pass | 1 / 10 | 1/0/0 |
| V-008 | `make test-slice OWNER=module.graphprojection` | `gp07-final-graph-01` | 0/pass | 11 / 27 | 7/3/1 |
| V-014/V-023/V-024/V-027 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.integration.graph_result_cleanup_races,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.integration.time_bucket_saved_graph_lifecycle` | `gp07-final-nf-01` | 0/pass | 5 / 17 | 4/1/0 |
| V-021/V-022 | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.graph_v2_exact_selection_and_rendering_f54d8b1a62,module.reporting.integration.graph_input_admission_reads_the_exact_selected_c_4c42af0393` | `gp07-final-reporting-01` | 0/pass | 4 / 15 | 3/0/1 |
| V-020 | `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c` | `gp07-final-recovery-01` | 0/pass | 1 / 10 | 0/0/1 |
| V-015 | `make test-slice OWNER=app.server ROWS=app.server.unit.post_claim_graph_activation_preserves_structural_f41e41a052` | `gp07-final-server-01` | 0/pass | 1 / 10 | 0/0/1 |
| V-003 | `make source-check CHECK=generated` | `gp07-final-generated-01` | 0/pass | 5 / 14 | 5/0/0 |
| V-011 | `make source-check CHECK=schema` | `gp07-final-schema-01` | 0/pass | 3 / 12 | 3/0/0 |
| V-025 | `make lint CHECK=go` | `gp07-final-go-01` | 0/pass | 6 / 15 | 6/0/0 |
| V-005 | `make lint CHECK=backend-boundary` | `gp07-final-boundary-01` | 0/pass | 4 / 13 | 4/0/0 |
| V-026 | `make test-evidence-audit OWNER=module.graphprojection EVIDENCE_ROOTS_FILE=/tmp/cartulary-graph-gp07-0n05g89d/graph-owner-roots.json` | `gp07-final-audit-01` | 0/pass | 1 / 10 | 1/0/0 |

Each graph-backed final run retained a manifest, terminal run/target summaries, unit results, completed cleanup and a passing retained-secret scan. No required unit was skipped; there were no missing declared outputs or failed/blocked cleanup steps. Harness-compatible hits are reused evidence. The complete Graph invocation includes the focused Graph obligations; they were not rerun separately after the freeze.

E-072 / complete owner: the [owner audit](../../.cartulary/test-results/gp07-final-audit-01/test-evidence-audit/audit.json) accepted only `gp07-final-graph-01` for all 11 required Graph rows (43 routed top-level selectors); no rejected or unused inputs. Its accepted graph digest is `sha256:4c7527fb2b658d921b4c535aee231c3003d0c029547e175581b2e787d948da7a`. The exact observed root was supplied through the external JSON manifest, independent of this tracker. Copies of the [audit input](../../.cartulary/evidence-manifests/graph-projection-gp07/graph-owner-roots.json) and [frozen file ledger](../../.cartulary/evidence-manifests/graph-projection-gp07/frozen-source.json) are retained with the workspace evidence.

E-073 / preservation and compatibility: all 8,064 original paths remain. Compared to E-060, 16 original files changed only within the affected surfaces in §2; the two new private Graph files bring the total to 8,066 (41 under Graph). The 28-path staged binary diff and every index entry (mode/object/stage/path) are identical to the entry snapshot. The logical index digest is `2890b589cb49603f10517636a42a96680d71c17ff2a11f7f4988545de3072909`; raw index metadata was refreshed, without changing the staged content. The GP-05/GP-06 execution and final-evidence sections retain their original bytes. Both worktree and staged whitespace checks pass. GP-07 has no routing changes: only the authored GP2-FIX-002 fixture and its Make-regenerated embedded representation change in the contract projections.

V-023/V-027 retain the complete production binding fixtures, including default result `gpres_775aade676fd013e8c86b573093f9f4362fa95f375a069ca3fd90af74b7c1a8b` and time-bucket result `gpres_d52be3bc28fa17b5aa610b184c394e2797c11172923c41adcb7897ca26c49eea`. Graph v2 interfaces and wire shapes, ID framing, restore algorithm/result v5, registry v5 and binding v7 remain unchanged. Corrected generic definitions can reject; numeric grouping and ordered provenance can change bytes and result identity. Exact reconstruction and real transaction rollback remain required; nothing migrates or rewrites stored bindings. Roll back coherent authored changes and regenerate dependent projections through Make, preserving pre-existing work.

Object-count enforcement bounds emitted vertices/edges, not heap usage. Browser, performance, deployment, full-repository and release qualification remain outside this scope; the sole optional omission is the retained warm-check maintenance in E-070.

E-074 / documentation: `CARTULARY_TEST_RUN_ID=gp07-final-markdown-01 make lint-markdown` exited 0, tool status pass with zero failures after recording the final portfolio. The acceptance gate on the completion text itself is the separate `CARTULARY_TEST_RUN_ID=gp07-final-markdown-02 make lint-markdown`. Its [final-text summary](../../.cartulary/test-results/gp07-final-markdown-02/adhoc/lint-markdown/tool-run-summary.json) and [native exit ledger](../../.cartulary/evidence-manifests/graph-projection-gp07/native-exits.json) are the authoritative result of that last check. Completion requires exit 0/pass on this exact text with no intervening or subsequent edits.

E-075 / S-021 handoff: S-016–S-021, WF-24–WF-29 and T-029–T-034 are DONE. Every reported defect has a pre-fix reproduction and an implemented correction. Final acceptance includes the complete Graph owner audit, affected peer/static checks, final-text documentation qualification and readback of all 7,645 frozen path/content/mode entries. No required check is omitted. Stage the complete reviewed changes together, including the two new files; committing the earlier index alone would omit this remediation. No staging or commit was performed in GP-07.

### GP-06 execution checkpoints

S-015 entry: E-053 recorded. Revalidated combined source, Graph/NF/Reporting/Recovery/server Make guidance and exact §8 row membership. V-027 adds the missing time-bucket obligations. Finalize, review generated provenance, freeze executable files, then run the complete final portfolio; optional warm-root maintenance remains skipped without a compatible full warm-check root.

S-014 entry: E-049 recorded; lexical/context validators, source/projected readers, inclusion predicates and NF producer paths revalidated. Reproduce canonical-name rejection and forbidden-alias acceptance across path consumers before changing production.

S-013 entry: E-047 recorded; ordered rules, two execution loops, source references, merge/endpoint handling and restore candidate proof revalidated. Add nonempty chain reproduction before replacing execution.

S-012 entry: E-044 checkpoint recorded; ordering, endpoint validation, contributor scopes and Graph §§3-J/4 revalidated. Characterize scope collisions using canonical property paths before production changes.

| Evidence / slice | Command or inspection | Outcome / next gate |
| --- | --- | --- |
| E-043 / S-011 entry | Revalidated root/backend/tooling/test instructions, baseline, Graph §§3-J/4/5, affected source and current NF producer | S-011 IN_PROGRESS. No new reproduction or check pass yet. Current NF produces no engine aggregation rules and uses property/metadata paths. |
| E-044 / S-011 completion | Graph §§3-J/4/5 clarified; findings/sequence/compatibility and supersession recorded. `make task-guide OWNER=module.graphprojection` exit 0; `CARTULARY_TEST_RUN_ID=gp06-s011-markdown-01 make lint-markdown` | Outer exit 0/pass; `.cartulary/test-results/gp06-s011-markdown-01/adhoc/lint-markdown/tool-run-summary.json`. S-011 DONE before S-012. |
| E-045 / S-012 reproduction setup | `CARTULARY_TEST_RUN_ID=gp06-s012-repro-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` | Outer exit 2; infra/preflight_error: test referenced a named config field as embedded. Corrected the test reference; no production behavior exercised. Exact root `.cartulary/test-results/gp06-s012-repro-01`, summary `run-summary.json`; cleanup 10, zero failed/blocked. |
| E-046 / S-012 reproduced | `CARTULARY_TEST_RUN_ID=gp06-s012-repro-02 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` | Outer exit 2; product/test_assertion_failure. Both scope-collision cases rejected as invalid_aggregation_rule; other admission cases passed. Production unchanged. Exact root `.cartulary/test-results/gp06-s012-repro-02`, `run-summary.json`; cleanup 10, zero failed/blocked. |
| E-047 / S-012 completion | Scope/kind producer index, deduplicated dependencies, endpoint admission and sorted ready-batch topological traversal. `make format` run `gp06-s012-format-01`; V-001/V-017–V-019 combined `make test-slice` run `gp06-s012-check-01` with exact §8 row IDs | Both outer exits 0/pass. Roots `.cartulary/test-results/<run-id>`, summaries `run-summary.json`; canonical goldens unchanged. S-012 DONE before S-013 entry. |
| E-048 / S-013 reproduced | `CARTULARY_TEST_RUN_ID=gp06-s013-repro-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` | Outer exit 2; product/test_assertion_failure: both nonempty chains emitted 3 vertices/1 edge rather than 4/2; downstream endpoint/grouping failures were missed. Production execution unchanged. Root `.cartulary/test-results/gp06-s013-repro-01`, `run-summary.json`; cleanup 10, zero failed/blocked. |
| E-049 / S-013 completion | One ordered pass with private per-rule emitters; contributor snapshots, shared available-object/endpoint state and once-built rule metadata. Mixed-chain values/refs/bytes/permutations, failure policies and exact restore/mismatch coverage. `make format` run `gp06-s013-format-01`; combined V-001/V-010/V-012/V-013/V-017–V-019 run `gp06-s013-check-01` | Both outer exits 0/pass. Roots `.cartulary/test-results/<run-id>`, `run-summary.json`. Real assembly/writer rollback rows pass; goldens unchanged. S-013 DONE before S-014 entry. |
| E-050 / S-014 reproduced | `CARTULARY_TEST_RUN_ID=gp06-s014-repro-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` | Outer exit 2; product/test_assertion_failure. Canonical source-kind paths rejected; bare kind accepted across filters/inclusion/definitions/grouping/endpoints. Exact root `.cartulary/test-results/gp06-s014-repro-01`, `run-summary.json`; cleanup 10, zero failed/blocked. Production unchanged at reproduction. |
| E-051 / S-014 coverage refinement | `gp06-s014-format-01/02/03/04` Make formatting; `gp06-s014-check-01/02` combined V-001/V-010/V-017–V-019; `gp06-s014-nf-01` V-014/V-023; `gp06-s014-temporal-01` V-027 | Observed outer exits 0/pass. Check-01/02 cleanup 12 each; NF/temporal cleanup 15 each, zero failed/blocked. Added graph-view/projected-context negatives and strengthened the chain to consume each preceding aggregation's actual output; subsequent focused check required. |
| E-052 / production binding baseline | Isolated `/tmp/cartulary-gp06-producer-baseline` at HEAD plus saved `starting.diff`, with temporary test-only binding capture. `gp06-s014-baseline-bindings-03` runs V-023 and V-027's unit row | Outer exit 0/pass; cleanup 10, zero failed/blocked. All production Go files equal the pre-GP-06 baseline. Two temporary test log additions and one unrelated CSV checkout line-ending difference are the only baseline content differences. This is characterization, not final-source acceptance. |
| E-052 / baseline capture and setup accounting | Same isolated checkout: `gp06-s014-baseline-capture-01` intentionally fails the two capture probes to retain their nine-member bindings in the published runner log | Outer exit 2, normalized product/test_assertion_failure; intentional capture, not a discovered product defect. Cleanup 10, zero failed/blocked. The captured tuples become fixed assertions in both existing production fixture tests. Earlier attempts `gp06-s014-baseline-bindings-01/02` exited 2 before graph execution for missing local Node/TypeScript; public `make bootstrap PART=node` (root `20261103T232138Z-p44109`) and `PART=frontend` (`gp06-s014-baseline-frontend-01`) both exited 0/pass. No main-checkout dependencies changed. |
| E-053 / S-014 completion | Shared lexical classifier and context-specific admission/resolution; inclusion predicates receive the same scoped semantic validation as filters. Bare kind and projected-kind substitutes reject before execution/defaults. `gp06-s014-check-03` repeats V-001/V-010/V-017–V-019 for the strengthened mixed-chain tests; `gp06-s014-bindings-01` runs V-023 plus V-027 unit assertions | Both outer exits 0/pass; cleanup 12/10, zero failed/blocked. Exact nine-member default and time-bucket bindings equal pre-GP-06 captured constants, including all three digests and result IDs. Existing Graph goldens unchanged. F-018/S-014 DONE before S-015 entry. |
| E-054 / S-015 finalization and freeze | `CARTULARY_TEST_RUN_ID=gp06-s015-finalize-01 make agent-finalize` | Outer exit 0/pass; `run-summary.json` and `unit-artifacts/finalize-summary.json`. Zero generated files changed; existing topology index equals GP-06 entry content and still derives from the GP-05 authored selector. Cleanup 10, zero failed/blocked. Frozen-source ledger `/tmp/cartulary-graph-gp06-nsfrn4_z/frozen-source.json` contains 7,643 non-document entries. Optional canonical-evidence/scheduler drift maintenance is `not_selected/results-dir-not-provided`: no compatible full warm-check root exists, and no redundant broad check was launched. |

S-015 requires V-016, then final frozen-source V-008, V-014/V-015, V-020–V-025, V-003/V-005/V-011/V-007, V-027 and V-026. Use explicit fresh run IDs and non-Markdown evidence manifest. Earlier compatible results count only when the harness admits them. No broad warm check solely for optional finalizer maintenance.

### GP-06 final-source validation and handoff

E-055 final portfolio uses repository-root `CARTULARY_TEST_RUN_ID=<run ID>` invocations after E-054. The executable source digest is `sha256:b52449efafba18f1ca2623911a015cfb385380a7c0a60ad45acb2834ddd77fce`, HEAD `2dc1519246657929dc7194de9db22987597f9a92`, dirty worktree. Toolchain digest is `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`; system digest is `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`. Environment is Bash/Linux/WSL2, pinned Go 1.27.3, Node 24.15.0, pnpm 10.33.0, with harness-owned isolated services where selected. Each `run-manifest.json` records exact selections, command identity and timestamps. Source/evidence compatibility excludes Markdown; human owner review establishes specification fidelity.

All roots in the table are under `/home/jochi/code/cartulary/.cartulary/test-results/` and retain `run-summary.json`, `run-manifest.json`, target/row/unit evidence, `cleanup-results.json` and `retained-secret-scan.json`. Process exits below were observed from their owning native sessions, separately from summaries. Cache triples are bypass/hit/miss work units, not test counts. Normal compatible cache reuse is not claimed as fresh execution.

| Obligation | Exact command | Run ID | Outer exit / summary; cleanup; cache |
| --- | --- | --- | --- |
| V-008 | `make test-slice OWNER=module.graphprojection` | `gp06-final-graph-01` | 0/pass; 27 steps, zero failed/blocked; 7/3/1 |
| V-014/V-023/V-024/V-027 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.integration.graph_result_cleanup_races,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.integration.time_bucket_saved_graph_lifecycle` | `gp06-final-nf-01` | 0/pass; 17 steps, zero failed/blocked; 4/1/0 |
| V-021/V-022 | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.graph_v2_exact_selection_and_rendering_f54d8b1a62,module.reporting.integration.graph_input_admission_reads_the_exact_selected_c_4c42af0393` | `gp06-final-reporting-01` | 0/pass; 15 steps, zero failed/blocked; 3/0/1 |
| V-020 | `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c` | `gp06-final-recovery-01` | 0/pass; 10 steps, zero failed/blocked; 0/0/1 |
| V-015 | `make test-slice OWNER=app.server ROWS=app.server.unit.post_claim_graph_activation_preserves_structural_f41e41a052` | `gp06-final-server-01` | 0/pass; 10 steps, zero failed/blocked; 0/0/1 |
| V-003 | `make source-check CHECK=generated` | `gp06-final-generated-01` | 0/pass; 14 steps, zero failed/blocked; 5/0/0 |
| V-011 | `make source-check CHECK=schema` | `gp06-final-schema-01` | 0/pass; 12 steps, zero failed/blocked; 3/0/0 |
| V-025 | `make lint CHECK=go` | `gp06-final-go-01` | 0/pass; 15 steps, zero failed/blocked; 6/0/0 |
| V-005 | `make lint CHECK=backend-boundary` | `gp06-final-boundary-01` | 0/pass; 13 steps, zero failed/blocked; 4/0/0 |
| V-026 | `make test-evidence-audit OWNER=module.graphprojection EVIDENCE_ROOTS_FILE=/tmp/cartulary-graph-gp06-nsfrn4_z/graph-owner-roots.json` | `gp06-final-audit-01` | 0/pass; 10 steps, zero failed/blocked |

E-056 complete-owner acceptance: [audit.json](../../.cartulary/test-results/gp06-final-audit-01/test-evidence-audit/audit.json) accepts exactly the explicit `gp06-final-graph-01` root for all 11 required Graph owner rows, with no rejected or unused inputs. The JSON root manifest is an external human-authored selection artifact, never derived from this Markdown. Catalog/routing digests in the audit bind the current 43 top-level selectors and their new subtests. Peer results remain scoped to their selected rows; no complete-peer or release claim is implied.

E-057 preservation review: the finalizer reports zero generated changes. All original 8,061 paths remain, with three new authored Graph files, for 8,064 paths. Compared with the GP-06 entry baseline, only ten existing files changed (Graph specification/tracker; aggregation, engine, engine tests, project, validation, restore service test; two NF fixture tests). The other original dirty files remain byte-identical. The logical index is unchanged; no commit or staging occurred. The combined diff retains the pre-existing GP-05 repairs and Make-generated selector projection. Final readback confirms all 7,643 frozen executable hashes match and the complete GP-05 implementation/final-evidence ledger remains byte-identical to the entry copy. No unrelated content change or missing baseline path was found.

Compatibility and rollback: corrected mixed-scope inputs can change normalized configuration, output and result identity; forbidden kind paths reject. Exact unaffected Graph goldens and both NF production tuples remain asserted. Public v2 interfaces, framing and restore generations are unchanged. Exact expected-binding mismatch still fails closed; nothing rewrites stored identities, deletes results or resets environments. Rollback is a coherent authored revert with Make regeneration of dependent routing, preserving unrelated work.

Optional finalizer retained-run maintenance remains skipped for the documented missing warm-check root. Browser/visual, performance, deployment, full-repository and release qualification are outside the authorized scope. No required gate was skipped or left failed. Every final graph invocation has matching source/toolchain/system identity, no missing declared unit outputs, a passing retained-secret scan and zero failed/blocked cleanup steps.

E-058 documentation qualification: `CARTULARY_TEST_RUN_ID=gp06-final-markdown-01 make lint-markdown` completed with observed outer exit 0 and tool summary pass/exit 0, zero failures, at `.cartulary/test-results/gp06-final-markdown-01/adhoc/lint-markdown/tool-run-summary.json`. Final status/evidence accounting is recorded after this check; no executable file changed. Read-only `git diff --check` is clean. A rejected `make help TARGET=bootstrap` and two rejected diagnostic `explain-run` argument combinations were usage errors only; corrected discovery/diagnosis did not rerun product tests.

E-059 / S-015 completion and handoff: S-011–S-015, WF-19–WF-23 and T-024–T-028 are DONE. F-016–F-018 are resolved; current E-055/E-056 evidence requalifies the combined F-011–F-018 remediation and supersedes the earlier broad S-006/S-010 acceptance claims. All required specification, implementation, regression, compatibility, transaction, static, evidence-audit and documentation gates pass. No unresolved finding or required acceptance limitation remains within this scope. Corrected generic inputs may change identities; unaffected Graph fixtures and both NF production bindings remain exact, and incompatible stored bindings still reject. The reviewed combined diff is ready to commit; staging/commit remain unperformed. This is the final implementation workstream/status update.

### GP-05 implementation checkpoints

S-010 entry: E-036 recorded after all S-009 gates passed. Finalize now, review generated changes, freeze executable source/routing, then qualify the complete Graph owner and bounded peer portfolio. No full warm-check root exists for optional retained-run maintenance; do not launch one solely for finalization.

S-009 entry: E-035 recorded before entry; Graph root, NF provider, Reporting mapper and REQ-RPT-076/078/079/099g revalidated. Correct §15.2 before code movement. Keep public reasons, admission ordering, retained-input independence and result-integrity ownership.

S-008 entry: E-034 recorded before entry. Private helper callers, unused work fields and duplicate normalization inspected again; no callers use the permissive numeric mode. Current identity framing and legitimate optional defaults remain required.

S-007 entry: S-006 checkpoint E-032 recorded; registry constructors/accessors, binding admission and existing service/rollback assertions revalidated against unchanged Graph §9. No new owner contradiction.

| Evidence / checkpoint | Changes / command / exact run | Outcome / remaining gate |
| --- | --- | --- |
| E-030 / baseline and S-006 entry | Explicit implementation authorization; baseline in §1; Graph owner discovery exited 0 | S-006 IN_PROGRESS; successors await checkpoints |
| E-031 / S-006 initial reproduction | `CARTULARY_TEST_RUN_ID=gp05-s006-repro-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism`; exact root `.cartulary/test-results/gp05-s006-repro-01`, `run-summary.json` | Outer exit 2; product/test_assertion_failure. Identical/conflicting duplicate IDs, two-rule cycle and cycle after a ready batch all returned nil error. Valid chain/ready-batch order and permutation case passed. Cleanup: 10 steps, zero failed/blocked; missing outputs empty. Production unchanged at reproduction. `explain-run` confirmed failures; two earlier diagnostic usage errors (missing/invalid TARGET) did not rerun tests. |
| E-032 / S-006 completion | Duplicate-safe fallible ordering; closed non-retryable validation failure; valid ready-batch order unchanged. `CARTULARY_TEST_RUN_ID=gp05-s006-check-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.engine.v2_maximum_semantic_bounds,module.graphprojection.storage.v2_contract_projection` | Outer exit 0, summary pass; exact root `.cartulary/test-results/gp05-s006-check-01`, `run-summary.json`. V-001/V-017–V-019 pass; F-011 closed. Cleanup 11 steps, zero failed/blocked. No golden, routing or format-generation change. S-006 DONE before S-007 entry. |
| E-033 / S-007 first validation | Runs `gp05-s007-graph-01` (V-010/V-012/V-013) and `gp05-s007-recovery-01` (V-020), under `.cartulary/test-results/`; summaries `run-summary.json` | Recovery outer exit 0/pass; Graph outer exit 2/product test_assertion_failure in the new malformed-binding fixture. Registry mutation cases and both database rows passed. The fixture had discarded the selected digest, violating required result identity; corrected it to retain that digest while passing the rejected definition. No production repair arose from this fixture issue. Cleanup 18/10 steps respectively, zero failed/blocked. Requalify V-010; final-source portfolio remains required. |
| E-034 / S-007 completion | Private per-entry copies isolate constructor/document/registration slices and preserve callbacks; exact packaged current-binding admission replaces generation iteration. V-010 rerun: `CARTULARY_TEST_RUN_ID=gp05-s007-restore-02 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_deterministic_current_restore_rebuild` | Outer exit 0/pass, `.cartulary/test-results/gp05-s007-restore-02/run-summary.json`. Mutation/rebuild and strict definition/pre-clear cases pass. E-033 supplies V-012/V-013/V-020 development evidence; final-source runs remain required. F-012 closed; S-007 DONE before S-008 entry. |
| E-035 / S-008 completion | Removed strict-only mode plumbing, unused private digests/arguments/initializers and redundant parser sorts (including aggregation presort; ready-batch normalization remains sole ordering owner). `make format` runs `gp05-s008-format-01/02`; `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism,module.graphprojection.engine.v2_semantic_matrix,module.graphprojection.engine.v2_maximum_semantic_bounds,module.graphprojection.storage.v2_contract_projection` run `gp05-s008-graph-01`; `make lint CHECK=go` run `gp05-s008-go-01` | All outer exits 0/pass; exact roots `.cartulary/test-results/<run-id>`, `run-summary.json`. V-001/V-017–V-019/V-025 pass with unchanged goldens. Go cleanup 15 steps, zero failed/blocked. Formatting stayed within already changed code. F-013 closed; S-008 DONE before S-009 entry. |
| E-036 / S-009 completion | Reporting §15.2 corrected before code movement; source errors moved without aliases; direct/wrapped mappings and real unselected/stale provider cases added; positive Graph exported-value guard added to authored routing. `make generate` (`gp05-s009-generate-01`) refreshed only the topology index; `make format` (`gp05-s009-format-01`) passed. Validation runs `gp05-s009-graph-01` (V-001), `gp05-s009-nf-01` (V-014/V-023/V-024), `gp05-s009-reporting-01` (V-021/V-022), `gp05-s009-generated-01` (V-003), `gp05-s009-schema-01` (V-011), `gp05-s009-boundary-01` (V-005), `gp05-s009-markdown-01` (V-007), using exact §8 commands with those run IDs | All outer exits 0/pass. Exact roots `.cartulary/test-results/<run-id>`; graph summaries `run-summary.json`; generate summary `generate/tool-run-summary.json`; Markdown summary `adhoc/lint-markdown/tool-run-summary.json`. Graph/Reporting/generated/schema/boundary cleanup counts 10/15/14/12/13, zero failed/blocked. NF summary and cleanup inspected before checkpoint. Public reason codes and admission/retention/cleanup behavior pass. F-014 closed; S-009 DONE before S-010 entry. |
| E-037 / S-010 finalization and freeze | `CARTULARY_TEST_RUN_ID=gp05-s010-finalize-01 make agent-finalize`; exact root `.cartulary/test-results/gp05-s010-finalize-01`, `run-summary.json` | Outer exit 0/pass. Reviewed generated diff: only the intended selector hash/topology input digest differs from baseline. Executable source/routing frozen; non-document fingerprint ledger at `/tmp/cartulary-graph-gp05-6j7yob3s/frozen-source.json`. Optional retained-run actions omitted because no compatible successful full warm-check root was supplied; required finalization completed. Final portfolio/audit still pending. |

### GP-05 final-source validation evidence

Executable source/routing was frozen after E-037. Every final graph invocation records source digest `sha256:2e6fa246edc3eebfb7078fb203038e3e0e22e4cbe5190650a3741c3887f1b958`, toolchain digest `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f` and system digest `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1` in its manifest. Environment: repository root, Bash/Linux/WSL2, pinned Go 1.27.3, Node 24.15.0 and pnpm 10.33.0, repository-managed isolated PostgreSQL for service rows. Exact timestamps, selection and environment remain in each `run-manifest.json`; normal cache policy applies, so compatible cache reuse is not described as fresh execution.

Every row below uses `CARTULARY_TEST_RUN_ID=<run ID>` followed by the exact command. Result roots are `.cartulary/test-results/<run ID>`; graph summaries are `run-summary.json` and cleanup records are `cleanup-results.json`. All completed rows have observed outer exit 0 and reported pass. Product qualification uses final-source invocations; earlier development passes remain historical.

| Evidence / validation | Exact command after run-ID assignment | Run ID | Outcome / cleanup |
| --- | --- | --- | --- |
| E-038 / V-008 | `make test-slice OWNER=module.graphprojection` | `gp05-final-graph-01` | pass; complete 11-row/43-selector owner; cleanup 30 steps, zero failed/blocked |
| E-038 / V-014/V-023/V-024 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.integration.graph_result_cleanup_races` | `gp05-final-nf-01` | pass; cleanup 17 steps, zero failed/blocked |
| E-038 / V-021/V-022 | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.graph_v2_exact_selection_and_rendering_f54d8b1a62,module.reporting.integration.graph_input_admission_reads_the_exact_selected_c_4c42af0393` | `gp05-final-reporting-01` | pass; cleanup 14 steps, zero failed/blocked |
| E-038 / V-020 | `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c` | `gp05-final-recovery-01` | pass; cleanup 10 steps, zero failed/blocked |
| E-038 / V-015 | `make test-slice OWNER=app.server ROWS=app.server.unit.post_claim_graph_activation_preserves_structural_f41e41a052` | `gp05-final-server-01` | pass; cleanup 10 steps, zero failed/blocked |
| E-038 / V-003 | `make source-check CHECK=generated` | `gp05-final-generated-01` | pass; cleanup 14 steps, zero failed/blocked |
| E-038 / V-011 | `make source-check CHECK=schema` | `gp05-final-schema-01` | pass; cleanup 12 steps, zero failed/blocked |
| E-038 / V-025 | `make lint CHECK=go` | `gp05-final-go-01` | pass; cleanup 15 steps, zero failed/blocked |
| E-038 / V-005 | `make lint CHECK=backend-boundary` | `gp05-final-boundary-01` | pass; cleanup 13 steps, zero failed/blocked |
| E-039 / V-026 | `make test-evidence-audit OWNER=module.graphprojection EVIDENCE_ROOTS_FILE=/tmp/cartulary-graph-gp05-6j7yob3s/graph-owner-roots.json` | `gp05-final-audit-01` | pass; `test-evidence-audit/audit.json` accepts exactly `gp05-final-graph-01` for the complete required owner union; zero rejected/unused inputs; cleanup 10 steps, zero failed/blocked |

E-039 uses an explicit JSON manifest containing the observed final Graph root, independent of this document. The audit validates the required complete Graph-owner evidence union against current source/catalog/routing; peer acceptance is limited to the selected rows. It does not certify specification completeness or release readiness.

E-040 preservation review: all 8,061 baseline paths remain, with no added/removed source files. Content differs only in the 20 intended files (this tracker, Reporting prose, 16 authored Go files, the authored Graph selector manifest, and the Make-generated topology index). Some generated outputs were rewritten with identical bytes by managed Make; their identity/mtime changes are not reported as content changes. The raw index changed during managed workflows, but read-only comparison of original/current index entries proves identical staged paths/modes/blob IDs; no stage/commit operation occurred. The GP-03 and GP-04 session evidence blocks remain byte-for-byte unchanged. All 7,640 frozen executable-source entries remain unchanged. The manual preservation record is `/tmp/cartulary-graph-gp05-6j7yob3s/final-preservation.json`; it is not an input to product tests or the evidence audit. Git whitespace inspection reports no diagnostics.

E-041 document qualification: `CARTULARY_TEST_RUN_ID=gp05-final-markdown-01 make lint-markdown` retains a pass with tool exit 0 in `.cartulary/test-results/gp05-final-markdown-01/adhoc/lint-markdown/tool-run-summary.json`; its outer exit was not retained across context compaction. After final evidence reconciliation, `CARTULARY_TEST_RUN_ID=gp05-final-markdown-02 make lint-markdown` passed with observed outer exit 0, zero failures, and summary `.cartulary/test-results/gp05-final-markdown-02/adhoc/lint-markdown/tool-run-summary.json`. V-007 is complete.

E-042 / S-010 completion and handoff: E-037–E-041 satisfy the required finalization, frozen-source portfolio, complete Graph-owner audit, preservation review and document validation. F-011–F-015, S-006–S-010, WF-14–WF-18 and T-019–T-023 are DONE. Inventory remains 36 files in five packages; the Graph owner has 11 rows and 43 top-level selectors. Valid canonical identities, current restore generations, transaction atomicity and retained Reporting input remain supported. Previously accepted malformed aggregations now reject, and two internal source-selection sentinels moved to the consumer contract without aliases. No migration or unresolved required gate remains. This completion/status record is the final implementation-slice update; earlier sessions and failed development reproductions remain historical evidence.

### GP-04 planning and document update

GP-04 is based on clean commit `2dc1519246657929dc7194de9db22987597f9a92`. Read-only planning preserved repository contents and index; the subsequent accepted request authorizes this tracker update and documentation lint only. Source observations below are inspection evidence, not executed reproductions or product passes. No S-006–S-010 implementation is performed or authorized by this session.

| Evidence / category | Date and inspected sources or command | Result / artifact / scope | Next action |
| --- | --- | --- | --- |
| E-022 / scope and preservation baseline | 2026-11-03; root and applicable backend/tooling instructions, refactor-tracker/test-ops guidance, Core precedence, current Graph owner and tracker; Git/path identity and 8,061-path snapshot | Clean `main` at the stated commit; original tracker, index identity and per-path snapshot retained at `/tmp/cartulary-graph-gp04-jug0qsnf`. Output admitted as a regular single-link tracked file. No unrelated write authorized | Preserve original history and baseline; final comparison E-029 |
| E-023 / backend correctness and contract | `internal/modules/graphprojection/engine_v2.go` (`parseProjectionInputV2`, `normalizeProjectionRequestV2`, `orderAggregationRulesV2`), `validation.go`, `aggregation.go`; applicable engine tests and Graph §§3–6 | Directly inspected ID-keyed insertion before later duplicate validation and dependency-stall fallback. F-011/C-015; no failing runtime case has been executed | Future S-006 records reproductions before correcting admission |
| E-024 / definition integrity and admission | `internal/modules/graphprojection/restore/contract.go` constructors/accessors; `service.go` binding validation and pre-clear admission; `contract_test.go`, `service_test.go`, `boundary_test.go`; current generated binding definitions and Graph §9 | Nested `SemanticQuerySchemaIDs` slices are shallow-copied; current binding validation iterates generation definitions. F-012/C-016. Existing tests/rollback evidence do not prove alias isolation | Future S-007 owns slices and validates exactly the current typed/canonical definition |
| E-025 / private cleanup and identities | Graph `canonical.go`, `types.go`, `project.go`, `input.go`, `engine_v2.go`, `validation.go`, relevant tests, schemas and repository caller searches | Strict-only canonical-number mode, assigned/unread work digests, ignored arguments, unreachable fallback, overwritten predicate initializers and repeated sorts identified. F-013; retain actual defaults, source order, cancellation and current identity-domain strings | Future S-008 makes a separate mechanical cleanup after correctness/integrity fixes |
| E-026 / consumer boundary, security and documentation | Graph `result_v2_ports.go`/`boundary_guard_test.go`; NF `reporting_graph_source.go`/`reporting_graph_source_integration_test.go`; Reporting `graph_source.go`, `graphsourcecontract/contract.go`, `graph_projection_refs_test.go`; Reporting §§8/15.2 and REQ-RPT-099g | F-014/C-017: two declaration-selection sentinels have NF producer/Reporting mapper users; destination package exists. Keep auth/admission precedence, locking, public reason codes and retained input. No frontend change or new authorization defect asserted; no browser review performed | Future S-009 relocates symbols without aliases and fixes superseded wording with consumer regressions |
| E-027 / test-routing and command discovery | `make help ALL=1`; `make task-guide OWNER=module.graphprojection DETAIL=rows`; `make task-guide OWNER=module.networkflow`; `make task-guide OWNER=module.reporting` | All discovery commands exited 0 during planning; outputs retained in session. Inspected Graph/NF/Reporting/Recovery/server authored test-family manifests and audit manifest schema. Exact row commands are in §8; baseline Graph has 11 rows/42 selectors. No product invocation, generation or evidence audit performed | Future S-010 requires current compatible owner evidence and an explicit non-Markdown root manifest |
| E-028 / document update and lint | `CARTULARY_TEST_RUN_ID=gp04-plan-lint-20261103-01 make lint-markdown` from the repository root | Outer exit 0 and retained status `pass`; exact root `.cartulary/test-results/gp04-plan-lint-20261103-01`, [tool-run-summary.json](../../.cartulary/test-results/gp04-plan-lint-20261103-01/adhoc/lint-markdown/tool-run-summary.json). UTC 2026-11-03 21:51:43–21:52:04; zero failures. GP-04 plan recorded; no product pass claimed | Record completed handoff; future implementation remains TODO |
| E-029 / final scope review | Read-only tracker/whitespace review, twelve-section/stable-ID comparison, exact authored routing inspection, all baseline paths and Git index comparison | All 8,061 paths retained, none added/removed; only tracker content/mtime changed. Its inode/mode/link count and all other baseline path content/identity match. HEAD and raw index hash/identity unchanged. `git --no-optional-locks diff --check` exited 0 | T-018 DONE; preserve findings and follow S-006 entry gates in a later implementation task |

E-028 completion checkpoint: after recording completed handoff statuses, `CARTULARY_TEST_RUN_ID=gp04-plan-lint-20261103-02 make lint-markdown` also exited 0 with retained status `pass`; exact root `.cartulary/test-results/gp04-plan-lint-20261103-02`, [tool-run-summary.json](../../.cartulary/test-results/gp04-plan-lint-20261103-02/adhoc/lint-markdown/tool-run-summary.json). E-029's document readback also confirmed 34 consistent tables, existing relative-link targets, all historical IDs, and unchanged GP-01–GP-03 session records.

Environment and coverage: repository root `/home/jochi/code/cartulary`, Linux/WSL2, Bash, normal managed cache paths under `/home/jochi/.cache/cartulary`; the E-028 summaries retain command/environment and timestamps. No database/service fixture, product test, performance check, generator, formatter, finalization or deployment was run for this documentation update; these checks are outside its authorized scope, not silently skipped implementation gates. Lint wrote ordinary ignored managed artifacts. No service cleanup was applicable, and no claim is made that ignored caches stayed unchanged. GP-03 toolchain/run metadata is not reused as GP-04 runtime qualification.

### GP-03 execution checkpoints

| Evidence / checkpoint | Date / baseline | Changes and commands | Result / artifact | Next action |
| --- | --- | --- | --- | --- |
| E-014 / bootstrap | 2026-11-03; main `f8ae96be698fb7f78c3f403e62cf5099752e6dfe` | Revalidated instructions, current lease drift and tracker identity; captured 8,060-path baseline outside repository | User authorized accepted S-001–S-005 plan; tracker updated before implementation; original tracker preserved in `/tmp/cartulary-graph-remediation-j9cs9xwz/tracker-before.md` | S-001 IN_PROGRESS |
| E-015 / S-001 DONE | 2026-11-03; GP-03 baseline plus editorial changes | Graph glossary, domain relationship/change-trigger/coverage rows, framework boundary, and Core 04 tracker correction; `make task-guide OWNER=module.graphprojection` exit 0; `CARTULARY_TEST_RUN_ID=gp-remediation-s001-lint-20261103-01 make lint-markdown` exit 0 | Lint passed: `.cartulary/test-results/gp-remediation-s001-lint-20261103-01/adhoc/lint-markdown/tool-run-summary.json`; reviewed against Core 00 frozen-input amendment and Graph §8; no semantic/generation change | S-002 IN_PROGRESS; comment revalidation next |
| E-016 / S-002 DONE | 2026-11-03; GP-03 | Re-read writer SQL/commit path, Graph §9, Recovery contribution and caller rollback; corrected Clear comment | `git diff -- internal/modules/graphprojection/postgresrestore/writer.go` contains only comment changes; no test added for prose | S-003 IN_PROGRESS |
| E-017 / S-003 DONE | 2026-11-03; GP-03 service/selector changes | Recording clear/enumerate/publish fake, metadata/source/semantic/binding/overflow/cancellation assertions, closed failed results, two renamed tests and one added test; generated topology index updated only by Make | `make format`, `make generate`, V-010, V-003 and V-011 all exit 0. Run roots `.cartulary/test-results/gp-remediation-s003-{format,generate,unit,generated,schema}-20261103-01`; graph commands retain `run-summary.json`, generate retains `generate/tool-run-summary.json`. Graph catalog now 11 rows/42 selectors, restore row 14 selectors; not a full-owner pass | S-004 IN_PROGRESS; RB-002 remains open until real rollback evidence |
| E-018 / S-004 DONE | 2026-11-03; final test/routing changes; source digest `sha256:6b8889aaa18a1ac32bf018ffea7c0181bf55b71441b00835c61562c609423d48` | Actual nonempty three-table rollback after enumeration error and cancellation; real NF materialization followed by worker shutdown and selected-binding mismatch; exact Graph/declaration/job/input snapshots; existing writer and lifecycle tests preserved | `make format`, `make generate`, V-012/V-013/V-014 all exit 0. Roots `.cartulary/test-results/gp-remediation-s004-{format,generate,assembly,writer,networkflow}-20261103-01`; test roots retain `run-summary.json`, row and cleanup results. Each service invocation has 14 cleanup steps, zero failed/blocked; no missing outputs. No executable production defect found | F-005/RB-002 closed; S-005 IN_PROGRESS before finalization |
| E-019 / S-005 source freeze | 2026-11-03; executable source digest `sha256:6b8889aaa18a1ac32bf018ffea7c0181bf55b71441b00835c61562c609423d48` | `CARTULARY_TEST_RUN_ID=gp-remediation-s005-finalize-20261103-01 make agent-finalize` | Exit 0; `.cartulary/test-results/gp-remediation-s005-finalize-20261103-01/run-summary.json`; reviewed generated topology index (only two selector hashes and aggregate digest changed); no additional source change. No RESULTS_DIR supplied: no full warm-check root exists; retained full-check maintenance not requested | Freeze executable sources/routing; execute final bounded portfolio and reconcile handoff |
| E-020 / final-source portfolio | 2026-11-03; E-019 frozen executable source | Eight executable-source invocations and canonical Markdown lint below; native process exits observed separately from retained summaries | All exit 0/pass; 26 selected top-level tests across five rows; generated/schema/boundary/Markdown checks pass; three managed-service invocations each completed all 14 cleanup steps with no failed/blocked steps or missing outputs | Required bounded validation complete; no unresolved check failure |
| E-021 / preservation and handoff review | 2026-11-03; compare against E-014 baseline | Read-only path/content/identity comparison, Git index/worktree diff review, tracker findings/contracts/selectors/history reconciliation | 8,060 baseline paths retained; ten existing files intentionally changed, one new NF test, no removed paths or unrelated content changes. Branch/commit unchanged; index entries remain identical to HEAD; raw index hash changed to `8725a519abab33c7feb374b662a63c29519c6fb96b90284bd7a234a5a90fbd28` from cached-stat refresh. Make rewrote some generated outputs with identical contents; no identity-preservation claim for those files. Existing untracked tracker and its historical session records retained | S-005/WF-13 DONE; handoff complete; no required next implementation slice |

### GP-03 final verification evidence

All commands below ran from `/home/jochi/code/cartulary` with `CARTULARY_TEST_RUN_ID=<run ID>`. Each artifact link gives the exact retained run root and summary. V-003/V-005/V-010–V-015 resolve the full command and selectors in §8; these are fresh executions after V-016, not inferred reuse of development results. Native outer exit codes and retained summary status both indicate success.

| Validation / exact Make command reference | Run ID | Retained summary | Outer exit / result |
| --- | --- | --- | --- |
| V-010 / restore unit row | `gp-remediation-s005-unit-20261103-01` | [run-summary.json](../../.cartulary/test-results/gp-remediation-s005-unit-20261103-01/run-summary.json) | 0 / pass |
| V-012 / Graph assembly row | `gp-remediation-s005-assembly-20261103-01` | [run-summary.json](../../.cartulary/test-results/gp-remediation-s005-assembly-20261103-01/run-summary.json) | 0 / pass |
| V-013 / writer failure row | `gp-remediation-s005-writer-20261103-01` | [run-summary.json](../../.cartulary/test-results/gp-remediation-s005-writer-20261103-01/run-summary.json) | 0 / pass |
| V-014 / NF saved-graph lifecycle row | `gp-remediation-s005-networkflow-20261103-01` | [run-summary.json](../../.cartulary/test-results/gp-remediation-s005-networkflow-20261103-01/run-summary.json) | 0 / pass |
| V-015 / server activation row | `gp-remediation-s005-activation-20261103-01` | [run-summary.json](../../.cartulary/test-results/gp-remediation-s005-activation-20261103-01/run-summary.json) | 0 / pass |
| V-011 / `make source-check CHECK=schema` | `gp-remediation-s005-schema-20261103-01` | [run-summary.json](../../.cartulary/test-results/gp-remediation-s005-schema-20261103-01/run-summary.json) | 0 / pass |
| V-003 / `make source-check CHECK=generated` | `gp-remediation-s005-generated-20261103-01` | [run-summary.json](../../.cartulary/test-results/gp-remediation-s005-generated-20261103-01/run-summary.json) | 0 / pass |
| V-005 / `make lint CHECK=backend-boundary` | `gp-remediation-s005-boundary-20261103-01` | [run-summary.json](../../.cartulary/test-results/gp-remediation-s005-boundary-20261103-01/run-summary.json) | 0 / pass |
| V-007 / `make lint-markdown` | `gp-remediation-s005-lint-20261103-01` | [tool-run-summary.json](../../.cartulary/test-results/gp-remediation-s005-lint-20261103-01/adhoc/lint-markdown/tool-run-summary.json) | 0 / pass |
| V-007 / `make lint-markdown` after handoff status reconciliation | `gp-remediation-s005-lint-20261103-02` | [tool-run-summary.json](../../.cartulary/test-results/gp-remediation-s005-lint-20261103-02/adhoc/lint-markdown/tool-run-summary.json) | 0 / pass |

Environment: Linux x64 on WSL2 kernel `6.6.114.1-microsoft-standard-WSL2`, Intel Core Ultra 9 285HX; managed owned service fixtures, normal cache mode, Go `go1.27.3`, `GOOS=linux`, `GOARCH=amd64`, `CGO_ENABLED=0`, no race/shuffle override. Every final executable-source run and finalization records commit `f8ae96be698fb7f78c3f403e62cf5099752e6dfe`, dirty worktree source digest `sha256:6b8889aaa18a1ac32bf018ffea7c0181bf55b71441b00835c61562c609423d48`, and toolchain digest `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`. Source manifests exclude Markdown; subsequent tracker edits do not affect executable-source compatibility. Markdown lint separately validates documentation.

For service runs, `make explain-run RESULTS_DIR=<exact run root above> DETAIL=accounting` exited 0 and confirmed completed cleanup, zero failed/blocked units and no missing outputs. Retained `run-manifest.json`, `execution-context.json`, row summaries and `cleanup-results.json` provide the environment, exact source inputs, selection and cleanup evidence. No required product check was skipped. Full-owner, browser, visual, performance, deployment and release qualification remain outside the agreed scope; no broad warm-check root was created for optional retained-run maintenance.

The remaining session tables preserve GP-01/GP-02 historical evidence; they are not current execution outcomes.

Session IDs are stable within this document. GP-01 is the earlier read-only planning session represented by the accepted plan; GP-02 is the tracker creation session on 2026-11-03. There was no prior tracker file/history to overwrite. Dates below identify sessions, not test-run timestamps. Files inspected are distinct from the sole file touched: `docs/handoffs/graphprojection-module-refactor-tracker.md` in GP-02.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-03, GP-01 | Codex / planning | Initial planning completed on main at the recorded commit; writing withheld in the earlier Plan Mode | Inspected framework first, owner/status documents, instruction/skill chains and source ledger below; touched none | Git inspection, `rg` enumeration/search, source reads; both Make discovery commands in E-011 | Read-only plan recommended retention; clean final status reported in the accepted plan | Earlier Plan Mode prevented writing; resolved for GP-02 | Recheck output and baseline, then create only the tracker. |
| 2026-11-03, GP-02 | Codex / tracker creation | Accepted documentation plan recorded; production scope unchanged | Re-read framework, instructions and relevant authority/source definitions; tracker is the only touched file | E-011/E-012; document readback E-013 | Authority and inspection depth are explicit; no owner contradiction found in the scoped sources after applying adoption precedence | RB-001 affects canonical lint only; Plan Mode is no longer a blocker | Future work needs its own authorized scope; recheck baseline then. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-03, GP-01/GP-02 | Codex / boundary review | Five Graph packages form a pure root with narrow Graph-owned adapters and assembly composition | Inspected all 36 target files at inventory depth, E-006 consumers and E-008 assembly; touched tracker only in GP-02 | Target enumeration, import/caller searches, exact-file reads | Retain boundary; no structural slices; F-002/F-003/F-006/F-007/F-010 | No planning blocker; F-004 comment and F-005 characterization deferred | Preserve root purity, borrowed capabilities, single transaction and source-owner policy; revalidate only for a concrete future change. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-03, GP-01/GP-02 | Codex / applicability review | No frontend implementation is in this target; target imports/ports and inspected HTTP consumers do not establish direct selector, controller or grid-vendor ownership | Inspected target boundary and NF route surfaces, not frontend implementation; touched tracker only in GP-02 | Backend import/caller and route searches/reads | Frontend workstream is inapplicable to the retained-boundary plan; C-013/C-014 | None for current plan; no claim of a full frontend audit | If future scope reaches rendered behavior, load its instructions and discover its browser/selector coverage before planning a move. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-03, GP-01/GP-02 | Codex / contract review | Current owner posture is Graph 2.2.1, NF 8.0.2, Reporting 1.4.1; restore/registry v5 and binding v7 | E-001/E-009 plus schema tests/ports; tracker only touched in GP-02 | Targeted owner/status searches, exact source/contract/generated-definition reads | v2 identities, registry/binding, three-table proof and frozen-input contracts mapped; generated projections left untouched | F-001 supporting prose drift is resolved for this plan by owner precedence, not by editing or inventing compatibility | Future authored-input changes regenerate through Make and check V-003; never use Markdown as runtime/test input. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-03, GP-01/GP-02 | Codex / evidence accounting | Eleven Graph rows select 41 tests; helpers require parent tests; catalog and assertions are evidence posture only | Target tests, selected consumer/assembly tests, pgtest/graphfixture use, E-010 Make/catalog definitions; touched tracker only | `make help ALL=1`; `make task-guide OWNER=module.graphprojection`; read-only JSON/assertion inspection; no product test | Both discovery commands exited zero; V-001–V-009 recorded; no product pass claim | RB-001 lint writes; RB-002 conditional Clear/rollback evidence gap | Preserve mappings; do not run product commands for this artifact. A future implementation task selects affected rows and retained-run handling through Make/test-ops. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-03, GP-01/GP-02 | Codex / owner review | NF admits incident-scoped operations; trusted invocation does not turn the pure engine into an HTTP authorization layer; source/Reporting admission remains owner-local | E-001d/e/g, NF routes/admission, saved-read failures, Reporting admission, Graph invocation/read/cleanup ports; tracker only touched | Role/admission/caller searches and selected exact function/assertion reads | C-008–C-010 capture roles, preconditions and retained-input admission; no security behavior change proposed | No newly asserted authorization defect; scoped review does not certify all routes | Any future cross-owner change must preserve auth-before-resource error precedence, result-before-declaration locking and retained execution independence. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-03, GP-02 | Codex / final handoff | Planning complete; retain boundary; production implementation is not authorized by this document | Tracker only touched; full inspected-source ledger below | E-012/E-013 final readback and comparison | Twelve sections, inventory, table columns, links and dependencies reconciled; non-output baseline unchanged; product checks deliberately not run | RB-001; RB-002 only for future sequencing changes | Next safe work is a separately scoped maintenance/characterization task if requested, not an assumed refactor. |

### Inspected repository evidence ledger

§2 enumerates every target file individually. The ledger below records adjacent files and inspection depth so later agents can revisit exact sources without treating search hits as verified assertions. Paths are repository-relative. Read-only inspection does not make these files part of the permitted write set.

| Evidence | Files inspected | Depth and use |
| --- | --- | --- |
| E-003 | All root files and `internal/modules/graphprojection/internal/semanticlimits/registry.go` listed in §2 | Imports/declarations and selected implementation/assertions across the pure engine; deeper reads of entrypoint/result ports/validation and positive boundary guards. Not exhaustive semantic certification. |
| E-004 | Every `postgresresult/` file plus `graph_projection_migration_test.go` in §2 | Result publication/read/traversal/cleanup entrypoints and relevant SQL/locking/proof assertions; helper-parent relation. Migration test assertions inspected; migrations were not modified or executed. |
| E-005 | Every `restore/` and `postgresrestore/` file in §2 | Contract types/validation, source registry/rebuild/recovery-state logic, writer phases and selected failure assertions. Service recording publisher and Clear limitation directly checked. |
| E-006 | `internal/modules/networkflow/graph_projection_adapter.go`; `graph_result_cleanup.go`; `graph_routes.go`; `graph_view_routes.go`; `saved_graph_read_application.go`; `operation_admission.go`; `module.go`; `graph_restore_source.go`; `reporting_graph_source.go`; `saved_graph_read_application_test.go`; `reporting_graph_source_integration_test.go` (same directory for abbreviated names) | Adapter, registry, route/admission/read/cleanup and retained-source functions; selected saved-read/retention assertions. Route registries and authorization paths were inspected, not every route implementation or every test body. |
| E-007 | `internal/modules/reporting/graph_source.go`; `graph_retention.go`; `graph_retention_test.go`; `graph_projection_refs_test.go` | Registry and retained-input consumption, retention helpers/assertions; helper-parent relationship and exact-read/owner/error-mapping assertions directly read in graph projection reference tests. No claim of exhaustive Reporting lifecycle review. |
| E-008 | `internal/app/graphassembly/rebuild.go`; `rebuild_test.go`; `internal/app/server/graph_activation.go`; `graph_activation_test.go` | Transaction assembly and one-transaction integration assertions; post-claim readiness/activation implementation and selected unit cases. Tests were read, not run. |
| E-009 | `tools/generated_artifact_policy.json`; `tools/execution_topology_manifest.json`; `contracts/graph-projection/index.json`; `contracts/graph-projection/storage-maintenance.v2.json`; `internal/gen/contractgraphprojection/artifacts_gen.go`; `internal/gen/contractrecovery/graph_projection_restore_binding_gen.go`; `internal/gen/networkflowroutes/catalog_gen.go` | Policy/output declarations, contract index/storage contract, generated markers and selected registry/catalog/binding entries. Additional schemas/fixtures were traced through inspected tests/index metadata; this is not a full schema-content or generated-drift audit. |
| E-010 | `tools/test_families/module.graphprojection.json`; `contracts/verification/owners/module.graphprojection.json`; peer owner descriptors `module.networkflow.json`, `module.reporting.json`, `app.server.json` in that directory; `Makefile`; `tools/task_surface_manifest.json`; `tools/task_surface.help-all.generated.json`; `tools/task_surface.generated.mk`; `tools/task_surface.runtime.generated.mk`; `tools/harness/workspace/admission.mk`; `tools/harness/workspace/admission.sh`; `tools/harness/execution/run-make-node-tool.sh`; `tools/harness/execution/run-make-node-tool-cli.mjs`; `tools/harness/command-surface/make-node-tools.mjs`; `tools/harness/static-analysis/markdownlint.sh`; `tools/harness/static-analysis/markdownlint-runner.sh`; `.markdownlint-cli2.jsonc`; `tools/toolchain_pins.json`; `internal/testutil/graphfixture/result.go` | Graph catalog rows/selectors and owner metadata; selected command definitions/wrappers/side effects; first toolchain pin entries; real-producer fixture helper. Paths reconciled with live enumeration; helpers are not independent catalog tests. |
| E-011 | Session command outputs retained in chat, not written to a new repository artifact | Both Make discovery commands returned exit zero. File/Git searches and reads establish current state only. See command log below. |
| E-012 | In-memory prewrite baseline and final comparison; §1 stores aggregate identities | Git HEAD/index/status, target content/identity, all tracked/non-ignored non-output paths, relevant static-analysis cache and test-results root. No broad claim about contents of all unrelated ignored outputs. |
| E-013 | This tracker, after writing | Document readback and read-only structural inspection found twelve sections, 23 consistent tables, 36 inventory rows, seven handoff tables, valid relative links and an acyclic seven-workflow graph. Catalog comparison confirmed 11 rows/41 selectors. Initial extra EOF blank line was corrected; subsequent Git whitespace inspection emitted no diagnostics. Not an executable product check and not Markdown lint. |

Search-only candidates: `internal/modules/networkflow/routes_integration_test.go` yielded `TestNetworkFlowSavedGraphLifecycleRoutes_Integration`, `TestNetworkFlowTimeBucketSavedGraphLifecycle_Integration` and `TestNetworkFlowGraphContributorsAndIndicatorLinkRoutes`; `graph_result_cleanup_race_integration_test.go` yielded `TestNetworkFlowGraphCleanupPublicationAndAdmissionRaces_Integration`. These declarations establish candidate follow-up coverage, not reviewed assertions or passing results. `graph_view_jobs.go` and `graph_view_store.go` were located as adjacent implementation candidates; no exhaustive job/store review is claimed.

### Command and write accounting

The following commands describe GP-02. E-081 records GP-09 document verification.

| Command family | Execution and result | Scope / side effects |
| --- | --- | --- |
| `make help ALL=1` | Executed successfully in discovery; exit 0 | Public command discovery. No product verification claim. |
| `make task-guide OWNER=module.graphprojection` | Executed successfully in discovery; exit 0 | Owner/row selection and command guidance. No tests executed. |
| `git --no-optional-locks` status / branch / rev-parse / ls-files / diff inspections | Executed for baseline, path universe and final scope | Ordinary read-only Git; no stage, commit, branch switch, reset or cleanup. |
| `git --no-optional-locks diff --no-index --check /dev/null docs/handoffs/graphprojection-module-refactor-tracker.md` | Initial extra EOF blank line corrected; repeat returned exit 1 with no whitespace diagnostics, reflecting the new-file diff | Inspection of the untracked artifact; no retained test artifact and no canonical lint claim. The companion `--stat` inspection showed one new file. |
| `rg --files`, targeted `rg` queries, `sed`/`cat` exact-file reads, filesystem identity/enumeration | Executed across the bounded sources listed above | Read-only inspection; no search snippet is an edit source. |
| Ephemeral read-only Python/JSON inspection | Executed for file counts, catalog accounting, hashes/path identity and tracker structure/link review | No script file or artifact written. An overly broad ignored-results fingerprint attempt was stopped and replaced with the bounded baseline stated in §1; this was not a product verification failure. |
| Patch application | GP-02 only; tracker creation and reconciliation | Sole admitted write: this tracker. No generated, production, test, harness or dependency changes. |
| `make lint-markdown` | Skipped | RB-001; inspected side effects exceed authorization. No direct-script bypass. |
| All product tests, code generation, `make agent-finalize`, broad checks | Not run | Outside this documentation-only task. |

## 11. Open Questions and Blockers

**RB-003 DONE, E-083.** The S-022 baseline exposed both malformed-result admissions and duplicate/overflow precedence. E-086 and E-088 close the behavioral findings. E-090/E-091 qualify the combined implementation.

### Historical blockers through GP-08

GP-08 closed F-022–F-028 on the combined candidate recorded in §10. No required gate is open for this scope. The following GP-07 decision is historical. Its reviewed rows passed (E-070–E-075). The following GP-06 gate is historical and superseded: **closed, E-059**. F-016–F-018 and all required final acceptance gates pass on the compatible combined source. GP-05 closure prose below is historical and is superseded by the more precise GP-06 compatibility boundary and qualification. No unresolved owner contradiction, implementation finding or required validation blocker remains within the selected scope.


No unresolved contradiction between applicable primary owners was identified. GP-03 found no runtime defect in its restore-failure remediation; GP-04 additionally records F-011/F-012 from source inspection and F-013/F-014 cleanup needs. GP-05 reproduced and corrected F-011 (E-031/E-032), hardened F-012 (E-034), simplified F-013 (E-035), and closed F-014 (E-036). S-010 final-source qualification and handoff close F-015 (E-037–E-042); historical evidence is not substituted for that gate.

| ID | Original gap / risk | Status | Resolution / remaining gate |
| --- | --- | --- | --- |
| RB-001 | GP-02 tracker-only scope prohibited lint artifacts | DONE | User authorized normal managed implementation validation; S-001 V-007 passed E-015 and final lint passed E-020. Original skip remains historical. |
| RB-002 | No observed Clear order or assembled rollback after source failure | DONE | S-003/E-017 and S-004/E-018 establish phase order, closed failures and real committed-state rollback. No production sequencing change required. |

F-001 and F-004 remain closed by E-015/E-016; F-005 remains closed by E-017/E-018. S-005 stays DONE with its historical finalization/checks/handoff E-019–E-021. Neither RB-001 nor RB-002 is reopened by a new iteration.

GP-05 satisfied authorization, owner/source revalidation, initial reproduction and sequential checkpoint gates. E-033 records a test-fixture identity failure and its diagnosed correction; E-034 passes the corrected case. No unresolved owner contradiction or runtime blocker remains in the implemented seams. S-010 product/static gates and owner audit pass (E-038/E-039); final documentation and handoff are complete (E-040–E-042).

No required gate remains open for F-011–F-015. §7 retains each remediation's rationale, compatibility and rollback decisions. No deployment, full semantic audit, performance or release certification follows from this bounded effort.

## 12. Binary Completion Criteria

### GP-09 completion criteria

| Gate | Status | Binary evidence |
| --- | --- | --- |
| Tracker update (T-035) | DONE | Twelve numbered sections and historical IDs preserved. E-081 records the initial lint and scope comparison. Final-text verification uses `CARTULARY_TEST_RUN_ID=gp09-tracker-final-01 make lint-markdown`. [Final lint receipt](../../.cartulary/test-results/gp09-tracker-final-01/adhoc/lint-markdown/tool-run-summary.json) |
| S-022 characterization (T-036) | DONE | E-083. Valid byte fixtures and observed malformed direct vertex/edge behavior recorded on the baseline, closing RB-003. |
| S-023 input encoding (T-037) | DONE | E-084. S-022 bindings and bytes, optional members and numerical spelling. |
| S-024 collection counts (T-038) | DONE | E-085. Nine ceiling/first-excess pairs and competing errors. |
| S-027 input admission (T-042) | DONE | E-145 closes GP-15. E-137–E-139 cover input depth, diagnostic paths, literal keys and Network Flow admission. E-133–E-134 cover GP-13. E-128–E-129 cover resource precedence, declared source kinds and disjoint mapping keys. E-123–E-125 cover scalar traversal and metadata mapping IDs. E-118–E-119 cover the terminal issue ceiling and definition IDs. Prior qualification at E-111. Previous repair: E-097–E-098 close the reopened admission gates. Original candidate: E-086. Both source collections, cross-collection overflow, distinct overflow, malformed shape, cancellation and the issue budget pass V-001/V-018 before S-025. |
| S-025 result content (T-039) | DONE | E-087. All result families and historical rows retain exact bytes and proof. |
| S-026 admission correction (T-040) | DONE | E-145 closes GP-15. E-137–E-139 cover raw provenance, scope defaults and the canonical output budget. E-133–E-134 cover GP-13. E-128–E-129 cover label ceilings, edge identity and label agreement, reachable nulls and configured producers. E-123–E-125 cover mandatory fields, relationship claims, label capabilities, object labels and restore. E-118–E-119 cover emitted fields, direct provenance and per-kind directions. Earlier receipts E-105–E-110 and E-112 cover inert declarations. Previous repair: E-094–E-096 close the reopened admission gates. Original candidate: E-088. Rebound malformed content fails in the pure model, before any publication write and on stored reads. The schema projection agrees. Valid historical rows and metadata-only locking pass. |
| Frozen-source qualification and handoff (T-041) | DONE | E-158–E-159 close the rollback interruption follow-up after 43 passed rows, the eleven-row Graph audit, all static gates, exact frozen-source comparison and final-text Markdown. E-154–E-155 qualify the preceding GP-15 follow-up after 43 passed rows, the eleven-row Graph audit, static gates and final-text Markdown. E-148–E-150 qualify the preceding GP-15 source. E-140–E-141 qualify GP-14. E-135–E-136 qualify GP-13. E-130–E-131 qualify GP-12. E-125–E-126 qualify GP-11. E-120 qualifies GP-10. E-121 closes the final-text gate. E-113–E-116 qualify the preceding candidate. Previous candidate at E-099–E-103. Frozen source, all eleven audited Graph rows, peer and static gates, diff preservation and final-text lint. E-089–E-092 retain the original qualification. |

GP-02 completed its bounded planning/documentation task, including E-012 preservation and E-013 document review; it ran no product checks and skipped Markdown lint under its then-authorized write limit. That historical outcome is not GP-03 implementation evidence.

| GP-03 historical criterion | Status | Evidence / applicability |
| --- | --- | --- |
| F-001 specification/vocabulary cleanup | DONE | S-001/E-015; adopted frozen input and transaction-scoped locks agree; history and unrelated leases preserved |
| F-004 restore comment | DONE | S-002/E-016; executable diff empty |
| F-005 service order and closed failures | DONE | S-003/E-017; observable phases, metadata/clear/source/semantic/binding/resource/cancellation failures; 14 selected restore tests |
| F-005/RB-002 real transaction rollback | DONE | S-004/E-018; nonempty real Graph seed, actual clear, source error/cancellation and real NF binding mismatch; independent exact snapshots before cleanup; no commit and successful rollback |
| Required finalization | DONE | V-016/E-019; executable source frozen; no product contract/generation changes |
| Final-source bounded portfolio and lint | DONE | V-003/V-005/V-007/V-010–V-015; E-020 |
| Tracker traceability, baseline preservation and handoff | DONE | S-005; E-021 |

GP-03 planning and implementation were complete for that iteration's bounded remediation. S-001–S-005 and WF-09–WF-13 remain DONE. Their evidence does not certify GP-04's future source changes, all Graph semantics, browser behavior, deployment or release readiness.

| GP-04 historical document/planning criterion (at that session's close) | Status | Evidence / applicability |
| --- | --- | --- |
| Accepted scope, clean baseline and preserved history recorded | DONE | E-022; twelve original sections and existing IDs retained; new GP-04 history is distinct |
| New findings mapped to owners, contracts, slices, risks and validation | DONE | F-011–F-015, C-015–C-017, §§5–8; runtime corrections separated from mechanical cleanup |
| Sequential implementation plan and checkpoint rule | DONE | WF-14 → WF-15 → WF-16 → WF-17 → WF-18; S-006–S-010 all TODO; update results before entering each successor |
| Document lint and final preservation review | DONE | E-028/E-029; no product verification is required or claimed for this update |
| Future implementation readiness distinguished from authorization | DONE | S-006 is ready for characterization after authorization/revalidation; S-007–S-010 await predecessors. This document update does not authorize implementation |
| Future runtime completion and production-hardening evidence | TODO | T-019–T-023; no new fix, reproduction, product check or owner evidence audit executed |

GP-04 completed planning and documentation only (E-028/E-029); its lack of implementation authorization and unfinished runtime criteria above are historical. GP-05 explicitly supersedes those restrictions and statuses. The accepted design retains module topology, valid Graph v2 identities, Recovery v5/registry v5/binding v7, read/traversal/cancellation capabilities, SQL history and retirement rejection. Malformed aggregations now reject; two repository-internal sentinels moved without aliases. No wire/storage migration, reset or compatibility layer is introduced.

### GP-05 completion criteria (historical; superseded by GP-06 review)

| Criterion | Status | Evidence / remaining gate |
| --- | --- | --- |
| Baseline, authorization, adopted owner reconciliation and sequential checkpoints | DONE | E-030–E-037; existing tracker changes preserved; each successor entered after predecessor checkpoint |
| S-006 duplicate/cycle rejection and valid canonical behavior | DONE | E-031/E-032; initial failures observed; V-001/V-017–V-019 pass |
| S-007 registry ownership and current-binding admission | DONE | E-033/E-034; metadata isolation, callbacks, pre-clear rejection and rollback evidence |
| S-008 private implementation simplification | DONE | E-035; semantic/identity/bounds/cancellation checks and Go lint pass |
| S-009 consumer-owned errors, specification correction and positive guard | DONE | E-036; production callers updated atomically; consumer/race/static/generated checks pass |
| Required finalization and executable-source freeze | DONE | E-037; optional retained-run maintenance omitted with reason |
| Final-source product/static portfolio and complete Graph-owner audit | DONE | E-038/E-039; all required rows and checks pass on the frozen source |
| Final Markdown lint | DONE | V-007/E-041; passed after final evidence reconciliation |
| Final baseline comparison, tracker reconciliation and handoff | DONE | E-040/E-042; S-010 completed last; existing changes and historical evidence preserved |

Rollback is a coherent authored-change revert plus Make regeneration of routing projections. No data conversion or automatic deletion is permitted. E-038/E-039 name exact compatible evidence roots. Browser/visual, performance, full-repository, deployment and release qualification were not run because they are outside the accepted scope. No required product/static check was skipped. Finalization's optional canonical-evidence and scheduler-retained-run actions were skipped (`results-dir-not-provided`): there was no compatible full warm-check root. The separately required Graph owner audit did run and pass. This is bounded module/direct-consumer hardening, not deployment or release certification.

### GP-06 historical completion criteria (superseded by GP-07)

| Criterion | Status | Evidence / gate |
| --- | --- | --- |
| S-011 owner and tracker reconciliation | DONE | E-044; V-007 passed |
| S-012 typed dependencies / F-016 | DONE | E-046/E-047 |
| S-013 ordered execution / F-017 | DONE | E-048/E-049 |
| S-014 canonical field paths / F-018 | DONE | E-050–E-053; positive/rejection matrix and exact baseline producer bindings |
| Required finalization and source freeze | DONE | V-016/E-054; generated unchanged; 7,643 executable entries frozen |
| Complete compatible Graph owner evidence | DONE | V-008/V-026; E-055/E-056 |
| Final peer/static/document portfolio | DONE | E-055/E-058; all owning exits 0, summaries/scans pass, cleanup succeeds |
| S-015 preservation and final handoff | DONE | E-057/E-059; source/index/history preserved; completion recorded last |

### GP-07 historical completion criteria

| Criterion | Status | Evidence / gate |
| --- | --- | --- |
| S-016 specification and acceptance reconciliation | DONE | E-061; V-007 passed |
| S-017 numeric grouping / F-019 | DONE | E-062/E-063 and supplemental E-069; final V-008/V-026 pass |
| S-018 definition contexts / F-018 | DONE | E-064/E-065; required Graph/NF checks pass |
| S-019 bounded emission / F-020 | DONE | E-066/E-067; required checks pass |
| S-020 contributor order / F-021 | DONE | E-068/E-069; order, restore and production binding checks passed before S-021 |
| S-021 final qualification and handoff | DONE | E-070–E-075: compatible final portfolio and owner audit, unchanged staged content and frozen source, final-text V-007 |

### GP-08 historical completion criteria

F-022–F-028 closed on the 48-path GP-08 candidate (36 modified, 12 added), subsequently committed as `9203107cd`. Section 10 records the red witnesses, focused repairs, final Graph and peer roots, native exits and explicit audit. The staged executable content matches the frozen inventory and the staged documentation matches the text checked by final lint. The candidate-tree receipt identifies the exact index qualified for commit.
