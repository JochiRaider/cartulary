# Cartulary Modular Refactor Planning Framework

## 1. Purpose and use

This framework is a reusable local planning artifact for generating specific module-by-module refactoring plans from the current repository state. It assumes Cartulary remains a modular monolith and that production code, test ownership, and evidence accounting are organized around durable module, platform, application, package, web, and harness owners.

Use this file before creating a Codex `/goal` prompt or before asking a local agent to plan or implement a refactor slice. The current repository is the final source of code truth. This framework does not claim that any repository file, test, package, or import currently exists until the local agent has inspected it.

## 2. Source and authority posture

The refactor plan must follow this authority order:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for current implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides for terminology, package boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior analysis files as evidence, not authority.

When owner documents conflict, mark `BLOCKED: owner contradiction` and do not pick a side. When repo state conflicts with the framework, record the conflict in the handoff and adapt the plan to the repo state without inventing behavior.

## 3. Refactor doctrine

A refactor plan must preserve observable behavior unless the scoped task explicitly
authorizes a behavior change or an adopted owner requires a normative correction.
Existing behavior is not a compatibility requirement merely because it exists.
When correction is authorized, separate it from structural movement, identify the
owner clause and migration impact, and characterize the intended behavior before
removing the old path. Observable behavior includes route shape, request and
response envelopes, WebSocket paths and event semantics, workbook interaction
behavior, storage semantics, authorization outcomes, generated contract surfaces,
and harness accounting.

The implementation must move from phase-shaped or UI-shaped production code toward module-shaped production code. Test rows and visual fixtures use semantic owner/family identities; historical delivery phases are not an execution, accounting, or compatibility boundary.

A module boundary is valid only when it hides a real design decision. Prefer deep modules with small public facades and private complexity over shallow helper scattering.

## 4. Module boundary catalog

Use this table as the default target-module registry. A local plan may split a row into smaller seams, but it must not merge unrelated concerns without an explicit reason.

| Module or package | Public responsibility | Complexity to hide |
| --- | --- | --- |
| `auth` | Login, session, credential lifecycle, safe auth outcomes. | Password/TOTP verification, session expiry, bootstrap, revocation, audit. |
| `incidents` | Incident lifecycle, membership, visible incident collection. | Visibility rules, versioning, membership audit, close/reopen races. |
| `artifacts` | Structured artifact source semantics for notes, coordination artifacts, findings, investigative queries, and forensic keywords. | Exact surface admission, subtype validation/defaults, authoritative artifact persistence, collection-field policy, source mutation atomicity, and thin owner contributions. It does not own workbook UI, generic relationship persistence, revision mechanics, projection lifecycle, or reporting orchestration. |
| `timeline` | Low-friction timeline capture and mutation. | Rough capture validation, mention extraction, row versions, projection triggers. |
| `entities` | Hosts, identities, parties, mentions, stubs, resolution. | Alias handling, provenance, auto/manual resolution, merge/dedupe. |
| `indicators` | Canonical indicators and observations. | Defanging, observation derivation, lifecycle intervals. |
| `evidence` | Evidence records, object blobs, handles, preview/download, attach/finalize. | Object-store details, safe preview states, blocked states, blob lifecycle. |
| `imports` | CSV/XLSX onboarding beyond ordinary clipboard paste. | Parser quirks, region detection, warnings, mapping fingerprints, provenance. |
| `extensions` | Immutable extension registry admission, claim resolution, serving-epoch plan construction, state admission, validation-condition admission, and pure deadline policy. | Generated catalog integrity, dependency closure, collision checks, initialization/migration protocol, and typed projections for application composition. |
| `incidentbundles` | Incident bundle export/import and extension portability orchestration. | State/claim blocking, participant admission, bounded preparation, scoped staging, and atomic target publication. |
| `crossownertransaction` | Bounded multi-owner final-commit protocol. | Prepare/write separation, global serialization order, cancellation boundaries, typed capabilities, and closed commit outcomes. |
| `stagedobjects` | Operation-scoped staged-byte lifecycle and reconciliation. | Allocation/readiness/abandon/transfer/publication, cutoff draining, retry policy, and readiness/fatal outcomes. |
| `recovery` | Backup capture/catalog, restore, restore verification, and extension binding/codec proof validation. | Artifact integrity, stopped/pristine-target admission, exact codec selection, ordered restore, post-restore validation, and failed-target gating. |
| `links` | Typed relationships, tags, analyst-work coordination links. | Relationship validation, confidence, link projection. |
| `revisions` | Change sets, row revisions, rollback, restore. | Mutation grouping, rollback safety, destructive contention. |
| `projections` | Grid projections, search, sort/filter/group materialization. | Caches, tokenization, disposable derived tables, cursor behavior. |
| `reference_data` | Reference packs and type registries. | Activation, verification, disconnected packs, optional overlays. |
| `reporting` | Snapshots, export model, render/release when claimed. | Immutable snapshot derivation, redaction, release binding. |
| `collaboration` | WebSocket subscriptions, presence, live row updates. | Event ordering, authorization recheck, pending queue interaction. |
| `/apps/web` | Workbook shell, controllers, query state, mutation submission, conflicts, inspector, presence. | Browser state, continuity, pending replay, status feedback. |
| `/packages/grid-adapter` | Direct grid vendor integration and Cartulary-native grid API. | Vendor coordinates, focus, selection, paste/fill, grouping, styling. |
| `/packages/view-contracts` | TypeScript adapters around generated view-schema contracts. | Contract parsing, field metadata, capabilities. |
| `/packages/ui-contracts` | Runtime-safe selector and test-id builders. | Stable UI selectors shared across runtime and tests. |
| `/packages/test-utils` | Browser/helper choreography for tests. | Wait discipline, fixture actions, diagnostics. |

## 5. Top-level work tracker

Copy this tracker into each concrete plan and update it at every checkpoint.

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define target module and scope | scope | TODO | none | TODO | target declaration | One module/seam and exclusions are explicit. |
| T-002 | Inspect current repo state | discovery | TODO | T-001 | TODO | inventory notes | Relevant files, imports, tests, generated paths, and commands are listed. |
| T-003 | Map owner contracts | contracts | TODO | T-002 | TODO | owner map | Public behavior and owner docs are mapped. |
| T-004 | Freeze characterization evidence | tests | TODO | T-003 | TODO | test matrix | Existing and missing characterization tests are known. |
| T-005 | Plan boundary guardrails | architecture | TODO | T-003 | TODO | guardrail plan | Import/generated/domain-platform guardrails are defined or gaps recorded. |
| T-006 | Plan behavior-preserving moves | implementation | TODO | T-004,T-005 | TODO | slice plan | Smallest safe move sequence is defined. |
| T-007 | Plan validation loop | validation | TODO | T-006 | TODO | command list | Cheapest sufficient validation targets are named. |
| T-008 | Update docs/contracts if required | docs | TODO | T-003 | TODO | doc patch plan | Owner docs or derived contracts are planned before codegen. |
| T-009 | Execute or hand off | handoff | TODO | T-006,T-007,T-008 | TODO | handoff log | Next actor can continue without rediscovery. |

Status values: `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, `DROPPED`.

## 6. Workflow dependency map

Every concrete plan must include only workflows that are in scope. A workflow marked `root` has no prerequisite except repository access. A workflow marked `chain` must run after its listed prerequisites. A workflow marked `parallel` may run after prerequisites and does not require peer workflow completion.

| Workflow | Name | Class | Required previous workflows | Required subsequent workflows |
| --- | --- | --- | --- | --- |
| WF-00 | Session and source bootstrap | root | none | WF-01 |
| WF-01 | Current-state repository scan | chain | WF-00 | WF-02, WF-03 |
| WF-02 | Module ownership inventory | chain | WF-01 | WF-04, WF-05 |
| WF-03 | Public contract freeze | chain | WF-01 | WF-04, WF-05 |
| WF-04 | Refactor slice selection | chain | WF-02, WF-03 | WF-05, WF-06 |
| WF-05 | Characterization test plan | chain | WF-03, WF-04 | WF-09 |
| WF-06 | Boundary guardrail plan | chain | WF-02, WF-04 | WF-09 |
| WF-07 | Backend module facade plan | parallel | WF-04, WF-05, WF-06 | WF-09 |
| WF-08 | Frontend package seam plan | parallel | WF-04, WF-05, WF-06 | WF-09 |
| WF-09 | Execution checkpoint plan | chain | WF-05 plus any of WF-07/WF-08 | WF-10 |
| WF-10 | Validation and harness accounting plan | chain | WF-09 | WF-11 |
| WF-11 | Documentation and generated-artifact plan | parallel | WF-03, WF-09 | WF-12 |
| WF-12 | Cleanup and anti-drift plan | chain | WF-10, WF-11 | WF-13 |
| WF-13 | Handoff and next-slice bootstrap | chain | WF-12 | none |

## 7. Workflow details

### WF-00: Session and source bootstrap

**Depends on:** none.  
**Precedes:** WF-01.

Objective: establish the exact local context for one refactor effort.

Steps:
1. Record branch, commit, dirty-tree state, target module, prior analysis path, framework path, and user constraints.
2. Read `AGENTS.md` when present before any edit.
3. Identify owner docs likely to govern the target module.
4. Record source limits and unknowns.

Output:
- Session header.
- Initial source list.
- Explicit scope and non-scope.

Acceptance:
- The target module or seam is named.
- The plan records whether implementation edits are allowed.
- Any missing context is marked `TODO:` rather than guessed.

### WF-01: Current-state repository scan

**Depends on:** WF-00.  
**Precedes:** WF-02 and WF-03.

Objective: inspect actual code, tests, manifests, generated paths, and commands before planning movement.

Steps:
1. Search by filenames, symbols, route names, package imports, field keys, Make targets, and module names.
2. Identify public callers, internal callers, tests, generated consumers, and fixture data.
3. Separate authored source from generated files.
4. Identify current failing tests or stale generated artifacts.

Output:
- File inventory table.
- Import/dependency graph notes.
- Current validation baseline.

Acceptance:
- Every file proposed for movement is directly inspected.
- Search snippets are not treated as final edit source.
- Unseen adjacent files are explicitly listed as unseen.

### WF-02: Module ownership inventory

**Depends on:** WF-01.  
**Precedes:** WF-04 and WF-05.

Objective: assign every in-scope file and behavior to one target module or platform/package boundary.

Inventory table:

| Path | Current responsibility | Target owner | Public/private | External contracts touched | Risk | Test evidence |
| --- | --- | --- | --- | --- | --- | --- |

Acceptance:
- Each production file has one owner or a `TODO: ownership decision required`.
- Platform code is separated from domain logic.
- Shared helpers are justified by semantic ownership, not convenience.

### WF-03: Public contract freeze

**Depends on:** WF-01.  
**Precedes:** WF-04 and WF-05.

Objective: identify behavior that must not drift during the refactor.

Contract surface table:

| Surface | Specific contract | Owner source | Current evidence | Characterization required? |
| --- | --- | --- | --- | --- |
| HTTP | TODO | TODO | TODO | TODO |
| WebSocket | TODO | TODO | TODO | TODO |
| Workbook UI | TODO | TODO | TODO | TODO |
| Storage/revision/projection | TODO | TODO | TODO | TODO |
| Generated artifacts | TODO | TODO | TODO | TODO |
| Harness accounting | TODO | TODO | TODO | TODO |

Acceptance:
- The plan states what is behavior-preserving.
- Any proposed behavior change is separated into its own task and owner-doc decision.

### WF-04: Refactor slice selection

**Depends on:** WF-02 and WF-03.  
**Precedes:** WF-05 and WF-06.

Objective: choose the smallest coherent slice that reduces risk without mixing unrelated behavior.

Slice rule:
- One module seam, one facade, one dependency direction, or one illegal import family per slice.
- Do not combine route behavior changes, schema changes, UI redesign, and test harness changes unless they are inseparable and the plan explains why.

Acceptance:
- The slice has an explicit stop condition.
- Out-of-scope files and behaviors are listed.
- The plan can be converted into one `/goal` prompt.

### WF-05: Characterization test plan

**Depends on:** WF-03 and WF-04.  
**Precedes:** WF-09.

Objective: preserve current external behavior before moving code.

Steps:
1. Reuse existing unit, integration, E2E, visual, and harness evidence when owner-aligned.
2. Add characterization tests only where movement risk is high and behavior lacks coverage.
3. Keep tests named by behavior and owner IDs where available, not by phase alone.

Acceptance:
- Risky movement has pre-move evidence or an explicit `BLOCKED: missing characterization`.
- Tests identify the observed behavior, not just implementation details.

### WF-06: Boundary guardrail plan

**Depends on:** WF-02 and WF-04.  
**Precedes:** WF-09.

Objective: prevent the refactor from creating new dependency leaks.

Default guardrails:
- Domain logic must not live under platform transport/storage packages.
- Direct `react-data-grid` imports must remain inside `/packages/grid-adapter`.
- Generated files must not be hand-edited.
- Modules must not import peer internals.
- Shared semantic coordinators must depend on narrow typed owner ports, never a
  broad peer-module facade or platform SQL/storage DTO.
- Application composition is the edge that translates immutable owner catalogs
  into executable adapters; profile owners and shared coordinators must not
  import one another to discover implementations.
- Public Make targets, output behavior, and artifacts must remain Make-owned when harness evidence is used.

Acceptance:
- Existing violations are listed separately from new violations.
- Each planned guardrail has a validation command or a `TODO:` owner.

### WF-07: Backend module facade plan

**Depends on:** WF-04, WF-05, and WF-06.  
**Precedes:** WF-09.  
**Use only when backend code is in scope.**

Default facade shape:
- `service` or equivalent public module facade.
- Command/query DTOs.
- Closed public module errors or mapped route errors.
- Private store interface and persistence implementation.
- Projection/collaboration hooks only through stable contracts.

Acceptance:
- Callers depend on facade/DTOs, not SQL rows, parser internals, projection-table details, or helper packages.
- Persistence moves do not change public behavior.

### WF-08: Frontend package seam plan

**Depends on:** WF-04, WF-05, and WF-06.  
**Precedes:** WF-09.  
**Use only when frontend code is in scope.**

Default package direction:
- `/apps/web` owns workbook shell and application state.
- `/packages/grid-adapter` owns vendor integration.
- `/packages/view-contracts` owns generated contract adapters.
- `/packages/ui-contracts` owns selectors and test IDs.
- `/packages/test-utils` owns browser helper choreography.

Acceptance:
- `/apps/web` does not learn vendor coordinate semantics.
- UI changes preserve grid-first capture, focus/selection continuity, and inspector default-closed behavior unless owner docs say otherwise.

### WF-09: Execution checkpoint plan

**Depends on:** WF-05 plus any implementation workflow used.  
**Precedes:** WF-10.

Objective: convert the slice into ordered edits.

Checkpoint format:

| Checkpoint | Edit scope | Validation | Expected diff | Rollback point |
| --- | --- | --- | --- | --- |
| CP-01 | TODO | TODO | TODO | TODO |

Acceptance:
- Each checkpoint is small enough to review.
- Validation is attached to each checkpoint.
- The plan has a rollback point before high-risk movement.

### WF-10: Validation and harness accounting plan

**Depends on:** WF-09.  
**Precedes:** WF-11.

Objective: identify the cheapest sufficient proof that behavior was preserved.

Validation table:

| Target | Purpose | Required? | Expected artifact | Failure handling |
| --- | --- | --- | --- | --- |
| `make check` | Default local correctness gate | TODO | TODO | TODO |
| `make test` | Broad test sweep | TODO | TODO | TODO |
| module-specific target | Fast slice check | TODO | TODO | TODO |
| browser/visual/a11y target | UI readiness only unless owner says otherwise | TODO | TODO | TODO |

Acceptance:
- The plan distinguishes product failures from harness/config/infra failures.
- Visual and accessibility evidence are not promoted into product conformance unless the owner contract allows it.

### WF-11: Documentation and generated-artifact plan

**Depends on:** WF-03 and WF-09.  
**Precedes:** WF-12.

Objective: align owner docs, derived contracts, generated code, and implementation support.

Rules:
- Behavior-affecting changes start in owner docs, then derived contracts, then generated code, then implementation.
- Pure behavior-preserving refactors usually update implementation-support docs or handoff notes only.
- Generated files are refreshed through generators, not edited manually.

Acceptance:
- The plan names whether docs are required, optional, or not applicable.
- Generated-artifact drift checks are included when generated outputs change.

### WF-12: Cleanup and anti-drift plan

**Depends on:** WF-10 and WF-11.  
**Precedes:** WF-13.

Objective: remove obsolete paths without deleting useful evidence.

Steps:
1. Remove dead wrappers, duplicate helpers, and old imports only after callers are moved.
2. Re-run import-boundary and generated-artifact checks.
3. Review diffs for out-of-scope behavior changes.
4. Record intentional deferrals.

Acceptance:
- No orphaned old path remains unless explicitly deferred.
- No out-of-scope diff remains.
- Handoff names any residual debt.

### WF-13: Handoff and next-slice bootstrap

**Depends on:** WF-12.  
**Precedes:** none.

Objective: preserve continuity across repeated refactor sessions.

Acceptance:
- Top tracker is current.
- Commands run and results are recorded.
- Changed files are listed by workstream.
- Remaining blockers and next recommended workflow are explicit.

## 8. Workstream notes

Use this section during a live session. Keep entries terse and factual.

### Scope and evidence

| Date | Note | Source or command | Impact |
| --- | --- | --- | --- |
| TODO | TODO | TODO | TODO |

### Contracts and docs

| Date | Owner section | Decision or conflict | Action |
| --- | --- | --- | --- |
| TODO | TODO | TODO | TODO |

### Backend modules

| Date | Module | Files | Current state | Next action |
| --- | --- | --- | --- | --- |
| TODO | TODO | TODO | TODO | TODO |

### Frontend packages

| Date | Package | Files | Current state | Next action |
| --- | --- | --- | --- | --- |
| TODO | TODO | TODO | TODO | TODO |

### Tests and harness

| Date | Target | Result | Artifact | Follow-up |
| --- | --- | --- | --- | --- |
| TODO | TODO | TODO | TODO | TODO |

### Generated artifacts

| Date | Generator or target | Outputs | Drift status | Follow-up |
| --- | --- | --- | --- | --- |
| TODO | TODO | TODO | TODO | TODO |

### Risks and blockers

| ID | Risk or blocker | Owner | Blocking workflow | Resolution condition |
| --- | --- | --- | --- | --- |
| B-001 | TODO | TODO | TODO | TODO |

## 9. Session handoff

Append one handoff record before ending a session.

### Handoff record template

| Field | Value |
| --- | --- |
| Date/time | TODO |
| Branch/commit | TODO |
| Target module or seam | TODO |
| Current workflow | TODO |
| Completed workflows | TODO |
| Changed files | TODO |
| Commands run | TODO |
| Passing validation | TODO |
| Failing validation | TODO |
| Decisions made | TODO |
| Open questions | TODO |
| Blockers | TODO |
| Next recommended workflow | TODO |
| Safe restart command | TODO |

### Handoff requirements

- Do not claim validation passed unless the exact command ran in this session or the retained artifact is named.
- Do not claim a file was preserved, compared, or verified unless it was inspected.
- Use `TODO:` for missing evidence.
- Record whether dirty worktree changes are intentional.
- Record any generated files that need regeneration or drift checks.

## 10. Top-level checklist

- [ ] Target module or seam is explicit.
- [ ] Source limits are recorded.
- [ ] Current repo state is inspected.
- [ ] Owner contracts are mapped.
- [ ] Public behavior to preserve is listed.
- [ ] Characterization coverage is identified.
- [ ] Boundary guardrails are planned.
- [ ] Implementation slice is small and reviewable.
- [ ] Docs/contracts/generation needs are classified.
- [ ] Validation commands are named.
- [ ] Handoff notes are current.
- [ ] No out-of-scope behavior change is planned.
- [ ] No generated file hand-edit is planned.
- [ ] No phase-shaped runtime dependency is introduced.

## 11. Binary acceptance criteria for a concrete refactor plan

| ID | Criterion |
| --- | --- |
| RF-AC-001 | The plan names exactly one primary target module, package, or seam. |
| RF-AC-002 | The plan lists all inspected in-scope source files and marks all relevant unseen files. |
| RF-AC-003 | The plan maps every public contract surface that could drift. |
| RF-AC-004 | The plan separates behavior-preserving refactors from behavior changes. |
| RF-AC-005 | The plan states required characterization tests or explains why existing evidence is sufficient. |
| RF-AC-006 | The plan contains a checkpoint sequence with validation after each risky move. |
| RF-AC-007 | The plan preserves module boundaries and package import boundaries. |
| RF-AC-008 | The plan does not hand-edit generated files. |
| RF-AC-009 | The plan does not make phase identity a runtime production dependency. |
| RF-AC-010 | The handoff section is sufficient for another session to resume without rediscovery. |

## Sources

[^core00]: `00_document_set_status_and_precedence.md`, §§1-5, lines 5-20, 24-35, 53-68, and 95-105. Used for authority, conformance, and owner-section posture.
[^core01]: `01_architecture_storage_and_view_contracts.md`, §§1-2, lines 5-26 and 30-79. Used for modular-monolith topology, required module boundaries, import isolation, and clipboard/import separation.
[^core02]: `02_domain_model_schema_and_history.md`, §§1-2, lines 5-17 and 50-53. Used for raw/canonical separation, source/projection separation, rough-capture preservation, and inspector non-workflow boundary.
[^core03]: `03_workbook_interaction_collaboration_and_workflows.md`, §§1-2, lines 5-28 and 36-54. Used for grid-first interaction, workbook surface, inspector role, and required built-in/system surfaces.
[^devguide]: `cartulary-dev-guide.md`, §§2.3.1, 3.1, and 4.1, lines 201-219, 377-403, and 415-421. Used for frontend grid-adapter isolation, package boundaries, generated-file policy, and contract derivation.
[^harness]: `testing-harness-nlspec.md`, §§1-2, lines 28-40 and 44-56. Used for canonical Make invocation, generated-file prohibition, direct-command boundary, and harness artifact requirements.
[^implguide]: `cartulary_implementation_testing_guide.md`, §1, lines 12-20 and 24-37. Used for phase-as-planning posture, test categories, mutation substrate timing, and test accounting discipline.
[^nlspec]: `nlspec-spec.md`, “Why NLSpecs Work When They Work,” lines 30-72. Used for behavioral completeness, interface precision, defaults/bounds, mapping tables, and binary acceptance criteria.
