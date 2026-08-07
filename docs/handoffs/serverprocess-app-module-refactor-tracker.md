# serverprocess-app Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

### 1.1 Status and authority

| Item | Required posture |
| --- | --- |
| Target path | `internal/app/serverprocess` |
| Target label | `serverprocess-app`, normalized to lowercase kebab case |
| Output path | `docs/handoffs/serverprocess-app-module-refactor-tracker.md` |
| Current task | Planning and documentation only; this tracker is the only permitted source change. |
| Later work | Production, test, contract, harness, topology, generated, or configuration changes require a separately authorized implementation task. |
| Repository baseline | `main` at `04af1d2c77e817be3bcebfa23829fcd3d2a6891a`; the tracker is an untracked workspace file carried forward from the preceding planning session. |
| Prior history | The 2026-08-06T22:20:53-04:00 handoff rows are retained. This revision appends the 2026-08-06T23:00:58-04:00 session. |

This tracker is a prescriptive implementation plan subordinate to adopted
owners. It is not an adopted product NLSpec, a production contract, an
executable harness input, or proof of repository conformance. Requirements in
this tracker constrain only a later authorized refactor. When this tracker
conflicts with an adopted owner, the owner controls and implementation MUST
stop with `BLOCKED: owner contradiction`.

Normative terms have these meanings:

- **MUST** and **MUST NOT** define binary requirements for the later refactor.
- **SHOULD** defines a requirement that may be waived only by recording the
  owner-backed reason and replacement evidence in Section 10.
- **MAY** grants intentional implementation freedom.
- Unspecified internal Go names and private decomposition are intentional
  implementation freedom when they do not alter a requirement below.

The source hierarchy is:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. This tracker, the planning framework, research, and analysis notes as
   planning evidence only.

Core 05 is not applicable. The affected rows carry
`claim_posture="implementation"`; no timed or fixture-sensitive claim is
published. Test-family rows, execution topology, result roots, and phase maps
remain evidence routing or execution support and MUST NOT define runtime
architecture.

### 1.2 Tracker requirements

| Requirement | Normative rule |
| --- | --- |
| SPT-REQ-001 | The refactor MUST follow the source hierarchy above, MUST NOT edit Core 00 through Core 04 for the decisions closed here, and MUST stop on an actual owner contradiction. |
| SPT-REQ-002 | `internal/app/serverprocess` MUST remain process-level test evidence. It MUST NOT become a production module, application facade, or behavioral owner merely because the directory exists. |
| SPT-REQ-003 | HTTP, WebSocket, authorization, storage, revision, projection, Recovery, generated-contract, harness-security, and packaged-process behavior MUST remain observable-equivalent unless a separate owner-backed behavior change is authorized. |
| SPT-REQ-004 | The Network Flow process-composition selector MUST migrate to the exact `module.networkflow` row specified in Section 3 and MUST route only through owner-local Network Flow verification. |
| SPT-REQ-005 | The restored-server selector MUST migrate to the exact `module.recovery` row specified in Section 3 and MUST remain the unique packaged-server proof of the AC-399-class result. |
| SPT-REQ-006 | Owner migration MUST allocate the exact new immutable row IDs and runtime-binary mappings in Section 3 while preserving the live runtime, resource, fixture, tier, and claim-posture values. |
| SPT-REQ-007 | Each migrated selector MUST have exactly one active row. The old row MUST be removed, the temporary crosswalk MUST remain non-executable, and the crosswalk MUST be removed after final reconciliation. |
| SPT-REQ-008 | The Recovery sentinel MUST prove the Section 4 AC-399 matrix and MUST NOT claim AC-401 cadence, due-selection, basis-change scheduling, or attestation-update behavior. |
| SPT-REQ-009 | A helper MUST remain local by default and MAY move only to the destination admitted by the Section 5 boundary-classification table. |
| SPT-REQ-010 | Every moved helper MUST have a complete caller-by-caller semantic-equivalence matrix and MUST satisfy every shared-helper admission rule in Section 5. |
| SPT-REQ-011 | `processtest` process identity, defaults, deadlines, listener behavior, termination, diagnostics, and resource-release observations MUST remain as specified in Section 4. |
| SPT-REQ-012 | Recovery-owned test support MUST expose the semantic output contract in Section 5, MUST preserve evidence safety and resource ownership, and MUST NOT start the final server sentinel. |
| SPT-REQ-013 | Authored catalog and topology changes MUST be generated and validated through Make, MUST preserve rollback atomicity, and MUST retain exact before/after result roots. Generated outputs MUST NOT be hand-edited. |
| SPT-REQ-014 | Every completed slice MUST update this tracker, record actual commands and outcomes, preserve prior handoff history, and leave no unexplained out-of-scope diff. |

Owner and support documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/domain.md`
- Core 00 through Core 04 under `docs/spec/`
- `docs/testing-harness-nlspec.md`
- `docs/network-flow-activity-nlspec.md`
- `docs/extension-subsystem-nlspec.md`
- `docs/research/nlspec-spec.md` as writing guidance, not owner authority
- `temp/analysis-notes.md` as decision input, not repository-state proof

Repository evidence inspected includes every path in Section 2; the server
facade, assembly, and production/harness profiles; `processtest` and
`appsupport`; relevant view-schema, WebSocket, Recovery, and Network Flow
contracts; verification owners; test-family manifests; the catalog owner;
execution topology; schemas; and backend boundary policy.

## 2. Current-State Repository Inventory

The target contains ten tracked Go test files and no production Go source. The
fifteen additional JSON files are ignored Recovery execution evidence, not
authored requirements.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/app/serverprocess/config_test.go` | Proves missing, incompatible, overlapping, or invalid configuration fails before readiness and leaves HTTP, readiness, and socket probes unavailable with structured diagnostics. | Test functions only. | `app.server.process` rows. | Platform config plus config, fixture, and process test support. | Self and package harness. | Configuration and diagnostic contracts indirectly. | `app.server` with `platform.config` collaboration. | High | A refusal-only probe uses an obsolete view-specific socket literal. |
| `internal/app/serverprocess/e2e_test.go` | Exercises real-process bootstrap, readiness, storage checks, database observations, environment construction, and shared process helpers. | Package-test environment, SQL, S3, readiness, and HTTP helpers. | Other target tests and app.server process rows. | Auth bootstrap, object store, audit/config/fixture/process/S3/security test support, SQL, and HTTP. | Self plus package callers. | Bootstrap, readiness, storage, diagnostics, and HTTP envelopes indirectly. | `app.server`; reusable mechanics only where Section 5 admits them. | Critical | Process evidence MUST remain subprocess-backed. |
| `internal/app/serverprocess/embedded_frontend_process_test.go` | Proves the packaged server serves the root HTML shell and one embedded asset. | Test function only. | One full-tier app.server row. | Standard HTTP and response parsing. | Self. | Embedded web assets indirectly. | `app.server`. | Medium | No frontend controller or grid adapter is present. |
| `internal/app/serverprocess/evidence_process_test.go` | Creates Evidence and Timeline records, uploads an object blob, attaches Evidence, and checks Timeline projections. | `CreateViewRow`, `PatchRecord`, `RequireTimelineEvidenceCount`, and `PutObject`; test-only. | Recovery sentinel and its own scenario. | Auth primitives, HTTP test utilities, and process support. | Self and Recovery sentinel. | Evidence and Timeline view schemas, record links, and projection fields indirectly. | Evidence, Timeline, and Projections at the app.server process edge. | Critical | These helpers remain owner/process-local by default. |
| `internal/app/serverprocess/incident_membership_process_test.go` | Exercises incidents, memberships, default workbook preferences, deployment-admin non-bypass, and extension discovery/reserved routes. | `CreateIncident` and extension-claim assertions; test-only. | Same-package scenarios and app.server rows. | Auth primitives, HTTP utilities, and process support. | Self plus Evidence and Recovery fixtures. | Incident, membership, workbook, and extension contracts indirectly. | Incidents with Auth, Workbook, Extensions, and app.server collaborators. | Critical | Physical location does not assign product ownership. |
| `internal/app/serverprocess/networkflow_runtime_routes_process_test.go` | Proves the server-harness binary composes Network Flow fault control, rejects malformed requests, arms a fault, and resets it. | Test function only. | One app.server full-tier row. | Network Flow harness control and HTTP test utilities. | Self plus module harness-control coverage. | Harness route accounting only. | `module.networkflow`, with app.server collaborator. | Critical | Current owner routing violates TH-HARNESS-REQ-658. |
| `internal/app/serverprocess/process_test.go` | Covers bootstrap, MFA, sessions, CSRF, authorization, session limits/revocation, admin lifecycle, and shared HTTP/socket helpers. | Package-test login, request, response, and socket helpers. | Same-package scenarios and app.server rows. | Auth flow support, auth primitives, fixtures, HTTP/process support, SQL, and WebSocket client. | Self plus package callers. | Auth HTTP envelopes and incident WebSocket semantics indirectly. | `module.auth` at the app.server process edge. | Critical | Direct SQL is test observation, not production persistence logic. |
| `internal/app/serverprocess/recovery_sentinel_test.go` | Captures retained backups, restores and verifies a fresh target, emits evidence, starts the actual server-harness binary, and queries restored state. | Recovery fixture types and helpers; test-only. | One app.server row and same-package helpers. | Extension, Recovery, Timeline, Evidence recovery provider, object store, auth, appsupport, fixtures, HTTP, and processtest. | Self plus reused incident/Evidence/Auth helpers. | Recovery, Timeline, Evidence, and view contracts indirectly. | `module.recovery`, with app.server, Evidence, and Timeline collaborators. | Critical | The explicit Evidence recovery-provider boundary allowance remains until an actual import move. |
| `internal/app/serverprocess/runtime_routes_process_test.go` | Proves exact harness enablement, token/host/origin/CORS guards, test clock, and reset behavior. | Test functions and request helper. | App.server process rows and Network Flow reset scenario. | Fixtures, HTTP/process support, networking, and JSON. | Self and Network Flow process scenario. | Testing Harness security and reset accounting indirectly. | Harness/platform mechanics at the app.server process edge. | Critical | Production profile isolation MUST remain fail-closed. |
| `internal/app/serverprocess/shared_process_harness_test.go` | Owns package `TestMain` and shared Postgres/S3 lifecycle. | Package lifecycle only. | Every target test. | `pgtest`, `s3test`, environment, and test lifecycle. | Entire package. | Execution topology indirectly. | Package-local test infrastructure. | High | Section 5 requires it to remain local. |
| `internal/app/serverprocess/.cartulary/test-results/local-backup-restore/{backend-process,test-slice}/backup-restore/{object-store-backup-manifest.json,object-store-backup-summary.json,restore-verification.json}` | Six ignored local Recovery artifacts. | None. | Local executions only. | Recovery artifact writer. | Not source tests. | Runtime instances of Recovery schemas. | Out of scope. | Low | MUST NOT be edited, deleted, hashed as source, or treated as authority. |
| `internal/app/serverprocess/tmp/op/{r1335c4-check,r925b0-check,re9a32-check}/backend-process/backup-restore/{object-store-backup-manifest.json,object-store-backup-summary.json,restore-verification.json}` | Nine ignored prior-run Recovery artifacts. | None. | Local executions only. | Recovery artifact writer. | Not source tests. | Runtime instances of Recovery schemas. | Out of scope. | Low | Same exclusion as the local artifacts above. |

## 3. Module Boundary Diagnosis

The target is a legitimate process-boundary evidence package with mixed owner
scenarios. It is not a production facade, persistence adapter, frontend
controller, saved-view implementation, imports/ingest module, or grid-vendor
integration layer.

| Responsibility | Required owner/boundary | Disposition | Normative evidence | Required result |
| --- | --- | --- | --- | --- |
| Packaged server startup, readiness, refusal, diagnostics, and shutdown | `app.server` process evidence | keep | Repository procedure and process tests | Only subprocess tests may close these postconditions. |
| Network Flow fault-control composition | `module.networkflow`; app.server collaborator | re-own row, keep selector in place | TH-HARNESS-REQ-658, 664, and 667 | Product-owner verification receives the process evidence without moving the test. |
| Coherent restored-server sentinel | `module.recovery`; app.server, Evidence, and Timeline collaborators | re-own row; split helpers later | Core 01 REQ-01-577/578 and Core 04 AC-399 | Recovery owns the aggregate result while actual server proof remains in this package. |
| Generic child-process mechanics | `internal/testutil/processtest` | keep or extract only equivalent mechanics | Existing process helper interface | No in-process replacement is permitted. |
| Pure in-process application composition | `internal/testutil/appsupport` | admit only pure helpers | Repository procedure | No process, fixture, environment, or Recovery semantics enter appsupport. |
| Recovery semantics | `internal/modules/recovery/testsupport` | create only in an authorized helper slice | Recovery owner boundary | Final server startup remains outside Recovery support. |
| Auth, incident, Evidence, Timeline, and extension scenarios | Named behavioral owners; app.server collaborator when process composition is material | keep physically unless a later selector split is justified | Core and subsystem owners | Package location remains implementation support only. |
| Shared containers and direct SQL observations | Package-local evidence | keep | Live callers and fixture lifecycle | No generic production-like facade is introduced. |

### 3.1 Exact catalog migration map

The live harness requires every verification ID to be owned by the row's
primary owner. Cross-owner secondary verification IDs are invalid. Other
participants MUST appear only in sorted, unique `collaborator_ids` unless a
separate non-overlapping owner row independently proves their postcondition.

| Field | Network Flow replacement | Recovery replacement |
| --- | --- | --- |
| Retired row | `app.server.process.the_standalone_server_composes_the_network_flow_3d2a5564ba` | `app.server.process.restoring_the_latest_successful_retained_backup_6456371896` |
| New row | `module.networkflow.process.the_packaged_standalone_server_composes_the_netw_400a31ad27` | `module.recovery.process.the_packaged_server_serves_the_coherently_restor_6b1731590e` |
| `owner_id` | `module.networkflow` | `module.recovery` |
| `family_id` | `module.networkflow.process` | `module.recovery.process` |
| `collaborator_ids` | `["app.server"]` | `["app.server", "module.evidence", "module.timeline"]` |
| `verification_ids` | `["module.networkflow.verification.behavior_contract"]` | `["module.recovery.verification.behavior_contract"]` |
| Go selector | `./internal/app/serverprocess` + `TestNetworkFlowHarnessRuntimeRouteServerProcessContribution` | `./internal/app/serverprocess` + `TestFreshEnvironmentRestoreWorkbookConsistency_Integration` |
| Runner/evidence | `go` / `integration` | `go` / `integration` |
| Runtime/resource | `default` / `standard` | `default` / `standard` |
| Fixture | `postgres_dedicated` | `postgres_dedicated` |
| Tier/posture/status | `full` / `implementation` / `active` | `standard` / `implementation` / `active` |
| Runtime binaries | New topology entry `module.networkflow.process → ["server-harness"]` | Existing topology entry becomes `module.recovery.process → ["operator", "server-harness"]` |

The IDs were derived through the public authoring command with these exact
inputs:

```text
make author-test-row-id FAMILY_ID=module.networkflow.process CLAIM='the packaged standalone server composes the Network Flow harness fault-control contribution' SELECTOR_KEY='go:./internal/app/serverprocess:TestNetworkFlowHarnessRuntimeRouteServerProcessContribution'
make author-test-row-id FAMILY_ID=module.recovery.process CLAIM='the packaged server serves the coherently restored latest successful retained backup' SELECTOR_KEY='go:./internal/app/serverprocess:TestFreshEnvironmentRestoreWorkbookConsistency_Integration'
```

The Recovery process family already supplies the operator binary to its
operator-process selectors. The required sorted binary union is intentional;
the refactor MUST NOT invent a row-level binary override or new topology schema
to avoid that shared prerequisite.

### 3.2 Temporary migration crosswalk

This table is documentation-only implementation evidence. It MUST NOT be read
by harness code, used as a selector, or retained after final reconciliation.

| Old immutable row ID | New immutable row ID | Current disposition |
| --- | --- | --- |
| `app.server.process.the_standalone_server_composes_the_network_flow_3d2a5564ba` | `module.networkflow.process.the_packaged_standalone_server_composes_the_netw_400a31ad27` | Planned; old row remains live until an authorized atomic catalog migration. |
| `app.server.process.restoring_the_latest_successful_retained_backup_6456371896` | `module.recovery.process.the_packaged_server_serves_the_coherently_restor_6b1731590e` | Planned; old row remains live until an authorized atomic catalog migration. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Existing evidence | Required future characterization | Risk | Frozen rule |
| --- | --- | --- | --- | --- | --- |
| Process startup, listener, readiness, exit, and diagnostics | Core 01/Core 04; app.server | `config_test.go`, `e2e_test.go`, processtest | Preserve every existing process row. | Critical | Startup failure remains pre-listener and diagnostics remain structured and secret-safe. |
| HTTP routes and envelopes | Core 01 and route owners | Most target tests | Preserve status, envelope, and error-code assertions. | Critical | A test/catalog move MUST NOT redefine a route. |
| Incident WebSocket | Core 01/Core 04 | `process_test.go` and refusal probes | Live socket evidence uses `/ws/v1/incidents/{incident_id}`. | Critical | The obsolete view-specific refusal literal may change only with unchanged refusal semantics. |
| Auth, MFA, CSRF, sessions, and authorization | Core 04; Auth and Incidents | `process_test.go`, membership scenarios | Preserve all public outcomes and revocation behavior. | Critical | Deployment admin remains distinct from incident authority. |
| Evidence/blob attachment and Timeline projections | Evidence, Timeline, Links, Projections | `evidence_process_test.go`, Recovery sentinel | Preserve upload, attach, collection link, count, and boolean projection results. | Critical | Evidence record and object blob remain distinct state. |
| Recovery and workbook probe | Recovery, Timeline, Workbook, Evidence | Recovery sentinel plus owner process/browser/integration rows | Close the AC-399 matrix below. | Critical | No public backup/restore route and no AC-401 overclaim. |
| Harness route security and reset | Testing Harness/platform | `runtime_routes_process_test.go` | Preserve exact `1`, token, host/origin/CORS, clock, and reset outcomes. | Critical | Harness controls remain non-product mechanics. |
| Network Flow harness control | Network Flow plus Testing Harness | Target process test and module harness-control tests | Preserve malformed, arm, reset, and rearm outcomes. | Critical | Fault vocabularies and routes are not product APIs. |
| Embedded browser assets | app.server packaging | Embedded asset smoke | Preserve only if packaging changes. | Medium | No frontend/controller/grid refactor is in scope. |
| Generated and harness projections | Adopted owners and authored inputs | Catalog/topology/drift tooling | Regenerate only from authored inputs. | High | Generated roots MUST NOT be hand-edited. |

### 4.1 Recovery AC-399 assertion matrix

The Recovery browser row starts `go run ./tools/recoverybrowserrestore`, whose
helper composes `server.Runtime` in-process. Recovery operator rows execute the
operator binary. Neither row starts the packaged server-harness binary against
the restored target. The target sentinel is therefore unique and MUST remain.

| AC-399-class postcondition | Current sentinel evidence | Required authorized-slice action | Acceptance |
| --- | --- | --- | --- |
| Latest successful retained backup selected | Explicit selected `backup_set_id` equality | retain | Exact expected UUID is selected. |
| Postgres and object state come from the same backup | Restore runner plus manifest consistency point; hashes are only checked non-empty | add explicit selected backup/consistency-point and source-to-restored digest equality | Structured and object identities bind the same selected set. |
| Projections rebuild before serving | Restore rebuilder, readiness gate, and final Timeline query | retain and assert readiness is marked only after rebuild | Query cannot pass before rebuild closure. |
| Incident identity preserved | Public GET uses the source incident ID | retain | Restored response returns the exact incident ID. |
| Record identity preserved | Timeline query searches the source record ID | retain | Exact record ID is present. |
| Row version preserved | Not asserted explicitly | add exact source/restored row-version equality | Positive expected version is unchanged. |
| Change-set count preserved | Change-set digest is checked non-empty | add exact source/restored count and digest equality | Count and digest match. |
| Blob hashes preserved | Manifest member hashes exist; restored digest is checked non-empty | add exact source/restored blob digest and durable blob hash equality | Lowercase SHA-256 values match. |
| Evidence/blob lifecycle remains coherent | Attachment count/boolean survive | add exact Evidence record, blob identity, lifecycle, and link checks | Lifecycle and attachment semantics match source. |
| Built-in workbook probe executes | Verification result passes, but registration/view binding is not asserted | decode the artifact and assert `timeline.base_restore_probe.v1`, `cartulary.view.timeline.v2`, and selected incident | Owner-defined probe is authoritative. |
| Actual packaged server serves restored target | `processtest.StartServer`, readiness, auth, incident GET, and query | retain | Server-harness reaches readiness and public query succeeds. |
| Shutdown completes | Deferred `server.Stop` | retain | Process group terminates within the frozen bound. |
| AC-401 cadence and due-verification | Not exercised | exclude from this row | No cadence, basis-change, due-selection, or attestation claim is attached. |

### 4.2 `processtest` interface and defaults

| Interface point | Required preserved behavior |
| --- | --- |
| Binary input | `CARTULARY_SERVER_HARNESS_BIN` is required and MUST identify a regular executable, non-symlink file. Omission fails before child start. |
| Environment | `ServerOptions.Env` is copied before listener variables are added; the caller retains ownership of its input map. |
| Listen default | Omitted `CARTULARY_HTTP_ADDR` means `127.0.0.1:0`. The helper passes the inherited listener as FD `3`. |
| Runtime identity | The helper MUST use an actual child process and MUST NOT call server constructors or runtime assembly in-process. |
| Readiness | Deadline `15s`; per-request timeout `500ms`; poll interval `200ms`; both `/healthz` and `/readyz` must return `200`. |
| Exit wait | Deadline `15s`; timeout fails with captured stdout/stderr. |
| Ordinary status probe | Request timeout `2s`. |
| Stop | Cancel context, signal the child process group with `SIGTERM`, and wait at most `5s`; timeout fails with captured output. |
| Diagnostics | Structured stderr remains observable; Go-run suffix normalization does not authorize arbitrary log parsing. |

## 5. Coupling and Boundary Findings

### 5.1 Findings

| Finding | Evidence | Risk | Classification | Required action |
| --- | --- | --- | --- | --- |
| Network Flow row is app.server-owned and app.server-verified. | Exact live row and TH-HARNESS-REQ-658/667. | Owner verification bypass. | `must_fix`; decision resolved | Apply S-03 exact mapping. |
| Recovery sentinel is app.server-owned. | Exact live row; Core 01/04 postcondition. | Recovery evidence misownership. | `must_fix`; decision resolved | Apply S-04 exact mapping. |
| Analysis notes proposed cross-owner secondary verifications. | Live validator requires each verification owner to equal row owner. | Invalid catalog. | `intentional/no_action` after correction | Use collaborators only, as fixed in Section 3. |
| Process-family binary requirements follow family identity. | Execution topology maps `app.server.process` to server-harness and `module.recovery.process` to operator. | Re-owned tests could lose required binaries. | `must_fix`; decision resolved | Apply the exact binary map in Section 3. |
| Some helpers have similar names across test packages. | `CreateIncident`, platform open helpers, and process environment builders have different runtime contexts. | Evidence weakening through false reuse. | `should_fix` by admission, not forced extraction | Default local; require Section 5.3 matrix. |
| Recovery sentinel contains reusable Recovery semantics. | Capture, selection, restore, verification, artifact, and target helpers are co-located with final process boot. | Second Recovery implementation at process edge. | `should_fix`; decision resolved | Move admitted Recovery semantics to Recovery testsupport after characterization. |
| Evidence recovery-provider import is allowlisted. | Backend boundary policy. | Stale policy after movement. | `intentional/no_action` | Change the allowance only if the import actually moves. |
| Direct SQL is test observation only. | Readiness/session fixture observations. | Accidental shared persistence facade. | `intentional/no_action` | Keep local unless an equally strong public observation replaces it. |
| Shared `TestMain` owns package fixtures. | Package-wide Postgres/S3 lifecycle. | Isolation and fixture-capability drift. | `intentional/no_action` | Keep local in this refactor. |
| No frontend/controller/grid/import implementation exists. | File and import inventory. | False scope. | `intentional/no_action` | No corresponding workstream. |

No owner contradiction exists. The two catalog defects are implementation
projection mismatches with resolved owner decisions.

### 5.2 Helper destination and default map

The default disposition is **local**. Textual similarity, shared names, or one
potential caller MUST NOT establish reuse.

| Helper class or named family | Boundary class | Required destination | Decision/default |
| --- | --- | --- | --- |
| Process start, inherited listener, readiness, exit, shutdown, and output capture | `subprocess` | `internal/testutil/processtest` | Existing canonical home; move only duplicate owner-neutral mechanics. |
| Pure configuration values and typed platform-settings builders | `in_process` or `transport_only` | `internal/testutil/appsupport` | Admit only with explicit inputs and no process, environment, fixture, or Recovery ownership. |
| `ServerEnv` and process-specific environment assembly | `subprocess` | local | Keep local. |
| `CreateIncident` | `owner_semantic` plus subprocess auth | local | Keep local unless a raw transport-only subset passes the matrix. |
| `CreateViewRow`, `PatchRecord`, `PutObject`, `RequireTimelineEvidenceCount` | `owner_semantic` | local Evidence/Timeline process support | Keep local. |
| Login, cookie, MFA, CSRF, session, WebSocket, refusal, and diagnostic choreography | `owner_semantic` or `subprocess` | local | Keep local. |
| Direct SQL observation | `owner_semantic` test observation | local | Keep local unless replaced by equally strong public evidence. |
| Shared Postgres/S3 `TestMain` | `fixture_lifecycle` | local | Keep local for this refactor. |
| Backup capture/selection, restore target, catalogs, codecs, markers, leases, rebuild, verification artifacts, and consistency checks | `owner_semantic` | `internal/modules/recovery/testsupport` | Move only in S-05 after the matrix and AC-399 characterization pass. |
| Final restored-server start, readiness, auth, query, and shutdown | `subprocess` | local `serverprocess` sentinel | MUST remain local. |

### 5.3 Mandatory semantic-equivalence matrix

Before a helper moves, the implementation handoff MUST contain one row per
candidate and caller. No field may be omitted.

| Field | Required characterization |
| --- | --- |
| Current definitions | Every implementation alleged to duplicate the helper. |
| Callers | Every direct and indirect caller. |
| Boundary class | Exactly one of `subprocess`, `in_process`, `transport_only`, `fixture_lifecycle`, or `owner_semantic`. |
| Inputs/defaults | Required, optional, omitted, derived, and environment-bound inputs. |
| Runtime identity | Actual server binary, operator binary, in-process runtime, or no runtime. |
| Side effects | Database, object store, network, process, files, environment, and retained evidence. |
| Resource ownership | Creator, borrower, resetter, closer, and partial-failure cleanup owner. |
| Fixture capability | Exact catalog capability for every caller. |
| Security | Cookies, CSRF, token, host, origin, CORS, redaction, and secret handling. |
| Failure semantics | Returned error, test failure, process exit, diagnostics, and cleanup behavior. |
| Observability | Required logs, readiness, exit status, artifacts, and row evidence. |
| Destination | `processtest`, `appsupport`, Recovery testsupport, or local. |
| Decision | `equivalent`, `partially_reusable`, or `not_reusable`, with one specific reason. |

A candidate moves only when it has at least two independent callers, or one
safety-critical protocol whose centralized implementation materially reduces
divergence; hides no owner behavior; has explicit dependencies; preserves
caller-owned lifecycle, fixture, runtime, security, failure, and evidence
parity; and introduces no catch-all cross-owner dependency. An unproven field
means `not_reusable` and the helper remains local.

### 5.4 Recovery-owned support interface

Private Go names and internal decomposition are intentionally unspecified. The
semantic input/output boundary is not.

| Input or output | Required contract |
| --- | --- |
| Prefix | Required non-empty fixture identity; no default. |
| As-of time | If omitted, use current UTC time truncated to one second; return the chosen value. |
| Postgres/S3 harnesses | Explicit borrowed dependencies; Recovery support MUST NOT close them. |
| Evidence location | Caller supplies normalized results root, run ID, target, and group. Recovery support MUST NOT read global environment to discover them. |
| Process environment | Required complete copied map for the restored target; caller owns it; secret-bearing values MUST NOT be logged or retained. |
| Backup identity | Required `backup_set_id` and UTC `consistency_point_at`. |
| Restored identities | Required incident ID, Timeline record ID, Evidence record ID, and object-blob ID. |
| Expected consistency | Required row version, change-set count/digest, blob SHA-256, evidence count, lifecycle state, and `has_evidence` value. |
| Authentication material | Required only for final public-process login; ephemeral and forbidden from artifacts or diagnostics. |
| Safe artifact references | Normalized references to manifest, summary, and verification artifacts; no raw bucket, storage ref, DSN, credential, or incident content. |
| Cleanup | Required idempotent caller-registered cleanup for resources created by support; borrowed resources remain open. |

For local execution, omitted target defaults to `backend-process`. If either the
results root or run ID is absent, the caller resolves both to
`.cartulary/test-results` and `local-backup-restore`. Evidence directories use
mode `0700`; evidence files use mode `0600`. These defaults preserve current
paths while keeping environment lookup outside generalized support.

## 6. Refactor Workstreams

| Workflow ID | Name | Class | Previous | Subsequent | Goal | Validation | Exit checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Source and authority bootstrap | root | none | WF-01 | Lock requirements, baseline, and authorization boundary. | Markdown and source review. | SPT-REQ-001 through 014 are internally consistent. |
| WF-01 | Exact baseline | chain | WF-00 | WF-02 | Run current affected selectors and owner slices before changes. | S-00 commands. | Before result roots are retained. |
| WF-02 | Characterization and duplicate review | chain | WF-01 | WF-03 | Complete AC-399 and helper matrices and confirm unique process evidence. | S-01/S-02 commands. | Every gap has a binary assertion or local disposition. |
| WF-03 | Owner/catalog migration | parallel | WF-02 | WF-04 | Apply Network Flow and Recovery row/topology transactions independently. | Catalog, shape, generation, exact rows. | Each selector resolves once under its normative owner. |
| WF-04 | Helper boundary extraction | chain | WF-03 | WF-05 | Move only admitted Recovery or process mechanics. | Owner slices and boundaries. | Final subprocess evidence and resource ownership remain exact. |
| WF-05 | Mechanical cleanup | chain | WF-04 | WF-06 | Normalize obsolete refusal literal and remove proven-unused wrappers. | Affected exact rows. | No stale helper, selector, or path remains. |
| WF-06 | Reconciliation and handoff | chain | WF-05 | none | Remove crosswalk after retained reconciliation and run final gates. | S-07/S-08 commands. | All SPT-AC criteria pass. |

## 7. Proposed Refactor Slice Plan

Every slice requires later authorization. Product behavior change is outside
these slices and requires separate owner adoption.

| Slice | Depends on | Exact intended change | Likely authored scope | Contract risk | Required tests/evidence | Validation | Rollback | Completion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Record clean current results for the two old rows and all app.server service-backed evidence. | No source change; retained results only. | Dirty baseline or unexecuted selector. | Exact old Network Flow and Recovery rows plus app.server slice. | `make service-backed-test-slice OWNER=app.server ROWS=<two-old-row-ids>`; `make service-backed-test-slice OWNER=app.server` | None. | Exact before roots and failures are classified. |
| S-01 | S-00 | Complete the duplicate-evidence report and every candidate helper/caller equivalence row. | Tracker and characterization tests only if a gap cannot be observed. | False deduplication or helper admission. | Browser helper, operator process, packaged server, and named helper comparisons. | `make test-slice OWNER=app.server`; affected owner slices. | Revert only newly added characterization tests. | Recovery sentinel remains unique and every helper has a disposition. |
| S-02 | S-01 | Add missing explicit AC-399 assertions from Section 4 without changing product behavior. | Recovery sentinel and owner test support admitted by S-01. | Overclaim, flaky identity, or hidden cadence behavior. | Every Section 4 matrix row except AC-401. | Exact old Recovery row; Recovery and app.server service-backed slices. | Revert assertion-only changes if they reveal a product defect and record the defect. | SPT-AC-008 passes before re-ownership. |
| S-03 | S-02 | Remove the old Network Flow row, add its exact replacement, and add `module.networkflow.process → server-harness`. | App.server and Network Flow family manifests; authored topology; generated outputs through Make. | Duplicate selector, lost binary, wrong owner verification. | Exact new row and existing module harness-control evidence. | Catalog; JSON; generate/drift/policy; exact new Network Flow row; app.server slice; backend-process. | Revert authored and generated catalog/topology transaction together. | SPT-AC-004, 006, and 007 pass for Network Flow. |
| S-04 | S-02 | Remove the old Recovery row, add its exact replacement, and extend `module.recovery.process` binary IDs to `operator, server-harness`. | App.server and Recovery family manifests; authored topology; generated outputs through Make. | Duplicate selector, operator prerequisite drift, lost packaged-server proof. | Exact new row, existing operator rows, app.server slice. | Catalog; JSON; generate/drift/policy; exact new Recovery row; app.server slice; backend-process; boundary check. | Revert authored and generated catalog/topology transaction together. | SPT-AC-005, 006, and 007 pass for Recovery. |
| S-05 | S-03, S-04 | Create Recovery testsupport and move only helpers admitted by the completed matrix; retain final subprocess edge locally. | Recovery testsupport, target tests, boundary policy only if imports move. | Lifecycle, secret, artifact, or runtime-identity drift. | Recovery support unit/negative tests plus exact Recovery/app.server rows. | Recovery and app.server focused/service-backed slices; backend-process; boundary check. | Move one helper family per commit-sized transaction and restore it on parity failure. | SPT-AC-009 through 012 pass. |
| S-06 | S-03, S-04 | Replace the obsolete refusal-only WebSocket literal with the canonical incident path and remove only proven-unused wrappers. | Target tests only. | Changed refusal or diagnostic behavior. | Exact startup/config rows. | Narrow app.server service-backed rows. | Revert literal/wrapper cleanup independently. | Refusal behavior is unchanged and no stale literal remains. |
| S-07 | S-05, S-06 | Reconcile active selectors, identities, binaries, collaborators, boundaries, and generated outputs; retain report, then remove the temporary crosswalk. | Authored inputs, generated outputs, tracker. | Permanent alias or stale migration dependency. | Catalog reconciliation and repository reference scan. | Catalog; JSON; generation/policy/drift; harness contract; backend-process. | Restore crosswalk only while reconciliation remains incomplete; never restore runtime aliases. | One active row per selector and no retired ID remains live. |
| S-08 | S-07 | Run final narrow-to-broad verification and complete handoff. | Validation artifacts and tracker. | False success from skipped rows or unrelated diff. | All affected owner and process evidence. | `make agent-finalize`; `make check`; Markdown/diff/status audit. | Roll back the last independent failing slice. | SPT-AC-013 and 014 pass with actual roots. |

## 8. Validation Plan

### 8.1 Current documentation revision

| Layer | Command | Required result |
| --- | --- | --- |
| Markdown | `make lint-markdown` | Pass after the final tracker edit. |
| Whitespace | `git diff --check -- docs/handoffs/serverprocess-app-module-refactor-tracker.md` | No output and exit `0`. |
| Scope | `git status --short` and `git diff --name-only` | Only this tracker is a source/worktree change. |

### 8.2 Later authorized implementation

The later implementation MUST run in this order and MUST record every result
root. A broad pass does not prove that an affected exact row executed.

| Order | Command | Purpose |
| --- | --- | --- |
| 1 | Current exact owner/service-backed rows from S-00 | Baseline before mutation. |
| 2 | `make test-catalog-check` | Owner, family, verification, collaborator, identity, and selector integrity. |
| 3 | `make json-shape-check` | Authored manifest and topology shape. |
| 4 | `make generate` | Refresh generated topology through the public generator. |
| 5 | `make generate-drift` | Prove authored/generated equivalence. |
| 6 | `make generated-artifact-policy-check` | Prove generated-root edit policy. |
| 7 | `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.process.the_packaged_standalone_server_composes_the_netw_400a31ad27` | Exact migrated Network Flow process evidence. |
| 8 | `make service-backed-test-slice OWNER=module.recovery ROWS=module.recovery.process.the_packaged_server_serves_the_coherently_restor_6b1731590e` | Exact migrated Recovery process evidence. |
| 9 | `make service-backed-test-slice OWNER=app.server` | Remaining app.server process evidence after re-ownership. |
| 10 | `make backend-module-boundary-check` | Import and explicit allowance integrity. |
| 11 | `make backend-process` | Canonical aggregate process topology. |
| 12 | `make harness-contract` | Harness contract, catalog, and topology closure. |
| 13 | `make agent-finalize` | Finalizer and retained-run maintenance; record when `RESULTS_DIR` is unset. |
| 14 | `make check` | Broad developer gate after all narrow evidence passes. |

`make browser-e2e` is not required unless a later slice changes packaged
frontend behavior. The duplicate-evidence review does not itself authorize a
browser change.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| SPT-001 | Lock authority, target, and one-file planning boundary | WF-00 | DONE | none | Sections 1 and 2 | Source posture and inventory are complete. |
| SPT-002 | Close Network Flow owner and identity decision | WF-03 | DONE | SPT-001 | Section 3 exact mapping | No architectural decision remains; implementation is S-03. |
| SPT-003 | Close Recovery owner, uniqueness, and identity decision | WF-02/WF-03 | DONE | SPT-001 | Sections 3 and 4 | No ownership question remains; implementation is S-02/S-04. |
| SPT-004 | Close helper classification decision | WF-02/WF-04 | DONE | SPT-001 | Section 5 | Default, destinations, matrix, and output boundary are exact. |
| SPT-005 | Establish current exact baseline | WF-01 | TODO | later authorization | S-00 | Before roots are retained. |
| SPT-006 | Complete characterization and AC-399 gaps | WF-02 | TODO | SPT-005 | S-01/S-02 | Matrices and assertions pass. |
| SPT-007 | Migrate Network Flow catalog/topology | WF-03 | TODO | SPT-006 | S-03 | Exact new row is active and old row absent. |
| SPT-008 | Migrate Recovery catalog/topology | WF-03 | TODO | SPT-006 | S-04 | Exact new row is active and old row absent. |
| SPT-009 | Extract admitted Recovery helpers | WF-04 | TODO | SPT-007, SPT-008 | S-05 | Recovery support and final subprocess edge satisfy Section 5. |
| SPT-010 | Normalize refusal literal and remove obsolete wrappers | WF-05 | TODO | SPT-007, SPT-008 | S-06 | Behavior is unchanged and no stale path remains. |
| SPT-011 | Reconcile and remove temporary crosswalk | WF-06 | TODO | SPT-009, SPT-010 | S-07 | Final retained report exists and no alias remains. |
| SPT-012 | Run final implementation verification | WF-06 | TODO | SPT-011 | S-08 | All affected and broad gates pass or are accurately blocked. |
| SPT-013 | Revise this tracker in NLSpec voice | WF-00 | DONE | SPT-001 through SPT-004 | This file and lint root `.cartulary/test-results/20260807T030530Z-p382756` | Markdown, whitespace, and sole-file scope checks pass. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T22:20:53-04:00 | Codex initial tracker creation | Authority, target, label, and documentation-only scope locked. | Inspected framework, domain, Core 00-04, adopted NLSpecs, target, and adjacent evidence; touched tracker only. | Read-only source searches, Make guidance, `git diff --check`, and `make lint-markdown`. | Initial tracker created and Markdown passed. | Later implementation unauthorized. | Refine decisions before implementation. |
| 2026-08-06T23:00:58-04:00 | Codex NLSpec-style revision | Requirements, exact owner migrations, defaults, helper contracts, and acceptance crosswalk are decision-closed. | Inspected analysis notes, NLSpec-writing guide, live owners, catalog schemas/validators, topology routing, target tests, Recovery browser/operator evidence, and test support; touched tracker only. | Read-only `sed`, `rg`, `jq`, `find`, `git`; Make target explanations and exact row-ID authoring; `git diff --check`; `make lint-markdown`. | Tracker rewritten from owner-backed live state; Markdown passed at `.cartulary/test-results/20260807T030530Z-p382756`; whitespace check passed. | No architectural blocker; implementation requires later authorization. | Begin S-00 only in a later authorized task. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T22:20:53-04:00 | Codex initial tracker creation | Target confirmed as mixed process-level evidence, not production module. | All ten target files, server assembly/profile, processtest, appsupport, and boundary policy; tracker touched. | Full source reads and caller/import searches. | Binary-only evidence retained; helper and Recovery splits identified. | RB-002 and RB-003 were open. | Characterize before movement. |
| 2026-08-06T23:00:58-04:00 | Codex NLSpec-style revision | Boundary decisions are closed: process edge remains, Recovery semantics move only to Recovery testsupport, and all other helpers follow the local default. | Rechecked target, processtest defaults, helper overlap, Recovery support patterns, and runtime-binary topology; tracker touched. | Targeted source and topology searches. | Exact destination and resource-ownership contracts are specified. | None. | Execute S-00 only after authorization. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T22:20:53-04:00 | Codex initial tracker creation | No frontend controller, saved-view, imports/ingest, or grid-vendor surface exists in target. | Target imports, asset test, workbook assertion, view-schema references, browser guidance; tracker touched. | Targeted source searches and browser target explanation. | Frontend workstream omitted beyond packaged assets. | None. | Browser validation only for an actual packaging change. |
| 2026-08-06T23:00:58-04:00 | Codex NLSpec-style revision | Frontend exclusion remains normative; Recovery browser evidence was reviewed only for duplicate classification. | Inspected `apps/web/e2e/restore.spec.ts` and `tools/recoverybrowserrestore`; tracker touched. | Exact source reads. | Browser helper uses in-process `server.Runtime` and does not replace packaged-server evidence. | None. | No frontend slice. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T22:20:53-04:00 | Codex initial tracker creation | HTTP, WebSocket, view-schema, Recovery, extension, and harness surfaces mapped. | Contracts, generated policy, owner catalog, topology, and verification inputs; tracker touched. | Source searches and Make target explanations. | No contract/generated change; Network Flow mismatch identified. | RB-001 was open. | Correct authored inputs only in later task. |
| 2026-08-06T23:00:58-04:00 | Codex NLSpec-style revision | Exact immutable replacement IDs, owner-local verifications, collaborators, and family-binary mappings are fixed. | Inspected family manifests, manifest schema, catalog validator, owner verifications, topology, and row-ID authoring implementation; tracker touched. | Two exact `make author-test-row-id` invocations plus read-only inspection. | Cross-owner verification from analysis notes rejected; exact live-compatible mapping specified. | None. | Apply S-03 and S-04 atomically when authorized. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T22:20:53-04:00 | Codex initial tracker creation | Fourteen app.server process rows and backend-process topology mapped. | App.server, Network Flow, Recovery manifests and topology; tracker touched. | Task guide, owner explanation, target explanations, and searches. | No product suite ran; Network Flow row was misrouted. | RB-001/RB-002 were open. | Establish service-backed baseline. |
| 2026-08-06T23:00:58-04:00 | Codex NLSpec-style revision | Owner decisions, exact row IDs, process routing, duplicate evidence, and validation order are closed. | Inspected target routing, runtime-binary compiler, schema constraints, Recovery browser and operator process tests; tracker touched. | Read-only inspection and exact ID authoring. | New process families retain backend-process routing and receive required binaries. | No planning blocker. | Run S-00 before catalog mutation. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T22:20:53-04:00 | Codex initial tracker creation | Auth, session, CSRF, membership, harness authorization, and production fail-closed behavior frozen. | Core 04, process tests, harness route tests, server profiles; tracker touched. | Source reads and security searches. | Authorization remained with product/platform owners. | None. | Preserve exact outcomes. |
| 2026-08-06T23:00:58-04:00 | Codex NLSpec-style revision | Helper security, secret handling, artifact permissions, and caller-owned lifecycle are explicit. | Rechecked process helpers, Recovery artifact writer, owner security requirements, and harness guard ownership; tracker touched. | Targeted reads/searches. | Generalized support cannot hide auth, environment, profile, redaction, or secret behavior. | None. | Enforce the matrix and SPT-AC-010/012 in S-05. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T22:20:53-04:00 | Codex initial tracker creation | Initial planning was characterization-first but retained three decision gaps. | Tracker and live repository; tracker touched. | Section/table review and Markdown validation. | S-00 through S-08 proposed. | RB-001 through RB-003. | Close decisions before implementation. |
| 2026-08-06T23:00:58-04:00 | Codex NLSpec-style revision | No architectural question remains; only bounded implementation and verification are pending. | Tracker reconciled with live owners, schemas, topology, tests, and analysis notes; tracker touched. | Final Markdown, whitespace, and sole-file scope validation. | S-00 through S-08 now have exact interfaces, defaults, rollback, and acceptance criteria; documentation validation passed. | Later implementation authorization only. | Begin S-00 only in an authorized task. |

## 11. Open Questions and Blockers

There are no remaining architectural questions. These stable IDs now record
resolved decisions and their implementation preconditions.

| ID | Resolved decision | Why it matters | Required implementation evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Network Flow owns the exact replacement process row; app.server is its collaborator; only Network Flow verification IDs are valid. | Satisfies TH-HARNESS-REQ-658/667 without treating the process package as owner. | S-03 catalog/topology transaction and exact new-row result root. | RESOLVED DECISION — IMPLEMENTATION PENDING |
| RB-002 | Recovery owns the exact replacement process row; app.server, Evidence, and Timeline are collaborators; the packaged-server sentinel is unique. | Assigns the AC-399-class aggregate postcondition to Recovery without losing real-process proof. | S-02 assertions, S-04 migration, and exact Recovery result root. | RESOLVED DECISION — IMPLEMENTATION PENDING |
| RB-003 | Default helper disposition is local; only a complete equivalence matrix can admit movement to processtest, appsupport, or Recovery testsupport. | Prevents in-process substitution, lifecycle drift, and a new catch-all helper package. | S-01 matrix and S-05 per-family before/after evidence. | RESOLVED DECISION — IMPLEMENTATION PENDING |

## 12. Binary Completion Criteria

### 12.1 Requirement-to-acceptance crosswalk

| Acceptance | Requirement | Binary criterion | Current state |
| --- | --- | --- | --- |
| SPT-AC-001 | SPT-REQ-001 | No owner contradiction is present; no Core 00-04 edit is proposed; authority order is explicit. | PASS for planning |
| SPT-AC-002 | SPT-REQ-002 | No production Go file exists in target and every future slice retains actual child-process evidence. | PASS for planning |
| SPT-AC-003 | SPT-REQ-003 | Section 4 maps every discovered observable contract to an owner, evidence, characterization, and frozen outcome. | PASS for planning |
| SPT-AC-004 | SPT-REQ-004 | Exactly one active row selects the Network Flow test; it has the exact new ID, owner, family, collaborator, verification, profiles, fixture, tier, and posture in Section 3. | PENDING S-03 |
| SPT-AC-005 | SPT-REQ-005 | Exactly one active row selects the Recovery test; it has the exact new ID and mapping in Section 3, and no browser/operator row replaces its packaged-server proof. | PENDING S-04 |
| SPT-AC-006 | SPT-REQ-006 | Row IDs reproduce through the recorded authoring commands and topology supplies server-harness to Network Flow process and operator plus server-harness to Recovery process. | PENDING S-03/S-04 |
| SPT-AC-007 | SPT-REQ-007 | Old IDs are absent from active catalogs and runtime inputs; no selector overlaps; final reconciliation is retained; temporary crosswalk is removed. | PENDING S-07 |
| SPT-AC-008 | SPT-REQ-008 | Every AC-399 matrix row passes with explicit equality where required, and the row carries no AC-401 claim. | PENDING S-02 |
| SPT-AC-009 | SPT-REQ-009 | Every candidate helper has the exact destination or local disposition in Section 5; no unadmitted helper moves. | PENDING S-01/S-05 |
| SPT-AC-010 | SPT-REQ-010 | Every moved helper has a complete matrix row for every caller and satisfies all admission predicates. | PENDING S-05 |
| SPT-AC-011 | SPT-REQ-011 | Process-support tests prove required binary, real child identity, listener/readiness defaults, bounded exit/stop, and diagnostics behavior. | PENDING affected S-05 evidence |
| SPT-AC-012 | SPT-REQ-012 | Recovery support returns every required semantic output, preserves permissions/redaction/ownership, and cannot start the final server sentinel. | PENDING S-05 |
| SPT-AC-013 | SPT-REQ-013 | All Section 8 implementation commands pass in order with exact roots, or a failure is recorded with relation and rollback; generated roots have no hand edit. | PENDING S-08 |
| SPT-AC-014 | SPT-REQ-014 | Each completed slice has a current handoff row, no unexplained diff, and no unreported skipped check. | PASS for planning; pending implementation handoffs |

### 12.2 Tracker revision completion

This documentation revision is complete only when:

- all ten tracked files and fifteen ignored runtime-evidence files remain
  inventoried;
- every exact mapping, default, interface, workflow, rollback, and acceptance
  criterion above is present without contradiction;
- RB-001 through RB-003 remain decision-resolved and implementation-pending;
- the preceding handoff history and the new session rows are both retained;
- `make lint-markdown` and `git diff --check` pass on the final text; and
- this tracker is the only changed repository source path.

Completion of this tracker MUST NOT be interpreted as authorization for any
production refactor, test movement, catalog edit, topology edit, generation,
or behavioral change.
