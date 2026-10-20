# Testing harness remediation 02

## Scope and baseline

Implementation of the four-blocker remediation plan approved on 2026-10-20.
Starting HEAD: `766846c72fe50328c83e301488ea5394a70dacd2`; working tree clean.
Historical CI: `remediation01-pg07-readiness-ci`, source
`sha256:c4b09f3c40723faeeec6a112370edfbdf8b578db1dd439d8567299fc533a4e5a`.
Historical results do not establish acceptance of the completed source.

The selected SQL readiness contract is 15 seconds, 250 ms polling, and a
5-second maximum individual probe, bounded by the parent deadline. Database
drop, cleanup, durability and ownership constraints remain unchanged.

## Current obligation matrix

| ID / status | Owner and boundary | Required implementation and regression evidence | Exit |
| --- | --- | --- | --- |
| W0 IMPLEMENTED | Testing Harness NLSpec; Core 03 contextual authoring | Contracts and authored projections amended; source/schema/catalog checks pass in final CI | Specification fidelity is ready for human review; Markdown remains outside executable inputs |
| W1 INVESTIGATION OPEN | harness.browser; pgtest/postgrescleanup | Bounded observations implemented; isolated and aggregate execution pass; historical storage cause unproved | Current sampled drops meet their deadlines; a causal workload, accounting or environment repair remains contingent on evidence |
| W2 IMPLEMENTED; HISTORICAL CAUSE OPEN | harness.browser; PostgreSQL provider and fixture broker | Partial ownership, explicit settlement, readiness stages and failure references implemented; final-source Graph Projection row launches and passes | Valid exclusive binding, actual assertions and completed cleanup demonstrated; old allocation-to-service association cannot be reconstructed |
| W3 VERIFIED | module.workbook; contextual reference picker tests | Controlled late-response regression, eight-row recovery selection and all 55 aggregate accessibility rows pass | Exact identities, cancellation and focus preserved; all seven previously skipped rows execute |
| W4 VERIFIED | harness.browser; native session/process ownership | Closed identity diagnostics, durable producer retirement, PID-reuse and descendant regressions pass; exact retained runtime recovered before cutover | Original proof and failure receipts preserved; final CI runtime closes successfully |
| V00 CURRENT EXECUTION VERIFIED; CAUSAL ACCEPTANCE OPEN | Harness and affected source owners | Finalization, final-source CI, required explicit selections, cleanup and evidence checks pass; documentation verification identified below | Overall remediation remains open for the W1/W2 causal investigations; passing execution is not a waiver |

## Sequencing and compatibility

First adopt contracts and bounded diagnostics. Diagnose and recover the old
private-detail authority while its format remains supported. Then cut over
producer lifecycle writers and readers together, without a permanent legacy
reader. Failed or ambiguous proof blocks the affected cutover.

Readiness, storage and identity hypotheses remain separate. No timeout increase,
automatic destructive retry, weakened durability, speculative relocation or
second resource owner is admitted. Resource-profile changes require causal
evidence. Service-heavy comparisons and aggregate validation run sequentially.

## Evidence log

- Discovery: `CARTULARY_TEST_RUN_ID=remediation02-discovery make help ALL=1`
  exited 0. Public recovery requires exact `RESULTS_DIR`, `RUNTIME_ROOT` and
  `RUNTIME_LEASE_ID`.
- A bounded read-only observation of the exact retained private authority found
  440 non-group producer identities absent and 750 groups absent (1190 total).
  This does not explain the historical failing comparison and is not deletion
  authority. At this observation the runtime remained fenced; the successful
  owner recovery below subsequently settled it.

## Handoff and acceptance

The implementation has passed fresh aggregate CI and required explicit checks.
The evidence handoff retains the unresolved causal obligations. Preserve original CI and recovery
receipts. Record native exits, exact run/source identities, row execution and
fixture/runtime settlement separately. Passing current execution does not prove
the historical storage, readiness or process-identity cause. Retained-run
maintenance requires a compatible successful full warm result root.

## Implementation evidence, 2026-10-20

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

## Intermediate comparison and audit results

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

## Completed-source baseline and aggregate acceptance

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

## PostgreSQL comparison and evidence limits

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

## Operating and migration guidance

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

## Unresolved causal obligations

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

## Explicit validation, phases and final handoff

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
