---
title: OpenTelemetry Implementation Goal Handoff
status: handoff
document_class: execution-handoff
handoff_type: codex-goal
created_at: 2026-06-07
source_plan: opentelemetry-implementation-plan.md
---

# OpenTelemetry Implementation Goal Handoff

## 1. Purpose

This handoff is for running `opentelemetry-implementation-plan.md` as a long-running Codex goal against a live Cartulary repository. It converts the plan into an execution contract, stop-condition set, evidence ledger, and goal prompt suitable for the 4,000-character goal-input constraint.

This handoff is not a telemetry behavior owner. Core 00 through Core 04 remain the current implementation-conformance corpus, Core 05 governs claim-bearing publication only, and the adopted OpenTelemetry NLSpec owns telemetry subsystem behavior only.[^1] The implementation plan owns implementation sequencing, repo-control materialization, canonical conformance-visible paths, harness wiring, fixture ownership, and acceptance mapping, but it must not replace the adopted telemetry NLSpec, Core 04 deployment-configuration mechanics, the harness NLSpec, repository layout owners, or Core 05 publication rules.[^2]

The handoff is complete only as a goal-run artifact. It is not complete as telemetry implementation evidence, product conformance evidence, benchmark evidence, or release evidence.

## 2. Document Contract and Normative Language

`document_class` is `execution-handoff`. This document is not a behavior NLSpec and is not a substitute for `docs/opentelemetry-instrumentation-nlspec.md`.

Inside this handoff, uppercase **MUST**, **MUST NOT**, and **MAY** are normative only for goal execution. A handoff-owned **MAY** is valid only when omission behavior appears in the same paragraph or table row; omission means the goal runner must continue under the controlling owner document and must not create a new handoff-owned behavior. Requirements in imported owner text retain their original owner and meaning.

A handoff requirement is valid only when it satisfies one of these rules:

- It imports an owner requirement by exact document section, `OTEL-REQ-*`, `OTEL-AC-*`, or `OIP-AC-*` identifier.
- It defines goal-run behavior local to this handoff, such as ledger fields, stop conditions, prompt text, or final reporting.

The handoff MUST NOT own telemetry behavior, deployment-configuration mechanics, product behavior, benchmark publication, harness mechanics, or repository layout outside paths made canonical by the implementation plan. When this handoff and an owner differ, the owner governs and the handoff must be revised before the goal continues.

| Contract family | Behavior owner | Handoff-owned local consequence | Forbidden handoff action | Drift handling |
| --- | --- | --- | --- | --- |
| Telemetry behavior | `docs/opentelemetry-instrumentation-nlspec.md`, especially `OTEL-REQ-001..120` and `OTEL-REQ-122..136` | Import telemetry behavior by requirement ID and stop on owner gaps. | Invent signal shapes, config values, resource attributes, retry behavior, or privacy rules. | Mark `owner_drift_conflict` in the Decision Reconciliation Registry or stop condition report. |
| Implementation sequencing | `docs/opentelemetry-implementation-plan.md`, phases 0-11 and `OIP-AC-*` | Require phase order, phase gates, canonical artifacts, and acceptance mapping during the goal run. | Treat sequencing prose as independent telemetry behavior authority. | Owner plan governs sequencing unless it conflicts with the adopted telemetry NLSpec. |
| Deployment configuration | Core 04 plus OTel NLSpec §6 and implementation-plan Phase 2 | Require that telemetry keys materialize through the Core 04 configuration owner model. | Define a second config source, overlay grammar, discovery rule, or validation envelope. | Stop if Core 04 and OTel config requirements conflict. |
| Harness mechanics | `docs/testing-harness-nlspec.md`; implementation-plan Phase 10 and `OIP-AC-014` | Require Make-owned target use, retained evidence recording, and harness summary inspection. | Redefine Make invocation, result roots, output classes, cleanup, scheduler, or summary schema mechanics. | Report the unresolved harness owner section and do not treat child commands as final evidence. |
| Conformance-visible paths | Implementation-plan Phase 1, Phase 10, and acceptance registry | Use canonical paths and reject optional aliases unless the plan is revised. | Move conformance-visible artifacts to alternate paths by handoff decision. | Stop on path conflict; do not create parallel owner artifacts. |
| Product/domain behavior | Core 00 through Core 04 and `docs/domain.md` vocabulary boundaries | Preserve product behavior and domain vocabulary while adding telemetry. | Treat telemetry as product state, case data, evidence, audit state, workflow state, or domain authority. | Stop if telemetry work requires product behavior changes outside telemetry ownership. |
| Benchmark publication | Core 05 | Prevent telemetry evidence from being represented as claim-bearing benchmark evidence. | Use telemetry captures as Core 05 publication evidence without Core 05 predicates. | Stop and report publication-boundary conflict. |
| Live-repo facts | Live repository inspection plus implementation-plan source-limit rules | Require inspection before asserting packages, targets, files, schemas, or conventions exist. | Treat uploaded-corpus assumptions as live facts. | Use permitted `TODO(repo-adoption)` only under §6; otherwise stop. |

## 3. Source Limits and Run Posture

The implementation plan was written from the uploaded specification corpus rather than live repository inspection. A goal run MUST inspect the live repository before asserting that any package, Make target, schema, file, directory, or convention already exists.[^3]

The current telemetry NLSpec is adopted and current. It governs only telemetry generation, configuration, export, log correlation, naming, attribute governance, resource identity, privacy boundaries, runtime telemetry behavior, and verification.[^4] It MUST NOT redefine product behavior owned by Core 00 through Core 04 or claim-bearing benchmark publication owned by Core 05.[^4]

The implementation MUST keep telemetry inside the existing modular monolith. It MUST NOT add a telemetry sidecar, required Collector deployable, monitoring-vendor dependency, Prometheus service, browser telemetry service, or additional Cartulary deployable.[^5]

## 4. Pasteable Pursue Goal Prompt

Use this when the handoff document itself is available in the repository.

```text
Execute docs/handoffs/opentelemetry-goal-handoff.md as the active Cartulary OpenTelemetry implementation handoff. Read opentelemetry-implementation-plan.md, opentelemetry-instrumentation-nlspec.md, testing-harness-nlspec.md, Core 00-05, docs/domain.md, and relevant repo-control files before editing. Preserve authority boundaries: the adopted OTel NLSpec owns telemetry behavior, Core 00-04 own product conformance, Core 05 owns publication, the harness NLSpec owns Make/harness mechanics, and the implementation plan owns sequencing, artifacts, phase gates, and acceptance mapping. Use the Decision Reconciliation Registry, not a blanket OTEL-DQ rule. Treat OTEL-DQ-007, OTEL-DQ-008, OTEL-DQ-009, OTEL-DQ-010, residual OTEL-DQ-011, and OTEL-DQ-012 as blockers unless the owner docs have been revised to close them. Do not invent owner values in the handoff or implementation. Inspect the live repo before asserting package, target, file, schema, or convention facts. Use TODO(repo-adoption): ... only under the handoff TODO semantics. Implement phases in dependency order unless an earlier phase is already fully satisfied. Keep telemetry inside the modular monolith; do not add a sidecar, required Collector, vendor dependency, browser exporter, product workflow, case-data authority, or default network export. Run the narrowest relevant Make target after each slice and record changed files, owner imports, evidence, failures, blockers, decision status, and next action. Final done is unavailable while any decision registry row is owner_blocking, partially_owner_resolved, or owner_drift_conflict. If owner specs conflict, stop and report the exact conflict; do not pick a side.
```

Use this shorter prompt when the handoff file is outside the repository and must be attached or pasted separately.

```text
Execute the attached OpenTelemetry goal handoff and opentelemetry-implementation-plan.md. Treat the adopted OTel NLSpec as telemetry behavior authority, Core 00-04 as product authority, Core 05 as publication-only authority, and the harness NLSpec as Make/harness authority. Use the handoff Decision Reconciliation Registry, not a blanket OTEL-DQ rule. Treat OTEL-DQ-007, OTEL-DQ-008, OTEL-DQ-009, OTEL-DQ-010, residual OTEL-DQ-011, and OTEL-DQ-012 as blockers unless owner docs close them. Do not invent owner values. Inspect live repo facts before claiming them. Final done is unavailable while any decision registry row is owner_blocking, partially_owner_resolved, or owner_drift_conflict, or while any blocking TODO(repo-adoption) remains.
```

## 5. Non-Negotiable Constraints

- The run MUST NOT treat telemetry output as product state, audit state, case data, history, projection state, evidence state, workflow state, or benchmark evidence.
- The run MUST NOT add user-facing row-edit rituals, approval gates, or capture friction.
- The run MUST NOT require a vendor backend, required Collector, public dashboard, browser-to-third-party export, raw incident-content logging, OpenTelemetry environment-driven egress, or Collector-side privacy enforcement.[^6]
- Ordinary instrumentation units MUST import only OpenTelemetry API packages and Cartulary telemetry bootstrap accessors. SDK, exporter, processor, sampler, metric-reader, log-processor, resource, declarative-config, autoconfiguration, and plugin-provider imports belong only in the telemetry bootstrap boundary or hostile-fixture tests.[^7]
- Browser packages MUST NOT include OpenTelemetry exporters, vendor telemetry SDKs, Collector clients, session replay SDKs, or third-party analytics initialization in the current profile.[^7]
- Partial implementation may exist only as implementation-in-progress. Omission behavior: before a phase gate passes, the goal run remains `adopted_incomplete`, network export remains unavailable except as harness-owned explicit test behavior, and no conformance claim or `otel_release_ready` state is permitted.[^8]

## 6. TODO(repo-adoption) and Omission Semantics

A `TODO(repo-adoption): <specific missing value>` is permitted only for one of these cases:

- a missing live-repo fact that the implementation plan permits as a repo-adoption placeholder; or
- an owner-blocked decision named in the Decision Reconciliation Registry.

Every `TODO(repo-adoption)` in a conformance-visible artifact is blocking unless the Decision Reconciliation Registry explicitly gives that row `handoff execution status='repo_materialized'` and the TODO has been removed from final evidence. Non-blocking TODOs may appear only in transient execution notes. They MUST NOT appear in conformance-visible artifacts, acceptance mappings, final evidence, release-readiness manifests, or final handoff status.

Missing implementation before phase completion means `adopted_incomplete`. It MUST NOT permit network export by default, conformance claims, or `otel_release_ready`. If a phase needs an owner value that is absent, the goal runner MUST stop at that phase and report the exact owner blocker. It MUST NOT choose a value, infer a default, or encode a provisional value as final conformance behavior.

## 7. Phase Execution Ledger

Use one ledger row per slice. Keep the current row short enough to paste into a status update.

| Field | Required content |
| --- | --- |
| `slice_id` | One of `OTEL-1` through `OTEL-10`, or `OTEL-11` for final readiness. |
| `phase_refs` | Implementation-plan phase numbers touched by the slice. |
| `owner_refs` | Source file, section, and `OTEL-REQ-*`, `OTEL-AC-*`, or `OIP-AC-*` identifiers inspected before editing. |
| `changed_paths` | Repo-relative regular files changed. Generated files must be identified as generated. |
| `tests_before` | First failing or missing tests observed before implementation. |
| `tests_after` | Narrowest successful or failing validation command after implementation. |
| `artifact_refs` | Retained run root, summary JSON, raw capture path, normalized golden path, or manifest path. |
| `blocking_todos` | Exact blocking `TODO(repo-adoption)` values remaining, or `none`. |
| `decision_status` | Decision Reconciliation Registry rows changed, blocked, or repo-materialized. |
| `owner_blockers` | Exact owner section or decision row preventing progress, or `none`. |
| `acceptance_ids` | `OIP-AC-*` and `OTEL-AC-*` rows exercised by the slice. |
| `next_action` | One concrete next action or `handoff complete`. |

## 8. Boundary Closure Matrix

| Boundary | Owner import | Local handoff closure | Divergence risk removed | Blocking condition | Acceptance evidence |
| --- | --- | --- | --- | --- | --- |
| Config defaults, bounds, omitted behavior, explicit null, empty env, secret status, and failure behavior | OTel NLSpec §6, `OTEL-REQ-024..030`, `OTEL-REQ-122..125`; implementation-plan Phase 2, `OIP-AC-005..006`; `OTEL-AC-003..006` | Materialize the full telemetry key table, parser families, cross-key matrix, and hostile environment/config fixtures before treating config as complete. | Different defaults, null handling, empty env handling, secret diagnostics, or startup failure behavior. | `OTEL-DQ-012` blocks header count, value-length, and total header-block bounds. | Generated schema/config tests plus `OIP-AC-005..006` and `OTEL-AC-003..006`. |
| TODO and omission semantics | Implementation-plan source limits and partial implementation sections; this handoff §6 | Permit placeholders only for live-repo facts or named decision blockers; partial work remains `adopted_incomplete`. | Silent provisional behavior, accidental conformance claims, or export before phase completion. | Any blocking TODO in conformance-visible evidence. | Ledger `blocking_todos`, conformance-status manifest, `OIP-AC-001..003`. |
| Resource identity and null omission | OTel NLSpec §7 and §8.2, `OTEL-REQ-031..038`, `OTEL-REQ-126..128`; implementation-plan Phase 4; `OTEL-AC-010..014A` | Require closed resource attributes, empty Resource `schema_url`, detector suppression, forbidden-value action tests, and no null attribute setter calls. | Divergent resource attributes, schema URL merge behavior, detector leakage, or null emission. | `OTEL-DQ-010` and residual `OTEL-DQ-011`. | Resource, detector, forbidden-value, SDK-limit, and null-omission tests; `OIP-AC-017`, `OIP-AC-027`. |
| Span lifecycle and parent/link/status behavior | OTel NLSpec §9, `OTEL-REQ-051..069`; implementation-plan Phase 5, `OIP-AC-009`; `OTEL-AC-015..020` | Require registered span names, lifecycle boundaries, local parent rules, disabled custom span events, and inbound context stripping. | Different span topology, status mapping, forbidden attributes, or remote-context trust. | Owner conflict in span registry or sampler profile. | Span-shape, sampler, deterministic trace-ID, and remote-context tests. |
| Metric identity, temporality, aggregation, overflow, and histogram buckets | OTel NLSpec §10, `OTEL-REQ-070..085`, `OTEL-REQ-131..134`; implementation-plan Phase 6, `OIP-AC-010`; `OTEL-AC-021..026A` | Require emitted metrics to match registry identity and cumulative temporality; final metric conformance cannot pass without row-count histogram bucket closure. | Divergent metric streams, View names, attribute filters, exemplars, overflow handling, or histogram buckets. | `OTEL-DQ-007` for `{row}` bucket sequence. | Metric registry, View, exemplar, overflow, temporality, and Bind tests. |
| Log bridge default, body bound, severity, EventName omission, and exception reduction | OTel NLSpec §11, `OTEL-REQ-086..094`; implementation-plan Phase 7, `OIP-AC-011`; `OTEL-AC-027..031` | Keep OTel log export disabled by default; when enabled, require exact top-level mapping, severity table, body bound behavior, no EventName, and safe exception reduction. | Divergent log export defaults, arbitrary severity text, unsafe exception fields, or span-event bridging. | Missing log bridge mapping tests. | Disabled bridge, enabled mapping, exception, EventName, and span-event tests. |
| Exporter endpoint, headers, User-Agent, retry, queue overflow, shutdown, and runtime invariance | OTel NLSpec §12, `OTEL-REQ-095..112`, `OTEL-REQ-135..136`; implementation-plan Phase 8, `OIP-AC-012`, `OIP-AC-021`, `OIP-AC-026`; `OTEL-AC-032..039A` | Require no default endpoint, deterministic OTLP/HTTP paths, one gRPC target, redacted headers, bounded retry envelope, `drop_new`, idempotent shutdown, and product-runtime invariance. | Different egress defaults, per-signal divergence, retry timing, header leakage, queue blocking, or product behavior changes. | `OTEL-DQ-009` blocks gRPC transport-security evidence; `OTEL-DQ-012` blocks header-bound evidence. | Endpoint, gRPC, header, User-Agent, retry, overflow, shutdown, recursion, and runtime-invariance tests. |
| Browser non-export and browser configuration absence | OTel NLSpec §13, `OTEL-REQ-113..116`; implementation-plan Phase 9, `OIP-AC-013`, `OIP-AC-022`; `OTEL-AC-040..042` | Require source, package graph, bundle, dynamic import, source-map, and runtime probes to fail closed when required artifacts cannot be inspected. | Browser exporter, vendor SDK, analytics path, or browser state configuring telemetry. | Required browser artifact cannot be inspected. | Browser bundle, config, and non-transfer tests. |
| Golden raw/normalized artifact separation and canonical normalization | OTel NLSpec §14, `OTEL-REQ-117..120`; implementation-plan Phase 10, `OIP-AC-014`, `OIP-AC-023`; `OTEL-AC-043..046` | Keep raw captures under run root and committed normalized goldens under `internal/testutil/golden/otel`; normalize only owner-permitted volatile fields. | Raw capture commits, unstable goldens, or normalization hiding shape facts. | `OTEL-DQ-008` blocks byte-stable version normalization. | Golden corpus, raw artifact, normalization, and dependency-update classification evidence. |
| Acceptance mapping and conformance-mode reporting | Implementation-plan Phase 0, Phase 11, `OIP-AC-002..003`, `OIP-AC-015..023` | Require exactly one conformance mode and require every phase gate to map to local or imported acceptance IDs. | Ambiguous readiness, skipped imported criteria, or inconsistent startup/harness/release status. | Any decision row remains `owner_blocking`, `partially_owner_resolved`, or `owner_drift_conflict`. | Conformance-status manifest, harness summary, acceptance mapping table, and final checklist. |

## 9. Phase Order and Completion Gate

| Phase | Required owner imports | Required artifacts or checks | Explicit blockers | Completion evidence | Acceptance IDs |
| --- | --- | --- | --- | --- | --- |
| 0 | Implementation-plan Phase 0; OTel NLSpec §1 and §16; Core 00-05 authority owners | Adoption posture, conformance-status manifest, decision registry update, threat-model update if release-impacting telemetry boundaries are introduced. | Any Decision Reconciliation Registry row without an allowed final status blocks `adopted_conformant`. | Exactly one conformance mode and no unsupported conformance claim. | `OIP-AC-001..003`, `OIP-AC-024` |
| 1 | Implementation-plan Phase 1; OTel NLSpec §4, §5.2; `OTEL-AC-001..002`, `OTEL-AC-007..009` | `contracts/otel/otel_source_snapshot.json`, `contracts/otel/generated_constants_manifest.json`, `contracts/otel/import_boundary.json`, static import checks, and generated-constant drift checks. | Placeholder source values, short SHAs, missing package versions, or forbidden imports. | Source baseline and import-boundary evidence exists or reports a blocking owner TODO under §6. | `OIP-AC-004`, `OIP-AC-016`, `OTEL-AC-001..002`, `OTEL-AC-007..009` |
| 2 | OTel NLSpec §6, `OTEL-REQ-024..030`, `OTEL-REQ-122..125`; implementation-plan Phase 2 | Full telemetry key table, parser fixtures for omitted, empty, valid, invalid, and explicit string `"null"` cases, cross-key matrix, secret redaction, and hostile OTel environment/config fixtures. | `OTEL-DQ-012` blocks final header count and size-bound evidence. | Generated schema and config tests cover every key default, bound, omission rule, explicit-null rule, empty-env rule, secret status, failure behavior, and hazard family. | `OIP-AC-005..006`, `OTEL-AC-003..006` |
| 3 | OTel NLSpec §4.3, §5.2, §5.3, §8.2; implementation-plan Phase 3 | API-facing accessors, registered scope validation, no-SDK lifecycle tests, and OTel API spy or emitted-capture null tests. | Unknown scope, SDK import in ordinary module, or unsafe no-SDK path. | Ordinary modules import only API packages/accessors, no-SDK product paths complete, and no null setter call occurs. | `OIP-AC-007`, `OIP-AC-016`, `OTEL-AC-007..009`, `OTEL-AC-014A` |
| 4 | OTel NLSpec §7 and §8, `OTEL-REQ-031..050`, `OTEL-REQ-126..128`, `OTEL-REQ-141..143`; implementation-plan Phase 4 | Closed resource registry, empty Resource `schema_url`, detector suppression, null omission, forbidden-value action tests, error-class registry, and redaction-before-recording corpus. | `OTEL-DQ-010` blocks final instance-id opacity; residual `OTEL-DQ-011` blocks final profile-claims collation/empty-set evidence. | Resource, detector, forbidden-value, unknown-attribute, SDK-limit, null-omission, and error-class mapping tests pass. | `OIP-AC-008`, `OIP-AC-017`, `OIP-AC-027`, `OTEL-AC-010..014A` |
| 5 | OTel NLSpec §9, `OTEL-REQ-051..069`, `OTEL-REQ-129..130`; implementation-plan Phase 5 | Sampler profile tests, fixed trace-ID corpus, inbound context stripping, span registry tests, lifecycle boundary tests, parent/link/status tests, and forbidden-attribute tests. | Sampler owner conflict or missing span registry evidence. | Every emitted span matches registered name, kind, boundary, parent/link rule, status rule, and allowed/forbidden attributes. | `OIP-AC-004`, `OIP-AC-009`, `OIP-AC-018`, `OTEL-AC-015..020` |
| 6 | OTel NLSpec §10, `OTEL-REQ-070..085`, `OTEL-REQ-131..134`; implementation-plan Phase 6 | Metric registry, identity validation, cumulative temporality, View filters, exemplar disablement, overflow handling, and Bind bypass tests. | `OTEL-DQ-007` blocks final metric conformance for `cartulary.workbook.rows.returned`. | Every emitted metric matches registered stream identity and no unregistered stream, exemplar, overflow, or bypass appears outside the negative fixture. | `OIP-AC-010`, `OIP-AC-019`, `OTEL-AC-021..026A` |
| 7 | OTel NLSpec §11, `OTEL-REQ-086..094`; implementation-plan Phase 7 | Disabled bridge tests, enabled LogRecord mapping tests, severity mapping, body truncation including `body_max_chars=0`, exception reduction, EventName omission, and no span-event bridge. | Missing disabled-default or exception-sanitization evidence. | Logs export only under enabled bridge and only in owner-approved shape. | `OIP-AC-011`, `OIP-AC-020`, `OTEL-AC-027..031` |
| 8 | OTel NLSpec §12, `OTEL-REQ-095..112`, `OTEL-REQ-135..136`; implementation-plan Phase 8 | OTLP/HTTP path tests, one-target gRPC tests, header redaction, User-Agent grammar, retry envelope, retry classification, queue overflow, shutdown, recursion guard, and runtime-invariance tests. | `OTEL-DQ-009` blocks final gRPC security evidence; `OTEL-DQ-012` blocks final header-bound evidence. | Endpoint, header, User-Agent, retry, overflow, shutdown, recursion, and product-invariance evidence pass or report owner blockers. | `OIP-AC-012`, `OIP-AC-021`, `OIP-AC-026`, `OTEL-AC-032..039A` |
| 9 | OTel NLSpec §13, `OTEL-REQ-113..116`; implementation-plan Phase 9 | Source import, package graph, built bundle, dynamic import, source-map, runtime-probe, and non-transfer checks. | Required browser artifact cannot be inspected, or forbidden runtime package is browser-reachable. | Browser direct export and browser configuration authority are absent. | `OIP-AC-013`, `OIP-AC-022`, `OTEL-AC-040..042` |
| 10 | OTel NLSpec §14, `OTEL-REQ-117..120`; implementation-plan Phase 10 | `make otel-conformance`, summary schema, retained raw artifacts under run root, committed normalized corpus under `internal/testutil/golden/otel`, canonical number rules, fixture IDs `OTEL-CORPUS-001..018`, and dependency-update classification. | `OTEL-DQ-008` blocks byte-stable version normalization; raw captures under committed golden paths block artifact gate. | Corpus covers every fixture ID, raw and normalized artifacts are separated, and dependency updates classify shape changes. | `OIP-AC-014`, `OIP-AC-023`, `OTEL-AC-043..046` |
| 11 | Implementation-plan Phase 11 and acceptance registry | Release-readiness state, final conformance-status manifest, acceptance mapping, final ledger, and retained evidence summary. | Any decision row remains `owner_blocking`, `partially_owner_resolved`, or `owner_drift_conflict`; any blocking TODO remains. | `otel_release_ready` only when `make otel-conformance` passes and every mapped criterion passes. | `OIP-AC-002..003`, `OIP-AC-015..023`, all imported `OTEL-AC-*` mapped by the plan |

The plan's recommended work breakdown maps these phases into `OTEL-1` through `OTEL-10`, starting with source snapshots and generated constants and ending with golden corpus and harness integration.[^9]

## 10. Canonical Artifacts to Inspect or Create

| Artifact family | Canonical path or boundary | Run rule |
| --- | --- | --- |
| OTel source snapshot | `contracts/otel/otel_source_snapshot.json` | Must reproduce the adopted baseline and reject `main`, short SHAs, placeholder digests, missing source paths, missing document status, and missing SDK package versions. |
| Generated constants manifest | `contracts/otel/generated_constants_manifest.json` | Must bind standard attributes and standard metrics to generated pinned sources or explicit allowlists. |
| Import-boundary manifest | `contracts/otel/import_boundary.json` | Must encode allowed API packages, bootstrap package roots, and forbidden package families. |
| Error-class registry | `contracts/otel/error_class_registry.json` | Must bind low-cardinality error classes and public error codes. |
| Golden corpus | `internal/testutil/golden/otel/**` | Must contain normalized committed corpus, not raw retained capture. |
| Conformance-status manifest | `contracts/otel/conformance_status.json` | Must declare exactly one closed conformance mode, read the Decision Reconciliation Registry state, and inventory blocking TODOs. |
| Telemetry bootstrap | `internal/platform/telemetry` | Only this boundary may import SDK, exporters, processors, samplers, readers, log processors, resource detectors, or equivalent bootstrap-only packages. |
| API-facing telemetry accessors | `internal/platform/telemetry/api` | Ordinary modules may import this package and OpenTelemetry API packages only. |

These paths are conformance-visible in the implementation plan and are not optional aliases unless the plan is revised.[^10]

## 11. Decision Reconciliation Registry

Every row in this registry has exactly one `handoff execution status`. Final done requires every row to have `handoff execution status` of `resolved_by_owner` or `repo_materialized`, and no row may remain `owner_blocking`, `partially_owner_resolved`, or `owner_drift_conflict`.

| Decision ID | Current owner status | Handoff execution status | Required value or blocker | Permitted provisional representation | Final-conformance rule |
| --- | --- | --- | --- | --- | --- |
| `OTEL-DQ-001` | Resolved by OTel NLSpec §16 and `OTEL-AC-001`; materialized through `contracts/otel/otel_source_snapshot.json`. | `resolved_by_owner` until repo-control artifact inspection; `repo_materialized` after artifact contains owner value. | Concrete semantic-convention model digest and digest algorithm from owner source snapshot requirements. | `TODO(repo-adoption): compute semconv_model_digest` only before source-baseline materialization. | Final conformance requires the owner digest in `contracts/otel/otel_source_snapshot.json`; placeholder or malformed digest fails. |
| `OTEL-DQ-002` | Resolved by OTel NLSpec §16, §4.1.3, and `OTEL-AC-001`. | `resolved_by_owner` until generated-constants artifact inspection; `repo_materialized` after artifact contains pinned provenance. | Exact generated-constant source, package, or generator version. | `TODO(repo-adoption): pin generated constants` only before generated-constant manifest materialization. | Final conformance requires generated constants provenance in `contracts/otel/generated_constants_manifest.json` or owner-equivalent repo-control evidence. |
| `OTEL-DQ-003` | Resolved by OTel NLSpec §16, §4.1.4, and `OTEL-AC-001..002`. | `resolved_by_owner` until package manifests and source snapshot are inspected; `repo_materialized` after exact versions are pinned. | Exact OTel API, SDK, exporter, logs, metrics, trace, semantic-convention constants, bridge, and adapter package versions. | `TODO(repo-adoption): pin OTel package versions` only before repo-control package materialization. | Final conformance requires exhaustive package-family rows and exact versions; missing family or placeholder version fails. |
| `OTEL-DQ-004` | Resolved by OTel NLSpec §8.5.1 and `OTEL-REQ-141..143`. | `resolved_by_owner` until repo error registry inspection; `repo_materialized` after telemetry registry and public error registry are bound. | Mapping from `cartulary.error_code` to `cartulary.error_class`. | `TODO(repo-adoption): bind error registry` only before registry materialization. | Final conformance requires the telemetry error-class registry to match the adopted public error-code registry. |
| `OTEL-DQ-005` | Resolved by OTel NLSpec §14.1 and implementation-plan Phase 10. | `resolved_by_owner` and `repo_materialized` only when committed and retained paths match owner paths. | Committed normalized corpus under `internal/testutil/golden/otel`; raw captures under harness run root only. | No placeholder is permitted for final artifact paths. | Final conformance fails if raw captures are committed under golden paths or normalized goldens use a different path without owner revision. |
| `OTEL-DQ-006` | Resolved by OTel NLSpec §12.4 and implementation-plan Phase 0/8. | `resolved_by_owner`. | Runtime retry jitter uses bounds-only conformance; deterministic RNG hook is optional test support and not runtime configuration. | No TODO needed. | Final conformance requires retry tests to assert bounds, cutoff, disabled retry, permanent rejection, timeout, shutdown, and non-blocking behavior. |
| `OTEL-DQ-007` | Not closed by current OTel NLSpec; implementation-plan Phase 0 identifies owner gap. | `owner_blocking`. | Explicit `{row}` histogram bucket sequence for `cartulary.workbook.rows.returned`. | `TODO(repo-adoption): define {row} histogram buckets` in phase evidence only. | Final metric conformance is unavailable until owner defines bucket sequence; implementation MUST NOT invent boundaries. |
| `OTEL-DQ-008` | Not closed by current OTel NLSpec §14.2; implementation-plan Phase 10 identifies owner gap. | `owner_blocking`. | Normalization rules for `service.version` and instrumentation-scope `version`. | `SERVICE_VERSION` and `SCOPE_VERSION` placeholders in provisional tests only after reporting owner blocker. | Final byte-stable golden conformance is unavailable until owner normalization rules are defined. |
| `OTEL-DQ-009` | Not closed by current OTel NLSpec §12.2; implementation-plan Phase 8 identifies owner gap. | `owner_blocking`. | OTLP/gRPC endpoint-scheme to transport-security mapping. | `TODO(repo-adoption): define gRPC scheme to TLS mapping` in endpoint-gate evidence only. | Final gRPC endpoint evidence is unavailable until owner defines allowed schemes and secure/insecure mapping. |
| `OTEL-DQ-010` | Current OTel NLSpec requires opacity and maximum length but does not define the full structural predicate or explicit unenforced provenance invariant. | `owner_blocking`. | Closed `service.instance.id` structural opacity predicate plus marked unenforced provenance invariant. | `TODO(repo-adoption): define instance-id opacity predicate` in resource-gate evidence only. | Final resource conformance is unavailable until owner defines deterministic accept/reject predicate and invariant wording. |
| `OTEL-DQ-011` | Partially resolved: OTel NLSpec §8.5 says string and sorted comma-delimited claimed profile IDs; sort collation, duplicate handling, and empty-set representation are not fully closed. | `partially_owner_resolved`. | Exact `cartulary.profile.claims` representation, sort collation, duplicate handling, and empty-set representation. | `TODO(repo-adoption): pin profile.claims representation` for unresolved subparts only. | Final resource byte-equality is unavailable until every listed subpart is owner-defined or the owner text is revised to close them. |
| `OTEL-DQ-012` | Not closed by current OTel NLSpec §6.1; implementation-plan Phase 2 identifies owner gap. | `owner_blocking`. | Maximum header count, maximum header value length, and maximum total header-block size for `telemetry.exporter.headers`. | `TODO(repo-adoption): bound header count and size` in config-gate evidence only. | Final config and exporter-header conformance are unavailable until owner defines exact bounds and counting rules. |

## 12. Validation Commands and Evidence Rules

Use Make as the canonical command surface. Direct package scripts, raw Go, Vitest, Playwright, Biome, Vite, pnpm, or tool-specific commands are developer conveniences unless invoked by a public Make target.[^11]

The final target is `make otel-conformance`. During implementation, run the narrowest relevant Make target. A package-level child command may be recorded only when the public target does not yet exist; omission behavior is that child-command output remains provisional evidence and cannot satisfy final harness evidence.

Every retained harness evidence reference MUST include result root, run ID, run root, target name, summary artifact, failure class, failure reason, rerun command, and investigation command when the harness exposes them. The harness owns result-root, retained-artifact, target-selection, output-mode, exit-code, failure-classification, and cleanup mechanics.[^12]

`CARTULARY_OUTPUT_MODE=summary` is the default operator-friendly mode. Use `verbose` only for live investigation and `machine` only when a target explicitly accepts exactly-one-JSON output.

## 13. Stop Conditions

Stop the goal run and produce a handoff update instead of continuing when any of these conditions is true:

1. Two owner documents appear to conflict and the conflict affects implementation behavior, conformance status, emitted telemetry shape, or release-readiness state.
2. A live-repo fact is unavailable and §6 does not permit a `TODO(repo-adoption)` for it.
3. A phase needs a value owned by another artifact and the Decision Reconciliation Registry row remains `owner_blocking` or `partially_owner_resolved`.
4. The only way to proceed would add a telemetry sidecar, required Collector, vendor backend, browser exporter, product-visible workflow change, case-data authority, or default network export.
5. `make otel-conformance` or its precursor target cannot emit retained evidence because harness ownership, result-root behavior, or artifact schema is unresolved.
6. A test failure is caused by product behavior outside the telemetry subsystem and cannot be fixed without changing Core 00 through Core 04 behavior.
7. The live owner documents close a decision differently from this handoff's registry. In that case, mark the row `owner_drift_conflict`, stop, and revise the handoff before implementation continues.

## 14. Final Done Checklist

The goal run is complete only when all conditions below are true:

- Source snapshots, semantic-convention model digest, generated constants, and SDK package versions are pinned and source-baseline checks pass.
- Config parsing satisfies OTel NLSpec §6 and implementation-plan `OIP-AC-005..006`, including defaults, bounds, omission rules, explicit-null rules, empty environment values, secret status, cross-key rules, and hostile OTel environment/config fixtures.
- External OpenTelemetry environment variables and declarative configuration cannot affect behavior.
- Ordinary instrumentation is API-only and no-SDK safe.
- Resource identity, resource `schema_url`, attributes, forbidden values, null omission, and redaction-before-recording satisfy `OTEL-AC-010..014A`.
- Traces, metrics, and logs emit only owner-registered shapes.
- Exporters are explicit, bounded by owner-imported retry/processor/shutdown rules, non-blocking, and redacted.
- Browser direct export and browser configuration authority are absent.
- Raw telemetry captures and normalized goldens are separated according to OTel NLSpec §14.1.
- Dependency-update classification is enforced.
- `make otel-conformance` passes and emits retained evidence under the harness owner model.
- Every local `OIP-AC-*` and imported `OTEL-AC-*` criterion mapped by implementation-plan Phase 11 passes.
- Every row in the Decision Reconciliation Registry has `handoff execution status` of `resolved_by_owner` or `repo_materialized`.
- No row in the Decision Reconciliation Registry remains `owner_blocking`, `partially_owner_resolved`, or `owner_drift_conflict`.
- No blocking `TODO(repo-adoption)` appears in a conformance-visible artifact, registry, fixture, phase gate, acceptance mapping, release-readiness manifest, or final evidence.

## 15. Handoff Update Template

Use this template at interruption, blocker, or final handoff.

```text
OTel goal handoff status:
- Current readiness state:
- Last completed slice:
- Changed paths:
- Tests run:
- Retained evidence:
- Imported acceptance IDs:
- Open failures:
- Blocking TODO(repo-adoption):
- Decision registry status:
- Owner blockers:
- Owner drift conflicts:
- Conformance-visible artifacts:
- Exact next action:
```

## 16. Document Revision Acceptance Criteria

- **HANDOFF-AC-001:** The document states that it is an execution handoff and not a telemetry behavior owner.
- **HANDOFF-AC-002:** Every normative handoff requirement imports an owner section or ID, or defines local goal-run behavior.
- **HANDOFF-AC-003:** The decision registry reconciles `OTEL-DQ-001` through `OTEL-DQ-012` with current owner state and gives each row one execution status.
- **HANDOFF-AC-004:** No blanket decision-row completion language remains outside the Decision Reconciliation Registry and final checklist.
- **HANDOFF-AC-005:** Every phase gate maps to at least one `OIP-AC-*` or `OTEL-AC-*` criterion.
- **HANDOFF-AC-006:** Every default, bound, omission rule, null rule, and edge-case family named in the handoff has an owner import or explicit blocking condition.
- **HANDOFF-AC-007:** Every handoff-owned uppercase **MAY** has explicit omission behavior.
- **HANDOFF-AC-008:** The final done checklist is impossible to satisfy while any blocking owner decision or blocking `TODO(repo-adoption)` remains.
- **HANDOFF-AC-009:** The handoff contains no invented values for `{row}` buckets, version normalization, gRPC transport security, instance-id opacity, profile-claim collation, or header size bounds.
- **HANDOFF-AC-010:** The handoff update template includes decision status, owner blockers, imported acceptance IDs, retained evidence, and exact next action.

## Sources

[^1]: `docs/spec/00_document_set_status_and_precedence.md`, §§1-4.1; `docs/opentelemetry-instrumentation-nlspec.md`, §1, `OTEL-REQ-001..004`.
[^2]: `docs/opentelemetry-implementation-plan.md`, §Contract ownership and closure model.
[^3]: `docs/opentelemetry-implementation-plan.md`, §Source limits.
[^4]: `docs/opentelemetry-instrumentation-nlspec.md`, §1, `OTEL-REQ-001..004`.
[^5]: `docs/opentelemetry-implementation-plan.md`, §Status, authority, and conformance modes; `docs/spec/01_architecture_storage_and_view_contracts.md`, §1.
[^6]: `docs/opentelemetry-instrumentation-nlspec.md`, §3 and §13, `OTEL-REQ-113..116`.
[^7]: `docs/opentelemetry-instrumentation-nlspec.md`, §4.3, §5.3, and §13.1.
[^8]: `docs/opentelemetry-implementation-plan.md`, §Partial implementation and omission semantics.
[^9]: `docs/opentelemetry-implementation-plan.md`, §Recommended work breakdown.
[^10]: `docs/opentelemetry-implementation-plan.md`, Phase 1 and Phase 10.
[^11]: `docs/testing-harness-nlspec.md`, §1; `docs/opentelemetry-implementation-plan.md`, Phase 10.
[^12]: `docs/testing-harness-nlspec.md`, §§1-3.
