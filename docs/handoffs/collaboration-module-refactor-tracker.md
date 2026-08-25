# collaboration Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Value |
| --- | --- |
| Target path | `internal/modules/collaboration` |
| Target label | `collaboration` |
| Output path | `docs/handoffs/collaboration-module-refactor-tracker.md` |
| Repository baseline | Commit `169ba53197681aa45767914b70cd86d8759d0d3f`; `main` was clean and three commits ahead of `origin/main` at discovery time. |
| Planning session | `2026-08-24T17:26:00-04:00` |
| NLSpec-style revision session | `2026-08-24T18:22:06-04:00`; the tracker was already staged as a new file and no other tracked change was present. |
| Execution session | Authorized `2026-08-24T19:09:01-04:00` from commit `169ba53197681aa45767914b70cd86d8759d0d3f`; the tracker remained the only staged path and no index change was authorized. |
| Status | Remediation complete; `WF-EXEC-00` through `WF-SL-08` are `DONE`, all binary exits pass, and no required follow-up remains. |
| Allowed change | The specification, authored contract, implementation, test, test-support, harness-policy, generated projection, and handoff paths required by `WF-EXEC-00` through `WF-SL-08`. Generated roots may change only through their Make-owned generators. |
| Non-goals | No public route, message family, CLI grammar, database schema, Incident Bundle version, authorization precedence, or telemetry-vocabulary change; no compatibility alias, dual path, feature flag, backfill, or production-data migration; no agent-created commit, staging change, or index reset. |
| Deployment posture | Pre-production. Incompatible disposable state uses the adopted reset-only profile; this posture does not weaken any public or security contract. |

The target label was derived from `internal/modules/collaboration` and normalized
to the safe lowercase kebab-case value `collaboration`. The target directory
exists and contained 28 files at the baseline above.

The source hierarchy used for this tracker is:

1. adopted subsystem NLSpecs within their named scopes;
2. Core 00 through Core 04 for implementation-conformance behavior;
3. Core 05 only for claim-bearing timed or fixture-sensitive publication;
4. Domain vocabulary, adopted implementation-boundary decisions, and harness
   mechanics;
5. current repository code and tests for implementation state; and
6. prior plans, handoffs, and the planning framework as evidence only.

Core 05 is not used because this effort makes no timed, fixture-sensitive,
benchmark, or claim-bearing publication. The Revisions decision's assignment
of Collaboration intent publication is an ownership contradiction to resolve
in `WF-SPEC-01`; no implementation cutover may precede that adoption.

Owner and doctrine documents inspected:

- `AGENTS.md`;
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`;
- `docs/domain.md`;
- `docs/spec/00_document_set_status_and_precedence.md`;
- the Collaboration, view-row, job-progress, storage, and operator portions of
  `docs/spec/01_architecture_storage_and_view_contracts.md`;
- the Collaboration and Workbook interaction portions of
  `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`;
- the session, authorization, requeue, and conformance portions of
  `docs/spec/04_security_deployment_and_conformance.md`;
- `docs/testing-harness-nlspec.md`;
- `docs/opentelemetry-instrumentation-nlspec.md`;
- `docs/extension-subsystem-nlspec.md`;
- `docs/network-flow-activity-nlspec.md`;
- `docs/decisions/revisions-module-boundary.md`;
- `docs/decisions/workbook-module-boundary.md`;
- `docs/decisions/projections-module-boundary.md`; and
- `docs/decisions/protocol-ts-public-type-compatibility.md` as historical
  compatibility evidence only.

Repository implementation and projection evidence inspected includes every
target file inventoried in Section 2 and these relevant external surfaces:

- `internal/app/server/runtime_assembly.go`, `collaboration_intents.go`,
  `collaboration_socket.go`, `module_settings.go`, and
  `runtime_dependencies.go`;
- `internal/app/operator/operator_collaboration.go`;
- `internal/app/recoveryassembly/state_catalog.go`;
- `internal/app/timelineassembly/assembly.go`;
- `internal/app/workbookassembly/catalog.go`;
- `internal/app/importassembly/owner_registry.go`;
- representative Collaboration consumers in Revisions, Evidence, Entity
  Mentions, and Entity Merge;
- `internal/testutil/collaborationsupport/intents.go`;
- `apps/web/src/collaboration/IncidentCollaborationSession.tsx`;
- `apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts` and
  `workbookCollaborationMessages.ts`;
- `apps/web/src/networkFlow/networkFlowCollaborationInterpreter.ts`;
- `packages/protocol-ts/src/entrypoints/collaboration.ts` and its generated
  Collaboration type/validator surfaces;
- `contracts/ws/index.schema.json`;
- `contracts/collaboration/index.json`, the v2 requeue registry, schema, and
  fixtures;
- `contracts/recovery/fixtures/recovery-state-catalog.v1.json`;
- `contracts/verification/owners/module.collaboration.json`;
- `db/migrations/00029_collaboration.sql`;
- `tools/backend_module_boundaries.json`;
- `tools/schema_object_ownership_manifest.json`;
- `tools/test_catalog_owner.json`;
- `tools/test_families/module.collaboration.json` and relevant collaborating
  owner manifests; and
- the authored Make task surface used by the discovery commands in Section 8.

The implementation, behavior corrections, contract changes, generated
refreshes, and harness repairs described here were explicitly authorized on
`2026-08-24`. This tracker is the controlling execution and handoff artifact.

This tracker is prescriptive for execution of the refactor, but it is not an
adopted behavioral NLSpec and does not create product behavior. In this
tracker, `MUST`, `MUST NOT`, and `REQUIRED` have binding force only when the
statement either traces to an adopted owner or is explicitly labeled as a
tracker execution gate. A tracker gate binds sequencing and evidence handling;
it does not supersede an adopted owner.

`temp/analysis-notes.md` was inspected as planning research, and
`docs/research/nlspec-spec.md` was inspected as writing-quality guidance. They
are not owner documents. The analysis recommendations are incorporated only
where they conform to the hierarchy above; any difference from an adopted
owner is recorded explicitly rather than silently resolved.

### 1.1 Load-bearing tracker requirements

Each requirement is defined once here and referenced by ID elsewhere.

| ID | Requirement | Authority/status |
| --- | --- | --- |
| `CT-REQ-001` | Work MUST preserve every observable contract in Section 4 unless an adopted owner explicitly requires the authorized correction. | Adopted-owner trace; active. |
| `CT-REQ-002` | Pre-production cutovers MUST be atomic. The target plan MUST NOT introduce compatibility aliases, dual readers or writers, shadow paths, feature flags, historical backfills, or production-data migrations. Incompatible disposable state MUST use the adopted reset-only path. | Adopted pre-production posture plus proposed refactor gate; active. |
| `CT-REQ-003` | Private Revisions conflict facts and public Collaboration publication facts MUST be separate inputs. Neither consumer may derive its facts from the other consumer's representation, a portable snapshot, a projection row, or arbitrary whole-row JSON comparison. | Adopted by REQ-00-075 and both implementation decisions. |
| `CT-REQ-004` | A committed live mutation, its revision and private facts, its projection effects, and exactly one Collaboration intent MUST share one authoritative transaction. Any pre-commit failure MUST leave all of those effects absent. | Core 01 transaction behavior; active. |
| `CT-REQ-005` | Collaboration MUST remain the owner of canonical public `record_changed` construction, validation, intent append, replay ordering, and wire delivery. Revisions MUST remain the owner of private revision/conflict persistence. | Core behavior and adopted REQ-00-075 allocation. |
| `CT-REQ-006` | Known `/ws/v1/` server messages MUST admit owner-allowed additive members without weakening required-member, known-type, duplicate-member, framing, or size validation. | Core 01; downstream correction explicitly authorized by `G-03`. |
| `CT-REQ-007` | Implementation MUST begin from retained B0 owner-slice evidence and MUST use B1 after the accounting-only `SL-01` cut. | Testing Harness NLSpec plus execution gate; B0 and B1 recorded, RB-002 resolved. |
| `CT-REQ-008` | Authored owner inputs MUST change before generated projections; generated roots MUST be refreshed only through Make and MUST never be hand-edited. | Repository generated-artifact policy; active. |
| `CT-REQ-009` | Every completed slice MUST retain its exact source SHA, commands, run roots, row outcomes, rollback point, and binary acceptance result. | Testing Harness NLSpec plus tracker handoff gate; active. |
| `CT-REQ-010` | Workstreams MUST execute in the order in Section 1.3. A successor may begin only after its predecessor's evidence, compatibility outcome, rollback point, residual risks, and Markdown-lint result are recorded and the predecessor is `DONE`. | Authorized execution gate; active. |

### 1.2 Document placement and authority

| Information | Controlling location | Tracker treatment | Prohibited treatment |
| --- | --- | --- | --- |
| Public WebSocket, replay, authorization, revision, and client behavior | Adopted Core and subsystem NLSpecs | Import by requirement reference and freeze in Section 4. | Redefining behavior from repository code, tests, or analysis notes. |
| Exact Revisions and Collaboration Go topology | Adopted implementation-boundary decisions | Record current authority and a proposed replacement topology separately. | Claiming a proposed tracker topology is already adopted. |
| Authored schemas and registries | `contracts/**` owner inputs | Treat as downstream executable projections and planned edit sources. | Letting a projection override its adopted owner. |
| Generated Go and TypeScript | Declared generated roots | Record expected generated consequences and Make validation. | Hand editing or treating generated output as behavioral authority. |
| Commands, run roots, row results, and failure classifications | Tracker and retained harness evidence | Record exact execution evidence and status. | Treating prose or a phase map as proof that a test passed. |
| Rationale, external implementation guidance, and research | Research notes or decision rationale | Use only to explain a conforming recommendation. | Promoting supporting research into Cartulary authority. |

### 1.3 Authorized remediation execution

The selected architecture is the independent-port topology: source owners
derive separate private Revisions facts and public Collaboration effects from
admitted semantic operations. Revisions and Collaboration MUST NOT derive from
or import one another's representations. Each source owner supplies both ports
through one borrowed source transaction where live publication is required.

| Gap | Required remediation | Areas | Durable outcome | Compatibility and migration | Risk if unresolved | Completion evidence |
| --- | --- | --- | --- | --- | --- | --- |
| `G-01` | Adopt the Collaboration boundary, amend Revisions ownership, add `REQ-00-075` and `AC-564`, complete traceability, and correct domain terminology. | Specification, decisions, conformance, tracker. | Import direction, composition, transaction, and compatibility rules become authoritative for later message families and owners. | Internal architecture and documentation only. | Wire publication can remain implicitly Revisions-owned and future code can redefine ownership. | Adopted texts agree; Markdown lint and final `AC-564` routing pass. |
| `G-02` | Establish B0/B1, correct the `/ws/v1/` capability label, remove the prose-only evidence index, and retain mechanically unique active evidence. | Tests, harness inputs, tracker. | Deleted or moved tests cannot leave false lifecycle evidence. | Test selectors may move atomically; runtime is unchanged. | Pre-existing failure and refactor regressions cannot be distinguished reliably. | Owner slices, harness contract, retained manifests, and one terminal result per selected row. |
| `G-03` | Complete the server-envelope and replayable-payload projections; generate a non-mutating known-member projector; replace the weak frontend `record_changed` guard. | Contracts, generator, generated Go/TypeScript, frontend, tests. | Additive `/ws/v1/` growth remains accepted while known fields stay strongly typed and bounded. | Valid additive messages become accepted, incomplete messages rejected, and admitted unknown members are omitted from decoder results; private `0.0.0` consumers migrate atomically. | Server additions can disable live updates and malformed messages can reach weak guards. | `DEC-AC-01` through `DEC-AC-07`, negative envelope cases, compile/frontend/browser checks, and generated drift/policy checks. |
| `G-04` | Introduce one runtime/application facade; retain one `RegisterRoutes`; move semantic protocol and private route/hub orchestration behind narrow owner boundaries. | Implementation, assembly, boundary policy, tests. | Lifecycle and security-sensitive route behavior can evolve independently of persistence and process mechanics. | Internal Go imports cut over atomically; public route behavior is unchanged; no aliases remain. | Concrete hub, replay, dispatcher, and persistence lifecycles keep a broad blast radius. | Collaboration/app-server slices, lifecycle, shutdown, socket/browser, and boundary checks. |
| `G-05` | Split private PostgreSQL stream storage from dispatcher scheduling, listening, retry, tailing, and retention. | Implementation, storage/runtime tests, boundary policy. | Failure domains and future retention or multi-process changes become isolated. | No schema or stored-data change; intent keys, order, resume, quarantine, and retention stay authoritative. | A small change can disturb append atomicity, replay, retry, and shutdown together. | Full stream suite including tailing, idempotency, reset, quarantine, retention, and ownership. |
| `G-06` | Replace shared row diff, broad intents, and the combined revision/publication operation with explicit live revision facts, public record effects, a distinct historical operation, and an immutable publication catalog. | Decisions, implementation, assembly, boundary policy, tests. | Private conflict reconstruction and public disclosure evolve independently with no whole-row inference. | Atomic internal API cutover; no compatibility path or schema change. | Public helpers remain private conflict authority and new private/computed fields can leak or cause false conflicts. | `CT-AC-01` through `CT-AC-10` and exhaustive source-owner/destructive-path coverage. |
| `G-07` | Move requeue SQL, locking, proof, and audit persistence behind private recovery adapters; expose only a narrow Operator capability and pure recovery-state contribution. | Implementation, Operator composition, tests, boundary policy. | Recovery can evolve without broadening or destabilizing the live protocol. | CLI bytes, exit codes, table state, audit behavior, and catalog remain unchanged. | Recovery SQL/audit concerns can leak into root API or gain an unintended network surface. | Operator unit/process, requeue integration/concurrency/outcome, role, and catalog checks. |
| `G-08` | Consolidate socket and semantic intent helpers in `internal/testutil/collaborationsupport`; remove the module-local runtime wrapper; retain direct SQL only for physical storage, recovery, or role tests. | Tests, test utilities, selectors/topology. | Storage and package moves require fewer unrelated owner edits while semantic atomicity stays testable. | Atomic test import/selector migration only. | Cross-owner tests become schema authorities and composition fixtures diverge. | All importing owner slices, harness contract, path/selector audits, and drift where inputs change. |
| `G-09` | Enforce the final topology against retired constructors, cross-imports, raw Collaboration SQL, broad intents, aliases, and old support paths; close retained evidence and handoff. | Policy, cleanup, tests, tracker. | Forbidden coupling fails immediately and the completed topology is reproducible. | Internal cleanup only; no deprecation period. | Retired paths can return and prose can claim completion without evidence. | Exact audits, boundary/generated checks, `make agent-finalize`, `make check`, and complete handoff entries. |

Workstreams are strictly sequenced. The tracker checkpoint is part of every
workstream, not a later documentation task.

| Sequence | Workstream | Scope | Status | Dependency | Binary exit |
| ---: | --- | --- | --- | --- | --- |
| 0 | `WF-EXEC-00` | Activate this tracker and establish retained B0. | `DONE` | None. | B0 passes or every non-pass is classified; checkpoint and Markdown lint pass. |
| 1 | `WF-SPEC-01` | Adopt specification and ownership boundaries. | `DONE` | `WF-EXEC-00` `DONE`. | Owner texts agree on topology, ports, transactions, cutover, and compatibility; lint passes. |
| 2 | `WF-SL-01` | Repair harness accounting and establish retained B1. | `DONE` | `WF-SPEC-01` `DONE`. | Canonical evidence only; B1 retained; no runtime behavior change. |
| 3 | `WF-SL-02` | Correct WebSocket projection, decoder, and frontend consumption. | `DONE` | `WF-SL-01` `DONE`. | Complete typed envelopes decode to fresh known-member projections; narrow and browser validation pass. |
| 4 | `WF-SL-03` | Introduce the runtime, route, protocol, and hub facade. | `DONE` | `WF-SL-02` `DONE`. | One registrar/lifecycle remains and route/security behavior passes. |
| 5 | `WF-SL-04` | Split durable stream storage from dispatcher runtime. | `DONE` | `WF-SL-03` `DONE`. | Root has no stream SQL and complete stream behavior passes. |
| 6 | `WF-SL-05` | Cut over explicit private/public fact ports and all callers. | `DONE` | `WF-SL-04` `DONE`. | All acceptance criteria and owner slices pass; no old diff, combined API, cross-import, or compatibility path remains. |
| 7 | `WF-SL-06` | Isolate recovery mechanics. | `DONE` | `WF-SL-05` `DONE`. | Operator/recovery contracts and local-only security pass unchanged. |
| 8 | `WF-SL-07` | Consolidate shared test support. | `DONE` | `WF-SL-06` `DONE`. | No stale helper import or duplicate runtime wrapper remains; affected owner and harness checks pass. |
| 9 | `WF-SL-08` | Enforce final topology, validate, and complete handoff. | `DONE` | `WF-SL-07` `DONE`. | Audits and broad checks pass or unrelated failure is completely classified; all workstreams are `DONE`. |

At every checkpoint record start/end time, starting HEAD and exact status,
files changed, requirement/gap IDs, substantive result, exact Make commands,
stable command/run identifiers, selected row count and non-pass outcomes,
compatibility result, rollback scope, residual risk, and binary exit. Run and
record `make lint-markdown` after the checkpoint edit. A failed workstream stays
`IN_PROGRESS` or becomes `BLOCKED`; its successor does not start.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/collaboration/codec.go` | Semantic WebSocket envelope codec, frame-kind policy, duplicate-member rejection, size limit, and message vocabulary. | `MaximumMessageBytes`, `MessageKind`, `Socket`, socket transport function types, `Codec`, decode failures, message-type predicates. | Collaboration routes; server socket adapter; socket test support. | Go JSON and HTTP abstractions only; the concrete WebSocket library remains in `internal/platform/ws`. | `codec_test.go`, `socket_test.go`. | `contracts/ws/index.schema.json`; generated protocol validators/types indirectly. | Collaboration semantic protocol, with transport implementation kept in Platform. | high | Correctly avoids importing the concrete WebSocket vendor. |
| `internal/modules/collaboration/codec_test.go` | Characterizes semantic codec behavior, additive envelope members, UTF-8, duplicate members, binary frames, and limits. | Test surface only. | `module.collaboration.support_unit` verification row. | `codec.go`. | Self. | WebSocket v1 compatibility posture. | Collaboration tests. | medium | Freezes additive envelope-member tolerance. |
| `internal/modules/collaboration/historical_restore.go` | Transaction-local suppression of historical `record_changed` intent publication. | `HistoricalIntentPolicy` constructor and pgx/database-sql suppression methods. | Server assembly and Revisions historical import/restore paths. | `pgx`, `database/sql`, and the Collaboration PostgreSQL setting. | `historical_restore_test.go`, Revisions restore/import tests. | Core 01 historical-intent rule; no generated file. | Collaboration publication policy, exposed through a narrow port to Revisions. | high | Persistence-adjacent but semantically owned by Collaboration publication behavior. |
| `internal/modules/collaboration/historical_restore_test.go` | Proves suppression is transaction-local and does not leak after commit or rollback. | Test surface only. | Collaboration integration verification. | Scenario runtime and PostgreSQL. | Self. | None. | Collaboration tests. | high | Required characterization before moving the policy. |
| `internal/modules/collaboration/hub_telemetry.go` | OpenTelemetry spans, counters, active-connection gauge, and safe event/drop vocabularies for the hub. | `Hub.ConfigureTelemetry`; remaining helpers are private. | `protocol.go` hub and server composition. | `internal/platform/telemetry`, OpenTelemetry APIs. | `hub_telemetry_test.go`. | OpenTelemetry semantic contracts, not WS code generation. | Collaboration telemetry over Platform telemetry substrate. | medium | Adopted telemetry NLSpec assigns Collaboration WS signal ownership here. |
| `internal/modules/collaboration/hub_telemetry_test.go` | Characterizes safe telemetry vocabularies and no-SDK behavior. | Test surface only. | Collaboration support-unit verification. | Hub telemetry. | Self. | OpenTelemetry conformance indirectly. | Collaboration tests. | medium | Prevents identifiers or unbounded values from becoming attributes. |
| `internal/modules/collaboration/hub_test.go` | Characterizes presence scope/order/expiry, revocation isolation, subscription teardown, and canonical patch/invalidate behavior. | Test surface only. | Collaboration support-unit verification and the local evidence index. | `protocol.go`, record-change payload construction. | Self. | WebSocket message semantics. | Collaboration tests. | high | The evidence index names two subtests that do not exist in this file. |
| `internal/modules/collaboration/incident_session_notifier.go` | Post-commit incident-close and membership-revocation notifications to active Collaboration sessions. | `IncidentSessionNotifier` and its constructor/notification methods. | `internal/app/collaborationassembly/incidenteffects` via server assembly. | `Hub` only; no storage or auth dependency. | Incident-effect coordinator tests and socket/integration tests. | Session-revocation wire contract. | Collaboration application capability consumed by Incidents effects. | high | Authored boundary policy explicitly keeps this notifier storage-free. |
| `internal/modules/collaboration/integration_test.go` | Real-client presence/replay, revocation sources, incident closure, replay-family filtering, and browser-origin rejection. | Test surface plus local setup helpers. | Five Collaboration integration verification rows. | Shared application runtime, auth flow, jobs, socket test support, PostgreSQL. | Self. | WS contract and harness evidence. | Collaboration integration tests. | critical | Primary end-to-end backend contract characterization. |
| `internal/modules/collaboration/intent_validation_test.go` | Validates all durable event families, including acceptance of additive job and extension payload members. | Test surface only. | Collaboration support-unit verification. | `NewEventIntent`. | Self. | Conflicts with the closed generated frontend payload validators described in Section 5. | Collaboration tests. | high | Concrete evidence of a projection/client-decoder mismatch. |
| `internal/modules/collaboration/protocol.go` | In-memory hub, presence, revocation/termination, replayable delivery, public message DTOs, row-change conversion, patch construction, job progress, extension invalidation, and presence validation. | `Message`, `Hub`, timing/status constants, payload types, hub methods, replay predicates, message builders/validators, `BuildViewRowPatch`, and record-change conversion. | Routes, dispatcher, incident notifier, server observers, tests, and test support. | UUID, JSON, time, synchronization, Collaboration record-change types. | `protocol_test.go`, `hub_test.go`, socket/integration tests. | `contracts/ws/index.schema.json`, generated Go WS artifact, generated protocol-ts Collaboration types/validators. | Collaboration protocol; row/view-specific materialization seam requires explicit owner resolution. | critical | Broadest mixed semantic surface in the target. |
| `internal/modules/collaboration/protocol_test.go` | Characterizes hub terminal behavior, generated WS payload shapes, job progress, extension invalidation, and extension-workspace presence. | Test surface only. | Collaboration support-unit verification. | Generated WS artifact through `contracttest`; protocol/hub. | Self. | Directly reads generated WS contract artifact. | Collaboration tests. | high | Existing schema assertions do not characterize additive replayable payload acceptance in protocol-ts. |
| `internal/modules/collaboration/record_change_intent.go` | Computes changed cell keys, canonicalizes view-row cells, and builds deterministic durable `record_changed` intents. | `RecordChange`, `AffectedViewChange`, `ChangedCellKeys`, `NewRecordChangeIntent`. | Revisions, Timeline assembly, Evidence, Entity Mentions, Entity Merge, and tests. | JSON, UUID, Collaboration stream intent. | `record_change_intent_test.go` and multiple source-owner integration tests. | WS `record_changed` and `view_row_patch_v1` contracts. | Collaboration retains wire-intent validation and canonicalization; source owners and Revisions receive separate proposed fact contracts under `CT-REQ-003`. | critical | Current exported helper is also used by Revisions for private conflict facts. The replacement topology is planning-complete but not owner-adopted. |
| `internal/modules/collaboration/record_change_intent_test.go` | Characterizes sorted field keys, affected-view ordering, patch compaction, duplicates, and invalid input. | Test surface plus a local helper invoked by `shared_harness_test.go`. | Collaboration support-unit verification. | Record-change intent builder. | Self and `shared_harness_test.go`. | WS row-change contract. | Collaboration tests. | high | Must survive any ownership split. |
| `internal/modules/collaboration/recovery.go` | Deployment-local stream-quarantine requeue transaction, repair proof checks, failure typing, state reset, and administrative journal write. | Requeue failure/request/result types, `RecoveryService`, constructor, `RequeueIncident`. | Operator application and operator process tests. | `postgres.DB`, direct SQL, `internal/platform/administrativeaudit`. | `stream_integration_test.go`, operator unit/process tests. | `contracts/collaboration/**`; operator result generated artifact. | Collaboration recovery capability with private persistence/audit adapters. | critical | Correct Collaboration owner, but persistence and audit mechanics should be hidden behind a narrower root surface. |
| `internal/modules/collaboration/recovery_state.go` | Declares four Collaboration tables and restore-generation invalidation algorithm to Recovery assembly. | `RecoveryStateContribution`. | `internal/app/recoveryassembly`. | `internal/platform/recoverystate`. | Recovery catalog tests. | Recovery-state catalog fixtures and generated recovery artifacts. | Collaboration source-owner recovery contribution. | high | Source owner should continue constructing the contribution. |
| `internal/modules/collaboration/routes.go` | Registers and runs the incident WebSocket route, session establishment, replay, heartbeat, authorization rechecks, presence, terminal effects, and close/error mapping. | `Service`, `Settings`, `RegisterRoutes`. | Server route catalog and tests. | Incident admission, platform auth/session, HTTP API/auth, codec, hub, and replay store. | `socket_test.go`, `integration_test.go`. | WS route and message contract. | Thin Collaboration application facade over private route orchestration. | critical | Transport-adjacent and authorization-sensitive; public route behavior must remain frozen. |
| `internal/modules/collaboration/shared_harness_test.go` | Declares socket-event harness coverage and a prose evidence index. | Test surface only. | `module.collaboration.support_unit.shared_harness...` row. | Target-local `incidentwstest` inventory. | Self. | Harness accounting only. | Collaboration test accounting. | high | It validates string shape but not referenced test existence; two hub subtest references are stale. |
| `internal/modules/collaboration/socket_test.go` | Characterizes first-message grammar, repeated handshake rejection, resume reset, strict frames, heartbeat/session expiry, and incident-scoped presence/replay. | Test surface and local helpers. | Collaboration store/integration verification rows. | Real server runtime, auth flow, timeline, socket test support. | Self. | WS route/message contract and harness rows. | Collaboration tests. | critical | Essential route-level regression suite. |
| `internal/modules/collaboration/stream.go` | Durable intent append, replay-token store, replay queries, dispatcher lifecycle, sequencing, retry/quarantine, PostgreSQL notification listening, multi-process tailing, and retention pruning. | Event-family constants, `EventIntent`, `IntentAppender`, `ReplayStore`, `ReplayResult`, `Dispatcher`, constructors and lifecycle methods. | Server assembly, routes, jobs/network-flow translators, Revisions, Timeline, Evidence, Entities, tests and test utilities. | `postgres.DB`, pgx, direct SQL, crypto token/hash, hub broadcaster. | `stream_integration_test.go`, intent tests, numerous cross-owner atomicity tests. | Migration 00029, schema ownership manifest, recovery fixtures, backend boundary policy. | Collaboration durable stream with private persistence and runtime sub-boundaries. | critical | A 1,032-line mixed persistence/runtime coordinator and the main refactor pressure point. |
| `internal/modules/collaboration/stream_integration_test.go` | Characterizes transaction atomicity, idempotent/divergent keys, ordering, retries, restart/tail behavior, quarantine/requeue, concurrency, commit outcomes, legacy payloads, tokens, and retention. | Test surface plus local helpers. | Collaboration integration/store verification rows. | Dedicated PostgreSQL runtime, application helpers, auth, timeline. | Self. | Collaboration tables, requeue contract, recovery behavior. | Collaboration integration tests. | critical | Slice-level baseline for persistence or recovery movement. |
| `internal/modules/collaboration/telemetry.go` | WebSocket lifecycle and dispatcher telemetry, safe operation/result classification, and HTTP error mapping. | No independent exported package API; used by routes/dispatcher. | Routes and stream dispatcher. | Platform telemetry and HTTP API error types. | `telemetry_test.go`. | OpenTelemetry semantic contract. | Collaboration telemetry. | medium | Semantic signal ownership is intentional; exporter/runtime plumbing remains Platform-owned. |
| `internal/modules/collaboration/telemetry_test.go` | Characterizes lifecycle vocabulary and public-error classification. | Test surface only. | Collaboration support-unit verification. | `telemetry.go`, platform HTTP errors. | Self. | OpenTelemetry conformance indirectly. | Collaboration tests. | medium | Preserve exact safe categories. |
| `internal/modules/collaboration/testsupport/incidentwstest/incidentwstest.go` | Shared real WebSocket client, hello/resume helpers, message reads, assertions, and close behavior. | Exported test-only client/options/helpers. | Collaboration socket/integration tests and Timeline test support. | Target protocol types, auth/test HTTP helpers, `internal/testutil/wstest`. | Collaboration and Timeline tests. | WS test evidence and owner rows. | Candidate `internal/testutil/collaborationsupport`; protocol-specific assertions may remain Collaboration-owned. | high | Cross-module use makes placement a real test-util boundary finding. |
| `internal/modules/collaboration/testsupport/incidentwstest/inventory.go` | Declares required socket-event harness capabilities and surface labels. | `HarnessID`, requirements, inventory entry, inventory/assertion helpers. | `shared_harness_test.go`. | Testing package only. | `shared_harness_test.go`. | Harness accounting. | Collaboration test accounting or shared testutil after consolidation. | high | Contains stale surface `GET /ws/incidents/{incident_id}` without `/v1`. |
| `internal/modules/collaboration/testsupport/incidentwstest/view_events.go` | Connects view-scoped socket clients and asserts `record_changed` view effects. | `ViewSocket`, connector and record-change assertion helpers. | Collaboration tests and cross-module scenario support. | Incident socket client and Collaboration record-change decoder. | Cross-owner integration tests. | `record_changed`/view schema test contract. | Candidate shared Collaboration testutil. | high | Semantically tied to the public Collaboration/view contract. |
| `internal/modules/collaboration/testsupport/intenttest/intents.go` | Direct SQL load/insert/count helpers for Collaboration intent fixtures, including legacy job progress. | `IntentRecord` and SQL test helpers. | Collaboration stream tests and potentially owner integration tests. | pgx pool and Collaboration tables. | Stream/legacy integration tests. | Migration 00029 table shape. | Collaboration owner-specific storage test support. | medium | Direct SQL is acceptable for owner storage tests but must not become production authority. |
| `internal/modules/collaboration/testsupport/scenariotest/harness.go` | Thin wrapper over generic application runtime/server test composition. | `RuntimeHarness`, `ServerHarness`, `StartRuntime`, `StartServer`. | Collaboration socket, stream, and historical integration tests. | `internal/testutil/appsupport`, HTTP route mode. | Collaboration integration tests. | Harness mechanics only. | `internal/testutil/appsupport` or a narrower shared Collaboration support facade. | medium | Duplicates generic reusable application test composition at a module-local path. |

No file under the target is out of scope for this inventory.

## 3. Module Boundary Diagnosis

The repository evidence supports a permanent Collaboration implementation and
supporting capability, but not a Collaboration domain bounded context and not
the current flat package as the ideal final topology. Presence and live
save/conflict interaction remain Workbook Interaction language. The target is
a mixed-responsibility package containing a legitimate application facade,
view/protocol orchestration, transport-adjacent logic,
persistence-adjacent adapters, and test-support surfaces. It is not a frontend
shell, grid-vendor adapter, generic projection owner, saved-view owner, or
catch-all for arbitrary source mutation.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Incident WebSocket semantic route and protocol | `codec.go`, `routes.go`, `protocol.go` | Collaboration, with concrete socket implementation in Platform and composition in Server | split | Core 01 WS contract; Platform boundary policy; server socket adapter | Preserve `RegisterRoutes` as the server-facing facade while hiding orchestration details. |
| Presence, hub fan-out, replay delivery, revocation, and terminal closure | `protocol.go`, `incident_session_notifier.go` | Collaboration | keep | Core 01/Core 03 and incident-effect composition | These are central Collaboration responsibilities. |
| Durable intent append and stream ordering | `stream.go`, `record_change_intent.go` | Collaboration | split | Core 01 same-transaction intent and replay rules; source modules use `IntentAppender` | Keep the consumer-facing intent capability; hide persistence/runtime implementation. |
| PostgreSQL replay/token/dispatcher storage | `stream.go` | Collaboration-owned private persistence adapter | split | Collaboration owns the four tables in schema and boundary manifests | Do not move Collaboration table meaning to generic Platform/Postgres. |
| Quarantine requeue and recovery-state declaration | `recovery.go`, `recovery_state.go` | Collaboration with operator/recovery application adapters | split | Core 00 owner matrix; Core 01/Core 03/Core 04 requeue allocations | Preserve deployment-local-only authority and exact CLI/result behavior. |
| WebSocket telemetry semantics | `hub_telemetry.go`, `telemetry.go` | Collaboration over Platform telemetry substrate | keep | Adopted OpenTelemetry NLSpec | No generic telemetry move is indicated. |
| Canonical `record_changed` envelope, arrays, and sparse patch fallback | `record_change_intent.go`, `protocol.go` | Collaboration as the Core 01 wire-contract implementer | keep | Core 01 REQ-01-267 and Core 00 owner matrix | Source owners supply semantic effects; Collaboration validates/canonicalizes the wire intent. |
| View-row cell diff used for Collaboration and Revisions conflict facts | `ChangedCellKeys` in `record_change_intent.go` | Separate source-owner inputs for Revisions private facts and Collaboration public facts | split | Revisions uses it for private conflict facts; source owners also use it for Collaboration consequences | The target topology is adopted by REQ-00-075, both boundary decisions, `CT-REQ-003`, and Sections 3.1-3.4. |
| Network Flow invalidation payload | Generic payload in `protocol.go`; source translation in server/Network Flow composition | Collaboration owns generic envelope; Network Flow owns resource semantics | keep | Extensions NLSpec EXT-REQ-093 and Core 01 | No Network Flow domain logic should enter Collaboration. |
| Frontend connection, reset, presence, row application, and authorization recovery | `apps/web/src/collaboration` and `apps/web/src/workbook/collaboration` | Frontend Collaboration session and Workbook coordinator | defer | Live frontend consumers | Out of target; freeze as contract consumers. |
| Reusable socket/scenario test composition | `testsupport/**` and `internal/testutil/collaborationsupport` | Shared `internal/testutil` with owner-semantic helpers retained near Collaboration | split | Cross-module imports and repository testutil boundary | Decide file-by-file during `SL-07`; do not expose test-only assumptions to production. |
| Grid-vendor integration | Not present in target | Existing grid adapter/frontend surfaces | keep | Search found no backend target imports or vendor tokens | `intentional/no_action`. |

### 3.1 Adopted topology status

REQ-00-075 adopts `docs/decisions/collaboration-module-boundary.md`, and the
same workstream amends `docs/decisions/revisions-module-boundary.md`. The
independent-port topology below is therefore the controlling implementation
architecture. Research remains rationale only.

| Concern | Current implementation | Proposed target | Adoption status |
| --- | --- | --- | --- |
| Private changed-field calculation | Revisions calls Collaboration `ChangedCellKeys` over live view-shaped rows. | A source owner supplies explicit private facts to Revisions. | Adopted; implementation pending `WF-SL-05`. |
| Public changed-field calculation | Revisions and several source owners derive or pass Collaboration fields through different paths. | A source owner supplies a separate public fact set to Collaboration. | Adopted; implementation pending `WF-SL-05`. |
| Revision append | Revisions persists canonical revision state and private conflict facts. | Revisions continues to own private persistence. | Adopted. |
| Public intent append | Revisions or a source-owner adapter constructs and appends the intent, depending on the caller. | Collaboration validates, canonicalizes, and appends through one transaction-bound capability. | Adopted; implementation pending `WF-SL-05`. |
| Wiring | Revisions imports Collaboration and app assembly also supplies direct Collaboration adapters. | Source-local one-method consumer interfaces are wired by application assembly; Revisions and Collaboration do not import each other. | Adopted; implementation pending `WF-SL-05`. |
| Compatibility | `ChangedCellKeys` is an exported transitional dependency. | One atomic cutover removes it without an alias or shared replacement helper. | Required by `CT-REQ-002`. |

### 3.2 Adopted transaction-bound interfaces

Under the independent-port alternative, each authoritative source owner MUST
declare the two one-method interfaces it consumes. Their semantic operations
are:

| Consumer interface | Operation | Input | Output | Transaction rule |
| --- | --- | --- | --- | --- |
| `liveRevisionAppender` | `AppendLiveRevisionTx(ctx, tx, input)` | `LiveRevisionInput` | `error` only; the caller already owns `change_set_id` | Uses the borrowed source transaction and MUST NOT begin, commit, roll back, or nest a transaction. |
| `recordChangeIntentAppender` | `AppendRecordChangedTx(ctx, tx, input)` | `RecordChangeIntentInput` | `error` only | Uses the same borrowed transaction and MUST append exactly one deterministic semantic intent. |

An authority-compatible alternative MAY keep one Revisions coordination call,
but it MUST expose the same two disjoint input models internally and MUST route
the public model through a Collaboration-owned appender. It MUST NOT re-create
one shared diff representation.

| Input model | Member | Type/shape | Required rule |
| --- | --- | --- | --- |
| `LiveRevisionInput` | `change_set_id` | Non-nil UUID | Identifies the already-admitted change set; no implicit creation. |
| `LiveRevisionInput` | `record_id` | Non-nil UUID | Exact authoritative record identity. |
| `LiveRevisionInput` | `row_version` | Integer `>= 1` | Exact post-mutation authoritative row version. |
| `LiveRevisionInput` | `before_snapshot`, `after_snapshot` | Revisions-owned canonical snapshot variants | At least one is present; schema and record identity MUST agree when both are present. |
| `LiveRevisionInput` | `conflict_facts[]` | Ordered set of `RevisionConflictFact` | Required for an ordinary live revision; omitted entirely on historical import. |
| `RevisionConflictFact` | `field_key` | Non-empty stable field key | Unique by exact byte value; visible headers, cell positions, and JSON property aliases are invalid substitutes. |
| `RevisionConflictFact` | `before_present`, `after_present` | Boolean | At least one MUST be true. |
| `RevisionConflictFact` | `before_value`, `after_value` | Private source-semantic value | A value is present if and only if its matching presence flag is true; these values never enter Collaboration or a portable bundle. |
| `RecordChangeIntentInput` | `incident_id`, `record_id`, `change_set_id`, `actor_user_id` | Non-nil UUIDs | Exact identities from the admitted mutation. |
| `RecordChangeIntentInput` | `row_version` | Integer `>= 1` | Exact committed-result row version. |
| `RecordChangeIntentInput` | `client_txn_id` | String | Exact admitted client transaction ID; defaults to the existing empty-string representation when the change set has no client transaction. |
| `RecordChangeIntentInput` | `mutation_ordinal` | Non-negative integer | Deterministic ordinal within the change set; no clock-derived or process-local default. |
| `RecordChangeIntentInput` | `created_at` | UTC timestamp | Exact change-set creation time; the appender MUST NOT substitute its current clock. |
| `RecordChangeIntentInput` | `changed_field_keys[]` | Public stable field-key set | Caller order is non-semantic. Collaboration validates the source-owner public registry, rejects duplicates or unknown/private keys, and serializes ascending exact lexicographic order. |
| `RecordChangeIntentInput` | `affected_views[]` | Non-empty set of `AffectedViewChange` | Unique by exact `view_schema_id`; Collaboration serializes ascending exact lexicographic order. |
| `AffectedViewChange` | `view_schema_id` | Non-empty base view-schema ID | Visible labels, saved-view IDs, and client-local state are forbidden. |
| `AffectedViewChange` | `change_kind` | Exactly `patch`, `invalidate`, or `remove` | No default and no insert-like v1 value. |
| `AffectedViewChange` | `patch_cells` | `view_row_patch_v1` or absent | Present if and only if `change_kind='patch'`; an unsafe or incomplete patch MUST become `invalidate`. |

The source owner MUST derive both fact sets from accepted normalized mutation
operations and declared deterministic side effects. It MUST NOT discover them
by comparing arbitrary full-row JSON, projection rows, portable snapshots,
database relations, or client-local state. Revisions MUST NOT reconstruct
private facts from a public Collaboration input. Collaboration MUST NOT receive
conflict classes, private source columns, private before/after values, or
relation names.

`changed_field_keys[]` MAY be empty only for an owner-declared effect with no
public field change; soft-delete and restore are the current required cases.
Every ordinary current-profile update MUST supply a non-empty set. An included
patch cell with `{ "value": null }` is an authoritative clear when admitted by
the field contract; omission means unchanged.

### 3.3 Required transaction sequence

| Step | Required action | Failure result |
| --- | --- | --- |
| 1 | Authenticate the actor and authorize the requested mutation against current state. | Reject without source, revision, projection, or intent effects. |
| 2 | Normalize and validate the submitted source-owner operations. | Reject without acquiring publication state. |
| 3 | Acquire required source record/version locks and revalidate the mutation. | Roll back the borrowed transaction on failure. |
| 4 | Derive the exact private and public field identities plus declared deterministic side effects. | Fail closed; no generic row diff fallback. |
| 5 | Apply the authoritative source mutation. | Roll back all transaction-local effects on failure. |
| 6 | Append the live revision and private conflict facts through Revisions. | Roll back the source mutation and every prior effect. |
| 7 | Materialize or refresh the required source-owned projection/view effects and construct safe affected-view facts. | Roll back; never guess a partial patch. |
| 8 | Append exactly one Collaboration semantic intent through the Collaboration-owned capability. | Roll back source, revision, private facts, projections, and idempotency success. |
| 9 | Commit the borrowed transaction exactly once. | A failed or commit-unknown outcome follows the existing transaction-owner policy; no early WebSocket delivery occurs. |
| 10 | After proven commit, make the intent available for stream sequencing and currently authorized delivery. | Delivery failure follows existing retry/quarantine behavior and does not undo committed source state. |

Exact idempotent replay MUST return the original result and MUST NOT append a
second revision or Collaboration intent. A rejected conflict-resolution
request MUST leave source state, revision facts, projection effects,
idempotency success, and Collaboration intents absent.

### 3.4 Historical path and pre-production default

Historical incident-bundle import is a separate variant. It MAY import
owner-valid historical revisions and canonical snapshots, but it MUST NOT
synthesize live conflict facts or append a live `record_changed` intent.
Transaction-local publication suppression MUST NOT leak after commit or
rollback.

Because this repository is pre-production, an incompatible pre-existing
private-fact shape MUST use the adopted reset-required diagnostic and database
reset. No backfill, inference from old snapshots, dual reader, shadow write,
feature flag, or compatibility alias is allowed.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /ws/v1/incidents/{incident_id}` route | Core 01; security consequences in Core 04 | `routes.go`, server route catalog, WS schema | Socket and integration tests | Preserve exact method/path, upgrade failures, and route registration through every slice | critical | Target-local harness inventory currently omits `/v1`. |
| WebSocket envelope, 32,768-byte limit, text-only semantics, duplicate members, and close behavior | Core 01 | `codec.go`, `routes.go` | Codec and socket frame tests | Preserve additive known-message members, malformed/duplicate handling, UTF-8, binary rejection, and close codes | critical | Concrete vendor adapter stays outside the module. |
| Client message vocabulary | Core 01 | WS schema and codec registry | Codec/socket tests | `hello`, `resume`, `pong`, `presence_update`; first-message and later-message closed grammar | high | Unknown message types remain invalid even while known messages tolerate additive members. |
| Server message vocabulary | Core 01 | WS schema, protocol DTOs | Protocol, hub, socket, and integration tests | All ten current server families, required envelopes, replayable versus ephemeral classification | critical | Generated TypeScript is an observable consumer surface. |
| Presence identity, validation, canonical ordering, TTL, snapshot/delta, and incident isolation | Core 01; UI interpretation in Core 03 | `protocol.go`, routes, frontend coordinator | Hub, socket, integration, frontend unit/browser tests | Expiry, exact connection ordering, extension workspace shape, local visibility and removal | critical | Presence is ephemeral and not revision authority. |
| Resume-token binding, high-water sequencing, replay/reset, retention, and gap handling | Core 01/Core 03 | `stream.go`, routes, frontend session | Socket, stream integration, frontend session, browser tests | Session/incident/client binding, expiry, too-old/future reset, replay order, duplicate/gap handling | critical | Preserve no partial replay on reset. |
| `record_changed` envelope, field/view canonicalization, patch/invalidate/remove | Core 01; client application in Core 03 | Intent builder, protocol conversion, frontend coordinator | Intent, hub, integration, frontend, browser tests | Canonical arrays, no duplicates, null/omission, replay/live equality, source-side atomicity, pending-txn correlation | critical | `client_txn_id` is current observable behavior and remains frozen. |
| `extension_resource_changed` | Core 01 generic envelope; Extensions and Network Flow for owner semantics | WS schema, protocol validator, server translator, frontend interpreter | Protocol, intent, Network Flow integration/frontend tests | Additive-member decoder tolerance, stable IDs, claimed profile, allowed reason/change pairs, no labels/diagnostics | high | `SL-02` requires later authorization. |
| Incident-scoped `job_progress` | Core 01 shared job resource semantics | WS schema, protocol validator, server translator | Protocol, intent, jobs and frontend tests | Incident scope match, exact statuses/progress, no deployment event, additive-member decoder tolerance | high | Generated frontend validator currently rejects additive payload members. |
| Same-transaction event intent and rollback behavior | Core 01 | `stream.go`, Revisions/source-owner callers | Stream and many source-owner integration tests | Exactly one deterministic intent per committed effect; no intent on rollback or exact replay | critical | Source owners retain mutation semantics. |
| Historical-intent suppression | Core 01 | `historical_restore.go`, Revisions integration | Historical suppression and restore/import tests | Transaction-local setting across commit, rollback, and unrelated transactions | high | No global or session-leaking switch. |
| Current authorization and terminal session effects | Core 04 with Core 01 route behavior | Routes, admission checker, incident notifier | Revocation/closure/heartbeat/socket/browser tests | Membership removal, role/session loss, incident close, no deployment-admin bypass, no unauthorized delivery | critical | Authorization is re-derived, not cached by replay tokens. |
| Collaboration stream tables and storage semantics | Collaboration source ownership under Core 01/Core 03 | Migration 00029, schema ownership manifest, `stream.go` | Stream integration, recovery catalog, PostgreSQL role tests | Atomic append/sequence, cursor locking, token hash/binding, quarantine and retention | critical | Four tables: intents, cursors, replay events, resume tokens. |
| Deployment-local requeue CLI and result v2 | Core 01 grammar/result, Core 03 transition, Core 04 authority | `recovery.go`, operator app, `contracts/collaboration/**` | Operator unit/process and stream requeue tests | Closed grammar, repaired-state proof, concurrency, audit, timeout/cancel/commit-unknown mapping, no network surface | critical | No HTTP, browser, WS, job, session, bearer, or public-audit surface. |
| Recovery-state contribution | Core 01 recovery mechanics with Collaboration source ownership | `recovery_state.go`, recovery assembly, recovery fixture | Recovery catalog tests | Exact four tables, recovery classification, invalidation algorithm | high | Generated recovery outputs must be regenerated, never hand-edited. |
| Frontend session/coordinator behavior | Core 03 client behavior over Core 01 wire | Incident session, Workbook coordinator, Network Flow interpreter | Vitest and browser owner rows | Reconnect, reset completion, protected-state invalidation, presence, active-surface refresh, pending edit anchoring | critical | Out of target but must be included in contract regression. |
| Collaboration telemetry vocabulary | Adopted OpenTelemetry NLSpec | Hub/lifecycle telemetry files | Telemetry unit and conformance tests | Safe bounded event/result/drop/operation vocabularies and active connection gauge | medium | No raw identifiers in telemetry attributes. |
| Generated Go/TypeScript Collaboration contract | Core 01 owner projected through `contracts/ws` | WS schema, generated Go artifact, protocol-ts types/validators | Protocol contract tests and `package.protocol_ts` rows | Source schema and generated outputs agree with owner; drift checks pass | critical | Generated roots must not be hand-edited. |
| Harness/test accounting | Testing Harness NLSpec for mechanics; product owners for behavior | verification contract, test catalog, 33-row family manifest | Catalog/harness checks and shared harness tests | Exact active selectors, valid evidence references, correct route labels and owner collaboration | high | Rows are verification accounting, not runtime architecture. |

### 4.1 `/ws/v1/` decoder outcome map

This map specifies the proposed `SL-02` correction. It does not authorize that
correction. `RB-003` MUST be cleared first.

| Input condition | Required decoder result | Downstream effect |
| --- | --- | --- |
| Known replayable message with every required member valid and an unknown additive scalar member | Accept. | Exclude the unknown member from the typed result; it does not affect dispatch or state. |
| Known replayable message with an unknown additive object or array member | Accept without behavior driven by the unknown subtree. | Exclude the member from the typed result. |
| Unknown additive member inside `view_row_v1`, `view_row_patch_v1`, or a cell object where its owner admits additive members | Accept at that object boundary. | Preserve only known typed members. |
| Required known member omitted | Reject. | Preserve the existing malformed-message failure path. |
| Required known member has the wrong type or an invalid closed token | Reject. | Preserve the existing malformed-message failure path. |
| Unknown top-level message `type` | Reject. | Do not dispatch through a fallback family. |
| Duplicate known JSON member | Reject at the raw byte-to-message boundary. | Ordinary parsed-object validation is insufficient evidence. |
| Duplicate unknown JSON member | Reject at the raw byte-to-message boundary. | Ordinary parsed-object validation is insufficient evidence. |
| Invalid UTF-8, binary frame, malformed JSON, or oversized message | Preserve the current rejection and close behavior. | No typed message reaches a consumer. |
| Unknown member resembles a route, action, discriminator, resource, or telemetry field outside its owning position | Accept only where additive, then ignore and exclude it. | It MUST NOT alter dispatch, navigation, persistence, telemetry, UI state, or rendered content. |
| Extra member in a deliberately closed client request/control object | Reject. | The correction MUST NOT globally open all JSON objects. |

The affected replayable families are `record_changed`, `job_progress`, and
`extension_resource_changed`. The current authored schema already leaves the
`record_changed` payload open, but all three families require positive
characterization. The current `job_progress` and
`extension_resource_changed` payload projections contain closed object levels
that reject owner-admitted additive input.

### 4.2 Recorded authorization for `WF-SL-02`

The 2026-08-24 remediation request authorizes the following correction after
`WF-SL-01` completes:

> **Authorized behavior correction - Collaboration replayable decoder v1.**
> Update the authored `/ws/v1/` contract projection and generated Protocol
> TypeScript decoder so known replayable server message families accept unknown
> additive object members at every object level that its adopted owner declares
> additive. The affected families are `record_changed`, `job_progress`, and
> `extension_resource_changed`. Preserve required known members and types, the
> closed message-type vocabulary, raw duplicate-member rejection, UTF-8 and
> text-frame rules, size limits, unknown-message-type rejection, authorization,
> and dispatch behavior. Ignore admitted unknown members, exclude them from the
> typed result, and prevent them from affecting dispatch, application state,
> persistence, domain mutation, telemetry, navigation, or rendering. Update
> authored schema and tests first and regenerate only through Make. This is a
> v1 conformance correction. It does not authorize a new route, message type,
> schema family, compatibility decoder, dual reader, or `/ws/v2/` root.

The authorized workstream, allowed authored and generated scopes, tests, Make
generation path, rollback rule, and frontend-consumer characterization are
recorded in Sections 1.3, 4.3, and 7. Generated roots remain Make-owned.

### 4.3 Decoder characterization requirements

| ID | Characterization | Required evidence |
| --- | --- | --- |
| `DEC-AC-01` | Additive scalar, object, and array members on each replayable envelope and payload are accepted. | Protocol TypeScript decoder tests for all three families. |
| `DEC-AC-02` | Additive nested members are accepted only at owner-declared additive row, patch, cell, scope, progress, summary, or workspace boundaries. | Per-object positive and deliberately closed negative fixtures. |
| `DEC-AC-03` | Missing required members, wrong types, and unknown message types remain rejected. | Existing negative tests plus one case per affected family. |
| `DEC-AC-04` | Duplicate known and unknown members are rejected before ordinary parsed-object validation. | Raw WebSocket byte-to-message characterization. |
| `DEC-AC-05` | Unknown discriminator-like members do not alter the typed value, dispatch target, UI state, persistence, or telemetry classification. | Decoder and frontend consumer assertions. |
| `DEC-AC-06` | Backend intent validation continues to accept the same owner-valid additive fixtures. | `module.collaboration` characterization. |
| `DEC-AC-07` | Authored schema, generated Go/TypeScript, and checked-in artifact policy agree. | `make generate-drift`, policy, and JSON-shape results. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Authored WS payload schemas close `job_progress` and `extension_resource_changed`, and generated frontend validators reject additive payload members, while Core 01 REQ-01-257 requires `/ws/v1/` clients to ignore unknown additive members. | `contracts/ws/index.schema.json`; generated protocol-ts validator; protocol-ts tests; intent validation tests | A current client can discard an owner-valid future additive event. | `must_fix` | Core 01 WS projection and protocol-ts generated surface | `WF-SL-02`; update authored projection, characterize client acceptance, and regenerate through Make under the recorded authorization. |
| Socket harness inventory names `GET /ws/incidents/{incident_id}`. | `testsupport/incidentwstest/inventory.go` versus live route/schema | Evidence accounting can describe the wrong public route. | `must_fix` | Collaboration test accounting | `SL-01`; repair authored inventory and validate it. |
| Socket lifecycle evidence index names nonexistent replay/reset hub subtests and checks only that strings contain `::Test`. | `shared_harness_test.go` versus actual functions/subtests in `hub_test.go` | Apparent evidence coverage can remain green after evidence disappears. | `must_fix` | Collaboration tests and harness accounting | `SL-01`; point to live evidence and add reference validation or remove the prose-only index in favor of owner rows. |
| `stream.go` combines intent writing, resume storage, dispatcher scheduling, retry/quarantine, LISTEN, tailing, and retention. | Exact source inspection; 1,032 lines and multiple public constructors | High change blast radius and difficult isolated testing. | `should_fix` | Collaboration private stream runtime and persistence subpackages | `SL-04`; split behind frozen public capabilities without moving table meaning to Platform. |
| Root Collaboration surface exposes persistence/runtime constructors alongside protocol DTOs and route facade. | `stream.go`, `routes.go`, `protocol.go`, server assembly | Callers can couple to implementation breadth instead of a thin application boundary. | `should_fix` | Collaboration facade plus application composition | `SL-03` then `SL-04`; retain atomic caller cutover and no aliases. |
| Requeue direct SQL, administrative audit, and recovery-state adapter live beside WebSocket protocol code. | `recovery.go`, `recovery_state.go` | Recovery changes can disturb the flat package and obscure local-authority boundaries. | `should_fix` | Collaboration recovery capability with private adapters | `SL-06`; freeze CLI/result/security semantics. |
| `ChangedCellKeys` is used both for Collaboration wire consequences and Revisions private conflict facts. | Revisions appender files and source-owner callers | A Collaboration import makes a public-wire helper the authority for private conflict calculation; a careless move could alter both semantics. | `must_fix` | Separate source-owner fact inputs, Revisions private persistence, and Collaboration public construction | Use adopted REQ-00-075, `CT-REQ-003`, and Sections 3.1-3.4 in `WF-SL-05`. |
| The prior Revisions decision assigned Collaboration publication to Revisions. | Pre-remediation `docs/decisions/revisions-module-boundary.md` and `temp/analysis-notes.md` | Leaving the mismatch unresolved would let implementation or research silently choose ownership. | `resolved_specification` | Revisions and Collaboration boundary decisions | `WF-SPEC-01` amended Revisions and adopted the independent-port boundary through REQ-00-075; implementation remains sequenced to `WF-SL-05`. |
| Source owners construct Collaboration record-change inputs and append through `IntentAppender`. | Timeline assembly, Revisions, Evidence, Entity Mentions/Merge | Broad input types can expose more wire structure than each owner needs, but same-transaction ownership is correct. | `should_fix` | Consumer-owned Collaboration port with source-owned adapters in assembly/owners | Include port narrowing in `SL-05` only after characterization. |
| Reusable socket/scenario helpers are split between target-local `testsupport` and `internal/testutil`. | Cross-module imports; `scenariotest` wraps `appsupport`; existing `internal/testutil/collaborationsupport` | Duplicated composition and unclear test fixture ownership. | `should_fix` | `internal/testutil` for reusable capabilities; Collaboration for semantic assertions | `SL-07`; move only after selector/import audit. |
| Collaboration routes import platform auth/HTTP primitives but not concrete WebSocket vendor code. | `routes.go`, `codec.go`, server socket adapter, backend boundary policy | Route orchestration is transport-adjacent; moving semantics into Platform would invert ownership. | `intentional/no_action` | Collaboration semantics; Platform transport primitives | Preserve the existing semantic/transport split while making route internals private. |
| Collaboration owns semantic telemetry while Platform supplies telemetry APIs. | Adopted telemetry NLSpec and telemetry files | Moving signal semantics to Platform would erase owner vocabulary. | `intentional/no_action` | Collaboration | Keep signal definitions; only infrastructure remains Platform-owned. |
| No grid-vendor imports or tokens occur under the backend target. | Target and frontend-boundary search | Inventing a backend grid-adapter workstream would create unsupported scope. | `intentional/no_action` | Existing frontend/grid adapter owners | Record the negative finding; no slice. |
| No saved-view persistence or view-schema registry ownership was found in the target; only stable sheet/view identities are consumed. | Protocol presence and record-change shapes; source-owner callers | Misclassifying wire identities as registry ownership could move behavior incorrectly. | `intentional/no_action` | Saved Views/View Schemas remain their current owners; Collaboration owns only wire usage | Freeze identities; do not move registries into Collaboration. |
| Generated contract and recovery artifacts are downstream of authored owners. | Generated artifact policy, WS/recovery generated files | Hand edits would drift or be overwritten. | `intentional/no_action` | Authored contracts plus Make generation | Every generated change must start at its owner input and use `make generate`. |
| Test-only direct SQL exists in owner storage helpers and cross-owner assertions. | `testsupport/intenttest`, source-owner integration tests | Tests can accidentally define production ownership. | `defer` | Collaboration owner storage tests and narrow shared assertions | Audit during `SL-07`; retain only behavior evidence required by owner tests. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish baseline, authority, allowed write, and handoff state. | This tracker and inspected owner documents | `make lint-markdown` for the tracker task | Baseline commit, status, time, authorities, and commands recorded. |
| WF-01 | Complete target inventory | chain | WF-00 | WF-02, WF-04 | Account for all 28 files and live package/caller surfaces. | `internal/modules/collaboration/**` plus direct callers | Read-only search and exact source inspection | Every target file has one inventory row. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-03, WF-05 | Bind observable contracts to adopted owners and projections. | Core owner sections, `contracts/ws`, Collaboration contracts, frontend consumers | Contract/test inventory review | Every contract risk has owner, evidence, and test posture. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-05, WF-07 | Identify missing behavior characterization and stale evidence. | Collaboration tests, protocol-ts tests, owner family manifests | `make task-guide` and `make explain-test-owner` discovery | Additive decoder and harness-evidence gaps recorded. |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-05 | Classify transport, persistence, source-owner, testutil, frontend, and generated coupling. | Target, app assembly, direct production callers, boundary policies | `make backend-module-boundary-check` and `make frontend-import-boundary-check` during implementation | Findings classified as `must_fix`, `should_fix`, `defer`, or `intentional/no_action`. |
| WF-05 | Facade and ownership redesign | chain | WF-02, WF-03, WF-04 | WF-06 | Adopt the fact-separation and port topology, plus the thin facade/private sub-boundaries. | Collaboration root, Revisions boundary, candidate private packages, source-owner adapters, application assembly | Owner review, owner slices, and boundary checks | Completed by `WF-SPEC-01`: REQ-00-075 and both boundary decisions adopt Sections 3.1-3.4. |
| WF-06 | Behavior-preserving slice sequencing | chain | WF-05 | WF-07 | Execute the smallest independently reversible cutovers without dual paths. | Files named by SL-01 through SL-08 | Slice-specific commands in Section 7 | Each slice has a clean diff, passing narrow checks, and recorded rollback point. |
| WF-07 | Harness/test/accounting update | chain | WF-03, WF-06 | WF-08 | Keep selectors, owner rows, evidence, and generated projections aligned with moved tests/contracts. | Test-family owner inputs, authored harness inventory, generated projections via Make | `make harness-contract`, generation drift, owner slices | No stale path/selector/reference; generated files changed only through Make. |
| WF-08 | Validation and final handoff | chain | WF-07 | none | Run finalization, broad verification, inspect residual imports, and close the tracker. | Repository-wide verification surfaces and this tracker | `make agent-finalize`, `make check` | Run roots/results, skipped checks, residual risks, and binary completion recorded. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-01 | WF-03 | Repair the target-local route label and stale socket evidence references; make evidence references mechanically correspond to live tests or rely on the authored owner rows. No runtime change. | `shared_harness_test.go`, `testsupport/incidentwstest/inventory.go`, authored harness/test-family inputs only if selectors change | Harness accounting only; must not create product behavior | Preserve all socket tests; add a failing case for nonexistent evidence reference if the local index remains | `make test-slice OWNER=module.collaboration`; `make harness-contract` | Revert this accounting-only slice; no production rollback needed. | Inventory says `/ws/v1/`, every evidence reference resolves, and owner/harness checks pass. |
| SL-02 | SL-01 | Implement the authorized decoder outcome map in Section 4.1, update authored WS projections and characterization first, then regenerate through Make. | `contracts/ws/index.schema.json`, protocol-ts authored tests/entrypoint as needed, generated Go/TS outputs via `make generate` | Changes accepted client input; MUST preserve required members, closed message types, raw duplicate rejection, framing, size, and non-execution of unknown content | `DEC-AC-01` through `DEC-AC-07`; preserve malformed, unknown-type, handshake, and backend intent tests | After authored edits: `make generate`; then `make test-slice OWNER=package.protocol_ts`; `make test-slice OWNER=module.collaboration`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Revert authored schema/tests and regenerate; never revert generated files independently. | Every `DEC-AC-*` row passes, generated state is clean, and the authorization record is linked. |
| SL-03 | SL-01 | Establish a thin Collaboration application facade and move route/hub implementation details behind private boundaries while preserving existing public composition symbols during the atomic cutover. | `routes.go`, `codec.go`, `protocol.go`, server collaboration composition, candidate private Collaboration packages | Route, authorization, message, presence, hub, shutdown, and observer behavior | Preserve codec, hub, socket, integration, incident-effect, frontend and browser tests | `make test-slice OWNER=module.collaboration`; `make service-backed-test-slice OWNER=module.collaboration`; `make backend-module-boundary-check` | Revert the atomic facade cutover; do not retain forwarding aliases or two active registrars. | One server-facing registrar remains, concrete WebSocket stays in Platform, and no observable route/message behavior changes. |
| SL-04 | SL-03 | Separate durable stream coordination from private PostgreSQL persistence while retaining narrow intent, replay, and dispatcher capabilities. | `stream.go`, candidate private stream store/runtime packages, server assembly | Transaction atomicity, intent-key collisions, order, multi-process tailing, retries, quarantine, token binding, retention, shutdown | Preserve full `stream_integration_test.go`, intent validation, socket replay, and source-owner atomicity tests | `make service-backed-test-slice OWNER=module.collaboration`; `make backend-module-boundary-check` | Revert the storage/runtime cutover as one slice; schema and data remain unchanged. | Root facade no longer contains direct stream SQL, four-table semantics remain Collaboration-owned, and all stream tests pass. |
| SL-05 | SL-03 | Implement adopted REQ-00-075 and Sections 3.1-3.4: source owners provide separate explicit fact sets, Revisions persists private facts, Collaboration validates and appends public intent facts, and `ChangedCellKeys` is removed atomically. | `record_change_intent.go`, Revisions appender and boundary files, Timeline assembly, Evidence, Entity Mentions/Merge, source-owner adapters, application assembly | Field identity, conflict reconstruction, affected-view ordering, patch/null semantics, historical suppression, transaction order, and exactly-one intent | `CT-AC-01` through `CT-AC-10`; preserve record-change, Revisions conflict, source-owner mutation, frontend reconciliation, and browser conflict coverage | `make test-slice OWNER=module.collaboration`; `make service-backed-test-slice OWNER=module.collaboration`; `make test-slice OWNER=module.revisions`; `make test-slice OWNER=module.timeline`; `make service-backed-test-slice OWNER=module.timeline`; `make test-slice OWNER=module.evidence`; `make service-backed-test-slice OWNER=module.evidence`; `make backend-module-boundary-check` | Revert the complete fact/port cutover; no source owner may remain on the old helper and no compatibility export may survive. | The adopted topology is implemented, all known callers are characterized, every `CT-AC-*` row passes, and no shared semantic diff or circular dependency remains. |
| SL-06 | SL-04 | Hide requeue SQL/audit and recovery-state mechanics behind stable Collaboration-owned recovery capabilities. | `recovery.go`, `recovery_state.go`, operator adapter, recovery assembly, candidate private persistence adapter | CLI grammar/result v2, local authority, repair proof, locking, journal, timeout/cancel/commit-unknown, recovery catalog | Preserve requeue stream tests, operator unit/process tests, recovery catalog and PostgreSQL role tests | `make test-slice OWNER=app.operator`; `make service-backed-test-slice OWNER=app.operator`; `make service-backed-test-slice OWNER=module.collaboration` | Revert the recovery adapter cutover; no migration or contract version change is permitted. | Public operator contract is byte/exit compatible, source-owner contribution is unchanged, and root recovery API is narrow. |
| SL-07 | SL-03, SL-04 | Consolidate reusable socket/scenario capabilities under `internal/testutil`; retain owner-semantic storage/protocol helpers locally where justified; update authored selectors/imports before regeneration. | `testsupport/**`, `internal/testutil/collaborationsupport`, importing test packages, test-family owner inputs | Test behavior, selectors, fixture ownership, and evidence routing only | Preserve all Collaboration and cross-owner tests that use the helpers | `make test-slice OWNER=module.collaboration`; `make service-backed-test-slice OWNER=module.collaboration`; `make harness-contract`; `make generate-drift` if projections change | Revert helper moves and selector owner inputs together; do not hand-edit generated topology. | No duplicated generic runtime wrapper remains, all imports/selectors resolve, and semantic helpers have an explicit owner. |
| SL-08 | SL-02, SL-05, SL-06, SL-07 | Atomically cut over remaining callers, remove transitional exports/paths, run finalization and broad verification, and record evidence. No compatibility aliases, fallback readers, or dual paths. | Collaboration root, app assembly, remaining direct callers, authored boundary/harness inputs, this tracker | Entire frozen contract map | Preserve all owner, protocol-ts, operator, frontend, browser, security, recovery, and generation coverage | `make agent-finalize`; `make check` | Revert the final cleanup if any caller or broad gate fails; earlier completed slices retain their own rollback points. | Old imports/exports are absent, generated state is clean, broad checks pass, and the handoff records run roots and residual risks. |

Any discovered change beyond the authorized owner-mandated decoder repair and
the internal topology in Section 1.3 requires a new adopted owner or explicit
scope expansion before implementation.

`CT-REQ-002` applies to every slice. Temporary candidate types MAY exist only
inside the isolated implementation diff before its atomic cutover. No committed
slice may retain an old and new runtime path, compatibility alias, fallback
reader, shadow write, or feature flag.

### 7.1 Cross-slice binary acceptance matrix

| ID | Required outcome | Applies to | Evidence posture | Failure meaning |
| --- | --- | --- | --- | --- |
| `CT-AC-01` | One successful live mutation commits source state, one live revision with private facts, required projection effects, and exactly one Collaboration intent in one transaction. | `SL-05` | Source-owner service-backed integration plus intent/revision counts. | Slice fails and is rolled back. |
| `CT-AC-02` | A forced failure at each boundary before commit leaves none of the source, revision, projection, idempotency-success, or intent effects. | `SL-04`, `SL-05` | Existing failure-injection integration rows extended for the new ports. | Transaction topology is non-conformant. |
| `CT-AC-03` | Exact idempotent replay returns the original result and creates no second revision or intent; divergent reuse preserves the current conflict result. | `SL-04`, `SL-05` | Source-owner and Collaboration store rows. | Slice fails and is rolled back. |
| `CT-AC-04` | A concurrent different-field edit retains existing auto-rebase behavior; a same-field edit retains the explicit conflict path. | `SL-05` | Revisions, Workbook/frontend, and browser conflict rows. | Private fact calculation changed behavior. |
| `CT-AC-05` | Private field values, conflict classifications, source columns, snapshots, and relation names appear in neither WebSocket messages nor portable bundles. | `SL-05` | Wire-shape, portability, and retained-secret assertions. | Boundary and disclosure failure. |
| `CT-AC-06` | Historical import creates no live conflict facts or Collaboration event, and suppression does not leak across commit, rollback, or a later live transaction. | `SL-05` | Historical suppression and incident-bundle integration rows. | Historical/live boundary failure. |
| `CT-AC-07` | Live and replay delivery serialize identical canonical `changed_field_keys[]` and `affected_views[]` order for semantically identical content. | `SL-04`, `SL-05` | Intent, stream replay, socket, and frontend rows. | Public wire conformance failure. |
| `CT-AC-08` | Delete, restore, sparse patch, authoritative null, remove, and invalidate behavior remains byte- and state-compatible with the Section 4 freeze map. | `SL-03`, `SL-05` | Protocol, hub, source-owner, frontend, and browser rows. | Observable behavior changed without authorization. |
| `CT-AC-09` | Revisions and Collaboration have no implementation import in either direction; no generic shared-diff package, `ChangedCellKeys` forwarding alias, dual path, or fallback reader remains. | `SL-05`, `SL-08` | Exact import/search audit and backend boundary check. | Cutover is incomplete. |
| `CT-AC-10` | Revisions, Timeline, Evidence, Entity Mentions, Entity Merge, rollback/restore, and application-assembly callers are characterized before the old export is removed. | `SL-05` | Caller inventory crosswalk to exact owner/test rows. | Removal is blocked until the missing caller has evidence. |

Every applicable acceptance row is binary. A passing proxy test, compilation,
or broad target does not substitute for a missing owner-routed row. Each result
MUST be recorded with the exact run root required by `CT-REQ-009`.

## 8. Validation Plan

Discovery established an active `module.collaboration` owner manifest with 33
rows: 12 no-service rows and 21 service-backed rows across Go, Vitest, and
Playwright. The commands below are repository-owned public Make targets. No
test, lint, generation, or conformance target had been executed before this
tracker was authored.

### 8.1 Required implementation baselines

`RB-002` required executable evidence rather than command discovery or
documentation review. Both named checkpoints are now established:

| Checkpoint | Exact point | Required commands | Disposition |
| --- | --- | --- | --- |
| B0 - pre-change baseline | Exact implementation-base SHA before any Collaboration product, contract, test, harness, or generated edit | `make task-guide ROLE=module-author OWNER=module.collaboration`; `make explain-test-owner OWNER=module.collaboration`; `make test-slice OWNER=module.collaboration`; `make service-backed-test-slice OWNER=module.collaboration` | Required before `SL-01`; a later repository advance invalidates an older B0. |
| B1 - post-accounting baseline | Exact source digest after `WF-SL-01` and before `WF-SL-02` through `WF-SL-08` | Repeat both owner-slice commands with the post-`WF-SL-01` catalog and selectors | **PASS:** final retained roots and source digest are recorded below. |

#### B0 retained execution record

| Field | B0 evidence |
| --- | --- |
| Workstream and gaps | `WF-EXEC-00`; enables `G-01` through `G-09` and closes the baseline portion of `G-02`. |
| Start and end | `2026-08-24T19:09:01-04:00` through `2026-08-24T19:16:09-04:00`. |
| Starting source | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`; exact status `A  docs/handoffs/collaboration-module-refactor-tracker.md`. |
| Execution source identity | The harness recorded source commit `169ba53197681aa45767914b70cd86d8759d0d3f`, `source_state=dirty`, and source digest `sha256:5f917cd41e415142ace66686976d3a2dc02a1c6fa6ea0a8c9f6178dec2333dae`. The dirtiness is the authorized tracker activation layered over the preserved staged tracker. |
| Selection | `make explain-test-owner OWNER=module.collaboration` reported 33 owner rows, 21 service-backed rows, and the authored manifest `tools/test_families/module.collaboration.json`. The no-service graph selected and passed 32 units; the service-backed graph selected and passed 23 units, including dependency/support work. Each graph reported complete terminal accounting. |
| Files changed | Only `docs/handoffs/collaboration-module-refactor-tracker.md`; status at evidence completion was `AM`, preserving the pre-existing staged addition and keeping execution edits unstaged. |
| Substantive result | Both owner slices and every required boundary/generated guard passed. No product, contract, test, harness, generated, or migration file changed. |
| Compatibility and migration | None. Tracker activation and retained evidence do not change runtime behavior or public/internal interfaces. |
| Rollback point | Revert only the unstaged tracker activation/checkpoint edits; the staged tracker content and repository implementation remain intact. |
| Residual risks | B0 does not validate the intended decoder correction or final topology. The harness `explain-run` summary reports graph identity but currently renders zero detailed work units; the authoritative Make graph summaries below contain complete pass counts. B1 remains required after the accounting-only workstream. |
| Binary exit | **PASS.** Owner and guard validations passed, the checkpoint was recorded, and Markdown lint passed. `WF-EXEC-00` is `DONE`. |

| Exact command | Stable command ID or discovery result | Result | Run root |
| --- | --- | --- | --- |
| `make task-guide ROLE=module-author OWNER=module.collaboration` | Focused owner commands resolved successfully. | Passed. | Non-retained discovery output. |
| `make explain-test-owner OWNER=module.collaboration` | 33 rows; 21 service-backed; 14 Go, 12 Playwright, and 7 Vitest runners. | Passed. | Non-retained discovery output. |
| `make test-slice OWNER=module.collaboration` | `cartulary.harness.command.test_slice.v2` | Passed, 32 of 32 units in 84.636 seconds; no non-pass classification. | `.cartulary/test-results/20260824T231017Z-p1047041` |
| `make service-backed-test-slice OWNER=module.collaboration` | `cartulary.harness.command.service_backed_test_slice.v2` | Passed, 23 of 23 units in 83.788 seconds; no non-pass classification. | `.cartulary/test-results/20260824T231148Z-p1094471` |
| `make backend-module-boundary-check` | `cartulary.harness.command.backend_module_boundary_check.v2` | Passed, 3 of 3 units. | `.cartulary/test-results/20260824T231323Z-p1141848` |
| `make generate-drift` | `cartulary.harness.command.generate_drift.v2` | Passed, 4 of 4 units. | `.cartulary/test-results/20260824T231328Z-p1142202` |
| `make generated-artifact-policy-check` | `cartulary.harness.command.generated_artifact_policy_check.v2` | Passed, 3 of 3 units. | `.cartulary/test-results/20260824T231339Z-p1145148` |
| `make json-shape-check` | `cartulary.harness.command.json_shape_check.v2` | Passed, 3 of 3 units. | `.cartulary/test-results/20260824T231343Z-p1145588` |
| `make lint-markdown` | Ad hoc Markdown lint runner. | Passed in 42.190 seconds. | `.cartulary/test-results/20260824T231517Z-p1147708` |

The final `DONE` status edit was revalidated by `make lint-markdown`; it passed
in 11.400 seconds with run root
`.cartulary/test-results/20260824T231633Z-p1148999`.

#### `WF-SPEC-01` execution record

| Field | Specification-adoption evidence |
| --- | --- |
| Workstream and gap | `WF-SPEC-01`; closes the adoption portion of `G-01` and establishes the authority required by `G-04` through `G-09`. |
| Start and end | `2026-08-24T19:17:08-04:00` through `2026-08-24T19:23:51-04:00`. |
| Starting source and status | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`; exact status `AM docs/handoffs/collaboration-module-refactor-tracker.md`. |
| Files changed | Added `docs/decisions/collaboration-module-boundary.md`; amended `docs/decisions/revisions-module-boundary.md`, `docs/domain.md`, `docs/spec/00_document_set_status_and_precedence.md`, `docs/spec/04_security_deployment_and_conformance.md`, `docs/spec/F_source_traceability_matrix.md`, and this tracker. |
| Requirements and result | Added REQ-00-075 and AC-564, adopted the independent-port architecture, removed Collaboration publication ownership from Revisions, made historical revision a distinct non-publication operation, fixed Collaboration's domain classification, and linked the acceptance criterion in the traceability matrix and Base claim manifest. |
| Commands before checkpoint lint | Exact `sed` and `rg` owner/ID/terminology audits; `git diff --check`; Git status and diff-stat inspection. All completed without a conflicting owner, duplicate ID, whitespace error, or unexpected path. |
| Compatibility and migration | Architecture/specification only. No public route, message, CLI, storage schema, portability version, authorization, telemetry, implementation, test, generated file, or persisted state changed. Internal callers will migrate atomically in later workstreams with no aliases. |
| Rollback point | Revert this workstream's seven documentation paths as one owner-adoption slice. Do not partially retain the Revisions amendment without the Collaboration decision and Core/traceability entries. B0 and tracker activation remain independently valid. |
| Residual risks | AC-564 is adopted but is not yet backed by the final boundary-policy and owner-routed implementation evidence; that is required in `WF-SL-08`. B1, decoder repair, facade, stream, facts, recovery, and test-support work remain sequenced. |
| Markdown validation | `make lint-markdown` passed in 11.430 seconds with run root `.cartulary/test-results/20260824T232336Z-p1152178`. |
| Binary exit | **PASS.** Owner texts agree, static review and Markdown lint pass, RB-001 is resolved, and `WF-SPEC-01` is `DONE`. |

#### `WF-SL-01` execution and B1 record

| Field | Accounting and B1 evidence |
| --- | --- |
| Workstream and gap | `WF-SL-01`; closes `G-02`. |
| Start and end | `2026-08-24T19:24:57-04:00` through `2026-08-24T19:36:33-04:00`. |
| Starting source and status | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`; seven specification/tracker paths from completed predecessors were present, with the tracker still `AM`. |
| Files changed | `internal/modules/collaboration/shared_harness_test.go`, `internal/modules/collaboration/testsupport/incidentwstest/inventory.go`, `tools/test_families/module.collaboration.json`, generated `tools/execution_topology_render_index.json`, and this tracker. |
| Requirements and result | Removed `TestSocketLifecycleEvidenceIndex`, which only shape-checked prose references and included nonexistent subtests; retained the mechanically unique socket capability inventory; corrected its route to `GET /ws/v1/incidents/{incident_id}`; removed the deleted selector; regenerated the topology render index through Make. |
| B1 source identity | Harness source commit `169ba53197681aa45767914b70cd86d8759d0d3f`, `source_state=dirty`, final source digest `sha256:f07488ebfbff12f6687a336cec01dd3fd030c01faefd78474fdf2289b1311184`; no-service graph digest `sha256:c1c95669d74d0ae1aaa80e73d717d0229d130e1afc4d3d821db841dbefd97655`; service-backed graph digest `sha256:8a9b65c6edcd85d45d24dcc8de10b9a81a90696753243a348e56d20b95d9f67c`. |
| Selection and terminal closure | The owner catalog remains 33 rows with 21 service-backed rows. Final B1 passed 32 of 32 no-service graph units and 23 of 23 service-backed graph units with no non-pass classification. The active selector references only `TestSocketEventInventoryCoverage`. |
| Compatibility and migration | Test accounting only. One non-behavioral test selector was removed atomically; runtime, public contracts, Go production APIs, database state, and generated product contracts are unchanged. |
| Rollback point | Revert the two test files, authored owner manifest, and Make-generated render index together. B0 and specification adoption remain independently valid. |
| Residual risks | B1 establishes behavior/accounting health but does not implement AC-564 or decoder acceptance. `make explain-run` still renders zero detailed units despite the authoritative graph pass counts; the retained graph summaries and manifests remain the execution evidence. |
| Markdown validation | `make lint-markdown` passed in 10.970 seconds with run root `.cartulary/test-results/20260824T233618Z-p1404765`. |
| Binary exit | **PASS.** Owner slices, harness contract, generation drift, generated artifact policy, and Markdown lint pass; `WF-SL-01` is `DONE`. |

| Exact command | Result | Run root or disposition |
| --- | --- | --- |
| Initial `make test-slice OWNER=module.collaboration` | Passed 32 of 32 units after the authored accounting edit. | `.cartulary/test-results/20260824T232542Z-p1154682` |
| Initial `make harness-contract` | Passed 2 of 2 units. | `.cartulary/test-results/20260824T232712Z-p1203153` |
| Pre-generation B1 owner reruns | No-service passed 32 of 32 and service-backed passed 23 of 23, then were superseded when the generated topology index changed. | `.cartulary/test-results/20260824T232731Z-p1203718`; `.cartulary/test-results/20260824T232900Z-p1251074` |
| `make generate-drift` before regeneration | Failed 3 of 4 units because `tools/execution_topology_render_index.json` retained the old authored-manifest digest. Classified as expected generated harness-projection drift caused by this slice, not a product failure. | `.cartulary/test-results/20260824T233029Z-p1298440` |
| `make generate` | Passed and changed only the expected generated topology render index. | `.cartulary/test-results/20260824T233053Z-p1301667` |
| Final `make generate-drift` | Passed 4 of 4 units. | `.cartulary/test-results/20260824T233116Z-p1304529` |
| `make generated-artifact-policy-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260824T233129Z-p1307482` |
| Final `make harness-contract` | Passed 2 of 2 units. | `.cartulary/test-results/20260824T233133Z-p1307979` |
| Final B1 `make test-slice OWNER=module.collaboration` | Passed 32 of 32 units in 82.990 seconds; no non-pass classification. | `.cartulary/test-results/20260824T233150Z-p1308506` |
| Final B1 `make service-backed-test-slice OWNER=module.collaboration` | Passed 23 of 23 units in 83.432 seconds; no non-pass classification. | `.cartulary/test-results/20260824T233325Z-p1355805` |

#### `WF-SL-02` execution record

| Field | WebSocket projection and decoder evidence |
| --- | --- |
| Workstream and gap | `WF-SL-02`; closes `G-03` and implements `DEC-AC-01` through `DEC-AC-07` under `CT-REQ-001`, `CT-REQ-002`, `CT-REQ-005`, `CT-REQ-006`, `CT-REQ-008`, `CT-REQ-009`, and `CT-REQ-010`. |
| Start and end | `2026-08-24T19:37:23-04:00` through `2026-08-24T20:14:00-04:00`. |
| Starting source and status | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`. The completed specification/accounting paths were dirty, the tracker remained `AM`, and no decoder, generated protocol, or frontend path had yet changed. |
| Files changed | Authored `contracts/ws/index.schema.json`, `tools/protocol-ts/generate-protocol-types.mjs`, `tools/contractgen/validation.go`, protocol-ts decoder/entrypoint/tests, frontend Collaboration consumers/tests, Collaboration protocol characterization, `tools/test_families/package.protocol_ts.json`, Make-generated Go/TypeScript contracts and topology index, and this tracker. Generated roots changed only through `make generate`. |
| Substantive result | The authored projection now covers the complete ten-family server envelope and required replayable payload shapes. Owner-declared additive boundaries validate additively, while a generated schema-derived projector returns a fresh known-member value, strips admitted unknown members, preserves free-form keyed-map values, and never mutates raw input. Server types exclude client control messages. The frontend consumes the generated discriminated `record_changed` type instead of a weak handwritten structural guard. Backend intent validation characterizes additive record, job, and extension payloads. |
| Acceptance closure | `DEC-AC-01` through `DEC-AC-03`, `DEC-AC-05`, and `DEC-AC-06` are covered by protocol-ts, compile, frontend, and backend tests. Existing raw-frame duplicate-key coverage continues to satisfy `DEC-AC-04`. Authored/generated drift, artifact policy, and JSON-shape passes satisfy `DEC-AC-07`. The final Collaboration service-backed slice includes the stateful reconnect, reset, invalidation, and multi-client browser rows. |
| Compatibility and migration | Known-valid messages remain equivalent. Owner-valid additive server members are now accepted but omitted from the decoded value; free-form map entries remain intact. Incomplete or wrong-family server messages are rejected, and client control families are no longer part of `IncidentStreamMessage`. The private `0.0.0` TypeScript package migrated atomically. No public route, message family, database schema, CLI, authorization precedence, or stored state changed. |
| Rollback point | Revert the authored WS schema, generator/contract validation, protocol-ts/frontend/Go test changes, and selector as one unit, then run `make generate`; generated outputs must not be reverted independently. Completed specification and B1 checkpoints remain valid. |
| Residual risks | No slice-local defect remains. Runtime facade, stream split, explicit fact ports, recovery isolation, test-support consolidation, final AC-564 enforcement, and broad validation remain sequenced. |
| Markdown validation | `make lint-markdown` passed in 42.610 seconds with run root `.cartulary/test-results/20260825T001317Z-p1690452`. |
| Binary exit | **PASS.** All implementation, behavior, generated-state, harness, and tracker validations pass; `WF-SL-02` is `DONE`. |

| Exact command | Result | Run root or disposition |
| --- | --- | --- |
| Three early `make generate` repair runs | Failed because contract validation did not yet admit standard `$defs`, a renamed protocol selector was stale, and the projection helper was initially scoped inside a generated template. Each was caused by this slice, repaired in authored inputs, and superseded by the final generation pass. | `.cartulary/test-results/20260824T235030Z-p1410905`; `.cartulary/test-results/20260824T235105Z-p1412913`; `.cartulary/test-results/20260824T235404Z-p1427447` |
| Final `make generate` | Passed; generated Go, TypeScript, and topology artifacts were refreshed from authored inputs. | `.cartulary/test-results/20260824T235435Z-p1429333` |
| Earlier frontend/compiler repair runs | Frontend typecheck exposed unsafe replayable narrowing and incomplete fixtures; frontend unit exposed an owner-invalid 60-second resume window; import-boundary check exposed a direct generated-type import; a later typecheck exposed lost TypeScript discrimination. All were slice-caused and repaired before final validation. | `.cartulary/test-results/20260824T235454Z-p1432226`; `.cartulary/test-results/20260824T235612Z-p1433972`; `.cartulary/test-results/20260825T000135Z-p1533008`; `.cartulary/test-results/20260825T000715Z-p1637795` |
| Earlier Collaboration owner repair run | Failed 31 of 32 because a stale characterization expected deployment-scoped jobs on the incident stream; repaired to the adopted incident-only scope. | `.cartulary/test-results/20260824T235838Z-p1480841` |
| `make format` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T000329Z-p1536191` |
| Final `make test-slice OWNER=package.protocol_ts` | Passed 7 of 7 units; no non-pass classification. | `.cartulary/test-results/20260825T000807Z-p1638860` |
| Final `make frontend-typecheck` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T000749Z-p1638391` |
| Final `make frontend-unit` | Passed 390 of 390 units. | `.cartulary/test-results/20260825T000820Z-p1644253` |
| Final `make frontend-import-boundary-check` | Passed 2 of 2 units after the adapter-layer import repair. | `.cartulary/test-results/20260825T000209Z-p1533595` |
| Final `make test-slice OWNER=module.collaboration` | Passed 32 of 32 units; no non-pass classification. | `.cartulary/test-results/20260825T000359Z-p1540447` |
| Final `make service-backed-test-slice OWNER=module.collaboration` | Passed 23 of 23 units; no non-pass classification. | `.cartulary/test-results/20260825T000534Z-p1590187` |
| Final `make generate-drift` | Passed 4 of 4 units. | `.cartulary/test-results/20260825T001015Z-p1685326` |
| Final `make generated-artifact-policy-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T001027Z-p1688256` |
| Final `make json-shape-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T001032Z-p1688690` |
| Final `make harness-contract` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T001040Z-p1689204` |

One malformed diagnostic invocation omitted the required `TARGET` argument,
and one targeted test invocation initially named a `web.workbook` row under
the Collaboration owner. Neither produced a product run root; both were
corrected immediately, and the correctly owned row passed 2 of 2 at
`.cartulary/test-results/20260825T000126Z-p1532416`.

The final `DONE` status edit was revalidated by `make lint-markdown`; it passed
in 11.260 seconds with run root
`.cartulary/test-results/20260825T001438Z-p1691822`.

#### `WF-SL-03` execution record

| Field | Runtime facade and protocol-boundary evidence |
| --- | --- |
| Workstream and gap | `WF-SL-03`; implements `G-04` under `CT-REQ-001`, `CT-REQ-002`, `CT-REQ-005`, `CT-REQ-006`, `CT-REQ-007`, and `CT-REQ-010`. |
| Start and end | `2026-08-24T20:15:26-04:00` through `2026-08-24T20:35:21-04:00`. |
| Starting source and status | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`; the completed specification, accounting, and decoder paths were dirty; the tracker remained `AM`; no facade or protocol-subpackage file existed. |
| Files changed | Added `internal/modules/collaboration/runtime.go` and `internal/modules/collaboration/protocol/{contracts,codec}.go`; removed the root codec and incident-notifier wrapper; made hub and dispatcher types/constructors owner-private; changed routes and telemetry to private orchestration; updated server composition/settings/socket transport, application observation, owner tests, and semantic socket-test imports; added a test-binary-only dispatcher constructor pending the stream split; updated this tracker. |
| Substantive result | Server assembly now constructs exactly one Collaboration runtime from borrowed PostgreSQL, semantic socket, clock, and version capabilities. The runtime owns live hub construction, intent capability, dispatcher lifecycle, Auth revocation, Incident terminal effects, and semantic test observation. `RegisterRoutes(runtime)` is the sole server registrar. Message/envelope/payload, codec, duplicate-member, close/error, and socket abstractions now live in `collaboration/protocol`; Platform remains the only concrete WebSocket-vendor owner. Root `Hub`, `Dispatcher`, `Service`, `Settings`, codec, socket, and notifier exports are gone. |
| Compatibility and migration | Internal Go imports and server observation migrated atomically. Public route, handshake, message family, close/error behavior, authorization precedence, telemetry vocabulary, database schema/state, CLI, and frontend contracts are unchanged. Borrowed PostgreSQL and socket resources are not closed by the Collaboration runtime; its private dispatcher is closed idempotently by the owning server runtime. |
| Rollback point | Revert the runtime/protocol package addition, root privacy changes, server assembly/settings/socket changes, and semantic test imports as one unit. Do not restore a concrete hub without also restoring its prior server lifecycle wiring. Decoder and B1 checkpoints remain independently valid. |
| Residual risks | The replay store and mixed stream implementation remain root exports until `WF-SL-04`; the test-binary-only dispatcher constructor exists solely for owner integration until test support is consolidated. Final boundary policy still must prohibit retired construction/import shapes in `WF-SL-08`. |
| Markdown validation | `make lint-markdown` passed in 11.080 seconds with run root `.cartulary/test-results/20260825T003556Z-p1914683`. |
| Binary exit | **PASS.** Facade, protocol, owner, browser/socket, app-server, boundary, and tracker validations pass; `WF-SL-03` is `DONE`. |

| Exact command | Result | Run root or disposition |
| --- | --- | --- |
| Early targeted Collaboration compile | Failed before test execution because external stream tests still referenced the retired dispatcher constructor; repaired with a test-binary-only private-constructor bridge. | `.cartulary/test-results/20260825T002012Z-p1699748` |
| Second targeted Collaboration compile | Failed because one mechanically migrated harness call acquired a duplicated package qualifier; repaired immediately. | `.cartulary/test-results/20260825T002146Z-p1710146` |
| Early app-server support row | Failed before execution because the settings test retained an unused Collaboration import; repaired immediately. | `.cartulary/test-results/20260825T002221Z-p1716312` |
| Protocol-move compile repairs | Two targeted runs failed before execution on remaining unqualified moved protocol constants; all diagnostics were slice-caused and repaired in the owner tests and route implementation. | `.cartulary/test-results/20260825T002758Z-p1723703`; `.cartulary/test-results/20260825T002819Z-p1728043` |
| Final targeted Collaboration semantic row | Passed 1 of 1. | `.cartulary/test-results/20260825T002843Z-p1732340` |
| Final targeted app-server support row | Passed 1 of 1. | `.cartulary/test-results/20260825T002855Z-p1734174` |
| Final `make test-slice OWNER=module.collaboration` | Passed 32 of 32 units; no non-pass classification. | `.cartulary/test-results/20260825T002909Z-p1735715` |
| Final `make service-backed-test-slice OWNER=module.collaboration` | Passed 23 of 23 units; no non-pass classification. | `.cartulary/test-results/20260825T003059Z-p1785723` |
| Final `make test-slice OWNER=app.server` | Passed 24 of 24 units; no non-pass classification. | `.cartulary/test-results/20260825T003232Z-p1833241` |
| Final `make service-backed-test-slice OWNER=app.server` | Passed 17 of 17 units; no non-pass classification. | `.cartulary/test-results/20260825T003336Z-p1874011` |
| Final `make backend-module-boundary-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T003052Z-p1785341` |
| Final `make format` before owner validation | Passed 2 of 2 units. | `.cartulary/test-results/20260825T002839Z-p1728511` |

One attempted targeted invocation used a guessed rather than owner-catalog row
ID, and one diagnostic invocation used unsupported `DETAIL=artifacts`; both
failed as non-product usage errors without retained product run roots. The
catalog row was then selected exactly and retained above.

#### `WF-SL-04` execution record

| Field | Durable stream split evidence |
| --- | --- |
| Workstream and gap | `WF-SL-04`; implements `G-05` and the stream portions of `CT-AC-02`, `CT-AC-03`, and `CT-AC-07` under `CT-REQ-001`, `CT-REQ-002`, `CT-REQ-004`, and `CT-REQ-010`. |
| Start and end | `2026-08-24T20:37:09-04:00` through `2026-08-24T20:52:19-04:00`. |
| Starting source and status | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`; completed predecessor paths were dirty; the tracker remained `AM`; durable append/replay and dispatcher SQL/process logic still shared the root `stream.go`. |
| Files changed | Replaced root `stream.go` with the narrow temporary publication facade `stream_facade.go`; added private `internal/stream/store.go`, `dispatcher.go`, `dispatcher_telemetry.go`, and `record_payload.go`; updated runtime/routes to compose one private PostgreSQL stream and dispatcher; migrated owner stream tests to private storage construction; updated the recovery verifier to use private payload validation pending recovery isolation; updated this tracker. |
| Substantive result | `PostgresStream` exclusively owns append/idempotency, stream cursor/sequence, replay token/replay, retry/quarantine persistence, and retention SQL. `Dispatcher` owns lifecycle, LISTEN notification waiting, bounded scheduling, exponential backoff, tail cursor state, delivery, and retention cadence, and invokes the store through narrow methods. Runtime composes one instance of each. Root production code exposes no replay store or dispatcher type/constructor and contains no live stream persistence SQL. The dispatcher file contains no persistence query; its sole SQL command is the owner-required PostgreSQL `LISTEN`. |
| Compatibility and migration | No database table, column, index, intent key, canonical payload, sequence order, resume token, replay/reset, retry/quarantine, retention, or public wire behavior changed. Existing generic intent callers remain temporarily source-compatible only through the root facade and are removed atomically in `WF-SL-05`; there is no persisted-state migration or dual data path. |
| Rollback point | Revert the private stream package, root facade, runtime/route composition, recovery validation import, and stream-test construction together, restoring the prior monolithic root stream file. Do not retain both stream implementations. Predecessor facade/protocol work remains independently valid. |
| Residual risks | The broad `EventIntent` facade and stateless intent constructor intentionally survive only until `WF-SL-05`. Recovery SQL remains root-local until `WF-SL-06`. The test-binary dispatcher bridge remains until shared test support is consolidated in `WF-SL-07`. |
| Markdown validation | `make lint-markdown` passed in 11.350 seconds with run root `.cartulary/test-results/20260825T005304Z-p2133418`. |
| Binary exit | **PASS.** Complete no-service and service-backed Collaboration graphs, stream behavior, boundary policy, formatting, structural audits, and tracker validation pass; `WF-SL-04` is `DONE`. |

| Exact command | Result | Run root or disposition |
| --- | --- | --- |
| First post-move targeted compile | Failed before execution on two unused imports left by the physical dispatcher extraction; repaired immediately. | `.cartulary/test-results/20260825T004307Z-p1922747` |
| Second post-move targeted compile | Failed before execution because root recovery validation still referenced the retired root pending-intent representation; repaired by calling private stream payload validation. | `.cartulary/test-results/20260825T004331Z-p1927050` |
| Preliminary full owner passes | Passed 32 of 32 and 23 of 23, then were superseded by final runs after the boundary-policy naming repair. | `.cartulary/test-results/20260825T004437Z-p1933401`; `.cartulary/test-results/20260825T004612Z-p1982222` |
| First `make backend-module-boundary-check` | Failed 2 of 3 because existing final-topology tokens correctly rejected generic private names `Store`, `NewStore`, and a free `AppendIntentTx`; repaired structurally as `PostgresStream`, `NewPostgresStream`, and a store method rather than weakening policy. | `.cartulary/test-results/20260825T004747Z-p2029675` |
| Final `make format` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T004838Z-p2030450` |
| Final `make backend-module-boundary-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T004842Z-p2034315` |
| Final targeted Collaboration semantic row | Passed 1 of 1. | `.cartulary/test-results/20260825T004848Z-p2034713` |
| Final `make test-slice OWNER=module.collaboration` | Passed 32 of 32 units; no non-pass classification. | `.cartulary/test-results/20260825T004901Z-p2036536` |
| Final `make service-backed-test-slice OWNER=module.collaboration` | Passed 23 of 23 units; no non-pass classification, including append/idempotency, multi-process tailing, replay/reset, retry/quarantine/requeue, retention, socket, and browser dependencies. | `.cartulary/test-results/20260825T005043Z-p2085576` |

#### `WF-SL-05` execution record

| Field | Independent fact-port cutover evidence |
| --- | --- |
| Workstream and gap | `WF-SL-05`; closes `G-06` and implements `CT-AC-01` through `CT-AC-10` under `CT-REQ-001` through `CT-REQ-005`, `CT-REQ-008`, `CT-REQ-009`, and `CT-REQ-010`. |
| Start and end | `2026-08-24T20:54:06-04:00` through `2026-08-24T23:23:36-04:00`. |
| Starting source and status | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`; all completed predecessor paths were dirty, the tracker remained `AM`, and no fact-port, owner-publication, or publication-catalog file yet existed. The existing index was not changed. |
| Files changed | Added explicit Collaboration publication/catalog code, Revisions record-publication and revision-fact code, source-owner publication/fact helpers, and `internal/testutil/collaborationsupport` composition. Updated application assembly; every live Artifacts, Assessments, Hosts/Identities, Indicators, Parties, Tasks/Decisions, Imports, Timeline, Evidence, Mentions, Merge, and Revisions destructive caller; affected tests and exact export/boundary manifests; generated contract/topology projections through `make generate`; and this tracker. Removed the combined Revisions/Collaboration appender, root record-diff/publication files, and historical suppression implementation/tests. |
| Substantive result | Revisions now accepts `LiveRevisionInput` with canonical, unique, presence-aware `RevisionConflictFact` values and a distinct `HistoricalRevisionInput`. Collaboration accepts `RecordChangeIntentInput` with catalog-validated public keys and affected views and creates exactly one canonical intent through the borrowed transaction. Application composition binds the two ports without a Revisions-to-Collaboration representation dependency. The immutable publication catalog rejects incomplete, duplicate, cross-view, non-public, invalid-patch, and undeclared-patch inputs. Merge now supplies source-owned Host and Identity conflict facts instead of omitting them. No Revisions or Collaboration production package imports the other. |
| Acceptance closure | Catalog/fact unit cases cover canonical ordering, duplicate/invalid keys, before/after presence including explicit JSON null, immutable declarations, identity/version agreement, view ownership, change kinds, private fields, and patch shape. All enumerated source owners, destructive delete/restore/rollback paths, atomicity/idempotency paths, and explicit conflict/reconciliation browser rows passed. Static audits found no combined appender, shared diff, historical Boolean/suppression policy, retired intent constructor, forwarding alias, or production cross-import. Boundary policy now makes the mutual-import and retired-symbol constraints executable. |
| Compatibility and migration | Internal Go consumers migrated atomically with no alias, adapter preserving a retired API, dual write/read, flag, backfill, schema change, or stored-data migration. Existing public record-change payloads, writable-field conflict behavior, routes, CLI, Incident Bundle, authorization, and telemetry remain unchanged. Historical reconstruction now has an explicit non-live operation and cannot publish. Accidental private/computed-field coupling is removed from the public port. |
| Rollback point | Revert the fact inputs, publication catalog/appender, all source-owner/application assembly migrations, test support, boundary policy, authored selector change, and Make-generated projections as one slice. Do not restore only the combined appender or retain both fact paths. `WF-SL-04` is the last independently valid checkpoint. |
| Residual risks | No `WF-SL-05` defect remains. Direct Collaboration-table assertions in ordinary cross-owner tests are intentionally deferred to `WF-SL-07`; recovery SQL and authority remain for `WF-SL-06`; final constructor/raw-SQL/support-path prohibitions and broad verification remain for `WF-SL-08`. |
| Markdown validation | `make lint-markdown` passed with run root `.cartulary/test-results/20260825T032502Z-p4104060`. |
| Binary exit | **PASS.** The owner matrix, browser conflict/reconciliation rows, fast suite, boundary, generated, JSON, harness, static, and tracker checks pass. `WF-SL-05` is `DONE`. |

Final owner-slice evidence (each graph had exactly one terminal result per
selected unit and no non-pass result):

| Owner | No-service result | Service-backed result |
| --- | --- | --- |
| Collaboration | 32/32, `.cartulary/test-results/20260825T020640Z-p2410095` | 23/23, `.cartulary/test-results/20260825T020813Z-p2457864` |
| Revisions | 27/27, `.cartulary/test-results/20260825T021228Z-p2570721` | 20/20, `.cartulary/test-results/20260825T021337Z-p2615869` |
| Artifacts | 7/7, `.cartulary/test-results/20260825T021449Z-p2660446` | 3/3, `.cartulary/test-results/20260825T021535Z-p2676605` |
| Assessments | 27/27, `.cartulary/test-results/20260825T021832Z-p2751393` | 18/18, `.cartulary/test-results/20260825T021933Z-p2794132` |
| Entities | 40/40, `.cartulary/test-results/20260825T022355Z-p2910807` | 31/31, `.cartulary/test-results/20260825T022553Z-p2965242` |
| Evidence | 35/35, `.cartulary/test-results/20260825T022752Z-p3019329` | 25/25, `.cartulary/test-results/20260825T023154Z-p3157753` |
| Indicators | 20/20, `.cartulary/test-results/20260825T023318Z-p3206978` | 8/8, `.cartulary/test-results/20260825T023405Z-p3223595` |
| Parties | 20/20, `.cartulary/test-results/20260825T023453Z-p3239769` | 17/17, `.cartulary/test-results/20260825T023549Z-p3279819` |
| Tasks/Decisions | 18/18, `.cartulary/test-results/20260825T023641Z-p3319836` | 15/15, `.cartulary/test-results/20260825T023730Z-p3359678` |
| Timeline | 51/51, `.cartulary/test-results/20260825T024419Z-p3477636` | 29/29, `.cartulary/test-results/20260825T024910Z-p3535550` |
| Imports | 23/23, `.cartulary/test-results/20260825T025352Z-p3593235` | 14/14, `.cartulary/test-results/20260825T031056Z-p3984842` |
| Incident Bundles | 8/8, `.cartulary/test-results/20260825T025827Z-p3678289` | 6/6, `.cartulary/test-results/20260825T025928Z-p3694607` |
| Workbook | 66/66, `.cartulary/test-results/20260825T030427Z-p3788735` | 37/37, `.cartulary/test-results/20260825T030644Z-p3846483` |
| App server | 24/24, `.cartulary/test-results/20260825T030859Z-p3903971` | 17/17, `.cartulary/test-results/20260825T030957Z-p3944524` |

Additional final validation:

| Exact command | Result | Run root or disposition |
| --- | --- | --- |
| `make test-fast` | Passed 430 of 430 units after the final API naming repair. | `.cartulary/test-results/20260825T031735Z-p4075517` |
| Explicit Collaboration conflict and reconciliation browser selection | Passed 14 of 14 graph units, including conflict-resolver submission and multi-client live-row update/presence. | `.cartulary/test-results/20260825T031227Z-p4026958` |
| Targeted durable-stream regression selection | Passed 3 of 3 graph units. | `.cartulary/test-results/20260825T020424Z-p2384144` |
| `make format` | Passed 2 of 2 units after the final authored Go edit. | `.cartulary/test-results/20260825T031728Z-p4071548` |
| `make generate` | Passed; refreshed only declared generated consequences of current authored inputs. | `.cartulary/test-results/20260825T031905Z-p4088551` |
| Final `make generate-drift` | Passed 4 of 4 units. | `.cartulary/test-results/20260825T032304Z-p4100393` |
| `make generated-artifact-policy-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T031934Z-p4094394` |
| Final `make json-shape-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T032203Z-p4098213` |
| Final `make harness-contract` | Passed 2 of 2 units after updating the mechanically asserted dedicated-fixture count from 262 to 261 for the deleted historical-suppression row. | `.cartulary/test-results/20260825T032053Z-p4097262` |
| Final `make backend-module-boundary-check` | Passed 3 of 3 units with executable mutual-import, retired-API, and construction constraints. | `.cartulary/test-results/20260825T032258Z-p4100002` |
| Exact `rg` retirement and reciprocal-import audits | No production Revisions-to-Collaboration import, Collaboration-to-Revisions import, combined/diff/historical-policy API, retired constructor, or generic historical wrapper remained. | Non-retained read-only audit at checkpoint. |

All non-pass executions were slice-caused repair evidence or classified
infrastructure, and none remains in the final source. Early atomic-cutover
compile/test repairs are retained at
`.cartulary/test-results/20260825T010340Z-p2143401`,
`.cartulary/test-results/20260825T010618Z-p2148873`,
`.cartulary/test-results/20260825T012216Z-p2159904`,
`.cartulary/test-results/20260825T012315Z-p2165017`,
`.cartulary/test-results/20260825T012358Z-p2170073`,
`.cartulary/test-results/20260825T013402Z-p2245915`,
`.cartulary/test-results/20260825T014252Z-p2263277`,
`.cartulary/test-results/20260825T014351Z-p2272365`,
`.cartulary/test-results/20260825T014711Z-p2288176`,
`.cartulary/test-results/20260825T014915Z-p2289311`,
`.cartulary/test-results/20260825T014935Z-p2290371`,
`.cartulary/test-results/20260825T015008Z-p2292462`,
`.cartulary/test-results/20260825T015711Z-p2306180`,
`.cartulary/test-results/20260825T015757Z-p2314386`, and
`.cartulary/test-results/20260825T020120Z-p2329322`. They respectively exposed
stale combined-port callers, interface/export mismatches, missing explicit
fact/publication dependencies, syntax/format damage during mechanical edits,
and a private stream error-sentinel mismatch; each diagnostic was repaired and
superseded by the final fast and owner matrices above.

The first complete Revisions, Assessments, Entities, Timeline, and Workbook
runs failed only because their direct test compositions had not yet received
the new publication capability; those exact roots are
`.cartulary/test-results/20260825T020950Z-p2505644`,
`.cartulary/test-results/20260825T021626Z-p2692608`,
`.cartulary/test-results/20260825T022031Z-p2836264`,
`.cartulary/test-results/20260825T023821Z-p3399508`, and
`.cartulary/test-results/20260825T030036Z-p3710762`. The first Evidence
service run had one transient object-upload HTTP 503 at
`.cartulary/test-results/20260825T022913Z-p3068790`; its exact row then passed
11/11 at `.cartulary/test-results/20260825T023054Z-p3118166`, followed by the
clean full service run above. The first Imports service run ended in a service
readiness timeout without a product-row failure at
`.cartulary/test-results/20260825T025505Z-p3635516`; the clean full rerun is
recorded above.

The first boundary runs at
`.cartulary/test-results/20260825T031329Z-p4068991` and
`.cartulary/test-results/20260825T032212Z-p4098759` correctly found an omitted
exact Assessments port-import allowance and a characterization string matching
the new retired-symbol rule. Both policies were narrowed explicitly, and the
final boundary result is green. Pre-generation drift at
`.cartulary/test-results/20260825T031833Z-p4085144` was the expected stale
generated projection and passed after `make generate`. The first harness
contract at `.cartulary/test-results/20260825T031947Z-p4095389` found the stale
dedicated-fixture count caused by the deleted active row; the authored count
was repaired and the final contract is green.

The final `DONE` status and handoff-log edit was revalidated by
`make lint-markdown`; it passed with run root
`.cartulary/test-results/20260825T032604Z-p4105317`.

#### `WF-SL-06` execution record

| Field | Recovery-isolation evidence |
| --- | --- |
| Workstream and gap | `WF-SL-06`; closes `G-07` under `CT-REQ-001`, `CT-REQ-002`, `CT-REQ-005`, `CT-REQ-008`, `CT-REQ-009`, and `CT-REQ-010`. |
| Start and end | `2026-08-24T23:26:30-04:00` through `2026-08-24T23:49:38-04:00`. |
| Starting source and status | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`; all completed predecessor paths were dirty and the tracker remained `AM`. Recovery SQL, locks, proof validation, audit append, and transaction outcome handling still occupied the root `recovery.go`. The existing index was not changed. |
| Files changed | Added `internal/modules/collaboration/internal/recovery/adapter.go`; reduced root `internal/modules/collaboration/recovery.go` to narrow capability vocabulary/delegation; updated `internal/app/operator/operator_collaboration.go`, the Collaboration durable-stream integration composition, the Operator process fixture, `tools/backend_module_boundaries.json`, and this tracker. |
| Substantive result | Operator now consumes `RecoveryCapability` and cannot observe a concrete recovery service. The private adapter exclusively owns transaction admission, quarantine and pending-intent `FOR UPDATE` locks, payload repair proof, retry/cursor mutation, administrative-audit persistence, rollback, and commit-unknown classification. The root facade maps the closed private failure vocabulary explicitly and contains no SQL, `pgx`, audit, begin/commit/rollback, or proof mechanics. `RecoveryStateContribution` remains a separate pure-data contribution. |
| Compatibility and migration | The internal constructor changed atomically from `NewRecoveryService` to `NewRecoveryCapability`; no alias or concrete compatibility service remains. Operator grammar, v2 JSON bytes, reason codes, exit codes, state transition, audit identity, concurrency, cancellation, timeout, commit-unknown behavior, recovery catalog, database schema, and stored data are unchanged. |
| Rollback point | Revert the private adapter, root capability facade, Operator/test composition, and recovery boundary rules together to restore the prior root service. Do not retain both constructors or two implementations. `WF-SL-05` is the last independently valid checkpoint. |
| Residual risks | No slice-local recovery defect remains. Physical recovery/role SQL stays in owner tests by design. Ordinary cross-owner intent assertions and module-local scenario support remain for `WF-SL-07`; final raw-SQL/support-path enforcement and broad checks remain for `WF-SL-08`. |
| Markdown validation | `make lint-markdown` passed with run root `.cartulary/test-results/20260825T035040Z-p457913`. |
| Binary exit | **PASS.** Collaboration, Operator, Recovery, PostgreSQL-role, fast, targeted requeue/process, boundary, JSON, generation-drift, formatting, and tracker validations pass; `WF-SL-06` is `DONE`. |

| Exact command | Result | Run root or disposition |
| --- | --- | --- |
| Initial `make test-fast` | Passed 430 of 430 units after the private package move. | `.cartulary/test-results/20260825T032918Z-p4112485` |
| Final `make test-slice OWNER=module.collaboration` | Passed 32 of 32 units. | `.cartulary/test-results/20260825T033308Z-p4161022` |
| Final `make service-backed-test-slice OWNER=module.collaboration` | Passed 23 of 23 units, including requeue concurrency, repair proof, rollback, audit, cancellation, and commit-unknown cases. | `.cartulary/test-results/20260825T033443Z-p17168` |
| Final `make test-slice OWNER=app.operator` | Passed 12 of 12 units; CLI grammar, v2 delivery bytes, failure mapping, and registry surfaces are unchanged. | `.cartulary/test-results/20260825T033937Z-p177098` |
| Final `make service-backed-test-slice OWNER=app.operator` | Passed 9 of 9 units, including the Collaboration requeue process contract. | `.cartulary/test-results/20260825T034015Z-p213364` |
| `make test-slice OWNER=module.recovery` | Passed 24 of 24 units, including exact recovery-state catalog checks. | `.cartulary/test-results/20260825T034056Z-p249306` |
| `make service-backed-test-slice OWNER=module.recovery` | Passed 19 of 19 units, including recovery-state database coverage. | `.cartulary/test-results/20260825T034216Z-p301869` |
| `make test-slice OWNER=platform.postgres` | Passed 4 of 4 units. | `.cartulary/test-results/20260825T034348Z-p354421` |
| `make service-backed-test-slice OWNER=platform.postgres` | Passed 3 of 3 units, including exact role ownership and privileges. | `.cartulary/test-results/20260825T034435Z-p370463` |
| Final targeted durable-stream requeue row | Passed 3 of 3 graph units after explicit private-to-public failure mapping. | `.cartulary/test-results/20260825T034621Z-p390893` |
| Final targeted Operator requeue process row | Passed 7 of 7 graph units. | `.cartulary/test-results/20260825T034720Z-p408202` |
| Final `make test-fast` | Passed 430 of 430 units. | `.cartulary/test-results/20260825T034814Z-p444333` |
| Final `make format` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T034610Z-p386978` |
| `make json-shape-check` | Passed 3 of 3 units for the authored boundary-policy edit. | `.cartulary/test-results/20260825T033219Z-p4159252` |
| Final `make backend-module-boundary-check` | Passed 3 of 3 units; construction and root-mechanics prohibitions are active. | `.cartulary/test-results/20260825T034925Z-p457056` |
| `make generate-drift` | Passed 4 of 4 units. | `.cartulary/test-results/20260825T034913Z-p454023` |
| Exact root/private recovery `rg` audits | Root recovery had no SQL/transaction/audit tokens; private recovery contained all Collaboration requeue SQL and mechanics; no retired root service remained. | Non-retained read-only audit at checkpoint. |

The first Operator owner runs at
`.cartulary/test-results/20260825T033614Z-p64773` and
`.cartulary/test-results/20260825T033811Z-p136972`, plus the exact-row run at
`.cartulary/test-results/20260825T033722Z-p101269`, were classified
`infrastructure_failed/missing_selector_result`: the package could not compile
because the SL-05 fixture migration left one reference to the retired
`intent.CanonicalPayload` variable. Replacing it with the already-authored raw
`intentPayload` restored both process selectors; the complete final Operator
graphs above supersede all three failures. The first recovery-boundary run at
`.cartulary/test-results/20260825T033229Z-p4159782` correctly found that the
retired token `NewRecoveryService` was too broad and also matched unrelated
`NewRecoveryServices`; the rule was narrowed to the qualified Collaboration
constructor, and the final boundary graph is green.

#### `WF-SL-07` execution record

| Field | Shared test-support consolidation evidence |
| --- | --- |
| Workstream and gap | `WF-SL-07`; closes `G-08` under `CT-REQ-001`, `CT-REQ-002`, `CT-REQ-005`, `CT-REQ-008`, `CT-REQ-009`, and `CT-REQ-010`. |
| Start and end | `2026-08-24T23:51:30-04:00` through `2026-08-25T01:04:48-04:00`. |
| Starting source and status | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`; all completed predecessor paths were dirty, the tracker remained `AM`, owner-local socket/intent/scenario support still existed, and ordinary cross-owner tests still queried Collaboration tables. The existing index was not changed. |
| Files changed | Moved socket inventory/view-event helpers and intent fixtures to `internal/testutil/collaborationsupport`; added centralized semantic intent/count/idle assertions and a narrow intent-only subpackage; removed the module-local scenario wrapper and the empty owner-local support tree; migrated affected Collaboration, application, source-owner, jobs, recovery, and shared-support tests; updated `tools/test_support_inventory.json`, `tools/backend_module_boundaries.json`, and this tracker. |
| Substantive result | Reusable socket and intent fixtures now have one shared owner. Collaboration tests compose through `appsupport` directly, with no duplicate runtime wrapper. Ordinary cross-owner tests assert canonical intent semantics through nullable-aware helpers instead of treating Collaboration table columns as their own contract. Direct table SQL remains only in Collaboration owner tests, recovery/role/physical-storage evidence, generated recovery code, and the centralized semantic adapter. A narrow `intenttest` helper avoids importing full application composition into same-package Jobs tests. |
| Harness and policy result | The authored test-support inventory no longer declares the retired owner-local root. Backend policy rejects both retired support imports and Collaboration-table SQL outside an explicit physical-evidence allowlist. No generated topology change was required; harness, JSON shape, generated policy, and drift checks pass. Exact source/import and raw-SQL audits are empty outside their policy declarations and intentional allowlists. |
| Compatibility and migration | Test imports, fixtures, assertions, and selectors migrated atomically. No production route, WebSocket payload, transaction, table, stored data, CLI, authorization, Incident Bundle, or telemetry behavior changed. No compatibility wrapper or forwarding alias remains. |
| Rollback point | Revert the shared support files, every importing test, semantic assertion migration, support inventory, and boundary rules together; restore the owner-local socket/intent/scenario tree only as one unit. `WF-SL-06` is the last independently valid checkpoint. |
| Residual risks | No slice-local test-support defect remains. Final removal of any remaining transitional implementation surface, complete topology audits, `make agent-finalize`, and `make check` remain exclusively in `WF-SL-08`. |
| Markdown validation | `make lint-markdown` passed with run root `.cartulary/test-results/20260825T050949Z-p2020981`. |
| Binary exit | **PASS.** All implementation, owner, harness, generation, policy, formatting, static, and tracker acceptance checks pass; `WF-SL-07` is `DONE`. |

Final affected-owner evidence (every graph had exactly one terminal result per
selected unit and no non-pass result):

| Exact commands | No-service result | Service-backed result |
| --- | --- | --- |
| `make test-slice OWNER=module.collaboration`; `make service-backed-test-slice OWNER=module.collaboration` | 32/32, `.cartulary/test-results/20260825T041312Z-p592435` | 23/23, `.cartulary/test-results/20260825T041444Z-p640743` |
| `make test-slice OWNER=module.artifacts`; `make service-backed-test-slice OWNER=module.artifacts` | 7/7, `.cartulary/test-results/20260825T041616Z-p688370` | 3/3, `.cartulary/test-results/20260825T041702Z-p704337` |
| `make test-slice OWNER=module.assessments`; `make service-backed-test-slice OWNER=module.assessments` | 27/27, `.cartulary/test-results/20260825T041745Z-p720289` | 18/18, `.cartulary/test-results/20260825T041845Z-p762441` |
| `make test-slice OWNER=module.entities`; `make service-backed-test-slice OWNER=module.entities` | 40/40, `.cartulary/test-results/20260825T041944Z-p804580` | 31/31, `.cartulary/test-results/20260825T042137Z-p858929` |
| `make test-slice OWNER=module.evidence`; `make service-backed-test-slice OWNER=module.evidence` | 35/35, `.cartulary/test-results/20260825T042335Z-p912908` | 25/25, `.cartulary/test-results/20260825T042458Z-p962360` |
| `make test-slice OWNER=module.indicators`; `make service-backed-test-slice OWNER=module.indicators` | 20/20, `.cartulary/test-results/20260825T042616Z-p1011571` | 8/8, `.cartulary/test-results/20260825T042708Z-p1027926` |
| `make test-slice OWNER=module.tasksdecisions`; `make service-backed-test-slice OWNER=module.tasksdecisions` | 18/18, `.cartulary/test-results/20260825T042756Z-p1044120` | 15/15, `.cartulary/test-results/20260825T042846Z-p1083868` |
| `make test-slice OWNER=module.imports`; `make service-backed-test-slice OWNER=module.imports` | 23/23, `.cartulary/test-results/20260825T042936Z-p1123573` | 14/14, `.cartulary/test-results/20260825T043055Z-p1165605` |
| `make test-slice OWNER=module.timeline`; `make service-backed-test-slice OWNER=module.timeline` | 51/51, `.cartulary/test-results/20260825T043210Z-p1207639` | 29/29, `.cartulary/test-results/20260825T043702Z-p1265381` |
| `make test-slice OWNER=module.workbook`; `make service-backed-test-slice OWNER=module.workbook` | 66/66, `.cartulary/test-results/20260825T044146Z-p1322982` | 37/37, `.cartulary/test-results/20260825T044405Z-p1381449` |
| `make test-slice OWNER=module.incidents`; `make service-backed-test-slice OWNER=module.incidents` | 27/27, `.cartulary/test-results/20260825T044623Z-p1438862` | 19/19, `.cartulary/test-results/20260825T044750Z-p1483430` |
| `make test-slice OWNER=module.links`; `make service-backed-test-slice OWNER=module.links` | 14/14, `.cartulary/test-results/20260825T044919Z-p1527982` | 13/13, `.cartulary/test-results/20260825T045006Z-p1567670` |
| `make test-slice OWNER=module.auth`; `make service-backed-test-slice OWNER=module.auth` | 35/35, `.cartulary/test-results/20260825T045053Z-p1607281` | 25/25, `.cartulary/test-results/20260825T045232Z-p1658939` |
| `make test-slice OWNER=module.revisions`; `make service-backed-test-slice OWNER=module.revisions` | 27/27, `.cartulary/test-results/20260825T045410Z-p1710478` | 20/20, `.cartulary/test-results/20260825T045522Z-p1755457` |
| `make test-slice OWNER=module.incidentbundles`; `make service-backed-test-slice OWNER=module.incidentbundles` | 8/8, `.cartulary/test-results/20260825T045628Z-p1800016` | 6/6, `.cartulary/test-results/20260825T050039Z-p1833110` |
| `make test-slice OWNER=platform.jobs`; `make service-backed-test-slice OWNER=platform.jobs` | 6/6, `.cartulary/test-results/20260825T040314Z-p544327` | 5/5, `.cartulary/test-results/20260825T050139Z-p1849281` |
| `make test-slice OWNER=module.extensions`; `make service-backed-test-slice OWNER=module.extensions` | 24/24, `.cartulary/test-results/20260825T050536Z-p1876975` | 22/22, `.cartulary/test-results/20260825T050628Z-p1948771` |
| `make test-slice OWNER=module.crossownertransaction`; `make service-backed-test-slice OWNER=module.crossownertransaction` | 4/4, `.cartulary/test-results/20260825T050536Z-p1876983` | 3/3, `.cartulary/test-results/20260825T050628Z-p1948768` |
| `make test-slice OWNER=module.stagedobjects`; `make service-backed-test-slice OWNER=module.stagedobjects` | 5/5, `.cartulary/test-results/20260825T050536Z-p1876990` | 3/3, `.cartulary/test-results/20260825T050628Z-p1948773` |

Additional final validation:

| Exact command | Result | Run root or disposition |
| --- | --- | --- |
| `make test-slice OWNER=package.test_utils` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T050358Z-p1866413` |
| `make format` | Passed 2 of 2 units after the final authored Go edit. | `.cartulary/test-results/20260825T050407Z-p1866916` |
| `make test-fast` | Passed 430 of 430 units. | `.cartulary/test-results/20260825T050415Z-p1871410` |
| `make json-shape-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T050415Z-p1871006` |
| `make backend-module-boundary-check` | Passed 3 of 3 units with the raw-SQL and retired-support prohibitions active. | `.cartulary/test-results/20260825T050415Z-p1871260` |
| `make harness-contract` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T050415Z-p1871279` |
| `make generated-artifact-policy-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T050415Z-p1870917` |
| `make generate-drift` | Passed 4 of 4 units. | `.cartulary/test-results/20260825T050432Z-p1873354` |
| Exact retired-path and cross-owner table-token `rg` audits | No production/test import uses the retired support path, and no ordinary cross-owner Go test contains Collaboration table tokens outside the intentional policy allowlist. | Non-retained read-only audit at checkpoint. |

The first fast run at
`.cartulary/test-results/20260825T035721Z-p466669` and subsequent Jobs owner
runs at `.cartulary/test-results/20260825T035830Z-p474372`,
`.cartulary/test-results/20260825T035920Z-p490691`, and
`.cartulary/test-results/20260825T040104Z-p523053` were classified
`infrastructure_failed/missing_selector_result`: same-package Jobs tests
imported the root shared helper, which composes Jobs and therefore formed an
import cycle. The narrow intent-only helper removed that cycle. The next Jobs
run at `.cartulary/test-results/20260825T040137Z-p527503` exposed a nullable
source-row-version column in the semantic assertion record; changing the
field to `*int64` preserved the table contract and produced the final green
owner runs above.

JSON-shape runs at `.cartulary/test-results/20260825T040936Z-p574319` and
`.cartulary/test-results/20260825T040945Z-p574795` found the now-empty retired
support directory; removing the empty tree made the authored inventory and
filesystem agree. The first boundary run at
`.cartulary/test-results/20260825T041032Z-p576478` found that an existing
retired-handler token also matched a generic test helper argument. Replacing
that stringly assertion with `CountImportApplyIntents` put the semantic name
inside the Collaboration support owner and made the final rule pass without
weakening it. The first Incident Bundles service graph at
`.cartulary/test-results/20260825T045729Z-p1816163` ended in a service
readiness timeout with no product-row failure; the complete clean rerun is
recorded above.

#### `WF-SL-08` execution record

| Field | Final topology, validation, and handoff evidence |
| --- | --- |
| Workstream and gap | `WF-SL-08`; closes `G-09`, supplies final `AC-564` implementation evidence, and completes `CT-AC-01` through `CT-AC-10`, `DEC-AC-01` through `DEC-AC-07`, and `CT-REQ-001` through `CT-REQ-010`. |
| Start and end | `2026-08-25T01:10:32-04:00` through `2026-08-25T01:37:53-04:00`. |
| Starting source and status | HEAD `169ba53197681aa45767914b70cd86d8759d0d3f`; branch `main` three commits ahead of `origin/main`; all predecessor workstream paths were dirty, the tracker remained the sole staged path as `AM`, and no index change was authorized or made. A test-binary dispatcher bridge, a root error forwarding alias, and a policy rule scanning a deleted notifier file remained for final cleanup. |
| Files changed | Removed `internal/modules/collaboration/export_test.go`; updated the durable-stream owner test to construct the owner-private dispatcher directly and assert its private collision sentinel; defined the root publication collision error independently; replaced the deleted-file policy with an executable prohibition on retired root lifecycle/concrete/test exports and forwarding aliases; supplied the missing Collaboration publication port to the Network Flow shared Indicator test composition; completed this tracker. |
| Substantive result | One runtime facade and one registrar remain. Concrete hub, dispatcher, stream/replay store, and recovery mechanics have no root lifecycle export. The public collision vocabulary is mapped explicitly rather than aliasing a private storage sentinel. Boundary policy prohibits retired constructors, lifecycle types, forwarding aliases, mutual production imports, retired fact APIs, owner-local support paths, and unauthorized Collaboration-table SQL. Exact audits found no old symbol, reciprocal production import, retired support import, or unauthorized table token. No database migration or Collaboration/recovery contract changed. |
| Acceptance and conformance | `AC-564` is backed by the adopted decision, source-owner independent-port matrix, immutable catalog tests, recovery/facade/stream/test-support owner evidence, exact audits, executable boundary rules, and the passing full repository graph. All `CT-AC-*` and `DEC-AC-*` criteria are closed by their slice records and the final 656/656 check. Generated state, harness accounting, frontend/backend imports, JSON shapes, security/browser rows, and retained-run maintenance pass. |
| Compatibility and migration | Internal code and tests cut over atomically with no aliases, deprecated constructors, fallback readers, dual paths, flags, backfills, schema changes, or stored-data migration. Public HTTP routes, WebSocket message families, known-message semantics except the authorized decoder correction, CLI grammar/bytes, Incident Bundle versions, authorization precedence, storage schema, and telemetry vocabulary are unchanged. The private `0.0.0` protocol package and internal Go consumers migrated repository-wide. |
| Rollback point | Revert only the final dispatcher-test construction, independent error sentinel, final boundary-rule replacement, and Network Flow test-composition dependency to return to the `WF-SL-07` checkpoint. Earlier workstreams retain their recorded rollback points. Do not restore the test export or error alias independently. |
| Residual risks | No known remediation defect, unexplained non-pass, owner contradiction, compatibility shim, or required follow-up remains. The working tree intentionally remains uncommitted, and the pre-existing staged tracker entry was preserved. |
| Markdown validation | `make lint-markdown` passed with run root `.cartulary/test-results/20260825T053753Z-p2594418`. |
| Binary exit | **PASS.** Implementation, narrow validation, broad validation, retained evidence, and handoff content pass; `WF-SL-08` and the overall remediation are `DONE`. |

Final validation and retained evidence:

| Exact command | Result | Run root or disposition |
| --- | --- | --- |
| Initial `make format` after final topology cleanup | Passed 2 of 2 units. | `.cartulary/test-results/20260825T051232Z-p2024532` |
| Initial `make json-shape-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T051232Z-p2024231` |
| Initial `make backend-module-boundary-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T051232Z-p2024468` |
| Final `make test-slice OWNER=module.collaboration` | Passed 32 of 32 units. | `.cartulary/test-results/20260825T051444Z-p2078291` |
| Final `make service-backed-test-slice OWNER=module.collaboration` | Passed 23 of 23 units. | `.cartulary/test-results/20260825T051647Z-p2126512` |
| `make test-fast` | Passed 430 of 430 units. | `.cartulary/test-results/20260825T051821Z-p2174784` |
| Final narrow `make backend-module-boundary-check` | Passed 3 of 3 units with all final Collaboration prohibitions active. | `.cartulary/test-results/20260825T051821Z-p2174606` |
| Final narrow `make json-shape-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T051821Z-p2174241` |
| `make harness-contract` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T051821Z-p2174689` |
| `make generated-artifact-policy-check` | Passed 3 of 3 units. | `.cartulary/test-results/20260825T051821Z-p2174266` |
| `make frontend-import-boundary-check` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T051821Z-p2174640` |
| `make generate-drift` | Passed 4 of 4 units. | `.cartulary/test-results/20260825T051917Z-p2183989` |
| Initial `make agent-finalize` without `RESULTS_DIR` | Passed 1 of 1; retained-run maintenance was initially skipped because no successful full warm-check root existed yet. | `.cartulary/test-results/20260825T051931Z-p2186979` |
| Final repair `make format` | Passed 2 of 2 units. | `.cartulary/test-results/20260825T052547Z-p2322596` |
| `make test-slice OWNER=module.networkflow` | Passed 35 of 35 units. | `.cartulary/test-results/20260825T052556Z-p2326540` |
| `make service-backed-test-slice OWNER=module.networkflow` | Passed 30 of 30 units. | `.cartulary/test-results/20260825T052811Z-p2399175` |
| `make test-slice OWNER=module.reporting` | Passed 5 of 5 units. | `.cartulary/test-results/20260825T052556Z-p2326542` |
| `make service-backed-test-slice OWNER=module.reporting` | Passed 4 of 4 units. | `.cartulary/test-results/20260825T052811Z-p2399229` |
| Final pre-check `make agent-finalize` | Passed 1 of 1 on repaired source. | `.cartulary/test-results/20260825T053026Z-p2471434` |
| Final `make check` | Passed 656 of 656 units with one terminal pass per selected unit. | `.cartulary/test-results/20260825T053042Z-p2474350` |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260825T053042Z-p2474350` | Passed 1 of 1 and completed retained-success maintenance. | `.cartulary/test-results/20260825T053524Z-p2590396` |
| Exact retired-symbol, reciprocal-production-import, retired-support-import, unauthorized-table-token, public-contract-path, and `git diff --check` audits | All returned empty or passed. The index remained exactly one pre-existing staged tracker addition. | Non-retained read-only audit at checkpoint. |

The first Collaboration owner run at
`.cartulary/test-results/20260825T051232Z-p2024315` passed 31 of 32 units but
failed its durable-stream row because the physical owner test still compared
the private-store error to the newly independent root sentinel. Updating that
physical test to assert `privatestream.ErrIntentKeyCollision` removed the last
representation alias; the complete no-service and service-backed owner runs
above supersede it.

The first `make check` at
`.cartulary/test-results/20260825T051949Z-p2189881` passed 651 of 656 units and
failed five Network Flow/Reporting Go units. Every failure had the same direct
cause: `newTestNetworkFlowStore` composed the Indicators owner without its new
required Collaboration publication capability. Supplying
`collaborationsupport.NewPublicationAppender()` once in that shared test
constructor closed the missed source-owner composition caller. The complete
Network Flow and Reporting owner graphs and final 656/656 repository graph
above supersede all five failures. No unrelated or unexplained non-pass
remains.

The baseline record MUST retain all of the following. A prose statement that a
command was green is insufficient.

| Evidence class | Required retained data |
| --- | --- |
| Source identity | Full commit SHA and exact clean/dirty status. |
| Selection identity | Owner ID, selected row IDs and count, catalog digest, verification digest, and topology/profile digests. |
| Invocation identity | Exact Make target and resolved stable command ID. |
| Run identity | Result root, run ID, run root, and start/completion timestamps. |
| Per-row closure | Exactly one terminal result for every selected row. |
| Failure evidence | Primary failure classification, exact row, logs, artifacts, and applicable service evidence. |
| Classification | One of pre-existing product defect, harness/infrastructure defect, authorized omission, dependency consequence, or cancellation. |
| Accountability | Owning module, remediation reference, and explicit effect on `SL-02`, `SL-03`, `SL-04`, and `SL-05`. |

The Testing Harness terminal-state vocabulary is closed:

| Terminal state | Baseline interpretation |
| --- | --- |
| `passed` | The selected row executed and passed with complete accounting. |
| `failed` | A product assertion or unauthorized skip failed. |
| `infrastructure_failed` | Harness, setup, service, fixture, or browser infrastructure prevented execution. |
| `skipped_dependency` | A failed declared dependency prevented execution. |
| `cancelled` | A supported signal or explicit cancellation stopped execution. |
| `skipped_authorized` | One exact unexpired verification-level authorization applies. |

A **PASSING BASELINE** has every selected row in `passed`. A successful
invocation containing `skipped_authorized` is instead a **CLASSIFIED BASELINE
WITH AUTHORIZED OMISSION**. Any other classified baseline MUST give every
non-pass one primary cause, owner, remediation reference, affected-slice
determination, and evidence that it reproduces independently of the proposed
refactor. Missing or duplicate results are harness-accounting failures, not
classifiable product baselines.

Failures in revision/conflict reconstruction, same-field versus
different-field detection, transaction atomicity, record-change
canonicalization, historical suppression, replay ordering/deduplication,
sparse patch/invalidate behavior, WebSocket codec/decoder behavior, or a known
source-owner caller are hard blockers for `SL-02` or `SL-05`. A failure proven
unrelated MAY permit an independent slice only after its exact exception and
evidence are written into this tracker.

The following guard baselines supplement but never replace B0/B1:

- `make backend-module-boundary-check`;
- `make generate-drift`;
- `make generated-artifact-policy-check`;
- `make json-shape-check`; and
- for `SL-02`, `make test-slice OWNER=package.protocol_ts`.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.collaboration` | Owner-selected no-service backend/frontend rows and explicit selected rows when `ROWS` is supplied | yes | Use `make task-guide ROLE=module-author OWNER=module.collaboration` before choosing narrower rows. |
| integration | `make service-backed-test-slice OWNER=module.collaboration` | Owner-selected PostgreSQL/object-store/browser-backed Collaboration rows | yes | Establish a passing or classified baseline under `RB-002`; rerun after persistence, route, recovery, or testutil slices. |
| e2e/browser | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Functional and stateful browser Collaboration behavior | no | Required after route, decoder, frontend-contract, authorization, reset, pending-edit, or presence changes. Use accessibility/visual targets only when those surfaces change. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Authored WS/harness/recovery inputs and generated Go/TypeScript/topology projections | yes | Run `make generate` only when an authorized authored input change requires regeneration. Never hand-edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check`; `make frontend-import-boundary-check` | Backend module imports/SQL ownership and frontend import ownership | yes | Required for facade, private package, testutil, or frontend entrypoint movement. |
| protocol package | `make test-slice OWNER=package.protocol_ts` | Generated/public protocol decoder behavior | no | Required for `SL-02` and any generated Collaboration type/validator change. |
| operator recovery | `make test-slice OWNER=app.operator`; `make service-backed-test-slice OWNER=app.operator` | Requeue grammar, result delivery, and process contract | no | Required for `SL-06`. |
| harness accounting | `make harness-contract` | Owner catalog, selector, evidence, fixture, and harness mechanics | no | Required for `SL-01`, `SL-07`, or any owner/test-family input change. |
| tracker documentation | `make lint-markdown` | Authored Markdown, including this tracker | yes | The only product-adjacent check appropriate to the present documentation-only task. |
| full check | `make agent-finalize`; `make check` | Final harness maintenance and standard repository verification gate | no | Required after implementation, not for source discovery. Record run roots and classify unrelated failures. |

Read-only discovery commands run for this planning session:

- `git status --short --branch`, `git rev-parse HEAD`, `date
  --iso-8601=seconds`, `find`, `wc`, `rg`, `sed`, and `jq` for baseline,
  inventory, exact source, owner, contract, caller, schema, and test-map
  inspection;
- `make task-guide ROLE=module-author OWNER=module.collaboration`;
- `make explain-test-owner OWNER=module.collaboration`;
- `make help`; and
- `make explain-target TARGET=<target> DETAIL=summary` for the owner-slice,
  browser, generation, boundary, harness, and full-check targets.

Current tracker-task validation result: `make lint-markdown` passed and the
updated tracker was revalidated on 2026-08-24 with run root
`.cartulary/test-results/20260824T214025Z-p1019868`.

That run root is historical evidence for the original tracker. The NLSpec-style
revision requires a new `make lint-markdown` result, recorded in Section 10;
the historical pass MUST NOT be cited as validation of later edits.

NLSpec-style revision validation: `make lint-markdown` passed on 2026-08-24
with run root `.cartulary/test-results/20260824T222840Z-p1033205`. No product
test, generation, conformance, owner-slice, or broad-check target was run.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| CT-001 | Establish source hierarchy, baseline, safe label, and single-file write boundary | WF-00 | DONE | none | Section 1; commit and session baseline | Authority and allowed scope are explicit. |
| CT-002 | Inventory all 28 target files and direct surfaces | WF-01 | DONE | CT-001 | Section 2 | Every file has an evidence-backed row. |
| CT-003 | Map public WS, stream, recovery, frontend, telemetry, generated, and harness contracts | WF-02 | DONE | CT-002 | Section 4 | Every discovered contract risk has an owner and test posture. |
| CT-004 | Identify characterization and evidence-accounting gaps | WF-03 | DONE | CT-003 | Additive decoder mismatch and stale harness findings | Gaps have classifications and planned slices. |
| CT-005 | Diagnose module boundary and coupling | WF-04 | DONE | CT-002 | Sections 3 and 5 | Collaboration is classified as a legitimate mixed-responsibility boundary with negative findings recorded. |
| CT-006 | Specify the row-diff/private-fact replacement and candidate port shape | WF-05 | DONE | CT-003, CT-004, CT-005 | `CT-REQ-003`; Sections 3.1-3.4 | The planning recommendation assigns both fact models, consumer ownership, transaction order, defaults, and acceptance criteria without claiming adoption. |
| CT-006A | Adopt the independent-port topology | `WF-SPEC-01` | DONE | CT-006, RB-001 | REQ-00-075; Collaboration decision; amended Revisions decision | The controlling owner decisions name exact ports, coordination ownership, import directions, cutover, and no-compatibility rule. |
| CT-007 | Define independently reversible implementation slices | WF-06 | DONE | CT-003, CT-004, CT-005 | Section 7 | Every slice has dependency, risk, validation, rollback, and completion criteria. |
| CT-008 | Define owner, generated, boundary, browser, operator, and harness validation | WF-07 | DONE | CT-004, CT-007 | Section 8 and owner manifests | Commands are discovered and not invented. |
| CT-009 | Establish passing B0 and post-accounting B1 owner baselines | `WF-EXEC-00`, `WF-SL-01` | DONE | RB-002 | B0/B1 run roots in Section 8.1 | No unexplained baseline failure can be attributed to later slices. |
| CT-010 | Implement SL-01 through SL-08 | WF-06 | DONE | CT-006A, CT-009 | Section 8.1 workstream records and final full-check root | All slices complete without unauthorized behavior change and every applicable `CT-AC-*`/`DEC-AC-*` row passes. |
| CT-011 | Run finalization/full validation and close handoff | WF-08 | DONE | CT-010 | Finalization and 656/656 full-check roots in Section 8.1 | Required checks pass, repair failures are classified, and no residual remediation remains. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-24T17:26:00-04:00 | Codex planning session | Authority and planning boundary established; no owner contradiction found. | Inspected framework, Domain, Core 00-04, relevant adopted NLSpecs/decisions; touched only this tracker. | Read-only `sed`/`rg`; Git baseline commands. | Target exists; safe label is `collaboration`; only a later task may implement changes. | None for tracker creation. | Keep all future work within the owner hierarchy in Section 1. |
| 2026-08-24T18:22:06-04:00 | Codex NLSpec-style tracker revision | Tracker authority vocabulary, pre-production defaults, requirement IDs, and document placement are explicit. | Inspected `temp/analysis-notes.md`, `docs/research/nlspec-spec.md`, the tracker, and controlling owner evidence; touched only this tracker. | `date`, Git status/revision/diff inspection, `wc`, `sed`, `rg`. | Research and style guidance remain non-authoritative; prior handoff history is preserved. | None for documentation completion. | Apply `CT-REQ-001` through `CT-REQ-010` in every later handoff. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-24T17:26:00-04:00 | Codex planning session | Collaboration implementation capability diagnosed as a mixed flat package with protocol, persistence, recovery, telemetry, and test-support responsibilities; later vocabulary review corrected the original bounded-context label. | Inspected all 28 target files and representative app/source-owner callers; touched only this tracker. | `find`, `wc`, `rg`, exact `sed` reads. | Keep semantic Collaboration implementation ownership without creating a domain context; plan private facade, stream, recovery, and testutil seams. | `RB-001` | Resolve the row-diff/private conflict topology before SL-05. |
| 2026-08-24T18:22:06-04:00 | Codex NLSpec-style tracker revision | The replacement fact models, candidate ports, transaction sequence, historical variant, and acceptance gates are decision-complete planning material. | Inspected Collaboration record-change code, Revisions appender/publication code, source-owner ports, the adopted Revisions boundary, and Core transaction/fact requirements; touched only this tracker. | Exact `sed`/`rg` reads and tracker diff checks. | The notes' independent-port recommendation differs from the adopted Revisions publication topology and is not claimed as adopted. | `RB-001` owner adoption. | Adopt the proposed topology or an explicitly equivalent owner-compliant topology before `SL-05`. |
| 2026-08-24T19:17:08-04:00 | Codex `WF-SPEC-01` | Independent-port architecture adopted and the domain classification corrected. | Added the Collaboration decision; amended Revisions, Domain, Core 00, Core 04, traceability, and this tracker. | Exact owner/ID/terminology `sed` and `rg` audits; `git diff --check`; `make lint-markdown`. | REQ-00-075 and AC-564 are adopted, Revisions no longer owns public Collaboration publication, presence remains Workbook Interaction language, and Markdown lint passed at `.cartulary/test-results/20260824T232336Z-p1152178`. | None; AC-564 implementation evidence remains a final-workstream obligation. | Begin accounting-only `WF-SL-01`; establish B1 before decoder or production work. |
| 2026-08-24T20:15:26-04:00 | Codex `WF-SL-03` | One Collaboration facade and one semantic protocol boundary now own lifecycle and wire semantics. | Added runtime/protocol packages; privatized hub/dispatcher/route service; updated server, Auth/Incident effects, tests, and this tracker. | Both Collaboration and app-server owner slices, both service-backed slices, backend boundary check, format, and Markdown lint. | Final owner graphs passed 32/32, 23/23, 24/24, and 17/17; boundary passed 3/3; checkpoint lint passed at `.cartulary/test-results/20260825T003556Z-p1914683`. | None. | Begin `WF-SL-04`; remove the remaining root replay-store surface and separate stream persistence from process orchestration. |
| 2026-08-24T20:37:09-04:00 | Codex `WF-SL-04` | Durable PostgreSQL storage and dispatcher process orchestration are separate private components. | Replaced root stream implementation with private store/dispatcher files and a temporary publication facade; updated runtime/routes/recovery verification/tests and this tracker. | Both Collaboration owner slices, targeted semantic row, boundary check, structural SQL/export audits, format, and Markdown lint. | Final owner graphs passed 32/32 and 23/23; boundary passed 3/3; checkpoint lint passed at `.cartulary/test-results/20260825T005304Z-p2133418`. | None. | Begin `WF-SL-05`; replace the temporary broad publication facade with explicit public effects while cutting Revisions to independent private facts. |
| 2026-08-24T20:54:06-04:00 | Codex `WF-SL-05` | Independent private revision facts and public Collaboration effects now cover every live owner and Revisions destructive path. | Added fact/publication inputs and immutable catalog; migrated application/source owners; removed combined/diff/suppression APIs; updated tests, generated projections, boundary policy, and this tracker. | All affected no-service/service owner slices, conflict/reconciliation browser rows, fast, format, generation, JSON, harness, boundary, static audits, and Markdown lint. | Final owner matrix and 430/430 fast graph passed; mutual imports and retired APIs are absent; checkpoint lint passed at `.cartulary/test-results/20260825T032502Z-p4104060`. | None. | Begin `WF-SL-06`; isolate recovery SQL, proof, lock, journal, and transaction outcomes behind the narrow Operator capability. |
| 2026-08-24T23:26:30-04:00 | Codex `WF-SL-06` | Operator sees only a narrow recovery capability; all recovery mechanics are Collaboration-private. | Added the private recovery adapter; reduced the root service to capability vocabulary/delegation; migrated Operator/tests; enforced root-mechanics and retired-constructor rules; updated this tracker. | Collaboration, Operator, Recovery, PostgreSQL, fast, targeted process/requeue, format, JSON, drift, boundary, static audits, and Markdown lint. | Final graphs passed with no compatibility change; exact run roots are in Section 8.1. | None. | Begin `WF-SL-07`; consolidate socket/intent/scenario test support and replace ordinary cross-owner table SQL. |
| 2026-08-24T23:51:30-04:00 | Codex `WF-SL-07` | Shared Collaboration test support is centralized; ordinary cross-owner tests no longer own Collaboration storage assertions. | Moved socket/intent helpers, removed the scenario wrapper and owner-local support root, added semantic assertion helpers, migrated all consumers, enforced support-path/SQL ownership, and updated this tracker. | All importing owner slices, shared test-utils, fast, format, JSON, harness, generated policy/drift, boundary, exact path/SQL audits, and Markdown lint. | Product and policy checks pass with test-only compatibility impact; exact run roots and classified repair runs are in Section 8.1. | None. | Begin `WF-SL-08`; enforce the final topology, run finalization and broad validation, and complete handoff. |
| 2026-08-25T01:10:32-04:00 | Codex `WF-SL-08` | Final topology is enforced and the remediation is complete. | Removed the dispatcher test export and final forwarding alias; replaced a stale deleted-file policy with executable lifecycle/export rules; repaired the last Network Flow test-composition caller; completed this tracker. | Exact symbol/import/SQL/path/public-contract audits; Collaboration, Network Flow, and Reporting owner slices; fast, format, JSON, harness, generated, frontend/backend boundary, finalization, and full check. | Final `make check` passed 656/656 and retained-evidence finalization passed; exact roots and the classified first broad failure are in Section 8.1. | None. | None; all workstreams are `DONE` and no required follow-up remains. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-24T17:26:00-04:00 | Codex planning session | Frontend session and Workbook coordinator are out-of-target contract consumers; Network Flow has a narrow interpreter; no backend grid-vendor coupling found. | Inspected incident session, Workbook coordinator/messages, Network Flow interpreter, protocol entrypoint; touched only this tracker. | `rg`, exact `sed` reads. | Freeze reconnect/reset, presence, row application, authorization recovery, and pending-edit anchoring. | Decoder projection repair requires later authorization. | Include protocol-ts and browser rows when SL-02 or route contracts change. |
| 2026-08-24T18:22:06-04:00 | Codex NLSpec-style tracker revision | Frontend remains an out-of-target consumer; decoder acceptance and non-effect requirements now cover typed output, dispatch, state, navigation, rendering, and telemetry. | Reused the inspected frontend consumer evidence and inspected protocol-ts generated/type surfaces; touched only this tracker. | `rg`, exact generated/entrypoint inspection. | `DEC-AC-01` through `DEC-AC-07` define the required frontend-facing characterization. | `RB-003` behavior-correction authorization. | Include the exact frontend consumer scope in the `SL-02` authorization. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-24T17:26:00-04:00 | Codex planning session | WS, requeue, recovery, and generated surfaces mapped. | Inspected authored WS/Collaboration contracts and generated consumers; touched only this tracker. | `jq`, `rg`, exact `sed` reads; `make explain-target` discovery. | Found owner/projection drift: replayable payload validators are closed against Core 01 additive-member policy. | `RB-003` | Under later authorization, repair authored projection, test decoder, run `make generate`, then drift checks. |
| 2026-08-24T18:22:06-04:00 | Codex NLSpec-style tracker revision | Decoder outcomes, authorization text, authored-first generation order, and generated drift acceptance are explicit. | Inspected Core 01 REQ-01-257/267, `contracts/ws/index.schema.json`, and generated protocol-ts types/validators; touched only this tracker. | `sed`, `jq`, `rg`. | The current closed job/extension object projections are separated from the owner-admitted additive behavior; no contract or generated file changed. | `RB-003`. | Obtain the Section 4.2 authorization before `SL-02`. |
| 2026-08-24T19:37:23-04:00 | Codex `WF-SL-02` | Complete server projection and non-mutating known-member decoder implemented. | Updated the authored WS contract, generator and validator, generated Go/TypeScript, protocol-ts/frontend consumers and tests, backend characterization, selector, topology index, and this tracker. | `make generate`; Protocol TS, frontend, Collaboration, browser-containing service slice, generated-state, harness, format, and Markdown targets. | Final runs passed 7/7 Protocol TS, 390/390 frontend unit, 32/32 Collaboration no-service, 23/23 service-backed, and all generated/harness guards. Checkpoint lint passed at `.cartulary/test-results/20260825T001317Z-p1690452`; classified repair runs are retained in Section 8.1. | None. | Begin `WF-SL-03`; keep public route and security behavior unchanged while introducing the runtime facade. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-24T17:26:00-04:00 | Codex planning session | 33 Collaboration owner rows discovered; local route inventory and evidence index contain stale entries. | Inspected target tests/testsupport, verification contract, owner/test-family catalogs; touched only this tracker. | `make task-guide`, `make explain-test-owner`, `make help`, target explanations, `jq`, `rg`, `sed`. | Canonical owner commands recorded; no product/harness validation was run during discovery. | `RB-002` | Run and classify owner baselines before implementation; repair accounting in SL-01. |
| 2026-08-24T18:22:06-04:00 | Codex NLSpec-style tracker revision | B0/B1, retained evidence, terminal states, passing/classified outcomes, hard blockers, and cross-slice acceptance are fully specified. | Inspected the Testing Harness terminal-state requirements and Collaboration/source-owner manifests; touched only this tracker. | `rg`, `sed`, owner-manifest file search. | No owner slice, product test, generation, harness, or conformance target was executed. | `RB-002`. | Establish B0 at the exact implementation base, then B1 after `SL-01`. |
| 2026-08-24T19:09:01-04:00 | Codex `WF-EXEC-00` | Tracker activated and B0 completed. | Touched only this tracker; inspected owner routing and retained run identities. | `make task-guide ROLE=module-author OWNER=module.collaboration`; `make explain-test-owner OWNER=module.collaboration`; both owner slices; boundary and generated-state guards; `make explain-run` for all six retained roots; `make lint-markdown`. | No-service 32/32, service-backed 23/23, boundary 3/3, generation drift 4/4, generated policy 3/3, and JSON shape 3/3 passed with no non-pass rows. Markdown lint passed at `.cartulary/test-results/20260824T231517Z-p1147708`. | None. | Begin `WF-SPEC-01`; do not begin accounting repair until its owner adoption and checkpoint pass. |
| 2026-08-24T19:24:57-04:00 | Codex `WF-SL-01` | Accounting repair and final-source B1 completed. | Removed the prose evidence test, corrected the socket inventory, updated the authored owner selector, generated the topology index through Make, and updated this tracker. | Owner slice, harness contract, B1 owner slices, failed drift investigation, `make generate`, final drift/policy/harness checks, final-source B1 reruns, retained-run inspection, `make lint-markdown`. | The final source digest is `sha256:f07488ebfbff12f6687a336cec01dd3fd030c01faefd78474fdf2289b1311184`; B1 passed 32/32 and 23/23. The one failed drift run was caused and resolved by the expected generated harness index update. Markdown lint passed at `.cartulary/test-results/20260824T233618Z-p1404765`. | None. | Begin authorized `WF-SL-02`; use B1 as its comparison point. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-24T17:26:00-04:00 | Codex planning session | Current-membership delivery, session revocation, incident closure, origin policy, and deployment-local requeue authority are frozen. | Inspected Core 04, routes, notifier, recovery/operator surfaces and tests; touched only this tracker. | `rg`, exact `sed` reads. | No authorization move is proposed; Platform supplies primitives while Collaboration owns route semantics. | None beyond later implementation authorization. | Preserve security precedence and negative surfaces in every slice. |
| 2026-08-24T18:22:06-04:00 | Codex NLSpec-style tracker revision | Pre-production cutover rules and decoder handling explicitly preserve authorization, route, duplicate, framing, and non-execution boundaries. | Reused inspected Core 04, route, notifier, recovery, and operator evidence; touched only this tracker. | Exact source review and tracker diff inspection. | No security owner, public surface, or local-authority rule changed. | Later implementation authorization only. | Apply `CT-REQ-001`, `CT-REQ-002`, and `CT-REQ-010` to every slice. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-24T17:40:09-04:00 | Codex planning session | Planning inventory, findings, workflows, slices, command discovery, and tracker validation are complete. | Touched only `docs/handoffs/collaboration-module-refactor-tracker.md`. | `make lint-markdown` | Passed and revalidated; latest recorded run root `.cartulary/test-results/20260824T214025Z-p1019868`. No production refactor was performed and implementation remains deferred. | `RB-001`, `RB-002`, `RB-003` | Hand off for a separately authorized implementation/design session. |
| 2026-08-24T18:22:06-04:00 | Codex NLSpec-style tracker revision | The tracker now has precise normative posture, proposed interfaces, defaults, mappings, baseline evidence, decoder authorization, and binary acceptance. | Touched only `docs/handoffs/collaboration-module-refactor-tracker.md`; prior handoff rows remain intact. | Read-only discovery/diff checks; `make lint-markdown`. | Markdown lint passed with run root `.cartulary/test-results/20260824T222840Z-p1033205`. No production, test, contract, generated, migration, package, or harness file changed. | `RB-001`, `RB-002`, `RB-003`. | Hand off for owner adoption, B0/B1 execution, and explicit decoder-correction authorization. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Has the fact-separation and port recommendation in Sections 3.1-3.4 been adopted by the controlling boundary owners? | Without adoption, implementation or research could silently choose ownership. | An amendment to the Revisions decision, an adopted Collaboration decision, Core adoption, and traceability. | **RESOLVED:** REQ-00-075 adopts the Collaboration decision; the Revisions decision is amended; AC-564 and traceability are recorded. Implementation remains sequenced. |
| RB-002 | What is the retained B0/B1 passing or fully classified baseline for the Collaboration owner slices? | Without exact per-row closure, existing product or harness failures can be misattributed to a slice and rollback decisions become unsafe. | Section 8.1 source, selection, invocation, run, row, failure, classification, and accountability evidence from both canonical owner-slice commands. | **RESOLVED:** B0 and final-source B1 both pass; exact run roots and the generated-index drift classification are recorded. |
| RB-003 | Has a later task recorded the exact Section 4.2 authorization for the additive replayable-payload decoder correction? | The correction aligns downstream projections with Core 01 but changes observable client acceptance, typed decoder output, authored contracts, tests, and generated artifacts. | Explicit authorization naming scope and owners, followed by `DEC-AC-01` through `DEC-AC-07` and Make-owned generation/validation. | **RESOLVED:** the 2026-08-24 remediation request explicitly authorizes `G-03` and `WF-SL-02`; implementation remains sequenced behind specification adoption and B1. |

There is no `BLOCKED: owner contradiction` finding in the inspected scope.

## 12. Planning Baseline Completion Criteria

This section records the completed pre-authorization planning baseline. Its
historical statements do not describe the current execution worktree; current
workstream closure is controlled by Sections 1.3, 8.1, 9, and 10.

- [x] Every file in `internal/modules/collaboration` is inventoried; no target
  file is silently out of scope.
- [x] Every discovered public contract risk has an owner and test posture.
- [x] Every proposed workflow has dependencies and a handoff/exit checkpoint.
- [x] Every proposed implementation slice is behavior-preserving unless it is
  explicitly marked `requires later authorization`.
- [x] Every load-bearing tracker requirement has a stable ID, authority/status
  posture, applicable slice or blocker, binary acceptance posture, and
  validation path.
- [x] The candidate Revisions/Collaboration interfaces name every required
  input, ownership boundary, omission/default rule, duplicate rule,
  transaction rule, and failure consequence needed for owner adoption.
- [x] `RB-001` is decision-complete for planning and explicitly blocked on
  owner adoption; the tracker does not claim that research supersedes the
  adopted Revisions boundary.
- [x] `RB-002` defines B0/B1, the complete retained evidence set, the closed
  terminal-state vocabulary, passing/classified outcomes, and hard blockers,
  without claiming that a baseline ran.
- [x] `RB-003` contains the exact authorization boundary, decoder outcome map,
  characterization matrix, and generated-artifact sequence without granting
  authorization itself.
- [x] The pre-production default requires reset-only incompatible-state
  handling and atomic cutover without compatibility aliases, dual paths,
  feature flags, backfills, or production-data migrations.
- [x] The document-placement table separates behavioral owners, topology
  owners, downstream projections, generated outputs, retained evidence, and
  non-authoritative rationale.
- [x] `CT-AC-01` through `CT-AC-10` and `DEC-AC-01` through `DEC-AC-07` provide
  binary, owner-routed acceptance coverage for the newly specified gaps.
- [x] Validation commands were discovered from public Make targets; none were
  invented.
- [x] No owner contradiction was found; future contradictions must be written
  as `BLOCKED: owner contradiction` without choosing a side.
- [x] Repository/framework mismatches and live projection/accounting drift are
  recorded as planning findings.
- [x] Handoff sections contain the baseline, inspected surfaces, command
  discovery, blockers, and next actions needed to continue without repeating
  repository discovery.
- [x] `make lint-markdown` passes for the NLSpec-style revision; run root
  `.cartulary/test-results/20260824T222840Z-p1033205`. The earlier tracker run
  `.cartulary/test-results/20260824T214025Z-p1019868` remains historical
  evidence only.
- [x] The only file touched by this planning task is this tracker.
- [x] The tracker explicitly states that no production refactor was performed
  and that implementation requires later authorization.

At planning completion, all criteria in this historical section passed and
`RB-001` through `RB-003` remained future gates. The authorized execution
record above subsequently resolved those gates and completed every
slice-specific criterion.
