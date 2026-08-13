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

## 2. Product-fixture bindings and observable-timing overlays

**REQ-05-004**
Claim-bearing publication MUST bind its `fixture_ids[]` and
`measurement_predicate_ids[]` to the current exact Core 04 registry. Core 05
does not redefine product-supported fixture shape, actions, start states, stop
predicates, thresholds, or implementation sampling policy.
Profiles: claim_publication
Verified by: PC-003

The Core 04 fixtures define incident shape and concurrent load only. They do not
by themselves define a claim-bearing benchmark environment.

**REQ-05-005**
Latency measurements used for claim-bearing publication MUST preserve the exact
Core 04 product-visible interval and sampling policy. Publication MAY add
environment qualification and retention requirements but MUST NOT change the
product predicate or estimator.
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

## 4. Measurement-predicate publication binding

**REQ-05-011**
The current claim-publication profile MUST bind every timed or fixture-sensitive
criterion to an exact current `measurement_predicate_id` and fixture ID owned by
Core 04 REQ-04-157 through REQ-04-159. A historical predicate ID MAY remain in a
retained historical manifest but MUST NOT satisfy a current publication claim.
Profiles: claim_publication
Verified by: PC-003

For the current profile, the timed or fixture-sensitive criteria are `AC-003`, `AC-008`, `AC-011`, `AC-016`, `AC-027`, `AC-030`, `AC-033`, `AC-043`, `AC-044`, `AC-045`, `AC-046`, `AC-047`, and `AC-132`.

**REQ-05-012**
The claim-bearing benchmark profile MUST use the Core 04 large-grid fixture's
`traffic_trace_id='cartulary.perf.live_updates_25sessions.v1'`, seed, and warm
state, and MUST declare `warmup_passes=1`. Publication bindings MUST preserve
the fixture and predicate selected by the implementation criterion.
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
- **PC-003**: For each timed or fixture-sensitive criterion in the current profile, the result binds to an exact current Core 04 `measurement_predicate_id` and fixture ID and preserves their start state, action, stop predicate, threshold, and sampling mode without local reinterpretation.
  - Verifies: REQ-05-004..REQ-05-005, REQ-05-011..REQ-05-012
- **PC-004**: A claim-bearing benchmark run is non-conformant if authentication, session handling, CSRF protection, sanitization, safe-preview restrictions, or integrity checks are disabled, or if a claim-bearing visible-state result is produced from headless browser mode.
  - Verifies: REQ-05-010
- **PC-005**: The benchmark artifact bundle retains the emitted `benchmark_manifest`, raw timing samples or event logs, harness identifier and version, environment identifiers, and an `artifact_bundle_sha256` sufficient to replay and audit the claim.
  - Verifies: REQ-05-009
- **PC-006**: A claim-bearing benchmark publication is conformant only when every requirement selected by `profile:claim_publication` is implemented, every applicable publication criterion listed in this manifest passes, and the underlying implementation claim already passes for each published timed or fixture-sensitive criterion.
  - Verifies: `profile:claim_publication`
