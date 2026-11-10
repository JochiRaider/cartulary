# Test Consolidation Refactor Tracker and Handoff

Version: 1.8.0  
Date: 2026-11-10  
Target: Cartulary testing ecosystem  
Authoring status: recovered-panic correction complete — scoped GO  
Artifact mode: corrective handoff; S-051 and final S-012/S-014 govern readiness

## 1. Scope and Source Posture

### 1.1 Authority, purpose, and authorization

**TCR-REQ-001 — Scoped authority.** This handoff MUST be the authoritative
change plan for the test-suite rationalization effort described here. Its
requirements govern consolidation decisions, preservation obligations, evidence,
sequencing, and completion. Adopted subsystem NLSpecs and normative Core sections
MUST continue to govern production behavior under
[Core 00 precedence](../spec/00_document_set_status_and_precedence.md).
The [Testing Harness NLSpec](../testing-harness-nlspec.md) MUST govern harness
semantics. This handoff MUST NOT override either authority, declare a new product
requirement, or promote an implementation observation into a requirement.

The [NLSpec grounding document](../research/nlspec-spec.md) supplies writing
discipline. The [modular refactor framework](cartulary_modular_refactor_planning_framework.md)
supplies planning doctrine. Neither establishes that a test, defect, or runtime
result exists. The [domain document](../domain.md) supplies vocabulary and owner
navigation; it does not establish exact test selectors or internal filenames.

**TCR-REQ-002 — Current authorization.** The maintainer's instruction to
"Implement the proposed plan" authorizes the remediation workstreams in §6,
including owner amendments, authored implementation and test changes, canonical
generation, targeted validation, and this tracker's checkpoint updates. The
tracker-only restriction applied to the initial authoring session and remains
historical evidence; it does not restrict this implementation session. Repository
procedure and generated-artifact safeguards continue to apply. Another session's
services and unrelated work remain outside scope.

**TCR-REQ-003 — Prose isolation.** Tests, generators, runtime metadata,
conformance, and release evidence MUST NOT read, stat, hash, parse, or depend on
this handoff, other Cartulary specification/adoption prose, `docs/`, or
README. The mappings below are human review records. They MUST NOT become a new
executable requirement registry. Existing canonical artifacts and machine-readable
expectations remain executable inputs. Offline source producers retain only the
pinned-upstream exception already allowed by repository procedure.

### 1.2 Target, scope, and evidence vocabulary

The supplied repository root is `/home/jochi/code/cartulary`; the canonical
repository-relative target is `.`. The target kind is a cross-repository
testing seam, not a new production module. The label is `test-consolidation`.
The controlling artifact is `docs/handoffs/test-consolidation-refactor-tracker.md`.
Its ancestors are ordinary directories; it is a regular, singly linked file.
The original absent-output admission remains in the historical baseline.

Scope includes all discoverable test responsibilities: product, architecture,
security, harness infrastructure, tooling, documentation tooling, producer,
development-only, generated projections, and vendored tests. Production code is
changed only where required by the adopted remediation and tested boundaries.
Unrelated product redesign, feature implementation, data migration,
unsupported platform expansion, and unrelated cleanup are outside this effort.

**TCR-REQ-004 — Evidence semantics.** Each assertion in this tracker MUST be
identified by its role:

| Role | Meaning | Permitted conclusion |
| --- | --- | --- |
| Normative requirement | A TCR requirement or an explicitly referenced adopted owner clause. | Determines required behavior within its authority. |
| Confirmed static observation | Supported by inspected source at the recorded baseline. | Establishes source structure or an assertion's actual scope; does not establish a runtime pass. |
| Inference / potential concern | Plausible consequence not established by runtime or complete semantic evidence. | Opens bounded investigation; does not permit removal. |
| Unresolved evidence | A named missing fact, artifact, owner decision, or characterization. | Blocks the dependent decision. |
| Rationale | Explanation of a decision. | Supports understanding; adds no requirement. |

Work status is exactly `TODO`, `IN_PROGRESS`, `BLOCKED`,
`DONE`, `DEFERRED`, or `DROPPED`. Finding classification is
exactly `must_fix`, `should_fix`, `defer`, or
`intentional/no_action`. Planning completeness, readiness, and authorization
are separate outcome fields. A planning item marked DONE MUST NOT imply that
its future implementation is complete.

Missing duration, resource, retry, or execution evidence defaults to
`unmeasured` or `not executed`, never zero, stable, or passed.
Missing authority defaults to `unresolved`, never unnecessary. An
unreviewed test defaults to `retain pending evidence`.

### 1.3 Baseline and preservation

**TCR-REQ-005 — Baseline compatibility.** A later session MUST compare HEAD,
affected source content, governing documents, adopted status, callers, selectors,
and fixture profiles before reusing evidence. It MUST revalidate changed
dependencies or mark dependent evidence stale. Historical success MUST NOT be
reported as a current pass. Unrelated concurrent changes MUST NOT be reverted,
absorbed into a slice, or attributed to this effort without evidence.

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Repository / branch / commit | `/home/jochi/code/cartulary`; `main`; `ba75e83990783f50baa25b5c8ccf94484870e356`. | Original baseline changed independently; final HEAD matches the immediate pre-write baseline, E-024. | Original survey used `a43248fbe3f762b0f448993126f74172f597f101`; runtime claims are not imported. Concurrent HEAD advanced to `40c4d787e8cc2a417b6903a5b03e79547c872313`; E-025 records affected authoring guidance. |
| Tracked and non-ignored untracked entries | 9,474 entries; no staged, unstaged, or untracked changes. | All 9,472 non-output entries in the immediate pre-write baseline match final comparison; only this tracker is untracked. E-024. | Hashing a file is identity accounting, not semantic inspection. |
| Content and file identity | SHA-256 of sorted path records containing mode, device, inode, size, link count, and content hash or link target: `1d8fefed774a02d66cbe40e42be57c9fe168e9992a41394b21f5b083b46d0dcd`. | Original snapshot changed independently; immediate pre-write content/identity digest matches final comparison, E-024. | No capture errors. Snapshot calculation was in memory; no manifest was written. |
| Git index | SHA-256 `fc5b571fa9f810e07523c3d8a64ebae4e5f06406af38c6fd1144affd4b4f7dd5`. | Original index changed independently; immediate pre-write index matches final comparison, E-024. | Git observations use `--no-optional-locks`. |
| Ignored output roots | Immediate-directory identity and child-name digest observed for `.cache/cartulary/static-analysis` and `.cartulary/test-results`; `.cache/cartulary/build` absent. | Selected root metadata matches the immediate pre-write baseline; E-024. | Existing descendant artifact contents were not exhaustively fingerprinted. No newest-run selection or service observation occurred. |
| Output identity | Requested tracker absent; existing parent components are not symlinks. | Absence and parent identities rechecked before creation; E-024. | An intervening file MUST be read and reconciled; it MUST NOT be overwritten unseen. |

The audit counts and observations remain pinned to ba75e839. During drafting,
HEAD advanced independently to 40c4d787 and root/tooling/test-operations guidance
changed in the working tree. The user identified the Markdown lint simplification
as their work. Those changes are not part of this tracker change. Current
instructions and lint invocation were reread only to determine safe authoring and
future validation procedure; no unrelated changes were modified. A separate
pre-write comparison baseline is recorded in E-024: HEAD advanced again to
`beaf0293634e6aef9e62aa5670e033a26b6458d3`, the working tree was clean, and
9472 non-output entries were fingerprinted. E-025 identifies stale
invocation details; other unchanged candidate source observations remain pinned
to their inspected baseline. Future sessions MUST revalidate affected evidence.

## 2. Current-State Repository Inventory

### 2.1 Inventory contract and coverage boundary

**TCR-REQ-006 — Complete accounting.** Inventory closure MUST reconcile four
sets: source-defined tests and registrations; supporting files and fixtures;
canonical catalog selections; and public/non-public execution entrypoints.
Filename conventions alone MUST NOT define the universe. TestMain, benchmarks,
fuzz targets, parameterized cases, registration modules, shell entrypoints,
generator checks, and tests excluded from canonical discovery MUST receive an
explicit disposition.

The original planning baseline enumerated all tracked/non-ignored paths and all
`tools/test_families/*.json` rows. Named-file accounting below is
metadata-only except where E-IDs identify direct inspection. Family metadata
was inspected for all 66 owners. Detailed assertion inspection is limited to
the named candidate groups and counterparts in Sections 4, 5, and 10.
The original unresolved registrations are closed by S-001 and the current
disposition delta in §2.5. Complete source/catalog accounting is established;
exhaustive product requirement coverage and full-suite runtime acceptance are
separate and are not claimed.

Generated output, fixture data, ignored caches, installed dependencies, and
vendored code MUST remain visible as separate accounting classes. They MUST NOT
be counted as authored runnable tests merely because they contain test-like
names. Symbolic links and submodules MUST be recorded without following external
targets. Ignored installed dependencies and retained result trees are excluded
from authored-test enumeration; explicitly selected historical artifacts may be
inspected only through the owning observation procedure.

**TCR-REQ-007 — Responsibility classification.** Each reviewed test or coherent
test group MUST record its actual assertion purpose separately from catalog
owner, behavior class, runner, evidence class, aggregate placement, and fixture
capability. Product, architecture, security, harness, tooling, documentation,
producer, development-only, and vendor responsibilities MAY coexist within a
group; the group MUST be split for decision accounting when those obligations
would otherwise be lost. Supporting classification MUST NOT imply low value.

### 2.2 Source and mechanism inventory

The naming census below is a reproducible lower-level accounting view, not a
runnable-test count: Go basenames ending `_test.go`; TypeScript paths ending
`.test.ts` or `.test.tsx`; browser paths ending `.spec.ts` or
`.spec.tsx`; MJS basenames `test-*.mjs` or `*.test.mjs`; shell
basenames `test-*.sh` or `test_*.sh`; Python basenames
`test_*.py` or `*_test.py`. At this baseline these sets contain 2,083
distinct paths. Different earlier naming queries MUST NOT be interpreted as
suite growth or shrinkage.

| Path / kind | Count and accounting depth | Responsibility / public surface | Dependencies / execution mechanism | Owner / risk / evidence |
| --- | ---: | --- | --- | --- |
| `internal/modules/**/*_test.go` | 920; metadata-only, named candidates directly inspected. | Domain, architecture, service, security, transport, and helper assurance. | Go catalog package/test selectors; PostgreSQL, object store, or no services according to row capability. | Module owners; no uniform classification inferred from path. E-002–E-011. |
| `internal/platform/**/*_test.go` | 123; same depth rule. | Transport, storage adapters, shared security/runtime contracts. | Go; route/serialization units and service-backed fixtures. | Platform owners; discovery tests can be owned by app.server. E-002, E-010. |
| `internal/app/**/*_test.go` | 103; same depth rule. | Composition and process witnesses; projection consistency. | Go application harness or managed process. | Application/module owners. E-008, E-011. |
| `internal/testutil/**/*_test.go` | 114; metadata-only except named dependency inspection. | Fixture, lifecycle, service, and assertion-helper assurance. | Go; some checks deliberately exercise failure/cleanup behavior. | Harness or semantic owner; package location does not determine value or tier. E-002, E-019. |
| `referencepacks/**/*_test.go` | 41: attack 14, format 8, entra 7, windows 7, veris 5; metadata-only. | Offline producer and format assurance. | Go modules and Make-owned producer operations. | Producer/Reference Pack ownership requires per-family semantic review. E-002, E-020; RB-002. |
| `tools/**/*_test.go` | 20: contractgen 14, testservices 2, embedwebassets 1, objectstoreprobe 1, openapi-operation-catalog 1, reference-pack-modules 1; metadata-only. | Generation, fixtures, embedding, probes, module discovery. | Go modules and owning Make operations. | Harness/product boundary varies by assertion. E-002, E-020. |
| `apps/web/**/*.test.ts{x}` | 416; metadata-only, listed architecture candidates directly inspected. | Frontend product behavior, architecture, fixtures, browser-support utilities. | Vitest project selected by catalog/source convention. | Web/module/harness owners. E-002, E-009. |
| `packages/*/**/*.test.ts{x}` | 33: grid-adapter 10, report-renderer 7, ui-contracts 7, view-contracts 5, protocol-ts 3, test-utils 1; metadata-only with selector assertions inspected. | Vendor adaptation, rendering, selectors, contract adapters, choreography. | Vitest; browser-unit, Node, or renderer projects as selected. | Package and product owners. E-002, E-014. |
| `apps/web/e2e/**/*.spec.ts{x}` | 163; metadata-only, named candidates/counterparts directly inspected. | Stateful UI, API probes, accessibility, visuals, measurements, support. | Playwright/Chromium; browser stack; selectors expand independently of file count. | Module/web/harness owners. E-002, E-010–E-015. |
| Test-shaped MJS paths under `tools` | 116: harness 104, release-evidence 8, reference-pack-modules 2, reference-packs 2; metadata-only except named checker suites. | Node tests, suite registrars, and some production helpers with test-prefixed filenames. | Shell catalog operations can register multiple Node files and cases. | Tool/harness/release owners. E-016, E-020; registration closure pending. |
| Test-shaped shell paths under `tools` | 24: harness 23, release-evidence 1; metadata-only except wrapper inspection. | Shell regression scripts plus `tools/harness/output/test-output.sh`, an implementation helper. | Make-owned shell operations; subprocess/fixture behavior varies. | Harness/release owners. E-016, E-017. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_*.py` | 10; metadata-only, vendored/upstream development tests. | Upstream catalog/design tooling; outside canonical product discovery. | Python execution mechanism and dependency fitness pending inspection. | Vendor/development purpose; retain pending evidence, not product coverage. E-002, E-020. |
| Catalog, verification contracts, runner configs, source discovery, generated policies | Direct metadata inspection at named paths; not counted as test cases. | Selection, ownership, profiles, artifacts, execution closure. | Authored JSON projected into generated task/schedule artifacts. | Harness owners. E-003, E-018–E-020. |
| Other tracked paths | Accounted by baseline identity; non-test content outside detailed semantic review unless referenced below. | Production dependencies, docs, fixtures, generated/vendor assets. | No claim that all hidden registrations are excluded; RB-002 closes that question. | Excluded from test-shaped subtotal by stated naming rules, not from future discovery. |

Go totals are 1,321 files. A lexical scan finds 281 without a conventional
top-level `func Test…(` declaration. These are not automatically dead files:
helpers, TestMain, benchmarks, and fuzz declarations require separate inspection.
Vitest titles, parameterized expansions, Go subtests, and catalog rows MUST NOT
be added together as if they counted the same unit.

### 2.3 Canonical catalog baseline

| Dimension | Observed counts |
| --- | --- |
| Catalog rows / owners | 2,611 / 66 |
| Runner | Go 1,205; Vitest 753; Playwright 616; shell 37 |
| Evidence class | unit 1,311; integration 567; browser 484; static 100; accessibility 57; visual 42; measurement 34; security 11; release 5 |
| Aggregate placement | fast 1,184; standard 600; full 719; explicit_only 66; release 42 |
| Fixture capability | none 1,455; browser_stack 616; postgres_dedicated 393; postgres_transaction 125; postgres_migration 18; object_store_namespace 4 |
| Declared verification behavior | product-only 2,022; architecture-only 320; harness-only 149; security-only 103; product/security 8; architecture/product 9 |
| Claim posture | implementation 2,506; informative 105 |

The behavior totals derive from the set of behavior classes of each row's
referenced verification contracts. They are declared metadata, not a completed
semantic classification. The 66 files linked below contain exact family IDs,
row IDs, selectors, profiles, and service dependencies; this human table is a
baseline navigation index, not a second executable catalog.

| Owner / authored family catalog | Rows | Runner counts | Aggregate placement counts |
| --- | ---: | --- | --- |
| [app.migrate](../../tools/test_families/app.migrate.json) | 3 | go=3 | fast=1, standard=2 |
| [app.operator](../../tools/test_families/app.operator.json) | 13 | go=13 | fast=8, full=2, standard=3 |
| [app.server](../../tools/test_families/app.server.json) | 52 | go=51, playwright=1 | fast=16, full=6, standard=30 |
| [harness.browser](../../tools/test_families/harness.browser.json) | 97 | go=46, playwright=12, shell=13, vitest=26 | explicit_only=20, fast=2, full=72, standard=3 |
| [harness.command_surface](../../tools/test_families/harness.command_surface.json) | 14 | go=7, shell=7 | fast=5, full=9 |
| [harness.evidence_accounting](../../tools/test_families/harness.evidence_accounting.json) | 14 | go=12, shell=2 | explicit_only=1, fast=8, full=5 |
| [harness.generated_artifacts](../../tools/test_families/harness.generated_artifacts.json) | 12 | go=9, shell=3 | fast=9, standard=3 |
| [harness.release](../../tools/test_families/harness.release.json) | 5 | shell=5 | release=5 |
| [harness.test_catalog](../../tools/test_families/harness.test_catalog.json) | 1 | shell=1 | full=1 |
| [module.artifacts](../../tools/test_families/module.artifacts.json) | 19 | go=19 | fast=10, full=9 |
| [module.assessments](../../tools/test_families/module.assessments.json) | 27 | go=16, playwright=5, vitest=6 | fast=11, full=5, standard=11 |
| [module.auth](../../tools/test_families/module.auth.json) | 72 | go=45, playwright=18, vitest=9 | explicit_only=1, fast=29, full=22, standard=20 |
| [module.collaboration](../../tools/test_families/module.collaboration.json) | 50 | go=22, playwright=20, vitest=8 | explicit_only=4, fast=10, full=24, release=3, standard=9 |
| [module.crossownertransaction](../../tools/test_families/module.crossownertransaction.json) | 2 | go=2 | fast=1, standard=1 |
| [module.database_migrations](../../tools/test_families/module.database_migrations.json) | 19 | go=19 | fast=8, full=1, standard=10 |
| [module.entities](../../tools/test_families/module.entities.json) | 49 | go=31, playwright=15, vitest=3 | explicit_only=2, fast=12, full=23, standard=12 |
| [module.evidence](../../tools/test_families/module.evidence.json) | 100 | go=44, playwright=42, vitest=14 | explicit_only=4, fast=29, full=40, release=1, standard=26 |
| [module.extensions](../../tools/test_families/module.extensions.json) | 44 | go=42, playwright=2 | fast=22, full=2, standard=20 |
| [module.graphprojection](../../tools/test_families/module.graphprojection.json) | 11 | go=11 | fast=4, standard=7 |
| [module.imports](../../tools/test_families/module.imports.json) | 56 | go=34, playwright=10, vitest=12 | fast=23, full=6, release=4, standard=23 |
| [module.incidentbundles](../../tools/test_families/module.incidentbundles.json) | 38 | go=38 | fast=16, standard=22 |
| [module.incidents](../../tools/test_families/module.incidents.json) | 87 | go=50, playwright=28, vitest=9 | fast=27, full=34, standard=26 |
| [module.indicators](../../tools/test_families/module.indicators.json) | 40 | go=38, vitest=2 | fast=26, full=1, standard=13 |
| [module.jobapi](../../tools/test_families/module.jobapi.json) | 5 | go=5 | fast=1, standard=4 |
| [module.links](../../tools/test_families/module.links.json) | 14 | go=13, playwright=1 | fast=8, full=1, standard=5 |
| [module.networkflow](../../tools/test_families/module.networkflow.json) | 153 | go=120, playwright=31, vitest=2 | explicit_only=1, fast=69, full=35, release=1, standard=47 |
| [module.parties](../../tools/test_families/module.parties.json) | 19 | go=17, playwright=2 | fast=11, full=2, standard=6 |
| [module.projections](../../tools/test_families/module.projections.json) | 24 | go=24 | fast=2, full=1, standard=21 |
| [module.records](../../tools/test_families/module.records.json) | 9 | go=8, shell=1 | fast=2, standard=7 |
| [module.recovery](../../tools/test_families/module.recovery.json) | 50 | go=47, playwright=3 | fast=18, full=3, standard=29 |
| [module.reference_data](../../tools/test_families/module.reference_data.json) | 101 | go=89, playwright=6, vitest=6 | explicit_only=3, fast=54, full=3, release=3, standard=38 |
| [module.reportcomposition](../../tools/test_families/module.reportcomposition.json) | 4 | go=4 | fast=1, standard=3 |
| [module.reporting](../../tools/test_families/module.reporting.json) | 50 | go=36, playwright=2, vitest=12 | fast=21, release=2, standard=27 |
| [module.revisions](../../tools/test_families/module.revisions.json) | 79 | go=54, playwright=24, vitest=1 | fast=17, full=25, release=3, standard=34 |
| [module.savedviews](../../tools/test_families/module.savedviews.json) | 27 | go=14, playwright=12, vitest=1 | explicit_only=1, fast=6, full=11, standard=9 |
| [module.stagedobjects](../../tools/test_families/module.stagedobjects.json) | 3 | go=3 | fast=2, standard=1 |
| [module.tabularingest](../../tools/test_families/module.tabularingest.json) | 2 | go=2 | fast=1, full=1 |
| [module.tasksdecisions](../../tools/test_families/module.tasksdecisions.json) | 17 | go=13, playwright=3, vitest=1 | fast=8, full=3, standard=6 |
| [module.timeline](../../tools/test_families/module.timeline.json) | 203 | go=43, playwright=129, vitest=31 | explicit_only=3, fast=45, full=121, release=13, standard=21 |
| [module.workbook](../../tools/test_families/module.workbook.json) | 297 | go=57, playwright=214, vitest=26 | explicit_only=14, fast=30, full=214, standard=39 |
| [package.grid_adapter](../../tools/test_families/package.grid_adapter.json) | 47 | playwright=2, vitest=45 | explicit_only=1, fast=43, full=2, standard=1 |
| [package.protocol_ts](../../tools/test_families/package.protocol_ts.json) | 8 | shell=1, vitest=7 | fast=4, standard=4 |
| [package.test_utils](../../tools/test_families/package.test_utils.json) | 1 | vitest=1 | full=1 |
| [package.ui](../../tools/test_families/package.ui.json) | 11 | vitest=11 | fast=7, standard=4 |
| [package.view_contracts](../../tools/test_families/package.view_contracts.json) | 5 | vitest=5 | fast=5 |
| [platform.audit](../../tools/test_families/platform.audit.json) | 3 | go=3 | fast=1, full=2 |
| [platform.bootstrap](../../tools/test_families/platform.bootstrap.json) | 2 | go=2 | fast=2 |
| [platform.config](../../tools/test_families/platform.config.json) | 17 | go=17 | fast=16, full=1 |
| [platform.cryptography](../../tools/test_families/platform.cryptography.json) | 4 | go=4 | fast=4 |
| [platform.httpruntime](../../tools/test_families/platform.httpruntime.json) | 8 | go=8 | fast=6, full=1, standard=1 |
| [platform.jobs](../../tools/test_families/platform.jobs.json) | 19 | go=19 | fast=5, standard=14 |
| [platform.objectstore](../../tools/test_families/platform.objectstore.json) | 9 | go=9 | fast=5, full=3, standard=1 |
| [platform.openapi](../../tools/test_families/platform.openapi.json) | 6 | go=6 | fast=5, full=1 |
| [platform.pagination](../../tools/test_families/platform.pagination.json) | 2 | go=2 | fast=2 |
| [platform.postgres](../../tools/test_families/platform.postgres.json) | 6 | go=6 | fast=4, standard=2 |
| [platform.rootedfs](../../tools/test_families/platform.rootedfs.json) | 2 | go=2 | fast=2 |
| [platform.telemetry](../../tools/test_families/platform.telemetry.json) | 11 | go=11 | fast=4, release=7 |
| [platform.viewquery](../../tools/test_families/platform.viewquery.json) | 13 | go=8, playwright=3, vitest=2 | fast=10, full=3 |
| [platform.viewschema](../../tools/test_families/platform.viewschema.json) | 6 | go=6 | fast=5, standard=1 |
| [platform.ws](../../tools/test_families/platform.ws.json) | 1 | go=1 | full=1 |
| [web.application](../../tools/test_families/web.application.json) | 95 | vitest=95 | fast=79, standard=16 |
| [web.architecture](../../tools/test_families/web.architecture.json) | 11 | shell=3, vitest=8 | full=2, standard=9 |
| [web.collaboration](../../tools/test_families/web.collaboration.json) | 7 | vitest=7 | fast=7 |
| [web.design](../../tools/test_families/web.design.json) | 32 | playwright=31, shell=1 | explicit_only=11, fast=1, full=20 |
| [web.networkflow](../../tools/test_families/web.networkflow.json) | 66 | vitest=66 | fast=65, standard=1 |
| [web.workbook](../../tools/test_families/web.workbook.json) | 337 | vitest=337 | fast=329, standard=8 |

### 2.4 Review record and registration reconciliation

**TCR-REQ-008 — Review interface.** Each future reviewed group MUST have a
documented record containing the fields below. A shared record MAY cover cases
only when their obligation, scope, profile, and disposition are identical.

| Field | Required meaning / default |
| --- | --- |
| Identity | Source path, test identifier or registration, parameterized case boundaries, current catalog row IDs or explicit uncatalogued status. |
| Responsibility | Actual assertion purpose; primary owner and collaborators; declared classification separately. |
| Contract | Requirement/acceptance references where established; otherwise explicit engineering purpose and owner. Absence of a clause is not a removal reason. |
| Implementation | Production or supporting path and interface actually exercised; mock, fixture, HTTP, process, browser, or static scope. |
| Assertion | Preconditions, stimulus, observable result, failure oracle, and relevant negative/edge cases. |
| Selection | Public Make selector, runner/project, runtime and fixture profiles, aggregate placement, isolation, and service dependencies. |
| Evidence | Baseline, inspection depth, exact execution command/environment/result/artifact when available; otherwise not executed. |
| Cost and reliability | Setup/execution/cleanup, resource and contention observations, retries/failures, measurement limits; default unmeasured. |
| Disposition | retain, consolidate, relocate, correct, investigate, or retire; retain pending evidence is the default. |
| Traceability | Finding, contract, slice, validation, blockers, and completion evidence IDs. |

#### Execution registration reconciliation

At beaf0293634e6aef9e62aa5670e033a26b6458d3, source and registration inspection
accounts for 2,086 test-shaped or nonconventional registration files. This extends
the original 2,083-name census with three registration modules, not three newly
added tests. The canonical catalog remains 2,611 rows across 66 owners.
`make source-check CHECK=catalog` passed under run ID
`tcr-s001-catalog-20261109-01` (outer exit 0); its exact retained root is
`.cartulary/test-results/tcr-s001-catalog-20261109-01`. This validates current
catalog/source closure, not the correctness or completeness of product assertions.

All 1,321 Go test files and 612 frontend test/spec files have accountable catalog
package/file ownership. A separate declaration inspection found no unmatched
ordinary Go Test declaration or frontend file. The 282 Go files without an
ordinary Test declaration are retained package helpers/setup; filename count is
not a count of assertions. There are three TestMain entrypoints: serverprocess
owns service setup/cleanup; processtest owns bounded subprocess fixtures; ATT&CK
producer owns test-only CLI child dispatch. None is an independent catalog test.
The single FuzzTimeBucketArithmetic context remains development-only; ordinary
TestTimeBucketArithmeticPropertySeeds_Unit supplies selected deterministic
arithmetic coverage. No Benchmark declaration was found in this source posture.
Linux-constrained helper tests remain in the current Linux/WSL execution context;
this does not qualify another operating system.

Node/shell discovery separately inspected registration imports and exported
registration functions, including initialization-tests.mjs,
contract-initialization-cases.mjs and run-observation-cases.mjs. Every registration
candidate has an authored operation or registrar reference. Source references are
not execution passes. Test-output and catalog implementation helpers are not
miscounted as assertions. Ten vendored Python files retain upstream development
assurance and are deliberately outside canonical product discovery. They are not
executed, copied into product runners, or credited as product conformance.

The inventory below is registration-level inspection with explicit retention
purpose; it does not claim every assertion has received deep semantic review.
Named candidate assertions have the deeper inspection recorded in E-004–E-016.
S-011 and S-012 own semantic classification, obligation gaps and runtime evidence.
Fixture data and other supporting sources remain accounted by their authored
support/fixture inventories and the original whole-repository enumeration; they
are not independently runnable tests. No inventory entry is retired by this slice.

<details>
<summary>Per-file registration and retention ledger (S-001 baseline)</summary>

| Path | Accountable owner / context | Disposition and inspection purpose |
| --- | --- | --- |
| `apps/web/e2e/account.a11y.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/account.visual.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/administrative-audit.spec.ts` | module.auth | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/assessments.a11y.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/assessments.recovery.spec.ts` | module.assessments | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/auth-and-incident-directory.spec.ts` | module.auth, module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/auth.a11y.spec.ts` | module.auth | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/auth.support.spec.ts` | module.auth | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/auth.visual.spec.ts` | module.auth | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/authoring-candidates.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/autoresolve.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/batch-history-review.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/clock.spec.ts` | module.auth | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/collaboration.a11y.spec.ts` | module.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/collaboration.spec.ts` | module.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/collaboration.visual.spec.ts` | module.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/contextual-coordination-create.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/contextual-create.spec.ts` | module.assessments, module.entities, module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/coordination-public-route.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/coordination.a11y.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/coordination.authoring.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/coordination.visual.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/decisions.recovery.spec.ts` | module.tasksdecisions | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/deployment-administration.a11y.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/deployment-administration.visual.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/design.a11y.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/design.visual.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/enterprise-auth.spec.ts` | module.auth | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/entities.a11y.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/entities.visual.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/entity-find.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/entity-party-assignments.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence-integration.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence.a11y.spec.ts` | module.evidence, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence.access.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence.attachments.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence.file-gestures.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence.file-recovery.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence.lifecycle.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence.metadata.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence.verification.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/evidence.visual.spec.ts` | module.evidence, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/extensions.stateful.spec.ts` | module.extensions, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/fixture-state.spec.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/fixtures.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/grid-interaction.a11y.spec.ts` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/grid-interaction.visual.spec.ts` | module.timeline, package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/grid-provenance.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/history-browsing.spec.ts` | module.revisions | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/history-recovery.spec.ts` | module.revisions | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/history.spec.ts` | module.revisions | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/import-assistant.spec.ts` | module.imports | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-administration.a11y.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-administration.spec.ts` | module.incidents, module.savedviews, module.workbook, web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-administration.visual.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-creation.a11y.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-creation.visual.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-import.a11y.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-import.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-import.visual.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-lifecycle.spec.ts` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-membership-audit.spec.ts` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-membership-management.spec.ts` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident-metadata-editing.spec.ts` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/incident.support.spec.ts` | app.server, module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/indicator-canonical-create.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/indicator-lifecycle.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/indicator-observations.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/indicators.a11y.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/indicators.visual.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/inspector-actions.spec.ts` | module.revisions, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/keyboard.spec.ts` | module.timeline, module.workbook, web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/linked-note-create.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/anchorTraceSupport.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/evidence-inspector.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/evidence-verification.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/evidenceInspectorTimingSupport.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/historyTimingSupport.test.ts` | module.revisions | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/import-apply.spec.ts` | module.imports | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/import-discovery.spec.ts` | module.imports | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/jobProgressPaint.test.ts` | module.evidence, module.imports, module.reference_data, module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/network-flow-grid.spec.ts` | module.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/pasteTimingSupport.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/presenceTimingSupport.test.ts` | module.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/reference-pack-progress.spec.ts` | module.reference_data | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/reporting-progress.spec.ts` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/timeline-anchor-trace.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/timeline-grid.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/timeline-history.spec.ts` | module.revisions | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/timeline-paste.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/timeline-presence.spec.ts` | module.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/timeline-view-change.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/timingSupport.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/measurement/viewChangeTimingSupport.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/mention-lifecycle.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/mentions.discovery-recovery.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/mentions.lifecycle.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/mentions.recovery.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/mentions.resolve.spec.ts` | module.entities, module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/merge-recovery.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/merge.spec.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/network-analysis.a11y.spec.ts` | module.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/network-analysis.visual.spec.ts` | module.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/network-flow-navigation.spec.ts` | module.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/network-flow-pagination.spec.ts` | module.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/network-flow.spec.ts` | module.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/note-associations.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/notes.a11y.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/notes.authoring.spec.ts` | module.links | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/notes.visual.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/parties.authoring.spec.ts` | module.parties | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/parties.recovery.spec.ts` | module.parties | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/reference-candidate-support.spec.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/reference-multiple.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/reference-pack.spec.ts` | module.reference_data | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/reference-packs.a11y.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/reference-packs.visual.spec.ts` | web.design | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/restore.spec.ts` | module.recovery | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/saved-views.a11y.spec.ts` | module.savedviews, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/saved-views.visual.spec.ts` | module.savedviews, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/session-recovery.spec.ts` | module.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/accessibility/assertions.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/accessibility/contrast.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/accessibility/presentation.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/auth/sessions.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/auth/visualPreferences.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/collaboration/replay.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/entities/mentions.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/evidence/uploads.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/performance/ac043Snapshot.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/runtime/cleanup.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/runtime/fixtureIdentity.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/runtime/harnessState.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/runtime/ownerCharacterization.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/runtime/privateState.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/runtime/timingSupport.dom.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/timeline/timelineInvestigation.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/transport/incidentSocket.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/transport/publicJsonClient.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/transport/testControlClient.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/visual/profile.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/workbook/autosaveFeedback.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/workbook/mutationAnchors.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/support/workbook/savedViews.test.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/tasks-decisions.authoring.spec.ts` | module.tasksdecisions | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/tasks.lifecycle.spec.ts` | module.tasksdecisions | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-attachment-feedback.spec.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-auto-resolution-feedback.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-autosave-feedback.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-bulk-tag.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-collection-input.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-deferred-continuity.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-find.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-grid-entry.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-public-route.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-query.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-range-selection.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-related-evidence.spec.ts` | module.evidence, module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-scalar-clipboard.spec.ts` | module.timeline, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline-workbook.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline.a11y.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline.support.spec.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/timeline.visual.spec.ts` | module.timeline, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/visual-harness.spec.ts` | harness.browser | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbench.a11y.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-column-sizing.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-destination-handoff.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-editing.a11y.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-editing.visual.spec.ts` | module.timeline, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-grid-autosave.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-inspector-edit.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-inspector-layout.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-inspector.a11y.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-inspector.visual.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-ordinary-create.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-ordinary-security.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-preferences.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-query.authority.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-query.continuation.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-query.filters.spec.ts` | module.workbook, platform.viewquery | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-query.grouping.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-query.recovery.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-query.routes.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-query.saved.spec.ts` | module.savedviews, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-registry.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-save-status.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-saved-view-discovery.spec.ts` | module.savedviews | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-shell.a11y.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-shell.visual.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-standardized-surfaces.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook-surface-continuity.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook.assessments.spec.ts` | module.assessments | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook.generic.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook.spec.ts` | module.savedviews, module.workbook, platform.viewquery | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/e2e/workbook.support.spec.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/AccountApplicationMenu.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/App.auth.support.test.tsx` | module.auth | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/App.auth.test.tsx` | module.auth, web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/App.landing.test.tsx` | module.incidents, web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/App.timeline-invalidation.support.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/AppWorkflows.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/DeploymentAuditPanel.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/IncidentAdminPanel.test.tsx` | module.incidents, web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/IncidentImportPanel.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/IncidentMembershipAuditPanel.test.tsx` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/IncidentMembershipManagementPanel.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/ReferencePackAdminPanel.test.tsx` | module.reference_data | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/accountAuthenticationLifecycle.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/accountFrontendBoundaries.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/accountSettingsEditing.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/accountSettingsModel.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/administrativeAuditController.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/administrativeAuditModel.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/api/administrativeAuditClient.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/api/incidentDirectoryClient.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/api/incidentImportClient.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/api/incidentLifecycleClient.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/api/incidentMembershipAuditClient.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/api/incidentMembershipManagementClient.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/api/incidentMetadataClient.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/api/shellHttpClients.routeBoundary.test.ts` | module.auth | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/appDepartureReview.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/appSessionController.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/createReportingWorkflow.test.tsx` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/deploymentUsersModel.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/fontBundle.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/fontRoles.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentCreationModel.node.test.ts` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentDirectoryModel.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentImportHardening.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentImportModel.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentImportState.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentLifecycleCharacterization.test.tsx` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentLifecycleController.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentLifecycleIntegration.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMembershipAuditCharacterization.test.tsx` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMembershipAuditController.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMembershipAuditLifecycle.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMembershipAuditModel.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMembershipManagementCharacterization.test.tsx` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMembershipManagementController.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMembershipManagementLifecycle.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMetadataCharacterization.test.tsx` | module.incidents | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMetadataController.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/incidentMetadataLifecycle.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/otelBoundary.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/referencePackAdminController.node.test.ts` | module.reference_data | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/referencePackAdminModel.node.test.ts` | module.reference_data | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/routeState.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/useNetworkFlowImport.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/useWorkbookImport.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/app/workbookPreferenceIntegration.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/collaboration/IncidentCollaborationSession.test.tsx` | module.collaboration, web.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/collaboration/incidentCollaborationSessionPlan.node.test.ts` | web.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/extensions/extensionAvailability.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/imports/CommittedSourcesPanel.test.tsx` | module.imports | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/imports/WorkbookImportController.node.test.ts` | module.imports | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/imports/importCoordinator.test.ts` | module.imports, web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/imports/importJobObservation.node.test.ts` | module.imports | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx` | module.networkflow, web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/NetworkFlowControls.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/NetworkFlowImportController.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/NetworkFlowQueryControls.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/NetworkFlowQueryPagination.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/NetworkFlowSemanticGrid.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/NetworkFlowTableController.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/SavedGraphController.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/SavedGraphRecovery.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/SavedGraphResultNavigation.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/explorationContinuity.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/explorationFocus.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowAliasFixture.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowBoundaryPolicy.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowClient.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowCollaborationInterpreter.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowController.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowErrors.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowExplorationNavigation.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowImportAdmission.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowImportModel.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowIndicatorLinkModel.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowIndicatorLinkOperation.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowPageRecovery.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowPresentation.node.test.ts` | module.networkflow, web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowQueryAuthoring.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/networkFlowQueryModel.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/savedGraphObservation.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/savedGraphTransport.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/tableConsumerContinuity.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/useNetworkFlowExtensionEvents.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/useNetworkFlowGridLayout.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/useNetworkFlowPagedQuery.test.tsx` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/networkFlow/useNetworkFlowQueryAuthoring.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/reporting/ReportingController.node.test.ts` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/services/browserApi.test.ts` | module.auth, web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/services/clientTransactionId.node.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/services/importClient.test.ts` | module.imports | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/services/importTargetContractAdapter.node.test.ts` | module.imports | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/services/networkFlowContractAdapter.node.test.ts` | web.networkflow | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/services/referencePacks.test.ts` | module.reference_data | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/services/workbookEvidence.test.ts` | module.evidence, web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/shared/publicError.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/shared/publicErrorPresentation.node.test.ts` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/shared/useRegisteredOverlayNavigation.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/shared/useTransientMessageController.test.tsx` | web.application | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/shared/workbookRecoveryNavigation.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/testing/irFixtureSupport.node.test.ts` | web.architecture | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/testing/networkFlowFixtureSupport.node.test.ts` | web.architecture | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/testing/sourceOwnershipPolicy.node.test.ts` | web.architecture | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/testing/timelineWorkbookTestSupport.test.tsx` | web.architecture | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/testing/transportBoundaryPolicy.node.test.ts` | web.architecture | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/testing/workbookInspectorTestSupport.test.tsx` | web.architecture | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.actionSequencing.test.tsx` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.assessments.test.tsx` | module.assessments, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.autosave.test.tsx` | module.timeline, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.collaboration.test.tsx` | module.collaboration, module.workbook, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.evidence.test.tsx` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.grid.test.tsx` | module.timeline, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.gridProvenance.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.history.test.tsx` | module.collaboration, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.inspector.test.tsx` | module.workbook, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.mentionChips.node.test.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.payload.test.tsx` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.query.test.tsx` | platform.viewquery, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.saveState.test.tsx` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.sentinel.test.tsx` | module.workbook, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.support.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx` | module.evidence, module.workbook, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/WorkbookShell.timelineQuery.test.tsx` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createIndicatorCreateTransport.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createIndicatorLifecycleAdapter.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createObservationTransport.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createOrdinaryCreateTransport.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookClipboardPasteAdapter.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookDecisionSupersessionAdapter.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookEntityMergeAdapter.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookEvidenceMetadataAdapter.node.test.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookPendingMutationAdapter.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookPreferenceAdapter.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookRecordHistoryAdapter.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookSavedViewAdapter.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookStartupAndIncidentAdapters.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/createWorkbookViewQueryAdapter.node.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/observationContracts.node.test.ts` | module.indicators | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/adapters/workbookOperationErrorPolicy.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.node.test.ts` | web.collaboration, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/collaboration/workbookCollaborationMessages.node.test.ts` | module.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/collaboration/workbookCollaborationTransitionModels.node.test.ts` | web.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/commands/WorkbookCommands.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/commands/workbookCommandIndex.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/GenericMutationControl.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookActiveSurfaceFrame.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookAuthoringReferencePicker.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookBatchRecordChoices.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookBatchRecovery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookGridControls.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookIncidentIdentityDisclosure.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookParkedGridDrafts.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookReferenceControl.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookRelationshipChip.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/components/WorkbookShellTopBar.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/composition/WorkbookMutationPresentation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/continuity/gridViewportContinuity.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/continuity/workbookContinuityPort.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/evidence/evidenceAccessPresentation.node.test.ts` | web.application, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/ImportAssistantFeature.test.tsx` | module.imports | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/assessments/assessmentDiscovery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/assessments/useAssessmentCreationController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/DecisionSupersessionEditor.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/WorkbookDecisionSupersessionOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/contextualCreateAuthoring.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/contextualCreateDiscovery.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/contextualCreateRecovery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/coordinationCreateAuthoring.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/coordinationCreateRecovery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/decisionSupersession.characterization.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/decisionSupersessionReconciliation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/decisionSupersessionRuntime.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/coordination/useCoordinationWorkflowController.test.tsx` | module.tasksdecisions | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/entities/WorkbookEntityMergeOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/entities/WorkbookEntityMergeRecovery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/entities/entityInspectorEditing.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/entities/useEntityMergeController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/evidence/EvidenceAttachmentEntry.test.tsx` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/evidence/evidenceFileRecovery.test.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/evidence/resolveSolePreviewableEvidence.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/evidence/timelineRelatedEvidenceAuthoring.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/evidence/timelineRelatedEvidenceRecovery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/evidence/useEvidenceWorkbookBindings.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/generic/GenericInspectorReferenceSummary.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/IndicatorCreateFromObservation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/IndicatorInspectorWorkflow.test.tsx` | module.indicators | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/IndicatorLifecycleWorkflow.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/ObservationCollection.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/WorkbookIndicatorCreateOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/WorkbookIndicatorLifecycleOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/WorkbookObservationOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/indicatorCreateAuthoring.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/indicatorCreateReconciliation.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/indicatorLifecycle.characterization.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/indicatorLifecycleModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/indicatorLifecycleReconciliation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/indicatorLifecycleRuntime.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/indicatorObservations.characterization.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/observationAuthoring.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/observationModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/observationPresentation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/indicators/observationReconciliation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/notes/noteAssociations.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/notes/noteCreateAuthoring.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/notes/noteCreateRecovery.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/ordinary/WorkbookOrdinaryCreateOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/ordinary/ordinaryCreateControls.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/ordinary/ordinaryCreateSubmission.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/parties/WorkbookPartyLinkOperationOwner.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/features/parties/partyLinkControls.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/find/WorkbookFindControl.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/find/WorkbookFindController.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/find/useWorkbookFind.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/history/HistoryActionLookup.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/history/WorkbookBatchHistoryReview.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/history/WorkbookHistoryRecovery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/history/WorkbookRecordHistoryOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/history/workbookHistoryBrowsing.characterization.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/history/workbookHistoryBrowsing.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/hooks/ordinaryInspectorRecovery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/hooks/useWorkbookExtensionAvailability.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/hooks/useWorkbookProjectionRefreshController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/hooks/useWorkbookQueryController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/hooks/useWorkbookSavedViewController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/hooks/useWorkbookSemanticGridFocus.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/hooks/useWorkbookStartupController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/WorkbookInspectorDeclaredPanelList.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/WorkbookInspectorDraftStore.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/WorkbookInspectorRecordHistory.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/inspectorCapabilityResolver.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/inspectorRelatedRecordModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/presentation/WorkbookInspectorPanelContent.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/presentation/WorkbookInspectorPresentation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/useInspectorCreateRelatedWorkflow.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/useRetainedInspectorRow.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/useWorkbookInspectorCoordinator.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/useWorkbookInspectorEditDraft.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/workbookHistoryPresentationModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/workbookHistoryRecovery.characterization.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/workbookInspectorErrorModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/workbookInspectorFieldFeedback.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/inspector/workbookRecordHistoryModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/layout/WorkbookColumnLayoutController.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/layout/useWorkbookMenuPlacement.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/layout/useWorkbookResponsiveLayout.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/layout/workbookDensity.node.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/layout/workbookLayoutPolicy.node.test.ts` | web.architecture | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/layout/workbookResponsiveLayout.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/WorkbookGridDraftStore.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/assessmentWorkbookModel.node.test.ts` | module.assessments, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/entityCellPresentation.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/entityClipboardPastePlan.node.test.ts` | module.entities | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/entityMergePlan.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/entityWorkbookModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/evidenceLifecycleViewModel.node.test.ts` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/genericWorkbookModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookClipboardPaste.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookGridEditValue.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookGridQueryControls.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookInspectorModel.node.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookQuery.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookRequestDecoders.node.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookSavedViewControl.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookSavedViewPaginationMachine.node.test.ts` | module.savedviews | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookSavedViewRuntime.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookSavedViews.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookStartup.node.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookSurfaceRegistration.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/models/workbookSurfaceRegistry.node.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/navigation/WorkbookNavigationPresentation.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/navigation/WorkbookNavigationStatus.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/navigation/WorkbookSessionNavigation.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/navigation/WorkbookWorkPanel.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/navigation/useWorkbookWorkbench.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/policies/workbookApplicationShortcuts.node.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/policies/workbookSurfaceOwnershipPolicy.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/preferences/WorkbookPreferenceController.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/preferences/workbookPreferenceCharacterization.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/query/WorkbookQueryBrowser.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/query/WorkbookQueryBrowsingControls.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/query/WorkbookSurfaceComposition.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/query/ordinaryCreateQuery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/query/useAssessmentSurfaceQuery.test.tsx` | module.assessments, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/query/useEntitySurfaceQuery.test.tsx` | module.entities, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/query/useGenericSurfaceQuery.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/query/workbookLatestRequest.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/query/workbookQueryMetadata.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/WorkbookBatchOperationOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/WorkbookEntityMergeAdmission.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/WorkbookExplicitPatchOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/WorkbookGridAutosave.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/WorkbookMutationRuntime.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/WorkbookRuntimeResponsibilities.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/WorkbookSurfaceRegistry.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/pending/workbookPendingQueue.node.test.ts` | module.collaboration, module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/workbookBatchOutcome.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/runtime/workbookConflictModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/savedviews/WorkbookSavedViewController.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/savedviews/savedViewReads.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/services/WorkbookCandidateDiscovery.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/services/workbookReferenceSelection.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/startup/useWorkbookStartupAdmission.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/startup/workbookStartupAdmissionMachine.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/actions/TimelineSupersessionEditor.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/actions/WorkbookTimelineCaptureActionOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/actions/WorkbookTimelineMentionOperationOwner.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/actions/reconcileTimelineMentionReceipt.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/actions/timelineMentionCandidates.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/adapters/createTimelineActionAdapters.test.ts` | module.evidence, module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/adapters/createTimelineMentionSourceReader.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/adapters/timelineCaptureTransport.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/adapters/timelineMentionProtocol.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/bulk/useTimelineClearController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/bulk/useTimelineFillController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/collaboration/useTimelineCollaborationBindings.test.tsx` | web.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/collaboration/useTimelinePresenceController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/components/TimelineAttachmentFeedback.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/components/TimelineCollectionCell.test.tsx` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/components/TimelineEvidenceMetadata.test.tsx` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.test.tsx` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/components/TimelineInspectorDetails.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/components/TimelineScalarEditor.test.tsx` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookNotices.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/composition/useTimelineSourceWriteCoordination.test.tsx` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/editing/useTimelineEditorDraftRegistry.test.tsx` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/focus/timelineInspectorElementRegistry.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/hooks/useTimelineClipboardPasteController.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/hooks/useTimelineEvidenceMetadata.test.tsx` | module.evidence | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineBulkTagPlan.node.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineClipboardPastePlan.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineEvidenceAttachmentPlan.node.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineFindText.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineKeyboardIntentModel.node.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineLoadMachine.node.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineMentionActionPlan.node.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineModelBoundaries.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineMutationDriverPlans.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineMutationModels.node.test.ts` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineRowsModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineViewportContinuityModel.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/timelineWorkbookRuntime.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/models/workbookRecordFreshness.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/mutations/timelineFileDraftRecovery.node.test.ts` | module.evidence, web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/mutations/useTimelineRowMutationCoordinator.test.tsx` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/timelineAutosaveIsolation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/timelineBulkTagInteraction.test.tsx` | module.timeline | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/timelineCaptureActions.characterization.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/timelineCompositionArchitecture.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineCommittedRecordIdle.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineCompositionLifecycle.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineCreateRelatedWorkflow.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineFind.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineInspectorFeatureController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineInspectorLifecycle.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineKeyboardController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineMentionActions.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineMutationRuntimeBindings.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineObservationSource.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineRowActionMenu.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineRows.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineSurfaceFoundation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/timeline/useTimelineViewportContinuityController.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/utils/GridAdapter.anchor.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/utils/workbookClipboard.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/utils/workbookPresence.test.tsx` | web.collaboration | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/utils/workbookRowReconciliation.node.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/utils/workbookStatusSecondary.node.test.ts` | module.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/utils/workbookValueFormat.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/workbookInspectorOwnershipArchitecture.node.test.ts` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/workbookRecoveryNavigation.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `apps/web/src/workbook/workbookSaveStatus.test.tsx` | web.workbook | registered frontend assertions; retain except named slices; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_catalog_refresh.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_core.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_core_data_quality.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_data_contracts.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_design_system_mode.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_native_desktop_stack_freshness.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_relevance_evaluator.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_style_taxonomy.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_text_layout_resilience.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests/test_web_stack_freshness.py` | upstream development tooling | vendored Python assurance; exclude from product conformance, retain upstream; registration-level inspection. |
| `internal/app/collaborationassembly/incidenteffects/coordinator_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/app/configassembly/configuration_test.go` | platform.config | registered Go assertions; retain; registration-level inspection. |
| `internal/app/configassembly/platform_settings_test.go` | platform.config | registered Go assertions; retain; registration-level inspection. |
| `internal/app/extensionassembly/configuration_test.go` | app.server, module.crossownertransaction, module.extensions, module.stagedobjects | registered Go assertions; retain; registration-level inspection. |
| `internal/app/extensionassembly/cross_owner_transaction_integration_test.go` | app.server, module.crossownertransaction, module.extensions, module.stagedobjects | registered Go assertions; retain; registration-level inspection. |
| `internal/app/extensionassembly/cross_owner_transaction_test.go` | app.server, module.crossownertransaction, module.extensions, module.stagedobjects | registered Go assertions; retain; registration-level inspection. |
| `internal/app/extensionassembly/job_reconciliation_integration_test.go` | app.server, module.crossownertransaction, module.extensions, module.stagedobjects | registered Go assertions; retain; registration-level inspection. |
| `internal/app/extensionassembly/publication_catalog_test.go` | app.server, module.crossownertransaction, module.extensions, module.stagedobjects | registered Go assertions; retain; registration-level inspection. |
| `internal/app/extensionassembly/staged_objects_integration_test.go` | app.server, module.crossownertransaction, module.extensions, module.stagedobjects | registered Go assertions; retain; registration-level inspection. |
| `internal/app/extensionassembly/state_store_integration_test.go` | app.server, module.crossownertransaction, module.extensions, module.stagedobjects | registered Go assertions; retain; registration-level inspection. |
| `internal/app/graphassembly/rebuild_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/app/graphassembly/reverse_reconstruction_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/app/importassembly/assessment_integration_test.go` | module.assessments, module.imports, module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/app/importassembly/assessment_membership_integration_test.go` | module.assessments, module.imports, module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/app/importassembly/owner_registry_test.go` | module.assessments, module.imports, module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/app/importassembly/tasksdecisions_integration_test.go` | module.assessments, module.imports, module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/app/incidentportabilityassembly/catalog_test.go` | module.incidentbundles | registered Go assertions; retain; registration-level inspection. |
| `internal/app/incidentportabilityassembly/storage_stream_test.go` | module.incidentbundles | registered Go assertions; retain; registration-level inspection. |
| `internal/app/incidentportabilityassembly/storage_test.go` | module.incidentbundles | registered Go assertions; retain; registration-level inspection. |
| `internal/app/jobassembly/catalog_test.go` | platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/app/migrate/admission_integration_test.go` | app.migrate | registered Go assertions; retain; registration-level inspection. |
| `internal/app/migrate/fresh_state_test.go` | app.migrate | registered Go assertions; retain; registration-level inspection. |
| `internal/app/migrate/migrate_test.go` | app.migrate | registered Go assertions; retain; registration-level inspection. |
| `internal/app/migrate/reference_schema_integration_test.go` | app.migrate | registered Go assertions; retain; registration-level inspection. |
| `internal/app/operator/internal/recoverycli/cli_test.go` | app.operator | registered Go assertions; retain; registration-level inspection. |
| `internal/app/operator/operator_migration_evidence_integration_test.go` | app.operator, module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/app/operator/operator_migration_evidence_test.go` | app.operator, module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/app/operator/operator_reference_admission_integration_test.go` | app.operator, module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/app/operator/operator_reference_pack_test.go` | app.operator, module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/app/operator/operator_registry_test.go` | app.operator, module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/app/operator/operator_surface_test.go` | app.operator, module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/app/operator/operator_test.go` | app.operator, module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/app/operator/reference_lifetime_test.go` | app.operator, module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/app/projectionassembly/catalog_manifest_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/app/projectionassembly/source_ownership_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/app/recoveryassembly/evidence_repository_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/recoveryassembly/extension_inventory_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/recoveryassembly/release_identity_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/recoveryassembly/state_catalog_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/recoveryassembly/state_coverage_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/recoveryassembly/storage_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/recoveryassembly/structural_completion_evidence_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/recoveryassembly/target_admission_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/recoveryassembly/target_control_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/referenceassembly/administration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/app/referenceassembly/cleanup_lifetime_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/app/referenceassembly/staging_bounds_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/app/referenceassembly/storage_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/app/referenceassembly/telemetry_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/app/reportingassembly/links_test.go` | module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/app/revisionassembly/history_projection_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/app/revisionassembly/revisions_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/cancellation_grace_process_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/collaboration_publication_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/environment_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/evidence_cleanup_telemetry_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/evidence_composition_characterization_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/extensions_publication_characterization_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/graph_activation_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/module_settings_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/network_flow_telemetry_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/openapi_contract_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/publication_deadline_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/publication_loss_process_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/recovery_target_admission_integration_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/reference_admission_integration_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/runtime_dependencies_test.go` | app.server, module.stagedobjects, platform.openapi | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/app/server/runtime_integration_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/runtime_lease_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/runtime_routes_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/runtime_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/server_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/server/staged_objects_test.go` | app.server, module.stagedobjects, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/authentication_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/authentication_support_test.go` | app.server, module.extensions, module.networkflow, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/app/serverprocess/bootstrap_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/config_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/embedded_frontend_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/evidence_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/extension_dispatch_integration_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/extensions_lease_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/incident_membership_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/incident_support_test.go` | app.server, module.extensions, module.networkflow, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/app/serverprocess/networkflow_fault_recovery_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/app/serverprocess/networkflow_runtime_routes_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/process_fixture_harness_test.go` | app.server, module.extensions, module.networkflow, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/app/serverprocess/process_fixture_test.go` | app.server, module.extensions, module.networkflow, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/app/serverprocess/recovery_sentinel_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/recovery_target_control_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/restored_networkflow_claim_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/runtime_routes_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/serverprocess/tls_process_test.go` | app.server, module.extensions, module.networkflow, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/app/tasksdecisionassembly/link_facts_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/artifacts_idempotency_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/assessment_idempotency_migration_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/assessment_idempotency_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/entity_admission_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/evidence_capabilities_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/indicator_adapter_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/party_capabilities_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/taskdecision_admission_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/tasksdecisions_idempotency_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/app/workbookassembly/timeline_capabilities_test.go` | module.artifacts, module.assessments, module.evidence, module.indicators, module.parties, module.tasksdecisions, module.timeline, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/artifact_collection_integration_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/artifact_conflict_integration_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/artifact_contract_support_test.go` | module.artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/artifacts/artifact_import_integration_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/artifact_patch_envelope_integration_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/artifact_rollback_integration_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/artifact_source_mutation_integration_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/coordination_source_integration_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/internal/providers/rollback/provider_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/internal/sourcecatalog/catalog_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/linked_notes_integration_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/mutation_admission_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/mutation_composition_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/note_associations_integration_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/note_associations_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/records_capability_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/risk_refs_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/source_policy_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/source_state_manifest_test.go` | module.artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/artifacts/surface_contract_test.go` | module.artifacts | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/artifacts/workbookprojection/contribution_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/admission/create_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/assembly_constructor_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/assessment_contract_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/assessments_integration_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/constructor_contract_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/create_identity_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/facade_contract_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/import_create_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/internal/providers/rollback/provider_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/merge_effects_constructor_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/policy_contract_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/projection_provider_contribution_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/assessments/workbookprojection/contribution_test.go` | module.assessments | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/admin_audit_contract_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/admin_audit_integration_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/authorization_integration_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/csrf_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/enterprise_auth_contract_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/enterprise_auth_integration_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/handlers_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/integration_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/openapi_contract_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/password_capacity_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/request_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/routes_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/saml_production_integration_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/secret_envelopes_integration_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/store_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/support_integration_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/support_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/testsupport/flowtest/actor_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/testsupport/internal/actorfixture/actor_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/auth/testsupport/storetest/rollback_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/codegenboundary/sqlc_boundary_test.go` | harness.generated_artifacts, module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/codegenboundary/testutil_import_boundary_test.go` | harness.generated_artifacts, module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/candidate_discovery_integration_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/codec_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/hub_state_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/hub_telemetry_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/hub_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/integration_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/intent_validation_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/internal/stream/dispatcher_lifecycle_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/protocol/replay_validation_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/publication_catalog_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/record_change_intent_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/runtime_construction_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/socket_session_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/socket_transport_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/stream_integration_entrypoint_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/collaboration/stream_integration_support_test.go` | module.collaboration | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/collaboration/stream_intent_integration_test.go` | module.collaboration | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/collaboration/stream_recovery_integration_test.go` | module.collaboration | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/collaboration/stream_retention_integration_test.go` | module.collaboration | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/collaboration/stream_tailing_integration_test.go` | module.collaboration | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/collaboration/telemetry_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/crossownertransaction/contract_admission_test.go` | module.crossownertransaction | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/crossownertransaction/coordinator_test.go` | module.crossownertransaction | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/crossownertransaction/limits_test.go` | module.crossownertransaction | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/crossownertransaction/validation_test.go` | module.crossownertransaction | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/database_migrations/baseline_integrity_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/catalog_capture_test.go` | module.database_migrations | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/database_migrations/catalog_characterization_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/crypto_admission_integration_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/graph_lease_retirement_integration_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/import_source_index_integration_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/incident_membership_baseline_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/incident_portability_cutover_integration_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/migration_contract_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/migration_lock_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/migration_remediation_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/migration_state_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/migrationevidence/evidence_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/migrations_context_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/migrations_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/public_contract_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/rebaseline_admission_integration_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/rebaseline_manifest_integration_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/rebaseline_policy_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/recovery_target_state_integration_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/reporting_failure_records_integration_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/schema_bootstrap_guard_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/schema_bootstrap_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/schema_readiness_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/database_migrations/sourcecatalog/provider_test.go` | module.database_migrations | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/active_identifier_claims_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/boundary_guard_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/candidate_discovery_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/create_idempotency_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/create_security_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/entity_alias_migration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/exact_match_unit_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/hostidentity/clipboard_paste_api_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/hostidentity/field_registry_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/hostidentity/party_assignment_admission_test.go` | module.entities | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/entities/hostidentity/patch_api_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/hostidentity/rollbackprovider/collections_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/hostidentity/rollbackprovider/host_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/hostidentity/rollbackprovider/identity_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/hostidentity/store_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/hostidentity/workbook_admission_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/incident_bundle_invariant_closure_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/mention_effects_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/mention_unit_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/mentions/ports_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/mentions/rollbackprovider/mention_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/merge/merge_protected_set_composition_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/merge/merge_protected_set_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/merge_route_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/merge_unit_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/mutation_admission_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/openapi_contract_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/origin_upsert_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/party_assignment_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/projectioncontract/contribution_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/record_schema_helpers_test.go` | module.entities | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/entities/record_schema_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/relationship_contracts_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/resolution_route_integration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/resolution_support_test.go` | module.entities | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/entities/revision_provider_contribution_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/source_integrity_migration_test.go` | module.entities | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/entities/unit_support_test.go` | module.entities | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/evidence/access_handle_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/access_handle_lifecycle_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/attach_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/attachment_authorization_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/blob_create_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/blob_idempotency_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/blob_quarantine_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/blob_verification_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/blobref/blobref_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/cleanup_dispatcher_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/cleanup_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/cleanup_objectstore_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/conformance_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/create_validation_contract_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/evidence_blob_uniqueness_migration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/handle_state_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/handles_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/internal/policy/policy_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/internal/providers/rollback/provider_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/lifecycle_boundary_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/lifecycle_projection_collaboration_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/lifecycle_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/lifecycle_test_support_test.go` | module.evidence | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/evidence/metadata_query_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/mutation_admission_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/mutation_coordinator_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/objectstore_dependency_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/party_reference_admission_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/projection_collaboration_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/projectioncontract/contribution_test.go` | module.evidence, module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/projectioncontract/row_test.go` | module.evidence, module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/provider_contract_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/quarantine_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/request_decoding_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/routes_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/service_composition_test.go` | module.evidence | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/evidence/surface_contracts_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/upload_lifecycle_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/upload_routes_integration_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/evidence/upload_token_test.go` | module.evidence | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/admission_validation_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/availability_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/binding_admission_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/catalog_isolation_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/collision_collection_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/contract_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/coordinator_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/dependency_execution_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/empty_registry_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/extension_job_cutover_migration_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/inactive_configuration_catalog_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/initialization_diagnostics_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/job_policy_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/jobs_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/owner_diagnostics_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/owner_identity_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/owner_input_limits_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/protocol_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/recoverycontribution/inventory_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/registry_contracts_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/registry_integrity_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/retired_participant_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/startup_diagnostics_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/state_capabilities_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/state_service_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/extensions/state_test_adapter_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/state_validation_regression_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/extensions/validation_vectors_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/admission_limits_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/aggregation_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/boundary_guard_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/contributor_order_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/cutover_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/edge_semantics_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/engine_v2_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/execution_bounds_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/field_paths_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/graph_projection_migration_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/input_counts_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/postgresrestore/writer_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/postgresresult/cleanup_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/postgresresult/contract_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/postgresresult/historical_numeric_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/postgresresult/limits_export_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/postgresresult/result_proof_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/postgresresult/store_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/restore/boundary_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/restore/contract_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/restore/service_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/result_claims_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/result_v2_validation_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/result_vectors_v2_test.go` | module.graphprojection | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/graphprojection/semantics_v2_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/graphprojection/v2_contract_projection_test.go` | module.graphprojection | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/apply_cancellation_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/behavior_contract_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/committed_source_portability_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/committed_source_transfer_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/current_workflow_faults_test.go` | module.imports, module.incidentbundles, module.timeline | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/imports/current_workflow_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/generic_projection_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/http_security_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/imports_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/ir_clipboard_public_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/ir_corpus_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/ir_investigation_lifecycle_test.go` | module.imports, module.incidentbundles, module.timeline | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/imports/ir_reporting_lifecycle_test.go` | module.imports, module.incidentbundles, module.timeline | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/imports/large_incident_transfer_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/mapping_compatibility_contract_test.go` | module.imports, module.incidentbundles, module.timeline | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/imports/mapping_source_protection_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/module_composition_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/network_flow_export_snapshot_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/network_flow_harness_fault_test.go` | module.imports, module.incidentbundles, module.timeline | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/imports/network_flow_history_transfer_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/network_flow_transfer_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/owner_errors_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/ownerfacade/compatibility_contract_test.go` | module.imports | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/ownerfacade/registry_test.go` | module.imports | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/ownerfacade/scalar_contract_test.go` | module.imports | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/source_admission_integration_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/source_admission_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/source_reader_integration_test.go` | module.imports, module.incidentbundles, module.timeline | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/imports/workbook_text_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/imports/xlsx_test.go` | module.imports, module.incidentbundles, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/actor_export_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/api_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/archive_stream_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/archive_test_support_test.go` | module.incidentbundles, module.parties | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidentbundles/extension_portability_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/import_fatal_completion_test.go` | module.incidentbundles, module.parties | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidentbundles/import_input_retention_test.go` | module.incidentbundles, module.parties | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidentbundles/incident_bundle_storage_reference_migration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/incidents_source_characterization_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/portable_commit_fault_test.go` | module.incidentbundles, module.parties | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidentbundles/portable_fatal_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/reference_attack_round_trip_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/reference_entra_round_trip_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/reference_export_test.go` | module.incidentbundles, module.parties | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidentbundles/reference_veris_round_trip_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/reference_windows_round_trip_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/routes_admission_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/routes_atomicity_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/routes_compatibility_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/routes_helpers_integration_test.go` | module.incidentbundles, module.parties | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidentbundles/routes_party_portability_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/routes_reference_verification_integration_test.go` | module.incidentbundles, module.parties | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidentbundles/routes_reporting_artifacts_integration_test.go` | module.incidentbundles, module.parties | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidentbundles/routes_round_trip_integration_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/source_catalog_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentbundles/source_failure_test.go` | module.incidentbundles, module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentportability/member_reads_test.go` | module.incidentbundles | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidentportability/strict_ndjson_test.go` | module.incidentbundles | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/admission/admission_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/application_helpers_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/audit_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/authorization_helpers_test.go` | module.incidents | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidents/authorization_incident_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/authorization_membership_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/authorization_observations_test.go` | module.incidents | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/incidents/authorization_timeline_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/authorization_workbook_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/extra_integration_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/http_conformance_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/httpapi/lifecycle_contract_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/httpapi/location_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/httpapi/membership_audit_contract_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/httpapi/membership_management_contract_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/httpapi/metadata_contract_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/httpapi/openapi_contract_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/httpapi/request_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/httpapi/support_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/httpapi/unit_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/idempotency_compatibility_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/incident_bundle_source_codec_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/integration_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/internal/sourcestate/manifest_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/lifecycle_integration_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/membership_audit_integration_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/membership_concurrency_integration_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/membership_version_history_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/metadata_intent_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/mutation_admission_lock_integration_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/mutation_admission_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/pagination_integration_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/receipt_concurrency_integration_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/source_state_contract_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/store_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/testsupport/storetest/store_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/incidents/transaction_participant_integration_test.go` | module.incidents | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/active_identity_claims_integration_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/admission/create_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/application_actor_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/application_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/application_test_helpers_test.go` | module.indicators | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/indicators/boundary_guard_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/child_routes_integration_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/composition_boundary_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/envelope_contract_migration_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/httpapi/construction_test.go` | module.indicators | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/indicators/httpapi/error_mapping_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/httpapi/vocabulary_admission_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/idempotency_hash_test.go` | module.indicators | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/indicators/identity_concurrency_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/identity_portability_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/indicators_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/internal/identity/identity_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/internal/providers/rollback/children_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/internal/providers/rollback/provider_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/internal/providers/rollback/vocabulary_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/internal/sourcestate/catalog_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/internal/vocabulary/vocabulary_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/lifecycle_integrity_migration_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/observation_origin_integration_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/observation_origin_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/portability_characterization_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/portability_final_state_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/portability_invariants_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/production_contract_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/projection_provider_contribution_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/provider_encapsulation_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/replay_codec_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/resolution_integration_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/revision_provider_contribution_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/source_state_contract_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/target_resolution_integration_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/transaction_atomicity_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/unit_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/vocabulary_contract_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/indicators/workbookprojection/contribution_test.go` | module.indicators | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/jobapi/error_test.go` | module.jobapi | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/jobapi/routes_integration_test.go` | module.jobapi | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/internal/mutationvalue/mutation_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/internal/sourcestate/manifest_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/internal/sourcestate/prepare_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/internal/valuecodec/valuecodec_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/internal/vocabulary/vocabulary_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/item_refs_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/merge_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/party_assignment_facts_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/portability_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/revision_provider_contribution_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/route_projection_history_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/source_state_contributions_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/source_state_portability_integration_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/store_test.go` | module.links | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/links/test_helpers_test.go` | module.links | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/application_composition_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/approved_mapping_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/approved_mapping_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/canonical_endpoint_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/canonical_migration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/canonical_success_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/canonical_upgrade_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/cisco_sna_example_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/cisco_sna_policy_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/cisco_sna_profile_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/configuration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/csv_decoder_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/direction_evidence_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/endpoint_direction_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/extension_state_v3_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_audit_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_comparison_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_contributors_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_graph_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_graph_probe_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/fixture_graph_retention_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_large_limits_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_link_dependency_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/fixture_linking_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_mapping_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_names_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_privacy_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_privacy_logs_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/fixture_publication_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_query_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_resource_limits_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_resource_probe_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/fixture_retention_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/fixture_source_change_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_capacity_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_cleanup_deadline_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/graph_materialization_payload_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/graph_materialization_timeout_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_query_application_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_restore_failure_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_result_cleanup_dispatcher_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_result_cleanup_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_result_cleanup_race_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_result_cleanup_test_bridge_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/graph_streaming_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_telemetry_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_temporal_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_view_authority_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_view_authority_test_bridge_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/graph_view_receipts_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/graph_view_store_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/harness_audit_integration_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/harness_authorization_integration_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/harness_randomness_integration_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/harness_worker_integration_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/harnesscontrol/control_json_test.go` | module.networkflow | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/harnesscontrol/network_flow_audit_assertion_integration_test.go` | module.networkflow | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/harnesscontrol/network_flow_auth_transition_integration_test.go` | module.networkflow | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/harnesscontrol/network_flow_fault_integration_test.go` | module.networkflow | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/harnesscontrol/network_flow_randomness_integration_test.go` | module.networkflow | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/harnesscontrol/test_helpers_test.go` | module.networkflow | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/import_preview_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/import_source_inspection_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/import_validation_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/indicator_link_boundary_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/indicator_link_fixture_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/indicator_link_replay_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/integration_surface_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/keyring_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/mapping_admission_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/mapping_admission_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/mapping_policy_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/network_flow_behavior_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/network_flow_contract_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/network_flow_unit_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/normalized_row_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/operation_admission_integration_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/orientation_corpus_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/orientation_corpus_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/orientation_diagnostics_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/portability_prepare_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/production_surface_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/query_authoring_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/query_sql_test_bridge_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/reporting_frozen_restart_integration_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/reporting_graph_budget_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/reporting_graph_source_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/retained_import_result_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/routes_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/saved_graph_admission_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/saved_graph_read_application_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/selected_source_history_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/semantic_boundary_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/semantic_reference_boundary_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/store_fixture_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/store_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/table_lifecycle_integration_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/table_lifecycle_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/table_receipt_validation_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/textcanon_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/timestamp_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/transaction_contracts_test.go` | module.networkflow, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/networkflow/transaction_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/v3_contract_projection_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/networkflow/validation_vectors_test.go` | module.networkflow, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/active_key_claims_integration_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/clock_cutover_integration_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/create_shared_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/internal/policy/value_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/internal/providers/incidentbundle/portability_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/internal/providers/rollback/provider_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/mutation_admission_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/mutation_composition_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/mutation_hash_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/parties_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/party_refs_test.go` | module.parties | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/parties/workbookprojection/contribution_test.go` | module.parties, module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/adapters/adapters_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/adapters/boundary_guard_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/queryengine/artifact_test.go` | module.artifacts, module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/queryengine/engine_test.go` | module.artifacts, module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/queryengine/host_test.go` | module.artifacts, module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/queryengine/identity_test.go` | module.artifacts, module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/catalog_construction_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/deletion_characterization_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/maintenance_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/provider_registry_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/query_plans_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/query_store_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/query_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/rebuild_failure_matrix_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/rebuild_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/telemetry_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/timeline_query_paging_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/runtime/timeline_query_schema_guard_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/internal/storage/artifact_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/projections/providercontract/descriptor_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/records/destructive_lock_integration_test.go` | module.records | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/records/envelope_store_integration_test.go` | module.records | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/records/incident_bundle_source_port_test.go` | module.records | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/records/reportingprovider/provider_integration_test.go` | module.records | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/records/route_target_integration_test.go` | module.records | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/records/subtypepresence/catalog_test.go` | module.records | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/application/due_batch_test.go` | module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/application/runtime_lifetime_test.go` | module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/application/target_admission_test.go` | module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/application/types_test.go` | module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/backup_capture_fixture_test.go` | module.extensions, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/recovery/backup_metadata_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/backup_metadata_store_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/binding_validation_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/current_metadata_capture_fixture_test.go` | module.extensions, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/recovery/current_restore_fixture_test.go` | module.extensions, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/recovery/current_restore_support_test.go` | module.extensions, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/recovery/encryption_policy_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/evidence_inventory_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/extension_catalog_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/extension_recovery_contract_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/extensions_internal_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/generation_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/object_store_backup_artifacts_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/operatortest/operator_process_test.go` | app.operator, module.recovery, platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/operatortest/portable_restore_process_test.go` | app.operator, module.recovery, platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/operatortest/target_control_process_test.go` | app.operator, module.recovery, platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/postgres_integrity_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/recovery_sentinel_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/reference_attack_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/reference_entra_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/reference_veris_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/reference_windows_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/reset_metadata_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/restore_owner_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/restore_projection_contract_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/restore_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/restore_verification_artifact_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/schema_projection_test.go` | module.extensions, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/recovery/streaming_encryption_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/structural_restore_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/target_capture_admission_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/target_restore_admission_integration_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/target_state_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/testsupport/recovery_test.go` | module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/transfer_test.go` | module.extensions, module.recovery | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/recovery/verification_internal_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/vnext_codec_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/vnext_graph_restore_artifacts_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/recovery/workbook_probe_test.go` | module.extensions, module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/api_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/attack_native_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/base_atomicity_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/base_reconciliation_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/canonical_persistence_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/concurrent_publication_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/consumer_contract_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/consumer_integrity_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/consumer_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/contractadmission/binding_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/contractadmission/content_profile_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/contractadmission/pending_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/contractadmission/strict_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/coordinator_admission_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/coordinator_execution_scope_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/coordinator_jobs_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/entra_lifecycle_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/generic_consumer_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/generic_consumer_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/generic_endpoint_verification_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/generic_recovery_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/http_admission_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/http_security_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/incident_export_integration_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/incident_portability_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/incident_retention_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/admission_fixture_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/archive_limit_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/archive_ratio_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/archive_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/attempt_result_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/attestation_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/builtin_binding_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/content_admission_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/content_diagnostics_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/content_limit_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/content_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/dataset_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/dependencies_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/dependency_fixture_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/diagnostic_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/diagnostics_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/fixture_contract_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/fixture_projection_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/fixture_scratch_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/framework_lookup_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/indicator_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/inventory_diagnostics_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/limit_diagnostics_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/manifest_limit_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/manifest_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/portability_retention_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/portable_content_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/portable_context_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/prepared_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/trust_fixture_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/trust_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/tuf_limit_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/tuf_root_shape_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packformat/tuf_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/internal/packstate/state_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/job_terminal_effects_integration_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/lifecycle_fixture_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/mixed_refresh_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/object_collection_integration_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/object_collection_storage_integration_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/object_collector_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/openapi_contract_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/operator_import_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/portable_admission_integration_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/portable_admission_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/portable_cleanup_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/portable_execution_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/portable_input_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/prepared_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/profile_consumers_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/provenance_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/public_types_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/reference_pack_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/reference_pack_storage_schema_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/reference_pack_test_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/refresh_admission_matrix_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/retired_format_fixture_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/staging_rejection_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/storage_port_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/successful_envelope_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/testsupport/entra_expectations_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/upload_limit_integration_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reference_data/verification_engine_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/verification_lifetime_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/verification_precedence_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/veris_lifecycle_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/windows_online_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/windows_release_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/worker_isolation_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/workspace_lifetime_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reference_data/workspace_recovery_integration_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reportcomposition/canonical_digest_test.go` | module.reportcomposition | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reportcomposition/local_validation_test.go` | module.reportcomposition | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reportcomposition/preview_options_test.go` | module.reportcomposition | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reportcomposition/release_tuple_facade_test.go` | module.reportcomposition | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reportcomposition/render_source_test.go` | module.reportcomposition | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reportcomposition/routes_integration_test.go` | module.reportcomposition | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/api_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/backup_recovery_integration_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/boundary_guard_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/bundle_members_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/canonical_json_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/composition_deck_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/composition_preview_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/composition_record_anchors_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/composition_render_input_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/composition_text_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/concurrent_admission_integration_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/container_redaction_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/content_classes_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/deleted_source_facts_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/diagram_direction_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/diagram_layout_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/disclosure_partitions_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/evidence_provider_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/evidence_provider_validation_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/execution_payload_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/export_materializer_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/export_model_paths_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/exportprovider/provider_test.go` | module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/failure_candidate_export_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/graph_projection_refs_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/graph_redaction_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/graph_result_render_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/graph_retention_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/mention_bindings_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/mermaid_binding_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/mermaid_source_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/model_fixture_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/model_identity_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/model_isolation_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/openapi_contract_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/output_options_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/portable_artifacts_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/portable_storage_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/preview_receipt_integration_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/preview_recovery_integration_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/record_lifecycle_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/redaction_action_parameters_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/redaction_profile_view_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/redaction_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/redaction_tokens_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/reference_binding_attack_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/reference_binding_entra_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/reference_binding_integration_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/reference_binding_veris_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/reference_binding_windows_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/reference_disclosure_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/release_binding_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/release_payload_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/release_recovery_integration_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/render_admission_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/render_attempt_deadline_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/render_failure_record_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/render_failure_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/render_fixture_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/render_issues_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/render_limits_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/render_model_admission_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/render_model_assembly_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/render_stage_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/render_supervision_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/render_validation_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/rendered_svg_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/reporting_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/reporting_job_worker_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/retained_fields_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/retained_model_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/retained_records_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/retained_sections_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/retained_structure_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/section_structure_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/sensitive_artifacts_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/sensitive_recovery_integration_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/slidev_binding_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/slidev_blocks_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/slidev_clicks_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/slidev_source_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/slidev_text_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/snapshot_admission_integration_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/snapshot_capture_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/snapshot_field_bindings_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/snapshot_payload_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/snapshot_records_test.go` | module.reportcomposition, module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/reporting/snapshot_validation_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/source_facts_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/source_groups_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/source_projection_integration_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/source_refs_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/structural_removals_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/subject_disclosure_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/subjects_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/support_refs_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/template_assets_integration_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/template_bindings_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/template_contract_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/template_derivation_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/template_pack_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/template_reachability_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/template_theme_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/timeline_diagram_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/timeline_sort_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/reporting/toolchain_binding_test.go` | module.reportcomposition, module.reporting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/revisions/application_boundary_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/catalog_admission_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/command_authorization_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/command_service_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/conflicts/conflict_token_internal_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/conflicts/conflict_token_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/conflicts/conflict_window_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/conflicts/consolidation_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/conflicts/field_resolver_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/conflicts/keep_saved_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/conflicts/text_merge_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/delete_restore_source_catalog_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/delete_restore_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/generic_dispatch_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/history_association_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/history_associations_migration_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/history_components_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/history_helpers_test.go` | module.revisions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/revisions/history_paging_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/history_source_paging_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/history_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/http_security_precedence_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/httpapi/history_pagination_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/imported_attribution_boundary_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/incident_bundle_allocator_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/incident_bundle_portability_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/indicator_children_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/integration_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/locks_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/projection_scope_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/revision_window_reader_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/rollback_changed_fields_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/rollback_components_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/rollback_failure_injection_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/rollback_participants_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/rollback_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/sourceboundary/resolver_integration_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/sourceboundary/resolver_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/target_semantics_nonrow_catalog_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/revisions/target_semantics_row_catalog_test.go` | module.revisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/savedviews/application_test.go` | module.savedviews | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/savedviews/export_test.go` | module.savedviews | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/savedviews/incident_bundle_portability_test.go` | module.savedviews | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/savedviews/read_contract_test.go` | module.savedviews | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/savedviews/read_query_test.go` | module.savedviews | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/savedviews/saved_views_storage_hardening_migration_test.go` | module.savedviews | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/savedviews/savedviews_test.go` | module.savedviews | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/stagedobjects/stagedobjects_test.go` | module.stagedobjects | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/stagedobjects/stream_test.go` | module.stagedobjects | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tabularingest/delimited_decoder_test.go` | module.tabularingest | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tabularingest/mapping_kernel_test.go` | module.tabularingest | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tabularingest/tabularingest_test.go` | module.tabularingest | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/decision_lifecycle_store_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/decision_reference_admission_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/decision_supersession_store_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/incident_bundle_source_port_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/internal/providers/rollback/provider_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/internal/sourcecatalog/catalog_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/mutation_admission_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/mutation_capabilities_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/mutation_composition_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/mutation_store_test_support_test.go` | module.tasksdecisions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/tasksdecisions/projectioncontract/contribution_test.go` | module.projections | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/reporting_contribution_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/tasksdecisions/task_mutation_store_test.go` | module.tasksdecisions | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/admission/assertions_test.go` | module.timeline | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/timeline/admission/batch_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/admission/clear_cells_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/admission/decoding_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/admission/error_classifier_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/admission/hashing_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/clear_cells_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/clipboard_paste_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/collectionfacts/reader_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/derivation_equivalence_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/import_create_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/ir_clipboard_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/mutationpolicy/policy_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/projection_contract_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/recovery_probe_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/replay_observations_test.go` | module.evidence, module.timeline | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/timeline/reportingprovider/sort_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/request_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/resolution_integration_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/rollbackprovider/provider_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/store_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/support_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/test_composition_test.go` | module.evidence, module.timeline | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/timeline/timecontract/timecontract_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/timeline_event_integration_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/unit_test.go` | module.evidence, module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/valuecodec/codec_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/versionid/format_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/timeline/workbookprojection/contribution_test.go` | module.timeline | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/viewschemas/viewschemas_integration_test.go` | platform.viewschema | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/catalog_artifact_surface_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/clear_cells_integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/clipboard_paste_integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/collection_conflict_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/conflict_route_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/contribution_catalog_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/coordination_surfaces_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/grouping_contract_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/locate_integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/locate_snapshot_integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/mutation_failure_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/mutation_observations_test.go` | module.assessments, module.tasksdecisions, module.workbook | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/workbook/mutation_replay_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/note_association_guards_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/note_association_query_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/note_associations_integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/notes_indicators_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/openapi_contract_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/optional_surfaces_catalog_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/owner_admission_contract_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/owner_setup_helpers_test.go` | module.assessments, module.tasksdecisions, module.workbook | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/modules/workbook/parties_integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/preferences_openapi_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/restoreprobe/registry_test.go` | module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/route_guard_characterization_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/routes_locate_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/row_wire_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/startup/api_test.go` | module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/startup/store_concurrency_internal_test.go` | module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/startup/store_test.go` | module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/startup_construction_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/surface_contracts_integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/tasksdecisions_conflict_contribution_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/telemetry_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/timestamp_contract_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/workbook_integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/workbook_mutation_integration_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/modules/workbook/workbook_startup_test.go` | module.assessments, module.tasksdecisions, module.workbook | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/administrativeaudit/audit_integration_test.go` | platform.audit | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/administrativeaudit/audit_test.go` | platform.audit | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/authn/idempotency_lock_test.go` | module.auth | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/authn/password_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/authn/primitives_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/authn/route_idempotency_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/authn/secrets_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/bootstrap/bootstrap_test.go` | platform.bootstrap | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/canonicaljson/canonicaljson_test.go` | module.extensions, module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/canonicaljson/extension_test.go` | module.extensions, module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/canonicaljson/strict_test.go` | module.extensions, module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/config/backup_config_test.go` | platform.config | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/config/catalog_test.go` | platform.config | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/config/claim_projection_test.go` | platform.config | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/config/core_config_test.go` | platform.config | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/config/extension_policy_test_support_test.go` | platform.config | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/config/inactive_extension_test.go` | platform.config | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/config/owner_namespace_test.go` | platform.config | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/config/validation_test_support_test.go` | platform.config | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/cryptography/execution_test.go` | platform.cryptography | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/cryptography/rsa_test.go` | platform.cryptography | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/cryptography/rsa_vectors_test.go` | platform.cryptography | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/cryptography/sealing_test.go` | platform.cryptography | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/cryptography/tls_admission_test.go` | platform.cryptography | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/cryptography/tls_test.go` | platform.cryptography | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/enterpriseauth/configuration_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/enterpriseauth/crypto_policy_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/enterpriseauth/manifest_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/enterpriseauth/production_verifier_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/enterpriseauth/saml_fixture_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/extensioncodec/codec_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/extensionstore/finalizer_integration_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/extensionstore/job_policy_integration_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/extensionstore/proof_bytes_test.go` | module.extensions | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/extensionstore/store_integration_test.go` | module.extensions | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/fieldnorm/text_test.go` | platform.viewquery | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/harnessredact/redaction_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/harnessruntime/reset_characterization_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/harnessruntime/reset_integration_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/harnessruntime/retired_controls_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/extension_dispatch_test.go` | app.server, platform.httpruntime, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/extensiondiscovery/routes_service_test.go` | app.server | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/extensiondiscovery/routes_test.go` | app.server | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/family_routes_test.go` | app.server, platform.httpruntime, platform.openapi | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/httpapi/httpapi_test.go` | app.server, platform.httpruntime, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/internal_error_test.go` | app.server, platform.httpruntime, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/operation_catalog_test.go` | app.server, platform.httpruntime, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/testclock_test.go` | app.server, platform.httpruntime, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/testroute_test.go` | app.server, platform.httpruntime, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/upload_envelope_test.go` | app.server, platform.httpruntime, platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpapi/webassets/webassets_test.go` | harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpruntime/publication_readiness_test.go` | platform.httpruntime | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/httpruntime/runtime_test.go` | platform.httpruntime | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/architecture_guard_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/composition_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/correction_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/execution_contract_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/expiry_internal_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/expiry_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/export_test.go` | module.extensions, platform.jobs | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/jobs/extension_identity_test.go` | module.extensions, platform.jobs | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/jobs/extension_policy_test.go` | module.extensions, platform.jobs | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/jobs/extensions_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/jobs_definition_migration_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/jobs_execution_migration_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/jobs_expiry_migration_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/jobs_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/resource_contract_test.go` | module.extensions, platform.jobs | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/jobs/restore_enumeration_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/retained_admission_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/runner_supervision_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/telemetry_integration_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/telemetry_test.go` | module.extensions, platform.jobs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/jobs/terminal_canonical_test.go` | module.extensions, platform.jobs | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/listquery/listquery_test.go` | platform.viewquery | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/objectstore/objectstore_adapter_contract_test.go` | platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/objectstore/objectstore_support_test.go` | platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/objectstore/objectstore_test.go` | platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/objectstore/s3_policy_test.go` | platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/objectstore/s3_streaming_integration_test.go` | platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/objectstore/telemetry_test.go` | platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/openapiassembly/assembly_test.go` | platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/openapicompat/compatibility_test.go` | platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/openapicompat/oauth_security_test.go` | platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/openapicompat/security_test.go` | platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/pagination/pagination_test.go` | platform.pagination | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/postgres/authentication_policy_test.go` | platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/postgres/postgres_admission_test.go` | platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/postgres/postgres_role_integration_test.go` | platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/postgres/postgres_settings_test.go` | platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/postgres/telemetry_test.go` | platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/postgres/transport_test.go` | platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/processlease/lease_test.go` | app.server | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/processlease/postgres_integration_test.go` | app.server | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/querypage/location_test.go` | platform.pagination, platform.viewquery | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/querypage/querypage_test.go` | platform.pagination, platform.viewquery | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/recoverystate/catalog_test.go` | module.recovery | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/renderprocess/installation_linux_test.go` | module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/renderprocess/process_linux_test.go` | module.reporting | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/rootedfs/directory_lease_linux_test.go` | platform.rootedfs | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/platform/rootedfs/inspect_linux_test.go` | platform.rootedfs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/rootedfs/production_boundary_test.go` | platform.rootedfs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/rootedfs/root_linux_test.go` | platform.rootedfs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/rootedfs/stage_linux_test.go` | platform.rootedfs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/securefile/read_linux_test.go` | platform.rootedfs | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/accessors_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/bootstrap_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/configuration_admission_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/exporter_identity_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/exporter_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/log_processor_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/logs_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/privacy_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/registry_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/resource_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/sampler_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/telemetry/transport_test.go` | platform.telemetry | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/viewquery/query_contract_test.go` | platform.viewquery | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/viewquery/query_test.go` | platform.viewquery | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/viewschema/layout_test.go` | platform.viewschema | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/viewschema/registry_test.go` | platform.viewschema | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/viewschema/viewschema_test.go` | platform.viewschema | registered Go assertions; retain; registration-level inspection. |
| `internal/platform/ws/ws_test.go` | platform.ws | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/appsupport/commit_fault_test.go` | app.server | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/appsupport/performancefixture/assembly_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/appsupport/performancefixture/evidence_fixture_integration_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/appsupport/performancefixture/evidence_inspector_integration_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/appsupport/runtime_guard_test.go` | app.server | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/appsupport/runtime_integration_test.go` | app.server | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/authcookietest/authcookietest_test.go` | module.auth | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/browserpki/pki_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/collaborationsupport/incidentwstest/client_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/collaborationsupport/incidentwstest/tls_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/collaborationsupport/protocoltest/protocoltest_test.go` | module.collaboration | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/configtest/configtest_test.go` | harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessaccount/cgroup_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessadmission/admission_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnesscollect/checkpoints_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnesscollect/collector_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnesscollect/container_observer_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessdiag/invocation_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessdiag/observation_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessdiag/receipt_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessdiag/rejection_test.go` | harness.evidence_accounting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessdiag/storage_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessdiag/token_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessfixtures/broker_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessfixtures/browser_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessfixtures/consumers_linux_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessfixtures/dependencies_linux_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessfixtures/observation_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessfixtures/proxy_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessfixtures/recover_command_linux_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessfixtures/recovery_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessfixtures/renderer_linux_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessfixtures/runtime_test_support_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessfixtures/startup_linux_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessfixtures/suite_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnesslocal/read_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessreview/adapter_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessreview/deadline_linux_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessreview/publication_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessreview/session_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessreview/transaction_linux_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessruntime/allocation_observation_linux_test.go` | harness.browser, harness.command_surface | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnessruntime/boundary/fixtures_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/boundary/graph_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/boundary/selected_services_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/boundary/session_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/browser_failure_linux_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/cache_linux_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/canonical_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/engine_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/nested_integration_linux_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/observation_linux_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/plan_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/prepare_linux_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/process_linux_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessruntime/producer_linux_test.go` | harness.browser, harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnessschema/schema_parity_generated_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnesssession/binding_linux_test.go` | harness.evidence_accounting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnesssession/cleanup_linux_test.go` | harness.evidence_accounting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnesssession/detail_linux_test.go` | harness.evidence_accounting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnesssession/diagnostic_linux_test.go` | harness.evidence_accounting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnesssession/lifecycle_linux_test.go` | harness.evidence_accounting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnesssession/preservation_linux_test.go` | harness.evidence_accounting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnesssession/protocol_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnesssession/recovery_locator_linux_test.go` | harness.evidence_accounting | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/harnesssession/runtime_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/harnesssession/scan_linux_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/httptestx/httptestx_test.go` | app.server, harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/httptestx/request_test.go` | app.server, harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/httptestx/response_test.go` | app.server, harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/irfixtures/corpus_test.go` | module.imports | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/networkflowfixtures/fixtures_test.go` | module.networkflow | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/objectstoreproxy/proxy_test.go` | platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/performancefixture/assembler_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/performancefixture/key_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/performancefixture/runtime_bundle_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/performancefixturelifecycle/capacity_linux_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/performancefixturelifecycle/lifecycle_integration_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/performancefixturelifecycle/lifecycle_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/performancefixturelifecycle/object_snapshot_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/performancefixturelifecycle/postgres_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/attachment_contract_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/database_cleanup_contract_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/database_contract_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/endpoint_diagnostic_test.go` | harness.browser, module.database_migrations, platform.postgres | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/pgtest/migration_contract_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/owned_cluster_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/pgtest_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/policy_contract_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/publication_diagnostic_test.go` | harness.browser, module.database_migrations, platform.postgres | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/pgtest/publication_qualification_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/template_contract_test.go` | harness.browser, module.database_migrations, platform.postgres | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/pgtest/template_preparation_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/pgtest/transaction_contract_test.go` | harness.browser, module.database_migrations, platform.postgres | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/postgrescleanup/attempt_observation_test.go` | harness.browser | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `internal/testutil/postgrescleanup/forced_drop_integration_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/postgrescleanup/forced_drop_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/processtest/processtest_test.go` | harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/recoverybrowsertest/drop_diagnostic_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/recoverybrowsertest/fixture_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/recoverybrowsertest/lifecycle_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/s3test/namespace_integration_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/s3test/namespace_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/s3test/ownership_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/s3test/readiness_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/sourcecheck/imports_test.go` | harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/suiteservices/activity_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/suiteservices/dependencies_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/suiteservices/diagnostics_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/suiteservices/env_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/testcontainersx/installed_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/testcontainersx/testcontainersx_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/testfailure/failure_test.go` | harness.evidence_accounting | registered Go assertions; retain; registration-level inspection. |
| `internal/testutil/wstest/wstest_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `packages/grid-adapter/src/clipboardCodec.test.ts` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/grid-adapter/src/columnSizing.test.tsx` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/grid-adapter/src/core.node.test.ts` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/grid-adapter/src/editorReveal.test.ts` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/grid-adapter/src/gridInteraction.test.ts` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/grid-adapter/src/index.test.tsx` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/grid-adapter/src/rangeKeyboard.node.test.ts` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/grid-adapter/src/semanticCellNavigation.node.test.ts` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/grid-adapter/src/semanticFocusRequest.test.ts` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/grid-adapter/src/semanticKernel.node.test.ts` | package.grid_adapter | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/protocol-ts/src/entrypoints/entrypoints.node.test.ts` | package.protocol_ts | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/protocol-ts/src/index.node.test.ts` | package.protocol_ts | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/protocol-ts/src/network-flow-mapping.node.test.ts` | package.protocol_ts | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/report-renderer/src/browser-policy.integration.test.ts` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/report-renderer/src/browser-policy.test.ts` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/report-renderer/src/mermaid.test.ts` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/report-renderer/src/module-assets.test.ts` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/report-renderer/src/pdf.test.ts` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/report-renderer/src/slidev.test.ts` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/report-renderer/src/stage.test.ts` | module.reporting | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/test-utils/src/index.test.ts` | package.test_utils | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/ui-contracts/src/application-selectors.node.test.ts` | package.ui | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/ui-contracts/src/design-presentation.node.test.ts` | package.ui | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/ui-contracts/src/design-tokens.node.test.ts` | package.ui | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/ui-contracts/src/network-flow-selectors.node.test.ts` | package.ui | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/ui-contracts/src/selector-core.node.test.ts` | package.ui | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/ui-contracts/src/workbook-interaction-selectors.node.test.ts` | package.ui | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/ui-contracts/src/workbook-shell-grid-selectors.node.test.ts` | package.ui | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/view-contracts/src/contracts.node.test.ts` | module.workbook, package.view_contracts | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/view-contracts/src/facade.node.test.ts` | package.view_contracts | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/view-contracts/src/references.node.test.ts` | package.view_contracts | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/view-contracts/src/rows.node.test.ts` | package.view_contracts | registered frontend assertions; retain except named slices; registration-level inspection. |
| `packages/view-contracts/src/surfaces.node.test.ts` | package.view_contracts | registered frontend assertions; retain except named slices; registration-level inspection. |
| `referencepacks/attack/attack_inputs_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/attack/audit_command_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/attack/audit_report_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/attack/command_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/attack/internal/native/endpoints_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/attack/internal/native/schema_boundaries_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/attack/internal/native/source_contract_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/attack/internal/native/source_envelope_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/attack/internal/native/source_limits_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/attack/internal/native/source_native_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/attack/internal/native/source_test_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/attack/schema_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/attack/source_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/attack/test_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/entra/command_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/entra/entra_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/entra/generic_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/entra/internal/native/source_envelope_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/entra/internal/native/source_native_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/entra/internal/native/source_test_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/entra/test_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/format/canonicaljson/canonicaljson_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/format/canonicaljson/strict_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/format/canonicaljson/tokens_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/format/dataset/dataset_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/format/dataset/limits_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/format/envelope/archive_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/format/envelope/tuf_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/format/publication/publication_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/veris/command_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/veris/internal/native/source_native_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/veris/internal/native/source_test_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/veris/test_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/veris/veris_source_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/windows/command_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/windows/internal/native/source_envelope_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/windows/internal/native/source_policy_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `referencepacks/windows/internal/native/source_test_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/windows/test_helpers_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/windows/windows_csv_test.go` | module.reference_data | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `referencepacks/windows/windows_online_test.go` | module.reference_data | registered Go assertions; retain; registration-level inspection. |
| `tools/contractgen/artifact_source_catalog_test.go` | harness.generated_artifacts | registered Go assertions; retain; registration-level inspection. |
| `tools/contractgen/collaboration_validation_test.go` | harness.generated_artifacts | registered Go assertions; retain; registration-level inspection. |
| `tools/contractgen/extensions_availability_test.go` | harness.generated_artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `tools/contractgen/extensions_collision_matrix_test.go` | harness.generated_artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `tools/contractgen/extensions_empty_test.go` | harness.generated_artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `tools/contractgen/extensions_identity_test.go` | harness.generated_artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `tools/contractgen/extensions_input_limits_test.go` | harness.generated_artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `tools/contractgen/extensions_owner_diagnostics_test.go` | harness.generated_artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `tools/contractgen/extensions_participant_test.go` | harness.generated_artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `tools/contractgen/extensions_transaction_test.go` | harness.generated_artifacts | Go fixture/helper or process setup; retain with package; registration-level inspection. |
| `tools/contractgen/import_targets_test.go` | harness.generated_artifacts | registered Go assertions; retain; registration-level inspection. |
| `tools/contractgen/recovery_codec_validation_test.go` | harness.generated_artifacts | registered Go assertions; retain; registration-level inspection. |
| `tools/contractgen/tasksdecisions_source_catalog_test.go` | harness.generated_artifacts | registered Go assertions; retain; registration-level inspection. |
| `tools/contractgen/validation_test.go` | harness.generated_artifacts | registered Go assertions; retain; registration-level inspection. |
| `tools/embedwebassets/main_test.go` | harness.generated_artifacts | registered Go assertions; retain; registration-level inspection. |
| `tools/harness/backend/tests/test-build-go-artifact.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/backend/tests/test-check-migrations.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/backend/tests/test-go-lpt-planning.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/backend/tests/test-go-row-runner.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/browser/tests/test-browser-acquisition.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-browser-lane-policy.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-browser-preparation.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-design-review.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/browser/tests/test-frontend-publication.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-measurement-diagnostics.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-measurement-evidence.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-measurement-fixtures.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-measurement-proofs.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-measurement-scheduler.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-playwright-row-runner.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-renderer-identity.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/browser/tests/test-run-playwright-step.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-ui-diagnostic.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-ui-review-analysis.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-ui-review-bundles.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-ui-review-contract.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-ui-review-execution.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-ui-review-lifecycle.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-ui-review-preparation.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-ui-review-seeded.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-ui-review-workflow.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-visual-contracts.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-visual-promotion.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/browser/tests/test-web-e2e-lifecycle.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/command-surface/tests/test-contracts.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/command-surface/tests/test-public-inputs.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/contract/test-output-context.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/contract/tests/test-artifact-reader.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/contract/tests/test-bootstrap-validator.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/contract/tests/test-cleanup.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/contract/tests/test-foundation-schema-validator.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/contract/tests/test-machine-state-config.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/diagnostics/tests/run-observation-cases.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/diagnostics/tests/test-canonical-target-plan.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/diagnostics/tests/test-plan-inputs.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/evidence-accounting/tests/test-canonical-unit-events.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/execution/tests/test-run-vitest-step.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/execution/tests/test-unit-selection.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/execution/tests/test-vitest-diagnostics.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/generated-artifacts/tests/test-boundary-projections.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/generated-artifacts/tests/test-execution-topology.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/generated-artifacts/tests/test-generate-drift.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/generated-artifacts/tests/test-generated-artifact-policy.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/generated-artifacts/tests/test-inspector-layout.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/generated-artifacts/tests/test-json-shapes.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/generated-artifacts/tests/test-task-surface-report.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/observability/tests/test-clock-placement.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/observability/tests/test-collection-engine.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/observability/tests/test-critical-path.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/observability/tests/test-execution-index.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/observability/tests/test-foundation-characterization.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/observability/tests/test-instrumentation.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/observability/tests/test-observability.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/observability/tests/test-owner-evidence-audit.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/observability/tests/test-wrapper-receipt.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/output/test-output.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/output/test-output.sh` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/output/tests/test-output-diagnostics.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/output/tests/test-shell-headlines.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/output/tests/test-step-process.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/output/tests/test-summary-accounting.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/performance-fixture/tests/test-evidence-generation.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/performance-fixture/tests/test-evidence-inspector-proof.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/performance-fixture/tests/test-registry-boundaries.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-acquisition-command.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-bootstrap-cache.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-bootstrap-node-runtime.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-bootstrap-shellcheck.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-build-input-discovery.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-cache-artifact.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-check-toolchain-pins.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-dev-services-lifecycle.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-dev-stack-lifecycle.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-frontend-artifact.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-frontend-install-self-hosting.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/readiness/tests/test-go-toolchain-readiness.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/runtime/tests/test-process-cleanup.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/runtime/tests/test-producer-umask.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/scheduler/work-graph/tests/test-browser-selection.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/scheduler/work-graph/tests/test-compilation.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/scheduler/work-graph/tests/test-composition.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/scheduler/work-graph/tests/test-contracts.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/scheduler/work-graph/tests/test-graph-contracts.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/scheduler/work-graph/tests/test-scheduler-contracts.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/scheduler/work-graph/tests/test-selection-policy.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/services/tests/test-local-session.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-backend-boundary-admission.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-backend-boundary-cli.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-backend-boundary-mechanics.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-backend-boundary-policy.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-exact-file-sets.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-fallow-failures.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-fallow-integration.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-fallow-policy.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-frontend-cli.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-frontend-policy.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-go-gosec-audit.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-go-gosec-targeted.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-go-govulncheck.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-go-import-inspection.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-go-staticcheck.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-harness-imports.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-lint-shell.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-markdown-biome.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-protocol-dead-code.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/static-analysis/tests/test-sql-relations.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/test-catalog/test-catalog.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/test-catalog/tests/test-contracts.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/contract-initialization-cases.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/initialization-tests.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/test-agent-finalize.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/tests/test-capacity-failure-contract.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/test-catalog-metadata-fixtures.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/test-command-failure.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/test-command-selections.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/test-frontend-producer-graph.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/test-frontend-producer-lifecycle.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/test-harness-boundary-contracts.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/tests/test-harness-command-surface-contracts.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/tests/test-harness-evidence-contracts.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/tests/test-harness-graph-contracts.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/tests/test-harness-scheduler-contracts.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/tests/test-make-node-tools.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/harness/tests/test-public-make-wrapper-smoke.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/harness/tests/test-tool-output-real-targets.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/objectstoreprobe/main_test.go` | platform.objectstore | registered Go assertions; retain; registration-level inspection. |
| `tools/openapi-operation-catalog/main_test.go` | platform.openapi | registered Go assertions; retain; registration-level inspection. |
| `tools/reference-pack-modules/closure.test.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/reference-pack-modules/isolation.test.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/reference-pack-modules/main_test.go` | harness.command_surface | registered Go assertions; retain; registration-level inspection. |
| `tools/reference-packs/explicit-input.test.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/reference-packs/source-audit.test.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/release-evidence/tests/test-benchmark-claim-check.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/release-evidence/tests/test-package-operation.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/release-evidence/tests/test-package-release.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/release-evidence/tests/test-package-smoke-cleanup.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/release-evidence/tests/test-release-readiness-evidence.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/release-evidence/tests/test-release-task-surface.sh` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/release-evidence/tests/test-sbom-license-evidence.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/release-evidence/tests/test-seaweedfs-release-evidence.mjs` | harness implementation/support | test-shaped helper/implementation; not an independent assertion; registration-level inspection. |
| `tools/release-evidence/tests/test-selected-evidence.mjs` | harness/release/producer operation owner | registered operation/registrar; preserve checker, lifecycle or producer assurance; registration-level inspection. |
| `tools/testservices/main_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |
| `tools/testservices/performance_fixture_test.go` | harness.browser | registered Go assertions; retain; registration-level inspection. |

</details>

### 2.5 S-011 portfolio disposition and current inventory

The implementation review closes accounting and disposition for the current
portfolio: **2,087 source/registration files, 2,609 catalog rows, 66 owners**.
The S-001 ledger remains a baseline; apply the following complete delta:

| File/context | Current disposition | Evidence depth and reason |
| --- | --- | --- |
| `internal/modules/codegenboundary/generated_contract_boundary_test.go` | Add; harness.generated_artifacts. | S-016 assertion/negative-fixture review and passing exact row; current-family authority. |
| `tools/harness/static-analysis/tests/test-projection-sql.mjs` | Add; harness.command_surface registrar. | S-006 assertion/negative-fixture review and passing boundary registrar; SQL-policy mechanics. |
| `apps/web/e2e/coordination-public-route.spec.ts` | Retired. | S-009 exact seven-field/authorization/effect transfer and qualified survivor; catalog migration records replacement. |
| Five rows in four `apps/web/src/testing/*TestSupport*` / fixture-admission files | Owner changes from web.architecture to harness.browser; files retained. | S-011 source inspection finds controlled input, identity/readiness, timeout/privacy and fixture-tamper oracles; full tier. Exact paths/rows are in the migration catalog. |
| Other baseline source/registration files | Retained with their owning group below and existing exact selectors. | Registration/family responsibility review; deeper assertion review only where named in the execution checkpoints. No hidden deletion or unsupported requirement-coverage assertion. |

Current runner counts: Go 1,205; Vitest 753; Playwright 613; shell 38.
Current placements: fast 1,172; standard 594; full 735; release 42;
explicit_only 66. These are rows, not test-case counts. The current source census
found no unaccounted test-shaped addition; canonical catalog validation checks
source registration identities independently of this prose.

The table below is the retention rationale and owner navigation for every family,
not a second executable catalog or a claim that each assertion proves a normative
requirement. The linked authored catalog supplies exact identities, cases,
profiles, fixtures, collaborators and selectors. The S-001 per-file ledger joins
these owners. Split review records for changed obligations are in Sections 5/10;
unchanged families retain independent engineering assurance against the failure
classes below. They are not carried forward merely to preserve a count.

| Accountable owner / exact catalog | Engineering purpose and retention rationale |
| --- | --- |
| [app.migrate](../../tools/test_families/app.migrate.json) | Migration command admission, upgrade execution and cleanup; retains the process/owner boundary. |
| [app.operator](../../tools/test_families/app.operator.json) | Operator command parsing, authorization and recovery dispatch; retains local command and process evidence. |
| [app.server](../../tools/test_families/app.server.json) | Runtime assembly, extension publication, incident/auth HTTP process flows and borrowed-resource ownership; fixture self-regressions are separately classified. |
| [harness.browser](../../tools/test_families/harness.browser.json) | Browser/Go fixtures, broker ownership, transport helpers, readiness, cleanup, diagnostics, measurement and visual evidence mechanics; retains failure/negative controls independently of product browser cases. |
| [harness.command_surface](../../tools/test_families/harness.command_surface.json) | Public Make admission, scheduler semantics, parser/CLI mechanics, toolchain/input and independent-producer boundaries; actual source gates and broad self-regressions have different placements. |
| [harness.evidence_accounting](../../tools/test_families/harness.evidence_accounting.json) | Event, receipt, schema, accounting, observation, collection and retained-evidence integrity; faulty/partial/unsafe inputs must fail rather than invent a pass. |
| [harness.generated_artifacts](../../tools/test_families/harness.generated_artifacts.json) | Typed projection/registry generation, artifact drift, migration input integrity and production imports/exports; current-source checks remain required gates. |
| [harness.release](../../tools/test_families/harness.release.json) | Package, selected evidence, license/SBOM, cleanup and release readiness; retains release-context evidence rather than relabeling it unit coverage. |
| [harness.test_catalog](../../tools/test_families/harness.test_catalog.json) | Registration identity, owner/profile/evidence closure and exact row selection; keeps invalid/overlapping/missing-selector regression fixtures. |
| [module.artifacts](../../tools/test_families/module.artifacts.json) | Notes, communication, handoff, review and lesson source mutations and contribution contracts; retains per-record validation and revision effects. |
| [module.assessments](../../tools/test_families/module.assessments.json) | Assessment authority, subject linkage, collaboration/recovery and workbook behavior; retains source, service and rendered distinctions. |
| [module.auth](../../tools/test_families/module.auth.json) | Identity/session/credential, CSRF, roles and administrative workflows; request/security behavior is distinct from screen observations. |
| [module.collaboration](../../tools/test_families/module.collaboration.json) | Replay/live event delivery, authorization recovery and client convergence; timing/product measurements and rendered accessibility remain distinct. |
| [module.crossownertransaction](../../tools/test_families/module.crossownertransaction.json) | Borrowed transaction commit/rollback across owners; keeps atomicity failures as integration obligations. |
| [module.database_migrations](../../tools/test_families/module.database_migrations.json) | Migration admission, order, transaction and schema lifecycle; separates actual migration contracts from fixture mechanics. |
| [module.entities](../../tools/test_families/module.entities.json) | Host/identity normalization, ownership, mentions, revisions and workbook behavior; retains private-source and public/rendered observations. |
| [module.evidence](../../tools/test_families/module.evidence.json) | Evidence metadata, blob custody, upload/download authorization, recovery and workbook interactions; service and browser effects are independent. |
| [module.extensions](../../tools/test_families/module.extensions.json) | Registry, claims, capability/route publication, startup and dispatch; current claim state is distinct from provider fixtures. |
| [module.graphprojection](../../tools/test_families/module.graphprojection.json) | Graph contribution/catalog ownership, ephemeral projection, lifecycle and exported surfaces; no dependency ban substitutes for export/shape assurance. |
| [module.imports](../../tools/test_families/module.imports.json) | Source admission, discovery/mapping, claim-gated targets, preview/apply and UI workflow; keeps invalid-input and durable effect checks. |
| [module.incidentbundles](../../tools/test_families/module.incidentbundles.json) | Portable bundle validation, import/export, state publication and participant boundaries; artifact safety and transaction failure remain required. |
| [module.incidents](../../tools/test_families/module.incidents.json) | Incident lifecycle, membership, authorization and landing/default preferences; assembled HTTP replaces only the identified API-only probes. |
| [module.indicators](../../tools/test_families/module.indicators.json) | Indicator normalization, matching/linking and source/projection authority; retains domain behavior and closed owner policy. |
| [module.jobapi](../../tools/test_families/module.jobapi.json) | Public job visibility, lifecycle and authorization; retains assembled route state/error checks. |
| [module.links](../../tools/test_families/module.links.json) | Link/tag eligibility, endpoint scope and atomic projection/revision updates; linked-refresh integration qualifies S-017. |
| [module.networkflow](../../tools/test_families/module.networkflow.json) | Claimed activity ingestion, temporal/row/graph behavior, limits, export and authorization; measurements/security and UI witnesses stay independently selected. |
| [module.parties](../../tools/test_families/module.parties.json) | Party fields, membership/mention semantics and recovery; retains public mutation and security observations. |
| [module.projections](../../tools/test_families/module.projections.json) | Provider source capabilities, table/catalog parity, deterministic queries and rebuild/refresh; SQL/static and runtime witnesses remain distinct. |
| [module.records](../../tools/test_families/module.records.json) | Envelope ownership, incident-scoped identity and transaction authority; record locks remain behind the owner capability. |
| [module.recovery](../../tools/test_families/module.recovery.json) | Backup/export/restore, quiescence, publication, trust and local authorization; seven command paths retain invocation-specific observations. |
| [module.reference_data](../../tools/test_families/module.reference_data.json) | Reference pack admission/trust/version state plus independent offline producers; product consumers use canonical artifacts, producers retain pinned-input checks. |
| [module.reportcomposition](../../tools/test_families/module.reportcomposition.json) | Report composition state, validation and revision contribution; retains owner-level mutation and integration assurance. |
| [module.reporting](../../tools/test_families/module.reporting.json) | Snapshot/render orchestration, live execution and output contracts; the import allowlist is architecture, not product-behavior evidence. |
| [module.revisions](../../tools/test_families/module.revisions.json) | History/conflicts/replay/revert, source capability ownership and current-role semantics; thin owner-local support-import selection is deliberately retained. |
| [module.savedviews](../../tools/test_families/module.savedviews.json) | Saved view ownership, versioned preferences and rendered selection; no UI observation is replaced by a request helper alone. |
| [module.stagedobjects](../../tools/test_families/module.stagedobjects.json) | Staged object admission, identity and durable lifecycle; retains failure/no-publication effects. |
| [module.tabularingest](../../tools/test_families/module.tabularingest.json) | Deterministic tabular parsing/normalization and malformed-input rejection; retains direct unit oracles. |
| [module.tasksdecisions](../../tools/test_families/module.tasksdecisions.json) | Task/decision state, relationships, supersession, revisions and workbook routes; seven-field HTTP transfer supplements source-owner tests. |
| [module.timeline](../../tools/test_families/module.timeline.json) | Timeline rows/collections, hierarchy, filters, continuity, keyboard and pending-edit behavior; browser support includes real interaction and remains meaningful. |
| [module.workbook](../../tools/test_families/module.workbook.json) | View/create/query/patch, conflict/history and browser editing contracts; source effects, API authorization and rendered focus are separate obligations. |
| [package.grid_adapter](../../tools/test_families/package.grid_adapter.json) | Vendor grid facade, selection/focus/geometry/editing and accessibility; preserves the actual adaptation boundary. |
| [package.protocol_ts](../../tools/test_families/package.protocol_ts.json) | Generated wire shape, decoding, bounds and closed facade access; generated-policy checks are distinct from consumer behavior. |
| [package.test_utils](../../tools/test_families/package.test_utils.json) | Shared browser choreography, passive observation and bounded identity/geometry waits; full-tier self-assurance remains directly selectable. |
| [package.ui](../../tools/test_families/package.ui.json) | Stable selector construction, token/theme and UI contracts; pure equality belongs here, mounted selectors stay in browser witnesses. |
| [package.view_contracts](../../tools/test_families/package.view_contracts.json) | View schema/field/control adapters and exported registry consistency; retains typed consumer interpretation. |
| [platform.audit](../../tools/test_families/platform.audit.json) | Audit attribution/content and integrity boundaries; retains persistence and validation oracles. |
| [platform.bootstrap](../../tools/test_families/platform.bootstrap.json) | Bootstrap admission and identity initialization; keeps invalid/incomplete inputs from silently creating authority. |
| [platform.config](../../tools/test_families/platform.config.json) | Configuration decoding, bounds, admission and owned roots; keeps malformed/unsafe-input tests separate from fixtures. |
| [platform.cryptography](../../tools/test_families/platform.cryptography.json) | Key/nonce, sealing and canonical cryptographic operations; direct known/invalid input behavior remains security-critical. |
| [platform.httpruntime](../../tools/test_families/platform.httpruntime.json) | HTTP errors, body/method/path handling and dispatch precedence; route-unit evidence complements assembled process witnesses. |
| [platform.jobs](../../tools/test_families/platform.jobs.json) | Lease/execution/job state and provider runtime contracts; service-backed ownership differs from shell harness lifecycle. |
| [platform.objectstore](../../tools/test_families/platform.objectstore.json) | Object storage adapters, capabilities and error/lifecycle behavior; keeps namespace/transport and product storage checks distinct. |
| [platform.openapi](../../tools/test_families/platform.openapi.json) | HTTP operation/schema projections and protocol identity; keeps machine contract consistency without reading specification prose. |
| [platform.pagination](../../tools/test_families/platform.pagination.json) | Cursor encoding/bounds/query admission; retains malformed and boundary-value oracles. |
| [platform.postgres](../../tools/test_families/platform.postgres.json) | Database binding/trust/admission and fixture contracts; ownership/profile distinctions determine placement, not the directory name. |
| [platform.rootedfs](../../tools/test_families/platform.rootedfs.json) | Rooted filesystem path confinement and operations; unsafe path rejection is a security engineering obligation. |
| [platform.telemetry](../../tools/test_families/platform.telemetry.json) | Telemetry privacy, bounds, configuration and export semantics; security checks remain required, unrelated exporter behavior is unchanged. |
| [platform.viewquery](../../tools/test_families/platform.viewquery.json) | Public query normalization, filtering/sorting/keyset and consumer behavior; preserves runtime/query boundary semantics. |
| [platform.viewschema](../../tools/test_families/platform.viewschema.json) | Typed view schema composition, field contracts and schema consumers; retains exact machine shape and integration behavior. |
| [platform.ws](../../tools/test_families/platform.ws.json) | WebSocket service behavior; its full-tier selected contract is retained independently of test-helper transport checks. |
| [web.application](../../tools/test_families/web.application.json) | Application routing, services, session/error recovery and client composition; static dependencies differ from user-visible behavior. |
| [web.architecture](../../tools/test_families/web.architecture.json) | Live source ownership, dependency direction, transport symbols and layout constraints; fixture/helper mechanics now have the harness owner. |
| [web.collaboration](../../tools/test_families/web.collaboration.json) | Client event/replay reconciliation and collaboration state; retains deterministic consumer and edge-case behavior. |
| [web.design](../../tools/test_families/web.design.json) | Design tokens, rendered accessibility and explicit-only visuals; no visual golden or accessibility witness is retired. |
| [web.networkflow](../../tools/test_families/web.networkflow.json) | Network Flow client ownership, mapping/import state, views and controls; claims/consumer behavior replace only source spelling. |
| [web.workbook](../../tools/test_families/web.workbook.json) | Workbook client reducers/composition, focus, mutation/recovery and owner-local controls; detailed type/symbol/behavior witnesses are retained. |

Node/shell registration and implementation material in the S-001 ledger has the
following accountable contexts; these replace the generic “operation owner”
labels for responsibility navigation without inventing independent test cases:

| Source scope | Accountable context / actual assurance |
| --- | --- |
| `tools/harness/backend`, `command-surface`, `execution`, `readiness`, `runtime`, `scheduler`, general `tests` | harness.command_surface; executable acquisition, CLI selection, scheduler/admission, failure/cleanup and producer mechanics. Exact operation registrars retain their case declarations. |
| `tools/harness/browser`, `performance-fixture`, `services` | harness.browser; runtime/fixture lifecycle, Playwright adapters, measurement/visual provenance and local-session behavior. UI-review/visual-specific operations remain explicit-only. |
| `tools/harness/contract`, `diagnostics`, `evidence-accounting`, `observability`, `output` and evidence registrars | harness.evidence_accounting with harness.command_surface composition; canonical shape, public output, safe artifact reading, receipt/event accounting and failure normalization. Output/context implementation files are not independent tests. |
| `tools/harness/generated-artifacts` | harness.generated_artifacts with harness.command_surface registrar; authored/generated parity, task topology, schema and projection integrity. |
| `tools/harness/test-catalog` | harness.test_catalog; catalog implementation and registration/selector/ownership regression. |
| `tools/harness/static-analysis` | harness.command_surface for backend/parser/tool mechanics; web.architecture for frontend/Fallow checker assurance. Current-source policy operations retain their semantic owners. |
| `tools/reference-pack-modules`, `tools/reference-packs` | harness.command_surface and module.reference_data; isolated producer closure, explicit pinned input and canonical output. |
| `tools/release-evidence` | harness.release; package/release/evidence and cleanup contracts. Registrar modules invoke assertions; filenames alone do not create cases. |
| 282 Go helper/setup files and three TestMain registrations | Their S-001 package owners; fixture construction, typed test ports and process teardown. Keep with their consuming tests, not as duplicate catalog tests. |
| `FuzzTimeBucketArithmetic` | module.networkflow; arithmetic exploration retained with its canonical deterministic seed witness. No fuzz campaign or benchmark result is claimed. |
| Ten Python files in the vendored UI tooling snapshot | Upstream/development context; retain snapshot fidelity, exclude from product runtime/conformance and do not promise a new supported Python test surface. |

Placement review applies TH-HARNESS-REQ-053/054/800 to assertion purpose.
Nineteen existing self-regression rows move to full: five frontend fixture/helper
rows, eight native evidence/accounting rows, two command/parser/archive rows,
one WebSocket helper row and three application-fixture rows. Five frontend rows
also move to the harness owner with migrations; their evidence class is unit.
Reporting's closed import allowlist now has an architecture verification; the
shared production Go import/family row has static evidence class. Named current
source policies, generated/schema/catalog/migration checks, embedded assets,
runtime admission/reset and all three smoke roles remain reachable in required
fast/standard gates. Existing product/architecture rows in helper packages keep
their owner-selected tiers. No row was moved solely for a faster reported gate.
The routing regression independently checks every corrected full-tier sentinel
against fast/check/ci/release and direct owner/row selection.

Cost disposition: no controlled before/after timing experiment was run and no
speed percentage, saved setup duration or universal stability claim is made.
The API-only retirements remove their selected browser scenarios; shared startup
cost depends on the remaining partition. Selector deduplication keeps its browser
startup. Full-tier corrections change gate membership without measuring marginal
CPU, service, test-body or cleanup cost. Current functional run durations are
retained diagnostics, not a performance baseline. Retention/placement decisions
are therefore justified by assurance and the adopted harness contract.

## 3. Module Boundary Diagnosis

### 3.1 Structural diagnosis

The testing ecosystem is a cross-owner assurance system. It MUST NOT be treated
as one production module. Central source-policy enforcement, local architectural
assertions, tests of checker mechanics, assembled product witnesses, and
browser choreography hide different decisions. Shared machinery is justified
only when it preserves those distinctions.

| Responsibility found | Current location | Owner candidate | Disposition | Evidence / finding / contract |
| --- | --- | --- | --- | --- |
| Production test-support and generated-contract import isolation | codegenboundary Go tests. | Existing harness.generated_artifacts and module.revisions selection owners. | Correct scanner first; consolidate subset only after effective survivor/routing proof. | E-004; F-001/F-002; C-002. |
| Package capability and import topology policies | Local module guards plus backend JSON/checker. | Semantic module policy owners; harness owns mechanics. | Consolidate overlapping mechanics after policy equivalence; unresolved file/package authority deferred. | E-005–E-007; F-003/F-004; C-003. |
| Projection SQL and descriptor consistency | projectionassembly tests, central SQL parser/policy. | module.projections; harness parser owner. | Separate reusable parsing from descriptor/source-authority obligations. | E-008; F-005; C-004. |
| Frontend dependency and semantic capability boundaries | Local Vitest guards and frontend policy. | web.architecture, web.workbook, module.imports and relevant owners. | Consolidate identical dependency rules; preserve distinct symbol/type/transport checks. | E-009; F-006/F-007; C-005. |
| Public API contracts executed through browser fixtures | incident.support and coordination-public-route. | app.server, module.incidents, module.workbook. | Evaluate backend relocation with retained rendered observations. | E-010–E-013; F-008–F-011; C-006–C-009. |
| Selector construction versus actual mounted selectors | ui-contracts unit tests and auth.support. | package.ui / module.auth. | Deduplicate pure equality; preserve attachment assertions until equivalent coverage exists. | E-014; F-012; C-009. |
| Checker, fixture, lifecycle, producer, and release self-assurance | tools, internal/testutil, referencepacks, package.test_utils. | Existing semantic/harness/producer owners; unresolved groups remain explicit. | Retain by default; investigate selection and marginal cost. | E-016/E-019/E-020; F-013/F-014; C-001/C-010. |
| Human source requirements and downstream routing fidelity | Core, harness NLSpec, verification/routing inputs. | Allocated specification owner, then harness projection owner. | Resolve drift through primary owner; never through passing tests alone. | E-001/E-021; F-015/F-016; C-011. |

### 3.2 Consolidation decision protocol

**TCR-REQ-009 — Retirement gate.** Removal or relocation MUST preserve every
active obligation of the original case, including useful engineering assurance
without an explicit product clause. Before retirement, the change record MUST
identify the original assertion, required behavior or engineering purpose,
surviving assertion, implementation boundary, execution scope and exemptions,
runtime/fixture profiles, owner selection, and compatible validation evidence.
A stricter name, larger suite, shared helper, or passing replacement label is
insufficient.

**TCR-REQ-010 — Decision algorithm.** A candidate MUST be evaluated in this order:

1. Establish its owner-required behavior or engineering purpose and actual
   assertion scope.
2. Separate ineffective or contradictory expectations from valid obligations.
   Correctness changes MUST be tracked separately from structural movement.
3. Identify the survivor for each valid obligation; preserve any unmatched
   obligation in the original or a separately characterized test.
4. Compare input/failure cases, observable outcomes, scopes, exemptions, execution
   boundaries, profiles, and owner selection.
5. Establish effectiveness with relevant negative and allowed-transformation
   fixtures, plus compatible runtime evidence where runtime behavior is claimed.
6. Update authored selection/accounting inputs atomically with any retirement;
   use canonical generation only during authorized implementation.
7. Retire only when every preservation and validation gate is satisfied.
   Otherwise retain and record the missing evidence.

Retirement for an obsolete purpose MUST identify the adopted owner decision or
bounded engineering-owner justification that ended that purpose. Mere absence
of a normative reference MUST NOT satisfy that condition. Independent process,
browser, order, security, or failure-cleanup evidence MUST remain distinct when
a lower layer cannot observe its boundary.

**Rationale — non-normative.** The primary structural burden is duplicated
policy maintenance and mismatched execution boundaries. Reducing test count
without establishing effective surviving assertions would worsen assurance.

## 4. Public Contract and Behavior Freeze Map

**TCR-REQ-011 — Preservation boundary.** Consolidation MUST preserve
owner-aligned HTTP/CLI envelopes and errors, authorization, mutations and
history, projections, generated contracts, browser interaction, UI selectors,
and harness accounting wherever the affected test supplies evidence. Existing
defects MUST NOT be frozen as compatibility. Unrelated behavior MUST remain
outside each slice.

| Contract / surface | Observed behavior and inspected tests | Authority / required behavior | Disposition / prerequisites | Finding, slice, validation |
| --- | --- | --- | --- | --- |
| C-001 Harness selection and evidence | 2,611 rows, mixed evidence classes, direct/aggregate/explicit selection. E-003/E-018/E-019. | Harness REQ-053/054, REQ-800/809; routing, smoke roles, explicit-only obligations, evidence compatibility. | Preserve selection meaning, fixture ownership, diagnostics, cleanup, and truthful accounting. | F-013/F-014/F-017; S-011/S-014; V-001/V-011. |
| C-002 Runtime dependency isolation | Three guards prune intended roots; broad testutil and revisionsupport share scanner/exemptions. E-004. | Repository production/support boundary; exact generated-contract allowlist is an engineering policy requiring owner confirmation before widening. | Separate scanner correction; preserve current exclusion intent pending authority review. | F-001/F-002; S-003/S-004; V-002. |
| C-003 Backend module capabilities | Closed local allowlists, named central prohibitions, exact-file provider grants. E-005–E-007. | Core 01 REQ-01-004/005; adopted decisions referenced by Core 00 REQ-00-070–074, AC-539/529/552/558/560 respectively. | Preserve known capability restrictions; exact topology decisions blocked where adopted decision text is unavailable. | F-003/F-004; S-002/S-005; V-003. |
| C-004 Projection and SQL isolation | SQLC import/export check; provider source ownership; descriptor/boundary/recovery/migration table equality. E-007/E-008. | Core 01 §8 projection ownership and REQ-01-649 source-port/read boundaries; exact package/file permissions remain separately justified policies. | Preserve all distinct guarantees; central table policy does not replace descriptor consistency. | F-005/F-018; S-005/S-006; V-004. |
| C-005 Frontend authority and generated consumption | Import policies, ambient fetch detection, source ownership, focus/geometry typing, import-target semantics. E-009. | Core 03 workbook behavior; applicable Network Flow owner; exact source spelling is not established as a normative interface. | Preserve semantic ownership and behavior; replace brittle spelling only with effective semantic checks. | F-006/F-007/F-018; S-007; V-005. |
| C-006 Incident request admission | Unknown create member and forbidden title patch are asserted in browser probes and request units; handlers invoke the admission functions. E-010. | Core 01 REQ-01-161/175 and §3.3.6.2 error details; associated incident AC family. | Preserve HTTP status, code, field, reason, admission/serialization and authorization boundary. | F-008; S-008; V-006. |
| C-007 Discovery and reserved dispatch | Generated registry payload, zero membership, pagination error, reserved root/descendant/outside paths; route/process counterparts. E-010/E-011. | Core 01 REQ-01-542–548; AC-370/371 and EXT-AC-156/157. | Preserve exact shape/order/claims and dispatch/error semantics. Unit/provider projections alone do not prove deployed assembly. | F-009/F-010; S-008; V-007. |
| C-008 Coordination mutations and current authorization | Seven public create/query/patch round trips; API membership demotion; denied task patch and unchanged state. E-012/E-013. | Core 01 §3.3.4/3.3.5, Core 04 incident-role authorization, corresponding source-owner contracts. Full per-field AC mapping pending RB-002. | Preserve seven fields/surfaces, same-session role change, public route effects, and rejection safety. | F-011; S-009; V-008. |
| C-009 Browser-visible state and selector contract | Empty incident landing, mounted error/status nodes, grid keyboard/focus/paste/virtualization support scenarios. E-014/E-015. | Core 04 AC-414; Core 03 workbook interaction; selector builders are a shared engineering interface. | Keep real rendered witnesses; pure builder equality cannot establish attachment or interaction. | F-009/F-010/F-012; S-008/S-010; V-009. |
| C-010 Supporting/producer/security assurance | Fixture lifecycle, checker rejection mechanics, generated drift, offline producer tests, release/security rows. E-016/E-020. | Harness owner, generated-artifact policy, adopted Reference Pack/Core security owners as applicable. | Do not remove by path/classification. Per-family normative or engineering purpose remains required. | F-013/F-014/F-017/F-018; S-001/S-011/S-012; V-010/V-012. |
| C-011 Recovery command authority | Harness REQ-655 lists five operations; Core lists seven including export and portable restore. E-021. | Core 01 REQ-01-593, Core 04 AC-428; Harness REQ-654 explicitly subordinates routing to product owner. | Separate correction to subordinate restatement; do not infer implementation/routing is defective without inspection. | F-015; S-013; V-013. |

WebSocket protocol, saved views, source storage, change-set/history semantics,
and generated wire schemas have no authorized production change in this effort.
Their tests MUST still be inventoried and their obligations retained. A later
slice touching their evidence MUST add the affected owner clauses and exact
case mapping before entry; a generic preservation statement cannot replace it.

## 5. Coupling and Boundary Findings

### 5.1 Findings register

| Finding | Evidence / confidence | Affected contracts | Classification / owner | Required action and scoped acceptance | Slice / validation |
| --- | --- | --- | --- | --- | --- |
| F-001 Three boundary tests have empty intended scan scope. | E-004; confirmed static. Immediate `cmd` and `internal` directories fail slash-suffixed prefix checks and are pruned. | C-002 | must_fix before guard consolidation; harness.generated_artifacts and module.revisions. | Prove production-root discovery and reject injected forbidden imports; zero selected files MUST fail. | S-003; V-002; AC-006. |
| F-002 Revisions test-support prohibition is a subset of broad testutil prohibition. | E-004; confirmed logical overlap, no effective runtime proof. | C-002/C-001 | should_fix after F-001; same owners. | Retain one effective broad prohibition only if Revisions-focused selection still reaches it. Generated-contract guard remains separate. | S-004; V-002/V-011; AC-007. |
| F-003 Local package capability guards only partially overlap central named prohibitions. | E-005/E-007; confirmed scope difference. | C-003 | should_fix; affected module owners and checker owner. | Compare full allow/deny semantics; retain closed allowlists where central policy lacks equivalent coverage. | S-005; V-003; AC-008. |
| F-004 Provider package permissions and exact-file grants differ. | E-006/E-007; confirmed policy difference. | C-003 | must_fix for the affected consolidation; authority RB-001. | Resolve whether file identity is intentional; demonstrate approved rename/split behavior and rejected bypasses. | S-002/S-005; V-003; AC-008. |
| F-005 Projection SQL ownership has duplicated parsing and a fixed source-file list. | E-008; confirmed structure; missed new-file/SQL-form coverage is a potential concern. | C-004 | should_fix; module.projections and checker owner. | Preserve descriptor authority checks while qualifying source enumeration and parser equivalence. | S-006; V-004; AC-009. |
| F-006 Frontend local/central dependency rules overlap without semantic parity. | E-009; confirmed differences in named imports, type-only treatment, resolved modules, selected paths. | C-005 | should_fix; web.architecture/web.workbook. | Transfer only equivalent rules; preserve unique typing, semantic focus, composition, fetch and ownership checks. | S-007; V-005; AC-010. |
| F-007 Import-target test asserts literal source spelling. | E-009; confirmed substring checks; alias/format sensitivity is inferred from predicates. | C-005 | should_fix; module.imports. | Replace spelling dependence only after bypass detection and behavioral registry/claim rejection remain effective. | S-007; V-005; AC-010. |
| F-008 Incident request-error browser test has no incident-UI assertion. | E-010; confirmed; admission units cover the same two cases. | C-006 | should_fix; module.incidents. | Preserve assembled HTTP status/code/details witness before removing browser execution. | S-008; V-006; AC-011. |
| F-009 Discovery browser test mixes API contracts with real empty-directory rendering. | E-010/E-011/E-015; confirmed. | C-007/C-009 | should_fix; app.server. | Preserve full deployed discovery profile and error details; separately retain a selected rendered landing witness. | S-008; V-007/V-009; AC-012. |
| F-010 Reserved-family browser coverage is narrower than apparent variable naming. | E-010/E-011; confirmed. Both profiles are enterprise_authentication; first family is /api/v1/auth/oidc, so placeholder replacement does nothing. | C-007/C-009 | should_fix; module.incidents/app.server. | Account for actual root/descendant cases, not a second profile or user-template case; prove surviving process/route and landing coverage. | S-008; V-007/V-009; AC-012. |
| F-011 Coordination API scenario uses a browser without rendered coordination assertions. | E-012/E-013; confirmed execution shape; complete redundancy not established. | C-008 | should_fix; module.workbook. | Relocate only with all seven patch round trips and API-driven same-session demotion preserved. Existing patch matrix is insufficient. | S-009; V-008; AC-013. |
| F-012 Auth support repeats a builder equality and consecutive attachment assertion. | E-014; confirmed duplicate assertions. Actual cross-screen attachment checks are distinct. | C-009 | should_fix; module.auth/package.ui. | Deduplicate pure assertions after narrow qualification; retain mounted-node coverage until mapped replacement exists. | S-010; V-009; AC-014. |
| F-013 Catalog behavior labels do not establish semantic purpose or necessary placement. | E-003; confirmed metadata limitation; e.g. Reporting boundary row references a product behavior verification. | C-001/C-010 | should_fix; catalog and semantic owners. | Review actual responsibility; correct classification only under adopted harness rules, with selection preserved. | S-001/S-011; V-001/V-010/V-011; AC-003/015. |
| F-014 Supporting tests may carry fixture, lifecycle, dependency, and workflow costs disproportionate to their selected purpose. | E-019/E-020; investigation hypothesis. No marginal cost established. | C-001/C-010 | defer optimization decision pending measurement; owning families. | Attribute setup, cleanup, resources, contention, and critical-path effects; distinguish selection reduction from executor improvement. | S-011; V-010; AC-016. |
| F-015 Recovery command restatement conflicts with primary Core grammar. | E-021; confirmed documentary drift, not demonstrated implementation failure. | C-011 | must_fix for complete recovery traceability; harness prose/routing owner. | Reconcile five-command restatement to seven-command Core authority; separately inspect actual routing and test coverage. | S-013; V-013; AC-017. |
| F-016 Referenced adopted architecture decision files are absent. | E-001; confirmed at baseline. | C-003 | must_fix for topology-dependent decisions; affected owner documents. | Locate authoritative adopted replacement or obtain owner resolution; do not invent exact topology policy. | S-002; V-003; AC-005. |
| F-017 Suite-wide requirement coverage, uncatalogued registration closure, and stability remain partly unassessed. | E-002/E-003; explicit evidence gap. | C-001/C-008/C-010 | must_fix for full-effort completion; all participating owners. | Complete review records and gap dispositions; no fabricated whole-suite stability or coverage claim. | S-001/S-012; V-001/V-012; AC-002/004/018. |
| F-018 Apparently similar checks retain unique assurance. | E-007–E-009/E-016; confirmed source distinctions. | C-004/C-005/C-010 | intentional/no_action for blanket deletion. | Retain SQLC export checks, descriptor consistency, ambient transport, ownership completeness, Graph exported-surface checks, and detector-negative fixtures unless their exact obligation transfers. | Preservation obligations in S-005/S-006/S-007/S-011; V-004/V-005/V-010. |
| F-019 Canonical Markdown lint does not select the requested path and can write outside it. | E-017/E-025; baseline wrapper effects and current direct Biome `--write` invocation confirmed. | C-001 | defer lint-configuration change; outside this author's write scope. | Record skipped lint and direct document review; do not rename output or edit configuration to manufacture a pass. | V-014; AC-019. |

Additional findings confirmed during planning and accepted for implementation:

| Finding | Evidence / confidence | Affected contracts | Classification / owner | Required action | Slice / validation |
| --- | --- | --- | --- | --- | --- |
| F-020 Duplicate Core acceptance identities AC-483/537/538/539/540. | Core 04 definitions directly inspected; confirmed static. | C-003/C-011 and affected behavioral owners. | must_fix for unambiguous evidence mapping; Core owners. | Assign unique IDs and update references by semantic meaning. | S-002; V-003/V-014. |
| F-021 Generated-contract guard targets retired internal/gen/contract rather than current family packages. | codegenboundary guard and contracts/index.json/current generated packages inspected; confirmed static. | C-002. | must_fix for effective generated boundaries; harness.generated_artifacts. | Registry-derived family discovery with explicit package-consumer authority and effective fixtures. | S-016; V-002/V-011. |
| F-022 Projections directly locks authoritative record envelopes while refreshing links. | S-006 recursive SQL qualification exposed the existing lock in runtime/link_effects.go; confirmed production owner defect. | C-003/C-004. | must_fix; Records, Projections and application composition. | Move exact ordered incident-scoped lock behind a Records-owned borrowed-transaction capability and inject it through the consumer contract; remove direct permission. | S-017; Records/Links integration and adapter/error fixtures PASS. |

### 5.2 Architectural preservation requirements

**TCR-REQ-012 — Effective guards.** Source-policy checks MUST prove that their
intended roots are selected. Empty selection MUST fail where production source
is required. Negative fixtures MUST prove the forbidden construct reaches the
detector. Allowed rename/split, alias, quote, and support-exemption fixtures MUST
remain accepted when the adopted policy permits them. The scanner correction
MUST NOT silently widen or narrow dependency policy.

**TCR-REQ-013 — Backend equivalence.** Consolidation MUST compare root-only
versus recursive scope, exact versus subtree imports, package versus file grants,
generated/support/test exclusions, and build-constraint treatment. Checker
mechanics/CLI tests MUST NOT be removed as duplicates of production-source
checks. The broad testutil ban MUST NOT be replaced solely by the pgtest ban.
The SQLC export check MUST retain rejection of generated SQL types in the
exported declaration forms it currently inspects; a dependency allowlist alone
does not supply this guarantee.

**TCR-REQ-014 — Projection preservation.** SQL parsing consolidation MUST retain
both relation access enforcement and provider-declared source authority. Active
descriptor, boundary-policy, recovery, and migration table sets MUST continue to
agree. Tests MUST distinguish reads, writes, and locks; qualify quoted/schema
names and CTEs; and demonstrate that an extracted/new SQL-bearing provider source
does not escape review. Existing test-support delegation to production assembly
and protected exported surfaces MUST retain their own evidence.

**TCR-REQ-015 — Frontend preservation.** Moving dependency checks MUST preserve
resolution of aliases, namespace imports, re-exports, dynamic imports, and
type-only distinctions where relevant to the rule. Ambient fetch detection,
complete source ownership, semantic grid focus/geometry, privileged capability
typing, and Network Flow browser-evidence/design-token restrictions MUST remain
separate obligations unless an equivalent surviving check demonstrably covers
them. Literal source-spelling assertions MUST NOT be replaced by import presence
alone when the intended obligation is actual claim-gated behavior.

### 5.3 Browser probe mapping

**TCR-REQ-016 — Browser relocation boundary.** A `page.request` assertion
MUST be credited for the API boundary actually exercised. Manual CSRF header
construction MUST NOT be credited as UI-client submission, negative CSRF,
cookie-policy, or CORS evidence. API fixture setup or post-UI persistence checks
MUST NOT be classified as redundant merely because they use requests.

A static search found 104 direct `page.request` occurrences in 34 browser
spec files. This excludes aliases/helper calls and does not classify all 34
files. The 18 catalog rows in browser-support families include real keyboard,
paste/fill, grouping, virtualization, and pending-navigation observations.
Those rows MUST NOT be retired as a class.

| Candidate / original selector | Existing counterpart and limitation | Unique assurance to preserve / acceptance gate |
| --- | --- | --- |
| `incident.support.spec.ts`: route-owned validation errors | `TestIncidentCreateAcceptsDeclaredMembersAndNormalizesIncidentKey_Unit` and `TestIncidentPatchAllowsPromotedFieldsAndKeepsNoOpVersionStable_Unit` call admission functions. | Assembled create and patch routes MUST reject the exact payloads with required HTTP status, code, field, and reason. Route binding, middleware, and serialization are not proven by admission alone. |
| Same file: zero-membership discovery and singleton pagination | Discovery response-shape/order and singleton-query units; service session-sliding test; process test checks six items and one claim. | Preserve zero-membership authentication, exact seven-member item contents/order, actual browser-stack claim set, and pagination reason details. Retain rendered empty directory separately. |
| Same file: reserved-family precedence | Dispatcher unit covers claimed/unclaimed/base/outside behavior; integration and process witnesses cover deployment assembly. | Preserve readiness/base dispatch, root and descendant unclaimed responses, canonical details, outside fallback, and compatible claim configuration. Do not count two enterprise-profile variables as two profiles. |
| `coordination-public-route.spec.ts`: public view/row mutation with current-role authorization | `TestWorkbookSurfaceContracts_Integration` covers all seven create/query surfaces; `TestWorkbookPatchSurfaceContracts_Integration` covers only Timeline and Parties and uses SQL role changes. | Preserve all seven field patch/read-back/version assertions, public API membership demotion, the same authenticated session, permitted reads, denied mutation details, and unchanged task state. |
| `auth.support.spec.ts`: stable selector builders | `application-selectors.node.test.ts` already checks bootstrap-setup-key equality; feedback attachment is repeated without an intervening action. | Pure duplicates may be retired after narrow evidence. Actual shell, landing, account, deployment, and login attachment/visibility observations remain separately mapped. |

The coordination preservation matrix is closed for S-009:

| Surface | Patched field | Required postcondition |
| --- | --- | --- |
| task_requests | `task.priority` | Response/read-back value and increased row version; after API demotion, read remains permitted, patch is denied, version/value unchanged. |
| decisions | `decision.status` | Response/read-back value and increased row version. |
| parties | `party.notes` | Response/read-back value and increased row version. |
| comm_log | `comm_log.summary` | Response/read-back value and increased row version. |
| handoff | `handoff.next_checks` | Response/read-back value and increased row version. |
| status_review | `status_review.active_risks_summary` | Response/read-back value and increased row version. |
| lesson | `lesson.closure_state` | Response/read-back value and increased row version. |

The initial discovery probe expects import, incident_portability,
snapshot_reporting, and reference_pack claimed; enterprise_authentication and
network_flow_activity unclaimed. This is a fixture configuration, not a universal
production default. Replacement evidence MUST bind the corresponding profile or
explicitly map a separately qualified equivalent configuration.

### 5.4 Stability, burden, and coverage requirements

**TCR-REQ-017 — Reliability evidence.** A flaky classification MUST distinguish
observed inconsistent outcomes under materially equivalent inputs from a
potential timing/shared-state concern. Records MUST identify source/profile,
exact command, attempts, failure signatures, and environmental differences.
A pass after retry MUST NOT erase the original failure. This handoff MUST NOT
authorize blind reruns, increased timeouts, service resets, or relaxed assertions.

**TCR-REQ-018 — Cost evidence.** Investigation MUST distinguish setup, test-body,
cleanup, shared-service, resource-admission, and critical-path costs. Evidence
MUST identify exact selected rows/cases, source, profiles, environment, cache
mode, and measurement limits. Removing an assertion within an existing browser
scenario MUST NOT be claimed to remove browser-stack startup cost. Reducing
full-tier rows MUST NOT be claimed to accelerate the fast gate. No percentage,
duration target, or harness performance qualification gate is introduced.

**TCR-REQ-019 — Supporting assurance.** Supporting tests MUST be evaluated by
the failure they detect and the boundary they prove. Required smoke roles,
generated drift/policy, migration integrity, product/static/security checks,
runtime admission/reset contracts, full-tier self-regressions, and explicit-only
UI-review/visual obligations MUST follow the adopted harness placement rules.
A test MUST NOT be moved solely to improve a reported gate duration.

**TCR-REQ-020 — Coverage gaps.** Each identified owner requirement without
established test evidence MUST record the required observation, existing partial
coverage, missing scenario, responsible owner, and blocking decision. Tests with
no explicit requirement MUST record their engineering purpose rather than be
discarded. Existing evidence that contradicts the primary owner MUST be
classified as characterization or negative evidence until separately corrected.
The five-command harness restatement MUST NOT override Core's seven commands.

## 6. Refactor Workstreams

### Current corrective execution graph: S-051

**Scoped GO for the documented pending change set.** F-050 and S-051 are complete:
every defer in a raw SQL helper's own body rejects its summary. Fresh mechanics
and production checks qualify both consumers; S-012 and S-014 are complete.
F-047–049 and S-047–050 retain compatible evidence for unchanged obligations,
including Links/Network Flow migrations. No general control-flow/heap/callback
solver or compatibility exemption was introduced. Whole-suite, full-owner, release
and universal SQL soundness claims remain outside this recommendation. No
production exploit path was established.

Baseline: main at `beaf0293634e6aef9e62aa5670e033a26b6458d3`; 524 tracked changes,
44 untracked files, nothing staged. The current content/file-identity baseline is
`/tmp/cartulary-s051-baseline-20261110.json`; the S-047 baseline is historical.
Preserve unrelated edits and the
index. No staging, commit, merge or deployment is authorized by this task.

| Slice | Prerequisite | Status | Exit |
| --- | --- | --- | --- |
| S-047 Contract and reproductions | Accepted plan/baseline | DONE | Harness and guide clarify completeness; retained regressions reproduce all gaps. |
| S-048 Helper returns | S-047 | DONE | Explicit/bare returns have complete bounded summaries; both consumers reject unsupported flow. |
| S-049 Result effects | S-048 | DONE | Result destinations with nested writers/escapes reject; the historical raw-helper cleanup allowance is superseded by S-051. |
| S-050 Source admission | S-047 | DONE | Shared strict admission covers all collectors; public inspection/CLI reject incomplete input. |
| S-051 Recovered-panic rejection | Accepted corrective plan | DONE | All own-body defers reject raw summaries; both consumers and unaffected controls qualified. |
| S-012 Coverage reconciliation | S-051 | DONE | Requirement/fixture/run mapping and compatible pending-change evidence are explicit. |
| S-014 Handoff | S-012 | DONE | Final checks, preservation review and bounded readiness recommendation. |

| Finding | Areas and remediation | Rationale/benefit | Compatibility and unresolved risk | Acceptance |
| --- | --- | --- | --- | --- |
| F-047 Omitted helper exits | Specification, implementation, tests, guide: private complete return summary; named returns; unsupported control flow rejects. | Explicit completeness prevents safe partial alternatives from becoming authority. | Raw tuple and unmodeled control flow reject; otherwise unrestricted runtime SQL can inherit a safe literal. | Sink assertions, both branch orders, switch/select/jump/loop and nested-function cases through both consumers. |
| F-048 Deferred result replacement | Same areas: use binding effects on named result destinations for every return form. | One effect rule avoids callback scheduling guesses. | Mutable/escaped raw result helpers need immutable or sealed construction; otherwise deferred writes bypass both policies. | Explicit/bare returns, nested/deferred/concurrent writers and address escapes reject; safe cleanup and bound data pass. |
| F-049 Incomplete source admission | Same areas: shared filesystem checks; required roots; linked/nonregular/unreadable inputs fail before completeness. | Strict mechanism prevents collectors drifting while preserving separate policy scopes. | Missing-root fixtures must create directories; no optional-root exemption. Otherwise forbidden SQL silently disappears. | Public backend inspection/CLI plus provider/generated/exact-set controls cover missing/link/read failures and valid inputs. |


### Historical corrective execution graph: S-043–046

**Historical scoped GO; superseded by S-047 NO-GO.** S-043–046 and reopened S-012/S-014 are complete under the
maintainer's implementation authorization. F-044–046 superseded the S-042 handoff
and are now closed by the final corrective evidence below. No production exploit
path was established. Earlier passes alone do not qualify these counterexamples;
the new regressions and production gates do. This is readiness for the recorded
changes, not full-suite, whole-owner or release acceptance.

Baseline: main at `beaf0293634e6aef9e62aa5670e033a26b6458d3`, 523 tracked changes,
37 untracked files, nothing staged. Content/file-identity inventory and originals
are retained privately at `/tmp/cartulary-s043-5aw5ruk3`; unrelated changes must be
preserved. No commit, merge or deployment is requested.

| Slice | Prerequisite | Status | Exit |
| --- | --- | --- | --- |
| S-043 Authority/readiness | Accepted plan and current baseline | DONE | Harness clarifies execution-state completeness; current GO superseded. |
| S-044 Shared state/branches | S-043 | DONE | All continuing states join; finite collection snapshots preserve alternatives in both policies. |
| S-045 Assignment transfer | S-044 | DONE | Every destination updates or becomes unknown, with no stale state. |
| S-046 Loop qualification | S-045 | DONE | Repeated boundaries account for writes before analysis; immutable controls pass. |
| S-012 Coverage reconciliation | S-044–046 | DONE | Current regressions, both production gates and catalog cover the changes. |
| S-014 Handoff | S-012 | DONE | Documentation, preservation review and exact evidence justify scoped readiness. |

| Finding | Areas/remediation | Rationale and benefit | Compatibility and unresolved risk | Acceptance |
| --- | --- | --- | --- | --- |
| F-044 Branch collections (P2, must_fix) | Shared scalar/collection/type/provenance state joins; implementation, tests, Harness and guide. | One state contract prevents branch-order dependence and future omitted channels. | Finite alternatives stay supported; unsupported inputs reject. Otherwise unauthorized relations disappear. | Both branch orders, switches, scope/return/empty/unknown alternatives and both policies qualify. |
| F-045 Multi-result assignment (P2, must_fix) | Complete destination transfer/invalidation across assignment/declaration; implementation, tests, Harness and guide. | One assignment rule prevents stale scalar, collection and provenance state. | Unsupported result positions reject; simplify raw construction or use sealed values. Otherwise runtime input inherits old authority. | Reassignment, :=, var, blank/later positions and parallel assignment qualify at the actual sink in both policies. |
| F-046 Loop-carried state (P2, must_fix) | Pre-invalidate loop writes and conservatively summarize exits; implementation, tests, Harness and guide. | Sound bounded iteration without a general solver. | Stateful raw builders must become immutable per-iteration expressions, bound parameters or sealed fragments. Otherwise later iterations execute unchecked SQL. | Body/post/continue/nested/range and zero-iteration exit cases reject when unresolved; immutable controls pass. |

The private inspection response and policy authority remain unchanged. No new
schema, grants, exclusions, tuple inference engine or compatibility fallback was
introduced. Domain vocabulary needs no amendment. This graph is historical; S-047–050 above owns current corrective sequencing.

### Historical corrective execution graph: S-036–042

**Historical GO, superseded by S-043 NO-GO.** F-038–040 and separately
recorded follow-ups F-041–043 are closed. Affected analyzer/native acceptance in
S-019/S-024/S-026/S-032/S-035/S-037/S-038 is restored; S-036–042 and reopened
S-012/S-014 satisfy their exits. The latest S-014 handoff below is authoritative.
Earlier assessments remain historical. Package grants, provider separation,
generated SQL coverage and earlier product migrations remain intact. This is
commit readiness within the recorded targeted scope, not full-suite or release
acceptance.

Baseline (2026-11-10): main at `beaf0293634e6aef9e62aa5670e033a26b6458d3`,
522 tracked changes and 35 untracked files; nothing staged. Content inventory:
`/tmp/tcr-s036-baseline`. No commit, merge or deployment is authorized.

This is the authoritative remaining graph. Update this tracker after every
workstream and before beginning its successor. No ownership expansion, exclusion
or compatibility fallback may conceal a failure. New defects require a separate
corrective slice before S-014.

| Workstream | Prerequisite | Status | Exit |
| --- | --- | --- | --- |
| S-036 Corrective baseline/specification | F-038–040 review | DONE | Authority, findings, evidence limits and contract agree. |
| S-037 Canonical callable admission | S-036 | DONE | SQL function values reject consistently; direct ports and non-SQL callbacks qualify. |
| S-038 Stable closure inputs | S-037 | DONE | Mutable captures and inherited stateful provenance reject; stable captures pass. |
| S-039 Unambiguous import admission | S-038 | DONE | All SQL-analysis dot imports reject; named/import-only controls and production gates pass. |
| S-040 Generic callable admission | S-039 reconciliation finding F-041 | DONE | Unbound SQL-result signatures and callable constraints cannot become SQL-free. |
| S-041 Closure collection isolation | S-040 handoff finding F-042 | DONE | No inherited or escaped mutable collection supplies inferred SQL; local finite controls pass. |
| S-042 Constant fallback identity | S-041 scope audit F-043 | DONE | Local constants cannot inherit package constants by name in unknown state. |
| S-012 Coverage reconciliation | S-042 | DONE | Every finding has current fixtures and production evidence; unchanged evidence is compatible. |
| S-014 Final validation/handoff | S-012 | DONE | Documentation, final diff and exact retained evidence justify readiness. |

| Finding | Areas / remediation | Rationale / benefit | Compatibility / unresolved risk | Acceptance |
| --- | --- | --- | --- | --- |
| F-038 Stale captures (P2, CLOSED) | Go binding/capture inventory, tests, Harness and guide; reject unstable inferred SQL and captured prepared/batch provenance. | Declaration identity and stability replace invocation-time guesses without heap/call-graph inference. | Internal builders may require immutable copies or local execution; otherwise later mutation substitutes unchecked SQL. | Deferred/concurrent/escaped callbacks, nested writers, address escapes, shadowing, copies and stateful handles qualify in both policies. |
| F-039 SQL callable values (P2, CLOSED) | Shared canonical callable classification, tests, Harness and guide; reject execution function values on invocation/extraction/transfer. | One semantic path makes inline/aliased/defined signatures equivalent and removes a second transport mechanism. | Use direct calls or narrow ports; ordinary application callbacks remain supported. Otherwise aliases or parameters erase execution. | Fields, parameters, locals, indexed/returned values and aliases reject; direct ports and non-SQL controls pass. |
| F-040 Dot-import admission (P2, CLOSED) | Go SQL-batch admission, tests, Harness and guide; reject every dot import with a located diagnostic. | Unambiguous name binding without a sensitive-package exception list. | Replace dot imports with named aliases; import-only inspection unchanged. Otherwise raw protocol operations disappear. | Protocol/driver/constructor/helper imports, generated/build-constrained inputs and named/import-only controls qualify. |
| F-041 Generic callable admission (P2, CLOSED) | Canonical type-parameter/constraint classification, tests, Harness and guide; reject ambiguous SQL function values. | Symbolic type identity cannot silently establish a non-SQL result; direct helper bodies retain their own authority. | Use concrete ports/signatures; no specialization or grants. Otherwise generic instantiation recreates F-039. | Generic results, constraints, conversions and extraction reject in both consumers; domain controls and production gates pass. |
| F-042 Closure collection state (P2, CLOSED) | Go environment isolation/binding inventory, fixtures, Harness and guide; reject captured or escaped finite-list SQL inference. | Every evaluator state channel obeys the capture proof without heap inference. | Use local finite lists, scalar copies or sealed fragments; otherwise list mutation can substitute unchecked relations. | Capture, nested writer, alias/address and field/index cases reject; local/non-SQL controls and both production gates pass. |
| F-043 Constant fallback identity (P2, CLOSED) | Go package declaration inventory, fixtures and guide; prevent name-only fallback for local constants. | Unknown state preserves lexical identity. | Unsupported contexts reject; otherwise local constants can inherit unrelated package authority. | Shadowing counterexamples fail explicitly; ordinary/cross-file/grouped constants and both policies pass. |

Core ownership remains authoritative. The located inspection response is sufficient
and remains unchanged. No public API, schema, data or deployment migration is
planned. Prior passing runs are historical for changed analysis; unchanged product,
query, restore and browser/HTTP obligations may retain compatible evidence.

### Historical S-030–035 execution graph

**Historical GO (superseded by S-036 NO-GO).** F-033–035 and the separately
recorded F-036/F-037 follow-ups are closed. S-019/S-024/S-026 acceptance is restored;
reopened S-012/S-014 satisfy their exits. The final S-014 checkpoint below records
current targeted evidence and its limits. Historical results remain evidence only
for compatible inputs and obligations. Core package/provider authority remains
unchanged. This is commit readiness, not full-suite or release acceptance.

Implementation baseline (2026-11-10): `main` at
`beaf0293634e6aef9e62aa5670e033a26b6458d3`, 521 tracked changes and 34 untracked
files, nothing staged. Baseline content hashes and focused source snapshots are
retained outside the repository at `/tmp/tcr-s030-baseline`. Preserve all existing
work; commit, merge and deployment are outside this effort.

The following prerequisite column is the authoritative execution graph.
Update this tracker after each completed workstream, before beginning the next.

| Workstream | Prerequisite | Status | Binary exit |
| --- | --- | --- | --- |
| S-030 Corrective baseline/specification | F-033–035 review | DONE | Findings, authority, acceptance and remaining graph agree. |
| S-031 Remove raw mutable SQL inference | S-030 | DONE | Raw builders fail at SQL boundaries; sealed/immutable alternatives and non-SQL controls pass. |
| S-032 Canonical type and boundary identity | S-031 | DONE | Alias-equivalent signatures agree; unresolved/conflicting identity fails; native inventory and production gates pass. |
| S-034 Generated SQL admission | S-032 | DONE | Generated dependencies resolve; generated accesses obey ownership scopes; linked/missing inputs fail. |
| S-033 Scoped SQL query operands | S-034 | DONE | Every admitted operand is classified; CTE/set/lock scopes and both consumers qualify. |
| S-035 Complete result signature admission | S-033 | DONE | Every result position participates; later database results and aliases cannot become SQL-free. |
| S-012 Coverage reconciliation | S-035 | DONE | No unresolved finding, missing selector, exemption or uncovered changed behavior. |
| S-014 Final validation/handoff | S-012 | DONE | Documentation, final workspace review and current targeted evidence justify readiness. |

| Finding | Areas and remediation | Rationale and benefit | Compatibility and unresolved risk | Acceptance |
| --- | --- | --- | --- | --- |
| F-033 Escaped builder mutation (P2, must_fix) | Go analysis, tests, Harness clarification and guide: remove raw strings.Builder SQL simulation; use immutable expressions or sealed fragments. | Removes mutable-object lifetime/alias inference and its escaping-method bypass class. | Raw-builder SQL becomes unsupported; ordinary text and bound parameter values remain valid. Otherwise a reset can hide unauthorized writes behind a stale comment prefix. | Escaped/direct builders reject at SQL boundaries in both consumers; immutable/sealed equivalents and non-SQL controls pass. |
| F-034 Aliased signature identity (P2, must_fix) | Shared declaration-aware type resolution, native/port analysis, tests and prose: canonical aliases; unresolved is never proven non-SQL. | Equivalent Go signatures retain equivalent capability decisions. | Private analysis only; hidden unresolved calls now reject. Defined types retain their own identity and methods. Otherwise a result alias disables ownership analysis. | Local/imported/chained/parameter aliases, fields, embeddings, handles, scope and build conflicts qualify; actual API inventory passes. |
| F-035 Hidden TABLE operands (P2, must_fix) | Shared SQL classifier, tests and prose: bounded query-operand decomposition with explicit CTE and set scopes. | Removes positional command exceptions and prevents partial successful classification. | Unsupported query structures fail; runtime SQL is unchanged unless a separately recorded migration is needed. Otherwise CTE/set operands conceal physical relations. | WITH/TABLE and set operands expose reads; CTE shadowing, writes, locks and unsupported structures qualify in both consumers. |
| F-036 Generated SQL scope omission (P2, must_fix) | Backend collection, authored policy scopes, tests and prose: include generated production Go in SQL analysis and enforcement. | Generated code receives the same construction and ownership checks as authored code. | No generator or capability grant changes. Otherwise generated helpers can resolve types while escaping table rules. | Missing/linked inputs fail; build-constrained generated declarations resolve; forbidden generated writes fail. |
| F-037 Non-leading database results (P2, must_fix) | Shared signature analysis, pinned API assurance, tests and prose: inspect every result position and inventory native result types. | Signature reordering cannot change boundary identity; driver result changes require review. | Private analysis only. Otherwise a leading data result hides a later database result. | Named/unnamed later results and aliases qualify in both policies; unresolved/derived results reject; pinned inventory and production gates pass. |

All findings map to Core REQ-01-690 and Harness SQL architecture analysis
completeness, downstream of the established C-004/TCR-REQ-014 ownership contract.
No ownership grant, scan exclusion or unknown-construction allowance may hide a
new failure. Newly exposed behavioral defects need their own corrective slice
before S-014. The existing inspector response remains unchanged; new failures
use located rejected dispositions. No public API/schema/data/deployment migration.

### Historical S-023–029 execution graph

F-027–029 and corrective follow-ups F-030–032 are closed. S-019/S-022 acceptance
is restored; S-020/S-021 capabilities remain intact. S-023–029 and the reopened
S-012/S-014 satisfy their exits. Earlier checkpoints remain historical, reusable
only where inputs and obligations are compatible. Readiness at that checkpoint
was **GO**; the current corrective graph and latest S-014 checkpoint supersede it.
Baseline: main at beaf0293634e6aef9e62aa5670e033a26b6458d3; the reviewed workspace
has 520 tracked changes and 30 untracked files, with nothing staged. Preserve
all existing work. No commit, merge, deployment or public/data migration is authorized.

| Workstream | Prerequisite | Status | Binary exit |
| --- | --- | --- | --- |
| S-023 Corrective baseline/specification | reviewed findings | DONE | Authority, acceptance and remaining graph agree. |
| S-024 Exhaustive SQL discovery | S-023 | DONE | Every SQL candidate is analyzed or rejected in both consumers. |
| S-025 Safe closed templates | S-024 | DONE | Shared runtime/static lexical validation rejects escapes; behavior passes. |
| S-026 Native API qualification | S-025; S-027/S-028 for production acceptance | DONE | Typed API inventory, arguments and explicit unsupported dispositions qualify. |
| S-027 Reference Data batch construction | S-026 admission implemented | DONE | Local batch construction preserves index rows, bounds and transaction ownership. |
| S-028 Closed test-support inputs | S-026 native detection | DONE | Existing helper calls use qualified construction without new transport grants. |
| S-029 Raw protocol construction admission | S-012 reconciliation observation | DONE | Direct protocol constructors and message construction fail explicitly. |
| S-012 Coverage reconciliation | S-026–029 | DONE | Current fixtures, selectors and behavior evidence close every finding. |
| S-014 Final validation/handoff | S-012 | DONE | Required checks pass and final assessment matches the workspace. |

Update this tracker after every workstream and before starting the next. Newly
exposed behavioral defects require a separate corrective slice before S-014;
no widened permission, exclusion or unknown-construction allowance may hide them.

| Finding | Areas / remediation | Rationale / benefit | Compatibility / unresolved risk | Acceptance |
| --- | --- | --- | --- | --- |
| F-027 AST omissions (P2) | Go analysis, private protocol, both checkers, specification, tests and guide: separate exhaustive discovery from bounded evaluation; located dispositions. | Syntax traversal no longer depends on evaluation support. | Internal protocol migrates atomically; otherwise field/index and nested calls conceal unauthorized access. | All expression contexts observed; unsupported candidates fail; scope and non-SQL controls pass. |
| F-028 Template comment escape (P2) | Shared pure lexical validator, runtime constructors, analyzer, tests, specification and guide. | One runtime/static contract removes context-sensitive substitution. | Closed templates reject comments and partial tokens; raw SQL comments remain supported. Otherwise quoted data can escape a comment and add relations. | Reported payload and quote/comment/token adversaries fail at runtime and analysis; valid query behavior passes. |
| F-029 Native API omissions (P2) | Typed/signature boundary registry, analyzer, API inventory assurance, specification and guide. | Driver changes cannot silently bypass enforcement. | Unused pipelines, named execution, textual low-level COPY and raw protocol access are rejected; no heuristic fallback. Otherwise raw native SQL escapes policy. | Every admitted signature/argument position qualified; unsupported APIs fail; same-named application controls pass. |

### Historical S-018–022 execution graph

The maintainer authorized the remediation plan after review identified three
remaining P2 enforcement gaps. This graph supersedes prior completion claims for
S-005/S-006 and reopens S-012/S-014. Previous completion records remain historical
source-scoped evidence, not current closure. Baseline: main at
beaf0293634e6aef9e62aa5670e033a26b6458d3, 60 tracked changes and 10 untracked
paths, no staged changes. All prior work is preserved.

| Workstream | Prerequisite | Status | Exit |
| --- | --- | --- | --- |
| S-018 Authority and corrective baseline | reviewed baseline | DONE | Owner rules, findings and current graph agree. |
| S-019 Shared bounded SQL analysis | S-018 | DONE | Both checkers use qualified Go AST SQL analysis; unsupported SQL fails. |
| S-020 Package/declaration permissions | S-019 | DONE | Allowed renames/splits pass both checks; forbidden capability changes fail. |
| S-021 Complete provider dependency coverage | S-020 | DONE | Recursive dependency coverage and read-capability separation qualify all providers. |
| S-022 Bounded SQL compatibility exposed by the shared backend gate | S-021 | DONE | Generic maintenance SQL has explicit closed capabilities; no unknown construction is silently admitted. |
| S-012 Coverage reconciliation (reopened) | S-022 | DONE | Every correction has compatible positive/negative and affected behavior evidence. |
| S-014 Final validation and handoff (reopened) | S-012 | DONE | Targeted checks pass and final handoff accurately closes the findings. |

Update this tracker after each completed workstream and before starting the next.
The selected SQL design is bounded AST analysis, not arbitrary Go interpretation.
Unknown SQL construction must fail; source owners retain legitimate write/lock
capabilities without granting them to provider read paths. No public/data migration,
commit, merge or deployment is part of these corrective slices.

| Finding | Areas / remediation | Rationale and benefit | Compatibility / unresolved risk | Validation |
| --- | --- | --- | --- | --- |
| F-023 Composed SQL escapes analysis | Specification, Go analysis, shared SQL checker and tests: extract complete expressions via Go AST; remove lexical Go fallback. | One syntax owner and explicit unsupported forms prevent silent partial analysis. | Internal protocol/builders change atomically; undetected reads/writes/locks otherwise pass gates. | Leading constants, UNION suffixes, assignments, branches, appended locks and unsupported sinks. |
| F-024 Extracted helpers escape coverage | Specification, dependency discovery, source read capabilities and tests: close repository-local imports for each provider and reject unassigned SQL. | Extraction stays visible and shared helpers retain each consumer's authority. | Internal ports/assembly may change without aliases; otherwise helpers can launder authority. | Sibling/transitive extraction, shared-provider permissions, missing/linked sources and source-lock separation. |
| F-025 Filename grants contradict package capabilities | Specification, policy/schema and tests: exact package grants plus declaration/statement restrictions. | File organization no longer determines permission. | Internal policy migration only; otherwise refactors fail inconsistently and exceptions accumulate. | Rename/split parity, denied peers/descendants, wrong tables/operations and stale grants. |
| F-026 Shared analysis exposes unrestricted generic SQL | Implementation, canonical generation, tests and documentation: typed structural fragments, exact transport/control grants and frozen-catalog maintenance commands. | Explicit capabilities preserve extensibility without arbitrary string inference or scan exclusions. | Internal contributed query ports migrate atomically; otherwise unresolved production SQL would either break enforcement or require unsafe exemptions. | Production gates, closed-constructor/maintenance negatives, source lifecycles, keysets, restore integrity and reset preservation. |

### Corrective checkpoint: S-018 complete

Core 01 now states filename-independent SQL capabilities and consumer-specific
read authority through helper dependencies. The harness completeness contract
requires Go AST extraction, explicit unsupported-construction failures and shared
classification. F-023–025 and all six corrective checkpoints are recorded above.
`make lint-markdown` passed (124 files, no changes). No production changes in
this slice. Next: S-019; final readiness remains NO-GO.

### Corrective qualification note: S-019 remains in progress

The AST extractor and shared relation classifier pass the projection declaration
ownership check (`tcr-s019-sql-20261109-05`), compiled-plan/keyset checks
(`tcr-s019-query-20261109-02`), closed fragment/predecessor checks
(`tcr-s019-fragments-20261109-02`), and pagination checks
(`tcr-s019-keyset-20261109-02`). All four owning commands exited zero. The source
inspection row also passed within `tcr-s019-mechanics-20261109-04`; that run remains
failed because three former partial-analysis fixtures must require rejection.

The shared backend gate (`tcr-s019-boundary-20261109-03`, exit 2, product /
`test_assertion_failure`) now exposes an additional compatibility obligation:
`internal/app/recoveryassembly/vnext.go`, `vNextSnapshot.StreamCanonicalTableRows`,
constructs its relation from `pgx.Identifier(...).Sanitize()`. Sanitizing an
identifier does not establish its complete permission set. **F-026 / S-022** must
qualify the generic maintenance construction and any further instances exposed
by this gate before coverage reconciliation and final handoff. This is separate
from provider authority; no fallback, exclusion, or blanket source-owner grant
is authorized. The remaining graph is S-019 → S-020 → S-021 → S-022 → S-012 → S-014.

Changed internal interfaces in S-019 include structured AST SQL results, sealed
`sqlshape.Fragment` query plans, fragment-based pagination, and descriptor-bound
literal restore-count queries. Query parameters and result contracts remain
unchanged. SQL text may contain different insignificant whitespace and explicit
PostgreSQL escaped string literals. No public API, schema, or stored-data
migration is intended. Final readiness remains **NO-GO**.

### Corrective checkpoint: S-019 complete

Both architecture checkers now consume the Go source inspector's structured SQL
declarations. The lexical Go extraction path is removed. Bounded assignments,
branches, constants, scoped names, builders and execution calls are qualified;
unresolved construction produces an explicit diagnostic. Projection query plans
and pagination now use closed read fragments; raw relation holes and constructor
escapes fail. Timeline source locking is visible under a temporary explicit owner
classification, to be separated from provider dependencies in S-021.

Retained evidence lives under `.cartulary/test-results/<run-id>/`. Commands and
selections used for this checkpoint:

- `make test-slice OWNER=harness.command_surface`: owner-boundaries row passed
  in `tcr-s019-mechanics-20261109-06` (exit 0); source-inspection row passed in
  `tcr-s019-ast-20261109-03` (retained summary: pass).
- `make test-slice OWNER=module.projections`: SQL declaration ownership passed
  in `tcr-s019-sql-20261109-07` (retained summary: pass); compiled query, query
  shape/keyset, Host and Identity SQL rows passed in
  `tcr-s019-query-20261109-03` (exit 0).
- `make test-slice OWNER=platform.viewquery`: closed SQL fragments and locator
  predecessor rows passed in `tcr-s019-fragments-20261109-02` (exit 0).
  Pagination keyset assertions passed in `tcr-s019-keyset-20261109-02` (exit 0).
- PostgreSQL-backed query/load parity and restore rows passed in
  `tcr-s019-db-20261109-01`; Evidence attached create/patch passed in
  `tcr-s019-evidence-20261109-01`; Links bundle isolation passed in
  `tcr-s019-links-20261109-01`. Each exited 0 with completed cleanup.
- Indicator projection delegation and its migrated fixture writer passed in
  `tcr-s019-indicator-20261109-01` (exit 0). `make format` passed in
  `tcr-s019-format-20261109-14` (exit 0).

Earlier failed qualification runs are retained. The latest global backend run,
`tcr-s019-boundary-20261109-04`, failed (exit 2, product assertion failure):
maintenance count permissions belong to S-020 and unresolved generic SQL belongs
to F-026/S-022. These are blocking follow-on obligations, not exclusions or passes.
Generated selector outputs remain pending reconciliation. Next: S-020 exact
package/declaration grants; F-024 and F-025 remain open and readiness is **NO-GO**.

### Corrective checkpoint: S-020 complete

Projection storage/query and affected Records/source-read grants now name exact
packages. Statement exceptions use package + declaration + normalized SQL shape;
the old filename statement form has no compatibility path. Count maintenance has
ten exact descriptor-table statements, without a blanket runtime-package grant.
Source mutation and read packages retained only their reviewed authority. Links
merge/revision packages and Records reporting reads have explicit package grants.
Test-support grants remain separate. Admission rejects obsolete affected filename
grants, missing package grants and stale statement bindings. Regression fixtures
exercise both checkers for rename/split parity, peers, descendants, wrong provider
tables, writes/locks and altered statement shape or declaration.

- `make test-slice OWNER=harness.command_surface
  ROWS=harness.command_surface.behavior.owner_boundaries` passed (114 cases,
  exit 0), run `tcr-s020-mechanics-20261109-05`.
- `make generate` passed (exit 0), run `tcr-s020-generate-20261109-01`, refreshing
  the changed schema validator and authored test-family projections. Generated
  validator dependency changes are mechanical output from the canonical producer.
- `make source-check CHECK=schema` passed (exit 0), run
  `tcr-s020-schema-20261109-02`. The preceding failed run exposed stale generated
  topology/validators and remains retained.
- `make format` passed (exit 0), run `tcr-s020-format-20261109-01`.
- `make lint CHECK=backend-boundary`, run `tcr-s020-boundary-20261109-02`, remains
  failed (exit 2, product assertion failure). Its diagnostics contain only
  `sql_construction_unresolved`; no table-access, stale binding or package-grant
  violations remain. Those construction failures belong to F-026/S-022.

Earlier qualification failures are retained. All run roots are under
`.cartulary/test-results/`. Next prerequisite: S-021 must replace recursive root
scanning with provider dependency provenance and remove mixed owner capabilities.
Global acceptance remains **NO-GO**; passing policy fixtures alone do not close
F-024 or the later coverage/handoff checkpoints.

### Corrective checkpoint: S-021 complete

Provider implementation and contributed-reader entry packages now drive a complete
repository-local import closure. Every reachable production Go file is inspected,
including generated and build-constrained files. Missing/empty packages, linked
files/directories, Go link directives, unclassified native code and unadmitted
foreign/standard-library imports fail explicitly. Shared dependencies retain each
provider's identity and must satisfy each consumer independently. Source coverage
is required for every active provider; storage coverage alone cannot satisfy it.
Unassigned SQL, stale declarations/package capabilities and source-owner bypass
roles fail. The old recursive source-root ceiling and all `source_owner:` roles
are removed.

Changed internal capabilities:

- Timeline read-only `sourcerepository` and owner-only `sourcelocking` share the
  immutable selection, scanner and envelope interpretation. The owner keeps the
  original source/envelope locks and borrowed transaction. A central SQL rule
  confines explicit Timeline source locks to the locking package.
- Evidence Timeline facts live in `evidence/timelinefacts`; Links facts live in
  `links/facts`, sharing public `links/vocabulary` and party-assignment metadata.
  The old aggregate fact-reader APIs were removed and all callers migrated.
- Records envelope data, scanners and unlocked reads live in
  `records/enveloperead`. Mutation APIs retain their existing lock semantics.
  Assessment projection fact adapters are separated from aggregate assembly.
- The unchanged PostgreSQL `DB` method set moved to `postgres/dbport`, with all
  consumers migrated and no alias at the old path. This prevents interface users
  from importing connection/configuration implementation dependencies.
- Go analysis resolves imported immutable query definitions and rejects ambiguous
  duplicate constant/function constructions. Diagnostic templates and value-only
  fragments do not create relation authority. Fragment fixtures also reject split
  share locks, SELECT INTO, COPY/CALL, and split TABLE relation construction.

Retained runs below are under `.cartulary/test-results/`, all successful commands
exited 0. Service runs completed every declared cleanup operation:

| Command / exact selection | Run ID | Result |
| --- | --- | --- |
| `make test-slice OWNER=harness.command_surface` owner-boundaries + Go source-inspection rows | `tcr-s021-mechanics-20261109-02` | pass |
| Same owner, `harness.command_surface.behavior.owner_boundaries`, final capability/fragment fixtures | `tcr-s021-mechanics-20261109-05` | pass, 116 cases |
| `make test-slice OWNER=module.projections` SQL ownership, assembly/catalog, single-connection query/load parity | `tcr-s021-sql-20261109-04` | pass |
| Same owner, `module.projections.architecture.sql_declaration_ownership`, final closure policy | `tcr-s021-policy-20261109-02` | pass |
| `make test-slice OWNER=module.records ROWS=module.records.envelope.store,module.records.envelope.destructive_lock` | `tcr-s021-records-20261109-02` | pass |
| `make test-slice OWNER=module.timeline` evidence-fact mapping + successful create/patch source/revision assertions | `tcr-s021-timeline-20261109-02` | pass |
| `make test-slice OWNER=module.links` structured links/tags persistence + party-assignment snapshot facts | `tcr-s021-links-20261109-01` | pass |
| Same owner, closed-vocabulary exact-round-trip row at its new package | `tcr-s021-vocabulary-20261109-01` | pass |
| `make test-slice OWNER=module.evidence ROWS=module.evidence.store.attached_evidence_create_and_patch_aa33ea0168` | `tcr-s021-evidence-20261109-01` | pass |
| `make test-slice OWNER=module.assessments ROWS=module.assessments.store.assessment_projection_query_remains_owner_backed_172051cb0d` | `tcr-s021-assessment-20261109-01` | pass |
| `make build COMPONENT=server` (planned embedded-web prerequisites included) | `tcr-s021-build-20261109-01` | pass |
| `make format` | `tcr-s021-format-20261109-09` | pass |

Failed qualification runs remain retained, including obsolete caller imports,
interface signatures, and the incomplete TABLE fragment fixture. No runtime
regression was observed in the selected source-reader/locking/projection tests.
The latest global gate, `make lint CHECK=backend-boundary`, run
`tcr-s021-boundary-20261109-02`, failed (exit 2, product assertion failure) with
63 `sql_construction_unresolved` diagnostics and no other boundary violations.
F-026/S-022 must qualify those generic SQL transports, contributed statements,
query builders and maintenance/catalog operations. No exclusion or permissive
fallback is authorized. Selector regeneration remains due for the vocabulary
package move. Next: S-022, then S-012 and S-014; readiness remains **NO-GO**.

### Corrective checkpoint: S-022 complete

F-026 is closed by explicit SQL capabilities, not by exclusions. The global
backend gate now accepts every affected production query without an unresolved
construction bypass. Both checkers use structured AST declarations; the projection
policy no longer accepts a legacy declaration-source response. Native acquired,
prepared, batch and bulk-copy execution boundaries retain type identity through
aliases, imported fields and embedded interfaces. SQL formatting accepts only
qualified `%s`/`%%` with exact arity; unsupported formatting remains visible in
SQL-producing declarations even without a local sink.

Changed interfaces and preserved behavior:

- `sqlshape.Fragment` replaces untyped structural query contributions in query
  compilers, source export/read ports and Recovery snapshot bindings. Sealed
  `Statement` values carry complete mutation/import/maintenance commands.
  Literal relations and token-separated fragments prevent arbitrary construction.
- Recovery and harness reset use `maintenancequery`, generated from the canonical
  Recovery catalog and ownership manifest through Make. Aggregate truncate/lock
  operations validate the complete relation set and retain their atomic SQL.
  Restore insertion excludes derived storage; protected metadata/control tables
  retain distinct operations. Generated SQL stays in ordinary ownership scans.
- Recovery FK validation consumes catalog descriptors and qualified identifiers/
  comparison operators rather than executing query strings returned by SQL.
  Existing MATCH SIMPLE/FULL meaning, generated-column verification, fencing,
  borrowed transaction identity and rollback behavior are preserved.
- Artifact, Task/Decision, Indicator, Network Flow, Reporting and shared source
  query builders now use bounded templates. Network Flow keyset construction
  shares one typed implementation. Host/Identity hydration uses UUID array
  parameters instead of generated placeholder lists; ordering/results are retained.
- Transparent DB transports have exact package/declaration grants only for an
  unchanged SQL parameter. Session controls have exact complete-command grants.
  Neither capability is available to a provider dependency. New constructors,
  stale grants and changed control statements are rejected.
- The final permission audit replaced six remaining Reporting production SQL
  filename grants with exact root-package read capabilities. Descendants and
  write/lock operations remain denied; separate test-support grants remain.
- Core 01 and the Harness NLSpec distinguish frozen-catalog physical maintenance
  from provider authority. The [SQL enforcement guide](../guides/sql-ownership-enforcement.md)
  records construction limits, provider/shared-reader admission, removed internal
  paths and migration notes. No public API, schema, stored-data or deployment
  migration is required.

All runs below live under `.cartulary/test-results/`. Target summaries pass,
cleanup outcomes are completed and retained-secret scans pass. Outer exits were
observed as zero for final gates, build, generation and the first six behavior
runs; for Entities/Reporting/Indicators/Links and the earlier four-row projection
run, the terminal artifacts were independently recovered after the native-session
completion output was unavailable. Those results are not represented as separately
captured outer-exit records.

| Command (exact ROWS where applicable) | Run ID | Result |
| --- | --- | --- |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | `tcr-s022-mechanics-20261109-09` | pass; AST/formatting/admission fixtures |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries` | `tcr-s022-mechanics-20261109-10` | pass; final Reporting grant fixtures |
| `make lint CHECK=backend-boundary` | `tcr-s022-boundary-20261109-08` | pass; final production policy |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | `tcr-s022-policy-20261109-04` | pass; final structured protocol/closure |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership,module.projections.assembly.catalog_and_source_ownership_927c3f08ce,module.projections.query.compiled_surface_iteration3,module.projections.query.contract_shape_and_keyset_59528aa56d` | `tcr-s022-policy-20261109-02` | pass; query, keyset and assembly behavior |
| `make test-slice OWNER=platform.viewquery ROWS=platform.viewquery.unit.closed_sql_fragments,platform.viewquery.unit.maintenance_sql_capabilities` | `tcr-s022-fragments-20261109-02` | pass |
| `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.reset_preserves_crypto_format_identity,module.recovery.integration.structural_restore_completes_with_independent_co_1bd8029ccb,module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c` | `tcr-s022-recovery-20261109-01` | pass; includes referential-integrity rejection |
| `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.pagination_recovery,module.networkflow.integration.bounded_graph_contributor_pipeline` | `tcr-s022-networkflow-20261109-01` | pass |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.managed_service_assertions_for_runtime_reset_rec_79940f98ab` | `tcr-s022-reset-20261109-01` | pass; Go reset integration, no browser |
| `make test-slice OWNER=module.artifacts ROWS=module.artifacts.surface_contracts.artifact_surface_registry_recognizes_all_eight_c_778e06f594,module.artifacts.source_mutations.managed_service_assertions_for_artifact_source_c_a67a87fbf9` | `tcr-s022-artifacts-20261109-01` | pass |
| `make test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.source_workbook.managed_service_assertions_for_task_requests_and_9063e4f330` | `tcr-s022-tasks-20261109-01` | pass |
| `make test-slice OWNER=module.entities ROWS=module.entities.store.explicit_merge_store_behavior_repoints_live_ment_0560122310` | `tcr-s022-entities-20261109-01` | pass; Host/Identity source hydration |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.provider_output_validate_and_fields_1d0b913cab,module.reporting.integration.managed_service_assertions_for_evidence_provider_0b8acd1609` | `tcr-s022-reporting-20261109-01` | pass |
| `make test-slice OWNER=module.indicators ROWS=module.indicators.store.indicators_persist_as_canonical_indicator_record_3998750e91` | `tcr-s022-indicators-20261109-01` | pass |
| `make test-slice OWNER=module.links ROWS=module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1` | `tcr-s022-links-20261109-01` | pass; source-bundle round trip and rejected-write atomicity |
| `make build COMPONENT=server` | `tcr-s022-build-20261109-02` | pass |
| `make generate` | `tcr-s022-generate-20261109-03` | pass; aggregate maintenance/catalog/schema projections |
| `make format` | `tcr-s022-format-20261109-11` | pass |

Qualification failures are retained. `tcr-s022-mechanics-20261109-01..03`,
`boundary-20261109-01..03`, `build-20261109-01` and
`policy-20261109-01` exposed incomplete typed callers, fixture admission,
unsupported LISTEN, and the previously unassigned Timeline bulk-copy fixture
writer. The corrected storage declaration has explicit Timeline ownership.
`format-20261109-03` and `06` rejected unsorted/unregistered authored catalog
inputs before successful generation. Final hardening runs
`tcr-s022-mechanics-20261109-08` and `tcr-s022-boundary-20261109-06` rejected the
existing combined reset controls because admission initially required exactly
one control statement; admission now requires the explicitly granted complete
control-only command and rejects added reads or altered controls. No failure was
hidden by an exemption, weakened behavioral assertion or same-input rerun.

Residual limits are deliberate finite analysis and explicit external-library
trust boundaries, documented in the guide. There is no outstanding construction
or policy diagnostic. Next prerequisite: S-012 must reconcile routing/generated
artifacts and compatible evidence; S-014 must validate documentation and final
handoff before readiness changes from **NO-GO**.

### Corrective checkpoint: S-012 complete

The corrective coverage ledger below supersedes the historical S-012 closure for
F-023–026. No identified finding, selector, capability exemption or changed
behavior lacks an explicit disposition. Qualification fixture counts are not
used as evidence of product semantics or complete specification conformance.

| Finding / obligation | Positive and negative qualification | Production / behavior evidence |
| --- | --- | --- |
| F-023 complete Go expressions | `TestAnalyzeSQL` and `TestSQLImportedDatabaseTypes`: leading/cross-file/imported constants, grouped declarations, parentheses, shadowing, assignment/`+=`, finite branches/helpers, builder escape, unresolved operands, format indexes/arity, acquired/prepared/batch/copy and embedded/imported DB types. Shared SQL fixtures cover UNION, CTE/nested scopes, qualification, all row locks, table locks, multiple statements and unresolved relations. | Final H/A runs in S-022; backend and projection production gates pass. S-021 Timeline owner locking and read mapping remain compatible. Compiled projection/keyset and Network Flow behavior pass after typed construction. |
| F-024 complete dependency coverage | Closure fixtures traverse aliased sibling/transitive and build-constrained generated packages; reject missing, linked/native and unadmitted foreign dependencies. Shared helpers satisfy every recipient; write/lock laundering and the retired source-owner role fail. Every active provider requires source coverage. | Final projection gate checks the complete active catalog. S-021 Records, Timeline, Evidence, Links and Assessment tests cover read meaning, borrowed transaction and owner-lock separation. S-022 Links bundle round trips qualify the subsequently typed common export port. |
| F-025 coherent package/declaration grants | Both checkers accept same-package renames/splits. Peers, descendants, wrong-provider tables and query-package writes/locks fail. Exact statement grants retain literals, parameters and complete SQL shape; stale grants/bindings fail. Reporting root read grants have equivalent rename/descendant qualification. | Final backend policy and projection ownership pass. Exact maintenance counts/restore enumeration remain separate from provider identities; test-support authority remains separate. |
| F-026 closed generic builders/maintenance | Closed fragment/command tests reject dynamic relation holes, split operations, unqualified constructors and post-conversion SQL. Maintenance tests reject unknown/injected/protected/partial/duplicate relation sets and restoring derived projections. Forwarding/control tests deny altered construction, changed commands, peers and stale grants. | S-022 Recovery integrity/restore, harness reset, Artifact, Task/Decision, Host/Identity, Indicator, Reporting, Links, Network Flow and projection query selections pass. No arbitrary dynamic-relation or opaque transport bypass was added. |

Catalog reconciliation includes the source-inspection selector's new AST/type
cases, closed-fragment and maintenance-capability rows, moved Links vocabulary
selector, and existing projection ownership routing. `make generate` produced the
maintenance Go catalog, schema validators and affected catalog projections from
authored inputs. Generated code was not hand-edited. Final scratch regeneration
found no drift; schema admission checks the current v4 policy. Executable inputs
remain canonical artifacts and source, with no specification/README dependency.

The following exact commands exited zero; their retained target summaries and
cleanup/secret-scan results pass:

| Command | Retained root under `.cartulary/test-results/` |
| --- | --- |
| `make source-check CHECK=catalog` | `tcr-s012-corrective-catalog-20261109-01` |
| `make source-check CHECK=schema` | `tcr-s012-corrective-schema-20261109-01` |
| `make source-check CHECK=generated` | `tcr-s012-corrective-generated-20261109-01` |

Compatible historical browser/HTTP evidence is reused: S-008
`tcr-s008-http-20261109-01` and `tcr-s008-browser-20261109-01`, S-009
`tcr-s009-http-20261109-01`, and S-010 unit/browser roots recorded below. Their
rendered/HTTP assertions were preserved. The common DB port move retains its
method set; current application builds and affected owner tests qualify the
internal migration. These historical passes do not substitute for the new SQL
counterexample fixtures or claim new browser execution.

Every observed corrective failure has a later changed-input qualification in
S-019–022. No unresolved runtime regression or policy diagnostic remains. There
is no full-suite, performance, visual or release acceptance claim. Next: S-014
must finish Markdown validation, final source/diff inspection and the handoff;
readiness remains **NO-GO** until that checkpoint.

### Historical corrective checkpoint: S-014 complete — superseded GO

**Historical recommendation: GO, subsequently superseded by the F-027–029 NO-GO review. The current graph and final S-014 checkpoint govern readiness.** F-023, F-024, F-025 and the additional F-026 are closed with
current compatible evidence. S-005/S-006 corrective acceptance is restored;
S-018–022 and the reopened S-012/S-014 all satisfy their exits. Earlier NO-GO and
completion statements are historical checkpoints, not competing current status.

| Finding | Final disposition | Decisive evidence |
| --- | --- | --- |
| F-023 | CLOSED | Full-expression and unsupported-construction regressions pass; native driver/type and format fixtures pass; shared production backend/projection analysis passes. |
| F-024 | CLOSED | Complete dependency closure, consumer-specific shared authority and source coverage pass; sibling/transitive extraction and write/lock laundering fail; changed fact-reader behavior passes. |
| F-025 | CLOSED | Package rename/split parity passes; wrong-provider, peer, descendant, operation and exact-statement mutations fail; current production policy has no affected filename-grant contradiction. |
| F-026 | CLOSED | Typed query/command ports and generated maintenance capabilities pass construction/policy qualification and affected source, keyset, restore-integrity and reset behavior checks. |

Final review covered authored/generated separation, internal interface migrations,
removed files and replacement capabilities, retained transaction/ordering/error
semantics, production grant scopes, selector changes and the original browser/HTTP
migrations. The 520 tracked changes and 25 untracked status paths at handoff
include the original work; much of the expanded footprint is the DB-port import
move and generated schema validators. New/untracked directories must be included
when the maintainer stages the changes. There are no staged changes, commits,
merges or deployments from this effort. Dependency lock files were not modified.

Final non-product validation:

- `make format` exited zero, run root
  `.cartulary/test-results/tcr-s014-corrective-format-20261109-01`.
- `make lint-markdown` passed over 125 files, with no fixes applied. It is rerun
  after this completion record is written.
- `git diff --check` passed. Source checks and final product/policy selections,
  exact commands, retained roots and failure history are recorded in the S-021,
  S-022 and corrective S-012 checkpoints above.
- The final source-only adjustment groups Timeline locking imports according to
  project style; it does not change the AST dependency graph or runtime behavior.
  Earlier compatible source-reader and production gate evidence is reused.

The [SQL ownership enforcement guide](../guides/sql-ownership-enforcement.md)
provides the supported construction contract, procedure for adding providers and
shared read capabilities, closed maintenance/control/transport capabilities,
removed compatibility paths and validation entry points. Internal query interfaces
and DB-port imports require atomic adoption in this commit. There is no public
API, database schema, stored-data or deployment migration. The prior Records
locking correction and substantive browser-to-HTTP assertions remain intact.

No required targeted check was skipped. A full suite, visual suite, release run
and `agent-finalize` were not required by this change and were not launched.
Compatible historical browser/HTTP evidence is identified in corrective S-012;
this is not a new browser execution or exhaustive specification-conformance claim.
Finite analysis limits and explicit foreign-library trust boundaries remain
intentional and documented. Unsupported construction requires a structural repair
or separately qualified capability, never a permissive fallback. Future SQL
features and external driver boundaries require their own admission fixtures.

The current corrective execution graph is complete. Handoff is ready for human
review and staging; commit, merge and deployment remain outside execution scope.

### Historical implementation graph (closure superseded where identified above)

**TCR-REQ-021 — Dependency semantics.** This is the single authoritative
execution graph. Each slice is a separate workstream, executed in table order.
After completing each workstream, update this tracker with its evidence and exit
status before beginning the next. Retirements require qualified surviving
obligations and selection. Any newly confirmed blocking defect receives a separate
corrective slice before S-014. Missing evidence is work to establish, not a pass.

The earlier grouped WF-01–WF-09 graph is superseded by the per-slice graph below.
Historical WF references identify the original investigation groups only.
WF-00 is satisfied by the revalidated execution baseline in E-026.

| Slice / workstream | Prerequisite | Status | Goal and binary exit | Validation |
| --- | --- | --- | --- | --- |
| S-015 Markdown validation coverage | WF-00 | DONE | Include this tracker in the existing direct Markdown lint; review its result and diff. | V-014 |
| S-001 Full inventory and purpose review | S-015 | DONE | Reconcile source registrations, catalog cases, helpers, fixtures, noncanonical contexts, producer/vendor material, and accountable dispositions. | V-001 |
| S-002 Architecture authority and acceptance identities | S-001 | DONE | Adopt durable owner rules replacing seven missing decisions; repair duplicate acceptance IDs and semantic references. | V-003/V-014 |
| S-013 Recovery specification cleanup | S-002 | DONE | Remove duplicated grammar; map all seven commands and per-invocation assertions without changing valid CLI behavior. | V-013/V-014 |
| S-003 Scanner correctness | S-013 | DONE | Qualify explicit production roots, empty/error handling, forbidden imports and exemptions. | V-002 |
| S-016 Current generated-contract boundaries | S-003 | DONE | Replace retired monolithic-package policy with complete current-family consumer policy and effective fixtures. | V-002/V-011 |
| S-004 Revisions selection decision | S-016 | DONE | Retain the thin Revisions-owned witness over the shared qualified detector; prove focused selection. | V-002/V-011 |
| S-005 Backend capabilities | S-004 | DONE | Reconcile exact package permissions; retain closed local policies, global topology and SQLC export protection with qualified fixtures. | V-003/V-004 |
| S-017 Linked-record envelope ownership | S-005; discovered during S-006 | DONE | Move direct envelope lookup/locking behind Records-owned capability; preserve ordering, transaction and linked-refresh behavior. | V-003/V-004; focused Records/Links validation |
| S-006 Projection SQL assurance | S-017 | DONE | Share SQL parsing, assign all SQL declarations to provider/maintenance authority, and preserve descriptor/table parity. | V-004 |
| S-007 Frontend semantic assurance | S-006 | DONE | Qualify dependency resolution and replace spelling assertions with behavior while retaining distinct semantic/type checks. | V-005 |
| S-008 Incident, discovery and dispatch evidence | S-007 | DONE | Qualify assembled HTTP replacements and retained rendered landing observations before browser retirement. | V-006/V-007/V-009 |
| S-009 Coordination HTTP workflow | S-008 | DONE | Qualify seven patches, public demotion, same-session authorization and rejection safety before browser retirement. | V-008 |
| S-010 Selector deduplication | S-009 | DONE | Remove pure duplicates and preserve distinct mounted-state observations. | V-009 |
| S-011 Supporting purpose, placement and cost | S-010 | DONE | Close evidence-backed keep/move/remove and placement decisions; distinguish measured from unmeasured cost. | V-010/V-011 |
| S-012 Coverage and reliability closure | S-011 | DONE | Close identified obligation gaps and blocking defects; preserve failure evidence and limits on stability claims. | V-012 |
| S-014 Final validation and handoff | S-012 and every required corrective slice | DONE | Reconcile compatible results, routing, all acceptance criteria, residual limits and final handoff. | V-011/V-014 |

## 7. Adopted Remediation Decisions

The maintainer accepted the complete plan and then requested implementation.
Authorization is granted; completion and test success are separate facts. The
following decisions supersede conflicting initial recommendations. Existing
finding and contract IDs remain stable.

| Findings / slice | Remediation and affected areas | Rationale / durable benefit | Compatibility and unresolved risk | Completion criterion |
| --- | --- | --- | --- | --- |
| F-001 / S-003 | Implementation/tests: explicit cmd/internal discovery, shared Go import inspection, policy-specific exclusions and positive selection. | Eliminate false-green traversal and duplicated scanner mechanics. | No public migration; hidden violations must be repaired rather than excluded. | Required-root, empty/error, forbidden and allowed fixtures pass. |
| F-002 / S-004 | Tests/docs: retain the thin Revisions owner selector using the shared evaluator. | One-owner-per-case routing makes a tiny owner-specific witness cleaner than cross-owner routing machinery. | No alias or selection loss; outright deletion would weaken Revisions-focused checks. | Both owner selections detect forbidden support imports. |
| F-003/F-004/F-018 / S-002/S-005 | Spec/implementation/tests: package capabilities by default, closed local allowlists, central cross-module rules, shared parsing, retained SQLC export and unique architectural witnesses. | Cohesive policy owners; allowed filename changes do not alter package authority. | No public migration; grants must not widen to entire subtrees accidentally. | Approved rename/split and forbidden peer/private-provider fixtures; exact obligation/selection matrix. |
| F-005/F-018 / S-006 | Implementation/tests/policy: existing SQL parser, complete source scopes, package defaults and declaration bindings in mixed packages; retain assembled descriptor and table-set parity. | Newly extracted SQL cannot evade review; no second authoritative provider registry. | Internal inputs migrate atomically; no data migration. Silent source or SQL-form gaps remain a blocking risk. | New/unassigned SQL, read/write/lock/CTE/qualified-name/dynamic-relation fixtures and all table equalities pass. |
| F-006/F-007/F-018 / S-007 | Implementation/tests: equivalent dependency rules share mechanics; semantic/type/transport checks stay distinct; replace spelling with generated-consumption and claim-gating behavior. | Resists bypasses without rejecting harmless formatting/aliases. | No UI or wire change; import presence alone cannot replace actual behavior. | Supported aliases/re-exports/dynamic/type-only cases, bypass fixtures and consumer behavior pass. |
| F-008/F-009/F-010 / S-008 | Tests/routing: assembled exact incident errors, complete discovery profiles and dispatch cases; preserve one compatible rendered zero-member landing witness. | Observe each contract at its appropriate boundary without unnecessary browser setup. | No route change; retire rows only after complete replacement, using catalog migrations. | Exact envelopes/profile claims/root-descendant-template-outside matrix and retained DOM observation pass. |
| F-011 / S-009 | Tests/routing: assembled seven-surface patch/read/version workflow, API role demotion, same session and unchanged denied-write effects. | Removes browser-only fixture burden while retaining current authorization assurance. | No field/role/history migration; SQL-only demotion or partial patches are insufficient. | Complete §5.3 matrix passes before browser retirement. |
| F-012 / S-010 | Tests: remove repeated selector equality and unchanged-state attachment assertions; keep distinct rendering observations. | Less noise without losing screen/state evidence. | No selector/golden change; blanket deletion loses mounted-state assurance. | Selector units and narrowed browser witness pass. |
| F-013/F-014/F-017/F-018 / S-001/S-011/S-012 | Tests/metadata/docs: complete registration/purpose/obligation review; correct classification; use typed, compatible diagnostics for concrete cost/reliability questions. | Explainable portfolio; supporting tests retain independent engineering purpose. | No invented speed/stability claims or new performance gate; discovered blockers receive separate fixes. | No unexplained registration, ownership or blocking coverage gap; evidence-backed retention/removal and accurate measurement limits. |
| F-015 / S-013 | Spec/docs: remove five-command restatement; Core owns grammar; each invocation maps uniquely while workflow cases may exercise several commands. | Single authority avoids repeated list drift and unnecessary workflow splitting. | Seven-command CLI already exists; no aliases or redesign. | Seven-command assertion/routing map and owner prose agree. |
| F-016/F-020 / S-002 | Spec/docs: replace seven dangling decision owners with durable Core/Graph rules; resolve duplicate AC-483/537/538/539/540 identities by semantic reference. | Accessible authority and unambiguous acceptance mapping. | No runtime compatibility; never globally replace ambiguous IDs without meaning review. | Adopted rules, unique definitions and all affected references resolve. |
| F-019 / S-015 | Tooling/docs: include *-refactor-tracker.md in direct Biome lint; retain intentional autofix behavior. | Canonical validation covers the controlling artifact without resurrected wrappers. | Historical tracker-only write limit is superseded; review actual lint diff. | Pattern includes this path and canonical lint passes. |
| F-021 / S-016 | Spec/implementation/tests/policy: discover contract families from the existing registry and explicitly assign approved consumer packages; fail unassigned families and forbidden fixture use. | Current-family enforcement cannot silently check a retired package. | Replace old guard/selection atomically; no monolithic compatibility shim. | Complete family disposition and allowed/forbidden/new-family fixtures pass. |
| F-022 / S-017 | Implementation/tests/policy: Records owns envelope identity locking; assembly adapts its DTO to the Projections consumer port. | Restores cohesive source authority and makes transaction/lock dependencies explicit. | Internal constructor change, no public/data migration; unresolved direct locking would couple Projections to authoritative storage and undermine source isolation. | Same ordering, incident filtering, borrowed transaction, failure/missing-endpoint handling and linked projection/history updates pass. |

Each workstream records exact entry evidence, commands, artifacts and rollback
notes at its checkpoint. Rollback restores authored routing, implementation and
generated projections together; original failure records remain retained.
New production behavior changes beyond this adopted plan remain separate owner
amendments and cannot be hidden in structural consolidation.

## 8. Validation Plan

### 8.1 Validation interface and canonical commands

**TCR-REQ-022 — Validation boundaries.** Future validation MUST use public Make
targets from the repository root and current test-operations guidance.
Selection/plan success MUST NOT count as execution success. Each slice MUST
select the smallest checks covering changed behavior and directly affected
dependencies, then stop when those checks and their declared cleanup pass.
Broader execution requires a concrete uncovered risk or an explicit broader
acceptance request. Complete-effort accounting MUST compare compatible evidence
without treating that accounting as permission to rerun complete owners.
Explicit-only, process/order, security, and measurement obligations MUST remain
explicit when the requested acceptance scope includes them. Generation and
finalization MUST be conditional on actual affected outputs or requested
maintenance; an implementation edit alone does not require them.

Each invocation record MUST contain its exact resolved command, fresh
literal run identity, source baseline, selected rows/cases, runtime/fixture
profiles, environment, outer exit, terminal result, artifact path, and remaining
applicability limits. Execution is authorized by TCR-REQ-002, subject to the workstream gates.

Baseline exact selectors for high-priority cases are recorded below. A later
session MUST revalidate them through current catalog/Make selection before
execution. `make task-guide OWNER=<owner>` and
`make target-plan TARGET=<target> DETAIL=rows JSON=1` are selection
templates; any substituted command MUST be recorded literally. Planned-target
inputs belong in the supported INPUTS object, not as OWNER/ROWS arguments to
target-plan. Omitted ROWS selects the owner's non-explicit-only scope; it MUST
NOT be treated as the preferred narrow selection.

| Selector key | Baseline catalog row ID (see migrations for retired identities) |
| --- | --- |
| R-01 | `harness.generated_artifacts.unit.runtime_go_files_do_not_import_internal_testutil_f162342f15` |
| R-02 | `module.revisions.support_unit.test_support_production_import_guard_6a4b2c1d9e` |
| R-03 | `harness.generated_artifacts.unit.sqlcimports_stay_behind_persistence_adapters_61a4cd5750` |
| R-04 | `module.projections.architecture.boundaries_and_source_ownership_9b8de61f81` |
| R-05 | `module.projections.assembly.catalog_and_source_ownership_927c3f08ce` |
| R-06 | `module.incidents.browser_support.supports_route_owned_incident_validation_errors_c277ce45ee` |
| R-07 | `app.server.browser_support.supports_zero_membership_extension_discovery_and_2d096f2dcd` |
| R-08 | `module.incidents.browser_support.supports_reserved_family_dispatch_precedence_pro_d831d613fa` |
| R-09 | `module.workbook.browser_stateful.verify_coordination_rows_can_be_queried_and_edit_77783541c2` |
| R-10 | `module.auth.browser_support.supplemental_browser_verifies_ordinary_shell_sel_f8a9da1957` |
| R-11 | `app.server.unit.extension_discovery_transport_7f6ad80724` |
| R-12 | `app.server.integration.extension_discovery_session_sliding_7f6ad80724` |
| R-13 | `app.server.integration.unclaimed_reserved_extension_families_return_can_57db0f8a45` |
| R-14 | `module.workbook.support_integration.workbook_surface_create_query_and_patch_contract_ae5c34d894` |

R-keys preserve the initial validation design. R-06/R-07/R-08/R-09 are retired;
resolve them through the authored row-migration catalog. Use current exact rows
from the execution checkpoints and Make discovery for new invocations. R-keys
and retired IDs MUST NOT be sent to Make.

| Validation / layer | Findings / contracts / slices | Stage and exact command or bounded selection procedure | Required environment / expected evidence | Failure consequence |
| --- | --- | --- | --- | --- |
| V-001 / inventory and classification | F-013/F-017; C-001/C-010; S-001 | Planning read-only Git/file/JSON inspection; later `make task-guide DETAIL=coverage OWNER=<reviewed-owner> JSON=1` for each affected owner. | No service execution for static accounting; preserve exact catalog/source reconciliation and unresolved registration list. | Inventory or classification remains partial; dependent retirement blocked. |
| V-002 / scanner correction and subset | F-001/F-002; C-002; S-003/S-004 | Pre-correction failure witnesses, then `make test-slice OWNER=harness.generated_artifacts ROWS=R-01` and `make test-slice OWNER=module.revisions ROWS=R-02` with resolved IDs. | Go readiness; both intended roots selected, empty scope rejected, forbidden/exempt imports qualified. | No scanner completion or subset retirement. |
| V-003 / backend policy | F-003/F-004/F-016; C-003; S-002/S-005 | `make lint CHECK=backend-boundary`; exact affected local guard and backend detector-assurance rows from current owner guides. `make harness-contract SUITE=contract` is permitted only when a documented uncovered registration/process contract requires that bundle. | Canonical checker and allowed/forbidden/rename fixtures; authority matrix. | Keep original guards; unresolved ownership remains blocked. |
| V-004 / SQL and projection | F-005/F-018; C-004; S-005/S-006 | `make test-slice OWNER=harness.generated_artifacts ROWS=R-03`; `make test-slice OWNER=module.projections ROWS=R-04,R-05`; V-003's smallest affected detector selection if parser changes. | Go/static dependencies; exported SQLC leak cases, quoted/CTE/lock cases, descriptor and new-source detection. | No parser or source-check retirement. |
| V-005 / frontend architecture | F-006/F-007/F-018; C-005; S-007 | Exact affected rows from web.architecture, web.workbook, and module.imports guides; `make test-slice OWNER=web.architecture ROWS=web.architecture.support.enforce_frontend_import_boundaries_apps_web_must_cc56151788` plus applicable detector assurance rows. | Frontend dependencies; Vitest node-unit/type and policy checks; alias/dynamic/type-only and bypass fixtures. | Preserve original rule until semantic equivalence established. |
| V-006 / incident HTTP errors | F-008; C-006; S-008 | Exact incident admission rows; new/existing assembled HTTP witness under module.incidents; compare resolved R-06 while it exists. | PostgreSQL/application fixtures for route witness; browser stack only for original comparison. | Admission-unit pass alone cannot permit removal. |
| V-007 / discovery and dispatch | F-009/F-010; C-007; S-008 | `make test-slice OWNER=app.server ROWS=R-07,R-11,R-12,R-13` and resolved module.incidents R-08; add exact missing assembled cases before retirement. | Bind claims and session profile; shape/order, zero membership, pagination details, root/descendant/base/outside dispatch. | Retain browser API assertions with uncovered obligations. |
| V-008 / coordination | F-011; C-008; S-009 | `make test-slice OWNER=module.workbook ROWS=R-09,R-14` before changes; qualify exact replacement rows afterward. | Assembled application and persisted state; seven-surface matrix and API role transition. | Partial create/query or two-surface patch evidence is insufficient. |
| V-009 / actual browser observations | F-009/F-010/F-012; C-009; S-008/S-010 | Exact auth selector unit, R-10, and incident-selection browser rows from current owner guides; only affected browser cases. | Real rendering/interaction; empty directory, mounted nodes, and retained UI obligations. | Builder equality or API success cannot replace DOM evidence. |
| V-010 / support placement and cost | F-013/F-014/F-018; C-001/C-010; S-011 | Exact affected owner/operation selections; `make explain-run RESULTS_DIR=<explicit-compatible-root> DETAIL=performance JSON=1` only when a supported root is identified. | Existing typed diagnostics; optional authorized instrumented executions; no new timing gate. | Report unmeasured/incomparable costs; do not invent savings. |
| V-011 / generated routing and final portfolio | F-002/F-013/F-017; C-001; S-004/S-014 | During authorized implementation, when authored changes affect generated outputs: `make generate` and affected `make source-check CHECK=generated`; `make source-check CHECK=schema` only for affected schema/structured-input contracts. Use `make agent-finalize` only for required refresh/validation actions or requested retained-run maintenance; reuse compatible completed checks. | Authored inputs first; exact old/new selector map; compatible final owner union and required non-row obligations. | No orphan rows, evidence loss, or completeness claim; investigate before rerun. |
| V-012 / stability and gaps | F-017; C-001/C-008/C-010; S-012 | Read exact identified terminal summaries through supported analysis; later narrow reproductions only when authorized and tied to a hypothesis. | Matching source/profile/environment; original failure preserved; gap-to-observation mapping. | Potential concerns stay potential; no broad flaky/covered claims. |
| V-013 / recovery traceability | F-015; C-011; S-013 | Read Core REQ-01-593 and Harness REQ-654/655; inspect app.operator/module.recovery command mappings and selected assertions before choosing exact future rows. | Seven-command obligation matrix and discovered routing; no current product execution claim. | Correct subordinate prose; actual routing remains unresolved until inspected. |
| V-014 / documentation and preservation | All document mappings and ACs | Direct document/Git review and canonical `make lint-markdown`; S-015 fixes filename coverage and S-014 records the final pass. | Authorized implementation diff; linked files, stable IDs, acyclic dependencies, normative/observational separation. | Confirm the tracker is selected and review formatter changes; no remaining lint limitation. |

The original design included discovery stages in V-001/V-003/V-005/V-006/
V-009/V-012/V-013. Those stages are complete for this effort; current selections,
retirement mappings and evidence are recorded in the workstream checkpoints.
The baseline table is not a fresh execution request or a substitute for current
Make discovery.

### 8.2 Initial authoring admission and results (historical)

| Validation | Invocation-path evidence / effects | Admission / execution result | Applicability |
| --- | --- | --- | --- |
| V-001 | Git/file reads and in-memory JSON aggregation; no runner invoked. | Executed as source inspection; E-001–E-003/E-020. | Establishes counts/metadata, not complete semantic review or runtime stability. |
| V-002–V-013 | Product/harness executions can create artifacts, compile, acquire services, or regenerate outputs. | Not executed; authoring task excludes implementation and test execution. | Future obligations only. Prior audit source observations remain baseline-qualified, not passed tests. |
| V-014 canonical Markdown lint | At the audit baseline, the wrapper wrote cache/run artifacts. During authoring the user simplified the target: current harness.lint_markdown directly invokes Biome lint with `--write --error-on-warnings .`, without the old wrapper. | SKIPPED: the current command can rewrite other selected Markdown files, outside the tracker-only boundary. No bypass invocation or run-then-restore was used. | biome.markdown.json still selects `docs/handoffs/*-module-refactor-tracker.md`, which does not match the requested filename. A generic success would not establish this file's inclusion. E-017/E-025. |
| V-014 direct review | Read-only checks of document references, identifiers, tables/dependencies, structure, and repository fingerprints. | Direct review completed; all checked internal IDs, links, table shapes, selectors, and workflow dependencies resolve. Preservation comparison passed; E-023/E-024. | Editorial verification, not canonical lint or product conformance. |

The Refactor Tracker skill used in authoring (host-local source:
`/mnt/c/Users/jochi/.codex/skills/refactor-tracker/SKILL.md`) states: “Do not run commands that may write outside the
tracker.” Its local installation is not a repository requirement or portable
dependency. That restriction applied to the initial planning-only session. The
subsequent explicit implementation request superseded it, as recorded in current
TCR-REQ-002; it is not a restriction on the completed execution work.

## 9. Top-Level Work Tracker

**TCR-REQ-023 — Checkpoint updates.** Each execution session MUST update status only
from recorded evidence, preserve stable IDs, and identify superseded conclusions.
The workstream dependency graph remains §6. The initial authoring work table
below is historical; current implementation status is recorded per slice in §6
and each execution checkpoint. A completed workstream MUST be recorded before
the next workstream begins.

| Work item | Planning or future implementation | Workflow / slice | Status | Evidence / blocker | Exit condition |
| --- | --- | --- | --- | --- | --- |
| T-001 Admit target and preserve baseline | Planning | WF-00 | DONE | E-001; final comparison E-024 pending. | Correct absent destination, scope, authority, and baseline recorded. |
| T-002 Establish catalog and named-file census | Planning | WF-01 / S-001 | DONE | E-002/E-003. | 66-owner catalog and naming census recorded with limitations. |
| T-003 Close all registrations and family-purpose review | Future investigation | WF-01 / S-001 | TODO | RB-002. | No unexplained discovered test/helper/registration and complete required dispositions. |
| T-004 Resolve missing architecture authority | Future owner resolution | WF-02 / S-002 | BLOCKED | RB-001. | Exact relevant file/package policy has adopted authority. |
| T-005 Qualify and correct ineffective scanners | Future correction | WF-03 / S-003 | BLOCKED | RB-003; F-001. | Required roots and negative/allowed fixtures proven. |
| T-006 Consolidate backend architectural checks | Future structural work | WF-04 / S-004–S-006 | BLOCKED | RB-001/RB-003/RB-004/RB-005. | Retirement mappings and unique assurance preserved. |
| T-007 Consolidate frontend policy checks | Future structural work | WF-04 / S-007 | BLOCKED | RB-005. | Dependency/symbol/type/behavior equivalence proven. |
| T-008 Relocate incident API evidence | Future structural work | WF-05 / S-008 | BLOCKED | RB-006. | Route/profile/error and rendered witnesses retained. |
| T-009 Relocate coordination API evidence | Future structural work | WF-05 / S-009 | BLOCKED | RB-007. | Seven-surface and public demotion matrix passes. |
| T-010 Remove proven selector assertion duplication | Future structural work | WF-05 / S-010 | BLOCKED | RB-003. | Narrow compatible evidence and mounted-node map complete. |
| T-011 Evaluate support placement and marginal cost | Future investigation | WF-06 / S-011 | TODO | RB-002/RB-008. | Purpose/placement decisions and honest cost conclusions. |
| T-012 Assess stability and requirement gaps | Future investigation | WF-07 / S-012 | TODO | RB-002/RB-008. | Confirmed versus potential concerns and missing observations resolved/dispositioned. |
| T-013 Correct recovery restatement and map evidence | Future correction/investigation | WF-02 / S-013 | TODO | E-021; actual routing RB-002. | Seven-command authority and actual test map agree or named dependent correction remains blocked. |
| T-014 Close routing and final evidence | Future completion | WF-08 / S-014 | BLOCKED | All applicable unresolved retirement gates. | Required owner union, selectors, projections, and non-row obligations close. |
| T-015 Deliver this handoff | Planning | WF-00 / V-014 | DONE | E-023/E-024; canonical lint explicitly skipped. | Tracker written, reviewed, limitations recorded, outside content/index preserved. |

## 10. Session Handoff Log

### Coverage and reliability closure: S-012

All identified remediation obligations have a survivor, accountable owner and
validation disposition below. Registration completeness, requirement coverage and
runtime stability are separate conclusions. The current portfolio has complete
registration/disposition accounting; unchanged families have the engineering
purposes in §2.5. This effort does not certify an exhaustive requirement-to-every-
assertion mapping or the absence of undiscovered product gaps. That limitation is
not used to remove a test, assert conformance, or leave an identified replacement
unqualified.

| Reviewed obligation / owner | Required observation and prior gap | Current evidence / disposition |
| --- | --- | --- |
| C-002; Core 01 §2.1; harness.generated_artifacts/module.revisions | Runtime roots must actually be scanned; nested imports, errors and current generated families must be enforceable; focused Revisions selection survives. | S-003/S-016/S-004 exact rows PASS with forbidden, malformed, empty and allowed-refactor fixtures. Thin Revisions selector retained by explicit decision. |
| C-003; Core 00/01 and Graph §8.1 | Seven missing architecture references and duplicate AC identities made source authority ambiguous. | S-002 adopted enduring owner sections; AC-598..602 replace only duplicate definitions. S-012 removes obsolete transition-order language from adoption summaries. Editorial/source review, not runtime conformance. |
| C-003/C-004; module.projections and source owners | Exact-file exceptions, provider SQL discovery/access and direct record-envelope locking could escape or violate owner capabilities. | S-005/S-006/S-017 source, detector, SQLC/export/parity, query, Records and Links witnesses PASS. F-022 formalizes the discovered production defect and its separate corrective slice. |
| C-005; web.architecture/module.imports/web.networkflow | Alias/module identity and literal spelling could hide dependencies or masquerade as actual consumption. | S-007 resolved module fixtures, selectable/forbidden workbook mapping, claim/availability and generated Network Flow identity PASS. Unique type/focus/transport/ownership checks remain. |
| C-006; module.incidents/app.server; REQ-01-161/175 | Browser requests alone had no incident UI witness; existing units lacked assembled exact errors. | S-008 process errors retain status/code/field/reason. Qualified before row retirement. |
| C-007/C-009; app.server; REQ-01-542..548 | Partial discovery shape/claim checks and duplicate enterprise variables did not establish all claimed dispatch distinctions. | S-008 complete default-profile discovery, zero membership, pagination, root/descendant/template/base/outside dispatch PASS; retained rendered empty landing PASS. Other dispatcher/claimed-profile tests remain unchanged. |
| C-008; module.workbook; §5.3 matrix | Only two prior HTTP patch surfaces and SQL demotion did not replace seven public API round trips. | S-009 all seven exact fields plus two pre-existing patch cases PASS, including public demotion, original session, allowed reads, denial details, unchanged source/history/row/version. |
| C-009; package.ui/module.auth | Builder equality and repeated attachment add no new observation after visibility. | S-010 selector unit and surviving mounted-state browser case PASS; cross-screen observations retained. |
| C-010; Harness REQ-053/054/800 | Metadata alone concealed purpose and supporting placement; current-source gates must not disappear. | S-011 19 corrected placements, five owner migrations and architecture/static metadata; routing, helper rows, catalog/schema/generated checks PASS. Production/smoke/current-artifact and explicit-only boundaries retained. |
| C-011; Core 01 §12.2.1 and Harness REQ-655 | Five-command subordinate restatement and unclear export/bundle mapping. | S-013 grammar ownership corrected. Seven parser cases and separately asserted process invocations reviewed; unchanged process behavior not rerun or claimed passed. |
| C-001/C-010; all 66 owners | Hidden registration and specification-free supporting material lacked a clear disposition. | S-001 source/registrar reconciliation plus S-011 current delta/accountable-purpose index; catalog PASS. Helpers, fuzz seed, three TestMain and vendor contexts accounted separately from cases. |

Recovery routing is explicit rather than a six-test/seven-command counting error:
`backup inspect latest`, `backup create`, `backup export latest`, `restore latest`,
`restore-verify latest` and `restore-verify due` are exercised by
`module.recovery.process.canonical_operator_process_evidence_maps_the_imp_9808fdd4e9`.
Export has its own result/operation, artifact/journal and occupied-output assertions
inside `TestCanonicalOperatorBackupCreate_Process`. `restore bundle` maps to
`module.recovery.process.portable_source_loss_restore` with result/replay/failure
assertions. All seven have cases in `TestRecoveryCLIParserAndInvalidInvocationContract`.
These are inspected routing/assertion facts, not new runtime acceptance results.

Task failure ledger (canonical summaries preserve their original classifications):

| Retained run under `.cartulary/test-results/` | Observed failure and relationship to this change | Repair / compatible qualification |
| --- | --- | --- |
| `tcr-s005-schema-20261109-01` | product/test_assertion_failure: synthetic scanner fixture used a prohibited prose path. No actual prose was read. | Renamed fixture scope to outside; schema-02 PASS. |
| `tcr-s005-policy-20261109-01` | product/test_assertion_failure: old Recovery filename-specific admission expectation contradicted adopted package authority. | Changed the allowed split-file fixture; policy-02 PASS. |
| `tcr-s006-generate-20261109-01` / `02` | Generation rejected unsorted catalog input, then a shell row whose verification lacked shell evidence. | Sorted authored row and added the correct evidence kind; generate-03/04 and subsequent generation PASS. |
| `tcr-s006-sql-20261109-01` / `02` | product/test_assertion_failure during qualification: constant context, maintenance/recursive coverage and actual F-022 owner violation were exposed. | Parser/policy refinements plus S-017 production correction; sql-05/06, mechanics-03, boundary-02 and query-01 PASS. |
| `tcr-s006-sql-20261109-04` | product/test_assertion_failure: strict stale-assignment admission detected 19 obsolete SQL declaration bindings. | Removed stale assignments; sql-05/06 PASS. |

These were deterministic authoring/qualification failures followed by changed
inputs. They are not evidence of an intermittent/flaky test or an equivalent-input
retry success. There are no unresolved observed task failures. No watchdog,
timeout, service reset, golden or assertion relaxation was used to hide a failure.
The S-008 error-detail adjustment follows Core's additive-member compatibility
rule and retains every required value; it did not repair a failing test.
Performance and whole-suite stability remain unmeasured. A future reliability
claim requires a concrete failure hypothesis and compatible repeated observations;
this effort introduces no blanket rerun or performance qualification gate.

### Historical execution checkpoint: S-014 complete (superseded by corrective review)

Final implementation/validation/handoff complete. Reconciled §2.5 inventory,
all 22 findings, all 20 acceptance criteria, eight original blockers, nine retired
row migrations and all 17 workstreams. Original authoring-only observations and
failures remain marked as historical; they do not override current authorization
or become runtime passes. No identified blocker or unsafe retirement remains.

Validation: `make lint-markdown` PASS, 124 files, no fixes; `git diff --check`
PASS. Human editorial review found no broken local links in the tracker or six
changed owner/navigation documents. Core 04 has 597 unique AC definitions;
AC-598..602 each occur once as definitions. All 24 TCR requirement identities
remain unique and all 12 required top-level sections remain present. This was
human document maintenance, not executable specification-dependent evidence.

Reused S-011 catalog/schema/generated/routing passes for unchanged authored
routing and the recorded per-slice targeted runtime passes for unchanged source.
No further full-owner, full-suite, Visual/UI-review, release or finalizer run is
required by the affected contracts. No claim of global conformance, universal
stability or measured speedup. Original failure roots and repair evidence remain
in the S-012 ledger.

HEAD remains beaf0293634e6aef9e62aa5670e033a26b6458d3 on main; no staged changes,
commit, merge or deployment. Ordinary Git inspection may refresh index stat
metadata, so the initial index byte digest is not reused as a final identity
claim. All substantive edits and generated projections belong to the accepted
slices. The workspace is ready for maintainer review; there is no next required
implementation slice.

### Execution checkpoint: S-012 complete

Closed the reviewed obligation/observation matrix and original failure ledger
above. All identified gaps have repaired or explicitly retained survivors;
unchanged source/registration/vendor groups retain engineering-purpose decisions.
Formalized F-022 and S-017 traceability, verified seven-command Recovery routing
including separately asserted export and portable restore, and removed historical
transition-order wording from Core 00 adoption summaries. No test or runtime
input reads these documents.

Reviewed all task-owned canonical graph summaries: the five failed graph runs
and two failed generation runs are accounted for with concrete repairs and later
compatible passing selections. No unresolved observed task failure, intermittent
failure conclusion, broad stability claim, timing improvement or relaxed timeout.
S-011 catalog/routing and prior slice passes remain compatible; this slice adds
only prose/traceability changes, so product tests are not repeated. S-014 owns
final Markdown validation, current reference/acceptance reconciliation and handoff.

### Execution checkpoint: S-011 complete

Completed the current inventory delta and accountable retention-purpose index
in §2.5: 2,087 source/registration files and 2,609 active rows across 66 owners;
no unaccounted test-shaped additions. This closes disposition, not an invented
assertion-by-assertion proof of all product requirements. Supporting producer,
fixture, registrar, helper and vendored contexts have explicit purposes.

Corrected Reporting's architecture verification and the shared Go source-policy
evidence class. Moved five fixture/DOM-helper rows from web.architecture to
harness.browser with exact migrations and unchanged case/profile selection.
Applied TH-HARNESS-REQ-054 full-tier placement to 19 reviewed self-regression
rows. Kept all required source/schema/catalog/generated/migration/security,
embedded-asset, runtime-admission/reset and smoke gates; explicit-only remains
66 rows. Strengthened existing scheduler tier sentinels for all corrected rows,
removing the obsolete assertions that two observation-helper suites were standard.

PASS/outer exit 0: `make test-slice OWNER=harness.command_surface
ROWS=harness.command_surface.behavior.scheduler_contract` in
`tcr-s011-routing-20261109-01`; the five exact migrated harness.browser rows in
`tcr-s011-helpers-20261109-01`; `make generate` and
`make source-check CHECK=catalog`, `CHECK=schema`, `CHECK=generated` in respective
`tcr-s011-{generate,catalog,schema,generated}-20261109-01` roots. The routing test
compiles and checks fast/check/ci/release and direct selections; it does not execute
those whole aggregates. No performance claim or new benchmark gate. Rollback
must restore old/new owner rows, verification references, migrations, independent
tier assertions and generated projections together. Next: S-012.

### Execution checkpoint: S-010 complete

Removed the bootstrap selector string equality already owned by the selector
unit table and both redundant attachment assertions after feedback visibility.
Renamed the scenario to describe its actual mounted observations; its stable row
and scenario identity remain selected. Shell/landing status and error nodes,
account security feedback, administration status and logged-out login/feedback
visibility are preserved. No selectors, screen behavior or goldens changed.

PASS/outer exit 0: `make test-slice OWNER=package.ui
ROWS=package.ui.frontend_unit.application_selector_contracts_675995f2ab` and
`make test-slice OWNER=module.auth
ROWS=module.auth.browser_support.supplemental_browser_verifies_ordinary_shell_sel_f8a9da1957`,
roots `tcr-s010-unit-20261109-01` and `tcr-s010-browser-20261109-01`.
`make generate`, `make format`, and `make source-check CHECK=generated` /
`CHECK=schema` PASS under respective `tcr-s010-*-20261109-01` roots in
`.cartulary/test-results/`. This removes assertions inside a surviving scenario;
it does not remove browser setup cost. Next: S-011.

### Execution checkpoint: S-009 complete

Extended `TestWorkbookPatchSurfaceContracts_Integration` with the seven §5.3
coordination fields. One assembled server/authenticated actor serves isolated
per-case incidents; the existing Timeline collection and Party primary-email
cases remain. Every case now queries the created row, checks a safe integer
version, patches and reads back value/version, verifies exact/divergent replay,
then demotes membership through the public API. The original session reads as
viewer, receives 403/authorization_denied with required_role for a new changed
write, and leaves row/version/source snapshots, receipts/history and publication
intents unchanged. Legitimate membership effects are captured before denial.

The former API-only coordination Playwright file and row are retired after the
assembled replacement passed; its migration points to the existing workbook
surface-contract row. No UI, wire, schema or data migration. Restoring a retired
row requires its source and batching projection together, not a silent alias.

PASS/outer exit 0: `make test-slice OWNER=module.workbook
ROWS=module.workbook.support_integration.workbook_surface_create_query_and_patch_contract_ae5c34d894`,
run `tcr-s009-http-20261109-01` (14 create/query plus 9 patch cases).
`make format` (01/02), `make generate`, and
`make source-check CHECK=generated` / `CHECK=schema` PASS under
`tcr-s009-{format,generate,generated,schema}-20261109-01` (final format 02), all
beneath `.cartulary/test-results/`. Original browser execution was unnecessary:
its source obligations were transferred individually and the replacement was run.
No full-tier or fast-gate timing claim. Next: S-010.

### Execution checkpoint: S-008 complete

Qualified the assembled process witness for unknown incident-create fields and
forbidden title patch, preserving status/code/required detail values. Discovery
now compares the complete seven-member resource array, registry order and all six
claim states under the same four-claimed default profile as the browser fixture;
it proves zero memberships and singleton-pagination rejection. The assembled
reserved-family witness adds readiness/base dispatch and an actual templated
enterprise descendant alongside root, descendant, Network Flow and outside 404.

Retired three browser catalog rows through explicit replacement mappings; retained
`app.server.browser_support.zero_membership_landing` as a real login plus visible
empty-directory witness. The catalog maps both prior empty landing observations
to that single case. No product behavior or public compatibility change.

PASS/outer exit 0: `make test-slice OWNER=app.server` selecting the reserved-family
integration and incident process-smoke rows, run `tcr-s008-http-20261109-01`;
then the exact new landing row, `tcr-s008-browser-20261109-01`.
`make generate`, `make format` (01 and 02), and
`make source-check CHECK=generated` / `CHECK=schema` pass under respective
`tcr-s008-{generate,format,generated,schema}-20261109-01` roots (format final 02).
All roots are under `.cartulary/test-results/`.
After the HTTP pass, removed an unused partial-claim helper and changed detail
comparison from whole-map equality to required-member equality: Core permits
additive error details. The prior stricter passing observations cover those same
required values; no new runtime path was introduced. Discovery producer shape
remains exact as required by REQ-01-542. Restore browser source, row migrations,
authored catalogs and generated batching together on rollback.
No timing improvement is claimed. Next: S-009.

### Execution checkpoint: S-007 complete

Implemented compiler-backed module resolution for nearest/inherited TypeScript
configuration, path aliases and package identities. Central restrictions and cycle
checks consume the resolved identity; aliased raw transport, private workspace,
vendor, re-export, dynamic, require and type-only bypasses have negative fixtures.
The resolver does not collect a TypeScript program or acquire prose inputs.
Replaced literal import-target source checks with actual workbook mapping across
all selectable targets, rejection of reserved/unknown targets and Network Flow
availability gating plus generated target/profile consumption.

Semantic parity disposition: central dependency direction owns resolved imports;
ambient fetch/symbol detection remains in transportBoundaryPolicy, exhaustive
source ownership remains in sourceOwnershipPolicy, Inspector focus/geometry and
Timeline composition/type checks remain local, and Network Flow contract/design
checks remain distinct. No unsupported blanket consolidation or deletion.

Validation: `make test-slice` with exact rows recorded in the manifests for
`tcr-s007-architecture-20261109-01` (two boundary rows),
`tcr-s007-imports-20261109-01` (generated-target and operator assistant rows), and
`tcr-s007-network-20261109-01` (analytical import ownership) all PASS, outer exit 0.
`make generate`, `make format`, `make lint CHECK=types`, and
`make source-check CHECK=generated` / `CHECK=schema` also PASS under respective
`tcr-s007-{generate,format,types,generated,schema}-20261109-01` run roots in
`.cartulary/test-results/`. No public API, UI, selector or data migration. Restore
checker, consumer cases and authored/generated catalogs together if reverting.
Next: S-008.

### Execution checkpoint: S-006 complete

Replaced TestProjectionProviderSQLSourceOwnership and
TestValidateProjectionProviderSQLReference with the module-owned
module.projections.architecture.sql_declaration_ownership shell row and shared
parser qualification in the owner-boundaries fixture operation. Existing
assembly row retains descriptor-mirror, exact descriptor/boundary/recovery/
migration-table equality and complete-port tests. No whole row was retired.

Go AST inspection supplies stable top-level declaration identities; recursive
reviewed source scopes reject empty/missing/linked input, new unassigned SQL,
unknown provider roles and stale explicit bindings. Dedicated source packages
have provider defaults; mixed packages have reviewed declaration assignments.
Nonprojection Evidence/Links operations are explicitly classified and retain
their independent backend policies. Restore incident enumeration and descriptor-
bounded row counting have separate maintenance permissions. Dynamic provider
relations and authoritative source writes/locks fail. Shared SQL parsing now
resolves package constants and nested expressions; private relation-alias helpers
and fragmented tag predicates were simplified without changing emitted SQL.
Fixtures exercise quoted/schema names, CTEs, reads/writes/locks, unknown authority,
renames/splits, new declarations, descendants, malformed nested SQL and stale
assignments. Prose is absent from executable inputs.

Qualification failures are retained at tcr-s006-sql-20261109-01/02/04: missing
constant context, maintenance distinction, S-017's actual owner defect, additional
recursive source declarations, fragmented expressions, and stale assignments.
The corrected implementation passed at tcr-s006-sql-20261109-05 (including Go
parity) and tcr-s006-sql-20261109-06 (final parser). Shared mechanics passed at
tcr-s006-mechanics-20261109-03; full backend policy at
tcr-s006-boundary-20261109-02; focused compiled query and shape/keyset rows at
tcr-s006-query-20261109-01. All outer exits 0. Generation and format passed at
tcr-s006-generate-20261109-04 and tcr-s006-format-20261109-02; generated and schema
checks passed at tcr-s006-generated-20261109-01 and
tcr-s006-schema-20261109-01. Later parser qualification changed neither authored
schemas nor generated inputs. Exact roots are under .cartulary/test-results.
No whole product suite or release pass is claimed. Next: S-007.


### Execution checkpoint: S-017 complete

Records now owns LockIncidentEnvelopeIdentitiesTx; application projection
assembly adapts its small owner DTO into a consumer-owned endpoint-lock function.
Projections construction rejects a missing function and linked refresh delegates
through the same caller transaction. The original deterministic NO KEY UPDATE
query is moved intact; missing-endpoint and lock-error handling remain explicit.
Removed Projections' direct envelope SQL permission. No public API or schema
migration, ambient fallback or peer-store construction added.

`make format`, Records envelope integration, Links projection/history workflow,
Projections adapter/private catalog rows, and backend-boundary lint all passed
with outer exit 0. Exact roots: tcr-s017-format-20261109-01,
tcr-s017-records-20261109-01, tcr-s017-links-20261109-01,
tcr-s017-adapters-20261109-01, tcr-s017-boundary-20261109-01 under
.cartulary/test-results. Fixtures cover ordering, incident isolation, borrowed
transaction usability, required construction, error propagation and missing
endpoints. The Links scenario proves production composition still updates
projections/history. Resume S-006 analyzer qualification and retirement.


### Execution checkpoint: S-006 qualification; S-017 corrective entry

The replacement SQL check is registered as a module.projections static shell
row, using the existing parser plus Go AST declaration identities. Generation
attempts tcr-s006-generate-20261109-01/02 failed on unsorted rows and missing
shell evidence-kind support; repaired authored metadata and generation passed
at tcr-s006-generate-20261109-03. Format passed; declaration/SQL mechanics passed
at tcr-s006-mechanics-20261109-01. Production check tcr-s006-sql-20261109-01
failed and cleanup passed: declaration isolation omitted shared constants,
restore incident enumeration needs a separate maintenance authority, and linked
refresh directly locks Records-owned envelopes. The first two are analyzer
qualification; the last is new F-022, a production conflict with REQ-01-649.
S-006 closure waits for corrective S-017, which is explicitly added rather than
widening provider authority to tolerate the defect. Remediation: a narrow
consumer-owned endpoint-lock port implemented by Records and injected in
application assembly. No public/schema migration; preserve NO KEY UPDATE,
deterministic ordering, same transaction and current missing-endpoint failure.
Validate constructor failure and actual linked-refresh behavior. Leaving the
bypass would undermine lock ownership and later Records evolution.


### Execution checkpoint: S-005 complete

Seven central rules now grant exact packages: four private projection-provider
consumers, Entities projection contract/port consumers, and Recovery SQLC.
Descendants/peers remain denied; fixtures qualify rename/split and bypass cases.
Projections import inspection now shares sourcecheck parsing. Local Entities,
Indicators, Reporting and Revisions positive allowlists remain because central
negative rules do not prove closed dependency membership. SQLC retains its
separate exported-declaration AST check; Projections retains production-assembly
and test-support delegation checks. Runtime, generated, support, test and
build-tag scope differences were reviewed, not equated by similar names.

Validation: format; backend-boundary lint; exact Projections architecture and
SQLC rows passed at tcr-s005-format-20261109-01,
tcr-s005-boundary-20261109-01, tcr-s005-projections-20261109-01 and
tcr-s005-sqlc-20261109-01. Schema and aggregate owner-boundary fixture checks
initially failed (outer 2; product/test_assertion_failure): the synthetic
out-of-scope fixture used a restricted prose-root name, and one old Recovery
fixture expected a filename grant. Changed the synthetic path to outside/ and
updated the expected package permission; retained failures at
 tcr-s005-schema-20261109-01 and tcr-s005-policy-20261109-01.
Rechecks passed with outer exit 0 at tcr-s005-schema-20261109-02 and
tcr-s005-policy-20261109-02 (107 fixtures). All roots are under
.cartulary/test-results. Policy input has no generated projection; schema
validation was required, generation was not. No services selected. Next: S-006.


### Execution checkpoint: S-004 complete

Retained the Revisions-owned thin witness: catalog cases have one owner and
collaborator metadata does not select another owner's case. A new fixture proves
that exact witness rejects a forbidden descendant. Shared runtime scanning now
also rejects an empty production scope; generic source inspection still permits
empty input. No cross-owner selection mechanism or compatibility alias added.
`make format` and exact Revisions and broad generated-artifact rows passed with
exit 0 at tcr-s004-format-20261109-01, tcr-s004-revisions-20261109-01 and
tcr-s004-broad-20261109-01. Next: S-005.


### Execution checkpoint: S-016 complete

The retired monolithic generated-contract path/file allowlist is replaced by
exact consumer packages for all 26 active families in contracts/index.json.
Explicit empty assignments deny production access; missing and stale family
assignments, unknown generated packages, descendants and peers fail closed.
Fixtures prove permitted rename/split, unauthorized cmd/peer/child access,
unknown/retired imports and explicit no-consumer disposition. The authored
registry supplies output membership, not authority to import. Tests consume no
prose. Existing catalog selectors still reach the moved test function.
`make format` and the exact generated_artifacts guard row passed (exit 0),
roots tcr-s016-format-20261109-01 and tcr-s016-guard-20261109-01 under
.cartulary/test-results. No generated inputs changed. Next: S-004.


### Execution checkpoint: S-003 complete

Both import guards now use explicit cmd/internal roots and shared Go AST import
inspection. Pruned generated/support directories are never opened; tests are
excluded, build-tagged imports retained, syntax/read/walk failures fail closed,
linked Go files rejected, and results ordered. Detector fixtures prove nested
cmd/internal offenders, aliases/grouped imports, comments/prefix lookalikes,
all support exemptions, malformed input, linked source and missing roots.
`make format` passed (tcr-s003-format-20261109-01); reviewed changes are limited
to intended files. Exact test-slice rows for generated_artifacts runtime import
guards and command_surface Go import inspection passed, with outer exit 0 at
`.cartulary/test-results/tcr-s003-guard-20261109-01` and
`.cartulary/test-results/tcr-s003-parser-20261109-01`. No service rows selected.
Next: S-016, because the now-effective generated guard still targets a retired
monolithic contract package.


### Execution checkpoint: S-013 complete

TH-HARNESS-REQ-655 delegates grammar to Core 01 §12.2.1 and requires mapping
per invocation, permitting separately asserted multi-command workflows. All
seven commands have explicit parser cases in recoverycli/cli_test.go. Process
witnesses are TestCanonicalOperatorBackupInspectLatest_Process,
TestCanonicalOperatorBackupCreate_Process (including separately asserted
backup_export_latest), TestCanonicalOperatorRestoreLatest_Process,
TestCanonicalOperatorRestoreVerifyLatest_Process,
TestCanonicalOperatorRestoreVerifyDue_Process and
TestCanonicalOperatorRestoreBundle_Process (portable_restore_process_test.go).
Source review confirmed export operation/result, journal and occupied-output
failure assertions and portable restore result, replay and failure assertions.
This is a specification correction; executable behavior and routing remain
unchanged. `make lint-markdown` passed (124 files); process tests were not rerun
and no new runtime pass is claimed. Next: S-003.


### Execution checkpoint: S-002 complete

Core 01 REQ-01-690 now owns durable package-capability rules for six modules;
Graph Projection §8.1 already owns its engine/restore topology. Core 00's seven
adoption entries and owner matrix, plus Domain navigation, resolve to those
present owners. No missing ADR was reconstructed from implementation history.
The second ambiguous Core 04 definitions were assigned AC-598 (clipboard),
AC-599 (migration), AC-600 (shortcuts), AC-601 (portable restore), AC-602
(package lifecycle); current references were reviewed by meaning, including
retention of AC-537 for Network Flow cleanup. Human editorial inspection found
no duplicate AC definitions or remaining current missing-decision references.
`make lint-markdown` passed (124 files). No executable prose dependency or
product runtime claim was added. Next: S-013.


### Execution checkpoint: S-001 complete

S-001 completed before S-002 entry. §2.4 records current registration closure,
per-file ownership/retention accounting, nonconventional registration mechanisms,
TestMain and fuzz dispositions, and explicit limits on assertion-level review.
`CARTULARY_TEST_RUN_ID=tcr-s001-catalog-20261109-01 make source-check CHECK=catalog`
exited 0; source-check reports pass in
`.cartulary/test-results/tcr-s001-catalog-20261109-01`. The planned operation union
was frontend readiness plus catalog validation, with no services or product rows.
No test, helper, fixture, runner, catalog selection or product code changed.
Full product requirement coverage and runtime stability are not inferred from
registration completeness; those obligations remain with S-011/S-012. Next:
S-002 specification authority and acceptance identity repair.

### Execution checkpoint: S-015 complete

S-015 completed before S-001 entry. `biome.markdown.json` now includes
`docs/handoffs/*-refactor-tracker.md`, covering both the prior module trackers and
this controlling artifact. `make lint-markdown` exited 0: 124 files checked,
no fixes applied. The diff contains only the intended include-pattern change and
tracker maintenance. No generator, product test or finalizer is required for this
configuration-only inclusion change. F-019 is resolved for this implementation
scope; prior skipped-lint records remain historical. Rollback is the single
include-pattern edit, with loss of this tracker's coverage explicitly understood.
Next: S-001 full registration and purpose accounting.

### Execution checkpoint: baseline and S-015 entry

2026-11-09: implementation authorized by the maintainer after acceptance of the
complete plan. HEAD is `beaf0293634e6aef9e62aa5670e033a26b6458d3`, branch `main`.
Tracked/index state was clean; this tracker was the only untracked file. Its
starting SHA-256 was `97d705039c234bacaec9c2104ca3cf69292d3645fef9385d82cfc631a7918ce7`;
index SHA-256 was `07bae1a6f4f6aee532d28f5b69823f5a3e1c6d221817ec65b01abe2ee2249668`.
Root/backend/tooling instructions and test-operations procedure were revalidated.
The initial authoring-only restrictions and grouped workflow graph are superseded
by the accepted execution plan. No runtime result is imported from planning.
S-015 is in progress: broaden the tracker filename pattern, then run the existing
canonical Markdown lint and review its diff before starting S-001.

The execution plan adds F-020 (duplicate acceptance identities) and F-021
(retired generated-contract package), expands F-016 to seven missing decisions,
and adds S-015/S-016 without renumbering earlier identities.


### 10.1 Initial authoring scope and authority (historical)

| Date / baseline | State / evidence | Files inspected or touched | Commands / results | Next boundary |
| --- | --- | --- | --- | --- |
| 2026-11-09 / ba75e839 | User selected full-suite scope; new tracker is scoped change authority, subordinate to adopted behavior owners. E-001. | Root/nested instructions, NLSpec guidance, framework, relevant Core/harness/domain clauses; only requested tracker to be written. | Read-only Git/file inspection; clean initial state. | Resolve RB-001; do not treat framework or machine projections as product authority. |

### 10.2 Backend boundary

| Date / baseline | State / evidence | Files inspected | Commands / results | Next boundary |
| --- | --- | --- | --- | --- |
| 2026-11-09 / ba75e839 | F-001–F-005/F-018 remain static findings. E-004–E-008. | codegenboundary, local guards, projection assembly, central policy/parser and negative fixtures. | Inspection only; no Go tests or injected violations executed. | Qualify S-003 before consolidation; resolve exact-file/package intent. |

### 10.3 Frontend boundary

| Date / baseline | State / evidence | Files inspected | Commands / results | Next boundary |
| --- | --- | --- | --- | --- |
| 2026-11-09 / ba75e839 | F-006/F-007/F-018 distinguish import, symbol, type and runtime obligations. E-009. | Inspector/Timeline/transport/source-ownership/Network Flow/import-target tests and policies. | Inspection only; no Vitest/browser execution. | Complete semantic parity matrix; retain distinct guarantees. |

### 10.4 Contract and codegen

| Date / baseline | State / evidence | Files inspected | Commands / results | Next boundary |
| --- | --- | --- | --- | --- |
| 2026-11-09 / ba75e839 | Typed policies/catalogs are downstream projections. E-003/E-018/E-021. | Verification owners, test families, source discovery, work graph/task surfaces; Core/harness recovery clauses. | JSON aggregation and source reads only; no generator. | Later authorized row changes update authored inputs first; recovery restatement requires separate correction. |

### 10.5 Tests and harness

| Date / baseline | State / evidence | Files inspected | Commands / results | Next boundary |
| --- | --- | --- | --- | --- |
| 2026-11-09 / ba75e839 | Catalog baseline and candidate assertions established; suite-wide semantic/reliability review partial. E-002/E-003/E-010–E-020. | Full family metadata; candidate/browser/backend counterparts; fixture/runner/checker sources. | No product/harness run or latest-run discovery; lint skipped with concrete effects and inclusion limitation. | Finish RB-002, select exact compatible retained evidence if later authorized, then narrow execution. |

### 10.6 Security and authorization

| Date / baseline | State / evidence | Files inspected | Commands / results | Next boundary |
| --- | --- | --- | --- | --- |
| 2026-11-09 / ba75e839 | Current-role demotion, CSRF/header behavior, reserved dispatch, and local recovery authorization remain distinct. E-010–E-013/E-021. | Browser request helpers, route/service/process tests, Core owner clauses. | Static inspection only; no security pass claimed. | Preserve rejection/no-side-effect witnesses; do not infer browser policy from APIRequestContext. |

### 10.7 Initial authoring risks and next session (historical)

| Date / baseline | Current state | Files touched | Results / blockers | Next action |
| --- | --- | --- | --- | --- |
| 2026-11-09 / ba75e839 audit; beaf0293 immediate pre-write state | Initial handoff; full-effort planning partial; implementation not authorized. | Requested tracker only; user lint/instruction changes remain independent. | RB-001–RB-008; E-025 requalifies lint/current guidance; E-023/E-024 record completed direct review and preservation comparison. | Revalidate baseline; begin S-001 and S-002 investigation independently; S-003 needs authorized characterization before correction. |

No historical runtime duration or pass from the earlier a43248fb survey is
used to close an acceptance criterion. The earlier survey and same-baseline
deep dive supplied discovery leads; current source/catalog evidence anchors
the findings here. Failed read-only inspection attempts, including a wrong
config path and output truncation, were corrected or excluded; they produced
no repository artifacts and are not verification passes.

### 10.8 Initial source evidence register and execution baseline

Paths below are repository-relative from this document through links.
“Inspected” means the stated metadata or bounded assertions, not every behavior
of the containing subsystem.

| Evidence | Source / exact inspection surface | Baseline / observation | Applicability |
| --- | --- | --- | --- |
| E-001 | [AGENTS](../../AGENTS.md), [backend instructions](../../internal/AGENTS.md), [tooling instructions](../../tools/AGENTS.md), [Core 00](../spec/00_document_set_status_and_precedence.md), [domain](../domain.md), [NLSpec guidance](../research/nlspec-spec.md), [framework](cartulary_modular_refactor_planning_framework.md). | ba75e839; authority, scope and clean main baseline; no docs/decisions directory in tracked tree. | Authoring/authority evidence; missing adopted decisions unresolved. |
| E-002 | `git --no-optional-locks ls-files --cached --others --exclude-standard -z`, in-memory naming/declaration counts and source metadata. | 9,474 initial entries; 2,083 test-shaped paths under stated patterns. | Enumeration/identity only outside specifically inspected groups; RB-002. |
| E-003 | [family catalogs](../../tools/test_families), [verification owners](../../contracts/verification/owners), [catalog owner](../../tools/test_catalog_owner.json). | All rows parsed in memory; 2,611 rows/66 owners and distributions in §2. | Metadata baseline, not asserted conformance. |
| E-004 | [testutil import guards](../../internal/modules/codegenboundary/testutil_import_boundary_test.go). | Walker pruning, no selected-count assertion, broad/subset policies and exemptions. | Confirms source defect and logical subset; no execution. |
| E-005 | [Entities](../../internal/modules/entities/boundary_guard_test.go), [Indicators](../../internal/modules/indicators/boundary_guard_test.go), [Reporting](../../internal/modules/reporting/boundary_guard_test.go), [Revisions](../../internal/modules/revisions/application_boundary_test.go) guards. | Local closed/negative policies, scope, positive selection, rename/alias fixtures inspected. | Preserves distinct local obligations; no aggregate parity claim. |
| E-006 | [Projection adapter boundaries](../../internal/modules/projections/adapters/boundary_guard_test.go), [Graph boundaries](../../internal/modules/graphprojection/boundary_guard_test.go). | Package permissions, production assembly support delegation, legacy selectors; Graph topology/export restrictions. | Unique assurance; legacy selector name-only matching can merit future false-positive investigation. |
| E-007 | [SQLC guard](../../internal/modules/codegenboundary/sqlc_boundary_test.go), [backend policy](../../tools/backend_module_boundaries.json), [backend checker](../../tools/harness/static-analysis/backend-module-boundary-check.mjs). | Global import/export syntax versus Recovery single-file grant; central provider imports and table policies. | Partial overlap; no global replacement equivalence. |
| E-008 | [Projection source ownership](../../internal/app/projectionassembly/source_ownership_test.go), [SQL relations parser](../../tools/harness/static-analysis/sql-relations.mjs). | Descriptor/table equality, fixed source list and regex, separate richer SQL parsing. | Coverage escape concern inferred; no injected escape executed. |
| E-009 | [Frontend policy](../../tools/frontend_import_boundaries.json), [checker](../../tools/harness/static-analysis/frontend-import-boundary-check.mjs), [source ownership](../../apps/web/src/testing/sourceOwnershipPolicy.node.test.ts), [transport](../../apps/web/src/testing/transportBoundaryPolicy.node.test.ts), [inspector](../../apps/web/src/workbook/workbookInspectorOwnershipArchitecture.node.test.ts), [Timeline](../../apps/web/src/workbook/timeline/timelineCompositionArchitecture.node.test.ts), [Network Flow](../../apps/web/src/networkFlow/networkFlowBoundaryPolicy.node.test.ts), [import target](../../apps/web/src/services/importTargetContractAdapter.node.test.ts). | Import/type/symbol/source-string/behavior distinctions inspected. | Specific assertions only; no frontend execution pass. |
| E-010 | [Incident probes](../../apps/web/e2e/incident.support.spec.ts), [CSRF/session helper](../../apps/web/e2e/support/auth/browserSession.ts), [request units](../../internal/modules/incidents/httpapi/request_test.go), [handlers](../../internal/modules/incidents/httpapi/incident_handlers.go), [admission adapter](../../internal/modules/incidents/httpapi/api.go). | Exact two rejection cases, request transport, profile/family selection. | API assurance scope; not UI-client or cookie-policy evidence. |
| E-011 | [Discovery route](../../internal/platform/httpapi/extensiondiscovery/routes.go), [unit tests](../../internal/platform/httpapi/extensiondiscovery/routes_test.go), [service test](../../internal/platform/httpapi/extensiondiscovery/routes_service_test.go), [dispatch unit](../../internal/platform/httpapi/extension_dispatch_test.go), [integration](../../internal/app/serverprocess/extension_dispatch_integration_test.go), [process tests](../../internal/app/serverprocess/incident_membership_process_test.go), [publication tests](../../internal/app/server/extensions_publication_characterization_test.go). | Response/query/auth/session and assembled/profile witnesses compared. | Counterparts are partial; do not treat item count as full payload parity. |
| E-012 | Baseline `apps/web/e2e/coordination-public-route.spec.ts` (retired in S-009; [current survivor](../../internal/modules/workbook/surface_contracts_integration_test.go)). | Seven surfaces, field patches, same-session public role demotion and denial-state checks. | Strong relocation candidate; full redundancy unproven. |
| E-013 | [Workbook surface integration](../../internal/modules/workbook/surface_contracts_integration_test.go). | Fourteen create/query cases; patch matrix only Timeline/Parties; SQL role changes. | Explicit gap relative to E-012. |
| E-014 | [Auth support](../../apps/web/e2e/auth.support.spec.ts), [selector units](../../packages/ui-contracts/src/application-selectors.node.test.ts). | Duplicate bootstrap equality and consecutive feedback attachment; real mounted observations. | Pure duplication versus browser witness distinguished. |
| E-015 | [Incident-selection a11y](../../apps/web/e2e/auth.a11y.spec.ts), [keyboard](../../apps/web/e2e/keyboard.spec.ts), [Timeline public route](../../apps/web/e2e/timeline-public-route.spec.ts), [fixtures](../../apps/web/e2e/fixtures.ts), [runtime origins](../../apps/web/e2e/support/runtime/configuration.ts). | Empty landing and real interaction witnesses; actor/context/session cleanup costs by source. | Supporting-browser label does not imply API-only work; no measured duration. |
| E-016 | [Boundary registrar](../../tools/harness/tests/test-harness-boundary-contracts.mjs), checker [mechanics](../../tools/harness/static-analysis/tests/test-backend-boundary-mechanics.mjs), [policy](../../tools/harness/static-analysis/tests/test-backend-boundary-policy.mjs), [CLI](../../tools/harness/static-analysis/tests/test-backend-boundary-cli.mjs), [Markdown wrapper assurance](../../tools/harness/static-analysis/tests/test-markdown-biome.mjs). | Detector fixtures/registrations versus checks of current source. | Negative/allowed-input and process-contract assurance is distinct. |
| E-017 | Baseline sources `tools/harness/static-analysis/markdown-biome.sh` and `markdown-biome-runner.sh`, recoverable with `git show ba75e83990783f50baa25b5c8ccf94484870e356:<path>`; [configuration](../../biome.markdown.json). | Baseline wrapper cache writes and filename-pattern exclusion identified. Wrapper files were later removed independently; current invocation is E-025. | Historical invocation evidence only; current skip is based on E-025, not an assumption that removed wrappers still run. |
| E-018 | [Task surface owner](../../tools/task_surface_owner.json), [work graph owner](../../tools/harness_work_graph_owner.json), [execution topology](../../tools/execution_topology_manifest.json), [generated policy](../../tools/generated_artifact_policy.json). | Execution/selection ownership and generated-file constraints inspected at relevant entries. | Not exhaustive implementation review of all graph operations. |
| E-019 | [Test-operations skill](../../.agents/skills/cartulary-test-ops/SKILL.md), [command recipes](../../.agents/skills/cartulary-test-ops/references/command-recipes.md), [Harness NLSpec](../testing-harness-nlspec.md). | Canonical selection, distinct obligations, exact-run observation, placement and no-repeat-without-purpose rules. | Procedure, not test-result evidence. |
| E-020 | [Source discovery](../../tools/harness/test-catalog/source-discovery.mjs), [referencepacks](../../referencepacks), [vendored Python tests](../cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/tests), [Playwright config](../../apps/web/playwright.config.ts). | Canonical exclusions include docs/vendor/testdata/generated; external naming census includes vendored tests. | Discovery exclusion is not absence or proof of no value; further registration/project tracing pending. |
| E-021 | [Core 01 §12.2.1](../spec/01_architecture_storage_and_view_contracts.md), [Core 04 AC-428](../spec/04_security_deployment_and_conformance.md), [Harness REQ-654/655](../testing-harness-nlspec.md). | Seven Core operations versus five in subordinate harness restatement. | Primary authority resolved to Core; actual routing completeness unexamined. |
| E-022 | In-memory SHA-256/stat baseline calculation in §1.3; Git optional writes disabled. | 9,474 entries hashed without read errors; Git index and selected ignored-root metadata captured. | Content/file identity accounting, not semantic inspection. |
| E-023 | Read-only review of this completed tracker: in-memory/source inspection of headings, references, rows, and workflow graph. | 12 required sections, 24 unique requirements, 19 findings, 11 contracts, 14 slices, 14 validation obligations, 20 acceptance criteria; 136 repository links resolve; all 14 named selector rows exist; no table-shape errors, unresolved internal references, or workflow cycles. | Editorial structure/traceability only, not canonical lint or product execution. |
| E-024 | Destination absent and ancestors non-symlinked immediately before controlled creation; final source/index/ignored-root comparison excludes this output. | Pre-write HEAD `beaf0293634e6aef9e62aa5670e033a26b6458d3`; clean tree; 9472 entries; content/identity digest `9278072e57a22f47e9d9dd84521681477354041b02ea36677d9da5fbef2d4512`; index SHA-256 `07bae1a6f4f6aee532d28f5b69823f5a3e1c6d221817ec65b01abe2ee2249668`. Final comparison matches all 9,472 non-output content/file-identity records, Git index, HEAD, and selected ignored-root metadata; no capture errors. Only the requested tracker is added. | Original ba75e839 baseline changed independently. This immediate pre-write baseline preserves the completed concurrent edits; ignored descendant contents not exhaustively compared. |
| E-026 | Implementation baseline and accepted plan, 2026-11-09; root/backend/tooling instructions and test-operations skill reread. | HEAD beaf0293634e6aef9e62aa5670e033a26b6458d3; clean tracked/index state; only pre-existing untracked tracker. | Implementation authorized; per-slice sequential checkpoint graph replaces initial grouping. | No product execution result. |
| E-025 | Current [AGENTS](../../AGENTS.md), [tooling instructions](../../tools/AGENTS.md), [test-operations guidance](../../.agents/skills/cartulary-test-ops/SKILL.md), [lint operation](../../tools/harness_work_graph_owner.json), and [Markdown configuration](../../biome.markdown.json); read-only Git comparison with ba75e839. | Concurrent HEAD 40c4d787 plus six modified instruction/skill files observed; user confirmed lint simplification. Direct Biome invocation uses `--write`; wrapper files removed. | Current authoring safeguards and targeted-check stopping rule respected. Audit census is not silently rebased; no concurrent change is claimed as this task's work. |

## 11. Open Questions and Blockers

Current blockers: none within the documented remediation scope. F-050, S-051 and
reopened S-012/S-014 are complete. Fresh evidence and the bounded acceptance claim
are recorded in the final S-051 reconciliation/handoff; earlier blockers remain
historical below.

Blockers below are scoped. They MUST NOT stop independent investigation or
authorize broad repair. Missing information MUST remain explicit until supported
by a cited owner decision or compatible evidence.

| Blocker | Type / issue | Affected decisions | Needed evidence / authority | Status / resolution |
| --- | --- | --- | --- | --- |
| RB-001 | Missing adopted architecture decision text. | S-002/S-005. | Durable owner rules and reachable adoption entries. | DONE; Core 01 §2.1 and Graph §8.1, S-002. |
| RB-002 | Incomplete semantic/registration/requirement coverage accounting. | S-001/S-011/S-012. | Account registrations, purpose, reviewed obligations and Recovery routing without claiming universal coverage. | DONE for scoped accounting/disposition; §2.5 and S-012 identify depth and non-claims. |
| RB-003 | No current runtime qualification for corrections. | S-003/S-004/S-010. | Focused fixtures and outer command exits with compatible terminal artifacts. | DONE; per-slice passing runs. Pre-correction source observations are not rebranded as passes. |
| RB-004 | Replacement selection equivalence. | S-004/S-011/S-014. | Exact obligation/row migrations and direct/aggregate reachability. | DONE; thin Revisions selection, nine explicit row migrations and S-011 routing/catalog checks. |
| RB-005 | Architectural semantic parity. | S-005/S-006/S-007/S-017. | Scope/capability/parser fixtures and distinct preserved guards. | DONE for changed boundaries; qualified checks and retained unique obligations. |
| RB-006 | Complete incident/discovery/dispatch and rendered replacement. | S-008. | Exact errors, full profile, dispatch and landing witness. | DONE; assembled HTTP and browser PASS. |
| RB-007 | Seven-surface public patch/role transition replacement. | S-009. | Seven fields, API demotion, same session and rejection effects. | DONE; expanded surface-contract row PASS. |
| RB-008 | Marginal cost and whole-suite stability unknown. | S-011/S-012. | Bound claims to observed inputs and preserve failures. | DONE as an explicit no-claim disposition; unmeasured global performance/stability does not justify further deletion. |

F-019 was corrected in S-015 under the implementation authorization. The final
canonical lint includes this tracker and passed; the earlier authoring-only skip
is retained solely as historical evidence.

## 12. Binary Completion Criteria

Historical execution outcome before the corrective reopenings: **DONE**. All 17 workstreams, including S-014 final
validation and handoff, are complete. No identified blocking implementation gap
or unsafe retirement remains. The original authoring-only outcomes in
§8.2/§9/§10 remain history; current results are the execution checkpoints and
acceptance matrix below. The current S-047–050 graph supersedes the S-043–046 final handoff; these
historical claims do not renew whole-suite acceptance.

### 12.1 Independent completion claims

**TCR-REQ-024 — Independent completion claims.** Delivery of this document,
completion of whole-suite analysis, readiness of a particular slice, and
implementation completion MUST be reported separately. A deferred requirement
that prevents safe retirement MUST keep that retirement blocked. A complete
rationalization outcome MAY retain tests when evidence supports their distinct
value; it MUST NOT invent removals to meet a count target.

| Outcome | Historical statement before corrective reopenings |
| --- | --- |
| Authorization | Explicit implementation request; all work remains within accepted remediation and separate S-017 corrective scope. |
| Inventory/disposition | Complete accounting and responsibility dispositions for 2,087 source/registration files, 2,609 active rows and 66 owners. Review depth is explicit; not every assertion received a new deep semantic audit. |
| Identified specification/implementation gaps | Repaired or explicitly retained with rationale; F-001..F-022 have decisions and evidence below. |
| Runtime verification | Targeted guards, detector fixtures, SQL/query, Records/Links, frontend consumers, assembled HTTP, browser and routing selections PASS with outer exit 0 as recorded. |
| Global coverage/stability/performance | No exhaustive product requirement-coverage certification, full-suite/release acceptance, fuzz campaign or before/after performance claim. These are not prerequisites for the scoped changes. |
| Final validation/handoff | DONE; canonical Markdown lint (124 files), local-link/acceptance-identity review, Git diff checks and compatible validation reconciliation. S-014. |
| Workspace state | Changes are in the working tree on main; no commit, merge or deployment requested. Initial tracked/index state was clean apart from the supplied tracker. |

### 12.2 Acceptance matrix

| Acceptance | Completion criterion | Evidence / result |
| --- | --- | --- |
| AC-001 | Scope, implementation authorization and adopted-owner precedence remain explicit. | DONE; §1, §6/7, E-026. |
| AC-002 | Every discovered test/helper/registration has depth, accountable context, mechanism and disposition, including noncanonical material. | DONE; S-001 and §2.5 current delta, canonical catalog PASS. |
| AC-003 | Actual purpose is separate from declared runner/evidence/behavior/placement. | DONE; §2.5, corrected Reporting/static metadata and five harness owner migrations; S-011. |
| AC-004 | Every reviewed obligation has evidence or a named limit; specification-free tests retain engineering purposes. | DONE; S-012 matrix and §2.5. No universal requirement-coverage claim. |
| AC-005 | Topology decisions have accessible adopted authority; subordinate prose does not compete. | DONE; S-002/S-013; Core 01 §2.1 and Graph §8.1. |
| AC-006 | Defective scanners select required roots, reject invalid/empty scope and preserve allowed exemptions. | DONE; S-003/S-016 negative and allowed fixtures PASS. |
| AC-007 | Revisions-focused selection remains effective with the broad prohibition; generated families are independently guarded. | DONE by accepted retention decision; S-004/S-016 PASS. |
| AC-008 | Backend package/file/import capability changes preserve reviewed scope and closed policies. | DONE; S-005 policy, SQLC and projection architecture checks PASS; unique local guards retained. |
| AC-009 | SQL source scope, authority, modes, SQLC export and descriptor/table parity remain effective. | DONE; S-006/S-017 parser, query/parity and Records/Links checks PASS. |
| AC-010 | Frontend resolution rejects bypasses and preserves distinct semantic/type/transport/claim obligations. | DONE; S-007 boundary and consumer selections plus TypeScript validation PASS. |
| AC-011 | Incident API relocation preserves both assembled exact error cases. | DONE; S-008 process PASS before retirement. |
| AC-012 | Discovery/dispatch transfer preserves full reviewed profile/error/precedence and real empty landing. | DONE; S-008 assembled and browser PASS. |
| AC-013 | Coordination transfer preserves seven field round trips and public same-session role transition/rejection safety. | DONE; S-009 expanded HTTP contract PASS. |
| AC-014 | Selector cleanup preserves mounted/visible state witnesses. | DONE; S-010 unit and browser PASS. |
| AC-015 | Supporting placement and direct/aggregate/explicit reachability follow adopted policy. | DONE; S-011 19 corrected sentinels and canonical routing/catalog checks PASS. |
| AC-016 | Every cost claim has compatible evidence and limits. | DONE; no unmeasured speed claim; §2.5. |
| AC-017 | Recovery grammar and actual invocation/assertion routing agree on all seven commands. | DONE; S-013/S-012 source review; unchanged process behavior not rerun. |
| AC-018 | Failure/gap conclusions preserve evidence and distinguish deterministic repairs from flakiness. | DONE; S-012 original failure ledger and explicit non-claims. |
| AC-019 | Links, unique acceptance identities, dependencies, generated projections and canonical Markdown validation agree. | DONE; S-014 lint/editorial review and compatible S-011 generated checks. |
| AC-020 | Every retirement has a qualified survivor; all findings and final outcomes are independently accounted. | DONE; S-014 closure, nine active-survivor migrations and separate scope/non-claims. |

### 12.3 Finding dispositions and maintenance handoff

| Findings | Final disposition / owner action |
| --- | --- |
| F-001/F-021 | Corrected root discovery and current generated-family consumer policy; extend explicit assignments and detector fixtures when adding a family or consumer. |
| F-002 | Retain thin Revisions owner selector over shared mechanics; no cross-owner selector indirection or compatibility alias. |
| F-003/F-004/F-016/F-020 | Adopted durable package capabilities, corrected source-policy projections and unique acceptance IDs. Future topology changes start at the owner rule and authored policy. |
| F-005/F-022 | Recursive declaration-aware SQL policy and Records-owned endpoint locking; new mixed-package SQL needs explicit authority. Preserve borrowed transaction and deterministic locking order. |
| F-006/F-007/F-018 | Resolved frontend dependency identity and actual generated-target consumption; keep independent symbol/type/focus/transport/export/parity witnesses. |
| F-008/F-009/F-010 | API probes replaced by qualified assembled HTTP; one rendered empty-directory witness retained. |
| F-011 | Seven coordination API patches/public demotion transferred to qualified assembled contract; API-only browser file retired. |
| F-012 | Pure selector duplicates removed; mounted-state browser checks retained. |
| F-013 | Purpose/classification/placement corrected with exact owner migrations and independent tier sentinels; required gates remain. |
| F-014 | No speculative optimization based on unmeasured marginal cost; retained groups have explicit assurance purposes. Future optimization needs a specific measured hypothesis. |
| F-015 | Core owns Recovery grammar; multi-invocation workflows retain separately asserted command results. |
| F-017 | Source/catalog accounting, family purpose and reviewed-gap/failure dispositions closed; global normative coverage and stability remain explicitly unclaimed. |
| F-019 | Generic refactor-tracker filenames now participate in canonical Markdown lint. |

There is no public API or data migration. Internal constructor consumers now
supply the Records endpoint-lock capability through assembly. Internal policy
inputs and catalog selectors changed atomically with generated projections;
`tools/test_catalog_row_migrations.json` maps all nine retired row identities to
active survivors. Callers should select the new rows, not add compatibility
aliases. No lockfiles, golden images, product timeout budgets or specification
inputs to executable verification were introduced.

Validation handoff reuses compatible completed runs. S-011's latest catalog,
schema, generated and tier-routing results cover the final authored routing;
subsequent changes are prose only. Earlier source checks and runtime selections
remain applicable to unchanged source dependencies. No full-owner/full-suite
rerun, Visual/UI-review, release gate or `agent-finalize` is required by these
changes; no required check is silently skipped. Exact run roots and failures
remain in §10. Roll back each slice's authored inputs, source and generated
projections together; do not restore a retired catalog row without its case.


### Corrective checkpoint: S-023 complete

F-027–029, the superseded GO and the single remaining graph are recorded.
Harness requirements now distinguish exhaustive discovery, bounded evaluation,
shared lexical admission and typed native boundaries. Core capabilities remain
unchanged; prose remains outside executable inputs. `make lint-markdown` passed
(125 files, no fixes). No production interface changed. Residual risks are the
three open defects. Next prerequisite: S-024; readiness remains **NO-GO**.


### Corrective checkpoint: S-024 complete

Discovery inspects assignment RHS/LHS expressions and switch tags/cases at their
program points, with an independent AST walk accounting for unvisited calls and
closures without guessed state. Both consumers require located structured
results with analyzed/rejected/transport dispositions. Closed and statically
empty execution inputs have explicit records; unresolved inputs still fail.
No policy grants or production query behavior changed.

- `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3`
  passed at `.cartulary/test-results/tcr-s024-mechanics-20261109-05`, outer exit 0.
- `make lint CHECK=backend-boundary` passed at
  `.cartulary/test-results/tcr-s024-boundary-20261109-02`, outer exit 0.
- Retained failures: `tcr-s024-red-20261109-01` proves original omissions;
  mechanics `02` exposed child initialization/duplicate reporting; `03` required
  updating a constructor assertion for explicit closed execution; boundary `01`
  exposed an empty error-return alternative now explicitly accounted for.
  Mechanics `04` independently passed the corrected Go assertion.

F-027 acceptance passes parser, backend and projection negative fixtures and the
production backend scan. Remaining risks are lexical templates and native API
coverage. Next prerequisite: S-025; readiness remains **NO-GO**.


### Corrective checkpoint: S-025 complete

The SQL-free `sqltemplate` leaf owns lexical admission shared by runtime
constructors and source inspection. Comments, quoted/partial-token holes, NUL,
unterminated tokens and unqualified Unicode escape forms fail explicitly.
Literal mode preserves literal percent text; formatting retains exact arity and
fragment separators. The SQL classifier distinguishes ordinary and E-string
backslashes. No query builder, capability grant or generated input needed repair.
The existing closed-fragment row exercises the new leaf through its public
constructors, so no duplicate selector or generated routing was added.

All commands below exited zero; roots are under `.cartulary/test-results/`:

| Command / exact selection | Run ID |
| --- | --- |
| `make test-slice OWNER=platform.viewquery ROWS=platform.viewquery.unit.closed_sql_fragments,platform.viewquery.unit.maintenance_sql_capabilities` | `tcr-s025-fragments-20261109-03` |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | `tcr-s025-mechanics-20261109-03` |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership,module.projections.assembly.catalog_and_source_ownership_927c3f08ce,module.projections.query.compiled_surface_iteration3,module.projections.query.contract_shape_and_keyset_59528aa56d` | `tcr-s025-projections-20261109-01` |
| `make lint CHECK=backend-boundary` | `tcr-s025-boundary-20261109-01` |
| `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.reset_preserves_crypto_format_identity,module.recovery.integration.structural_restore_completes_with_independent_co_1bd8029ccb,module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c` | `tcr-s025-recovery-20261109-01` |

`tcr-s025-red-20261109-01` retains the failing runtime reproductions. Fragment `02`
and mechanics `01..02` also passed before NUL/ordinary-string qualification and a
final F-027 assurance extension for unvisited raw assignment/declaration/return
expressions. This extension preserves bounded evaluation and passes mechanics
`03`; final native-gate qualification will recheck production discovery.
F-028 is closed with runtime, policy and affected behavior evidence. Remaining
risk: F-029 native API admission. Next: S-026; readiness remains **NO-GO**.


### S-026 qualification: explicit batch compatibility correction

The typed native inventory reproduces and rejects F-029. Qualification retained
`tcr-s026-red-20261109-01`; an invocation typo (`mechanics-02`) was rejected by
selection admission before execution. Mechanics `03..04` exposed old fixtures
with undeclared receiver names; they now declare driver types. Mechanics `05`
passes. `tcr-s026-inventory-20261109-01` retains initial pinned API inventory
qualification; every exported method is explicitly classified and signature
snapshots were reviewed before adoption.

`tcr-s026-boundary-20261109-01` exposed analyzer false positives for non-SQL
build-variant methods and ordinary fields on embedded transaction wrappers; these
are being corrected without weakening native admission. It also exposed **F-030**:
Reference Data sends a mutable native batch across a callback boundary. Runtime
behavior is valid, but proving arbitrary borrowed batch contents would expand
analysis beyond the selected bounded model. **S-027** will retain domain entries
in the private index builder and construct/execute the native batch locally.
It must preserve query/parameter order, row/byte limits, hashing, error propagation,
borrowed transactions, and owned commit/rollback. No permission exception or
opaque-batch allowance is authorized. Sequence: finish S-026 mechanics → S-027 →
S-026 production acceptance → S-012 → S-014. Readiness remains **NO-GO**.


### S-026 mechanics checkpoint before S-027

The exact native registry covers pgx/pgxpool, pgconn and database/sql methods,
SQL/relation argument indexes and returned handle types. Pinned exported method
signatures are frozen in authored assurance data. Unqualified prepared handles,
borrowed batches, low-level protocols and escaped methods fail. Alias, embedded,
local/global type and unrelated application-method controls pass. Build variants
retain all declared methods instead of choosing the final definition.
`make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3`
passed at `.cartulary/test-results/tcr-s026-mechanics-20261109-07` (outer exit 0).
S-026 remains IN PROGRESS pending the S-027 correction and production gates.
Next authorized workstream: S-027; no batch or transport permission was widened.


### Corrective checkpoint: S-027 complete

Reference Data callbacks carry immutable domain entries, and one private function
constructs and sends each local native batch. SQL, parameter order, batching
thresholds, source hashes and transaction ownership are preserved. Retained rows
copy caller-owned byte/key slices. No public interface, schema or capability
permission changed. F-030 is closed by structural construction, not inference.

`make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.index_batch,module.reference_data.integration.canonical_persistence`
passed at `.cartulary/test-results/tcr-s027-reference-20261109-02`, outer exit 0.
The new exact unit row covers row/byte boundaries, value isolation, parameters,
borrowed transactions, commit and failed-batch rollback. The integration row
qualifies publication and retained generations. `make generate` passed at
`tcr-s027-generate-20261109-01` after the authored row was added.

Retained `tcr-s027-reference-20261109-01` exposed the external-package boundary
of the existing persistence tests; the private unit test now has its own exact
row. `tcr-s027-boundary-20261109-01` exposed absent non-SQL fields in another build
variant; field resolution now accounts for present declarations and rejects
conflicting database identities. S-026 mechanics `08` passes these controls and
native method-expression/global/preparation/batch-escape cases. Next: finish
S-026 production acceptance, then S-012/S-014. Readiness remains **NO-GO**.


### S-026 qualification: test-support transport correction

Production qualification `tcr-s026-boundary-20261109-03` detects eight previously
unseen raw-query forwarders in type-switched test-support functions. **F-031** is
an internal compatibility gap exposed by correct discovery, not a reason to add
transport grants. **S-028** changes their private inputs to closed `Statement`
and read `Fragment` values. Callers retain the same SQL and parameter order;
construction receives existing package/table authority. No policy is widened.
Validate through Entities merge tests exercising all five helper packages and
compiled database branch signatures, plus production boundary enforcement. Earlier boundary
`02` identified unresolved HTTP URL, ranged participant and type-switch types;
declaration resolution now qualifies these and mechanics `10` passes. Sequence:
S-028 → S-026 acceptance → S-012 → S-014. Readiness remains **NO-GO**.


### Corrective checkpoint: S-028 complete

Eight private helpers now accept sealed read fragments or complete statements.
All callers retain their existing SQL and parameter ordering. The helpers perform
only terminal conversion before dispatch; no raw-string forwarding grant was added.
`make test-slice OWNER=module.entities ROWS=module.entities.store.explicit_merge_store_behavior_repoints_live_ment_0560122310`
passed at `.cartulary/test-results/tcr-s028-entities-20261109-01` (outer exit 0),
covering fixture effects across Assessments, Entities, Links, Records and Timeline.
Both driver branches compile; the selected integration behavior uses the managed
PostgreSQL port. `make lint CHECK=backend-boundary` passed at
`tcr-s028-boundary-20261109-01` (outer exit 0). F-031 is closed. Next: finish S-026
protocol/provenance assurance before S-012 reconciliation. Readiness is **NO-GO**.


### Corrective checkpoint: S-026 complete

F-029 is closed. The exact typed registry covers every supported native SQL
argument position, local batch/prepared provenance, acquired handles, aliases,
embedded ports and build variants. Unresolved receivers, method escapes,
method expressions, mutable/borrowed execution state and unqualified protocol
APIs fail. Every admitted native type has an authored exported-signature inventory;
independent call fixtures cover all SQL/copy/prepared and explicitly rejected
surfaces. Prepared/batch reassignment invalidates provenance instead of inferring
state across branches. Non-SQL URL methods, imported Query types and application
methods are qualified through declarations. No permission policy was broadened.

The private SQL protocol now requires exact located records and consistent
analyzed/rejected/transport dispositions. Invalid and missing fields fail adapter
admission. The guide documents this replacement and the supported/rejected API
qualification process. S-019/S-022 corrective acceptance is satisfied; S-012 and
S-014 remain the closure prerequisites.

Current evidence (all outer exits 0; roots `.cartulary/test-results/<run-id>`):

| Exact command | Run ID |
| --- | --- |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | `tcr-s026-mechanics-20261109-13` |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | `tcr-s026-mechanics-20261109-14` |
| `make lint CHECK=backend-boundary` | `tcr-s026-boundary-20261109-05` |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership,module.projections.assembly.catalog_and_source_ownership_927c3f08ce,module.projections.query.compiled_surface_iteration3,module.projections.query.contract_shape_and_keyset_59528aa56d` | `tcr-s026-projections-20261109-01` |
| `make format` | `tcr-s026-format-20261109-03` |

Mechanics `13` includes malformed-response and type-reference controls; `14`
extends the independent native call matrix without changing implementation.
Projection behavior remains compatible after the declaration-only non-SQL type
reference correction; final backend `05` covers all production SQL sources.
Boundary `04` retains that false-positive qualification failure. Mechanics
`09` retains four outdated undeclared-receiver fixtures; `10..12` passed their
repairs and further protocol/provenance cases. Earlier failures are historical,
not erased or reclassified. Next: S-012 coverage/catalog reconciliation.
Readiness remains **NO-GO** until closure evidence and handoff are complete.


### S-012 reconciliation: raw protocol qualification follow-up

Catalog, schema and generated checks pass at the `tcr-s012-final-*` roots.
Reconciliation exposed **F-032**: rejected raw-protocol methods are insufficient
if a caller constructs a protocol client directly and its factory result has no
admitted native type. **S-029** will reject raw-protocol package calls, escaped
constructors and message construction, plus pgconn's raw `Construct` factory.
This enforces F-029's existing default rejection without admitting another API
or changing driver permissions. Targeted analyzer/policy fixtures and production
backend/projection checks must pass before S-012 closes. Readiness is **NO-GO**.


### Corrective checkpoint: S-029 complete

F-032 is closed. Direct/escaped raw protocol package values and alias-based
message/client construction now fail, as does pgconn's raw connection factory.
This closes factory entry paths into F-029's already-rejected protocol surface;
no low-level API was admitted. The normative disposition wording also explicitly
includes unchanged-transport candidates, still requiring the existing exact
policy grant. The guide records these deliberate limits.

`tcr-s029-red-20261109-01` retains four failing direct/escaped factory probes.
The following commands passed with outer exit 0 (all roots under
`.cartulary/test-results/`):

| Exact command | Run ID |
| --- | --- |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | `tcr-s029-mechanics-20261109-01` |
| `make lint CHECK=backend-boundary` | `tcr-s029-boundary-20261109-01` |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | `tcr-s029-projections-20261109-01` |
| `make format` | `tcr-s029-format-20261109-01` |

The final source/fixture analysis and both production gates pass. These changes
have no runtime behavior or generated-input effect. Next: resume S-012 and then
complete S-014. Readiness remains **NO-GO** pending those exits.


### Final corrective checkpoint: S-012 complete

All current findings have fixture, production-policy and affected-behavior evidence.
No remaining exemption, missing selector, unresolved diagnostic or changed behavior
lacks a disposition. S-019/S-022 acceptance remains satisfied after S-029.

| Finding / obligation | Qualification | Current production / behavior evidence |
| --- | --- | --- |
| F-027 exhaustive discovery | `TestAnalyzeSQL`, native construction controls and both-policy fixtures cover fields/indices, nested calls, switches, returns, closures, loops, defer/go, scope/shadowing, escapes and non-SQL controls. Strict adapter fixtures reject missing/malformed/inconsistent located records. | S-029 mechanics, backend and projection runs pass. Unsupported contexts fail explicitly; no empty-success fallback remains. |
| F-028 closed lexical construction | Runtime `TestClosedSQLFragments` and shared-analysis fixtures cover reported comment escape, line/block/nested comments, ordinary/E/dollar/identifier quotes, partial tokens, arity, NUL, composition and adversarial quoted data. | S-025 fragments `03`, Recovery `01`, S-026 projections `01`; final production gates qualify all closed templates. SQL parameters, ordering and keysets remain covered. |
| F-029 native API identity | Independent call matrix qualifies all supported SQL argument positions, copy, local batches, prepared binding and every explicitly rejected method; pinned exported-signature inventory detects drift. Alias/embedded/field/global/type-switch/handle-return and unrelated-method controls pass. | S-029 mechanics and both production gates pass. Existing exact transports and provider restrictions remain unchanged. |
| F-030 Reference Data batches | Exact new `module.reference_data.unit.index_batch` row tests row/byte limits, retained values, SQL parameters, errors and borrowed/owned transactions. | S-027 reference `02` includes canonical publication, historical pins, backup/restore and index rebuild; no additional restore rerun is needed. |
| F-031 test-support inputs | Private adapters take sealed fragments/statements; inspection confirms all five packages retain identical raw SQL. Both driver signatures compile. | S-028 Entities merge and backend `01` pass, covering seeded relationship/source/projection effects through the managed port. No new raw transport grants. |
| F-032 raw protocol factories | Direct/escaped protocol factories, protocol message aliases and raw pgconn construction fail in Go and both-policy fixtures. | S-029 mechanics/backend/projection `01` pass; the rejected surface remains unused by production. |
| Preserved ownership model | Shared-provider, sibling/transitive closure, source-owner locks, maintenance limits, exact transports, wrong tables/operations, stale bindings and rename/split parity retain positive/negative fixtures. | S-029 mechanics and production gates pass; S-021 fact-reader/transaction evidence and S-022 maintenance behavior remain compatible. |

Catalog reconciliation includes the new Reference Data unit row; SQL lexical and
native tests extend existing platform/source-inspection rows. `make generate`
ran after the authored catalog change (`tcr-s027-generate-20261109-01`, exit 0).
No generated output was hand-edited. Source checks discovered through target-plan
passed (outer exits 0, roots `.cartulary/test-results/<run-id>`):

| Exact command | Run ID |
| --- | --- |
| `make source-check CHECK=catalog` | `tcr-s012-final-catalog-20261109-01` |
| `make source-check CHECK=schema` | `tcr-s012-final-schema-20261109-01` |
| `make source-check CHECK=generated` | `tcr-s012-final-generated-20261109-01` |

The subsequent S-029 analyzer/fixture repair changes no catalog/schema/generated
input or output, so these checks remain compatible. Run summaries and retained
secret scans pass; every recorded cleanup outcome is completed without failures.
This was also confirmed for the final mechanics/production selections and retained
S-025 runtime, S-027 reference and S-028 Entities results.

Unchanged browser/HTTP evidence is retained, with passing summaries/scans and
completed cleanup verified: `tcr-s008-http-20261109-01`,
`tcr-s008-browser-20261109-01`, `tcr-s009-http-20261109-01`,
`tcr-s010-unit-20261109-01`, `tcr-s010-browser-20261109-01`. Their exact commands
and assertions are recorded in the historical S-008–010 checkpoints. This is
compatible evidence reuse, not new browser execution or an exhaustive suite claim.
No required selection is skipped. Next: S-014 documentation, final diff and handoff;
readiness remains **NO-GO** until that checkpoint completes.


### Historical corrective checkpoint: S-014 complete — superseded GO

**Historical recommendation: GO; superseded by the F-033–035 review and current NO-GO graph.** F-027,
F-028 and F-029 are closed, as are F-030–032 discovered during qualification.
All corrective workstreams and reopened acceptance/closure exits are satisfied.
No known unresolved implementation, specification, ownership or test-selection
blocker remains. This is targeted implementation readiness, not a full-suite,
release, performance or exhaustive specification-conformance claim.

Final workspace inspection accounts for **521 tracked changes and 34 untracked
files**, with **nothing staged**, on `main`. The footprint includes the original
refactor, DB-port import migration and generated outputs; this corrective effort
preserves that work. Include the untracked source, tests, guide and tracker when
staging. No commit, merge or deployment was performed. Dependency lock files are
unchanged. Authored/generated separation, SQL templates/parameters, batch bounds,
transaction ownership, native/protocol rejection, policy scopes and replacement
selectors were inspected; `git diff --check` passes.

Final canonical enforcement evidence is S-029 mechanics, backend and projection
`01` (exact commands/roots in its checkpoint). Their outer exits are 0, summaries
and retained secret scans pass, and cleanup is completed without failures. The
S-012 matrix identifies compatible runtime and historical browser/HTTP evidence.
The latest formatting run is `tcr-s029-format-20261109-01` (exit 0); catalog,
schema and generated checks are the three `tcr-s012-final-*-20261109-01` roots.
`make lint-markdown` passed over 125 files with no fixes and is rerun after this
record. No required targeted check was skipped. Full-suite, visual, release and
`agent-finalize` runs were not needed for these scoped obligations.

Handoff contract:

- Go discovery is exhaustive independently of bounded evaluation. Located private
  records require explicit dispositions; malformed/legacy responses fail. Native
  execution uses declared identity/signatures and exact arguments, with no
  receiver-name or permissive unknown-method fallback.
- Runtime constructors and inspection share the pure lexical validator. Closed
  templates reject comments, unsafe holes and unterminated/unsupported tokens;
  quoted data, parameters and complete-token composition remain supported.
- The [enforcement guide](../guides/sql-ownership-enforcement.md) documents native
  API dispositions, pinned-inventory review, how to add providers/shared readers,
  exact package/declaration authority, rejected examples and maintenance limits.
- Reference Data callbacks carry immutable domain entries until local batch
  construction. Private test-support adapters take sealed queries/statements.
  No opaque-batch or raw-forwarding compatibility exception was introduced.
- Adopt the private analysis response and callers atomically. There is no public
  API, schema, stored-data or deployment migration. Records locking, source read
  separation, package grants and browser-to-HTTP assertions are retained.
- Deliberate limits remain: finite construction, no arbitrary heap/call-graph
  inference, explicit foreign-library boundaries, immutable local prepared/batch
  provenance, and rejection of unqualified stateful/raw-protocol execution.
  New uses require qualification; they must not inherit broader authority.

Earlier failed qualification runs remain available and are explained in the
workstream checkpoints. Later passes apply to corrected inputs and do not erase
those failures. All requested implementation, reconciliation, validation and
handoff work is complete; staging and committing are the maintainer's next step.


### Corrective checkpoint: S-030 complete

2026-11-10: registered F-033–035 and one authoritative remaining graph; superseded
the historical GO and reopened affected analyzer/classifier and closure acceptance.
Harness completeness now explicitly covers rejected raw builders, canonical alias
identity and scoped query operands. Core grants and the private wire response are
unchanged. The guide records the intentional raw-builder contract narrowing.
Baseline revalidation: 521 tracked/34 untracked, nothing staged, main at beaf029.
Validation: inspected applicable owner clauses, current counterexample sites,
existing selectors and generated policy; documentation lint is retained for final
S-014 after all prose changes. Remaining risks are F-033–035; next prerequisite
is S-031. Readiness remains **NO-GO**.


### Corrective checkpoint: S-031 complete

Removed raw strings.Builder content simulation, including direct writes/resets,
String conversion, fmt.Fprintf and builder-specific alias/escape state. Unknown
output now follows the existing located rejection path at SQL boundaries; child
discovery remains independent. No production builder migration was needed.
Added direct/escaped/method-expression/alias/callback/defer/go builder regressions
and non-SQL/parameter controls to TestAnalyzeSQL; both-policy fixtures include
existing valid source coverage and positive immutable/sealed alternatives.

Exact commands and retained roots (under `.cartulary/test-results/`):

| Command | Run ID | Outcome |
| --- | --- | --- |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | `tcr-s031-red-20261110-01` | Expected test assertion failures in both selected rows; outer exit 2. |
| Same exact mechanics selection | `tcr-s031-mechanics-20261110-01` | Pass, outer exit 0. |
| `make lint CHECK=backend-boundary` | `tcr-s031-boundary-20261110-01` | Pass, outer exit 0. |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | `tcr-s031-projections-20261110-01` | Pass, outer exit 0. |

All four roots have passing retained-secret scans and completed cleanup without
failures. Task-guide and target-plan selections were inspected before execution.
Private response and permissions are unchanged. F-033 is closed; remaining risks
are F-034/F-035. Next prerequisite: S-032. Readiness remains **NO-GO**.


### Corrective checkpoint: S-032 complete; F-036 requires S-034

Shared alias resolution now preserves declaration scope, alias/defined-type
identity, pointers, fields and returned handles. Method admission explicitly
distinguishes non-SQL signatures from unresolved/rejected boundaries; removed
the separate nonSQLMethod heuristic. Canonical parameter/result aliases, local
and imported scopes, build conflicts/cycles/missing declarations, generic data
results, defined native-derived results and non-SQL function fields are qualified.
The independently versioned Reference Pack format module is explicitly foreign;
its existing backend runtime-dependency check remains required. Response wire
format and grants are unchanged.

S-031's exact mechanics selection was run as `tcr-s032-mechanics-20261110-01`
through `06`. Run 01 exposed a missed old identity-helper caller (build/preflight
failure); 02 exposed a fixture counting discovered literals as executions;
03–05 passed. Run 06's Go row passed all current type/inventory fixtures, while
the boundary row exposed the generated-input integration gap described below.
Backend runs `tcr-s032-boundary-20261110-01` and `02` failed on data/function-field
classification and omitted generated declarations; `03` passes (outer exit 0).
Projection SQL-only runs `tcr-s032-projections-20261110-01` failed on those data
false positives; `02` and `03` pass, with `03` outer exit 0. Current successful
production roots have passing secret scans and completed cleanup. Failed runs
remain retained; they are implementation qualification evidence, not regressions
in product runtime behavior.

**F-036 — Generated SQL scope omission (P2, must_fix).** Adding generated source
to shared analysis exposed that projection/Records/Timeline table rule scopes
excluded generated packages. A negative fixture deleting host_grid_projection
from generated source passed despite complete extraction. S-034 must extend
those enforcement scopes without granting authority, qualify generated access
and missing/linked inputs, and update CLI fixtures to provide the now-required
generated source root. Specification/guide clarify that authored-source excludes
are not SQL exclusions. Central policy changes require schema validation.

This separate corrective slice is a new prerequisite of S-033 and closure. The
current table is authoritative: S-030 → S-031 → S-032 → S-034 → S-033 → S-012 →
S-014. F-034 is closed at the type/signature boundary; overall acceptance remains
**NO-GO** until F-035/F-036 and closure pass. Next prerequisite: S-034.


### Corrective checkpoint: S-034 complete

F-036 is closed. Generated Go is an explicit additional SQL source collection,
with missing/linked/nonregular inputs rejected and build-constrained files included.
Backend policy scopes for projection tables, Records envelopes and Timeline source
now include generated packages. No read/write/lock grant was added. Generated SQL
passes through construction, controls and table policy enforcement; it is not
unexamined type metadata. CLI fixtures supply the required generated source root.
Regression fixtures prove alias resolution, rejected generated writes and linked
input rejection. Ordinary authored-source exclusion behavior remains separately
qualified.

Retained roots under `.cartulary/test-results/`, all pass with outer exit 0,
passing secret scan and completed cleanup without failures:

| Command | Run ID |
| --- | --- |
| S-031 exact two-row mechanics selection | `tcr-s034-mechanics-20261110-01` |
| `make lint CHECK=backend-boundary` | `tcr-s034-boundary-20261110-01` |
| `make source-check CHECK=schema` | `tcr-s034-schema-20261110-01` |

Target-plan expansion for schema was inspected. No generator input/output changed;
the central policy JSON remains authored. F-033/F-034/F-036 are closed. F-035 is
the remaining corrective prerequisite. Next: S-033; readiness remains **NO-GO**.


### Corrective checkpoint: S-033 complete

F-035 is closed. Shared classification now separates statements, CTE bindings and
set operands with explicit command positions and relation/lock scopes. TABLE
operands after WITH and UNION/INTERSECT/EXCEPT are classified, including ALL,
DISTINCT, parentheses, qualified names and recursive/ordinary CTE visibility.
INSERT source queries are classified independently of their write target; missing
operands, unsupported heads and stray TABLE positions fail explicitly. No runtime
query or SQL classification wire interface changed.

Retained roots under `.cartulary/test-results/`:

| Command | Run ID | Result |
| --- | --- | --- |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries` | `tcr-s033-red-20261110-01` | Expected pre-fix assertion failures; outer exit 2. |
| S-031 exact two-row mechanics selection | `tcr-s033-mechanics-20261110-01` | Pass, outer exit 0. |
| `make lint CHECK=backend-boundary` | `tcr-s033-boundary-20261110-01` | Pass, outer exit 0. |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership,module.projections.assembly.catalog_and_source_ownership_927c3f08ce,module.projections.query.compiled_surface_iteration3,module.projections.query.contract_shape_and_keyset_59528aa56d` | `tcr-s033-projections-20261110-01` | Pass, outer exit 0. |

All current passing roots have passing retained-secret scans and completed cleanup
without failures. The four-row Projection expansion was inspected before execution.
Expected access inventories and both-policy fixtures qualify forbidden physical
reads alongside permitted CTE shadowing and relation-free queries. Earlier SQL
write/lock/maintenance/transport/rename fixtures continue passing. No product
query migration was needed. Remaining work is coverage reconciliation and final
review/documentation. Next: S-012; readiness remains **NO-GO** pending closure.


### Corrective checkpoint: reopened S-012 complete

| Obligation | Current qualification | Evidence and applicability |
| --- | --- | --- |
| F-033 mutable construction | TestAnalyzeSQL raw-builder matrix; both-policy rejected builders and accepted immutable/sealed/parameter controls. | S-033 mechanics and both production gates pass; no runtime builder migration. |
| F-034 signature identity | TestSQLImportedDatabaseTypes canonical signature matrix plus native API inventory; both-policy alias counterexamples. | Current Go and JS selected rows pass. Equivalent aliases, unsupported defined database results, missing/cyclic/conflicting types, fields/handles and non-SQL controls are represented. |
| F-035 query scope | Independent expected-access inventories for WITH/TABLE, every set operator/modifier, parenthesized/recursive/qualified operands, INSERT sources, writes and locks; both-policy decisions. | S-033 mechanics and full four-row Projection selection pass. |
| F-036 generated SQL | Strict generated collector, positive generated aliases, forbidden generated write, build-constrained input, missing root and symlink fixtures; exact scope additions. | S-034 mechanics/schema and S-033 current production gates pass; no capability grant added. |
| Preserved policy | Wrong providers/tables/operations, sibling/shared helpers, exact transport/control shapes, maintenance and rename/split controls remain selected. | Existing owner-boundaries suite passes on final semantic inputs. |
| Product behavior | Compiled projection surfaces, query/keyset contracts and assembly/catalog parity. | S-033 four-row selection passes; no runtime source or query was edited. |
| Routing and outputs | New tests are registered inside existing TestAnalyzeSQL/TestSQLImportedDatabaseTypes and owner-boundaries registrars; no new top-level selector. | `make source-check CHECK=catalog`, root `tcr-s012-catalog-20261110-01`, passes with outer exit 0, passing scan and completed cleanup. |

Re-inspected compatible retained roots: `tcr-s025-fragments-20261109-03`,
`tcr-s025-recovery-20261109-01`, `tcr-s027-reference-20261109-02`,
`tcr-s028-entities-20261109-01`, `tcr-s009-http-20261109-01`,
`tcr-s008-http-20261109-01`, `tcr-s008-browser-20261109-01`,
`tcr-s010-unit-20261109-01`, `tcr-s010-browser-20261109-01`, and
`tcr-s012-final-generated-20261109-01`. Their exact commands/rows remain in the
earlier checkpoints. Summaries/scans pass and cleanup is completed. These remain
historical evidence for unchanged runtime, generated and browser/HTTP obligations;
they do not substitute for current analyzer or policy qualification. Session
content comparison confirms those production sources/tests/generated outputs
were untouched.

The policy JSON is authored and does not change generated producers, schemas or
selector inputs; generation and another generated-output check are unnecessary.
Schema validation is current at `tcr-s034-schema-20261110-01`. Formatting passes
at `tcr-s014-format-20261110-01` (pre-handoff preparation, outer exit 0).
`git diff --check` passes. No unresolved finding or missing selector remains.
Next prerequisite: S-014 final documentation/workspace review; readiness remains
**NO-GO** until it completes.


### Final review reopening: F-035 declaration-level parentheses

S-014 source review found that SQL operand qualification at execution sites did
not cover a standalone Go declaration containing `(TABLE users)`: Go nested-query
discovery recognized SELECT/WITH only. Reopened S-033 and S-012/S-014 to extend
that existing finding's end-to-end acceptance. Add declaration-level expected
access fixtures and qualify the actual inspector plus both gates again. No runtime
SQL migration is involved. Readiness remains **NO-GO**.


### Corrective checkpoint: S-033 declaration acceptance restored

Extended Go nested-query discovery to TABLE/VALUES. Declaration-level TABLE and
expression-nested TABLE expected-access fixtures pass through the real inspector.
Fresh roots `tcr-s033-mechanics-20261110-02` (exact two-row selection),
`tcr-s033-boundary-20261110-02` (backend lint) and
`tcr-s033-projections-20261110-02` (SQL declaration ownership row) pass with outer
exit 0. The other three Projection rows retain compatible S-033 run 01 evidence.
No runtime or policy input changed in this follow-up.

### New corrective prerequisite: S-035 / F-037

**F-037 — Non-leading database results (P2, must_fix).** Final type review and a
read-only probe against the current inspector confirmed zero SQL expressions for
a Reader.Query signature returning `(bool, Rows, error)` with Rows aliasing
pgx.Rows. The signature classifier inspected only the first result. S-035 must
classify every result, preserving aliases and rejecting unresolved/derived native
results, and qualify positive/negative cases in both policies. This is analysis,
tests, specification clarification and guide work; it grants no capability and
changes no runtime interface. Left unresolved, an interface result reordering
hides unqualified SQL. Exit: all result positions participate, ordinary data
results remain non-SQL, and current mechanics/production gates pass.

The authoritative graph now includes S-035 after S-033 and before S-012/S-014.
Those closure slices remain reopened; readiness is **NO-GO**.


### Corrective checkpoint: S-035 complete

F-037 is closed. Repository-port admission now examines every result field,
including grouped/named results. Shared canonical resolution preserves aliases;
unresolved or newly defined native-derived results reject. The recognized result
inventory includes pgconn readers, queued queries and preparation descriptions,
and pinned native API assurance checks every non-error SQL result. Existing
preparation/batch provenance and exact transport restrictions remain unchanged.
The private response, runtime APIs and ownership grants are unchanged.

Retained roots under `.cartulary/test-results/`:

| Command | Run ID | Result |
| --- | --- | --- |
| S-031 exact two-row mechanics selection | `tcr-s035-red-20261110-01` | Expected pre-fix result-position assertion failures; outer exit 2. |
| Same exact mechanics selection | `tcr-s035-mechanics-20261110-01` | Summary/scan pass, cleanup completed; outer exit was not retained in the session handoff. |
| Same exact mechanics selection | `tcr-s035-mechanics-20261110-02` | Pass, outer exit 0; repeated narrowly to retain independent owning-command completion. |
| `make lint CHECK=backend-boundary` | `tcr-s035-boundary-20261110-01` | Pass, outer exit 0. |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | `tcr-s035-projections-20261110-01` | Pass, outer exit 0. |
| `make format` | `tcr-s035-format-20261110-01` | Summary/scan pass, cleanup completed; outer exit was not retained in the session handoff. |

All listed roots have passing retained-secret scans and completed cleanup with
no failures. Final mechanics run 02, backend and Projection roots supply current
invocation evidence. No product behavior changed; the other three Projection
rows retain compatible S-033 run 01 evidence. F-033–037 are closed. Next prerequisite:
reopened S-012 evidence reconciliation; readiness remains **NO-GO** until closure.


### Final coverage reconciliation: S-012 complete after S-035

This checkpoint supersedes the earlier reopened S-012 assessment. F-033–037
have positive controls and negative regressions in the selected Go/JavaScript
mechanics, plus current backend and projection policy execution. F-035 includes
declaration-only parenthesized TABLE; F-037 includes later/named result positions
and all pinned native SQL result types. No stale grant, permissive fallback,
unexplained exemption, missing selector or unqualified changed behavior remains
in this corrective scope.

Current analyzer/policy evidence is `tcr-s035-mechanics-20261110-02`,
`tcr-s035-boundary-20261110-01` and `tcr-s035-projections-20261110-01`, with
independently observed outer exit 0, passing summaries/scans and completed cleanup.
S-033's four-row Projection run 01 retains the unchanged assembly, compiled-query
and shape/keyset coverage. S-034 schema and S-012 catalog inputs remain compatible;
no new top-level test or routing input was introduced. The previously listed ten
runtime/generated/browser/HTTP roots were re-inspected: summaries and scans pass,
cleanup is completed without failures. Those runs qualify unchanged obligations,
not the corrected analysis.

The content comparison against the S-030 baseline identifies 17 changed paths:
three specification/guide/tracker documents, six sourcecheck implementation/test
files, the authored backend policy, the backend checker and six JavaScript
classifier/fixture files. All other tracked and non-ignored untracked contents
match the session baseline. No runtime source, generated output, dependency or
selector changed. Generation remains unnecessary; no full-suite, visual, release
or agent-finalize execution is required for these changes.

Final workspace before documentation closure: main at beaf029, 522 tracked
changes (515 modified, seven deleted) and 35 untracked files, nothing staged.
The existing changes are preserved. All corrective acceptance is complete;
next prerequisite is S-014 final documentation and handoff. Readiness remains
**NO-GO** until that final check completes.


### Final corrective handoff: S-014 complete, 2026-11-10

**Historical recommendation: GO, superseded by S-036 NO-GO.** F-033–035
and follow-ups F-036/F-037 have regression and production-check evidence. S-019,
S-024 and S-026 acceptance is restored; S-030–035 and reopened S-012/S-014 are DONE.
There are no known unresolved findings in the approved corrective scope. This
assessment covers targeted implementation and policy obligations, not full-suite,
visual, release or conformance acceptance. No commit, merge or deployment occurred.

Current retained validation, with roots under `.cartulary/test-results/`:

| Exact command | Retained run ID | Obligation |
| --- | --- | --- |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | `tcr-s035-mechanics-20261110-02` | Go discovery/construction/signatures/native inventory, SQL operands and both-policy positive/negative regressions. |
| `make lint CHECK=backend-boundary` | `tcr-s035-boundary-20261110-01` | Current authored/generated SQL, package/table/control/transport policy and runtime dependency boundaries. |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | `tcr-s035-projections-20261110-01` | Current provider dependency closure, declaration and table authority. |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership,module.projections.assembly.catalog_and_source_ownership_927c3f08ce,module.projections.query.compiled_surface_iteration3,module.projections.query.contract_shape_and_keyset_59528aa56d` | `tcr-s033-projections-20261110-01` | Compatible unchanged assembly, compiled query and shape/keyset behavior; its older SQL row is superseded above. |
| `make source-check CHECK=schema` | `tcr-s034-schema-20261110-01` | Current authored structured policy. |
| `make source-check CHECK=catalog` | `tcr-s012-catalog-20261110-01` | Current selectors and ownership routing. |
| `make format` | `tcr-s014-final-format-20261110-01` | Final formatting and independently retained command completion. |

Every listed invocation has observed outer exit 0, a passing summary and retained
secret scan, and completed cleanup without failures. Final formatting was selected
to retain owning-command completion missing from S-035's formatting handoff;
no semantic implementation changed. `make lint-markdown` passes (125 files, no
fixes); `git diff --check` passes. Earlier qualification failures and incomplete
outer-exit records remain documented rather than overwritten. Compatible unchanged
runtime, maintenance/restore, browser/HTTP and generated evidence is enumerated
in S-012; it is not claimed as freshly executed.

The final guide describes the construction and migration contract:

- Raw strings.Builder output no longer qualifies as SQL, even when local. Use
  finite immutable expressions or sealed sqlshape fragments. Non-SQL text and
  bound values remain valid. No production caller needed migration this round.
- Shared declaration-aware alias resolution determines native/port identity.
  Every result position participates. Missing, cyclic or conflicting identity
  cannot establish non-SQL behavior; defined types do not inherit native grants.
  Pinned native API assurance now also checks SQL result identities.
- The classifier separates statement, CTE and set-operation scopes, including
  TABLE/VALUES and declaration-only parenthesized TABLE. Unsupported operands
  reject. Generated Go participates in construction and access enforcement;
  missing/linked inputs fail. No capability grant or scan exclusion was widened.
- Provider/shared-reader additions still require explicit source entries, exact
  package/declaration authority, per-provider dependency qualification and owner
  behavior evidence. The [guide](../guides/sql-ownership-enforcement.md) contains
  the maintained procedure. No compatibility alias or lexical fallback was added.
- The private analysis wire response is unchanged in these slices. Earlier
  atomic private-interface migrations remain documented above. There is no public
  API, database schema, stored-data or deployment migration.

Deliberate limits remain bounded construction and query syntax, explicit foreign
library boundaries, no arbitrary heap or call-graph inference, and rejection of
unqualified stateful/native protocol execution. The architecture gate does not
prove every PostgreSQL semantic property. New surfaces require explicit
qualification; widening grants or silently treating unresolved inputs as data
is not a migration path.

Final state: main at `beaf0293634e6aef9e62aa5670e033a26b6458d3`, 522 tracked
changes and 35 untracked files, nothing staged. The final content comparison
preserves all pre-existing work outside the 17 recorded corrective paths.
No generated output, dependency lockfile or product runtime source changed in
this session. Generation, full suites, visual/release checks and agent-finalize
were not required and were not launched. Handoff is complete; staging and committing
remain maintainer actions.


### S-036 completion: corrective authority, 2026-11-10

Recorded the unchanged HEAD/index and all tracked/non-ignored untracked content
in `/tmp/tcr-s036-baseline`; counts are 522 tracked changes and 35 untracked files,
nothing staged. Current GO is superseded by NO-GO. F-038–040 and reopened acceptance
have one authoritative graph. Harness and guide now state the agreed stable-capture,
direct-call/port and universal dot-import admission contracts. Core and machine
policies are unchanged; the located response needs no protocol migration.

Validation: read authority and applicable repository/tooling/test-operations
instructions; inspected current implementation and retained evidence limitations.
Documentation lint is deferred to final documentation closure, not claimed passed.
Prior passes do not qualify these new counterexamples. Remaining risks are callable
alias/escape completeness and conservative capture admission; S-037 is next.


### S-037 completion: canonical callable admission, 2026-11-10

F-039 is implemented through shared callable classification for invocation and
value transfer. Inline, local/imported aliases, defined function types, pointers,
fields, parameters, indices and returned values retain identity; every result
position and extracted factory-method signature participates. SQL callable values
reject with located diagnostics. Direct native/port methods and ordinary callbacks
remain distinct; type declarations are not execution. Generic domain type parameters
retain declared identity. No production caller, grant or wire protocol changed.

Discovery: `make task-guide OWNER=harness.command_surface DETAIL=rows JSON=1` and
`make target-plan TARGET=test-slice INPUTS=/tmp/tcr-s036-baseline/mechanics-selection.json DETAIL=rows JSON=1`.
The exact selection is
`make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3`.
Current evidence: `tcr-s037-mechanics-20261110-06` and
`make lint CHECK=backend-boundary`, `tcr-s037-boundary-20261110-04`.
Both owning commands exited 0; summaries/scans pass and cleanup completed without
failures. All roots are under `.cartulary/test-results/`.

Retained qualification history: `tcr-s037-red-20261110-01` failed the new bypass
fixtures (exit 2). Mechanics 01 exposed constructor/conversion ambiguity; 04 exposed
lost returned callable identity (exit 2). Mechanics 02/03/05 passed (exit 0), superseded
by 06. Backend 01 exposed ordinary fields and generic domain identity; 03 exposed
unresolved dereference/type confusion (exit 2). Backend 02 passed (exit 0), superseded
by 04. These were analyzer qualification defects, repaired without production
exemptions. All retained scans pass and cleanup completed without failures.

Remaining obligation: capture stability is not established by callable discovery.
S-038 is next; overall readiness remains NO-GO. Existing runtime and source-capability
separation evidence remains compatible; formatting/documentation closure is pending.


### S-038 completion: stable closure inputs, 2026-11-10

F-038 now uses a declaration-identity inventory before evaluation. Captures require
one explicit reaching initialization and no reassignment/address escape/nested
writer; constants retain authority and unknown values remain unknown. Nested writers
and addresses invalidate raw SQL at outer sinks too. Closures receive isolated
value/type environments and no inherited prepared/batch provenance. Ordinary DB
identity remains intact. No scheduling or heap inference was introduced. The private
inspection protocol and exact transparent-transport dispositions are unchanged.

Two production construction changes were necessary: Links copies each finite
validation statement into an immutable local before its row callback; Recovery
captures sealed sqlshape fragments for its rebuild and validation queries. SQL text,
parameters, transaction ownership, result/error handling and operation order are
unchanged. No grants, scopes, selectors or dependencies were added.

Current retained evidence (all roots under `.cartulary/test-results/`):

| Exact command | Run ID | Result |
| --- | --- | --- |
| S-037 exact two-row harness selection | `tcr-s038-mechanics-20261110-04` | Pass, observed outer exit 0. |
| `make lint CHECK=backend-boundary` | `tcr-s038-boundary-20261110-02` | Pass, observed outer exit 0. |
| `make test-slice OWNER=module.links ROWS=module.links.store.bundle_validation_commit_isolation,module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1` | `tcr-s038-links-20261110-01` | Pass, observed outer exit 0; source round-trip, failure atomicity and transaction scratch isolation. |
| `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.reset_preserves_crypto_format_identity` | `tcr-s038-recovery-20261110-01` | Pass, observed outer exit 0; both rebuild algorithms, rejected invalid restore and reset identity. |

Owner task guides and target-plan expansions were inspected before execution;
selection/plan snapshots are in `/tmp/tcr-s036-baseline`. Scans pass and cleanup
completed without failures. Retained failures: `tcr-s038-red-20261110-01` reproduces
stale captures; mechanics 01 exposes local-batch overrejection; backend 01 exposes
the two production builders and accidental callback-parameter diagnostic drift.
All exited 2. The transport diagnostic was restored, local batches remain qualified,
and affected callers were simplified. Mechanics 02/03 passed (exit 0), superseded
by 04's additional global/unknown/ordinary-defer controls. All retained scans and
cleanup pass. S-039 import admission is next; readiness remains NO-GO.


### S-039 completion: unambiguous import admission, 2026-11-10

F-040 is implemented at batch admission: every dot import fails with file, line,
column and package before SQL analysis. No package exception list exists. Named
imports/aliases and import-only parsing retain their behavior. Fixtures cover driver,
protocol, constructor, helper and ordinary-library imports, generated/build-constrained
paths, message encoding/writes and factory extraction. Qualified raw-protocol controls
still reject. The authored/generated production collections pass admission; a source
survey across internal/cmd/db found no production dot imports to migrate.

After `make format` (`tcr-s039-format-20261110-01`), current validation is:

| Exact command | Run ID |
| --- | --- |
| S-037 exact two-row harness selection | `tcr-s039-mechanics-20261110-01` |
| `make lint CHECK=backend-boundary` | `tcr-s039-boundary-20261110-01` |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | `tcr-s039-projections-20261110-01` |

All four owning commands exited 0; summaries and retained scans pass; cleanup
completed without failures. Target-plan expansions and owner rows were inspected.
Both consumers qualify F-038–040 alongside existing construction, native inventory,
read/write/lock, transport, maintenance, provider/closure and file-move fixtures.
No schema/protocol/routing/generated input changed; generation is unnecessary.
The baseline comparison identifies exactly 12 changed paths, including two new
private analyzer files and the two S-038 runtime migrations. Other pre-existing
content is preserved. Next: S-012 coverage reconciliation. Readiness remains NO-GO.


### S-012 completion: corrective coverage reconciliation, 2026-11-10

F-038–040 now map to current Go/JavaScript qualification and both production gates:

| Finding / obligation | Positive and negative qualification | Current evidence |
| --- | --- | --- |
| F-038 capture stability | Deferred/concurrent/escaped/immediate callbacks; writes before/after capture, branch/loop/nested writes, address escapes, globals, named results, field/index state, scope/shadowing, immutable copies, unknown input, sealed values, local versus captured prepared/batch handles and ordinary defer arguments. | TestAnalyzeSQL capture subtests; both-consumer capture fixture; S-039 mechanics/backend/Projection runs. |
| F-039 callable identity | Inline/local/imported/chained aliases and defined signatures; fields, parameters, locals, pointers, indices, all returned positions, extracted factories and parentheses; missing/cyclic identity and ordinary callback controls. Existing canonical-identity/build-variant and native inventory fixtures remain selected. | TestAnalyzeSQL callable subtest; TestSQLImportedDatabaseTypes; both-consumer callable fixture; S-039 gates. |
| F-040 import admission | Driver/protocol/constructor/helper/ordinary dot imports; generated/build-constrained paths; protocol encoding and escaped factories; qualified controls and separate import-only parsing. | TestAnalyzeSQL dot-admission subtest and TestParseImports; shared adapter/backend fixtures; complete production collections in S-039 gates. |
| Preserved policy | Wrong-provider/table/read/write/lock authority, shared-reader closure, exact transports, maintenance permissions, generated SQL and same-package rename/split parity. | Existing owner-boundaries fixtures plus current S-039 backend and Projection checks. |
| Runtime migration | Links validation SQL/parameters, source round-trip/failure atomicity/scratch isolation; both Recovery rebuilds, invalid restore rejection and reset identity. | S-038 Links and Recovery owner-selected runs, outer exits 0. |
| Routing | Existing top-level source-inspection tests contain new cases; no new registration or authored/generated selector. | `make source-check CHECK=catalog`, `tcr-s012-corrective-catalog-20261110-01`, outer exit 0. |

Current roots are recorded under S-038/S-039. Catalog summary/scan pass and cleanup
completed without failures. The 12-path content comparison preserves all earlier
work outside this scope. No new exemption, widened authority, protocol/schema change,
missing selector or unresolved production diagnostic remains. No generated input
changed, so generation/schema reruns are unnecessary.

Compatible historical evidence was re-inspected (summary/scan pass; cleanup complete):
`tcr-s033-projections-20261110-01` for unchanged assembly/compiled/query-keyset rows;
`tcr-s034-schema-20261110-01` for unchanged machine policies;
`tcr-s025-fragments-20261109-03` for unchanged closed-template runtime;
`tcr-s025-recovery-20261109-01` for unchanged maintenance surfaces (the changed rebuild
path has fresh S-038 evidence);
`tcr-s027-reference-20261109-02`, `tcr-s028-entities-20261109-01`,
`tcr-s009-http-20261109-01`, `tcr-s008-http-20261109-01`,
`tcr-s008-browser-20261109-01`, `tcr-s010-unit-20261109-01`,
`tcr-s010-browser-20261109-01` and `tcr-s012-final-generated-20261109-01`
for unchanged product/generated obligations. Their original commands and owning
exits remain in earlier checkpoints; these are retained compatibility claims,
not freshly executed or proof of the corrected analyzer. Qualification failures
remain recorded in S-037/S-038. No full-owner, full-suite, visual or release acceptance
is claimed. Next: S-014 documentation, final diff and readiness. Overall NO-GO remains
until that exit completes.


### Corrective admission: F-041 / S-040, 2026-11-10

**P2, must_fix:** final S-012 review found that `QueryFunc[T any]
func(context.Context,string,...any)(T,error)` instantiated with `pgx.Rows` is
classified as data. A read-only sourcecheck probe returned `Read.sql=[]` for
`query(ctx,input)`. Generic domain identity added in S-037 does not establish a
non-SQL result at an ambiguous SQL-signature boundary. This reopens F-039 callable
acceptance and S-012/S-014; the preceding S-012 checkpoint is superseded.

Areas: shared Go identity/callable admission, tests, Harness clarification and guide.
S-040 must distinguish unbound type parameters from proven data at callable
boundaries, reject ambiguous raw-SQL signatures, and inspect function constraints.
Do not introduce generic specialization/call-graph inference or a grant exception.
Ordinary generic domain callbacks without a possible raw SQL argument remain data;
direct SQL in their bodies is still checked. This preserves bounded analysis while
preventing generic spelling from erasing SQL authority. Compatibility is private
analysis; newly ambiguous signatures must use concrete ports/types. Leaving this
unresolved permits the F-039 bypass through a generic type. No runtime exploit is
established. Exit: generic result/constraint fixtures reject in both consumers,
non-SQL controls and complete production checks pass. Next graph is
S-040 → S-012 → S-014; readiness remains NO-GO.


### S-040 completion: generic callable qualification, 2026-11-10

F-041 and reopened F-039 acceptance are implemented. Canonical callable admission
now inspects function constraints and rejects function values with a possible raw
SQL argument and an unbound result, including later results, pointers, generic
conversions and extracted generic functions. Ordinary constrained/domain callbacks
retain non-SQL identity. Direct repository helper calls remain analyzed through
their bodies under declaration authority; extraction does not inherit that
classification. No generic specialization, grant, exclusion or compatibility
fallback was added. No additional production caller or private protocol changed.

Current evidence, all with observed owning-command exit 0:

| Exact command | Run ID |
| --- | --- |
| S-037 exact two-row harness selection | `tcr-s040-mechanics-20261110-02` |
| `make lint CHECK=backend-boundary` | `tcr-s040-boundary-20261110-02` |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | `tcr-s040-projections-20261110-02` |
| `make format` | `tcr-s040-format-20261110-03` |

All summaries/scans pass; cleanup completed without failures. Retained
`tcr-s040-red-20261110-01` fails the generic bypass fixtures (outer exit 2).
Mechanics 01 passes, while backend/Projection 01 reject direct generic domain
helpers incorrectly (outer exits 2); the shared direct-call/value distinction
was corrected and all three 02 runs pass. Formatting 01/02 also passed (exit 0),
superseded by 03. Scans and cleanup pass for every retained qualification run.
The exact source-inspection selectors are unchanged; the new cases are subtests
of the existing registered tests. S-012 reconciliation is next; NO-GO remains until
S-014 closure. Earlier S-038 runtime and historical compatible evidence is unaffected.


### Final S-012 reconciliation after S-040, 2026-11-10

The earlier S-012 matrix remains valid and is extended by F-041's generic result,
constraint, conversion and extracted-function regressions plus ordinary generic
callback/direct-helper controls. Current analyzer and policy evidence is S-040
mechanics/backend/Projection 02, each with independently observed outer exit 0.
`make source-check CHECK=catalog` passed again as
`tcr-s012-final-catalog-20261110-02` (outer exit 0, passing summary/scan, completed
cleanup without failures). No selector or generated-output change was needed.

F-038–041 have current negative and positive coverage and no unexplained production
diagnostic. Links/Recovery behavior evidence from S-038 remains compatible: S-040
changes only analysis and prose. The twelve compatible historical roots listed in
the prior reconciliation retain unchanged obligations and do not establish current
analyzer acceptance. All earlier failed/incomplete evidence remains historical.

Exactly 12 paths differ from `/tmp/tcr-s036-baseline/hashes.json`; all other tracked
and non-ignored untracked contents match. Final readiness remains NO-GO pending
S-014 documentation and handoff. No additional corrective slice is currently known
necessary, and no permissions, scan exclusions or protocol fallback were added.


### Corrective admission: F-042 / S-041, 2026-11-10

**P2, must_fix:** handoff review found a second capture-state channel in finite-list
analysis. `tables := []string{"artifacts"}; defer func(){ for _, table := range tables
{ tx.Exec(ctx,"SELECT * FROM "+table) } }(); tables=[]string{input}` reports only
analyzed Artifact SQL. The read-only sourcecheck probe returned no unresolved input.
The restricted scalar environment did not restrict the collection environment.
This reopens F-038 acceptance and S-012/S-014.

Areas: Go binding/collection state, tests, Harness clarification and guide. S-041
must isolate every closure evaluation environment, discard inherited collection
inference, and reject raw collection inference after nested mutation, address or
alias escape. Mutable field/index state must not retain synthetic values. Keep
local finite literal collections only where their use is bounded; do not add heap
alias inference. Rationale: one stability contract must govern every evaluator
state channel. Compatibility: use immutable scalar copies or sealed fragments for
SQL callbacks; no grants or protocol changes. Otherwise callbacks or escaped lists
can substitute unauthorized relations while checks observe an earlier value.
Exit: capture/list-alias/nested-writer regressions and both consumers reject, local
finite/non-SQL controls and production gates pass. Next graph is
S-041 → S-012 → S-014; NO-GO remains in force.


### S-041 completion: isolated collection evaluation, 2026-11-10

F-042 is closed and F-038 acceptance restored. Closure analysis now isolates scalar,
type, collection and return environments. Inherited collections cannot supply raw
SQL alternatives. Local finite-list inference rejects alias/address exposure,
nested writers and mutable fields/indices; ordinary bound data remains supported.
No runtime source, capabilities or private response changed in this slice.

`make format` (`tcr-s041-format-20261110-01`), the exact two-row harness selection
(`tcr-s041-mechanics-20261110-01`), `make lint CHECK=backend-boundary`
(`tcr-s041-boundary-20261110-01`) and the exact Projection SQL row
(`tcr-s041-projections-20261110-01`) passed with independently observed outer exits
0, passing summaries/scans and completed cleanup without failures. The red fixture
run `tcr-s041-red-20261110-01` is retained (exit 2, expected counterexample failures,
passing scan and cleanup). No additional production migration was needed.

### Corrective admission: F-043 / S-042, 2026-11-10

**P2, must_fix:** a lexical-scope audit found a local constant in a labeled block
could fall back to a same-named package constant after discovery cleared value
state. With package `table="artifacts"` and local `table="users"`, the read-only
probe reports only analyzed Artifact SQL. This is an identity error in unknown-state
fallback, not a new SQL permission. Areas: Go constant identity, tests and guide.
S-042 must allow package fallback only for actual package declarations (or unresolved
cross-file references resolved by the package inventory); locals without evaluation
state must reject. Benefit: unsupported control flow cannot change lexical binding.
Compatibility: private analysis only; unsupported builders must simplify rather than
inherit outer constants. Otherwise scoped names can conceal unauthorized relations.
Exit: supported shadowing remains correct, unknown-state shadowing rejects in both
consumers, cross-file/grouped constants and production gates pass. No capability
expansion is allowed. S-042 → S-012 → S-014 is the remaining graph; NO-GO remains.


### S-042 completion: constant fallback identity, 2026-11-10

F-043 is closed. The package-constant inventory records declaration objects;
unknown evaluation state cannot substitute a package constant for a local constant
with the same name. Cross-file/package/grouped constant behavior stays qualified.
Unsupported local evaluation contexts produce located unresolved input. No runtime
caller, grant, exclusion, routing or response protocol changed.

The exact two-row harness selection (`tcr-s042-mechanics-20261110-01`),
`make lint CHECK=backend-boundary` (`tcr-s042-boundary-20261110-01`), the exact
Projection SQL row (`tcr-s042-projections-20261110-01`) and `make format`
(`tcr-s042-format-20261110-01`) all passed with observed owning-command exit 0,
passing summaries/scans and completed cleanup without failures. The red run
`tcr-s042-red-20261110-01` retains the expected scope counterexample failures
(outer exit 2; scan and cleanup pass). This closes the last known corrective
finding. S-012 reconciliation is next; overall readiness remains NO-GO until S-014.


### Final S-012 reconciliation after S-042, 2026-11-10

F-038–043 have current positive/negative regression coverage. The previous coverage
matrix is extended by S-041's inherited-list/alias/nested-writer/field-state cases
and S-042's unknown-state lexical shadowing cases, each exercised through Go and
both policy consumers. Current mechanics/backend/Projection evidence is S-042 run
01, each with an independently observed successful outer exit. The runtime changes
remain exactly the S-038 Links/Recovery constructions and their selected passing
behavior evidence remains compatible. No later runtime input changed.

`make source-check CHECK=catalog` passed as `tcr-s012-final-catalog-20261110-03`,
outer exit 0, passing summary/scan and completed cleanup without failures. The
registered source-inspection tests contain all new cases; no new selector,
generated output or machine protocol was introduced. Earlier compatible evidence
and failed qualification histories remain retained as enumerated above.

All pre-existing content outside the same 12 corrective paths remains byte-for-byte
unchanged. No unresolved finding, unexplained production diagnostic, unqualified
changed behavior or widened authority remains known. Final workspace: main at
beaf029, 523 tracked changes and 37 untracked files, nothing staged. S-014 is ready;
readiness remains NO-GO until documentation and final handoff complete.


### Final S-014 handoff after S-042, 2026-11-10

**Historical recommendation: GO, superseded by S-043 NO-GO.** F-038–043
are closed with regression and production evidence. All corrective workstreams and
reopened S-012/S-014 are DONE; affected earlier analyzer/native acceptance is restored.
No known unresolved finding remains in this corrective scope. This assessment does
not claim full-owner, full-suite, visual, release or conformance acceptance.
Commit, merge and deployment were not performed.

The supported contract and migration instructions are maintained in the
[SQL enforcement guide](../guides/sql-ownership-enforcement.md):

- Capture stability uses declaration identity and an independent inventory of
  initialization, writes and escapes. Unknown values stay unknown. Closure scalar,
  type, collection and return state is isolated; captured prepared/batch provenance
  and mutable collection inference do not carry across the callback boundary.
- SQL function values reject through one canonical callable path, including aliases,
  defined/generic types, constraints, fields, parameters, indices, pointers and
  returned/extracted values. Direct native calls/ports and existing exact transports
  remain qualified. Repository helper bodies retain their own declaration analysis;
  extracting a helper does not inherit direct-call admission.
- SQL-analysis batches reject all dot imports, including generated and constrained
  inputs. Import-only parsing remains separate. Unknown-state local constants cannot
  inherit same-named package constants. Unsupported inputs fail explicitly.
- Links captures an immutable statement copy; Recovery captures sealed rebuild and
  validation fragments. Parameters, SQL text, ordering, borrowed transaction identity,
  result/error behavior and restore/reset effects retain targeted behavioral coverage.
- Provider/shared-reader additions still require explicit entries, complete dependency
  closure, per-provider authority, exact package/declaration capabilities and affected
  owner behavior evidence. No grant, subtree permission, exclusion or forwarding
  exemption was broadened. No compatibility fallback was introduced.
- The located private inspection response is unchanged. Internal Go analysis helpers
  changed atomically. No public API, schema, stored-data or deployment migration is
  required. No structured policy, selector, dependency lockfile or generated artifact
  changed in these slices.

Final retained validation is below. Every command was launched from the repository
root with `CARTULARY_TEST_RUN_ID=<listed ID> CARTULARY_OUTPUT_MODE=ci`. Each result root
is `.cartulary/test-results/<listed ID>`. Owning-command exits were observed
independently; every listed run exited 0, reports pass, has a passing retained-secret
scan and completed cleanup with zero failures.

| Exact command | Run ID | Obligation |
| --- | --- | --- |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | `tcr-s042-mechanics-20261110-01` | Go discovery/binding/callable/capture/import/constant qualification, native inventory, SQL classifier and both-policy regressions. |
| `make lint CHECK=backend-boundary` | `tcr-s042-boundary-20261110-01` | Complete current authored/generated SQL and package/declaration/transport authority. |
| `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | `tcr-s042-projections-20261110-01` | Complete provider closure and source/declaration/table authority. |
| `make test-slice OWNER=module.links ROWS=module.links.store.bundle_validation_commit_isolation,module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1` | `tcr-s038-links-20261110-01` | Changed validation callback: source round-trip, rejected-write atomicity and transaction scratch isolation. |
| `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.reset_preserves_crypto_format_identity` | `tcr-s038-recovery-20261110-01` | Changed rebuild construction: both algorithms, invalid restore rejection and reset identity. |
| `make source-check CHECK=catalog` | `tcr-s012-final-catalog-20261110-03` | Current owner/selector closure. |
| `make format` | `tcr-s042-format-20261110-01` | Final implementation formatting. |

`make lint-markdown` passes for 125 files with no fixes; `git diff --check` passes.
Exact owner rows and target-plan expansions were inspected before execution.
Historical qualification failures remain recorded at their originating checkpoints,
including red fixtures and corrected false positives. Earlier compatible assembly,
compiled-query, keyset, closed-template, maintenance, generated and browser/HTTP
roots are enumerated in S-012. They are reused only for unchanged obligations;
older analyzer passes do not substitute for the S-042 checks.

Final workspace assessment: main at
`beaf0293634e6aef9e62aa5670e033a26b6458d3`, 523 tracked changes (516 modified,
seven deleted) and 37 untracked files, with nothing staged. The full tracked and
non-ignored untracked content inventory was compared with the S-036 baseline.
Only these 12 corrective paths differ; all other pre-existing content is preserved:

- Harness NLSpec, this tracker and the SQL enforcement guide.
- Links source-state port and Recovery PostgreSQL rebuild construction.
- Sourcecheck sql.go, sql_types.go, sql_resolution.go and sql_test.go, plus new
  private sql_callable.go and sql_capture.go.
- The shared JavaScript projection/backend SQL regression fixture.

The final diff review covers the authority changes, shared admission paths,
independent negative/positive fixtures, runtime migrations and preservation of the
remaining workspace delta. Deliberate limits remain bounded Go/SQL analysis,
explicit foreign-library boundaries, no arbitrary heap/call-graph/generic
specialization, and rejection of unsupported stateful execution. Passing this
architecture gate is not a proof of every PostgreSQL semantic property.

No schema or generation check was required because their authored inputs and
protocols did not change; compatible existing schema/generated evidence is retained.
No full suite, visual/release run or agent-finalize was needed for the selected
obligations. Handoff is complete; staging and committing remain maintainer actions.


### S-043 completion: execution-state authority, 2026-11-10

Current baseline and the three reported defects were re-inspected. The Harness
clarification covers complete joins, construction-time collection snapshots,
complete assignment transfer and conservative repeated execution. Current readiness
is NO-GO. Planning is complete; implementation is authorized by the current user
request. S-044 is active; no current passing implementation evidence exists yet.
Make help, both task guides and exact target-plan expansions were inspected.


### S-044 completion: complete branch state, 2026-11-10

Scalar, collection and type/provenance snapshots now join together across if and
switch alternatives. Collection elements retain construction-time values; branch
order cannot select the only retained list. Closure/discovery isolation and scope
restoration use the same state representation. Existing finite bounds remain.

The exact two-row mechanics selection failed as expected in
`tcr-s044-red-20261110-01` (outer exit 2; both Go and policy fixtures) and passed
in `tcr-s044-mechanics-20261110-01` (outer exit 0). Both have passing retained scans
and completed cleanup. Assertions target execution locations and both policies;
positive controls include var declarations, scoped lists and finite bounds.
F-044 is qualified at mechanics level; current production qualification belongs
to S-012 after the remaining corrections. S-045 is next; readiness stays NO-GO.


### S-045 completion: complete assignment transfer, 2026-11-10

Assignments, short declarations and var/const declarations now share evaluation
before destination writes. Every non-blank destination receives a result or
explicit unknown state; synthetic fields and collection state cannot survive
replacement. Unsupported later result positions reject without tuple inference.
Native result identity and local provenance rules remain separately qualified.

The exact mechanics rows retained the expected counterexample failures in
`tcr-s045-red-20261110-01` (outer exit 2; Go and both-policy fixtures), followed by
`tcr-s045-mechanics-20261110-01` (outer exit 0, pass). Both scans and cleanup pass.
Cases cover reassignment, :=, var, later/blank positions, collection replacement,
parallel values and bound-parameter controls. S-046 is next; readiness is NO-GO.


### S-046 qualification review, 2026-11-10

The original loop counterexamples failed in `tcr-s046-red-20261110-01` (outer
exit 2) and passed in `tcr-s046-mechanics-20261110-01` (exit 0). The expanded
matrix run `tcr-s046-mechanics-20261110-02` failed on two fixture defects: a missing
separator in the combined-file fixture and an expectation that unchanged parameter
forwarding had a rejected disposition rather than the existing transport
disposition. The corrected matrix passed in `tcr-s046-mechanics-20261110-03`.
Unqualified forwarding is independently rejected by both policy consumers.

Review then found two companion state obligations: preserve a binding's declared
port identity across assignment to different concrete implementations, and emit
unresolved compound SQL before loop-exit state is discarded. Focused Go-row runs
`tcr-s046-composition-red-20261110-01` and `tcr-s046-port-red-20261110-01` retain
these failures (exit 2). The corrections share assignment transfer; both-consumer
fixtures now cover them. These refine F-044/F-046 rather than add SQL capabilities.
S-046 remains IN_PROGRESS pending final qualification.


### S-046 completion: conservative repeated execution, 2026-11-10

Loop writes are inventoried before conditions/body/post qualification. Finite ranges
use construction-time element alternatives; outer mutations remain unknown after
zero or repeated iterations. Separate break/continue exits prevent unreachable
writes from restoring values. Initializer scopes, closure isolation, and unsupported
jump/fallthrough handling share the state rules. Assignment retains declared Go
port types and emits compound SQL before state disposal.

The final expanded two-row mechanics run `tcr-s046-mechanics-20261110-04` passes
with independently observed outer exit 0. It covers F-044–046 through both consumers,
exact sink locations, limits, helper rename/split, native provenance, control-flow
interactions and immutable/sealed/bound-data controls. Formatting run
`tcr-s046-format-20261110-03` also exits 0. No runtime caller or authority grant has
changed. S-012 production and coverage qualification is next; readiness is NO-GO.


### S-012 qualification correction: ordinary Go data, 2026-11-10

The first production gates failed in `tcr-s046-boundary-20261110-01` and
`tcr-s046-projections-20261110-01` (outer exits 2). They exposed analyzer false
positives on named results, type-switch fields, range values and a boolean `Query`
field. No runtime SQL migration or authority exception was warranted. The focused
Go regression `tcr-s046-data-red-20261110-01` preserves the failure (outer exit 2).

Named results now start with their declared types; missing inference at a join
does not conflict with a known static type. Selectors are observed in their
lexical type context, and type-switch bindings use complete scope restoration.
Conflicting known types and unqualified native provenance still reject. Both
consumers have ordinary-data positive controls alongside the negative port,
assignment and loop cases. This correction refines F-044–046 qualification.

`tcr-s046-mechanics-20261110-05` reports pass with a passing scan and completed
cleanup; its outer exit was unavailable after native-session output truncation,
so it is not final invocation evidence. The final source, including type-switch
scope restoration, passes the complete targeted two-row selection in
`tcr-s046-mechanics-20261110-06` with independently observed outer exit 0.
`tcr-s046-format-20261110-05` exits 0. Production qualification follows; previous
failures remain retained and are not reclassified as flakiness.


### Final S-012 reconciliation after S-046, 2026-11-10

F-044–046 are closed in the bounded analysis contract. The following matrix binds
requirements to observations at actual execution sites and to each policy decision.
The source-inspection suite calls the new private helpers from the existing
`TestAnalyzeSQL`; JavaScript fixtures remain registered under owner boundaries.
No catalog rows, acceptance IDs, capability grants or response schema changed.

| Obligation | Implementation/fixture witness | Current qualifying evidence |
| --- | --- | --- |
| F-044: all continuing collection alternatives | Complete state joins; both branch orders and switch orders; omitted alternatives; empty/unknown/unequal/record collections; construction snapshots; scopes/returns; finite limits; helper rename/split. Artifact reads and backend protected-table deletions fail under their distinct policies. | Mechanics 06; backend 02; projections 02 below. |
| F-045: every destination replaces prior state | Shared parallel assignment transfer for `=`, `:=`, var/const; blank/second/third results; collection overwrite; branch interaction; swaps/repeated targets; exact sink assertions. Unsupported tuple positions become unknown. | Mechanics 06 through Go and both policy consumers. |
| F-046: repeated boundaries see conservative values | Body mutation after execution, loop condition/post/continue, nested/range loops, existing range variables, after-loop use, break/return and unsupported jump/fallthrough handling. Compound SQL survives state disposal. | Mechanics 06 through Go and both policy consumers. |
| Preserve qualified behavior | Immutable and per-iteration expressions, sealed fragments, bound data, directly initialized prepared/batch handles; static port identity; named results and ordinary type-switch/range fields. | Mechanics 06; both production gates 02. |
| Routing/accounting | Existing two mechanics rows and projection architecture row still register the new helpers/cases; no generated or authored routing inputs changed. | Catalog 04, reused for unchanged registration/routing obligations only. |

Private implementation responsibilities are explicit: `sql_state.go` owns deep
snapshots, joins, collection values and scope restoration; `sql_assignment.go`
owns destination transfer; `sql_flow.go` owns write inventories and conservative
loop flow. `sql.go` coordinates AST discovery and boundaries; `sql_capture.go`
shares state isolation and binding facts. `sql_state_test.go` adds execution-site
qualification without creating another public test selector or analyzer protocol.

Both production gates now pass without changing callers, grants or exclusions.
Links' immutable capture and Recovery's sealed fragments are unchanged from this
task's baseline. Their historical runtime runs are not fresh evidence for the
changed analyzer and are not used to close these counterexamples. No public API,
database, deployment or stored-data migration is needed. Newly rejected raw SQL
patterns must use the guide's immutable/sealed/bound-data forms; there is no legacy
acceptance switch or exception. Domain vocabulary and normative SQL authority are
unchanged; the Harness clarification owns analysis completeness.

All roots in this checkpoint are under `.cartulary/test-results/`. Exact commands
run from the repository root with `CARTULARY_OUTPUT_MODE=ci` and the listed literal
`CARTULARY_TEST_RUN_ID` are:

| Run ID | Command / scope | Observed result |
| --- | --- | --- |
| `tcr-s046-mechanics-20261110-06` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | Outer exit 0; pass; 5/5 work units. |
| `tcr-s046-boundary-20261110-02` | `make lint CHECK=backend-boundary` | Outer exit 0; pass. |
| `tcr-s046-projections-20261110-02` | `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | Outer exit 0; pass. |
| `tcr-s012-catalog-20261110-04` | `make source-check CHECK=catalog` | Outer exit 0; pass; reused unchanged catalog obligations. |
| `tcr-s046-format-20261110-05` | `make format` | Outer exit 0; pass; subsequent source checks use the formatted source. |

The final mechanics, backend, projection and format manifests have source digest
`sha256:bb738e92303237594691f9e806193cc594523444f1b44f23abe039c835e4236c`,
commit `beaf0293634e6aef9e62aa5670e033a26b6458d3`, dirty source state and toolchain
digest `sha256:0e302b723a654ff49d1df64d0c655f27c266127f9f49480b0ea7a88b56be8893`.
Catalog 04 has the earlier source digest
`sha256:456142fb96aa1b630c43703bf632658f44bf29da6068d30b9826c81e3a487d3b`;
subsequent changes affect analyzer internals, cases inside existing registrations,
and prose, so this reuse supplies routing evidence only. No whole-owner evidence
audit is claimed.

All 24 corrective roots explicitly listed in S-044–046 and this S-012 checkpoint
have passing retained-secret scans and only completed cleanup outcomes, including
failed reproductions. Failure summaries retain `product/test_assertion_failure`:
S-044/S-045/S-046 red roots reproduce the reported defects; composition/port/data
red roots preserve companion qualification defects; mechanics 02 records fixture
errors; production gate 01 roots record ordinary-data false positives. Later passes
close their repaired obligations; they do not erase failures or establish flakiness.
Intermediate format roots 01–04 and mechanics roots 01–05 remain historical; only
the final compatible selections above support current implementation acceptance.

S-012 is DONE. S-014 must finish canonical Markdown lint and the final workspace
preservation review before current readiness changes from NO-GO.


### Final S-014 handoff after S-046, 2026-11-10

**Historical recommendation: scoped GO; superseded by S-047 NO-GO.** F-044, F-045 and F-046 are closed. S-043–046,
coverage reconciliation S-012 and handoff S-014 are DONE. The previous GO assessment
was reopened before implementation; this assessment relies on the final corrective
runs, including the ordinary-data qualification repair, rather than prior passes.

The final S-012 table records the exact passing Make commands, source identities
and independently observed outer exits. All final runs have passing retained scans
and completed cleanup. `make lint-markdown` checked 125 files with no fixes and
outer exit 0; this closing evidence-only/status update receives a final repeat.
`git diff --check` passes. Documentation records completeness obligations,
conservative rejection, migration choices and actual validation limits; executable
checks do not consume specification, guide or tracker prose.

The final baseline comparison covers all 9,509 original tracked/nonignored entries
and four new files. Exactly eleven paths differ: the Harness spec, this tracker,
the SQL enforcement guide; `sql.go`, `sql_capture.go`, `sql_test.go`; new private
`sql_state.go`, `sql_assignment.go`, `sql_flow.go`, `sql_state_test.go`; and the
shared `test-projection-sql.mjs` fixtures. All other baseline contents and file
identities match. The Git index digest remains
`3df1e41fe672cb561b7cd191b83cd37cff1090ea5fb4f15cd89f82367aac9247`;
HEAD is unchanged and nothing was staged, committed, merged or deployed.
The pre-existing 523 tracked changes and 37 untracked files were preserved;
four new private Go files bring the untracked-file count to 41. Ignored test/cache
artifacts are outside that source-preservation inventory and are retained as
normal harness outputs.

No runtime, owner grants, routing inputs, output schemas, generated outputs,
dependencies or public interfaces changed in this correction. Schema/generation
checks and agent-finalize were therefore unnecessary. No runtime service suite,
full suite, visual check, release acceptance, universal SQL soundness proof or
performance/stability claim is made. The architecture gate remains deliberately
bounded: unsupported mutable/aliased/raw tuple or loop-carried SQL must reject
or be rewritten into the qualified forms documented in the guide. Passing these
checks does not establish a production exploit path or certify every SQL semantic
property. Broader acceptance is not required to close these targeted defects.


### S-047–050 implementation and qualification, 2026-11-10

S-047 authority and red qualification are complete. `tcr-s047-red-20261110-01`
ran the exact two-row mechanics selection and exited 2; both rows report
`product/test_assertion_failure`, with passing secret scan and completed cleanup.
The regressions demonstrate omitted helper exits, deferred literal replacement
and missing/linked source inputs. A new fixture also incorrectly required unique
policy violations across statement and execution records; its assertion was
corrected to test input-path deduplication only. No retained failure is erased.

S-048/S-049 use a private return-summary context with result binding, finite values,
completeness and explicit reasons. Unsupported control admission is scoped to the
helper, nested callbacks have independent summaries, and named-result effects
apply even to explicit literals. Existing bounds/native result identity remain
separate. S-050 extracts common path/directory/file admission while preserving
collector authority. Missing-root CLI fixtures now materialize support directories.
New IO failure fixtures are deterministic; public inspection/CLI exercise real
symlinks and a nonregular socket source. Basename-only source exclusions are removed.

`tcr-s050-mechanics-20261110-01` and final extended `02` pass, with observed outer
exits 0, passing scans and completed cleanup. They cover the original state cases,
new return/effect cases, nested calls, rename/split, empty results and actual
execution/composition decisions through both policies. Format roots `01` and `02`
exit 0. The projection production gate `tcr-s050-projections-20261110-01` passes.

Backend production gate `tcr-s050-boundary-20261110-01` exits 2 with six
`sql_construction_unresolved` diagnostics in two runtime helper families: Links'
`sourceQuery` and Network Flow's `rowSQLExpression`. Both return raw string tuples,
which the accepted contract deliberately no longer infers. This is a required
caller migration, not a reason to relax admission. Both now return sealed fragments
through their existing error/bool contracts; callers retain parameter bindings,
query operations, ordering and borrowed transaction handling. Network Flow's live
SQL parity helper changes only its terminal text conversion. No capability grant,
exclusion, policy schema or public API changes. Production gates and narrowly
selected Links/Network Flow runtime rows must qualify this migration before S-012.
Readiness remains NO-GO.


### Final S-012 reconciliation after S-050, 2026-11-10

F-047–049 and S-047–050 are complete. The first backend production run correctly
identified raw tuple callers; the required Links and Network Flow sealed-fragment
migrations now pass production and affected runtime qualification. No table grant,
transport exception, exclusion, schema, dependency or generated output changed.

| Obligation | Implementation and qualification | Evidence |
| --- | --- | --- |
| F-047 complete return paths | Private `sql_return.go` summary; explicit/bare/zero-value results, finite alternatives, unsupported control rejection, loops, closure isolation and nested-call/rename/split sink assertions. | Mechanics 02; final consumers 03; backend 02; projection 02. |
| F-048 result effects | Existing binding inventory checks named destinations regardless of return expression; nested/deferred/concurrent/escaped writers and address aliases reject; cleanup and immutable controls pass. | Mechanics 02 and consumers 03 through both policies. |
| F-049 source admission | Private `source-admission.mjs` supplies normalized paths and strict directory/file reads to backend, migration, support, generated, projection and exact-set collectors. Real link/socket/missing fixtures and deterministic IO failures; no basename-only skips; sorted/deduplicated root inputs. | Consumers 03, including public inspection/CLI and existing strict-collector regressions; backend 02 and projection 02. |
| Caller migration | Links exports/compares typed query fragments; Network Flow filter/order/keyset and graph contributor expressions remain sealed. Status results, parameters, ordering and transaction handling are retained. | Links 01; Network Flow 01, including live directional SQL parity and graph contributor queries. |
| Routing and generated projections | New cases remain inside existing Go and shell registrations; no owner, selector, policy-schema or generator input changed. | Reuse catalog `tcr-s012-catalog-20261110-04` and unchanged schema/generated evidence below. |

All roots below are under `.cartulary/test-results/`. Each command used
`CARTULARY_OUTPUT_MODE=ci` and the literal listed `CARTULARY_TEST_RUN_ID`; owning
command exits were observed independently. Every passing run below also has a
passing retained-secret scan and only completed cleanup outcomes.

| Run ID | Exact command | Result and scope |
| --- | --- | --- |
| `tcr-s050-mechanics-20261110-02` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | Exit 0; shared Go analyzer and all three gap regressions. |
| `tcr-s050-consumers-20261110-03` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries` | Exit 0; final consumer/discovery fixtures, including sealed tuple migration controls. |
| `tcr-s050-boundary-20261110-02` | `make lint CHECK=backend-boundary` | Exit 0; production authored/generated SQL and package capabilities after migration. |
| `tcr-s050-projections-20261110-02` | `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | Exit 0; provider dependency/SQL closure after migration. |
| `tcr-s050-links-20261110-01` | `make test-slice OWNER=module.links ROWS=module.links.store.bundle_validation_commit_isolation,module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1` | Exit 0; bundle round-trip, invalid-write atomicity, validation/commit isolation. |
| `tcr-s050-networkflow-20261110-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.pagination_recovery,module.networkflow.integration.bounded_graph_contributor_pipeline` | Exit 0; PostgreSQL filter/order/keyset parity, pagination recovery and bounded contributor queries. |
| `tcr-s050-format-20261110-03` | `make format` | Exit 0; final Go/runtime formatting. |

Source compatibility is explicit: mechanics 02 has digest
`sha256:c34f06276249edb4a1d9682866d6baf41069b39998a4939a83ba6372869aae9b`.
The analyzer and admission implementation are unchanged afterward. Backend 02,
Links 01, Network Flow 01 and format 03 have digest
`sha256:1ba372d37e66ad4de0f935b6f536d14f1414959013057f2234134fcbfc2f2c3a`.
Projection 02 has digest
`sha256:00b1cc7b3c5c8abd0b59cfa77fa2a05d05a20ad337dd5d5187343c7bca2e7c6f`.
Subsequent changes are a consumer-fixture type inventory and its negative assertion,
plus prose. The runtime, collectors and analyzer did not change. Thus no unrelated
runtime or Go rerun is needed to repair the JavaScript fixture. The final consumer
run supplies current fixture evidence; older passes do not substitute for it.
All these runs have commit `beaf0293634e6aef9e62aa5670e033a26b6458d3`, dirty source
state and toolchain digest
`sha256:0e302b723a654ff49d1df64d0c655f27c266127f9f49480b0ea7a88b56be8893`.

Failure history: S-047 red 01 and S-050 backend 01 retain the reproduced defects
and required caller migration. Consumers 01 omitted the sealed Fragment type from
its synthetic source batch, correctly triggering unresolved callable identity.
Consumers 02 then expected a table-write diagnostic for a prohibited read-fragment
write; construction rejects that earlier. The final fixture instead proves a
sealed unauthorized read receives the actual table-read denial in both policies.
These are fixture corrections, not analyzer relaxations or flaky outcomes.
All failed roots have passing secret scans and completed cleanup. Intermediate
mechanics/format/projection passes remain historical, with their original scope.

#### Pending-change evidence reconciliation

The full 9,513-entry baseline preserves all earlier implementation and routing
changes outside this remediation. The historical S-012 obligation matrix remains
applicable only to unchanged behavior. The following 28 exact retained roots were
re-inspected: every summary/secret scan passes and every cleanup outcome is completed.
Their outer exits remain the historical observations in the originating checkpoints;
this is read-only evidence review, not new execution or a full-owner evidence audit.

| Unchanged obligation group | Compatible retained roots |
| --- | --- |
| Records locks, Timeline/Links/Evidence/Assessment facts, vocabulary, assembly adapters | `tcr-s021-records-20261109-02`, `tcr-s021-timeline-20261109-02`, `tcr-s021-links-20261109-01`, `tcr-s021-vocabulary-20261109-01`, `tcr-s021-evidence-20261109-01`, `tcr-s021-assessment-20261109-01`, `tcr-s017-adapters-20261109-01` |
| Runtime consumers, reset and source mutations | `tcr-s022-networkflow-20261109-01`, `tcr-s022-reset-20261109-01`, `tcr-s022-artifacts-20261109-01`, `tcr-s022-tasks-20261109-01`, `tcr-s022-reporting-20261109-01`, `tcr-s022-indicators-20261109-01` |
| Sealed runtime/maintenance, reference data, entity operations, compiled projection queries | `tcr-s025-fragments-20261109-03`, `tcr-s025-recovery-20261109-01`, `tcr-s027-reference-20261109-02`, `tcr-s028-entities-20261109-01`, `tcr-s033-projections-20261110-01` |
| Prior immutable-capture and sealed Recovery migrations | `tcr-s038-links-20261110-01`, `tcr-s038-recovery-20261110-01` |
| Generated artifacts, structured policies, catalog/routing | `tcr-s012-final-generated-20261109-01`, `tcr-s034-schema-20261110-01`, `tcr-s012-catalog-20261110-04` |
| Browser-to-HTTP replacement and retained rendered observations | `tcr-s008-http-20261109-01`, `tcr-s008-browser-20261109-01`, `tcr-s009-http-20261109-01`, `tcr-s010-unit-20261109-01`, `tcr-s010-browser-20261109-01` |

The S-050 Links and Network Flow rows supersede earlier runs for newly changed
query builders. Current analyzer/admission decisions rely on S-050, never earlier
runtime or architecture passes. No generated/authored routing input changed;
regeneration, schema/catalog reruns and agent-finalize are unnecessary. Broader
product, browser/visual, full-owner, release and performance acceptance remain
unclaimed. The documented pending changes have scoped evidence; universal
specification completeness or SQL soundness is not asserted.

S-012 is DONE. S-014 must finish canonical Markdown lint and the final preservation
review before readiness changes from NO-GO.

Final consumer 03 source digest: `sha256:763c89840646cd333de3f37ded07ebe1942b9cb885dbfda0440af2d88fb2ac59`.


### Final S-014 handoff after S-050, 2026-11-10

**Historical recommendation, superseded by F-050: scoped GO for committing the documented pending change set.**
The three identified P2 enforcement defects are closed. S-047–050, S-012 and
S-014 are DONE. The current graph and blocker rollup supersede prior GO/NO-GO
checkpoints; historical evidence and failures remain intact. No staging, commit,
merge or deployment was performed.

The S-012 tables bind each obligation to its implementation, regression and exact
Make invocation. Both policy consumers reject omitted returns and deferred/escaped
result replacements; source admission rejects partial inputs before completeness.
Qualified controls preserve named/explicit returns, harmless defers, sealed values,
ordinary bound data, native type identity and existing branch/assignment/loop rules.
The architecture gate remains bounded and does not claim general SQL or Go soundness.

Links and Network Flow now carry sealed fragments through their tuple-returning
builders. All callers were updated atomically; targeted PostgreSQL tests validate
source export/compare, failure atomicity, transaction isolation, live directional
query/keyset parity and graph contributors. There is no external API, schema,
stored-data or deployment migration. Future raw tuple callers must adopt the
same documented pattern; no legacy switch or new authority grant exists.

All 15 explicitly identified S-047/S-050 invocation roots were inspected. Eleven
report pass and four retain the documented assertion failures. Every retained
secret scan passes; every cleanup outcome is completed. The 28 compatible earlier
roots listed in S-012 were also checked. Their evidence is limited to unchanged
obligations; the migrated builders use fresh S-050 runtime evidence. No mandatory
targeted check is skipped. Generation, schema/catalog reruns and agent-finalize
are unnecessary because their authored inputs and registrations are unchanged.
No full suite, visual run, whole-owner audit, release acceptance or performance
claim is made.

`make lint-markdown` checked 125 files without fixes; `git diff --check` passes.
This final evidence/status update receives a closing Markdown check. The final
baseline comparison covers all 9,513 original tracked/nonignored entries. Exactly
16 baseline paths changed: the three specification/guide/tracker documents;
`sql.go`, `sql_capture.go`, `sql_test.go`; backend/projection/exact-set collectors;
three existing JavaScript fixture files; Links source-state queries; Network Flow
query construction, graph source reader and live SQL test bridge. Three private
files were added: `sql_return.go`, `sql_return_test.go`, `source-admission.mjs`.
All other baseline contents and file identities match. Ignored harness artifacts
are retained through the normal Make lifecycle rather than counted as source.

HEAD remains `beaf0293634e6aef9e62aa5670e033a26b6458d3`; the index digest remains
`3df1e41fe672cb561b7cd191b83cd37cff1090ea5fb4f15cd89f82367aac9247`, with nothing
staged. The existing 523 tracked changes and 41 untracked files were preserved;
changing the previously clean exact-set collector and adding three private files
brings the workspace to 524 tracked changes and 44 untracked files. The commit
recommendation concerns this reviewed pending change set with the explicit evidence
limits above; it does not authorize an automatic commit or publish operation.


### F-050 / S-051: recovered-panic return completeness, 2026-11-10

**P2, must_fix; NO-GO.** A named result initialized from unrestricted input can
be returned by deferred recovery before a later safe literal return. The existing
summary admits the defer, observes only the literal and both consumers accept it.
Implicit bounds panics and named deferred recovery functions reproduce the same
omission. This is an enforcement bypass; no production request exploit was shown.

The accepted contract rejects every defer in a raw helper's own body before value
traversal, including apparently harmless cleanup and unreachable/conditional
statements. This intentionally replaces the earlier cleanup-defer allowance within
raw helper inference. Nested functions retain independent contexts; ordinary SQL
cleanup, bound data, sealed values and independent body discovery remain supported.
The Harness NLSpec and enforcement guide now state this boundary. No Core ownership,
public API, policy grant, response schema or database migration is required.

Fresh baseline: HEAD `beaf0293634e6aef9e62aa5670e033a26b6458d3`, 524 tracked
changes, 44 untracked files, nothing staged. The 9,516-entry content/file-identity
snapshot is `/tmp/cartulary-s051-baseline-20261110.json`; index SHA-256 is
`3df1e41fe672cb561b7cd191b83cd37cff1090ea5fb4f15cd89f82367aac9247`.
S-051 → S-012 → S-014 is the remaining graph. Prior passing runs do not close
this counterexample. Production inspection found no raw SQL builder requiring a
cleanup-defer migration. Existing synthetic cleanup positives will become negatives;
separate ordinary-cleanup and nested-context controls retain the supported scope.

S-051 retained reproduction: `CARTULARY_TEST_RUN_ID=tcr-s051-red-20261110-01
CARTULARY_OUTPUT_MODE=ci make test-slice OWNER=harness.command_surface
ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3`
finished with outer exit 2 and normalized `product/test_assertion_failure`.
Both selected rows failed as expected: the Go execution record retained only
`SELECT * FROM artifacts` with `analyzed` disposition for explicit/implicit panic
and named recovery variants; the consumer test observed the same disposition.
Independent forbidden SQL discovery already worked, but the caller lacked the
required unresolved diagnostic. The ordinary-cleanup, bound-data, nested-context
and sealed-value controls passed. Secret scan passed; all 14 cleanup steps completed.
Source digest: `sha256:75bfa909257f37514db8b13a1b74e80e617ce34216283210036d7e570b2b7655`.

The repair now rejects `ast.DeferStmt` in the existing admission preflight and
uses the existing helper-named, source-located unresolved summary. It adds no
policy branch, wire field, recovery detector or deferred-call interpretation.
Qualification is pending; readiness remains NO-GO.


### Final S-012 reconciliation after S-051, 2026-11-10

F-050 and S-051 are DONE. The shared helper-summary preflight rejects every own-body
`ast.DeferStmt` before evaluating returns, with the helper name and defer location
in existing unresolved fields. The 16 negative variants run through five boundary
forms in both the Go analyzer and shared consumer fixtures. The negative assertion
checks the execution/composition record, disposition, source location and specific
reason, followed by Artifact projection and backend policy rejection. No policy
workaround, routing row, public field, capability or generated output was added.

| Requirement | Implementation | Fixture / validation evidence |
| --- | --- | --- |
| Recovered exits cannot qualify a raw helper | `sql_return.go` admission rejects defer before value traversal, including named/unnamed results and all return forms. | `runSQLHelperReturnTests` / `assertSQLDeferredReturn`; shared test "deferred raw helper exits reject at execution and composition in both consumers". Mechanics 01 covers explicit and bounds panics, called panic, inline/named/function-value recovery, and unrestricted or concrete unauthorized results. |
| Syntactic rejection has no cleanup/reachability exemption | Same shared admission rule across ordinary blocks, branches, switches and loops. | Conditional/multiple/recovery-free defers and a defer after an earlier return reject; former raw cleanup positives intentionally become negatives. Mechanics 01. |
| Rejection propagates without retaining only safe alternatives | Existing unresolved summary and expression/composition propagation. | Direct call, wrapper, split/renamed helper file, concatenation, formatting and sealed-literal composition. Mechanics 01 asserts the affected record and both policy diagnostics. |
| Nested scopes and ordinary cleanup remain supported | Nested function literals remain separate; existing binding effects and root-body discovery unchanged. | "defer admission preserves nested scopes ordinary cleanup bound data and sealed values" plus existing named/explicit/finite/zero-value, nondeferred mutation/address-escape and branch/assignment/loop controls. Mechanics 01. |
| Rejection does not suppress independent SQL discovery | Normal declaration and closure inspection remains independent of helper inference. | "rejected deferred helpers retain independent SQL discovery" verifies two distinct forbidden execution sites plus the caller rejection through both consumers. Mechanics 01. |
| Production scope and callers remain qualified | Existing consumers use the repaired analyzer without migration or exemptions. | Backend 01 and projections 01 pass; no runtime source changed. |
| Specification and guide remain non-executable | Harness SQL execution-state completeness and guide state the bounded language and migration pattern. | Human source review; Markdown lint. No executable input reads or depends on these documents. |

All S-051 roots below are under `.cartulary/test-results/`; each command used
`CARTULARY_OUTPUT_MODE=ci` and the explicit `CARTULARY_TEST_RUN_ID` shown. Native
outer exits were observed separately. Every root, including the failing
reproduction, has a passing retained-secret scan and only completed cleanup outcomes.

| Run ID | Exact command | Native exit / result |
| --- | --- | --- |
| `tcr-s051-red-20261110-01` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | 2 / expected assertion failures before the fix; diagnosis above. |
| `tcr-s051-format-20261110-01` | `make format` | 0 / pass; intermediate fixture source. |
| `tcr-s051-format-20261110-02` | `make format` | 0 / pass after making the unreachable-defer fixture end with a return. |
| `tcr-s051-mechanics-20261110-01` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.unit.go_import_inspection_distinguishes_declarations_8c659e57a3` | 0 / both selected rows pass, including previous return/effect/admission regressions. |
| `tcr-s051-boundary-20261110-01` | `make lint CHECK=backend-boundary` | 0 / production authored/generated/support SQL and package policy pass. |
| `tcr-s051-projections-20261110-01` | `make test-slice OWNER=module.projections ROWS=module.projections.architecture.sql_declaration_ownership` | 0 / production provider dependency and SQL closure pass. |

Final format 02, mechanics 01, boundary 01 and projections 01 share source digest
`sha256:c28f518ece11f139ce55d909f4d653ae6d07dd365d928751c52c2f3169549cbf`.
Intermediate format 01 has digest
`sha256:5e8425285030b589c7c29b6aaaa8cd4a5886e0ee4cf909e86a26565b605bd924`;
red 01 retains the pre-fix digest stated above. All six share source commit
`beaf0293634e6aef9e62aa5670e033a26b6458d3`, dirty source state and toolchain digest
`sha256:0e302b723a654ff49d1df64d0c655f27c266127f9f49480b0ea7a88b56be8893`.
Only documentation changes follow final qualification.

The 28 exact earlier roots enumerated in S-050's pending-change reconciliation,
plus `tcr-s050-links-20261110-01` and `tcr-s050-networkflow-20261110-01`, were
re-inspected: all 30 summaries and secret scans pass, with completed cleanup.
Their original native exits remain historical observations. These roots support
only unchanged runtime, generated, schema, routing and browser/HTTP obligations.
S-051 supplies fresh evidence for the changed analyzer and its consumers; the
S-050 raw-helper cleanup allowance is no longer an accepted control. Current
source admission and prior analyzer protections are also exercised by mechanics 01.

No production caller migration was required. Fixtures alone migrated away from
expecting raw helpers with cleanup defers to qualify. No runtime, policy owner,
registration, dependency or generated input changed, so broader product checks,
generation, schema/catalog reruns and agent-finalize are unnecessary. Full-owner,
whole-suite, visual, release and performance acceptance remain unclaimed. No
required targeted check is omitted.


### Final S-014 handoff after S-051, 2026-11-10

**Recommendation: scoped GO for committing the documented pending change set.**
F-050, S-051 and reopened S-012/S-014 are DONE. The preceding S-050 GO is historical;
this checkpoint includes the recovered-panic correction and supersedes the S-051
NO-GO checkpoints. There is no remaining identified P2 within the reviewed scope.
This is bounded enforcement qualification, not proof of universal Go/SQL soundness.

The accepted compatibility change is rejection of every defer in a raw SQL helper's
own body, including apparently harmless cleanup. Ordinary cleanup, bound data,
independently qualified sealed fragments and independent SQL discovery remain
supported. There is no database, deployment or external API migration.

Five new invocations pass; the sixth is the retained failing reproduction. All six
secret scans pass and all cleanup outcomes are completed. Markdown lint checked
125 files without fixes and `git diff --check` passed; the final tracker edit
receives the same closing checks. The final obligation/fixture/run matrix and the
30 compatible reused roots are explicitly scoped above.

The fresh baseline comparison covers all 9,516 original tracked/nonignored entries.
Exactly six existing paths changed: Harness specification, SQL enforcement guide,
this tracker, `sql_return.go`, `sql_return_test.go`, and the shared
`test-projection-sql.mjs` fixtures. No paths were added or removed; all other
contents and file identities match. HEAD and the index digest remain unchanged:
`beaf0293634e6aef9e62aa5670e033a26b6458d3` and
`3df1e41fe672cb561b7cd191b83cd37cff1090ea5fb4f15cd89f82367aac9247`.
There are still 524 tracked changes and 44 untracked files, nothing staged.
Ignored harness artifacts follow the Make lifecycle and are excluded from source
preservation counts. No staging, commit, merge or deployment was performed.
