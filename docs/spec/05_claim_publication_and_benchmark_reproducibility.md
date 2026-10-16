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
Unless a criterion declares another benchmark profile explicitly, the current claim-bearing benchmark profile is `cartulary.perf.desktop_ref.v3`.
Profiles: claim_publication
Verified by: PC-002

**REQ-05-008**
`cartulary.perf.desktop_ref.v1` is the immutable historical PostgreSQL 16
benchmark-profile object with these exact values:

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

Retained historical publication evidence MUST compare this benchmark environment
by exact value match and MUST NOT be reinterpreted as current evidence.
Profiles: claim_publication
Verified by: PC-002

**REQ-05-009**
Claim-bearing benchmark runs MUST retain a closed
`cartulary.benchmark_manifest.v2` envelope with exactly `schema_id`,
`claim_bearing=true`, `benchmark_profile_id`, `environment`, `source_digest`,
`criterion_ids`, `security_controls_state`, and `bundle`. The environment MUST
match the current profile exactly. Security controls MUST be explicit Boolean
`true` values for authentication, session handling, CSRF protection,
sanitization, safe-preview restrictions, and integrity checks. Unknown members,
duplicate JSON members, retired schemas/profiles and unknown criteria MUST fail.

`bundle` contains a normalized relative `path` and a `sha256:<hex>` digest of
the actual retained `cartulary.benchmark_evidence_bundle.v1` JSON bytes. This
bundle is a closed index with `schema_id` and nonempty `runs[]`. Each run has
`root`, `manifest`, `summary`, and nonempty `measurements[]`. `root` is relative
to the bundle directory. `manifest` and `summary` are digested relative file
references to the canonical run manifest and run summary. Each measurement
contains digested relative `summary`, `report`, and `browser_environment`
references within that run. Every referenced digest MUST be recomputed from
actual bytes; the same rule applies recursively to the observation, physical
fixture build and clone/cleanup evidence referenced by a measurement summary.
The envelope stays beside the bundle and is excluded from its own hash.
There is no metadata-only hash substitute or archive-path fallback.

All runs MUST be successful, fresh, complete and bound to the envelope's exact
source digest. Each selected measurement MUST qualify under the current
owner-controlled predicate/fixture binding, current physical snapshot contract,
independently recomputed retained samples/statistics, required cleanup and zero
scheduler overlap. The corresponding Playwright result MUST be a single passed
execution without retries, joined by the shared result adapter to the exact
catalog project, file and full suite/test title path. Its observation and browser-environment attachment
MUST match the referenced retained artifacts. A
`cartulary.browser_execution_environment.v1` receipt records the actual
`browser_engine`, `browser_build`, `browser_mode`, `playwright_version`,
`node_version`, `platform`, and `architecture` from the executing measurement
fixtures. Browser build MUST match the managed renderer attestation. Browser mode MUST
come from the browser-level `Browser.getVersion` product/user-agent tokens of
that pinned Chromium runtime, with unknown or inconsistent identities rejected.
Page/context user-agent overrides and local launch-option defaults cannot attest
a connected browser. Raw user-agent strings MUST NOT be retained in the receipt. A manifest assertion cannot override a headless or mismatched receipt.

Every current predicate required for each declared criterion MUST occur exactly
once. A missing, extra, unknown, duplicated or retired predicate, mismatched
fixture or criterion, incomplete sample set, failed threshold or unqualified
run MUST fail. A criterion without an admitted current exercising binding is
blocked from publication. No default global sample count, warm-up count,
fixture ID or traffic trace may substitute for the criterion's own contract.

Readers MUST reject documentation and Markdown paths before filesystem access,
reject symlinks in every ancestor and at the leaf, reject hard-linked or
non-regular files, and reject mutation during a read. Paths MUST be normalized
and contained in their declared root; absolute artifact references, traversal
and repository-relative fallbacks are unsupported. Envelope and bundle-index
reads are bounded to 1 MiB each; referenced machine evidence retains its owning
reader's declared size bound. These checks MUST NOT read specifications,
document hashes, adoption status or human review records. Machine evidence
validation does not establish the human fidelity or implementation-claim
prerequisite of REQ-05-002 and PC-006.
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
Claim-bearing publication MUST use the fixture, geometry, seed, traffic,
warm state and sampling policy selected by each current Core 04 predicate.
The environment profile MUST NOT impose large-grid traffic on Evidence-heavy
criteria or add a warm-up to a single observation. The existing owner-controlled
typed fixture and predicate contracts are the executable projection; publication
MUST NOT introduce another product workload or predicate catalog.
Profiles: claim_publication
Verified by: PC-002, PC-003

**REQ-05-013**
Additional benchmark profiles MAY be defined later, but they are informative only unless a Core 05 publication claim manifest or a publication criterion explicitly promotes them to claim-bearing status.
Profiles: claim_publication
Verified by: PC-002

**REQ-05-014**
`cartulary.perf.desktop_ref.v2` is immutable historical evidence: it has the
v1 values with only its profile ID and PostgreSQL image changed to
`cartulary.bench.ubuntu_24_04_postgres_18_6.2026q3`. The exact v1/v2 machine
objects remain recoverable at commit
`00d24aa5354284dfcf6533b8f650e9eb0291ad44`; current execution MUST NOT load a
legacy manifest reader or silently upgrade those objects.

The current `cartulary.perf.desktop_ref.v3` uses the v2 hardware, OS-image,
storage, network, security and warm-state values. Its browser build is
`153.0.8010.12`, its benchmark harness ID is
`cartulary.bench.harness.playwright.v2`, and its harness version is the pinned
Playwright version `1.63.0`. The harness source is separately bound by the
canonical source digest. Headed mode, no browser extensions and 100% zoom
remain mandatory. Product viewport, device scale factor, traffic trace, seed
and warm-up are removed from the global environment object and come from the
selected Core 04 contracts. The machine owner is the closed
`cartulary.benchmark_profile.v1` object for v3, with `schema_id`,
`benchmark_profile_id`, `benchmark_manifest_schema_id`, and `environment`.
The only current manifest generation is v2. Historical profiles remain human
historical references and cannot satisfy a current claim.

Claim-bearing publication MUST compare every environment field by exact value.
Hardware and deployment environment qualification remains part of the retained
publication review; a browser receipt alone does not attest cloud hardware,
OS images, network conditions or implementation conformance. Equivalent hardware,
equivalent browser, or similar LAN language is insufficient.
Profiles: claim_publication
Verified by: PC-002

## 5. Publication claim manifest

The manifest below defines the companion-local publication claim without redefining implementation claim boundaries.

Definition of Done:

- prerequisite claim: relevant implementation claim for each published timed or fixture-sensitive criterion
- requirement selector: `profile:claim_publication`
- required publication criteria: `PC-001..PC-006`

- **PC-001**: A claim-bearing timed or fixture-sensitive result is non-conformant if no `benchmark_manifest` exists, if `benchmark_profile_id` is missing, or if any required field in `cartulary.benchmark_manifest.v2` is absent, or an unknown member is supplied.
  - Verifies: REQ-05-003, REQ-05-006, REQ-05-009
- **PC-002**: A current claim is non-conformant if its profile is historical or unknown, its closed environment differs from `cartulary.perf.desktop_ref.v3`, its actual browser receipt differs from that environment, or its fixture, geometry, traffic, seed, warm state or sampling differs from the current Core 04 predicate. Historical results remain informative only.
  - Verifies: REQ-05-003, REQ-05-006..REQ-05-008, REQ-05-012..REQ-05-014
- **PC-003**: For each timed or fixture-sensitive criterion in the current profile, the result binds to an exact current Core 04 `measurement_predicate_id` and fixture ID and preserves their start state, action, stop predicate, threshold, and sampling mode without local reinterpretation.
  - Verifies: REQ-05-004..REQ-05-005, REQ-05-011..REQ-05-012
- **PC-004**: A claim-bearing benchmark run is non-conformant if authentication, session handling, CSRF protection, sanitization, safe-preview restrictions, or integrity checks are disabled, or if a claim-bearing visible-state result is produced from headless browser mode.
  - Verifies: REQ-05-010
- **PC-005**: The bundle retains the envelope, recursively digested run/measurement/fixture/cleanup artifacts, raw timing samples, Playwright results and actual browser receipts needed to replay and audit the claim. Tampering, aliases, documentation paths, mismatched source, retries, missing bindings, incomplete criteria or invalid measurement qualification reject before a claim can qualify.
  - Verifies: REQ-05-009
- **PC-006**: A claim-bearing benchmark publication is conformant only when every requirement selected by `profile:claim_publication` is implemented, every applicable publication criterion listed in this manifest passes, and the underlying implementation claim already passes for each published timed or fixture-sensitive criterion.
  - Verifies: `profile:claim_publication`
