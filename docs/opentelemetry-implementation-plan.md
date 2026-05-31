---
title: OpenTelemetry Implementation Plan
status: implementation-plan
document_class: nlspec-style closed implementation plan
created_at: 2026-05-30
---

## Source limits

This plan is based on the uploaded specification corpus, not on live repository inspection. It does not claim that any named package, Make target, schema, file, or directory already exists unless the cited documents state it.

This plan is a revised implementation-sequencing and repository-integration contract. It is written in NLSpec style, but it does not replace the adopted OpenTelemetry instrumentation NLSpec as the telemetry behavior owner.[^23] It closes implementation-plan boundaries by defining conformance modes, repo-control artifacts, canonical paths, phase gates, required fixture sets, and acceptance evidence.

A repo-local fact that is not available from the uploaded corpus MAY appear only as `TODO(repo-adoption): <specific missing value>`. A blocking `TODO(repo-adoption)` MUST prevent `adopted_conformant` and `otel_release_ready` status. A conformant implementation MUST contain zero blocking `TODO(repo-adoption)` values in conformance-visible artifacts, phase gates, fixture registries, acceptance mappings, and release-readiness manifests.

| Value class | Allowed in this plan | Allowed in `adopted_conformant` | Required treatment |
| --- | --- | --- | --- |
| Uploaded-corpus fact | Yes | Yes, if still adopted by repository authority | Cite exact owner source. |
| Live-repo fact not inspected | Yes, only as `TODO(repo-adoption)` | No | Mark blocking and name the required owner artifact. |
| Proposed conformance-visible path | Yes, only if made canonical by this plan | Yes | Put the path in the conformance-visible artifact registry. |
| Alternative conformance-visible path | No | No, unless this plan is revised | Treat as incompatible until revised. |
| Exact package version | Yes, only as repo-control placeholder | Yes, only when pinned in repo-control files | Pin in `go.mod`, `go.sum`, package manifests, workspace manifests, lockfiles, or a declared generated-constant manifest before conformance. |

## Status, authority, and conformance modes

This plan is derived from the Cartulary document set. Core 00 through Core 04 remain the current implementation-conformance corpus. Core 05 governs only claim-bearing timed or fixture-sensitive publication. Adopted subsystem NLSpecs may define bounded requirements for their named subsystem, and an adopted telemetry NLSpec owns telemetry generation, telemetry configuration, telemetry export, resource identity, privacy, and telemetry verification only.[^1]

The OpenTelemetry instrumentation NLSpec is `draft/proposed` until adopted. Before adoption, it does not widen the accepted deployment-configuration schema. After adoption, its closed `telemetry.*` namespace may be accepted only under the Core 04 deployment-configuration owner model, which requires exact keys, types, defaults, bounds, omitted behavior, explicit-`null` behavior, validation errors, secret handling, cross-key rules, and startup failure behavior for the adopted namespace.[^2]

This implementation plan MUST use these conformance modes:

| Mode | Entry condition | `telemetry.*` config behavior | Export behavior | Harness behavior | Claim behavior |
| --- | --- | --- | --- | --- | --- |
| `pre_adoption` | OTel NLSpec is not adopted by repository authority | `telemetry.*` keys are unknown keys and fail under ordinary deployment-config unknown-key handling | No telemetry export is permitted | `make otel-conformance` MAY be absent or MAY fail with not-adopted status | No OTel conformance claim is permitted |
| `experimental_non_conformant` | Repository explicitly opts into experimental implementation before adoption | Experimental keys MUST be outside `telemetry.*` or MUST be marked non-conformant by startup diagnostics and harness status | Network export default MUST be disabled | Any existing OTel harness target MUST emit non-conformant status | No OTel conformance claim is permitted |
| `adopted_incomplete` | OTel NLSpec is adopted, but one or more phase gates, decision rows, or acceptance criteria are incomplete | Closed `telemetry.*` namespace is parsed; invalid keys and values fail closed | Network export MUST remain disabled unless exporter, privacy, browser-boundary, and corpus gates pass | `make otel-conformance` MUST fail with incomplete criteria | No OTel conformance claim is permitted |
| `adopted_conformant` | OTel NLSpec is adopted, all blocking decisions are closed, all local `OIP-AC-*` and imported `OTEL-AC-*` criteria pass, and no blocking TODO remains | Closed `telemetry.*` namespace is accepted according to the adopted contract | Export is allowed only by explicit valid configuration | `make otel-conformance` passes and emits the required summary schema | OTel subsystem conformance claim is permitted |

Startup diagnostics, harness summaries, conformance manifests, and release-readiness output that report OTel status MUST distinguish `pre_adoption`, `experimental_non_conformant`, `adopted_incomplete`, and `adopted_conformant`.

The telemetry implementation MUST remain a logical boundary inside the modular monolith. It MUST NOT add a telemetry sidecar, required Collector deployable, monitoring-vendor dependency, Prometheus service, browser telemetry service, or additional Cartulary deployable. Cartulary’s base deployment remains one application deployable plus Postgres and S3-compatible object storage.[^3]

## Contract ownership and closure model

A contract family that appears in more than one normative document MUST have one primary owner. Non-owner sections MAY restate the owner only to declare local consequences, implementation sequencing, storage realization, or conformance checks; when a non-owner restatement and owner differ, the owner governs and the restatement is editorial drift to repair.[^1]

| Contract family | Primary owner | This plan owns | This plan MUST NOT own |
| --- | --- | --- | --- |
| Telemetry behavior | Adopted OpenTelemetry instrumentation NLSpec | Phase sequencing, repo-control materialization, canonical conformance-visible paths, harness wiring, fixture ownership, and acceptance mapping | Independent signal semantics, privacy policy, exporter behavior, or configuration behavior that is already owned by the adopted OTel NLSpec |
| Deployment-config container | Core 04 plus adopted subsystem NLSpec namespace | Telemetry namespace integration tasks and validation wiring | Core 04 config artifact discovery, overlay grammar, unknown-key rejection, validation envelope, or startup validation mechanics |
| Harness mechanics | Adopted Testing Harness NLSpec | Required use of `make otel-conformance`, required retained-artifact separation, and local acceptance evidence mapping | Make invocation semantics, scheduler semantics, result-root normalization, output classes, and schema-validation mechanics |
| Repository implementation layout | Development/bootstrap guides plus repo-control files | Canonical conformance-visible OTel paths required by this plan | Product runtime behavior, ordinary domain route behavior, or non-telemetry module ownership |
| Claim-bearing publication | Core 05 | Requirement that telemetry evidence is not a publication claim by itself | Benchmark profiles, benchmark manifests, and public timed or fixture-sensitive publication predicates |

A phase MAY summarize an owner contract only as an implementation consequence. It MUST NOT restate owner behavior as a second source of truth. If this plan and the adopted owner differ, the owner governs and this plan MUST be repaired.

## Terminology and forbidden ambiguity

| Term | Required meaning |
| --- | --- |
| `adopted OTel NLSpec` | The repository-adopted version of `opentelemetry-instrumentation-nlspec.md` or its successor. |
| `ordinary instrumentation unit` | Product code that records telemetry through API-facing accessors only. |
| `telemetry bootstrap boundary` | The only boundary that may configure SDK providers, processors, metric readers, exporters, samplers, shutdown, and telemetry self-diagnostics. |
| `conformance-visible artifact` | A file, target, schema, registry, fixture, or retained artifact used to decide OTel conformance. |
| `raw capture` | Unnormalized telemetry retained under the harness run root. |
| `normalized golden` | Canonical, committed, volatile-normalized telemetry expectation. |
| `blocking TODO` | A repo-adoption value that prevents `adopted_conformant` status. |
| `OpenTelemetry Core package` | An upstream OpenTelemetry package or plugin family, not Cartulary Core 00 through Core 05. |

The plan MUST NOT use `OpenTelemetry Core package` as permission for ordinary instrumentation units to import SDK, exporter, processor, sampler, propagator, metric-reader, log-processor, resource-detector, autoconfiguration, declarative-configuration, or plugin-provider packages.

Vague completion language is forbidden in conformance-visible requirements. A requirement that describes scope by example, vague representativeness, preference, proof without predicate, matching without equality type, appropriateness, need-based latitude, or unseen repository alternatives is invalid unless a closed table, exact fixture set, or binary predicate defines the allowed latitude.

## Partial implementation and omission semantics

Partial implementation MAY exist only as implementation-in-progress. It MUST NOT create a conformance claim, MUST NOT enable network export by default, and MUST NOT change product-visible behavior or committed state relative to an export-disabled baseline.

| Feature family | Omitted implementation before phase complete | Runtime default | Export behavior | Harness behavior |
| --- | --- | --- | --- | --- |
| Source snapshot | Missing or TODO snapshot | OTel conformance unavailable | No export | `otel-conformance` fails source-baseline gate |
| Configuration parser | Missing telemetry parser | `telemetry.*` unknown before adoption; invalid after adoption | No export | Config tests fail |
| Accessors | Missing wrappers | Product code MUST NOT call SDK directly | No export | Import-boundary tests fail |
| Traces | Missing span registry | Tracing disabled outside harness fixtures | No network export | Span-shape tests fail |
| Metrics | Missing metric registry | Metrics disabled outside harness fixtures | No network export | Metric tests fail |
| Logs | Missing bridge | Bridge disabled | No OTel logs | Log tests fail if enabled path absent |
| Exporter | Missing exporter | `telemetry.exporter.kind='none'` only | No export | Exporter tests fail |
| Browser boundary | Missing scan contract | Browser MUST NOT export telemetry | No direct browser export | Browser gate fails |
| Golden corpus | Missing normalized corpus | No conformance claim | No release-ready export | Corpus gate fails |

## Phase 0: Adoption and decision closure

### Objective

Make the draft OTel NLSpec implementable as a repository-controlled subsystem contract before production instrumentation is written.

### Decision-closure registry

All rows in this registry MUST be closed before `adopted_conformant` status. The OTel NLSpec lists these open decisions as unresolved until repository adoption; this plan makes their closure and failure behavior conformance-visible.[^4]

| Decision ID | Decision | Final required value | Owner artifact | Temporary placeholder | Conformance failure if unresolved | Verification |
| --- | --- | --- | --- | --- | --- | --- |
| `OTEL-DQ-001` | Semantic-convention model digest | Concrete lowercase 64-character SHA-256 digest over adopted semantic-convention model files | `contracts/otel/otel_source_snapshot.json` | `TODO(repo-adoption): compute semconv_model_digest` | Source-baseline validation fails | Digest parser rejects missing, placeholder, malformed, or non-64-hex values |
| `OTEL-DQ-002` | Generated-constant source | Exact generator source, package, or codegen version | `contracts/otel/generated_constants_manifest.json` | `TODO(repo-adoption): pin generated constants` | Generated-constant gate fails | Drift check verifies emitted standard names derive from pinned source or explicit allowlist |
| `OTEL-DQ-003` | SDK package set | Exact API, SDK, exporter, logs, metrics, trace, semantic-convention constants, bridge, and adapter package versions | Repo package manifests plus `contracts/otel/otel_source_snapshot.json` | `TODO(repo-adoption): pin OTel package versions` | Source-baseline validation fails | Static package inventory includes every used OTel family |
| `OTEL-DQ-004` | Public error binding | Mapping from `cartulary.error_code` and `cartulary.error_class` to the adopted public error registry | `contracts/errors/*` plus `contracts/otel/error_class_registry.json` | `TODO(repo-adoption): bind error registry` | Attribute-registry validation fails | Golden corpus includes safe public error-code and closed error-class cases |
| `OTEL-DQ-005` | Golden corpus path | `internal/testutil/golden/otel` for committed normalized corpus; raw captures retained only under harness run root | This plan plus harness registry | None | Artifact gate fails if path differs without plan revision | Corpus comparison reads committed path and run-root raw captures separately |
| `OTEL-DQ-006` | Retry jitter testing | Default: bounds-only conformance tests; deterministic RNG hook is optional test support and MUST NOT affect runtime behavior | This plan plus `otel-conformance` fixtures | None | Retry gate fails if tests require exact jitter without a declared deterministic hook | Retry tests assert bounds, cutoff, disabled retry, permanent rejection, timeout, shutdown, and non-blocking behavior |

### Required work items

1. Adopt the OTel NLSpec through the repository authority process or keep the implementation in `pre_adoption` or `experimental_non_conformant` mode.
2. Materialize all Phase 0 decision closures in the owner artifacts listed above.
3. Update the STRIDE threat model before any release that adds telemetry exporter configuration, telemetry exporter headers or secret references, retained telemetry artifacts, diagnostics, telemetry redaction, attribute governance, or browser diagnostics and browser telemetry-boundary behavior.[^5]
4. Add a conformance-status manifest entry that reports one of the four conformance modes exactly.

### Acceptance gates

| Gate | Required pass condition | Local criterion |
| --- | --- | --- |
| Adoption gate | The OTel NLSpec is adopted, or the implementation remains in `pre_adoption` or `experimental_non_conformant` and makes no conformance claim. | `OIP-AC-002` |
| Decision gate | `OTEL-DQ-001` through `OTEL-DQ-006` have final values. No blocking TODO remains in `adopted_conformant`. | `OIP-AC-003` |
| Config gate | Before adoption, `telemetry.*` keys fail as unknown. After adoption, only the closed namespace is accepted. | `OIP-AC-005` |
| Threat-model gate | STRIDE material covers telemetry endpoints, headers, source snapshot, generated constants, raw captures, redaction, browser non-export, and runtime-failure invariance. | `OIP-AC-024` |

## Phase 1: Repo-control baseline and dependency boundary

### Objective

Pin the external standard baseline and prevent SDK/exporter leakage into ordinary instrumentation code.

### Conformance-visible artifact registry

The bootstrap guide reserves `contracts/otel` and `internal/testutil/golden/otel`; it also states that exact OTel versions, generated constants provenance, and source snapshot values remain repo-control facts when an adopted OTel NLSpec is active.[^6]

| Artifact | Canonical path | Encoding | Owner | Required contents | Failure behavior |
| --- | --- | --- | --- | --- | --- |
| OTel source snapshot | `contracts/otel/otel_source_snapshot.json` | UTF-8 canonical JSON | This plan plus adopted OTel NLSpec | OTel spec ref, semantic-convention ref, commit SHAs, source paths, document statuses, digest, generated constants version, SDK versions | `otel-conformance` fails before provider tests |
| Generated constants manifest | `contracts/otel/generated_constants_manifest.json` | UTF-8 canonical JSON | This plan | Generator/package source, version, input digest, output digest | Generated-constant gate fails |
| OTel import-boundary manifest | `contracts/otel/import_boundary.json` | UTF-8 canonical JSON | This plan | Allowed API packages, bootstrap package roots, forbidden package families | Static import-boundary gate fails |
| Error-class registry | `contracts/otel/error_class_registry.json` | UTF-8 canonical JSON | This plan plus error-registry owner | Closed low-cardinality error classes and public error-code binding | Attribute validation fails |
| Golden corpus | `internal/testutil/golden/otel/**` | Canonical JSON files | This plan plus harness | Normalized emitted telemetry corpus | Corpus gate fails |

### Implementation layout

| Boundary | Canonical package path | May vary without plan revision? | Required invariant |
| --- | --- | ---: | --- |
| Telemetry bootstrap | `internal/platform/telemetry` | No | Only this boundary may import SDK, exporters, processors, samplers, metric readers, log processors, resource detectors, declarative configuration, autoconfiguration, or plugin providers. |
| API-facing accessors | `internal/platform/telemetry/api` | No | Ordinary modules may import only this accessor package and OTel API packages. |
| Hostile fixtures | `internal/testutil/golden/otel` and OTel conformance fixtures | No | Hostile fixtures may intentionally import forbidden packages only to assert containment. |

### Static import-boundary matrix

| Package family | Ordinary instrumentation units | Telemetry bootstrap | Hostile fixture tests |
| --- | ---: | ---: | ---: |
| OTel API | Allowed | Allowed | Allowed |
| Telemetry accessor package | Allowed | Allowed | Allowed |
| OTel SDK | Forbidden | Allowed | Allowed only for negative fixture setup |
| Exporter packages | Forbidden | Allowed | Allowed only for negative fixture setup |
| Processor or metric-reader packages | Forbidden | Allowed | Allowed only for negative fixture setup |
| Sampler packages | Forbidden | Allowed | Allowed only for negative fixture setup |
| Resource detector packages | Forbidden | Forbidden unless explicitly wrapped and disabled | Allowed only for detector-suppression fixtures |
| Declarative config or autoconfiguration | Forbidden | Forbidden unless used only for named containment fixtures | Allowed only for hostile fixtures |
| Plugin-provider packages | Forbidden | Forbidden | Allowed only for hostile fixtures |

The development guide already states that ordinary instrumentation units import only OTel API packages and telemetry bootstrap accessors, while SDK, exporter, processor, sampler, metric-reader, log-processor, resource, declarative-config, autoconfiguration, and plugin-provider imports belong only to the telemetry bootstrap boundary or hostile-fixture tests.[^7]

### Required work items

1. Generate `contracts/otel/otel_source_snapshot.json` from the adopted OTel baseline object.
2. Generate `contracts/otel/generated_constants_manifest.json` from the adopted semantic-convention source and code-generation process.
3. Generate `contracts/otel/import_boundary.json` from the package-family matrix above.
4. Add static checks that fail on forbidden imports outside the telemetry bootstrap and hostile fixtures.
5. Add generated-constant drift checks for standard attributes and metric names adopted from the pinned semantic-convention model.

### Acceptance gates

| Gate | Required pass condition | Local criterion |
| --- | --- | --- |
| Snapshot gate | Conformance fails if `main`, short SHAs, placeholder digests, missing source paths, missing document status, or missing SDK package versions appear in `otel_source_snapshot`. | `OIP-AC-004` |
| Import-boundary gate | Static tests assert SDK/exporter/package-family imports are absent outside allowed boundaries. | `OIP-AC-007` |
| Generated-constant gate | Emitted standard attributes and standard metrics either come from generated pinned sources or are explicitly allowlisted. | `OIP-AC-004` |

## Phase 2: Deployment configuration

### Objective

Add a closed, fail-closed telemetry configuration namespace under the existing Core 04 deployment-configuration surface.

### Telemetry configuration closure table

The adopted OTel NLSpec owns the exact telemetry key namespace. This plan requires the implementation to materialize the full key table into tests, generated schema, and startup validation rather than relying on prose references.[^8]

For `Environment binding`, `Core04 overlay` means the key is populated only through the Core 04 `CARTULARY__` overlay grammar after file load. Empty server-side environment values are omitted and use the key default. Invalid non-empty values fail startup. Raw values and secret-bearing values MUST NOT appear in diagnostics.[^9]

| Key | Type | Default | Bounds or values | Omitted behavior | Explicit `null` behavior | Empty server env | Environment binding | Secret-bearing? | Failure code |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `telemetry.enabled` | boolean | `true` | `true`, `false` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.otel_env_passthrough` | boolean | `false` | exactly `false` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.exporter.kind` | enum | `none` | `none`, `otlp_http`, `otlp_grpc` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.exporter.endpoint` | URL string or null | `null` | `http` or `https`, no userinfo, query, or fragment | Required when exporter kind is not `none`; otherwise default | Valid only when exporter kind is `none` | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.exporter.headers` | map string to string | `{}` | Header keys are non-empty ASCII token using letters, digits, `_`, `.`, or `-`; values are non-empty strings | Use default | Invalid | Omit | Core04 overlay or structured secret binding | Yes | `invalid_deployment_config` |
| `telemetry.exporter.protocol` | enum | Derived from exporter kind | `grpc`, `http/protobuf` | Derive from exporter kind | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.exporter.compression` | enum | `none` | `none`, `gzip` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.exporter.retry.enabled` | boolean | `true` | `true`, `false` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.exporter.retry.max_elapsed_ms` | integer | `30000` | `0..300000` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.exporter.retry.initial_interval_ms` | integer | `100` | `50..30000` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.exporter.retry.max_interval_ms` | integer | `5000` | `100..60000` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.exporter.retry.multiplier` | decimal | `2.0` | `1.0..5.0` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.traces.enabled` | boolean | `true` | `true`, `false` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.traces.sample_ratio` | decimal | `0.10` | `0.0..1.0` inclusive | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.traces.sampler_profile` | enum | `auto` | `auto`, `cartulary.sampler.always_on.v1`, `cartulary.sampler.always_off.v1`, `cartulary.sampler.traceidratio_compat.v1` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.traces.accept_remote_context` | boolean | `false` | exactly `false` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.metrics.enabled` | boolean | `true` | `true`, `false` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.metrics.temporality_profile` | enum | `cartulary.metrics.temporality.cumulative.v1` | exactly `cartulary.metrics.temporality.cumulative.v1` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.metrics.exemplars.enabled` | boolean | `false` | exactly `false` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.logs.bridge_enabled` | boolean | `false` | `true`, `false` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.logs.body_max_chars` | integer | `2048` | `0..8192` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.processor.max_queue_size` | integer | `2048` | `1..65536` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.processor.max_export_batch_size` | integer | `512` | `1..telemetry.processor.max_queue_size` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.processor.traces.schedule_delay_ms` | integer | `5000` | `100..300000` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.processor.metrics.schedule_delay_ms` | integer | `60000` | `100..300000` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.processor.logs.schedule_delay_ms` | integer | `1000` | `100..300000` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.processor.export_timeout_ms` | integer | `2000` | `100..10000` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.processor.overflow_policy` | enum | `drop_new` | exactly `drop_new` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.shutdown.flush_timeout_ms` | integer | `5000` | `100..30000` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.self_diagnostics.enabled` | boolean | `true` | `true`, `false` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.self_diagnostics.recursion_guard` | enum | `drop_telemetry_about_telemetry` | exactly `drop_telemetry_about_telemetry` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.resource.service_name` | string | `cartulary.app` | `1..128` ASCII letters, digits, `.`, `_`, or `-` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.resource.service_namespace` | string | `cartulary` | `1..128` ASCII letters, digits, `.`, `_`, or `-` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.resource.service_version` | string | Build version, else `0.0.0+unknown` | `1..128` non-empty string | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.resource.service_instance_id` | string or null | Generated UUID v4 per process start | Non-empty opaque string, maximum `128` Unicode scalar values | Generate default | Generate default | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.resource.deployment_environment_name` | string or null | `null` | `development`, `test`, `staging`, `production`, or custom non-empty token of maximum `128` ASCII letters, digits, `.`, `_`, or `-` | Omit attribute | Omit attribute | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.attribute.incident_correlation` | enum | `none` | `none`, `hmac_64bit` | Use default | Invalid | Omit | Core04 overlay | No | `invalid_deployment_config` |
| `telemetry.attribute.hmac_secret_ref` | string or null | `null` | Server-side secret reference | Required only when incident correlation is `hmac_64bit` | Valid only when incident correlation is `none`; otherwise invalid | Omit | Core04 overlay or secret reference | Yes | `invalid_deployment_config` |

### Cross-key validation matrix

| Rule ID | Inputs | Invalid condition | Required error code | Required reason code | Startup behavior | Secret-redaction rule |
| --- | --- | --- | --- | --- | --- | --- |
| `OTEL-CFG-001` | `exporter.kind`, `exporter.endpoint` | Exporter kind is `otlp_http` or `otlp_grpc` and endpoint is omitted or null | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | Endpoint value MAY be shown only after URL sanitization; no headers |
| `OTEL-CFG-002` | `exporter.kind`, `exporter.endpoint` | Exporter kind is `none` and endpoint is non-null | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | Redact endpoint if supplied by secret-bearing environment |
| `OTEL-CFG-003` | `exporter.kind`, `exporter.protocol` | `otlp_http` uses protocol other than `http/protobuf`, or `otlp_grpc` uses protocol other than `grpc` | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |
| `OTEL-CFG-004` | `processor.max_export_batch_size`, `processor.max_queue_size` | Batch size exceeds queue size | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |
| `OTEL-CFG-005` | Retry interval keys | `retry.max_interval_ms < retry.initial_interval_ms` | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |
| `OTEL-CFG-006` | `incident_correlation`, `hmac_secret_ref` | `hmac_64bit` without non-empty secret reference | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | Secret reference MAY be identified by safe reference name only; secret value forbidden |
| `OTEL-CFG-007` | Remote-context config | Any attempt to set `accept_remote_context=true` | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |
| `OTEL-CFG-008` | Sampler profile | Unsupported sampler or inconsistent sampler profile and sample ratio | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |
| `OTEL-CFG-009` | Export protocol | Unsupported protocol including `http/json` | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |
| `OTEL-CFG-010` | Effective config | Any per-signal endpoint key appears in effective behavior | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | Secret values forbidden |
| `OTEL-CFG-011` | Effective config | Any per-signal protocol or header divergence appears in effective behavior | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | Header values always redacted |
| `OTEL-CFG-012` | Exemplars | `telemetry.metrics.exemplars.enabled` is not `false` | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |
| `OTEL-CFG-013` | Log body bound | `telemetry.logs.body_max_chars` outside `0..8192` | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |
| `OTEL-CFG-014` | Environment passthrough | `telemetry.otel_env_passthrough=true` | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | OTel env values redacted |
| `OTEL-CFG-015` | External OTel config | Declarative config, autoconfig, plugin provider, ConfigProvider state, or resource detector attempts to create or alter runtime telemetry components | `invalid_deployment_config` | `invalid_telemetry_config` | Reject, ignore, or override before provider activation; fail if containment cannot be verified | Raw config content redacted |
| `OTEL-CFG-016` | Semantic-convention opt-in | `OTEL_SEMCONV_STABILITY_OPT_IN` would alter emitted telemetry shape | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | Raw value redacted |
| `OTEL-CFG-017` | Resource schema URL | Effective provider Resource has non-empty schema URL | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |
| `OTEL-CFG-018` | Metric View | View creates an exported stream name not exactly registered | `invalid_deployment_config` | `invalid_telemetry_config` | Fail before readiness | No secret values |

### OTel environment hazard registry

External OTel configuration inputs MUST have no behavior-selection effect. The adopted OTel NLSpec closes the environment hazard families and declarative-configuration hostile fixture set.[^10]

| Environment family | Members or pattern | Required behavior | Forbidden effect | Fixture ID |
| --- | --- | --- | --- | --- |
| SDK disablement | `OTEL_SDK_DISABLED` | Ignore for behavior selection | Disable providers or invert no-SDK tests | `OTEL-ENV-001` |
| Resource environment merge | `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_SERVICE_NAME`, `OTEL_ENTITIES` | Ignore or redact presence only | Add host, process, incident, user, deployment, entity, or service attributes | `OTEL-ENV-002` |
| SDK internal logging | `OTEL_LOG_LEVEL` | Ignore for product behavior | Change product logs, operator diagnostics, or secret exposure | `OTEL-ENV-003` |
| Propagators | `OTEL_PROPAGATORS` | Ignore | Enable Baggage, vendor propagators, or remote context propagation | `OTEL-ENV-004` |
| Trace sampler | `OTEL_TRACES_SAMPLER`, `OTEL_TRACES_SAMPLER_ARG` | Ignore | Override sampler profile | `OTEL-ENV-005` |
| Global attribute limits | `OTEL_ATTRIBUTE_COUNT_LIMIT`, `OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT` | Ignore for privacy | Act as privacy enforcement or alter redaction-before-recording | `OTEL-ENV-006` |
| Span limits | `OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT`, `OTEL_SPAN_ATTRIBUTE_VALUE_LENGTH_LIMIT`, `OTEL_SPAN_EVENT_COUNT_LIMIT`, `OTEL_SPAN_LINK_COUNT_LIMIT` | Ignore for authorization and privacy | Authorize extra attributes, events, links, or privacy bypass | `OTEL-ENV-007` |
| LogRecord limits | `OTEL_LOGRECORD_ATTRIBUTE_COUNT_LIMIT`, `OTEL_LOGRECORD_ATTRIBUTE_VALUE_LENGTH_LIMIT` | Ignore for privacy | Authorize raw log fields or privacy bypass | `OTEL-ENV-008` |
| Batch span processor | `OTEL_BSP_*` | Ignore | Override processor bounds, delay, timeout, or queue policy | `OTEL-ENV-009` |
| Batch LogRecord processor | `OTEL_BLRP_*` | Ignore | Enable or alter log export outside Cartulary config | `OTEL-ENV-010` |
| Exporter selection | `OTEL_TRACES_EXPORTER`, `OTEL_METRICS_EXPORTER`, `OTEL_LOGS_EXPORTER` | Ignore | Enable OTLP, stdout, zipkin, prometheus, vendor-native, or other export | `OTEL-ENV-011` |
| Global OTLP | `OTEL_EXPORTER_OTLP_*` | Ignore | Set endpoint, headers, protocol, timeout, compression, certificates, mTLS, or per-signal behavior | `OTEL-ENV-012` |
| Per-signal OTLP | `OTEL_EXPORTER_OTLP_TRACES_*`, `OTEL_EXPORTER_OTLP_METRICS_*`, `OTEL_EXPORTER_OTLP_LOGS_*` | Ignore | Create divergent per-signal endpoints, headers, protocol, or credentials | `OTEL-ENV-013` |
| Non-OTLP exporters | `OTEL_EXPORTER_ZIPKIN_*`, `OTEL_EXPORTER_PROMETHEUS_*`, implementation-specific exporter env vars | Ignore | Enable unsupported exporters | `OTEL-ENV-014` |
| Metrics exemplar | `OTEL_METRICS_EXEMPLAR_FILTER` | Ignore | Enable exemplars | `OTEL-ENV-015` |
| Metrics periodic reader | `OTEL_METRIC_EXPORT_INTERVAL`, `OTEL_METRIC_EXPORT_TIMEOUT` | Ignore | Override metric export cadence or timeout | `OTEL-ENV-016` |
| Declarative configuration | `OTEL_CONFIG_FILE`, `OTEL_EXPERIMENTAL_CONFIG_FILE` | Ignore and contain | Instantiate SDK components from file configuration | `OTEL-ENV-017` |
| Semantic-convention opt-in | `OTEL_SEMCONV_STABILITY_OPT_IN` | Ignore unless later revision adopts migration profile | Change emitted attribute names or duplicate conventions | `OTEL-ENV-018` |
| Language-specific OTel env vars | `OTEL_{LANGUAGE}_{FEATURE}` | Ignore unless later revision maps exact variables by name | Affect emitted telemetry or runtime behavior | `OTEL-ENV-019` |

### Required work items

1. Extend deployment-config parsing with the exact telemetry key set.
2. Generate schema validation for type, default, bound, omission, explicit-null, and empty-environment behavior.
3. Implement environment overlay parsing through the canonical deployment-configuration object before telemetry bootstrap.
4. Reject or contain external OTel environment variables, declarative configuration, SDK autoconfiguration, resource detectors, plugin providers, and ConfigProvider state before provider activation.
5. Add cross-key validation tests for every row in the matrix above.

### Acceptance gates

| Gate | Required pass condition | Local/imported criterion |
| --- | --- | --- |
| Unknown-key gate | Unknown `telemetry.*` keys fail before readiness. | `OTEL-AC-003`, `OIP-AC-005` |
| Null gate | Explicit `null` fails except where the key row permits nullable behavior. | `OIP-AC-005` |
| Environment-parser gate | Omitted, empty, valid, invalid, and explicit string `"null"` cases are tested for every parser family. | `OTEL-AC-004A` |
| Cross-key gate | Every cross-key rule fails before readiness with the required deployment-config error. | `OTEL-AC-004`, `OIP-AC-006` |
| Hostile-env gate | Every hazard-registry family has no runtime effect. | `OTEL-AC-006`, `OIP-AC-006` |
| Export-disabled gate | With `telemetry.exporter.kind='none'`, no exporter is created and no default localhost OTLP endpoint is contacted. | `OTEL-AC-005` |

## Phase 3: API-only instrumentation wrappers

### Objective

Give product modules safe telemetry APIs without exposing SDK mechanics or creating no-SDK failure modes.

### Accessor interface contract

Ordinary instrumentation units MUST obtain tracers, meters, and loggers only through API-facing provider accessors. The adopted OTel NLSpec requires registered scopes, build-version scope version where known, null schema URL, empty scope attributes, and no-SDK safety.[^11]

| Accessor | Input | Valid values | Output | Unknown input | No-SDK behavior | Concurrency behavior |
| --- | --- | --- | --- | --- | --- | --- |
| `tracer_for(scope_id)` | `scope_id` string | One registered instrumentation scope ID | API tracer handle | Static validation or startup validation fails; ad hoc scope creation forbidden | Returns no-op/API-safe tracer | Safe for concurrent calls |
| `meter_for(scope_id)` | `scope_id` string | One registered instrumentation scope ID | API meter handle | Static validation or startup validation fails; ad hoc scope creation forbidden | Returns no-op/API-safe meter | Safe for concurrent calls |
| `logger_for(scope_id)` | `scope_id` string | One registered instrumentation scope ID | API logger handle | Static validation or startup validation fails; ad hoc scope creation forbidden | Returns no-op/API-safe logger | Safe for concurrent calls |

### Instrumentation-scope registry

| Scope ID | Name | Version rule | Schema URL | Scope attributes | Owning module |
| --- | --- | --- | --- | --- | --- |
| `otel.scope.httpapi` | `cartulary.httpapi` | Build version else `0.0.0+unknown` | `null` | Empty | HTTP API |
| `otel.scope.workbook` | `cartulary.workbook` | Build version else `0.0.0+unknown` | `null` | Empty | Workbook/query/mutation |
| `otel.scope.collaboration` | `cartulary.collaboration` | Build version else `0.0.0+unknown` | `null` | Empty | WebSocket and presence |
| `otel.scope.jobs` | `cartulary.jobs` | Build version else `0.0.0+unknown` | `null` | Empty | Background jobs |
| `otel.scope.postgres` | `cartulary.postgres` | Build version else `0.0.0+unknown` | `null` | Empty | Postgres dependency |
| `otel.scope.objectstore` | `cartulary.objectstore` | Build version else `0.0.0+unknown` | `null` | Empty | Object storage dependency |
| `otel.scope.telemetry` | `cartulary.telemetry` | Build version else `0.0.0+unknown` | `null` | Empty | Telemetry self-diagnostics |

### No-SDK lifecycle table

| Runtime state | Provider installed? | Exporter installed? | Wrapper behavior | Product behavior |
| --- | ---: | ---: | --- | --- |
| `telemetry.enabled=false` | No | No | Accessors return API-safe no-op handles or leave API defaults in place | Product paths identical to no-instrumentation baseline |
| `telemetry.enabled=true`, SDK construction fails before readiness | No active provider | No | Startup fails before readiness | No product surface exposed |
| Test no-SDK profile | No | No | Accessors callable; no emitted telemetry | HTTP, workbook, jobs, Postgres, object-store, log-call, and WebSocket paths complete normally |

### Required work items

1. Add the accessor package at `internal/platform/telemetry/api`.
2. Add compile-time or startup-time scope registry validation.
3. Add OTel API spies or equivalent emitted-telemetry capture to assert null-like internal values are omitted before OTel API calls.
4. Add no-SDK runtime tests for the closed corpus paths listed in Phase 3: HTTP, workbook, jobs, Postgres, object-store, log-call, and WebSocket.

### Acceptance gates

| Gate | Required pass condition | Local/imported criterion |
| --- | --- | --- |
| Scope gate | No tracer, meter, or logger uses an unregistered scope. | `OTEL-AC-009`, `OIP-AC-007` |
| No-SDK gate | Product paths run unchanged when no SDK provider is installed. | `OTEL-AC-008`, `OIP-AC-007` |
| Null gate | API spy or emitted-capture tests assert no null attribute setter calls for spans, metrics, logs, resource attributes, or self-diagnostics. | `OTEL-AC-014A` |

## Phase 4: Resource identity and privacy substrate

### Objective

Build resource-identity closure and privacy controls before emitting meaningful telemetry.

### Resource registry

The resource registry is closed by the adopted OTel NLSpec. No other resource attribute may be emitted unless an SDK requires a `telemetry.sdk.*` attribute.[^12]

| Attribute | Requiredness | Source | Default | Null behavior | Forbidden source classes | Exported when |
| --- | --- | --- | --- | --- | --- | --- |
| `service.name` | Required | `telemetry.resource.service_name` | `cartulary.app` | Invalid after default resolution | SDK `unknown_service:*`, host, process, user, incident, object store | Always |
| `service.namespace` | Required | `telemetry.resource.service_namespace` | `cartulary` | Invalid after default resolution | Host, tenant, customer, incident | Always |
| `service.version` | Required | `telemetry.resource.service_version` | Build version else `0.0.0+unknown` | Invalid after default resolution | Operator text, paths, incident data | Always |
| `service.instance.id` | Required | Configured value or generated UUID v4 | Generated UUID v4 per process start | Generate default | Host, user, incident, process, container, filesystem, object-store identity | Always |
| `cartulary.deployment.profile` | Required | Deployment profile | Active profile | Invalid after default resolution | Incident data, customer names | Always |
| `cartulary.profile.claims` | Required | Sorted claimed profile IDs | Current claims | Invalid after default resolution | Incident data, extension secrets | Always |
| `deployment.environment.name` | Optional | `telemetry.resource.deployment_environment_name` | `null` | Omit attribute | Tenancy, authorization, incident identity | Only when configured |
| `telemetry.sdk.language` | SDK-required | SDK | SDK value | SDK-defined | Overridden product values | Always when SDK emits |
| `telemetry.sdk.name` | SDK-required | SDK | SDK value | SDK-defined | Overridden product values | Always when SDK emits |
| `telemetry.sdk.version` | SDK-required | SDK | SDK value | SDK-defined | Overridden product values | Always when SDK emits |

The effective provider Resource `schema_url` MUST be empty. Host, process, OS, container, Kubernetes, cloud, FaaS, browser, device, telemetry entity, environment-variable, and vendor-specific resource detectors MUST be disabled or bypassed before provider activation.[^12]

### Forbidden-value action matrix

The forbidden value families are closed by the adopted OTel NLSpec and include incident-authored content, evidence content and identity, stable product identifiers, security material, request content, database content, network and infrastructure identity, exception detail, and Baggage or trace-state detail.[^13]

| Family | Detection basis | Default action | Replacement token allowed? | Drop item when | Drop metric labels | Diagnostic behavior |
| --- | --- | --- | --- | --- | --- | --- |
| Incident-authored content | Provenance from incident fields, artifacts, notes, handoffs, status reviews, lessons, findings, source text | Omit value before OTel API call | Only closed `cartulary.result` or `cartulary.error_class` tokens not derived from content | The telemetry item requires the content to retain meaning and has no safe class | Remove label set; record drop metric if non-recursive | Redacted diagnostic with family `incident_authored_content` only |
| Evidence content and identity | Provenance from evidence bytes, filenames, blob hashes, storage refs, object keys, handles, upload IDs | Omit value | No, except safe operation class such as `preview`, `download`, `attach` | Span/log/metric cannot be emitted without object-identifying value | Remove object labels; record drop metric if non-recursive | Redacted diagnostic with family `evidence_identity` only |
| Stable product identifiers | Field name or value provenance from incident, record, user, party, session, job, connection, transaction, conflict, or handle IDs | Omit value | Only public error code or closed type vocabulary when independently safe | A required telemetry field would otherwise contain a stable ID | Remove label set; record drop metric if non-recursive | Redacted diagnostic with family `stable_identifier` only |
| Security material | Secret source, auth routes, credentials, tokens, exporter headers, signing keys, object-store credentials | Omit value | No | Any part of item would carry security material | Remove label set; record drop metric if non-recursive | Diagnostic MUST NOT include raw or transformed secret |
| Request content | Raw URL path, query string, body, headers, cookies, auth header, response body | Omit value | Route family or route template only | No route template or safe class exists | Remove label set; record drop metric if non-recursive | Redacted diagnostic with family `request_content` only |
| Database content | SQL, bind values, query JSON, table names, projection names, source headers, search text | Omit value | Closed operation class and closed error class only | Database span/log depends on SQL-like detail for meaning | Remove label set; record drop metric if non-recursive | Redacted diagnostic with family `database_content` only |
| Network and infrastructure identity | Client IP, hostnames, container IDs, DB addresses, object endpoints, filesystem roots | Omit value | Deployment profile token only when from closed config vocabulary | Item would disclose infrastructure identity | Remove label set; record drop metric if non-recursive | Redacted diagnostic with family `infrastructure_identity` only |
| Exception detail | Exception message, stacktrace, cause chain, panic string, wrapped detail | Reduce to safe low-cardinality class | `cartulary.error_class` and safe `exception.type` class only | Message, stacktrace, or wrapped details cannot be separated | Remove label set; record drop metric if non-recursive | Redacted diagnostic with family `exception_detail` only |
| Baggage and trace-state detail | Inbound Baggage, inbound tracestate, vendor trace state | Ignore before context creation | No | Any telemetry item would preserve inbound values | Remove label set; record drop metric if non-recursive | Diagnostic MAY record redacted presence only |

Detection MUST be provenance-first. A value known to originate from an incident field, evidence field, object-store identifier, request body, SQL text, credential, or stable product identifier remains forbidden even if it does not match a literal scanner. Literal scanning is required for the conformance corpus but is not sufficient as the runtime privacy model. SDK truncation, SDK attribute limits, exporter queue overflow, Collector filtering, and backend filtering MUST NOT be treated as redaction.[^13]

### Required work items

1. Construct the Resource only from the closed registry.
2. Generate `service.instance.id` as an opaque value unless a valid opaque value is configured.
3. Disable or wrap resource detectors so forbidden resource attributes cannot become effective.
4. Implement pre-recording forbidden-value classification and action mapping.
5. Add a corpus value for every forbidden family across traces, metrics, logs, self-diagnostics, retained artifacts, exporter requests, and User-Agent.

### Acceptance gates

| Gate | Required pass condition | Local/imported criterion |
| --- | --- | --- |
| Resource gate | Exported telemetry contains only registered resource attributes plus SDK-required `telemetry.sdk.*`; Resource schema URL is empty. | `OTEL-AC-010` |
| Detector gate | Host/process/container/cloud/resource-detector data never appears even when environment variables and detector packages are present. | `OTEL-AC-011` |
| Forbidden-value gate | Every forbidden family is absent from spans, metrics, logs, resource attributes, diagnostics, retained artifacts, exporter requests, and User-Agent. | `OTEL-AC-012`, `OIP-AC-008` |
| Unknown-attribute gate | Unknown `cartulary.*` attributes are rejected or omitted before recording. | `OTEL-AC-013` |
| SDK-limit gate | SDK truncation or attribute limits are not used as redaction controls. | `OTEL-AC-014` |

## Phase 5: Trace instrumentation

### Objective

Emit only registered, low-cardinality spans for HTTP, workbook, collaboration, jobs, Postgres, and object storage.

### Sampler and context rules

The sampler profile MUST follow the adopted OTel NLSpec: `sample_ratio=0.0` maps to AlwaysOff, `1.0` maps to AlwaysOn, and fractional ratios use the `TraceIdRatioBased`-compatible profile. ProbabilitySampler, remote/adaptive sampling, remote parent sampling, sampler plugin providers, and environment-driven samplers are not adopted. Inbound `traceparent`, `tracestate`, Baggage, vendor trace headers, and sampled flags MUST be ignored before root span creation.[^14]

### Span registry implementation table

The adopted OTel NLSpec owns span family names and required/forbidden attributes. This table adds implementation-plan lifecycle boundaries and default span kind decisions.[^15]

| Span family | Span name | Kind | Start boundary | End boundary | Parent rule | Link rule | Required attributes | Optional attributes | Forbidden attributes | Status mapping | Events |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| HTTP server | `<HTTP_METHOD> <route_template>` | `server` | After route template is resolved and before route handler work starts | After response status and envelope are committed | Root server-owned span; inbound remote context stripped | None | OTel HTTP allowlist, `cartulary.route_family`, `cartulary.result` | `cartulary.error_code` when safe | Raw path, query, headers, cookies, user agent, client IP, body, IDs | Result table below | Disabled |
| Workbook query | `cartulary.workbook.query` | `internal` | Before workbook query execution against source/projection boundary | After result rows and metadata are assembled or rejection is returned | Child of current HTTP span when present; otherwise root server-owned span | None | `cartulary.view_schema_id`, `cartulary.operation='query'`, `cartulary.result` | `cartulary.error_code` | Saved-view ID, filters, search text, row values, projection table name | Result table below | Disabled |
| Workbook mutation | `cartulary.workbook.mutation` | `internal` | Before authoritative mutation validation begins | After commit, conflict response, or rejection response | Child of current HTTP span when present; otherwise root server-owned span | None | `cartulary.view_schema_id`, `cartulary.record_type`, `cartulary.operation`, `cartulary.result` | `cartulary.error_code` | Record ID, row version, client transaction ID, field values | Result table below | Disabled |
| Projection maintenance | `cartulary.workbook.projection` | `internal` | Before affected projection row computation starts | After projection update or failure classification | Child of current workbook mutation span when present | None | `cartulary.view_schema_id`, `cartulary.operation`, `cartulary.result` | `cartulary.error_code` | Projection table name, SQL, row IDs | Result table below | Disabled |
| WebSocket lifecycle | `cartulary.collaboration.websocket` | `internal` | Before subscribe/authorize/connect lifecycle step begins | After close, rejection, or accepted lifecycle step completes | Root server-owned span for connection lifecycle step | None | `cartulary.operation`, `cartulary.result` | `cartulary.error_code` | Connection ID, user ID, incident ID, payload content | Result table below | Disabled |
| WebSocket event send | `cartulary.collaboration.event_send` | `internal` | Before one server-to-client event send attempt | After send, drop, or rejection | Child of current server context when present; otherwise root server-owned span | None | `cartulary.websocket.event_type`, `cartulary.result` | `cartulary.drop_reason` | Event payload, record ID, user ID, connection ID | Result table below | Disabled |
| Job enqueue | `cartulary.jobs.enqueue` | `internal` | Before durable job admission attempt | After accepted, rejected, or failed admission | Child of current HTTP span when present; otherwise root server-owned span | None | `cartulary.job_kind`, `cartulary.operation='enqueue'`, `cartulary.result` | `cartulary.error_code` | Job ID, incident ID, request body | Result table below | Disabled |
| Job run | `cartulary.jobs.run` | `internal` | When worker begins job execution | After terminal job state is committed | Root server-owned job span | None in this implementation plan | `cartulary.job_kind`, `cartulary.job_terminal_status`, `cartulary.result` | `cartulary.error_code` | Job ID, artifact path, incident ID, evidence ID | Result table below | Disabled |
| Postgres dependency | `cartulary.postgres.operation` | `client` | Before acquiring connection or issuing safe operation wrapper | After operation returns or timeout/error class is assigned | Child of current operation span when present | None | `db.system.name='postgresql'`, `cartulary.operation`, `cartulary.result` | `cartulary.error_class` | SQL, query summary, bind values, table names, database name, server address, port | Result table below | Disabled |
| Object-store dependency | `cartulary.objectstore.operation` | `client` | Before safe object-store operation wrapper starts | After operation returns or timeout/error class is assigned | Child of current operation span when present | None | `cartulary.operation`, `cartulary.result` | `cartulary.error_class` | Bucket, key, filename, hash, upload ID, copy source, storage ref | Result table below | Disabled |

### Span result and status mapping

| Product or dependency outcome | `cartulary.result` | OTel span status | Error attributes |
| --- | --- | --- | --- |
| Completed intended operation | `success` | Unset or OK according to selected SDK convention | None |
| Request rejected by validation or authorization | `rejected` | Error only when the adopted span registry classifies the rejection as an error; otherwise unset | Safe public `cartulary.error_code` when available |
| Optimistic-concurrency or same-field conflict | `conflict` | Unset unless the route owner classifies it as error | Safe public `cartulary.error_code` when available |
| Canceled product operation or job | `canceled` | Error only for abnormal cancellation | Safe public `cartulary.error_code` when available |
| Timeout | `timeout` | Error | `cartulary.error_class='timeout'` or public timeout error code when safe |
| Telemetry or product item dropped | `dropped` | Unset for intentional telemetry drop; error for product failure only | `cartulary.drop_reason` where metric/log context allows |
| Internal or dependency failure | `failed` | Error | Safe public `cartulary.error_code` or closed `cartulary.error_class` |

### Required work items

1. Implement sampler profile selection and fixed trace-ID corpus tests.
2. Strip inbound remote context before root span creation.
3. Implement required span families using the lifecycle table above.
4. Enforce HTTP, Postgres, and object-store forbidden-attribute checks.
5. Add span-shape tests for each span family and each status mapping row.

### Acceptance gates

| Gate | Required pass condition | Local/imported criterion |
| --- | --- | --- |
| Sampler gate | `sample_ratio=0.0`, `1.0`, and fractional ratios select exact sampler profiles and expose `sampler_profile_review_after='2027-01-01'`. | `OTEL-AC-015` |
| Determinism gate | Fixed server-owned trace-ID corpus verifies deterministic allow/drop behavior for selected SDK version and ratio. | `OTEL-AC-016` |
| Remote-context gate | Inbound trace and Baggage headers do not alter trace IDs, parentage, sampling, attributes, logs, or metrics. | `OTEL-AC-017` |
| Span-shape gate | Every emitted span conforms to registered name, kind, boundary, attributes, and forbidden-attribute rules. | `OTEL-AC-018..020`, `OIP-AC-009` |

## Phase 6: Metrics instrumentation

### Objective

Emit only registered metrics with fixed identity, cumulative temporality, allowed attributes, explicit buckets, and no exemplars.

### Metric registry implementation table

The adopted OTel NLSpec owns metric identity, histogram defaults, temporality, View behavior, overflow handling, exemplar policy, and Metric `Bind` non-adoption.[^16]

| Metric name | SDK instrument kind | Unit | Measurement source | Sync/async | Collection cadence | Aggregation | Temporality | Bucket family | Allowed attributes | Overflow behavior |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `cartulary.http.server.request.duration` | Histogram | `s` | HTTP server request lifecycle duration | Sync | Per observation | Explicit bucket histogram | Cumulative | Duration | `http.request.method`, `http.response.status_code`, `http.route`, `cartulary.route_family`, `cartulary.result`, optional `cartulary.error_code` | Fail corpus on `otel.metric.overflow=true`; record metric-overflow drop when non-recursive |
| `cartulary.workbook.query.duration` | Histogram | `s` | Workbook query duration | Sync | Per query | Explicit bucket histogram | Cumulative | Duration | `cartulary.view_schema_id`, `cartulary.operation`, `cartulary.result`, optional `cartulary.error_code` | Same |
| `cartulary.workbook.mutation.duration` | Histogram | `s` | Workbook mutation duration | Sync | Per mutation | Explicit bucket histogram | Cumulative | Duration | `cartulary.view_schema_id`, `cartulary.record_type`, `cartulary.operation`, `cartulary.result`, optional `cartulary.error_code` | Same |
| `cartulary.workbook.rows.returned` | Histogram | `{row}` | Returned row count | Sync | Per query | Explicit registry histogram | Cumulative | Count | `cartulary.view_schema_id`, `cartulary.result` | Same |
| `cartulary.collaboration.connections.active` | ObservableGauge | `{connection}` | Accepted WebSocket connection count | Async | Each metric collection | Last value | Cumulative-equivalent | None | No per-connection labels | Same |
| `cartulary.collaboration.events.sent` | Counter | `{event}` | Server-to-client event sends | Sync | Per send/drop | Monotonic sum | Cumulative | None | `cartulary.websocket.event_type`, `cartulary.result`, optional `cartulary.drop_reason` | Same |
| `cartulary.jobs.active` | ObservableGauge | `{job}` | Active background jobs by kind | Async | Each metric collection | Last value | Cumulative-equivalent | None | `cartulary.job_kind` | Same |
| `cartulary.jobs.duration` | Histogram | `s` | Background job runtime | Sync | Per terminal job | Explicit bucket histogram | Cumulative | Duration | `cartulary.job_kind`, `cartulary.job_terminal_status`, `cartulary.result`, optional `cartulary.error_code` | Same |
| `cartulary.postgres.operation.duration` | Histogram | `s` | Postgres dependency operation duration | Sync | Per operation | Explicit bucket histogram | Cumulative | Duration | `db.system.name`, `cartulary.operation`, `cartulary.result`, optional `cartulary.error_class` | Same |
| `cartulary.objectstore.operation.duration` | Histogram | `s` | Object-store operation duration | Sync | Per operation | Explicit bucket histogram | Cumulative | Duration | `cartulary.operation`, `cartulary.result`, optional `cartulary.error_class` | Same |
| `cartulary.objectstore.transfer.bytes` | Histogram | `By` | Safe object-store transfer size | Sync | Per transfer | Explicit bucket histogram | Cumulative | Byte | `cartulary.operation`, `cartulary.result` | Same |
| `cartulary.telemetry.export.failure` | Counter | `{failure}` | Export failures by signal and exporter kind | Sync | Per failure | Monotonic sum | Cumulative | None | `cartulary.signal_kind`, `cartulary.telemetry.exporter_kind`, `cartulary.error_class` | Same |
| `cartulary.telemetry.item.dropped` | Counter | `{item}` | Pre-recording or export-time telemetry drops | Sync | Per drop | Monotonic sum | Cumulative | None | `cartulary.signal_kind`, `cartulary.drop_reason` | Same |
| `cartulary.telemetry.queue.depth` | ObservableGauge | `{item}` | Current processor queue depth by signal | Async | Each metric collection | Last value | Cumulative-equivalent | None | `cartulary.signal_kind` | Same |

Default duration histogram buckets in seconds are `0.005`, `0.010`, `0.025`, `0.050`, `0.100`, `0.250`, `0.500`, `1.000`, `2.500`, `5.000`, `10.000`, and `30.000`. Default byte buckets are `1024`, `4096`, `16384`, `65536`, `262144`, `1048576`, `4194304`, `16777216`, `67108864`, and `268435456`.[^16]

### Required work items

1. Register only the metric instruments listed above.
2. Add identity validation for case-insensitive duplicate names and same-name/same-kind/same-unit/different-description divergence.
3. Configure cumulative temporality for every current-profile instrument.
4. Install Views or equivalent attribute filters for every metric.
5. Disable exemplars and assert environment variables cannot re-enable them.
6. Prohibit Metric `Bind` or assert bound attributes cannot bypass validation.

### Acceptance gates

| Gate | Required pass condition | Imported criterion |
| --- | --- | --- |
| Registry gate | Every emitted metric is byte-equivalent after canonical normalization to the exported metric-stream identity registry for name, instrument kind, unit, aggregation, temporality, resource, scope, and allowed attributes. | `OTEL-AC-021` |
| Identity gate | Duplicate metric-name and divergent-description tests reject invalid registrations. | `OTEL-AC-022` |
| View gate | Views cannot export unregistered streams or widen attributes. | `OTEL-AC-023` |
| Exemplar gate | No exemplars, exemplar trace IDs, span IDs, or filtered attributes appear. | `OTEL-AC-024` |
| Overflow gate | No `otel.metric.overflow=true` datapoint appears except explicit negative test; negative test records metric-overflow drop when non-recursive. | `OTEL-AC-025` |
| Temporality gate | All emitted metrics are cumulative. | `OTEL-AC-026` |
| Bind gate | Metric `Bind` or pre-bound attributes are absent or cannot bypass validation. | `OTEL-AC-026A` |

## Phase 7: Logs and correlation

### Objective

Allow safe local trace correlation and optional OTel LogRecord export without raw exception or incident-content leakage.

### Local log correlation field table

The adopted OTel NLSpec permits only safe trace correlation fields and closed Cartulary fields in local structured logs.[^17]

| Field | Source | Requiredness | Exported to OTel logs? | Forbidden derivations |
| --- | --- | --- | --- | --- |
| `trace_id` | Current server-owned trace ID | Optional | Yes, when bridge enabled and present | Inbound remote context |
| `span_id` | Current server-owned span ID | Optional | Yes, when bridge enabled and present | Inbound remote context |
| `trace_flags` | Current server-owned trace flags | Optional | Yes, sampled flag only | Vendor flags or inbound flags |
| `cartulary.module` | Registry value | Optional | Yes | User, incident, record, or free text |
| `cartulary.route_family` | Registry value | Optional | Yes | Raw route path or parameter |
| `cartulary.operation` | Registry value | Optional | Yes | User-provided text |
| `cartulary.result` | Registry value | Optional | Yes | Raw error message |
| `cartulary.error_code` | Public error-code token | Optional | Yes | Private server detail |
| `cartulary.error_class` | Closed low-cardinality class | Optional | Yes | Exception message, SQL, path, object key |

### OTel LogRecord mapping table

| LogRecord field | Mapping | Requiredness | Bounds | Forbidden values | Disabled-bridge behavior |
| --- | --- | --- | --- | --- | --- |
| `Timestamp` | Local log event timestamp when available | Optional | Timestamp value only | Incident or user content | No LogRecord exported |
| `ObservedTimestamp` | Time bridge observes the log | Required when exported | Timestamp value only | Incident or user content | No LogRecord exported |
| `TraceId` | Current server-owned trace ID | Optional | Trace ID shape | Inbound remote trace ID | No LogRecord exported |
| `SpanId` | Current server-owned span ID | Optional | Span ID shape | Inbound remote span ID | No LogRecord exported |
| `TraceFlags` | Current server-owned trace flags | Optional | Sampled flag only | Vendor flags | No LogRecord exported |
| `SeverityNumber` | Severity mapping below | Required when exported | Closed numeric table | Arbitrary numeric severity | No LogRecord exported |
| `SeverityText` | Severity mapping below | Required when exported | Maximum 32 ASCII characters | Arbitrary local severity text outside mapping | No LogRecord exported |
| `Body` | String-only, redacted before construction | Required when exported; may be empty string | Maximum `telemetry.logs.body_max_chars` Unicode scalar values | Exception message, stacktrace, request body, SQL, paths, object keys, incident text | No LogRecord exported |
| `Resource` | Closed resource registry | Required when exported | Registry values only | Host/process/container/cloud extras | No LogRecord exported |
| `InstrumentationScope` | Registered scope | Required when exported | Scope registry only | Ad hoc scopes | No LogRecord exported |
| `Attributes` | Attributes allowed by attribute registry and local log field table | Optional | Registry value shapes only | Unknown attributes or forbidden values | No LogRecord exported |
| `EventName` | Omitted | Forbidden | Not applicable | Any value | No LogRecord exported |

### Severity mapping

| Local severity | OTel `SeverityNumber` | `SeverityText` |
| --- | ---: | --- |
| `trace` | 1 | `TRACE` |
| `debug` | 5 | `DEBUG` |
| `info` | 9 | `INFO` |
| `warn` | 13 | `WARN` |
| `error` | 17 | `ERROR` |
| `fatal` | 21 | `FATAL` |
| unknown or omitted | 9 | `INFO` |

A local severity string outside this table MUST map to unknown or omitted behavior. It MUST NOT be exported as arbitrary `SeverityText`.[^17]

### Required work items

1. Add safe local structured-log correlation fields.
2. Keep the OTel log bridge disabled by default.
3. Implement the enabled LogRecord mapping table.
4. Reduce exception facts to safe low-cardinality class fields only.
5. Do not install a log-event-to-span-event bridge.

### Acceptance gates

| Gate | Required pass condition | Imported criterion |
| --- | --- | --- |
| Disabled bridge gate | No OTel LogRecords export when bridge is disabled. | `OTEL-AC-027` |
| Mapping gate | Enabled bridge emits only approved top-level fields. | `OTEL-AC-028` |
| Exception gate | Exception message, stacktrace, cause chain, panic string, SQL snippets, paths, object keys, request bodies, and incident text are absent. | `OTEL-AC-029`, `OTEL-AC-030` |
| Span-event gate | Logs do not create span events. | `OTEL-AC-031` |

## Phase 8: Exporter, processor, retry, and shutdown

### Objective

Make telemetry export explicit, bounded, non-blocking, and deterministic enough to test.

### Endpoint normalization algorithm

The adopted OTel NLSpec forbids default exporter endpoints and per-signal divergence, and it defines deterministic OTLP endpoint behavior.[^18]

1. If `telemetry.exporter.kind='none'`, the implementation MUST create no exporter.
2. If `telemetry.exporter.kind='otlp_http'`, parse `telemetry.exporter.endpoint` as a base HTTP or HTTPS URL.
3. Reject userinfo, query, and fragment during configuration validation.
4. Normalize the base path by removing exactly one trailing `/` for path-join purposes.
5. Append `/v1/traces`, `/v1/metrics`, or `/v1/logs` to the normalized base path.
6. Preserve scheme, host, port, and any non-root base path.
7. If `telemetry.exporter.kind='otlp_grpc'`, use the configured endpoint as one gRPC target and do not derive per-signal endpoints.

| Configured endpoint | Trace URL | Metrics URL | Logs URL |
| --- | --- | --- | --- |
| `http://collector:4318` | `http://collector:4318/v1/traces` | `http://collector:4318/v1/metrics` | `http://collector:4318/v1/logs` |
| `http://collector:4318/` | `http://collector:4318/v1/traces` | `http://collector:4318/v1/metrics` | `http://collector:4318/v1/logs` |
| `http://collector:4318/otel` | `http://collector:4318/otel/v1/traces` | `http://collector:4318/otel/v1/metrics` | `http://collector:4318/otel/v1/logs` |
| `http://collector:4318/otel/` | `http://collector:4318/otel/v1/traces` | `http://collector:4318/otel/v1/metrics` | `http://collector:4318/otel/v1/logs` |

### Header and User-Agent contract

| Element | Required behavior |
| --- | --- |
| Exporter request headers | Only `telemetry.exporter.headers` plus protocol-required safe headers. |
| Header values | Secret-bearing until verified otherwise; redacted from logs, spans, metrics, diagnostics, retained artifacts, and test reports. |
| User-Agent segment 1 | `Cartulary/<service.version>` when known; otherwise `Cartulary/0.0.0+unknown`. |
| User-Agent segment 2 | `OTel-OTLP-Exporter-<language>/<version>` when SDK/exporter exposes this identity. |
| Forbidden User-Agent material | Incident identifiers, deployment hostnames, usernames, object-store identifiers, environment names, runtime roots, local paths, or arbitrary operator-provided text. |

### Retry algorithm

Exporter retry behavior MUST follow this deterministic envelope.[^18]

| Retry element | Required rule |
| --- | --- |
| First attempt | No retry delay; bounded by `telemetry.processor.export_timeout_ms`. |
| First retry index | `1`. |
| Base interval | `base_interval_ms = min(max_interval_ms, initial_interval_ms * multiplier^(retry_attempt_index - 1))`. |
| Jitter | Uniform integer milliseconds in `[0, base_interval_ms]`. |
| Runtime randomness | Process CSPRNG or equivalent strong source. |
| Test behavior | Bounds-only by default; deterministic RNG hook optional. |
| Max elapsed cutoff | Do not start retry if `elapsed_since_first_failed_attempt_start + sampled_delay_ms > max_elapsed_ms`. |
| `max_elapsed_ms=0` | Disables retries. |
| Shutdown | No new retry loops after shutdown begins; in-progress retry may finish only inside flush timeout. |
| Permanent rejection | No retry; drop and record `exporter_permanent_discard` when non-recursive. |
| Timeout | Attempt aborts at export timeout; classify transient only when transport or status family is transient. |

### Runtime telemetry failure matrix

| Error point | Product behavior | Telemetry behavior | Required evidence |
| --- | --- | --- | --- |
| Invalid deployment telemetry config | Fail startup before readiness | Startup diagnostics only | Config failure fixture |
| SDK provider construction failure after valid config | Fail startup before readiness | Local startup diagnostics only | Provider-construction fixture |
| Exporter endpoint unavailable | Product remains available | Retry/drop according to retry contract | Runtime-invariance fixture |
| Processor queue full | Product remains available | Drop new telemetry item and increment drop metric when non-recursive | Overflow fixture |
| Redaction cannot establish safety | Product remains available unless product operation failed independently | Drop telemetry item and increment drop metric when non-recursive | Redaction rejection fixture |
| Log bridge mapping failure | Product remains available | Drop LogRecord and increment drop metric when non-recursive | Log mapping failure fixture |
| Shutdown flush timeout | Continue shutdown after timeout | Increment drop metric or local diagnostic when non-recursive and possible | Shutdown fixture |
| Telemetry self-diagnostic error | Product remains available | Suppress recursive telemetry | Recursion fixture |

Runtime telemetry failures MUST NOT alter product route results, workbook mutation commit behavior, WebSocket event-send behavior, evidence-handle issuance and redemption behavior, or background-job state transitions relative to the same scenario with telemetry export disabled.[^18]

### Required work items

1. Implement OTLP/HTTP and OTLP/gRPC endpoint construction exactly.
2. Implement header redaction and User-Agent grammar.
3. Implement bounded full-jitter retry and bounds-only tests by default.
4. Enforce processor queue bounds with `drop_new`.
5. Implement graceful shutdown flush timeout and idempotent shutdown.
6. Add telemetry-about-telemetry recursion guard.
7. Add black-box runtime-invariance tests across product surfaces.

### Acceptance gates

| Gate | Required pass condition | Imported criterion |
| --- | --- | --- |
| Endpoint gate | OTLP/HTTP and OTLP/gRPC tests assert exact endpoint behavior and no per-signal divergence. | `OTEL-AC-032`, `OTEL-AC-033` |
| Header gate | Secret headers are absent from every diagnostic and retained artifact. | `OTEL-AC-034` |
| User-Agent gate | User-Agent contains only allowed identity segments and no forbidden value families. | `OTEL-AC-035` |
| Retry gate | Retry tests cover full-jitter bounds, max elapsed cutoff, disabled retry, permanent rejection, timeout, shutdown, and non-blocking behavior. | `OTEL-AC-036` |
| Overflow gate | Processor overflow uses `drop_new` and does not block product work. | `OTEL-AC-037` |
| Shutdown gate | Flush respects timeout and shutdown continues after timeout. | `OTEL-AC-038` |
| Recursion gate | Self-diagnostics do not recurse unboundedly. | `OTEL-AC-039` |
| Runtime-invariance gate | Product responses and committed state match no-export baseline under runtime telemetry failures. | `OTEL-AC-039A` |

## Phase 9: Browser boundary

### Objective

Prevent browser code from becoming a telemetry exporter or configuration source.

### Browser telemetry boundary verification contract

The adopted OTel NLSpec forbids browser direct telemetry export and states that browser state cannot configure telemetry exporters, headers, endpoints, samplers, processors, metric readers, resource attributes, or log bridges.[^19]

| Verification area | Required input | Algorithm | Failure condition | Allowed override |
| --- | --- | --- | --- | --- |
| Source imports | Frontend source files | Static import scan against forbidden package/module registry | Forbidden import outside test fixtures | None |
| Package manifests | `package.json`, workspace manifests, lockfiles | Dependency graph scan for forbidden direct and transitive browser runtime packages | Forbidden runtime dependency reachable from browser bundle | Plan revision required |
| Built bundle | Production browser bundles | Text and module graph scan after minification | OTLP exporter, vendor SDK, Collector client, session replay, analytics initialization, remote log/metric/trace exporter present | Only documented false positive with artifact hash and no executable import path |
| Dynamic imports | Source and build graph | Treat all browser-reachable dynamic imports as bundle members | Forbidden module reachable dynamically | None |
| Source maps | Source maps when emitted | Scan original source names and module IDs | Forbidden module hidden by minification | None |
| Runtime probes | Browser E2E harness | Assert no telemetry export globals, no remote telemetry requests, no browser config influence | Any network request or runtime config path to telemetry exporter | None |

### Forbidden package and symbol registry

| Family | Forbidden browser runtime content |
| --- | --- |
| OTLP exporters | Any OTLP HTTP/gRPC/protobuf exporter initialized or bundled for browser runtime |
| OpenTelemetry SDK providers | Browser code constructing SDK providers, processors, readers, exporters, samplers, or resources |
| Vendor telemetry SDKs | Vendor-native telemetry clients or hosted analytics SDK initialization |
| Collector clients | Any browser-side Collector client or direct Collector sender |
| Session replay SDKs | Any session replay package or initialization path |
| Third-party analytics | Any analytics initialization sending to third-party endpoints |
| Remote log exporters | Browser-side log export to remote destination |
| Remote metrics exporters | Browser-side metric export to remote destination |
| Remote trace exporters | Browser-side trace export to remote destination |
| Browser telemetry configuration | Endpoint, header, sampler, processor, resource, metric-reader, or log-bridge config through local storage, session storage, IndexedDB, Service Workers, cookies, DOM, React props, grid coordinates, or URL parameters |

### Required work items

1. Add source import, package graph, built-bundle, dynamic-import, source-map, and runtime-probe checks.
2. Fail closed when a required artifact cannot be inspected.
3. Allow local performance marks only when they remain same-origin, contain no incident content or identifiers, do not export directly, and do not persist across sessions unless a later revision defines persistence.
4. Add non-transfer tests proving forbidden OTel concepts remain absent.

### Acceptance gates

| Gate | Required pass condition | Local/imported criterion |
| --- | --- | --- |
| Bundle gate | Browser artifacts contain no forbidden telemetry exporter or vendor SDK families. | `OTEL-AC-040`, `OIP-AC-013` |
| Config gate | Browser state cannot configure telemetry export or provider behavior. | `OTEL-AC-041`, `OIP-AC-013` |
| Non-transfer gate | Collector requirement, Collector-side privacy, Baggage, environment autoconfig, declarative providers, default OTLP localhost, per-signal exporters, Prometheus, Zipkin, Jaeger-native, vendor-native, SQL commenter, resource detectors, Resource schema-url merge, S3 semantic attributes, exemplars, Metric Bind bypass, EventName, Logs API Exception parameter, log-to-span bridge, and ProbabilitySampler remain absent. | `OTEL-AC-042` |

## Phase 10: Golden corpus and harness integration

### Objective

Make telemetry conformance testable through the adopted harness surface.

### Harness contract

The adopted testing harness defines `otel-conformance` as a public active Make target with command ID `cartulary.harness.command.otel_conformance.v1`, output class `summary_with_artifacts`, summary schema `cartulary.otel_conformance_summary.v1`, and retained-artifact behavior. It validates source snapshot, generated constants evidence, emitted telemetry goldens, browser non-export, retained raw capture policy, and telemetry security boundaries.[^20]

Retained raw telemetry captures from `otel-conformance` MUST be target-owned artifacts below the normalized run root and MUST NOT be written below committed golden directories such as `internal/testutil/golden/otel/`.[^20]

### Golden corpus directory contract

| Path | Contents | Mutable by ordinary tests? | Committed? |
| --- | --- | ---: | ---: |
| `internal/testutil/golden/otel/corpus_manifest.json` | Corpus case registry and schema version | No | Yes |
| `internal/testutil/golden/otel/cases/<case_id>/input.json` | Declared stimulus metadata | No | Yes |
| `internal/testutil/golden/otel/cases/<case_id>/normalized_traces.json` | Normalized trace output | No | Yes |
| `internal/testutil/golden/otel/cases/<case_id>/normalized_metrics.json` | Normalized metric output | No | Yes |
| `internal/testutil/golden/otel/cases/<case_id>/normalized_logs.json` | Normalized log output | No | Yes |
| `.cartulary/test-results/<run-id>/...` | Raw captures, diffs, summaries | Yes | No |

### Canonical JSON rules

| Rule | Requirement |
| --- | --- |
| Encoding | UTF-8 without BOM |
| Line endings | LF |
| Object members | Sorted by Unicode code point of member name |
| Arrays | Sorted only where normalization table states deterministic sort; otherwise preserve semantic order |
| Numbers | Decimal JSON numbers; no NaN or Infinity |
| Timestamps | Placeholder tokens only after forbidden-literal precheck |
| Unknown fields | Fail unless registry and change-classification allow them |
| File termination | Exactly one trailing LF |

### Placeholder token registry

| Volatile family | Token form | Preservation rule |
| --- | --- | --- |
| Trace IDs | `TRACE_ID_<ordinal>` | Preserve parent/link topology |
| Span IDs | `SPAN_ID_<ordinal>` | Preserve parent/link topology |
| Timestamps | `TS_<ordinal>` | Preserve event order |
| `service.instance.id` | `SERVICE_INSTANCE_ID_1` | Validate opacity before replacement |
| Durations | Not replaced unless wall-clock-only | Preserve metric bucket membership and asserted numeric measurements |

The adopted OTel NLSpec requires normalization to replace only volatile timestamps and identifiers while preserving shape facts, sorting attributes/resources/spans/metrics/logs deterministically, failing on forbidden values before normalization, and failing on unknown fields unless the owner registry permits them.[^21]

### Corpus fixture set

| Fixture ID | Corpus area | Required coverage |
| --- | --- | --- |
| `OTEL-CORPUS-001` | No-SDK mode | HTTP, workbook, job, Postgres, object-store, log-call, and WebSocket paths execute without SDK providers and without telemetry-induced errors |
| `OTEL-CORPUS-002` | Source baseline | Immutable OTel and semantic-convention refs, full SHAs, source paths, document statuses, model digest, generated constants, SDK versions |
| `OTEL-CORPUS-003` | Cartulary environment binding parser | Omitted, empty, valid, invalid, and explicit `"null"` fixtures for every parser family |
| `OTEL-CORPUS-004` | Hostile environment | Every environment family in hazard registry set to hostile value and no effect |
| `OTEL-CORPUS-005` | Hostile declarative config | Exporter, resource detector, sampler, metric reader, View, exemplar, log bridge, plugin-provider attempts have no effect |
| `OTEL-CORPUS-006` | HTTP route shape | Route-template spans and metrics without raw paths, queries, headers, or IDs |
| `OTEL-CORPUS-007` | Workbook query and mutation | Query, create, patch, conflict, projection update, and row-refresh telemetry with safe attributes only |
| `OTEL-CORPUS-008` | WebSocket collaboration | Connect, authorize, subscribe, event-send, replay, close, overflow, unauthorized subscribe |
| `OTEL-CORPUS-009` | Background jobs | Enqueue, run success, cancellation, failure, timeout, terminal-state metrics |
| `OTEL-CORPUS-010` | Postgres dependency | Success, timeout, unavailable, serialization conflict, constraint violation |
| `OTEL-CORPUS-011` | Object storage | Upload target, attach, preview/download handle issuance, unavailable, transfer size, forbidden identifiers |
| `OTEL-CORPUS-012` | Resource identity | Closed resource, empty schema URL, default/detector suppression, conflicting schema URL rejection |
| `OTEL-CORPUS-013` | Attribute boundary | Null-like values omitted before OTel API calls across all signal families |
| `OTEL-CORPUS-014` | Logs | Bridge disabled/enabled, severity mapping, forbidden exception values, omitted EventName, no optional Exception, no span-event bridge |
| `OTEL-CORPUS-015` | Metrics | Registration identity, stream identity, duplicate rejection, temporality, View rejection, filters, no Bind bypass, no exemplars, no overflow |
| `OTEL-CORPUS-016` | Exporter | Export disabled, endpoint construction, headers, User-Agent, retry, timeout, shutdown |
| `OTEL-CORPUS-017` | Runtime invariance | HTTP request, workbook query, workbook mutation, WebSocket send, evidence access, and job transition match no-export baseline under failure modes |
| `OTEL-CORPUS-018` | Redaction | Representative forbidden values from every family across spans, metrics, logs, self-diagnostics, retained artifacts, exporter attempts |

### Dependency update gate

Any OTel API, SDK, exporter, semantic-convention constants, resource detector package, log bridge package, instrumentation adapter, metric View, sampler, generated constant, or retry-behavior update MUST run the golden corpus and produce one of the adopted change classifications before merge.[^21]

| Change classification | Merge behavior |
| --- | --- |
| `registry_equivalent` | May merge without NLSpec revision if all existing criteria pass unchanged. |
| `additive_non_breaking` | Requires NLSpec revision and added criteria proving intentional privacy-safe addition. |
| `privacy_tightening` | Requires NLSpec revision and criteria proving removed element is absent and required operational questions retain coverage. |
| `breaking_shape_change` | Requires NLSpec revision and updated criteria for old-shape absence and new-shape presence. |

### Required work items

1. Implement `make otel-conformance` using the public harness registry.
2. Emit and validate `cartulary.otel_conformance_summary.v1` before target success.
3. Store raw captures only under the normalized run root.
4. Store normalized goldens only under `internal/testutil/golden/otel`.
5. Implement canonical normalization and deterministic diff output.
6. Add dependency-update classification gate.

### Acceptance gates

| Gate | Required pass condition | Local/imported criterion |
| --- | --- | --- |
| Harness gate | `make otel-conformance` is present in the public registry and produces the required summary schema. | `OIP-AC-014` |
| Artifact gate | Raw captures are retained under run root; normalized goldens remain committed separately. | `OIP-AC-014` |
| Corpus gate | Corpus covers every fixture ID in `OTEL-CORPUS-001` through `OTEL-CORPUS-018`. | `OTEL-AC-043`, `OIP-AC-014` |
| Drift gate | Dependency-only updates are accepted without NLSpec revision only when normalized corpus output is `registry_equivalent`; other shape changes require NLSpec revision. | `OTEL-AC-044..046` |

## Phase 11: Release readiness

### Objective

Treat telemetry as complete only when it is safe, testable, non-interfering, and free of blocking adoption gaps.

### Readiness-state table

| State | Entry condition | Export allowed? | Conformance claim allowed? |
| --- | --- | ---: | ---: |
| `otel_not_adopted` | OTel NLSpec not adopted | No | No |
| `otel_adopted_not_integrated` | OTel NLSpec adopted; source/config/accessor work incomplete | No | No |
| `otel_local_capture_only` | Phases 0-7 pass; Phase 8 not complete | No network export | No |
| `otel_exporter_under_test` | Phase 8 passes; browser and corpus gates incomplete | Harness only | No |
| `otel_release_ready` | Phases 0-11 pass, all `OIP-AC-*` and imported `OTEL-AC-*` mapped criteria pass, and no blocking TODO remains | Yes, by explicit valid config | Yes |

### Acceptance mapping table

| Local criterion | Imported criterion | Required evidence |
| --- | --- | --- |
| `OIP-AC-015` | `OTEL-AC-001..006` | Source baseline, config, hostile environment/config, export-disabled tests |
| `OIP-AC-016` | `OTEL-AC-007..009` | Static import checks, no-SDK runtime tests, scope tests |
| `OIP-AC-017` | `OTEL-AC-010..014A` | Resource, detector, forbidden-value, attribute, null-omission tests |
| `OIP-AC-018` | `OTEL-AC-015..020` | Sampler, remote-context, HTTP, Postgres, object-store span tests |
| `OIP-AC-019` | `OTEL-AC-021..026A` | Metric identity, View, exemplar, overflow, temporality, Bind tests |
| `OIP-AC-020` | `OTEL-AC-027..031` | Log disabled/enabled mapping, exception, EventName, span-event tests |
| `OIP-AC-021` | `OTEL-AC-032..039A` | Endpoint, gRPC, headers, User-Agent, retry, overflow, shutdown, self-diagnostics, runtime-invariance tests |
| `OIP-AC-022` | `OTEL-AC-040..042` | Browser boundary and non-transfer tests |
| `OIP-AC-023` | `OTEL-AC-043..046` | Golden normalization, dependency update, shape-classification tests |

### Definition of done

Implementation is done only when all of the following are true:

- Source snapshots, semantic-convention model digest, generated constants, and SDK package versions are pinned.
- Config parsing is closed and fail-closed.
- External OTel environment and declarative configuration cannot affect behavior.
- Ordinary instrumentation is API-only and no-SDK safe.
- Resource identity, attributes, forbidden values, and redaction-before-recording are closed.
- Traces, metrics, and logs emit only registered shapes.
- Exporters are explicit, bounded, non-blocking, and redacted.
- Browser direct export and browser configuration authority are absent.
- Raw telemetry captures and normalized goldens are separated.
- Dependency-update classification is enforced.
- `make otel-conformance` passes.
- Every local `OIP-AC-*` and imported `OTEL-AC-*` criterion mapped by this plan passes.
- No blocking `TODO(repo-adoption)` appears in a conformance-visible artifact, registry, fixture, phase gate, or acceptance mapping.

The adopted OTel NLSpec completion standard requires pinned source snapshots and package versions, API-only instrumentation, no-SDK safety, closed configuration, environment containment, closed resources, null omission, redaction-before-recording, registered signals, fixed metrics, disabled default logs, explicit OTLP export, bounded retry, runtime invariance, browser non-export, golden corpus comparison, and binary acceptance criteria.[^22]

## Acceptance criteria registry

| ID | Verifies | Required evidence | Pass condition |
| --- | --- | --- | --- |
| `OIP-AC-001` | Source-limit closure | Revised source-limit section | No conformance-visible behavior depends on unseen live-repo convention. |
| `OIP-AC-002` | Conformance modes | Status manifest and harness summary | Every OTel status is one of the four closed modes with required config, export, harness, and claim behavior. |
| `OIP-AC-003` | Decision closure | `OTEL-DQ-*` registry artifacts | No decision row is unresolved in `adopted_conformant`. |
| `OIP-AC-004` | Source baseline and generated constants | Snapshot and generated-constant manifests | No `main`, short SHA, placeholder digest, missing source path, missing status, or missing SDK package version appears. |
| `OIP-AC-005` | Config namespace closure | Generated schema and config tests | Every `telemetry.*` key has type, default, bound, omission rule, explicit-null rule, empty-env rule, secret status, and failure behavior. |
| `OIP-AC-006` | Cross-key and hazard closure | Config matrix and hazard fixtures | Every cross-key rule and OTel environment hazard family has explicit fixture and forbidden-effect assertion. |
| `OIP-AC-007` | Accessor interface closure | Static import checks, no-SDK tests, scope tests | Accessor input, output, unknown-scope behavior, no-SDK behavior, and concurrency behavior are implemented. |
| `OIP-AC-008` | Forbidden-value action closure | Redaction action tests | Each forbidden family deterministically omits, replaces with closed class, or drops before recording, and records drop metric when required and non-recursive. |
| `OIP-AC-009` | Span lifecycle closure | Span registry tests | Every span family has deterministic name, kind, lifecycle boundary, status rule, parent rule, link rule, and allowed/forbidden attribute set. |
| `OIP-AC-010` | Metric implementation closure | Metric registry tests | Every metric row states instrument kind, unit, aggregation, temporality, allowed attributes, and overflow behavior. |
| `OIP-AC-011` | Log mapping closure | Log bridge tests | Disabled mode, enabled mapping, severity mapping, body bound, exception reduction, EventName omission, and span-event non-creation are tested. |
| `OIP-AC-012` | Exporter algorithm closure | Exporter, retry, shutdown tests | Endpoint, header, User-Agent, retry, overflow, shutdown, and recursion behavior pass exact predicates. |
| `OIP-AC-013` | Browser boundary closure | Source, package, bundle, source-map, dynamic-import, runtime-probe tests | Browser boundary tests fail closed when any required artifact cannot be inspected. |
| `OIP-AC-014` | Golden corpus closure | Corpus manifest, normalized files, raw retained artifacts | Same input bytes, runtime config, and corpus revision produce byte-identical normalized golden files. |
| `OIP-AC-015` | Config and source-baseline imported criteria | Imported `OTEL-AC-001..006` evidence | All imported criteria pass. |
| `OIP-AC-016` | API, SDK, and instrumentation-boundary criteria | Imported `OTEL-AC-007..009` evidence | All imported criteria pass. |
| `OIP-AC-017` | Resource and attribute criteria | Imported `OTEL-AC-010..014A` evidence | All imported criteria pass. |
| `OIP-AC-018` | Trace and span criteria | Imported `OTEL-AC-015..020` evidence | All imported criteria pass. |
| `OIP-AC-019` | Metrics criteria | Imported `OTEL-AC-021..026A` evidence | All imported criteria pass. |
| `OIP-AC-020` | Logs criteria | Imported `OTEL-AC-027..031` evidence | All imported criteria pass. |
| `OIP-AC-021` | Exporter, processor, runtime, shutdown criteria | Imported `OTEL-AC-032..039A` evidence | All imported criteria pass. |
| `OIP-AC-022` | Browser and non-transfer criteria | Imported `OTEL-AC-040..042` evidence | All imported criteria pass. |
| `OIP-AC-023` | Golden corpus and drift criteria | Imported `OTEL-AC-043..046` evidence | All imported criteria pass. |
| `OIP-AC-024` | Threat-model update | STRIDE threat-model diff or checked-in document | Telemetry exporter config, headers/secrets, retained artifacts, redaction, attribute governance, browser boundary, and runtime failure invariance are covered before release. |
| `OIP-AC-025` | Phase-gate traceability | Phase gates and acceptance mapping | Every phase gate references at least one local `OIP-AC-*` row or imported `OTEL-AC-*` row. |

## Recommended work breakdown

| Slice | Primary deliverable | First tests to write |
| --- | --- | --- |
| `OTEL-1` | `otel_source_snapshot` and generated constants lock | Source snapshot rejects `main`, missing digest, short SHA, missing SDK versions. |
| `OTEL-2` | Closed telemetry config parser | Unknown keys, explicit nulls, invalid enums, invalid endpoint, invalid cross-key combinations. |
| `OTEL-3` | Bootstrap boundary and import guard | Static import checks and no-SDK runtime path tests. |
| `OTEL-4` | Resource and redaction substrate | Resource closure, detector suppression, forbidden-value corpus. |
| `OTEL-5` | Trace registry | HTTP/workbook/WebSocket/jobs/Postgres/object-store span-shape tests. |
| `OTEL-6` | Metric registry and Views | Metric identity, temporality, attribute filter, no exemplars, no Bind bypass. |
| `OTEL-7` | Log bridge | Disabled-default test, enabled mapping test, exception-sanitization test. |
| `OTEL-8` | Exporter and processor | Endpoint construction, headers, User-Agent, retry, queue overflow, shutdown, runtime invariance. |
| `OTEL-9` | Browser boundary | Source import, package graph, bundle, dynamic import, source-map, runtime probe tests. |
| `OTEL-10` | Golden corpus and harness | `otel-conformance` summary schema, normalized-golden comparison, dependency-update classification. |

## Sources

[^1]: `00_document_set_status_and_precedence.md`, §1 Status, §2 Precedence, §5 Document map, and §5.1 Contract-owner matrix, lines 5-18, 22-33, 93-112.
[^2]: `04_security_deployment_and_conformance.md`, §12.1 Scope and owner, lines 1643-1651.
[^3]: `01_architecture_storage_and_view_contracts.md`, §1 Architecture pattern and §2 Required modules and boundaries, lines 3-49; `opentelemetry-instrumentation-nlspec.md`, §3 Non-goals and §5.1 Subsystem boundary, lines 43-59 and 191-212.
[^4]: `opentelemetry-instrumentation-nlspec.md`, §16 Open decisions, lines 1280-1293.
[^5]: `04_security_deployment_and_conformance.md`, §4.4 STRIDE threat model and telemetry update triggers, lines 386-417.
[^6]: `cartulary_repository_bootstrap_guide.md`, Step 1 directory command and OTel repo-control note, lines 63-85.
[^7]: `cartulary-dev-guide.md`, §2.1.1 OpenTelemetry dependency boundary, lines 164-170; `opentelemetry-instrumentation-nlspec.md`, §4.3 OpenTelemetry component boundary, lines 130-154.
[^8]: `opentelemetry-instrumentation-nlspec.md`, §6.1 Configuration keys, lines 244-288.
[^9]: `opentelemetry-instrumentation-nlspec.md`, §6.2.1 Cartulary server-side environment binding parser, lines 304-323; `04_security_deployment_and_conformance.md`, §12.2 Canonical artifact and discovery, lines 1655-1674.
[^10]: `opentelemetry-instrumentation-nlspec.md`, §6.3 OpenTelemetry external configuration containment and §6.4 OpenTelemetry environment hazard registry, lines 325-388.
[^11]: `opentelemetry-instrumentation-nlspec.md`, §5.2 Instrumentation scopes and §5.3 No-SDK instrumentation mode, lines 214-238.
[^12]: `opentelemetry-instrumentation-nlspec.md`, §7.1 Resource attribute registry and §7.2 Resource detector and environment merge policy, lines 423-481.
[^13]: `opentelemetry-instrumentation-nlspec.md`, §8.2 Attribute value-shape policy, §8.4 Forbidden telemetry values and attributes, and §8.7 Redaction-before-recording invariant, lines 501-524, 548-573, and 613-625.
[^14]: `opentelemetry-instrumentation-nlspec.md`, §9.2 Sampler profile and §9.8 Inbound trace context and Baggage, lines 645-687 and 754-760.
[^15]: `opentelemetry-instrumentation-nlspec.md`, §9.1 General span rules, §9.3 Required span families, §9.4 HTTP server standard attribute allowlist, §9.5 Database span details, §9.6 Object-store non-adoption rule, and §9.7 Trace causality and links, lines 631-644, 689-744, and 746-752.
[^16]: `opentelemetry-instrumentation-nlspec.md`, §10 Metrics contract, lines 762-892.
[^17]: `opentelemetry-instrumentation-nlspec.md`, §11 Logs and correlation, lines 894-976.
[^18]: `opentelemetry-instrumentation-nlspec.md`, §12 Exporter, processor, runtime, and shutdown behavior, lines 978-1096.
[^19]: `opentelemetry-instrumentation-nlspec.md`, §13 Browser telemetry boundary and non-transfer rules, lines 1098-1145; `cartulary-dev-guide.md`, §2.1.1 OpenTelemetry dependency boundary, lines 168-170.
[^20]: `testing-harness-nlspec.md`, public target registry, raw telemetry artifact rule, and schema attachment registry, lines 220-229, 547-555, and 688-698.
[^21]: `opentelemetry-instrumentation-nlspec.md`, §14 Golden telemetry corpus and drift verification and §15.9 Golden-corpus and drift criteria, lines 1147-1199 and 1273-1278.
[^22]: `opentelemetry-instrumentation-nlspec.md`, §15 Verification and acceptance criteria and §17 Completion standard, lines 1201-1278 and 1295-1325.

[^23]: `nlspec-spec.md`, “The Nature of the Artifact,” “Behavioral Completeness,” “Unambiguous Interfaces,” “Explicit Defaults and Boundaries,” “Mapping Tables for Translation,” “Testable Acceptance Criteria,” and “Spec Economy,” lines 11-72 and 142-158.
