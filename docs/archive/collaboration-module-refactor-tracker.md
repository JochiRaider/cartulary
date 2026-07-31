# collaboration Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Value |
| --- | --- |
| Target path | `internal/modules/collaboration` |
| Target label | `collaboration`, derived from the final target-path component and normalized to lowercase kebab case |
| Output path | `docs/handoffs/collaboration-module-refactor-tracker.md` |
| Repository state inspected | Execution baseline: `main` at `c26f8ff1e1344a8cf8c14d143a74a7f82595be35`; worktree clean at the start of the authorized remediation. Final validation covers the complete authorized worktree of 176 changed paths: 158 tracked paths and 18 new paths. |
| Execution status | End-to-end remediation complete; WS-00 and SL-00 through SL-07 are `DONE` |
| Controlling artifact | This tracker is updated after every workstream and before the next workstream begins |
| Non-goals | No database migration, new Collaboration NLSpec, new public route, or JSON message-shape change; `docs/domain.md` remains unchanged |
| Authorization | The 2026-07-27 implementation task authorizes WS-00 and SL-00 through SL-07, subject to each slice gate and owner precedence |

The target exists and contains 27 files after the SL-03 capability split added
`recovery.go` and SL-05 removed `.gitkeep`. This tracker treats the planning
framework as doctrine and a template, not as evidence of current repository
state. Live source, contracts, tests, ownership projections, and Make-owned
command discovery were inspected before recording repository claims.

Normative language in this tracker constrains only the authorized refactor and
its acceptance boundary. It does not create, amend, or supersede public product
requirements. Execution MUST apply this tracker beneath the authority order
below, MUST preserve owner-defined observable behavior except for the explicit
Core-adopted G-01 compatibility cleanup, and MUST stop with `BLOCKED: owner
contradiction` if an adopted owner document conflicts with a requirement
recorded here.

The authority order used here is:

1. adopted subsystem NLSpecs within their named subsystem;
2. Core 00 through Core 04 for implementation-conformance behavior;
3. Core 05 only for claim-bearing timed or fixture-sensitive publication;
4. domain vocabulary, repository boundaries, and implementation-support
   guides;
5. current repository code and tests;
6. prior plans, handoffs, and the planning framework as evidence only.

Owner documents inspected:

- `AGENTS.md`;
- `docs/spec/00_document_set_status_and_precedence.md`;
- `docs/spec/01_architecture_storage_and_view_contracts.md`, especially the
  Collaboration WebSocket contract;
- `docs/spec/02_domain_model_schema_and_history.md`, for revision and
  append-only history boundaries;
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`,
  especially the Collaboration-owned sequencer, live-session behavior, and
  frontend pending-queue rules;
- `docs/spec/04_security_deployment_and_conformance.md`, for session,
  authorization, Origin, test-route, and incident-membership rules;
- `docs/extension-subsystem-nlspec.md`, for typed
  `extension_resource_changed` ownership and WebSocket invalidation;
- `docs/network-flow-activity-nlspec.md`, for the Network Flow source-owner
  boundary;
- `docs/opentelemetry-instrumentation-nlspec.md`, for Collaboration
  instrumentation, signal names, attributes, and privacy;
- `docs/testing-harness-nlspec.md`, for Make ownership, owner slices,
  generated artifacts, and evidence accounting;
- `docs/domain.md`, for domain vocabulary and bounded-context navigation.

Core 05 was considered and is not applicable to this structural,
non-claim-bearing plan. No timed or fixture-sensitive publication is proposed.

Supporting planning material inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, used as
  planning doctrine and structure only;
- `docs/research/nlspec-spec.md`, used for prescriptive language,
  completeness, explicit defaults, mapping tables, and binary acceptance
  criteria only;
- `temp/analysis-notes.md`, used as design-analysis evidence for RB-002 and
  RB-003 only;
- `docs/guides/cartulary-dev-guide.md`, whose current platform WebSocket
  ownership description is narrowed in SL-04 after the implementation moves.

These supporting documents MUST NOT be treated as behavioral authorities.
WS-00 amends Core 01, Core 03, and Core 04 for the explicit G-01 wire-resource
and live-intent rules before implementation begins. Core 02 requires no
amendment because no logical or physical storage change is planned.
`docs/domain.md` remains unchanged because no vocabulary or bounded-context
decision changes.

Repository evidence inspected includes:

- every file under `internal/modules/collaboration`;
- application composition in `internal/app/server`,
  `internal/app/operator`, `internal/app/timelineassembly`, and
  `internal/app/revisionassembly`;
- Collaboration callers in Revisions, Incident Bundles, Entities, Evidence,
  Timeline, Incidents, Jobs, and Network Flow;
- `internal/platform/ws`, including its presence, delivery, message, payload,
  and transport tests;
- the Collaboration durable-stream migrations and
  `tools/schema_object_ownership_manifest.json`;
- `contracts/ws/index.schema.json`,
  `contracts/protocol-ts/frontend-entrypoints.v1.json`, the protocol
  generator, and generated TypeScript output locations;
- frontend session, workbook collaboration-effect, and browser transport
  consumers;
- `contracts/verification/owners/module.collaboration.json`,
  `tools/test_families/module.collaboration.json`,
  `tools/test_families/web.collaboration.json`,
  `tools/test_support_inventory.json`, and relevant backend/frontend boundary
  projections.

No owner contradiction was discovered. If a later session finds one, it must
record `BLOCKED: owner contradiction` and stop the affected design choice.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/collaboration/.gitkeep` | Removed obsolete directory placeholder | None | None | None | None | None | Collaboration repository housekeeping | none | Deleted in SL-05 after the final package layout stabilized; no repository reference remains. |
| `internal/modules/collaboration/codec.go`, `codec_test.go` | Sole Collaboration semantic encoder/strict decoder and exact byte/failure evidence | `Codec`, decoded client message types, semantic decode errors | Collaboration routes and replay/live delivery | Standard JSON and Collaboration protocol registry only | Codec and socket failure tests | Existing WS shapes remain unchanged | Collaboration wire semantics | critical | Added in SL-04; text-only, duplicate-member rejection, and LF-free output are complete. |
| `internal/modules/collaboration/protocol.go`, `protocol_test.go` | Collaboration DTOs, payload validation, patch construction, socket port, semantic Hub, and presence | Collaboration messages/payloads, `Socket`, `AcceptSocket`, `CheckBrowserOrigin`, `Hub` | Collaboration routes, server adapter, source-owner/test consumers | Standard library, UUID, and transport-neutral ports | Protocol and Hub tests plus service-backed consumers | Existing WS contracts and generated TypeScript are unchanged | Collaboration semantic lifecycle | critical | Added in SL-04; no platform semantic alias or second Hub remains. |
| `internal/modules/collaboration/hub_telemetry.go`, `hub_telemetry_test.go`, `hub_test.go` | Hub delivery telemetry, safe public event classification, slow-consumer and ordering evidence | Hub methods and package-private telemetry helpers | Collaboration Hub | Platform telemetry primitives only | Exact Hub/telemetry unit tests and OTel conformance | OTel fixtures and conformance inputs | Collaboration telemetry and live delivery | high | Added in SL-04; event types derive from the Collaboration registry. |
| `internal/modules/collaboration/historical_restore_test.go` | Direct transaction-local suppression and leakage evidence | Test symbols only | Collaboration owner routing | Managed Postgres fixture | The file itself | No generated output | Collaboration verification | medium | Added in SL-00 and retained after the SL-01 policy cutover. |
| `internal/modules/collaboration/historical_restore.go` | Owns the transaction-local historical-intent setting and exposes suppression/query capabilities without leaking its literal | `HistoricalIntentPolicy`, `NewHistoricalIntentPolicy`, `SuppressTx`, `IsSuppressedTx`, `SuppressSQLTx` | Revision assembly and Incident Bundles composition/tests | `pgx.Tx` and a narrow SQL transaction fixture port | Direct transaction-local isolation, Incident Bundles import, and ordinary restore/rollback tests | Collaboration durable-intent behavior; no generated artifact | Collaboration policy exposed through consumer-owned ports | medium | Implemented in SL-01; no setting literal exists outside this file. |
| `internal/modules/collaboration/incident_session_notifier.go` | Converts incident close and membership revocation into active-session termination and WebSocket notifications | `IncidentSessionNotifier`, `NewIncidentSessionNotifier`, `NotifyIncidentClosed`, `NotifyIncidentMembershipRevoked` | Server runtime composition and Incidents collaboration-session port | `platform/authn`, `platform/postgres`, Collaboration `Hub` | Collaboration revocation integration tests and Incidents tests | Core WebSocket `error` and `session_revoked` contracts | Collaboration session orchestration | high | Collaboration owns semantic session termination; transport mechanics remain behind the socket adapter. |
| `internal/modules/collaboration/integration_test.go` | Real-server characterization for two-client presence/replay, revocation sources, closed incidents, replay filtering, and Origin rejection | Test symbols only | Collaboration owner test catalog | Auth, Incidents, Jobs, Timeline, platform WebSocket, shared test support | The file itself | `contracts/ws/index.schema.json`; module Collaboration evidence rows | Collaboration verification | high | Service-backed behavior evidence. |
| `internal/modules/collaboration/intent_validation_test.go` | Exact append-time semantic validation for all replayable event families, including additive-member compatibility | Test symbols only | Collaboration owner routing | Collaboration intent and payload APIs | The file itself | Authored Collaboration family routing | Collaboration verification | high | Added and routed in SL-02. |
| `internal/modules/collaboration/producer_catalog.go` | Removed in SL-01; its mapping is now source-owned and compiled by Revisions | None | None | None | Replacement catalog tests live in `internal/app/revisionassembly` | Public identifiers remain unchanged | Revisions coordination over source-owner contributions | none | No fallback or runtime reference remains. |
| `internal/modules/collaboration/producer_catalog_test.go` | Removed in SL-01 after equivalent and stronger immutable-catalog evidence moved to revision assembly | None | None | None | `TestRevisionsRuntimeBuildsExactImmutableRecordViewCatalog` and failure-matrix tests | Verification routing moved to `module.revisions` | Revisions verification | none | The former characterization row was replaced rather than left stale. |
| `internal/modules/collaboration/record_change_intent.go` | Canonicalizes changed cell keys and builds deterministic `record_changed` semantic intents | `RecordChange`, `ChangedCellKeys`, `NewRecordChangeIntent` | Revisions; Entities mentions and merge; Evidence; Timeline assembly and source-owner adapters | UUIDs and Collaboration-owned protocol helpers | `record_change_intent_test.go`, shared harness, stream/integration tests, source-owner tests | `record_changed` in `contracts/ws/index.schema.json` and generated protocol TypeScript | Collaboration semantic intent construction | high | Semantic ownership and deterministic identity are complete. |
| `internal/modules/collaboration/record_change_intent_test.go` | Characterizes sorted, compact patch construction and deterministic intent identity | Package-private test helper | `TestSocketEventInventoryCoverage` in `shared_harness_test.go` | Collaboration public intent API | The file itself through shared harness | `record_changed` payload and field-key conformance | Collaboration verification | medium | Accounted indirectly through the shared harness row. |
| `internal/modules/collaboration/recovery.go` | Operator-only incident-stream quarantine recovery | `RecoveryService`, `NewRecoveryService`, `RequeueIncident` | Operator application facade; durable-stream integration evidence | Postgres transaction capability only | Durable-stream quarantine/requeue and operator argument tests | Collaboration durable-stream storage; no generated artifact | Collaboration recovery capability | high | Split from `stream.go` in SL-03; it exposes no append, replay, dispatch, or token authority. |
| `internal/modules/collaboration/routes.go` | Registers and serves the incident WebSocket route, handshake/resume, replay, presence, heartbeat, authorization rechecks, terminal errors, and close behavior | `Service`, `Settings`, `RegisterRoutes` | `internal/app/server/runtime.go`, `internal/app/server/module_settings.go` | Incidents access, Authn, HTTP API/auth, raw socket port, Collaboration Hub, replay capability | `socket_test.go`, `integration_test.go`, testsupport clients, browser/e2e tests | Full Collaboration WebSocket contract and generated message types | Collaboration application adapter | critical | SL-04 removed direct third-party/platform WebSocket semantics; SL-03 narrows the remaining replay dependency. |
| `internal/modules/collaboration/shared_harness_test.go` | Validates semantic harness inventory and indexes live tests for socket obligations | `TestSocketEventInventoryCoverage`, `TestSocketLifecycleEvidenceIndex` | Collaboration owner test catalog | Owner-local incident WebSocket test support | The file itself plus referenced tests | Verification/accounting projections only | Collaboration verification | medium | Evidence routing is not runtime architecture. |
| `internal/modules/collaboration/socket_test.go` | Service-backed route tests for closed message vocabulary, resume reset, heartbeat expiry, and incident-scoped ephemeral presence | Test symbols; test-only `SessionTiming` | Collaboration owner test catalog | Auth, Incidents, Timeline, platform WebSocket, owner-local test support | The file itself | WebSocket contract and security behavior | Collaboration verification | high | Despite `_Unit` names, catalog rows use managed-service integration profiles. |
| `internal/modules/collaboration/stream.go` | Owns narrow intent admission, route-facing replay/token persistence, dispatcher-private sequencing/retry/quarantine/retention persistence, and their semantic validation | Event-family constants, `ErrIntentKeyCollision`, `EventIntent`, `IntentAppender`, `NewIntentAppender`, `NewEventIntent`, `ReplayStore`, `NewReplayStore`, `ReplayResult`, `Dispatcher`, `NewDispatcher`, `Start`, `Close`, `RunOnce` | One server-composed append capability shared by producer adapters; Collaboration routes; server dispatcher lifecycle | Postgres/pgx, crypto, Collaboration protocol validation | Intent validation, stream integration, route/socket/integration, and source-owner tests | Collaboration DB ownership manifest; replayable WebSocket event contracts | Collaboration durable stream with least-capability construction | critical | SL-03 removed the broad `Store`, `NewStore`, and exported free append function; the dispatcher repository is private and recovery is isolated in `recovery.go`. |
| `internal/modules/collaboration/stream_integration_test.go` | Characterizes atomic insertion, exact replay/collision, post-commit sequencing, outage/restart, fan-out, quarantine/requeue, hashed tokens, and conjunctive retention | Test symbols only | Collaboration owner test catalog | Postgres, server harness, Timeline, Auth, platform WebSocket | The file itself | Durable-stream DB contract and replay behavior | Collaboration verification | critical | Primary durable-stream behavior evidence. |
| `internal/modules/collaboration/telemetry.go` | Emits Collaboration WebSocket lifecycle and dispatcher-run telemetry using allowlisted vocabulary | Package-private methods on `Service` and `Dispatcher` | `routes.go`, `stream.go` | Platform HTTP API and telemetry; OpenTelemetry API | `telemetry_test.go`; OTel conformance references | OTel source snapshot/generated constants are downstream owner projections | Collaboration instrumentation using platform telemetry primitives | high | Signal names and privacy rules are observable operational contracts. |
| `internal/modules/collaboration/telemetry_test.go` | Characterizes public-error classification and closed telemetry vocabulary | Test symbols only | Go package test execution; OTel conformance source reference | Collaboration telemetry helpers, HTTP API errors | The file itself | OTel conformance and generated-constant evidence | Collaboration/Telemetry verification | medium | Current Collaboration owner-family rows do not name these tests. |
| `internal/modules/collaboration/testsupport/incidentwstest/incidentwstest.go` | Owner-local real WebSocket client, hello/resume parsing, dial assertions, terminal-event assertions, and presence sorting | `ConnectOptions`, ack DTOs, `Client`, connection/assertion helpers | Collaboration, Auth, Entities, Evidence, Incidents, Links, Timeline, Workbook, and other integration tests | platform WebSocket messages, generic `internal/testutil` HTTP/WebSocket helpers | Broad integration suites | WebSocket contract as test support; no generated output | Collaboration owner-local test support | medium | Explicitly registered as `owner_local`, service-starting support in `tools/test_support_inventory.json`. |
| `internal/modules/collaboration/testsupport/incidentwstest/inventory.go` | Declares semantic harness capabilities required by Collaboration socket evidence | Harness ID/requirement types, constants, inventory functions | `shared_harness_test.go` | Testing only | Shared harness tests | Harness/accounting projections only | Collaboration owner-local test support | low | Verification inventory must not become product authority. |
| `internal/modules/collaboration/testsupport/incidentwstest/view_events.go` | Connects view sockets and waits for `record_changed` events without leaking transport details to source-owner tests | `RecordChangeSocketPayload`, `ConnectViewSocket`, `RequireRecordChanged`, `ExpectNoSocketMessage` | Entities, Evidence, Links, Timeline, Workbook, and related integration tests | platform WebSocket and generic test utilities | Importing source-owner integration tests | `record_changed` contract as test support | Collaboration owner-local test support | medium | Remains semantic owner support unless generalized beyond Collaboration. |
| `internal/modules/collaboration/testsupport/intenttest/intents.go` | Collaboration-owned durable-intent inspection and legacy Jobs v1 fixture support | Intent record loader and v1/v2 coexistence helpers | Jobs producer integration tests | Managed Postgres fixture only | Jobs producer routing | No product contract | Collaboration owner-local test support | medium | Keeps Collaboration table knowledge out of Jobs and Network Flow tests as well as production code. |
| `internal/modules/collaboration/testsupport/scenariotest/harness.go` | Starts shared application/runtime fixtures for Collaboration scenarios | `RuntimeHarness`, `ServerHarness`, `StartRuntime`, `StartServer` | Collaboration service-backed tests | `internal/testutil/appsupport`, HTTP API, HTTP test helpers | Collaboration integration/socket/stream tests | Harness execution only | Collaboration owner-local test support | medium | Explicit service-starting support; not production assembly. |

The original 19-file diagnosis plus the execution-delta rows above account for
all 27 current files, the SL-05 placeholder deletion, and the two SL-01
deletions. No current file is omitted.

## 3. Module Boundary Diagnosis

The current target is a legitimate Collaboration module with a
mixed-responsibility implementation package. It is not merely an accidental
catch-all: durable per-incident sequencing, replay, presence/session
orchestration, and Collaboration telemetry have coherent owners in Core and the
adopted telemetry subsystem. It is also not a thin facade because it contains
substantial persistence, dispatch, route, authorization-adjacent, protocol, and
catalog logic.

The target is currently:

- a durable Collaboration application/service boundary;
- a WebSocket transport-adjacent adapter;
- a persistence-adjacent adapter;
- a collaboration mutation-notification coordinator, but not the owner of
  source mutations;
- partially a view/projection invalidation orchestration layer;
- a mixed-responsibility package with some logic owned by source modules or
  obscured by platform dependencies.

It is not a frontend shell/controller or grid-vendor integration layer.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Durable semantic intent admission, per-incident sequencing, replay, retries, quarantine, and requeue | `stream.go` and Collaboration DB tables | Collaboration | keep, then split internal capabilities | Core 03 Collaboration-owned durable sequencer; migrations and schema ownership manifest | Preserve post-commit sequence assignment and source-transaction atomicity. |
| Incident socket handshake, resume, heartbeat, and terminal behavior | `routes.go` | Collaboration application adapter over platform transport | keep | Core 01 WebSocket contract; current route tests | Exact route and envelopes remain frozen. |
| Presence and live delivery semantics | `routes.go` plus `internal/platform/ws.Hub` | Collaboration semantics; platform transport/runtime mechanics | split | Core 01 and Core 03; platform Hub implementation and tests | RB-003 defines the dependency-safe split; implementation begins in SL-04 after SL-00 characterization. |
| Record-change canonicalization and deterministic event construction | `record_change_intent.go` | Collaboration | keep | Core 01 event contract; Core 03 sequencer rule; source-owner callers | Platform-owned semantic payload helpers should be narrowed later. |
| Record type and artifact prefix to public view-schema mapping | `producer_catalog.go` | Source-owner `revisions.ProviderContribution` values aggregated by `revisionassembly`, validated by Revisions | move | Existing `RevisionProviderContribution()` producer functions, `ProviderContribution`, projection descriptors, and repository composition rule | RB-002 defines one immutable catalog and one appender per composition. Public `view_schema_id` values MUST NOT change and dual event writers are prohibited. |
| Historical live-intent suppression during bundle restore | `historical_restore.go`, read by Revisions | Collaboration port used by Incident Bundles/Revisions | keep / split | Incident Bundles caller and transaction-local setting | Add direct characterization before changing the seam. |
| Incident close/membership revocation notification | `incident_session_notifier.go` | Collaboration facade with an Incidents-owned narrow port | keep | Server composition, Incidents tests, revocation integration tests | Do not move incident authorization decisions into Collaboration. |
| Job progress semantic intent construction | `internal/platform/jobs/collaboration_producer.go` | Jobs semantics with Collaboration persistence port | keep split boundary | Core 03 source-owner intent rule; v2/coexistence/rollback evidence; SQL guard | Implemented in SL-02: Jobs retains transaction, payload, and key ownership through its configured `TransactionService`. |
| Network Flow invalidation semantic intent construction | `internal/modules/networkflow/collaboration_producer.go` | Network Flow semantics with Collaboration persistence port | keep split boundary | Extensions and Network Flow NLSpecs; replay/privacy/rollback evidence; SQL guard | Implemented in SL-02 through `ModuleDependencies.ResourceIntents`; payload/key/privacy remain Network Flow-owned. |
| WebSocket messages, payload validators, patch builders, presence, and Hub | `internal/modules/collaboration` | Collaboration | keep | Collaboration protocol/Hub/codec tests and SL-04 owner evidence | Implemented in SL-04 without aliases or a generated-contract change. |
| Generic WebSocket frame I/O, upgrade, Origin mechanics, and close mechanics | `internal/platform/ws` plus the server socket adapter | Platform transport and application composition | keep split boundary | Platform/server owner slices and boundary guards | Platform has no Collaboration semantic symbol or module import. |
| Frontend reconnect/resume/dedupe and workbook pending edits | `apps/web/src/collaboration` and workbook collaboration services | Frontend Collaboration session and Workbook controllers | keep | Core 03 and live frontend tests | Backend refactor must not retarget pending edits or move client state server-side. |
| Collaboration test fixtures and semantic assertions | `internal/modules/collaboration/testsupport` | Collaboration owner-local test support | keep | `tools/test_support_inventory.json` | `intentional/no_action` unless helpers become owner-neutral. |
| Grid rendering/vendor adaptation | Grid adapter and Workbook frontend packages | Existing frontend owners | keep outside target | Import search found no direct target/grid-vendor coupling | `intentional/no_action`. |

### 3.1 Required Catalog Owner Graph

The later RB-002 implementation MUST realize exactly this dependency direction:

| Producer or layer | May know | Must provide | Must not own |
| --- | --- | --- | --- |
| Source-owner modules | Their record types, artifact variants, revision providers, and public view identities | Data-only `RecordViewRouteContribution` values in the existing `revisions.ProviderContribution` | A mutable registry, Collaboration persistence, public sequence assignment, or another event writer |
| `internal/app/revisionassembly` | Source-owner contributions, projection descriptors, generated view-schema registry, Revisions, and the Collaboration intent sink | One validated `Runtime` containing one immutable catalog, one shared `*revisions.Appender`, and the command service | Source semantics, fallback mapping, or package-global registration |
| Revisions | The compiled catalog and a consumer-owned intent-appender interface | Catalog validation, record/view resolution, revision append coordination, and one intent append request | A second semantic event family, durable stream, or Collaboration storage implementation |
| Collaboration | `RecordChange`, deterministic intent rules, and durable intent storage | The sole durable intent admission path, sequencing, replay, and dispatch | Source-owned record/view declarations or revision coordination |

The permitted production import direction is source owner to Revisions,
`revisionassembly` to source owners plus Revisions plus composition
dependencies, and Revisions to Collaboration's semantic intent surface.
Collaboration MUST NOT import Revisions or source-owner packages.

### 3.2 Required WebSocket Owner Graph

| Layer | Required ownership | Prohibited ownership |
| --- | --- | --- |
| `internal/platform/ws` | HTTP upgrade, configured Origin mechanics, generic frame read/write, explicit message-size limit, write serialization, low-level control frames, and generic close mechanics | Incident, session, membership, resume, presence, record, view, job, extension, or Collaboration DTO semantics |
| `internal/modules/collaboration/wire` | The sole Collaboration envelope, DTO family, validation, patch construction, and JSON codec | HTTP upgrade, third-party WebSocket connection types, authorization lookup, Hub state, or durable storage |
| `internal/modules/collaboration` | Consumer-owned socket port, Hub, presence, hello/resume, replay/live delivery, application heartbeat, authorization choreography, terminal behavior, and Collaboration telemetry | HTTP listener mechanics or a direct third-party WebSocket import |
| `internal/app/server` | Concrete platform-to-Collaboration socket adapter and dependency composition | A second codec, Hub, semantic state machine, or route contract |

The semantic types MUST move rather than be copied. Slice completion permits
one codec and one Hub only; platform aliases, compatibility DTOs, duplicate
delivery paths, and dual-write migration modes are prohibited.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /ws/v1/incidents/{incident_id}` route and cookie-authenticated upgrade | Core 01/04; Collaboration adapter | `routes.go`, WS contract index | socket/integration tests; browser socket transport tests | Preserve exact method/path, error envelopes, and Origin rejection | critical | No second WebSocket mutation API may appear. |
| First application message and hello/resume acknowledgements | Core 01; Collaboration | Route state machine and contract schemas | handshake/resume unit and integration tests | Preserve closed first-message vocabulary and ack envelopes | critical | Repeated or invalid first messages remain rejected. |
| Resume-token binding, replay, high-water, and reset-required behavior | Core 01/03; Collaboration durable stream | `stream.go`, route logic, DB migrations | socket and durable-stream integration tests | Preserve invalid, expired, mismatched, future, and too-old reset cases with no partial replay | critical | Tokens remain hashed at rest and scoped to incident/client/session. |
| Presence snapshot/delta and ephemeral incident scoping | Core 01/03; Collaboration semantics | Route logic and platform Hub | socket, integration, platform presence tests, browser tests | Preserve deterministic sorting, expiry, removal, and non-replayability | high | Presence is not durable incident state. |
| Heartbeat and connection expiry | Core 01/04; Collaboration | Route timers and inbound-message tracking | heartbeat/idle tests | Preserve 15-second ping intent, 45-second inbound-dead posture, and no auth-expiry sliding | high | Timer implementation may move; behavior may not. |
| Current-session and current-membership authorization | Core 04; Auth/Incidents with Collaboration rechecks | Route and notifier code | revocation, membership, logout, concurrency, and close tests | Preserve recheck before delivery and terminal reason codes | critical | Deployment admin is not an incident-access bypass. |
| Closed incident terminal behavior | Core 01/03/04; Incidents and Collaboration | Route terminal handling and notifier | closed-incident integration and browser stateful tests | Preserve terminal error and close before writable ack | critical | No mutation acknowledgement after closure. |
| `record_changed` envelope, canonical changed keys, patch/invalidate/remove | Core 01/03; Collaboration plus source owners | WS schema, `record_change_intent.go`, platform payload helper | shared harness, platform presence tests, replay integration, Workbook tests | Add exhaustive source-owner view-schema mapping and historical-suppression characterization | critical | Entity, Evidence, Timeline, artifact, and other row behavior must remain unchanged. |
| Public `view_schema_id` routing and projection invalidation | Source view owners; Collaboration event contract | producer catalog, projection/workbook catalogs, frontend effects | artifact catalog tests; source-owner and Workbook tests | Add complete mapping parity across every current record owner and artifact subtype | high | Saved-view semantics are not owned here; only event routing to affected views is frozen. |
| `job_progress` payload and replayability | Jobs semantics; Collaboration delivery | Jobs producer, WS schema, durable stream | platform WS contract tests; Collaboration replay tests | Add source-producer atomicity, exact replay, and collision characterization | high | Job-private leases and payloads remain absent. |
| `extension_resource_changed` for Network Flow | Extensions/Network Flow semantics; Collaboration delivery | Network Flow producer, WS schema, Extensions NLSpec | Network Flow route integration and platform WS tests | Preserve rename/remove reason, workspace refs, privacy, atomicity, and exact replay | high | No generic callback bus or label disclosure. |
| Intent insertion, deterministic identity, and collision behavior | Collaboration persistence; source owners create semantics | `AppendIntentTx` and direct Jobs/Network Flow writers | durable-stream integration tests | Characterize direct producer parity before centralizing persistence | critical | Source write and intent insert remain one transaction. |
| Dispatcher ordering, retries, quarantine, and requeue | Collaboration | `stream.go`, Operator composition | durable-stream integration and operator tests | Preserve stable event ID/sequence through outage/restart and explicit requeue | critical | Public sequence is assigned only after source commit. |
| Replay retention and cleanup | Core 01 minimum; Collaboration storage | `stream.go` retention logic and migrations | conjunctive retention integration tests | Preserve at least five minutes or 10,000 incident events, whichever retains more | high | Current safety caps must not weaken the Core minimum. |
| Historical restore suppression | Collaboration with Incident Bundles/Revisions | transaction setting and callers | indirect rollback/import evidence | Add direct no-live-intent-on-historical-restore test | high | Historical import must not impersonate a live mutation. |
| Collaboration telemetry | OpenTelemetry NLSpec; Collaboration instrumentation | `telemetry.go`, telemetry registry | telemetry unit tests and OTel conformance references | Preserve registered span/metric names, vocabulary, and forbidden identifiers | high | Telemetry failure must not alter product behavior. |
| Authored and generated Collaboration protocol | Core 01; `contracts/ws` authored projection; protocol package generation | WS index, frontend entrypoint owner, generator | platform contract tests, frontend compile/tests, generate drift | No schema change planned; run drift if imports or authored inputs change | high | Never hand-edit generated TypeScript. |
| Frontend session and Workbook live-row behavior | Core 03; frontend Collaboration/Workbook owners | session component, socket transport, collaboration effects | Vitest, browser, stateful, accessibility, and visual rows | Preserve reconnect/dedupe/gap/reset, pending drafts, save-state, focus, and selectors | critical | No direct grid-vendor coupling was found in the backend target. |
| Harness/test accounting | Testing Harness NLSpec; authored owner/family catalogs | verification owner and family manifests | owner slice explanation reports | Add rows only for new or renamed exact test identities; regenerate through Make | medium | Phase/family maps are evidence accounting, not runtime ownership. |

### 4.1 Record/View Resolution Contract

The current base profile contains exactly the following 17 Collaboration-visible
record/view routes. A later implementation MUST preserve every identifier and
MUST validate this complete set against `projections.Catalog.Descriptors()` and
the generated `viewschema` registry. The existing
`contracts/projection-providers/index.json`,
`contracts/view-schemas/index.json`, and generated view-schema registry are
sufficient verification projections; this refactor MUST NOT introduce a new
record/view verification manifest.

| Contribution ID | Source owner | Record type | Variant | Canonical view schema |
| --- | --- | --- | --- | --- |
| `artifacts.artifact.comm_log.v1` | `artifacts` | `artifact` | `artifact_type=comm_log` | `cartulary.view.comm_log.v1` |
| `artifacts.artifact.finding.v1` | `artifacts` | `artifact` | `artifact_type=finding` | `cartulary.view.findings.v1` |
| `artifacts.artifact.handoff.v1` | `artifacts` | `artifact` | `artifact_type=handoff` | `cartulary.view.handoff.v1` |
| `artifacts.artifact.forensic_keyword.v1` | `artifacts` | `artifact` | `artifact_type=forensic_keyword` | `cartulary.view.forensic_keywords.v1` |
| `artifacts.artifact.lesson.v1` | `artifacts` | `artifact` | `artifact_type=lesson` | `cartulary.view.lesson.v1` |
| `artifacts.artifact.note.v1` | `artifacts` | `artifact` | `artifact_type=note` | `cartulary.view.notes.v1` |
| `artifacts.artifact.investigative_query.v1` | `artifacts` | `artifact` | `artifact_type=investigative_query` | `cartulary.view.investigative_queries.v1` |
| `artifacts.artifact.status_review.v1` | `artifacts` | `artifact` | `artifact_type=status_review` | `cartulary.view.status_review.v1` |
| `assessments.assessment.v1` | `assessments` | `assessment` | none | `cartulary.view.assessments.v1` |
| `tasksdecisions.decision.v1` | `tasksdecisions` | `decision` | none | `cartulary.view.decisions.v1` |
| `evidence.evidence.v1` | `evidence` | `evidence` | none | `cartulary.view.evidence.v1` |
| `entities.host.v1` | `entities` | `host` | none | `cartulary.view.hosts.v1` |
| `entities.identity.v1` | `entities` | `identity` | none | `cartulary.view.identities.v1` |
| `indicators.indicator.v1` | `indicators` | `indicator` | none | `cartulary.view.indicators.v1` |
| `parties.party.v1` | `parties` | `party` | none | `cartulary.view.parties.v1` |
| `tasksdecisions.task_request.v1` | `tasksdecisions` | `task_request` | none | `cartulary.view.task_requests.v1` |
| `timeline.timeline_event.v1` | `timeline` | `timeline_event` | none | `cartulary.view.timeline.v2` |

The existing `revisions.ProviderContribution` MUST gain one
`RecordViewRoutes []RecordViewRouteContribution` member. The descriptor
vocabulary is closed:

| Type or member | Required shape and default |
| --- | --- |
| `RecordVariant` | `{Kind, Value string}`; omission means no variant; the only admitted non-empty current-profile `Kind` is `artifact_type` |
| `RevisionSourceMatcher` | `{Kind, Value string}`; `Kind` is exactly `exact` or `prefix`; `prefix` values are non-empty and delimiter-bounded |
| `RecordViewRouteContribution` | Stable `ContributionID`, exact `SourceOwnerModule`, exact `RecordType`, optional `Variant`, non-empty canonical `ViewSchemaIDs`, and compatibility-only `RevisionSourceMatchers` |
| Missing optional slices | Normalize to immutable empty slices, not mutable shared `nil` state |
| Ordering | Routes sort by record type, variant kind, variant value, then contribution ID; view IDs sort bytewise; matchers sort by kind then value |

Constructors MUST defensively copy all maps and slices. Callers MUST receive no
mutable reference to catalog state. Regex, callbacks, globs, reflection,
runtime plugins, package initialization registration, package-global mutation,
and arbitrary metadata are forbidden.

Artifact resolution MUST preserve the current algorithm:

1. select the after-row when it exists; otherwise select the before-row;
2. if the selected row has cells, sort cell keys bytewise and use the prefix
   before the first `.` in the first key;
3. if no cell prefix exists, use exact `source.artifact_type`;
4. resolve exactly one route or return a typed internal invariant error.

No fallback guess is permitted after successful startup validation. A known
revision-backed type that intentionally emits no live event MUST declare
nonparticipation explicitly; absence MUST NOT imply nonparticipation.

### 4.2 Revisions Runtime and Single-Writer Contract

`internal/app/revisionassembly` MUST build one composition-scoped runtime
containing one immutable `RecordViewCatalog`, one concurrency-safe
`*revisions.Appender`, and the Revisions command service. The appender MUST be
constructed with the catalog and a Revisions-consumer-owned
`RecordChangeIntentAppender` interface satisfied by Collaboration. Every
source-owner service in that application or test composition MUST receive the
same appender pointer. The runtime is not a process-global singleton; distinct
server, operator, or isolated test compositions MAY build distinct runtimes.

The intended internal construction surface is fixed as:

```go
// package revisions
type RecordChangeIntentAppender interface {
	AppendIntentTx(context.Context, pgx.Tx, collaboration.EventIntent) error
}

func NewAppender(
	catalog *RecordViewCatalog,
	intentAppender RecordChangeIntentAppender,
) (*Appender, error)

// package revisionassembly
type Runtime struct {
	Catalog        *revisions.RecordViewCatalog
	Appender       *revisions.Appender
	CommandService *revisions.CommandService
}

type Dependencies struct {
	Database                    postgres.DB
	ImportedAttributionResolver revisions.ImportedAttributionResolver
	ProjectionServices          revisions.ProjectionServices
	ProjectionDescriptors       []projections.ProviderDescriptor
	CollaborationIntentAppender revisions.RecordChangeIntentAppender
}

func Build(
	dependencies Dependencies,
	contributions ...revisions.ProviderContribution,
) (*Runtime, error)
```

These are internal package interfaces and do not alter a public HTTP, WebSocket,
database, or generated contract. `Build` MUST copy contributions and projection
descriptors before validation. Production composition MUST supply the complete
source-owner contribution list. A source owner MAY define a local one-method
test interface around `*revisions.Appender`; Revisions MUST NOT export a broader
speculative interface.

`Appender` MAY retain only immutable catalog and intent-sink references. It
MUST NOT retain transaction, request, mutation, retry, or caller-specific
state, and its methods MUST be safe for concurrent use after construction.

The appender constructor MUST fail before route or worker readiness for an
empty or duplicate contribution ID, owner mismatch, duplicate or missing
selector, unexpected current-profile selector, unknown or duplicate view ID,
unsupported variant kind, empty or overlapping matcher, or undeclared
nonparticipation. It MUST NOT construct a partially usable runtime.

The only admitted live row-change path is:

```text
source-owner transaction
  -> source mutation
  -> revision/change-set mutation
  -> shared Revisions appender
  -> one deterministic Collaboration EventIntent
  -> Collaboration-owned intent append in the same transaction
  -> source transaction commit
  -> Collaboration sequencing, replay storage, and live dispatch
```

Source owners MUST NOT insert an additional `record_changed` intent, Revisions
MUST NOT own another durable stream, and the old and new catalogs MUST NOT be
production-reachable together. The old catalog MAY be compiled into a test-only
parity oracle before cutover. The production cutover MUST switch every
constructor path and remove `producer_catalog.go` from runtime use in one
completed slice.

### 4.3 Collaboration Raw-Socket Interface

Collaboration MUST own this exact transport-consumer surface:

```go
type MessageKind uint8

const (
	MessageText MessageKind = iota + 1
	MessageBinary
)

type Socket interface {
	Read(context.Context) (MessageKind, []byte, error)
	Write(context.Context, MessageKind, []byte) error
	Close(code uint16, reason string) error
}

type AcceptSocket func(http.ResponseWriter, *http.Request) (Socket, error)
type CheckBrowserOrigin func(http.ResponseWriter, *http.Request) bool
```

`CheckBrowserOrigin` MUST remain separate from `AcceptSocket` so the existing
cookie-browser Origin rejection can occur before upgrade and retain its current
HTTP `403` precedence. The server adapter MUST supply both functions and MUST
map only generic message kinds and close codes. It MUST NOT encode semantic
messages or maintain Collaboration state.

Collaboration route composition MUST use:

```go
type Settings struct {
	CheckBrowserOrigin CheckBrowserOrigin
	AcceptSocket       AcceptSocket
	Hub                *Hub
	ServiceVersion     string
}
```

`Settings` MUST fail route registration when either function or `Hub` is nil.
The server runtime field MUST be named
`CollaborationHub *collaboration.Hub`. `httpapi.DependencySet` MUST NOT retain
a WebSocket Hub field after the cutover.

The later split MUST move semantic messages, validators, patch builders, and
the codec into `internal/modules/collaboration/wire`; it MUST move Hub and
presence state into Collaboration; and it MUST leave only generic WebSocket
mechanics in platform. Collaboration MUST NOT import
`github.com/coder/websocket`, and `internal/platform/ws` MUST NOT import a
module package.

Server composition MUST expose a Collaboration-owned Hub, remove
`httpapi.DependencySet.WSHub`, inject the Hub through Auth's existing
consumer-owned revocation port and the Incidents notifier, remove unused Hub
fields from Imports, Job API, and Reference Data, and migrate test access from
`Runtime.WSHub` to the Collaboration-owned runtime surface.

### 4.4 Adopted WebSocket Defaults and Failure Mapping

The execution task intentionally removes incidental compatibility inherited
from `wsjson`: application messages are text-only and encoder-added trailing
line feeds are not part of the protocol. If an officially supported
binary-frame client outside this repository is discovered, SL-04 MUST stop as
`BLOCKED` and the owner MUST define a `/ws/v2/` migration rather than weakening
the v1 rule below.

| Behavior | Required value |
| --- | --- |
| Public route | Exactly `GET /ws/v1/incidents/{incident_id}` |
| Inbound application frames | Exactly one UTF-8 JSON object in one reassembled text message; binary application messages are rejected |
| Outbound application frames | Exactly one text message from the sole Collaboration codec; no encoder-added trailing LF |
| Message read limit | Exactly 32,768 bytes per reassembled application message |
| JSON objects | Duplicate members are rejected; unknown additive members remain accepted |
| Compression | Disabled |
| Subprotocol | None |
| First application message timeout | 10 seconds |
| Per-write timeout | 2 seconds |
| Incident subscriber buffer | 32 messages |
| Application ping interval | 15 seconds while otherwise idle |
| Inbound-dead timeout | 45 seconds without any inbound frame |
| Presence TTL | 45 seconds |
| Resume window | 5 minutes, without weakening Core replay-retention minima |
| Origin policy | With a session cookie and non-empty `Origin`, the header must equal `application.public_origin` exactly or the HTTP request fails `403` before authentication or upgrade. With an empty configured origin, any non-empty cookie-browser `Origin` fails that precheck. An absent `Origin` is not rejected by the precheck. Platform accept uses the configured origin's exact scheme and host as its allowed pattern; when configuration is empty, the pinned library's same-origin default remains active. |
| Authorization | Current session and incident membership are re-derived at connect/resume and before delivery; passive traffic does not extend idle expiry |

| Condition | Required close behavior |
| --- | --- |
| Binary application message | `1003`, reason `binary_message_unsupported` |
| Malformed JSON or invalid UTF-8 | `1007`, reason `invalid_json` |
| Message larger than 32,768 bytes | `1009`, reason `message_too_large` |
| Valid JSON but invalid first semantic message | Send `invalid_websocket_handshake`, then `1008`, reason `invalid_first_message` |
| Invalid later type, payload, duplicate member, or repeated establishment | Send `invalid_websocket_message`, then `1008`, reason `invalid_message` |
| Application heartbeat timeout | `1008`, reason `heartbeat_timeout` |
| Session or membership revocation | Send `session_revoked`, then `1008`, reason `session_revoked` |
| Incident closure | Send terminal `error`, then `1008`, reason `incident_closed`, before any writable acknowledgement |
| Subscriber queue overflow | Send current reset-required `resume_ack`, then `1013`, reason `slow_consumer` |
| Normal test/client shutdown | `1000`; this does not create a server terminal behavior |
| Unexpected transport failure | Preserve current abnormal-close behavior; do not invent a new semantic status |

### 4.5 Traceable Refactor Acceptance Requirements

| Requirement ID | Normative acceptance condition | Required evidence |
| --- | --- | --- |
| `CAT-001` | The assembled catalog contains exactly the 17 current routes and rejects a missing, duplicate, unexpected, unknown, or ambiguous route before readiness. | Catalog validation unit tests and projection/view-schema parity test |
| `CAT-002` | Every production source-owner service receives the same composition-scoped `*revisions.Appender`; no production source package calls `revisions.NewAppender`. | Constructor inventory and import/construction guard |
| `CAT-003` | A successful current mutation creates exactly one byte-equivalent deterministic intent; rollback creates none; historical restore creates none. | Service-backed mutation, rollback, and restore tests |
| `CAT-004` | The production cutover leaves no runtime reference to `producer_catalog.go`, no fallback catalog, and no new direct SQL writer to Collaboration intent tables. | Static search, boundary check, and source-backed tests |
| `WSB-001` | Platform owns only generic WebSocket mechanics; Collaboration owns the only semantic codec, DTO family, Hub, and presence state. | Import-boundary tests and exact symbol/import inventory |
| `WSB-002` | Live and replay delivery produce byte-equivalent canonical envelopes through one codec as one text message without an encoder-added trailing LF. | Golden live/replay wire fixtures |
| `WSB-003` | Origin, authorization, hello/resume/reset, heartbeat, presence, revocation, incident-close, slow-consumer, and failure mappings equal the frozen behavior above. | Collaboration, platform, Auth, Incidents, frontend, and browser tests |
| `WSB-004` | Authored WS contracts and generated TypeScript projections have no semantic diff. | `make generate-drift` and generated policy checks |
| `INT-001` | Jobs and Network Flow construct source-owned payloads and keys but append only through Collaboration-owned validation and persistence; source transactions remain atomic. | Producer success, rollback, replay, collision, and static SQL-boundary tests |
| `INT-002` | All three replayable event families are semantically validated before insertion and again before sequencing; existing invalid durable rows retain quarantine behavior. | Family validation and dispatcher quarantine tests |
| `INT-003` | Historical suppression is available only through `HistoricalIntentPolicy`, remains transaction-local, and cannot suppress a later transaction. | Incident import, ordinary restore/rollback, rollback, and isolation tests |
| `CAP-001` | Append, replay/token, dispatcher persistence, and recovery capabilities are separately constructed and injected; `Store` and `NewStore` are absent. | Constructor inventory, boundary checks, stream/replay/operator tests |
| `TEL-001` | Semantic telemetry is Collaboration-owned and event-type classification derives from the wire registry, admitting `resume_ack` and `extension_resource_changed` and rejecting `resume_result`. | Exact telemetry/privacy tests and `make otel-conformance` |
| `VER-001` | Every changed behavior test has one authored owner/family route with the correct runtime and fixture profile; generated topology is current. | Owner explanations, generators, drift and shape checks |
| `DOC-001` | The development guide describes platform as transport-only after WSB completion and does not contradict Core or the implemented boundary. | Documentation review in the authorized WSB slice |

### 4.6 SL-00 Characterization Baseline

This matrix records executable baseline evidence without elevating current
implementation deviations into requirements. In particular, the existing
binary-frame acceptance and encoder-added trailing LF are observed defects,
not preservation requirements.

| Requirement | Baseline evidence | Execution-time deviation or preservation result |
| --- | --- | --- |
| `CAT-001` | `TestRevisionsRuntimeBuildsExactImmutableRecordViewCatalog` and `TestRevisionsRuntimeRejectsIncompleteOrAmbiguousRecordViewCatalogs` | All 17 source-owned routes, canonical ordering, defensive copying, sorted artifact-prefix/deleted-row selection, and missing/duplicate/unexpected/unknown/unsupported/ambiguous rejection pass before readiness. |
| `CAT-002` | `revisionassembly.Runtime`, constructor inventory, exact runtime tests, and boundary guards | Assembly is the only production `revisions.NewAppender` caller; one runtime-owned pointer is reused by command and source services. |
| `CAT-003` | Collaboration/Revisions/source-owner service-backed slices, Incident Bundles import rows, and delete/restore/rollback live-event evidence | Commit/rollback/restore behavior passes, each committed participating revision has one deterministic intent, and historical imports have none. |
| `CAT-004` | Static inventory, retired-token/construction guards, SQL-ownership guards, and owner slices | `producer_catalog.go` is deleted, no runtime fallback exists, and Jobs/Network Flow have no Collaboration-table SQL. |
| `INT-001` | Jobs v2/coexistence and rollback tests; Network Flow replay/privacy and rollback tests; Collaboration collision evidence | Jobs and Network Flow own payload/key construction through consumer-owned ports; one shared stateless Collaboration appender owns canonical validation, exact replay, collision diagnostics, and insertion. Jobs v1 rows coexist unchanged with Go-derived v2 keys. |
| `INT-002` | `TestEventIntentValidatesEveryEventFamily_Unit` and durable legacy-family sequencing/quarantine evidence | Record, Jobs, and extension payloads are validated before insert and again before sequence; legacy invalid rows retain the dispatcher retry/quarantine policy. |
| `INT-003` | `TestHistoricalIntentSuppressionIsTransactionLocal_Integration`, Incident Bundles import evidence, and ordinary delete/restore/rollback event tests | The setting string is private to Collaboration, suppression is transaction-local, historical import emits no live record event, and ordinary restore/rollback still emits. |
| `WSB-001` | Collaboration semantic Hub/payload/codec rows and generic platform frame/origin row | SL-04 moved the semantic DTOs, codec, Hub, presence, and telemetry into Collaboration. Platform retains only generic transport mechanics, and the server owns the sole adapter. |
| `WSB-002` | Collaboration codec, replay/live, and strict socket failure tests | Live/replay semantics remain canonical; application input is text-only and valid output is one LF-free text message from the sole codec. |
| `WSB-003` | Collaboration service-backed lifecycle rows, generic platform origin/frame coverage, server, frontend, and browser suites | Origin, replay, presence, heartbeat, revocation, incident closure, slow-consumer behavior, and every adopted frame/failure mapping pass after the cutover. |
| `WSB-004` | `make generate-drift`, `make json-shape-check`, and frontend import-boundary check | All pass; no WS schema or generated protocol semantic change occurred, and repository browser clients send JSON text rather than binary frames. |
| `TEL-001` | Collaboration Hub telemetry rows, privacy tests, and OTel corpus/conformance | Telemetry is Collaboration-owned and derives safe labels from the wire registry; `resume_ack` and `extension_resource_changed` are exact while `resume_result` classifies as `other`. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Collaboration is a legitimate durable sequencer and live-session boundary, but the package is not a thin facade. | Core 03, `stream.go`, `routes.go`, migrations | Incorrectly dissolving the module would scatter sequencing and replay invariants. | `intentional/no_action` | Collaboration | Keep the module; refactor internal seams rather than eliminate the boundary. |
| Source-owner record/view mappings are centralized in `producer_catalog.go`. | Hard-coded catalog versus source-owner revision contributions, projection descriptors, and generated view-schema registry | New record owners or view IDs can drift across catalogs. | `must_fix` | Source owners, `revisionassembly`, Revisions validation | Implement RB-002 exactly: extend `ProviderContribution`, validate one immutable 17-route catalog, inject one shared appender, and atomically remove the Collaboration-owned duplicate after parity tests. |
| Jobs and Network Flow write Collaboration-owned intent tables directly. | Exact SQL in both producer files; schema ownership manifest | Collision, validation, and storage behavior can drift across producers. | `should_fix` | Source owners for semantics; Collaboration for persistence | Introduce narrow transaction-aware append ports without moving source mutation ownership. |
| Semantic messages, patch construction, presence, and Hub behavior live under platform WebSocket transport. | `internal/platform/ws/ws.go`, Collaboration routes, and presence tests | Platform becomes an owner of Collaboration behavior and complicates dependency direction. | `should_fix` | Collaboration semantic/wire packages; platform transport primitives; server adapter | Implement RB-003 after SL-00 through the exact raw-socket port and owner graph in Sections 3 and 4; one codec and one Hub are mandatory. |
| `Store` combines appender, query/replay, token, dispatcher, and recovery capabilities. | Exported surface and diverse callers | Callers can acquire broader storage capability than required. | `should_fix` | Collaboration | Introduce narrow internal capabilities and preserve compatibility during migration. |
| Revisions imports Collaboration and automatically emits intents while mapping source-owner view identities. | `internal/modules/revisions/appender.go`; production `NewAppender` construction inventory | Revisions can become a second owner for source mappings and Collaboration semantics. | `must_fix` | Revisions coordinates validated source-owner contributions; Collaboration builds and persists the intent | Use the resolved RB-002 composition-scoped appender and Revisions-owned sink interface. Parallel writers, fallback lookup, and per-module constructor calls are forbidden. |
| Historical suppression is a string-backed transaction setting with indirect coverage. | `historical_restore.go`, Revisions read, Incident Bundles call | Movement could publish historical rows as live edits. | `should_fix` | Collaboration integration port | Add direct characterization before changing ownership or API. |
| Collaboration owner-local test support is imported broadly. | Import search and `tools/test_support_inventory.json` | Moving it casually would erase semantic ownership or duplicate helpers. | `intentional/no_action` | Collaboration test support | Retain; move only a helper proven owner-neutral. |
| Producer-catalog and telemetry tests are not named by Collaboration owner-family rows. | Exact-symbol search in `tools/test_families/module.collaboration.json` | Focused owner slices can omit relevant unit evidence. | `should_fix` | Collaboration verification owner | Add or consolidate authored rows when later tests are changed, then regenerate. |
| Frontend Collaboration and Workbook controllers consume the protocol without backend-to-grid imports. | Frontend source/import inspection | Moving pending state server-side or coupling to vendor coordinates would change behavior. | `intentional/no_action` | Frontend Collaboration, Workbook, and grid adapter owners | Preserve current consumer boundary and selectors. |
| Generated protocol roots are downstream of authored contracts. | Generated-artifact policy, WS entrypoint owner, generator | Hand edits would drift or be overwritten. | `intentional/no_action` | Contract owners and generator | Change authored inputs only if a separately authorized public contract change is required. |
| Framework catalog understates the live target responsibilities. | Framework Collaboration catalog versus the 19-file inventory | A plan based only on the framework would miss storage, restore, operator, telemetry, and test-support risks. | `must_fix` | This tracker as planning evidence | Use the live inventory and record the mismatch; do not alter the framework in this task. |

No direct grid-vendor coupling, production test-only import, or generated-file
hand edit was found in the target. No owner contradiction was found.

### 5.1 Authorized Remediation Gap Register

This register is the execution-time disposition of every identified gap. The
specification column names the adopted owner when an observable rule changes;
the tracker remains subordinate execution evidence.

| Gap | Remediation and change areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation and completion |
| --- | --- | --- | --- | --- | --- |
| `G-01` | **Specification, tests, tracker:** Core 01 defines text-only UTF-8 JSON messages, the `32768`-byte reassembled limit, duplicate-member rejection, no encoder-added trailing LF, and stable error/close mappings; Core 03 and Core 04 define exactly one current live revision intent and no historical-import intent. | Owner-defined limits and failures survive library changes; incidental codec behavior stops constraining future transport work. | Repository clients already send text. Binary clients must migrate to text; any officially supported external binary client blocks SL-04 pending `/ws/v2/`. JSON shapes and routes do not change. | Invalid traffic remains ambiguous, library defaults become accidental contracts, and historical import can look like a current edit. | Owner prose and tracker agree; Markdown lint, frame/failure tests, intent-count tests, and generated drift pass. |
| `G-02` | **Implementation, tests:** source-owner `ProviderContribution` values carry closed record/view routes; Revisions compiles an immutable catalog; `revisionassembly.Build` creates one catalog and appender before source facades, and `Runtime.NewCommandService` later binds live projections. Delete the Collaboration catalog and all production `NewAppender` construction outside assembly. | One authoritative contribution per source restores cohesion and lets new phases add an owner without editing Collaboration. | Internal constructors change; public identifiers, payloads, transactions, and storage do not. | Catalog drift causes missing live refresh, and per-module appenders permit inconsistent routing or duplicate intent paths. | Exact 17-route parity, defensive-copy/order, missing/duplicate/unexpected/unknown/ambiguous rejection, shared-pointer inventory, mutation/rollback/restore tests, and construction guards pass. |
| `G-03` | **Implementation, tests:** Jobs and Network Flow expose consumer-owned producer ports; server/test adapters call a stateless Collaboration `IntentAppender`. Jobs uses `job_progress:v2:<job_id>:<sha256(canonical-payload)>`; Network Flow keys remain unchanged. Validate every replayable family before insert and sequence. | Source transactions keep semantic authority without access to Collaboration tables; validation, replay comparison, and collision diagnostics become uniform. | No migration or rewrite. Existing Jobs v1 intents remain readable; new writes use v2. | Direct SQL writers can drift, malformed durable events can poison streams, and storage evolution remains coupled to producers. | No external Collaboration-table SQL; exact replay, collision, rollback, invalid-family, privacy, legacy-v1/new-v2 coexistence, and quarantine tests pass. |
| `G-04` | **Implementation, tests:** Collaboration owns `HistoricalIntentPolicy` with `SuppressTx` and `IsSuppressedTx`; Revisions and Incident Bundles depend on consumer-owned narrow ports; the setting string is private to Collaboration. | The exceptional restore rule gains one evolvable owner and direct evidence. | No database or bundle change; the same transaction-local setting remains. | Imported history can be emitted as current live activity or suppression can leak across transactions. | Historical import emits none; ordinary restore/rollback still emits as required; rollback and later unrelated transactions are unsuppressed. |
| `G-05` | **Implementation, tests:** remove broad `Store`; inject only `IntentAppender`, `ReplayStore`, dispatcher-private persistence, or `RecoveryService`; split implementation files only where it improves cohesion and add no speculative interface. | Least-capability construction reduces accidental authority and storage coupling. | Internal composition only; durable data and replay behavior remain stable; no compatibility facade survives the slice. | Source callers can acquire replay/recovery authority and unrelated storage changes remain coupled. | Route, dispatcher, operator, and producer constructors expose only their required capability; stream/replay/retention/requeue tests pass; `Store`/`NewStore` are absent. |
| `G-06` | **Implementation, tests, documentation:** move envelopes, payloads, validation, patch construction, codec, Hub, presence, semantic queues, and send telemetry to Collaboration; platform retains only upgrade/Origin/frame/control/close mechanics behind the raw-socket port; server supplies the adapter. | The semantic owner can grow without turning platform into a second domain layer, and authorization cannot be bypassed through a generic Hub. | JSON schemas and frontend semantics remain stable; only the G-01 frame/whitespace cleanup is deliberate. | Cross-layer edits, import cycles, duplicated codecs/Hubs, and silently accepted invalid traffic persist. | One codec/Hub; no third-party socket import in Collaboration; no semantic platform symbols or module imports; fail-closed dependency registration; origin/replay/presence/heartbeat/revocation/closure/slow-consumer tests pass. |
| `G-07` | **Implementation, tests, generated conformance inputs:** move semantic telemetry to Collaboration and derive safe types from its wire registry; replace `resume_result` with `resume_ack`; admit `extension_resource_changed`; update privacy code and OTel conformance sources/fixtures. | One vocabulary eliminates label drift while retaining privacy and bounded cardinality. | Operational label values are corrected; signal names do not change. | Extension traffic is mislabeled and runtime behavior contradicts OTel conformance evidence. | Every public server type maps exactly or to `other`, forbidden identifiers remain absent, no-SDK/exporter failure is inert, and OTel conformance passes. |
| `G-08` | **Tests, harness inputs, generated outputs:** each behavior slice adds or moves its exact authored owner/family rows and boundary policy before closure; SL-06 audits identity, profiles, fixtures, and generated topology through Make. | Focused owner slices remain reliable after package and symbol movement. | Harness accounting changes; product behavior does not. | Narrow checks can pass while critical tests never execute. | Every changed test routes exactly once with correct service profile; no stale selector; owner explanation, generation, drift, policy, shape, and boundary checks pass. |
| `G-09` | **Documentation, repository hygiene:** update the development guide to the transport/semantic boundary; remove `.gitkeep` only after layout stabilizes; keep `docs/domain.md` unchanged. | Contributors see the implemented owner graph and do not recreate coupling. | Documentation and an obsolete placeholder only. | Stale guidance drives new boundary violations and the package tree disagrees with the tracker. | Guide, tracker, imports, and tree agree; Markdown lint and boundary checks pass; no tooling references `.gitkeep`. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01, WF-02 | Establish scope, authority, repository identity, and sole-write constraint. | Tracker; owner documents; repository control files | Read-only repository inspection | Target exists, tracker initialized, no non-tracker modification. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Account for all 19 target files, exact package surfaces, callers, dependencies, and tests. | Collaboration target and callers | Inventory count and import/symbol searches | Every target file has one inventory row. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04, WF-05 | Bind each affected observable to Core/subsystem/source-owner authority. | Core/NLSpecs, WS contracts, migrations, telemetry and frontend consumers | Contract/source cross-check | Every contract risk has an owner and evidence posture. |
| WF-03 | Characterization test gap analysis | parallel | WF-02 | WF-06 | Identify missing mapping, producer, suppression, and accounting coverage before code movement. | Collaboration tests, Jobs, Network Flow, Revisions, test-family owners | `make task-guide` and owner explanations; later owner slices | Exact gaps and required authored rows are named. |
| WF-04 | Boundary and coupling scan | parallel | WF-01, WF-02 | WF-05 | Diagnose source-owner mapping, direct SQL, platform semantics, Store breadth, and test-support posture. | Collaboration, platform WebSocket, source owners, app assembly | Boundary manifests and import searches | Every finding has classification and proposed owner. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-04 | WF-06 | Define the behavior-preserving source contribution/runtime contract, persistence ports, narrow capabilities, and raw-socket semantic/transport split. | Collaboration, Revisions, revision assembly, source owners, platform WebSocket, server adapter | Requirements `CAT-001` through `CAT-004` and `WSB-001` through `WSB-004` | RB-002 and RB-003 have one decision-complete owner graph each; no dual writer, duplicate codec/Hub, cycle, schema change, or widened capability remains. |
| WF-06 | Slice sequencing plan | chain | WF-03, WF-05 | WF-07, WF-08 | Order the smallest reversible implementation slices. | Packages named by SL-00 through SL-07 | Slice-specific commands in Section 8 | Each slice has dependency, rollback, and binary exit condition. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-06 | WF-08 | Update authored evidence routing only when exact test identities change. | Verification owner/family inputs and generator-owned outputs | `make generate-drift`; owner slices | Authored inputs precede generated refresh; no hand edits. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run narrow-to-broad verification and record exact retained evidence and failures. | Changed implementation/test/owner inputs and tracker | Section 8 validation ladder; `make agent-finalize` before broad end-of-run checks | Clean generated drift, recorded commands/results, current next action. |

## 7. Authorized Remediation Slice Plan

The slices are authorized in the order below. Every row is a hard gate: mark
the matching Section 9 work item `IN_PROGRESS`, implement and validate, record
files/commands/results/run roots and set `DONE` or `BLOCKED`, run tracker
diff/Markdown checks, and only then begin the next row. No slice authorizes a
database migration, new public route, or JSON message-shape change.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WS-00 | None | Refresh HEAD/status, replace planning-only posture and RB-001, add G-01 through G-09, adopt the wire/resource rules in Core 01/Core 03/Core 04, revise dependencies, and record that `docs/domain.md` needs no edit. | Tracker and adopted Core owners | The tracker must not become a parallel behavioral owner. | Owner/tracker consistency and no generated WS semantic change. | `make lint-markdown`; `make generate-drift`; tracker diff checks | Revert the owner/tracker alignment together if an owner contradiction is found. | Owner precedence remains explicit; CI-000 is `DONE`; no owner contradiction or generated semantic drift. |
| SL-00 | WS-00 | Add characterization for the 17-route mapping, historical suppression, Jobs/Network Flow atomicity and collision, live/replay semantics, security/lifecycle defaults, and telemetry; record known deviations without freezing binary frames or trailing LF. | Collaboration, Revisions, Incident Bundles, Jobs, Network Flow, platform WS tests and authored family rows | Tests must derive from adopted requirements and distinguish current deviations from required behavior. | Direct preservation matrix for every `CAT-*`, `INT-*`, and `WSB-*` invariant. | Focused slices for `module.collaboration`, `platform.ws`, `module.revisions`, `module.networkflow`, `module.extensions`, and `module.incidentbundles` | Revert tests and authored rows together; retain a failing test only for a confirmed implementation gap. | Every invariant has direct executable evidence and the focused characterization baseline is recorded. |
| SL-04 | SL-00 | Cut over once to the raw-socket port, Collaboration wire/Hub/presence, strict text codec, strict later-message handling, and corrected semantic telemetry; update the development guide. No alias, duplicate Hub, or dual delivery path may remain. | Collaboration routes/wire/Hub; platform WS; HTTP API; server; Auth; Incidents; Imports; Job API; Reference Data; OTel inputs; guide | Highest risk is auth, Origin, replay, close, presence, and live/replay byte regression. | Frame/failure mapping, codec golden, lifecycle/security, telemetry/privacy, browser and boundary evidence. | Focused `platform.ws`, `module.collaboration`, `app.server`, `web.collaboration`; OTel, boundary, browser webserver/stateful, generated drift | Revert the entire semantic ownership seam; no bridge alias is allowed. | G-01, G-06, G-07 and all `WSB-*`/`TEL-*` criteria pass with one codec and one Hub. |
| SL-01 | SL-04 | Implement nested source-owned route contributions, immutable catalog, cycle-safe `revisionassembly.Runtime`, shared pointer injection, `HistoricalIntentPolicy`, and atomic producer-catalog removal. | Revisions, revision assembly, projections, view-schema registry, Collaboration, Incident Bundles, and all participating source owners | Missed constructor paths, duplicate intents, catalog ambiguity, or historical suppression regression. | Exact catalog validation, construction guard, mutation/rollback/restore/import and socket tests. | Collaboration/Revisions/app-server plus affected source-owner unit/service slices; boundary and drift checks | Revert the full catalog/appender/policy cutover together. | G-02/G-04 and `CAT-*`/`INT-003` pass; all 17 routes contribute once; one appender pointer is shared. |
| SL-02 | SL-01 | Introduce Jobs and Network Flow consumer-owned producer ports, transaction services, central insert/family validation, Jobs v2 keys, and remove direct intent-table SQL. | Collaboration; platform Jobs; Network Flow; server/test assembly; Extensions, Imports, Incident Bundles, Reference Data, Reporting, Job API | Source-transaction atomicity, idempotency, legacy coexistence, or extension privacy regression. | Success/replay/collision/rollback/invalid-family/privacy/v1-v2 coexistence/quarantine tests. | Affected focused and service-backed owner slices plus boundary checks | Revert adapters and callers together; durable v1 rows require no rewrite. | G-03 and `INT-001`/`INT-002` pass; only Collaboration accesses its intent tables. |
| SL-03 | SL-01, SL-02 | Replace broad Store construction with append, replay/token, dispatcher-private persistence, and recovery capabilities; remove temporary adapters in this slice. | Collaboration stream implementation; server/operator/timeline/source composition | Shutdown, retry, retention, replay, or operator requeue regression. | Stream, replay, retention, quarantine/requeue, server/operator and source-owner tests. | Collaboration, app-server, app-operator, affected owner slices, boundary checks | Revert caller conversions coherently; do not retain a compatibility facade. | G-05 and `CAP-001` pass; broad `Store`/`NewStore` are absent. |
| SL-05 | SL-03, SL-04 | Remove `.gitkeep`, stale aliases, obsolete test helpers, dead imports, and no-longer-used platform semantic files after layout stabilizes. | Collaboration/platform package tree and authored routing references | Deleting evidence still named by verification routing. | Static symbol/path/import scans. | `make backend-module-boundary-check` and applicable narrow checks | Restore only an item proven still authoritative. | G-09 passes; no obsolete reference or placeholder remains. |
| SL-06 | All behavior slices | Audit exact test identity, owner/collaborator, runtime/resource/fixture profile, OTel fixtures, boundary policies, and generated topology; regenerate only through public Make targets. | Authored verification/boundary/conformance inputs and Make-generated projections | Lost, duplicate, stale, or misprofiled evidence. | Owner explanation and accounting checks. | `make generate`; drift, generated policy, JSON shape, OTel, boundary checks | Revert authored input and its generated projection together. | G-08 and `VER-001` pass; every changed test routes exactly once. |
| SL-07 | SL-06 | Run final narrow-to-broad validation on final sources, inspect retained evidence, refresh every tracker/handoff section, and record compatibility and skipped checks. | Entire final diff and tracker | False completion from stale evidence or concealed failures. | All affected owner, service, browser, generation, static, fast, and broad suites. | `make agent-finalize`; affected slices; browser webserver/stateful; `make test-fast`; `make check` | Do not weaken requirements or generated evidence; a failing applicable criterion leaves SL-07 `BLOCKED`. | All `CI-*` rows are `DONE`, binary criteria pass, drift is clean, and exact commands/run roots are recorded. |

## 8. Validation Plan

Commands were discovered from the current public Make surface using
`make help`, `make help-all`, `make task-guide`,
`make explain-test-owner`, and `make explain-target`. Row and phase maps are
verification accounting only.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.collaboration` | All 33 active Collaboration owner rows when `ROWS` is omitted, including unit/frontend and service-backed rows selected by the owner catalog | yes | Use exact `ROWS` only for a deliberately bounded slice; do not infer tests from filenames. |
| integration | `make service-backed-test-slice OWNER=module.collaboration` | The 21 currently service-backed Collaboration rows | yes | Covers real Postgres/server/browser-backed profiles owned by the catalog. |
| frontend unit | `make test-slice OWNER=web.collaboration` | Frontend Collaboration session regression | no | Required when frontend imports or the session consumer change; otherwise use as final contract-consumer confidence where risk warrants. |
| e2e/browser | `make browser-e2e-webserver-backed` and `make browser-e2e-stateful` | Cross-stack Collaboration, revocation, multi-client, pending-edit, and live-update behavior | no | Required before final handoff for route/protocol/presence changes; add a11y/visual only when visible behavior or fixtures change. |
| generated drift | `make generate-drift` | Generated protocol, harness topology, schedules, and other policy-owned outputs | yes | Mandatory after authored contract or harness owner input changes; generated files are never hand-edited. |
| generated policy/shape | `make generated-artifact-policy-check` and `make json-shape-check` | Generated-root policy and machine projection shapes | no | Required with SL-06 or any authored machine-contract edit. |
| import-boundary/static | `make backend-module-boundary-check` | Backend module imports, source ownership, and SQL/access constraints | yes | Run after each ownership/capability slice. |
| frontend import boundary | `make frontend-import-boundary-check` | Frontend Collaboration/Workbook/grid import boundaries | no | Required only if frontend files or imports move. |
| focused aggregate | `make test-fast` | Common backend/frontend and service-backed verification | no | Run after focused owner slices pass. |
| full check | `make check` | Current default local repository verification | no | Final broad validation. An initial `make agent-finalize` was run before broad verification; retained-run completion was then rerun against the successful full warm-check root required by the finalizer. |

Every slice has run its required owner, service-backed, formatting, generation,
drift, shape, boundary, and Markdown validation. SL-07 reran every affected
owner, the complete service-backed matrix, the explicit browser suites,
`test-fast`, `check`, and `agent-finalize` against the final source state.
Exact run roots and the resolved final-validation findings are retained in
Section 10.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| CT-001 | Read framework first and establish authority/source posture | WF-00 | DONE | none | Framework and owner documents listed in Section 1 | Authority order and non-goals are explicit. |
| CT-002 | Inventory all 19 target files and live package surfaces | WF-01 | DONE | CT-001 | Section 2 | Every file has an evidence-backed row. |
| CT-003 | Map observable contracts and current tests | WF-02 | DONE | CT-002 | Section 4 | Every discovered contract risk has owner and test posture. |
| CT-004 | Diagnose boundary and coupling findings | WF-04 | DONE | CT-002, CT-003 | Sections 3 and 5 | Findings use only allowed classifications and identify proposed owners. |
| CT-005 | Discover canonical validation commands and owner routing | WF-03, WF-07 | DONE | CT-003 | Section 8; Make discovery output | Commands are current Make-owned targets or explicitly conditional. |
| CT-006 | Create target-specific planning tracker and current handoff | WF-00 through WF-08 | DONE | CT-001 through CT-005 | This file | Required 12 sections exist and this is the sole changed file. |
| CT-007 | Resolve catalog injection and WebSocket boundary decisions in NLSpec voice | WF-05, WF-06 | DONE | CT-003, CT-004 | Sections 3, 4, 7, 11, and 12 | RB-002 and RB-003 have exact interfaces, defaults, owner graphs, failure rules, and binary acceptance criteria. |
| CI-000 | Align the tracker and adopted Core owners with the authorized remediation | WF-02, WF-06 | DONE | CT-007 | WS-00; Core 01, Core 03, and Core 04 | Baseline `c26f8ff1e1344a8cf8c14d143a74a7f82595be35`; G-01 through G-09 and the revised order are recorded; Markdown and generated drift pass. |
| CI-001 | Add mapping, suppression, producer, WebSocket, and telemetry characterization/accounting | WF-03, WF-07 | DONE | CI-000 | SL-00 | Every `CAT-*`, `INT-*`, `WSB-*`, and `TEL-*` invariant has direct evidence and known deviations are explicit. |
| CI-005 | Separate semantic Collaboration protocol from platform WebSocket mechanics | WF-05, WF-06 | DONE | CI-001 | SL-04 | `WSB-*`, `TEL-001`, and `DOC-001` pass with one codec, one Hub, strict traffic handling, and the adopted frame behavior. |
| CI-002 | Move record/view resolution to source-owner contributions and encapsulate historical suppression | WF-05, WF-06 | DONE | CI-005 | SL-01 | `CAT-*` and `INT-003` pass; the complete catalog and shared appender replace the duplicate and the policy string has one owner. |
| CI-003 | Centralize Jobs and Network Flow intent persistence | WF-05, WF-06 | DONE | CI-002 | SL-02 | `INT-001` and `INT-002` pass; no direct source-owner SQL reaches Collaboration tables; atomicity and legacy coexistence hold. |
| CI-004 | Narrow Collaboration Store capabilities | WF-05, WF-06 | DONE | CI-002, CI-003 | SL-03 | `CAP-001` passes; callers depend on minimal ports and `Store`/`NewStore` are absent. |
| CI-006 | Remove obsolete structure and helpers | WF-06 | DONE | CI-004, CI-005 | SL-05 | `.gitkeep`, stale aliases/helpers/imports, and obsolete references are absent. |
| CI-007 | Audit authored verification accounting and generated projections | WF-07 | DONE | CI-001 through CI-006 | SL-06 | `VER-001` passes; exact rows route once and generated drift is clean. |
| CI-008 | Run final validation and implementation handoff | WF-08 | DONE | CI-007 | SL-07; Section 10 final evidence | All applicable requirements and commands pass with exact run roots recorded. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Scope and owner posture complete; implementation remains unauthorized | Inspected `AGENTS.md`, framework, Core 00-05 applicability, domain and adopted subsystem NLSpecs; touched only this tracker | `sed`, `rg`, `git status --short --branch`, `git rev-parse HEAD`, `date --iso-8601=seconds` | Target and authority confirmed; no owner contradiction | RB-001 | Seek later authorization for a selected implementation slice. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Normative posture is explicit and subordinate to adopted owners; RB-002 and RB-003 are resolved design decisions | Inspected `docs/research/nlspec-spec.md`, `temp/analysis-notes.md`, owner passages, framework, development guide, and this tracker; touched only this tracker | `sed`, `rg`, `jq`, `git diff --check -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git diff --check HEAD -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git status --short --branch`, `date --iso-8601=seconds` | Supporting material was incorporated without elevating it to behavioral authority; both diff checks passed; no owner contradiction was found | RB-001 only | Obtain later authorization for SL-00 before any implementation. |
| 2026-07-27T22:57:43-04:00 | Codex WS-00 execution | Owner and tracker alignment complete | Changed this tracker plus Core 01, Core 03, and Core 04; confirmed `docs/domain.md` needs no change | `git diff --check`; `make lint-markdown`; `make generate-drift` | All pass; lint run `.cartulary/test-results/20260728T025713Z-p2892113`; drift run `.cartulary/test-results/20260728T025725Z-p2893569`; no generated semantic diff | No open blocker; external binary-client discovery remains an SL-04 stop condition | Mark SL-00 active and add direct characterization without freezing obsolete binary/LF behavior. |
| 2026-07-27T23:17:56-04:00 | Codex SL-00 execution | Characterization and exact authored routing complete; SL-04 activated only after the gate | Added exact Collaboration catalog and historical-policy tests, Jobs intent persistence characterization, standalone claimed Network Flow composition, and authored Collaboration/platform WS/Extensions/Network Flow rows; regenerated topology and schedules through Make | `make format`; focused and service-backed slices for Collaboration, platform WS, Revisions, Network Flow, Extensions, and Incident Bundles; `make generate`; `make generate-drift`; `make json-shape-check`; tracker diff checks; `make lint-markdown` | All final runs pass; representative roots: Collaboration unit `.cartulary/test-results/20260728T030610Z-p2932709`, Collaboration service `.cartulary/test-results/20260728T030826Z-p2960412`, Revisions unit `.cartulary/test-results/20260728T031039Z-p2987116`, Revisions service `.cartulary/test-results/20260728T031132Z-p3007742`, Network Flow `.cartulary/test-results/20260728T031703Z-p3038594`, Extensions `.cartulary/test-results/20260728T031717Z-p3039964`, Incident Bundles `.cartulary/test-results/20260728T031724Z-p3041285`, drift `.cartulary/test-results/20260728T031733Z-p3042544`, shape `.cartulary/test-results/20260728T031744Z-p3046336`, Markdown `.cartulary/test-results/20260728T031857Z-p3047273` | No open blocker; one initial Network Flow 404 exposed and corrected an unclaimed standalone test composition | Perform the single SL-04 semantic WebSocket cutover. |
| 2026-07-27T23:53:27-04:00 | Codex SL-04 execution | Semantic WebSocket ownership, strict wire behavior, telemetry vocabulary, and development guidance are complete; SL-01 activated after the gate | Added Collaboration codec/protocol/Hub/telemetry source and exact tests; reduced platform WS to generic transport; added the server adapter; narrowed HTTP/Auth/runtime dependencies; moved authored rows and OTel evidence; updated boundary policy and the development guide | `make format`; focused and service-backed Collaboration, platform WS, app-server, and web owner slices; `make otel-conformance`; backend/frontend boundary checks; browser webserver/stateful suites; `make generate`; `make generate-drift`; `make json-shape-check`; static import/client scans; `git diff --check`; `make lint-markdown` | All final runs pass. Roots: platform `.cartulary/test-results/20260728T033743Z-p3101135`, app server `.cartulary/test-results/20260728T033852Z-p3107772`, web `.cartulary/test-results/20260728T034003Z-p3137839`, Collaboration unit `.cartulary/test-results/20260728T034012Z-p3138207`, Collaboration service `.cartulary/test-results/20260728T034226Z-p3166075`, OTel `.cartulary/test-results/20260728T033845Z-p3104326`, backend boundary `.cartulary/test-results/20260728T033843Z-p3104071`, browser webserver `.cartulary/test-results/20260728T034439Z-p3192849`, browser stateful `.cartulary/test-results/20260728T034918Z-p3220756`, drift `.cartulary/test-results/20260728T035127Z-p3243146`, frontend boundary `.cartulary/test-results/20260728T035311Z-p3248551`, shape `.cartulary/test-results/20260728T035315Z-p3249019`, Markdown gate `.cartulary/test-results/20260728T035414Z-p3249909` | No blocker. Repository clients use text frames; no supported binary client was discovered. Residual stale test aliases and placeholders are intentionally deferred to SL-05 after the final package shape stabilizes. | Implement the source-owned catalog/runtime and historical-policy cutover in SL-01. |
| 2026-07-28T01:01:45-04:00 | Codex SL-01 execution | Source-owned record/view routing, the shared Revisions runtime, and Collaboration-owned historical policy are complete; SL-02 activated only after this tracker gate | Added nested route/policy contributions across the nine Revisions source owners; added immutable catalog/runtime source and exact failure tests; injected one appender through server, Timeline, Workbook, Imports, and focused test compositions; removed Collaboration producer catalog/tests; replaced raw suppression knowledge with `HistoricalIntentPolicy`; added historical-import and ordinary live-intent assertions; updated authored Revisions routing and boundary guards | `make format`; `make backend-unit`; `make backend-store`; exact and full owner slices for Collaboration, Revisions, app-server, Artifacts, Assessments, Entities, Evidence, Indicators, Parties, Tasks/Decisions, Timeline, Workbook, Imports, and Incident Bundles; corresponding service-backed slices where offered; `make generate`; `make generate-drift`; `make backend-module-boundary-check`; static searches; `git diff --check` | Product and exact accounting evidence pass. Key roots: catalog `.cartulary/test-results/20260728T044634Z-p3773392`, Collaboration `.cartulary/test-results/20260728T043038Z-p3514133`, app-server `.cartulary/test-results/20260728T043253Z-p3543082`, Artifacts `.cartulary/test-results/20260728T043503Z-p3584596`, Assessments `.cartulary/test-results/20260728T043511Z-p3585943`, Entities `.cartulary/test-results/20260728T043546Z-p3603033`, Evidence `.cartulary/test-results/20260728T043711Z-p3627356`, Indicators `.cartulary/test-results/20260728T043859Z-p3653926`, Parties `.cartulary/test-results/20260728T043913Z-p3655940`, Tasks/Decisions `.cartulary/test-results/20260728T043950Z-p3673032`, Timeline `.cartulary/test-results/20260728T044023Z-p3689834`, Workbook `.cartulary/test-results/20260728T044149Z-p3714268`, Incident Bundles `.cartulary/test-results/20260728T044545Z-p3769581`, Revisions service `.cartulary/test-results/20260728T042728Z-p3478473`, Collaboration service `.cartulary/test-results/20260728T044637Z-p3773759`, Workbook service `.cartulary/test-results/20260728T045645Z-p3950610`, Imports service `.cartulary/test-results/20260728T045954Z-p3984628`, Incident Bundles service `.cartulary/test-results/20260728T050039Z-p4003298`, backend unit `.cartulary/test-results/20260728T042956Z-p3503590`, backend store `.cartulary/test-results/20260728T043434Z-p3576100`, boundary `.cartulary/test-results/20260728T042725Z-p3478176`, drift `.cartulary/test-results/20260728T043014Z-p3510251` | No implementation blocker. Existing browser wrapper rows in broad Revisions, Imports, Entities-service, and Evidence-service runs reported `missing_selector_result` after their child target passed; no Go/product test failed, and the changed exact/service rows have clean pass roots. Preserve this harness-accounting observation for SL-06 rather than weakening a selector. | Implement Jobs and Network Flow producer ports and family validation in SL-02. |
| 2026-07-28T01:43:56-04:00 | Codex SL-02 execution | Jobs and Network Flow producer persistence is centralized; SL-03 activated only after this tracker gate | Added the stateless Collaboration `IntentAppender`; added server and separated test adapters for Jobs/Network Flow consumer-owned ports; introduced configured Jobs `TransactionService` and injected it into Imports, Incident Bundles, Reference Data, Reporting, composition preview, extension finalization, and the Jobs manager; replaced Jobs SQL hashing with canonical Go SHA-256 v2 keys; removed Jobs/Network Flow Collaboration-table SQL; added semantic family validation before insert and sequence; added rollback, legacy v1/v2 coexistence, privacy, and legacy-invalid sequencing evidence; updated exact routing and boundary guards | `make format`; `make backend-unit`; `make backend-store`; focused and service-backed slices for Collaboration, Extensions, Network Flow, Imports, Incident Bundles, Reference Data, Reporting, Job API, and app-server; exact producer/validation rows; `make generate`; `make generate-drift`; `make json-shape-check`; `make backend-module-boundary-check`; static SQL/construction searches; tracker diff checks; `make lint-markdown` | All applicable product and accounting evidence passes. Key roots: Collaboration durable `.cartulary/test-results/20260728T052121Z-p4104324`, family validation `.cartulary/test-results/20260728T054056Z-p173702`, Collaboration service `.cartulary/test-results/20260728T053607Z-p135339`, Extensions `.cartulary/test-results/20260728T052400Z-p4127054`, Extensions service `.cartulary/test-results/20260728T052911Z-p23103`, Jobs v2/rollback exact `.cartulary/test-results/20260728T054219Z-p177823`, Network Flow `.cartulary/test-results/20260728T051419Z-p4038575`, Network Flow service `.cartulary/test-results/20260728T053014Z-p40666`, Network Flow rollback exact `.cartulary/test-results/20260728T054042Z-p172336`, Imports `.cartulary/test-results/20260728T052624Z-p4172778`, Imports service `.cartulary/test-results/20260728T053228Z-p69013`, Incident Bundles `.cartulary/test-results/20260728T052712Z-p4191931`, Incident Bundles service `.cartulary/test-results/20260728T053315Z-p87876`, Reference Data `.cartulary/test-results/20260728T052751Z-p4194192`, Reference Data service `.cartulary/test-results/20260728T053349Z-p89508`, Reporting `.cartulary/test-results/20260728T052825Z-p17628`, Reporting service `.cartulary/test-results/20260728T053422Z-p106168`, Job API `.cartulary/test-results/20260728T052840Z-p19435`, Job API service `.cartulary/test-results/20260728T053437Z-p107488`, app-server `.cartulary/test-results/20260728T052503Z-p4145085`, app-server service `.cartulary/test-results/20260728T053452Z-p108813`, boundary `.cartulary/test-results/20260728T054206Z-p177465`, drift `.cartulary/test-results/20260728T054239Z-p179207`, shape `.cartulary/test-results/20260728T054259Z-p183031`, backend unit `.cartulary/test-results/20260728T054309Z-p183612`, Markdown gate `.cartulary/test-results/20260728T054556Z-p187417` | No implementation blocker. Existing durable v1 Jobs intents are not rewritten or re-enqueued; the new v2 key prefix applies only to new writes. | Split the broad Collaboration Store into replay/token, dispatcher persistence, and recovery capabilities in SL-03. |
| 2026-07-28T02:19:28-04:00 | Codex SL-03 execution | Least-capability Collaboration persistence is complete; SL-05 remains pending until this tracker gate passes | Replaced broad `Store` with `IntentAppender`, route-only `ReplayStore`, dispatcher-private repository, and operator-only `RecoveryService`; made the low-level append function private; injected one pointer-backed intent appender through Revisions, Timeline, Evidence, Workbook, Imports, Jobs, and Network Flow composition; removed every production fallback constructor; added `recovery.go`; updated durable-stream tests, authored collaborator routing, and construction/broad-facade guards | `make format`; `make backend-unit`; focused and service-backed Collaboration; focused app-server, app-operator, Revisions, Timeline, Evidence, Imports, and Workbook; app-server service-backed; `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make backend-module-boundary-check`; static constructor/symbol searches | All applicable final runs pass. Roots: backend unit `.cartulary/test-results/20260728T055731Z-p198245`, Collaboration focused `.cartulary/test-results/20260728T060006Z-p210111`, Collaboration service `.cartulary/test-results/20260728T060229Z-p239949`, app-server `.cartulary/test-results/20260728T060500Z-p267135`, app-server service `.cartulary/test-results/20260728T060635Z-p297428`, app-operator `.cartulary/test-results/20260728T060624Z-p295900`, Revisions `.cartulary/test-results/20260728T060744Z-p324148`, Timeline `.cartulary/test-results/20260728T060837Z-p344691`, Evidence final `.cartulary/test-results/20260728T061227Z-p395738`, Imports `.cartulary/test-results/20260728T061424Z-p420603`, Workbook `.cartulary/test-results/20260728T061507Z-p439692`, generation `.cartulary/test-results/20260728T061823Z-p476211`, drift `.cartulary/test-results/20260728T061836Z-p478462`, generated policy `.cartulary/test-results/20260728T061847Z-p482248`, shape `.cartulary/test-results/20260728T061849Z-p482594`, boundary `.cartulary/test-results/20260728T061852Z-p483171` | No implementation blocker. The first Evidence owner run `.cartulary/test-results/20260728T061006Z-p369929` had one isolated browser-backed upload timeout after every backend work unit and five sibling browser cases passed; the unchanged exact rerun passed all 35 tests. | Pass the tracker diff/Markdown gate, then activate SL-05 structural cleanup. |
| 2026-07-28T02:23:30-04:00 | Codex SL-05 execution | Structural cleanup is complete; SL-06 remains pending until this tracker gate passes | Deleted `internal/modules/collaboration/.gitkeep`; audited production Collaboration/platform aliases and imports, deleted platform semantic paths, producer-catalog paths, authored routing, contracts, tools, and guide references; retained the owner-local `ServerHarness` test-support alias because it remains an actively used scoped test composition rather than a compatibility surface | Exact static path/symbol/import scans; `make backend-module-boundary-check`; `make generate-drift` | Placeholder and obsolete production references are absent; boundary root `.cartulary/test-results/20260728T062219Z-p486241`; drift root `.cartulary/test-results/20260728T062221Z-p486514` | No blocker; generic generated-artifact tooling still recognizes `.gitkeep` sentinel filenames outside the Collaboration path and is intentionally unrelated to this cleanup. | Pass tracker diff/Markdown checks, then activate the SL-06 verification/accounting audit. |
| 2026-07-28T02:27:13-04:00 | Codex SL-06 execution | Verification and generated accounting are complete; SL-07 remains pending until this tracker gate passes | Audited every added or renamed Go test identity across Collaboration, platform WS, Revisions assembly, Jobs, and Network Flow; inspected owner, collaborator, verification, evidence, runtime, resource, and fixture profiles; rechecked the changed Collaboration durable-stream collaborator routing and OTel evidence; regenerated topology only through Make | Exact selector occurrence audit for 22 identities; `make explain-test-owner` for `module.collaboration`, `platform.ws`, `module.revisions`, `module.networkflow`, and `module.extensions`; `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make otel-conformance`; `make backend-module-boundary-check` | Every added/renamed identity resolves in exactly one authored row with a compatible profile. Roots: generation `.cartulary/test-results/20260728T062631Z-p495408`, drift `.cartulary/test-results/20260728T062637Z-p497604`, generated policy `.cartulary/test-results/20260728T062648Z-p501374`, shape `.cartulary/test-results/20260728T062650Z-p501720`, OTel `.cartulary/test-results/20260728T062653Z-p502208`, boundary `.cartulary/test-results/20260728T062700Z-p505637` | No blocker and no stale selector. Generated schedule/topology changes are projections of the authored owner rows, not hand edits. | Pass tracker diff/Markdown checks, then activate final validation and handoff SL-07. |
| 2026-07-28T03:47:53-04:00 | Codex SL-07 execution | End-to-end remediation and handoff complete; every `CI-*` row is `DONE` | Final diff covers 176 paths: five owner/guide/handoff documents, 157 backend/application/test-support paths including 18 new paths, and 14 authored or Make-generated verification/tooling paths. `docs/domain.md`, database migrations, public routes, WS JSON shapes, and generated protocol roots remain unchanged. | Every affected owner slice; complete service-backed matrix; explicit browser webserver/stateful suites; generation, drift, policy, shape, OTel, boundary, vulnerability, baseline-coverage, fast, full-check, and finalizer targets; exact evidence follows below | All applicable tests and accounting pass. Superseding final check root `.cartulary/test-results/20260728T080512Z-p2363828`; finalizer root `.cartulary/test-results/20260728T080750Z-p2460339`; final Entities owner root `.cartulary/test-results/20260728T074617Z-p2048101`. | No implementation, owner, security, generated-evidence, or validation blocker. Post-deployment operational verification remains deployment-time work because this task did not authorize a deployment or provide a deployment environment. | Preserve the adopted boundary; during the next deployment, perform readiness, dispatcher, connection/close metric, and real two-client reconnect/replay checks. |

#### SL-07 exact validation evidence

The final owner-slice runs all passed:

- `app.operator` `.cartulary/test-results/20260728T065010Z-p836755`;
  `app.server` `.cartulary/test-results/20260728T065010Z-p836727`;
  `module.artifacts` `.cartulary/test-results/20260728T065010Z-p836725`;
  `module.assessments` `.cartulary/test-results/20260728T065010Z-p836739`.
- `module.auth` `.cartulary/test-results/20260728T065122Z-p884085`;
  final `module.collaboration`
  `.cartulary/test-results/20260728T075350Z-p2080327`;
  `module.evidence` `.cartulary/test-results/20260728T065122Z-p884098`;
  final `module.entities`
  `.cartulary/test-results/20260728T074617Z-p2048101`.
- `module.extensions` `.cartulary/test-results/20260728T065423Z-p992242`;
  `module.imports` `.cartulary/test-results/20260728T065423Z-p992238`;
  `module.incidentbundles`
  `.cartulary/test-results/20260728T065423Z-p992251`;
  `module.incidents` `.cartulary/test-results/20260728T065423Z-p992263`.
- `module.indicators`
  `.cartulary/test-results/20260728T065554Z-p1053003`;
  `module.jobapi` `.cartulary/test-results/20260728T065554Z-p1053012`;
  `module.links` `.cartulary/test-results/20260728T065554Z-p1053035`;
  `module.networkflow`
  `.cartulary/test-results/20260728T065554Z-p1053056`.
- `module.parties` `.cartulary/test-results/20260728T065804Z-p1102523`;
  `module.projections`
  `.cartulary/test-results/20260728T065804Z-p1102543`;
  `module.reference_data`
  `.cartulary/test-results/20260728T065804Z-p1102531`;
  `module.reporting` `.cartulary/test-results/20260728T065804Z-p1102521`.
- `module.revisions` `.cartulary/test-results/20260728T065851Z-p1139919`;
  `module.tasksdecisions`
  `.cartulary/test-results/20260728T065851Z-p1139971`;
  `module.timeline` `.cartulary/test-results/20260728T065851Z-p1139935`;
  `module.workbook` `.cartulary/test-results/20260728T065851Z-p1139949`.
- `platform.telemetry`
  `.cartulary/test-results/20260728T070217Z-p1237533`;
  final `platform.ws`
  `.cartulary/test-results/20260728T075350Z-p2080332`;
  `web.collaboration`
  `.cartulary/test-results/20260728T070217Z-p1237527`.

The final complete service-backed matrix passed at
`.cartulary/test-results/20260728T072129Z-p1459139`. The late exact Entities
merge row passed independently at
`.cartulary/test-results/20260728T072052Z-p1456971` and is included in the
final Entities owner, fast, and check runs. The explicit browser targets passed:
webserver-backed at
`.cartulary/test-results/20260728T070226Z-p1238502` and stateful at
`.cartulary/test-results/20260728T070722Z-p1266457`.

Final generated and static evidence includes:

- `make generate`:
  `.cartulary/test-results/20260728T074116Z-p1937417`;
- `make generate-drift`:
  `.cartulary/test-results/20260728T074134Z-p1939772`;
- `make generated-artifact-policy-check`:
  `.cartulary/test-results/20260728T074134Z-p1939794`;
- `make json-shape-check`:
  `.cartulary/test-results/20260728T074134Z-p1939828`;
- `make go-test-duration-baseline-coverage`:
  `.cartulary/test-results/20260728T074134Z-p1939986`;
- `make otel-conformance`:
  `.cartulary/test-results/20260728T072541Z-p1538477`;
- `make backend-module-boundary-check`:
  `.cartulary/test-results/20260728T072541Z-p1538911`;
- `make go-vulncheck`:
  `.cartulary/test-results/20260728T072103Z-p1458402`;
- tracker `make lint-markdown` gate:
  `.cartulary/test-results/20260728T075107Z-p2073128`.

The superseding `make test-fast` passed 891 tests with no missing or unmapped
evidence at `.cartulary/test-results/20260728T075611Z-p2108640`. The retained
final `make check` passed all 172 work units and 747 tests with no failures,
missing, or unmapped evidence at
`.cartulary/test-results/20260728T080512Z-p2363828`.
`make agent-finalize` accepted that cache-warmed full-check root and passed all
run-health, shape, coverage, and drift checks at
`.cartulary/test-results/20260728T080750Z-p2460339`. The earlier successful
finalizer `.cartulary/test-results/20260728T074435Z-p2040997` refreshed the
three Make-owned browser/service/smoke timing artifacts; the superseding
finalizer found the generated state unchanged.

SL-07 exposed and closed five evidence gaps without weakening requirements:

1. the full service-backed matrix found one Network Flow fixture-profile
   mismatch and a direct Revisions-to-Projections module import; the authored
   transaction/clone profiles were corrected and assembly now projects a
   narrow Revisions-owned descriptor shape;
2. `test-fast` found stale generated shard identity after a duration refresh;
   topology was regenerated from the authored inputs;
3. `go-vulncheck` found two stale Entities merge test constructors outside the
   prior selectors; the tests now use a real minimal shared Revisions
   composition and have one exact authored owner row;
4. finalizer timing health found recurring Workbook store-shard imbalance;
   full service-backed coverage and full-check timing observations were merged
   without lowering the 15-second readiness or 1.25-times peer thresholds, and
   the rebalanced full check and finalizer passed;
5. the final static scan found that two external-package Collaboration tests
   still imported the third-party WebSocket client; generic frame/status aliases
   and JSON helpers moved to `internal/testutil/wstest`, leaving no third-party
   WebSocket import anywhere under the Collaboration package. The affected
   owner, fast, check, and finalizer runs all pass.

Compatibility remains deliberate: repository clients use text frames; binary
application frames now close `1003`; outbound JSON has no trailing LF; durable
intent/replay rows and resume tokens remain readable; Jobs v1 rows coexist
with v2 keys for new writes; public JSON shapes and storage schemas are
unchanged; no migration or data rewrite is required; rollback to the prior
binary remains storage-compatible.

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Mixed-responsibility but legitimate Collaboration module diagnosed | Inspected all 19 target files, server/operator/timeline/revision assembly, Revisions, source-owner callers, platform WebSocket, Jobs, Network Flow, migrations, and boundary inputs; touched only this tracker | `find`, `rg`, `sed`, `jq` | Source mapping, direct SQL, Store breadth, and semantic/platform seams classified | RB-002, RB-003 | Begin SL-00 only under later authorization. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Catalog injection and semantic/transport boundaries are decision-complete | Inspected live `ProviderContribution`, `Appender`, `revisionassembly`, projection descriptors, generated view-schema registry, platform WebSocket, Collaboration routes, server runtime, HTTP dependency set, and Hub consumers; touched only this tracker | `sed`, `rg`, `jq` | One immutable 17-route catalog, one shared appender, one semantic codec, one Hub, and one server adapter are now mandatory | RB-001 only | Execute SL-00, then implement SL-01 and SL-04 independently under authorization. |
| 2026-07-28T03:47:53-04:00 | Codex SL-07 execution | Backend owner graph is implemented and final | Collaboration owns semantic wire, Hub, telemetry, intent admission/replay/dispatch/recovery capabilities, and historical suppression; Revisions owns the immutable catalog and composition-scoped appender; source owners contribute routes and use narrow ports; platform WS is transport-only | Final affected owner slices, complete service-backed matrix, `make backend-module-boundary-check`, `make go-vulncheck`, `make test-fast`, `make check` | One codec, one Hub, one shared Revisions appender per composition, one Collaboration intent writer, no broad Store, no direct source-owner Collaboration SQL, and no production construction/import violation remain | None | Preserve narrow capabilities and source-owned contributions when adding future event families or record/view routes. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Frontend Collaboration session and Workbook effects confirmed as consumers/owners of client state | Inspected Collaboration session, browser socket transport, Workbook collaboration message/effect/controller files, frontend boundary inputs, and selected tests; touched only this tracker | `rg`, `sed`, `jq` | No backend/grid-vendor coupling found; pending queue remains frontend-owned | None for current plan | Preserve frontend behavior; run frontend owner/browser validation only when affected. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Frontend remains an unchanged contract consumer of the proposed backend seam migration | Reused live frontend consumer and boundary evidence; inspected no frontend edit source; touched only this tracker | `rg`, `sed` | Pending queues, workbook effects, selectors, reconnect/dedupe, and grid adaptation remain frontend-owned and unchanged | RB-001 only for later implementation | Run web owner and browser parity tests during SL-04; do not move client state. |
| 2026-07-28T03:47:53-04:00 | Codex SL-07 execution | Frontend remains an unchanged text-frame contract consumer | No frontend source file changed; backend test clients and shared WebSocket support moved to the Collaboration semantic API | `make test-slice OWNER=web.collaboration`; explicit browser webserver/stateful suites; full service-backed and check suites | Frontend owner and reconnect/live-update browser behavior pass; no binary-frame repository client was found | None | No client migration is required for repository consumers. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Authored WS and downstream generated surfaces mapped; no contract edit planned | Inspected WS index, protocol frontend entrypoints, protocol generator references, generated policy, migrations, and schema ownership manifest; touched only this tracker | `rg`, `sed`, `jq`, `make explain-target TARGET=generate-drift DETAIL=summary` | Internal refactor can preserve current schemas; generated roots remain generator-owned | Any future public contract change requires separate authorization | Keep authored contracts unchanged for the proposed behavior-preserving slices. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Existing projection-provider and view-schema projections are sufficient; no new manifest or WS contract change is planned | Inspected projection-provider manifest, view-schema index and canonical filters, generated registry, and current producer catalog; touched only this tracker | `jq`, `rg`, `sed` | All 17 current routes are mapped; generated artifacts remain downstream and hand edits are prohibited | RB-001 only | Validate parity through existing projections in SL-00/SL-01 and require clean generated drift. |
| 2026-07-28T03:47:53-04:00 | Codex SL-07 execution | Owner specifications and generated projections agree | Core 01, Core 03, and Core 04 now own the explicit frame/resource/failure/live-intent rules; `contracts/ws`, generated protocol roots, migrations, and `docs/domain.md` remain unchanged; authored verification and timing inputs generate the final scheduler/topology | `make generate`; `make generate-drift`; generated policy and JSON-shape checks; `make otel-conformance`; finalizer drift/shape checks | No semantic WS artifact drift, unsupported generated edit, schema change, or data migration | None | Any future JSON shape or `/ws/v2/` change requires its own owner and migration work. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Collaboration has 28 active owner rows, 19 service-backed; focused unit accounting gaps recorded | Inspected verification owner, module/web test-family catalogs, test-support inventory, target tests, and referenced consumer tests; touched only this tracker | `make help`, `make help-all`, `make task-guide ROLE=module-author OWNER=module.collaboration`, `make explain-test-owner OWNER=module.collaboration`, `make task-guide ROLE=module-author OWNER=web.collaboration`, `make explain-target` for focused targets, plus `jq`/`rg` | Canonical validation ladder discovered; no tests or product validation run | Later characterization tests need authored row decisions | Under later authorization, implement SL-00 and update authored routing before regeneration. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Characterization requirements now trace to `CAT-001` through `CAT-004` and `WSB-001` through `WSB-004` | Inspected exact tests, constructor sites, source-owner callers, third-party decoder defaults, and existing harness rows; touched only this tracker | `rg`, `sed`; no Make test target run | Required golden, construction-guard, catalog-failure, rollback, restore, Origin, close, frontend, and browser evidence is explicit | RB-001 only | Add and route exact SL-00 tests before either seam moves. |
| 2026-07-27T23:17:56-04:00 | Codex SL-00 execution | Direct characterization is routed and green | Changed Collaboration catalog/historical tests, Jobs and Network Flow integration tests, four authored family catalogs, and Make-generated topology/schedules | Focused and service-backed owner slices; `make generate`; `make generate-drift`; `make json-shape-check` | Exact tests route once with appropriate service profiles; the Section 4.6 matrix separates preserved behavior from known defects | None | Move semantic WebSocket tests with their implementation during SL-04 and update authored rows in that slice. |
| 2026-07-28T03:47:53-04:00 | Codex SL-07 execution | Final accounting is complete: 33 active Collaboration rows, 21 service-backed, and every remediation test routes exactly once | Added exact catalog, suppression, producer, codec/failure, telemetry, and late Entities merge rows; corrected Network Flow transaction/clone profiles; regenerated schedules from merged complete-coverage and full-check timing observations | All affected owner slices; full service-backed matrix; baseline coverage; `test-fast`; `check`; `agent-finalize` | No stale selector, missing/unmapped test, incompatible fixture profile, timing-health failure, or generated drift remains | None | Use owner slices for future narrow work and refresh full evidence before pruning duration observations. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Session, membership, Origin, heartbeat, revocation, and closed-incident behavior frozen | Inspected Core 04, Collaboration routes/notifier, Auth/Incidents callers, socket/integration tests, and platform presence/revocation tests; touched only this tracker | `sed`, `rg` | Current authorization is rederived and incident-scoped; no contradiction found | None for current plan | Treat authorization parity tests as mandatory for route, Hub, or notifier movement. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Origin precedence, frame limits, authorization rechecks, heartbeat, revocation, closure, and slow-consumer failures have exact preserved defaults | Inspected Core 01/Core 04, Collaboration route order, platform acceptor, `coder/websocket` v1.8.14 read/JSON behavior, and relevant tests; touched only this tracker | `rg`, `sed` against repository and local pinned module source | Pre-upgrade Origin gate, 32,768-byte limit, accepted frame kinds, and close mappings are explicit; no new security semantics were invented | RB-001 only | Prove every `WSB-003` case during SL-00 and SL-04. |
| 2026-07-28T03:47:53-04:00 | Codex SL-07 execution | Security and lifecycle defaults are implemented and validated | Strict semantic decoder, raw socket adapter, fail-closed route dependencies, Origin gate, authorization rechecks, revocation/closure handling, safe telemetry classification, and family validation | Collaboration/platform/server/Auth owner slices; explicit browser suites; OTel conformance; boundary checks; `make go-vulncheck`; full check | Stable close codes/reasons, text-only application frames, 32,768-byte limit, duplicate rejection, replay/session binding, privacy vocabulary, and exporter-failure invariance pass | None | Monitor close classifications, quarantines, and active connection/event signals after deployment. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Planning tracker complete; implementation intentionally not started | Inspected repository state and this tracker; touched only this tracker | `git status --short --branch`, inventory/search and Make discovery commands; final read-only diff checks planned | No production refactor, test edit, contract edit, generated edit, migration, or harness edit | RB-001, RB-002, RB-003 | Authorize SL-00 first, then close RB-002 before SL-01; leave SL-04 deferred until RB-003 closes. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | RB-002 and RB-003 are closed as architecture decisions; implementation remains intentionally unstarted | Inspected supporting analysis, live types/defaults, the complete tracker diff, and final status; touched only this tracker | `git diff --check -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git diff --check HEAD -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git diff --stat HEAD -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git status --short`; no product validation run | Both diff checks passed; status reported only the existing staged tracker addition plus its unstaged revision; no production, test, contract, generated, migration, package, or harness change occurred | RB-001 only | Authorize and execute SL-00; then implement the resolved catalog cutover and raw-socket split as independent slices. |
| 2026-07-27T23:17:56-04:00 | Codex SL-00 execution | Baseline complete; implementation deviations remain intentionally unfixed until their authorized slices | Inspected final characterization diff and Make evidence | `git diff --check`; owner/service slices; generation/drift/shape checks | Known defects are explicit: platform semantic ownership, binary acceptance, trailing LF, multiple appenders, duplicate catalog, direct producer SQL, raw suppression string, and drifting telemetry vocabulary | External supported binary-frame client discovery remains the SL-04 stop condition; none is present in repository clients | Pass the tracker gate and begin SL-04. |
| 2026-07-28T03:47:53-04:00 | Codex SL-07 execution | No remediation blocker remains | Inspected final status/diff, all retained run summaries, generated projections, static ownership scans, and tracker completion criteria | Final validation ladder and tracker gate | All implementation and local validation criteria pass. No supported binary-frame client was discovered. No application check was skipped. Deployment-only health verification was not run because no deployment was authorized or available. | None for code handoff; deployment observations are intentionally pending until rollout | Deploy as one code-only binary when authorized, then verify readiness, dispatcher failure/quarantine counts, connection/event/close metrics, and one real two-client reconnect/replay scenario. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | `RESOLVED`: the 2026-07-27 task explicitly authorizes WS-00 and SL-00 through SL-07 with the per-slice tracker gate. | Removes the planning-only prohibition while preserving owner precedence and incremental validation. | The current task text and Section 7 gate | RESOLVED |
| RB-002 | `RESOLVED`: `revisionassembly` collects data-only record/view routes from the existing source-owner `revisions.ProviderContribution` values, validates the exact current-profile catalog against projection descriptors and the generated view-schema registry, constructs one immutable `RecordViewCatalog`, and constructs one scoped `*revisions.Appender` that is injected into every source-owner service in that composition. Revisions owns validation and coordination; source owners own their contributions; Collaboration remains the sole durable intent writer. Package-global registration, per-module appender construction, fallback catalogs, shadow writes, and dual writers are prohibited. | The decision removes the duplicated Collaboration catalog without an import cycle or a second event owner. | SL-00 characterization; `CAT-001` through `CAT-004`; complete constructor inventory; catalog parity; rollback, historical-suppression, byte-parity, single-intent, and boundary-guard evidence | RESOLVED — implemented and validated in SL-01 |
| RB-003 | `RESOLVED`: Collaboration owns the exact raw-socket consumer port in Section 4, the sole semantic DTO/codec family, Hub, presence, hello/resume, application heartbeat, replay/live delivery, authorization choreography, terminal behavior, and semantic telemetry. Platform owns only configured upgrade/Origin mechanics, frame I/O, the explicit read limit, write serialization, low-level control frames, and close mechanics. Server supplies the pre-upgrade Origin gate and concrete adapter. Semantic types move rather than copy; no platform alias, duplicate codec, second Hub, or generated-contract change may remain. | The decision restores lower-level platform direction while adopting text-only application messages, strict invalid-traffic handling, and LF-free semantic encoding. | SL-00 characterization; `WSB-001` through `WSB-004`; golden live/replay bytes; Origin, limits, heartbeat, presence, revocation, closure, boundary, frontend, browser, and generated-drift evidence | RESOLVED — implemented and validated in SL-04 |

There is no open authorization blocker. RB-001 through RB-003 are retained by
stable ID as closed decisions and MUST NOT be reintroduced as implementation
discretion. No owner contradiction is currently known.

## 12. Binary Completion Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/collaboration` is inventoried or explicitly out of scope. | PASS | Section 2 accounts for the original inventory, all 27 current files, the SL-05 placeholder deletion, and the two SL-01 deletions. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 covers route, session, presence, replay, events, views, storage, authorization, telemetry, frontend, generated protocol, and harness accounting. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 defines WF-00 through WF-08 with predecessors, successors, validation, and checkpoints. |
| Every authorized implementation slice has a dependency, risk, rollback posture, validation, and binary exit condition. | PASS | Section 7 defines WS-00 and SL-00 through SL-07 and the mandatory tracker gate. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 records current Make-owned commands and applicability. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No contradiction was found; Sections 1, 5, and 11 define the required fail-closed posture if one appears. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Sections 3 and 5 record the framework’s narrower catalog versus the live target. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Section 10 records scope, backend/frontend, contract/codegen, harness, security, risks, commands, and next actions. |
| The execution baseline and explicit compatibility changes are adopted before implementation begins. | PASS | WS-00 updates Core 01/Core 03/Core 04 and Sections 1, 4, 5, 7, 9, and 11. |
| Normative tracker language is subordinate to adopted owner documents and does not create public behavior. | PASS | Section 1 states the authority boundary and fail-closed contradiction rule. |
| RB-002 defines one owner graph, one complete immutable catalog, one composition-scoped appender, and one durable intent writer. | PASS | Sections 3.1, 4.1, 4.2, 4.5, 7, and 11 define `CAT-001` through `CAT-004` without an unresolved design choice. |
| RB-003 defines one owner graph, one raw-socket port, one codec, one Hub, exact defaults, and exact failure mapping. | PASS | Sections 3.2, 4.3, 4.4, 4.5, 7, and 11 define `WSB-001` through `WSB-004` without an unresolved design choice. |
| Existing machine projections are sufficient for catalog validation and no parallel authority is proposed. | PASS | Section 4.1 names the projection-provider and generated view-schema evidence and prohibits a new manifest. |
| Every normative catalog requirement has a binary acceptance condition and required evidence. | PASS | `CAT-001` through `CAT-004` cover completeness, construction, transaction behavior, cutover, and storage boundaries. |
| Every normative WebSocket-boundary requirement has a binary acceptance condition and required evidence. | PASS | `WSB-001` through `WSB-004` cover imports/ownership, wire bytes, security/lifecycle behavior, and generated drift. |
| Core owner changes precede implementation and the development-guide follow-up is explicit. | PASS | Core 01/Core 03/Core 04 and Sections 1, 4.5 `DOC-001`, and SL-04 agree. |
| The top-level tracker contains no duplicate ID and uses only the allowed status vocabulary. | PASS | Section 9 has unique `CT-*` and `CI-*` rows using `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, or `DROPPED`. |
| RB-001 through RB-003 are resolved and no authorization blocker remains. | PASS | Sections 7, 9, and 11. |
| Every authorized implementation and completion slice is done on the final source state. | PASS | Every `CI-*` row is `DONE`; Section 10 records all affected owner slices, full service-backed/browser/fast/check evidence, and the passing finalizer root. |
| Compatibility and rollout posture are explicit and require no hidden migration. | PASS | Section 10 records text-only/LF-free wire cleanup, durable Jobs v1/v2 coexistence, unchanged JSON/storage schemas, binary rollback compatibility, and deployment-only health checks. |

The tracker is complete only while every `PASS` statement above remains true.
Any later repository change that invalidates a statement MUST change it to
`FAIL` or `BLOCKED`, identify the exact evidence gap, and prevent a completion
claim until the gap is closed.
