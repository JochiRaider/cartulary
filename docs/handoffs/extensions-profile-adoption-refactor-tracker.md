# Extensions Profile Adoption Refactor Tracker and Handoff

## 1. Scope, Authority, Assumptions, and Non-Goals

| Field | Recorded value |
| --- | --- |
| Output path | `docs/handoffs/extensions-profile-adoption-refactor-tracker.md` |
| Planning baseline | Clean `main` at `c0c9c02aae53762d25316afaa8808aa5e424b0db`, observed on 2026-07-24 |
| Upstream posture | `main` is one commit ahead of `origin/main` (`b796a9c5b80b43e7d8284cb1e96de971364e169b`) |
| Task posture | Ordered remediation execution under the mandatory checkpoint protocol |
| Execution status | `EPA-00` through `EPA-04` are `DONE`; `EPA-05` through `EPA-12` are `NOT_STARTED` |
| Controlling authority | Core 00 through Core 04, adopted profile-owner specifications, the Extensions NLSpec, `docs/domain.md`, and the Harness NLSpec |
| Historical evidence | `docs/handoffs/extensions-module-refactor-tracker.md` and `docs/handoffs/extensions-subsystem-implementation-tracker.md`; neither is normative |
| Domain posture | `domain vocabulary unchanged` |

Core 00 recognizes five canonical in-scope profile IDs:

| Requested audit domain | Canonical profile ID | Contract major | Claimable | Primary normative owner |
| --- | --- | ---: | --- | --- |
| Enterprise Authentication Extension Profile | `enterprise_authentication` | 1 | true | Core 01, with Core 04 security/configuration |
| Import Extension Profile | `import` | 1 | true | Core 01, with Core 03 interaction and Core 04 conformance |
| Incident Portability Extension Profile | `incident_portability` | 1 | true | Core 01 |
| Reference Pack Extension Profile | `reference_pack` | 1 | true | Core 01, with Core 04 security/configuration |
| Snapshot capability domain | `snapshot_reporting` | 1 | true | Reporting owner plus Core 01/Core 04 |
| Reporting capability domain | `snapshot_reporting` | 1 | true | Reporting and Report Composition owners plus Core 01/Core 04 |

Snapshot and Reporting are two audit domains under one profile. No
`snapshot`, `reporting`, or other replacement profile ID may be introduced.

`network_flow_activity` is recognized at contract major 2 and depends on
`import@1`. It is a reference and regression target only. Its adopted owner
explicitly declares no profile-owned job kind and no profile-owned worker kind
because Import owns its scheduling. A shared publication, Import, portability,
or recovery correction may require Network Flow adapter or regression changes;
that does not turn Network Flow into a profile-specific remediation target.

### Assumptions

- Public profile IDs, contract majors, route families, schemas, configuration
  keys, and common job response shapes remain compatible.
- Core-managed profile state remains part of the Core database and base backup.
  It does not acquire Extensions state metadata, migrations, state locks, or
  backup codecs merely because the profile uses durable Core state.
- Report Composition consumes no generic Extensions transaction, portability,
  backup/restore, state-presence, or Snapshot/Reporting participant interface
  under REQ-RC-076a.
- A report-builder UI is optional under REQ-RC-079. Absence is valid when
  builder-UI conformance is not claimed.
- Generated artifacts remain downstream. Authored owner inputs change first,
  followed by `make generate`.
- The profile-adoption job format is a clean pre-release cutover. Repository
  history has no release tags, the package version is `0.0.0`, and migration
  `00034` was introduced only in the current local baseline. Databases
  containing jobs from the pre-cutover profile handlers are unsupported and
  must be reset or reseeded; no legacy reader, backfill, or synthetic proof is
  permitted.

### Explicit non-goals

- Implementing any remediation in this planning task.
- Editing the completed Extensions trackers.
- Promoting the draft Reference Pack NLSpec or treating it as current authority.
- Adding a feature flag, compatibility alias, fallback reader, dual publication
  architecture, or second profile identity.
- Preserving jobs created by the pre-cutover handler names.
- Moving feature behavior into `cmd/*`.
- Reclassifying Core-managed state as extension-versioned state.
- Adding Network Flow jobs or workers that its adopted owner forbids.
- Making Core 05 evidence claims.
- Editing `docs/domain.md` without a genuine vocabulary or concept-boundary
  change.

## 2. Current Source Baseline and Repository Status

| Check | Result |
| --- | --- |
| `git status --short --branch` | `main...origin/main [ahead 1]`; the controlling tracker is the expected staged task input and carries the active EPA-00 checkpoint edit |
| `git log -1` | `c0c9c02a Extensions Module End-to-End Remediation` |
| Target tracker before this task | Absent |
| Existing completed tracker | Present and unchanged |
| Generated roots | Unchanged |
| Product sources, tests, contracts, and Harness inputs | Unchanged by this task |

The prior completed tracker says that all publication consumers use one typed
epoch and that no fallback path remains. Live source does not fully support
those historical claims:

- `internal/app/server/runtime.go` derives claims with a profile-ID switch,
  registers profile routes from a fixed application list, and acknowledges
  listeners before publication commit.
- `internal/platform/httpapi/extensions.go` retains generated-registry bytes,
  hard-coded default claim values, nil fallbacks, and global testing mutation.
- route services and workbook startup consume a broad discovery-shaped
  `ExtensionEpochProvider`, not their exact route or workspace plan projection.
- the generated worker plan is empty even though four in-scope profiles recover
  durable handlers.

These are new live-source findings, not grounds to rewrite history.

## 3. Documents and Evidence Inspected

### Normative owners

| Document | Status and use |
| --- | --- |
| `docs/spec/00_document_set_status_and_precedence.md` | Recognition, canonical IDs, claimability, contract majors, primary ownership, and precedence |
| `docs/spec/01_architecture_storage_and_view_contracts.md` | discovery, route contracts, common jobs, Import, snapshots/releases, Reference Packs, Incident Portability, staged objects, transactions, backup/restore |
| `docs/spec/02_domain_model_schema_and_history.md` | state authority, history, and cross-owner domain constraints |
| `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Import Assistant and extension workspace/client behavior |
| `docs/spec/04_security_deployment_and_conformance.md` | configuration, authorization, lifecycle, readiness, egress, and acceptance evidence |
| `docs/extension-subsystem-nlspec.md` | generic registry, claims, publication, state, participants, jobs, reconciliation, and conformance coordination |
| `docs/reporting-subsystem-nlspec.md` | adopted `snapshot_reporting` render/export semantics and `snapshot_reporting.render_export_v1` |
| `docs/report-composition-nlspec.md` | adopted composition authoring, validation, preview, and optional builder-UI boundary |
| `docs/opentelemetry-instrumentation-nlspec.md` | resolved-claim telemetry projection |
| `docs/testing-harness-nlspec.md` | command, selector, scheduling, fixture, artifact, and evidence accounting |
| `docs/domain.md` | vocabulary and concept boundaries |

### Reference and implementation-support evidence

| Document | Posture |
| --- | --- |
| `docs/network-flow-activity-nlspec.md` | adopted reference profile; remediation target only for shared regression |
| `docs/reference-pack-subsystem-nlspec.md` | draft 0.1.0; support evidence only |
| `docs/handoffs/extensions-module-refactor-tracker.md` | completed historical implementation evidence |
| `docs/handoffs/extensions-subsystem-implementation-tracker.md` | completed historical adoption evidence |
| `docs/guides/cartulary-dev-guide.md` | implementation support; its worker/plan wording overstates current wiring |
| `docs/guides/cartulary-ui-ux-design-guide.md` | design direction for deployment administration, Reference Packs, and Incident Import |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | frontend ownership and evidence mechanics |
| `docs/guides/cartulary_implementation_testing_guide.md` | repository implementation workflow |
| `docs/guides/cartulary_report_builder_implementation_stack_guide.md` | non-normative builder implementation support |
| `docs/guides/cartulary_browser_design_readiness_workflow.md` | browser inspection support, not conformance authority |

### Authored, generated, runtime, and evidence inputs

- every authored file under `contracts/extensions`;
- embedded generated Extensions artifacts in
  `internal/gen/contracts/contracts_gen.go`;
- generated TypeScript contract consumers under
  `packages/protocol-ts/src/generated` and
  `packages/ui-contracts/src/generated`;
- `contracts/openapi/cartulary.openapi.yaml`;
- profile migrations and reporting queries under `db/migrations` and
  `db/queries`;
- all files in the profile implementation roots inventoried in Section 4;
- application composition, Extensions coordination, platform HTTP, config,
  jobs, telemetry, extension storage, recovery, staged objects, transactions,
  revisions, workbook startup, frontend consumers, and browser evidence;
- verification owners in `contracts/verification/owners`;
- owner manifests in `tools/test_families`;
- command output from `make help`, `make help-all`, `make task-guide`, and
  `make explain-test-owner`.

## 4. Exhaustive Profile File and Caller Inventory

The inventory is exhaustive for profile-specific code and its direct
composition, generated, storage, frontend, and evidence callers at the recorded
baseline. Unrelated Base behavior in a shared package is not listed as profile
code.

### 4.1 `enterprise_authentication`

| Layer | Exact files or callers |
| --- | --- |
| Owner inputs | `contracts/extensions/fragments/core00.recognition.json`; `core01.discovery.json`; `core01.profile-classifications.json`; `core04.claim-configuration.json`; `profiles/enterprise_authentication/configuration.json`; `build/implementation-bindings.json` |
| Profile implementation | `internal/modules/auth/enterprise_errors.go`; `enterprise_protocol.go`; `enterprise_protocol_helpers.go`; `enterprise_routes.go`; shared registration in `routes.go`; binding handling in `user_admin_handlers.go`; session mapping in `auth_session_helpers.go`; user/provider persistence in `users.go` and `internal/platform/authn/enterprise_store.go` |
| Platform adapter | `internal/platform/enterpriseauth/enterpriseauth.go`; `manifest.go`; config in `internal/platform/config/config.go` and `validation.go` |
| Application composition | `internal/app/server/runtime.go` loads provider definitions and reconciles them from raw `cfg.EnterpriseAuthentication.Claimed` |
| HTTP/profile callers | `auth.RegisterRoutes`; `httpapi.ExtensionProfileClaimedIn`; discovery and reserved-family middleware |
| Frontend | `apps/web/src/app/AuthGateway.tsx`; `AccountAdministrationPanels.tsx`; `App.tsx`; `app/api/appShellClient.ts`; `DeploymentAuditPanel.tsx` |
| Tests | enterprise auth contract/integration tests; platform manifest/verifier tests; `apps/web/e2e/enterprise-auth.spec.ts`; auth frontend tests; incident support probes |
| Harness | `module.auth`; `platform.config`; app-server publication/process evidence |

The profile ID is duplicated as a literal in auth routes, frontend availability,
tests, runtime claim projection, and the platform HTTP fallback. Profile-owner
code may retain its own canonical constant; shared generic and composition code
must use generated identity and exact plan rows.

### 4.2 `import`

| Layer | Exact files or callers |
| --- | --- |
| Owner inputs | Core recognition/discovery/classification/claim fragments; `profiles/import/configuration.json`; implementation binding |
| Profile implementation | all 14 files under `internal/modules/imports`, including `api.go`, `routes.go`, `store.go`, `source_streams.go`, `xlsx.go`, `targets.go`, `extension_facade.go`, `owner_apply.go`, and `ownerfacade/*` |
| Source-owner callers | owner create/apply facades in artifacts, assessments, evidence, host identity, indicators, parties, tasks/decisions, timeline, and other importable Core owners |
| Extension caller | `internal/modules/networkflow/module.go`; `import_facade.go`; `transaction_participants.go`; Network Flow target declarations in `imports/targets.go` |
| Application composition | import store and routes plus Network Flow facade override in `internal/app/server/runtime.go` |
| Storage | `db/migrations/00018_imports.sql`; `00028_import_source_streams_and_targets.sql`; common jobs |
| Frontend | `apps/web/src/shared/importCoordinator.ts`; debug harness; Network Flow mapping/import controller; no production general Workbook Import Assistant |
| Tests | three service-backed Import selectors plus Network Flow import and shared frontend helper tests |
| Harness | `module.imports`; relevant `module.networkflow`, `web.networkflow`, and missing general Import browser ownership |

Current handler identities are `imports.discovery` and `imports.apply`.
Neither maps to a declared `import.*` job-kind contract or worker fact.

### 4.3 `incident_portability`

| Layer | Exact files or callers |
| --- | --- |
| Owner inputs | Core recognition/discovery/classification/claim fragments; `profiles/incident_portability/configuration.json`; portability shared protocols and result contracts; implementation binding |
| Profile implementation | all 14 files under `internal/modules/incidentbundles`; `internal/modules/incidentportability/import_targets.go` and `portability.go` |
| Source participants | incident bundle portability providers in artifacts, assessments, entities, evidence, incidents, indicators, links, parties, records, revisions, saved views, tasks/decisions, and timeline |
| Shared coordinators | `internal/modules/crossownertransaction`; `internal/modules/stagedobjects`; Extensions portability catalog; revisions attribution registry |
| Application composition | `internal/app/extensionassembly/incident_portability.go`; runtime portability policy, state-presence, staged-object, and transaction assembly |
| Storage | `db/migrations/00023_incident_bundles.sql`; common jobs; staged-object and proof tables from migration `00034` |
| Frontend | `apps/web/src/app/IncidentImportPanel.tsx`; `App.tsx`; support E2E |
| Tests | Incident Bundle unit/integration rows, Network Flow retained-state blocking, staged-object and transaction owner evidence |
| Harness | `module.incidentbundles`, `module.crossownertransaction`, `module.stagedobjects`, and app-server composition |

The production handler is `incident_bundles.execute`; the profile payload selects
export or import. No `incident_portability.*` job-kind or worker fact exists.

### 4.4 `reference_pack`

| Layer | Exact files or callers |
| --- | --- |
| Owner inputs | Core recognition/discovery/classification/claim fragments; `profiles/reference_pack/configuration.json`; implementation binding |
| Profile implementation | all 11 files under `internal/modules/reference_data`, including routes, store, verifier, minimum disconnected bundle, and worker hooks |
| Application composition | fixed route contribution in `internal/app/server/runtime.go`; disconnected seeding from current profile rows |
| Storage | `db/migrations/00019_reference_data.sql`; reference entries in Incident Bundle storage; common jobs |
| Frontend | `ReferencePackAdminPanel.tsx`; admin client/model; landing admin layout and app availability |
| Tests | Reference Pack unit/integration, frontend panel, webserver-backed browser, bootstrap/object-store/config support |
| Harness | `module.reference_data`; `web.application`; related workbook/view-contract rows |

The production handler is `reference_data.execute`; payload kinds include
`import`, `reverify`, and `refresh`. None has a `reference_pack.*` job contract,
worker fact, or common-job ownership metadata.

### 4.5 `snapshot_reporting`: Snapshot capability domain

| Layer | Exact files or callers |
| --- | --- |
| Owner inputs | Core fragments; `snapshot_reporting.participation.json`; configuration and participant contracts; Reporting owner manifest; implementation binding |
| Profile implementation | reporting `api.go`, `application_service.go`, `routes.go`, `store.go`, `reporting_job_worker.go`, export materializer, redaction, render bundle, and provider ports |
| Application composition | reporting registrar in `internal/app/server/runtime.go` |
| Storage | `db/migrations/00021_reporting.sql`; `db/queries/reporting.sql`; common jobs |
| HTTP callers | `/api/v1/snapshots` and `/api/v1/snapshots/*`; common job status routes |
| Frontend | no current production snapshot client; Core 03 permits omission |
| Tests | reporting unit/integration tests for snapshot replay, boundary stability, exact shapes, and redaction |
| Harness | `module.reporting`; app-server composition; no browser requirement solely from current owner |

The reporting handler is `reporting.execute`; payload kind
`snapshot_create` is not declared as an extension job kind.

### 4.6 `snapshot_reporting`: Reporting capability domain

| Layer | Exact files or callers |
| --- | --- |
| Owner inputs | Reporting and Report Composition owner manifests; `snapshot_reporting.participation.json`; `reporting-participant.json`; validation surfaces; implementation binding |
| Reporting implementation | all 18 files under `internal/modules/reporting` |
| Composition implementation | all 8 files under `internal/modules/reportcomposition` |
| Source provider callers | reporting providers under incidents, artifacts, evidence, entities, links, parties, records, tasks/decisions, and timeline |
| Application composition | reporting and report-composition registrars in `internal/app/server/runtime.go` |
| Storage | migrations `00021_reporting.sql` and `00022_report_composition.sql`; reporting query input; common jobs |
| HTTP callers | `/api/v1/releases`; `/api/v1/incidents/{incident_id}/report-compositions`; preview creates a common job |
| Frontend | no production report builder; optional under REQ-RC-079 |
| Tests | reporting unit/integration; Report Composition traceability and release-tuple unit tests |
| Harness | `module.reporting` has 10 rows/4 service-backed; `module.reportcomposition` has one unit row and no service-backed row |

`snapshot_reporting.render_export_v1` occurs in adopted documents and authored
contracts, generated bindings, validation inputs, and tests of contract
accounting. No runtime implementation or invocation was found. Composition
preview creates a job with no handler name and no reporting dispatch.

## 5. Profile-by-Capability Adoption Matrix

Every cell uses one required disposition verbatim.

| Capability layer | `enterprise_authentication` | `import` | `incident_portability` | `reference_pack` | `snapshot_reporting` / Snapshot | `snapshot_reporting` / Reporting |
| --- | --- | --- | --- | --- | --- | --- |
| Normative declaration and ownership | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven |
| Canonical ID, major, claimability, compatibility | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven |
| Runtime profile dependencies | declared not applicable | conformant and proven | declared not applicable | declared not applicable | declared not applicable | declared not applicable |
| Authored manifests/fragments/configuration | conformant and proven | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap |
| Generated registry and integrity | conformant and proven | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap |
| Coordinator admission and claims | conformant but insufficiently tested | conformant but insufficiently tested | conformant but insufficiently tested | conformant but insufficiently tested | conformant but insufficiently tested | conformant but insufficiently tested |
| Dependency order and collision validation | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven |
| Immutable publication-plan projection | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap |
| Application-owned publication epoch | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap |
| Discovery visibility and ordering | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven |
| Reserved route dispatch | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap |
| HTTP route preparation | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap |
| WebSocket contribution | declared not applicable | declared not applicable | declared not applicable | declared not applicable | declared not applicable | declared not applicable |
| Workspace contribution | declared not applicable | declared not applicable | declared not applicable | declared not applicable | declared not applicable | declared not applicable |
| Worker/job preparation | declared not applicable | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap |
| Exact acknowledgments and serving gate | implementation gap | implementation gap | implementation gap | implementation gap | implementation gap | implementation gap |
| Client support/frontend availability | conformant and proven | implementation gap | conformant and proven | conformant and proven | declared not applicable | declared not applicable |
| Configuration and inactive policy | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap | composition/wiring gap |
| Extension-versioned state | declared not applicable | declared not applicable | declared not applicable | declared not applicable | declared not applicable | declared not applicable |
| Core-managed state | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant but insufficiently tested | conformant but insufficiently tested |
| Cross-owner transaction | declared not applicable | conformant and proven | conformant and proven | declared not applicable | declared not applicable | declared not applicable |
| Staged-object publication | declared not applicable | declared not applicable | conformant and proven | declared not applicable | declared not applicable | declared not applicable |
| Incident portability participation | declared not applicable | declared not applicable | conformant and proven | declared not applicable | declared not applicable | declared not applicable |
| Backup/restore extension binding | declared not applicable | declared not applicable | declared not applicable | declared not applicable | declared not applicable | declared not applicable |
| Snapshot/Reporting participant | declared not applicable | declared not applicable | declared not applicable | declared not applicable | implementation gap | implementation gap |
| Telemetry projection | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven | conformant and proven |
| Revision-claim projection | declared not applicable | declared not applicable | conformant and proven | declared not applicable | declared not applicable | declared not applicable |
| Implementation binding/conformance | conformant but insufficiently tested | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap | authored-contract or generation gap |
| Fatal lifecycle/readiness | implementation gap | implementation gap | implementation gap | implementation gap | implementation gap | implementation gap |
| Harness ownership/selectors | test or Harness-accounting gap | test or Harness-accounting gap | test or Harness-accounting gap | test or Harness-accounting gap | test or Harness-accounting gap | test or Harness-accounting gap |
| Developer documentation | documentation gap | documentation gap | documentation gap | documentation gap | documentation gap | documentation gap |

The core-managed rows do not absolve a profile from common job ownership. A
durable job is a separate Extensions obligation even when its resource tables
remain Core managed.

## 6. Profile-by-Runtime-Consumer Matrix

| Runtime consumer | Enterprise Authentication | Import | Incident Portability | Reference Pack | Snapshot | Reporting |
| --- | --- | --- | --- | --- | --- | --- |
| Claim configuration | raw config switch plus generated descriptor | raw config switch plus descriptor | raw config switch plus descriptor | raw config switch plus descriptor | raw config switch plus descriptor | same profile claim |
| State admission | no extension state plan | no extension state plan | no extension state plan | no extension state plan | no extension state plan | no extension state plan |
| Discovery | publication controller discovery projection | same | same | same | same | same |
| Reserved HTTP matcher | broad epoch profiles | broad epoch profiles | broad epoch profiles | broad epoch profiles | broad epoch profiles | broad epoch profiles |
| Route construction | fixed `auth` registrar and claim check | fixed `imports` registrar | fixed `incident_bundles` registrar | fixed `reference_data` registrar | fixed `reporting` registrar | fixed `reporting` plus `report_composition` registrars |
| Exact route-plan rows | not consumed | not consumed | not consumed | not consumed | not consumed | not consumed |
| Workspace registry | not applicable | no registered workspace; missing assistant UI | not applicable | not applicable | not applicable | not applicable |
| Worker plan | empty | empty despite two handlers | empty despite one handler | empty despite one handler | empty despite reporting handler | empty despite reporting/preview work |
| Job recovery | none | invoked during route preparation | invoked during route preparation | invoked during route preparation | invoked during route preparation | reporting recovery; preview has no handler |
| Listener/worker acknowledgment | listener-only, before commit | listener ack before handler registration | listener ack before handler registration | listener ack before handler registration | listener ack before handler registration | listener ack before handler registration |
| Client availability | discovery-gated Auth UI | helper exists; production assistant absent | discovery-gated import panel | discovery- and role-gated admin panel | no required UI | optional builder absent |
| Telemetry | resolved claim set | resolved claim set | resolved claim set | resolved claim set | resolved claim set | resolved claim set |
| Revisions attribution | none | none | exact imported attribution resolver | none | none | none |
| Portability | explicit no incident state | explicit no incident state | shared orchestrator and NF blocking | explicit no incident state | explicit no incident state | Composition explicitly consumes none |
| Recovery codec | Core/base backup | Core/base backup | Core/base backup | Core/base backup | Core/base backup | Core/base backup |
| Conformance | generated manifest/binding | generated manifest omits live jobs | generated manifest omits live jobs | generated manifest omits live jobs | generated manifest omits live jobs | binding advertises unimplemented participant |

## 7. Authored-Input and Generated-Artifact Provenance

| Provenance stage | Exact owner/input | Downstream result | Current finding |
| --- | --- | --- | --- |
| Recognition | `fragments/core00.recognition.json` | descriptor ID, major, owner, claimability | correct for all six recognized profiles |
| Discovery | `fragments/core01.discovery.json` | route families and NF workspace | correct |
| Classification | `fragments/core01.profile-classifications.json`; `snapshot_reporting.participation.json` | state, egress, portability, reporting, contributions | missing job/worker facts; reporting participant declared |
| Claim configuration | `fragments/core04.claim-configuration.json`; `profiles/*/configuration.json` | claim key and inactive-key policy | authored correctly; runtime projection is hard-coded |
| Dependencies | `dependencies.json` and owner manifests | dependency snapshot and owner locator integrity | correct; only Network Flow has runtime profile dependency |
| Implementation assertions | `build/implementation-bindings.json` | per-profile binding and binding-set digest | falsely omits live workers/jobs and claims reporting participant implementation |
| Participant contract | `profiles/snapshot_reporting/reporting-participant.json` | participant registry and binding digest | uses unsupported profile-local context/result schema IDs |
| Client support | `build/client-support.json` | generated client support registry | correctly contains only NF workspace support |
| Generation | `tools/contractgen/extensions_*.go` | embedded generated artifacts | generator supports worker/job facts but receives none |
| Embedded backend | `internal/gen/contracts/contracts_gen.go` | coordinator artifact source | do not edit by hand |
| Generated frontend | protocol/UI generated roots | packaged decoders/types | do not edit by hand |
| Runtime admission | `internal/modules/extensions/coordinator.go` | descriptors, claims, plan, policies | validates authored/generated bytes but not actual handler existence |
| Application epoch | publication controller and runtime | process-local installed plan | typed plan exists; exact projections are not fully consumed |
| Harness | verification owners and `tools/test_families/*` | selector topology/evidence accounting | missing adoption-path selectors |

Generated outputs implicated by later work include descriptors, the registry,
implementation bindings and binding-set digest, owner inputs, dependency
snapshot, participant registry, job-kind artifacts, closure catalogs,
conformance manifests/index, registry accounting/integrity, generated schemas,
and generated Go/TypeScript embeddings. The exact set must come from
`make generate`; no workstream may predict it by editing generated paths.

## 8. Current Import and Composition DAG

```text
cmd/server
  -> internal/app/server
       -> internal/modules/extensions
       -> internal/app/extensionassembly
            -> extensions catalogs and semantic ports
            -> networkflow portability/state/recovery adapters
            -> incidentbundles / crossownertransaction / stagedobjects
       -> concrete profile route registrars
            -> auth
            -> imports -> ownerfacade implementations
            -> incidentbundles -> incident portability source owners
            -> reference_data
            -> reporting -> reporting/exportprovider source adapters
                         -> reportcomposition release tuple facade
            -> reportcomposition
       -> platform adapters
            -> config / httpapi / authn / jobs / postgres / object store / ws
```

Positive findings:

- `cmd/*` imports none of the profile or Extensions implementations directly.
- production `internal/modules/extensions` imports no concrete profile owner.
- concrete profile owners do not import the broad Extensions facade.
- application composition is already the legitimate dependency-inversion edge.
- platform `pgx`, HTTP, auth, and storage DTOs have not leaked through the
  extracted Extensions state, staged-object, portability, or transaction
  semantic ports.

Negative findings:

- app composition validates neither its fixed route catalog nor job handlers
  against exact generated contribution identities;
- route and workspace consumers down-convert the plan into broad discovery rows;
- the common job runner has no publication gate or profile/job contract;
- Reporting has no injected participant boundary;
- the platform HTTP compatibility path rereads registry bytes independently.

## 9. Per-Profile Findings

### Enterprise Authentication

- The four reserved route contributions match Core and authored discovery.
- Enterprise provider, binding, and browser evidence is substantial.
- Runtime uses raw claim configuration for manifest loading and reconciliation
  after the coordinator has already resolved claims. This creates a second claim
  decision.
- Routes use broad profile rows and a repeated literal profile ID.
- No job, worker, workspace, portability participant, reporting participant,
  state plan, transaction, staged object, or extra backup codec applies.
- Required correction: bind provider preparation and routes to the resolved
  epoch, retain owner-declared egress and existing public behavior, and prove
  every absence.

### Import

- The backend implements CSV/XLSX discovery, mapping, selection, apply, common
  jobs, owner facades, and Network Flow target dispatch.
- Its two durable handlers are undeclared and can recover before serving.
- Network Flow correctly declares that Import scheduling remains Import-owned.
- A shared frontend import coordinator exists, but production exposes it only
  through Network Flow/debug paths. The Core 03 Phase 2 Workbook Import
  Assistant and AC-027 browser behavior are missing.
- Required correction: job contracts, exact worker readiness, proof/metadata,
  claim-loss behavior, and a production Workbook Import Assistant.

### Incident Portability

- Route, staged-object, cross-owner final transaction, atomic import, attribution,
  and Network Flow blocking behavior are present.
- The profile itself correctly declares no authoritative incident extension
  state; it coordinates other owners.
- Export/import work uses one undeclared common-job handler.
- Required correction: exact export/import job kinds, publication gating,
  commit/cancellation proof integration, and production-composition evidence.

### Reference Pack

- Core-owned route, disconnected verification, activation, deployment-admin UI,
  and browser evidence exist.
- The draft Reference Pack NLSpec cannot widen current conformance.
- Import, reverify, and refresh jobs share one undeclared handler.
- Core-managed tables remain covered by base backup; no Extensions state or
  recovery codec applies.
- Required correction: job contracts/gating/proofs while preserving current
  disconnected and administrative behavior.

### Snapshot capability

- Snapshot routes and service-backed lifecycle evidence exist.
- Snapshot creation is a durable reporting job with no Extensions job or worker
  declaration.
- No snapshot UI is required by the current owner.
- Required correction: bind snapshot work to the common publication/job
  lifecycle without adding a client requirement.

### Reporting capability

- Release and composition route implementations exist.
- The adopted Reporting owner requires
  `snapshot_reporting.render_export_v1`; runtime never invokes it.
- The authored participant contract names
  `cartulary.snapshot_reporting_extension_context.v1` and
  `cartulary.snapshot_reporting_extension_result.v1`, neither of which has an
  adopted owner or generated schema.
- Reporting jobs bypass the participant and common Extensions job metadata.
- Composition preview creates an unhandled common job.
- No report builder is required unless builder conformance is claimed.
- Required correction: admitted typed participant invocation, reporting-owned
  result admission, complete snapshot/release/preview jobs, and production-path
  evidence.

## 10. Cross-Profile Systemic Findings and Gap Ledger

### Gap definitions

| ID | Profile/capability | Exact evidence | Classification | Required remediation | Affected areas | Workstream |
| --- | --- | --- | --- | --- | --- | --- |
| GAP-001 | Import, Incident Portability, Reference Pack, Snapshot/Reporting jobs | handler registration/recovery in profile routes; empty worker/job binding arrays | authored-contract or generation gap | adopt job/worker facts and exact implementation bindings | specification, contracts, generation, implementation, tests, Harness | EPA-02, EPA-04, EPA-06..09 |
| GAP-002 | common job ownership/proofs | migration `00034` columns/tables exist; `jobs.CreateParams` and inserts omit them; proof adapter unused | implementation gap | write metadata and atomic proofs/cancellation observations | implementation, storage, composition, tests | EPA-04 |
| GAP-003 | inactive jobs and pre-serving dequeue | no `reconcile_inactive_extension_jobs_v1`; route construction calls `RecoverHandler` before publication | implementation gap | generic reconciliation and one dequeue gate | implementation, composition, tests, Harness | EPA-04 |
| GAP-004 | all profiles, publication lifecycle | controller accepts acknowledgments only in `prepared`; commit requires all acks | implementation gap | commit plan first; ack only in `committed`; serve after all acks | implementation, composition, tests | EPA-03 |
| GAP-005 | all route/workspace/worker consumers | `PublicationPlan.Routes()`/`Workspaces()` unused; fixed registrars and broad epoch provider | composition/wiring gap | exact contribution catalog and typed epoch projections | implementation, composition, tests, docs | EPA-03 |
| GAP-006 | all claims; Enterprise preparation | `extensionClaimRequest` switch; raw enterprise claim checks | composition/wiring gap | generated config-owned claim projection and resolved-claim-only preparation | contracts if needed, config, composition, tests | EPA-03, EPA-05 |
| GAP-007 | shared HTTP path | registry-byte global defaults and nil fallback in `httpapi/extensions.go` | implementation gap | require explicit epoch; move test fixtures to testsupport; delete global mutation | implementation, tests, docs | EPA-11 |
| GAP-008 | Snapshot/Reporting participant | owner and binding claim `render_export_v1`; no runtime symbol/invocation | implementation gap | inject typed admitted participant invocation | contracts, implementation, composition, tests, Harness | EPA-02, EPA-09 |
| GAP-009 | Snapshot/Reporting participant schemas | profile-local context/result IDs occur only in participant JSON | authored-contract or generation gap | use exact adopted shared specialization schemas and regenerate digests | contracts, generation, tests | EPA-02 |
| GAP-010 | Report Composition preview | `jobs.CreateQueuedTx` has no handler; no later dispatch | implementation gap | create declared preview job and route it through Reporting worker/participant | implementation, storage, tests, Harness | EPA-09 |
| GAP-011 | Import client | backend exists; no production Workbook Import Assistant or browser selector | implementation gap | add claimed-profile assistant and stateful browser evidence | frontend, tests, Harness, docs | EPA-06 |
| GAP-012 | production-path evidence | profile tests prove features; app-server tests encode wrong ack order and no exact binding parity | test or Harness-accounting gap | add owner-correct production composition selectors | tests, Harness | EPA-01, EPA-10 |
| GAP-013 | implementation guidance | dev guide says workspace/worker admission binds to one plan | documentation gap | update after runtime and selectors are true | documentation | EPA-10 |
| GAP-014 | historical common jobs | no owner rule maps pre-adoption jobs/proofs into the new contracts | specification gap | adopt the clean pre-release cutover, reject databases containing pre-cutover handler jobs, and require reset/reseed without backfill | specification, contracts, migration, tests, docs | EPA-00, EPA-02, EPA-04 |

### Gap impact, dependencies, and validation

| ID | Rationale and long-term benefit | Compatibility/migration impact | Risk if unresolved | Dependencies | Validation criteria |
| --- | --- | --- | --- | --- | --- |
| GAP-001 | makes declared profiles match actual work and lets future profiles reuse one mechanism | internal digests change; public job shape does not | inactive jobs run without owner policy; bindings are false | owner amendment before generation | binding, descriptor, closure, and live handler parity |
| GAP-002 | establishes durable ownership and exact terminal evidence | complete unreleased migration `00034` in place; no compatibility migration | resource publication can be unprovable or misclassified | GAP-001, GAP-014 | atomic resource/job/proof commit and cancellation fixtures |
| GAP-003 | enforces fail-closed inactive behavior and one publication instant | pending jobs start later, after serving | inactive or pre-ready code performs work | GAP-001, GAP-002, GAP-004 | no dequeue before serve; deterministic inactive reconciliation |
| GAP-004 | conforms to EXT-REQ-214 and makes acknowledgment meaningful | internal startup ordering only | partial or false-ready epoch | EPA-01 characterization | ack-before-commit fails; commit/ack/serve succeeds |
| GAP-005 | prevents discovery-shaped down-conversion and future profile hard-coding | internal application assembly changes | registered code diverges from generated plan | GAP-004 | exact contribution parity and unclaimed absence |
| GAP-006 | removes second claim vocabulary | config keys unchanged | incompatible/inactive profile code may run | generated claim catalog | unknown future profile needs no runtime switch |
| GAP-007 | removes the last registry fallback and hidden defaults | test fixtures change; no public effect | split epoch and tests that bypass production | GAP-005 | no production global registry/default reader |
| GAP-008 | satisfies the adopted shared participant boundary | internal reporting pipeline change | conformance binding is knowingly false | GAP-009, GAP-004 | admitted call only, closed result, timeout/cancel/failure matrix |
| GAP-009 | binds code to owner-supported schemas | participant digest changes | generator blesses ownerless schema IDs | owner contract correction | generated participant registry matches owners |
| GAP-010 | makes preview terminate and remain Reporting-owned | pre-cutover pending previews are unsupported and rejected by the clean-cutover migration precondition | indefinite jobs and false preview behavior | GAP-001, GAP-002, GAP-008, GAP-014 | preview job reaches correct terminal state through production path |
| GAP-011 | closes current Import profile UI acceptance | additive UI only; public API already exists | claimed Import lacks required user workflow | GAP-005, GAP-003 | browser proves progress, cancel, apply, loss, Base fallback |
| GAP-012 | ensures evidence proves deployable composition | Harness manifests/topology regenerate | helper tests mask broken startup | all behavior slices | exact selectors pass under correct owners |
| GAP-013 | keeps contributors on the real architecture | documentation only | future code repeats obsolete patterns | implementation complete | guide matches source and Make surface |
| GAP-014 | avoids invented proofs and dual readers | development databases containing any pre-cutover profile-handler job must be reset or reseeded | upgrade can otherwise lose or fabricate job evidence | clean-cutover decision recorded in EPA-00 | fresh migration passes; every retired handler fixture fails before mutation with a reset-required diagnostic |

## 11. Target Ownership and Composition Model

```text
cmd/server
  -> internal/app/server
       -> internal/app/extensionassembly
            -> generated identity/configuration projection
            -> extensions.Coordinator
                 -> immutable descriptors, claims, policies, and plan
                 -> generic participant and job-reconciliation protocols
            -> profile contribution catalog
                 -> auth provider/routes
                 -> Import routes + discovery/apply workers
                 -> Incident Portability routes + export/import worker
                 -> Reference Pack routes + job worker
                 -> Reporting routes + snapshot/release/preview worker
                 -> Report Composition routes
                 -> Network Flow reference route/workspace/participants
            -> platform adapters
                 -> HTTP/listener gate
                 -> WebSocket/listener gate
                 -> job runner/dequeue gate
                 -> config, Postgres, object store, auth, telemetry
       -> PublicationController
            Prepare(plan and quiescent components)
            Commit(install immutable epoch, gate closed)
            Acknowledge(exact listeners and claimed workers)
            Serve(one atomic admission gate)
```

### Ownership rules

- `internal/modules/extensions` owns generated artifact admission, claim
  resolution, immutable policy and plan projection, generic validation,
  participant coordination, and inactive-job reconciliation semantics.
- Profile owners construct their route, worker, participant, and profile-specific
  proof behavior. They do not import the broad Extensions facade.
- `internal/app/extensionassembly` is the dependency-inversion edge. It imports
  generic Extensions contracts, concrete profile owners, and platform adapters,
  then builds an exact catalog keyed by generated contribution and worker
  identities.
- `internal/app/server` owns process lifecycle and one publication controller,
  not feature semantics.
- platform packages own TOML parsing, HTTP/auth DTOs, SQL/pgx mechanics, object
  storage, listener loops, and common job storage/runner mechanics.
- semantic ports expose logical profile IDs, job kinds, proofs, participant
  contexts/results, and commit outcomes; they do not expose `pgx.Tx`,
  `http.Request`, storage DTOs, or auth/session DTOs.
- the profile contribution catalog must be complete and exact. Extra and missing
  route, worker, participant, or binding implementations fail before commit.
- Base routes may remain an application-owned explicit list. Recognized
  extension route contributions may not be inferred from that list.
- profile code remains quiescent while prepared. No profile handler may recover
  or dequeue a job before the shared serving gate opens.
- the installed plan is process-local, immutable, unlogged, and unpersisted.

### Planned internal interfaces

Names are descriptive planning names; implementation may use repository naming
conventions without changing their ownership or behavior.

| Interface | Owner | Required behavior |
| --- | --- | --- |
| Claim configuration projection | platform config | copied sorted `{profile_id, claimed}` values derived from generated claim contracts; no runtime profile switch |
| Contribution catalog | application assembly | exact maps for route contribution ID, worker kind, and participant ID; validates against the plan |
| Publication epoch view | application server | copied typed routes, workspaces, workers, bindings, claims, discovery; no registry bytes |
| Prepared component | application assembly | prepare quiescent, acknowledge exact digest after commit, stop on failure/loss |
| Dequeue admission | platform jobs | one gate shared with HTTP, WebSocket, readiness, discovery, and workspace availability |
| Extension job definition | profile owner/common jobs | canonical owner profile, job kind, handler identity, proof/cancel policy, public result schema |
| Job reconciliation store | Extensions semantic port | ordered bounded read/classify/atomic-update operations with logical values only |
| Snapshot/Reporting participant invoker | Reporting-owned consumer port | invoke only the admitted specialization with closed context/result and deadline |

## 12. Execution Plan and Workstream Dependencies

```text
EPA-00
  -> EPA-01
       -> EPA-02
            -> EPA-03
                 -> EPA-05
                 -> EPA-04
                      -> EPA-06
                      -> EPA-07
                      -> EPA-08
                      -> EPA-09
                           -> EPA-10
                                -> EPA-11
                                     -> EPA-12
```

`EPA-06` through `EPA-09` may be implemented in any order after `EPA-04`, but
only one may be `IN_PROGRESS`. `EPA-05` depends on `EPA-03` rather than
`EPA-04`. `EPA-10` waits for all profile slices so it can assign final ownership
once. `EPA-11` waits for all replacement callers before deleting obsolete paths.

Intermediate commits are independently verifiable development checkpoints, not
independently deployable releases. No production rollout is permitted before
`EPA-12`.

## 13. Workstream Specifications

### EPA-00 — Baseline and controlling-tracker activation

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | none |
| Ownership surfaces | this tracker; read-only source/contract/Harness inspection |
| Intended changes | mark `EPA-00` `IN_PROGRESS`; reconfirm commit, worktree, owner digests, generated inputs, inventories, and command surface; update only checkpoint fields |
| Compatibility posture | no behavior or contract change |
| Risks | beginning remediation against a moved or dirty baseline |
| Narrow verification | `git status --short --branch`; exact `rg`/`jq` inventory checks; owner command discovery |
| Service-backed verification | none |
| Generated/drift checks | read-only digest comparison; no generation |
| Exit criteria | baseline differences are reconciled in this tracker; no remediation file changed; next slice is selected |
| Checkpoint | files inspected, commands/results, baseline delta, blockers, `domain vocabulary unchanged` |
| Completion checkpoint | Baseline remains `c0c9c02a`; the tracker is the only authored change. No release tags exist, package version is `0.0.0`, migration `00034` entered at `00522cfe`, and the authored Extensions input inventory digest is `af1c8e2ea6b232d9fd0a038ec3beb37e0a012ca56e6630434208772868b6f5bf`. `make task-guide ROLE=module-author OWNER=module.extensions` and `make explain-test-owner OWNER=module.extensions` confirmed 20 rows/9 service-backed. `git diff --check`, `make generated-artifact-policy-check`, `make json-shape-check`, and `make lint-markdown` passed. Run roots: `.cartulary/test-results/20260724T181025Z-p71088` and `.cartulary/test-results/20260724T181028Z-p71438`. BLOCK-001 is resolved by clean cutover; BLOCK-002 remains for EPA-02. Rollback is tracker-only. Domain vocabulary unchanged. |

### EPA-01 — Production-path characterization

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-00 |
| Ownership surfaces | `module.extensions`, `app.server`, current profile owners, production test composition |
| Intended changes | add pre-change selectors for the five canonical IDs, two Snapshot/Reporting domains, correct owner requirements, current publication order, exact contribution parity, pre-serving dequeue, inactive code, and NF regression |
| Compatibility posture | characterization only; do not bless behavior that contradicts adopted owners |
| Risks | tests can accidentally reproduce helper paths or encode invalid publication order |
| Narrow verification | focused `module.extensions` and `app.server` rows |
| Service-backed verification | app-server/profile composition with real Postgres/job runner where state is observable |
| Generated/drift checks | Harness catalog shape if authored selectors change |
| Exit criteria | every subsequent gap has a failing or passing production-path selector; wrong-order tests are identified for correction, not frozen |
| Checkpoint | exact selector names, owner assignment, current results, expected post-remediation outcomes |
| Completion checkpoint | Added `TestExtensionProfileAdoptionMatrix_Static` under `module.extensions` for the exact five-profile, ten-job, five-worker target identity set, the shared Snapshot/Reporting participant identity, and the explicit Network Flow non-worker posture. Added `TestRuntime_ExtensionPublication_MixedClaimProfileDomains` under `app.server` for all five canonical IDs, both Snapshot/Reporting route domains, exact claimed route contribution presence, and unclaimed Network Flow. The existing `TestRuntime_ExtensionPublication_*` row remains temporary current-state characterization: its acknowledgment-before-commit calls are owner-invalid and must be rewritten in EPA-03, not treated as compatibility. Existing process selector `TestNetworkFlowHarnessRuntimeRouteServerProcessContribution` remains the service-backed Network Flow regression. Focused rows passed at `.cartulary/test-results/20260724T181632Z-p85776` and `.cartulary/test-results/20260724T181710Z-p89267`; service-backed process evidence passed at `.cartulary/test-results/20260724T181724Z-p89718`. `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make harness-contract`, `make lint-scripts`, and `git diff --check` passed after generated selector accounting was refreshed. Related failures corrected during the slice were two test compile errors in the new participant assertion, stale topology reported by `json-shape-check`, and Harness's expected selector total (`1473` to `1475`); no unrelated failure remains. Files changed: this tracker, `internal/modules/extensions/coordinator_test.go`, `internal/modules/extensions/contract_test.go`, `internal/app/server/extensions_publication_characterization_test.go`, both affected test-family manifests, generated contract embedding/render index, and the Harness contract count. Residual risk is intentionally characterized implementation gaps assigned to EPA-02 onward; rollback is the EPA-01 commit only and has no data effect. Domain vocabulary unchanged. |

### EPA-02 — Owner, authored-contract, and generation closure

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-01 |
| Ownership surfaces | Core 01/Core 04; Reporting; Report Composition; Extensions; `contracts/extensions`; generator inputs |
| Intended changes | adopt evidence-based worker/job facts, exact job contracts, explicit not-applicable facts, supported participant schemas, binding truth, closure/traceability inputs, and the clean pre-release job cutover |
| Compatibility posture | patch-level internal coordination; no public profile major or route/schema change |
| Risks | false conformance, invented legacy proofs, mismatched owner digests, partial generated set |
| Narrow verification | Extensions contract/accounting and owner locator tests |
| Service-backed verification | none; clean-cutover migration fixtures belong to EPA-04 |
| Generated/drift checks | `make generate`; `make generate-drift`; policy and JSON shape |
| Exit criteria | authored facts match adopted owners; every live job kind and worker has one contract; participant schemas resolve; all generated outputs are clean |
| Checkpoint | owner decisions, exact new identifiers/contracts, generated files, digest changes, commands/results, blocker status |
| Completion checkpoint | Amended Core 01 REQ-01-634, Extensions EXT-REQ-130, Reporting REQ-RPT-019a, and Report Composition REQ-RC-076a; versions are Extensions `0.6.2`, Reporting `1.1.1`, and Report Composition `1.1.1`. Adopted exactly ten versioned job kinds and five canonical worker kinds, route-scoped idempotency, canonical terminal success/resource-ref schemas, required terminal proofs, precommit-observable cancellation, and the clean-cutover rejection rule. Replaced the obsolete Reporting participant contract with the `snapshot_reporting` `emit` specialization using the adopted shared context/result schemas, Reporting export model, ordering, authorization, redaction, errors, and closed limits. The generator now rejects binding, worker, job, participant, schema, algorithm, locator, and digest mismatches; emits ten standalone job-contract artifacts; marks Enterprise and Network Flow `job_contract` closure explicitly `not_applicable/no_jobs`; and admits exact job and participant catalogs at runtime. The previously ignored `contracts/extensions/build` authored inputs are now tracked so a clean checkout can regenerate the catalog. Generated embeddings and TypeScript projections were refreshed. `make test-slice OWNER=module.extensions` passed at `.cartulary/test-results/20260724T184421Z-p37560`; `make service-backed-test-slice OWNER=module.extensions` passed 5/5 work units and 9 tests at `.cartulary/test-results/20260724T184532Z-p56828`; `make test-fast` passed 716 tests at `.cartulary/test-results/20260724T184835Z-p86145`. Final `make generate` passed at `.cartulary/test-results/20260724T184816Z-p81293`; `make generate-drift` passed at `.cartulary/test-results/20260724T184822Z-p82925`; generated-artifact policy, JSON-shape, Markdown lint, formatting, and `git diff --check` passed. Two related generation failures were corrected: empty preview resource-ref arrays were initially rejected at `.cartulary/test-results/20260724T184005Z-p22932`, and empty participant-implementation arrays were initially rejected at `.cartulary/test-results/20260724T184036Z-p26646`; no unrelated failure remains. `BLOCK-002` is resolved. Remaining risk is deliberately deferred runtime adoption in EPA-03 onward; this contract-only intermediate is not deployable. Rollback is the EPA-02 commit plus regeneration and has no data mutation. Domain vocabulary unchanged. |

The exact job and worker identities must use the canonical profile namespace and
the already persisted operation vocabulary. The owner amendments, not
implementation constants, become authoritative. The expected content families
are:

- Import discovery and apply;
- Incident Portability export and import;
- Reference Pack import, reverify, and refresh;
- Snapshot creation, release creation, and composition preview;
- one or more worker kinds matching the actual independently acknowledged
  handler components.

### EPA-03 — Shared publication and application composition

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-02 |
| Ownership surfaces | Extensions coordinator/plan, app publication controller, `internal/app/extensionassembly`, config projection, platform listener gates |
| Intended changes | implement `Prepare -> Commit -> acknowledge -> Serve`; expose copied typed projections; build and validate the exact contribution catalog; replace claim switch and raw enterprise check |
| Compatibility posture | public behavior unchanged; startup fails earlier on catalog mismatch |
| Risks | deadlock, partial admission, duplicate route registration, plan leakage, incorrect component identity |
| Narrow verification | coordinator plan, app-server publication, config projection, boundary guards |
| Service-backed verification | real runtime startup with mixed claims and quiescent handlers |
| Generated/drift checks | generation drift only if claim projection has generated support |
| Exit criteria | all consumers bind to one installed plan; ack-before-commit fails; no public/dequeue/workspace exposure before one gate |
| Checkpoint | component catalog, state transitions, removed derivations, exact tests/results, rollback note |
| Completion checkpoint | Added immutable application catalogs keyed by contribution ID, worker kind, job kind, participant ID, and profile binding; catalog construction rejects missing, duplicate, cross-profile, or unbound identities. The publication controller now installs one copied plan through `Prepare -> Commit -> Acknowledge -> Serve`, rejects acknowledgment before commit and duplicate/unknown/wrong-digest/failed acknowledgments, exposes only installed typed projections, and opens the HTTP, WebSocket, job-dequeue, readiness, discovery, and workspace admission gate once. Generic generated claim-key projection replaced the runtime profile switch; Enterprise preparation consumes the resolved claim set rather than raw configuration. Base route registration remains explicit, while profile route groups are selected and accounted from exact claimed HTTP contribution rows. Production consumers now receive narrow claim, route, discovery, and workspace providers; old global HTTP fallback APIs remain isolated for removal in EPA-11. The job runner queues recovery while gated and performs recovery/dequeue only after serving. Changed authored surfaces are `internal/app/extensionassembly/{configuration.go,configuration_test.go,publication_catalog.go,publication_catalog_test.go}`, `internal/app/server/{publication_controller.go,runtime.go,runtime_routes.go,runtime_routes_test.go,extensions_publication_characterization_test.go}`, `internal/modules/extensions/{claims.go,coordinator.go}`, the affected auth/import/incident-bundle/Network-Flow/reference-data/reporting/report-composition/workbook registrars and test support, `internal/platform/{config,httpapi,jobs}`, Harness manifests/accounting, one classified Enterprise fixture, generated contract/topology outputs, and this tracker. Narrow Extensions passed at `.cartulary/test-results/20260724T191353Z-p34100`; service-backed Extensions passed at `.cartulary/test-results/20260724T191453Z-p51767`; app-server narrow and service-backed slices passed at `.cartulary/test-results/20260724T192415Z-p12445` and `.cartulary/test-results/20260724T192512Z-p38674`; Auth service-backed passed at `.cartulary/test-results/20260724T191642Z-p73791`; job API and config slices passed at `.cartulary/test-results/20260724T192608Z-p63335` and `.cartulary/test-results/20260724T192616Z-p63779`; final focused catalog/publication evidence passed at `.cartulary/test-results/20260724T193832Z-p15345`. Backend unit and integration passed at `.cartulary/test-results/20260724T192643Z-p64502` and `.cartulary/test-results/20260724T193610Z-p98269`. Final `make generate`, duration-baseline coverage, Harness contract, generation drift, generated-artifact policy, JSON-shape, and `git diff --check` passed at `.cartulary/test-results/20260724T193737Z-p10030`, `.cartulary/test-results/20260724T193743Z-p11759`, `.cartulary/test-results/20260724T193934Z-p17206`, `.cartulary/test-results/20260724T194115Z-p18672`, `.cartulary/test-results/20260724T194115Z-p18671`, and `.cartulary/test-results/20260724T194115Z-p18664`. Related failures corrected during the slice were direct-registry route expectations, stale characterization tokens, missing production-shaped Enterprise test configuration, stale generated topology/counts, fixture inventory classification/order, duration accounting for a temporary row that was removed, and one missing test import. Two `make test-fast` attempts exposed unrelated nondeterministic WorkbookShell frontend failures in different surface fixtures at `.cartulary/test-results/20260724T191554Z-p68056` and `.cartulary/test-results/20260724T192134Z-p7357`; no frontend source changed, backend unit/integration and affected slices pass, and the flakes remain a final-ladder risk rather than an EPA-03 blocker. This shared runtime intermediate is not deployable; rollback is the EPA-03 commit and has no data mutation. Domain vocabulary unchanged. |

### EPA-04 — Shared job admission, proof, migration, and reconciliation

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-03 |
| Ownership surfaces | Extensions jobs protocol; platform jobs/extensionstore; app adapters; unreleased migration `00034` |
| Intended changes | complete migration `00034` with a pre-cutover-job rejection; persist owner/job metadata; validate digest-bound job contracts; gate dequeue; atomically commit success/proof; record cancellation; reconcile inactive profiles in sorted bounded transactions |
| Compatibility posture | public job resources unchanged; clean pre-release database cutover; no fallback reader or backfill |
| Risks | indeterminate commit, proof mismatch, double execution, starvation, fabricated historical evidence |
| Narrow verification | job contract validation, limit+1, ordering, failure precedence, gate tests |
| Service-backed verification | fresh schema, deterministic rejection of every retired handler identity, proof transactions, cancel races, inactive profiles, crash/restart |
| Generated/drift checks | generation; JSON shape; migration drift for the corrected authored migration |
| Exit criteria | every new extension job carries metadata; proof policy comes only from binding; inactive code never runs; no dequeue before serving |
| Checkpoint | schema/migration posture, port boundaries, legacy outcomes, files, commands/results, remaining operational risk |
| Completion checkpoint | Corrected unreleased migration `00034` in place. Its first mutation is a deterministic preflight rejecting every retained/nonterminal job under the five retired handler identities with a reset/reseed diagnostic; the migration history digest is updated, fresh and penultimate scratch application pass, and there is no alias, reader, backfill, or synthetic proof. Common-job admission now persists closed internal profile/job ownership, route-scoped idempotency identity, route/scope keys, and normalized-request digest without changing the public job resource. All ten producer paths use the canonical job kinds and all five runtime handlers use the canonical worker kinds. Application assembly derives digest-bound job contracts from the installed exact catalog, validates their claimed worker/profile pairing and resource bounds, and configures one runtime finalizer. The owner finalizer atomically joins the authoritative owner callback, canonical terminal success, immutable proof, terminal job row, and final idempotency outcome; failure rolls back every mutation, accepted cancellation is observed in its transaction, valid success wins a cancellation race, and an indeterminate final commit emits the fatal sink. Generic inactive reconciliation uses the Extensions logical port with sorted `limit+1` reads, complete pre-mutation classification, exact closed identity/result/proof validation, proof/cancellation/absence precedence, JSONB re-canonicalization, row-locked evidence revalidation, one ordered profile transaction, rollback on evidence races, and fatal indeterminate commit handling. Recovery remains queued behind the EPA-03 serve gate, so inactive profile code and pre-serving dequeue do not execute. Authored implementation surfaces are `db/migrations/00034_extension_coordination.sql`; `internal/platform/jobs/{jobs.go,extensions.go,extensions_test.go}`; `internal/platform/extensionstore/{finalizer.go,finalizer_integration_test.go,reconciliation.go,reconciliation_integration_test.go}`; `internal/modules/extensions/{coordinator.go,jobs.go,jobs_test.go,contract_test.go}`; `internal/app/extensionassembly/{jobs.go,job_reconciliation.go,publication_catalog_test.go}`; `internal/app/server/runtime.go`; the Import, Incident Portability, Reference Pack, Reporting, and Report Composition enqueue/worker files; `tools/{migration_history_manifest.json,test_families/module.extensions.json,go_test_duration_baselines.json}`; Harness accounting; generated SQL/contracts/topology outputs; and this tracker. Final Extensions focused and service-backed slices passed at `.cartulary/test-results/20260724T204139Z-p52790` and `.cartulary/test-results/20260724T204239Z-p71967`; the full service-backed scheduler passed at `.cartulary/test-results/20260724T202904Z-p39928`. Import focused/service-backed passed at `.cartulary/test-results/20260724T201913Z-p40380` and `.cartulary/test-results/20260724T201922Z-p41486`; Incident Portability at `.cartulary/test-results/20260724T201935Z-p41884` and `.cartulary/test-results/20260724T201956Z-p42892`; Reference Pack at `.cartulary/test-results/20260724T202018Z-p43456` and `.cartulary/test-results/20260724T202112Z-p59600`; Reporting at `.cartulary/test-results/20260724T202215Z-p75437` and `.cartulary/test-results/20260724T202225Z-p76219`; common jobs at `.cartulary/test-results/20260724T202254Z-p77470` and `.cartulary/test-results/20260724T202304Z-p78085`; app-server at `.cartulary/test-results/20260724T202315Z-p78482` and `.cartulary/test-results/20260724T202407Z-p4471`; Report Composition focused passed at `.cartulary/test-results/20260724T202238Z-p76616` and has no service-backed row, as confirmed by the expected zero-row diagnostic at `.cartulary/test-results/20260724T202242Z-p76986`. Backend unit passed 302 tests at `.cartulary/test-results/20260724T204339Z-p88059`. Final generation drift, migration drift, duration-baseline coverage, Harness contract, generated-artifact policy, and JSON-shape checks passed at `.cartulary/test-results/20260724T204410Z-p94146`, `.cartulary/test-results/20260724T204420Z-p97234`, `.cartulary/test-results/20260724T204427Z-p99499`, `.cartulary/test-results/20260724T204446Z-p329`, `.cartulary/test-results/20260724T204513Z-p1440`, and `.cartulary/test-results/20260724T204515Z-p1783`; `git diff --check` passed. Related failures corrected during the slice were a missing Goose statement boundary, stale service schema hash after changing the migration, merged Harness selector accounting, empty isolated user fixtures, JSONB canonical-byte comparison, artificial Harness family IDs, missing duration baselines, and cancel-requested finalization admission. `make go-test-duration-baseline-drift` without `RESULTS_DIR` produced the expected usage failure at `.cartulary/test-results/20260724T204429Z-p99704`; coverage and Harness contract are the applicable EPA-04 gates. Profile-specific authoritative owner mutations consume this finalizer in EPA-06 through EPA-09, so the EPA-04 commit remains a deliberately non-deployable shared intermediate. Rollback before rollout is the EPA-04 commit boundary plus development database reset/reseed; after eventual rollout, proof and cancellation rows are immutable and must never be deleted or reinterpreted. No blocker remains. Domain vocabulary unchanged. |

Pre-cutover jobs are not adopted. The owner amendment must prescribe a
fail-closed reset/reseed action for a database containing any job associated with
the retired handler identities. The implementation must not create a synthetic
proof, backfill ownership, or treat a null metadata row as a runtime
compatibility format.

### EPA-05 — Enterprise Authentication adoption

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-03 |
| Ownership surfaces | auth, enterpriseauth platform adapter, app assembly, enterprise frontend |
| Intended changes | prepare provider/secret/egress behavior only from resolved claims; bind all four contribution IDs; centralize owner-local profile constant |
| Compatibility posture | preserve routes, sessions, bindings, provider manifest, authorization, and UI |
| Risks | provider secrets loaded for inactive profile; route family partially registered; login regression |
| Narrow verification | auth contract/unit, config inactive policy, contribution parity |
| Service-backed verification | OIDC/SAML/provider/binding integration and app startup |
| Generated/drift checks | generation drift; no migration |
| Exit criteria | raw config cannot independently activate code; exact routes prepare and ack; all N/A capabilities are proven |
| Checkpoint | provider preparation path, route IDs, frontend result, commands, risks, domain posture |

### EPA-06 — Import adoption

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-04 |
| Ownership surfaces | imports, common jobs, source-owner facades, Network Flow adapter, frontend application |
| Intended changes | bind discovery/apply jobs and workers; emit proofs; preserve owner dispatch; add the Phase 2 Workbook Import Assistant and claim-loss/Base behavior |
| Compatibility posture | existing Import routes/resources remain unchanged; additive UI |
| Risks | duplicate apply, incorrect target ownership, source loss, UI presenting partial success |
| Narrow verification | Import integration plus frontend unit/model tests |
| Service-backed verification | CSV/XLSX discovery/apply, cancel, replay, source change, NF target, job proof |
| Generated/drift checks | generation drift; no migration beyond shared EPA-04 decision |
| Exit criteria | exact jobs/workers match binding; assistant meets AC-027; NF scheduling remains Import-owned |
| Checkpoint | job kinds, source-owner matrix, browser scenarios, commands/results, compatibility |

### EPA-07 — Incident Portability adoption

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-04 |
| Ownership surfaces | incidentbundles, incidentportability, stagedobjects, crossownertransaction, revisions, source participants |
| Intended changes | bind export/import jobs and proofs to publication; retain atomic staged import, exact participant order, and NF blocking |
| Compatibility posture | bundle format/routes and base-only compatibility unchanged |
| Risks | payload loss, partial imported incident, stale staged object, attribution drift |
| Narrow verification | unit portability matrix and job contract tests |
| Service-backed verification | export/import, cancellation, final publication recheck, NF blocked state, atomic cleanup/proof |
| Generated/drift checks | generation drift; migration drift only if EPA-04 requires it |
| Exit criteria | export/import terminal state and resources share final proof boundary; inactive profile code is never invoked |
| Checkpoint | participant/state matrix, staged refs, transaction result, commands/results, risks |

### EPA-08 — Reference Pack adoption

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-04 |
| Ownership surfaces | reference_data, common jobs, config/egress, deployment-admin frontend |
| Intended changes | bind import/reverify/refresh jobs, proofs, cancellation, recovery, and readiness to exact plan |
| Compatibility posture | pack routes, verification, disconnected seed, admin authorization, and UI unchanged |
| Risks | network access while inactive, prior-active loss, pack activation without proof |
| Narrow verification | Reference Pack unit/frontend and contract rows |
| Service-backed verification | import, reverify, refresh, cancel, replay, inactive/disconnected startup, prior-active retention |
| Generated/drift checks | generation drift; no extension state/codec artifact |
| Exit criteria | every job kind is contract-bound; inactive profile performs no refresh/verification; UI follows serving epoch |
| Checkpoint | job matrix, egress posture, N/A state/recovery evidence, commands/results |

### EPA-09 — Snapshot and Reporting adoption

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-04 |
| Ownership surfaces | reporting, reportcomposition, participant coordinator/adapter, common jobs |
| Intended changes | invoke admitted `render_export_v1`; bind snapshot/release/preview jobs; route preview through Reporting; preserve Report Composition non-participant boundary |
| Compatibility posture | routes, snapshot/release/composition schemas, rendering ownership, and optional UI posture unchanged |
| Risks | recursion in self-owned participant adapter, render output published before validation, preview becoming approval evidence |
| Narrow verification | participant closed-result/deadline tests; reporting and composition units |
| Service-backed verification | snapshot/release/preview success/failure/cancel/replay, composition tuple, proof/resource refs |
| Generated/drift checks | generation and JSON shape; no extra recovery/state artifact |
| Exit criteria | only the admitted participant runs; output admission remains Reporting-owned; preview terminates; all jobs are exact |
| Checkpoint | invocation diagram, result admission, job kinds, optional UI statement, commands/results |

### EPA-10 — Harness ownership and documentation reconciliation

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-05, EPA-06, EPA-07, EPA-08, EPA-09 |
| Ownership surfaces | verification owners, test families, generated topology, docs/guides, tracker |
| Intended changes | assign final selectors to semantic owners; regenerate topology; update developer/testing guidance and all tracker checkpoints |
| Compatibility posture | evidence/docs only |
| Risks | duplicate selector, wrong owner, helper evidence presented as production proof |
| Narrow verification | `make explain-test-owner` for every affected owner; exact row runs |
| Service-backed verification | all new service-backed owner rows |
| Generated/drift checks | Harness-owned generation, generation drift, policy, JSON shape |
| Exit criteria | every behavior has one owner and production selector; docs match code; `domain vocabulary unchanged` recorded |
| Checkpoint | owner row totals, added/retained/retired selectors, generated topology, docs, commands/results |

### EPA-11 — Obsolete-path removal and static DAG guards

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-10 |
| Ownership surfaces | platform HTTP fallback, unused proof wrappers, application/profile imports, backend/frontend boundary manifests |
| Intended changes | delete global registry defaults, nil epoch readers, duplicate claim/route derivations, unused wrappers, and compatibility test mutation; add static guards |
| Compatibility posture | no aliases; tests use explicit production-shaped fixtures |
| Risks | hidden caller, test coupling, accidental concrete-profile import in generic owner |
| Narrow verification | caller scans, boundary tests, affected units |
| Service-backed verification | app/profile owner slices after deletion |
| Generated/drift checks | boundary/policy and generation drift |
| Exit criteria | no obsolete token/caller remains; target DAG is mechanically enforced |
| Checkpoint | removed symbols/files, caller scans, import graph, commands/results, rollback |

### EPA-12 — Final validation and handoff

| Field | Plan |
| --- | --- |
| Initial status | `NOT_STARTED` |
| Dependencies | EPA-11 |
| Ownership surfaces | entire affected repository and this tracker |
| Intended changes | validation and checkpoint updates only; no new behavior after final ladder begins |
| Compatibility posture | cold restart; no partial or mixed generated/runtime deployment |
| Risks | broad regression, retained evidence mismatch, incomplete migration rehearsal |
| Narrow verification | every owner slice and exact selector |
| Service-backed verification | every affected service-backed owner and process path |
| Generated/drift checks | full generation/policy/shape/migration/Harness ladder |
| Exit criteria | all binary criteria pass; every workstream is `DONE`; final handoff is complete |
| Checkpoint | final commit/status, files, commands/run roots, failures/retries, rollout/rollback, blockers, domain posture |

## 14. Rollback and Safe-Reversion Boundaries

| Boundary | Safe reversion rule |
| --- | --- |
| EPA-01 characterization | revert only its tests/Harness inputs; never retain a test that blesses owner-invalid behavior |
| EPA-02 owner/contracts/generated | one atomic source+generated commit; revert owner inputs and regenerate, never edit generated files backward |
| EPA-03 publication | revert controller, app catalog, config projection, and component adapters together; do not retain two epochs |
| EPA-04 common jobs | old binary may ignore additive metadata/proof rows; never delete proof history; quiesce extension jobs before binary rollback |
| Profile slices | revert profile adapter, tests, and binding inputs together; do not leave a generated implemented contribution without implementation |
| Frontend Import Assistant | may revert independently before release because public backend is unchanged; remove its support advertisement with it |
| Migration | migration `00034` is corrected in place because it is unreleased; databases containing pre-cutover handler jobs are rejected and must be reset/reseeded |
| EPA-11 deletion | revert only to the immediately preceding single-path architecture; do not restore obsolete globals as a second live path |
| Deployment | no intermediate workstream deploys; final rollout is a monitored cold stop/start with one generated/runtime version |

If rollback occurs after proof-bearing jobs have run, the database keeps their
metadata, proofs, and cancellation observations. An older binary may ignore
additive internal rows but must not erase or reinterpret them.

## 15. Harness Ownership and Selector Plan

### Current owner baseline

| Owner | Rows | Service-backed | Evidence retained |
| --- | ---: | ---: | --- |
| `module.extensions` | 20 | 9 | coordinator, contract accounting, state, process, one NF stateful browser row |
| `app.server` | 28 | 22 | publication characterization, readiness, process/runtime composition |
| `module.auth` | 48 | 29 | enterprise backend/frontend/browser plus Base auth |
| `module.imports` | 3 | 3 | CSV mapping/apply, invalid metadata, XLSX discovery |
| `module.incidentbundles` | 20 | 10 | portability, export/import, atomicity, worker behavior |
| `module.reference_data` | 16 | 10 | pack lifecycle, jobs, admin frontend/browser |
| `module.reporting` | 10 | 4 | snapshot/release service evidence and units |
| `module.reportcomposition` | 1 | 0 | traceability fixture only |
| `module.jobapi` | 3 | 3 | common job authorization |
| `platform.config` | current live manifest | no service-backed recommendation | inactive and deployment configuration |
| `web.application` | 60 | 0 | frontend units and boundary support |

### Existing selectors to retain or strengthen

- Extensions coordinator registry, claims, collisions, binding admission,
  publication plan, contract accounting, and state admission selectors.
- app-server
  `TestRuntime_ExtensionPublication_*` selectors, rewritten where they encode
  acknowledgment before commit.
- enterprise OIDC/SAML/provider/binding integration and
  `apps/web/e2e/enterprise-auth.spec.ts`.
- all three Import service-backed selectors.
- Incident Bundle portability state/claim matrix, atomic import, final
  publication recheck, NF blocking, and job authorization selectors.
- Reference Pack import/lifecycle/disconnected and admin browser selectors.
- Reporting snapshot/release service-backed selectors.
- Report Composition traceability and release-tuple fixtures.
- common Job API authorization selectors.
- `apps/web/e2e/extensions.stateful.spec.ts` as Network Flow/shared publication
  regression, not evidence for the five profile-specific paths.

### Missing selectors to author

| Owner | Planned selector or browser scenario |
| --- | --- |
| `module.extensions` | `TestExtensionProfileAdoptionMatrix_Static`; `TestInactiveExtensionJobReconciliation_Integration`; exact job/worker/binding accounting |
| `app.server` | `TestRuntime_ExtensionPublication_CommittedBeforeAcknowledgment`; exact route/workspace/worker/participant preparation; no pre-serving dequeue; mixed-claim process path |
| `platform.config` | generated claim projection handles every descriptor and a future test descriptor without a runtime switch |
| `module.auth` | enterprise provider preparation follows resolved claim rather than raw config |
| `module.imports` | discovery/apply metadata, proof, cancellation, inactive reconciliation, and production composition |
| Import browser owner | claimed Import Assistant discovery/progress/cancel/map/apply; claim loss; Base fallback |
| `module.incidentbundles` | export/import job proof and shared serving-gate behavior |
| `module.reference_data` | import/reverify/refresh job contracts and inactive no-code/no-egress behavior |
| `module.reporting` | admitted participant invocation; snapshot/release job contracts and proof behavior |
| `module.reportcomposition` | service-backed preview job reaches Reporting terminal result and never becomes approval evidence |
| `module.jobapi` | internal extension metadata is not leaked in the public common-job resource |
| `app.server` process evidence | all five IDs claimed/unclaimed in production assembly; Network Flow reference still works |

Harness ownership follows semantics: generic reconciliation belongs to
`module.extensions`, process/catalog parity to `app.server`, profile jobs to
their profile owner, job wire compatibility to `module.jobapi`, and browser
behavior to the owning frontend/profile row. A collaborator ID does not transfer
semantic ownership.

## 16. Generated-Artifact Implications

Later implementation may change authored Extensions fragments, owner manifests,
configuration/participant/job contracts, implementation binding sources,
closure/traceability inputs, validation surfaces, and Harness catalogs.

The implementation agent must:

1. update adopted owner documents before owner fragments or manifests;
2. update authored contract inputs;
3. run `make generate`;
4. inspect the exact generated diff;
5. run drift, policy, and JSON shape checks;
6. never hand-edit `internal/gen/**`,
   `packages/protocol-ts/src/generated/**`,
   `packages/ui-contracts/src/generated/**`, generated topology, or lockfiles.

Expected digest changes are internal and require one cold generated/runtime
deployment epoch. Contract major 1 remains correct because no public profile
wire shape or semantics change.

## 17. Compatibility and Migration Analysis

| Surface | Posture |
| --- | --- |
| Profile discovery | IDs, majors, routes, claimability, ordering, and workspace arrays unchanged |
| HTTP/OpenAPI | existing routes and public schemas unchanged |
| Configuration | existing keys/defaults unchanged; internal claim projection changes |
| Frontend | Import Assistant is additive; other availability remains discovery-driven |
| Common jobs | public resource unchanged; internal owner/job metadata added |
| Database | complete unreleased migration `00034` in place; fresh or reset/reseeded databases only |
| Existing nonterminal jobs | unsupported; migration precondition rejects the database before mutation |
| Existing terminal successes | unsupported; no proof backfill or historical job-kind adoption |
| Backup/restore | base Postgres captures core-managed profile tables; only NF retains extension binding/codecs |
| Incident bundles | bundle format and current optional/base behavior unchanged |
| Generated contracts | internal digests change atomically; no mixed binary/artifact version |
| Rollout | monitored cold restart after full EPA-12 evidence |

No compatibility shim is justified. Internal migration effort is not continuing
value for a dual reader.

## 18. Validation Commands from the Live Make Surface

### Per-workstream selection

Use the narrowest applicable command first:

```text
make task-guide ROLE=module-author OWNER=<owner-id>
make explain-test-owner OWNER=<owner-id>
make test-slice OWNER=<owner-id> [ROWS=<row-id,...>]
make service-backed-test-slice OWNER=<owner-id> [ROWS=<row-id,...>]
```

Confirmed relevant owner IDs are:

```text
module.extensions
app.server
module.auth
module.imports
module.incidentbundles
module.reference_data
module.reporting
module.reportcomposition
module.jobapi
module.crossownertransaction
module.stagedobjects
module.recovery
platform.config
platform.telemetry
web.application
web.networkflow
```

### Browser and boundary validation

```text
make browser-e2e-stateful
make browser-e2e-webserver-backed
make backend-module-boundary-check
make frontend-import-boundary-check
```

Use stateful browser evidence for publication and Import availability. Use
webserver-backed evidence for enterprise auth, Reference Packs, and any
profile-specific production route/UI path.

### Generated, migration, documentation, and Harness validation

```text
make generate-drift
make generated-artifact-policy-check
make json-shape-check
make migration-drift
make harness-contract
make lint-markdown
make lint-scripts
make lint-shell
```

### Final validation order

1. run all affected focused owner slices;
2. run all applicable service-backed owner slices;
3. run stateful and webserver-backed browser evidence;
4. run backend/frontend boundaries;
5. run Harness, generation, policy, JSON, migration, Markdown, script, and shell
   checks;
6. run `make check` and retain its successful warm results root;
7. run `make agent-finalize RESULTS_DIR=<successful-check-run-root>`;
8. rerun any drift checks changed by finalization;
9. run final `make check`;
10. run `make release-check`.

If no successful full warm check root exists, run `make agent-finalize` without
`RESULTS_DIR` and record that retained-run maintenance was skipped because the
variable was unset. Never invent a results path.

## 19. Execution Protocol and Session-Handoff Tables

Every implementation session must follow this protocol:

1. mark exactly one workstream `IN_PROGRESS`;
2. complete its implementation and validation;
3. record files changed, decisions, commands, results, failures, remaining
   risks, blockers, and domain-vocabulary posture;
4. mark it `DONE` only after every exit criterion passes;
5. mark the next workstream `IN_PROGRESS` before changing its files;
6. finish with dedicated `EPA-12` validation and handoff.

### Top-level work tracker

| Workstream | Status | Dependency | Deployment posture | Checkpoint |
| --- | --- | --- | --- | --- |
| EPA-00 | `DONE` | none | no behavior | baseline and clean-cutover decision recorded; tracker-only gates passed |
| EPA-01 | `DONE` | EPA-00 | evidence only | exact target matrix and mixed-claim production characterization pass |
| EPA-02 | `DONE` | EPA-01 | non-deployable contract intermediate | exact owner jobs/workers, shared schemas, participant specialization, binding parity, and generated closure pass |
| EPA-03 | `DONE` | EPA-02 | non-deployable shared runtime intermediate | one installed typed epoch, exact catalogs, generic claims, commit-before-ack, and shared serve/dequeue gate pass |
| EPA-04 | `DONE` | EPA-03 | non-deployable job intermediate | canonical metadata/workers, clean migration, atomic finalization/cancellation, gated recovery, and inactive reconciliation pass |
| EPA-05 | `NOT_STARTED` | EPA-03 | wait for final rollout | pending |
| EPA-06 | `NOT_STARTED` | EPA-04 | wait for final rollout | pending |
| EPA-07 | `NOT_STARTED` | EPA-04 | wait for final rollout | pending |
| EPA-08 | `NOT_STARTED` | EPA-04 | wait for final rollout | pending |
| EPA-09 | `NOT_STARTED` | EPA-04 | wait for final rollout | pending |
| EPA-10 | `NOT_STARTED` | EPA-05..09 | evidence/docs only | pending |
| EPA-11 | `NOT_STARTED` | EPA-10 | no dual path | pending |
| EPA-12 | `NOT_STARTED` | EPA-11 | final cold-rollout candidate | pending |

### Workstream checkpoint template

| Field | Required entry |
| --- | --- |
| Date/session | timestamp and agent/session identity |
| Workstream/status | one EPA ID and before/after status |
| Files changed | exact authored and generated paths |
| Decisions | owner, compatibility, migration, and evidence decisions |
| Domain posture | `domain vocabulary unchanged` or exact owner-approved edit |
| Narrow commands/results | exact commands and pass/fail |
| Service-backed commands/results | exact commands, run roots, and pass/fail |
| Generated/drift results | exact commands and diff posture |
| Failures | related/unrelated classification and summary artifact |
| Remaining risks | explicit residual risk |
| Blockers | blocker IDs and owner needed |
| Rollback point | commit/boundary and data posture |
| Next action | one next workstream only |

### Session handoff log

| Date | Session | Workstream | Status | Files changed | Validation/result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-24 | Tracker research and planning | planning only | complete | this tracker only | documentation-only checks recorded at handoff | BLOCK-001 was open at planning handoff | activate EPA-00 |
| 2026-07-24 | Remediation execution | EPA-00 | complete | this tracker only | tracker-only validation passed; run roots recorded in EPA-00 | BLOCK-001 resolved; BLOCK-002 remains for EPA-02 | commit EPA-00, then activate EPA-01 |
| 2026-07-24 | Remediation execution | EPA-01 | complete | tracker; Extensions/app-server characterization tests; Harness manifests, generated embedding/render index, and count | focused and service-backed selectors plus generation/Harness drift passed; run roots recorded in EPA-01 | BLOCK-002 remains for EPA-02 | commit EPA-01, then execute EPA-02 |
| 2026-07-24 | Remediation execution | EPA-02 | complete | owner specs/manifests/fragments; Extensions build/specification/traceability inputs; generator/runtime catalogs/tests; generated Go and TypeScript projections; tracker and `.gitignore` | exact job/worker and participant closure, focused/service-backed owner slices, 716-test fast suite, generation/drift/policy/shape/Markdown gates passed; run roots recorded in EPA-02 | none; BLOCK-002 resolved | commit EPA-02, then activate EPA-03 |
| 2026-07-24 | Remediation execution | EPA-03 | complete | tracker; Extensions/app-server/config/jobs/HTTP composition; profile registrars and test support; Harness manifests/accounting; generated embedding/render index | exact catalogs, generic claims, commit-before-ack, typed copied projections, one serve/dequeue gate, affected owner slices, backend unit/integration, generation/Harness/drift gates passed; run roots and two unrelated frontend flakes recorded in EPA-03 | none | commit EPA-03, then activate EPA-04 |
| 2026-07-24 | Remediation execution | EPA-04 | complete | tracker; migration 00034/history; common jobs and extension store; Extensions logical reconciliation; app assembly/runtime; five profile enqueue/worker identities; Harness manifests/accounting/baselines; generated SQL/contracts/topology | exact metadata/catalog binding, clean-cutover rejection, atomic proof/cancel/finalization, proof-first bounded reconciliation, race rollback, gated recovery, affected owner slices, full service-backed scheduler, backend unit, generation/migration/Harness/drift gates passed; exact roots recorded in EPA-04 | none | commit EPA-04, then activate EPA-05 |

## 20. Open Contradictions and Blockers

| ID | Kind | Exact issue | Required resolution | Blocks |
| --- | --- | --- | --- | --- |
| BLOCK-001 | resolved specification decision | pre-cutover profile-handler jobs have no trustworthy proof transition | clean pre-release cutover selected: reject affected databases and require reset/reseed; no reader, backfill, or synthetic proof | none; EPA-02 records the owner rule and EPA-04 enforces it |
| BLOCK-002 | resolved authored-contract contradiction | `reporting-participant.json` named context/result schema IDs without an admitted specialization implementation contract | adopted the exact shared Snapshot/Reporting context/result schemas and Reporting-owned `emit` specialization; regenerated all dependent digests and runtime catalogs | none |

No contradiction remains about profile identity: Core 00 conclusively maps the
two requested Snapshot and Reporting audits to `snapshot_reporting`.

No contradiction remains about Network Flow scope: its adopted owner makes it a
shared regression reference and explicitly leaves its scheduling with Import.

No contradiction remains about Reference Pack authority: the dedicated NLSpec is
draft, so current Core owners prevail.

## 21. Binary Completion Criteria

Tracker authoring is complete only when:

- [x] scope, authority, assumptions, and non-goals are explicit;
- [x] current source baseline and repository status are recorded;
- [x] normative and implementation-support documents are inventoried;
- [x] every in-scope profile and both `snapshot_reporting` domains have exact
  file/caller inventories;
- [x] every capability cell uses one allowed disposition;
- [x] runtime consumers are mapped;
- [x] authored/generated provenance is mapped;
- [x] current and target DAGs are recorded;
- [x] per-profile and systemic findings cite exact live evidence;
- [x] every gap has remediation, area, rationale, benefit, compatibility,
  unresolved risk, dependencies, validation, and owner workstream;
- [x] workstreams are independently verifiable and dependency ordered;
- [x] rollback and safe-reversion boundaries are explicit;
- [x] Harness owners, retained selectors, and missing evidence are identified;
- [x] generated implications and migration posture are explicit;
- [x] live Make commands are recorded without inventing targets;
- [x] execution protocol, top-level status, and checkpoint templates are present;
- [x] blockers are explicit;
- [x] `network_flow_activity` remains reference-only except for shared regression;
- [x] `domain vocabulary unchanged` is recorded;
- [x] documentation-only validation in Section 22 passes.

Remediation completion is separately binary:

- [ ] all five canonical profile IDs and both `snapshot_reporting` domains match
  their adopted owners;
- [ ] every applicable capability is `conformant and proven` and every absence
  is `declared not applicable`;
- [ ] no specification, contract/generation, implementation, wiring, test,
  Harness, documentation, or owner-contradiction disposition remains;
- [ ] every live extension job is contract-bound, gated, reconciled, and proven;
- [ ] publication uses the exact committed-then-acknowledged lifecycle;
- [ ] every runtime consumer uses the installed typed epoch;
- [ ] no registry fallback, hard-coded generic profile list, dual path, or unused
  wrapper remains;
- [ ] Network Flow regression evidence passes without giving it profile-specific
  work;
- [ ] EPA-00 through EPA-12 are `DONE`;
- [ ] final `make check` and `make release-check` pass.

## 22. Tracker-Authoring Validation and Handoff

This task changes only this Markdown tracker. Required validation:

```text
git diff --check
make generated-artifact-policy-check
make json-shape-check
make lint-markdown
```

Intentionally skipped for this documentation-only planning task:

- `make format`, because no Go or frontend source changed and AGENTS.md says not
  to run it solely for Markdown;
- product unit/integration/process and service-backed suites, because no product
  behavior changed;
- browser suites, because no frontend behavior changed;
- `make generate` and `make generate-drift`, because no generator input changed;
- `make migration-drift`, because no SQL or storage behavior changed;
- `make agent-finalize`, because no Harness maintenance input changed;
- `make check` and `make release-check`, because the narrow documentation gates
  cover this tracker-only change and the broad product/release gates belong to
  EPA-12.

Final validation results and the final diff summary must be entered here before
handoff.

| Command | Result |
| --- | --- |
| `git diff --check` | passed |
| `GIT_INDEX_FILE=<temporary-index> git diff --check` | passed; an alternate index with the tracker marked intent-to-add validates the new file itself |
| `make generated-artifact-policy-check` | passed; run root `.cartulary/test-results/20260724T175401Z-p57911` |
| `make json-shape-check` | passed; run root `.cartulary/test-results/20260724T175401Z-p57897` |
| `make lint-markdown` | passed |

### Planning handoff summary

- Planning result: five canonical profile IDs, six requested audit domains, 14
  gaps, two blockers, and 13 ordered workstreams.
- Files inspected: normative owners, applicable guides and historical trackers,
  all authored Extensions inputs, embedded generated registries, application
  composition, profile implementations, platform adapters, storage inputs,
  frontend consumers, tests, verification owners, and Harness manifests listed
  above.
- Tracker created:
  `docs/handoffs/extensions-profile-adoption-refactor-tracker.md`.
- Major findings: publication order is inverted; exact plan projections are not
  consumed; four profiles have undeclared durable jobs; inactive reconciliation
  and dequeue gating are absent; Reporting's declared participant is not
  implemented; Import lacks its production assistant.
- Proposed workstreams: EPA-00 through EPA-12.
- Unresolved blockers: historical job adoption policy and the unsupported
  Snapshot/Reporting participant schema IDs.
- Domain vocabulary: unchanged.
