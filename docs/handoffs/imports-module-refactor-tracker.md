# Imports Production-Readiness Refactoring Tracker

## 1. Authority and Current Status

| Field | Value |
| --- | --- |
| Target | `internal/modules/imports` |
| Tracker | `docs/handoffs/imports-module-refactor-tracker.md` |
| Repository baseline | Branch `main`; commit `9ce2db1c9f91fc611438fa404b9f897c72efbb20`; clean worktree observed at `2026-08-30T21:22:58-04:00` |
| Current package inventory | 38 Go files under `internal/modules/imports`, including `ownerfacade`; WS-24 removed two historical source-shape test files after routing durable replacements |
| Previous iteration | WS-00 through WS-18 and RB-001 through RB-006 are closed |
| Current iteration | WS-20 through WS-26 are complete; the Imports remediation is closed |
| Authorized change in this task | Complete WS-20 through WS-26 serially, including specification, production, test, harness, generated, validation, and tracker changes. |
| Implementation authority | The user explicitly authorized the complete remediation plan on 2026-08-30. |

Adopted Core owner sections and subsystem NLSpecs remain authoritative for
behavior in their scopes. Typed contracts are machine projections of those
owners. This tracker controls sequencing and handoff only; it does not create,
replace, or amend product requirements.

The applicable source order is:

1. Adopted subsystem NLSpecs for their named subsystem.
2. Core 00 through Core 04 for current product behavior and conformance.
3. `docs/domain.md` for vocabulary and owner navigation.
4. Authored contracts and harness inputs for executable projections and routing.
5. Current code and tests for implementation state.
6. This tracker and prior handoff evidence for execution history only.

Core 05 is not applicable to this document update because it publishes no
timed, benchmark, fixture-sensitive, or claim-bearing result.

## 2. Previous Iteration Closure Archive

The previous remediation established the current secure and owner-aligned
baseline. Its detailed discovery notes, superseded slices, historical TODOs,
and per-session logs have been retired from the controlling tracker. The
closure facts below are retained because they constrain future work.

| Workstream | Status | Retained result |
| --- | --- | --- |
| WS-00 | DONE | Rebaselined the original remediation and installed serial execution controls. |
| WS-01 | DONE | Reconciled executable Network Flow major 5, state 4, and Extensions 0.10.0; major 4 is unsupported. WS-20 repairs one missed owner-specification row. |
| WS-02 | DONE | Adopted the exhaustive Imports route-security matrix. |
| WS-03 | DONE | Enforced cookie CSRF on all six state-changing routes while preserving bearer and read-only behavior. |
| WS-04 | DONE | Separated HTTP transport from private application operations. |
| WS-05 | DONE | Installed exact, fail-closed transport-file accounting in boundary schema v3. |
| WS-06 | DONE | Characterized file and clipboard mapping bytes and semantics. |
| WS-07 | DONE | Adopted the shared pure tabular mapping kernel contract. |
| WS-08 | DONE | Routed file and clipboard view-target mapping through the shared kernel without byte drift. |
| WS-09 | DONE | Characterized every source-owner finalization caller and effect. |
| WS-10 | DONE | Migrated Artifacts to owner-local Revisions, Projections, and Collaboration ports. |
| WS-11 | DONE | Migrated Parties while preserving create/reuse distinctions. |
| WS-12 | DONE | Migrated Indicators to canonical owner identity and changed-field facts/effects. |
| WS-13 | DONE | Added Assessment-owned fact/effect derivation. |
| WS-14 | DONE | Migrated Host behavior independently. |
| WS-15 | DONE | Migrated Identity behavior independently. |
| WS-16 | DONE | Migrated Tasks and Decisions while preserving mutation sequence and link ordering. |
| WS-17 | DONE | Removed the aggregate revision/publication finalizer and every compatibility alias. |
| WS-18 | DONE | Completed narrow-to-broad verification and closed RB-001 through RB-006. |

### 2.1 Retained final evidence

| Validation | Retained result |
| --- | --- |
| Imports service-backed | 14/14 passed at `.cartulary/test-results/20260830T175638Z-p1417244` |
| Browser-backed | 60/60 passed at `.cartulary/test-results/20260830T175757Z-p1462105` |
| Refreshed fast suite | 441/441 passed at `.cartulary/test-results/20260830T181358Z-p1697214` |
| Finalization | Passed at `.cartulary/test-results/20260830T181452Z-p1706292`; retained-run maintenance was skipped because `RESULTS_DIR` was unset |
| Repository check | 671/671 passed at `.cartulary/test-results/20260830T181512Z-p1709285` |
| Markdown checkpoint | Passed at `.cartulary/test-results/20260830T182205Z-p1826723` |

The retained results prove the previous baseline only. They are not reusable
success evidence for WS-20 through WS-26 after production or harness changes.

### 2.2 Non-regression decisions

- Network Flow major 4, the combined finalizer, and duplicate mapping paths
  remain intentionally unsupported; no shim or forwarding alias may return.
- `http_handlers.go` remains the sole Imports transport-binding file unless an
  owner-backed route change deliberately updates exact-file accounting.
- File and clipboard view targets continue to use the shared mapping kernel;
  analytical mapping remains target-owned.
- Imports owns session, unit, source, mapping integrity, dispatch, job, outcome,
  and finalizer coordination. Source and analytical owners retain their exact
  durable resource semantics.
- Existing cookie-CSRF admission and diagnostic precedence remain unchanged.

## 3. Current Findings and Retention Boundaries

| ID | Finding | Evidence | Required disposition |
| --- | --- | --- | --- |
| PR-001 | The non-transactional discovery update and obsolete rectangle helper had no production or test caller. | WS-23 deleted both declarations; the transactional discovery path and internal region rectangle remain the only live behaviors. | DONE — no alias, wrapper, declaration, or call remains. |
| PR-002 | Imports analytical facades were transported through an untyped `ModuleOverrides` map and assembled with a string key helper. | WS-22 replaced the override with self-identifying facade contributions checked field-for-field against generated binding facts. | DONE — no Imports override, key builder, decoder, or server merge path remains. |
| PR-003 | Route construction accepted independently optional Jobs collaborators and silently skipped worker registration when the runner was absent. | WS-22 introduced one fail-closed module constructor and separate explicit route/worker registration. | DONE — incomplete composition fails before registration or serving. |
| PR-004 | Application assembly constructed the general Imports store solely to satisfy Network Flow's narrow source and transaction ports. | WS-21 replaced that edge with a two-method source capability carrying bounded filename provenance and direct transactional validation. | DONE — the server no longer constructs or exports the general store to Network Flow. |
| PR-005 | The Imports root exported transport, persistence, and application implementation details without live cross-package consumers. | WS-23 privatized the general store/service and package-local commands, results, errors, decoders, DTOs, methods, and constants. An AST contract now classifies every retained root and `ownerfacade` export by production role. | DONE — unclassified, stale, empty-role, or duplicate-role inventory entries fail. |
| PR-006 | Go tests froze historical filenames, removed paths, retired symbols, and exact source spellings already covered by stronger behavior or repository policy. | WS-24 moved Imports-to-peer ownership into the authored backend boundary policy, added checker fixtures and architecture routing, deleted both source-shape test files, and replaced the old harness row with composition/export contracts. | DONE — ownership is checked structurally and product invariants remain behavior-tested. |
| PR-007 | Tests and harness rows used “characterization” names after the behavior became adopted/current. | WS-25 renamed Imports and `ownerfacade` files, tests, fixtures, helpers, messages, and facade fixture identity to contract/compatibility terms; the old row and selectors were replaced without aliases. | DONE — active Imports test and harness artifacts contain no discovery-phase naming. |
| PR-008 | NF-REQ-088d still declares analytical-facade `contract_major=2` while its owner reference, adopted profile, typed binding, and generated registry are major 5. | Owner/specification comparison found the stale exact-value row; machine projections already agree on 5. | Amend the Network Flow NLSpec to 5; retain no major-2 or major-4 compatibility path. |

### 3.1 Active compatibility that must remain

The following versioned or durable surfaces may look historical but are active
contracts and are not cleanup targets:

- HTTP routes, methods, operation IDs, envelopes, error codes, and security order.
- Parser/profile identifiers, job kinds and handler identities, result codes,
  import resource references, and recovery `VNext` contribution identities.
- Approved-mapping bytes, mapping fingerprints, replay hashes, unit plans,
  outcomes, journals, source-stream references, and provenance.
- Current `v1` schema, binding, target-registry, and error-translation IDs.
- Generated target admission, Network Flow facade payloads, and source-owner
  facade semantics.

Any workstream that discovers a required change to one of these surfaces must
stop. The change requires the affected owner amendment, projection update,
migration or compatibility decision, and separately authorized implementation.

## 4. Controlling Workstreams

Execution is serial. Only one implementation workstream may be in progress,
and no dependent row may begin until its predecessor is `DONE`.

### WS-19 — Rebaseline the tracker

**Status:** DONE

**Scope:** This document only.

Replace the completed remediation's verbose controlling material with this
compact closure archive and install its successor work. WS-19 closes when the
baseline is accurate, implementation remains unstarted, the final diff touches
only this tracker, and `make lint-markdown` passes.

**Closure:** The tracker was reduced from 946 to 325 lines before final status
updates, the baseline and future workstreams were revalidated, the diff touches
only this file, and Markdown validation passed at
`.cartulary/test-results/20260831T012821Z-p32338`.

### WS-20 — Reconcile authority and rebaseline

**Status:** DONE

**Depends on:** WS-19

Correct NF-REQ-088d to analytical-facade major 5 and update Core 04 AC-465 to
use generated `availability_kind='enabled'` and owner-contract-test language.
Rebaseline this tracker to PR-008 and WS-20 through WS-26. `docs/domain.md`
already states the correct ownership boundary and remains unchanged.

**Exit:** Owner text agrees with the existing major-5 projection;
`make lint-markdown`, `make json-shape-check`, and `make generate-drift` pass.

**Closure:** NF-REQ-088d now uses major 5, Core 04 AC-465 now names the
generated enabled-target and owner-contract-test evidence, and this tracker
contains the complete serial WS-20 through WS-26 execution. No machine
projection changed. Markdown passed at
`.cartulary/test-results/20260831T020423Z-p48191`, JSON shape passed at
`.cartulary/test-results/20260831T020515Z-p49285`, and generated drift passed at
`.cartulary/test-results/20260831T020519Z-p49708`. Compatibility and migration:
owner text only; runtime major 5 is unchanged. Rollback boundary: the two owner
text corrections plus this tracker rebaseline. Residual risk: none identified.

### WS-21 — Introduce the exact analytical source port

**Status:** DONE

**Depends on:** WS-20

Add `ExtensionSourcePort` and `NewExtensionSourcePort(postgres.DB)` with only
source-stream opening and transactional apply-precondition validation. Carry
the original filename on the internal stream result, migrate Network Flow away
from the session-resource map and transaction type assertion, and remove the
server's general-store construction.

**Exit:** Network Flow consumes only the exact port; source bytes, provenance,
filename behavior, and atomic apply are unchanged.

**Closure evidence:** `ExtensionSourcePort` now owns the exact stream-open and
transactional precondition capabilities over the narrow `postgres.DB` port.
The stream result carries `OriginalFilename`, so Network Flow no longer reads
the session resource map, and its consumer-owned interface requires the
transaction method directly with no type assertion or late capability failure.
Server assembly constructs this port before Network Flow and contains no
`imports.NewStore` call. Changed production files:
`internal/modules/imports/source_streams.go`,
`internal/modules/imports/regions.go`,
`internal/modules/networkflow/module.go`,
`internal/modules/networkflow/import_facade.go`,
`internal/modules/networkflow/transaction_participants.go`, and
`internal/app/server/runtime_assembly.go`.

Compatibility and migration: repository-internal constructor/interface cutover
only; source bytes, digest validation, source identity, filename sanitization,
and in-transaction incident/source revalidation remain on the same durable
paths. No alias, fallback, database, wire, frontend, or owner-resource change
was introduced. `make test-slice OWNER=module.imports` passed 23/23 at
`.cartulary/test-results/20260831T021032Z-p91343`; the Network Flow unit slice
passed 34/34 at `.cartulary/test-results/20260831T021152Z-p36904`; Imports
service-backed passed 14/14 at
`.cartulary/test-results/20260831T021402Z-p95575`; and Network Flow
service-backed passed 28/28 at
`.cartulary/test-results/20260831T021516Z-p39805`. An earlier Imports unit run
failed at `.cartulary/test-results/20260831T020946Z-p59911`; it was related to
the change, classified as a compile-time unused import, corrected without an
interface fallback, and superseded by the passing run. Rollback boundary: all
six production files above as one source-port cutover. Residual risk: broad
server and browser composition remains for WS-26; focused owner and
service-backed behavior is clean.

The mandatory Markdown checkpoint passed at
`.cartulary/test-results/20260831T021804Z-p98199`.

### WS-22 — Replace composition and analytical injection

**Status:** DONE

**Depends on:** WS-21

Introduce `NewModule(ModuleDependencies)`, explicit route and worker
registration, complete self-describing analytical facade bindings, generated
binding identity/major projections, and exhaustive startup validation. Remove
all Imports route options, override keys, string-key builders, map decoding,
and the server merge helper in one atomic cutover.

**Exit:** Invalid composition fails at startup; valid composition binds 11
operations and two workers; Imports has no `ModuleOverrides` path.

**Closure evidence:** `NewModule(ModuleDependencies)` now validates the narrow
PostgreSQL port, each Jobs collaborator, all Imports/archive limits, the owner
registry, Revisions appender, profile-admission resolver, job finalizer, auth
key material, and analytical contributions before producing a module. Clock
injection and cursor fallback remain explicit. Construction has no route or
worker side effect; `Module.RegisterRoutes` binds the 11 generated operations,
and `Module.RegisterWorkers` registers discovery and apply together. The
general store now uses `postgres.DB` and `BeginTx`. Server assembly constructs
the module directly, validates the exact published Imports worker job set,
registers workers explicitly, and preserves unrelated `ModuleOverrides` only.

`ExtensionImportFacade.Binding` now returns all 14 Table 17.2-F facts. The
private selector registry rejects nil and typed-nil contributions, unknown or
duplicate selectors, duplicate facade IDs, every metadata mismatch, and a
claimed target without its facade; a known unclaimed target may omit it.
Network Flow publishes the complete major-5 binding. The contract generator
projects binding schema ID and contract major into both generated Go target
and adapter descriptors; `make generate` passed at
`.cartulary/test-results/20260831T022350Z-p2756` and final generated drift
passed at `.cartulary/test-results/20260831T024109Z-p1419`.

Changed authored production/generator files:
`internal/modules/imports/service.go`, `extension_facade.go`, `targets.go`,
`jobs.go`, `store.go`, `source_streams.go`, `regions.go`, `unit_outcomes.go`,
`apply_coordination.go`, `mapping.go`, `application_operations.go`,
`internal/modules/networkflow/import_facade.go`,
`internal/app/extensionassembly/publication_catalog.go`,
`internal/app/server/runtime_assembly.go`, and
`tools/contractgen/import_targets.go`. Generated projections changed only in
`internal/gen/importtargetregistry/registry_gen.go` and
`internal/gen/contractimports/artifacts_gen.go`. Contract tests changed or were
added in `internal/modules/imports/module_composition_test.go`,
`owner_errors_test.go`, `boundary_guard_test.go`, Network Flow's contract test,
and `tools/contractgen/import_targets_test.go`.

Focused validation passed: Imports unit 23/23 at
`.cartulary/test-results/20260831T023924Z-p52198`; app-server unit 24/24 at
`.cartulary/test-results/20260831T023319Z-p45752`; Network Flow unit 34/34 at
`.cartulary/test-results/20260831T024141Z-p4586`; Imports service-backed 14/14
at `.cartulary/test-results/20260831T024353Z-p63331`; app-server
service-backed 17/17 at `.cartulary/test-results/20260831T024507Z-p7685`; and
Network Flow service-backed 28/28 at
`.cartulary/test-results/20260831T024605Z-p49504`. JSON shape passed at
`.cartulary/test-results/20260831T024823Z-p7879`. Two related intermediate
Imports runs failed at `.cartulary/test-results/20260831T022854Z-p11201`
(test facade lacked the new binding method) and
`.cartulary/test-results/20260831T023009Z-p56240` (the existing responsibility
test still named the retired constructor); both were corrected in tests and
superseded by the passing slice.

Compatibility and migration: atomic repository-internal Go composition and
generated-projection cutover only. No alias, fallback injection, public route,
operation, schema, database, durable identity, owner-resource, frontend, or
behavior change was introduced. Rollback boundary: the full module,
facade-binding, generator, Network Flow, and server cutover above. Residual
risk: broad browser and repository validation remains for WS-26; focused
startup, owner, persistence, and transaction evidence is clean.

The mandatory Markdown checkpoint passed at
`.cartulary/test-results/20260831T024911Z-p8590`.

### WS-23 — Delete dead code and close exports

**Status:** DONE

**Depends on:** WS-22

Delete `MarkDiscovered` and `SourceRect`, privatize the general store, service,
transport and persistence details, and add an exact AST export/production-role
guard for the Imports root and `ownerfacade`.

**Exit:** Dead and broad surfaces are absent; every retained export has an
explicit live role and unapproved growth fails.

Implementation closure: the non-transactional discovery method and obsolete
rectangle helper were deleted without forwarders. The general store, service,
transport requests and responses, request decoders, persistence commands and
results, implementation errors, package-local methods, and package-local
identity constants are private. The exact analytical source port, module
composition, limits, typed analytical DTOs and binding, Jobs finalization,
recovery contributions, registry join, and `ownerfacade` cross-owner protocol
remain exported because each has a live production consumer or interface role.

`export_surface_contract_test.go` parses production Go syntax independently of
source filenames and inventories exported declarations, methods, interface
members, embedded fields, and DTO fields. It rejects new, stale, empty, or
multiply classified roles and includes a synthetic unclassified-export
fixture. The existing boundary verification row invokes the guard so focused
owner verification cannot omit it. Repository search confirms no Imports
`NewStore`, exported general store/service type, dead discovery method, or dead
rectangle function remains.

Changed files: all affected production and package-local tests under
`internal/modules/imports`, plus new
`internal/modules/imports/export_surface_contract_test.go`. External
integration tests were changed only to consume durable string identities
instead of implementation-only constants. No other owner, public API,
database, generated product contract, durable identifier, frontend, or runtime
behavior changed in this slice.

Focused validation passed:

- `make test-slice OWNER=module.imports`: 23/23 at
  `.cartulary/test-results/20260831T030908Z-p26480`.
- `make service-backed-test-slice OWNER=module.imports`: 14/14 at
  `.cartulary/test-results/20260831T031026Z-p70874`.
- `make test-slice OWNER=app.server`: 24/24 at
  `.cartulary/test-results/20260831T031149Z-p15329`.
- `make test-slice OWNER=module.networkflow`: 34/34 at
  `.cartulary/test-results/20260831T031721Z-p17927`.
- `make format`: passed at
  `.cartulary/test-results/20260831T030904Z-p22361`.

Related intermediate failures at
`.cartulary/test-results/20260831T025324Z-p16250` and
`.cartulary/test-results/20260831T025423Z-p51511` exposed mechanical rename
collisions and one external test's implementation-constant dependency; both
were corrected without compatibility exports. The intentional empty-inventory
guard failure at `.cartulary/test-results/20260831T030326Z-p27758` established
the exact surface before role classification. The first Network Flow rerun at
`.cartulary/test-results/20260831T031249Z-p57855` was environmental: 33/34
units passed and the remaining unit reported an object-store readiness timeout;
the unchanged-tree retry passed 34/34.

Compatibility and migration: atomic repository-internal caller renames only;
no alias, wrapper, wire, schema, durable-state, database, owner-resource, or
frontend migration was added. Rollback boundary: all WS-23 privacy renames,
dead-code deletions, caller/test updates, and the export-role guard together.
Residual risk: WS-24 must remove the remaining historical source-shape tests,
and WS-26 must run full repository validation. The mandatory Markdown
checkpoint passed at
`.cartulary/test-results/20260831T032027Z-p76286`.

### WS-24 — Replace historical boundary tests

**Status:** DONE

**Depends on:** WS-23

Move peer-owner, Workbook, and Projections import policy into the authored
backend boundary manifest; remove filename, retired-path, retired-symbol, and
call-spelling assertions; retain their underlying invariants through boundary,
compile-time, and behavior evidence; update verification routing and generated
topology through `make generate`.

**Exit:** No active test freezes the completed refactor layout; boundary,
harness, generated-policy, and drift checks pass.

Implementation closure: `tools/backend_module_boundaries.json` now contains
the production-only `imports-use-owner-facades-not-peer-implementations` rule.
It rejects subtree imports of peer source owners, Workbook, and Projections
from every current or future Imports production filename. The boundary checker
has a conforming capability-import fixture plus failing peer-owner, Workbook,
and Projections fixtures whose filenames did not exist during the refactor.
`module.imports.verification.architecture_boundary` routes this owner to
`backend-module-boundary-check` as shell/static architecture evidence.

The historical `boundary_guard_test.go` and `boundary_test.go` files were
deleted. Their path-stat, responsibility-filename, aggregate-symbol blacklist,
peer-import substring, and exact call-spelling assertions have no aliases. The
authored exact transport-file rule and repository-wide aggregate-finalizer
rule remain their single enforcement points. A stale boundary-policy allowance
for the deleted Imports test was also removed.

Every removed product assertion retains stronger executable evidence:

- Generated target projection is checked by generator contracts and the target
  inventory behavior test.
- Owner-facade dispatch is covered by registry contracts and the durable
  view-owner integration row.
- Analytical ownership is covered by typed binding construction and atomic
  Network Flow apply behavior.
- File/clipboard mapping parity is covered by shared-kernel compatibility
  bytes, fingerprints, transforms, and cross-path plans.
- Safe owner/error translation is covered by unit and integration rows.
- Route parity and security are covered by route-contract and CSRF matrix rows.
- XLSX semantics are covered by locator, archive-abuse, bounded-range, hidden
  sheet, and durable-region tests.
- Replay and transaction invariants are covered by upload/region replay, crash
  recovery, current-state revalidation, owner rollback, and atomic analytical
  apply rows.

The former `boundary_complete` selector was removed, not aliased. Authored
`export_surface_contract` and `module_composition_contract` rows now execute
the WS-22/WS-23 contracts directly. `make generate` refreshed
`tools/execution_topology_render_index.json`; no hand edit was made to generated
topology outputs.

Changed files: `tools/backend_module_boundaries.json`,
`tools/harness/static-analysis/backend-module-boundary-check-cli.mjs`,
`contracts/verification/owners/module.imports.json`,
`tools/test_families/module.imports.json`, generated
`tools/execution_topology_render_index.json`, and deletion of the two Imports
boundary source-test files.

Focused validation passed:

- `make test-slice OWNER=module.imports`: 23/23 at
  `.cartulary/test-results/20260831T032639Z-p86416`.
- `make backend-module-boundary-check`: 3/3 at
  `.cartulary/test-results/20260831T032853Z-p36903`.
- `make harness-contract`: 2/2 at
  `.cartulary/test-results/20260831T032757Z-p32398`.
- `make json-shape-check`: 3/3 at
  `.cartulary/test-results/20260831T032811Z-p32880`.
- `make generate`: passed at
  `.cartulary/test-results/20260831T032606Z-p83328`.
- `make generate-drift`: 4/4 at
  `.cartulary/test-results/20260831T032855Z-p37232`.
- `make generated-artifact-policy-check`: 3/3 at
  `.cartulary/test-results/20260831T032822Z-p36206`.

The related intermediate `make generate` failure at
`.cartulary/test-results/20260831T032454Z-p80019` rejected an unsorted authored
row insertion. The rows were reordered by semantic ID and the next generation
passed. Compatibility and migration: test/harness selector replacement only;
no product, wire, schema, durable-state, database, owner-resource, frontend, or
runtime behavior changed. Rollback boundary: the boundary rule and fixtures,
verification contract, test-row replacement, generated render index, and both
test-file deletions together. Residual risk: WS-25 must remove the remaining
discovery-phase naming without changing protected behavior or bytes. The
mandatory Markdown checkpoint passed at
`.cartulary/test-results/20260831T032942Z-p40424`.

### WS-25 — Normalize contract-test vocabulary

**Status:** DONE

**Depends on:** WS-24

Rename Imports and `ownerfacade` characterization files, tests, fixtures,
helpers, and harness selectors to contract/compatibility language. Retain no
old row or selector alias and preserve all golden bytes.

**Exit:** No active Imports test or harness artifact uses discovery-phase
naming; row coverage and protected bytes remain equivalent.

Implementation closure: the Imports route/target/XLSX file is now
`behavior_contract_test.go`; the mapping golden and cross-path parity support
is `mapping_compatibility_contract_test.go`; and the `ownerfacade` normalization
evidence is `compatibility_contract_test.go`. Test functions, workbook/ZIP
helpers, assertion helpers, diagnostics, and the test-only facade identity now
use contract or compatibility vocabulary. Production code was not changed.

The authored `characterization_complete` row and both old test selectors were
removed without migration aliases. `module.imports.unit.behavior_contract` now
selects the generated target registry, route contract, safe internal error,
and hidden-sheet semantics tests. The composition, export-surface,
architecture, owner-facade, compatibility, replay, and transaction rows remain
separate and unchanged in coverage. `make generate` refreshed the execution
topology render index from those authored identities.

Repository-wide active-artifact search confirms that no Imports filename,
test, helper, selector, row, or generated topology artifact contains
“characterization.” Focused Imports verification executed the renamed
behavior-contract row, including the byte-exact approved mapping, normalized
request, fingerprint, owner-request, transform, failure-envelope, and
cross-path parity assertions; all protected goldens remained byte-identical.

Changed files: three Imports test-file renames and their terminology-only
contents, `tools/test_families/module.imports.json`, and generated
`tools/execution_topology_render_index.json`.

Focused validation passed:

- `make format`: 2/2 at
  `.cartulary/test-results/20260831T033218Z-p43398`.
- `make generate`: passed at
  `.cartulary/test-results/20260831T033222Z-p47488`.
- `make test-slice OWNER=module.imports`: 23/23 at
  `.cartulary/test-results/20260831T033237Z-p50285`.
- `make harness-contract`: 2/2 at
  `.cartulary/test-results/20260831T033359Z-p95362`.
- `make json-shape-check`: 3/3 at
  `.cartulary/test-results/20260831T033413Z-p95849`.
- `make generate-drift`: 4/4 at
  `.cartulary/test-results/20260831T033416Z-p96248`.
- `make generated-artifact-policy-check`: 3/3 at
  `.cartulary/test-results/20260831T033424Z-p99172`.

The first `make format` preflight failed for a related authored selector-order
error and produced no usable retained run root. Ordering the renamed tests by
ASCII identity corrected the harness input; the next format and all subsequent
generation/harness checks passed. Compatibility and migration: test identity
replacement only; no legacy selector, product, wire, schema, durable-state,
database, owner-resource, frontend, or runtime compatibility path was added.
Rollback boundary: all three file renames, terminology changes, authored row
replacement, and generated render-index refresh together. Residual risk: only
the final WS-26 narrow-to-broad validation remains. The mandatory Markdown
checkpoint passed at `.cartulary/test-results/20260831T033539Z-p384`.

### WS-26 — Verify and close

**Status:** DONE

**Depends on:** WS-25

Run the complete validation ladder in Section 6, classify every failure, and
record run roots, compatibility outcome, rollback boundary, and residual risk.
Close only when every criterion in Section 8 passes.

Validation result: all prescribed narrow owner slices, boundary/harness and
generation gates, browser coverage, fast tests, finalization, and the complete
672-unit repository check pass. The user explicitly authorized repair of the
remaining grid-adapter CSS lint failure. The scoped header/selection/gutter
cell declarations now follow the lower-specificity declarations they override,
and their four unnecessary `!important` flags are absent. The adapter rules
remain unlayered and therefore retain precedence over React Data Grid's
declared `rdg` layer without exceptional cascade priority.

Final-tree validation completed in the required order:

1. Imports test slice: 23/23 at
   `.cartulary/test-results/20260831T033654Z-p3961`.
2. Imports service-backed slice: 14/14 at
   `.cartulary/test-results/20260831T033807Z-p47752`.
3. App-server test slice: 24/24 at
   `.cartulary/test-results/20260831T033927Z-p91445`.
4. App-server service-backed slice: 17/17 at
   `.cartulary/test-results/20260831T034028Z-p34326`.
5. Network Flow test slice: 34/34 at
   `.cartulary/test-results/20260831T034126Z-p76096`.
6. Network Flow service-backed slice: 28/28 at
   `.cartulary/test-results/20260831T034333Z-p34930`.
7. Backend module boundary: 3/3 at
   `.cartulary/test-results/20260831T034543Z-p92774`.
8. Harness contract: 2/2 at
   `.cartulary/test-results/20260831T034545Z-p93185`.
9. JSON shape: 3/3 at
   `.cartulary/test-results/20260831T034558Z-p93673`.
10. Generation passed at
    `.cartulary/test-results/20260831T034602Z-p94076`.
11. Generated drift: 4/4 at
    `.cartulary/test-results/20260831T034610Z-p96790`.
12. Generated-artifact policy: 3/3 at
    `.cartulary/test-results/20260831T034618Z-p99749`.
13. Webserver-backed browser: 60/60 at
    `.cartulary/test-results/20260831T034624Z-p587`.
14. Fast suite: 443/443 at
    `.cartulary/test-results/20260831T035210Z-p57190`.
15. Agent finalization: 1/1 at
    `.cartulary/test-results/20260831T035423Z-p98015`; retained-run maintenance
    was skipped with `results-dir-not-provided` because no successful current-
    tree full warm-check root existed before the prescribed final `check`.
16. Full check first reached 669/672 at
    `.cartulary/test-results/20260831T035442Z-p1435`.
17. Authorized CSS remediation passed `make lint-biome` 2/2 at
    `.cartulary/test-results/20260831T114012Z-p65734`.
18. Grid-adapter test slice passed 38/38 at
    `.cartulary/test-results/20260831T114041Z-p66655`; its service-backed slice
    passed 13/13 at
    `.cartulary/test-results/20260831T114148Z-p15065`.
19. Final full check passed 672/672 at
    `.cartulary/test-results/20260831T114247Z-p58908`.
20. Post-check retained-run finalization passed 1/1 at
    `.cartulary/test-results/20260831T114847Z-p10573`, using the successful
    full-check root from item 19.

Failure classification and disposition:

- Related: staticcheck found `columnName`, made unused by the WS-23 dead-method
  deletion. The helper was deleted without an alias; `make lint-go` then
  passed. The full unchanged-product rerun confirms this unit is clean.
- Environmental: the first full check's recovery integration unit received an
  object-store `PutObject` unavailable error. It passed in the next full-check
  run with no recovery-code change.
- Initially unrelated, then explicitly authorized: Biome reported eight
  warnings in the previously untouched grid-adapter CSS. Direct
  `make lint-biome` reproduced them at
  `.cartulary/test-results/20260831T040049Z-p48406`. The structural cascade
  correction removed the warnings without a suppression; focused unit and
  service-backed grid-adapter behavior checks and the final full check pass.

The second full check reached 671/672 at
`.cartulary/test-results/20260831T040226Z-p49526`; its only failure was the same
Biome unit. After explicit scope authorization and the CSS repair, the final
full check passes all 672 units.

Compatibility and migration: the Imports remediation remains repository-
internal or specification/test/harness-only. Public HTTP operations and
schemas, database shape, durable identifiers and bytes, Jobs identities,
frontend interfaces, and source-/analytical-owner resource behavior are
unchanged. The authorized frontend change removes exceptional CSS priority
while preserving the scoped grid-cell density declarations and their effective
precedence over the vendor layer. No migration, rollout flag, fallback,
deprecated alias, or second composition path exists.

Changed-file summary: two owner specification files; the controlling tracker;
Imports production, contract tests, and three terminology-only test renames;
server/extension assembly; Network Flow facade/port composition; import-target
generator inputs/tests and generated Go projections; the authored backend
boundary policy/checker; Imports verification/test-family inputs; and the
generated topology render index; and the authorized grid-adapter stylesheet
cascade cleanup. No database, SQL, OpenAPI, dependency lock, or migration file
changed.

Rollback boundary: WS-20 authority corrections; WS-21 exact source-port and
Network Flow cutover; WS-22 module/facade/generator/server cutover; WS-23
privacy/dead-code/export guard; WS-24 boundary/harness replacement; WS-25
terminology/selector replacement; and the single grid-adapter declaration-
ordering/priority cleanup are individually atomic boundaries. No known
residual correctness or compatibility risk remains: focused grid-adapter
behavior, browser-backed remediation coverage, the full repository check, and
retained-run finalization all pass. The earlier blocked-state Markdown
checkpoint passed at `.cartulary/test-results/20260831T040906Z-p74816`; the
pre-closure final-tree Markdown checkpoint passed at
`.cartulary/test-results/20260831T115035Z-p13979`, and the completed tracker
passed at `.cartulary/test-results/20260831T115158Z-p15343`.

### 4.1 Mandatory checkpoint

At the end of each implementation workstream:

1. Mark the row `DONE` or `BLOCKED`; do not begin its dependent while blocked.
2. Update the findings, top-level tracker, completion criteria, and handoff log.
3. Record changed files, substantive changes, compatibility and migration
   outcome, commands and run roots, failure relation, rollback, and residual risk.
4. Run `make lint-markdown` and record the result.

Rollback is by complete workstream. A failing removal is rolled back with its
consumer cutover; it is never repaired with a legacy alias, a second injection
path, or a duplicate mapping/finalization implementation.

## 5. Planned Internal Interface Changes

| Current internal seam | Planned replacement | Compatibility posture |
| --- | --- | --- |
| Variadic route options and six `With*` functions | `NewModule(ModuleDependencies)`, `Module.RegisterRoutes`, and `Module.RegisterWorkers` implemented in WS-22 | Old constructors removed without wrappers |
| Imports `ModuleOverrides` entry plus string facade keys | Self-describing typed analytical facade collection implemented in WS-22 | Untyped key, decoder, and merge paths are absent |
| Facade interface without binding identity | Complete `ExtensionImportFacadeBinding` validated against generated target facts | Startup rejects all identity or schema drift |
| General `NewStore` previously passed from server assembly to Network Flow | `NewExtensionSourcePort` and exact two-method `ExtensionSourcePort` implemented in WS-21 | General store is no longer exposed across this owner edge |
| Unclassified root and `ownerfacade` exports | Exact AST role inventory implemented in WS-23 | New, stale, empty, or duplicate roles fail |

These are breaking changes only to repository-internal Go composition. No
external route, OpenAPI schema, database schema, frontend interface, generated
product type, or owner behavior change is planned.

## 6. Validation Plan

Before selecting rows, use the public task guide and owner explanation for
`module.imports`, `app.server`, and `module.networkflow`.

| Layer | Command | Required coverage |
| --- | --- | --- |
| Imports unit and contract | `make test-slice OWNER=module.imports` | Composition failures, exports, route catalog, mapping and owner contracts |
| Imports services | `make service-backed-test-slice OWNER=module.imports` | Persistence, replay, apply, recovery, owner dispatch, cancellation |
| Server assembly | `make test-slice OWNER=app.server` | Typed composition and production route/worker wiring |
| Network Flow unit | `make test-slice OWNER=module.networkflow` | Facade binding and source-port compatibility |
| Network Flow services | `make service-backed-test-slice OWNER=module.networkflow` | Mapping/apply source validation and atomic publication |
| Backend boundaries | `make backend-module-boundary-check` | Exact transport set, peer-import ban, aggregate prohibition |
| Harness contracts | `make harness-contract` and `make json-shape-check` | Authored manifest and schema consistency |
| Harness generation | `make generate`, then `make generate-drift` | Generated topology refresh and equivalence |
| Generated policy | `make generated-artifact-policy-check` | No hand-edited generated root |
| Browser workflow | `make browser-e2e-webserver-backed` | Claimed and unclaimed assistant workflow |
| Broad local | `make test-fast` | Cross-owner regression coverage |
| Finalization | `make agent-finalize` | End-of-run policy and retained-evidence maintenance |
| Repository | `make check` | Full repository validation |
| Tracker | `make lint-markdown` | Controlling handoff validity |

Required construction cases are: each missing individual dependency; each
single missing member of the Jobs bundle; zero or invalid limits; nil, unknown,
duplicate, facade-ID-mismatched, schema-mismatched, and target-mismatched facade
bindings; claimed target without a facade; unclaimed target without a facade;
and valid registration of exactly 11 routes and two workers.

Required behavior preservation includes route and error bytes, cookie-CSRF
precedence, job identities, mapping and replay fingerprints, source provenance,
owner transaction atomicity, recovery contributions, Network Flow import
binding behavior, and the claimed/unclaimed browser workflow.

## 7. Top-Level Tracker

| Workstream | Status | Depends on | Completion evidence |
| --- | --- | --- | --- |
| WS-19 | DONE | WS-18 | Tracker-only rewrite; Markdown passed at `.cartulary/test-results/20260831T012821Z-p32338` |
| WS-20 | DONE | WS-19 | Owner-specification and tracker reconciliation; Markdown/shape/drift passed |
| WS-21 | DONE | WS-20 | Exact source-port unit/service evidence and Markdown checkpoint |
| WS-22 | DONE | WS-21 | Typed module composition/facade validation and Markdown checkpoint |
| WS-23 | DONE | WS-22 | Dead-code deletion, exact export-role evidence, and Markdown checkpoint |
| WS-24 | DONE | WS-23 | Durable boundary, behavior-test, harness, generation, and Markdown evidence |
| WS-25 | DONE | WS-24 | Contract vocabulary, selector, harness, generation, and Markdown evidence |
| WS-26 | DONE | WS-25 | Authorized CSS repair, 672/672 final check, retained-run finalization, and Markdown checkpoint passed |

No production, owner, contract, harness, generated, database, or frontend work
was performed as part of WS-19.

## 8. Binary Completion Criteria

| Criterion | Current status |
| --- | --- |
| Previous WS-00 through WS-18 closure and final run roots remain visible | PASS |
| Baseline and authority are accurate; current 38-file inventory accounts for the WS-24 historical-test deletions | PASS |
| WS-20 through WS-26 have dependencies, exact outcomes, stop rules, tests, and rollback posture | PASS |
| WS-19 changes only this tracker and passes `make lint-markdown` | PASS |
| Owner specification and analytical binding agree on Network Flow major 5 | PASS |
| Network Flow receives only the exact Imports source capability | PASS |
| Imports has no untyped analytical-facade override path | PASS |
| Incomplete production composition fails during construction | PASS |
| Dead functions and broad store exposure are absent | PASS |
| Every retained export has a live production role | PASS |
| Historical file/symbol assertions are absent | PASS |
| Characterization names and selectors are absent | PASS |
| Durable boundary, compatibility, harness, and generated evidence pass | PASS |
| All narrow, browser, fast, finalization, and full checks pass with recorded roots | PASS |
| No public wire, durable-state, owner, database, or frontend compatibility change occurred | PASS |

## 9. Compact Handoff Log

| Time | Workstream | Result | Changes | Verification | Next action |
| --- | --- | --- | --- | --- | --- |
| `2026-08-30T14:20:54-04:00` | WS-18 | Previous remediation closed RB-001 through RB-006. | Final implementation and tracker closure. | Imports 14/14, browser 60/60, fast 441/441, finalization passed, repository 671/671; roots retained in Section 2.1. | Rebaseline only under a new authorized task. |
| `2026-08-30T21:22:58-04:00` | WS-19 | Compact production-readiness plan installed and WS-19 closed; implementation remained unstarted. | This tracker only. | `make lint-markdown` passed at `.cartulary/test-results/20260831T012821Z-p32338`. | WS-20 authority reconciliation. |
| `2026-08-30T21:50:00-04:00` | WS-20 | DONE — owner authority and controlling sequence reconciled. | NF-REQ-088d major 5; AC-465 generated-target/contract-test language; tracker WS-20 through WS-26. | Markdown, JSON shape, and generated drift passed; roots recorded in WS-20 closure. | Begin WS-21. |
| `2026-08-30T22:06:00-04:00` | WS-21 | DONE — exact source-port cutover and checkpoint passed. | Two-method Imports source port; bounded filename provenance; direct Network Flow transaction capability; no server general-store edge. | Imports unit 23/23 and service-backed 14/14; Network Flow unit 34/34 and service-backed 28/28; Markdown passed. One related compile failure was corrected; roots are recorded in WS-21 closure. | Begin WS-22. |
| `2026-08-30T22:18:00-04:00` | WS-22 | DONE — typed composition, analytical injection, and checkpoint passed. | Fail-closed module; separate 11-route/two-worker registration; full facade bindings; generated schema/major facts; no Imports overrides. | All Imports, app-server, and Network Flow unit/service slices, generation drift, JSON shape, and Markdown passed. Two related intermediate test failures were corrected; roots are recorded in WS-22 closure. | Begin WS-23. |
| `2026-08-30T22:50:00-04:00` | WS-23 | DONE — dead-code deletion, privacy cutover, export-role guard, and checkpoint passed. | Deleted dead declarations; privatized implementation-only surface; installed exact root/`ownerfacade` role guard. | Imports unit/service, app-server, Network Flow retry, formatting, and Markdown pass; roots and the unrelated object-store timeout are recorded in WS-23 closure. | Begin WS-24. |
| `2026-08-30T23:21:00-04:00` | WS-24 | DONE — durable boundary/test replacement and checkpoint passed. | Authored peer-owner rule and fixtures; architecture verification; composition/export rows; historical source tests deleted. | Imports, boundary, harness, JSON shape, generation, generated policy, drift, and Markdown pass; roots and one corrected row-order failure are recorded in WS-24 closure. | Begin WS-25. |
| `2026-08-30T23:30:00-04:00` | WS-25 | DONE — contract vocabulary/selector normalization and checkpoint passed. | Three test-file renames; contract/compatibility helpers and tests; behavior-contract row; generated render index. | Imports, format, harness, JSON shape, generation, generated policy, drift, and Markdown pass; roots and one corrected selector-order failure are recorded in WS-25 closure. | Begin WS-26. |
| `2026-08-31T07:49:20-04:00` | WS-26 | DONE — final validation and handoff complete. | Removed one related dead helper; repaired the authorized CSS cascade structurally; recorded final compatibility, rollback, and residual-risk disposition. | Biome 2/2, grid-adapter 38/38 and 13/13, repository check 672/672, retained-run finalization 1/1, and pre-closure Markdown passed; roots are recorded in WS-26. | None; WS-20 through WS-26 and the overall remediation are closed. |

WS-20 through WS-26 are complete. Reopen only under a new authorized change;
the atomic rollback boundaries and final successful run roots are recorded
above.
