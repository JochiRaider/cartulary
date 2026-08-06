# protocol-ts Module Refactoring Tracker and Handoff

## Current Controlling Posture

Iteration 1 is complete at commit
`d29322b6dc403fb718a6635308c5756becfd97d8`. Sections 1 through 12 are the
closed Iteration 1 execution record and retain their original scope, evidence,
and decisions. Sections 13 through 19 are the controlling plan for Iteration 2.

Iteration 2 is **ITERATION 2 COMPLETE** on branch `main` from the clean
`d29322b6dc403fb718a6635308c5756becfd97d8` implementation baseline. Only this
tracker differed from that baseline at activation. The tracker amendment and
I2-00 through I2-07 completed in the required sequence, every checkpoint
passed, and Section 19 is the final restart and handoff record.

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Target path | `packages/protocol-ts` |
| Target label | `protocol-ts`; derived from the target path and normalized to lowercase kebab case. |
| Output path | `docs/handoffs/protocol-ts-packages-refactor-tracker.md` |
| Session baseline | Branch `main`; commit `8ccf5de48a80ec370764c5840a9e0fc2f22e5912`; target and output-path existence rechecked on 2026-08-03. |
| Status | COMPLETE — S-00 through S-04, S-06, and S-07 are complete; S-05 is explicitly deferred. This file is the final controlling execution and handoff artifact. |
| Allowed change | The specification, authored contracts and generator inputs, generated projections through Make-owned generation, implementation, tests, harness owner inputs/projections, and documentation named by S-00 through S-07. |
| Non-goals | Adding or removing public package entrypoints; exporting private facade paths; adding audit/revisions entrypoints; changing unrelated product behavior; hand-editing generated files or lockfiles; changing `docs/domain.md`. |
| Implementation boundary | Execute only the authorized slices in this order: tracker activation, S-00, S-06 red gate, S-01, S-06 green gate, S-02, S-03, S-04, S-05 deferred decision, and S-07. Each slice must remain independently reviewable and revertible. |

The authority order used for this tracker is:

1. Adopted subsystem NLSpecs for their named scopes.
2. Core 00 through Core 04 for current implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication. The
   authorized remediation does not create such a claim unless S-05 is reopened.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests for implementation state.
6. Prior plans, handoffs, and the planning framework as evidence only.

### Normative language and default behavior

Within this tracker, `MUST` and `MUST NOT` state mandatory conditions for a
conforming refactor, `SHOULD` states a requirement that may be waived only by a
retained owner-approved exception, and `MAY` states an optional action. Current
repository observations are descriptive rather than normative.

The refactor preserves every public symbol and the five supported package
specifiers. Behavior and wire shape remain unchanged unless an authorized slice
explicitly repairs an owner/projection defect. Existing behavior has no
independent preservation priority when it conflicts with an adopted owner or
the approved forward-compatible design. Silence, repository-local non-use,
generated presence, or a historical label does not authorize public drift.

### Execution checkpoint protocol

After completing each workstream, the implementer **MUST** update this tracker
with changed files, commands, retained run roots, results, blockers, and the
next slice, then run `make lint-markdown`. The next workstream **MUST NOT** begin
until that checkpoint passes. Expected failure at the S-06 red gate is retained
evidence rather than slice completion; S-06 completes only at its post-S-01
green gate. Generated roots and Make-owned generated harness projections are
never hand-edited. Tests and production evidence **MUST NOT** read, stat, hash,
or parse this tracker or any other Markdown file.

This tracker is implementation-support material. Adopted owner documents remain
authoritative for product behavior. S-04 clarifies Core 01's existing
string-valued upload-header requirement and repairs downstream projections; no
other Core amendment is required. If current owners contradict one another,
the affected work **MUST** be recorded as `BLOCKED: owner contradiction` and
**MUST NOT** proceed until the owners are repaired. No Core amendment is
required for the behavior-preserving decisions recorded here; a Core amendment
is required only when the applicable Core behavior itself changes or is found
ambiguous.

Owner documents inspected for the relevant boundaries were:

- `docs/spec/00_document_set_status_and_precedence.md`, especially authority,
  profile, contract-owner, and generated-projection posture;
- `docs/spec/01_architecture_storage_and_view_contracts.md`, especially public
  HTTP/WebSocket, workbook, projection, evidence, import, and extension routes;
- `docs/spec/02_domain_model_schema_and_history.md`, especially record,
  entity, evidence, saved-view, relationship, and revision semantics;
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`,
  especially workbook, collaboration, mutation, saved-view, evidence, import,
  and projection-facing interaction behavior;
- `docs/spec/04_security_deployment_and_conformance.md`, especially session,
  authorization, trust-boundary, and public-interface criteria;
- `docs/extension-subsystem-nlspec.md`, especially generic discovery,
  compatible-client decoding, extension workspace, and generated-registry
  boundaries;
- `docs/network-flow-activity-nlspec.md`, especially public schema, import
  target, graph adapter, route, authorization, and frontend contract scopes;
- `docs/graph_projection_nlspec.md`, only for the graph-projection boundary
  imported by Network Flow;
- `docs/testing-harness-nlspec.md`, especially Make ownership, generated-file
  policy, owner/row routing, and evidence limitations;
- `docs/domain.md`, especially the implementation-detail distinction, bounded
  contexts, published-language relationships, and source/projection boundary;
- `docs/guides/cartulary-dev-guide.md`, especially the live-intended frontend
  package boundaries, contract derivation, generated-artifact policy, and
  canonical validation commands; and
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, used as
  planning doctrine rather than repository-state evidence.

Supporting material inspected for this revision was:

- `docs/research/nlspec-spec.md`, used for specification economy, interface
  precision, explicit defaults, mapping completeness, and binary acceptance
  posture;
- `temp/analysis-notes.md`, used as the approved intent that closes RB-001
  through RB-003;
- local research R01, R03, R08, and R09 where they support facade decomposition
  and curated package boundaries; and
- official [TypeScript optional-property](https://www.typescriptlang.org/tsconfig/exactOptionalPropertyTypes.html),
  [Node.js package-export](https://nodejs.org/api/packages.html), and
  [Rollup](https://rollupjs.org/configuration-options/) documentation as
  non-normative implementation rationale.

| Requirement subject | Binding owner or record | Required placement |
| --- | --- | --- |
| HTTP/JSON member requiredness, nullability, omission, scalar type, object openness, and unknown-member behavior | Core 01 and authored OpenAPI/schema inputs | Owner document and machine projection; never this tracker alone. |
| Evidence, history, or revision domain meaning | Core 02 | Core 02 only when that meaning changes. |
| Workbook-visible behavior | Core 03 | Core 03 only when interaction behavior changes. |
| Authorization, redaction, secret handling, or trust boundaries | Core 04 | Core 04 only when security behavior changes. |
| Contract-family ownership | Core 00 section 5.1 | Core 00 only when a new public contract family is created. |
| TypeScript aliases, package exports, private facade layout, and compile fixtures | Package implementation plus this tracker and the S-00 decision record | Supporting implementation records; they **MUST NOT** redefine wire behavior. |
| Current public entrypoint list | `contracts/protocol-ts/frontend-entrypoints.v1.json` and `package.json` | Both authored surfaces **MUST** be revised before any future entrypoint change. |
| Type-compatibility decisions | `docs/decisions/protocol-ts-public-type-compatibility.md` plus compile/runtime evidence | Created by S-00; tests **MUST NOT** read the Markdown record. |
| Bundle graph, chunk sizes, cycles, and materiality result | Retained production-build evidence | Supporting evidence only; S-05 is optional. |
| Browser artifact allowlist and import restrictions | Frontend import-boundary configuration and executable tests | Authored policy plus Make-owned evidence. |
| Generated-file mechanics | Testing Harness, generator inputs, and generated-artifact policy | Generated outputs remain downstream and **MUST NOT** be hand-edited. |

Repository evidence inspected included every tracked file listed in Section 2,
the generated markers and exported surfaces in those files, and these direct
supporting sources:

- `contracts/index.json`, `contracts/protocol-ts/frontend-entrypoints.v1.json`,
  and `contracts/protocol-ts/http-operations.v1.json`;
- `tools/protocol-ts/generate-protocol-types.mjs`, relevant TypeScript-output
  code in `tools/contractgen/main.go` and `tools/contractgen/import_targets.go`,
  and `tools/generated_artifact_policy.json`;
- `tools/frontend_import_boundaries.json`, `.fallowrc.json`,
  `tools/test_families/package.protocol_ts.json`,
  `contracts/verification/owners/package.protocol_ts.json`, and relevant
  collaborator-family rows;
- the declared workspace/package/TypeScript configuration; and
- direct consumers under `apps/web`, `packages/view-contracts`, and
  `packages/ui-contracts` identified by import search and opened before use as
  evidence.

The target path exists. The tracker did not previously exist, so there is no
prior target-specific session history to preserve. Ignored dependency/cache
content under `packages/protocol-ts/node_modules/**` and the ignored
`packages/protocol-ts/tsconfig.tsbuildinfo` are explicitly outside the tracked
source inventory: they are install/build products, not authored or generated
repository inputs.

## 2. Baseline and Final Repository Inventory

The activation baseline contained exactly 24 tracked files under
`packages/protocol-ts`: four authored or configuration files, one
generated-root sentinel, and 19 generated TypeScript files. The final change
set adds seven authored files, for 31 target files once adopted. Generated rows
below remain projections only; none was hand-edited.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `packages/protocol-ts/package.json` | Declares private workspace package, runtime/dev dependencies, and five entrypoints. | `@cartulary/protocol-ts`, root `.`, and `./collaboration`, `./core-http`, `./http`, `./network-flow`. | Workspace resolution; `apps/web`, `packages/view-contracts`, and `packages/ui-contracts`. | Runtime `ajv`; development `vitest`; generated entrypoint files. | Package owner slice, frontend typecheck, import-boundary and build checks. | Exposes generated files through four declared subpaths. | `package.protocol_ts` package surface. | High | Repository non-use of a declared subpath does not authorize removal. |
| `packages/protocol-ts/tsconfig.json` | TypeScript project configuration for the package source. | Referenced project in root `tsconfig.json`; no runtime export. | Root TypeScript project graph. | `../../tsconfig.base.json`; `src`. | `make frontend-typecheck`. | Includes generated source in typechecking. | `package.protocol_ts` build support. | Medium | `noEmit`; does not own protocol behavior. |
| `packages/protocol-ts/src/index.ts` | Composition-only authored public facade. | Six private-facade re-exports preserving every root symbol. | Direct root imports in frontend protocol adapters, collaboration, workbook adapters, package consumers, and tests. | Six authored files under `src/facade`. | Twelve package Vitest titles plus compile and consumer evidence. | No direct generated import. | `package.protocol_ts` facade composition. | High | Public surface is stable; private facade paths remain unexported. |
| `packages/protocol-ts/src/index.test.ts` | Characterizes the exact root facade, runtime validation, decoders, discovery, and artifact identities. | Twelve Vitest titles routed through five active package Vitest rows. | Harness owner catalog. | Root facade and Vitest. | Is the package-owned runtime test file. | Exercises generated HTTP bindings, artifact registries, types, and validators. | `package.protocol_ts` verification. | Medium | Includes all 14 WebSocket families, seven discovery members, additive row/cell tolerance, header typing, exact exports, and payload-free failures. |
| `packages/protocol-ts/src/generated/.gitkeep` | Generated-root sentinel accepted by policy. | None. | Generated-artifact policy only. | None. | Generated-artifact policy checks. | Ignored sentinel named by policy. | Generated-root maintenance. | Low | Not a behavior or generation owner. |
| `packages/protocol-ts/src/generated/artifact.ts` | Shared generated `Artifact` shape and artifact-index helper. | Protected `Artifact` and `indexArtifacts`; transitively used by generated artifact families. | Generated artifact-family modules and generated barrel. | None beyond language runtime. | Package artifact helper tests transitively. | Generated by `tools/contractgen` from active contract-family configuration. | `package.protocol_ts` generated projection support. | Medium | Never hand-edit. |
| `packages/protocol-ts/src/generated/audit-artifacts.ts` | Embeds the administrative-audit registry JSON and digest. | Protected `auditArtifacts` and index; exported by protected generated barrel, not selected by the authored root artifact helpers. | Protected generated barrel; no supported direct package consumer found. | `artifact.ts`; `contracts/audit/index.json`. | Drift/policy/shape checks; no root facade assertion. | Audit contract family. | Administrative-audit owner projected through `package.protocol_ts`. | Medium | Disposition is closed: remain generated/protected/root-omitted/export-omitted/browser-unreachable. Generation does not grant browser exposure. |
| `packages/protocol-ts/src/generated/collaboration-types.ts` | Generated union and interfaces for 14 incident-stream message types. | Root-exported `IncidentStreamMessage`; declared `./collaboration` subpath exposes full generated surface. | Authored root facade and package subpath consumers; current app collaboration imports the root. | WS contract selected by frontend entrypoints. | Collaboration consumers; package root lacks a focused decoder case. | `contracts/ws/index.schema.json`; generated by protocol generator. | Collaboration wire projection; lifecycle remains collaboration/web owned. | High | Message types do not own connection, resume, authorization, or application semantics. |
| `packages/protocol-ts/src/generated/core-http-types.ts` | Generated selected OpenAPI component and operation-reachable request/response types. | Declared `./core-http` subpath; selected types re-exported or aliased by root/HTTP bindings. | Authored root, HTTP bindings, shared `SheetRef`, saved-view browser support, and workbook E2E. | Generated OpenAPI plus selected operation closure. | Package tests and broad consumer unit/browser evidence. | `contracts/openapi/cartulary.openapi.yaml`; generated by protocol generator. | Core/adopted route owners projected through `package.protocol_ts`. | High | Includes workbook, saved-view, evidence, account, admin, import/job, revision, and extension shapes. |
| `packages/protocol-ts/src/generated/errors-artifacts.ts` | Embeds public error and reason-code registry JSON and digest. | Protected values surfaced through root artifact/registry helpers. | Generated barrel and authored root. | `artifact.ts`; `contracts/errors/index.json`. | Package error/reason-code registry test. | Error contract family generated by `tools/contractgen`. | Core error owner projected through `package.protocol_ts`. | High | Error semantics remain owner-defined; this file is static projection data. |
| `packages/protocol-ts/src/generated/extensions-artifacts.ts` | Embeds client-support, descriptor, and profile-registry artifacts. | Protected values surfaced through root extension/artifact helpers. | Generated barrel and authored root. | `artifact.ts`; eight generated extension contract artifacts. | Package extension registry and discovery tests; extension consumers. | Extensions contract family generated by `tools/contractgen`. | Extensions owner projected through `package.protocol_ts`. | High | Does not authorize or activate extensions. |
| `packages/protocol-ts/src/generated/http-operation-bindings.ts` | Generated operation metadata, path/query builders, success response validation, and request/response maps for 66 selected operations. | Root re-export and declared `./http` subpath. | Authored root, browser transport/workbook adapters through root; no current direct `./http` import found. | `core-http-types.ts`, `protocol-validators.ts`, selected OpenAPI operations. | Package HTTP binding test and consumer tests. | `contracts/protocol-ts/http-operations.v1.json` and generated OpenAPI. | `package.protocol_ts` wire/transport projection. | High | Three operations are explicitly forward tolerant; package does not implement routes. |
| `packages/protocol-ts/src/generated/import-target-registry.ts` | Generated frontend projection of import target identity, ownership, and availability rows. | Root re-export of types and `importTargetRegistry`. | Authored root and `apps/web/src/services/importTargetContractAdapter.ts`. | Authored import target registry and contractgen inputs. | Package drift plus module-imports generated-target consumer tests. | Imports target registry generated by `tools/contractgen`. | Imports owner projection; interpretation remains import adapter/module owned. | High | Must not become parser, mapping, apply, or lifecycle authority. |
| `packages/protocol-ts/src/generated/index.ts` | Protected generated barrel for artifact families and import target registry. | No declared package entrypoint and no authored runtime consumer. | Generated output only. | Generated artifact-family files. | Import-boundary, drift, policy, and browser-reachability checks. | Generated by `tools/contractgen`. | `package.protocol_ts` generated aggregation. | High | Unconditionally denied to authored imports; audit/revisions modules and bytes are browser-unreachable. |
| `packages/protocol-ts/src/generated/network-flow-artifacts.ts` | Embeds ten Network Flow contract artifacts and digests. | Protected values surfaced through root artifact/error helpers. | Generated barrel and authored root. | `artifact.ts`; Network Flow indexes, schemas, mappings, presentation, errors, and provenance. | Package Network Flow tests and owner-aligned Network Flow tests. | Network Flow contract family generated by `tools/contractgen`. | Network Flow owner projected through `package.protocol_ts`. | High | Does not own Network Flow resource or UI behavior. |
| `packages/protocol-ts/src/generated/network-flow-descriptor.ts` | Generated profile identity and current contract-major descriptor. | Root-exported `networkFlowContractDescriptor`. | Authored root and Network Flow contract adapter through root. | Network Flow contract index. | Package and Network Flow adapter tests. | Generated by protocol generator. | Network Flow owner projection. | High | Compatibility is contract-major based, not package-name based. |
| `packages/protocol-ts/src/generated/network-flow-mapping-registry.ts` | Generated immutable mapping registry constant. | Root-exported `networkFlowMappingRegistry`. | Authored root and Network Flow adapter. | `contracts/network-flow/mapping-registry.v2.json`. | Network Flow adapter/owner tests; drift checks. | Generated by protocol generator. | Network Flow mapping owner projection. | High | Mapping semantics remain the adopted Network Flow owner's. |
| `packages/protocol-ts/src/generated/network-flow-presentation.ts` | Generated immutable presentation registry constant. | Root-exported `networkFlowPresentationRegistry`. | Authored root and Network Flow adapter. | `contracts/network-flow/presentation.v2.json`. | Network Flow presentation/adapter tests; drift checks. | Generated by protocol generator. | Network Flow presentation owner projection. | High | Does not own React, grid coordinates, or workspace controller state. |
| `packages/protocol-ts/src/generated/network-flow-types.ts` | Generated public Network Flow schema types, including rows, tables, queries, graph results, contributors, indicator links, and import results. | Root type re-export and declared `./network-flow` subpath. | Authored root and Network Flow adapter through root; no current direct subpath import found. | Public Network Flow schema bundle. | Package decoder test and extensive Network Flow consumer tests. | Generated by protocol generator from Network Flow frontend entrypoints and schema bundle. | Network Flow owner projection. | High | Graph result types do not transfer Graph Projection or source-state ownership. |
| `packages/protocol-ts/src/generated/protocol-validators.ts` | Standalone AJV validators for selected Core HTTP, collaboration, and Network Flow schemas. | Protected validator functions; not a package entrypoint. | Authored root decoder factory and HTTP operation bindings. | AJV runtime helpers and generated schema bundles. | Package decoder/HTTP validation tests and consumer validation paths. | Generated by protocol generator. | `package.protocol_ts` runtime validation projection. | High | Failure output must remain payload-safe; never hand-edit generated code. |
| `packages/protocol-ts/src/generated/revisions-artifacts.ts` | Embeds revisions registry and conflict-token key-ring schema artifacts. | Protected `revisionsArtifacts` and index; exported by generated barrel, not selected by root artifact helpers. | Protected generated barrel; no supported direct consumer found. | `artifact.ts`; two revisions contracts. | Drift/policy/shape checks; no root facade assertion. | Revisions contract family generated by `tools/contractgen`. | Revisions owner projected through `package.protocol_ts`. | Medium | Disposition is closed: remain generated/protected/root-omitted/export-omitted/browser-unreachable; no key material may enter frontend artifacts. |
| `packages/protocol-ts/src/generated/view-schema-source-types.ts` | Generated types for authored view-schema source documents, fields, inspector configuration, and route/seed bindings. | Root type re-export. | Authored root and `packages/view-contracts` through the root. | View-schema typed source shape. | View-contracts tests plus package typecheck. | Generated by `tools/contractgen`. | View-schema owners projected through `package.protocol_ts`; semantic adapter is `package.view_contracts`. | High | Raw source types are not workbook interaction semantics. |
| `packages/protocol-ts/src/generated/view-schemas-artifacts.ts` | Embeds 17 view-schema artifacts plus the registry index. | Protected values surfaced through root view-schema helpers. | Generated barrel, authored root, and `packages/view-contracts` through root helpers. | `artifact.ts`; base/optional view-schema contracts and index. | Package registry tests, view-contracts tests, UI selector tests, browser consumers. | View-schema contract family generated by `tools/contractgen`. | View-schema owners projected through `package.protocol_ts`. | High | Surface meaning, field capability, and write behavior remain owner/view-contract concerns. |
| `packages/protocol-ts/src/generated/ws-artifacts.ts` | Embeds the WebSocket contract index and digest. | Protected values surfaced through root WS artifact helpers. | Generated barrel and authored root. | `artifact.ts`; `contracts/ws/index.schema.json`. | Package artifact test indirectly and collaboration consumers. | WS contract family generated by `tools/contractgen`. | Collaboration transport projection. | High | Separate from the generated collaboration TypeScript union and runtime decoder. |

Final authored target additions are exhaustive:

- `packages/protocol-ts/src/public-compatibility.compile.ts` owns the real
  TypeScript-project compile matrix for all five specifiers and 17 compatibility
  candidates.
- `packages/protocol-ts/src/facade/runtimeValidation.ts` owns generated HTTP
  binding projection and payload-safe decoder construction.
- `packages/protocol-ts/src/facade/collaboration.ts` owns the incident-stream
  decoder facade.
- `packages/protocol-ts/src/facade/extensionDiscovery.ts` owns tolerant,
  fail-closed extension discovery decoding.
- `packages/protocol-ts/src/facade/contractArtifacts.ts` owns curated direct
  artifact-family imports and registry helpers.
- `packages/protocol-ts/src/facade/networkFlow.ts` owns Network Flow descriptor,
  registry, error, and decoder projection without owning Network Flow behavior.
- `packages/protocol-ts/src/facade/compatibilityTypes.ts` owns the retained root
  compatibility declarations and the two proven aliases.

No new private facade file is a removable forwarding shim: each owns runtime
behavior, a least-privilege generated-import boundary, or continuing public
compatibility. The package export map still exposes exactly the original five
specifiers.

## 3. Module Boundary Diagnosis

The current target is a legitimate transport-adjacent protocol facade and
generated-projection boundary, but it is not a permanent semantic module for
the many domain names present in its generated types. The authored root is also
a mixed-responsibility package surface: it combines static type exposure,
runtime schema decoding, extension-specific tolerant decoding, contract-byte
lookup, registry interpretation, Network Flow adapter construction, and
handwritten compatibility types.

The target is therefore classified as:

- a legitimate protocol package facade, but not currently a thin facade;
- a transport-adjacent adapter and generated-contract projection boundary;
- a mixed-responsibility authored root;
- a raw view/projection contract exposure layer, not projection orchestration;
- a mutation-envelope and revision-contract exposure layer, not a mutation
  coordinator;
- not a persistence adapter, frontend shell/controller, test-util module, or
  grid-vendor integration layer; and
- a misplaced permanent home for any semantic behavior inferred only from the
  generated names of timeline, imports, entities, indicators, evidence, links,
  saved views, revisions, collaboration, or projections.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Package entrypoints and stable root facade | `package.json`, `src/index.ts` | `package.protocol_ts` | keep | Declared workspace exports and import-boundary policy. | Preserve every current export path and root symbol by default. |
| Generic generated-schema decoding and safe failure projection | `src/index.ts` | Private `package.protocol_ts` facade module | split | `createGeneratedDecoder`, generated validators, payload-free `DecodeFailure`. | Does not own domain behavior. |
| Collaboration message wire decoder | `src/index.ts`, generated collaboration types | Protocol facade for decode; collaboration module/web facade for lifecycle | split | `incidentStreamMessageDecoder`; app collaboration consumer. | Keep WebSocket connection/resume/application effects outside the package. |
| Compatible extension discovery item decoding | `src/index.ts` | Protocol extension-decoder facade implementing adopted Extensions algorithm | split | Seven known members, unknown additive member removal, sorted/token checks. | This tolerant decoder is owner-required behavior, not generic JSON parsing. |
| Static contract artifact lookup and registry access | `src/index.ts` | Protocol artifact facade; semantic interpretation by each owner adapter | split | Artifact family arrays and parse/get/require helpers. | Unchecked generic parse is safe only for immutable packaged bytes followed by owner adapter validation. |
| Network Flow wire decoders and metadata access | `src/index.ts` plus generated Network Flow files | Protocol Network Flow facade; semantics in Network Flow owners | split | Nine decoders, descriptor, error, mapping, presentation registries. | Must not absorb graph, import, indicator, or workspace behavior. |
| HTTP operation metadata and generated request/response maps | Generated HTTP/type files | `package.protocol_ts` generated projection of exact route owners | keep | 66 selected operations and three tolerant responses. | Generated files remain unchanged by structural facade work. |
| Handwritten account wire declarations | `src/index.ts` | Generated Core HTTP types behind root compatibility aliases | move | Equivalent member families also exist in `core-http-types.ts`. | Prove TypeScript assignability before replacing declarations. |
| Handwritten evidence/view compatibility declarations | `src/index.ts` | Root compatibility surface until later type-contract decision | move | Generated variants differ in readonlyness, required values, index signatures, header values, and empty-object typing. | Convergence is not automatically behavior-preserving. |
| View-schema bytes/source types | Generated files and root helpers | Raw exposure in protocol package; interpretation in `package.view_contracts` | keep | `packages/view-contracts/src/index.ts` validates and maps raw documents. | Do not move field capability or inspector behavior into protocol-ts. |
| Import target registry projection | Generated registry and root re-export | Raw projection in protocol package; semantics in imports adapter/module | keep | `importTargetContractAdapter.ts` partitions and validates rows. | No parser, mapping, apply, or lifecycle behavior belongs here. |
| Audit and revisions artifact exposure | Generated barrel only | Respective owners plus protocol package projection policy | keep | Generated outputs exist, but authored root omits their artifact families and no supported browser consumer requires them. | They **MUST** remain generated and protected, **MUST NOT** enter root helpers or package exports, and **MUST** be unreachable from production browser bundles. |
| Public family subpath strategy and root bundle breadth | `package.json`, root/generated barrels | `package.protocol_ts` package API with consumer owners | keep | Only `./core-http` currently has repository consumers, but repository reachability cannot prove external safety. | The root and all four subpaths **MUST** remain unchanged. Optional optimization is governed by S-05 and defaults to no API change. |

### Generated artifact browser disposition

Browser exposure is a closed allowlist. Generation alone never grants browser
eligibility. Authored browser-runtime code **MUST** import each approved family
directly through the private facade and **MUST NOT** import the protected
generated barrel.

| Generated artifact family | Current root exposure | Required browser disposition | Required enforcement |
| --- | --- | --- | --- |
| WebSocket | Root artifact helpers | Approved through the stable facade. | Direct private-facade import; preserve artifact bytes, digest, order, and helper results. |
| View schemas | Root artifact helpers | Approved through the stable facade and view-contract adapter. | Direct private-facade import; preserve registry order and lookup behavior. |
| Errors and reason codes | Root artifact helpers | Approved through the stable facade. | Direct private-facade import; preserve tolerant owner-defined semantics. |
| Extensions | Root artifact and discovery helpers | Approved through the stable facade. | Direct private-facade import; preserve seven-member tolerant decoding and unknown-member inertness. |
| Network Flow | Root artifact and Network Flow helpers | Approved through the stable facade. | Direct private-facade import; preserve contract-major, mapping, presentation, error, and decoder identity. |
| Administrative audit | No root helper or package entrypoint | Generated and protected; forbidden from browser runtime. | Negative import-boundary fixture plus production-bundle absence evidence; no `./audit` entrypoint. |
| Revisions and conflict-token key-ring schema | No root helper or package entrypoint | Generated and protected; forbidden from browser runtime. | Negative import-boundary fixture plus production-bundle absence evidence; no `./revisions` entrypoint and no secret/key material in frontend artifacts. |

## 4. Public Contract and Behavior Freeze Map

The package does not implement server behavior, but moving its facade can break
compile-time contracts, runtime decoding, browser request construction, and
consumer admission. Those observable effects are frozen below.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Root and four declared package entrypoints | `package.protocol_ts` package surface | `package.json`, root facade, direct import inventory. | Typecheck, package tests, consumer tests. | Import/compile smoke for every entrypoint and exact root compatibility exports. | High | No entrypoint removal or rename is authorized. |
| 66 selected HTTP operations | Exact Core/adopted route owners; protocol projection selection owned by `package.protocol_ts`. | HTTP selection contract, generated bindings, OpenAPI. | Package HTTP binding test; consumer adapter tests. | Preserve exact operation ID, method, path template, path parameters, query members, statuses, response schema, and tolerance. | High | Package constructs/validates requests; it does not handle routes. |
| Three forward-tolerant HTTP responses | Core route owners and protocol selection input | HTTP selection contract. | Package test covers administrative audit and extension discovery paths. | Explicitly freeze the three-operation set and unknown-value handling. | High | Do not generalize tolerance to closed responses. |
| Standard success/error envelopes | Core 01 | Generated Core HTTP types, error artifacts, browser transport. | Package error tests and consumer tests. | Payload-free validation failure and error-envelope decode characterization. | High | Error mapping and authorization outcomes remain server owned. |
| Workbook query, row, create, patch, bulk, paste, conflict, and startup wire shapes | Core 01/Core 03 and source owners | Generated Core HTTP types/bindings and workbook adapters. | Package HTTP test; workbook unit/browser evidence. | Preserve request/response assignability and malformed-success rejection for every moved adapter surface. | High | No row/query/mutation semantics are owned locally. |
| Saved-view, `SheetRef`, view-schema, and workbook-preference shapes | Core 01/Core 03 | Core HTTP types, view artifacts, view-contracts, sheet-ref consumers. | Package registry test, view-contracts/UI tests, saved-view browser rows. | Entry/subpath type compatibility and registry ordering/lookup characterization. | High | Visible labels are never identity. |
| Projection refresh/invalidation-facing wire data | Core 01/Core 03; Graph Projection only for its named subsystem | View rows, WebSocket `record_changed`/extension messages, Network Flow graph results. | Collaboration, workbook, and Network Flow tests. | Decoder/application boundary characterization only where facade movement touches it. | High | Protocol-ts performs no refresh or projection lifecycle. |
| Fourteen WebSocket message types and incident-stream decoder | Core 01 transport, Core 03 behavior, adopted Extensions where applicable | WS contract, generated union, decoder. | Consumer collaboration tests; package lacks focused coverage. | Valid replayable/non-replayable examples, unknown type/member rejection posture, and no-payload failure. | High | Authorization recheck and event application remain outside target. |
| Extension discovery seven-member producer/compatible-client decoder | Core 01 plus adopted Extensions NLSpec | Custom root decoder and generated envelope validator. | Package discovery test. | Known-member malformed cases, sorted uniqueness, unknown profile/workspace/member inertness, and fail-closed complete-response behavior. | High | Current major requires capabilities to remain empty. |
| Evidence blob creation, upload target, attach, preview/download handle shapes | Core 01/Core 02/Core 04 | Generated Core HTTP types and handwritten root compatibility types. | Package type test, evidence service tests, targeted browser evidence. | Manual/generated assignability matrix and runtime HTTP validation before type convergence. | High | Storage keys, authorization, and lifecycle remain server owned. |
| Account profile/preferences and deployment administration shapes | Core 01/Core 04 | Generated Core HTTP types/bindings and handwritten root account types. | Package type test and app shell/admin tests. | Readonly and request/response alias compatibility. | High | Package performs no authentication or authorization. |
| Import session/unit/mapping/job shapes and generated target registry | Core 01 generic imports; target-specific owner for target schema | HTTP bindings/types, import target registry, frontend import adapters. | Package HTTP test and module-imports collaborator tests. | Target registry partition/integrity and request/response compatibility. | High | Clipboard behavior and file import remain distinct owners. |
| Revision, history, rollback, delete/restore, and conflict shapes | Core 01/Core 02/Revisions | HTTP bindings/types and revisions artifacts. | Package HTTP validation and workbook history/conflict tests. | Preserve aliases and decoder outcomes used by workbook adapters. | High | No mutation grouping or side effect occurs in protocol-ts. |
| Network Flow descriptor, types, decoders, mapping, presentation, errors, and artifacts | Adopted Network Flow NLSpec; Graph Projection for imported graph output behavior | Generated files and authored Network Flow root facade. | Package Network Flow test; module/web Network Flow rows. | Decoder matrix for all public result schemas and exact failure redaction; metadata identity parity. | High | Flow rows never become Core records through this package. |
| View-schema artifact/source projections | Core 01 with Core 02/Core 03 consequences | Generated artifacts/source types and view-contracts adapter. | Package registry, view-contracts, UI selectors, browser tests. | Preserve artifact bytes/digests, registry order, and raw source type exposure. | High | Semantic validation stays in view-contracts. |
| Administrative-audit and revisions artifact browser disposition | Core 01 for public HTTP shapes, Core 04 for authorization/redaction/secrets, respective subsystem owners for internal registries | Generated artifact modules, current root-helper omission, package export map, and frontend boundary policy. | Drift/policy checks; browser absence row is required by S-06. | Direct-import policy, negative source fixtures, and initial/dynamic production-chunk absence. | High | Generated presence does not grant browser eligibility; browser consumers use existing public HTTP types. |
| Generated artifact and drift boundary | Testing Harness and generated-artifact policy for mechanics; each contract owner for behavior | Generated markers, generator inputs, generation targets. | Drift, policy, and JSON-shape checks. | No new characterization beyond official regeneration equivalence. | High | Generated files must never be hand-edited. |
| Grid adapter and UI selector contracts | Grid/UI package owners | No grid-vendor import in target; UI selectors consume registry identity indirectly. | Import-boundary, UI-contracts, grid/browser tests. | Only owner-aligned consumer tests if registry exposure changes. | Medium | No grid coordinate or vendor behavior may enter protocol-ts. |
| Harness/test accounting | Testing Harness NLSpec and authored owner catalogs | Five package rows: three static and two unit families; zero service-backed rows. | `make test-slice OWNER=package.protocol_ts`. | Add exact titles to the authored package family only when characterization titles are added. | Medium | Routing proves execution accounting, not specification completeness. |

### Public package specifier freeze

The following list is exhaustive. The refactor **MUST NOT** add, remove, rename,
or migrate a public entrypoint. Private `src/facade/**` paths **MUST NOT** become
package exports.

| Supported specifier | Current target | Required refactor behavior | Default when no optimization study runs |
| --- | --- | --- | --- |
| `@cartulary/protocol-ts` | `src/index.ts` | Retain as the compatibility composition facade with the same exported names, types, values, initialization behavior, and artifact ordering. | Keep unchanged. |
| `@cartulary/protocol-ts/collaboration` | `src/generated/collaboration-types.ts` | Preserve resolution and the complete generated collaboration surface. | Keep unchanged. |
| `@cartulary/protocol-ts/core-http` | `src/generated/core-http-types.ts` | Preserve resolution and the complete generated Core HTTP surface. | Keep unchanged. |
| `@cartulary/protocol-ts/http` | `src/generated/http-operation-bindings.ts` | Preserve resolution, all 66 operation bindings, encoding, aliases, validation, and the three tolerant operations. | Keep unchanged. |
| `@cartulary/protocol-ts/network-flow` | `src/generated/network-flow-types.ts` | Preserve resolution and the complete generated Network Flow type surface. | Keep unchanged. |

### Public compatibility-type disposition

A declaration may become a generated alias only when S-00 proves equivalent
caller acceptance, generated-to-legacy output compatibility, permitted writes,
and runtime schema behavior. Structural assignability by itself is
insufficient.

| Variance class | Required decision | Permitted implementation |
| --- | --- | --- |
| Exact structural and behavioral equivalence | Alias is permitted after characterization. | Replace the handwritten member list with a direct generated type alias and retain the current root-exported name. |
| Optional versus required property | In-place convergence is prohibited. | Preserve the public compatibility declaration; use the generated type internally or through `./core-http`. |
| Mutable versus `readonly` property | In-place convergence is prohibited unless write capability is identical. | Preserve every currently permitted caller write and characterize writes explicitly. |
| Index signature added, removed, narrowed, or widened | In-place convergence is prohibited. | Preserve current undeclared-key lookup behavior; do not add a broad index signature to silence errors. |
| Header value `string` versus `unknown` | Owner resolution is required. | Correct an owner projection defect at its source, or retain `unknown` at the canonical trust boundary with checked narrowing. Preserve any existing public `string` declaration until a versioned migration. |
| `Record<string, never>` versus generated empty interface | Treat as non-equivalent. | Preserve the closed empty-object compatibility type or correct the generated owner projection; do not weaken it to an open interface. |
| `any`, double assertion, or schema widening used to force compatibility | Prohibited. | Retain the compatibility type, correct the owner projection, or introduce an explicit checked adapter. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Handwritten account, evidence, and view wire declarations overlap generated Core HTTP declarations. | `src/index.ts` versus `generated/core-http-types.ts`; examples include optional versus required cell values, broad row index signatures, readonly differences, string versus unknown header values, and `Record<string, never>` versus generated empty interface. | Public TypeScript drift can accept or reject different caller code even when runtime JSON is unchanged. | `must_fix` | Generated Core HTTP projection plus root compatibility facade. | S-00 **MUST** inventory every candidate symbol. Internal consumers **MUST** use canonical generated types; only proven-equivalent declarations **MAY** alias, and every divergent public declaration **MUST** remain exact. |
| The authored root combines six separable protocol responsibilities. | 692 lines, 65 export statements, decoder factory, discovery decoder, artifact helpers, Network Flow setup, compatibility types, generated re-exports. | Small changes can initialize or disturb unrelated contract families. | `should_fix` | Private `src/facade/` modules composed by the stable root. | Split only after root characterization; preserve all exports. |
| Root imports selected artifacts through a broad protected generated barrel. | `src/index.ts` imports from `generated/index.ts`, whose exports also include audit, revisions, and import registry. | Hidden browser reachability and coupling to prohibited artifact families. | `must_fix` | Private protocol artifact facade and frontend import-boundary owner. | S-01 **MUST** import each approved artifact family directly and authored browser-runtime code **MUST NOT** import `generated/index.ts`. |
| Public family subpaths exist but only `./core-http` has current repository consumers. | Package export map and import search; Node package exports define the supported package interface. | Removing unused-looking paths can break consumers outside static repository reachability. | `intentional/no_action` | `package.protocol_ts` public API owner. | The root and all four subpaths **MUST** remain. S-05 is optional and defaults to no package API change. |
| Audit and revisions artifact files are generated and barrel-exported but omitted from root artifact helper lists. | Generated artifact files/barrel and root `contractArtifactLists`; no supported browser consumer was found. | Generic exposure can pull server/trust-boundary contracts into browser runtime and misstate public API. | `must_fix` | Audit/Revisions owners with protocol package and frontend-boundary owners. | Keep both families generated/protected/root-omitted/export-omitted; add import-boundary and production-bundle absence evidence. |
| Blanket public-export reachability suppression prevents repository dead-code evidence from authorizing removal. | `.fallowrc.json` ignores all root exports and marks the package public. | False confidence from zero in-repo callers. | `intentional/no_action` | Static-analysis/package API policy. | Treat every current export as compatibility surface until an authorized API review. |
| The package exposes raw static JSON through a generic typed parse cast. | `parseContractArtifact<T>`; view-contracts performs subsequent invariants. | A consumer can assert a wrong compile-time type if it skips an owner adapter. | `defer` | Raw artifact facade plus semantic owner adapters. | Preserve current behavior; assess typed family helpers only in a separate API-design slice. |
| Protected generated imports are limited to the authored root today. | `frontend-generated-protocol-boundary` allows only `src/index.ts`. | A private facade split would violate current policy unless policy moves with it. | `must_fix` | Frontend import-boundary owner. | Plan the policy/test update in the same structural slice; never weaken protection outside `src/facade/**`. |
| No platform storage, SQL, authorization decision, route handler, WebSocket lifecycle, projection refresh, React state, test-util choreography, or grid-vendor code exists in the target. | Direct target search and dependencies show generated/typed protocol concerns only. | Moving unrelated semantic behavior here would create a new violation. | `intentional/no_action` | Existing semantic modules and frontend package owners. | Preserve the negative boundary and add no semantic convenience helpers. |
| Generated files are downstream projections with two official generator paths. | Generated markers, `tools/contractgen`, protocol generator, and generated-artifact policy. | Hand edits would drift or be overwritten. | `intentional/no_action` | Contract owners and Make-owned generators. | Plan source/input changes before regeneration; never patch generated files. |
| No test-only import or helper leaks into production protocol source. | Package runtime dependencies and source imports exclude test-utils, Playwright, Vitest, Node builtins, and grid adapter. | Low current risk; regression would pollute browser runtime boundary. | `intentional/no_action` | `package.protocol_ts` and frontend boundary policy. | Retain runtime/test-helper and Node builtin checks. |
| Guide shorthand calls the package generated protocol types, while live state also contains an authored facade and decoders. | Development guide tree shorthand versus package source and the guide's more precise frontend boundary text. | Planning could incorrectly treat all target files as generated. | `should_fix` | This tracker as implementation-support clarification; guide owner if later edited. | Record the mismatch; do not change guide or behavior in this task. |

No owner contradiction was found in the inspected scopes. The adopted
Extensions and Network Flow documents align on the generic seven-member
discovery item and Network Flow contract major 2. If a later session finds two
current owners that conflict, it **MUST** record
`BLOCKED: owner contradiction` and **MUST** stop the affected slice. No Core
amendment is required by RB-001 through RB-003 because their dispositions
preserve current public wire, package, authorization, and workbook behavior.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Freeze branch, commit, constraints, authority, and generated-file posture. | Tracker only during planning. | `git status`, target/output existence, source list. | Scope, authority, and no-implementation statement are current. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Account for every tracked target file, caller, dependency, test, and generator. | `packages/protocol-ts/**` plus direct caller/generator evidence. | Read-only file/import inventory; `make task-guide ROLE=module-author OWNER=package.protocol_ts`. | All 24 files have one inventory row. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Separate wire projection ownership from semantic owner behavior. | Owner documents, contract selections, generator inputs, package facade. | Human owner review plus `make explain-test-owner OWNER=package.protocol_ts`. | Every contract risk names an owner and test posture. |
| WF-03 | Characterization test gap analysis | parallel | WF-01, WF-02 | WF-05, WF-06 | Identify pre-move evidence for each public root/subpath and decoder/type seam. | `src/index.test.ts`, family manifests, consumer tests. | Existing package slice and owner-aligned test discovery. | Missing entrypoint, WS, decoder, and type compatibility cases are explicit. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Identify illegal dependencies, mixed responsibilities, and intentional negative boundaries. | Root facade, package metadata, import-boundary and Fallow policy. | `make frontend-import-boundary-check`. | Findings are classified with required actions. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06, WF-07 | Define private facade seams, direct approved-family imports, exact compatibility preservation, and the fixed five-specifier surface. | Future `src/facade/**`, stable root, boundary policy. | Characterization, typecheck/import boundary, production build, and protected-family reachability evidence. | Private seams, compatibility dispositions, and browser artifact allowlist are decision-complete. |
| WF-06 | Slice sequencing plan | chain | WF-03, WF-05 | WF-07, WF-08 | Order characterization, movement, policy, alias, and optional public-surface work into rollback-safe slices. | Files named per Section 7. | Validation attached per slice. | Every slice has dependencies, rollback, and completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-05 | WF-08 | Keep new test titles, boundary policy, and browser-reachability evidence represented by authored owner inputs without treating routing as behavior. | Package family manifest, frontend boundary input/test, private bundle helper; generated harness outputs only through Make when required. | Owner slice, boundary, build, bundle-reachability row, JSON shape, generation drift. | Required authored evidence is specified; no generated hand edit or Markdown-dependent test is planned. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Prove the selected implementation slice and leave current restart evidence. | Changed implementation slice plus tracker update in a later task. | Narrow owner checks, then static/generated/build/full gates according to risk. | Commands, results, blockers, and next slice are recorded. |

WF-00 through WF-08 are complete as planning workflows in this tracker. Their
future implementation artifacts remain `TODO` or `DEFERRED` in Section 9.

## 7. Proposed Refactor Slice Plan

These are the authorized implementation slices. Every required slice preserves
the current public package surface; runtime or wire behavior changes only where
S-04 repairs an owner/projection disagreement. Any other non-equivalent public
type or entrypoint change is outside this remediation and requires a separately
authorized versioned API migration.

Current execution state: tracker activation, S-00 through S-04, the full S-06
red/green workstream, and S-07 are complete. S-05 remains explicitly
`DEFERRED`. No implementation slice remains open.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Add characterization for all five package entrypoints, collaboration decoding, payload-free decoder failures, extension tolerant decoding, and every manual/generated compatibility candidate. Create `docs/decisions/protocol-ts-public-type-compatibility.md` with the exact interface below. | Package tests, compile fixtures, authored package family manifest only when selectors change, and the decision record. | A test that encodes a proposed stricter type or decoder behavior would manufacture a contract. | Existing ten titles; positive and negative compile fixtures; representative accepted/rejected runtime payloads; root and `./core-http` import cases. | `make test-slice OWNER=package.protocol_ts`; `make frontend-typecheck`; `make json-shape-check` when the manifest changes. | Revert the characterization, authored selector updates, and decision record together. | Every candidate symbol has exactly one complete decision row, all required compile/runtime cases pass, and no product behavior or generated file changes. |
| S-01 | S-00; S-06 bundle row | Split the authored root into private `src/facade/runtimeValidation.ts`, `collaboration.ts`, `extensionDiscovery.ts`, `contractArtifacts.ts`, `networkFlow.ts`, and `compatibilityTypes.ts`. `src/index.ts` remains the exact public composition facade. `contractArtifacts.ts` imports approved generated families directly and no authored browser-runtime file imports `generated/index.ts`. | Authored package source only; generated files untouched. | Export identity, initialization order, decoder lookup, artifact bytes/order, and prohibited audit/revisions reachability. | All S-00 characterization, package tests, entrypoint smoke, artifact identity/order, and bundle absence evidence. | `make test-slice OWNER=package.protocol_ts`; `make frontend-typecheck`; `make build-web`; bundle-reachability row once catalogued. | Revert the private files and root composition as one slice. | All five entrypoints and root symbols are unchanged; approved helpers return identical values; audit/revisions modules and embedded bytes are absent from production browser chunks. |
| S-02 | S-01 | Restrict protected generated imports to the private protocol facade and keep every other consumer on package exports. Add positive private-facade fixtures and negative non-facade, audit, and revisions fixtures. | Frontend import-boundary configuration/test plus authored facade imports. | An overbroad allowlist could expose generated internals or server/trust-boundary artifacts. | Existing boundary fixtures; approved facade positive case; non-facade, broad-barrel, audit, and revisions negative cases. | `make frontend-import-boundary-check`; `make json-shape-check`; `make test-slice OWNER=package.protocol_ts`. | Revert policy, fixtures, and facade imports together. | Only named private facade files can import approved generated families; no authored browser file can import the generated barrel, audit, or revisions modules. |
| S-03 | S-00, S-01 | Replace an account/profile/preferences declaration with a generator-backed root alias only when its S-00 row has disposition `alias`. Otherwise retain the exact compatibility declaration. | Private compatibility facade, root composition, package tests. | Mutability, optionality, generic inference, and diagnostic surface can change despite apparent structural compatibility. | Bidirectional assignment, property writes, omission/`undefined`, excess-property, generic-position, and root/subpath import fixtures for every candidate. | `make test-slice OWNER=package.protocol_ts`; `make frontend-typecheck`; affected app owner unit rows. | Revert each alias group as one compatibility unit. | Every changed declaration satisfies all equivalence fields; unchanged candidates retain their exact public member lists. |
| S-04 | S-00, S-01 | Clarify Core 01's string-valued upload headers, repair authored OpenAPI row/cell openness and closed-empty generation, regenerate through Make, alias only the now-equivalent handle request, and adopt generated operation types internally. | Core 01, authored OpenAPI owners and release change set, protocol generator and projections, private compatibility facade, Evidence consumers, compile/runtime tests. | Optionality, readonlyness, index signatures, header values, closed-empty-object behavior, additive response handling, release compatibility, and generated drift. | S-00 compile/runtime matrix; numeric-header rejection; additive row/cell acceptance; Evidence and Workbook owner unit/integration/browser rows. | Generation/drift/policy/compatibility checks; package/OpenAPI owner slices; typecheck/import/build/reachability; affected Evidence/Workbook rows. | Revert owner inputs, generator, generated projections, internal adoption, and release authorization as one slice; divergent root declarations remain stable. | Owners and projections agree, generic closed-empty generation fails closed, internal canonical uses are operation-specific, all divergent root types remain exact, and affected owner evidence passes. |
| S-05 | S-01, S-02 | **Optional optimization study.** Retain the current package surface. A candidate family-entrypoint design may be proposed only under the measurement contract below. Failure to run the study or meet every gate results in no package API change. | Retained production build-analysis evidence only unless a separately authorized candidate proceeds. | Public imports, initialization timing, side effects, chunk reachability, cycles, and undeclared external consumers. | Five-entrypoint smoke; exact export comparison; normalized module/chunk/cycle comparison across two builds per variant. | `make build-web`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make check` only for a separately authorized retained candidate. | The unchanged root and four subpaths are the mandatory fallback. | Either no study runs and status remains `DEFERRED`, or retained evidence satisfies all compatibility gates and one materiality gate before a new proposal is authorized. |
| S-06 | S-00 and any selector/evidence change | Update authored package test-family selectors and verification collaboration only for new or moved evidence. Add a Make-owned package row for bundle reachability before S-01 claims browser exclusion; generated harness outputs remain Make-owned. | Authored package test-family/verification inputs and Make-generated harness projections. | Missing, duplicated, or misclassified rows can hide evidence; Markdown cannot become machine authority. | Catalog/shape validation, exact owner resolution, and bundle-reachability row execution after `make build-web`. | `make json-shape-check`; `make agent-finalize` when tracked harness projections require refresh; `make test-slice OWNER=package.protocol_ts`. | Revert authored and generated harness updates as one slice while preserving unrelated evidence. | Every active title/check resolves exactly once, package-owned service-backed rows remain zero unless explicitly authorized, and no test reads Markdown. |
| S-07 | S-01, S-02, S-03, S-04, and any required S-06 | Remove only wrappers proven obsolete after compatible callers move; divergent public compatibility declarations remain. Run the complete risk-based validation ladder and update the handoff. S-05 is not a dependency. | Changed facade/consumers/policy/tests and tracker in a later implementation task. | Premature compatibility removal, hidden export loss, generated drift, browser artifact leakage, or owner regression. | All package characterization, browser-reachability evidence, and affected owner-aligned unit/browser rows. | Narrow owner checks, static/generated/build checks, then `make check`. | Revert only the failing slice; do not combine unrelated cleanup. | No removable wrapper remains, all intentional compatibility declarations are retained, required checks pass, and the handoff is current. |

### S-00 compatibility decision-record interface

The S-00 record **MUST** contain exactly one row per affected public symbol.
Tests **MUST NOT** parse, hash, stat, or otherwise depend on the Markdown file.

| Required field | Required meaning |
| --- | --- |
| `public_symbol` | Current root-exported name. |
| `generated_symbol` | Candidate generated Core HTTP name. |
| `contract_owner` | Exact Core 01 section or adopted subsystem owner. |
| `wire_position` | Exactly one of `request`, `response`, `shared`, or `artifact_source`. |
| `property_variances` | Exhaustive optionality, nullability, readonly, index-signature, scalar, and empty-object differences. |
| `legacy_to_generated` | Whether every current caller value is assignable to the generated type. |
| `generated_to_legacy` | Whether every generated value satisfies the current declaration. |
| `write_capability_equal` | Whether callers retain exactly the same permitted writes. |
| `runtime_schema_equal` | Whether both declarations describe the same accepted and rejected JSON. |
| `disposition` | Exactly one of `alias`, `retain_compatibility`, `owner_projection_defect`, or `future_versioned_change`. |
| `removal_gate` | Binary conditions required before deleting the compatibility declaration. |

Compile characterization **MUST** cover bidirectional assignment, writes,
omission versus explicit `undefined`, arbitrary-key lookup, excess object
members, un-narrowed header access, generic parameter and return positions, and
imports from both the root and `./core-http`. Negative expectations **MUST** use
compile-time assertions such as `@ts-expect-error`. Runtime characterization
**MUST** validate representative accepted payloads, reject owner-forbidden
payloads, preserve payload-free failures, invent no omitted defaults, preserve
owner-required additive tolerance, and add no undeclared request namespace.

### S-05 optional measurement contract

Decision recorded 2026-08-03: **DEFERRED**. S-01 removed the demonstrated
security and bundle defect without expanding the public API. No additional
entrypoint or permanent measurement infrastructure has sufficient continuing
value absent a separately prioritized study that satisfies this contract.
Deferral is not a dependency or blocker for S-07.

Every candidate **MUST** preserve all five current specifiers and exported
symbols, decoder behavior, artifact order, registry identity, error projection,
initialization order, generated-file protection, and workbook initial loading.
It **MUST NOT** introduce a production strongly connected component or enlarge
the largest existing component.

A candidate is material only when it satisfies at least one gate:

1. It reduces initial production JavaScript attributable to the affected entry
   by at least 10 KiB gzip **and** 3 percent, with no regression greater than
   2 KiB gzip or 1 percent on another primary entry.
2. It removes at least one production static-import cycle involving
   `protocol-ts`, introduces no static or dynamic cycle, and causes no bundle
   regression greater than 2 KiB gzip.

The comparison **MUST** use the actual `make build-web` path with identical
source except for the candidate, identical lock/tool/build inputs, production
mode, normalized static and dynamic graphs, chunk membership, entry
reachability, raw/gzip/Brotli sizes, strongly connected components, side-effect
findings, and duplicate modules. Each variant **MUST** produce the same
normalized graph and size totals in two clean builds. The helper **MUST** remain
private and Make-owned; it **MUST NOT** add a public target without a separate
Testing Harness/task-surface owner revision.

## 8. Validation Plan

Canonical commands were discovered through `make help-all`,
`make task-guide ROLE=module-author OWNER=package.protocol_ts`,
`make explain-test-owner OWNER=package.protocol_ts`, and target explanations.
The package owns five active Vitest rows and zero service-backed rows.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=package.protocol_ts` | All five active package rows and ten current titles. | yes | Passed baseline: `.cartulary/test-results/20260803T221341Z-p2950064`. Run before and after every implementation slice. |
| compile compatibility | `make frontend-typecheck` | Actual workspace compiler configuration, including `strict`, `exactOptionalPropertyTypes`, and `noUncheckedIndexedAccess`; S-00 positive/negative fixtures. | yes | Passed pre-refactor baseline: `.cartulary/test-results/20260803T221403Z-p2953658`. S-00 **MUST** record effective options without changing them. |
| integration | `TODO: no package-owned integration or service-backed row` | Owner-aligned consuming module only when a changed contract crosses into that module. | no | Discover exact affected rows with `make explain-test-owner OWNER=<owner-id>`; do not invent a generic protocol integration row. |
| e2e/browser | `make browser-e2e` | Broad browser behavior affected by HTTP/WS/type changes. | no | Prefer an exact owner row where known, such as `make test-slice OWNER=module.evidence ROWS=module.evidence.browser.verify_attach_flow_uses_generated_protocol_types_7a778c9178`; required after an authorized evidence wire change, not for a private no-behavior move by default. |
| generated drift | `make generate-drift` | Official regeneration equivalence. | yes | Passed baseline: `.cartulary/test-results/20260803T221351Z-p2950824`. Also use `make generated-artifact-policy-check` (passed `.cartulary/test-results/20260803T221403Z-p2953467`) and `make json-shape-check` (passed `.cartulary/test-results/20260803T221403Z-p2953462`). |
| import-boundary/static | `make frontend-import-boundary-check` | Generated-root protection, workspace package boundaries, direct approved-family imports, and negative broad-barrel/audit/revisions cases. | yes | Passed baseline: `.cartulary/test-results/20260803T221346Z-p2950477`. S-02 **MUST** add the new positive and negative fixtures before claiming the revised boundary. |
| production build | `make build-web` | Actual Vite/Rollup browser output and module initialization. | no | Required for S-01, S-02, S-04, and any S-05 study. A successful build alone does not prove audit/revisions absence. |
| browser artifact reachability | `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b` | Absence of generated audit/revisions modules and their embedded contract bytes from every emitted JavaScript chunk and source map. | yes | S-06 red failed as expected at `.cartulary/test-results/20260803T233505Z-p3017178`, with both generated modules plus all three forbidden paths, digests, schema IDs, and embedded JSON signatures. The identical row passed after S-01 within `.cartulary/test-results/20260803T234215Z-p3041053`. |
| full check | `make check` | Developer verification gate. | no | Required before final implementation handoff after all selected slices; not required for this documentation-only tracker. |
| documentation | `make lint-markdown` | Authored Markdown structure. | no | Revised tracker passed: `.cartulary/test-results/20260803T225833Z-p2974554`. |

The baseline commands prove the inspected repository state passed those checks
at their named run roots. They do not prove specification completeness, future
slice correctness, product conformance, or Core 05 claim publication.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| PT-001 | Freeze target, constraints, and authority posture | WF-00 | DONE | none | Section 1; branch/commit and source hierarchy. | One target, one allowed output, and explicit non-goals are recorded. |
| PT-002 | Inventory all tracked target files and ignored cache exclusions | WF-01 | DONE | PT-001 | Section 2; 24 tracked rows. | Every tracked file has exactly one row; caches are explicitly excluded. |
| PT-003 | Map owner documents and projection boundaries | WF-02 | DONE | PT-002 | Sections 1, 3, and 4. | Every discovered public risk names a current owner or defer decision. |
| PT-004 | Freeze public package and wire behavior | WF-03 | DONE | PT-003 | Section 4 and existing test matrix. | Root/subpaths, HTTP, WS, decoder, generated, and harness surfaces are listed. |
| PT-005 | Diagnose mixed responsibilities and dependency direction | WF-04 | DONE | PT-002, PT-003 | Sections 3 and 5. | Every finding is classified and has a required planning action. |
| PT-006 | Add missing characterization coverage | WF-03 | DONE | PT-004 | S-00 compile fixture, runtime characterization, decision record, package family selectors, and generated topology index. | All five entrypoints, 17 compatibility candidates, 14 WS families, safe failures, extension discovery, exports, artifacts, and identities are frozen. |
| PT-007 | Split the authored root into private facade modules and remove the broad generated-barrel import | WF-05 | DONE | PT-006, PT-017 | S-01 six-module private facade, composition-only root, characterization evidence, web build, and reachability row. | Stable root composes private seams, approved families are imported directly, public behavior is unchanged, and protected artifacts are absent from browser output. |
| PT-008 | Keep generated imports protected after facade split | WF-04, WF-07 | DONE | PT-007 | S-02 least-privilege import rules, extension-normalized path matching, and positive/negative checker fixtures. | Only the owning private facade modules can import approved generated files; barrel, audit, and revisions imports are always denied. |
| PT-009 | Replace proven-equivalent account declarations with generated-backed aliases | WF-05 | DONE | PT-006, PT-007 | S-03 `DensityMode` alias, operation-specific app-shell types, compile matrix, and affected auth/application rows. | `DensityMode` has no independent root member list; six readonly account compatibility declarations remain exact and application source does not import them. |
| PT-010 | Define divergent evidence/view compatibility policy | WF-05 | DONE | PT-004 | RB-001 and the Section 4 variance table. | Equivalent aliases only; every non-equivalent root declaration is preserved and S-04 is behavior-preserving. |
| PT-011 | Fix the current public package-surface decision | WF-05 | DONE | PT-004 | RB-002 and the Section 4 specifier table. | Root plus four existing subpaths are the exhaustive unchanged surface; optimization is optional. |
| PT-012 | Define harness/accounting updates for new characterization | WF-07 | DONE | PT-004 | Section 6/7 and current package family manifest. | Authored selector update rule and Make-owned generation posture are explicit. |
| PT-013 | Discover validation ladder and establish baseline | WF-08 | DONE | PT-002 | Section 8 and six retained passing run roots. | Commands are exact or marked TODO with reason. |
| PT-014 | Create current tracker and handoff | WF-08 | DONE | PT-001 through PT-005, PT-012, PT-013 | This file and Section 10. | Another agent can start S-00 without repository rediscovery. |
| PT-015 | Fix audit/revisions artifact disposition | WF-04, WF-05 | DONE | PT-004, PT-005 | RB-003 and the Section 3 browser-disposition table. | Both families are generated/protected/root-omitted/export-omitted/browser-forbidden. |
| PT-016 | Adopt generated evidence/view types internally while preserving public compatibility | WF-05 | DONE | PT-006, PT-007, PT-010 | S-04 Core/OpenAPI/generator projection repair, generated handle alias, operation-specific Evidence consumers, release authorization, and affected owner evidence. | Internal canonical uses are generated-backed and every divergent root declaration remains exact. |
| PT-017 | Add Make-owned browser-reachability evidence for protected artifact families | WF-07 | DONE | PT-006, PT-015 | S-06 private scanner, stable command ID, package verification/row, red and green retained runs, and finalized Make projections. | The catalogued package row proves audit/revisions modules and embedded bytes are absent after `make build-web`. |
| PT-018 | Study alternative family entrypoints only if optimization is separately prioritized | WF-05 | DEFERRED | PT-007, PT-008, PT-011 | Optional S-05 retained measurement report. | No study or a failed materiality gate produces no package API change. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T18:16:17-04:00 | Codex / protocol-ts tracker planning | Scope, authority, and no-implementation boundary are complete on `main` at `8ccf5de48a80ec370764c5840a9e0fc2f22e5912`. | Inspected framework, Domain, Core 00-04, adopted subsystem/harness docs, guide, and live target/support files; touched only this tracker. | Read-only `sed`, `rg`, `git`, `find`, `wc`, `jq`; Make discovery commands. | Target exists, normalized label is safe, tracker was absent, and no owner contradiction was found. | Future implementation needs explicit authorization. | Authorize S-00 characterization only before structural movement. |
| 2026-08-03T18:52:02-04:00 | Codex / protocol-ts NLSpec-style closure | The tracker is revised on `main` at `8ccf5de48a80ec370764c5840a9e0fc2f22e5912`; it remains implementation-support material and no Core amendment is required. | Inspected the current tracker, `temp/analysis-notes.md`, `docs/research/nlspec-spec.md`, live package/configuration evidence, and official supporting documentation; touched only this tracker. | Read-only `sed`, `rg`, `jq`, `git`, Make target discovery, and documentation review. | Normative terms, preservation default, owner placement, and fail-closed contradiction handling are explicit. | No planning blocker; implementation still requires separate authorization. | Authorize S-00 characterization before any structural movement. |
| 2026-08-03T20:32:00-04:00 | Codex / S-07 final handoff | The authorized remediation is complete on the original `main` baseline; S-05 alone remains intentionally deferred. | Audited all changed/untracked paths, final package inventory, public exports, generated policy, retained compatibility, consumer imports, guide wording, `docs/domain.md`, validation evidence, and tracker state. | `make format`; `make agent-finalize` with `RESULTS_DIR` unset; full validation ladder ending in `make check` and `make lint-markdown`. | No public specifier or root symbol was removed; no private facade path was exported; no generated file or lockfile was hand-edited; no owner contradiction was found. | Two unrelated harness exceptions are retained below, each with replacement evidence where applicable. | No restart action is required. A future task should begin from the final worktree diff and this completed tracker. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T18:16:17-04:00 | Codex / protocol-ts tracker planning | No backend package is in target scope; target contains no route handler, SQL/storage adapter, auth decision, mutation coordinator, or projection lifecycle. | Inspected target source and relevant Core/backend owner boundaries; touched only this tracker. | Target dependency/search inventory. | Backend semantics remain with existing modules/platform; protocol-ts exposes wire projections only. | None for S-00/S-01; semantic changes remain out of scope. | Keep future facade slices frontend/package-local. |
| 2026-08-03T18:52:02-04:00 | Codex / protocol-ts NLSpec-style closure | RB-001 through RB-003 require no backend or Core behavior change. | Inspected Core-placement rules in the approved notes and the existing target inventory; touched only this tracker. | Read-only owner/source comparison. | Evidence, revisions, audit, authorization, route, storage, and projection semantics remain with their existing owners. | None for the behavior-preserving package refactor. | Escalate to the applicable Core owner only if S-00 finds a real wire-contract ambiguity or projection defect. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T18:16:17-04:00 | Codex / protocol-ts tracker planning | Legitimate protocol facade with a mixed authored root; no shell/controller, test-util, or grid-vendor behavior found. | Inspected package root/config, direct app/view/ui consumers, frontend import policy, Fallow policy; touched only this tracker. | Import searches; `make frontend-import-boundary-check`; `make frontend-typecheck`. | Both checks passed at the Section 8 run roots; private facade split and protected-import policy are sequenced. | RB-002 blocks public entrypoint migration without evidence. | Start S-00, then split privately while preserving root exports. |
| 2026-08-03T18:52:02-04:00 | Codex / protocol-ts NLSpec-style closure | RB-002 is closed: the root and all four current subpaths are the exhaustive unchanged package surface. RB-003 fixes a closed browser artifact allowlist. | Inspected `package.json`, frontend entrypoint contract, root generated imports, import-boundary policy/tests, TypeScript options, and build surface; touched only this tracker. | `jq`, `rg`, `sed`, and `make help-all` discovery. | Private facade decomposition may proceed; private paths remain non-public and audit/revisions remain browser-forbidden. | PT-017 is an implementation prerequisite, not a planning ambiguity. | Execute S-00, then author the S-06 bundle row before completing S-01. |
| 2026-08-03T19:40:35-04:00 | Codex / S-01 private facade | S-01 is complete; `src/index.ts` is a six-line composition facade and the package export map is unchanged. | Added the six private facade modules; moved validation, collaboration, extension discovery, artifacts, Network Flow, and compatibility declarations without changing their public names or implementations. | `make frontend-typecheck`; five non-shell package rows; `make build-web`; exact S-06 reachability row; source search and `git diff --check`; `make lint-markdown`. | Typecheck passed at `.cartulary/test-results/20260803T233946Z-p3025514`; package rows passed at `.cartulary/test-results/20260803T234001Z-p3026001`; build passed at `.cartulary/test-results/20260803T234009Z-p3027020`; reachability passed at `.cartulary/test-results/20260803T234023Z-p3031321`; Markdown lint passed at `.cartulary/test-results/20260803T234113Z-p3035101`. Exact runtime exports, artifact order/digests, decoder identity, and five specifiers remain characterized. | Current import-boundary policy still names the former root importer and is intentionally repaired in S-02. | Complete the distinct S-06 green accounting checkpoint before S-02. |
| 2026-08-03T19:51:02-04:00 | Codex / S-02 import enforcement | S-02 is complete; generated access is family-owned and deny-by-default. | Updated frontend import rules and checker extension normalization; added fixtures for approved direct imports and public-root, application, wrong-facade, broad-barrel, audit, and revisions bypasses. | `make frontend-import-boundary-check`; `make run-harness-smoke-extended` sequential and parallel; `make json-shape-check`; full package slice; `make build-web`; `make lint-markdown`. | Boundary check passed at `.cartulary/test-results/20260803T234943Z-p3074140`; all new frontend-boundary fixtures passed within `.cartulary/test-results/20260803T234950Z-p3074755`; JSON shape passed at `.cartulary/test-results/20260803T235048Z-p3093403`; package slice passed at `.cartulary/test-results/20260803T235033Z-p3086344`; build passed at `.cartulary/test-results/20260803T235041Z-p3090080`; Markdown lint passed at `.cartulary/test-results/20260803T235151Z-p3094181`. | The aggregate extended smoke target failed independently because `harness-smoke-toolchain-pins` could not find its synthetic `toolchain-drift` target summary; sequential roots `.cartulary/test-results/20260803T234610Z-p3050441` and `.cartulary/test-results/20260803T234701Z-p3056299` reproduce it. | Preserve the unrelated harness failure as a validation exception; begin S-03. |
| 2026-08-03T19:54:35-04:00 | Codex / S-03 account types | S-03 is complete; only `DensityMode` aliases generated Core HTTP, while all six readonly root account declarations remain explicit. | Updated `compatibilityTypes.ts` and migrated `appShellClient.ts` to generated operation request/response aliases with resource types derived from response `data`. | `make frontend-typecheck`; full package slice; exact module.auth app-shell row; two affected web.application rows; application import search; `git diff --check`; `make lint-markdown`. | Typecheck passed at `.cartulary/test-results/20260803T235343Z-p3098013`; package slice passed at `.cartulary/test-results/20260803T235358Z-p3098501`; auth row passed at `.cartulary/test-results/20260803T235412Z-p3103458`; application rows passed at `.cartulary/test-results/20260803T235421Z-p3103972`; Markdown lint passed at `.cartulary/test-results/20260803T235510Z-p3104856`. No application source imports a retained root account compatibility type. | None. | Begin S-04 at owner and generator inputs. |
| 2026-08-03T20:10:40-04:00 | Codex / S-04 Evidence adoption | S-04 is complete; the repaired generated Evidence shapes are canonical internally while every non-equivalent root declaration remains explicit. | Updated the private compatibility facade, compile/runtime characterization, Evidence service/tests, Workbook mutation ports, Evidence E2E compile consumers, and decision record. | `make frontend-typecheck`; full package slice; import-boundary check; `make build-web`; exact browser-reachability row; affected Evidence/Workbook unit and browser rows; `make lint-markdown`. | Typecheck passed at `.cartulary/test-results/20260804T000301Z-p3116394`; package slice at `.cartulary/test-results/20260804T000316Z-p3116860`; boundary check at `.cartulary/test-results/20260804T000433Z-p3128499`; build at `.cartulary/test-results/20260804T000439Z-p3128976`; reachability at `.cartulary/test-results/20260804T000446Z-p3132317`; the broader browser run passed both targeted rows at `.cartulary/test-results/20260804T000633Z-p3141623`; Markdown lint passed at `.cartulary/test-results/20260804T001212Z-p3195197`. | The exact service-backed browser selection has a pre-existing fixture-bootstrap routing defect recorded in the tests/harness log; the broader Make-owned browser target passed the same rows. | Record the S-05 deferred decision without adding entrypoints or measurement infrastructure. |
| 2026-08-03T20:32:00-04:00 | Codex / S-07 cleanup and guide | No transitional implementation debris remains and the development guide describes the live boundary accurately. | Updated the two stale `protocol-ts` guide descriptions; audited all six private facade modules, compatibility imports, broad-barrel/protected-family imports, package export map, and domain diff. | `make format`; source/import searches; `git diff --check`; `git diff -- docs/domain.md`. | No new private forwarding shim is unused; retained public compatibility declarations remain regardless of repository-local use; application/E2E Evidence consumers use operation aliases; `docs/domain.md` has no diff. | None. | Preserve the final inventory and validation roots in this handoff. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T18:16:17-04:00 | Codex / protocol-ts tracker planning | All 19 generated TypeScript files are inventoried and traced to `tools/contractgen` or the protocol generator and typed inputs. | Inspected generated markers/exports, contract indexes/selections, generator source, generated policy; touched only this tracker. | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`. | All passed at the named Section 8 run roots; no generated drift or hand edit was introduced. | RB-001 blocks divergent type convergence; RB-003 blocks audit/revisions exposure changes. | Use owner inputs and Make generation only in a later authorized task. |
| 2026-08-03T18:52:02-04:00 | Codex / protocol-ts NLSpec-style closure | RB-001 is closed with compatibility preservation; RB-003 is closed with generated/protected/browser-omitted disposition. | Inspected handwritten/generated type overlaps, TypeScript compiler settings, generated barrel/root helper selection, and approved notes; touched only this tracker. | Read-only symbol/import/configuration comparison. | The tracker defines the S-00 record interface, variance dispositions, direct approved-family imports, and prohibited escape hatches. | No owner contradiction; a later projection defect must be repaired at its owner input. | Create the S-00 matrix and executable characterization; do not edit generated files. |
| 2026-08-03T20:10:40-04:00 | Codex / S-04 owner projection repair | Core 01, authored OpenAPI, generated TypeScript/runtime validators, and release compatibility now agree. | Clarified REQ-01-244; updated the Evidence header map, Workbook row/cell openness, the 2.0.0 change set, and generic closed-empty normalization; regenerated OpenAPI, Go artifacts/catalog, and protocol projections only through `make generate`. | `make generate`; `make generate-drift`; `make openapi-compatibility-check`; `make generated-artifact-policy-check`; `make json-shape-check`; `make test-slice OWNER=platform.openapi`. | Initial generation retained exact missing-change evidence at `.cartulary/test-results/20260803T235752Z-p3108685`; the first generic rule correctly stopped on a patterned dynamic-map edge at `.cartulary/test-results/20260803T235838Z-p3110081`; final generation passed at `.cartulary/test-results/20260803T235915Z-p3111761`. Drift passed at `.cartulary/test-results/20260804T000411Z-p3124521`; compatibility at `.cartulary/test-results/20260804T000423Z-p3127179`; generated policy at `.cartulary/test-results/20260804T000423Z-p3127160`; JSON shape at `.cartulary/test-results/20260804T000423Z-p3127178`; OpenAPI owner rows at `.cartulary/test-results/20260804T000326Z-p3121812`. | None; no owner contradiction was found. | Record S-05 as deferred, then begin final cleanup and validation. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T18:16:17-04:00 | Codex / protocol-ts tracker planning | Five active package rows cover ten current Vitest titles; the owner has zero service-backed rows. | Inspected package test, package family manifest, verification owner, collaborator rows, and harness doctrine; touched only this tracker. | `make task-guide ROLE=module-author OWNER=package.protocol_ts`; `make explain-test-owner OWNER=package.protocol_ts`; `make test-slice OWNER=package.protocol_ts`. | Owner slice passed at `.cartulary/test-results/20260803T221341Z-p2950064`; missing characterization is explicit. | No current execution blocker. | Add S-00 titles and update the authored family manifest only when authorized. |
| 2026-08-03T18:52:02-04:00 | Codex / protocol-ts NLSpec-style closure | S-00 compile/runtime coverage and the S-06 browser-reachability row now have explicit interfaces and completion gates. | Inspected package family/verification manifests, compiler options, Make target surface, and existing baseline evidence; touched only this tracker. | Read-only manifest/schema/target inspection; prior six product baselines retained; `make lint-markdown`. | Tests remain machine-grounded and Markdown remains non-executable; package-owned service-backed row count remains zero; Markdown lint passed at `.cartulary/test-results/20260803T225833Z-p2974554`. | The bundle-reachability row does not exist yet and must be authored before S-01 completion. | Authorize S-00; add or move authored selectors only when the corresponding executable evidence exists. |
| 2026-08-03T19:30:00-04:00 | Codex / S-00 characterization | S-00 is complete; product runtime and generated files are unchanged. | Added `public-compatibility.compile.ts`, expanded `index.test.ts`, created the 17-row compatibility decision, updated package selectors, and refreshed the generated topology render index through `make generate`. | `make frontend-typecheck`; `make test-slice OWNER=package.protocol_ts`; `make generate`; `make json-shape-check`; `make lint-markdown`. | Typecheck passed at `.cartulary/test-results/20260803T232756Z-p2994372`; package slice passed at `.cartulary/test-results/20260803T232848Z-p2996217`; generation passed at `.cartulary/test-results/20260803T232919Z-p2997999`; JSON shape passed at `.cartulary/test-results/20260803T232932Z-p3000286`; Markdown lint passed at `.cartulary/test-results/20260803T233031Z-p3001133`. The package still owns zero service-backed rows. | None. | Author the S-06 Make-owned browser-reachability row and retain its expected pre-S-01 failure. |
| 2026-08-03T19:35:32-04:00 | Codex / S-06 red gate | The private scanner, command ID, verification, and package shell row are authored; S-06 is intentionally red before S-01. | Added the reachability helper; updated task-surface owner inputs, package verification/test-family inputs, and Make-generated task/topology projections. No facade implementation file changed. | `make author-test-row-id`; `make generate`; `make build-web`; exact reachability row; `make json-shape-check`; `make lint-markdown`. | Generation passed at `.cartulary/test-results/20260803T233442Z-p3010573`; web build passed at `.cartulary/test-results/20260803T233456Z-p3012943`; the exact row failed as expected at `.cartulary/test-results/20260803T233505Z-p3017178`; JSON shape passed at `.cartulary/test-results/20260803T233525Z-p3020693`; Markdown lint passed at `.cartulary/test-results/20260803T233630Z-p3021518`. | Expected product failure: the current root imports the broad generated barrel. | Begin S-01 and make this exact row green. |
| 2026-08-03T19:42:34-04:00 | Codex / S-06 green gate | S-06 is complete; the same owned row that failed before S-01 now passes and all harness projections are current. | Ran finalization with `RESULTS_DIR` unset, inspected the owner catalog, and changed no S-01 product source during this checkpoint. | `make agent-finalize`; `make explain-test-owner OWNER=package.protocol_ts`; full package slice; `make json-shape-check`; `make generated-artifact-policy-check`; `make lint-markdown`. | Finalization passed at `.cartulary/test-results/20260803T234157Z-p3038346`; package slice passed at `.cartulary/test-results/20260803T234215Z-p3041053`; JSON shape passed at `.cartulary/test-results/20260803T234222Z-p3044662`; generated policy passed at `.cartulary/test-results/20260803T234229Z-p3045091`; Markdown lint passed at `.cartulary/test-results/20260803T234315Z-p3045840`. Exactly six package rows resolve: five Vitest and one shell; service-backed rows remain zero. | None. | Begin S-02 generated-import enforcement. |
| 2026-08-03T20:10:40-04:00 | Codex / S-04 affected-owner evidence | S-04 owner and consumer evidence is complete, including service-backed route behavior and both targeted browser scenarios. | Exercised the exact package, OpenAPI, Evidence, and Workbook rows; no harness catalog input changed in this slice. | Package/OpenAPI owner slices; exact Evidence frontend unit and Workbook OpenAPI rows; three exact service-backed Evidence Go rows; broader `make browser-e2e-webserver-backed`. | Evidence frontend unit passed at `.cartulary/test-results/20260804T000455Z-p3135867`; Workbook OpenAPI row at `.cartulary/test-results/20260804T000455Z-p3135866`; all three Evidence Go rows passed within `.cartulary/test-results/20260804T000507Z-p3136883`; all 62 broader browser units passed at `.cartulary/test-results/20260804T000633Z-p3141623`, including the targeted Evidence and Workbook rows. | `make service-backed-test-slice` could not start the targeted Playwright row because browser fixture bootstrap dropped `OWNER`; roots `.cartulary/test-results/20260804T000507Z-p3136883` and `.cartulary/test-results/20260804T000537Z-p3139661` reproduce it. Product evidence passed through the broader canonical target, so this unrelated routing defect is not an S-04 blocker. | Preserve the routing exception for final reporting; begin S-05 only after Markdown lint. |
| 2026-08-03T20:32:00-04:00 | Codex / S-07 final validation | The final narrow and broad validation ladder is complete. | Finalized harness projections, updated the row-count and transport-boundary assertions exposed by the first broad run, and reran their exact owner rows before the complete graph. | Generation drift/policy/shape/compatibility; package/OpenAPI/account/application/Evidence/Workbook owner rows; typecheck/import/build/reachability; service-backed Evidence rows; `make check`; `make lint-markdown`. | Finalization passed at `.cartulary/test-results/20260804T002433Z-p3386631`; final static projection roots are `.cartulary/test-results/20260804T001524Z-p3207991`, `p3208022`, `p3208013`, and `.cartulary/test-results/20260804T001525Z-p3208062`; typecheck passed at `.cartulary/test-results/20260804T001625Z-p3221581`; package/account/application/Evidence rows at `.cartulary/test-results/20260804T001726Z-p3230233`, `p3230243`, `p3230260`, and `p3230273`; Workbook and service-backed Evidence at `.cartulary/test-results/20260804T001739Z-p3235816` and `p3235820`; build and reachability at `.cartulary/test-results/20260804T001641Z-p3222166` and `.cartulary/test-results/20260804T001654Z-p3226425`; `make check` passed 713/713 units at `.cartulary/test-results/20260804T003100Z-p3553215`; final Markdown lint passed at `.cartulary/test-results/20260804T003549Z-p3628608`. | First `make check` root `.cartulary/test-results/20260804T001803Z-p3237982` exposed the two corrected stale assertions and `/tmp` exhaustion. Second root `.cartulary/test-results/20260804T002456Z-p3389213` proved 710/713 units but used an overlong/in-repo temp path. The final short external temp path passed and was removed. | Handoff complete. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T18:16:17-04:00 | Codex / protocol-ts tracker planning | Target performs schema admission and payload-safe failure projection but no authentication, authorization, storage credential, or route-time access decision. | Inspected Core 04, Extensions tolerant-decoder rules, evidence contracts, decoder implementation, and browser transport consumers; touched only this tracker. | Source/contract search and package owner slice. | Unknown extension data remains inert; decoder failures expose schema/path/category rather than raw payload. | S-00 must characterize WebSocket and decoder failure cases before movement. | Preserve fail-closed and no-payload behavior in all facade slices. |
| 2026-08-03T18:52:02-04:00 | Codex / protocol-ts NLSpec-style closure | RB-003 fixes a deny-by-default browser boundary for administrative audit and revisions artifacts. | Inspected Core-placement guidance, audit/revisions generated artifacts, root helper omissions, and browser import/build boundaries; touched only this tracker. | Read-only artifact/import/security review. | No audit/revisions package entrypoint is permitted; conflict-token key material and deployment key-ring state are permanently excluded from frontend artifacts under the current profile. | Future browser exposure requires a separate owner-backed Core 04 and package API authorization. | Preserve payload-safe failures and prove protected-family bundle absence in S-01/S-06. |
| 2026-08-03T19:35:32-04:00 | Codex / S-06 browser evidence | The previously inferred exposure is now exact executable evidence. | Scanned every emitted `.js` and `.map`, including parsed source-map module lists and source content, using signals derived from the protected generated artifact projections. | Exact package reachability row after `make build-web`. | The row found 26 unique matches: both forbidden generated modules and the three protected artifacts' paths, SHA-256 digests, schema identifiers, and embedded JSON signatures. Failure output contains signatures only from repository contracts, not runtime secrets. | S-06 remains red until S-01 removes broad-barrel reachability. | Split the root into direct-family private facades without weakening the scanner or package surface. |
| 2026-08-03T19:40:35-04:00 | Codex / S-01 browser exclusion | The concrete browser exposure is removed without adding an entrypoint or weakening generated protection. | Verified the authored runtime contains no `generated/index` import; `contractArtifacts.ts` names only WS, view-schema, error, extension, and Network Flow artifact families. | Production build plus the unchanged exact reachability row. | All emitted JavaScript chunks and source maps contain zero protected module-list entries, artifact paths, digests, schema identifiers, or embedded JSON signatures. | S-06 accounting still must prove catalog uniqueness and generated projection currency. | Complete the S-06 green gate, then enforce family-specific imports in S-02. |
| 2026-08-03T20:10:40-04:00 | Codex / S-04 validation boundaries | Generated validation now rejects non-string upload headers, admits inert additive row/cell members, and rejects member-bearing handle bodies at compile and runtime owners. | Inspected the Core/OpenAPI trust boundary, regenerated AJV validation, package runtime cases, service-backed route rows, and the unchanged protected-artifact scanner. | Package/OpenAPI/Evidence owner rows; build and exact bundle-reachability row; broader webserver-backed browser evidence. | Valid string header maps and empty maps remain accepted; numeric values fail; forward additive row/cell members pass without interpretation; protected audit/revisions signals remain absent from emitted JavaScript and maps. | None. | Keep these gates in S-07 final validation. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T18:16:17-04:00 | Codex / protocol-ts tracker planning | Planning is complete; implementation rows remain TODO/DEFERRED. | Inspected all evidence named above; touched only this tracker. | Six passing baseline Make checks plus `make lint-markdown` and read-only discovery. | Baselines and Markdown lint passed; three deferred decisions are preserved: divergent type convergence, public entrypoint/bundle strategy, and audit/revisions artifact exposure. | RB-001 through RB-003 block only their named later slices. | Next session: authorize and execute S-00 characterization only, then update this handoff. |
| 2026-08-03T18:52:02-04:00 | Codex / protocol-ts NLSpec-style closure | All three former blockers are closed; required implementation remains TODO and optional entrypoint research remains DEFERRED. | Inspected all revision inputs named above; touched only this tracker. | Read-only repository/source review; `make lint-markdown`; final one-file audit. | RB-001 preserves divergent public types, RB-002 keeps five specifiers, RB-003 protects and browser-omits audit/revisions, and Markdown lint passed. | No open planning blocker. PT-017 is a required evidence task; S-05 is optional and defaults to no change. | Authorize S-00 characterization before S-01 or any compatibility alias. |
| 2026-08-03T19:20:07-04:00 | Codex / protocol-ts remediation activation | Tracker activation is complete on `main` at `8ccf5de48a80ec370764c5840a9e0fc2f22e5912`; no implementation file changed at this checkpoint. | Rechecked repository status, current commit, tracker scope, slice dependencies, generated-file restrictions, and command routing; touched only this tracker. | Read-only `git`, `date`, `sed`, and `rg`; `make lint-markdown`. | Authorization, fixed public-interface decisions, exact execution order, and checkpoint protocol are recorded; Markdown lint passed at `.cartulary/test-results/20260803T232040Z-p2987795`. | None. | Begin S-00 characterization and compatibility decision record. |
| 2026-08-03T19:30:00-04:00 | Codex / S-00 checkpoint | The characterization baseline is executable and complete. | Inspected all candidate generated/root declarations, WebSocket projection inputs, extension discovery rules, artifact exports, and package selectors; touched only S-00 artifacts plus this tracker. | Final passing S-00 commands are recorded in the tests/harness log; earlier calibration failures were resolved within the slice. | Fixed dispositions are executable: `DensityMode=alias`, `EvidenceHandleIssueRequest=owner_projection_defect`, and the other fifteen candidates remain compatibility declarations. | S-06 intentionally begins red because the current broad generated barrel reaches forbidden browser artifacts. | Complete tracker Markdown lint, then begin the S-06 red gate without facade implementation edits. |
| 2026-08-03T19:35:32-04:00 | Codex / S-06 red checkpoint | The new evidence is correctly catalogued and fails for the intended product reason, not a harness error. | Inspected generated audit/revisions modules, current built chunks/maps, task-surface schemas, verification profiles, row identity, and generated projections. | Retained commands and run roots are recorded above. | The row resolves through one stable command ID, scans all 14 JavaScript chunks and 14 source maps in the current build, and reports the prohibited bytes precisely. | Expected red gate only. | Pass Markdown lint, then begin S-01; do not mark PT-017 or S-06 complete until green. |
| 2026-08-03T19:40:35-04:00 | Codex / S-01 checkpoint | Structural decomposition and prohibited-bundle removal are complete and independently revertible. | Touched only authored package facade source plus this tracker in S-01; S-00/S-06 evidence remained unchanged. | Final S-01 runs are recorded above. | Public value exports and type imports remain exact; browser exposure changed only by removing prohibited administrative contract bytes. | None for S-01. | Pass Markdown lint, then execute S-06 green accounting and finalization. |
| 2026-08-03T19:42:34-04:00 | Codex / S-06 green checkpoint | The S-06 red/green evidence and catalog accounting are complete. | Verified generated task surface and topology currency, exact owner resolution, runner/evidence classification, and zero service-backed package rows. | Final passing commands and run roots are recorded in the tests/harness log. | PT-017 is DONE; the row is durable package-owned security evidence and no Markdown participates in execution. | None. | Pass Markdown lint, then begin S-02. |
| 2026-08-03T19:51:02-04:00 | Codex / S-02 checkpoint | Least-privilege generated imports are enforced without weakening facade access. | Inspected checker resolution semantics, normalized JavaScript/TypeScript module extensions structurally, and verified the live package imports against family-specific rules. | Retained runs are recorded above; the relevant extended-smoke evidence target passed even though the aggregate stopped on an unrelated target. | Public root and application code cannot import generated paths; wrong facades cannot import another family's projection; broad barrel, audit, and revisions have unconditional deny rules. | Unrelated toolchain-pin smoke defect is retained for final skipped/failed-check reporting. | Pass Markdown lint, then begin S-03. |
| 2026-08-03T19:54:35-04:00 | Codex / S-03 checkpoint | Account/profile deduplication is complete without expanding caller write capability. | Compared operation aliases to the S-00 matrix, retained readonly root declarations, and verified application consumers use generated operation types. | Passing narrow runs are recorded above. | Root names and shapes are stable; request bodies now receive operation-specific compile checking. | None. | Pass Markdown lint, then begin S-04. |
| 2026-08-03T20:10:40-04:00 | Codex / S-04 checkpoint | Owner/projection repair and canonical Evidence adoption are complete; PT-016 is DONE. | Verified owner text, authored OpenAPI, release authorization, generator normalization, generated projections, retained root compatibility declarations, live consumers, and targeted owner evidence. | Retained passing and calibration roots are recorded in the frontend, contract/codegen, and tests/harness logs. | No application or E2E source imports the retained root Evidence/blob compatibility types; only `DensityMode` and `EvidenceHandleIssueRequest` alias generated declarations; `docs/domain.md` remains unchanged. | Preserve two unrelated harness exceptions for final reporting: extended-smoke toolchain summary and targeted service-backed browser `OWNER` propagation. | Run the tracker Markdown checkpoint, then record S-05 as explicitly deferred. |
| 2026-08-03T20:14:00-04:00 | Codex / S-05 deferred decision | S-05 remains `DEFERRED`; no package entrypoint, export, implementation, or measurement infrastructure changed. | Rechecked the post-S-01 production result, fixed five-specifier surface, quantified materiality gates, reproducibility requirements, and SCC constraints; touched only this tracker in S-05. | Read-only tracker/package review; `make lint-markdown`. | The concrete security defect is already removed. Possible residual bundle inefficiency does not justify new public API or permanent tooling without two reproducible builds per variant and the 10 KiB/3 percent or cycle-removal gate. Markdown lint passed at `.cartulary/test-results/20260804T001330Z-p3198685`. | None; deferral is explicitly not a dependency of S-07. | Begin S-07 cleanup and final validation. |
| 2026-08-03T20:32:00-04:00 | Codex / final restart state | No remediation work remains open; all required slices are complete and PT-018 remains deliberately deferred. | Reconciled tracker statuses, final inventory, guide/domain decisions, actual run roots, failures, replacements, and binary criteria. | Final status/diff audit; `make lint-markdown`. | Required product and security evidence is green, including 713/713 `make check` units and both targeted browser rows in the retained webserver-backed run. | Known external exceptions only: `run-harness-smoke-extended` still has its previously recorded synthetic toolchain-summary defect, and exact owner-selected browser fixture bootstrap drops `OWNER`; neither affects the passing product rows or final `make check`. | Handoff complete. If work resumes, treat it as a new task; do not reopen S-05 without its quantified materiality gate. |

## 11. Resolved Questions and Decisions

No open decision blocks the authorized remediation. Each slice retains its own
evidence prerequisites and may begin only after the preceding tracker and
Markdown checkpoint is complete.

| ID | Binding decision | Required evidence or implementation | Status |
| --- | --- | --- | --- |
| RB-001 | Only declarations proven equivalent in caller acceptance, output compatibility, write capability, and runtime schema behavior **MAY** alias generated Core HTTP types. Every evidence/view declaration still affected by optionality, readonlyness, index signatures, or nested-shape differences after owner repair **MUST** remain an exact root compatibility declaration. Internal protocol consumers **MUST** use canonical generated operation types through supported package exports. | S-00 symbol-by-symbol decision matrix; positive and negative compile fixtures under the actual TypeScript configuration; S-04 owner/projection repair; runtime validator parity; affected Evidence/Workbook owner evidence. `DensityMode` and the repaired closed-empty `EvidenceHandleIssueRequest` are the only aliases. | CLOSED |
| RB-002 | `@cartulary/protocol-ts` and the existing `./collaboration`, `./core-http`, `./http`, and `./network-flow` subpaths **MUST** remain. The current refactor **MUST NOT** add, migrate, or remove a public entrypoint. Private facade paths **MUST NOT** be exported. | Five-entrypoint compile smoke, exact export preservation, import-boundary checks, and build evidence. Optional S-05 requires every compatibility gate and one quantified materiality gate; otherwise no API change occurs. | CLOSED — KEEP |
| RB-003 | Audit and revisions artifact families **MUST** remain generated and protected, **MUST NOT** be enumerated by root artifact helpers, **MUST NOT** gain package entrypoints, and **MUST** be unreachable from initial and dynamically reachable production browser chunks. Browser code **MUST** use existing public audit/history/conflict HTTP types. Conflict-token key material, secret references, resolved secrets, and deployment key-ring instances **MUST NOT** enter frontend artifacts or bundles. | Direct approved-family imports from the private artifact facade; negative broad-barrel/audit/revisions import fixtures; production bundle absence assertion; generated drift/policy checks. Any future browser-safe projection requires a separate owner-backed package API and Core 04 security authorization. | CLOSED — PROTECTED AND BROWSER-OMITTED |

## 12. Binary Completion Criteria

| Criterion | Result | Evidence |
| --- | --- | --- |
| Every file in `packages/protocol-ts` is inventoried or explicitly out of scope. | PASS | Section 2 inventories the 24-file activation baseline and all seven final authored additions, yielding 31 target files once adopted; ignored install/build caches are excluded. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 maps package, HTTP, WS, workbook, evidence, account, imports, revisions, Network Flow, generated, and harness surfaces. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 defines WF-00 through WF-08 with predecessors, successors, validation, and checkpoints. |
| Normative terms, authority, and the default behavior are unambiguous. | PASS | Section 1 defines `MUST`, `MUST NOT`, `SHOULD`, and `MAY`, subordinates the tracker to adopted owners, and makes unchanged behavior the default. |
| The supported package interface is exhaustive and frozen. | PASS | Section 4 names exactly the root and four current subpaths and prohibits additions, removals, renames, migrations, or private-facade exports. |
| Every compatibility variance has a deterministic disposition. | PASS | Section 4 maps equivalence, optionality, readonlyness, index signatures, header values, empty objects, and prohibited escape hatches. |
| Every affected compatibility symbol has a complete future decision-record interface and executable posture. | PASS | Section 7 defines all required S-00 fields and compile/runtime cases; Markdown remains non-executable. |
| S-04 preserves root compatibility while repairing owner projections. | PASS | Internal consumers adopt generated operation types; every non-equivalent root declaration remains exact; `DensityMode` and the repaired closed-empty handle request are the only aliases; Core HTTP intentionally widens row/cell additive tolerance and rejects invalid headers/handle bodies. |
| The public-entrypoint default and optional optimization threshold are explicit. | PASS | RB-002 closes on keep; S-05 defaults to no change and defines compatibility, reproducibility, bundle, regression, and cycle gates. |
| Audit and revisions disposition is complete and security-bounded. | PASS | Sections 3, 5, 7, and 11 require generated/protected/root-omitted/export-omitted/browser-unreachable posture and prohibit frontend key material. |
| Every required implementation slice follows the authorized compatibility posture. | PASS | S-00 through S-04, S-06, and S-07 preserve all five specifiers and root symbols; only S-04's owner-backed projection repairs change generated subpath acceptance. S-05 cannot authorize an API change. |
| Validation commands and retained evidence are complete. | PASS | Section 8 and Section 10 record narrow owner, service-backed, browser, static, generation, build, security, and final 713/713 `make check` evidence. The package correctly owns zero service-backed rows. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No contradiction was found; Section 5 defines the required fail-closed handling if one is later discovered. |
| All three prior refactor blockers have binding closure decisions. | PASS | Section 11 records RB-001 through RB-003 as closed; remaining TODO rows are implementation prerequisites, not design ambiguity. |
| The work tracker uses only allowed statuses and separates decisions from implementation. | PASS | Every required PT row is `DONE`; PT-018 alone is `DEFERRED` under the explicit S-05 materiality gate. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Section 5 records guide shorthand versus the live authored-plus-generated package. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Section 10 preserves activation through final validation, actual run roots, known exceptions, and a closed restart state. |
| Tracker activation changes only the tracker before implementation begins. | PASS | Activation `git status` found only this untracked controlling artifact; implementation starts after its Markdown checkpoint. |
| Repository-wide validation passes. | PASS | `make check` passed all 713 units at `.cartulary/test-results/20260804T003100Z-p3553215` using a short external temp path; both temporary directories created by S-07 were removed afterward. |
| Markdown validation passes for the completed tracker. | PASS | `make lint-markdown` passed at `.cartulary/test-results/20260804T003549Z-p3628608`. |

This tracker is complete as the controlling remediation artifact. All required
ordered slices satisfy their exit criteria and checkpoint protocol; S-05
remains intentionally deferred and authorizes no package API change.

## 13. Iteration 2 Scope, Authority, and Decisions

### 13.1 Goal and authorized future scope

Iteration 2 removes repository-unreachable compatibility and aggregation
surfaces, replaces raw generated package entrypoints with cohesive authored
family entrypoints, minimizes browser runtime projections, and establishes an
enforced authored dead-code gate. The package remains private and versioned
`0.0.0`. Repository reachability is the complete compatibility boundary for
this iteration; no external-consumer or semantic-version deprecation cycle is
required.

The future implementation MAY change the package export map, authored package
source, repository consumers, contract-family generation metadata and schema,
the protocol generator, generated projections through `make generate`, import
and reachability policy, package-owned tests, affected owner tests, the
development guide, the prior compatibility decision record, and this tracker.
It MUST NOT change adopted wire behavior, domain vocabulary, authentication or
authorization behavior, unrelated application behavior, generated outputs by
hand, lockfiles by hand, or `docs/domain.md`.

Execution order is mandatory:

`tracker amendment -> I2-00 -> I2-01A -> I2-02 -> I2-03 -> I2-04 -> I2-05 -> I2-06 -> I2-01B -> I2-07`

After every workstream, the implementer **MUST** update Sections 17 through 19
with changed files, commands, actual run roots, failures, replacements,
blockers, and the next workstream, then pass `make lint-markdown`. The next
workstream **MUST NOT** begin before that checkpoint passes. Tests, generators,
runtime metadata, and conformance evidence **MUST NOT** read, stat, hash, or
parse this or any other Markdown file.

### 13.2 Superseded and retained Iteration 1 decisions

| Prior decision | Iteration 2 disposition | Binding reason |
| --- | --- | --- |
| RB-001 retained fifteen handwritten compatibility declarations and two aliases. | **SUPERSEDED.** Remove all handwritten compatibility declarations and aliases after repository consumers use generated operation-specific types. | The package is private, all fifteen handwritten declarations have no repository consumer outside package characterization, and internal consumers already use canonical generated operation types. Retention would create permanent type drift and write-capability ambiguity without continuing value. |
| RB-002 retained the root plus four subpaths and made entrypoint change conditional on S-05. | **SUPERSEDED.** The final API contains exactly seven authored family subpaths and no root or `./core-http`. | The new iteration explicitly authorizes the structural change. Family cohesion, protected generated projections, and the elimination of eager aggregate reachability are primary design requirements rather than an optional size optimization. |
| RB-003 kept audit and revisions TypeScript artifact modules generated but protected and browser-omitted. | **STRENGTHENED.** Retain their backend/Go generation and the browser-forbidden security rule, but stop generating their unused TypeScript runtime projections. | A protected unused browser-language projection provides no continuing value and preserves a re-exposure hazard. The bundle scanner remains owner-input-driven after the modules are deleted. |
| S-05 required a materiality gate before an entrypoint study. | **SUPERSEDED FOR ITERATION 2.** Retain before/after production-build evidence, but do not condition the authorized family boundary on a byte threshold. | The API change is authorized for cohesion, dependency direction, reachability, and maintainability. Bundle size is validation evidence, not the design owner. |

No adopted Core or subsystem owner requires the retired package aliases,
specifier layout, generated barrel, or generic artifact helpers. If two adopted
owners contradict during implementation, the affected slice **MUST** stop and
be recorded as `BLOCKED: owner contradiction`.

### 13.3 Fixed compatibility and migration policy

- The final package remains private and at version `0.0.0`.
- Removal is immediate inside one repository migration; no deprecated aliases,
  forwarding entrypoints, wildcard compatibility barrels, or removal shims
  survive I2-05.
- Repository consumers use operation-specific HTTP request/response aliases and
  indexed access into those responses for resource types.
- Generated Core HTTP types remain protected implementation projections. The
  authored HTTP entrypoint explicitly exposes only `ErrorEnvelope` and
  `SheetRef` in addition to generated operation bindings and decoder
  primitives.
- Owner-declared generated type pass-throughs are not dead merely because an
  individual selected protocol type lacks a current caller. Authored wrappers,
  helpers, types, and values receive no such exemption.
- Unexpected removal candidates outside the inventory in Section 14 do not
  authorize opportunistic deletion. Record them and resolve their owner before
  expanding a slice.

## 14. Iteration 2 Baseline Findings and Removal Manifest

Read-only discovery on 2026-08-03 used the clean `d29322b6` worktree. It found
the following concrete seams.

| Finding | Evidence | Final disposition | Unresolved risk |
| --- | --- | --- | --- |
| Fifteen explicit handwritten account, view, evidence, and blob compatibility declarations have no repository import outside package compile/runtime characterization. | Import search excluding generated sources, the compatibility file, the package test, and compile fixture. Same-named application/E2E local types are not package imports. | Delete the declarations and the two generated aliases with the root compatibility facade in I2-05. | Continued divergence in readonlyness, optionality, openness, and operation ownership. |
| Root imports dominate live consumers even though collaboration, HTTP, Network Flow, import targets, view schemas, extensions, and errors have distinct owners. | Direct package-specifier search across `apps` and `packages`. Only `./core-http` has live subpath consumers; the other current family subpaths are bypassed through root. | Migrate every consumer to one owning authored family subpath in I2-04, then remove root and `./core-http`. | Aggregate coupling, accidental cross-family initialization, and unclear API ownership. |
| The custom extension-discovery decoder has no production consumer. | Its only external references are package tests; application transport already calls generated `validateHTTPOperationResponse` for `listDeploymentExtensions`. | Move its owner cases to generated operation validation and delete the custom decoder in I2-05. | Two validation paths can drift and disagree on tolerated additive members or invalid known members. |
| Most generic artifact enumeration, lookup, require, profile, reason-code, and registry helpers are package-test-only. The generic parser has only narrow error, extension, and Network Flow uses. | Symbol-by-symbol repository search of `contractArtifacts.ts` and its callers. | Replace live uses with typed generated registry values and a view-schema-specific artifact surface; delete the generic API. | Arbitrary artifact-path coupling, eager family aggregation, unchecked casts, and a growing public helper burden. |
| The protected generated barrel and audit/revisions TypeScript artifact modules have no supported consumer. | Generated import search plus package export map and enforced browser scanner. | Delete their projections in I2-06 while retaining backend generation and scanner signatures. | A future import can reintroduce administrative data into browser output. |
| Additional generated artifact arrays are broader than runtime need. | Production source maps contain `view-schemas-artifacts.ts` (about 310 KB of source), `network-flow-artifacts.ts` (about 104 KB), `errors-artifacts.ts` (about 55 KB), `extensions-artifacts.ts` (about 20 KB), and unused `ws-artifacts.ts` (about 6 KB). | Keep only view-schema source artifacts; replace fixed registries with typed JSON-value projections and remove the other arrays. | Unnecessary browser bytes and artifact-family initialization. |
| The public `frontend-fallow-static` route currently fails before analysis. | `make frontend-fallow-static` failed with `frontend-fallow-static has no matching registered cache profile`. | Register it in the frontend-analysis cache profile in I2-01. | The repository advertises a diagnostic that cannot execute through its supported surface. |
| Fallow exempts every Protocol-TS root export and treats the package as externally public. | `.fallowrc.json` lists the root as an entry, ignores all its exports, and includes the package in `publicPackages`. | Remove blanket exemptions and add a package-scoped enforced gate with explicit compile/generated handling. | Dead authored API cannot be detected or prevented from returning. |

### 14.1 Exact legacy surfaces to remove

I2-05 and I2-06 remove these known surfaces after their replacements are live:

- root `@cartulary/protocol-ts` and `@cartulary/protocol-ts/core-http`;
- `src/index.ts`, `src/facade/compatibilityTypes.ts`, the generic contract
  artifact facade, custom extension discovery, obsolete root compile matrix,
  monolithic root characterization, and private forwarding modules superseded
  by final entrypoints;
- root compatibility declarations and aliases for view, account, evidence, and
  blob shapes;
- generic artifact enumeration, arbitrary path get/require/parse, extension
  profile lookup, reason-code lookup, and registry lookup helpers;
- `generated/index.ts`, `audit-artifacts.ts`, `revisions-artifacts.ts`,
  `ws-artifacts.ts`, `errors-artifacts.ts`, `extensions-artifacts.ts`, and
  `network-flow-artifacts.ts`; and
- the application extension-contract forwarding shim once its consumers use
  the typed Extensions entrypoint.

Historical Markdown remains audit evidence. The Iteration 1 compatibility
decision record is marked superseded rather than deleted and remains
non-executable.

### 14.2 I2-00 exact live-consumer replacement map

The current export map has exactly five supported specifiers: root,
`./collaboration`, `./core-http`, `./http`, and `./network-flow`. The future
`./errors`, `./extensions`, `./import-targets`, and `./view-schemas` specifiers
are absent. Current root and `./core-http` consumers migrate as follows; this
list excludes package-only characterization fixtures.

- `./http`: `apps/web/src/app/api/publicHttpTypes.ts` moves close/reopen and
  administrative-audit aliases; `apps/web/src/app/api/appShellClient.ts` moves
  session, account, user, and extension-discovery aliases and derives its
  extension resource from `ListDeploymentExtensionsResponse`;
  `apps/web/src/services/browserApi.ts` moves operation IDs, metadata, path and
  query builders, query values, and response validation;
  `apps/web/src/services/httpTransport.ts` moves `Decoder`;
  `apps/web/src/services/workbookEvidence.ts` moves object-blob aliases and
  `ErrorEnvelope`; `apps/web/e2e/evidence-integration.spec.ts` moves its six
  object-blob/evidence operation aliases; workbook mutation, view-query,
  pending-mutation, history, mention, clipboard, and evidence adapters under
  `apps/web/src/workbook/**` move their operation aliases and decoder
  primitives; `apps/web/src/services/importContractAdapter.ts` moves every
  imported import/job alias plus `ImportSourceColumnMapping`; and
  `apps/web/e2e/workbook.spec.ts`,
  `apps/web/e2e/support/workbook/savedViews.ts`, and
  `apps/web/src/shared/sheetRef.ts` move `SheetRef` from `./core-http`.
- `./collaboration`:
  `apps/web/src/collaboration/IncidentCollaborationSession.tsx` moves
  `IncidentStreamMessage` and `incidentStreamMessageDecoder`.
- `./network-flow`: `apps/web/src/services/networkFlowContractAdapter.ts`
  moves every imported Network Flow type, `DecodeFailure`, `Decoder`, the
  descriptor, mapping/presentation registries, typed error registry, and all
  decoders; `apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx` moves
  `networkFlowDecoders`.
- `./import-targets`: `apps/web/src/services/importTargetContractAdapter.ts`
  moves `ImportTargetFrontendRow` and `importTargetRegistry` without moving its
  interpretation logic.
- `./view-schemas`: `packages/view-contracts/src/index.ts` moves all five
  view-schema source types, the registry value/accessor, and source artifacts;
  `packages/ui-contracts/src/viewSchemaSelectors.ts` moves the registry-entry
  accessor.
- `./errors`: `apps/web/e2e/incident.support.spec.ts` replaces generic parsing
  of `contracts/errors/index.json` with the typed immutable error registry.
- `./extensions`: `apps/web/e2e/incident.support.spec.ts` replaces generic
  profile-registry parsing with the typed immutable profile registry. The
  forwarding `apps/web/src/services/extensionContractAdapter.ts` is deleted;
  its packaged-registry responsibility moves here and its discovery-resource
  type moves to `./http`.

No live consumer needs a handwritten compatibility declaration, root-only
generic artifact helper, custom extension decoder, generated barrel, or direct
generated/private path.

## 15. Final Package Interface and Ownership

The final `package.json` export map contains exactly these seven specifiers and
points only to authored files under `src/entrypoints`:

| Specifier | Required authored surface | Generated dependencies permitted behind it |
| --- | --- | --- |
| `@cartulary/protocol-ts/collaboration` | `IncidentStreamMessage` and the frozen, payload-safe incident-stream decoder. | Collaboration types and protocol validator only. |
| `@cartulary/protocol-ts/errors` | Typed immutable `errorRegistry`. | Exact public error-registry JSON-value projection only. |
| `@cartulary/protocol-ts/extensions` | Typed immutable `extensionClientSupportRegistry` and `extensionProfileRegistry`. | The two exact Extensions JSON-value projections only. |
| `@cartulary/protocol-ts/http` | HTTP operation metadata, path/query builders, response validation, operation request/response types, `Decoder`, decode result/failure types, `createGeneratedDecoder`, `ErrorEnvelope`, and `SheetRef`. | HTTP operation bindings, protected Core HTTP types for the two explicit component types, and protocol validators. |
| `@cartulary/protocol-ts/import-targets` | Generated import-target registry and row types without added interpretation. | Import-target registry projection only. |
| `@cartulary/protocol-ts/network-flow` | Generated public Network Flow types, descriptor, mapping and presentation registries, typed error registry, and frozen decoders. | Exact Network Flow type/descriptor/registry/validator projections only. |
| `@cartulary/protocol-ts/view-schemas` | View-schema source types, immutable source artifacts, typed registry value, registry entry types, and narrowly named view-schema accessors only when a live consumer requires them. | View-schema source types, selected source artifact collection, and exact registry JSON-value projection. |

No `.` export exists. `./core-http`, `./generated`, `./facade`, `./entrypoints`,
and `./internal` are unsupported. Application and package import-boundary policy
must reject private and generated paths even when filesystem resolution could
find them.

`contracts/protocol-ts/frontend-entrypoints.v2.json` becomes the machine owner
of the exact seven specifiers and their authored source paths. A shape or
package-owned check compares it exactly with `package.json`; neither surface
may drift independently.

## 16. Iteration 2 Workstream Plan

### I2-00 — Characterization and removal manifest

**Areas:** tests, retained build evidence, and this tracker.

**Remediation:** Freeze the current five supported specifiers, record the four
future family specifiers as currently absent, map every live import and symbol
to its final family, characterize runtime export identity, decoder failure
redaction, registry identities, and two independent pre-change production
module graphs. Add passing characterization for the custom extension decoder's
adopted owner behavior and the discrepancy between that decoder and generated
HTTP response validation. Do not compile future entrypoints or reject the
still-supported root and `./core-http` in this slice.

**Rationale and benefit:** Removal begins from executable behavior and exact
consumer ownership rather than names or apparent non-use. The retained
evidence distinguishes intended removals from protocol regressions.

**Compatibility impact:** Test-only. No export or runtime change occurs.

**Risk if unresolved:** A later slice could silently remove required protocol
validation, preserve a dead alias, or invent a replacement API during
migration.

**Exit:** Current package compile/runtime coverage, package slice, frontend
typecheck, import-boundary check, two independent `make build-web` runs, the
current security scanner, and Markdown checkpoint pass. Section 14 has one
deterministic disposition and replacement for every current authored export;
the checkpoint records `domain vocabulary unchanged`.

### I2-01A — Enforced dead-code red gate

**Areas:** harness owner inputs/projections, Fallow configuration, tests, and
package verification accounting. No package implementation changes occur.

**Remediation:** Add `frontend-fallow-static` to the frontend-analysis cache
profile. Add a Make-owned `protocol-ts-dead-code-check` backed by package-scoped
Fallow analysis with entry-export reporting and failure on findings. Remove the
Protocol-TS blanket entry/export/public-package exemptions, register compile
fixtures explicitly, and exempt only generated roots and pure generated
pass-throughs declared by the entrypoint owner. Add fixtures proving unused
authored files, types, values, and exports fail, while live imports, compile
fixtures, and owner-declared generated pass-throughs pass. Per-file or
per-symbol authored suppressions are forbidden.

**Rationale and benefit:** Repository reachability becomes executable policy,
and future dead compatibility helpers cannot accumulate silently.

**Compatibility impact:** Internal verification only.

**Risk if unresolved:** The cleanup is one-time and regression-prone, while the
advertised static target remains broken.

**Exit:** Fixture tests and harness smoke pass; `make frontend-fallow-static`
executes through the supported cache profile; the package gate returns nonzero
with the expected legacy-authored findings and retained evidence. I2-01A is
complete at that red state. The Markdown checkpoint records
`domain vocabulary unchanged` before I2-02 begins.

### I2-02 — Durable projections, validators, and tolerance repair

**Areas:** contract-family registry/schema, frontend-entrypoint and HTTP
operation machine owners, protocol/contract generators, generator and
conformance tests, generated outputs through Make, and policy accounting.

**Remediation:** Replace contract-family registry v3 with v4. Replace
`typescript_runtime_artifact_prefixes` with validated projection records for
an `artifact_collection` selected by exact prefixes and a `json_value` selected
by one exact artifact path. Support an optional generated barrel during the
migration; absence deletes the prior barrel. Validate repository containment,
exact artifact matches, unique output paths and identifiers, deterministic
ordering, declared generated outputs, and deletion of removed projections. Add
deeply readonly stable-singleton values for public errors, Extensions client
support, Extensions profiles, Network Flow errors, and the view-schema
registry. Retain legacy collections and the barrel until I2-06 so the current
root stays buildable; only the view-schema collection survives finally.

Adopt `frontend-entrypoints.v2` as the target-state owner of the exact seven
specifier/path pairs and per-entrypoint generated-module allowlists while
preserving current HTTP, collaboration, and Network Flow selections. Adopt
`http-operations.v2` with executable response-validation policy identifiers
and a closed-schema default. Generate separate Core HTTP, collaboration, and
Network Flow validator modules, retaining a compatibility validator barrel
only until I2-06. Generated HTTP validation implements the adopted owner
policies: extension discovery tolerates inert additive item members while
enforcing known-member types, bounds, sorted and unique collections, and
fail-closed complete-response behavior; administrative audit tolerates only
owner-declared future vocabulary while retaining closed structure and
redaction. Every validation failure remains payload-free.

**Rationale and benefit:** Browser projections and tolerance become executable
owner data rather than a broad barrel, descriptive label, or handwritten
one-off parser. Future families can grow without widening unrelated graphs.

**Compatibility impact:** Additive generated outputs and owner-aligned
conformance repair only; the legacy root continues to compile.

**Risk if unresolved:** Family entrypoints initialize unrelated validators,
custom and generated discovery validation disagree, and typed registries
require unchecked parsing.

**Exit:** Generator positive/negative tests, `make generate`,
`make generate-drift`, generated-artifact policy, JSON shape,
`make test-fast`, package slice, Extensions owner slice, and Markdown
checkpoint pass; the legacy facade still compiles and the checkpoint records
`domain vocabulary unchanged`.

### I2-03 — Authored family entrypoints

**Areas:** package implementation, package export map, import policy, compile
fixtures, runtime tests, and package catalog rows.

**Remediation:** Introduce the seven modules in Section 15 under
`src/entrypoints` and move shared decoder mechanics to `src/internal`. Repoint
the existing collaboration, HTTP, and Network Flow specifiers to authored
modules; add errors, extensions, import-targets, and view-schemas. Retain root
and `./core-http` temporarily for I2-04 migration. Each entrypoint imports only
generated modules allowed by `frontend-entrypoints.v2`; decoder construction
receives exact family validators instead of dynamically indexing a namespace.

**Rationale and benefit:** Every public import maps to one cohesive owner and
generated implementation files remain protected. Future phase growth extends
one family without widening an aggregate root.

**Compatibility impact:** Four additive specifiers plus
implementation-preserving repointing of three current subpaths. The temporary
export count is nine.

**Risk if unresolved:** Consumer migration would target raw generated files or
another temporary aggregate.

**Exit:** Seven positive compile fixtures become active; private/generated
paths remain negative; legacy root fixtures remain positive; family runtime
tests, package slice, frontend typecheck, import-boundary check, web build,
browser scanner, and Markdown checkpoint pass. The checkpoint records
`domain vocabulary unchanged`.

### I2-04 — Repository consumer migration

**Areas:** application and package imports/types, affected tests, and the
obsolete application forwarding shim.

**Remediation:** Move HTTP and `SheetRef` consumers to `./http`, using
operation-specific request/response aliases and indexed response types for
resources. Move collaboration, Network Flow, import targets, view schemas,
Extensions registries, and error registries to their owning entrypoints.
Replace arbitrary artifact parsing with typed registry access. Replace
extension discovery resource aliases with indexed
`ListDeploymentExtensionsResponse` types. Delete the extension-contract
forwarding shim after its responsibilities move to `./extensions` and
`./http`.

**Rationale and benefit:** Callers express their actual dependency and no
longer initialize unrelated runtime families through the root.

**Compatibility impact:** Repository-local import and type changes only; wire
requests, responses, validation, registry data, and runtime identity remain
owner-compatible.

**Risk if unresolved:** Root removal is unsafe and family isolation cannot be
established.

**Exit:** Search finds no live root or `./core-http` imports outside
characterization fixtures. Package, view-contracts, UI-contracts, Extensions,
Imports, Network Flow, collaboration, application, Workbook, and Evidence
owner slices pass, followed by typecheck, import-boundary, build, scanner, and
Markdown checkpoint. The checkpoint records `domain vocabulary unchanged`.

### I2-05 — Remove legacy authored APIs

**Areas:** package metadata/source/tests, entrypoint owner/schema, import
policy, decision documentation, and harness selectors.

**Remediation:** Remove root and `./core-http`; delete handwritten compatibility
declarations and aliases, generic artifact enumeration/parsing/lookups, the
custom extension decoder, obsolete root characterization, and superseded
private facade modules. Enable exact comparison between `package.json` and
`frontend-entrypoints.v2`. Mark the Iteration 1 compatibility decision record
superseded without deleting historical evidence.

**Rationale and benefit:** No indefinite transition layer or duplicate type
authority survives. The final API is small, explicit, and mechanically owned.

**Compatibility impact:** Intentional breaking change limited to the private
workspace package after all repository consumers migrate; no bridge remains.

**Risk if unresolved:** The aggregate facade and compatibility burden become
permanent.

**Exit:** Exactly seven authored exports match the v2 owner; root,
`./core-http`, private, internal, and generated imports fail compile; no
compatibility declaration, generic parser, or duplicate discovery path
remains; affected tests, build, scanner, and Markdown checkpoint pass. The
checkpoint records `domain vocabulary unchanged`.

### I2-06 — Remove obsolete generated projections and harden policy

**Areas:** v4 contract-family projections, generated outputs through Make,
generator/import-policy tests, browser scanner, and generated accounting.

**Remediation:** Remove transitional artifact and validator barrels and
obsolete artifact collections from their v4 owners, regenerate, and delete the
seven artifact modules named in Section 14 plus the monolithic validator
module. Retain backend/Go audit and revisions generation. Derive forbidden
browser paths, digests, identifiers, and content signatures directly from the
audit/revisions contract-family owner inputs rather than deleted TypeScript
files. Deny reintroduction of a generated barrel, monolithic validator barrel,
or audit/revisions browser projection.

**Rationale and benefit:** Least-privilege browser artifacts, family-confined
validators, and security evidence independent of disposable TypeScript files
make the final boundary explicit and extensible.

**Compatibility impact:** Protected generated implementation only.

**Risk if unresolved:** Sensitive administrative material can be reintroduced,
and broad runtime assets continue to inflate unrelated feature graphs.

**Exit:** Generation/drift/policy/shape checks pass; deleted files stay absent;
backend audit/revisions outputs remain; scanner finds no forbidden signals;
error/WS arrays are absent; Network Flow code and validators are confined to
its owning graph; package and Markdown checkpoints pass. The checkpoint
records `domain vocabulary unchanged`.

### I2-01B — Enforced dead-code green gate

**Areas:** verification and this tracker only.

**Remediation:** Run the exact package gate established in I2-01A without
changing its semantics. Any finding returns to its owning workstream for
cleanup; this slice adds no suppression or product cleanup.

**Rationale and benefit:** The remediation is incomplete until the permanent
gate proves the final authored graph and rejects future compatibility debris.

**Compatibility impact:** None.

**Risk if unresolved:** The repository has no durable evidence that cleanup is
complete.

**Exit:** `make protocol-ts-dead-code-check` reports zero authored findings and
zero suppressions; `make frontend-fallow-static`, harness smoke, JSON shape,
package slice, and Markdown checkpoint pass. The checkpoint records
`domain vocabulary unchanged`.

### I2-07 — Final validation, documentation, and handoff

**Areas:** documentation, validation evidence, tracker, and handoff.

**Remediation:** Update the development guide for the seven family entrypoints
and protected generated boundary. Run `make agent-finalize` before broad
validation, execute Section 18, compare two independent final production
graphs with I2-00, and close every tracker row with actual evidence. No product
cleanup occurs here; a code defect returns to its owning workstream and reruns
that checkpoint. Leave `docs/domain.md` unchanged.

**Rationale and benefit:** Completion is reproducible and restartable rather
than a source-only assertion.

**Compatibility impact:** None beyond I2-05.

**Risk if unresolved:** Stale documentation, hidden drift, or incomplete
security/build evidence could be mistaken for readiness.

**Exit:** Every Section 19 criterion passes, all required run roots and skipped
checks are recorded, and the tracker becomes `ITERATION 2 COMPLETE`.

## 17. Iteration 2 Top-Level Work Tracker

Allowed statuses are `TODO`, `IN_PROGRESS`, `DONE`, `DEFERRED`, and `BLOCKED`.
At most one implementation row may be `IN_PROGRESS`. The planning activation
row becomes `DONE` only after the tracker-only Markdown checkpoint passes.

| ID | Work item | Workstream | Status | Dependencies | Completion evidence |
| --- | --- | --- | --- | --- | --- |
| I2-PT-001 | Activate the Iteration 2 controlling tracker on the clean Iteration 1 commit. | planning checkpoint | DONE | Iteration 1 complete | Only this tracker changed; current posture, decisions, findings, sequence, and restart state are present; Markdown lint passed at `.cartulary/test-results/20260804T010606Z-p3646124`. |
| I2-PT-001A | Amend the controlling plan into independently attainable workstream checkpoints. | tracker amendment | DONE | I2-PT-001 | Red/green gates and fixture activation states match the executable sequence; tracker-only diff passed and Markdown lint passed at `.cartulary/test-results/20260804T013054Z-p3660359`. |
| I2-PT-002 | Freeze export, consumer, owner, replacement, compile, decoder, and build baselines. | I2-00 | DONE | I2-PT-001A | Five current specifiers and four absent future specifiers are recorded; owner behavior and validator discrepancy pass; two build graphs share normalized source-set digest `8533fbd491ddaa7f95314aa37bc4b4a10ae7c2c34199ffe8d42a456ed583b310`. |
| I2-PT-003 | Add and retain the package dead-code red gate. | I2-01A | DONE | I2-PT-002 | Supported Fallow route passes; fixture/harness smoke passes; the retained package gate reports 49 legacy authored findings and zero authored suppressions at `.cartulary/test-results/20260804T015406Z-p3710722`. |
| I2-PT-004 | Adopt v4 projections, v2 machine owners, family validators, and executable tolerance. | I2-02 | DONE | I2-PT-003 | Registry v4, both v2 owners, exact frozen projections, family validators, policy conformance, generation/drift/policy/shape, package, type, fast, and filtered Extensions owner checks pass while the legacy root remains buildable; Markdown checkpoint passed at `.cartulary/test-results/20260804T023149Z-p3822048`. |
| I2-PT-005 | Add seven authored family entrypoints and least-privilege generated imports. | I2-03 | DONE | I2-PT-004 | Seven public compile fixtures, family runtime tests, exact owner-allowlist enforcement, package/type/boundary/build/scanner checks, and the temporary nine-export manifest pass; Markdown checkpoint passed at `.cartulary/test-results/20260804T024724Z-p3852178`; the root and `./core-http` remain transitional. |
| I2-PT-006 | Migrate every repository consumer to its owning family entrypoint. | I2-04 | DONE | I2-PT-005 | No live root or `./core-http` imports remain outside retained characterization; typed registry values and operation-specific/indexed HTTP types replace generic parsing and component aliases; affected owner slices pass; Markdown checkpoint passed at `.cartulary/test-results/20260804T030319Z-p3932815`. |
| I2-PT-007 | Remove the root, `./core-http`, compatibility types, generic artifacts, and redundant decoder. | I2-05 | DONE | I2-PT-006 | Exact v2 seven-export match, negative legacy compile fixtures, family tests, and no authored legacy source pass; Markdown checkpoint passed at `.cartulary/test-results/20260804T032140Z-p3979847`. |
| I2-PT-008 | Remove aggregate validators, generated barrel, and unused TypeScript projections at their source. | I2-06 | DONE | I2-PT-007 | Regeneration deleted the exact modules; backend projections remain byte-unchanged; drift, policy, shape, scanner, owner slices, and bundle isolation pass; Markdown checkpoint passed at `.cartulary/test-results/20260804T033318Z-p4085529`. |
| I2-PT-009 | Close the package dead-code gate green with zero authored findings or suppressions. | I2-01B | DONE | I2-PT-008 | Exact Make-owned gate reports zero authored findings and zero suppressions; public Fallow, harness, package, shape, type, boundary, build, and scanner checks pass; Markdown checkpoint passed at `.cartulary/test-results/20260804T034111Z-p4116941`. |
| I2-PT-010 | Complete final validation, documentation, tracker evidence, and handoff. | I2-07 | DONE | I2-PT-009 | Both development guides describe the seven-entrypoint and protected-generated boundaries; agent finalization, every required Section 18 gate, two deterministic final graphs, final boundary audits, and the Markdown checkpoint pass; all evidence and skips are recorded below. |

## 18. Iteration 2 Validation Matrix

The implementer selects the narrowest owner rows first and broadens according
to the workstream risk. `make agent-finalize` is required before the final broad
run. `RESULTS_DIR` is passed only when a suitable successful warm run exists;
otherwise the tracker records that retained-run maintenance was skipped because
it was unset.

| Gate | Required command or evidence | Required by | Final expectation |
| --- | --- | --- | --- |
| Generation | `make generate` | I2-02, I2-06, I2-07 | Pass; generated outputs change only through the generator. |
| Drift | `make generate-drift` | I2-02, I2-06, I2-07 | Pass with no generated drift. |
| Generated policy | `make generated-artifact-policy-check` | I2-02, I2-06, I2-07 | Pass; deleted projections are not expected outputs. |
| JSON ownership | `make json-shape-check` | I2-01A, I2-02, I2-05, I2-06, I2-01B, I2-07 | Pass for registry v4, entrypoints v2, HTTP operations v2, and harness owners/projections. |
| Package runtime | `make test-slice OWNER=package.protocol_ts` | Every implementation slice | Pass with exact family tests and stable catalog resolution. |
| Type graph | `make frontend-typecheck` | I2-00, I2-03 through I2-07 | Pass; final fixture accepts seven specifiers and rejects removed/private paths. |
| Import policy | `make frontend-import-boundary-check` | I2-03 through I2-07 | Pass with exact entrypoint generated allowlists and unconditional private/generated denies. |
| General reachability | `make frontend-fallow-static` | I2-01A, I2-01B, I2-07 | Public target executes successfully through its cache profile. |
| Package reachability | `make protocol-ts-dead-code-check` | I2-01A, I2-01B, I2-07 | Nonzero legacy findings are required only for I2-01A; final result has zero authored findings or suppressions. |
| Production build | `make build-web` | I2-00, I2-03 through I2-07 | Two clean final builds succeed and preserve deterministic entry/module graphs. |
| Protected artifacts | `make protocol-ts-browser-artifact-reachability` | I2-00, I2-03 through I2-07 | No audit/revisions path, digest, identifier, signature, or removed module appears in JS or source maps. |
| View schemas | `make test-slice OWNER=package.view_contracts` | I2-04, I2-07 | Typed registry and source artifacts preserve current package behavior. |
| Extensions | `make test-slice OWNER=module.extensions` | I2-04, I2-07 | Typed client/profile registries and discovery validation preserve owner behavior. |
| Imports | `make test-slice OWNER=module.imports` | I2-04, I2-07 | Import-target registry identity and interpretation remain correct. |
| Network Flow | `make test-slice OWNER=module.networkflow` and `make test-slice OWNER=web.networkflow` | I2-04, I2-06, I2-07 | Types, typed error registry, metadata, decoders, and feature graph pass. |
| Workbook consumers | Narrow affected `module.workbook` and `web.workbook` rows selected through `make task-guide` and `make explain-test-owner` | I2-04, I2-07 | Operation-specific request/response migration preserves behavior. |
| Browser behavior | `make browser-e2e-webserver-backed` and `make browser-e2e-stateful` | I2-07 | Pass after the final build and owner slices. |
| Broad repository | `make check` | I2-07 | Pass; any unrelated failure is reported with its run root and replacement evidence, never silently ignored. |
| Documentation | `make lint-markdown` | Every checkpoint | Pass before the next workstream. |

The final build comparison is graph-first. It MUST show that error and WS
artifact arrays are absent, audit/revisions signals remain absent, Network Flow
runtime assets are reachable only through their owning feature graph, the
Extensions client-support registry occurs exactly once, and required view
schema artifacts remain reachable. Byte counts are recorded and MUST NOT
regress without an owner-backed explanation, but no byte threshold can waive a
failed ownership or reachability condition.

## 19. Iteration 2 Binary Completion and Restart State

### 19.1 Binary completion criteria

| Criterion | Current result | Completion evidence required |
| --- | --- | --- |
| Exactly seven authored package specifiers exist. | PASS | `package.json`, entrypoints v2, seven positive compile fixtures, exact boundary comparison, and typecheck agree. |
| Root and `./core-http` are unsupported. | PASS | Negative compile/import fixtures, exact export-map enforcement, and repository search pass. |
| No handwritten compatibility declaration or alias remains. | PASS | The root, compatibility declarations, and all six facade modules are deleted; source search, typecheck, and family tests pass. |
| No generic artifact parser or custom extension decoder remains. | PASS | Source search is empty; typed registry access and generated HTTP policy tests pass. |
| No authored Protocol-TS dead file, value, type, or export remains. | PASS | The enforced package gate reports zero findings and zero suppressions while automatically retaining 268 owner-selected generated pass-through exports. |
| Generated barrel and unused runtime artifact projections are absent. | PASS | Registry v4 retains only the view-schema artifact collection; managed regeneration, drift, policy, and file inventory agree. |
| Backend audit/revisions generation remains intact. | PASS | Generated Go outputs are byte-unchanged and the platform Audit and filtered non-browser Revisions owner slices pass. |
| Browser audit/revisions security remains fail-closed. | PASS | The scanner derives paths, canonical digests, identifiers, and signatures from both protected contract roots and passes against every emitted JS chunk and map. |
| Error artifacts are absent and Network Flow is family-confined in production. | PASS | Removed-module source-map denies pass; every runtime Protocol-TS Network Flow projection and validator is confined to the lazy `NetworkFlowFeature` graph. |
| Wire validation and payload redaction are unchanged. | PASS | Final Protocol-TS, Extensions, Network Flow, collaboration, application, Workbook, Evidence, browser, and repository-wide checks pass, including the HTTP tolerance and payload-free failure characterization. |
| Every repository consumer uses its owning family entrypoint. | PASS | Import search, boundary policy, typecheck, and affected owner slices pass. |
| Documentation and handoff reflect the final boundary. | PASS | Both development guides, the retained superseded compatibility decision, this final tracker evidence, and Markdown lint pass; `docs/domain.md` is unchanged. |
| Repository-wide validation passes or every unrelated exception has replacement evidence. | PASS | Final `make check` passed 713/713 at `.cartulary/test-results/20260804T035325Z-p93926`; the one overload-sensitive application-row failure and pre-execution selector-limit failures have passing replacement evidence recorded below. |

### 19.2 Planning checkpoint log

| Time | Agent/session | State | Files touched | Commands and evidence | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T21:02:05-04:00 | Codex / Iteration 2 planning activation | Iteration 1 is complete at `d29322b6`; Iteration 2 is planned and not started; I2-PT-001 is done. | This tracker only. | Clean branch/commit/status inspection; targeted package, consumer, generated projection, source-map, and Fallow discovery; `git diff --check`; `make lint-markdown` passed at `.cartulary/test-results/20260804T010606Z-p3646124`. | None. | Stop after a final Markdown checkpoint. A later implementation task begins I2-00 only. |
| 2026-08-03T21:30:54-04:00 | Codex / tracker amendment | I2-PT-001A is done; no implementation workstream has started. | This tracker only. | `git diff --check` passed; `make lint-markdown` passed at `.cartulary/test-results/20260804T013054Z-p3660359`. Domain vocabulary unchanged. | None. | Begin I2-00 only after the amended tracker checkpoint passes again with its evidence recorded. |
| 2026-08-03T21:38:26-04:00 | Codex / I2-00 | I2-00 is done; current/future specifiers, replacements, identities, redaction, owner-compatible discovery decoding, generated-validator discrepancy, and repeatable pre-change graphs are frozen. | `packages/protocol-ts/src/index.test.ts`; this tracker. | Package slice passed at `.cartulary/test-results/20260804T013413Z-p3664564`; typecheck passed at `.cartulary/test-results/20260804T013413Z-p3664703`; import boundary passed at `.cartulary/test-results/20260804T013511Z-p3670859`; two cache-disabled builds passed at `.cartulary/test-results/20260804T013539Z-p3675675` and `.cartulary/test-results/20260804T013657Z-p3687080`; both normalized source sets produced digest `8533fbd491ddaa7f95314aa37bc4b4a10ae7c2c34199ffe8d42a456ed583b310`; the scanner passed after each retained graph, including through the package slice. Domain vocabulary unchanged. | The first boundary run failed at `.cartulary/test-results/20260804T013413Z-p3664671` because expected-unresolved imports were placed in package source; the file was deleted and the passing boundary run is replacement evidence. No owner contradiction. | Run the I2-00 Markdown checkpoint, then begin I2-01A only. |
| 2026-08-03T21:56:42-04:00 | Codex / I2-01A | I2-01A is done at the required red state; no package implementation changed. | `.fallowrc.json`; `tools/harness_cache_registry.json`; `tools/task_surface_owner.json`; generated task-surface/topology projections through `make generate`; `tools/harness/static-analysis/protocol-ts-dead-code-check.mjs`; its fixture shell; Fallow smoke and harness contract tests; this tracker. | `make frontend-fallow-static` passed at `.cartulary/test-results/20260804T015345Z-p3709791`; `make protocol-ts-dead-code-check` produced the required nonzero report with 49 authored findings, zero suppressions, and retained root `.cartulary/test-results/20260804T015406Z-p3710722`; final Fallow fixture/harness smoke passed at `.cartulary/test-results/20260804T015606Z-p3719766`; `make generate` passed at `.cartulary/test-results/20260804T015553Z-p3717446`; generation drift passed at `.cartulary/test-results/20260804T015622Z-p3721889`; JSON shape passed at `.cartulary/test-results/20260804T015500Z-p3713464`; harness contract passed at `.cartulary/test-results/20260804T015539Z-p3717045`; script and shell lint passed at `.cartulary/test-results/20260804T015622Z-p3722074` and `.cartulary/test-results/20260804T015622Z-p3722104`. Domain vocabulary unchanged. | Initial harness contract run `.cartulary/test-results/20260804T015500Z-p3713644` exposed the intentional internal-target count change; the authored contract expectation was updated from 144 to 145 and the passing rerun replaces it. No finding was suppressed; no owner contradiction. | Run the I2-01A Markdown checkpoint, then begin I2-02 only. |
| 2026-08-03T22:31:09-04:00 | Codex / I2-02 | I2-02 is done; v4 exact projections, v2 entrypoint/HTTP owners, family validator modules, and executable response policies are active while the legacy root remains supported. | `contracts/index.json`; both `contracts/protocol-ts/*.v2.json` owners; v4/v2 schemas and schema attachments; `tools/contractgen` family planning, generation, and negative/positive tests; the Protocol-TS generator; JSON-shape semantic enforcement; generated registries, family validators, compatibility validator barrel, HTTP bindings, artifact collections, and affected integrity output through `make generate`; package conformance tests; this tracker. | Final `make generate` passed at `.cartulary/test-results/20260804T022138Z-p3751152`; drift at `.cartulary/test-results/20260804T022457Z-p3806941`; generated policy at `.cartulary/test-results/20260804T022508Z-p3809519`; JSON shape at `.cartulary/test-results/20260804T022151Z-p3753348`; `make test-fast` at `.cartulary/test-results/20260804T022220Z-p3757131`; package slice at `.cartulary/test-results/20260804T021923Z-p3744716`; frontend typecheck at `.cartulary/test-results/20260804T021935Z-p3749735`; the non-Playwright Extensions owner rows at `.cartulary/test-results/20260804T022949Z-p3818667`; format, script lint, and Biome lint at `.cartulary/test-results/20260804T022202Z-p3753877`, `.cartulary/test-results/20260804T023053Z-p3820947`, and `.cartulary/test-results/20260804T023058Z-p3821348`; Markdown checkpoint at `.cartulary/test-results/20260804T023149Z-p3822048`; `git diff --check` passed. Domain vocabulary unchanged. | Generation first failed for exhausted `/tmp` cache capacity at `.cartulary/test-results/20260804T021152Z-p3732479`; the reproducible `/tmp/cartulary-go-build` cache was removed and recreated. The next run `.cartulary/test-results/20260804T021251Z-p3734176` correctly stopped on the not-yet-migrated v1 generator path; the final generation run replaces both. JSON-shape runs `.cartulary/test-results/20260804T021131Z-p3731730` and `.cartulary/test-results/20260804T021951Z-p3750168` exposed stale topology and the v3 semantic checker; regeneration and the v4 checker replace them. Full Extensions owner runs `.cartulary/test-results/20260804T022516Z-p3809977` and `.cartulary/test-results/20260804T022826Z-p3815186` exposed that its Playwright row is group-owned and cannot execute as a raw owner row; the passing filtered owner slice replaces product coverage, and the grouped row remains assigned to I2-07's canonical `browser-e2e-stateful` run. No product assertion failed and no owner contradiction was found. | Begin I2-03 only. |
| 2026-08-03T22:46:34-04:00 | Codex / I2-03 | I2-03 is done; seven cohesive authored entrypoints now project exact family-owned generated modules through shared payload-safe decoder mechanics, while the root and `./core-http` remain supported only for the next migration slices. | `packages/protocol-ts/package.json`; seven `src/entrypoints/*.ts` modules; `src/internal/decoder.ts`; seven positive `*-entrypoint.compile.ts` fixtures; family runtime tests; the collaboration, Network Flow, and runtime-validation transitional facades; package characterization imports; frontend import-boundary owner enforcement and its positive/negative fixture; this tracker. | Package slice passed at `.cartulary/test-results/20260804T024234Z-p3830440`; final frontend typecheck at `.cartulary/test-results/20260804T024559Z-p3851333`; import-boundary check at `.cartulary/test-results/20260804T024400Z-p3836302`; its `web.architecture` owner slice, including the owner-allowlist fixture, at `.cartulary/test-results/20260804T024418Z-p3836926`; production build at `.cartulary/test-results/20260804T024514Z-p3842683`; protected-artifact scanner at `.cartulary/test-results/20260804T024524Z-p3850469`; format, script lint, and Biome lint at `.cartulary/test-results/20260804T024431Z-p3838467`, `.cartulary/test-results/20260804T024439Z-p3841579`, and `.cartulary/test-results/20260804T024529Z-p3850742`; Markdown checkpoint at `.cartulary/test-results/20260804T024724Z-p3852178`; `git diff --check` passed. The manifest has the required temporary nine exports, the seven owner paths exist, and domain vocabulary is unchanged. | Typecheck runs `.cartulary/test-results/20260804T023824Z-p3827188` and `.cartulary/test-results/20260804T024020Z-p3828571` exposed a narrow inferred registry-map key and unchecked compile-fixture indexing; the final typecheck replaces both. Import-boundary run `.cartulary/test-results/20260804T024156Z-p3829488` correctly rejected characterization imports that still reached generated registries; the tests now use the family entrypoints and the passing boundary run replaces it. Biome run `.cartulary/test-results/20260804T024447Z-p3842016` rejected non-null assertions in a compile fixture; the typed collection assertion replaces them. No owner contradiction was found. | Begin I2-04 only. |
| 2026-08-03T23:02:31-04:00 | Codex / I2-04 | I2-04 is done; every live repository consumer imports one owning family, typed singleton registries replace arbitrary artifact-path parsing, operation aliases/indexed response resources replace component aliases, and the application extension-contract forwarding shim is deleted. Root and `./core-http` remain only as package characterization/compatibility inputs for I2-05. | HTTP, collaboration, Network Flow, import-target, error, extension, and view-schema consumers under `apps/web` and shared packages; `packages/protocol-ts` registry type projections and compile/runtime tests; `tools/frontend_import_boundaries.json`; deletion of `apps/web/src/services/extensionContractAdapter.ts`; the web source guide row; this tracker. | Protocol-TS package slice passed at `.cartulary/test-results/20260804T025446Z-p3863713`; view-contracts and UI package slices at `.cartulary/test-results/20260804T025500Z-p3868819` and `.cartulary/test-results/20260804T025500Z-p3868825`; non-Playwright Extensions and Imports rows at `.cartulary/test-results/20260804T025550Z-p3871400` and `.cartulary/test-results/20260804T025622Z-p3875087`; affected module/web Network Flow rows at `.cartulary/test-results/20260804T025706Z-p3880274` and `.cartulary/test-results/20260804T025706Z-p3880298`; affected module/web collaboration rows at `.cartulary/test-results/20260804T025706Z-p3880284` and `.cartulary/test-results/20260804T025706Z-p3880314`; web application and Workbook slices at `.cartulary/test-results/20260804T025748Z-p3887232` and `.cartulary/test-results/20260804T025748Z-p3887237`; affected module Workbook and Evidence rows at `.cartulary/test-results/20260804T025848Z-p3906189` and `.cartulary/test-results/20260804T025848Z-p3906194`; final typecheck and import boundary at `.cartulary/test-results/20260804T030141Z-p3924549` and `.cartulary/test-results/20260804T030141Z-p3924551`; final cached production build and scanner at `.cartulary/test-results/20260804T030203Z-p3925396` and `.cartulary/test-results/20260804T030217Z-p3932282`; format and Biome lint at `.cartulary/test-results/20260804T030120Z-p3920910` and `.cartulary/test-results/20260804T030131Z-p3923980`; Markdown checkpoint at `.cartulary/test-results/20260804T030319Z-p3932815`; live root/`./core-http`, generic parser, shim, duplicate discovery alias, and generated resource-alias searches are empty; `git diff --check` passed. Domain vocabulary unchanged. | Typecheck runs `.cartulary/test-results/20260804T025142Z-p3858293` and `.cartulary/test-results/20260804T025302Z-p3859141` exposed overly narrow literal-union registry types, two raw generated component imports, and unchecked fixture indexing; widened readonly family row types, operation-specific/indexed types, and safe fixtures replace them. Boundary run `.cartulary/test-results/20260804T025905Z-p3910092` rejected direct imports outside the declared application adapter layer; Workbook now consumes its extension application type and extension availability is the explicit protocol adapter. Typecheck runs `.cartulary/test-results/20260804T030002Z-p3911153` and `.cartulary/test-results/20260804T030024Z-p3911706` refined that application type to accept owner data and existing test constructors; the final pass replaces them. Playwright, visual, accessibility, and measurement rows did not run because this slice changes no UI geometry, accessibility contract, or claim-bearing measurement; canonical browser behavior remains assigned to I2-07. No owner contradiction or product-test failure was found. | Begin I2-05 only. |
| 2026-08-03T23:20:29-04:00 | Codex / I2-05 | I2-05 is done; the private package now exposes exactly the seven v2-owned family specifiers, while the aggregate root, `./core-http`, handwritten compatibility layer, generic artifact APIs, and redundant discovery decoder are removed. | `packages/protocol-ts/package.json`; deletion of `src/index.ts`, `src/public-compatibility.compile.ts`, and all six `src/facade/*.ts` modules; family conformance/runtime tests; `src/unsupported-entrypoints.compile.ts`; entrypoint runtime tests; exact export-map enforcement and facade-free fixtures in the import-boundary checker; Protocol-TS test-family owner and generated topology; removal of the deleted extension shim from frontend source ownership; the retained compatibility decision marked superseded; this tracker. | `make generate` passed at `.cartulary/test-results/20260804T031813Z-p3950080`; drift at `.cartulary/test-results/20260804T031828Z-p3952553`; final JSON shape at `.cartulary/test-results/20260804T032020Z-p3979044`; generated policy at `.cartulary/test-results/20260804T031955Z-p3972724`; package slice, including family conformance and scanner evidence, at `.cartulary/test-results/20260804T031828Z-p3952536`; typecheck at `.cartulary/test-results/20260804T031828Z-p3952755`; final import boundary at `.cartulary/test-results/20260804T031916Z-p3964722`; `web.architecture` at `.cartulary/test-results/20260804T031955Z-p3972857`; production build at `.cartulary/test-results/20260804T031943Z-p3965588`; the explicit protected-artifact target also exited zero; final format at `.cartulary/test-results/20260804T031913Z-p3961698`; Biome and script lint at `.cartulary/test-results/20260804T031955Z-p3973043` and `.cartulary/test-results/20260804T031955Z-p3973079`; Markdown checkpoint at `.cartulary/test-results/20260804T032140Z-p3979847`; exact export/source/deletion searches and `git diff --check` passed. Domain vocabulary unchanged. | Typecheck `.cartulary/test-results/20260804T031037Z-p3939563` exposed one stale deleted-index import, impossible literal comparisons, and a test-only undeclared compiler dependency; the final compile fixtures and passing typecheck replace it. A pre-execution package catalog selector failure exposed stale facade test titles; the family selectors replace them. Format first rejected non-ASCII catalog row ordering, then `.cartulary/test-results/20260804T031713Z-p3943513` rejected the unnecessary compiler dependency; the final format replaces both. Import-boundary run `.cartulary/test-results/20260804T031828Z-p3952828` correctly rejected the negative compile fixture until it received an explicit rule exclusion; the final passing boundary run replaces it. No owner contradiction or product-test failure was found. | Begin I2-06 only. |
| 2026-08-03T23:32:26-04:00 | Codex / I2-06 | I2-06 is done; only the live view-schema artifact collection remains, every aggregate browser projection and validator barrel is managed-deleted, protected owner inputs drive scanner evidence directly, and Network Flow runtime projections are source-map-confined to their lazy owning graph. | `contracts/index.json`; the Protocol-TS generator's obsolete-output reconciliation; generated deletion of `index.ts`, `protocol-validators.ts`, and audit, revisions, WS, errors, Extensions, and Network Flow artifact arrays through `make generate`; browser reachability scanner; production and fixture import policy with explicit permanent denies; this tracker. Backend audit/revisions Go files were inspected and are unchanged. | `make generate` passed at `.cartulary/test-results/20260804T032624Z-p3987412`; final drift at `.cartulary/test-results/20260804T033152Z-p4075880`; generated policy at `.cartulary/test-results/20260804T033221Z-p4084791`; final JSON shape at `.cartulary/test-results/20260804T033152Z-p4075915`; `make test-fast` at `.cartulary/test-results/20260804T032701Z-p3998667`; final package/scanner slice at `.cartulary/test-results/20260804T033059Z-p4069831`; typecheck at `.cartulary/test-results/20260804T032701Z-p3998466`; final import boundary at `.cartulary/test-results/20260804T033059Z-p4070007`; filtered non-Playwright Network Flow backend rows at `.cartulary/test-results/20260804T032910Z-p4054034` and all web Network Flow rows at `.cartulary/test-results/20260804T032701Z-p3998360`; platform Audit at `.cartulary/test-results/20260804T033152Z-p4075947`; filtered non-Playwright Revisions at `.cartulary/test-results/20260804T033152Z-p4076014`; fresh production build at `.cartulary/test-results/20260804T032921Z-p4055604`; the explicit strengthened scanner exited zero; final format at `.cartulary/test-results/20260804T033048Z-p4063084`; script and Biome lint at `.cartulary/test-results/20260804T033059Z-p4070063` and `.cartulary/test-results/20260804T033059Z-p4070072`; Markdown checkpoint at `.cartulary/test-results/20260804T033318Z-p4085529`; deleted-file, backend-diff, source-map, source-reference, and `git diff --check` inspections passed. Domain vocabulary unchanged. | Full Network Flow owner run `.cartulary/test-results/20260804T032701Z-p3998339` completed 46 non-browser units but its 15 grouped Playwright rows failed preflight because raw owner slicing does not forward `OWNER` into the session launcher; the filtered passing owner slice replaces backend/product coverage and final browser groups remain assigned to I2-07. One attempted filtered rerun passed an empty shell variable before Make and failed usage without a run root; the corrected explicit `ROWS` invocation passed. No product assertion, owner contradiction, or generated/backend regression was found. | Run I2-01B without changing gate semantics or suppressions. |
| 2026-08-03T23:40:07-04:00 | Codex / I2-01B | I2-01B is done; the permanent package reachability gate is green with zero authored findings and zero suppressions, and owner-selected generated pass-throughs remain automatically retained. | The package gate's owner-allowlist collector and `export type *` fixture; removal of unused HTTP response-decoder helpers, two view lookup helpers, and unused authored alias exports from family entrypoints; matching compile/runtime characterization; a live view-contracts consumer of the required view-schema entry type; this tracker. | Final `make protocol-ts-dead-code-check` passed with `authored_findings=0`, `authored_suppressions=0`, eight automated compile fixtures, and 268 automated generated pass-through exports at `.cartulary/test-results/20260804T033954Z-p4105132`; `make frontend-fallow-static` passed at `.cartulary/test-results/20260804T033922Z-p4102425`; final harness contract, including the named/star pass-through fixture, at `.cartulary/test-results/20260804T033954Z-p4104616`; JSON shape at `.cartulary/test-results/20260804T033922Z-p4102156`; final package slice at `.cartulary/test-results/20260804T033954Z-p4104387`; view-contracts slice at `.cartulary/test-results/20260804T033824Z-p4094623`; typecheck at `.cartulary/test-results/20260804T033824Z-p4094726`; boundary at `.cartulary/test-results/20260804T033922Z-p4102392`; replacement production build at `.cartulary/test-results/20260804T034019Z-p4109786`; the strengthened scanner exited zero; format at `.cartulary/test-results/20260804T033814Z-p4091379`; final script and Biome lint at `.cartulary/test-results/20260804T033954Z-p4104606` and `.cartulary/test-results/20260804T033954Z-p4104608`; Markdown checkpoint at `.cartulary/test-results/20260804T034111Z-p4116941`; `git diff --check` passed. Domain vocabulary unchanged. | Initial green-gate run `.cartulary/test-results/20260804T033413Z-p4089057` reported 176 findings: four unused authored helpers and aliases plus owner-allowed generated star pass-throughs misclassified by the gate. The collector was repaired under the I2-01A semantics without changing policy, producing 23 genuine/duplicate findings at `.cartulary/test-results/20260804T033556Z-p4090207`; the I2-05-owned dead APIs were then removed rather than suppressed. Intermediate zero run `.cartulary/test-results/20260804T033824Z-p4095235` and the final fixture-backed run replace both failures. No suppression, product assertion failure, or owner contradiction occurred. | Begin I2-07 only. |
| 2026-08-03T23:59:01-04:00 | Codex / I2-07 final handoff | I2-07 and Iteration 2 are complete; exactly seven family entrypoints, family-scoped validators and registries, executable tolerance, migrated consumers, least-privilege browser projections, and permanent dead-code enforcement are validated end to end. | `docs/guides/cartulary-dev-guide.md`; `docs/guides/cartulary_frontend_implementation_testing_guide.md`; final evidence and completion state in this tracker. No product cleanup occurred in this slice, and `docs/domain.md` is unchanged. | `make agent-finalize` passed before broad verification at `.cartulary/test-results/20260804T034243Z-p4120447`; retained-run maintenance was skipped because `RESULTS_DIR` was unset. Generation, drift, generated policy, JSON shape, package, type, boundary, and Fallow passed at `.cartulary/test-results/20260804T034301Z-p4122917`, `.cartulary/test-results/20260804T034313Z-p4125215`, `.cartulary/test-results/20260804T034313Z-p4125219`, `.cartulary/test-results/20260804T034313Z-p4125238`, `.cartulary/test-results/20260804T034313Z-p4125317`, `.cartulary/test-results/20260804T034324Z-p4132505`, `.cartulary/test-results/20260804T034324Z-p4132553`, and `.cartulary/test-results/20260804T034324Z-p4132546`. The package dead-code gate exited zero again with no changed policy. Two cache-disabled builds and protected-artifact scans passed at `.cartulary/test-results/20260804T034342Z-p4134390` and `.cartulary/test-results/20260804T034359Z-p4141497`; both graphs have 14 JS files, 14 maps, 2,966,453 JS bytes, 8,182,228 map bytes, 310 sources, source digest `20182c4c85bfb33a06aa2ceb9754e136a5fde39fc9838f0a754012a91d3cfd40`, and asset digest `b3a5ea110b3f7183d5573c6e0460b6b3c5fdd0baad854b9fdc253af7fe3bb3b2`. The matching final graphs intentionally differ from I2-00's aggregate-facade source digest `8533fbd491ddaa7f95314aa37bc4b4a10ae7c2c34199ffe8d42a456ed583b310`. View-contracts, UI-contracts, Extensions, Imports, backend/web Network Flow, backend/web collaboration, Evidence, web Workbook, and three bounded backend Workbook slices passed at `.cartulary/test-results/20260804T034447Z-p4148845`, `.cartulary/test-results/20260804T034447Z-p4148848`, `.cartulary/test-results/20260804T034447Z-p4148846`, `.cartulary/test-results/20260804T034447Z-p4148855`, `.cartulary/test-results/20260804T034520Z-p4158007`, `.cartulary/test-results/20260804T034520Z-p4158015`, `.cartulary/test-results/20260804T034520Z-p4158028`, `.cartulary/test-results/20260804T034520Z-p4158052`, `.cartulary/test-results/20260804T034601Z-p4171195`, `.cartulary/test-results/20260804T034634Z-p4176678`, `.cartulary/test-results/20260804T034634Z-p4176770`, `.cartulary/test-results/20260804T034634Z-p4176795`, and `.cartulary/test-results/20260804T034634Z-p4176807`. Webserver-backed browser tests passed 62/62 at `.cartulary/test-results/20260804T034759Z-p11012`; stateful browser tests passed 36/36 at `.cartulary/test-results/20260804T035131Z-p68587`; final `make check` passed 713/713 at `.cartulary/test-results/20260804T035325Z-p93926`; final Markdown checkpoint passed at `.cartulary/test-results/20260804T040020Z-p249021`. Exact export parity, deleted-file absence, backend-output presence, live-import removal, `docs/domain.md` no-diff, and `git diff --check` audits passed. | The five-way parallel owner run produced one lazy debug-harness timeout after 52/53 application rows passed at `.cartulary/test-results/20260804T034634Z-p4176672`; its exact row passed alone at `.cartulary/test-results/20260804T034748Z-p10437`, and both canonical browser groups plus final `make check` replace the overload-sensitive failure. Initial web application, web Workbook, and backend Workbook selections exceeded the 4096-byte row-selection limit before product execution; full web owners and three bounded backend chunks replace them. Dedicated visual, accessibility, and measurement suites were skipped because no UI geometry, accessibility contract, or claim-bearing measurement changed. No owner contradiction, product regression, or open blocker remains. | Handoff complete; no implementation workstream remains active. |

### 19.3 Restart instructions

Iteration 2 is complete and no workstream is active. Preserve the dead-code gate
semantics, zero-suppression rule, zero-finding green result, 49-finding red
baseline, automated owner-selected generated pass-through handling, v4/v2
owners, family-specific validators, executable response policies, exact
generated-module allowlists, seven authored family entrypoints, exact
seven-export package boundary, negative removed/private compile fixtures,
facade-free authored source graph, and the fully migrated consumer graph. Also
preserve the final generated inventory, unchanged backend audit/revisions
outputs, owner-input-derived security scanner, and Network Flow graph
confinement. A future change begins from this completed record and opens a new
tracker or explicitly amends this one before changing any of those boundaries.
