# collaboration Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Value |
| --- | --- |
| Target path | `internal/modules/collaboration` |
| Target label | `collaboration`, derived from the final target-path component and normalized to lowercase kebab case |
| Output path | `docs/handoffs/collaboration-module-refactor-tracker.md` |
| Repository state inspected | `main` at `8ce5d3f6640e4f7054ebe06420ac6180afc49f38`; worktree was clean before this tracker was created |
| Planning status | Planning and documentation only |
| Allowed change | This tracker file only |
| Non-goals | No production refactor, behavior change, test edit, contract edit, generated-artifact edit, dependency change, migration, harness change, or package reconfiguration |
| Later authorization | Every implementation slice in this tracker requires a later explicitly authorized task |

The target exists and contains 19 files. This tracker treats the planning
framework as doctrine and a template, not as evidence of current repository
state. Live source, contracts, tests, ownership projections, and Make-owned
command discovery were inspected before recording repository claims.

Normative language in this tracker constrains only the proposed refactor and
its acceptance boundary. It does not create, amend, or supersede public product
requirements. A later implementer MUST apply this tracker beneath the authority
order below, MUST preserve owner-defined observable behavior, and MUST stop with
`BLOCKED: owner contradiction` if an adopted owner document conflicts with a
requirement recorded here.

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
  ownership description must be narrowed only after an authorized RB-003
  implementation.

These supporting documents MUST NOT be treated as behavioral authorities.
Core 01, Core 03, and Core 04 already own the behavior frozen by this tracker;
their amendment is not a prerequisite for the behavior-preserving internal
refactor. Core 02 requires no amendment because no logical or physical storage
change is planned. A later public behavior or owner-boundary change requires
the ordinary adopted-document process and is outside this tracker.

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
| `internal/modules/collaboration/.gitkeep` | Obsolete directory placeholder | None | None | None | None | None | Collaboration repository housekeeping | low | Package is non-empty; remove only in a later authorized implementation slice. |
| `internal/modules/collaboration/historical_restore.go` | Sets the transaction-local suppression flag used to prevent live Collaboration intents during historical restore | `SuppressHistoricalIntentsTx` | `internal/modules/incidentbundles/source.go` | `pgx.Tx`; Postgres `set_config` | No direct named test; rollback tests set the same flag through SQL | Collaboration durable-intent behavior; no generated artifact | Collaboration, exposed as a narrow Incident Bundles integration port | medium | Behavior needs direct characterization before movement. |
| `internal/modules/collaboration/incident_session_notifier.go` | Converts incident close and membership revocation into active-session termination and WebSocket notifications | `IncidentSessionNotifier`, `NewIncidentSessionNotifier`, `NotifyIncidentClosed`, `NotifyIncidentMembershipRevoked` | Server runtime composition and Incidents collaboration-session port | `platform/authn`, `platform/postgres`, `platform/ws.Hub` | Collaboration revocation integration tests and Incidents tests | Core WebSocket `error` and `session_revoked` contracts | Collaboration session orchestration | high | Legitimate facade; platform Hub dependency participates in the semantic/transport seam. |
| `internal/modules/collaboration/integration_test.go` | Real-server characterization for two-client presence/replay, revocation sources, closed incidents, replay filtering, and Origin rejection | Test symbols only | Collaboration owner test catalog | Auth, Incidents, Jobs, Timeline, platform WebSocket, shared test support | The file itself | `contracts/ws/index.schema.json`; module Collaboration evidence rows | Collaboration verification | high | Service-backed behavior evidence. |
| `internal/modules/collaboration/producer_catalog.go` | Admits record producers and maps record types or artifact prefixes to public `view_schema_id` values | `RecordProducerDescriptor`, `RecordProducerCatalog`, `RecordProducerViewSchema` | `internal/modules/revisions/appender.go`; local tests | Standard library only; hard-coded identities owned by Artifacts, Assessments, Entities, Evidence, Indicators, Parties, Tasks/Decisions, and Timeline | `producer_catalog_test.go` | Public view-schema identities consumed by `record_changed`; related UI/protocol projections are downstream | Source-owner contributions aggregated for Revisions, not Collaboration | high | `must_fix`: duplicates knowledge already distributed across source owners and assembly catalogs. |
| `internal/modules/collaboration/producer_catalog_test.go` | Characterizes artifact prefix and deleted-row source-type mapping | Test symbols only | Go package test execution | `producer_catalog.go` | The file itself | Public view-schema identities | Collaboration verification until mapping ownership moves | medium | Current module owner-family rows do not name these tests. |
| `internal/modules/collaboration/record_change_intent.go` | Canonicalizes changed cell keys and builds deterministic `record_changed` semantic intents | `RecordChange`, `ChangedCellKeys`, `NewRecordChangeIntent` | Revisions; Entities mentions and merge; Evidence; Timeline assembly and source-owner adapters | UUIDs and semantic payload helpers in `internal/platform/ws` | `record_change_intent_test.go`, shared harness, stream/integration tests, source-owner tests | `record_changed` in `contracts/ws/index.schema.json` and generated protocol TypeScript | Collaboration semantic intent construction | high | Keep semantic intent ownership; reverse or narrow the platform semantic dependency later. |
| `internal/modules/collaboration/record_change_intent_test.go` | Characterizes sorted, compact patch construction and deterministic intent identity | Package-private test helper | `TestSocketEventInventoryCoverage` in `shared_harness_test.go` | Collaboration public intent API | The file itself through shared harness | `record_changed` payload and field-key conformance | Collaboration verification | medium | Accounted indirectly through the shared harness row. |
| `internal/modules/collaboration/routes.go` | Registers and serves the incident WebSocket route, handshake/resume, replay, presence, heartbeat, authorization rechecks, terminal errors, and close behavior | `Service`, `Settings`, `RegisterRoutes` | `internal/app/server/runtime.go`, `internal/app/server/module_settings.go` | Incidents access, Authn, HTTP API/auth, platform WebSocket, durable `Store` | `socket_test.go`, `integration_test.go`, testsupport clients, browser/e2e tests | Full Collaboration WebSocket contract and generated message types | Collaboration transport-adjacent application adapter | critical | Exact route and externally visible session behavior must not change. |
| `internal/modules/collaboration/shared_harness_test.go` | Validates semantic harness inventory and indexes live tests for socket obligations | `TestSocketEventInventoryCoverage`, `TestSocketLifecycleEvidenceIndex` | Collaboration owner test catalog | Owner-local incident WebSocket test support | The file itself plus referenced tests | Verification/accounting projections only | Collaboration verification | medium | Evidence routing is not runtime architecture. |
| `internal/modules/collaboration/socket_test.go` | Service-backed route tests for closed message vocabulary, resume reset, heartbeat expiry, and incident-scoped ephemeral presence | Test symbols; test-only `SessionTiming` | Collaboration owner test catalog | Auth, Incidents, Timeline, platform WebSocket, owner-local test support | The file itself | WebSocket contract and security behavior | Collaboration verification | high | Despite `_Unit` names, catalog rows use managed-service integration profiles. |
| `internal/modules/collaboration/stream.go` | Owns durable intent insertion, idempotency collision checks, per-incident sequencing, replay storage, resume tokens, dispatch retries/quarantine, retention, and operator requeue | Event-family constants, `ErrIntentKeyCollision`, `EventIntent`, `IntentAppender`, `Store`, `NewStore`, `NewEventIntent`, `AppendIntentTx`, `ReplayResult`, `CurrentHighWater`, `IssueResumeToken`, `ReplayMessages`, `Dispatcher`, `NewDispatcher`, `Start`, `Close`, `RunOnce`, `RequeueIncident` | Server runtime, Operator, Timeline assembly, Revisions, Evidence, Entities, and other source-owner adapters | Postgres/pgx, crypto, platform WebSocket messages and validation | `stream_integration_test.go`, route/socket/integration tests, source-owner integration tests | Collaboration DB ownership manifest; replayable WebSocket event contracts | Collaboration durable stream and persistence adapter | critical | Legitimate Collaboration core, but `Store` exposes several independent capabilities. |
| `internal/modules/collaboration/stream_integration_test.go` | Characterizes atomic insertion, exact replay/collision, post-commit sequencing, outage/restart, fan-out, quarantine/requeue, hashed tokens, and conjunctive retention | Test symbols only | Collaboration owner test catalog | Postgres, server harness, Timeline, Auth, platform WebSocket | The file itself | Durable-stream DB contract and replay behavior | Collaboration verification | critical | Primary durable-stream behavior evidence. |
| `internal/modules/collaboration/telemetry.go` | Emits Collaboration WebSocket lifecycle and dispatcher-run telemetry using allowlisted vocabulary | Package-private methods on `Service` and `Dispatcher` | `routes.go`, `stream.go` | Platform HTTP API and telemetry; OpenTelemetry API | `telemetry_test.go`; OTel conformance references | OTel source snapshot/generated constants are downstream owner projections | Collaboration instrumentation using platform telemetry primitives | high | Signal names and privacy rules are observable operational contracts. |
| `internal/modules/collaboration/telemetry_test.go` | Characterizes public-error classification and closed telemetry vocabulary | Test symbols only | Go package test execution; OTel conformance source reference | Collaboration telemetry helpers, HTTP API errors | The file itself | OTel conformance and generated-constant evidence | Collaboration/Telemetry verification | medium | Current Collaboration owner-family rows do not name these tests. |
| `internal/modules/collaboration/testsupport/incidentwstest/incidentwstest.go` | Owner-local real WebSocket client, hello/resume parsing, dial assertions, terminal-event assertions, and presence sorting | `ConnectOptions`, ack DTOs, `Client`, connection/assertion helpers | Collaboration, Auth, Entities, Evidence, Incidents, Links, Timeline, Workbook, and other integration tests | platform WebSocket messages, generic `internal/testutil` HTTP/WebSocket helpers | Broad integration suites | WebSocket contract as test support; no generated output | Collaboration owner-local test support | medium | Explicitly registered as `owner_local`, service-starting support in `tools/test_support_inventory.json`. |
| `internal/modules/collaboration/testsupport/incidentwstest/inventory.go` | Declares semantic harness capabilities required by Collaboration socket evidence | Harness ID/requirement types, constants, inventory functions | `shared_harness_test.go` | Testing only | Shared harness tests | Harness/accounting projections only | Collaboration owner-local test support | low | Verification inventory must not become product authority. |
| `internal/modules/collaboration/testsupport/incidentwstest/view_events.go` | Connects view sockets and waits for `record_changed` events without leaking transport details to source-owner tests | `RecordChangeSocketPayload`, `ConnectViewSocket`, `RequireRecordChanged`, `ExpectNoSocketMessage` | Entities, Evidence, Links, Timeline, Workbook, and related integration tests | platform WebSocket and generic test utilities | Importing source-owner integration tests | `record_changed` contract as test support | Collaboration owner-local test support | medium | Remains semantic owner support unless generalized beyond Collaboration. |
| `internal/modules/collaboration/testsupport/scenariotest/harness.go` | Starts shared application/runtime fixtures for Collaboration scenarios | `RuntimeHarness`, `ServerHarness`, `StartRuntime`, `StartServer` | Collaboration service-backed tests | `internal/testutil/appsupport`, HTTP API, HTTP test helpers | Collaboration integration/socket/stream tests | Harness execution only | Collaboration owner-local test support | medium | Explicit service-starting support; not production assembly. |

All 19 files are in scope for diagnosis. No file is omitted.

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
| Presence and live delivery semantics | `routes.go` plus `internal/platform/ws.Hub` | Collaboration semantics; platform transport/runtime mechanics | split | Core 01 and Core 03; platform Hub implementation and tests | RB-003 defines the dependency-safe split. Implementation remains blocked only by RB-001 and SL-00 characterization. |
| Record-change canonicalization and deterministic event construction | `record_change_intent.go` | Collaboration | keep | Core 01 event contract; Core 03 sequencer rule; source-owner callers | Platform-owned semantic payload helpers should be narrowed later. |
| Record type and artifact prefix to public view-schema mapping | `producer_catalog.go` | Source-owner `revisions.ProviderContribution` values aggregated by `revisionassembly`, validated by Revisions | move | Existing `RevisionProviderContribution()` producer functions, `ProviderContribution`, projection descriptors, and repository composition rule | RB-002 defines one immutable catalog and one appender per composition. Public `view_schema_id` values MUST NOT change and dual event writers are prohibited. |
| Historical live-intent suppression during bundle restore | `historical_restore.go`, read by Revisions | Collaboration port used by Incident Bundles/Revisions | keep / split | Incident Bundles caller and transaction-local setting | Add direct characterization before changing the seam. |
| Incident close/membership revocation notification | `incident_session_notifier.go` | Collaboration facade with an Incidents-owned narrow port | keep | Server composition, Incidents tests, revocation integration tests | Do not move incident authorization decisions into Collaboration. |
| Job progress semantic intent construction | `internal/platform/jobs/collaboration_producer.go` | Jobs semantics with Collaboration persistence port | split | Core 03 source-owner intent rule; direct SQL inspection | Jobs must retain authoritative transaction ownership. |
| Network Flow invalidation semantic intent construction | `internal/modules/networkflow/collaboration_producer.go` | Network Flow semantics with Collaboration persistence port | split | Extensions and Network Flow NLSpecs; direct SQL and integration test | Preserve typed owner-admitted invalidation and privacy. |
| WebSocket messages, payload validators, patch builders, presence, Hub, and wire I/O | `internal/platform/ws` | Collaboration semantic wire package plus platform transport primitives | split | Platform source, Collaboration route source, and presence transport tests | RB-003 fixes the final owner graph. No generated-contract change is permitted. |
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

### 4.4 Preserved WebSocket Defaults and Failure Mapping

| Behavior | Required preserved value |
| --- | --- |
| Public route | Exactly `GET /ws/v1/incidents/{incident_id}` |
| Inbound application frames | Text and binary frames containing valid JSON are accepted, matching the current `wsjson.Read` behavior |
| Outbound application frames | Text frames only; serialized semantic bytes, field ordering, and current trailing LF behavior remain unchanged |
| Message read limit | Explicitly 32,768 bytes per message |
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
| Malformed JSON | `1007`, reason `failed to unmarshal JSON` |
| Message larger than 32,768 bytes | `1009` |
| Invalid first message | `1008`, reason `invalid_first_message` |
| Invalid later semantic message | Send the current `error` envelope, then `1008`, reason `invalid_message` |
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
| `WSB-002` | Live and replay delivery produce byte-equivalent canonical envelopes through one codec, including current text-frame and LF behavior. | Golden live/replay wire fixtures |
| `WSB-003` | Origin, authorization, hello/resume/reset, heartbeat, presence, revocation, incident-close, slow-consumer, and failure mappings equal the frozen behavior above. | Collaboration, platform, Auth, Incidents, frontend, and browser tests |
| `WSB-004` | Authored WS contracts and generated TypeScript projections have no semantic diff. | `make generate-drift` and generated policy checks |
| `DOC-001` | The development guide describes platform as transport-only after WSB completion and does not contradict Core or the implemented boundary. | Documentation review in the authorized WSB slice |

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

## 7. Proposed Refactor Slice Plan

All slices below require later authorization. They preserve observable behavior
unless a row explicitly says otherwise. No slice in this tracker authorizes a
public API, WebSocket schema, storage-schema, or product behavior change.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | WF-03 | `requires later authorization`: add characterization for the complete 17-route mapping, historical suppression, Jobs/Network Flow producer atomicity and collision parity, current WebSocket bytes/defaults/failures, and telemetry. Update authored accounting for new exact test identities. | Collaboration tests; Revisions; Incident Bundles; Jobs; Network Flow; platform WebSocket; authored test-family owners | Tests MUST freeze current behavior and MUST NOT introduce a new product requirement. | Existing owner rows plus exact catalog, suppression, producer, WebSocket golden, and telemetry rows | `make test-slice OWNER=module.collaboration`; `make service-backed-test-slice OWNER=module.collaboration`; affected owner slices | Revert tests and authored rows together; regenerate outputs through Make. | Requirements `CAT-001` through `CAT-004` and `WSB-001` through `WSB-004` have direct failing-when-broken evidence against unchanged behavior. |
| SL-01 | SL-00 | `requires later authorization`: extend `revisions.ProviderContribution` with the closed route descriptors in Section 4; build one immutable catalog and one shared appender in `revisionassembly`; validate against projection descriptors and generated view schemas; switch every production constructor; remove the Collaboration catalog in the same cutover. | Collaboration producer catalog; Revisions provider/appender; revision assembly; Projections/view-schema registry; Artifacts, Assessments, Entities, Evidence, Indicators, Parties, Tasks/Decisions, Timeline | Public view IDs, patch/invalidate/remove, change-set identity, source transaction, historical suppression | Complete catalog parity; missing/duplicate/ambiguous validation; construction guard; source-owner mutation, rollback, restore, and socket tests | Affected owner slices; `make backend-module-boundary-check`; `make generate-drift` | Revert the whole code-only cutover. The old catalog MAY exist as a test oracle but MUST NOT remain production-reachable. | `CAT-001` through `CAT-004` pass; all 17 routes contribute exactly once; emitted bytes are unchanged; no production source calls `NewAppender`; `producer_catalog.go` is removed. |
| SL-02 | SL-00 | `requires later authorization`: introduce narrow transaction-aware Collaboration intent persistence adapters for Jobs and Network Flow. Retain source-owned payload/key construction and authoritative transaction ownership; replace duplicated insert/collision SQL with the common validation and persistence path. | Collaboration intent appender; app/server assembly; platform Jobs; Network Flow | Source atomicity, idempotency, deterministic keys, collision failure, extension privacy | Producer-specific atomicity/replay/collision tests; durable stream and Network Flow socket tests | Affected owner slices; `make service-backed-test-slice OWNER=module.collaboration`; `make backend-module-boundary-check` | Revert adapters and callers together; no DB migration is planned. | Jobs and Network Flow no longer access Collaboration tables directly, with byte-identical payloads and unchanged transaction behavior. |
| SL-03 | SL-01, SL-02 | `requires later authorization`: split Collaboration internal capabilities for intent append, replay/token queries, dispatch lifecycle, and operator recovery while keeping an exported compatibility facade only until each caller has moved to its narrowest port. | `stream.go`; server/operator/timeline assembly; source-owner adapters | Capability wiring, lifecycle close, retry/requeue, nil configuration | Existing stream, runtime, operator, and source-owner tests | Collaboration owner slices; `make backend-module-boundary-check` | Revert caller conversions independently; remove temporary compatibility only after the last caller moves. | Each production caller receives only its required capability; no temporary adapter is left; sequence, replay, and storage behavior are unchanged. |
| SL-04 | SL-00 | `requires later authorization; ready after characterization`: implement the exact raw-socket port and owner graph in Sections 3 and 4; move semantic DTO/codec/patch behavior into `collaboration/wire`; move Hub/presence into Collaboration; adapt in server; remove generic and unused Hub dependencies; update the development guide. Authored WS schemas and generated protocol semantics MUST NOT change. | Collaboration routes/intents/wire; `internal/platform/ws`; `internal/platform/httpapi`; server assembly/runtime; Auth; Incidents notifier; Imports; Job API; Reference Data; development guide; frontend only as unchanged consumer | Highest risk: envelope bytes, frame kinds, close mappings, presence order, replayability, authorization filtering, Origin precedence, telemetry, import cycles | Platform WS contract/presence tests; Collaboration socket/integration and golden wire tests; Auth/Incidents tests; frontend and browser tests | Collaboration and web owner slices; boundary checks; browser stateful/webserver-backed; `make generate-drift` | Perform one semantic ownership cutover with one codec and one Hub; revert the whole seam if parity fails. A duplicate implementation or alias is not an allowed rollback bridge. | `WSB-001` through `WSB-004` and `DOC-001` pass; platform is transport-only, Collaboration is semantic owner, and every frozen byte/default/failure remains equivalent. |
| SL-05 | SL-03 | `requires later authorization`: remove obsolete `.gitkeep` after the non-empty package layout is stable. | `internal/modules/collaboration/.gitkeep` | None | None beyond ordinary repository checks | `make backend-module-boundary-check` | Restore the placeholder if repository tooling unexpectedly requires it. | Placeholder is absent and no tooling refers to it. |
| SL-06 | SL-00 through the last implemented behavior slice | `requires later authorization`: update authored verification owner/family inputs only for new or moved exact tests; run the owner generator and drift checks. Never edit generated topology, schedule, or protocol output by hand. | Contracts verification owner, test-family owner, execution-topology owner only when required; generated outputs through Make | Lost or duplicated evidence rows; unsupported generated drift | Owner slices and catalog/accounting checks | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Revert authored input and generated projection from the same slice. | Every active changed test is routed exactly once and generated drift is clean. |
| SL-07 | Implemented slices and SL-06 when applicable | `requires later authorization`: perform final narrow-to-broad validation, retained evidence review, tracker update, and handoff. | All files changed by the later authorized task; tracker | False success from stale evidence or unrelated failures | All preserved contract tests | `make agent-finalize`; affected owner slices; `make test-fast`; `make check` | Revert only the failing implementation slice; do not rewrite requirements or generated evidence to make a failure pass. | Every applicable Section 4 requirement passes or the exact failure and run root are recorded; the tracker names the next action without claiming false success. |

Any later proposal to alter public messages, route shapes, retention,
authorization, storage schema, or frontend behavior is a behavior change and
must be marked `requires later authorization` before implementation.

## 8. Validation Plan

Commands were discovered from the current public Make surface using
`make help`, `make help-all`, `make task-guide`,
`make explain-test-owner`, and `make explain-target`. Row and phase maps are
verification accounting only.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.collaboration` | All active Collaboration owner rows when `ROWS` is omitted, including unit/frontend and service-backed rows selected by the owner catalog | yes | Use exact `ROWS` only for a deliberately bounded slice; do not infer tests from filenames. |
| integration | `make service-backed-test-slice OWNER=module.collaboration` | The 19 currently service-backed Collaboration rows | yes | Covers real Postgres/server/browser-backed profiles owned by the catalog. |
| frontend unit | `make test-slice OWNER=web.collaboration` | Frontend Collaboration session regression | no | Required when frontend imports or the session consumer change; otherwise use as final contract-consumer confidence where risk warrants. |
| e2e/browser | `make browser-e2e-webserver-backed` and `make browser-e2e-stateful` | Cross-stack Collaboration, revocation, multi-client, pending-edit, and live-update behavior | no | Required before final handoff for route/protocol/presence changes; add a11y/visual only when visible behavior or fixtures change. |
| generated drift | `make generate-drift` | Generated protocol, harness topology, schedules, and other policy-owned outputs | yes | Mandatory after authored contract or harness owner input changes; generated files are never hand-edited. |
| generated policy/shape | `make generated-artifact-policy-check` and `make json-shape-check` | Generated-root policy and machine projection shapes | no | Required with SL-06 or any authored machine-contract edit. |
| import-boundary/static | `make backend-module-boundary-check` | Backend module imports, source ownership, and SQL/access constraints | yes | Run after each ownership/capability slice. |
| frontend import boundary | `make frontend-import-boundary-check` | Frontend Collaboration/Workbook/grid import boundaries | no | Required only if frontend files or imports move. |
| focused aggregate | `make test-fast` | Common backend/frontend and service-backed verification | no | Run after focused owner slices pass. |
| full check | `make check` | Current default local repository verification | no | Final broad validation; run `make agent-finalize` first as required by repository procedure. |

For this tracker-only session, no product validation, Markdown harness target,
or implementation test was run. Both tracker-scoped `git diff --check` commands
passed. `git status --short` reported only
`AM docs/handoffs/collaboration-module-refactor-tracker.md`; the existing staged
addition was not modified, and this revision remains the sole unstaged path.

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
| CI-001 | Add mapping, suppression, producer, WebSocket, and telemetry characterization/accounting | WF-03, WF-07 | BLOCKED | RB-001 | SL-00 | Every `CAT-*` and `WSB-*` invariant has direct evidence without changing product behavior. |
| CI-002 | Move record/view resolution to source-owner contributions | WF-05, WF-06 | BLOCKED | CI-001, RB-001 | SL-01 | `CAT-001` through `CAT-004` pass and the complete validated catalog replaces the Collaboration duplicate with identical behavior. |
| CI-003 | Centralize Jobs and Network Flow intent persistence | WF-05, WF-06 | BLOCKED | CI-001, RB-001 | SL-02 | No direct source-owner SQL to Collaboration tables; atomic behavior is preserved. |
| CI-004 | Narrow Collaboration Store capabilities | WF-05, WF-06 | BLOCKED | CI-002, CI-003, RB-001 | SL-03 | Callers depend on minimal ports, temporary compatibility is removed, and tests remain green. |
| CI-005 | Separate semantic Collaboration protocol from platform WebSocket mechanics | WF-05, WF-06 | BLOCKED | CI-001, RB-001 | SL-04 | `WSB-001` through `WSB-004` and `DOC-001` pass with one codec, one Hub, and byte/behavior parity. |
| CI-006 | Remove obsolete target placeholder | WF-06 | BLOCKED | CI-004, RB-001 | SL-05 | `.gitkeep` is removed with no tooling dependency. |
| CI-007 | Update authored verification accounting and generated projections | WF-07 | TODO | Any changed/new test identity | SL-06 | Exact rows route once and drift is clean. |
| CI-008 | Run final validation and implementation handoff | WF-08 | TODO | Implemented slices, CI-007 when applicable | SL-07 | Commands/results/run roots and remaining blockers are current. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Scope and owner posture complete; implementation remains unauthorized | Inspected `AGENTS.md`, framework, Core 00-05 applicability, domain and adopted subsystem NLSpecs; touched only this tracker | `sed`, `rg`, `git status --short --branch`, `git rev-parse HEAD`, `date --iso-8601=seconds` | Target and authority confirmed; no owner contradiction | RB-001 | Seek later authorization for a selected implementation slice. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Normative posture is explicit and subordinate to adopted owners; RB-002 and RB-003 are resolved design decisions | Inspected `docs/research/nlspec-spec.md`, `temp/analysis-notes.md`, owner passages, framework, development guide, and this tracker; touched only this tracker | `sed`, `rg`, `jq`, `git diff --check -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git diff --check HEAD -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git status --short --branch`, `date --iso-8601=seconds` | Supporting material was incorporated without elevating it to behavioral authority; both diff checks passed; no owner contradiction was found | RB-001 only | Obtain later authorization for SL-00 before any implementation. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Mixed-responsibility but legitimate Collaboration module diagnosed | Inspected all 19 target files, server/operator/timeline/revision assembly, Revisions, source-owner callers, platform WebSocket, Jobs, Network Flow, migrations, and boundary inputs; touched only this tracker | `find`, `rg`, `sed`, `jq` | Source mapping, direct SQL, Store breadth, and semantic/platform seams classified | RB-002, RB-003 | Begin SL-00 only under later authorization. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Catalog injection and semantic/transport boundaries are decision-complete | Inspected live `ProviderContribution`, `Appender`, `revisionassembly`, projection descriptors, generated view-schema registry, platform WebSocket, Collaboration routes, server runtime, HTTP dependency set, and Hub consumers; touched only this tracker | `sed`, `rg`, `jq` | One immutable 17-route catalog, one shared appender, one semantic codec, one Hub, and one server adapter are now mandatory | RB-001 only | Execute SL-00, then implement SL-01 and SL-04 independently under authorization. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Frontend Collaboration session and Workbook effects confirmed as consumers/owners of client state | Inspected Collaboration session, browser socket transport, Workbook collaboration message/effect/controller files, frontend boundary inputs, and selected tests; touched only this tracker | `rg`, `sed`, `jq` | No backend/grid-vendor coupling found; pending queue remains frontend-owned | None for current plan | Preserve frontend behavior; run frontend owner/browser validation only when affected. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Frontend remains an unchanged contract consumer of the proposed backend seam migration | Reused live frontend consumer and boundary evidence; inspected no frontend edit source; touched only this tracker | `rg`, `sed` | Pending queues, workbook effects, selectors, reconnect/dedupe, and grid adaptation remain frontend-owned and unchanged | RB-001 only for later implementation | Run web owner and browser parity tests during SL-04; do not move client state. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Authored WS and downstream generated surfaces mapped; no contract edit planned | Inspected WS index, protocol frontend entrypoints, protocol generator references, generated policy, migrations, and schema ownership manifest; touched only this tracker | `rg`, `sed`, `jq`, `make explain-target TARGET=generate-drift DETAIL=summary` | Internal refactor can preserve current schemas; generated roots remain generator-owned | Any future public contract change requires separate authorization | Keep authored contracts unchanged for the proposed behavior-preserving slices. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Existing projection-provider and view-schema projections are sufficient; no new manifest or WS contract change is planned | Inspected projection-provider manifest, view-schema index and canonical filters, generated registry, and current producer catalog; touched only this tracker | `jq`, `rg`, `sed` | All 17 current routes are mapped; generated artifacts remain downstream and hand edits are prohibited | RB-001 only | Validate parity through existing projections in SL-00/SL-01 and require clean generated drift. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Collaboration has 28 active owner rows, 19 service-backed; focused unit accounting gaps recorded | Inspected verification owner, module/web test-family catalogs, test-support inventory, target tests, and referenced consumer tests; touched only this tracker | `make help`, `make help-all`, `make task-guide ROLE=module-author OWNER=module.collaboration`, `make explain-test-owner OWNER=module.collaboration`, `make task-guide ROLE=module-author OWNER=web.collaboration`, `make explain-target` for focused targets, plus `jq`/`rg` | Canonical validation ladder discovered; no tests or product validation run | Later characterization tests need authored row decisions | Under later authorization, implement SL-00 and update authored routing before regeneration. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Characterization requirements now trace to `CAT-001` through `CAT-004` and `WSB-001` through `WSB-004` | Inspected exact tests, constructor sites, source-owner callers, third-party decoder defaults, and existing harness rows; touched only this tracker | `rg`, `sed`; no Make test target run | Required golden, construction-guard, catalog-failure, rollback, restore, Origin, close, frontend, and browser evidence is explicit | RB-001 only | Add and route exact SL-00 tests before either seam moves. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Session, membership, Origin, heartbeat, revocation, and closed-incident behavior frozen | Inspected Core 04, Collaboration routes/notifier, Auth/Incidents callers, socket/integration tests, and platform presence/revocation tests; touched only this tracker | `sed`, `rg` | Current authorization is rederived and incident-scoped; no contradiction found | None for current plan | Treat authorization parity tests as mandatory for route, Hub, or notifier movement. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | Origin precedence, frame limits, authorization rechecks, heartbeat, revocation, closure, and slow-consumer failures have exact preserved defaults | Inspected Core 01/Core 04, Collaboration route order, platform acceptor, `coder/websocket` v1.8.14 read/JSON behavior, and relevant tests; touched only this tracker | `rg`, `sed` against repository and local pinned module source | Pre-upgrade Origin gate, 32,768-byte limit, accepted frame kinds, and close mappings are explicit; no new security semantics were invented | RB-001 only | Prove every `WSB-003` case during SL-00 and SL-04. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T21:47:39-04:00 | Codex planning/tracker session | Planning tracker complete; implementation intentionally not started | Inspected repository state and this tracker; touched only this tracker | `git status --short --branch`, inventory/search and Make discovery commands; final read-only diff checks planned | No production refactor, test edit, contract edit, generated edit, migration, or harness edit | RB-001, RB-002, RB-003 | Authorize SL-00 first, then close RB-002 before SL-01; leave SL-04 deferred until RB-003 closes. |
| 2026-07-27T22:28:44-04:00 | Codex NLSpec tracker revision | RB-002 and RB-003 are closed as architecture decisions; implementation remains intentionally unstarted | Inspected supporting analysis, live types/defaults, the complete tracker diff, and final status; touched only this tracker | `git diff --check -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git diff --check HEAD -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git diff --stat HEAD -- docs/handoffs/collaboration-module-refactor-tracker.md`, `git status --short`; no product validation run | Both diff checks passed; status reported only the existing staged tracker addition plus its unstaged revision; no production, test, contract, generated, migration, package, or harness change occurred | RB-001 only | Authorize and execute SL-00; then implement the resolved catalog cutover and raw-socket split as independent slices. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Implementation is outside this planning-only authorization. | Production, test, contract, migration, generated, and harness changes are forbidden in this task. | A later task explicitly authorizing selected slices and their verification scope | BLOCKED |
| RB-002 | `RESOLVED`: `revisionassembly` collects data-only record/view routes from the existing source-owner `revisions.ProviderContribution` values, validates the exact current-profile catalog against projection descriptors and the generated view-schema registry, constructs one immutable `RecordViewCatalog`, and constructs one scoped `*revisions.Appender` that is injected into every source-owner service in that composition. Revisions owns validation and coordination; source owners own their contributions; Collaboration remains the sole durable intent writer. Package-global registration, per-module appender construction, fallback catalogs, shadow writes, and dual writers are prohibited. | The decision removes the duplicated Collaboration catalog without an import cycle or a second event owner. | SL-00 characterization; `CAT-001` through `CAT-004`; complete constructor inventory; catalog parity; rollback, historical-suppression, byte-parity, single-intent, and boundary-guard evidence | RESOLVED — implementation pending RB-001 and SL-00 |
| RB-003 | `RESOLVED`: Collaboration owns the exact raw-socket consumer port in Section 4, the sole semantic DTO/codec family, Hub, presence, hello/resume, application heartbeat, replay/live delivery, authorization choreography, terminal behavior, and semantic telemetry. Platform owns only configured upgrade/Origin mechanics, frame I/O, the explicit read limit, write serialization, low-level control frames, and close mechanics. Server supplies the pre-upgrade Origin gate and concrete adapter. Semantic types move rather than copy; no platform alias, duplicate codec, second Hub, or generated-contract change may remain. | The decision restores lower-level platform direction while preserving public bytes, frame behavior, replay, authorization, and frontend semantics. | SL-00 characterization; `WSB-001` through `WSB-004`; golden live/replay bytes; Origin, limits, heartbeat, presence, revocation, closure, boundary, frontend, browser, and generated-drift evidence | RESOLVED — implementation pending RB-001 and SL-00 |

RB-001 is the only open blocker. RB-002 and RB-003 are retained by stable ID as
closed design decisions and MUST NOT be reintroduced as implementation
discretion. No owner contradiction is currently known.

## 12. Binary Completion Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/collaboration` is inventoried or explicitly out of scope. | PASS | Section 2 contains 19 rows and states that none is omitted. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 covers route, session, presence, replay, events, views, storage, authorization, telemetry, frontend, generated protocol, and harness accounting. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 defines WF-00 through WF-08 with predecessors, successors, validation, and checkpoints. |
| Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | PASS | Section 7 marks every slice as requiring later authorization and permits no public behavior/schema change. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 records current Make-owned commands and applicability. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No contradiction was found; Sections 1, 5, and 11 define the required fail-closed posture if one appears. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Sections 3 and 5 record the framework’s narrower catalog versus the live target. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Section 10 records scope, backend/frontend, contract/codegen, harness, security, risks, commands, and next actions. |
| This planning task changes only the tracker and performs no refactor. | PASS | Scope posture, Section 10 session log, passing diff checks, and `git status --short` reporting only this tracker. |
| Normative tracker language is subordinate to adopted owner documents and does not create public behavior. | PASS | Section 1 states the authority boundary and fail-closed contradiction rule. |
| RB-002 defines one owner graph, one complete immutable catalog, one composition-scoped appender, and one durable intent writer. | PASS | Sections 3.1, 4.1, 4.2, 4.5, 7, and 11 define `CAT-001` through `CAT-004` without an unresolved design choice. |
| RB-003 defines one owner graph, one raw-socket port, one codec, one Hub, exact defaults, and exact failure mapping. | PASS | Sections 3.2, 4.3, 4.4, 4.5, 7, and 11 define `WSB-001` through `WSB-004` without an unresolved design choice. |
| Existing machine projections are sufficient for catalog validation and no parallel authority is proposed. | PASS | Section 4.1 names the projection-provider and generated view-schema evidence and prohibits a new manifest. |
| Every normative catalog requirement has a binary acceptance condition and required evidence. | PASS | `CAT-001` through `CAT-004` cover completeness, construction, transaction behavior, cutover, and storage boundaries. |
| Every normative WebSocket-boundary requirement has a binary acceptance condition and required evidence. | PASS | `WSB-001` through `WSB-004` cover imports/ownership, wire bytes, security/lifecycle behavior, and generated drift. |
| Core owner changes are not prerequisites and the development-guide follow-up is explicit. | PASS | Sections 1, 4.5 `DOC-001`, and SL-04 distinguish owner authority from implementation-support maintenance. |
| The top-level tracker contains no duplicate ID and uses only the allowed status vocabulary. | PASS | Section 9 has unique `CT-*` and `CI-*` rows using `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, or `DROPPED`. |
| RB-001 is the only open blocker; RB-002 and RB-003 are closed and removed from slice/work-item dependencies. | PASS | Sections 7, 9, and 11. |

The tracker is complete only while every `PASS` statement above remains true.
Any later repository change that invalidates a statement MUST change it to
`FAIL` or `BLOCKED`, identify the exact evidence gap, and prevent a completion
claim until the gap is closed.
