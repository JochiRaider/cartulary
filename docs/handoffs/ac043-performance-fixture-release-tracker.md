# AC-043 Performance Fixture Snapshot and Release Handoff Tracker

## 1. Purpose, authority, and scope

This document is the controlling execution ledger for the remaining AC-043 fixture-snapshot and release work. It is implementation-support evidence, not a behavioral owner. It MUST NOT be read, hashed, or otherwise consumed by product code, test code, generators, conformance tooling, or release tooling.

The primary implementation seam is the AC-043 fixture-snapshot boundary between browser catalog routing and `tools/testservices`. The work replaces four independent, per-predicate production-API fixture builds with one suite-scoped populated PostgreSQL template and one isolated clone per predicate.

Authority is resolved in this order:

1. [Core 04](../spec/04_security_deployment_and_conformance.md) owns the supported fixture shape, interaction predicates, sampling policy, and thresholds.
2. [Testing Harness NLSpec](../testing-harness-nlspec.md) owns snapshot construction, fixture-broker leases, clone isolation, cleanup, qualification evidence, and scheduler behavior.
3. Typed contracts and authored harness inputs project those owners into executable data.
4. Generated consumers, catalog rows, test support, browser tests, and retained artifacts implement or verify the typed projections.
5. This tracker records decisions, sequence, evidence, and handoff state. If it conflicts with an adopted owner, the owner wins and this tracker is corrected.

The writing and acceptance posture follows the [NLSpec writing standard](../research/nlspec-spec.md): interfaces, defaults, boundaries, failure behavior, and completion criteria are explicit and testable. The workflow and append-only handoff structure follows the [modular refactoring framework](cartulary_modular_refactor_planning_framework.md).

### 1.1 Repository baseline

The following state was rechecked immediately before this tracker was created on 2026-08-13:

| Field | Recorded value |
| --- | --- |
| Branch | `main` |
| Commit | `cc76c32682bc35cc766e500e3b5efd60be398907` |
| Worktree | Clean before this tracker was added |
| Primary seam | Browser catalog routing to `tools/testservices` fixture provisioning |
| Current implementation | Deterministic key helper exists, but all four predicates still assemble the large fixture through production APIs |
| Current release claim | No green current-source release is claimed |

The implementing agent MUST refresh branch, commit, worktree state, and retained-run availability at the start of each implementation session. A changed baseline is recorded in the append-only handoff log; it is not silently substituted here.

### 1.2 Evidence baseline

| Evidence | Status | Interpretation | Preservation rule |
| --- | --- | --- | --- |
| `.cartulary/test-results/20260813T194448Z-p3008841` | Failed, 925/929, 1,979,786 ms | The planned release run predates the final AC-043 fixes. Its failed units include the pre-fix Timeline measurement failure and an independent object-store readiness failure. It does not qualify current source. | Preserve the result unchanged. Do not retry it, rewrite it, or replace it with a later result. |
| `.cartulary/test-results/20260813T212600Z-p3589519` | Passed, 18/18, 1,083,828 ms | Narrow corroboration for the four isolated predicates: one warm-up and 100 samples each, unchanged thresholds, and zero scheduler overlap. The approximately 18-minute duration demonstrates the remaining fixture-build cost. | Preserve as narrow corroboration only. It is not release qualification and does not close snapshot work. |

The failed release and narrow pass are separate observations. The narrow pass does not retroactively reclassify the release failure, and neither result is a retry of the other.

### 1.3 Fixed decisions and non-goals

The remaining implementation observes all of the following boundaries:

- AC-043 remains a Core 04 supported-envelope implementation gate.
- Selection, focus, and typing acknowledgment retain a 100 ms p95 limit. Blank-row creation retains a 150 ms p95 limit.
- Each predicate retains one warm-up pass, exactly 100 measured operations, nearest-rank p95, and zero retries.
- Core 05 claim publication is not added or implied.
- No public API, production storage schema, database migration, threshold, or retry-policy change is introduced.
- The snapshot is suite-scoped test infrastructure. Cross-run fixture reuse and measurement-result caching are out of scope.
- The implementation does not use raw E2E SQL, per-predicate fixture assembly, catalog-string inference, aliases, or fallback assembly.
- Snapshot optimization does not weaken security controls, foreground-worker provisioning, traffic qualification, quiet scheduling, fixture validation, or evidence redaction.
- A later passing result never replaces or mutates an earlier failed result.

## 2. Top-level work tracker

Allowed status values are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and `DROPPED`. At most one work item is `IN_PROGRESS` at a time. Status changes require a dated append-only handoff entry.

| ID | Work item | Workstream | Status | Depends on | Owner | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| AC043-001 | Core 04 predicate, threshold, fixture, and sampling ownership | Prior remediation | DONE | None | Core 04 | Core 04 requirements and `contracts/performance/ac043.v1.json` | Product-visible predicates, exact limits, exact fixture, and sampling are normatively and executably defined. |
| AC043-002 | Quiet measurement scheduling and zero-overlap qualification | Prior remediation | DONE | AC043-001 | Testing Harness | `browser_measurement_quiet`, scheduler evidence, narrow run | Measurements hold the exclusive quiet session and retain overlap proof. |
| AC043-003 | Four independent paint-qualified predicate measurements and summaries | Prior remediation | DONE | AC043-001, AC043-002 | Timeline, web application, Testing Harness | Four catalog rows and `cartulary.frontend_measurement_summary.v1` evidence | All four predicates execute independently with exact thresholds and 100 measured samples. |
| AC043-004 | Snapshot authority and typed fixture-profile contract | WS-01 | TODO | AC043-001..003 | Testing Harness, Core 04 projection reviewers | Adopted NLSpec edit, registry, schema, catalog field, generated projections | Owner review and contract, schema, generation, drift, and JSON-shape checks pass. |
| AC043-005 | Harness-only source-owner fixture contribution assembler | WS-02 | TODO | AC043-004 | Auth, Incidents, Entities, Timeline, Links, Projections, Testing Harness | Contribution manifest, assembler, deterministic validation tests | Repeated builds produce the same semantic digest and exact fixture counts; invalid manifests fail closed. |
| AC043-006 | Snapshot broker, populated template, clones, and cleanup lifecycle | WS-03 | TODO | AC043-005 | Testing Harness, testservices, PostgreSQL test support | Builder prerequisite, leases, artifact, lifecycle tests | One build and four isolated clones occur; concurrency, failure, and stale-resource cleanup are proven. |
| AC043-007 | Browser and catalog cutover to explicit fixture profile | WS-04 | TODO | AC043-006 | Testing Harness, Timeline, web application | Four catalog bindings, generated topology, browser support changes | Direct row, owner slice, aggregate measurement, and release graph use identical snapshot semantics with no fallback. |
| AC043-008 | Narrow current-source qualification | WS-05 | TODO | AC043-007 | Testing Harness, Timeline | Current-source narrow result root and four v2 summaries | All four predicates pass unchanged limits, snapshot proof is complete, and overlap is zero. |
| AC043-009 | Single planned current-source release qualification | WS-05 | TODO | AC043-008 | Release operator | One retained `make release-check` result root | Every selected release unit passes in the one planned run. No automatic retry occurs. |
| AC043-010 | Final handoff, cleanup accounting, and tracker closure | WS-05 | TODO | AC043-009 | Implementing agent | Complete diff review, handoff log, migration notes, retained artifacts | All binary acceptance criteria are satisfied and the final handoff is complete. |

## 3. Current-state inventory and gap assessment

### 3.1 Relevant implementation surface

| Path or surface | Current responsibility | Remaining problem |
| --- | --- | --- |
| `contracts/performance/ac043.v1.json` | Projects the exact AC-043 fixture, predicates, sampling, and thresholds | It does not select or describe a harness snapshot profile. |
| `tools/test_families/module.timeline.json` | Routes four release-tier measurement rows through `browser_measurement_quiet` and `browser_stack` | The rows have no explicit populated-fixture profile. |
| `apps/web/e2e/support/performance/ac043Fixture.ts` | Computes a deterministic cache-key-shaped value, assembles the fixture, validates it, and creates background users | It rebuilds the entire fixture through production APIs for every predicate and does not provide a populated snapshot. |
| `apps/web/e2e/measurement/timeline-grid.spec.ts` | Executes four independent paint-qualified measurements | Every test invokes live fixture assembly and background-account creation. |
| `tools/testservices/main.go` | Creates a migrated template, prepares browser fixtures, and owns service cleanup | It cannot select, build, seal, lease, or clone a populated performance template. |
| `internal/testutil/pgtest` | Provides PostgreSQL test lifecycle support, including template cloning | It needs the broker-facing lifecycle and failure-cleanup coverage required by populated templates. |
| Browser measurement evidence | Retains predicate timing, quietness, and qualification details | Version 1 summaries have no snapshot provenance and cannot close this work. |

### 3.2 Gap evaluation

| Gap | Remediation and areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criterion |
| --- | --- | --- | --- | --- | --- |
| Deterministic key without reusable populated state | Add the typed profile, suite-scoped builder, sealed template, per-predicate clones, and lifecycle artifact. Areas: specification, contracts, harness, tests, documentation. | Moves expensive deterministic setup to the correct shared prerequisite while keeping measured state isolated. It gives later performance fixtures a reusable, explicit lifecycle. | Internal catalog and generated topology evolve. No production schema or public API changes. | Four redundant builds keep exact qualification near 18 minutes and invite tactical fixture shortcuts. | A four-row aggregate records exactly one build and four independent clones. |
| Fixture meaning is implicit in browser code | Add `fixture_profile_id` and a closed typed snapshot registry. Areas: contracts, catalog, generation, validation. | Explicit routing is stable under file, title, scenario, and predicate renames. | All four current rows receive the new field; non-snapshot rows remain unaffected under the schema rule. | Filename or predicate inference creates hidden coupling and silent misrouting. | Catalog negatives reject missing, unknown, and incompatible profiles for the four rows. |
| Fixture construction lacks source-owner closure | Introduce a harness-only contribution assembler whose contributors use production services and validators. Areas: test support, source-owner test APIs, contracts, tests. | Preserves domain invariants and makes ownership, ordering, drift, and extension explicit without raw E2E SQL. | New internal test-support interfaces only. | A monolithic loader can bypass invariants, depend on insertion accidents, and become difficult to extend. | Duplicate, omitted, reordered, or contract-mismatched contributions fail before sealing. |
| Populated template lifecycle is unspecified | Adopt harness requirements for builder admission, sealing, leases, clones, cleanup, and janitor behavior. Areas: Testing Harness NLSpec, harness implementation, tests. | Distinguishes fixture reuse from forbidden result caching and gives failures one owner and one cleanup contract. | Snapshot resources exist only inside one suite invocation. | Partial templates, leaked databases, clone aliasing, and accidental cross-run reuse can contaminate results. | Concurrency and injected-failure tests prove single flight, isolation, and complete cleanup. |
| Existing evidence lacks snapshot provenance | Add `cartulary.performance_fixture_snapshot.v1` and link it from `cartulary.frontend_measurement_summary.v2`. Areas: schemas, finalization, tests, security policy. | Qualified results become reproducible and comparable without exposing data or infrastructure secrets. | Historical v1 summaries remain readable history but cannot qualify the snapshot implementation. | A fast run could use an unknown, stale, partial, or shared fixture and still appear qualified. | Pass and failure fixtures emit schema-valid, redacted provenance; missing or invalid provenance fails closed. |
| No current-source release result exists | Complete narrow qualification, finalization, and exactly one planned release run. Areas: validation and handoff. | Separates implementation confidence from full release qualification and preserves failure integrity. | New retained run roots are additive. | The project may imply release health from narrow evidence or retry away a real failure. | One current-source release root passes every selected unit; otherwise release remains failed. |

## 4. Decision-complete target architecture

### 4.1 Ownership and interface boundary

| Concern | Owner | Required interface |
| --- | --- | --- |
| Fixture dimensions and semantic distribution | Core 04 | Versioned performance contract referenced by fixture contributions |
| Predicate actions, completion states, sample count, estimator, and limits | Core 04 | Existing performance contract and generated UI facade |
| Snapshot registry, build admission, lease scope, clone lifecycle, and cleanup | Testing Harness | Authored typed registry, generated harness consumer, and `tools/testservices` broker |
| Source data construction and validation | Auth, Incidents, Entities, Timeline, Links, Projections | Test-support contribution interfaces implemented with production services and validators |
| Catalog binding and graph compilation | Verification catalog and Testing Harness | Explicit `fixture_profile_id` resolved during graph compilation |
| Browser preparation and measurement | Timeline and web application tests | Snapshot discovery, query-based validation, background login, and v2 summary emission |
| Retained qualification evidence | Testing Harness | Snapshot artifact plus measurement summary link and finalizer validation |

No contributor owns the suite lifecycle. No central assembler owns source meaning. The harness validates the closed contribution set, orders it by declared dependencies, coordinates execution, and owns cleanup.

### 4.2 Fixture profile and catalog contract

The authored profile ID is `ac043_large_grid_snapshot_v1`. A new catalog field named `fixture_profile_id` binds this profile to exactly these four current measurement rows:

- `module.timeline.measurement.committed_timeline_summary_typing_acknowledgment_b615aabfe6`
- `module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13`
- `module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3`
- `module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95`

The compiler resolves the field through the typed fixture-snapshot registry. It never infers the profile from a file name, test title, scenario ID, verification ID, predicate ID, family ID, or resource profile.

The catalog schema applies these rules:

- The four rows above MUST declare `fixture_profile_id: ac043_large_grid_snapshot_v1`.
- A row that declares a fixture profile MUST use a compatible fixture capability and runtime profile.
- An unknown, inactive, duplicate, or incompatible profile is a catalog error.
- A measurement row that is contractually bound to a populated fixture cannot omit the field.
- Rows that do not require a populated fixture omit the field; omission never means AC-043 by default.

Generated topology carries the resolved profile and snapshot prerequisite identity to both direct-row and aggregate target plans. No constructor duplicates the profile mapping.

### 4.3 Typed fixture-snapshot registry

Add an authored registry under `tools/` with schema ID `cartulary.performance_fixture_snapshot_owner.v1`. The exact file name may follow an established owner-registry naming convention, but there MUST be one canonical authored source and one schema. Generated consumers are produced through the existing Make-owned generation path; generated roots are never hand-edited.

Each profile entry defines:

| Field | Meaning and rule |
| --- | --- |
| `fixture_profile_id` | Stable catalog-facing ID; for this work, `ac043_large_grid_snapshot_v1`. |
| `fixture_version` | Stable semantic fixture version; `cartulary.perf.large_grid.v1`. |
| `seed` | Integer seed `20260405`. |
| `source_contract_refs` | Closed, sorted references to the typed owner contracts used by contributors. |
| `contributions` | Closed contribution list with stable ID, version, owner, dependencies, contract references, and expected semantic output. |
| `validation_rules` | Exact count, relationship, distribution, default view, security, and no-session rules. |
| `cleanup_policy` | Suite owns populated template; predicate lease owns clone and empty object-store bucket; failure and stale-resource cleanup are mandatory. |
| `artifact_schema_id` | `cartulary.performance_fixture_snapshot.v1`. |

The initial closed contribution manifest is ordered by declared dependency, not by source-file discovery:

1. `auth.background_analysts.v1`
2. `incidents.workspace.v1`
3. `entities.hosts_identities.v1`
4. `timeline.large_grid.v1`
5. `links.timeline_associations.v1`
6. `projections.timeline_entities.v1`
7. harness semantic validation

The exact names become contract values once adopted. Changing membership, order dependencies, version, or contract references changes `source_contract_digest` and therefore changes the snapshot key.

### 4.4 Snapshot identity

The snapshot key uses only stable typed inputs:

- `migration_digest` is the existing canonical `pgschema.Hash()` identity.
- `source_contract_digest` is SHA-256 over canonical JSON for the closed typed fixture-contribution manifest, including contribution IDs, versions, owners, dependencies, and source-contract identities.
- `fixture_version` is `cartulary.perf.large_grid.v1`.
- `seed` is integer `20260405`.
- `schema_id` is the versioned schema ID for the snapshot-key envelope.

The canonical key input has exactly this semantic shape:

```json
{
  "schema_id": "cartulary.performance_fixture_snapshot_key.v1",
  "migration_digest": "<lowercase canonical digest>",
  "source_contract_digest": "<lowercase canonical digest>",
  "fixture_version": "cartulary.perf.large_grid.v1",
  "seed": 20260405
}
```

`snapshot_key` is the lowercase hexadecimal SHA-256 digest of the repository's canonical JSON encoding of that object. Key generation rejects absent values, unknown fields, non-canonical digest encodings, and unsupported schema or fixture versions. Tests prove that map ordering and caller formatting do not change the key, while every semantic component does.

### 4.5 Source-owner fixture assembler

The assembler lives under reusable test support and is callable by the harness snapshot builder, not by a Playwright scenario. It observes these rules:

- Each source owner constructs its contribution using production owner services and production validators.
- The harness supplies deterministic seed context and already-opened suite resources.
- The assembler validates the complete registered contribution catalog before the first mutation.
- Dependencies form one acyclic closed graph. Missing, duplicate, unknown, cyclic, or incompatible contributions fail before sealing.
- Contributors do not issue raw fixture SQL from E2E or browser support. Narrow test-support adapters may call production persistence interfaces when the source owner owns that path.
- Contributors return bounded semantic receipts containing version, counts, and safe digests, not payloads or credentials.
- Repeated builds from the same migrated template and key produce the same semantic validation digest.

The populated template contains exactly the Core 04 large-grid shape:

- one deterministic fixture workspace and incident;
- 20,000 Timeline rows;
- 1,000 Hosts;
- 1,000 Identities;
- 1,000 deterministic tags, 1,000 deterministic mentions, and 1,000 deterministic links distributed across every twentieth Timeline row;
- 24 deterministic background analyst accounts and their required workspace and incident memberships;
- the default sort, filter, and grouping state required by the performance contract;
- all production security and authorization controls needed by the measured workflow; and
- no authenticated sessions, access tokens, refresh tokens, browser state, active traffic, object-store payloads, or predicate-local target mutations.

The validation pass uses production query and view surfaces where available and verifies exact counts, relationship distribution, projection readiness, target-row pools, default view state, and absence of active sessions. It computes a semantic digest from stable typed receipts and safe aggregate query results. Database physical layout, sequence positions that are not product semantics, generated record IDs, database names, and timestamps do not enter the semantic digest.

### 4.6 Snapshot builder, broker, and lifecycle

Graph compilation creates one shared snapshot-builder prerequisite for each distinct tuple of runtime profile, fixture profile, and snapshot key. Its logical identity is:

`fixture_snapshot:<runtime_profile_id>:<fixture_profile_id>:<snapshot_key>`

The lifecycle is:

1. Validate the catalog profile, registry entry, contribution closure, migration digest, and key.
2. Admit exactly one builder for the key within the suite. Concurrent dependants join the same in-flight result.
3. Acquire shared `host_activity` for construction. Measurement quiet sessions remain exclusive and do not begin during the build.
4. Create a suite-owned database from the already migrated template.
5. Run the closed source-owner contribution assembler.
6. Run exact semantic validation and emit the build portion of the snapshot artifact.
7. Close every connection to the populated database and terminate no unknown external connection.
8. Mark the database as a PostgreSQL template only after validation and connection closure succeed.
9. Release the builder's shared host activity and make the sealed snapshot available to dependent predicate leases.
10. For each predicate, create one isolated database from the sealed populated template and one empty isolated object-store bucket.
11. Start the predicate's existing exclusive quiet session across clone preparation, services, readiness, traffic stabilization, warm-up, measurement, artifact finalization, and cleanup.
12. Drop only that predicate's clone and bucket during browser-session cleanup.
13. After all dependent leases finish, suite cleanup unseals and drops the populated template.

The suite never exposes the populated template DSN to Playwright. A predicate never connects to another predicate's clone. A clone cannot become another template. Build success is not inferred from the presence of a database name; a sealed state and validated artifact are both required.

Cross-run lookup, persistent template caches, and retained-result reuse remain out of scope. The word "snapshot" in this tracker means suite-scoped populated fixture reuse only. It does not change the harness rule that browser and measurement work results are uncached.

### 4.7 Browser cutover

After the broker is proven:

- `timeline-grid.spec.ts` discovers its provisioned snapshot clone through the normal browser fixture environment.
- The existing `assembleAc043Fixture` path is deleted. No alias, compatibility wrapper, empty-template fallback, or live production-API fallback remains.
- Browser support validates the cloned fixture through production query surfaces before warm-up; it does not assemble or repair the fixture.
- The foreground Playwright worker is provisioned normally for its clone and is added to the cloned incident by the ordinary supported test path.
- Background setup logs in the 24 preseeded analyst accounts. It does not create users, workspaces, memberships, or fixture rows.
- No active background session exists until the predicate clone and quiet session are ready.
- The measured user remains foreground. The 24 background analysts retain evenly staggered updates to non-target rows every five seconds, totaling 4.8 committed updates per second, with presence enabled.
- Traffic stops before clone cleanup, and cleanup proves that no browser, backend, worker, or database connection remains.

### 4.8 Snapshot and measurement evidence

Add `cartulary.performance_fixture_snapshot.v1` as a versioned, redacted harness artifact. It records:

- fixture profile ID, fixture version, and seed;
- snapshot key schema ID and snapshot key;
- migration and source-contract digests;
- ordered contribution IDs and versions;
- exact aggregate counts and distribution checks;
- semantic validation digest;
- build state: `built`, `reused_within_suite`, or safe partial failure state;
- suite-scoped builder identity and build ordinal;
- clone proof: predicate row ID, opaque clone ordinal, parent snapshot key, creation state, isolation check, and cleanup state;
- validation outcome and bounded failure code;
- artifact creation and lifecycle timestamps needed for ordering; and
- redaction-policy version.

The artifact MUST NOT contain entered values, raw row or record IDs, transaction IDs, user identifiers, email addresses, passwords, credentials, tokens, cookies, DSNs, database or bucket names, process environment, SQL text, object payloads, or production-like sensitive data.

Add `cartulary.frontend_measurement_summary.v2`. It preserves the v1 measurement fields and adds a required snapshot provenance object containing `fixture_profile_id`, `snapshot_key`, snapshot artifact reference and SHA-256 digest, clone ordinal, clone-isolation result, and clone-cleanup result. The browser target finalizer validates both artifacts and their cross-reference.

Historical v1 summaries remain readable immutable evidence. They cannot satisfy the snapshot-provenance requirement and cannot close AC043-006 through AC043-010.

### 4.9 Fail-closed behavior

| Condition | Required outcome | Classification and evidence |
| --- | --- | --- |
| Unknown or absent required fixture profile | Do not compile or start the measurement | Catalog or topology contract failure naming row and profile field |
| Migration or source-contract digest mismatch | Do not build or clone | Harness contract/setup failure with safe expected and actual digests |
| Duplicate, missing, cyclic, or invalid contribution | Roll back construction and clean the partial database | Harness setup failure plus partial redacted snapshot artifact |
| Validation count or semantic digest mismatch | Do not seal the populated database | `environment_not_qualified`; artifact records failed rule without payload data |
| Open connection prevents safe sealing | Abort sealing and clean the partial database | Harness lifecycle failure with bounded connection-count diagnostic |
| Corrupt, partial, unsealed, or missing snapshot | Do not fall back to an empty template or live assembly | `environment_not_qualified` and cleanup proof |
| Concurrent request for the same key | Join the single in-flight build | One build ordinal; all dependants reference the same snapshot key |
| Clone failure | Do not start that predicate | Harness setup failure; other cleanup remains mandatory |
| Snapshot or clone cannot be cleaned | Keep the target failed and invoke bounded janitor handling | Harness cleanup failure with opaque resource class and lease identity |
| Missing, malformed, unredacted, or inconsistent provenance | Fail target finalization | Evidence-integrity or security failure; no measurement qualification |
| Qualified p95 exceeds its unchanged limit | Keep predicate and release failed | Product failure routed by existing timing stages; no retry |
| Object-store readiness fails | Keep affected unit and release failed separately | Harness/infrastructure failure; preserve without reclassification or retry |

## 5. Workflow dependency map

The implementation sequence is serial at the checkpoint boundary:

`WS-00 → WS-01 → WS-02 → WS-03 → WS-04 → WS-05`

Source-owner contributions inside WS-02 may be developed independently after the typed contribution interface is fixed, but WS-02 does not exit until the closed aggregate is deterministic. Browser cutover does not begin until broker lifecycle and cleanup tests pass. Release qualification does not begin until all four direct and aggregate routes use the same profile.

## 6. Ordered workstreams

### WS-00: Tracker creation and evidence freeze

**Status:** `DONE`.

**Depends on:** Prior AC-043 remediation and retained evidence.

**Edit scope:** This tracker only.

**Expected diff:** One new Markdown file under `docs/handoffs`; no specification, contract, generated, implementation, test, or retained-result mutation.

**Validation:**

- `make lint-markdown`
- `git diff --check`
- `git status --short`
- complete diff review of this file

**Rollback point:** Remove only this uncommitted tracker. Historical result roots remain untouched.

**Risks:** Recording inferred state as current qualification, conflating narrow and release evidence, or letting this tracker become an executable source.

**Exit criteria:** The file passes documentation hygiene, matches the retained summaries, records no green release claim, and the append-only handoff row captures the exact result.

### WS-01: Authority and typed fixture contract

**Status:** `TODO`

**Depends on:** WS-00.

**Edit scope:** Testing Harness NLSpec, Core 04 typed projection only if owner review finds a missing executable reference, authored fixture registry, registry schema, catalog schema and rows, generator inputs and implementations, generated consumers, contract tests, and catalog/topology validators.

**Required edits:**

1. Update the Testing Harness NLSpec to define suite-scoped populated fixture snapshots and distinguish them from forbidden browser-work and measurement-result caching.
2. Define snapshot registry ownership, key identity, build admission, sealing, leases, cleanup, evidence, and fail-closed behavior.
3. Add `fixture_profile_id` to the catalog contract and bind all four rows to `ac043_large_grid_snapshot_v1`.
4. Add the authored registry and schema with the exact fields in section 4.3.
5. Generate typed harness and topology consumers through the Make-owned path.
6. Add negative fixtures for missing profiles, unknown profiles, incompatible capabilities, malformed digests, contribution closure errors, and unsupported versions.

**Expected diff:** Normative harness changes plus authored machine contracts, generator changes, generated projections, and focused tests. No database migration, public API, production storage, threshold, sampling, retry, or Core 05 publication change.

**Validation:** Owner review; schema negative tests; `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make harness-contract`; applicable catalog and topology drift checks selected through the task guide.

**Rollback point:** Revert WS-01 as one atomic contract/projection change before any consumer depends on `fixture_profile_id`. Do not leave duplicate old and new catalog meanings.

**Risks:** Treating the snapshot as a result cache, duplicating fixture declarations across Core 04 and harness inputs, or permitting implicit catalog defaults.

**Exit criteria:** Adopted owners and typed projections agree; every profile resolves uniquely; invalid combinations fail closed; generation and drift are clean.

### WS-02: Source-owner fixture contribution assembler

**Status:** `TODO`

**Depends on:** WS-01.

**Edit scope:** Reusable harness/test support, source-owner contribution adapters and tests, fixture semantic validation, and narrow owner-facing test utilities. No raw SQL is added to E2E support.

**Required edits:**

1. Define the contribution interface and deterministic seed context.
2. Implement Auth, Incidents, Entities, Timeline, Links, and Projections contributions through production owner services and validators.
3. Seed fixed background analysts and memberships without sessions.
4. Produce exact Core 04 counts and every-twentieth-row association distribution.
5. Validate the closed graph before mutation and exact fixture semantics after assembly.
6. Emit safe per-contribution receipts and one deterministic semantic digest.
7. Add tests for repeatability, dependency order, duplicate IDs, omissions, cycles, contract drift, partial contribution failure, and payload redaction.

**Expected diff:** Test-support assembler and owner adapters with unit/integration tests. Product runtime composition and production APIs remain unchanged.

**Validation:** Focused owner slices for every changed owner; repeated-build digest comparison; exact-count and distribution tests; security/authorization validation; absence-of-session validation; invalid-manifest tests.

**Rollback point:** Revert the assembler slice while retaining the unused typed registry. No catalog route moves until this workstream exits.

**Risks:** Hidden order dependence, uncontrolled clock or ID entropy, bypassed business invariants, excessive coupling between contributors, or sensitive values entering receipts.

**Exit criteria:** Two independent builds with identical key inputs yield the same semantic digest and exact counts; all malformed contribution cases fail before sealing and clean partial state.

### WS-03: Snapshot broker and lifecycle

**Status:** `TODO`

**Depends on:** WS-02.

**Edit scope:** `tools/testservices`, reusable PostgreSQL test support, broker leases, execution-graph prerequisite compilation, cleanup/janitor logic, snapshot artifact schema and writer, and focused harness tests.

**Required edits:**

1. Compile and schedule one shared builder prerequisite per resolved snapshot key.
2. Implement single-flight construction under shared `host_activity`.
3. Build from the existing migrated template, validate, close connections, and seal the populated template.
4. Clone one database per predicate and associate it with that predicate's existing quiet session.
5. Preserve empty per-predicate object-store buckets.
6. Implement cleanup ownership for clones, partial builds, sealed templates, failed dependants, cancellation, and stale resources.
7. Emit `cartulary.performance_fixture_snapshot.v1` and validate it before dependants qualify.

**Expected diff:** Harness and PostgreSQL test-lifecycle implementation plus schemas and focused tests. Browser tests still use their old assembly path until WS-04.

**Validation:** Snapshot key tests; one-build/four-clone aggregate fixture; same-key concurrency; different-key separation; clone isolation; corruption; unsealed-template rejection; connection-closure enforcement; injected failure at every lifecycle step; cancellation; suite cleanup; janitor cleanup; artifact schema and redaction.

**Rollback point:** Revert broker consumers and builder compilation together. Delete only suite resources created by the failed test run through their recorded leases; never target a broad database namespace.

**Risks:** Duplicate builds, PostgreSQL template connection races, leaked clones, cleanup of the wrong resource, cross-predicate state, or builder activity overlapping measurement.

**Exit criteria:** The aggregate graph constructs one template, produces four isolated clones, serializes duplicate construction, proves zero builder/measurement overlap, and cleans all resources after success and injected failure.

### WS-04: Browser and catalog cutover

**Status:** `TODO`

**Depends on:** WS-03.

**Edit scope:** Four Timeline catalog rows, authored topology inputs, generated topology/schedules, Playwright fixture support, `timeline-grid.spec.ts`, traffic setup, measurement summary v2 schema and finalization, and focused browser/harness tests.

**Required edits:**

1. Activate `fixture_profile_id` for the four rows in direct, slice, aggregate measurement, and release routing.
2. Replace `assembleAc043Fixture` with clone discovery and production-query validation.
3. Remove background-user creation and log in the 24 preseeded users.
4. Delete the per-predicate assembly path without alias or fallback.
5. Link each measurement summary v2 to its clone and snapshot artifact.
6. Fail target finalization on missing, corrupt, mismatched, unredacted, or overlapping provenance.
7. Prove traffic rate, foreground isolation, target-row exclusion, predicate independence, and cleanup remain unchanged.

**Expected diff:** Atomic catalog/browser cutover, generated routing updates, v2 evidence, and deletion of obsolete assembly code. The four scenario IDs, predicates, limits, and release posture do not change.

**Validation:** Direct-row plans; Timeline owner slice; harness browser slice; four-row service-backed aggregate; `make browser-e2e-measurement`; summary/artifact schema fixtures; traffic and cleanup tests; generated drift.

**Rollback point:** Revert the catalog and browser cutover atomically to the last known implementation. Do not ship dual routing, an alias, or a runtime fallback.

**Risks:** Direct-row and aggregate semantic divergence, foreground account omission, stored sessions in the template, profile inference, background traffic targeting measured rows, or accepting v1 evidence.

**Exit criteria:** Every route selects the same profile and key; no live assembly symbol or background account-creation path remains; all four summaries contain valid v2 provenance; one build and four clones are retained in aggregate evidence.

### WS-05: Qualification, release, and handoff

**Status:** `TODO`

**Depends on:** WS-04.

**Edit scope:** Validation and retained evidence only, followed by tracker and handoff updates. Product changes discovered by qualified failures return to the owning earlier workstream and repeat narrow validation before release.

**Execution order:**

1. Recheck the worktree and use task guides to select exact owner verification.
2. Run focused contract, contribution, broker, browser, and Timeline checks.
3. Run generation/drift and frontend/harness checks in section 8.
4. Run the four-row measurement aggregate once and inspect snapshot, clone, traffic, quietness, and timing evidence.
5. Run `make agent-finalize` with the appropriate retained result input when required by its contract.
6. Run one planned `make release-check`.
7. Do not automatically retry any failure.
8. Record every result root, failed unit, classification, skipped check, cleanup result, and next owner in the append-only log.

**Expected diff:** No source diff from qualification itself except intentional append-only documentation updates or an owner-routed fix begun as a separately recorded iteration.

**Validation:** Full matrix in section 8 and binary acceptance criteria in section 11.

**Rollback point:** Retained evidence is immutable and is never rolled back. If a source change fails, revert only the identified implementation slice or prepare a new owner-reviewed fix; do not mutate result roots.

**Risks:** Treating narrow success as release success, automatic rerun, ignoring snapshot provenance, combining object-store readiness with product latency, or optimizing without stage evidence.

**Exit criteria:** One current-source release run passes every selected unit, all snapshot and measurement artifacts qualify, cleanup is complete, and the handoff names all changes and evidence. If the planned release fails, this workstream remains open and the failure is preserved.

## 7. Checkpoint ledger

Each checkpoint is serial. The implementing agent appends a handoff row when a checkpoint begins, exits, or becomes blocked.

| Checkpoint | Edit scope | Validation | Expected diff | Rollback point |
| --- | --- | --- | --- | --- |
| CP-00 Tracker baseline | This file only | Markdown lint, diff check, status, complete review | One documentation file | Remove only the uncommitted tracker |
| CP-01 Authority and typed contract | Harness NLSpec, registry/schema, catalog field, generators/projections, validators/tests | Owner review, schema negatives, generation/drift, JSON shape, harness contract | One atomic owner-to-projection contract slice | Revert the full contract slice before consumers depend on it |
| CP-02 Contribution assembler | Test support, owner contributions, validators, deterministic tests | Exact counts, stable digest, graph negatives, security and no-session tests | Internal test-support slice only | Revert assembler; retain inactive contract if useful |
| CP-03 Broker lifecycle | Testservices, PostgreSQL support, graph prerequisite, leases, cleanup, artifact v1 | Single flight, isolation, corruption, injected failure, cancellation, janitor, redaction | Harness lifecycle and evidence slice | Revert broker/compiler consumers together; clean only leased resources |
| CP-04 Browser/catalog cutover | Four rows, topology inputs/generated outputs, E2E support, traffic, summary v2 | Direct and aggregate plans, owner slices, measurement target, provenance negatives | Atomic replacement of live assembly by snapshot discovery | Revert route and browser consumer together; no dual path |
| CP-05 Qualification and release | Commands, retained artifacts, tracker/handoff | Narrow qualification, finalizer, one release run | Evidence and documentation only unless a failure creates a new owner slice | Preserve all results; return code changes to owning checkpoint |

## 8. Validation matrix

### 8.1 Planning and owner routing

Run before choosing narrow rows or targets:

```text
make task-guide ROLE=module-author OWNER=harness.browser
make task-guide ROLE=module-author OWNER=module.timeline
make explain-test-owner OWNER=harness.browser
make explain-test-owner OWNER=module.timeline
```

Use `make explain-target TARGET=<target> DETAIL=rows` and `make target-plan` or `make target-plan-json` to confirm that direct rows, owner slices, aggregate browser measurement, and release compilation share the same snapshot prerequisite semantics.

### 8.2 Focused contract and implementation checks

The task guides determine exact rows. At minimum, the implementation validation includes:

```text
make test-slice OWNER=harness.browser
make test-slice OWNER=module.timeline
make service-backed-test-slice OWNER=harness.browser
make service-backed-test-slice OWNER=module.timeline
make frontend-typecheck
make frontend-unit
make harness-contract
make browser-e2e-measurement
```

When the owner slices are broader than the changed seam, use `ROWS=<row-id,...>` first, record the selection, then run the full owner slice when checkpoint risk requires it.

### 8.3 Required fixture-snapshot tests

| Test class | Required proof |
| --- | --- |
| Key identity | Same canonical inputs yield one key; each schema, migration, source-contract, fixture-version, or seed change changes the key; malformed input fails. |
| Contribution closure | Exact registered set succeeds; duplicates, omissions, cycles, unknown owners, incompatible versions, and contract drift fail before sealing. |
| Determinism | Repeated builds from the same migrated input produce exact counts and the same semantic digest. |
| Source invariants | Production authorization, owner validation, default view, projection, and relationship rules remain active. |
| No stored sessions | Template inspection proves no access, refresh, browser, or active session state exists. |
| Single flight | Concurrent requests for one key run one builder; different keys never share state. |
| Clone isolation | Four clones have a common validated parent key and cannot observe predicate-local mutations from each other. |
| Corruption and partial state | Missing artifact, unsealed template, count mismatch, digest mismatch, or partial contribution fails closed with no fallback. |
| Resource cleanup | Success, setup failure, validation failure, clone failure, measurement failure, cancellation, and finalizer failure leave no template, clone, session, incident lease, connection, or bucket. |
| Janitor safety | Stale resources are selected only through validated lease identity and scope; cleanup never targets a broad namespace. |
| Traffic | Twenty-four logged-in background analysts plus one foreground user are live; commits remain evenly staggered at 4.8 updates/s and exclude target rows. |
| Redaction | Artifacts contain allowed IDs, digests, counts, states, and ordinals only; forbidden payload and infrastructure values are rejected. |
| Route parity | Direct row, row slice, owner slice, aggregate measurement, and release plan resolve the same fixture profile and snapshot-key algorithm. |

### 8.4 Generation and drift

```text
make generate
make generate-drift
make generated-artifact-policy-check
make json-shape-check
```

Run any catalog, schedule, or topology drift checks selected by `make task-guide` and the repository's current public target inventory. Generated files are reviewed but never hand-edited.

### 8.5 Finalization and release

After every narrow check and the four-row aggregate pass on current source:

```text
make agent-finalize
make release-check
```

`make release-check` runs once as the planned qualification attempt. A failure remains the release result. The implementing agent diagnoses it from the retained root and does not issue an automatic retry. A later run is a separately planned validation after an identified change, with its own handoff entry and unchanged earlier evidence.

### 8.6 Tracker-only validation

Creating this tracker changes documentation only. Its checkpoint runs:

```text
make lint-markdown
git diff --check
git status --short
```

The author also reviews the complete diff. Implementation, contract, generated, browser, and release checks are intentionally not run for the tracker-only checkpoint because this file is not executable input.

## 9. Qualification and failure routing

The narrow measurement is qualified only when all four predicates use the expected snapshot key, exact fixture digest, distinct clone proofs, one warm-up, 100 finite measured samples, exact traffic, zero scheduler overlap, and schema-valid redacted evidence.

A qualified latency breach remains a product and release failure. Route it by the dominant retained stage:

| Dominant evidence | Owning investigation |
| --- | --- |
| Driver dispatch or invalid quietness | Harness or environment qualification |
| Accepted action to request | UI event handling or mutation queueing |
| Request round trip | Server, database, or transport telemetry |
| Response decode or client apply | Client normalization and state application |
| Apply to visible paint | React/grid rendering, virtualization, layout, or paint |

Optimize product code only after qualified evidence identifies the repeatedly dominant stage. Revalidate the changed owner slice and repeat the planned qualification sequence; preserve the original breach.

An object-store readiness recurrence is a distinct harness or infrastructure failure. Preserve its target result, readiness stage, attempt count, and cleanup evidence. Do not combine it with Timeline latency, reclassify it as measurement jitter, or automatically rerun the release.

## 10. Risks, blockers, and compatibility posture

### 10.1 Risk register

| ID | Risk | Prevention or mitigation | Closure evidence |
| --- | --- | --- | --- |
| R-01 | Suite snapshot is mistaken for forbidden result caching | Adopt explicit harness language and keep snapshot lifetime within one invocation | NLSpec review and cross-run reuse negative test |
| R-02 | Fixture profile selection depends on unstable strings | Require catalog `fixture_profile_id` and typed registry resolution | Catalog negatives and route-parity tests |
| R-03 | Source-owner rules are bypassed for speed | Require production services/validators and prohibit raw E2E SQL | Import/boundary tests and owner review |
| R-04 | Nondeterministic IDs, clocks, or ordering change semantics | Seed deterministic inputs and hash safe semantic receipts | Repeated-build digest test |
| R-05 | Sessions or credentials are cloned | Seed accounts/memberships only and validate no-session state before sealing | No-session and redaction tests |
| R-06 | PostgreSQL template has active connections or is partially sealed | Close and verify connections, then mark sealed only after validation | Injected sealing failure and cleanup tests |
| R-07 | Predicate clones share mutable state | One clone and bucket per row with clone proof and mutation isolation test | Four-clone aggregate artifact |
| R-08 | Cleanup deletes unrelated resources | Lease-scoped opaque identities and bounded janitor selection | Janitor safety tests and cleanup artifact |
| R-09 | Runtime improvement weakens measured load or predicates | Preserve exact Core 04 projection and compare qualification fields | Four v2 summaries and traffic proof |
| R-10 | Narrow pass is reported as release qualification | Separate tracker rows and require current-source full release root | AC043-009 exit evidence |
| R-11 | A failed release is retried without a corrective decision | One planned run and append-only evidence policy | Handoff log contains one attempt and disposition |
| R-12 | Snapshot artifact leaks sensitive or infrastructure data | Allowlist schema fields and fail finalization on forbidden content | Redaction negative tests |

### 10.2 Compatibility and migration

This is an internal benchmark-contract and harness-lifecycle cutover:

- Four catalog rows gain an explicit fixture profile.
- Generated topology gains a shared snapshot-builder prerequisite and profile metadata.
- Browser support stops assembling fixture data and stops creating background users.
- Current qualified summaries move from v1 to v2 and link a new snapshot artifact.
- Historical row identities and v1 result artifacts remain readable evidence.
- No runtime alias or fallback preserves the old assembly path.
- No public HTTP API, storage schema, database migration, production fixture, threshold, retry policy, or Core 05 claim changes.

Dashboard or result consumers that display current qualified AC-043 evidence must accept v2 and its snapshot link. They may continue to render v1 as historical, explicitly non-snapshot-qualified evidence.

### 10.3 Current blockers

No external blocker is recorded at tracker creation. The absence of a green current-source release is unfinished validation, not a reason to weaken acceptance. Any new blocker records its exact failing command, result root or artifact, affected work item, owner, cleanup state, and next safe action in the handoff log.

## 11. Binary acceptance criteria

The tracker closes only when every criterion is true:

- [ ] AC043-SR-AC-001: Core 04 and the Testing Harness NLSpec have non-overlapping, reviewed ownership for fixture meaning and snapshot mechanics.
- [ ] AC043-SR-AC-002: `ac043_large_grid_snapshot_v1` exists in one typed registry and all four rows bind it through `fixture_profile_id`.
- [ ] AC043-SR-AC-003: Snapshot key generation uses canonical JSON and exactly the specified schema, migration digest, source-contract digest, fixture version, and seed.
- [ ] AC043-SR-AC-004: The closed source-owner contribution assembler uses production services and validators and contains no raw E2E SQL path.
- [ ] AC043-SR-AC-005: The populated template has exact Core 04 counts and distribution, deterministic semantic digest, enabled security controls, and no stored sessions.
- [ ] AC043-SR-AC-006: The four-row aggregate builds one populated template and creates four isolated predicate clones.
- [ ] AC043-SR-AC-007: Same-key requests are single-flight; corrupt, partial, unsealed, absent, or mismatched snapshots fail closed without live assembly.
- [ ] AC043-SR-AC-008: Success, failure, cancellation, and stale-resource tests leave no populated template, clone, connection, session, or object-store bucket.
- [ ] AC043-SR-AC-009: `assembleAc043Fixture` and background account creation are removed without alias or fallback.
- [ ] AC043-SR-AC-010: Direct-row, row-slice, owner-slice, aggregate measurement, and release graphs resolve identical snapshot semantics.
- [ ] AC043-SR-AC-011: Every current measurement emits a valid redacted snapshot v1 artifact and frontend measurement summary v2 cross-reference.
- [ ] AC043-SR-AC-012: Historical v1 evidence remains readable but cannot close current snapshot qualification.
- [ ] AC043-SR-AC-013: Selection, focus, and typing pass at 100 ms p95 and blank-row creation passes at 150 ms p95, each with one warm-up and 100 measured samples.
- [ ] AC043-SR-AC-014: Qualification proves 25 live analysts, 4.8 committed background updates per second, target-row exclusion, and zero ordinary scheduler overlap.
- [ ] AC043-SR-AC-015: Focused owner, contract, frontend, harness, browser, generation, drift, schema, cleanup, and redaction checks pass.
- [ ] AC043-SR-AC-016: `make agent-finalize` passes with its required retained-evidence posture.
- [ ] AC043-SR-AC-017: One planned current-source `make release-check` passes every selected release unit without an automatic retry.
- [ ] AC043-SR-AC-018: The final handoff lists changed files, removed paths, contract/schema versions, catalog/profile changes, commands, result roots, failures, skipped checks, cleanup state, and remaining risks.

If AC043-SR-AC-017 is false, no green release is claimed and AC043-009 through AC043-010 remain open.

## 12. Append-only handoff log

Existing rows are never edited to rewrite history. Corrections are new rows that name the superseded statement.

| Date | Checkpoint | Status change | Branch and commit | Work completed | Validation and evidence | Risks, blockers, and cleanup | Next action and owner |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-13 | CP-00 | WS-00 `IN_PROGRESS`; AC043-001..003 retained `DONE`; AC043-004..010 initialized `TODO` | `main` at `cc76c32682bc35cc766e500e3b5efd60be398907`; worktree clean before tracker creation | Recorded authority, target interfaces, serial workstreams, failure policy, binary acceptance, and immutable evidence posture | Failed release root `20260813T194448Z-p3008841`: 925/929, 1,979,786 ms. Narrow root `20260813T212600Z-p3589519`: 18/18, 1,083,828 ms, four qualified predicates and zero overlap. Tracker checks pending. | No green current-source release. Historical release predates final fixes and includes a separate object-store readiness failure. No retained result was mutated. | Complete CP-00 documentation checks, append the result, then begin WS-01 with Testing Harness and Core 04 projection review. |
| 2026-08-13 | CP-00 | WS-00 `IN_PROGRESS` to `DONE` | `main` at `cc76c32682bc35cc766e500e3b5efd60be398907`; only this tracker untracked | Completed the tracker-only checkpoint; no specification, contract, generated, implementation, test, or retained-result file was changed | `make lint-markdown` passed; run root `20260813T220910Z-p3633404`. `git diff --check`, final status, and complete-diff review are recorded in the delivery handoff. | No cleanup required. No green current-source release is claimed. | Begin WS-01 with Testing Harness owner review; implementation owner records a new baseline before editing. |

### 12.1 Handoff entry requirements

Every future entry includes:

- the work item and checkpoint entered or exited;
- branch, commit, and initial worktree state;
- files or ownership surfaces changed;
- exact validation commands and outcome;
- retained result root and relevant artifact references;
- failing unit and classification when any command fails;
- snapshot key, builder count, clone count, and cleanup state when applicable;
- any skipped check and the reason;
- whether a release run was planned and whether it was the only attempt;
- unresolved risks or blockers; and
- one explicit next action with owner.

## 13. Completion and closure rule

This tracker is complete only after AC043-001 through AC043-010 are `DONE`, all binary acceptance criteria are checked, the final current-source release root is retained and green, and cleanup is proven. The final entry records that result without deleting or diminishing the failed `20260813T194448Z-p3008841` release or the narrow `20260813T212600Z-p3589519` corroboration.

If the planned release fails, the tracker remains open. The handoff records the failure as product, harness, infrastructure, evidence-integrity, or cleanup work according to retained evidence; assigns it to the owning workstream; and does not authorize retries, threshold changes, percentile slack, failure reclassification, Core 05 promotion, or fallback fixture assembly.
