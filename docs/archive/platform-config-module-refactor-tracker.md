# platform-config Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current value |
| --- | --- |
| Target path | `internal/platform/config` |
| Target label | `platform-config` |
| Output path | `docs/handoffs/platform-config-module-refactor-tracker.md` |
| Repository snapshot | `main` at `781e4f015d81edc9e5d2c51f05724f828d622615`; this tracker was already staged as a new file and no other path was changed when execution began |
| Migration head at activation | `00036_reporting_composition_preview_outputs.sql` |
| Status | Complete; `PC-SL-00` through `PC-SL-15` are `DONE` and RB-001 through RB-003 are `CLOSED` |
| Allowed changes | Owner specifications and authored contracts, generated outputs through Make, backend implementation/tests, append-only migrations, harness inputs, guides, and this tracker |
| Non-goals | Frontend/browser behavior, view schemas, revisions, projections, and authorization logic unless live-call inventory proves a direct dependency |
| Implementation authorization | User authorized the complete remediation plan on 2026-07-24 |
| Compatibility decisions | Network Flow document `2.0.1` with contract major 2 and no diagnostic alias; pre-release database reset with no path backfill or dual-read compatibility |

Source hierarchy used for this tracker:

1. Adopted subsystem NLSpecs for their named subsystem.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides for terminology, package boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans, handoffs, and the modular-refactor framework as evidence and planning doctrine only.

Owner and support documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` was read first and used as planning doctrine, not as proof of repository state.
- `docs/spec/00_document_set_status_and_precedence.md`.
- `docs/spec/04_security_deployment_and_conformance.md`, especially the deployment-configuration, root-binding, validation, and startup requirements.
- `docs/extension-subsystem-nlspec.md`.
- `docs/opentelemetry-instrumentation-nlspec.md`.
- `docs/network-flow-activity-nlspec.md`.
- `docs/testing-harness-nlspec.md`.
- `docs/domain.md`.

Repository evidence inspected:

- Every file under `internal/platform/config`, inventoried in Section 2.
- Application callers in `internal/app/extensionassembly/configuration.go`, `internal/app/server/runtime.go`, `internal/app/server/server.go`, `internal/app/migrate/migrate.go`, and `internal/app/operator/operator.go`.
- Platform consumers in `internal/platform/httpapi/httpapi.go`, `internal/platform/postgres/postgres.go`, `internal/platform/objectstore/objectstore.go`, `internal/platform/telemetry/bootstrap.go`, and `internal/platform/viewquery/query.go`.
- Module consumers in `internal/modules/imports/routes.go`, `internal/modules/imports/xlsx.go`, `internal/modules/incidentbundles/bundle.go`, `internal/modules/reference_data/verifier.go`, and `internal/modules/networkflow/keyring.go`.
- Reusable test configuration in `internal/testutil/configtest/configtest.go`.
- Harness ownership in `tools/test_families/platform.config.json`, `contracts/verification/owners/platform.config.json`, and `contracts/verification/owners/platform.telemetry.json`.
- Extension configuration inputs in `contracts/extensions/fragments/core04.claim-configuration.json`, `contracts/extensions/profiles/network_flow_activity/configuration.json`, `contracts/extensions/profiles/enterprise_authentication/configuration.json`, and `contracts/extensions/validation/surfaces.json`.
- Telemetry inputs and evidence in `contracts/otel/telemetry_config_schema.v1.json`, `contracts/otel/config_hazard_fixture_matrix.v1.json`, `contracts/otel/conformance_status.json`, and `tools/otel/check-otel-conformance.mjs`.
- The development configuration input `configs/dev/config.toml` and generated consumers `internal/gen/contracts/contracts_gen.go` and `packages/protocol-ts/src/generated/contracts.ts`; generated files were read only.

The framework describes a target-specific module boundary, but the live target is not a single permanent domain module. It is a legitimate platform configuration facade combined with subsystem-specific validation, secret handling, extension policy, and DTOs consumed across module boundaries. This mismatch is a planning finding, not an implementation decision. The package also has no direct HTTP route handlers, WebSocket publishers, SQL, entity row mutations, saved-view persistence, projection refresh, revision coordination, authorization decisions, frontend controller state, or grid-vendor integration.

No prior tracker existed at this path. The planning history remains in the appendices, and this tracker now controls the authorized implementation. No workstream may begin until the previous workstream is marked complete here and the next is marked `IN_PROGRESS`.

## 2. Final-State Repository Inventory

The final target contains one strict generic parser/container, one immutable
catalog/snapshot boundary, private wire representations for contributed
namespaces, neutral inactive-policy application, and focused tests. The
directory has no placeholder or owner-specific policy package.

| Path | Final responsibility | Production boundary |
| --- | --- | --- |
| `config.go` | Artifact selection, strict TOML/environment parsing, closed-key admission, core wire DTOs, diagnostic envelope, and exact claim projection | No exported monolithic configuration value or legacy loader/validator |
| `catalog.go` | Immutable typed keys, contribution ownership checks, owner overlay dispatch, presence-aware source, and deterministic contribution order | Imports no concrete owner |
| `snapshot.go` | Immutable materialization, typed owner values, narrow decode for application composition, and startup-root validation entry point | Snapshot is consumed only by `internal/app/configassembly` |
| `telemetry_document.go` | Private wire-shaped telemetry decode target | Contains no telemetry policy, defaults, validation, secret handling, or owner import |
| `inactive_policy.go` | Neutral interface applied by the parser | Extensions owns catalog construction and application assembly adapts it |
| `validation.go` | Core structural validation, root admission/canonicalization, overlap checks, and explicit claim lookup | Contains no telemetry, Network Flow, enterprise-auth, secret, or manifest policy |
| `backup_config_test.go`, `claim_projection_test.go`, `core_config_test.go`, `inactive_extension_test.go` | Core and integration characterization | Canonical `platform.config` owner evidence |
| `catalog_test.go`, `config_otel_test.go`, `inactive_policy_test_support_test.go` | Catalog lifecycle, telemetry-owner integration, and neutral policy test support | No production compatibility aliases |

The directory was reviewed against `docs/domain.md`; no domain term or concept
boundary changed, so the vocabulary document requires no edit.

### Activation inventory (historical)

The table below is the pre-execution inventory retained for audit history. Its
symbols, callers, package paths, and risk notes describe the starting state,
not the final repository.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/platform/config/.gitkeep` | Historical placeholder in a now-populated directory | None | None | None | None | None | No semantic owner | low | Obsolete but harmless; remove only in a later structural cleanup slice. |
| `internal/platform/config/backup_config_test.go` | Characterizes backup-storage root binding and defaults | `TestBackupStorageRootBindingConfig` test selector | `tools/test_families/platform.config.json` through the package test target | Target configuration API and test helpers | This file is test evidence | Core 04 root-binding behavior and `platform.config` harness row | Platform configuration tests | low | Preserve selector identity if tests move because the harness names it. |
| `internal/platform/config/claim_projection_test.go` | Characterizes Boolean claim projection and rejection of unknown or duplicate paths | Two `TestBooleanValuesAtPaths*` selectors | `platform.config` test-family manifest | `BooleanValuesAtPaths`, base config fixtures | This file is test evidence | Generated extension descriptors supply claim paths consumed by application assembly | Extensions configuration boundary tests | medium | Protects application composition from concrete profile-field knowledge. |
| `internal/platform/config/config.go` | Defines the deployment configuration model, defaults, path selection, TOML decoding, environment overlays, closed-key admission, claim projection, and diagnostic envelope | `Config`, nested configuration and limit types, `LoadOptions`, `Diagnostic`, `DiagnosticsError`, constants, `BooleanValuesAtPaths`, `ResolvePath*`, `Load*`, `DiagnosticsFromError`, `NewDiagnosticsError` | Server, migrate, operator, extension assembly, platform adapters, domain modules, test utilities, process tests, and tools | Standard library, `github.com/BurntSushi/toml`, `extensioninactive` | All target tests plus caller tests | Core 04 deployment configuration; extension claim descriptors; telemetry schema; generated Go/TypeScript descriptors indirectly | Keep a narrower generic platform configuration facade; split owner-specific DTOs and policy | critical | Central public package surface. Changes can alter startup, diagnostics, defaults, config wire keys, and many constructors. |
| `internal/platform/config/config_otel_test.go` | Characterizes the closed `telemetry.*` namespace, defaults, hostile OTel environment isolation, cross-key validation, secret references, redaction, and header bounds | Four `TestOpenTelemetry*` selectors | Static OTel conformance evidence; not currently selected by the `platform.config` owner manifest | Target configuration and telemetry secret helpers | This file is test evidence | Telemetry schema, hazard matrix, conformance status, and static OTel checker | OpenTelemetry configuration owner tests, executed through platform configuration | high | All four selectors are missing from canonical `platform.config` owner-slice accounting; see RB-002. |
| `internal/platform/config/core_config_test.go` | Broad characterization for config discovery, enterprise authentication, Network Flow claimability, extension runtime values, roots, filesystem paths, disconnected defaults, and resource limits; also exports package test helpers | Eight `Test*` selectors plus `BaseConfig`, root-name, managed-service, and environment helpers | `platform.config` test-family manifest selects seven of the eight tests; other package tests reuse helpers | Target configuration API, filesystem, environment, and test-only root helpers | This file is test evidence | Core 04, Extensions, enterprise-authentication, Network Flow, and resource-limit configuration | Platform configuration tests plus named subsystem-owner tests | high | `TestNetworkFlowActivityConfigDefaultsAndClaimability` is not in canonical owner-slice accounting and exposes RB-001. |
| `internal/platform/config/extensioninactive/catalog.go` | Builds a key-policy catalog and validates/discards inactive extension configuration using forbidden or syntax-only policy | `PolicyKind`, `PolicyForbidden`, `PolicySyntaxOnly`, `Policy`, `Finding`, `Catalog`, `NewCatalog`, `Catalog.Policy`, `Catalog.Keys`, `Catalog.ValidateAndDiscard` | `config.go`, `validation.go`, extension application assembly, package tests | Standard library only | `inactive_extension_test.go` and configuration-load tests | Extension validation surfaces and generated descriptor metadata indirectly | Extensions owner behind a neutral configuration contribution | high | Its current platform location is suspect; exact permanent package/interface is deferred until the provider boundary is characterized. |
| `internal/platform/config/inactive_extension_test.go` | Characterizes forbidden and syntax-only inactive extension configuration, discard behavior, and diagnostics | `TestInactiveExtensionConfiguration_Unit` | `platform.config` test-family manifest | Target load/validation and `extensioninactive` catalog | This file is test evidence | Extensions inactive-policy contracts | Extensions/configuration integration tests | medium | Preserve syntax-only inert behavior and exact diagnostic output. |
| `internal/platform/config/validation.go` | Normalizes and validates schema/profile/origin, roots, bootstrap manifests, subsystem claims, runtime values, resource limits, telemetry, filesystem readiness, and telemetry secrets | `Validate*`, `ValidateForStartup*`, `ResolveTelemetrySecretReferences`, `RegisterTelemetrySecretPurposes`, `ResolveTelemetryExporterHeaders` | Server startup and config loading; telemetry bootstrap; target and caller tests | Filesystem and network APIs, UUID, `x/sys/unix`, `extensioninactive`, `secretpurpose` | Core, inactive-extension, telemetry, server-runtime, and platform-adapter tests | Core 04 deployment validation, subsystem configuration owners, telemetry contracts | Split generic validation coordination from owner-specific validators and secret handling | critical | Contains high-impact startup behavior and test-referenced filesystem helpers that are not direct production call paths; see RB-003. |

Package-surface freeze for later work:

- Keep current config keys, TOML shapes, default values, environment overlay names, error envelope, diagnostic paths/reason codes/messages/order, and startup timing unchanged unless an owner-authorized behavior task says otherwise.
- Retain compatibility for `Config`, `Load*`, `ResolvePath*`, `Validate*`, `ValidateForStartup*`, `DiagnosticsError`, claim projection, root-binding DTOs, limit DTOs, telemetry DTOs, and secret helpers while callers migrate through isolated slices.
- Treat the exported test helpers in `core_config_test.go` as test surface rather than production API, while preserving named harness selectors.

## 3. Final Module Boundary Diagnosis

- `internal/platform/config` is a cohesive Core 04 platform parser and immutable
  configuration-container owner.
- `internal/app/configassembly` is the only full deployment projection and
  owner-value materialization boundary.
- Telemetry, Extensions inactive policy, Network Flow, and Enterprise
  Authentication own their pure configuration semantics.
- Modules and HTTP transport import none of `config`, `rootedfs`, or
  `securefile`; adapters accept narrow settings and ports.
- Rooted filesystem and secure-file primitives remain neutral platform
  mechanisms, wrapped by application-owned adapters.
- Database and job records store strict relative logical references, with no
  compatibility column or dual-read path.

### Activation diagnosis (historical)

Current classification:

- legitimate platform configuration facade;
- mixed-responsibility package;
- transport-adjacent configuration source through `httpapi.DependencySet`;
- persistence-adjacent adapter input through Postgres and object-store constructors;
- misplaced home for some logic owned by Extensions, OpenTelemetry, Network Flow, enterprise authentication, imports, evidence/portability, and view-query contracts.

It is not currently a view/projection orchestration layer, mutation coordinator, frontend shell/controller surface, or grid-vendor integration layer.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Config artifact discovery, default path, absolute selector override, TOML decoding, `CARTULARY__` overlays, and unknown-key rejection | `config.go` | Generic platform configuration facade under Core 04 | keep | Core 04 and live `Load*` implementation/tests | This is legitimate platform orchestration. |
| Diagnostic aggregation and stable deployment-config error envelope | `config.go`, `validation.go` | Generic platform configuration facade | keep | `DiagnosticsError`, load/startup callers, Core 04 | Owner validators may contribute findings, but the facade should coordinate the envelope and ordering. |
| Root-binding DTOs, profile compatibility, overlap, canonicalization, and startup readiness | `config.go`, `validation.go` | Core 04 platform configuration plus a possible filesystem-boundary adapter | split | Core 04, server startup, Postgres/object-store consumers, root tests | Keep config admission and startup coordination; defer the final home of canonical filesystem operations pending RB-003. |
| Deployment-profile and application public-origin validation | `validation.go` | Generic platform configuration facade under Core 04 | keep | Core 04 and server runtime | Public origin indirectly affects HTTP/WebSocket admission but is not route ownership. |
| Generated extension claim projection | `BooleanValuesAtPaths`, extension assembly | Extensions owner with platform config projection support | split | Extension descriptors and `claim_projection_test.go` | Keep neutral Boolean path projection; Extensions owns which paths are claims. |
| Inactive extension forbidden/syntax-only policy | `extensioninactive/catalog.go`, config load/validation | Extensions subsystem | move | Extensions NLSpec, generated validation surfaces, extension assembly | Final package and interface are deferred until a neutral contribution contract is characterized. |
| Enterprise-authentication claimed/provider-manifest validation | `Config` and `validation.go` | Core 04 configuration with `internal/platform/enterpriseauth` contribution | split | Core 04 and enterprise-auth manifest consumer | Preserve keys and diagnostic behavior; avoid making generic config know provider semantics beyond the adopted owner boundary. |
| Network Flow claimed/key-ring-manifest validation | `Config` and `validation.go` | Network Flow owner contributing config validation | move | Network Flow NLSpec, extension contract, `networkflow/keyring.go` | Blocked for the disputed unclaimed-key diagnostic by RB-001. |
| Telemetry DTOs, defaults, hostile environment isolation, cross-key validation, secret resolution, and purpose registration | `config.go`, `validation.go` | OpenTelemetry subsystem with platform config parsing coordination | move | OpenTelemetry NLSpec, telemetry contracts/tests, telemetry bootstrap | Preserve the `telemetry.*` wire namespace and fail-before-readiness semantics. |
| Resource-limit registry and owner-specific limit DTOs | `config.go`, `validation.go` | Generic numeric admission plus owner-local options in imports, object storage, evidence/portability, reference data, and Extensions | split | Direct module imports and runtime consumers | Application composition should map config values to owner-local option types. |
| Public sort/filter/change/collection-action ceilings | `config.go` | View-query and public protocol/mutation contract owners | move | `internal/platform/viewquery/query.go`; no non-test use found for change/action constants | Values remain frozen; exact change/action owner needs caller discovery before movement. |
| Full `Config` propagation to HTTP/module dependencies | `httpapi.DependencySet` and module consumers | Application composition mapping to narrow capabilities/options | move | Live importer and symbol-use scan | Configuration must not remain a service locator for domain modules. |
| Test-only filesystem path/write helper behavior in production source | Private helpers in `validation.go`, exercised from `core_config_test.go` | Actual filesystem-writing owners or test support after characterization | defer | No direct production call found for the tested helpers | Do not delete or move until real import/extract/write consumers are characterized under RB-003. |

This diagnosis is an architectural finding only. It does not authorize moving any symbol or changing a permanent module boundary.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `/etc/cartulary/config.toml`, `CARTULARY_CONFIG_FILE`, explicit path precedence, and absolute-path rejection | Core 04 / platform configuration | Core 04; `config.go` | `TestConfigDiscovery_Unit` | Preserve missing/default/explicit/env combinations and exact diagnostics | high | File selection occurs before application assembly. |
| `cartulary.deployment_config.v1`, deployment profile, TOML shape, `CARTULARY__` overlays, closed keyspace, and unknown-key failure | Core 04 plus named subsystem owners for their namespaces | Core 04, owner NLSpecs, `config.go` | Core and OTel configuration tests | Add an owner-composition case proving contributed namespaces retain strict unknown-key behavior | critical | Owner extraction must not make configuration open-ended. |
| `invalid_deployment_config` envelope and ordered diagnostic path/reason/message/JSON output | Core 04 / platform configuration | `Diagnostic`, `DiagnosticsError`, load/validation code | Core and telemetry tests; server/operator tests | Add golden multi-owner aggregation and stable-order cases before validator extraction | critical | CLI/startup consumers depend on the envelope. |
| Root binding kinds, service refs, profile compatibility, lexical/canonical/writable/overlap checks | Core 04 | Core 04; `validation.go`; Postgres/object-store consumers | Runtime-root and filesystem-path tests | Exercise actual import/extract/write consumers against traversal, symlink escape, and unwritable roots | critical | RB-003 blocks cleanup of tested helpers. |
| Invalid configuration prevents HTTP, WebSocket, and background jobs from reaching readiness | Core 04 / application server | Core 04; `server/runtime.go` | Server runtime and integration tests | Preserve failure-before-listener/job behavior with owner-contributed validation | critical | Validation phase ordering is observable. |
| `application.public_origin` and browser WebSocket Origin admission | Core 04 config plus HTTP/WebSocket platform | `validation.go`, server/http composition | Config and server tests | Preserve exact allowed/rejected origin outcomes after facade narrowing | high | Indirect contract; config owns the value, not route logic. |
| Extension claim Boolean projection and immutable claim request | Extensions subsystem | Generated extension descriptors, extension assembly, `BooleanValuesAtPaths` | Claim-projection and server assembly tests | Preserve missing/duplicate/unresolved path failure and descriptor-driven projection | high | Do not hardcode claim paths in application composition. |
| Inactive extension forbidden and syntax-only policy, including discard behavior | Extensions subsystem | Extensions NLSpec, validation surfaces, `extensioninactive` | `TestInactiveExtensionConfiguration_Unit` | Add provider-boundary cases with multiple inactive owners and stable diagnostic ordering | high | No generated descriptor may be hand-edited. |
| Extension timeouts, intervals, and limits | Core 04 and Extensions | Core 04, Extensions NLSpec, config DTOs | `TestExtensionRuntimeConfig_Unit`, resource-limit tests | Preserve defaults, bounds, and startup failure across a contribution seam | high | Values are consumed by runtime coordinators. |
| Enterprise-authentication claim and provider-manifest path | Core 04 and enterprise-authentication owner | Core 04; enterprise-auth contract/config; manifest loader | `TestEnterpriseAuthenticationConfig_Unit`, enterprise-auth tests | Preserve claimed/unclaimed/default/path and startup validation cases | high | Extraction must retain existing manifest-load order. |
| Network Flow claim and key-ring-manifest path | Network Flow and Extensions owners | Both adopted NLSpecs, generated validation surface, live owner contribution and key-ring code | `TestNetworkFlowActivityConfigDefaultsAndClaimability`, exact config matrix, Network Flow key-ring and startup tests | Claimed/unclaimed/path/file-state matrix is executable and passing | critical | RB-001 is closed: inactive configuration uses only `extension_config_without_claim`, with no alias and no file or secret work. |
| `telemetry.*` defaults, closed namespace, hostile OTel environment isolation, cross-key rules, secret refs, redaction, and purpose registry | OpenTelemetry subsystem | OTel NLSpec, telemetry schema/hazard matrix/conformance status | Four `TestOpenTelemetry*` tests and telemetry package tests | Put all four tests under canonical owner execution; preserve aggregation, redaction, header-size, and purpose-reuse cases | critical | Static `otel-conformance` evidence does not replace executing these tests. |
| Object, import, archive, incident-bundle, reference-pack, preview, extension, sort, and filter limits | Core 04 configuration plus consuming owners | Config DTOs/defaults; module and platform consumers | `TestResourceLimits_Unit` and consumer tests | Add mapping tests at application composition for each narrowed owner option | high | Storage and request outcomes must not change. |
| Generated extension/profile/protocol descriptors that carry claim keys and inactive policy | Extensions/contract owners; generated layer downstream | Extension fragments/profiles/surfaces and generated Go/TS files | Generation/drift and claim-projection tests | Ensure owner-input changes regenerate through Make and produce no unrelated drift | high | Target imports no generated package directly, but application assembly uses descriptor-derived paths. |
| `platform.config` harness rows, selectors, scheduling, and evidence ownership | Testing Harness NLSpec and owner manifests | `tools/test_families/platform.config.json`, verification owners | Eighteen active rows; all sixteen intended top-level symbols resolve exactly once | Accounting gap closed; retain exact selector and verification identity | high | Phase/test rows are evidence accounting, not runtime architecture. |
| HTTP routes, WebSocket event shapes, entity row/query/mutation envelopes, saved views/view schemas, projection refresh, revisions, authorization, UI/grid selectors | Their named owners; not this target | Target source and caller inspection found no direct ownership | Existing owner tests outside this package | None unless a later slice changes origin, limit, or visibility behavior | low | Explicitly out of direct scope; indirect configuration effects remain frozen above. |

## 5. Coupling and Boundary Findings

Every activation `must_fix` and `should_fix` item below is closed by
PC-SL-03 through PC-SL-14. The retained table records the starting diagnosis;
Sections 7, 9, 10, and 12 state the implemented boundary and evidence.

### Activation findings (historical)

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Domain modules import platform configuration DTOs directly | Imports, incident bundles, reference data, recovery, and Network Flow import/use scans | Domain logic remains coupled to a broad platform schema and unrelated changes fan out | `must_fix` | Each domain owner, mapped by application composition | Define owner-local option types and one-way mapping from validated config. |
| `httpapi.DependencySet` carries the full `config.Config` | `internal/platform/httpapi/httpapi.go` and module consumers | Transport becomes a configuration service locator and modules can read unrelated settings | `must_fix` | Application composition plus narrow transport capabilities | Inventory each field read, inject only owner-specific options/capabilities, then remove the broad field. |
| Generic config hardcodes telemetry schema, validation, and secret behavior | `config.go`, `validation.go`, OTel contracts and tests | OpenTelemetry owner changes require edits in a broad platform package | `must_fix` | OpenTelemetry subsystem | Design a strict namespace contribution that preserves parsing, diagnostics, and pre-readiness validation. |
| Generic config hardcodes extension inactive policy | `extensioninactive`, extension assembly, generated validation surfaces | Generated owner policy and runtime validation can drift | `must_fix` | Extensions subsystem | Move policy construction behind the Extensions owner; keep neutral application in the config facade. |
| Generic config hardcodes Network Flow key-ring admission | `NetworkFlowActivityConfig`, validation, key-ring consumer | Current owners contradict one another for an observable diagnostic | `must_fix` | Network Flow plus Extensions owner resolution | Block the affected move on RB-001; do not choose a code path. |
| Resource-limit DTOs cross into unrelated modules | Direct imports of `ImportLimits`, `ArchiveLimits`, `ReferencePackLimits`, and `LimitConfig` | Boundary leaks make platform configuration part of domain APIs | `must_fix` | Consuming modules | Introduce owner-local immutable options with exact value mapping and tests. |
| Public sort/filter and unused change/action constants live in config | View-query use scan and config test references | Protocol policy is mislabeled as deployment configuration | `must_fix` | View-query and public mutation contract owners | Preserve numeric values; discover exact owner for unused constants before movement. |
| Platform adapters accept the full `Config` when they use roots, telemetry, or a small subset | Postgres, object-store, telemetry, and recovery callers | Broad dependencies hide required inputs and complicate testing | `should_fix` | Relevant platform adapters | Narrow constructors after compatibility tests establish the used fields. |
| Load-time validation and server startup validation overlap | `LoadWithOptions`, `Validate*`, `server/runtime.go` | Refactor can change ordering, repeat side effects, or alter diagnostics | `should_fix` | Generic config facade and application server | Characterize the exact stages and make validation phases explicit without behavior change. |
| Filesystem containment/write helpers are tested from production source but lack direct production callers | Symbol-use scan and `core_config_test.go` | Tests can overstate real enforcement while dead helpers look authoritative | `must_fix` | Actual import/extract/write owners | Close RB-003 with consumer-level characterization before deleting or relocating helpers. |
| `.gitkeep` remains in a populated directory | Target inventory | Cosmetic drift only | `should_fix` | Repository housekeeping | Remove in the last cleanup slice, not in this planning task. |
| Config appropriately uses OS/environment/TOML/filesystem APIs | `config.go`, `validation.go`, Core 04 | Moving all platform work into domain owners would invert dependencies | `intentional/no_action` | Platform configuration facade | Retain platform mechanics and strict parsing in the facade. |
| Application facades load config and compose the extension catalog | Server, migrate, operator, and extension assembly | This is the intended composition-root direction | `intentional/no_action` | Exact `internal/app/*` facades | Preserve construction in application assembly while narrowing inputs. |
| No direct SQL, grid-vendor, frontend, route, WebSocket, saved-view, projection, revision, or authorization logic exists in the target | Target and caller inspection | Expanding the refactor into those systems would create unrelated behavior risk | `intentional/no_action` | Existing named owners | Keep them out of scope except for explicit indirect contract characterization. |
| Exact final package/interface for `extensioninactive` is not proven | Current package, extension assembly, generated metadata | Premature placement can create a reverse dependency | `defer` | Extensions plus generic config facade | Decide after the contribution interface and package-direction tests exist. |
| Exact home of canonical root-path operations is not proven | Root validation plus RB-003 | Moving validation away from actual filesystem effects can weaken enforcement | `defer` | Core 04 config or a narrow filesystem-boundary adapter | Decide only after real consumers and failure timing are characterized. |
| Generated files are downstream evidence, not edit sources | Generated policy and inspected Go/TS outputs | Hand edits would drift from owner contracts | `intentional/no_action` | Contract generators | Change authored owner inputs and run Make-owned generation/drift targets only. |

No duplicated platform-config row/view-schema implementation, direct grid-vendor coupling, hidden mutation/revision/projection side effect, misplaced authorization check, or test-only package import in production was found. The filesystem-helper concern is instead production-private code whose only direct callers are tests.

## 6. Authorized Workstreams

Workstreams are sequential even where dependencies would permit parallel work. After each workstream, this tracker must be updated and validated before the next workstream changes any non-tracker file.

| Slice | Depends on | Authorized change | Exit evidence |
| --- | --- | --- | --- |
| PC-SL-00 | None | Activate this tracker and record fixed compatibility decisions. | Tracker-only diff and passing `make lint-markdown`. |
| PC-SL-01 | PC-SL-00 | Reconcile Network Flow, Core 04, Core 00, and filesystem owner requirements plus authored contract inputs. | Markdown/JSON checks and generated drift pass. |
| PC-SL-02 | PC-SL-01 | Add five omitted config rows, establish rooted-filesystem evidence ownership, and complete the production filesystem-effect inventory. | Sixteen config selectors resolve exactly once; harness and OTel gates pass. |
| PC-SL-03 | PC-SL-02 | Add immutable config catalog/snapshot and explicit validation phases behind temporary internal adapters. | Config owner, boundary, and fast tests pass. |
| PC-SL-04 | PC-SL-03 | Move telemetry configuration policy and mapping to the telemetry owner. | Config owner and OTel gates pass. |
| PC-SL-05 | PC-SL-04 | Move inactive-extension policy construction to Extensions and retain neutral config application. | Extensions and config owner gates pass. |
| PC-SL-06 | PC-SL-05 | Add race-resistant `rootedfs` and read-only `securefile` platform capabilities. | Security matrix and targeted security checks pass. |
| PC-SL-07 | PC-SL-06 | Move Network Flow configuration and manifest admission to its owner. | Config and Network Flow owner gates pass; RB-001 closes. |
| PC-SL-08 | PC-SL-07 | Move enterprise-auth configuration and manifest admission to its owner. | Auth, config, and startup gates pass. |
| PC-SL-09 | PC-SL-08 | Migrate incident-bundle paths, storage effects, schema, and tests to logical references and rooted ports. | Incident-bundle service-backed matrix and migration drift pass. |
| PC-SL-10 | PC-SL-09 | Migrate reference-pack paths, storage effects, schema, and tests to logical references and rooted ports. | Reference-data service-backed matrix and reset/reseed checks pass. |
| PC-SL-11 | PC-SL-10 | Migrate recovery, object store, and Postgres filesystem effects; close inventory and add bypass guards. | Affected owner gates pass; RB-003 closes. |
| PC-SL-12 | PC-SL-11 | Remove full config use from modules and HTTP transport through owner-local options. | No prohibited config import; app/server and owner gates pass. |
| PC-SL-13 | PC-SL-12 | Move public ceilings, owner limit DTOs, and narrow remaining platform adapter inputs. | View-query, adapter, boundary, and fast tests pass. |
| PC-SL-14 | PC-SL-13 | Remove legacy config/root helpers, temporary adapters, stale tests/artifacts, and update guides. | Zero legacy callers and clean lint/boundary/drift results. |
| PC-SL-15 | PC-SL-14 | Run focused-to-broad validation and complete handoff evidence. | All required gates pass and final tracker state is complete. |

## 7. Permanent Architecture and Slice Rules

- Core 04 and `internal/platform/config` own artifact selection, TOML and environment parsing, strict key admission, the deployment diagnostic envelope, root-binding admission, and validation-phase coordination.
- Application assembly supplies an immutable catalog of owner contributions. Contributions declare stable IDs and exact key ownership and perform only pure decode, default, normalization, and validation work.
- The ordered lifecycle is parse, claim overlays, inactive-key admission, key ownership and unknown-key rejection, decode/default/normalization, pure structural diagnostics, root capability construction, and application-owned preflight/startup.
- Typed snapshot values are projected only by application facades. Modules, HTTP transport, and platform adapters do not receive the complete snapshot.
- Telemetry, Extensions, Network Flow, and enterprise-auth configuration semantics live with their named owners.
- `internal/platform/rootedfs` enforces every root-relative production effect without exposing raw roots; `internal/platform/securefile` handles bounded read-only absolute manifests without following symlinks.
- Incident-bundle and reference-pack database records use strict relative logical references. Populated pre-release databases must be reset; there is no backfill, dual column, or dual read.
- Temporary compatibility is internal, recorded by exact caller count after every slice, and removed by PC-SL-14.
- Generated roots and topology are changed only through authored inputs and Make generation.

## 8. Validation Plan

| Layer | Command or rule | Required point |
| --- | --- | --- |
| Tracker/docs | `make lint-markdown` | Every tracker checkpoint and specification slice |
| Contract shape | `make json-shape-check` | Authored JSON contract/harness changes |
| Harness | `make harness-contract` | Owner/family changes |
| Generation | `make generate`, `make generate-drift`, `make generated-artifact-policy-check` | Authored inputs or generators |
| Migration | `make migration-drift` | Schema/reference slices |
| Boundary | `make backend-module-boundary-check` | Before and after boundary migrations |
| Telemetry | `make otel-conformance` | Telemetry/config slices |
| Focused owner | `make task-guide ROLE=module-author OWNER=<owner-id>`, `make explain-test-owner OWNER=<owner-id>`, then `make test-slice OWNER=<owner-id>` | Every affected owner |
| Service-backed owner | `make service-backed-test-slice OWNER=<owner-id>` | Incident bundles, reference data, recovery, object store, Postgres, and any discovered live writer |
| Security | `make go-gosec-targeted` | Rooted filesystem and consumer completion |
| Fast broad | `make test-fast` | Each structural implementation group |
| Finalization | `make agent-finalize` | Before final broad verification |
| Broad | `make check`, `make release-check` | PC-SL-15 |

If `RESULTS_DIR` is unset at finalization, retained-run maintenance is recorded as skipped for that reason. A skipped mandatory gate prevents completion.

## 9. Top-Level Work Tracker

| Slice | Status | Current evidence | Remaining exit condition |
| --- | --- | --- | --- |
| PC-SL-00 | DONE | Tracker-only activation; `make lint-markdown` passed; status/diff showed only this pre-existing staged tracker | Complete |
| PC-SL-01 | DONE | Network Flow `2.0.1`, Core 00 recognition, Core 04 inactive-diagnostic mapping and operation-time root contract, refreshed owner manifests/dependencies, and generated Go contracts; all required gates passed | Complete; exact runtime matrix remains PC-SL-07 evidence |
| PC-SL-02 | DONE | Five exact selector rows, five focused verification contracts, stable OTel evidence references, rooted-filesystem verification owner, generated topology, and the complete effect inventory below | Complete; sixteen selectors resolve once and all fourteen rows passed |
| PC-SL-03 | DONE | Immutable typed catalog/snapshot, deterministic contribution ordering and diagnostics, explicit phase order, application-owned materialization, and temporary defensive legacy projection | Complete; temporary projection remains explicitly bounded for later owner slices |
| PC-SL-04 | DONE | Telemetry-owned DTO/default/overlay/validation/endpoint/secret policy, typed catalog contribution, narrow bootstrap/instrumentation/transport mapping, and retained wire behavior | Complete; config DTO aliases remain temporary until the monolithic facade is removed |
| PC-SL-05 | DONE | Extensions-owned immutable inactive-configuration catalog, neutral config policy interface, application adaptation, owner tests/row, and no reverse config-to-module import | Complete; no legacy `extensioninactive` implementation or caller remains |
| PC-SL-06 | DONE | Linux descriptor-anchored `rootedfs`, immutable bounded `securefile`, strict logical references, race/identity rollback, active owner rows, and unsupported-platform fail-closed implementations | Complete; real owner consumers intentionally remain in PC-SL-07 through PC-SL-11 |
| PC-SL-07 | DONE | Owner-local Network Flow configuration/errors, app-owned catalog projection, immutable `securefile` manifest admission, exact inactive diagnostic/details, and no module platform-config/filesystem dependency | Complete; RB-001 closed |
| PC-SL-08 | DONE | Enterprise-auth-owned immutable configuration/error types, typed catalog projection, bounded owner document-reader port, application `securefile` preflight for manifests and referenced certificates, exact owner rows, and no owner config/filesystem dependency | Complete; active and inactive matrices, auth/config/startup/rooted gates, generation, boundary, security, and fast tests pass |
| PC-SL-09 | DONE | Owner-local logical references and storage port; rooted application adapter; atomic private staging/publication; migration `00037` reset gate and lexical checks; canonical owner, migration, security, harness, drift, boundary, and broad evidence | Complete; only the deliberately deferred archive-limit config import remains, with no storage-path or raw-filesystem compatibility surface |
| PC-SL-10 | DONE | Owner-local logical references and storage port; rooted application adapter; atomic private staging/publication; migration `00038` reset gate and lexical checks; activation/reverification/seeding/cancellation cleanup; canonical owner, migration, security, harness, drift, boundary, and broad evidence | Complete; only the deliberately deferred archive/reference-limit config imports remain, with no storage-path or raw-filesystem compatibility surface |
| PC-SL-11 | DONE | Recovery-owned rooted storage adapter, rooted filesystem object store, bounded/no-follow Postgres DSN and bootstrap reads, production-effect inventory, exact bypass guard, and affected owner matrices | Complete; RB-003 closed |
| PC-SL-12 | DONE | Owner-local module settings and limits, recovery deployment projection, narrow runtime process settings, deleted `httpapi.DependencySet.Config`, exact static guard, and complete affected-owner matrices | Complete; only test-support fixtures retain full config values |
| PC-SL-13 | DONE | Neutral archive-policy values, view-query-owned public ceilings, narrow Postgres/object-store/bootstrap/telemetry/recovery adapter inputs, application projections, adapter boundary rule, and focused mapping/signature evidence | Complete; retained monolithic config projection is isolated to application composition for PC-SL-14 removal |
| PC-SL-14 | DONE | Generic private document plus typed catalog/snapshot; app-only deployment projection; owner overlay/normalization; broad callers migrated; aliases/helpers, config `.gitkeep`, and stale guide language removed; final boundary rules active | Complete; zero prohibited imports, concrete-owner imports, compatibility aliases, or production raw storage mutations |
| PC-SL-15 | DONE | Complete owner matrix, security/drift gates, corrected stale telemetry test dependency, green focused retry, `test-fast`, `check`, and `release-check`, plus final status/symbol/import/path audit | Complete; no blocker or implementation follow-up remains |

## 10. Session Handoff Log

| Time | Slice | State | Files changed | Commands/result | Compatibility or blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-07-24T23:00:22-04:00 | PC-SL-00 | IN_PROGRESS | Tracker only | Baseline discovery complete; lint pending | Network Flow `2.0.1`/major 2; database reset; tracker was pre-existing staged work | Validate activation, then mark PC-SL-00 done and PC-SL-01 active |
| 2026-07-24T23:05:46-04:00 | PC-SL-00 | DONE | Tracker only | `make lint-markdown` PASS; status/diff confirmed no other path changed | Fixed decisions recorded; no blocker | PC-SL-01 marked active before owner edits |
| 2026-07-24T23:13:30-04:00 | PC-SL-01 | DONE | Network Flow/Core 00/Core 04 owner docs; extension owner fragments, manifests, dependencies; generated Go contracts; tracker | Markdown, JSON shape, generate, generate drift, and generated policy PASS; roots `20260725T031252Z-p44947`, `20260725T031257Z-p46589`, `20260725T031306Z-p49684`, `20260725T031308Z-p50049` | Initial generate stopped on stale authored digest at `20260725T030832Z-p38699`; refreshed authored owner inputs with the contract generator and reran successfully; no alias or major bump | PC-SL-02 marked active before harness edits |
| 2026-07-24T23:19:10-04:00 | PC-SL-02 | DONE | Config family/verification owner; rooted-filesystem verification owner/registry; OTel status; harness count test; generated topology; inventory; tracker | Shape, harness, generation/drift/policy, exact selector scan, config owner 14/14 at `20260725T031856Z-p66812`, and OTel at `20260725T031859Z-p67078` PASS; harness root `20260725T031813Z-p61846` | Shape first exposed registry ordering/profile errors; first harness run exposed the intentionally changed row-count fixture; both were corrected and rerun | RB-002 closed; PC-SL-03 marked active before config implementation |
| 2026-07-24T23:31:27-04:00 | PC-SL-03 | DONE | `internal/platform/config/catalog.go`, `snapshot.go`, and catalog tests; `internal/app/configassembly/configuration.go`; server/migrate/operator loading; config family, harness count, and generated topology; tracker | Format, config owner 15/15 at `20260725T033055Z-p42989`, app server 7/7 at `20260725T032739Z-p53498`, app operator 1/1 at `20260725T032739Z-p53502`, boundary at `20260725T032716Z-p43029`, generation/shape/harness/drift/policy at `20260725T032325Z-p75695` and `20260725T032337Z-p77531`, `-p77809`, `-p77543`, `-p77555`, and final fast 743 tests at `20260725T033101Z-p43446` PASS | One concurrent fast run at `20260725T032716Z-p43121` hit an unrelated frontend timing failure and passed alone on rerun; 24 non-test Go files still mention `config.Config`, while four facade call sites use the one defensive `LegacyConfig` projection; no direct facade loader call remains | PC-SL-04 marked active before telemetry-owner edits |
| 2026-07-24T23:51:34-04:00 | PC-SL-04 | DONE | `internal/platform/telemetry/configuration/*` and owner catalog/preflight adapter; telemetry runtime/resource/retry/exporter/self-metrics plus tests; config DTO aliases and neutral delegation; app runtime, HTTP telemetry settings, collaboration/workbook mappings; tracker | Format at `20260725T034747Z-p15365`, config owner 15/15 at `20260725T034757Z-p18057`, OTel conformance at `20260725T034757Z-p17929`, boundary at `20260725T034757Z-p18219`, backend unit 312 tests at `20260725T034809Z-p21920`, app server 7/7 at `20260725T034809Z-p21925`, and fast 743 tests at `20260725T034941Z-p58959` PASS | `platform.telemetry` is a verification-contract owner but not an active test-family owner, so task-guide/explain correctly reported it unavailable; canonical config rows, OTel conformance, backend unit, and app-server execution cover the change; 21 non-test Go files still mention `config.Config`; temporary config telemetry type aliases remain for later facade removal | PC-SL-05 marked active before Extensions policy edits |
| 2026-07-25T00:06:21-04:00 | PC-SL-05 | DONE | `internal/platform/config/inactive_policy.go` and neutral load/validation integration; `internal/modules/extensions/inactive_configuration_catalog.go` plus owner tests; application extension assembly and server/migrate/operator mappings; Extensions test family/contract count and generated topology; removal of `internal/platform/config/extensioninactive`; tracker | Format at `20260725T040347Z-p93695`; config owner 15/15 at `20260725T040357Z-p96522`; Extensions owner 26 tests at `20260725T040357Z-p96529`; app server 29 tests at `20260725T040157Z-p66178`; boundary at `20260725T040357Z-p96721`; fast 744 tests at `20260725T040357Z-p96853`; JSON, harness, generation/drift, and generated-policy gates passed in the `20260725T035802Z`–`20260725T035813Z` runs | The first Extensions accounting run failed because its package contract still expected 25 rows (`20260725T035852Z-p23464`); the corrected run passed. The first boundary run caught config tests importing the owner module (`20260725T040157Z-p66263`); neutral test policy support removed the reverse dependency and the rerun passed. Wire admission, generic reason/message/details, discard semantics, and fail-before-preflight behavior remain unchanged. No `extensioninactive`, `ExtensionInactiveCatalog`, or old validation entry point remains; 21 non-test files still mention `config.Config`/`Snapshot` for later slices. | PC-SL-06 marked active only after the corrected required gates and this tracker checkpoint |
| 2026-07-25T00:23:06-04:00 | PC-SL-06 | DONE | `internal/platform/rootedfs/{reference,root_linux,root_unsupported,root_linux_test}.go`; `internal/platform/securefile/{document,read_linux,read_unsupported,read_linux_test}.go`; rooted owner registry/family/count inputs; Make-generated topology; tracker | Rooted owner 2/2 rows at `20260725T042247Z-p76779`; targeted gosec at `20260725T042250Z-p77276`; JSON shape `20260725T041839Z-p97024`; harness 91/91 at `20260725T041839Z-p97507`; generation `20260725T041818Z-p94528`, drift `20260725T041839Z-p97009`, policy `20260725T041839Z-p97054`, boundary `20260725T041839Z-p97498`, and isolated fast 751 tests at `20260725T041954Z-p29393` PASS | The initial owner activation correctly reported no active owner before its executable manifest existed. First catalog runs found unsorted selectors (`20260725T041453Z-p77461`, `-p77599`) and then stale topology (`20260725T041521Z-p79028`); sorting and Make generation corrected both. First owner execution found one unused test import (`20260725T041601Z-p82825`); corrected owner runs passed. A concurrent fast run failed only in unchanged WorkbookShell frontend timing (`20260725T041839Z-p97707`); the isolated rerun passed. The capability is intentionally Linux-specific; every other platform constructor fails with `ErrUnsupportedPlatform`, never an insecure fallback. Old helpers and real consumers remain until PC-SL-07 through PC-SL-11, so RB-003 remains open. | PC-SL-07 marked active only after this checkpoint; migrate Network Flow manifest admission through `securefile` |
| 2026-07-25T00:51:58-04:00 | PC-SL-07 | DONE | Network Flow owner configuration/finding/error types and tests; key-ring parsing from immutable bytes with no config or filesystem import; application catalog projection/test and runtime `securefile` preflight; typed secure-file failure classification; generic inactive message/details; config/app/module owner rows and generated topology; tracker | Config 16/16 at `20260725T043548Z-p16033`; rooted 2/2 at `20260725T043548Z-p15827`; app/server 29 at `20260725T043958Z-p81104`; Network Flow 119 at `20260725T044053Z-p7365`; Network Flow service-backed 20 at `20260725T044525Z-p39848`; backend unit 323 at `20260725T043456Z-p5757`; gosec at `20260725T044526Z-p39999`; fast 754 at `20260725T044956Z-p93124`; JSON, harness, generation/drift/policy, and boundary at `20260725T043532Z`–`20260725T043548Z` PASS | Concurrent app/server and Network Flow owner runs contended for the same browser service names (`20260725T043548Z-p15919`, `-p15877`); all product partitions passed, and sequential full reruns closed both owners. Unclaimed file/overlay values now emit the adopted exact message plus `profile_id` and `config_path`, before active grammar or `securefile`; `profile_incompatible_binding` is absent from this path. Claimed missing, invalid, unreadable, directory, oversized, symlinked, and malformed cases are owner-defined and fail before readiness without host-path disclosure. The internal legacy loader/parser APIs were removed, not aliased. Twenty non-test files still mention `config.Config`/`Snapshot` for later slices. | RB-001 closed; PC-SL-08 marked active only after this checkpoint |
| 2026-07-25T01:24:33-04:00 | PC-SL-08 | DONE | Core 04 manifest/certificate no-follow and byte-bound requirements; refreshed Core 04 extension owner inputs; corrected enterprise configuration contract reference; enterpriseauth configuration/finding/error and document-reader port; manifest/parser tests; application catalog projection and `securefile` adapter; server fail-before-readiness tests; config/auth/app family inputs, auth verification wording, harness counts, generated contracts/topology, and tracker | Format `20260725T051855Z-p46569`; generation `20260725T052047Z-p55020`; JSON `20260725T052053Z-p56667`; harness `20260725T052053Z-p56735`; config full 17/17 `20260725T050746Z-p71038` plus final owner projection `20260725T052135Z-p58619`; auth full 49 `20260725T050801Z-p73184`, service-backed 29 `20260725T051250Z-p8509`, and final owner row `20260725T052126Z-p58219`; app startup row `20260725T051815Z-p45110`; rooted 2/2 `20260725T051825Z-p46005`; drift `20260725T052144Z-p59040`; generated policy `-p59042`; boundary `-p59102`; gosec `-p59138`; fast 763 `-p59130` PASS | `enterpriseauth` no longer imports config/rootedfs/securefile or performs manifest filesystem access. Inactive configuration performs no document read; active missing, invalid, unavailable, directory, oversized, symlinked, malformed, referenced-symlink, and referenced-oversize cases fail in owner diagnostics without host paths. Startup order for provider secret registration and reconciliation is unchanged. The manifest is capped at `1048576` bytes and each certificate at `262144` bytes under existing reason-code families. Initial generation stopped on the intentionally stale Core 04 owner digest (`20260725T050502Z-p60196`) and passed after contract-owner refresh. The first app startup row exposed malformed JSON being preempted by duplicate scanning (`20260725T051733Z-p41605`); syntax validation now precedes duplicate-member validation and the rerun passed. Nineteen non-test files still mention `config.Config`/`Snapshot`; no compatibility alias was added. | PC-SL-09 marked active only after this checkpoint; migrate incident-bundle storage and database references |
| 2026-07-25T02:09:15-04:00 | PC-SL-09 | DONE | `db/migrations/00037_incident_bundle_storage_references.sql`; `internal/modules/incidentbundles/{storage_port,routes,store,worker_service}.go` and owner tests; removal of `bundle_files.go`; `internal/app/server/{incident_bundle_storage,incident_bundle_storage_test,runtime}.go`; `internal/platform/rootedfs/{root_linux,root_unsupported,root_linux_test}.go`; Postgres migration tests; dev guide; migration/test-family/boundary/baseline authored inputs; Make-generated SQL models, catalog, scheduler, and topology index; tracker | Format `20260725T055511Z-p32279`; incident owner 22/22 `20260725T054835Z-p57653` and service-backed 11/11 `20260725T054804Z-p56321`; migration row `20260725T054743Z-p55896`; rooted owner and adapter rows `20260725T055515Z-p34871`, `20260725T055518Z-p35286`; app bootstrap rows `20260725T055600Z-p36398`; migration drift `20260725T060617Z-p87792`; full service-backed 167/167 `20260725T060144Z-p13489`; duration refresh/coverage `20260725T060515Z-p85866`, `20260725T060531Z-p86208`; harness `20260725T060537Z-p86489`; generation `20260725T060628Z-p93738`; drift/JSON `20260725T060638Z-p95479`, `-p95489`; generated policy `20260725T060617Z-p87761`; boundary `20260725T060853Z-p65650`; gosec `20260725T060651Z-p99600`; fast 767 tests `20260725T060651Z-p99587` PASS | Existing incident rows deliberately block migration with “development database reset required”; reset/reseed is mandatory, with no backfill, dual column, or dual read. Stored and payload values are strict relative POSIX references; owners never receive roots or absolute paths. Staging and publication are descriptor-anchored, private, exclusive/atomic, cleanup-safe, root-identity checked, and fail without path disclosure. Initial migration drift found the unregistered migration and Goose statement boundary; both authored inputs were corrected. The first broad run exposed absent valid roots and was fixed by secure no-follow `OpenOrCreate`; the next passed every work unit but revealed the new migration row was misclassified as a template clone. Correct migration-scratch ownership made the final broad run and budget pass. The boundary guard then required the new adapter’s exact file allowlist; no wildcard was added. Remaining incident legacy scan finds only the archive-limit config import intentionally deferred to PC-SL-12/13; old path spellings remain solely in migration history and the still-unmigrated reference-pack owner. | PC-SL-10 marked active only after this checkpoint; migrate reference-pack storage and database references |
| 2026-07-25T02:39:37-04:00 | PC-SL-10 | DONE | `db/migrations/00038_reference_pack_storage_references.sql`; `internal/modules/reference_data/{storage_port,storage_port_test,api,store,routes,minimum_disconnected_bundle,job_finalization,reference_pack_integration_test}.go`; `internal/app/server/{reference_pack_storage,reference_pack_storage_test,runtime}.go`; `internal/app/extensionassembly/reference_pack_jobs.go`; Postgres migration tests; dev guide; migration/test-family/boundary/baseline authored inputs; Make-generated SQL models, catalog, scheduler, and topology index; tracker | Format `20260725T063435Z-p41051`; backend unit 337 tests `20260725T061919Z-p87294`; logical-reference/adapter rows `20260725T061943Z-p91327`; full service-backed baseline `20260725T062125Z-p98732`; duration refresh/coverage `20260725T062458Z-p81475`, `20260725T062712Z-p17756`; reference-data owner 19/19 `20260725T062505Z-p81707` and service-backed 11/11 `20260725T063133Z-p75902`; activation, cancellation, and symlink entry-point rows `20260725T062909Z-p52620`, `20260725T063122Z-p75312`, `20260725T063442Z-p43688`; generation `20260725T062700Z-p15799`; JSON/harness/drift/policy `20260725T062712Z-p17651`, `-p17912`, `20260725T063912Z-p20334`, `20260725T062747Z-p22712`; migration drift `20260725T063912Z-p20378`; boundary `20260725T063912Z-p20664`; gosec `20260725T063912Z-p20854`; isolated fast 772 tests `20260725T063544Z-p49702` PASS | Existing reference-pack rows deliberately block migration with “development database reset required”; reset/reseed is mandatory, with no backfill, dual column, or dual read. Stored/job values are strict relative POSIX references; owner code receives no root or absolute path. Publication is private, descriptor-anchored, exclusive/atomic, bounded, root-identity checked, and cleanup-safe across duplicate admission, cancellation, precommit failure, and deterministic finalizer failure; indeterminate commits retain recoverable artifacts. Activation and reverification read the persisted object through the real port. Initial focused migration execution found a stale scratch schema, activation characterization exposed a test-owned missing artifact, and the cancellation matrix exposed a pre-execution cleanup gap; rebuilt schema evidence, test restoration, and owner cleanup fixed each. An exact extension-assembly boundary allow was required; no broad wildcard was added. One broad fast run hit only the unchanged intermittent WorkbookShell timing case (`20260725T063454Z-p44205`); isolated rerun passed. Production scans find no old reference-pack path spelling, raw filesystem effect, or rooted/secure package import; two config imports remain solely for limit DTOs deferred to PC-SL-12/13. | PC-SL-11 marked active only after this checkpoint; migrate and characterize all remaining filesystem consumers and close RB-003 |
| 2026-07-25T03:15:34-04:00 | PC-SL-11 | DONE | Core 04 bootstrap-manifest requirements and refreshed extension-owner inputs; `internal/platform/rootedfs/{reference,root_linux,root_unsupported,root_linux_test,production_boundary_test}.go`; `internal/platform/objectstore/{objectstore,objectstore_test}.go`; `internal/platform/postgres/{postgres,postgres_settings_test}.go`; `internal/platform/bootstrap/bootstrap.go`; bootstrap startup integration test and oversized golden; `internal/modules/recovery/{capture,encryption}.go`, recovery tests, and `operatorops/operations.go`; `internal/app/recoveryassembly/{storage,storage_test}.go`; operator/serverprocess/recovery-browser callers; rooted/recovery/Postgres family inputs, harness counts, and Make-generated topology/catalog outputs; obsolete config root-write helpers/tests; tracker | Format `20260725T070523Z-p72040`; backend unit 339 `20260725T065809Z-p95626`; rooted owner 2/2 `20260725T070122Z-p31856`; object-store owner 2/2 `20260725T070228Z-p37186` and service-backed 2/2 `20260725T070819Z-p20811`; recovery owner 11/11 `20260725T070859Z-p21966` and service-backed 9/9 `20260725T070527Z-p74645`; Postgres owner 4/4 `20260725T071112Z-p75454` and service-backed 2/2 `20260725T070838Z-p21355`; Imports service-backed 5/5 `20260725T070643Z-p939`; app bootstrap row `20260725T070329Z-p44463`; generation `20260725T070217Z-p35438`; JSON `20260725T071030Z-p47922`; harness PASS; drift `20260725T071030Z-p48142`; generated policy `20260725T071030Z-p47937`; boundary `20260725T071030Z-p48401`; gosec `20260725T071030Z-p48498`; migration drift `20260725T071135Z-p76201`; fast 775 `20260725T071135Z-p76433` PASS | Filesystem object-store keys and recovery references are now strict relative POSIX references; managed bindings never fall back locally. All real owner operations are descriptor anchored, bounded where read-only, no-follow, identity checked, cleanup safe, and path redacted. Bootstrap and Postgres preserve admitted settings and reason families while rejecting symlinks, hard links, special files, replacement, and oversize input. Imports remain byte/in-memory only; Reporting has no live filesystem writer, so no reporting owner execution was required. Raw product effects are limited by an exact AST guard to the repository-local Postgres migration-evidence read and Goose artifact log. The first backend run exhausted `/tmp`; only the reproducible `/tmp/cartulary-go-build` cache was removed, with no source or retained data affected. Initial generation/shape failures were expected stale Core 04 digest/topology input and passed after Make-owned refresh. First bootstrap and recovery runs exposed an existing message-compatibility expectation and a removed-helper import; both were corrected and rerun. No obsolete config root helper or unguarded storage consumer remains; RB-003 is closed. | PC-SL-12 marked active only after this checkpoint; replace module and HTTP full-config dependencies with owner-local options |
| 2026-07-25T04:16:27-04:00 | PC-SL-12 | DONE | `internal/platform/httpapi/httpapi.go`; owner-local route/settings and limit types in Auth, Collaboration, Evidence, Imports, Incident Bundles, and Reference Data; `internal/modules/recovery/operatorops/operations.go`; `internal/app/operator/operator_recovery.go`; `internal/app/server/{module_settings,module_settings_test,runtime,server}.go`; harness-runtime reset callback and tests; Network Flow harness-control helper; `internal/testutil/httptestx/httptestx.go`; affected owner tests; backend boundary input; app-server family/harness-count input; Make-generated catalog/topology; tracker | Format `20260725T080750Z-p44770`; backend unit 341 `20260725T080754Z-p47392`; app server 29/29 `20260725T080900Z-p52354` and service-backed 22/22 `20260725T080955Z-p79387`; Auth 49/49 `20260725T073229Z-p25492` and service-backed 29/29 `20260725T073708Z-p59156`; Collaboration 27/27 `20260725T074153Z-p91773` and service-backed 18/18 `20260725T074611Z-p24262`; Evidence 34/34 `20260725T075029Z-p54210` and service-backed 29/29 `20260725T075328Z-p81537`; Imports 5/5 `20260725T075626Z-p8394` and service-backed 5/5 `20260725T075802Z-p27954`; Incident Bundles 22/22 `20260725T075932Z-p47257` and final service-backed 11/11 `20260725T081055Z-p5515`; Reference Data 19/19 `20260725T080039Z-p50636` and final service-backed 11/11 `20260725T081122Z-p6640`; Recovery 11/11 `20260725T080235Z-p83812` and final service-backed 9/9 `20260725T081210Z-p23240`; Operator 2/2 `20260725T080443Z-p35958` and service-backed 1/1 `20260725T080457Z-p36513`; four affected Network Flow harness reset rows `20260725T080816Z-p51746`; generation `20260725T072955Z-p66869`; final JSON `20260725T081325Z-p49066`; harness PASS; drift `20260725T081354Z-p50698`; generated policy `20260725T081403Z-p53786`; boundary `20260725T081406Z-p54187`; OTel `20260725T081408Z-p54440`; gosec `20260725T081413Z-p57696`; fast 776 `20260725T081428Z-p78449` PASS | Wire keys, defaults, numeric values, public-origin behavior, telemetry service version, readiness, route behavior, and resource ownership are unchanged. Modules receive owner-local value objects; recovery receives a narrow deployment projection with prebound application factories; HTTP transport has no config field; Runtime retains only four process settings. No compatibility alias was added. The only non-test-file module config import is the reusable Workbook test-support harness, which the production-only guard excludes explicitly; production modules and HTTP have none. Initial backend execution found Incident Bundle tests still constructing config DTOs (`20260725T072235Z-p39389`); owner-local fixtures fixed it. The first boundary run required exact application-assembly paths (`20260725T072902Z-p65831`), with no wildcard. Removing Runtime’s full config exposed test-only artifact-path reads (`20260725T080643Z-p40403`); the test harness now retains its own config copy, while production Runtime does not. Remaining full-config adapter APIs are explicitly PC-SL-13 scope; the temporary application `LegacyConfig` projection is PC-SL-14 scope. | PC-SL-13 marked active only after this checkpoint; move remaining ceilings and replace full-config adapter inputs |
| 2026-07-25T04:50:50-04:00 | PC-SL-13 | DONE | `internal/platform/archivepolicy/limits.go`; owner-local limit aliases in Imports, Incident Bundles, and Reference Data; view-query ceilings/tests; narrow Postgres, object-store, bootstrap, migration-evidence, telemetry-instrumentation, and recovery assembly inputs; app config projections plus migrate/operator/server callers and test support; adapter boundary input; config/Postgres family rows and harness count; Make-generated catalog/topology; tracker | Format `20260725T083927Z-p17246`; config 18/18 `20260725T083637Z-p92767`; view-query 6/6 `20260725T083653Z-p93783`; Postgres final 5/5 `20260725T084109Z-p26384`; object store 2/2 `20260725T083823Z-p14428`; bootstrap 2/2 `20260725T083841Z-p15476`; operator final 2/2 `20260725T083930Z-p19869`; app server 29/29 `20260725T083151Z-p49871`; Imports 5/5 `20260725T084201Z-p30628`; Incident Bundles 22/22 `20260725T084342Z-p51159`; Reference Data 19/19 `20260725T084529Z-p53125`; Recovery 11/11 `20260725T084631Z-p70717`; backend unit 349 `20260725T084131Z-p27246`; build migrate/operator/server `20260725T084733Z-p97177`, `-p98847`, `20260725T084743Z-p7875`; generation `20260725T084037Z-p23555`; JSON `20260725T084758Z-p16357`; harness PASS; drift `20260725T084801Z-p16854`; generated policy `20260725T084810Z-p19946`; boundary `20260725T084812Z-p20347`; fast 784 `20260725T084815Z-p20715` PASS | Sort `8` and filter `16` are unchanged and owned by view-query. Dead change/action constants were removed after the live-call scan found no production owner. Archive values are a neutral shared value object; no module imports config limit DTOs. Postgres, object store, bootstrap, telemetry bootstrap, enterprise-auth manifest admission, migration evidence, and recovery storage accept narrow owner settings/ports; no adapter accepts `config.Config` or `config.Snapshot`. Borrowed resource and reverse-order/idempotent cleanup evidence remains green. The first backend run exposed a server fixture without the newly enforced root-bound DSN; the corrected fixture exercises the secure read. The first operator slice exposed a missing narrowed-type import and passed after correction. No old adapter symbol or config-owned public ceiling remains. | PC-SL-14 marked active only after this checkpoint; delete the application legacy projection, config aliases, obsolete tests/artifacts, and stale guide language |
| 2026-07-25T05:32:53-04:00 | PC-SL-14 | DONE | `internal/platform/config/{config,catalog,snapshot,validation,telemetry_document}.go` and tests; `internal/platform/telemetry/{configuration,bootstrap_test}.go`; `internal/app/configassembly/**`; server/migrate/operator composition and tests; config/recovery/object-store test support and recovery-browser tool; backend boundary input; developer guide; removed config `.gitkeep` and test-only legacy facade; tracker | Format `20260725T092737Z-p11437`; config owner 18/18 `20260725T092830Z-p23735`; app server 29/29 `20260725T092838Z-p25851`; backend unit 349 `20260725T092407Z-p81264`; build server/migrate/operator `20260725T092432Z-p89970`, `20260725T092441Z-p98677`, `20260725T092444Z-p731`; JSON `20260725T092749Z-p15680`; boundary `20260725T092752Z-p16233`; generated policy `20260725T092755Z-p16524`; drift `20260725T092757Z-p16863`; migration drift `20260725T092806Z-p19977`; final fast 784 `20260725T092940Z-p54257` PASS; Markdown PASS | Wire keys/defaults/diagnostics remain stable. Telemetry supplies overlay/default/validation policy through its contribution; generic config has no concrete-owner import. No monolithic `Config`, `LegacyConfig`, test-only alias, exported legacy load/validate helper, inactive-policy package, full adapter input, module/HTTP platform import, or owner raw filesystem mutation remains. Old database column names occur only in append-only migration history and cutover tests. The first post-extraction config run `20260725T092319Z-p75459` exposed owner TOML tags with options not handled by overlay lookup; comma-option parsing was fixed and all reruns passed. `docs/domain.md` was reviewed and requires no vocabulary edit. | PC-SL-15 marked active only after this checkpoint; run the mandatory final owner, security, broad, release, and handoff gates |
| 2026-07-25T05:59:05-04:00 | PC-SL-15 | FAILED | Tracker only; retained failure inspection | Config 18/18 `20260725T093434Z-p13783`; rootedfs 2/2 `20260725T093439Z-p14566`; Extensions 26/26 `20260725T093444Z-p15347`; Network Flow 119/119 `20260725T093542Z-p35189` and service-backed 20/20 `20260725T094000Z-p68910`; Auth 49/49 `20260725T094420Z-p2562` and service-backed 29/29 `20260725T094853Z-p36541`; Incident Bundles 22/22 `20260725T095334Z-p70123` and service-backed 11/11 `20260725T095410Z-p71521`; Reference Data 19/19 `20260725T095447Z-p72660` and service-backed 11/11 `20260725T095619Z-p89883` PASS. Recovery stopped at `20260725T095714Z-p7122`: four of five work units passed; the browser restore target failed before readiness. | Retained stderr proves the direct cause: `tools/recoverybrowserrestore.targetConfig` constructs an in-memory deployment with a populated `limits` table but omits the two required Extensions limit defaults, so strict admission observes explicit zero values. No later owner or broad gate started. Resume only after the tracker records the same-workstream correction scope. |
| 2026-07-25T06:11:59-04:00 | PC-SL-15 | FAILED | Tracker only; retained final-gate inspection | Recovery correction format `20260725T100049Z-p37763`; Recovery 11/11 `20260725T100054Z-p40810` and service-backed 9/9 `20260725T100208Z-p70620`; config rerun 18/18 `20260725T100330Z-p96914`; Extensions service-backed 12/12 `20260725T100335Z-p97808`; object store 2/2 `20260725T100431Z-p17459` and service-backed 2/2 `20260725T100443Z-p18087`; Postgres 5/5 `20260725T100456Z-p19000` and service-backed 2/2 `20260725T100514Z-p19749`; imports 5/5 `20260725T100537Z-p20708` and service-backed 5/5 `20260725T100711Z-p40532` PASS. Markdown, JSON `20260725T100851Z-p61748`, harness contract, generation `20260725T100920Z-p63403`, generation drift `20260725T100925Z-p65051`, generated policy `20260725T100935Z-p68163`, migration drift `20260725T100937Z-p68548`, boundary `20260725T100945Z-p70762`, OTel `20260725T100947Z-p71025`, and targeted gosec `20260725T100952Z-p74281` PASS. Final `test-fast` failed at `20260725T101001Z-p96220`; its backend unit surface passed 349 tests, while frontend unit had one failure in `WorkbookShell.surfaces.test.tsx` (`keeps party-link mutations syncing until workbook and references refresh`) because the generic edit record test node was not observed. | The failed frontend file is unchanged and this exact intermittent assertion passed on isolated reruns earlier in the effort. No production correction is authorized from this evidence. `agent-finalize`, `check`, and `release-check` have not started. Resume only after recording an isolated `frontend-unit` retry followed by `test-fast`. |
| 2026-07-25T06:12:47-04:00 | PC-SL-15 | IN_PROGRESS | Tracker only | Markdown validation passed after the required failure checkpoint. | Resume is limited to `make frontend-unit`, then `make test-fast`; no production source change is planned. |
| 2026-07-25T06:17:23-04:00 | PC-SL-15 | FAILED | Tracker only; retained broad-gate evidence | Isolated frontend unit passed at `20260725T101320Z-p10642`; repeated `test-fast` passed 784 tests at `20260725T101345Z-p12560`; `agent-finalize` passed unchanged at `20260725T101602Z-p62935` with retained-run maintenance skipped because `RESULTS_DIR` was unset. Broad `check` failed at `20260725T101625Z-p68223` after 38 of 155 scheduled work units; the sole failed work unit was `go-vulncheck`, reported as `govulncheck: loading packages`. | The scheduler skipped the remaining 117 dependent work units. `release-check` has not started. Diagnose the retained vulnerability log before deciding whether the same slice can resume. |
| 2026-07-25T06:18:10-04:00 | PC-SL-15 | IN_PROGRESS | Tracker only | Retained stderr identifies six compile references in `internal/platform/telemetry/exporter_test.go` to the deliberately removed `config.TelemetryExporterRetryConfig`; production telemetry already accepts `telemetry/configuration.ExporterRetryConfig`. This is a stale test dependency, not a vulnerability finding or product failure. | Resume is limited to migrating that test to its owner-local type, formatting, running focused telemetry/config verification and `go-vulncheck`, then repeating `agent-finalize` and `check`; `release-check` remains gated. |
| 2026-07-25T06:27:15-04:00 | PC-SL-15 | DONE | `internal/platform/telemetry/exporter_test.go`; final tracker; complete implementation status/diff inspected across 162 entries | Stale test migration format `20260725T101902Z-p14185`; backend unit 349 tests `20260725T101913Z-p16888`; OTel `20260725T101913Z-p16838`; vulnerability scan `20260725T101913Z-p16960`; repeated `agent-finalize` unchanged `20260725T101929Z-p22591`; broad `check` 155/155 work units and 644 tests `20260725T101937Z-p23089`; `release-check` 12/12 work units and 644 tests `20260725T102155Z-p14728` PASS. Final Markdown, `git diff --check`, exact imports, legacy exports, current schema, raw-effect, generated-lock, and status scans PASS. | Make generation provenance is `20260725T100920Z-p63403`; generated drift and policy passed at `20260725T100925Z-p65051` and `20260725T100935Z-p68163`. Generated Go contract and SQL model changes are downstream of owner inputs and migrations; no lock file or generated topology was hand-edited. Retained-run maintenance was skipped because `RESULTS_DIR` was unset before the first successful full warm check. Old storage-path spellings remain only in append-only migration history, reversible migration clauses, and cutover tests. Raw filesystem calls found by the audit are test fixtures or the exact repo-runner Postgres log allowlist; the boundary gate prohibits owner bypasses. The two reset-gated migrations intentionally require `make db-reset` and reseeding for affected pre-release data. No absolute stored reference, compatibility alias, broad module/HTTP config import, raw owner bypass, stale generated-policy change, or unresolved RB remains. No next action is required beyond normal review/commit. |

## 11. Open Questions and Blockers

| ID | Required closure | Status |
| --- | --- | --- |
| RB-001 | Adopt Network Flow `2.0.1`, generic `extension_config_without_claim`, aligned Core 00/Core 04/contracts/generation, and exact executable matrix. | CLOSED |
| RB-002 | Add five independent active `platform.config` selectors, stable verification references, exact-once accounting, owner execution, and OTel conformance. | CLOSED |
| RB-003 | Complete effect inventory, adopt operation-time requirements, implement rooted capability, migrate every real consumer, pass entry-point security matrices and static guards, and remove obsolete helpers. | CLOSED |

Fixed RB-001 result:

```text
error.code = invalid_deployment_config
reason_code = extension_config_without_claim
message = "Extension configuration is present while the profile is inactive."
details.profile_id = "network_flow_activity"
details.config_path = "$.network_flow_activity.key_ring_manifest_path"
```

This result precedes active-value validation and all file, secret, Network Flow, and egress effects. `profile_incompatible_binding` is not an alias.

## 12. Binary Completion Criteria

| Criterion | Status | Required evidence |
| --- | --- | --- |
| Tracker protocol followed for every slice | DONE | Sequential `IN_PROGRESS`/`DONE` entries, explicit failure/resume checkpoints, and tracker validation before the next slice |
| RB-001 owner contradiction closed | DONE | Adopted owners, generated contracts, exact generic diagnostic/details, no-file precedence, and config/app/Network Flow owner evidence |
| RB-002 executable accounting closed | DONE | Sixteen exact-once selectors and successful retained owner rows |
| RB-003 production containment closed | DONE | Real consumer matrices, static guard, and no obsolete helper |
| Config facade is generic and owner-extensible | DONE | Typed catalog/snapshot tests, private wire document, owner overlay dispatch, and no production concrete-owner import |
| Modules/transport/adapters use narrow inputs | DONE | Owner-local options, adapter signature/mapping tests, import/call scans, and boundary gate |
| Persistent artifact locations are logical references | DONE | Migrations `00037`/`00038`, reset gates, owner ports, database/payload/log scans, and incident/reference-data owner evidence |
| No compatibility aliases or stale surfaces remain | DONE | Symbol/import/call scans, removed `.gitkeep`, and cleanup diff |
| Required focused, security, drift, broad, and release gates pass | DONE | Exact commands, results, and run roots in Section 10; final `check` and `release-check` are green |
| Final handoff is reproducible and unambiguous | DONE | Final status/diff and invariant scans, generated provenance, reset impact, skipped retained-run rationale, and no remaining next action |

### PC-SL-02 Production Filesystem-Effect Inventory

| Owner or call path | Current production effect | Required remediation slice |
| --- | --- | --- |
| Imports XLSX parsing | Workbook/archive bytes remain in memory; no temporary file or extraction path is passed to a library. | No rooted storage implementation; retain archive admission tests and the future raw-effect guard. |
| Incident bundles | PC-SL-09 routes every production staging/read/publication/cleanup effect through an owner port backed by two `rootedfs` capabilities; records and job payloads contain strict relative references only. | Complete. PC-SL-13 moved archive values to the neutral archive-policy object and removed the final config DTO dependency. |
| Reference Pack | PC-SL-10 owner ports stage, publish, verify, activate, seed, and clean up through rooted capabilities; retained pack/job rows store only strict relative references. | Complete. Migration `00038`, owner and adapter matrices, reset/reseed evidence, static scans, and PC-SL-13 limit ownership all pass. |
| Recovery | Completed in PC-SL-11: owner code consumes an opaque storage port; application assembly supplies bounded, strict-reference, descriptor-anchored storage with no managed-service fallback. | Complete; actual owner and operator-process matrices pass. |
| Filesystem object store | Completed in PC-SL-11: all key reads, writes, listings, deletion, and replacement use `rootedfs`; no `os.Root` or raw fallback remains. | Complete; actual adapter matrices and boundary guard pass. |
| Filesystem-root Postgres binding | Completed in PC-SL-11: the root-relative DSN is a bounded, no-follow, regular single-link read with redacted errors. | Complete; focused and service-backed Postgres owner matrices pass. |
| Bootstrap, enterprise-auth, and Network Flow manifests | Completed in PC-SL-07, PC-SL-08, and PC-SL-11: application assembly supplies immutable bounded `securefile` bytes; owners never reopen host paths. | Complete; inactive and active startup matrices pass. |
| Reporting/preview/export builders | No live operator-root writer exists; reporting materialization is database/job-payload based and has no filesystem/archive import. | No rooted adapter is required; the exact production bypass guard makes any future writer an explicit boundary change. |
| External libraries receiving output paths | No current import/report library receives an operator-controlled output path; Imports XLSX and report paths remain byte/in-memory only. | Complete inventory classification; the static guard and review rule treat any future path-taking call as a filesystem effect. |

`platform.rootedfs` verification ownership is now registered. Its active test-family owner and exact selectors are created with the package tests in PC-SL-06; the Harness schema permits only active owners with resolvable nonempty rows, so a non-executable placeholder family is intentionally forbidden.

## Appendix A. Superseded Planning-Only Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish authority, repository snapshot, write scope, and initial handoff ledger | This tracker and owner documents only | `make lint-markdown`; diff/status inspection | Tracker exists, source order is recorded, and only the tracker is changed. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-04 | Account for every target file, export, importer, dependency, and test | `internal/platform/config/**` and direct callers | Repository inventory/import/symbol searches | All nine files have concrete Section 2 rows. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-03, WF-05 | Freeze observable behavior and map each contract to its owner | Owner docs, config code, callers, contract inputs | Source/contract inspection; `make explain-test-owner OWNER=platform.config` | Every discovered risk has evidence, owner, and test posture; RB-001 remains explicit. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-05, WF-07 | Separate executable characterization from static and accounting evidence | Target/caller tests, test-family and verification manifests | `make test-slice OWNER=platform.config`; `make otel-conformance` | RB-002 and RB-003 are closed or affected slices remain blocked. |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-05 | Identify broad config propagation, owner-policy placement, storage/transport adjacency, and generated/test leaks | Target, app composition, platform/module consumers | `make backend-module-boundary-check` | Each finding has a classification and planning action. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Specify a narrow generic config facade and strict owner contributions while preserving public compatibility | Config, app assembly, named owner packages | Design review against Sections 3 through 5 | Interfaces, diagnostic aggregation, validation order, and compatibility sequence are decision-complete. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order reversible behavior-preserving implementation slices | Packages listed in Section 7 | Slice-specific Make targets | Every slice has dependencies, risks, tests, validation, rollback, and a binary exit condition. |
| WF-07 | Harness/test/accounting update plan | chain | WF-03, WF-06 | WF-08 | Add missing authored owner rows and regenerate downstream artifacts through Make only | Authored test-family/verification owners; generated outputs downstream | `make harness-contract`; `make generate-drift`; owner slice | All live tests are canonically accounted for without hand-editing generated files. |
| WF-08 | Validation and final handoff | chain | WF-02 through WF-07 | none | Run focused-to-broad validation and leave a resumable result ledger | Tracker plus later authorized implementation diff | Section 8 commands | Results, failures, skips, blockers, and next action are recorded. |

## Appendix B. Superseded Proposed Slice Plan

All slices below require a later implementation authorization. They preserve observable behavior unless expressly marked otherwise.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| PC-SL-00 | WF-02 | Resolve RB-001 in the normative owner documents and derived contract inputs before changing Network Flow config behavior; `requires later authorization` | Network Flow and Extensions owner documents; authored extension validation inputs; generated outputs only through Make | Selecting either current diagnostic changes one owner's required observable behavior | Preserve and update the claimed/unclaimed key-ring matrix only after owner resolution | `make lint-markdown`<br>`make json-shape-check`<br>`make generate-drift`<br>`make test-slice OWNER=platform.config` | Revert the owner resolution and all generated downstream changes together | Adopted owners agree on one diagnostic, derived contracts match it, and the live test proves it. |
| PC-SL-01 | WF-03; PC-SL-00 only for the Network Flow selector's final expectation | Add canonical authored test-family/verification ownership for the five omitted tests and regenerate downstream accounting through Make | `tools/test_families/platform.config.json`, verification owner inputs, generated harness artifacts downstream | Incorrect ownership can hide tests or double-count evidence | All sixteen live target test functions remain executable; exact five omitted selectors gain owners | `make harness-contract`<br>`make generate-drift`<br>`make test-slice OWNER=platform.config` | Revert authored catalog rows and generated outputs together; never patch generated files | `make explain-test-owner OWNER=platform.config` resolves every live target selector exactly as intended. |
| PC-SL-02 | PC-SL-01 | Add or realign characterization for diagnostic aggregation/order, load precedence, roots, claims, inactive keys, telemetry, limits, and actual filesystem consumers without changing production behavior | Target tests, server/app tests, and owner consumer tests | Tests may reveal current nonconformance or unsupported assumptions | Preserve all target tests; add multi-owner diagnostics and real traversal/symlink/write consumer cases | `make test-slice OWNER=platform.config`<br>`make otel-conformance`<br>affected-owner slices discovered with `make explain-test-owner` | Revert characterization and authored accounting changes together | RB-002 and RB-003 are closed; every later move has executable behavior evidence. |
| PC-SL-03 | PC-SL-02 | Introduce a compatibility-preserving generic config facade and strict owner-contribution interface; retain existing exported entry points as delegating compatibility surface | `internal/platform/config`, `internal/app/extensionassembly`, server/migrate/operator assembly | Unknown-key handling, diagnostic ordering, validation phases, and startup timing | Existing config/server tests plus contribution ordering, duplicate namespace, and unknown-key cases | `make test-slice OWNER=platform.config`<br>`make backend-module-boundary-check`<br>`make test-fast` | Keep the current direct implementation until all contribution tests pass; revert interface and wiring as one slice | Generic loading remains authoritative, owner namespaces are strict, and public callers observe identical output. |
| PC-SL-04 | PC-SL-03; Network Flow portion also depends on PC-SL-00 | Move telemetry, extension inactive policy, Network Flow key-ring rules, and enterprise-auth policy behind their owners while the facade coordinates parsing and diagnostics | Config, telemetry, Extensions/application assembly, Network Flow, enterprise authentication | Config keys/defaults, exact diagnostics, secret redaction, claim behavior, and startup order | Owner-specific config suites, all target tests, server readiness tests | `make test-slice OWNER=platform.config`<br>`make otel-conformance`<br>`make backend-module-boundary-check`<br>`make test-fast` | Move one owner policy per commit and retain a delegating adapter until its focused gates pass | Generic config no longer embeds owner policy; all frozen wire and startup behavior remains unchanged. |
| PC-SL-05 | PC-SL-03, PC-SL-04 | Replace domain and transport use of full config or platform DTOs with owner-local immutable options mapped at application composition | `httpapi.DependencySet`, imports, incident bundles, reference data, recovery, Network Flow, server assembly | Limits, storage semantics, route outcomes, and constructor defaults | Mapping tests for each option plus existing owner unit/integration tests | `make backend-module-boundary-check`<br>`make test-fast`<br>`TODO: exact affected-owner service-backed rows after make explain-test-owner` | Migrate one consumer at a time; keep the old field until that consumer passes, then remove it in the final sub-slice | No domain module imports `internal/platform/config` and transport does not expose full `Config`. |
| PC-SL-06 | PC-SL-05 | Move public sort/filter/change/action ceilings to their contract owners and narrow Postgres, object-store, telemetry, and recovery adapter inputs | View-query/public mutation owners, platform adapters, server assembly, config compatibility aliases during migration | Numeric limit changes, query rejection, storage selection, and telemetry startup | Preserve limit tests and add value-mapping/constructor tests for each adapter | `make backend-module-boundary-check`<br>`make test-fast`<br>`TODO: exact public-query/mutation owner rows after catalog discovery` | Preserve compatibility constants until all callers migrate; revert one owner/adapter independently | Values are unchanged, all callers use semantic owner types, and compatibility aliases can be removed safely. |
| PC-SL-07 | PC-SL-02 through PC-SL-06 | Remove obsolete compatibility aliases, dead/test-only filesystem helpers where RB-003 proves them unnecessary, and `.gitkeep` | Config package and moved consumer tests | Accidental removal of an enforcement path or external package surface | Full target and affected-owner characterization | `make test-slice OWNER=platform.config`<br>`make backend-module-boundary-check`<br>`make test-fast` | Revert cleanup only; structural slices remain independently valid | No live caller or contract uses removed surfaces and real filesystem enforcement remains covered. |
| PC-SL-08 | PC-SL-01 through PC-SL-07 as selected | Run focused, drift, boundary, broad, and finalization gates; update this handoff with exact results | Tracker and later authorized diff | False completion claims, generated drift, or stale harness evidence | All affected owner tests and contract checks | `make generated-artifact-policy-check`<br>`make json-shape-check`<br>`make generate-drift`<br>`make agent-finalize`<br>`make check` | Revert the failing implementation slice; do not mark complete on a failed or skipped required gate | Required gates pass, result roots and skips are recorded, and only intentional files remain changed. |

## Appendix C. Planning-Session Validation

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=platform.config` | Canonical focused config-owner rows | yes | Discovered through `make task-guide` and `make explain-test-owner`; currently omits five live tests under RB-002. |
| integration | `TODO: exact affected-owner slice after make explain-test-owner` | Server startup and consumer behavior affected by a selected slice | no | No current service-backed `platform.config` rows were found. Do not invent row IDs; select affected owners before each implementation slice. |
| e2e/browser | Not applicable to the current target | No direct frontend or browser surface | no | Discover an exact owner row only if a later slice changes public-origin, extension visibility, or browser WebSocket behavior. |
| generated drift | `make generate-drift` | Generated contracts and harness outputs downstream of authored owners | no | Required whenever authored contract, generator, or harness inputs change; never hand-edit generated outputs. |
| generated policy | `make generated-artifact-policy-check` | Generated-root edit policy | no | Required before handoff for any later task touching owner inputs or generation. |
| contract shape | `make json-shape-check` | JSON contracts, manifests, and owner inputs | no | Required for later contract/harness changes. |
| import-boundary/static | `make backend-module-boundary-check` | Backend package-direction and module-boundary rules | yes | Run before and after facade/consumer migration. |
| telemetry static | `make otel-conformance` | OpenTelemetry contract/evidence consistency | yes | This static gate cites a config test but does not replace executing the test through an owner slice. |
| focused broad | `make test-fast` | Fast repository verification after focused owner checks | no | Broaden after the narrow target passes. |
| full check | `make check` | Repository-wide developer verification | no | Run after `make agent-finalize` when later implementation risk warrants it. If `RESULTS_DIR` is unset, report retained-run maintenance as skipped. |
| tracker documentation | `make lint-markdown` | This authored Markdown tracker | yes | Required for this planning-only session. |

Command discovery used the live Make task surface. Repository commands were not invented, and no direct `go`, pnpm, Vitest, Playwright, or raw test-script invocation is proposed.

For this planning-only session, only `make lint-markdown` is an execution validation. The `make task-guide`, `make explain-test-owner`, `make explain-target`, `make help-all`, repository searches, and Git state checks used to build the tracker were read-only discovery commands, not validation runs.

Tracker-only validation result:

| Command | Result | Notes |
| --- | --- | --- |
| `make lint-markdown` | PASS | Completed with exit code 0 and no emitted failure output. |

## Appendix D. Superseded Planning Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| PC-PLAN-001 | Establish scope, authority, snapshot, and tracker | WF-00 | DONE | none | Section 1, this tracker, and tracker-only validation | Markdown validation passed and only this tracker is changed. |
| PC-PLAN-002 | Inventory all nine target files and their live surface | WF-01 | DONE | PC-PLAN-001 | Section 2 | Every target file has a concrete responsibility, caller, dependency, test, owner, and risk posture. |
| PC-PLAN-003 | Map current contracts and owner authority | WF-02 | DONE | PC-PLAN-002 | Sections 3 and 4 | Every discovered contract is mapped; contradictions are blockers rather than decisions. |
| PC-PLAN-004 | Analyze characterization and harness gaps | WF-03 | DONE | PC-PLAN-003 | RB-002, RB-003, Sections 4 and 8 | Missing evidence/accounting is explicit and later closure criteria are defined. |
| PC-PLAN-005 | Classify boundary and coupling findings | WF-04 | DONE | PC-PLAN-002 | Section 5 | Every finding has an allowed classification, proposed owner, and planning action. |
| PC-PLAN-006 | Define facade and ownership redesign direction | WF-05 | DONE | PC-PLAN-003, PC-PLAN-004, PC-PLAN-005 | Sections 3, 5, and 7 | Safe architecture direction is documented without authorizing a permanent boundary change. |
| PC-PLAN-007 | Define implementation and verification slices | WF-06, WF-08 | DONE | PC-PLAN-006 | Sections 6 through 8 | Every slice has dependencies, risk, tests, validation, rollback, and completion criteria. |
| PC-BLOCK-001 | Reconcile the Network Flow/Extensions diagnostic contradiction | WF-02 | BLOCKED | PC-PLAN-003 | RB-001 | Adopted owners and derived contracts agree before affected code changes. |
| PC-TEST-001 | Add canonical ownership for five omitted tests | WF-03, WF-07 | TODO | PC-BLOCK-001 for the Network Flow expectation only | RB-002; `tools/test_families/platform.config.json` | All intended live selectors resolve under canonical owner inspection and execute in the owner slice. |
| PC-TEST-002 | Prove root containment at actual filesystem consumers | WF-03 | TODO | PC-PLAN-004 | RB-003 | Traversal, symlink, writability, and escape behavior are characterized at production call paths. |
| PC-REF-001 | Implement the behavior-preserving facade/owner slices | WF-05, WF-06 | DEFERRED | PC-BLOCK-001, PC-TEST-001, PC-TEST-002 as applicable | Section 7 | A later authorized task selects and completes slices with their gates. |
| PC-HAND-001 | Validate implementation and finalize handoff | WF-08 | DEFERRED | Selected implementation slices | Section 8 and future result roots | Required gates pass and the session logs contain exact results and remaining blockers. |

## Appendix E. Planning Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-24T22:00:41-04:00 | Codex planning session | Authority, one-file write scope, repository snapshot, initial tracker, and documentation validation complete | Framework, owner documents, `docs/domain.md`, target files, and this tracker; only this tracker touched | `sed`; `rg`; `find`; `wc`; `git status --short`; `git branch --show-current`; `git rev-parse HEAD`; `date -Iseconds`; `make lint-markdown` | Tracker created and Markdown validation passed | RB-001 through RB-003 | A later authorized task begins with owner resolution and characterization/accounting. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-24T22:00:41-04:00 | Codex planning session | Mixed-responsibility platform facade diagnosed; no implementation performed | All target files; app, platform, and selected module consumers inspected; tracker touched | Import/symbol searches; `make task-guide ROLE=module-author OWNER=platform.config`; `make explain-target TARGET=backend-module-boundary-check DETAIL=summary` | Broad config propagation and owner-policy placement classified | RB-001 and RB-003 | A later task starts with PC-SL-02/PC-SL-03 only after applicable blockers close. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-24T22:00:41-04:00 | Codex planning session | `intentional/no_action`; target has no direct frontend, controller, selector, view-schema, or grid-adapter surface | Target and caller/import evidence inspected; tracker touched | Repository searches for target imports and relevant contract terms | Frontend work excluded unless a later slice changes origin or extension visibility behavior | None specific to frontend | Preserve current browser-visible behavior; discover an exact owner row only if later scope reaches it. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-24T22:00:41-04:00 | Codex planning session | Configuration, extension, Network Flow, telemetry, and generated-descriptor risks mapped; generated files read only | Owner specs; extension and OTel contract inputs; generated Go/TS consumers; tracker touched | `sed`; `rg`; `make explain-target TARGET=generate-drift DETAIL=summary`; `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary` | No generated edit proposed; exact owner contradiction recorded | RB-001 | Reconcile owners first, then update authored inputs and regenerate only through Make in a later task. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-24T22:00:41-04:00 | Codex planning session | Sixteen live target test functions compared with canonical owner accounting | Target tests, `tools/test_families/platform.config.json`, verification owner manifests, and tracker inspected; tracker touched | `rg -n '^func Test'`; `make task-guide ROLE=module-author OWNER=platform.config`; `make explain-test-owner OWNER=platform.config`; `make help-all`; `make explain-target TARGET=test-slice DETAIL=summary`; `make explain-target TARGET=otel-conformance DETAIL=summary` | Eleven selectors are accounted for; five live selectors are omitted; no service-backed `platform.config` rows found | RB-002 and RB-003 | Later PC-SL-01 adds authored ownership; PC-SL-02 adds real consumer characterization. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-24T22:00:41-04:00 | Codex planning session | No authorization decisions live in the target; secret and root-security boundaries are configuration concerns | `validation.go`, telemetry bootstrap/contract evidence, Network Flow key-ring consumer, owner specs, and tracker inspected; tracker touched | `sed`; `rg` | Secret resolution/redaction and fail-before-readiness behavior frozen; root enforcement evidence gap recorded | RB-001 and RB-003 | Preserve exact security outcomes; obtain owner resolution and consumer-level root tests before movement. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-24T22:00:41-04:00 | Codex planning session | Tracker complete; implementation deliberately deferred | Tracker touched; all evidence groups in Section 1 inspected | Discovery commands listed above; `make lint-markdown`; no test or refactor command run | Documentation validation passed; three stable blockers/gaps remain; no production refactor performed | RB-001 through RB-003 | A later authorized session begins with owner resolution and characterization/accounting. |
| 2026-07-24T22:36:41-04:00 | Codex closure-guidance session | Canonical closure decisions documented without claiming repository closure | `temp/analysis-notes.md` inspected; only this tracker touched | `git status --short`; `date -Iseconds`; `sed`; `make lint-markdown` | Closure guidance added and Markdown validation passed; required owner edits, implementation, and retained evidence do not yet exist | RB-001 through RB-003 retain their existing statuses | The next authorized task begins with RB-001 owner reconciliation, RB-002 authored rows, and RB-003 production-effect inventory. |

## Appendix F. Planning Closure Guidance

`temp/analysis-notes.md` supplies closure guidance and rationale for this section. It is not a normative owner document or proof that repository changes or executable evidence exist. Adopted owner documents, authored machine contracts, live implementation, canonical test execution, and retained evidence remain controlling.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | The adopted correction is that unclaimed `network_flow_activity.key_ring_manifest_path` uses `extension_config_without_claim`. | The reason code, message/details, validation precedence, and absence of file access are observable startup behavior. | Network Flow/Core 04/Core 00 reconciliation, aligned contracts, drift-clean generation, and passing claimed/unclaimed matrix | CLOSED |
| RB-002 | Five independent active `platform.config` rows own the omitted top-level test symbols; no aggregate OTel row or duplicate subsystem ownership exists. | Focused owner execution must retain every material runtime-configuration selector. | Authored family and verification-owner rows, exact selector resolution, successful full owner slice, and passing OTel conformance | CLOSED |
| RB-003 | `platform.config` owns startup root admission and neutral rooted capabilities enforce race-safe containment at every production read/write/extract effect. | Helper-only tests cannot prove operation-time security at real owner entry points. | Complete inventory, adopted requirements, rooted capabilities, actual owner matrices, exact static boundary, and obsolete-helper deletion | CLOSED |

### RB-001 closure guidance: inactive Network Flow configuration

For this exact condition:

```text
network_flow_activity.claimed = false or omitted
network_flow_activity.key_ring_manifest_path is present
```

the coordinated owner amendments must establish:

```text
top-level error.code = invalid_deployment_config
reason_code = extension_config_without_claim
message = "Extension configuration is present while the profile is inactive."
details.profile_id = "network_flow_activity"
details.config_path = "$.network_flow_activity.key_ring_manifest_path"
```

Inactive-extension structural admission must fail before active-value grammar, defaulting, effective-value retention, file existence/readability/type/size checks, secret or reference resolution, Network Flow invocation, or egress. `profile_incompatible_binding` remains available for actual deployment-profile or root-binding incompatibility and must not alias this inactive extension-key result.

Required owner work:

1. Amend the Network Flow paragraph following the §20.1 configuration table and `NF-REQ-183` to use the Extensions-owned generic inactive-key result.
2. Expand `NF-AC-111` to verify the exact code, message, profile ID, extension config path, validation precedence, no file access, and no compatibility alias.
3. Amend Core 04 `REQ-04-143` and its acceptance coverage to bind forbidden inactive extension keys to `extension_config_without_claim`, import the Extensions diagnostic registry into the deployment-config envelope, and define the exact finding-to-`items[]` translation and ordering.
4. Copy the current live Core 04 `items[].path` spelling into the owner amendment; do not infer it from the Extensions JSON path.
5. Update Core 00 adoption/version records and the authored Network Flow/extension validation-condition inputs, verification ownership, dependencies, digests, and conformance status. Regenerate all derived outputs through Make.
6. Keep Extensions generic semantics unchanged unless a clarification is needed; no profile-local replacement is allowed without the complete replacement mapping required by the Extensions owner.

Version classification remains a mandatory release-history gate:

- Use Network Flow document version `2.0.1`, retaining contract major 2, only if release/conformance inventory proves no conforming released artifact or accepted claim emitted `profile_incompatible_binding` for this condition.
- If that proof fails, use contract major 3. Do not provide a dual-code compatibility alias.

Executable closure matrix:

| Claim state | Path state | Required result |
| --- | --- | --- |
| Omitted or `false` | Absent | Valid; no Network Flow file work. |
| Omitted or `false` | Present with any value spelling or file state | `extension_config_without_claim`; no active-value validation or file access. |
| `true` | Absent | Claimed-profile missing-required-key failure. |
| `true` | Present but syntactically invalid | Active Network Flow path-validation failure. |
| `true` | Valid path with unreadable, non-regular, or malformed manifest | Active Network Flow manifest failure before readiness. |
| Any unclaimed state | Path present | Never `profile_incompatible_binding`. |

RB-001 closes only after coordinated owner adoption, recorded version classification, aligned authored contracts, drift-clean generation, and canonical owner-slice execution of this matrix.

### RB-002 closure guidance: five omitted executable rows

Add five active rows to the authored `platform.config` family. Each row must select exactly one symbol:

| Semantic row suffix; reuse an existing equivalent ID if present | Exact Go test symbol | Primary postcondition | Collaborators |
| --- | --- | --- | --- |
| `extension_claims.network_flow_activity` | `TestNetworkFlowActivityConfigDefaultsAndClaimability` | Network Flow claim default and inactive-key admission | Extensions and Network Flow owners |
| `telemetry.defaults_closed_namespace` | `TestOpenTelemetryConfigDefaultsAndClosedNamespace` | Defaults and closed `telemetry.*` namespace | OpenTelemetry / `platform.telemetry` |
| `telemetry.environment_binding_parser` | `TestOpenTelemetryEnvironmentBindingParser` | Cartulary environment-binding parser | OpenTelemetry / `platform.telemetry` |
| `telemetry.validation` | `TestOpenTelemetryConfigValidation` | Cross-key, endpoint, and configuration validation | OpenTelemetry / `platform.telemetry` |
| `telemetry.secret_references` | `TestOpenTelemetrySecretReferences` | Secret references, headers, redaction, and purpose isolation | OpenTelemetry / `platform.telemetry` |

Each row must use:

```text
owner_id = platform.config
runner = go
selector.symbols = [one exact Test... symbol]
evidence_class = unit
claim_posture = implementation
status = active
default_check = true
```

Copy package spelling and runtime, resource, and fixture profiles from adjacent no-managed-service `platform.config` Go unit rows. Do not add a second package alias, new runner, schema, public Make target, or service-backed profile without contrary live evidence. `platform.config` remains primary owner because the tests exercise deployment parsing, defaulting, closed-key admission, diagnostics, and pre-readiness behavior. Subsystem owners are collaborators and must not own duplicate executable rows.

Verification mapping:

| Test | Required verification references |
| --- | --- |
| Network Flow claimability | Core 04 claim/config envelope, Extensions generic claim/inactive-policy rules, and the corrected Network Flow requirements |
| OTel defaults/closed namespace | `OTEL-REQ-023..025` |
| OTel environment parser | `OTEL-REQ-122..124` |
| OTel validation | `OTEL-REQ-029..030` plus the endpoint/config rules actually exercised |
| OTel secret references | `OTEL-REQ-144..146`, Core 04 `secret_ref_v1`, and at least one security-class verification ID |

At minimum, later authorized work updates `tools/test_families/platform.config.json` and `contracts/verification/owners/platform.config.json`. Any OTel conformance-status owner input that treats a raw test filename as executable coverage must instead reference stable verification or row identity. `contracts/verification/owners/platform.telemetry.json` may reference the rows for traceability but must not duplicate ownership. Generated accounting and topology remain generator-owned.

Required later validation sequence:

```sh
make json-shape-check
make harness-contract
make generate
make generate-drift
make generated-artifact-policy-check
make explain-test-owner OWNER=platform.config
make test-slice OWNER=platform.config
make otel-conformance
```

RB-002 closes only when:

1. Active `platform.config` Go selectors cover all sixteen intended live target tests.
2. Each new symbol is selected exactly once and overlaps no other active row.
3. Owner explanation lists all five obligations and an omitted `ROWS` value produces a `full_owner` plan containing them.
4. Retained owner accounting has one successful terminal row per new row.
5. `make otel-conformance` passes as supplemental static evidence, not as a substitute for owner-slice execution.
6. No service-backed row is introduced without evidence that a test needs managed services.

### RB-003 closure guidance: production filesystem containment

The permanent responsibility split is:

- `platform.config` validates and canonicalizes configured root bindings before readiness.
- A neutral platform capability enforces containment at every actual read, write, create, rename, remove, or extract effect.
- Application composition converts validated roots into owner-specific capabilities; domain modules do not receive raw root paths or the complete `config.Config`.
- Each archive/write owner retains its public or job error mapping, cleanup, publication, transaction, and audit consequences.
- Read-only absolute manifest references for bootstrap, enterprise authentication, and Network Flow remain a separate concern and must not be forced through an archive/write abstraction.

The concrete package and method names remain implementation latitude until the call-path inventory is complete. A package such as `internal/platform/filesystemroot` and an interface conceptually named `RootedFilesystem` are acceptable only if the boundary:

- constructs capabilities solely from validated `filesystem_root` bindings;
- accepts validated relative POSIX child paths and rejects absolute paths, empty or dot components, `..`, NUL, backslashes, and normalization collisions;
- exposes no arbitrary absolute-path or raw-root escape hatch to domain modules;
- anchors effects to the startup-canonical root without following child or final-target symlinks;
- makes containment and the filesystem effect race-safe rather than checking a prefix and reopening an unconstrained path;
- verifies the actual opened object type, uses exclusive creation or an owner-defined atomic replacement protocol, and keeps temporary/final renames inside one root;
- fails safely if the root disappears, is replaced, or loses writability after startup, without partial publication;
- provides owner-controlled cleanup on success, failure, cancellation, and timeout;
- never instantiates a local fallback for `managed_service` bindings; and
- prevents diagnostics or retained evidence from leaking absolute roots, raw hostile names, secrets, object keys, or credentials.

Core 04 must own these common operation-time invariants and distinguish:

| Failure source | Required error family |
| --- | --- |
| Invalid configured root at startup | `invalid_deployment_config` with existing root/config reason codes |
| Hostile member or child path supplied to an operation | The owning route/job error, such as `invalid_member_path` or `path_traversal` |
| Root removed, replaced, or unwritable after readiness | Safe owner storage/dependency failure with no partial success or publication |

Named owners retain their narrower archive formats, cleanup, publication, and error semantics. Supporting material may define Go interfaces and OS strategy, but race, containment, type, cleanup, and failure consequences must live in Core 04 or the adopted named owner.

Minimum production-effect inventory:

| Owner/call path | Required finding |
| --- | --- |
| `internal/modules/imports/routes.go`, `internal/modules/imports/xlsx.go` | Prove whether bytes remain in memory or use temporary storage; if written, prove temporary-work containment and cleanup. |
| `internal/modules/incidentbundles/bundle.go` | Prove container/member containment, unsupported-member rejection, staged handling, cleanup, and no incident publication on failure. |
| `internal/modules/reference_data/verifier.go` | Prove private temporary extraction, reference-pack publication, integrity checking, and cleanup. |
| Backup/recovery and operator paths | Prove backup-root containment, staging/publication safety, target separation, and no raw-path disclosure. |
| Report, preview, and export-output builders | Find every live `temporary_work` or `export_outputs` writer and characterize each actual effect. |
| Postgres/object-store filesystem adapters | Separate binding/startup/open behavior from hostile member-path handling. |
| Bootstrap, enterprise-authentication, and Network Flow manifests | Confirm read-only absolute regular-file handling outside the rooted write/extract capability. |
| External libraries receiving output paths | Treat the library call as a filesystem effect; prove the boundary is not bypassed. |

The inventory must search direct and indirect uses of create/open-for-write, temporary directory/file creation, rename, remove, extraction callbacks, libraries receiving output paths, and write-to-path helpers. A no-effect classification is valid only when the inspected production path proves data remains in memory or in a managed service.

Every actual write/extract owner needs entry-point tests for applicable cases:

- parent traversal, absolute paths, backslashes, NUL, empty/dot components, and normalization collisions;
- existing child/final symlinks, symlink loops, and a directory replaced by a symlink between admission and effect;
- archive symlinks, hard links, devices, FIFOs, sockets, duplicate normalized paths, and file/directory prefix collisions;
- existing destinations, in-root success, root loss or writability loss after startup, and managed-service no-fallback behavior;
- failure, cancellation, timeout, cleanup, atomic publication, and diagnostic non-disclosure.

Tests must enter through the real owner facade, job, import, or extraction path and use a real temporary filesystem. Helper-only normalization tests remain support evidence and cannot close the obligation. Writability cases must assert the real creation/open result rather than only permission bits.

Evidence ownership remains distributed:

| Evidence | Primary owner |
| --- | --- |
| Root syntax, profile compatibility, overlap, and startup writability | `platform.config` |
| Incident-bundle extraction/publication | Incident Portability |
| Reference-pack extraction/activation eligibility | Reference Pack |
| XLSX/import temporary handling | Imports |
| Report/preview/export writes | Reporting owner |
| Backup/restore writes | Recovery |
| Common rooted-filesystem mechanism and package guard | Platform filesystem/security support |

Add an exact production-import/static-analysis guard so only approved rooted-filesystem adapters perform generic archive extraction or arbitrary filesystem mutation for these flows. Owner packages consume narrow ports, test-only imports grant no production permission, and the allowlist uses exact package paths.

Keep the current `validation.go` helpers until production characterization is complete. Then move and retain them only if their semantics belong to the rooted-filesystem boundary, or delete them if call-graph inspection proves zero production use and real consumers cover every relevant behavior. Do not retain compatibility aliases solely for helper tests, expose a generic raw-path utility, or move the helpers into a domain owner.

RB-003 closes only after the production-effect inventory is complete, Core/owner requirements are adopted, every actual owner has executable containment evidence, the static boundary passes, and moving or deleting the existing helpers leaves no uncovered production effect.

### Final closure gates

The following are later-task gates, not validation results from this documentation-only update:

```text
RB-001:
  coordinated owner amendments adopted
  release-history/version classification recorded
  authored contract inputs aligned
  generated outputs drift-clean
  exact diagnostic matrix passes

RB-002:
  five exact-selector active rows present
  verification references resolve once
  full platform.config owner slice passes
  retained accounting contains all five terminal rows
  static OTel conformance also passes

RB-003:
  production filesystem-effect inventory complete
  operation-time rooted-filesystem boundary implemented
  every actual owner has real consumer tests
  symlink/race/writability/escape matrix passes
  exact static package boundary passes
  old helpers moved or deleted without coverage loss
```

No other question currently blocks safe planning. The exact final location of `extensioninactive` remains deferred inside the slice plan. Rooted-filesystem package and method names likewise remain deferred pending the production-effect inventory; the required security semantics and ownership split are fixed.

## Appendix G. Planning Completion Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/platform/config` is inventoried or explicitly out of scope. | PASS | Section 2 contains all nine files, including `.gitkeep` and `extensioninactive/catalog.go`. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 records owner, evidence, existing tests, characterization gap, risk, and notes; absent direct surfaces are explicitly identified. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 records required previous/subsequent workflows and binary handoff checkpoints. |
| Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | PASS | Section 7 freezes behavior; PC-SL-00 is explicitly owner/behavior-authorized only, and all implementation is deferred to a later task. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 uses live Make targets and records why integration/browser owner rows require later discovery. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | RB-001 uses the required wording and prevents a diagnostic choice. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Sections 1 and 3 diagnose a legitimate facade combined with mixed owner responsibilities. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Sections 9 through 11 record status, evidence, commands, blockers, and the next safe actions. |
| Only the tracker file is changed and tracker Markdown validation passes. | PASS | `make lint-markdown` passed; final diff and Git status inspection found only this tracker. |

The tracker is structurally complete. Its completion does not resolve RB-001 through RB-003 and does not authorize any production refactor.
