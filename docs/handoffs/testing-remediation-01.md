# Testing Remediation 01: PostgreSQL Fixture Simplification

## 1. Scope and Source Posture

Document date: 2026-10-26. **Current action: closure implementation COMPLETE.**
The approved PostgreSQL Remediation Closure Plan is complete: **F-021–F-026,
WF-17–WF-19, S-017–S-019 and T-015–T-017 are DONE**. E-045, E-046 and
E-047 record the ordered checkpoints. Final acceptance uses 22 successful,
compatible audited roots, 42 distinct selected rows, and completed cleanup.
E-044 retains the entry baseline; E-037–E-043 remain historical evidence.
The bounded completion does not assert complete-owner, release or performance
acceptance. The separate preceding-handoff WF-33 remains outside this effort.

The earlier user instruction approved the PostgreSQL Fixture Remediation Plan and execution of WF-12–WF-16,
including scoped verification, the specification terminology clarification and
required tracker checkpoints. The selected scope is backend fixtures,
PostgreSQL first, with structural cleanup rather than test rearrangement alone.
**WF-12–WF-16 are DONE.** E-037–E-041 record the ordered checkpoints,
qualified assertion successors, final compatible evidence and completed handoff.
Completed WF-00–WF-11, their stable IDs,
decisions and evidence remain in the historical snapshot in Section 10; their
authorization and passes do not apply to this iteration. The separate open WF-33
in [the preceding handoff](testing-remediation.md) is unchanged.

Repository: `/home/jochi/code/cartulary`; established supplied and canonical target
`.`; kind: repository directory; label: `postgres-fixture-simplification`. This
continues the existing tracker's repository identity with a bounded subsystem
inspection. The earlier implementation write boundary was the seam described below, its
direct callers, tests, authored routing and generated projections, this tracker,
and the bounded fixture-model terminology clarification in the NLSpec. The
current authorized closure permits scoped fixture implementation, tests, direct
caller migration, authored routing and Make-generated projections, and tracker
updates. Adopted behavior is unchanged; no product API or data migration is required.
The output and parent components are not symlinks; the output is a regular,
singly linked file. Artifact mode: **normal for this bounded iteration**.
The [modular refactor framework](cartulary_modular_refactor_planning_framework.md)
and Refactor Tracker format supply planning structure, not behavioral authority.

The change boundary is `internal/testutil/pgtest`, its PostgreSQL composition in
`tools/testservices`, direct owned-cluster callers, and affected test routing.
Object-store internals, scheduler architecture, browser workflows, performance
fixture policy and production modules are outside this iteration. Their direct
PostgreSQL boundary assertions remain regression witnesses. Do not create a
generic service manager, public dependency-injection API, compatibility adapter,
new runner or new fixture-policy language to execute this plan.

| Authority | Applicable requirement and disposition |
| --- | --- |
| [Core 00](../spec/00_document_set_status_and_precedence.md), §§2–5 | Adopted owner allocation; tests and typed projections do not define requirements. Markdown remains excluded from executable tests, generators, runtime metadata, conformance and release evidence. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md), `adopted/current` | REQ-804: explicit attachment, physical ownership, scoped borrowing, recovery and quarantine. REQ-810: opaque migration capability and canonical source. REQ-811: provisioning identities, pristine prerequisites and clone grants. REQ-405 / §11.0: transaction settlement, cloning and cleanup bounds. REQ-410: declared dependencies before acquisition. Sections 3, 8 and 9: exact selection and truthful evidence. |
| [Core 01](../spec/01_architecture_storage_and_view_contracts.md), §2.1A / REQ-01-657 | Canonical migration lineage and ownership; targeted migration operations remain limited to harness-issued disposable scratch capabilities. |
| [Core 04](../spec/04_security_deployment_and_conformance.md), §12.3.1A / REQ-04-153 | Pinned PostgreSQL profile, checksums, verified certificate-only TLS 1.3 and purpose-specific roles/privileges. Pins remain authored in `tools/toolchain_pins.json`. |
| [Core 05](../spec/05_claim_publication_and_benchmark_reproducibility.md) | No timing, performance or release-publication claim is made by this plan or ordinary fixture checks. Preserve distinct qualification contexts without converting them into new claims. |

No inspected owner requires the mixed `Harness` shape, three owned-start wrappers,
mutable global test hooks, reflection method count, duplicate setup algorithms or
synthetic borrowed-service shutdown errors. No normative amendment is proposed to
preserve them. If implementation reveals a genuine primary-owner contradiction,
bound the affected slice and obtain the required owner decision before changing
that guarantee; do not accommodate the incidental test in the specification.

The following table records the earlier planning-only baseline. Authorized
execution revalidation is E-036; final execution preservation is E-041.

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Branch / commit | `main`; `867e97ed49829d671c58576232999d1fc9a4e3fb` | E-035 records final comparison | Historical frontend results qualify their own baseline only |
| Index / staged state | Clean; index-entry digest `2f8393a2bede4eadc41b7ccc1246f9daa19005bae7c33cc7c84b15d4b9155ce4`; raw index fingerprint retained | E-035 | No staging authorized |
| Worktree / untracked | Clean; no nonignored untracked files; 7,637 path identities/content fingerprints retained in E-030 | E-035 | Repository enumeration is not exhaustive behavioral inspection |
| Existing tracker | 80,994 bytes; inode 348826; SHA-256 `1898d970f68fd055fc5c1232d1ac297951694c756f5c98178e1b42575295cc47` | E-034 pre-write identity check; E-035 final result | Historical snapshot is retained with demoted heading levels |
| Ignored outputs / dependencies | Existing installed Node/pnpm and frontend readiness; existing workspace lock; no installations authorized | Docs lint uses external result/cache directories and installed-only admission | No exhaustive fingerprint audit of ignored dependency/build trees; invocation-path review bounds effects |

## 2. Current-State Repository Inventory

**Review overlay, E-042:** the earlier inventory below describes E-030/E-031,
not the final package shape. The completed source now separates physical
`OwnedCluster`, fixture `Harness`, attachment registry and template recipe;
private operation values replace the mutable function hooks. The review baseline
accounts for 7,649 tracked/nonignored entries, including 12 new pgtest source files.
Every entry is metadata/content-identity-only unless inspected below or in the
retained baseline inventory. No external links or nested repositories were traversed.

| Review boundary | Inspection depth / current conclusion |
| --- | --- |
| `pgtest/{owned_cluster,attachment,template,operations,initialization,provisioning,pgtest}.go` | Direct inspection of acquisition, preparation, admission, lifecycle events, migration issuance and cleanup; F-021–F-026 below. |
| `pgtest/{attachment,template,policy,migration,database,database_cleanup}_contract_test.go` | Direct assertion inspection; cancellation at connection-open, failed-readiness readmission, migration lease mismatch and pre-provisioning ownership are not characterized by the inspected cases. |
| `owned_cluster_test.go`, real template/transport tests, diagnostics | Relevant callback, failure, shutdown and qualification boundaries inspected; remaining diagnostic bodies are metadata/search-only. No exhaustive diagnostic audit claimed. |
| testservices composition and direct process/reset/forced-drop callers | Changed call sites and relevant startup/publication/cleanup bodies inspected; owner API migration and delegated template composition remain sound. |
| PostgreSQL `OpenSQL` / connection initializer; postgrescatalog create/clone/coordinator; suiteservices env/dependency/event functions | Direct dependency inspection confirms parent-context loss, creation timing and ambient event/dependency lookup. |
| Authored family selectors, generated render-index diff, compiler fixture environment | Direct affected-row/mapping inspection; no retired selector aliases or profile changes found. Compiler already exports `CARTULARY_FIXTURE_CAPABILITY`; no new policy language is needed. |
| Final retained evidence | Thirteen roots' manifests, summaries, cleanup and secret-scan records inspected read-only; 64 transport service scopes counted. All 7,338 frozen non-Markdown source/input hashes still match. No product tests or observability commands rerun. |
| Remaining repository, ignored outputs and unrelated dependencies | Metadata-only or excluded as in the earlier inventory. Retained evidence is inspection data, not a new validation run; ignored trees were not exhaustively fingerprinted. |

E-030 enumerates tracked and nonignored untracked entries under `.` without
following external symlinks or nested repositories. Every entry is metadata-only
unless the boundary inspection below says otherwise. Ignored dependency, cache,
build and retained-run trees are excluded from the source inventory; relevant
lint prerequisites were inspected separately. Generated projections are read-only
accounting, not source owners. All authored boundaries proposed for movement were
opened directly; diagnostics were inspected at their ownership and call sites,
not audited line by line. Remaining product callers are metadata/search-only.

| Path / kind | Disposition / inspection depth | Responsibility and known callers | Dependencies / tests / generated surface | Owner / risk / evidence |
| --- | --- | --- | --- | --- |
| `internal/testutil/pgtest/pgtest.go`, `pgtest_test.go` | Directly inspected lifecycle, attachment, preparation, migration, transaction, policy and cleanup bodies/assertions | Mixed physical owner and database fixture; ordinary and owned entrypoints; 116 Go importers found repository-wide | testcontainers, PostgreSQL ports, migrations, suite bindings/events, generated policy; tests currently construct mixed states and patch package globals | pgtest; high ownership/settlement risk; C-011–C-016; E-031 |
| `pgtest/{provisioning,initialization,tls}.go` | Directly inspected relevant provisioning, migration and security functions | Cluster initialization, pristine database prerequisites, clone finalization and TLS material | Purpose roles, crypto admission, canonical source and postgrescatalog | Retain narrow implementation owners; C-012/C-014/C-015; E-031 |
| `pgtest/{startup_diagnostic,endpoint_diagnostic,publication_diagnostic}.go`, `endpoint_diagnostic_test.go`, `publication_diagnostic_test.go` | Directly inspected startup/diagnostic boundaries and their callers; partial body depth | Startup diagnostics and safe qualification evidence | Existing redaction and endpoint observation assertions | Keep semantics; move only ownership receiver/call sites needed by S-014; C-011/C-015; E-031 |
| `pgtest/{template_preparation,publication_qualification}_test.go` | Directly inspected real lifecycle and transport qualification assertions | Template publication ordering and fresh physical startup lifetimes | Durable template state, checkpoint placement, clone admission, exact transport qualification profile | Retain distinct integration witnesses; C-012/C-015; E-031/E-033 |
| `tools/testservices/{main,main_test}.go` | Directly inspected PostgreSQL startup callbacks, template creation/publication, attachment, cleanup and related contract tests | Physical service composition, binding/lease publication; duplicate canonical template recipe | pgtest, suiteservices, postgrescatalog; existing lifecycle-contract catalog row | Keep composition here; move recipe to pgtest; C-011/C-012/C-016; E-031 |
| `internal/app/serverprocess/{process_fixture_harness_test,config_test}.go` | Directly inspected shared attachment and partial-startup helper | Process harness attaches PostgreSQL then object store; fake PostgreSQL shutdown error models an impossible attached-owner path | Process config row; ordinary service admission | Migrate detach call and replace synthetic shutdown assertion; C-011; E-031/E-033 |
| `internal/testutil/appsupport/{runtime,runtime_guard_test}.go` | Directly inspected fixture acquisition and declared-dependency guards | Reusable application fixture composition | pgtest and existing object-store composition; app.server support row | Keep behavior; no S3 redesign; C-011/C-016; E-031 |
| `internal/testutil/postgrescleanup/forced_drop_integration_test.go`; `internal/platform/harnessruntime/reset_characterization_test.go` | Directly inspected owned-start callers and lifecycle assertions | Self-owned cleanup barrier and reset peer cluster | Physical startup, prepared databases and owner close | Migrate directly to OwnedCluster; preserve integration contexts; C-011/C-013; E-031 |
| `internal/testutil/performancefixturelifecycle/postgres.go`; `internal/testutil/postgrescatalog/clone.go` | Directly inspected clone ownership, created/error result and finalization | Performance snapshot lifecycle and target-specific clone coordination | Shared low-level pgtest clone primitive, ownership ledger and grants | Keep boundary; performance template is a distinct snapshot workflow, not the ordinary duplicate recipe; C-012/C-013; E-031 |
| `internal/testutil/suiteservices` attachment, dependency and lifecycle boundaries | Directly inspected relevant env/event/diagnostic consumers; remaining bodies metadata-only | Admitted service bindings, scope and evidence | Generated catalog/policy and lifecycle result consumers | No protocol or scheduler redesign; C-011/C-016; E-031 |
| `tools/test_families/{harness.browser,platform.postgres,module.database_migrations,app.server,module.recovery}.json`; relevant `contracts/verification/owners` | Directly inspected affected selectors, capabilities and contexts | Authored routing to unit, integration and process evidence | Generated catalog, fixture policy and schedule projections | Migrate rows atomically after assertion mapping; C-016; E-033 |
| Root, internal and tools `AGENTS.md`; framework; owner clauses in Section 1; test-ops skill/recipes | Directly inspected applicable instructions and owner sections | Authority, boundaries and future public Make workflow | Documentation is planning input only | E-032/E-033 |
| Makefile; generated task surface; generated-artifact policy; execution topology; Markdown lint/readiness/step wrappers and config | Directly inspected relevant command, ownership and side-effect paths | Canonical validation and doc-only command admission | Authored inputs versus generated projections; installed-only readiness; external artifacts | E-033/E-035 |
| Remaining enumerated entries | Metadata-only, including unrelated authored code and generated/vendor/binary entries; links accounted as entries | No movement proposed | Textual import search does not prove absence of dynamic/external consumers | Next boundary is any newly discovered physical-owner or template-builder caller |

The core package inventory is 12 authored Go files, 5,579 lines, 28 top-level test
functions and seven catalog rows at this baseline. These are observations, not
target counts. Source fingerprints account for all 7,637 entries; the twelve core
files have the direct-inspection depths above. Supporting paths have only their
stated boundary inspection; every other entry retains metadata-only disposition.
The 116 importer search is caller discovery, not an audit of all their assertions.

## 3. Module Boundary Diagnosis

E-042 confirms that the owner/borrower split is useful and no additional package
hierarchy is warranted. The table below retains the original diagnosis; completed
moves are recorded in E-038–E-041. Remaining work belongs at the existing context,
creation-notification and attachment-admission seams, not in another manager or
generic dependency container. `attachmentRegistry.active` is redundant state
(F-026), distinct from the generation history required for quarantine.

At E-031 the package mixed physical service ownership, fixture provision and
mutation coordination. The completed split within the existing package addresses
that diagnosis; a new package hierarchy remains unjustified.

| Responsibility | Current location | Owner / disposition | Evidence / finding / contract |
| --- | --- | --- | --- |
| Container startup, readiness, TLS lifetime, partial acquisition and termination | `Harness`, three owned-start entrypoints, lifecycle callbacks | **Split** into `OwnedCluster`; keep physical composition in testservices | E-031; F-015; C-011/C-015 |
| Attached fixture access, allocations, per-database ownership and quarantine | Same `Harness` plus shared/attached flags | **Keep** in database `Harness`; attachment cannot acquire or terminate a physical owner | E-031; F-015/F-017; C-011/C-013 |
| Canonical template creation, migrations, handle closure and sealing | `ensureLocalTemplateDatabase` and testservices `createTemplateDatabase` | **Move/consolidate** into one pgtest preparation operation | E-031; F-016; C-012/C-014 |
| Clone admission, scratch migrations and rollback isolation | pgtest plus narrow postgrescatalog/postgrescleanup helpers | **Keep** purpose-specific operations and existing narrow SQL/cleanup owners | E-031; F-018/F-019; C-013/C-014 |
| Policy selection and operation dependencies | Duplicate resolver and mutable package-level functions | **Consolidate** one real resolver; **split** private instance dependencies by responsibility | E-031; F-017/F-018; C-013/C-016 |
| Assertion organization and catalog selectors | Large helper test file, repeated mode variants and mixed routing | **Consolidate** by guarantee and failure boundary; retain distinct real contexts | E-031/E-033; F-018/F-019/F-020; C-011–C-016 |
| Performance snapshots, S3, broker/scheduler and product behavior | Existing separate owners | **Defer** redesign; retain affected boundary regression witnesses | E-031/E-032; F-020; no independent change slice |

## 4. Public Contract and Behavior Freeze Map

Freeze required guarantees, not the test implementation. C-001–C-010 remain
historical. No HTTP, WebSocket, entity query/mutation, saved-view schema,
projection, revision, grid selector or user-facing authorization change is
proposed; these product contracts remain with their existing adopted owners.

| ID / surface | Observed behavior and owner-required guarantee | Disposition / assertions | Entry evidence / risk / slices / validation |
| --- | --- | --- | --- |
| C-011 — service ownership and attachment | REQ-804/410: ordinary paths require admitted explicit bindings; borrowed cancellation never destroys a shared owner; physical acquisition publishes cleanup authority before readiness; failed cleanup cannot make a generation reusable | Preserve fail-closed dependency admission, partial-acquisition cleanup and primary/secondary errors. Retire borrowed physical close and unreachable owned branch in `StopShared`; retain quarantine across repeated attachment | E-031/E-032; F-015/F-017; S-012/S-014; V-015/V-017/V-019; RB-005 |
| C-012 — provisioning and template publication | REQ-811 and §11.0: cluster init once per physical generation; prerequisites once per pristine database; canonical migrations/crypto admission before handle closure and template sealing; clone-specific grants before connectable publication; no replay on ordinary clones | Preserve real durability, checkpoint-placement, isolation and clone grants assertions. One recipe and one authoritative lifecycle event per fact; service composition still owns binding publication | E-031/E-032; F-016; S-013/S-014; V-015/V-016/V-019 |
| C-013 — isolation, ownership and settlement | REQ-804/405, §11.0: exact generation/provisioning/template/allocation scopes; transaction-only access; rollback and connection close each receive independent detached five-second budgets; invalidate before unlock on uncertain settlement; eager owned DB cleanup | Preserve quarantine/recovery authority and typed cleanup errors without replacing product failures. Preserve on-demand clones, single forced drop, target-specific advisory lock on one connection, 35-second total / 15-second admission / 2-second release limits; no pool, global quiescence or polling | E-031/E-032; F-017/F-018/F-020; S-014/S-015; V-015/V-017/V-018/V-019 |
| C-014 — migrations and capability | Core 01 REQ-01-657; REQ-810: unexported issued identity, canonical embedded source, apply target > 0 and rollback target >= 0 validated before DB access; database/error result and lifecycle evidence rather than a second preparation status record | Preserve rejection of arbitrary handles, source selection, targeted history and fresh-schema/crypto checks. Drop exact exported-method-count reflection assertion; future helper methods are not production behavior | E-031/E-032; F-016/F-019; S-013/S-015; V-015/V-016/V-018/V-019 |
| C-015 — PostgreSQL security and transport | Core 04 REQ-04-153 and adopted harness startup/qualification requirements: pinned engine, checksums, verified TLS 1.3, purpose roles/ACLs, bounded readiness and safe diagnostics | Preserve real role and transport witnesses, failure classification and fresh-lifetime qualification; mocks cannot replace these. Keep diagnostic semantics when changing receiver ownership | E-031/E-032/E-033; F-015/F-020; S-014/S-016; V-015/V-017/V-019 |
| C-016 — routing and evidence | Harness Sections 3/8/9 and REQ-405/410: catalog capability is the row request; exact selectors, declared dependencies, distinct execution contexts and truthful failure/evidence attribution | Preserve meaningful qualification contexts; remove redundant selectors/rows only with assertion successors. No retired-ID aliases, Markdown consumption or automatic owner-completeness claim from a subset | E-033; F-018/F-020; S-012/S-015/S-016; V-015/V-018/V-019 |

Test presence and inspected assertions establish planned witnesses, not a current
pass or specification completeness. Human review of projection fidelity remains
required. Existing deficiencies are reasons for simpler successors, not new
constraints to carry forward.

## 5. Coupling and Boundary Findings

| ID / classification | Evidence and problem | Action / benefit / compatibility disposition | Contract / slice / validation |
| --- | --- | --- | --- |
| F-015 / must_fix / DONE E-041 | E-031: `Harness` combines Container/Host/Port/TLS/readiness with databases, templates and transactions; shared/attached flags admit synthetic states; `StopShared` retains an owned branch although ordinary startup requires attachment | Separate physical ownership from fixture borrowing. One options-bearing `StartOwned`; direct caller migration; delete wrappers and borrowed no-op close instead of preserving them | C-011/C-015; S-014; V-015/V-017/V-019 |
| F-016 / must_fix / DONE, requalified E-047 | E-031: local and service templates independently create, provision, migrate, close and seal; `prepareDatabase` retains fresh-migration fallback after paths that require or prepare a template. Confirmed creation can be lost when later provisioning fails; an attempted name alone cannot grant deletion authority | One canonical template operation; explicit owned template preparation; ordinary fixture preparation clones an admitted sealed template or fails. Keep intentional raw scratch creation separate. Retain confirmed partial creation in the recovery ledger through provisioning/settlement failures, quarantine uncertain outcomes, and never delete a target whose CREATE was rejected. Delete duplicate recipe and implicit fallback | C-012/C-014; S-013; V-015/V-016 |
| F-017 / must_fix / DONE, requalified E-047 | E-031: package-global `*Fn` substitutions and sharedHarness resets couple otherwise independent tests and preserve invalid mixed states | Private dependency values and a small attachment registry instance for contract tests; production process reuse remains deliberate. Quarantine belongs to admitted generation state and survives borrower abandonment; detach state is removed. No public injection mode or universal mock framework | C-011/C-013; S-014; V-017 |
| F-018 / must_fix / DONE, requalified E-047 | E-031: policy resolver is only exercised by its own test while `ExplicitPostgresFixturePolicyT` duplicates logic; standalone/attached clone and migration cleanup repeat assertions | Route the real entrypoint through one resolver. Table-drive guarantee dimensions; preserve persistent borrower/owner-context cases where materially different. Remove duplicate test setup and top-level names after successor qualification | C-013/C-016; S-015; V-018 |
| F-019 / must_fix / DONE, requalified E-047 | E-031: migration capability test freezes exactly three methods; smoke startup test checks only preparation and nonempty DSN; process fixture helper injects borrowed PostgreSQL cleanup failure | Test issuance/rejection and canonical migration semantics; merge smoke obligation into substantive migrated-schema/crypto test. Replace impossible borrowed shutdown error with detach/nontermination assertion; physical cleanup failures stay in owner tests | C-011/C-014; S-014/S-015; V-017/V-018 |
| F-020 / intentional retention and clarification / DONE E-041 | E-031/E-033: real template durability, transaction isolation, drop/barrier, transport qualification, purpose ACLs, reset and process workflows cross different boundaries | Retain independent witnesses even where setup looks similar. Performance snapshot sealing has different ownership semantics; defer its redesign and S3/scheduler work. Catalog changes preserve execution meaning rather than test counts. The bounded specification wording now describes quarantine/readmission directly, without implying a writable pool | C-011–C-016; S-012/S-016; V-015/V-019 |

`must_fix` is limited to the named slice acceptance conditions. There is no
identified production behavior correction or data migration in this plan. The
desired improvement is fewer owners, states, recipes and duplicated assertions;
line/test counts are diagnostics, never a deletion quota or proof of coverage.

### Authorized remediation decisions

| Gap | Areas | Durable remediation and rationale | Compatibility / unresolved risk | Completion evidence |
| --- | --- | --- | --- | --- |
| F-015 | Implementation, tests, caller documentation | Separate physical `OwnedCluster` from borrowed `Harness`; one options-bearing constructor makes cleanup authority explicit and eliminates mixed states | Internal callers migrate atomically without aliases; leaving the mixed type permits invalid lifecycle operations | V-017: owner-only shutdown, recoverable partial acquisition, idempotent cleanup and preserved diagnostics |
| F-016 | Implementation, tests, composition | One template recipe owns provisioning, canonical migration, crypto admission, close, seal and event facts. Confirmed creation remains owned after later errors; never infer ownership from an attempted name | Explicit preparation replaces ordinary fallback; no product/data migration. Unresolved duplication risks premature publication and lost cleanup authority | V-016: fault matrix, primary/secondary errors, real durable sealing and grants; no replay or usable publication after failure |
| F-017 | Implementation, tests | Private instance dependencies and generation-scoped attachment state isolate fault injection without a public injection framework | No borrower detach API remains; unresolved global hooks permit order dependence and false readmission | V-017: independent registries/dependencies, exact binding validation, poison survives repeated attachment, unrelated generations remain admissible |
| F-018 | Implementation, tests, authored routing and projections | One real policy resolver and guarantee-shaped assertion matrices reduce duplicated logic while preserving material ownership differences | Retired selectors are removed without aliases; unresolved duplication certifies unused logic or obsolete witnesses | V-018: real helper policy selection, fail-closed capability admission and qualified assertion successors |
| F-019 | Tests, narrow fixture composition, routing | Replace method-count and smoke-only constraints with issuance, bounds, source and migrated-schema assertions; replace impossible borrower shutdown with borrower nontermination | Internal test identities may retire; object-store startup remains the primary error. Unresolved tests impose needless API compatibility | V-017/V-018: forged/zero rejection before access, canonical operations, crypto format and borrower nontermination |
| F-020 | Specification clarification, tests, routing, handoff | Preserve distinct unit/database/process contexts; replace misleading pool wording with quarantine/reuse terminology | No pooling, scheduler, S3 or performance-policy change. Unresolved ambiguity and cosmetic deduplication invite unsupported designs or coverage loss | V-019: real contexts remain selected, including 64 fresh transport lifetimes and unchanged four-way profile; human semantic review remains separate |

### Post-implementation review findings — E-042

The completion assessments for F-016/F-017/F-018/F-019 are reopened by the
specific gaps below. Their implemented structural changes and E-037–E-041 test
results remain valid historical evidence. Findings are source-backed deductions;
no new reproduction or product check was run in this review. P1 means prompt
correction; P2 means required follow-up; P3 is optional simplification.

| ID / priority / disposition | Concrete trigger, evidence and consequence | Owner / original obligation | Remediation and affected areas | Validation / follow-up |
| --- | --- | --- | --- | --- |
| F-021 / P1 / must_fix / DONE, E-045 | `template.go:72` opens through a context-free dependency; `pgtest.go:859` uses `context.Background()` for `postgres.OpenSQL`, whose `PingContext` and role/admission queries perform I/O. `initialization.go:18` opens the second admission handle the same way. A stalled connection can exceed the two-minute preparation bound and caller cancellation, holding `templateMu` and delaying disposal. The mutex wait at `template.go:39` also ignores a waiting caller's shorter deadline. | C-012/C-015; F-016; REQ-804 bounded provisioning/cancellation; existing TemplatePreparationTimeout | Implementation/tests: pass the preparation context through both opens and production admission; make waiting for the one preparation attempt cancellation-aware. Keep one owner/recipe and the existing independent disposal budget. No requirement amendment. | V-021, WF-17: stall each open/admission independently, cancel/expire the caller and a concurrent waiter; no publication, preserved initiating/close/drop outcomes, prompt bounded settlement. |
| F-022 / P2 / must_fix / DONE, E-046 | In `attachment.go:83–91`, generation inspection can confirm the advertised identity, then readiness can fail at line 87 before any generation record exists. A subsequent attach (with or without detach) repeats readiness and admits that same generation if the probe now succeeds. The current test poisons only an already admitted handle. | C-011/C-013; F-017; REQ-804 explicitly retires generations on failed/canceled/uncertain readiness and forbids later success restoring eligibility | Implementation/tests: retain a failed admission record for the verified generation and exact binding before returning the readiness failure. Reattach must fail closed without another probe or preparation; preserve the primary error and leave independent generations admissible. Do not infer physical ownership from an unverified name. | V-022, WF-18: matching generation then failed/canceled readiness; repeated attach and detach/attach reject without new I/O; another verified generation works. |
| F-023 / P2 / must_fix / DONE, E-046 | `pgtest.go:293–339` issues fresh/targeted migration capabilities without checking the selected broker capability. Creation checks only the service declaration. The compiler exports `CARTULARY_FIXTURE_CAPABILITY` but pgtest never reads it. A transaction or dedicated row declaring postgres can therefore request migration scratch and obtain targeted migration authority; on a shared transaction generation its create/drop also violates the declared isolation. The consolidated policy test rejects only clone misuse. | C-013/C-014/C-016; F-018/F-019; REQ-405 helper capability admission and REQ-810 issuance for a disposable postgres_migration lease | Implementation/tests, caller/routing review: validate the existing admitted capability before scratch creation/issuance, including both full-chain and through-version factories. Preserve explicit physical-owner lifecycle paths and the compiler's intentional allowance for current-head clones within migration rows. No new fixture-policy vocabulary or compatibility wrapper. | V-022, WF-18: absent, transaction and dedicated capabilities reject migration factories before create/open/source; admitted migration succeeds. Audit real scratch callers and retain canonical-source/target-bound tests. |
| F-024 / P2 / must_fix / DONE, E-045 | `pgtest.go:680–696` records exact ownership and the created event only after the low-level operation returns. That operation has already run `PrepareFreshDatabase` (`pgtest.go:843`), or clone ACL finalization under `postgrescatalog/clone.go:28–32`. A provisioning stall or process loss occurs after confirmed CREATE but before the per-database recovery ledger sees it. Returned errors do retain the bool; whole-cluster recovery authority also remains. Those improvements do not satisfy the plan's explicit pre-risk registration ordering. | C-012/C-013; F-016; Section 7 requires registering partial database ownership before risky work; REQ-804/811 preserve exact ownership through failed provisioning | Implementation/tests: report confirmed CREATE to the existing owner ledger immediately, before prerequisite/grant work; retain target-scoped catalog admission and the created/error settlement result. A failed CREATE must never invoke this ownership notification. A narrow callback/creation boundary is sufficient. | V-021, WF-17: inside provisioning/finalization, assert exact local and journal ownership already exists; stop/fail there and retain recovery proof. Rejected existing target remains unowned and undropped; publication/cleanup errors stay joined. |
| F-025 / P2 / must_fix / DONE, E-046 | `attachment.go:90–92` constructs a registry-local fixture without an event sink or retained environment. `Harness.recordEvent` falls back to `RecordEvent(nil, ...)` at `pgtest.go:906`, and create admission uses `CheckServiceDependencies(nil, ...)` at line 673. Independent registries therefore still write to the ambient suite journal and react to ambient dependency declarations. A local postgres binding can attach successfully then fail preparation because an unrelated process-global declaration differs. | C-011/C-016; F-017's independent registry/dependency instances; REQ-804 exact admitted binding and evidence ownership | Implementation/tests: bind event publication and dependency admission to the registry's admitted environment/private operation values, with explicit production process-environment composition. Copy the relevant binding inputs if needed; no generic injection framework. | V-022, WF-18: two independent registries with different suite/result/dependency bindings cannot write each other's journals or inherit admission/publication failures, including when ambient env differs. |
| F-026 / P3 / simplify / DONE, E-046 | `attachmentRegistry.active` is assigned, returned once and cleared; every attach lookup uses `generations`, and the retained generation entry still holds the same handle. Clearing active has no effect on admission, lifetime or retained references. Tests assert the implementation field rather than an additional guarantee. | F-017/F-019; requested conceptual clarity and minimal state | Selected closure: remove active state, DetachShared and synthetic detach plumbing. Borrowers release references; retain generation quarantine/recovery records and behavioral nontermination tests. No no-op compatibility API or second lease abstraction. | Selected removal is complete: borrower nontermination, primary-error preservation and retained quarantine qualify at E-046; no detach API remains. |

All required corrections affect internal test support; none requires product API,
schema or persisted-data migration. Context propagation may intentionally change
internal helper signatures and requires atomic caller migration without aliases.
Capability admission will intentionally reject previously tolerated misdeclared
scratch requests. The expected long-term benefits are bounded failure handling,
unambiguous recovery proof, capability-correct isolation and genuinely independent
contract tests. Leaving F-021–F-025 unresolved retains timeout hangs, poisoned
readmission, underdeclared cluster mutation, delayed ownership evidence and test
coupling respectively. F-026 was optional in E-042; the closure request selected its removal as a required gate, now qualified at E-046/E-047.

The single recipe, owner/borrower separation, canonical migration source, nonzero
issuer token, meaningful smoke successor, process borrower nontermination, retained real
security/transport contexts and bounded specification terminology change are
supported by inspection. No regression in their inspected success paths or
retired-selector routing was found. This is not an exhaustive security or
whole-repository conformance review.

## 6. Refactor Workstreams

This is the authoritative dependency list. Each checkpoint records source
identity, assertion dispositions, exact commands/results/artifacts, unresolved
failures and the next gate before a successor starts. The original five workflows
below are completed history; the authorized closure chain is completed at E-045–E-047.
Neither chain bypasses its predecessor or silently consumes historical passes.

| ID / slice | Class / prerequisites | Goal / likely files | Validation / checkpoint / binary exit |
| --- | --- | --- | --- |
| WF-12 / S-012 — DONE, E-037 | root; authorized; baseline revalidation | Clarify specification terminology, freeze essential guarantee-to-assertion map and run focused existing witnesses | V-015; every planned retirement has a successor or an incidental-behavior rationale; baseline failures classified before movement |
| WF-13 / S-013 — DONE, E-038 | chain; WF-12 | Consolidate ordinary template preparation in pgtest; testservices delegates | V-016; one canonical create/provision/migrate/close/seal path with qualified failure disposal and publication order |
| WF-14 / S-014 — DONE, E-039 | chain; WF-13 | Separate OwnedCluster and fixture Harness; migrate physical callers, localize hooks and detach semantics | V-017; borrowed handles cannot terminate owners; all owned callers use one lifecycle API; quarantine and primary errors preserved |
| WF-15 / S-015 — DONE, E-040 | chain; WF-14 | Consolidate tests and policy resolution; update authored selectors/owners; regenerate projections | V-018; assertion successors cover meaningful guarantees and no dead test hook, duplicate resolver, obsolete selector or compatibility shim remains |
| WF-16 / S-016 — DONE, E-041 | chain; WF-15 | Finalize, qualify compatible final portfolio and save handoff | V-019; final source identity and artifacts support the bounded completion criteria in Section 12 |

### Authorized closure chain — E-044

**WF-17 → WF-18 → WF-19** is the authorized bounded correction chain. Preserve
completed WF-12–WF-16 and failed evidence; checkpoint this tracker after each
future exit and before its successor. No rollback of the useful structural split
is recommended.

| Workstream / slice / status | Entry gate and readiness | Work / dependencies / principal risk | Exit criteria |
| --- | --- | --- | --- |
| WF-17 / S-017 / DONE, E-045 | Authorized by closure request; source/owner revalidation and fault characterization precede correction | Correct preparation contexts and confirmed-create notification; keep catalog locking and physical recovery authority. Risk: notification before confirmed ownership or disposal before lock release. | V-021, focused A/B and real template E witness; fault witnesses qualified; close/seal/publication order and cleanup bounds retained; tracker checkpoint. |
| WF-18 / S-018 / DONE, E-046 | WF-17 exit satisfied; F-022/F-023/F-025 characterization qualified | Retain failed verified-generation admission, enforce scratch lease capability, bind local dependencies; selected F-026 removal. Risk: rejecting valid lifecycle callers or clearing poison while simplifying. | V-022, A/B/D and real template/transaction E witnesses; caller/routing changes atomic; generated/static checks if inputs change; tracker checkpoint. |
| WF-19 / S-019 / DONE, E-047 | WF-18 exit satisfied; finalization and frozen final portfolio qualified at E-047 | Finalize, inspect generated diffs, freeze source, run compatible A–E successor/static portfolio and audit retained roots. Risk: reusing E-041 as proof of untested correction paths. | V-023; all required findings closed with exact roots/outer exits/source identities; scoped limitations and handoff saved. RESULTS_DIR only for compatible successful full warm check. |

Rollback stays slice-sized: correction, callers/tests and authored routing
reverted together, generated outputs rebuilt through Make; preserve prior edits
and failure evidence. No specification amendment is needed for these fixes.

## 7. Authorized Refactor Slice Plan

S-012–S-016 were authorized by the earlier implementation instruction.
The current closure request authorizes S-017–S-019 and their scoped validation. Readiness records evidence gates. RB-005 and RB-006 are closed by the ordered
E-037–E-041 checkpoints; each predecessor finished before its successor began.

| ID / kind | Findings / contracts | Entry conditions / readiness | Intended change / completion | Validation / rollback |
| --- | --- | --- | --- | --- |
| S-012 / characterization and disposition | F-015–F-020; C-011–C-016 | **DONE**, E-037; current characterization qualified movement | Complete assertion ledger and record current focused baseline; distinguish fixture failure from product failure before proceeding | V-015; save baseline checkpoint; no data migration |
| S-013 / structural preservation | F-016; C-012/C-014 | **DONE**, E-038; WF-12 exit satisfied | One ordinary template builder owns operation order, cleanup and event facts; testservices supplies identity and publishes ready bindings afterward; ordinary clones have no fresh-migration fallback | V-016; revert source/tests/routing together if guarantee gaps arise; no legacy builder wrapper |
| S-014 / structural preservation; private support API changes | F-015/F-017/F-019; C-011/C-013/C-015 | **DONE**, E-039; WF-13 exit satisfied | OwnedCluster owns physical lifecycle; Harness owns fixture capabilities; single owned-start API, explicit template setup, instance-local dependencies and safe detach; direct consumers migrate atomically | V-017; revert the slice and caller changes together; do not retain old APIs as aliases |
| S-015 / assertion and routing consolidation | F-018/F-019/F-020; C-013/C-014/C-016 | **DONE**, E-040; successors qualified before retirement | One policy resolver, coherent test files, consolidated cleanup cases, meaningful capability tests, migrated-schema witness and exact authored/generated routing | V-018; restore source and authored projections together, regenerate via Make; no retired-ID compatibility map |
| S-016 / final qualification and handoff | F-015–F-020; C-011–C-016 | **DONE**, E-041; final portfolio and handoff complete | Finalize before broader validation; confirm compatible portfolio, no unintended production changes and scoped completion | V-019; keep incomplete work explicitly incomplete; no production data migration |
| S-017 / preparation recovery | F-021/F-024; C-012/C-013/C-015 | **DONE**, E-045 | Context-aware opens, one terminal attempt and pre-provision ownership notification | V-021; qualified fault/real lifecycle witnesses |
| S-018 / bound admission and routing | F-022/F-023/F-025/F-026; C-011/C-013/C-014/C-016 | **DONE**, E-046 | Durable verified admission, immutable inputs, migration capability and detach removal; atomic row splits | V-022; complete caller inventory, generation and static checks |
| S-019 / frozen final validation and handoff | All closure findings and retained guarantees | **DONE**, E-047 | Finalize, freeze, execute compatible portfolio, audit evidence and save handoff | V-023; all required closure work complete |

### Intended final ownership and API

`OwnedCluster` holds the container, endpoint/TLS lifetime, readiness observations,
physical generation and terminal cleanup. `StartOwned(ctx, StartOptions)` is the
single owned constructor. Labels and observers remain useful options; acquisition
and settlement callbacks receive the physical owner. Migrate
`StartOwnedWithLabels` and `StartOwnedWithOptions` callers directly, then remove
those variants. Move physical `Close`, `ContainerID`, readiness and related
diagnostic ownership here; preserve bounded startup and partial-resource recovery.

`Harness` holds admitted fixture binding, template identity, database ownership,
transaction registry and quarantine. Keep useful ordinary `Start(t)` and
`StartShared(ctx)` entrypoints, which only attach. It exposes no physical close.
Remove `DetachShared` and redundant active attachment state; borrowers release
references normally. Retained generation records prevent readmission after
verified admission failure or quarantine. Preserve
the process-level registry where reuse is useful, but make unit tests use their
own private registry/dependency instances rather than resetting package globals.

Owned callers explicitly prepare a template before requesting migrated clones;
raw pristine and migration scratch operations stay explicit. An owned cluster
provides access to its fixture handle after physical initialization and an
explicit template-preparation operation returning the prepared database and
error. Keep the common builder in pgtest, with one implementation for local
owned tests and suite-service startup. The service layer supplies suite/template
identity and retains lease and binding publication. Do not introduce a parallel
preparation-status record or duplicate lifecycle events. Register partial
database ownership before risky work; close migration handles, seal, then publish
the template; retain cleanup failures alongside the initiating error and
quarantine uncertainty. The final API must make omitted template preparation a
clear error instead of silently migrating a fresh ordinary fixture.

Use small private dependencies for attachment, physical startup and database
preparation. Do not build one extensible dependency container for unrelated
operations. Keep narrow SQL/provisioning helpers where they already have an
owner. Move code only to support this responsibility split; unrelated diagnostic
or performance-policy redesign is deferred.

### Assertion successor and retirement ledger

| Existing assertion group | Final disposition / meaningful successor | Retirement gate |
| --- | --- | --- |
| `TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness`, explicit-attachment and omitted-dependency cases | Attachment tests prove missing/malformed binding fails before acquisition; borrower abandonment does not terminate owner; quarantined generation cannot be reused | C-011; V-017; remove invalid mixed Harness construction and unreachable owned-shutdown branch |
| Owned labels, partial startup, readiness/cancellation/authentication failures and diagnostic helpers | OwnedCluster tests use local startup dependencies; retain callback ordering, bounded errors, exact cleanup authority and real transport qualification | C-011/C-015; V-017/V-019; remove old start wrappers only after direct caller migration |
| Duplicate local/service template sequence tests | One operation-order/fault matrix plus thin service publication-order assertion; real `TestTemplatePreparationCompletesCopyBeforePublication` remains | C-012; V-016; no duplicate migration or premature published binding |
| Standalone/attached isolated-clone cleanup and standalone/attached migration-scratch cleanup | **Qualified/replaced, E-040:** `TestDatabaseFixtureCleanup` covers clone/head scratch/targeted scratch × standalone/attached/persistent borrower, exact owned name, one eager drop and handle-before-drop order; independent context/failure authority witnesses remain | C-013/C-014; V-018; keep differences that affect actual cleanup authority |
| Policy resolver-only test plus `ExplicitPostgresFixturePolicyT` | **Qualified, E-040:** public helper uses the sole resolver; its actual attribution path is tested for top-level/package/default selection and rejected policy before acquisition | C-016; V-018; delete unused alternative resolver |
| `TestMigrationDatabaseTargetedOperationValidation` | **Qualified, E-040:** nil/zero/forged rejection; invalid targets before counted source loading or execution; legal canonical-source apply/rollback; no field/method inventory | C-014; V-018; changing private method layout must not become compatibility work |
| `TestHarnessStartsPostgresAndRunsCurrentMigrationPath` and `TestPrepareIsolatedDatabaseTReturnsMigratedDatabase` | **Qualified/replaced, E-040:** `TestPrepareIsolatedDatabaseTReturnsMigratedDatabase` retains schema and `application_crypto_format` queries; smoke function/selector removed | C-012/C-014; V-018; no weaker nonempty-DSN replacement |
| `requireProcessHarnessStartupErrorContract` | Preserve object-store attach failure as primary and perform no PostgreSQL owner operation; physical cleanup-error joins remain in OwnedCluster tests | C-011; V-017; remove synthetic borrowed cleanup error without redesigning S3 |
| Rollback isolation/uncertain settlement, real forced drop/barrier, reset, ACL, process readiness/recovery and performance snapshot lifecycle | Retain separate real contexts and failure boundaries; share setup only when ownership semantics match | C-011–C-016; V-019; no deletion based solely on similar test titles |

Organize focused unit assertions by attachment/policy, owned-cluster lifecycle,
database preparation/cleanup, and migration/transaction capabilities. Keep real
PostgreSQL lifecycle and transport qualification separate from those unit
fixtures. Filenames and top-level test names may change with their routing;
neither their present layout nor their population is an acceptance requirement.

### Closure assertion successors — F-021–F-026

| Gap / retained guarantee | Qualified successor and exact selection | Disposition |
| --- | --- | --- |
| F-021 cancellation, deadline, one terminal preparation, independent waiters and settlement | `TestPrepareDatabaseTemplateModeClonesWithoutMigrationReplay/template_preparation_contract` in C's certificate_fixture_lifecycle; real template witness in E | Added work/guard stalls, real TCP open, cancellation and waiting deadline, one result, no retry and no premature publication; V-021/V-022 and final C/E requalification pass, E-047. |
| F-024 confirmed CREATE before dependent work; rejected targets never owned | Same template contract plus `TestTemplatePreparationCompletesCopyBeforePublication` in E | Added pre-provision local/journal checks for pristine/clone callbacks; retained primary/close/drop/release and unowned-target boundaries. |
| F-022 terminal verified admission and independent generations | `TestStartSharedRequiresExplicitAttachment` / `TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness` in A | Added readiness failure/cancellation/publication, no repeated probes, pending caller cancellation and concurrent independent generation. |
| F-023 declared migration issuance and canonical target bounds | `TestValidateSelectedPostgresFixturePolicy/migration_issuance_requires_capability` in A; `TestDatabaseFixtureCleanup` and `TestMigrationDatabaseTargetedOperationValidation` in B; recorded caller inventory | Capability/scope cross-product rejects both factories before acquisition; valid issuance, source identity and forged/zero/invalid targets remain. No new policy language or bypass. |
| F-025 stable authority, policy and evidence destination | A's attachment binding/authority/journal cases; `TestMigrationActivityPrivatePublicationAndShutdown` in timing row | Conflicting ambient values do not redirect admission, guards, policy, creation journal or timing publication. Changed cached bindings fail; one journal fault does not poison another registry. |
| F-026 borrower nontermination and initiating error | A's borrowed-handle capability/quarantine assertions; E's real owner usability; D's `requireProcessHarnessStartupErrorContract` | Remove detach field/callback assertions and API together. Existing top-level selectors now protect behavior; no retired selector or forwarding stub is needed. |

The three new row IDs are exactly those approved in the closure plan. Graph
current-head schema moves to its dedicated row; six canonical cutover tests move
together to the reference-data migration row; reporting identity cutover moves
to its migration row. Remaining ordinary tests stay dedicated. The final
portfolio includes both sides of every split. No meaningful assertion was deleted
without its qualified replacement; existing transport/reset/ACL/recovery and
performance snapshot contexts remain separate witnesses.

## 8. Validation Plan

V-020 records the earlier documentation session. V-015–V-019 are now authorized
implementation checks; only the current execution checkpoints establish which
have actually run. Run commands at the repository root
using the pinned installed tools and Make-managed service admission. No direct
`go test`, script execution or manually provisioned database is a canonical
substitute. Consult `make help ALL=1` and `make task-guide OWNER=<owner>` before
the future run; exact row selection below is the inspected starting portfolio.

| ID / layer | Stage / contracts / slices | Exact command or selection | Environment / expected evidence / failure gate |
| --- | --- | --- | --- |
| V-015 / characterization | Pre-move; C-011–C-016; S-012 | Focused selections A–D below; inspect `make task-guide OWNER=harness.browser`, `OWNER=platform.postgres`, `OWNER=module.database_migrations`, `OWNER=app.server` | Pinned Go/Node, Make-managed PostgreSQL and process dependencies; retain exact run roots, source IDs and row results. Unexplained baseline failures block the affected move, not a reason to broaden/rerun blindly |
| V-016 / preparation | Post S-013; C-012/C-014 | A and B; C for fresh admission/roles; selected lifecycle integration rows in E | Require fault coverage and real sealed-template/clone witness. Publication or disposal failure blocks ownership split |
| V-017 / ownership | Post S-014; C-011/C-013/C-015 | A, B and D; E for physical startup/cleanup/reset; transport qualification row remains explicit | Container runtime and exact row scopes; no borrower termination, lost cleanup authority, masked primary failure or poison reuse. Preserve qualification profile |
| V-018 / consolidation/routing | Post S-015; C-013/C-014/C-016 | Qualified successors of A–D; `make generate`; `make source-check CHECK=generated`; `make source-check CHECK=schema`; `make source-check CHECK=catalog`; `make source-check CHECK=identity` | Authored owner changes precede generation; update this ledger with exact successor IDs. No dangling selector, extra execution, weakened fixture capability or Markdown input |
| V-019 / final regression | Final; C-011–C-016; S-016 | `make agent-finalize`, then compatible final union of A–E and affected owner-guided static/import checks; `make harness-observability-check RESULTS_DIR=<exact-run-root> CHECK=all` for applicable retained runs | Freeze source after finalization. Passes must share compatible source/tool/config identity. Observability check does not make a subset an owner-complete run. Failed obligations keep implementation incomplete |
| V-020 / documentation | Planning-session check; T-008 | `make lint-markdown` with installed-only admission and external artifacts as recorded in E-035; `git --no-optional-locks diff --check`; manual structure/link/trace review | No tests, service changes, generators or dependency installation. Configured lint globs omit this handoff; report that limit and review the changed document directly |

### Review follow-up validation obligations

| ID | Required proof / selection | Current result / gate |
| --- | --- | --- |
| V-021 | F-021/F-024 context and creation-notification witnesses; A/B plus `platform.postgres.support_unit.certificate_fixture_lifecycle` and `harness.browser.integration.postgres_template_preparation` from E | PASS, E-045. Added the certificate/lifecycle row because it owns the template fault contract; A alone does not select it. Historical evidence is retained separately. |
| V-022 | F-022/F-023/F-025 admission/identity/isolation witnesses; A/B/D and template/transaction E rows; task guides and affected routing/source checks if changed | PASS, E-046: A/B/D, template/transaction, all 19 caller/split rows, timing evidence and generated/schema/catalog/identity plus Go/boundary checks. Final frozen-source requalification remains V-023. |
| V-023 | WF-19: `make agent-finalize` before broader compatible A–E successor portfolio, applicable Go/boundary/source checks and `make harness-observability-check RESULTS_DIR=<exact-run-root> CHECK=all` | PASS, E-047: finalized/frozen source, exact final portfolio, 22 compatible successful graph roots with outer exits 0, all retained audits pass, final preservation and handoff complete; bounded claims only. |
| V-024 | Review-only tracker save: external-output installed-only `make lint-markdown`, direct tracker review and optional-write-disabled diff/preservation checks | E-043; lint globs exclude this tracker, so direct review is required. Product tests, generation and finalization are outside this review. |

### Exact starting selections

A — helper and composition contracts:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.unit.testservices_lifecycle_contract
```

B — scratch capability and migration admission:

```sh
make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.test_harness_targeted_operation_validation,module.database_migrations.integration.cryptographic_state_admission,module.database_migrations.integration.schema_bootstrap
```

C — certificate fixtures, migrated admission and purpose privileges:

```sh
make test-slice OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle,platform.postgres.integration.fresh_fixture_admission,platform.postgres.integration.purpose_identity_roles_and_acl
```

D — reusable app admission and real process attachment/configuration:

```sh
make test-slice OWNER=app.server ROWS=app.server.support_unit.incident_create_commit_fault_requires_validated_efdc344831,app.server.process.effective_deployment_configuration_fails_the_rea_c17cbafdfc,app.server.process.a_fresh_process_reaches_health_and_ready_only_af_c93818219a
```

E — distinct real lifecycle and recovery witnesses, selected by exact row rather
than an automatic full-suite run:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.integration.postgres_template_preparation,harness.browser.integration.postgres_transaction_isolation,harness.browser.integration.postgres_owned_barrier_observation,harness.browser.integration.postgres_single_drop,harness.browser.integration.postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277,harness.browser.integration.reset_checkpoint_characterization,harness.browser.integration.managed_service_assertions_for_runtime_reset_rec_79940f98ab,harness.browser.integration.performance_fixture_snapshot_lifecycle
make test-slice OWNER=harness.browser ROWS=harness.browser.integration.published_postgre_sql_transport_qualification_ac_38acf197c9
make test-slice OWNER=module.recovery ROWS=module.recovery.process.the_packaged_server_serves_the_coherently_restor_6b1731590e
```

The explicit transport qualification retains its existing 64 fresh lifetimes and
four-way qualification profile; startup receiver movement is not grounds to
replace it with a unit mock. Template preparation, transaction borrowing,
physical-owner barriers and persistent borrower cleanup have different scopes.
Retain each applicable context. Existing performance snapshot lifecycle coverage
is regression evidence for shared clone primitives, not permission to redesign
performance policy or assert a speed improvement.

After S-015, retired selectors in these baseline selections must be replaced by
the recorded qualified successors; do not keep obsolete IDs alive to make old
commands pass. For finalization, supply `RESULTS_DIR` only if a compatible
successful full warm-check root actually exists; otherwise state that retained-run
maintenance was skipped because it was unset. Do not manufacture a full check
solely for maintenance. Broaden validation only for uncovered changed ownership
or a diagnosed failure; do not automatically stack full check, CI and every test.
Browser/visual/golden and timed benchmark runs are not required by this bounded
backend support change. Production static/import obligations discovered through
the affected owner guides still apply during implementation.

| Validation | Invocation-path review / side effects | Current admission / result |
| --- | --- | --- |
| V-015–V-019 | Tests acquire services and produce evidence; generation/finalization may mutate outputs; command surfaces and authored rows inspected in E-033 | V-015–V-019 executed at E-037–E-041; successful scoped evidence retained, but E-042 reopens the uncovered obligations V-021–V-023 |
| V-020 | Make admission uses existing lock; installed-only readiness checks existing runtime/dependencies; Markdown runner has fixing disabled; static cache/results redirected outside repository; current globs exclude this target | E-035 records actual command, result, manual review and final preservation comparison |

## 9. Top-Level Work Tracker

| ID | Work item / phase | Workstream | Status | Dependencies / evidence | Exit condition |
| --- | --- | --- | --- | --- | --- |
| T-008 | Save and review PostgreSQL plan / documentation | Current session | DONE | E-030–E-035; V-020 | Only this handoff changes; current plan and historical evidence distinguished; check limits recorded |
| T-009 | Specification clarity, assertion dispositions and current baseline | WF-12 | DONE | S-012; V-015; E-036/E-037 | Current source and focused witnesses qualify movement |
| T-010 | Single template preparation owner | WF-13 | DONE | Section 6; S-013; V-016; E-038 | Duplicate algorithm and ordinary fallback removed with equivalent assurance |
| T-011 | Physical ownership and local dependencies | WF-14 | DONE | Section 6; S-014; V-017; E-039 | Explicit owner/borrower split; no compatibility wrappers/global hook mutation |
| T-012 | Consolidated assertions, policy and routing | WF-15 | DONE | Section 6; S-015; V-018; E-040 | Meaningful successors qualified, redundant tests/hooks/rows removed |
| T-013 | Finalization, final portfolio and handoff | WF-16 | DONE | Section 6; S-016; V-019; E-041 | Bounded implementation criteria met on final compatible source |
| T-014 | Review completed remediation and save findings | Current review | DONE | E-042/E-043; V-024 | Tracker-only review complete, evidence limitations and follow-up gates saved |
| T-015 | Preparation failure boundaries | WF-17 | DONE, E-045 | F-021/F-024; S-017; V-021 | Context and ownership-notification gaps closed |
| T-016 | Attachment and migration admission, local dependencies | WF-18 | DONE, E-046 | F-022/F-023/F-025/F-026; S-018; V-022 | Failed generation remains retired, scratch issuance is admitted, local registries isolated |
| T-017 | Correction validation and final handoff | WF-19 | DONE, E-047 | S-019; V-023; WF-18 exit | All required review findings qualified on compatible correction source |

Historical T-001–T-007 and WF-00–WF-11 remain DONE at their recorded baselines;
their IDs, F-001–F-014, C-001–C-010, S-000–S-011, V-001–V-014,
E-001–E-029 and RB-001–RB-004 are reserved. Documentation completion does not
complete or authorize any future work item.

## 10. Session Handoff Log

### WF-19 completion / E-047 — final handoff

**The closure effort is complete.** WF-17, WF-18 and WF-19 each finished
validation and saved its tracker checkpoint before the next began. All six
findings, reopened F-016–F-019 obligations, T-015–T-017, V-021–V-023 and RB-007
are closed against their actual witnesses. No implementation work remains in
this bounded plan. The next action is human review of the changes and projection
fidelity; there is no new requirement amendment or pending execution gate here.

Finalization ran before the broader portfolio and changed no executable input.
All final runs below share the exact identities:

- Source: `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- System/config capability: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
- Git source: dirty `main` at `867e97ed49829d671c58576232999d1fc9a4e3fb`;
  each manifest retains its selected graph/config, command and exact declared inputs.

The freeze-entry commands immediately below are the exact executed final
portfolio. Every root below has native outer exit **0**, terminal status **pass**,
retained secret scan **pass**, and observability audit exit **0**. For each exact
root, the audit command was
`CGO_ENABLED=0 make harness-observability-check RESULTS_DIR=.cartulary/test-results/<basename> CHECK=all`;
all returned `full_validity=1`. The native command exits were captured separately
from summary status. A later read-only `make test-run-status` on the transport
root exited 0 and reported terminal pass; its observation limitations do not
replace the captured native exit or the subsequent evidence audit.

| Exact root under `/home/jochi/code/cartulary/.cartulary/test-results/` | Result / outer / audit | Completed cleanup records |
| --- | --- | --- |
| `tr01-wf19-finalize-20261026-01` | pass / 0 / 0 | 10 |
| `tr01-wf19-lifecycle-20261026-01` | pass / 0 / 0 | 30 |
| `tr01-wf19-migrations-20261026-01` | pass / 0 / 0 | 21 |
| `tr01-wf19-postgres-20261026-01` | pass / 0 / 0 | 17 |
| `tr01-wf19-process-20261026-01` | pass / 0 / 0 | 21 |
| `tr01-wf19-transport-20261026-01` | pass / 0 / 0 | 12 |
| `tr01-wf19-recovery-20261026-01` | pass / 0 / 0 | 20 |
| `tr01-wf19-assessments-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-entities-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-evidence-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-graphprojection-20261026-01` | pass / 0 / 0 | 17 |
| `tr01-wf19-indicators-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-parties-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-reference_data-20261026-01` | pass / 0 / 0 | 17 |
| `tr01-wf19-reporting-20261026-01` | pass / 0 / 0 | 17 |
| `tr01-wf19-activity-20261026-01` | pass / 0 / 0 | 10 |
| `tr01-wf19-source-generated-20261026-01` | pass / 0 / 0 | 14 |
| `tr01-wf19-source-schema-20261026-01` | pass / 0 / 0 | 12 |
| `tr01-wf19-source-catalog-20261026-01` | pass / 0 / 0 | 11 |
| `tr01-wf19-source-identity-20261026-01` | pass / 0 / 0 | 12 |
| `tr01-wf19-lint-go-20261026-01` | pass / 0 / 0 | 15 |
| `tr01-wf19-lint-backend-boundary-20261026-01` | pass / 0 / 0 | 12 |

The portfolio covers **42 distinct rows**. All **338** reported cleanup records
completed, with no failed or blocked cleanup. Transport additionally retains
**64** service scopes with distinct physical allocation identities; each reports
physical settlement completed, passed child cleanup and no failures. The
qualification source/profile remains 16 serial lifetimes, 32 in four-way waves,
and 16 competing lifetimes, with the existing four-lane admission. Template
security/durability, transaction settlement, forced-drop/barrier bounds, reset,
roles/ACLs, process recovery and separately owned performance snapshots remain
selected. No product migration, S3 internals, browser workflow or performance
policy redesign was introduced.

Deliberate internal API changes and migration impact:

- `OpenPurposeDatabase` requires a caller context. All repository call sites were
  migrated atomically; no context-free wrapper remains.
- A failed owner template attempt permanently retires its generation. Start a
  fresh physical owner for another attempt; cleanup authority remains available.
- `DetachShared`, registry active/detach state and process detach callbacks are
  deleted. Borrowers release references; physical ownership stays with OwnedCluster.
- Migration factories require the admitted postgres_migration/exclusive binding.
  Three entity rows and evidence cutover retain IDs with corrected capability;
  graph, reference-data and reporting mixed rows split under the three approved
  IDs. Every assertion is selected exactly once in authored/generated routing.
- Migration timing publication accepts the explicit retained environment, removing
  the last ambient evidence dependency found in this seam. No legacy adapter,
  generic dependency container or new package hierarchy was added.

Failures and limitations are retained, not converted into passes. E-045 preserves
expected template characterization and nil-context lint failure; E-046 preserves
expected attachment characterization, two intermediate build failures and the
quarantine-diagnostic regression. Their corrective witnesses pass on the final
frozen source. The historical 13 E-041 roots still qualify only their original
source/selections. No failed final run was replaced or hidden; all 22 final graph
roots succeeded on their first frozen-source invocation.

`make lint-markdown` exited 0 at `tr01-wf19-lint-docs-20261026-01`. Its configured
globs exclude this tracker, so this document's statuses, IDs, tables, commands,
links, assertion successors and historical block were reviewed directly after
saving. Documentation lint is not a product/specification completeness test.

Retained-run maintenance was **skipped because RESULTS_DIR was unset** during
agent-finalize; no compatible successful full warm-check root exists. No blanket
owner-wide check, release gate, benchmark acceptance or instrumentation-required
performance assessment was run or claimed. The observability audits report
optional execution/resource diagnostics unavailable/not_observed in this normal
profile; that is not a performance claim. Primary-owner semantic review remains
human-owned. Separate preceding-handoff WF-33 remains unchanged and outside scope.

Preservation: all 7,338 frozen non-Markdown source/input hashes still match; the
historical details block is byte-identical to closure entry; staged Git entries
are unchanged. Exactly 33 entries differ from the closure baseline, all within
the approved fixture/test/caller/routing/tracker scope. Existing edits outside
those entries are preserved. Make may refresh index stat metadata and ignored
build/cache/results trees; those reproducible trees were not exhaustively
fingerprinted. No commit or staging was performed. Reproducible evidence and
comparison inventories are under `/tmp/cartulary-pg-closure-_vbuwtfo/`; the
controlling handoff remains this document alone.

### WF-19 freeze and exact final portfolio — E-047 entry

`env -u RESULTS_DIR CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-finalize-20261026-01 make agent-finalize`
exited 0. Inspection found no executable/input changes against the pre-finalize
snapshot. Retained-run maintenance was skipped because RESULTS_DIR was unset;
there is no compatible successful full warm-check root. Executable source is
now frozen. The final portfolio below is recorded before execution. Historical
13-root evidence and WF-17/WF-18 runs are not substituted for these obligations.

Frozen source digest: `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185`.
Toolchain/system identities are those recorded at E-045. The private freeze
inventory is `/tmp/cartulary-pg-closure-_vbuwtfo/final-source.json`; it excludes
Markdown, as do executable source identity inputs. Only tracker/documentation
updates may follow without reopening the final evidence compatibility review.

```sh
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-lifecycle-20261026-01 make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.unit.testservices_lifecycle_contract,harness.browser.integration.postgres_template_preparation,harness.browser.integration.postgres_transaction_isolation,harness.browser.integration.postgres_owned_barrier_observation,harness.browser.integration.postgres_single_drop,harness.browser.integration.postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277,harness.browser.integration.reset_checkpoint_characterization,harness.browser.integration.managed_service_assertions_for_runtime_reset_rec_79940f98ab,harness.browser.integration.performance_fixture_snapshot_lifecycle
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-migrations-20261026-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.concurrent_apply_locking,module.database_migrations.integration.production_ddl_v2_recurrence,module.database_migrations.integration.production_preflight_state_matrix,module.database_migrations.integration.schema_bootstrap,module.database_migrations.integration.schema_readiness_matrix,module.database_migrations.unit.test_harness_targeted_operation_validation,module.database_migrations.integration.cryptographic_state_admission
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-postgres-20261026-01 make test-slice OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle,platform.postgres.integration.fresh_fixture_admission,platform.postgres.integration.purpose_identity_roles_and_acl
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-process-20261026-01 make test-slice OWNER=app.server ROWS=app.server.support_unit.incident_create_commit_fault_requires_validated_efdc344831,app.server.process.effective_deployment_configuration_fails_the_rea_c17cbafdfc,app.server.process.a_fresh_process_reaches_health_and_ready_only_af_c93818219a
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-transport-20261026-01 make test-slice OWNER=harness.browser ROWS=harness.browser.integration.published_postgre_sql_transport_qualification_ac_38acf197c9
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-recovery-20261026-01 make test-slice OWNER=module.recovery ROWS=module.recovery.process.the_packaged_server_serves_the_coherently_restor_6b1731590e
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-assessments-20261026-01 make test-slice OWNER=module.assessments ROWS=module.assessments.store.assessment_create_idempotency_migration_preserve_b451ac58f2
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-entities-20261026-01 make test-slice OWNER=module.entities ROWS=module.entities.support_integration.entity_identifier_claims_37_3ba3e2aaaf,module.entities.support_integration.entity_source_integrity_migration_36_e6d97d1e54,module.entities.support_integration.timeline_mention_projection_upgrade_preserves_so_dabe8cb065
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-evidence-20261026-01 make test-slice OWNER=module.evidence ROWS=module.evidence.integration.verification_cutover
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-graphprojection-20261026-01 make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.migration_reset,module.graphprojection.storage.graph_projection_current_head_schema_preserves_r_cf74406ebd
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-indicators-20261026-01 make test-slice OWNER=module.indicators ROWS=module.indicators.storage.lifecycle_support_reference_integrity
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-parties-20261026-01 make test-slice OWNER=module.parties ROWS=module.parties.claims.party_active_claims_are_migration_safe_concurren_9ad2db3764
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-reference_data-20261026-01 make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.integration.canonical_reference_data_cutovers_reject_incompa_3682e9346f
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-reporting-20261026-01 make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c,module.reporting.integration.reporting_identity_cutover_preserves_incompatibl_62c39cd14e
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-activity-20261026-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-source-generated-20261026-01 make source-check CHECK=generated
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-source-schema-20261026-01 make source-check CHECK=schema
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-source-catalog-20261026-01 make source-check CHECK=catalog
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-source-identity-20261026-01 make source-check CHECK=identity
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-lint-go-20261026-01 make lint CHECK=go
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr01-wf19-lint-backend-boundary-20261026-01 make lint CHECK=backend-boundary
```

### WF-18 checkpoint / E-046 — admission, binding and routing qualified

WF-18/S-018, T-016 and V-022 are DONE. F-022/F-023/F-025/F-026 are corrected;
all six findings now have implementation witnesses, with overall acceptance
still waiting for WF-19. The adopted requirements were not changed.

Attachment admission now stores a pending/terminal result per advertised
generation. Inspection must confirm identity before readiness/publication
failure can retire that generation. Same-binding callers share the result;
cancelled waiters leave the initiating attempt intact, and another generation
can proceed during stalled admission. Unverified claims confer no ownership.
Readiness cancellation and journal failure retain original causes and quarantine;
subsequent calls neither probe nor prepare again.

Fixtures copy explicit binding inputs at entry. Registry reuse compares connection,
capability, policy, dependency and evidence-destination inputs. Creation guards,
policy resolution, migration admission and events use that retained binding.
Production entrypoints capture process inputs; private registry calls receive
explicit inputs. The existing explicit testservices publisher is preserved.
Migration timing publication was another ambient evidence consumer, discovered
while tracing the same boundary: `suiteservices.MeasureMigration(ctx, env, fn)`
now takes the admitted inputs at its two pgtest callers. Its existing private
publication/shutdown witness passes with deliberately conflicting ambient inputs.
No generic dependency container, scheduler change or compatibility adapter was added.

Both migration factories require postgres_migration plus exclusive scope before
CREATE/open/source access. The negative matrix crosses missing, unknown,
transaction, dedicated and migration capability with absent/shared/exclusive/
unknown scope, for both factories; the valid pair is exercised by existing fake
issuance/canonical-source contracts and real caller rows. Invalid target bounds
and forged/zero capability rejection remain selected. Physical ownership alone
adds no migration authority. Four migration-only rows change capability; the
three mixed rows split exactly as recorded below, retaining IDs, verification
links, aggregate placement, dependencies and runtime/resource profiles.

`DetachShared`, registry active state/detachment and process detach callbacks
are removed without aliases. The process attachment helper returns its original
object-store error. The real template witness checks the owner remains usable
and the same generation remains admitted after borrower use; quarantine tests
check repeated attachment cannot restore eligibility. No physical shutdown is
exposed by Harness. Historical detach descriptions in E-039/E-041 are history,
not current API guidance.

The caller audit traces 43 top-level product tests through local helper calls to
15 migration rows, with no unmapped test. The exact 19-row portfolio below also
includes the unit capability row and ordinary assertions retained by row splits.
All nine owner task guides and harness.evidence_accounting/app.server guides
exited 0. No owner-wide rerun was substituted for exact caller selections.

Validation used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<basename>` from the repo
root. For each root, run-manifest.json retains exact target/OWNER/ROWS/CHECK and
source/tool/system identities. Commands are public `make test-slice` using A/B/D,
the recorded caller selection, or the following additions: harness A appends
`harness.browser.integration.postgres_template_preparation` and
`harness.browser.integration.postgres_transaction_isolation`; migration selection
is the recorded module.database_migrations caller list plus B's cryptographic
admission row; unit-postgres is C's certificate_fixture_lifecycle row;
unit-migrations is B's test_harness_targeted_operation_validation row.
`activity` uses the exact evidence-accounting row recorded below.

| Root under `.cartulary/test-results/` | Public target / selected owner or check | Outer exit | Source digest | Result |
| --- | --- | --- | --- | --- |
| `tr01-wf18-activity-20261026-01` | `make test-slice OWNER=harness.evidence_accounting` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-assessments-20261026-01` | `make test-slice OWNER=module.assessments` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-characterization-20261026-01` | `make test-slice OWNER=harness.browser` | 2 | `sha256:f2c2e40a8ef08268561bbf910d7029ee6bf78f0ffa665b5f19b1999d2347b98e` | fail |
| `tr01-wf18-entities-20261026-01` | `make test-slice OWNER=module.entities` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-evidence-20261026-01` | `make test-slice OWNER=module.evidence` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-format-20261026-01` | `make format` | 0 | `sha256:7536ca4652ec653d2d1a4b58c32bbf539cf4595defc12d0492aef9a275d4c75b` | pass |
| `tr01-wf18-format-20261026-02` | `make format` | 0 | `sha256:468116d6bb0bcdce931be8ea2c157b3466f37d1c3b5c03c85fe1228485bceb45` | pass |
| `tr01-wf18-format-20261026-03` | `make format` | 0 | `sha256:f5a456c76eb789b6cb0cd33181e59e157af77ef3609da6688c0bcf6d7ad32e01` | pass |
| `tr01-wf18-format-20261026-04` | `make format` | 0 | `sha256:aad088c4b1fe25695954555aebe6c74f98f0cbbde9ee98a624a7513c626a53ee` | pass |
| `tr01-wf18-graphprojection-20261026-01` | `make test-slice OWNER=module.graphprojection` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-harness-20261026-01` | `make test-slice OWNER=harness.browser` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-indicators-20261026-01` | `make test-slice OWNER=module.indicators` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-lint-boundary-20261026-01` | `make lint CHECK=backend-boundary` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-lint-go-20261026-01` | `make lint CHECK=go` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-lint-go-20261026-02` | `make lint CHECK=go` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-lint-go-20261026-03` | `make lint CHECK=go` | 0 | `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185` | pass |
| `tr01-wf18-migrations-20261026-01` | `make test-slice OWNER=module.database_migrations` | 2 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | fail |
| `tr01-wf18-migrations-20261026-02` | `make test-slice OWNER=module.database_migrations` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-parties-20261026-01` | `make test-slice OWNER=module.parties` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-process-20261026-01` | `make test-slice OWNER=app.server` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-reference_data-20261026-01` | `make test-slice OWNER=module.reference_data` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-reporting-20261026-01` | `make test-slice OWNER=module.reporting` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-source-catalog-20261026-01` | `make source-check CHECK=catalog` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-source-generated-20261026-01` | `make source-check CHECK=generated` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-source-identity-20261026-01` | `make source-check CHECK=identity` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-source-schema-20261026-01` | `make source-check CHECK=schema` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-unit-20261026-01` | `make test-slice OWNER=harness.browser` | 0 | `sha256:0d13eb2ee84ae2202101891588a2376295c0d6b69c2e7b14e21290562bfb5ff9` | pass |
| `tr01-wf18-unit-harness-20261026-02` | `make test-slice OWNER=harness.browser` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-unit-migrations-20261026-02` | `make test-slice OWNER=module.database_migrations` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-unit-migrations-20261026-03` | `make test-slice OWNER=module.database_migrations` | 0 | `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185` | pass |
| `tr01-wf18-unit-postgres-20261026-01` | `make test-slice OWNER=platform.postgres` | 0 | `sha256:a1b23df04a5f7651c8da16cb1d662fa0043b2d47df17e984546bd2763b9eab8b` | pass |
| `tr01-wf18-unit-postgres-20261026-02` | `make test-slice OWNER=platform.postgres` | 0 | `sha256:1c29ea64d0f1bbb70a149f33e1182f262773c29740e0a9651fcd20d4b567dbc1` | pass |
| `tr01-wf18-unit-postgres-20261026-03` | `make test-slice OWNER=platform.postgres` | 0 | `sha256:e84fcbaa0923e1b4ca2bc0798fe16a6116926c9f7516398e1d1fba7c56ed2185` | pass |

Additional commands: `make generate` exited 0 at
`tr01-wf18-generate-20261026-01`; `make lint-markdown` exited 0 at
`tr01-wf18-lint-docs-20261026-01`. The tracker is outside lint globs and was
reviewed directly. Two intermediate `make format` invocations exited 2 before
the graph could start, roots `20261026T231734Z-p72355` and
`20261026T231758Z-p73018`: incomplete receiver/binding signature migration
prevented building the harness. The build stderr identified the unset binding
in startOwnedAttempt; passing the entry snapshot through that private call
closed the defect. Subsequent format and Go lint roots above pass.

Characterization failed as expected with product/test_assertion_failure in the
helper row. The first full migration selection then found a real diagnostic
regression: new admission returned generic text for quarantine, and policy
resolution could obscure the quarantine witness. The repair gives quarantine
precedence and uses the existing typed fixture-failure publisher. Unit successors
and the entire seven-row migration selection pass afterward. These failed roots
are preserved. Final small test changes add a waiting preparation deadline and
remove unused policy table fields; their affected unit rows and Go lint pass.

All successful roots report completed cleanup, with no failed or blocked cleanup.
Toolchain/system identities remain those in E-045. Source identity changes are
recorded per slice run; WF-19 must freeze and requalify compatible final evidence.
The historical details block remains byte-identical to closure entry. No files
outside the 33 scoped source/test/routing/tracker entries differ from the closure
baseline; generated render-index changes came from Make. The Git index's stat
cache may refresh during canonical tools; no staging was requested or performed.

Next gate: only after this saved checkpoint, run `make agent-finalize` with
RESULTS_DIR unset (there is no compatible successful full warm check root),
inspect its edits, freeze executable inputs, run final A–E plus the exact caller
and timing selections, and audit successful roots. Acceptance, T-017 and V-023
remain open until the final handoff checkpoint is saved.

### WF-18 exact migration caller selections — recorded before execution

The 16 prior caller rows plus three approved splits are the additional caller
portfolio. Includes helper-mediated calls (`migratedSchemaReadinessDatabase`,
`emptyMigrationDatabase`, reference-data cutover helpers, and party clock cutover).
Retained ordinary assertions in split rows are included to detect routing loss.
Run each command with `CGO_ENABLED=0` and a fresh recorded run ID. B and this
inventory overlap intentionally at two migration rows; combine them into one
owner selection for the final portfolio. The narrow evidence-accounting row is
added because migration timing publication now receives explicit bound inputs.

```sh
make test-slice OWNER=module.assessments ROWS=module.assessments.store.assessment_create_idempotency_migration_preserve_b451ac58f2
make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.concurrent_apply_locking,module.database_migrations.integration.production_ddl_v2_recurrence,module.database_migrations.integration.production_preflight_state_matrix,module.database_migrations.integration.schema_bootstrap,module.database_migrations.integration.schema_readiness_matrix,module.database_migrations.unit.test_harness_targeted_operation_validation
make test-slice OWNER=module.entities ROWS=module.entities.support_integration.entity_identifier_claims_37_3ba3e2aaaf,module.entities.support_integration.entity_source_integrity_migration_36_e6d97d1e54,module.entities.support_integration.timeline_mention_projection_upgrade_preserves_so_dabe8cb065
make test-slice OWNER=module.evidence ROWS=module.evidence.integration.verification_cutover
make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.migration_reset,module.graphprojection.storage.graph_projection_current_head_schema_preserves_r_cf74406ebd
make test-slice OWNER=module.indicators ROWS=module.indicators.storage.lifecycle_support_reference_integrity
make test-slice OWNER=module.parties ROWS=module.parties.claims.party_active_claims_are_migration_safe_concurren_9ad2db3764
make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.integration.canonical_reference_data_cutovers_reject_incompa_3682e9346f
make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c,module.reporting.integration.reporting_identity_cutover_preserves_incompatibl_62c39cd14e
make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.local_migration_activities
```

### WF-17 checkpoint / E-045 — preparation recovery qualified

WF-17/S-017 and T-015 are DONE. F-021 and F-024 are corrected against the
unchanged adopted obligations. `OpenPurposeDatabase(ctx, dsn, purpose)` now
rejects nil/cancelled contexts before I/O and preserves context causes from
production admission. All repository callers supply an operation/test context.
The independent crypto guard receives the preparation context. One owner-local
attempt runs synchronously with the existing two-minute bound; waiters cancel
independently, and a failed attempt is terminal only after settlement. Invalid
requests do not retire the owner. There is no in-place retry or background recipe.

Pristine CREATE and the existing clone finalization callback synchronously record
local ownership and publish the created event before provisioning. Notification
failure stops dependent work; disposal runs after catalog admission returns.
Rejected CREATE grants no ownership. Exported low-level CloneDatabase remains
available to recovery and performance owners. The real template witness observes
both ownership ledgers at notification for pristine and cloned databases.

Assertion additions are nested beneath the existing selected template contract:
retained terminal failures, successful concurrent preparation, independently
cancelled waiter, both stalled opens with cancellation/deadline, settlement-before-
terminal publication, nil/invalid pre-admission, and real TCP connection-open
cancellation/deadline. Existing close/drop failure and unowned-target assertions
remain. No top-level test or selector was retired in this slice.

Commands below ran from the repository root with `CGO_ENABLED=0` and
`CARTULARY_TEST_RUN_ID=<root basename>`. Exact selections: **A** is Section 8 A;
**B** is Section 8 B; **C-unit** is
`OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle`;
**A+template** is A with
`harness.browser.integration.postgres_template_preparation` appended to ROWS.
All test commands are `make test-slice` with those exact arguments.

| Run root under `.cartulary/test-results/` | Command / selection | Outer exit | Source digest | Result |
| --- | --- | --- | --- | --- |
| `tr01-wf17-characterization-20261026-01` | A | 0 | `sha256:b3365868908ddc3c0d7ad51ec0941edc8995988041bf097b3f68acfa45264807` | pass |
| `tr01-wf17-characterization-20261026-02` | C-unit | 2 | `sha256:b3365868908ddc3c0d7ad51ec0941edc8995988041bf097b3f68acfa45264807` | fail |
| `tr01-wf17-unit-20261026-01` | C-unit | 0 | `sha256:e586dd2240d92f5e92f08c3b0c4ee9e19d2c92231d430957168c2a1ad2433133` | pass |
| `tr01-wf17-unit-20261026-02` | C-unit | 0 | `sha256:d941b537a26c01b20dc30cd6d0522d132c9c84e657009e1c63e42a43ee31bb37` | pass |
| `tr01-wf17-harness-20261026-01` | A+template | 0 | `sha256:d941b537a26c01b20dc30cd6d0522d132c9c84e657009e1c63e42a43ee31bb37` | pass |
| `tr01-wf17-migrations-20261026-01` | B | 0 | `sha256:d941b537a26c01b20dc30cd6d0522d132c9c84e657009e1c63e42a43ee31bb37` | pass |
| `tr01-wf17-lint-go-20261026-01` | make lint CHECK=go | 2 | `sha256:d941b537a26c01b20dc30cd6d0522d132c9c84e657009e1c63e42a43ee31bb37` | fail |
| `tr01-wf17-lint-go-20261026-02` | make lint CHECK=go | 0 | `sha256:70d6064db870a4258e31837e5e077ff77860d39fa68aed2ab0353909bbd5bd6f` | pass |

The first characterization A pass did not select the newly changed template
contract; C-unit then failed with typed `product/test_assertion_failure` (unit
exit 10, outer exit 2), qualifying the retained-failure/waiter gap. No pass was
claimed for that uncovered boundary. Go lint's first failure was SA1012 for the
intentional nil-context rejection test; a one-line test-only annotation explains
that negative case. The corrected lint passes and compiles migrated callers.
No product code suppression was added.

All successful test roots report completed cleanup with no cleanup failures.
Toolchain digest remains `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`;
system digest remains `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Source digests differ during correction and comment-only lint repair; these are
slice evidence, not the frozen final portfolio. WF-19 must requalify final source.
`make help ALL=1`, task guides for harness.browser, platform.postgres and
module.database_migrations all exited 0. `make format` exited 0 at roots
`20261026T230846Z-p15398`, `20261026T231004Z-p22998`, and
`20261026T231100Z-p29881`. `make lint-markdown` exited 0 at
`tr01-wf17-lint-docs-20261026-01`; its configured globs omit this tracker,
which was re-read and reviewed directly before this checkpoint.

Next gate: WF-18 may begin only after this checkpoint is saved. Characterize
verified-generation readmission, binding isolation and migration admission;
remove DetachShared and redundant state; correct/split caller routing; then
regenerate and validate V-022. WF-19 final acceptance remains open. No
complete-owner, release or performance acceptance is asserted.

### Closure entry / E-044 — WF-17 authorized

The implementation request adopts the complete closure plan in this conversation.
Two decisions supersede the review's optional alternatives: remove DetachShared
and its unused state/callbacks, and retain one terminal preparation result per
physical owner, retiring the generation after failed preparation. Cancellation
of a waiting caller does not cancel the initiating attempt. Invalid requests
rejected before an attempt do not retire a healthy owner.

Source baseline is main at `867e97ed49829d671c58576232999d1fc9a4e3fb`, with 33
pre-existing changed/new paths and 7,649 tracked/nonignored entries. Preservation
material is `/tmp/cartulary-pg-closure-_vbuwtfo`; starting tracker SHA-256 is
`d14d9f9130113893ff75b2abc7a86c1cfdf458cb8c3d869304f0ee6a2d7089d5`.
Read root/internal/tools instructions and test-operations skill. REQ-804/405/810/811
already own the required behavior; no normative amendment is needed. Supporting
API guidance will be corrected with callers. No executable check consumes Markdown.

WF-17 implements context-aware purpose opens, one terminal preparation attempt,
and synchronous ownership notification after confirmed CREATE and before
prerequisites/clone grants. WF-18 implements verified-generation failure retention,
immutable fixture inputs/event destinations, strict migration-capability admission,
and DetachShared removal. Its routing correction changes the four migration-only
rows in entities/evidence and splits mixed graph/reference-data/reporting rows.
The three new IDs are the previously derived IDs in the approved plan. Preserve
profiles, dependencies, verification links and exclusive isolation. WF-19 finalizes,
freezes source, validates A–E plus migration-caller rows, audits exact roots and
saves handoff. Every exit updates findings, assertion mapping, selections, work
status, blockers and exact evidence before the next slice begins.

Initial validation is V-021: A, B and real template preparation. Add fault
characterization before implementation and classify expected failures. Run all
repository verification through Make, record native sessions and outer exits,
and retain failures. Broader final verification waits for WF-19's finalization.
No correction validation has run at this entry checkpoint.

### Post-implementation review — E-042

This review supersedes E-041's unconditional completion assessment, not its
historical commands/results or completed workstream checkpoints. F-021–F-025
remain required; F-026 is optional. The existing design is worth retaining, with
bounded corrections in WF-17–WF-19. Implementation authorization is **not granted
by this review invocation**. No production refactor, repair or product rerun was
performed.

Review baseline: `main`, HEAD `867e97ed49829d671c58576232999d1fc9a4e3fb`;
21 already modified tracked files and 12 nonignored new pgtest files; 7,649
path/content/identity entries. Starting tracker SHA-256:
`80eaf2857cdb07dc35a22564206920881aaae4d5262d716b5a353091010fee9f`.
The admitted output is the same regular, singly linked canonical file. External
baseline and review records are at `/tmp/cartulary-pg-review-ghyk5hap`.
A plain initial Git status preceded the optional-write-disabled baseline; any
prior stat-cache refresh is outside the preservation comparison. All subsequent
Git inspection disabled optional writes. No staging or commits were performed.

Source review used bounded `rg`, `sed`, numbered source reads and
`git --no-optional-locks diff`, with relevant dependencies opened directly.
Searches that referenced nonexistent guessed helper paths were corrected by file
and symbol discovery; they produced no edits or validation evidence. Adopted
Harness REQ-804/405/810/811 and the plan govern the findings; `domain.md` is
vocabulary/navigation. No owner contradiction or necessary spec amendment was
found. The existing pool/readmission wording change preserves the requirement.

All thirteen E-041 run manifests and summaries still report pass (70 total units,
zero failed/skipped/cancelled). Source/toolchain/system identities match E-041;
all retained cleanup outcomes are `completed`, secret scans are `pass`, and the
transport root contains 64 service-scope records. Direct hash comparison of the
7,338-file E-041 freeze found no executable/input change. This corroborates
retained results; it does not reproduce faults, rerun observability qualification,
or establish coverage of F-021–F-025. `evidence-review.json` contains the inspected
root identities and results. Tests, services, generation, `agent-finalize` and
new observability runs were intentionally not invoked for this review-only task.

### Review documentation save — E-043

V-024 completed. The inspected lint path used installed-only Node/frontend
readiness, the existing workspace lock, fixing disabled, and external result/cache
directories. Exact command from the repository root:

```sh
env GIT_OPTIONAL_LOCKS=0 \
  CARTULARY_PREPARATION_POLICY=installed_only \
  CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-pg-review-ghyk5hap/results \
  CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-pg-review-ghyk5hap/static-analysis \
  CARTULARY_TEST_RUN_ID=tr01-postgres-review-doc-20261026-01 \
  COREPACK_ENABLE_NETWORK=0 make lint-markdown
```

The lint summary records **pass, exit 0**, 17,650 ms, completed
`2026-10-26T22:27:16.628Z`, zero tests. Exact artifact:
`/tmp/cartulary-pg-review-ghyk5hap/results/tr01-postgres-review-doc-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The asynchronous command's outer Make exit was not retained by this review's
monitor; the recorded exit above is the lint command's tool summary. No retry was
run to replace missing monitor metadata. Current lint globs exclude this tracker.
Direct review checked its twelve sections, balanced fences/tables, local links,
stable IDs, contract/finding/slice/validation references, preserved historical
snapshot and distinction between source findings and executed evidence.
`git --no-optional-locks diff --check` passed. Final evidence annotations do not
change the linted file population and were reviewed directly.

The final comparison against all 7,649 starting entries found **only this tracker
changed**, no new/deleted paths, and identical raw index and index entries. The
embedded historical snapshot remains byte-identical. All other tracked and
nonignored untracked content/file identities, including the prior implementation
edits, are preserved; no staging or commit occurred. External comparison records
are beside the review baseline. Ignored build/dependency/output trees were not
exhaustively fingerprinted; reviewed command paths redirected their outputs.
T-014 is DONE. Next entry is later-authorized WF-17 characterization, not an
implementation action in this review. Required acceptance remains open.

### Current implementation — WF-12 entry / E-036

The user authorized WF-12 through WF-16 in dependency order, with a saved exit
checkpoint before every successor. Starting HEAD is
`867e97ed49829d671c58576232999d1fc9a4e3fb`; the index is unchanged and the sole
pre-existing modification is this tracker. Content/file-identity fingerprints
for 7,637 tracked/nonignored entries, the original tracker, HEAD and index are
retained in `/tmp/cartulary-pg-remediation-kmpj28se`. Ignored outputs and installed
dependencies are not exhaustively fingerprinted. E-030–E-035 remain historical
planning evidence, not this implementation baseline.

Source and owner revalidation confirms F-015–F-020. F-016 additionally covers
discarded close/drop errors and preservation of confirmed CREATE ownership after
later preparation failures. Failed CREATE never grants ownership of an existing
target. The authorized specification edit replaces “returned to the pool” with
“readmitted for reuse” in §11.0; it introduces no pooling or behavioral change.
Domain vocabulary and product specifications require no amendment.

V-015 includes A–D plus real template preparation, transaction isolation,
owned cleanup/barrier and the explicit transport qualification before movement.
Each run uses a fresh literal identity and `CGO_ENABLED=0`; exact commands,
outer exits, artifacts and source identities will be saved at the WF-12 exit.
The existing assertion successor ledger is controlling; no incidental method
count, smoke-only assertion or borrowed-shutdown error becomes a requirement.

Final API decisions: `StartOwned(ctx, StartOptions) (*OwnedCluster, error)`,
`OwnedCluster.Fixture()`, owner-only explicit template preparation and physical
shutdown, attachment-only `Start`/`StartShared`, and no-I/O `DetachShared()`.
There will be no constructor aliases or physical shutdown on `Harness`.

### WF-12 exit / E-037 — characterization qualified

WF-12 completed before WF-13 began. All six commands used
`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<ID> make test-slice OWNER=<owner> ROWS=<selection>`
from the repository root, exited 0, and share source
`sha256:daaea890979e0349dc4392aeffd167b41b172981b29d13dca99008fbad112eb9`
and toolchain
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
Run roots are `.cartulary/test-results/<ID>`; the existing selections in §8
give exact A–D row arguments.

| Run ID | Selection | Duration ms | Result |
| --- | --- | --- | --- |
| `tr01-wf12-contract-20261026-01` | A; harness.browser | 10977 | pass, 2/2 units |
| `tr01-wf12-migrations-20261026-01` | B; module.database_migrations | 22391 | pass, 9/9 units |
| `tr01-wf12-postgres-20261026-01` | C; platform.postgres | 14078 | pass, 5/5 units |
| `tr01-wf12-process-20261026-01` | D; app.server | 88243 | pass, 10/10 units |
| `tr01-wf12-lifecycle-20261026-01` | harness.browser: `postgres_template_preparation`, `postgres_transaction_isolation`, `postgres_owned_barrier_observation`, `postgres_single_drop`, `postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277`, all under `.integration.` | 19655 | pass, 6/6 units |
| `tr01-wf12-transport-20261026-01` | harness.browser: `harness.browser.integration.published_postgre_sql_transport_qualification_ac_38acf197c9` | 96331 | pass; 64 fresh lifetimes |

Each exact root passed `make harness-observability-check RESULTS_DIR=<root>
CHECK=all` with outer exit 0 and `full_validity=1`; cleanup completed and retained
secret scans passed. There were no failed/skipped/cancelled units. These are
pre-move evidence only. `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=tr01-wf12-markdown-20261026-01 make lint-markdown` exited 0;
summary: `.cartulary/test-results/tr01-wf12-markdown-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
Its configured globs exclude this tracker, which received direct structure and
diff review; `git diff --check` passed. No product/source movement occurred in
WF-12. The meaningful successor/retirement ledger in §7 remains complete; RB-005
is resolved and S-013 is ready. Human projection-fidelity review is not inferred
from test success. Next: canonical preparation, explicit admission and failure
ownership in WF-13.

### WF-13 exit / E-038 — canonical preparation qualified

One pgtest operation now prepares the ordinary template and returns database/error.
Owned callers prepare explicitly; ordinary clone preparation fails without a sealed
template. Service composition delegates through its physical service capability
and supplies the lifecycle publisher; duplicate create/migrate events and its
SQL/migration recipe were removed. Confirmed CREATE is retained through subsequent
errors; rejected CREATE grants no deletion authority. Template handle/guard close,
drop and event-publication failures are retained with their primary cause and
uncertainty prevents reuse. Fault cases cover create rejection, committed-create
failure, create/migrate event publication, open, initialization, cancellation,
close, seal and combined close/drop failure. Real negative cases now use independent
owned lifetimes rather than accidentally stopping at a previous quarantine.

Validation used `CGO_ENABLED=0`, fresh IDs and public Make commands. Two initial
build failures were diagnosed and corrected: an event receiver typo blocked
`tr01-wf13-format-20261026-01` (build step summary under `harness-runtime-build`),
and an unused import after deleting the duplicate recipe caused
`tr01-wf13-contract-20261026-01` to report `infra/preflight_error` with row
`missing_selector_result`. These are implementation compile failures, not service
instability; their artifacts remain. Format passed at `tr01-wf13-format-20261026-03`.

The exit portfolio shares source
`sha256:9c4615b746a1bab5ae757a313fe7de211200adbf915be1651a22b403d7e6f6a9`:

| Run ID | Exact selection | Duration ms | Result |
| --- | --- | --- | --- |
| `tr01-wf13-lifecycle-20261026-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.unit.testservices_lifecycle_contract,harness.browser.integration.postgres_template_preparation` | 29441 | pass |
| `tr01-wf13-migrations-20261026-01` | Selection B in §8 | 21595 | pass |
| `tr01-wf13-postgres-20261026-02` | Selection C in §8, including the new fault matrix | 20809 | pass |

Each root `.cartulary/test-results/<ID>` passed `make
harness-observability-check RESULTS_DIR=<root> CHECK=all`, outer exit 0,
`full_validity=1`. Native commands exited 0; cleanup completed and retained-secret
scans passed. Earlier WF-13 passes qualify earlier source only. No selector change
was required; the new fault matrix is selected by the existing template contract.
This checkpoint closes WF-13 before WF-14 begins. Next: move physical lifetime and
the explicit preparation method to OwnedCluster, then isolate dependencies and
attachment registries without aliases.

### WF-14 exit / E-039 — ownership and dependencies qualified

`StartOwned(ctx, StartOptions)` now returns `*OwnedCluster`. Its callbacks retain
physical recovery authority before readiness; `Fixture()` and template preparation
remain unavailable until admission completes. The owner holds termination, TLS,
readiness, diagnostics and template preparation. `Harness` has database ownership,
transactions and quarantine, with no physical close or embedded owner. Direct
callers migrated without wrappers or aliases; closure retires fixture admission.
Private startup/database/template/diagnostic operations replace mutable function
hooks. Tests construct independent attachment registries. Every cache hit validates
the complete admitted binding identity; detach preserves generation records,
ownership and quarantine. Independently admitted generations remain usable.

The process fixture now uses `DetachShared()` on object-store attachment failure
and returns the initiating error unchanged. Real template integration proves local
detach leaves PostgreSQL alive and keeps the admitted fixture record. C-011/C-013/
C-015 and F-015/F-017 are implemented and qualified at this checkpoint; F-019's
process boundary is complete, with capability/smoke consolidation left to WF-15.
The original assertion IDs remain selected; owner lifecycle assertions now use
local operations and `OwnedCluster` callbacks. No mutable PG hook substitutions,
borrowed shutdown, `StartOwnedWithLabels`, or `StartOwnedWithOptions` remain.

All commands used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<ID> make test-slice` with
§8 selections. Composite commands use the comma-joined union of the stated rows
under `OWNER=harness.browser`. Results are under `.cartulary/test-results/<ID>`:

| Run ID | Exact row selection | Source digest | Duration ms | Native result / outer exit |
| --- | --- | --- | --- | --- |
| `tr01-wf14-migrations-20261026-01` | B | `sha256:37bb7ebfb485107a7915cd4fb8d6cf5dd7c2e7d0c95fe2d89511cf17273a609b` | 24022 | pass / 0 |
| `tr01-wf14-lifecycle-20261026-02` | A plus first E command | `sha256:d1a13e83352aad66d823eace2650f6d922cf5a170fc28643b078ba0ce3417c8c` | 60262 | pass / 0 |
| `tr01-wf14-process-20261026-01` | D | `sha256:d1a13e83352aad66d823eace2650f6d922cf5a170fc28643b078ba0ce3417c8c` | 78568 | pass / 0 |
| `tr01-wf14-postgres-20261026-02` | C | `sha256:cb6328e926106f4d204cbca33db4c1d5302c8ba253c9a46c96d1a428154ef5a8` | 16978 | pass / 0 |
| `tr01-wf14-transport-20261026-01` | A plus template preparation and explicit transport row | `sha256:cb6328e926106f4d204cbca33db4c1d5302c8ba253c9a46c96d1a428154ef5a8` | 105201 | pass / 0 |

The final transport command selected exactly A's two rows,
`harness.browser.integration.postgres_template_preparation`, and
`harness.browser.integration.published_postgre_sql_transport_qualification_ac_38acf197c9`.
It retained all 64 fresh lifetimes and the existing four-way profile. Each listed
root passed `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`,
outer exit 0, `full_validity=1`; cleanup and retained-secret checks passed. The
source identities above deliberately distinguish checks made before the final
pre-readiness admission guard; C, A, real preparation and transport were rerun
after that guard. WF-16 will establish the compatible final portfolio.

Two implementation build failures were diagnosed before retry: format run
`tr01-wf14-format-20261026-01` had moved-import/receiver signature errors, recorded
in `harness-runtime-build/build-harness-runtime/step-summary.json`; lifecycle run
`tr01-wf14-lifecycle-20261026-01` had two borrowed forced-drop callers incorrectly
changed to `Fixture()`, reported as `infra/preflight_error` and
`missing_selector_result` in its two postgrescleanup unit logs. Both exited 2;
both are corrected and retained. Final format run
`tr01-wf14-format-20261026-04` passed. An attempted `make test-status` observation
was rejected because that target does not exist; corrected to public
`make test-run-status` without executing tests. No unresolved failure remains.

This checkpoint closes WF-14 before WF-15 starts. Next: consolidate policy and
cleanup assertions, qualify capability/smoke successors, update authored selectors,
and regenerate projections. RB-006 now gates only WF-15 exit and WF-16 entry.

### WF-15 exit / E-040 — assertions and routing consolidated

The cleanup successor passed first as a nested case under the existing targeted
capability selector (`tr01-wf15-successors-20261026-01`, selection B, exit 0,
24418 ms). Only then were the four standalone/attached clone/scratch cleanup tests
retired and `TestDatabaseFixtureCleanup` selected directly by the existing
`module.database_migrations.unit.test_harness_targeted_operation_validation` row.
Its nine cases preserve resource kind and standalone/attached/persistent borrowing
contexts, issue real private capabilities, and verify closure before exact eager
drop. Single-drop ownership, caller-context propagation and failed-recovery tests
remain separate because they prove distinct failure boundaries.

The weak `TestHarnessStartsPostgresAndRunsCurrentMigrationPath` and selector were
removed after the existing `TestPrepareIsolatedDatabaseTReturnsMigratedDatabase`
passed C at WF-14; the surviving witness proves schema and crypto admission.
All row IDs, owners, service dependencies, isolation and qualification profiles
remain unchanged. No alias maps were added. Authored changes are only the test
arrays in `module.database_migrations.json` and `platform.postgres.json`;
`make generate` updated only the corresponding generated topology input hashes.

The public policy helper now calls one resolver. Its test exposed attribution
that skipped test files inside pgtest along with implementation frames; retaining
actual `_test.go` caller frames fixes package selection without inventing a test
entrypoint. The first policy run `tr01-wf15-policy-20261026-01` failed with
`product/test_assertion_failure` (outer exit 2, 6354 ms); its logs identify
`package_priority`. After correction `tr01-wf15-policy-20261026-02` passed in
7028 ms. Capability assertions now reject nil/zero/forged identities, check source
and execution counters before rejected targets, and compare legal operations to
the canonical embedded source. The private identity token is nonzero-sized so
its pointer cannot alias an unrelated zero-sized allocation. Reflection API-shape
assertions and obsolete local restore scaffolding are gone. Tests are grouped by
responsibility in the same package, without introducing a package hierarchy.
Concurrent initial template preparation now proves exactly one recipe/publication.

All A–D successors passed on source
`sha256:b3ef01cf9ea1fa6973c4e3aec96dc8522a01c3fff3af3b366bd31bd262e55e1a`:

| Run ID | Exact §8 selection | Duration ms | Native result / outer exit |
| --- | --- | --- | --- |
| `tr01-wf15-contract-20261026-01` | A | 7372 | pass / 0 |
| `tr01-wf15-migrations-20261026-01` | B successor | 30654 | pass / 0 |
| `tr01-wf15-postgres-20261026-01` | C successor | 14558 | pass / 0 |
| `tr01-wf15-process-20261026-01` | D | 78087 | pass / 0 |

The added private canonical-source load counter was then qualified by B at
`tr01-wf15-migrations-20261026-02`, exit 0, 28593 ms, source
`sha256:04e6276a37b9a7e9ca342911b43a12258842424698d8722ff213eac36879c036`.
Each listed test root passed `make harness-observability-check
RESULTS_DIR=.cartulary/test-results/<ID> CHECK=all`, exit 0, `full_validity=1`.
Cleanup completed and retained-secret scans passed. These workstream roots are
not substituted for WF-16's compatible final-source portfolio.

Public commands used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<ID>`. Generation passed
at `tr01-wf15-generate-20261026-01` (10280 ms). `make source-check CHECK=<check>`
passed for `generated`, `schema`, `catalog`, `identity`, using roots
`tr01-wf15-source-<check>-20261026-01` (15645, 6206, 1643, 724 ms respectively),
all exits 0. A format build failed once while moving declarations because two
still-used imports had been removed (`tr01-wf15-format-20261026-02`, exit 2,
`harness-runtime-build/build-harness-runtime/step-summary.json`); corrected before
execution. Final format `tr01-wf15-format-20261026-04` passed (5450 ms).

F-018/F-019 and routing portion of F-020 are implemented. No retired selector or
old API reference remains in executable sources or routing. Human authority and
Markdown isolation are unchanged. This checkpoint closes WF-15 before WF-16
starts. Next: agent-finalize, inspect generated changes, freeze executable source,
run A–E successor union plus required static checks, audit every successful root,
and finish the handoff. No compatible successful full warm-check root exists;
agent-finalize will leave RESULTS_DIR unset and report retained maintenance skipped.

### WF-16 exit / E-041 — final validation and handoff complete

WF-12–WF-16, T-009–T-013, F-015–F-020 and C-011–C-016 are complete for this
bounded PostgreSQL effort. The following evidence is the final acceptance
portfolio; earlier runs retain their own identities and are not substituted.

Finalization ran before broader verification. The initial final Go lint found
`U1000` on the now-unused testservices template-detail helper and the nonzero-sized
capability identity marker (`tr01-wf16-lint-go-20261026-01`, outer exit 2;
`unit-logs/operation-harness.lint_go_staticcheck/stdout.log`). Its normalized result
was `product/test_assertion_failure`; these were implementation cleanup errors,
not a service failure. The helper was removed and the nonzero-sized marker made
unnamed. `make agent-finalize` was repeated, again reporting generated outputs
unchanged (zero updated files), then executable source was frozen. Superseded
WF-16 `-01` lifecycle/migration/static results remain retained but are excluded
from the final table below. No unresolved failure remains.

Final executable/input identity:
`sha256:fbc666088d1ff36c404fa5bc33c090ffab488aa0726f0002e88897436d3a2f97`.
Every graph root below has the same toolchain identity
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and system identity
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Graph/workload digests intentionally differ by exact selection and remain in each
`run-manifest.json` and target summary. Authored configuration, capabilities,
isolation scopes and qualification profiles were frozen with source; no profiles
were replaced or mixed to claim coverage. HEAD remains
`867e97ed49829d671c58576232999d1fc9a4e3fb`; worktree source is intentionally dirty.

Each command used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<ID>`. Each exact run root
is `/home/jochi/code/cartulary/.cartulary/test-results/<ID>`:

| Run ID | Public command / exact §8 selection | Duration ms | Native result / outer exit |
| --- | --- | --- | --- |
| `tr01-wf16-finalize-20261026-02` | `make agent-finalize` (RESULTS_DIR unset) | 23730 | pass / 0 |
| `tr01-wf16-lifecycle-20261026-02` | A plus first E command; exact union below | 54585 | pass / 0 |
| `tr01-wf16-migrations-20261026-02` | B successor, unchanged row ID | 21641 | pass / 0 |
| `tr01-wf16-postgres-20261026-01` | C successor, unchanged row IDs | 20945 | pass / 0 |
| `tr01-wf16-process-20261026-01` | D | 75496 | pass / 0 |
| `tr01-wf16-transport-20261026-01` | Second E command: explicit transport qualification | 91987 | pass / 0 |
| `tr01-wf16-recovery-20261026-01` | Third E command: packaged recovery | 80833 | pass / 0 |
| `tr01-wf16-lint-go-20261026-02` | `make lint CHECK=go` | 15909 | pass / 0 |
| `tr01-wf16-lint-boundary-20261026-02` | `make lint CHECK=backend-boundary` | 2677 | pass / 0 |
| `tr01-wf16-source-generated-20261026-02` | `make source-check CHECK=generated` | 15904 | pass / 0 |
| `tr01-wf16-source-schema-20261026-02` | `make source-check CHECK=schema` | 18294 | pass / 0 |
| `tr01-wf16-source-catalog-20261026-02` | `make source-check CHECK=catalog` | 1744 | pass / 0 |
| `tr01-wf16-source-identity-20261026-02` | `make source-check CHECK=identity` | 690 | pass / 0 |

The lifecycle union command was exactly:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.pgtest_helper_contract,harness.browser.unit.testservices_lifecycle_contract,harness.browser.integration.postgres_template_preparation,harness.browser.integration.postgres_transaction_isolation,harness.browser.integration.postgres_owned_barrier_observation,harness.browser.integration.postgres_single_drop,harness.browser.integration.postgre_sql_cleanup_coordinates_real_database_dr_8a5f7f8277,harness.browser.integration.reset_checkpoint_characterization,harness.browser.integration.managed_service_assertions_for_runtime_reset_rec_79940f98ab,harness.browser.integration.performance_fixture_snapshot_lifecycle
```

All 13 graph roots passed `make harness-observability-check
RESULTS_DIR=<exact-run-root> CHECK=all`, outer exit 0, `full_validity=1`.
Every retained cleanup outcome is `completed`; every retained-secret scan passed.
The transport root contains exactly 64 lifetime scopes and preserves the
16 serial / 32 four-way concurrent / 16 competing lifetimes. Recovery qualified
the packaged process boundary independently of reset and snapshot lifecycle.

`make lint-markdown` also passed, root
`.cartulary/test-results/tr01-wf16-markdown-20261026-01`, outer exit 0, 8000 ms;
summary `adhoc/lint-markdown/tool-run-summary.json`. Its legacy adhoc artifact is
separate from graph validity. Current lint globs omit this tracker: the changed
tracker was reviewed directly for structure, links, stable IDs, successor mapping,
phase statuses, source identities, scope and truthful limitations; `git diff
--check` passed. The specification diff only replaces misleading pool terminology
with quarantine/readmission wording. Specification/projection fidelity remains
human-owned; no executable check reads or certifies Markdown requirements.

#### Completed interfaces and assertion handoff

- Physical construction is only `StartOwned(ctx, StartOptions) (*OwnedCluster,
  error)`. Acquisition/settlement callbacks receive the owner. Retain a non-nil
  partial owner on failure for its cleanup receipt and recovery authority.
- Owned template preparation is explicit: `owner.PrepareTemplate(ctx, name)`;
  empty name uses existing canonical naming, service composition supplies its
  existing identity. `owner.Fixture()` becomes available after physical admission.
  Ordinary clones require the sealed template; scratch operations remain explicit.
- `Harness.Close`, physical fields, owned `StartOwnedWithLabels` and
  `StartOwnedWithOptions`, the interim `Harness.PrepareTemplate`, and PostgreSQL
  `StopShared` are removed without aliases. Borrowers use attachment-only
  `Start`/`StartShared`; `DetachShared()` has no context, I/O or shutdown error.
  Detach preserves admitted identity, ownership and generation quarantine.
- `TestDatabaseFixtureCleanup` succeeds the four standalone/attached clone/scratch
  cleanup tests; the prior matrix qualified before removal. The migrated-database
  witness succeeds the weak startup smoke. Capability and policy tests retain
  their row IDs and now test the actual paths. The assertion ledger in §7 and
  E-040 give the detailed mapping. Distinct failure/context witnesses remain.
- Direct callers and authored/generated routing are migrated atomically. No
  product API, schema/data migration, new policy language, new package hierarchy,
  S3 redesign, browser workflow change or performance-policy change is required.

#### Preservation, limits and operational handoff

The execution inventory comparison found 21 changed existing files and 12 new
pgtest files, all inside the authorized boundary; no baseline file was deleted.
The pre-existing tracker history inside `<details>` remains byte-identical.
Seventy generated-policy files had metadata/inode refreshes from public generation
with byte-identical contents; there were no metadata-only changes outside that
policy. Index entries remain byte-identical (digest
`2f8393a2bede4eadc41b7ccc1246f9daa19005bae7c33cc7c84b15d4b9155ce4`);
raw index stat-cache metadata refreshed, with no staging or commits. The final
7,338-file executable/input freeze was rechecked unchanged after validation.
Ignored build/results/dependency trees were not exhaustively fingerprinted.

Both finalizer invocations left RESULTS_DIR unset: no compatible successful full
warm-check root exists. Retained-run maintenance and warm-check scheduler/timing
checks were therefore skipped (`results-dir-not-provided` in
`unit-artifacts/finalize-summary.json`). Successful scoped graph roots were audited
explicitly as above. Full-owner, CI/release and performance acceptance are not
claimed; visual/browser workflows and preceding-handoff WF-33 remain outside scope.
These are bounded acceptance limits, not incomplete PostgreSQL remediation.

The final handoff is this tracker plus the working-tree diff and exact retained
roots. No further implementation workstream remains. Human review can inspect
projection fidelity and the bounded specification wording without treating tests
as specification authority. Rollback remains workstream-sized as described in §7;
preserve the original tracker edits and all failed evidence.

### Earlier planning-session scope and authority (superseded by E-036–E-041)

| Session / baseline | Inspected / touched | Result and supersession | Next action |
| --- | --- | --- | --- |
| 2026-10-26; `867e97ed49829d671c58576232999d1fc9a4e3fb` | Root/internal/tools instructions, tracker framework and adopted owners; sole edit this handoff | E-030/E-032/E-034: completed frontend iteration becomes historical; PostgreSQL plan is current; no owner amendment or implementation authorized | Future WF-12 starts with source and owner revalidation |

### Backend boundary

| Session / baseline | Inspected boundaries | Decision / evidence | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | pgtest, testservices and direct physical/process callers | E-031: split physical lifetime from fixture access; one template recipe; narrow helper owners retained | Qualify pre-move assertions, then WF-13 before WF-14 |

### Frontend boundary

Not applicable to the new iteration: no frontend change. The completed frontend
work and qualification limits are retained in the historical snapshot below.

### Contract and codegen

| Session / baseline | Inspected boundaries | Decision / evidence | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | Core/Harness owners, family rows, generated policy/topology | E-032/E-033: authored routing first, Make projections only; method count, wrappers and duplicate recipes are not requirements | Future WF-15 records exact old-to-successor assertion/selector mapping |

### Tests and harness

| Session / baseline | Inspection / check | Result / evidence | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | Assertions and routing inspected; no tests run | E-031/E-033; RB-005 remains open; historical passes do not clear it | Run V-015 only in a later implementation task |
| Document save | V-020 | E-035; lint scope and manual review limits explicitly recorded | No product verification for this update |

### Security and authorization

| Session / baseline | Inspected requirements | Decision / evidence | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | TLS/purpose admission, migration capability, owned cleanup and redaction boundaries | E-031/E-032: preserve C-011/C-013/C-014/C-015; no public injection or arbitrary-handle migration path | Keep real transport/ACL and capability-negative witnesses in future portfolio |

### Open risks and next session

| Session / baseline | Risk / blocker | Current disposition | Next action |
| --- | --- | --- | --- |
| Earlier planning baseline | RB-005/RB-006: no current execution evidence and predecessor gates | Planning complete for bounded seam; movement blocked, future tests explicitly specified | S-012 may enter after authorization; report failure causes before moving or broadening |
| Document save | Concurrent changes and command effects | E-034 pre-write identity check; E-035 final index/content comparison | Revalidate affected sources on any later resumption |

### Current evidence register

| ID | Source / command | Baseline / environment | Observation / artifact / applicability |
| --- | --- | --- | --- |
| E-030 | `git --no-optional-locks` status/index/tree inspection; external content/identity snapshot | Current `main` commit above; clean start | `/tmp/cartulary-pgfixture-plan-bf09ql7x/baseline.json`, `prior-handoff.md`; 7,637 entries; retained fingerprints support preservation, not behavioral coverage |
| E-031 | Source/assertion/import inspections in Section 2 | Same source commit; read-only | Mixed ownership, duplicate recipes, mutable hooks, resolver duplication and incidental assertions directly observed; 116 importer search is not a full product-caller audit |
| E-032 | Adopted status/Core precedence, Harness REQ-804/810/811/405/410 and §11.0; Core 01/04/05; repository instructions | Same source commit | C-011–C-016 distinguish required assurance from private test design; no primary-owner contradiction found in this seam |
| E-033 | Authored family/verification rows, Make surface, policy/topology, test-ops command recipes and lint path | Same source commit; no future command executed | Exact future A–E portfolio; generated outputs and invocation effects identified; V-020 lint exclusion known before execution |
| E-034 | Immediate pre-write canonical path/link/content recheck | Same commit; prior tracker digest from Section 1 | Matched E-030 across all 7,637 entries and both index fingerprints before writing; output remained regular/singly linked; `prewrite-comparison.json` retained beside baseline; historical snapshot retained with heading demotion |
| E-035 | V-020 and final preservation review | Documentation session, 2026-10-26; external cache/results | Lint passed at 20:33:49 UTC, exit 0, 7.570 seconds; no tests. Direct document review and `git --no-optional-locks diff --check` passed. Final comparison found only this tracker changed, with index and all other enumerated content/identities preserved; details below |
| E-042 | Post-implementation source/owner review; read-only final-manifest/summary/cleanup/scan inspection and source-freeze comparison | Review baseline and limits above; same E-041 executable source | Five required gaps and optional redundant attachment state; no product rerun or repair; follow-up chain WF-17–WF-19 |
| E-043 | V-024 documentation validation and preservation | Tracker-only review; external results/cache | Lint tool summary pass/0; outer Make exit not retained; direct review/diff/preservation passed; tracker-only change confirmed |
| E-044 | Closure implementation entry and preservation baseline | Authorized F-021–F-026; WF-17–WF-19; entry backup | Required decisions reconciled before executable edits |
| E-045 | WF-17 preparation recovery checkpoint | V-021 and exact roots above | F-021/F-024 qualified; terminal preparation and early ownership; predecessor saved |
| E-046 | WF-18 admission/binding/routing checkpoint | V-022 and exact caller/static roots above | F-022/F-023/F-025/F-026 qualified; predecessor saved; failed evidence preserved |
| E-047 | WF-19 final validation and handoff | V-023; 22 compatible roots; all native/audit exits 0 | 42 selected rows, 64 fresh transport lifetimes, completed cleanup, preserved source/history and bounded acceptance COMPLETE |

### Documentation save validation — E-035

The executed command from the repository root was:

```sh
env GIT_OPTIONAL_LOCKS=0 \
  CARTULARY_PREPARATION_POLICY=installed_only \
  CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-pgfixture-plan-bf09ql7x/results \
  CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-pgfixture-plan-bf09ql7x/static-analysis \
  CARTULARY_TEST_RUN_ID=tr01-postgres-plan-doc-20261026-01 \
  COREPACK_ENABLE_NETWORK=0 make lint-markdown
```

It passed with summary at
`/tmp/cartulary-pgfixture-plan-bf09ql7x/results/tr01-postgres-plan-doc-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The wrapper labels this documentation invocation `adhoc`; its summary records
zero tests. The pass covers the configured Markdown globs, which exclude
`docs/handoffs/testing-remediation-01.md`. Direct review of this document confirmed
the twelve ordered sections, balanced fences and consistent tables, seven local
links, 21 exact catalog row references, preserved historical text apart from
heading demotion, and coherent finding/contract/slice/validation references.
`document-review.json` records the structural/reference inspection; it is
documentation review, not executable harness or conformance evidence.

`final-comparison.json` beside the baseline records unchanged HEAD/branch,
index bytes and entries, path population, and content/file identities for every
enumerated entry outside this tracker. The tracker retained its file identity.
Ignored dependency/build trees were not exhaustively fingerprinted; the reviewed
invocation used existing dependencies, fixing disabled and external result/cache
paths. No install, format, generation, `agent-finalize`, product check, service
startup or test was run, as required for this document-only update. Final evidence
annotations were reviewed with the same document/baseline checks; they do not
change the linted file population.

### Completed iterations — historical snapshot

The following snapshot retains the preceding document's decisions, stable IDs,
commands, artifacts, failed-attempt diagnoses and completion limits. Heading
levels are demoted for nesting. Statements such as “current”, “authorized” and
“DONE”, and numbered section references, refer to that historical snapshot only.
It is evidence history, not authorization or a validation result for WF-12–WF-16.

<details>
<summary>Completed frontend and graph/scanner iterations (WF-00–WF-11)</summary>

### Testing Remediation 01: Frontend Test-Policy Simplification

#### 1. Scope and Source Posture

Document date: 2026-10-26. **Current action: completed implementation and handoff.** The user
approved the revised WF-06–WF-11 plan and instructed its implementation after
selecting semantic readiness, simplified E2E owner rules, and retention of the
existing bounded Fallow gate. This supersedes the earlier document-only scope.
WF-06–WF-11 are DONE, including final validation and handoff. Historical WF-00–WF-05 remain complete with their original
qualification limits in Section 10. The preceding
[remediation handoff](testing-remediation.md) and its separate open WF-33 effort
are unchanged.

Repository: `/home/jochi/code/cartulary`; supplied and established canonical target
`.`; kind: repository directory; label: `frontend-test-policy-assurance`. This
continues the existing tracker's repository identity with a newly bounded seam.
The controlling plan and handoff is `docs/handoffs/testing-remediation-01.md`.
Authorized edits also include the specification, implementation, tests, authored
routing and Make-generated projections listed below. The output is a regular, singly linked file; neither it
nor its parent components is symlinked. Artifact mode: **normal**. Planning
completeness, workstream completion and implementation authorization are separate
outcomes; see Section 12.

The [modular refactor framework](cartulary_modular_refactor_planning_framework.md)
and Refactor Tracker format supply planning structure. They do not make current
test arrangements requirements. This iteration covers the two bespoke frontend
policy scanners, their useful successor assurance, E2E transport ownership,
Fallow E2E roots, and directly affected routing and support guidance. It does not
redesign production behavior, browser workflows, service fixtures, scheduling,
runner protocols, security thresholds, visual goldens or performance policy.

| Authority | Applicable boundary and disposition |
| --- | --- |
| [Core 00](../spec/00_document_set_status_and_precedence.md), §§2–5 | Adopted primary-owner precedence and the documentation-free executable boundary. Tests, typed projections and this handoff cannot define product requirements. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md), adopted/current | REQ-001/003/004: scope and projections; REQ-070: existing Fallow policy; REQ-075/082: discovery and canonical assurance; REQ-103: semantic frontend readiness; REQ-665: stable post-mutation identity and observation-only continuity; REQ-676/677: E2E discovery/reachability and private raw transport; Sections 3, 8–10: exact selection, evidence and execution. |
| Core 01–04 and adopted product owners | Production HTTP, authorization, storage, history, collaboration and UI behavior remain outside the change. Core 03 REQ-03-283 remains the product owner for row-local focus restoration. |
| [Domain vocabulary](../domain.md) and [frontend testing guide](../guides/cartulary_frontend_implementation_testing_guide.md), §7 | Owner navigation and implementation-support boundaries; the guide is not a second requirements authority. |
| [Core 05](../spec/05_claim_publication_and_benchmark_reproducibility.md) | Not applicable: no claim-bearing timed or fixture-sensitive publication is proposed. |

Executable tests, generators, runtime metadata, conformance and release evidence
must never read, stat, hash or otherwise depend on this document or other Markdown.
Human inspection and documentation lint are not product evidence consumers.

The following table preserves the **prior document-only baseline** (E-010–E-017).
The implementation baseline, including the pre-existing tracker changes, is
E-018; implementation preservation and final evidence are recorded in E-029.

| Historical baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Checkout | `main`; HEAD `971d6423c1bfc5f1f3ac6569fc524b596cf33f61` | E-017 records the final comparison | No commit, push or PR requested |
| Index | Clean; entry fingerprint `005bc70a5d18167d1c962ef3029fb6530fab6a16d8746efb8e367b54ddf104f2` | Must remain identical | Optional Git writes disabled |
| Tracked/non-ignored content | 7,638 paths; combined fingerprint `7ed6019c9a137bac0f549c20351d540cf11fbac8c0775d09e82945df30d91ad3`; no unstaged or non-ignored untracked changes | Excluding this output, fingerprint must remain `a7df1bc81868dab69454c7e23a14d67e0ced15e124fc1cab6dfb76033db16215` | Per-path hashes and file identities retained outside repository |
| Output | Inode `348826`, one link; initial SHA-256 `70dc2e851a16e6c07ed52e483ccbc853b5cb15dc9589396fe271bddcca3058b0` | Identity and contents rechecked immediately before replacement | Historical canonical target `.` established from existing text |
| Comparison material | `/tmp/cartulary-tr01-doc-hla6fi0g/baseline.json` and `prior-handoff.md` | E-017 | Temporary local resumption aids, not executable evidence inputs |
| Ignored state | Installed dependencies, caches, builds and retained results excluded from refactor inputs | Documentation validation must use installed dependencies and external temporary outputs, or be skipped with its reason | Ignored trees are not exhaustively content-audited; no broad preservation claim for concurrent external activity |

Those document-only fingerprints matched the preceding planning turn. They do
not qualify implementation results. The subsequent explicit implementation
instruction authorizes WF-06–WF-11; E-018 preserves its distinct starting state.

#### 2. Inspected Baseline Repository Inventory

The enumeration universe is tracked and non-ignored untracked paths under `.`.
Every path is metadata-only for this bounded analysis unless explicitly identified
below as directly inspected. Generated, binary and vendored entries remain
metadata-only; ignored dependency installs, service data and result directories
are excluded because they are not authored refactor inputs. No submodule or
external symlink target is traversed as source. Enumeration is not behavioral
inspection. Exploratory Go fixture and remaining shell-suite reads did not select
those subsystems for change.

| Path / kind | Disposition and inspection depth | Responsibility, consumers and dependencies | Owner / contracts / risk / evidence |
| --- | --- | --- | --- |
| `apps/web/e2e/support/runtime/architecturePolicy.test.ts` | Directly inspected: complete 655-line authored scanner and nine assertion groups | Own source walker, TypeScript import/reachability/export analyzer, literal and naming restrictions; one `harness.browser` row, project `harness-node` | Split across existing owners; C-007/C-009/C-010; loss of unique checks if simply deleted; E-012 |
| `apps/web/src/testing/selectorContractPolicy.node.test.ts` | Directly inspected: complete 671-line authored scanner and five assertion groups | Source walker, selector-prefix inventory, mock exceptions, regex and AST scans, prescribed heading grammar; five `web.architecture` rows, project `node-unit` | UI Contracts and behavioral support; C-008/C-010; REQ-103 amendment needed before retiring grammar; E-013 |
| `tools/harness/static-analysis/frontend-import-boundary-check.mjs` and `tests/test-frontend-policy.mjs` | Directly inspected at parser, resolver, evaluator, policy fixture and result boundaries | Existing explicit-root checker with one source enumeration and TypeScript parsing; real CLI and canonical assurance consumers | Frontend static analysis; C-007/C-009; retain independent positive/negative expectations; E-012 |
| `tools/frontend_import_boundaries.json` | Directly inspected relevant authored rules and scan scopes | Already checks grid/vendor, generated artifacts, package facades and E2E/private-app imports; scanner-specific core-support exception | Authored projection, not requirements authority; C-007/C-010; E-012/E-015 |
| `.fallowrc.json`, `tools/fallow/reachability_owner.json` | Directly inspected configuration and relevant owner fields | Blanket `apps/web/e2e/**/*.ts` root hides helper orphaning; existing Vitest entry patterns omit E2E support tests | Fallow; C-007; root correction precedes analyzer deletion; E-012 |
| `tools/harness/static-analysis/fallow-reachability.mjs`, `tests/fallow-fixtures.mjs`, `tests/test-fallow-integration.mjs` | Directly inspected root construction and real analyzer fixtures/assertions | Effective config merges authored roots; fixtures cover live/dead peers, but not the required E2E root distinction | Fallow; C-007/C-010; pinned analyzer qualification is future work; E-012 |
| `apps/web/vite.config.ts`, `apps/web/playwright.config.ts`, `apps/web/playwright.shared.config.ts` | Directly inspected test discovery and setup bindings | `harness-node` owns E2E `*.test.ts`; Playwright owns specs and explicit global setup/teardown; no config behavior change planned | Existing runtime-discovery owners; C-007/C-010; E-012 |
| `apps/web/e2e/support/transport/publicJsonClient.ts`, `publicHttpOperationClient.ts`, `publicJsonClient.test.ts` | Directly inspected complete execution and test boundaries; import search across E2E | Origin/type consumers span E2E; `requestPublicJson` is used only by direct transport tests; the operation client uses `requestPublicJsonObserved` | Transport; C-009; guard names the wrong seam; E-014 |
| `apps/web/src/testing/timelineWorkbookTestSupport.ts`, `timelineWorkbookTestSupport.test.tsx`, `workbookInspectorTestSupport.ts`, `workbookInspectorTestSupport.test.tsx` | Directly inspected row/inspector waits and existing assertions | Finite configured waits; stable row IDs and exact inspector view/record/version; existing delayed hydration and diagnostic tests | Application test support; C-008; add negative identity and observation witnesses; E-014 |
| `packages/ui-contracts/src/selector-core.node.test.ts`, `application-selectors.node.test.ts` | Directly inspected selector-identity, escaping, registry and authentication assertions | Public package facade tests already cover deterministic stable identities and malformed inputs | `package.ui`; C-008; preserve existing runtime selector output; E-014 |
| `tools/test_families/{web.architecture,harness.browser,package.ui}.json`, `contracts/verification/owners/web.architecture.json`, `tools/frontend_source_ownership.json` | Directly inspected affected rows, verification kinds and path membership | Six scanner rows; existing frontend/Fallow operations; readiness, transport and selector successors; authored source ownership | Routing owners; C-010; atomic source/selector/projection migration; E-015 |
| Makefile, generated task-surface includes, generated-artifact policy, documentation lint and readiness wrappers | Directly inspected relevant invocation paths and mutation boundaries | Canonical commands, installed-only admission, external result/cache outputs, generated ownership | Harness plumbing; C-010; documentation-only invocation boundary; E-015/E-016 |
| Remaining enumerated paths | Metadata-only; not proposed for movement | Not a claim of assertion completeness or exhaustive callers | Product/backend/browser workflow audit deferred outside this seam |

The primary inventory is two scanners, fourteen top-level assertion groups and six
catalog rows. These counts identify the inspected baseline; they are not future
acceptance constraints. Supporting files are inspected only to the depth stated.
Direct import and routing searches establish the known consumers, not the absence
of dynamic external consumers. E-018/E-019 revalidated this baseline. WF-10 also
inspected and updated the graph assurance tier sentinel discovered during cutover.

#### 3. Module Boundary Diagnosis

| Responsibility found | Baseline location | Required owner / disposition | Evidence / finding / contract |
| --- | --- | --- | --- |
| Facades, private imports and vendor boundaries | E2E scanner plus existing frontend checker | **Keep** existing checker; **move** useful residual checks into its existing source pass | E-012; F-007; C-007 |
| Reachability and unused exports/types | Bespoke E2E analyzer plus Fallow | **Move** E2E assurance to Fallow after root and real-analyzer qualification | E-012; F-007/F-008; C-007 |
| Selector identity and encoding | Selector scanner plus UI Contracts tests | **Keep** package behavioral assertions; retire duplicate syntax/family inventories | E-013/E-014; F-009; C-008 |
| Bounded readiness and observation-only continuity | Source-expression grammar plus actual helpers | **Keep/extend** helper behavioral tests; separately amend incidental grammar | E-013/E-014; F-009; C-008 |
| Raw versus validated HTTP execution | Symbol-specific scanner guard and mixed transport file | **Split** raw execution from the reusable context/origin boundary; enforce actual module ownership | E-014; F-010; C-009 |
| Catalog ownership and execution | Browser/architecture scanner rows and existing assurance operations | **Move/consolidate** obligations into existing owners; retire six redundant rows when qualified | E-015; F-011; C-010 |
| Production HTTP, WebSocket, storage, revision and UI semantics | Existing production owners | **Keep**; not proposed for movement or correction | C-006; no production slice |

No new public production API, runner, schema, registry, scheduler, compatibility
mode or general fixture language is needed. Private support exports can be
removed or narrowed; forwarding wrappers are not a completion strategy.

#### 4. Public Contract and Behavior Freeze Map

Historical C-001–C-005 remain reserved for the completed graph/scanner iteration.

| ID / surface | Observed behavior and owner | Required disposition / successor | Prerequisite / risk / slice / validation |
| --- | --- | --- | --- |
| C-006 — production behavior | Core 01–04 and applicable adopted subsystem owners retain HTTP, authorization, storage/history, collaboration and UI requirements; this iteration changes support only | Preserve routes, wire shapes, authentication, persistence, revision semantics, DOM identity and actual focus/accessibility behavior | No product correction or product-data migration; all slices must remain within support scope |
| C-007 — architectural integrity | Existing checker enforces meaningful import boundaries; E2E scanner adds its own graph and source-shape rules; REQ-070/082 govern analyzer and consolidated assurance mechanics | Preserve facade/vendor separation, route-neutral shared helpers and meaningful reachability; retire closed-root/catch-all naming and parser-specific restrictions | F-007/F-008; S-007/S-008/S-010; V-009/V-010/V-012 |
| C-008 — selectors/readiness | Amended REQ-103 requires finite waits, stable identity and observation-only readiness; REQ-665 requires committed identity/version convergence | Preserve semantic guarantees and runtime selector values; the authorized amendment removes matcher grammar before scanner retirement | RB-002; F-009; S-006/S-009; V-008/V-011 |
| C-009 — support HTTP transport | Origin binding, body/header omission, non-2xx results, generated method/path/query, success validation, observed headers and private diagnostics; guide §7 and Harness evidence integrity apply | Preserve all meaningful operation/transport assertions; replace ineffective wrapper-name guard with module ownership | RB-001; F-010; S-007; V-009 |
| C-010 — execution/evidence | REQ-003/004/075/082, Section 3 and Sections 8–10 govern authored routing, exact selectors, canonical operations and truthful publication | Preserve exact resolution and distinct contexts; remove obsolete rows with explicit successor/retirement mapping; no historical-ID aliases | F-011; S-010/S-011; V-012/V-013 |

Source literals, helper filenames, test titles, directory placement, exact test
populations and the old scanners' parser limitations are not production
requirements. Actual requirements and necessary safety/evidence invariants need
successor assurance. Do not invent normative text to preserve incidental test
arrangements. Test presence is not execution success, and no result proves owner
projection fidelity without human review.

#### 5. Coupling and Boundary Findings

Findings describe the inspected baseline. Their completed remediation and
validation are recorded in E-019–E-029; the risk column below describes what
retaining that baseline would have left unresolved.

| ID / classification | Finding and evidence | Required action / scoped acceptance | Contract / slice / validation |
| --- | --- | --- | --- |
| F-007 / must_fix | E-012: the E2E test duplicates graph traversal, import/export interpretation and existing boundary enforcement | One established owner per meaningful rule; no second analyzer or new generic policy engine | C-007; S-007/S-008/S-010; V-009/V-010/V-012 |
| F-008 / must_fix | E-012: every E2E TypeScript file is currently a Fallow root, so existing Fallow coverage is not equivalent to orphan detection | Admit true test/setup roots, prove imported helpers live and dead peers dead using the real pinned analyzer before deletion | C-007; S-008; V-010 |
| F-009 / must_fix | E-013/E-014: selector prefixes, mock allowlists and heading-expression grammar duplicate identity knowledge and constrain legitimate assertions | Adopt semantic REQ-103 replacement; qualify readiness failures and unchanged focus directly; remove syntax inventories | C-008; S-006/S-009; V-008/V-011 |
| F-010 / must_fix | E-014: name guard bans test-only `requestPublicJson`, not actual raw execution through `requestPublicJsonObserved` | Establish a transport-private execution module and one validator path; remove the unused wrapper without a shim | C-009; S-007; V-009 |
| F-011 / must_fix | E-015: six scanner rows and a scanner-specific boundary exception preserve redundant work and split ownership | Migrate authored routing, source ownership and generated projections atomically; preserve successor qualification contexts | C-010; S-010/S-011; V-012/V-013 |
| F-012 / intentional/no_action | Existing real checker/analyzer, selector-package, transport and helper tests observe different boundaries | Keep those distinct witnesses; matching topic names do not make them duplicate coverage | C-007/C-008/C-009; no separate removal slice; V-009–V-011 |
| F-013 / must_fix | Revised planning found that REQ-676 owns the closed E2E layout/importer accounting and REQ-677 names the wrapper slated for removal | Amend these owners before dependent implementation; retain semantic discovery and validated transport authority | C-007/C-009; S-006; V-008 |
| F-014 / must_fix | REQ-070/help describe advisory execution ambiguously although the existing implementation blocks bounded web/package findings | Specify the existing bounded gate, correct authored help, and qualify E2E finding normalization without expanding local checks | C-007/C-010; S-006/S-008/S-010; V-010/V-012 |

`must_fix` applies only to these slices and their completion conditions. No
production defect or unrelated cleanup is authorized by a finding. Lower test
counts alone are not evidence of improvement; the result must remove machinery
and duplicated decisions while retaining meaningful assurance.

##### Decision, migration and unresolved-risk ledger

The action and validation columns above define each fix. This ledger records why
that disposition is preferable over preserving the baseline, its affected areas,
and its long-term implications.

| Gap | Areas changed | Rationale and long-term benefit | Compatibility / migration impact | Risk if left unresolved |
| --- | --- | --- | --- | --- |
| F-007 | Implementation, tests, routing, guidance | Consolidate useful residual checks into one parsed source pass and established import policy; eliminate duplicate traversal and parser semantics | Retire the E2E scanner after successor qualification; no replacement analyzer or rule language | Divergent enforcement, false positives and multiple maintenance sites for every syntax change |
| F-008 | Analyzer configuration, tests, specification | True runtime roots expose dead support files; real-analyzer positive and negative controls establish capabilities instead of assuming parity | Specs/setup/support tests remain roots; imported helpers remain ordinary dependencies; dead code becomes visible to the existing bounded gate | Blanket entries silently hide orphans and make removal of the old scanner unsafe |
| F-009 | Specification, tests, routing, documentation | Stable identity, finite deadlines and observation-only behavior directly test the needed postconditions; removing source grammar permits future legitimate assertion styles | Remove five scanner rows and private syntax/fixture exceptions; preserve selector output and production behavior | Brittle parser restrictions grow while wrong-record, wrong-version or action-retrying helpers can still pass |
| F-010 | Specification, implementation, tests, import policy | Restrict the actual execution module and share one same-origin validator; callers have a clear validated-operation boundary | Direct private imports migrate; remove the unused result wrapper without alias; no public API or data migration | The ineffective wrapper-name ban permits bypass of operation-derived method/path/status/schema validation |
| F-011 | Authored routing, generated projections, graph/source ownership tests, documentation | One explicit successor map and atomic cutover keep discovery, tiers and evidence coherent | Six obsolete row IDs disappear; no historical aliases; standard readiness/full analyzer placement retained | Dangling selectors, duplicated execution, orphan exceptions and misleading coverage claims |
| F-012 | Retained tests and evidence accounting | Keep checks that observe different boundaries, including real repository analysis and isolated fixtures; this retains independent failure detection | No compatibility change; each retained context has a distinct qualification purpose | Over-consolidation could discard runtime behavior, integration or failure-normalization evidence |
| F-013 | Specification and downstream projections | Amend actual adopted owners before implementation, removing accidental filename/importer rules and naming the real transport boundary | Owner text changes with its private projections; canonical discovery and identity remain | Implementation would conflict with adopted requirements or keep obsolete behavior solely to satisfy a scanner |
| F-014 | Specification, authored/generated help, tests | Describe the existing bounded blocking policy accurately and test normalized findings despite analyzer exit 0 | No new gate or local broad scan; advisory findings outside current scopes remain advisory | Operators misread a failing analysis as advisory, and future work may weaken or accidentally broaden enforcement |

#### 6. Refactor Workstreams

The prerequisite column is authoritative. All workstreams are included and
sequential; none conditionally skips a predecessor. Each handoff checkpoint
records source identity, assertion dispositions, commands/exits/artifacts,
remaining gates and next action before the successor begins.

| ID / slice | Class / authoritative prerequisite | Goal and likely changes | Validation / handoff checkpoint / binary exit |
| --- | --- | --- | --- |
| WF-06 / S-006 | root; baseline revalidation and explicit implementation authorization | Final assertion ledger, focused characterization and concrete owner review packet | V-008; record baseline and REQ-103 disposition; no ambiguity about preserved obligations or retired restrictions |
| WF-07 / S-007 | chain; WF-06 | Existing frontend checker, transport-private execution and direct tests | V-009; boundary/transport checkpoint; actual raw execution is confined and transport assertions pass |
| WF-08 / S-008 | chain; WF-07 | Authored Fallow roots and existing real-analyzer fixtures | V-010; analyzer checkpoint; legitimate E2E roots/helpers and dead peers are distinguished |
| WF-09 / S-009 | chain; WF-08 plus adoption of the concrete REQ-103 amendment | Readiness witnesses and selector scanner/row retirement | V-011; behavioral checkpoint; semantic readiness and observation guarantees qualify without syntax policing |
| WF-10 / S-010 | chain; WF-09 | E2E scanner/row retirement, remaining routing, source ownership and guidance cleanup | V-012; cutover checkpoint; successors account for every obligation and no obsolete binding remains |
| WF-11 / S-011 | chain; WF-10 | Finalization, compatible final portfolio and completion handoff | V-013; final checkpoint; Section 12 implementation criteria satisfied with explicit exclusions |

Phase risks follow the same dependency chain: WF-06 can omit a normative owner
(mitigated by explicit REQ-103/676/677/070 review); WF-07 can expose raw execution
or duplicate validation (module allow/deny and transport witnesses); WF-08 can
mistake analyzer success for coverage (independent live/dead and normalized-failure
controls); WF-09 can lose readiness semantics (negative identity, timeout and
no-action witnesses); WF-10 can leave routing drift (generation, catalog,
identity, source ownership and tier assurance); WF-11 can accept incompatible or
incomplete evidence (freeze, exact-root audits, cleanup/scan review and explicit
coverage limits). A failed exit blocks the next phase; rollback boundaries are
defined in Section 7.

#### 7. Authorized Refactor Slice Plan

Implementation of all six slices is authorized. Readiness describes prerequisite
evidence, not completion. Later slices inherit their workstream prerequisites
from Section 6; each checkpoint must be saved before the successor begins.

| Slice / kind | Findings / contracts | Entry evidence and readiness | Intended change / tests | Completion and rollback |
| --- | --- | --- | --- | --- |
| S-006 / owner proposal and characterization | F-007–F-011; C-006–C-010 | **DONE**, E-018/E-019; baseline passed and owner amendments adopted | V-008 baseline; adopt REQ-103/676/677/070 and AC-021 with support guidance | Record exact baseline and adoption status; roll back owners with dependent projections if the coordinated change is reverted |
| S-007 / structural simplification | F-007/F-010; C-007/C-009 | **DONE**, E-020/E-021; boundary and transport witnesses passed | Existing checker source pass and private transport split; V-009 | Positive/negative boundary controls and transport witnesses pass; rollback source/tests/authored binding changes together |
| S-008 / analyzer configuration correction | F-007/F-008; C-007/C-010 | **DONE**, E-022/E-023; real analyzer and repository scan passed | Narrow existing roots and extend real pinned-analyzer fixtures; V-010 | Orphan and unused-export controls fail correctly without suppressions; restore configuration and fixtures together on rollback |
| S-009 / adopted harness-policy change and test simplification | F-009/F-011; C-008/C-010 | **DONE**, E-024/E-025; semantic witnesses and selector retirement passed | Add negative readiness witnesses; remove selector scanner and five rows; V-011 | No grammar/selector-family inventory remains; meaningful behaviors qualify; rollback owner/source/routing/projections together |
| S-010 / structural cutover | F-007/F-011; C-007/C-010 | **DONE**, E-026/E-027; scanner retirement and routing qualified | Remove E2E scanner, row and exception; regenerate and reconcile inventories; V-012 | Every assertion group has a successor or justified retirement, with no orphan selectors or compatibility aliases |
| S-011 / final qualification | F-007–F-014; C-006–C-010 | **DONE**, E-028/E-029; compatible final qualification and handoff complete | Finalize, freeze source and execute V-013 once per required context | Compatible results and limits recorded; rollback is coordinated, slice-scoped and preserves unrelated work |

##### S-006 — requirement amendment and baseline

Use the assertion successor ledger below as the initial coverage map. Revalidate
its source evidence and execute V-008 before implementation movement. Add characterization only for missing consequential behavior; do not
freeze accidental source shapes or failures.

Adopted replacement for the named-heading paragraphs of Harness REQ-103:

> Readiness helpers MUST observe the semantic state needed by the next operation
> within the existing finite, configuration-backed deadline. When identity
> matters, the predicate MUST use the required stable surface, record and version
> identity rather than visible text or count alone. Accessible role/name queries
> MAY assert accessible content or actual focus. Enforcement MUST validate the
> helper's behavior and postcondition; it MUST NOT prescribe a particular matcher
> grammar or prohibit aliases merely because a source scanner cannot interpret
> them. Helpers MUST remain observation-only after the invoking action and MUST
> NOT move focus or repeat an action to manufacture a passing result.

Retain REQ-103's existing diagnostic fields, payload exclusion and product-failure
classification, and retain REQ-665's response/version convergence. Preserve exact
inspector-version matching separately from mutation convergence's accepted-version
floor; do not collapse these different postconditions. Update AC-021's references
to selector-policy assurance to name the corresponding behavioral assurance.
The user selected this concrete semantic amendment and authorized its adoption
with implementation. WF-06 applies it to the NLSpec and AC-021. Human projection
review remains distinct from passing tests. REQ-676 now retires the closed root
layout/custom importer rule in favor of qualified Fallow analysis; REQ-677 names
the actual private execution boundary. REQ-070 explicitly retains bounded web
and package enforcement, with advisory findings outside those scopes.

##### S-007 — one boundary implementation and a real transport seam

Reuse the existing frontend checker for package facades, private application
imports, generated-source ownership and grid-vendor imports. Add only two narrow
owner-local checks: grid-vendor selector leakage in E2E source and application
route/control-token leakage into shared test-utils. Consume the same enumerated
and parsed file records; do not add a second walker, parser, configurable syntax
rule language or copied rule registry. Preserve the narrow ownership restrictions
with allowed/forbidden examples; comments or fixture descriptions are not runtime
dependencies. Existing authored import rules remain independently tested.

Move raw JSON execution into `apps/web/e2e/support/transport/rawPublicJson.ts`.
Keep origin binding and request-context types in `publicJsonClient.ts` for actual
consumers. Maintain one same-origin path validator shared inside the transport
owner; do not duplicate validation or introduce a circular runtime dependency.
Remove the result-only `requestPublicJson` wrapper. The operation client uses the
raw module, and direct transport tests observe that same implementation.

Declare the raw module restricted through the existing path-prefix import policy;
allow its operation client and direct transport test only. Public context/origin
consumers do not gain raw execution through a re-export. Preserve omission versus
supplied body/headers, method/path/query generation, non-2xx payloads, success
status/schema validation, repeated-header availability after validated success,
and diagnostics that omit capabilities, payload values and private URLs. Update
private imports directly; no deprecated facade or wrapper remains.

##### S-008 — qualify E2E reachability through Fallow

Replace `.fallowrc.json`'s `apps/web/e2e/**/*.ts` entry with
`apps/web/e2e/**/*.spec.ts`, `apps/web/e2e/**/*.spec.tsx`, and the existing global
setup/teardown files. Add `apps/web/e2e/**/*.test.ts` to the current Vitest owner
entry patterns, matching `harness-node` discovery. `fixtures.ts`, support and page
helpers are ordinary imports, not unconditional roots. Use existing config/owner
shapes; no new dynamic-reachability family or schema is needed.

Extend the existing real-analyzer fixture with a live spec, setup/teardown,
support test, imported helper, orphan helper, unused exported value and type,
namespace import, re-export and literal dynamic import. Include a negative control
where the blanket E2E root would hide an orphan. Expected live/dead identities are
independent assertions; fixture configuration projects the actual authored E2E
root policy rather than an unrelated copied policy. Keep the real repository scan
as a separate configuration-integration witness.

Do not assume Fallow is equivalent before these witnesses run. If the pinned
analyzer cannot supply required assurance, block deletion of the affected bespoke
check and revise this plan. Do not introduce a replacement analyzer, suppression,
compatibility import rule or blanket helper entrypoint. Unexpected unrelated
findings require scoped diagnosis rather than silent cleanup or weaker policy.

##### S-009 — behavior instead of selector-expression policing

Preserve UI Contracts tests for deterministic stable IDs, escaping, collision
resistance, invalid identifier rejection and view-registry validation. Preserve
runtime selector outputs and existing production tests; removing a source ban
does not require rewriting their selectors.

Extend the existing row/inspector helper suites with controlled transitions that
prove same text on the wrong record, wrong surface, wrong exact inspector version
and a not-ready inspector cannot satisfy identity readiness. Prove delayed correct
identity succeeds, persistent mismatch times out with bounded payload-free
diagnostics, and observation does not change focus or dispatch an action. Retain
configuration-backed deadlines; do not increase them or introduce arbitrary
settling sleeps, retry wrappers or test-only production APIs.

After amendment adoption and qualification, delete the selector scanner and its
five rows. Remove prefix lists, blanket raw-literal prohibitions, fixture-name
exceptions and heading-expression grammar. The package facade and actual semantic
postconditions remain the assurance owners. Update source ownership and guidance
with the deletion, regenerating through Make as needed.

##### S-010 — retire the E2E scanner and complete routing

Delete the E2E scanner only after S-007/S-008 successors and S-009 behavior qualify.
Remove its catalog row and the scanner-specific core-support exception. Retire
closed E2E root placement, catch-all filename bans, mandatory external-importer
counting and parser-specific namespace/dynamic-expression restrictions. Dead-code
assurance uses Fallow's actual semantics; semantic catalog identities remain with
the existing identity owner.

Update affected authored family manifests, frontend source ownership, backing-file
declarations and support README/guide links. Reuse current canonical frontend and
Fallow assurance operations. Preserve standard placement for behavioral readiness
and full placement for analyzer assurance; consolidation must not promote broad
analyzer work into local gates. Regenerate downstream artifacts through Make and
remove obsolete active references without aliases. Human history may name retired
paths; executable consumers may not depend on it.

##### Assertion successor ledger

| Baseline assertion group | Classification / successor | Retirement condition |
| --- | --- | --- |
| E2E current-view-schema literal declarations | Incidental syntax restriction; UI Contracts registry validation and public facade usage remain | Delete the literal ban without adding a new schema-ID source scanner |
| E2E raw grid-vendor selectors | Useful architectural boundary; existing frontend source pass | Positive allowed owner and forbidden consumer witnesses qualify |
| E2E phase/catch-all helper filenames | Retire catch-all naming restrictions; retain canonical semantic identity rules where applicable | Do not move the old filename blacklist into another specification |
| E2E closed root and reachable support | Root layout is incidental; reachability moves to Fallow | True roots and live/dead helper controls qualify |
| Every support/page export needs an external importer | Replace custom importer counting with actual unused export/type assurance | Real Fallow detects independently seeded unused exports/types |
| Namespace and ambiguous dynamic-import bans | Parser limitations, not production requirements | Remove extra grammar restrictions; preserve existing canonical catalog discovery requirements |
| `requestPublicJson` named import/call ban | Defective proxy for transport ownership | Actual raw execution module is confined; validated operation behavior qualifies |
| Routes/payload/control-token literals in test-utils | Retain route-neutral shared support boundary in current frontend source pass | Allowed neutral helper and forbidden route/control witnesses qualify |
| E2E contract/package facade string checks | Duplicate existing checker/package exports | Retain canonical positive/negative import witnesses; delete duplicate string searches |
| Direct heading focus versus other heading expressions | Former owner grammar; explicitly amended in WF-06 | Adopt REQ-103 replacement and qualify semantic readiness/focus witnesses |
| Delivery-shaped selector helper names in production source | Retire duplicate private pattern scan; retain applicable canonical semantic identity validation | No second naming registry or compatibility alias |
| Hard-coded authentication selector-family classification | Incidental duplicate inventory | UI Contracts application-selector behavior remains qualified |
| Blanket selector template/raw-literal scan and fixture allowlists | Incidental construction constraints | Deterministic identity, registry validation and helper behavior remain qualified |
| Authentication helper-level literal sinks | Incidental spelling constraints | Preserve actual operation/selector behavior, not a parallel list of accepted source spellings |

S-011 rolls back only the coordinated authored, test, routing, documentation and
generated changes for this iteration. No production data, dependency, golden or
public artifact migration is required. Do not broaden a rollback into unrelated
work or recreate private aliases to accommodate old test arrangements.

#### 8. Validation Plan

All commands run from `/home/jochi/code/cartulary`. Product/harness execution below
is authorized; actual execution and results are recorded in Section 10. Installed pinned Node/pnpm/Fallow and ordinary native
harness prerequisites are required; use the pin registry, not versions copied
into this document. No product browser workflow is selected. The WF-10 execution
contract preset includes its declared build and service-backed harness checks;
these are part of that canonical qualification, not a product browser suite.
Use fresh explicit run IDs and retain exact roots and native outer exits.

| ID / layer | Stage / slice / contracts | Exact command or defined portfolio | Expected result / failure gate |
| --- | --- | --- | --- |
| V-008 / focused characterization | Pre-move; S-006; C-007–C-010 | `make test-slice OWNER=web.architecture`; the pre-move browser selection below; the UI selector selection below | Current owner-aligned baseline plus known incidental checks; failures diagnosed before movement; not current qualification from historical runs |
| V-009 / boundaries and transport | Post-slice; S-007; C-007/C-009 | `make test-slice OWNER=web.architecture ROWS=web.architecture.support.frontend_import_boundary_checker_assurance_393c407a70`; `make lint CHECK=frontend-boundary`; the transport selection below | Real checker integration, fresh positive/negative fixtures, transport semantics and privacy pass; a failed required witness blocks S-008 |
| V-010 / reachability | Post-slice; S-008; C-007/C-010 | `make test-slice OWNER=web.architecture ROWS=web.architecture.support.fallow_static_and_protocol_analyzer_assurance_0794130901`; `make frontend-fallow-static` | Pinned analyzer detects E2E dead peers, keeps true roots live and accepts repository configuration; no suppression growth |
| V-011 / behavioral support | Post-slice; S-009; C-008 | The readiness and UI selector selections below | Negative identity, finite timeout, payload-safe diagnostics and unchanged focus qualify; no syntax scanner needed |
| V-012 / projection/cutover | Post-slice; S-009/S-010; C-010 | `make generate`; `make source-check CHECK=generated`; `make source-check CHECK=schema`; `make source-check CHECK=catalog`; `make source-check CHECK=identity`; `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19` | Authored inputs agree with generated outputs; exact selection and source ownership reconcile; obsolete references absent |
| V-013 / final portfolio | Final regression gate; S-011; C-006–C-010 | `make agent-finalize`, then the frozen-source portfolio below | Review any finalizer mutations before freezing; all selected obligations pass on compatible source identities |
| V-014 / documentation | Document-save check; current invocation | `make lint-markdown` with installed-only preparation and external temporary result/cache paths when safe; manual handoff/diff review | No product execution, installation, generation or authored auto-fix; document actual lint coverage and any skip reason |

##### Exact focused selections

Pre-move browser selection:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131,harness.browser.boundary_support.publicjsonclient_suite_a3aedb76f9,harness.browser.boundary_support.public_http_diagnostic_privacy,harness.browser.boundary_support.source_publicjsonclient_test_ts_a6c5c0d17c
```

Transport selection, retained after scanner deletion:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.publicjsonclient_suite_a3aedb76f9,harness.browser.boundary_support.public_http_diagnostic_privacy,harness.browser.boundary_support.source_publicjsonclient_test_ts_a6c5c0d17c
```

Readiness selection; add new named cases to the same behavioral rows' exact title
lists, keeping independent cases rather than introducing a scenario DSL:

```sh
make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.timelineworkbooktestsupport_waits_for_delayed_wo_8de9145b31,web.architecture.boundary_support.workbookinspectortestsupport_suite_b85a10b67a
```

UI selector selection:

```sh
make test-slice OWNER=package.ui ROWS=package.ui.frontend_unit.selector_core_contracts_375732dc80,package.ui.frontend_unit.application_selector_contracts_675995f2ab
```

##### Final portfolio and evidence limits

After focused exits, `make agent-finalize` precedes the broader final portfolio.
Supply `RESULTS_DIR` only for an existing compatible successful full warm-check
root; otherwise record retained-run maintenance skipped. Do not launch a redundant
check solely to produce that root.

Freeze source after reviewing finalizer changes, then qualify this portfolio.
Compatible complete focused evidence may be retained after the final validity
audit; do not rerun a passing context solely to change its run ID:

- `make test-slice OWNER=web.architecture` — includes actual frontend boundary
  enforcement, its assurance, Fallow assurance, source ownership and readiness.
- The exact transport and UI selector selections above.
- `make frontend-fallow-static` — actual repository reachability, distinct from
  isolated analyzer fixtures.
- `make source-check CHECK=all`.
- `make harness-contract SUITE=execution` — canonical graph assurance for the
  retired-row sentinel and retained standard/full tier placement; the existing
  execution preset is the narrowest public preset containing that assurance.
- `make lint CHECK=types`, `make lint CHECK=frontend`, and `make lint CHECK=scripts`.
- `make lint-markdown` with its actual configured coverage recorded.

Do not separately replay a checker operation already covered by this owner union.
For each accepted root retain the native exit, selected obligations, source and
toolchain identities, summary, cleanup outcome and retained-secret scan. Use
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all` for required
canonical validity, not as a substitute for test execution or full-owner audit.
Use `make test-evidence-audit OWNER=web.architecture EVIDENCE_ROOTS_FILE=<path>`
with an explicit manifest of exact compatible roots before claiming complete
architecture-owner acceptance. The transport and UI selections remain subsets.
After a repair, reassess compatibility and missing obligations before rerunning.
No blanket `check`, CI, product browser/visual/accessibility/UI-review, benchmark
or release qualification is required by this bounded change. Service-backed
checks are limited to the existing execution-contract preset's declared work.

| Validation / invocation admission | Side-effect inspection | Current execution posture |
| --- | --- | --- |
| Discovery: `make help ALL=1`, `make task-guide OWNER=web.architecture`, `make task-guide OWNER=harness.browser` | Public help/diagnostic paths and existing workspace admission inspected; no readiness installation or service work | Revalidated for implementation in E-018; final public preset confirmed in E-026/E-027 |
| V-008–V-013 | Generate/finalizer can rewrite projections; tests and lint create execution artifacts | Authorized and executed through workstream checkpoints; final coverage and limitations in E-029 |
| V-014 | `lint-markdown` can ordinarily install dependencies and write cache/results; installed-only readiness is read-only; supported environment paths permit external temporary outputs | E-016 is historical; implementation lint and manual handoff review are recorded in E-029. Configured lint globs exclude this handoff |

#### 9. Top-Level Work Tracker

| ID | Work item / phase | Workstream | Status | Prerequisite / references | Exit condition |
| --- | --- | --- | --- | --- | --- |
| T-001 | Save and review the document-only plan | Historical planning session | DONE | E-010–E-015; V-014 | Only admitted handoff changed; validation and limits recorded |
| T-002 | Baseline, assertion mapping and concrete owner amendments | WF-06 | DONE | S-006; V-008 | Baseline and amendment disposition recorded before code movement |
| T-003 | Consolidate boundaries and raw execution ownership | WF-07 | DONE | Section 6; F-007/F-010; V-009 | Transport and boundary checkpoint qualifies |
| T-004 | Qualify Fallow E2E reachability | WF-08 | DONE | Section 6; F-008; V-010 | Real live/dead E2E controls qualify |
| T-005 | Replace selector grammar with semantic behavior assurance | WF-09 | DONE | Section 6; F-009; V-011 | Amendment adopted and scanner/five rows removed with qualified successors |
| T-006 | Remove E2E scanner and reconcile routing | WF-10 | DONE | Section 6; F-011; V-012 | No obsolete active scanner/row/exception; generated consistency |
| T-007 | Finalize and qualify compatible final portfolio | WF-11 | DONE | Section 6; V-013 | Section 12 implementation criteria met |

Historical WF-00–WF-05 / S-000–S-005 remain DONE. Their F-001–F-006,
C-001–C-005 and V-001–V-007 identifiers are reserved and are not reassigned to this
iteration. Their evidence is historical and does not qualify the current
implementation.

#### 10. Session Handoff Log

##### Authorized implementation — current checkpoints

The current implementation follows WF-06 → WF-07 → WF-08 → WF-09 → WF-10 → WF-11.
Each completed checkpoint is saved before its successor starts. The earlier
document-only authorization statements and preservation evidence below are
historical; they do not restrict the explicitly authorized implementation.

| Evidence / workstream | State, changes and qualification | Next action |
| --- | --- | --- |
| E-018 / WF-06 entry | IN_PROGRESS. HEAD `971d6423c1bfc5f1f3ac6569fc524b596cf33f61`, branch `main`; only this tracker was dirty. All 7,638 tracked/non-ignored paths and index captured in `/tmp/cartulary-tr01-implementation-r5ea9f4a`; pre-existing tracker preserved as `prior-handoff.md`. Make help and guides for web.architecture, harness.browser and package.ui passed. | Execute V-008, amend REQ-103/676/677/070 and acceptance/support text; retain baseline results before code movement. |

WF-06 exit — E-019: DONE. The user-selected REQ-103 semantic amendment,
REQ-676 discovery/liveness simplification, REQ-677 real transport boundary, and
REQ-070 bounded enforcement are adopted in the NLSpec; AC-021 and the frontend
support guide agree. The fourteen assertion dispositions were revalidated.
Projection fidelity remains subject to human review; execution is not a substitute.

All commands used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run ID>` from the
repository root. Native outer exits were 0:

| Run ID | Command / exact selection | Result / artifact |
| --- | --- | --- |
| `tr01-wf06-architecture-20261026-01` | `make test-slice OWNER=web.architecture` | PASS; `run-summary.json`, 6/6 units |
| `tr01-wf06-transport-20261026-01` | V-008 exact four-row browser/scanner selection | PASS; `run-summary.json` |
| `tr01-wf06-selectors-20261026-01` | V-008 exact two-row package.ui selection | PASS; `run-summary.json` |
| `tr01-wf06-markdown-20261026-01` | `make lint-markdown` | PASS; `adhoc/lint-markdown/tool-run-summary.json`; handoff excluded by lint globs and reviewed manually |

Run roots are `.cartulary/test-results/<run ID>`. Baseline executable source:
`sha256:b296423ca57d61784fda96b379452ee93e05f609ad2fda8d08a584a56b1d8efc`;
toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
`git diff --check` passed. No executable source moved before these baselines.
This checkpoint closes RB-001/RB-002 and is saved before WF-07 starts.

WF-07 entry — E-020: IN_PROGRESS after saved E-019. Boundary checker source
records, transport imports/tests and authored path restrictions revalidated;
V-009 is the exit gate.

WF-07 exit — E-021: DONE. The canonical checker shares source text, AST and
import records, including existing design-token evaluation. Two owner-local
executable-syntax checks replace residual vendor-selector/route-control scans;
comments are ignored. Raw execution moved to `rawPublicJson.ts`, with only the
operation client/direct tests admitted; context retains the sole path validator
and actual shared types. The result-only wrapper is removed. Public/control
separation now covers all three public transport modules. No aliases were added.

V-009 native exits were 0 (same environment/root convention as E-019):

| Run ID | Command / selection | Result |
| --- | --- | --- |
| `tr01-wf07-boundary-20261026-01` | Exact frontend boundary checker assurance row | PASS; positive/negative syntax, raw import/re-export and privileged-control fixtures |
| `tr01-wf07-transport-20261026-01` | Exact three transport rows | PASS; actual raw execution, omission/supplied values, absolute/protocol-relative/backslash escape rejection, operation validation and diagnostic privacy |
| `tr01-wf07-integration-20261026-01` | `make lint CHECK=frontend-boundary` | PASS; actual repository source integration |

Each root retains `run-summary.json`; focused source identities are recorded in
its manifest. These are slice witnesses, not the frozen final portfolio. This
checkpoint is saved before WF-08 begins.

WF-08 entry — E-022: IN_PROGRESS after saved E-021. Existing Fallow owner,
real-analyzer fixture and publication tests revalidated; V-010 is the exit gate.

WF-08 exit — E-023: DONE. Authored roots now select E2E specs and explicit
setup/teardown, and the Vitest owner includes E2E support tests. Ordinary helpers
are not unconditional roots. Real pinned-analyzer fixtures qualified live/dead
helpers, unused values/types, type-only/namespace/re-export/literal dynamic imports,
and the negative blanket-root control. Publication rejects seeded E2E findings
even with analyzer exit 0; existing advisory behavior remains covered.

V-010 native exits were 0:

| Run ID | Command / selection | Result |
| --- | --- | --- |
| `tr01-wf08-analyzer-20261026-01` | Exact Fallow analyzer assurance row | PASS; real analyzer and publication controls |
| `tr01-wf08-repository-20261026-01` | `make frontend-fallow-static` | PASS; repository integration with zero bounded web/package findings |

Roots use the E-019 convention and retain run/target summaries, resolved config
and Fallow reports. Advisory findings remain diagnostic evidence; no suppressions
or unrelated cleanup were introduced. `git diff --check` passed. RB-003 is closed.
This checkpoint is saved before WF-09 begins.

WF-09 entry — E-024: IN_PROGRESS after saved E-023 and adopted semantic
readiness amendments. V-011 plus projection consistency gate scanner retirement.

WF-09 exit — E-025: DONE. Eight new controlled behavioral cases cover wrong
record/surface, lower and higher incorrect inspector versions, persistent mismatch,
finite configured timeouts, payload exclusion and observation without focus,
click or event dispatch. Existing helpers needed no behavior correction. Fake
clocks drive existing polling/deadlines; no production timeout was increased.
The selector scanner and its five rows/source-ownership entry are removed after
successor qualification; README guidance now describes semantic assurance.

The initial `make generate` root `tr01-wf09-generate-20261026-01` failed with
native exit 2 (normalized harness/unknown_failure) because newly authored title
lists were not ASCII-sorted. Its `generate/tool-run-summary.json` and
`generate/generate-task-surface-and-execution-topology/stderr.log` retain the
failure. Sorting the two affected row title lists repaired the authored input;
no generator or generated output was hand-edited.

Subsequent native exits were all 0:

| Run ID | Command / selection | Result |
| --- | --- | --- |
| `tr01-wf09-generate-20261026-02` | `make generate` | PASS after authored sorting repair |
| `tr01-wf09-format-20261026-01` | `make format` | PASS; only intended authored TypeScript formatting changed |
| `tr01-wf09-readiness-20261026-01` | Exact two readiness rows with new title inventories | PASS; 11 selected cases |
| `tr01-wf09-selectors-20261026-01` | Exact two package.ui selector rows | PASS |
| `tr01-wf09-cutover-generate-20261026-01` | `make generate` | PASS after scanner/row removal |
| `tr01-wf09-generated-20261026-01` | `make source-check CHECK=generated` | PASS |
| `tr01-wf09-schema-20261026-01` | `make source-check CHECK=schema` | PASS |
| `tr01-wf09-catalog-20261026-01` | `make source-check CHECK=catalog` | PASS |
| `tr01-wf09-identity-20261026-01` | `make source-check CHECK=identity` | PASS |
| `tr01-wf09-ownership-20261026-01` | Exact source-ownership row | PASS |

Roots use the E-019 convention. Generate roots retain `generate/tool-run-summary.json`;
graph roots retain `run-summary.json`. Active source/routing references to the
selector scanner are absent; `git diff --check` passed. V-011 and this cutover's
V-012 gates are satisfied. This checkpoint is saved before WF-10 begins.

WF-10 entry — E-026: IN_PROGRESS after saved E-025. All scanner successors
qualified; V-012 gates final removal and authored/generated routing reconciliation.

Cutover inspection also found the retired E2E row in the graph assurance's
independent full-tier sentinel list. Remove that obsolete sentinel and assert
standard placement for both retained readiness rows; keep both existing analyzer
assurance sentinels in full. V-012 therefore additionally requires
`make harness-contract SUITE=execution`, the existing public graph-assurance
route. No scheduler implementation, preset, or tier policy changes are needed.

WF-10 exit — E-027: DONE. Removed the 655-line E2E scanner, its catalog row,
core-support exception and graph sentinel. Both readiness rows are asserted as
standard tier and both existing analyzer assurance rows remain full tier.
Authored Fallow help now describes bounded web/package gates; Make regenerated
the task-surface projections and topology render index. No executable reference
to either retired scanner or its rows remains under `apps`, `tools` or `contracts`.

V-012 native exits were 0, under the E-019 environment/root convention:

| Run ID | Selection | Result / duration ms |
| --- | --- | --- |
| `tr01-wf10-generate-20261026-01` | `make generate` | PASS; 11170 |
| `tr01-wf10-generated-20261026-01` | `make source-check CHECK=generated` | PASS; 75219 |
| `tr01-wf10-schema-20261026-01` | `make source-check CHECK=schema` | PASS; 55972 |
| `tr01-wf10-catalog-20261026-01` | `make source-check CHECK=catalog` | PASS; 46176 |
| `tr01-wf10-identity-20261026-01` | `make source-check CHECK=identity` | PASS; 49530 |
| `tr01-wf10-ownership-20261026-01` | Exact source-ownership row | PASS; 45726 |
| `tr01-wf10-graph-20261026-01` | `make harness-contract SUITE=execution` | PASS; 20/20 units; 174681 |

These graph roots identify source
`sha256:daaea890979e0349dc4392aeffd167b41b172981b29d13dca99008fbad112eb9`.
Durations include concurrent scheduling and make no performance claim.
An exploratory `make help TARGET=harness-contract` was rejected with exit 2
because `TARGET` is not a help input; `make help ALL=1` supplied the public
selection instead. This discovery rejection did not launch verification.
`git diff --check` passed. This completed checkpoint is saved before WF-11 starts.

WF-11 entry — E-028: IN_PROGRESS after saved E-027. Run `make agent-finalize`
without `RESULTS_DIR`: no compatible successful full warm-check root was supplied,
so retained-run maintenance is intentionally skipped. Review generated mutations,
then freeze executable source and qualify V-013. Reuse E-027's complete execution
contract only if finalizer review and canonical validity establish compatibility;
otherwise rerun it on the final source. Documentation cleanup cannot supply
executable evidence and is reviewed separately.

WF-11 exit — E-029: DONE. Finalizer, compatible executable qualification,
documentation review, preservation comparison and handoff are complete. The finalizer
reported `generated.status=unchanged`, zero updated files and no rollback.
Executable source is frozen at
`sha256:daaea890979e0349dc4392aeffd167b41b172981b29d13dca99008fbad112eb9`;
all roots below use toolchain
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
The finalizer and final portfolio used `CGO_ENABLED=0`, fresh explicit run IDs and
the default `.cartulary/test-results/<run-id>` roots. Native outer exits were 0.

| Run ID | Exact selection / distinct obligation | Passed work units / duration ms |
| --- | --- | --- |
| `tr01-wf11-finalize-20261026-01` | `make agent-finalize`; required maintenance before final verification | 1/1; 29097 |
| `tr01-wf11-architecture-20261026-01` | `make test-slice OWNER=web.architecture`; complete owner, including real checker, analyzer assurance, readiness and ownership | 6/6; 8078 |
| `tr01-wf11-transport-20261026-01` | Section 8's exact three-row `harness.browser` transport selection; raw/validated semantics and privacy | 2/2; 3949 |
| `tr01-wf11-selectors-20261026-01` | Section 8's exact two-row `package.ui` selection; stable selector identity | 2/2; 2338 |
| `tr01-wf11-fallow-20261026-01` | `make frontend-fallow-static`; actual repository reachability, separate from fixtures | 2/2; 21623 |
| `tr01-wf11-source-20261026-01` | `make source-check CHECK=all`; complete generated/schema/catalog/surface/identity/toolchain/migration-input consistency | 9/9; 25696 |
| `tr01-wf11-types-20261026-01` | `make lint CHECK=types`; frontend compile/type compatibility | 2/2; 27089 |
| `tr01-wf11-frontend-20261026-01` | `make lint CHECK=frontend`; authored frontend lint | 2/2; 3273 |
| `tr01-wf11-scripts-20261026-01` | `make lint CHECK=scripts`; authored harness syntax/lint | 2/2; 749 |
| `tr01-wf10-graph-20261026-01` | Retained compatible `make harness-contract SUITE=execution`; graph/tier cutover and preset's declared supporting checks | 20/20; 174681 |
| `tr01-wf11-owner-audit-20261026-01` | `make test-evidence-audit OWNER=web.architecture EVIDENCE_ROOTS_FILE=/tmp/cartulary-tr01-implementation-r5ea9f4a/web-architecture-evidence.json`; complete-owner evidence admission | 1/1; 2868 |

Each graph root retains `run-manifest.json`, `run-summary.json`, target/row
projections, `cleanup-results.json` and `retained-secret-scan.json`. All eleven
passed `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all` with
native exit 0 and `full_validity=1`. All selected units passed, no failed/skipped/
cancelled units remain, cleanup outcomes are `completed`, and retained-secret
scans passed. The ten execution/finalization roots had zero action-cache hits;
this is functional evidence, not performance qualification. Optional resource
diagnostics were not collected.

The owner audit's `test-evidence-audit/audit.json` reports `complete_owner`, all
nine required `web.architecture` rows accepted, and no rejected or unused inputs.
Its manifest names only the exact final architecture run. Transport and UI
selector runs qualify their explicit subsets, not their complete owners. This
portfolio closes the bounded remediation scope without claiming whole-project
acceptance.

The final Fallow report has zero blocking web or package findings. Its raw health
report exits 1, retained as the existing nonblocking warning; findings outside
the bounded policies remain advisory. See
`tr01-wf11-fallow-20261026-01/frontend-fallow-static/fallow-static-summary.json`
and `tool-run-summary.json` under the result parent. No suppression, baseline,
dependency or analyzer policy was weakened to obtain the pass.

Finalizer retained-evidence and scheduler-drift maintenance were skipped because
`RESULTS_DIR` was unset; the detailed reason is `results-dir-not-provided` in
`tr01-wf11-finalize-20261026-01/unit-artifacts/finalize-summary.json`. No compatible
full warm check was supplied or manufactured. The per-root validity and complete
architecture-owner audit above are separate completed obligations.

Full `check`/CI, product browser workflows, visual/accessibility/UI-review,
dedicated race testing, benchmarks and release qualification were not selected:
no behavior in those areas changed. The execution-contract preset did run its
declared build and service-backed harness checks; their cleanup passed. The
previously diagnosed WF-09 title-order generation failure remains recorded in
E-025; its repaired generation and final source checks passed.

Manual preservation review against E-018 found only the scoped 26 changed
existing files, the new raw transport module and the two planned scanner
deletions. The index and HEAD are unchanged. Generated outputs also received
content-identical file replacements during Make generation; these are file
identity changes, not additional source changes. Comparison details are in
`/tmp/cartulary-tr01-implementation-r5ea9f4a/implementation-comparison.json`.
The pre-existing tracker is retained as `prior-handoff.md`; the separate
`testing-remediation.md` is unchanged. Ignored builds, dependencies and retained
results were not exhaustively content-audited.

Documentation validation: `CGO_ENABLED=0
CARTULARY_TEST_RUN_ID=tr01-wf11-markdown-20261026-01 make lint-markdown` passed
with native exit 0 in 8080 ms. Its leaf summary is
`.cartulary/test-results/tr01-wf11-markdown-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The configured globs cover the changed NLSpec, frontend guide and source READMEs,
but exclude this handoff. Manual review additionally checked all twelve handoff
sections, table widths, fences, local links, whitespace, the fourteen assertion
dispositions and historical/current status separation; no issues were found.
`git diff --check` passed. This leaf documentation result is not a graph result
or executable acceptance evidence.

Handoff: all six authorized workstreams are complete. Review the adopted owner
amendments and downstream implementation together; passing checks do not replace
human projection-fidelity review. There is no pending implementation gate in
this bounded effort. No public API, production data, dependency, golden or
release migration is required. Private callers and exact test routing were
migrated without shims. Continue any separate WF-33 work from its own preceding
handoff; this completion does not close or alter that effort.

Confirmed specification decisions: replace heading matcher grammar with bounded
identity-based, observation-only readiness; retire the closed E2E filename list
and custom external-importer counting; describe the real raw execution boundary;
retain the existing bounded Fallow web and package gates and advisory findings
elsewhere. Fallow stays out of local `check` report execution. No new analyzer,
schema, runtime mode, compatibility wrapper or suppression growth is admitted.

##### Historical document-only planning session — E-010–E-017

The following subsections retain that session’s observations and limits. Phrases
such as “current” and “next action” below refer to that earlier session; the
implementation checkpoints above supersede its pending work.

###### Scope and authority

| Time / baseline | State and evidence | Next action |
| --- | --- | --- |
| 2026-10-26; `971d6423c1bfc5f1f3ac6569fc524b596cf33f61` | E-010/E-011: user selected in-place update, frontend policies and simplification at source. Prior Plan Mode allowed inspection only. Current task authorizes saving that document-only plan. | Revalidate baseline before later implementation; do not inherit historical implementation permission |

###### Backend boundary

Not applicable: no backend, Go fixture, production assembly, storage or service
change is part of this iteration. Initial exploratory fixture reads selected no
backend workstream.

###### Frontend boundary

| Evidence | Inspected state / affected files | Next action |
| --- | --- | --- |
| E-012/E-013 | Two bespoke scanners duplicate established analysis and encode incidental constraints; Fallow roots hide a real successor-coverage gap | S-007/S-008 establish successors before S-009/S-010 removals |

###### Contract and codegen

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-013/E-015 | REQ-103 grammar is currently adopted; authored row/source inputs and generated outputs are identified; no owner/projection edits occurred | Propose/adopt only the concrete amendment; regenerate from authored inputs during separately authorized implementation |

###### Tests and harness

| Evidence | Commands/results and applicability | Next action |
| --- | --- | --- |
| E-014/E-015 | Assertions and command bindings inspected; read-only help and two task guides passed; no tests ran | Use V-008 as the future baseline, not historical passes |
| E-016 | `make lint-markdown` passed, outer exit 0; installed-only preparation and external result/cache directories. Handoff structure, links and diff reviewed separately because lint globs exclude this file. | No tests, installation, generation or finalizer run; detailed command and coverage below |

###### Security and authorization

| Evidence | Required distinction | Next action |
| --- | --- | --- |
| E-014 | Preserve actual HTTP validation, same-origin boundaries and private diagnostics; current wrapper-name guard is incomplete | Enforce the module that actually performs raw execution; keep success/error/observed-header witnesses |

###### Open risks and next session

| Evidence | Current state / limits | Next action |
| --- | --- | --- |
| E-017 | All 7,638 tracked/non-ignored paths compared: only this handoff content changed; no added/removed paths or file-identity changes; HEAD and index unchanged | Later actor starts with baseline/adoption/characterization revalidation, not broad rediscovery |

###### Historical planning evidence register

| ID | Source / observation | Baseline and applicability |
| --- | --- | --- |
| E-010 | Git status/HEAD/index and per-path content/file-identity capture; 7,638 paths; exact comparison files in Section 1 | Current planning/save baseline; matches prior-turn fingerprints; no current test evidence |
| E-011 | Root/tooling guidance, Refactor Tracker format and modular framework; target/output identity and link checks | Applies to this sole-file write; prior completion authorization is historical |
| E-012 | E2E scanner, frontend checker/rules, Fallow configuration/fixtures and actual runner configuration | Static source evidence for F-007/F-008; real analyzer successor qualification remains future |
| E-013 | Selector scanner and adopted REQ-103/AC-021; proposed semantic amendment | Current requirement distinguished from desired future requirement; no adoption inferred |
| E-014 | Transport modules/tests, row/inspector waits/tests, selector package tests and direct import search | Test assertions inspected, not executed; no exhaustive external-consumer claim |
| E-015 | Relevant catalog, source ownership, task-surface and generated policy inputs; Make help and owner guides | Commands discovered and routing known; generated products not edited |
| E-016 | Document write, successful Markdown lint (outer exit 0), manual structural/link review and clean `git diff --check` | Lint covers configured globs, which exclude this handoff; supplemental inspection found twelve required sections, fourteen assertion dispositions and no structural/link issues |
| E-017 | Final index/content/file-identity comparison; `/tmp/cartulary-tr01-doc-hla6fi0g/comparison.json` | Only admitted output content changed; no added/removed paths or identity changes; HEAD/index unchanged. Ignored trees were not exhaustively audited |

###### Historical document validation — E-016/E-017

Executed from the repository root:

```sh
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-tr01-doc-hla6fi0g/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-tr01-doc-hla6fi0g/static-analysis \
CARTULARY_TEST_RUN_ID=tr01-frontend-plan-doc-20261026-01 \
COREPACK_ENABLE_NETWORK=0 make lint-markdown
```

Outer exit: **0**. Step summary:
`/tmp/cartulary-tr01-doc-hla6fi0g/results/tr01-frontend-plan-doc-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
The existing lint configuration excludes this handoff; the pass qualifies only
its configured documents. Supplemental document inspection covered section
structure, table widths, fenced blocks, whitespace, local links, the fourteen
assertion dispositions and preservation of historical evidence. `git diff
--check` passed. These are documentation checks, not test execution or product
qualification. Review notes and the content/index comparison are retained in
`document-review.json` and `comparison.json` beside the temporary baseline.
No tests, format, generation, finalizer or product verification ran, as required
by the current document-only scope.

##### Historical graph/scanner iteration — E-001–E-009

The previous iteration completed WF-00–WF-05 / S-000–S-005: canonical graph
assurance replaced duplicate modes; graph policy, diagnostics and execution
fixtures gained clear owners; four scanner shell suites became directly registered
Node tests; Staticcheck now fails before scanning when package discovery fails or
is empty. Public commands, security thresholds and distinct diagnostic/publication
witnesses were preserved. No aliases or production test-only APIs were retained.

The full prior tracker is retained in Git at
`971d6423c1bfc5f1f3ac6569fc524b596cf33f61:docs/handoffs/testing-remediation-01.md`
and in the temporary pre-edit copy identified in Section 1. Its statements about
uncommitted work describe that historical session, not this clean starting
checkout. Artifact contents have not been re-audited in this document session.

| Evidence | Historical outcome and retained reference | Qualification limit |
| --- | --- | --- |
| E-001 | Prior implementation resumed on HEAD `3d192eea610867bac59f4eaf79ba54f4dd05dd68` with preceding dirty work present | Historical authorization only |
| E-002 | Baseline `tr01-baseline-20261026-01`: regression passed, 55 graph units; source `sha256:aa432d498946013285d924d51e7f754a7d1bfc37a40492be0cb5c09e413ca9d0` | Pre-move historical baseline |
| E-003 | Graph policy/wiring split; first expectation failure repaired; `tr01-wf01-graph-20261026-02` passed, 60 tests | Slice evidence, earlier source |
| E-004 | Canonical matrix and explicit published-unit fixtures; generation and invalid fixture-ID failures repaired; `tr01-wf02-graph-20261026-02` passed, 74 tests | Earlier-source generation/schema roots remain in prior tracker |
| E-005 | Four scanner suites consolidated; `tr01-wf03-scanner-20261026-01`, `tr01-wf03-graph-20261026-01`, `tr01-wf03-source-20261026-01` passed | Wrapper refactor before discovery correction |
| E-006 | `tr01-wf04-negative-20261026-01` exposed partial-output discovery failure; corrected scanner and actual Go lint passed at `tr01-wf04-scanner-20261026-01` / `tr01-wf04-go-20261026-01` | Intentional fail-closed correction; earlier source |
| E-007 | Initial final contract failed because graph umbrella imported a backend-private test; direct operation registration repaired ownership | Initial portfolio superseded |
| E-008 | Regeneration, exact boundary row and second finalizer passed; final compatible portfolio below accepted | Retained maintenance excluded because RESULTS_DIR unset |
| E-009 | Historical `tr01-final-markdown-20261026-01` lint passed; prior bounded iteration closed | No current-document or current-source test qualification |

##### Historical final acceptance portfolio

The following retained portfolio is reproduced from E-008/E-009. Its statements
report the earlier session; they do not certify this iteration or a new run.


Every command below used `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run ID>`. Every
native command exited 0. Roots are `.cartulary/test-results/<run ID>`; each contains
`run-manifest.json`, `run-summary.json`, target/unit results, `cleanup-results.json`
and `retained-secret-scan.json`.

Executable source: `sha256:b296423ca57d61784fda96b379452ee93e05f609ad2fda8d08a584a56b1d8efc`.
Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
Earlier-source passes and failed characterization runs are historical evidence,
not part of this final acceptance union.

| Run ID | Command | Duration ms | Added assurance |
| --- | --- | --- | --- |
| `tr01-final-contract-20261026-02` | `make harness-contract` | 114368 | Default contracts, graph matrix, owner boundaries and real Make/private diagnostic channel; 11/11 units |
| `tr01-final-scanner-20261026-02` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.go_scanner_wrappers` | 14457 | All 20 wrapper cases, including discovery failure before scan |
| `tr01-final-output-20261026-02` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.output_accounting` | 27139 | Distinct step/target/tool publication, including valid and malformed Govulncheck findings |
| `tr01-final-source-20261026-02` | `make source-check CHECK=all` | 35700 | Generated, schema, catalog, identity, surface, toolchain and migration-input consistency |
| `tr01-final-scripts-20261026-02` | `make lint CHECK=scripts` | 625 | Authored script validation |
| `tr01-final-shell-20261026-02` | `make lint CHECK=shell` | 23784 | Shell analysis including the changed Staticcheck wrapper |
| `tr01-final-go-20261026-02` | `make lint CHECK=go` | 14032 | Actual Go formatting, vet and Staticcheck invocation |

For every listed root, `make harness-observability-check RESULTS_DIR=<exact-root>
CHECK=all` passed with outer exit 0 and `full_validity=1`. All cleanup outcomes
were `completed`; retained-secret scans passed; no failed, skipped or cancelled
units remain. This qualifies the selected obligations, not complete acceptance
of every owner or the whole project. No performance claim follows from these
functional timings; optional resource diagnostics were not collected.

Full CI, browser/visual/UI-review, dedicated race-suite execution, benchmarks,
release qualification and complete audits of other owners were not run: no
production behavior in those areas changed. Their selection/isolation guarantees
remain covered by the graph assertions. Finalizer canonical retained-evidence and
scheduler-drift maintenance were skipped because `RESULTS_DIR` was unset; no
compatible full warm check was supplied and no redundant check was created for
maintenance. This does not replace the per-run validity checks recorded above.

#### 11. Open Questions and Blockers

No unresolved scope or implementation-design choice blocks this implementation.
The following entry gates are closed by actual workstream evidence, not planning
completion or historical passes.

| ID | Blocker / affected slice | Needed evidence or authority | Status / resolution condition |
| --- | --- | --- | --- |
| RB-001 | Current pre-move characterization for S-007 | V-008 on a revalidated source baseline | DONE; E-019 retains all three passing baseline roots |
| RB-002 | Owner amendments for S-007–S-010 | User selected concrete semantic/simplification/bounded-gate decisions and authorized implementation | DONE; E-019 records adopted REQ-103/676/677/070 and AC-021 changes |
| RB-003 | Real Fallow E2E successor assurance | V-010 real analyzer, publication and repository integration | DONE; E-023 |
| RB-004 | Successor and final portfolio gates; S-009–S-011 | Predecessor exits and V-011–V-013 | DONE; E-025/E-027/E-029 close all gates |

The authorized owner amendments are adopted in E-019. F-014 resolved the
ambiguous Fallow specification/help wording while retaining the existing bounded
implementation policy. Human projection review remains distinct from passing tests.

#### 12. Binary Completion Criteria

| Outcome | Current statement |
| --- | --- |
| Planning completeness | **complete** for the bounded frontend test-policy iteration; not an exhaustive suite audit |
| Workstream completion | WF-06–WF-11 **DONE**; each predecessor checkpoint was saved before the next entry; validation and handoff completed last |
| Implementation authorization | **granted by the current implementation instruction** for WF-06–WF-11 and their scoped verification |
| Implementation write boundary | **Verified**, E-029: scoped source/specification/test/routing edits and generated projections only; index/HEAD unchanged. E-017’s sole-file result belongs to historical planning |

Historical document-only completion criteria (E-016/E-017):

- [x] Existing canonical target and requested output identity established.
- [x] Current observations distinguish direct inspection from metadata accounting.
- [x] Each finding maps to a contract, slice and validation obligation or justified no-action disposition.
- [x] All fourteen baseline assertion groups have a successor or reasoned retirement.
- [x] Dependency order, entry gates, rollback and exact validation selections are defined.
- [x] Incidental rule changes are separated from unchanged production requirements.
- [x] Historical completion IDs, final portfolio and exclusions are retained.
- [x] No tests, production changes, owner amendments or generated edits were executed.
- [x] Documentation validation and its actual coverage/limitations recorded.
- [x] Final content/index comparison verifies the admitted sole-file write.

Implementation completion criteria (updated at workstream exits):

- [x] Both bespoke scanners and obsolete routing are removed without aliases.
- [x] Every meaningful obligation has qualified successor coverage at its appropriate boundary.
- [x] Incidental naming, literal and matcher restrictions are retired explicitly.
- [x] The concrete REQ-103 amendment is adopted before dependent removals.
- [x] Actual raw transport execution is confined and all retained transport semantics qualify.
- [x] Fallow detects dead E2E peers without blanket roots or suppression growth.
- [x] Selector identity and bounded observation-only readiness remain qualified.
- [x] Authored/generated state and exact selected-row accounting agree.
- [x] No new analyzer, compatibility mechanism or generic policy framework is introduced.
- [x] Final compatible evidence, exclusions and all six checkpoints are recorded.

T-001 records the historical document update. Current implementation progress is
recorded by T-002–T-007 and the current checkpoints in Section 10. The separate
WF-33 remains open in the preceding handoff; historical passes do not qualify
the current implementation.

</details>

## 11. Open Questions and Blockers

| ID | Blocker / affected scope | Required evidence / authority | Status | Resolution condition |
| --- | --- | --- | --- | --- |
| RB-005 | Current pre-move characterization; C-011–C-016 and S-013 onward | V-015 baseline and completed assertion ledger | DONE | E-037 qualifies all required pre-move boundaries |
| RB-006 | Predecessor exits for S-014–S-016, and WF-12 gate for S-013 | Exact Section 6 checkpoints and qualified successor assertions | DONE | E-037–E-041 satisfy the ordered exits; WF-16 validation and handoff complete |
| RB-007 | Review follow-up characterization and completion; S-017–S-019 | F-021–F-026 fault witnesses and V-021–V-023; current implementation authorized | DONE, E-047 | All six findings corrected; final compatible validation and handoff saved |

There is no unresolved planning choice about selecting the subsystem, including
S3, retaining mixed ownership, or preserving incidental test behavior. Those
choices are settled in Sections 1–7. No primary-owner contradiction was found.
The next discovery boundary is a newly encountered physical-start/template caller
or a failure that demonstrates an omitted guarantee; investigate that boundary
without expanding into unrelated test-platform redesign.

## 12. Binary Completion Criteria

| Outcome | Current statement |
| --- | --- |
| Planning completeness | **Complete for the bounded PostgreSQL fixture iteration**: ownership, essential guarantees, assertion dispositions, ordered slices, exact starting validation and deferrals are specified. This is not an exhaustive repository test audit. |
| Per-slice completion / readiness | **S-012–S-016 DONE historically**, E-037–E-041. **S-017 DONE, E-045; S-018 DONE, E-046; S-019 DONE, E-047; all three closure checkpoints saved.** |
| Implementation authorization | The current closure implementation request authorizes WF-17–WF-19, atomic callers/routing and required validation. Review-only restrictions in E-042/E-043 describe those historical invocations. |
| Session write-boundary verification | E-043 retains the review-only boundary. E-047 verifies 33 closure-modified entries, prior edits/history preserved, unchanged staged Git entries, and all 7,338 frozen executable/input hashes unchanged after finalization; ignored-tree limits disclosed. |
| Implementation and final validation | **Acceptance requalified and closed by E-047.** E-041 passes remain valid for their selected witnesses; F-021/F-024 and V-021 are qualified at E-045; F-022/F-023/F-025/F-026 and V-022 are qualified at E-046; V-023 and bounded overall acceptance are COMPLETE at E-047. Historical failures and historical-source successes remain preserved separately. |

The earlier planning-only acceptance required: admitted output and baseline; accounted enumeration
with honest inspection depth; adopted owners for each essential guarantee;
findings linked to slices and validation or explicit deferral; acyclic workflow
dependencies; readiness/rollback/exit conditions; exact future selections and
truthful then-unexecuted status; preserved historical rationale; pre-write recheck;
and final proof that only the admitted tracker changed. E-030–E-035 carry that
evidence. Documentation lint alone does not validate the production guarantees.

| Planning acceptance group | Result / evidence |
| --- | --- |
| Target/output identity, pre-write check and final preservation | Met; E-030/E-034/E-035, with ignored-tree limit disclosed |
| File accounting and bounded direct inspection | Met for this seam; Section 2 and E-031; no whole-repository assertion audit claimed |
| Required guarantees separated from incidental test design | Met; C-011–C-016 and F-015–F-020; E-031/E-032 |
| Traceability, workflow dependencies, slice gates and rollback | Met; Sections 5–7; WF-12–WF-16 form one acyclic chain with explicit binary exits |
| Command discovery and truthful evidence scope | Planning evidence E-033/E-035 retained; authorized execution completed at E-037–E-041 |
| Durable history and next-session handoff | Met; preserved snapshot, current evidence register, RB-005/RB-006 and separate readiness/authorization outcomes |

The following E-041 criteria are requalified by the F-021–F-026 corrections and
the final E-047 portfolio. All bounded implementation criteria are met:

- Physical PostgreSQL ownership and fixture borrowing have distinct types and
  lifetimes; one owned-start API remains; borrowed handles cannot shut down an owner.
- One canonical ordinary template preparation path remains, with close-before-seal
  and seal-before-publication; ordinary clones do not replay migrations or fall
  back to implicit provisioning; explicit scratch capabilities remain supported.
- Quarantine, recovery authority, exact owned cleanup, transaction settlement,
  role/transport admission and canonical migration requirements remain qualified.
- Contract tests use local dependencies and coherent assertion groups; unused
  resolver logic, duplicate cleanup/smoke assertions and method-count constraints
  are removed with recorded meaningful successors.
- Direct callers and authored/generated routing agree without wrappers, retired-ID
  aliases, redundant evidence records or new test-only compatibility behavior.
- A compatible final validation portfolio qualifies distinct unit, real database
  and process boundaries after finalization; no unexecuted, stale or subset result
  is represented as a complete owner, release or performance claim.
- WF-12–WF-16 checkpoints and final evidence explicitly record remaining limits;
  no product API/schema/data behavior or unrelated subsystem was changed to keep
  historical test assumptions alive.
