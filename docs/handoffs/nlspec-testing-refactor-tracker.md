# NLSpec-First Testing Refactor Tracker and Handoff

## 1. Mission and Source Posture

This tracker governs the repository-wide replacement of the current
machine-first testing authority with this derivation chain:

**Intent → adopted NLSpec/Core specification → versioned machine projection →
implementation and behavioral tests**

This is a fundamental testing-program overhaul, not an incremental cleanup.
Production behavior is frozen unless a separately adopted specification change
authorizes a correction. Harness metadata, traceability schemas, generated
artifacts, and retained evidence may break through the coordinated cutover
defined below.

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
| Repository baseline | Commit `b45905b1`; clean tracked worktree before tracker creation |
| Archaeological boundary | Commit `a7b4eced` introduced the current requirement-catalog authority; it is evidence, not a revert target |
| Program scope | Repository-wide, prioritizing harness, OpenAPI, Extensions, Graph Projection, Reporting, and Report Composition |
| Production behavior | Frozen; only testing authority, internal contracts, traceability, generation, and evidence compatibility are authorized to change |
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
| The generic executable-input restriction already prevents dependencies on `docs/` without parsing documentation | Executable-input policy and restricted-input boundary | Retain and reframe as architecture protection |
| The old extension document parser checked headings, byte ranges, counts, hashes, and formatting | Deleted code in the parent history of `a7b4eced` | Do not restore it or any equivalent document parser |

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

### 3.2 Optional specification trace annotations

Verification contract v3 may contain an optional ASCII-sorted,
duplicate-free `spec_trace_ids` array. These tokens are human provenance aids
only:

- there is no minimum item count;
- migration must not mechanically rename or preserve all current
  `requirement_ids`;
- unknown tokens are accepted if they satisfy the token shape;
- no loader resolves them against a file, catalog, registry, or document;
- they do not affect routing, selection, pass/fail, coverage, conformance, or
  skip policy;
- they are excluded from verification routing digests and all evidence
  identities; and
- they are forbidden from production contracts and runtime structures.

### 3.3 Fixed disposition matrix

| Disposition | Apply when | Examples | Required result |
| --- | --- | --- | --- |
| `REMOVE` | The test or artifact exists only to compare documentation or account for copied requirements | Document parsers; prose/heading/anchor/line/hash checks; requirement statements/statuses; AC counts; planned placeholders; traceability-only corpora; catalog-hash gates | Delete the test/artifact and its catalog or schema registration |
| `REFACTOR` | A test mixes useful behavior with traceability accounting | Runtime test plus requirement-array assertion; output-kind behavior inside a fixture-corpus accounting test | Retain the behavior assertion in an owner-aligned behavioral test and remove accounting |
| `RETAIN` | The test exercises production-relevant behavior or validates a versioned machine contract | Runtime/API/persistence/security/failure tests; real fixture/golden execution; schema validation; generator drift | Preserve behavior and route it through verification v3 |
| `EDITORIAL_ONLY` | The check maintains documentation quality but does not establish product conformance | Markdown lint | Keep outside product, conformance, release, and generated evidence |
| `ARCHITECTURE_PROTECTION` | The check prevents executable dependencies on specifications without inspecting their content | Generic `docs/` restricted-input boundary | Retain; test policy behavior with synthetic executable references |

Planned specification behavior remains in the specification or a handoff
tracker. It must not appear as an executable placeholder, active verification,
passing row, or release-evidence claim.

## 4. Coordinated Cutover Strategy

All implementation work from `NS-01` through `NS-07` belongs to cutover set
`NLTEST-CUTOVER-1`. The branch is non-releasable after `NS-01` begins and
remains so until `NS-07` passes. Intermediate commits are rollback checkpoints,
not supported repository states.

The order is deliberate:

1. Repair the specifications and human guidance first.
2. Decouple production and generated consumers from requirement catalogs while
   the old central harness remains temporarily intact.
3. Remove traceability-only subsystem tests and artifacts.
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
| `NS-02` | Contract-family and OpenAPI decoupling | parallel | `NS-01` | `platform.openapi`, contract generation owner | Non-releasable |
| `NS-03` | Extensions typed-contract migration | chain | `NS-01`, `NS-02` contract-ID rules | `module.extensions` | Non-releasable |
| `NS-04` | Graph, Reporting, and Composition test cleanup | parallel | `NS-01` | Subsystem owners | Non-releasable |
| `NS-05` | Repository-wide authority-consumer audit | chain | `NS-02`, `NS-03`, `NS-04` | Cross-owner harness maintainer | Non-releasable |
| `NS-06` | Verification v3, evidence epoch, and requirement-layer deletion | chain | `NS-05` | Harness catalog/evidence owners | Atomic central cutover; non-releasable |
| `NS-07` | Generation, drift closure, and evidence rebaseline | chain | `NS-06` | Generated-artifact and affected owners | Releasable only after all exit checks pass |
| `NS-08` | Full validation and final handoff | chain | `NS-07` | Cross-owner release verifier | Program closure |

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

**Status:** `TODO`.

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

**Status:** `TODO`.

**Contract-family changes:**

1. Replace `cartulary.contract_family_registry.v2` with v3.
2. Remove `owner_requirement_ids` from the schema, authored registry,
   generator types, validation, and tests.
3. Retain non-empty, versioned `owner_contract_ids`.
4. Retain `generation_status` only as generator lifecycle configuration. It
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

**Status:** `TODO`.

**Actions:**

1. Delete the extension traceability mapping source and stop generating
   requirement-coverage artifacts and schemas.
2. Remove requirement registry/catalog paths, schema IDs, hashes, imported
   requirement IDs, acceptance-criterion IDs, owner requirement IDs, required
   specification statuses, catalog-coverage checks, and requirement-bearing
   closure members from authored and generated extension contracts.
3. Replace operational references as follows:

| Old field family | Target field family |
| --- | --- |
| `owner_requirement_id` | `owner_contract_id` |
| `resolution_requirement_id` | `resolution_contract_id` |
| `authorization_requirement_id` | `authorization_contract_id` |
| `redaction_requirement_id` | `redaction_contract_id` |
| `error_requirement_id` | `error_contract_id` |
| `codec_requirement_id` | `codec_contract_id` |

4. A target contract ID must identify a versioned machine contract that contains
   the executable fact. If none exists, author the smallest versioned typed
   contract before migrating the reference; do not copy the specification
   sentence.
5. Update dependency snapshots, owner manifests/fragments, shared-owner
   resolutions, profile contracts, conformance manifests, contract closures,
   integrity artifacts, generator/validator code, production catalog types,
   and tests.
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
- Extension runtime behavior and public descriptors are unchanged except for
  removal/renaming of internal traceability fields that are not public product
  behavior.

**Validation:**

- `make json-shape-check`
- `make generate`
- `make generate-drift`
- `make generated-artifact-policy-check`
- `make test-slice OWNER=module.extensions`
- `make test-slice OWNER=harness.generated_artifacts`

### NS-04: Graph, Reporting, and Composition test cleanup

**Status:** `TODO`.

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
2. Remove the fixture-corpus verification and test-family routing entry.
3. Add no replacement unless a production behavior assertion was previously
   hidden in the deleted test.

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

### NS-05: Repository-wide authority-consumer audit

**Status:** `TODO`.

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
- the old digest names `catalog_semantic_digest` and
  `verification_semantic_digest`;
- production or generated `spec_trace_ids`;
- executable references to `docs/`;
- tests whose only outcome is traceability, completeness, count, status, or
  document-shape accounting; and
- planned placeholders in active verification or test routing.

Allowed post-audit surfaces include typed contract IDs, versioned executable
facts, verification IDs, test row IDs, operational service requirements,
behavior fixture IDs, and optional verification-only `spec_trace_ids`.

**Acceptance:** A repository-wide search result is attached to the handoff with
every match classified as removed, renamed, or a documented operational
non-specification use.

**Validation:** Narrow owner slices for every changed owner, followed by
`make json-shape-check`, `make generate-drift`, and `make harness-contract`.

### NS-06: Verification v3, evidence epoch, and requirement-layer deletion

**Status:** `TODO`.

This is the atomic central cutover. It is complete only when all changes below
land together.

**Verification changes:**

1. Replace verification registry/contract v2 with v3.
2. Keep `verification_id`, `behavior_class`, `profile`, `evidence_kinds`,
   optional `public_target`, skip policy, and active routing status.
3. Remove required `requirement_ids`; add optional `spec_trace_ids` under
   Section 3.2.
4. Remove the requirement loader, requirement-to-verification map, active or
   planned requirement checks, and requirement completeness enforcement.
5. Retain routing integrity: unique owner-qualified verification IDs, valid
   profiles/evidence kinds/skip policies, valid row references, and the rule
   that each active verification has a test row or declared public target.
6. Compute `verification_routing_digest` from routing-semantic fields only,
   explicitly omitting `spec_trace_ids`.

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

1. Unknown but lexically valid `spec_trace_ids` load successfully.
2. Adding, removing, or reordering normalized trace annotations does not change
   `verification_routing_digest`, test selection, or evidence identity.
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

### NS-07: Generation, drift closure, and evidence rebaseline

**Status:** `TODO`.

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

### NS-08: Full validation and final handoff

**Status:** `TODO`.

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

## 7. Public and Internal Interface Migration

| Surface | Breaking change | Compatibility rule |
| --- | --- | --- |
| Requirement registry/catalog v1 | Removed | No replacement statement/status catalog |
| Verification registry/contract v2 | Replaced by v3 with optional non-semantic `spec_trace_ids` | v2 rejected; no dual reader |
| Contract-family registry v2 | Replaced by v3 without `owner_requirement_ids` | v2 rejected |
| OpenAPI source manifest v2 | Replaced by v3 without `requirements_registry` | v2 rejected |
| Graph fixture manifest v2 | Replaced by v3 without `requirement_ids` | v2 rejected |
| Extension contracts | Requirement/AC/catalog fields removed; operational references become typed contract IDs | Affected schemas receive new majors; no aliases |
| Test/evidence digests | Renamed to test-catalog and verification-routing digests | Old fields rejected |
| Test evidence | Requires `cartulary.test_evidence.nlspec.v1` epoch | Earlier evidence cannot close current gates |
| Reporting/composition trace corpora | Removed | Specifications retain future requirements; no placeholder replacement |
| Graph conformance matrix/corpus registry | Removed | Executable fixture directories and behavior tests remain |

Production HTTP, WebSocket, storage, authorization, domain, and browser behavior
are not authorized to change. If implementation reveals that removing a
traceability field changes a public product contract, mark the slice
`BLOCKED`, identify the owning specification, and separate that behavior change
from this program.

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

Use `make explain-test-owner`, `make explain-target`, and `make explain-run`
when failures or execution scope are unclear.

### Required behavioral scenarios

| Scenario | Expected result |
| --- | --- |
| Specification prose, headings, formatting, or file arrangement changes | Product/harness execution and evidence identity are unchanged; only editorial tooling may react |
| Executable source or machine input references `docs/` | Restricted-input architecture check fails without opening or hashing the document |
| Versioned machine artifact has invalid shape or unknown schema identity | Shape validation fails before product assertions |
| Typed contract changes a value consumed by production/tests | Generation/drift and the relevant behavior tests reflect the contract change |
| `spec_trace_ids` contains an unknown valid token | Verification loads; execution and evidence digests are unchanged |
| Planned specification behavior has no implementation | It remains absent from active rows and passing evidence |
| Graph fixture executes | Real code runs and observable artifacts/state effects are compared |
| Reporting supported output kinds are queried | Production code returns Mermaid and Slidev |
| Pre-cutover retained evidence is supplied to a current gate | Epoch/schema mismatch rejects reuse |
| Requirement catalog, requirement statement, or AC completeness input is added | Schema/boundary/audit checks reject it |

### Documentation-only validation

`make lint-markdown` maintains Markdown quality. It is not product,
conformance, generated-artifact, or release evidence.

## 9. Top-Level Work Tracker

| ID | Work item | Status | Depends on | Owner | Evidence | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| `NT-001` | Establish authority, scope, baseline, and cutover doctrine | `DONE` | None | Handoff owner | Sections 1–5 | Decisions and quantitative baseline are recorded |
| `NT-002` | Repair specification and guidance precedence | `TODO` | `NT-001` | Specification owners | `NS-01` diff and Markdown run | All authority prose is specification-first |
| `NT-003` | Decouple contract family and OpenAPI assembly | `TODO` | `NT-002` | Contract/OpenAPI owners | `NS-02` schemas, tests, generation | No requirement registry dependency remains |
| `NT-004` | Replace extension requirement traceability with typed contracts | `TODO` | `NT-002`, `NT-003` | `module.extensions` | `NS-03` authored/generated inventory | No requirement/AC/catalog field remains in Extensions |
| `NT-005` | Remove Graph, Reporting, and Composition accounting-only tests | `TODO` | `NT-002` | Subsystem owners | `NS-04` behavior test results | Valuable behavior remains; traceability-only artifacts are gone |
| `NT-006` | Complete repository-wide consumer audit | `TODO` | `NT-003`, `NT-004`, `NT-005` | Harness maintainer | Classified search inventory | Every forbidden match is removed or classified |
| `NT-007` | Execute central verification/evidence cutover and delete requirements | `TODO` | `NT-006` | Harness catalog/evidence owners | `NS-06` schema and harness results | v3 routing and new epoch pass without requirement catalogs |
| `NT-008` | Regenerate and establish first new-epoch evidence | `TODO` | `NT-007` | Generated-artifact owners | `NS-07` drift and retained run root | Generated state is clean and current |
| `NT-009` | Run full closure validation and hand off | `TODO` | `NT-008` | Cross-owner verifier | `NS-08` command/run ledger | All binary completion criteria pass |

## 10. Session Handoff Log

Append one row after every implementation session. Do not rewrite earlier rows
except to correct a factual error.

| Timestamp | Baseline commit | Slice | Status transition | Files/contracts changed | Commands and results | Run roots/artifacts | Generated outputs | Blockers/failures | Next unblocked slice |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-26 | `b45905b1` | `NS-00` | `TODO` → `DONE` | Added this tracker only | Planning baselines in Section 2; `make lint-markdown` passed | Section 2 roots and `.cartulary/test-results/20260726T205341Z-p2826022` | None | None | `NS-01` |

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
| Optional trace IDs become de facto coverage | Exclude them from digests, selection, pass/fail, completeness, and production contracts |
| Requirement catalogs are deleted before downstream consumers are decoupled | Enforce `NS-02` through `NS-05` before atomic `NS-06` |
| Old and new evidence are accidentally mixed | Require the exact new evidence epoch and reject old schemas |
| Generated files are manually repaired during the large cutover | Change authored owners only, run official generation, and inspect generated policy |
| Behavior coverage is lost with traceability tests | Apply `REFACTOR` to mixed tests and preserve real production assertions before deletion |
| Conformance is overclaimed after completeness counters disappear | Treat test reports as execution evidence only; review conformance against adopted specifications |
| A machine fact has no versioned projection | Author the smallest versioned typed contract before a test or runtime consumer uses it |
| Specification text re-enters executable inputs | Preserve the generic restricted-input boundary and synthetic negative test |
| The cutover produces an intermediate green but unsupported state | Keep the branch non-releasable until `NS-07`; no partial merge or release |

Mark a slice `BLOCKED` rather than guessing when:

- adopted owner specifications contradict each other;
- removing a traceability field changes public product behavior;
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

- [ ] Adopted NLSpecs and normative Core owner sections are explicitly the
  primary behavioral authority.
- [ ] The testing harness is restored as an adopted/current NLSpec and forbids
  executable consumption of specification documents.
- [ ] Machine-readable facts consumed by tests have versioned schemas or
  contract identities and contain no copied narrative authority.
- [ ] `contracts/requirements/**`, its schemas, attachments, loaders, and
  completeness enforcement are absent.
- [ ] Verification registry/contract v3 routes tests without resolving or
  counting specification requirements.
- [ ] Optional `spec_trace_ids` exist only in verification metadata and have no
  semantic or evidence effect.
- [ ] Contract-family registry v3 and OpenAPI source manifest v3 have no
  requirement-registry coupling.
- [ ] Extensions contain typed operational contract IDs and no requirement,
  acceptance, catalog, document-status, or requirement-hash fields.
- [ ] Graph behavior fixtures remain executable while the conformance matrix,
  redundant corpus registry, and requirement IDs are gone.
- [ ] Reporting and Report Composition traceability-only corpora/tests are
  removed, with production behavior assertions preserved.
- [ ] Old digest names are absent and all affected artifacts use
  `test_catalog_digest` and `verification_routing_digest`.
- [ ] Current evidence requires
  `cartulary.test_evidence.nlspec.v1`; pre-cutover evidence cannot close current
  gates.
- [ ] No product, test, generator, conformance, or release input reads, stats,
  hashes, or otherwise depends on `docs/`.
- [ ] Markdown lint remains editorial-only and cannot count as product
  conformance.
- [ ] Repository-wide audit finds no traceability-only test, planned executable
  placeholder, requirement statement catalog, or specification completeness
  counter.
- [ ] Official generation and drift checks pass with no hand-edited generated
  files.
- [ ] All required narrow owner checks, `make test-fast`, `make check`,
  `make ci`, and `make release-check` pass, with run roots recorded.
- [ ] The final handoff records changed files, schema migrations, test
  dispositions, generated outputs, verification results, skipped checks, and
  any unrelated pre-existing worktree changes.
