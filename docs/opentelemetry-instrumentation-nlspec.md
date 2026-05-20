---
title: Cartulary OpenTelemetry Instrumentation NLSpec
status: draft/proposed
document_class: nlspec
created_at: 2026-05-19
---

## 1. Status, scope, and authority

Status: `draft/proposed`.

This NLSpec defines Cartulary's OpenTelemetry instrumentation subsystem. It is not adopted implementation-conformance authority until the Cartulary repository authority process adopts it. This revision preserves the authority boundary of the uploaded draft and hardens it against OpenTelemetry configuration, semantic-convention, privacy, and signal-shape drift.[^1]

**OTEL-REQ-001**
This NLSpec governs only telemetry generation, telemetry configuration, telemetry export, log correlation, signal naming, attribute governance, privacy boundaries, telemetry runtime behavior, and instrumentation verification.

**OTEL-REQ-002**
This NLSpec MUST NOT redefine product behavior owned by Cartulary Core 00 through Core 04. It MUST NOT redefine claim-bearing benchmark publication owned by Core 05. Runtime telemetry MAY support engineering diagnosis and operational SRE practice, but telemetry observations MUST NOT satisfy claim-bearing timed or fixture-sensitive publication unless the Core 05 benchmark-manifest and measurement-predicate requirements are also satisfied.[^2][^3]

**OTEL-REQ-003**
When this NLSpec conflicts with Core 00 through Core 04 before adoption, the conflict is a draft defect in this NLSpec. When a later adopted version of this NLSpec conflicts with older non-normative appendices or guides, the adopted NLSpec governs only the telemetry subsystem.

**OTEL-REQ-004**
The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are normative inside this NLSpec. **MUST** and **MUST NOT** define conformance requirements. **SHOULD** and **SHOULD NOT** define strong defaults whose exceptions must remain compatible with all MUST-level requirements. **MAY** defines optional behavior whose omission semantics are explicit.

## 2. Purpose

**OTEL-REQ-005**
Cartulary MUST provide first-class observability because long-term support requires operators to diagnose availability, latency, error, queueing, persistence, evidence access, and collaboration failures without inspecting incident content, exposing stable incident identifiers, or weakening the workbook hot path.

The instrumentation subsystem MUST make these operational questions answerable:

| Question | Required signal support |
| --- | --- |
| Is the application deployable accepting and completing HTTP requests? | HTTP traces and bounded HTTP metrics. |
| Are workbook queries, mutations, and projection updates healthy? | Workbook and projection spans plus duration, row-count, conflict, and result metrics. |
| Are WebSocket subscriptions, presence updates, and live row updates healthy? | WebSocket active gauges, event counters, bounded operation spans, and low-cardinality close or drop classification. |
| Are background jobs queued, running, canceled, failed, or completing? | Job enqueue and run spans, job active gauges, terminal duration metrics, and terminal-status attributes. |
| Are Postgres or object-storage dependencies degraded? | Dependency spans, dependency duration metrics, and low-cardinality error classification. |
| Are telemetry exporters failing or dropping data? | Telemetry self-metrics and bounded local diagnostics. |
| Can operators correlate local logs with traces without exposing secrets or incident content? | Trace correlation fields, LogRecord mapping, bounded body rules, and redaction-before-recording rules. |

## 3. Non-goals

**OTEL-REQ-006**
This NLSpec MUST NOT introduce any behavior in the following table:

| Non-goal | Boundary |
| --- | --- |
| A new product workflow | Telemetry MUST NOT add row-edit rituals, approval gates, or user-facing capture friction. |
| A new case-data source of truth | Telemetry MUST NOT become authoritative incident state, audit state, history state, projection state, workflow state, evidence state, or benchmark evidence. |
| A monitoring vendor dependency | The implementation MUST remain vendor-neutral and MUST NOT require Datadog, Grafana, Honeycomb, New Relic, Splunk, Elastic, Jaeger, Prometheus, or any other specific backend. |
| A required OpenTelemetry Collector | A Collector MAY be an external receiver, but it is not a Cartulary deployable and is not required for Cartulary telemetry conformance. |
| A public metrics dashboard contract | Dashboards, alerts, and runbooks MAY be derived later, but this NLSpec owns emitted telemetry, not dashboard layout. |
| Browser-to-third-party telemetry export | Browser code MUST NOT send telemetry directly to an external collector or vendor endpoint. |
| Raw incident-content logging | Incident-authored values, evidence bytes, raw queries, note text, timeline details, filenames, credential material, and object-store keys MUST NOT be exported as telemetry attributes or logs. |
| Environment-driven telemetry egress | OpenTelemetry SDK environment defaults MUST NOT override Cartulary's deployment configuration or enable export when Cartulary export is disabled. |
| Collector-side privacy enforcement | A Collector, backend, exporter, or vendor pipeline MUST NOT be required to make emitted telemetry privacy-conformant. |

## 4. External standard baseline

OpenTelemetry is the selected telemetry framework because it is the vendor-neutral project baseline for generating and exporting observability signals. OpenTelemetry clients separate API, SDK, semantic conventions, and contrib packages; ordinary instrumentation code must depend on API packages rather than SDK packages, while the application owner manages SDK installation and configuration.[^6][^7]

### 4.1 Baseline object

**OTEL-REQ-007**
The initial external standard baseline MUST be the closed object in this table:

| Field | Required value or rule |
| --- | --- |
| `otel_spec_version` | `1.56.0` until this NLSpec is revised. |
| `otel_spec_source` | `https://opentelemetry.io/docs/specs/otel/`. |
| `otel_spec_observed_at` | `2026-05-20`. |
| `semconv_version` | `1.41.0` until this NLSpec is revised. |
| `semconv_source` | `https://opentelemetry.io/docs/specs/semconv/`. |
| `semconv_model_source` | Semantic-convention YAML model files from the pinned semantic-conventions source, not prose-only extraction. |
| `semconv_generated_constants_version` | Exact generated-constant package or code-generation source version used by the implementation. This value MUST be pinned in repo-control files before implementation adoption. |
| `semconv_stability_policy` | The policy in §4.3 and the change classification in §4.4. |
| `migration_note_required` | Derived from §4.4 rather than a single unconditional boolean. |

**OTEL-REQ-008**
The baseline versions above MUST be treated as an adoption lock, not as a claim that later OpenTelemetry releases are incompatible. A later adopted revision MAY rebaseline to newer OpenTelemetry or semantic-convention versions only after applying §4.4 and updating all affected signal registries, configuration rules, and acceptance criteria.

### 4.2 OpenTelemetry component boundary

**OTEL-REQ-009**
Cartulary MUST use the component meanings in this table:

| Term | Required Cartulary meaning |
| --- | --- |
| `OpenTelemetry API` | The only OTel package family that ordinary Cartulary instrumentation code may call directly. |
| `OpenTelemetry SDK` | Installed, configured, and shut down only by the server-side telemetry bootstrap boundary. |
| `Instrumentation unit` | Cartulary code that records telemetry for one internal module or platform concern through the OTel API. |
| `Instrumentation scope` | The OTel `(name, version, schema_url, attributes)` identity used when obtaining tracers, meters, or loggers for a Cartulary instrumentation unit. |
| `Exporter` | Server-side component that sends telemetry to the configured OTLP endpoint. Exporters are never configured by browser code. |
| `Processor` | Server-side component that batches, bounds, drops, flushes, and forwards telemetry to exporters. |
| `Collector` | Optional external receiver. It is not a Cartulary deployable and is not required for Cartulary telemetry conformance. |
| `Semantic conventions` | OTel standard names and meanings for common telemetry concepts. They are adopted by stability policy and registry generation, not copied ad hoc. |

**OTEL-REQ-010**
Ordinary instrumentation units MUST obtain tracers, meters, and loggers only through API-facing provider accessors supplied by the telemetry bootstrap boundary. They MUST NOT import, construct, or configure SDK providers, exporters, processors, samplers, propagators, metric readers, log processors, declarative configuration, SDK autoconfiguration, or plugin-provider packages.

**OTEL-REQ-011**
The telemetry bootstrap boundary MAY import SDK packages only for provider setup, configuration validation, processor construction, exporter construction, metric-reader construction, sampler construction, shutdown, and bounded self-diagnostics. It MUST NOT make SDK construction or exporter configuration callable from ordinary instrumentation units.

### 4.3 Semantic-convention stability policy

**OTEL-REQ-012**
Cartulary MUST apply this semantic-convention adoption matrix:

| OTel convention status | Cartulary default |
| --- | --- |
| Stable and applicable | Emit by default when it does not violate Cartulary privacy, cardinality, configuration, or deployment-boundary rules. |
| Stable but privacy-conflicting | Do not emit the conflicting attribute. Record the omission in the signal-specific allowlist or non-adoption table that would otherwise own it. |
| Development or experimental | Do not emit by default. Adoption requires an explicit NLSpec revision or an explicit opt-in configuration key defined by this NLSpec. |
| Deprecated | Do not emit unless a migration-compatibility profile explicitly requires it. |
| Migration-period duplicated conventions | Do not duplicate by default. Duplication requires a bounded compatibility rule and an acceptance criterion proving both old and new forms are intentional. |
| Unknown or unpinned | Do not emit. |

**OTEL-REQ-013**
Every emitted standard attribute and standard metric name MUST be generated or imported from the pinned semantic-convention model source, or MUST be explicitly listed as a standard attribute allowlist exception in the owning signal registry. Every emitted Cartulary custom attribute MUST be listed in §8.4.

### 4.4 Telemetry shape change classification

**OTEL-REQ-014**
Every dependency update, semantic-convention update, SDK update, exporter update, instrumentation change, or signal-registry change MUST be classified by this table before it is accepted as conformant:

| Change class | Definition | NLSpec revision required | Migration note required | Acceptance impact |
| --- | --- | ---: | ---: | --- |
| `registry_equivalent` | Same emitted spans, span names, span kinds, metrics, metric identities, resource attributes, log mappings, standard attributes, custom attributes, and forbidden-value exclusions for the same conformance corpus. | No, if dependency-only. | No. | Existing acceptance criteria must pass unchanged. |
| `additive_non_breaking` | Adds a new emitted telemetry element or optional low-cardinality attribute without removing, renaming, retyping, or changing existing emitted telemetry. | Yes. | Yes. | Add criteria proving the addition is intentional and privacy-safe. |
| `privacy_tightening` | Removes or suppresses telemetry to reduce disclosure, cardinality, unsafe correlation, or unbounded retention risk. | Yes. | Yes. | Add criteria proving the removed element is absent and required operational questions retain required coverage. |
| `breaking_shape_change` | Removes, renames, retypes, changes requiredness, changes default emission, changes span parent/link topology, changes metric temporality or aggregation, changes resource identity, or changes log mapping. | Yes. | Yes. | Add or update criteria for old-shape absence and new-shape presence. |

**OTEL-REQ-015**
Dependency-only updates MAY occur without an NLSpec revision only when emitted telemetry remains `registry_equivalent` for the conformance corpus. A semantic-convention update, SDK update, generated-constant update, or `OTEL_SEMCONV_STABILITY_OPT_IN` setting MUST NOT become active if it causes any `additive_non_breaking`, `privacy_tightening`, or `breaking_shape_change` effect without this NLSpec being revised.

## 5. Instrumentation ownership

The application deployable owns the browser-facing UI host, API surface, WebSocket hub, and background-job runners because Cartulary's base deployment is one web application deployable plus Postgres and S3-compatible object storage.[^4]

### 5.1 Subsystem boundary

**OTEL-REQ-016**
The instrumentation subsystem MUST be a logical internal boundary inside the modular monolith. It MUST NOT require a separate application deployable, sidecar, microservice, Collector, vendor backend, Prometheus server, or browser telemetry service.

**OTEL-REQ-017**
Instrumentation ownership MUST follow this table:

| Runtime area | Instrumentation owner | Required coverage | Forbidden ownership leak |
| --- | --- | --- | --- |
| HTTP API | Application server instrumentation | Server spans, request metrics, status/error classification. | Route handlers MUST NOT configure exporters or SDK providers. |
| WebSocket subscription | Collaboration instrumentation | Connect, authorize, subscribe, close, event-send, replay, and overflow metrics. | WebSocket payload content MUST NOT become telemetry attributes. |
| Background jobs | Jobs instrumentation | Enqueue, start, terminal state, cancellation, duration, active count, and error metrics. | Job IDs MUST NOT be emitted. |
| Postgres access | Platform/Postgres instrumentation | Standard database client spans and duration metrics without SQL text or bind values. | Workbook modules MUST NOT directly emit raw database query text. |
| Object storage | Platform/object-store instrumentation | Cartulary object-store dependency spans and byte/duration metrics without bucket names, keys, hashes, filenames, or handles. | Object-store implementation details MUST NOT leak into evidence telemetry. |
| Workbook query and mutation | Workbook/projection instrumentation | Query, create, patch, conflict, projection-maintenance, and refresh spans. | Projection table names and visible row positions MUST NOT become telemetry identity. |
| Browser UI | Browser controller instrumentation | Local performance marks MAY exist. | Browser direct OTLP, vendor-native, or third-party telemetry export is forbidden. |
| Telemetry bootstrap | Server-side telemetry boundary | SDK provider setup, processors, exporters, readers, shutdown, and self-diagnostics. | Ordinary instrumentation units MUST NOT configure SDK, exporter, processor, reader, sampler, propagator, or Collector behavior. |

### 5.2 Instrumentation scopes

**OTEL-REQ-018**
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

**OTEL-REQ-019**
No instrumentation unit may create an unregistered instrumentation scope. A future revision that adds scope attributes MUST define a closed attribute table with names, types, allowed values, cardinality bound, default, and forbidden-value tests.

## 6. Configuration contract

Telemetry configuration lives in the Cartulary deployment configuration surface. Core 04 owns the operator-facing deployment configuration artifact, discovery precedence, binding keys, and fail-closed startup validation; this NLSpec adds telemetry keys under that same surface rather than defining a second configuration model.[^5]

### 6.1 Configuration keys

**OTEL-REQ-020**
The effective telemetry configuration MUST be the closed key set in this table. Unknown `telemetry.*` keys are invalid unless a later revision defines them.

| Key | Type | Default | Bounds or values | Omitted behavior | Explicit `null` behavior | Required behavior |
| --- | --- | --- | --- | --- | --- | --- |
| `telemetry.enabled` | boolean | `true` | `true`, `false` | Use default. | Invalid. | When `false`, no OpenTelemetry providers, exporters, log bridges, or instrumentation hooks are installed except no-op placeholders needed for code safety. |
| `telemetry.otel_env_passthrough` | boolean | `false` | `true`, `false` | Use default. | Invalid. | When `false`, OTel SDK environment variables, declarative config, and SDK autoconfig MUST NOT enable exporters, propagators, samplers, processors, endpoints, headers, config files, views, metric readers, or plugins. |
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
| `telemetry.traces.sample_ratio` | decimal | `0.10` | `0.0..1.0` inclusive | Use default. | Invalid. | Uses parent-based trace-ID-ratio sampling over server-owned trace IDs. |
| `telemetry.traces.accept_remote_context` | boolean | `false` | exactly `false` in this revision | Use default. | Invalid. | Remote trace context is not trusted in this revision. |
| `telemetry.metrics.enabled` | boolean | `true` | `true`, `false` | Use default. | Invalid. | Has effect only when `telemetry.enabled=true`. |
| `telemetry.metrics.exemplars.enabled` | boolean | `false` | exactly `false` in this revision | Use default. | Invalid. | Configure the SDK exemplar filter to `AlwaysOff` or equivalent. Exemplar emission is non-conformant in this revision. |
| `telemetry.logs.bridge_enabled` | boolean | `false` | `true`, `false` | Use default. | Invalid. | When `false`, local structured logs MAY include trace correlation fields but MUST NOT be exported as OpenTelemetry logs. |
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
| `telemetry.resource.service_instance_id` | string or null | Generated UUID v4 per process start | Non-empty opaque string, maximum `128` Unicode scalar values | Generate default. | Generate default. | Maps to `service.instance.id`. Must satisfy §7. |
| `telemetry.resource.deployment_environment_name` | string or null | `null` | `development`, `test`, `staging`, `production`, or custom non-empty token of maximum `128` ASCII letters, digits, `.`, `_`, or `-` | Omit attribute. | Omit attribute. | Maps to `deployment.environment.name` when present. |
| `telemetry.attribute.incident_correlation` | enum | `none` | `none`, `hmac_64bit` | Use default. | Invalid. | `none` forbids incident correlation attributes. `hmac_64bit` is a narrowed opt-in under §8.5. |
| `telemetry.attribute.hmac_secret_ref` | string or null | `null` | Server-side secret reference | Required only when incident correlation is `hmac_64bit`. | Valid only when incident correlation is `none`; otherwise invalid. | Secret value MUST NOT be exported. |

### 6.2 Configuration precedence

**OTEL-REQ-021**
Configuration precedence MUST be exactly this table:

| Precedence | Source | Required behavior |
| --- | --- | --- |
| 1 | Cartulary deployment configuration | Authoritative for all telemetry behavior. |
| 2 | Cartulary server-side environment bindings | MAY populate Cartulary deployment configuration keys only. Empty values are treated as omitted. |
| 3 | OTel SDK environment variables | Ignored unless `telemetry.otel_env_passthrough=true`; even then constrained by §6.3. |
| 4 | OTel declarative configuration and plugin providers | Not authoritative in this revision; constrained by §6.3. |
| 5 | OTel SDK defaults | MAY apply only inside the effective Cartulary configuration envelope. MUST NOT enable export when Cartulary export is `none`. |
| 6 | Browser state or browser environment | Never a telemetry exporter, processor, sampler, propagator, metric-reader, or header configuration source. |

### 6.3 OpenTelemetry external configuration containment

OpenTelemetry supports programmatic, environment-variable, declarative, and other configuration mechanisms, and declarative configuration includes file-based SDK component configuration and custom plugin components.[^8][^9]

**OTEL-REQ-022**
External OpenTelemetry configuration inputs MUST be contained by this matrix:

| External input | Current-profile treatment | Allowed effect |
| --- | --- | --- |
| OTel SDK environment variables | Ignored unless `telemetry.otel_env_passthrough=true`. | May fill only otherwise omitted OTel-equivalent settings explicitly mapped by this NLSpec. |
| `OTEL_CONFIG_FILE` | Ignored when `telemetry.otel_env_passthrough=false`; unsupported as an authority even when passthrough is true unless a future revision maps it. | No exporter, processor, reader, sampler, propagator, header, exemplar, view, or plugin creation. |
| OTel declarative configuration file | Not an authoritative Cartulary configuration source. | None in current profile. |
| OTel instrumentation ConfigProvider state | Not an authoritative Cartulary configuration source. | None unless a future revision defines a closed mapping. |
| OTel plugin component provider | Forbidden as a runtime configuration authority. | None. |
| `OTEL_SEMCONV_STABILITY_OPT_IN` | Ignored unless a future revision defines a semantic-convention migration profile. | No emitted-shape change. |
| Per-signal OTLP endpoint env vars | Rejected or ignored according to effective config validation; never authoritative. | None in current profile. |
| Per-signal OTLP protocol or header env vars | Rejected or ignored according to effective config validation; never authoritative. | None in current profile. |
| Browser state | Never an exporter or processor configuration source. | None. |

**OTEL-REQ-023**
When `telemetry.otel_env_passthrough=false`, the implementation MUST ignore the following environment variables for behavior selection. It MAY retain their presence only in redacted startup diagnostics that do not expose values.

| Environment variable family | Members |
| --- | --- |
| Exporter selection | `OTEL_TRACES_EXPORTER`, `OTEL_METRICS_EXPORTER`, `OTEL_LOGS_EXPORTER` |
| Global OTLP | `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_HEADERS`, `OTEL_EXPORTER_OTLP_PROTOCOL` |
| Per-signal OTLP endpoints | `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`, `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`, `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` |
| Per-signal OTLP protocols | `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL`, `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL`, `OTEL_EXPORTER_OTLP_LOGS_PROTOCOL` |
| Per-signal OTLP headers | `OTEL_EXPORTER_OTLP_TRACES_HEADERS`, `OTEL_EXPORTER_OTLP_METRICS_HEADERS`, `OTEL_EXPORTER_OTLP_LOGS_HEADERS` |
| Context and sampling | `OTEL_PROPAGATORS`, `OTEL_TRACES_SAMPLER`, `OTEL_TRACES_SAMPLER_ARG` |
| Processors | `OTEL_BSP_*`, `OTEL_BLRP_*`, `OTEL_METRIC_EXPORT_INTERVAL` |
| Declarative config | `OTEL_CONFIG_FILE` |
| Semantic-convention migration | `OTEL_SEMCONV_STABILITY_OPT_IN` |

**OTEL-REQ-024**
When `telemetry.otel_env_passthrough=true`, OTel SDK environment variables MAY fill only otherwise omitted OTel-equivalent runtime settings. They MUST NOT override an explicit Cartulary configuration value. They MUST NOT permit unsupported exporters, unsupported protocols, remote-context acceptance, Baggage correlation, Prometheus scrape exposure, Zipkin export, Jaeger native export, vendor-native export, SQL commenter propagation, per-signal endpoints, per-signal headers, per-signal protocol divergence, OTel declarative configuration, SDK plugin providers, metric exemplars, unregistered metric views, or browser export.

### 6.4 Configuration validation

**OTEL-REQ-025**
Invalid telemetry configuration MUST fail deployment-configuration validation before readiness with `error.code='invalid_deployment_config'` and `reason_code='invalid_telemetry_config'` or an equivalent deployment-config reason-code family if the repository owner has already defined a narrower code. Exporter endpoint network unavailability MUST NOT fail startup when the endpoint is syntactically valid.

**OTEL-REQ-026**
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
| Unsupported protocol | Any exporter protocol value other than `grpc` or `http/protobuf`. |
| Per-signal endpoints | Any per-signal endpoint key appears in effective behavior. |
| Per-signal protocol or header divergence | Any per-signal protocol or header key appears in effective behavior. |
| Exemplar enablement | `telemetry.metrics.exemplars.enabled` is any value other than `false`. |
| Log body bound | `telemetry.logs.body_max_chars` is outside `0..8192`. |
| External OTel config authority | Any OTel declarative config, SDK autoconfig, plugin provider, or ConfigProvider state attempts to create or alter exporters, processors, propagators, samplers, metric readers, log processors, header capture, metric views, exemplars, or SDK plugin components outside declared `telemetry.*` keys. |
| Semantic-convention environment opt-in | `OTEL_SEMCONV_STABILITY_OPT_IN` would alter emitted telemetry shape in the current profile. |

## 7. Resource attributes

OpenTelemetry service semantic conventions define stable service identity attributes and recommend opaque service-instance identity because underlying host, pod, or machine identity can be confidential. `service.criticality` is Development-status and is not adopted in this profile.[^10]

**OTEL-REQ-027**
The instrumentation subsystem MUST attach the following resource attributes to all emitted traces, metrics, and exported logs:

| Attribute | Requiredness | Value source | Export rule | Privacy rule |
| --- | --- | --- | --- | --- |
| `service.name` | Required | `telemetry.resource.service_name` | Always export. | MUST NOT resolve to SDK-generated `unknown_service:*`. |
| `service.namespace` | Required | `telemetry.resource.service_namespace` | Always export. | Default is `cartulary`. |
| `service.version` | Required | `telemetry.resource.service_version` | Always export. | Omission is non-conformant after default resolution. |
| `service.instance.id` | Required | configured value or generated UUID v4 | Always export. | MUST be opaque. |
| `deployment.environment.name` | Optional | `telemetry.resource.deployment_environment_name` | Export only when configured. | Descriptive only. MUST NOT drive tenancy, service identity, authorization, or incident identity. |
| `telemetry.sdk.language` | SDK-required | SDK-provided | Do not override. | SDK value only. |
| `telemetry.sdk.name` | SDK-required | SDK-provided | Do not override. | SDK value only. |
| `telemetry.sdk.version` | SDK-required | SDK-provided | Do not override. | SDK value only. |
| `cartulary.deployment.profile` | Required | Deployment profile | Always export. | Low-cardinality custom attribute. |
| `cartulary.profile.claims` | Optional | Claimed profile IDs known at startup | Export as sorted comma-delimited profile tokens only when known at startup. | MUST NOT include incident, user, or customer identifiers. |

**OTEL-REQ-028**
`service.instance.id` MUST NOT be a hostname, pod name, container ID, IP address, MAC address, user identifier, incident identifier, customer name, filesystem path, object-store key, deployment root, or secret reference. The default generated value MUST be a canonical lowercase UUID v4 generated exactly once during telemetry bootstrap.

**OTEL-REQ-029**
A generated default `service.instance.id` MUST remain stable for the lifetime of the application process. A new application process start with no configured `telemetry.resource.service_instance_id` MUST generate a different value with high probability. Conformance tests MAY assert difference across two controlled process starts.

**OTEL-REQ-030**
`service.criticality` MUST NOT be emitted in the current profile because it is Development-status and is not required to answer any operational question in §2.

## 8. Attribute governance

### 8.1 Attribute classes

**OTEL-REQ-031**
Telemetry attributes MUST follow this governance table:

| Class | Required behavior |
| --- | --- |
| Standard OTel stable attributes | MAY be emitted only when allowed by the active signal-specific registry. |
| Standard OTel development attributes | MUST NOT be emitted unless a named opt-in key exists in this NLSpec. No such opt-in key exists in this revision. |
| Standard OTel privacy-conflicting attributes | MUST NOT be emitted even if OTel defines them. |
| Cartulary custom attributes | MUST use `cartulary.` and MUST appear in §8.4. |
| Reserved namespace attributes | MUST NOT be created by Cartulary as custom attributes. |
| Unknown custom attributes | MUST NOT be emitted. |
| High-cardinality values | MUST NOT be emitted unless the exact attribute and bound are listed. |

### 8.2 Reserved namespaces

**OTEL-REQ-032**
Cartulary custom attributes MUST NOT use any prefix in this table:

| Prefix | Rule |
| --- | --- |
| `otel.` | Reserved for OpenTelemetry. |
| `service.` | Reserved for OTel resource attributes. |
| `telemetry.` | Reserved for OTel SDK resource attributes. |
| `http.` | Reserved for OTel HTTP semantic conventions. |
| `url.` | Reserved for OTel URL semantic conventions. |
| `db.` | Reserved for OTel DB semantic conventions. |
| `aws.` | Reserved for OTel or AWS semantic conventions. |
| `exception.` | Reserved for OTel exception semantic conventions. |
| `enduser.` | Not used because user identity is forbidden in telemetry. |
| `user.` | Not used because user identity is forbidden in telemetry. |
| `session.` | Not used because session identity is forbidden in telemetry. |

### 8.3 Forbidden telemetry values and attributes

**OTEL-REQ-033**
The forbidden value family is closed by this table. The implementation MUST prevent these values from reaching spans, metrics, logs, events, exception fields, resource attributes, exporter headers, exporter artifacts, self-diagnostics, retained test artifacts, and any app-mediated future telemetry route.

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

**OTEL-REQ-034**
When a value belongs to a forbidden family, the implementation MUST do one of the following before recording telemetry:

| Treatment | Required behavior |
| --- | --- |
| Omit | Do not set the attribute, body text, event field, or diagnostic field. |
| Replace with closed class | Emit only a closed low-cardinality class such as `validation_error`, `permission_denied`, `timeout`, `queue_full`, or `redaction_rejected`. |
| Drop item | Drop the telemetry item when safe omission cannot be proven. |

### 8.3.1 Redaction-before-recording invariant

**OTEL-REQ-035**
Forbidden-value detection, redaction, omission, replacement, or rejection MUST occur before any telemetry item reaches an OTel span attribute setter, metric measurement, log bridge mapper, SDK processor, exporter queue, retained telemetry artifact, or self-diagnostic sink.

**OTEL-REQ-036**
A Collector, backend, exporter, vendor pipeline, or external scrubber MUST NOT be relied on to satisfy Cartulary privacy conformance. Cartulary telemetry must already be conformant before SDK processor or exporter handoff.

**OTEL-REQ-037**
When redaction cannot prove an item safe, the implementation MUST drop the telemetry item and increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='redaction_rejected'` when doing so does not recurse.

**OTEL-REQ-038**
The forbidden-value test corpus MUST include at least one representative value from every forbidden family in OTEL-REQ-033 and MUST exercise traces, metrics, logs, self-diagnostics, and retained telemetry artifacts.

### 8.4 Cartulary custom attribute registry

**OTEL-REQ-039**
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

**OTEL-REQ-040**
Unknown `cartulary.*` attributes MUST NOT be emitted. Adding a new `cartulary.*` attribute is an `additive_non_breaking` change unless it changes requiredness, type, or emitted shape, in which case §4.4 decides the change class.

### 8.5 Incident correlation opt-in

**OTEL-REQ-041**
The default `telemetry.attribute.incident_correlation='none'` MUST omit incident correlation attributes. In default configuration, no incident identifier, incident key, incident title, customer name, or incident-derived hash may appear in telemetry.

**OTEL-REQ-042**
When `telemetry.attribute.incident_correlation='hmac_64bit'`, the implementation MAY emit only `cartulary.incident.hash64`. The value MUST be the first 64 bits of `HMAC-SHA-256(secret, canonical_incident_id_bytes)`, encoded as exactly 16 lowercase hexadecimal characters. The secret value MUST be resolved server-side from `telemetry.attribute.hmac_secret_ref`, MUST NOT be exported, and MUST NOT be available to browser code.

**OTEL-REQ-043**
`cartulary.incident.hash64` is an operational grouping key, not a stable public incident identifier. It MUST NOT appear in product API responses, workbook rows, WebSocket payloads, evidence handles, export snapshots, benchmark manifests, or user-facing UI labels solely because telemetry correlation is enabled.

## 9. Tracing contract

OpenTelemetry spans carry operation name, timestamps, attributes, events, parent span identity, links to causally related spans, and SpanContext. OpenTelemetry links are the appropriate model for batch, async, or trusted-boundary relationships that do not have a single synchronous parent.[^14]

### 9.1 General span rules

**OTEL-REQ-044**
Span names MUST use route templates, module operations, or stable operation names. Span names MUST NOT include path IDs, incident IDs, record IDs, user-supplied strings, search text, filenames, object keys, SQL text, visible row values, saved-view IDs, job IDs, user IDs, or handle tokens.

**OTEL-REQ-045**
Span status and `error.type` MUST follow the signal-specific error rules. `cartulary.error_code` carries Cartulary public error-code tokens. The implementation MUST NOT overload `error.type` with Cartulary public error codes unless the value is also the low-cardinality OTel error class for that span.

**OTEL-REQ-046**
Exception recording MUST NOT emit `exception.message` or `exception.stacktrace` in the base telemetry profile. `exception.type` MAY be emitted only when it is a low-cardinality class name that does not include incident, user, record, object, path, SQL, or secret material.

### 9.2 Required span families

**OTEL-REQ-047**
The implementation MUST provide the span families in this table when the corresponding operation executes and tracing is enabled:

| Span family | Span name | SpanKind | Instrumentation scope | Parent rule | Link rule | Required standard attributes | Required Cartulary attributes | Forbidden attributes | Start predicate | End predicate | Error rule |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| HTTP server | `{http.request.method} {http.route}` when route is known; otherwise `{http.request.method}` | `SERVER` | `cartulary.httpapi` | Root; inbound remote context is ignored in this revision. | Forbidden. | `http.request.method`; `http.route` when known; `http.response.status_code` when sent; optional `url.scheme`. | `cartulary.route_family`, `cartulary.result`, optional `cartulary.error_code`. | `url.full`, `url.path`, `url.query`, concrete path-derived name, request headers, response headers, client IP attributes, server address attributes. | Request accepted by router. | Response body completion or route failure. | 4xx server responses are not errors solely because they are 4xx; 5xx sets error status unless a narrower owner rule says otherwise. |
| Workbook query | `cartulary.workbook.query` | `INTERNAL` | `cartulary.workbook` | Current HTTP server span or job span. | Forbidden. | None. | `cartulary.view_schema_id`, `cartulary.result`, optional `cartulary.incident.hash64`. | Raw filters, search text, saved-view ID, incident ID, record ID. | Query validation begins. | Response rows serialized or query rejected. | Validation rejection uses `cartulary.result='rejected'`; runtime failure sets error status and low-cardinality `error.type`. |
| Record mutation | `cartulary.record.mutate` | `INTERNAL` | `cartulary.workbook` | Current HTTP server span or job span. | Forbidden. | None. | `cartulary.view_schema_id`, `cartulary.record_type`, `cartulary.operation`, `cartulary.result`, optional `cartulary.error_code`, optional `cartulary.incident.hash64`. | Record ID, user ID, client transaction ID, submitted value, persisted value, conflict token. | Mutation validation begins. | Mutation commits, conflicts, rejects, fails, or is canceled. | Same-field conflict uses `cartulary.result='conflict'` and MUST NOT set OTel error status solely because conflict occurred. |
| Projection maintenance | `cartulary.projection.maintenance` | `INTERNAL` | `cartulary.workbook` | Current mutation span, job span, or root for maintenance rebuild. | Optional when triggered by batch or replay operation. | None. | `cartulary.record_type`, `cartulary.operation`, `cartulary.result`. | Projection table names, record IDs, incident IDs. | Projection unit begins. | Projection unit completes or fails. | Runtime failure sets error status. |
| WebSocket subscribe | `cartulary.websocket.subscribe` | `INTERNAL` | `cartulary.collaboration` | Current HTTP upgrade/request span when available; otherwise root. | Forbidden. | None. | `cartulary.result`, optional `cartulary.error_code`. | Connection ID, user ID, incident ID, record ID, field key. | Subscription authorization begins. | Subscription accepted, rejected, or closed during setup. | Auth or membership rejection uses `cartulary.result='rejected'`. |
| WebSocket send | `cartulary.websocket.send` | `INTERNAL` | `cartulary.collaboration` | Current operation span when send is synchronous; otherwise current job span or root. | Optional when caused by async mutation fan-out. | None. | `cartulary.websocket.event_type`, `cartulary.result`. | Payload content, connection ID, user ID, incident ID, record ID, changed field values. | Event serialization begins. | Event sent, dropped, or failed. | Drop due to disconnected client is not an error unless implementation error caused the drop. |
| Job enqueue | `cartulary.job.enqueue` | `PRODUCER` | `cartulary.jobs` | Current HTTP, mutation, or system-process span. | Forbidden. | None. | `cartulary.job_kind`, `cartulary.result`. | Job ID, incident ID, user ID, request payload. | Enqueue validation begins. | Job accepted, rejected, or fails to enqueue. | Enqueue rejection uses `cartulary.result='rejected'`; runtime failure sets error status. |
| Job run | `cartulary.job.run` | `CONSUMER` | `cartulary.jobs` | Root. | Required when job was enqueued from a traced request or traced mutation; optional for scheduler-only maintenance. | None. | `cartulary.job_kind`, `cartulary.job_terminal_status` when terminal, `cartulary.result`, optional `cartulary.incident.hash64`. | Job ID, user ID, record ID, incident ID unless hashed opt-in applies. | Job leaves queued state. | Job reaches terminal state. | Terminal failed job sets error status only when failure is an implementation or dependency error; ordinary cancellation uses `cartulary.result='canceled'`. |
| Object-store dependency | `cartulary.objectstore.operation` | `CLIENT` | `cartulary.objectstore` | Current HTTP, mutation, job, or evidence operation span. | Forbidden. | None. | `cartulary.module='objectstore'`, `cartulary.operation`, `cartulary.result`. | `aws.s3.*`, bucket, key, upload ID, copy source, part number, filename, hash, evidence handle, storage ref. | Object operation begins. | Object operation completes or fails. | Dependency failure sets error status and low-cardinality `error.type`. |
| Postgres dependency | `postgresql {db.operation.name}` when operation name is known; otherwise `postgresql` | `CLIENT` | `cartulary.postgres` | Current HTTP, mutation, job, or projection span. | Forbidden. | `db.system.name='postgresql'`; `db.operation.name` when available; optional `db.query.summary` only from fixed low-cardinality labels. | `cartulary.module='postgres'`, `cartulary.result`. | `db.query.text`, `db.query.parameter.*`, SQL fragment, bind value, table name, projection name, saved-view query JSON, search text. | Database operation begins. | Database operation completes or fails. | Dependency failure sets error status and low-cardinality database error class. |

**OTEL-REQ-048**
The implementation MUST NOT emit `cartulary.postgres.operation` as the only database dependency span family. Postgres dependency telemetry MUST use the database-client contract in §9.5. Cartulary internal workbook spans MAY remain parents of those dependency spans.

### 9.3 HTTP server span details

OpenTelemetry HTTP semantic conventions are stable, require low-cardinality route templates for `http.route`, and specify that server 4xx responses do not by themselves set span error status while 5xx responses generally do.[^11]

**OTEL-REQ-049**
HTTP server spans MUST use the Cartulary route template as `http.route` when a matched route exists. They MUST NOT use concrete URL paths, raw route parameters, query strings, incident IDs, record IDs, saved-view IDs, object IDs, handle tokens, or search text in span names.

**OTEL-REQ-050**
An exported HTTP server span MUST never be exported with a concrete path-derived name. If the route template is not known at span creation, the implementation MUST either create the span with `{http.request.method}` only and set the final name to `{http.request.method} {http.route}` before export when the route becomes known, or delay span creation until route selection is known. The implementation MUST NOT emit an intermediate span event, diagnostic, metric attribute, or log field containing the concrete path as a fallback.

**OTEL-REQ-051**
Cartulary intentionally omits `url.path`, `url.query`, and `url.full` from HTTP server telemetry in this revision even when a generic OTel HTTP instrumentation would normally emit them. `http.route` is the only allowed path-like HTTP attribute. If no low-cardinality route template is available, `http.route` MUST be omitted rather than replaced by a URI path.

**OTEL-REQ-052**
HTTP server 4xx responses MUST NOT set span error status solely because the response is 4xx. HTTP server 5xx responses MUST set span error status unless a narrower owner rule proves the status code is not an implementation or dependency error. Intentional caller cancellation MUST NOT be classified as an error solely because the request ended early.

### 9.4 HTTP standard attribute allowlist

**OTEL-REQ-053**
HTTP server spans MUST use this standard-attribute allowlist:

| Attribute | Current-profile behavior |
| --- | --- |
| `http.request.method` | Required. Unknown methods map to `_OTHER` according to OTel rules. |
| `http.route` | Required when a low-cardinality matched route template is available. Omitted when unavailable. |
| `http.response.status_code` | Required when a response status was sent. |
| `url.scheme` | Optional if known without exposing untrusted headers. |
| `url.path` | Forbidden. |
| `url.query` | Forbidden. |
| `url.full` | Forbidden. |
| `http.request.header.*` | Forbidden. |
| `http.response.header.*` | Forbidden. |
| `client.address`, `network.peer.address`, `network.peer.port` | Forbidden. |
| `server.address`, `server.port` | Forbidden in HTTP server spans in this revision. |
| `error.type` | Allowed only as a low-cardinality OTel error class under §9.1 and §9.3. |

### 9.5 Database span details

OpenTelemetry database semantic conventions are stable and define database spans, `db.query.text`, `db.query.summary`, `db.query.parameter.*`, and the well-known `db.system.name='postgresql'` value. Cartulary adopts only a privacy-bounded subset.[^12]

**OTEL-REQ-054**
Postgres dependency telemetry MUST NOT emit raw SQL, sanitized SQL, parameterized SQL text, bind values, saved-view query JSON, filter payloads, source headers, table names, projection table names, index names, schema names, or user-authored search text. `db.query.summary` MAY be emitted only when generated from a fixed low-cardinality operation label, not from SQL text.

**OTEL-REQ-055**
SQL commenter, database trace-context comments, or equivalent database-context propagation MUST be disabled in the current revision and MUST NOT be enabled by OTel environment variables.

**OTEL-REQ-056**
Postgres spans MUST use this standard-attribute allowlist:

| Attribute | Required current-profile behavior |
| --- | --- |
| `db.system.name` | Required with exact value `postgresql`. |
| `db.operation.name` | Allowed only from a closed low-cardinality operation vocabulary defined by the signal registry. |
| `db.query.summary` | Optional; allowed only from fixed low-cardinality labels that are not derived from SQL text, table names, projection names, filter values, or saved-view query JSON. |
| `db.query.text` | Forbidden. |
| `db.query.parameter.*` | Forbidden. |
| `db.namespace` | Omitted in current profile unless a future revision privacy-reviews it. |
| `db.collection.name` | Omitted in current profile. |
| `server.address` | Omitted in current profile. |
| `server.port` | Omitted in current profile. |
| `network.peer.address`, `network.peer.port` | Omitted in current profile. |
| Any table, projection, index, schema, or query-family name that identifies Cartulary storage realization | Forbidden. |

### 9.6 Object-store non-adoption rule

OpenTelemetry S3 semantic conventions are Development-status and include object-identifying attributes such as bucket, key, copy source, part number, and upload ID. Cartulary does not adopt those conventions in the base telemetry profile.[^13]

**OTEL-REQ-057**
Cartulary object-store telemetry MUST be a Cartulary dependency span and metric family over the object-storage abstraction. It MUST NOT be an AWS SDK span and MUST NOT be an S3 semantic-convention span, even when the configured object store is S3-compatible.

**OTEL-REQ-058**
Object-store operation telemetry MAY expose only the following values:

| Value family | Allowed members |
| --- | --- |
| Operation identity | `cartulary.module='objectstore'`, `cartulary.operation`. |
| Result identity | `cartulary.result`, optional low-cardinality `cartulary.error_class`. |
| Measurements | Duration metric, byte-count metric when payload size is known. |

**OTEL-REQ-059**
Object-store operation telemetry MUST NOT emit `aws.s3.*`, bucket, key, upload ID, copy source, part number, filename, blob hash, evidence handle, preview handle, download handle, object-blob ID, evidence-record ID, storage ref, evidence title, object-store endpoint, or object-store credential material.

### 9.7 Trace causality and links

**OTEL-REQ-060**
Trace topology MUST follow this table:

| Scenario | Required trace topology |
| --- | --- |
| HTTP request starts and completes synchronous workbook work | Child spans under the HTTP server span. |
| HTTP request enqueues a background job | `cartulary.job.enqueue` is a child of the request or mutation span; `cartulary.job.run` is a root `CONSUMER` span linked to the enqueue span context when a linkable context exists. |
| Background job starts from scheduler maintenance without a causal request | Job run span is a root span with only allowed system-process attribution attributes. |
| WebSocket fan-out caused by a mutation | Do not create false synchronous parentage from the original HTTP mutation to every pushed event. Use links or compact send spans only when causality matters. |
| Batch or replay operation processes multiple causal inputs | Use zero or more span links rather than a false single parent. |

**OTEL-REQ-061**
When an HTTP request or mutation enqueues a job and the job later executes asynchronously, `cartulary.job.run` MUST be a root `CONSUMER` span. It MUST include a link to the enqueue span context when the enqueue span was sampled or otherwise linkable. The link MUST be added at job-run span creation. The job-run span MUST NOT be parented under the original HTTP span after the request has returned.

**OTEL-REQ-062**
Batch or replay operations with multiple causal inputs MUST use zero or more links and MUST NOT invent a single parent solely for implementation convenience.

### 9.8 Inbound trace context and Baggage

OpenTelemetry Baggage is intended to propagate observability key-value pairs and can be consumed as attributes or context for metrics, logs, and traces. Cartulary does not transfer Baggage into the base profile.[^14]

**OTEL-REQ-063**
Inbound remote trace context is not trusted in this revision. Browser requests and API requests MUST start server-owned root traces. No configuration, environment variable, request header, proxy header, or browser state may change that behavior in this revision.

**OTEL-REQ-064**
Incoming `traceparent`, `tracestate`, and `baggage` headers MUST NOT become Cartulary telemetry identity, incident correlation, sampling input, export-routing input, span attributes, metric attributes, log attributes, log bodies, self-diagnostics, or retained telemetry artifacts.

**OTEL-REQ-065**
Incoming `traceparent` and `tracestate` MAY be observed only enough to ignore them safely. They MUST NOT establish parentage in the current profile. Incoming `baggage` MUST be discarded before any instrumentation scope can consume it.

## 10. Metrics contract

OpenTelemetry metric instruments are identified by name, kind, description, and unit. Views allow the SDK to filter attributes and configure aggregation. OpenTelemetry metrics otherwise place minimal constraints on data and generally avoid validation or sanitization; Cartulary intentionally applies stricter pre-export privacy and cardinality controls.[^15]

### 10.1 Metric naming rules

**OTEL-REQ-066**
All custom metric names MUST start with `cartulary.` and MUST satisfy the following rules:

| Rule | Required behavior |
| --- | --- |
| Character grammar | Lowercase dotted names using ASCII letters, digits, and `_` inside name components. |
| Unit handling | Metric names MUST NOT include unit suffixes when the unit metadata carries the unit. |
| Duration | Duration histograms MUST end with `.duration` and use unit `s`. |
| Count units | Curly-brace count units MUST use singular quantity annotations such as `{event}`, `{job}`, `{item}`, `{connection}`, `{row}`, or `{conflict}`. |
| Active counts | Active-count metric names MUST use singular object names with `.active`. |
| No `_total` suffix | Counter names MUST NOT use `_total`. |
| Namespace collision | Custom metric names MUST NOT use reserved standard namespaces outside `cartulary.`. |
| Stability | Runtime values MUST NOT appear in metric names. |

### 10.2 Histogram defaults

**OTEL-REQ-067**
The implementation MUST use these explicit histogram bucket boundaries unless a metric registry row declares a different closed bucket set:

| Histogram family | Unit | Explicit bucket boundaries |
| --- | --- | --- |
| Duration | `s` | `0.005`, `0.010`, `0.025`, `0.050`, `0.100`, `0.250`, `0.500`, `1`, `2.5`, `5`, `10` |
| Bytes | `By` | `1024`, `4096`, `16384`, `65536`, `262144`, `1048576`, `4194304`, `16777216`, `67108864`, `268435456` |
| Rows | `{row}` | `1`, `10`, `50`, `100`, `500`, `1000`, `5000`, `10000`, `20000` |

### 10.3 Required custom metric registry

**OTEL-REQ-068**
The implementation MUST emit the custom metrics in this table when the corresponding operation occurs and metrics are enabled:

| Metric name | Instrument kind | Unit | Description | Value type | Aggregation | Temporality | Allowed attributes | Cardinality bound | Reset behavior | Lifecycle behavior | Required bucket boundaries |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `cartulary.workbook.query.duration` | Histogram | `s` | Duration of one workbook view query. | floating point | Explicit bucket histogram | Cumulative unless SDK/exporter requires translation | `cartulary.view_schema_id`, `cartulary.result` | `view_schema_id count * 7 results` | Process-local series reset on process restart. | Record one measurement per workbook query attempt. | Duration defaults. |
| `cartulary.workbook.query.row_count` | Histogram | `{row}` | Rows serialized in one workbook query response. | integer | Explicit bucket histogram | Cumulative unless SDK/exporter requires translation | `cartulary.view_schema_id` | `view_schema_id count` | Process-local series reset on process restart. | Record one measurement per successful query response. | Rows defaults. |
| `cartulary.record.mutation.duration` | Histogram | `s` | Duration of one row-create, patch, delete, restore, rollback, merge, attach, or conflict-resolution attempt. | floating point | Explicit bucket histogram | Cumulative unless SDK/exporter requires translation | `cartulary.view_schema_id`, `cartulary.record_type`, `cartulary.operation`, `cartulary.result` | `view_schema_id count * record_type count * operation count * 7 results` | Process-local series reset on process restart. | Record one measurement per mutation attempt. | Duration defaults. |
| `cartulary.record.mutation.conflict` | Counter | `{conflict}` | Same-field conflict outcomes returned to clients. | integer | Sum | Cumulative | `cartulary.view_schema_id`, `cartulary.record_type` | `view_schema_id count * record_type count` | Counter restarts at zero on process restart. | Increment once per same-field conflict response. | N/A. |
| `cartulary.projection.maintenance.duration` | Histogram | `s` | Duration of one projection-maintenance unit. | floating point | Explicit bucket histogram | Cumulative unless SDK/exporter requires translation | `cartulary.record_type`, `cartulary.result` | `record_type count * 7 results` | Process-local series reset on process restart. | Record one measurement per projection-maintenance unit. | Duration defaults. |
| `cartulary.websocket.connection.active` | ObservableGauge | `{connection}` | Current accepted WebSocket incident subscriptions in the process. | integer | Last value | Gauge observation | none | 1 series | Observed as `0` after process start before accepted connections. | Callback observes current active accepted subscription set. Abnormal close removal MUST be reflected by the next observation. | N/A. |
| `cartulary.websocket.event.sent` | Counter | `{event}` | Attempted server-to-client WebSocket events. | integer | Sum | Cumulative | `cartulary.websocket.event_type`, `cartulary.result` | `event_type count * 7 results` | Counter restarts at zero on process restart. | Increment once per attempted event send. | N/A. |
| `cartulary.job.active` | ObservableGauge | `{job}` | Current active running jobs in the process. | integer | Last value | Gauge observation | `cartulary.job_kind` | `job_kind count` | Observed as `0` after process start before running jobs. | Callback observes current running-job registry. Terminal completion and cancellation MUST be reflected by the next observation. | N/A. |
| `cartulary.job.duration` | Histogram | `s` | Duration of one terminal job run. | floating point | Explicit bucket histogram | Cumulative unless SDK/exporter requires translation | `cartulary.job_kind`, `cartulary.job_terminal_status` | `job_kind count * 4 terminal statuses` | Process-local series reset on process restart. | Record one measurement per terminal job. | Duration defaults. |
| `cartulary.objectstore.operation.duration` | Histogram | `s` | Duration of one object-store operation. | floating point | Explicit bucket histogram | Cumulative unless SDK/exporter requires translation | `cartulary.operation`, `cartulary.result` | `operation count * 7 results` | Process-local series reset on process restart. | Record one measurement per object-store operation. | Duration defaults. |
| `cartulary.objectstore.operation.bytes` | Histogram | `By` | Payload bytes for upload, download, preview-source fetch, or generated-output store operation. | integer | Explicit bucket histogram | Cumulative unless SDK/exporter requires translation | `cartulary.operation` | `operation count` | Process-local series reset on process restart. | Record one measurement when payload byte count is known. | Bytes defaults. |
| `cartulary.postgres.operation.duration` | Histogram | `s` | Duration of one fixed-label Postgres operation or transaction. | floating point | Explicit bucket histogram | Cumulative unless SDK/exporter requires translation | `cartulary.result` | `7 results` | Process-local series reset on process restart. | Record one measurement per named database operation or transaction. | Duration defaults. |
| `cartulary.telemetry.export.duration` | Histogram | `s` | Duration of one telemetry export attempt. | floating point | Explicit bucket histogram | Cumulative unless SDK/exporter requires translation | `cartulary.signal_kind`, `cartulary.telemetry.exporter_kind`, `cartulary.result` | `3 signal kinds * 3 exporter kinds * 7 results` | Process-local series reset on process restart. | Record one measurement per export attempt. This self-metric MUST NOT trigger recursive export telemetry. | Duration defaults. |
| `cartulary.telemetry.item.dropped` | Counter | `{item}` | Telemetry items dropped locally before successful export. | integer | Sum | Cumulative | `cartulary.signal_kind`, `cartulary.drop_reason` | `3 signal kinds * 6 drop reasons` | Counter restarts at zero on process restart. | Increment once per item dropped for queue overflow, redaction rejection, exporter permanent discard, shutdown timeout, metric overflow, or recursion guard. | N/A. |

### 10.4 Metric view and attribute-filter contract

**OTEL-REQ-069**
Every emitted metric, including standard OTel metrics, MUST have an explicit View or equivalent implementation-level filter. The View/filter MUST allow only attributes named in the owning metric registry row. The View/filter MUST reject, drop, or never construct all non-allowlisted attributes before export.

**OTEL-REQ-070**
Standard metrics MUST follow this table:

| Metric source | Current-profile requirement |
| --- | --- |
| Required Cartulary custom metrics | Emit only registry-listed attributes. |
| Standard HTTP metrics | MAY emit only if filtered to allowed stable low-cardinality attributes and no URL path, query, header, client address, server address, incident, user, record, session, or handle values. |
| Standard database metrics | MAY emit only if filtered to allowed stable low-cardinality attributes and no SQL, parameter, table, projection, namespace, server address, or storage-realization values. |
| Runtime or SDK metrics | MAY emit only if resource and metric attributes pass §8 and no high-cardinality process, host, path, incident, user, storage, thread, runtime root, or object-store identifiers appear. |
| Unknown metric instruments | Disabled or dropped. |

**OTEL-REQ-071**
A standard metric with attributes that cannot be filtered safely MUST be disabled. A metric producer that cannot prove safe attribute filtering MUST NOT be enabled in the current profile.

### 10.5 SDK metric cardinality overflow

OpenTelemetry SDK cardinality limits can emit an `otel.metric.overflow=true` synthetic aggregation when unique attribute sets exceed the configured cardinality limit.[^15]

**OTEL-REQ-072**
The implementation MUST configure metric attributes, Views, and cardinality limits so the conformance corpus emits no `otel.metric.overflow` attribute.

**OTEL-REQ-073**
`otel.metric.overflow` is not a Cartulary custom attribute and MUST NOT be treated as a permitted application metric attribute in the current profile. If an SDK nevertheless emits an overflow point in the conformance corpus, conformance fails because it proves the metric registry or View filters are insufficient.

### 10.6 Exemplar policy

OpenTelemetry exemplars can carry trace and span linkage plus filtered measurement attributes, which makes them a possible privacy side channel in Cartulary unless separately specified.[^15]

**OTEL-REQ-074**
Metric exemplars are disabled in the current profile. `telemetry.metrics.exemplars.enabled=true` MUST fail configuration validation.

**OTEL-REQ-075**
A future revision that enables exemplars MUST apply the same forbidden-value, attribute allowlist, cardinality, and trace/span-correlation restrictions as ordinary metric points. Exemplar trace/span linkage MUST NOT become a backdoor for incident, user, evidence, record, session, connection, job, or handle identity.

## 11. Logs and correlation

OpenTelemetry logs define a data model with trace and span correlation, severity, body, resource, instrumentation scope, attributes, and optional event identity.[^16]

### 11.1 Local structured-log fields

**OTEL-REQ-076**
Cartulary MUST use structured application logging as the primary local log substrate. OTel log export is a bridge. Enabling the bridge MUST NOT change local log content, product behavior, or log redaction outcomes.

**OTEL-REQ-077**
When a request, job, WebSocket event, object-store operation, or database operation runs under a sampled trace/span, local structured logs MUST include the available correlation fields in this table:

| Local log field | Requiredness | Rule |
| --- | --- | --- |
| `trace_id` | Required when active span context is valid | Lowercase trace ID from active span context. |
| `span_id` | Required when active span context is valid | Lowercase span ID from active span context. |
| `trace_flags` | Optional when available | Lowercase two-character hex trace flags or equivalent stable string. |
| `cartulary.module` | Required | Closed module value. |
| `cartulary.result` | Required on terminal logs | Closed result value. |
| `cartulary.error_code` | Required on public error logs | Public error-code token or `internal_error`. |

### 11.2 OTel LogRecord mapping

**OTEL-REQ-078**
When `telemetry.logs.bridge_enabled=true`, exported OTel LogRecords MUST use this mapping:

| Local structured-log field | OTel LogRecord field | Required behavior |
| --- | --- | --- |
| active trace ID | `TraceId` | Present when the log is emitted under a valid active span context. |
| active span ID | `SpanId` | Present only when `TraceId` is present. |
| active trace flags | `TraceFlags` | Present when available from active context. |
| local severity enum | `SeverityNumber`, `SeverityText` | Deterministic mapping in §11.3. |
| redacted message | `Body` | String-only in the current profile. MUST NOT contain forbidden values. |
| `cartulary.module` | `Attributes["cartulary.module"]` | Required. |
| `cartulary.result` | `Attributes["cartulary.result"]` | Required on terminal logs. |
| `cartulary.error_code` | `Attributes["cartulary.error_code"]` | Required on public error logs. |
| instrumentation identity | `InstrumentationScope` | Same scope discipline as traces and metrics. |
| resource identity | `Resource` | Same resource contract as §7. |
| event name | `EventName` | Forbidden in the base profile unless a later revision adds a closed event-name registry. |

**OTEL-REQ-079**
In the current profile, OTel LogRecord `Body` MUST be a string. The string MUST be redacted before LogRecord construction and truncated after redaction to `telemetry.logs.body_max_chars` Unicode scalar values. Truncation MUST NOT split a Unicode scalar value. If raw bytes reach the logging boundary, invalid UTF-8 input MUST be decoded with replacement before redaction and truncation.

**OTEL-REQ-080**
`EventName` MUST be absent from all exported LogRecords in the current profile. It MUST NOT be emitted as an empty string or null placeholder.

### 11.3 Severity mapping

**OTEL-REQ-081**
Local severity MUST map to OTel log severity by this table:

| Local severity | `SeverityText` | `SeverityNumber` |
| --- | --- | --- |
| `trace` | `TRACE` | `1` |
| `debug` | `DEBUG` | `5` |
| `info` | `INFO` | `9` |
| `warn` | `WARN` | `13` |
| `error` | `ERROR` | `17` |
| `fatal` | `FATAL` | `21` |

**OTEL-REQ-082**
Logs MUST NOT include forbidden values from §8.3. When log bridge export is enabled, redaction MUST run before the log record reaches an OTel processor, exporter queue, retained exporter artifact, or diagnostic capture.

## 12. Privacy and security invariants

Core 04 requires authenticated mutation origin and audit fidelity, but telemetry does not replace that audit substrate. Core 04 also requires a project-local STRIDE threat model for the current architecture, deployment profiles, and high-risk workflows.[^19]

**OTEL-REQ-083**
Telemetry MUST be classified as operational engineering evidence only. It MUST NOT replace change sets, record revisions, administrative audit records, evidence custody events, snapshot manifests, release approvals, benchmark manifests, or claim-bearing measurement artifacts.

**OTEL-REQ-084**
The project-local STRIDE threat model MUST cover these telemetry-specific scopes before this NLSpec is treated as implemented:

| STRIDE class | Telemetry-specific scope | Required control |
| --- | --- | --- |
| Spoofing | Inbound trace context, fake telemetry source identity | Remote trace context disabled by default; server-owned resource identity. |
| Tampering | Exporter endpoint, telemetry headers, emitted attributes | Config validation, header redaction, closed attribute registry. |
| Repudiation | Telemetry used as audit substitute | Telemetry cannot replace change sets, record revisions, or administrative audit records. |
| Information disclosure | Incident content, credentials, object keys, user identities | Forbidden-value policy, browser-direct-export prohibition, no SQL text export, no object key export. |
| Denial of service | Export queue growth, exporter timeouts, high-cardinality metrics | Queue bounds, export timeouts, attribute registry, cardinality controls, drop metrics. |
| Elevation of privilege | Telemetry endpoint or headers exposed to browser | Server-side-only exporter configuration and no browser direct export. |
| Recursion | Telemetry about telemetry producing more telemetry without bound | Recursion guard and no exporter self-spans. |

**OTEL-REQ-085**
Telemetry self-diagnostics MUST be bounded. The implementation MUST NOT produce unbounded logs, metrics, spans, or retained artifacts in response to exporter failure, processor overflow, redaction failure, or recursion guard activation.

## 13. Exporter, processor, runtime, and shutdown behavior

OTLP exporter configuration defines transport protocols, endpoints, per-signal endpoint overrides, headers, compression, timeouts, and retry behavior. OTLP/HTTP supports binary protobuf and JSON protobuf encodings, but Cartulary adopts only a bounded transport profile.[^17]

### 13.0 Intentional divergence from OTel exporter defaults

**OTEL-REQ-086**
Cartulary MUST apply these intentional divergences from broader OTel exporter capability and defaults:

| OTel exporter capability or default | Cartulary current-profile rule | Reason |
| --- | --- | --- |
| Default OTLP/HTTP endpoint `localhost:4318` | Not adopted. Default exporter is `none`. | Prevent implicit network egress. |
| Default OTLP/gRPC endpoint `localhost:4317` | Not adopted. Endpoint required only when exporter kind is not `none`. | Prevent implicit network egress. |
| Per-signal endpoints | Unsupported and rejected. | Avoid unreviewed signal-specific routing and privacy divergence. |
| Per-signal protocols and headers | Unsupported and rejected. | Avoid unreviewed signal-specific behavior and secret handling. |
| `http/json` OTLP | Unsupported and rejected. | Keep one bounded HTTP encoding profile. |
| OTel header env vars | Not authoritative. | Prevent secret/header injection outside Cartulary config. |
| Declarative config plugin components | Not authoritative. | Prevent unreviewed exporter, processor, reader, sampler, propagator, view, exemplar, or plugin construction. |
| Collector-side scrubbing | Not a conformance control. | Cartulary must redact before SDK processor/exporter handoff. |

### 13.1 Exporter contract

**OTEL-REQ-087**
Exporter behavior MUST follow this table:

| Property | `none` | `otlp_http` | `otlp_grpc` |
| --- | --- | --- | --- |
| Network export | Disabled. | Enabled. | Enabled. |
| Required endpoint | Must be `null` or omitted. | Required `http` or `https` URL. | Required `http` or `https` URL. |
| Protocol | None. | `http/protobuf`. | `grpc`. |
| `http/json` | Unsupported. | Unsupported. | Unsupported. |
| Path construction | None. | Remove one trailing slash from base endpoint, then append `/v1/traces`, `/v1/metrics`, or `/v1/logs` for the signal. | Not applicable. |
| Per-signal endpoint support | Not supported in current revision. | Reject per-signal endpoint keys. | Reject per-signal endpoint keys. |
| Headers | None. | Server-side only; redacted everywhere. | Server-side only; redacted everywhere. |
| Compression | None. | `none` or `gzip`. | `none` or `gzip`. |
| Retry | None. | Bounded retry for transient errors. | Bounded retry for transient errors. |
| Startup unreachable endpoint | No effect. | Must not fail startup. | Must not fail startup. |
| Runtime failure | No effect. | Product behavior unchanged; self-metrics increment. | Product behavior unchanged; self-metrics increment. |

**OTEL-REQ-088**
For `otlp_http`, the configured endpoint MUST be treated as a base endpoint. It MUST NOT already be a per-signal endpoint ending in `/v1/traces`, `/v1/metrics`, or `/v1/logs`. Per-signal endpoint configuration is unsupported in this revision.

### 13.2 Retry contract

**OTEL-REQ-089**
Transient exporter errors MAY be retried only while all of the following remain true:

| Guard | Required behavior |
| --- | --- |
| Retry enabled | `telemetry.exporter.retry.enabled=true`. |
| Elapsed bound | Elapsed retry time is less than or equal to `telemetry.exporter.retry.max_elapsed_ms`. |
| Shutdown state | Process shutdown has not begun. |
| Processor bound | Queue and batch limits remain enforced. |
| Product isolation | Retrying export MUST NOT block product HTTP responses, WebSocket event delivery, mutation commit, evidence access, or job terminal transitions. |

**OTEL-REQ-090**
Retry delay MUST use exponential intervals starting at `telemetry.exporter.retry.initial_interval_ms`, multiplied by `telemetry.exporter.retry.multiplier`, and capped at `telemetry.exporter.retry.max_interval_ms`. Jitter MAY vary each retry delay within `50%..150%` of the computed capped interval. Tests MUST assert the configured bounds rather than exact jitter values.

### 13.3 Processor overflow contract

**OTEL-REQ-091**
Processor behavior MUST follow this table:

| Event | Required behavior |
| --- | --- |
| Processor queue accepts item | Record normally. |
| Processor queue full | Drop the telemetry item being offered. Retain already queued items. Increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='queue_full'`. |
| Redaction rejects item | Drop the item before enqueue. Increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='redaction_rejected'`. |
| Metric overflow observed | Treat conformance corpus emission as a test failure; in production diagnostics increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='metric_overflow'` when safe and non-recursive. |
| Exporter timeout | Mark export attempt as timed out. Product operation remains unchanged. |
| Exporter transient error | Retry within configured retry bound. |
| Exporter permanent error | Drop or retain only according to bounded processor behavior. Product operation remains unchanged. |
| Malformed exporter response | Treat the export attempt as failed or dropped according to bounded exporter behavior. Product operation remains unchanged. |
| Metric reader collection timeout | Product operation continues; bounded diagnostic MAY be emitted. |
| Shutdown flush timeout | Stop waiting after `telemetry.shutdown.flush_timeout_ms`; process shutdown continues. |
| Self-diagnostic emission | Must not recursively create unbounded telemetry about telemetry. |

### 13.4 Error-handling matrix

OpenTelemetry's error-handling guidance treats telemetry as non-essential relative to application behavior, allows fail-fast initialization for bad configuration, and requires runtime telemetry failures not to throw unhandled exceptions or alter application business behavior.[^18]

**OTEL-REQ-092**
Failure handling MUST follow this table:

| Failure point | Required behavior |
| --- | --- |
| Invalid Cartulary telemetry configuration | Fail before readiness. |
| OTel SDK invalid-use panic risk | Contain through bootstrap validation or test-only strict handling; production MUST NOT expose unhandled telemetry exceptions. |
| Exporter endpoint unreachable at startup | Application starts. |
| Runtime exporter failure | Product request, job, WebSocket, evidence access, and mutation continue. |
| Malformed exporter response | Export attempt fails or is dropped according to bounded exporter behavior; product operation continues. |
| Exporter TLS or connection failure | Product operation continues; bounded self-diagnostic or self-metric records the failure. |
| SDK callback, processor, reader, or exporter runtime error | Must be contained so production product behavior does not fail due to telemetry runtime error. |
| Processor queue overflow | Drop according to declared policy and increment drop metric. |
| Redaction failure | Drop affected telemetry item. |
| Log bridge export failure | Local logs remain intact; exported log item may be dropped. |
| Metric reader collection timeout | Product operation continues; bounded diagnostic MAY be emitted. |
| Shutdown flush timeout | Continue shutdown after timeout. |

### 13.5 Startup and shutdown contract

**OTEL-REQ-093**
Startup behavior MUST satisfy these rules:

| Rule | Required behavior |
| --- | --- |
| Telemetry disabled | If `telemetry.enabled=false`, startup installs no-op telemetry providers and MUST NOT create exporters. |
| Invalid config | Syntactically or semantically invalid telemetry config fails deployment configuration validation. |
| Endpoint unreachable | A syntactically valid but unreachable endpoint does not fail startup. |
| Initialization order | Telemetry initialization completes before HTTP listener, WebSocket listener, and background-job runner startup so startup diagnostics can be correlated after initialization. |
| SDK env defaults | SDK defaults cannot enable export outside the Cartulary effective configuration envelope. |

**OTEL-REQ-094**
Shutdown behavior MUST satisfy these rules:

| Rule | Required behavior |
| --- | --- |
| Flush request | On graceful shutdown, the implementation requests trace, metric, and enabled log bridge flush. |
| Timeout | Flush wait uses `telemetry.shutdown.flush_timeout_ms`. |
| Product shutdown | Flush timeout MUST NOT prevent process shutdown after the timeout expires. |
| Diagnostics | Shutdown records a bounded local diagnostic when telemetry flush times out. |
| Drop metric | Items abandoned because of flush timeout increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='shutdown_timeout'` when self-diagnostics are enabled and the increment can occur without recursion. |

### 13.6 Telemetry-about-telemetry recursion guard

**OTEL-REQ-095**
The implementation MUST NOT emit OpenTelemetry spans for telemetry export attempts. Telemetry export attempts MAY emit only the self-metrics in §10.3 and bounded local diagnostics. Those self-metrics MUST NOT themselves create additional telemetry export spans, span events, or recursive self-metrics.

## 14. Browser telemetry boundary and non-transfer rules

### 14.1 Browser boundary

**OTEL-REQ-096**
Browser direct export is forbidden. Browser code MUST NOT contain direct OTLP, Prometheus, Zipkin, Jaeger native, vendor telemetry, or third-party telemetry endpoint configuration.

**OTEL-REQ-097**
The browser MAY create local performance marks for workbook controller diagnostics, but those marks MUST remain local unless a later adopted revision defines an app-mediated client telemetry route.

**OTEL-REQ-098**
A future app-mediated browser telemetry route MUST define, at minimum:

| Required future contract element |
| --- |
| Route path. |
| Authentication context. |
| CSRF behavior. |
| Request schema. |
| Batch size limits. |
| Allowed event names. |
| Allowed attributes. |
| Redaction rules. |
| Replay and idempotency behavior. |
| Rejection errors. |
| Sampling behavior. |
| Whether events are emitted as OTel logs, spans, metrics, or local diagnostics. |

### 14.2 Concepts not transferred into Cartulary

**OTEL-REQ-099**
The following OTel concepts or ecosystem patterns MUST NOT be transferred into the Cartulary base telemetry profile:

| OTel concept or ecosystem pattern | Cartulary base-profile treatment |
| --- | --- |
| Browser direct OTLP export | Forbidden. |
| Browser vendor telemetry endpoint | Forbidden. |
| Browser-configured exporter headers | Forbidden. |
| Baggage for incident, user, party, evidence, saved-view, job, or record correlation | Forbidden. |
| Inbound `traceparent` as trusted parentage | Ignored in current profile. |
| Inbound `tracestate` as trusted identity | Ignored in current profile. |
| OpenTelemetry Collector as required deployable | Forbidden as a conformance dependency. Optional external receiver only. |
| Collector-side scrubbing as privacy conformance | Not sufficient. Redaction must happen before SDK/exporter handoff. |
| Prometheus scrape endpoint | Deferred. Not part of this revision. |
| SQL commenter or database context propagation | Forbidden by default and not configurable in this revision. |
| OTel SDK environment autoconfiguration | Not authoritative. MUST NOT override Cartulary configuration. |
| OTel declarative config as telemetry authority | Forbidden as an authority; only declared `telemetry.*` keys may affect behavior. |
| OTel plugin providers as deployment-time behavior | Forbidden unless a future revision defines a closed plugin profile. |
| Default localhost OTLP export | Not adopted. Default exporter remains `none`. |
| Per-signal OTLP endpoints | Rejected in current profile. |
| Per-signal OTLP protocols and headers | Rejected in current profile. |
| OTLP `http/json` | Rejected in current profile. |
| `OTEL_SEMCONV_STABILITY_OPT_IN` shape changes | Ignored unless a future revision defines the migration. |
| AWS S3 semantic-convention attributes | Not adopted in the base telemetry profile. |
| OTel metric-pipeline non-sanitization guidance as a privacy rule | Not adopted. Cartulary validates and redacts before recording because incident content and identifiers are forbidden. |
| OTel exemplars | Disabled in current profile. |
| `service.criticality` | Not emitted in current profile. |

## 15. Verification and acceptance criteria

### 15.1 Configuration and bootstrap criteria

- **OTEL-AC-001:** With all telemetry keys omitted, effective config resolves to `telemetry.enabled=true`, `telemetry.exporter.kind='none'`, tracing enabled, metrics enabled, log bridge disabled, exemplars disabled, and no network export.
- **OTEL-AC-002:** With `telemetry.enabled=false`, no OTel SDK providers, exporters, metric readers, log processors, or log bridge are installed except no-op placeholders needed for code safety.
- **OTEL-AC-003:** Invalid config combinations in OTEL-REQ-026 fail before readiness with the configured deployment-config error path.
- **OTEL-AC-004:** A syntactically valid but unreachable OTLP endpoint does not fail startup.
- **OTEL-AC-005:** With `telemetry.exporter.kind='none'`, no outbound telemetry occurs when global OTLP env vars, per-signal OTLP env vars, per-signal protocol/header vars, `OTEL_CONFIG_FILE`, and `OTEL_SEMCONV_STABILITY_OPT_IN` are set.
- **OTEL-AC-006:** With `telemetry.otel_env_passthrough=true`, an explicit Cartulary config value wins over conflicting OTel environment variables.
- **OTEL-AC-007:** OTel declarative config cannot create exporters, processors, propagators, samplers, metric readers, log processors, header capture, metric views, exemplars, or plugin components in the current profile.

### 15.2 API, SDK, and instrumentation-scope criteria

- **OTEL-AC-008:** A static import-boundary test fails if any ordinary instrumentation module imports OTel SDK provider, exporter, processor, sampler, propagator, metric-reader, log-processor, declarative-config, or SDK-autoconfiguration packages.
- **OTEL-AC-009:** Every tracer, meter, and logger uses a registered instrumentation scope name, the build-version fallback rule, `schema_url=null`, and no scope attributes in the current profile.
- **OTEL-AC-010:** A dependency-only SDK update with registry-equivalent emitted telemetry passes existing acceptance criteria without requiring an NLSpec revision.
- **OTEL-AC-011:** Any additive, privacy-tightening, or breaking telemetry-shape change requires a migration note and matching acceptance criteria.

### 15.3 Resource and attribute criteria

- **OTEL-AC-012:** Every exported signal contains the required resource attributes in §7 and contains no forbidden resource attributes.
- **OTEL-AC-013:** `service.instance.id` is stable across multiple requests, jobs, metric collections, and log exports in one process and differs across two default-generated process starts.
- **OTEL-AC-014:** `service.criticality` is absent from resource attributes.
- **OTEL-AC-015:** The custom attribute registry is closed: every emitted `cartulary.*` attribute appears in §8.4 and has the declared type and allowed-value behavior.
- **OTEL-AC-016:** The forbidden-value test injects representative credentials, incident text, evidence filenames, object keys, SQL bind values, user IDs, record IDs, incident IDs, client transaction IDs, handle tokens, Baggage values, and exporter headers, then verifies none appear in exported spans, metrics, logs, self-diagnostics, SDK processor inputs, exporter queues, or retained telemetry test artifacts.
- **OTEL-AC-017:** A redaction failure drops the telemetry item, increments the drop metric when non-recursive, and does not fail the product operation.
- **OTEL-AC-018:** A Collector, backend transform, or exporter-side scrubber is not required to pass forbidden-value criteria.

### 15.4 Trace and span criteria

- **OTEL-AC-019:** HTTP spans for routes containing concrete IDs are never exported with concrete path-derived names, `url.path`, `url.query`, raw route params, concrete IDs, logs, diagnostics, or retained artifacts.
- **OTEL-AC-020:** A request to `/api/v1/records/{record_id}` with a concrete record ID exports only the route-template span name and `http.route`; the concrete path and query are absent from all spans, metrics, logs, self-diagnostics, and retained telemetry artifacts.
- **OTEL-AC-021:** A route miss or unknown-route request uses a low-cardinality fallback span name and does not emit the raw path.
- **OTEL-AC-022:** HTTP server 4xx responses leave span status unset unless a narrower owner rule marks a real implementation error; HTTP server 5xx responses set error status.
- **OTEL-AC-023:** Postgres spans include `db.system.name='postgresql'` and omit SQL text, bind values, table names, projection table names, saved-view query JSON, search text, `db.namespace`, `db.collection.name`, `server.address`, and `server.port`.
- **OTEL-AC-024:** Object-store telemetry is a Cartulary dependency span/metric family and never emits S3 semantic attributes or object-identifying values.
- **OTEL-AC-025:** Upload, preview-source fetch, download-source fetch, generated-output write, and failed object-store operation tests prove no S3 semantic attributes, bucket names, object keys, upload IDs, copy sources, filenames, blob hashes, evidence handles, storage refs, or evidence titles are emitted.
- **OTEL-AC-026:** Job execution spans created from asynchronous enqueue are root `CONSUMER` spans with enqueue links added at span creation and are not children of returned HTTP requests.
- **OTEL-AC-027:** A batch or replay operation with two causal inputs uses links and no false single parent.
- **OTEL-AC-028:** Requests carrying `traceparent`, `tracestate`, and `baggage` produce server-owned root traces, no Baggage-derived attributes, no Baggage-derived logs, and no sampling or routing change.
- **OTEL-AC-029:** Exception recording exports `exception.type` only when safe and never exports `exception.message` or `exception.stacktrace`.

### 15.5 Metrics criteria

- **OTEL-AC-030:** Every custom metric has one registry entry defining exact instrument kind, unit, description, allowed attributes, aggregation, lifecycle behavior, cardinality bound, reset behavior, and bucket boundaries when applicable.
- **OTEL-AC-031:** Active WebSocket and active job metrics return to zero on normal close, abnormal close, terminal job completion, cancellation, and process restart according to the declared reset semantics.
- **OTEL-AC-032:** Histogram bucket-boundary tests prove workbook duration, DB duration, object-store duration, byte count, and row count measurements use the registered explicit buckets.
- **OTEL-AC-033:** Every emitted metric has an effective View/filter matching its registry row; injected non-allowlisted attributes are absent.
- **OTEL-AC-034:** Standard HTTP, database, runtime, and SDK metrics are disabled when their attributes cannot be filtered to the current-profile allowlist.
- **OTEL-AC-035:** The conformance corpus emits no `otel.metric.overflow`.
- **OTEL-AC-036:** No metric point contains exemplars when `telemetry.metrics.exemplars.enabled` is omitted or explicitly `false`.
- **OTEL-AC-037:** Setting `telemetry.metrics.exemplars.enabled=true` fails deployment-configuration validation in this revision.

### 15.6 Logs criteria

- **OTEL-AC-038:** Local structured logs under an active span include `trace_id` and `span_id` when available.
- **OTEL-AC-039:** When the OTel log bridge is enabled, exported LogRecords map trace ID, span ID, trace flags, severity, body, resource, instrumentation scope, and attributes according to the log mapping table.
- **OTEL-AC-040:** Log bridge export emits string-only `Body`.
- **OTEL-AC-041:** A long redacted message is truncated to the configured character bound after redaction without splitting a Unicode scalar value.
- **OTEL-AC-042:** Invalid UTF-8 input is made safe before export.
- **OTEL-AC-043:** `EventName` is absent from all exported LogRecords.
- **OTEL-AC-044:** OTel log export emits no forbidden values in body, attributes, event name, exception fields, resource attributes, or retained telemetry artifacts.

### 15.7 Exporter, processor, runtime, and shutdown criteria

- **OTEL-AC-045:** Exporter kind `none` creates no outbound OTLP request and no exporter instance capable of network egress.
- **OTEL-AC-046:** OTLP HTTP export uses only `http/protobuf`; `http/json` configuration fails validation.
- **OTEL-AC-047:** Per-signal endpoint, protocol, or header configuration is rejected or ignored according to §6.3 and cannot change routing or enable export.
- **OTEL-AC-048:** Exporter outage, retry exhaustion, TLS failure, connection refusal, malformed remote responses, queue overflow, redaction rejection, metric reader collection timeout, and shutdown flush timeout do not fail workbook HTTP requests, WebSocket processing, evidence access decisions, or background jobs.
- **OTEL-AC-049:** Shutdown attempts telemetry flush and exits after the configured flush timeout even when the exporter does not respond.
- **OTEL-AC-050:** Telemetry export attempts do not create spans.
- **OTEL-AC-051:** Telemetry export failure, queue overflow, redaction rejection, and flush timeout increment bounded self-metrics or bounded local diagnostics without recursive telemetry generation.

### 15.8 Browser and non-transfer criteria

- **OTEL-AC-052:** Browser code contains no direct OTLP, Prometheus, Zipkin, Jaeger native, vendor telemetry, or third-party telemetry export endpoint configuration.
- **OTEL-AC-053:** Browser local performance marks, if present, remain local and are not exported without a later app-mediated telemetry route contract.
- **OTEL-AC-054:** Baggage is not used to propagate incident, record, user, party, evidence, saved-view, job, handle, or customer identifiers.
- **OTEL-AC-055:** The Collector is not required for Cartulary telemetry conformance; a deployment with `telemetry.exporter.kind='none'` passes the default telemetry criteria without any Collector present.
- **OTEL-AC-056:** A non-transfer audit test proves each row in §14.2 is enforced by configuration validation, static import checks, emitted telemetry assertions, retained-artifact assertions, or browser bundle inspection.

### 15.9 Security and conformance criteria

- **OTEL-AC-057:** The project-local STRIDE threat model includes telemetry exporter credentials, telemetry retention, inbound trace context, SDK environment override risk, declarative configuration risk, telemetry queue denial-of-service, metric cardinality overflow, exemplar risk, and telemetry self-diagnostic recursion.
- **OTEL-AC-058:** Telemetry emitted during functional tests is classified as operational engineering evidence only and is not used as claim-bearing benchmark evidence unless Core 05 publication requirements are separately satisfied.
- **OTEL-AC-059:** The emitted telemetry conformance corpus can be validated without reading product databases, object-store evidence bytes, Collector-side transformed data, vendor dashboards, or backend query results.
- **OTEL-AC-060:** Every emitted telemetry name, attribute, metric, span, log mapping, exporter setting, and processor setting has exactly one owner table in this NLSpec.
- **OTEL-AC-061:** Acceptance evidence proves the default profile emits no raw incident content, no stable product identifiers, no SQL text, no DB parameters, no object keys, no filenames, no evidence handles, no Baggage values, no exporter headers, no exemplars, and no metric overflow attribute.

## 16. Open decisions

**OTEL-REQ-100**
Open decisions MUST use this disposition table:

| ID | Decision | Disposition |
| --- | --- | --- |
| `OTEL-DQ-001` | Adopted spec path | Open. Choose final repository path and adoption process. Suggested path remains `docs/cartulary-opentelemetry-instrumentation-nlspec.md`. |
| `OTEL-DQ-002` | Exact Go OpenTelemetry SDK versions | Open. Pin exact module versions in repo-control files before implementation adoption. |
| `OTEL-DQ-003` | Semantic-convention update policy | Closed by §4.1 through §4.4. Semantic-convention updates are spec revisions unless emitted telemetry remains `registry_equivalent`. |
| `OTEL-DQ-004` | Browser telemetry | Deferred. Browser direct export remains forbidden; app-mediated telemetry requires a later route contract. |
| `OTEL-DQ-005` | Prometheus compatibility | Deferred. Prometheus scrape support is not in this revision and MUST NOT be implied by OTel exporter lists. |
| `OTEL-DQ-006` | Incident correlation | Narrowed. Default remains `none`; `hmac_64bit` requires production-specific approval through configuration, server-side secret, bounded attribute use, and the collision-risk posture in §8.5. |
| `OTEL-DQ-007` | Operator dashboards and alerts | Closed as separate SRE/runbook scope. This NLSpec owns emitted telemetry, not dashboard artifacts. |
| `OTEL-DQ-008` | Metric exemplars | Closed for this revision. Disabled and invalid if enabled. |
| `OTEL-DQ-009` | OTel declarative config and plugin providers | Closed for this revision. Not authoritative and cannot create SDK behavior. |
| `OTEL-DQ-010` | HTTP `url.path` adoption | Closed for this revision. Forbidden; `http.route` is the only path-like HTTP attribute. |

## 17. Completion standard

This revision is complete only when all of the following are true:

- The existing authority model remains intact.
- Every emitted telemetry name, attribute, metric, span, log mapping, exporter setting, and processor setting has one owner table in this NLSpec.
- Every configurable telemetry value has a default, allowed values, bounds, omitted behavior, explicit-`null` behavior, and startup-validation behavior.
- Every standard OTel semantic convention used by Cartulary is classified by stability and privacy safety.
- Every Cartulary custom key is listed in the custom registry and passes the naming policy.
- Every forbidden value family has tests covering traces, metrics, logs, events, exception fields, exporter artifacts, self-diagnostics, SDK processor inputs, exporter queues, and retained test artifacts.
- OTel environment variables, declarative config, plugin providers, and semconv opt-ins cannot alter telemetry behavior outside declared `telemetry.*` keys.
- No ordinary instrumentation unit can configure SDK providers, exporters, processors, samplers, propagators, metric readers, log processors, or declarative config.
- HTTP spans never emit concrete paths, concrete IDs, raw query strings, headers, or path-derived names.
- Postgres spans emit `db.system.name='postgresql'` and never emit SQL text, bind values, table names, projection names, DB namespace, collection name, server address, or server port.
- Object-store telemetry never emits S3 semantic attributes or object-identifying values.
- Every metric has an allowlisted View/filter; no `otel.metric.overflow` appears in the conformance corpus.
- Exemplars are disabled.
- LogRecord bodies are bounded, string-only, and redacted before construction.
- Browser direct export, Collector requirement, Prometheus scrape export, SQL commenter propagation, Baggage correlation, AWS S3 attributes, OTel environment autoconfiguration, OTel declarative config authority, OTel plugin providers, raw HTTP URLs, raw SQL, raw DB parameters, object keys, filenames, and evidence handles are explicitly forbidden or explicitly deferred.
- The acceptance criteria are binary and sufficient for an implementation agent to determine conformance without guessing.

## Sources

[^1]: `opentelemetry-instrumentation-nlspec.md`, uploaded draft, §§1-17. The revised document uses that draft as the base artifact and preserves its draft/proposed authority boundary while adding the gap closures requested by the revision plan.

[^2]: `00_document_set_status_and_precedence.md`, Core 00 §1-§4. Core 00 defines the current normative core, separates Core 05 from implementation conformance, and places future adopted Cartulary NLSpecs above the current core in precedence.

[^3]: `05_claim_publication_and_benchmark_reproducibility.md`, Core 05 §1-§4. Core 05 separates implementation correctness from claim-bearing timed or fixture-sensitive benchmark publication.

[^4]: `01_architecture_storage_and_view_contracts.md`, Core 01 §1-§2. Core 01 defines the modular monolith, the single application deployable, Postgres, S3-compatible object storage, and logical internal module boundaries.

[^5]: `04_security_deployment_and_conformance.md`, Core 04 §6 and §12. Core 04 owns runtime roots and the operator-facing deployment configuration surface.

[^6]: OpenTelemetry Specification page, `https://opentelemetry.io/docs/specs/otel/`, lines 122-123 as observed on 2026-05-20. The page identifies the current observed specification as `OpenTelemetry Specification 1.56.0`.

[^7]: OpenTelemetry Overview, `https://opentelemetry.io/docs/specs/otel/overview/`, lines 151-166 as observed on 2026-05-20. The page defines the API/SDK split, states that instrumentation authors must not directly reference SDK packages, and states that semantic-convention YAML files are the source of truth for generated constants.

[^8]: OpenTelemetry Configuration, `https://opentelemetry.io/docs/specs/otel/configuration/`, lines 127-145 as observed on 2026-05-20. The page describes programmatic, environment-variable, and declarative configuration, including SDK component configuration, instrumentation configuration, custom plugin components, and file-based representation.

[^9]: OpenTelemetry Environment Variable Specification, `https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/`, lines 142-153 as observed on 2026-05-20. The page describes environment-variable configuration goals and implementation guidance.

[^10]: OpenTelemetry Service attribute registry, `https://opentelemetry.io/docs/specs/semconv/registry/attributes/service/`, lines 352-360 as observed on 2026-05-20. The page marks `service.criticality` Development, defines `service.instance.id`, and recommends opaque UUID-style service-instance identity because underlying identity can be confidential.

[^11]: OpenTelemetry HTTP span semantic conventions, `https://opentelemetry.io/docs/specs/semconv/http/http-spans/`, lines 356-390, 398-415, 632-648, and 689-690 as observed on 2026-05-20. The page identifies HTTP span conventions as stable, defines span naming around method and target, documents semconv migration opt-ins, specifies server 4xx and 5xx status handling, and defines low-cardinality `http.route` behavior.

[^12]: OpenTelemetry Database span semantic conventions, `https://opentelemetry.io/docs/specs/semconv/db/database-spans/`, lines 347-380, 419-453, 482-490, and 494-529 as observed on 2026-05-20. The page identifies database span conventions as stable, defines database span naming, query summary, query parameters, and the well-known `postgresql` `db.system.name` value.

[^13]: OpenTelemetry AWS S3 semantic conventions, `https://opentelemetry.io/docs/specs/semconv/object-stores/s3/`, lines 336-359 and 364-392 as observed on 2026-05-20. The page marks S3 conventions Development and lists object-identifying attributes such as bucket, key, copy source, part number, and upload ID.

[^14]: OpenTelemetry Overview, `https://opentelemetry.io/docs/specs/otel/overview/`, lines 209-228 and 271-280 as observed on 2026-05-20. The page defines span state, span links, and Baggage consumption into metrics, logs, and traces.

[^15]: OpenTelemetry Overview and Metrics SDK, `https://opentelemetry.io/docs/specs/otel/overview/` and `https://opentelemetry.io/docs/specs/otel/metrics/sdk/`, lines 254-264, 263-325, 554-567, and 660-665 as observed on 2026-05-20. These pages describe metric instrument families, Views, attribute filtering, cardinality overflow, and exemplars.

[^16]: OpenTelemetry Logs Data Model, `https://opentelemetry.io/docs/specs/otel/logs/data-model/`, lines 147-181 as observed on 2026-05-20. The page defines the stable log data model and top-level LogRecord field families.

[^17]: OpenTelemetry Protocol Exporter and OTLP Specification, `https://opentelemetry.io/docs/specs/otel/protocol/exporter/` and `https://opentelemetry.io/docs/specs/otlp/`, lines 135-149 and 1011-1043 as observed on 2026-05-20. These pages define OTLP exporter endpoint configuration, per-signal endpoint precedence, default endpoints, OTLP/gRPC and OTLP/HTTP ports, and protobuf binary/JSON encodings.

[^18]: OpenTelemetry Error Handling, `https://opentelemetry.io/docs/specs/otel/error-handling/`, lines 127-136 as observed on 2026-05-20. The page states that telemetry should not significantly change application behavior, permits fail-fast initialization for bad configuration, and forbids unhandled runtime telemetry exceptions.

[^19]: `04_security_deployment_and_conformance.md`, Core 04 §3-§4.5. Core 04 requires a project-local STRIDE threat model and defines relevant trust-boundary controls.
