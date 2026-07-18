# Test Harness Subsystem Migration Refactor Tracker and Handoff

## 1. Tracker control

| Field | Value |
| --- | --- |
| Target | Test-harness execution, accounting, diagnostics, generated topology, and phase-shaped test sources |
| Tracker | `docs/handoffs/test-harness-subsystem-migration-refactor-tracker.md` |
| Baseline date | 2026-07-17 |
| Baseline branch | `revision/grid-adapter` |
| Baseline commit | `37cfdd727b3172046fbc3c5194d896a1197a381c` |
| Baseline worktree | Clean |
| Tracker state | `IN_PROGRESS` — WS-07 owner slice and evidence APIs |
| Active start | Clean tree at `ddfd2afa` |
| Active tasks | T-033 |
| Migration mode | Hard cutover; no aliases, compatibility readers, dual catalogs, or retained phase interfaces |
| Completion model | Binary; partial owner adoption is not a releasable end state |

This tracker is the resumable implementation authority for migrating the test harness from delivery-phase organization to logical owner-module-family organization. It records the required decisions, dependency order, evidence, deletion obligations, and completion gates. It does not itself change product behavior or claim conformance.

The inspected baseline contains 456 backend phase-map rows, 87 frontend phase-map rows, and 5 distinct Graph Projection subsystem rows, for a provisional reconciliation population of exactly 548 legacy identities. The 26 checked-in Markdown coverage ledgers are deletion inventory and are not part of that population. WS-00 MUST reproduce `456 + 87 + 5 = 548` from the frozen baseline before the first migration edit. It may change 548 only by proving identity overlap in the schema-valid baseline report and revising every dependent total before downstream migration begins.

### Status vocabulary

Use only these work-item states:

- `TODO`: not started.
- `IN_PROGRESS`: actively owned in the current handoff.
- `BLOCKED`: cannot advance without an identified decision or prerequisite.
- `DONE`: exit condition and evidence are both satisfied.
- `DROPPED`: inspected and intentionally removed with a Section 9 closed reason and owner approval.
- `DEFERRED`: valid work outside this tracker, with an explicit owner and follow-up record; it MUST NOT cover a Section 15 completion obligation.

`DONE` never means “code was written.” It means the named validation evidence exists and no listed deletion or reconciliation obligation remains.

## 2. Authority, scope, and decision record

### 2.1 Source precedence

| Rank | Source | Role in this migration |
| --- | --- | --- |
| 1 | Core 00 through Core 04 | Normative current product behavior, architecture, security, and implementation-conformance authority. |
| 2 | Adopted subsystem NLSpecs | Normative only for the subsystem each document explicitly owns. Draft subsystem documents do not create implementation obligations. |
| 3 | `docs/testing-harness-nlspec.md` | Final-state v2 owner of harness command invocation, catalog and runner validation, owner/row selection, scheduling, fixtures, artifacts, evidence audit, cleanup, and verification gates. Its `adopted/current` status is valid only in the atomic cutover; the document change MUST NOT merge while the running surface remains v1. |
| 4 | Core 05 | Authority only for claim-bearing timed, benchmark, fixture-sensitive, or publication evidence. It is not general harness architecture authority. |
| 5 | `docs/domain.md` | Vocabulary and concept-boundary reference. It distinguishes product/domain phase concepts from historical delivery-phase labels. |
| 6 | Machine contracts under `contracts/` and authored owner inputs under `tools/` | Executable derived validation inputs. They implement adopted requirements but do not supersede normative owner documents. |
| 7 | `docs/design.md` and guides | Design direction and implementation support. They are not Base Profile or extension-profile conformance evidence. |
| 8 | Current code and tests | Evidence of current implementation, not evidence that unsupported behavior must survive. |
| 9 | Prior trackers, archived plans, and the modular-refactor framework | Historical evidence and general guidance only. |

If current sources conflict at the same authority level, mark the affected item `BLOCKED: owner contradiction`; do not encode a guessed resolution into a machine contract. If a test validates behavior unsupported by the current owners, delete or narrow the test rather than preserving it as a regression by inertia.

### 2.2 Locked decisions

These decisions may not be reopened by an implementation slice without updating this tracker and obtaining explicit authority:

1. The migration includes harness/accounting code and phase-shaped test filenames, symbols, browser titles, fixtures, and golden paths.
2. Authored production-source phase names are separate module-owner follow-ups unless a harness or test selector cannot be decoupled without changing one. Generated downstream code is never hand-edited.
3. The cutover is hard. There are no old-name aliases, fallback readers, dual-written manifests, compatibility schema unions, or temporary public phase targets in the merged state.
4. All 26 generated Markdown phase ledgers are deleted. Owner manifests and machine-generated diagnostics replace their operational use; no owner-ledger document format is introduced.
5. Tests and validation tools do not open documentation or specification files. Executable requirements are represented by machine-readable owners outside `docs/` and `docs/spec/`.
6. The generic modular-refactor framework remains useful, but its allowance for retaining phase-based evidence is superseded for this subsystem migration.
7. Earlier handoff decisions that preserve phase paths or titles, including the web-E2E refactor tracker, are superseded only for the harness and test-source surfaces covered here.
8. Historical retained runs using phase schemas are investigation-only after cutover and cannot close current verification gates.
9. The cutover publishes `docs/testing-harness-nlspec.md` as `cartulary.testing_harness.v2` with conformance profile `cartulary.testing_harness.current.v2` and `status: adopted/current`; no implementation slice may claim that adoption before the atomic cutover criteria are satisfied.
10. `docs/domain.md` requires no change. It already excludes implementation modules and test rows from product-domain vocabulary; the harness NLSpec owns the local owner, family, row, verification, runner, and profile terms.

### 2.3 In scope

- Harness command surface, helper ownership, target planning, selection, scheduling, execution, fixtures, cleanup, diagnostics, accounting, artifact emission, finalization, and retained-run auditing.
- Backend and frontend phase registries, maps, joins, policy exceptions, schedules, ledgers, and generated topology derived from phase identity.
- The Graph Projection subsystem test registry and its loader, which must be absorbed rather than maintained as a second catalog.
- Phase-shaped Go, Vitest, Playwright, shell, fixture, visual, accessibility, and measurement test paths and identifiers.
- Direct or indirect documentation parsing performed to determine whether a test passes, whether a row is owned, or whether evidence is complete.
- Public Make targets, input variables, schemas, artifact names, duration baselines, repository procedures, and active guides affected by the new owner-first surface.
- Deletion of unsupported, draft-only, duplicate, vestigial, or production-irrelevant test behavior discovered during row adjudication.

### 2.4 Out of scope

- New product behavior, public route changes, protocol changes, authorization changes, or new conformance claims.
- Renaming legitimate runtime/domain uses of “phase,” including incident phase, job progress phase, operator phase, or the normatively named Workbook Import Assistant phase.
- A general production-source phase-name refactor. Such findings go into the module-owner follow-up ledger in Section 10.
- Hand editing generated roots, generated topology outputs, `go.sum`, `pnpm-lock.yaml`, or tool-managed dependencies.
- Converting human-readable specifications into executable code generation from Markdown. Machine contracts are reviewed derived inputs, not Markdown-extraction caches.
- Publishing new timed or visual claims. Existing claim boundaries remain governed by Core 05.

### 2.5 Terminology guardrail

In this tracker, “delivery phase” means the historical `phase0` through `phase12` or `FE-P0` through `FE-P12` accounting identity. “Execution step” means a scheduler state or command lifecycle stage. “Product phase” means a domain or operator value whose semantics are owned by a product contract. Automated anti-drift checks must distinguish these meanings; a raw repository-wide ban on the word `phase` is incorrect.

## 3. Current-state baseline

### 3.1 Inventory summary

| Surface | Baseline | Migration implication |
| --- | ---: | --- |
| Backend phase registry | 13 active entries | Delete after all rows move to owner manifests. |
| Backend phase maps | 13 files / 456 rows | Every row receives one terminal crosswalk disposition. |
| Frontend phase registry | 13 active entries | Delete cumulative dependencies, guide digests, and base-phase joins. |
| Frontend phase maps | 13 files / 87 rows | Merge runner-neutral ownership into the same owner catalog as backend rows. |
| Checked-in Markdown coverage ledgers | 26 files | Delete with their generators, drift checks, digest fields, and ignore rules. |
| Harness implementation files | 350 tracked files at baseline | Reclassify by semantic family; do not move files merely to satisfy a directory aesthetic. |
| `tools/harness/phase-accounting` | 58 files at baseline | Split retained behavior among catalog, execution, evidence accounting, diagnostics, and generated-artifact owners; then delete the directory. |
| Subsystem test registries | Graph Projection only / 5 rows | Freeze all five identities, migrate them into the unified catalog, then delete the subsystem registry and special loader. |
| Vitest accounting classifications | 228 entries, including 142 `unowned_regression` entries at discovery | Adjudicate each entry; no unowned catch-all survives. |
| Phase-shaped test references | Broadly present in Go, Vitest, Playwright, shell, fixtures, and goldens | WS-00 freezes the exact path/symbol/title population before renaming. |
| Public Make targets | Phase slices, phase diagnostics, phase ledgers/schedules, and frontend phase audit are exposed | Replace atomically with owner-based targets. |

The inventory counts above were gathered by read-only inspection. They are a checkpoint seed. The implementation must create an exact machine-readable crosswalk from the live baseline rather than treating prose totals as sufficient reconciliation evidence.

### 3.2 Current phase-owned inputs

| Input family | Current schema or role | Coupling to remove |
| --- | --- | --- |
| `tools/phase_registry.json` | `cartulary.phase_registry.v1` | Ordered delivery identity, ledger paths, and prose owner references. |
| `tools/phase0_test_map.json` through `tools/phase12_test_map.json` | `cartulary.phase_test_map.v2` | Backend test ownership and execution selection by phase. |
| `tools/frontend_phase_registry.json` | `cartulary.frontend_phase_registry.v5` | Guide path/digests, cumulative dependencies, ledger digests, and base-phase joins. |
| `tools/frontend_phase_maps/*.json` | `cartulary.frontend_phase_test_map.v4` | Frontend row ownership, guide headings, requirements, and phase joins. |
| `tools/phase_policy_exceptions.json` | Empty exception mechanism | Delete instead of carrying an unused policy layer forward. |
| `tools/subsystem_test_registry.json` | `cartulary.subsystem_test_registry.v1` | Parallel special-case ownership model used only by Graph Projection. |
| `tools/test_accounting_classification.json` | Vitest/Go accounting classification | `unowned_regression` and phase-derived classification. |
| `tools/task_surface_owner.json` | Public target and input owner | Phase targets, variables, role names, scripts, and help text. |
| `tools/execution_topology_manifest.json` | Authored topology owner | Phase ledger/schedule nodes and phase-derived work units. |

### 3.3 Current public surface and required disposition

| Current surface | Disposition | Successor |
| --- | --- | --- |
| `make phase-slice PHASE=<phase>` | Delete | `make test-slice OWNER=<owner-id> [ROWS=<row-id,...>]` |
| `make service-backed-slice PHASE=<phase>` | Delete | `make service-backed-test-slice OWNER=<owner-id> [ROWS=<row-id,...>]` |
| `make explain-phase PHASE=<phase>` | Delete | `make explain-test-owner OWNER=<owner-id> [JSON=1]` |
| `make task-guide ROLE=phase-author ...` | Replace role and inputs | `make task-guide ROLE=module-author OWNER=<owner-id> [JSON=1]` |
| `make frontend-evidence-audit PHASE_NAMESPACE=frontend PHASE=<phase>` | Delete | `make test-evidence-audit OWNER=<owner-id> CHECK_RESULTS_DIR=<root> ...` |
| `make phase-ledgers` | Delete without replacement | On-demand owner diagnostics read machine manifests. |
| `make phase-ledger-drift` | Delete without replacement | `test-catalog-check` plus schema and generated-artifact drift checks. |
| `make phase-schedules` | Delete as a public target | Necessary topology generation becomes part of `make generate`. |
| `make phase-schedule-drift` | Delete as a public target | Necessary topology drift becomes part of `make generate-drift`. |
| Internal `phase-map-check` | Delete | Internal/check-level `test-catalog-check`. |
| Internal `phase-test-name-check` | Replace | Semantic owner/test-name check with product-phase allow rules. |
| `PHASE`, `PHASE_NAMESPACE` | Delete | `OWNER`; `ROWS` remains an optional exact row selector. |
| `phase-author` | Delete | `module-author`. |

The owner-slice targets retain the current optional worker controls where applicable: `VITEST_MAX_WORKERS`, `PLAYWRIGHT_WORKERS`, and `JSON`. Public input validation must reject removed phase variables instead of silently ignoring them.

### 3.4 Current schema and artifact debt

The following identities are representative mandatory-removal items. WS-00 must expand this into an exhaustive generated reference list before edits.

| Old identity | Target identity or action |
| --- | --- |
| `cartulary.phase_slice_plan.v2` | `cartulary.test_slice_plan.v1` |
| `cartulary.phase_slice_scheduler_summary.v4` | `cartulary.test_slice_scheduler_summary.v1` |
| `cartulary.frontend_row_accounting.v5` | `cartulary.test_evidence_accounting.v1` |
| `cartulary.frontend_evidence_audit_summary.v1` | `cartulary.test_evidence_audit_summary.v1` |
| `cartulary.test_phase_summary.v3` | `cartulary.test_owner_summary.v1` |
| `phase-summary.json` where it means command lifecycle | `step-summary.json` |
| `phase-slice-plan.json` | `test-slice-plan.json` |
| `phase-runtime.sh` where it is a generic command runner | `command-runtime.sh` |
| `manifest_phase`, `phase_id`, `selected_phase` | `owner_id` and, where needed, `family_id` |
| `phase_namespace`, `base_phase_join` | Delete; there is one catalog and no backend/frontend ownership namespace split. |

Old schema IDs are not accepted in new readers. Old retained artifacts can be displayed by historical tooling outside release closure, but no compatibility parser is added to production harness code.

## 4. Target architecture and machine contracts

### 4.1 Boundary and data flow

```text
Core 00-04 and adopted subsystem NLSpecs
                 |
                 | reviewed derivation; never runtime Markdown parsing
                 v
contracts/verification/registry.json
contracts/verification/owners/*.json
                 |
                 v
tools/test_catalog_owner.json ---> tools/test_families/*.json
                 |                         |
                 +-----------+-------------+
                             v
                test-catalog validation and selection
                             |
                             v
                  owner-slice execution planning
                             |
                             v
             generic scheduler, fixtures, and runner adapters
                             |
                             v
          step summaries and owner-based evidence accounting
                             |
                 +-----------+-----------+
                 v                       v
       retained-run audit        diagnostics / task guide

Authored owner inputs --Make generation--> generated topology and schedules
```

Documentation explains intent and remains review authority, but no arrow from `docs/` enters executable validation. Generated topology is downstream of authored machine owners and is changed only through Make-owned generation.

### 4.2 Harness module families

| Family | Responsibility | May depend on | Must not own |
| --- | --- | --- | --- |
| `test-catalog` | Load and validate owner/family manifests, verification references, selectors, evidence classes, fixture/runtime profiles, and resources. | Machine schemas, verification contracts, runner-neutral types. | Process execution, service lifecycle, Markdown parsing, scheduling policy. |
| `execution` | Convert selected catalog rows into semantic work units and invoke runner adapters. | `test-catalog`, scheduler interface, output interface. | Owner inference, evidence publication, generated-file editing. |
| `evidence-accounting` | Record selected/executed/passed/failed/skipped rows, validate retained evidence, and emit owner summaries. | Catalog snapshots, step summaries, artifact schemas. | Test discovery, command execution, documentation-derived ownership. |
| `scheduler` | Generic DAG, resource, concurrency, dependency, cancellation, and cleanup coordination. | Work-unit contracts and lifecycle services. | Delivery phases, module ownership policy, runner-specific manifest parsing. |
| `diagnostics` | Explain owners, rows, target plans, runs, and recommended narrow commands. | Read-only catalog, topology, and evidence views. | Alternate selection rules or private manifest fallbacks. |
| `generated-artifacts` | Render and drift-check topology outputs from authored owner inputs. | Catalog/topology schemas and generator policy. | Authored ownership, Markdown ledgers, hand-edited outputs. |
| Existing backend/browser/contract/readiness families | Runner-specific execution and evidence behavior already coherently owned. | Execution contracts and shared lifecycle interfaces. | Phase identity or duplicate catalog loading. |

`tools/harness` remains the subsystem root. The current `phase-accounting` root is deleted after its retained responsibilities are placed in the families above. No new catch-all facade may re-export all internal families.

### 4.3 Canonical owner and family IDs

The schemas and validator MUST implement this grammar exactly:

```text
segment         = [a-z][a-z0-9_]{0,62}
owner_id        = (module|platform|app|web|package|harness) "." segment
family_id       = owner_id "." segment
row_id          = family_id "." segment
verification_id = owner_id ".verification." segment
```

Every complete ID is at most 191 ASCII bytes. Unicode, whitespace, empty segments, slashes, backslashes, percent escaping, shell metacharacters, and any segment matching `phase[0-9]+` or `fe_p[0-9]+` case-insensitively are invalid. IDs are globally unique within their category, immutable, never recycled, and serialized in ascending ASCII-byte order. An ownership move creates a new ID plus a crosswalk entry; runtime aliases are forbidden. Omitted display metadata renders as the machine ID and never participates in semantic identity or digests.

### 4.4 Verification contract model

Create these authored inputs:

- `contracts/verification/registry.json`: registry of adopted machine verification owners and contract paths.
- `contracts/verification/owners/<owner-id>.json`: verification entries grouped by normative product or support owner.
- Registered JSON schemas under `tools/schemas/` for both files.

Each verification entry contains:

| Field | Rule |
| --- | --- |
| `verification_id` | Stable, globally unique, owner-qualified identifier. |
| `owner_id` | One current product or support owner. |
| `behavior_class` | One of `product`, `security`, `architecture`, `build`, `harness`, `claim_publication`. |
| `profile` | Exactly `base`, `support`, `extension.<profile_id>`, or `claim.<profile_id>`. |
| `requirement` | Concise machine-owned assertion semantics; not copied prose used as a document checksum. |
| `evidence_kinds` | Allowed evidence types, such as Go test, Vitest, Playwright, static check, artifact audit, or release check. |
| `status` | `active` only in executable catalogs. Draft or unsupported entries are not executable. |
| `documentation_refs` | Optional inert strings. Consumers may emit or compare them but MUST NOT open, stat, resolve, or hash the referenced documents. |

A verification contract is a reviewed derived contract. A change to one requires review against its normative owner, schema validation, row-reference validation, and relevant behavior tests. It must not be regenerated by scraping headings, requirement IDs, prose, TODO markers, or file digests from documentation.

`behavior_class` is closed to `product`, `security`, `architecture`, `build`, `harness`, and `claim_publication`. Product and security behavior uses `base` or an adopted `extension.*` profile; build and harness behavior uses `support`; claim-publication behavior uses `claim.*` and resolves to an active Core 05 claim. Claim posture and profile MUST agree. Informative evidence cannot close conformance or release, and informative measurement defaults to `default_check=false`. The default skip policy is `forbid`. An authorized skip requires an exact reason, verification owner, approval evidence, and expiry; an expired authorization is invalid.

### 4.5 Test catalog model

Create these authored inputs:

- `tools/test_catalog_owner.json`, schema `cartulary.test_owner_registry.v1`.
- `tools/test_families/<owner-id>.json`, schema `cartulary.test_family_manifest.v1`.

The registry contains `owner_id`, manifest path, status, and optional display metadata. It contains no execution order, ledger path, document digest, phase join, or cumulative dependency chain.

Each manifest row contains:

| Field | Rule |
| --- | --- |
| `row_id` | Stable, globally unique, semantic ID with no phase token. |
| `owner_id` | Exactly one primary owner, equal to the manifest owner. |
| `family_id` | Semantic owner-qualified family. |
| `collaborator_ids` | Required sorted unique owner IDs; the array may be empty. |
| `verification_ids` | Required, sorted, duplicate-free, nonempty active machine verification IDs. |
| `runner` | Exactly `go`, `vitest`, `playwright`, or `shell` in the v1 runner registry. |
| `selector` | Runner-specific exact package/path/name/tags or target selector; no prose discovery. |
| `evidence_class` | `unit`, `integration`, `browser`, `accessibility`, `visual`, `measurement`, `static`, `security`, or `release`. |
| `runtime_profile_id` | One registered runtime profile ID. |
| `resource_profile_id` | One registered resource profile ID; capacity remains owned by the scheduler resource registry. |
| `fixture_profile_id` | One registered fixture profile ID. |
| `default_check` | Whether the row participates in the default check topology. |
| `claim_posture` | `implementation`, `informative`, or an explicit Core 05 publication profile. |
| `status` | Exactly `active`. |

Rows MUST NOT embed commands, ports, capacities, service topology, environment variables, or fixture paths. The authored execution-topology owner supplies exhaustive closed `runtime_profiles`, `resource_profiles`, and `fixture_profiles` collections before adoption. The loader validates all manifests as one catalog and rejects unknown properties, enum values, profiles, runners, duplicate rows, unresolved or multiply resolved references, unsorted or duplicate arrays, unsupported or ambiguous selectors, documentation-backed ownership, inactive requirements, and an owner with zero executable rows. There is no subsystem exception path and no `unowned_regression` classification.

The runner registry closes selectors as follows: Go uses a repository package plus a nonempty sorted array of exact top-level `Test...` symbols; Vitest uses a repository-relative test file plus nonempty exact full test titles; Playwright uses a repository-relative file, project ID, stage, stable scenario IDs, and diagnostic titles; shell uses a registered stable command ID and forbids raw shell, argv, and executable paths. Playwright stage is exactly `webserver_backed`, `stateful`, `support`, `visual`, `accessibility`, or `measurement`. Preflight rejects zero or multiple resolution, active-row overlap, globs, regular expressions, missing paths, symlink escape, and paths outside approved roots before setup. A new runner requires an adopted registry revision, selector and result schemas, an allowlisted checked-in adapter, negative fixtures, and an NLSpec revision; dynamic plugin or package loading is outside v2.

### 4.6 Owner selection and evidence rules

1. Assign the owner of the normative verification postcondition.
2. For cross-module integration, assign the owner of the externally visible postcondition or primary durable mutation.
3. For mechanism-only evidence, assign the platform or harness mechanism owner.
4. Record every other participant as a collaborator.
5. If these rules do not produce one owner, mark the row `BLOCKED`; filename, package, runner, and maintainer are never tie-breakers.
6. Record the governing requirement or support contract, rule applied, collaborators, review revision, and owner-review evidence for every adjudication.
7. Backend and frontend rows share one ownership namespace. Runner and evidence class determine execution, not a frontend registry.
8. Scheduler dependencies use work-unit/resource relationships, not owner ordering or historical phase order.
9. Every selected row emits exactly one terminal record from `passed`, `failed`, `infrastructure_failed`, `skipped_dependency`, `cancelled`, or `skipped_authorized`.
10. Evidence audits compare retained semantic identity and the selected row set with emitted row records. They do not infer expected rows from filenames, time, or documentation.
11. Rows that only prove stale documentation text, obsolete scaffolding, or unsupported behavior are `DROPPED`, not migrated.

## 5. Delivery-phase to owner-family migration map

This table provides the mandatory first-pass destinations. WS-00 and the row crosswalk decide the exact owner for every row; a historical phase never becomes a target owner.

| Current source | Required candidate owner families | Adjudication notes |
| --- | --- | --- |
| `phase0` / `FE-P0` | `app.server`, `app.migrate`, `platform.bootstrap`, `platform.config`, `platform.postgres`, `platform.objectstore`, `harness.startup` | Separate product startup behavior from repository readiness and harness smoke checks. |
| `phase1` / `FE-P1` | `module.auth`, `platform.authn`, `platform.httpauth`, `web.auth` | Deployment-admin controls remain platform/security evidence; UI login behavior belongs to web auth. |
| `phase2` / `FE-P2` | `module.incidents`, `module.extensions`, `module.savedviews`, `web.shell`, `web.workbook` | Membership and preferences follow their actual module owner; reserved extension dispatch is not general incident ownership. |
| `phase3` / `FE-P3` | `module.timeline`, `module.workbook`, `module.projections`, `module.collaboration`, `package.grid_adapter`, `web.workbook` | Split grid choreography from Timeline behavior and projection refresh assertions. |
| `phase4` / `FE-P4` | `module.entities`, `module.indicators`, `module.parties`, `module.assessments`, `module.evidence`, `module.links`, `module.workbook`, `web.inspector` | Large mixed phase; row-level adjudication is mandatory and phase-wide moves are forbidden. |
| `phase5` / `FE-P5` | `module.evidence`, `platform.objectstore`, `web.evidence` | Preserve safe-handle and access-control boundaries; storage adapter tests are not Evidence domain tests. |
| `phase6` / `FE-P6` | `module.collaboration`, `module.workbook`, `platform.ws`, `web.collaboration` | Conflict semantics, pending edits, transport behavior, and UI presence have separate primary owners. |
| `phase7` / `FE-P7` | `module.revisions`, `module.records`, `web.history` | History, rollback, delete, and restore assertions follow the mutation/revision owner they principally validate. |
| `phase8` / `FE-P8` | `module.links`, `module.savedviews`, `module.projections`, `module.viewschemas`, `platform.viewquery`, `web.workbook` | Query plumbing and projection semantics must not be collapsed into saved-view ownership. |
| `phase9` / `FE-P9` | `module.workbook`, `module.indicators`, `module.parties`, `module.assessments`, `module.tasksdecisions`, `package.grid_adapter`, `web.workbook` | Keyboard/clipboard tests belong to grid or web interaction owners; record behavior follows domain modules. |
| `phase10` / `FE-P10` | `module.recovery`, `app.operator`, `platform.config`, `platform.objectstore`, `app.serverprocess`, `web.recovery` | Public route absence, backup semantics, process orchestration, and UI restore flows remain distinct. |
| `phase11` / `FE-P11` | `module.imports`, `module.reporting`, `module.reportcomposition`, `module.reference_data`, `module.incidentportability`, `module.incidentbundles`, `module.jobapi`, `platform.jobs`, `platform.enterpriseauth`, `module.extensions`, `web.extensions` | Admit only behavior supported by current Core/adopted profiles; draft subsystem documents do not preserve rows. |
| `phase12` / `FE-P12` | `module.networkflow`, `module.graphprojection`, `module.indicators`, `module.extensions`, `web.networkflow` | Graph Projection owns only its adapter/projection interface; Network Flow owns its extension workflow. |
| Graph Projection subsystem map | `module.graphprojection` | Migrate into the unified registry and delete the special subsystem loader and registry. |
| Tooling/support classifications | Appropriate `harness.*`, `platform.*`, `package.*`, or `app.*` owner | Every support row needs a named support verification contract; no generic unowned bucket. |

Owner IDs in the crosswalk must resolve against the live owner registry. If the correct destination is not represented by an existing module boundary, record `BLOCKED: owner boundary missing`; do not assign the nearest convenient module.

## 6. Documentation and specification decoupling ledger

The following known couplings are mandatory migration items. WS-00 must search for additional direct reads, path joins, digests, headings, requirement-ID extraction, source locators, and documentation-derived generated data.

| Current consumer/input | Current coupling | Required disposition | Target machine owner | Status |
| --- | --- | --- | --- | --- |
| `tools/harness/tests/test-harness-contracts.mjs` | Parses the harness NLSpec command registry/input matrix and Network Flow source text. | Validate `tools/task_surface_owner.json`, catalog owners, and verification contracts directly. Delete prose-extraction assertions. | `harness.command_surface`, `harness.test_catalog` | DONE |
| Harness schema-attachment validation | Parses owner-facade rows from the harness NLSpec for exact parity. | Validate `tools/harness_helper_ownership.json` and schema attachments against registered machine schemas. | `harness.generated_artifacts` | DONE |
| Harness JSON-shape validation | Parses harness requirement IDs and Network Flow spec/locators. | Restrict JSON-shape checks to registered machine owners and references. | `harness.generated_artifacts` | DONE |
| Frontend phase accounting | Reads guide paths, headings, digests, requirement IDs, acceptance IDs, and cumulative joins. | Replace executable document inputs with verification IDs now; replace and delete the remaining v1 frontend accounting structure in T-026. | `harness.test_catalog`, `harness.evidence_accounting` | DONE |
| Network Flow Go unit test | Reads the Network Flow NLSpec, phase map, and document text. | Assert behavior and machine verification/accounting contracts; remove file-content assertions. | `module.networkflow` | DONE |
| Network Flow activity accounting | Stores `source_spec` and Core/NLSpec locator fragments. | Retain only stable machine verification references and pure accounting fields. | `module.networkflow` | DONE |
| Reporting traceability test | Parses reporting NLSpec and Core documents. | Validate reporting corpus and owner verification contracts without opening documentation. | `module.reporting` | DONE |
| Report Composition traceability test | Parses report-composition NLSpec. | Validate a report-composition machine contract/corpus. | `module.reportcomposition` | DONE |
| OpenTelemetry conformance tool | Parses OTEL NLSpec and Core documents for status, TODOs, hooks, and error codes. | Move executable telemetry requirements to an authored machine contract under the telemetry contract owner. | `platform.telemetry` | DONE |
| SeaweedFS release-evidence tool | Reads Core documents for storage and threat-model references. | Validate an authored storage/release threat-policy contract. | `platform.objectstore`, `harness.release` | DONE |
| Design-token generator/checks | Parse `docs/design.md` and depend on documentation paths. | Move executable tokens and frontend design policy to machine owners; keep `docs/design.md` descriptive. | `package.ui`, `web.design` | DONE |
| Frontend import-boundary configuration/tests | Extract or validate design direction through documentation references. | Encode executable boundary rules in machine configuration; retain optional inert documentation references. | `web.architecture` | DONE |
| Visual browser spec | Asserts a visual registry `guide_path`. | Assert owner IDs, fixture IDs, scenario IDs, and contract version; remove guide-path behavior. | `harness.visual` | DONE |
| Visual fixture registry | Carries guide/document ownership. | Replace with stable fixture and owner IDs; preserve golden bytes during path-only moves. | `harness.visual` | DONE |
| Release-readiness evidence | Emits documentation owner references. | Emit stable machine owner/verification IDs; optional docs references remain display-only. | `harness.release` | DONE |
| `contracts/index.json` | Contains `owner_document` metadata. | Keep only as inert traceability or replace with stable owner IDs; no validator may open the document. | Contract registry | DONE |

### Decoupling acceptance rule

The final repository check MUST scan executable and validation sources under `cmd`, `internal`, `apps`, `packages`, `scripts`, `tools`, Makefiles, and package scripts for direct, runtime-computed, indirect-helper, realpath, and symlink-escape reads under `docs/` or `docs/spec/`. `tools/documentation_read_exceptions.json` MUST validate as `cartulary.documentation_read_exceptions.v1`; only documentation lint, link-check, or generation commands may receive an exact exception. Product, catalog, security, release, and conformance validators receive no exception. Documentation references remain inert strings.

`tools/delivery_phase_semantic_allowlist.json` MUST validate as `cartulary.delivery_phase_semantic_allowlist.v1`. The case-insensitive scan covers phase-shaped paths, symbols, titles, selectors, variables, schema IDs, artifact names, and target identities. A legitimate product phase or execution step requires normalized location, a stable semantic locator or bounded pattern, classification, owner, and reason. Line-number-only entries are invalid, and an ambiguous match blocks closure.

## 7. Public contracts, schemas, and artifacts

### 7.1 Successor command contracts

| Target | Command ID | Required inputs | Optional inputs | Behavior |
| --- | --- | --- | --- | --- |
| `test-slice` | `cartulary.harness.command.test_slice.v1` | `OWNER` | `ROWS`, worker inputs, `JSON` | Omitted `ROWS` selects every active executable row owned by `OWNER`; `default_check` never narrows the slice. |
| `service-backed-test-slice` | `cartulary.harness.command.service_backed_test_slice.v1` | `OWNER` | `ROWS`, worker inputs, `JSON` | Omitted `ROWS` selects every owned row whose runtime profile has `managed_services_required=true`; an explicit non-service row is invalid. |
| `explain-test-owner` | `cartulary.harness.command.explain_test_owner.v1` | `OWNER` | `JSON` | Report manifest, families, row counts, runner/evidence distribution, profiles, default topology participation, and exact narrow commands. |
| `task-guide` | `cartulary.harness.command.task_guide.v2` | `ROLE=module-author`, `OWNER` | `JSON` | Recommend owner-slice, generation, and broader evidence-derived commands from the same catalog snapshot. |
| `test-evidence-audit` | `cartulary.harness.command.test_evidence_audit.v1` | `OWNER`, `CHECK_RESULTS_DIR` | Explicit browser support, visual, accessibility, and measurement result roots | Reconcile owner rows with compatible semantic-identity evidence and emit an owner audit summary. |

`test-catalog-check` is private check-level work. It has no public command ID or target-local inputs and is selected by `make check`.

`OWNER` has no default, is trimmed, remains case-sensitive, and resolves to exactly one active owner. `ROWS=` and every zero-row selection are invalid. The `ROWS` parser retains raw comma tokens, trims them, rejects blank tokens and post-trim duplicates, proves every row belongs to `OWNER`, and then sorts accepted IDs. `VITEST_MAX_WORKERS` defaults to `4` and `PLAYWRIGHT_WORKERS` defaults to `3`; both accept only decimal integers `1..16`. Valid unused worker values are recorded in `unused_inputs`; invalid unused values still fail. `JSON` accepts only exact `1`; omitted or empty means human output, and target-local JSON combined with global machine-output mode is a usage error. Unknown owners, rows, removed phase inputs, and undeclared Make command-line inputs fail before child execution with `usage_error`; undeclared inherited environment variables retain the NLSpec ignore/strip rule.

### 7.2 New schema set

| Schema ID | Owns |
| --- | --- |
| `cartulary.verification_registry.v1` | Verification owner registry. |
| `cartulary.verification_contract.v1` | Per-owner executable verification entries. |
| `cartulary.test_owner_registry.v1` | Test owner registry. |
| `cartulary.test_family_manifest.v1` | Per-owner cross-runner test rows. |
| `cartulary.test_runner_registry.v1` | Closed runner and selector contracts. |
| `cartulary.test_slice_plan.v1` | Selected owner, rows, runtime profiles, work units, resources, and scheduler DAG. |
| `cartulary.test_slice_scheduler_summary.v1` | Slice execution terminal state and work-unit summaries. |
| `cartulary.test_evidence_accounting.v1` | Per-row expected and observed evidence. |
| `cartulary.test_evidence_audit_summary.v1` | Retained-run closure audit for one owner selection. |
| `cartulary.test_owner_summary.v1` | Aggregate owner result used by diagnostics/finalization. |
| `cartulary.test_owner_explanation.v1` | Read-only owner explanation JSON. |
| `cartulary.task_guide_summary.v2` | Read-only module-author guidance JSON. |
| `cartulary.test_catalog_check_summary.v1` | Private catalog validation summary. |
| `cartulary.documentation_read_exceptions.v1` | Closed documentation-tool read exceptions. |
| `cartulary.delivery_phase_semantic_allowlist.v1` | Classified legitimate product-phase and execution-step matches. |

All schemas are attached through the existing machine schema-attachment owner and validated by `make json-shape-check`. Old and new schema IDs are never accepted by the same production reader.

### 7.3 Artifact rules

- Every evidence artifact carries `schema_id`, `command_id`, `run_id`, `owner_id`, `selected_rows`, `source_snapshot_digest`, `catalog_semantic_digest`, `verification_semantic_digest`, `runtime_profile_digest`, `resource_profile_digest`, `fixture_profile_digest`, `started_at`, `finished_at`, and `duration_ms`.
- `source_snapshot_digest_v1` is SHA-256 over RFC 8785 canonical JSON containing every tracked and non-ignored untracked repository file sorted by repository-relative path with stable file-kind/mode identity and byte digest. Result roots, caches, and ignored files are excluded; symlinks are recorded as links and never followed.
- Artifact directories use run ID, target, owner ID, and semantic step/work-unit IDs; they never use a delivery phase as identity.
- A row result includes owner, family, verification IDs, runner, selector digest, runtime profile, terminal state, duration, and redacted artifact references.
- Evidence audits reject mismatched owner, selected-row inventory, source snapshot, catalog digest, verification digest, schema, or profile digests; missing conditional roots; duplicate candidates; mixed evidence; and ambiguous row results. Auditors inspect only caller-listed roots and never choose the newest artifact.
- `CHECK_RESULTS_DIR` is always required. Support, visual, accessibility, or measurement roots are required exactly when the selected owner rows contain the matching Playwright stage. Valid unnecessary roots are reported in `unused_inputs`.
- Semantic JSON digests use RFC 8785 and SHA-256 and reject duplicate members, non-I-JSON numbers, non-finite values, and negative zero. Timestamps are UTC RFC 3339 with uppercase `T`/`Z` and millisecond precision; durations are monotonic integer milliseconds. No wall-clock TTL applies, and release closure additionally requires a clean worktree.
- Duration baselines are rebuilt only from fresh successful owner-based warm runs. Phase-keyed baseline entries are deleted rather than translated heuristically.
- Generated topology and schedule artifacts remain reproducible outputs and are never hand-edited.
- Deleting Markdown ledgers requires deleting ledger paths/digests, generators, drift checks, finalizer steps, topology nodes, generated-artifact policy entries, and Markdown-lint exclusions.

Slice implementation MUST follow this order: validate all inputs; resolve `OWNER`; parse explicit `ROWS`; select all owner rows or all service-backed owner rows when omitted; reject an empty set; resolve selectors and profiles; reject zero, multiple, or overlapping resolution; sort by `row_id`; emit an immutable plan; execute with `stop_on_first_failure=false`; drain running work; do not start dependency-blocked work; always run finalizers; and emit exactly one terminal record per selected row. Owner success requires every row to be `passed` or validly `skipped_authorized`. No automatic product-test retry is permitted; existing bounded service-readiness retries remain unchanged, all attempts are retained, and only the final row state participates in aggregation.

Failure mapping is exact: invalid inputs or selection are `usage_error/config/2`; invalid catalog, schema, profile, selector, or duplicate ownership are `configuration_error/config/2`; product assertions retain product exit `10`; missing or inconsistent row accounting is `scheduler_accounting_error/harness/11`; missing, stale, mixed, or incompatible evidence is `artifact_error/artifact/11`; unauthorized documentation access is `boundary_policy_violation/harness/11`; cleanup with no earlier failure retains cleanup exit `12`; timeout retains `13`; supported signals retain the existing `130`, `143`, or signal-specific mapping. Configuration failures occur before setup. Existing primary-failure precedence remains, and cleanup never overwrites an earlier failure.

Slice targets retain `<target>/tool-run-summary.json`, `test-slice-plan.json`, `test-slice-scheduler-summary.json`, `test-evidence-accounting.json`, and `test-owner-summary.json`. The evidence audit retains its tool summary and `test-evidence-audit-summary.json`. Explanation and task-guide commands are read-only. With `JSON=1`, each read-only command emits exactly one command-schema object plus LF. For slice commands, `JSON=1` changes stdout only, execution still occurs, and stdout is the final `cartulary.test_slice_scheduler_summary.v1` object plus LF.

## 8. Test-source and fixture migration rules

### 8.1 Rename scope

Rename all delivery-phase-shaped test surfaces to behavior/owner names:

- Go test filenames and function/subtest names.
- Vitest and Playwright filenames, suite titles, test titles, tags, and helper symbols.
- Shell test names, fixture directories, scenario IDs, row IDs, visual snapshot paths, accessibility scenario IDs, and measurement scenario IDs.
- Phase-specific helper names when the helper is actually semantic test support.
- Generated topology labels and artifact paths that use delivery phase identity.

Do not mechanically replace legitimate product-phase terminology. Every candidate is classified as `delivery_phase`, `product_phase`, `execution_step`, or `ambiguous`; ambiguous cases require owner review.

### 8.2 Behavioral preservation and deletion

- Preserve supported observable behavior, route shapes, authorization outcomes, fixture isolation, service ownership, cleanup, redaction, and evidence classifications during renames.
- Delete tests whose sole purpose is documentation text parity, phase completeness, obsolete compatibility, unsupported feature behavior, or a superseded implementation detail without a current production requirement.
- Consolidate duplicates only when the surviving row preserves all distinct supported assertions and evidence classes; record both old rows in the crosswalk.
- Do not promote informative measurements, visual direction, or draft extension behavior into conformance evidence.
- Do not retain a test merely to keep row counts stable. Reconciliation correctness matters; numerical preservation does not.

### 8.3 Visual and fixture safety

Path-only golden changes are byte-preserving moves. Before and after each visual slice, record sorted SHA-256 digests keyed by scenario ID and prove unchanged bytes. Do not run a visual-update target to accomplish a rename. Any pixel change is a separate visual-maintenance task governed by the visual golden guide and appropriate review.

Fixture identity moves from phase labels to semantic owner/scenario IDs. Cleanup ownership, isolation scope, runtime profile, and resource locks must remain explicit in the manifest.

## 9. Migration crosswalk and reconciliation contract

WS-00 creates a temporary authored migration file outside generated roots. It is retained through cutover, used by validation, and deleted only after the final reconciliation report is captured in the handoff evidence.

Create temporary Draft 2020-12 schemas `cartulary.test_migration_baseline.v1` and `cartulary.test_migration_crosswalk.v1`, with `additionalProperties: false`, closed enums, sorted duplicate-free arrays, and constant `schema_id`. The immutable baseline key is `(source_registry_id, legacy_row_id)` and records the exact legacy selector and its digest.

The crosswalk is a closed tagged union:

| Disposition | Required content | Forbidden content |
| --- | --- | --- |
| `migrated` | Baseline key, target `owner_id`, `row_id`, nonempty `verification_ids`, governing owner/requirement, applied adjudication rule, collaborators, review revision, and owner-review evidence. | Deletion fields. |
| `consolidated` | Every `migrated` field plus the surviving row and assertion-preservation evidence proving all distinct supported assertions and evidence classes remain represented. | Deletion fields. |
| `deleted` | Baseline key, one closed reason from `unsupported`, `duplicate`, `obsolete`, `documentation_only`, or `non_production`, governing owner/requirement, review revision, and owner-review evidence. | Target owner, row, verification, and selector fields. |

New v2 rows are not legacy dispositions. A separate `new_rows` collection requires target owner, row, verification IDs, and either an authorization ID or an adopted owner-requirement ID.

Reconciliation fails unless:

1. All 548 old identities appear exactly once.
2. Every new row appears in `new_rows` with provenance or authorization.
3. No old selector remains live after cutover.
4. Consolidation is many-to-one only and has an assertion-preservation explanation.
5. Deletions cite unsupported, duplicate, obsolete, or non-production grounds and the governing owner review.
6. Totals by source registry, disposition, and target owner are exact and emitted in the final handoff.
7. The five Graph Projection identities are present unless the frozen baseline proves exact identity overlap and revises the population before migration work begins.

The temporary crosswalk is not a runtime input, alias map, old-schema reader, or permanent historical compatibility artifact.

## 10. Production-source follow-up ledger

WS-00 records delivery-phase-shaped authored production paths and symbols without renaming them by default. At minimum, inspect known SQL/query and module filenames under incidents, timeline, recovery, and reporting, plus generated SQL descendants.

| Finding | Disposition |
| --- | --- |
| Test selector depends on a production phase-shaped filename but can target a semantic package/symbol | Decouple the selector; record a module-owner follow-up. |
| Production rename is required to make the test or harness owner boundary coherent | Mark `BLOCKED: production-owner authorization`; execute only after explicit scope expansion. |
| Generated code contains phase-shaped descendants | Update the authored owner input only in a separately authorized module change, then regenerate through Make. |
| Name represents a legitimate product/runtime phase | Keep; add to semantic scan allow rules with its owner and reason. |
| Name is historical delivery metadata with no runtime meaning | Add a module-owner follow-up with path, proposed semantic name, risk, and validation target. |

These follow-ups do not block harness completion unless the old production identity remains part of a live test row, public harness interface, selector, artifact identity, or generated topology.

WS-06 closure is recorded in `tools/delivery_identity_followup_ledger.json`. It gives all 13 frozen follow-ups, three additional live helper paths, two symbols, and four retained product protocol identifiers explicit owner dispositions. The executable closure check requires every renamed path to be absent, every semantic replacement to exist, every retained protocol identifier to resolve exactly once without entering catalog selectors, and the repository path scan to contain no unclassified delivery-shaped identity.

## 11. Dependency-ordered workstreams

### 11.1 Workstream summary

| ID | Workstream | Status | Depends on | Primary evidence | Rollback boundary |
| --- | --- | --- | --- | --- | --- |
| WS-00 | Freeze baseline and build migration crosswalk | DONE | None | `tools/test_migration_baseline.json`; `tools/test_migration_crosswalk.json` | Revert inventory-only commit. |
| WS-01 | Revise owner contracts and repository procedure | DONE | WS-00 | Adopted harness contract and command-policy review | Revert contract commit before implementation consumers land. |
| WS-02 | Create machine verification owners and remove documentation parsing | DONE | WS-01 | Schema-valid contracts and decoupling ledger closure | Revert an owner contract and its consumers together. |
| WS-03 | Build unified test catalog | DONE | WS-01, WS-02 | Registry/manifests loader and catalog checks | Revert catalog stack; old catalog remains authoritative only before cutover. |
| WS-04 | Migrate backend rows | DONE | WS-03 | 456 backend dispositions and focused tests | Revert complete owner slices, never individual aliases. |
| WS-05 | Migrate frontend rows | DONE | WS-03 | 87 frontend dispositions and browser accounting tests | Revert complete owner slices with fixture metadata. |
| WS-06 | Rename tests, symbols, fixtures, and goldens | DONE | WS-04, WS-05 | Semantic scan, visual digest report, and production follow-up closure ledger | Revert complete rename slices; no duplicate old/new tests. |
| WS-07 | Replace slice, audit, schema, and artifact APIs | IN_PROGRESS | WS-03–WS-06 | Successor CLI contract/smoke tests | Revert the whole interface checkpoint before atomic cutover. |
| WS-08 | Migrate browser stages and scheduler topology | TODO | WS-05, WS-07 | Owner-based DAG, lifecycle, and browser tests | Revert authored topology and generated outputs together. |
| WS-09 | Update task surface, generation, finalization, and baselines | TODO | WS-07, WS-08 | Generated surface/drift and fresh baseline plan | Revert owner inputs plus regenerated outputs together. |
| WS-10 | Atomic deletion and hard cutover | TODO | WS-02–WS-09 | Deletion manifest and zero-reference scans | Revert the entire cutover commit; never add shims. |
| WS-11 | Focused and broad verification | TODO | WS-10 | Successful fresh run roots and audit summaries | Forward-fix or revert the full cutover; old evidence is invalid. |
| WS-12 | Finalize handoff and remove temporary migration inputs | TODO | WS-11 | Final reconciliation, retained-run evidence, handoff log | Reopen tracker if any closure invariant fails. |

Before WS-10, old and new implementations may coexist only on unmerged migration branches to enable comparison. Public authority remains singular, and no dual-reader or dual-writer state may be merged. WS-10 is one cohesive checkpoint. After WS-10, remediation is a forward fix or full checkpoint revert, never a compatibility shim.

At most one workstream may be `IN_PROGRESS`. WS-03 MUST NOT start until WS-01 and WS-02 are both `DONE`. Baseline rebasing resets WS-00 to `TODO` and invalidates every downstream reconciliation artifact whose recorded baseline digest changes. The final state permits no `TODO`, `IN_PROGRESS`, or `BLOCKED`. `DEFERRED` cannot cover any Section 15 completion obligation. `DROPPED` requires the closed disposition reason and owner approval defined in Section 9.

### 11.2 WS-00 — Baseline and crosswalk

Prerequisites: repository at or intentionally rebased from the recorded baseline; clean understanding of any intervening changes.

Deliverables:

- Reproduce registry, row, ledger, path, schema, artifact, target, helper, topology, duration-baseline, and direct-document-read inventories.
- Freeze every old row and selector in the temporary crosswalk.
- Inventory phase-shaped test paths, symbols, titles, fixtures, goldens, generated labels, and production follow-ups.
- Capture existing visual golden digests and current relevant focused-test state.
- Record generated versus authored paths before any move.

Exit: all 548 rows are represented; discovery commands and outputs are retained; unknown counts are resolved; current failures are distinguished from migration failures.

If exact identity overlap changes the provisional population, WS-00 MUST update the baseline schema artifact, crosswalk cardinality, task totals, acceptance fixtures, risk table, completion criteria, and every prose total in this tracker before WS-01 begins.

### 11.3 WS-01 — Owner contracts and procedures

Prerequisites: WS-00 complete.

Deliverables:

- Revise `docs/testing-harness-nlspec.md` to make owner catalogs, owner slices, machine verification contracts, and on-demand diagnostics the adopted mechanics.
- Attach every new requirement to at least one acceptance criterion; reserve retired requirement IDs rather than assigning them new semantics.
- Remove normative mandates for phase maps, registries, slices, schedules, ledgers, frontend joins, guide digests, and documentation parsing.
- Update `AGENTS.md`, active guides, and the modular-refactor framework where command guidance or generic phase-preservation language would misdirect future work.
- Preserve Core/adopted subsystem authority and domain terminology; do not turn the tracker into product authority.

Exit: active documentation describes one owner-based public surface and contains no instruction to validate tests by reading documentation.

### 11.4 WS-02 — Machine verification and documentation decoupling

Prerequisites: WS-01 adopted contract.

Deliverables:

- Add verification registry/contracts and registered schemas.
- Migrate harness, Network Flow, reporting, report composition, telemetry, object-store/release, design, visual, import-boundary, and release-readiness executable requirements.
- Delete prose checksums, heading/digest comparisons, TODO scans, and source-locator parsing.
- Add a focused no-document-read policy check with a narrow documentation-linter exception model.

Exit: every item in Section 6 is `DONE` or explicitly `DROPPED`; executable validation is schema-backed and documentation-independent.

### 11.5 WS-03 — Unified catalog

Prerequisites: WS-01 and WS-02 complete enough to supply active verification IDs.

Deliverables:

- Add owner registry, family manifests, schemas, loader, selector resolvers, and catalog check.
- Register runner-specific selector validation and runtime/resource profiles.
- Migrate Graph Projection from the subsystem registry.
- Replace `unowned_regression` with explicit owners or delete unsupported rows.
- Add import-direction guardrails between catalog, execution, accounting, scheduler, and diagnostics.

Exit: the unified loader can represent every retained baseline row across all runners and rejects all invalid ownership/reference cases.

### 11.6 WS-04 — Backend row migration

Prerequisites: WS-03 complete.

Deliverables:

- Adjudicate all 456 backend phase-map rows owner by owner.
- Preserve test layer, runtime needs, security posture, and exact supported assertions.
- Remove duplicate, obsolete, draft-only, documentation-only, and unsupported rows with recorded reasons.
- Update the crosswalk and add focused catalog/selector tests after every owner slice.

Exit: all 456 backend rows have terminal dispositions and every retained selector resolves through the unified catalog.

### 11.7 WS-05 — Frontend row migration

Prerequisites: WS-03 complete.

Deliverables:

- Adjudicate all 87 frontend rows into the same owner manifests as backend rows.
- Replace guide headings/digests, requirements/acceptance IDs, cumulative dependencies, and base-phase joins with verification IDs and resource dependencies.
- Preserve functional, webserver-backed, stateful, accessibility, visual, and measurement evidence distinctions.
- Replace frontend evidence accounting and audit logic with generic owner accounting.

Exit: all 87 frontend rows have terminal dispositions; no frontend ownership namespace or documentation-derived activation state remains.

### 11.8 WS-06 — Semantic source and fixture names

Prerequisites: WS-04 and WS-05 row ownership stable.

Deliverables:

- Rename test files, functions, suites, titles, helpers, fixtures, scenarios, row IDs, and golden paths to semantic names.
- Update selectors and accounting atomically with each rename.
- Prove byte identity for path-only golden moves.
- Populate the production follow-up ledger and semantic allow rules.

Exit: the delivery-phase scan is empty in harness/test identity surfaces; all remaining `phase` matches are classified product/runtime concepts or execution steps.

### 11.9 WS-07 — Owner slice and evidence APIs

Prerequisites: WS-03 through WS-06 complete.

Deliverables:

- Implement the successor Make/CLI contracts in Section 7.
- Replace slice plan, scheduler summary, accounting, audit, owner summary, and step artifact schemas and filenames.
- Update target-plan output to use owner/family/work-unit identities.
- Add usage, JSON output, invalid-input, selection, service-backed, cancellation, cleanup, and retained-evidence smoke tests.

Exit: all successor targets operate end to end from the unified catalog, and no successor reader accepts a phase schema.

### 11.10 WS-08 — Browser and scheduler topology

Prerequisites: WS-05 and WS-07 complete.

Deliverables:

- Derive browser stages and batches from runner/resource/runtime profiles rather than phase ordering.
- Preserve external/shared server behavior, worker controls, fixture lifecycle, cleanup, artifact capture, redaction, fail-fast, and failure-priority semantics.
- Remove frontend row-accounting environment variables and phase scheduler DAG helpers.
- Update authored topology and regenerate downstream artifacts through Make.

Exit: default and owner-slice browser scheduling is phase-neutral and existing lifecycle characterization remains green.

### 11.11 WS-09 — Task surface, generation, finalization, and baselines

Prerequisites: WS-07 and WS-08 complete.

Deliverables:

- Replace public target/input/role definitions in the task-surface owner and generated Make surface.
- Fold schedule generation into `generate` and drift validation into `generate-drift`.
- Remove ledger/schedule steps from finalization and replace phase summary consumption with owner/step summaries.
- Update helper ownership, schema attachments, topology manifests, smoke suites, help text, task guidance, and repository command docs.
- Delete phase-keyed duration baselines and define the fresh warm-run refresh path.

Exit: generated surfaces are owner-first, drift-clean, and contain no removed command or input.

### 11.12 WS-10 — Atomic hard cutover

Prerequisites: WS-02 through WS-09 complete on the migration branch.

Deliverables:

- Delete both phase registries, all phase maps, all frontend phase maps, all 26 ledgers, the policy-exception file, subsystem registry/maps, and superseded classification entries.
- Delete the `phase-accounting` implementation and all obsolete phase diagnostic, slice, scheduler, generator, finalizer, schema, smoke, and compatibility code.
- Publish the v2 NLSpec frontmatter, new schemas, task-surface bindings, catalog, topology profiles, generated outputs, and v1 deletion manifest in the same cutover checkpoint.
- Remove generated-artifact policy and Markdown-lint entries that existed only for ledgers.
- Regenerate permitted outputs through public Make targets.
- Run zero-reference scans for every removed path, schema, target, variable, and artifact identity.

Exit: deletion manifest is complete; the repository has one owner-based authority path; no compatibility layer or checked-in ledger remains.

### 11.13 WS-11 — Verification

Prerequisites: WS-10 complete.

Deliverables:

- Run schema, catalog, boundary, docs-decoupling, semantic-name, focused owner, frontend, browser, accessibility, visual, security, generation, and release checks.
- Run `make agent-finalize` before broad end-of-run verification.
- Produce a successful fresh warm `make check` run, finalize with that successful run root, and repeat broad verification after retained baseline refresh.
- Run `make release-check` because public harness and release-evidence behavior changed.

Exit: all required commands pass from a clean tree; failures include target, summary artifact/run root, and relation to the migration; no old retained run is used as closure evidence.

### 11.14 WS-12 — Final handoff

Prerequisites: WS-11 complete.

Deliverables:

- Capture final crosswalk totals by disposition and owner.
- Remove the temporary crosswalk after retaining its final reconciliation report.
- Close every tracker row, documentation-decoupling item, production follow-up, and risk.
- Record final branch/commit/tree state, changed owner inputs, generated outputs, commands, run roots, and skipped checks.

Exit: every binary completion criterion in Section 15 is true and another engineer can verify closure without reconstructing migration history.

## 12. Top-level implementation tracker

| ID | Work item | Workstream | Status | Depends on | Evidence | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Reproduce 13/456 backend baseline | WS-00 | DONE | None | Migration baseline | Registry and exact row identities frozen. |
| T-002 | Reproduce 13/87 frontend baseline | WS-00 | DONE | None | Migration baseline | Registry and exact row identities frozen. |
| T-002A | Reproduce five-row Graph Projection baseline | WS-00 | DONE | None | Migration baseline | All five subsystem identities and selectors are frozen. |
| T-003 | Inventory 26 ledgers and all consumers | WS-00 | DONE | None | Migration baseline deletion inventory | Every ledger path is frozen; consumer discovery remains a WS-10 deletion check. |
| T-004 | Create 548-row temporary crosswalk | WS-00 | DONE | T-001,T-002,T-002A | Baseline and crosswalk validation | Every frozen identity appears once as a pending baseline key; terminal dispositions are populated by WS-04/WS-05. |
| T-005 | Inventory phase-shaped test identities and goldens | WS-00 | DONE | None | Migration baseline semantic inventory and 56 golden digests | Exact rename population and initial visual bytes frozen. |
| T-006 | Inventory direct/indirect documentation reads | WS-00 | DONE | None | 68-candidate decoupling inventory | Known executable/document coupling candidates are frozen for WS-02 adjudication. |
| T-007 | Inventory production-source follow-ups | WS-00 | DONE | None | 13 classified follow-ups | Seven authored identities and six generated descendants have owners and dispositions. |
| T-008 | Adopt owner-based harness NLSpec | WS-01 | DONE | T-004,T-006 | 216-requirement/72-acceptance editorial audit | Phase mechanics are retirement history only; owner mechanics are normative. |
| T-009 | Update active command/procedure docs | WS-01 | DONE | T-008 | Markdown lint and active-doc command scan | No active procedure directs contributors to a phase command. |
| T-010 | Register verification schemas | WS-02 | DONE | T-008 | JSON shape evidence | Registry and owner contracts validate. |
| T-011 | Migrate harness executable requirements | WS-02 | DONE | T-010 | Focused harness tests | Harness tests consume machine owners only. |
| T-012 | Migrate product/tool documentation readers | WS-02 | DONE | T-010 | Per-owner tests and scan | Section 6 is closed. |
| T-013 | Add no-document-read guard | WS-02 | DONE | T-011,T-012 | Negative fixtures | Unauthorized doc reads fail deterministically. |
| T-014 | Add owner registry and family schemas | WS-03 | DONE | T-010 | JSON shape evidence | Owner/family files validate. |
| T-015 | Implement unified catalog loader and validator | WS-03 | DONE | T-014 | Unit/smoke tests | Invalid references/selectors are rejected. |
| T-016 | Absorb Graph Projection subsystem map | WS-03 | DONE | T-015 | Crosswalk/catalog evidence | No special subsystem path needed. |
| T-017 | Eliminate `unowned_regression` | WS-03 | DONE | T-015 | Classification reconciliation | Every retained support row has an owner. |
| T-018 | Add module-family import guardrails | WS-03 | DONE | T-015 | Boundary tests | Dependency direction is enforced. |
| T-019 | Migrate phase0–phase4 backend rows | WS-04 | DONE | T-015 | 163 dispositions, 24 support rows, catalog/migration checks, and focused tests | Rows have terminal dispositions. |
| T-020 | Migrate phase5–phase8 backend rows | WS-04 | DONE | T-015 | 86 dispositions, 6 support rows, catalog/migration checks, and focused tests | Rows have terminal dispositions. |
| T-021 | Migrate phase9–phase12 backend rows | WS-04 | DONE | T-015 | 207 dispositions, 7 support rows, catalog/migration checks, and focused tests | Rows have terminal dispositions. |
| T-022 | Reconcile all 456 backend rows | WS-04 | DONE | T-019,T-020,T-021 | 456-row/550-selector reconciliation report and 37-row/118-selector support report | Count and selector coverage close. |
| T-023 | Migrate FE-P0–FE-P4 rows | WS-05 | DONE | T-015 | 35 dispositions, exact selector reconciliation, catalog and frontend tests | Rows have terminal dispositions. |
| T-024 | Migrate FE-P5–FE-P8 rows | WS-05 | DONE | T-015 | 22 dispositions, 47 exact title selectors, catalog and frontend tests | Rows have terminal dispositions. |
| T-025 | Migrate FE-P9–FE-P12 rows | WS-05 | DONE | T-015 | 30 dispositions, 39 exact title atoms, catalog and frontend tests | Rows have terminal dispositions. |
| T-026 | Remove frontend guide/cumulative accounting | WS-05 | DONE | T-023,T-024,T-025 | Catalog-native owner accounting projection and negative tests | One successor owner accounting model remains. |
| T-027 | Reconcile all 87 frontend rows | WS-05 | DONE | T-026 | 87-row/216-atom reconciliation report | Count and selector coverage close. |
| T-028 | Rename Go test identities | WS-06 | DONE | T-022 | 102 path, 632 test-entry, and 333 helper-declaration renames; catalog reconciliation; semantic scan; 1,295-test fast suite | No delivery-phase Go test identity remains. |
| T-029 | Rename Vitest/Playwright identities | WS-06 | DONE | T-022,T-027 | 53 path renames; semantic source/catalog scan; frontend and harness gates | No delivery-phase frontend test identity remains. |
| T-030 | Rename fixtures/scenarios/goldens | WS-06 | DONE | T-027 | 56 visual and 15 diagnostic golden moves, digest report, and 29-scenario visual run | Paths are semantic and bytes preserved. |
| T-031 | Close production follow-up classifications | WS-06 | DONE | T-007,T-028,T-029 | 16-path/2-symbol/4-protocol follow-up ledger and zero-match scan | Every remaining match has owner/disposition. |
| T-032 | Implement owner-slice planning/execution | WS-07 | DONE | T-015,T-022,T-027 | Deterministic plan contracts, exact Go/Vitest execution, managed-service smoke, and fail-closed selection fixtures | Both successor slice targets work. |
| T-033 | Implement owner diagnostics/task guide | WS-07 | IN_PROGRESS | T-032 | Text/JSON contract tests | Commands report exact owner topology. |
| T-034 | Implement owner evidence accounting/audit | WS-07 | TODO | T-032 | Audit fixtures | Fresh compatible evidence reconciles. |
| T-035 | Replace schemas and artifact identities | WS-07 | TODO | T-032,T-034 | Schema attachment/drift checks | No successor artifact uses delivery phase identity. |
| T-036 | Migrate browser stages/runtime profiles | WS-08 | TODO | T-027,T-032 | Browser plan tests | Selection derives from catalog resources. |
| T-037 | Migrate scheduler DAG and lifecycle | WS-08 | TODO | T-036 | Scheduler/lifecycle tests | Phase ordering has no execution role. |
| T-038 | Regenerate owner-first topology | WS-08 | TODO | T-037 | Generation/drift evidence | Authored and generated topology agree. |
| T-039 | Cut task surface to successor targets | WS-09 | TODO | T-032,T-033,T-034 | Help/contract/smoke tests | Removed inputs fail and new targets work. |
| T-040 | Fold schedule generation into standard targets | WS-09 | TODO | T-038 | Generate/generate-drift | No public schedule target remains. |
| T-041 | Replace finalizer and duration baseline flow | WS-09 | TODO | T-034,T-038 | Finalizer tests and refresh plan | Owner summaries drive finalization. |
| T-042 | Delete ledgers and ledger machinery | WS-10 | TODO | T-039,T-041 | Deletion manifest/zero scan | All 26 and every consumer are gone. |
| T-043 | Delete phase/subsystem registries and maps | WS-10 | TODO | T-016,T-022,T-027,T-039 | Deletion manifest/zero scan | Unified catalog is sole owner. |
| T-044 | Delete phase-accounting and compatibility code | WS-10 | TODO | T-035,T-037,T-043 | Boundary and zero-reference scans | No old reader or shim remains. |
| T-045 | Regenerate all permitted outputs | WS-10 | TODO | T-042,T-043,T-044 | Generated drift checks | Clean owner-first generated tree. |
| T-045A | Prove atomic v2 parity and retirement | WS-10 | TODO | T-045 | NLSpec/schema/task-surface/topology parity and zero-reference report | v2 is complete and no v1 alias, reader, writer, catalog, or artifact identity is active. |
| T-046 | Run focused verification matrix | WS-11 | TODO | T-045A | Command results/run roots | All focused gates pass. |
| T-047 | Run agent finalization and first warm check | WS-11 | TODO | T-046 | Successful warm run root | Fresh broad evidence exists. |
| T-048 | Refresh retained baselines and repeat broad checks | WS-11 | TODO | T-047 | Finalized root and second results | No phase baseline is reused. |
| T-049 | Run release check | WS-11 | TODO | T-048 | Release-check result | Public/release harness changes pass. |
| T-050 | Capture reconciliation and remove crosswalk | WS-12 | TODO | T-049 | Final reconciliation report | Temporary compatibility-free migration input removed. |
| T-051 | Complete handoff and closure audit | WS-12 | TODO | T-050 | Handoff log and clean status | Section 15 is fully satisfied. |

## 13. Verification matrix

Choose the narrowest public Make target that covers each slice, then broaden only after focused evidence is green. Direct Go, pnpm, Vitest, Playwright, or raw-script invocations are developer diagnostics, not closure evidence unless a Make-owned wrapper invokes them.

| Gate | Required coverage | Minimum evidence |
| --- | --- | --- |
| Machine shape | Verification registry/contracts, owner registry/manifests, schema attachments, artifact schemas | `make json-shape-check`, `make generated-artifact-policy-check`, focused schema negative fixtures |
| Generation | Contract/code generation, topology, task surface, schedules | `make generate-drift`, relevant generator-focused tests; generation performed through `make generate` only |
| Catalog integrity | Unique rows, one owner, valid collaborators/verifications/selectors, active profiles, no unowned bucket | `test-catalog-check` and negative fixtures |
| Reconciliation | All 548 old rows and every new row accounted for | Crosswalk validator and final totals report |
| Documentation decoupling | No executable validator reads docs/specs | No-document-read guard plus repository scan |
| Semantic naming | No delivery-phase identity in harness/test paths, symbols, titles, fixtures, artifacts, variables, or topology | Semantic scan with reviewed product/runtime allow rules |
| Boundaries | Catalog/execution/accounting/scheduler/diagnostics direction | Harness boundary tests and frontend import-boundary check |
| Backend behavior | Every migrated backend owner | Focused owner slices plus the generated evidence-class gates below |
| Frontend behavior | Vitest/type/import boundaries | `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make lint-biome` |
| Browser behavior | Webserver-backed/stateful/runtime profile and lifecycle | Owner slices plus `make browser-e2e-webserver-backed` and `make browser-e2e-stateful` |
| Accessibility | Owner scenarios and retained audit | `make browser-e2e-a11y` and owner audit summary |
| Visual | Scenario ownership, registry, byte-preserving moves, strict comparison | Digest report and `make browser-e2e-visual`; no update target for renames |
| Measurement | Informative versus claim-bearing posture and fixture identity | `make browser-e2e-measurement`; Core 05 gates only where a claim profile is active |
| Security/redaction | Auth, test routes, artifacts, logs, object handles, retained evidence | Relevant targeted security tests, `make go-gosec-targeted`, and release evidence checks |
| Harness scripts/docs | Shell/script/Markdown correctness and public help | `make lint-scripts`, `make lint-shell`, `make lint-markdown`, `make help`, `make help-all` |
| Finalization | Fresh retained owner evidence and clean generated state | `make agent-finalize`, successful warm `make check`, finalization with its results root, second broad check |
| Release | Public command, artifact, and release-evidence compatibility | `make release-check` |

The verification registry MUST generate this applicability mapping; human reviewers cannot waive a required row:

| Evidence class | Required gate |
| --- | --- |
| `unit` | Owner slice plus unit/fast gate |
| `integration` | Owner slice plus integration gate |
| `browser` | Runtime-profile-selected webserver-backed, stateful, or support gate |
| `accessibility` | `browser-e2e-a11y` |
| `visual` | Golden digest validation and `browser-e2e-visual` |
| `measurement` | `browser-e2e-measurement`; informative unless claim-authorized |
| `static` | Exact static target named by the verification contract |
| `security` | Exact security target named by the verification contract |
| `release` | `release-check` |

A gate is `not_applicable_zero_rows` only when the owner has no rows of that evidence class. A required row cannot be manually declared inapplicable.

Required negative fixtures cover missing, blank, duplicate, unknown, cross-owner, and zero-row selections; worker bounds `0`, `1`, `16`, and `17`; `JSON` omitted, empty, `1`, and invalid values; retired targets and `PHASE` inputs; missing, extra, null, duplicate, unordered, unknown, and overlength schema fields; selector zero/multiple/overlap, traversal, symlink, glob, and regular-expression cases; every terminal state and simultaneous primary/cleanup failures; every semantic identity digest mismatch; missing, duplicate, unnecessary, and mixed evidence roots; direct, indirect, and symlinked documentation reads; legitimate and ambiguous semantic allowlist matches; and all 548 baseline identities, new-row authorization, consolidation proof, deletion proof, and Graph Projection absorption.

If a target fails, record the target, exit status, relevant summary artifact or run root, whether the failure reproduces on the baseline, whether it is related to the migration, and the chosen remediation. Do not mark a work item `DONE` on a rerun that silently excludes its rows.

## 14. Risks, blockers, and controls

| Risk | Control | Blocking condition |
| --- | --- | --- |
| Silent row loss or duplicate execution across 548 rows | Validated one-time crosswalk, reverse coverage from new rows, owner totals, selector-resolution tests | Any old row lacks one terminal disposition. |
| Owner catalog becomes another shallow catch-all | Qualified owners, semantic families, one primary owner, import guardrails | A row cannot be assigned without convenience ownership. |
| Public harness churn strands local/CI workflows | Atomic task-surface generation, contract tests, active docs update, hard rejection of old inputs | Successor surface is not end-to-end before cutover. |
| Removing documentation parsers permits unnoticed normative drift | Reviewed derived contracts, stable verification IDs, owner review, contract diffs | Normative owner is contradictory or not adopted. |
| Visual path renames alter golden bytes | Before/after digest maps, move-only operation, strict visual run | Any unexpected digest change. |
| Product/runtime phase terminology is falsely removed | Semantic classification and reviewed allow rules | Match meaning is ambiguous. |
| Draft extension behavior is accidentally promoted | Check owner status before creating verification IDs | Only a draft source supports the behavior. |
| Old retained evidence appears green under new schemas | Hard schema rejection and fresh warm baseline | Evidence does not carry a compatible owner catalog snapshot. |
| Scheduler behavior drifts while identity changes | Characterize cancellation, cleanup, resources, failure priority, external server modes | Lifecycle characterization is incomplete. |
| Generated outputs are hand-edited during broad cutover | Generated policy and Make-only regeneration | Required change has no authored owner input. |
| Prior trackers conflict with this transformation | Record their decisions as historical and apply Section 2.2 precedence | Conflict would change product behavior rather than harness/test organization. |
| Production phase names leak into test selectors | Decouple selectors or obtain module-owner scope expansion | Live owner catalog still requires a delivery-phase production identity. |

## 15. Binary completion criteria

The migration is complete only when every statement below is true:

- [ ] All 548 baseline rows have exactly one reviewed terminal disposition and every retained new row has provenance.
- [ ] One unified owner catalog is the only executable test-ownership source.
- [ ] Every retained row has one primary owner, valid verification IDs, a resolvable selector, explicit evidence class, and runtime/resource metadata.
- [ ] No `unowned_regression`, subsystem special-case registry, frontend ownership namespace, or phase exception mechanism remains.
- [ ] No backend/frontend phase registry or phase map remains active or referenced.
- [ ] All 26 Markdown phase ledgers and every generator, digest, drift check, ignore, finalizer step, and topology consumer are deleted.
- [ ] No public or private phase-slice, phase-ledger, phase-schedule, explain-phase, or frontend-phase-audit interface remains.
- [ ] `PHASE`, `PHASE_NAMESPACE`, base-phase joins, and phase-author roles are absent from the harness command contract.
- [ ] No old phase schema reader, artifact fallback, alias target, dual catalog, or compatibility shim remains.
- [ ] No product, harness, release, visual, security, or conformance validator reads `docs/` or `docs/spec/`.
- [ ] Executable requirements are schema-valid machine contracts outside documentation directories and reference current adopted owners.
- [ ] Unsupported, draft-only, documentation-only, obsolete, and duplicate rows have been removed with recorded reasons.
- [ ] Delivery-phase identity is absent from harness/test paths, symbols, titles, fixtures, scenario IDs, artifact identities, and generated topology.
- [ ] Remaining uses of “phase” are reviewed product/runtime concepts or semantic execution steps.
- [ ] Path-only visual moves preserve bytes, and all visual/accessibility evidence resolves through owner IDs.
- [ ] Generated artifacts are owner-first, reproducible through Make, and drift-clean.
- [ ] Successor commands, task guidance, diagnostics, owner slices, and evidence audits pass their contract tests.
- [ ] Every generated evidence-class gate is either passing or exactly `not_applicable_zero_rows`; no required row is manually skipped.
- [ ] A fresh successful warm check supplies retained baselines; no historical phase run closes evidence.
- [ ] `make agent-finalize`, the repeated broad verification, and `make release-check` pass from a clean tree.
- [ ] The temporary migration crosswalk has been removed after its final reconciliation report is retained.
- [ ] Production-source follow-ups have explicit owners and do not leak delivery-phase identity into the completed harness.

## 16. Handoff protocol and log

At every handoff, update the top-level tracker and append one entry below. Do not rewrite prior truthful entries.

Each entry must include:

- Date, branch, commit, and dirty-tree state.
- Workstream/task IDs started and completed.
- Authored and generated files changed, clearly distinguished.
- Substantive decisions and owner-contract interpretations.
- Crosswalk totals before and after the slice.
- Commands run, results, run roots, and relevant summary artifacts.
- Failures, relation to the migration, and remaining blockers.
- Skipped checks with reasons.
- Exact next safe task and prerequisites.
- Rollback boundary for the completed slice.

### Handoff log

#### 2026-07-17 — Tracker creation

- Branch/commit: `revision/grid-adapter` at `625e5dea34d7ad743075ad766f00928924ee96b1`.
- Initial tree: clean; this tracker was absent.
- Work completed: planning baseline, target architecture, decision record, migration map, documentation-decoupling ledger, workstream dependency graph, implementation tracker, validation matrix, and binary closure gates.
- Migration work completed: none. All implementation items remain `TODO`.
- Read-only discovery: inspected the harness owner specification, domain vocabulary, modular-refactor framework, comparable handoffs, phase/frontend/subsystem registries, task surface, topology references, schemas, harness families, module/platform/app/package roots, and known documentation-reading validators.
- Validation at tracker creation: `git diff --check`, `make lint-markdown`, `make generated-artifact-policy-check`, `make json-shape-check`, and command-surface inspection through `make help-all` passed. The generated-policy run root was `.cartulary/test-results/20260717T215489Z-p14438`; the JSON-shape run root was `.cartulary/test-results/20260717T215440Z-p14592`. No subsystem migration test was represented as complete.
- Next safe task: WS-00/T-001 through T-007, beginning with a reproducible exact inventory and temporary 548-row crosswalk.
- Rollback boundary: this documentation-only tracker addition can be reverted independently; it has no runtime or generated consumer.

#### 2026-07-17 — NLSpec v2 final-state contract revision

- Branch/commit: `revision/grid-adapter` at `625e5dea34d7ad743075ad766f00928924ee96b1`; dirty tree contains this tracker and `docs/testing-harness-nlspec.md` only.
- Workstream/task state: no migration workstream or implementation task was started or completed. The final-state v2 owner contract, migration decisions, hard-cutover obligations, and acceptance plan were authored; T-008 remains `TODO` until the implementation, schemas, generated outputs, and v1 deletion pass atomically.
- Authored files changed: `docs/testing-harness-nlspec.md` and this tracker. Generated files changed: none.
- Decisions recorded: sole NLSpec ownership; closed owner/family/row/verification identifiers; exact catalogs, runners, profiles, commands, defaults, terminal accounting, failure mapping, semantic evidence identity, documentation/path boundaries, evidence-class gate routing, and atomic v1 retirement. The five Graph Projection rows remain distinct, so the provisional population remains exactly 548.
- Crosswalk totals: before `0/548` frozen identities represented; after `0/548`. No baseline or crosswalk artifact was fabricated by the document revision.
- Passed checks: the local requirement audit found 216 unique requirement definitions, each with at least one `Verified by` reference; 72 unique acceptance rows with no unresolved reference; no reuse of TH-HARNESS-REQ-104 through TH-HARNESS-REQ-110 or TH-HARNESS-REQ-351; and no forbidden open-language marker in Sections 1 through 17. `git diff --check`, `make lint-markdown`, and `make generated-artifact-policy-check` passed. The final generated-policy run root was `.cartulary/test-results/20260717T230644Z-p94453`.
- Expected migration failures: `make json-shape-check` failed at `.cartulary/test-results/20260717T230231Z-p86027` because the v2 semantic helper-facade table does not match the still-v1 `tools/harness_helper_ownership.json`. `make harness-contract` failed at `.cartulary/test-results/20260717T230249Z-p86824` because the current task-surface registry has 97 public rows while the v2 table has 93 and the retired `frontend-evidence-audit` input contract still exposes v1 phase inputs. `make help-all` inspection still listed `phase-slice`, `service-backed-slice`, and `explain-phase` and no v2 owner replacements. These are hard-cutover blockers, not accepted exceptions.
- Skipped checks: `make generate`, `make generate-drift`, owner slices, catalog/audit fixtures, broad evidence gates, `make agent-finalize`, full warm `make check`, and `make release-check` were not run because WS-00 through WS-07 implementation inputs do not yet exist and generation would project the current v1 owners. No v2 conformance or atomic adoption is represented as complete.
- Next safe task: WS-00/T-001 through T-007. Freeze the exact legacy population and produce schema-valid baseline and crosswalk artifacts before changing owner catalogs or runtime surfaces.
- Rollback boundary: revert both authored-document edits together. Do not merge the adopted/current v2 text without the implementation and deletion work required by T-045A and T-046.

#### 2026-07-17 — WS-00 baseline freeze and crosswalk checkpoint

- Branch/baseline: `revision/grid-adapter` at clean starting commit `37cfdd727b3172046fbc3c5194d896a1197a381c`; the tracker baseline was intentionally rebased from `625e5dea34d7ad743075ad766f00928924ee96b1` before migration edits.
- Workstream/task state: WS-00 and T-001 through T-007 are `DONE`; no downstream workstream was started in this checkpoint.
- Authored migration evidence: `tools/test_migration_baseline.json`, `tools/test_migration_crosswalk.json`, their two temporary closed schemas under `tools/harness/migration/schemas/`, and the deterministic freeze/check helper `tools/harness/migration/freeze-test-migration.mjs`. Generated production artifacts changed: none.
- Frozen totals: 456 backend authoritative identities, 87 frontend identities, five Graph Projection identities, 548 total pending baseline keys, 37 backend support candidates, 228 Vitest classification candidates including 142 `unowned_regression`, 26 ledgers, 290 phase-shaped tracked paths, 56 visual golden digests, 68 documentation-coupling candidates, and 13 classified production follow-ups (seven authored identities and six generated descendants).
- Crosswalk totals: before `0/548` represented; after `548/548` represented as immutable pending baseline keys and `0/548` terminally adjudicated. Pending keys are replaced one-for-one by reviewed terminal dispositions in WS-04 and WS-05; they are not runtime aliases.
- Passed checks: deterministic baseline regeneration/check, strict Draft 2020-12 AJV validation for both temporary schemas, `git diff --check`, `make lint-markdown`, and `make generated-artifact-policy-check`. The generated-policy run root was `.cartulary/test-results/20260717T232058Z-p13767`.
- Expected migration failure: `make json-shape-check` failed at `.cartulary/test-results/20260717T232113Z-p14536` on the pre-existing v2 helper-facade versus v1 `tools/harness_helper_ownership.json` mismatch. The temporary schemas are intentionally outside the current schema-attachment root and were validated directly; they are removed in WS-12.
- Skipped checks: owner catalog, owner slices, generation, browser, broad verification, finalization, and release checks remain inapplicable before WS-02 through WS-10 create the v2 implementation.
- Next safe task: WS-01/T-008 and T-009, auditing the v2 owner contract against the frozen baseline and removing active procedure guidance that would direct contributors to the v1 phase surface.
- Rollback boundary: revert the WS-00 migration evidence and this tracker checkpoint together; no runtime or generated consumer depends on it.

#### 2026-07-17 — WS-01 owner contract and procedure checkpoint

- Branch/commit at start: `revision/grid-adapter` at `9a4ad28b`; the tree contained only this workstream's documentation edits.
- Workstream/task state: WS-01, T-008, and T-009 are `DONE`; no WS-02 machine owner or implementation consumer was started.
- Authored files changed: `AGENTS.md`, the modular-refactor framework, active development/bootstrap/browser/frontend/implementation/visual guides, and this tracker. Generated files changed: none. The frontend implementation-testing guide was intentionally replaced with a compact owner-catalog guide; product behavior remains owned by Core/adopted owner documents.
- Contract audit: the v2 NLSpec contains 216 unique requirement definitions and 72 unique acceptance rows with no duplicate IDs. Its current normative command, catalog, evidence, and retirement rules remain the final-state contract; the migration branch is still non-releasable until atomic parity.
- Procedure result: active contributor guidance now uses `OWNER`, semantic `ROWS`, `module-author`, `test-slice`, `service-backed-test-slice`, `explain-test-owner`, and `test-evidence-audit`. No active procedure or guide directs contributors to a retired phase command or phase input; historical retirement mentions remain in the NLSpec and migration tracker only.
- Passed checks: the requirement/acceptance uniqueness audit, active-doc command scan, `git diff --check`, and `make lint-markdown`.
- Expected migration drift: current v1 guide digests, phase-map guide restatements, generated help, and task-surface behavior do not yet match these final-state procedures. They remain branch-local blockers assigned to WS-05, WS-09, and WS-10 rather than compatibility exceptions.
- Skipped checks: JSON shape, harness contract, generation, owner slices, broad verification, finalization, and release checks remain blocked on WS-02 through WS-10 implementation inputs.
- Next safe task: WS-02/T-010 through T-013, creating reviewed verification owner contracts and replacing executable documentation parsing with machine-owner validation.
- Rollback boundary: revert the WS-01 documentation checkpoint as one unit before any consumer relies on the new commands; do not restore phase-preservation language selectively.

#### 2026-07-17 — WS-02 machine verification and documentation-decoupling checkpoint

- Branch/commit at start: `revision/grid-adapter` at `9a4a6930a7e1`; the tree contains the completed WS-02 owner contracts, consumer migrations, generated projections, validation fixtures, and this tracker checkpoint. WS-03 has not started.
- Workstream/task state: WS-02 and T-010 through T-013 are `DONE`. All 16 Section 6 coupling rows are closed. The remaining v1 frontend registry/accounting structure has no executable documentation dependency and remains an explicit T-026 deletion obligation, not a compatibility exception.
- Authored machine owners: added the closed verification registry and 14 owner contracts under `contracts/verification/`, machine design-token and object-store release-threat contracts, registered schemas, empty closed documentation-read and delivery-phase semantic policy files, the verification loader, and the documentation-boundary policy/CLI. Documentation references are excluded from semantic verification digests.
- Consumer migration: harness contract/schema/JSON validation, Network Flow accounting/fixtures/tests, reporting and report-composition traceability, OpenTelemetry conformance, SeaweedFS and release-readiness evidence, design-token generation/import policy, visual registry/spec, frontend accounting, and finalizer cache inputs now consume machine owners or inert metadata without opening documentation. `contracts/index.json.owner_document` remains inert and is not statted, resolved, hashed, or opened.
- Policy enforcement: the no-document-read guard covers tracked and untracked executable sources, Makefiles, package scripts, direct and computed filesystem operations, local helper-mediated reads, lexical containment, realpath containment, and symlink escape. Exact schema-valid exceptions are limited to documentation lint, link-check, or generation purposes; the active exception registry is empty.
- Generated outputs: refreshed `internal/gen/contracts/contracts_gen.go`, `packages/protocol-ts/src/generated/contracts.ts`, `tools/execution_topology_render_index.json`, and `tools/scheduler_manifest.json` only through Make-owned generation. The transitional v1 schedule refresh remains deletion inventory for T-040/T-045 and does not change the final owner-first contract.
- Crosswalk totals: before and after this slice, 548 frozen authoritative identities remain represented as pending baseline keys, with zero terminal dispositions, 37 backend support candidates, and 228 Vitest classification candidates still awaiting WS-03 through WS-05 adjudication. WS-02 neither created nor deleted a test identity.
- Passed validation: `make generate` at `.cartulary/test-results/20260717T234819Z-p51303`; transitional `make phase-schedules` at `.cartulary/test-results/20260718T000642Z-p62545`; `make json-shape-check` at `.cartulary/test-results/20260718T000647Z-p62819`; `make harness-contract` at `.cartulary/test-results/20260718T000647Z-p62856`; `make generate-drift` at `.cartulary/test-results/20260718T000647Z-p62815`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260718T000647Z-p62842`; `make otel-conformance` at `.cartulary/test-results/20260717T235224Z-p66991`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260717T235224Z-p66997`; `make seaweedfs-release-evidence` at `.cartulary/test-results/20260717T235237Z-p70495`; `make backend-unit` with 235 passing tests at `.cartulary/test-results/20260717T235632Z-p95206`; `make format` at `.cartulary/test-results/20260718T000553Z-p56691`; and `make test-fast` with 1,307 passing tests at `.cartulary/test-results/20260718T000159Z-p12196`. `make lint-scripts`, `make lint-shell`, and `make lint-markdown` also passed.
- Resolved migration-related failures: frontend freshness digests were refreshed after the schema-owner change; array equality in Network Flow shape validation was corrected; stale guide-ref fixtures were migrated; the reporting traceability test was corrected to call the package-owned unexported selector; the telemetry error-class registry was reconciled to the machine public-error registry; and generate-drift scratch execution now uses its actual internal target identity and avoids a nested `contracts/contracts` copy caused by a redundant input path.
- Skipped checks: owner-catalog negative fixtures, owner slices, browser stages, broad `make check`, finalization, and release check remain inapplicable until WS-03 through WS-10 complete the catalog, command, topology, and hard-cutover work. No historical phase run is accepted as v2 closure evidence.
- Next safe task: WS-03/T-014 through T-018. Mark only WS-03 `IN_PROGRESS`, then add the owner/family/runner/profile registries and unified loader before adjudicating any backend or frontend row.
- Rollback boundary: revert the WS-02 verification-owner contracts, migrated consumers, schemas, fixtures, and generated projections as one checkpoint. Do not restore individual documentation readers or retain a machine contract without its consumer validation.

#### 2026-07-18 — WS-03 unified owner-catalog checkpoint

- Branch/commit at start: `revision/grid-adapter` at tracker checkpoint `b0ae8400197a`; the implementation began from the clean completed-WS-02 tree `a5db27faf1b75e29f880fe80fee965f0346d3cae`. This entry closes the dirty WS-03 implementation tree before its cohesive commit; WS-04 has not started.
- Workstream/task state: WS-03 and T-014 through T-018 are `DONE`. The next workstream remains unstarted so there is no overlap between catalog construction and backend adjudication.
- Authored catalog contracts: added the owner registry, 12 owner manifests, runner registry, four runner-adapter contracts, exact runtime/resource/fixture profiles, four registered schemas, closed catalog loader, strict semantic JSON digesting, exact Go/Vitest/Playwright/shell selector resolution, private catalog-check CLI, and catalog import-direction enforcement. Display and documentation metadata are excluded from semantic identity.
- Ownership result: the active v2 catalog contains 12 owners, 18 families, 246 rows, and 519 exact selectors. Five Graph Projection rows are owned by `module.graphprojection`; 241 exact Vitest rows are assigned to module, web, package, architecture, or harness boundaries without a frontend or catch-all namespace. No catalog row ID contains a delivery-phase segment.
- Auxiliary adjudication: all 228 frozen Vitest candidates have a terminal auxiliary disposition. 226 candidates are retained and exact pattern expansion produces 241 non-overlapping rows. `vitest:212` is deleted because its classified title has no executable assertion and duplicates compile-time type enforcement; the absent `vitest:083` Unmapped fixture is deleted as non-production. The active classification schema and file contain zero `unowned_regression` values. All 37 frozen backend support candidates remain intentionally pending for WS-04 owner adjudication.
- Crosswalk totals: before this slice, 548 authoritative identities were pending with zero dispositions and no authorized new rows. After this slice, 543 authoritative identities remain pending, all five Graph Projection identities have reviewed `migrated` dispositions, 228 Vitest auxiliary dispositions are recorded, and 241 new catalog rows carry frozen-candidate provenance. The migration check proves every one of the 246 live catalog rows has exactly one disposition or new-row authorization.
- Structural controls: catalog validation rejects duplicate or zero-row owners, delivery-phase IDs, unresolved and cross-owner verifications, unresolved collaborators, unknown or mutated profiles, unsupported or mismatched runners, unordered references, zero/multiple/overlapping selectors, traversal, symlink components, globs, regex-like selectors, unknown shell commands, and unregistered manifests. Runner characterization covers all four closed selector kinds. Catalog code may import only its own layer and the generic contract layer; execution, accounting, scheduler, browser, and diagnostics dependencies are rejected.
- Generated output: `tools/execution_topology_render_index.json` was regenerated through Make after the authored topology and runner-adapter inputs changed. Transitional v1 phase-schedule generation remains branch-local deletion inventory for WS-09/WS-10 and is not accepted as a v2 compatibility path.
- Passed validation: migration baseline/crosswalk schema, cardinality, provenance, and catalog-authorization check; catalog check summary `sha256:7dad0a3da49268a9f050bd423addf6a08d6533ec512a43c5f2d8f336eb868a1c`; `make format` at `.cartulary/test-results/20260718T003118Z-p95288`; `make generate` at `.cartulary/test-results/20260718T003322Z-p1135`; transitional `make phase-schedules` at `.cartulary/test-results/20260718T003558Z-p8102`; `make harness-contract` with 59 passing tests at `.cartulary/test-results/20260718T003540Z-p7160`; `make json-shape-check` at `.cartulary/test-results/20260718T003603Z-p8376`; `make generate-drift` at `.cartulary/test-results/20260718T003603Z-p8371`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260718T003338Z-p2780`; and `make test-fast` with 1,307 passing tests at `.cartulary/test-results/20260718T003616Z-p10815`. `make lint-scripts` and `git diff --check` also passed.
- Resolved migration-related failures: schema-attachment ordering was corrected; the helper-ownership schema gained the `test_catalog` boundary; parameterized Vitest titles are statically expanded for `%` and named-object cases; title-specific and file-wide classifications are partitioned without selector overlap; a stale classified Network Flow title and the non-production Unmapped fixture received explicit deletions; and the migration validator now treats the frozen baseline as immutable after dispositions begin while continuing to reject cardinality, schema, authorization, and provenance drift.
- Skipped checks: backend owner slices, frontend authoritative-row slices, browser stages, semantic source renames, broad `make check`, finalization, and release check remain assigned to WS-04 through WS-11. The legacy Graph subsystem file and v1 catalogs remain only as isolated-branch rollback/reconciliation inputs until their atomic WS-10 deletion; no successor reader consumes them.
- Next safe task: create a tracker-only checkpoint marking WS-04/T-019 through T-022 `IN_PROGRESS`, then migrate backend rows in complete owner slices. Start with phase0 through phase4 identities plus the implicated frozen backend support candidates, update the crosswalk after each owner slice, and never begin WS-05 concurrently.
- Rollback boundary: revert the complete WS-03 catalog stack, owner contracts, auxiliary reconciliation, generated topology index, tests, and this tracker checkpoint together. Do not preserve a partial registry, runner, or manifest subset and do not restore `unowned_regression` as a fallback owner.

#### 2026-07-18 — WS-04 phase0–phase4 backend owner-slice checkpoint

- Branch/commit at start: `revision/grid-adapter`; WS-04 was activated by tracker checkpoint `212b88e15c95` from the clean completed-WS-03 tree `a4bfa828fa9173e6990d1c757d35fa928b0dfbb9`. This entry closes T-019 before any phase5–phase8 catalog change begins.
- Workstream/task state: WS-04 remains the only `IN_PROGRESS` workstream. T-019 is `DONE`, T-020 is the sole active task, and T-021/T-022 remain `TODO`.
- Ownership result: all 163 authoritative phase0–phase4 rows have reviewed `migrated` dispositions and all 24 implicated frozen backend-support candidates are retained as exact, explicit-only owner rows. The catalog now contains 28 owners, 77 families, 433 rows, and 794 exact selectors: the prior five Graph Projection rows and 241 Vitest auxiliary rows plus this slice's 163 authoritative and 24 support rows. No catalog row ID contains a delivery-phase segment.
- Baseline metadata correction: selector reconciliation proved that the frozen baseline had mislabeled all 29 backend Vitest identities as Go rows with empty symbol lists. The source files, source digest, baseline commit/tree, immutable keys, and `456 + 87 + 5 = 548` population did not change. The baseline now records the correct backend runner split of 335 Go, 29 Vitest, and 92 Playwright identities with exact Vitest file/title selectors; its semantic digest and the crosswalk's `baseline_digest` were updated atomically. Because no source digest, key, selector meaning, or cardinality changed, completed work was revalidated rather than reset.
- Crosswalk totals: before this slice, 543 authoritative identities were pending, five Graph Projection dispositions and 228 Vitest auxiliary dispositions were recorded, and 241 new rows were authorized. After this slice, 380 authoritative identities remain pending, 168 authoritative dispositions are recorded, 252 auxiliary dispositions are recorded, and 265 new rows are authorized. Every one of the 433 live catalog rows has exactly one disposition or new-row authorization.
- Structural result: backend rows use primary owners selected by normative postcondition, owner-qualified semantic families, immutable semantic row IDs, exact Go/Vitest/Playwright selectors, owner-local verification IDs, explicit collaborators, preserved default-check posture, and explicit runtime/resource/fixture profiles. Playwright title resolution now treats `test.describe` as ancestry and accepts a leaf or full title only when it resolves to one executable test.
- Passed validation: catalog and migration baseline/crosswalk schema, cardinality, runner-population, provenance, authorization, selector, and delivery-name checks; catalog semantic digest `sha256:9905e8a18eeb1c907b6952eea30da4f6d5f8c27216631135bf6a8c273c5a8069`; `make format` at `.cartulary/test-results/20260718T005307Z-p68672`; `make harness-contract` with 59 passing tests at `.cartulary/test-results/20260718T005316Z-p71071`; `make json-shape-check` at `.cartulary/test-results/20260718T005316Z-p71102`; `make test-fast` with 1,307 passing tests at `.cartulary/test-results/20260718T005316Z-p71094`; and `git diff --check`.
- Skipped checks: phase5–phase12 owner rows, frontend authoritative rows, semantic source renames, owner commands, browser/scheduler cutover, broad `make check`, finalization, and release checks remain assigned to T-020 onward. Existing v1 phase maps remain frozen migration inputs and no successor execution command consumes the new rows yet.
- Next safe task: T-020 only. Run the idempotent backend catalog builder for phase5 through phase8, validate exact selectors and crosswalk authorization, then update and commit this tracker before activating T-021.
- Rollback boundary: revert the complete T-019 owner slice—verification contracts, owner manifests, baseline selector correction, crosswalk dispositions, selector resolver, tests, migration builder, and this tracker entry—together. Do not retain a partial phase0–phase4 owner set or reintroduce empty Vitest baseline selectors.

#### 2026-07-18 — WS-04 phase5–phase8 backend owner-slice checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean T-019 checkpoint `fb4c1e69d0b4ec75379e27430bb150a05427fb41`. This entry closes T-020 before any phase9–phase12 catalog change begins.
- Workstream/task state: WS-04 remains the only `IN_PROGRESS` workstream. T-019 and T-020 are `DONE`, T-021 is the sole active task, and T-022 remains `TODO`.
- Ownership result: all 86 authoritative phase5–phase8 rows have reviewed `migrated` dispositions and all six implicated frozen backend-support candidates are retained as exact, explicit-only owner rows. The slice adds 105 authoritative and 10 support selector atoms across evidence, collaboration, revisions, saved views, workbook, links, view-query, view-schema, and server owners. Query grammar/control rows are owned by `platform.viewquery`, startup and generic workbook support by `module.workbook`, and saved-view persistence/lifecycle by `module.savedviews`; historical phase grouping did not override the normative postcondition owner.
- Crosswalk totals: before this slice, 380 authoritative identities were pending, 168 authoritative dispositions, 252 auxiliary dispositions, and 265 new rows were recorded. After this slice, 294 authoritative identities remain pending, 254 authoritative dispositions are recorded, 258 auxiliary dispositions are recorded, and 271 new rows are authorized. The unified catalog now contains 35 owners, 108 families, 525 rows, and 909 exact selectors, with every live row authorized once and no delivery-phase segment in a catalog row ID.
- Passed validation: exact selector preservation for 86/86 authoritative rows and 6/6 support candidates; catalog and migration schema/cardinality/provenance/authorization checks; catalog semantic digest `sha256:57f31753937655078bca30d58d205d9d65f4f51766a30798cb3dd888879622f2`; `make format` at `.cartulary/test-results/20260718T005751Z-p18833`; `make harness-contract` with 59 passing tests at `.cartulary/test-results/20260718T005759Z-p21253`; `make json-shape-check` at `.cartulary/test-results/20260718T005759Z-p21244`; `make test-fast` with 1,307 passing tests at `.cartulary/test-results/20260718T005759Z-p21240`; and `git diff --check`.
- Skipped checks: phase9–phase12 owner rows, final 456-row reconciliation, frontend authoritative rows, semantic source renames, owner commands, browser/scheduler cutover, broad `make check`, finalization, and release checks remain assigned to T-021 onward. No historical retained run is accepted as owner closure evidence.
- Next safe task: T-021 only. Run the idempotent backend catalog builder for phase9 through phase12, validate every selector and owner assignment, then checkpoint the tracker before T-022 performs full backend reconciliation.
- Rollback boundary: revert the complete T-020 owner slice—verification contracts, owner manifests, crosswalk dispositions/authorizations, migration-builder owner rules, and this tracker entry—together. Do not retain only a subset of the evidence, collaboration, revision, saved-view, or query ownership moves.

#### 2026-07-18 — WS-04 phase9–phase12 backend owner-slice checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean T-020 checkpoint `06205d601abcc9cff2de0bb400d34bdd849cf67f`. This entry closes T-021 before full backend reconciliation begins.
- Workstream/task state: WS-04 remains the only `IN_PROGRESS` workstream. T-019 through T-021 are `DONE`; T-022 is now the sole active task.
- Ownership result: all 207 authoritative phase9–phase12 rows have reviewed `migrated` dispositions and all seven implicated frozen backend-support candidates are retained as exact, explicit-only owner rows. The slice preserves 249 authoritative and 29 support selector atoms across workbook, party, assessment, task/decision, view-schema, recovery, server, authentication, job, reporting, Reference Pack, incident-bundle, import, and Network Flow owners. Network Flow measurement uses the claimed extension runtime while the unclaimed-availability row remains on the default runtime; informative measurement remains non-default and non-claim-bearing.
- Crosswalk totals: before this slice, 294 authoritative identities were pending, 254 authoritative dispositions, 258 auxiliary dispositions, and 271 new rows were recorded. After this slice, only the 87 frontend identities remain pending, all 456 backend identities plus five Graph Projection identities have `migrated` dispositions, all 37 backend support candidates plus 228 Vitest candidates have terminal auxiliary dispositions, and 278 new rows are authorized. The unified catalog contains 44 owners, 135 families, 739 rows, and 1,187 exact selectors.
- Passed validation: exact selector preservation for 207/207 authoritative rows and 7/7 support candidates; catalog and migration schema/cardinality/provenance/authorization checks; catalog semantic digest `sha256:e71c1c64ade6af4223a0a0ecb64ba0032618fe1f7afb0e1f893dec30f4ed6173`; `make format` at `.cartulary/test-results/20260718T010130Z-p67053`; `make harness-contract` with 59 passing tests at `.cartulary/test-results/20260718T010138Z-p69437`; `make json-shape-check` at `.cartulary/test-results/20260718T010138Z-p69470`; `make test-fast` with 1,307 passing tests at `.cartulary/test-results/20260718T010138Z-p69462`; and `git diff --check`.
- Skipped checks: the independent 456-row full reconciliation, frontend authoritative rows, semantic source renames, owner commands, browser/scheduler cutover, broad `make check`, finalization, and release checks remain assigned to T-022 onward. No v1 retained evidence closes a v2 owner row.
- Next safe task: T-022 only. Reconcile all 456 backend source identities and 37 support candidates against live catalog rows, exact selector atoms, owner/verifications, profiles, default-check posture, and crosswalk authorization; then close WS-04 before activating WS-05.
- Rollback boundary: revert the complete T-021 owner slice—verification contracts, owner manifests, runtime assignment, crosswalk dispositions/authorizations, migration-builder rules, and this tracker entry—together. Do not retain a partial late-phase owner population.

#### 2026-07-18 — WS-04 backend reconciliation and workstream completion

- Branch/commit at start: `revision/grid-adapter` at clean T-021 checkpoint `b22913b234ace4af4883b29690db0232c9bdedfa`. This entry closes T-022 and WS-04; WS-05 remains `TODO` and has not started.
- Workstream/task state: WS-04 and T-019 through T-022 are `DONE`. No workstream or task is active until a separate tracker activation checkpoint starts WS-05.
- Reconciliation result: the independent source-to-catalog reconciler proves all 456 frozen backend identities have one reviewed `migrated` disposition and map to 456 distinct owner rows with exactly 550 selector atoms. All 37 frozen backend-support candidates have one retained auxiliary disposition, one authorized owner row, and exactly 118 selector atoms. The combined backend catalog population is 493 distinct rows; no backend identity remains pending, overlaps another row, or retains a delivery-phase row ID.
- Preserved semantics: reconciliation checks exact package/file selectors, owners, semantic families, verifications, collaborators, runners, evidence classes, fixture/runtime/resource profiles, default-check participation, and claim posture against the frozen source rows. The authoritative split is 335 Go, 29 Vitest, and 92 Playwright rows; evidence is 186 unit, 178 integration, 79 browser, 11 visual, and two informative measurement rows. Fixture profiles reconcile as 186 none, 43 transaction, 129 template-clone, four group-clone, two migration-scratch, and 92 service-stack rows. Runtime profiles reconcile as 186 none, 261 default, and nine Network Flow claimed-extension rows. Exactly 419 authoritative rows participate in default check; 37 do not. All 13 visual/measurement rows remain informative and non-claim-bearing.
- Owner totals: `app.operator=2`, `app.server=19`, `module.assessments=6`, `module.auth=40`, `module.collaboration=15`, `module.entities=12`, `module.evidence=29`, `module.imports=3`, `module.incidentbundles=17`, `module.incidents=29`, `module.indicators=3`, `module.jobapi=3`, `module.links=3`, `module.networkflow=107`, `module.parties=3`, `module.recovery=10`, `module.reference_data=16`, `module.reporting=10`, `module.revisions=18`, `module.savedviews=6`, `module.tasksdecisions=2`, `module.timeline=43`, `module.workbook=43`, `platform.bootstrap=2`, `platform.config=6`, `platform.objectstore=1`, `platform.postgres=1`, `platform.viewquery=5`, and `platform.viewschema=2`.
- Crosswalk/catalog totals: 87 frontend identities are the only remaining pending authoritative keys; 461 authoritative dispositions cover 456 backend plus five Graph Projection rows; 265 auxiliary dispositions cover 37 backend-support plus 228 Vitest candidates; and 278 new rows are authorized. The live catalog contains 44 owners, 135 families, 739 rows, and 1,187 exact selectors with catalog digest `sha256:e71c1c64ade6af4223a0a0ecb64ba0032618fe1f7afb0e1f893dec30f4ed6173` and verification digest `sha256:c18e4e9e1d22d67a736a6b5dcc0e44f405af737167a60bd4bde1d4ca117d3a4c`.
- Passed validation: the reproducible `cartulary.test_backend_reconciliation_summary.v1` report; migration baseline/crosswalk schema, runner-population, cardinality, provenance, authorization, and catalog checks; `make format` at `.cartulary/test-results/20260718T010617Z-p16789`; `make generate` at `.cartulary/test-results/20260718T010631Z-p19242`; `make harness-contract` with 60 passing tests at `.cartulary/test-results/20260718T010643Z-p20717`; `make json-shape-check` at `.cartulary/test-results/20260718T010643Z-p20827`; `make generate-drift` at `.cartulary/test-results/20260718T010643Z-p20719`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260718T010643Z-p20886`; `make lint-scripts`; and `git diff --check`. The three row-slice checkpoints each passed `make test-fast` with 1,307 tests; the latest run root is `.cartulary/test-results/20260718T010138Z-p69462`.
- Skipped checks: owner slice commands do not exist until WS-07, and browser-stage execution, semantic renames, broad `make check`, finalization, and release checks remain assigned to WS-06 through WS-11. Exact Playwright resolution is complete, but current runtime evidence is not misrepresented as final v2 closure evidence.
- Next safe task: create a tracker-only checkpoint marking WS-05 and T-023 through T-027 `IN_PROGRESS`, then migrate FE-P0 through FE-P4 rows without reopening backend ownership or activating WS-06 concurrently.
- Rollback boundary: revert the complete WS-04 sequence from `fb4c1e69` through this completion checkpoint, including backend verification contracts, owner manifests, baseline selector correction, crosswalk dispositions/authorizations, migration/reconciliation tools, selector resolution, and focused tests. Never restore individual phase readers or keep a partial backend owner population.

#### 2026-07-18 — WS-05 FE-P0–FE-P4 owner-slice checkpoint

- Branch/commit at start: `revision/grid-adapter`; WS-05 was activated by tracker checkpoint `1ada8275a48794593f90ce6881c79663bbf515de` from the clean completed-WS-04 tree `5eb409a19104134552341f2b2fdee318ef5e39f6`. This entry closes T-023 before FE-P5–FE-P8 changes begin.
- Workstream/task state: WS-05 is the only `IN_PROGRESS` workstream. T-023 is `DONE`, T-024 is the sole active task, and T-025 through T-027 remain `TODO`.
- Reconciliation result: all 35 FE-P0–FE-P4 baseline rows have terminal dispositions: 29 `migrated`, five `consolidated`, and one `deleted`. Multi-file Vitest identities are split into exact file/title rows with nine explicit new-row authorizations. The slice preserves 107 legacy title atoms and five supported shell-target claims; six duplicate title atoms consolidate into existing backend owner rows, while three unique shell commands and 101 unique title selectors add 104 live selector identities.
- Structural decisions: generated-protocol, view-contract, selector-contract, grid-adapter, auth, workbook, saved-view, and architecture behavior is assigned to its durable owner rather than a frontend namespace. Existing auth unit, auth browser-support, Timeline browser-support, import-boundary, and generate-drift evidence is consolidated rather than double-executed. `FE-S-P0-03` is deleted as obsolete because its only behavior was phase-ledger completion accounting. The supported generated-policy, generated-drift, and frontend import-boundary commands are represented once as exact shell rows.
- Accessibility decoupling: FE-P1 through FE-P4 accessibility scenarios now obtain exact titles from test-local constants rather than reading frontend phase maps at Playwright startup. The resolver evaluates safe local static constants and array indexing, treats describe blocks as ancestry, and still requires each leaf/full title to identify one executable test.
- Crosswalk/catalog totals: before this slice, 87 frontend identities were pending, 461 authoritative dispositions and 278 new rows were recorded. After this slice, 52 frontend identities remain pending, 496 authoritative dispositions and 287 new rows are recorded. The catalog contains 45 owners, 150 families, 777 rows, and 1,291 exact selectors with digest `sha256:7939b389dee5181c85094dac51ff2cf9902accff78a7cc28c8bb71a725307fa4`.
- Passed validation: exact legacy-title and shell-command reconciliation, catalog/migration schema/cardinality/provenance/authorization checks, delivery-phase row-ID scan, `make format` at `.cartulary/test-results/20260718T011616Z-p36598`, `make harness-contract` with 60 passing tests at `.cartulary/test-results/20260718T011723Z-p43319`, `make json-shape-check` at `.cartulary/test-results/20260718T011637Z-p39247`, `make frontend-typecheck` at `.cartulary/test-results/20260718T011637Z-p39237`, `make frontend-unit` at `.cartulary/test-results/20260718T011637Z-p39222`, `make frontend-import-boundary-check` at `.cartulary/test-results/20260718T011723Z-p43334`, `make lint-scripts`, and `git diff --check`.
- Resolved validation failure: the first harness-contract run at `.cartulary/test-results/20260718T011637Z-p39224` correctly exposed migration-test assumptions that counted consolidated dispositions as independent catalog authorizations and expected all 87 frontend keys to remain pending. The tests now count only migrated/new authorizations and require zero backend pending keys; the corrected rerun passed.
- Skipped checks: FE-P5–FE-P12 rows, generic evidence accounting/audit replacement, final frontend reconciliation, semantic source renames, owner commands, browser-stage execution, broad `make check`, finalization, and release checks remain assigned to T-024 onward. Exact selector resolution is not represented as final browser runtime evidence.
- Next safe task: T-024 only. Replace FE-P5 through FE-P8 accessibility title reads, migrate those 22 rows, reconcile exact title/file selectors and the existing FE-B-P8 support overlap, then checkpoint the tracker before T-025.
- Rollback boundary: revert the complete T-023 slice—verification contracts, owner manifests, shell rows, accessibility title decoupling, crosswalk dispositions/authorizations, frontend migration builder, selector resolver, tests, and this tracker entry—together. Do not restore phase-ledger completion as a compatibility row.

#### 2026-07-18 — WS-05 FE-P5–FE-P8 owner-slice checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean T-023 checkpoint `013e7a544bf7ae1dcb491e58e4e942cc458e6284`. This entry closes T-024 before FE-P9–FE-P12 changes begin.
- Workstream/task state: WS-05 is the only `IN_PROGRESS` workstream. T-023 and T-024 are `DONE`, T-025 is the sole active task, and T-026 through T-027 remain `TODO`.
- Reconciliation result: all 22 FE-P5–FE-P8 baseline rows have reviewed `migrated` dispositions. They preserve all 47 distinct legacy title atoms in 26 owner rows; four additional rows are explicitly authorized for multi-file Vitest identities. No row was deleted or consolidated. The separately frozen Phase 8 browser-support execution remains its existing exact auxiliary owner row, while the authoritative `FE-B-P8-01` execution now has a distinct webserver-backed selector for `apps/web/e2e/phase8.workbook.spec.ts`; the duplicate title therefore cannot collapse the two execution profiles.
- Structural decisions: entity/mention behavior is owned by `module.entities`, evidence lifecycle by `module.evidence`, collaboration/conflict behavior by `module.collaboration`, query compilation by `platform.viewquery`, saved-view behavior by `module.savedviews`, and browser command-helper behavior by `module.workbook`. Playwright stage assignment distinguishes support files from webserver-backed files even when one legacy row listed both targets. The migration builder accepts a title in multiple files only by emitting or reusing an exact file/title row for every occurrence, and it preserves owner-authored verification requirements and profiles while unioning runner evidence kinds.
- Accessibility decoupling: FE-P5 through FE-P8 accessibility scenarios now use exact test-local title constants instead of opening frontend phase maps at Playwright startup. FE-P9 through FE-P11 remain assigned to T-025; no final no-map claim is made by this checkpoint.
- Crosswalk/catalog totals: before this slice, 52 frontend identities were pending, 496 authoritative dispositions and 287 new rows were recorded. After this slice, 30 frontend identities remain pending, 518 authoritative dispositions and 291 new rows are recorded. The catalog contains 45 owners, 161 families, 803 rows, and 1,338 exact selectors; runner counts are 377 Go, 123 Playwright, three shell, and 300 Vitest rows. Catalog digest is `sha256:8698e073586d28ef0a1830802a727a77ac7d1e4276e2a02fcd14edb9521954c5`; verification digest is `sha256:befde47b5d32e94ec22f30b05e47a1fa38bd8a7f963f6616e39db3ebc287cb58`.
- Passed validation: exact catalog and migration-crosswalk checks; `make format` at `.cartulary/test-results/20260718T012218Z-p51769`; final `make harness-contract` with 60 passing tests at `.cartulary/test-results/20260718T012501Z-p61784`; final `make json-shape-check` at `.cartulary/test-results/20260718T012501Z-p61767`; `make frontend-typecheck` at `.cartulary/test-results/20260718T012225Z-p54175`; `make frontend-unit` at `.cartulary/test-results/20260718T012225Z-p54206`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260718T012331Z-p58914`; `make lint-biome` at `.cartulary/test-results/20260718T012331Z-p58918`; final `make lint-scripts` at `.cartulary/test-results/20260718T012501Z-p61797`; final `make lint-markdown` at `.cartulary/test-results/20260718T012501Z-p61772`; and `git diff --check`.
- Resolved validation failure: the first migration-builder pass rejected `FE-B-P8-01` because its title occurs once in both the webserver-backed and support Playwright files. Exact multi-file expansion and file-sensitive stage selection now retain both executions without ambiguity or duplicate authorization. Diff review also caught and removed an attempted replacement of established owner-contract prose/profile fields; the builder now changes only the runner evidence-kind set for an existing verification.
- Skipped checks: FE-P9–FE-P12 rows, generic evidence accounting/audit replacement, final frontend reconciliation, semantic source renames, owner commands, browser-stage runtime execution, broad `make check`, finalization, and release checks remain assigned to T-025 onward. Exact Playwright resolution is structural evidence, not final browser runtime evidence.
- Next safe task: T-025 only. Decouple the remaining FE-P9 through FE-P11 accessibility titles, migrate the final 30 frontend identities, preserve FE-P12 Network Flow runtime/fixture posture, and checkpoint before T-026.
- Rollback boundary: revert the complete T-024 slice—owner manifests, the `module.entities` Vitest verification-kind addition, accessibility title decoupling, crosswalk dispositions/authorizations, multi-file and file-sensitive-stage builder behavior, and this tracker entry—together. Do not collapse the webserver-backed and support Playwright executions into one selector.

#### 2026-07-18 — WS-05 FE-P9–FE-P12 owner-slice checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean T-024 checkpoint `5dd2033ce5eeb665fac43c932c71d6fa0b88c96c`. This entry closes T-025 before generic accounting changes begin.
- Workstream/task state: WS-05 is the only `IN_PROGRESS` workstream. T-023 through T-025 are `DONE`, T-026 is the sole active task, and T-027 remains `TODO`.
- Reconciliation result: all final 30 frontend baseline rows have terminal dispositions: 26 `migrated`, one `consolidated`, and three `deleted`. Their 39 legacy title atoms reconcile as 38 selectors in 29 newly authorized owner rows plus the existing exact Timeline browser-support survivor for `FE-B-P10-02`; three additional rows are explicit multi-file Vitest splits. The three FE-P11 support rows are deleted as duplicate command-composition accounting: their supported generic gates remain independently owned, while phase-ledger/schedule completion and frontend-phase release joins are not retained behavior.
- Structural decisions: revision/history behavior is owned by `module.revisions`; inspector, coordination, keyboard, clipboard, and visual behavior by `module.workbook`; claimed extension behavior by `module.networkflow`; and global visual/accessibility readiness by `web.design`. The four `web.design` rows use the existing `web.design.verification.readiness_direction` contract rather than creating a generic parallel verification. All newly retained Network Flow Playwright rows use `network_flow_claimed`, `browser_exclusive`, and `service_stack`; accessibility, visual, and measurement rows remain explicit-only informative evidence.
- Documentation decoupling: FE-P9 through FE-P11 accessibility titles are local exact constants, FE-P12 was already literal, and the now-unused `apps/web/e2e/support/accessibility/phaseMap.ts` documentation/map reader is deleted. No Playwright test loads frontend phase maps for title discovery or activation.
- Migration-check durability: the temporary baseline validator now validates the immutable stored baseline and crosswalk directly once migration dispositions exist. It no longer reconstructs or opens retired baseline paths, while preserving schema, digest, 548-key cardinality, auxiliary-candidate, authorization, and live-catalog parity checks.
- Crosswalk/catalog totals: before this slice, 30 frontend identities were pending, 518 authoritative dispositions and 291 new rows were recorded. After this slice, no authoritative key is pending, all 548 identities have terminal dispositions, and 294 new rows are authorized. Frontend dispositions total 77 migrated, six consolidated, and four deleted. The catalog contains 46 owners, 169 families, 832 rows, and 1,376 exact selectors; runner counts are 377 Go, 143 Playwright, three shell, and 309 Vitest rows. Catalog digest is `sha256:a96f2dc00d868302e0c208958b14ed6d38a5b50ade29abd17148a946ebf7e733`; verification digest is `sha256:6c4b33052cf1c0646066c9420dcb4c137807f7ffc5d7fafe3f49f8d0d58e8e57`.
- Passed validation: exact catalog and migration-crosswalk checks; `make format` at `.cartulary/test-results/20260718T012857Z-p68773`; `make harness-contract` with 60 passing tests at `.cartulary/test-results/20260718T012904Z-p71275`; `make frontend-typecheck` at `.cartulary/test-results/20260718T012904Z-p71300`; `make frontend-unit` at `.cartulary/test-results/20260718T012904Z-p71213`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260718T012904Z-p71162`; `make lint-biome` at `.cartulary/test-results/20260718T012904Z-p71203`; `make lint-scripts` at `.cartulary/test-results/20260718T012904Z-p71241`; corrected `make json-shape-check` at `.cartulary/test-results/20260718T013035Z-p77214`; and `git diff --check`.
- Resolved validation failure: the first `make json-shape-check` run at `.cartulary/test-results/20260718T012904Z-p71153` saw the intended helper deletion as an unstaged tracked path and attempted to open it. Staging that deletion removed it from the authored tracked-file inventory; the unchanged check then passed. This was an index/worktree transition during the cohesive slice, not a retained reader or compatibility restoration.
- Skipped checks: generic owner accounting/audit replacement, independent 87-row frontend reconciliation, semantic source renames, owner commands, browser-stage runtime execution, broad `make check`, finalization, and release checks remain assigned to T-026 onward. Structural selector/profile closure is not final browser runtime evidence.
- Next safe task: T-026 only. Replace frontend row accounting and audit internals with runner-neutral owner accounting, remove registry/guide/cumulative/base-phase joins from executable paths, and preserve current public v1 wiring only until the later atomic interface cutover.
- Rollback boundary: revert the complete T-025 slice—owner manifests/registry, Network Flow and design verification references, accessibility constants and helper deletion, crosswalk dispositions/authorizations, migration-builder verification routing, stored-baseline checker behavior, and this tracker entry—together. Do not restore document-driven title discovery.

#### 2026-07-18 — WS-05 catalog-native owner accounting checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean T-025 checkpoint `b40742ede50876a7692cfeb74edee51cd8c1d760`. This entry closes T-026 before final frontend reconciliation begins.
- Workstream/task state: WS-05 is the only `IN_PROGRESS` workstream. T-023 through T-026 are `DONE`, and T-027 is the sole active task.
- Successor accounting result: `tools/harness/evidence-accounting` is a registered harness owner boundary with one catalog-native facade. It resolves an active owner and optional exact row set from the unified catalog, rejects blank, duplicate, unknown, foreign-owner, zero-row, and target-incompatible selections, orders rows by immutable owner-qualified ID, and projects exact family, verification, runner, selector-digest, evidence-class, profile, default-check, claim-posture, and target identities.
- Target and profile semantics: Vitest rows route to `frontend-unit`; Playwright rows route only through the closed stage-to-target map; shell rows resolve through registered command IDs; Go rows deliberately retain no guessed target because WS-07 owner-slice planning will supply their exact work units. Target-specific projections recompute semantic profile digests from only the selected rows. Catalog, verification, runtime-profile, resource-profile, fixture-profile, and selector digests all use the existing strict semantic-JSON implementation.
- Coupling removal: the successor accounting core contains no frontend namespace, phase ID, cumulative dependency, base-phase join, guide path/digest, documentation reference, or legacy registry/map read. Pre-cutover v1 accounting/audit files remain comparison-only inputs on the isolated branch and are a mandatory WS-10 deletion; they are not imported by the successor core and do not define its ownership or selection. Public artifacts, final schemas, and audit CLI integration remain assigned to T-034/T-035 as planned rather than introducing a temporary schema or translated artifact.
- Ownership guardrail: helper ownership gained the `evidence_accounting` boundary and facade, and the import-boundary scanner recognizes that top-level owner root. Catalog code still has no dependency on execution or accounting; dependency direction is accounting-to-catalog only.
- Passed validation: `make format` at `.cartulary/test-results/20260718T013843Z-p87994`; final `make harness-contract` with 62 passing tests at `.cartulary/test-results/20260718T013850Z-p90389`; `make json-shape-check` at `.cartulary/test-results/20260718T013850Z-p90369`; `make lint-scripts` at `.cartulary/test-results/20260718T013850Z-p90349`; and `git diff --check`.
- Resolved validation failure: the first harness-contract run at `.cartulary/test-results/20260718T013753Z-p86568` passed all new accounting behaviors but correctly rejected the stale exact helper-facade population of 35. Registering the new facade changes that closed population to 36; the final rerun passed all 62 tests.
- Skipped checks: final evidence-accounting/audit schemas and retained artifact integration are owned by T-034/T-035; public command cutover by T-039; v1 reader deletion by T-043/T-046; final browser-stage runtime, broad `make check`, finalization, and release checks remain later workstream gates. No v1 artifact is represented as successor evidence.
- Next safe task: T-027 only. Independently reconcile all 87 frontend baseline identities, terminal dispositions, exact source selectors, catalog rows, owners, verifications, profiles, and the 548-row crosswalk closure before closing WS-05.
- Rollback boundary: revert the evidence-accounting owner root, helper ownership/schema enum, import-boundary registration, focused tests, and this tracker entry together. Do not move accounting into the catalog owner or restore documentation-derived selection in the successor path.

#### 2026-07-18 — WS-05 frontend reconciliation and workstream completion

- Branch/commit at start: `revision/grid-adapter` at clean T-026 checkpoint `fc6937c62775997387b97b6f0822607bc7af6fb5`. This entry closes T-027 and WS-05; WS-06 remains `TODO` and has not started.
- Workstream/task state: WS-05 and T-023 through T-027 are `DONE`. No workstream or task is active until a separate tracker activation checkpoint starts WS-06.
- Independent reconciliation result: the source-to-catalog reconciler proves all 87 frozen frontend identities have one terminal reviewed disposition: 77 migrated, six consolidated, and four deleted. All 193 unique legacy title atoms are retained by exact catalog selectors; they resolve 194 times because `FE-B-P8-01` intentionally has separate webserver-backed and support executions. All 23 command atoms are terminally adjudicated: seven resolve once through registered shell rows and 16 belong to the four explicitly deleted phase-ledger/release-composition rows.
- Catalog authorization and selector closure: the frontend migration authorizes 16 additional split rows and references 101 distinct live catalog rows after consolidation and exact duplicate execution are included. Every retained row is authorized by a migrated baseline identity or explicit new-row provenance, every exact selector resolves through the unified loader, every consolidated survivor remains independently authorized, and no frontend baseline key remains pending. The complete crosswalk has exactly 548 terminal authoritative dispositions and 265 terminal auxiliary dispositions.
- Semantic preservation: reconciliation independently checks durable owner, owner-local verification, runner, exact title/command coverage, evidence class, Playwright stage, runtime/resource/fixture profile, default-check posture, claim posture, and delivery-neutral row IDs. The frozen runner split remains 56 Playwright, 25 Vitest, and six shell identities. Direct retained row references comprise 13 accessibility, 26 browser, one measurement, seven static, 43 unit, and 14 visual records; repeated references are expected where consolidated identities share a survivor.
- Owner reference totals: `harness.generated_artifacts=4`, `module.auth=11`, `module.collaboration=10`, `module.entities=5`, `module.evidence=5`, `module.networkflow=7`, `module.revisions=2`, `module.savedviews=5`, `module.timeline=2`, `module.workbook=36`, `package.grid_adapter=6`, `package.protocol_ts=1`, `package.ui=1`, `package.view_contracts=1`, `platform.viewquery=1`, `web.architecture=3`, and `web.design=4`. These are reconciled row references rather than unique catalog-owner populations.
- Final WS-05 catalog state: 46 owners, 169 families, 832 rows, and 1,376 exact selectors; 377 Go, 143 Playwright, three shell, and 309 Vitest rows. Catalog digest is `sha256:a96f2dc00d868302e0c208958b14ed6d38a5b50ade29abd17148a946ebf7e733`; verification digest is `sha256:6c4b33052cf1c0646066c9420dcb4c137807f7ffc5d7fafe3f49f8d0d58e8e57`.
- Passed validation: reproducible `cartulary.test_frontend_reconciliation_summary.v1`; complete migration baseline/crosswalk schema, cardinality, authorization, and catalog parity checks; `make format` at `.cartulary/test-results/20260718T014440Z-p98836`; `make harness-contract` with 63 passing tests at `.cartulary/test-results/20260718T014449Z-p1580`; `make json-shape-check` at `.cartulary/test-results/20260718T014449Z-p1598`; `make frontend-typecheck` at `.cartulary/test-results/20260718T014449Z-p1657`; `make frontend-unit` at `.cartulary/test-results/20260718T014449Z-p1584`; `make lint-scripts` at `.cartulary/test-results/20260718T014449Z-p1630`; and `git diff --check`.
- Skipped checks: owner-slice commands do not exist until WS-07; semantic filenames/titles/goldens are WS-06; browser-stage runtime execution and topology are WS-08; broad `make check`, finalization, and release checks are WS-11. Exact selector/profile reconciliation is not represented as final browser runtime evidence.
- Next safe task: create a tracker-only checkpoint marking WS-06 and T-028 through T-031 `IN_PROGRESS`, then perform semantic renames in complete backend, frontend, and fixture/golden slices without starting WS-07 concurrently.
- Rollback boundary: revert the complete WS-05 sequence from `013e7a54` through this completion checkpoint, including frontend owner manifests, verification-kind changes, title decoupling, accounting core, crosswalk dispositions/authorizations, migration/reconciliation tools, and focused tests. Never restore documentation-driven activation or a separate frontend ownership namespace.

#### 2026-07-18 — WS-06 activation checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean WS-05 completion checkpoint `4cf06958b46649f76d79f1232379beb2738dc201`.
- Workstream/task state: WS-06 is the only `IN_PROGRESS` workstream. T-028 is the sole active task; T-029 through T-031 remain `TODO`. WS-07 has not started.
- Starting invariants: all 548 authoritative identities and 265 auxiliary candidates are terminally reconciled; the catalog has 46 owners, 169 families, 832 rows, and 1,376 selectors; all 56 frozen visual golden digests remain the WS-00 byte-identity baseline. Renames must update catalog selectors atomically and must not reopen ownership adjudication.
- Next safe task: T-028 only. Classify the frozen Go delivery-identity population, rename complete package-owner slices, update exact catalog selectors, and run focused Go plus semantic scans before activating T-029.
- Rollback boundary: revert a complete semantic owner slice, including source names, exact catalog selectors, semantic allow classifications, tests, and tracker evidence. Never leave duplicate old/new tests or a selector that resolves both identities.

#### 2026-07-18 — T-028 semantic Go identity checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean WS-06 activation checkpoint `001f67f8`. This entry closes T-028; T-029 becomes the sole active task while WS-06 remains the only active workstream.
- Rename result: 102 delivery-shaped Go test paths were moved to semantic paths; 632 `Test`/`Benchmark`/`Fuzz` entry points and 333 test helper/type/fixture declarations had delivery-phase, sprint, or legacy phase-row identity removed. Collision checks ran before each mechanical pass. The two names that would otherwise overlap production `ApprovedMapping` and `FlowRow` types were resolved as the explicit `approvedMappingFixture` and `flowRowFixture` helpers.
- Selector and policy result: all 456 authoritative backend dispositions and 37 retained support candidates still reconcile to 550 authoritative plus 118 support selector atoms. Phase-map and owner-family selectors changed atomically, support regex selectors became closed exact/alternation selectors, and the catalog remains 46 owners, 169 families, 832 rows, and 1,376 exact selectors. The catalog semantic digest is now `sha256:2dcbd9b4380fe1479e768ff25e049d5ed3696ef077dba80e40f73a8d6a3f0ef4`; the verification digest remains `sha256:6c4b33052cf1c0646066c9420dcb4c137807f7ffc5d7fafe3f49f8d0d58e8e57`.
- Structural result: `tools/harness/test-catalog/semantic-identity-check-cli.mjs` now rejects delivery-shaped Go filenames, declarations, and catalog selectors, including legacy phase-row fragments. The pre-cutover `phase-test-name-check` entry point delegates to that inverse policy until WS-09/WS-10 replace and delete the old target/path. Transitional phase schedules were regenerated through `make phase-schedules`; this does not authorize retaining the generator or generated phase topology after WS-10.
- Passed validation: idempotent migration dry run; direct semantic identity scan; exact zero counts for delivery-shaped Go test paths, entry points, and declarations; backend reconciliation; `make format` at `.cartulary/test-results/20260718T015935Z-p47439`; `make test-fast` with 1,295 tests and zero failures at `.cartulary/test-results/20260718T015938Z-p49713`; `make json-shape-check` at `.cartulary/test-results/20260718T020150Z-p93139`; `make phase-schedule-drift` at `.cartulary/test-results/20260718T020150Z-p93133`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260718T020150Z-p93194`; `make harness-contract`; `make phase-test-name-check`; `make lint-scripts`; and `git diff --check`.
- Resolved failures: the first JSON-shape runs exposed unstaged rename discovery, stale support selection prefixes, legacy manifest row-ID naming enforcement, and expected generated schedule drift; staging the atomic moves, closing support selectors, replacing the naming policy, and regenerating from owners resolved them. The first `make test-fast` root `.cartulary/test-results/20260718T015819Z-p37028` exposed the two production-type collisions described above; the full rerun passed.
- Skipped checks: frontend unit/typecheck and browser stages are reserved for T-029/T-030; the full semantic repository scan and production classifications are T-031; owner commands do not exist until WS-07; broad `make check`, finalization, and release checks remain WS-11.
- Next safe task: T-029 only. Rename Vitest and Playwright files, suites, titles, helpers, and exact catalog/transition selectors in closed frontend-owner slices, then run frontend and semantic validation before activating T-030.
- Rollback boundary: revert this complete T-028 checkpoint, including source moves, Go declarations, catalog and transitional selectors, semantic policy, generated schedule outputs, and ledger references. Never restore duplicate old/new files or the former requirement that executable names carry delivery-row IDs.

#### 2026-07-18 — T-029 semantic frontend identity checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean T-028 checkpoint `16835f3361f80f84f74237cac67f500bc7deac9b`. This entry closes T-029; T-030 becomes the sole active task while WS-06 remains the only active workstream.
- Rename result: 53 Vitest/Playwright test paths moved to semantic filenames, including functional, support, accessibility, visual, and measurement surfaces. The migration rewrote 403 distinct delivery-shaped or legacy title identities and 29 helper/type declarations, updated exact source references, and preserved dotted `.support.` filenames where support-stage discovery depends on that semantic role. Its check mode is idempotent and explicitly excludes caches such as `.tsbuildinfo`.
- Catalog and accounting result: all 548 authoritative identities and 265 frozen auxiliary dispositions remain terminal. A previously executable Timeline WebSocket assertion was found to have depended only on the v1 `Phase 3 support` title heuristic; it is now the explicitly authorized `web.workbook.regression.timelineworkbook_websocket_invalidation_focus_continuity` row with `web.collaboration` as collaborator and an exact `.support.` selector. The catalog therefore contains 46 owners, 169 families, 833 rows, and 1,377 exact selectors: 377 Go, 143 Playwright, three shell, and 310 Vitest rows. Catalog digest is `sha256:e54272a89ceec50eeb2b6a1ef32413d66c94220d938e50eaa2a29215530d1981`; verification digest remains `sha256:6c4b33052cf1c0646066c9420dcb4c137807f7ffc5d7fafe3f49f8d0d58e8e57`.
- Structural result: the semantic-identity policy rejects delivery-shaped frontend test paths, declarations, direct suite/test titles, and catalog selectors. Transitional frontend maps, ledgers, registry digests, duration inputs, and phase schedules were updated atomically for pre-cutover validation; this does not authorize retaining those v1 inputs after WS-10. Freshness regeneration uses the declared phase-accounting facade rather than a private import.
- Passed validation: idempotent frontend migration dry run; direct semantic identity scan; complete baseline/crosswalk authorization check; backend and frontend reconciliation; `make format` at `.cartulary/test-results/20260718T022915Z-p56351`; `make frontend-unit` with 600 tests and zero unmapped entries at `.cartulary/test-results/20260718T022623Z-p47724`; `make frontend-typecheck` at `.cartulary/test-results/20260718T022724Z-p52496`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260718T022727Z-p52972`; `make lint-biome` at `.cartulary/test-results/20260718T022633Z-p49144`; `make json-shape-check` at `.cartulary/test-results/20260718T022626Z-p48243`; final `make harness-contract` with 63 passing tests at `.cartulary/test-results/20260718T022720Z-p52149`; `make phase-ledger-drift` at `.cartulary/test-results/20260718T022730Z-p53489`; and `make phase-schedule-drift` at `.cartulary/test-results/20260718T022800Z-p54841`.
- Resolved failures: focused runs exposed a stale relative test filename, one implicit support-title accounting dependency, support-file stage misclassification, stale frontend manifest/freshness digests, a local helper collision, and a private phase-accounting import. Exact semantic references, an owner-qualified catalog row, dotted support filenames, deterministic digest regeneration, the `editExtendedSurfaceCell` helper, and facade-only access resolved them without a catch-all classification or compatibility alias.
- Skipped checks: fixture/scenario/golden renames and browser runtime gates are T-030; the complete repository semantic classification and production follow-up ledger are T-031; owner commands begin in WS-07; broad `make check`, finalization, and release checks remain WS-11.
- Next safe task: T-030 only. Rename delivery-shaped fixture/scenario/artifact/golden identities, update exact consumers, prove all 56 frozen golden byte digests are unchanged by path moves, and run the applicable browser/harness gates before activating T-031.
- Rollback boundary: revert this complete T-029 checkpoint, including 53 file moves, titles/helpers, catalog/crosswalk authorization, transitional map/ledger/registry/topology updates, semantic policy, and focused tests. Never restore title inference, duplicate old/new test files, or the former delivery-shaped selectors.

#### 2026-07-18 — T-030 semantic fixture and golden identity checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean T-029 checkpoint `465e39ea`. This entry closes T-030; T-031 becomes the sole active task while WS-06 remains the only active workstream.
- Rename result: all 56 frozen visual goldens moved to semantic snapshot paths and all 15 bootstrap diagnostic goldens moved from the delivery-shaped `phase0` directory to `bootstrap/diagnostics`. The checked migration report records 16 retained current visual fixtures, six deleted unowned placeholder fixtures, zero residual text rewrites, and SHA-256 identity for every before/after golden pair; its final report digest is `ff339ae9190b27321717125bb9b16169a6e2e8a3bc1fd17224a5ac0bf7b13f03`.
- Fixture-contract result: the visual registry now has only the 16 executable current fixtures, with immutable semantic `visual.fixture.*` and `visual.seed.*` identities, semantic golden filenames, and exact owner-catalog scenario resolution. Placeholder/missing fixture states no longer participate in live selection. The visual fixture schema, loader, Playwright assertions, transitional maps/ledgers, support inventory, and generated topology references changed atomically; no visual-update target was used.
- Behavioral preservation: exact rendered labels that form part of frozen captures remain test data rather than accounting identity. The migration script explicitly preserves those values in the visual spec and its Network Flow workspace helper while still rejecting delivery-shaped fixture paths, IDs, metadata, and artifact labels. This keeps the pixel contract byte-stable without treating the labels as selectors or ownership metadata.
- Reconciliation result: all 548 authoritative dispositions and 265 auxiliary dispositions remain terminal. The catalog remains 46 owners, 169 families, 833 rows, and 1,377 exact selectors with digest `sha256:e54272a89ceec50eeb2b6a1ef32413d66c94220d938e50eaa2a29215530d1981`; backend reconciliation remains 456 authoritative rows/550 selector atoms plus 37 support rows/118 selector atoms, and frontend reconciliation remains 87 authoritative rows with 193 legacy title atoms resolving through 194 exact title selectors.
- Passed validation: idempotent fixture migration check; frozen digest comparison; semantic identity scan; migration baseline/crosswalk validation; backend and frontend reconciliation; final `make format` at `.cartulary/test-results/20260718T024727Z-p77195`; `make browser-e2e-visual` with all 29 scenarios passing at `.cartulary/test-results/20260718T024735Z-p79582`; `make frontend-unit` with 600 tests at `.cartulary/test-results/20260718T023807Z-p72653`; `make frontend-typecheck` at `.cartulary/test-results/20260718T024014Z-p2880`; `make lint-biome` at `.cartulary/test-results/20260718T024019Z-p4947`; `make test-fast` with 1,295 tests at `.cartulary/test-results/20260718T023904Z-p82646`; final `make json-shape-check` at `.cartulary/test-results/20260718T025009Z-p2181`; `make harness-contract`; `make phase-ledger-drift` at `.cartulary/test-results/20260718T025009Z-p2218`; `make phase-schedule-drift` at `.cartulary/test-results/20260718T025009Z-p2276`; `make lint-markdown`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260718T025009Z-p2292`; and `git diff --check`.
- Resolved failures: early JSON-shape and fast-test runs exposed a stale support-inventory golden path and stale Go path components. The first two browser runs then proved that semantic test-data rewrites had changed rendered pixels in Timeline, collaboration, and Network Flow captures. Restoring only the frozen rendered values and excluding those values from the idempotent migration pass resolved the failures; semantic paths, fixture IDs, snapshot names, selectors, and owner accounting remain migrated.
- Skipped checks: the complete repository semantic classification and production follow-up ledger are T-031; owner commands begin in WS-07; broad `make check`, finalization, and release checks remain WS-11.
- Next safe task: T-031 only. Reconcile the 13 frozen production-source follow-ups, populate reviewed semantic allow rules only for legitimate product/runtime phases or execution steps, strengthen the repository scan and negative fixtures, and close WS-06 before activating WS-07.
- Rollback boundary: revert this complete T-030 checkpoint, including semantic fixture IDs and paths, exact consumers, visual registry/schema simplification, all 71 golden moves, generated transitional outputs, semantic policy, and tracker evidence. Never restore missing placeholder fixtures or change golden bytes to make a rename pass.

#### 2026-07-18 — T-031 and WS-06 semantic identity closure checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean T-030 checkpoint `86fc4b877464`. This entry closes T-031 and WS-06; no workstream or task remains active until a separate tracker checkpoint starts WS-07.
- Production-path result: the 13 frozen SQL/query, module, and generated descendants now use semantic incident, recovery, reporting, saved-view, timeline, and workbook-startup filenames. The additional authentication and incident-directory debug harnesses and recovery browser-restore helper are semantic paths as well. Generated SQL descendants were changed only by `make generate` at `.cartulary/test-results/20260718T025500Z-p12537`.
- Selector-contract result: delivery-phase UI-selector aliases were removed instead of retained. Application, Vitest, and Playwright consumers use the semantic authentication, account, deployment-administration, incident-directory, app-route, and public-error helpers directly. Debug selector IDs, transaction prefixes, fixture row IDs, policy-test titles, catalog selectors, and transitional accounting classifications changed atomically.
- Follow-up result: `tools/delivery_identity_followup_ledger.json` and its migration-only schema close 16 path dispositions, two symbol dispositions, and four retained product protocol identifiers with explicit owners and reasons. The retained identifiers are persisted product/import formats, not harness identities; the closure checker proves that each occurs exactly once in its owner source and never appears in an owner-catalog selector. No legitimate numbered product phase or execution-step identity required an allowlist entry, so `tools/delivery_phase_semantic_allowlist.json` remains closed and empty.
- Policy result: the semantic identity checker validates allowlist shape, owner existence, ordering, uniqueness, and exact candidate resolution, and scans live helper/production identity paths and declarations in addition to tests and catalog selectors. Harness contract fixtures cover exact legitimate matches plus unmatched, duplicate, line-number, and unknown-owner rejection. The closure scan reports zero unclassified delivery-shaped paths.
- Reconciliation result: all 548 authoritative dispositions and 265 auxiliary dispositions remain terminal. The catalog remains 46 owners, 169 families, 833 rows, and 1,377 exact selectors with digest `sha256:8e065437a6f7c03ceb4c7b3a752b0cedf3cf34c767ee098581b1bca5f9cc4aa8`; no owner, family, row, selector, or verification population changed.
- Passed validation: migration baseline/crosswalk check; delivery-identity follow-up closure; direct semantic identity and catalog checks; `make format` at `.cartulary/test-results/20260718T030541Z-p99542`; final `make json-shape-check` at `.cartulary/test-results/20260718T030544Z-p2151`; `make frontend-unit` with 649 tests and zero unmapped entries at `.cartulary/test-results/20260718T030846Z-p10247`; `make harness-contract` at `.cartulary/test-results/20260718T030918Z-p12815`; `make test-fast` with 1,295 tests at `.cartulary/test-results/20260718T030935Z-p13723`; `make frontend-typecheck` at `.cartulary/test-results/20260718T031134Z-p55032`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260718T031134Z-p55047`; `make lint-biome` at `.cartulary/test-results/20260718T031134Z-p55068`; `make generate-drift` at `.cartulary/test-results/20260718T031134Z-p55087`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260718T031158Z-p58267`; `make phase-ledger-drift` at `.cartulary/test-results/20260718T031158Z-p58234`; `make phase-schedule-drift` at `.cartulary/test-results/20260718T031158Z-p58255`; `make browser-e2e-stateful` with 25 tests at `.cartulary/test-results/20260718T031222Z-p59447`; and final script/shell/Markdown lint plus staged diff checks.
- Resolved failures: initial shape, type, and unit runs identified stale renamed imports/catalog titles; exact semantic references resolved them. Two later frontend runs identified exact transitional accounting titles still using the former wording; the residual dossier named both entries, and exact classification updates restored zero-unmapped accounting without a catch-all. One concurrently executed harness-contract run timed out during SeaweedFS readiness; the isolated rerun passed. These were migration-related reference updates except the concurrency-sensitive service-readiness failure.
- Skipped checks: the public `test-catalog-check` target does not exist before T-032/T-035, so the attempted target failed with Make usage status and the catalog’s checked CLI supplied the T-031 structural evidence. Owner commands, owner artifacts, and their public smoke tests begin in WS-07; broad `make check`, finalization, and release checks remain WS-11.
- Next safe task: create a tracker-only checkpoint marking WS-07 and T-032 `IN_PROGRESS`, then implement owner-slice planning and execution from the unified catalog without starting WS-08 or exposing a compatibility alias.
- Rollback boundary: revert this complete T-031 checkpoint, including semantic production/helper paths, regenerated SQL descendants, selector facade removal, exact consumer updates, follow-up ledger/schema/checker, semantic policy fixtures, and tracker evidence. Never restore delivery-phase aliases or let a retained product protocol identifier become a test selector.

#### 2026-07-18 — WS-07 activation checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean WS-06 closure checkpoint `ddfd2afa`. WS-07 is the only `IN_PROGRESS` workstream and T-032 is the sole active task; T-033 through T-035 remain `TODO`.
- Active scope: implement catalog-native `test-slice OWNER=<owner-id> [ROWS=...]` and `service-backed-test-slice OWNER=<owner-id> [ROWS=...]` planning/execution with exact selection, usage, JSON, service-backed, cancellation, cleanup, and failure-normalization fixtures. This checkpoint does not activate diagnostics, evidence audit, artifact replacement, browser topology, or public task-surface retirement assigned to later tasks.
- Rollback boundary: revert the complete T-032 implementation checkpoint before activating T-033. Do not add phase aliases, translated plans, old-catalog readers, or fallback selection.

#### 2026-07-18 — T-032 owner-slice planning and execution checkpoint

- Branch/commit at start: `revision/grid-adapter` at clean WS-07 activation checkpoint `743fd3f5`. T-032 is `DONE`, T-033 becomes the sole active task, and WS-07 remains the only `IN_PROGRESS` workstream.
- Selection result: both successor commands resolve only the unified owner catalog. Omitted `ROWS` selects the complete active owner inventory for `test-slice` and the runtime-profile service-backed subset for `service-backed-test-slice`; explicit selection is owner-qualified, sorted, duplicate-free, non-empty, and all-or-nothing. `default_check` is absent from slice narrowing. Missing, malformed, blank, duplicate, unknown, cross-owner, zero-row, and explicit non-service selections fail with usage exit `2` before a result directory is created.
- Plan result: `cartulary.test_slice_plan.v1` is a closed attached schema carrying the exact selection, semantic source/catalog/verification/profile digests, normalized rows, deterministic runner/profile work units, resource claims, dependencies, finalizers, expected artifacts, and unused worker inputs. The source snapshot hashes every tracked and non-ignored untracked regular file or symlink without following links. Identical semantic inputs produce byte-identical plans apart from caller-supplied run/time identity.
- Execution result: exact Go tests are selected by package and escaped test-name alternation, exact Vitest cases by file and escaped title alternation, shell rows by their registered command target, and managed-service work through the owned test-services lifecycle. Runner arguments are passed without shell interpretation and retained work-unit logs are redacted. Playwright rows remain plan-valid but cannot execute through a phase fallback; their catalog-native stage executor is the explicit WS-08/T-036 dependency.
- Boundary result: the new owner-slice facade is registered in helper ownership under the scheduler boundary. Top-level harness-root recognition now derives owner roots from that registry, preventing a second hard-coded ownership list. Go evidence-target routing is semantic-family based, so owner accounting no longer leaves Go rows targetless.
- Behavioral evidence: exact one-row `platform.config` Go execution passed at `.cartulary/test-results/20260718T032638Z-p13940`; omitted-row JSON execution selected and passed all seven owner rows, including non-default rows, at `.cartulary/test-results/20260718T032649Z-p14986`; an exact `package.protocol_ts` Vitest row passed at `.cartulary/test-results/20260718T033242Z-p82534`; and an exact managed-service `module.assessments` row passed at `.cartulary/test-results/20260718T032922Z-p26441`. Target-local JSON emitted exactly one `cartulary.test_slice_scheduler_summary.v1` object, and an explicit empty `ROWS=` fixture exited `2` with no retained run directory.
- Passed validation: `make format` at `.cartulary/test-results/20260718T032853Z-p23077`; `make harness-contract` with 68 tests at `.cartulary/test-results/20260718T033255Z-p82912`; `make test-fast` with 1,295 tests at `.cartulary/test-results/20260718T032951Z-p34701`; final `make json-shape-check` at `.cartulary/test-results/20260718T033255Z-p82928`; `make generate-drift` at `.cartulary/test-results/20260718T033204Z-p75812`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260718T033255Z-p82888`; and script/Biome lint plus staged diff checks.
- Resolved failures: schema attachment and helper-facade arrays initially violated required ASCII ordering, and the import boundary initially treated the new facade directory as unknown. Sorting the owner inputs and deriving known roots from registered facades resolved those migration-related failures without an exception. No product assertion failed.
- Skipped checks: Playwright owner execution is assigned to WS-08, full per-row terminal accounting and audit are T-034, complete artifact/schema replacement is T-035, generated public help/input ownership is WS-09, and broad `make check`, finalization, and release validation remain WS-11.
- Next safe task: T-033 only. Implement read-only `explain-test-owner` and `task-guide ROLE=module-author OWNER=...` from the same catalog/plan topology with exact human/JSON contracts and no retained artifacts.
- Rollback boundary: revert the complete T-032 checkpoint, including hidden pre-cutover Make entrypoints, command dispatch, selection/planning/execution facade, source snapshot, schemas/attachments, evidence-target routing, boundary ownership, tests, and tracker evidence. Never replace it with a translated phase plan or target-wide over-selection.

## 17. First-resumer checklist

1. Read this tracker, `docs/testing-harness-nlspec.md`, `docs/domain.md`, Core 00 through Core 04, and owner NLSpecs implicated by the first selected rows.
2. Confirm branch, commit, worktree, and any changes since the recorded baseline.
3. Mark only one workstream `IN_PROGRESS` and assign exact task IDs.
4. Run `make help`, `make help-all`, and applicable `make task-guide`/explain commands to inspect the live public surface before changing it.
5. Complete WS-00 inventories and the 548-row crosswalk before owner-contract or implementation edits.
6. Never hand-edit generated roots or generated topology outputs; update owners and use Make generation.
7. Keep public authority singular. Do not merge a dual catalog, fallback reader, alias target, or partial hard cutover.
8. Update this tracker and append a handoff entry before yielding the work.
