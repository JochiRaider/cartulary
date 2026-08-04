# view-contracts Module Refactoring Tracker and Handoff

## Current Control — Production-Readiness Iteration

This section controls the authorized S-09 through S-14 production-readiness
iteration. Sections 1 through 12 retain the completed first-iteration analysis
until the S-09 rebaseline replaces them; section 13 is immutable historical
evidence for S-00 through S-08.

| Item | Value |
| --- | --- |
| Status | `DONE` 2026-08-04; S-09 through S-14 complete and handed off |
| Baseline | Clean worktree at `25cd6532dfd9a6531ac7fffb10b6d051478225b2` |
| Target | Replace browser-time parsing of trusted view-schema artifacts with a generator-owned normalized projection, then remove obsolete protocol, package, web, test, and harness seams |
| Compatibility boundary | Private workspace packages at `0.0.0`; repository-wide importer scans and frontend typecheck; no aliases or fallback APIs |
| Preserved identities | All current view-schema IDs, `cartulary.inspector_config.v1`, HTTP routes, row wire shapes, and the `@cartulary/view-contracts` `.` entrypoint |
| Generated-file rule | Authored generator and policy inputs change first; generated roots and topology outputs change only through Make |

After each slice, and before beginning the next slice, this tracker MUST record
the status, substantive authored and generated files, validation commands and run
roots, compatibility impact, rollback posture, blocker state, and next slice.
`make lint-markdown` and `git diff --check` MUST pass before a slice is `DONE`.

| Slice | State | Depends on | Exit summary |
| --- | --- | --- | --- |
| S-09 | `DONE` | S-08 | Generator-owned normalized projection, policy ownership, and parity evidence pass without a production cutover |
| S-10 | `DONE` | S-09 | Package initialization consumes the projection and contains no source parser or Inspector parser |
| S-11 | `DONE` | S-10 | Raw TypeScript artifacts, source decoder/types, runtime Inspector registry value, and generic artifact-collection machinery are removed |
| S-12 | `DONE` | S-11 | Dead package and web APIs are removed without aliases and all consumers migrate to the immutable list API |
| S-13 | `DONE` | S-12 | Focused tests and authored harness accounting replace the monolith; obsolete sentinels and seams are absent |
| S-14 | `DONE` | S-13 | Exact focused-to-broad validation passes and the implementation handoff is complete |

### Current work items

| Work item | Slice | State | Binary outcome |
| --- | --- | --- | --- |
| VC-014 | S-09 | `DONE` | A generator-owned normalized browser projection has exact parity with all 17 contracts and surfaces |
| VC-015 | S-10 | `DONE` | Production package initialization has no raw-source parser, decoder, or duplicate Inspector validator |
| VC-016 | S-11 | `DONE` | Protocol and contractgen expose no raw view artifact, TypeScript source type, source decoder, runtime Inspector registry value, or artifact-collection facility |
| VC-017 | S-12 | `DONE` | Dead façade and workbook construction seams are absent without aliases |
| VC-018 | S-13 | `DONE` | Package tests are focused, owner-routable, and free of parser/builder and sentinel residue |
| VC-019 | S-14 | `DONE` | The prescribed focused-to-broad validation sequence and production handoff are complete |

### Current blockers and completion criteria

No blocker is open. A slice is binary-complete only when its implementation,
importer audit, focused validation, generated-policy validation where applicable,
tracker checkpoint, `make lint-markdown`, and `git diff --check` all pass. The
overall iteration is complete only when VC-014 through VC-019 are `DONE`, every
S-14 command passes in order, removed symbols and outputs are absent, generated
drift is empty, and the final compatibility and rollback handoff is recorded.

### Current handoff log

| Date | Slice | Event | Next action |
| --- | --- | --- | --- |
| 2026-08-04 | S-09 | Generated projection and policy ownership completed | Cut package runtime over to the projection |
| 2026-08-04 | S-10 | Browser source parsing and Inspector validation removed | Remove the now-unreachable protocol and generator facilities |
| 2026-08-04 | S-11 | Registry v5 and protocol/generator cleanup completed | Remove dead façade and workbook construction seams in S-12 |
| 2026-08-04 | S-12 | Dead package, workbook, and E2E construction/re-export seams removed | Split and account the focused package test architecture in S-13 |
| 2026-08-04 | S-13 | Five-module production graph and four owner-routable test files completed | Run the exact S-14 validation and handoff sequence |
| 2026-08-04 | S-14 | Corrected one dead staticcheck helper, restarted the sequence, and passed every focused and broad gate | Implementation is ready for review and commit |

The selected target and removals are the decision-complete plan in the task
authorization. In particular, the prior decisions to retain `.gitkeep`, runtime
source parsing, injected surface builders, capability/visibility helpers, and
test-only optional-surface omission are superseded. Optional surface status
remains owner metadata; absence of a declared build artifact is a generation
failure unless a future adopted owner introduces an explicit availability model.

## 1. Scope, Authority, and Status

This tracker controls the S-09 through S-14 production-readiness iteration for
`@cartulary/view-contracts`. Adopted Core owners and authored contracts remain
the behavioral authority. This tracker controls sequencing, compatibility
decisions, evidence, rollback, and handoff; no runtime or test reads it.

The iteration began from clean commit
`25cd6532dfd9a6531ac7fffb10b6d051478225b2`. No Core owner, domain vocabulary,
database, HTTP route, view identity, or Inspector identity change is in scope.
Generated files and harness topology MUST change only through Make-owned
generation.

The tracker status is `DONE`. S-09 through S-14 are complete. Section 13 is the
preserved S-00 through S-08
historical execution record. Section 14 records this iteration's checkpoints.

## 2. Rebaseline and Repository Evidence

| Evidence | Baseline conclusion |
| --- | --- |
| Package importers | 83 TypeScript files referenced `@cartulary/view-contracts` at rebaseline; repository typecheck is the private-workspace compatibility boundary |
| Removable source path | Raw artifacts, source types, decoder/parser code, and duplicate Inspector validation totaled approximately 453 KB before S-09 |
| Package identity | Private workspace package at `0.0.0`, one public `.` entrypoint, no supported subpaths |
| Runtime trust boundary | Authored view-schema sources are trusted generation inputs; no browser API accepts untrusted source documents |
| Surface corpus | All 17 declared contracts and surfaces are required build outputs; optional status is metadata, not a silent omission profile |
| Baseline checks | Package, protocol, typecheck, import-boundary, generation-drift, JSON-shape, and Markdown checks passed before implementation |

The production package previously reparsed escaped JSON, decoded already
validated authored sources, repeated cross-field Inspector validation, built
optional subsets through injected tests, and exposed helpers with no production
consumer. Those behaviors are compatibility burden rather than product value.

## 3. Production Target Architecture

The final package MUST contain five authored production modules plus focused
tests:

1. `index.ts` — the only public façade.
2. `types.ts` — the coherent View, Inspector, row, and surface metadata graph.
3. `rows.ts` — full and sparse branded row normalization.
4. `projection.ts` — the sole adapter over the package-local generated browser projection.
5. `contracts.ts` — immutable contract/surface lookup, schema constants, and status partitions.

The generated `view-contract-projection.ts` contains normalized, registry-ordered,
deeply frozen browser values. It MUST NOT contain raw JSON strings, hashes,
defaults, fallback validation, or package imports. `projection.ts` alone may
import it and derives maps while preserving shared contract identity between
contract and surface lists.

The package MUST NOT own source-schema parsing, runtime source decoding,
authorization, routes, mutation execution, workbook state, feature availability,
or a generic plugin system.

## 4. Interface and Compatibility Decisions

Retained public runtime APIs are row normalization, contract list/get/require,
header-sort resolution, surface listing, 17 schema constants, and three status
arrays. Retained public types are the branded full/sparse row results and the
coherent View, Inspector, field, row-cell, and surface metadata graph.

The following are intentionally removed without aliases:

- `fieldCapability`, `visibleFields`, and `ViewFieldCapability`;
- `buildWorkbookSurfaceContracts`, `getWorkbookSurfaceContract`, and
  `requireWorkbookSurfaceContract`;
- web `buildWorkbookSurfaceRegistry` and its injectable impossible-state tests;
- the E2E workbook-surface re-export module;
- protocol `viewSchemaArtifacts`, `viewSchemaSourceDocumentDecoder`, every
  `ViewSchemaSource*` TypeScript type, and runtime `viewInspectorRegistry`;
- generated `artifact.ts`, raw view-schema artifacts, source types, and source decoder;
- contractgen artifact-collection and shared TypeScript artifact support.

`viewSchemaRegistry`, `listViewSchemaRegistryEntries`, literal-derived Inspector
types, and Go view-schema source types remain. No deprecation alias, fallback,
package subpath, identity bump, database migration, or route migration is
introduced.

Prior S-00 through S-08 decisions to retain `.gitkeep`, runtime parsing,
injected surface builders, capability/visibility helpers, and optional-surface
omission through test injection are superseded.

## 5. Gap Remediation Matrix

| Gap | Structural remediation | Areas | Long-term benefit | Risk if unresolved | Completion evidence |
| --- | --- | --- | --- | --- | --- |
| Duplicate browser source validation | Generate one normalized frozen browser projection and adapt it locally | Generator, policy, package, tests | Single validation authority and deterministic startup | Drift and browser-only initialization failures | 17-contract parity, shared identity, freeze, drift, and package evidence |
| Raw artifact projection facility | Remove artifact collections, source writer/decoder, and adopt registry v5 | Contracts, generator, protocol, tests | Smaller generated surface and no raw-source coupling path | Future features may ship owner sources into browsers | v5 shape, generation, obsolete-output absence, protocol exact surface |
| Dead façade/build seams | Remove helpers/builders and migrate workbook/E2E callers to immutable lists | Package, web, E2E, tests | Fewer impossible states and supported APIs | Dead behavior constrains future structure | Importer scans, export test, typecheck, browser behavior |
| Browser-owned invalid-source tests | Move rejection evidence to contractgen; test browser values and public behavior only | Generator and package tests | Tests follow the real trust boundary | Parser architecture survives only to support tests | Generator rejection matrix and no parser test seam |
| Monolithic tests and sentinel | Split façade, row, contract, and surface tests; remove `.gitkeep` | Tests, harness, package | Focused failures and accountable growth | Owner slices can omit or obscure evidence | Four owner-routable paths, topology drift-clean, no sentinel |

## 6. Sequential Workstreams

| Slice | Scope | Dependency | Exit condition |
| --- | --- | --- | --- |
| S-09 | Generate normalized browser projection, adapter, ownership policy, and parity tests | S-08 | Projection is deterministic and unused by production until parity passes |
| S-10 | Cut package initialization to the projection and delete source/Inspector parsers | S-09 | Production graph contains no parser, source decoder, or duplicate Inspector validation |
| S-11 | Remove protocol/generator legacy and adopt contract-family registry v5 | S-10 | Obsolete outputs and artifact-collection machinery are absent |
| S-12 | Remove dead façade, workbook, and E2E seams | S-11 | All importers use retained immutable APIs and browser behavior passes |
| S-13 | Consolidate five-module production graph, split tests, and finalize accounting | S-12 | Every retained test is owner-routable and cleanup scans are empty |
| S-14 | Run prescribed focused-to-broad validation and complete handoff | S-13 | Every command passes in order and VC-014 through VC-019 are `DONE` |

Each slice is independently reversible within its stated compatibility unit.
The tracker MUST be updated after a slice and before the next begins. A failed
slice is `BLOCKED`; later slices MUST NOT start.

## 7. Dependencies and Sequencing Rules

- S-09 establishes parity before any runtime switch.
- S-10 removes browser parsing only while S-11 protocol rollback remains available.
- S-11 removes the old protocol/generator unit only after the package graph no longer imports it.
- S-12 removes public and web seams only after a final importer audit.
- S-13 moves tests and topology only after runtime/public structure is stable.
- S-14 is validation and handoff only; first-time implementation or test registration is prohibited.

After every slice, update the current work table, blocker state, handoff log,
checkpoint, generated/authored distinction, compatibility impact, rollback, and
next slice. Run `make lint-markdown` and `git diff --check`; mark `DONE` only
after both pass.

## 8. Validation Strategy

Narrow evidence is selected by owner before broad checks. Generator or harness
input changes require generation through Make and drift checks. The minimum
focused owners are `platform.viewschema`, `package.protocol_ts`,
`package.view_contracts`, `web.architecture`, `web.workbook`, and the affected
`module.workbook` collaborator row.

Validation covers:

- exact public runtime and type surfaces;
- all 17 contract/surface values, order, freezing, and shared identity;
- full/sparse row behavior and type non-assignability;
- generator-time malformed source and cross-field rejection;
- generated ownership, JSON shape, drift, and frontend boundaries;
- package, workbook, browser, typecheck, Biome, and web build behavior;
- authored title/path ownership and generated topology.

S-14 runs the user-prescribed command sequence in order and stops at the first
failure. `make agent-finalize` runs without `RESULTS_DIR` unless an equivalent
successful retained warm-check root exists.

## 9. Top-Level Work Tracker

| Work item | Slice | State | Exit evidence |
| --- | --- | --- | --- |
| VC-014 | S-09 | `DONE` | Generated projection parity and policy ownership |
| VC-015 | S-10 | `DONE` | Runtime projection cutover and parser deletion |
| VC-016 | S-11 | `DONE` | Registry v5 and protocol/generator legacy removal |
| VC-017 | S-12 | `DONE` | Dead façade/web/E2E seams removed and importers migrated |
| VC-018 | S-13 | `DONE` | Five-module graph, four focused test files, topology and cleanup audit |
| VC-019 | S-14 | `DONE` | Complete validation and handoff |

## 10. Risk and Rollback Register

| Risk | Prevention | Rollback unit |
| --- | --- | --- |
| Generated values differ from characterized runtime | Full semantic parity before S-10 | S-09 projection, adapter, and policy additions |
| Cutover changes order, identity, or freezing | Exact object identity/order/freeze tests | S-10 package parser/projection unit |
| Generator v5 removes a still-used output | Repository import scan plus protocol/typecheck gates | S-11 registry v4, writers, outputs, and entrypoint together |
| Dead API removal breaks a hidden consumer | Final repository scan and private-workspace typecheck | S-12 façade, web tests, and E2E import migration |
| Test split loses evidence | Same-slice manifests, Make generation, focused owner execution | S-13 tests and authored topology inputs |
| Broad unrelated failure obscures handoff | Stop on first failure and classify relatedness from run artifacts | No code rollback until failure is diagnosed |

No database, owner, identity, route, or wire rollback is required by this
iteration.

## 11. Blockers, Assumptions, and Open Questions

No blocker or open design question remains. The production cutover was selected
by the task owner. Authored schemas and registries are trusted build inputs.
Optional surface status remains metadata, but silent artifact absence is not a
supported build profile. Repository-wide typecheck and importer scans are the
compatibility boundary for private `0.0.0` TypeScript packages.

If a future phase needs feature-gated surface availability, it MUST introduce an
explicit owner model rather than restore injected builders or omission fallback.

## 12. Binary Completion Criteria

The iteration is `DONE` only when all statements are true:

- VC-014 through VC-019 and S-09 through S-14 are `DONE`.
- The package has one `.` entrypoint, five authored production modules, four
  focused test files, and no `.gitkeep`.
- Every retained API/type remains and every selected removal is absent without an alias.
- No production browser source imports raw artifacts, source types/decoder,
  runtime Inspector registry values, or parses view-schema JSON.
- Contractgen supports no artifact collection or shared raw TypeScript artifact output.
- Contract-family registry v5, generated policy, JSON shapes, and generation drift pass.
- All 17 contracts and surfaces preserve value, registry/status order, freezing,
  shared identity, schema constants, and Inspector metadata.
- Every focused test path/title is owned and generated topology is current.
- Every S-14 command passes in order, with exact run roots recorded.
- The final checkpoint records generated-file confirmation, API removals,
  compatibility impact, source-size reduction, remaining risks, rollback, and handoff.

## 13. Execution Checkpoints

### 13.1 S-00 — Owner, projection, and Indicator behavior repair

| Item | Result |
| --- | --- |
| Status | `DONE` on 2026-08-04; completed as one owner/projection/application compatibility unit |
| Dependencies | Implementation authority granted; baseline and compatibility posture recorded |
| Authored specification and contract changes | Core 01 exact vocabularies/rows/precedence/exclusion; Core 03/Core 04 references and AC-532; Appendices E/F; `view-inspector` registry family and JSON Schema; Timeline/Indicators feature objects; eight protocol operation catalog entries; protocol entrypoint/import-boundary ownership |
| Generated projections | Regenerated only through `make generate`: Go contract registry, view-schema/extension artifacts, protocol HTTP request/response/validator/binding types, Inspector registry TypeScript, and execution-topology render index |
| Implementation changes | `view-contracts` consumes the owner-derived registry; workbook Indicator operations use `WorkbookOperationExecutor`; exact complete-tuple dispatch mounts observations in `relationships` and lifecycle in `history`; paging, mutation metadata, retry/loading/empty states, stable-ID refresh, and unsupported-tuple omission are explicit |
| Test/accounting changes | Added exact transport/dispatch/paging tests and a dedicated accessible UI/paging/metadata row; `module.indicators` now has 24 rows, two frontend rows, and 11 service-backed rows; topology regenerated through Make |
| Compatibility/migration | Preserved `cartulary.view.timeline.v2`, `cartulary.view.indicators.v1`, `cartulary.inspector_config.v1`, HTTP paths, server authorization, and package `.`. No database/storage migration, identity bump, alias, permissive fallback, or generic patch compatibility was added. Panel placement and unsupported inert-control omission are intentional UI changes. |
| Rollback | Revert the complete S-00 authored owner/contract/generator/package/web/test/harness/golden change set together, restore the prior authored registry inputs, and rerun `make generate`; do not selectively retain generated outputs. |
| Next slice | S-01 characterization baseline; do not begin until the checkpoint Markdown and diff checks pass. |

#### S-00 passing evidence

| Command | Result | Run root or evidence |
| --- | --- | --- |
| `make generate` | PASS | `.cartulary/test-results/20260804T145847Z-p779607` |
| `make lint-markdown` | PASS after the completed checkpoint write | `.cartulary/test-results/20260804T150535Z-p800915` |
| `make lint-biome` | PASS | `.cartulary/test-results/20260804T150425Z-p796340` |
| `make generate-drift` | PASS | `.cartulary/test-results/20260804T150436Z-p796751` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260804T150446Z-p799320` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260804T150451Z-p799757` |
| `make test-slice OWNER=platform.viewschema` | PASS, 3/3 units | `.cartulary/test-results/20260804T144139Z-p621256` |
| `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 units | `.cartulary/test-results/20260804T144916Z-p663365` |
| `make test-slice OWNER=package.view_contracts` | PASS, 5/5 units | `.cartulary/test-results/20260804T144927Z-p668379` |
| `make test-slice OWNER=module.indicators` | PASS, 27/27 units | `.cartulary/test-results/20260804T145949Z-p789179` |
| `make service-backed-test-slice OWNER=module.indicators` | PASS, 13/13 units | `.cartulary/test-results/20260804T144956Z-p672848` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 units | `.cartulary/test-results/20260804T145011Z-p674298` |
| `make service-backed-test-slice OWNER=module.workbook` | PASS, 66/66 units | `.cartulary/test-results/20260804T145049Z-p687286` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T150006Z-p792964` |
| `make frontend-import-boundary-check` | PASS, 2/2 units | `.cartulary/test-results/20260804T150501Z-p800288` |
| `make browser-e2e-webserver-backed` | PASS, 62/62 units | `.cartulary/test-results/20260804T145253Z-p723486` |
| `make explain-test-owner OWNER=module.indicators` | PASS | 24 rows; frontend 2; service-backed 11 |
| `make target-plan` | PASS | 614 units; digest `sha256:0fb8b2a04bf627a25e4b61b2c5dc7d20b28a1f7b794ebc9aa33788e7efc49be5` |
| `git diff --check` | PASS | No output after the S-00 tracker checkpoint |

#### S-00 failures encountered and resolved

| Run root | Failure and disposition |
| --- | --- |
| `.cartulary/test-results/20260804T140454Z-p404667` | Initial generation exposed an accidentally duplicated panel; authored schema corrected, then regeneration passed. |
| `.cartulary/test-results/20260804T141738Z-p412825` | Formatter exposed an implicit-any branch and forbidden test assertions; source/test typing corrected. |
| `.cartulary/test-results/20260804T141911Z-p425638`, `.cartulary/test-results/20260804T141935Z-p426262` | JSON-shape validation exposed missing active-family accounting and then stale generated topology; authored checker updated and generation rerun. |
| `.cartulary/test-results/20260804T142022Z-p435138` | Package tests exposed deterministic registry-order diagnostic changes; assertions aligned with the owner-derived ordered registry. |
| `.cartulary/test-results/20260804T142305Z-p455919`, `.cartulary/test-results/20260804T142611Z-p467140`, `.cartulary/test-results/20260804T142755Z-p475106` | Owner-scoped service validation exposed missing selection provenance and illegal direct Playwright row scheduling; the harness now forwards selection identity and projects Playwright rows through filtered browser groups. |
| `.cartulary/test-results/20260804T143206Z-p514340` | Functional browser rows passed; the intentional Indicator panel move produced reviewed visual deltas. Canonical refresh and later comparison validation passed. |
| `.cartulary/test-results/20260804T144323Z-p645189`, `.cartulary/test-results/20260804T144442Z-p649530`, `.cartulary/test-results/20260804T144756Z-p656203` | Type/import gates exposed an over-wide parsed-type field, an untyped spy, protocol imports outside adapters, a missing generated-entrypoint allowlist, and a panel narrowing gap; each was structurally corrected. |
| `.cartulary/test-results/20260804T145859Z-p781853` | The new UI row used unavailable DOM matcher extensions; assertions were rewritten against standard DOM properties and the owner slice passed. |

All failures were related to S-00 implementation or its required owner-scoped validation path and are resolved. No S-00 validation was skipped. Broad `agent-finalize`, `test-fast`, and `check` remain deliberately sequenced in S-08, not claimed here.

#### S-00 visual refresh record

- Accepted trigger: adopted Inspector behavior intentionally moved Indicator observation management from `workflow` to `relationships` and omitted unsupported inert relationship controls.
- Affected owner rows: `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` and `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`.
- Affected fixtures: `visual.fixture.base_inspector`, `visual.fixture.destructive_actions`, and `visual.fixture.mention_chip_state_matrix`.
- Changed goldens: `entity-mention-chip-states-linux.png`, `workbook-inspector-destructive-confirmation-linux.png`, `workbook-inspector-history-linux.png`, `workbook-inspector-public-error-linux.png`, `workbook-inspector-relationships-linux.png`, and `workbook-inspector-rollback-preview-linux.png` under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
- Capture contract: viewport, browser zoom, masks, scroll normalization, screenshot scope, fixture identities, typography, grid layout, and focus contract did not change.
- Refresh: `make browser-e2e-visual-update` passed at `.cartulary/test-results/20260804T143614Z-p554985`.
- Review: all six refreshed images were inspected; deltas were limited to the supported Indicator control and the resulting downstream Inspector vertical layout.
- Ordinary comparison validation: workbook owner service-backed comparison passed at `.cartulary/test-results/20260804T145049Z-p687286`; standalone webserver-backed browser validation passed at `.cartulary/test-results/20260804T145253Z-p723486`.

### 13.2 S-01 — Characterization baseline

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-00 complete; owner-derived Indicator semantics already active |
| Substantive authored files | `packages/view-contracts/src/index.test.ts`; `tools/test_families/package.view_contracts.json`; this tracker |
| Generated files | `tools/execution_topology_render_index.json`, regenerated by `make generate`; no generated file was hand-edited |
| Test accounting | Added `package.view_contracts.frontend_unit.characterization_baseline_31fd4bbd8a` with 11 exact titles under the existing shared-contract verification owner; the package manifest now has 5 active rows |
| Coverage added | Supported facade availability; 17-artifact initialization; freeze boundaries; invalid JSON and malformed roots; missing/malformed row identity and version; full-cell completeness; technical-cell rejection; additive row/cell members; sparse patches; view mismatch; required/optional surface handling; registry mismatch and order |
| Intentional exclusions | No assertion accepts `row_version=0`; no assertion treats missing required view-source members as supported. Strict raw source-schema validation remains an intentional S-04 change. |
| Compatibility impact | Test and harness metadata only. Existing valid runtime behavior is unchanged; the baseline protects deliberate later structural moves without preserving approved malformed-input or zero-version cleanup targets. |
| Rollback | Revert the new test block and its authored family row, rerun `make generate`, and remove only the resulting topology digest change. S-00 remains independently intact. |
| Next slice | S-02 shared types and invariant infrastructure |

#### S-01 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units after topology regeneration | `.cartulary/test-results/20260804T151059Z-p809954` |
| `make generate` | PASS | `.cartulary/test-results/20260804T151021Z-p807514` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T151145Z-p810612` |
| `git diff --check` | PASS | No output after the S-01 checkpoint |

The first owner-slice run failed at `.cartulary/test-results/20260804T150937Z-p805317` because the export-availability assertion accidentally depended on JavaScript module-key enumeration order. The assertion was corrected to verify each supported function independently; this was a characterization-test defect, not a product behavior failure. No web implementation or web test changed in S-01, so no web owner slice was affected or skipped.

### 13.3 S-02 — Shared types and invariant infrastructure

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-01 characterization baseline complete |
| Substantive authored files | `packages/view-contracts/src/types.ts`; `packages/view-contracts/src/invariants.ts`; `packages/view-contracts/src/index.ts`; `packages/protocol-ts/src/entrypoints/view-schemas.ts`; the S-01 facade test import was made statically lintable; this tracker |
| Generated files | None; no generated root or harness output changed in S-02 |
| Structure | `types.ts` owns the coherent public metadata/row/surface type graph; `invariants.ts` owns source-aware contract/row failures and shared key/object/enum primitives; `index.ts` explicitly type-re-exports the public graph and consumes both modules |
| Dependency direction | Protocol registry literal types → `types.ts`; no-dependency invariant primitives → semantic code in `index.ts`; the two extracted modules do not import the façade or each other, so the internal graph is acyclic |
| Behavior | Existing exception envelopes, validation order, frozen runtime objects, module initialization, and public names remain unchanged; the historical stage-row comment was replaced by owner-neutral invariant wording |
| Compatibility impact | No runtime or public import-path change. Two additive type-only aliases (`ViewInspectorIncidentRole` and `ViewInspectorSpecializedActionKey`) expose facts already present in the protocol registry. |
| Test accounting | No title or path was added/renamed; the S-01 authored package row still executes the affected behavior, so topology regeneration was not required |
| Rollback | Restore public types and invariant helpers to `index.ts`, remove the two internal files and the two additive protocol aliases; no generated rollback is needed |
| Next slice | S-03 full and sparse row adapters |

#### S-02 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260804T151539Z-p815783` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T151547Z-p818938` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T151548Z-p819054` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T151548Z-p819091` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T151702Z-p821018` |
| `git diff --check` | PASS | No output after the S-02 checkpoint |

The first Biome run failed at `.cartulary/test-results/20260804T151507Z-p814794` because the newly added files needed formatting and the S-01 facade test used dynamic namespace access. `make format` normalized the authored sources and the test now statically imports every retained function. The concurrent typecheck already passed at `.cartulary/test-results/20260804T151507Z-p814764`; the complete post-fix gate set above also passes.

### 13.4 S-03 — Full and sparse row adapters

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-02 shared types and invariant infrastructure complete |
| Substantive authored files | `packages/view-contracts/src/rows.ts`; `packages/view-contracts/src/types.ts`; `packages/view-contracts/src/index.ts`; `packages/view-contracts/src/index.test.ts`; `tools/test_families/package.view_contracts.json`; this tracker |
| Generated files | `tools/execution_topology_render_index.json`, regenerated by `make generate`; no generated file was hand-edited |
| Type contract | `NormalizedViewRowV1` and new `NormalizedViewRowPatchV1` carry distinct required `unique symbol` phantom brands. The brands emit no runtime members and are non-assignable in either direction. |
| Runtime contract | Full rows still require every declared data cell; patches remain sparse; technical/unknown cells fail; additive row/cell members remain ignored; nested values, group semantics, source identity, and freeze boundaries are unchanged |
| Concurrency cleanup | Both adapters now require a positive safe integer and reject zero, negative, fractional, string, missing, infinite, and unsafe versions with one deterministic invariant |
| Consumer migration | All sparse consumers already use `ReturnType<typeof normalizeViewRowPatchV1>`, so they received the stronger type without local casts or aliases; repository typecheck and the full web owner slice confirm compatibility |
| Test accounting | Added the exact title `keeps normalized full rows and sparse patches non-assignable` to the existing characterization row; package manifest remains 5 rows and topology was regenerated through Make |
| Compatibility impact | Intentional compile-time distinction and intentional rejection of `row_version=0`; no valid wire-object shape or schema identity change |
| Rollback | Restore the S-02 row types/functions, remove `rows.ts` and the added test title, then rerun `make generate`; no other slice must be reverted |
| Next slice | S-04 view parser and helper extraction with strict generated decoding |

#### S-03 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260804T151951Z-p824989` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T152001Z-p828088` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 units | `.cartulary/test-results/20260804T152019Z-p829496` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T152001Z-p828160` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T152019Z-p829595` |
| `make generate` | PASS | `.cartulary/test-results/20260804T152057Z-p842883` |
| `make explain-test-owner OWNER=package.view_contracts` | PASS | 5 rows; 4 frontend-unit and 1 regression row; Vitest owner route confirmed |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T152206Z-p845654` |
| `git diff --check` | PASS | No output after the S-03 checkpoint |

No S-03 validation failed. No database, route, generated schema, or web implementation file changed. The complete `web.workbook` slice supplies the requested query and collaboration regression evidence rather than relying only on narrower local tests.

### 13.5 S-04 — View parser and helper extraction

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-01 characterization and S-02 shared foundations complete; S-03 remains independently intact |
| Substantive authored files | `tools/protocol-ts/generate-protocol-types.mjs`; `contracts/protocol-ts/frontend-entrypoints.v2.json`; `packages/protocol-ts/src/internal/decoder.ts`; `packages/protocol-ts/src/entrypoints/view-schemas.ts`; protocol tests; `packages/view-contracts/src/index.ts`; `packages/view-contracts/src/view-contracts.ts`; `packages/view-contracts/src/invariants.ts`; package tests and family accounting; this tracker |
| Generated files | Added `packages/protocol-ts/src/generated/view-schema-source-validator.ts`; refreshed generated protocol outputs and `tools/execution_topology_render_index.json` through `make generate`; no generated file was hand-edited |
| Decoder source | The existing authored `tools/schemas/cartulary.view_schema_source.v1.schema.json` is compiled by the protocol generator into a standalone AJV validator; the view-schema entrypoint exposes `viewSchemaSourceDocumentDecoder` typed as `ViewSchemaSourceDocument` |
| Validation order | JSON syntax → generated source-schema decoder → semantic and cross-document invariants. Parser defaults can no longer turn missing required members into accepted values. |
| Diagnostics | Syntax and shape failures use `View contract source validation failed: <source> path=<JSON pointer> reason=<category>`; required and additional-property decoder paths include the exact member while payload values remain absent |
| Extraction | The package `index.ts` is now a one-line façade over `view-contracts.ts`; source parsing, generated artifact initialization, lookup, sort, field-reference, capability, visibility, and still-to-be-extracted surface/Inspector semantics remain behind it without import cycles |
| Behavior evidence | All 17 generated artifacts initialize in registry order; semantic duplicate/reference/Inspector checks remain deterministic; valid field capabilities, ordering, visibility, mutation metadata, and freeze results are unchanged |
| Compatibility impact | Intentionally rejects invalid JSON, malformed roots, missing required members, wrong types, unknown source-schema members, and invalid closed values. Valid source artifacts and browser-facing runtime objects are unchanged. |
| Test accounting | Added one exact package title for source-schema shape branches to the existing characterization row and regenerated topology; protocol decoder evidence extends an already-accounted title/path |
| Rollback | Move the implementation back to `index.ts`, remove the source decoder export and generator output/input changes, restore prior decoder path behavior/tests, remove the added package title, and rerun `make generate`; row adapters remain untouched |
| Next slice | S-05 workbook surface assembly |

#### S-04 passing evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make generate` | PASS after final authored test accounting | `.cartulary/test-results/20260804T153039Z-p876424` |
| `make format` | PASS | `.cartulary/test-results/20260804T153006Z-p863541` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T153016Z-p866842` |
| `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 units | `.cartulary/test-results/20260804T153016Z-p866851` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T153016Z-p867093` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T153016Z-p867248` |
| `make generate-drift` | PASS, 4/4 units | `.cartulary/test-results/20260804T153051Z-p878657` |
| `make json-shape-check` | PASS, 3/3 units | `.cartulary/test-results/20260804T153051Z-p878659` |
| `make explain-test-owner OWNER=package.view_contracts` | PASS | 5 rows; package owner route confirmed |
| `make explain-test-owner OWNER=package.protocol_ts` | PASS | 6 rows; protocol owner route confirmed |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T153219Z-p882541` |
| `git diff --check` | PASS | No output after the S-04 checkpoint |

#### S-04 failures encountered and resolved

| Run root | Failure and disposition |
| --- | --- |
| `.cartulary/test-results/20260804T152436Z-p849488` | The first generation run rejected an unsorted authored generated-module allowlist. The new validator path was placed in lexical order; generation passed at `.cartulary/test-results/20260804T152500Z-p851069`. |
| `.cartulary/test-results/20260804T152648Z-p853855`, `.cartulary/test-results/20260804T152708Z-p854895`, `.cartulary/test-results/20260804T152820Z-p856142` | Strict decoding exposed intentionally incomplete test fixtures and assertions that expected later semantic errors. Fixtures were made schema-valid and shape assertions were updated to the new deterministic decoder boundary; the package slice then passed at `.cartulary/test-results/20260804T152918Z-p857338` and again in the final gate set. |
| `.cartulary/test-results/20260804T152928Z-p858250` | Exact-member decoder paths intentionally changed an existing Network Flow unknown-member expectation from the parent path to `/raw_source_value`. The payload-safe assertion was aligned and the protocol slice passed in the final gate set. |
| `.cartulary/test-results/20260804T153016Z-p866768`, `.cartulary/test-results/20260804T153016Z-p866828` | Drift and JSON-shape gates correctly detected the newly added package title after the preceding generation. A final `make generate` refreshed topology; both gates then passed. |

All failures were related to the S-04 implementation or its authored test-accounting order and are resolved. No database, backend route, view identity, Inspector identity, alias, fallback, or package subpath was introduced.

### 13.6 S-05 — Workbook surface assembly

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-02 shared foundations and S-04 parser module complete |
| Substantive authored files | `packages/view-contracts/src/workbook-surfaces.ts`; `packages/view-contracts/src/view-contracts.ts`; `packages/view-contracts/src/index.ts`; package tests and family accounting; this tracker |
| Generated files | `tools/execution_topology_render_index.json`, regenerated by `make generate`; no generated file was hand-edited |
| Extracted responsibility | Registry-order joining, injected source construction, required/optional handling, registry/contract agreement, frozen surface records, surface lookup index, 17 schema constants, and three status partitions |
| Dependency direction | `workbook-surfaces.ts` depends on the protocol registry, shared types, and `listViewContracts`; parser/view helpers do not depend on surface assembly, and the façade exports both without a cycle |
| Retained seam | `buildWorkbookSurfaceContracts(sourceContracts?)` remains public and injectable. Its default source is the initialized view-contract list, preserving the web surface registry’s production construction path. |
| Validation matrix | Missing required surface throws; missing optional surface omits; surface-kind and ordered reference-pack mismatches throw; registry order and all constants/partitions match; output arrays and entries remain frozen |
| Compatibility impact | None. Public names, package `.` entrypoint, identities, ordering, omission behavior, errors, and production injection are unchanged. |
| Test accounting | Added one exact constant/partition title to the existing characterization row and regenerated topology; no test path changed |
| Rollback | Restore the surface block to `view-contracts.ts`, remove `workbook-surfaces.ts` and its façade export, remove the added title, and rerun `make generate`; parser and row modules remain intact |
| Next slice | S-06 Inspector module extraction |

#### S-05 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260804T153610Z-p886743` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T153621Z-p889922` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T153622Z-p890050` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T153622Z-p890072` |
| `make generate` | PASS | `.cartulary/test-results/20260804T153640Z-p891657` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 units | `.cartulary/test-results/20260804T153653Z-p893895` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T153855Z-p907395` |
| `git diff --check` | PASS | No output after the S-05 checkpoint |

No S-05 validation failed or was skipped. No web implementation file needed migration because its existing package-level import and injected builder use were preserved exactly.

### 13.7 S-06 — Inspector module extraction

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-00 owner/projection repair, S-02 shared foundations, and S-04 strict parser complete |
| Substantive authored files | `packages/view-contracts/src/inspector.ts`; `packages/view-contracts/src/view-contracts.ts`; package Inspector tests; `tools/test_families/package.view_contracts.json`; this tracker |
| Generated files | `tools/execution_topology_render_index.json`, regenerated by `make generate`; no generated file was hand-edited |
| Extracted responsibility | Inspector source parsing, closed-value validation, cross-field route and seed invariants, per-view feature closure, exact specialized signatures, and frozen metadata construction |
| Registry authority | `inspector.ts` derives every closed vocabulary, ordered per-view feature list, and specialized signature from `viewInspectorRegistry`; repository search finds no second independent TypeScript registry or feature-key map |
| Exact behavior | All four Indicator features match their complete owner-derived tuple before wildcard families, use the required panels/owners/actions/conditions/results, and reject any `record_patch` substitution or other tuple mismatch |
| Scope boundary | Authorization, transport execution, workflow state, frontend controls, and server enforcement remain outside the package. Inspector roles and disabled conditions remain presentation metadata only. |
| Compatibility impact | None. Valid metadata, public types, initialized artifacts, exception envelopes, and package entrypoint are unchanged. Unknown tokens and unsupported tuples continue to fail closed. |
| Test accounting | Added one exact four-signature title to the existing characterization row and regenerated topology; the package manifest remains at 5 active rows |
| Rollback | Restore Inspector parsing to `view-contracts.ts`, remove `inspector.ts` and the added title, then rerun `make generate`; the S-00 owner repair remains independently required |
| Next slice | S-07 façade and compatibility cleanup |

#### S-06 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260804T154322Z-p911864` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T154335Z-p915065` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T154335Z-p915170` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T154335Z-p915197` |
| `make generate` | PASS | `.cartulary/test-results/20260804T154351Z-p916773` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 units | `.cartulary/test-results/20260804T154409Z-p919037` |
| `make service-backed-test-slice OWNER=module.workbook` | PASS, 66/66 units | `.cartulary/test-results/20260804T154448Z-p932038` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T154851Z-p969688` |
| `git diff --check` | PASS | No output after the S-06 checkpoint |

No S-06 validation failed or was skipped. The extraction changed module cohesion only; it did not move security, workflow, transport, or UI responsibilities into the package.

### 13.8 S-07 — Façade and compatibility cleanup

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Dependencies | S-03 through S-06 complete |
| Substantive authored files | `packages/view-contracts/src/index.ts`; `packages/view-contracts/src/types.ts`; `packages/view-contracts/src/view-contracts.ts`; `packages/view-contracts/src/index.test.ts`; this tracker |
| Generated files | None; no generated root or harness output changed in S-07 |
| Final façade | `index.ts` contains only explicit runtime and type re-exports. The package manifest still exposes only `.`, and repository search finds no package subpath import. |
| Approved removals | `parseViewContractJSON` is absent from the public runtime export set but remains directly available to package tests and module initialization in the internal parser module. `ViewRowV1` was removed entirely. |
| Retained API | Both normalized row types/adapters, all used view/surface functions and constants, `buildWorkbookSurfaceContracts`, and the coherent Inspector metadata type graph remain public. |
| Importer audit | No production or test consumer imported either removed name through `@cartulary/view-contracts`; no workspace caller required migration or an alias. All 83 direct package-import occurrences remain on the `.` entrypoint. |
| Initialization evidence | The package slice still loads and validates all 17 artifacts. Its exact runtime-key assertion confirms the raw parser is not accidentally re-exported and every intended value remains available. |
| Test accounting | The existing façade test was strengthened without changing its title or path; no authored family or topology update was required |
| Compatibility impact | Intentional removal of two unused private-workspace API seams only. There is no valid wire behavior, identity, package path, or runtime initialization change. |
| Rollback | Restore the S-06 star exports and `ViewRowV1`; restore the test import to the façade. No generator or consumer rollback is needed. |
| Next slice | S-08 validation and handoff completion |

#### S-07 evidence

| Command | Result | Run root or detail |
| --- | --- | --- |
| `make format` | PASS after final test correction | `.cartulary/test-results/20260804T155105Z-p979049` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 units | `.cartulary/test-results/20260804T155108Z-p982034` |
| `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T155116Z-p983182` |
| `make frontend-import-boundary-check` | PASS, 2/2 units | `.cartulary/test-results/20260804T155116Z-p983201` |
| `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T155116Z-p983214` |
| Repository-wide removed-symbol and subpath scans | PASS | Parser references are internal implementation/test references only; `ViewRowV1` and package subpath imports have zero matches |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T155248Z-p984940` |
| `git diff --check` | PASS | No output after the S-07 checkpoint |

The first package slice failed at `.cartulary/test-results/20260804T155029Z-p976568` because the new expected export-name array was not itself sorted before comparison; the runtime export set was already correct. Sorting both sides fixed the assertion. One initial composite audit shell command also had a quoting error and was replaced by separate read-only scans. No product, consumer, or harness-accounting failure remains, and no S-07 check was skipped.

### 13.9 S-08 — Validation and handoff completion

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04; overall remediation `DONE` |
| Dependencies | S-07 and every preceding slice complete |
| Authored owner accounting | `platform.viewschema`: 3 rows; `package.protocol_ts`: 6 rows; `package.view_contracts`: 5 rows; `module.indicators`: 24 rows, 11 service-backed; `web.workbook`: 119 rows; `module.workbook`: 89 rows, 56 service-backed |
| Test-title/path audit | All new titles resolve through their authored manifests. Only `module.indicators.json` and `package.view_contracts.json` required title/row changes in this effort; no S-08 title or path changed. |
| Topology audit | `make target-plan` reported 616 planned `check` units, 572 cacheable units, 214 fixtures, digest `sha256:a299ea816654335788b36bbd2288d0a9382b93b67931a2e20bfa96c08ed3a27a`; final generation/drift and full execution passed. |
| S-08 authored fixes | Added the two new Indicator source paths to `tools/frontend_source_ownership.json`; removed the now-unused `requireEnumStringFromSet` helper from `tools/contractgen/validation.go`. Both omissions were exposed by the first broad `check`, not by a change in selected product semantics. |
| Generated confirmation | Final `make generate` passed at `.cartulary/test-results/20260804T160836Z-p1276381`. All Go/protocol/view-schema/Inspector validators, bindings, artifacts, topology outputs, and visual goldens were produced through Make-owned generation/update paths; none was hand-edited. |
| Compatibility summary | Existing HTTP paths, `cartulary.view.timeline.v2`, `cartulary.view.indicators.v1`, `cartulary.inspector_config.v1`, and the package `.` entrypoint are preserved. No database migration, alias, fallback, identity bump, or backend authorization transfer was introduced. Intentional incompatibilities are limited to malformed source rejection, `row_version < 1` rejection, compile-time full/sparse separation, corrected Inspector panel placement/dispatch, and removal of two unused private-workspace exports. |
| Retained-run maintenance | Skipped by design because `RESULTS_DIR` was unset for `make agent-finalize`; the target itself passed and no successful retained full warm-check root was claimed. |
| Remaining risks | No known functional, specification, generated-drift, ownership-accounting, security-boundary, or validation blocker remains. The uncommitted implementation still requires ordinary reviewer acceptance of the owner edits and refreshed visual baselines. |
| Rollback | Revert S-07 through S-01 in reverse checkpoint order for isolated package changes. Revert S-00 owner/projection/application/UI/tests/goldens as one compatibility unit, then rerun `make generate`; do not retain Indicator runtime changes against the prior owner corpus. The two S-08 cleanup edits should follow the files that made them necessary. |
| Handoff | Review the owner amendments and generated Inspector registry first, then the typed Indicator executor/handler/UI flow, package module split, strict decoder/row types, test accounting, and S-00 visual evidence. Every required focused and broad gate passes on this worktree. |
| Tracker closure | `make lint-markdown` passed at `.cartulary/test-results/20260804T162619Z-p1560317`; `git diff --check` passed with no output after the final handoff was written |

#### S-08 final validation sequence

| Ordered command | Result | Run root or detail |
| --- | --- | --- |
| 1. `make lint-markdown` | PASS | `.cartulary/test-results/20260804T160853Z-p1278732` |
| 2. `make lint-biome` | PASS, 2/2 units | `.cartulary/test-results/20260804T160901Z-p1280253` |
| 3. `make generate-drift` | PASS, 4/4 units | `.cartulary/test-results/20260804T160911Z-p1280677` |
| 4. `make generated-artifact-policy-check` | PASS, 3/3 units | `.cartulary/test-results/20260804T160920Z-p1283252` |
| 5. `make json-shape-check` | PASS, 3/3 units | `.cartulary/test-results/20260804T160924Z-p1283680` |
| 6. `make test-slice OWNER=platform.viewschema` | PASS, 3/3 units | `.cartulary/test-results/20260804T160932Z-p1284133` |
| 7. `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 scheduled units | `.cartulary/test-results/20260804T160938Z-p1284703` |
| 8. `make test-slice OWNER=package.view_contracts` | PASS, 6/6 scheduled units | `.cartulary/test-results/20260804T160947Z-p1288932` |
| 9. `make test-slice OWNER=module.indicators` | PASS, 27/27 scheduled units | `.cartulary/test-results/20260804T160953Z-p1289838` |
| 10. `make service-backed-test-slice OWNER=module.indicators` | PASS, 13/13 scheduled units | `.cartulary/test-results/20260804T161010Z-p1293859` |
| 11. `make test-slice OWNER=web.workbook` | PASS, 120/120 scheduled units | `.cartulary/test-results/20260804T161025Z-p1295311` |
| 12. `make service-backed-test-slice OWNER=module.workbook` | PASS, 66/66 scheduled units | `.cartulary/test-results/20260804T161108Z-p1308312` |
| 13. `make frontend-typecheck` | PASS, 2/2 units | `.cartulary/test-results/20260804T161305Z-p1344026` |
| 14. `make frontend-import-boundary-check` | PASS, 2/2 units | `.cartulary/test-results/20260804T161318Z-p1344537` |
| 15. `make browser-e2e-webserver-backed` | PASS, 62/62 units | `.cartulary/test-results/20260804T161324Z-p1344980` |
| 16. `make agent-finalize` | PASS, 1/1 unit | `.cartulary/test-results/20260804T161655Z-p1397125`; `RESULTS_DIR` unset |
| 17. `make test-fast` | PASS, 349/349 units | `.cartulary/test-results/20260804T161711Z-p1399669` |
| 18. `make check` | PASS, 715/715 units | `.cartulary/test-results/20260804T161836Z-p1435478` |
| 19. `git diff --check` | PASS | No output |
| 20. `git status --short` | PASS, expected dirty handoff | Only the intended authored, generated, visual-evidence, test-accounting, and tracker paths were present |

#### S-08 failure and recovery record

The first exact sequence stopped at `make check` root `.cartulary/test-results/20260804T160326Z-p1148447` with 713/715 units passing. Both failures were related to this remediation:

- `web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19` found the new `IndicatorInspectorWorkflow.test.tsx` and `indicatorInspectorHandlers.ts` absent from the authored frontend source-ownership manifest.
- Go staticcheck found `requireEnumStringFromSet` unused after Inspector closed-vocabulary validation moved to the owner-derived registry.

The ownership manifest and obsolete helper were corrected. `make test-slice OWNER=web.architecture` then passed 12/12 at `.cartulary/test-results/20260804T160824Z-p1275215`; `make lint-go-staticcheck` passed; `make generate` refreshed downstream artifacts; and the complete 20-command sequence above was restarted from command 1. No final validation was skipped and no unrelated failure remains.

## 14. Production-Readiness Execution Checkpoints

### 14.1 S-09 — Generated production projection

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Baseline | Clean `25cd6532dfd9a6531ac7fffb10b6d051478225b2`; S-08 complete |
| Authored implementation | Added the contractgen browser-projection writer and the package-local `projection.ts` adapter; declared the exact generated output; added semantic parity evidence |
| Authored policy | Added `packages/view-contracts/src/generated` to generated-artifact, drift-scratch, Biome, import-boundary, generated-path, and JSON-shape ownership inputs |
| Generated output | `packages/view-contracts/src/generated/view-contract-projection.ts`, produced only by `make generate`; 17 registry-ordered normalized rows and no raw JSON strings or hashes |
| Runtime posture | Production still uses the S-08 parser in this slice; the projection adapter is test-only parity evidence, so rollback is deletion of the new output/adapter/policy entries |
| Compatibility | Projected contracts equal the characterized runtime values; `listViewContracts` retains its historical stable ID order and projected surfaces retain registry order and shared contract identity |
| Test accounting | Added one ASCII-sorted title to the existing characterization owner row and regenerated topology through Make |
| Failure record | The first package slice found the projected contract list in registry rather than historical contract-list order; the adapter now preserves both established contract order and registry surface order. The first Biome gate found formatting/import order only; `make format` corrected it. |
| Next slice | S-10 package runtime cutover |

| Command | Result | Run root |
| --- | --- | --- |
| `make generate` | PASS | `.cartulary/test-results/20260804T165806Z-p1592139` |
| `make test-slice OWNER=package.view_contracts` | PASS, 6/6 | `.cartulary/test-results/20260804T165912Z-p1595916` |
| `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 | `.cartulary/test-results/20260804T165923Z-p1597199` |
| `make frontend-typecheck` | PASS, 2/2 | `.cartulary/test-results/20260804T165713Z-p1591136` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `.cartulary/test-results/20260804T165923Z-p1597516` |
| `make generate-drift` | PASS, 4/4 | `.cartulary/test-results/20260804T165923Z-p1596967` |
| `make generated-artifact-policy-check` | PASS, 3/3 | `.cartulary/test-results/20260804T165923Z-p1597005` |
| `make json-shape-check` | PASS, 3/3 | `.cartulary/test-results/20260804T165923Z-p1597045` |
| `make test-fast` | PASS, 349/349 | `.cartulary/test-results/20260804T165923Z-p1597801` |
| `make format` | PASS after formatting-only failure | `.cartulary/test-results/20260804T170034Z-p1638758` |

### 14.2 S-10 — Package runtime cutover

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Authored implementation | Replaced runtime source parsing with the package-local projection adapter; made surface initialization consume the projected surfaces; folded row-only invariant helpers into `rows.ts` |
| Removed implementation | Deleted `inspector.ts`, `invariants.ts`, and the complete JSON/source-decoder parser path from `view-contracts.ts` |
| Tests and accounting | Replaced parser-invalidity characterization with generated-runtime, row, metadata, and projection-boundary evidence; reduced the package manifest to four authored rows and migrated the module-workbook collaborator title |
| Compatibility | All 17 view contracts retain historical list order; surfaces retain registry order, shared contract identity, freezing, fields, synthetic filters, Inspector metadata, and lookup behavior |
| Runtime audit | No authored package production source references raw artifacts, source types, the source decoder, the Inspector registry value, or `JSON.parse` |
| Rollback | Restore the S-09 parser/Inspector/invariants modules and production surface builder while the legacy protocol surface still exists; no generator rollback is required |
| Failure record | Topology initially rejected a removed module-workbook collaborator title and then an overlapping replacement title. Ownership was made singular under module-workbook. One compatibility test removed the first alphabetically sorted contract rather than Timeline; it now filters by stable ID. |
| Next slice | S-11 protocol and generator legacy removal |

| Command | Result | Run root |
| --- | --- | --- |
| `make generate` | PASS | `.cartulary/test-results/20260804T170719Z-p1667775` |
| `make format` | PASS | `.cartulary/test-results/20260804T170831Z-p1697756` |
| `make test-slice OWNER=package.view_contracts` | PASS, 5/5 | `.cartulary/test-results/20260804T170839Z-p1700807` |
| `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 | `.cartulary/test-results/20260804T170739Z-p1673396` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 | `.cartulary/test-results/20260804T170739Z-p1673433` |
| `make frontend-typecheck` | PASS, 2/2 | `.cartulary/test-results/20260804T170739Z-p1673713` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `.cartulary/test-results/20260804T170739Z-p1673769` |
| `make lint-biome` | PASS, 2/2 | `.cartulary/test-results/20260804T170739Z-p1673850` |
| `make build-web` | PASS | `.cartulary/test-results/20260804T170739Z-p1674290` |

### 14.3 S-11 — Protocol and generator legacy removal

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Authored generator change | Adopted `cartulary.contract_family_registry.v5`; removed the shared TypeScript artifact support output, barrel seam, `artifact_collection`, raw artifact writer, and TypeScript view-source type writer while retaining the Go source-type writer |
| Authored protocol change | Reduced `@cartulary/protocol-ts/view-schemas` to the frozen view registry, registry listing, and literal-derived Inspector types; the generated Inspector registry is now a type-only dependency |
| Generated removal | `artifact.ts`, `view-schemas-artifacts.ts`, `view-schema-source-types.ts`, and `view-schema-source-validator.ts` were deleted only through their owning Make generators |
| Policy and schema change | Replaced the v4 registry schema and attachment with v5; removed obsolete entrypoint allowlist and import-boundary members; added removed modules to browser artifact-reachability enforcement |
| Compatibility | Registry values and literal Inspector types remain; browser raw artifacts, runtime decoding, runtime Inspector registry access, and all TypeScript source-document types are intentionally removed without aliases |
| Rollback | Restore registry v4, the raw projection/support writers, source decoder generator, four generated outputs, and the protocol entrypoint as one generator/protocol unit |
| Failure record | No validation failure. Regeneration updated the extension integrity artifact because `main.go` is an integrity-tracked generator source. |
| Next slice | S-12 dead façade and workbook seam removal |

| Command | Result | Run root |
| --- | --- | --- |
| `make generate` | PASS | `.cartulary/test-results/20260804T171537Z-p1705669` |
| `make format` | PASS | `.cartulary/test-results/20260804T171556Z-p1708102` |
| `make json-shape-check` | PASS, 3/3 | `.cartulary/test-results/20260804T171617Z-p1711295` |
| `make generated-artifact-policy-check` | PASS, 3/3 | `.cartulary/test-results/20260804T171617Z-p1711304` |
| `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 | `.cartulary/test-results/20260804T171617Z-p1711412` |
| `make frontend-typecheck` | PASS, 2/2 | `.cartulary/test-results/20260804T171617Z-p1711523` |
| `make generate-drift` | PASS, 4/4 | `.cartulary/test-results/20260804T171638Z-p1717603` |
| `make test-slice OWNER=platform.viewschema` | PASS, 3/3 | `.cartulary/test-results/20260804T171638Z-p1717724` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `.cartulary/test-results/20260804T171638Z-p1717865` |
| `make lint-biome` | PASS, 2/2 | `.cartulary/test-results/20260804T171638Z-p1717923` |
| `make test-fast` | PASS, 348/348 | `.cartulary/test-results/20260804T171657Z-p1721619` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T171947Z-p1771098` |
| `git diff --check` | PASS | Repository root after S-11 checkpoint |

### 14.4 S-12 — Dead façade and workbook seam removal

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Package removal | Removed `fieldCapability`, `visibleFields`, `ViewFieldCapability`, `buildWorkbookSurfaceContracts`, `getWorkbookSurfaceContract`, and `requireWorkbookSurfaceContract` without aliases |
| Web removal | Workbook startup now consumes `listWorkbookSurfaceContracts()` directly; removed the injectable `buildWorkbookSurfaceRegistry` and its impossible-state/optional-omission tests |
| E2E removal | Deleted the workbook-surface re-export module and migrated 30 E2E import sites directly to `@cartulary/view-contracts`; the import-boundary rule now forbids coupling to the web registry rather than requiring the deleted seam |
| Test migration | Replaced the remaining `fieldCapability` assertion with direct `ViewFieldContract.gridEditable` metadata and removed parser/builder compatibility titles from authored manifests before regenerating topology |
| Importer audit | Exact repository scans are empty for every removed function/type and the deleted E2E module path; package export-surface evidence passes |
| Compatibility | Intentional private-workspace API removal only; all production consumers use immutable generated-projection lists and all 17 surfaces remain present |
| Rollback | Restore the façade exports, injected package/web builders, E2E re-export, migrated imports, and removed test titles as one local façade/web unit; no protocol rollback is required |
| Failure record | No validation failure |
| Next slice | S-13 focused test architecture and package cleanup |

| Command | Result | Run root |
| --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260804T172229Z-p1775287` |
| `make generate` | PASS | `.cartulary/test-results/20260804T172236Z-p1778328` |
| `make test-slice OWNER=package.view_contracts` | PASS, 5/5 | `.cartulary/test-results/20260804T172251Z-p1780828` |
| `make test-slice OWNER=web.workbook` | PASS, 120/120 | `.cartulary/test-results/20260804T172251Z-p1780849` |
| `make frontend-typecheck` | PASS, 2/2 | `.cartulary/test-results/20260804T172251Z-p1781042` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `.cartulary/test-results/20260804T172251Z-p1781084` |
| `make lint-biome` | PASS, 2/2 | `.cartulary/test-results/20260804T172251Z-p1781143` |
| `make test-slice OWNER=web.architecture` | PASS, 12/12 | `.cartulary/test-results/20260804T172338Z-p1795822` |
| `make build-web` | PASS | `.cartulary/test-results/20260804T172339Z-p1796034` |
| `make browser-e2e-webserver-backed` | PASS, 62/62 | `.cartulary/test-results/20260804T172350Z-p1801690` |
| `make explain-test-owner OWNER=package.view_contracts` | PASS, four rows | Repository root after S-12 |
| `make explain-test-owner OWNER=module.workbook` | PASS, 89 rows | Repository root after S-12 |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T172756Z-p1854872` |
| `git diff --check` | PASS | Repository root after S-12 checkpoint |

### 14.5 S-13 — Test architecture and package cleanup

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Production structure | Consolidated contract and surface lookup into `contracts.ts`; the production graph is exactly `index.ts`, `types.ts`, `rows.ts`, `projection.ts`, and `contracts.ts` with one-way dependencies and no cycle |
| Test structure | Replaced the monolith with `facade.test.ts`, `rows.test.ts`, `contracts.test.ts`, and `surfaces.test.ts`; retained meaningful behavior titles and removed parser/builder-only cases |
| Harness accounting | Rebuilt four ASCII-sorted package rows around the four test files, moved the module-workbook collaborator selector to `contracts.test.ts`, and regenerated topology through Make |
| Cleanup | Removed `.gitkeep`; source scans find no raw-source/parser APIs, obsolete generated outputs, removed façade symbols, old internal modules, or unaccounted package test path |
| Tracker revision | Replaced stale sections 1 through 12 with the production-readiness baseline, target architecture, compatibility decisions, workstreams, risk register, and binary completion criteria; section 13 remains the S-00 through S-08 historical record |
| Compatibility | Public runtime behavior is unchanged from S-12; this slice changes internal module and test organization only |
| Rollback | Restore the two lookup modules, monolithic test, prior manifests, and sentinel as the isolated S-13 unit; no generator/protocol rollback is required |
| Failure record | `make format` initially rejected unsorted authored row IDs; sorting the four manifest rows resolved the owner-input failure before topology generation |
| Next slice | S-14 validation and handoff completion |

| Command | Result | Run root |
| --- | --- | --- |
| `make format` | PASS after authored-row sort | `.cartulary/test-results/20260804T173233Z-p1859544` |
| `make generate` | PASS | `.cartulary/test-results/20260804T173243Z-p1862598` |
| `make test-slice OWNER=package.view_contracts` | PASS, 5/5 | `.cartulary/test-results/20260804T173254Z-p1864823` |
| `make test-slice OWNER=module.workbook` | PASS, 99/99 | `.cartulary/test-results/20260804T173340Z-p1865959` |
| `make generate-drift` | PASS, 4/4 | `.cartulary/test-results/20260804T173339Z-p1865900` |
| `make frontend-typecheck` | PASS, 2/2 | `.cartulary/test-results/20260804T173340Z-p1866118` |
| `make lint-biome` | PASS, 2/2 | `.cartulary/test-results/20260804T173340Z-p1866135` |
| `make generated-artifact-policy-check` | PASS, 3/3 | `.cartulary/test-results/20260804T173809Z-p1911049` |
| `make json-shape-check` | PASS, 3/3 | `.cartulary/test-results/20260804T173809Z-p1911075` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `.cartulary/test-results/20260804T173809Z-p1911220` |
| `make explain-test-owner OWNER=package.view_contracts` | PASS, four focused rows | Repository root after S-13 |
| `make target-plan` | PASS, 615 units and digest `sha256:4c5fc36ce6502f58756fc835cea41d87d4803f320e61906afb65508684ca7ec7` | Repository root after S-13 |
| Source/output/import scans | PASS | Repository root after S-13 |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260804T173858Z-p1912553` |
| `git diff --check` | PASS | Repository root after S-13 checkpoint |

### 14.6 S-14 — Validation and handoff

| Checkpoint field | Result |
| --- | --- |
| Status | `DONE` 2026-08-04 |
| Final generated confirmation | All authored generator/policy changes were regenerated through Make; the package-local projection and integrity-tracked Go artifact are current; four obsolete protocol outputs are absent; final drift and policy checks pass |
| API removal confirmation | Raw artifact/source decoder/types, runtime Inspector registry value, capability/visibility helpers, package and web surface builders, package surface lookup helpers, and the E2E re-export module are absent without aliases |
| Compatibility | All view/Inspector identities, routes, row wire shapes, 17 contracts, 17 surfaces, status arrays, retained metadata types, list/get/require contract APIs, sort resolution, and row normalizers remain; private dead APIs and malformed-source browser behavior are intentionally removed |
| Source-size result | The selected legacy runtime path was 459,654 bytes at baseline; its generated projection and authored adapters are 388,306 bytes, a 71,348-byte or 15% reduction, while moving validation entirely to generation |
| Final source graph | Exactly five authored production modules and four focused test files; no cycle, `.gitkeep`, raw-source reference, obsolete output, removed-symbol importer, or unaccounted package test remains |
| Remaining risk | The normalized projection remains a substantial generated browser input and must stay drift/policy gated. Any future conditional surface availability requires a new explicit owner model rather than restoring omission or injection seams. No known correctness blocker remains. |
| Rollback guidance | Revert S-13/S-12 locally for façade/test structure; revert S-11 as the registry-v4/protocol/generator unit; restore S-10 parser initialization before reverting S-09 projection ownership. No database, route, identity, or wire rollback is required. |
| Retained-run maintenance | Skipped because `RESULTS_DIR` was intentionally unset for `make agent-finalize`; the command passed with a fresh run root |
| Final worktree | `git status --short` contains only the expected authored/generated implementation, test, policy, topology, and tracker changes; no unrelated pre-existing edit was present |
| Handoff | Production reads one generator-owned projection, protocol exposes only registry values and Inspector literal types, the package has one small public façade and five-module graph, workbook/E2E consumers use immutable lists directly, and all focused/broad validation is green |

The first S-14 attempt stopped correctly at `make check`: 713 of 714 units
passed in `.cartulary/test-results/20260804T174710Z-p2038804`, and Go
staticcheck identified the now-unused `marshalJSONString` helper left after the
artifact/barrel writer removal. The helper was deleted, `make generate` passed at
`.cartulary/test-results/20260804T175149Z-p2168396`, and
`make lint-go-staticcheck` passed. The complete sequence was then restarted from
command 1; only the successful restarted sequence is completion evidence.

| Ordered command | Final result | Run root |
| --- | --- | --- |
| 1. `make lint-markdown` | PASS | `.cartulary/test-results/20260804T175231Z-p2172270` |
| 2. `make lint-biome` | PASS, 2/2 | `.cartulary/test-results/20260804T175238Z-p2173784` |
| 3. `make generate-drift` | PASS, 4/4 | `.cartulary/test-results/20260804T175246Z-p2174205` |
| 4. `make generated-artifact-policy-check` | PASS, 3/3 | `.cartulary/test-results/20260804T175255Z-p2176796` |
| 5. `make json-shape-check` | PASS, 3/3 | `.cartulary/test-results/20260804T175303Z-p2177233` |
| 6. `make test-slice OWNER=harness.generated_artifacts` | PASS, 2/2 | `.cartulary/test-results/20260804T175315Z-p2177732` |
| 7. `make test-slice OWNER=platform.viewschema` | PASS, 3/3 | `.cartulary/test-results/20260804T175326Z-p2180150` |
| 8. `make test-slice OWNER=package.protocol_ts` | PASS, 7/7 | `.cartulary/test-results/20260804T175334Z-p2180722` |
| 9. `make test-slice OWNER=package.view_contracts` | PASS, 5/5 | `.cartulary/test-results/20260804T175343Z-p2184941` |
| 10. `make test-slice OWNER=web.architecture` | PASS, 12/12 | `.cartulary/test-results/20260804T175401Z-p2185793` |
| 11. `make test-slice OWNER=web.workbook` | PASS, 120/120 | `.cartulary/test-results/20260804T175409Z-p2187267` |
| 12. `make frontend-typecheck` | PASS, 2/2 | `.cartulary/test-results/20260804T175452Z-p2200283` |
| 13. `make frontend-import-boundary-check` | PASS, 2/2 | `.cartulary/test-results/20260804T175524Z-p2200867` |
| 14. `make build-web` | PASS | `.cartulary/test-results/20260804T175534Z-p2201357` |
| 15. `make browser-e2e-webserver-backed` | PASS, 62/62 | `.cartulary/test-results/20260804T175544Z-p2204837` |
| 16. `make agent-finalize` | PASS, 1/1 | `.cartulary/test-results/20260804T175916Z-p2256864` |
| 17. `make test-fast` | PASS, 348/348 | `.cartulary/test-results/20260804T175929Z-p2259416` |
| 18. `make check` | PASS, 714/714 | `.cartulary/test-results/20260804T180057Z-p2295567` |
| 19. `git diff --check` | PASS | Repository root after final broad validation |
| 20. `git status --short` | PASS, expected implementation set captured | Repository root after final broad validation |

After this checkpoint was authored, `make lint-markdown` passed at
`.cartulary/test-results/20260804T180608Z-p2420608` and `git diff --check`
passed at the repository root. No command or generated output remains pending.
