# NLSpec-First Testing Refactor Tracker and Handoff

## 1. Mission and Source Posture

This tracker governs the repository-wide replacement of the current
machine-first testing authority with this derivation chain:

**Intent → adopted NLSpec/Core specification → versioned machine projection →
implementation and behavioral tests**

This is a fundamental testing-program overhaul, not an incremental cleanup.
Existing behavior has no independent preservation priority: each affected
current Definition-of-Done obligation is retained or implemented when it has
continuing value, or revised in the owning specification before it is pruned or
moved to a future profile. Harness metadata, traceability schemas, generated
artifacts, public descriptors that expose accidental provenance, and retained
evidence may break through the coordinated cutover defined below.

The source hierarchy for this effort is:

1. Adopted subsystem NLSpecs for their named scopes and normative Core owner
   sections for the current profile. Core 00 resolves document ownership and
   precedence.
2. Versioned schemas, enums, limits, mappings, algorithms, fixtures, and other
   machine-readable contracts derived from those specifications.
3. Test-routing and evidence catalogs, which describe execution and results but
   do not define requirements or prove specification completeness.
4. Current implementation, generated outputs, guides, research, and handoffs,
   which cannot override adopted specifications.

`docs/research/nlspec-spec.md` supplies the planning doctrine: an NLSpec is
prescriptive and generative; its Definition of Done defines what must be true,
not how tests must prove it; tests are downstream; and each concept should be
specified once. `docs/testing-harness-nlspec.md` is the foundation and intended
authoritative owner for harness behavior after its status repair in `NS-01`.
This tracker is a downstream execution ledger and is not itself a normative
product or harness specification.

When a specification and a machine artifact disagree, the machine artifact is
defective. Behavior-affecting work starts with the owning specification, then
updates derived machine contracts, implementation, and behavioral tests.
Executable validation never settles such a conflict by reading, parsing,
hashing, or comparing the specification text.

### Program boundaries

| Item | Decision |
| --- | --- |
| Target artifact | `docs/handoffs/nlspec-testing-refactor-tracker.md` |
| Repository baseline | Tracker creation at `b45905b1`; implementation planning reconciled at `286c91a1` with a clean tracked worktree |
| Archaeological boundary | Commit `a7b4eced` introduced the current requirement-catalog authority; it is evidence, not a revert target |
| Program scope | Repository-wide, prioritizing harness, OpenAPI, Extensions, Graph Projection, Reporting, and Report Composition |
| Product behavior | Governed by the final adopted owner specifications; accidental provenance receives no compatibility support |
| Compatibility posture | One coordinated breaking cutover; no compatibility readers, aliases, dual writers, or release window with two authorities |
| Generated files | Update authored owners and run official generators; never hand-edit policy-managed generated roots |
| Specification access | Documentation tooling may lint documentation; product, test, generator, conformance, and release inputs may not consume `docs/` |
| Public conformance posture | Behavioral evidence supports review against specifications; mapping counts cannot publish or imply complete conformance |

Status values in this tracker are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`,
`DEFERRED`, and `DROPPED`.

## 2. Current-State Baseline and Diagnosis

The current repository is internally consistent but implements the wrong
authority direction. Planning baselines passed as follows:

| Command | Result | Run root or summary | What it proves |
| --- | --- | --- | --- |
| `make generate-drift` | Passed | `.cartulary/test-results/20260726T204021Z-p2814881` | Current authored/generated inputs agree with the current machine-first model |
| `make json-shape-check` | Passed | `.cartulary/test-results/20260726T204021Z-p2814883` | Current registered JSON artifacts satisfy their current schemas |
| `make harness-contract` | Passed | `.cartulary/test-results/20260726T204021Z-p2815081/harness-contract/target-summary.json` | Current harness catalogs are mutually consistent |

These results are characterization evidence only. They do not validate the
current precedence model or justify retaining duplicated natural-language
requirements.

### Quantitative inventory

| Surface | Current quantity | Diagnosis |
| --- | ---: | --- |
| Requirement owner catalogs | 73 | Parallel natural-language authority under `contracts/requirements/owners/` |
| Requirement entries | 723 | 607 `active`, 116 `planned`; copied statements and lifecycle status influence executable coverage |
| Verification owners | 60 | Verification v2 requires requirement IDs and resolves them against the parallel catalogs |
| Verification entries | 94 | Every active machine requirement must map to verification |
| Test-family manifests | 54 | Routing inputs are useful, but currently inherit requirement authority through verification v2 |
| Test rows | 936 | 895 `implementation`, 41 `informative`; row accounting is useful but is not specification completeness |
| Extension traceability mappings | 2,590 | Requirement and acceptance-ID accounting is generated into runtime-adjacent contracts |
| Graph acceptance-accounting rows | 69 | A conformance matrix asserts completeness independently from executing behavior |
| Graph executable fixture directories | 36 | Valuable behavior fixtures to retain after removing requirement IDs and redundant registries |
| Reporting placeholder rows | 80 | One `implemented`, 79 `planned`; rows contain IDs/status only and are not executable fixtures |
| Report Composition placeholder rows | 27 | Twenty-six `implemented`, one `planned`; rows contain IDs/status only and are not executable fixtures |
| Network Flow activity-accounting rows | 107 | Copied owner requirements, acceptance IDs, exact selectors, and fixture mappings duplicate specification structure |
| Network Flow executable fixture directories | 28 | Valuable behavior fixtures to retain after removing accounting metadata |
| OTel checked-in conformance declarations | 1 | Static pass/claim status is trusted by the conformance checker instead of being derived solely from execution |

### Structural findings

| Finding | Evidence | Required disposition |
| --- | --- | --- |
| Core 00 and repository guidance say machine contracts override specifications | Core 00 opening sections, `AGENTS.md`, README/developer guidance | Repair authority prose in `NS-01` |
| The harness specification is demoted to human reference despite containing normative requirements and acceptance criteria | Testing harness frontmatter and Sections 1, 3, 8, and 16–18 | Restore adopted NLSpec status and clarify the no-document-input boundary |
| Requirement statements and `active`/`planned` statuses participate in semantic digests and coverage gates | Requirement and verification loaders | Delete the requirement layer and its completeness gates in `NS-06` |
| Test and release evidence identity incorporates requirement-derived verification semantics | Slice, owner, browser, accounting, audit, and release consumers | Introduce the new digest vocabulary and evidence epoch in `NS-06` |
| Contract-family activation is linked through synthetic requirement IDs | `contracts/index.json` and contract generator validation | Move to contract-family registry v3 in `NS-02` |
| OpenAPI assembly consults the global requirement registry only to recognize unit owners | OpenAPI source manifest and assembler | Remove the external registry dependency in `NS-02` |
| Extensions embed requirement IDs, catalog paths, hashes, and coverage mappings in generated/runtime-adjacent contracts | Extension dependencies, owner fragments, traceability source, generated contracts, and module tests | Replace operational requirement references with typed contract IDs in `NS-03` |
| Graph, Reporting, and Report Composition contain accounting-only tests and corpora | Conformance matrix, fixture corpus registries, and traceability tests | Delete accounting-only artifacts while preserving behavior tests in `NS-04` |
| Network Flow duplicates owner text and acceptance-to-selector mappings | Activity-accounting manifest, fixture manifests, scenario files, and AC-switch unit helpers | Delete accounting and retain direct behavior execution in `NS-04A` |
| OTel conformance consumes a checked-in pass/claim declaration | OTel conformance status manifest and checker | Derive conformance only from real checks in `NS-04B` |
| The generic executable-input restriction already prevents dependencies on `docs/` without parsing documentation | Executable-input policy and restricted-input boundary | Retain and reframe as architecture protection |
| The old extension document parser checked headings, byte ranges, counts, hashes, and formatting | Deleted code in the parent history of `a7b4eced` | Do not restore it or any equivalent document parser |

### Gap decision register

| Gap | Areas | Remediation | Long-term benefit | Migration impact | Risk if unresolved | Completion evidence |
| --- | --- | --- | --- | --- | --- | --- |
| `G1` authority inversion | specifications, guidance, implementation, tests | Restore adopted NLSpec/Core authority and remove machine completeness authority | One semantic source of truth | Derived contracts may break; runtime compatibility is not implied | Catalogs can silently redefine behavior | Authority audit and Markdown lint pass |
| `G2` executable document coupling | specifications, policy, tests | Use typed projections only when executable facts are needed; preserve the generic `docs/` boundary | Prose remains editable and non-executable | Executable `docs/` references become hard failures | Formatting and paths become accidental APIs | Synthetic reference is rejected without document I/O |
| `G3` contract-family/OpenAPI coupling | specifications, contracts, generators, tests | Remove both contract-family owner arrays and the OpenAPI requirement-registry dependency | Local, cohesive owner validation | Registry and manifest v2 are rejected | Requirement deletion breaks generation | Bytes, shapes, generation, and owner slices pass |
| `G4` Extensions provenance | specification, contracts, runtime, tests | Remove traceability fields; use direct typed facts or real resolvable contracts only | Smaller, reusable, future-facing contracts | Coordinated schema majors; accidental public metadata is removed | Thousands of fragile mappings become permanent | Forbidden-field audit and behavior/security slices pass |
| `G5` Graph/Reporting/Composition accounting | specifications, fixtures, tests, routing | Delete accounting-only corpora while retaining/rerouting behavior | Tests assert observable behavior | Internal fixture/routing schemas break | Counts remain green while behavior is absent | Real fixtures and behavior rows pass |
| `G6` Network Flow accounting | specifications, contracts, fixtures, tests | Delete copied acceptance accounting and selectors; retain direct behavior fixtures | Stable test identities independent of prose structure | Fixture/scenario schema majors; no old readers | AC switches hide missing behavior | All real fixture sets and direct behavior checks pass |
| `G7` OTel self-attestation | specification, contracts, conformance, tests | Delete checked-in pass/claim status and derive success from execution | Honest, reproducible conformance evidence | Static v1 claim format is removed | Stale declarations can publish readiness | OTel real checks pass without a status file |
| `G8` verification/evidence coupling | contracts, harness, release evidence | Introduce routing-only verification v3, new evidence majors/epoch, and delete requirements | Evidence identifies execution, not prose | Atomic breaking cutover; historical evidence is not migrated | Stale evidence and requirement edits remain coupled | Harness passes with v3 and rejects old evidence |
| `G9` unresolved current obligations | specifications, implementation, tests, handoff | Record `RETAIN-ROUTE`, `IMPLEMENT`, or `SPEC-PRUNE/FUTURE`; forbid planned executable rows | Honest current scope and extensible future phases | Product behavior may change only after its owner specification | Cleanup can hide real product gaps | Every affected current obligation has a durable disposition |

## 3. Authority and Test-Disposal Contract

### 3.1 Machine-projection rule

A test that needs an enum, schema, limit, mapping, algorithm, fixture, or other
machine-testable fact must consume a machine-readable artifact with a versioned
`schema_id` or contract identity. That artifact:

- contains only executable facts needed by production, generation, or testing;
- is reviewed as a projection of the owning specification;
- does not copy a narrative requirement statement or specification lifecycle
  status;
- does not identify source lines, headings, anchors, paths, byte offsets, or
  document hashes;
- cannot override the owning specification; and
- is registered with the appropriate schema attachment and shape validation.

Human review establishes that a projection faithfully expresses its owner
specification. No executable tool proves that relationship by consuming the
specification document.

### 3.2 No specification trace annotations

Verification contract v3 contains routing semantics only. It does not accept
`requirement_ids`, `spec_trace_ids`, copied acceptance identifiers,
specification lifecycle status, or any equivalent provenance field. Human
review may record a temporary obligation disposition in this tracker, but no
executable loader resolves or carries that documentation.

### 3.3 Fixed disposition matrix

| Disposition | Apply when | Examples | Required result |
| --- | --- | --- | --- |
| `REMOVE` | The test or artifact exists only to compare documentation or account for copied requirements | Document parsers; prose/heading/anchor/line/hash checks; requirement statements/statuses; AC counts; planned placeholders; traceability-only corpora; catalog-hash gates | Delete the test/artifact and its catalog or schema registration |
| `REFACTOR` | A test mixes useful behavior with traceability accounting | Runtime test plus requirement-array assertion; output-kind behavior inside a fixture-corpus accounting test | Retain the behavior assertion in an owner-aligned behavioral test and remove accounting |
| `RETAIN` | The test exercises production-relevant behavior or validates a versioned machine contract | Runtime/API/persistence/security/failure tests; real fixture/golden execution; schema validation; generator drift | Preserve behavior and route it through verification v3 |
| `EDITORIAL_ONLY` | The check maintains documentation quality but does not establish product conformance | Markdown lint | Keep outside product, conformance, release, and generated evidence |
| `ARCHITECTURE_PROTECTION` | The check prevents executable dependencies on specifications without inspecting their content | Generic `docs/` restricted-input boundary | Retain; test policy behavior with synthetic executable references |

Every affected current Definition-of-Done obligation receives one human
disposition before its old accounting is removed:

- `RETAIN-ROUTE`: existing valuable behavior and real evidence are retained;
- `IMPLEMENT`: valuable current behavior is implemented and tested; or
- `SPEC-PRUNE/FUTURE`: the owning specification is revised first because the
  behavior does not belong in the current future-facing profile.

Future specification behavior may remain in an explicitly future profile or a
handoff tracker. It must not appear as an executable placeholder, active
verification, passing row, or release-evidence claim.

## 4. Coordinated Cutover Strategy

All implementation work from `NS-01` through `NS-07` belongs to cutover set
`NLTEST-CUTOVER-1`. The branch is non-releasable after `NS-01` begins and
remains so until `NS-07` passes. Intermediate commits are rollback checkpoints,
not supported repository states.

The order is deliberate:

1. Repair the specifications and human guidance first.
2. Decouple production and generated consumers from requirement catalogs while
   the old central harness remains temporarily intact.
3. Remove traceability-only subsystem tests and artifacts, including Network
   Flow acceptance accounting and OTel static conformance declarations.
4. Audit all remaining consumers.
5. In one central cutover, migrate verification and evidence schemas and delete
   the requirement layer.
6. Regenerate, rebaseline, and run full validation.

No step adds a reader or writer that accepts both old and new schema versions.
No pre-cutover evidence can satisfy a post-cutover gate.

The new required evidence epoch is:

`cartulary.test_evidence.nlspec.v1`

Every post-cutover plan, owner summary, accounting result, audit summary,
browser plan/index, scheduler summary, and release-evidence artifact that
currently carries catalog/verification digests must carry this exact
`evidence_epoch`. Current tools reject absent or mismatched epochs. Historical
artifacts remain historical files only.

## 5. Workstream Dependency Map

| Slice | Name | Class | Depends on | Owner | Merge/release state |
| --- | --- | --- | --- | --- | --- |
| `NS-00` | Tracker and characterization baseline | root | None | Handoff owner | Tracker-only; releasable |
| `NS-01` | Specification authority repair | chain | `NS-00` | Core and testing-harness specification owners | Begins non-releasable cutover |
| `NS-02` | Contract-family and OpenAPI decoupling | chain | `NS-01` | `platform.openapi`, contract generation owner | Non-releasable |
| `NS-03` | Extensions typed-contract migration | chain | `NS-01`, `NS-02` contract-ID rules | `module.extensions` | Non-releasable |
| `NS-04` | Graph, Reporting, and Composition test cleanup | chain | `NS-03` | Subsystem owners | Non-releasable |
| `NS-04A` | Network Flow accounting removal | chain | `NS-04` | `module.networkflow` and affected owners | Non-releasable |
| `NS-04B` | OTel derived conformance | chain | `NS-04A` | `platform.telemetry` | Non-releasable |
| `NS-05` | Repository-wide authority-consumer audit | chain | `NS-04B` | Cross-owner harness maintainer | Non-releasable |
| `NS-06` | Verification v3, evidence epoch, and requirement-layer deletion | chain | `NS-05` | Harness catalog/evidence owners | Atomic central cutover; non-releasable |
| `NS-07` | Generation, drift closure, and evidence rebaseline | chain | `NS-06` | Generated-artifact and affected owners | Releasable only after all exit checks pass |
| `NS-08` | Full validation and final handoff | chain | `NS-07` | Cross-owner release verifier | Program closure |

The slices execute strictly serially. After each slice, run its validation,
append the handoff entry, record schema and obligation dispositions, run
`make lint-markdown`, and mark the slice `DONE` or `BLOCKED`. The next slice
must remain `TODO` until that tracker checkpoint is complete; no slices run in
parallel.

## 6. Implementation Slice Ledger

### NS-00: Tracker and characterization baseline

**Status:** `DONE`; tracker-only Markdown validation passed with run root
`.cartulary/test-results/20260726T205341Z-p2826022`.

**Actions:**

1. Record the source posture, counts, hotspots, known consumers, current passing
   baselines, cutover decision, and evidence epoch.
2. Treat `a7b4eced` as archaeological evidence and preserve its useful generic
   no-document-input boundary.
3. Do not change product code, contracts, harness inputs, or generated files
   while creating this tracker.

**Exit:** Sections 1–13 are decision-complete and `make lint-markdown` passes.

### NS-01: Specification authority repair

**Status:** `DONE`.

**Actions:**

1. Restore `docs/testing-harness-nlspec.md` to an adopted/current NLSpec with
   its original harness identity and a title that describes a specification,
   not a human reference.
2. Rewrite its authority, input, catalog, integration, acceptance, and evidence
   provisions so specifications define required outcomes while versioned
   machine projections and behavioral tests implement verification.
3. State explicitly that acceptance criteria define what must be true, not a
   mandatory one-to-one test or machine-requirement mapping.
4. Preserve the generic rule that executable product, test, generator,
   conformance, and release inputs cannot read `docs/`. Preserve standalone
   Markdown lint as editorial maintenance.
5. Restore Core 00's specification-first precedence. Remove language saying a
   machine contract wins conflicts or that machine status/coverage determines
   specification conformance.
6. Align `AGENTS.md`, README/developer guidance, and `docs/domain.md`. Domain
   remains vocabulary and owner navigation; replace proposed direct Domain-text
   validation with versioned projection validation.
7. Audit every adopted subsystem NLSpec for wording that demotes specifications
   or grants machine artifacts independent authority. Do not rewrite unrelated
   behavioral requirements.

**Acceptance:**

- Adopted specifications are the only behavioral authority.
- Derived artifacts are explicitly downstream.
- No specification tells tests to read or structurally validate the document.
- Human guidance consistently describes the same precedence.

**Validation:** `make lint-markdown`.

### NS-02: Contract-family and OpenAPI decoupling

**Status:** `DONE`.

**Contract-family changes:**

1. Replace `cartulary.contract_family_registry.v2` with v3.
2. Remove both `owner_requirement_ids` and `owner_contract_ids` from the
   schema, authored registry, generator types, validation, and tests. The
   current contract IDs are synthetic and do not resolve to executable
   contracts.
3. Retain `generation_status` only as generator lifecycle configuration. It
   must not claim specification adoption or conformance and must not be checked
   against a requirement status.

**OpenAPI changes:**

1. Replace `cartulary.openapi_source_manifest.v2` with v3.
2. Remove `requirements_registry` from the manifest, assembler type,
   validation, fixtures, and tests.
3. Validate owners locally: one root unit owned by `platform.openapi`, one unit
   per owner, owner IDs matching the existing owner-ID lexical form, paths under
   `unit_root/<owner_id>/`, unique paths, permitted roles, regular
   non-symlink files, containment, and existing resource limits.
4. Do not introduce another global owner or requirement registry.

**Acceptance:**

- Contract generation and OpenAPI assembly contain no path or reference to
  `contracts/requirements`.
- Generated OpenAPI bytes and production HTTP behavior remain unchanged.
- Registry and manifest v2 are removed without compatibility readers.

**Validation:**

- `make json-shape-check`
- `make generate`
- `make generate-drift`
- `make generated-artifact-policy-check`
- `make test-slice OWNER=platform.openapi`
- `make test-slice OWNER=harness.generated_artifacts`

### NS-03: Extensions typed-contract migration

**Status:** `DONE`.

**Actions:**

1. Delete the extension traceability mapping source and stop generating
   requirement-coverage artifacts and schemas.
2. Remove requirement registry/catalog paths, schema IDs, hashes, imported
   requirement IDs, acceptance-criterion IDs, owner requirement IDs, required
   specification statuses, catalog-coverage checks, and requirement-bearing
   closure members from authored and generated extension contracts.
3. Classify every former requirement reference:
   - remove it when it is provenance only;
   - embed a direct typed value when the fact is single-use and needs no
     independent lookup; or
   - use a typed contract ID only when it identifies a real versioned machine
     contract and a consumer resolves that contract.
4. Do not mechanically rename `owner_requirement_id`,
   `resolution_requirement_id`, `authorization_requirement_id`,
   `redaction_requirement_id`, `error_requirement_id`, or
   `codec_requirement_id`. If a reusable executable fact has no projection,
   author the smallest versioned typed contract; do not copy the specification
   sentence or invent an unresolvable ID.
5. Update dependency snapshots, retained owner fragments, profile contracts,
   integrity artifacts, generator/validator code, production catalog types,
   and tests. Delete owner manifests, shared-owner resolutions, conformance
   manifests, and contract closures because their remaining purpose is
   provenance or completeness accounting.
6. Bump every affected extension schema major, remove old versions, and
   regenerate policy-managed outputs. Do not hand-edit generated roots.
7. Refactor mixed extension tests to assert descriptor admission, dependency
   closure through contract IDs, runtime composition, collisions, schema
   behavior, security behavior, and generation integrity. Delete assertions
   over requirement/AC arrays, counts, or catalog hashes.

**Acceptance:**

- No extension authored, generated, runtime, or test artifact contains
  requirement-catalog references or specification requirement/AC fields.
- Typed contracts preserve every production-relevant executable fact.
- Extension behavior follows the revised adopted specification. Public
  descriptors remove requirement-derived provenance unless the adopted product
  contract gives it continuing value.

**Validation:**

- `make json-shape-check`
- `make generate`
- `make generate-drift`
- `make generated-artifact-policy-check`
- `make test-slice OWNER=module.extensions`
- `make test-slice OWNER=harness.generated_artifacts`

**Execution checkpoint:**

- The Extensions NLSpec now defines typed facts and behavioral verification
  without document hashes, requirement/acceptance arrays, status gates, or
  bidirectional traceability. Current behavior is `RETAIN-ROUTE`: descriptor
  admission, dependency and artifact resolution, authorization, redaction,
  codecs, errors, containment, state, participant, and generation behavior
  remain exercised. The two completeness-only tests and rows were removed;
  descriptor-provenance and contract-accounting tests were refactored into
  descriptor-projection and behavior-routing tests.
- Deleted the requirement mapping source, shared-owner resolution set,
  closure-mapping source, ten owner contract manifests, and their generated
  conformance, closure, requirement-coverage, and registry-accounting
  projections. These artifacts had no independent product behavior.
- Major-versioned `base_route_reservation_registry` v2→v3 and these Extension
  contracts: authored-input catalog v2→v3, dependency declaration v2→v3,
  dependency snapshot v1/v2→v3, owner fragment v2→v3, owner-input registry
  v1→v2, profile configuration v2→v3, profile descriptor v1/v2→v3, registry
  integrity v1→v2, validation condition registry v1→v2, validation surface
  declaration/set v2→v3, participant specialization v2→v3, transaction
  participant v2→v3, backup binding codec v2→v3, generated-schema source set
  v2→v3, contract-definition set v2→v3, and shared-protocol set v2→v3.
  Removed owner-contract-manifest v2, shared-owner-resolution v2,
  closure-derivation source v1, conformance manifest/index v1/v2,
  contract-closure catalog v1/v2, requirement mapping/coverage v2, and
  registry-accounting v1 identities without replacement.
- The final semantic scan is empty across authored Extension contracts,
  generated Go/TypeScript projections, runtime catalog types, generator code,
  and Extension test-family inputs. The only temporarily retained
  requirement-linked value is in the central verification-v2 owner contract;
  it is an explicitly enumerated `NS-06` atomic deletion target, not an
  Extension product or generation input.
- Official `make generate` produced all generated changes; no generated root
  was hand-edited. The evidence epoch did not change and no retained result was
  reused. The baseline worktree was clean at `286c91a1`; no unrelated
  pre-existing change was present.
- Initial generation failures identified omitted typed configuration fields,
  four rounded 64-bit limits, a stale participant digest, and misassigned codec
  hashes. They were corrected in authored inputs. The final major-version audit
  also corrected a stale operational configuration-schema reference.
  Harness validation initially found stale row/selector counts and a renamed
  duration baseline; the old weight was preserved under the behavior-routing
  symbol because partial slice evidence cannot refresh a full baseline.

### NS-04: Graph, Reporting, and Composition test cleanup

**Status:** `DONE`.

Before deleting accounting, revise the Graph Projection, Reporting, and Report
Composition specifications so behavioral scenarios remain normative but no
one-to-one fixture/test mapping is required. Record the human disposition of
each affected current obligation.

**Graph Projection actions:**

1. Delete the acceptance-coverage matrix, its schema/attachment, and the
   fixture-count conformance test.
2. Delete the redundant fixture corpus registry and its bespoke shape/count
   validation.
3. Replace `cartulary.graph_projection_fixture_manifest.v2` with v3, removing
   `requirement_ids`.
4. Preserve all 36 executable fixture directories, inputs, expected artifacts,
   digests, deterministic controls, state-effect checks, operations, test
   symbols, and real implementation adapters.
5. Update the fixture loader and its tests to validate behavior-bearing
   manifest fields without requiring traceability metadata.

**Reporting actions:**

1. Delete the 80-row status/ID-only corpus and
   `TestReportingTraceabilityAndFixtureCorpus`.
2. Preserve the supported output-kind assertion in an ordinary behavioral test
   that calls production code and expects Mermaid and Slidev.
3. Remove the fixture-corpus verification and test-family routing entry.

**Report Composition actions:**

1. Delete the 27-row status/ID-only corpus and
   `TestReportCompositionTraceabilityAndFixtureCorpus`.
2. Replace the fixture-corpus verification with a behavior verification and
   reroute the real preview-to-reporting integration row through it.
3. Remove only the traceability test row; preserve the real integration test.

**Acceptance:**

- Graph fixtures still execute real behavior and compare expected outputs.
- Reporting output-kind behavior remains covered.
- No test passes merely because an ID, count, or `implemented`/`planned` marker
  exists.

**Validation:**

- `make json-shape-check`
- `make test-slice OWNER=module.graphprojection`
- `make test-slice OWNER=module.reporting`
- `make test-slice OWNER=module.reportcomposition`
- `make harness-contract`

**Completion checkpoint (2026-07-26, baseline `286c91a1`):**

- The session began from the clean baseline recorded by `NS-00/R`; all dirty
  worktree entries at this checkpoint belong to `NS-01` through `NS-04`.
  No unrelated pre-existing change was present or overwritten.
- Graph Projection retained all 36 behavior-bearing fixture directories.
  Fixture manifests moved from v2 to v3 and dropped `requirement_ids`; the
  loader, schema attachment, shape validator, and tests now validate only
  executable fixture data. Aggregate unit and service-backed tests execute
  every fixture against its real adapter and compare complete expected
  response or state artifacts.
- `GP-FIX-005` had a stale issue reason exposed by the newly complete fixture
  route. The fixture now uses the adopted
  `empty_kind_registry_not_allowed` reason, and its reviewed expected response
  and digest were regenerated through the official candidate target.
- Reporting retained its production `supportedOutputKinds` behavior through a
  direct Mermaid/Slidev assertion. Report Composition retained its real
  preview-to-reporting integration row under
  `module.reportcomposition.verification.behavior_contract`.
- The Graph matrix and redundant corpus, Reporting and Composition
  ID/status-only corpora, and their completeness/trace tests were
  `SPEC-PRUNE`: each owning specification first removed the one-to-one
  accounting mandate. The 36 Graph fixtures, Reporting output kinds, and
  Composition integration behavior were `RETAIN-ROUTE`. No current product
  behavior was pruned or moved to a future profile.
- The two subsystem-wide Reporting and Composition requirement catalogs were
  deleted after their adopted specifications ceased to define those copied
  row identities. The temporary owner-level behavior requirements remain
  only for verification v2 and are enumerated `NS-06` deletion inputs.
- Final validations passed: `make generate`
  (`.cartulary/test-results/20260726T220022Z-p3037762`),
  `make json-shape-check` (`.../20260726T220133Z-p3041812`),
  `make generate-drift` (`...p3041811`),
  `make generated-artifact-policy-check` (`...p3041841`),
  `make harness-contract` (`...p3042076`), owner slices for Graph,
  Reporting, and Composition (`.../20260726T220209Z-p3047671`,
  `...p3047669`, and `...p3047688`), and their service-backed slices
  (`.../20260726T220233Z-p3051950`,
  `.../20260726T220254Z-p3053524`, and
  `.../20260726T220333Z-p3054798`). Duration-baseline coverage passed at
  `.../20260726T220020Z-p3037600`.
- Expected discovery failures were resolved and retained for diagnosis:
  stale generated topology (`.../20260726T215220Z-p3008800`), unsorted
  generated inputs (`.../20260726T215450Z-p3022366` and
  `.../20260726T215514Z-p3024944`), stale `GP-FIX-005` behavior
  (`.../20260726T215556Z-p3029616`), missing duration baseline
  (`.../20260726T215846Z-p3036418`), partial-run baseline refresh refusal
  (`.../20260726T215952Z-p3036961` and
  `.../20260726T215958Z-p3037208`), and stale harness cardinalities
  (`.../20260726T220032Z-p3040085`). No required validation was skipped.

### NS-04A: Network Flow accounting removal

**Status:** `DONE`.

**Actions:**

1. Revise the Network Flow and testing-harness fixture provisions first:
   retain behavioral scenarios and fixture integrity, but remove exact
   acceptance-to-selector accounting and copied owner text.
2. Delete `tools/network_flow_activity_accounting.json`, its schema,
   attachment, validator, and completeness-only tests.
3. Replace `cartulary.network_flow_fixture_manifest.v1` with v2. Remove
   acceptance IDs, verification IDs, execution/phase selectors, copied
   requirements, and document provenance from fixture and scenario shapes.
4. Preserve all 28 real fixture directories, sources, expected artifacts,
   transcripts, sizes, digests, freeze/revision semantics, containment,
   symlink rejection, and real execution.
5. Replace AC-ID switch helpers and machine-completeness assertions with direct
   behavior tests. Route useful tests through Network Flow behavior
   verification; remove the accounting-only verification.
6. Relocate independently useful generation activation, fixture-integrity,
   scratch-copy, and public-target checks to their owning validators without
   retaining AC maps or counts.
7. Remove `document_version` only where it denotes source-document provenance.
   Retain genuine runtime contract versions and product
   `conformance_status`.

**Acceptance:**

- No Network Flow executable input contains acceptance IDs, copied owner
  requirements, phase selectors, or accounting-only mappings.
- All real fixtures and direct unit/store/process/browser behavior remain
  executable.

**Validation:**

- `make json-shape-check`
- `make generate-drift`
- `make test-slice OWNER=module.networkflow`
- affected `web.networkflow` and generated-artifact owner slices
- `make harness-contract`

**Completion checkpoint (2026-07-26, baseline `286c91a1`):**

- Revised Network Flow `2.0.1` to `2.0.2` and the Testing Harness fixture
  provisions before changing executable contracts. The specifications now
  retain the behavioral scenarios, fixture integrity, and product
  conformance states while rejecting acceptance-to-selector accounting,
  copied owner text, and source-document provenance.
- Deleted the 107-row Network Flow activity-accounting manifest, its v2
  schema, attachment, validator, and completeness-only harness test. Removed
  the accounting-only Network Flow verification and its generic owner
  requirement.
- Migrated the contract index, fixture manifest, fixture scenario, mapping
  registry, presentation registry, frontend entrypoints, and timezone
  provenance to v2/v2/v2/v2/v2/v3/v2 respectively. Old readers and schemas
  were removed; no alias or dual-read path was added.
- Preserved all 28 fixture directories and recalculated their authored
  file/bundle digests after removing acceptance, verification, selector,
  copied-requirement, and document-provenance fields. The fixture validator
  still enforces containment, regular files, symlink rejection, exact
  authored-file closure, resource limits, sizes, hashes, scenario identity,
  and bundle integrity.
- Replaced AC-switch dispatch with cohesive direct behavior tests. The
  frozen-corpus test now executes every fixture through production parsing,
  mapping, canonicalization, validation, and request-admission paths while
  retaining complete expected/transcript integrity. The owner catalog
  removed 64 duplicate accounting wrappers and now routes 55 real rows.
- DoD disposition: all current Network Flow product behavior and all 28
  fixtures are `RETAIN-ROUTE`; no product behavior was pruned or moved.
  The static acceptance ledger, duplicate AC wrappers, selector switches,
  and document/catalog-completeness assertions are `SPEC-PRUNE` after the
  owning specification amendments. The remaining central planned
  `network_flow_activity` requirement owner is an enumerated `NS-06`
  deletion target and is not execution evidence.
- Final validation passed: `make test-slice OWNER=module.networkflow`
  (`.cartulary/test-results/20260726T222136Z-p3078146`);
  `make service-backed-test-slice OWNER=module.networkflow`
  (`.cartulary/test-results/20260726T222840Z-p3117785`, 20 tests);
  `make test-slice OWNER=web.networkflow`
  (`.cartulary/test-results/20260726T223312Z-p3151679`, 31 tests);
  `make test-slice OWNER=harness.generated_artifacts`
  (`.cartulary/test-results/20260726T223335Z-p3153375`);
  `make generate`
  (`.cartulary/test-results/20260726T223352Z-p3157404`);
  `make generate-drift`
  (`.cartulary/test-results/20260726T223400Z-p3159624`);
  `make generated-artifact-policy-check`
  (`.cartulary/test-results/20260726T223413Z-p3163426`);
  `make json-shape-check`
  (`.cartulary/test-results/20260726T223418Z-p3163808`); and
  `make harness-contract`
  (`.cartulary/test-results/20260726T223423Z-p3164410`). The tracker
  checkpoint passed `make lint-markdown`
  (`.cartulary/test-results/20260726T223545Z-p3166035`).
- Focused scans found no old Network Flow schema identity, activity-accounting
  reference, executable AC ID, phase selector, copied owner requirement, or
  document-provenance field. The only `acceptance_ids` matches are negative
  closed-shape harness fixtures that prove both new schemas reject the field.
- Expected discovery failures were resolved and retained: scenario schema
  identity (`.../20260726T221952Z-p3070433`), attachment ordering
  (`.../20260726T222027Z-p3071272`), stale generated topology
  (`.../20260726T222041Z-p3071865`), an over-broad manifest-version edit
  (`.../20260726T222058Z-p3074651`), and stale harness cardinalities
  (`.../20260726T222630Z-p3114745`). The final catalog has 868 rows and
  1,544 selectors. No required check was skipped, no generated root was
  hand-edited, and the clean starting worktree had no unrelated change.

### NS-04B: OTel derived conformance

**Status:** `DONE`.

**Actions:**

1. Revise the OTel specification provisions first so conformance is derived
   from real behavior and contract checks, not a checked-in pass declaration.
2. Delete `contracts/otel/conformance_status.json`, its v1 schema,
   attachment, loader, pass/claim decisions, and static acceptance rows.
3. Refactor `make otel-conformance` to aggregate actual source-snapshot,
   configuration, generated-constant, import-boundary, golden, runtime,
   redaction, failure-invariance, and shutdown checks.
4. Route `platform.telemetry` verification only through real tests or the real
   conformance target.

**Acceptance:**

- No checked-in value can self-assert OTel readiness or conformance.
- OTel failure, redaction, shutdown, and import-boundary behavior remains
  executable and release-visible.

**Validation:**

- `make otel-conformance`
- `make test-slice OWNER=platform.telemetry`
- `make json-shape-check`
- `make harness-contract`

**Completion checkpoint (2026-07-26, baseline `286c91a1`):**

- Added `OTEL-REQ-150` before implementation changes. The adopted OTel
  specification now requires current-run source, configuration, generated
  constant, import-boundary, browser-build, golden, runtime, privacy,
  failure-invariance, queue, retry, and shutdown evidence and explicitly
  rejects authored pass, release-readiness, acceptance-completion, and
  `claim_allowed` declarations.
- Deleted `contracts/otel/conformance_status.json` and removed its
  `cartulary.otel_conformance_status.v1` loader and all pass/claim/decision
  consumption. The repository had no status-schema file or attachment, so
  there was no hidden schema/attachment compatibility surface to retain.
  Core 04's threat-model hook now points only to the behavior fixture, real
  exporter tests, and `make otel-conformance`.
- Added active owner `platform.telemetry` with six cohesive rows covering all
  50 existing platform telemetry test selectors. The current verification now
  accepts both `go_test` and `static_check` evidence. The conformance checker
  executes that owner slice in the current run before evaluating snapshots,
  configuration, generated constants, import boundaries, browser artifacts,
  normalized goldens, and error mappings. The retained summary remains output,
  never an input to a later pass.
- Routing real tests exposed two hidden implementation gaps. `IMPLEMENT`:
  exact closed enum membership now admits the specified
  `authorization` error class without applying a generic secret-word
  heuristic, and digest-bound future profile identities validate canonical
  profile-token shape without a duplicated telemetry profile vocabulary.
- DoD disposition: the 50 behavior, privacy, redaction, exporter, retry,
  queue, shutdown, resource, and no-SDK selectors are `RETAIN-ROUTE`; the two
  uncovered defects are `IMPLEMENT`; the checked-in self-attestation and its
  loader are `SPEC-PRUNE`. No current product behavior was removed or moved
  to a future profile.
- Final validation passed: `make format`
  (`.cartulary/test-results/20260726T224056Z-p3174965`);
  `make test-slice OWNER=platform.telemetry`
  (`.cartulary/test-results/20260726T224105Z-p3177755`);
  `make otel-conformance`
  (`.cartulary/test-results/20260726T224114Z-p3178121`, including the nested
  current owner slice under the same run root); `make json-shape-check`
  (`.cartulary/test-results/20260726T224205Z-p3181760`);
  `make harness-contract`
  (`.cartulary/test-results/20260726T224246Z-p3183596`);
  `make generate-drift`
  (`.cartulary/test-results/20260726T224322Z-p3184676`); and
  `make generated-artifact-policy-check`
  (`.cartulary/test-results/20260726T224335Z-p3188489`). Initial generation
  after registering the owner passed at
  `.cartulary/test-results/20260726T223958Z-p3171667`; the tracker checkpoint
  passed `make lint-markdown` at
  `.cartulary/test-results/20260726T224434Z-p3189261`.
- Focused scans found the removed status identity/path and pass/claim fields
  only in the specification's prohibition and the negative harness regression
  that proves the file is absent and the checker has no loader. The catalog
  now has 55 owners, 194 families, 874 rows, and 1,594 exact selectors.
- Expected discovery failures were resolved and retained: the first newly
  routed owner run
  (`.cartulary/test-results/20260726T224009Z-p3173927`) exposed the two real
  implementation gaps above, and the initial harness run
  (`.cartulary/test-results/20260726T224213Z-p3182378`) exposed the intentional
  selector-cardinality increase. No required check was skipped, the evidence
  epoch did not change, no retained result was reused, no generated root was
  hand-edited, and the clean starting worktree had no unrelated change.

### NS-05: Repository-wide authority-consumer audit

**Status:** `DONE`.

Apply the Section 3 disposition matrix to all production, test, generator,
harness, schema, and release inputs. The audit must distinguish specification
requirements from legitimate operational resource fields such as
`service_requirements`.

The following are forbidden after the audit:

- paths under or references to `contracts/requirements`;
- requirement registry/catalog schema IDs, paths, hashes, loaders, or
  attachments;
- specification-oriented fields ending in `requirement_id`,
  `requirement_ids`, or `acceptance_criterion_ids`;
- specification-oriented `acceptance_ids`, `owner_requirements`,
  acceptance-to-selector maps, phase selectors, document
  status/version/path/hash/heading/byte references, and synthetic document
  contract IDs;
- checked-in coverage, pass, release-ready, or claim-allowed values that assert
  specification completeness instead of recording an actual run;
- the old digest names `catalog_semantic_digest` and
  `verification_semantic_digest`;
- executable or generated `spec_trace_ids`;
- executable references to `docs/`;
- tests whose only outcome is traceability, completeness, count, status, or
  document-shape accounting; and
- planned placeholders in active verification or test routing.

Allowed post-audit surfaces include typed contract IDs, versioned executable
facts, verification IDs, test row IDs, operational service requirements,
behavior fixture IDs, real contract versions, and runtime conformance fields
that affect product behavior.

**Acceptance:** A repository-wide search result is attached to the handoff with
every match classified as removed, renamed, or a documented operational
non-specification use.

**Validation:** Narrow owner slices for every changed owner, followed by
`make json-shape-check`, `make generate-drift`, and `make harness-contract`.

**Completion checkpoint (2026-07-26, baseline `286c91a1`):**

- The audit removed three previously unenumerated provenance surfaces. OpenAPI
  release approvals now use
  `cartulary.openapi_release_change_set.v2`, which contains only the semantic
  fingerprint, compatibility class, owner, and rationale. OTel configuration,
  hazard, corpus, normalized-signal, dependency-classification, raw-capture,
  and generated-constant contracts moved to v2 identities and no longer carry
  copied owner prose, static conformance labels, or the unused
  `repo_materialized_seed` status. The draft Reference Pack NLSpec moved from
  document version `0.1.0` to `0.1.1` and explicitly excludes specification
  requirement identifiers from public validation diagnostics.
- The residual requirement search is an exact `NS-06` atomic deletion set:
  71 owner catalogs under `contracts/requirements/`, 60 verification v2 owner
  contracts, the requirement and verification loaders, their two v1/v2
  schemas, and the central negative/coverage tests. There are no
  `owner_requirements`, `spec_trace_ids`, phase selectors, or planned entries
  in active verification/test routing. The one `acceptance_ids` file is the
  negative closed-schema regression for the removed Network Flow field.
- The two old digest names occur in one 23-file harness/evidence union: 14
  writers, readers, finalizers, diagnostics, and tests plus the nine schemas
  enumerated by `NS-06`. These are kept together until the atomic epoch/schema
  cutover; no additional serializer was discovered.
- The four document-field matches belong solely to the immutable OpenAPI
  release registry and compatibility loader. Its `document_path`, SHA-256,
  byte length, source commit, and publication state identify a versioned
  machine OpenAPI release under `contracts/openapi-releases/`, not a
  specification document. OTel source-snapshot `present`/`absent` values are
  operational third-party input inventory. Product/runtime statuses, service
  requirements, verification/test identities, fixture identities, release
  lifecycle state, and actual run statuses remain legitimate operational
  concepts.
- No executable JSON, production source, generator, conformance input, or
  release input references `docs/`. The only search matches are human-facing
  Markdown pointers. `claim_allowed` and release-ready vocabulary occurs only
  in the OTel negative regression; `coverage_status` occurs only in shell
  variables holding command exit codes.
- `make format` passed at
  `.cartulary/test-results/20260726T225245Z-p3217267`; final
  `make generate` and `make otel-conformance` passed at
  `.cartulary/test-results/20260726T225249Z-p3220016` and
  `.cartulary/test-results/20260726T225255Z-p3222171`. The telemetry, OpenAPI,
  and generated-artifact owner slices passed at
  `.cartulary/test-results/20260726T225052Z-p3204020`,
  `.cartulary/test-results/20260726T225055Z-p3204336`, and
  `.cartulary/test-results/20260726T225102Z-p3206688`.
  `make json-shape-check`, `make generate-drift`,
  `make generated-artifact-policy-check`, and `make harness-contract` passed
  at `.cartulary/test-results/20260726T225120Z-p3210778`,
  `.cartulary/test-results/20260726T225122Z-p3211272`,
  `.cartulary/test-results/20260726T225134Z-p3215046`, and
  `.cartulary/test-results/20260726T225306Z-p3225692`.
- All removed metadata was `SPEC-PRUNE`; executable OTel and OpenAPI behavior
  was `RETAIN-ROUTE`. No generated root was hand-edited, no required check was
  skipped, no retained result was reused, the evidence epoch has not yet
  changed, and the clean starting baseline had no unrelated pre-existing
  worktree change.

### NS-06: Verification v3, evidence epoch, and requirement-layer deletion

**Status:** `DONE`.

This is the atomic central cutover. It is complete only when all changes below
land together.

**Verification changes:**

1. Replace verification registry/contract v2 with v3.
2. Keep `verification_id`, `behavior_class`, `profile`, `evidence_kinds`,
   optional `public_target`, and skip policy.
3. Remove `requirement_ids`, `spec_trace_ids`, and redundant verification and
   registry-owner status. Presence in v3 defines the current routing set.
4. Remove the requirement loader, requirement-to-verification map, active or
   planned requirement checks, and requirement completeness enforcement.
5. Retain routing integrity: unique owner-qualified verification IDs, valid
   profiles/evidence kinds/skip policies, valid row references, and the rule
   that each verification has a test row or declared public target.
6. Compute `verification_routing_digest` from routing-semantic fields only,
   with no specification-provenance input.

**Requirement-layer deletion:**

1. Delete `contracts/requirements/**`.
2. Delete requirement registry/catalog v1 schemas and schema attachments.
3. Delete all loader, shape-check, test, generator, and release code that
   expects the requirement layer.
4. Do not replace it with another statement/status catalog.

**Evidence interface changes:**

Replace fields everywhere:

| Old field | New field |
| --- | --- |
| `catalog_semantic_digest` | `test_catalog_digest` |
| `verification_semantic_digest` | `verification_routing_digest` |

Major-version and update every schema, writer, reader, equality check, fixture,
test, diagnostic, browser planner, scheduler, finalizer, retained-run
validator, and release consumer carrying those fields:

| Current schema | Target schema |
| --- | --- |
| `cartulary.browser_owner_index.v1` | `cartulary.browser_owner_index.v2` |
| `cartulary.task_guide_summary.v2` | `cartulary.task_guide_summary.v3` |
| `cartulary.test_catalog_check_summary.v1` | `cartulary.test_catalog_check_summary.v2` |
| `cartulary.test_evidence_accounting.v1` | `cartulary.test_evidence_accounting.v2` |
| `cartulary.test_evidence_audit_summary.v1` | `cartulary.test_evidence_audit_summary.v2` |
| `cartulary.test_owner_explanation.v1` | `cartulary.test_owner_explanation.v2` |
| `cartulary.test_owner_summary.v1` | `cartulary.test_owner_summary.v2` |
| `cartulary.test_slice_plan.v2` | `cartulary.test_slice_plan.v3` |
| `cartulary.test_slice_scheduler_summary.v1` | `cartulary.test_slice_scheduler_summary.v2` |

All listed target schemas require
`evidence_epoch="cartulary.test_evidence.nlspec.v1"`. Any additional schema
found to serialize either old digest must receive the same next-major treatment
and be added to this table before `NS-06` is marked `DONE`.

**Required regression tests:**

1. Verification v3 rejects requirement, trace, and redundant status fields.
2. Specification-only changes do not change `verification_routing_digest`,
   test selection, or evidence identity.
3. An invalid v3 verification routing field fails before test execution.
4. A row referencing an unknown verification still fails routing validation.
5. Missing or mismatched evidence epochs prevent retained-run reuse.
6. Pre-cutover schemas are rejected; there are no compatibility readers.
7. The executable-input boundary rejects a synthetic machine input that
   references `docs/` without reading the referenced document.
8. Markdown-only specification changes do not alter test or verification
   digests.

**Acceptance:**

- The requirement layer no longer exists.
- Verification v3 and all evidence consumers use only routing semantics.
- No current gate reads or accepts pre-cutover evidence.
- `make harness-contract` passes without a machine requirement catalog.

**Completion checkpoint (2026-07-26, baseline `286c91a1`):**

- Verification registry and all 60 owner contracts moved from v2 to
  routing-only v3. Registry owners now contain only `owner_id` and
  `contract_path`; verification entries retain routing and skip-policy
  semantics and reject `requirement_ids`, `spec_trace_ids`, and redundant
  status fields. The routing loader validates local identity, shape,
  qualification, profile, evidence-kind, skip-policy, row, and public-target
  integrity without importing a requirement loader.
- The complete 72-file `contracts/requirements/` tree, its two v1 schemas,
  schema attachments, loader, exports, and completeness tests were deleted.
  No replacement statement/status catalog, compatibility reader, alias, dual
  writer, or historical evidence migration was added. These accounting
  artifacts were `SPEC-PRUNE`; routing and observable harness behavior were
  `RETAIN-ROUTE`.
- Evidence fields changed atomically from `catalog_semantic_digest` and
  `verification_semantic_digest` to `test_catalog_digest` and
  `verification_routing_digest`. The nine planned schema majors were the
  complete serializer set: browser owner index v1→v2, task-guide summary
  v2→v3, catalog-check summary v1→v2, evidence accounting v1→v2, evidence
  audit summary v1→v2, owner explanation v1→v2, owner summary v1→v2, slice
  plan v2→v3, and slice scheduler summary v1→v2. All require the exact
  `cartulary.test_evidence.nlspec.v1` epoch; no additional serializer was
  discovered.
- Catalog loading now resolves command identities from the authored
  `tools/task_surface_owner.json`, keeping test-catalog validation upstream of
  generated topology. Three owner-routable behavior families were added for
  `harness.test_catalog`, `harness.evidence_accounting`, and
  `harness.release`; their Make-owned internal targets exercise catalog,
  epoch/accounting, and release-evidence behavior rather than placeholder
  counts. The current catalog contains 58 owners, 197 families, 877 rows, and
  1,597 exact selectors.
- New negative regressions reject v3 registry status, verification requirement
  IDs, trace IDs, and verification status. Evidence audit regressions reject
  both a pre-cutover v1 artifact with no epoch and a current-shape artifact
  with the wrong epoch. Release readiness now verifies epoch, test-catalog
  digest, verification-routing digest, profile, run, and owner-partition
  identity against current routing before accepting evidence.
- Final owner slices passed for `harness.test_catalog`,
  `harness.evidence_accounting`, `harness.browser`, `harness.release`, and
  `harness.generated_artifacts` at
  `.cartulary/test-results/20260726T230821Z-p3265192`,
  `.cartulary/test-results/20260726T230920Z-p3271149`,
  `.cartulary/test-results/20260726T230932Z-p3271787`,
  `.cartulary/test-results/20260726T230848Z-p3266293`, and
  `.cartulary/test-results/20260726T230904Z-p3267098`.
  `make format`, `make json-shape-check`, `make generate-drift`,
  `make generated-artifact-policy-check`, and `make harness-contract` passed
  at `.cartulary/test-results/20260726T231028Z-p3274758`,
  `.cartulary/test-results/20260726T231031Z-p3277495`,
  `.cartulary/test-results/20260726T231034Z-p3278003`,
  `.cartulary/test-results/20260726T231045Z-p3281758`, and
  `.cartulary/test-results/20260726T231058Z-p3282282`. The final generation
  root is `.cartulary/test-results/20260726T230811Z-p3262929`.
- Discovery failures were retained and resolved: owner routing initially
  rejected three unroutable owners at
  `.cartulary/test-results/20260726T225909Z-p3236084`; generation then exposed
  missing command registration and a public-target ownership mismatch at
  `.cartulary/test-results/20260726T230337Z-p3247890`,
  `.cartulary/test-results/20260726T230442Z-p3250776`, and
  `.cartulary/test-results/20260726T230559Z-p3253792`; release validation
  exposed an unintended release partition at
  `.cartulary/test-results/20260726T230742Z-p3261921`. The authored task
  surface and internal owner routing fixes above resolved each failure.
- A final executable search finds old requirement paths/loaders/schema IDs,
  old digest fields, and removed v1 evidence identities only in deliberate
  negative regressions. The clean starting baseline had no unrelated
  pre-existing change; generated roots changed only through official
  generation; no required validation was skipped; and no retained result was
  reused. This slice establishes, but does not yet retain, the new evidence
  epoch. The completed `NS-06` tracker checkpoint passed
  `make lint-markdown` at
  `.cartulary/test-results/20260726T231304Z-p3283967`.

**Reopened completion checkpoint (2026-07-26, baseline `286c91a1`):**

- `agent-finalize` against the initial `NS-07` warm root failed at
  `.cartulary/test-results/20260726T232623Z-p3577259` because the required
  tracker checkpoint changed `source_snapshot_digest`. Investigation found
  that Testing Harness still specified both pre-cutover digest names and
  `source_snapshot_digest_v1` over every repository file, contradicting its
  no-document executable boundary and the program's required Markdown
  invariance. The retained root was rejected and no finalizer mutation ran.
- Testing Harness now specifies `evidence_epoch`, `test_catalog_digest`,
  `verification_routing_digest`, and `source_snapshot_digest_v2`.
  Source-snapshot v2 derives excluded roots from the typed executable-input
  policy, excludes all Markdown by extension, and filters normalized paths
  before any filesystem inspection or hashing. Editing, adding, deleting,
  renaming, or rearranging specifications and Markdown guidance cannot change
  the digest or file count.
- The new regression covers Markdown inside and outside `docs/`, non-Markdown
  specification material inside `docs/`, a missing-target specification
  symlink, and rearranged/deleted guidance. It also proves a non-document
  source edit still changes the digest. The final executable scan contains
  only the deliberate `spec_trace_ids` v3 rejection fixture.
- `make format` and `make lint-markdown` passed at
  `.cartulary/test-results/20260726T233100Z-p3594721` and
  `.cartulary/test-results/20260726T232926Z-p3583210`. Official generation
  passed at `.cartulary/test-results/20260726T233145Z-p3599204`; drift,
  generated policy, shape, and harness passed at
  `.cartulary/test-results/20260726T233157Z-p3601626`,
  `.cartulary/test-results/20260726T233157Z-p3601611`,
  `.cartulary/test-results/20260726T233157Z-p3601631`, and
  `.cartulary/test-results/20260726T233157Z-p3601897`.
- Final owner slices passed for `harness.test_catalog`,
  `harness.evidence_accounting`, `harness.browser`, `harness.release`, and
  `harness.generated_artifacts` at
  `.cartulary/test-results/20260726T233236Z-p3607574`,
  `.cartulary/test-results/20260726T233236Z-p3607562`,
  `.cartulary/test-results/20260726T233236Z-p3607603`,
  `.cartulary/test-results/20260726T233236Z-p3607594`, and
  `.cartulary/test-results/20260726T233236Z-p3607613`. Their source snapshot is
  `sha256:8ec22bab12a258a175a07c12a14b2b069eb97b3588ac0a8fea7abefaa2edfaa7`;
  catalog and routing digests are unchanged.
- The expected pre-generation shape failure at
  `.cartulary/test-results/20260726T233108Z-p3597533` reported only stale
  topology input and passed after official generation. No schema attachment
  changed, no policy-managed output was hand-edited, no validation was
  skipped, no unrelated pre-existing change was present, and no retained
  result was reused. The evidence-identity correction is `IMPLEMENT`;
  documentation remains `EDITORIAL_ONLY`.
  The reopened completion checkpoint passed `make lint-markdown` at
  `.cartulary/test-results/20260726T233333Z-p3615753`.

### NS-07: Generation, drift closure, and evidence rebaseline

**Status:** `DONE`.

**Actions:**

1. Run official generation after all authored schemas, contracts, manifests,
   and topology inputs are final.
2. Inspect generated Go and TypeScript diffs for traceability leakage and
   unintended public product changes.
3. Rebuild schema attachments and generated topology through their owners.
4. Create the first retained evidence root for
   `cartulary.test_evidence.nlspec.v1`.
5. Verify old requirement, digest, traceability, and schema tokens are absent
   from executable and generated inputs.
6. Record all changed schema IDs and the new run root in the handoff log.

**Acceptance:** Generation and drift checks pass, the new evidence root is
internally consistent, and the tracked worktree contains no unexplained
generated diff.

**Validation:**

- `make generate`
- `make generate-drift`
- `make generated-artifact-policy-check`
- `make json-shape-check`
- `make harness-contract`

**Completion checkpoint (2026-07-26, baseline `286c91a1`):**

- Official `make generate` passed at
  `.cartulary/test-results/20260726T231339Z-p3287117`. Inspection of every
  changed policy-managed output found only the intended Extension and Network
  Flow contract-major/provenance removals, their derived digest changes,
  verification-v3 input hashes, and the three new internal harness routes.
  No generated file was hand-edited and `git diff --check` is clean.
- `make generate-drift`, `make generated-artifact-policy-check`,
  `make json-shape-check`, and `make harness-contract` passed at
  `.cartulary/test-results/20260726T231405Z-p3289677`,
  `.cartulary/test-results/20260726T231405Z-p3289655`,
  `.cartulary/test-results/20260726T231405Z-p3289646`, and
  `.cartulary/test-results/20260726T231406Z-p3290007`. After the discoveries
  below, the affected harness contract passed again at
  `.cartulary/test-results/20260726T232155Z-p3473577`, and the final full
  check reran generation, drift, shape, policy, and affected behavior in one
  successful schedule.
- The first warm run at
  `.cartulary/test-results/20260726T231447Z-p3295547` failed
  `go-vulncheck` because the earlier Extension cleanup removed `os` and
  `path/filepath` imports still used by independent contract-generator
  security tests. It also exposed a WorkbookShell test race: the party-link
  tests could open an inspector before the observable row load completed,
  after which the data reset closed it. The imports were restored, and both
  tests now wait for their production-facing row state before opening the
  inspector instead of extending a timeout. `go-vulncheck`, the generated
  owner slice, the focused WorkbookShell row, and the complete frontend unit
  suite passed at `.cartulary/test-results/20260726T231634Z-p3353421`,
  `.cartulary/test-results/20260726T231634Z-p3353193`,
  `.cartulary/test-results/20260726T231730Z-p3360390`, and
  `.cartulary/test-results/20260726T231810Z-p3363890`.
- The second warm run at
  `.cartulary/test-results/20260726T231837Z-p3365788` passed 163 of 164
  scheduled units and exposed two OTel cutover gaps. The nested telemetry
  behavior slice now carries the complete prepared `test-slice` artifact
  identity during scheduled execution, and the conformance checker treats
  verification-v3 presence rather than a removed `status` as current routing.
  Standalone OTel conformance and its six-row owner slice passed at
  `.cartulary/test-results/20260726T232155Z-p3473246` and
  `.cartulary/test-results/20260726T232155Z-p3473395`.
- The first internally consistent new-epoch warm run is
  `.cartulary/test-results/20260726T232236Z-p3478312`: `make check` passed all
  164 work units and 689 tests with no missing or failed result. All 252
  epoch-bearing artifacts use
  `cartulary.test_evidence.nlspec.v1`,
  `test_catalog_digest=sha256:1abae38334f41f22892be86b6fce50075d8f1b6d57b66350db312d1a98c0a921`,
  and
  `verification_routing_digest=sha256:6e2c5117d6cea5d79a2189a34d4935c8eea7118b41d865628a2784d657ebd002`.
  The root contains zero old digest, requirement path, or pre-cutover
  evidence-schema matches.
- The final executable-input scan contains only the negative regression that
  injects `spec_trace_ids` to prove v3 rejection. There are no unexplained
  generated diffs, no schema additions beyond the `NS-06` ledger, no skipped
  checks, no unrelated pre-existing worktree changes, and no retained result
  was consumed during this slice. The successful warm root is the retained
  evidence candidate for `NS-08` finalization. The completed tracker
  checkpoint passed `make lint-markdown` at
  `.cartulary/test-results/20260726T232546Z-p3574153`.

**Rebaseline checkpoint after reopened `NS-06` (2026-07-26):**

- Final official generation passed at
  `.cartulary/test-results/20260726T233408Z-p3618818`. Drift, generated policy,
  shape, and harness passed at
  `.cartulary/test-results/20260726T233420Z-p3621161`,
  `.cartulary/test-results/20260726T233420Z-p3621153`,
  `.cartulary/test-results/20260726T233420Z-p3621194`, and
  `.cartulary/test-results/20260726T233420Z-p3621496`. Generated diffs remain
  the explained Extension, Network Flow, verification-v3, and topology
  projections; no generated root was hand-edited.
- The replacement warm root is
  `.cartulary/test-results/20260726T233503Z-p3627091`: `make check` passed all
  164 work units and 689 tests with no failure, missing result, or unmapped
  evidence. All 252 evidence artifacts share
  `source_snapshot_digest=sha256:8ec22bab12a258a175a07c12a14b2b069eb97b3588ac0a8fea7abefaa2edfaa7`,
  the exact `cartulary.test_evidence.nlspec.v1` epoch, and the catalog/routing
  digests recorded above. The root contains zero pre-cutover schema, digest,
  or requirement-path matches.
- No schema or product behavior changed in this rebaseline, no check was
  skipped, and no retained result was consumed. The program worktree was
  already dirty with only this effort; no unrelated pre-existing change was
  present. This root supersedes the rejected initial candidate and is the sole
  `NS-08` finalizer input. The rebaseline checkpoint passed
  `make lint-markdown` at
  `.cartulary/test-results/20260726T233809Z-p3722714`.

**Timing-health replacement checkpoint (2026-07-26):**

- A fresh `make check` passed all 164 work units and 689 tests at
  `.cartulary/test-results/20260726T233928Z-p3728219`. The exact
  finalizer-equivalent timing gate,
  `make scheduler-summary-timing-drift RESULTS_DIR=<root> TARGET=check`,
  passed at `.cartulary/test-results/20260726T234226Z-p3822713`.
- An unscoped diagnostic invocation at
  `.cartulary/test-results/20260726T234159Z-p3822104` correctly audited every
  scheduler directory and reported that the nested OTel owner `test-slice`
  does not emit a top-level scheduler target summary. Finalization scopes
  timing health to the retained `check` scheduler; the nested owner slice
  retains its own scheduler summary and is not a top-level check target. No
  code or threshold changed in response to the broader diagnostic.
- This root supersedes the timing-contaminated candidate. No generated,
  schema, specification, implementation, or test file changed during the
  timing-health retry; no validation was skipped and no retained result was
  reused. The checkpoint passed `make lint-markdown` at
  `.cartulary/test-results/20260726T234251Z-p3823040`.

### NS-08: Full validation and final handoff

**Status:** `DONE`.

Run the narrowest affected-owner checks first. Before broad end-of-run
verification, run `make agent-finalize`; pass `RESULTS_DIR` when a successful
full warm-check run exists, otherwise record that retained-run maintenance was
skipped because `RESULTS_DIR` was unset.

Required final commands:

1. `make lint-markdown`
2. `make test-fast`
3. `make check`
4. `make ci`
5. `make release-check`

The handoff must record each command, result, run root or summary artifact,
failure relationship, and any skipped check with its reason.

**Closure checkpoint (2026-07-26, baseline `286c91a1`):**

- `make agent-finalize
  RESULTS_DIR=.cartulary/test-results/20260726T233928Z-p3728219` passed at
  `.cartulary/test-results/20260726T234327Z-p3826127`. The retained root passed
  epoch, source-snapshot-v2, routing, accounting, and exact `check` timing
  compatibility. The authorized finalizer refreshed
  `tools/browser_e2e_duration_baselines.json`,
  `tools/harness_smoke_duration_baselines.json`, and
  `tools/service_backed_make_target_duration_baselines.json`; its generation,
  drift, shape, coverage, timing, and rollback checks all passed.
- The required final sequence passed without interruption:
  `make lint-markdown` at
  `.cartulary/test-results/20260726T234416Z-p3833125`,
  `make test-fast` at
  `.cartulary/test-results/20260726T234422Z-p3834590` with 2 of 2 work units
  and 833 tests,
  `make check` at
  `.cartulary/test-results/20260726T234646Z-p3883854` with 164 of 164 work
  units and 689 tests,
  `make ci` at
  `.cartulary/test-results/20260726T234921Z-p3978073` with 5 of 5 work units
  and 689 tests, and
  `make release-check` at
  `.cartulary/test-results/20260726T235217Z-p4084990` with 12 of 12 work units
  and 689 tests. There were no failed, missing, or unmapped results.
- The final executable-input audit has one classified match: the negative
  verification-v3 regression deliberately injects `spec_trace_ids` and
  asserts rejection. There is no requirement tree or v2
  verification/requirement schema, both new digest names are active, and
  `git diff --check` passes. Contract-family, OpenAPI, Graph, Network Flow,
  Extensions, OTel, verification, and evidence migrations remain exactly as
  recorded by their owning slices.
- No semantic, schema, specification, implementation, or product-test change
  was introduced in `NS-08`. Generated policy roots changed only through
  earlier official generation; finalization changed only its three declared
  authored duration-baseline outputs. The retained result was consumed only
  by the explicit compatible finalizer invocation. No required check was
  skipped, and the initially clean worktree contains no unrelated
  pre-existing change.
- The completed closure checkpoint passed `make lint-markdown` at
  `.cartulary/test-results/20260726T235743Z-p35990`.

## 7. Public and Internal Interface Migration

| Surface | Breaking change | Compatibility rule |
| --- | --- | --- |
| Requirement registry/catalog v1 | Removed | No replacement statement/status catalog |
| Verification registry/contract v2 | Replaced by routing-only v3 without requirement, trace, or redundant status fields | v2 rejected; no dual reader |
| Contract-family registry v2 | Replaced by v3 without either owner ID array | v2 rejected |
| OpenAPI source manifest v2 | Replaced by v3 without `requirements_registry` | v2 rejected |
| Graph fixture manifest v2 | Replaced by v3 without `requirement_ids` | v2 rejected |
| Network Flow fixture manifest v1 | Replaced by v2 without acceptance, verification, selector, copied-requirement, or document-provenance fields | v1 rejected |
| Extension contracts | Requirement/AC/catalog fields removed; operational references become typed contract IDs | Affected schemas receive new majors; no aliases |
| Test/evidence digests | Renamed to test-catalog and verification-routing digests | Old fields rejected |
| Test evidence | Requires `cartulary.test_evidence.nlspec.v1` epoch | Earlier evidence cannot close current gates |
| Reporting/composition trace corpora | Removed | Specifications retain future requirements; no placeholder replacement |
| Graph conformance matrix/corpus registry | Removed | Executable fixture directories and behavior tests remain |
| Network Flow activity accounting | Removed | Behavior fixtures and owner-specific integrity checks remain |
| OTel static conformance status v1 | Removed | Conformance is derived only from actual checks and run evidence |

Public HTTP, WebSocket, storage, authorization, domain, and browser behavior
follows the final adopted owner specifications. If a provenance field leaks
through a public descriptor without continuing product value, remove it in the
owning slice. Any other externally meaningful behavior change starts with the
owning specification and is recorded in the obligation disposition.

## 8. Verification Strategy

### Per-slice owner routing

Before running an owner slice, use:

`make task-guide ROLE=module-author OWNER=<owner-id>`

The expected narrow owner set is:

- `harness.generated_artifacts`
- `platform.openapi`
- `module.extensions`
- `module.graphprojection`
- `module.reporting`
- `module.reportcomposition`
- `module.networkflow`
- `platform.telemetry`

Use `make explain-test-owner`, `make explain-target`, and `make explain-run`
when failures or execution scope are unclear.

### Required behavioral scenarios

| Scenario | Expected result |
| --- | --- |
| Specification prose, headings, formatting, or file arrangement changes | Product/harness execution and evidence identity are unchanged; only editorial tooling may react |
| Executable source or machine input references `docs/` | Restricted-input architecture check fails without opening or hashing the document |
| Versioned machine artifact has invalid shape or unknown schema identity | Shape validation fails before product assertions |
| Typed contract changes a value consumed by production/tests | Generation/drift and the relevant behavior tests reflect the contract change |
| Verification v3 contains a requirement, trace, or redundant status field | Schema/routing validation rejects it |
| Valuable current specification behavior has no implementation | It is implemented and behavior-tested before the owning slice completes |
| Behavior does not belong in the current future-facing profile | The owning specification is revised before its executable placeholder is removed |
| Graph fixture executes | Real code runs and observable artifacts/state effects are compared |
| Network Flow fixture executes | Real code runs and complete expected artifacts, state, and digests are compared without acceptance metadata |
| Reporting supported output kinds are queried | Production code returns Mermaid and Slidev |
| OTel conformance executes | Actual source/config/import/golden/runtime checks decide the result; no status file is consumed |
| Pre-cutover retained evidence is supplied to a current gate | Epoch/schema mismatch rejects reuse |
| Requirement catalog, requirement statement, or AC completeness input is added | Schema/boundary/audit checks reject it |

### Documentation-only validation

`make lint-markdown` maintains Markdown quality. It is not product,
conformance, generated-artifact, or release evidence.

## 9. Top-Level Work Tracker

| ID | Work item | Status | Depends on | Owner | Evidence | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| `NT-001` | Establish authority, scope, baseline, and cutover doctrine | `DONE` | None | Handoff owner | Sections 1–5 | Decisions and quantitative baseline are recorded |
| `NT-002` | Repair specification and guidance precedence | `DONE` | `NT-001` | Specification owners | `NS-01` diff and Markdown run | All authority prose is specification-first |
| `NT-003` | Decouple contract family and OpenAPI assembly | `DONE` | `NT-002` | Contract/OpenAPI owners | `NS-02` schemas, tests, generation | No requirement registry dependency remains |
| `NT-004` | Replace extension requirement traceability with typed contracts | `DONE` | `NT-002`, `NT-003` | `module.extensions` | `NS-03` authored/generated inventory | No requirement/AC/catalog field remains in Extensions |
| `NT-005` | Remove Graph, Reporting, and Composition accounting-only tests | `DONE` | `NT-004` | Subsystem owners | `NS-04` behavior test results | Valuable behavior remains; traceability-only artifacts are gone |
| `NT-006` | Remove Network Flow acceptance accounting | `DONE` | `NT-005` | `module.networkflow` | `NS-04A` fixture and behavior results | Fixtures execute without acceptance metadata |
| `NT-007` | Replace OTel self-attestation with derived conformance | `DONE` | `NT-006` | `platform.telemetry` | `NS-04B` real conformance results | No checked-in pass/claim input remains |
| `NT-008` | Complete repository-wide consumer audit | `DONE` | `NT-007` | Harness maintainer | Classified search inventory | Every forbidden match is removed or classified |
| `NT-009` | Execute central verification/evidence cutover and delete requirements | `DONE` | `NT-008` | Harness catalog/evidence owners | `NS-06` schema and harness results | v3 routing and new epoch pass without requirement catalogs |
| `NT-010` | Regenerate and establish first new-epoch evidence | `DONE` | `NT-009` | Generated-artifact owners | `NS-07` drift and retained run root | Generated state is clean and current |
| `NT-011` | Run full closure validation and hand off | `DONE` | `NT-010` | Cross-owner verifier | `NS-08` command/run ledger | All binary completion criteria pass |

## 10. Session Handoff Log

Append one row after every implementation session. Do not rewrite earlier rows
except to correct a factual error.

| Timestamp | Baseline commit | Slice | Status transition | Files/contracts changed | Commands and results | Run roots/artifacts | Generated outputs | Blockers/failures | Next unblocked slice |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-26 | `b45905b1` | `NS-00` | `TODO` → `DONE` | Added this tracker only | Planning baselines in Section 2; `make lint-markdown` passed | Section 2 roots and `.cartulary/test-results/20260726T205341Z-p2826022` | None | None | `NS-01` |
| 2026-07-26 | `286c91a1` | `NS-00/R` | plan reconciliation; `NS-01` → `IN_PROGRESS` | Tracker authority, decisions, gap register, serial slices, Network Flow/OTel scope, interfaces, risks, and exits | `make lint-markdown` passed | `.cartulary/test-results/20260726T211151Z-p2834566` | None | Initial worktree clean; no pre-existing changes; new evidence not created | `NS-01` active |
| 2026-07-26 | `286c91a1` | `NS-01` | `IN_PROGRESS` → `DONE`; `NS-02` → `IN_PROGRESS` | Testing Harness/Core 00 authority; `AGENTS.md`; README; Domain editorial projection boundary; Appendix I and developer guidance | `make lint-markdown` passed | `.cartulary/test-results/20260726T211514Z-p2839390` | None | No blockers; documentation-only authority repair; product checks skipped by slice design | `NS-02` active |
| 2026-07-26 | `286c91a1` | `NS-02` | `IN_PROGRESS` → `DONE`; `NS-03` → `IN_PROGRESS` | Contract-family registry v2→v3 removed both owner arrays; OpenAPI source manifest v2→v3 removed requirement registry; local owner/path/security validation and tests | `make generate`, `json-shape-check`, both owner slices, `generate-drift`, generated policy, and `harness-contract` passed | Generate `.cartulary/test-results/20260726T211817Z-p2844975`; shape `.cartulary/test-results/20260726T211833Z-p2847429`; owner slices `.../20260726T211839Z-p2848054` and `...p2848060`; drift `.../20260726T211857Z-p2853013`; policy `...p2853016` | `tools/execution_topology_render_index.json` refreshed only through `make generate`; OpenAPI bytes unchanged | Initial shape root `.cartulary/test-results/20260726T211809Z-p2844433` stopped on expected stale generated topology input and passed after generation | `NS-03` active |
| 2026-07-26 | `286c91a1` | `NS-03` | `IN_PROGRESS` → `DONE`; `NS-04` → `IN_PROGRESS` | Revised Extensions NLSpec; removed provenance/accounting sources and outputs; major-versioned typed Extension contracts; refactored descriptor/routing tests; preserved behavioral and security coverage | Final `generate`, shape, both owner slices, drift, generated policy, and `harness-contract` passed | Generate `.cartulary/test-results/20260726T214512Z-p2973651`; shape `.../20260726T214522Z-p2976091`; owner slices `...p2976267` and `...p2976197`; drift `...p2976083`; policy `...p2976148` | Go/TypeScript Extension projections and topology changed only through official generation | Initial generate roots `.../20260726T212659Z-p2863995`, `...T212735Z-p2867986`, `...T212803Z-p2869546`, and `...T212848Z-p2871239` exposed authored-input issues; harness root `.../20260726T213810Z-p2927146` exposed stale counts/baseline; all resolved. Baseline refresh targets `...T213929Z-p2929452` and `...T213935Z-p2929705` correctly rejected partial evidence. No epoch change, retained reuse, skipped required check, or unrelated pre-existing change | `NS-04` active |
| 2026-07-26 | `286c91a1` | `NS-04` | `IN_PROGRESS` → `DONE`; `NS-04A` → `IN_PROGRESS` | Revised Graph, Reporting, and Composition NLSpecs; Graph fixture v2→v3; deleted accounting-only matrices/corpora/tests and copied subsystem catalogs; retained 36 real fixtures, Reporting output kinds, and Composition preview integration | Final generation, shape, drift, generated policy, harness, duration coverage, all three owner slices, and all three service-backed slices passed | Generate `.cartulary/test-results/20260726T220022Z-p3037762`; shape/drift/policy/harness `.../20260726T220133Z-p3041812`, `...p3041811`, `...p3041841`, `...p3042076`; owner roots `.../20260726T220209Z-p3047671`, `...p3047669`, `...p3047688`; service roots `.../20260726T220233Z-p3051950`, `.../20260726T220254Z-p3053524`, `.../20260726T220333Z-p3054798` | Official generation refreshed topology; Graph expected candidate/digest refreshed through the public candidate target; no generated root was hand-edited | Clean starting baseline; no unrelated pre-existing change. Accounting-only rows were `SPEC-PRUNE`; all product behavior was `RETAIN-ROUTE`. Expected discovery failures and their resolutions are recorded in the slice checkpoint; no required check was skipped | `NS-04A` active |
| 2026-07-26 | `286c91a1` | `NS-04A` | `IN_PROGRESS` → `DONE`; `NS-04B` → `IN_PROGRESS` | Revised Network Flow and Testing Harness fixture provisions; migrated Network Flow contract/fixture/scenario/registry schemas; deleted activity accounting; retained 28 complete fixtures; replaced AC switches with direct behavior tests | Final module, service-backed, web, generated-artifact, generation, shape, drift, generated policy, and harness checks passed | Module `.cartulary/test-results/20260726T222136Z-p3078146`; service `.cartulary/test-results/20260726T222840Z-p3117785`; web `.../20260726T223312Z-p3151679`; generated owner `.../20260726T223335Z-p3153375`; generation/shape/drift/policy/harness roots are in the slice checkpoint | Network Flow Go/TypeScript projections and execution topology refreshed only through `make generate`; fixture digests were recalculated from the authored fixture tree | Clean starting baseline; no unrelated pre-existing change. All product behavior/fixtures were `RETAIN-ROUTE`; the 107-row ledger and 64 duplicate wrappers were `SPEC-PRUNE`. Expected discovery failures are recorded in the checkpoint; no required check was skipped | `NS-04B` active |
| 2026-07-26 | `286c91a1` | `NS-04B` | `IN_PROGRESS` → `DONE`; `NS-05` → `IN_PROGRESS` | Revised OTel run-derived conformance; deleted checked-in status/pass claim; added a six-row/50-selector telemetry owner; checker now executes current runtime tests; fixed closed error-class and future-profile validation | OTel target, telemetry owner, format, generation, shape, harness, drift, and generated policy passed | OTel `.cartulary/test-results/20260726T224114Z-p3178121`; owner `.../20260726T224105Z-p3177755`; shape/harness `.../20260726T224205Z-p3181760`, `.../20260726T224246Z-p3183596`; other roots are in the slice checkpoint | Execution topology refreshed only through `make generate`; no generated root was hand-edited | Clean starting baseline; no unrelated pre-existing change. Existing behavior was `RETAIN-ROUTE`, two hidden gaps were `IMPLEMENT`, and self-attestation was `SPEC-PRUNE`. Expected discovery failures are recorded in the checkpoint; no epoch change, retained reuse, or skipped required check | `NS-05` active |
| 2026-07-26 | `286c91a1` | `NS-05` | `IN_PROGRESS` → `DONE`; `NS-06` → `IN_PROGRESS` | Removed OpenAPI release-approval requirement IDs; migrated remaining OTel corpus/config/generated-constant metadata to behavior-only v2 contracts; removed a draft Reference Pack diagnostic requirement ID; classified every residual semantic-search match | Format, generation, OTel conformance, telemetry/OpenAPI/generated owner slices, shape, drift, generated policy, harness, and tracker Markdown lint passed | Final generation/OTel `.cartulary/test-results/20260726T225249Z-p3220016`, `.../20260726T225255Z-p3222171`; owner roots `.../20260726T225052Z-p3204020`, `.../20260726T225055Z-p3204336`, `.../20260726T225102Z-p3206688`; tracker lint `.../20260726T225421Z-p3227288`; remaining roots are in the checkpoint | Official generation refreshed the OTel generator-source digest and topology; no generated root was hand-edited | Clean starting baseline; no unrelated pre-existing change. Removed metadata was `SPEC-PRUNE`; behavior was `RETAIN-ROUTE`. No failure or skipped required check; no retained result reused; old epoch intentionally remains until atomic `NS-06` | `NS-06` active |
| 2026-07-26 | `286c91a1` | `NS-06` | `IN_PROGRESS` → `DONE`; `NS-07` → `IN_PROGRESS` | Verification v2→v3; deleted requirements; nine evidence schema majors; renamed both digests; enforced the new epoch; added three harness owner behavior routes | Five required owner slices, format, generation, shape, drift, generated policy, harness, and tracker lint passed | Owner and validation roots are in the checkpoint; final harness `.cartulary/test-results/20260726T231058Z-p3282282`; tracker lint `.../20260726T231304Z-p3283967` | Schema attachments, Extension projections, and topology refreshed only through official generation | Clean starting baseline; no unrelated pre-existing change. Requirements were `SPEC-PRUNE`; routing/evidence behavior was `RETAIN-ROUTE`. Expected discovery failures are recorded and resolved. No skipped check or retained result reuse; epoch changed to `cartulary.test_evidence.nlspec.v1` | `NS-07` active |
| 2026-07-26 | `286c91a1` | `NS-07` | `IN_PROGRESS` → `DONE`; `NS-08` → `IN_PROGRESS` | Official rebaseline; restored contract-generator security-test imports; synchronized two WorkbookShell tests with observable row readiness; completed OTel scheduled v3 identity/routing cutover | Generation, drift, policy, shape, harness, OTel, affected owner/unit checks, full warm check, and tracker lint passed | Retained candidate `.cartulary/test-results/20260726T232236Z-p3478312`; tracker lint `.../20260726T232546Z-p3574153`; all other roots are in the checkpoint | Policy-managed outputs changed only through official generation; inspected diffs are explained | Program worktree was already dirty; no unrelated pre-existing change. Fixes were `IMPLEMENT`; behavior remained `RETAIN-ROUTE`. Two failed warm roots are recorded and resolved. No skipped check or retained reuse; 252 artifacts use the new epoch | `NS-08` active |
| 2026-07-26 | `286c91a1` | `NS-08/NS-06R` | `NS-08` → `TODO`; `NS-06` reopened | Finalizer exposed that source snapshot v1 still hashed specification and Markdown bytes and that Testing Harness retained old digest names | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260726T232236Z-p3478312` failed as designed | Failure `.cartulary/test-results/20260726T232623Z-p3577259`; incompatible `source_snapshot_digest` in retained owner evidence | None | Program worktree already dirty; no unrelated pre-existing change. No artifact removed. Required downstream finalizer actions were skipped after preflight failure. Retained evidence was rejected, not reused | `NS-06` reopened |
| 2026-07-26 | `286c91a1` | `NS-06R` | `IN_PROGRESS` → `DONE`; `NS-07` → `IN_PROGRESS` | Testing Harness digest fields repaired; source snapshot v1→v2 excludes restricted documentation roots and Markdown before filesystem access; added invariance regression | Five harness owner slices, format, Markdown, generation, drift, policy, shape, and harness passed | Owner and validation roots are in the reopened checkpoint; tracker lint `.cartulary/test-results/20260726T233333Z-p3615753` | Topology refreshed only through official generation; no generated code changed for the snapshot algorithm | Program worktree already dirty; no unrelated pre-existing change. Fix was `IMPLEMENT`; documentation is `EDITORIAL_ONLY`. One expected stale-topology shape failure was resolved. No skipped check or retained reuse | `NS-07` active |
| 2026-07-26 | `286c91a1` | `NS-07R` | `IN_PROGRESS` → `DONE`; `NS-08` → `IN_PROGRESS` | Regenerated after source snapshot v2 and produced a replacement internally consistent new-epoch warm root | Generation, drift, policy, shape, harness, full warm check, and tracker lint passed | Retained candidate `.cartulary/test-results/20260726T233503Z-p3627091`; tracker lint `.../20260726T233809Z-p3722714` | Official generation only; inspected diffs remain explained | Program worktree already dirty; no unrelated pre-existing change. No removal or new schema, skipped check, or retained reuse; 252 artifacts share source snapshot v2 and the new epoch | `NS-08` active |
| 2026-07-26 | `286c91a1` | `NS-08/NS-07R2` | `NS-08` → `TODO`; `NS-07` reopened | Retained-run preflight accepted source snapshot v2 after tracker edits, then scheduler timing health rejected one contaminated non-isolated store lane | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260726T233503Z-p3627091` failed only `scheduler-summary-timing-drift` | Failure `.cartulary/test-results/20260726T233837Z-p3725776`; `module-workbook-store` 12.556s versus 5.705s peer median | None | Program worktree already dirty; no unrelated pre-existing change. No mutation ran, no check was skipped by choice, and no retained result was reused. Timing threshold remains unchanged | `NS-07` active |
| 2026-07-26 | `286c91a1` | `NS-07R2` | `IN_PROGRESS` → `DONE`; `NS-08` → `IN_PROGRESS` | Replaced the contaminated warm root without semantic changes | Full warm check, exact finalizer-scoped timing health, and tracker lint passed | Retained root `.cartulary/test-results/20260726T233928Z-p3728219`; timing gate `.../20260726T234226Z-p3822713`; tracker lint `.../20260726T234251Z-p3823040` | None | Program worktree already dirty; no unrelated pre-existing change. No schema/removal, skipped check, threshold change, or retained reuse. One unscoped diagnostic result is classified in the checkpoint | `NS-08` active |
| 2026-07-26 | `286c91a1` | `NS-08` | `IN_PROGRESS` → `DONE`; program complete | Finalized compatible retained evidence, refreshed three declared duration-baseline inputs, completed the broad release sequence, and classified the final forbidden-token scan | `agent-finalize`, Markdown, fast, check, CI, release, tracker Markdown, and repository sanity checks passed | Finalizer `.cartulary/test-results/20260726T234327Z-p3826127`; release `.cartulary/test-results/20260726T235217Z-p4084990`; all roots are in the closure checkpoint | No policy-managed generated output changed in this slice; prior outputs came only from official generation | Program worktree already dirty only with this effort; no unrelated pre-existing change. All removals, schema majors, and obligation/test dispositions are in their owning checkpoints. Compatible retained evidence was consumed only by finalization; no check skipped or unresolved failure remains | None |

Each handoff entry must additionally state:

- whether the worktree was dirty before the session and which pre-existing
  changes were preserved;
- the disposition of every removed test or artifact;
- every schema ID added, removed, or major-versioned;
- whether generated files changed only through official generation;
- whether the evidence epoch changed or a retained result was reused;
- the exact failing target and summary artifact when a command failed; and
- any skipped validation with the reason.

## 11. Risks and Blocker Rules

| Risk | Prevention or response |
| --- | --- |
| A new JSON catalog recreates prose authority under a different name | Reject statement/status fields and require machine artifacts to contain only executable facts |
| Trace metadata reappears under a new name | Verification v3 accepts routing semantics only; reject requirement/trace/status provenance fields |
| Requirement catalogs are deleted before downstream consumers are decoupled | Enforce `NS-02` through `NS-05` before atomic `NS-06` |
| Old and new evidence are accidentally mixed | Require the exact new evidence epoch and reject old schemas |
| Generated files are manually repaired during the large cutover | Change authored owners only, run official generation, and inspect generated policy |
| Behavior coverage is lost with traceability tests | Apply `REFACTOR` to mixed tests and preserve real production assertions before deletion |
| Conformance is overclaimed after completeness counters disappear | Treat test reports as execution evidence only; review conformance against adopted specifications |
| Current behavior gaps are hidden when placeholders disappear | Record `RETAIN-ROUTE`, `IMPLEMENT`, or `SPEC-PRUNE/FUTURE` before removing each affected accounting surface |
| A machine fact has no versioned projection | Author the smallest versioned typed contract before a test or runtime consumer uses it |
| Specification text re-enters executable inputs | Preserve the generic restricted-input boundary and synthetic negative test |
| The cutover produces an intermediate green but unsupported state | Keep the branch non-releasable until `NS-07`; no partial merge or release |

Mark a slice `BLOCKED` rather than guessing when:

- adopted owner specifications contradict each other;
- a public behavior change has no clear owner-specification disposition;
- a required executable fact has no owner specification;
- a generated artifact cannot be recreated from authored machine inputs; or
- the only proposed verification requires reading specification text.

## 12. Tracker-Creation Handoff

This initial session creates the tracker only. It does not execute `NS-01`
through `NS-08`, edit specifications, delete catalogs, modify schemas, or run
generation. The handoff must record:

- added file: `docs/handoffs/nlspec-testing-refactor-tracker.md`;
- substantive edit: decision-complete authority, disposal, cutover, schema,
  slice, validation, risk, and handoff program;
- verification: `make lint-markdown` passed with run root
  `.cartulary/test-results/20260726T205341Z-p2826022`;
- skipped product checks: intentionally skipped because this is a
  documentation-only tracker addition; and
- next slice: `NS-01`.

## 13. Binary Completion Criteria

The program is complete only when every item below is true:

- [x] Adopted NLSpecs and normative Core owner sections are explicitly the
  primary behavioral authority.
- [x] The testing harness is restored as an adopted/current NLSpec and forbids
  executable consumption of specification documents.
- [x] Machine-readable facts consumed by tests have versioned schemas or
  contract identities and contain no copied narrative authority.
- [x] `contracts/requirements/**`, its schemas, attachments, loaders, and
  completeness enforcement are absent.
- [x] Verification registry/contract v3 routes tests without resolving or
  counting specification requirements.
- [x] Verification v3 rejects requirement, specification-trace, and redundant
  status fields.
- [x] Contract-family registry v3 and OpenAPI source manifest v3 have no
  requirement-registry coupling.
- [x] Extensions contain typed operational contract IDs and no requirement,
  acceptance, catalog, document-status, or requirement-hash fields.
- [x] Graph behavior fixtures remain executable while the conformance matrix,
  redundant corpus registry, and requirement IDs are gone.
- [x] Reporting and Report Composition traceability-only corpora/tests are
  removed, with production behavior assertions preserved.
- [x] Network Flow fixtures execute without acceptance IDs, copied owner
  requirements, phase selectors, or activity-accounting inputs.
- [x] OTel conformance is derived from actual checks and no checked-in
  pass/claim status remains.
- [x] Every affected current Definition-of-Done obligation is retained/routed,
  implemented/tested, or revised in its owner specification before being
  pruned or moved to a future profile.
- [x] Old digest names are absent and all affected artifacts use
  `test_catalog_digest` and `verification_routing_digest`.
- [x] Current evidence requires
  `cartulary.test_evidence.nlspec.v1`; pre-cutover evidence cannot close current
  gates.
- [x] No product, test, generator, conformance, or release input reads, stats,
  hashes, or otherwise depends on `docs/`.
- [x] Markdown lint remains editorial-only and cannot count as product
  conformance.
- [x] Repository-wide audit finds no traceability-only test, planned executable
  placeholder, requirement statement catalog, or specification completeness
  counter.
- [x] Official generation and drift checks pass with no hand-edited generated
  files.
- [x] All required narrow owner checks, `make test-fast`, `make check`,
  `make ci`, and `make release-check` pass, with run roots recorded.
- [x] The final handoff records changed files, schema migrations, test
  dispositions, generated outputs, verification results, skipped checks, and
  any unrelated pre-existing worktree changes.
