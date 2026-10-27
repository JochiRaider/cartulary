# Testing Remediation 01: Object-Store Fixture Simplification

## 1. Scope and Source Posture

Document date: 2026-10-26. **Current action: completed review and corrections**, authorized by the
user's follow-up request to review guarantees, regressions and complexity, fix
issues, run focused validation and update this handoff.
Execution order is WF-20 → WF-21 → WF-25 → WF-22 → WF-23 → WF-24.
The earlier document-only saves remain historical. This invocation admits the
bounded source, test, caller and routing changes below, their Make validation and
handoff updates; it does not authorize normative amendments or unrelated redesign.
Artifact mode remains **normal for this bounded iteration**.

Repository: `/home/jochi/code/cartulary`; established supplied and canonical target
`.`; kind: repository directory; iteration label: `object-store-fixture-simplification`.
The handoff is `docs/handoffs/testing-remediation-01.md`. Its parent
components are not symlinks; the output is a regular, singly linked file. The
existing tracker's repository identity is retained. The [modular refactor
framework](cartulary_modular_refactor_planning_framework.md) and Refactor Tracker
format supply planning structure, not behavioral authority.

The implementation change boundary is `internal/testutil/s3test`, object-store composition
in `tools/testservices`, directly affected process/application/recovery consumers,
and their authored routing and generated projections. Keep physical service
ownership distinct from borrowed fixture access. Remove accidental test
constraints at their source; do not add compatibility adapters, public injection
frameworks, another service manager, a fixture-policy language, or a new runner.
Production APIs/data schemas, provider replacement, PostgreSQL internals, scheduler
architecture, browser presentation, and performance-fixture policy are excluded.
Affected consumers remain regression witnesses, not invitations to redesign them.

WF-00–WF-19 remain completed at their historical baselines. E-047 records the
PostgreSQL closure's 22 compatible roots and 42 selected rows; those results do
not qualify this object-store iteration. The separate open WF-33 in [the
preceding handoff](testing-remediation.md) is neither resumed nor closed here.
Historical instructions inside this handoff describe their original sessions;
they do not enlarge the current bounded implementation request.

| Authority / status | Required distinction and applicable clauses |
| --- | --- |
| [Core 00](../spec/00_document_set_status_and_precedence.md), current normative owner allocation | Adopted owners govern requirements. Tests, typed projections and this handoff do not establish specification completeness. Tests, generators, runtime metadata, conformance and release evidence must not consume Markdown. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md), `adopted/current` | REQ-804: explicit attachment, independent physical owners, cancellation, exact recovery proof and quarantine. REQ-410 and §11: declared capabilities/dependencies, namespace ownership and cleanup. §11.4: shared mutation-readiness contract, deadlines and package admission. REQ-414 and §9: bounded diagnostics and primary/secondary failure classification. §§3/8: exact routing and truthful evidence. |
| [Core 04](../spec/04_security_deployment_and_conformance.md), §1.3 | Preserve verified transport, the shared S3 transport policy and secret confinement. Pins remain machine-owned in `tools/toolchain_pins.json`; no new pin inventory or qualification claim. |
| [Domain](../domain.md), §16 / §16.1 | Vocabulary and owner navigation: object storage backs binary evidence; buckets and keys are not public evidence identity. No domain-model change is planned. |
| [NLSpec guidance](../research/nlspec-spec.md) | Supporting guidance on requirements versus implementation choices; not adopted authority for fixture APIs, test names, layouts or mock behavior. |
| [Core 05](../spec/05_claim_publication_and_benchmark_reproducibility.md) | No timing, performance or release-publication claim. Functional fixtures do not replace distinct claim-bearing qualification. |

No inspected owner requires a mixed `Harness`, three owned-start wrappers,
package-global test substitutions, fake deletion that succeeds while reporting
failure, wait-strategy description matching, or one particular cleanup request
count. No specification amendment is proposed to preserve those arrangements.
Binding validation and quarantine corrections implement existing requirements;
they are explicitly identified separately from mechanical movement below.

| Baseline item | Implementation entry, E-057 | Current comparison | Limits |
| --- | --- | --- | --- |
| Branch / commit | `main`; `9e6fa4053e2de4367ce106606da02e74fafc72d5` | Same commit; no commits or branch changes | Historical evidence retains its own source scope |
| Index / staged state | No staged changes; entry SHA-256 `5370d11e07f02f067cde89dc1b15d83aa93adfb200c82034f6ab8db1a31b9f04`; raw-index SHA-256 `443632109419c6b351af3837dc3abbf68e1eed208ab4403c7bbfb01a285284c8` | Staged entries unchanged; raw index metadata differs during implementation, with no staged changes | Raw index byte identity is not claimed |
| Worktree / untracked | Pre-existing edits confined to this handoff; 7,649 tracked/nonignored entries | Exactly 20 changed/added/removed entries, including this handoff and the generated topology index; 7,654 current entries; E-063, with six existing entries updated during review at E-064 | Ignored build/cache/runtime trees excluded |
| Tracker | 229,205 bytes; inode 348826; SHA-256 `43960f81a5e8c28a8853e425330bb017d9b939f504aa2b3de02c8fd24b0cd7cc` | Existing edits and all prior historical log text preserved | E-053/E-056 describe the preceding save, not this entry |
| Ignored outputs / dependencies | Installed runtime/frontend prerequisites and cached pinned Go | Invocations use explicit result roots and pinned Go executable/PATH; no dependency or pin updates | No exhaustive ignored-tree fingerprint |

Implementation preservation material is in
`/tmp/cartulary-s3-implementation-p5bpj7gf/`: `baseline.json`,
`original-handoff.md`, `pre-finalize.json` and `frozen-source.json`.
E-048–E-056 remain historical document-save evidence. Source accounting includes
tracked/nonignored entries under `.`; generated products remain projections and
never behavioral owners. None of the executable changes consumes Markdown.

## 2. Current-State Repository Inventory

This table records the inspected **pre-refactor state**, revalidated at WF-20
(E-057/E-058). The original five S3 Go files contained 30 top-level tests and
3,076 lines; those counts were inventory, not a reduction quota. Current owners,
source groups and qualified successors are recorded in E-059–E-063. The bounded
caller inspection does not claim an exhaustive audit of the repository.

| Path / kind | Disposition / inspection depth | Responsibility and inbound callers | Dependencies / tests / projections | Owner / risk / evidence |
| --- | --- | --- | --- | --- |
| `internal/testutil/s3test/s3test.go` | Directly inspected, authored | Container ownership, borrowed attachment, cached client, readiness, namespace creation/reset/cleanup, diagnostics and attribution; used by testservices, appsupport, process tests and performancefixturelifecycle | testcontainersx, suiteservices, S3 transport, private resource journal; 25 tests in companion file | Split physical owner from fixture operations; C-017–C-020; E-049 |
| `s3test/s3test_test.go` | Directly inspected, authored | Startup/attachment, mutation probe, package admission and real namespace tests | Global substitution helpers; fake readiness state; real service fixtures; five catalog rows across the package | Consolidate assertions, localize dependencies and strengthen absence proof; F-028–F-031/F-033; E-049/E-054 |
| `s3test/object_cleanup.go` | Directly inspected, authored | Exact-prefix pagination, bounded delete batches and independent emptiness proof | Narrow private client port; no new algorithm/framework needed | Keep useful boundary; C-019; E-049 |
| `s3test/object_cleanup_test.go` | Directly inspected, authored | Five tests cover cleanup protocol faults, package quarantine, real pagination and pre-client ownership reservation | Local cleanup fake already has a narrow port; package cases still patch globals | Move package cases to their owner and preserve cleanup faults; C-019; E-049 |
| `s3test/tls.go` | Directly inspected, authored | Owner TLS directory, server/proxy identities and root-certificate access | tlstest, suiteservices runtime root and container configuration | Move private material ownership only; verified client trust remains; C-017/C-020 |
| `tools/testservices/main.go`, `main_test.go`, `partial_ownership.go` | Directly inspected affected composition, callback, recovery and publication bodies/assertions | Sole external owned-start caller; broker probe bucket, lease publication, browser namespace reset/cleanup; two replaceable object-store globals | Existing dependency seams, partial startup journal, bounded failure summaries; lifecycle-contract row | Keep orchestration; use local object-store seam, no new manager; C-017/C-020 |
| `internal/app/serverprocess/process_fixture_harness_test.go`, relevant config/bootstrap callers | Directly inspected setup/teardown and affected assertions; remaining bodies metadata-only | TestMain borrows services, then calls misleading `StopShared`; evidence process uses fixture bucket | Existing app.server process selectors and borrowed owner lifetime | Migrate local transport cleanup; no product assembly changes; C-017/C-021 |
| `internal/testutil/appsupport/runtime.go`, runtime guards; `performancefixturelifecycle/lease.go` | Directly inspected service acquisition and consequential namespace cleanup; remaining bodies metadata-only | Composite fixture acquisition and snapshot consumer attachment | Existing support guard, managed-service and process regression witnesses | Keep consumer semantics and performance policy; C-017/C-019 |
| `internal/platform/objectstore/s3transport/connection.go`; suiteservices env/dependency/event boundaries | Directly inspected shared connection policy, local close and relevant binding/journal paths | Verified HTTPS, single-attempt product requests, captured suite inputs and cleanup authority | Local transport Close releases idle connections; it does not terminate a service | Keep production transport; bind support inputs in S-025; C-017/C-020 |
| Native boundary `fixtures_linux_test.go`, `selected_services_linux_test.go` | Directly inspected affected lifecycle/selection boundaries; remaining bodies metadata-only | Owned/borrowed physical integration and nested exact-row selection | Existing broker/runtime and service fixtures | Retain independent physical evidence; migrate changed nested row IDs; C-017/C-021 |
| Affected `tools/test_families` rows, verification owners, task/catalog/topology and generated policies | Directly inspected affected selectors and ownership/projection declarations | Authored routing; a real reset is currently classed `unit` | Make generation and generated/schema/catalog/identity checks are future work | Repair authored inputs atomically; F-032/C-021 |
| Instructions, framework, owner clauses, command recipes and lint paths | Directly inspected relevant sections | Authority, future verification selection and document-only check admission | Installed-only readiness, fixing disabled, external caches/results | E-050 historically; E-054/E-055 for this refinement; no Markdown added to executable inputs |
| Remaining entries / ignored roots | Metadata-only / excluded as declared in Section 1 | No movement proposed | Text searches do not establish absence of dynamic/external consumers | Next boundary: newly encountered owned-start, attachment, namespace cleanup or exact-selector consumer |

## 3. Module Boundary Diagnosis

The problem is mixed ownership and test coupling inside an existing support
package, not a missing platform layer. Keep `s3test` as the package boundary.
Responsibility-oriented files are sufficient; do not create a generic fixture
manager or force PostgreSQL and S3 into a shared lifecycle abstraction.

| Responsibility | Current location | Recommended owner / disposition | Evidence / finding / contract |
| --- | --- | --- | --- |
| Container acquisition, partial settlement, TLS material and physical close | Mixed `Harness`, three start wrappers | **Split** into `OwnedStore`; testservices keeps composition | E-049; F-027; C-017/C-020 |
| Borrowed client, attachment reuse and local transport release | Shared singleton plus shared/attached flags | **Keep/simplify** in `Harness`; separate mechanical extraction (S-021) from bound-admission correction (S-025) | E-049/E-054; F-028; C-017 |
| Mutation-readiness algorithm and diagnostics | Existing shared probe plus repeated test scaffolding | **Keep** algorithm; **split** local test dependencies and owner-specific wiring assertions | E-049; F-029/F-030; C-018/C-020 |
| Namespace ownership, package admission, reset and cleanup | Harness plus existing cleanup module | **Keep** lifecycle; **move** package tests to coherent namespace group and require typed absence plus sibling payload proof | E-049/E-054; F-029–F-031/F-033; C-019 |
| Source assertions, exact selectors and evidence classification | Two mixed unit umbrellas, separate origins row and split real rows | **Consolidate** by observation boundary; integration class for real service work | E-049/E-050; F-032; C-021 |
| Provider, scheduler, PostgreSQL and performance policy | Independent existing owners | **Defer** redesign; retain affected regression boundaries | User-selected scope; no independent change slice |

## 4. Public Contract and Behavior Freeze Map

Freeze owner-required guarantees, not test implementation details. No HTTP,
WebSocket, domain mutation/query, saved-view, projection, revision, public evidence
identity or UI-selector change is proposed. Public Make commands and evidence
schemas retain their current contracts. Internal support APIs and row identities
may change atomically without aliases.

| Contract / surface | Observed implementation and adopted owner | Required behavior / disposition | Inspected witnesses / prerequisites / trace |
| --- | --- | --- | --- |
| C-017 — ownership and attachment | Mixed Harness, process singleton and StopShared; REQ-804/410, §11 | Preserve explicit dependencies/attachment, independent physical owners, partial recovery authority and borrower nontermination. Separate types and bind cached admission to its inputs; changed bindings cannot bypass admission. | Startup/attachment tests, testservices callbacks, process TestMain and native boundaries; current characterization required by RB-008; F-027/F-028; S-020/S-021/S-025; V-025/V-026/V-031/V-029 |
| C-018 — mutation readiness | One probe used by owned, attached, broker and package paths; §11.4 and startup retry table | Preserve put, exact-size head, delete and not-found proof; bounded same-lane polling, terminal capability rejection, separate startup/readiness windows, parent cancellation and cleanup on every outcome. No timeout expansion or product retry. | Mutation probe, owned-start and package admission cases; add deterministic cancellation/failure witnesses as needed; F-029/F-030; S-021/S-022; V-026/V-027/V-029 |
| C-019 — namespace lifecycle | Existing reserved ownership, package lock/quarantine, stable reset and cleanup algorithm; REQ-804 and §11 | Preserve exact ownership before risky acquisition, unique scopes, package reuse/concurrency, failed-generation rejection, independent empty proof, sibling protection, stable reset and final deletion. No cleanup success inferred from a fake's error path or an arbitrary read error; absence requires typed not-found, and surviving siblings retain their payloads. | All cleanup/package tests and four real-service tests; F-029/F-031/F-033; S-022; V-025/V-027/V-029 |
| C-020 — security and failure evidence | TLS roots/owner keys, S3 transport, typed readiness projection and testservices summaries; Core 04 §1.3, REQ-414, §9 | Preserve verified TLS, exact secure origins, private keys/credentials, primary versus secondary errors, timeout exit 3, cancellation and capability classifications, bounded public/retained diagnostics and correct journal destination. | Origins/configuration, cleanup diagnostic, setup-envelope and composition publication assertions; F-027–F-030; S-021/S-025/S-022; V-026/V-031/V-027/V-029 |
| C-021 — selection and assurance | Five S3 rows; real ResetBucket case labelled unit; §§3/8 and REQ-410 | Preserve meaningful independent qualification contexts and exact service/capability declarations. Give each assertion one coherent selector; retired names need no compatibility routing. A subset is not complete-owner acceptance. | Authored selectors, native nested selections, app/platform consumer rows; F-032; S-020/S-023/S-024; V-025/V-028/V-029 |

## 5. Coupling and Boundary Findings

`must_fix` is limited to the named slices and acceptance conditions, not unrelated
cleanup. Findings describe the pre-refactor defects. E-060 reproduced cached-binding
bypass; E-059–E-063 record their fixes and qualification without adopting the
old tests' incidental constraints. F-027 is resolved by E-059, F-028 by
E-059/E-060, F-029/F-030/F-031/F-033 by E-061, and F-032 by E-062; E-063
qualified all successors together at that checkpoint. E-064 records the
follow-up review, its missed failure paths and their focused requalification.
F-034–F-037 are resolved by E-064; these corrections do not rewrite the earlier
checkpoint outcomes.

| Finding / classification | Evidence and consequence | Required action / compatibility disposition | Contract / slice / validation |
| --- | --- | --- | --- |
| F-027 / must_fix | Harness holds Container, TLS directory, client, shared/attached flags and bucket state; owned entrypoints return that same type. StopShared describes owner shutdown although ordinary startup only attaches. | Give physical ownership one type and constructor; keep borrowed local transport cleanup separate. Migrate direct callers, remove wrappers and StopShared; no forwarding stubs. | C-017/C-020; S-021; V-026/V-029; E-049 |
| F-028 / must_fix | Startup/readiness/package operations and testservices object-store composition use mutable package globals. StartShared returns the cached handle without comparing the new binding; later guards/events reread ambient values. | S-021 introduces private instance dependencies and independent test admission objects. S-025 separately corrects cached-binding bypass, ambient event/dependency reads and terminal-failure retention under REQ-804/410. No public injection API or generic service registry. | C-017/C-020; S-021/S-025; V-026/V-031/V-029; E-049/E-054 |
| F-029 / must_fix | Fake readiness DeleteNamespace removes namespace/objects even when recordAndFail returns an error, allowing no-residue assertions to succeed on a purported failure. | Default failure leaves state unchanged; explicit cases model response loss after mutation. Deleting a nonempty namespace rejects without recursively erasing objects. Assert retained uncertainty/cleanup outcome rather than fabricate success; preserve the earlier readiness failure as primary. | C-018/C-019/C-020; S-022; V-027/V-029; E-049 |
| F-030 / must_fix | Wait-strategy test matches dependency description text; cleanup tests freeze exact batch/list counts; short real timers and repeated readiness setup mix protocol, controller and publication responsibilities. | Assert mapped-endpoint admission, deadline separation, bounded batches, complete traversal and postconditions. Use local deterministic fault seams and small readable case tables; retain true boundary constraints and distinct publication witnesses. | C-018/C-020; S-022; V-027/V-029; E-049 |
| F-031 / should_fix | Package quarantine sits in cleanup-algorithm tests; repeated real roundtrip/prefix setup and separate readiness/package scaffolds obscure ownership. | Group namespace lifecycle assertions; consolidate redundant real setup while preserving an independent stable-reset scenario. Keep narrow cleanup and readiness fakes separate: they model different protocols. | C-019; S-022; V-027/V-029; E-049 |
| F-032 / must_fix | S3 assertions span two mixed unit umbrellas and origins, while a real reset has unit evidence class. Native selected-service tests reference exact existing row IDs. | Three coherent service-free groups (ownership/attachment, readiness/diagnostics, namespace contracts), plus real integration namespace/reset evidence; atomic authored/generated/nested-selector cutover. No aliases or duplicate execution to preserve old names. | C-021; S-023; V-028/V-029; E-049/E-050 |
| F-033 / must_fix | `TestCleanupPrefixPreservesSiblingObjects` and `TestResetBucketPreservesNamespaceAndProvesMutation` accept any HeadObject error as deletion proof; transport or authorization failure can satisfy the assertion. The sibling check proves availability but not payload preservation. | Require the existing typed not-found classification for deleted objects, fail on other errors, and compare surviving sibling bytes with their original payload. Require successful new mutation after stable reset. | C-019/C-021; S-022; V-027/V-029; E-054 |
| F-034 / must_fix | Readiness cancellation during polling sets outcome `cancelled` but drops the cancellation from the returned cause; package setup then reports `harness/fixture_error`. The earlier test cancelled before readiness instead. | Retain polling termination alongside the original probe error; deterministic cancellation between probes must preserve the interrupted envelope and clean probe state. | C-018/C-020; S-022; V-033; E-064 |
| F-035 / must_fix | The readiness controller counts terminal cleanup errors as primary capability failures, counts cleanup timeouts as probe timeouts, and can hide a transient cleanup failure behind a later successful attempt. Pure probe tests did not exercise the controller boundary. | End admission on uncertain cleanup, retaining the original probe stage/cause/class and secondary cleanup status. Qualify transient, deadline and cancellation cases through the controller and setup envelope. | C-018/C-019/C-020; S-022; V-033; E-064 |
| F-036 / should_fix | The cancelled-waiter test can cancel before its goroutine enters admission; it can pass without exercising waiting. Runtime/lease/mode binding changes and post-deletion publication failure also lacked direct negative witnesses. | Coordinate on the wait context's Done observation, cover all required binding categories, and prove journal-write failure remains failure after a real deletion. Keep existing rows and narrow fakes; add no runtime test hooks or routing aliases. | C-017/C-019/C-020; S-025/S-022; V-033; E-064 |
| F-037 / must_fix | A missing fixture CA passes attachment-shape validation. The shared transport then resolves its process-level system trust, potentially using ambient inputs rather than the captured fixture binding. Existing incomplete-binding cases omitted trust. | Require an explicit fixture CA in attachment/recovery binding before readiness or network acquisition. Keep system-trust behavior in the production transport unchanged; managed fixture owners already publish the CA. | C-017/C-020; S-025; V-033; E-064 |

The shared mutation probe, exact-prefix cleanup algorithm, borrowed client cache,
package reuse, broker publication checks and real provider witness each have a
useful purpose. Keep them; simplification is not a deletion quota. Scoped source/routing changes are now authorized. No adopted-owner conflict or
necessary specification amendment was found.

## 6. Refactor Workstreams

The prerequisites in this table are authoritative. Execution order is WF-20 →
WF-21 → WF-25 → WF-22 → WF-23 → WF-24; existing IDs retain their identity.
There is no optional branch. Save each exit checkpoint before its successor.
Every slice migrates affected selectors with its source, before final routing
consolidation where necessary. S-025 is the separate owner-required correction;
mechanical movement must not silently include it.

| Workflow | Name / class | Prerequisites | Goal / likely files | Validation | Checkpoint and binary exit |
| --- | --- | --- | --- | --- | --- |
| WF-20 | Characterization / root | Current implementation authorization; E-057 revalidation | Complete assertion/caller ledger and focused baseline; s3test and consequential callers | V-025 | All 30 tests and consequential assertions have dispositions, exact baseline roots and classified gaps; known defects are not frozen as requirements |
| WF-21 | Ownership and local dependencies / chain | WF-20 exit | OwnedStore/Harness split, private operations and direct testservices/process migration | V-026 | Ownership, partial settlement and local cleanup qualified; obsolete lifecycle wrappers/global test substitutions removed; admission correction remains separately tracked |
| WF-25 | Bound attachment correction / chain | WF-21 exit | Immutable attachment/dependency/event inputs, reuse validation, terminal failure and cancellable waiting; recovery uses saved binding | V-031 | Negative binding, failure-retention and cancellation witnesses pass; no fresh admission in recovery cleanup or ambient journal redirection |
| WF-22 | Assertion consolidation / chain | WF-25 exit | Repair fake semantics, strengthen typed absence/sibling proof, organize readiness/namespace tests and consolidate real setup | V-027 | Every retained guarantee has qualified successor coverage; incidental string/count/mock constraints retired; arbitrary errors cannot prove deletion |
| WF-23 | Routing consolidation / chain | WF-22 exit | Authored family/verification inputs, nested selections and generated projections | V-028 | Three service-free groups and two real integration scenarios have exact selectors and correct capability/evidence declarations; no duplicate/dangling/alias rows |
| WF-24 | Final qualification and handoff / chain | WF-23 exit | Finalize, freeze source, qualify compatible portfolio and close handoff | V-029 | Required distinct boundaries pass; canonical evidence, cleanup and source limits recorded; no broader acceptance inferred |

## 7. Proposed Refactor Slice Plan

| Slice / kind | Findings / contracts | Entry and readiness | Intended change / validation | Rollback and binary completion |
| --- | --- | --- | --- | --- |
| S-020 / characterization | F-027–F-033; C-017–C-021 | **complete**, E-058 | Complete assertion/subcase and caller dispositions, establish V-025 baseline, record missing witnesses and known defects separately | Do not encode incidental old behavior as a baseline obligation. Exit is WF-20 checkpoint; unexpected failures require diagnosis |
| S-021 / structural preservation | F-027/F-028; C-017/C-018/C-020 | **complete**, E-059; RB-008 resolved | Separate physical and borrowed lifetimes, localize dependencies and migrate callers; V-026 | Revert source/tests/callers/selectors together, without aliases; ownership and settlement witnesses pass without claiming the S-025 correction |
| S-025 / owner-required admission correction | F-028; C-017/C-020 | **complete**, E-060; RB-010 resolved | Bind inputs and journals, reject changed cached bindings, retain terminal failures and cancel waiting locally; qualify recovery's saved-binding path; V-031 | Revert the correction and its witnesses together if unsafe; keep F-028 open. Exit requires negative-case coverage; no product API/data migration |
| S-022 / test and fixture simplification | F-029/F-030/F-031/F-033; C-018/C-019/C-020/C-021 | **complete**, E-061 | Correct fake semantics, consolidate setup and assertion groups, require typed absence and sibling bytes; V-027 | Restore affected assertions/fixtures together if successor gaps emerge; no compatibility mock behavior; exit is WF-22 checkpoint |
| S-023 / routing cutover | F-032; C-021 | **complete**, E-062 | Consolidate semantic rows and real integration classification; retain verification identities with unchanged meaning; regenerate through Make; V-028 | Restore authored inputs/source and regenerate together; every intended source selector resolves once in the catalog |
| S-024 / final qualification | All current findings/contracts | **complete**, E-063; RB-009 resolved | Finalizer before broader verification, frozen compatible portfolio and saved completion; V-029 | Keep incomplete work explicitly incomplete; no historical-result rewrite or unsupported acceptance claim |

All implementation slices are authorized subject to their entry gates. S-025 corrects existing
REQ-804/410 obligations; no normative amendment or user-facing compatibility
migration is proposed. Introduce its negative witnesses before changing admission,
record expected failures against the defect, and require passing corrected
behavior before WF-22. Historical success cannot satisfy these gates.

### Implemented ownership and interface

- `OwnedStore` owns container acquisition, startup settlement, TLS material,
  broker probe-bucket lifetime and physical close. One
  `StartOwned(ctx context.Context, options StartOptions) (*OwnedStore, error)`
  returns this owner, including a partial handle when settlement is uncertain.
  Existing before/acquired/settled callbacks retain
  recovery ordering and become owner-typed; testservices retains orchestration.
- `Harness` owns borrowed client access and namespace operations. Composition is
  explicit, without embedding that exposes physical shutdown. Retain `Start` and
  `StartShared` for ordinary consumers. Local transport release must not terminate
  the container, delete the broker bucket, remove owner keys or erase quarantine.
- Replace `StartOwnedWithLabels`, `StartOwnedWithOptions` and `StopShared` at every
  direct caller, then delete them. Process teardown releases its borrowed client;
  it does not report termination of a shared testcontainer.
- S-021 extracts a private process admission object for necessary reuse; tests
  construct independent objects. S-025 then binds a copied set of
  attachment, dependency and event-destination inputs at entry. Validate every
  call against the admitted binding; changed inputs fail instead of using stale
  credentials or a different journal.
  Successful reuse and terminal failure are explicit; waiters honor cancellation
  without shutting down an owner or cancelling another caller's admission.
  Do not add cross-service registries or replacement-generation machinery.
- Keep operation values private and limited to startup, attachment, readiness or
  package lifecycle. Testservices uses its existing dependency style with an
  instance-local object-store seam; no exported fake constructor is necessary.
- Preserve the one mutation probe and cleanup algorithm. Keep the error-returning
  package admission core; its testing.TB wrapper only publishes setup failure and
  registers successful release. Isolate test cases without replacing global
  functions or parsing fatal text.
- A recovery cleanup caller retains its saved exact binding and ownership proof;
  migrating its Harness construction must not trigger fresh admission, provisioning
  or cleanup of unrelated resources. Owner TLS/key cleanup remains recoverable
  through the existing journal/runtime, not a new evidence schema.
- Keep only three service-free assertion groups: ownership/attachment,
  readiness/diagnostics, and namespace lifecycle. Consolidate real roundtrip,
  pagination, sibling protection and final cleanup into one namespace scenario;
  retain stable reset as a separate real scenario. Both are integration evidence.
  Keep publication and native process/lifecycle witnesses distinct.

### Assertion successor and retirement ledger

Every baseline top-level test is named once below (30 total). These are assertion
dispositions; executed successor evidence is recorded at the slice checkpoints
and final source in E-063. The WF-20 supplement below completes
subcase-level mapping and consequential caller assertions before movement; counts and test
names themselves are not acceptance requirements. Each removed assertion must
name its successor guarantee and observation boundary, or explain why it only
constrains incidental behavior. Missing coverage is a prerequisite, not a reason
to preserve misleading assertions or add compatibility behavior.

| Existing test | Successor / retirement disposition | Contract / validation |
| --- | --- | --- |
| `TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness` | Local admission cases retain explicit binding and no physical acquisition in S-021; S-025 qualifies same-binding reuse and changed-binding rejection | C-017; V-026/V-031 |
| `TestStartSharedRejectsMissingAttachment` | Missing/incomplete attachment and dependency rejection before I/O/publication; recheck against captured inputs in S-025 | C-017; V-026/V-031 |
| `TestBucketNamesAreUniqueAcrossSimulatedProcesses` | Namespace identity cases retain valid bounded names and cross-process noncollision | C-019; V-027 |
| `TestObjectStoreContainerWaitStrategyOnlyWaitsForPortMapping` | Retire description-string match; preserve mapped-endpoint-before-readiness behavior and declared startup bound in ownership cases | C-017/C-018; V-026 |
| `TestSeaweedFSS3AllowedOriginsDefaultSupportsLoopbackBrowserPortRange` | Ownership/configuration cases retain exact secure loopback origins and reject wildcard/plaintext assumptions | C-020; V-026 |
| `TestOwnedObjectStoreAppliesContainerConfiguration` | One owner startup group retains labels, single-master topology and TLS configuration | C-017/C-020; V-026 |
| `TestOwnedObjectStoreReadinessTimeoutIsTerminalForTheStartedLane` | Owner terminal-failure matrix: no replacement after readiness starts; exact cleanup and bounded evidence | C-017/C-018; V-026 |
| `TestOwnedObjectStoreCleanupFailureRemainsSecondaryTypedEvidence` | Same owner group retains partial handle, memoized settlement result, primary cause and diagnostic redaction | C-017/C-020; V-026 |
| `TestOwnedObjectStoreRetriesOnlyPreReadinessStartupAndUsesFreshReadinessContext` | Preserve allowed pre-readiness retry and startup/readiness budget separation; parent cancellation still applies | C-018; V-026 |
| `TestOwnedObjectStoreDoesNotRetryAuthenticationReadinessFailure` | Terminal capability case in owner matrix; no new container, preserved cleanup/classification | C-018/C-020; V-026 |
| `TestObjectStoreMutationProbeRetriesTransientStagesOnSameLane` | Readiness controller cases retain bounded transient put/head retries on one lane, without short-timer call-count assumptions | C-018; V-027 |
| `TestObjectStoreMutationProbeRejectsCapabilitiesImmediately` | Readiness protocol/controller cases retain immediate typed rejection and owned-probe cleanup | C-018/C-020; V-027 |
| `TestObjectStoreMutationProbeDeadlineRetainsCleanupEvidence` | Deterministic deadline case with bounded terminal failure and truthful secondary cleanup | C-018/C-020; V-027 |
| `TestObjectStoreMutationProbePreservesPrimaryFailureDuringCleanupFailure` | Repair fake deletion semantics; assert original put failure, cleanup failure and unresolved state; separate response-loss case | C-018/C-019/C-020; V-027 |
| `TestAttachedObjectStoreUsesMutationProbeContract` | Attached wiring case proves shared probe use, unique child scope and preservation of broker parent bucket | C-017/C-018; V-027 |
| `TestPackageBucketAdmissionProbesWithoutRemovingTheBucket` | Package wiring case retains mutation sequence/size proof, clean probe and preserved reusable bucket | C-018/C-019; V-027 |
| `TestObjectStoreMutationProbeRejectsSizeMismatch` | Protocol negative case retains exact size verification and cleanup outcome | C-018; V-027 |
| `TestCreateBucketRejectsOmittedServiceBeforeClientAcquisition` | Bound dependency guard rejects before client/network acquisition; captured-input correction belongs to S-025 | C-017; V-026/V-031 |
| `TestPreparePackageBucketCoreClassifiesPutDeadlineAndCleanupFailure` | Package failure matrix retains error-returning core and timeout envelope; publication stays independently tested | C-019/C-020; V-027 |
| `TestPreparePackageBucketCoreRejectsCapabilityAndCancellation` | Package failure matrix retains terminal capability/cancellation classes and no successful release | C-019/C-020; V-027 |
| `TestPreparePackageBucketCoreReusesPackageBucket` | Namespace lifecycle group retains same package identity, reset-before-reuse and readiness-before-admission | C-019; V-027 |
| `TestPreparePackageBucketCoreAdmitsDistinctPackagesConcurrently` | Independent-package case retains concurrent admission and separate namespaces; no global test hooks | C-019; V-027 |
| `TestHarnessStartsSeaweedFSS3AndRoundTripsObjects` | Consolidated real namespace lifecycle retains exact payload roundtrip and idempotent final cleanup; retire misleading implication that Start owns startup | C-017/C-019; V-027 |
| `TestCleanupPrefixPreservesSiblingObjects` | Consolidated real namespace lifecycle requires typed not-found for removed keys and exact surviving sibling payloads; other read failures fail the assertion | C-019; V-027 |
| `TestResetBucketPreservesNamespaceAndProvesMutation` | Keep independent real stable-reset scenario: same bucket, typed not-found for prior objects, new mutation/roundtrip succeeds; route as integration | C-019/C-021; V-027/V-028 |
| `TestObjectCleanupBatchesAndProvesExactEmptyNamespace` | Retain 0/1/boundary/multi-page cases, bounded batches and fresh empty proof; retire unnecessarily exact request totals | C-019; V-027 |
| `TestObjectCleanupRejectsUncertainCompletion` | Retain list/delete/verification faults, partial errors, residue, malformed pages/tokens, out-of-prefix keys, cancellation and absent bucket | C-019/C-020; V-027 |
| `TestPackageBucketFailedGenerationCannotBeReadmitted` | Move to package lifecycle group; retain create/reset/readiness/cancelled-reset poisoning, cleanup identity and no repeated mutation | C-019; V-027 |
| `TestPinnedObjectStoreBatchCleanupTraversesMultiplePages` | Consolidated real namespace scenario retains more than one provider page, <=1,000-key batches, fresh empty proof and exact sibling payload; no fixed two-request requirement | C-019; V-027 |
| `TestOwnedBucketReservesCleanupBeforeClientFailure` | Namespace ownership case retains journal reservation before acquisition and no false success event | C-019/C-020; V-027 |

Keep testservices failure-summary/publication assertions distinct from pure probe
assertions: they qualify public classification, retained serialization and exit
propagation. Keep native owner/borrower/selection witnesses distinct from real S3
protocol coverage. Add fault cases for incomplete bindings, separate instances
with conflicting ambient inputs, pending-waiter cancellation, terminal admission
failure, partial acquisition/publication failure and borrower local-close safety
where the baseline lacks them. A preserved test function is not proof that its
assertions cover these boundaries.

### WF-20 assertion and caller supplement — E-057

The preceding ledger's row numbers (1–30, in order) identify the baseline tests.
These details complete the assertion-level dispositions. No planned successor is
credited as passing before its own checkpoint. Known F-028/F-029/F-030/F-032/F-033
defects are not requirements even when their existing tests pass.

| Ledger rows | Assertions/subcases carried to the named successor | Retirement or missing witness to qualify |
| --- | --- | --- |
| 1–2 | Explicit endpoint/credentials/TLS/probe binding, one admission and no physical start; missing attachment returns no published fixture | Container-field inspection becomes structural absence of shutdown capability plus borrower-close witness. Add incomplete and changed bindings, retained failure and cancelled waiters in S-025 |
| 3 | DNS-safe names, <=63 characters, alphanumeric ends, distinct repeated and simulated-process scopes | Literal sample names/count eight are data, not requirements |
| 4–6 | Startup bound, mapped endpoint available before readiness, labels, single standalone master, exact HTTPS loopback origins and no wildcard/plaintext | Retire wait dependency's description text. Add TLS file confinement and owner/borrower lifetime assertions at S-021 |
| 7–10 | Readiness deadline/capability rejection never replaces a started lane; single settled close, partial handle on uncertain cleanup; internal secondary cause, public redaction and bounded phase/stage/outcome/count fields; permitted transient pre-start retry with fresh readiness budget and parent context | One terminal acquisition/close is causal evidence, not an incidental total. Add before/acquired/settled ordering, cancellation during acquisition and publication failure; keep readiness correction separate |
| 11–14 | Transient put/head recovery on same lane; terminal create capability rejection; deadline and cleanup evidence; primary put cause survives secondary namespace-delete failure | Replace timer-sensitive minimum retries by coordinated transient recovery/deadline cases. Retire false no-residue assertion after failed namespace deletion; separately model mutation then response loss |
| 15–17 | Attached probe uses existing broker namespace; package probe preserves reusable namespace; put → exact-size head → delete → typed absence, no probe residue; size mismatch fails with cleanup | Protocol sequence is causal. Namespace create/delete prohibition protects borrowed ownership. Add typed non-absence failures; failed fake operations cannot silently mutate |
| 18–20 | Omitted object-store dependency before client acquisition; package core returns no release on timeout; infra/readiness_timeout stage/count/cleanup envelope; capability→harness and cancellation→interrupted | S-025 binds guards to saved inputs. Testservices separately proves serialized/public exit behavior; pure probe tests need not duplicate that publication boundary |
| 21–22 | Stable package identity, create once then reset-before-readmission, readiness each admission; distinct packages can hold different leases concurrently | Preserve per-package serialization, not a global lock or global hooks. Use completion channels; timeout only bounds a deadlock failure |
| 23–25 | Real payload roundtrip, repeated final cleanup, owned-prefix removal, sibling protection; independent stable reset keeps bucket and admits a new put/get | Merge roundtrip/prefix into pagination scenario. Replace arbitrary HeadObject errors with typed absence; compare sibling and new-mutation bytes, not mere availability |
| 26 | Empty, single, 1,000-boundary and 2,507-key cleanup; no owned residue, sibling retained, each batch 1–1,000, independent empty proof | Retire exact batch/list totals. Observe final independent listing and bounded complete deletion rather than a prescribed partition |
| 27 | `list`, `delete`, `verify` retain cause; `partial` redacts provider details; `residue`, `nil_page`, `nil_result`, `missing_truncation`, `missing_token`, `duplicate_token`, `cyclic_token` reject uncertainty; `outside_prefix`, `empty_key`, `common_prefix` do not mutate; `cancel_list`, `cancel_delete` retain cancellation; `missing_bucket`, `missing_at_verification` accept typed absence | Preserve finite pagination/cycle rejection; do not freeze a universal three-request ceiling. Cleanup fake remains narrow and separate from readiness |
| 28 | `create`, `reset`, `readiness`, `cancelled_reset`: failed generation retains cause/cleanup identity, releases lock, cannot mutate or retry upon reuse | Move to namespace lifecycle group; no compatibility reset of quarantine |
| 29–30 | Real 1,007-key traversal and sibling survival; journal reserves cleanup identity before failed client acquisition, no false bucket-created event | Merge row 29 into real namespace scenario, bounded batches replace exact [1,000,7]; row 30 keeps journal truth and gains saved-destination failure cases |

| Consequential caller / exact qualification | Assertions and migration boundary |
| --- | --- |
| testservices lifecycle row (A) | Service-specific S3 2-minute acquisition bound, original caller context, before/acquired/settled partial journal ordering, probe publication, primary/secondary readiness diagnostics, retained serialization and public exit 3. Replace only S3 global substitutions with a local seam |
| native physical/borrowed/selected-service rows (B) | Physical lifecycle and persistent borrower nontermination are independent of unit fakes. Selected-service nested rows must follow any renamed S3 selector in the same slice |
| app support guard and evidence-upload process rows (C) | Explicit validated service dependency before fixture construction; standalone process still uploads real evidence. TestMain releases local S3 transport after StartShared; it gains no physical owner |
| managed object-store adapter row (D) | Production adapter exercises the managed verified-TLS fixture binding; keep this independent of fixture protocol unit tests |
| browser recovery cleanup in testservices | Use the saved environment and exact bucket ownership proof, without fresh admission or provisioning; capture journal destination. Qualify binding failures/publication failures in S-025 |
| appsupport and performancefixturelifecycle | Retain borrowed acquisition, scoped namespace cleanup and existing performance policy. Their Harness references must not gain owner methods; no unrelated hook or scheduler redesign |

### Guarantee and retirement decisions

| Required guarantee / owner | Meaningful successor and observation boundary | Incidental constraint retired / deletion gate |
| --- | --- | --- |
| Physical owner versus borrower; REQ-804/402 | Ownership unit cases plus native physical/borrowed lifecycle; partial ownership, cancellation, one settlement and borrower local-close safety | Mixed Harness flags, redundant start wrappers and StopShared are not compatibility contracts; V-026 qualifies their removal |
| Exact attachment and recovery binding; REQ-804/410 | Independent admission instances; incomplete/changed binding rejection, same-binding reuse, terminal failure, cancelled waiter and saved-binding cleanup | Cached success cannot bypass new binding checks; ambient environment or journal mutation is not a supported rebind; V-031 |
| Readiness mutation and deadlines; §11.4 | Shared probe protocol cases plus owned/attached/package wiring; exact size, typed not-found, bounded same-lane polling and primary/secondary failures | Dependency description strings and scheduling-sensitive retry counts are removed; retain causal ordering and declared budgets; V-026/V-027 |
| Exact cleanup and truthful uncertainty; REQ-804/402, §11 | Narrow cleanup fake covers bounded batches, malformed pages/tokens, partial failures, cancellation, fresh emptiness proof and no out-of-prefix mutation | Exact request totals and fake recursive namespace deletion are removed. Explicit mutation-then-error cases do not redefine default failure semantics; V-027 |
| Isolation, stable reset and data preservation; REQ-804, §11 | Real namespace scenario proves roundtrip, multi-page cleanup, typed absence and unchanged sibling bytes; separate reset proves same identity and new mutation | Duplicate real setup and acceptance of arbitrary HeadObject errors are removed only with these successor observations; F-033/V-027 |
| Package reuse and quarantine; REQ-804, §11.4 | Namespace lifecycle cases prove reset-before-reuse, readiness-before-admission, distinct-package concurrency and failed-generation rejection | Global substitution scaffolds and misplaced quarantine tests are consolidated; failure cause and cleanup identity remain; V-027 |
| Secure transport and truthful publication; Core 04 §1.3, REQ-414, §9 | Configuration/diagnostic unit cases plus distinct testservices serialization/exit and managed-adapter/process witnesses | Repeated protocol assertions at publication boundaries may be removed; serialization, classification, secret confinement and real transport witnesses may not; V-026/V-027/V-029 |
| Exact selection and declared dependencies; §§3/8, REQ-410 | Three service-free groups, two real integration scenarios and nested selection witness | Historical row names, duplicate routes and real-reset-as-unit classification are retired atomically; existing verification identities remain when their meaning is unchanged; V-028 |

Consolidation is justified by guarantees and independent observation boundaries,
not test-count reduction. Do not replace narrow fakes with a general S3 simulator,
expose a public injection API, or introduce a generic clock framework. Use local
controllable waits/contexts where timing matters; production deadlines remain
unchanged. WF-20 establishes the baseline before source or test movement.

## 8. Validation Plan

**The original implementation passed WF-24 at E-063; the follow-up review is
qualified separately by V-033/E-064.** Public discovery was
revalidated at entry (E-057): `make help ALL=1` and
`make task-guide OWNER=<owner>` for `harness.browser`, `platform.objectstore`,
and `app.server`, with installed-only preparation and diagnostics/accounting off.
Source/catalog inspection supplies the exact
starting selections below. WF-20 revalidated them at E-058; E-063 records the
qualified final successors. No historical pass is promoted to the current source.

V-024 and V-030 retain their historical documentation-check identities. V-031
qualifies the separate admission correction; V-032 covers the prior document
refinement only; E-063 records this implementation's final documentation checks. No historical validation ID or result is repurposed.

| Validation | Findings/contracts/slices / stage | Exact selection or command / expected artifact | Gate |
| --- | --- | --- | --- |
| V-025 | All; S-020; pre-move characterization | Groups A–D below through Make; exact roots, source identity, native exits and assertion ledger | Required baseline before S-021; classify known defects and unexpected failures separately, without freezing incidental behavior; diagnose before rerun/broadening |
| V-026 | F-027/F-028; C-017/C-018/C-020; S-021; post-slice | Ownership/attachment successors in A, testservices row, applicable B/C/D boundary witnesses | No mixed shutdown capability or global hook substitution; uncertain settlement remains recoverable. S-025 owns the separate binding correction |
| V-031 | F-028; C-017/C-020; S-025; post-correction | Attachment successors in A plus testservices recovery/publication and applicable B/C witnesses; exact rows recorded after migration | Changed/incomplete bindings fail before I/O; reuse retains terminal failure; waiter cancellation is local; saved recovery binding and event destination remain exact |
| V-027 | F-029–F-031/F-033; C-018–C-021; S-022; post-slice | Readiness/namespace successors in A; real namespace/reset scenarios; testservices publication witness | Truthful fake failures, typed absence, unchanged sibling payload and successful post-reset mutation qualified before redundant assertions are deleted |
| V-028 | F-032; C-021; S-023; post-slice | `make generate`; `make source-check CHECK=generated`, `CHECK=schema`, `CHECK=catalog`, `CHECK=identity`; `make lint CHECK=go`, `CHECK=backend-boundary`; selected-service row | Authored inputs/projections and exact selectors agree; real reset has integration evidence class |
| V-029 | All; S-024; final regression | `make agent-finalize`, review output, freeze source; compatible successor A–D union plus V-028 static checks; audit every exact root with `make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all` | Native exits, valid evidence and required cleanup pass; do not claim whole-owner, release or performance acceptance |
| V-030 | Historical document save only | Earlier documentation checks, retained verbatim in E-051/E-052 | Historical result; not this refinement's validation |
| V-032 | Historical document refinement only | External-output installed-only `make lint-markdown`, direct tables/links/IDs/dependency/assertion review, `git --no-optional-locks diff --check` and baseline comparison | E-055/E-056; configured lint excludes this handoff, so direct review is required; no product assurance claim |
| V-033 | F-034–F-037; post-completion review | Negative ownership/readiness cases, five S3 rows plus publication, finalizer/freeze, Go lint, generated/catalog checks and native borrower boundary; exact-root audits and documentation review | Implementation and documentation checks passed, E-064; explicit fixture CA, preserved primary/secondary failure and cancellation, terminal cleanup uncertainty, deterministic waiting and truthful deletion publication; prior E-063 results remain historical |

All implementation commands run from `/home/jochi/code/cartulary`, with fresh explicit
`CARTULARY_TEST_RUN_ID`, repository-pinned prerequisites and native session/exit
capture. Use `CGO_ENABLED=0` for the existing nested Go-runner profile; retain
current resource claims, fixture scopes and deadlines. Do not widen service
acquisition or make unavailable infrastructure a passing skip. Current-source
pre-move characterization passed at E-058; each successor checkpoint was saved
before its dependent changes.

### Exact starting selections

A — all five existing S3 rows plus the composition/publication witness:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.s3test_helper_contract,harness.browser.unit.service_free_contract_assertions_for_object_stor_aac986f098,harness.browser.unit.object_store_secure_origins,harness.browser.integration.object_store_fixture_roundtrip,harness.browser.unit.managed_service_assertions_for_object_store_fixt_7b174abeda,harness.browser.unit.testservices_lifecycle_contract
```

B — distinct native physical ownership, persistent borrowing and nested selection:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.native_borrowed_fixture_lifecycle,harness.browser.integration.native_selected_service_lifecycle
```

C — application dependency guard and evidence-upload process consumer:

```sh
make test-slice OWNER=app.server ROWS=app.server.support_unit.incident_create_commit_fault_requires_validated_efdc344831,app.server.process.the_standalone_server_process_can_create_evidenc_4ca93ab308
```

D — production adapter against the managed fixture binding:

```sh
make test-slice OWNER=platform.objectstore ROWS=platform.objectstore.support_integration.managed_service_object_store_binding_4d6aaf4bd9
```

These are exact baseline commands, not permanent names. Record qualified
successors when A is consolidated; migrate nested uses of its old selectors in
B in the same slice. Recompute the final union against target plans so nested or
already compatible execution is not repeated without a missing obligation or
distinct qualification purpose. Changed evidence serialization would additionally
require its existing evidence-accounting owner selection; this plan changes no
schema or serializer. Newly discovered affected callers get exact row mappings
before their source changes, not an automatic full-suite run.

Full CI, visual/golden/UI-review, benchmarks, release qualification and preceding
WF-33 are excluded. No production behavior in those scopes is planned to change.
Finalization left `RESULTS_DIR` unset because there is no compatible successful
full warm-check root. E-063 records the retained-maintenance skips; the scoped
portfolio does not substitute for full-owner or release qualification.

| Invocation group | Side effects / admission | Current result |
| --- | --- | --- |
| Discovery and source inspection | Public help/task guides and file reads; no tests/services/generation | Revalidated for implementation at E-057; E-053/E-054 retain the preceding document-only discovery |
| V-025–V-029 and V-031 | Scoped implementation, services, test results, generation and finalizer writes | V-025/V-026/V-031/V-027/V-028 passed at E-058–E-062; V-029 passed at E-063 |
| V-033 | Five existing S3 source/test files and this handoff; focused tests and required finalization/checks | Passed, E-064: six compatible audited graph roots, seven selected test rows, two generated-policy rows; separate documentation lint/direct review passed |
| V-030 | Historical invocation | E-051/E-052 retained; not current validation |
| V-032 | Existing workspace lock; installed-only Node/frontend readiness; lint fixing disabled; result/cache outputs external | Passed, E-055/E-056; direct review covers the handoff excluded by configured lint |

## 9. Top-Level Work Tracker

| Work item | Workstream | Status | Dependencies / evidence | Exit condition |
| --- | --- | --- | --- | --- |
| T-018 — initial object-store plan save / documentation | Historical save | DONE | E-048–E-052; V-030 | Sole-file update, reviewed plan/history, truthful checks and preservation proof |
| T-019 — current characterization | WF-20 | DONE | S-020; V-025; E-057/E-058 | Current baseline and complete assertion/caller dispositions saved |
| T-020 — ownership and local dependencies | WF-21 | DONE | T-019; RB-008; S-021/V-026 | Ownership/settlement guarantees qualified; wrappers/globals removed; S-025 remains separate |
| T-025 — bound attachment correction | WF-25 | DONE | T-020; RB-010; S-025/V-031 | Binding rejection, terminal failure, waiter cancellation and saved recovery inputs qualified |
| T-021 — test consolidation | WF-22 | DONE | T-025; RB-009; S-022/V-027 | Meaningful successors qualified; truthful fake behavior and typed absence replace misleading assertions |
| T-022 — routing cutover | WF-23 | DONE | T-021; RB-009; S-023/V-028 | Authored/generated/nested selectors and execution classes agree |
| T-023 — final qualification and handoff | WF-24 | DONE | T-022; RB-009; S-024/V-029; E-063 | Frozen compatible portfolio, cleanup and bounded acceptance recorded |
| T-024 — refine object-store plan / documentation | Historical save | DONE | E-053–E-056; V-032 | Refined plan saved, documentation checks passed and sole-file preservation verified |
| T-026 — post-completion review and correction | Scoped follow-up | DONE | F-034–F-037; V-033; E-064 | Correct identified gaps, qualify exact focused portfolio and preserve historical checkpoints |

T-001–T-017, WF-00–WF-19, F-001–F-026, C-001–C-016,
S-000–S-019, V-001–V-024, E-001–E-047 and RB-001–RB-007 are
reserved historical IDs. Document completion cannot complete a future workflow.

## 10. Session Handoff Log

### Post-completion review and correction complete — E-064

The user authorized review and correction after E-063. Entry remained dirty
`main` at `9e6fa4053e2de4367ce106606da02e74fafc72d5`, with the same 20 entries
from the completed iteration and no unrelated edits. The review-entry snapshot
and full handoff are under `/tmp/cartulary-s3-review-lt5j7ior/`. All prior log
text, including E-063, remains historical and is preserved verbatim.

Read root/backend/tooling instructions and the test-operations skill and its
references. `make help ALL=1` and `make task-guide OWNER=harness.browser` passed.
Reviewed all S3 implementation/test files, the 30-test assertion ledger, affected
composition/recovery/process callers, authored/nested routing, shared transport
and event-binding boundaries, and REQ-804/402/414 plus §11.4. The separate owner,
borrower, narrow fakes, three unit groups and two real scenarios remain coherent;
no additional framework, compatibility route or service abstraction is needed.

| Finding | Correction and successor assurance |
| --- | --- |
| F-034: polling cancellation lost from error chain | Retain the polling termination alongside the original probe cause. Deterministic cancellation between probes now proves the interrupted setup envelope, both causes and probe cleanup. |
| F-035: secondary cleanup masks or disappears behind readiness | End admission on uncertain cleanup. Preserve primary probe stage, normalized cause and outcome; count probe timeouts separately. Controller cases cover transient cleanup, cleanup-only timeout, primary deadline and primary cancellation, and prohibit another probe from hiding the failure. |
| F-036: incomplete negative-case coverage | The waiter observes entry into its cancellation select before cancellation. Binding cases now include runtime/lease identities and lifecycle modes. The existing real namespace scenario proves deletion can succeed while required publication fails, and the failure still reaches its caller; sibling/final cleanup assertions remain. |
| F-037: omitted fixture CA reaches process-level trust | Require the captured fixture CA before attachment or recovery readiness/network access. The managed owner already publishes it. The missing-CA case proves rejection before verification; the shared production transport retains its separate system-trust mode. |

F-034/F-035 were pre-existing controller defects missed by the earlier
consolidation's guarantee validation. F-037 completes S-025's captured-input
boundary: `cryptography.systemTLSRoots` reads process-level trust on its first
use when no explicit root is supplied. No production transport policy, timeout,
poll interval, permitted startup retry, provider, schema, row identity or caller
API signature changed. No new top-level test was necessary; the existing exact
selectors cover the additional subcases. The only source edits in this review
are `s3test.go`, `attachment.go`, `readiness_test.go`, `ownership_test.go` and
`namespace_integration_test.go` in `internal/testutil/s3test`.

All commands use E-062's pinned Go executable/PATH, `CGO_ENABLED=0`,
`GIT_OPTIONAL_LOCKS=0`, installed-only preparation and fresh explicit run IDs.
Every run ID below resolves under the exact parent
`/home/jochi/code/cartulary/.cartulary/test-results/tr01-s3-review/`.

Negative reproduction and intermediate results remain separate from acceptance:

| Command / run ID | Session / native exit | Outcome / relationship |
| --- | --- | --- |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_readiness,harness.browser.unit.object_store_ownership` / `review-negative-20261026-01` | 19710 / Make 2 | 3,263 ms; readiness row/unit exit 10, `product/test_assertion_failure`; ownership row passed; all 10 cleanup outcomes completed. Added controller negatives expose F-034/F-035. |
| `make format` / `review-format-20261026-01` | 1279 / 0 | 5,459 ms; first correction formatted. |
| Six-row selection below / `review-focused-20261026-01` | 38912 / 0 | 12,790 ms; precedes trust correction, so checkpoint evidence only. |
| `make agent-finalize` / `review-finalize-20261026-01` | 68069 / 0 | 23,411 ms; superseded after the later trust finding. |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_ownership` / `review-trust-negative-20261026-01` | 64584 / Make 2 | 1,834 ms; row/unit exit 10, `product/test_assertion_failure`; all 10 cleanup outcomes completed. Missing-CA witness fails before F-037 correction. |
| `make format` / `review-format-20261026-02` | 66283 / 0 | 5,378 ms; final correction formatted. |

The negative roots' structured summaries and row/unit outcomes were inspected;
`make explain-run` confirmed the first failure. Its retained stdout was empty,
so individual subcase-log completeness is not claimed. Failures are preserved,
not overwritten by later passes.

The final focused command is:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_ownership,harness.browser.unit.object_store_readiness,harness.browser.unit.object_store_namespaces,harness.browser.integration.object_store_namespace_lifecycle,harness.browser.integration.object_store_stable_reset,harness.browser.unit.testservices_lifecycle_contract
```

The finalizer then passed schema shape, tier coverage and generated refresh,
with zero file changes. Both retained canonical-evidence and scheduler-drift
maintenance were skipped as `results-dir-not-provided`: RESULTS_DIR was explicitly
unset because no compatible successful full warm-check exists. Its summary is
`review-finalize-20261026-02/unit-artifacts/finalize-summary.json`. Source froze at
`sha256:d7b53f17a7e34c2ec60fbdb5f8e64e484deefa482346c352c4cb6ddafa0b8be7`;
`/tmp/cartulary-s3-review-lt5j7ior/frozen-source.json` retains non-Markdown content
hashes. The focused result remained compatible, so it was not repeated.

`make target-plan TARGET=test-slice
INPUTS=/tmp/cartulary-s3-review-lt5j7ior/final-selection.json DETAIL=rows JSON=1`
passed (session 23741, exit 0), confirming the native borrower row's existing
self-owned integration profile without adding broker dependencies.

| Accepted command / run ID | Session / native exit | Result / completed cleanup outcomes |
| --- | --- | --- |
| Focused six-row selection / `review-focused-20261026-02` | 60410 / 0 | Pass, 10,828 ms / 16 |
| `make agent-finalize` / `review-finalize-20261026-02` | 29784 / 0 | Pass, 22,885 ms / 10 |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_borrowed_fixture_lifecycle` / `review-native-20261026-01` | 88929 / 0 | Pass, 13,801 ms / 12 |
| `make lint CHECK=go` / `review-lint-20261026-01` | 21232 / 0 | Pass, 9,889 ms / 15 |
| `make source-check CHECK=generated` / `review-generated-20261026-01` | 54426 / 0 | Pass, 18,022 ms / 14 |
| `make source-check CHECK=catalog` / `review-catalog-20261026-01` | 69799 / 0 | Pass, 3,123 ms / 11 |

All six accepted roots share the frozen source digest and passed
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, each with
native exit 0, full_validity=1 and read_only=1. Inspected `run-manifest.json`,
`run-summary.json`, row results and `cleanup-results.json` show zero failed,
skipped or cancelled graph units, seven passing selected test rows plus two
generated-policy rows, and all 78 cleanup outcomes completed. Execution-result
cache hits are zero. Optional execution/resource observations remain unavailable
or not observed; no performance or resource claim is made.

This portfolio qualifies the affected readiness, attachment/trust, package and
namespace boundaries, publication and borrower lifetime. Schema/tier checks also
passed inside finalization. No routing/schema input changed, so separate identity,
full routing, application-process, production-adapter, full-owner, CI or release
runs were not repeated. E-063's wider portfolio remains historical on its own
source. This focused review does not promote it to the corrected source.

`make lint-markdown` passed as `review-markdown-20261026-01` (session 62410,
native exit 0, 17,240 ms); its adhoc summary is
`adhoc/lint-markdown/tool-run-summary.json` beneath that exact root. Configured
lint globs exclude this handoff, so direct review also verified all twelve
sections, tables, local links, IDs and dependencies, the 30 baseline assertion
dispositions, and the review's failure/successor mapping. This documentation run
is separate from the six graph qualification roots. `git --no-optional-locks
diff --check` passed.

Final comparison with the review-entry snapshot finds exactly six modified files:
the five S3 files listed above and this handoff, with no added/removed paths.
No source content changed after finalization/freeze. The original 20-entry
implementation delta and 7,654 tracked/nonignored files remain; HEAD and staged
entries are unchanged, with no staged diff. The handoff remains the same regular,
singly linked inode, and all prior section-10 log text from E-063 onward remains
byte-for-byte preserved. Raw Git-index byte identity and exhaustive ignored-tree
preservation are not claimed. Final direct checks were repeated after saving
completion status.

F-034–F-037, V-033 and T-026 are complete. No unresolved finding remains from this
bounded review. Acceptance is the focused portfolio above, with retained-run
maintenance skips and broader historical evidence kept explicitly separate.

### WF-24 exit / object-store iteration complete — E-063

`make agent-finalize` completed as `wf24-finalize-20261026-01` (session 56369,
native exit 0, 23,045 ms). Its `unit-artifacts/finalize-summary.json` reports
schema shape, tier coverage and generated refresh passed. Retained canonical
evidence and scheduler drift maintenance were skipped with
`results-dir-not-provided`: RESULTS_DIR was explicitly unset because no compatible
full warm-check root exists. These skips do not establish release/whole-owner
acceptance. Comparison against `pre-finalize.json` found no changed source files.

Source is frozen at dirty `main`, commit
`9e6fa4053e2de4367ce106606da02e74fafc72d5`, digest
`sha256:17120726ceef8adf9447044354b3fecf807b8a99432e2800dc921794e1759603`.
The non-Markdown content snapshot is
`/tmp/cartulary-s3-implementation-p5bpj7gf/frozen-source.json`.
All final qualification uses the common E-057 environment plus E-062's pinned
PATH. No timeout, retry, capacity or fixture policy was widened.

Public `make target-plan TARGET=test-slice INPUTS=<file> DETAIL=rows JSON=1`
passed for `/tmp/cartulary-s3-implementation-p5bpj7gf/final-{a,b,c,d}.json`
(sessions 55361/12573/33351/52064, all native exit 0). Each corresponding
`.plan.json` is preserved externally. A selects the five E-062 S3 rows plus
`harness.browser.unit.testservices_lifecycle_contract`; B selects physical and
borrowed native lifecycle; C and D retain the exact Section 8 selections. The
plans confirm three service-free S3 groups and two object-store integration
scenarios. B's native selected-service row is supplied by the already compatible
`wf23-selected-20261026-01` root, preserving its distinct nested selection witness
without repeating it. The independent routing-contract row, Go lint and backend
boundary also retain compatible E-062 roots. Earlier differing source digests
supply checkpoint evidence only. This portfolio is bounded, not full-owner CI.

| Final invocation / run ID | Session / native exit | State |
| --- | --- | --- |
| A / `wf24-a-20261026-01` | 54578 / 0 | Pass, 9,448 ms; 16 completed cleanup outcomes |
| B physical/borrowed / `wf24-b-20261026-01` | 10774 / 0 | Pass, 138,875 ms; 21 completed cleanup outcomes |
| `make source-check CHECK=generated` / `wf24-generated-20261026-01` | 42874 / 0 | Pass, 14,844 ms; 14 completed cleanup outcomes |
| `make source-check CHECK=schema` / `wf24-schema-20261026-01` | 88570 / 0 | Pass, 6,930 ms; 12 completed cleanup outcomes |
| `make source-check CHECK=catalog` / `wf24-catalog-20261026-01` | 27551 / 0 | Pass, 2,084 ms; 11 completed cleanup outcomes |
| `make source-check CHECK=identity` / `wf24-identity-20261026-01` | 11120 / 0 | Pass, 1,124 ms; 12 completed cleanup outcomes |
| C / `wf24-c-20261026-01` | 38717 / 0 | Pass, 73,974 ms; 21 completed cleanup outcomes |
| D / `wf24-d-20261026-01` | 92614 / 0 | Pass, 8,178 ms; 14 completed cleanup outcomes |

Every run ID in this checkpoint and the four retained E-062 roots resolves under
`/home/jochi/code/cartulary/.cartulary/test-results/tr01-s3-implementation/`.
Their exact `run-manifest.json`, `run-summary.json`, `rows/*.json` and
`cleanup-results.json` are the retained evidence. The thirteen accepted graph
roots (four compatible E-062 roots, finalizer, A–D and four static checks) passed
`make harness-observability-check RESULTS_DIR=<exact-root> CHECK=all`, each with
native exit 0, full_validity=1 and read_only=1. All thirteen source digests match
the frozen source. Earlier static roots preceded final layout normalization and
were requalified, not silently treated as compatible.

All accepted summaries report pass with zero failed, skipped or cancelled units.
The six test-slice roots cover 13 distinct selected rows: A's six, B's two,
C's two, D's one, native selected-service and routing-contract. The generated
check contributes two additional passing policy rows and backend-boundary lint
one, for 16 distinct retained row results overall. All 179 recorded cleanup
outcomes completed. Finalizer's two internal retained-maintenance skips remain
explicit above; they are not hidden by the graph-unit counts. Optional execution
and resource observations are unavailable or not observed, so no performance or
resource claim is made.

Final A and B commands replace only their retired baseline selections:

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_ownership,harness.browser.unit.object_store_readiness,harness.browser.unit.object_store_namespaces,harness.browser.integration.object_store_namespace_lifecycle,harness.browser.integration.object_store_stable_reset,harness.browser.unit.testservices_lifecycle_contract
make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.native_borrowed_fixture_lifecycle
```

C and D are exactly Section 8's commands. The compatible nested-selection root
used `make test-slice OWNER=harness.browser
ROWS=harness.browser.integration.native_selected_service_lifecycle`; the additional
routing witness used `make test-slice OWNER=harness.test_catalog
ROWS=harness.test_catalog.behavior.routing_contract`. These are independent
qualification boundaries, not duplicate S3 selectors.

Comparison with the implementation-entry snapshot finds exactly 20 changed,
added or removed entries: this handoff; eleven entries in `internal/testutil/s3test`
(including seven additions and two removals); the process teardown and native
selection consumers; testservices composition and its tests; authored browser
family, historical row migration and tier-selection assertions; and the
Make-generated topology render index. The 7,654 remaining tracked/nonignored
files include no unrelated changes. Non-Markdown content exactly matches
`frozen-source.json`; HEAD and staged entries are unchanged, with no staged diff.
Raw index metadata changed, so raw byte identity is not claimed. The handoff
remains a regular singly linked file at inode 348826, and all prior log text from
E-053–E-056 through the historical section-10 tail is byte-for-byte preserved.
Ignored runtime/cache/build trees were not exhaustively fingerprinted.

`make lint-markdown` passed as `wf24-markdown-20261026-01` (session 67753,
native exit 0, 17,470 ms). This is a separate adhoc documentation invocation,
not an additional graph qualification root; its authoritative summary is
`adhoc/lint-markdown/tool-run-summary.json` beneath that exact run root.
Configured lint globs exclude this handoff. Direct review therefore checked the
twelve sections, consistent tables, resolvable local links, retained IDs and
ordered gates, the 30-test assertion ledger and its successor coverage, and
agreement among completion/readiness/validation records. `git --no-optional-locks
diff --check` passed. Source preservation and history checks were repeated after
the final document save; only this handoff changed after source freeze.

WF-20/WF-21/WF-25/WF-22/WF-23/WF-24, S-020–S-025 and
T-019/T-020/T-025/T-021/T-022/T-023 are complete. F-027–F-033 are resolved within
this scope. All required production guarantees have meaningful successor
witnesses; acceptance covers the bounded object-store refactor, with no full-suite,
release, performance or preceding WF-33 closure claim. No implementation work
remains in this iteration.

### WF-23 exit / WF-24 entry — E-062

Authored routing now has exactly these S3 groups, all retaining
`harness.browser.verification.support_contract` because its meaning is unchanged:

| Current row | Selection / execution declaration |
| --- | --- |
| `harness.browser.unit.object_store_ownership` | 10 ownership/attachment/configuration tests; service-free, capability `none` |
| `harness.browser.unit.object_store_readiness` | 7 mutation-readiness/diagnostic tests; service-free, capability `none` |
| `harness.browser.unit.object_store_namespaces` | 9 namespace/package/cleanup tests; service-free, capability `none` |
| `harness.browser.integration.object_store_namespace_lifecycle` | `TestObjectStoreNamespaceLifecycle`; integration, capability `object_store_namespace`, dependency `object_store` |
| `harness.browser.integration.object_store_stable_reset` | `TestResetBucketPreservesNamespaceAndProvesMutation`; integration, same capability/dependency |

The native selected-service tests and scheduler tier assertions now name these
current rows. The existing historical migration record points directly to current
successors; no executable alias or new compatibility mechanism was introduced.
Source/catalog inspection finds no old row IDs in executable selectors. The
standalone publication, native, application and production-adapter rows remain.
A review also kept owner-only CORS inputs out of borrowed binding equality: owner
startup reads them once, but changing an irrelevant ambient CORS variable cannot
invalidate an otherwise identical borrower. Imports/field layout were normalized
for readability; no new policy or source owner was introduced.

Common environment and result parent remain E-057. Current validation:

| Command / run ID | Native session / exit | Result |
| --- | --- | --- |
| `make generate` / `wf23-generate-20261026-01` | 29809 / 0 | Pass, 9,980 ms; generated topology render index refreshed |
| `make source-check CHECK=generated` / `wf23-generated-20261026-01` | 83477 / 0 | Pass, 13,627 ms; precedes final layout cleanup and remains checkpoint evidence |
| `make source-check CHECK=schema` / `wf23-schema-20261026-01` | 19113 / 0 | Pass, 6,122 ms |
| `make source-check CHECK=catalog` / `wf23-catalog-20261026-01` | 99254 / 0 | Pass, 1,820 ms |
| `make source-check CHECK=identity` / `wf23-identity-20261026-01` | 53867 / 0 | Pass, 1,036 ms |

`make format` passed as `wf23-format-20261026-01` (session 35691, exit 0,
5,422 ms). `make lint CHECK=go` initially failed as
`wf23-lint-go-20261026-01` (session 85353, Make exit 2; Staticcheck unit exit 1,
`product/test_assertion_failure` normalization). Formatting/vet passed; bounded
Staticcheck stderr identified its internally launched `go` as 1.27.1 with
`GOTOOLCHAIN=local`. It uses PATH independently of Make's GO variable.

Subsequent commands additionally prepend
`/home/jochi/.cache/cartulary/go/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin`
to PATH. No pin, tool or verification rule changed. Corrected Go lint passed as
`wf23-lint-go-20261026-02` (session 66254, exit 0, 7,818 ms).
`make lint CHECK=backend-boundary` passed as `wf23-boundary-20261026-01`
(session 89717, exit 0, 2,890 ms). The exact routing contract passed as `wf23-routing-20261026-01`
(session 84190, exit 0, 98,154 ms; all 11 cleanup outcomes completed).
The native selected-service witness passed as `wf23-selected-20261026-01`
(session 36224, exit 0, 39,425 ms). Go lint's 15 and boundary lint's 12 cleanup
outcomes completed; all these graph summaries have no failed/skipped/cancelled
units. V-028/WF-23/S-023/T-022 are DONE. Earlier lint failure remains historical.

The post-format lint and routing roots share source digest
`sha256:17120726ceef8adf9447044354b3fecf807b8a99432e2800dc921794e1759603`.
WF-24 enters: run `make agent-finalize` with RESULTS_DIR unset (no compatible
successful full warm-check exists), inspect its writes and freeze source. Reuse
focused roots only if compatibility survives; run missing A–D coverage and static
checks, then audit every accepted graph root. Earlier focused results with a
different digest are checkpoint evidence only. This exit was saved before finalization.

### WF-22 exit / WF-23 entry — E-061

Unit assertions now reside in `ownership_test.go`, `readiness_test.go` and
`namespace_test.go`; real scenarios reside in `namespace_integration_test.go`.
Package quarantine moved with namespace lifecycle. Readiness and cleanup retain
separate narrow fakes. No mutable package-global substitutions remain.

The readiness fake leaves state unchanged on ordinary failure, explicitly marks
namespace-delete response loss after mutation, and rejects nonempty namespace
deletion. The old false no-residue assertion is replaced by actual retained-state
and primary/secondary failure assertions. A private per-fixture poll wait lets
unit cases coordinate recovery/expiry while production keeps its existing ticker
and deadlines. Testservices' independent publication witness uses an already
expired deadline instead of sleeping for a nanosecond timer.

The startup-description assertion retires to mapped-endpoint-before-readiness
coverage plus `TestObjectStoreStartupMappingDeadline`. Cleanup request totals
retire to bounded batches, remaining-state inspection, fresh empty observation
and rejection before requesting beyond a repeated token. Required one-time
admission/settlement/quarantine effects and causal mutation order remain.

Three real tests (`TestHarnessStartsSeaweedFSS3AndRoundTripsObjects`,
`TestCleanupPrefixPreservesSiblingObjects`,
`TestPinnedObjectStoreBatchCleanupTraversesMultiplePages`) now have one successor,
`TestObjectStoreNamespaceLifecycle`: exact roundtrip, 1,007-key pagination,
bounded batches, independent empty listing, typed deleted-key absence, unchanged
sibling bytes, saved-binding recovery despite ambient conflict, verified bucket
removal and repeated final cleanup. The independent stable-reset case now requires
typed absence and exact new payload. Transport/authentication errors cannot prove
absence; negative readiness cases explicitly reject them.

The existing integration row selects the new scenario and the existing unit row
selects the renamed deadline test. Authored selectors changed with source;
retired names have no aliases. Final semantic row consolidation remains WF-23.
`make format` passed as `wf22-format-20261026-01` (session 7223, exit 0,
5,541 ms). `make generate` passed as `wf22-generate-20261026-01`
(session 75834, exit 0, 10,420 ms), updating the generated topology render index.
A with the migrated source selectors passed as `wf22-a-20261026-01` (session
60455, exit 0, 11,945 ms): six selected rows, zero failed/skipped/cancelled units,
all 17 cleanup outcomes completed. V-027 and WF-22/S-022/T-021 are DONE.

WF-23 entry inspection additionally found exact references in
`tools/harness/scheduler/work-graph/tests/tier-selection.mjs` and the existing
object-store entry in `tools/test_catalog_row_migrations.json`. These are scoped
routing consumers, not scheduler architecture changes. The migration loader
validates retired/current identities; it does not execute retired aliases. Update
that existing record's successors, all tier/nested selectors, and the authored
family atomically. Add the existing exact
`harness.test_catalog.behavior.routing_contract` row to qualification for the
changed tier assertion boundary; no new harness or policy is needed.
This checkpoint was saved before routing consolidation.

### WF-25 exit / WF-22 entry — E-060

Before correction, new negative subcases were added under the existing
`TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness` selector. The exact
service-free row was run as `wf25-negative-20261026-01` (native session 69868,
Make exit 2, row/unit exit 10). Its run summary reports
`product/test_assertion_failure`; a bounded stdout inspection confirmed changed
bindings reused cached admission. This is an expected defect witness, not a
passing baseline. The terminal row/summary remain authoritative; the subsequently
read stdout file was empty, so no complete retained-log claim is made.

The correction captures a copied allowlist of attachment, dependency, TLS,
runtime and event-destination inputs. One private admission instance retains its
binding and terminal result. Reuse compares the full binding; pending callers
wait on completion with their own context, outside the admission lock. No new
registry or replacement generation exists. Fixture guards, names and journals
use saved inputs; event publication errors propagate instead of reporting success.
Physical startup captures its runtime/CORS inputs once as well.

Recovery now calls `CleanupBucketFromBinding` with the supplied saved environment
only, after testservices' generated-name ownership proof. It cannot fall back to
StartShared or perform fresh admission/provisioning. The saved-binding borrower
closes its local transport. Missing/incomplete bindings fail before network use.
New cases cover each required attachment field, malformed/omitted dependencies,
independent copied bindings, changed credentials/dependencies/destinations,
terminal readiness/publication failure and local waiter cancellation. Existing
selectors include these subcases; no compatibility routing is added.

All roots use E-057's parent and common pinned environment.

| Invocation / run ID | Session / native exit | Result |
| --- | --- | --- |
| Negative service-free row / `wf25-negative-20261026-01` | 69868 / 2 | Expected assertion failure before correction, summary 2,843 ms |
| `make format` / `wf25-format-20261026-01` | 79163 / 0 | Pass, 5,583 ms |
| A / `wf25-a-20261026-01` | 66040 / 0 | Pass, 12,460 ms; six rows including all new negative witnesses and distinct testservices publication |
| Native borrowed lifecycle / `wf25-b-20261026-01` | 2190 / 0 | Pass, 10,207 ms; all 12 cleanup outcomes completed |

C (both original rows) passed as `wf25-c-20261026-01`, session 53694, native
exit 0, summary 76,943 ms; all 21 cleanup outcomes completed. A's 17 cleanup
outcomes completed. A/B/C summaries contain zero failed, cancelled or skipped
units. V-031 is qualified; WF-25/S-025/T-025 are DONE. The expected pre-correction
failure is retained. S-022 may now consolidate the qualified negative cases with
the other assertion groups. This checkpoint was saved before WF-22 edits.

### WF-21 exit / WF-25 entry — E-059

The structural extraction introduces `owned_store.go` (`OwnedStore`, one
`StartOwned(ctx, StartOptions)` constructor, owner-typed settlement callbacks,
physical/TLS close and broker-probe construction), `attachment.go` (one process
admission instance; independent instances in tests), and `operations.go` (private
fixture dependencies). `Harness` cannot terminate containers or access proxy keys;
its close releases local transport and retains package quarantine. No embedding
or compatibility aliases remain. Testservices uses a local start/probe seam;
process TestMain closes its borrowed transport. The existing 30 selector names
remain valid; no routing migration is needed for this extraction.

New subcases under the existing ownership/configuration selector qualify
before-acquisition failure, acquisition cancellation, publication failure,
before→acquire→publish→start→ready ordering, TLS confinement/removal, memoized
owner close and borrower nontermination/quarantine retention. Existing partial
cleanup uncertainty and readiness terminal/retry assertions remain. Admission
still has the old caching/ambient semantics deliberately; WF-25 owns that correction.

All roots use E-057's result parent and common pinned environment.

| Invocation / run ID | Session / native exit | Result |
| --- | --- | --- |
| `make format` / `wf21-format-20261026-01` | 54979 / 2 | Prerequisite Go build failed before formatting: method expression had receiver/context in the wrong order. Step summary and bounded stderr at `harness-runtime-build/build-harness-runtime/` retain the failure; corrected with a local function |
| `make format` / `wf21-format-20261026-02` | 67981 / 0 | Pass, 5,433 ms; only scoped authored Go formatting |
| A / `wf21-a-20261026-01` | 59326 / 0 | Pass, 11,486 ms; all six rows including new ownership subcases and publication witness |
| B physical/borrowed rows / `wf21-b-20261026-01` | 91260 / 0 | Pass, 136,258 ms; both rows, all 21 cleanup outcomes completed. Nested selected-service qualification remains a later routing/final obligation |
| C process row only / `wf21-c-20261026-01` | 24453 / 0 | Pass, 76,561 ms; all 20 cleanup outcomes completed |
| D / `wf21-d-20261026-01` | 77436 / 0 | Pass, 8,078 ms; all 14 cleanup outcomes completed |

A's 17 cleanup outcomes completed. All four test summaries report zero failed,
skipped or cancelled units. The scoped selectors remain unchanged; no generated
routing edits were needed. `git diff --check` passed. WF-21/S-021/T-020 are DONE;
RB-010 is resolved. WF-25 enters with separate negative witnesses before any
admission correction. This exit was saved before those witnesses were added.

### WF-20 exit / WF-21 entry — E-058

All four V-025 baseline selections passed on the unchanged implementation. Commands
use the exact Section 8 rows and E-057 common environment, including the explicit
cached pinned GO executable. Each exact root is
`/home/jochi/code/cartulary/.cartulary/test-results/tr01-s3-implementation/<run-id>`.

| Run ID / group | Native session / exit | Selected rows / summary wall | Cleanup |
| --- | --- | --- | --- |
| `wf20-a-baseline-20261026-02` / A | 10144 / 0 | 6 / 10,354 ms | 17 completed |
| `wf20-b-baseline-20261026-01` / B | 93217 / 0 | 3 / 145,219 ms | 22 completed |
| `wf20-c-baseline-20261026-01` / C | 23974 / 0 | 2 / 73,003 ms | 21 completed |
| `wf20-d-baseline-20261026-01` / D | 51528 / 0 | 1 / 7,399 ms | 14 completed |

All summaries report pass with zero failed, skipped or cancelled units; selected
rows are distinct from setup/work-unit counts. The first A attempt remains a
pre-test toolchain invocation failure (E-057), not a deleted or passing result.
The ledger covers every baseline test/assertion family and affected caller;
known defects and missing negative witnesses remain explicit. No historical
PostgreSQL evidence was promoted. WF-20/S-020/T-019 are DONE and RB-008 is resolved.

WF-21 now enters. Its exit must qualify physical-versus-borrowed ownership,
partial settlement and local dependencies before WF-25 changes admission.
There is no S-025 behavior claim at this checkpoint. This checkpoint was saved
before source movement.

### Object-store implementation entry — E-057

2026-10-26, America/New_York. Current user authorization supersedes the earlier
planning-only restriction. WF-20 is in progress; no successor slice has entered.
The implementation entry baseline is `main` at
`9e6fa4053e2de4367ce106606da02e74fafc72d5`, with pre-existing changes confined to
this handoff. All 7,649 tracked/nonignored entries, HEAD/index and the original
handoff were saved in `/tmp/cartulary-s3-implementation-p5bpj7gf/baseline.json`
and `original-handoff.md`. The E-053/E-056 baseline table above describes the
prior document save, not this invocation's original handoff bytes. No staging or
commit is part of this work. Historical checkpoints below remain unchanged.

Public `make help ALL=1` and task guides for `harness.browser`, `app.server` and
`platform.objectstore` succeeded. The 30-test ledger is supplemented above at
assertion/subcase and consequential-caller level. F-028/F-029/F-030/F-032/F-033
remain known defects; baseline passes do not adopt their incidental behavior.

V-025 uses the exact A–D selections in Section 8. Common command environment:
`CGO_ENABLED=0 GIT_OPTIONAL_LOCKS=0 CARTULARY_PREPARATION_POLICY=installed_only
CARTULARY_TEST_RESULTS_DIR=/home/jochi/code/cartulary/.cartulary/test-results/tr01-s3-implementation`.
Each invocation supplies the run ID below. After the first failure, it also
supplies `GO=/home/jochi/.cache/cartulary/go/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go`.

| Run ID / group | Native session / exit | Result and cleanup |
| --- | --- | --- |
| `wf20-a-baseline-20261026-01` / A | 9411 / 2 | Before tests: cryptographic module toolchain identity rejected. No run root published. Installed launcher VERSION is go1.27.1; required cached executable VERSION is go1.27.3. No pin or verification bypass; corrected explicit GO for the new invocation |
| `wf20-a-baseline-20261026-02` / A | 10144 / 0 | Pass, 6 selected rows; summary wall 10,354 ms; all 17 published cleanup outcomes completed, no failed/skipped units. Root is the result parent above plus this run ID |
| `wf20-b-baseline-20261026-01` / B | 93217 / pending | Native lifecycle validation running; not yet credited |

WF-20 exit is not yet saved. Source refactoring remains gated on the complete
V-025 baseline; no source changes or generation have occurred at this entry.

### Object-store plan refinement — E-053–E-056

2026-10-26, America/New_York. The user selected refinement of the unimplemented
object-store iteration, chose to retain the attachment defect as a separate
correction, and authorized saving this document-only plan. Only this handoff may
change. Source refactoring, tests, generation, services and normative amendments
remain future work. Entries below this refinement record retain their original
historical status, authorization and evidence limits; they are not current results.

#### Refinement decisions and source posture

- E-053 establishes the actual starting worktree: the same `main` commit, clean
  staged entries and pre-existing edits only in this handoff. The previous
  planning pass left its fingerprint unchanged. Do not replace those edits with
  the committed version or reuse the earlier clean-worktree claim.
- E-054 revalidates the five authored S3 files (3,076 lines; 30 top-level tests),
  affected composition/consumer boundaries and routing against the unchanged
  source commit. The focused inventory remains bounded, with partial caller
  inspection and metadata-only remaining entries; no exhaustive suite audit.
- F-033 records weak absence assertions in the prefix-cleanup and stable-reset
  tests. Successors require typed not-found results, surviving sibling bytes and
  successful post-reset mutation; unavailable transport cannot prove deletion.
- S-021/WF-21 now contains structural movement only. S-025/WF-25/V-031 corrects
  bound admission after that checkpoint and before S-022/WF-22 consolidation.
  The owner clauses are REQ-804/410; no specification amendment is proposed.
- F-029 now explicitly rejects implicit recursive fake bucket deletion. Default
  failure preserves state; response loss after mutation is an explicit case.
- The 30-test ledger and guarantee/retirement table distinguish required assurance
  from historical strings, request counts, globals, names and redundant setup.
  Existing verification identities remain when their meaning is unchanged.
- No frontend presentation, production API/data schema, PostgreSQL internals,
  scheduler, provider or performance-policy change is part of this iteration.

#### Refinement evidence register

| Evidence | Source / baseline | Result and applicability |
| --- | --- | --- |
| E-053 | `/tmp/cartulary-s3-plan-refinement-rv0caa_o/baseline.json` and `original-handoff.md` | Fingerprints and file identities for 7,649 tracked/nonignored entries; HEAD, raw index and staged-entry hashes captured before editing; prior document edits retained |
| E-054 | Preceding refinement inspection; current unchanged source commit, applicable root/internal/tools guidance and owner clauses | F-027–F-033, C-017–C-021, separate correction and successor coverage plan; read-only Make help and task-guide discovery succeeded for harness.browser, platform.objectstore and app.server; no tests ran |
| E-055 | V-032 documentation validation; exact command and artifacts below | Outer lint exit 0, configured lint passed; direct document review and whitespace check passed. Configured lint excludes this handoff; no product assurance claim |
| E-056 | Pre-write identity/content check; `/tmp/cartulary-s3-plan-refinement-rv0caa_o/comparison.json` | Only this handoff changed across 7,649 entries; no added/deleted paths or identity changes; HEAD/branch/raw index/staged entries unchanged. Historical log preserved byte-for-byte; ignored-tree limits disclosed |

#### Refinement documentation validation

Executed from `/home/jochi/code/cartulary`, with installed prerequisites,
fixing disabled, external result/cache destinations and network-disabled Corepack:

```sh
GIT_OPTIONAL_LOCKS=0 \
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-s3-plan-refinement-rv0caa_o/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-s3-plan-refinement-rv0caa_o/static-analysis \
CARTULARY_TEST_RUN_ID=tr01-object-store-plan-refinement-20261026-01 \
COREPACK_ENABLE_NETWORK=0 HARNESS_DIAGNOSTICS=off HARNESS_ACCOUNTING=off \
make lint-markdown
```

V-032 passed. The native outer Make exit was **0**; the summary reports **pass**,
zero failures, zero tests and 17,200 ms. Exact summary:
`/tmp/cartulary-s3-plan-refinement-rv0caa_o/results/tr01-object-store-plan-refinement-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
Configured lint excludes this handoff, so that pass applies only to its configured
documents. Supplemental review in `document-review.json` beside the baseline
passed: twelve top-level sections, 17 current tables, eight current local links,
stable IDs and finding traceability, the six-workflow dependency chain, all 30
baseline test mappings and four commands selecting 12 existing catalog rows.
The prior 168,640-byte session log is preserved byte-for-byte. These are document
checks, not executed test successors or product verification.

`git --no-optional-locks diff --check -- docs/handoffs/testing-remediation-01.md`
passed. The E-056 comparison found only this handoff changed across all 7,649
tracked/nonignored entries; no added/deleted paths or file-identity changes, and
unchanged HEAD, branch, raw index and staged entries. Final evidence annotations
were directly reviewed and preservation was rechecked; the linted population did
not change. Ignored build/dependency/output trees were not exhaustively audited;
reviewed command paths used installed prerequisites and external result/cache
outputs. The initial uncommitted plan and historical rationale remain preserved
within the refined document and its retained starting copy.

No format, generation, agent-finalize, product tests, service commands or
installation ran. Those operations are excluded by this document-only scope.
T-024 is DONE; V-025–V-029 and V-031 remain unexecuted future obligations.
The next source-work entry remains later-authorized WF-20 characterization.


### Object-store planning save — E-048–E-052

2026-10-26, America/New_York. The user approved implementing the document-update
plan, explicitly keeping source changes and tests as future work. Revalidation
found the same clean source as the planning turn. The sole admitted write is this
handoff; no instructions embedded in its historical entries expand that boundary.

#### Scope and authority

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-048/E-050 | Existing canonical repository target/output admitted; current owner allocation and instructions revalidated; no normative amendment | Save plan only; later actor revalidates source and obtains implementation authorization |

#### Backend boundary

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-049 | Mixed S3 ownership, globals, ambient binding and misleading fake cleanup identified; all five core files inspected | WF-20 qualifies guarantees before owner split; preserve narrow S3 transport/cleanup owners |

#### Frontend boundary

Not applicable to source changes: browser presentation and UI selectors remain
outside scope. Native fixture/selection boundaries are harness witnesses, not a
request for rendered UI inspection.

#### Contract and codegen

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-049/E-050 | Authored rows and nested selectors identified; real reset currently classed unit; no generated edits | WF-23 updates authored inputs and Make projections after successor qualification |

#### Tests and harness

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-049/E-050 | Thirty baseline tests have planned dispositions; A–D exact starting commands discovered; no tests executed | WF-20 completes subcase/caller ledger and current baseline; preserve distinct physical/publication witnesses |
| E-051 | V-030 passed: outer lint exit 0; direct handoff review and diff check passed | No product test evidence is implied; retain configured lint coverage limit |

#### Security and authorization

| Evidence | Current state | Next action |
| --- | --- | --- |
| E-049/E-050 | Preserve TLS roots/private owner keys, exact namespaces, saved recovery proof, bounded diagnostics and production fail-fast transport | Add/retain fault cases without serializing raw provider errors or adding product retries |

#### Open risks and next session

| Evidence | Current state | Next action |
| --- | --- | --- |
| RB-008/RB-009 | Planning decisions settled; current characterization and predecessor exits are future gates | Begin only later-authorized WF-20; do not resume unrelated WF-33 |
| E-052 | All 7,649 entries compared; only this handoff changed; no identity/path-set/index/HEAD change | Later actor revalidates the baseline; ignored trees were not exhaustively fingerprinted |

#### Current evidence register

| Evidence | Source / baseline | Observation and applicability |
| --- | --- | --- |
| E-048 | Clean `main` at `9e6fa4053e2de4367ce106606da02e74fafc72d5`; `/tmp/cartulary-s3-plan-doc-ohx6z1fv/baseline.json` | All 7,649 path identities/content captured; raw index/entries and original tracker retained; no current test result |
| E-049 | Five s3test files and consequential caller/dependency/routing boundaries in Section 2; unchanged planning/save source | Static findings F-027–F-032, C-017–C-021 and 30-test ledger; no exhaustive repository assertion audit |
| E-050 | Root/internal/tools instructions, Refactor Tracker format/framework, adopted owner clauses, supporting domain/NLSpec guidance, Make/task guides and lint invocation path | Authority distinguished from incidental behavior; future commands known; V-024 collision resolved by preserving it and starting at V-025 |
| E-051 | V-030 document checks; exact command and artifact below | Outer Make exit 0; configured lint passed; direct section/table/link/ledger/catalog review and diff check passed; no product test validation implied |
| E-052 | Pre-write identity/content recheck; `/tmp/cartulary-s3-plan-doc-ohx6z1fv/comparison.json` | Only this handoff content changed across 7,649 entries; HEAD, raw index, staged entries and path identities unchanged; verbatim historical checkpoint log retained |

#### Document validation

Executed from the repository root with installed prerequisites, fixing disabled,
external caches/results and network-disabled Corepack:

```sh
GIT_OPTIONAL_LOCKS=0 \
CARTULARY_PREPARATION_POLICY=installed_only \
CARTULARY_TEST_RESULTS_DIR=/tmp/cartulary-s3-plan-doc-ohx6z1fv/results \
CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/cartulary-s3-plan-doc-ohx6z1fv/static-analysis \
CARTULARY_TEST_RUN_ID=tr01-object-store-plan-doc-20261026-01 \
COREPACK_ENABLE_NETWORK=0 HARNESS_DIAGNOSTICS=off HARNESS_ACCOUNTING=off \
make lint-markdown
```

The native outer Make exit was **0**; the lint summary reports **pass**, zero
failures and 17,280 ms. Exact summary:
`/tmp/cartulary-s3-plan-doc-ohx6z1fv/results/tr01-object-store-plan-doc-20261026-01/adhoc/lint-markdown/tool-run-summary.json`.
No test ran. The configured lint globs exclude this handoff, so this pass applies
only to their configured documents. Supplemental document review checked all
twelve top-level sections, tables/fences/whitespace, local links, ID traceability,
all 30 baseline test dispositions, four exact catalog selections and preservation
of the historical checkpoint log. Results are in `document-review.json` beside
the baseline. `git --no-optional-locks diff --check` passed. Final evidence
annotations were directly reviewed; they do not change the linted population.

The comparison in E-052 found no added/removed paths or file-identity changes;
only this handoff's content changed. HEAD, raw index and staged entries remained
identical. Ignored build/dependency/output trees were not exhaustively audited;
the reviewed command paths redirected outputs externally. No formatting,
generation, finalizer, product verification, installation or source refactor ran.
Those operations are excluded by the document-only scope. T-018 is DONE; the
next implementation step remains later-authorized WF-20 characterization.

### Completed PostgreSQL iteration — historical ID index

The following index replaces obsolete current-state summaries; it does not
reopen completed work or discard its evidence. Full original planning detail is
retained in Git at
`9e6fa4053e2de4367ce106606da02e74fafc72d5:docs/handoffs/testing-remediation-01.md`
and in the external original-handoff copy. The detailed checkpoint log beginning
with E-047 below is preserved verbatim. Any “current” or authorization wording in
that historical log refers only to the session it records.

| Historical finding / contract | Completed disposition / slices | Validation / work items / evidence |
| --- | --- | --- |
| F-015; C-011/C-015 | OwnedCluster separates physical PostgreSQL ownership from borrowed Harness; one owned-start API; S-014 | V-017/V-019; T-011; E-039/E-041 |
| F-016; C-012/C-014 | One template recipe, no ordinary migration fallback, confirmed partial-create ownership; S-013/S-017 | V-016/V-021; T-010/T-015; E-038/E-045/E-047 |
| F-017; C-011/C-013 | Local dependencies, bound generation admission and persistent quarantine; S-014/S-018 | V-017/V-022; T-011/T-016; E-039/E-046/E-047 |
| F-018; C-013/C-016 | One real policy resolver and consolidated cleanup cases; S-015/S-018 | V-018/V-022; T-012/T-016; E-040/E-046/E-047 |
| F-019; C-011/C-014 | Capability/meaningful schema assertions replace method counts, smoke-only and impossible borrowed shutdown; S-014/S-015 | V-017/V-018; T-011/T-012; E-039/E-040 |
| F-020; C-011–C-016 | Distinct real transport/reset/ACL/recovery contexts retained; bounded terminology clarified; S-012/S-016 | V-015/V-019; T-009/T-013; E-037/E-041 |
| F-021/F-024; C-012/C-013/C-015 | Context-aware template opens, cancellation-aware wait, one terminal attempt, early confirmed ownership; S-017 | V-021; T-015; E-045 |
| F-022/F-023/F-025/F-026; C-011/C-013/C-014/C-016 | Failed verified generation stays retired; admitted migration capability, immutable inputs and no redundant detach state; S-018 | V-022; T-016; E-046 |
| S-019 / closure | WF-17–WF-19 COMPLETE; frozen compatible final portfolio and handoff | V-023; T-017; E-047; RB-007 DONE |
| Historical planning/review | T-008/T-014 DONE; no implementation authority arose from those document-only saves | V-020/V-024; E-030–E-035/E-042/E-043; RB-005/RB-006 DONE through E-041 |

#### Historical PostgreSQL assertion dispositions


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

#### Closure assertion successors — F-021–F-026

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

### Historical checkpoint log — E-001–E-047

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

| ID | Blocker / affected scope | Needed authority or evidence | Status | Resolution condition |
| --- | --- | --- | --- | --- |
| RB-008 | Current pre-move characterization; S-021 | Current authorization (E-057), V-025 roots and completed subcase/caller dispositions | RESOLVED, E-058 | Current A–D baseline and assertion/caller dispositions qualify movement |
| RB-010 | Separate attachment correction; S-025 | WF-21 exit plus intended negative witnesses for F-028; V-031 qualifies the correction | RESOLVED, E-059/E-060 | Ownership split and V-031 correction qualified before consolidation |
| RB-009 | Ordered successor gates; S-022–S-024 | WF-25/WF-22/WF-23 exits, respectively, with V-031/V-027/V-028 evidence | RESOLVED, E-060–E-063 | Every entry followed its saved predecessor exit; final qualification passed |

No unresolved blocker remains for this bounded iteration or the E-064 review. No applicable
primary-owner contradiction was found. Newly encountered physical-start,
attachment, recovery-cleanup or selector consumers would require their own
bounded discovery and qualification; this completion does not claim an exhaustive
suite audit or justify preserving incidental test behavior.
RB-001–RB-007 remain historical and retain their recorded completion evidence.

## 12. Binary Completion Criteria

| Outcome | Current statement |
| --- | --- |
| Planning completeness | **complete** for the bounded object-store iteration: owners, structural/correction separation, assertion retirement/successor decisions, sequencing, validation and exclusions are specified; not an exhaustive suite audit |
| Per-slice readiness | S-020 **complete**, E-058; S-021 **complete**, E-059; S-025 **complete**, E-060; S-022 **complete**, E-061; S-023 **complete**, E-062; S-024 **complete**, E-063 |
| Post-completion review | **complete**, E-064; F-034–F-037 corrected and qualified, T-026 DONE; no new framework, row, schema or compatibility layer |
| Implementation authorization | **granted by the current implementation request** for the bounded slices and required validation |
| Session write-boundary verification | **Verified**, E-064: this review changed five existing S3 source/test files and the handoff; the original 20-entry implementation delta and 7,654 current entries remain. Final source frozen; all prior logs preserved; HEAD/staged entries/path identity unchanged; raw-index and ignored-tree limits disclosed |
| Implementation / test execution | V-025/V-026/V-031 passed at E-058–E-060; V-027 passed at E-061; V-028 passed at E-062; V-029 passed at E-063; follow-up V-033 passed at E-064 on its corrected source with seven selected test rows and required static/evidence/documentation checks |

The preceding document-only acceptance remains historical; E-063 supplies the
current implementation and final documentation evidence.

| Historical document acceptance group | Result / evidence at that save |
| --- | --- |
| Target/output admission and current baseline | E-053 and pre-write recheck establish the existing repository target, same source commit, dirty document baseline and safe sole-file destination |
| Evidence-backed bounded inspection | Section 2 distinguishes five fully inspected core files, partial caller boundaries, metadata-only entries and excluded ignored trees; E-054 revalidates source posture |
| Required guarantees versus incidental constraints | C-017–C-021 and F-027–F-033 identify owner obligations, meaningful successors and explicit retirement reasons; F-033 excludes arbitrary errors as absence proof |
| Traceability and executable handoff | All 30 baseline tests mapped; WF-20 → WF-21 → WF-25 → WF-22 → WF-23 → WF-24 has explicit gates and rollback; S-025 correction is separate from structural movement |
| Honest verification | Met, E-055: V-032 documentation checks passed with configured lint coverage limits disclosed; E-054 discovery is not passing test evidence; V-030 remains historical |
| Durable history | Prior checkpoints, IDs and assertion dispositions retained; no historical result promoted to current evidence; preceding WF-33 remains open |
| Sole-file preservation | Met, E-056: only this handoff changed; content/path identities/HEAD/index compared with E-053; prior historical log preserved byte-for-byte |

The following implementation criteria are met at E-059–E-063:

- Physical ownership and borrowing have separate types/lifetimes; one owned-start
  API remains; no old wrapper, StopShared alias or borrowed physical shutdown.
- Private dependencies replace mutable test hooks. The separately qualified S-025
  correction binds attachment/dependency/event inputs, rejects stale-binding reuse,
  retains terminal failure and supports local waiter cancellation without destroying
  the owner or changing another caller's admission. Recovery uses saved exact proof.
- Readiness, namespace isolation, quarantine, exact cleanup, stable reset, TLS and
  safe failure publication retain meaningful negative and real-service witnesses.
- Fake failures preserve state unless explicitly modeling response loss after
  mutation; fake namespace deletion cannot erase remaining objects. Incidental
  description and request-count constraints are retired with meaningful successors.
- Deleted-object checks require typed not-found; sibling payloads survive unchanged;
  stable reset preserves bucket identity and supports a successful new mutation.
- Three coherent service-free groups and two real integration scenarios replace
  mixed selectors and redundant setup. Distinct publication, native lifecycle,
  application-process and production-adapter witnesses remain. Authored/generated/
  nested selectors agree without duplicate or compatibility routes.
- Finalization precedes a frozen compatible validation portfolio; required cleanup
  and evidence audits pass, historical failures remain and acceptance limits are
  explicit. Documentation lint alone never establishes production guarantees;
  completion has no prescribed test-count reduction quota.
