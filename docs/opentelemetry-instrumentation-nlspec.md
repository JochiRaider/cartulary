---
title: Cartulary OpenTelemetry Instrumentation NLSpec
status: adopted/current
document_class: nlspec
created_at: 2026-05-19
---

## 1. Status, scope, and authority

Status: `adopted/current`.

This NLSpec defines Cartulary's OpenTelemetry instrumentation subsystem. It is adopted implementation-conformance authority for the telemetry subsystem only. It retains OpenTelemetry as the telemetry substrate and closes source baselining, OpenTelemetry configuration containment, semantic-convention drift control, signal-shape determinism, privacy, exporter behavior, and conformance testing.

**OTEL-REQ-001**
This NLSpec governs only telemetry generation, telemetry configuration, telemetry export, log correlation, signal naming, attribute governance, resource identity, privacy boundaries, telemetry runtime behavior, and instrumentation verification.

**OTEL-REQ-002**
This NLSpec MUST NOT redefine product behavior owned by Cartulary Core 00 through Core 04. It MUST NOT redefine claim-bearing benchmark publication owned by Core 05. Runtime telemetry MAY support engineering diagnosis and operational SRE practice, but telemetry observations MUST NOT satisfy claim-bearing timed or fixture-sensitive publication unless Core 05 benchmark-manifest and measurement-predicate requirements are also satisfied.

**OTEL-REQ-003**
When this NLSpec conflicts with Core 00 through Core 04 outside the telemetry subsystem, the conflict is a defect in this NLSpec. When this NLSpec conflicts with non-normative appendices or guides, this NLSpec governs only the telemetry subsystem.

**OTEL-REQ-004**
The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are normative inside this NLSpec. **MUST** and **MUST NOT** define conformance requirements. **SHOULD** and **SHOULD NOT** define strong defaults whose exceptions must remain compatible with all MUST-level requirements. **MAY** defines optional behavior whose omission semantics are explicit.

## 2. Purpose

**OTEL-REQ-005**
Cartulary MUST provide first-class observability because long-term support requires operators to diagnose availability, latency, error, queueing, persistence, evidence access, collaboration, and telemetry-pipeline failures without inspecting incident content, exposing stable incident identifiers, or weakening the workbook hot path.

The instrumentation subsystem MUST make these operational questions answerable:

| Question | Required signal support |
| --- | --- |
| Is the application deployable accepting and completing HTTP requests? | HTTP server spans, bounded HTTP duration metrics, and low-cardinality status classification. |
| Are workbook queries, mutations, and projection updates healthy? | Workbook spans plus duration, row-count, conflict, and result metrics. |
| Are WebSocket subscriptions, presence updates, and live row updates healthy? | WebSocket active gauges, event counters, bounded operation spans, and low-cardinality close or drop classification. |
| Are background jobs queued, running, canceled, failed, or completing? | Job enqueue and run spans, active gauges, terminal duration metrics, and terminal-status attributes. |
| Are Postgres or object-storage dependencies degraded? | Dependency spans, dependency duration metrics, retry/drop counters, and low-cardinality error classification. |
| Are telemetry exporters failing or dropping data? | Telemetry self-metrics and bounded local diagnostics. |
| Can operators correlate safe local logs with traces? | Trace correlation fields, OTel LogRecord mapping when enabled, bounded body rules, and redaction-before-recording rules. |

## 3. Non-goals

**OTEL-REQ-006**
This NLSpec MUST NOT introduce any behavior in the following table:

| Non-goal | Boundary |
| --- | --- |
| A new product workflow | Telemetry MUST NOT add row-edit rituals, approval gates, or user-facing capture friction. |
| A new case-data source of truth | Telemetry MUST NOT become authoritative incident state, audit state, history state, projection state, workflow state, evidence state, or benchmark evidence. |
| A monitoring vendor dependency | The implementation MUST remain vendor-neutral and MUST NOT require Datadog, Grafana, Honeycomb, New Relic, Splunk, Elastic, Jaeger, Prometheus, or any other specific backend. |
| A required OpenTelemetry Collector | A Collector MAY be an external receiver, but it is not a Cartulary deployable and is not required for Cartulary telemetry conformance. |
| A public dashboard contract | Dashboards, alerts, and runbooks MAY be derived later, but this NLSpec owns emitted telemetry, not dashboard layout. |
| Browser-to-third-party telemetry export | Browser code MUST NOT send telemetry directly to an external collector or vendor endpoint. |
| Raw incident-content logging | Incident-authored values, evidence bytes, raw queries, note text, timeline details, filenames, credential material, and object-store keys MUST NOT be exported as telemetry attributes or logs. |
| Environment-driven telemetry egress | OpenTelemetry SDK environment defaults MUST NOT override Cartulary deployment configuration or enable export when Cartulary export is disabled. |
| Collector-side privacy enforcement | A Collector, backend, exporter, or vendor pipeline MUST NOT be required to make emitted telemetry privacy-conformant. |
| OpenTelemetry completeness | Cartulary MUST NOT adopt every OpenTelemetry feature merely because the upstream specification supports it. |

## 4. External standard baseline and source snapshot

OpenTelemetry is the selected telemetry framework because it is vendor-neutral and defines common API, SDK, signal, resource, semantic-convention, and OTLP concepts. OpenTelemetry clients separate API packages used by instrumentation from SDK packages managed by the application owner, and instrumentation authors must not directly reference SDK packages.

### 4.1 Baseline object

**OTEL-REQ-007**
The external standard baseline for this NLSpec MUST be the closed object in this table. Every repo-control `otel_source_snapshot` used for implementation conformance MUST reproduce these values or the rules in the value cell exactly.

| Field | Required value or rule |
| --- | --- |
| `otel_spec_version` | `1.57.0`. |
| `otel_spec_source` | `https://opentelemetry.io/docs/specs/otel/`. |
| `otel_spec_repo` | `https://github.com/open-telemetry/opentelemetry-specification`. |
| `otel_spec_ref` | `v1.57.0`. `main` is forbidden in conformance snapshots. |
| `otel_spec_commit_sha` | `d4a91bddb53b4c308df3e40171a60059183efd88`. |
| `otel_spec_observed_at` | `2026-05-20`. |
| `semconv_version` | `1.41.0`. |
| `semconv_source` | `https://opentelemetry.io/docs/specs/semconv/`. |
| `semconv_repo` | `https://github.com/open-telemetry/semantic-conventions`. |
| `semconv_ref` | `v1.41.0`. `main` is forbidden in conformance snapshots. |
| `semconv_commit_sha` | `e018fe6f91862f5ed63c082f87697cddac596784`. |
| `semconv_model_source` | Semantic-convention YAML model files from the pinned semantic-conventions source, not prose-only extraction. |
| `semconv_model_digest` | `3f8f80a2ed04521dfe29e50fcddd7f7de70145a6aee01959f985a65fbb4c8632`, computed by `semconv_model_digest_v1` in §4.1.1. |
| `semconv_generated_constants` | Generated constants provenance object defined in §4.1.3. |
| `language_sdk_versions` | Exhaustive package-family table defined in §4.1.4. Exact package paths and versions are repo-control facts. |
| `document_status_by_path` | Required for every adopted source path in §4.2. |
| `migration_note_required` | `true` for this revision because the baseline moved from the prior observed 1.56.0 source posture to the 1.57.0 source baseline. For later revisions, derive from §4.5. |

#### 4.1.1 Semantic-convention model digest

**OTEL-REQ-137**
`semconv_model_digest_v1` is the only valid semantic-convention model digest algorithm for this revision. It MUST use the pinned semantic-conventions source identified by `semconv_ref='v1.41.0'` and `semconv_commit_sha='e018fe6f91862f5ed63c082f87697cddac596784'`.

The input tree is the semantic-conventions repository model root. In the pinned upstream repository this root is `model/`; in Cartulary source-family notation the corresponding registry path is `semantic-conventions/model/**`. The digest input MUST include every regular file under that model root.

`semconv_model_digest_v1` MUST reject symlinks, directory entries as digest records, duplicate normalized paths, unreadable files, paths outside the model root, absolute paths, path traversal, and platform-specific path separators. Each digest path MUST be a POSIX path relative to the model root with no leading `./`.

Files MUST be ordered by bytewise ascending UTF-8 normalized path. The hash input MUST be:

```text
cartulary.semconv_model_digest_v1 NUL
for each file:
  normalized_path_bytes NUL decimal_byte_length NUL raw_file_bytes NUL
```

The output MUST be lowercase 64-character SHA-256 hexadecimal. The adopted digest for this revision is `3f8f80a2ed04521dfe29e50fcddd7f7de70145a6aee01959f985a65fbb4c8632`.

#### 4.1.2 Repo-control source snapshot object

**OTEL-REQ-138**
The repo-control `otel_source_snapshot` object MUST contain exactly the top-level fields in this table. Unknown top-level keys are invalid unless a later adopted NLSpec revision declares them.

| Field | Required rule |
| --- | --- |
| `schema_id` | Exactly `cartulary.otel_source_snapshot.v1`. |
| `otel_spec_version` | Exact value from §4.1. |
| `otel_spec_ref` | Exact value from §4.1; `main` is invalid. |
| `otel_spec_commit_sha` | Full 40-character lowercase SHA. |
| `semconv_version` | Exact value from §4.1. |
| `semconv_ref` | Exact value from §4.1; `main` is invalid. |
| `semconv_commit_sha` | Full 40-character lowercase SHA. |
| `semconv_model_digest_algorithm` | Exactly `semconv_model_digest_v1`. |
| `semconv_model_digest` | Concrete lowercase 64-character hex digest from §4.1.1. |
| `semconv_generated_constants` | Object defined in §4.1.3. |
| `language_sdk_versions` | Array defined in §4.1.4. |
| `source_paths` | One row for each §4.2 path with `family`, `path`, `status`, and `source_commit_sha`. |
| `created_at` | RFC3339 timestamp generated by repo-control tooling. |
| `created_by_tool` | Stable tool identifier, not a human name. |

#### 4.1.3 Generated constants provenance

**OTEL-REQ-139**
`semconv_generated_constants` MUST use one of the allowed source kinds in this table.

| `source_kind` | Required fields | Rule |
| --- | --- | --- |
| `repo_codegen` | `generator_name`, `generator_version`, `generator_source_ref`, `generator_source_sha`, `input_model_digest`, `output_package_or_path` | `input_model_digest` MUST equal `semconv_model_digest`. |
| `upstream_package` | `package_path`, `package_version`, `package_checksum_source`, `input_model_digest_or_equivalence_proof` | The equivalence proof MUST bind the package constants to the pinned semantic-convention model digest. |

Ordinary instrumentation code MUST NOT use ad hoc string literals for emitted standard attributes or metric names when a generated or pinned constant exists. Generated-constants drift is a conformance failure unless the normalized golden corpus proves `registry_equivalent` under §4.5.

#### 4.1.4 Language SDK package set

**OTEL-REQ-140**
`language_sdk_versions` MUST be an exhaustive package-family table for every OpenTelemetry package family that can affect emitted telemetry shape or runtime behavior.

Each row MUST declare `package_family` and `status`. Rows with `status='present'` MUST include the required snapshot fields in the table below. Rows with `status='absent'` MUST identify the package family and the repo-control source that proves absence. Rows with `status='hostile_fixture_only'` MUST include the present-package fields plus `fixture_scope`, and runtime import or initialization outside that fixture scope is non-conformant.

| Package family | Required if used? | Required snapshot fields |
| --- | ---: | --- |
| API | Yes | `language`, `package_path`, `version`, `module_or_lockfile_source`. |
| SDK | Yes when SDK providers are installed | Same as above. |
| Trace SDK | Yes when traces enabled | Same as above. |
| Metrics SDK | Yes when metrics enabled | Same as above. |
| Logs SDK or bridge | Yes when log bridge implemented | Same as above. |
| OTLP HTTP exporters | Yes when `otlp_http` supported | Same as above. |
| OTLP gRPC exporters | Yes when `otlp_grpc` supported | Same as above. |
| Semantic-convention constants | Yes | Same as above plus generated-constants binding. |
| Instrumentation adapters | Yes if imported | Same as above. |
| Resource-detector packages | Must be absent unless a future revision adopts them | If present, validation fails in this revision. |
| Autoconfiguration or declarative-config packages | Must be absent unless wrapped as a hostile-fixture test dependency | Runtime use fails conformance. |
| Browser-side OTel packages | Must be absent in the current profile | Any browser exporter, vendor SDK, Collector client, or session replay package fails browser-boundary validation. |

The package-family table MUST make the no-browser-export rule in §13.1 statically checkable against browser bundles and frontend package manifests.

**OTEL-REQ-008**
The baseline versions above are adoption locks, not claims that later OpenTelemetry or semantic-convention releases are incompatible. A later adopted revision MAY rebaseline to newer OpenTelemetry or semantic-convention versions only after applying §4.5 and updating all affected source paths, signal registries, configuration rules, non-transfer rules, golden-corpus expectations, and acceptance criteria.

**OTEL-REQ-121**
A repo-control `otel_source_snapshot` is invalid for conformance if any of the following are true:

| Invalid state | Required result |
| --- | --- |
| `otel_spec_ref` or `semconv_ref` equals `main` | Conformance mode fails before tests are treated as evidence. |
| Any pinned commit SHA is omitted, shortened, malformed, or not 40 lowercase hexadecimal characters | Conformance mode fails before provider bootstrap tests. |
| `semconv_model_digest` is omitted, a placeholder, malformed, or not 64 lowercase hexadecimal characters | Conformance mode fails before generated-constant or emitted-shape tests. |
| `schema_id` is missing or not exactly `cartulary.otel_source_snapshot.v1` | Conformance mode fails before source-baseline tests. |
| `semconv_model_digest_algorithm` is missing or not exactly `semconv_model_digest_v1` | Conformance mode fails before generated-constant or emitted-shape tests. |
| `semconv_generated_constants` is missing, malformed, inconsistent with §4.1.3, or bound to a different model digest | Conformance mode fails before emitted-shape tests. |
| `language_sdk_versions` omits any required package family, package path, version, language, lockfile or module source, or semantic-convention constants binding | Conformance mode fails before emitted-shape tests. |
| Resource-detector, runtime autoconfiguration, runtime declarative-config, browser exporter, vendor telemetry SDK, Collector client, session replay, or browser analytics initialization packages are present outside allowed hostile-fixture scope | Conformance mode fails before runtime tests. |
| The snapshot records a source path not assigned by §4.2 or a later revision | The snapshot fails source-baseline validation unless the later revision assigns the path. |
| Any §4.2 source path is missing, duplicated, or lacks a status row | The snapshot fails source-baseline validation. |
| Any source path row has an unknown status or a source commit SHA different from the owning source family commit | The snapshot fails source-baseline validation. |
| Any top-level value is a placeholder such as `TODO`, `TBD`, `unknown`, an empty string, or a shortened SHA | Conformance mode fails before tests are treated as evidence. |

### 4.2 Source path registry

**OTEL-REQ-009**
The implementation MUST maintain a repo-control `otel_source_snapshot` artifact with the source paths in this table. The artifact MAY add paths only when an NLSpec revision assigns the added path to a source family and change classifier. Wildcard configuration paths are forbidden because configuration source scope affects conformance reproducibility.

| Source family | Required source path or page | Required use |
| --- | --- | --- |
| Specification overview | `specification/overview.md` | API/SDK/component boundary, OpenTelemetry Core-package terminology, signals, spans, links, Baggage, Resources, Collector, instrumentation-library terminology. |
| Configuration overview | `specification/configuration/README.md` | Programmatic, environment-variable, and declarative configuration boundaries. |
| Configuration data model | `specification/configuration/data-model.md` | Declarative configuration object families that Cartulary refuses as runtime authority. |
| Configuration API | `specification/configuration/api.md` | Instrumentation configuration API non-adoption as Cartulary runtime authority. |
| Configuration SDK | `specification/configuration/sdk.md` | File configuration, in-memory configuration, custom plugin component providers, parse/load operations, and non-adoption fixtures. |
| Common configuration parsing | `specification/configuration/common.md` | OTel numeric, duration, timeout, string, and enum parsing reference for containment analysis only. |
| SDK environment variables | `specification/configuration/sdk-environment-variables.md` | Environment hazard registry, default OTel propagators, sampler, exporter, resource, exemplar, declarative-config, and SDK-disabled behavior. |
| Trace SDK | `specification/trace/sdk.md` | Sampler profile, TraceIdRatioBased compatibility, ProbabilitySampler non-adoption, parent-based semantics. |
| Metrics API | `specification/metrics/api.md` | Instrument identity, name syntax, unit length, synchronous/asynchronous instrument families, and Metric `Bind` non-adoption. |
| Metrics SDK | `specification/metrics/sdk.md` | Views, stream names, temporality, cardinality overflow, exemplars, metric readers, shutdown/flush. |
| Logs API | `specification/logs/api.md` | Log bridge, LogRecord emission parameters, optional `Exception` parameter, EventName input. |
| Logs data model | `specification/logs/data-model.md` | LogRecord top-level field mapping, severity mapping, Body, Attributes, EventName. |
| Resource SDK | `specification/resource/sdk.md` | Resource immutability, resource schema URL, resource merge rules, resource detectors, environment-derived resource attributes. |
| Protocol exporter | `specification/protocol/exporter.md` | OTLP endpoint construction, per-signal endpoint precedence, compression, retry, User-Agent. |
| Common concepts | `specification/common/README.md` | Attribute value shapes, AnyValue, attribute limits, uniqueness. |
| Versioning and stability | `specification/versioning-and-stability.md` | Semantic-convention stability and shape-change classification. |
| Semantic conventions model | `semantic-conventions/model/**` | Generated standard constants and standard allowlists. |
| Semantic conventions docs | `semantic-conventions/docs/**` | Human-readable interpretation only. The model files are the source of generated constants. |

### 4.3 OpenTelemetry component boundary

**OTEL-REQ-010**
Cartulary MUST use the component meanings in this table:

| Term | Required Cartulary meaning |
| --- | --- |
| `OpenTelemetry API` | The only OTel package family that ordinary Cartulary instrumentation code may call directly. |
| `OpenTelemetry SDK` | Installed, configured, and shut down only by the server-side telemetry bootstrap boundary. |
| `OpenTelemetry Core package` | An upstream OpenTelemetry specification-defined package or plugin family across API, SDK, exporter, propagator, or related package categories. This term MUST NOT mean Cartulary Core 00, Core 01, Core 02, Core 03, Core 04, or Core 05, and MUST NOT justify ordinary instrumentation units importing SDK, exporter, processor, sampler, propagator, metric-reader, log-processor, autoconfiguration, resource-detector, or plugin-provider packages. |
| `Instrumentation unit` | Cartulary code that records telemetry for one internal module or platform concern through the OTel API. |
| `Instrumentation scope` | The OTel `(name, version, schema_url, attributes)` identity used when obtaining tracers, meters, or loggers for a Cartulary instrumentation unit. |
| `Exporter` | Server-side component that sends telemetry to the configured OTLP endpoint. Exporters are never configured by browser code. |
| `Processor` | Server-side component that batches, bounds, drops, flushes, and forwards telemetry to exporters. |
| `MetricReader` | Server-side metric collection/export boundary owned by telemetry bootstrap. It defines temporality and collection cadence. |
| `Sampler` | Trace decision component owned by telemetry bootstrap and selected only through the sampler profile in §9.2. |
| `Propagator` | Context injection/extraction component. Inbound remote extraction is disabled in this revision. |
| `Collector` | Optional external receiver. It is not a Cartulary deployable and is not required for Cartulary telemetry conformance. |
| `Semantic conventions` | OTel standard names and meanings for common telemetry concepts. They are adopted by stability policy and registry generation, not copied ad hoc. |

**OTEL-REQ-011**
Ordinary instrumentation units MUST obtain tracers, meters, and loggers only through API-facing provider accessors supplied by the telemetry bootstrap boundary. They MUST NOT import, construct, or configure SDK providers, exporters, processors, samplers, propagators, metric readers, log processors, declarative configuration, SDK autoconfiguration, resource detectors, or plugin-provider packages.

**OTEL-REQ-012**
The telemetry bootstrap boundary MAY import SDK packages only for provider setup, configuration validation, resource construction, sampler construction, processor construction, exporter construction, metric-reader construction, shutdown, and bounded self-diagnostics. It MUST NOT make SDK construction or exporter configuration callable from ordinary instrumentation units.

### 4.4 Semantic-convention stability policy and change classification

**OTEL-REQ-013**
Cartulary MUST apply this semantic-convention adoption matrix:

| OTel convention status | Cartulary default |
| --- | --- |
| Stable and applicable | Emit by default only when it does not violate Cartulary privacy, cardinality, configuration, or deployment-boundary rules. |
| Stable but privacy-conflicting | Do not emit the conflicting attribute or telemetry element. Record the omission in the signal-specific non-adoption table. |
| Development or experimental | Do not emit by default. Adoption requires an explicit NLSpec revision or a closed opt-in profile defined by this NLSpec. |
| Deprecated | Do not emit unless a migration-compatibility profile explicitly requires it. |
| Migration-period duplicated conventions | Do not duplicate by default. Duplication requires a bounded compatibility rule and an acceptance criterion proving both old and new forms are intentional. |
| Unknown or unpinned | Do not emit. |

**OTEL-REQ-014**
Every emitted standard attribute and standard metric name MUST be generated or imported from the pinned semantic-convention model source, or MUST be explicitly listed as a standard attribute allowlist exception in the owning signal registry. Every emitted Cartulary custom attribute MUST be listed in §8.5.

**OTEL-REQ-015**
Cartulary is stricter than OpenTelemetry semantic-convention stability. Even when OpenTelemetry treats an additive standard attribute, metric, event, or signal element as compatible, Cartulary MUST classify it by §4.5 and MUST NOT silently widen emitted telemetry.

### 4.5 Telemetry shape change classification

**OTEL-REQ-016**
Every dependency update, semantic-convention update, SDK update, exporter update, instrumentation change, resource change, metric-reader change, sampler-profile change, or signal-registry change MUST be classified by this table before it is accepted as conformant:

| Change class | Definition | NLSpec revision required | Migration note required | Acceptance impact |
| --- | --- | ---: | ---: | --- |
| `registry_equivalent` | Same emitted spans, span names, span kinds, span events, span links, metric identities, metric temporality, resource attributes, log mappings, instrumentation scopes, standard attributes, custom attributes, and forbidden-value exclusions for the same conformance corpus. | No, if dependency-only. | No. | Existing acceptance criteria must pass unchanged. |
| `additive_non_breaking` | Adds a new emitted telemetry element or optional low-cardinality attribute without removing, renaming, retyping, or changing existing emitted telemetry. | Yes. | Yes. | Add criteria proving the addition is intentional and privacy-safe. |
| `privacy_tightening` | Removes or suppresses telemetry to reduce disclosure, cardinality, unsafe correlation, or unbounded retention risk. | Yes. | Yes. | Add criteria proving the removed element is absent and required operational questions retain required coverage. |
| `breaking_shape_change` | Removes, renames, retypes, changes requiredness, changes default emission, changes span parent/link topology, changes metric temporality or aggregation, changes resource identity, changes log mapping, or changes sampler profile. | Yes. | Yes. | Add or update criteria for old-shape absence and new-shape presence. |

**OTEL-REQ-017**
Dependency-only updates MAY occur without an NLSpec revision only when emitted telemetry remains `registry_equivalent` for the conformance corpus. A semantic-convention update, SDK update, generated-constant update, or `OTEL_SEMCONV_STABILITY_OPT_IN` setting MUST NOT become active if it causes any `additive_non_breaking`, `privacy_tightening`, or `breaking_shape_change` effect without this NLSpec being revised.

## 5. Instrumentation ownership and scopes

The application deployable owns the browser-facing UI host, API surface, WebSocket hub, and background-job runners because Cartulary's base deployment is one web application deployable plus Postgres and S3-compatible object storage.

### 5.1 Subsystem boundary

**OTEL-REQ-018**
The instrumentation subsystem MUST be a logical internal boundary inside the modular monolith. It MUST NOT require a separate application deployable, sidecar, microservice, Collector, vendor backend, Prometheus server, or browser telemetry service.

**OTEL-REQ-019**
Instrumentation ownership MUST follow this table:

| Runtime area | Instrumentation owner | Required coverage | Forbidden ownership leak |
| --- | --- | --- | --- |
| HTTP API | Application server instrumentation | Server spans, request metrics, status/error classification. | Route handlers MUST NOT configure exporters or SDK providers. |
| WebSocket subscription | Collaboration instrumentation | Connect, authorize, subscribe, close, event-send, replay, and overflow metrics. | WebSocket payload content MUST NOT become telemetry attributes. |
| Background jobs | Jobs instrumentation | Enqueue, start, terminal state, cancellation, duration, active count, and error metrics. | Job IDs MUST NOT be emitted. |
| Postgres access | Platform/Postgres instrumentation | Standard database client spans and duration metrics without SQL text, bind values, table names, projection names, or connection endpoints. | Workbook modules MUST NOT directly emit raw database query text. |
| Object storage | Platform/object-store instrumentation | Cartulary object-store dependency spans and byte/duration metrics without bucket names, keys, hashes, filenames, upload IDs, copy sources, or handles. | Object-store implementation details MUST NOT leak into evidence telemetry. |
| Workbook query and mutation | Workbook/projection instrumentation | Query, create, patch, conflict, projection-maintenance, and refresh spans. | Projection table names and visible row positions MUST NOT become telemetry identity. |
| Browser UI | Browser controller instrumentation | Local performance marks MAY exist. | Browser direct OTLP, vendor-native, or third-party telemetry export is forbidden. |
| Telemetry bootstrap | Server-side telemetry boundary | SDK provider setup, processors, exporters, metric readers, samplers, shutdown, and self-diagnostics. | Ordinary instrumentation units MUST NOT configure SDK, exporter, processor, reader, sampler, propagator, or Collector behavior. |

### 5.2 Instrumentation scopes

**OTEL-REQ-020**
Each tracer, meter, or logger MUST be created with one of the following instrumentation scopes. Scope version MUST be the Cartulary build version when known, else `0.0.0+unknown`. Schema URL MUST be `null` in this revision. Scope attributes MUST be empty in this revision.

| Scope name | Owning instrumentation area | Allowed signals |
| --- | --- | --- |
| `cartulary.httpapi` | HTTP API and route-family classification | traces, metrics, logs |
| `cartulary.workbook` | Workbook query, row create, row mutation, conflict, projection work, and row refresh | traces, metrics, logs |
| `cartulary.collaboration` | WebSocket subscription and server-to-client events | traces, metrics, logs |
| `cartulary.jobs` | Background-job enqueue and execution | traces, metrics, logs |
| `cartulary.postgres` | Postgres dependency spans and pool/dependency metrics | traces, metrics |
| `cartulary.objectstore` | S3-compatible object storage abstraction | traces, metrics |
| `cartulary.telemetry` | Telemetry self-metrics and bounded diagnostics | metrics, logs |

**OTEL-REQ-021**
No instrumentation unit may create an unregistered instrumentation scope. A future revision that adds scope attributes MUST define a closed attribute table with names, types, allowed values, cardinality bound, default, and forbidden-value tests.

### 5.3 No-SDK instrumentation mode

**OTEL-REQ-022**
Ordinary instrumentation units MUST execute safely when no SDK provider is installed. In no-SDK mode, API calls MUST be no-op or API-default behavior, product behavior MUST remain unchanged, and instrumentation MUST NOT throw because exporters, processors, metric readers, log processors, samplers, or SDK providers are absent.

**OTEL-REQ-023**
When `telemetry.enabled=false`, the telemetry bootstrap boundary MUST install no exporters, no processors, no metric readers, no log bridge, no resource detectors, and no SDK autoconfiguration. It MAY install no-op API providers or leave API default providers in place only when ordinary instrumentation units remain safe.

## 6. Configuration contract

Telemetry configuration lives in the Cartulary deployment configuration surface. Core 04 owns the operator-facing deployment configuration artifact, discovery precedence, binding keys, and fail-closed startup validation; this NLSpec adds telemetry keys under that same surface rather than defining a second configuration model.

### 6.1 Configuration keys

**OTEL-REQ-024**
The effective telemetry configuration MUST be the closed key set in this table. Unknown `telemetry.*` keys are invalid unless a later revision defines them.

| Key | Type | Default | Bounds or values | Omitted behavior | Explicit `null` behavior | Required behavior |
| --- | --- | --- | --- | --- | --- | --- |
| `telemetry.enabled` | boolean | `true` | `true`, `false` | Use default. | Invalid. | When `false`, no SDK providers, exporters, processors, metric readers, log bridges, resource detectors, or instrumentation hooks are installed except no-op placeholders needed for code safety. |
| `telemetry.otel_env_passthrough` | boolean | `false` | exactly `false` in this revision | Use default. | Invalid. | OTel SDK environment variables, declarative config, and SDK autoconfig MUST NOT enable exporters, propagators, samplers, processors, endpoints, headers, config files, views, metric readers, resource detectors, or plugins. |
| `telemetry.exporter.kind` | enum | `none` | `none`, `otlp_http`, `otlp_grpc` | Use default. | Invalid. | `none` disables network export. `otlp_http` and `otlp_grpc` require `telemetry.exporter.endpoint`. |
| `telemetry.exporter.endpoint` | URL string or null | `null` | `http` or `https` URL with no userinfo, query, or fragment | Required when exporter kind is not `none`; otherwise default. | Valid only when exporter kind is `none`. | Cartulary MUST NOT rely on OpenTelemetry's default localhost OTLP endpoint. |
| `telemetry.exporter.headers` | map of string to string | `{}` | Header keys non-empty ASCII token using letters, digits, `_`, `.`, or `-`; values non-empty strings | Use default. | Invalid. | Secret-bearing. MUST be server-side only and redacted everywhere. |
| `telemetry.exporter.protocol` | enum | Derived from `telemetry.exporter.kind` | `grpc`, `http/protobuf` | Derive from exporter kind. | Invalid. | `otlp_grpc` maps to `grpc`; `otlp_http` maps to `http/protobuf`; `http/json` is unsupported. |
| `telemetry.exporter.compression` | enum | `none` | `none`, `gzip` | Use default. | Invalid. | Applies to OTLP export attempts when supported by the selected transport. |
| `telemetry.exporter.retry.enabled` | boolean | `true` | `true`, `false` | Use default. | Invalid. | Enables bounded OTLP transient-error retry. |
| `telemetry.exporter.retry.max_elapsed_ms` | integer | `30000` | `0..300000` | Use default. | Invalid. | `0` disables retries even when retry is enabled. |
| `telemetry.exporter.retry.initial_interval_ms` | integer | `100` | `50..30000` | Use default. | Invalid. | Initial retry interval for transient export failures. |
| `telemetry.exporter.retry.max_interval_ms` | integer | `5000` | `100..60000` | Use default. | Invalid. | Maximum retry interval. Must be greater than or equal to `initial_interval_ms`. |
| `telemetry.exporter.retry.multiplier` | decimal | `2.0` | `1.0..5.0` | Use default. | Invalid. | Exponential retry multiplier. |
| `telemetry.traces.enabled` | boolean | `true` | `true`, `false` | Use default. | Invalid. | Has effect only when `telemetry.enabled=true`. |
| `telemetry.traces.sample_ratio` | decimal | `0.10` | `0.0..1.0` inclusive | Use default. | Invalid. | Selects the sampler profile by §9.2. |
| `telemetry.traces.sampler_profile` | enum | `auto` | `auto`, `cartulary.sampler.always_on.v1`, `cartulary.sampler.always_off.v1`, `cartulary.sampler.traceidratio_compat.v1` | Use default. | Invalid. | `auto` derives from `sample_ratio`; non-auto must be consistent with `sample_ratio` by §9.2. |
| `telemetry.traces.accept_remote_context` | boolean | `false` | exactly `false` in this revision | Use default. | Invalid. | Remote trace context is not trusted in this revision. |
| `telemetry.metrics.enabled` | boolean | `true` | `true`, `false` | Use default. | Invalid. | Has effect only when `telemetry.enabled=true`. |
| `telemetry.metrics.temporality_profile` | enum | `cartulary.metrics.temporality.cumulative.v1` | exactly `cartulary.metrics.temporality.cumulative.v1` | Use default. | Invalid. | MetricReader MUST export cumulative temporality for all registered current-profile instruments. |
| `telemetry.metrics.exemplars.enabled` | boolean | `false` | exactly `false` in this revision | Use default. | Invalid. | Configure the SDK exemplar filter to `AlwaysOff` or equivalent. Exemplar emission is non-conformant in this revision. |
| `telemetry.logs.bridge_enabled` | boolean | `false` | `true`, `false` | Use default. | Invalid. | When `false`, local structured logs MAY include trace correlation fields but MUST NOT be exported as OTel logs. |
| `telemetry.logs.body_max_chars` | integer | `2048` | `0..8192` | Use default. | Invalid. | Maximum Unicode scalar values in exported OTel LogRecord `Body` after redaction. `0` exports an empty string body. |
| `telemetry.processor.max_queue_size` | integer | `2048` | `1..65536` | Use default. | Invalid. | Queue bound for each enabled signal processor. |
| `telemetry.processor.max_export_batch_size` | integer | `512` | `1..telemetry.processor.max_queue_size` | Use default. | Invalid. | Maximum batch size for each export attempt. |
| `telemetry.processor.traces.schedule_delay_ms` | integer | `5000` | `100..300000` | Use default. | Invalid. | Trace processor periodic export delay. |
| `telemetry.processor.metrics.schedule_delay_ms` | integer | `60000` | `100..300000` | Use default. | Invalid. | Metric processor periodic export delay. |
| `telemetry.processor.logs.schedule_delay_ms` | integer | `1000` | `100..300000` | Use default. | Invalid. | Log processor periodic export delay when log bridge is enabled. |
| `telemetry.processor.export_timeout_ms` | integer | `2000` | `100..10000` | Use default. | Invalid. | Cartulary exporter attempt timeout; MUST compile into SDK/exporter config. |
| `telemetry.processor.overflow_policy` | enum | `drop_new` | exactly `drop_new` | Use default. | Invalid. | When a queue is full, the telemetry item being offered is dropped and older queued telemetry is retained. |
| `telemetry.shutdown.flush_timeout_ms` | integer | `5000` | `100..30000` | Use default. | Invalid. | Maximum graceful flush wait. |
| `telemetry.self_diagnostics.enabled` | boolean | `true` | `true`, `false` | Use default. | Invalid. | Enables bounded local self-diagnostics and self-metrics. |
| `telemetry.self_diagnostics.recursion_guard` | enum | `drop_telemetry_about_telemetry` | exactly `drop_telemetry_about_telemetry` | Use default. | Invalid. | Prevents exporter self-telemetry from recursively generating unbounded exporter telemetry. |
| `telemetry.resource.service_name` | string | `cartulary.app` | `1..128` ASCII letters, digits, `.`, `_`, or `-` | Use default. | Invalid. | Maps to `service.name`. Startup MUST reject SDK-generated `unknown_service:*` effective identity. |
| `telemetry.resource.service_namespace` | string | `cartulary` | `1..128` ASCII letters, digits, `.`, `_`, or `-` | Use default. | Invalid. | Maps to `service.namespace`. |
| `telemetry.resource.service_version` | string | Build version, else `0.0.0+unknown` | `1..128` non-empty string | Use default. | Invalid. | Maps to `service.version`. |
| `telemetry.resource.service_instance_id` | string or null | Generated UUID v4 per process start | Non-empty opaque string, maximum `128` Unicode scalar values | Generate default. | Generate default. | Maps to `service.instance.id`; MUST satisfy §7. |
| `telemetry.resource.deployment_environment_name` | string or null | `null` | `development`, `test`, `staging`, `production`, or custom non-empty token of maximum `128` ASCII letters, digits, `.`, `_`, or `-` | Omit attribute. | Omit attribute. | Maps to `deployment.environment.name` when present. |
| `telemetry.attribute.incident_correlation` | enum | `none` | `none`, `hmac_64bit` | Use default. | Invalid. | `none` forbids incident correlation attributes. `hmac_64bit` is a narrowed opt-in under §8.6. |
| `telemetry.attribute.hmac_secret_ref` | string or null | `null` | Server-side secret reference | Required only when incident correlation is `hmac_64bit`. | Valid only when incident correlation is `none`; otherwise invalid. | Secret value MUST NOT be exported. |

### 6.2 Configuration precedence

**OTEL-REQ-025**
Configuration precedence MUST be exactly this table:

| Precedence | Source | Required behavior |
| --- | --- | --- |
| 1 | Cartulary deployment configuration | Authoritative for all telemetry behavior. |
| 2 | Cartulary server-side environment bindings | MAY populate Cartulary deployment configuration keys only. Empty values are treated as omitted under §6.2.1. |
| 3 | OTel SDK environment variables | Ignored in this revision because `telemetry.otel_env_passthrough` is fixed to `false`. |
| 4 | OTel declarative configuration and plugin providers | Not authoritative in this revision. |
| 5 | OTel SDK defaults | MAY apply only inside the effective Cartulary configuration envelope. MUST NOT enable export when Cartulary export is `none`. |
| 6 | Browser state or browser environment | Never a telemetry exporter, processor, sampler, propagator, metric-reader, header, or resource configuration source. |

### 6.2.1 Cartulary server-side environment binding parser

**OTEL-REQ-122**
A server-side environment binding MAY populate a Cartulary deployment-configuration key only after it is parsed into the same canonical deployment-configuration object shape used by file-based deployment configuration. Parsing MUST complete before telemetry bootstrap, SDK provider activation, exporter construction, resource construction, or readiness.

**OTEL-REQ-123**
Cartulary server-side environment bindings MUST use this parser contract. This table governs Cartulary deployment keys only. It does not adopt OTel SDK environment-variable fallback behavior.

| Input family | Omitted or unset | Empty string | Valid non-empty value | Invalid non-empty value | Explicit string `"null"` |
| --- | --- | --- | --- | --- | --- |
| Boolean | Omit key and use deployment-config default. | Omit key and use deployment-config default. | Case-insensitive `true` or `false` only. | Fail startup with `invalid_deployment_config` and a telemetry-specific reason code. | Invalid unless the target key type is string and the owning row allows that literal. |
| Integer | Omit key and use default. | Omit key and use default. | Base-10 ASCII digits only, no sign, no separators, within the target key bounds. | Fail startup. | Invalid. |
| Decimal | Omit key and use default. | Omit key and use default. | Finite decimal token within target key bounds; exponent form is invalid unless the owning key row explicitly allows it. | Fail startup. | Invalid. |
| Enum | Omit key and use default. | Omit key and use default. | Exact case-sensitive token from the target key's closed vocabulary. | Fail startup. | Invalid. |
| URL | Omit key and use default. | Omit key and use default. | URL satisfying the `telemetry.exporter.endpoint` row: `http` or `https`, no userinfo, no query, no fragment. | Fail startup. | Invalid. |
| Header map | Omit key and use `{}`. | Omit key and use `{}`. | Deterministic comma-separated `key=value` pairs or a repository-owned structured secret binding. Keys must satisfy §6.1. Duplicate keys after exact byte comparison fail. Values are secret-bearing non-empty strings. | Fail startup; never log values. | Invalid. |
| String or nullable string | Omit key and use default. | Omit key unless the target row explicitly permits empty string. | Target row validation applies after Unicode scalar counting or ASCII-token validation. | Fail startup. | Literal string `"null"` is not JSON null and is invalid unless the target row explicitly allows that string. |

**OTEL-REQ-124**
Invalid Cartulary environment bindings MUST fail startup. They MUST NOT be gracefully ignored as if they were OTel SDK environment variables. Secret-bearing parsed values, invalid values, and raw environment values MUST NOT appear in startup diagnostics, logs, spans, metrics, retained test artifacts, or exporter headers.

### 6.3 OpenTelemetry external configuration containment

OpenTelemetry defines environment-variable configuration and declarative configuration mechanisms that can alter SDK behavior, including SDK disablement, resource attributes, service name, propagators, samplers, processors, exporters, OTLP endpoints, metrics exemplars, metric export cadence, attribute limits, and file-based SDK component creation.

**OTEL-REQ-026**
External OpenTelemetry configuration inputs MUST be contained by this matrix:

| External input | Current-profile treatment | Allowed effect |
| --- | --- | --- |
| OTel SDK environment variables | Ignored for behavior selection. | May appear only as redacted presence diagnostics. |
| `OTEL_CONFIG_FILE` | Ignored for behavior selection. | No exporter, processor, reader, sampler, propagator, header, resource detector, exemplar, view, or plugin creation. |
| `OTEL_EXPERIMENTAL_CONFIG_FILE` | Ignored for behavior selection. | None. |
| OTel declarative configuration file | Not an authoritative Cartulary configuration source. | None. |
| OTel instrumentation ConfigProvider state | Not an authoritative Cartulary configuration source. | None. |
| OTel plugin component provider | Forbidden as runtime configuration authority. | None. |
| `OTEL_SEMCONV_STABILITY_OPT_IN` | Ignored unless a future revision defines a semantic-convention migration profile. | No emitted-shape change. |
| Per-signal OTLP endpoint env vars | Ignored and never authoritative. | None. |
| Per-signal OTLP protocol or header env vars | Ignored and never authoritative. | None. |
| Browser state | Never an exporter or processor configuration source. | None. |

**OTEL-REQ-125**
The hostile declarative-configuration fixture set MUST include each attempted component below. Each attempted component MUST be rejected, ignored, or overridden before provider activation so that it has no externally observable telemetry effect.

| Hostile declarative-config attempt | Required result |
| --- | --- |
| Instantiate OTLP/HTTP exporter while `telemetry.exporter.kind='none'` | No exporter is created; no network export attempt occurs. |
| Instantiate OTLP/gRPC exporter with per-signal endpoint | No per-signal endpoint is used; no exporter is created unless Cartulary deployment config explicitly enables export. |
| Instantiate resource detector for host/process/container/cloud/service | Detector is not invoked; effective Resource remains closed. |
| Instantiate sampler, adaptive sampler, remote sampler, or sampler plugin | §9.2 sampler profile remains unchanged. |
| Instantiate metric reader or alter export interval | §10.3 temporality and configured schedule remain unchanged. |
| Instantiate View that renames a stream or changes attributes | The View is rejected, ignored, or overridden by Cartulary-owned metric view registry; no unregistered metric stream is exported. |
| Enable exemplars | Exemplars remain disabled. |
| Configure log bridge or log exporter | Log bridge remains controlled only by `telemetry.logs.bridge_enabled`. |
| Reference custom plugin component provider | Plugin is not loaded or invoked. |

### 6.4 OpenTelemetry environment hazard registry

**OTEL-REQ-027**
The current profile MUST ignore the following environment variable families for behavior selection. The implementation MAY record only their presence in redacted startup diagnostics that do not expose values.

| Environment variable family | Members or pattern | Hazard contained |
| --- | --- | --- |
| SDK disablement | `OTEL_SDK_DISABLED` | Cannot disable Cartulary-selected providers or invert no-SDK tests. |
| Resource environment merge | `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_SERVICE_NAME`, `OTEL_ENTITIES` | Cannot add host, process, incident, user, deployment, entity, or service attributes. |
| SDK internal logging | `OTEL_LOG_LEVEL` | Cannot change product logs, operator diagnostics, or secret exposure. |
| Propagators | `OTEL_PROPAGATORS` | Cannot enable Baggage, vendor propagators, or remote context propagation. |
| Trace sampler | `OTEL_TRACES_SAMPLER`, `OTEL_TRACES_SAMPLER_ARG` | Cannot override §9.2 sampler profile. |
| Global attribute limits | `OTEL_ATTRIBUTE_COUNT_LIMIT`, `OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT` | Cannot be treated as privacy enforcement and cannot change Cartulary redaction-before-recording. |
| Span limits | `OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT`, `OTEL_SPAN_ATTRIBUTE_VALUE_LENGTH_LIMIT`, `OTEL_SPAN_EVENT_COUNT_LIMIT`, `OTEL_SPAN_LINK_COUNT_LIMIT`, `OTEL_ATTRIBUTE_COUNT_LIMIT`, `OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT` | Cannot authorize extra attributes, events, or links, and cannot be used as privacy controls. |
| LogRecord limits | `OTEL_LOGRECORD_ATTRIBUTE_COUNT_LIMIT`, `OTEL_LOGRECORD_ATTRIBUTE_VALUE_LENGTH_LIMIT` | Cannot authorize raw log fields or be used as log privacy controls. |
| Batch span processor | `OTEL_BSP_*` | Cannot override Cartulary processor bounds, delay, timeout, or queue policy. |
| Batch LogRecord processor | `OTEL_BLRP_*` | Cannot enable or alter log export outside Cartulary config. |
| Exporter selection | `OTEL_TRACES_EXPORTER`, `OTEL_METRICS_EXPORTER`, `OTEL_LOGS_EXPORTER` | Cannot enable OTLP, stdout, zipkin, prometheus, vendor-native, or other export. |
| Global OTLP | `OTEL_EXPORTER_OTLP_*` | Cannot set endpoint, headers, protocol, timeout, compression, certificates, mTLS, or per-signal behavior. |
| Per-signal OTLP | `OTEL_EXPORTER_OTLP_TRACES_*`, `OTEL_EXPORTER_OTLP_METRICS_*`, `OTEL_EXPORTER_OTLP_LOGS_*` | Cannot create divergent per-signal endpoints, headers, protocol, or credentials. |
| Non-OTLP exporters | `OTEL_EXPORTER_ZIPKIN_*`, `OTEL_EXPORTER_PROMETHEUS_*`, implementation-specific exporter env vars | Cannot enable unsupported exporters. |
| Metrics exemplar | `OTEL_METRICS_EXEMPLAR_FILTER` | Cannot enable exemplars. |
| Metrics periodic reader | `OTEL_METRIC_EXPORT_INTERVAL`, `OTEL_METRIC_EXPORT_TIMEOUT` | Cannot override metric export cadence or timeout. |
| Declarative configuration | `OTEL_CONFIG_FILE`, `OTEL_EXPERIMENTAL_CONFIG_FILE` | Cannot instantiate SDK components from file configuration. |
| Semantic-convention opt-in | `OTEL_SEMCONV_STABILITY_OPT_IN` | Cannot change emitted attribute names or duplicate conventions. |
| Language-specific OTel env vars | `OTEL_{LANGUAGE}_{FEATURE}` | Ignored unless a future revision maps exact variables by name. |

**OTEL-REQ-028**
Any OTel SDK autoconfiguration package, declarative configuration package, resource detector package, exporter package, log bridge package, or plugin package that reads the environment before Cartulary configuration validation MUST be disabled, bypassed, or wrapped so it cannot create externally observable telemetry behavior outside this NLSpec.

### 6.5 Configuration validation

**OTEL-REQ-029**
Invalid telemetry configuration MUST fail deployment-configuration validation before readiness with `error.code='invalid_deployment_config'` and `reason_code='invalid_telemetry_config'` or an equivalent deployment-config reason-code family if the repository owner has already defined a narrower code. Exporter endpoint network unavailability MUST NOT fail startup when the endpoint is syntactically valid.

**OTEL-REQ-030**
Configuration validation MUST enforce these cross-key rules:

| Rule | Failure condition |
| --- | --- |
| Export endpoint requirement | `telemetry.exporter.kind` is `otlp_http` or `otlp_grpc` and `telemetry.exporter.endpoint` is omitted or null. |
| Endpoint absence for disabled export | `telemetry.exporter.kind='none'` and `telemetry.exporter.endpoint` is a non-null URL. |
| Protocol consistency | `telemetry.exporter.kind='otlp_http'` with protocol other than `http/protobuf`, or `telemetry.exporter.kind='otlp_grpc'` with protocol other than `grpc`. |
| Batch bound | `max_export_batch_size > max_queue_size`. |
| Retry interval order | `retry.max_interval_ms < retry.initial_interval_ms`. |
| HMAC secret | `incident_correlation='hmac_64bit'` and `hmac_secret_ref` is omitted, null, or empty. |
| Remote-context attempt | Any configuration attempts to set `telemetry.traces.accept_remote_context=true`. |
| Unsupported sampler | Any sampler profile outside §9.2, including `ProbabilitySampler`, `JaegerRemoteSampler`, or implementation-specific remote/adaptive samplers. |
| Unsupported protocol | Any exporter protocol value other than `grpc` or `http/protobuf`. |
| Per-signal endpoints | Any per-signal endpoint key appears in effective behavior. |
| Per-signal protocol or header divergence | Any per-signal protocol or header key appears in effective behavior. |
| Exemplar enablement | `telemetry.metrics.exemplars.enabled` is any value other than `false`. |
| Log body bound | `telemetry.logs.body_max_chars` is outside `0..8192`. |
| Environment passthrough | `telemetry.otel_env_passthrough=true`. |
| External OTel config authority | Any OTel declarative config, SDK autoconfig, plugin provider, ConfigProvider state, or resource detector attempts to create or alter exporters, processors, propagators, samplers, metric readers, log processors, header capture, metric views, exemplars, resources, or SDK plugin components outside declared `telemetry.*` keys. |
| Semantic-convention environment opt-in | `OTEL_SEMCONV_STABILITY_OPT_IN` would alter emitted telemetry shape in the current profile. |
| Resource schema URL | Any effective provider Resource would have a non-empty schema URL in the current profile. |
| View rename | Any metric View would create an exported stream name not exactly registered in §10.4. |

## 7. Resource identity and detector policy

OpenTelemetry Resources are immutable entity descriptors associated with telemetry providers. SDKs may provide default resource attributes, merge environment-derived attributes, and support resource detector packages for process, host, container, service, cloud, and similar runtime facts.

### 7.1 Resource attribute registry

**OTEL-REQ-031**
The instrumentation subsystem MUST attach the following resource attributes to all emitted traces, metrics, and exported logs:

| Attribute | Requiredness | Value source | Export rule | Privacy rule |
| --- | --- | --- | --- | --- |
| `service.name` | Required | `telemetry.resource.service_name` | Always export. | MUST NOT resolve to SDK-generated `unknown_service:*`. |
| `service.namespace` | Required | `telemetry.resource.service_namespace` | Always export. | Default is `cartulary`. |
| `service.version` | Required | `telemetry.resource.service_version` | Always export. | Omission is non-conformant after default resolution. |
| `service.instance.id` | Required | Configured value or generated UUID v4 | Always export. | MUST be opaque and MUST NOT encode host, user, incident, process ID, container ID, filesystem path, or object-store identity. |
| `deployment.environment.name` | Optional | `telemetry.resource.deployment_environment_name` | Export only when configured. | Descriptive only. MUST NOT drive tenancy, service identity, authorization, or incident identity. |
| `telemetry.sdk.language` | SDK-required | SDK-provided | Do not override. | SDK value only. |
| `telemetry.sdk.name` | SDK-required | SDK-provided | Do not override. | SDK value only. |
| `telemetry.sdk.version` | SDK-required | SDK-provided | Do not override. | SDK value only. |
| `cartulary.deployment.profile` | Required | Deployment profile | Always export. | Low-cardinality deployment profile token only. |
| `cartulary.profile.claims` | Required | Sorted claimed profile IDs | Always export. | Profile identifiers only; no incident data. |

**OTEL-REQ-032**
No other resource attribute may be emitted in this revision unless an SDK requires a `telemetry.sdk.*` attribute by specification. If an SDK injects a default resource attribute outside this registry, bootstrap MUST remove it, overwrite the provider resource with the closed registry, or fail startup before provider activation.

**OTEL-REQ-126**
The current profile's effective provider Resource `schema_url` MUST be the empty string. Resource construction MUST NOT merge a non-empty schema URL from SDK defaults, environment resource attributes, resource detectors, declarative configuration, or plugin providers. If the selected SDK path cannot construct an effective Resource with the closed attribute registry and empty schema URL, telemetry bootstrap MUST fail before readiness.

**OTEL-REQ-127**
Resource schema URL merge behavior MUST be fail-closed by this table:

| Source | Current-profile rule |
| --- | --- |
| SDK default Resource schema URL | Must not become effective unless empty. |
| Environment-derived Resource schema URL | Must not become effective. |
| Detector-derived Resource schema URL | Detector must not run; schema URL must not merge. |
| Two different non-empty schema URLs | Fail before provider activation; do not allow implementation-specific merged Resource contents. |
| Future pinned semantic-convention schema URL | Reserved for a later revision and classified as `breaking_shape_change` unless the golden corpus proves registry equivalence. |

### 7.2 Resource detector and environment merge policy

**OTEL-REQ-033**
Cartulary MUST NOT enable host, process, OS, container, Kubernetes, cloud, FaaS, browser, device, telemetry entity, environment-variable, or vendor-specific resource detectors in the current profile.

**OTEL-REQ-034**
The resource MUST be constructed from the closed Cartulary registry in §7.1 only. `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_SERVICE_NAME`, `OTEL_ENTITIES`, SDK-provided process facts, detector-provided host facts, detector-provided container facts, and detector-provided cloud facts MUST NOT be merged into the effective resource.

**OTEL-REQ-035**
The following resource attributes are explicitly not adopted and MUST NOT be emitted:

| Attribute family | Reason |
| --- | --- |
| `host.*` | Host identity and topology disclosure. |
| `os.*` | Host and runtime environment disclosure. |
| `process.*` | Process ID, executable path, command line, owner, and runtime disclosure. |
| `container.*` | Container identity disclosure. |
| `k8s.*` | Cluster, namespace, pod, workload, and node disclosure. |
| `cloud.*` | Cloud provider, account, region, resource, and instance disclosure. |
| `faas.*` | Deployment topology disclosure. |
| `service.criticality` | Development-status semantic convention and potential incident-severity confusion. |
| `enduser.*` | User identity is forbidden in telemetry. |
| `session.*` | Session identity is forbidden in telemetry. |
| `telemetry.distro.*` | Not adopted unless a future revision defines distribution identity. |

## 8. Attribute governance, forbidden values, and value shapes

### 8.1 Attribute classes

**OTEL-REQ-036**
Every telemetry attribute MUST belong to exactly one class in this table:

| Attribute class | Rule |
| --- | --- |
| OTel stable standard attributes | MAY be emitted only when allowed by the signal-specific allowlist and privacy rules. |
| OTel development or experimental attributes | MUST NOT be emitted unless a future revision adopts them. |
| OTel deprecated attributes | MUST NOT be emitted unless a migration profile requires them. |
| OTel privacy-conflicting attributes | MUST NOT be emitted even if OTel defines them. |
| Cartulary custom attributes | MUST use `cartulary.` and MUST appear in §8.5. |
| Reserved namespace attributes | MUST NOT be created by Cartulary as custom attributes. |
| Unknown custom attributes | MUST NOT be emitted. |
| High-cardinality values | MUST NOT be emitted unless the exact attribute and bound are listed. |

### 8.2 Attribute value-shape policy

OpenTelemetry permits broad `AnyValue` shapes, including primitive values, arrays, byte arrays, maps, and empty values. Cartulary uses a narrower profile because telemetry values must be privacy-safe before SDK processor or exporter handoff.

**OTEL-REQ-037**
Cartulary telemetry attributes MUST use only the value shapes in this table:

| Value shape | Default rule | Exception rule |
| --- | --- | --- |
| string scalar | Allowed only when the owning attribute registry row permits string. | String values must satisfy the registry cardinality and forbidden-value rules. |
| integer scalar | Allowed only when the owning attribute registry row permits integer. | Numeric values must not encode identifiers, hashes, timestamps, row positions, byte offsets into evidence, or object sizes unless the metric or span registry explicitly owns that measurement. |
| floating-point scalar | Allowed only for metric measurements and histogram values unless a registry row permits otherwise. | NaN and Infinity MUST NOT be emitted. |
| boolean scalar | Allowed only when the owning registry row permits boolean. | Must not encode hidden presence of forbidden content unless the registry names the meaning. |
| null | Means omission by default. | May be emitted only when the owning registry row explicitly says emitted null is observable. No current-profile registry row permits emitted null attributes. |
| homogeneous array | Forbidden by default. | May be emitted only by a future revision with exact item type, maximum length, ordering, and forbidden-value tests. |
| byte array | Forbidden. | No current-profile exception. |
| map or nested object | Forbidden. | No current-profile exception. |
| arbitrary JSON | Forbidden. | No current-profile exception. |

**OTEL-REQ-128**
A Cartulary instrumentation wrapper MUST NOT call an OTel API with an attribute whose value is `null`, `undefined`, `None`, `nil`, or the language-equivalent empty value. For nullable internal inputs, the wrapper MUST omit the attribute key before calling the OTel API unless the owning registry row explicitly permits emitted null. No current-profile registry row permits emitted null. The string value `"null"` is not a null value and MUST pass ordinary string validation and forbidden-value checks.

**OTEL-REQ-038**
SDK attribute count limits, attribute value length limits, log body truncation, exporter queue overflow, batch dropping, or Collector/backend filtering MUST NOT be treated as privacy enforcement. Redaction and omission MUST occur before recording.

### 8.3 Reserved namespaces

**OTEL-REQ-039**
Cartulary custom attributes MUST NOT use any prefix in this table:

| Prefix | Rule |
| --- | --- |
| `otel.` | Reserved for OpenTelemetry. |
| `service.` | Reserved for OTel resource attributes. |
| `telemetry.` | Reserved for OTel SDK resource attributes. |
| `http.` | Reserved for OTel HTTP semantic conventions. |
| `url.` | Reserved for OTel URL semantic conventions. |
| `network.` | Reserved for OTel network semantic conventions. |
| `server.` | Reserved for OTel server semantic conventions. |
| `client.` | Reserved for OTel client semantic conventions. |
| `db.` | Reserved for OTel DB semantic conventions. |
| `aws.` | Reserved for OTel or AWS semantic conventions. |
| `exception.` | Reserved for OTel exception semantic conventions. |
| `enduser.` | Not used because user identity is forbidden in telemetry. |
| `user.` | Not used because user identity is forbidden in telemetry. |
| `session.` | Not used because session identity is forbidden in telemetry. |

### 8.4 Forbidden telemetry values and attributes

**OTEL-REQ-040**
The forbidden value family is closed by this table. The implementation MUST prevent these values from reaching spans, metrics, logs, events, exception fields, resource attributes, exporter headers, exporter artifacts, self-diagnostics, retained telemetry artifacts, and any app-mediated future telemetry route.

| Family | Forbidden examples |
| --- | --- |
| Incident-authored content | Timeline summary, note text, evidence title, finding narrative, handoff summary, status-review text, lesson text, analyst-provided source text. |
| Evidence content and identity | Evidence bytes, filenames, `filename_hint`, blob hashes, storage refs, object keys, bucket names, preview handles, download handles, upload IDs, copy sources, object-blob IDs. |
| Stable product identifiers | `incident_id`, `record_id`, `row_version`, `user_id`, `party_id`, `identity_id`, `host_id`, `saved_view_id`, `change_set_id`, `job_id`, `session_id`, `connection_id`, `client_txn_id`, conflict tokens, handle tokens. |
| Security material | Passwords, TOTP seeds, bootstrap tokens, session tokens, bearer tokens, CSRF tokens, exporter header values, API keys, object-store credentials, reference-pack signing keys. |
| Request content | Raw URL path containing concrete IDs, raw query strings, request bodies, response bodies, request headers, response headers, cookies, authorization headers. |
| Database content | SQL text, sanitized SQL, parameterized SQL text, bind values, saved-view query JSON, table names, projection table names, index names, schema names, source headers, search text. |
| Network and infrastructure identity | Client IP, hostnames, pod names, container IDs, MAC addresses, database addresses, object-store endpoints, local filesystem roots, deployment root paths. |
| Exception detail | `exception.message`, `exception.stacktrace`, panic string, wrapped error detail when it includes any forbidden value. |
| Baggage and trace-state detail | Inbound `baggage` values, inbound `tracestate` values, vendor trace-state values. |

**OTEL-REQ-041**
When a value belongs to a forbidden family, the implementation MUST do one of the following before recording telemetry:

| Treatment | Required behavior |
| --- | --- |
| Omit | Do not set the attribute, body text, event field, or diagnostic field. |
| Replace with closed class | Emit only a closed low-cardinality class such as `validation_error`, `permission_denied`, `timeout`, `queue_full`, or `redaction_rejected`. |
| Drop item | Drop the telemetry item when safe omission cannot be proven. |

### 8.5 Cartulary custom attribute registry

**OTEL-REQ-042**
The current profile permits only the custom attributes in this registry:

| Attribute | Type | Allowed values or source | Cardinality bound | Allowed signals |
| --- | --- | --- | --- | --- |
| `cartulary.deployment.profile` | string | Active deployment profile token | Deployment profile vocabulary | resource |
| `cartulary.profile.claims` | string | Sorted comma-delimited claimed profile IDs | Current profile vocabulary combinations | resource |
| `cartulary.module` | string | `httpapi`, `workbook`, `collaboration`, `jobs`, `postgres`, `objectstore`, `telemetry` | 7 | spans, metrics, logs |
| `cartulary.route_family` | string | Closed route-family token, not a route path | Route-family registry count | spans, metrics, logs |
| `cartulary.view_schema_id` | string | Standardized `view_schema_id` token only | Current profile view-schema count | spans, metrics |
| `cartulary.record_type` | string | Closed record-type vocabulary | Current profile record-type count | spans, metrics |
| `cartulary.operation` | string | Closed operation token owned by signal registry | Operation registry count | spans, metrics, logs |
| `cartulary.result` | string | `success`, `rejected`, `conflict`, `canceled`, `failed`, `timeout`, `dropped` | 7 | spans, metrics, logs |
| `cartulary.error_code` | string | Public error-code token or `internal_error` | Public error-code registry count | spans, logs |
| `cartulary.error_class` | string | Closed low-cardinality implementation or dependency class | Error-class registry count | spans, logs |
| `cartulary.websocket.event_type` | string | Public WebSocket event type, not payload content | WebSocket event vocabulary count | spans, metrics |
| `cartulary.job_kind` | string | Background-job kind vocabulary, not job ID | Job-kind vocabulary count | spans, metrics |
| `cartulary.job_terminal_status` | string | `succeeded`, `failed`, `canceled`, `expired` | 4 | spans, metrics |
| `cartulary.signal_kind` | string | `traces`, `metrics`, `logs` | 3 | metrics, logs |
| `cartulary.telemetry.exporter_kind` | string | `none`, `otlp_http`, `otlp_grpc` | 3 | metrics, logs |
| `cartulary.drop_reason` | string | `queue_full`, `redaction_rejected`, `exporter_permanent_discard`, `shutdown_timeout`, `recursion_guard`, `metric_overflow` | 6 | metrics, logs |
| `cartulary.incident.hash64` | string | Optional HMAC-derived 16-character lowercase hex value | Incident count after opt-in approval | spans, metrics |

**OTEL-REQ-043**
Unknown `cartulary.*` attributes MUST NOT be emitted. Adding a new `cartulary.*` attribute is an `additive_non_breaking` change unless it changes requiredness, type, or emitted shape, in which case §4.5 decides the change class.

#### 8.5.1 Error code and error class registry

**OTEL-REQ-141**
`cartulary.error_code` MUST be either:

- one public `error.code` token from Core 01 §3.3.6.1, or
- `internal_error` only when no public API error code exists because the product did not expose a public error.

Core 01 §3.3.6.1 remains the canonical public error-code registry. This NLSpec consumes that registry only for low-cardinality telemetry classification. Runtime inference from prefixes, transport status, retry hints, route family, or string matching is forbidden.

**OTEL-REQ-142**
Every Core 01 §3.3.6.1 public `error.code` token MUST map to exactly one `cartulary.error_class` using this table.

| `cartulary.error_class` | Public `error_code` values |
| --- | --- |
| `request_invalid` | `invalid_view_query`, `invalid_pagination_request`, `invalid_mutation_payload`, `invalid_evidence_handle_request`, `invalid_blob_create_request`, `invalid_incident_create`, `invalid_incident_patch`, `invalid_rollback_request`, `invalid_auth_request`, `invalid_enterprise_auth_request`, `invalid_import_request`, `invalid_snapshot_request`, `invalid_release_request`, `invalid_reference_pack_request`, `invalid_incident_bundle_request` |
| `authentication` | `invalid_credentials`, `mfa_required`, `mfa_setup_required`, `credential_bootstrap_rejected`, `invalid_current_password`, `invalid_second_factor`, `totp_setup_not_pending`, `enterprise_auth_transaction_rejected`, `provider_response_rejected`, `provider_identity_rejected` |
| `capability_unavailable` | `extension_profile_not_claimed`, `auth_provider_not_found`, `auth_provider_disabled` |
| `concurrency_conflict` | `client_txn_conflict`, `row_version_conflict`, `incident_key_conflict`, `incident_version_conflict`, `same_field_conflict`, `user_version_conflict`, `membership_version_conflict`, `auth_binding_conflict` |
| `lifecycle_conflict` | `job_cancel_rejected`, `illegal_transition`, `record_deleted_use_restore`, `record_already_deleted`, `record_delete_blocked`, `record_not_deleted`, `record_locked`, `rollback_precondition_failed`, `last_deployment_admin`, `user_inactive`, `membership_exists_use_patch`, `last_incident_admin`, `merge_precondition_failed`, `import_state_conflict`, `import_apply_blocked`, `release_state_conflict`, `release_approval_rejected`, `reference_pack_state_conflict`, `reference_pack_activation_rejected` |
| `not_found` | `entity_mention_not_found`, `resolved_record_not_found`, `rollback_target_not_found`, `evidence_record_not_found`, `handle_not_found_or_revoked`, `job_not_found`, `auth_binding_not_found`, `user_not_found`, `membership_not_found`, `import_session_not_found`, `import_unit_not_found`, `snapshot_not_found`, `release_not_found`, `reference_pack_not_found`, `incident_bundle_not_found` |
| `expired_or_consumed` | `handle_expired`, `handle_consumed` |
| `policy_rejected` | `blob_create_rejected`, `evidence_attach_rejected`, `import_source_unsupported`, `import_source_rejected`, `release_render_failed`, `reference_pack_verification_failed`, `incident_bundle_export_rejected`, `incident_bundle_import_rejected` |
| `dependency_unavailable` | `evidence_access_unavailable` |

**OTEL-REQ-143**
The complete `cartulary.error_class` registry additionally includes the following dependency and runtime classes.

| Additional `cartulary.error_class` | Scope |
| --- | --- |
| `timeout` | Dependency or exporter timeout. |
| `serialization_conflict` | Database serialization conflict. |
| `constraint_violation` | Safe data-integrity constraint class after redaction. |
| `exporter_transient` | Retryable exporter transport/status class. |
| `exporter_permanent` | Non-retryable exporter rejection class. |
| `redaction_rejected` | Telemetry item dropped because safety could not be proven. |
| `queue_full` | Telemetry processor overflow class when class is needed rather than `cartulary.drop_reason`. |
| `shutdown_timeout` | Telemetry flush exceeded shutdown budget. |
| `recursion_guard` | Telemetry-about-telemetry recursion prevented. |
| `internal_error` | Safe non-detail class for internal failures. |

Adding, removing, or renaming a Core 01 public `error.code` token without updating this mapping is conformance drift. Adding a second independent `cartulary.error_class` registry anywhere in this NLSpec is forbidden.

### 8.6 Incident correlation opt-in

**OTEL-REQ-044**
The default `telemetry.attribute.incident_correlation='none'` MUST omit incident correlation attributes. In default configuration, no incident identifier, incident key, incident title, customer name, or incident-derived hash may appear in telemetry.

**OTEL-REQ-045**
When `telemetry.attribute.incident_correlation='hmac_64bit'`, the implementation MAY emit only `cartulary.incident.hash64`. The value MUST be the first 64 bits of `HMAC-SHA-256(secret, canonical_incident_id_bytes)`, encoded as exactly 16 lowercase hexadecimal characters. The secret value MUST be resolved server-side from `telemetry.attribute.hmac_secret_ref`, MUST NOT be exported, and MUST NOT be available to browser code.

**OTEL-REQ-046**
`cartulary.incident.hash64` is an operational grouping key, not a stable public incident identifier. It MUST NOT appear in product API responses, workbook rows, WebSocket payloads, evidence handles, export snapshots, benchmark manifests, or user-facing UI labels solely because telemetry correlation is enabled.

### 8.7 Redaction-before-recording invariant

**OTEL-REQ-047**
Forbidden-value detection, redaction, omission, replacement, or rejection MUST occur before any telemetry item reaches an OTel span attribute setter, metric measurement, log bridge mapper, SDK processor, exporter queue, retained telemetry artifact, or self-diagnostic sink.

**OTEL-REQ-048**
A Collector, backend, exporter, vendor pipeline, or external scrubber MUST NOT be relied on to satisfy Cartulary privacy conformance. Cartulary telemetry must already be conformant before SDK processor or exporter handoff.

**OTEL-REQ-049**
When redaction cannot prove an item safe, the implementation MUST drop the telemetry item and increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='redaction_rejected'` when doing so does not recurse.

**OTEL-REQ-050**
The forbidden-value test corpus MUST include at least one representative value from every forbidden family in OTEL-REQ-040 and MUST exercise traces, metrics, logs, self-diagnostics, and retained telemetry artifacts.

## 9. Tracing contract

OpenTelemetry spans carry operation name, timestamps, attributes, events, parent span identity, links to causally related spans, and SpanContext. OpenTelemetry links are the appropriate model for batch, async, or trusted-boundary relationships that do not have a single synchronous parent.

### 9.1 General span rules

**OTEL-REQ-051**
Span names MUST use route templates, module operations, or stable operation names. Span names MUST NOT include path IDs, incident IDs, record IDs, user-supplied strings, search text, filenames, object keys, SQL text, visible row values, saved-view IDs, job IDs, user IDs, or handle tokens.

**OTEL-REQ-052**
Span status and `error.type` MUST follow the signal-specific error rules. `cartulary.error_code` carries Cartulary public error-code tokens. The implementation MUST NOT overload `error.type` with Cartulary public error codes unless the value is also the low-cardinality OTel error class for that span.

**OTEL-REQ-053**
Exception recording MUST NOT emit `exception.message` or `exception.stacktrace` in the base telemetry profile. `exception.type` MAY be emitted only when it is a low-cardinality class name that does not include incident, user, record, object, path, SQL, or secret material.

**OTEL-REQ-054**
Span events are disabled by default for Cartulary custom spans. A future revision that permits span events MUST define the exact event names, attributes, cardinality, ordering, and forbidden-value tests.

### 9.2 Sampler profile

OpenTelemetry marks `TraceIdRatioBased` as deprecated in favor of `ProbabilitySampler`, forbids SDK implementors from removing or modifying original `TraceIdRatioBased` behavior until at least January 1, 2027, and warns that exact TraceIdRatioBased algorithms were never specified for cross-SDK compatibility. `ProbabilitySampler` is Development-status in the inspected baseline.

**OTEL-REQ-055**
The current profile MUST use this sampler-profile table:

| `sample_ratio` | `sampler_profile='auto'` result | Required SDK construction | Notes |
| --- | --- | --- | --- |
| `0.0` | `cartulary.sampler.always_off.v1` | AlwaysOff or equivalent | Drops root spans while preserving safe no-op behavior. |
| `1.0` | `cartulary.sampler.always_on.v1` | AlwaysOn or equivalent | Records and samples all root spans. |
| `0.0 < ratio < 1.0` | `cartulary.sampler.traceidratio_compat.v1` | ParentBased with root TraceIdRatioBased-equivalent sampler; local-parent sampled -> AlwaysOn; local-parent not sampled -> AlwaysOff; remote parent branches unused because remote context is stripped | Bounded compatibility profile until a future revision adopts a ProbabilitySampler profile. |

**OTEL-REQ-056**
An explicit `telemetry.traces.sampler_profile` MUST be consistent with `sample_ratio`:

| Explicit profile | Valid `sample_ratio` values | Invalid condition |
| --- | --- | --- |
| `cartulary.sampler.always_off.v1` | exactly `0.0` | Any other ratio. |
| `cartulary.sampler.always_on.v1` | exactly `1.0` | Any other ratio. |
| `cartulary.sampler.traceidratio_compat.v1` | `0.0 < ratio < 1.0` | `0.0` or `1.0`. |

**OTEL-REQ-057**
`ProbabilitySampler`, `ComposableProbability`, `CompositeSampler`, `JaegerRemoteSampler`, adaptive sampling, remote sampling, sampler plugin providers, W3C Trace Context Level 2 randomness adoption, OpenTelemetry TraceState threshold propagation, and sampler environment variables are not adopted in this revision.

**OTEL-REQ-129**
The sampler profile metadata is closed by this table:

| Field | Required value |
| --- | --- |
| `sampler_profile_review_after` | `2027-01-01` |
| `sampler_profile_current_fractional` | `cartulary.sampler.traceidratio_compat.v1` |
| `probability_sampler_status` | Not adopted because the inspected upstream status is Development. |
| `migration_rule` | Any move to `ProbabilitySampler`, `ComposableProbability`, `CompositeSampler`, W3C Trace Context Level 2 randomness, or OpenTelemetry TraceState threshold propagation requires an NLSpec revision and a golden-corpus update. |

**OTEL-REQ-130**
A dependency update that silently replaces `TraceIdRatioBased` behavior with `ProbabilitySampler` or another probability-sampling implementation MUST be classified as `breaking_shape_change` unless a later NLSpec revision defines a new sampler profile and the golden corpus proves emitted-shape equivalence for the selected SDK version.

**OTEL-REQ-058**
Inbound remote parent sampling decisions MUST NOT control Cartulary root-span sampling. Because `telemetry.traces.accept_remote_context=false`, inbound `traceparent`, `tracestate`, Baggage, vendor headers, or sampled flags MUST be ignored before root span creation.

**OTEL-REQ-059**
Sampling tests MUST use a fixed corpus of server-owned trace IDs. For the traceidratio compatibility profile, the same trace ID and ratio MUST produce the same allow/drop decision within the selected SDK version. Cross-language equivalence is not claimed in this profile.

### 9.3 Required span families

**OTEL-REQ-060**
The implementation MUST emit spans for the following families when tracing is enabled and the selected sampler records the span:

| Span family | Span name format | Required attributes | Forbidden attributes |
| --- | --- | --- | --- |
| HTTP server | `<HTTP_METHOD> <route_template>` | Stable OTel HTTP allowlist from §9.4 plus `cartulary.route_family`, `cartulary.result`. | Raw URL path, raw query, headers, request body, response body, user/session/incident IDs. |
| Workbook query | `cartulary.workbook.query` | `cartulary.view_schema_id`, `cartulary.operation='query'`, `cartulary.result`. | Saved-view ID, filters, search text, row values, projection table name. |
| Workbook mutation | `cartulary.workbook.mutation` | `cartulary.view_schema_id`, `cartulary.record_type`, `cartulary.operation`, `cartulary.result`. | Record ID, row version, client transaction ID, field values. |
| Projection maintenance | `cartulary.workbook.projection` | `cartulary.view_schema_id`, `cartulary.operation`, `cartulary.result`. | Projection table name, SQL, row IDs. |
| WebSocket lifecycle | `cartulary.collaboration.websocket` | `cartulary.operation`, `cartulary.result`. | Connection ID, user ID, incident ID, payload content. |
| WebSocket event send | `cartulary.collaboration.event_send` | `cartulary.websocket.event_type`, `cartulary.result`. | Event payload, record ID, user ID, connection ID. |
| Job enqueue | `cartulary.jobs.enqueue` | `cartulary.job_kind`, `cartulary.operation='enqueue'`, `cartulary.result`. | Job ID, incident ID, request body. |
| Job run | `cartulary.jobs.run` | `cartulary.job_kind`, `cartulary.job_terminal_status`, `cartulary.result`. | Job ID, artifact path, incident ID, evidence ID. |
| Postgres dependency | `cartulary.postgres.operation` | `db.system.name='postgresql'`, `cartulary.operation`, `cartulary.result`. | SQL text, query summary, bind values, table names, database name, server address, port. |
| Object-store dependency | `cartulary.objectstore.operation` | `cartulary.operation`, `cartulary.result`. | Bucket, key, filename, object hash, upload ID, copy source, storage ref. |

### 9.4 HTTP server standard attribute allowlist

**OTEL-REQ-061**
HTTP server spans MAY emit only the following standard attributes:

| Attribute | Rule |
| --- | --- |
| `http.request.method` | HTTP method token only. Unknown methods MUST be mapped according to adopted stable OTel HTTP rules without raw user values. |
| `http.response.status_code` | Integer status code. |
| `http.route` | Route template only, never concrete path. |
| `url.scheme` | `http` or `https` only when derived from trusted server config. |
| `network.protocol.name` | `http` when available. |
| `network.protocol.version` | Protocol version only. |
| `error.type` | Low-cardinality OTel-compatible error class only. |

**OTEL-REQ-062**
HTTP telemetry MUST NOT emit `url.full`, `url.path`, `url.query`, `user_agent.original`, `client.address`, `server.address`, `server.port`, `http.request.header.*`, `http.response.header.*`, cookies, authorization headers, or route parameters.

### 9.5 Database span details

**OTEL-REQ-063**
Postgres spans MAY emit `db.system.name='postgresql'`. They MUST NOT emit `db.query.text`, `db.query.summary`, `db.response.status_code`, `db.namespace`, `db.collection.name`, SQL statement text, SQL parameters, table names, projection table names, index names, database address, database port, or connection string details.

**OTEL-REQ-064**
Database errors MUST be classified into `cartulary.error_class` using the global registry in §8.5.1. Database instrumentation MAY use only `dependency_unavailable`, `timeout`, `serialization_conflict`, `constraint_violation`, and `internal_error` unless a later adopted revision adds a database-specific class to the global registry.

### 9.6 Object-store non-adoption rule

**OTEL-REQ-065**
Cartulary MUST NOT adopt S3 semantic conventions in the current profile because they expose object-store identity and are Development-status in the observed semantic conventions. Object-store spans and metrics MUST use only Cartulary custom attributes listed in §8.5 plus safe numeric measurements such as duration and byte count.

### 9.7 Trace causality and links

**OTEL-REQ-066**
Span links MAY be emitted only for causal relationships that cross an async or batch boundary and do not expose product identifiers. Link attributes are forbidden in this revision. Linked SpanContexts MUST be server-owned contexts generated inside the same process or trusted deployment boundary.

**OTEL-REQ-067**
The implementation MUST NOT use links to encode incident IDs, job IDs, record IDs, evidence IDs, user IDs, request bodies, or queue item IDs.

### 9.8 Inbound trace context and Baggage

**OTEL-REQ-068**
Inbound `traceparent`, `tracestate`, Baggage, vendor trace headers, SQL commenter propagation, or other remote context carriers MUST NOT be extracted into active Cartulary telemetry context in this revision.

**OTEL-REQ-069**
Outbound propagation from Cartulary to Postgres, object storage, browser clients, or other dependencies MUST NOT inject trace context, Baggage, or SQL commenter fields unless a future revision defines an explicit trusted-boundary propagation profile.

## 10. Metrics contract

OpenTelemetry metric instruments have names, kinds, units, and descriptions. Instrument names are case-insensitive ASCII strings with syntax and length limits, and instrument units are case-sensitive ASCII strings with a maximum length of 63 characters.

### 10.1 Metric identity rules

**OTEL-REQ-070**
Every metric instrument MUST be declared in §10.4 before use. The declaration owns SDK instrument registration identity, exported metric-stream identity, description, allowed attributes, view filtering, aggregation, and temporality.

**OTEL-REQ-071**
Metric names MUST be lowercase ASCII, MUST satisfy OpenTelemetry instrument name syntax when used as SDK instrument names, MUST be no more than 255 characters, and MUST NOT collide after ASCII case-folding. A case-insensitive duplicate metric name is invalid even if the SDK would return a functional instrument.

**OTEL-REQ-072**
Metric units MUST be ASCII strings with maximum length 63. Duration metrics MUST use `s`. Byte metrics MUST use `By`. Count-like metrics MUST use `{request}`, `{operation}`, `{event}`, `{item}`, `{connection}`, `{job}`, or another exact unit declared in the registry.

**OTEL-REQ-073**
Metric identity MUST be evaluated using the two identity families in this table:

| Identity | Required Cartulary rule | Shape-change rule |
| --- | --- | --- |
| SDK instrument registration identity | Includes case-folded SDK instrument `name`, instrument kind, unit, and description because OpenTelemetry identifies instruments by those fields. The implementation MUST NOT register two SDK instruments whose upstream identity conflicts with or diverges from the §10.4 registry. | A change in this identity requires registry validation and dependency-update review. |
| Exported metric stream identity | Includes exported metric name, instrument kind, unit, aggregation, temporality, allowed attribute set, Resource identity, and instrumentation-scope context. | A change in this identity is classified by §4.5. |
| Diagnostic description | Description MUST be a static registry string. It MAY appear in SDK/exporter-specific metadata, but Cartulary does not treat it as privacy, routing, grouping, or operational authority. | A description-only change is not an emitted-shape change when exported telemetry shape remains otherwise registry-equivalent; it is still an SDK registration identity change. |

**OTEL-REQ-131**
Metric descriptions MUST NOT contain incident, deployment-host, operator, object-store, query, error-detail, path, or credential material. Duplicate metric tests MUST include same-name/different-case and same-name/same-kind/same-unit/different-description cases.

### 10.2 Histogram defaults

**OTEL-REQ-074**
Duration histogram metrics MUST use explicit bucket boundaries in seconds unless a metric registry row overrides them. The default duration boundaries are:

| Bucket sequence |
| --- |
| `0.005`, `0.010`, `0.025`, `0.050`, `0.100`, `0.250`, `0.500`, `1.000`, `2.500`, `5.000`, `10.000`, `30.000` |

**OTEL-REQ-075**
Byte histogram metrics MUST use explicit bucket boundaries in bytes unless a metric registry row overrides them. The default byte boundaries are:

| Bucket sequence |
| --- |
| `1024`, `4096`, `16384`, `65536`, `262144`, `1048576`, `4194304`, `16777216`, `67108864`, `268435456` |

### 10.3 Temporality and aggregation profile

OpenTelemetry MetricReader selection controls aggregation temporality, and temporality affects whether successive collections emit cumulative or delta data points.

**OTEL-REQ-076**
The current profile MUST use `cartulary.metrics.temporality.cumulative.v1`:

| Instrument kind | Aggregation temporality | Aggregation rule |
| --- | --- | --- |
| Counter | Cumulative | Monotonic sum. |
| UpDownCounter | Cumulative | Non-monotonic sum. |
| Histogram | Cumulative | Explicit bucket histogram unless registry row overrides. |
| ObservableGauge | Cumulative-equivalent current observation | Last value for each collection. |

**OTEL-REQ-077**
A change from cumulative to delta temporality, a change to histogram aggregation, a change to bucket boundaries, or a change from sync to async instrument kind is a `breaking_shape_change` unless OpenTelemetry and Cartulary registry rules both classify it as equivalent for the specific metric.

### 10.4 Required custom metric registry

**OTEL-REQ-078**
When metrics are enabled, the implementation MUST emit only the metric instruments in this registry:

| Metric name | Instrument kind | Unit | Description | Allowed attributes |
| --- | --- | --- | --- | --- |
| `cartulary.http.server.request.duration` | Histogram | `s` | HTTP server request duration. | `http.request.method`, `http.response.status_code`, `http.route`, `cartulary.route_family`, `cartulary.result`, optional `cartulary.error_code`. |
| `cartulary.workbook.query.duration` | Histogram | `s` | Workbook query duration. | `cartulary.view_schema_id`, `cartulary.operation`, `cartulary.result`, optional `cartulary.error_code`. |
| `cartulary.workbook.mutation.duration` | Histogram | `s` | Workbook mutation duration. | `cartulary.view_schema_id`, `cartulary.record_type`, `cartulary.operation`, `cartulary.result`, optional `cartulary.error_code`. |
| `cartulary.workbook.rows.returned` | Histogram | `{row}` | Rows returned by a workbook query. | `cartulary.view_schema_id`, `cartulary.result`. |
| `cartulary.collaboration.connections.active` | ObservableGauge | `{connection}` | Active accepted WebSocket connections. | `cartulary.result` omitted; no per-connection labels. |
| `cartulary.collaboration.events.sent` | Counter | `{event}` | WebSocket events sent. | `cartulary.websocket.event_type`, `cartulary.result`, optional `cartulary.drop_reason`. |
| `cartulary.jobs.active` | ObservableGauge | `{job}` | Active background jobs by kind. | `cartulary.job_kind`. |
| `cartulary.jobs.duration` | Histogram | `s` | Background job runtime duration. | `cartulary.job_kind`, `cartulary.job_terminal_status`, `cartulary.result`, optional `cartulary.error_code`. |
| `cartulary.postgres.operation.duration` | Histogram | `s` | Postgres dependency operation duration. | `db.system.name`, `cartulary.operation`, `cartulary.result`, optional `cartulary.error_class`. |
| `cartulary.objectstore.operation.duration` | Histogram | `s` | Object-store dependency operation duration. | `cartulary.operation`, `cartulary.result`, optional `cartulary.error_class`. |
| `cartulary.objectstore.transfer.bytes` | Histogram | `By` | Safe object-store transfer size. | `cartulary.operation`, `cartulary.result`; no object labels. |
| `cartulary.telemetry.export.failure` | Counter | `{failure}` | Telemetry export failures. | `cartulary.signal_kind`, `cartulary.telemetry.exporter_kind`, `cartulary.error_class`. |
| `cartulary.telemetry.item.dropped` | Counter | `{item}` | Telemetry items dropped before or during export. | `cartulary.signal_kind`, `cartulary.drop_reason`. |
| `cartulary.telemetry.queue.depth` | ObservableGauge | `{item}` | Current processor queue depth by signal. | `cartulary.signal_kind`. |

**OTEL-REQ-079**
Unknown metrics MUST NOT be emitted. Adding a metric, changing an instrument kind, changing a unit, changing allowed attributes, changing temporality, or changing aggregation MUST be classified under §4.5.

### 10.5 Metric view and attribute-filter contract

OpenTelemetry Views can customize metric stream names, aggregation, and metric attributes. Upstream specifies that a View-provided stream name is not required to conform to instrument-name syntax and must not be validated against that syntax by the SDK. Attributes removed by View configuration may still appear on exemplars unless exemplars are disabled.

**OTEL-REQ-080**
Every metric MUST have an SDK View or equivalent implementation-owned filter that retains only the allowed attributes in §10.4. All other attributes MUST be dropped before SDK export.

**OTEL-REQ-081**
Metric attributes MUST NOT include `record_id`, `incident_id`, `job_id`, `user_id`, `session_id`, `connection_id`, object-store identifiers, DB table names, HTTP header names or values, route parameters, free-text search, or exception detail.

**OTEL-REQ-132**
Metric View stream-name behavior MUST follow this table:

| View name condition | Required behavior |
| --- | --- |
| View omits stream `name` | Allowed. The SDK instrument name is used, and the exported stream must still match §10.4. |
| View supplies a stream `name` exactly equal to a §10.4 registered exported metric name for the selected instrument | Allowed only when the View's selector selects at most one registered SDK instrument and attribute filtering remains closed. |
| View supplies an unregistered stream `name` | Reject before provider activation or suppress before export. No unregistered metric stream may be exported. |
| View supplies a stream `name` with invalid instrument-name syntax | Do not rely on SDK rejection. Reject before provider activation unless the name is exactly a registered exported stream name in §10.4. |
| Wildcard View supplies a stream `name` | Forbidden in the current profile. Fail before provider activation. |

**OTEL-REQ-133**
Instrument advisory attributes and pre-bound metric attributes MUST NOT widen §10.4 allowed-attribute sets. If the selected SDK would retain attributes because a View omitted `attribute_keys` or because an instrument advisory parameter supplies broader attributes, Cartulary bootstrap MUST install a stricter View, wrap recording calls, or fail before provider activation.

### 10.6 SDK metric cardinality overflow

OpenTelemetry SDKs can synthesize an `otel.metric.overflow=true` attribute when cardinality limits are exceeded.

**OTEL-REQ-082**
The conformance corpus MUST produce zero metric data points with `otel.metric.overflow=true`.

**OTEL-REQ-083**
If `otel.metric.overflow=true` appears in emitted telemetry, the implementation MUST increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='metric_overflow'` when doing so does not recurse, and the test must fail unless the specific corpus case was written to verify overflow handling.

### 10.7 Exemplar policy

OpenTelemetry exemplars may include trace/span identifiers and filtered attributes; the SDK supports `AlwaysOff` as an exemplar filter.

**OTEL-REQ-084**
Metric exemplars MUST be disabled in the current profile. The effective exemplar filter MUST be `AlwaysOff` or an equivalent implementation. Exported metric points MUST NOT contain exemplars, exemplar trace IDs, exemplar span IDs, or filtered exemplar attributes.

**OTEL-REQ-085**
`OTEL_METRICS_EXEMPLAR_FILTER` and any SDK default exemplar filter MUST NOT re-enable exemplars.

**OTEL-REQ-134**
The OpenTelemetry Metric `Bind` API is not adopted as a Cartulary instrumentation mechanism in the current profile. Any bound or pre-bound attributes supplied by an SDK, helper, or future dependency MUST pass the same §8 attribute validation, null-omission, forbidden-value, and §10.4 allowed-attribute checks before recording. Bound attributes MUST NOT bypass wrappers, Views, or registry validation.

## 11. Logs and correlation

OpenTelemetry LogRecords define top-level fields including timestamps, trace context fields, severity fields, Body, Resource, InstrumentationScope, Attributes, and EventName. The Logs API supports emitting LogRecords and may accept an exception object and EventName.

### 11.1 Local structured-log fields

**OTEL-REQ-086**
Local structured logs MAY include trace correlation fields only in the following safe forms:

| Field | Rule |
| --- | --- |
| `trace_id` | Current server-owned trace ID only. MUST NOT be derived from inbound remote context. |
| `span_id` | Current server-owned span ID only. |
| `trace_flags` | May include sampled flag only. |
| `cartulary.module` | Registry value from §8.5. |
| `cartulary.route_family` | Registry value from §8.5. |
| `cartulary.operation` | Registry value from §8.5. |
| `cartulary.result` | Registry value from §8.5. |
| `cartulary.error_code` | Public error-code token only. |
| `cartulary.error_class` | Closed low-cardinality class only. |

**OTEL-REQ-087**
Local structured logs MUST NOT use trace correlation as a reason to add user identity, session identity, incident identity, record identity, job identity, object identity, or raw request details.

### 11.2 OTel LogRecord mapping

**OTEL-REQ-088**
When `telemetry.logs.bridge_enabled=false`, the implementation MUST NOT export OTel logs. Local logs may remain local application diagnostics.

**OTEL-REQ-089**
When `telemetry.logs.bridge_enabled=true`, exported OTel LogRecords MUST use this mapping:

| LogRecord field | Required mapping |
| --- | --- |
| `Timestamp` | Local log event timestamp when available. |
| `ObservedTimestamp` | Time the bridge observes the log. |
| `TraceId` | Current server-owned trace ID when present. |
| `SpanId` | Current server-owned span ID when present. |
| `TraceFlags` | Current server-owned trace flags when present. |
| `SeverityNumber` | Mapping in §11.3. |
| `SeverityText` | Local severity token after mapping; maximum 32 ASCII characters. |
| `Body` | String only. Redacted before construction. Bounded by `telemetry.logs.body_max_chars`. |
| `Resource` | Closed resource from §7. |
| `InstrumentationScope` | Registered scope from §5.2. |
| `Attributes` | Only attributes allowed by §8.5 and §11.1. |
| `EventName` | MUST be omitted. |

**OTEL-REQ-090**
The log bridge MUST NOT pass raw exception objects to the OTel Logs API and MUST NOT supply the OTel Logs API optional `Exception` parameter. Local application logs MAY carry internal exception objects, but the OTel bridge MUST reduce exception facts before OTel API emission to safe fields only:

| Candidate fact | Current-profile treatment |
| --- | --- |
| `cartulary.error_class` | Allowed when the value is in the closed low-cardinality registry. |
| `exception.type` | Allowed only when it is a low-cardinality class name that does not include incident, user, record, object, path, SQL, or secret material. |
| exception message | Forbidden. |
| stacktrace | Forbidden. |
| cause chain or wrapped error string | Forbidden. |
| panic string | Forbidden. |
| SQL snippet, filesystem path, object-store key, request body, incident-authored text | Forbidden. |

**OTEL-REQ-091**
The implementation MUST NOT emit `exception.message`, `exception.stacktrace`, exception cause chains, panic strings, SQL snippets, file paths, object keys, request bodies, or incident-authored text through LogRecord Body, Attributes, EventName, optional `Exception` parameter, or span events.

**OTEL-REQ-092**
No log-event-to-span-event bridge may be installed in this revision. A LogRecord MUST NOT create a span event merely because it has an event-like name, error class, severity, or local logger event type.

### 11.3 Severity mapping

**OTEL-REQ-093**
Local severity mapping MUST be exactly this table:

| Local severity | OTel `SeverityNumber` | OTel range | `SeverityText` |
| --- | ---: | --- | --- |
| `trace` | 1 | TRACE | `TRACE` |
| `debug` | 5 | DEBUG | `DEBUG` |
| `info` | 9 | INFO | `INFO` |
| `warn` | 13 | WARN | `WARN` |
| `error` | 17 | ERROR | `ERROR` |
| `fatal` | 21 | FATAL | `FATAL` |
| unknown or omitted | 9 | INFO | `INFO` |

**OTEL-REQ-094**
A local severity string outside the closed table MUST map to unknown or omitted behavior. It MUST NOT be exported as arbitrary `SeverityText`.

## 12. Exporter, processor, runtime, and shutdown behavior

### 12.1 Intentional divergence from OTel exporter defaults

OpenTelemetry OTLP exporters define default endpoints, per-signal endpoint precedence, OTLP/HTTP path construction, supported compression values, retry concepts, and User-Agent guidance. Cartulary intentionally uses stricter rules to avoid environment-driven egress and signal divergence.

**OTEL-REQ-095**
Cartulary MUST NOT rely on OTel exporter default endpoints. Network export occurs only when `telemetry.exporter.kind` is `otlp_http` or `otlp_grpc` and `telemetry.exporter.endpoint` is explicitly configured.

**OTEL-REQ-096**
Cartulary MUST NOT support per-signal endpoint, per-signal header, or per-signal protocol divergence in this revision. Trace, metric, and log export MUST use one configured endpoint family and one configured header map.

### 12.2 OTLP endpoint normalization

**OTEL-REQ-097**
For `telemetry.exporter.kind='otlp_http'`, the implementation MUST construct signal URLs from the configured base endpoint by appending signal paths exactly as shown:

| Configured endpoint | Trace URL | Metrics URL | Logs URL |
| --- | --- | --- | --- |
| `http://collector:4318` | `http://collector:4318/v1/traces` | `http://collector:4318/v1/metrics` | `http://collector:4318/v1/logs` |
| `http://collector:4318/` | `http://collector:4318/v1/traces` | `http://collector:4318/v1/metrics` | `http://collector:4318/v1/logs` |
| `http://collector:4318/otel` | `http://collector:4318/otel/v1/traces` | `http://collector:4318/otel/v1/metrics` | `http://collector:4318/otel/v1/logs` |
| `http://collector:4318/otel/` | `http://collector:4318/otel/v1/traces` | `http://collector:4318/otel/v1/metrics` | `http://collector:4318/otel/v1/logs` |

**OTEL-REQ-098**
For `telemetry.exporter.kind='otlp_grpc'`, the implementation MUST use the configured endpoint as the gRPC target. It MUST NOT derive per-signal gRPC endpoints from environment variables.

**OTEL-REQ-099**
OTLP/HTTP JSON export is unsupported in this revision. OTLP/HTTP MUST use `http/protobuf`. OTLP/gRPC MUST use protobuf over gRPC.

### 12.3 Header and User-Agent policy

**OTEL-REQ-100**
Exporter request headers MUST come only from `telemetry.exporter.headers` plus protocol-required safe headers. Headers MUST NOT be read from OTel environment variables, browser state, request headers, incident records, or user-controlled workbook values.

**OTEL-REQ-101**
Exporter headers and header values are secret-bearing unless proven otherwise. They MUST be redacted from logs, spans, metrics, diagnostics, retained artifacts, and test reports.

**OTEL-REQ-102**
Exporter `User-Agent` MAY identify the Cartulary application version, OpenTelemetry exporter language, and OpenTelemetry exporter version. It MUST NOT include incident identifiers, deployment hostnames, usernames, object-store identifiers, environment names, runtime roots, local paths, or arbitrary operator-provided text.

**OTEL-REQ-103**
The preferred User-Agent form is:

| Segment | Required rule |
| --- | --- |
| `Cartulary/<service.version>` | Present when `service.version` is known; otherwise `Cartulary/0.0.0+unknown`. |
| `OTel-OTLP-Exporter-<language>/<version>` | Present when the SDK/exporter exposes this identity. |
| Additional product identifiers | Forbidden in this revision unless a future revision defines them. |

### 12.4 Retry contract

**OTEL-REQ-104**
Exporter retries MUST be bounded by the configured retry keys in §6.1. Retries MUST NOT block HTTP responses, WebSocket sends, workbook writes, evidence access, or background-job state transitions beyond normal telemetry-recording overhead.

**OTEL-REQ-105**
Exporter retry behavior MUST follow this deterministic envelope:

| Retry element | Required behavior |
| --- | --- |
| First export attempt | Occurs without retry delay. It remains bounded by `telemetry.processor.export_timeout_ms`. |
| Retry attempt index | First retry after the initial failed attempt uses `retry_attempt_index=1`. |
| Base interval | `base_interval_ms = min(max_interval_ms, initial_interval_ms * multiplier^(retry_attempt_index - 1))`. |
| Jitter mode | Full jitter. Runtime delay is sampled uniformly from integer milliseconds in `[0, base_interval_ms]`. |
| Randomness source | Runtime uses a process CSPRNG or equivalent strong random source. Conformance tests assert delay bounds and terminal behavior rather than exact runtime samples. |
| Max elapsed cutoff | The implementation MUST NOT start a retry if `elapsed_since_first_failed_attempt_start + sampled_delay_ms > max_elapsed_ms`. |
| `max_elapsed_ms=0` | Disables retries even when retry is enabled. |
| Shutdown | After shutdown begins, no new retry loop may be started. A retry already in progress may finish only within the shutdown flush timeout. |
| Permanent rejection | No retry; drop and record `exporter_permanent_discard` when non-recursive. |
| Timeout | Each attempt aborts at `telemetry.processor.export_timeout_ms`; timeout is classified as transient only when the transport or status family is transient. |
| Disabled export | No network exporter exists when `telemetry.exporter.kind='none'`; any export attempt in this state is an implementation defect. |

No product configuration key, browser state, public route, or production-visible test hook may inject deterministic runtime retry randomness. Unit tests MAY exercise a pure retry-planner helper with supplied sample values, but that helper MUST NOT be a runtime configuration surface.

**OTEL-REQ-135**
The retry loop MUST use the transient-versus-permanent classification exposed by the selected OTLP transport. It MUST treat retryable OTLP/gRPC status codes and retryable OTLP/HTTP status codes as transient, and it MUST treat permanent rejections as terminal. The implementation MUST NOT retry merely because a product operation failed independently of export.

### 12.5 Processor overflow contract

**OTEL-REQ-106**
Each enabled signal processor MUST enforce `telemetry.processor.max_queue_size`. When the queue is full, the implementation MUST apply `drop_new`: retain already queued telemetry, drop the newly offered telemetry item, and increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='queue_full'` when non-recursive.

**OTEL-REQ-107**
Processor overflow MUST NOT block product operations waiting for telemetry queue capacity.

### 12.6 Error-handling matrix

OpenTelemetry error-handling guidance distinguishes initialization failures from runtime telemetry failures and states that telemetry should not significantly change application behavior.

**OTEL-REQ-108**
Telemetry errors MUST follow this table:

| Error point | Product behavior | Telemetry behavior |
| --- | --- | --- |
| Invalid deployment telemetry config | Fail startup before readiness. | Emit only startup diagnostics permitted by Core 04 and this NLSpec. |
| SDK provider construction failure after valid config | Fail startup before readiness. | Emit only local startup diagnostics. |
| Exporter endpoint unavailable | Product remains available. | Export attempts fail, retry/drop according to §12.4. |
| Processor queue full | Product remains available. | Drop new telemetry item and increment drop metric when non-recursive. |
| Redaction cannot prove safety | Product remains available unless the product operation itself failed independently. | Drop telemetry item and increment drop metric when non-recursive. |
| Log bridge mapping failure | Product remains available. | Drop the LogRecord and increment drop metric when non-recursive. |
| Shutdown flush timeout | Continue shutdown after timeout. | Increment drop metric or local diagnostic if non-recursive and possible. |
| Telemetry self-diagnostic error | Product remains available. | Suppress recursive telemetry. |

**OTEL-REQ-136**
For runtime telemetry failures, product route results, workbook mutation commit behavior, WebSocket event-send behavior, evidence-handle issuance and redemption behavior, and background-job state transitions MUST match the same scenario with telemetry export disabled, except for bounded local diagnostics and telemetry self-metrics. This invariant applies during exporter endpoint unavailability, exporter timeout, retry backoff, queue overflow, redaction rejection, log bridge mapping failure, and self-diagnostic failure. Invalid telemetry configuration and SDK provider construction failure remain startup failures before readiness.

### 12.7 Startup and shutdown contract

**OTEL-REQ-109**
Telemetry bootstrap MUST occur after deployment configuration validation and before readiness. Providers MUST NOT be activated until resource identity, sampler profile, processors, metric readers, exporters, log bridge, environment containment, and forbidden resource attributes have been validated.

**OTEL-REQ-110**
On graceful shutdown, the application MUST attempt to flush enabled telemetry for at most `telemetry.shutdown.flush_timeout_ms`. If the timeout expires, shutdown MUST continue. Telemetry loss during shutdown MUST NOT prevent process termination.

**OTEL-REQ-111**
Shutdown MUST call provider, processor, metric-reader, and exporter shutdown exactly once per active SDK provider instance. Repeated shutdown calls MUST be idempotent from the application perspective.

### 12.8 Telemetry-about-telemetry recursion guard

**OTEL-REQ-112**
Telemetry self-diagnostics MUST NOT recursively produce unbounded telemetry. When self-diagnostic telemetry would cause telemetry about telemetry about telemetry, the implementation MUST drop the recursive item and, when possible without recursion, increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='recursion_guard'`.

## 13. Browser telemetry boundary and non-transfer rules

### 13.1 Browser boundary

**OTEL-REQ-113**
Browser code MUST NOT initialize an OpenTelemetry SDK exporter, OTLP exporter, vendor SDK, remote log exporter, remote metrics exporter, remote traces exporter, session replay SDK, or third-party analytics client in the current profile.

**OTEL-REQ-114**
Browser code MAY create local performance marks or local diagnostic events only when all of the following are true:

| Condition | Required behavior |
| --- | --- |
| Same-origin only | Marks stay in browser memory or are sent to the Cartulary application through an explicitly specified product diagnostic route in a future revision. |
| No incident content | Marks contain no row values, field values, evidence values, query text, object keys, filenames, or identifiers. |
| No direct export | Browser does not send marks to OTLP, vendor, Collector, or third-party endpoints. |
| No persistence surprise | Marks do not persist across sessions unless a future revision defines the persistence contract. |

**OTEL-REQ-115**
No browser local-storage, session-storage, IndexedDB, Service Worker, cookie, DOM attribute, React prop, grid vendor coordinate, or URL parameter may configure telemetry exporters, headers, endpoints, samplers, processors, metric readers, resource attributes, or log bridges.

### 13.2 Concepts not transferred into Cartulary

**OTEL-REQ-116**
The OpenTelemetry concepts in this table MUST NOT be transferred into current-profile Cartulary behavior:

| OTel concept | Current-profile rule |
| --- | --- |
| Required Collector deployment | Forbidden as conformance requirement. Collector MAY be external receiver only. |
| Collector-side privacy enforcement | Forbidden. Privacy must hold before SDK processor/exporter handoff. |
| Baggage correlation | Forbidden. Baggage values MUST NOT become spans, metrics, logs, resource attributes, or local correlation keys. |
| Environment-driven autoconfiguration | Forbidden. Cartulary deployment config remains authoritative. |
| Declarative configuration plugin providers | Forbidden as runtime authority. They must not instantiate exporters, processors, readers, samplers, resource detectors, views, log bridges, exemplar filters, or plugins. |
| Default OTLP localhost export | Forbidden. Export must be explicit. |
| Per-signal exporters/endpoints | Forbidden. Single exporter profile applies across enabled signals. |
| Prometheus scrape endpoint | Not adopted in current profile. |
| Prometheus Metrics Exporter compatibility | Not adopted in current profile. Upstream stabilization does not create a Cartulary Prometheus endpoint or scrape contract. |
| Zipkin or Jaeger-native export | Not adopted in current profile. |
| Vendor-native exporters | Not adopted in current profile. |
| SQL commenter propagation | Forbidden. |
| Host/process/container/cloud resource detectors | Forbidden. |
| OTel Resource schema-url merge semantics | Not adopted as dynamic behavior. Current profile uses empty effective Resource schema URL unless a future revision pins one exact schema URL. |
| S3 semantic conventions | Not adopted in current profile. |
| Metric exemplars | Disabled in current profile. |
| Metric `Bind` API | Not adopted in current profile. Bound attributes must not bypass Cartulary attribute validation or registry allowlists. |
| Log `EventName` | Omitted in current profile. |
| Logs API optional `Exception` parameter | Not adopted. Exception facts must be reduced to safe allowlisted attributes before OTel API emission. |
| Log-event-to-span-event bridge | Forbidden. |
| ProbabilitySampler | Not adopted until a future revision defines a profile. |

## 14. Golden telemetry corpus and drift verification

### 14.1 Corpus contract

**OTEL-REQ-117**
The repository MUST maintain a golden telemetry conformance corpus for this NLSpec.

The corpus artifact paths are fixed by this table:

| Artifact family | Required path or location |
| --- | --- |
| Committed normalized golden telemetry | `internal/testutil/golden/otel/` |
| Committed source-snapshot fixture | `internal/testutil/golden/otel/source-snapshot/` |
| Committed emitted-shape fixtures | `internal/testutil/golden/otel/signals/` |
| Retained raw captures | Target-owned retained artifact directory under the harness run root, not committed source. |
| Retained normalized comparison summary | Target-owned retained artifact directory under the harness run root. |

Raw captures MUST NOT be written under committed golden directories. Harness-owned run-root identity and cleanup behavior are owned by `docs/testing-harness-nlspec.md` §6.

The corpus MUST exercise at least:

| Corpus area | Required coverage |
| --- | --- |
| No-SDK mode | Representative HTTP, workbook, job, Postgres, object-store, log-call, and WebSocket paths execute without SDK providers and without telemetry-induced errors. |
| Source baseline | Immutable OTel spec and semantic-convention refs, full commit SHAs, source paths, document statuses, semantic-convention model digest, generated constants version, and SDK package versions. |
| Cartulary environment binding parser | Omitted, empty, valid, invalid, and explicit `"null"` fixtures for boolean, integer, decimal, enum, URL, header-map, and string key families. |
| Hostile environment | Every environment family in §6.4 is set to a hostile value and does not alter effective telemetry behavior. |
| Hostile declarative configuration | Exporter, resource detector, sampler, metric reader, View, exemplar, log bridge, and plugin-provider instantiation attempts in §6.3 have no runtime effect. |
| HTTP route shape | Route-template spans and metrics appear without raw paths, query strings, headers, or IDs. |
| Workbook query and mutation | Query, create, patch, conflict, projection update, and row-refresh telemetry appears with safe attributes only. |
| WebSocket collaboration | Connect, authorize, subscribe, event-send, replay, close, overflow, and unauthorized subscribe paths. |
| Background jobs | Enqueue, run success, cancellation, failure, timeout, and terminal-state metrics. |
| Postgres dependency | Success, timeout, unavailable, serialization conflict, and constraint-violation classes. |
| Object storage | Upload target, attach, preview/download handle issuance, object-store unavailable, transfer size, and forbidden object identifiers. |
| Resource identity | Closed resource attributes, empty Resource `schema_url`, SDK default/resource-detector suppression, and conflicting non-empty schema URL rejection. |
| Attribute boundary | Null-like internal values are omitted before OTel API calls across spans, metrics, logs, resource attributes, and self-diagnostics. |
| Logs | Bridge disabled, bridge enabled, safe severity mapping, exception with forbidden values, omitted EventName, no optional `Exception` parameter, and no span-event bridge. |
| Metrics | SDK registration identity, exported stream identity, case-insensitive duplicate rejection, divergent-description duplicate rejection, temporality, View name rejection, attribute filters, no Metric `Bind` bypass, no exemplars, and no `otel.metric.overflow`. |
| Exporter | Export disabled, OTLP/HTTP path construction, OTLP/gRPC target construction, headers, User-Agent, full-jitter retry envelope, timeout, and shutdown. |
| Product-boundary runtime invariance | Representative HTTP request, workbook query, workbook mutation, WebSocket send, evidence access, and background-job transition match the no-export baseline under exporter failure, timeout, queue overflow, and redaction rejection. |
| Redaction | Representative forbidden values from every family in §8.4 across spans, metrics, logs, self-diagnostics, retained artifacts, and exporter attempts. |

### 14.2 Canonical telemetry-shape normalization

**OTEL-REQ-118**
Golden telemetry comparison MUST normalize volatile values before comparison and MUST NOT normalize away shape facts. The normalization table is:

| Field family | Normalization rule |
| --- | --- |
| Timestamps and durations | Replace wall-clock timestamps with ordinal placeholders; retain metric duration buckets and numeric measurements needed for assertions. |
| Trace IDs and span IDs | Replace with stable placeholders preserving parent/child/link topology. |
| `service.instance.id` | Replace with placeholder after validating opacity and registry presence. |
| Attribute order | Sort by attribute key. |
| Resource records | Compare as sorted key/value maps after applying allowed placeholder normalization. |
| Spans | Sort by deterministic corpus operation order, then span name, then normalized parent placeholder. |
| Metrics | Sort by metric name, instrument kind, unit, temporality, data point attributes, and collection ordinal. |
| Logs | Sort by deterministic corpus operation order, timestamp placeholder, severity, and normalized trace/span placeholder. |
| Forbidden values | MUST NOT be normalized into safety. Any forbidden literal observed before normalization fails the test. |
| Unknown fields | Unknown emitted telemetry fields fail unless the owning registry allows additive fields and §4.5 classifies the change. |

### 14.3 Dependency update gate

**OTEL-REQ-119**
A dependency update to OTel API, SDK, exporter, semantic-convention constants, resource detector packages, log bridge packages, or instrumentation adapters MUST run the golden telemetry corpus and produce one of the §4.5 change classifications before merge.

**OTEL-REQ-120**
The update gate MUST compare span names, span kinds, span status policy, span events, links, SDK instrument registration identity, exported metric-stream identity, metric temporality, aggregation, View stream-name behavior, exemplar behavior, log top-level fields, Logs API exception-parameter behavior, resource attributes, Resource schema URL, instrumentation-scope names, instrumentation-scope schema URLs, instrumentation-scope attributes, sampler profile behavior, retry envelope, standard attributes, custom attributes, User-Agent, endpoint paths, non-transfer rules, and forbidden-value exclusions.

## 15. Verification and acceptance criteria

### 15.1 Configuration and source-baseline criteria

- **OTEL-AC-001:** The repo contains a complete `otel_source_snapshot` with `schema_id='cartulary.otel_source_snapshot.v1'`, immutable OTel spec ref `v1.57.0`, full OTel spec commit SHA, semantic-conventions ref `v1.41.0`, full semantic-conventions commit SHA, exact source paths from §4.2, document statuses, `semconv_model_digest_algorithm='semconv_model_digest_v1'`, semantic-convention model digest `3f8f80a2ed04521dfe29e50fcddd7f7de70145a6aee01959f985a65fbb4c8632`, generated constants provenance, and exhaustive SDK package-family rows.
- **OTEL-AC-002:** Conformance mode fails on `main`, unresolved source-snapshot placeholder values, short-only commit SHA, missing source path, duplicated source path, unknown source path, missing document status, missing model digest, missing generated constants provenance, missing SDK package family, or prohibited browser/runtime OTel package family.
- **OTEL-AC-003:** Unknown `telemetry.*` keys fail deployment-config validation.
- **OTEL-AC-004:** Invalid cross-key combinations in §6.5 fail before readiness.
- **OTEL-AC-004A:** Cartulary server-side environment parser fixtures cover omitted, empty, valid, invalid, and explicit `"null"` cases for every parser family in §6.2.1; invalid cases fail before readiness and expose no secret values.
- **OTEL-AC-005:** With `telemetry.exporter.kind='none'`, no network exporter is created and no OTel default localhost endpoint is contacted.
- **OTEL-AC-006:** With every §6.4 environment variable and every hostile declarative-config fixture in §6.3 present, export remains disabled when Cartulary export is disabled, resource attributes remain closed, Resource schema URL remains empty, propagators remain closed, exemplars remain disabled, no plugin provider loads, and no declarative-config component becomes active.

### 15.2 API, SDK, and instrumentation-boundary criteria

- **OTEL-AC-007:** Static dependency checks prove ordinary instrumentation units import only OTel API packages and telemetry bootstrap accessors, not SDK, exporter, processor, sampler, propagator, metric-reader, log-processor, resource-detector, autoconfiguration, declarative-configuration, or plugin-provider packages; the check also proves `OpenTelemetry Core package` terminology is not used as permission for ordinary instrumentation imports.
- **OTEL-AC-008:** A no-SDK runtime profile exercises representative HTTP, workbook, job, Postgres, object-store, log-call, and WebSocket paths without telemetry-induced errors.
- **OTEL-AC-009:** Every tracer, meter, and logger uses a registered instrumentation scope with version, null schema URL, and empty scope attributes.

### 15.3 Resource and attribute criteria

- **OTEL-AC-010:** Exported telemetry contains only resource attributes listed in §7.1 plus SDK-required `telemetry.sdk.*` values, and the effective Resource `schema_url` is empty.
- **OTEL-AC-011:** With SDK resource detector packages present and `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_SERVICE_NAME`, and `OTEL_ENTITIES` set, including conflicting non-empty Resource schema URLs, exported telemetry contains no host, process, container, Kubernetes, cloud, entity, filesystem, object-store, or non-empty schema URL resource state.
- **OTEL-AC-012:** The forbidden-value corpus proves representative forbidden values are absent from spans, metrics, logs, resource attributes, self-diagnostics, retained telemetry artifacts, exporter requests, and User-Agent.
- **OTEL-AC-013:** Unknown `cartulary.*` attributes are rejected or omitted before recording.
- **OTEL-AC-014:** SDK truncation or attribute limits are not used as redaction controls; forbidden values fail before any SDK limit can truncate them.
- **OTEL-AC-014A:** Instrumentation-wrapper tests use an OTel API spy or emitted telemetry capture to prove no null attribute setter call occurs for representative spans, metrics, logs, resource attributes, and self-diagnostics.

### 15.4 Trace and span criteria

- **OTEL-AC-015:** `sample_ratio=0.0`, `1.0`, and a fractional ratio each select the exact sampler profile defined in §9.2, and the source snapshot or generated config exposes `sampler_profile_review_after='2027-01-01'`.
- **OTEL-AC-016:** A fixed server-owned trace-ID corpus proves deterministic allow/drop behavior for the selected SDK version and fractional ratio without claiming cross-language equivalence.
- **OTEL-AC-017:** Inbound `traceparent`, `tracestate`, Baggage, and vendor trace headers do not alter root-span sampling, trace IDs, span parentage, attributes, logs, or metrics.
- **OTEL-AC-018:** HTTP spans emit route templates only and never concrete paths, route parameter values, query strings, headers, cookies, user agents, client IPs, or body content.
- **OTEL-AC-019:** Postgres spans emit `db.system.name='postgresql'` and never emit SQL text, query summaries, bind values, table names, schema names, projection names, DB namespace, server address, or server port.
- **OTEL-AC-020:** Object-store telemetry never emits S3 semantic attributes or object-identifying values.

### 15.5 Metrics criteria

- **OTEL-AC-021:** Every emitted metric matches §10.4 by exported metric-stream identity: exported metric name, instrument kind, unit, allowed attributes, aggregation, temporality, Resource identity, and instrumentation-scope context.
- **OTEL-AC-022:** Registry validation rejects case-insensitive duplicate metric names and duplicate SDK registration identity divergence, including same-name/same-kind/same-unit/different-description cases.
- **OTEL-AC-023:** Metric View tests reject or suppress unregistered View stream names, invalid View stream names, wildcard Views with stream names, Views that add non-allowlisted attributes, and Views that omit required attribute filtering where instrument advisory attributes would retain too many attributes.
- **OTEL-AC-024:** The conformance corpus emits no exemplars, no exemplar trace IDs, no exemplar span IDs, and no exemplar filtered attributes; `OTEL_METRICS_EXEMPLAR_FILTER` attempts to enable exemplars have no effect.
- **OTEL-AC-025:** The conformance corpus emits no `otel.metric.overflow=true` datapoint except an explicit overflow-negative test, and that negative test records `cartulary.drop_reason='metric_overflow'` when non-recursive.
- **OTEL-AC-026:** Metric temporality is cumulative for every current-profile instrument.
- **OTEL-AC-026A:** Metric `Bind` or pre-bound attribute paths are absent from Cartulary instrumentation or are proven unable to bypass §8 and §10.4 validation.

### 15.6 Logs criteria

- **OTEL-AC-027:** With `telemetry.logs.bridge_enabled=false`, no OTel LogRecords are exported.
- **OTEL-AC-028:** With log bridge enabled, LogRecords use only the top-level fields and mapping rules in §11.2.
- **OTEL-AC-029:** A local log containing an exception whose message, stacktrace, cause, and wrapped text include incident-like strings, SQL-like strings, object-store-like strings, and filesystem paths exports none of those values.
- **OTEL-AC-030:** The log bridge never passes raw exception objects, never supplies the OTel Logs API optional `Exception` parameter, never emits `EventName`, and never emits exception message, stacktrace, cause chain, or wrapped error detail.
- **OTEL-AC-031:** No log-event-to-span-event bridge is installed; log records do not create span events.

### 15.7 Exporter, processor, runtime, and shutdown criteria

- **OTEL-AC-032:** OTLP/HTTP exporter tests inspect actual request URLs and prove the exact path construction in §12.2 for traces, metrics, and logs.
- **OTEL-AC-033:** OTLP/gRPC exporter tests prove one configured target and no per-signal endpoint divergence.
- **OTEL-AC-034:** Exporter request headers contain only configured safe headers and protocol-required headers; secret header values are absent from logs, spans, metrics, diagnostics, and retained artifacts.
- **OTEL-AC-035:** User-Agent includes only allowed Cartulary and exporter identity segments and contains no forbidden value families.
- **OTEL-AC-036:** Exporter retry tests assert full-jitter retry bounds, max-elapsed cutoff, disabled retry behavior, `max_elapsed_ms=0`, permanent rejection behavior, timeout behavior, shutdown interaction, no production-visible deterministic runtime RNG hook, and no product hot-path blocking.
- **OTEL-AC-037:** Processor queue overflow uses `drop_new`, does not block product work, and records drop metrics when non-recursive.
- **OTEL-AC-038:** Shutdown flush respects `telemetry.shutdown.flush_timeout_ms` and continues process shutdown after timeout.
- **OTEL-AC-039:** Telemetry self-diagnostics do not recurse unboundedly.
- **OTEL-AC-039A:** Black-box tests run representative HTTP request, workbook query, workbook mutation, WebSocket send, evidence access, and background-job transition scenarios under exporter failure, timeout, queue overflow, and redaction rejection; each must match the no-telemetry-export baseline for product response and committed state.

### 15.8 Browser and non-transfer criteria

- **OTEL-AC-040:** Browser bundles contain no OTLP exporter, vendor telemetry SDK, Collector client, session replay SDK, or third-party analytics initialization.
- **OTEL-AC-041:** Browser state cannot configure telemetry exporters, headers, endpoints, samplers, processors, resource attributes, or log bridges.
- **OTEL-AC-042:** Required Collector deployment, Collector-side privacy enforcement, Baggage correlation, environment-driven autoconfiguration, declarative plugin providers, default OTLP localhost export, per-signal exporters/endpoints, Prometheus scrape/exporter compatibility, Zipkin export, Jaeger-native export, vendor-native export, SQL commenter propagation, resource detectors, Resource schema-url merge semantics, S3 semantic attributes, exemplars, Metric `Bind` bypass, LogRecord EventName, Logs API optional `Exception` parameter, log-event-to-span-event bridging, and ProbabilitySampler are absent from runtime behavior and emitted telemetry.

### 15.9 Golden-corpus and drift criteria

- **OTEL-AC-043:** Golden telemetry normalization preserves shape facts while replacing only volatile identifiers and timestamps listed in §14.2.
- **OTEL-AC-044:** Any OTel dependency, SDK, exporter, semantic-convention, generated-constant, resource-detector, log-bridge, instrumentation-adapter, metric-view, sampler, or retry-behavior update runs the golden corpus and records a §4.5 classification before merge. The comparison includes SDK instrument registration identity, exported metric-stream identity, Resource schema URL, retry envelope, non-transfer rules, and sampler profile behavior.
- **OTEL-AC-045:** A dependency-only update can be accepted without NLSpec revision only when the normalized corpus is `registry_equivalent`.
- **OTEL-AC-046:** Any `additive_non_breaking`, `privacy_tightening`, or `breaking_shape_change` result requires NLSpec revision and updated acceptance criteria before adoption. Sampler migration, Resource schema-url adoption, View-renaming change, Prometheus exporter adoption, or any non-registry-equivalent change is non-conformant without such revision.

## 16. Resolved adoption decisions

The following adoption decisions are resolved for this revision. A later revision that reopens any row MUST keep this NLSpec non-adopted until the replacement owner section is complete.

| Decision ID | Closure status | Closure owner |
| --- | --- | --- |
| `OTEL-DQ-001` | resolved | §4.1.1 and concrete `otel_source_snapshot.semconv_model_digest`. |
| `OTEL-DQ-002` | resolved | §4.1.3 and repo-control generated constants evidence. |
| `OTEL-DQ-003` | resolved | §4.1.4 and repo-control dependency manifests. |
| `OTEL-DQ-004` | resolved | §8.5.1 public-code/class mapping. |
| `OTEL-DQ-005` | resolved | §14.1 committed and retained corpus paths. |
| `OTEL-DQ-006` | resolved | §12.4 bounds-only runtime retry-jitter conformance. |

The OTel specification baseline and semantic-convention release baseline are not reopenable implementation choices in this revision; they are pinned in §4.1. A later rebaseline requires §4.5 classification and §15 acceptance updates.

## 17. Completion standard

This NLSpec is complete for implementation when all of the following are true:

- The OpenTelemetry source snapshot is pinned to `v1.57.0` and full commit SHA `d4a91bddb53b4c308df3e40171a60059183efd88`.
- The semantic-conventions source snapshot is pinned to `v1.41.0` and full commit SHA `e018fe6f91862f5ed63c082f87697cddac596784`.
- Configuration source paths are exact, not wildcarded.
- The semantic-convention model digest, generated constants provenance, and language SDK package-family rows are pinned in repo-control files.
- OpenTelemetry `Core packages` terminology cannot be confused with Cartulary Core 00 through Core 05.
- Ordinary instrumentation units are API-only and SDK-independent.
- No-SDK mode executes representative product paths without telemetry-induced errors.
- Cartulary deployment configuration is the sole telemetry behavior authority.
- Cartulary environment-binding parsing is closed, deterministic, and fail-closed.
- Hostile OTel environment variables cannot enable egress, alter resources, alter sampling, enable Baggage, enable exemplars, load declarative config, or instantiate plugin providers.
- Declarative configuration and plugin-provider hostile fixtures cannot instantiate runtime components.
- Resource attributes are closed, resource detectors are disabled, effective Resource `schema_url` is empty, and conflicting non-empty schema URLs fail before provider activation.
- Null internal values are omitted before OTel API calls.
- Forbidden values are removed before recording, not by SDK limits or downstream scrubbing.
- The sampler profile is explicit, has review metadata `sampler_profile_review_after='2027-01-01'`, and inbound remote context cannot control Cartulary root spans.
- HTTP, Postgres, object-store, workbook, collaboration, jobs, metrics, and logs emit only registered signal shapes.
- Metric SDK registration identity and exported metric-stream identity are separately defined and tested.
- Metric Views cannot create unregistered exported streams, wildcard-renamed streams, or attribute widening.
- Metrics have fixed identity, allowed attributes, cumulative temporality, no Metric `Bind` bypass, and no exemplars.
- Logs are disabled by default; when enabled they emit bounded, string-only, redacted LogRecords with no EventName, no raw exception object, and no optional Logs API `Exception` parameter.
- OTLP export requires explicit Cartulary configuration and uses deterministic endpoint, header, User-Agent, retry, timeout, and shutdown rules.
- Retry jitter is bounded, testable, and non-blocking for product hot paths.
- Runtime telemetry failures do not alter product-visible behavior or committed state.
- Browser direct export is absent.
- Non-transferred OpenTelemetry concepts are explicitly forbidden or deferred.
- Golden telemetry corpus comparison is required for OTel dependency, SDK, exporter, semantic-convention, generated-constant, sampler, metric-view, and retry-behavior updates.
- All acceptance criteria in §15 are binary and pass.
- No `TODO`, `TBD`, placeholder digest, placeholder package version, placeholder generated constants source, shortened SHA, `main` ref, or unresolved adoption-decision row remains in the adopted source snapshot.
- `otel_source_snapshot` validates before conformance tests are accepted, and the snapshot is byte-stable for identical inputs.
- Static dependency checks prove API-only ordinary instrumentation and SDK/exporter-only bootstrap import boundaries.
- Browser bundle checks prove no browser SDK, exporter, vendor telemetry package, Collector client, session replay package, analytics initialization path, or telemetry configuration path.
- Committed normalized goldens and retained raw captures follow §14.1 paths.
- Every acceptance criterion in §15 is executable by a named harness target or explicitly mapped to a repo-control validation tool.
