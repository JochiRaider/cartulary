# Test Harness Resource Lifecycle Remediation

Status: explicit-suite and lane hard cutover implemented; current completion
qualification is recorded in the final section. Earlier lifecycle runs below
are historical and do not qualify the cutover. This document is a development
handoff, not conformance or release evidence. The adopted
[Testing Harness NLSpec](../testing-harness-nlspec.md) governs mechanics; Core 04
and Core 05 continue to govern production security and publication respectively.

## Earlier lifecycle cutover: scope and decisions

This is a pre-production hard cutover. PostgreSQL services already have an
invocation owner; the waste comes from replaying cluster and database-local
provisioning after template copies. Renderer containers previously belonged to
each browser group. Reuse moves physical renderer ownership into the native
fixture owner while retaining independently owned browser connections, fresh
test contexts, database clones and existing application reset boundaries.

One invocation owns its renderer pool. No renderer sharing crosses chats or
invocations. That initial pool had at most four slots. The subsequent cutover below
replaces this limit with topology-owned four-default/six-maximum policy, still
subject to host admission. PostgreSQL may instead belong to an explicitly attached local session;
that owner retains both physical cleanup authority and resident capacity.
Measurement and seeded review use dedicated lifetimes through the same native
allocation path. Explicit dev-origin review keeps its separately specified host
browser and charges that browser throughout its review session; it cannot borrow
a canonical renderer or alter the borrowed application service. Persistent
Chromium reuse across test groups is excluded.

No new baseline, benchmark, speedup claim, golden refresh or broad coverage
consolidation belongs to this change. Structural acceptance is sufficient:
initialize each physical PostgreSQL generation once, avoid global provisioning
on ordinary clones, reuse compatible renderer generations, and release physical
capacity only after proven cleanup.

## Remediation workstreams

| Gap | Remediation and affected areas | Rationale and lasting benefit | Migration impact | Risk if unresolved | Completion criteria |
| --- | --- | --- | --- | --- | --- |
| Conflicting lifecycle requirements | Specification, contracts, tests and guides distinguish physical allocations, logical borrows and active work; demand-created clones replace replenishment wording | One authoritative model prevents helpers from restoring unnecessary setup | Replace current contracts together; historical evidence stays historical | Conflicting requirements, stale acceptance | Requirements 803, 804, 806, 811 and 816 agree with producers and readers; acceptance 085/086 includes ownership and isolation |
| Cluster and database preparation coupled | Implementation splits native cluster initialization, fresh database preparation and clone finalization; tests retain positive/negative privilege coverage | Removes global role DDL and extension/ACL replay from each clone | Monolithic provisioning API and implicit owned-service fallback removed | Serialized clones and consumers repairing shared prerequisites | Initialization precedes publication; current policy/generation identity required; clones finalize under their target lock with connections disabled |
| Template identity omits provisioning | Native compiled policy digest binds executed SQL, pinned image, initdb and TLS configuration; migration and policy identify templates | Invalidates incompatible services/templates without duplicate digest algorithms | Old local sessions must be stopped and recreated | Stale privileges or templates accepted | Native attachment checks reject each missing or incompatible identity; JavaScript validates the native capability |
| Per-stack CA prevents renderer reuse | Native PKI owner keeps a run/profile authority in memory, issues distinct leaf keys and combines its public root with the exact object-store root | Stable immutable trust permits reuse without accumulating roots or sharing signing keys | Replace stack/trust attachments together; interrupted owners recover only by cleanup | TLS bypasses, key leakage or repeated trust installation | Same-scope roots match, leaf keys differ, foreign authority/wrong name reject, partial issuance fails closed |
| Retained resources disappear from accounting | Native host admission tracks durable resident allocations separately from active grants; cleanup proofs retain capacity after owner death | Safe sharing under capacity pressure with one scheduler and explainable ownership | Admission state generation changes; drain old invocations before deployment | Oversubscription, orphaned capacity, cleanup or measurement deadlock | Resident claims persist, borrowers do not double-charge, queued work holds no host grant, cleanup and exclusive work progress |
| Groups own renderer startup and teardown | Native slot owner provisions, lends, settles, quarantines and retires; group and seed consumers require private bound attachments | Removes repeated package staging, trust installation and renderer identity checks | Per-group ownership and optional canonical local launch removed | Repeated startup and borrowers destroying shared infrastructure | Compatible groups reuse a generation; incompatible identities reject; interrupted clients retire; exact removal settles once |
| Evidence assumes one physical lifetime per group | Logical results carry borrow/allocation/generation IDs and physical attestation references; existing resource journal controls recovery | Independent row outcomes remain attributable despite infrastructure sharing | Changed result/schema generation adopted atomically | Stale or ambiguous acceptance | Every selected row has exactly one outcome and shared physical resources have one lifecycle with terminal cleanup |
| Validation repeats expensive setup | Extend current tests with fake-adapter transition cases and a real multi-group lifecycle case; retain run-bound build receipts | Keeps assertions tied to observable guarantees while reducing development cost | Preserve semantic row identities where assurance is unchanged | Mechanism-specific tests hide regressions or keep the old cost | Narrow milestone gates, one consolidated integration gate and one comprehensive completion gate |

## Ownership and isolation contracts

1. PostgreSQL cluster roles, logins, memberships and parameter grants are fixed
   before service publication. Ordinary consumers may create isolated databases
   but do not alter cluster prerequisites. Every connection still performs its
   production purpose, role, TLS, version and checksum checks.
2. A migrated template is sealed and never a writable test database. Template
   copies use `WAL_LOG`, preserve normal durability and start with connections
   disabled. Finalization changes only the target database's grants and enables
   connections. A post-create failure may delete only the target proven created.
3. Native renderer compatibility binds run, runtime profile, admitted renderer
   identity, immutable trust and launch options. Suite name and application
   affinity do not select renderer ownership. Only one borrower occupies a slot.
4. A group closes its Playwright clients and contexts. The lifecycle adapter
   checks the exact container for a live server and absence of client sockets,
   connection-owned browser descendants and forwarding listeners. Only closed
   results leave that inspection. Uncertainty retires the generation; product
   work is never retried to hide a failure.
5. Resident reservations retain memory, process, port and service occupancy.
   Acquisition, execution, reset, inspection and cleanup require active
   admission. Resource ownership survives controller death until exact cleanup
   proves release. Admission mutexes never enclose Docker operations.
6. Cancellation stops admission and drains exact owned descendants. Cleanup
   failures preserve primary failures, unresolved dependencies and capacity.
   Repeated logical release observes the recorded settlement; explicit recovery
   creates a separate bounded attempt against retained proofs.

## Ordered implementation and gates

| Phase | Dependencies and components | Principal risk | Exit criteria and minimum validation |
| --- | --- | --- | --- |
| 0. Authority and cutover | NLSpec, acceptance clauses, authored schemas/topology and guides | Partial contract adoption or edits conflicting with another chat | Review source changes together; documentation lint; old invocations drained before executing current contracts |
| 1. PostgreSQL | `pgtest`, `postgrescatalog`, test-service owner, recovery/performance clone consumers, local-session metadata | Privilege regression or deleting a pre-existing target | Focused helper cases and one owned PostgreSQL selection cover initialization, attachment rejection, grants, durable clones, concurrent isolation and failure cleanup |
| 2. Trust and accounting | `browserpki`, native admission, residency and dependency cleanup; requires phase 0 | Key scope error, uncharged resources, deadlock | PKI cases and native fake-owner/admission cases cover immutable roots, leaf separation, durable accounting, owner death and drain |
| 3. Renderer owner | Native fixture composition, physical adapter and group attachments; requires phase 2 | Reuse before connection settlement | Fake transition tests plus the installed multi-group lifecycle case cover reuse, fresh storage, client interruption and retirement |
| 4. Integrated replacement | Graph compilation, seeded review, populated clones, evidence readers and generated projections; requires phases 1–3 | A stale executable path or schema remains | One consolidated selection covers functional, visual/accessibility, reset, populated-fixture and dedicated measurement lifecycles; source/schema checks reject stale projections |
| 5. Completion | Stable implementation and current generated output | Incomplete evidence or unnecessary reruns | `make agent-finalize`, applicable source checks, one `make ci`, reviewed diff and final handoff; no separate repeat of `check` |

Use `make task-guide` and exact catalog selections through `make test-slice`.
Diagnose a failed selection before rerunning; rerun the affected selection only.
Do not run another warm suite merely to supply `RESULTS_DIR`; report retained-run
maintenance as skipped when no compatible successful full warm result exists.

## Contract and component map

| Obligation | Machine contract and implementation | Focused existing evidence |
| --- | --- | --- |
| 803: physical lifetime and admission | Native admission generation 2; runtime-resource v2 and resident-allocation v1; native residency/dependency cleanup | Admission weighted-tree, quiet ordering, native ownership/recovery tests |
| 804/811: PostgreSQL provisioning and isolation | Service lease v2, PostgreSQL provisioning v1, native compiled digest and three provisioning stages | PostgreSQL helper, template preparation, privilege and migration owner selections |
| 804/806: renderer borrowing and trust | Work-graph v7, topology v13, renderer physical/attachment/event/startup v1, browser-group-result v9 | Native fixture adapter, PKI, installed physical lifecycle and browser group selections |
| 816: explicit persistent attachment | Local-session v2, native generation validation and durable owner reservation | Local-session contract and native borrowed-service lifecycle |
| Recovery and independent evidence | Lifecycle policy v2, exact resource journal, logical borrow results and shared attestations | Existing interrupted acquisition, failed cleanup, row accounting and evidence-reader selections |

Runtime code and generators consume authored machine inputs only. This mapping
is for review; neither it nor other Markdown is executable acceptance input.

## Handoff requirements

Completion requires the final implementation mapping, removed legacy paths,
current interface/lifecycle summary, exact commands and row selections, run roots
and process exits, cleanup outcomes, skipped checks and remaining defects.
Do not treat an unresolved lifecycle or isolation failure as a performance
follow-up. Human review of specification projection remains necessary; passing
tests do not prove specification completeness or authorize publication.

## Implementation record

The replacement is implemented in native `pgtest`, `harnessadmission` and
`harnessfixtures`, with the existing Node browser adapter limited to physical
renderer operations. `ProvisionDatabase` and implicit PostgreSQL owned-service
fallback are removed. Browser groups no longer create or remove containers,
stage packages, install trust, or copy physical attestations into each group.
Group results bind the exact unit and logical borrow to one physical attestation
and trust record. Current retained readers verify those references and the borrow
event; target finalization rejects duplicate borrow attribution.

The retired schema generations are work graph v6, topology v12, runtime resource
v1, lifecycle policy v1, service lease v1, local session v1 and browser group
result v8. Producers, readers, synthetic fixtures and generated validators now
use the current generations listed above. No translator or legacy launch mode
was added. Existing image/font and public trust attestation schemas are unchanged.

The host admission state generation is 2. Before adopting it, drain previous
harness invocations and close old local sessions through their owners. Do not
translate live admission state, remove sibling resources, or attach superseded
capabilities. The configured port-lane ceiling is now 16, still CPU-bounded:
resident database, object-store, stack and renderer listeners consume these lanes
throughout their physical lifetimes instead of disappearing with group activity.
The capacity table also now explicitly records the existing 25 percent CPU
safety margin and three object-store lanes. These conservative host bounds remain
intentional under the new model: resident service occupancy and additional
execution claims both consume the declared budget. They are reservations, not
measured usage or a throughput claim. The former prose values of four
object-store lanes and four port lanes must not be used to interpret current
admission evidence.

Validation uses the default result parent `.cartulary/test-results/`. Each run
name below identifies its exact directory, never a newest-run selector. Commands
were launched with `CARTULARY_TEST_RUN_ID` set to the recorded run name. A future
invocation must use a fresh name; retained directories must not be cleared or
reused.

| Command/selection | Run | Process result and interpretation |
| --- | --- | --- |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_fixture_lifecycle` | `lifecycle-installed-owner-20261018c` | Exit 0: real multi-group reuse, fresh cookies/storage, interrupted client, borrowed services and exact recovery |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.postgres_template_preparation,harness.browser.boundary_support.local_session` | `lifecycle-installed-owner-20261018a` | Exit 2: PostgreSQL clone/ACL/concurrency and local-session selections passed; renderer acquisition failed. Superseded by the repaired installed lifecycle run; all outer cleanup steps completed |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_fixtures,harness.browser.unit.pgtest_helper_contract,harness.browser.boundary_support.playwright_adapter,harness.browser.boundary_support.ui_review_contract` | `lifecycle-owner-units-20261018c` | Exit 0: native transitions, compatibility rejection, independent borrow evidence and retained-reader checks |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission` | `lifecycle-admission-20261018b` | Exit 0: weighted accounting, durable ownership, owner death and bounded drain |
| `make test-slice OWNER=platform.postgres ROWS=platform.postgres.support_unit.certificate_fixture_lifecycle` | `lifecycle-pg-policy-20261018a` | Exit 0: initialize-before-publication and terminal readiness/initialization failures |
| `make agent-finalize` | `lifecycle-finalize-20261018c` | Exit 0; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make source-check CHECK=generated` | `lifecycle-source-generated-20261018a` | Exit 0 |
| `make source-check CHECK=schema` | `lifecycle-source-schema-20261018a` | Exit 0 |
| `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | `lifecycle-graph-contract-20261018b` | Exit 0: active claims and physical allocations agree with authored topology |
| `make test-slice OWNER=module.auth ROWS=module.auth.browser.browser_sign_in_exposes_the_ordinary_session_sur_2ec8df6229,module.auth.browser.invalid_credentials_leave_no_session_cookie_and_47a0f6a4d2,module.auth.visual.capture_the_anonymous_auth_gateway_across_initia_755030aa99,module.auth.accessibility.verify_session_mfa_incident_forbidden_loading_an_e8f5d5f3a9` | `lifecycle-browser-integration-20261018b` | Exit 0: four independently accounted catalog rows, 16 work units and all 31 cleanup steps completed; concurrent groups used separate generations, while the installed lifecycle selection proves sequential reuse |
| `make ci` | `lifecycle-ci-20261018a` | Exit 130 after a confirmed admission stall: 924 work units passed, 6 failed and 272 were cancelled. All 1,299 cleanup steps completed; host leases and resident reservations returned to zero |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission,harness.evidence_accounting.behavior.native_collection` | `lifecycle-repair-admission-20261018a` | Exit 2: collection passed; admission retained a superseded strict-FIFO assertion, subsequently corrected |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission` | `lifecycle-repair-admission-20261018b` | Exit 0: retained 14-of-16 port reservations, oldest-fitting admission, exclusive priority and unsuccessful request withdrawal |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_fixtures,harness.browser.boundary_support.ui_review_lifecycle` | `lifecycle-repair-fixtures-20261018a` | Exit 0: idle retirement, concurrent last-consumer settlement and current service-lease fixtures |
| `make agent-finalize` | `lifecycle-repair-finalize-20261018a` | Exit 0; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_graph_observations,harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.ui_review_seeded_default,harness.browser.integration.ui_review_seeded_network_flow_claimed` | `lifecycle-repair-installed-20261018a` | Exit 2: five work units passed, including native fixture lifecycle, real container observations and claimed-network-flow review; default review exposed sibling diagnostic admission. All 15 cleanup steps completed and host reservations returned to zero |
| `make agent-finalize` | `lifecycle-review-finalize-20261018a` | Exit 0 after binding review operations beneath their fixture's active reservation; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.ui_review_seeded_default,harness.browser.integration.ui_review_seeded_network_flow_claimed` | `lifecycle-review-repair-20261018a` | Exit 0: both real seeded review workflows passed; all 13 cleanup steps completed |
| `make lint CHECK=go` | `lifecycle-lint-go-20261018a` | Exit 2: formatting and vet passed; staticcheck found two obsolete unused declarations, subsequently removed |
| `make agent-finalize` | `lifecycle-static-finalize-20261018a` | Exit 0 after removing obsolete declarations; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make lint CHECK=go` | `lifecycle-lint-go-20261018b` | Exit 0: formatting, vet and staticcheck |
| `make go-gosec-targeted` | `lifecycle-security-targeted-20261018a` | Exit 0 |
| `make go-gosec-audit` | `lifecycle-security-audit-20261018a` | Exit 0 |
| `make go-vulncheck` | `lifecycle-vulncheck-20261018a` | Exit 0 |
| `make source-check CHECK=generated` | `lifecycle-source-final-20261018a` | Exit 0 after final source cleanup |
| `make migration-drift` | `lifecycle-migration-20261018a` | Exit 0: full-chain fresh database migration/schema bootstrap and authored migration drift |
| `make otel-conformance` | `lifecycle-otel-20261018a` | Exit 0 |
| `make lint CHECK=shell` | `lifecycle-lint-shell-20261018a` | Exit 0 |
| `make test-slice SUITE=backend-process` | `lifecycle-process-20261018a` | Exit 0: all 17 work units and all 40 cleanup steps completed |
| `make browser-e2e SUITE=all` | `lifecycle-browser-final-20261018a` | Exit 130 after correcting an overbroad selection: `all` includes 29 release-tier measurement rows beyond CI. The evidence-heavy snapshot rejected insufficient `/tmp` capacity, causing two dependent summary failures. At interruption: 21 units passed, 3 failed, 3 skipped and 335 cancelled. All 58 cleanup steps completed and host reservations returned to zero |
| `make browser-e2e SUITE=functional` | `lifecycle-browser-functional-20261018a` | Exit 0: all 153 work units and all 308 cleanup steps completed; independently retained group results demonstrate repeated renderer generations |
| `make format` | `lifecycle-eviction-format-20261018a` | Exit 0 after pressure-aware retirement changes |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission` | `lifecycle-eviction-admission-20261018a` | Exit 0: initial pressure-aware retirement regression cases |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_fixtures` | `lifecycle-eviction-fixtures-20261018a` | Exit 0: retire exactly one eligible pool slot |
| `make agent-finalize` | `lifecycle-eviction-finalize-20261018a` | Terminal summary passed after centralizing weighted pressure accounting; outer process exit was not retained in the handoff. Retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission` | `lifecycle-eviction-admission-20261018b` | Exit 0: shared weighted calculation, required ancestor claims, temporary versus resident pressure and exact release boundaries |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_fixtures` | `lifecycle-eviction-fixtures-20261018b` | Exit 0 after final pressure-accounting refinement |
| `make lint CHECK=go` | `lifecycle-eviction-lint-20261018a` | Exit 0: formatting, vet and staticcheck after final refinement |
| `make source-check CHECK=generated` | `lifecycle-eviction-generated-20261018a` | Exit 0 after final refinement |
| `make browser-e2e SUITE=stateful` | `lifecycle-browser-stateful-20261018a` | Exit 0: all 43 work units and all 86 cleanup steps completed; 17 groups reused exactly two physical renderer generations, separated by runtime profile |
| `make browser-e2e SUITE=accessibility` | `lifecycle-browser-accessibility-20261018a` | Exit 0: all 21 work units and all 42 cleanup steps completed |
| `make browser-e2e SUITE=visual` | `lifecycle-browser-visual-20261018a` | Exit 0: all 97 work units and all 194 cleanup steps completed against existing goldens; 43 default-profile groups reused one generation, with the claimed-profile group isolated in another |
| `make browser-e2e SUITE=support` | `lifecycle-browser-support-20261018a` | Exit 0: all 20 work units and all 39 cleanup steps completed |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.measurement.ac047_anchor_trace` | `lifecycle-populated-clone-20261018a` | Exit 0: all 15 work units and all 28 cleanup steps completed; sealed populated-database snapshot, clone finalization and dedicated measurement renderer exercised by the existing acceptance scenario |

Earlier development failures exposed missing standalone schema registration,
zero-valued resident quantity validation, a missing trust-digest handoff, a stale
active-claim assertion, and the topology key allowlist. They were repaired within
the affected owner paths. An attempted fixture-build reuse across distinct run
identities was removed because it violated receipt binding. These failures do
not justify weakening artifact, privilege, cleanup or isolation checks.

Completion gate status is recorded below. No new benchmark,
performance baseline, golden refresh, release/package target, or extra warm run
for retained-run maintenance has been performed. The unintended release-tier
measurement selection is recorded explicitly above and below.

The comprehensive run exposed head-of-line blocking: idle application stacks
occupied 14 of 16 port reservations, the oldest queued request required three
additional reservations, and strict host FIFO prevented fitting consumers from
running. The controller was alive but no active work had admission. The run was
interrupted through its verified native execution session; unrelated resources
were not swept. The repair admits the oldest fitting shared request while
preserving exclusive priority, withdraws unsuccessful pooled-renderer admission
before releasing the logical slot, and retires idle ordinary stacks before
exclusive allocations. Reacquisition uses the same fresh database or sealed
snapshot path; reset operations remain in place. Focused regression coverage
establishes those boundaries; the affected installed-owner selections and
remaining browser suites subsequently passed.

Other failures identified stale acquisition fixtures without the current
PostgreSQL capability field and a collector still constructing the former
five-dimensional admission quantities. The collector is a runtime consumer as
well as a test fixture; both require the current complete resource domain.

The installed repair selection also exposed a review ownership error: diagnostic
work was a sibling of the fixture's active reservation and could exceed capacity
by counting the same execution twice. Seeded review now places browser,
diagnostic and artifact operations beneath that fixture reservation. Independent
physical allocations remain charged; this changes nesting, not capacity limits.

The `all` browser selection also included release-tier performance qualification,
which was not needed for this cutover. Its evidence-heavy physical object fixture
rejected the 16 GiB tmpfs before populating it; the declared population and
admission guard were preserved. That run was stopped through its verified native
session to avoid unnecessary qualification work, not treated as a pass. Remaining
CI browser coverage uses the functional, stateful, accessibility, visual and
support selectors. Populated-database clone coverage uses one explicit fixture
scenario, without repeating all release-tier measurements or changing scratch
storage, goldens, population sizes or performance thresholds.

Functional validation exposed avoidable renderer churn in the first admission
repair: every temporary capacity miss could retire healthy idle renderers, even
when their resources could not relieve the blocking service workload. The native
admission owner now uses the same weighted capacity calculation to distinguish
temporary work from resident pressure, preserving required ancestor claims.
The physical owner retires one eligible allocation and re-evaluates; a full pool
also evicts only the one slot needed. Last-consumer retirement and exclusive
measurement cleanup remain unconditional. Focused tests cover transient memory
pressure, persistent port pressure, unrelated resources, ancestor claim floors,
and replacement of exactly one stable slot.

The completed stateful run provides direct structural evidence after that final
refinement. Its 13 default-profile groups reference generation
`6d2d2e30-7058-487c-a085-fa2366c56aee`, while its four
`network_flow_claimed` groups reference generation
`4e06bbe8-749a-4ddd-910c-019cb791d2ea`. Each group retains its own borrow/result
and the existing intervening application resets. This is lifecycle evidence,
not a benchmark or an estimate of speedup.

## Authorized CI completion and contention repairs

The subsequent request authorized another complete `make ci`, failure repair,
cleanup verification and this handoff update. The first completion attempt,
`lifecycle-ci-completion-20261018a`, exposed four installed-boundary failures and
then a confirmed admission stall. It ended through its verified native session
with outer exit 130: 938 passed, four failed and 260 cancelled work units. All
1,316 cleanup records completed, the retained-secret scan passed, and host leases
and residents returned to zero. `harness.deployable_shape` passed in this run;
its former unexecuted-operation gap is closed.

The four failed rows were native fixture lifecycle, native selected-service
lifecycle, seeded default UI review and seeded claimed-network-flow UI review.
The borrowed native fixture case reported persistent-owner startup failure
without its child stderr; the immediate cause of that historical failure is
unproven. The selected-service case exhausted its nested object-store selection
deadline. Both review cases reported inner browser-stack acquisition expiry,
while their outer assertions retained the published product classification.
After independent work drained, idle stacks still held 14 of 16 port lanes and
every dispatched browser request required three more. Renderer-only eviction
could not release those stack reservations. Cancellation followed this proven
stall, not merely a failed row or silent output.

The repairs keep host limits and product assertions unchanged:

- `harnessadmission` preserves an older shared waiter's priority when independent
  active work can drain enough capacity. Backfill is permitted only when durable
  resident pressure still blocks it. The same weighted calculation preserves
  required ancestor claims and exclusive fences.
- `harnessfixtures` retires one eligible idle renderer, rechecks admission, and
  can retire one idle application stack when renderer retirement cannot relieve
  persistent pressure. A pending logical borrower pins its chosen stack
  generation before admission; active and reserved stacks cannot be evicted.
  Retirement uses the existing bounded drain and settlement path, outside the
  admission-state mutex. Fake-provider coverage exercises seven retained stacks
  occupying 14 port lanes and proves that exactly one retirement admits the next
  stack and all allocations eventually settle.
- Local-service creation and closure await the existing native admission queue.
  Their lifecycle lock spans the wait; no descriptor or service is published
  before a grant. Cancellation withdraws queued work, cleanup retains durable
  resident ownership, and startup failures retain bounded child stderr.
  Existing preparation and drain limits remain unchanged.
- The authored `harness_installed_browser` profile declares the actual nested
  peak: four CPU, one I/O, 2,048 MiB, 20 processes, five port lanes, two PostgreSQL
  lanes, one object-store lane and one browser stack. Native fixture lifecycle
  and both seeded review rows use it with no unused outer service dependencies.
  The native fixture row uses the PostgreSQL-capable default runtime profile.
  This changes claims and routing, not physical host capacity.

REQ-803 and REQ-816, the implementation guide, authored topology, applicable
schemas, catalog validation and generated projections were updated together.
The final restore repair also removes the remaining contradictory asynchronous
database-supply sentence in the specification. No clone pool is implemented.

All commands below use the default result parent and the exact run name as
`CARTULARY_TEST_RUN_ID`, as in the earlier execution table.

| Command/selection | Run | Observed result |
| --- | --- | --- |
| `make ci` | `lifecycle-ci-completion-20261018a` | Exit 130; 938 passed, four failed, 260 cancelled; all 1,316 outer cleanup records completed |
| `make generate` | `lifecycle-ci-repair-generate-20261018b` | Exit 0 |
| `make format` | `lifecycle-ci-repair-format-20261018a` | Exit 0 |
| `make agent-finalize` | `lifecycle-ci-repair-finalize-20261018d` | Exit 0 before focused repair validation |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.native_fixtures,harness.browser.boundary_support.local_session` | `lifecycle-ci-repair-units-20261018a` | Exit 0; three units and all 12 cleanup records completed |
| `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | `lifecycle-ci-repair-routing-20261018a` | Exit 0; two units and all 11 cleanup records completed |
| `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.observation_admission` | `lifecycle-ci-repair-admission-20261018c` | Exit 0; one unit and all ten cleanup records completed |
| `make agent-finalize` | `lifecycle-ci-repair-finalize-20261018e` | Exit 0 after the final fairness assertions, before installed validation and CI |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_fixture_lifecycle,harness.browser.integration.native_selected_service_lifecycle,harness.browser.integration.ui_review_seeded_default,harness.browser.integration.ui_review_seeded_network_flow_claimed` | `lifecycle-ci-repair-installed-20261018a` | Exit 0; all four formerly failed rows passed, all 13 cleanup records completed, retained-secret scan passed, zero host leases/residents |
| `make ci` | `lifecycle-ci-repaired-completion-20261018a` | Exit 2; 1,200 passed and two failed units, zero cancelled or skipped; all 1,897 outer cleanup records completed |

Development checkpoints were retained rather than overwritten.
`lifecycle-ci-repair-finalize-20261018a` and `b` exited 2 for the omitted resource
profile allowlist and unsorted authored claims. Read-only target-plan inspection
then identified the stale compiled profile enum and the native fixture's
incompatible `none` runtime profile. `lifecycle-ci-repair-generate-20261018a`
exited 2 for that runtime mismatch; generation `b` corrected the projections.
`lifecycle-ci-repair-finalize-20261018c` exited 2 for stale generated topology
before the successful generation/format/finalization sequence above.
`lifecycle-ci-repair-admission-20261018a` and `b` exited 2 because old tests still
expected transient-pressure backfill; the corrected fairness assertions passed
in `c`. Only the affected narrow selections were repeated for these failures.

## Final CI evidence and restore cleanup repair

The complete, uninterrupted command was:

```sh
CARTULARY_TEST_RUN_ID=lifecycle-ci-repaired-completion-20261018a make ci
```

Its native process exited 2 after retained scanning. The exact run root is
`.cartulary/test-results/lifecycle-ci-repaired-completion-20261018a`.
The manifest binds command `cartulary.harness.command.ci.v4`, source digest
`sha256:d8d29b7a5dc20e3be1261159909dbb0c3131544698f0a18ad5cae513b8c0e5fb`
and graph digest
`sha256:fa0aa68761f64b8d08d53deb0f43fb8429427f9a3b525ecf363b4e4e35e136b2`.
The cache summary reports one hit, 682 misses and 519 bypasses. The four earlier
installed failures and `operation:harness.deployable_shape` all passed in this
full invocation. No selected work was cancelled or skipped.

The remaining failure was
`browser_group:webserver-backed:functional-support-default-restore`, reported as
`harness/cleanup_error`, unit exit 12. Its dependent
`browser_target_summary:browser-e2e-webserver-backed` carries the same failure;
these are one underlying defect, not two independent failures. The browser row
was
`module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0`.
Its target database's direct forced-drop operation exhausted the fixture's
10-second aggregate context while PostgreSQL reported `CheckpointStart`, a
checkpointer waiting on `DataFileSync`, and no blocking sessions. The source
database subsequently dropped successfully. This evidence identifies a cleanup
failure; it is not a product assertion failure or a renderer admission stall.

`recoverybrowsertest` now uses the existing common normal-drop/fallback owner
with separate five-second normal and 15-second forced-operation limits, each
with its own target admission. It no longer truncates them with a local aggregate
database context. The existing browser process drain guard remains unchanged.
Acquisition no longer force-drops a proposed database name before CREATE; only
successful CREATE proves ownership for post-creation failure cleanup. The
installed drop-observation test also verifies that a collision preserves the
existing database identity and that an owned clone with an active connection is
deleted without disturbing its peer. Common cleanup tests already cover parent
cancellation, normal-attempt expiry, bounded fallback and admission release.

The full invocation's outer lifecycle journal contains 1,897 completed records.
That does not erase the nested eager-cleanup failure: the suite owner ultimately
removed its physical service. The retained-secret scan passed over 11,934 files,
and host admission contained zero leases and zero residents after termination.
No sibling resources were swept.

After the user's final wrap-up instruction, verification was limited to the
affected selections and static/documentation checks. No further full CI was
launched. The repair's observed results are:

| Command/selection | Run | Observed result |
| --- | --- | --- |
| `make format` | `lifecycle-restore-format-20261018a` | Exit 0 |
| `make agent-finalize` | `lifecycle-restore-finalize-20261018a` | Exit 0 before the final focused validation; retained-run maintenance skipped with `RESULTS_DIR` unset |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.recovery_drop_observation` | `lifecycle-restore-contract-20261018a` | Exit 0; all three work units and all 14 cleanup records completed; retained-secret scan passed |
| `make test-slice OWNER=module.recovery ROWS=module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0` | `lifecycle-restore-browser-20261018a` | Exit 0; all 12 work units and all 23 cleanup records completed; the formerly failing restore row passed |
| `make lint CHECK=go` | `lifecycle-restore-lint-go-20261018a` | Exit 0: formatting, vet and staticcheck |
| `make lint-markdown` | `lifecycle-ci-handoff-20261018a` | Exit 0; `adhoc/lint-markdown/tool-run-summary.json` reports pass |

The repair selection's source digest is
`sha256:117ab34186bebf1d14516d93ad7e7e71e821c517d552f0b7d898a290b8e15e20`.
The only executable changes after the full CI are the recovery fixture's cleanup
and collision behavior and its regression case. Both focused validation
processes terminated successfully. Host admission again contained zero leases
and zero resident allocations after their cleanup. No known failing lifecycle or
isolation assertion remains in these focused selections; the absence of another
complete CI pass remains an acceptance limitation, not a claimed success.

A read-only discovery attempt using `make target-plan TARGET=test-slice ROWS=...`
was rejected because `target-plan` does not accept `ROWS`. It launched no tests.
The two explicit owner selections above used the discovered public task guide.

## Completion scope

The implementation, specification and current contract generations are in place.
The previous deployable-shape execution gap is closed. Production assertions,
application resets, existing goldens and isolation boundaries remain intact.
The historical interrupted runs and the final failed CI remain failures in their
original evidence; focused repairs do not rewrite them into a clean CI result.
A clean comprehensive pass on the final repaired source remains unestablished.

Retained-run maintenance was skipped because `RESULTS_DIR` was unset and no
compatible successful full warm check was available at finalization. No separate
`check`, extra warm run, golden refresh, new benchmark/baseline, or release/package
target was added. The earlier accidental release-tier selection remains recorded
above. Human review of specification projection and Core 05 publication remain
separate obligations; test results do not replace them.


## Explicit-suite and lane cutover (2026-10-18)

This section records the subsequent cutover on source revision `808203eaf` plus
uncommitted worktree changes. It does not reinterpret earlier failed/interrupted
runs or claim their passes qualify the current source. Completion evidence below
will identify the final source digest rather than inventing a commit.

### Workstreams and sequencing

| Phase / gap | Remediation and areas | Rationale and durable benefit | Migration / unresolved risk | Exit criteria |
| --- | --- | --- | --- | --- |
| 1: implicit suite inclusion | NLSpec REQ-800/054/056/082, all family manifests, work-graph owner, selection compiler and diagnostics adopt closed aggregate placement | A single row contract excludes optional expensive assurance before readiness | v9 catalogs and the old field reject; no aliases. Leaving policy expansion unconstrained can silently restore excluded work | Four aggregates and broad owner/suite selections exclude 58 rows; dedicated selectors and exact rows remain complete; invalid dependency fails before provisioning |
| 2: uncertain fixture reuse | PostgreSQL cleanup and object-store helper implementation/tests enforce bounded settlement, poisoning and explicit attachment | Shared infrastructure does not imply shared dirty state or hidden provisioning | Implicit object-store callers must declare attachment or use owned lifecycle API. Ignored cleanup can contaminate later tests | Independent five-second rollback/close budgets, cancellation-safe cleanup, later borrowers rejected, ownership retained |
| 3: competing browser controls | Topology owns four-default/six-maximum policy; public inputs, runtime plan, native slot owner and worker config use it | One concurrency control and one account/worker per group; resource capacities stay authoritative | Retire PLAYWRIGHT_WORKERS and browser SUITE=all; no translation. Divergent planner/pool limits cause wasted provisioning and contention | Synthetic six slots, waiting seventh, invalid-input/profile rejection, pressure/exclusive admission and fresh state remain proved |
| 4: excess fixture/environment dependencies | Split Assessments pure row, correct Evidence dedicated fixture, rename two Node-only files and update Vite/runner/discovery/catalog | Dependency declarations match assertions and avoid needless PostgreSQL/jsdom setup | No old-path aliases or assertion deletion; orphaned/duplicate selection would lose assurance | Same symbol/title union exactly once; pure rows have no services/DOM; Evidence retains real concurrent database behavior |
| 5: misleading qualification | Advance current schemas, generate projections, revise guides/recipes and retain exact completion evidence | Current contracts and result descriptions agree without executable Markdown dependencies | Historical results stay immutable. Aggregate success cannot mean omitted suites passed | Static/narrow checks, finalization and one consolidated CI plus explicit Visual/UI-Review completion gate |

Phases 2 and 3 depend on phase 1; phase 3 also depends on lifecycle safety.
Phase 4 depends on catalog migration. Phase 5 closes all source and generated
changes. Synthetic tests replace no product assertions or physical completion
checks. Six-way physical throughput is host-dependent and is not promised by
six-slot synthetic qualification. The later user-authorized capacity increase
is recorded below; physical resource accounting remains authoritative.

No-change dispositions: production password hashing, per-test report-renderer
browsers, potential semantic duplicates, Chromium lifetime, authenticated-context
isolation, backend resets and migration replay retain their assurance contracts.
No benchmark, baseline, golden refresh, production API or database migration is
required. No CI-provider integration or new release qualification gate is added.

### Development validation

- `make generate`: initial run `20261018T212435Z-p41965` failed on the new
  shell row's incorrect Go-only verification binding. Corrected the authored
  verification; `cutover-generate-02` and `cutover-generate-03` passed, outer exit 0.
- Initial focused/format attempts stopped in graph preparation (outer exit 2).
  Corrected topology default projection and removed the remaining default
  UI-Review operation. `cutover-format-02` passed, outer exit 0.
- `cutover-focused-lifecycle-02`: combined native fixtures, pgtest helper, s3test
  helper, object-store admission and browser-lane-policy rows passed, outer exit 0.
  Native fixture and lane-policy checks used synthetic providers. The admission
  row used its declared shared object-store service; this selection did not
  provision PostgreSQL or a real browser.
- Command-surface validation: `cutover-focused-command-01` passed public registry
  parity but failed native runtime on an outdated nested aggregate row count.
  The assertion now excludes the 58 explicit-only rows;
  `cutover-focused-runtime-02` passed the native runtime selection (outer exit 0).
- Routing: `cutover-focused-routing-01` exposed outdated default UI-Review and
  broad Visual expectations plus a task-guide schema that required a nonempty
  optional-command list. After correction, `cutover-focused-routing-02` passed
  (outer exit 0).
- Pure selections: `cutover-pure-assessments-01`, `cutover-pure-ui-node-01` and
  `cutover-pure-views-node-01` passed (outer exit 0). The two renamed source files
  retain their assertions and execute under Node, outside jsdom.
- Worker/lane validation: `cutover-focused-workers-01` passed strict one-worker
  fixtures and lane policy. `cutover-lane-final-01` also passed actual compiler
  graph affinity checks at four and six lanes (outer exit 0).
- Formatting/generation: `cutover-format-03` and `cutover-generate-05` passed.
  `cutover-generated-check-01` passed generated verification. Schema validation
  found the second topology shape reader missing `browser_lane_policy`; updating
  it required refreshing its render index. `cutover-generate-06` and
  `cutover-schema-check-03` then passed (outer exit 0). Earlier schema roots
  `cutover-schema-check-01/02` remain failed evidence.
- `cutover-markdown-01` passed Markdown lint; later handoff prose receives a final
  Markdown-only check. `cutover-finalize-01` passed generated refresh, source,
  schema and catalog validation. Its retained-run evidence and scheduler
  maintenance actions were explicitly skipped because RESULTS_DIR was unset.
- `cutover-static-01` passed `make lint CHECK=standard`, including Go, TypeScript,
  frontend, script, shell, import-boundary and protocol checks (outer exit 0).

All run IDs above are below `.cartulary/test-results/`. No historical full CI
pass is assumed, and no warm baseline was created solely for finalizer upkeep.

### Interrupted completion gate and focused repairs

`CARTULARY_TEST_RUN_ID=cutover-ci-six-01 make ci BROWSER_LANES=6` was interrupted
at the user's request to stop early and repair its failures. The native process
exited 130 after finalization. Its terminal summary records 1,050 passed, two
failed and 51 cancelled work units out of 1,103. This is incomplete qualification,
not a CI pass. The run binds source digest
`sha256:5e61580830de08e628b27f44142d1aa8b79ce0ca11ddcffebf7a32c4f102ded3`.

The retained `cleanup-results.json` records 1,633 completed cleanup operations
and no failed operation. `retained-secret-scan.json` passed over 10,464 files.
These results describe that invocation; they do not qualify later repairs.
Six lanes were requested with unchanged host capacities. Published observations
showed two browser groups running together; they do not establish a global peak
or six-way physical throughput.

The two failing units exposed three concrete gaps:

- `operation:harness.public_wrapper_smoke` failed because the v5 task-guide
  schema still restricted `explicit_commands` to owner-slice syntax. The schema
  now accepts only the two dedicated commands. Routing tests validate guides
  with Visual and UI-Review commands as well as guides without either command.
- The owned case of `harness.browser.integration.native_fixture_lifecycle`
  omitted its resolved renderer lane limit. Its explicit single-slot fixture
  now declares one lane, preserving its repeated-borrow reuse assertions.
- The borrowed case timed out starting an independently reserved persistent
  service owner. Its enclosing installed-browser profile also claimed physical
  service capacity. The borrowed case is now a separately admitted row,
  `harness.browser.integration.native_borrowed_fixture_lifecycle`, with the
  control-process `harness_contract` profile. The persistent owner retains its
  independent resident reservation. Owned and failed-cleanup cases retain their
  installed-browser envelope. All assertions remain; no capacity or deadline
  was raised. Overlapping claims are confirmed by static inspection; their
  contribution to the original timeout is an inference, not a measured cause.

The resource contract in REQ-803 now distinguishes these ownership models.
A compiler regression verifies that the borrowed row requests control resources
without duplicate service, browser or port claims. This change belongs in the
specification, catalog, implementation fixtures and tests; it removes accidental
coupling between persistent-owner admission and browser-lifecycle admission.
It adds an exact selectable row without a compatibility alias. Leaving the
overlap would retain avoidable capacity pressure and timeout risk.

`cutover-fix-generate-01` failed because the new catalog row was not ASCII-sorted.
After sorting the authored manifest, `cutover-fix-generate-02` passed generation
(outer exit 0). `cutover-fix-format-01` passed formatting and
`cutover-fix-routing-01` passed the exact routing-contract selection (outer exit 0).
The interrupted full CI invocation is not automatically restarted; repair
validation uses the affected narrow selections.

The following focused repair commands passed with outer exit 0. Their run roots
are `/home/jochi/code/cartulary/.cartulary/test-results/<run-id>`:

| Run ID | Command | Outcome |
| --- | --- | --- |
| `cutover-fix-routing-01` | `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | Routing and guide schemas pass, including dedicated optional commands |
| `cutover-fix-smoke-01` | `make harness-contract SUITE=smoke` | Public Make wrapper, scheduler and fixture-broker smoke checks pass |
| `cutover-fix-lifecycle-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.native_borrowed_fixture_lifecycle,harness.browser.integration.native_fixture_lifecycle,harness.browser.unit.browser_lane_policy` | All four graph units pass; real owned/borrowed lifecycle cases and synthetic lane policy pass |

The lifecycle run binds source digest
`sha256:33ad80009fb264b9384527457167ed7f169e9096b52eb7b80348db80a9052a94`.
Its default resolved lane limit is four; the physical reuse test explicitly owns
one slot and verifies compatible generation reuse, fresh cookies/local storage,
and retirement after client interruption. Its persistent-service case verifies
borrower protection and owner survival; the failed-cleanup case verifies exact
recovery. Its 13 cleanup operations completed, and its retained-secret scan
passed over 113 files. The synthetic lane checks cover four/six configuration
and exclude duplicate physical claims from the borrowed-service test. These
passes do not establish six-way physical throughput or resolve the incomplete
full-suite qualification.

`cutover-fix-finalize-01` passed `make agent-finalize` (outer exit 0), with zero
generated changes. Its schema, catalog and tier validations bind the same
`sha256:33ad80009fb264b9384527457167ed7f169e9096b52eb7b80348db80a9052a94`
source digest. Retained-run evidence and scheduler maintenance were skipped
because `RESULTS_DIR` was unset. This is the repair implementation identity
before the capacity follow-up below, on `808203eaf` plus the uncommitted cutover;
no new commit has been created.

`cutover-fix-schema-01`, `cutover-fix-generated-01` and
`cutover-fix-static-01` passed `make source-check CHECK=schema`,
`make source-check CHECK=generated` and `make lint CHECK=standard`, respectively
(outer exit 0), before the capacity follow-up.

### Remaining acceptance scope

The full `make ci BROWSER_LANES=6` completion gate remains incomplete after the
requested early stop. `make browser-e2e SUITE=visual` and
`make harness-contract SUITE=ui-review` have not been executed for this cutover.
Their selection contracts are verified, but no Visual or UI-Review pass is
claimed. A later completion effort must use the final source identity and fresh
run roots; the interrupted run cannot be relabelled or combined into a pass.
`release-check` and release measurements were intentionally not run for this
harness refactor. No full warm run was created for retained-run maintenance.

### User-authorized capacity increase

The subsequent request to increase resource limits supersedes the initial
instruction to retain the old capacities. Inspection of the already retained CI
snapshot found 19 CPU tokens, 38 I/O tokens, 76 process slots and approximately
31 GiB of memory, but only eight PostgreSQL tokens, three object-store tokens
and 16 port lanes. The shared object-store owner reserves one token; three
tokens therefore permit only two ordinary functional groups requiring one
additional token each. Six groups plus shared infrastructure require 13
PostgreSQL tokens, seven object-store tokens and 20 port lanes under the current
authored reservation model. This is static accounting, not a performance
measurement or a claim about actual utilization.

REQ-353 and the resource registry now set default ceilings of 16 PostgreSQL
tokens, eight object-store tokens and 32 port lanes. PostgreSQL and object-store
ceilings remain bounded by admitted CPU tokens; port lanes are bounded by twice
that number. Smaller hosts scale down. The single registry owns these values;
the semantic reader no longer independently pins PostgreSQL to eight. Capacity
resolution reuses its existing pure automatic-policy calculation rather than
duplicating the fixed-cap calculations.

CPU, I/O, memory and process limits already accommodate the six-group envelope
on this host and remain unchanged. The 25 percent CPU safety margin, cleanup
headroom, resident accounting, exclusive measurement priority, one worker per
group and four-default/six-maximum browser policy remain in force. This change
raises admission ceilings, not production database connection settings or
physical resource claims. It requires no migration or compatibility alias.

Synthetic tests check the actual compiled six-lane graph's 8,960 MiB, 52-process,
20-port, 13-PostgreSQL and seven-object-store envelope against a 24-CPU/32-GiB
host policy, including additional cleanup capacity. They also check bounded
defaults on four-CPU and one-CPU hosts, and rejection of a port override above
the host-derived bound. Actual simultaneous throughput remains unmeasured.

Focused validation passed with outer exit 0:

- `cutover-capacity-generate-01`: `make generate` refreshed authored projections.
- `cutover-capacity-lanes-01`: `make test-slice OWNER=harness.browser ROWS=harness.browser.unit.browser_lane_policy,harness.browser.unit.native_fixtures`
  passed the six-group envelope, lower-host bounds, native slot, contention,
  cancellation, ownership and cleanup contracts.
- `cutover-capacity-scheduler-01`: `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.scheduler_contract`
  passed capacity override and scheduler contracts. Its retained snapshot
  resolves 19 CPU, 38 I/O, 76 process, 16 PostgreSQL, eight object-store and 32
  port tokens, with 33,353,125,888 memory bytes.

Both focused test invocations bind
`sha256:3de6e779aa7a8a838f63704342c744f93ad2a911ad3f2fd8659f0b9509b3a22a`.
No comprehensive CI run was restarted and no new benchmark was performed.
Earlier physical lifecycle passes remain evidence of those earlier source
identities; the increased limits have focused contract qualification, not a
full current-source CI or six-way physical-throughput qualification.

`cutover-capacity-finalize-01` passed `make agent-finalize` (outer exit 0),
reporting zero generated changes. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset. The current final implementation digest is the
`sha256:3de6e779aa7a8a838f63704342c744f93ad2a911ad3f2fd8659f0b9509b3a22a`
identity above, on revision `808203eaf` plus uncommitted changes.

`cutover-capacity-schema-01`, `cutover-capacity-generated-01` and
`cutover-capacity-static-01` passed `make source-check CHECK=schema`,
`make source-check CHECK=generated` and `make lint CHECK=standard`, respectively
(outer exit 0). All run roots remain under
`/home/jochi/code/cartulary/.cartulary/test-results/`.
