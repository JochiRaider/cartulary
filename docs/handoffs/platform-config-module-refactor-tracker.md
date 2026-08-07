# platform-config Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/platform/config`
- **Target label:** `platform-config` (derived from the target path and normalized to
  lowercase kebab case)
- **Output path:** `docs/handoffs/platform-config-module-refactor-tracker.md`
- **Status:** Prior iteration complete; WS-00 through WS-10 passed their required gates.
  The production-readiness cleanup iteration in Section 13 is planned; WS-11 completed
  its documentation gate and WS-12 is the sole next implementation slice, pending
  separate authorization.
- **Allowed change:** This request authorizes only the WS-11 tracker update. The
  implementation, tests, boundary policy, verification inputs, and generated artifacts
  named by WS-12 through WS-17 require a separately authorized implementation task.
- **Non-goals:** No HTTP, database, browser-facing, or deployment-config v2 schema
  migration; no v1 compatibility reader; no deprecated alias for removed internal Go
  APIs; no unrelated product work.
- **Default posture:** Preserve the valuable external `cartulary.deployment_config.v2`
  contract, strict parsing, overlay behavior, defaults, diagnostics, and fail-closed
  admission. Remove internal compatibility only when the controlling workstream proves
  that it has no continuing value.
- **Implementation authorization:** Explicitly granted and completed on 2026-08-07 for
  WS-00 through WS-10. WS-11 is authorized as a documentation-only step; WS-12 through
  WS-17 are planned but not yet authorized for implementation.

The source hierarchy used for this tracker is:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication; it is not
   applicable to this planning-only target.
4. Domain vocabulary and implementation-support guidance.
5. Live repository code, tests, contracts, fixtures, and harness maps.
6. This framework and other prior handoffs as evidence only.

The initial inventory found a self-contradiction inside the adopted Extensions NLSpec:
its current projection boundary requires descriptor/configuration v3 without a
compatibility reader, while later requirements and tables require v1. The authorized
remediation treats the declared v3/v2 current projection as the intended owner state and
repairs the stale normative clauses before any dependent implementation change. The
former RB-003 blocker is therefore converted into the WS-01 prerequisite rather than
silently resolved in implementation.

### 1.1 Implementation control and next iteration

This subsection supersedes the planning-only status and deferred implementation rows in
Sections 2 through 8 and the legacy rows in Sections 9 and 11. Those entries remain as
dated discovery history. The external
deployment-config v2 surface is frozen; internal Go surfaces may break atomically and do
not receive compatibility shims.

| ID | Workstream | Depends on | Status | Exit evidence | Next gate |
| --- | --- | --- | --- | --- | --- |
| WS-00 | Rebaseline this tracker with the authorized end-to-end plan and mandatory evidence gate | None | DONE | This subsection; `make lint-markdown` | WS-01 |
| WS-01 | Normalize Extensions, Core 04, OTel, and Core 01 specification ownership and current schema terminology | WS-00 | DONE | Normative search/review; `make lint-markdown` run root `.cartulary/test-results/20260807T163557Z-p3341552` | WS-02 |
| WS-02 | Harden authored extension projections and generator validation; regenerate declared outputs | WS-01 | DONE | Generate `.cartulary/test-results/20260807T164359Z-p3355367`; drift `.cartulary/test-results/20260807T164420Z-p3357844`; policy/shape and Extensions owner results in Section 10 | WS-03 |
| WS-03 | Add external-behavior and claim-boundary characterization | WS-02 | DONE | Config 19/19 `.cartulary/test-results/20260807T164808Z-p3393805`; Extensions 37/37 `.cartulary/test-results/20260807T164817Z-p3395625` | WS-04 |
| WS-04 | Remove dead kernel APIs and privatize defaults after recovery-default equivalence coverage | WS-03 | DONE | Config 19/19 `.cartulary/test-results/20260807T165151Z-p3422982`; recovery 37/37 `.cartulary/test-results/20260807T165151Z-p3422975`; removed-symbol search clean | WS-05 |
| WS-05 | Contract configassembly and move explicit-path loading to test support | WS-04 | DONE | Unit and service-backed run roots recorded in Section 10 | WS-06 |
| WS-06 | Enforce the owner-neutral platform-config dependency boundary | WS-05 | DONE | Boundary 3/3 `.cartulary/test-results/20260807T170115Z-p3626189`; config 19/19 `.cartulary/test-results/20260807T170121Z-p3626568`; tracker lint `.cartulary/test-results/20260807T170201Z-p3631648` | WS-07 |
| WS-07 | Implement explicit owner-supplied namespace decoding, projection, validation, and cloning | WS-06 | DONE | Config 19/19 `.cartulary/test-results/20260807T171814Z-p3774363`; telemetry 6/6 `.cartulary/test-results/20260807T171705Z-p3766838`; tracker lint `.cartulary/test-results/20260807T171931Z-p3782981`; affected owner/server/boundary rows in Section 10 | WS-08 |
| WS-08 | Move telemetry wire semantics and verification ownership to platform.telemetry | WS-07 | DONE | Telemetry 10/10 `.cartulary/test-results/20260807T172538Z-p3804410`; config 15/15 `.cartulary/test-results/20260807T172550Z-p3805815`; OTel 14/14 `.cartulary/test-results/20260807T172551Z-p3805998`; tracker lint `.cartulary/test-results/20260807T172711Z-p3814843` | WS-09 |
| WS-09 | Replace hard-coded Boolean paths with descriptor-derived typed requested claims | WS-02, WS-07, WS-08 | DONE | Config 15/15 `.cartulary/test-results/20260807T175144Z-p4075770`; Extensions unit 37/37 `.cartulary/test-results/20260807T174214Z-p3875258` and service-backed 24/24 `.cartulary/test-results/20260807T174252Z-p3902644`; Network Flow 71/71 `.cartulary/test-results/20260807T174322Z-p3925490`; server unit 42/42 `.cartulary/test-results/20260807T174951Z-p4042969` and service-backed 33/33 `.cartulary/test-results/20260807T174448Z-p3963442`; stateful browser 36/36 `.cartulary/test-results/20260807T174752Z-p4016912`; tracker lint `.cartulary/test-results/20260807T175232Z-p4077806` | WS-10 |
| WS-10 | Reconcile harness ownership, run final validation, and complete handoff | WS-01 through WS-09 | DONE | Finalize `.cartulary/test-results/20260807T175350Z-p4081227`; fast 349/349 `.cartulary/test-results/20260807T175533Z-p4100146`; check 722/722 `.cartulary/test-results/20260807T175802Z-p4193493`; release 893/893 `.cartulary/test-results/20260807T180201Z-p146336`; tracker lint `.cartulary/test-results/20260807T181236Z-p335003`; remaining evidence in Section 12 | Complete |
| WS-11 | Rebaseline this tracker with the production-readiness cleanup iteration | WS-10 | DONE | Section 13 audit and plan; initial and post-status `make lint-markdown` runs `.cartulary/test-results/20260807T185329Z-p348875` and `.cartulary/test-results/20260807T185415Z-p350661` | WS-12 after separate implementation authorization |
| WS-12 | Remove dead request-state projections | WS-11 | READY — AUTHORIZATION REQUIRED | Removed-symbol audit and focused config/Extensions/configassembly tests | WS-13 |
| WS-13 | Contract snapshots and owner presence to typed, namespace-scoped seams | WS-12 | PLANNED | Config, telemetry, OTel, isolation, and defensive-copy evidence | WS-14 |
| WS-14 | Separate one-time structural admission from startup filesystem readiness | WS-13 | PLANNED | Structural/startup diagnostic parity and config/server tests | WS-15 |
| WS-15 | Remove test-only production admission and migrate application/test loading | WS-14 | PLANNED | Builds plus affected unit, service-backed, recovery, and stateful-browser evidence | WS-16 |
| WS-16 | Enforce the contracted boundary and prune residual legacy support | WS-15 | PLANNED | Negative boundary fixtures, removed-symbol audit, and generated drift | WS-17 |
| WS-17 | Run final validation and complete the production-readiness handoff | WS-12 through WS-16 | PLANNED | Finalization, focused, broad, security, release, and tracker evidence | Complete |

After every workstream, before work begins on the next one:

1. Record timestamp, agent, owners, files, substantive decision, and compatibility
   effect in Section 10.
2. Record every validation command and its run root or summary artifact; explain any
   failure and do not advance while a required result is unresolved.
3. Mark the completed row `DONE`, mark exactly one successor `READY`, and update risks
   or follow-ups.
4. Run `make lint-markdown` and record the result.

The implementation decisions controlling all workstreams are:

- v3/v2 Extensions projections are canonical and v1 artifacts are rejected without a
  compatibility reader.
- `internal/platform/config` ends as an owner-neutral Core 04 kernel; owners supply
  namespace decoding and semantics through explicit static application assembly.
- Package initialization, runtime/source scanning, reflection-based owner discovery,
  mutable global registries, and dynamic plugin loading remain prohibited.
- Telemetry wire semantics move to `internal/platform/telemetry`; generic parsing and
  admission remain Core 04 responsibilities.
- Claim registrations derive from validated Extensions artifacts. Requested claims,
  resolved claims, and resolved-claim-set identity remain mechanically distinct.
- `docs/domain.md` changes only if a new domain term emerges; implementation vocabulary
  alone is not sufficient.

### 1.2 Gap closure register

| Gap | Required areas | Compatibility decision | Risk if unresolved | Validation criterion |
| --- | --- | --- | --- | --- |
| Contradictory Extensions versions and terminology | Specification, authored contracts, generator tests | Reject obsolete v1; no deployment-config bump | Competing authoritative interpretations | One coherent v3/v2 normative model and current examples |
| Insufficient projection validation | Contracts, generator implementation, fixtures | Generated digests change atomically | Stale or malformed security metadata is admitted | Exact shapes, ordering, schema resolution, and digest parity fail closed |
| Missing characterization | Tests | No product change | Refactor encodes or changes incidental behavior | External discovery, overlay, diagnostics, admission, immutability, and claim truth table covered |
| Dead generic APIs and exported defaults | Implementation, tests, recovery tool | Intentional internal break; no aliases | Dead surfaces become permanent coupling | No references; effective recovery defaults unchanged |
| Duplicated application facade and test API | Implementation, test support | Atomic caller migration | Multiple projections diverge | One defensive deployment projection; no production path-loader helper |
| Unenforced kernel boundary | Boundary policy and tests | Future imports require owner rationale | Owner semantics return to the kernel | Forbidden directions and profile registries rejected statically |
| Static owner DTOs in config | Specification, implementation, tests | Stable external TOML | Every new owner edits central code | Fixture owner registers without a kernel source edit |
| Telemetry misownership | Specification, implementation, verification inputs | Stable telemetry keys; new owner-qualified row IDs | OTel evolves through unrelated kernel changes | No telemetry DTO/policy in config; telemetry and OTel suites pass |
| Hard-coded and untyped claims | Specification, contracts, implementation, tests | Stable six public paths; no arbitrary-lookup shim | Stale metadata or raw requests bypass resolution boundary | Descriptor parity and future-profile tests; distinct request/resolution types |
| Verification/handoff drift | Verification contracts, generated topology, tracker | New test IDs receive a crosswalk, not aliases | Green checks prove the wrong owner | Harness/drift checks and final evidence are complete |

### Owner documents inspected

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` was read first and
  used as planning doctrine, not repository-state evidence.
- `docs/domain.md` and `docs/research/nlspec-spec.md` supplied vocabulary, authority,
  completeness, and owner-mapping rules.
- `docs/spec/00_document_set_status_and_precedence.md` supplied document precedence
  and the Core 04 deployment-configuration allocation.
- `docs/spec/01_architecture_storage_and_view_contracts.md` supplied the runtime-root
  relationship to Core 04. Targeted searches of Core 02 and Core 03 found no direct
  deployment-configuration ownership for this package.
- `docs/spec/04_security_deployment_and_conformance.md` supplied the configuration
  artifact, discovery, overlay, profile, root, limit, claim, error, and startup
  requirements.
- `docs/opentelemetry-instrumentation-nlspec.md` supplied the adopted telemetry
  configuration contract.
- `docs/extension-subsystem-nlspec.md` supplied the adopted claim and inactive
  extension configuration boundary.
- `docs/network-flow-activity-nlspec.md` supplied the adopted Network Flow
  configuration owner boundary.
- `temp/analysis-notes.md` supplied a proposed resolution of the earlier architecture
  questions. It is implementation-planning evidence only and does not override an
  adopted owner document.

### Repository files inspected

- Every file under `internal/platform/config`, listed individually in Section 2.
- `internal/app/configassembly/configuration.go`, `configuration_test.go`,
  `deployment.go`, `platform_settings.go`, and `platform_settings_test.go`.
- `internal/app/extensionassembly/configuration.go` and application composition in
  `internal/app/server`, `internal/app/migrate`, and `internal/app/operator`.
- Configuration callers under `internal/platform/bootstrap`,
  `internal/platform/postgres`, `internal/platform/telemetry`,
  `internal/testutil/configtest`, and `tools/recoverybrowserrestore`.
- `configs/dev/config.toml`, `internal/testutil/fixtures/config/valid.toml`, and
  `internal/testutil/fixtures/config/invalid_missing_required.toml`.
- `tools/backend_module_boundaries.json`,
  `tools/test_catalog_owner.json`, `tools/test_families/platform.config.json`,
  `contracts/verification/owners/platform.config.json`,
  `contracts/verification/registry.json`, and the OTel configuration contracts under
  `contracts/otel`.

At tracker creation, the tracked worktree was clean on `main`, one commit ahead of
`origin/main`. The tracker did not previously exist, so this session establishes its
handoff history.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/platform/config/catalog.go` | Typed source-owner catalog, namespace registration, validation phases, immutable clone/project lifecycle | `Source`, `ValidationPhase`, `ValidationPhases`, `Key`, `NewKey`, `Definition`, `CatalogBuilder`, `Register`, `Catalog`, `IDs` | `configassembly`, target snapshot loader, target and application tests | Standard library | `catalog_test.go`, `config_otel_test.go`, configassembly tests | Referenced by `platform.config` harness rows; no generated file imported | `platform.config` generic engine | High | Central extension seam; changes can affect every contributed owner configuration. |
| `internal/platform/config/snapshot.go` | Admitted immutable snapshot, loading, typed decoding, Boolean claim projection, startup validation | `Snapshot`, `LoadSnapshotWithOptions`, `LoadSnapshotFromTOML`, `Decode`, `SnapshotBooleanValuesAtPaths`, `ValidateSnapshotForStartup`, `Value` | `configassembly`, extension assembly, target tests | Standard library and package-private decoder/validator | `catalog_test.go`, `claim_projection_test.go`, `core_config_test.go`, `inactive_extension_test.go`, `config_otel_test.go` | Claim projection is consumed alongside generated extension descriptors; no generated file imported | `platform.config` generic engine | High | Snapshot immutability, typed lookup, and deterministic projection are frozen behavior. |
| `internal/platform/config/config.go` | Config path discovery, strict TOML decode, `CARTULARY__` overlays, diagnostics, base DTOs, raw inactive-section handling, owner-shaped document sections | Defaults and error constants; `LoadOptions`; diagnostics types/helpers; application, root, bootstrap, claim, timeout, interval, and limit DTOs | `configassembly`; server, migrate, operator; bootstrap, Postgres, telemetry; test utilities; recovery-browser tool | Standard library, `BurntSushi/toml` | `core_config_test.go`, `backup_config_test.go`, `inactive_extension_test.go`, `config_otel_test.go` and downstream application/process tests | `cartulary.deployment_config.v2` fixtures and OTel contracts exercise its behavior; no generated file imported | Core 04 platform config plus contributed owner projections | High | Legitimate facade with mixed generic and owner-shaped responsibilities. |
| `internal/platform/config/validation.go` | Structural and startup validation, root canonicalization/writability/overlap checks, defaults and limits, hard-coded claim lookup, extension runtime validation | Package-internal validation surface used through loaders and `ValidateSnapshotForStartup` | Target loaders and tests; behavior reaches all application roots | Standard library, `x/sys/unix` | Primarily `core_config_test.go`, `backup_config_test.go`, `claim_projection_test.go`, `inactive_extension_test.go` | Harness rows cover roots, limits, claims, runtime defaults, and inactive config | Core 04 platform config, with some source-owner policy candidates | High | Primary coupling hotspot; extraction must retain exact diagnostics and startup ordering. |
| `internal/platform/config/inactive_policy.go` | Narrow port for syntax-only parse/validate/discard of unclaimed extension configuration | `InactivePolicy` | `extensionassembly` constructs it; loaders consume it | Standard library | `inactive_extension_test.go`, `inactive_policy_test_support_test.go` | Extension descriptors inform the application implementation; no generated dependency in this file | `platform.config` port implemented by Extensions/application assembly | Medium | Direction is intentional: config owns artifact handling while the source owner supplies policy. |
| `internal/platform/config/telemetry_document.go` | Private wire-shaped telemetry TOML/environment document DTOs | No exported symbols | Package decoder; telemetry contribution materializes admitted values | Standard library | `config_otel_test.go` | `contracts/otel/telemetry_config_schema.v2.json` and hazard fixtures cover the resulting behavior | `platform.telemetry` contribution or a generic owner-document mechanism | Medium | Permanent placement is deferred; the adopted OTel NLSpec owns telemetry semantics. |
| `internal/platform/config/backup_config_test.go` | Characterizes backup-root presence, binding kind, and overlap rules | Go test `TestBackupStorageRootBindingConfig` | `platform.config` test family | Testing plus target package | Self | `platform.config` backup-root harness row | `platform.config` | Medium | Core 04 deployment-root evidence. |
| `internal/platform/config/catalog_test.go` | Characterizes typed catalog registration, order, clone immutability, type safety, duplicate handling, and phases | Go test `TestCatalogSnapshotLifecycle_Unit` | `platform.config` test family | Testing plus target package | Self | `platform.config` catalog lifecycle row | `platform.config` | High | Required before changing the owner-contribution seam. |
| `internal/platform/config/claim_projection_test.go` | Characterizes requested-only Boolean projection and rejection of unknown/duplicate paths | Two unit tests for `SnapshotBooleanValuesAtPaths` | `platform.config` test family | Testing plus target package | Self | Extension/runtime claim projection harness row | Core 04/Extensions integration at `platform.config` | High | Required parity evidence before replacing hard-coded claim lookup. |
| `internal/platform/config/config_otel_test.go` | Characterizes telemetry defaults, closed namespace, environment containment, validation, endpoint/enums/bounds, secret references, and redaction | Four OTel test groups | `platform.config` rows and OTel conformance evidence | Testing, target config, platform telemetry configuration | Self | OTel schema and hazard matrices cite this evidence | Shared `platform.config` integration and `platform.telemetry` semantics | High | Test ownership may split, but platform integration coverage must remain. |
| `internal/platform/config/core_config_test.go` | Characterizes discovery, overlays, schema/origin, claims, runtime defaults, roots, paths, resource limits, and non-configurable process model | Core configuration test groups and shared test helpers | Multiple `platform.config` harness rows | Testing, target package, test fixtures | Self and downstream configassembly/process tests | `platform.config` owner manifest maps its test groups | Core 04 `platform.config`, with collaborator-owner projections | High | Broad contract-freeze suite; split only after equivalent owner-focused coverage exists. |
| `internal/platform/config/inactive_extension_test.go` | Characterizes syntax-only inactive configuration and environment overlays without admitting or leaking values | `TestInactiveExtensionConfiguration_Unit` | `platform.config` test family | Testing plus target package | Support policy in `inactive_policy_test_support_test.go` | Inactive extension harness row | Core 04/Extensions boundary | High | Protects fail-closed and secret-discard behavior. |
| `internal/platform/config/inactive_policy_test_support_test.go` | Supplies a test-only inactive policy implementation and observed-call evidence | Test-only helpers | `inactive_extension_test.go` | Testing, target package, UUID helper | `inactive_extension_test.go` | No generated artifact | Target test support | Low | Test-only support; no production assumption was found to depend on it. |

`internal/platform/config/extensioninactive` exists as an empty filesystem directory.
It contains no source file or tracked artifact requiring a separate inventory row.

The production target imports no `internal/modules/*` package and contains no direct
SQL, object-storage operation, HTTP route, frontend, or grid-vendor dependency.

### Usage-proven cleanup inventory

The next iteration removes only surfaces whose lack of continuing value was confirmed
by repository-wide Go usage searches. No compatibility alias or deprecated wrapper is
planned because these are repository-internal APIs and all callers are updated in the
same slice.

| Candidate | Live usage evidence | Disposition | Reason |
| --- | --- | --- | --- |
| `ValidationPhase`, its constants and backing list, and `ValidationPhases` | Referenced only by `catalog_test.go`; runtime control flow never consumes the values | remove | A disconnected phase catalog can drift from the real loader and makes future phase growth less safe. Reintroduce a phase API only if runtime execution consumes it. |
| `Key.ID` | No caller found | remove | Exposes an opaque identity without a consumer. |
| `Catalog.IDs` | Referenced only by `catalog_test.go` | remove | Catalog order is already characterized by materialization behavior; a test-only introspection API adds no production value. |
| `Definition.ClaimPath` and `catalogEntry.claimPath` | Two registrations populate it, but no runtime code reads it | remove | The field is inert and is not sufficient for the future typed claim-registration contract. Do not repurpose it. |
| `DefaultConfigPath` | Used only inside `platform.config` and its same-package test | make private | The path behavior remains public through artifact selection; the constant itself has no external caller. |
| Exported limit default constants | Used by `tools/recoverybrowserrestore` as well as target validation/tests | retain | They have a live production-tool consumer. Consolidation requires a separately characterized recovery configuration builder. |
| `configassembly.LoadPath` | Called only by recovery operator process tests | move to `internal/testutil/configtest` | Test fixture construction does not justify a production application-facade method. |
| `Loaded.Revisions`, `Loaded.EnterpriseAuthentication`, and `Loaded.NetworkFlow` | Revisions is test-only; the other two duplicate fields already present in the defensive `Deployment()` projection | remove | One immutable application projection is simpler and eliminates impossible post-admission lookup error branches. |
| Telemetry `ConfigurationKey` | Referenced only inside `internal/platform/telemetry` | make private | The owner contribution constructor and value mapper are the useful package surface. |

The following surfaces have confirmed continuing value and remain: `LoadSnapshotFromTOML`,
`Snapshot.Decode`, `SnapshotBooleanValuesAtPaths`, `Value`, `configassembly.Admit`, the
diagnostic types/helpers, Core 04 DTOs, `ConfigFileEnv`, and
`InvalidDeploymentConfigCode`.

## 3. Module Boundary Diagnosis

The current package is a **legitimate platform configuration facade plus a
mixed-responsibility owner-projection host**. It owns real Core 04 platform behavior:
artifact discovery, strict decoding, overlays, diagnostics, root capability validation,
admission, and startup checks. It is not evidence for a permanent monolithic boundary:
the live application already constructs owner contributions in
`internal/app/configassembly`, while telemetry, enterprise authentication, Network Flow,
and Revisions supply or receive owner-shaped values.

It is not a frontend shell/controller surface, grid integration layer, route handler,
projection layer, persistence adapter, mutation coordinator, or domain catch-all.

The permanent direction is a three-layer design:

1. `platform.config` owns the generic catalog/snapshot kernel, Core 04 base document,
   artifact selection, overlays, diagnostics, admission, and startup-root validation.
2. Source owners own namespace-native types and semantic validation. Where the live
   boundary registry prohibits an owner package from importing deployment config, a
   private application-layer contribution adapter bridges `config.Source` to that
   owner; this does not move owner semantics into application composition.
3. `internal/app/configassembly` statically constructs the exact contribution catalog
   and translates the admitted snapshot into immutable runtime settings.

Registration remains explicit and deterministic. Package initialization, reflection
discovery, mutable global registries, source-tree scanning, and dynamic plugin loading
are prohibited. Domain modules, HTTP handlers, and storage adapters do not consume raw
deployment configuration.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Artifact selection, TOML decoding, environment overlay, and diagnostic envelope | `config.go` | `platform.config` | keep | Core 04 Section 12 and live server/migrate/operator loaders | Preserve path selection, namespace closure, error codes, ordering, and redaction. |
| Typed contribution catalog and immutable admitted snapshot | `catalog.go`, `snapshot.go` | `platform.config` generic engine | keep | `configassembly` registers source-owner contributions through this API | Public typed surface must be frozen before internal rearrangement. |
| Runtime-root and bootstrap capability validation | `config.go`, `validation.go` | `platform.config` under Core 04 | keep | Core 04 root model and target root/startup tests | Storage adapters consume admitted capabilities; config does not perform storage operations. |
| Core 04 deployment defaults and resource limit registry | `config.go`, `validation.go` | `platform.config`, projected to consuming owners by application assembly | split | Core 04 owns the registry; `configassembly`/server project settings to modules | Split mechanics from downstream consumption only; do not relocate normative defaults without owner review. |
| Owner-shaped telemetry wire document | `telemetry_document.go` | `platform.telemetry` behind a generic owner-namespace decoding seam | defer | OTel NLSpec owns semantics; telemetry already builds a catalog contribution | Ownership is resolved, but movement is deferred until strict unknown-key, source, overlay, ordering, and redaction behavior can be preserved without `platform.config` importing telemetry. |
| Enterprise-authentication, Network Flow, and Revisions document sections | `config.go`, then `configassembly` contributions | Respective adopted owners plus application composition | split | Existing owner contribution constructors and projection tests | Retain exact TOML and environment keys while reducing platform-config semantic knowledge. |
| Boolean claim path lookup | `validation.go`, `snapshot.go` | Generic config projection supplied by an Extensions-owned descriptor adapter | defer | Hard-coded switch, extensionassembly request paths, and the adopted Extensions v1/v3 contradiction | Generalize only after the owner repairs the contradiction, regenerates projections, and proves unknown, duplicate, inactive, and closed-set parity. |
| Inactive extension raw parse/validate/discard | Decoder plus `InactivePolicy` supplied by extension assembly | Core 04 artifact handling with Extensions-owned policy | keep | Core 04 and Extensions NLSpec allocation; live port direction | This is a deliberate cross-owner seam, not hidden domain logic. |
| Application catalog construction and runtime settings translation | `internal/app/configassembly`, server application composition | Application assembly | keep | Live code constructs contributions and maps deployment values | Modules and HTTP packages must continue to avoid importing deployment config. |

Planning-framework examples involving timeline, projections, test-util row mutation,
collaboration, imports/tabular ingest, entities/indicators, evidence, links, saved views,
view contracts, frontend controller state, and grid adapters are not present inside this
target. That framework/repository difference is a planning finding, not a reason to
invent those workstreams here. Likewise, the `platform.config` harness owner label is
evidence routing, not proof of the permanent architecture.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Schema ID `cartulary.deployment_config.v2`, default `/etc/cartulary/config.toml`, absolute `CARTULARY_CONFIG_FILE`, strict TOML namespace, and `CARTULARY__` overlay grammar | Core 04 / `platform.config` | Core 04 Section 12, fixtures, `config.go` | Config discovery and OTel parser tests | Preserve file-not-found, relative override, empty overlay, unknown path, and type-conversion cases | High | Wire keys and precedence are frozen. |
| Diagnostic envelope and deterministic ordering | Core 04 / `platform.config` | `Diagnostic`, `DiagnosticsError`, loader and validator behavior | Catalog, core config, inactive config, and OTel tests | Add parity assertions at any new internal boundary, including secret values | High | Code, path, source, message safety, ordering, and redaction must not drift. |
| Deployment profiles and public origin | Core 04 | Core 04 profile/origin requirements and config fixtures | Core config and server tests | Preserve profile-specific root rules and origin validation | High | `public_origin` indirectly affects WebSocket/HTTP origin enforcement but route shapes are not owned here. |
| Runtime roots, bootstrap, backup root, canonicalization, writability, overlap, and startup fail-closed behavior | Core 04 / `platform.config` | `validation.go`, fixtures, bootstrap/Postgres/server consumers | Core config, backup, serverprocess, migrate, operator, recovery tests | Characterize ordering when multiple root/startup failures coexist | High | Storage semantics remain in adapters; this contract admits capabilities only. |
| Catalog registration, validation phases, immutable snapshots, typed decode, and clone behavior | `platform.config` generic engine | `catalog.go`, `snapshot.go`, configassembly | Catalog lifecycle tests | Add caller-boundary parity before moving types or constructors | High | Exported Go API is a public package surface within the repository. |
| Boolean claims, unknown/duplicate paths, inactive config parsing, and parse/validate/discard semantics | Core 04 plus Extensions owner | `snapshot.go`, `validation.go`, `InactivePolicy`, extensionassembly | Claim projection and inactive-extension tests | Cover every adopted claim path and owner manifest combination before generalizing | High | Do not infer claims from filenames or historical phases. |
| Telemetry defaults, closed namespace, environment containment, validation, endpoints, headers, and secret references | Adopted OTel NLSpec / `platform.telemetry`; integrated by `platform.config` | OTel NLSpec and `contracts/otel` | Four OTel target test groups and `make otel-conformance` rows | Preserve TOML/environment parity and redaction across any ownership move | High | Test/accounting ownership may change without changing runtime behavior. |
| Enterprise-authentication claim and provider manifest projection | Core 04 and authentication owner | Core config and configassembly contribution | Enterprise-auth target and projection tests; auth integration test | Preserve unclaimed-manifest rejection and admitted owner value | High | No authorization decision is performed inside this package. |
| Network Flow claim and key-ring manifest projection | Network Flow NLSpec plus Core 04 generic claim rules | Core config, contribution constructor, Network Flow integration | Target config/projection and Network Flow tests | Preserve claimed/unclaimed key-ring behavior | High | Owner-specific semantic validation stays with Network Flow. |
| Revisions conflict-token key-ring path projection | Core 04/Revisions allocation | Config fixture and revisions contribution | Target configassembly and downstream revisions/server tests | Preserve absolute-path and startup-readiness behavior | High | Revision/change-set semantics are not owned by config. |
| Resource limit and extension timeout/interval/limit projections | Core 04 registry, consuming module owners | DTO/default validation and `platform_settings.go` | Resource-limit, extension-runtime, and settings projection tests | Preserve every default, bound, unit, and consumer mapping | High | Moving a value must not alter its normative owner or application mapping. |
| Generated protocol/view contracts | Their respective contract owners | No target production import or route/view-schema surface found | Not applicable | None unless a later slice unexpectedly reaches those surfaces | Low | Hand-editing generated outputs is prohibited. |
| Harness and test accounting | Verification owners and harness manifests | `platform.config` verification and test-family JSON | 19 Go rows: 18 unit and 1 support-unit; no service-backed row | Update authored owner inputs only if tests move, then regenerate | Medium | Accounting is evidence routing, not architecture. |

Removing repository-internal Go surfaces listed in the cleanup inventory is not an
observable configuration compatibility change. The implementation iteration must still
update every repository caller atomically and must not retain aliases, deprecated
wrappers, dual registries, or fallback paths. The deployment schema, TOML/environment
keys, defaults, diagnostic fields and ordering, snapshot values, runtime settings, and
startup outcomes remain frozen.

No direct HTTP route, WebSocket event format, entity row/query/mutation contract,
saved-view/view-schema behavior, projection refresh, frontend selector, or grid-adapter
contract is implemented by this package. Later changes can still affect service startup
availability or origin authorization indirectly, so those outcomes remain frozen.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Runtime-unused phase metadata presents itself as a normative lifecycle | Phase constants/list are read only by their same-package test | It can drift from real control flow and mislead future phase additions | `must_fix` | Actual kernel control flow plus behavior tests | Delete the metadata; add a future phase API only with an executable consumer and owner contract. |
| Catalog identity accessors and claim metadata have no consumer | `Key.ID` has no callers; `Catalog.IDs` is test-only; `ClaimPath` is stored but never read | Dead public surface creates compatibility burden and invites accidental reuse | `must_fix` | `platform.config` | Remove atomically with registrations/tests and do not add shims. |
| Test fixture loading and owner-value accessors enlarge production application APIs | `LoadPath` is process-test-only; three `Loaded` accessors duplicate `Deployment()` or are test-only | Extra paths allow application projections to diverge | `must_fix` | `internal/testutil/configtest` for fixture loading; `configassembly.Deployment` for admitted values | Move explicit-path test loading and contract the `Loaded` facade. |
| `DefaultConfigPath` and telemetry `ConfigurationKey` are exported without external callers | Repository-wide qualified-symbol searches | Low but permanent package-surface noise | `should_fix` | Their current packages | Make them private in the cleanup slice. |
| Generic configuration mechanics share files with owner-shaped document DTOs and policy | `config.go`, `validation.go`, `telemetry_document.go`; existing configassembly contributions | Ownership changes can alter wire decode or diagnostics | `should_fix` | Generic engine in `platform.config`; semantics in adopted source owners | Characterize, then design a split that retains exact artifact shape and error behavior. |
| Boolean claim lookup duplicates a closed list of known claim paths | Hard-coded lookup in `validation.go`; paths requested by extensionassembly | New or moved claims can drift between decoder, descriptors, and projections | `must_fix` | Core 04 generic config plus Extensions-owned descriptor adapter | Defer removal while RB-003 is `BLOCKED: owner contradiction`; then prove full parity before cutover. |
| Telemetry semantics are owner-supplied, but wire DTOs and most configuration tests remain under platform config | OTel contribution in platform telemetry and `config_otel_test.go` | Test movement could lose integration or secret-redaction evidence | `should_fix` | Shared platform-config integration and platform-telemetry owner coverage | Ownership is resolved; defer movement until the generic decoder seam and test-row crosswalk are decision complete. |
| Application assembly constructs owner providers and translates admitted deployment values | `internal/app/configassembly` and server module settings | Moving composition into config would reverse the intended boundary | `intentional/no_action` | Application assembly | Preserve this direction and add boundary checks to slice validation. |
| Exported limit defaults have a live recovery-tool caller | `tools/recoverybrowserrestore/main.go` constructs an explicit disconnected deployment | Removing them now would force duplicated literals or an uncharacterized builder | `intentional/no_action` | Core 04 `platform.config` | Retain for this iteration; reconsider only with a tested recovery configuration factory. |
| Diagnostics are shared by bootstrap, Postgres readiness, telemetry, and application roots | Callers use `Diagnostic` and `NewDiagnosticsError` | A premature owner split could fragment the Core 04 startup error envelope | `intentional/no_action` | `platform.config` | Freeze the shared envelope; refactor internals behind it. |
| Production target has no domain-module import, direct SQL/storage call, HTTP handler, frontend/grid dependency, or generated-file read | Production imports and repository searches | Inventing a move would create unsupported scope | `intentional/no_action` | Existing platform/application owners | Record absence and do not add speculative workstreams. |
| Test-only inactive-policy support does not leak into production | `_test.go` helper is referenced only by target tests | Low | `intentional/no_action` | Target test support | Retain until tests are reorganized; do not create a production helper. |
| Framework candidate boundaries do not match the live target | Direct target inventory versus generic framework examples | Treating examples as repository facts would mis-scope the refactor | `intentional/no_action` | Tracker planning process | Use live evidence and explicitly mark the mismatch. |

No direct grid-vendor coupling, duplicated saved-view/view-schema logic, hidden
mutation/revision/projection side effect, authorization decision in the wrong layer,
direct persistence coupling, or hand-edited generated drift was found in the target.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 | Establish authority, clean state, scope, and single-file write boundary | This tracker and inspected owner documents | `git status --short --branch`; tracker existence check | Source posture and session log established. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-04 | Account for every target file, caller, dependency, and current test | `internal/platform/config/**` and direct callers | Repository searches and exact source reads | Section 2 complete with no generic rows. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04 | Map every observable configuration contract to its normative and implementation owner | Core 04, adopted subsystem NLSpecs, configassembly, contracts | Owner-document and live-code comparison | Freeze map has an owner and evidence posture for each contract. |
| WF-03 | Characterization test gap analysis | parallel | WF-02 | WF-05 | Identify missing parity coverage without changing behavior | Target tests, configassembly tests, caller tests, test catalog | `make task-guide ROLE=module-author OWNER=platform.config`; `make explain-test-owner OWNER=platform.config` | Test gaps and affected collaborator owners are enumerated. |
| WF-04 | Boundary and coupling scan | parallel | WF-01, WF-02 | WF-05 | Separate intentional platform seams from misplaced owner semantics | Target production files, configassembly, boundary registry | `make backend-module-boundary-check` when implementation begins | Every finding is classified and assigned a planning action. |
| WF-05 | Facade and ownership redesign plan | chain | WF-03, WF-04 | WF-06 | Establish the three-layer boundary and separate the cleanup iteration from later telemetry/claim redesign | Target, configassembly, test support, telemetry, boundary registry | Live usage proof and owner review | RB-001 and RB-002 are resolved; RB-003 is explicitly blocked without blocking dead-surface cleanup. |
| WF-06 | Legacy-removal slice sequencing | chain | WF-05 | WF-07 | Remove only proven-unused surfaces through independently reversible behavior-preserving slices | Target, configassembly, server runtime assembly, config test support | Per-slice commands in Section 7 | Every removal has live usage evidence, an atomic caller update, and no compatibility shim. |
| WF-07 | Boundary and test-accounting confirmation | chain | WF-06 | WF-08 | Harden dependency direction and prove unchanged test names/postconditions require no harness remap | Backend boundary owner and existing test-family inputs | `make backend-module-boundary-check`; harness inspection | New boundary rule passes and all existing rows remain valid without authored harness changes. |
| WF-08 | Validation and final handoff | chain | WF-07 | None | Run narrow-to-broad verification and publish continuation evidence | All authorized implementation files and this tracker | `make agent-finalize`, owner slices, `make test-fast`, risk-based `make check` | Commands, results, run roots, blockers, and next action are current. |

## 7. Proposed Refactor Slice Plan

These slices replace the earlier broad sequencing plan for the next implementation
iteration. They are not authorized by this tracker-update task. Every slice preserves
observable behavior, removes only a usage-proven burden, and updates all repository
callers atomically. No compatibility aliases, deprecated wrappers, fallback registries,
or dual behavior paths are retained.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | WF-03 complete | Complete characterization needed to replace dead introspection assertions with behavioral assertions for catalog order, snapshots, diagnostics, owner projections, defaults, claims, and startup | Target and configassembly tests only | Removing a test-only getter can accidentally remove the behavior it was indirectly checking | Preserve all 19 `platform.config` rows; assert materialization order and defensive projections directly | `make test-slice OWNER=platform.config` | Revert only the new/reworked characterization assertions | The owner slice passes and every removed surface is either behavior-free or covered through its real postcondition. |
| S-01 | S-00 | Remove `ValidationPhase` metadata, `ValidationPhases`, `Key.ID`, `Catalog.IDs`, `Definition.ClaimPath`, `catalogEntry.claimPath`; make `DefaultConfigPath` private | `internal/platform/config` and its catalog/core tests | Catalog order, registration rejection, or artifact-path behavior could drift during mechanical cleanup | Catalog build/materialization order, duplicate/overlap checks, snapshot immutability, default path | `make test-slice OWNER=platform.config` | Revert this catalog/config surface-contraction slice as one unit | All dead symbols are absent, registrations compile without `ClaimPath`, real behavior tests pass, and no replacement metadata API exists. |
| S-02 | S-01 | Move explicit-path process-test loading behind `internal/testutil/configtest`; remove `configassembly.LoadPath` and redundant `Loaded` owner accessors; make telemetry `ConfigurationKey` private; have runtime assembly consume one defensive `Deployment()` projection | Configassembly, config test support, server runtime assembly, recovery operator process tests, platform telemetry | Owner normalization, defensive-copy behavior, server startup, or process fixtures could change | Rework configassembly immutability tests around repeated `Deployment()` calls; preserve server and operator/recovery process tests | `make test-slice OWNER=platform.config`; `make test-slice OWNER=app.server`; `make service-backed-test-slice OWNER=app.server`; `make service-backed-test-slice OWNER=module.recovery`; `make service-backed-test-slice OWNER=app.operator` | Revert facade contraction and its caller updates together | Removed methods have no callers, process fixtures load identical configuration, runtime receives identical owner values, and all affected owner slices pass. |
| S-03 | S-02 | Add an authored backend boundary preventing `internal/platform/config/**` from importing application packages, domain modules, or owner runtime packages; keep static catalog construction | Backend module-boundary owner and its tests/manifest inputs as required by the existing boundary mechanism | An overbroad rule could prohibit intentional standard/platform primitives; an underbroad rule permits kernel creep | Existing boundary fixtures plus positive coverage for current allowed imports and negative coverage for prohibited directions | `make backend-module-boundary-check`; `make test-slice OWNER=platform.config` | Revert the single authored rule and its boundary tests | The rule rejects each prohibited direction, accepts the current kernel imports, and introduces no allowlist exception for application/owner runtimes. |
| S-04 | S-03 | Remove imports, comments, test branches, and helper code made obsolete by S-01 through S-03; run final narrow-to-broad verification and update this handoff | Authorized cleanup files and tracker | Cleanup can mask a behavior change or broaden scope into telemetry/claim redesign | All preceding suites; unchanged harness row names and postconditions | `make agent-finalize`; `make test-fast`; risk-based `make check` | Revert only the smallest failing cleanup slice; do not fold telemetry or claim work into the repair | Required checks pass, only planned surfaces were removed, no compatibility shim remains, and telemetry/claim redesign is still deferred. |

The next iteration stops at S-04. Moving `telemetry_document.go` requires a later,
decision-complete owner-namespace decoder design and test-row crosswalk. Removing the
hard-coded Boolean claim switch requires an adopted repair of RB-003, regenerated
projections, and the complete claim parity matrix. Neither later slice may be folded
into cleanup work.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | This tracker | yes | Required for this planning-only write. It does not validate product behavior. |
| unit | `make test-slice OWNER=platform.config` | Current 19-row owner slice | yes | Baseline passed 19/19 on 2026-08-07; rerun after every implementation slice. |
| application unit | `make test-slice OWNER=app.server` | Server application composition after S-02 | no | Required for S-02 and final validation. |
| integration | `make service-backed-test-slice OWNER=app.server`; `make service-backed-test-slice OWNER=module.recovery`; `make service-backed-test-slice OWNER=app.operator` | Runtime projection and explicit-path process fixtures affected by S-02 | no | Required after S-02. Current `platform.config` itself has no service-backed row. |
| e2e/browser | Not applicable for the current target | No direct browser/frontend contract | no | Discover a browser target only if a later slice changes server startup or externally observable availability. |
| OTel conformance | `make otel-conformance` | Adopted telemetry configuration and instrumentation evidence | no | Deferred with telemetry extraction; making its private key private is covered by compile/unit checks. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Generated projections and policy | no | Required after any authorized authored contract/harness-input change; never hand-edit generated outputs. |
| import-boundary/static | `make backend-module-boundary-check` | Existing and new backend dependency direction | yes | Baseline passed 3/3 on 2026-08-07; required after S-03. |
| security | `make go-gosec-targeted` | Changed Go security surface | no | Required for a security-sensitive implementation slice. |
| fast broad check | `make test-fast` | Broad non-service verification | no | Run after the narrow owner slice for cross-caller risk. |
| full check | `make check` | Full repository gate | no | Run after `make agent-finalize` when implementation risk warrants it; report the run root. |

The 2026-08-07 baseline executions are product validation of the pre-change state only:

- `make test-slice OWNER=platform.config` passed 19/19 with run root
  `.cartulary/test-results/20260807T155629Z-p3318358`.
- `make backend-module-boundary-check` passed 3/3 with run root
  `.cartulary/test-results/20260807T155636Z-p3320189`.

The tracker-only `make lint-markdown` validation passed with run root
`.cartulary/test-results/20260807T161351Z-p3326350`; it does not validate a production
refactor.

Other command-discovery invocations recorded in this tracker are not test executions and
must not be reported as product validation.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| TR-001 | Establish source posture and planning boundary | WF-00 | DONE | None | Section 1 and initial session log | Authority, allowed write, and non-goals are explicit. |
| TR-002 | Inventory every target file and direct dependency | WF-01 | DONE | TR-001 | Section 2 | Every target file is represented; empty directory is noted. |
| TR-003 | Map observable contracts and owner posture | WF-02 | DONE | TR-002 | Sections 3 and 4 | Each discovered contract has an owner, evidence, and test posture. |
| TR-004 | Classify current coupling findings | WF-04 | DONE | TR-002, TR-003 | Section 5 | Every supported finding uses an allowed classification. |
| TR-005 | Complete cleanup-specific characterization | WF-03 | DONE | TR-003 | WS-03 characterization and current platform-config 15/15 evidence | Behavioral assertions replace all reliance on dead introspection surfaces. |
| TR-006 | Approve generic-kernel/application/owner design | WF-05 | DONE | TR-004 | Section 3 and resolved RB-001/RB-002 decisions | Three-layer direction and package-boundary realization are explicit. |
| TR-007 | Implement legacy-surface removal S-00 through S-04 | WF-06 | DONE | TR-005, TR-006 | Superseded and completed by authorized WS-03 through WS-07 | Removed internal surfaces have no aliases or remaining callers. |
| TR-008 | Reassign harness accounting if postconditions or test names move | WF-07 | DONE | TR-007 | WS-08 crosswalk and regenerated authored topology | Telemetry evidence has current immutable owner-qualified IDs. |
| TR-009 | Run final implementation validation and handoff | WF-08 | DONE | TR-007, TR-008 as applicable | Section 12 final run matrix | Product, release, and final Markdown gates pass. |
| TR-010 | Change observable configuration behavior | Separate authorization | DROPPED | None | Contract freeze map | Excluded from this behavior-preserving refactor plan. |
| TR-011 | Extract telemetry wire DTO and split test ownership | WS-07, WS-08 | DONE | TR-009, RB-002 prerequisites | Owner namespace seam, WS-08 crosswalk, telemetry/OTel evidence | No telemetry DTO or semantic policy remains in config. |
| TR-012 | Replace hard-coded claim projection | WS-01, WS-02, WS-09 | DONE | RB-003 | Repaired owner NLSpec, regenerated projections, exact claim parity tests | Descriptor-derived typed request flow is admitted without arbitrary path lookup. |

## 10. Session Handoff Log

### WS-08 telemetry verification ownership crosswalk

The prior IDs are retired rather than retained as aliases. The new rows execute from
`internal/platform/telemetry/configuration_admission_test.go`, remain collaborative with
the generic parser through `platform.config`, and are owned by the telemetry source
owner.

| Retired row and verification ID | Current row and verification ID |
| --- | --- |
| `platform.config.unit.telemetry_defaults_closed_namespace`; `platform.config.verification.telemetry_defaults_namespace` | `platform.telemetry.unit.configuration_defaults_closed_namespace`; `platform.telemetry.verification.configuration_defaults_namespace` |
| `platform.config.unit.telemetry_environment_binding_parser`; `platform.config.verification.telemetry_environment_parser` | `platform.telemetry.unit.configuration_environment_binding`; `platform.telemetry.verification.configuration_environment_parser` |
| `platform.config.unit.telemetry_secret_references`; `platform.config.verification.telemetry_secret_references` | `platform.telemetry.unit.configuration_secret_references`; `platform.telemetry.verification.configuration_secret_references` |
| `platform.config.unit.telemetry_validation`; `platform.config.verification.telemetry_validation` | `platform.telemetry.unit.configuration_validation`; `platform.telemetry.verification.configuration_validation` |

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 14:50 EDT | Codex / WS-11 | Prior WS-00 through WS-10 evidence is retained; a new production-readiness cleanup iteration is rebaselined from current usage and caller evidence | Inspected the tracker, every current `internal/platform/config` production file, configassembly, extensionassembly, telemetry configuration, server/migrate/operator composition, test-support loaders, verification ownership, and backend boundary policy; touched only this tracker | `git status --short --branch`; exact source and owner-document reads; repository-wide symbol and caller searches; `git diff --check`; `make lint-markdown` | PASS: only this tracker is modified; initial and post-status Markdown lint run roots `.cartulary/test-results/20260807T185329Z-p348875` and `.cartulary/test-results/20260807T185415Z-p350661`; six cleanup gaps are decision-complete without an external deployment-config change; WS-12 through WS-17 remain implementation-planning rows only | RB-005: later implementation authorization | WS-12 is the sole ready slice after separate authorization. |
| 2026-08-07 14:10 EDT | Codex / WS-10 | Harness ownership and generated topology are reconciled; every required narrow, service-backed, stateful, security, broad, full, release, and tracker gate is green; final source audits find no removed claim API, production profile registry, retired telemetry owner ID, whitespace error, or `docs/domain.md` change | Reconciled the controlling tracker and final worktree; no product source changed after the successful WS-09 gate | `make agent-finalize`; `make harness-contract`; generation/artifact/JSON/boundary checks; platform.telemetry owner and OTel conformance; retained WS-09 focused owner/service-backed/stateful evidence; `make test-fast`; `make go-gosec-targeted`; `make check`; `make release-check`; `git diff --check`; exact final source/ownership/domain searches; `make lint-markdown` | PASS: finalize 1/1 `.cartulary/test-results/20260807T175350Z-p4081227`; harness 2/2 `.cartulary/test-results/20260807T175406Z-p4083877`; drift 4/4 `.cartulary/test-results/20260807T175413Z-p4084256`; artifact policy 3/3 `.cartulary/test-results/20260807T175426Z-p4086984`; JSON shape 3/3 `.cartulary/test-results/20260807T175429Z-p4087431`; boundary 3/3 `.cartulary/test-results/20260807T175439Z-p4087933`; OTel 14/14 `.cartulary/test-results/20260807T175456Z-p4088274`; telemetry 10/10 `.cartulary/test-results/20260807T175511Z-p4099748`; fast 349/349 `.cartulary/test-results/20260807T175533Z-p4100146`; gosec 4/4 `.cartulary/test-results/20260807T175747Z-p4167348`; check 722/722 `.cartulary/test-results/20260807T175802Z-p4193493`; release 893/893 `.cartulary/test-results/20260807T180201Z-p146336`; tracker lint `.cartulary/test-results/20260807T181236Z-p335003`. Finalize retained-run checks were intentionally skipped with `results-dir-not-provided` because no successful full warm check root existed before broad verification. | None; no unrelated exception is needed. | Complete. |
| 2026-08-07 13:52 EDT | Codex / WS-09 | One generated, immutable Extensions configuration policy now supplies exact claim registrations and inactive-key behavior; the kernel collects only registered Booleans; `extensionassembly.RequestedClaims`, coordinator resolution, and resolved identity are distinct types; no arbitrary Boolean/path API remains | Added the generic kernel claim-policy seam and future-profile/malformed-catalog/truth-table tests; rewrote extensionassembly policy/request composition; moved claim-only wire values out of config; migrated configassembly, server, migrate, operator, test support, and authored test-family selectors; regenerated harness outputs | `make format`; `make generate`; owner unit/service-backed slices for platform.config, module.extensions, module.networkflow, app.server, and app.operator; `make browser-e2e-stateful`; `make backend-module-boundary-check`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make build-migrate`; exact removed-symbol/raw-map/hard-coded-path searches; `make lint-markdown` | PASS: format `.cartulary/test-results/20260807T174726Z-p4013307`; generate `.cartulary/test-results/20260807T174648Z-p4010264`; config 15/15 `.cartulary/test-results/20260807T175144Z-p4075770`; Extensions unit 37/37 `.cartulary/test-results/20260807T174214Z-p3875258` and service-backed 24/24 `.cartulary/test-results/20260807T174252Z-p3902644`; Network Flow 71/71 `.cartulary/test-results/20260807T174322Z-p3925490`; server unit 42/42 `.cartulary/test-results/20260807T174951Z-p4042969` and service-backed 33/33 `.cartulary/test-results/20260807T174448Z-p3963442`; operator 17/17 `.cartulary/test-results/20260807T174601Z-p3988536`; browser 36/36 `.cartulary/test-results/20260807T174752Z-p4016912`; boundary 3/3 `.cartulary/test-results/20260807T174737Z-p4016441`; drift 4/4 `.cartulary/test-results/20260807T175044Z-p4070126`; artifact policy 3/3 `.cartulary/test-results/20260807T175054Z-p4072855`; JSON shape 3/3 `.cartulary/test-results/20260807T175057Z-p4073286`; migrate build `.cartulary/test-results/20260807T175105Z-p4073842`; tracker lint `.cartulary/test-results/20260807T175232Z-p4077806`. Initial format checks failed on stale/unsorted authored test selectors; config runs `.cartulary/test-results/20260807T173756Z-p3825816` and `.cartulary/test-results/20260807T173915Z-p3834433` failed on related caller/test-policy migration; boundary run `.cartulary/test-results/20260807T174700Z-p4012634` correctly rejected the broad Extensions module import. Each related cause was structurally repaired before the passing evidence. | None. External deployment v2 paths and request truth table are unchanged; internal Boolean lookup and inactive-only policy APIs are intentionally removed without shims. | Run WS-10 final validation and complete the handoff. |
| 2026-08-07 13:26 EDT | Codex / WS-08 | Telemetry wire/semantic configuration evidence is physically and contractually owned by platform.telemetry; the kernel retains only generic admission collaboration | Moved telemetry configuration tests to an external telemetry test package; reassigned four authored rows and four verification contracts; regenerated harness projections; added the immutable crosswalk above | `make format`; `make generate`; platform.telemetry and platform.config owner slices; `make otel-conformance`; `make generate-drift`; `make harness-contract`; exact retired-ID and kernel telemetry searches; `make lint-markdown` | PASS: generate `.cartulary/test-results/20260807T172426Z-p3791104`; telemetry 10/10 `.cartulary/test-results/20260807T172538Z-p3804410`; config 15/15 `.cartulary/test-results/20260807T172550Z-p3805815`; OTel 14/14 `.cartulary/test-results/20260807T172551Z-p3805998`; drift 4/4 `.cartulary/test-results/20260807T172602Z-p3811295`; harness 2/2 `.cartulary/test-results/20260807T172602Z-p3811447`; tracker lint `.cartulary/test-results/20260807T172711Z-p3814843`. Initial telemetry slice `.cartulary/test-results/20260807T172439Z-p3793444` failed on two remaining private-helper calls after the test move; those calls were migrated to the public application admission path before the passing run. | None | Build descriptor-derived claim registration and typed requested-claim boundaries in WS-09. |
| 2026-08-07 13:19 EDT | Codex / WS-07 | Four source-owned namespaces use explicit decoder, closed-path, overlay, projection/validation, and clone callbacks; snapshots retain opaque typed namespaces; the kernel telemetry wire mirror and owner-specific clone are deleted | Updated config catalog/loader/snapshot and tests; configassembly; telemetry contribution; enterprise-auth, Network Flow, and Revisions overlay owners/tests; deleted `telemetry_document.go` | `make format`; platform.config and platform.telemetry owner slices; focused module.auth, module.networkflow, module.revisions, and app.server rows; `make backend-module-boundary-check`; production owner-identifier search; `make lint-markdown` | PASS: config 19/19 `.cartulary/test-results/20260807T171814Z-p3774363`; telemetry 6/6 `.cartulary/test-results/20260807T171705Z-p3766838`; auth `.cartulary/test-results/20260807T171834Z-p3779981`; Network Flow `.cartulary/test-results/20260807T171834Z-p3779973`; Revisions `.cartulary/test-results/20260807T171834Z-p3779986`; server startup `.cartulary/test-results/20260807T171656Z-p3766222`; boundary 3/3 `.cartulary/test-results/20260807T171843Z-p3782314`; tracker lint `.cartulary/test-results/20260807T171931Z-p3782981`. The first config run `.cartulary/test-results/20260807T171002Z-p3641349` failed on a related test import cycle; the next config run `.cartulary/test-results/20260807T171129Z-p3649824` and concurrent broader owner runs failed on related scoped-TOML-tag lookup; both causes were repaired and superseded by the passing evidence. | None | Move telemetry semantic tests and immutable row ownership to platform.telemetry in WS-08. |
| 2026-08-07 13:01 EDT | Codex / WS-06 | Platform config has an authored production import allowlist and a fail-closed guard against owner/profile path registries; hard-coded Boolean switching was replaced by generic typed path resolution | Updated backend boundary policy/checker fixtures and the kernel Boolean field resolver | `make backend-module-boundary-check`; `make test-slice OWNER=platform.config`; `make lint-markdown` | PASS: boundary 3/3 `.cartulary/test-results/20260807T170115Z-p3626189`; config 19/19 `.cartulary/test-results/20260807T170121Z-p3626568`; tracker lint `.cartulary/test-results/20260807T170201Z-p3631648` | None | Implement explicit owner-supplied namespace ownership and generic snapshot storage in WS-07. |
| 2026-08-07 12:57 EDT | Codex / WS-05 | `Loaded.Deployment()` is the sole owner projection; test-only path selection moved to appsupport; telemetry key is private; server consumes the one projection | Updated configassembly API/tests, server assembly, telemetry contribution key, recovery operator tests, and added appsupport configuration loading | Unit slices for platform.config, app.server, app.operator, platform.postgres; service-backed slices for app.server, app.operator, module.recovery | PASS unit: config 19/19 `.cartulary/test-results/20260807T165529Z-p3476027`, server 42/42 `.cartulary/test-results/20260807T165529Z-p3476025`, operator 17/17 `.cartulary/test-results/20260807T165529Z-p3476049`, Postgres/migrate 9/9 `.cartulary/test-results/20260807T165529Z-p3476038`. PASS service: server 33/33 `.cartulary/test-results/20260807T165639Z-p3542588`, operator 9/9 `.cartulary/test-results/20260807T165639Z-p3542601`, recovery 25/25 `.cartulary/test-results/20260807T165639Z-p3542587` | None | Add the permanent platform-config boundary rule in WS-06. |
| 2026-08-07 12:52 EDT | Codex / WS-04 | Dead lifecycle/catalog identities and inert claim metadata are removed; config path and limit defaults are private; recovery tooling relies on ordinary admission defaults | Updated platform config catalog/defaults/tests, configassembly registrations, and recovery browser/restore target construction | Repository-wide removed-symbol search; `make test-slice OWNER=platform.config`; `make test-slice OWNER=module.recovery` | PASS: no removed symbol remains; config 19/19 `.cartulary/test-results/20260807T165151Z-p3422982`; recovery 37/37 `.cartulary/test-results/20260807T165151Z-p3422975` | None | Contract application facade and test-only path loading in WS-05. |
| 2026-08-07 12:49 EDT | Codex / WS-03 | External claim truth-table, in-memory limit defaulting, and defensive deployment projection behavior are characterized without asserting private layout | Added platform-config/configassembly tests and mapped them into the existing platform.config verification rows; regenerated harness projections | `make generate`; `make test-slice OWNER=platform.config`; `make test-slice OWNER=module.extensions` | PASS: generate `.cartulary/test-results/20260807T164758Z-p3391495`, config 19/19 `.cartulary/test-results/20260807T164808Z-p3393805`, Extensions 37/37 `.cartulary/test-results/20260807T164817Z-p3395625` | None | Remove usage-proven dead kernel surfaces in WS-04. |
| 2026-08-07 12:45 EDT | Codex / WS-02 | Current configuration contracts have exact closed rows, resolvable value schemas, canonical dependency inputs, and digest-bound claim facts; generated extension bytes were refreshed from authored inputs | Updated extension dependencies, Core 04 claim fragment, enterprise-auth configuration contract, contract generator validation/tests, and generated contractextensions bytes | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make test-slice OWNER=module.extensions` | PASS: generate `.cartulary/test-results/20260807T164359Z-p3355367`, drift `.cartulary/test-results/20260807T164420Z-p3357844`, policy `.cartulary/test-results/20260807T164431Z-p3360621`, shape `.cartulary/test-results/20260807T164431Z-p3360616`, Extensions 37/37 `.cartulary/test-results/20260807T164440Z-p3361449`. Earlier generate runs failed on a new helper-name collision, empty-array validator choice, and an existing unsorted dependency pair; each related failure was repaired before the passing run. | None | Add contract-focused characterization in WS-03. |
| 2026-08-07 12:36 EDT | Codex / WS-01 | Specification authority is coherent on the declared v3/v2 projection and the generic-kernel, telemetry-owner, static-composition, and requested/resolved boundaries | Updated Extensions NLSpec 0.7.2 and Core 01/Core 04/OTel owner clauses; `docs/domain.md` unchanged because no domain term emerged | Exact normative searches; `make lint-markdown` | PASS, run root `.cartulary/test-results/20260807T163557Z-p3341552`; no superseded-family v1 or `*_schema_ref` references remain in the affected clauses | None | Harden authored projections and generator validation in WS-02. |
| 2026-08-07 12:32 EDT | Codex / WS-00 | Authorized tracker rebaseline complete; external deployment v2 is frozen while internal compatibility shims are prohibited | Updated this tracker; preserved its staged discovery history | `git status --short --branch`; exact tracker reads; `make lint-markdown` | PASS, run root `.cartulary/test-results/20260807T163227Z-p3337104`; WS-00 is complete | None | Normalize adopted owner specifications in WS-01. |
| 2026-08-07 11:13 EDT | Codex / tracker creation | Planning tracker initialized; implementation remains unauthorized | Inspected framework, domain/NLSpec guidance, Core 00/01/04 and targeted Core 02/03; touched only this tracker | `git status --short --branch`; target/tracker existence checks; exact document reads | Target exists, tracker was new, tracked worktree was initially clean on `main` ahead of `origin/main` by one | None for tracker creation | Validate this file with `make lint-markdown`. |
| 2026-08-07 12:09 EDT | Codex / legacy-removal planning | Tracker updated with a bounded cleanup iteration; production remains unauthorized | Inspected `analysis-notes.md`, NLSpec grounding, live owner documents, target/caller usage; touched only this tracker | `git status`; exact reads; repository-wide `rg` usage searches; `make lint-markdown` | Three-layer direction adopted as implementation planning; tracker lint passed; evidence notes remain subordinate to owners | RB-003 and later implementation authorization | Authorize S-00 separately. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 11:13 EDT | Codex / tracker creation | Legitimate platform facade with mixed owner-projection responsibilities | All target files; configassembly; extensionassembly; server, migrate, operator, bootstrap, Postgres, telemetry, testutil, recovery-browser callers; touched only tracker | `rg --files`; targeted `rg`; `sed` exact reads; exported-symbol scan | No production domain import, route, SQL/storage operation, or grid/frontend dependency found | RB-001 | Complete characterization, then approve generic-engine/application-owner division. |
| 2026-08-07 12:09 EDT | Codex / legacy-removal planning | Dead catalog metadata and redundant application/test surfaces are usage-proven; three-layer boundary is decision complete | Catalog/snapshot/config sources, configassembly, server runtime assembly, telemetry configuration, configtest, recovery process tests, backend boundary registry; touched only tracker | Qualified-symbol and caller searches; `make backend-module-boundary-check` | Boundary baseline passed 3/3; RB-001 resolved; removal list and retained surfaces are explicit | Later implementation authorization | Execute S-00/S-01 atomically before application-facade contraction. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 11:13 EDT | Codex / tracker creation | Not applicable to the current target | Repository import/caller searches; touched only tracker | Targeted repository `rg` searches | No frontend import, shell/controller state, selector, view contract, or grid-vendor surface is implemented here | None | Re-open only if a later startup change reaches an observable browser contract. |
| 2026-08-07 12:09 EDT | Codex / legacy-removal planning | Still not applicable; next iteration changes no route, browser, selector, or grid contract | Rechecked target/caller scope; touched only tracker | Usage and boundary searches | No frontend validation is required for S-00 through S-04 | None | Keep frontend work out of the cleanup iteration. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 11:13 EDT | Codex / tracker creation | Deployment and OTel contracts mapped; no generated edit authorized | Core 04, adopted OTel/Extensions/Network Flow NLSpecs, config fixtures, `contracts/otel`, verification registry; touched only tracker | Exact reads; `rg` for schema/environment references; `make explain-target TARGET=generate-drift DETAIL=summary` | No generated production dependency found; OTel matrices and harness manifests reference target tests as evidence | RB-002, RB-003 | Preserve wire shapes; use authored inputs and generators only if later test ownership moves. |
| 2026-08-07 12:09 EDT | Codex / legacy-removal planning | Telemetry ownership resolved but deferred; claim cutover blocked by adopted owner self-contradiction | Exact Extensions NLSpec v1/v3 clauses, current v3 projections, telemetry model/adapter, analysis notes; touched only tracker | Targeted schema/version and generated-projection searches | Owner document simultaneously requires v3 and v1; projections cannot choose the normative side | `BLOCKED: owner contradiction` for RB-003 | Extensions owner repairs and adopts one coherent contract before generated claim work. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 11:13 EDT | Codex / tracker creation | Existing target evidence cataloged; no tests executed | Seven target test files, `tools/test_families/platform.config.json`, verification owner/registry; touched only tracker | `make task-guide ROLE=module-author OWNER=platform.config`; `make explain-test-owner OWNER=platform.config`; `make help-all`; relevant `make explain-target` calls | Discovery commands succeeded; owner has 19 Go rows, no service-backed rows; this is not validation success | Affected caller-owner integration rows remain slice-dependent | Run `make lint-markdown` now; later start with `make test-slice OWNER=platform.config`. |
| 2026-08-07 12:09 EDT | Codex / legacy-removal planning | Pre-change baseline established and affected owner routes discovered | Target/configassembly tests and owner manifests for platform config, telemetry, Extensions, server, operator, and recovery; touched only tracker | `make test-slice OWNER=platform.config`; task guides and owner explanations | Platform-config baseline passed 19/19; app/server/operator/recovery service-backed commands are recorded | No harness change planned; implementation not authorized | Preserve test names/postconditions and run the Section 8 matrix per slice. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 11:13 EDT | Codex / tracker creation | Config controls startup inputs and secret-bearing references but performs no domain authorization decision | Core 04, OTel NLSpec, enterprise-auth and Network Flow config projections; touched only tracker | Exact source/owner reads and secret-reference searches | Fail-closed startup, origin, manifest, diagnostic, and redaction behavior added to freeze map | Later ownership moves must retain these outcomes | Add security-specific validation only for an authorized security-sensitive slice. |
| 2026-08-07 12:09 EDT | Codex / legacy-removal planning | Cleanup removes no security policy or diagnostic envelope; later telemetry movement retains explicit redaction prerequisites | Diagnostic callers, telemetry secret configuration, inactive-policy boundary; touched only tracker | Usage searches and owner reads | Removed candidates are metadata/access surfaces, not authorization or secret behavior | RB-002 prerequisites for later telemetry work | Keep security/OTel work outside S-00 through S-04. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 11:13 EDT | Codex / tracker creation | Discovery and planning complete; production work intentionally stopped | This tracker; no other file touched | Worktree/status inspection and command discovery | No owner contradiction found; three architecture/test-ownership questions remain | RB-001, RB-002, RB-003; later implementation authorization absent | Resolve blockers during WF-03/WF-05, then authorize and execute S-00 first. |
| 2026-08-07 12:09 EDT | Codex / legacy-removal planning | RB-001/RB-002 decisions recorded; next cleanup iteration is bounded; telemetry and claims are excluded | Tracker, analysis notes, owner docs, live usage and boundary evidence | Baseline tests, boundary check, task guides, exact source searches, tracker lint | S-00 through S-04 are decision complete; staged tracker history preserved; Markdown lint passed | RB-003; later implementation authorization | Begin a separately authorized S-00 task. |

## 11. Resolved Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | What is the permanent division between the generic `platform.config` engine and `internal/app/configassembly`? | Moving construction or translation in the wrong direction could make domain/application logic depend on deployment mechanics or make config own application composition. | Three-layer responsibility allocation in Section 3, Core 04, and live boundary policy | RESOLVED — kernel owns generic/Core 04 behavior; owner semantics remain owner-owned; application composition statically adapts and translates. |
| RB-002 | Should telemetry wire DTOs and configuration tests move to `platform.telemetry`, remain platform integration evidence, or split between both? | The OTel owner must own semantics, but moving all tests could lose strict decoder, environment containment, and secret-redaction integration evidence. | Adopted OTel owner, generic owner-namespace decoder, and immutable test-row crosswalk | RESOLVED AND IMPLEMENTED — WS-07 supplies the generic seam; WS-08 moves semantics and owner evidence while preserving parser collaboration. |
| RB-003 | Which descriptor and configuration-contract versions govern descriptor-derived Boolean claim projection? | The adopted Extensions NLSpec formerly named v3 as current while later clauses required v1. | Adopted owner repair, regenerated projections/integrity evidence, and complete claim parity matrix | RESOLVED AND IMPLEMENTED — WS-01 makes v3/v2 canonical without a v1 reader; WS-02 and WS-09 validate and consume that projection. |
| RB-004 | Is production/test/boundary implementation authorized? | The original tracker task was planning-only. | Explicit user authorization for WS-00 through WS-10 | RESOLVED — authorization was recorded before WS-00 and every gated workstream is complete or at the final Markdown gate. |
| RB-005 | Is the production-readiness cleanup implementation authorized? | The current request is explicitly a document-update step; treating it as production authorization would bypass the new per-workstream gates. | A later explicit implementation request covering WS-12 through WS-17 | OPEN AUTHORIZATION BOUNDARY — WS-11 may update this tracker only. This is not a product or specification blocker. |

There is no active product, specification, or release blocker and no approval exception
against the completed prior iteration. RB-005 limits this request to WS-11 documentation;
the current cleanup gaps are planned rather than silently deferred. Historical planning
labels in Sections 2 through 8 are retained only to explain the decisions that the
authorized WS-00 through WS-10 sequence superseded.

## 12. Prior Iteration Final Validation and Handoff

### Planning and implementation summary

The completed design has three durable boundaries:

1. `internal/platform/config` owns only generic Core 04 artifact discovery, strict TOML
   parsing, deterministic overlays, diagnostics, structural admission, immutable
   snapshots, and startup-root validation.
2. Source owners own namespace wire types, defaults, overlays, semantic validation, and
   cloning; `internal/app/configassembly` registers those contributions explicitly.
3. Extensions application assembly derives one immutable configuration policy from the
   admitted descriptor/configuration catalog, config collects only its registered
   Boolean paths, and a typed `RequestedClaims` value crosses into coordinator
   resolution. Resolved claims and published claim-set identity remain separate types.

No package-init registration, runtime scanning, reflection-based owner discovery,
dynamic plugin loading, compatibility reader, deprecated alias, or arbitrary Boolean
path lookup was introduced.

### Specifications, contracts, and implementation changed

- Normalized the Extensions NLSpec and Core 01/Core 04/OTel owner clauses to the current
  v3/v2 projection, generic owner-delegation seam, telemetry ownership, and
  request/resolution boundary.
- Hardened authored Extensions dependencies, claim/configuration facts, schema
  resolution, canonical ordering, row exactness, and digest validation; regenerated
  `internal/gen/contractextensions/artifacts_gen.go` through `make generate`.
- Contracted config/catalog/snapshot and configassembly APIs; privatized defaults;
  removed telemetry DTOs, owner-specific clone branches, raw Boolean maps, and unused
  compatibility surfaces.
- Added explicit namespace contributions for telemetry, enterprise authentication,
  Network Flow, and Revisions, plus a permanent backend boundary rule and negative
  fixtures.
- Reassigned telemetry verification ownership to `platform.telemetry`, regenerated the
  execution-topology render index, and recorded the retired-to-current ID crosswalk.
- Added claim registration, future-profile, malformed-catalog, inactive-policy,
  truth-table, immutability, dependency-failure, and no-partial-resolution coverage.

`docs/domain.md` is unchanged: configuration kernel, namespace decoder, and claim
registration are implementation/design vocabulary rather than new domain terms.

### Compatibility and migration result

- `cartulary.deployment_config.v2`, its six current claim paths, supported keys,
  default path, `CARTULARY_CONFIG_FILE`, strict parsing/overlays, effective defaults,
  diagnostics, and fail-closed startup behavior remain supported.
- There is no HTTP, database, browser schema, or deployment-config version migration.
- Obsolete Extensions v1 artifacts remain unsupported and have no compatibility reader.
- Removed internal Go APIs intentionally have no aliases or shims; all repository
  callers migrated atomically.
- Telemetry test IDs changed owner with an explicit crosswalk; retired IDs are not
  aliases.

### Final validation matrix

| Gate | Result and evidence |
| --- | --- |
| Finalization | 1/1, `.cartulary/test-results/20260807T175350Z-p4081227` |
| Harness contract | 2/2, `.cartulary/test-results/20260807T175406Z-p4083877` |
| Generation drift | 4/4, `.cartulary/test-results/20260807T175413Z-p4084256` |
| Generated-artifact policy | 3/3, `.cartulary/test-results/20260807T175426Z-p4086984` |
| JSON shape | 3/3, `.cartulary/test-results/20260807T175429Z-p4087431` |
| Backend module boundary | 3/3, `.cartulary/test-results/20260807T175439Z-p4087933` |
| Platform telemetry owner | 10/10, `.cartulary/test-results/20260807T175511Z-p4099748` |
| OTel conformance | 14/14, `.cartulary/test-results/20260807T175456Z-p4088274` |
| Platform config owner | 15/15, `.cartulary/test-results/20260807T175144Z-p4075770` |
| Extensions owner unit/service-backed | 37/37 and 24/24, `.cartulary/test-results/20260807T174214Z-p3875258` and `.cartulary/test-results/20260807T174252Z-p3902644` |
| Network Flow owner | 71/71, `.cartulary/test-results/20260807T174322Z-p3925490` |
| App server unit/service-backed | 42/42 and 33/33, `.cartulary/test-results/20260807T174951Z-p4042969` and `.cartulary/test-results/20260807T174448Z-p3963442` |
| Stateful browser | 36/36, `.cartulary/test-results/20260807T174752Z-p4016912` |
| Fast broad suite | 349/349, `.cartulary/test-results/20260807T175533Z-p4100146` |
| Targeted Go security | 4/4, `.cartulary/test-results/20260807T175747Z-p4167348` |
| Full check | 722/722, `.cartulary/test-results/20260807T175802Z-p4193493` |
| Release check | 893/893, `.cartulary/test-results/20260807T180201Z-p146336` |

The pre-broad `make agent-finalize` run intentionally omitted `RESULTS_DIR` because no
successful full warm check run existed yet. Its retained-run maintenance steps record
`results-dir-not-provided`; no check was skipped after a reusable warm run had been
selected. No unrelated failure exception is required.

### Residual risks and completion criteria

- [x] Authoritative specifications contain one coherent current v3/v2 contract.
- [x] Platform config has no owner DTO, telemetry rule, or hard-coded profile registry.
- [x] A fixture owner namespace and future claimable profile require no kernel edit.
- [x] Malformed, stale, duplicate, misordered, unresolved-schema, path-mismatched, and
  digest-mismatched owner projections fail closed.
- [x] Requested claims, coordinator resolution, and resolved-claim-set identity are
  mechanically distinct and expose no partial result on failure.
- [x] Removed internal APIs have no caller, alias, or compatibility shim.
- [x] Verification ownership, generated artifacts, stateful coverage, security checks,
  full checks, and release evidence are current and green.
- [x] Final `make lint-markdown` passes after this handoff reconciliation.

No known residual product or migration risk remains. The only continuing maintenance
obligation is architectural: new namespaces and profiles must use the explicit owner
contribution and admitted policy seams rather than widening the kernel boundary.

## 13. Current Production-Readiness Cleanup Iteration

### 13.1 Objective, authority, and preservation boundary

This iteration removes usage-proven legacy and test-shaped internal surfaces left after
the successful WS-00 through WS-10 architecture cutover. It is an atomic boundary
cleanup across `internal/platform/config` and only the adjacent application/test
packages required to delete those surfaces cleanly. It does not introduce a feature,
new configuration namespace, new extension profile, or new runtime policy.

The following external behavior remains frozen:

- `cartulary.deployment_config.v2`, all currently supported keys, and the six current
  claim paths;
- default artifact selection, `CARTULARY_CONFIG_FILE`, and the `CARTULARY__` overlay
  grammar and precedence;
- strict TOML and overlay type handling, unknown-key rejection, effective defaults,
  deterministic diagnostics, and fail-closed startup ordering;
- HTTP, database, browser, route, storage, secret-reference, and readiness contracts.

The current Core 04, Core 01, Extensions, and OTel owner text already requires an
owner-neutral kernel, explicit static namespace assembly, immutable admitted settings,
and distinct requested/resolved claim boundaries. The cleanup below is an implementation
contraction inside those requirements. No specification, authored product contract,
generated production artifact, or `docs/domain.md` edit is planned. If implementation
finds a genuine normative contradiction, the active slice stops and records a blocker
instead of expanding this cleanup through an incidental specification change.

### 13.2 Current gap register

| ID | Current evidence | Remediation and affected areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if unresolved | Completion criteria |
| --- | --- | --- | --- | --- | --- | --- |
| PRG-001 | `configassembly.Deployment` still carries `Import`, `IncidentPortability`, `ReferencePack`, and `SnapshotReporting`; repository search finds no reader outside their construction. `extensionassembly.ClaimConfiguration` exists only for those fields. | **Implementation and tests:** remove the four fields, their construction map, and `ClaimConfiguration`. Keep `Loaded.RequestedClaims()` as the only request-state projection. | Removes a second representation of requested claims and preserves the security distinction between deployment settings, requested claims, resolved claims, and published identity. New claim-only profiles no longer require application projection edits. | Intentional internal Go break with atomic caller/test updates and no alias. Operator-facing claim keys remain accepted through the generated claim policy. | A caller can treat an unused Boolean mirror as authoritative, and each future profile can recreate a central static projection list. | Removed-symbol searches are empty; current claim truth-table and Extensions/configassembly tests pass; requested claims remain unchanged for omitted, false, and true values. |
| PRG-002 | `config.Source.Decode` has no production caller; `Snapshot.Decode` is used only by configassembly to copy eight Core sections. Their support retains the full document and reflection helpers `documentSource`, `copyConfigurationValue`, `fieldByTOMLName`, `configurationFieldName`, `cloneConfig`, and `cloneReflectValue`. | **Implementation and tests:** add `CoreConfiguration` and defensive `Snapshot.Core()`; replace `Source` with presence-only `NamespacePresence`; store only Core configuration, sorted requested-claim IDs, and owner-cloned values in snapshots; remove arbitrary document decode and reflection-based document/owner cloning. | Gives the kernel a typed, closed projection and prevents owners from inspecting or copying unrelated namespaces. Owner-provided clone behavior becomes the only authority for owner values, making future map/slice/private-field types safe. | Internal catalog and snapshot API break only. Configassembly and telemetry migrate atomically; TOML shape and diagnostics remain unchanged. | Full-document access becomes an accidental compatibility surface, namespace isolation is advisory, and reflection can silently mishandle future owner types. | A fixture owner can observe presence only inside its registered namespace; `Snapshot.Core()` and `Value` are defensive; no removed helper remains; platform-config, telemetry, and OTel checks pass. |
| PRG-003 | Package-private `validate`, `validateWithExtensionPolicy`, `validateForStartup`, and `validateForStartupWithExtensionPolicy` are used only by tests or by startup revalidation. Startup currently repeats structural validation on an already admitted snapshot without the original extension policy. | **Implementation and tests:** run structural validation exactly once during admission; make startup validation reject an unadmitted snapshot and perform only filesystem canonicalization, overlap, and writability readiness checks; migrate tests to the production loader or narrow `_test.go` helpers around the production structural validator. | Establishes one owner for each validation phase, removes phase drift, and avoids reapplying defaults or validating with an incomplete policy. | No accepted input or diagnostic change. Tests stop depending on production-only wrappers and private document lifecycle shortcuts. | Structural behavior can diverge by call path, and an invalid zero snapshot could be mistaken for admitted input after simplification. | Structurally invalid input never yields a snapshot; zero/unadmitted snapshots fail closed; admitted startup diagnostics and ordering match the baseline; wrappers have no references. |
| PRG-004 | `configassembly.Admit` is documented as composition-test support; `config.LoadSnapshotFromTOML` supports it and one unit test. `server.NewRuntime` accepts a raw `Deployment` only to re-admit it. Production `Deployment` TOML tags and a thin appsupport path loader exist for these tests. | **Implementation and test support:** remove both in-memory admission APIs; make `server.NewRuntime` accept `configassembly.Loaded`; consolidate explicit path and fixture loading in `internal/testutil/configtest`; construct test variants through valid TOML fixtures plus explicit overlays; remove `Deployment` TOML tags and the duplicate appsupport loader. | Makes production construction consume admitted state only and ensures integration tests exercise the same discovery, overlay, catalog, and policy path as binaries. The application projection ceases pretending to be a wire DTO. | Intentional internal Go/test-harness break. Server, migrate, operator, recovery, HTTP test harness, and runtime tests migrate in one slice. No serialization shim is retained. | Tests can normalize or admit shapes that production never accepts, and test convenience continues defining production APIs and projection metadata. | No production API accepts an unadmitted `Deployment`; no test serializes `Deployment`; removed symbols/tags have no callers; affected builds and owner/service-backed/stateful slices pass. |
| PRG-005 | `configassembly.Load` accepts `config.LoadOptions`, then silently overwrites its `ExtensionPolicy`; application callers can supply a value that is ignored. | **Implementation and tests:** add `configassembly.LoadOptions { Path string; Env map[string]string }` and construct kernel `config.LoadOptions` plus the generated policy inside application assembly. | Makes application assembly's accepted inputs truthful and prevents generic kernel policy details from leaking into composition roots or tests. | Internal call sites migrate atomically. Platform `config.LoadOptions` remains the kernel input and retains its policy field for the explicit assembly edge. | A caller can believe it selected admission policy while application assembly silently substitutes another, obscuring ownership and tests. | No caller passes platform options to configassembly; path/env behavior is unchanged; generated policy remains mandatory and fail-closed. |
| PRG-006 | The current boundary check prevents owner imports and hard-coded profile registries, but it does not reject reintroduction of arbitrary snapshot/document access or production test-only admission. `inactive_policy.go` now owns the complete extension policy rather than an inactive-only abstraction. | **Tooling, implementation hygiene, and tests:** extend the authored backend boundary with negative fixtures for the removed access/admission forms; rename stale policy filenames; prune obsolete helpers and imports. | Converts cleanup decisions into durable architecture rules and leaves names aligned with current responsibility, reducing the chance that later work revives the legacy seam. | Boundary policy becomes stricter for future internal changes. No runtime or product compatibility impact. | The repository can regress to the same bypasses while all behavioral tests remain green, and stale names mislead future maintainers. | Representative forbidden fixtures fail; the repository passes `make backend-module-boundary-check`; exact symbol searches and generated drift are clean. |

### 13.3 Controlling interface decisions

The cleanup uses these exact internal boundaries:

- `config.CoreConfiguration` is the typed Core 04 projection containing schema ID,
  deployment profile, application, roots, bootstrap, timeouts, intervals, and limits.
  `Snapshot.Core()` returns a defensive value and does not expose the decoded document.
- `config.NamespacePresence` exposes only `Defined(path ...string) bool`. Paths remain
  full deployment paths for consistency with contribution path declarations and
  diagnostics, but a contribution receives `false` for any path outside its registered
  namespace. It cannot decode or inspect another namespace.
- A snapshot contains an admitted marker, `CoreConfiguration`, a sorted defensive list
  of requested-claim registration IDs, and typed owner values protected by each
  contribution's `Clone`. The transient decoded document and namespace map end at
  materialization.
- `configassembly.LoadOptions` contains exactly `Path` and `Env`. `Load` derives the
  generated Extensions configuration policy and passes a kernel `config.LoadOptions`
  internally.
- `server.NewRuntime` accepts a `configassembly.Loaded`, not a mutable application
  projection. Tests obtain `Loaded` through `configtest` fixture/path helpers and express
  configuration variants through explicit overlays before admission.
- `Loaded.Deployment()` remains a precomputed, infallible, defensive application
  projection. The live catalog builder, `NamespaceDecoder`, owner `Project` and `Clone`
  callbacks, typed `Value`, diagnostics, kernel `LoadOptions`, and `RequestedClaims`
  remain because they carry continuing production value.
- Reflection remains permitted for generic TOML decoding, typed environment overlay,
  and registered inactive-value traversal where it is live. This iteration removes only
  reflection used to provide arbitrary full-document copies or generic owner cloning.

Internal removals receive no deprecated aliases, compatibility readers, replacement
builders, or dual paths. Each slice updates all repository callers atomically.

### 13.4 Ordered workstreams

#### WS-11 — Tracker rebaseline

- **Areas:** documentation only.
- **Work:** record this gap register, exact interface decisions, workstream dependency
  chain, evidence template, compatibility boundary, and final validation plan. Preserve
  WS-00 through WS-10 and Sections 2 through 12 as dated historical evidence.
- **Dependency:** WS-10 completion.
- **Risk:** rewriting history or implying production authorization. Mitigate by adding
  this section, retaining prior evidence, and recording RB-005.
- **Exit:** this tracker is the only tracked file changed; `make lint-markdown` passes;
  WS-11 is `DONE`; WS-12 is the sole `READY` slice and remains authorization-gated.

#### WS-12 — Dead request-state projection removal

- **Areas:** implementation and tests in extensionassembly/configassembly.
- **Work:** remove the four claim-only `Deployment` fields,
  `extensionassembly.ClaimConfiguration`, the transient requested-set map used to fill
  them, and the now-unnecessary requested argument to deployment projection.
- **Dependency:** WS-11 and explicit implementation authorization.
- **Risk:** unintentionally removing live owner claim state. Enterprise Authentication
  and Network Flow owner configuration, generated claim registrations, and
  `Loaded.RequestedClaims()` are retained and tested separately.
- **Verification:** `make format`; exact removed-symbol searches;
  `make test-slice OWNER=platform.config`; `make test-slice OWNER=module.extensions`.
- **Exit:** PRG-001 completion criteria pass and the tracker/Markdown gate records the
  evidence before WS-13 begins.

#### WS-13 — Typed snapshot and namespace-presence contraction

- **Areas:** platform config, configassembly, telemetry configuration, and tests.
- **Work:** introduce `CoreConfiguration`, `Snapshot.Core()`, admitted/sorted snapshot
  state, and namespace-scoped `NamespacePresence`; migrate configassembly and telemetry;
  delete arbitrary source/snapshot decode and reflection-copy helpers.
- **Dependency:** WS-12.
- **Risk:** changing owner default presence, clone isolation, projection errors, or claim
  ordering. Add focused parity and hostile fixture-owner cases before deleting helpers.
- **Verification:** `make format`; `make test-slice OWNER=platform.config`;
  `make test-slice OWNER=platform.telemetry`; `make otel-conformance`; exact source and
  removed-symbol searches.
- **Exit:** PRG-002 criteria pass, no owner can observe another namespace, and the
  tracker/Markdown gate records the evidence before WS-14 begins.

#### WS-14 — Admission and startup-phase simplification

- **Areas:** platform-config validation, config/configassembly tests, and server startup
  characterization.
- **Work:** remove test-only structural wrappers and repeated structural validation from
  startup; make an unadmitted snapshot fail closed; retain only runtime filesystem
  canonicalization, overlap, and writability work in the startup phase.
- **Dependency:** WS-13.
- **Risk:** accepting a zero snapshot or changing diagnostic precedence. Preserve the
  existing contract tests and add explicit admitted/unadmitted phase tests.
- **Verification:** `make format`; `make test-slice OWNER=platform.config`;
  `make test-slice OWNER=app.server`; focused diagnostic-golden comparison.
- **Exit:** PRG-003 criteria pass and the tracker/Markdown gate records the evidence
  before WS-15 begins.

#### WS-15 — Application and test-loading boundary

- **Areas:** configassembly, server/migrate/operator composition, configtest/appsupport,
  HTTP/runtime harnesses, recovery test support, and affected tests.
- **Work:** introduce application `LoadOptions`; remove `Admit` and in-memory snapshot
  loading; make runtime construction accept `Loaded`; replace mutable `Deployment`
  fixtures with admitted TOML fixture plus overlay helpers; remove projection TOML tags
  and the duplicate appsupport loader.
- **Dependency:** WS-14.
- **Risk:** startup ordering, borrowed-resource ownership, dynamic test roots, or service
  binding setup can change during widespread caller migration. Keep runtime assembly
  order unchanged and express every prior test mutation as a pre-admission overlay.
- **Verification:** `make format`; `make build-server`; `make build-migrate`;
  `make build-operator`; unit slices for `platform.config`, `app.server`, `app.operator`,
  and `platform.postgres`; service-backed slices for `app.server`, `app.operator`, and
  `module.recovery`; `make browser-e2e-stateful`; exact removed-symbol and TOML-tag
  searches.
- **Exit:** PRG-004 and PRG-005 criteria pass and the tracker/Markdown gate records all
  run roots before WS-16 begins.

#### WS-16 — Permanent enforcement and residual pruning

- **Areas:** authored backend boundary policy/checker fixtures, filenames, tests, and
  verification metadata only if an existing row must move.
- **Work:** add fail-closed fixtures for arbitrary document/snapshot access and
  production test-only admission; rename inactive-only policy files to their current
  extension-policy responsibility; remove obsolete helpers and imports. Preserve test
  function names and ownership where possible to avoid meaningless topology churn.
- **Dependency:** WS-15.
- **Risk:** a token rule can reject legitimate decoder or overlay reflection. Match only
  the removed APIs/patterns and retain the current TOML/overlay mechanisms explicitly.
- **Verification:** `make format`; `make backend-module-boundary-check`;
  `make harness-contract`; `make generate-drift`;
  `make generated-artifact-policy-check`; exact forbidden-symbol searches.
- **Exit:** PRG-006 criteria pass, generated outputs are unchanged or regenerated only
  from authored inputs, and the tracker/Markdown gate records the evidence before
  WS-17 begins.

#### WS-17 — Validation and handoff completion

- **Areas:** verification, tracker evidence, and generated topology only if required by
  an authored verification change.
- **Work:** reconcile every gap and workstream; run finalization, focused owner and
  service-backed checks, stateful coverage, security, broad, and release validation;
  document failures, run roots, compatibility, generated artifacts, residual risks, and
  skipped checks.
- **Dependency:** WS-12 through WS-16.
- **Risk:** broad failures obscure ownership. Use `make explain-test-owner`,
  `make explain-target`, and `make explain-run`, then rerun the narrow owning slice
  before repeating broad verification.
- **Verification:** run the complete matrix in Section 13.6.
- **Exit:** every required command passes or an explicitly approved unrelated exception
  is documented; every PRG row is closed; no cleanup row is deferred; final tracker lint
  passes; WS-17 is `DONE`.

The dependency chain is strictly:

`WS-11 -> WS-12 -> WS-13 -> WS-14 -> WS-15 -> WS-16 -> WS-17`

No slice advances around a failed or blocked predecessor.

### 13.5 Mandatory tracker and evidence gate

At the end of every workstream, before the next workstream begins:

1. Record timestamp, agent/session, affected owners, files changed, and the substantive
   ownership or compatibility decision.
2. Record every verification command, result, run root or summary artifact, and whether
   each failure was related. A failed required command is never omitted from the log.
3. Record removed/changed internal interfaces, external compatibility effects, newly
   discovered risks, generated artifacts, and follow-up work.
4. Mark the current workstream `DONE` only after every exit criterion passes. Mark
   exactly one successor `READY`; all later rows remain `PLANNED`.
5. Run `make lint-markdown` after the tracker update and record its run root. Do not
   begin the successor until this gate passes.

Use this evidence shape for each new Section 10 row:

| Field | Required content |
| --- | --- |
| Time and agent | Local timestamp, agent, and workstream ID |
| Current state | Resulting architecture or behavior, not an activity summary |
| Owners and files | Every affected semantic owner and authored/generated file family |
| Commands | Exact Make targets and any read-only source audits |
| Results | Counts, run roots, summary artifacts, and related failure resolution |
| Compatibility and risks | External stability, intentional internal breaks, and new risks |
| Next action | Exactly one successor or final completion |

### 13.6 Final validation matrix

WS-17 runs at least the following after `make agent-finalize`. Supply
`RESULTS_DIR=<successful-full-warm-run-root>` only when such a retained run exists;
otherwise record `results-dir-not-provided` as the reason retained-run maintenance was
skipped.

| Validation family | Required command or evidence |
| --- | --- |
| Tracker | `make lint-markdown` |
| Harness contract | `make harness-contract` |
| Generation | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` |
| Architecture | `make backend-module-boundary-check`; exact removed-symbol and forbidden-pattern searches |
| Platform config | `make test-slice OWNER=platform.config` |
| Telemetry | `make test-slice OWNER=platform.telemetry`; `make otel-conformance` |
| Extensions | `make test-slice OWNER=module.extensions`; service-backed slice when routed by the owner guide |
| Server | `make test-slice OWNER=app.server`; `make service-backed-test-slice OWNER=app.server` |
| Operator and migration | Unit/service-backed `app.operator`; `platform.postgres` owner slice; server/migrate/operator builds |
| Recovery | `make service-backed-test-slice OWNER=module.recovery` |
| Stateful runtime | `make browser-e2e-stateful` |
| Fast broad suite | `make test-fast` |
| Targeted security | `make go-gosec-targeted` |
| Full repository | `make check` |
| Release | `make release-check` |

Use `make task-guide ROLE=module-author OWNER=<owner-id>` to confirm the current routed
commands before each focused slice. If test functions or packages move, update authored
owner rows, assign new immutable IDs only when semantic ownership changes, record a
crosswalk, regenerate through Make, and include harness/topology evidence.

### 13.7 Completion criteria

The cleanup iteration is complete only when:

- request state exists only as typed `RequestedClaims`; no claim-only Boolean mirror or
  static application projection list remains;
- the kernel exposes no arbitrary full-document or snapshot-path decode, retains no
  owner namespace document after materialization, and performs no reflective owner
  cloning;
- owner presence is mechanically namespace-scoped and owner values remain defensively
  cloned;
- structural admission runs once, startup filesystem validation runs only on an
  admitted snapshot, and zero/unadmitted state fails closed;
- production composition accepts admitted configuration, while test configuration uses
  the same file discovery, overlays, catalog, and generated policy as production;
- configassembly accepts only truthful path/environment options and never silently
  replaces a caller-supplied policy;
- boundary enforcement rejects reintroduction of the removed seams;
- all internal breaks have no caller, shim, deprecated alias, or compatibility reader;
- the external deployment v2 contract and startup behavior remain unchanged;
- final focused, service-backed, stateful, security, full, release, generated-artifact,
  and tracker evidence is current and green.
