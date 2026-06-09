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

This handoff is for running `opentelemetry-implementation-plan.md` as a long-running Codex goal against a live Cartulary repository. It converts the plan into an execution contract, stop-condition set, evidence ledger, boundary-gap registry, and goal prompt suitable for the 4,000-character goal-input constraint.

This handoff is not a telemetry behavior owner. Core 00 through Core 04 remain the current implementation-conformance corpus, Core 05 governs claim-bearing publication only, and the adopted OpenTelemetry NLSpec owns telemetry subsystem behavior only.[^1] The implementation plan owns implementation sequencing, repo-control materialization, canonical conformance-visible paths, harness wiring, fixture ownership, and acceptance mapping, but it must not replace the adopted telemetry NLSpec, Core 04 deployment-configuration mechanics, the harness NLSpec, repository layout owners, or Core 05 publication rules.[^2]

The handoff is complete only as a goal-run artifact. It is not complete as telemetry implementation evidence, product conformance evidence, benchmark evidence, or release evidence.

## 2. Document Contract and Normative Language

`document_class` is `execution-handoff`. This document is not a behavior NLSpec and is not a substitute for `docs/opentelemetry-instrumentation-nlspec.md`.

Inside this handoff, uppercase **MUST**, **MUST NOT**, and **MAY** are normative only for goal execution. A handoff-owned **MAY** is valid only when omission behavior appears in the same paragraph or table row; omission means the goal runner must continue under the controlling owner document and must not create a new handoff-owned behavior. Requirements in imported owner text retain their original owner and meaning.

A handoff requirement is valid only when it satisfies one of these rules:

- It imports an owner requirement by exact document section, `OTEL-REQ-*`, `OTEL-AC-*`, or `OIP-AC-*` identifier.
- It defines goal-run behavior local to this handoff, such as ledger fields, stop conditions, prompt text, closure-state vocabulary, registry schema, or final reporting.

The handoff MUST NOT own telemetry behavior, deployment-configuration mechanics, product behavior, benchmark publication, harness mechanics, or repository layout outside paths made canonical by the implementation plan. When this handoff and an owner differ, the owner governs and the handoff must be revised before the goal continues.

| Contract family | Behavior owner | Handoff-owned local consequence | Forbidden handoff action | Drift handling |
| --- | --- | --- | --- | --- |
| Telemetry behavior | `docs/opentelemetry-instrumentation-nlspec.md`, especially `OTEL-REQ-001..120` and `OTEL-REQ-122..136` | Import telemetry behavior by requirement ID and stop on owner gaps. | Invent signal shapes, config values, resource attributes, retry behavior, or privacy rules. | Mark `owner_drift_conflict` in the Decision Reconciliation Registry or Boundary Gap Registry. |
| Implementation sequencing | `docs/opentelemetry-implementation-plan.md`, phases 0-11 and `OIP-AC-*` | Require phase order, phase gates, canonical artifacts, and acceptance mapping during the goal run. | Treat sequencing prose as independent telemetry behavior authority. | Owner plan governs sequencing unless it conflicts with the adopted telemetry NLSpec. |
| Deployment configuration | Core 04 plus OTel NLSpec §6 and implementation-plan Phase 2 | Require that telemetry keys materialize through the Core 04 configuration owner model. | Define a second config source, overlay grammar, discovery rule, or validation envelope. | Stop if Core 04 and OTel config requirements conflict. |
| Harness mechanics | `docs/testing-harness-nlspec.md`; implementation-plan Phase 10 and `OIP-AC-014` | Require Make-owned target use, retained evidence recording, and harness summary inspection. | Redefine Make invocation, result roots, output classes, cleanup, scheduler, or summary schema mechanics. | Report the unresolved harness owner section and do not treat child commands as final evidence. |
| Conformance-visible paths | Implementation-plan Phase 1, Phase 10, and acceptance registry | Use canonical paths and reject optional aliases unless the plan is revised. | Move conformance-visible artifacts to alternate paths by handoff decision. | Stop on path conflict; do not create parallel owner artifacts. |
| Product/domain behavior | Core 00 through Core 04 and `docs/domain.md` vocabulary boundaries | Preserve product behavior and domain vocabulary while adding telemetry. | Treat telemetry as product state, case data, evidence, audit state, workflow state, or domain authority. | Stop if telemetry work requires product behavior changes outside telemetry ownership. |
| Benchmark publication | Core 05 | Prevent telemetry evidence from being represented as claim-bearing benchmark evidence. | Use telemetry captures as Core 05 publication evidence without Core 05 predicates. | Stop and report publication-boundary conflict. |
| Live-repo facts | Live repository inspection plus implementation-plan source-limit rules | Require inspection before asserting packages, targets, files, schemas, or conventions exist. | Treat uploaded-corpus assumptions as live facts. | Use permitted `TODO(repo-adoption)` only under §7; otherwise stop. |

## 3. Closure-State Model

This handoff uses one closure-state vocabulary across the phase ledger, Decision Reconciliation Registry, Handoff Boundary Gap Registry, prompts, stop conditions, conformance-status material, and final readiness summaries. A goal runner MUST NOT introduce any other closure state.

| Closure state | Meaning | Final-allowed? | Required handling |
| --- | --- | ---: | --- |
| `owner_closed` | The controlling owner document defines the value, rule, interface, or predicate exactly enough for implementation and verification, and no additional repo artifact must be materialized before the goal can rely on it. | Yes | Import the owner section or requirement ID and verify the owner value without restating it as handoff-owned behavior. |
| `repo_materialization_required` | The owner requirement is closed, but the live repository fact, generated artifact, manifest, fixture, target, or schema has not yet been inspected or created. | No | Inspect or materialize the repo artifact; until then, use only the permitted `TODO(repo-adoption)` form in §7. |
| `repo_materialized` | A required live repository fact, generated artifact, manifest, fixture, target, or schema has been inspected or created and matches the owner requirement. | Yes | Record the path or evidence reference in the ledger and remove the corresponding blocking `TODO(repo-adoption)` from final evidence. |
| `owner_closure_required` | The controlling owner document does not yet define the value, rule, interface, or predicate needed for deterministic implementation. Partially specified owner text is in this state until every listed subpart is owner-defined. | No | Stop at the affected phase, report the exact owner gap, and do not choose, infer, normalize, default, or encode a provisional owner value as final behavior. |
| `handoff_local_closed` | The gap concerns only goal-run mechanics owned by this handoff, such as ledger fields, prompt content, stop conditions, status reporting, registry schema, acceptance mapping, or closure-state vocabulary. | Yes | Apply the handoff rule exactly and keep it out of telemetry behavior, deployment-configuration behavior, product behavior, and harness mechanics. |
| `owner_drift_conflict` | The live owner documents or repo-control artifacts close a value differently from this handoff's registry, or two owner documents appear to conflict. | No | Stop, report the exact conflict, and revise this handoff before implementation continues. |

Final handoff completion, `adopted_conformant`, `otel_release_ready`, release-readiness evidence, and any final conformance claim are unavailable while any Decision Reconciliation Registry row or Handoff Boundary Gap Registry row has `closure_state` of `repo_materialization_required`, `owner_closure_required`, or `owner_drift_conflict`.

This handoff MAY define the required shape of an owner-closure payload when the shape is needed to prevent implementation drift. Omission behavior: if the owner document has not supplied that payload, the goal runner MUST stop and report the missing payload; it MUST NOT fill the owner value in this handoff or in implementation code.

The words or phrases `appropriate`, `preferred`, `when known`, `according to selected SDK convention`, `equivalent`, `representative`, and `as needed` are invalid in a final closure rule unless the same sentence, paragraph, or table row supplies a binary predicate that makes the allowed latitude testable. Imported owner text that still contains one of those terms remains owner-governed; this handoff records the resulting closure requirement as a boundary gap rather than interpreting the term locally.[^13]

## 4. Source Limits and Run Posture

The implementation plan was written from the uploaded specification corpus rather than live repository inspection. A goal run MUST inspect the live repository before asserting that any package, Make target, schema, file, directory, or convention already exists.[^3]

The current telemetry NLSpec is adopted and current. It governs only telemetry generation, configuration, export, log correlation, naming, attribute governance, resource identity, privacy boundaries, runtime telemetry behavior, and verification.[^4] It MUST NOT redefine product behavior owned by Core 00 through Core 04 or claim-bearing benchmark publication owned by Core 05.[^4]

The implementation MUST keep telemetry inside the existing modular monolith. It MUST NOT add a telemetry sidecar, required Collector deployable, monitoring-vendor dependency, Prometheus service, browser telemetry service, or additional Cartulary deployable.[^5]

## 5. Pasteable Pursue Goal Prompt

Use this when the handoff document itself is available in the repository.

```text
Execute docs/handoffs/opentelemetry-goal-handoff.md as the active Cartulary OpenTelemetry implementation handoff. Read opentelemetry-implementation-plan.md, opentelemetry-instrumentation-nlspec.md, testing-harness-nlspec.md, Core 00-05, docs/domain.md, and relevant repo-control files before editing. Preserve authority boundaries: the adopted OTel NLSpec owns telemetry behavior, Core 00-04 own product conformance, Core 05 owns publication, the harness NLSpec owns Make/harness mechanics, and the implementation plan owns sequencing, artifacts, phase gates, and acceptance mapping. Use both the Decision Reconciliation Registry and the Handoff Boundary Gap Registry; do not use a blanket OTEL-DQ rule. Treat every Decision Registry or Boundary Gap Registry row with closure_state repo_materialization_required, owner_closure_required, or owner_drift_conflict as blocking; rows with owner_closed still require implementation evidence. Do not invent owner values in the handoff or implementation. Inspect the live repo before asserting package, target, file, schema, or convention facts. Use TODO(repo-adoption): ... only under §7 TODO semantics. Implement phases in dependency order unless an earlier phase is already fully satisfied. Keep telemetry inside the modular monolith; do not add a sidecar, required Collector, vendor dependency, browser exporter, product workflow, case-data authority, or default network export. Run the narrowest relevant Make target after each slice and record changed files, owner imports, evidence, failures, blockers, decision_status, gap_status, and next_action. Final done is unavailable while any Decision Registry or Boundary Gap Registry row has a blocking closure_state or while any blocking TODO(repo-adoption) remains. If owner specs conflict, stop and report the exact conflict; do not pick a side.
```

Use this shorter prompt when the handoff file is outside the repository and must be attached or pasted separately.

```text
Execute the attached OpenTelemetry goal handoff and opentelemetry-implementation-plan.md. Treat the adopted OTel NLSpec as telemetry behavior authority, Core 00-04 as product authority, Core 05 as publication-only authority, and the harness NLSpec as Make/harness authority. Use both the Decision Reconciliation Registry and Handoff Boundary Gap Registry; do not use a blanket OTEL-DQ rule. Treat any closure_state of repo_materialization_required, owner_closure_required, or owner_drift_conflict as blocking; owner_closed rows still require implementation evidence before final done. Do not invent owner values. Inspect live repo facts before claiming them. Final done is unavailable while any Decision Registry or Boundary Gap Registry row has a blocking closure_state or while any blocking TODO(repo-adoption) remains.
```

## 6. Non-Negotiable Constraints

- The run MUST NOT treat telemetry output as product state, audit state, case data, history, projection state, evidence state, workflow state, or benchmark evidence.
- The run MUST NOT add user-facing row-edit rituals, approval gates, or capture friction.
- The run MUST NOT require a vendor backend, required Collector, public dashboard, browser-to-third-party export, raw incident-content logging, OpenTelemetry environment-driven egress, or Collector-side privacy enforcement.[^6]
- Ordinary instrumentation units MUST import only OpenTelemetry API packages and Cartulary telemetry bootstrap accessors. SDK, exporter, processor, sampler, metric-reader, log-processor, resource, declarative-config, autoconfiguration, and plugin-provider imports belong only in the telemetry bootstrap boundary or hostile-fixture tests.[^7]
- Browser packages MUST NOT include OpenTelemetry exporters, vendor telemetry SDKs, Collector clients, session replay SDKs, or third-party analytics initialization in the current profile.[^7]
- Partial implementation may exist only as implementation-in-progress. Omission behavior: before a phase gate passes, the goal run remains `adopted_incomplete`, network export remains unavailable except as harness-owned explicit test behavior, and no conformance claim or `otel_release_ready` state is permitted.[^8]

## 7. TODO(repo-adoption) and Omission Semantics

A `TODO(repo-adoption): <specific missing value>` is permitted only for one of these cases:

- a missing live-repo fact that the implementation plan permits as a repo-adoption placeholder; or
- an owner-blocked decision named in the Decision Reconciliation Registry or Handoff Boundary Gap Registry.

Every `TODO(repo-adoption)` in a conformance-visible artifact is blocking unless the Decision Reconciliation Registry or Handoff Boundary Gap Registry explicitly gives the corresponding row `closure_state` of `repo_materialized`, `owner_closed`, or `handoff_local_closed`, and the TODO has been removed from final evidence. Non-blocking TODOs may appear only in transient execution notes. They MUST NOT appear in conformance-visible artifacts, acceptance mappings, final evidence, release-readiness manifests, prompts, or final handoff status.

Missing implementation before phase completion means `adopted_incomplete`. It MUST NOT permit network export by default, conformance claims, or `otel_release_ready`. If a phase needs an owner value that is absent, the goal runner MUST stop at that phase and report the exact owner blocker. It MUST NOT choose a value, infer a default, or encode a provisional value as final conformance behavior.

## 8. Phase Execution Ledger

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
| `decision_status` | Decision Reconciliation Registry rows changed, blocked, or set to a final-allowed closure state. |
| `gap_status` | Handoff Boundary Gap Registry rows changed, blocked, or set to a final-allowed closure state. |
| `owner_blockers` | Exact owner section, decision row, or gap row preventing progress, or `none`. |
| `acceptance_ids` | `OIP-AC-*` and `OTEL-AC-*` rows exercised by the slice. |
| `next_action` | One concrete next action or `handoff complete`. |

## 9. Boundary Closure Matrix

| Boundary | Owner import | Local handoff closure | Divergence risk removed | Blocking condition | Acceptance evidence |
| --- | --- | --- | --- | --- | --- |
| Config defaults, bounds, omitted behavior, explicit null, empty env, secret status, and failure behavior | OTel NLSpec §6, `OTEL-REQ-024..030`, `OTEL-REQ-122..125`, `OTEL-REQ-144..146`; Core 04 `secret_ref_v1`; implementation-plan Phase 2, `OIP-AC-005..006`; `OTEL-AC-003..006` | Materialize the full telemetry key table, parser families, cross-key matrix, hostile environment/config fixtures, and secret-reference diagnostics before treating config as complete. | Different defaults, null handling, empty env handling, secret diagnostics, raw header handling, or startup failure behavior. | Owner closure is imported from `OTEL-DQ-012`, `HANDOFF-GAP-001`, and `HANDOFF-GAP-011`; implementation evidence remains required. | Generated schema/config tests plus `OIP-AC-005..006` and `OTEL-AC-003..006`. |
| TODO and omission semantics | Implementation-plan source limits and partial implementation sections; this handoff §7 | Permit placeholders only for live-repo facts or named decision blockers; partial work remains `adopted_incomplete`. | Silent provisional behavior, accidental conformance claims, or export before phase completion. | Any blocking TODO in conformance-visible evidence. | Ledger `blocking_todos`, conformance-status manifest, `OIP-AC-001..003`. |
| Resource identity and null omission | OTel NLSpec §7 and §8.2, `OTEL-REQ-031..038`, `OTEL-REQ-126..128`; implementation-plan Phase 4; `OTEL-AC-010..014A` | Require closed resource attributes, empty Resource `schema_url`, detector suppression, forbidden-value action tests, UUID v4 instance IDs, profile-claims serialization, and no null attribute setter calls. | Divergent resource attributes, schema URL merge behavior, detector leakage, null emission, instance ID acceptance, or profile-claim byte shape. | Owner closure is imported from `OTEL-DQ-010`, `OTEL-DQ-011`, `HANDOFF-GAP-005`, and `HANDOFF-GAP-006`; implementation evidence remains required. | Resource, detector, forbidden-value, SDK-limit, and null-omission tests; `OIP-AC-017`, `OIP-AC-027`. |
| Span lifecycle and parent/link/status behavior | OTel NLSpec §9, `OTEL-REQ-051..069`; implementation-plan Phase 5, `OIP-AC-009`; `OTEL-AC-015..020` | Require registered span names, lifecycle boundaries, local parent rules, disabled custom span events, inbound context stripping, and owner-closed status mapping evidence. | Different span topology, status mapping, forbidden attributes, or remote-context trust. | Owner closure is imported from `HANDOFF-GAP-008`; implementation evidence or owner-conflict detection remains required. | Span-shape, sampler, deterministic trace-ID, status mapping, and remote-context tests. |
| Metric identity, temporality, aggregation, overflow, and histogram buckets | OTel NLSpec §10, `OTEL-REQ-070..085`, `OTEL-REQ-131..134`, `OTEL-REQ-147`; implementation-plan Phase 6, `OIP-AC-010`; `OTEL-AC-021..026A` | Require emitted metrics to match registry identity, cumulative temporality, and the owner-closed row-count histogram buckets. | Divergent metric streams, View names, attribute filters, exemplars, overflow handling, or histogram buckets. | Owner closure is imported from `OTEL-DQ-007` and `HANDOFF-GAP-002`; implementation evidence remains required. | Metric registry, View, exemplar, overflow, temporality, bucket, and Bind tests. |
| Log bridge default, body bound, severity, EventName omission, and exception reduction | OTel NLSpec §11, `OTEL-REQ-086..094`; implementation-plan Phase 7, `OIP-AC-011`; `OTEL-AC-027..031` | Keep OTel log export disabled by default; when enabled, require exact top-level mapping, severity table, body bound behavior, no EventName, and safe exception reduction. | Divergent log export defaults, arbitrary severity text, unsafe exception fields, or span-event bridging. | Missing log bridge mapping tests. | Disabled bridge, enabled mapping, exception, EventName, and span-event tests. |
| Exporter endpoint, headers, User-Agent, retry, queue overflow, shutdown, and runtime invariance | OTel NLSpec §12, `OTEL-REQ-095..112`, `OTEL-REQ-135..136`; implementation-plan Phase 8, `OIP-AC-012`, `OIP-AC-021`, `OIP-AC-026`; `OTEL-AC-032..039A` | Require no default endpoint, deterministic OTLP/HTTP paths, one gRPC target, redacted `secret_ref_v1` headers, exact User-Agent grammar, bounded retry envelope, `drop_new`, idempotent shutdown, and product-runtime invariance. | Different egress defaults, per-signal divergence, retry timing, header leakage, queue blocking, User-Agent shape, URL canonicalization, or product behavior changes. | Owner closure is imported from `OTEL-DQ-009`, `OTEL-DQ-012`, `HANDOFF-GAP-001`, `HANDOFF-GAP-003`, `HANDOFF-GAP-009`, `HANDOFF-GAP-010`, and `HANDOFF-GAP-011`; implementation evidence remains required. | Endpoint, gRPC, header, User-Agent, retry, overflow, shutdown, recursion, and runtime-invariance tests. |
| Browser non-export and browser configuration absence | OTel NLSpec §13, `OTEL-REQ-113..116`; implementation-plan Phase 9, `OIP-AC-013`, `OIP-AC-022`; `OTEL-AC-040..042` | Require source, package graph, bundle, dynamic import, source-map, and runtime probes to fail closed when required artifacts cannot be inspected. | Browser exporter, vendor SDK, analytics path, or browser state configuring telemetry. | Required browser artifact cannot be inspected. | Browser bundle, config, and non-transfer tests. |
| Golden raw/normalized artifact separation and canonical normalization | OTel NLSpec §14, `OTEL-REQ-117..120`; implementation-plan Phase 10, `OIP-AC-014`, `OIP-AC-023`; `OTEL-AC-043..046` | Keep raw captures under run root and committed normalized goldens under `internal/testutil/golden/otel`; normalize only owner-permitted volatile fields including version placeholders. | Raw capture commits, unstable goldens, or normalization hiding shape facts. | Owner closure is imported from `OTEL-DQ-008` and `HANDOFF-GAP-004`; implementation evidence remains required. | Golden corpus, raw artifact, normalization, and dependency-update classification evidence. |
| Acceptance mapping and conformance-mode reporting | Implementation-plan Phase 0, Phase 11, `OIP-AC-002..003`, `OIP-AC-015..023` | Require exactly one conformance mode and require every phase gate to map to local or imported acceptance IDs. | Ambiguous readiness, skipped imported criteria, or inconsistent startup/harness/release status. | Any Decision Reconciliation Registry or Handoff Boundary Gap Registry row has a blocking closure state. | Conformance-status manifest, harness summary, acceptance mapping table, and final checklist. |

## 10. Phase Order and Completion Gate

| Phase | Required owner imports | Required artifacts or checks | Explicit blockers | Completion evidence | Acceptance IDs |
| --- | --- | --- | --- | --- | --- |
| 0 | Implementation-plan Phase 0; OTel NLSpec §1 and §16; Core 00-05 authority owners | Adoption posture, conformance-status manifest, decision registry update, threat-model update if release-impacting telemetry boundaries are introduced. | Any Decision Reconciliation Registry or Handoff Boundary Gap Registry row without a final-allowed closure state blocks `adopted_conformant`. | Exactly one conformance mode and no unsupported conformance claim. | `OIP-AC-001..003`, `OIP-AC-024` |
| 1 | Implementation-plan Phase 1; OTel NLSpec §4, §5.2; `OTEL-AC-001..002`, `OTEL-AC-007..009` | `contracts/otel/otel_source_snapshot.json`, `contracts/otel/generated_constants_manifest.json`, `contracts/otel/import_boundary.json`, static import checks, and generated-constant drift checks. | Placeholder source values, short SHAs, missing package versions, or forbidden imports. | Source baseline and import-boundary evidence exists or reports a blocking owner TODO under §7. | `OIP-AC-004`, `OIP-AC-016`, `OTEL-AC-001..002`, `OTEL-AC-007..009` |
| 2 | OTel NLSpec §6, `OTEL-REQ-024..030`, `OTEL-REQ-122..125`, `OTEL-REQ-144..146`; Core 04 `secret_ref_v1`; implementation-plan Phase 2 | Full telemetry key table, parser fixtures for omitted, empty, valid, invalid, and explicit string `"null"` cases, cross-key matrix, secret redaction, header limits, secret-reference resolution, and hostile OTel environment/config fixtures. | Missing config/header/secret fixture or schema evidence. | Generated schema and config tests cover every key default, bound, omission rule, explicit-null rule, empty-env rule, secret status, failure behavior, and hazard family. | `OIP-AC-005..006`, `OTEL-AC-003..006` |
| 3 | OTel NLSpec §4.3, §5.2, §5.3, §8.2; implementation-plan Phase 3 | API-facing accessors, registered scope validation, no-SDK lifecycle tests, and OTel API spy or emitted-capture null tests. | Unknown scope, SDK import in ordinary module, or unsafe no-SDK path. | Ordinary modules import only API packages/accessors, no-SDK product paths complete, and no null setter call occurs. | `OIP-AC-007`, `OIP-AC-016`, `OTEL-AC-007..009`, `OTEL-AC-014A` |
| 4 | OTel NLSpec §7 and §8, `OTEL-REQ-031..050`, `OTEL-REQ-126..128`, `OTEL-REQ-141..143`; implementation-plan Phase 4 | Closed resource registry, empty Resource `schema_url`, detector suppression, null omission, forbidden-value action tests, error-class registry, UUID v4 instance-id tests, profile-claims serialization tests, and redaction-before-recording corpus. | Missing resource, privacy, instance-id, profile-claims, or error-class evidence. | Resource, detector, forbidden-value, unknown-attribute, SDK-limit, null-omission, instance-id, profile-claims, and error-class mapping tests pass. | `OIP-AC-008`, `OIP-AC-017`, `OIP-AC-027`, `OTEL-AC-010..014A` |
| 5 | OTel NLSpec §9, `OTEL-REQ-051..069`, `OTEL-REQ-129..130`; implementation-plan Phase 5 | Sampler profile tests, fixed trace-ID corpus, inbound context stripping, span registry tests, lifecycle boundary tests, parent/link/status tests, and forbidden-attribute tests. | Sampler owner conflict, missing span registry evidence, or status mapping mismatch. | Every emitted span matches registered name, kind, boundary, parent/link rule, status rule, and allowed/forbidden attributes. | `OIP-AC-004`, `OIP-AC-009`, `OIP-AC-018`, `OTEL-AC-015..020` |
| 6 | OTel NLSpec §10, `OTEL-REQ-070..085`, `OTEL-REQ-131..134`, `OTEL-REQ-147`; implementation-plan Phase 6 | Metric registry, identity validation, cumulative temporality, View filters, exemplar disablement, row-count bucket tests, overflow handling, and Bind bypass tests. | Missing metric registry, bucket, overflow, exemplar, View, temporality, or Bind evidence. | Every emitted metric matches registered stream identity and no unregistered stream, exemplar, overflow, or bypass appears outside the negative fixture. | `OIP-AC-010`, `OIP-AC-019`, `OTEL-AC-021..026A` |
| 7 | OTel NLSpec §11, `OTEL-REQ-086..094`; implementation-plan Phase 7 | Disabled bridge tests, enabled LogRecord mapping tests, severity mapping, body truncation including `body_max_chars=0`, exception reduction, EventName omission, and no span-event bridge. | Missing disabled-default or exception-sanitization evidence. | Logs export only under enabled bridge and only in owner-approved shape. | `OIP-AC-011`, `OIP-AC-020`, `OTEL-AC-027..031` |
| 8 | OTel NLSpec §12, `OTEL-REQ-095..112`, `OTEL-REQ-135..136`; implementation-plan Phase 8 | OTLP/HTTP path tests, one-target gRPC tests, secret header redaction, exact User-Agent grammar, retry envelope, retry classification, queue overflow, shutdown, recursion guard, and runtime-invariance tests. | Missing endpoint, header, User-Agent, retry, overflow, shutdown, recursion, or product-invariance evidence. | Endpoint, header, User-Agent, retry, overflow, shutdown, recursion, and product-invariance evidence pass. | `OIP-AC-012`, `OIP-AC-021`, `OIP-AC-026`, `OTEL-AC-032..039A` |
| 9 | OTel NLSpec §13, `OTEL-REQ-113..116`; implementation-plan Phase 9 | Source import, package graph, built bundle, dynamic import, source-map, runtime-probe, and non-transfer checks. | Required browser artifact cannot be inspected, or forbidden runtime package is browser-reachable. | Browser direct export and browser configuration authority are absent. | `OIP-AC-013`, `OIP-AC-022`, `OTEL-AC-040..042` |
| 10 | OTel NLSpec §14, `OTEL-REQ-117..120`; implementation-plan Phase 10 | `make otel-conformance`, summary schema, retained raw artifacts under run root, committed normalized corpus under `internal/testutil/golden/otel`, canonical number rules, fixture IDs `OTEL-CORPUS-001..018`, version normalization, and dependency-update classification. | Raw captures under committed golden paths, missing fixture ID coverage, or any unregistered provisional placeholder. | Corpus covers every fixture ID, raw and normalized artifacts are separated, and dependency updates classify shape changes. | `OIP-AC-014`, `OIP-AC-023`, `OTEL-AC-043..046` |
| 11 | Implementation-plan Phase 11 and acceptance registry | Release-readiness state, final conformance-status manifest, acceptance mapping, final ledger, retained evidence summary, and final registry states. | Any Decision Reconciliation Registry or Handoff Boundary Gap Registry row has a blocking closure state; any blocking TODO remains. | `otel_release_ready` only when `make otel-conformance` passes and every mapped criterion passes. | `OIP-AC-002..003`, `OIP-AC-015..023`, all imported `OTEL-AC-*` mapped by the plan |

The plan's recommended work breakdown maps these phases into `OTEL-1` through `OTEL-10`, starting with source snapshots and generated constants and ending with golden corpus and harness integration.[^9]

## 11. Canonical Artifacts to Inspect or Create

| Artifact family | Canonical path or boundary | Run rule |
| --- | --- | --- |
| OTel source snapshot | `contracts/otel/otel_source_snapshot.json` | Must reproduce the adopted baseline and reject `main`, short SHAs, placeholder digests, missing source paths, missing document status, and missing SDK package versions. |
| Generated constants manifest | `contracts/otel/generated_constants_manifest.json` | Must bind standard attributes and standard metrics to generated pinned sources or explicit allowlists. |
| Import-boundary manifest | `contracts/otel/import_boundary.json` | Must encode allowed API packages, bootstrap package roots, and forbidden package families. |
| Error-class registry | `contracts/otel/error_class_registry.json` | Must bind low-cardinality error classes and public error codes. |
| Golden corpus | `internal/testutil/golden/otel/**` | Must contain normalized committed corpus, not raw retained capture. |
| Conformance-status manifest | `contracts/otel/conformance_status.json` | Must declare exactly one closed conformance mode, read the Decision Reconciliation Registry and Handoff Boundary Gap Registry states, and inventory blocking TODOs. |
| Telemetry bootstrap | `internal/platform/telemetry` | Only this boundary may import SDK, exporters, processors, samplers, readers, log processors, resource detectors, or bootstrap-only package families. |
| API-facing telemetry accessors | `internal/platform/telemetry/api` | Ordinary modules may import this package and OpenTelemetry API packages only. |

These paths are conformance-visible in the implementation plan and are not optional aliases unless the plan is revised.[^10]

## 12. Decision Reconciliation Registry

Every row in this registry has exactly one `closure_state` from §3. Final done requires every row to have `closure_state` of `owner_closed`, `repo_materialized`, or `handoff_local_closed`.

### Decision-State Compatibility Table

| Decision row family | New closure state before implementation evidence | New closure state after required closure | Required transition |
| --- | --- | --- | --- |
| `OTEL-DQ-001..004` repo-control decisions | `repo_materialization_required` | `repo_materialized` | Inspect or create the named repo-control artifact, remove the blocking placeholder, and record evidence in the ledger. |
| `OTEL-DQ-005..006` owner-closed plan decisions | `owner_closed` | `owner_closed` | Keep the owner-imported rule unless live repo inspection proves artifact drift, in which case set `owner_drift_conflict`. |
| `OTEL-DQ-007..012` owner-closed coordination decisions | `owner_closed` | `owner_closed` | Use the owner-imported rule from the adopted OTel NLSpec and fail implementation evidence on mismatch. |
| Any owner mismatch discovered during the run | `owner_drift_conflict` | final state unavailable until handoff revision | Stop and revise the handoff before implementation continues. |

| Decision ID | Owner import | Closure state | Closure payload required | Permitted temporary representation | Final done rule | Linked handoff gap |
| --- | --- | --- | --- | --- | --- | --- |
| `OTEL-DQ-001` | OTel NLSpec §16 and `OTEL-AC-001`; `contracts/otel/otel_source_snapshot.json`. | `repo_materialization_required` | Concrete semantic-convention model digest and digest algorithm in `contracts/otel/otel_source_snapshot.json`. | `TODO(repo-adoption): compute semconv_model_digest` only before source-baseline materialization. | Final done requires `repo_materialized`; placeholder, malformed digest, or missing artifact blocks. | none |
| `OTEL-DQ-002` | OTel NLSpec §16, §4.1.3, and `OTEL-AC-001`; `contracts/otel/generated_constants_manifest.json`. | `repo_materialization_required` | Exact generated-constant source, package, or generator version with pinned provenance. | `TODO(repo-adoption): pin generated constants` only before generated-constant manifest materialization. | Final done requires `repo_materialized`; missing provenance or placeholder blocks. | none |
| `OTEL-DQ-003` | OTel NLSpec §16, §4.1.4, and `OTEL-AC-001..002`; package manifests and source snapshot. | `repo_materialization_required` | Exact OTel API, SDK, exporter, logs, metrics, trace, semantic-convention constants, bridge, and adapter package versions. | `TODO(repo-adoption): pin OTel package versions` only before repo-control package materialization. | Final done requires `repo_materialized`; missing family or placeholder version blocks. | none |
| `OTEL-DQ-004` | OTel NLSpec §8.5.1 and `OTEL-REQ-141..143`; `contracts/errors/*`; `contracts/otel/error_class_registry.json`. | `repo_materialization_required` | Mapping from `cartulary.error_code` to `cartulary.error_class` bound to the adopted public error-code registry. | `TODO(repo-adoption): bind error registry` only before registry materialization. | Final done requires `repo_materialized`; mismatch with the public error-code registry blocks. | none |
| `OTEL-DQ-005` | OTel NLSpec §14.1 and implementation-plan Phase 10. | `owner_closed` | Committed normalized corpus path under `internal/testutil/golden/otel`; raw captures under harness run root only. | No placeholder is permitted for final artifact paths. | Final done fails if raw captures are committed under golden paths or normalized goldens use a different path without owner revision. | none |
| `OTEL-DQ-006` | OTel NLSpec §12.4 and implementation-plan Phase 0/8. | `owner_closed` | Runtime retry jitter uses bounds-only conformance; deterministic RNG hook is optional test support and not runtime configuration. | No TODO needed. | Final done requires retry tests to assert bounds, cutoff, disabled retry, permanent rejection, timeout, shutdown, and non-blocking behavior. | none |
| `OTEL-DQ-007` | OTel NLSpec §10.2 and §10.4; implementation-plan Phase 0/6. | `owner_closed` | Explicit `{row}` bucket sequence `0`, `1`, `5`, `10`, `25`, `50`, `100`, `250`, `500`; measurement is serialized `rows[]` length in one successful view-query response, and upper bounds are inclusive. | No placeholder is permitted after owner closure. | Final metric evidence must prove the exact bucket sequence, zero-row behavior, and synthetic overflow handling. | `HANDOFF-GAP-002` |
| `OTEL-DQ-008` | OTel NLSpec §5.2 and §14.2; implementation-plan Phase 10. | `owner_closed` | Version source order is explicit config, Make-owned build metadata, then `0.0.0+unknown`; SemVer or unknown values are validated before replacing resource `service.version` with `SERVICE_VERSION` and scope versions with `SCOPE_VERSION`. | No placeholder is permitted after owner closure except the normalized golden tokens themselves. | Final golden evidence must prove byte-stable normalized output across version changes and reject invalid or inconsistent versions before normalization. | `HANDOFF-GAP-004` |
| `OTEL-DQ-009` | OTel NLSpec §12.2; implementation-plan Phase 8. | `owner_closed` | OTLP/gRPC accepts only absolute `http://host:port` or `https://host:port`; `https` uses TLS with system trust and `http` uses insecure transport; normalized target is `host:port`; custom CAs, mTLS, client certs, and `insecure_skip_verify` are not adopted. | No placeholder is permitted after owner closure. | Final gRPC endpoint evidence must prove exact accept/reject grammar, target normalization, and scheme-to-security selection. | `HANDOFF-GAP-003` |
| `OTEL-DQ-010` | OTel NLSpec §7.1 and §14.2; implementation-plan Phase 4. | `owner_closed` | `service.instance.id` defaults to a fresh canonical lowercase UUID v4 per process start; configured values must be canonical lowercase UUID v4 and not nil; provenance opacity is an unenforced operator invariant. | No placeholder is permitted after owner closure. | Final resource evidence must prove valid/invalid configured IDs, fresh defaults, and golden replacement after validation. | `HANDOFF-GAP-005` |
| `OTEL-DQ-011` | OTel NLSpec §7.1 and §8.2; implementation-plan Phase 4. | `owner_closed` | `cartulary.profile.claims` is a string scalar containing `base` plus claimed extension `profile_id` tokens, exact-token duplicates coalesced, ASCII byte sorted, comma-delimited, no spaces, no escaping, and never empty. | No placeholder is permitted after owner closure. | Final resource evidence must prove scalar shape, sort order, duplicate handling, Base-only output, and full claimed-profile output. | `HANDOFF-GAP-006` |
| `OTEL-DQ-012` | OTel NLSpec §6.1, §6.2.1, and §12.3; Core 04 `secret_ref_v1`; implementation-plan Phase 2. | `owner_closed` | Exporter headers are a map from header name to `secret_ref_v1`; raw literal values are invalid; max 16 configured headers, max 4096 UTF-8 bytes per resolved value, max 8192 bytes total configured block, lowercase duplicate detection, protocol-header override rejection, and fail-closed secret resolution are required. | No placeholder is permitted after owner closure. | Final config and exporter-header evidence must prove grammar, limits, duplicate handling, protocol-header rejection, missing-secret failure, and redacted diagnostics. | `HANDOFF-GAP-001` |

## 13. Handoff Boundary Gap Registry

Every row in this registry has exactly one `closure_state` from §3. Final done requires every row to have `closure_state` of `owner_closed`, `repo_materialized`, or `handoff_local_closed`.

| Gap ID | Boundary | Owner import | Closure state | What remains open | Why drift can occur | Required owner or repo closure payload | Permitted temporary representation | Final done rule | Acceptance evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `HANDOFF-GAP-001` | `telemetry.exporter.headers` grammar and limits | OTel NLSpec §6.1, §6.2.1, §12.3; Core 04 `secret_ref_v1`; implementation-plan Phase 2 and Phase 8; `OTEL-DQ-012` | `owner_closed` | None; owner text defines header grammar, secret binding, duplicate handling, protocol-owned header rejection, count/value/block limits, and diagnostics. | Implementations can still drift if they accept raw literals, count bytes differently, leak secret values, or let configured headers override protocol-owned headers. | Implement `map<header_name, secret_ref_v1>` with header-name regex `[A-Za-z0-9_.-]{1,64}`, ASCII lowercase duplicate detection, max 16 headers, max 4096-byte resolved value, max 8192-byte configured block counted as `lowercase_name + ": " + resolved_value`, and fail-closed secret resolution. | No placeholder is permitted after owner closure. | Final config and exporter-header conformance require tests for grammar, secret refs, duplicates, forbidden headers, limits, missing secrets, invalid values, and redacted diagnostics. | `OIP-AC-005..006`, `OTEL-AC-003..006`, `OTEL-AC-034` |
| `HANDOFF-GAP-002` | `{row}` histogram buckets for `cartulary.workbook.rows.returned` | OTel NLSpec §10.2 and §10.4; implementation-plan Phase 6; `OTEL-DQ-007` | `owner_closed` | None; owner text defines the ordered bucket sequence, inclusive upper bounds, serialized-row measurement source, zero-row behavior, and overflow fixture scope. | Implementations can still drift if they measure total matches, choose different buckets, or record overflow outside synthetic fixtures. | Use explicit boundaries `0`, `1`, `5`, `10`, `25`, `50`, `100`, `250`, `500`; measure serialized `rows[]` length in one successful view-query response; record `0` for empty success; reserve `>500` for negative or synthetic overflow fixtures. | No placeholder is permitted after owner closure. | Final metric conformance requires corpus assertions for `0`, `1`, `5`, `6`, `500`, and `501`. | `OIP-AC-010`, `OIP-AC-019`, `OTEL-AC-021..026A` |
| `HANDOFF-GAP-003` | OTLP/gRPC endpoint security | OTel NLSpec §12.2; implementation-plan Phase 8; `OTEL-DQ-009` | `owner_closed` | None; owner text defines allowed schemes, normalized target form, TLS/insecure mapping, credential non-adoption, and failure cases. | SDKs and exporters can still drift if wrapper code accepts host-only targets, strips schemes, adopts custom TLS material, or creates per-signal targets. | Accept only absolute `http://host:port` or `https://host:port` with explicit port and path empty or `/`; normalize to `host:port`; `https` uses system-trust TLS; `http` uses insecure transport; reject custom CA, mTLS, client certs, `insecure_skip_verify`, userinfo, query, fragment, non-empty path, and per-signal targets. | No placeholder is permitted after owner closure. | Final gRPC endpoint evidence requires accepted/rejected endpoint fixtures and secure/insecure channel assertions. | `OIP-AC-012`, `OIP-AC-021`, `OTEL-AC-033` |
| `HANDOFF-GAP-004` | Version normalization for `service.version` and instrumentation-scope `version` | OTel NLSpec §5.2 and §14.2; implementation-plan Phase 10; `OTEL-DQ-008` | `owner_closed` | None; owner text defines version source order, known/unknown predicate, valid format, placeholder names, pre-normalization validation, and resource/scope equality. | Goldens can still drift if implementations normalize before validation, use different fallback versions, or let scope versions diverge from Resource `service.version`. | Resolve version from explicit config, then Make-owned build metadata, then `0.0.0+unknown`; accept SemVer 2.0.0 or exact unknown only; use the same resolved value for Resource, scopes, and User-Agent; replace with `SERVICE_VERSION` and `SCOPE_VERSION` after validation. | No placeholder is permitted after owner closure except normalized golden tokens. | Final golden conformance requires byte-stable output across build versions and rejection of invalid or inconsistent raw version fields. | `OIP-AC-014`, `OIP-AC-023`, `OTEL-AC-043..046` |
| `HANDOFF-GAP-005` | `service.instance.id` opacity | OTel NLSpec §7.1, §8.2, §14.2; implementation-plan Phase 4; `OTEL-DQ-010` | `owner_closed` | None; owner text defines canonical UUID v4 structure, default generation, nil rejection, ASCII/lowercase policy, and the unenforced operator provenance invariant. | Implementations can still drift if they accept arbitrary operator text, Unicode, uppercase UUIDs, nil UUID, or stable host/process-derived defaults. | Generate a fresh canonical lowercase UUID v4 per process start; accept configured canonical lowercase UUID v4 values only when non-nil; document that UUIDs must not encode host, user, incident, customer, path, object-store, or process identity. | No placeholder is permitted after owner closure. | Final resource conformance requires UUID v4 accept/reject fixtures, fresh default-generation evidence, and post-validation golden normalization. | `OIP-AC-017`, `OIP-AC-027`, `OTEL-AC-010..014A` |
| `HANDOFF-GAP-006` | `cartulary.profile.claims` representation | OTel NLSpec §7.1 and §8.2; implementation-plan Phase 4; `OTEL-DQ-011` | `owner_closed` | None; owner text defines scalar value shape, delimiter, sort collation, duplicate rule, Base-only representation, and invalid empty-set behavior. | Implementations can still drift if they emit arrays, preserve duplicates, sort locale-aware, add spaces, escape tokens, or omit `base`. | Emit a string scalar containing `base` plus every claimed extension `profile_id`; coalesce duplicates by exact token; sort by ASCII byte ascending order; serialize with comma delimiter, no spaces, no escaping; emit `base` for Base-only; never emit empty string. | No placeholder is permitted after owner closure. | Final resource byte equality requires scalar, sorting, duplicate, Base-only, and full-profile fixtures. | `OIP-AC-017`, `OTEL-AC-010..014A` |
| `HANDOFF-GAP-007` | Imported fixture coverage using `representative` | OTel NLSpec §14.1, §15.2, §15.3, §15.7; implementation-plan Phase 10 | `owner_closed` | None; owner text replaces representative wording with `OTEL-CORPUS-001..018` fixture IDs and binary pass/fail predicates. | Implementations can still drift if corpus manifests do not materialize every fixture ID or silently narrow a predicate. | Materialize corpus fixtures for no-SDK, source baseline, parser, hostile env/config, HTTP, workbook, WebSocket, jobs, Postgres, object storage, resource identity, attribute boundary, logs, metrics, exporter, runtime invariance, and redaction exactly as `OTEL-CORPUS-001..018`. | No placeholder is permitted after owner closure. | Final corpus evidence requires every fixture ID to be present, mapped to input and expected output, and pass/fail binary. | `OIP-AC-014`, `OIP-AC-023`, `OTEL-AC-008`, `OTEL-AC-012`, `OTEL-AC-014A`, `OTEL-AC-039A`, `OTEL-AC-043..046` |
| `HANDOFF-GAP-008` | Span status mapping | OTel NLSpec §9.1; implementation-plan Phase 5 span result/status table | `owner_closed` | None; owner text defines outcome-to-status mapping and safe error attributes. | Implementations can still drift if they set `Ok`, leave rejected/conflict unset, model telemetry drops as product spans, or leak exception detail. | Do not set OTel `Ok`; success and normal user-requested job cancellation use `Unset`; rejected, conflict, timeout, failed, abnormal canceled, and product dropped use `Error`; telemetry-only drops use metrics/log diagnostics; error spans may emit only `cartulary.error_code`, `cartulary.error_class`, and `error.type=cartulary.error_class`. | No placeholder is permitted after owner closure. | Final span-shape conformance requires per-result fixtures across HTTP, workbook, WebSocket, jobs, Postgres, and object-store families. | `OIP-AC-009`, `OIP-AC-018`, `OTEL-AC-015..020` |
| `HANDOFF-GAP-009` | Exporter User-Agent grammar | OTel NLSpec §12.3; implementation-plan Phase 8 | `owner_closed` | None; owner text defines exact grammar, segment order, version sources, exporter-version failure behavior, and forbidden extras. | Implementations can still drift by adding environment names, comments, product identifiers, hostnames, or by omitting unresolved exporter-version failures. | Emit exactly `Cartulary/<SERVICE_VERSION> OTel-OTLP-Exporter-go/<EXPORTER_VERSION>` with one ASCII space; use the resolved service version and repo-control exporter package version; fail export-enabled startup if exporter version is unavailable; forbid comments, parentheses, extra identifiers, paths, hosts, environment names, or operator text. | No placeholder is permitted after owner closure. | Final User-Agent evidence requires accept/reject fixtures and forbidden-family scans of exporter requests and retained artifacts. | `OIP-AC-012`, `OIP-AC-021`, `OTEL-AC-035` |
| `HANDOFF-GAP-010` | OTLP/HTTP endpoint URL grammar | OTel NLSpec §12.2; implementation-plan Phase 8 | `owner_closed` | None; owner text defines accepted URL grammar, canonicalization, rejected edge cases, and path joining. | Implementations can still drift on default ports, uppercase canonicalization, percent-encoded paths, duplicate slashes, dot segments, IPv6 hosts, IDNA, and path normalization. | Accept only absolute `http`/`https` URLs with explicit port, ASCII host or bracketed IPv6, no userinfo/query/fragment; lowercase scheme and host; reject missing ports, percent-encoded paths, duplicate slashes, dot segments, empty internal segments, Unicode/IDNA hosts, and unsupported schemes; append `/v1/traces`, `/v1/metrics`, or `/v1/logs` after stripping one trailing slash. | No placeholder is permitted after owner closure. | Final OTLP/HTTP endpoint evidence requires accepted and rejected URL fixture matrices, including uppercase canonicalization, IPv6, path joining, and rejected percent/dot/duplicate-slash cases. | `OIP-AC-012`, `OIP-AC-021`, `OTEL-AC-032` |
| `HANDOFF-GAP-011` | Structured secret references | OTel NLSpec §6.1, §6.2.1; Core 04 deployment configuration owner; implementation-plan Phase 2 | `owner_closed` | None; Core 04 owns `secret_ref_v1`; OTel imports it for exporter headers and HMAC secret references. | Implementations can still drift if they resolve the wrong environment variable, permit unsafe names, resolve after readiness, or leak reference values. | Implement `secret_ref_v1` as `{kind='env', name='<safe_name>'}` with current `kind` exactly `env`; `name` matches `[A-Za-z0-9][A-Za-z0-9_.-]{0,63}` and resolves to `CARTULARY_SECRET_<NORMALIZED_NAME>`; missing, empty, invalid, CR/LF/NUL, leading/trailing whitespace, non-visible-ASCII, or unresolved secrets fail before readiness with redacted diagnostics. | No placeholder is permitted after owner closure. | Final config and secret diagnostics evidence requires fixtures for valid refs, invalid refs, missing env, empty env, unsafe values, redacted diagnostics, exporter headers, and HMAC secret references. | `OIP-AC-005..006`, `OTEL-AC-004A`, `OTEL-AC-034` |

## 14. DQ-to-Gap Mapping Table

| Decision ID | Handoff gap | Mapping rule |
| --- | --- | --- |
| `OTEL-DQ-007` | `HANDOFF-GAP-002` | Row-count histogram evidence imports owner closure from OTel NLSpec §10.2 and §10.4; implementation evidence still must prove the exact buckets and measurement source. |
| `OTEL-DQ-008` | `HANDOFF-GAP-004` | Byte-stable version normalization imports owner closure from OTel NLSpec §5.2 and §14.2; implementation evidence still must prove validation before replacement. |
| `OTEL-DQ-009` | `HANDOFF-GAP-003` | OTLP/gRPC channel-security evidence imports owner closure from OTel NLSpec §12.2; implementation evidence still must prove exact endpoint grammar and scheme-to-security selection. |
| `OTEL-DQ-010` | `HANDOFF-GAP-005` | Resource conformance for `service.instance.id` imports owner closure from OTel NLSpec §7.1 and §14.2; implementation evidence still must prove UUID v4 behavior. |
| `OTEL-DQ-011` | `HANDOFF-GAP-006` | Resource byte equality for `cartulary.profile.claims` imports owner closure from OTel NLSpec §7.1 and §8.2; implementation evidence still must prove scalar serialization. |
| `OTEL-DQ-012` | `HANDOFF-GAP-001` | Config and exporter-header conformance imports owner closure from OTel NLSpec §6.1, §6.2.1, §12.3, and Core 04 `secret_ref_v1`; implementation evidence still must prove limits and redaction. |
| none | `HANDOFF-GAP-007..011` | These additional boundary-completeness rows are now owner-closed; implementation evidence still must materialize fixture coverage, span status, User-Agent, OTLP/HTTP endpoint, and secret-reference tests. |

## 15. Validation Commands and Evidence Rules

Use Make as the canonical command surface. Direct package scripts, raw Go, Vitest, Playwright, Biome, Vite, pnpm, or tool-specific commands are developer conveniences unless invoked by a public Make target.[^11]

The final target is `make otel-conformance`. During implementation, run the narrowest relevant Make target. A package-level child command may be recorded only when the public target does not yet exist; omission behavior is that child-command output remains provisional evidence and cannot satisfy final harness evidence.

Every retained harness evidence reference MUST include result root, run ID, run root, target name, summary artifact, failure class, failure reason, rerun command, and investigation command when the harness exposes them. The harness owns result-root, retained-artifact, target-selection, output-mode, exit-code, failure-classification, and cleanup mechanics.[^12]

`CARTULARY_OUTPUT_MODE=summary` is the default operator-friendly mode. Use `verbose` only for live investigation and `machine` only when a target explicitly accepts exactly-one-JSON output.

## 16. Stop Conditions

Stop the goal run and produce a handoff update instead of continuing when any of these conditions is true:

1. Two owner documents appear to conflict and the conflict affects implementation behavior, conformance status, emitted telemetry shape, or release-readiness state.
2. A live-repo fact is unavailable and §7 does not permit a `TODO(repo-adoption)` for it.
3. A phase needs a value from a Decision Reconciliation Registry or Handoff Boundary Gap Registry row whose `closure_state` is `repo_materialization_required`, `owner_closure_required`, or `owner_drift_conflict`.
4. A new boundary gap is found that could affect deterministic implementation or verification and no `HANDOFF-GAP-*` row exists for it.
5. Final evidence contains a blocking `TODO(repo-adoption)` or an unregistered provisional placeholder.
6. The only way to proceed would add a telemetry sidecar, required Collector, vendor backend, browser exporter, product-visible workflow change, case-data authority, or default network export.
7. `make otel-conformance` or its precursor target cannot emit retained evidence because harness ownership, result-root behavior, or artifact schema is unresolved.
8. A test failure is caused by product behavior outside the telemetry subsystem and cannot be fixed without changing Core 00 through Core 04 behavior.
9. The live owner documents close a decision differently from this handoff's registry. In that case, mark the affected row `owner_drift_conflict`, stop, and revise the handoff before implementation continues.

## 17. Final Done Checklist

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
- Every row in the Decision Reconciliation Registry has `closure_state` of `owner_closed`, `repo_materialized`, or `handoff_local_closed`.
- Every row in the Handoff Boundary Gap Registry has `closure_state` of `owner_closed`, `repo_materialized`, or `handoff_local_closed`.
- No row in either registry has `closure_state` of `repo_materialization_required`, `owner_closure_required`, or `owner_drift_conflict`.
- No blocking `TODO(repo-adoption)` appears in a conformance-visible artifact, registry, fixture, phase gate, acceptance mapping, release-readiness manifest, prompt, or final evidence.
- No unregistered provisional placeholder appears in conformance-visible evidence.

## 18. Handoff Update Template

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
- Decision registry closure_state:
- Boundary gap registry closure_state:
- Owner blockers:
- Owner drift conflicts:
- Conformance-visible artifacts:
- Exact next action:
```

## 19. Document Revision Acceptance Criteria

- **HANDOFF-AC-001:** The document states that it is an execution handoff and not a telemetry behavior owner.
- **HANDOFF-AC-002:** Every normative handoff requirement imports an owner section or ID, or defines local goal-run behavior.
- **HANDOFF-AC-003:** The closure-state vocabulary is singular, exhaustive, and used by the phase ledger, Decision Reconciliation Registry, Handoff Boundary Gap Registry, prompts, stop conditions, final checklist, and handoff update template.
- **HANDOFF-AC-004:** The decision registry reconciles `OTEL-DQ-001` through `OTEL-DQ-012` with current owner state, gives each row one `closure_state`, and uses only the §3 closure-state vocabulary.
- **HANDOFF-AC-005:** Every `HANDOFF-GAP-*` row has the exact registry schema columns: `Gap ID`, `Boundary`, `Owner import`, `Closure state`, `What remains open`, `Why drift can occur`, `Required owner or repo closure payload`, `Permitted temporary representation`, `Final done rule`, and `Acceptance evidence`.
- **HANDOFF-AC-006:** `HANDOFF-GAP-001` through `HANDOFF-GAP-011` are present and cover header grammar/limits, row histogram buckets, gRPC security, version normalization, instance-id opacity, profile-claims representation, fixture coverage predicates, span status mapping, User-Agent grammar, OTLP/HTTP URL grammar, and structured secret references.
- **HANDOFF-AC-007:** The DQ-to-gap mapping table links `OTEL-DQ-007..012` to `HANDOFF-GAP-001..006` and states that `HANDOFF-GAP-007..011` are additional owner-closed implementation-evidence rows.
- **HANDOFF-AC-008:** Every phase gate maps to at least one `OIP-AC-*` or `OTEL-AC-*` criterion.
- **HANDOFF-AC-009:** Every default, bound, omission rule, null rule, and edge-case family named in the handoff has an owner import, final-allowed registry state, or explicit blocking registry row.
- **HANDOFF-AC-010:** Every handoff-owned uppercase **MAY** has explicit omission behavior.
- **HANDOFF-AC-011:** The final done checklist is impossible to satisfy while any Decision Reconciliation Registry or Handoff Boundary Gap Registry row has a blocking closure state, while any blocking `TODO(repo-adoption)` remains, or while any unregistered provisional placeholder remains.
- **HANDOFF-AC-012:** The handoff imports owner-defined values for `{row}` buckets, version normalization, gRPC transport security, instance-id opacity, profile-claim collation, header size bounds, span status mapping, User-Agent grammar, OTLP/HTTP URL edge cases, and structured secret references from the controlling Core 04 and OTel NLSpec sections.
- **HANDOFF-AC-013:** No handoff-owned final rule uses the invalid latitude phrases from §3 unless the same sentence, paragraph, or table row supplies a binary predicate or the unresolved phrase is represented by a blocking `HANDOFF-GAP-*` row.
- **HANDOFF-AC-014:** The handoff update template includes decision registry state, boundary gap registry state, owner blockers, owner drift conflicts, imported acceptance IDs, retained evidence, and exact next action.

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
[^13]: `docs/research/nlspec-spec.md`, especially "Explicit Defaults and Boundaries", "What Completeness Means", and "Intentional vs. Accidental Ambiguity".
