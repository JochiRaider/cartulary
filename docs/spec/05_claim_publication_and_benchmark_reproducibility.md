# Cartulary Normative Companion 05: Claim Publication and Benchmark Reproducibility

## 1. Scope and separation

Core 00 through Core 04 define implementation correctness for the current profile. This companion defines claim-bearing publication requirements for timed or fixture-sensitive criteria only. It is normative companion material and is not part of Base Profile or extension-profile implementation conformance.

Within this companion, `Profiles: claim_publication` is a companion-local traceability tag. It is not a Base or extension implementation profile and MUST NOT appear in runtime extension discovery or Core 04 implementation claim manifests.

**REQ-05-001**
Core 00 through Core 04 define implementation correctness. This companion defines only claim-bearing publication requirements for timed or fixture-sensitive criteria and MUST NOT add product runtime behavior or broaden Base Profile or extension-profile implementation conformance.
Profiles: claim_publication
Verified by: PC-006

**REQ-05-002**
A public claim that an implementation satisfies one or more timed or fixture-sensitive criteria MUST NOT be claim-bearing unless the underlying implementation claim already passes for the cited criteria.
Profiles: claim_publication
Verified by: PC-006

**REQ-05-003**
Results produced without satisfying this companion MAY be reported as informative engineering measurements but MUST NOT satisfy a claim-bearing publication claim.
Profiles: claim_publication
Verified by: PC-001, PC-002, PC-006

## 2. Benchmark fixtures and observable-timing rules

**REQ-05-004**
For performance-sensitive publication criteria, the following reference performance fixtures apply:

- **Fixture A: large-grid incident**
  - 20,000 timeline rows
  - 1,000 host rows
  - 1,000 identity rows
  - 25 concurrently connected analyst sessions on one incident, with presence enabled and representative live row-update traffic
  - representative tags, mentions, and links, but not evidence-heavy per row
- **Fixture B: evidence-heavy incident**
  - 5,000 timeline rows
  - 10,000 evidence records
  - tens of GB of binary evidence stored in object storage
  - at least one timeline row linked to 100 evidence records
  - evidence blobs MAY be stubbed for throughput tests, but evidence metadata, counts, attachment state, and preview handles MUST be real
Profiles: claim_publication
Verified by: PC-003

These fixtures define incident shape and concurrent load only. They do not by themselves define a claim-bearing benchmark environment.

Unless a publication criterion states otherwise, `reference incident` in this companion means Fixture A.

**REQ-05-005**
Latency measurements used for claim-bearing publication MUST use end-user-observable completion time from the initiating user action to the required visible UI state. Any criterion expressed as p95 MUST be evaluated over at least 100 completed operations of the named interaction after one warm-up pass on the named fixture.
Profiles: claim_publication
Verified by: PC-003

## 3. Benchmark profile registry and benchmark-manifest contract

**REQ-05-006**
Each timed or fixture-sensitive criterion used for claim-bearing publication MUST be evaluated under a named `benchmark_profile_id`. Results produced without a named benchmark profile MAY be reported as informative engineering measurements but MUST NOT satisfy a claim-bearing publication claim.
Profiles: claim_publication
Verified by: PC-001, PC-002

**REQ-05-007**
Unless a criterion declares another benchmark profile explicitly, the current claim-bearing benchmark profile is `cartulary.perf.desktop_ref.v1`.
Profiles: claim_publication
Verified by: PC-002

**REQ-05-008**
`cartulary.perf.desktop_ref.v1` MUST be a closed benchmark-profile object with these exact values:

- `browser.engine='chromium'`
- `browser.build='134.0.6998.35'`
- `browser.mode='headed'`
- `browser.extensions='none'`
- `browser.viewport_css_px='1440x900'`
- `browser.device_scale_factor=1`
- `browser.zoom_percent=100`
- `client_runner_id='aws.ec2.c7i.2xlarge'`
- `client_os_image_id='cartulary.bench.ubuntu_24_04_client.2026q1'`
- `client_reserved_vcpu=8`
- `client_reserved_memory_gib=16`
- `client_power_mode='performance'`
- `app_runner_id='aws.ec2.c7i.2xlarge'`
- `app_os_image_id='cartulary.bench.ubuntu_24_04_app.2026q1'`
- `app_reserved_vcpu=8`
- `app_reserved_memory_gib=16`
- `postgres_runner_id='aws.ec2.i4i.2xlarge'`
- `postgres_os_image_id='cartulary.bench.ubuntu_24_04_postgres.2026q1'`
- `postgres_reserved_vcpu=8`
- `postgres_reserved_memory_gib=32`
- `postgres_storage_class='instance_store_nvme'`
- `object_store_runner_id='aws.ec2.c7i.xlarge'`
- `object_store_os_image_id='cartulary.bench.ubuntu_24_04_object.2026q1'`
- `object_store_reserved_vcpu=4`
- `object_store_reserved_memory_gib=8`
- `object_store_storage_class='gp3_ssd'`
- `client_to_app_link_mbps=1000`
- `client_to_app_rtt_ms_max=2`
- `client_to_app_loss_percent=0`
- `client_to_app_jitter_ms_max=1`
- `app_to_postgres_rtt_ms_max=1`
- `app_to_object_store_rtt_ms_max=1`
- `traffic_trace_id='cartulary.perf.live_updates_25sessions.v1'`
- `seed=20260405`
- `warmup_passes=1`
- `authenticated_session_state='complete'`
- `incident_open_state='open'`
- `surface_warm_state='loaded'`
- `benchmark_harness_id='cartulary.bench.harness.playwright.v1'`
- `benchmark_harness_version='2026.04.0'`

Claim-bearing publication MUST compare this benchmark environment by exact value match. `Equivalent hardware`, `equivalent browser`, `similar LAN`, or similar open-ended language is non-conformant for claim-bearing publication.
Profiles: claim_publication
Verified by: PC-002

**REQ-05-009**
Claim-bearing benchmark runs MUST emit one durable `benchmark_manifest` conforming to `cartulary.benchmark_manifest.v1`. The manifest MUST be retained with the benchmark artifact bundle and MUST include, at minimum:

- `benchmark_manifest_schema_id`
- `benchmark_profile_id`
- `criterion_ids[]`
- `measurement_predicate_ids[]`
- `fixture_ids[]`
- `traffic_trace_id`
- `seed`
- `warmup_passes`
- `browser_engine`
- `browser_build`
- `client_runner_id`
- `client_os_image_id`
- `app_runner_id`
- `app_os_image_id`
- `postgres_runner_id`
- `postgres_os_image_id`
- `postgres_storage_class`
- `object_store_runner_id`
- `object_store_os_image_id`
- `object_store_storage_class`
- `benchmark_harness_id`
- `benchmark_harness_version`
- `run_started_at`
- `run_completed_at`
- `sample_count`
- `artifact_bundle_sha256`
- `security_controls_state`

A timed or fixture-sensitive claim without a conformant `benchmark_manifest` is non-conformant.
Profiles: claim_publication
Verified by: PC-001, PC-005

**REQ-05-010**
Claim-bearing benchmark runs MUST keep ordinary security controls enabled. Benchmark execution MUST NOT disable authentication, session handling, CSRF protection, sanitization, safe-preview restrictions, or integrity checks. Headless browser runs MAY be used for engineering diagnostics, but claim-bearing visible-state criteria MUST use the benchmark profile's headed browser mode.
Profiles: claim_publication
Verified by: PC-004

## 4. Measurement-predicate registry

**REQ-05-011**
The current profile MUST define a closed `measurement_predicate_id` registry for every timed or fixture-sensitive acceptance criterion it uses for claim-bearing publication. Each registry entry MUST define the initiating user action, required start state, stop predicate, bound fixture, warm-state assumptions, and whether the criterion is p95-sampled or single-observation. When a visible-state measurement predicate would otherwise vary by editor family, field class, or harness realization, the registry entry MUST bind an exact `view_schema_id`, exact `field_key`, exact literal initiating payload when input content matters, and exact anchor invariants sufficient to make the stop predicate decidable without local interpretation. One `measurement_predicate_id` MUST NOT denote a family of interchangeable editor-specific or harness-specific observables. A later profile that needs a different anchor, editor family, or stop predicate MUST define a different exact `measurement_predicate_id`.
Profiles: claim_publication
Verified by: PC-003

For the current profile, the timed or fixture-sensitive criteria are `AC-003`, `AC-008`, `AC-011`, `AC-016`, `AC-027`, `AC-030`, `AC-033`, `AC-043`, `AC-044`, `AC-045`, `AC-046`, `AC-047`, and `AC-132`.

| `measurement_predicate_id` | Bound criterion or criteria | Start state and initiating action | Stop predicate | Fixture and sampling |
| --- | --- | --- | --- | --- |
| `perf.timeline_paste_20x5.v1` | `AC-003` | Timeline surface loaded with default sort, filter, and grouping state; target range visible; timing starts when the paste commit is accepted. | 20 committed rows are visible with mapped values in five writable visible columns and stable `record_id` plus `row_version` binding on each new row. | Fixture A; single observation |
| `perf.presence_delta.rendered.v1` | `AC-008`, `AC-132` | Analyst A commits a workbook-surface, focused-row, or same-cell edit-state presence change on an incident. | Analyst B renders the corresponding workbook-header, row-gutter, or same-cell indicator from matching `sheet_ref`, `record_id`, and `field_key`. | Fixture A; single observation |
| `perf.rollback_or_row_restore.rendered.v1` | `AC-011` | A reviewer confirms rollback or whole-row restore. | The visible row reflects the new attributed revision and the history surface shows the new entry. | Fixture A; single observation |
| `perf.job_progress.visible_with_cancel.v1` | `AC-016`, `AC-027`, `AC-030`, `AC-033`, `AC-046` | The user submits an import, evidence-processing action, projection rebuild, snapshot generation, report generation, or reference-pack action that must remain backgrounded. | Visible progress UI and a cancel affordance render, and another grid row can be selected and accept text input without modal capture. | Fixture A or Fixture B as named by the criterion; single observation |
| `perf.selection_change.v1` | `AC-043` | A loaded workbook surface is visible and timing starts when a user action changes the selected grid cell or row. | The new selection is painted and keyboard input would target the newly selected item. | Fixture A; p95 |
| `perf.focus_change.v1` | `AC-043` | A loaded workbook surface is visible and timing starts when a user action changes the focused editable cell or row. | The focus state is visibly rendered and direct typing would target that field. | Fixture A; p95 |
| `perf.typing_ack.v1` | `AC-043` | `cartulary.view.timeline.v1` is loaded; an existing visible Timeline row is already in active edit mode for `field_key='timeline.summary'`; the same `record_id` remains selected; the text editor is focused; the caret is collapsed at the end of the current rendered text; there is no selection; IME composition is inactive; timing starts when the editor accepts the literal printable ASCII character `x`. | The same active text editor for the same `record_id` and `field_key` visibly renders the pre-keypress text with one trailing literal `x` appended, and input remains anchored to that same `record_id` and `field_key`. `Saved`, `Syncing`, validation badges, conflict badges, DOM mutation counts, network completion, collaboration messages, or `row_version` change MUST NOT satisfy the predicate. | Fixture A; p95 |
| `perf.timeline_blank_row_create.v1` | `AC-043` | Timeline surface loaded with default sort, filter, and grouping state; a blank row is visible; timing starts when the user commits Enter on a blank row containing one qualifying non-empty value. | The committed row is visible with stable `record_id`, stable `row_version`, and the entered value rendered in the target field. | Fixture A; p95 |
| `perf.view_change.first_useful_viewport.v1` | `AC-044` | The active surface is loaded in its default current state; timing starts when the user submits a sort, filter, or grouping change. | The first useful viewport defined in Core 04 §9 is visible. | Fixture A; p95 |
| `perf.view_change.stable_viewport.v1` | `AC-044` | The active surface is loaded in its default current state; timing starts when the user submits a sort, filter, or grouping change. | The stable viewport defined in Core 04 §9 is visible. | Fixture A; p95 |
| `perf.evidence_inspector.metadata_shell.v1` | `AC-045` | A user opens the inspector on a Timeline row linked to 100 evidence records. | The selected-row summary, total linked-evidence count, and first rendered evidence-list window are visible, and each evidence item in that first rendered window shows filename or media-type label, attachment state, and preview-handle availability. Binary preview bytes are not required. | Fixture B; p95 |
| `perf.anchor_stability_under_live_updates.v1` | `AC-047` | A deterministic live-update trace begins while an analyst holds a pending edit anchored to one `record_id`. | Throughout the trace, scrolling, sorting, filtering, grouping, and live updates never retarget the pending edit away from that `record_id`, and viewport stabilization completes without a full-sheet rerender. | Fixture A; pass or fail seeded scenario |

**REQ-05-012**
The claim-bearing benchmark profile MUST use `traffic_trace_id='cartulary.perf.live_updates_25sessions.v1'` and `warmup_passes=1`. Unless a measurement predicate declares a criterion-specific override, authentication MUST already be complete, the incident MUST already be open, the relevant workbook surface MUST already be loaded, and the active sort, filter, and grouping state MUST be the default current state before timing starts.
Profiles: claim_publication
Verified by: PC-002, PC-003

**REQ-05-013**
Additional benchmark profiles MAY be defined later, but they are informative only unless a Core 05 publication claim manifest or a publication criterion explicitly promotes them to claim-bearing status.
Profiles: claim_publication
Verified by: PC-002

## 5. Publication claim manifest

The manifest below defines the companion-local publication claim without redefining implementation claim boundaries.

Definition of Done:

- prerequisite claim: relevant implementation claim for each published timed or fixture-sensitive criterion
- requirement selector: `profile:claim_publication`
- required publication criteria: `PC-001..PC-006`

- **PC-001**: A claim-bearing timed or fixture-sensitive result is non-conformant if no `benchmark_manifest` exists, if `benchmark_profile_id` is missing, or if any required field in `cartulary.benchmark_manifest.v1` is absent.
  - Verifies: REQ-05-003, REQ-05-006, REQ-05-009
- **PC-002**: A claim-bearing timed or fixture-sensitive result is non-conformant if the emitted benchmark profile, browser build, browser mode, runner IDs, storage classes, network values, `traffic_trace_id`, `seed`, `warmup_passes`, or declared warm state differ from `cartulary.perf.desktop_ref.v1`; such a run MAY be reported only as informative.
  - Verifies: REQ-05-003, REQ-05-006..REQ-05-008, REQ-05-012..REQ-05-013
- **PC-003**: For each timed or fixture-sensitive criterion in the current profile, the result binds to an exact `measurement_predicate_id` from the companion registry, and that predicate makes the start state, stop predicate, fixture, and sampling mode decidable without local interpretation. When the predicate depends on a specific editor or visible edit artifact, the registry also binds an exact `view_schema_id`, exact `field_key`, exact initiating payload where relevant, and exact anchor invariants rather than an interchangeable equivalent observable.
  - Verifies: REQ-05-004..REQ-05-005, REQ-05-011..REQ-05-012
- **PC-004**: A claim-bearing benchmark run is non-conformant if authentication, session handling, CSRF protection, sanitization, safe-preview restrictions, or integrity checks are disabled, or if a claim-bearing visible-state result is produced from headless browser mode.
  - Verifies: REQ-05-010
- **PC-005**: The benchmark artifact bundle retains the emitted `benchmark_manifest`, raw timing samples or event logs, harness identifier and version, environment identifiers, and an `artifact_bundle_sha256` sufficient to replay and audit the claim.
  - Verifies: REQ-05-009
- **PC-006**: A claim-bearing benchmark publication is conformant only when every requirement selected by `profile:claim_publication` is implemented, every applicable publication criterion listed in this manifest passes, and the underlying implementation claim already passes for each published timed or fixture-sensitive criterion.
  - Verifies: `profile:claim_publication`
