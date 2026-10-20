# Testing harness remediation 02 — Efficiency assessment and handoff

## Executive summary

The authorized implementation has completed runner grouping, incremental Vitest
evidence, incident test setup consolidation, canonical native assurance routing,
durable pre-ready recovery authority and Markdown coverage. Specification
amendments preceded implementation. Section 10 records source/run identities,
failed attempts, migration limits and validation; Section 9 is the execution
tracker. Final validation and handoff (S-008) have been performed. Aggregate acceptance
remains blocked by a current PostgreSQL SQL-readiness failure; the earlier missing-proof
runtime also remains unresolved. Neither is hidden by the successful focused
contracts, source checks or the earlier CI run.

Optional isolation relaxation and cross-work-unit PostgreSQL reuse were reviewed
and deferred for missing independence, benefit or lifecycle proof. No global cost, capacity, worker or deadline tuning was adopted. Canonical
assurance routing and pre-ready recovery authority were repaired after final
validation exposed their gaps. These are explicit terminal dispositions,
not implemented features or claims of measured improvement. The original
candidate priorities and gates remain below for traceability.

| Priority / slice | Opportunity | Expected benefit, not measured savings | Complexity / reliability risk | Decision gate |
| --- | --- | --- | --- | --- |
| 1 / S-001 | Remove unnecessary runner partitions, starting with repeated selections from the same Vitest file | Medium–high if coordinator, import and setup costs dominate | Medium / medium: selection and owner attribution | Establish which partitions are required; preserve every selected outcome and adopt any necessary owner amendment |
| 1 / S-002 | Incremental Vitest reconciliation and publication | Medium if cumulative processing and writes are material | Medium / medium: incomplete or incorrectly superseded evidence | Preserve interruption checkpoints, redaction, native-report reconciliation and invocation-wide failure semantics |
| 2 / S-003 | Replace redundant incident-creation setup while retaining distinct assertions | Low–medium per affected journey | Low–medium / medium: accidental coverage loss | Map each removed step to surviving coverage; retain real creation, authorization, preference and shell assertions |
| 2 / S-004 | Revisit isolation only for newly proved candidates | Potentially high where service/reset costs dominate | Medium–high / high without full effect tracing | Prior exclusions remain valid until disproved; no blanket conversion |
| 3 / S-005 | Sequential exclusive PostgreSQL service reuse across compatible work units | Potentially high if physical startup/template initialization dominate | High / high: cluster residue, checkpoint interference and ownership | Adopt a changed lifetime contract and prove safe settlement; retain fresh test databases |
| 3 / S-006 | Calibrate execution costs and resource admission | Workload dependent | Medium / medium: oversubscription and starvation | Measure the bottleneck and uncertainty before changing capacity or packing |

The executed sequence was WF-00 → S-007 → S-002 → S-001 → S-003 → S-004 →
S-005 → S-006 → S-008 review → reopened S-006 → S-009 → S-008 completion.
The tracker was updated at every workstream boundary.
Remediation 01's completed DOM, credential, rollback, cleanup and cohort work
was retained; its excluded candidates were not silently reopened.

The historical W1 first-drop and W2 readiness investigations remain open. Successful
historical runs do not establish their causes, current performance, or acceptance
of these proposed changes. The complete prior record is retained in
[Appendix A](#appendix-a--historical-remediation-record).

## 1. Scope and Source Posture

- Repository and supplied workspace: `/home/jochi/code/cartulary`; canonical
  repository-relative anchor: `.`. The primary target is the cross-directory
  test-execution and fixture-lifecycle seam, not a new permanent module.
- Output: `docs/handoffs/testing-harness-remediation-02.md`; label:
  `testing-harness-remediation-02`. The existing document's harness identity was
  established from its contents and source references. Its path and ancestors
  were checked for symlinks; the existing output is a regular, single-link file.
- Artifact mode: **normal execution tracker**, with bounded source coverage and
  explicitly gated optional optimizations. The earlier static assessment remains
  provenance; it is not an exhaustive test inventory or performance benchmark.
- Authorization: the user approved implementation of the proposed plan on
  2026-10-20. Each implementation slice is a separate sequential workstream.
  Updating this tracker after each workstream and before starting the next is
  mandatory. Historical W1/W2 causal closure is separate from this effort's
  implementation and handoff completion.
- Framework: [modular refactor planning framework](cartulary_modular_refactor_planning_framework.md).
  The adopted/current [Testing Harness NLSpec v3](../testing-harness-nlspec.md)
  owns mechanics. Its REQ-379/803/804/805/806/811 and AC-078/085/086/087/088/095
  constrain this plan. Root, tooling and backend instructions apply to the
  corresponding proposed work; the framework and this handoff do not amend them.
- Product behavior, authorization, storage, HTTP/WebSocket and UI contracts remain
  with their adopted product owners. This handoff does not establish completeness
  of those specifications. Markdown must remain outside executable test,
  generation, conformance, runtime-metadata and release inputs.

The following table records the **earlier planning session**, not the authorized
implementation's final state. Section 10 records the execution baseline and
current evidence.

| Baseline item | Starting evidence / identity | Planning-session outcome | Limits |
| --- | --- | --- | --- |
| Branch / commit | `codex/specification-remediation-20261016`; `f9f94168816c9c14438c4efb1e2aa24ccdcdd710` | Unchanged on final inspection | Same baseline as the preceding static audit |
| Index, tracked working tree, non-ignored untracked files | Clean `git --no-optional-locks status --porcelain=v1`; index-entry SHA-256 `402244a8201275e56ffa7914cfc2c3c15943d686e65bb84431e83d90294cc6cd` | Index-entry fingerprint unchanged; only this handoff differs from HEAD; no non-ignored untracked files | Logical index entries fingerprinted with `git ls-files --stage -z`; no index refresh requested |
| Existing handoff | 37,978 bytes; SHA-256 `d16ab6f04cbc810484db0823e8e2b6d0322e8cd360e9dfc05cc7b9f8eb740236` | Original body matches Appendix A exactly after the declared heading transformation | Appendix transformation removes the old H1 and demotes its subsection headings; original body retained |
| Ignored output and live services | Deliberately not inspected or operated | No commands directed at these locations | No claim of content-level preservation across concurrent external activity |

That planning session did not open historical run artifacts. Execution used new
literal run IDs and retained current results separately. Appendix A remains
historical evidence; current passes do not establish its unresolved causes.

## 2. Current-State Repository Inventory

The declared inspection universe is the bounded source set below plus the 66
catalog JSON files listed at the end of this section. This is not a full recursive
repository inventory. Direct inspection means the relevant definitions and
call sites were read; it does not mean every line or transitive caller was audited.
Unless explicitly labeled generated, listed code and policy inputs are tracked
authored source. Catalog files were parsed for metadata; generated projections
were not treated as behavior owners.

| Paths / inspection depth | Responsibility, callers and dependencies | Ownership / risk / evidence |
| --- | --- | --- |
| [compiler.mjs](../../tools/harness/scheduler/work-graph/compiler.mjs), [go-lpt-shards.mjs](../../tools/harness/backend/go-lpt-shards.mjs), [go.mjs](../../tools/harness/execution/runners/go.mjs): directly inspected | Selected rows become compatibility groups and exact package/file invocations; depends on catalog, topology and runtime profiles | harness.command_surface / harness.test_catalog; selection and accounting risk; E-001 |
| [vitest-invocation-cli.mjs](../../tools/harness/execution/vitest-invocation-cli.mjs), [vitest-observations.mjs](../../tools/harness/diagnostics/vitest-observations.mjs), [row-results.mjs](../../tools/harness/execution/runners/row-results.mjs), [vitest-worker-lifecycle.mjs](../../tools/harness/execution/runners/vitest-worker-lifecycle.mjs): directly inspected | Coordinator, fresh file workers, report reconciliation and atomic row checkpoints; public Vitest API and secure publication dependencies | harness.evidence_accounting / harness.command_surface; failure attribution risk; E-002 |
| [services_linux.go](../../internal/testutil/harnessfixtures/services_linux.go), [renderer_linux.go](../../internal/testutil/harnessfixtures/renderer_linux.go), [pgtest.go](../../internal/testutil/pgtest/pgtest.go): directly inspected | Native broker service identity/retirement, renderer leases, clone and transaction fixtures; shared owners with per-borrow settlement | harness.browser; contamination and physical-settlement risk; E-003 |
| [browser-cohorts.mjs](../../tools/harness/browser/browser-cohorts.mjs), [browser.mjs](../../tools/harness/scheduler/work-graph/browser.mjs), [browser-functional-lanes.mjs](../../tools/harness/scheduler/work-graph/browser-functional-lanes.mjs): relevant definitions inspected in the audit | State-compatible packing, reset chains and allocation lifetimes; scheduler and broker consumers | harness.browser / harness.command_surface; reset ordering risk; E-004 |
| [fixtures.ts](../../apps/web/e2e/fixtures.ts), [Playwright config](../../apps/web/playwright.shared.config.ts), [Vite config](../../apps/web/vite.config.ts): directly inspected | Test-owned actors/contexts, one worker per cohort, remote renderer attachment; separate Node/jsdom projects | harness.browser; mutable identity and environment isolation; E-004 |
| [incident-administration.spec.ts](../../apps/web/e2e/incident-administration.spec.ts), [auth-and-incident-directory.spec.ts](../../apps/web/e2e/auth-and-incident-directory.spec.ts): named creation journeys directly inspected | Real creation overlaps; administration adds preference/shell assertions, directory adds listing/reopening assertions | module.incidents with affected UI owners; coverage mapping required; E-005 |
| [App.auth.test.tsx](../../apps/web/src/app/App.auth.test.tsx): selected assertions inspected in the audit | Seven catalog rows across module.auth and web.application; retained jsdom requirement | A starting partition candidate, not proof all seven rows are compatible; E-001/E-006 |
| [harness_work_graph_owner.json](../../tools/harness_work_graph_owner.json), [execution_topology_manifest.json](../../tools/execution_topology_manifest.json), [postgres fixture policy](../../tools/postgres_fixture_policy_registry.json): relevant metadata inspected | Cost estimates, resource claims, browser limits, explicit transaction admission and exclusive capabilities | Authored projections; amendments must follow governing owner; E-006/E-007 |
| [performance explanation schema](../../tools/schemas/cartulary.harness_performance_explanation.v5.schema.json): metadata inspected in the audit | Existing performance/queue/resource evidence surface; not a measurement collected here | harness.evidence_accounting; bounded observations can be incomplete; E-007 |
| [test operations skill](../../.agents/skills/cartulary-test-ops/SKILL.md), [command recipes](../../.agents/skills/cartulary-test-ops/references/command-recipes.md), [task surface owner](../../tools/task_surface_owner.json), generated Make target declarations: relevant procedures/declarations inspected | Future public-command selection; generated declarations are discovery evidence only | No Make command executed; E-008 |
| [remediation 01](testing-harness-remediation-01.md), prior contents of this handoff: relevant dispositions and complete prior handoff inspected | Prior implementation history and explicit exclusions | Historical evidence only; E-009/E-010 |
| Remaining test bodies, transitive effects, collector/adapter consumers and generated schemas affected by future protocol changes: pending | Required discovery boundary for S-001–S-005 | RB-001/RB-003; no exhaustive caller or equivalence claim |
| `.cartulary`, caches, installed dependencies, private runtime data, binaries, golden images and running processes: excluded | Outside the static documentation task; avoid interference | No traversal of external links/submodules or execution of repository code |

There are 2,365 active catalog rows across 66 files:

| Dimension | Static count |
| --- | --- |
| Runners | Go 1,000; Vitest 751; Playwright 586; shell 28 |
| Fixture capabilities | None 1,363; browser stack 586; dedicated PostgreSQL 288; transaction PostgreSQL 113; migration PostgreSQL 11; object namespace 4 |
| PostgreSQL scope | Exclusive 890; shared 113; none 1,362 |
| Browser state scope | Test-owned 465; isolated 88; measurement 33 |
| Go source/owner grouping | 188 package selectors, 225 package/owner pairs; 21 package selectors have multiple owners |
| Vitest source/owner grouping | 443 files, 475 file/owner pairs; 28 files have multiple owners |

These are catalog metadata counts, not executed assertions, work units, database
starts or browser launches. The 37 extra Go package/owner pairs and 32 extra
Vitest file/owner pairs identify fragmentation to investigate; they are not
counts of removable invocations. Runtime/evidence/fixture compatibility can
require additional partitions. Literal-selector inspection found no exact
identity duplicates in the earlier audit; it does not establish resolved alias
uniqueness or absence of semantic redundancy.

All 66 files under `tools/test_families/` received metadata-only accounting:
`app.migrate.json`, `app.operator.json`, `app.server.json`,
`harness.browser.json`, `harness.command_surface.json`,
`harness.evidence_accounting.json`, `harness.generated_artifacts.json`,
`harness.release.json`, `harness.test_catalog.json`, `module.artifacts.json`,
`module.assessments.json`, `module.auth.json`, `module.collaboration.json`,
`module.crossownertransaction.json`, `module.database_migrations.json`,
`module.entities.json`, `module.evidence.json`, `module.extensions.json`,
`module.graphprojection.json`, `module.imports.json`, `module.incidentbundles.json`,
`module.incidents.json`, `module.indicators.json`, `module.jobapi.json`,
`module.links.json`, `module.networkflow.json`, `module.parties.json`,
`module.projections.json`, `module.records.json`, `module.recovery.json`,
`module.reference_data.json`, `module.reportcomposition.json`,
`module.reporting.json`, `module.revisions.json`, `module.savedviews.json`,
`module.stagedobjects.json`, `module.tabularingest.json`,
`module.tasksdecisions.json`, `module.timeline.json`, `module.workbook.json`,
`package.grid_adapter.json`, `package.protocol_ts.json`, `package.test_utils.json`,
`package.ui.json`, `package.view_contracts.json`, `platform.audit.json`,
`platform.bootstrap.json`, `platform.config.json`, `platform.cryptography.json`,
`platform.httpruntime.json`, `platform.jobs.json`, `platform.objectstore.json`,
`platform.openapi.json`, `platform.pagination.json`, `platform.postgres.json`,
`platform.rootedfs.json`, `platform.telemetry.json`, `platform.viewquery.json`,
`platform.viewschema.json`, `platform.ws.json`, `web.application.json`,
`web.architecture.json`, `web.collaboration.json`, `web.design.json`,
`web.networkflow.json`, and `web.workbook.json`.

## 3. Module Boundary Diagnosis

Keep the existing execution coordinator, fixture broker and runner adapters.
The structural issue is deciding which compatibility boundaries are necessary
and preventing already completed work from being processed again. Evidence
ownership and physical execution identity are related but need not automatically
be the same partition; any separation must still satisfy adopted requirements.

| Responsibility | Current boundary and owner candidate | Disposition | Evidence / finding |
| --- | --- | --- | --- |
| Selected-row identity and evidence | Catalog and runner adapters; harness.test_catalog / harness.evidence_accounting | Keep; each row remains independently attributable even if execution is shared | E-001/E-002; F-001/F-002 |
| Execution compatibility | Graph compiler; harness.command_surface | Split physical compatibility from reporting-only distinctions only after owner review | E-001; F-001 |
| Service identity, admission and retirement | Native fixture broker; harness.browser | Keep; extend this owner if S-005 is admitted; no second pool or service manager | E-003; F-005 |
| Test-state isolation | pgtest and browser fixtures with test owners | Keep; change only proved candidate boundaries | E-003/E-004; F-004 |
| Report reconciliation/publication | Existing Vitest coordinator and adapters | Keep responsibility together; process deltas instead of rescanning completed work | E-002; F-002 |
| Test choreography | Incident test owner and existing public API helpers | Simplify setup; preserve distinct behavior assertions | E-005; F-003 |
| Cost and capacity model | Existing authored graph/topology inputs | Keep one deterministic model; recalibrate only from applicable measurements | E-007; F-006 |

### Infrastructure reuse is not test-state reuse

| Resource | Existing sharing to retain | Isolation that must survive |
| --- | --- | --- |
| PostgreSQL service and sealed migrated template | Explicit shared attachment for approved rollback consumers; templates initialized by physical owner | No implicit fallback startup; cluster-changing, lifecycle, migration and deadline evidence remains exclusive |
| Rollback database | Package bucket reused in a test process; `BeginRollbackDBT` serializes the full borrow | Fresh connection/transaction; every consequential helper stays on it; rollback/close failure poisons reuse |
| Dedicated/migration database | Compatible tests can share a Go process; dedicated databases are cloned, migration fixtures replay deliberately | Fresh test-owned database and eager teardown; real commit, cross-connection visibility and recovery tests retain their capability |
| Browser stack and renderer | Compatible cohorts share stack setup; healthy physical renderer service can serve later compatible borrowers | Fresh context and actor/session per test; one worker; state-scope proof; required resets and quarantine; no claim that all browser processes are shared |
| Object storage | Managed service with owned namespaces and existing cleanup | Confirm namespace emptiness; protect siblings; failed or uncertain cleanup prevents reuse |
| Vitest | One coordinator for compatible bounded cohorts; browser-unit is jsdom, not Chromium | Fresh isolated file worker and acknowledged termination before the next file; exclusive cases remain singleton |

No new framework, daemon, session shared across chats, mutable global fixture,
writable database pool, or shared browser context is proposed. Reuse stays bounded
to an authorized invocation and its existing ownership model.

## 4. Public Contract and Behavior Freeze Map

The cited harness clauses are adopted/current. Any proposed change to them remains
unadopted until reviewed at that owner. Production HTTP, WebSocket, persistence,
query/mutation, saved-view, projection, revision and authorization contracts are
preserved; this plan changes no product API or schema. S-003 must map the affected
product assertions to their owners before removing any step. UI selector changes
and golden updates are not part of that simplification.

| ID / surface | Observed and owner-required behavior | Disposition and prerequisites | Slices / validation |
| --- | --- | --- | --- |
| C-001 Selection and accounting | Amended REQ-379/805: exact symbols/titles, explicit physical compatibility, bounded cohorts, isolated files, native outcomes and attributable failures | Preserve exact owners/outcomes; owner-only partitioning removed after adoption and validation | S-001/S-002; V-001/V-002 |
| C-002 Resource ownership | REQ-803/804/811: explicit generation, admission/residency, immutable templates, completed cleanup; dedicated DB create/drop requires an exclusive cluster because DROP checkpoint work is cluster-wide | Preserve; S-005 needs adoption to extend non-browser ownership beyond one work unit. No weakened durability, timeout expansion or destructive retry | S-004/S-005/S-006; V-003/V-005 |
| C-003 State isolation | REQ-804/806/811: test-owned mutable state, transactional effect confinement, poisoned failed reuse, required browser resets; measurement quiet lifetime | Preserve. Role/configuration mutation, independent commits, concurrent transactions, migration and lifecycle assertions retain stronger isolation | S-003/S-004/S-005; V-003/V-004 |
| C-004 Evidence publication | REQ-805/806 and AC-087/088: independently settled outcomes, interruption behavior, teardown failures, bounded redacted diagnostics and canonical results | Preserve; inspect every consumer before changing checkpoint representation. If required, adopt/version one replacement protocol and migrate writers/readers together | S-001/S-002; V-002/V-005 |
| C-005 Product assurance | Distinct creation/admin, preferences/shell, directory/reopening, integration, visual and accessibility assertions remain meaningful | Preserve; no test deletion without an assertion-to-survivor map and appropriate product-owner review | S-003/S-004; V-004 |
| C-006 Deterministic scheduling | REQ-803/805/806: declared resource claims, stable packing, bounded concurrency, quiet measurements and single-attempt product assertions | Preserve; authored costs must have evidence provenance. No adaptive runtime-learning subsystem or automatic capacity escalation | S-006; V-001/V-005 |

Compatibility is not a default requirement for accidental implementation behavior.
Remove obsolete helpers, wrappers and protocol paths once their consumers migrate;
retain neither parallel implementations nor permanent legacy readers to make a
cutover appear harmless. Unsettled old runtime authority must be recovered through
its supported owner before an incompatible ownership-format cutover. Historical
artifacts are retained as history, not silently translated into current evidence.

## 5. Coupling and Boundary Findings

| ID / finding | Evidence and limitation | Classification / affected contract | Required action / trace |
| --- | --- | --- | --- |
| F-001 Owner-sensitive runner fragmentation | E-001/E-006: compiler grouping includes owner identity; some files/packages serve multiple owners. Other differences may still require separation | `should_fix` only where proved unnecessary; C-001 | S-001/V-001: compare complete compatibility tuples, preserve exact rows and measure fewer launches; do not infer savings from pair counts |
| F-002 Cumulative Vitest processing | E-002: every completed file appends observations/reports, calls `publish(false)`, normalizes/reconciles cumulative data and republishes all completed rows | `should_fix`; C-001/C-004. With similarly sized files, cumulative serialization alone grows quadratically in file count; elapsed impact is unknown | S-002/V-002: index observation identities, process new reports once and publish only new/changed rows; preserve final authoritative reconciliation |
| F-003 Repeated creation journey setup | E-005: both named incident tests use UI creation; administration adds preference/shell checks and directory adds list/reopen checks | `should_fix` for duplicated setup, not wholesale deletion; C-005 | S-003/V-004: retain creation in the directory journey and use public API setup for the preference/shell journey after complete coverage mapping |
| F-004 Over-broad isolation is not established for the remaining catalog | E-009: prior inventory explicitly excluded metadata browser rows because of shared directory/administration traversal; transaction eligibility requires complete effect tracing | `defer`; C-002/C-003/C-005 | S-004/V-003/V-004: reopen only with new source proof; otherwise retain the exclusion. Unique IDs or page-local routes alone do not prove isolation |
| F-005 Exclusive service lifetime can repeat physical initialization | E-003: ordinary exclusive cluster keys include UnitID and release retires the generation; REQ-804 currently requires this lifetime | `defer`; C-002/C-003. No redundant per-clone migration is alleged | S-005/V-003/V-005: measure physical starts/template cost, then consider sequential exclusive reuse at the existing broker after adoption |
| F-006 Estimates and concurrency need workload evidence | E-007: evidence-class estimates and fixed startup/cleanup estimates guide packing; browser default 4/max 6. Rollback buckets serialize within a process | `defer` until measured; C-006 | S-006/V-005: distinguish queueing, CPU, memory and storage pressure; increasing workers can increase checkpoint and startup contention |
| F-007 Existing structural improvements should not be rebuilt | E-003/E-004/E-009: templates, actor isolation, immutable prepared credential inputs, namespace cleanup, cohorts and renderer pooling already exist | `intentional/no_action`; C-002/C-003 | Preserve these foundations; no new slice without a demonstrated residual defect |
| F-008 Historical causal obligations remain open | E-010: W1 storage/drop and W2 readiness have evidence gaps despite reported successful later runs | `must_fix` only for closing those historical obligations, not a prerequisite to all independent optimization; C-002 | WF-07/V-005 and RB-005; do not label a performance optimization a historical root-cause repair without captured proof |
| F-009 Duplicate native-boundary assurance bypasses catalog contracts | S-008 observed the whole-package operation time out twice while its HTTPS child waited for host admission; the same five tests already belong to four CI-validated catalog rows | `must_fix` for current assurance closure; C-001/C-006 | Reopened S-006: select canonical rows, remove the private duplicate operation/alias, preserve existing profiles and validate exact coverage and aggregate completion |
| F-010 Startup recovery depends on success-only environment | Current failed child retains three managed suites whose environment files are missing; traced producer/broker publication ordering | `must_fix`; C-002/C-004 | S-009: durable prelaunch authority, separate bindings, fault-injection and recovery proof; no fabricated historical proof |

There is no evidence here for deleting entire coverage layers. A DOM test,
real-service integration test, accessibility interaction and visual capture can
exercise different failure modes despite sharing setup. Likewise, migration replay
is required evidence when migration behavior is the subject. Prior explicit
exclusions for real browser launches in renderer tests are not overturned by a
general desire to reuse infrastructure.

## 6. Refactor Workstreams

The prerequisite column is the authoritative workflow dependency list. The
approved execution order is WF-00 → WF-08 → WF-02 → WF-01 → WF-03 → WF-04 →
WF-05 → WF-06 → WF-07. Each completion includes its tracker checkpoint before
the successor begins. Service-heavy comparisons do not overlap.

| ID / class | Prerequisites | Goal / likely surface | Validation / checkpoint / binary exit |
| --- | --- | --- | --- |
| WF-00 / root | User implementation authorization, granted | Freeze contracts, characterization populations, baseline and metric availability | V-001/V-005; identities, observations, acceptance criteria and proposed owner amendments recorded |
| WF-08 / chain | WF-00 | S-007: include this controlling handoff in the existing Markdown lint selection | V-007; exact file included, lint passes and tracker checkpoint recorded |
| WF-02 / chain | WF-08 | S-002: incremental Vitest reconciliation and durable file checkpoints | V-002/V-005; outcome, privacy and interruption contracts preserved with linear evidence processing/publication |
| WF-01 / chain | WF-02 | S-001: separate physical compatibility from reporting ownership | V-001/V-002; exact attribution preserved with unnecessary launches removed, or necessary partitions retained with evidence |
| WF-03 / chain | WF-01 | S-003: remove redundant incident-creation choreography | V-004; surviving journeys independently preserve every distinct assertion |
| WF-04 / chain | WF-03; new evidence required to reopen an exclusion | S-004: candidate-specific isolation proof | V-003/V-004; admitted candidates verified or exclusions explicitly retained |
| WF-05 / chain | WF-04 terminal disposition; benefit, settlement and adoption gates for implementation | S-005: evaluate sequential exclusive PostgreSQL reuse | V-003/V-005; safety and material benefit established, otherwise explicit deferral/drop |
| WF-06 / chain | WF-05 terminal disposition | S-006: calibrate the resulting deterministic model | V-001/V-005; measured improvement without reliability/resource regression, otherwise supported no-change disposition |
| WF-09 / chain | Reopened WF-06 complete; F-010 demonstrated during final assurance | S-009: separate prelaunch recovery authority from success-only bindings | Failure/publication/recovery contracts; ordinary no-retry and exact authority preserved |
| WF-07 / chain | WF-06, WF-09 and every preceding terminal disposition | S-008: final-source validation and handoff completion | V-006/V-007; accepted changes and cleanup verified; obsolete paths removed; W1/W2 separately assigned |

A deferred/dropped optional workstream has no implementation dependency to satisfy;
its recorded disposition satisfies the corresponding terminal-disposition gate.
Mandatory correctness failures cannot be dropped as optional optimizations. If
measurements reject every candidate, close the optimization effort with a supported
no-change recommendation and retain unresolved historical investigations.

## 7. Implementation Slice Plan and Dispositions

Implementation is authorized. Slice entry remains conditional on the stated
evidence and owner-amendment gates. Optional isolation, service reuse and tuning
are not mandatory features; a supported deferral or drop satisfies their
terminal-disposition gate without weakening any correctness obligation.

| ID / kind | Finding / contract | Depends on / entry gates | Intended change and readiness | Validation / migration / binary completion |
| --- | --- | --- | --- | --- |
| S-001 / structural, possibly owner amendment | F-001; C-001/C-004 | WF-01; RB-001/RB-002/RB-004 | Start with `App.auth.test.tsx` selections; compare complete tuples, then combine only proved compatible exact selections while retaining row owners. Review Go package pairs afterward. **DONE**; compatibility, exact owner subsets and 749-row frontend execution verified | V-001/V-002. Update authored owners before generated projections. No old/new grouping flag or compatibility alias. Complete only with unchanged exact outcomes and measured reduction in launches/setup; otherwise retain partitions |
| S-002 / structural; protocol amendment only if necessary | F-002; C-001/C-004 | WF-02; RB-001/RB-004 | Index by project/file/title, reconcile each completed native report once, and publish new or semantically changed rows. Build final cumulative views once where the contract allows. **DONE**; immutable packets, disposition, shared reader and regression checks adopted; S-008 adds the final command-boundary edge | V-002/V-005. Preserve completed-file durability; global teardown can supersede affected rows. If removing cumulative checkpoint rewrites needs a new representation, migrate all consumers in one adopted cutover. Complete with equivalent failures/evidence and less processing/writing; do not claim a full improvement from indexes alone |
| S-003 / test restructuring | F-003; C-003/C-005 | WF-03; RB-001 | Keep landing-screen creation/list/reopen coverage in `auth-and-incident-directory.spec.ts`; seed the preference/shell case in `incident-administration.spec.ts` through an existing public API helper. **DONE**; product owners traced, coverage map retained and both survivors independently passed | V-004. Preserve preference confirmations, shell/grid visibility, page-error checks and creator/admin assurance at named surviving locations. Update authored title/routing together if purpose changes; no selector aliases. Complete when independently run survivors cover all prior obligations with less repeated setup |
| S-004 / candidate-specific isolation change | F-004; C-002/C-003/C-005 | WF-04; RB-003/RB-004 | Trace all connections, helpers, background effects and shared listings. Rollback is allowed only for transaction-confined behavior; browser sharing only for proved test-owned state. **DEFERRED after audit**; named metadata/concurrency exclusions remain justified | V-003/V-004. Preserve independently committed, concurrent, recovery and lifecycle tests. Revert a candidate's fixture declaration if benefit or independence fails, rather than adding a fallback mode. Complete each candidate with positive proof or explicit retained exclusion |
| S-005 / proposed lifetime behavior change | F-005; C-002/C-003 | WF-05; RB-001/RB-002/RB-004 | One run-owned broker lane may sequentially lease a compatible exclusive generation after full prior settlement, using fresh test databases and the sealed template. **DEFERRED after evaluation**; REQ-804's exclusive cluster lifetime remains unchanged and benefit/safety gates are unmet | V-003/V-005. No overlapping unrelated borrowers, cross-run daemon, pooled dirty DB, or delayed per-test drop. Role/configuration, migration, deadline, checkpoint and measurement cases remain dedicated. Complete only with adopted ownership/residency semantics, residue and failure proof, and net savings |
| S-006 / measured scheduling change | F-006; C-006 | WF-06; RB-004 | Calibrate the existing deterministic cost model and, only when justified, admission profiles for the post-refactor workload. **DONE after reopening during S-008**: canonical assurance routing now reuses catalog claims; global tuning remains unadmitted | V-001/V-005. Change one variable per comparison; no permanent tuning subsystem or automatic worker escalation. Complete with improved throughput/resource use and no meaningful reliability or tail-latency regression |
| S-007 / documentation validation | F-007; RB-006 | WF-08 after WF-00 | **DONE**: existing Markdown selection includes this exact tracker | Successful explicit lint receipt; no new linter or runtime dependency on Markdown |
| S-009 / startup recovery structural fix | F-010; C-002/C-004 | WF-09 after reopened S-006; before S-008 completion | **DONE**: prelaunch recovery authority separated from success-only bindings | Failure injection, ordinary no-retry, exact recovery, projection checks; missing historical proof remains unresolved |
| S-008 / final acceptance and handoff | All accepted changes | WF-07 after all terminal dispositions | **HANDOFF COMPLETE; ACCEPTANCE BLOCKED**: finalization, source checks and assurance completed; current readiness failure prevents successful acceptance | Final assurance 60/61 units passed, native exit 2; current cleanup/scans complete; W1/W2/W5 remain assigned; final Markdown receipt below |

For every slice, record a concrete acceptance threshold and experimental population
before runtime comparison. Require a material reduction in the targeted cost and
no unexplained regression in wall time, peak resource use or assurance. Drop an
optional optimization when its benefit is within observed variability or too small
to justify the additional architecture. Retain the simpler implementation.

## 8. Validation Plan

The implementation task authorizes the validation in this section. Commands run
from `/home/jochi/code/cartulary` through public Make targets.
Fresh execution uses a new literal `CARTULARY_TEST_RUN_ID`; record the exact command,
source/toolchain/capacity identities, native session and exit, result root and
physical settlement. Never reuse the historical run IDs in Appendix A.

| ID / stage | Cases and required evidence | Public command / selection route | Failure consequence |
| --- | --- | --- | --- |
| V-001 / pre-move and post-slice | Exact/partial selections; repeated titles across files/projects; multiple owners; incompatible environment/fixture/evidence classes; deterministic cohort identities and accounting; process-exclusive cases | `make task-guide OWNER=harness.command_surface`, then `make task-guide OWNER=harness.test_catalog`; freeze exact rows before `make test-slice OWNER=<owner> ROWS=<frozen-rows>` | Missing, duplicated or widened execution blocks S-001/S-006 |
| V-002 / pre-move and post-slice | Two or more completed files followed by interruption; suite load/file teardown failure; run-wide close failure; unhandled errors; contradictory/duplicate/missing native observations; redaction; worker termination uncertainty; completed-row preservation/supersession | `make task-guide OWNER=harness.evidence_accounting` and command-surface discovery; exact routed rows for the Vitest diagnostics/coordinator/adapter contracts, then representative compatible application cohorts | Incorrect attribution, lost completed evidence or false pass blocks S-001/S-002; no workaround that suppresses failure |
| V-003 / pre-move and post-slice | Isolated and reordered borrowers; connection escape and independent commits; concurrent borrow denial; cancellation; rollback/close/drop failure; poisoned reuse; lost owner; partial startup; cluster residue; sibling object namespace safety | `make task-guide OWNER=harness.browser`; candidate owner discovery followed by exact-row `make test-slice`; Go-only order comparisons use a recorded positive `GO_SHUFFLE_SEED` | Uncertain settlement keeps the resource quarantined and capacity owned; no second destructive cleanup attempt |
| V-004 / post-slice | Each surviving journey alone and in permitted combinations; creation/admin behavior, preference confirmations, shell/grid and page errors; preserved distinct visual/accessibility behavior for any changed browser fixture | `make task-guide OWNER=module.incidents`; freeze the two creation row IDs below; affected browser-owner slices. `make browser-e2e SUITE=visual` and accessibility selection only when fixture/visual coverage requires them | Coverage gap or order dependence blocks S-003/S-004; no blanket golden refresh |
| V-005 / baseline and comparison | Wall time, process/container CPU, peak resident memory, I/O, queueing, setup/assertion/cleanup/publication costs; physical starts, templates, clones/drops, renderer generations, resets, coordinator/file-worker counts; observation omissions | Matched exact-row `make test-slice ... HARNESS_DIAGNOSTICS=basic`; then `make explain-run RESULTS_DIR=<exact-root> DETAIL=performance JSON=1` and `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=evidence` | Missing dimensions remain unknown; reader success is not an execution pass or complete evidence audit. W1/W2 close only with separate causal evidence |
| V-006 / final authorized implementation | Affected owner regression, generated/source checks where inputs changed, aggregate and required explicit coverage, exact cleanup and current handoff | `make agent-finalize` before broader end-of-run checks; when projections change, `make generate`, `make source-check CHECK=generated` and `make source-check CHECK=schema`; `make ci` plus required explicit selections derived from changed owners | Failed acceptance remains open. Retained-run maintenance requires a compatible successful full warm `RESULTS_DIR`; otherwise report it skipped |
| V-007 / documentation closure | Re-read links, IDs, dependency graph, historical preservation, source limits and file boundary | `make lint-markdown` after final tracker edits; explicit handoff selection added in S-007 | Require a successful receipt for the completed handoff |

The S-003 starting rows are
`module.incidents.browser.real_browser_incident_create_bootstraps_the_crea_6b8248fece`
and
`module.incidents.browser.the_ordinary_authenticated_landing_screen_create_fa80fbfef5`.
S-001 starts with the seven catalog rows selecting `App.auth.test.tsx`, across
`module.auth` and `web.application`; compare their full tuples before proposing a
merged invocation. These are bounded starting populations, not preapproved merges.

`<owner>`, `<frozen-rows>` and `<exact-root>` are explicitly unresolved invocation
parameters, not runnable commands. RB-001 is resolved for admitted changes: the
exact diagnostics, scheduler and application selections and public discovery
results are recorded in Section 10. Optional S-004/S-005 selections must still be
resolved if their reopening gates are met.

Hold workload, fixture population, toolchain, cache mode, diagnostic mode, capacity
and storage environment constant within each comparison. Measure one treatment at
a time in an agreed window; no simultaneous service experiments or diagnostic
polling during quiet measurement. For this work, use cache-disabled runs to measure
executed work rather than cache hits, and record warm/cold dependency state.
Predeclare a bounded repeat count and report variability, failed runs and omissions;
do not retry until green. Compare critical-path elapsed time separately from summed
phase durations, which can overlap.

Existing typed diagnostics are preferred. If they cannot observe publication CPU,
bytes or lifecycle counts reliably, first scope bounded instrumentation at the
existing owner; do not infer missing metrics or build a second observer. An
`explain-run` equivalent-comparison request can reject changed-source identities;
retain its incompatibility reasons instead of forcing a delta. Cross-source
experiments need their own explicitly controlled interpretation. Descriptive cross-source observations are recorded separately from any release
or benchmark claim; no such qualifying performance claim is made here.

The following command posture is the **historical planning-session record**.
It is superseded for authorized execution by Section 10.

| Session command family | Side-effect assessment | Planning-session result |
| --- | --- | --- |
| `rg`, bounded `sed`/`cat`, `stat`, SHA-256; `git --no-optional-locks` status/revision/index inspection | File inspection only; no repo modules imported or services contacted | Executed for static evidence |
| Small Python standard-library JSON parsing and document transformation | Catalog metadata reads; only this admitted handoff is written | Executed; no repository code or tests run |
| `make lint-markdown` | Public target invokes the harness Markdown operation, writes a static-analysis cache stamp and can create run outputs; current configured globs omit this handoff | Deferred under the continuing no-execution/document-only boundary; not reported passed; lint configuration left unchanged |
| Tests, builds, `make generate`, `make format`, `make agent-finalize`, maintenance and service commands | Write outputs or operate infrastructure | Not run; not required for this document-only change; no retained warm `RESULTS_DIR` supplied |

## 9. Top-Level Work Tracker

New work uses the statuses below. Historical W/P/R identifiers and their original
status language are preserved only in the historical record and references to it;
they are not recast as new implementation completions.

| ID | Work item | Workflow / references | Status | Exit condition |
| --- | --- | --- | --- | --- |
| T-001 | Reorganize the current assessment and preserve historical evidence | Documentation task; E-001–E-011; V-007 | DONE | Static content and write-boundary comparison recorded; historical body preserved; skipped checks explicit |
| T-002 | Capture the bounded workload and metric baseline | WF-00; RB-001/RB-004 | DONE | WF-00 completion below records exact selections, identities, metric gaps and thresholds |
| T-003 | Review and reduce unnecessary runner partitioning | WF-01; S-001 | DONE | Cross-owner planner contracts, 749-row frontend suite, exact owner subset and evidence audit passed |
| T-004 | Simplify Vitest result processing | WF-02; S-002 | DONE | Immutable checkpoints, shared recovery, indexed reconciliation and stable rows validated; evidence below |
| T-005 | Consolidate repeated incident setup | WF-03; S-003 | DONE | Independent directory and preference cases, catalog migration and both evidence audits passed |
| T-006 | Reopen only evidenced isolation candidates | WF-04; S-004 | DEFERRED | Audit complete: existing exclusions retained with source reasons and explicit reopening criteria |
| T-007 | Evaluate sequential exclusive PostgreSQL reuse | WF-05; S-005 | DEFERRED | Evaluation complete; no admitted cross-work-unit reuse or cluster-lifetime amendment; reopening gates recorded |
| T-008 | Calibrate resulting cost/admission model | WF-06; S-006 | DONE | Duplicate whole-package routing removed; scheduler equivalence, four native rows, cleanup and evidence audit pass; no global tuning |
| T-009 | Complete authorized implementation acceptance and handoff | WF-07; S-008; V-006 | BLOCKED (handoff complete) | Implementation, finalization, source checks and validation/handoff performed; full assurance failed on W2 readiness, so successful aggregate acceptance is not claimed |
| T-010 | Preserve W1/W2 causal investigations | F-008; RB-005; historical record | TODO | Captured causes and owner repairs satisfy original closure criteria; passing optimization tests alone cannot close this item |
| T-012 | Preserve pre-ready recovery authority | WF-09; S-009; F-010 | DONE | Fault-injection, explicit recovery and all four real native-boundary rows pass; cleanup and evidence audits pass |
| T-013 | Settle the earlier runtime missing recovery proof | W5; F-010 migration limit; RB-007 | BLOCKED | Supported exact-owner recovery failed; validated authority and physical settlement required before releasing three resident allocations |
| T-011 | Include the controlling handoff in Markdown validation | WF-08; S-007; RB-006 | DONE | Explicit configuration selection and successful r02e-s007-markdown receipt |

## 10. Session Handoff Log

### Authorized implementation — 2026-10-20

The user requested implementation of the proposed plan. This entry supersedes
the earlier documentation-only authorization and workflow ordering, not the
historical results. The starting commit remains
`f9f94168816c9c14438c4efb1e2aa24ccdcdd710` on
`codex/specification-remediation-20261016`. The only initial working-tree change
was this handoff, SHA-256
`f56b013e6dbebcaa7ac5de06effe89a3d4e61a4c5d20fd321cec470aea28e1a5`.
Its original bytes are preserved at
`/tmp/cartulary-remediation02-execution-baseline/testing-harness-remediation-02.md`.
The logical index fingerprint was
`402244a8201275e56ffa7914cfc2c3c15943d686e65bb84431e83d90294cc6cd`.
Appendix A is preserved.

| Slice | Approved implementation, rationale and benefit | Compatibility and unresolved-risk treatment | Acceptance |
| --- | --- | --- | --- |
| S-007 | Add this handoff to the existing explicit Markdown lint list; avoid a new command or lint subsystem | Documentation selection only; no Markdown in product/runtime/generation evidence inputs | Exact file covered and `make lint-markdown` passes |
| S-002 | Index project/file/full-title observations; normalize/reconcile each settled file once; atomically retain immutable native-report/observation checkpoints and invocation disposition; publish only changed rows; assemble terminal envelopes once | Preserve terminal v1 report/v3 detail shapes and row IDs; one new intermediate protocol; remove cumulative checkpoint readers/writers in the same cutover | Interruption, partial publication, invalid identities, teardown, global failure, redaction and linear processing/publication checks |
| S-001 | Explicit physical-compatibility tuples replace reporting-owner partitioning, beginning with the seven App.auth rows, then other proven Vitest/Go groups | Exact row/verification owners survive; mixed Vitest execution uses harness ownership; work/cache identity changes, no aliases or grouping flags | Exact selections and failures attributed once; deterministic graph; real launch reduction; incompatible cases stay separate |
| S-003 | Keep real creation/admin/list/reopen in the directory journey; use authenticated public API setup for preference/shell journey | Preserve confirmations, shell/grid visibility, page errors and actor isolation; coordinate semantic row identity changes without aliases | Full assertion-to-survivor map and independent/combined executions; one redundant creation sequence removed |
| S-004 | Reopen only newly proved candidates; complete transaction/helper/background and browser shared-state tracing | Prior exclusions remain by default; no fallback isolation mode | Positive independence and benefit proof or named retained exclusion |
| S-005 | Only after benefit/safety admission, extend the existing broker to one borrower per run-local exclusive generation with fresh eagerly dropped databases | REQ-804 amendment before implementation; capacity retained through idle/cleanup; quarantine is irreversible; incompatible old ownership settled before cutover | Residue, concurrency, partial startup, owner loss and capacity tests plus at least 10% median elapsed improvement exceeding observed variability; otherwise defer/drop |
| S-006 | Calibrate existing deterministic costs only from the resulting workload; preserve ceilings by default | No adaptive tuner, second scheduler or worker increase to hide overhead | One variable per comparison; measured benefit without unexplained resource/cleanup-tail regression, otherwise no change |
| S-008 | Finalize source, verify affected owners and aggregate/required explicit coverage, settle resources, remove obsolete paths and complete handoff | Existing templates, actor/credential/object/renderer foundations retained; W1/W2 and W4 attribution limits remain separate | Actual native exits, canonical results, scans, cleanup, generated checks and final Markdown receipt recorded |

The concrete initial characterization routes are
`harness.command_surface.behavior.public_registry_parity` (including Vitest
diagnostics), `harness.command_surface.behavior.scheduler_contract`,
`harness.test_catalog.behavior.routing_contract`,
`harness.evidence_accounting.behavior.current_epoch_evidence`,
`harness.browser.unit.fixture_broker_smoke` and
`harness.browser.unit.pgtest_helper_contract`. Public task-guide discovery
revalidates these before execution.

For admitted runtime comparisons, predeclare three baseline and three treatment
executions with identical exact selections, dependency warmth, cache-disabled
policy, diagnostic mode, capacity and storage environment. Keep native sessions
and fresh run IDs. Retain failures and missing metrics; never retry until green.
S-001–S-003 require demonstrably removed redundant work plus no unexplained
regression. Added S-005 lifetime complexity requires at least 10% median elapsed
improvement exceeding observed variability. Investigate non-target regressions
above the larger of 5% and observed variability. Missing safety-critical
measurements block the affected optional optimization. Cross-source comparisons
are controlled experiments, not equivalent-source or Core 05 claims.

Final acceptance runs generation and affected source checks, then
`make agent-finalize` before broader final-source owner verification and `make ci`.
Explicit visual/accessibility/UI-review and measurement selections follow changed
ownership, not blanket reruns. Evidence audits use exact result roots. Retained-run
maintenance requires an eligible successful full warm root; otherwise its skip
is explicit. The last step is this tracker's completion update and Markdown lint.

### WF-00 completion

**DONE before WF-08 began.** `make help ALL=1` and task-guide for
`harness.command_surface`, `harness.evidence_accounting`,
`harness.test_catalog`, `harness.browser`, `package.view_contracts` and
`module.incidents` all exited 0. REQ-805/379 and AC-087 now state the approved
execution/reporting separation and incremental settled-file publication contract.
REQ-804's PostgreSQL lifetime remains unchanged. Human normative-fidelity review
remains distinct from executable projection checks.

All executions below used `CARTULARY_HARNESS_CACHE_MODE=off` and command-line
`HARNESS_DIAGNOSTICS=basic`. Each exact run root is
`/home/jochi/code/cartulary/.cartulary/test-results/<run-id>`.

| Run ID | Public command following environment prefix | Native session / exit | Graph duration / cleanup |
| --- | --- | --- | --- |
| r02e-baseline-contract | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity HARNESS_DIAGNOSTICS=basic` | 41414 / 0 | 59,018 ms; all 11 cleanup steps completed, including runtime closure |
| r02e-baseline-view-1 | `make test-slice OWNER=package.view_contracts HARNESS_DIAGNOSTICS=basic` | 48381 / 0 | 1,938 ms; all 11 cleanup steps completed |
| r02e-baseline-view-2 | same exact five-row owner selection | 49383 / 0 | 1,990 ms; all 11 cleanup steps completed |
| r02e-baseline-view-3 | same exact five-row owner selection | 99056 / 0 | 1,961 ms; all 11 cleanup steps completed |
| r02e-baseline-incidents | `make test-slice OWNER=module.incidents ROWS=module.incidents.browser.real_browser_incident_create_bootstraps_the_crea_6b8248fece,module.incidents.browser.the_ordinary_authenticated_landing_screen_create_fa80fbfef5 HARNESS_DIAGNOSTICS=basic` | 16435 / 0 | 79,253 ms; all 26 cleanup steps completed, including runtime closure |

The five view rows are `package.view_contracts.frontend_unit.contracts`,
`.facade`, `.references`, `.rows` and `.surfaces` under that same family prefix.
They use one five-file coordinator cohort. Median graph duration is 1,961 ms;
range is 52 ms (2.65% of the median). These are bounded characterization data,
not a benchmark claim. The incident execution characterizes preserved behavior;
no three-repeat browser performance comparison is claimed.

All five manifests have executable source
`sha256:c732e69ee9fe463ab88109e5748e6145c014af9cde40843c30ad82b09dac7ddb`,
system/capacity
`sha256:008fe637b622cd8932b55cc1d8be5346bae02c09ab35d7518fc9f72b66eb6c21`
and toolchain
`sha256:f0968e3ace0b7282a384ecb609ccc16db0e9e49c646cf0b45ba457ed4669ae6c`.
Public `harness-observability-check CHECK=evidence` on the contract and incident
roots returned 0. The performance reader on `r02e-baseline-view-2` returned 0;
its resource sampling covers only observed intervals and cannot establish
whole-invocation CPU/I/O totals or Windows-host storage behavior. Publication
bytes and reconciliation work need deterministic contract characterization in
S-002. These omissions do not justify a capacity change or PostgreSQL reuse.

The S-003 assertion map retains Core 01 REQ-01-163 creation/admin admission in
the real directory journey, alongside initial URL, key/title listing and reopen.
Core 01 workbook-preference authorization and Core 03 startup preferences remain
exercised by the administration journey's home/default confirmations, shell and
grid visibility. Its page-error collection and root-route assertion stay there.
Its authenticated `createIncident` helper uses the public operation and CSRF
headers; no storage shortcut or authorization bypass is introduced.

### WF-08 / S-007 completion

**DONE before WF-02 began.** Added this exact handoff to the existing explicit
handoff list in `.markdownlint-cli2.jsonc`. No command, product check or runtime
document dependency was added. `CARTULARY_TEST_RUN_ID=r02e-s007-markdown make
lint-markdown` completed in native session 41617 with exit 0. Its receipt is
`.cartulary/test-results/r02e-s007-markdown/adhoc/lint-markdown/tool-run-summary.json`.
The structural lint reports pass, 13,310 ms. RB-006 is resolved for this source;
the final tracker update will be linted again at handoff completion.

### WF-02 / S-002 completion

Pre-move characterization `r02e-s002-characterization` passed the exact
`harness.command_surface.behavior.public_registry_parity` selection with cache
off and basic diagnostics (native session `16304`, exit 0, graph 59,078 ms).
Two independently settled files survive interruption; invocation teardown
supersedes both. Source digest:
`sha256:ab17e03f02f8940cb72c4afeadd3c059b8b9da509c07c719c211fd657703529a`.

The initial post-change runs are retained, not erased:

| Run / native session | Result and diagnosis | Disposition |
| --- | --- | --- |
| `r02e-s002-contract-1` / `41777` | Exact public-registry row failed; native Make exit 2, normalized product/test_assertion_failure; graph 47,330 ms; all 11 cleanup steps completed | New timestamp test used a noncanonical fixture row ID. Corrected the fixture. An additional adapter optimization made during this run means it is not final-source acceptance evidence |
| `r02e-s002-contract-2` / `14054` | Same row failed; native exit 2, graph 61,970 ms; all 11 cleanup steps completed | New interruption inventory assertion read `name` instead of the existing `symbol_or_title` field. Retained summary proves the completed `ok` inventory entry survived; corrected the test |
| `r02e-s002-schema` / `53117` | `make source-check CHECK=schema` failed, native exit 2; all 12 cleanup steps completed | New schema attachments were out of lexical order. Sorted authored attachments and regenerated through Make |

`r02e-s002-contract-3` (session `67853`) passed all test units (graph
62,123 ms, all 11 cleanup steps completed), but the owning Make command exited
2 with artifact/artifact_error after summary publication. No retained scan
receipt was published. The new malformed-packet fixture recreated its file with
default permissions; the retained scanner requires mode 0600. The fixture now
creates that file with private permissions. This run is not accepted as an
overall pass. The corrected schema check `r02e-s002-schema-2` (session `6330`)
passed with native exit 0.

Run roots are `.cartulary/test-results/<run-id>`; the canonical target summaries
and referenced unit stderr retain the failure evidence. These failures are
related to this work and require successful replacement checks.

**DONE before WF-01 began.** The coordinator writes immutable file packets
and an atomic invocation disposition; the row adapter and output reporter share
one bounded, identity-validating reader. Native/normalized evidence is durable
before row publication. Only final settlement builds cumulative terminal
v1/v3 envelopes. Indexed project/file/title reconciliation replaces repeated
full-observation searches. Semantically unchanged rows retain their original
publication and timestamps. Global failures durably supersede provisional
success. Cumulative `runner-checkpoint-*` writers/readers are removed.

New packet/disposition schemas and their attachments were generated through
Make. Completed-file uniqueness remains enforced by the shared reconciler's
set check; removing the schema's quadratic object-array `uniqueItems` scan
preserves the semantic constraint without changing the v3 envelope shape.
Tests cover repeated titles, exact selection, skipped/missing/duplicate
observations, load/file/environment/global failures, SIGTERM/SIGKILL after two
settled files, partial terminal publication, unsafe/oversized/foreign packets,
redaction, row timestamp preservation/supersession, and 4/8/16-file publication
growth. One packet is written per file; terminal envelopes are each written once.

| Acceptance run | Native session / exit | Result |
| --- | --- | --- |
| `r02e-s002-contract-4` | `36119` / 0 | Exact public-registry row, cache off/basic; graph 62,226 ms; all contracts passed |
| `r02e-s002-view-1` | `40902` / 0 | Same five-row/five-file owner workload as baseline; graph 1,904 ms |
| `r02e-s002-view-2` | `4686` / 0 | Same selection; graph 1,847 ms |
| `r02e-s002-view-3` | `81812` / 0 | Same selection; graph 1,882 ms |

All four runs share source
`sha256:a912ec73c00e302d927ff3d82f11de24eedde797d4f6ac3275ab6c2e220c22e9`,
the baseline toolchain, private retained-scan PASS receipts and 11 completed
cleanup steps including runtime close. Public `CHECK=evidence` audits of the
contract run and view-2 passed with native exit 0 (`full_validity=0`; these are
evidence audits, not full resource-sampling claims). `make generate` and the
corrected schema check passed.

Treatment median 1,882 ms versus baseline 1,961 ms is a descriptive 4.0%
reduction; the structural acceptance is removal of cumulative writes, not a
broad throughput claim. Same coordinator/worker count is expected in this slice.
The final adapter's measured parse activity rose from 3.245 to 5.290 ms because
it now validates binding, normalized causes and committed-envelope digests;
this approximately 2 ms bounded integrity cost does not overturn the removed
cumulative path or the non-regressing whole-cohort result. Short-lived process
sampling still cannot establish whole-run CPU/I/O totals or Windows-host cost.
The public performance report is retained at
`/tmp/cartulary-remediation02-execution-baseline/s002-performance.json`.

### WF-01 / S-001 completion

The compiler now excludes reporting ownership from otherwise unchanged physical
compatibility contracts. Mixed Vitest units use `harness.command_surface`;
Go units retain `harness.backend`. Exact row identities and owner selection are
unchanged. Scheduler contracts `r02e-s001-scheduler` (session `91751`, native exit
0, graph 12,541 ms) demonstrate App.auth's seven cross-owner rows combining from
two owner invocations into one, and a compatible reporting/networkflow Go pair
combining from two into one. Negative project/resource/preparation/isolation
cases remain separate. The six-row App.auth `module.auth` subset passed as
`r02e-s001-auth-subset` (session `9751`, native exit 0, graph 5,206 ms).

**DONE before WF-03 began.** `make test-slice SUITE=frontend-unit` with cache
off/basic completed as `r02e-s001-frontend` (session `55732`, native exit 0,
graph 207,914 ms). All 749 rows passed in 11 Vitest cohorts plus preparation;
all seven App.auth rows belonged to one physical unit owned by
`harness.command_surface` and each has its own passed row result. All 23 cleanup
steps completed, the retained-secret scan passed, and public `CHECK=evidence`
audit passed (exit 0, `full_validity=0`). Source:
`sha256:80d50332e52557c3f13061da0546d9cd791318f3929f3c43a86aa5ef157a95bf`.
No same-selection full frontend baseline exists, so this establishes functional
acceptance and deterministic launch consolidation, not a suite latency gain.
The public performance report is retained as `s001-performance.json` alongside
the S-002 report. Broader final validation will exercise mixed Go execution.
Work-unit/cache identities change naturally with compatibility and membership;
row IDs and verification ownership do not require migration or aliases.

S-008 follow-up from source review: exercise interruption after the final file
settles and retain a private command failure even if the durable disposition
has no command failure. The enclosing command must not become successful merely
because independently settled rows passed. This edge is not closed by the
current interruption tests, which leave an unfinished selected file.

### WF-03 / S-003 completion and assertion map

Core 01 REQ-01-163 owns atomic creator/admin and initial preference creation;
REQ-01-148–151 own preference resources/authorization; Core 03 REQ-03-027–031
own startup selection. These product requirements are unchanged.

| Prior obligation | Surviving independent browser case |
| --- | --- |
| Real landing-screen create, successful public POST, returned incident identity, creator admin role, incident key, directory listing and reopen | Unchanged `auth-and-incident-directory.spec.ts` case `creates an incident from the landing screen, lists it, and opens the workbook as incident admin` |
| Home/default action confirmations, preference controls, shell composition, root workbook URL, no page errors, ready shell and visible Timeline grid | `incident-administration.spec.ts` case `confirms workbook home and incident default preferences in a ready workbook shell`, using existing authenticated public `createIncident` setup |

The new semantic row is
`module.incidents.browser.workbook_home_and_incident_default_preferences_c_6591154e0f`,
derived through `make author-test-row-id`; the old creation-named row is retired
without an alias. The opaque scenario identity remains stable for the evolved
case; the exact authored title and row binding change together. No product code,
fixture lifecycle, isolation scope, timeout, or verification owner changes.
Generated browser grouping is refreshed through Make.

**DONE before WF-04 began.** The independently selected survivors passed:
`r02e-s003-preferences` (session `91887`, exit 0, graph 89,064 ms) and
`r02e-s003-directory` (session `31777`, exit 0, graph 88,934 ms). Each used its
own exact-row Make invocation with cache off/basic; both used source
`sha256:9be1088ffadd06ed8731d65458fa93c23e6287005904053e35f3f3aa6330293e`.
Each recorded 23 completed cleanup steps, retained-secret-scan PASS and a public
`CHECK=evidence` PASS (exit 0, `full_validity=0`). Catalog check
`r02e-s003-catalog` (session `34495`) and generation
`r02e-s003-generate` (session `29255`) also exited 0. No selectors were aliased
and no assertions were removed without a named survivor. These independent
runs overlapped and are not matched latency comparisons with the paired
baseline; the demonstrated saving is one duplicated UI-creation setup path.

### WF-04 / S-004 disposition

**Audit complete; optimization DEFERRED before WF-05 began.** No fixture or
isolation metadata changed. The remaining candidates do not have new proof
that permits a weaker boundary:

- Incident metadata cases and their `incidentMetadata.ts` /
  `incidentMembershipManagement.ts` helpers still enter through the shared
  directory and cover membership, departure, role, closure or authorization
  state. Their explicit remediation-01 exclusions remain valid. S-003 removes
  duplicated setup but does not establish independence of all observed state.
- `mutation_admission_lock_integration_test.go` uses three independent
  connections, two concurrent transactions, committed setup and a timed
  lifecycle writer. One rollback transaction would invalidate its assertion.
- `membership_concurrency_integration_test.go` uses a real pool and concurrent
  independently committed cross-admin deletion. Its dedicated database is
  required evidence, not redundant setup.

Rationale/benefit: retain faithful concurrency and workflow tests without adding
fixture adapters that obscure their effects. Compatibility impact is none.
The residual cost is dedicated setup; weakening these declarations now risks
false assurance, cross-test interference and cluster checkpoint contention.
Reopen only for a named candidate with a complete connection/background-effect
trace, assertions confined to test-owned state, order/parallel independence
proof and a matched measurable benefit. Migration, role/configuration, lifecycle,
measurement and independently committed concurrency cases remain excluded.
No runtime check was added for this no-code disposition; final validation covers
the unchanged declarations with the implementation changes.

### WF-05 / S-005 disposition

**Evaluation complete; extension DEFERRED before WF-06 began.** Existing
`services_linux.go::clusterKey` binds ordinary exclusive PostgreSQL to UnitID,
and `releaseCluster` retires it at that lifetime's end. This implements REQ-804;
it is not an accidental compatibility restriction. Shared rollback consumers,
sealed templates, browser allocation reuse and measured-fixture ownership
already have distinct supported lifetimes. Existing residency proof survives
until settlement and must not be bypassed.

The available paired incident baseline exposes two released browser leases
(8,164 and 10,091 ms held, via public `DETAIL=fixtures`, exit 0). This does not
measure amortizable ordinary non-browser physical initialization, nor prove safe
idle residency between exclusive work units. No matched three-run ordinary
PostgreSQL treatment or complete host-storage attribution exists. The predeclared
10% median-benefit and safety gates therefore do not admit the extension.

Recommendation: retain the existing broker and one-exclusive-unit lifetime;
add no pool, daemon, compatibility flag or delayed database drop. There is no
specification, implementation or migration change for this slice. This avoids
unproved residue, role/configuration leakage, idle admission under-accounting
and cancellation/recovery coupling. The unresolved cost is potentially repeated
physical startup; no per-clone migration redundancy is claimed.

Reopening requires a named non-lifecycle workload, matched physical-start /
template / clone / eager-drop / retirement evidence, median benefit above both
10% and baseline variability, one borrower at a time, fresh databases, complete
idle residency accounting, quarantine after failed settlement and fault-injected
owner-loss/recovery proof. Adopt REQ-804's new lifetime before implementing it in
the existing broker. Migration, role/configuration, deadline/checkpoint and quiet
measurement fixtures remain excluded. W1/W2 causes are not closed by these passes.

### WF-06 / S-006 disposition

**Review complete; no policy change before WF-07 began.** The deterministic
model already charges startup/cleanup once per formed cohort. Owner removal now
changes the actual cohort membership, so its summed estimates reflect sharing
without introducing another policy source. Capacity, worker counts, deadlines,
resource profiles and estimate inputs are unchanged.

The public S-001 report shows the largest 383-row cohort estimated at 384,250 ms
and executed in 207,608 ms; a 282-row cohort estimated at 283,250 ms executed in
51,970 ms. The renderer build estimated 15,000 ms executed in 41,461 ms, and one
renderer cohort estimated 6,250 ms executed in 19,495 ms. These are observations
from different workload classes, not evidence for one global multiplier.
The largest cohort had no admission queue; whole-run CPU reservation saturation
was 3,569 ms. Raising capacity would not address that serial critical path.

Resource collection has 1,240 samples and no discovery truncation, but covers
observed subsets rather than whole-run consumption; the bounded report omits
185 process scopes and some attributed CPU lower bounds are zero. Its sampled
simultaneous RSS sum of 8,522,039,296 bytes is not a per-unit peak or an additive
reservation requirement. Windows-host/storage attribution remains absent. The
full suite has no matched before/after repetition and overlapped a separate
owner-subset verification. S-002's three matched small-cohort treatments showed
no whole-cohort regression, but cannot calibrate renderer or aggregate capacity.

Rationale/benefit: preserve predictable admission and avoid speculative cost or
worker changes. Compatibility impact is none. Residual risks are coarse estimates
and a long compatible-cohort tail; neither was hidden by a claimed throughput
improvement. Reopen calibration with predetermined representative, uncontended,
matched repetitions, useful per-unit resource attribution, and one estimate or
cohort-packing variable at a time. Validate latency/tail, queueing, memory, I/O,
cleanup and exact outcomes; do not introduce an automatic tuner or increase
workers to mask storage or lifecycle costs.

### WF-07 / S-008 validation and handoff — acceptance blocked

Final source review found and closed the command-boundary edge noted in WF-01:
settled rows survive interruption after the last file, but the enclosing
invocation now retains failure; a private command failure is not overwritten by
a null durable disposition. Zero exit without final settlement fails closed.
The specification makes this separation explicit. The added real-process
regression kills the coordinator during global close after both files settle.
`r02e-s008-edge-contract` (session `93845`, native exit 0, graph 62,513 ms) passed
that edge together with all existing diagnostics contracts. `make format`
passed as `r02e-s008-format` (session `89580`, exit 0).

The first finalizer `r02e-s008-finalize` (session `8162`) exited 2 at
`json-shape-check`, before mutation. Its
`unit-artifacts/finalize-summary.json` reports stale topology inputs for the
final row-runner/adapter change. This is related to S-008; no product test failed.
All 10 cleanup steps completed. `make generate` then passed as `r02e-s008-generate` (session `34203`,
exit 0). The fresh finalizer `r02e-s008-finalize-2` (session `80754`, exit 0,
graph 25,201 ms) passed with generated outputs unchanged. Its public
`unit-artifacts/finalize-summary.json` explicitly records retained-run maintenance
as skipped because `RESULTS_DIR` was unset.

Final selection discovery confirms `make ci` includes all command-surface,
evidence-accounting and catalog owner rows, both incident survivors, and the
mixed reporting/networkflow Go pair (2,271 rows / 657 work units before final
refresh). These owners need no duplicate post-CI invocation if that final-source
aggregate succeeds. `make source-check CHECK=all` passed as
`r02e-s008-source` (session `64671`, exit 0, graph 15,971 ms), after finalization.
The source digest for this CI acceptance run is
`sha256:8a62033d2cb106b1ff4b6b0337a57b0051bddc4dbc48b11a8b3bf958e9a1de5e`
(the later explicit-only fixture correction has its own digest below).
Cache-disabled `make ci HARNESS_DIAGNOSTICS=basic` passed as
`r02e-s008-ci` (session `67529`, native exit 0, graph 3,069,649 ms):
657/657 work units passed, with no skipped, failed or cancelled units. All 1,291
cleanup steps completed without failures or blocks. This is correctness and
settlement evidence, not a matched performance comparison. All 2,271 selected row
results are passed, including both incident journeys and the mixed-owner Go
cases. `retained-secret-scan.json` reports PASS over 10,738 retained files.
`make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/r02e-s008-ci CHECK=evidence`
passed in session `48172` with native exit 0 (`full_validity=0`, read-only audit). At that finalization, no successful full warm check was available for these
changes; retained-run maintenance was explicitly skipped.

Explicit-coverage discovery compared both row and operation identities. All
artifact-consumer units are included in CI. The full harness contract selection
has 52 additional operation units, including wrapper and scheduler assurance;
`CARTULARY_TEST_RUN_ID=r02e-s008-contracts CARTULARY_HARNESS_CACHE_MODE=off
make harness-contract SUITE=full HARNESS_DIAGNOSTICS=basic` ran in
session `63216`, started after the CI process exited; its two failures and
replacement validation are recorded below. Rendered visual and
UI-review selections are not required: no product rendering or browser fixture
semantics changed.

The additional cache-disabled full contract run hit a `timing/timeout_failure`
at `operation:harness.harness_runtime_boundary_contract` (SIGTERM, no assertion
report; unit stdout/stderr empty). At inspection its HTTPS child had completed
10 of 12 units and was waiting for host admission; the preceding PostgreSQL
child passed all 7 units and completed all 18 cleanup steps. Concurrent real
`check` output assertions were actively executing their 339-unit child graphs.
These observations establish contention and the outer deadline, not a product
failure or a complete causal model. The full run finished with native exit 2 (session `63216`, graph 1,194,498 ms),
56 passed units and 2 failed units; all 67 cleanup steps completed and the retained
scan passed over 12,799 files. The second failure was
`harness.assurance_web_e2e_lifecycle`: its live-owner security fixture omitted the
now-required `runtime-root/tls/trust.pem`, so `prepare_runtime_root` returned 2
before its positive admission assertions. This fixture was unchanged by S-001–S-003;
repair it using the existing PKI fixture tool without relaxing production trust.
The unchanged runtime boundary passed alone as `r02e-s008-runtime-isolated`
(session `94600`, native exit 0, cache off/basic, graph 255,903 ms). This supports
an aggregate-contention explanation for the deadline failure; it does not prove a
general scheduling guarantee. No timeout or capacity policy was changed.

S-008's fixture correction uses the existing `tools/browserpki` tool to provision
private trust before the positive runtime-admission assertion, matching the
existing artifact fixture. Production trust and ownership guards are unchanged.
`r02e-s008-lifecycle-repair` (session `87706`, native exit 2, graph 7,431 ms)
failed because the new setup omitted the PKI tool's required private parent
directory. All 13 cleanup steps completed. The fixture now creates that parent
with mode 0700 before provisioning; the fresh focused verification is
`r02e-s008-lifecycle-repair-2` (session `24121`, native exit 0, graph 7,428 ms).
All four focused units passed. The isolated runtime evidence audit also passed
(`CHECK=evidence`, native exit 0, `full_validity=0`), with all 10 parent cleanup
steps completed.
After the focused fixture pass, generation/finalization and full assurance were
repeated in normal cache mode: command-output contracts still execute, while
unchanged child product checks may use the existing cache. The separate CI run
already executed every selected product row with caching disabled.

The fixture correction passed `make generate` as `r02e-s008-fixture-generate`
(session `6187`, native exit 0), then `make agent-finalize` as
`r02e-s008-fixture-finalize` (session `36534`, exit 0, graph 25,473 ms), with
outputs unchanged and retained-run maintenance again explicitly skipped because
`RESULTS_DIR` was unset. `make source-check CHECK=all` passed as
`r02e-s008-fixture-source` (session `13502`, exit 0, graph 16,174 ms).
Current executable source:
`sha256:ccdddc0c4bdcb093e9b18214256cb3267727e0d84d45f70126213604ad2ff5ed`.
The pre/post fixture-correction CI plans retain identical row selections, unit IDs
and unit semantic digests; the correction belongs to explicit lifecycle assurance,
not a CI execution input. Preserve the distinct source identities rather than
claiming CI executed after that fixture-only edit.

`CARTULARY_TEST_RUN_ID=r02e-s008-contracts-normal make harness-contract SUITE=full
HARNESS_DIAGNOSTICS=basic` finished in session `53128`, using the standard normal
cache policy: 55 units passed, three failed, graph 1,232,508 ms, and retained scan
passed over 12,401 files. The final native-session response was lost at context
compaction; its exit code is unavailable, so the failed summary is not substituted
for an observed native exit. A diagnostic `target-plan` invocation with direct cache/diagnostic
arguments was rejected as undeclared usage (exit 2); the supported `INPUTS` route
and default-plan comparison provide the selection evidence.

#### WF-06 / S-006 reopened before further implementation

The normal-cache full assurance run reproduced the same runtime-boundary timeout.
Static tracing identifies a second execution path for the exact same five native
boundary tests: `harness.harness_runtime_boundary_contract` runs the whole Go
package with fallback one-CPU/one-process claims, while four existing catalog rows
already own these tests with the appropriate ordinary, installed-browser and
exclusive-process contracts. All four rows passed in CI. The duplicate operation
bypasses those declarations and aggregates their deadlines, setup and attribution.

Remediation belongs to authored verification routing, scheduler contract tests,
generated projections and this handoff. Replace the duplicate operation and its
private alias selection with the four canonical row references; retain the public
`harness-contract SUITE=runtime-boundary` selection. This restores one execution
contract per test, avoids a second profile/timeout policy, and lets full assurance
deduplicate and schedule the same rows used by CI. Existing capacity, fixture
isolation and deadlines stay authoritative. Retire the private operation without
an alias; retained failures remain historical evidence. Leaving it unresolved
permits under-reserved nested execution, aggregate starvation and divergent test
coverage. Validate exact five-symbol coverage, canonical row/unit equivalence,
non-widening selection, the scheduler regression and full assurance completion.

The second run also failed its real `check` output assertion because the nested
`module.recovery` unit hit a forced DROP cleanup deadline: generation
`7698846237154906152`, admission 5 ms, operation 20,000 ms against 19,999 ms,
`CheckpointDone` / checkpointer `DataFileSync`, with requested/done counts 17/16.
The public nested root is
`.cartulary/test-results/r02e-s008-contracts-normal/unit-artifacts/operation-harness.assurance_output_check/real-command-output/runs/real-check`.
All 449 nested cleanup-ledger steps completed, but the unit retained its primary
cleanup failure and subsequent capability rejections. This is new W1 evidence;
it does not establish the host/storage cause or justify weaker cleanup semantics.

The machine-output assertion also failed its nested `real-check-machine` run
(334/339 units passed; graph 458,042 ms; `infra/service_readiness_timeout`). Its
401 cleanup steps include five failures: three host releases, `services_close`
with `resource_graph/io/errno=2`, and `runtime_close` with
`recovery_proof/io/errno=2`. Exact root:
`.cartulary/test-results/r02e-s008-contracts-normal/unit-artifacts/operation-harness.assurance_output_check_machine/real-command-output/runs/real-check-machine`.
Owner-checked recovery ran once as `r02e-s008-nested-recovery`, session `62014`,
native exit 2, and failed closed at `resource_graph/io/errno=2`. The exact binding
was runtime root
`/home/jochi/.cache/cartulary/harness-scratch/suite-runtime/suite-real-check-machine-c0149ef6-b5f5-4b40-bee5-c7c53ebfab6d`
and lease `ce924c85-30f5-4e77-8b89-2c79756a04ac`. No authority files were recreated,
resources force-deleted, or failures rewritten. Physical settlement of this child
remains unresolved and requires the harness runtime/service owner to establish
why the required recovery dependency is absent and recover through validated authority.
The parent cleanup ledger does not establish child settlement. Two diagnostic
reader attempts (`DETAIL=failures JSON=1`, then `DETAIL=summary JSON=1`) were
rejected as usage errors (exit 2); supported `DETAIL=summary` succeeded without
rerunning the child.

REQ-102 now explicitly requires native-boundary assurance to reuse the catalog
contracts. The four canonical rows cover all five package tests, including the
separately isolated borrowed fixture and the installed-browser fixture. Their
existing profiles and isolation distinctions must remain intact.

S-008 remained open while this reopened workstream executed. The normal-cache
run was terminal before executable inputs changed; the S-006 completion checkpoint
was recorded before the subsequent workstream. Unresolved child recovery is an explicit acceptance
limitation and must not be hidden by a later passing selection.

#### WF-06 completion checkpoint, before WF-09 begins

S-006 is complete. `make generate` passed as `r02e-s006-routing-generate`
(session `22296`, exit 0). The focused scheduler row passed as
`r02e-s006-routing-scheduler` (session `42748`, exit 0; graph 12,580 ms).
`CARTULARY_TEST_RUN_ID=r02e-s006-routing-boundary CARTULARY_HARNESS_CACHE_MODE=off
make harness-contract SUITE=runtime-boundary HARNESS_DIAGNOSTICS=basic` passed
(session `43377`, exit 0; graph 154,941 ms): four canonical rows/units passed,
all 13 cleanup steps completed, retained scan passed over 378 files, and public
`CHECK=evidence` passed with exit 0. Source digest:
`sha256:08ed8fe3d714ec7190994ed1dde4c4d27458d7cafebcc1c636bf11b3b4572bd5`.
The regression verifies exact five-symbol coverage and equality of the direct,
execution, regression and full-assurance boundary units, including their existing
resource/isolation/deadline projections. No broad throughput guarantee is claimed.

#### WF-09 / S-009: durable startup recovery authority

**DONE; entered after the S-006 completion checkpoint above.** Bounded
inspection of the exact failed child's six resource records found three acquired
managed suites and their three resident allocations. All three suites lack their
`suite-environment.json`; the journals and resident records remain present.
The parties allocation's public scope shows `sql_connectivity/deadline_expired`,
15,001 ms against 15,000 ms, with unresolved cleanup. This narrows W2's current
failure stage without identifying its service/environment cause.

F-010 is a demonstrated implementation gap: `StartSuite` records the environment
file as a recovery dependency before invoking the helper, but the helper writes
that file only after readiness. Failed startup or owner loss can therefore retain
physical authority without the environment required by recovery. The positive
mock always wrote the file and missed this gap.

Remediation: the broker publishes its bounded, allowlisted recovery environment
before recording launch intent. The helper writes success-only service bindings
to a separate disposable private file; the broker validates those bindings and
atomically updates recovery authority before admitting consumers. Recovery stays
fail-closed for missing, corrupt or foreign proof. This belongs to the normative
clarification, implementation, fault-injection tests and handoff; no product API,
capacity, deadline, new service manager or old-proof fallback is introduced.
The long-term benefit is explicit ownership of pre-ready recovery independently
of readiness data. Leaving the gap permits unrecoverable private runtimes and
stranded residency after failed acquisition.

Validate absence of ready bindings after failed startup, durable minimal proof at
launch, preservation and exact recovery after one unresolved ordinary cleanup,
no duplicate ordinary cleanup, successful ready binding validation, and private
publication failure before launch. Old runs already missing proof remain an
explicit migration limitation; never fabricate their environment files. Complete
this workstream and update the tracker before returning to S-008 finalization,
source checks, aggregate assurance and handoff.

#### WF-09 completion checkpoint, before S-008 resumes

The broker now durably publishes its allowlisted recovery environment before
recording launch intent, uses private disposable `suite-bindings-*` output for
ready data, and removes that output after validation. Successful bindings update
recovery authority atomically. No reader fallback or historical proof mutation
was added. All earlier foreign/missing/unsafe evidence checks remain active.

`make format` passed as `r02e-s009-format` (session `79817`, exit 0), and
`make generate` passed as `r02e-s009-generate` (session `79335`, exit 0).
Cache-disabled focused verification used `HARNESS_DIAGNOSTICS=basic`:

| Run | Public selection | Native session / exit | Outcome |
| --- | --- | --- | --- |
| `r02e-s009-fixtures` | `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_fixtures` | 26430 / 0 | One selected row passed, including launch-publication rejection, absent/partial bindings, preservation, no repeated ordinary cleanup and exact explicit recovery; graph 3,003 ms; 10 cleanup steps completed; retained scan passed over 17 files |
| `r02e-s009-boundary` | `make harness-contract SUITE=runtime-boundary` | 45755 / 0 | Four real native-boundary rows passed; graph 164,587 ms; 13 cleanup steps completed; retained scan passed over 378 files |

Both exact roots passed public `CHECK=evidence` with native exit 0. Source digest:
`sha256:1fd7ba2a61e318324c2df8b9bc4f4805806f6d10d3a446df93292365cba7b21a`.
S-009 is complete and the tracker is updated before resuming S-008. The old
missing-proof runtime remains W5; these results establish the new publication
contract, not settlement of that old runtime or the cause of its SQL deadline.

#### S-008 final-source validation resumed

`make agent-finalize` passed as `r02e-s008-final-finalize` (session `78930`,
native exit 0; graph 25,577 ms). Generated outputs were unchanged. Retained-run
maintenance was skipped because `RESULTS_DIR` was unset; no eligible successful
full warm `check` root was available. `make source-check CHECK=all` passed as
`r02e-s008-final-source` (session `19299`, native exit 0; graph 16,215 ms).
Both retain the S-009 source digest above. Final broad assurance is
`CARTULARY_TEST_RUN_ID=r02e-s008-final-contracts make harness-contract SUITE=full
HARNESS_DIAGNOSTICS=basic`, session `36725`, normal cache policy. Its final failed
outcome is recorded below; earlier full-assurance failures are preserved rather
than replaced.

The earlier 2,271-row cache-disabled CI proves the S-001/S-002/S-003 and final
Vitest edge changes at its recorded source identity. Later changes affect the
explicit lifecycle fixture, canonical assurance routing and broker startup
recovery. Focused contracts and real native-boundary runs cover those changes;
final full assurance also exercises real nested `check` commands. Do not relabel
the earlier CI as a run of the final broker source. Product visual/measurement
behavior is unchanged, so visual/UI-review and measurement selections are not
added solely for these harness changes.

During final assurance, the machine-output child reported
`infra/service_readiness_timeout` for `go:08fcce5155edfae4:ed1de841f34f`, the eight
Network Flow integration rows. The exact public child root is
`.cartulary/test-results/r02e-s008-final-contracts/unit-artifacts/operation-harness.assurance_output_check_machine/real-command-output/runs/real-check-machine`.
Its scope `_shared/test-services/23f113d6c2b14cd58d26954a/service-scope.json`
binds allocation `postgres:unit:go:08fcce5155edfae4:ed1de841f34f` and records
`sql_connectivity/deadline_expired`: 15,001 ms against 15,000 ms, 60 probes,
after successful container start (660 ms) and endpoint resolution (7 ms).
Physical settlement is **completed**, with no implicit startup retry. This is
current W2 evidence and a failed acceptance gate, not a product assertion or
proof of the service/environment cause. The parent and independent work were
allowed to drain; no deadline, capacity or rerun-until-green change is authorized
by this observation.

Both nested checks have now published terminal summaries and retained scans:

| Final child | Reported outcome | Work units / cache | Physical closure |
| --- | --- | --- | --- |
| `real-check` | PASS; graph 1,154,899 ms | 339 passed; 157 cache hits, 20 misses, 162 bypasses | All 431 cleanup steps completed; scan passed over 3,787 files |
| `real-check-machine` | FAIL; graph 1,113,148 ms; readiness failure above | 338 passed, one failed; 147 cache hits, 30 misses, 162 bypasses | All 439 cleanup steps completed; scan passed over 3,791 files |

The successful child's public `CHECK=evidence` passed (session `86687`, native
exit 0). The failed child's `CHECK=evidence` returned exit 2 (session `80834`,
`artifact/artifact_error`, `invalid-retained-diagnostics`). This check requires a
passing invocation and therefore cannot qualify the failed child; its generic
rejection is not itself proof of a separate corruption defect. Its separate
`CHECK=events` passed (session `58428`, native exit 0). Neither cache hits nor
reader success establish fresh product execution or waive the readiness failure. The nested command handles belong to
the parent assurance process; final owning native exit is recorded below.

#### Final validation outcome and handoff completion

`r02e-s008-final-contracts` finished in native session `36725` with **exit 2**:
60/61 work units passed, one failed, none skipped or cancelled; graph 1,836,623 ms.
The sole failed parent unit is
`operation:harness.assurance_output_check_machine` (`product/test_assertion_failure`,
unit exit 1), because its real child command reported the Network Flow
`infra/service_readiness_timeout` described above. The normal-output check and
all canonical boundary, scheduler, diagnostics, lifecycle and other assurance
operations passed. All **70 parent cleanup steps completed**, and the final
retained-secret scan passed over **12,951 files**. Public `CHECK=events` passed
with native exit 0; this validates event closure, not successful acceptance.
The parent cannot qualify for `CHECK=evidence` because that check requires a
passing invocation, so no redundant failed qualification attempt was made.

Final executable source is
`sha256:1fd7ba2a61e318324c2df8b9bc4f4805806f6d10d3a446df93292365cba7b21a`;
final full-assurance graph is
`sha256:f8d3aaa5f11de338205aa55df0e684e168c1ef3bd85ca61fff8d8950af0ef6eb`.
Toolchain and system/capacity digests remain those recorded at WF-00. No executable
source changed during this validation. Normal-cache evidence includes the child
hits/misses listed above and is not presented as another cache-disabled CI.

**Implementation and handoff are complete; successful aggregate acceptance is
not complete.** S-008 therefore retains a blocked acceptance disposition rather
than a false DONE/pass. No extra rerun, weakened deadline, capacity increase,
cleanup bypass or reconstructed historical proof was used to obtain a green run.
The finalizer ran before broad verification; retained-run maintenance was skipped
with `RESULTS_DIR` unset. A successful nested warm `check` became available only
during this final assurance and was audited independently; no later finalizer
mutation was introduced after the source freeze.

Next owner actions, in order:

1. Preserve W5's exact runtime, lease and failed recovery receipt. The fixture and
   service owners must establish supported recovery authority and physical
   settlement before releasing its three resident allocations. Do not synthesize
   missing environment files from current configuration or infer authority from
   process/container absence.
2. Investigate W2 using the exact final allocation/scope/phase above and the
   existing W1 checkpoint evidence. Define a representative controlled comparison
   and capture service/guest/available physical-host storage observations before
   retirement. Repair only the demonstrated provider, workload, accounting or
   environment cause. A successful unrelated or isolated rerun does not explain it.
3. After the owner repair, verify its focused contracts and rerun the failed
   full-assurance gate with fresh identity and unchanged budgets. Close aggregate
   acceptance only with native exit 0, complete selected outcomes, physical
   settlement, retained scans and successful public evidence qualification.

Final editorial verification compares Appendix A byte-for-byte with the captured
execution baseline, verifies the original logical Git index digest, resolves local
links and runs `git diff --check`. Changes remain unstaged for review. The final
configured documentation command is
`CARTULARY_TEST_RUN_ID=r02e-s008-final-markdown-2 make lint-markdown`; its native
exit and `.cartulary/test-results/r02e-s008-final-markdown-2/adhoc/lint-markdown/tool-run-summary.json`
provide the documentation outcome. No product rerun follows Markdown-only edits.

#### Migration and reviewer handoff

Deploy the coordinator, shared evidence reader, row adapter, wrapper and generated
validators together. The terminal native-report v1 and diagnostic v3 shapes stay
stable, but current consumers require the new invocation disposition and immutable
file packets. Old cumulative checkpoint layouts have no compatibility reader;
retain old run summaries as historical evidence and execute a fresh selection when
current acceptance evidence is needed. No database or public application API
migration is required.

Cohort identities change when compatible owners merge. Reports must use each
catalog row's retained owner/verification binding; a mixed Vitest unit's execution
owner is `harness.command_surface`, and Go execution ownership remains
`harness.backend`. Owner-specific exact selections remain narrow. Update any
saved selector for the retired incident-creation case to
`module.incidents.browser.workbook_home_and_incident_default_preferences_c_6591154e0f`;
there is no old-selector alias. Real UI creation remains in the directory case.

The evidence owner enforces 16 MiB per terminal/disposition artifact and 32 MiB
aggregate file checkpoints, exact request identity, secure file access and
fail-closed contradiction handling. Future evidence-volume expansion must retain
bounded memory/read/write behavior and the interruption/global-failure contracts.
Human review should compare the REQ-102/379/804/805 and AC-087 amendments with these typed
projections; passing tests do not establish normative completeness.

| Separate open investigation | Owner and required closure |
| --- | --- |
| W1 historical first-DROP/storage overrun | `harness.browser` and pgtest/postgrescleanup with the execution-environment owner: capture aligned allocation, generation, drop/checkpoint and available host/storage pressure evidence on reproduction; apply and validate the demonstrated workload, accounting or environment repair. Current successful drops do not prove the old cause |
| W2 historical readiness failure | `harness.browser`, PostgreSQL provider and fixture broker: preserve exact allocation-to-service/scope binding, readiness phase and budget observations on reproduction; demonstrate the causal owner repair and physical settlement. The missing old association cannot be reconstructed by inference |
| W5 current incomplete nested-runtime recovery | `harness.browser` fixture/runtime and service owners: use the exact `real-check-machine` result root, runtime root and lease recorded above. Three acquired suites lack their saved environment. S-009 prevents this ordering defect in new starts; it cannot reconstruct old proof. Closure requires validated physical settlement and release of all three resident allocations through supported ownership, with the original failed result preserved |

### Historical planning-session scope

| Session / baseline | State, files and commands | Next boundary |
| --- | --- | --- |
| 2026-10-20 documentation session; `f9f94168816c9c14438c4efb1e2aa24ccdcdd710` | User chose full reorganization with history preserved and authorized this document update. Root/tooling/backend guidance and tracker framework consulted; static commands only | Revalidate source and this file before later implementation; no harness implementation authorization follows from this handoff |

#### Backend boundary — planning record

| Evidence | State | Next boundary |
| --- | --- | --- |
| E-003/E-007 | Existing broker and pgtest own physical service and transaction lifetimes; no backend code changed | Trace complete connection/effect closure before S-004; amend lifetime owner before S-005 |

#### Frontend boundary — planning record

| Evidence | State | Next boundary |
| --- | --- | --- |
| E-004/E-005/E-009 | Fresh contexts/actors and one-worker cohorts retained; distinct incident assertions identified. Metadata conversion is explicitly deferred, superseding the earlier audit's tentative pilot suggestion | Complete S-003 coverage map; do not relabel metadata state scopes without disproving their recorded exclusions |

#### Contract and codegen — planning record

| Evidence | State | Next boundary |
| --- | --- | --- |
| E-001/E-002/E-003 | No adopted owner, typed contract, catalog, generated file or public API changed | Identify amendment needs and migrate all affected consumers before changing grouping, checkpoint or exclusive-lifetime semantics |

#### Tests and harness — planning record

| Evidence | State | Next boundary |
| --- | --- | --- |
| E-006/E-008 | Recomputed catalog counts match the preceding audit; no test/linter/build execution and no current result root | Use the V-001–V-006 discovery/validation routes only in a separately authorized execution window |

#### Security and authorization — planning record

| Evidence | State | Next boundary |
| --- | --- | --- |
| E-002/E-003/E-004/E-010 | Preserve redaction, test-owned sessions, exact ownership proof, quarantine and private-protocol cutover discipline; no private runtime or credential artifact inspected | Verify cancellation/partial cleanup and failure publication without weakening authority checks |

#### Open risks and next session — planning record

| State | Commands/results and preservation | Next action |
| --- | --- | --- |
| Static findings actionable; implementation discovery and performance proof partial | Only this document changed; logical index unchanged; historical body compared exactly; ignored outputs remain outside inspection | WF-00, then S-002 characterization and S-001 compatibility review; keep W1/W2 separate |

#### Evidence register — planning record

| ID | Inspected source / baseline | Observation and applicability |
| --- | --- | --- |
| E-001 | Compiler Vitest grouping near `vitestGroups` and Go `compatibility`; `buildGoInvocations`; baseline HEAD above | Owner-sensitive grouping and exact selections; startup reduction remains a hypothesis |
| E-002 | `publish`, file loop, `normalizeVitestObservations`, `reconcileVitestReport`, `publishRowResult`, `settledForkPool`; same HEAD | Direct cumulative processing and isolated-worker settlement evidence; no runtime cost measurement |
| E-003 | `clusterKey`, `services`, `releaseCluster`, renderer `reserve`/settlement, `PrepareIsolatedDatabaseT`, `MigrationDatabaseT`, `BeginRollbackDBT`; same HEAD | Physical generation ownership, fresh clone/migration boundaries, serialized rollback buckets and poisoning |
| E-004 | Browser cohort/graph/lane definitions, `webE2EBaseConfig`, `testActor`/`context`, Vite projects; same HEAD | Existing coordinated reuse with per-test state isolation; jsdom distinguished from actual browsers |
| E-005 | Named incident creation tests around lines 649 and 781 respectively at the inspected HEAD | Shared creation choreography plus distinct assertions; no claim of total semantic equivalence |
| E-006 | Standard-library JSON parse of every active row in the 66 named manifests; 2026-10-20, same HEAD | Counts in Section 2; metadata accounting only, not executed workload or source-body coverage |
| E-007 | Graph evidence/invocation/lifecycle estimates, topology lane/physical claims, performance explanation schema; same HEAD | Estimates are not measurements; existing bounded diagnostic interfaces available |
| E-008 | Test-operations skill/recipes, task-surface owner and generated Make declarations; Markdown wrapper/runner and `.markdownlint-cli2.jsonc`; same HEAD | Public command routes discovered statically; none executed. Markdown wrapper writes a cache stamp, and configured globs omit this handoff |
| E-009 | Remediation 01 workstream/disposition tables, including metadata rows around lines 5838–5842, 6341 and 6361 | Prior completed work and explicit exclusion reasons; historical execution not revalidated |
| E-010 | Original handoff SHA-256 in Section 1, preserved in Appendix A | Historical commands, source/run identities, outcomes, omitted observations and W1/W2/W4 causality limits; no artifacts reopened |
| E-011 | Static comparison against `git show HEAD:docs/handoffs/testing-harness-remediation-02.md`, local-link inspection and Git state comparison | Historical body preserved after heading demotion; all new local file links resolve; twelve numbered sections present; logical index unchanged; only this file modified. These are editorial inspections, not repository linter or product verification results |

## 11. Open Questions and Blockers

| ID | Blocker / affected scope | Needed evidence or authority | Status / resolution |
| --- | --- | --- | --- |
| RB-001 | Consumer/selector/assertion mapping for admitted changes | Exact regression and application rows; complete shared-reader cutover and assertion map | RESOLVED for S-001/S-002/S-003; S-005 is explicitly not admitted |
| RB-002 | Adopted compatibility and protocol authority | REQ-379/805 amendments, typed packet/disposition projections and exact attribution | RESOLVED for admitted changes; REQ-804 cluster lifetime unchanged and reuse deferred; S-009 separately clarifies pre-ready authority |
| RB-003 | Isolation proof incomplete or prior explicit exclusion; S-004 | All connections, asynchronous effects, mutable identities, listings and control-plane actions must be traced; retained exclusion is the default | DEFERRED; metadata rows remain isolated |
| RB-004 | Limits on broader performance and capacity claims | Matched representative repetitions and useful host/resource attribution | PARTIAL: S-002 bounded comparison and deterministic S-001/S-003 work removal verified; S-005/S-006 tuning not admitted |
| RB-005 | W1 first-drop and W2 readiness causes remain unproved; historical closure only | Original closure criteria, aligned allocation/scope/phase and storage observations; absence of old bindings cannot be reconstructed by inference | TODO; independent optimization may proceed without claiming these causes repaired |
| RB-007 | Current full-assurance readiness failure and old unresolved recovery | Completed aggregate validation, captured service/environment cause, and validated ownership for W5 settlement | OPEN; exact current W2 scope and failed W5 recovery recorded in Section 10; no budget changes or fabricated proof |
| RB-006 | Controlling handoff was omitted from Markdown lint | Exact handoff added to the existing list in authorized S-007 | DONE; r02e-s007-markdown native exit 0 and tool receipt |

## 12. Binary Completion Criteria

| Outcome | Current disposition |
| --- | --- |
| Plan and authorization | Complete plan with implementation explicitly authorized; this file controls sequencing and handoff |
| Implemented slices | S-007, S-002, S-001, S-003, S-006 and S-009 completed with recorded validation and migration limits |
| Optional dispositions | S-004/S-005 deferred with specific proof/value gates; no global tuning adopted in S-006 |
| Final validation and handoff | Handoff complete; S-008 successful acceptance BLOCKED: final source checks passed, full assurance 60/61 passed with native exit 2 due W2; exact evidence and follow-up recorded |
| Specification fidelity | Owner clauses and downstream projections updated within scope; human review of projection fidelity remains the repository's authority boundary |
| Causal investigations and historical settlement | W1/W2 remain open with current recurrence evidence; W5 retains the exact failed runtime and recovery binding. These are not closed by implemented fixes or later passing runs |

| Completion condition | Evidence / exit criterion |
| --- | --- |
| Every identified gap has a remedy or justified terminal disposition | F-001–F-010, Sections 7, 9 and 10 |
| Historical material and initial user changes preserved | Appendix A byte comparison against the captured execution baseline; final Git/index inspection |
| Exact outcomes and ownership preserved | Focused contracts, 749-row frontend, independent browser survivors and earlier 2,271-row uncached CI passed; final-source assurance is not fully passing |
| Obsolete compatibility removed | No cumulative checkpoint reader/writer, old incident row alias, owner-only grouping barrier or duplicate whole-package native-boundary operation |
| Generated source and documentation consistent | Make generation/finalization, source checks and final explicit Markdown lint |
| Final command success and cleanup independently established | NOT MET for overall acceptance: final assurance native exit 2; the parent and both real-check cleanup ledgers and scans passed, but W5 historical settlement remains unresolved |
| Handoff is actionable | Final source/run identities, migration notes, skipped checks and W1/W2/W5 owner/closure criteria recorded |

## Appendix A — Historical remediation record

The following is the prior handoff, preserved as a historical record. Its original
H1 is replaced by this appendix heading and its subsection headings are demoted
one level; the remaining text is unchanged. Statements such as “current,” “final,”
“verified,” or “changes remain in the working tree” refer to that recorded session,
not this documentation update. Commands and retained artifact references are
historical evidence, not instructions to rerun or recover anything now.

### Scope and baseline

Implementation of the four-blocker remediation plan approved on 2026-10-20.
Starting HEAD: `766846c72fe50328c83e301488ea5394a70dacd2`; working tree clean.
Historical CI: `remediation01-pg07-readiness-ci`, source
`sha256:c4b09f3c40723faeeec6a112370edfbdf8b578db1dd439d8567299fc533a4e5a`.
Historical results do not establish acceptance of the completed source.

The selected SQL readiness contract is 15 seconds, 250 ms polling, and a
5-second maximum individual probe, bounded by the parent deadline. Database
drop, cleanup, durability and ownership constraints remain unchanged.

### Current obligation matrix

| ID / status | Owner and boundary | Required implementation and regression evidence | Exit |
| --- | --- | --- | --- |
| W0 IMPLEMENTED | Testing Harness NLSpec; Core 03 contextual authoring | Contracts and authored projections amended; source/schema/catalog checks pass in final CI | Specification fidelity is ready for human review; Markdown remains outside executable inputs |
| W1 INVESTIGATION OPEN | harness.browser; pgtest/postgrescleanup | Bounded observations implemented; isolated and aggregate execution pass; historical storage cause unproved | Current sampled drops meet their deadlines; a causal workload, accounting or environment repair remains contingent on evidence |
| W2 IMPLEMENTED; HISTORICAL CAUSE OPEN | harness.browser; PostgreSQL provider and fixture broker | Partial ownership, explicit settlement, readiness stages and failure references implemented; final-source Graph Projection row launches and passes | Valid exclusive binding, actual assertions and completed cleanup demonstrated; old allocation-to-service association cannot be reconstructed |
| W3 VERIFIED | module.workbook; contextual reference picker tests | Controlled late-response regression, eight-row recovery selection and all 55 aggregate accessibility rows pass | Exact identities, cancellation and focus preserved; all seven previously skipped rows execute |
| W4 VERIFIED | harness.browser; native session/process ownership | Closed identity diagnostics, durable producer retirement, PID-reuse and descendant regressions pass; exact retained runtime recovered before cutover | Original proof and failure receipts preserved; final CI runtime closes successfully |
| V00 CURRENT EXECUTION VERIFIED; CAUSAL ACCEPTANCE OPEN | Harness and affected source owners | Finalization, final-source CI, required explicit selections, cleanup and evidence checks pass; documentation verification identified below | Overall remediation remains open for the W1/W2 causal investigations; passing execution is not a waiver |

### Sequencing and compatibility

First adopt contracts and bounded diagnostics. Diagnose and recover the old
private-detail authority while its format remains supported. Then cut over
producer lifecycle writers and readers together, without a permanent legacy
reader. Failed or ambiguous proof blocks the affected cutover.

Readiness, storage and identity hypotheses remain separate. No timeout increase,
automatic destructive retry, weakened durability, speculative relocation or
second resource owner is admitted. Resource-profile changes require causal
evidence. Service-heavy comparisons and aggregate validation run sequentially.

### Evidence log

- Discovery: `CARTULARY_TEST_RUN_ID=remediation02-discovery make help ALL=1`
  exited 0. Public recovery requires exact `RESULTS_DIR`, `RUNTIME_ROOT` and
  `RUNTIME_LEASE_ID`.
- A bounded read-only observation of the exact retained private authority found
  440 non-group producer identities absent and 750 groups absent (1190 total).
  This does not explain the historical failing comparison and is not deletion
  authority. At this observation the runtime remained fenced; the successful
  owner recovery below subsequently settled it.

### Handoff and acceptance

The implementation has passed fresh aggregate CI and required explicit checks.
The evidence handoff retains the unresolved causal obligations. Preserve original CI and recovery
receipts. Record native exits, exact run/source identities, row execution and
fixture/runtime settlement separately. Passing current execution does not prove
the historical storage, readiness or process-identity cause. Retained-run
maintenance requires a compatible successful full warm result root.

### Implementation evidence, 2026-10-20

The first diagnostic-only native session selection passed under
`remediation02-identity-diagnostics` (native exit 0). The exact old runtime was
then recovered with:

```sh
CARTULARY_TEST_RUN_ID=remediation02-retained-recovery make test-run-recover RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/remediation01-pg07-readiness-ci RUNTIME_ROOT=/home/jochi/.cache/cartulary/harness-scratch/suite-runtime/suite-remediation01-pg07-readiness-ci-9a3f8492-4969-45fd-8569-7c5bcb76bb8a RUNTIME_LEASE_ID=caaa4e01-8bb2-4af8-bac0-5effd9853c6c
```

Native session 29442 exited 0. The tool receipt is
`.cartulary/test-results/remediation02-retained-recovery/test-run-recover/tool-run-summary.json`.
The owner reported recovery complete, original execution result unchanged; the
exact runtime directory is absent. This recovery used the original detail v1
reader and unchanged identity decisions. No historical retirement acknowledgements
were manufactured. The earlier PID-reuse explanation remains a hypothesis.

The subsequent private detail v2 cutover records immutable producer and launch
proofs plus active/retired state. Native unit and lifecycle-command supervisors
acknowledge retirement before reaping their exact settled group leaders. The
runtime scope still owns inherited descendants. Purge never revisits retired
identities, continues independent safe settlement, and requires durable receipts
before removing private data. Cleanup diagnostic v4 adds closed identity-boundary
categories and opaque producer ordinals.

PostgreSQL startup now preserves a partially acquired handle on cleanup failure,
returns the same ordinary cleanup receipt on subsequent Close calls, and vetoes
retries after logical readiness or unresolved cleanup. Container creation is
separate from start: private lease/credential dependencies are published first.
The fixture broker releases residency only on explicit physical settlement.
Failed prelaunch work carries validated service references through unit and row
results. Service scope records the allocation reference; readiness diagnostics
retain closed stages, budgets, probe counts and outcomes rather than transport
messages. Drop observations cover successful as well as failed attempts, with a
bounded retained sample and an explicit omitted count. Deadlines and durability
are unchanged.

The existing picker component required no production change. The accessibility
assertion now waits for an accepted page and compares exact fixture-derived
identities with polling. A deferred-response component regression exercises
clear, Escape, reopen, late response, focus and explicit Apply.

| Fresh run ID | Selection / action | Native exit | Result |
| --- | --- | ---: | --- |
| remediation02-diagnostic-generate | make generate, diagnostic v4 | 0 | Passed |
| remediation02-lifecycle-generate | make generate after new picker regression | 2 | Failed: new test lacked authored catalog routing; corrected before rerun |
| remediation02-lifecycle-generate-routed | make generate after routing correction | 0 | Passed |
| remediation02-format | make format | 0 | Passed |
| remediation02-native-session | harness.evidence_accounting.behavior.native_session, cache off | 0 | Passed, including live group retirement and deterministic identity substitution |
| remediation02-postgres-startup | platform.postgres.support_unit.certificate_fixture_lifecycle, cache off | 0 | Passed, including unresolved partial container cleanup |
| remediation02-picker-delayed | web.workbook.regression.authoring_candidate_presentation, cache off | 0 | Passed, including controlled late response |
| remediation02-native-runtime | harness.command_surface.unit.native_runtime, cache off | 0 | Passed |
| remediation02-native-fixtures | native_fixtures, pgtest_helper_contract, testcontainersx_helper_contract, testservices_lifecycle_contract under harness.browser, cache off | 2 | Three units passed; service-selection test doubles omitted acquired container identities; corrected |
| remediation02-native-fixtures-repaired | Same four harness.browser rows after test-double correction | 2 | Three passed; pgtest tests lacked the new postgrescleanup import; corrected |
| remediation02-postgres-acquisition | platform.postgres.support_unit.certificate_fixture_lifecycle | 2 | The same missing test import blocked compilation; corrected before further validation |
| remediation02-observation-generate | make generate, readiness/drop/reference projections | 0 | Passed |
| remediation02-ownership-format | make format | 0 | Passed |

These are intermediate-source checks, not final-source acceptance. Every root is
`.cartulary/test-results/<fresh-run-id>`. Historical failed allocation evidence
identifies the Graph Projection unit, but the retained service scope lacks an
allocation binding. Its association with the 17,635 ms service attempt remains
unproved; new scope linkage must supply that evidence on reproduction.

The predetermined storage comparison remains: each exact affected PostgreSQL
selection alone, then representative aggregate CI with basic diagnostics, the
same source/storage/durability/capacity policy, and sequential service-heavy
invocations. Correlate scope allocation, generation and drop timestamps with
checkpoint completions and existing host-pressure observations. Do not equate
unavailable host/storage observations with zero pressure, and do not infer a
performance guarantee from this bounded sample.

### Intermediate comparison and audit results

`make agent-finalize` passed under `remediation02-source-finalize` (native
75433/0), with generated artifacts unchanged. The comparison source was
`sha256:512ad296b489a4eaa2380fa78cd3ec9de1a504cdfbdd7ea523cf47f7b9a22c35`.
Its retained-run maintenance was explicitly skipped because `RESULTS_DIR` was
unset and no compatible successful full warm root had been established.

All service selections below used `CARTULARY_HARNESS_CACHE_MODE=off`,
`HARNESS_DIAGNOSTICS=basic`, fresh explicit run IDs, and sequential invocation.

| Run ID | Exact selected owner / rows | Native / exit | Result |
| --- | --- | --- | --- |
| remediation02-incident-isolated | module.incidentbundles / module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8 | 71403 / 0 | Passed; first drop 1841 ms of 19999 ms effective operation budget; SQL readiness 1265 ms |
| remediation02-networkflow-isolated | module.networkflow / module.networkflow.integration.saved_graph_lifecycle_v2 | 93540 / 0 | Passed; first drop 1674 ms of 19999 ms; SQL readiness 1015 ms |
| remediation02-collaboration-isolated | module.collaboration / module.collaboration.integration.entity_candidate_discovery_does_not_publish | 73386 / 0 | Passed; first drop 1655 ms of 19999 ms; SQL readiness 1014 ms |
| remediation02-graphprojection-isolated | module.graphprojection / module.graphprojection.storage.result_v2_cleanup_locking | 64681 / 0 | Launched and passed, 2200 ms execution; SQL readiness 1014 ms |
| remediation02-picker-recovery | module.workbook / the eight accessibility rows below | 63659 / 0 | All eight passed; 12/12 units passed |

Every listed run completed fixture and runtime cleanup, with no non-completed
cleanup step. These observations do not establish the cause of the historical
readiness timeout or storage overrun.

The exact eight-row selection was:

```sh
CARTULARY_TEST_RUN_ID=remediation02-picker-recovery CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.contextual_task_decision_creation,module.workbook.accessibility.ordinary_create_authoring_recovery,module.workbook.accessibility.coordination_create_authoring_recovery,module.workbook.accessibility.timeline_related_evidence,module.workbook.accessibility.note_create_authoring_recovery,module.workbook.accessibility.timeline_row_action_menu,module.workbook.accessibility.timeline_clear,module.workbook.accessibility.batch_recovery HARNESS_DIAGNOSTICS=basic
```

Further caller audit extended memoized partial-failure handling to object-store
startup. One invocation-owned serialized receipt now preserves both concurrent
providers' acquired containers before either starts readiness. This is a shared
acquisition-contract correction, not a common root-cause claim.

Native review and physical-cleanup workers retire producer registrations before
reaping their settled direct children. Registrations bind through revalidated
ancestry to the runtime owner's exact direct child, including nested producers
that create another group/session. A controlled live regression proves that a
detached descendant can survive launch retirement and is still settled by the
inherited runtime scope before purge. No missing or reused anchor grants
retirement, and no diagnostic grants authority.

The subsequent readiness refinement records security/capability validation
separately from SQL connection, using the same SQL probe and remaining deadline.
It adds no query, readiness window, or retry. Binding publication remains owned
by the service lifecycle and is success-only after provisioning.

Additional intermediate checks:

| Run ID | Action | Native exit / result |
| --- | --- | --- |
| remediation02-focused-lifecycle | Six focused harness.browser lifecycle rows | 2; five passed, object-store test did not compile due to a missing new import; corrected |
| remediation02-native-final-a | Same six rows after correction | 0; six passed |
| remediation02-partial-publication | PostgreSQL, object-store and testservices helper contracts | 0; passed |
| remediation02-retirement-supervisors | native_session | 0; passed |
| remediation02-native-review | native_review | 0; passed |
| remediation02-enclosing-launch | native_session including live detached descendant | 0; passed |
| remediation02-capability-generate | make generate | 0; passed |

The final ancestry and readiness-stage refinements change source identity.
The preceding runs remain intermediate evidence; final-source validation must
cover those refinements before overall acceptance.

### Completed-source baseline and aggregate acceptance

The final executable snapshot is
`sha256:54b53ac3f86228bdc8967604ce0d103488e5c3ae4196f65b690bc44ae7ecec3a`.
`make agent-finalize` under `remediation02-scope-finalize` completed with native
session 26022 / exit 0; generated artifacts were unchanged. Later handoff edits
are documentation-only and do not enter executable source identity.

The final acquisition audit also covers object-store callers of the shared
startup wrapper. Both providers publish partial authority before starting a
created container. One serialized invocation receipt preserves concurrent
provider acquisitions. Failed ordinary cleanup is memoized, retains credential
dependencies and resident claims, and is not silently retried.

Producer retirement now requires an exact, unreaped supervised launch anchor.
Transfer for a producer outside that launch group also requires revalidated
ancestry and verified inherited runtime-scope membership. Unmarked descendants
retain their individual active proof. A failed retirement publication cannot
grant successful closure. The new boundary regressions cover shorter parent
readiness deadlines, terminal capability rejection, failed retirement writes,
marked detached descendants and unmarked descendants.

All commands in the following table ran from the repository root. Service runs
used fresh IDs, `CARTULARY_HARNESS_CACHE_MODE=off` and
`HARNESS_DIAGNOSTICS=basic`, sequentially.

| Run ID | Public command / exact selection | Native session / exit | Observed result |
| --- | --- | --- | --- |
| remediation02-load-incident | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8 HARNESS_DIAGNOSTICS=basic` | 38126 / 0 | Passed; 33,682 ms graph duration |
| remediation02-load-networkflow | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2 HARNESS_DIAGNOSTICS=basic` | 47504 / 0 | Passed; 66,971 ms |
| remediation02-load-collaboration | `make test-slice OWNER=module.collaboration ROWS=module.collaboration.integration.entity_candidate_discovery_does_not_publish HARNESS_DIAGNOSTICS=basic` | 64119 / 0 | Passed; 14,526 ms |
| remediation02-load-graphprojection | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.result_v2_cleanup_locking HARNESS_DIAGNOSTICS=basic` | 79753 / 0 | Launched and passed; 8,646 ms graph duration |
| remediation02-aggregate-ci | `make ci HARNESS_DIAGNOSTICS=basic` | 71887 / 0 | 741/741 work units passed, no failed/skipped/cancelled units; 3,308,909 ms graph duration |
| remediation02-ci-evidence | `make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/remediation02-aggregate-ci CHECK=evidence` | 11223 / 0 | Canonical execution, selected coverage, cleanup and retained-boundary evidence passed; scoped evidence check, not a complete diagnostic-bundle claim |

Aggregate graph identity:
`sha256:bdaece9976376e9e457a35c40ab4b70e8f13a08a6d9b3b7f2b33ef84cee95584`.
The owning command exited only after the retained-secret scan passed over
10,639 files. The terminal receipt reports child exit 0 and 3,418,140 ms outer
duration; graph duration excludes the scan and other outer work.

All 1,398 cleanup steps completed: 741 host releases, 494 fixture releases,
154 fixture detachments, seven provider closures, one services closure and one
runtime closure. The canonical event stream records 396 acquired and 396 released
fixture leases. Runtime closure has no failure diagnostic. Fixture and runtime
settlement are separate verified outcomes.

The Graph Projection row executed its locking/rollback test in 1,770 ms. The
Incident Bundles, Network Flow and Collaboration affected rows passed in
54,770 ms, 59,810 ms and 6,800 ms respectively. All 55 accessibility rows passed,
including the exact contextual creation row and all seven formerly skipped
rows listed in the eight-row recovery selection above.

CI also executed all four required Network Flow measurement rows:
`exploration_graph_dom_ceiling`,
`measure_bounded_dom_growth_and_semantic_reachabi_297b6a7322`,
`network_flow_selector_covers_the_actual_accepted_f68f4c7a57`, and
`saved_graph_dom_ceiling` (each prefixed `module.networkflow.measurement.`).
Their durations were 7,722, 2,325, 4,214 and 728 ms. These passes are reused;
the release-tier backend capacity certification is not claimed.

### PostgreSQL comparison and evidence limits

The isolated and aggregate runs share executable source, toolchain and capacity
identities. The system/capacity identity is
`sha256:008fe637b622cd8932b55cc1d8be5346bae02c09ab35d7518fc9f72b66eb6c21`;
the toolchain identity is
`sha256:f0968e3ace0b7282a384ecb609ccc16db0e9e49c646cf0b45ba457ed4669ae6c`.
Capacity remained 19 CPU tokens, 38 I/O tokens, 16 PostgreSQL lanes, eight
object-store lanes, 76 process slots and 32 port lanes. Durability, storage
placement, deadlines and resource profiles were not changed.

The table reports the first recorded drop of each selected generation. It does
not assign every generation-level observation to a particular test within a
batched unit. Every listed drop had a 19,999 ms effective operation budget and
a 15,000 ms admission budget.

| Generation owner | Isolated first drop / remaining operation budget (ms) | Aggregate first drop / remaining budget (ms) | Isolated / aggregate SQL readiness (ms) |
| --- | ---: | ---: | ---: |
| Incident Bundles | 1,878 / 18,121 | 12,016 / 7,983 | 1,263 / 8,510 |
| Network Flow | 1,615 / 18,384 | 4,071 / 15,928 | 1,011 / 2,761 |
| Collaboration | 1,603 / 18,396 | 5,655 / 14,344 | 1,010 / 3,012 |
| Graph Projection | 1,613 / 18,386 | 908 / 19,091 | 1,009 / 1,262 |

Aggregate generation/scope bindings are retained under the aggregate run root:

| Owner | PostgreSQL generation | Service-scope reference |
| --- | --- | --- |
| Incident Bundles | 7698725448361181224 | `_shared/test-services/786465407c6e454591f8080d/service-scope.json` |
| Network Flow | 7698725886402932776 | `_shared/test-services/e60b2bc31fbd4b32b46473f1/service-scope.json` |
| Collaboration | 7698725938882084904 | `_shared/test-services/9a7b321faf094a808ae914bf/service-scope.json` |
| Graph Projection | 7698730972110041128 | `_shared/test-services/f55831947cf2433b9a928326/service-scope.json` |

Each scope links `postgres.allocation_ref` to its exact canonical unit. The
checkpoint diagnostic links the same scope to the generation. All four service
scopes explicitly report completed physical settlement.

All 196 PostgreSQL readiness extensions report ready; maximum observed SQL
connectivity was 9,516 ms within the selected 15,000 ms window. The 745 retained
drop observations all completed; maximum recorded operation was 18,829 ms,
below its effective budget. This maximum is a sampled observation, not a future
performance guarantee or a claim that omitted observations had zero cost.

Incident Bundles retains 32 of 36 drop observations and explicitly reports four
omissions. Its largest recorded drop was 13,658 ms. Its largest observed
checkpoint synchronized 3,894 files in 13,286 ms. Network Flow's largest
checkpoint synchronized 1,398 files in 3,967 ms; Collaboration's synchronized
1,397 files in 5,546 ms. These observations are consistent with greater sync
cost under aggregate load, but do not identify the responsible competing
workload or physical storage cause.

The aggregate checkpoint collector retains 820 records for 196 service scopes.
Four scopes are unavailable because of measurement quiet periods. The execution
index is bounded at 4,096 records and reports 2,887 omitted references. Resource
coverage is explicitly partial: 10 discovery truncations and 2,909 omitted
observations. Linux guest pressure and process/container I/O are available only
for observed intervals; they do not measure the Windows host storage layer.
No whole-run resource total or causal attribution is inferred from partial data.

The initial isolated runs select one row; CI coalesces 11 Incident Bundles,
eight Network Flow and six Collaboration rows into their respective units.
An additional predetermined isolated execution of those exact CI batches holds
the fixture population constant. All three completed successfully:

| Run ID | Native session / exit | Rows / work units | Graph duration (ms) | First / maximum retained drop (ms) | SQL readiness (ms) |
| --- | --- | --- | ---: | ---: | ---: |
| remediation02-matched-incident | 32417 / 0 | 11 / 3 | 60,231 | 2,935 / 2,935 | 1,264 |
| remediation02-matched-networkflow | 76275 / 0 | 8 / 3 | 76,921 | 1,613 / 1,613 | 1,010 |
| remediation02-matched-collaboration | 91446 / 0 | 6 / 3 | 17,162 | 1,550 / 1,550 | 1,009 |

The canonical unit IDs, runner command, row bindings, resource claims, physical
allocations and PostgreSQL isolation match the corresponding aggregate units.
Incident Bundles created 37 databases (one migrated template and 36 clones),
Network Flow 14 (one template and 13 clones), and Collaboration seven (one
template and six clones), identically in isolation and aggregate CI. No redundant
migration is demonstrated by these observations.

The matched scope references are respectively
`_shared/test-services/45d0c71fcf064c0d82ab8081/service-scope.json`,
`_shared/test-services/64b448f40acc45a2b3fd10b2/service-scope.json`, and
`_shared/test-services/2ceb9d8990d641f99855f256/service-scope.json` under their
exact run roots. Each completed all 14 cleanup steps and passed its retained
scan and public evidence check. Incident Bundles again explicitly omits four
drop observations after its first 32. The original single-row comparisons remain
useful execution evidence but are not controlled estimates of concurrency's effect.

Exact matched commands (fresh run IDs and unchanged cache/diagnostic policy):

```sh
CARTULARY_TEST_RUN_ID=remediation02-matched-incident CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_descriptor_reads_reject_singleto_ea18aa4fe5,module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8,module.incidentbundles.integration.incident_bundle_export_jobs_use_the_incident_sco_d94a8de941,module.incidentbundles.integration.incident_bundle_export_uses_the_common_job_resou_a1e51b82d7,module.incidentbundles.integration.incident_bundle_import_final_publication_recheck_b7b07b3e93,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755,module.incidentbundles.integration.incident_bundle_import_upload_envelope_failures_bf5449e3d4,module.incidentbundles.integration.incident_bundle_import_z_evidence_cleanup_a3d619b413,module.incidentbundles.integration.incident_bundle_retired_version_is_rejected_with_1ec7e171fe,module.incidentbundles.integration.layout_legacy_conversion_replay,module.incidentbundles.integration.network_flow_retained_state_blocks_bundle_export_7dcd6f85c2 HARNESS_DIAGNOSTICS=basic
CARTULARY_TEST_RUN_ID=remediation02-matched-networkflow CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.integration.effective_resource_limit_discovery,module.networkflow.integration.pagination_recovery,module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5,module.networkflow.integration.saved_graph_cutover_v6,module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.table_lifecycle_admission_replay,module.networkflow.integration.time_bucket_saved_graph_lifecycle HARNESS_DIAGNOSTICS=basic
CARTULARY_TEST_RUN_ID=remediation02-matched-collaboration CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.collaboration ROWS=module.collaboration.integration.closed_incident_socket_terminates_before_writabl_8057ea1833,module.collaboration.integration.cookie_authenticated_browser_socket_upgrades_rej_73faec66da,module.collaboration.integration.durable_incident_stream_commits_intents_atomically_s_80128fe153,module.collaboration.integration.entity_candidate_discovery_does_not_publish,module.collaboration.integration.resume_with_a_valid_replay_token_replays_replaya_1826b38095,module.collaboration.integration.two_real_clients_connect_to_the_same_incident_so_257bcc588c HARNESS_DIAGNOSTICS=basic
```

### Operating and migration guidance

The Testing Harness NLSpec owns lifecycle behavior (especially REQ-502 and
TH-HARNESS-REQ-804/811); Core 03 owns contextual-reference cancellation.
`domain.md` already provides the appropriate vocabulary/navigation and required
no change. Generated schemas, validators and routing were refreshed through
Make. Human review of the normative amendments remains distinct from automated
projection and conformance checks.

- On a PostgreSQL deadline failure, inspect the exact allocation, unit, service
  scope and readiness extension. Distinguish container start, endpoint resolution,
  SQL connectivity, capability validation and generation initialization from
  success-only provisioning/binding publication. Compare configured and effective
  budgets before diagnosing a timeout. Authentication and capability rejection
  are terminal, not startup-transient retry candidates.
- On partial startup, preserve the acquired private lease, credential dependencies
  and primary/secondary outcomes. `startup_failed` is not a physical settlement
  receipt. Only explicit completed settlement releases residency. An unresolved
  ordinary cleanup receipt does not authorize another ordinary Close/Stop.
  Provider death uses the existing owner recovery path and exact published
  authority; absent proof remains unresolved.
- On database-drop timeout, preserve the first attempt, its effective deadline,
  generation, checkpoint evidence and quarantine. Do not retry DROP, clear the
  quarantine or accept a late server completion as restored reuse eligibility.
  Runtime/pool shutdown, admission, DROP, unlock and connection-close failures
  remain separately attributable.
- On process-identity failure, use the closed boundary category and opaque
  registration reference to locate the private owner proof. Diagnostics cannot
  authorize signaling or deletion. A missing, reused, unreadable or unanchored
  identity remains protected. Continue independently safe settlement while
  retaining unresolved authority.
- A producer exit is not descendant settlement. Retirement requires durable
  acknowledgement at the supervised launch boundary. Inherited runtime-scope
  descendants remain the runtime's responsibility; unmarked descendants retain
  individual proof. Failure to publish retirement leaves closure unsuccessful.
- For recovery, use `make test-run-recover` with the exact result root, runtime
  root and lease from the owning record. Do not manually delete runtime data,
  infer ownership from a PID/name, or manufacture old retirement records.
- For asynchronous picker assertions, wait for the accepted candidate page,
  then compare the complete fixture-derived selected-identity collection with
  retrying equality. Preserve empty staged state, dismissal, focus and explicit
  Apply semantics. Loading completion is not inferred from Apply being enabled.

The exact retained detail-authority v1 runtime was drained through its supported
owner before the v2 writer/reader cutover. Cleanup diagnostics cut over from
v3 to v4. New readiness and prelaunch service-failure projections cut over with
all their writers, validators and readers. Historical artifacts are preserved,
not rewritten into the current epoch. Other environments with unresolved old
private runtimes must recover them using the old supported owner before deploying
this cutover; there is no permanent legacy reader.

Product HTTP contracts, persisted data and draft schemas did not change. The
picker production implementation did not change. No compatibility alias,
additional scheduler, deadline increase or resource-capacity change was added.
The significant migration risk is mixing private protocol epochs while old
runtime ownership remains outstanding.

One directly discovered guidance defect was corrected: the test-operations
command recipe listed `DETAIL=resources`, which the current public `explain-run`
rejects. The recipe now uses `DETAIL=performance JSON=1`, which includes resource
observations. The two diagnostic reader attempts using the stale option exited
2 with `config/usage_error`; neither reran or changed the retained CI. The
subsequent performance reader exited 0. These are reader outcomes, not test
failures or acceptance audits.

### Unresolved causal obligations

Current successful execution does not explain the historical failure. The
following investigations stay open; no speculative repair or waiver is applied.

| Obligation | Evidence and remaining uncertainty | Owner / next bounded action | Closure criterion |
| --- | --- | --- | --- |
| W1 historical first-drop overrun | Historical `CheckpointDone` / `DataFileSync` overrun is confirmed. Current aggregate drops pass but sync remains slower under aggregate work; the responsible workload/storage cause is unproved. | harness.browser / pgtest and environment owner: use the matched-batch evidence below, then capture one predetermined representative aggregate comparison with aligned guest and physical-host storage observations if the overrun recurs. Preserve original budgets/capacity and capture before retirement. | A demonstrated redundant-write, resource-accounting, throughput or supported-environment defect is repaired at its owner and the affected first drops meet their effective budgets under representative load. |
| W2 historical readiness timeout | The old allocation identifies the Graph Projection unit, but the 17,635 ms scope has no allocation binding. All 196 current PostgreSQL acquisitions report ready. | harness.browser service owner: retain the unreconstructable association explicitly; on a bounded reproduction use the new allocation/scope link and phase/deadline outcome to identify the failed boundary. | A captured cause is repaired at its owner and current execution/physical settlement continue to pass. Until the cause is established, the historical investigation remains open. |
| W4 historical identity comparison | Old registered processes were absent at later observation; owner recovery succeeded. The failing historical comparison was not captured, so PID reuse remains a hypothesis. | harnesssession owner: use the new closed boundary diagnostics for any recurrence. Existing deterministic comparison-race tests and durable retirement address the demonstrated lifecycle weakness. | No historical root-cause claim without captured proof; foreign-process protection and final runtime closure remain required. |

The next storage comparison must not be a retry-until-green loop. Record its
population, source, admission profile, storage environment, duration and missing
observations before deciding whether another implementation change is warranted.
Guest pressure alone cannot identify the Windows host's competing workload.

### Explicit validation, phases and final handoff

| Run ID | Exact public command after the run/cache environment prefix | Native session / exit | Result and physical settlement |
| --- | --- | --- | --- |
| remediation02-visual | `make browser-e2e SUITE=visual HARNESS_DIAGNOSTICS=basic` | 51616 / 0 | 53/53 work units and all 49 rows passed; 677,023 ms; all 106 cleanup steps completed, including runtime closure; retained scan passed over 766 files |
| remediation02-ui-review | `make harness-contract SUITE=ui-review HARNESS_DIAGNOSTICS=basic` | 11981 / 0 | 10/10 work units passed; 102,704 ms; all 19 cleanup steps completed, including runtime closure; retained scan passed |
| remediation02-imports | `make test-slice OWNER=module.imports ROWS=module.imports.measurement.csv_apply,module.imports.measurement.csv_discovery,module.imports.measurement.xlsx_apply,module.imports.measurement.xlsx_discovery HARNESS_DIAGNOSTICS=basic` | 82601 / 0 | All four rows and 21/21 work units passed; 314,412 ms; all 40 cleanup steps completed, including runtime closure; retained scan passed |

The prefix is `CARTULARY_TEST_RUN_ID=<listed-run-id>
CARTULARY_HARNESS_CACHE_MODE=off`. All three manifests retain the completed-source
identity above. Imports measurement intervals ran without concurrent service
experiments, diagnostic readers or documentation work.

Each explicit root and matched-batch root passed:

```sh
CARTULARY_TEST_RUN_ID=<exact-execution-run-id>-evidence make harness-observability-check RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/<exact-execution-run-id> CHECK=evidence
```

The exact execution IDs are the six rows in the matched/explicit tables above.
Each evidence command returned native exit 0. This check validates canonical
results, coverage, cleanup and retained-boundary closure. It deliberately does
not claim complete diagnostic sampling when bounded collectors report omissions.

| Phase | Disposition and dependencies satisfied | Remaining risk / exit status |
| --- | --- | --- |
| P0 contracts | Implemented owner amendments and obligation matrix; baseline separated from historical source | Human review must assess normative fidelity; generated checks cannot substitute for that review |
| P1 diagnostics | Bounded readiness/drop/identity evidence and prelaunch references implemented; fault-injection and redaction checks pass in focused validation and CI | Declared truncation and measurement quiet periods limit attribution; no diagnostic grants authority |
| P2 lifecycle and recovery | Partial-acquisition ownership and producer retirement implemented; exact old runtime recovered before cutover; W3 regression also passes | Other environments must settle old proof before private-format cutover |
| P3 storage/readiness | Matched isolated and aggregate execution pass with unchanged budgets and capacity | Historical W1/W2 causes remain open; no demonstrated storage/workload repair is claimed |
| P4 final-source validation | Finalization, focused comparisons, CI, visual, UI-review, Network Flow and Imports measurements, retained scans and cleanup pass | Current execution acceptance is complete; it does not close P3's causal investigation |
| P5 handoff | Controlling tracker updated; exact identities, commands, recovery, compatibility and unresolved-owner actions retained here | Overall plan is not declared fully remediated while P3 remains open |

Required retained-run maintenance was skipped with `RESULTS_DIR` unset. There is
no compatible successful full warm `make check` root with `cache_mode=normal`;
the successful cache-disabled `make ci` root is not eligible. No extra warm run
or relaxed preflight was introduced to manufacture maintenance eligibility.

Final documentation verification uses
`CARTULARY_TEST_RUN_ID=remediation02-final-markdown make lint-markdown`, with
receipt root `.cartulary/test-results/remediation02-final-markdown`.
The owning terminal result and receipt establish its outcome. Documentation
does not trigger another product generation or CI run.

Changes remain in the working tree for review. Original failure artifacts and
recovery receipts are unchanged. Current passing execution, safe lifecycle
repair, and unresolved historical causation are separate dispositions.
