---
title: Cartulary OpenTelemetry Instrumentation NLSpec
status: draft/proposed
document_class: nlspec
version: 0.2.0-draft
created_at: 2026-05-19
suggested_repository_path: docs/cartulary-opentelemetry-instrumentation-nlspec.md
---

# Cartulary OpenTelemetry Instrumentation NLSpec

## 1. Status, scope, and authority

Status: `draft/proposed`.

This NLSpec defines Cartulary's OpenTelemetry instrumentation subsystem. It is not adopted implementation-conformance authority until the Cartulary repository authority process adopts it.

**OTEL-REQ-001**
This NLSpec governs only telemetry generation, telemetry configuration, telemetry export, log correlation, signal naming, attribute governance, privacy boundaries, telemetry runtime behavior, and instrumentation verification.

**OTEL-REQ-002**
This NLSpec MUST NOT redefine product behavior owned by Cartulary Core 00 through Core 04. It MUST NOT redefine claim-bearing benchmark publication owned by Core 05. Runtime telemetry MAY support engineering diagnosis and operational SRE practice, but telemetry observations MUST NOT satisfy claim-bearing timed or fixture-sensitive publication unless the Core 05 benchmark-manifest and measurement-predicate requirements are also satisfied.[^1][^2]

**OTEL-REQ-003**
When this NLSpec conflicts with Core 00 through Core 04 before adoption, the conflict is a draft defect in this NLSpec. When a later adopted version of this NLSpec conflicts with older non-normative appendices or guides, the adopted NLSpec governs only the telemetry subsystem.

## 2. Purpose

**OTEL-REQ-004**
Cartulary MUST provide first-class observability because long-term support requires operators to diagnose availability, latency, error, queueing, persistence, evidence access, and collaboration failures without inspecting incident content or weakening the workbook hot path.

The instrumentation subsystem MUST make these operational questions answerable:

| Question | Required signal support |
| --- | --- |
| Is the application deployable accepting and completing HTTP requests? | HTTP traces and metrics. |
| Are workbook queries, mutations, and projection updates healthy? | Workbook and projection spans plus duration and row-count metrics. |
| Are WebSocket subscriptions, presence updates, and live row updates healthy? | WebSocket metrics and bounded operation spans. |
| Are background jobs queued, running, canceled, failed, or completing? | Job lifecycle spans, job active gauges, terminal duration metrics, and terminal-status attributes. |
| Are Postgres or object-storage dependencies degraded? | Dependency spans, dependency duration metrics, and low-cardinality error classification. |
| Are telemetry exporters failing or dropping data? | Telemetry self-metrics and bounded local diagnostics. |
| Can operators correlate local logs with traces without exposing secrets or incident content? | Trace correlation fields, log-record mapping, and redaction rules. |

## 3. Non-goals

**OTEL-REQ-005**
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

## 4. External standard baseline

OpenTelemetry is the selected telemetry framework because it is the vendor-neutral project baseline for generating and exporting observability signals. OpenTelemetry clients separate API, SDK, semantic conventions, and contrib packages; instrumentation authors are required to depend on API packages rather than SDK packages.[^3]

### 4.1 Baseline object

**OTEL-REQ-006**
The initial external standard baseline MUST be the closed object in this table:

| Field | Required value or rule |
| --- | --- |
| `otel_spec_version` | `1.56.0` until this NLSpec is revised.[^4] |
| `otel_spec_source` | `https://opentelemetry.io/docs/specs/otel/`. |
| `otel_spec_observed_at` | `2026-05-19`. |
| `semconv_version` | `1.41.0` until this NLSpec is revised.[^5] |
| `semconv_source` | `https://opentelemetry.io/docs/specs/semconv/`. |
| `semconv_model_source` | Semantic-convention YAML model files from the pinned semantic-conventions source, not prose-only extraction.[^6] |
| `semconv_generated_constants_version` | Exact generated-constant package or code-generation source version used by the implementation. This value MUST be pinned in repo-control files before implementation adoption. |
| `semconv_stability_policy` | The policy in §4.3. |
| `migration_note_required` | `true` for any baseline change. |

**OTEL-REQ-007**
Any later adopted revision that changes `otel_spec_version`, `semconv_version`, emitted metric names, span names, attribute names, resource attributes, stability status, default adoption, or privacy exclusions MUST include a migration note identifying the changed telemetry shape and dashboard, alert, test, or exporter consequences.

**OTEL-REQ-008**
Dependency-only updates MAY occur without an NLSpec revision only when emitted telemetry remains registry-equivalent for the conformance test corpus. Registry-equivalent means the same spans, span names, span kinds, metrics, metric instrument identities, resource attributes, log mappings, standard attributes, custom attributes, and forbidden-value exclusions are emitted for the same tested operations.

### 4.2 OpenTelemetry component boundary

**OTEL-REQ-009**
Cartulary MUST use the component meanings in this table:

| Term | Required Cartulary meaning |
| --- | --- |
| `OpenTelemetry API` | The only OTel package family that ordinary Cartulary instrumentation code may call directly. |
| `OpenTelemetry SDK` | Installed, configured, and shut down only by the server-side telemetry bootstrap boundary. |
| `Instrumentation library` | Code that records telemetry for a Cartulary module through the OTel API and MUST NOT configure exporters, processors, SDK providers, or environment autoconfiguration. |
| `Instrumentation scope` | The OTel `(name, version, schema_url, attributes)` identity used when obtaining tracers, meters, or loggers for a Cartulary instrumentation unit. |
| `Exporter` | Server-side component that sends telemetry to the configured OTLP endpoint. Exporters are never configured by browser code. |
| `Processor` | Server-side component that batches, bounds, drops, flushes, and forwards telemetry to exporters. |
| `Collector` | Optional external receiver. It is not a Cartulary deployable and is not required for Cartulary telemetry conformance. |
| `Semantic conventions` | OTel standard names and meanings for common telemetry concepts. They are adopted by stability policy and registry generation, not copied ad hoc. |

### 4.3 Semantic-convention stability policy

**OTEL-REQ-010**
Cartulary MUST apply this semantic-convention adoption matrix:

| OTel convention status | Cartulary default |
| --- | --- |
| Stable and applicable | Emit by default when it does not violate Cartulary privacy, cardinality, or deployment-boundary rules. |
| Stable but privacy-conflicting | Do not emit the conflicting attribute. Record the omission in the signal-specific table that would otherwise own it. |
| Development or experimental | Do not emit by default. Adoption requires an explicit NLSpec revision or an explicit opt-in configuration key defined by this NLSpec. |
| Deprecated | Do not emit unless a migration-compatibility profile explicitly requires it. |
| Migration-period duplicated conventions | Do not duplicate by default. Duplication requires a bounded compatibility rule and an acceptance criterion proving both old and new forms are intentional. |
| Unknown or unpinned | Do not emit. |

**OTEL-REQ-011**
Every emitted standard attribute and standard metric name MUST be generated or imported from the pinned semantic-convention model source, or MUST be explicitly listed as a standard attribute allowlist exception in the signal registry. Every emitted Cartulary custom attribute MUST be listed in §8.4.

## 5. Instrumentation ownership

The application deployable owns the browser-facing UI host, API surface, WebSocket hub, and background-job runners because Cartulary's base deployment is one web application deployable plus Postgres and S3-compatible object storage.[^7]

**OTEL-REQ-012**
The instrumentation subsystem MUST be a logical internal boundary inside the modular monolith. It MUST NOT require a separate application deployable, sidecar, microservice, Collector, vendor backend, Prometheus server, or browser telemetry service.

**OTEL-REQ-013**
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
| Telemetry bootstrap | Server-side telemetry boundary | SDK provider setup, processors, exporters, shutdown, and self-diagnostics. | Ordinary instrumentation libraries MUST NOT configure SDK, exporter, processor, or Collector behavior. |

## 6. Configuration contract

Telemetry configuration lives in the Cartulary deployment configuration surface. Core 04 owns the operator-facing deployment configuration artifact, discovery precedence, binding keys, and fail-closed startup validation; this NLSpec adds telemetry keys under that same surface rather than defining a second configuration model.[^8]

### 6.1 Configuration keys

**OTEL-REQ-014**
The effective telemetry configuration MUST be the closed key set in this table. Unknown `telemetry.*` keys are invalid unless a later revision defines them.

| Key | Type | Default | Bounds or values | Omitted behavior | Explicit `null` behavior | Required behavior |
| --- | --- | --- | --- | --- | --- | --- |
| `telemetry.enabled` | boolean | `true` | `true`, `false` | Use default. | Invalid. | When `false`, no OpenTelemetry providers, exporters, log bridges, or instrumentation hooks are installed except no-op placeholders needed for code safety. |
| `telemetry.otel_env_passthrough` | boolean | `false` | `true`, `false` | Use default. | Invalid. | When `false`, OTel SDK environment variables MUST NOT enable exporters, propagators, samplers, processors, endpoints, headers, or config files. |
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
| `telemetry.traces.sample_ratio` | decimal | `0.10` | `0.0..1.0` inclusive | Use default. | Invalid. | Uses parent-based trace-id ratio sampling over server-owned trace IDs. |
| `telemetry.traces.accept_remote_context` | boolean | `false` | exactly `false` in this revision | Use default. | Invalid. | Remote trace context is not trusted in this revision. |
| `telemetry.metrics.enabled` | boolean | `true` | `true`, `false` | Use default. | Invalid. | Has effect only when `telemetry.enabled=true`. |
| `telemetry.logs.bridge_enabled` | boolean | `false` | `true`, `false` | Use default. | Invalid. | When `false`, local structured logs MAY include trace correlation fields but MUST NOT be exported as OpenTelemetry logs. |
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
| `telemetry.attribute.incident_correlation` | enum | `none` | `none`, `hmac_64bit` | Use default. | Invalid. | `none` forbids incident correlation attributes. `hmac_64bit` is a narrowed opt-in under §8.4. |
| `telemetry.attribute.hmac_secret_ref` | string or null | `null` | Server-side secret reference | Required only when incident correlation is `hmac_64bit`. | Valid only when incident correlation is `none`; otherwise invalid. | Secret value MUST NOT be exported. |

### 6.2 Configuration precedence

**OTEL-REQ-015**
Configuration precedence MUST be exactly this table:

| Precedence | Source | Required behavior |
| --- | --- | --- |
| 1 | Cartulary deployment configuration | Authoritative for all telemetry behavior. |
| 2 | Cartulary server-side environment bindings | MAY populate Cartulary deployment configuration keys only. Empty values are treated as omitted. |
| 3 | OTel SDK environment variables | Ignored unless `telemetry.otel_env_passthrough=true`. |
| 4 | OTel SDK defaults | MAY apply only inside the effective Cartulary configuration envelope. MUST NOT enable export when Cartulary export is `none`. |
| 5 | Browser state or browser environment | Never a telemetry exporter configuration source. |

**OTEL-REQ-016**
When `telemetry.otel_env_passthrough=false`, the implementation MUST ignore `OTEL_TRACES_EXPORTER`, `OTEL_METRICS_EXPORTER`, `OTEL_LOGS_EXPORTER`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_HEADERS`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_PROPAGATORS`, `OTEL_TRACES_SAMPLER`, `OTEL_BSP_*`, `OTEL_BLRP_*`, `OTEL_METRIC_EXPORT_INTERVAL`, and `OTEL_CONFIG_FILE` for behavior selection. It MAY retain their presence only in redacted startup diagnostics that do not expose values.

**OTEL-REQ-017**
When `telemetry.otel_env_passthrough=true`, OTel SDK environment variables MAY fill only otherwise omitted OTel-equivalent runtime settings. They MUST NOT override an explicit Cartulary configuration value. They MUST NOT permit unsupported exporters, unsupported protocols, remote-context acceptance, Baggage correlation, Prometheus scrape exposure, Zipkin export, Jaeger native export, vendor-native export, SQL commenter propagation, or browser export.

### 6.3 Configuration validation

**OTEL-REQ-018**
Invalid telemetry configuration MUST fail deployment-configuration validation before readiness. Exporter endpoint network unavailability MUST NOT fail startup when the endpoint is syntactically valid.

**OTEL-REQ-019**
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
| Per-signal endpoints | Any per-signal endpoint key appears in the current revision. |

## 7. Resource attributes

OpenTelemetry resource conventions define service identity attributes and caution that some potential instance identifiers can expose confidential infrastructure information.[^9]

**OTEL-REQ-020**
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

**OTEL-REQ-021**
`service.instance.id` MUST NOT be a hostname, pod name, container ID, IP address, MAC address, user identifier, incident identifier, customer name, filesystem path, object-store key, deployment root, or secret reference. The default generated value MUST be a canonical lowercase UUID v4 generated per process start.

## 8. Attribute governance

### 8.1 Attribute classes

**OTEL-REQ-022**
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
| Secret-bearing values | MUST NOT be emitted. |
| Incident-authored content | MUST NOT be emitted. |
| Baggage keys and values | MUST NOT be emitted, propagated, or used as Cartulary correlation attributes. |

### 8.2 Reserved namespaces

**OTEL-REQ-023**
Cartulary MAY emit standard attributes from reserved namespaces only when a signal table explicitly allows them. Cartulary MUST NOT create custom attributes inside these namespaces:

| Reserved namespace |
| --- |
| `http.` |
| `url.` |
| `db.` |
| `server.` |
| `client.` |
| `network.` |
| `service.` |
| `deployment.` |
| `telemetry.sdk.` |
| `otel.` |
| `error.` |
| `exception.` |
| `aws.` |
| `rpc.` |
| `log.` |
| `event.` |

### 8.3 Forbidden telemetry values and attributes

**OTEL-REQ-024**
The implementation MUST NOT export any forbidden value from this section as a span name, metric name, metric attribute, span attribute, span event name, span event attribute, log field, log body, log attribute, resource attribute, exemplar attribute, exception field, exporter diagnostic, retained telemetry test artifact, or telemetry self-diagnostic.

| Forbidden value family | Examples |
| --- | --- |
| Authentication secrets | Passwords, password hashes, TOTP secrets, TOTP codes, bootstrap tokens, session tokens, cookies, bearer tokens, CSRF tokens. |
| Provider secrets and assertions | OIDC tokens, SAML responses, provider assertions, provider access tokens, ID tokens. |
| Incident-authored text | Timeline summary, timeline details, source text, notes, findings, query text, communication summaries, handoff text, lesson text, analyst-authored search text. |
| Evidence-sensitive material | Blob bytes, object-store keys, bucket names, filenames, path hints, file hashes, MIME-derived filenames, preview handles, download handles, upload IDs, storage refs. |
| Stable user or party identifiers | `user_id`, `party_id`, email address, display name, external subject, auth-provider subject. |
| Stable incident or record identifiers | `incident_id`, `record_id`, `row_version`, `client_txn_id`, `entity_mention_id`, `object_blob_id`, `evidence_record_id`, `job_id`, `saved_view_id`, handle token. |
| Raw query structure | Filter values, search strings, SQL text, SQL bind values, saved-view query JSON, workbook pasted content, import-source headers. |
| Infrastructure secrets | Database DSNs, object-store credentials, OTLP headers, TLS private keys, config secret refs after resolution. |

**OTEL-REQ-025**
The following standard attributes and equivalent custom replacements are forbidden in the base telemetry profile:

| Forbidden attribute or family | Required reason |
| --- | --- |
| `url.full` | May include raw paths, queries, incident IDs, record IDs, saved-view IDs, and search text. |
| `url.path` | Forbidden when it contains concrete route parameters. Route templates MUST use `http.route`. |
| `url.query` | Always forbidden. |
| `http.request.header.*` | Forbidden by default because headers may contain session, CSRF, bearer, or customer-sensitive data. |
| `http.response.header.*` | Forbidden by default unless a later revision explicitly allowlists individual headers. |
| `db.query.text` | Forbidden even when parameterized or sanitized. |
| `db.query.parameter.*` | Always forbidden. |
| `aws.s3.*` | Forbidden in the base telemetry profile. |
| `exception.message` | Forbidden unless a signal-specific table explicitly allowlists a redacted value. No such allowlist exists in this revision. |
| `exception.stacktrace` | Forbidden by default. |
| Baggage keys and values | Forbidden as a Cartulary incident, user, party, evidence, saved-view, job, or record correlation channel. |

### 8.4 Allowed custom attribute registry

**OTEL-REQ-026**
The implementation MUST emit Cartulary custom attributes only from this registry:

| Attribute | Type | Allowed values or bounds | Signals | Notes |
| --- | --- | --- | --- | --- |
| `cartulary.module` | string | `auth`, `incidents`, `timeline`, `entities`, `evidence`, `imports`, `links`, `revisions`, `projections`, `reference_data`, `reporting`, `collaboration`, `jobs`, `httpapi`, `postgres`, `objectstore`, `config`, `telemetry` | traces, metrics, logs | Closed module vocabulary. |
| `cartulary.route_family` | string | `auth`, `users`, `incidents`, `memberships`, `view_schemas`, `saved_views`, `workbook_preferences`, `views_query`, `views_rows`, `records`, `entity_mentions`, `object_blobs`, `evidence_access`, `jobs`, `extensions`, `telemetry_client` | traces, metrics | Route family, not raw path. |
| `cartulary.view_schema_id` | string | Standardized current-profile `view_schema_id` values only | traces, metrics | No saved-view IDs. |
| `cartulary.record_type` | string | Current closed record-type vocabulary | traces, metrics | No record IDs. |
| `cartulary.operation` | string | Operation-specific closed value from the owning signal table | traces, metrics | Examples are signal-specific; raw SQL and object keys are forbidden. |
| `cartulary.result` | string | `success`, `error`, `conflict`, `canceled`, `timeout`, `dropped`, `rejected` | traces, metrics, logs | No raw error messages. |
| `cartulary.error_code` | string | Public error-code registry value or `internal_error` | traces, metrics, logs | No stack traces or exception messages. |
| `cartulary.job_kind` | string | `import_discovery`, `import_apply`, `reference_pack_verify`, `reference_pack_activate`, `snapshot_build`, `report_render`, `portability_export`, `portability_import`, `projection_rebuild`, `maintenance` | traces, metrics | Job ID forbidden. |
| `cartulary.job_terminal_status` | string | `succeeded`, `failed`, `canceled`, `expired` | traces, metrics | Terminal status only. |
| `cartulary.websocket.event_type` | string | `presence_snapshot`, `record_changed`, `job_progress`, `session_revoked`, `subscription_error` | traces, metrics | Payload content forbidden. |
| `cartulary.deployment.profile` | string | `disconnected`, `on_prem`, `cloud`, `test`, or implementation-declared profile token | resource, traces, metrics | Low-cardinality deployment descriptor. |
| `cartulary.incident.hash64` | string | 16 lowercase hex chars | traces, metrics | Emitted only when `incident_correlation='hmac_64bit'`. |
| `cartulary.telemetry.exporter_kind` | string | `none`, `otlp_http`, `otlp_grpc` | metrics, logs | Self-observability only. |
| `cartulary.signal_kind` | string | `traces`, `metrics`, `logs` | metrics, logs | Telemetry self-metrics. |
| `cartulary.drop_reason` | string | `queue_full`, `redaction_rejected`, `exporter_permanent_error`, `shutdown_timeout`, `recursion_guard` | metrics, logs | Telemetry self-metrics. |

**OTEL-REQ-027**
`cartulary.incident.hash64` MUST NOT be emitted unless all conditions in this table are true:

| Condition | Required behavior |
| --- | --- |
| Configuration | `telemetry.attribute.incident_correlation='hmac_64bit'`. |
| Secret | `telemetry.attribute.hmac_secret_ref` resolves to a server-side secret with at least `256 bits` of entropy. |
| Algorithm | Compute HMAC-SHA-256 over the canonical incident identifier bytes, take the first 64 bits, and encode as exactly 16 lowercase hex characters. |
| Scope | The value MAY appear only on workbook, record, job, and dependency telemetry where incident-level aggregation is operationally necessary. |
| Prohibition | The raw incident ID, incident key, title, customer name, or any reversible encoding MUST NOT appear. |

## 9. Tracing contract

### 9.1 General span rules

**OTEL-REQ-028**
Span names MUST use route templates, module operations, or stable operation names. Span names MUST NOT include path IDs, incident IDs, record IDs, user-supplied strings, search text, filenames, object keys, SQL text, visible row values, saved-view IDs, or handle tokens.

**OTEL-REQ-029**
Span status and `error.type` MUST follow the signal-specific error rules. `cartulary.error_code` carries Cartulary public error-code tokens. The implementation MUST NOT overload `error.type` with Cartulary public error codes unless the value is also the low-cardinality OTel error class for that span.

**OTEL-REQ-030**
Exception recording MUST NOT emit `exception.message` or `exception.stacktrace` in the base telemetry profile. `exception.type` MAY be emitted only when it is a low-cardinality class name that does not include incident, user, record, object, path, SQL, or secret material.

### 9.2 Instrumentation scopes

**OTEL-REQ-031**
Each tracer, meter, or logger MUST be created with one of the following instrumentation scopes. Scope version MUST be the Cartulary build version when known, else `0.0.0+unknown`. Schema URL is `null` in this revision because Cartulary custom telemetry does not define an OTel schema URL.

| Scope name | Owning instrumentation area | Allowed signals |
| --- | --- | --- |
| `cartulary.httpapi` | HTTP API and route-family classification | traces, metrics, logs |
| `cartulary.workbook` | Workbook query, row create, row mutation, conflict, and projection work | traces, metrics, logs |
| `cartulary.collaboration` | WebSocket subscription and server-to-client events | traces, metrics, logs |
| `cartulary.jobs` | Background-job enqueue and execution | traces, metrics, logs |
| `cartulary.postgres` | Postgres dependency spans and pool/dependency metrics | traces, metrics |
| `cartulary.objectstore` | S3-compatible object storage abstraction | traces, metrics |
| `cartulary.telemetry` | Telemetry self-metrics and bounded diagnostics | metrics, logs |

### 9.3 Required span families

**OTEL-REQ-032**
The implementation MUST provide the span families in this table when the corresponding operation executes and tracing is enabled:

| Span family | Span name | SpanKind | Instrumentation scope | Parent rule | Link rule | Required standard attributes | Required Cartulary attributes | Forbidden attributes | Start predicate | End predicate | Error rule |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| HTTP server | `{http.request.method} {http.route}` when route is known; otherwise `{http.request.method}` | `SERVER` | `cartulary.httpapi` | Root; inbound remote context is ignored in this revision. | Forbidden. | `http.request.method`; `http.route` when known; `http.response.status_code` when sent. | `cartulary.route_family`, `cartulary.result`, optional `cartulary.error_code`. | `url.full`, `url.query`, concrete path-derived name, request headers, response headers. | Request accepted by router. | Response body completion or route failure. | 4xx server responses are not errors solely because they are 4xx; 5xx sets error status unless a narrower owner rule says otherwise. |
| Workbook query | `cartulary.workbook.query` | `INTERNAL` | `cartulary.workbook` | Current HTTP server span or job span. | Forbidden. | None. | `cartulary.view_schema_id`, `cartulary.result`, optional `cartulary.incident.hash64`. | Raw filters, search text, saved-view ID, incident ID, record ID. | Query validation begins. | Response rows serialized or query rejected. | Validation rejection uses `cartulary.result='rejected'`; runtime failure sets error status and low-cardinality `error.type`. |
| Record mutation | `cartulary.record.mutate` | `INTERNAL` | `cartulary.workbook` | Current HTTP server span or job span. | Forbidden. | None. | `cartulary.view_schema_id`, `cartulary.record_type`, `cartulary.operation`, `cartulary.result`, optional `cartulary.error_code`, optional `cartulary.incident.hash64`. | Record ID, user ID, client transaction ID, submitted value, persisted value, conflict token. | Mutation validation begins. | Mutation commits, conflicts, rejects, fails, or is canceled. | Same-field conflict uses `cartulary.result='conflict'` and MUST NOT set OTel error status solely because conflict occurred. |
| Projection maintenance | `cartulary.projection.maintenance` | `INTERNAL` | `cartulary.workbook` | Current mutation span, job span, or root for maintenance rebuild. | Optional when triggered by batch or replay operation. | None. | `cartulary.record_type`, `cartulary.operation`, `cartulary.result`. | Projection table names, record IDs, incident IDs. | Projection unit begins. | Projection unit completes or fails. | Runtime failure sets error status. |
| WebSocket subscribe | `cartulary.websocket.subscribe` | `INTERNAL` | `cartulary.collaboration` | Current HTTP upgrade/request span when available; otherwise root. | Forbidden. | None. | `cartulary.result`, optional `cartulary.error_code`. | Connection ID, user ID, incident ID, record ID, field key. | Subscription authorization begins. | Subscription accepted, rejected, or closed during setup. | Auth or membership rejection uses `cartulary.result='rejected'`. |
| WebSocket send | `cartulary.websocket.send` | `INTERNAL` | `cartulary.collaboration` | Current operation span when send is synchronous; otherwise current job span or root. | Optional when caused by async mutation fan-out. | None. | `cartulary.websocket.event_type`, `cartulary.result`. | Payload content, connection ID, user ID, incident ID, record ID, changed field values. | Event serialization begins. | Event sent, dropped, or failed. | Drop due to disconnected client is not an error unless implementation error caused the drop. |
| Job enqueue | `cartulary.job.enqueue` | `PRODUCER` | `cartulary.jobs` | Current HTTP, mutation, or system-process span. | Forbidden. | None. | `cartulary.job_kind`, `cartulary.result`. | Job ID, incident ID, user ID, request payload. | Enqueue validation begins. | Job accepted, rejected, or fails to enqueue. | Enqueue rejection uses `cartulary.result='rejected'`; runtime failure sets error status. |
| Job run | `cartulary.job.run` | `CONSUMER` | `cartulary.jobs` | Root. | Required when job was enqueued from a traced request or traced mutation; optional for scheduler-only maintenance. | None. | `cartulary.job_kind`, `cartulary.job_terminal_status` when terminal, `cartulary.result`, optional `cartulary.incident.hash64`. | Job ID, user ID, record ID, incident ID unless hashed opt-in applies. | Job leaves queued state. | Job reaches terminal state. | Terminal failed job sets error status only when failure is an implementation or dependency error; ordinary cancellation uses `cartulary.result='canceled'`. |
| Object-store dependency | `cartulary.objectstore.operation` | `CLIENT` | `cartulary.objectstore` | Current HTTP, mutation, job, or evidence operation span. | Forbidden. | None. | `cartulary.module='objectstore'`, `cartulary.operation`, `cartulary.result`. | `aws.s3.*`, bucket, key, upload ID, filename, hash, evidence handle, storage ref. | Object operation begins. | Object operation completes or fails. | Dependency failure sets error status and low-cardinality `error.type`. |
| Postgres dependency | `postgresql {db.operation.name}` when operation name is known; otherwise `postgresql` | `CLIENT` | `cartulary.postgres` | Current HTTP, mutation, job, or projection span. | Forbidden. | `db.system.name='postgresql'`; `db.operation.name` when available; optional `db.query.summary` only from fixed low-cardinality labels. | `cartulary.module='postgres'`, `cartulary.result`. | `db.query.text`, `db.query.parameter.*`, SQL fragment, bind value, saved-view query JSON, search text. | Database operation begins. | Database operation completes or fails. | Dependency failure sets error status and low-cardinality database error class. |

**OTEL-REQ-033**
The implementation MUST NOT emit `cartulary.postgres.operation` as the only database dependency span family. Postgres dependency telemetry MUST use the database-client contract in §9.3. Cartulary internal workbook spans MAY remain parents of those dependency spans.

### 9.4 HTTP server span details

**OTEL-REQ-034**
HTTP server spans MUST use the Cartulary route template as `http.route`. They MUST NOT use concrete URL paths, raw route parameters, query strings, incident IDs, record IDs, saved-view IDs, object IDs, handle tokens, or search text in span names.

**OTEL-REQ-035**
HTTP server 4xx responses MUST NOT set span error status solely because the response is 4xx. HTTP server 5xx responses MUST set span error status unless a narrower owner rule proves the status code is not an implementation or dependency error. Intentional caller cancellation MUST NOT be classified as an error solely because the request ended early.[^10]

### 9.5 Database span details

**OTEL-REQ-036**
Postgres dependency telemetry MUST NOT emit raw SQL, sanitized SQL, parameterized SQL text, bind values, saved-view query JSON, filter payloads, source headers, or user-authored search text. `db.query.summary` MAY be emitted only when generated from a fixed low-cardinality operation label, not from SQL text.

**OTEL-REQ-037**
SQL commenter, database trace-context comments, or equivalent database-context propagation MUST be disabled in the current revision and MUST NOT be enabled by OTel environment variables.

### 9.6 Object-store non-adoption rule

**OTEL-REQ-038**
Cartulary MUST NOT emit `aws.s3.*` semantic-convention attributes in the base telemetry profile. Cartulary object storage is an S3-compatible abstraction over binary evidence storage, not evidence that AWS S3 development conventions are safe for incident telemetry.[^11]

**OTEL-REQ-039**
Object-store operation telemetry MUST be limited to operation, result, duration, and byte-count measurements. Bucket names, object keys, upload IDs, copy sources, filenames, evidence handles, preview handles, download handles, file hashes, storage refs, and evidence titles are forbidden.

### 9.7 Trace causality and links

**OTEL-REQ-040**
Trace topology MUST follow this table:

| Scenario | Required trace topology |
| --- | --- |
| HTTP request starts and completes synchronous workbook work | Child spans under the HTTP server span. |
| HTTP request enqueues a background job | `cartulary.job.enqueue` is a child of the request or mutation span; `cartulary.job.run` is a root or consumer span linked to the enqueueing span when it executes asynchronously. |
| Background job starts from scheduler maintenance without a causal request | Job run span is a root span with only allowed system-process attribution attributes. |
| WebSocket fan-out caused by a mutation | Do not create false synchronous parentage from the original HTTP mutation to every pushed event. Use links or compact send spans only when causality matters. |
| Batch or replay operation processes multiple causal inputs | Use span links rather than a false single parent. |

OpenTelemetry spans have a single parent and may carry links to causally related spans; links are the correct mechanism for batch or async causal relationships that do not have a single synchronous parent.[^12]

### 9.8 Inbound trace context and Baggage

**OTEL-REQ-041**
Inbound remote trace context is not trusted in this revision. Browser requests and API requests MUST start server-owned root traces. No configuration, environment variable, or request header may change that behavior in this revision.

**OTEL-REQ-042**
Baggage MUST NOT be accepted, propagated, synthesized, exported, or used as a Cartulary incident, user, party, evidence, saved-view, job, or record correlation channel.

## 10. Metrics contract

OpenTelemetry metric instruments are identified by name, kind, description, and unit, and views define how measurements are processed, aggregated, and exported.[^13]

### 10.1 Metric naming rules

**OTEL-REQ-043**
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

**OTEL-REQ-044**
The implementation MUST use these explicit histogram bucket boundaries unless a metric registry row declares a different closed bucket set:

| Histogram family | Unit | Explicit bucket boundaries |
| --- | --- | --- |
| Duration | `s` | `0.005`, `0.010`, `0.025`, `0.050`, `0.100`, `0.250`, `0.500`, `1`, `2.5`, `5`, `10` |
| Bytes | `By` | `1024`, `4096`, `16384`, `65536`, `262144`, `1048576`, `4194304`, `16777216`, `67108864`, `268435456` |
| Rows | `{row}` | `1`, `10`, `50`, `100`, `500`, `1000`, `5000`, `10000`, `20000` |

### 10.3 Required custom metric registry

**OTEL-REQ-045**
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
| `cartulary.telemetry.item.dropped` | Counter | `{item}` | Telemetry items dropped locally before successful export. | integer | Sum | Cumulative | `cartulary.signal_kind`, `cartulary.drop_reason` | `3 signal kinds * 5 drop reasons` | Counter restarts at zero on process restart. | Increment once per item dropped for queue overflow, redaction rejection, exporter permanent discard, shutdown timeout, or recursion guard. | N/A. |

**OTEL-REQ-046**
Standard HTTP, database, runtime, and SDK metrics MAY be emitted under their OpenTelemetry names only when their attributes pass §8 and the relevant stable semantic convention is allowed by §4.3. If a standard metric would include forbidden attributes or high-cardinality values by default, Cartulary MUST configure a view or equivalent filter to drop those attributes or MUST disable that standard metric.

## 11. Logs and correlation

OpenTelemetry logs define a data model with trace and span correlation, severity, body, resource, instrumentation scope, attributes, and optional event identity.[^14]

**OTEL-REQ-047**
Cartulary MUST use structured application logging as the primary local log substrate. OTel log export is a bridge. Enabling the bridge MUST NOT change local log content, product behavior, or log redaction outcomes.

### 11.1 Local structured-log fields

**OTEL-REQ-048**
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

**OTEL-REQ-049**
When `telemetry.logs.bridge_enabled=true`, exported OTel LogRecords MUST use this mapping:

| Local structured-log field | OTel LogRecord field | Required behavior |
| --- | --- | --- |
| active trace ID | `TraceId` | Present when the log is emitted under a valid active span context. |
| active span ID | `SpanId` | Present only when `TraceId` is present. |
| active trace flags | `TraceFlags` | Present when available from active context. |
| local severity enum | `SeverityNumber`, `SeverityText` | Deterministic mapping in §11.3. |
| redacted message | `Body` | MUST NOT contain forbidden values. |
| `cartulary.module` | `Attributes["cartulary.module"]` | Required. |
| `cartulary.result` | `Attributes["cartulary.result"]` | Required on terminal logs. |
| `cartulary.error_code` | `Attributes["cartulary.error_code"]` | Required on public error logs. |
| instrumentation identity | `InstrumentationScope` | Same scope discipline as traces and metrics. |
| resource identity | `Resource` | Same resource contract as §7. |
| event name | `EventName` | Forbidden in the base profile unless a later revision adds a closed event-name registry. |

### 11.3 Severity mapping

**OTEL-REQ-050**
Local severity MUST map to OTel log severity by this table:

| Local severity | `SeverityText` | `SeverityNumber` |
| --- | --- | --- |
| `trace` | `TRACE` | `1` |
| `debug` | `DEBUG` | `5` |
| `info` | `INFO` | `9` |
| `warn` | `WARN` | `13` |
| `error` | `ERROR` | `17` |
| `fatal` | `FATAL` | `21` |

**OTEL-REQ-051**
Logs MUST NOT include forbidden values from §8.3. When log bridge export is enabled, redaction MUST run before the log record reaches an OTel processor, exporter queue, retained exporter artifact, or diagnostic capture.

## 12. Privacy and security invariants

Core 04 requires authenticated mutation origin and audit fidelity, but telemetry does not replace that audit substrate. Core 04 also requires a project-local STRIDE threat model for the current architecture, deployment profiles, and high-risk workflows.[^15]

**OTEL-REQ-052**
Telemetry MUST be classified as operational engineering evidence only. It MUST NOT replace change sets, record revisions, administrative audit records, evidence custody events, snapshot manifests, release approvals, benchmark manifests, or claim-bearing measurement artifacts.

**OTEL-REQ-053**
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

**OTEL-REQ-054**
Telemetry self-diagnostics MUST be bounded. The implementation MUST NOT produce unbounded logs, metrics, spans, or retained artifacts in response to exporter failure, processor overflow, redaction failure, or recursion guard activation.

## 13. Exporter, processor, runtime, and shutdown behavior

### 13.1 Exporter contract

OTLP exporter configuration defines transport protocols, endpoints, per-signal path construction for HTTP, headers, compression, timeouts, and retry behavior for transient failures.[^16]

**OTEL-REQ-055**
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

**OTEL-REQ-056**
For `otlp_http`, the configured endpoint MUST be treated as a base endpoint. It MUST NOT already be a per-signal endpoint ending in `/v1/traces`, `/v1/metrics`, or `/v1/logs`. Per-signal endpoint configuration is unsupported in this revision.

### 13.2 Retry contract

**OTEL-REQ-057**
Transient exporter errors MAY be retried only while all of the following remain true:

| Guard | Required behavior |
| --- | --- |
| Retry enabled | `telemetry.exporter.retry.enabled=true`. |
| Elapsed bound | Elapsed retry time is less than or equal to `telemetry.exporter.retry.max_elapsed_ms`. |
| Shutdown state | Process shutdown has not begun. |
| Processor bound | Queue and batch limits remain enforced. |
| Product isolation | Retrying export MUST NOT block product HTTP responses, WebSocket event delivery, mutation commit, evidence access, or job terminal transitions. |

**OTEL-REQ-058**
Retry delay MUST use exponential intervals starting at `telemetry.exporter.retry.initial_interval_ms`, multiplied by `telemetry.exporter.retry.multiplier`, and capped at `telemetry.exporter.retry.max_interval_ms`. Jitter MAY vary each retry delay within `50%..150%` of the computed capped interval. Tests MUST assert the configured bounds rather than exact jitter values.

### 13.3 Processor overflow contract

**OTEL-REQ-059**
Processor behavior MUST follow this table:

| Event | Required behavior |
| --- | --- |
| Processor queue accepts item | Record normally. |
| Processor queue full | Drop the telemetry item being offered. Retain already queued items. Increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='queue_full'`. |
| Redaction rejects item | Drop the item before enqueue. Increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='redaction_rejected'`. |
| Exporter timeout | Mark export attempt as timed out. Product operation remains unchanged. |
| Exporter transient error | Retry within configured retry bound. |
| Exporter permanent error | Drop or retain only according to bounded processor behavior. Product operation remains unchanged. |
| Shutdown flush timeout | Stop waiting after `telemetry.shutdown.flush_timeout_ms`; process shutdown continues. |
| Self-diagnostic emission | Must not recursively create unbounded telemetry about telemetry. |

### 13.4 Error-handling matrix

OpenTelemetry's error-handling guidance treats telemetry as non-essential relative to application behavior and encourages self-diagnostics for relevant errors.[^17]

**OTEL-REQ-060**
Failure handling MUST follow this table:

| Failure point | Required behavior |
| --- | --- |
| Invalid Cartulary telemetry configuration | Fail before readiness. |
| OTel SDK invalid-use panic risk | Contain through bootstrap validation or test-only strict handling; production MUST NOT expose unhandled telemetry exceptions. |
| Exporter endpoint unreachable at startup | Application starts. |
| Runtime exporter failure | Product request, job, WebSocket, evidence access, and mutation continue. |
| Processor queue overflow | Drop according to declared policy and increment drop metric. |
| Redaction failure | Drop affected telemetry item. |
| Log bridge export failure | Local logs remain intact; exported log item may be dropped. |
| Shutdown flush timeout | Continue shutdown after timeout. |

### 13.5 Startup and shutdown contract

**OTEL-REQ-061**
Startup behavior MUST satisfy these rules:

| Rule | Required behavior |
| --- | --- |
| Telemetry disabled | If `telemetry.enabled=false`, startup installs no-op telemetry providers and MUST NOT create exporters. |
| Invalid config | Syntactically or semantically invalid telemetry config fails deployment configuration validation. |
| Endpoint unreachable | A syntactically valid but unreachable endpoint does not fail startup. |
| Initialization order | Telemetry initialization completes before HTTP listener, WebSocket listener, and background-job runner startup so startup diagnostics can be correlated after initialization. |
| SDK env defaults | SDK defaults cannot enable export outside the Cartulary effective configuration envelope. |

**OTEL-REQ-062**
Shutdown behavior MUST satisfy these rules:

| Rule | Required behavior |
| --- | --- |
| Flush request | On graceful shutdown, the implementation requests trace, metric, and enabled log bridge flush. |
| Timeout | Flush wait uses `telemetry.shutdown.flush_timeout_ms`. |
| Product shutdown | Flush timeout MUST NOT prevent process shutdown after the timeout expires. |
| Diagnostics | Shutdown records a bounded local diagnostic when telemetry flush times out. |
| Drop metric | Items abandoned because of flush timeout increment `cartulary.telemetry.item.dropped` with `cartulary.drop_reason='shutdown_timeout'` when self-diagnostics are enabled and the increment can occur without recursion. |

### 13.6 Telemetry-about-telemetry recursion guard

**OTEL-REQ-063**
The implementation MUST NOT emit OpenTelemetry spans for telemetry export attempts. Telemetry export attempts MAY emit only the self-metrics in §10.3 and bounded local diagnostics. Those self-metrics MUST NOT themselves create additional telemetry export spans, span events, or recursive self-metrics.

## 14. Browser telemetry boundary and non-transfer rules

### 14.1 Browser boundary

**OTEL-REQ-064**
Browser direct export is forbidden. Browser code MUST NOT contain direct OTLP, Prometheus, Zipkin, Jaeger native, vendor telemetry, or third-party telemetry endpoint configuration.

**OTEL-REQ-065**
The browser MAY create local performance marks for workbook controller diagnostics, but those marks MUST remain local unless a later adopted revision defines an app-mediated client telemetry route.

**OTEL-REQ-066**
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

**OTEL-REQ-067**
The following OTel concepts or ecosystem patterns MUST NOT be transferred into the Cartulary base telemetry profile:

| OTel concept or ecosystem pattern | Cartulary base-profile treatment |
| --- | --- |
| Browser direct OTLP export | Forbidden. |
| Browser vendor telemetry endpoint | Forbidden. |
| Browser-configured exporter headers | Forbidden. |
| Baggage for incident, user, party, evidence, saved-view, job, or record correlation | Forbidden. |
| OpenTelemetry Collector as required deployable | Forbidden as a conformance dependency. Optional external receiver only. |
| Prometheus scrape endpoint | Deferred. Not part of this revision. |
| SQL commenter or database context propagation | Forbidden by default and not configurable in this revision. |
| OTel SDK environment autoconfiguration | Not authoritative. MUST NOT override Cartulary configuration. |
| AWS S3 semantic-convention attributes | Not adopted in the base telemetry profile. |
| OTel metric-pipeline non-sanitization guidance as a privacy rule | Not adopted. Cartulary validates and redacts before recording because incident content and identifiers are forbidden. |

## 15. Verification and acceptance criteria

Each criterion below is binary. A conformant implementation must pass every criterion in this section that applies to the implemented telemetry signals.

### 15.1 Baseline and registry criteria

- **OTEL-AC-001:** With default telemetry configuration, the application starts without an OTLP endpoint, exports no network telemetry, and still serves HTTP, WebSocket, and background-job surfaces.
- **OTEL-AC-002:** Setting `telemetry.enabled=false` results in no registered exporter, no outbound telemetry network connection, and no emitted OpenTelemetry spans, metrics, or exported logs.
- **OTEL-AC-003:** The revised spec contains a closed OTel component-boundary table defining API, SDK, instrumentation library, instrumentation scope, exporter, processor, Collector, and semantic conventions.
- **OTEL-AC-004:** A registry test fails if any emitted standard attribute is not generated or imported from the pinned semantic-convention model source or explicitly allowlisted.
- **OTEL-AC-005:** A registry test fails if any custom attribute uses a reserved OTel namespace or is absent from the Cartulary custom attribute registry.
- **OTEL-AC-006:** A naming lint fails if any custom metric or attribute violates the adopted grammar, unit, namespace, or collision rules.

### 15.2 Configuration and exporter criteria

- **OTEL-AC-007:** Setting `telemetry.exporter.kind=otlp_http` without `telemetry.exporter.endpoint` fails deployment-configuration validation before readiness.
- **OTEL-AC-008:** A syntactically valid but unreachable OTLP endpoint does not fail startup and does not fail a representative HTTP request, workbook query, mutation, WebSocket subscription, or background job.
- **OTEL-AC-009:** With default telemetry configuration, no outbound OTLP, Prometheus, Zipkin, Jaeger native, vendor-native, or browser telemetry connection is attempted.
- **OTEL-AC-010:** Setting `OTEL_TRACES_EXPORTER=otlp`, `OTEL_METRICS_EXPORTER=otlp`, `OTEL_LOGS_EXPORTER=otlp`, `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318`, and `OTEL_CONFIG_FILE` does not enable export when Cartulary config sets `telemetry.exporter.kind='none'`.
- **OTEL-AC-011:** `telemetry.exporter.kind=otlp_http` emits OTLP over `http/protobuf` and constructs `/v1/traces`, `/v1/metrics`, and `/v1/logs` paths from the configured base endpoint.
- **OTEL-AC-012:** `telemetry.exporter.kind=otlp_grpc` emits OTLP over gRPC and does not use HTTP path construction.
- **OTEL-AC-013:** `http/json`, per-signal endpoints, and unsupported exporter protocols fail deployment-configuration validation unless a later revision defines them.
- **OTEL-AC-014:** Forcing exporter blockage fills the processor queue, applies `drop_new`, increments `cartulary.telemetry.item.dropped`, and does not fail a representative product operation.
- **OTEL-AC-015:** Exporter headers are accepted from server-side configuration and are redacted from diagnostics, logs, API responses, browser state, test artifacts, telemetry attributes, and retained telemetry artifacts.

### 15.3 HTTP and route criteria

- **OTEL-AC-016:** A request to a route containing incident and record identifiers emits an HTTP server span named from method plus route template, not concrete URL path or query.
- **OTEL-AC-017:** HTTP server spans emit `http.response.status_code` when known, do not emit `url.full` or `url.query`, and do not set error status for ordinary server-side 4xx responses solely because they are 4xx.
- **OTEL-AC-018:** A representative 5xx response sets span error status and a low-cardinality `error.type` without leaking exception message, stacktrace, incident text, IDs, query payload, headers, or route parameters.

### 15.4 Workbook, mutation, and conflict criteria

- **OTEL-AC-019:** A workbook query emits exactly one `cartulary.workbook.query` span and one `cartulary.workbook.query.duration` metric measurement with `cartulary.view_schema_id` and without `incident_id`, `record_id`, query text, filter values, or saved-view IDs.
- **OTEL-AC-020:** A row mutation emits exactly one `cartulary.record.mutate` span and one `cartulary.record.mutation.duration` metric measurement with route-family, view-schema, record-type, operation, and result attributes.
- **OTEL-AC-021:** A same-field conflict increments `cartulary.record.mutation.conflict` and does not emit the client-submitted value, persisted server value, record ID, user ID, or conflict token.

### 15.5 Database and object-store criteria

- **OTEL-AC-022:** A Postgres-backed workbook query emits database dependency telemetry with `SpanKind.CLIENT` and `db.system.name='postgresql'` when DB semantic conventions are enabled.
- **OTEL-AC-023:** Postgres telemetry emits no `db.query.text`, no `db.query.parameter.*`, no saved-view query JSON, no filter payload, no search text, no SQL fragment, and no bind value.
- **OTEL-AC-024:** Object-store upload, preview-source fetch, and download-source fetch emit no `aws.s3.*`, bucket name, object key, filename, file hash, evidence handle, preview handle, download handle, storage ref, upload ID, or evidence title.
- **OTEL-AC-025:** SQL commenter or equivalent database-context propagation is disabled by default and cannot be enabled by OTel environment variables in the current revision.

### 15.6 WebSocket, job, and trace-causality criteria

- **OTEL-AC-026:** A WebSocket subscription updates `cartulary.websocket.connection.active` on accept and close, including abnormal close.
- **OTEL-AC-027:** A background job emits `cartulary.job.enqueue` when enqueued, emits `cartulary.job.run` when executed, updates `cartulary.job.active`, records one terminal `cartulary.job.duration` measurement, and does not emit `job_id`.
- **OTEL-AC-028:** An HTTP action that queues a background job produces a job execution span linked to the enqueueing request span or enqueue span.
- **OTEL-AC-029:** A background job started by scheduler maintenance without a causal request produces a root job span without incident, record, job ID, user, or party identifiers.
- **OTEL-AC-030:** WebSocket fan-out telemetry does not create false synchronous parentage from the original HTTP mutation to every pushed event.

### 15.7 Metrics criteria

- **OTEL-AC-031:** Every custom metric has one registry entry defining exact instrument kind, unit, description, allowed attributes, aggregation, lifecycle behavior, cardinality bound, and reset behavior.
- **OTEL-AC-032:** Active WebSocket and active job metrics return to zero on normal close, abnormal close, terminal job completion, cancellation, and process restart according to the declared reset semantics.
- **OTEL-AC-033:** Histogram bucket-boundary tests prove workbook duration, DB duration, object-store duration, byte count, and row count measurements use the registered explicit buckets.

### 15.8 Logs criteria

- **OTEL-AC-034:** Local structured logs under an active span include `trace_id` and `span_id` when available.
- **OTEL-AC-035:** When the OTel log bridge is enabled, exported LogRecords map trace ID, span ID, trace flags, severity, body, resource, instrumentation scope, and attributes according to the log mapping table.
- **OTEL-AC-036:** OTel log export emits no forbidden values in body, attributes, event name, exception fields, resource attributes, or retained telemetry artifacts.

### 15.9 Privacy, security, and non-transfer criteria

- **OTEL-AC-037:** The forbidden-value test injects representative credentials, incident text, evidence filenames, object keys, SQL bind values, user IDs, record IDs, incident IDs, client transaction IDs, handle tokens, and exporter headers, then verifies none appear in exported spans, metrics, logs, self-diagnostics, or retained telemetry test artifacts.
- **OTEL-AC-038:** Shutdown attempts telemetry flush and exits after the configured flush timeout even when the exporter does not respond.
- **OTEL-AC-039:** Telemetry emitted during functional tests is classified as operational engineering evidence only and is not used as claim-bearing benchmark evidence unless Core 05 publication requirements are separately satisfied.
- **OTEL-AC-040:** Browser code contains no direct OTLP, vendor telemetry, or third-party telemetry export endpoint configuration.
- **OTEL-AC-041:** Baggage is not used to propagate incident, record, user, party, evidence, saved-view, job, handle, or customer identifiers.
- **OTEL-AC-042:** The Collector is not required for Cartulary telemetry conformance; a deployment with `telemetry.exporter.kind='none'` passes the default telemetry criteria without any Collector present.
- **OTEL-AC-043:** The project-local STRIDE threat model includes telemetry exporter credentials, telemetry retention, inbound trace context, SDK environment override risk, telemetry queue denial-of-service, and telemetry self-diagnostic recursion.
- **OTEL-AC-044:** Telemetry export failure, queue overflow, redaction rejection, and flush timeout do not alter product HTTP responses, product state mutations, background-job state, WebSocket event semantics, evidence access decisions, or shutdown completion.

## 16. Open decisions

**OTEL-REQ-068**
Open decisions MUST use this disposition table:

| ID | Decision | Disposition |
| --- | --- | --- |
| `OTEL-DQ-001` | Adopted spec path | Open. Choose final repository path and adoption process. Suggested path remains `docs/cartulary-opentelemetry-instrumentation-nlspec.md`. |
| `OTEL-DQ-002` | Exact Go OpenTelemetry SDK versions | Open. Pin exact module versions in repo-control files before implementation adoption. |
| `OTEL-DQ-003` | Semantic-convention update policy | Closed by §4.1 and §4.3. Semantic-convention updates are spec revisions when emitted telemetry changes; dependency-only updates are allowed only when emitted telemetry remains registry-equivalent. |
| `OTEL-DQ-004` | Browser telemetry | Deferred. Browser direct export remains forbidden; app-mediated telemetry requires a later route contract. |
| `OTEL-DQ-005` | Prometheus compatibility | Deferred. Prometheus scrape support is not in this revision and MUST NOT be implied by OTel exporter lists. |
| `OTEL-DQ-006` | Incident correlation | Narrowed. Default remains `none`; `hmac_64bit` requires production-specific approval through configuration, server-side secret, bounded attribute use, and the collision-risk posture in §8.4. |
| `OTEL-DQ-007` | Operator dashboards and alerts | Closed as separate SRE/runbook scope. This NLSpec owns emitted telemetry, not dashboard artifacts. |

## 17. Completion standard

This revision is complete only when all of the following are true:

- Every emitted telemetry name, attribute, metric, span, log mapping, exporter setting, and processor setting has one owner table in this NLSpec.
- Every configurable telemetry value has a default, allowed values, bounds, omitted behavior, explicit-`null` behavior, and startup-validation behavior.
- Every standard OTel semantic convention used by Cartulary is classified by stability and privacy safety.
- Every Cartulary custom key is listed in the custom registry and passes the naming policy.
- Every forbidden value family has tests covering traces, metrics, logs, events, exception fields, exporter artifacts, self-diagnostics, and retained test artifacts.
- Browser direct export, Collector requirement, Prometheus scrape export, SQL commenter propagation, Baggage correlation, AWS S3 attributes, OTel environment autoconfiguration, raw HTTP URLs, raw SQL, raw DB parameters, object keys, filenames, and evidence handles are explicitly forbidden or explicitly deferred.
- The acceptance criteria are binary and sufficient for an implementation agent to determine conformance without guessing.

## Sources

[^1]: `00_document_set_status_and_precedence.md`, Core 00 §1-§4, lines 5-20 and 22-66. Core 00 defines the current normative core, separates Core 05 from implementation conformance, and places future adopted Cartulary NLSpecs above the current core in precedence.

[^2]: `05_claim_publication_and_benchmark_reproducibility.md`, Core 05 §1-§3, lines 3-22 and 53-67. Core 05 separates implementation correctness from claim-bearing timed or fixture-sensitive benchmark publication.

[^3]: OpenTelemetry Specification overview, `https://opentelemetry.io/docs/specs/otel/overview/`, lines 154-166 as observed by web retrieval on 2026-05-19. The page defines the API/SDK split, states that instrumentation authors must not directly reference SDK packages, and states that semantic-convention YAML files are the source of truth for generated constants.

[^4]: OpenTelemetry Specification page, `https://opentelemetry.io/docs/specs/otel/`, lines 24 and 123 as observed by web retrieval on 2026-05-19. The page identifies the current observed specification as `OpenTelemetry Specification 1.56.0`.

[^5]: OpenTelemetry Semantic Conventions page, `https://opentelemetry.io/docs/specs/semconv/`, line 24 as observed by web retrieval on 2026-05-19. The page identifies the current observed semantic-convention baseline as `Semantic conventions 1.41.0`.

[^6]: OpenTelemetry Specification overview, `https://opentelemetry.io/docs/specs/otel/overview/`, lines 161-166 as observed by web retrieval on 2026-05-19. The page states that semantic conventions are in their own repository and that YAML files must be used as the source of truth for generated constants.

[^7]: `01_architecture_storage_and_view_contracts.md`, Core 01 §1-§2, lines 3-49. Core 01 defines the modular monolith, the single application deployable, Postgres, S3-compatible object storage, and logical internal module boundaries.

[^8]: `04_security_deployment_and_conformance.md`, Core 04 §6 and §12, lines 459-478 and 1618-1646. Core 04 owns runtime roots and the operator-facing deployment configuration surface.

[^9]: OpenTelemetry service resource attributes, `https://opentelemetry.io/docs/specs/semconv/registry/attributes/service/`, observed by web retrieval on 2026-05-19. Used for service resource attribute alignment and opaque service-instance identity constraints.

[^10]: OpenTelemetry HTTP span semantic conventions, `https://opentelemetry.io/docs/specs/semconv/http/http-spans/`, lines 356-422 as observed by web retrieval on 2026-05-19. The page defines stable HTTP span conventions, low-cardinality span naming with `http.route`, status handling, and cancellation guidance.

[^11]: OpenTelemetry AWS S3 semantic conventions, `https://opentelemetry.io/docs/specs/semconv/object-stores/s3/`, observed by web retrieval on 2026-05-19. Used as evidence that S3 conventions include object-identifying attributes and are not adopted by Cartulary's base profile.

[^12]: OpenTelemetry Specification overview and trace API, `https://opentelemetry.io/docs/specs/otel/overview/` and `https://opentelemetry.io/docs/specs/otel/trace/api/`, observed by web retrieval on 2026-05-19. Used for span parent/link topology and span-kind alignment.

[^13]: OpenTelemetry Specification overview and Metrics API, `https://opentelemetry.io/docs/specs/otel/overview/` and `https://opentelemetry.io/docs/specs/otel/metrics/api/`, observed by web retrieval on 2026-05-19. Used for metric instrument identity, instrument kinds, aggregation/view, and naming requirements.

[^14]: OpenTelemetry Logs Data Model, `https://opentelemetry.io/docs/specs/otel/logs/data-model/`, observed by web retrieval on 2026-05-19. Used for OTel LogRecord field mapping.

[^15]: `04_security_deployment_and_conformance.md`, Core 04 §3-§4.5, lines 360-419. Core 04 requires a project-local STRIDE threat model and defines relevant trust-boundary controls.

[^16]: OpenTelemetry Protocol Exporter specification, `https://opentelemetry.io/docs/specs/otel/protocol/exporter/`, observed by web retrieval on 2026-05-19. Used for OTLP protocol, endpoint, headers, timeout, compression, and retry assumptions.

[^17]: OpenTelemetry Error Handling guidance, `https://opentelemetry.io/docs/specs/otel/error-handling/`, observed by web retrieval on 2026-05-19. Used for the fail-open runtime telemetry-loss posture and bounded self-diagnostics.
