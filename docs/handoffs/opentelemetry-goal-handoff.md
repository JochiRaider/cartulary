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
Execute docs/handoffs/opentelemetry-goal-handoff.md as the active Cartulary OpenTelemetry implementation handoff. Read opentelemetry-implementation-plan.md, opentelemetry-instrumentation-nlspec.md, testing-harness-nlspec.md, Core 00-05, docs/domain.md, and relevant repo-control files before editing. Preserve authority boundaries: the adopted OTel NLSpec owns telemetry behavior, Core 00-04 own product conformance, Core 05 owns publication, the harness NLSpec owns Make/harness mechanics, and the implementation plan owns sequencing, artifacts, phase gates, and acceptance mapping. Use both the Decision Reconciliation Registry and the Handoff Boundary Gap Registry; do not use a blanket OTEL-DQ rule. Treat every Decision Registry or Boundary Gap Registry row with closure_state repo_materialization_required, owner_closure_required, or owner_drift_conflict as blocking. Treat HANDOFF-GAP-007 through HANDOFF-GAP-011 as blockers even though they do not have matching OTEL-DQ rows. Do not invent owner values in the handoff or implementation. Inspect the live repo before asserting package, target, file, schema, or convention facts. Use TODO(repo-adoption): ... only under §7 TODO semantics. Implement phases in dependency order unless an earlier phase is already fully satisfied. Keep telemetry inside the modular monolith; do not add a sidecar, required Collector, vendor dependency, browser exporter, product workflow, case-data authority, or default network export. Run the narrowest relevant Make target after each slice and record changed files, owner imports, evidence, failures, blockers, decision_status, gap_status, and next_action. Final done is unavailable while any Decision Registry or Boundary Gap Registry row has a blocking closure_state or while any blocking TODO(repo-adoption) remains. If owner specs conflict, stop and report the exact conflict; do not pick a side.
```

Use this shorter prompt when the handoff file is outside the repository and must be attached or pasted separately.

```text
Execute the attached OpenTelemetry goal handoff and opentelemetry-implementation-plan.md. Treat the adopted OTel NLSpec as telemetry behavior authority, Core 00-04 as product authority, Core 05 as publication-only authority, and the harness NLSpec as Make/harness authority. Use both the Decision Reconciliation Registry and Handoff Boundary Gap Registry; do not use a blanket OTEL-DQ rule. Treat any closure_state of repo_materialization_required, owner_closure_required, or owner_drift_conflict as blocking. Treat HANDOFF-GAP-007 through HANDOFF-GAP-011 as blockers even without matching OTEL-DQ rows. Do not invent owner values. Inspect live repo facts before claiming them. Final done is unavailable while any Decision Registry or Boundary Gap Registry row has a blocking closure_state or while any blocking TODO(repo-adoption) remains.
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
| Config defaults, bounds, omitted behavior, explicit null, empty env, secret status, and failure behavior | OTel NLSpec §6, `OTEL-REQ-024..030`, `OTEL-REQ-122..125`; implementation-plan Phase 2, `OIP-AC-005..006`; `OTEL-AC-003..006` | Materialize the full telemetry key table, parser families, cross-key matrix, and hostile environment/config fixtures before treating config as complete. | Different defaults, null handling, empty env handling, secret diagnostics, or startup failure behavior. | `OTEL-DQ-012`, `HANDOFF-GAP-001`, and `HANDOFF-GAP-011`. | Generated schema/config tests plus `OIP-AC-005..006` and `OTEL-AC-003..006`. |
| TODO and omission semantics | Implementation-plan source limits and partial implementation sections; this handoff §7 | Permit placeholders only for live-repo facts or named decision blockers; partial work remains `adopted_incomplete`. | Silent provisional behavior, accidental conformance claims, or export before phase completion. | Any blocking TODO in conformance-visible evidence. | Ledger `blocking_todos`, conformance-status manifest, `OIP-AC-001..003`. |
| Resource identity and null omission | OTel NLSpec §7 and §8.2, `OTEL-REQ-031..038`, `OTEL-REQ-126..128`; implementation-plan Phase 4; `OTEL-AC-010..014A` | Require closed resource attributes, empty Resource `schema_url`, detector suppression, forbidden-value action tests, and no null attribute setter calls. | Divergent resource attributes, schema URL merge behavior, detector leakage, null emission, instance ID acceptance, or profile-claim byte shape. | `OTEL-DQ-010`, `OTEL-DQ-011`, `HANDOFF-GAP-005`, and `HANDOFF-GAP-006`. | Resource, detector, forbidden-value, SDK-limit, and null-omission tests; `OIP-AC-017`, `OIP-AC-027`. |
| Span lifecycle and parent/link/status behavior | OTel NLSpec §9, `OTEL-REQ-051..069`; implementation-plan Phase 5, `OIP-AC-009`; `OTEL-AC-015..020` | Require registered span names, lifecycle boundaries, local parent rules, disabled custom span events, inbound context stripping, and closed status mapping evidence. | Different span topology, status mapping, forbidden attributes, or remote-context trust. | `HANDOFF-GAP-008` or owner conflict in span registry or sampler profile. | Span-shape, sampler, deterministic trace-ID, status mapping, and remote-context tests. |
| Metric identity, temporality, aggregation, overflow, and histogram buckets | OTel NLSpec §10, `OTEL-REQ-070..085`, `OTEL-REQ-131..134`; implementation-plan Phase 6, `OIP-AC-010`; `OTEL-AC-021..026A` | Require emitted metrics to match registry identity and cumulative temporality; final metric conformance cannot pass without row-count histogram bucket closure. | Divergent metric streams, View names, attribute filters, exemplars, overflow handling, or histogram buckets. | `OTEL-DQ-007` and `HANDOFF-GAP-002`. | Metric registry, View, exemplar, overflow, temporality, bucket, and Bind tests. |
| Log bridge default, body bound, severity, EventName omission, and exception reduction | OTel NLSpec §11, `OTEL-REQ-086..094`; implementation-plan Phase 7, `OIP-AC-011`; `OTEL-AC-027..031` | Keep OTel log export disabled by default; when enabled, require exact top-level mapping, severity table, body bound behavior, no EventName, and safe exception reduction. | Divergent log export defaults, arbitrary severity text, unsafe exception fields, or span-event bridging. | Missing log bridge mapping tests. | Disabled bridge, enabled mapping, exception, EventName, and span-event tests. |
| Exporter endpoint, headers, User-Agent, retry, queue overflow, shutdown, and runtime invariance | OTel NLSpec §12, `OTEL-REQ-095..112`, `OTEL-REQ-135..136`; implementation-plan Phase 8, `OIP-AC-012`, `OIP-AC-021`, `OIP-AC-026`; `OTEL-AC-032..039A` | Require no default endpoint, deterministic OTLP/HTTP paths, one gRPC target, redacted headers, bounded retry envelope, `drop_new`, idempotent shutdown, and product-runtime invariance. | Different egress defaults, per-signal divergence, retry timing, header leakage, queue blocking, User-Agent shape, URL canonicalization, or product behavior changes. | `OTEL-DQ-009`, `OTEL-DQ-012`, `HANDOFF-GAP-001`, `HANDOFF-GAP-003`, `HANDOFF-GAP-009`, `HANDOFF-GAP-010`, and `HANDOFF-GAP-011`. | Endpoint, gRPC, header, User-Agent, retry, overflow, shutdown, recursion, and runtime-invariance tests. |
| Browser non-export and browser configuration absence | OTel NLSpec §13, `OTEL-REQ-113..116`; implementation-plan Phase 9, `OIP-AC-013`, `OIP-AC-022`; `OTEL-AC-040..042` | Require source, package graph, bundle, dynamic import, source-map, and runtime probes to fail closed when required artifacts cannot be inspected. | Browser exporter, vendor SDK, analytics path, or browser state configuring telemetry. | Required browser artifact cannot be inspected. | Browser bundle, config, and non-transfer tests. |
| Golden raw/normalized artifact separation and canonical normalization | OTel NLSpec §14, `OTEL-REQ-117..120`; implementation-plan Phase 10, `OIP-AC-014`, `OIP-AC-023`; `OTEL-AC-043..046` | Keep raw captures under run root and committed normalized goldens under `internal/testutil/golden/otel`; normalize only owner-permitted volatile fields. | Raw capture commits, unstable goldens, or normalization hiding shape facts. | `OTEL-DQ-008` and `HANDOFF-GAP-004`. | Golden corpus, raw artifact, normalization, and dependency-update classification evidence. |
| Acceptance mapping and conformance-mode reporting | Implementation-plan Phase 0, Phase 11, `OIP-AC-002..003`, `OIP-AC-015..023` | Require exactly one conformance mode and require every phase gate to map to local or imported acceptance IDs. | Ambiguous readiness, skipped imported criteria, or inconsistent startup/harness/release status. | Any Decision Reconciliation Registry or Handoff Boundary Gap Registry row has a blocking closure state. | Conformance-status manifest, harness summary, acceptance mapping table, and final checklist. |

## 10. Phase Order and Completion Gate

| Phase | Required owner imports | Required artifacts or checks | Explicit blockers | Completion evidence | Acceptance IDs |
| --- | --- | --- | --- | --- | --- |
| 0 | Implementation-plan Phase 0; OTel NLSpec §1 and §16; Core 00-05 authority owners | Adoption posture, conformance-status manifest, decision registry update, threat-model update if release-impacting telemetry boundaries are introduced. | Any Decision Reconciliation Registry or Handoff Boundary Gap Registry row without a final-allowed closure state blocks `adopted_conformant`. | Exactly one conformance mode and no unsupported conformance claim. | `OIP-AC-001..003`, `OIP-AC-024` |
| 1 | Implementation-plan Phase 1; OTel NLSpec §4, §5.2; `OTEL-AC-001..002`, `OTEL-AC-007..009` | `contracts/otel/otel_source_snapshot.json`, `contracts/otel/generated_constants_manifest.json`, `contracts/otel/import_boundary.json`, static import checks, and generated-constant drift checks. | Placeholder source values, short SHAs, missing package versions, or forbidden imports. | Source baseline and import-boundary evidence exists or reports a blocking owner TODO under §7. | `OIP-AC-004`, `OIP-AC-016`, `OTEL-AC-001..002`, `OTEL-AC-007..009` |
| 2 | OTel NLSpec §6, `OTEL-REQ-024..030`, `OTEL-REQ-122..125`; implementation-plan Phase 2 | Full telemetry key table, parser fixtures for omitted, empty, valid, invalid, and explicit string `"null"` cases, cross-key matrix, secret redaction, and hostile OTel environment/config fixtures. | `OTEL-DQ-012`, `HANDOFF-GAP-001`, and `HANDOFF-GAP-011` block final config/header evidence. | Generated schema and config tests cover every key default, bound, omission rule, explicit-null rule, empty-env rule, secret status, failure behavior, and hazard family. | `OIP-AC-005..006`, `OTEL-AC-003..006` |
| 3 | OTel NLSpec §4.3, §5.2, §5.3, §8.2; implementation-plan Phase 3 | API-facing accessors, registered scope validation, no-SDK lifecycle tests, and OTel API spy or emitted-capture null tests. | Unknown scope, SDK import in ordinary module, or unsafe no-SDK path. | Ordinary modules import only API packages/accessors, no-SDK product paths complete, and no null setter call occurs. | `OIP-AC-007`, `OIP-AC-016`, `OTEL-AC-007..009`, `OTEL-AC-014A` |
| 4 | OTel NLSpec §7 and §8, `OTEL-REQ-031..050`, `OTEL-REQ-126..128`, `OTEL-REQ-141..143`; implementation-plan Phase 4 | Closed resource registry, empty Resource `schema_url`, detector suppression, null omission, forbidden-value action tests, error-class registry, and redaction-before-recording corpus. | `OTEL-DQ-010`, `OTEL-DQ-011`, `HANDOFF-GAP-005`, and `HANDOFF-GAP-006`. | Resource, detector, forbidden-value, unknown-attribute, SDK-limit, null-omission, instance-id, profile-claims, and error-class mapping tests pass. | `OIP-AC-008`, `OIP-AC-017`, `OIP-AC-027`, `OTEL-AC-010..014A` |
| 5 | OTel NLSpec §9, `OTEL-REQ-051..069`, `OTEL-REQ-129..130`; implementation-plan Phase 5 | Sampler profile tests, fixed trace-ID corpus, inbound context stripping, span registry tests, lifecycle boundary tests, parent/link/status tests, and forbidden-attribute tests. | `HANDOFF-GAP-008`, sampler owner conflict, or missing span registry evidence. | Every emitted span matches registered name, kind, boundary, parent/link rule, status rule, and allowed/forbidden attributes. | `OIP-AC-004`, `OIP-AC-009`, `OIP-AC-018`, `OTEL-AC-015..020` |
| 6 | OTel NLSpec §10, `OTEL-REQ-070..085`, `OTEL-REQ-131..134`; implementation-plan Phase 6 | Metric registry, identity validation, cumulative temporality, View filters, exemplar disablement, overflow handling, and Bind bypass tests. | `OTEL-DQ-007` and `HANDOFF-GAP-002` block final metric conformance for `cartulary.workbook.rows.returned`. | Every emitted metric matches registered stream identity and no unregistered stream, exemplar, overflow, or bypass appears outside the negative fixture. | `OIP-AC-010`, `OIP-AC-019`, `OTEL-AC-021..026A` |
| 7 | OTel NLSpec §11, `OTEL-REQ-086..094`; implementation-plan Phase 7 | Disabled bridge tests, enabled LogRecord mapping tests, severity mapping, body truncation including `body_max_chars=0`, exception reduction, EventName omission, and no span-event bridge. | Missing disabled-default or exception-sanitization evidence. | Logs export only under enabled bridge and only in owner-approved shape. | `OIP-AC-011`, `OIP-AC-020`, `OTEL-AC-027..031` |
| 8 | OTel NLSpec §12, `OTEL-REQ-095..112`, `OTEL-REQ-135..136`; implementation-plan Phase 8 | OTLP/HTTP path tests, one-target gRPC tests, header redaction, User-Agent grammar, retry envelope, retry classification, queue overflow, shutdown, recursion guard, and runtime-invariance tests. | `OTEL-DQ-009`, `OTEL-DQ-012`, `HANDOFF-GAP-001`, `HANDOFF-GAP-003`, `HANDOFF-GAP-009`, `HANDOFF-GAP-010`, and `HANDOFF-GAP-011`. | Endpoint, header, User-Agent, retry, overflow, shutdown, recursion, and product-invariance evidence pass or report owner blockers. | `OIP-AC-012`, `OIP-AC-021`, `OIP-AC-026`, `OTEL-AC-032..039A` |
| 9 | OTel NLSpec §13, `OTEL-REQ-113..116`; implementation-plan Phase 9 | Source import, package graph, built bundle, dynamic import, source-map, runtime-probe, and non-transfer checks. | Required browser artifact cannot be inspected, or forbidden runtime package is browser-reachable. | Browser direct export and browser configuration authority are absent. | `OIP-AC-013`, `OIP-AC-022`, `OTEL-AC-040..042` |
| 10 | OTel NLSpec §14, `OTEL-REQ-117..120`; implementation-plan Phase 10 | `make otel-conformance`, summary schema, retained raw artifacts under run root, committed normalized corpus under `internal/testutil/golden/otel`, canonical number rules, fixture IDs `OTEL-CORPUS-001..018`, and dependency-update classification. | `OTEL-DQ-008`, `HANDOFF-GAP-004`, raw captures under committed golden paths, or any unregistered provisional placeholder. | Corpus covers every fixture ID, raw and normalized artifacts are separated, and dependency updates classify shape changes. | `OIP-AC-014`, `OIP-AC-023`, `OTEL-AC-043..046` |
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
| `OTEL-DQ-007..012` owner-coordination decisions | `owner_closure_required` | `owner_closed` only after the controlling owner document supplies every required payload field | Stop at affected phases until the owner document closes the value; do not invent values in implementation or handoff text. |
| Any owner mismatch discovered during the run | `owner_drift_conflict` | final state unavailable until handoff revision | Stop and revise the handoff before implementation continues. |

| Decision ID | Owner import | Closure state | Closure payload required | Permitted temporary representation | Final done rule | Linked handoff gap |
| --- | --- | --- | --- | --- | --- | --- |
| `OTEL-DQ-001` | OTel NLSpec §16 and `OTEL-AC-001`; `contracts/otel/otel_source_snapshot.json`. | `repo_materialization_required` | Concrete semantic-convention model digest and digest algorithm in `contracts/otel/otel_source_snapshot.json`. | `TODO(repo-adoption): compute semconv_model_digest` only before source-baseline materialization. | Final done requires `repo_materialized`; placeholder, malformed digest, or missing artifact blocks. | none |
| `OTEL-DQ-002` | OTel NLSpec §16, §4.1.3, and `OTEL-AC-001`; `contracts/otel/generated_constants_manifest.json`. | `repo_materialization_required` | Exact generated-constant source, package, or generator version with pinned provenance. | `TODO(repo-adoption): pin generated constants` only before generated-constant manifest materialization. | Final done requires `repo_materialized`; missing provenance or placeholder blocks. | none |
| `OTEL-DQ-003` | OTel NLSpec §16, §4.1.4, and `OTEL-AC-001..002`; package manifests and source snapshot. | `repo_materialization_required` | Exact OTel API, SDK, exporter, logs, metrics, trace, semantic-convention constants, bridge, and adapter package versions. | `TODO(repo-adoption): pin OTel package versions` only before repo-control package materialization. | Final done requires `repo_materialized`; missing family or placeholder version blocks. | none |
| `OTEL-DQ-004` | OTel NLSpec §8.5.1 and `OTEL-REQ-141..143`; `contracts/errors/*`; `contracts/otel/error_class_registry.json`. | `repo_materialization_required` | Mapping from `cartulary.error_code` to `cartulary.error_class` bound to the adopted public error-code registry. | `TODO(repo-adoption): bind error registry` only before registry materialization. | Final done requires `repo_materialized`; mismatch with the public error-code registry blocks. | none |
| `OTEL-DQ-005` | OTel NLSpec §14.1 and implementation-plan Phase 10. | `owner_closed` | Committed normalized corpus path under `internal/testutil/golden/otel`; raw captures under harness run root only. | No placeholder is permitted for final artifact paths. | Final done fails if raw captures are committed under golden paths or normalized goldens use a different path without owner revision. | none |
| `OTEL-DQ-006` | OTel NLSpec §12.4 and implementation-plan Phase 0/8. | `owner_closed` | Runtime retry jitter uses bounds-only conformance; deterministic RNG hook is optional test support and not runtime configuration. | No TODO needed. | Final done requires retry tests to assert bounds, cutoff, disabled retry, permanent rejection, timeout, shutdown, and non-blocking behavior. | none |
| `OTEL-DQ-007` | OTel NLSpec §10.2 and §10.4; implementation-plan Phase 0/6. | `owner_closure_required` | Explicit ordered `{row}` histogram bucket sequence for `cartulary.workbook.rows.returned`. | `TODO(repo-adoption): define {row} histogram buckets` in phase evidence only. | Final metric conformance is unavailable until the owner defines bucket sequence and HANDOFF-GAP-002 is final-allowed. | `HANDOFF-GAP-002` |
| `OTEL-DQ-008` | OTel NLSpec §14.2; implementation-plan Phase 10. | `owner_closure_required` | Normalization rules for `service.version` and instrumentation-scope `version`. | `SERVICE_VERSION` and `SCOPE_VERSION` placeholders in provisional tests only after reporting owner blocker. | Final byte-stable golden conformance is unavailable until the owner defines version normalization and HANDOFF-GAP-004 is final-allowed. | `HANDOFF-GAP-004` |
| `OTEL-DQ-009` | OTel NLSpec §12.2; implementation-plan Phase 8. | `owner_closure_required` | OTLP/gRPC endpoint-scheme to transport-security mapping. | `TODO(repo-adoption): define gRPC scheme to TLS mapping` in endpoint-gate evidence only. | Final gRPC endpoint evidence is unavailable until the owner defines allowed schemes and secure/insecure mapping and HANDOFF-GAP-003 is final-allowed. | `HANDOFF-GAP-003` |
| `OTEL-DQ-010` | OTel NLSpec §7.1 and §14.2; implementation-plan Phase 4. | `owner_closure_required` | Closed `service.instance.id` structural opacity predicate plus marked unenforced provenance invariant. | `TODO(repo-adoption): define instance-id opacity predicate` in resource-gate evidence only. | Final resource conformance is unavailable until the owner defines deterministic accept/reject predicate and invariant wording and HANDOFF-GAP-005 is final-allowed. | `HANDOFF-GAP-005` |
| `OTEL-DQ-011` | OTel NLSpec §7.1 and §8.2; implementation-plan Phase 4. | `owner_closure_required` | Exact `cartulary.profile.claims` representation, sort collation, duplicate handling, and empty-set representation. | `TODO(repo-adoption): pin profile.claims representation` for unresolved subparts only. | Final resource byte equality is unavailable until every listed subpart is owner-defined and HANDOFF-GAP-006 is final-allowed. | `HANDOFF-GAP-006` |
| `OTEL-DQ-012` | OTel NLSpec §6.1; implementation-plan Phase 2. | `owner_closure_required` | Maximum header count, maximum header value length, and maximum total header-block size for `telemetry.exporter.headers`. | `TODO(repo-adoption): bound header count and size` in config-gate evidence only. | Final config and exporter-header conformance are unavailable until the owner defines exact bounds and counting rules and HANDOFF-GAP-001 is final-allowed. | `HANDOFF-GAP-001` |

## 13. Handoff Boundary Gap Registry

Every row in this registry has exactly one `closure_state` from §3. Final done requires every row to have `closure_state` of `owner_closed`, `repo_materialized`, or `handoff_local_closed`.

| Gap ID | Boundary | Owner import | Closure state | What remains open | Why drift can occur | Required owner or repo closure payload | Permitted temporary representation | Final done rule | Acceptance evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `HANDOFF-GAP-001` | `telemetry.exporter.headers` grammar and limits | OTel NLSpec §6.1, §6.2.1, §12.3; implementation-plan Phase 2 and Phase 8; `OTEL-DQ-012` | `owner_closure_required` | Header-pair grammar, whitespace and escaping rules, duplicate-key handling beyond exact-byte comparison, maximum header count, maximum value length, maximum total header-block size, and counting units. | Implementations can parse comma-separated values, secret bindings, whitespace, duplicate keys, bytes, Unicode scalars, and protocol-required headers differently. | Owner must define exact header grammar, structured secret-binding shape or owner pointer, duplicate handling, max count, max value length, max total size, counting units, whether protocol-required headers count, and boundary fixtures. | `TODO(repo-adoption): bound header count and size` in config-gate evidence only. | Blocks final config and exporter-header conformance until closure payload is owner-defined and tests pass. | `OIP-AC-005..006`, `OTEL-AC-003..006`, `OTEL-AC-034` |
| `HANDOFF-GAP-002` | `{row}` histogram buckets for `cartulary.workbook.rows.returned` | OTel NLSpec §10.2 and §10.4; implementation-plan Phase 6; `OTEL-DQ-007` | `owner_closure_required` | Ordered bucket sequence, boundary inclusivity, zero-row behavior, overflow bucket behavior, and golden fixture assertions. | Implementations can choose different row-count buckets and emit different metric streams for the same workbook result. | Owner must define the exact ordered bucket sequence, unit rule, inclusivity, zero-row handling, overflow handling, and corpus assertions for byte-stable output. | `TODO(repo-adoption): define {row} histogram buckets` in metric phase evidence only. | Blocks final metric conformance until closure payload is owner-defined and metric corpus passes. | `OIP-AC-010`, `OIP-AC-019`, `OTEL-AC-021..026A` |
| `HANDOFF-GAP-003` | OTLP/gRPC endpoint security | OTel NLSpec §12.2; implementation-plan Phase 8; `OTEL-DQ-009` | `owner_closure_required` | Allowed endpoint schemes, target normalization, TLS/insecure mapping, default port behavior, certificate and mTLS non-adoption or adoption rules, and failure cases. | SDKs and exporters can interpret `http`, `https`, host-only targets, stripped schemes, and credentials differently. | Owner must define allowed schemes, rejected schemes, normalized target form, secure/insecure credential selection, default-port behavior, certificate/mTLS rule, and fixtures for accepted and rejected endpoint families. | `TODO(repo-adoption): define gRPC scheme to TLS mapping` in endpoint-gate evidence only. | Blocks final gRPC endpoint evidence until closure payload is owner-defined and endpoint tests pass. | `OIP-AC-012`, `OIP-AC-021`, `OTEL-AC-033` |
| `HANDOFF-GAP-004` | Version normalization for `service.version` and instrumentation-scope `version` | OTel NLSpec §5.2 and §14.2; implementation-plan Phase 10; `OTEL-DQ-008` | `owner_closure_required` | Build-version source, known/unknown predicate, valid format, placeholder names, and comparison behavior for resource and scope versions. | Goldens can vary across build versions, and implementers can disagree on when build version exists or whether resource and scope versions must match. | Owner must define build-version source, unknown fallback predicate, valid format, placeholder tokens, pre-normalization validation, comparison behavior, and whether resource and scope versions must match. | `SERVICE_VERSION` and `SCOPE_VERSION` placeholders in provisional tests only after reporting owner blocker. | Blocks final byte-stable golden conformance until owner normalization is defined and corpus proves version-invariant equality. | `OIP-AC-014`, `OIP-AC-023`, `OTEL-AC-043..046` |
| `HANDOFF-GAP-005` | `service.instance.id` opacity | OTel NLSpec §7.1, §8.2, §14.2; implementation-plan Phase 4; `OTEL-DQ-010` | `owner_closure_required` | Structural accept/reject predicate, UUID v4 default form, forbidden detectable patterns, Unicode/ASCII policy, and marked unenforced provenance invariant. | Implementations can accept different configured strings or normalize different values before replacing instance IDs in goldens. | Owner must define exact structural predicate, configured-value accept/reject rules, generated UUID v4 form, forbidden detectable patterns, Unicode or ASCII policy, and the explicit unenforced operator provenance invariant. | `TODO(repo-adoption): define instance-id opacity predicate` in resource-gate evidence only. | Blocks final resource conformance until predicate and invariant are owner-defined and resource tests pass. | `OIP-AC-017`, `OIP-AC-027`, `OTEL-AC-010..014A` |
| `HANDOFF-GAP-006` | `cartulary.profile.claims` representation | OTel NLSpec §7.1 and §8.2; implementation-plan Phase 4; `OTEL-DQ-011` | `owner_closure_required` | Value shape, delimiter or array rule, sort collation, duplicate rule, empty-set representation, and byte-stable fixtures. | Implementations can emit string or array values, choose different sorting, preserve or collapse duplicates, or represent empty claims differently. | Owner must define the exact value shape, delimiter or array structure, escaping impossibility or grammar, sort collation, duplicate handling, empty-set representation, and byte-stable fixture expectations. | `TODO(repo-adoption): pin profile.claims representation` for unresolved subparts only. | Blocks final resource byte equality until every listed subpart is owner-defined and resource corpus passes. | `OIP-AC-017`, `OTEL-AC-010..014A` |
| `HANDOFF-GAP-007` | Imported fixture coverage using `representative` | OTel NLSpec §14.1, §15.2, §15.3, §15.7; implementation-plan Phase 10 | `owner_closure_required` | Owner acceptance text names sample-style coverage without exact fixture IDs or binary coverage predicates for every affected product path and forbidden-value family. | Implementations can choose different HTTP, workbook, job, Postgres, object-store, log, WebSocket, forbidden-value, and runtime-invariance scenarios. | Owner or repo-control corpus manifest must map each such phrase to exact fixture IDs, corpus inputs, expected outputs, and pass/fail predicates. | `TODO(repo-adoption): close fixture coverage predicate` in corpus planning notes only. | Blocks final corpus and runtime-invariance evidence until all sample-style coverage is mapped to exact fixture IDs or binary predicates. | `OIP-AC-014`, `OIP-AC-023`, `OTEL-AC-008`, `OTEL-AC-012`, `OTEL-AC-014A`, `OTEL-AC-039A`, `OTEL-AC-043..046` |
| `HANDOFF-GAP-008` | Span status mapping | OTel NLSpec §9.1; implementation-plan Phase 5 span result/status table | `owner_closure_required` | Closed outcome-to-status mapping for success, rejected, conflict, canceled, timeout, dropped, and failed outcomes per span family. | Implementations can emit different OTel status values for the same product or dependency outcome. | Owner must define one status and error-attribute rule per outcome per span family, including validation rejection, authorization rejection, conflict, abnormal cancellation, telemetry drop, and dependency failure cases. | `TODO(repo-adoption): close span status mapping` in span-shape evidence only. | Blocks final span-shape conformance until status mapping is closed and tested. | `OIP-AC-009`, `OIP-AC-018`, `OTEL-AC-015..020` |
| `HANDOFF-GAP-009` | Exporter User-Agent grammar | OTel NLSpec §12.3; implementation-plan Phase 8 | `owner_closure_required` | Exact segment order, inclusion predicates, build-version fallback, exporter-version omission behavior, allowed characters, and forbidden extras. | Implementations can include or omit exporter identity differently, vary segment order, or leak unsafe text while satisfying a broad allowlist. | Owner must define exact User-Agent grammar, segment order, inclusion and omission predicates, allowed character set, version fallback, exporter-version absence behavior, and forbidden-extra tests. | `TODO(repo-adoption): close user-agent grammar` in exporter evidence only. | Blocks final User-Agent evidence until grammar is owner-defined and tests pass. | `OIP-AC-012`, `OIP-AC-021`, `OTEL-AC-035` |
| `HANDOFF-GAP-010` | OTLP/HTTP endpoint URL grammar | OTel NLSpec §12.2; implementation-plan Phase 8 | `owner_closure_required` | Accepted URL grammar, canonicalization rules, rejected edge cases, and fixture matrix beyond the four path examples. | Implementations can diverge on uppercase schemes, default ports, percent-encoded paths, duplicate slashes, dot segments, IPv6 hosts, IDNA, and path normalization. | Owner must define accepted URL grammar, canonicalization algorithm, rejected URL families, path-join behavior for edge cases, and fixture matrix for accepted and rejected inputs. | `TODO(repo-adoption): close OTLP HTTP URL grammar` in endpoint evidence only. | Blocks final OTLP/HTTP endpoint evidence until grammar and fixture matrix are owner-defined and tests pass. | `OIP-AC-012`, `OIP-AC-021`, `OTEL-AC-032` |
| `HANDOFF-GAP-011` | Structured secret references | OTel NLSpec §6.1, §6.2.1; Core 04 deployment configuration owner; implementation-plan Phase 2 | `owner_closure_required` | Secret-reference shape, safe reference-name predicate, missing-secret behavior, diagnostics rule, and fixture coverage for exporter headers and HMAC secret references. | Implementations can parse secret references differently, expose unsafe reference names, or handle missing secrets differently. | Owner must define secret-reference data shape, allowed reference-name grammar, resolution timing, missing-secret failure behavior, redacted diagnostics, and fixture coverage for `telemetry.exporter.headers` and `telemetry.attribute.hmac_secret_ref`. | `TODO(repo-adoption): close structured secret reference contract` in config evidence only. | Blocks final config and secret diagnostics evidence until owner or repo-control artifact closes secret-reference behavior and tests pass. | `OIP-AC-005..006`, `OTEL-AC-004A`, `OTEL-AC-034` |

## 14. DQ-to-Gap Mapping Table

| Decision ID | Handoff gap | Mapping rule |
| --- | --- | --- |
| `OTEL-DQ-007` | `HANDOFF-GAP-002` | Row-count histogram conformance remains blocked until both rows are final-allowed. |
| `OTEL-DQ-008` | `HANDOFF-GAP-004` | Byte-stable version normalization remains blocked until both rows are final-allowed. |
| `OTEL-DQ-009` | `HANDOFF-GAP-003` | OTLP/gRPC channel-security evidence remains blocked until both rows are final-allowed. |
| `OTEL-DQ-010` | `HANDOFF-GAP-005` | Resource conformance for `service.instance.id` remains blocked until both rows are final-allowed. |
| `OTEL-DQ-011` | `HANDOFF-GAP-006` | Resource byte equality for `cartulary.profile.claims` remains blocked until both rows are final-allowed. |
| `OTEL-DQ-012` | `HANDOFF-GAP-001` | Config and exporter-header conformance remain blocked until both rows are final-allowed. |
| none | `HANDOFF-GAP-007..011` | These are additional boundary-completeness blockers. Their absence from the original decision registry does not make them optional. |

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
- **HANDOFF-AC-007:** The DQ-to-gap mapping table links `OTEL-DQ-007..012` to `HANDOFF-GAP-001..006` and states that `HANDOFF-GAP-007..011` are additional blockers.
- **HANDOFF-AC-008:** Every phase gate maps to at least one `OIP-AC-*` or `OTEL-AC-*` criterion.
- **HANDOFF-AC-009:** Every default, bound, omission rule, null rule, and edge-case family named in the handoff has an owner import, final-allowed registry state, or explicit blocking registry row.
- **HANDOFF-AC-010:** Every handoff-owned uppercase **MAY** has explicit omission behavior.
- **HANDOFF-AC-011:** The final done checklist is impossible to satisfy while any Decision Reconciliation Registry or Handoff Boundary Gap Registry row has a blocking closure state, while any blocking `TODO(repo-adoption)` remains, or while any unregistered provisional placeholder remains.
- **HANDOFF-AC-012:** The handoff contains no invented values for `{row}` buckets, version normalization, gRPC transport security, instance-id opacity, profile-claim collation, header size bounds, span status mapping, User-Agent grammar, OTLP/HTTP URL edge cases, or structured secret references.
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
