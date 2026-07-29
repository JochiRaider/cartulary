---
doc_id: cartulary.testing_harness.v2
title: Testing Harness NLSpec
conformance_profile_id: cartulary.testing_harness.current.v2
doc_type: nlspec
status: adopted/current
authority_boundary: Harness mechanics only; command invocation, target selection, scheduling, fixture lifecycle, service ownership, artifact emission, summary emission, cleanup, and harness verification gates.
---

## 1. Status, Scope, and Authority

This NLSpec defines the Cartulary testing harness subsystem. It is the sole
adopted current authority for the harness mechanics identified in
`authority_boundary`, including owner and row selection, catalog validation,
runner adaptation, retained-evidence auditing, and the existing scheduler,
fixture, service, artifact, cleanup, and verification-gate contracts. Adoption
does not make harness readiness evidence product conformance or Core 05
claim-publication evidence.

No harness test, generator, conformance command, or release command reads,
stats, hashes, or otherwise consumes this file or any other documentation file
as executable evidence. Versioned machine inputs are reviewed projections of
this NLSpec and the applicable product/support specifications; executable tools
validate those projections without parsing their source documents.
OpenAPI release change sets are operational review records: each entry contains
only its semantic fingerprint, compatibility classification, owning component,
and rationale. They MUST NOT copy specification requirement identifiers or use
a requirement catalog as approval authority.

**TH-HARNESS-REQ-001**
This NLSpec owns only harness mechanics: command invocation, owner and row
selection, catalog and runner validation, scheduling, fixture lifecycle,
service ownership, artifact emission, summary emission, retained-evidence
auditing, cleanup, and harness verification gates. Core 00 through Core 04
remain the sole owners of product behavior and product-profile authority. Core
05 remains the sole owner of claim-publication and benchmark-publication
activation.

Owner catalogs, verification contracts, schemas, task-surface inputs, and
execution-topology inputs are reviewed derived contracts. They MUST implement
this NLSpec and their cited product or support owners, but they MUST NOT
supersede either source. Migration trackers and handoffs are implementation
authorities only; they MUST NOT create lasting harness behavior. Guides,
generated outputs, repository source, tests, and retained artifacts MUST NOT
become alternate behavior owners.

Frontend readiness mechanics introduced by `browser-e2e-visual`, `browser-e2e-a11y`, `test-evidence-audit`, owner test-family manifests, `tools/frontend_visual_fixture_registry.json`, `cartulary.test_evidence_accounting.v1`, `cartulary.frontend_visual_fixture_registry.v5`, `cartulary.frontend_accessibility_summary.v4`, and `cartulary.frontend_claim_publication_review.v1` are harness and implementation-readiness mechanics only. They MUST NOT define Core product behavior, MUST NOT promote visual or accessibility evidence into product-conformance evidence, and MUST NOT activate Core 05 claim-publication review unless a claim-bearing publication predicate is active.

Harness v1 command IDs, delivery-phase registries, schemas, artifact identities, and retained runs are historical investigation evidence after the v2 cutover. A v2 implementation MUST NOT provide v1 aliases, fallback readers, dual catalogs, dual writers, or newest-artifact fallback. Historical v1 artifacts MUST NOT close a v2 verification or release gate.

Where repository planning, handoffs, or command descriptions use the phrase "production evidence", the current harness interpretation is release-readiness evidence. Release-readiness evidence is a release gate input, not product conformance by itself and not Core 05 claim-publication evidence unless an owner document later promotes a narrower claim-bearing publication boundary.

Harness-local cache mechanics introduced by `cartulary.cache.readiness.v1`, `cartulary.cache.build_artifact.v1`, `cartulary.cache.static_analysis.v1`, `cartulary.agent_finalize_action_cache_record.v1`, and `cartulary.execution_topology_render_cache.v1` are local acceleration mechanics only. They MUST NOT define product behavior, MUST NOT weaken public-target summary emission, MUST NOT replace drift, security, service-readiness, cleanup, runtime-reset, or generated-artifact verdicts, and MUST NOT be cited as product-conformance, release, benchmark, or Core 05 publication evidence.

Same-run helper artifact references introduced by `cartulary.same_run_helper_artifact_ref.v2` are retained harness diagnostics only. They MAY explain that an aggregate consumed helper/setup artifacts produced earlier in the same run, but they MUST NOT mark scheduler work as `reused`, MUST NOT reference old retained runs, MUST NOT skip selected conformance rows, and MUST NOT be cited as product-conformance, release, benchmark, or Core 05 publication evidence.

Fallow static-analysis mechanics introduced by `frontend-fallow-static`, `.fallowrc.json`, `tools/fallow/*`, `cartulary.fallow_reachability_owner.v1`, and `cartulary.fallow_static_summary.v2` are harness and implementation-support mechanics only. They MUST NOT define product behavior, MUST NOT replace TypeScript, Biome, frontend import-boundary checks, tests, security scans, generated-artifact drift checks, or harness gates, MUST NOT activate Fallow Runtime behavior, and MUST NOT be cited as Core 05 publication evidence.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-016, TH-HARNESS-AC-022, TH-HARNESS-AC-026, TH-HARNESS-AC-028

**TH-HARNESS-REQ-002**
A harness conformance claim MUST identify this NLSpec version, the exact public Make target or target set under evaluation, the conformance environment from Section 14, and the retained result root/run ID/run root when retained harness artifacts are used as evidence.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-016

**TH-HARNESS-REQ-003**
The current canonical public invocation surface is Make. In the current profile, a command invocation is canonical only when invoked as `make <target>` from the repository root or through a Make-owned wrapper that preserves the target identity.

Each public target MUST also declare one stable `command_id` using the form `cartulary.harness.command.<name>.v<positive_integer>`. The version is a decimal integer without a leading zero. The `command_id` identifies the command's semantic contract. The Make target name is the current invocation binding for that semantic command. A later adopted NLSpec MAY add additional invocation bindings only when they preserve the same `command_id`, configuration contract, output contract, artifact contract, failure mapping, and cleanup behavior.

Public behavior change gate. Any change to public Make target identity, stable `command_id`, declared schema IDs, retained artifact paths, output shape, failure mapping, or task-surface/topology metadata MUST be specified here before implementation and before generated task-surface or topology outputs are refreshed. Private module movement that preserves those public contracts MAY proceed without changing this gate.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005

**TH-HARNESS-REQ-004**
Generated files under `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, generated task/schedule artifacts, and generated Make includes are downstream generated artifacts. They MUST NOT be hand-edited and MUST NOT become behavior owners unless a later adopted NLSpec explicitly promotes one of them. `tools/task_surface_owner.json` is the authored machine owner for task metadata and Make binding profiles; `tools/task_surface_manifest.json` and generated Make includes are projections of that owner. `tools/execution_topology_manifest.json` owns execution topology and the closed runtime, resource, and fixture profile registries only; it MUST NOT embed a second task-surface or test-catalog owner.

Standalone generated-artifact renderers that produce service-backed schedule source files MUST write only to an explicit caller-supplied output path. `tools/harness/generated-artifacts/render-service-backed-schedule-manifest.mjs` MUST NOT create an implicit repo-local `tools/scheduler_service_sources.json`; Make-owned generation is the only current owner for checked-in scheduler and topology artifacts derived from service-backed schedule sources.

When a harness helper or catalog path change affects task-surface, topology, or schedule outputs, the owner input MUST be updated first and downstream artifacts MUST be refreshed through `make generate`. The current verification ladder is `make generate-drift`, `make generated-artifact-policy-check`, and `make json-shape-check`. Generated files MUST NOT be edited as the source of truth for helper-path or catalog migration.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-016

**TH-HARNESS-REQ-005**
Direct package scripts, raw scripts, raw Go/Vitest/Playwright/Biome/Vite/pnpm commands, and tool-specific reports are developer conveniences or child commands unless a public Make target invokes them. Direct invocation of those surfaces MUST NOT be treated as equivalent to a canonical harness run.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-005

## 2. Purpose, Non-Goals, and Conformance Boundary

The testing harness exists to provide a reproducible repository command surface for local developers, CI entrypoints, coding agents, and release verification. It provides deterministic target selection, bounded output, structured artifacts, explicit service ownership, controlled fixture lifecycle, stable failure classification, and destructive cleanup gates.

**TH-HARNESS-REQ-006**
The harness MUST provide all of the following for public Make targets:

- deterministic target-class and target-selection metadata;
- declared configuration resolution;
- output-mode behavior;
- exit-code mapping;
- retained artifact identity when a target declares artifacts;
- failure classification that separates product assertion failures from harness operational failures;
- cleanup predicates for every destructive operation.
Verified by: TH-HARNESS-AC-001..TH-HARNESS-AC-017

**TH-HARNESS-REQ-007**
The harness MUST NOT claim provider-specific hosted CI behavior, benchmark publication, release publication readiness, macOS support, Windows-native support, Podman support, or Playwright artifact schema stability unless those areas are explicitly included in this NLSpec's current conformance profile or in a later adopted NLSpec. The current visual-snapshot refresh authority is limited to the helper-only maintenance contract stated in Sections 6, 8, 11, 15, and 17.
Verified by: TH-HARNESS-AC-012, TH-HARNESS-AC-016

**TH-HARNESS-REQ-008**
Logical scheduler resources are execution constraints inside the harness. They MUST NOT be represented as guarantees about physical CPU, I/O, Docker, database, object-store, browser, or network capacity.
Verified by: TH-HARNESS-AC-006

**TH-HARNESS-REQ-009**
Adopted Sections 1 through 17 MUST close every current-conformance harness behavior they name. Unbounded delegation to a target, producer, tool, implementation, or applicability judgment is invalid unless the same requirement cites a closed table, schema attachment, algorithm, or explicitly non-normative diagnostic boundary. Generated manifests and generated Make includes MAY mirror a closed contract, but they MUST NOT be the only current-conformance owner for a public harness behavior.

In adopted Sections 1 through 17, `MAY` means true implementation freedom whose divergent realizations remain interchangeable to callers. Acceptance-bearing behavior MUST use `MUST` or `MUST NOT`; this document does not use an advisory normative keyword.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-029

**TH-HARNESS-REQ-010**
The normative main body in Sections 1 through 17 MUST contain every load-bearing default, bound, omission rule, mapping, algorithm, failure consequence, and acceptance obligation. Appendices and Section 18 MAY contain rationale, examples, sample objects, and research traceability only. A supporting section MUST NOT be the sole owner of current-conformance behavior.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-071

**TH-HARNESS-REQ-011**
The current normative imported format dependencies are JSON Schema Draft 2020-12, RFC 8785 JSON Canonicalization Scheme with its verified errata, SHA-256 as specified by FIPS 180-4, and RFC 3339 timestamps constrained by Section 8. A later revision MUST replace an imported dependency explicitly; implementation-library defaults MUST NOT silently change the format contract.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-066, TH-HARNESS-AC-071

**TH-HARNESS-REQ-021**
Repository test-harness observability is a harness diagnostic subsystem, not
application runtime telemetry. This NLSpec owns harness trace reconstruction,
harness performance metrics, retained diagnostic artifacts, explicit post-run
OTLP export, and the public commands that inspect or export them. The pinned
OpenTelemetry source and protocol baseline is consumed from machine snapshot
and dependency contracts under `contracts/otel/`; this reference MUST NOT
duplicate or silently rebaseline it. Application telemetry scopes, application
`cartulary.module` attributes, deployment telemetry configuration, and the
server telemetry bootstrap MUST NOT become harness configuration or harness
evidence owners.
Verified by: TH-HARNESS-AC-072, TH-HARNESS-AC-076

## 3. Terminology

| Term                     | Meaning                                                                                                                                                    |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| harness run              | One invocation of a canonical harness command or one child invocation explicitly tied to a result root and run ID.                                         |
| target                   | A named Make target or scheduler target selected by the harness.                                                                                           |
| public target            | A target classified as public in the public command registry and canonical only through Make.                                                              |
| child target             | A target invoked by an aggregate, sequence, scheduler, or wrapper target.                                                                                  |
| work unit                | One scheduler-visible executable unit with an identity, dependencies, resource claims, logs, status, and optional completion keys.                         |
| scheduler                | A harness runner that executes a manifest-defined DAG using logical resource claims and emits scheduler events and summaries.                              |
| lifecycle machine        | A normative finite state contract for one harness lifecycle, including its closed states, closed events, allowed transitions, failure mapping, and evidence. |
| representational lifecycle diagram | A non-normative diagram or list that explains an existing lifecycle without adding requirements.                                                   |
| state                    | One named lifecycle condition inside a lifecycle machine.                                                                                                  |
| event                    | One named input signal that can be presented to a lifecycle machine.                                                                                       |
| transition               | One allowed movement from a source state to a destination state for a specific event and guard.                                                            |
| terminal state           | A state that ends a lifecycle machine instance. No later transition is allowed from a terminal state.                                                      |
| result root              | The root directory that contains run artifacts. The default is `.cartulary/test-results`.                                                                  |
| run ID                   | The run directory name under the result root. The default format is defined in Section 6.                                                                  |
| run root                 | The directory `normalize_result_root(CARTULARY_TEST_RESULTS_DIR) / normalize_run_id(CARTULARY_TEST_RUN_ID)`.                                               |
| harness artifact         | A file or directory produced by a harness run, child command, service, scheduler, test runner, or diagnostic tool.                                         |
| retained artifact        | An artifact preserved after command exit for a specific result root, run ID, and target.                                                                   |
| release-readiness evidence | A retained harness artifact that aggregates product, frontend, visual, accessibility, harness, release, and support evidence for release gating while preserving each record's owner-derived semantic effect. |
| generated artifact       | A file produced from owner inputs by a generator and checked for drift.                                                                                    |
| cache record             | A repo-local JSON record that proves a cache key, input digest, output digest, and profile-specific output contract for a local acceleration profile.      |
| cacheable output         | A deterministic file or directory output whose digest is declared in a cache record and may be reused only when all profile validation succeeds.           |
| non-cacheable side effect | Target behavior that must still execute or be emitted for the current run, such as summaries, failure classification, cleanup, service readiness, or drift/security verdicts. |
| same-run helper artifact ref | A schema-owned retained reference that links an aggregate consumer to helper/setup artifacts produced earlier under the same run root, with declared inputs, producer artifact digests, consumer refs, and fail-closed scope. |
| fixture                  | Test setup state created for a test, package, target, scheduler group, browser stack, or service suite.                                                    |
| service-backed fixture   | A fixture that uses Postgres, object-store services, Docker/testcontainers, browser processes, or Compose-backed services.                                  |
| backing services         | Postgres, object-store services, Docker/testcontainers, Compose services, backend processes, frontend processes, and browser runtime dependencies used by harness targets. |
| output mode              | The resolved mode from Section 7 that controls stdout, stderr, and artifact summary behavior.                                                              |
| machine output           | The `machine` output mode defined in Section 7. For public Make targets that accept it, stdout is exactly one UTF-8 JSON object followed by LF.            |
| failure class            | A coarse normalized grouping for failed harness commands: `product`, `config`, `infra`, `harness`, `artifact`, `timing`, `interrupted`, or `unknown`.      |
| failure reason           | A detailed snake-case reason code used for diagnostics, exit-code mapping, automation, and handoff.                                                        |
| cleanup tier             | A named cleanup scope such as repo-local clean, repo-local distclean, service-suite cleanup, browser-stack cleanup, or stale janitor cleanup.              |
| stale janitor            | A cleanup routine that removes previously generated DBs, buckets, containers, or browser fixtures only when proof predicates match.                        |
| diagnostic-only artifact | An artifact retained for human investigation whose internal shape is not a machine-readable harness conformance contract.                                  |
| harness observability bundle | A deterministic diagnostic projection of native retained harness evidence into one invocation trace, OTLP request payloads, and one hotspot summary. |
| invocation trace         | One trace rooted at a top-level public harness invocation. Child targets, sequence steps, scheduler work, services, runners, and finalizers are spans or links inside that trace. |
| actual dependency critical path | The longest observed dependency-respecting chain of executable and wait intervals. It is distinct from the scheduler envelope stored in `critical_path_wall_duration_ms`. |
| unattributed envelope    | Parent wall time not covered by the union of its directly attributable child intervals; overlapping child time is counted once. |
| test owner               | The one module, platform, application, package, or harness boundary accountable for a catalog row's verification postcondition.                         |
| test family              | An owner-qualified semantic grouping of related catalog rows; it is not a runner, file, target, evidence class, or delivery milestone.                 |
| catalog row              | One active cross-runner executable evidence contract with one owner, exact selectors, verification references, and execution profiles.                 |
| collaborator             | A participating owner that is not accountable for the row's verification postcondition.                                                                 |
| verification contract    | A reviewed machine-readable derivation of one adopted product or support requirement used for evidence routing.                                         |
| runner adapter           | An allowlisted harness implementation that translates one closed selector kind into executable child work and normalized results.                      |
| runtime profile          | A closed startup and managed-service identity referenced by a catalog row.                                                                               |
| resource profile         | A closed set of logical scheduler-resource claims referenced by a catalog row; it does not define physical capacity.                                    |
| fixture profile          | A closed fixture-policy and budget identity referenced by a catalog row.                                                                                 |
| retired delivery identity | Any ordinal execution identity from the predecessor harness, unsupported by the current owner-based runtime.                                           |

Domain and product terms keep their meanings from the product specs and `docs/domain.md`.

### 3.1 Identifier grammar and lifecycle

**TH-HARNESS-REQ-012**
The current identifier grammar is:

```text
segment         = [a-z][a-z0-9_]{0,62}
owner_id        = (module|platform|app|web|package|harness) "." segment
family_id       = owner_id "." segment
row_id          = family_id "." segment
verification_id = owner_id ".verification." segment
```

Each complete identifier MUST contain at most 191 ASCII bytes. Unicode, whitespace, empty segments, `/`, `\\`, percent escaping, and shell metacharacters are invalid. No segment may case-insensitively match `phase[0-9]+` or `fe_p[0-9]+`. Serialized identifier arrays MUST be unique and ordered by ascending ASCII bytes.

Identifiers are immutable, globally unique within their category, and never recycled. An owner migration MUST allocate a new ID and record the migration crosswalk; a runtime alias is forbidden. When display metadata is omitted, diagnostics MUST render the machine ID. Display metadata MUST NOT participate in semantic identity or semantic digests.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-068

### 3.2 Owner, verification, and family contracts

**TH-HARNESS-REQ-013**
The current owner inputs are `contracts/verification/registry.json`,
`contracts/verification/owners/*.json`, `tools/test_catalog_owner.json`, and
`tools/test_families/*.json`. They validate respectively as
`cartulary.verification_registry.v3`, `cartulary.verification_contract.v3`,
`cartulary.test_owner_registry.v1`, and `cartulary.test_family_manifest.v2`.
Every schema uses JSON Schema Draft 2020-12, requires its exact `schema_id`,
rejects unknown properties, and closes every current enum.

Every verification resolves to an active catalog row or registered public
target. Verification v3 contains no requirement, acceptance,
specification-trace, or specification-status field. Specification completeness
is assessed against the adopted owner by human review, not inferred from
routing counts.

An active test owner registry row MUST contain `owner_id`, `manifest_path`, and
`status="active"`; it MAY contain display metadata. A verification registry
owner row contains only `owner_id` and `contract_path`. Each path MUST be a
normalized repository-relative path under the matching owner root, MUST
resolve exactly once, and MUST remain inside the repository after realpath
resolution. Every active test owner MUST own at least one active executable
row.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-067

**TH-HARNESS-REQ-014**
Every active catalog row MUST contain `row_id`, `owner_id`, `family_id`, `collaborator_ids`, `verification_ids`, `runner`, `selector`, `evidence_class`, `runtime_profile_id`, `resource_profile_id`, `fixture_profile_id`, `default_check`, `claim_posture`, and `status="active"`.

`collaborator_ids` is required and MAY be empty. `verification_ids` is required
and MUST be nonempty. Verification entries contain only routing semantics:
`verification_id`, `behavior_class`, `profile`, `evidence_kinds`, optional
`public_target`, and optional `skip_policy`. Reference arrays MUST be sorted and
duplicate-free, and every reference MUST resolve exactly once. `owner_id` MUST
equal the containing manifest owner. A row MUST NOT embed commands, ports,
capacities, service topology, child environment variables, fixture paths,
documentation paths, or document-derived behavior.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-067

**TH-HARNESS-REQ-015**
Verification contracts use the closed `behavior_class` set `product`, `security`, `architecture`, `build`, `harness`, and `claim_publication`. Catalog rows use the closed `evidence_class` set `unit`, `integration`, `browser`, `accessibility`, `visual`, `measurement`, `static`, `security`, and `release`.

Verification profiles are `base`, `support`, `extension.<profile_id>`, or `claim.<profile_id>`. Claim posture is `implementation`, `informative`, or `claim.<claim_id>`. Product and security behavior MUST use `base` or an adopted extension; build and harness behavior MUST use `support`; claim-publication behavior MUST use an active Core 05 `claim.*` profile. A `claim.*` posture and profile MUST resolve to the same active claim. Informative evidence MUST NOT satisfy conformance or release closure, and informative measurement rows MUST set `default_check=false`.

The default skip policy is `forbid`. An authorized skip MUST identify one closed reason, the verification owner, approval evidence, and an RFC 3339 expiry. Missing or expired authorization is equivalent to `forbid`.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-065, TH-HARNESS-AC-069

### 3.3 Runner registry and selectors

**TH-HARNESS-REQ-016**
`tools/test_runner_registry.json` MUST validate as `cartulary.test_runner_registry.v1`. The current runner set and selectors are closed by this table:

| Runner | Exact selector contract |
| --- | --- |
| `go` | One repository package and one nonempty ASCII-sorted array of exact top-level `Test...` symbols. |
| `vitest` | One repository-relative test file and one nonempty ASCII-sorted array of exact full test titles. |
| `playwright` | One repository-relative file, one project ID, one stage, nonempty stable scenario IDs, and matching diagnostic titles. |
| `shell` | One stable registered command ID; raw shell, argv, executable paths, and row-defined environment are forbidden. |

Playwright stage is exactly one of `webserver_backed`, `stateful`, `support`, `visual`, `accessibility`, or `measurement`. Before setup, selector validation MUST reject zero resolution, multiple resolution, overlap between active rows, globs, regular expressions, missing paths, symlink escape, paths outside approved roots, and shell command IDs absent from the task-surface registry.

A later runner requires an adopted runner-registry and NLSpec revision, selector and result schemas, an allowlisted checked-in adapter, and positive and negative contract fixtures. Dynamic package, plugin, or executable loading is outside the current profile.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-067

### 3.4 Runtime, resource, and fixture profiles

**TH-HARNESS-REQ-017**
`tools/execution_topology_manifest.json` MUST own closed top-level `runtime_profiles`, `resource_profiles`, and `fixture_profiles` collections. Catalog rows reference profile IDs only. Resource profiles MAY reference logical resources from `tools/scheduler_resource_registry.json` but MUST NOT redefine capacity. Unknown profiles, duplicate profiles, cross-kind references, and inline row overrides MUST fail before child work.

The current runtime profiles are:

| ID | Managed services required | Contract |
| --- | ---: | --- |
| `none` | no | No managed service or browser startup. |
| `default` | yes | Ordinary unclaimed isolated test-service/browser configuration. |
| `network_flow_claimed` | yes | Network Flow claimed startup configuration with its separately owned key-ring and secret-handling rules. |

The current resource profiles are `none`, `go_balanced`, `go_cpu_heavy`, `go_io_heavy`, `go_transaction_heavy`, `go_reset_heavy`, `go_clone_heavy`, and `browser_exclusive`. The current fixture profiles are `none`, `postgres_transaction`, `postgres_package_reset`, `postgres_group_clone`, `postgres_template_clone`, `postgres_migration_scratch`, `object_store_isolated`, and `service_stack`. Each profile's exact claims, budgets, and compatibility keys MUST be present in the authored topology; omission has no implicit fallback except the explicit `none` profile.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-065

### 3.5 Browser runtime profiles

**TH-HARNESS-REQ-018**
Browser evidence MUST obtain `runtime_profile_id` through the catalog row and the authored execution-topology registry. Unknown profiles MUST fail before Playwright starts. Arbitrary per-test environment injection and runtime routes that toggle extension claims are forbidden.

A browser runtime profile is immutable startup identity. Generated groups, shards, and sessions MUST carry the profile ID; incompatible profiles MUST use distinct browser session groups. A mixed-profile session, an attach request whose expected profile differs from the retained stack, or a profile/configuration fingerprint mismatch MUST fail before product assertions. Runtime reset is data-only and MUST NOT change the profile, extension claims, key-ring identity, or child-process environment.

Claimed-profile secrets MUST be generated in memory for each owned stack, passed only in the child server environment, and redacted from commands, logs, diagnostics, summaries, retained metadata, and failure messages. Retained browser-stack metadata MUST contain the non-secret runtime profile ID and deterministic configuration fingerprint and MUST NOT contain secret values or secret digests.

Every Make-owned browser invocation MUST resolve its selected rows and their
runtime profiles before service setup. A resolved browser session with one or
more managed services has `service_requirement=test-services` and MUST either
own one isolated suite or attach to an exact compatible active suite. A resolved
session with no managed services has `service_requirement=none` and MUST start
no service suite. Incompatible runtime profiles MUST remain separate sessions,
including when one public target selects more than one profile. Target or stage
names MUST NOT determine service need, and no browser path may fall back to the
shared development Postgres, object store, Compose project, bucket, port, or
proxy.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-066

### 3.6 Semantic digests and extension boundary

**TH-HARNESS-REQ-019**
Semantic JSON digests MUST reject duplicate object members, non-I-JSON numbers, non-finite values, and negative zero before RFC 8785 canonicalization. The digest is lowercase `sha256:` followed by 64 hexadecimal characters. Semantic projections MUST omit display metadata, inert documentation references, diagnostic timestamps, and other fields explicitly classified as non-semantic by the owning schema. A producer and consumer MUST use the same schema ID and semantic projection version.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-066

**TH-HARNESS-REQ-020**
Internal module layout, in-memory data structures, adapter function signatures, human diagnostic prose, and scheduling order among simultaneously ready independent work units remain intentionally unspecified when Sections 4 through 17 are preserved. Implementations MUST remain interchangeable at the public command, schema, artifact, failure, security, and cleanup boundaries.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-071

## 4. Public Command Surface

**TH-HARNESS-REQ-050**
The public command registry MUST be owned by this NLSpec, represented in the authored `tools/task_surface_owner.json`, and mirrored by generated `tools/task_surface_manifest.json` entries with `target_class="public"`. The implementation MUST provide exactly the public targets listed in the target registry below unless the owner input and this NLSpec are revised together.

`tools/execution_topology_manifest.json` MAY provide scheduler topology, child-work topology, generated schedule inputs, or resource-profile inputs. It MUST NOT independently add, remove, rename, reclassify, or change the output class, artifact policy, schema policy, side-effect declaration, command identity, or public lifecycle state of a public target.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-027

**TH-HARNESS-REQ-051**
Every public target MUST declare exactly one output class, exactly one stable-summary schema policy, and exactly one artifact policy. The output-class behavior is owned by Section 7. The schema policy is owned by Section 8.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005, TH-HARNESS-AC-023

**TH-HARNESS-REQ-052**
A Make-owned wrapper MAY invoke package scripts, raw scripts, or external tools as implementation mechanisms. The wrapper remains responsible for the public target's configuration, output, artifact, failure, exit-code, and cleanup contract.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004

**TH-HARNESS-REQ-058**
A public target MUST provide a semantic harness operation, not merely an alias for one or more child commands. A target qualifies as semantic only when it owns at least one observable behavior from the table below in addition to invoking child work.

| Semantic behavior | Observable requirement |
| --- | --- |
| `configuration_resolution` | Resolves and validates declared harness configuration before child work. |
| `evidence_normalization` | Emits or validates retained artifacts under a stable schema. |
| `failure_normalization` | Maps child or harness failures to Section 9 `failure_class`, `failure_reason`, and public exit code. |
| `service_lifecycle` | Owns service startup, readiness, fixture lifecycle, lease, or cleanup proof. |
| `scheduler_orchestration` | Selects, orders, and executes work units using the scheduler contract. |
| `destructive_safety` | Applies cleanup or reset proof predicates before mutation. |
| `security_boundary` | Applies redaction, token gating, artifact-safety, or secret-handling behavior. |
| `diagnostic_synthesis` | Converts retained evidence into a bounded human or machine diagnostic that cannot be obtained from raw child output alone. |

A target that provides none of these behaviors MUST be private child work or a developer convenience outside the public command registry.
Verified by: TH-HARNESS-AC-020

**TH-HARNESS-REQ-053**
The default `check-harness-smoke` gate MUST remain a small semantic smoke surface rather than a broad harness regression suite. Its fast tier MUST contain exactly one check for each gate role: public Make/wrapper projection, check-scheduler semantics, and service-backed scheduler semantics. Broader field-shape, topology-rendering, and sequence-detail checks MUST live in owner-aligned validation such as `json-shape-check`, generated drift checks, the explicit `harness-contract` extended target, or non-default diagnostic smoke tiers. The `harness-contract` target MUST be selected by CI and release gates, not by default local `check`.

Fast smoke fixtures that create disposable repo-shaped files, directories, fake Make surfaces, manifests, or child-run workspaces MUST create them through `tools/harness/test-support/harness-scratch.sh` outside the repository checkout. `CARTULARY_HARNESS_SCRATCH_ROOT` MAY redirect that scratch root only when it still resolves outside the repository. Repo-local `tmp/` remains reserved for durable tool caches, retained run artifacts, and operator-inspectable local outputs; fast smoke fixtures MUST NOT place transient package-shaped or source-shaped scratch trees there, so concurrent source traversal such as `go list ./...` cannot observe disappearing non-package directories.

Non-default harness smoke tiers MAY carry owner-specific regression checks for harness maintenance surfaces such as finalization, evidence audit, topology generation, scheduler behavior, and wrapper compatibility. A retained harness self-test MUST be reachable from an owner-controlled tier, merged into another active owner test, or deleted with named replacement coverage; manual-only harness self-test files outside owner-controlled tiers MUST NOT be treated as active coverage.

The current owner-controlled harness smoke tiers are `fast`, `execution`, `extended`, `lifecycle`, and `full`. Make helper targets MUST expose the active tiers as `run-harness-smoke-fast`, `run-harness-smoke-execution`, `run-harness-smoke-extended`, `run-harness-smoke-lifecycle`, and `run-harness-smoke-full`. The execution helper target is the narrow validation surface for shell command execution, runner wrappers, Make-node dispatch, service-backed runner delegation, and fast Make-sequence wrapper behavior. The lifecycle helper target is the narrow validation surface for browser/dev stack lifecycle, reset, readiness, and teardown harness changes. The execution and lifecycle helper targets MUST remain helper-only and MUST NOT become default local `check` work by themselves.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-006

**TH-HARNESS-REQ-054**
The default local `check` gate MUST prioritize correctness evidence and MUST NOT enforce duration-baseline drift. Duration-baseline coverage MAY remain in `check` because it validates scheduler input completeness. Duration-baseline drift MUST remain available through explicit duration drift targets and MUST be enforced by CI, `agent-finalize RESULTS_DIR=<dir>`, or another timing-maintenance surface.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-006, TH-HARNESS-AC-027

Default `check` placement MUST be selected from the active owner catalog. `default_check=true` selects a row for `make check`; `default_check=false` excludes it from `make check` but never narrows an omitted owner slice. Informative measurement rows MUST default to `false`. Every selected row MUST retain its verification, evidence class, runtime profile, resource profile, and fixture profile. Task-surface default inclusion and scheduler reachability MUST agree with the generated evidence-to-gate mapping in TH-HARNESS-REQ-667. An unknown row, inactive row, missing verification, or row selected by a gate inconsistent with its evidence class MUST fail catalog validation before child work.

Scheduled Go shard identity MUST be resolved against the exact catalog-row universe used to generate that schedule. Every generated Go shard and its aggregate finalizer MUST receive the same closed selection scope: `all`, `default_check`, or `rows`. The `rows` scope additionally requires a sorted, duplicate-free, non-empty semantic row-ID set; the other scopes forbid row-ID payloads. Capture, fixture planning, report collation, and manifest reconciliation MUST derive one exact row universe from that atomic tuple and reconstruct shard names only from that universe. A shard name from a default-check, owner, or other bounded projection MUST NOT be reinterpreted against the full target inventory. A partial, missing, unknown, duplicate, broadened, or inconsistent scheduled selection MUST fail before Go product work or artifact collation rather than execute a different positional shard.

**TH-HARNESS-REQ-055**
The check scheduler MUST NOT skip required correctness work through digest-only input stamps in the current local profile. Default local `make check` MUST execute every selected static, drift, security, product, service-backed, browser, service-resource mutation, runtime reset, and scratch database apply work unit unless a future NLSpec revision defines a reusable artifact cache with complete retained provenance.

`local_input_stamp`, `CARTULARY_CHECK_DISABLE_INPUT_STAMPS`, `tmp/check-stamps/`, and `cartulary.check_input_stamp.*` records are retired in the current profile. Scheduler manifests and execution-topology owner inputs MUST reject `local_input_stamp` rather than treating it as diagnostic metadata. A future reusable artifact cache MAY classify work as `reused` only when the cache record validates the relevant tracked and untracked inputs, tool versions or binary digests, configuration inputs, expected outputs, summary schema, and artifact references needed to diagnose the reused work from a single retained run.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-006, TH-HARNESS-AC-028

**TH-HARNESS-REQ-060**
The current profile permits only these local acceleration cache families:

- `cartulary.cache.readiness.v1` for toolchain and install readiness, including pinned Node/pnpm readiness, frontend install readiness, Playwright install readiness, pinned Go helper binaries, ShellCheck, test-service image readiness, and scheduler readiness units that only validate provisioning state.
- `cartulary.cache.build_artifact.v1` for deterministic build artifacts, including `build-server`, `build-server-harness`, `build-migrate`, `build-operator`, `build-web`, `testservices-build`, and embedded web asset preparation.
- `cartulary.cache.static_analysis.v1` for deterministic non-security static analysis success stamps, currently limited to `lint-markdown` and strict `lint-shell`.
- `cartulary.agent_finalize_action_cache_record.v1` for the closed `agent-finalize` action IDs listed in Section 8.2.
- `cartulary.execution_topology_render_cache.v1` for internal deterministic execution-topology render content only.

Embedded web asset preparation is a build-artifact producer in the current profile. When the embedded output is consumed by Go `//go:embed`, the publisher MUST update the embedded source-tree artifact atomically from Go embed's point of view. It MUST NOT delete or rewrite a directory of hashed frontend assets in place while concurrently scheduled Go compilation can traverse that directory. Harness-only readiness stamps, cache records, and other operational metadata MUST remain outside the embedded content root unless the owning product spec explicitly makes that file served application content.

Every cached Go binary whose transitive package closure consumes the embedded web asset root MUST depend on the complete embedded asset producer tuple before compilation starts and MUST include that root in its build-artifact cache key. The authored execution topology MUST represent the same dependency whenever the producer and consumer are scheduled together. A source-file-only key, ambient Make ordering, or concurrent publisher/consumer execution is not valid evidence for such a binary.

All permitted cache families MUST be content-addressed. A valid key MUST include the cache schema ID, cache scope, profile ID, platform identity where relevant, declared tool or runtime versions, declared command/profile inputs, helper implementation digests, and every declared output contract needed by the profile. Broad timestamp-only caching is not a valid harness cache mechanism.

Cached Go build-artifact profiles MUST pass `-buildvcs=false`. Git revision and
dirty-worktree stamping are undeclared repository-state inputs and MUST NOT alter a
binary whose cache key is otherwise unchanged. Source-snapshot, release, and audit
provenance remain owned by their retained harness evidence; a cached binary MUST NOT
silently acquire a second provenance identity from ambient Git metadata.

Readiness, build-artifact, and static-analysis cache hits MAY skip only the deterministic provisioning, build, or analyzer command guarded by the cache profile. Static-analysis cache hits MUST be limited to closed non-security profiles whose keys include tool version or binary digest, configuration digests, selected input digests, helper implementation digests, analyzer arguments, and output stamp digests. They MUST NOT skip the public target wrapper, public target summary emission, failure classification, output validation, drift comparison, security scan execution, service lifecycle, service readiness, fixture cleanup, runtime reset, scratch database apply, object-store mutation, destructive-operation guard, or aggregate success/failure computation. Build-artifact and static-analysis cache hits inside `make check`, `make ci`, or `make release-check` MUST NOT be reported as scheduler `reused` work in the current profile.

Test-service image readiness is a readiness cache profile only. Its cache key MUST include the testservices binary digest, image-owner source digests, helper implementation digests, and toolchain pins. A cache hit MUST still prove that every pinned service image named by the testservices helper is locally present before accepting the readiness stamp; missing images invalidate the stamp and force the ordinary image warmup command. This cache profile MUST NOT replace service startup readiness, test service lifecycle evidence, browser reset, cleanup, fixture preparation, or product-conformance evidence.

Cache records that are missing, disabled, forced, invalid, corrupt, or whose declared outputs are missing or digest-mismatched MUST NOT produce success by reuse. They MUST either execute the underlying command and emit a miss/disabled/invalid-cache artifact, or fail as `configuration_error` or `artifact_error` when the target cannot prove the output contract. Security targets, drift verdicts, generated-artifact drift detection, service-backed/browser/live-state tests, cleanup targets, destructive reset targets, and aggregate `check`/`ci`/`release-check` success MUST remain uncached.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-006, TH-HARNESS-AC-018, TH-HARNESS-AC-019, TH-HARNESS-AC-024, TH-HARNESS-AC-028

**TH-HARNESS-REQ-070**
The current `frontend-fallow-static` profile MUST build an effective Fallow configuration from `.fallowrc.json` plus the schema-validated owner input `tools/fallow/reachability_owner.json` using schema ID `cartulary.fallow_reachability_owner.v1`. The effective configuration MUST be retained under the target's run root as `frontend-fallow-static/fallow/resolved-fallowrc.json`, and every Fallow child invocation in the target MUST use that retained configuration. The four authored read-only report invocations (`dead-code` JSON/SARIF, `dead-code` Markdown, `dupes`, and `health`) MUST start as independent concurrent children after the effective configuration is retained. The parent MUST drain every child before normalization, retain every successfully produced artifact when siblings fail, and normalize reports, logs, warnings, and failures in that authored order so concurrent completion order cannot change output identity or primary-failure selection.

Fallow reachability MUST represent durable owner patterns, not per-file suppression growth. Current owner patterns are Vitest setup files selected by `apps/web/vite.config.ts`, task-surface and harness-check backing scripts declared by `tools/task_surface_owner.json`, Vite public asset URLs under the declared public root, and executable tooling dependencies invoked by owner scripts such as `pnpm exec cdxgen`. A missing owner-declared file, missing public asset reference, invalid owner input, or missing executable dependency declaration MUST fail `frontend-fallow-static` as `configuration_error`. Non-blocking static findings in valid Fallow reports MAY still be retained as warnings.

The effective Fallow configuration MUST NOT enable Fallow Runtime, Fallow security scans, baseline enforcement, automatic source mutation, generated-file mutation, inline suppression insertion, or lockfile mutation. A future expansion of changed-code scoping, runtime coverage, blocking enforcement, or additional dynamic reachability owner families MUST revise this NLSpec and the owner schema before implementation.
Verified by: TH-HARNESS-AC-022, TH-HARNESS-AC-028, TH-HARNESS-AC-029

**TH-HARNESS-REQ-056**
The default local `check` gate MUST keep ordinary browser measurement evidence out of the warm `check-service-backed` critical path. `browser-e2e-measurement` MUST remain available as an explicit public target and MAY remain required by CI, release, or explicit browser aggregate targets, but default local `make check` MUST NOT schedule the measurement browser stage as a `check-service-backed` child.

Default local `make check` MUST NOT schedule current full visual or accessibility browser work. Direct `make browser-e2e-visual`, direct `make browser-e2e-a11y`, explicit browser aggregates, and CI/release profiles that select full visual or accessibility evidence MUST retain their full-fidelity target behavior. A later bounded visual or accessibility readiness projection MAY enter default `check` only after this NLSpec and the task-surface metadata declare a separate bounded-readiness row with `check_projection.mode=projection`, `full_target_equivalent=false`, and a reason that identifies the downstream default-check consumer.

`test-slice` and `service-backed-test-slice` select visual work only from the resolved owner rows whose evidence class is `visual` and whose Playwright stage is `visual`. The ordinary `browser-e2e-visual` adapter MUST receive the exact resolved row IDs. It MUST NOT broaden the selection to other owners or infer selection from a filename, title prefix, historical registry, or visual fixture identifier.

Browser rows selected by default `check` MUST declare `default_check=true`, and their verification contract MUST identify the cross-stack postcondition unavailable from a cheaper required gate. Informative measurement, full visual, and full accessibility rows MUST use `default_check=false` unless a later NLSpec revision adopts a separate bounded readiness row with its own verification ID and selector. Direct public browser targets MUST remain full-fidelity for their catalog-selected inventories. A bounded default-check row MUST NOT be represented as proof that a different full target inventory ran.

Default browser schedule generation MUST apply `default_check` after evidence-class and profile resolution and before work-unit emission. An empty browser group MUST be omitted from the default schedule. A group that retains rows MUST preserve reset, taint, browser-session, teardown, and target-summary behavior. Direct public browser targets, explicit `test`, `ci`, `release-check`, browser aggregates, and owner slices MUST use their own closed selection scopes and MUST NOT inherit `make check` filtering.

Default browser projection artifacts MUST expose the exact selected owner rows. Retained `cartulary.test_evidence_accounting.v1` artifacts MUST use the plan's `resolved_row_ids` without broadening. A target-level browser run MUST account for its catalog-selected target inventory in the uniform `<target>/owners/<owner-id>/` shard layout, and its tool-run summary MUST reference the target's `cartulary.browser_owner_index.v1`; absence of one runner family MUST NOT be represented as success for a row that was selected but did not execute.

A scheduler- or owner-slice-selected browser group MUST apply its exact sorted
row-ID subset after resolving the generated batch group and before constructing
Playwright title filters. The subset MUST be non-empty, unique, and contained by
the generated group. Reopening a full batch group by stage and group name MUST NOT
discard, broaden, or replace the scheduler-owned row selection.

`cartulary.browser_e2e_batch_manifest.v7` MUST be generated from the active catalog plus authored stage/runtime/fixture/isolation policy. Every generated group MUST contain exactly one selector file, one runtime-profile-derived `service_requirement`, and a sorted non-empty set of semantic catalog row IDs; delivery-phase IDs, phase-selected batches, title-prefix inference, runtime translation of retired IDs, and renderer-owned dependency lists are forbidden. Across a direct evidence target, every applicable Playwright catalog row MUST occur in exactly one generated group for its stage and runtime profile. Task-surface ownership supplies the public Make binding, execution topology supplies row/profile/session/resource/fixture dependencies, and runtime-binary and image-readiness owners supply their existing producer prerequisites; generators join those owners without duplicating authority.

For a managed browser session, the suite-scoped browser lifecycle adapter is the
only owner of backend and frontend startup, readiness, startup events,
terminal startup diagnostics, v4 stack publication, and teardown. The
Playwright-facing adapter is attach-only. Before workers start it MUST validate
an exact `cartulary.web_e2e_stack.v4`, the suite/session/profile identities, all
referenced byte digests, the active schema/template/bucket/endpoint identities,
the frontend build digest, and live backend/frontend process proofs.
Missing, stale, v3-only, profile-mismatched, digest-mismatched, development-stack,
or incomplete attachment evidence MUST fail before Playwright assertions.
Canonical Playwright configuration MUST NOT start a web server, reuse an
existing listener, or derive origins from defaults. `--no-deps` MUST NOT bypass
the outer attachment guard.

Each session MUST retain one append-only validated event stream and one
immutable terminal diagnostic under
`_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/`.
Only the session lifecycle adapter may write them. A ready terminal requires
the complete ordered state graph `initializing`, `service_attached`,
`fixture_ready`, `backend_ready`, `frontend_ready`, `ready`; `failed` may close
any nonterminal state and can never regress. The terminal ready diagnostic MUST
be published before one immutable v4 stack binds its exact digest together with
the service-scope snapshot, lease, database, object-store namespace, fixture,
process, and frontend-build identities. Group and target results MUST carry
ordered run-relative session artifact references and SHA-256 digests. Shared or
multi-profile target projections MUST consume those artifacts without gaining
write authority. V3 stack and v1 diagnostic schemas are historical-validation
inputs only and MUST NOT be active-run admission fallbacks.

The `browser-e2e-stateful` public target MAY use generated `stateful_partition` groups when each partition declares explicit semantic row IDs and an explicit browser session group. Partitioning MUST preserve the same row inventory as the unpartitioned target. An empty adapter invocation MUST be omitted rather than represented as product success. Direct execution MUST reset between selector-file partitions. Scheduler execution MUST serialize stateful partitions that share a browser session group in authored order, and each partition's reset MUST complete before the next partition starts. Distinct browser session groups MAY overlap only when each group owns an isolated retained lifecycle. The `network_flow_claimed` profile MUST always have a distinct startup session for stateful, accessibility, visual, measurement, and webserver-backed evidence. Partitioning MUST NOT remove reset, taint, teardown, route-token, runtime identity, evidence-accounting, or target-summary evidence.

A browser session cleanup unit is a failure-tolerant scheduler finalizer. It MUST run after successful, failed, or dependency-skipped group outcomes, MUST release every retained resource claim owned by that session, and MUST NOT be dependency-skipped. A finalizer whose producer never started and therefore acquired no resource MUST complete as a successful no-op; it MUST NOT invoke lease cleanup, emit a target pass summary, or add a secondary missing-artifact failure. Valid aggregate Go evidence requires the complete declared shard set. After early scheduler stop, an aggregate finalizer for which none or only a proper subset of declared shards started MUST complete as a successful no-op and MUST NOT emit partial target evidence; once every declared shard started, missing shard metadata remains fail-closed under the ordinary artifact classification. Missing browser lease metadata for work that did start likewise remains fail-closed under the ordinary artifact or cleanup classification. A failed browser group MUST still produce the normalized scheduler and target summaries needed to retain its row/group failure classification; cleanup after failure MUST NOT degrade a product assertion into scheduler deadlock, missing-summary inference, or `unknown_failure`.

The Playwright result adapter MUST join observations by exact normalized selector file and exact catalog title. Zero observations, multiple observations, aggregate process success without selector observations, and an unauthorized Playwright skip are accounting failures; a product assertion failure MUST remain a product failure. Reports, stdout, and stderr MUST be retained through the redaction boundary, and only exact selector observations may close catalog rows.

Ordinary Playwright configuration MUST set `updateSnapshots=none`. Missing
goldens therefore fail ordinary visual validation. Snapshot mutation is
authorized only by `browser-e2e-visual-update`, which MUST retain the same row
selection, profile-derived service resolution, session grouping, attachment
evidence, and exact row accounting as validation and MUST NOT publish ordinary
passing visual evidence.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-022, TH-HARNESS-AC-063, TH-HARNESS-AC-069

**TH-HARNESS-REQ-057**
Warm steady-state `check-service-backed` timing is a harness health contract, not product performance evidence. For the supported WSL2 compatibility profile, a successful warm `check` run accepted for harness maintenance MUST keep `check-service-backed` wall time at or below the hard compatibility cap of `155000ms` unless the caller explicitly supplies a different maintenance budget to the timing checker. Non-isolated backend and browser peer lanes MUST NOT exceed `125%` of their peer median after applying the `5000ms` materiality floor. A lane is excluded only when the immutable plan marks it isolated or it is the only lane in its peer group.

Timing-accepted runs MUST be warm-ready runs. Cold provisioning, pinned tool installation, Go build-cache population, frontend install, Playwright install, service image build or warmup, and hidden helper builds are valid correctness evidence but MUST be attributed as readiness or provisioning work rather than accepted as warm scheduler-health evidence. At minimum, the service test binary build MUST be modeled as first-class scheduler readiness work separate from service image warmup.

Retained-run finalizer duration and finalizer-maintained baseline refresh effects are not part of the evaluated `make check` wall time. When `agent-finalize RESULTS_DIR=<run-root>` updates duration baselines, schedule inputs, generated schedule outputs, or topology render artifacts that can affect later warm scheduling, the updated files MUST be treated as affecting the next warm baseline only. A timing improvement claim after such a refresh MUST be based on a later warm-ready `make check` run produced after the refresh, not on the pre-refresh run or on finalizer duration.
Verified by: TH-HARNESS-AC-018

### 4.1 Mechanism Boundary

Harness implementation packages are owner-local realization details unless a row in this NLSpec explicitly promotes a specific path. Module names and directory names under `tools/harness/**` MUST NOT be treated as public harness invocation bindings, product conformance surfaces, or durable compatibility promises. Implementation code MUST depend on owner facades for contract, catalog, output, execution, finalization, diagnostics, smoke, frontend, browser, backend, scheduler, evidence-accounting, and generated-artifact behavior instead of importing private mixed-responsibility helper paths. The harness import-boundary check MUST reject private catch-all package imports once an owner-specific facade exists for the behavior.

Catalog shape, registry semantics, evidence naming, fixture policy, row selection, selector verification, and Go/Vitest/Playwright/shell result adaptation are catalog and evidence-accounting behavior. Task guidance, target explanation, owner explanation, retained-run explanation, target-plan display, and fixture diagnostics are diagnostics behavior. Runner-specific discovery remains behind runner adapters; test-output indexing and target-start statistics belong to test-output helpers; summary topology belongs to execution helpers; task-surface validation and generated Make rendering belong to generated-artifact helpers. Public Make target names, stable `command_id` values, schema IDs, output/failure contracts, and retained artifact paths remain the compatibility surface. A diagnostic surface that combines catalog, runner, task-surface, and scheduler reachability data MUST depend on the owner facades for each source family rather than importing private owner-local helpers.

### 4.1A Private Helper Ownership, Facades, and Compatibility

**TH-HARNESS-REQ-061**
Private helper paths under `tools/harness/**` MUST be classified by semantic ownership, not by historical directory name. A helper path MAY move, be renamed, or be deleted when public Make target names, stable `command_id` values, schema IDs, output shape, retained artifact paths, failure mapping, cleanup behavior, and declared public input contracts remain unchanged.

The following compatibility statuses are closed:

| Status | Definition |
| --- | --- |
| `public_contract` | A Make target, `command_id`, schema ID, output/failure contract, retained artifact path, or declared public input contract owned by this NLSpec. |
| `owner_facade` | A declared semantic import boundary for an implementation helper family. Non-owner callers MUST import this facade when the behavior is needed from another harness subsystem. |
| `unsupported_private` | A helper path that is not a public compatibility path and may move, be renamed, or be deleted without public contract revision when public harness behavior is preserved. |

The following helper ownership registry is closed for the current profile:

Repository attachment mechanics for these rows are owned by the
schema-validated `tools/harness_helper_ownership.json` input. Its `key` set
MUST exactly match the semantic `owner_facade` rows below. Each entry declares
only current facade paths and explicit cross-owner consumers; historical paths
and moved-module aliases are not owner input.

The scheduler helper rows below own harness orchestration only. They MUST NOT define product HTTP routes, WebSocket behavior, workbook mutation/query behavior, saved-view behavior, product transport adapters, frontend shell state, or grid-vendor integration; those questions remain owned by Core 00 through Core 04 and adopted subsystem owner documents.

| Helper family | Facade key | Owner boundary | Compatibility status | Observable behavior to preserve |
| --- | --- | --- | --- | --- |
| Backend Go target row discovery and backend target-plan DTOs. | `backend_target_plan` | Backend harness target planning. | `owner_facade` | Backend row identity, ordering, target selection, backend field values consumed by `target-plan-json`, and Go fixture-policy details. |
| Backend Go shard planning and shard-plan DTOs. | `backend_shard_plan` | Backend harness shard planning. | `owner_facade` | Shard names, ordering, fixture policy, target mapping, and shard-plan JSON. |
| Backend Go target execution, capture, finalization, and report emission. | `backend_target_execution` | Backend harness target execution. | `owner_facade` | CLI command set, command text, Go test environment, shared report locking and reuse, runtime-binary validation, owner summaries, target summaries, timing spans, finalizer diagnostics, and failure classification. |
| Go duration baseline coverage, drift, update inputs, and retained Go shard observations. | `backend_duration_accounting` | Backend duration accounting. | `owner_facade` | Baseline schema identity, coverage semantics, read-only drift, update inputs, and default weights. |
| Migration-history and schema-object ownership validation. | `database_contract_drift` | Generated-drift `database_contract_drift` sub-boundary. | `owner_facade` | Manifest schema IDs, current-line lineage validation, deterministic diagnostics when present, fresh and penultimate scratch migration behavior for `migration-drift`, and `json-shape-check` compatibility. |
| Govulncheck findings normalization and redaction. | `static_analysis_security_findings` | Static-analysis/security sub-boundary. | `owner_facade` | Redaction behavior, blocking/non-blocking classification, exit-code mapping, `cartulary.govulncheck_findings.v1`, and `GOVULNCHECK_DB` as the current public vulnerability database override. |
| Task-surface owner loading, topology, generated drift, schema attachment validation, and generated Make rendering. | `generated_artifact_surface` | Generated-artifact helpers. | `owner_facade` | Authored task-surface ownership, generated artifact source-of-truth parity, schema validation, drift failure mapping, and no hand-editing of generated outputs. |
| Public Make Node-tool command registry, input filtering, child environment construction, and invocation argument synthesis. | `command_surface_node_tool_dispatch` | Command-surface helpers; current facade `tools/harness/command-surface/make-node-tools.mjs`. | `owner_facade` | Public Make node-tool input matrix parity, inherited-environment stripping, runtime-env forwarding, usage diagnostics, and generated Make dispatch behavior. |
| Shared shell command runtime, timing, redaction, artifact directory policy, output mode resolution, step summaries, target summaries, and Vitest watchdog substrate. | `command_execution_runtime` | Execution runtime helper boundary. | `owner_facade` | Step and target timing spans, redacted stdout/stderr handling, retained artifact directories, output mode behavior, step-summary emission, public exit-code propagation, and watchdog sidecar behavior. |
| Cross-runner owner catalogs, row IDs, selectors, profiles, and slice planning. | `test_catalog` | Catalog and selection boundary. | `owner_facade` | Owner/verification/runner registry validation, exact selector resolution, profile resolution, catalog semantic digest, and immutable slice-plan semantics. |
| Owner evidence-accounting artifact generation and validation. | `test_evidence_accounting` | Evidence-accounting boundary. | `owner_facade` | `cartulary.test_evidence_accounting.v1`, terminal row closure, selected-row scope validation, retained artifact path `test-evidence-accounting.json`, and target-summary failure injection. |
| Owner retained-evidence audit. | `test_evidence_audit` | Evidence-accounting audit boundary. | `owner_facade` | `test-evidence-audit` retained-root inputs, semantic digest validation, closure audit semantics, and `cartulary.test_evidence_audit_summary.v1`. |
| Frontend unit target execution. | `frontend_target_execution` | Execution helper boundary. | `owner_facade` | `frontend-unit` command behavior, Vitest invocation shape, owner summaries, target summaries, exact selected-row filtering, and public failure mapping. |
| Vitest execution diagnostics and sidecars. | `vitest_execution_diagnostics` | Execution, test-output, and diagnostics helper boundaries. | `owner_facade` | Vitest row wrappers, exact-title filtering, watchdog handling, `cartulary.vitest_failure_details.v1`, and retained `vitest-failure-details.json` sidecar paths. |
| Frontend toolchain and dependency-install readiness. | `frontend_toolchain_readiness` | Readiness helper boundary. | `owner_facade` | Pinned repo-local Node/pnpm preparation, frozen-lockfile install behavior, readiness cache keys, stamp content, and configuration-error exit mapping. |
| Web build and embedded web asset artifacts. | `web_build_artifact` | Readiness/build-artifact helper boundary. | `owner_facade` | `build-web` and embedded asset cache behavior, Vite build invocation, embed archive/stamp atomicity, and public build target behavior. |
| Design-token generation. | `design_token_generation` | Generated-artifact design-token sub-boundary. | `owner_facade` | `docs/design.md` token parsing, generated token TypeScript content, generated provenance identity, and generated-artifact drift behavior. |
| Font asset validation. | `font_asset_validation` | Static-analysis helper boundary. | `owner_facade` | Font manifest validation, vendored font checksum/license checks, local CSS activation checks, and remote-font ban diagnostics. |
| Browser batch manifest loading, normalization, and target/stage metadata. | `browser_batch_manifest` | Browser manifest helper boundary; current canonical path `tools/harness/browser/browser-batch-manifest.mjs`. | `owner_facade` | `cartulary.browser_e2e_batch_manifest.v7`, catalog-derived semantic stage/group normalization, exact selector-file grouping, runtime-profile service resolution and session identity, target/stage metadata, and diagnostics used by generated-artifact and execution helpers. |
| Browser scheduler adapter command/dependency projection. | `browser_scheduler_adapter` | Scheduler browser adapter boundary; current facade path `tools/harness/scheduler/adapters/browser.mjs`. | `owner_facade` | Browser work-unit command paths, stage/session completion keys, group dependency keys, worker slot ranges, worker environment variables, and scheduler expansion semantics. |
| Scheduler DAG execution, events, finalizers, summaries, and failure mapping. | `scheduler_execution_core` | Scheduler execution helper boundary; current facade `tools/harness/scheduler/scheduler-runner.mjs`. | `owner_facade` | Work-unit ordering, logical resource scheduling, scheduler event stream, scheduler summary emission, finalizer handling, stable failure classes/reasons, public Make target behavior, and retained scheduler artifact paths. |
| Scheduler manifest, resource, family, and reporting helpers. | `scheduler_contract_helpers` | Scheduler contract helper boundary; current facades `tools/harness/scheduler/scheduler-family-contract.mjs`, `tools/harness/scheduler/scheduler-manifest.mjs`, `tools/harness/scheduler/scheduler-resources.mjs`, and `tools/harness/scheduler/scheduler-reporting.mjs`. | `owner_facade` | `cartulary.scheduler_manifest.v2`, scheduler family tokens, scheduler resource registry behavior, resource default/auto policies, retained scheduler paths, bounded reporting lines, and artifact references. |
| Scheduler child process execution and log handling. | `scheduler_process_execution` | Scheduler process adapter boundary; current facade `tools/harness/scheduler/process-executor.mjs`. | `owner_facade` | Child environment construction, sensitive environment stripping, dry-run detection, log-name sanitization, log replay, and child exit/failure propagation. |
| Owner-slice plan construction and selected work-unit accounting. | `test_slice_planning` | Catalog and execution planning boundary. | `owner_facade` | Owner/row selection, `cartulary.test_slice_plan.v2`, selected work-unit ordering, profile-backed evidence accounting, runtime-binary readiness, and scheduler-runner coordination inputs. |
| Service-backed schedule manifest expansion and topology planning. | `service_backed_schedule_planning` | Execution/service-backed planning boundary; current facade `tools/harness/execution/service-backed/schedule-planning.mjs`. | `owner_facade` | Service-backed schedule source validation, expansion, browser group/session mapping, resource mapping, topology checks, and generated scheduler manifest behavior. |
| Scheduler-retained duration baseline and drift accounting. | `scheduler_duration_accounting` | Duration-accounting helper boundary; current facade `tools/harness/duration-accounting/index.mjs`. | `owner_facade` | Shared duration thresholding, retained-run contamination detection, service-backed target duration baselines, harness-smoke duration baselines, read-only drift behavior, and retained-run validation before baseline mutation. |
| Scheduler retained event/timing diagnostics. | `scheduler_evidence_drift` | Diagnostics helper boundary; current entrypoints `tools/harness/diagnostics/scheduler-event-order-drift-cli.mjs` and `tools/harness/diagnostics/scheduler-summary-timing-drift-cli.mjs`. | `owner_facade` | Event-order drift, summary timing drift, warm-run eligibility checks, lane/session accounting diagnostics, bounded output, and retained-run evidence requirements. |
| Browser target execution wrappers and stage dispatch. | `browser_target_execution` | Browser target execution helper boundary; current canonical private runner `tools/harness/browser/run-browser-e2e-target.sh`. | `owner_facade` | Browser target wrapper entrypoint behavior, batch stage dispatch, target/stage summaries, stack session wrapping, reset boundaries, artifact paths, and public failure mapping. |
| Browser Playwright selection, webserver-batch execution, report parsing, and selection artifacts. | `browser_playwright_execution` | Browser Playwright execution plus test-output adapter boundary; current stable adapter `tools/harness/output/test-output/playwright-artifacts.mjs`. | `owner_facade` | Exact catalog row and scenario selection, Playwright runner report interpretation, selected-test title/file indexing, merged report behavior, owner summaries, stdout/stderr/output artifact paths, and failure normalization. |
| Browser duration accounting, shard planning, drift, and baseline refresh inputs. | `browser_duration_accounting` | Browser duration accounting boundary; current canonical facade `tools/harness/browser/browser-duration-accounting.mjs`. | `owner_facade` | `cartulary.browser_e2e_duration_baselines.v3`, `cartulary.browser_e2e_shard_plan.v2`, manifest row ID joining, frontend readiness row inclusion, read-only drift semantics, explicit retained-run validation before mutation, and default weights. |
| Browser owned-stack lifecycle, runtime identity proof, and runtime reset adapter. | `browser_lifecycle_adapter` | Browser lifecycle/test-route adapter boundary; current entrypoints `tools/harness/browser/start-web-e2e.sh` and `tools/harness/browser/reset-web-e2e-stack.sh`. | `owner_facade` | `cartulary.web_e2e_stack.v4`, per-session startup event and terminal-diagnostic ownership, immutable attachment evidence, preview-mode startup, port ownership, runtime root/session files, process-group cleanup, runtime identity proof, reset route token/origin/host predicates, reset taint handling, and Playwright state cleanup. |
| Browser accessibility evidence summaries. | `browser_accessibility_evidence` | Browser helper boundary; current canonical path `tools/harness/browser/browser-catalog-group-cli.mjs`. | `owner_facade` | Accessibility summary schema, contrast record handling, retained Playwright runner references, and browser a11y target artifact paths. |
| Browser visual snapshot update helper. | `browser_visual_update_helper` | Browser visual maintenance helper boundary; current entrypoint `tools/harness/browser/run-browser-e2e-visual-update.sh`. | `owner_facade` | Helper-only visual update target posture, snapshot-update mode propagation, authorized authored snapshot write path, retained browser evidence, and exclusion from default `check`, `test`, `ci`, and release gates unless separately declared. |

Migration-history and schema-object ownership validators are implementation-support evidence for database-contract drift. They MUST NOT become Core product behavior owners and MUST NOT be imported by backend target execution code except through a declared `database_contract_drift` facade.

Govulncheck findings normalization belongs to `static_analysis_security_findings`. Non-current backend helper paths for that behavior are private implementation history and are invalid as current compatibility paths.

Verified by: TH-HARNESS-AC-038, TH-HARNESS-AC-040, TH-HARNESS-AC-041

**TH-HARNESS-REQ-062**
Private helper compatibility MUST NOT be preserved by default. Archive references, historical handoffs, raw script paths, and old implementation imports do not establish compatibility support.

The current profile rejects private cross-owner imports through semantic
boundary rules derived from current ownership. The validator MUST retain the
general bans on `tools/harness/core/**`, unknown top-level harness owner roots,
private backend, browser, frontend catch-all, catalog, and evidence-accounting imports,
and scheduler-to-browser imports that bypass the browser scheduler adapter.
It MUST NOT maintain exact or prefix tombstones for historical paths.

Deletion or contraction of an `unsupported_private` family is allowed only
after current live callers are moved to declared owner facades,
generated/task-surface metadata does not reference the old path, semantic
import-boundary tests reject cross-owner access, and characterization tests
for the relevant public targets pass. A moved private module receives no
forwarding shim unless a continuing external consumer is demonstrated.

Verified by: TH-HARNESS-AC-039

**TH-HARNESS-REQ-063**
Harness implementation code MUST import declared owner facades rather than arbitrary private backend or frontend helper paths once the facade for that behavior exists. The harness import-boundary check MUST reject new non-owner imports from private implementation helpers and private catch-all imports where an owner facade is declared.

Tests MAY import private implementation fixtures only from declared test-support paths. Generated files MAY mirror declared owner paths but MUST NOT independently widen import allowances. Non-owner harness code MUST NOT import private browser implementation helpers directly once a browser owner facade exists; scheduler code MUST use the scheduler browser adapter instead of direct browser helper imports. Import-boundary failures are harness failures, not product failures.

Verified by: TH-HARNESS-AC-038, TH-HARNESS-AC-039

**TH-HARNESS-REQ-064**
Go duration-baseline maintenance MUST use the retained-run eligibility declared for each command family in the following table. `duration_retained_run` means retained run evidence that contains the target summaries, scheduler summaries, scheduler events, and Go shard/duration artifacts required to bind observed durations to current target and shard identities.

| Command family | Retained-run rule |
| --- | --- |
| `go-test-duration-baseline-coverage` | Does not require retained-run evidence. It verifies planned baseline coverage only, is read-only, and may remain ordinary `check` evidence. |
| `go-test-duration-baseline-drift` | MAY use `RESULTS_DIR` explicitly supplied by the caller or the current retained run when invoked inside a retained-run context. It MUST be read-only and MUST reject missing, failed, incomplete, or artifact-insufficient evidence before producing a drift verdict. |
| `go-test-duration-baselines` | MUST require explicit `RESULTS_DIR`. It MUST reject ambiguous result roots unless exactly one retained run is resolved by Sections 5 and 6. It MAY mutate baseline files only after retained-run validation succeeds. |
| `agent-finalize duration_baseline_refresh` | MUST use an existing retained full warm `make check` run root. Service-backed-only, owner-slice, browser-only, and other partial roots are invalid. |
| `agent-finalize duration_baseline_drift_validation` | MUST use the same retained-run requirement as `duration_baseline_refresh`, but remains read-only. |

Duration retained-run evidence MUST be rejected if it is failed, incomplete, contaminated, non-warm where warm evidence is required, missing full-check markers where full-check evidence is required, artifact-insufficient, or older than the latest sibling retained check run without the Section 8.2 older-run override where that override applies. A mutating baseline update MUST fail before the first mutation when retained-run validation fails.

Verified by: TH-HARNESS-AC-042

**TH-HARNESS-REQ-065**
Migration-history evidence capture is database-contract or migration-evidence evidence unless a later adopted owner explicitly promotes another boundary. Harness rows for migration manifest audit, embedded SQL source audit, goose ledger inspection, schema-object ownership drift, and migration-history diagnostics MUST route through database-contract or migration-evidence ownership. They MUST NOT be classified as operator-recovery conformance evidence merely because an implementation exposes a deployment-local operator wrapper.

A deployment-local wrapper for migration evidence MAY exist as an implementation mechanism. Omission behavior: if the wrapper is absent, harness conformance MAY still be satisfied through owner-backed database-contract drift or migration-evidence targets, provided the catalog rows cite the correct owner and retained artifacts satisfy the declared schema.

Verified by: TH-HARNESS-AC-045

**TH-HARNESS-REQ-066**
When production code or tests move across packages, harness accounting source-of-truth inputs MUST be updated before generated artifacts. The required order is:

1. Update the owner catalog row, target-map row, runtime-binary declaration, helper ownership row, task-surface input, topology input, or schedule input that owns the changed path or selection.
2. Regenerate downstream task-surface, schedule, topology, and evidence-accounting artifacts only through Make-owned generation.
3. Run drift and schema checks before treating the move as complete.
4. Treat any generated-ledger or generated-manifest hand edit as non-conformant.

A path-only test move that preserves row ID, title, target, owner references, evidence class, runtime-binary use, and retained artifact shape MAY remain behavior-preserving. A move that changes command grammar, result schema, authorization outcome, runtime-binary use, public target membership, default-check membership, scheduler topology, fixture lifecycle, or retained artifact shape is a public harness behavior change and MUST be specified in this NLSpec before implementation.

Required validation after moved-test accounting changes is:

| Change class | Required validation |
| --- | --- |
| Catalog or row ownership only | `make json-shape-check`, private `test-catalog-check`, and the affected owner slice. |
| Task-surface or public-target metadata | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` and public-target parity checks. |
| Scheduler topology or generated schedule | `make generate`, `make generate-drift`, and the affected scheduler target. |
| Generated artifacts | `make generate-drift` and `make generated-artifact-policy-check`. |
| Operator runtime-binary rows | `make build-operator` and the affected scheduler-selected operator work. |

Verified by: TH-HARNESS-AC-046

**TH-HARNESS-REQ-067**
Private backend test-support packages MUST be named for their semantic owner
and responsibility. Historical delivery labels are forbidden as evidence-accounting
identifiers, test IDs, catalog selectors, helper packages, exported helper types,
and exported helper symbols. A legitimate product-phase or execution-step term
requires the Section 15 semantic allowlist classification.

The repository MUST maintain one schema-validated test-support inventory that
classifies every shared and owner-local Go support root and every committed
shared fixture or golden root. The inventory MUST record semantic owner,
shared or owner-local posture, runtime-scan treatment, support-scan treatment,
and whether a Go root exposes service-starting `Start*` entrypoints. A support
root that is absent, duplicated, missing on disk, excluded from both runtime
and support security profiles, or inconsistent with its service-starting
classification MUST fail harness validation. Private compatibility packages
MUST NOT be retained solely to preserve an old helper import.

Verified by: TH-HARNESS-AC-056

### 4.1B Command-runtime profiles and legacy-surface disposition

**TH-HARNESS-REQ-068**
Cartulary has exactly three deployable executable identities: `server`, `migrate`,
and `operator`. `build-server-harness` is a harness-only build profile of the
existing `server` identity, not a fourth deployable, release artifact, or public
product command. `build-server` produces the production profile; it MUST NOT link
harness-route contributors or inherited-listener support.

The harness profile is selected only by the Make-owned build target and the private
`cartulary_harness` build tag. It is the only profile permitted to consume
`CARTULARY_ENABLE_TEST_ROUTES` or `CARTULARY_HTTP_LISTEN_FD`. The production profile
MUST reject either key before application runtime construction or listener
acquisition. Product HTTP, WebSocket, authorization, diagnostics, and packaged-asset
behavior shared by the two profiles remain owned by Core 00 through Core 04.

All black-box consumers MUST receive a declared runtime binary from the topology
runtime-binary registry. The injected file MUST be a scheduler-produced regular,
executable, digest-matched artifact for its declared profile. Go tests MUST NOT build
or run a command binary through nested `make`, `go build`, or `go run` fallbacks;
missing injection is a configuration failure that directs the caller to the relevant
public Make target.

| Surface | Owner | Supported caller and final execution surface | Retirement trigger |
| --- | --- | --- | --- |
| `server` production profile | `internal/app/server` plus `internal/platform/httpruntime` | Development, packaging, release, stand-up, and deployable-shape targets through `build-server`. | None; this is the production server identity. |
| `server` harness profile | Testing Harness and the `internal/app/server` harness contribution | Process and browser evidence through `build-server-harness` and declared `server-harness` runtime-binary rows only. | Remove when no harness route or inherited-listener consumer remains. |
| `migrate up` | `internal/app/migrate` and Postgres migration owner | Deployment/bootstrap, `db-migrate`, `db-reset`, and Make-owned migration targets only. | None while forward migration remains a deployment requirement. |
| Penultimate migration application | Postgres database-contract test support | `migration-scratch-apply` and migration-drift support only; never the `migrate` executable. | Remove when migration-line verification is retired by its owner. |
| Five recovery commands | Recovery module | Declared operator runtime-binary rows and deployment-local operator execution. | None while Core 01 requires them. |
| `migration-evidence capture` | Postgres migration-evidence owner | Declared database-contract or migration-evidence harness rows. | Remove when owner-defined evidence no longer consumes it. |
| `object-store init` | Object-store platform owner | Stand-up packaging and declared operator rows. | Remove when configured-bucket initialization is no longer needed. |

Retired recovery aliases, implicit migration commands, arbitrary Goose arguments,
contextless runner wrappers, and raw-Go or nested-build test fallbacks are unsupported
private compatibility. Historical archives may describe them but MUST NOT restore them
as current workflows.

Verified by: TH-HARNESS-AC-057

Private child-runner paths are implementation details and MUST be registered through the checked-in runner registry and invoked through Make-owned execution adapters. Legacy root runners, frontend catch-all runners, and core diagnostic shims MUST NOT be recreated as compatibility paths; callers MUST use the owning execution, browser, or diagnostics boundary.

`tools/harness/execution/cartulary-runner-cli.mjs` private direct use MUST select an explicit runner subcommand. Backend Go target execution is available only through `go-target <target-or-command> [...]`; direct aliases such as `backend-unit` or `backend-store` are unsupported private compatibility and MUST fail with usage status `2`. Quiet successful child logs MUST remain suppressed for public summaries regardless of unrelated environment variables.

Harness import-boundary validation MUST cover statically resolvable shell `source` and `.` references to repo-local `tools/harness/...` paths in addition to JavaScript imports. Unsupported private helper rules apply equally to those shell-source edges. Dynamic shell source expressions that cannot be resolved to a repo-local static path remain outside this static rule and are covered by shell lint plus target-level tests.

| Surface                                                  |                                  Normative? | Required contract                                                                      |
| -------------------------------------------------------- | ------------------------------------------: | -------------------------------------------------------------------------------------- |
| Public Make target name                                  |                                         yes | Stable command surface invoked as `make <target>` from the repository root.            |
| `tools/task_surface_owner.json` public target metadata | yes | Required authored machine owner downstream of this NLSpec. |
| `tools/task_surface_manifest.json` public target_class | yes | Required generated machine-readable mirror of the public target registry. |
| Root/package `pnpm` scripts                              |                                          no | Developer convenience unless invoked by a Make-owned public target. Successful raw package-script output MUST NOT be reported as completion evidence for public harness targets. |
| Raw owner helper scripts and child CLIs                   | no | May change when public Make behavior remains unchanged.                                |
| Make-owned harness contract helper implementation path    | no as an additional invocation binding; yes as the owner implementation path used by Make wrappers | Make-owned public wrappers invoke this owner CLI for preflight, cleanup, schema validation, and related harness-contract mechanics. Direct CLI invocation remains implementation support unless a public Make target adopts it. |
| Make-owned test-output helper implementation path         | yes only for Make-owned command behavior; no for JavaScript module boundaries | Make-owned wrappers may invoke these helpers to emit lifecycle lines, machine output, summaries, and retained artifacts. Command names, arguments, output modes, schema IDs, failure taxonomy, exit codes, and retained artifact paths are the contract; helper filenames, exports, imports, and private module locations are implementation details unless explicitly promoted by this NLSpec. |
| `tools/testservices` binary path                         |                                          no | Service lifecycle behavior is normative; binary path is an implementation realization. |
| Public output classes and schema IDs listed in Section 8 |                                         yes | Required machine-output and artifact validation contracts.                              |
| Docker image tag for Postgres or object-store services   | no unless declared in a service fixture row | Exact tag is not normative unless it defines fixture semantics in Section 11.          |
| Generated Make include names, helper binaries, helper target classes, priority-band names, and generator constants | no | Implementation detail unless promoted by an explicit requirement.                      |

**TH-HARNESS-REQ-069**
Task-surface Make binding profiles are private generated-artifact implementation
details. The current closed profile types are `artifact_binding`, `aggregate`,
`readiness_projection`, `cleanup`, `print_help`, `sequence`, `check_schedule`,
`go_target`, `service_backed_target`, `service_backed_schedule`, `browser_batch`,
`owner_command`, `summary_target`, and `node_tool`. The former catch-all `alias`
profile is unsupported. Generated bindings MUST factor invariant preflight,
prerequisite, input sanitization, and summary behavior through a shared generated
runtime rather than repeating the global public-input inventory per target.

`tools/task_surface.generated.mk` MUST be no larger than 180 KiB; the shared generated
runtime include MUST be no larger than 64 KiB; their combined size MUST be no larger
than 220 KiB; and neither file may contain a physical line longer than 512 bytes. A
synthetic addition of 25 ordinary owner-style targets MUST grow generated output by no
more than 512 bytes per target on average. These are implementation-maintainability
gates and do not change public target behavior.
Verified by: TH-HARNESS-AC-058

**TH-HARNESS-REQ-071**
The current owner-first public commands are closed by this table:

| Target | Command ID | Required target-local inputs | Optional target-local inputs |
| --- | --- | --- | --- |
| `test-slice` | `cartulary.harness.command.test_slice.v1` | `OWNER` | `ROWS`, `VITEST_MAX_WORKERS`, `PLAYWRIGHT_WORKERS`, `JSON` |
| `service-backed-test-slice` | `cartulary.harness.command.service_backed_test_slice.v1` | `OWNER` | `ROWS`, `VITEST_MAX_WORKERS`, `PLAYWRIGHT_WORKERS`, `JSON` |
| `explain-test-owner` | `cartulary.harness.command.explain_test_owner.v1` | `OWNER` | `JSON` |
| `task-guide` | `cartulary.harness.command.task_guide.v2` | `ROLE=module-author`, `OWNER` | `JSON` |
| `test-evidence-audit` | `cartulary.harness.command.test_evidence_audit.v1` | `OWNER`, `EVIDENCE_ROOTS_FILE` | none |

`test-catalog-check` is private check-level work, accepts no target-local public input, has no stable public command ID, and MUST be selected by `make check`. A public or private v1 delivery-phase command, alias, reader, or dual writer is unsupported.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-064, TH-HARNESS-AC-070

**TH-HARNESS-REQ-072**
`test-slice` and `service-backed-test-slice` MUST use `scheduler_summary_with_artifacts`; `test-evidence-audit` MUST use `summary_with_artifacts`; `explain-test-owner` and `task-guide` MUST use `human_summary` and remain read-only. Section 7 owns the exact human, machine, and target-local JSON behavior. Section 8 owns the exact artifacts.
Verified by: TH-HARNESS-AC-004, TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-073**
The public commands, machine contracts, selection inputs, artifact paths, and schema IDs declared by this specification are closed. Any undeclared predecessor identity is unsupported and MUST NOT be accepted, translated, discovered, read, or emitted. Historical migration mappings belong only in the immutable reconciliation handoff and Git history; they are not normative runtime inputs.
Verified by: TH-HARNESS-AC-068, TH-HARNESS-AC-070

**TH-HARNESS-REQ-074**
Public command preflight MUST validate every target-local input, owner, row, selector, and profile before setup or child work. Removed inputs and unknown Make command-line inputs MUST fail as `usage_error`; undeclared inherited environment variables retain the Section 5 ignore rule. A caller MUST NOT obtain partial selection by supplying an invalid row.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-065

**TH-HARNESS-REQ-075**
`explain-test-owner` MUST report the selected manifest, semantic families, row counts, runner and evidence distribution, execution profiles, default-check participation, and exact narrow commands. `task-guide ROLE=module-author OWNER=<owner_id>` MUST derive its focused, generation, and broader commands from the same catalog and topology snapshot. Neither command may infer ownership from paths or documentation.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-076**
The hard cutover MUST update this NLSpec, authored task surface, catalog, verification contracts, topology, schemas, generated outputs, public help, active guides, and deletion obligations in one merged change set. A state in which v1 and v2 public selection are both supported is nonconforming.
Verified by: TH-HARNESS-AC-070, TH-HARNESS-AC-071

### 4.2 Command Family Defaults

Target membership for each family is defined only by `### 4.3 Public Target Registry`. The family-default table owns shared behavior for targets whose registry row declares that family. If a target appears in a prose command list and not in the registry, the registry governs and the prose is editorial drift.

| Family | Family ID | Required inputs | Optional inputs and defaults | Output class family | Scheduler use | Backing services | Artifact behavior | Failure contract |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| help and discovery | `help_discovery` | Target-local inputs declared by the Section 5.3 per-target input registry. | Omitted optional inputs select documented summary views according to the target's `input_contract`. | The target row's Section 4.3 output class, constrained by Section 7.2. | None | None | Does not create central run evidence unless the target row declares a summary schema. | Usage/config errors use Section 9. |
| bootstrap and toolchain | `bootstrap_toolchain` | Required local tools according to target | Tool paths default from Section 5. | `summary_with_artifacts` | None | May download/install repo-local tools. | Tool-run summary required; readiness cache artifacts MAY be retained when a cache profile is active. | Tool/config failures are `configuration_error` or `preflight_error`. |
| local services and dev | `local_services_dev` | Docker Compose and local config required by the target row. | `CONFIG_FILE=configs/dev/config.toml`, `OBJECT_STORE_BUCKET=cartulary` for rows that read those variables. | `service_summary` or `interactive_raw` | None | Compose Postgres/SeaweedFS S3 and local processes. | `service_summary` rows emit service summaries; `dev` has no verification artifact contract. | Startup/readiness/config failures are harness operational failures. |
| generated and drift | `generated_drift` | Owner inputs and manifests | `RESULTS_DIR` rows in the Section 5.3 input matrix select retained evidence. | `summary_with_artifacts` | Child work is scheduled only through Section 10 normalized scheduler work units. | Migration drift may use scratch Postgres when scheduled. | Tool summary and command-specific drift/finalizer files. | Drift mismatch is `artifact_error` or `scheduler_accounting_error`; unsafe retained-run finalization evidence is `artifact_error` or `configuration_error`. |
| owner and service slices | `test_owner_slices` | `OWNER` and any exact row selector declared by Section 5.3. | Omitted `ROWS` uses TH-HARNESS-REQ-360; worker defaults and `JSON` use Section 5.3. | `scheduler_summary_with_artifacts` | Uses owner selection and the generic scheduler. | Service-backed owner work requires its declared managed services. | Target, run, scheduler, accounting, and owner artifacts. | Missing/invalid owner or row is `usage_error`; child failures retain child class. |
| backend and frontend leaf tests | `backend_frontend_leaf_tests` | Toolchain, catalog, and package inputs | Parallelism and worker variables from Section 5. | `summary_with_artifacts` | Service-backed targets may use a scheduler or testservices. | Store/integration/process targets require Postgres/object-store services when service-backed. | Owner, target, tool, logs, and reports. | Product assertion failures are `test_assertion_failure`; setup failures are operational. |
| browser E2E | `browser_e2e` | Node/pnpm, Playwright browser, backend/migrate/server support, services | `PLAYWRIGHT_WORKERS=3`, `BROWSER_E2E_FUNCTIONAL_SHARDS=auto` unless overridden by Section 5. | `summary_with_artifacts` | Uses browser batch and service-backed scheduler only for rows closed by Section 10. | Postgres, object-store service, backend, frontend, browser runtime. | Browser stack, Playwright, reset, target, scheduler artifacts. | Product assertions are product failures; stack/readiness/reset failures are operational. |
| aggregates and gates | `aggregates_gates` | Toolchain and child inputs | `summary` output mode by default; `ci` defaults to `ci` mode through Make-owned CI target environment. | aggregate or scheduler output classes | `check` uses the check scheduler. | Service-backed and browser children require backing services. | Aggregate run, child target, scheduler, and tool summaries. | Exit nonzero if any required child fails or artifact validation fails. |
| static analysis and security | `static_analysis_security` | Toolchain and source roots | Rule, flag, package, and Fallow static profiles named by public target rows and Section 5.3 inputs. Shell lint is blocking for public Make targets. `GOVULNCHECK_DB` is the only public security-profile override in the current profile. Fallow Runtime and Fallow security-scan commands are not selected by this profile. | `summary_with_artifacts` | Scheduled inside `check` only through Section 10 normalized scheduler work units. | None | Tool summary and logs; security scans are uncached. | Findings are gate failures for scheduled local correctness targets. Advisory targets MUST be explicitly selected outside local `check`. |
| builds | `builds` | Build inputs and toolchain | Output paths from Make variables. | `summary_with_artifacts` | Scheduled as readiness work only through Section 10 normalized scheduler work units. | None | Tool summary and build logs; build cache artifacts MAY be retained when a cache profile is active. | Build failures are gate failures. |
| cleanup | `cleanup` | None | Uses Make path registries. | `destructive_human` | None | Does not stop Docker Compose services. | No central summary contract. | Unsafe path guard failure exits nonzero; missing paths are not failures. |
| formatting | `formatting` | Toolchain | None | `summary_with_artifacts` | None | None | Tool summary and formatter logs. | Formatter failure is operational; formatter rewrites are mutating. |

For `test_owner_slices`, omitted `ROWS` selects the inventory closed by TH-HARNESS-REQ-360. An explicit `ROWS` value selects only active executable row IDs owned by the exact requested owner. Every `service-backed-test-slice` has `dependency_scope="service_backed"`; omission still produces `completion_scope="full_owner"` over the owner's service-backed inventory, while an explicit selection produces `completion_scope="selected_subset"`.

`VITEST_MAX_WORKERS` and `PLAYWRIGHT_WORKERS` on this family apply only to
selected child work in the matching runner family. A slice that selects no
matching child MAY accept and report the bounded input but MUST NOT use it to
change another runner's concurrency or scheduler resource limits.

### 4.3 Public Target Registry

Every command below inherits the matching family defaults. `Default inclusion sets` lists direct full-target default memberships only; bounded default-check evidence for a full target is described by `check_projection` metadata instead of ordinary `check` membership. `helper_only` means the target is public and directly invocable, but is not selected by default by `test`, `check`, `ci`, or `release-check` unless another registry row explicitly includes it. `helper_only` MUST NOT mean private, uncontracted, or exempt from public-target output, configuration, failure, and cleanup contracts.

`Command ID` is the stable semantic command contract; the Make target is the current invocation binding. `Family ID` binds the target to Section 4.2 family defaults. `Semantic behaviors` declares the observable harness operation required by TH-HARNESS-REQ-058. `Side effects` declares the target's intentional mutation and resource contract from TH-HARNESS-REQ-059. The visible `Side effects` column MUST match the `side_effects[].class` list in `tools/task_surface_manifest.json`. `Lifecycle state` is defined by Section 4.6.

Public aggregate targets MAY be represented in scheduler manifests by typed internal helper work when the stable public command contains distinct resource, policy, or lifecycle boundaries. `migration-drift` is the current example: direct `make migration-drift` MUST remain the public aggregate that runs static migration-input validation plus scratch database apply evidence, while default `check` MUST schedule `migration-input-drift` and `migration-scratch-apply` as separate internal helper work units so static policy validation and scratch Postgres evidence keep separate resource and artifact identities.

| Target | Command ID | Family ID | Default inclusion sets | Output class | Stable summary schema | Semantic behaviors | Side effects | Lifecycle state | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `help` | `cartulary.harness.command.help.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `help-all` | `cartulary.harness.command.help_all.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `doctor` | `cartulary.harness.command.doctor.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `bootstrap` | `cartulary.harness.command.bootstrap.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `bootstrap-node-runtime` | `cartulary.harness.command.bootstrap_node_runtime.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `frontend-toolchain` | `cartulary.harness.command.frontend_toolchain.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `frontend-install` | `cartulary.harness.command.frontend_install.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `playwright-install` | `cartulary.harness.command.playwright_install.v1` | `bootstrap_toolchain` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `tool_install` | `public_active` |  |
| `db-up` | `cartulary.harness.command.db_up.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start` | `public_active` |  |
| `db-migrate` | `cartulary.harness.command.db_migrate.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` | Proves local Postgres readiness under Section 11, starting an owned instance when absent, and applies current-line migrations without resetting the database or object storage. |
| `db-reset` | `cartulary.harness.command.db_reset.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `destructive_safety` (Section 13), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `destructive_cleanup` | `public_active` | Requires `CARTULARY_DESTRUCTIVE_CONFIRM=db-reset` unless `CARTULARY_CLEANUP_DRY_RUN=1`; resets only the local database and does not reset object storage. |
| `services-up` | `cartulary.harness.command.services_up.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start` | `public_active` |  |
| `services-down` | `cartulary.harness.command.services_down.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `destructive_safety` (Section 13), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `destructive_cleanup` | `public_active` | Stops local Compose services with named volumes preserved. |
| `object-store-init` | `cartulary.harness.command.object_store_init.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` | Proves local object-store readiness under Section 11, starting an owned instance when absent, and initializes the configured bucket without requiring Postgres. |
| `object-store-reset` | `cartulary.harness.command.object_store_reset.v1` | `local_services_dev` | `helper_only` | `service_summary` | `cartulary.tool_run_summary.v5` | `destructive_safety` (Section 13), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `destructive_cleanup` | `public_active` | Requires `CARTULARY_DESTRUCTIVE_CONFIRM=object-store-reset` unless `CARTULARY_CLEANUP_DRY_RUN=1`; clears objects only from the configured local object-store bucket. |
| `dev` | `cartulary.harness.command.dev.v1` | `local_services_dev` | `helper_only` | `interactive_raw` | none | `service_lifecycle` (Section 11) | `service_start` | `public_active` |  |
| `generate` | `cartulary.harness.command.generate.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `generate-drift` | `cartulary.harness.command.generate_drift.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `generated-artifact-policy-check` | `cartulary.harness.command.generated_artifact_policy_check.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `json-shape-check` | `cartulary.harness.command.json_shape_check.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `toolchain-drift` | `cartulary.harness.command.toolchain_drift.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `migration-drift` | `cartulary.harness.command.migration_drift.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_resource_mutation` | `public_active` |  |
| `agent-finalize` | `cartulary.harness.command.agent_finalize.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `test-evidence-audit` | `cartulary.harness.command.test_evidence_audit.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Audits one owner's catalog rows across compatible retained broad-check, support, visual, accessibility, and measurement roots. |
| `benchmark-claim-check` | `cartulary.harness.command.benchmark_claim_check.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Validates retained Core 05 benchmark claim artifacts when the default benchmark manifest exists; absence of the default manifest is a no-claim pass, while an explicitly configured non-default missing manifest remains a harness failure. |
| `task-surface-report` | `cartulary.harness.command.task_surface_report.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `task-guide` | `cartulary.harness.command.task_guide.v2` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` | Requires `ROLE=module-author` and `OWNER`. |
| `test-slice` | `cartulary.harness.command.test_slice.v1` | `test_owner_slices` | `helper_only` | `scheduler_summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `scheduler_orchestration` (Section 10), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Selects every active owner row when `ROWS` is omitted. |
| `service-backed-test-slice` | `cartulary.harness.command.service_backed_test_slice.v1` | `test_owner_slices` | `helper_only` | `scheduler_summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `scheduler_orchestration` (Section 10), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Selects only rows whose runtime profile requires managed services. |
| `backend-unit` | `cartulary.harness.command.backend_unit.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `backend-store` | `cartulary.harness.command.backend_store.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` |  |
| `backend-integration` | `cartulary.harness.command.backend_integration.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` |  |
| `backend-process` | `cartulary.harness.command.backend_process.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` |  |
| `otel-conformance` | `cartulary.harness.command.otel_conformance.v1` | `backend_frontend_leaf_tests` | `check`, `ci` | `summary_with_artifacts` | `cartulary.otel_conformance_summary.v1` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `security_boundary` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Validates source snapshot, generated constants evidence, emitted telemetry goldens, browser non-export, retained raw capture policy, and telemetry security boundaries. |
| `target-plan` | `cartulary.harness.command.target_plan.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `target-plan-json` | `cartulary.harness.command.target_plan_json.v1` | `help_discovery` | `helper_only` | `machine_stdout_json` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `fixture-report` | `cartulary.harness.command.fixture_report.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `explain-run` | `cartulary.harness.command.explain_run.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `harness-observability-check` | `cartulary.harness.command.harness_observability_check.v1` | `generated_drift` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4), `security_boundary` (Section 15) | `none` | `public_active` | Read-only validation of deterministic harness diagnostic projection for one exact retained run. |
| `harness-otel-export` | `cartulary.harness.command.harness_otel_export.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4), `security_boundary` (Section 15) | `external_network` | `public_active` | Explicit post-run OTLP export; ordinary harness commands never invoke it. |
| `harness-performance-check` | `cartulary.harness.command.harness_performance_check.v2` | `generated_drift` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4), `failure_normalization` (Section 9) | `none` | `public_active` | Validates exact target/provider-scoped baseline and candidate evidence windows under Section 10.5. |
| `harness-public-target-duration-baselines` | `cartulary.harness.command.harness_public_target_duration_baselines.v2` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `diagnostic_synthesis` (Section 4), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` | Sole writer of the public-target duration baseline artifact from exact qualified target/provider windows. |
| `explain-test-owner` | `cartulary.harness.command.explain_test_owner.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` | Read-only catalog and topology explanation for one owner. |
| `explain-target` | `cartulary.harness.command.explain_target.v1` | `help_discovery` | `helper_only` | `human_summary` | none | `diagnostic_synthesis` (Section 4) | `none` | `public_active` |  |
| `go-test-duration-baselines` | `cartulary.harness.command.go_test_duration_baselines.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `go-test-duration-baseline-coverage` | `cartulary.harness.command.go_test_duration_baseline_coverage.v1` | `generated_drift` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `go-test-duration-baseline-drift` | `cartulary.harness.command.go_test_duration_baseline_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `browser-e2e-duration-baselines` | `cartulary.harness.command.browser_e2e_duration_baselines.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `browser-e2e-duration-baseline-drift` | `cartulary.harness.command.browser_e2e_duration_baseline_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `service-backed-make-target-duration-baselines` | `cartulary.harness.command.service_backed_make_target_duration_baselines.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `service-backed-make-target-duration-baseline-drift` | `cartulary.harness.command.service_backed_make_target_duration_baseline_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `harness-smoke-duration-baselines` | `cartulary.harness.command.harness_smoke_duration_baselines.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `generated_artifacts` | `public_active` |  |
| `harness-smoke-duration-baseline-drift` | `cartulary.harness.command.harness_smoke_duration_baseline_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `scheduler-event-order-drift` | `cartulary.harness.command.scheduler_event_order_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `scheduler-summary-timing-drift` | `cartulary.harness.command.scheduler_summary_timing_drift.v1` | `generated_drift` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `frontend-typecheck` | `cartulary.harness.command.frontend_typecheck.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `frontend-unit` | `cartulary.harness.command.frontend_unit.v1` | `backend_frontend_leaf_tests` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `frontend-import-boundary-check` | `cartulary.harness.command.frontend_import_boundary_check.v1` | `backend_frontend_leaf_tests` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `backend-module-boundary-check` | `cartulary.harness.command.backend_module_boundary_check.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `security_boundary` (Section 8), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Enforces backend module ownership boundaries from `tools/backend_module_boundaries.json` and emits `cartulary.backend_module_boundary_summary.v1`. |
| `frontend-fallow-static` | `cartulary.harness.command.frontend_fallow_static.v1` | `static_analysis_security` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Current helper-only Fallow static profile. Emits `cartulary.fallow_static_summary.v2`; Fallow Runtime and Fallow security scans are not selected. |
| `lint-biome` | `cartulary.harness.command.lint_biome.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `lint-scripts` | `cartulary.harness.command.lint_scripts.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `lint-markdown` | `cartulary.harness.command.lint_markdown.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Structural Markdown lint over authored active docs; generated ledgers remain generator-owned. |
| `lint-shell` | `cartulary.harness.command.lint_shell.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `format` | `cartulary.harness.command.format.v1` | `formatting` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `authored_source_write` | `public_active` |  |
| `browser-e2e` | `cartulary.harness.command.browser_e2e.v1` | `browser_e2e` | `test`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Direct aggregate remains full browser evidence and is not a default local `check` member. |
| `browser-e2e-webserver-backed` | `cartulary.harness.command.browser_e2e_webserver_backed.v1` | `browser_e2e` | `test`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Default `check` uses a `check-service-backed` `webserver-backed` projection classified as default-local cross-stack conformance with `full_target_equivalent=false`; direct target evidence remains full-fidelity. |
| `browser-e2e-stateful` | `cartulary.harness.command.browser_e2e_stateful.v1` | `browser_e2e` | `test`, `check`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Default `check` uses `check_projection.mode=direct` for `check-service-backed` `stateful` evidence with `full_target_equivalent=true`. |
| `browser-e2e-measurement` | `cartulary.harness.command.browser_e2e_measurement.v1` | `browser_e2e` | `test`, `ci` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `browser-e2e-a11y` | `cartulary.harness.command.browser_e2e_a11y.v1` | `browser_e2e` | `test`, `ci`, `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Emits `cartulary.frontend_accessibility_summary.v4` for active selected accessibility rows only. Direct target evidence remains full-fidelity and default local `check` does not schedule current accessibility work. Release-check uses it as release-readiness evidence only. |
| `browser-e2e-visual` | `cartulary.harness.command.browser_e2e_visual.v1` | `browser_e2e` | `test`, `ci`, `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Direct target evidence runs the full active visual catalog inventory. Owner slices pass exact selected visual row IDs. Default local `check` does not schedule full visual work. Release-check uses it as release-readiness evidence only. |
| `browser-e2e-visual-update` | `cartulary.harness.command.browser_e2e_visual_update.v1` | `browser_e2e` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `authored_source_write`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Helper-only refresh command for committed Playwright visual goldens. It runs the same visual stack and row selection as `browser-e2e-visual` with snapshot-update mode, mutates only authored snapshot PNGs under the visual spec snapshot directory, and MUST NOT be selected by `check`, `test`, `ci`, or release gates. |
| `harness-contract` | `cartulary.harness.command.harness_contract.v1` | `static_analysis_security` | `ci`, `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Runs extended harness topology, schema, and field-shape contract checks outside default local `check`. |
| `test-fast` | `cartulary.harness.command.test_fast.v1` | `aggregates_gates` | `test`, `check`, `ci` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `test` | `cartulary.harness.command.test.v1` | `aggregates_gates` | `test`, `check`, `ci` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `lint` | `cartulary.harness.command.lint.v1` | `aggregates_gates` | `helper_only` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Sequence aggregate that emits run and target summaries for blocking lint/typecheck children. |
| `go-vulncheck` | `cartulary.harness.command.go_vulncheck.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Runs Govulncheck with structured JSON output and retains `govulncheck-findings.json` with schema ID `cartulary.govulncheck_findings.v1`. Symbol-reachable vulnerability findings are blocking `security_finding` failures; package-only and module-only findings are retained as diagnostic security evidence unless a later profile promotes them. |
| `go-gosec-targeted` | `cartulary.harness.command.go_gosec_targeted.v1` | `static_analysis_security` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` |  |
| `go-gosec-audit` | `cartulary.harness.command.go_gosec_audit.v1` | `static_analysis_security` | `ci`, `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `security_boundary` (Section 15), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Advisory no-fail audit evidence. It MUST NOT be selected by default local `check`. |
| `seaweedfs-compatibility` | `cartulary.harness.command.seaweedfs_compatibility.v1` | `local_services_dev` | `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation` | `public_active` | Runs the dedicated SeaweedFS S3 compatibility profile and emits the full `SWFS-COMP-*` report outside `services-up` as a command-specific retained artifact. |
| `standup-package-smoke` | `cartulary.harness.command.standup_package_smoke.v1` | `local_services_dev` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs`, `service_start`, `service_resource_mutation` | `public_active` | Builds and smokes the MVP on-prem stand-up package with one app image plus local Postgres and SeaweedFS S3, verifies embedded browser assets, package-local object-store init, readiness, persistent Docker-volume roots, and WebSocket Origin behavior. It is package smoke evidence only and MUST NOT be represented as disconnected-profile or backup/restore conformance evidence. |
| `standup-operational-recovery-smoke` | `cartulary.harness.command.standup_operational_recovery_smoke.v1` | `local_services_dev` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `service_lifecycle` (Section 11), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs`, `service_start`, `service_resource_mutation` | `public_active` | Builds and smokes the MVP on-prem operational recovery workflow, creates a backup through the canonical operator recovery result schema, inspects the latest retained backup through the canonical inspect command, runs due restore verification against an isolated target, proves public recovery route-family absence, and retains command-specific recovery artifacts. It is not disconnected-profile evidence and does not reclassify `standup-package-smoke` as backup/restore conformance evidence. |
| `seaweedfs-release-evidence` | `cartulary.harness.command.seaweedfs_release_evidence.v1` | `static_analysis_security` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `security_boundary` (Section 15) | `retained_artifacts` | `public_active` | Runs current SeaweedFS compatibility as a prerequisite, emits SeaweedFS release evidence, and emits a non-enforcing release-gate summary as command-specific retained artifacts; missing strict child evidence is reported as blocked evidence rather than hidden. |
| `seaweedfs-release-gate` | `cartulary.harness.command.seaweedfs_release_gate.v1` | `static_analysis_security` | `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9), `security_boundary` (Section 15) | `retained_artifacts` | `public_active` | Runs current SeaweedFS compatibility and enforces the strict SeaweedFS release gate from current-run compatibility, backup-integrity, redaction, storage-ref owner, security, license, and occurrence evidence. The release-gate summary is a command-specific retained artifact. |
| `check` | `cartulary.harness.command.check.v1` | `aggregates_gates` | `check` | `scheduler_summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `scheduler_orchestration` (Section 10), `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `ci` | `cartulary.harness.command.ci.v1` | `aggregates_gates` | `ci` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` |  |
| `release-check` | `cartulary.harness.command.release_check.v1` | `aggregates_gates` | `release-check` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `service_start`, `service_resource_mutation`, `runtime_reset` | `public_active` | Runs release child helpers for extended harness contract checks, advisory security audit evidence, SBOM/license evidence, SeaweedFS S3 release-gate evidence, builds, deployable-shape evidence, frontend support/visual/accessibility readiness children, and final release-readiness aggregation. |
| `release-readiness-evidence` | `cartulary.harness.command.release_readiness_evidence.v1` | `aggregates_gates` | `release-check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts` | `public_active` | Aggregates current-run target summaries and the exact per-owner `test-evidence-accounting.json` and `test-owner-summary.json` partitions for catalog-backed visual, accessibility, and support gates, plus release artifacts and harness-contract outputs, into `cartulary.release_readiness_evidence.v2`. Retired frontend row-accounting artifacts are not ingested. The aggregate records release-gate effects without promoting design/support evidence to product conformance or Core 05 publication evidence. |
| `build` | `cartulary.harness.command.build.v1` | `builds` | `helper_only` | `aggregate_summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` |  |
| `build-server` | `cartulary.harness.command.build_server.v1` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | In default local `check`, build evidence is readiness for downstream service-backed work, not release deployable-shape evidence. |
| `build-server-harness` | `cartulary.harness.command.build_server_harness.v1` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | Builds the harness-only profile of the existing `server` identity. Default local `check` selects it as process/browser runtime readiness; it is never release deployable-shape evidence. |
| `build-migrate` | `cartulary.harness.command.build_migrate.v1` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | In default local `check`, build evidence is readiness for migration and service-backed work, not release deployable-shape evidence. |
| `build-operator` | `cartulary.harness.command.build_operator.v1` | `builds` | `helper_only` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | Selected through `build`, CI, release-shaped gates, and scheduler-visible operator runtime-binary readiness. Default local `check` builds the operator only when selected runtime-binary work declares it. |
| `build-web` | `cartulary.harness.command.build_web.v1` | `builds` | `check` | `summary_with_artifacts` | `cartulary.tool_run_summary.v5` | `evidence_normalization` (Section 8), `failure_normalization` (Section 9) | `retained_artifacts`, `build_outputs` | `public_active` | In default local `check`, build evidence is readiness for browser preview work, not release deployable-shape evidence. |
| `clean` | `cartulary.harness.command.clean.v1` | `cleanup` | `helper_only` | `destructive_human` | none | `destructive_safety` (Section 13), `failure_normalization` (Section 9) | `destructive_cleanup` | `public_active` |  |
| `distclean` | `cartulary.harness.command.distclean.v1` | `cleanup` | `helper_only` | `destructive_human` | none | `destructive_safety` (Section 13), `failure_normalization` (Section 9) | `destructive_cleanup` | `public_active` |  |

**TH-HARNESS-REQ-059**
Every public target MUST declare one or more side-effect classes in the public target registry source. The declaration MUST be represented as `side_effects[]`, where each entry is an object with `class`, `owner_section`, and the class-specific details required by the table below. A target that performs an undeclared side effect is non-conformant. `none` is mutually exclusive with every other side-effect class.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-020, TH-HARNESS-AC-023

| Side-effect class | Meaning | Required declaration |
| --- | --- | --- |
| `none` | No intentional file, service, or resource mutation outside ordinary terminal output. | Target row declares only `side_effects[].class=none`. |
| `retained_artifacts` | Writes retained run-root artifacts. | Artifact policy declares retained artifact families or paths. |
| `generated_artifacts` | Mutates checked-in generated or maintenance artifacts. | Target row declares exact generated file families. |
| `authored_source_write` | Mutates authored source files. | Target row declares source families or paths. |
| `build_outputs` | Writes reproducible build outputs. | Target row declares output roots or artifact families. |
| `tool_install` | Installs or updates repo-local tools or dependencies. | Target row declares install root and cleanup behavior. |
| `service_start` | Starts local or harness-owned services or runtime processes. | Target row declares ownership mode and lifecycle machine. |
| `service_resource_mutation` | Creates, modifies, or deletes service resources such as scratch databases, buckets, fixture resources, or local service bootstrap resources. | Target row declares ownership mode, resource families, and lifecycle machine. |
| `destructive_cleanup` | Deletes files, directories, services, databases, buckets, or other resources. | Target row cites Section 13 predicates. |
| `runtime_reset` | Mutates test runtime state through the test-only reset boundary. | Target row cites Section 12 predicates. |
| `external_network` | Sends a caller-selected diagnostic payload to an external endpoint. | Target row cites the exact input, protocol, privacy, timeout, redirect, and failure contract. |

**TH-HARNESS-REQ-077**
The authored task-surface owner MUST declare a closed `observability_policy`
that assigns every public target exactly one explicit disposition of `required`,
`excluded`, or `out_of_scope`. An `excluded` or `out_of_scope` entry MUST have a
nonempty owner section and reason. Every required target MUST bind exactly one
stable measurement profile; its target-row command ID and resolved profile's
canonical inputs, direct or aggregate eligibility, warm-up policy, and
performance gate form the measurement identity. Parameterized
slice and audit profiles MUST use `OWNER=module.auth`; the audit profile MUST
consume that owner's retained slice evidence. A target binding MAY declare one
target-specific normalized policy transition when a shared measurement
profile's gates and canonical inputs remain unchanged; that override MUST be a
closed validator-owned transition and MUST NOT affect any sibling binding.
Duplicate target steps are invalid
until an occurrence-aware artifact contract is adopted. Generated task-surface
projections MUST preserve the policy and validation MUST enumerate the complete
public surface against it, rejecting omissions, overlap, unknown targets, and
unowned exclusions. A check-internal target MAY have a measurement-profile
binding outside the public disposition sets only when a named acceptance gate
requires that exact aggregate child; `release-browser-readiness` is the sole
such target. Execution-context capture and performance qualification MUST retain
and evaluate every measurement-profile binding, including this check-internal
subject. Defaults, runtime-family inference, and target-name inference are
forbidden. `scheduler-event-order-drift` and
`scheduler-summary-timing-drift` are explicitly `out_of_scope`: they validate
caller-selected retained evidence, so their duration describes that external
evidence selection rather than a stable command workload.
Verified by: TH-HARNESS-AC-072, TH-HARNESS-AC-073

**TH-HARNESS-REQ-078**
Every successful, failed, or interrupted top-level invocation with
`observability.disposition=required` MUST attempt local observability
finalization after native summaries and cleanup evidence are stable. One
top-level invocation produces one invocation trace. Nested public targets are
child spans selected from sequence, scheduler, summary-group, and timing
relationships; they MUST NOT become disconnected root traces merely because
their Make targets are independently public.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-074

**TH-HARNESS-REQ-079**
The current public observability and performance-maintenance commands are
`harness-observability-check`, `harness-otel-export`, and
`harness-performance-check`, plus the sole baseline writer
`harness-public-target-duration-baselines`; `explain-run` additionally accepts
`DETAIL=performance`. `harness-observability-check` reads one retained run,
reconstructs the selected invocation bundle in memory, validates retained
output and deterministic equivalence, and fails with `artifact_error`, exit
`11`, on missing, partial, malformed, unsafe, or nondeterministic evidence.
It MUST be strictly read-only and MUST NOT create a check-summary artifact in
the selected run. Check and export selection MUST name an exact retained run
directory or provide a result root together with `RUN_ID`; newest-run selection
is forbidden for these commands.
`harness-performance-check` reads the exact evidence-roots manifest supplied by
the caller and fails with `duration_baseline_drift`, exit `13`, when Section
10.5 acceptance fails.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-079

**TH-HARNESS-REQ-080**
`harness-otel-export` is the only current public harness network-export
surface. It MUST read a complete validated retained bundle without modifying
the selected run. `HARNESS_OTLP_ENDPOINT` is required and accepted from the
Make command line only. `HARNESS_OTLP_HEADERS_FILE` is optional and accepted
from the Make command line only. No ordinary test, aggregate, scheduler,
summary, finalizer, or explanation command may perform telemetry network
export.
Verified by: TH-HARNESS-AC-076

### 4.4 Direct Script and Package Boundary

| Surface                                                  | Classification                           | Contract                                                                                                  |
| -------------------------------------------------------- | ---------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Root `package.json` scripts `build`, `test`, `typecheck` | Developer convenience                    | They do not promise Make result roots, run IDs, scheduler summaries, cleanup, or machine output.          |
| `apps/web/package.json` scripts                          | Developer convenience or child command   | Browser and unit scripts become harness child work only when invoked through Make wrappers.               |
| Raw owner-local helper scripts and CLIs                  | Tool-owned diagnostics or child commands | Their direct usage and exit codes are not public harness contracts unless a Make target adopts them.      |
| Raw Go, Vitest, Playwright, Biome, Vite, pnpm commands   | Tool-owned                               | Tool output schemas remain external or diagnostic unless consumed and normalized by a Make-owned wrapper. |

### 4.5 Public Wrapper Lifecycle

Every Make-owned public wrapper that is not `interactive_raw` MUST execute this observable lifecycle:

1. establish wrapper identity and target identity;
2. resolve output mode;
3. resolve and validate harness configuration;
4. compute result-root and run-id identity if the output class declares retained artifacts;
5. initialize redaction before capturing child output;
6. run the target's semantic behavior;
7. validate required schema-owned artifacts before success;
8. select the primary failure using Section 9.1;
9. run required cleanup or finalizers;
10. emit the target's public output according to Section 7;
11. expose the normalized public exit code through retained summaries and failure output, and exit nonzero on failure.

A target MAY skip a step only when its output class or target row explicitly declares that the step does not apply. A skipped step MUST NOT be implemented as an implicit child-command side effect.

When a public Make wrapper recipe directly invokes repo-owned Node tooling before delegating to child or scheduled work, the wrapper MUST make pinned repo-local Node readiness an explicit precondition before semantic behavior begins. The current Make binding satisfies this with a `$(NODE_BIN)` prerequisite rendered from owner inputs. If the pinned Node runtime cannot be resolved, installed, downloaded, verified, or executed, the wrapper MUST fail before semantic work with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2`; it MUST NOT surface a raw shell, raw `curl`, or `env` executable-not-found failure as the public target result. Node runtime bootstrap MUST serialize mutation of the repo-local runtime and archive paths, download into temporary files, publish an archive only after checksum verification, remove corrupt or partial archive candidates, and use bounded retry for transient download failures.

### 4.6 Public Target Lifecycle

A target has one of these public-lifecycle states:

| State | Meaning | Invocation behavior |
| --- | --- | --- |
| `candidate_child` | Internal or generated child work, not a public command. | MUST NOT be required for public conformance by direct invocation. |
| `public_active` | Current public command. | MUST satisfy all public target contracts. |
| `removed` | No longer public. | MUST NOT appear in the public registry. |

A target may move to `public_active` only when it passes the semantic-value test from TH-HARNESS-REQ-058. A pre-release target may move to `removed` by revising the registry and generated mirrors in one change set. `removed` is represented by absence from the public registry, not by a retained registry row.

## 5. Configuration Resolution Contract

**TH-HARNESS-REQ-100**
Every public Make target MUST resolve harness configuration through `resolve_harness_config()` before child work begins. A target that cannot resolve or validate configuration MUST fail with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2`.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-003, TH-HARNESS-AC-014

`resolve_harness_config()` is the normative configuration-resolution contract. Repository implementation entrypoints such as preflight helpers MAY wrap this resolver, but MUST NOT define a narrower public-target configuration contract.

**TH-HARNESS-REQ-101**
Generated manifests are execution inputs, not caller configuration. A caller-supplied variable that attempts to override a non-overridable manifest field MUST fail with `configuration_error` before child work.
Verified by: TH-HARNESS-AC-002

**TH-HARNESS-REQ-102**
When a scheduler work unit invokes a child runner that starts its own worker pool, the scheduler input MUST either declare logical resource claims equal to the child worker budget or constrain that child worker budget through scheduler-owned environment. The scheduler-owned value wins for that scheduled work unit even when the same variable has a different direct public-target default. In the current check profile, scheduled `frontend-unit` MUST run Vitest with `VITEST_MAX_WORKERS=2` and MUST claim `host_cpu=2`; direct `make frontend-unit` MAY keep a faster developer default outside the check scheduler.

Auto-derived scheduler capacity MUST NOT resolve below the largest declared claim for that resource in the normalized work-unit set. Caller overrides MAY still choose lower limits, but such overrides are configuration errors when they cannot satisfy a declared work unit.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-021

**TH-HARNESS-REQ-103**
Frontend unit harness tests that depend on asynchronous jsdom rendering, workbook row hydration, inspector-subject hydration, controlled input replacement, row-history rendering, or virtualized grid mounting MUST use shared bounded wait helpers and stable selector builders with actionable diagnostics. The default wait budget MUST be finite and configuration-backed. When identity matters, the wait predicate MUST use stable workbook-row, inspector-subject, or row-history-item identity rather than visible count or text alone. An inspector-subject readiness helper MUST be observation-only after the invoking action and MUST match the expected `view_schema_id`, `record_id`, and `row_version`; it MUST NOT retry the action. Exact human diagnostic prose is non-normative; the diagnostic record MUST identify the expected row IDs, mounted row IDs, expected and mounted inspector-subject identity, received row-history item references, surface, inspector state, and failing selector class without including record payload values or reclassifying ordinary assertions away from `failure_class=product`.
Verified by: TH-HARNESS-AC-021

**TH-HARNESS-REQ-665**
Browser E2E helpers that perform a mutating UI action and then drive another action that depends on the committed result MUST wait for the server success response and for the rendered workbook projection to converge on the response's stable source-record identity before continuing. When the response supplies `source_record.row_version`, convergence MUST require the rendered source row to reach at least that version; a concurrent accepted version above the response version satisfies the floor, while a stale lower version does not. When the dependent action relies on optimistic concurrency, convergence MUST include the returned `row_version` rendered under the stable row identifier. A visible global save-state label such as `Saved` MAY be asserted after convergence, but it MUST NOT be the only completion predicate for a dependent mutation sequence.

A helper that validates post-mutation focus or viewport continuity MUST be observation-only after the mutating action: it MUST NOT focus, scroll, click, press a key, or dispatch an input event to manufacture the postcondition it is measuring. Setup-time navigation and scrolling before the action remain allowed. A later passing invocation is distinct evidence and MUST NOT retry, replace, or reclassify an earlier product assertion failure.

This requirement owns browser synchronization and evidence only. Core 03 `REQ-03-283` remains the unchanged product authority for deterministic row-local focus restoration after same-surface follow-up rendering.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-021, TH-HARNESS-AC-080

Browser E2E tests that need the authenticated incident directory MUST navigate to that surface explicitly before asserting directory UI. Tests MUST NOT assume raw authenticated `/` always renders the incident directory, because Core 01 `REQ-01-580` makes raw root cardinality-sensitive and requires the sole visible incident to auto-open.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-021

**TH-HARNESS-REQ-115**
Browser E2E helpers that select a workbook surface through a menu, popover, selector, or tab strip MUST use stable selector builders rooted in canonical `view_schema_id` or `sheet_ref` identity. Such a helper MUST use bounded retries, reacquire locators after every render-sensitive action, require a single target option before selecting, and after selection converge on the active workbook shell surface signal, the canonical direct `view_schema_id` URL representation when selecting a base surface, and the target grid shell before returning. Final helper diagnostics MUST include the requested target identity, current URL, active shell surface, menu-open state, visible candidate target identities, and final retry error. This requirement governs browser-test synchronization only and MUST NOT define product behavior beyond the owning product specifications.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-021

**TH-HARNESS-REQ-111**
Make-owned frontend dependency installation MUST use the pinned repo-local Node and pnpm toolchain, MUST bind pnpm's content-addressable store to the repo-local `.pnpm-store` path through project configuration, MUST run without requiring a TTY or interactive confirmation, and MUST use a frozen lockfile. `frontend-install` is an install/readiness target, not a dependency-update target; if `pnpm-lock.yaml` is out of sync with workspace manifests, the target MUST fail with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2` rather than mutating the lockfile. A package-manager repair that purges and recreates repo-local `node_modules` is allowed only as part of this non-interactive install contract and only for repo-local workspace dependency roots.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-014, TH-HARNESS-AC-023

**TH-HARNESS-REQ-116**
`OWNER` is required for every owner-first public command. It has no default, uses `trim` normalization, remains case-sensitive after trimming, and MUST resolve to exactly one active owner. Missing, empty, malformed, inactive, and unknown values are `usage_error` with public exit `2` before setup.
Verified by: TH-HARNESS-AC-064

**TH-HARNESS-REQ-117**
`ROWS` is optional only for the two slice commands. The parser MUST retain raw comma-separated tokens until it rejects empty tokens and duplicates after trimming. Accepted values MUST be active row IDs owned by the exact `OWNER`, then normalized to unique ASCII-bytewise order. `ROWS=` and unknown, malformed, duplicate, inactive, cross-owner, or non-executable rows are `usage_error` with exit `2`. No valid subset may run when any requested token is invalid.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-065

**TH-HARNESS-REQ-118**
Omitted `ROWS` has the exact selection behavior in TH-HARNESS-REQ-360. `default_check` affects `make check` only and MUST NOT narrow an owner slice. A selection that resolves to zero rows is `usage_error`; no support-only or migration exception exists in the adopted profile.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-068

**TH-HARNESS-REQ-119**
`VITEST_MAX_WORKERS` defaults to decimal `4`; `PLAYWRIGHT_WORKERS` defaults to decimal `3`. Each accepts only a base-10 integer from `1` through `16`, inclusive. Validation occurs even when the selection contains no matching runner. A valid unused value MUST be recorded in `unused_inputs` and MUST NOT change another runner or scheduler resource limit.
Verified by: TH-HARNESS-AC-064

**TH-HARNESS-REQ-120**
Target-local `JSON` accepts only exact `1`; omitted or empty means human output. `JSON=1` and `CARTULARY_OUTPUT_MODE=machine` together are `usage_error` with exit `2`. For `test-slice` and `service-backed-test-slice`, `JSON=1` changes stdout only and execution still occurs; successful stdout is exactly one `cartulary.test_slice_scheduler_summary.v1` object followed by LF. For `explain-test-owner` and `task-guide`, successful stdout is exactly one command-specific schema object followed by LF.
Verified by: TH-HARNESS-AC-004, TH-HARNESS-AC-064

**TH-HARNESS-REQ-121**
`test-evidence-audit` requires `EVIDENCE_ROOTS_FILE`, a caller-owned
`cartulary.test_evidence_root_manifest.v1` file containing the exact owner ID and
ASCII-sorted unique `{target_id, run_root}` entries to audit. The auditor MUST derive
the required target partitions from the current catalog and verification contracts;
it MUST NOT widen the manifest, search sibling roots, or select a newest run. One
physical run root MAY be named by multiple explicit target entries. A known supplied
target that is not applicable to the selected owner MUST be reported in
`unused_inputs`; an unknown or duplicate target is `usage_error`. Missing required
target entries are `usage_error`; unsafe roots or incompatible contents are
`artifact_error`. As an atomic alternative to leaf target partitions, a manifest
MAY contain one `target_id=test-slice` entry whose retained accounting artifact
selects the owner's complete active row set. In that mode the auditor MUST verify
the full-owner selection and all compatibility fields directly from that artifact;
it MUST treat any additional supplied target entries as unused and MUST NOT infer,
split, or search for leaf evidence. A partial or service-backed-only slice MUST NOT
satisfy the full-owner alternative.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-066

**TH-HARNESS-REQ-122**
`task-guide` requires exact `ROLE=module-author` and a valid `OWNER`; neither has a default. Every other role token and every delivery-phase input is unsupported in the current task-guide contract.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-123**
Any Make command-line input not declared by the selected target MUST fail before child work as `usage_error`; an undeclared inherited environment value MUST be ignored and MUST NOT reach child environments, as required by TH-HARNESS-REQ-112.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-070

### 5.1 Precedence

| Precedence | Source                                 | Rule                                                                                                   |
| ---------: | -------------------------------------- | ------------------------------------------------------------------------------------------------------ |
|          1 | Make-owned wrapper CLI flags           | Highest priority only for flags explicitly declared by the target.                                     |
|          2 | Make command-line variables            | `VAR=value make target` overrides inherited environment for the same canonical variable.               |
|          3 | Exported environment inherited by Make | Accepted only for variables declared in the configuration table.                                       |
|          4 | Target manifest values                 | Source inputs for scheduler and target behavior, not caller overrides.                                 |
|          5 | Config files                           | Apply only to the application-under-test runtime unless the variable table declares a harness binding. |
|          6 | Hardcoded harness defaults             | Used only when all higher layers omit the value.                                                       |

### 5.2 Configuration Algorithm

```text
resolve_harness_config(target, raw_make_vars, raw_env, wrapper_cli_args):
  assert target is a public Make target or a Make-owned wrapper target
  declared = global_configuration_table + per_target_input_registry[target]
  resolved = empty map

  reject undeclared wrapper CLI flags
  reject caller overrides of manifest/internal fields
  reject undeclared public harness Make variables supplied on the Make command line

  for each declared variable in stable table order:
    candidates = [
      wrapper_cli_args value if variable has declared CLI binding,
      raw_make_vars value if supplied on the Make command line,
      raw_env value if exported before Make invocation,
      manifest value if variable has declared manifest binding,
      config_file value if variable has declared config-file binding,
      hardcoded default
    ]

    select the first candidate whose layer is allowed for the variable
    apply the variable's empty-string rule
    normalize the selected value
    validate the normalized value
    if validation fails:
      emit configuration_error summary when the target has a summary layer
      fail before child work with exit code 2
    record selected value, source layer, and normalized value

  ignore undeclared inherited environment variables
  strip undeclared public harness variables from child process environments
  emit resolved values required by Section 8 summaries
```

### 5.3 Per-Target Input Registry

**TH-HARNESS-REQ-112**
Every public target MUST declare a closed per-target input contract. This NLSpec owns the current public target-local input contract. Authored `tools/task_surface_owner.json` MUST use `cartulary.task_surface_owner.v1` or a later adopted owner schema, and generated `tools/task_surface_manifest.json` MUST use `cartulary.task_surface_manifest.v15` or a later adopted projection schema. Both MUST contain `input_contract` for every row with `target_class="public"`; neither may independently widen, narrow, or reinterpret the closed contract below. Execution topology may reference a target name but MUST NOT own or override its input contract.

Make-to-wrapper source transport is private. The current transport is one
`CARTULARY_MAKE_INPUT_SOURCES` value containing whitespace-separated
`NAME=cli|env|file|unset` entries for the closed public input inventory. Names MUST be
unique safe Make variable names, and unknown names or source tokens MUST fail before
child work. Per-variable `CARTULARY_MAKE_ORIGIN_<NAME>` variables are unsupported and
MUST NOT affect source resolution.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-002, TH-HARNESS-AC-027, TH-HARNESS-AC-058

Each `input_contract` MUST contain:

| Field | Required value |
| --- | --- |
| `undeclared_make_command_line` | `usage_error` |
| `undeclared_inherited_env` | `ignore` |
| `inputs[]` | Stable ordered array of accepted target-local inputs. Empty array means the public target accepts no target-local Make variables. |

Each `inputs[]` row MUST contain `name`, `binding`, `allowed_sources`, `required`, `type`, `default`, `empty_string`, `normalization`, `invalid_reason`, `summary_emission`, and `child_forwarding`. Rows MAY additionally contain bounded type metadata such as `values`, `min`, or `max`.

| Row field | Meaning |
| --- | --- |
| `name` | Uppercase Make variable name accepted by this target. |
| `binding` | Public invocation binding. The current profile accepts `make_variable`; a later profile MAY add wrapper CLI bindings only when Section 5.2 precedence remains preserved. |
| `allowed_sources` | Subset of `make_command_line`, `environment`, `makefile_default`, `internal_default`, and `manifest`. A source not listed for the row MUST NOT supply the value. |
| `required` | Whether omission after all allowed sources is a usage/configuration failure. |
| `type` | One of the Section 5.3 type tokens. |
| `default` | Default value or `null`; defaults are valid only when their source is declared. |
| `empty_string` | One of `invalid`, `omitted`, or `false`. |
| `normalization` | One of `none`, `trim`, `trim_lowercase`, or `path_token`. |
| `invalid_reason` | `usage_error` for caller selection mistakes or `configuration_error` for invalid paths, retained evidence, internal state, or manifest-derived configuration. |
| `summary_emission` | One of `none`, `value`, `redacted_value`, or `source_and_value`. |
| `child_forwarding` | One of `none`, `argv`, `runtime_env`, or `argv_and_runtime_env`; undeclared public harness inputs MUST NOT reach child environments. |

The closed target-local public input set in the current profile consists only of documented uses of `ROLE`, `OWNER`, `ROWS`, `TARGET`, `RESULTS_DIR`, explicit retained-evidence root selectors, `ALLOW_OLDER_RESULTS_DIR`, `RUN_ID`, `DETAIL`, `JSON`, worker controls, fixture report limits, duration-maintenance knobs, scheduler timing knobs, destructive-safety controls, the explicit Govulncheck database override, and the explicit `HARNESS_OTLP_ENDPOINT` and `HARNESS_OTLP_HEADERS_FILE` post-run export inputs. A public target accepts one of these names only when it appears in the normative input matrix below.

`frontend-fallow-static` accepts no target-local Make variables in the current Fallow static profile. A future changed-code audit base such as `FALLOW_CHANGED_SINCE` MUST be added to this registry before it becomes public input.

**TH-HARNESS-REQ-114**
Every public target's accepted target-local input set is closed by the normative input matrix. A public target that is not listed in the matrix accepts no target-local Make variables beyond the global variables in Section 5.5. A grouped row is valid only when every listed target has identical `type`, `default`, `allowed_sources`, `required`, `empty_string`, `normalization`, `values` or `min`/`max` bounds, `invalid_reason`, `summary_emission`, and `child_forwarding`.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-029

| Target(s) | Input | Type | Required | Allowed sources | Default | Omission behavior | Empty-string behavior | Normalization | Values/bounds | Invalid behavior | Summary emission | Child forwarding |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `db-reset`, `services-down`, `object-store-reset`, `clean`, `distclean` | `CARTULARY_CLEANUP_DRY_RUN` | `exact_1_bool` | no | Make command line, environment, Makefile default | `false` | false | false | `trim` | exact `1` means true | `usage_error`, exit `2` | value | runtime env |
| `db-reset` | `CARTULARY_DESTRUCTIVE_CONFIRM` | `enum` | no | Make command line only | none | omitted allowed only for dry-run | invalid | `trim` | `db-reset` | `usage_error`, exit `2` | value | none |
| `object-store-reset` | `CARTULARY_DESTRUCTIVE_CONFIRM` | `enum` | no | Make command line only | none | omitted allowed only for dry-run | invalid | `trim` | `object-store-reset` | `usage_error`, exit `2` | value | none |
| `agent-finalize` | `ALLOW_OLDER_RESULTS_DIR` | `exact_1_bool` | no | Make command line, environment, Makefile default | none | older retained root rejected | false | `trim` | exact `1` means true | `usage_error`, exit `2` | value | runtime env |
| `agent-finalize` | `RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | none | actions that require retained evidence are not selected | omitted | `path_token` | existing retained full warm `check` run root when supplied | `usage_error`, exit `2` | value | runtime env |
| `test-evidence-audit` | `OWNER` | `owner_id` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `trim` | active Section 3 owner ID | `usage_error`, exit `2` | value | runtime env |
| `test-evidence-audit` | `EVIDENCE_ROOTS_FILE` | `evidence_root_manifest` | yes | Make command line, environment, Makefile default | none | missing manifest | invalid | `path_token` | existing non-symlink regular file containing `cartulary.test_evidence_root_manifest.v1` for exact `OWNER` | `usage_error`, exit `2` when missing or malformed | value | argv |
| `task-surface-report` | `TASK_SURFACE_REPORT_ARGS` | `task_surface_report_args` | no | Make command line, environment, Makefile default | none | compact default report | omitted | `trim` | empty, `--all`, `--check`, `--check --all`, `--all --check` | `usage_error`, exit `2` | value | argv |
| `task-guide` | `ROLE` | `enum` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `trim` | exact `module-author` | `usage_error`, exit `2` | value | argv |
| `task-guide`, `test-slice`, `service-backed-test-slice`, `explain-test-owner` | `OWNER` | `owner_id` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `trim` | active Section 3 owner ID | `usage_error`, exit `2` | value | argv |
| `test-slice`, `service-backed-test-slice` | `ROWS` | `test_row_ids` | no | Make command line, environment, Makefile default | none | Section 5 and TH-HARNESS-REQ-360 owner default | invalid | `trim` | comma-separated active row IDs owned by exact `OWNER` | `usage_error`, exit `2` | value | argv |
| `test-slice`, `service-backed-test-slice` | `VITEST_MAX_WORKERS` | `positive_integer` | no | Make command line, environment, Makefile default | `4` | `4` | invalid | `trim` | `1..16`; affects selected Vitest work only | `usage_error`, exit `2` | value | runtime env |
| `test-slice`, `service-backed-test-slice` | `PLAYWRIGHT_WORKERS` | `positive_integer` | no | Make command line, environment, Makefile default | `3` | `3` | invalid | `trim` | `1..16`; affects selected Playwright work only | `usage_error`, exit `2` | value | runtime env |
| `task-guide`, `test-slice`, `service-backed-test-slice`, `fixture-report`, `explain-test-owner`, `explain-target` | `JSON` | `exact_1_bool` | no | Make command line, environment, Makefile default | `false` | human output | false | `trim` | exact `1` means JSON target-local output | `usage_error`, exit `2` | value | argv |
| `frontend-unit` | `VITEST_MAX_WORKERS` | `positive_integer` | no | Make command line, environment, Makefile default | `4` | `4` | invalid | `trim` | `1..16` | `usage_error`, exit `2` | value | runtime env |
| `target-plan`, `target-plan-json`, `fixture-report`, `explain-run`, `scheduler-event-order-drift`, `scheduler-summary-timing-drift` | `TARGET` | `target_name` | no | Make command line, environment, Makefile default | none | all target rows accepted by the target | omitted | `trim` | public or scheduler target name present in the task-surface manifest | `usage_error`, exit `2` | value | argv |
| `fixture-report` | `RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | `.cartulary/test-results` | default result root | omitted | `path_token` | existing result root or retained run root | `usage_error`, exit `2` | value | argv |
| `fixture-report`, `explain-run` | `RUN_ID` | `run_id` | no | Make command line, environment, Makefile default | none | latest retained run under selected result root for human investigation only | omitted | `trim` | Section 6.2 `run_id_v1` | `usage_error`, exit `2` | value | argv |
| `harness-observability-check`, `harness-otel-export` | `RUN_ID` | `run_id` | conditionally yes | Make command line only | none | allowed only when `RESULTS_DIR` is itself one exact retained run directory | invalid when `RESULTS_DIR` is a result root | `trim` | Section 6.2 `run_id_v1` | `usage_error`, exit `2` | value | argv |
| `fixture-report` | `FIXTURE_THRESHOLD_MS` | `positive_integer` | no | Make command line, environment, Makefile default | `30000` | `30000` | omitted | `trim` | `1..999999999` | `usage_error`, exit `2` | value | argv |
| `fixture-report` | `FIXTURE_TOP` | `positive_integer` | no | Make command line, environment, Makefile default | `5` | `5` | omitted | `trim` | `1..999999999` | `usage_error`, exit `2` | value | argv |
| `explain-run`, `harness-observability-check`, `harness-otel-export`, `go-test-duration-baselines`, `browser-e2e-duration-baselines`, `service-backed-make-target-duration-baselines`, `harness-smoke-duration-baselines` | `RESULTS_DIR` | `result_selector` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `path_token` | existing result root or retained run root; check/export additionally obey the exact-selection `RUN_ID` row | `usage_error`, exit `2` | value | argv |
| `explain-run` | `DETAIL` | `enum` | no | Make command line, environment, Makefile default | `summary` | `summary` | omitted | `trim` | `summary`, `children`, `logs`, `progress`, `accounting`, `performance` | `usage_error`, exit `2` | value | argv |
| `harness-otel-export` | `HARNESS_OTLP_ENDPOINT` | `url` | yes | Make command line only | none | missing required input | invalid | `trim` | HTTPS base URL or loopback HTTP URL with no credentials, query, or fragment | `usage_error`, exit `2` | redacted value | argv |
| `harness-otel-export` | `HARNESS_OTLP_HEADERS_FILE` | `path` | no | Make command line only | none | no extra headers | omitted | `path_token` | owner-only non-symlink regular JSON file with bounded string header values | `configuration_error`, exit `2` | redacted value | argv |
| `harness-performance-check` | `EVIDENCE_ROOTS_FILE` | `path` | yes | Make command line only | none | missing manifest | invalid | `path_token` | existing non-symlink regular `cartulary.harness_performance_evidence_roots.v2` manifest with `mode=comparison`, explicit deduplicated reference/candidate windows, and exact target bindings | `usage_error`, exit `2` when missing or malformed | value | argv |
| `harness-public-target-duration-baselines` | `EVIDENCE_ROOTS_FILE` | `path` | yes | Make command line only | none | missing baseline window | invalid | `path_token` | existing non-symlink regular `cartulary.harness_performance_evidence_roots.v2` manifest with `mode=baseline`, explicit target/provider bindings, one accepted warm-up root, and exactly two accepted measured roots per window | `usage_error`, exit `2` when missing or malformed | value | argv |
| `explain-target` | `TARGET` | `target_name` | yes | Make command line, environment, Makefile default | none | missing required input | invalid | `trim` | public or scheduler target name present in the task-surface manifest | `usage_error`, exit `2` | value | argv |
| `explain-target` | `DETAIL` | `enum` | no | Make command line, environment, Makefile default | `summary` | `summary` | omitted | `trim` | `summary`, `rows`, `artifacts` | `usage_error`, exit `2` | value | argv |
| `go-test-duration-baselines` | `PRUNE_OBSERVED_PACKAGES` | `exact_1_bool` | no | Make command line, environment, Makefile default | `false` | false | false | `trim` | exact `1` means true | `usage_error`, exit `2` | value | argv |
| `go-test-duration-baselines` | `ALLOW_COMMAND_OVERHEAD_DECREASE` | `exact_1_bool` | no | Make command line, environment, Makefile default | `false` | false | false | `trim` | exact `1` means true | `usage_error`, exit `2` | value | argv |
| `go-test-duration-baselines`, `go-test-duration-baseline-coverage`, `go-test-duration-baseline-drift` | `GO_TEST_DURATION_BASELINE` | `path` | no | Make command line, environment, Makefile default | none | omitted caller input lets the Makefile default source supply `tools/go_test_duration_baselines.json` | omitted | `path_token` | filesystem path token | `configuration_error`, exit `2` | value | argv |
| `go-test-duration-baseline-drift`, `browser-e2e-duration-baseline-drift`, `service-backed-make-target-duration-baseline-drift`, `harness-smoke-duration-baseline-drift`, `scheduler-event-order-drift`, `scheduler-summary-timing-drift` | `RESULTS_DIR` | `result_selector` | no | Make command line, environment, Makefile default | `current-run` | current retained run when available | omitted | `path_token` | existing result root or retained run root | `usage_error`, exit `2` | value | argv |
| `browser-e2e-duration-baselines`, `browser-e2e-duration-baseline-drift` | `BROWSER_E2E_DURATION_BASELINE` | `path` | no | Make command line, environment, Makefile default | none | omitted caller input lets the Makefile default source supply `tools/browser_e2e_duration_baselines.json` | omitted | `path_token` | filesystem path token | `configuration_error`, exit `2` | value | argv |
| `service-backed-make-target-duration-baselines`, `service-backed-make-target-duration-baseline-drift` | `SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE` | `path` | no | Make command line, environment, Makefile default | none | omitted caller input lets the Makefile default source supply `tools/service_backed_make_target_duration_baselines.json` | omitted | `path_token` | filesystem path token | `configuration_error`, exit `2` | value | argv |
| `harness-smoke-duration-baselines`, `harness-smoke-duration-baseline-drift` | `HARNESS_SMOKE_DURATION_BASELINE` | `path` | no | Make command line, environment, Makefile default | none | omitted caller input lets the Makefile default source supply `tools/harness_smoke_duration_baselines.json` | omitted | `path_token` | filesystem path token | `configuration_error`, exit `2` | value | argv |
| `scheduler-summary-timing-drift` | `SCHEDULER_WARM_CHECK_BUDGET_MS` | `positive_integer` | no | Make command line, environment, Makefile default | `155000` | `155000` | omitted | `trim` | `1..999999999` | `usage_error`, exit `2` | value | argv |
| `scheduler-summary-timing-drift` | `SCHEDULER_WARM_CHECK_BALANCE_RATIO` | `positive_decimal` | no | Make command line, environment, Makefile default | `1.25` | `1.25` | omitted | `trim` | `>=1` | `usage_error`, exit `2` | value | argv |
| `build-operator` | `OPERATOR_BIN` | `path` | no | Make command line, environment, Makefile default | `$(CURDIR)/operator` | build writes the default repo-local operator binary | invalid | `path_token` | concrete build-output file path; not `/`, `.`, `..`, not under protected repo roots, no NUL, no POSIX backslash, not an existing directory, and not a symlink path | `configuration_error`, exit `2` | source_and_value | none |
| `build-server-harness` | `SERVER_HARNESS_BIN` | `path` | no | Make command line, environment, Makefile default | `$(CURDIR)/server-harness` | build writes the harness-profile server artifact | invalid | `path_token` | concrete build-output file path; not `/`, `.`, `..`, not under protected repo roots, no NUL, no POSIX backslash, not an existing directory, and not a symlink path | `configuration_error`, exit `2` | source_and_value | none |
| `go-vulncheck` | `GOVULNCHECK_DB` | `path` | no | Make command line, environment, Makefile default | none | Govulncheck default DB | omitted | `path_token` | optional Govulncheck vulnerability DB path or endpoint token | `usage_error`, exit `2` | value | runtime env |

`fixture-report` remains a `human_summary` target by default. `JSON=1` selects the target-local diagnostic JSON path and is not equivalent to `CARTULARY_OUTPUT_MODE=machine`. When `JSON=1`, stdout MUST be exactly one `cartulary.fixture_report.v1` JSON object followed by one LF, and stderr follows the Section 7 failure budget for `human_summary` targets. `CARTULARY_OUTPUT_MODE=machine make fixture-report` MUST continue to fail before child work under Section 7.2 unless a later adopted registry row changes the target's output class.

**TH-HARNESS-REQ-113**
Undeclared public harness inputs MUST have one shared result:

| Caller input class | Required behavior |
| --- | --- |
| Undeclared wrapper CLI flag | Reject before child work with `failure_reason=usage_error`, exit `2`. |
| Undeclared public harness Make variable supplied on the Make command line | Reject before child work with `failure_reason=usage_error`, exit `2`. |
| Undeclared inherited environment variable | Ignore for resolution and strip from child process environments. |
| Caller override of manifest/internal fields | Reject before child work with `failure_reason=configuration_error`, exit `2`. |

Manifest and internal fields include at least `TASK_SURFACE_MANIFEST`, `CARTULARY_TASK_SURFACE_MANIFEST`, `EXECUTION_TOPOLOGY_MANIFEST`, `CARTULARY_EXECUTION_TOPOLOGY_MANIFEST`, `SCHEDULER_MANIFEST`, and `CARTULARY_OPERATOR_BIN` when supplied through public Make command-line variables. Script-level environment fallbacks such as broad manifest-path overrides, broad passthrough argument strings such as `VITEST_FLAGS`, or unbounded threshold variables are non-canonical implementation inputs unless a public target row declares a bounded `input_contract` entry.

For public `frontend-unit` evidence, `VITEST_MAX_WORKERS` is the only current public Vitest input. `VITEST_FLAGS` MUST NOT be accepted from the Make command line for `frontend-unit`; inherited `VITEST_FLAGS` from the caller environment MUST be stripped before child Vitest execution and MUST NOT narrow the canonical runner report. Filtered Vitest diagnostics, when needed, are private developer commands outside public harness evidence.
Verified by: TH-HARNESS-AC-001, TH-HARNESS-AC-002, TH-HARNESS-AC-003

| Type token | Valid values |
| --- | --- |
| `enum` | One of the row's `values[]` tokens after normalization. |
| `exact_1_bool` | Exact `1` when true; empty string is false only when the row says `empty_string=false`. |
| `owner_id` | One active owner ID satisfying TH-HARNESS-REQ-012. |
| `test_row_ids` | Comma-separated active row IDs parsed by TH-HARNESS-REQ-117. |
| `target_name` | A target name present in the task-surface manifest. |
| `run_id` | `run_id_v1` from Section 6.2. |
| `result_selector` | Existing result root or retained run-root path accepted by the target. |
| `path` | Filesystem path token; path existence and safety follow the Section 5.3 matrix row's `Values/bounds` and `Invalid behavior` cells unless a later adopted row narrows them. |
| `positive_integer` | Decimal integer greater than zero and inside row bounds when declared. |
| `positive_decimal` | Decimal number greater than zero and inside row bounds when declared. |
| `task_surface_report_args` | Empty string, `--all`, `--check`, `--check --all`, or `--all --check`. |

### 5.4 Empty-String Rules

| Variable family                              | Empty string behavior                                                                |
| -------------------------------------------- | ------------------------------------------------------------------------------------ |
| Output mode                                  | Treated as omitted; default resolution applies.                                      |
| Result root                                  | Invalid.                                                                             |
| Run ID                                       | Invalid.                                                                             |
| Boolean exact-`1` flags                      | Empty string is false.                                                               |
| Integer limits                               | Treated as omitted; default applies.                                                 |
| Required DSN, endpoint, credential, or token | Invalid.                                                                             |
| Optional config path                         | Treated as omitted.                                                                  |
| Comma-separated lists                        | Empty string is an empty list only when the variable row says so; otherwise invalid. |

### 5.5 Configuration Variable Table

| Name or family                                                                                  | Scope                   | Type and valid values                                                                                                 | Default                                                                                   | Allowed sources                                 | Empty-string behavior                                   | Normalization                                                                                                 | Invalid behavior                                                                   | Summary emission                                   |
| ----------------------------------------------------------------------------------------------- | ----------------------- | --------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | -------------------------------------------------- |
| `CARTULARY_TEST_RESULTS_DIR`                                                                    | global                  | `result_root_path_v1` from Section 6                                                                                  | `.cartulary/test-results`                                                                 | Make variable, env, default                     | invalid                                                 | Path normalized by Section 6                                                                                  | `configuration_error`, exit `2`                                                    | normalized path and cleanup scope                  |
| `CARTULARY_TEST_RUN_ID`                                                                         | global                  | `run_id_v1` from Section 6                                                                                            | generated by Section 6                                                                    | Make variable, env, default                     | invalid                                                 | Grammar validation only                                                                                       | `configuration_error`, exit `2`                                                    | run ID and whether generated                       |
| `CARTULARY_OUTPUT_MODE`                                                                         | global                  | `quiet`, `summary`, `ci`, `verbose`, `debug`, `machine`                                                               | resolved by Section 7                                                                     | Make variable, env, default                     | omitted                                                 | lower-case exact token                                                                                        | `configuration_error`, exit `2`                                                    | resolved output mode and source                    |
| `VERBOSE`                                                                                       | global                  | exact `1` means verbose request; any other value false                                                                | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` ignored as false                                                           | boolean source when true                           |
| `CI_VERBOSE`                                                                                    | global                  | exact `1` means CI-output request; any other value false                                                              | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` ignored as false                                                           | boolean source when true                           |
| `CI`                                                                                            | global                  | exact `1` marks CI environment                                                                                        | false                                                                                     | env, default                                    | false                                                   | exact string compare                                                                                          | non-`1` ignored as false                                                           | boolean source when true                           |
| Scheduler resource limits                                                                       | scheduler               | positive integer `1..256` unless resource row declares narrower bound                                                 | resource registry default                                                                 | CLI flag, Make variable, env, manifest, default | omitted                                                 | decimal parse with no separators                                                                              | `configuration_error`, exit `2`                                                    | normalized limit and source                        |
| `--resource-limit name=value`                                                                   | scheduler               | declared resource name and positive integer value                                                                     | none                                                                                      | CLI flag only                                   | invalid                                                 | name exact; value decimal                                                                                     | `usage_error` for malformed flag, `configuration_error` for invalid declared value | normalized override                                |
| `GO`, `GO_CACHE_DIR`, `GO_MOD_CACHE_DIR`, `GOCACHE`, `GOMODCACHE`                               | toolchain               | executable path or filesystem path                                                                                    | Go auto-discovered; external caches `/tmp/cartulary-go-build` and `/tmp/cartulary-go-mod` | Make variable, env, default                     | invalid for paths, omitted for executable               | path normalization by target helper                                                                           | `configuration_error`, exit `2`                                                    | executable path and cache path unless redacted     |
| `NODE_VERSION`, `PNPM_VERSION`, `NODE_RUNTIME_DIR`, `NODE_BIN`, `PNPM`, `COREPACK_HOME`, `PATH` | toolchain               | version token or filesystem path                                                                                      | Node `24.15.0`, pnpm `10.33.0`, repo-local `tmp/node-runtime`                             | Make variable, env, default                     | invalid for paths/versions unless row-specific optional | exact version token; path normalization                                                                       | `configuration_error`, exit `2`                                                    | version and runtime path                           |
| `CARTULARY_READINESS_CACHE_DIR`, `CARTULARY_READINESS_DISABLE_CACHE`, `CARTULARY_FORCE_REINSTALL` | harness cache           | repo-local path for cache dir; exact `1` for disable or force reinstall                                                | `.cache/cartulary/readiness`; false; false                                                 | Make variable, env, default                     | invalid for path; false for boolean flags                | path normalization; exact string compare for flags                                                           | invalid path is `configuration_error`; non-`1` flags are false                     | cache state and record path only                   |
| `CARTULARY_BUILD_CACHE_DIR`, `CARTULARY_BUILD_CACHE_DISABLE`, `CARTULARY_FORCE_REBUILD`          | harness cache           | repo-local path for cache dir; exact `1` for disable or force rebuild                                                  | `.cache/cartulary/build-artifacts`; false; false                                           | Make variable, env, default                     | invalid for path; false for boolean flags                | path normalization; exact string compare for flags                                                           | invalid path is `configuration_error`; non-`1` flags are false                     | cache state and record path only                   |
| `CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR`, `CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE`      | harness cache           | repo-local path for cache dir; exact `1` disables action cache                                                         | `.cache/cartulary/agent-finalize-action-cache`; false                                      | env, default                                    | invalid for path; false for boolean flag                 | path normalization; exact string compare for flag                                                            | invalid path is `configuration_error`; non-`1` flag is false; `CI=1` disables cache | action cache state in finalizer summary            |
| `CARTULARY_EXECUTION_TOPOLOGY_RENDER_CACHE_DIR`, `CARTULARY_EXECUTION_TOPOLOGY_RENDER_DISABLE_CACHE` | harness cache        | repo-local path for cache dir; exact `1` disables render cache                                                         | `.cache/cartulary/execution-topology-render`; false                                        | env, default                                    | invalid for path; false for boolean flag                 | path normalization; exact string compare for flag                                                            | invalid path is `configuration_error`; non-`1` flag is false                       | debug/cache diagnostics only                       |
| `CONFIG_FILE`, `CARTULARY_CONFIG_FILE`                                                          | app runtime             | config file path                                                                                                      | `configs/dev/config.toml` for local/dev/browser targets                                   | Make variable, env, config binding, default     | omitted                                                 | path normalization; `CARTULARY_CONFIG_FILE` wins only inside application runtime when both are passed through | harness invalid path: `configuration_error`; app invalid config: target failure    | path, not file contents                            |
| `TEST_SERVICES_BIN`, `CARTULARY_TEST_SERVICES_BIN`                                              | service suite           | executable path                                                                                                       | `tmp/toolbin/cartulary-test-services`                                                     | Make variable, env, default                     | invalid                                                 | path normalization                                                                                            | `configuration_error`, exit `2`                                                    | normalized path                                    |
| `CARTULARY_OPERATOR_BIN`                                                                         | runtime binary          | scheduler-owned executable path for operator scenario Go tests                                                        | produced by `build-operator` from `OPERATOR_BIN`; current default `operator`              | scheduler/runtime wiring only; not public Make command line | invalid for canonical scheduler-selected operator scenario work | path normalization; existing regular executable file; symlinks rejected | missing, empty, non-regular, non-executable, or caller command-line override is `configuration_error`, exit `2`; build-artifact digest/provenance mismatch is `artifact_error`, exit `11` | source, normalized path, producer target, file digest, build-artifact reference |
| `CARTULARY_TEST_SERVICES_ACTIVE`                                                                | service suite           | exact `1` selects attach mode                                                                                         | owned mode                                                                                | env, Make variable, default                     | false                                                   | exact string compare                                                                                          | non-`1` selects owned mode                                                         | mode only                                          |
| `CARTULARY_TEST_SUITE_ID`, `CARTULARY_TEST_TARGET`                                              | service suite           | non-empty ASCII token; suite ID is 24 lowercase hex in owned mode                                                     | generated in owned mode                                                                   | service manifest, env in attach mode            | invalid in attach mode                                  | exact grammar validation                                                                                      | `configuration_error`, exit `2`                                                    | suite ID, target                                   |
| Postgres attach set                                                                             | service suite           | `CARTULARY_PGTEST_ADMIN_DSN`, `CARTULARY_PGTEST_DSN_TEMPLATE` containing `{database}`, `CARTULARY_PGTEST_TEMPLATE_DB`, optional `CARTULARY_PGTEST_SCHEMA_HASH` | none                                                                                      | env, Make variable                              | invalid                                                 | DSN redacted; template exact placeholder validation; schema hash exact match when supplied                     | partial or malformed set or schema-hash mismatch: `configuration_error`, exit `2`  | redacted DSN, attach mode, schema hash             |
| Postgres fixture policy                                                                         | service suite           | `template_clone`, `transaction`, `package_reset`, `group_clone`, `migration_scratch`; table lists use comma-separated SQL identifiers | target-plan policy; direct helper fallback `template_clone`                                | Make variable, env, manifest, default           | omitted                                                 | lower-case exact token; table identifiers sorted for summaries                                                | unknown policy or bad identifier: `configuration_error`, exit `2`                  | policy, fixture class, schema hash, table count    |
| Object-store S3 attach set                                                                      | service suite           | endpoint, access key, secret key, secure bool through `CARTULARY_S3TEST_*`                                            | none                                                                                      | env, Make variable                              | invalid for required members                            | endpoint normalized; credentials redacted; secure bool exact `true`/`false` or `1`/`0`                        | partial set or invalid bool: `configuration_error`, exit `2`                       | endpoint, secure flag, credential redaction marker |
| `CARTULARY_TEST_SERVICES_WEB_E2E_CLEANUP_WORKERS`                                               | browser/service cleanup | integer `1..16`                                                                                                       | `4`                                                                                       | Make variable, env, default                     | omitted                                                 | decimal parse                                                                                                 | invalid value falls back to `4` and records warning                                | resolved value and warning when fallback used      |
| Compose env                                                                                     | local services          | `CARTULARY_COMPOSE_FILE`, ready timeouts, `OBJECT_STORE_BUCKET`                                                       | `docker-compose.dev.yml`, Postgres `180s`, object-store `120s`, bucket `cartulary`        | Make variable, env, default                     | omitted for optional values                             | path and duration normalization                                                                               | missing Docker/Compose: Section 9 class                                            | non-secret values                                  |
| Browser owned-stack env                                                                         | browser                 | runtime roots, origins, backend/frontend port overrides, built frontend preview artifact                               | dynamic ports; frontend served from `apps/web/dist` by non-watching preview; `build-web` is a first-class prerequisite | Make variable, env, manifest, default           | invalid for required values or missing built frontend artifact | origin values lower-case scheme and host; ports decimal; backend readiness must prove owned process identity through the token-protected test runtime identity route; frontend readiness must report `frontend_mode="preview"` and `frontend_command_kind="vite-preview"`; service-backed frontend auto-allocation uses stage-owned CORS windows, with stateful browser work in `39100-39199` and current non-stateful browser work in `39000-39099`; frontend port selection MUST use `CARTULARY_BROWSER_STAGE` or scheduler session metadata rather than target-name substring matching | config or port collision: `resource_conflict`; missing preview artifact or invalid config: `configuration_error` | origins, ports, runtime root, ownership proof, frontend mode |
| `PLAYWRIGHT_WORKERS`, worker count/index/offset envs                                            | browser                 | positive integers; worker offset `0..1024`                                                                            | Make `3`; shared config fallback `2`; direct isolated offset `0`; scheduled browser groups require scheduler-owned count and offset | Make variable, env, default, scheduler manifest | omitted only for direct isolated browser invocation      | decimal parse                                                                                                 | `configuration_error`, exit `2`                                                    | worker counts and scheduled worker slot range      |
| `VITEST_MAX_WORKERS`                                                                            | frontend unit           | positive integer `1..16`                                                                                              | direct public `frontend-unit` default `4`; scheduled `frontend-unit` default-check override `2` | Make variable, env, default, scheduler manifest | invalid                                                 | decimal parse                                                                                                 | `usage_error`, exit `2` for public target input                                     | worker count and scheduler-owned source            |
| Webserver-backed shard env                                                                      | browser                 | required grep/file values declared by target, plus selected-test artifact path for manifest-aware shard verification   | none                                                                                      | Make variable, manifest                         | invalid                                                 | exact string after JSON/shell decoding; selected tests validated as `cartulary.playwright_manifest_selection.v1` | missing required value: `configuration_error`, exit `2`                            | declared shard IDs as compatibility fallback; selected `(row_id,file,title)` entries when artifact-backed |
| `CARTULARY_ENABLE_TEST_ROUTES`                                                                  | reset/browser           | exact `1` enables test routes                                                                                         | disabled                                                                                  | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` means disabled                                                             | enabled boolean                                    |
| `CARTULARY_TEST_ROUTE_TOKEN`                                                                    | reset/browser           | non-empty opaque string with at least 128 bits entropy                                                                | generated by harness stack when reset route enabled                                       | harness generated, env for attach mode          | invalid                                                 | not normalized                                                                                                | missing/low-entropy token: `configuration_error`, exit `2` before stack use        | redaction token only                               |
| Object-store runtime env                                                                        | app runtime             | `CARTULARY_S3_OBJECT_PRIMARY_*` endpoint, credentials, secure bool, bucket                                            | browser/dev SeaweedFS S3 local values                                                     | Make variable, env, config binding, default     | invalid for required members                            | endpoint normalized; credentials redacted                                                                     | app startup/reset failure according to Section 12                                  | redacted credential fields                         |
| Runtime root envs                                                                               | app runtime             | `CARTULARY__ROOTS__*__PATH` filesystem paths                                                                          | browser stack creates under runtime root                                                  | Make variable, env, config binding, default     | invalid                                                 | path normalization                                                                                            | invalid/unwritable path: `configuration_error` or app startup failure              | normalized path                                    |
| `CARTULARY_HARNESS_REPO_ROOT`, `CARTULARY_HARNESS_SCRATCH_ROOT`, `TMPDIR`                       | harness scratch         | filesystem path                                                                                                       | `${TMPDIR:-/tmp}/cartulary-harness-scratch`                                               | env, default                                    | invalid for explicit scratch                            | path normalization; scratch root must be outside repo                                                         | in-repo scratch root: `configuration_error`, exit `2`                              | normalized scratch root                            |
| `CARTULARY_CLEANUP_DRY_RUN`                                                                     | cleanup                 | exact `1`                                                                                                             | false                                                                                     | Make variable, env, default                     | false                                                   | exact string compare                                                                                          | non-`1` false                                                                      | dry-run boolean                                    |
| `CARTULARY_DESTRUCTIVE_CONFIRM`                                                                 | destructive local reset | enum equal to the target name, currently `db-reset` or `object-store-reset`                                            | none                                                                                      | Make command line only                          | invalid when supplied empty; omitted allowed only for dry-run | trim exact token                                                                                              | wrong Make command-line token is `usage_error`; inherited-env-only confirmation is ignored and cannot satisfy reset confirmation; missing token on real reset fails before mutation | selected target token                            |
| `LINT_SHELL_STRICT`                                                                             | lint                    | exact `1`; public Make lint targets force strict blocking behavior                                                    | `1` for public Make targets; raw script fallback may default false                       | Make recipe                                      | false outside public Make                            | exact string compare                                                                                          | public Make target overrides ignored by recipe-owned strict value                   | boolean when true                                  |
| `STATICCHECK_CHECKS`                                                                             | static analysis          | closed; not a public Make target input                                                                                | Staticcheck default fixed by the public target row and wrapper                           | raw script only                                  | not applicable                                           | none                                                                                                          | Make command-line use is `usage_error`, exit `2`                                  | none                                               |
| Gosec rule, flag, and pattern variables (`GOSEC_*`, `GOSEC_TARGETED_*`, `GOSEC_AUDIT_*`)         | security                | closed; not public Make target inputs                                                                                 | curated profiles from this NLSpec and task-surface owner inputs                          | raw script only                                  | not applicable                                           | none                                                                                                          | Make command-line use is `usage_error`, exit `2`                                  | retained security profile metadata                 |
| `GOVULNCHECK_DB`                                                                                 | security                | optional Govulncheck vulnerability DB path or endpoint token                                                          | omitted; Govulncheck default DB                                                           | Make variable, env, default                     | omitted                                                 | path-token validation                                                                                         | invalid value is `usage_error`, exit `2`                                           | value                                              |
| `GOVULNCHECK_FLAGS`, `GOVULNCHECK_PATTERNS`                                                      | security                | closed; not public Make target inputs                                                                                 | `-test -json` flags and authored package roots fixed by the public target row and wrapper | raw script only                                  | not applicable                                           | none                                                                                                          | Make command-line use is `usage_error`, exit `2`                                  | none                                               |

## 6. Result Roots, Run IDs, and Artifact Identity

**TH-HARNESS-REQ-150**
A public Make target that emits retained artifacts MUST compute artifact identity as:

```text
run_root = normalize_result_root(CARTULARY_TEST_RESULTS_DIR) / normalize_run_id(CARTULARY_TEST_RUN_ID)
```
Verified by: TH-HARNESS-AC-003, TH-HARNESS-AC-015

Retained raw telemetry captures from `otel-conformance` MUST be retained artifacts owned by the `otel-conformance` target below the normalized run root. They MUST NOT be written below committed golden directories such as `internal/testutil/golden/otel/`.

### 6.1 Result Root Normalization

```text
normalize_result_root(input):
  if input is omitted:
    input = ".cartulary/test-results"
  reject empty string
  reject NUL
  reject path equal to "/" after lexical normalization
  reject path equal to "." after lexical normalization
  reject any caller-supplied segment equal to ".."
  reject backslash on POSIX conformance hosts
  if relative:
    resolve against repository root
  if absolute:
    allow for artifact writing and set cleanup_scope = "external_or_custom"
  create the directory if missing
  create retained run roots and target artifact directories with owner-only permissions on POSIX conformance hosts
  fail with configuration_error if parent is not writable
  fail with configuration_error if a caller-supplied custom result root is world-writable without sticky bit
```

### 6.2 Run-ID Grammar

```text
run_id = 1*96(ALPHA / DIGIT / "-" / "_" / ".")
run_id MUST NOT equal "." or ".."
run_id MUST NOT contain "/"
run_id MUST NOT contain "\\"
run_id MUST NOT contain whitespace
```

When `CARTULARY_TEST_RUN_ID` is omitted, the wrapper MUST generate:

```text
YYYYMMDDTHHMMSSZ-p<PID>
```

`YYYYMMDDTHHMMSSZ` is the UTC wrapper start time. `<PID>` is the decimal process ID of the Make-owned top-level wrapper.
Every public target that retains artifacts MUST freeze the resolved run ID once
at target scope before its first recipe line. Public preflight, prerequisite
coordination, child invocation, summary emission, cleanup, and observability
finalization MUST reuse that exact value; no recipe line may re-evaluate the
default run-ID expression or create a sibling root for the same invocation.

### 6.3 Collision Rules

| Case                                                | Required behavior                                                                        |
| --------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Omitted run ID                                      | Generate a default run ID.                                                               |
| Caller-supplied run ID path does not exist          | Create it.                                                                               |
| Caller-supplied run ID path exists and is empty     | Reuse it.                                                                                |
| Caller-supplied run ID path exists and is non-empty | Fail before child work with `configuration_error`, exit `2`.                             |
| Generated default run ID collides                   | Append `-n<N>` with the smallest positive decimal `N` that produces a non-existing path. |

### 6.4 Artifact-Proof Rule

**TH-HARNESS-REQ-151**
Retained artifacts prove only the explicit `{result_root, run_id, target}` triple. A newest-run fallback MAY be used for human investigation, but MUST NOT satisfy harness conformance evidence.
Verified by: TH-HARNESS-AC-015

**TH-HARNESS-REQ-152**
Every post-cutover row-bearing evidence artifact MUST identify `schema_id`,
`evidence_epoch`, `command_id`, `run_id`, `owner_id`, `selected_rows`,
`source_snapshot_digest`, `test_catalog_digest`,
`verification_routing_digest`, `runtime_profile_digest`,
`resource_profile_digest`, `fixture_profile_digest`, `started_at`,
`finished_at`, and `duration_ms`. `selected_rows` MUST be sorted and
duplicate-free. Missing identity fields make the artifact incompatible, not
stale-by-age.

An internal evidence partition MUST retain its own `target_id` while using the command identity of the public semantic command that owns the partition. The `backend-integration-support` partition therefore MUST use `target_id=backend-integration-support` and `command_id=cartulary.harness.command.backend_integration.v1`. This ownership route MUST be explicit catalog-side policy; runtime target-name inference and a new public or private support command identity are forbidden.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-153**
`source_snapshot_digest_v2` is SHA-256 over RFC 8785 canonical JSON
describing every Git-tracked and non-ignored untracked executable repository
input, ordered by repository-relative path. Each entry contains path, stable
file-kind or Git-mode identity, and byte digest. Ignored paths, result roots,
cache roots, the complete `docs/` subtree, and every Markdown file
(`.md` or `.markdown`, ASCII case-insensitive) are excluded.

Documentation exclusion MUST use only the normalized repository-relative path
returned by Git and MUST occur before any `lstat`, `stat`, open, read,
`readlink`, or hashing operation for that path. Editing, adding, deleting,
renaming, or rearranging an excluded specification or Markdown guidance file
MUST NOT change the snapshot digest or file count. A retained symlink outside
the excluded documentation surfaces is recorded as a link and its link bytes
are hashed; the digest algorithm MUST NOT follow it. Duplicate, escaping,
unreadable, or unstable executable-input paths MUST fail snapshot
construction.
Verified by: TH-HARNESS-AC-066, TH-HARNESS-AC-067

**TH-HARNESS-REQ-154**
Evidence from different retained roots is compatible only when `owner_id`, source
snapshot digest, test-catalog digest, and verification-routing digest are
identical. Selected-row inventories are target partitions and therefore MAY differ
between gates. Each artifact's selected rows and runtime, resource, and fixture
profile digests MUST exactly equal the current catalog-derived partition for that
target. The union of accepted `(owner_id, target_id, row_id)` records MUST equal the
full applicability set derived by TH-HARNESS-REQ-668. Two candidate artifacts for the
same required target and row are ambiguous and MUST fail. An auditor MUST NOT search
unlisted roots or select a newest candidate.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-155**
Evidence freshness has no wall-clock TTL. A matching semantic identity remains fresh regardless of age, subject to the target success, completeness, contamination, and release rules in this NLSpec. Release closure additionally requires a clean worktree. Local dirty-tree evidence is compatible only when its exact source snapshot digest matches the audited worktree.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-156**
Owner evidence timestamps MUST be UTC RFC 3339 strings using uppercase `T` and `Z` and exactly three fractional-second digits. `duration_ms` MUST be a non-negative integer measured from a monotonic clock. Wall-clock timestamps MUST NOT be subtracted to determine duration.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-157**
Result-root readers MUST apply Section 6 normalization and containment before opening retained evidence. They MUST reject NUL, traversal, symlink escape, non-directory roots, world-writable custom roots without sticky bit, and target artifacts outside the declared run root. Invalid caller path syntax is `usage_error`; a syntactically valid root containing unsafe or incompatible evidence is `artifact_error`.
Verified by: TH-HARNESS-AC-066, TH-HARNESS-AC-067

**TH-HARNESS-REQ-158**
Prepared artifact identity is atomic. A child MAY reuse prepared identity only when
`CARTULARY_HARNESS_IDENTITY_PREPARED=1` and the complete normalized
`{result_root, run_id, target}` tuple is present and validates before any artifact
write. A partial tuple, a marker value other than exact `1`, an undeclared target,
or a target mismatch MUST fail closed with `configuration_error`, exit `2`, before
artifact creation.

A scheduler child that requires an independent artifact identity MUST clear the
complete parent prepared-identity tuple, parent selectors, worker overrides, and
Make command-line override metadata as one operation. The scheduler MUST give that
child a distinct result root contained below the parent-owned diagnostic namespace,
and the child MUST perform ordinary identity preflight. Child diagnostic artifacts
MUST NOT overwrite parent plans, scheduler logs, work-unit results, accounting,
summaries, or cleanup evidence, and MUST NOT substitute for the parent's row-bearing
evidence.
Verified by: TH-HARNESS-AC-003, TH-HARNESS-AC-015, TH-HARNESS-AC-064, TH-HARNESS-AC-071

## 7. Output Modes and Machine Output

**TH-HARNESS-REQ-200**
Public Make targets MUST recognize exactly these output-mode tokens: `quiet`, `summary`, `ci`, `verbose`, `debug`, and `machine`. Unknown output modes MUST fail with `configuration_error` and exit `2` before child work. A recognized mode is accepted only when the target's output class allows it. When a recognized mode is not accepted for that output class, the target MUST fail before child work with the Section 7.2 output-class rejection behavior.
Verified by: TH-HARNESS-AC-002, TH-HARNESS-AC-004, TH-HARNESS-AC-005

### 7.1 Output-Mode Resolution

```text
resolve_output_mode(CARTULARY_OUTPUT_MODE, VERBOSE, CI_VERBOSE, CI, target):
  if CARTULARY_OUTPUT_MODE is present and non-empty:
    return exact token after lower-case validation
  if VERBOSE == "1":
    return "verbose"
  if CI_VERBOSE == "1":
    return "ci"
  if target == "ci" or CI == "1":
    return "ci"
  return "summary"
```

`quiet`, `debug`, and `machine` are selected only by `CARTULARY_OUTPUT_MODE`.

### 7.2 Output Class Matrix

| Output class                       | Public targets                                                                                    | `machine` accepted? | `machine` stdout                                        | `machine` stderr                                                                      | Success artifacts                                         | Failure behavior                                                       |
| ---------------------------------- | ------------------------------------------------------------------------------------------------- | ------------------: | ------------------------------------------------------- | ------------------------------------------------------------------------------------- | --------------------------------------------------------- | ---------------------------------------------------------------------- |
| `summary_with_artifacts`           | Leaf, toolchain, non-aggregate build/lint child, drift, browser-stage, and formatting targets with wrapper summaries |                 yes | One `cartulary.tool_run_summary.v5` JSON object plus LF | Empty after wrapper starts; pre-wrapper diagnostics allowed only before JSON emission | Tool-run summary required                                 | Same schema with failure fields and nonzero exit                       |
| `service_summary`                  | `db-up`, `db-reset`, `services-up`, `services-down`, `object-store-init`, `object-store-reset` |                 yes | One `cartulary.tool_run_summary.v5` JSON object plus LF | Empty after wrapper starts                                                            | Tool-run summary and service diagnostics for service-owning rows | Same schema with service failure fields                                |
| `aggregate_summary_with_artifacts` | `test-fast`, `test`, `lint`, `ci`, `release-check`, `build`                                       |                 yes | One `cartulary.tool_run_summary.v5` JSON object plus LF | Empty after wrapper starts                                                            | Aggregate summary plus child references                   | Same schema with primary failure                                       |
| `scheduler_summary_with_artifacts` | `check`, `test-slice`, `service-backed-test-slice`                                                |                 yes | One `cartulary.tool_run_summary.v5` JSON object plus LF | Empty after wrapper starts; no scheduler progress prose                               | Scheduler summary/events and run summary                  | Same schema with scheduler or child failure                            |
| `machine_stdout_json`              | `target-plan-json` and other explicitly declared JSON discovery targets                           |                 yes | One closed target JSON value plus LF                    | Empty on success                                                                      | None unless target declares artifacts                     | Invalid input exits `2`; error JSON only when Section 7.4 declares it |
| `human_summary`                    | `help`, `help-all`, text discovery/explanation targets                                            |                  no | Empty                                                   | Bounded diagnostic allowed on failure                                                 | None unless target row declares diagnostic artifacts      | `machine` rejected as `usage_error`, exit `2`                          |
| `interactive_raw`                  | `dev`                                                                                             |                  no | Empty when `machine` requested                          | Diagnostic allowed                                                                    | None                                                      | `machine` rejected as `usage_error`, exit `2`                          |
| `destructive_human`                | `clean`, `distclean`                                                                              |                  no | Empty when `machine` requested                          | Diagnostic allowed                                                                    | None                                                      | `machine` rejected as `usage_error`, exit `2`                          |

Successful `frontend-fallow-static` `summary` and `quiet` output MUST expose only bounded result and artifact references. Raw Fallow JSON, SARIF, Markdown, stdout, and stderr diagnostics MUST be retained as artifacts rather than emitted in successful summary stdout; `verbose` and `debug` modes MAY expose raw child diagnostics.

### 7.3 Human Output Budgets

| Mode      |             Success stdout budget | Success stderr                                              | Child logs                                       |
| --------- | --------------------------------: | ----------------------------------------------------------- | ------------------------------------------------ |
| `quiet`   |                    At most 1 line | Empty unless failure                                        | No child logs.                                   |
| `summary` |   At most 30 lines and 8192 bytes | Empty                                                       | Retained in artifacts only.                      |
| `ci`      | At most 120 lines and 32768 bytes | Empty unless CI wrapper failure occurs before summary layer | Retained in artifacts; bounded progress allowed. |
| `verbose` |              No fixed line budget | Tool-dependent                                              | May stream child logs.                           |
| `debug`   |              No fixed line budget | Tool-dependent                                              | May stream wrapper telemetry.                    |

### 7.4 Machine Output

**TH-HARNESS-REQ-201**
For every public target whose output class accepts `machine`, stdout MUST be exactly one UTF-8 JSON object followed by one LF, except that `machine_stdout_json` discovery targets MAY emit one closed target JSON value followed by one LF when Section 7.4 defines that value. The JSON payload MAY contain artifact pointers. Stdout MUST NOT be a pointer-only payload, multiple JSON values, scheduler progress prose, child logs, or human summary text.
Verified by: TH-HARNESS-AC-004

**TH-HARNESS-REQ-202**
For every public target whose output class rejects `machine`, setting `CARTULARY_OUTPUT_MODE=machine` MUST fail before child work with `failure_class=config`, `failure_reason=usage_error`, public exit code `2`, empty stdout, and bounded stderr diagnostic.
Verified by: TH-HARNESS-AC-005

For `target-plan-json`, the current closed JSON contract is a UTF-8 JSON array followed by one LF. Each array item MUST be a catalog projection containing `row_id`, `owner_id`, `family_id`, `collaborator_ids`, `verification_ids`, `runner`, `selector`, `evidence_class`, `runtime_profile_id`, `resource_profile_id`, `fixture_profile_id`, `default_check`, `claim_posture`, `status`, and the target and work-unit identities derived from that row. Arrays and rows MUST be ordered by ascending ASCII-byte ID order and duplicate-free. The projection MUST NOT contain a historical delivery selector, legacy registry field, documentation-derived activation field, raw executable path, raw shell, inline port, capacity, service topology, environment variable, or fixture path. Unknown `TARGET` input MUST fail with `usage_error`, exit `2`, empty stdout, and a bounded diagnostic; no partial JSON is emitted.

## 8. Artifact and Schema Contract

**TH-HARNESS-REQ-250**
A public Make-owned command that declares a stable schema ID MUST emit JSON that validates against the matching normative schema attachment before command success. If required artifact validation fails, the public target MUST fail with `artifact_error` or `scheduler_accounting_error` according to Section 9.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-004, TH-HARNESS-AC-025

Current owner-row producers MUST provide both logical and executed duration in their schema-owned records; an executed row sets both to the same measurement, while a derived accounting record sets executed duration to zero. Artifact construction MUST NOT infer either value from a historical generic duration field. Current failure records accept only `lifecycle_step`,
`scheduler_event_sequence`, and `child_registry_order` for deterministic ordering;
alternate spellings are unsupported artifact shape. Current target summaries MUST reference `cartulary.test_evidence_accounting.v1` as an artifact and MUST fail if a retired frontend-row-accounting extension duplicate is present.

The following schema IDs are public contracts. Schema file paths are repository attachments, not behavioral owners. `tools/harness_schema_attachments.json` is the authored machine owner for the current repository attachment path, classification, producer class, and validation target of every schema under `tools/schemas`. It MUST contain exactly the current schema files; an unregistered file, duplicate ID/path, or missing registered attachment is invalid. The table below owns public behavioral schema IDs and validation points; the machine owner supplies attachment mechanics and MUST remain parity-checked with this table.

| Schema ID                                       | Repository attachment path                                               | Status            | Producer class           | Required validation point                 |
| ----------------------------------------------- | ------------------------------------------------------------------------- | ----------------- | ------------------------ | ----------------------------------------- |
| `cartulary.harness_artifact_ref.v1`             | `tools/schemas/cartulary.harness_artifact_ref.v1.schema.json`             | present           | Shared schema component  | Whenever a current schema declares structured retained harness artifact references. |
| `cartulary.tool_run_summary.v5`                 | `tools/schemas/cartulary.tool_run_summary.v5.schema.json`                 | present           | Centralized wrappers     | Before wrapper exits.                     |
| `cartulary.fallow_reachability_owner.v1`        | `tools/schemas/cartulary.fallow_reachability_owner.v1.schema.json`        | present           | Fallow reachability owner | During JSON shape checks and before `frontend-fallow-static` builds its effective config. |
| `cartulary.fallow_static_summary.v2`            | `tools/schemas/cartulary.fallow_static_summary.v2.schema.json`            | present           | Fallow static target     | Before `frontend-fallow-static` success.  |
| `cartulary.test_owner_summary.v1`               | `tools/schemas/cartulary.test_owner_summary.v1.schema.json`               | present           | Owner accounting handlers | Before target summary consumes it.       |
| `cartulary.vitest_failure_details.v1`           | `tools/schemas/cartulary.vitest_failure_details.v1.schema.json`           | present           | Vitest wrappers          | Before owner summaries consume failure diagnostics. |
| `cartulary.test_target_summary.v4`              | `tools/schemas/cartulary.test_target_summary.v4.schema.json`              | present           | Target summary generator | Before aggregate/run summary consumes it. |
| `cartulary.test_run_summary.v6`                 | `tools/schemas/cartulary.test_run_summary.v6.schema.json`                 | present           | Run summary generator    | Before public aggregate success.          |
| `cartulary.same_run_helper_artifact_ref.v2`     | `tools/schemas/cartulary.same_run_helper_artifact_ref.v2.schema.json`     | present           | Run summary generator    | Before an aggregate reports same-run helper artifact reuse. |
| `cartulary.task_surface_owner.v1`               | `tools/schemas/cartulary.task_surface_owner.v1.schema.json`               | present           | Authored task-surface owner validation | Before task-surface, scheduler, or generated Make projection. |
| `cartulary.contract_family_registry.v3`         | `tools/schemas/cartulary.contract_family_registry.v3.schema.json`         | present           | Contract generator family registry | During JSON shape checks and before `tools/contractgen` emits generated contract roots. |
| `cartulary.check_scheduler_summary.v10`          | `tools/schemas/cartulary.check_scheduler_summary.v10.schema.json`          | present           | Check scheduler          | Before scheduler target success.          |
| `cartulary.service_backed_scheduler_summary.v10` | `tools/schemas/cartulary.service_backed_scheduler_summary.v10.schema.json` | present           | Service-backed scheduler | Before scheduler target success.          |
| `cartulary.test_slice_scheduler_summary.v1`     | `tools/schemas/cartulary.test_slice_scheduler_summary.v1.schema.json`     | present           | Owner-slice scheduler    | Before an owner-slice scheduler target succeeds. |
| `cartulary.scheduler_event.v7`                  | `tools/schemas/cartulary.scheduler_event.v7.schema.json`                  | present           | Scheduler                | During scheduler JSONL validation.        |
| `cartulary.scheduler_pressure_summary.v4`       | `tools/schemas/cartulary.scheduler_pressure_summary.v4.schema.json`       | present           | Scheduler reporter       | Before scheduler target success.          |
| `cartulary.fixture_tier_proof.v2`               | `tools/schemas/cartulary.fixture_tier_proof.v2.schema.json`               | present           | Scheduler reporter and fixture-proof validators | Before a retained fixture-tier proof artifact is accepted. |
| `cartulary.test_slice_plan.v2`                  | `tools/schemas/cartulary.test_slice_plan.v2.schema.json`                  | present           | Owner-slice planner      | Before setup, retained plan emission, or owner-slice JSON output is accepted. |
| `cartulary.test_evidence_root_manifest.v1`      | `tools/schemas/cartulary.test_evidence_root_manifest.v1.schema.json`      | present           | Evidence-audit caller    | Before any retained root is opened. |
| `cartulary.requirement_registry.v1`             | `tools/schemas/cartulary.requirement_registry.v1.schema.json`             | present           | Requirement registry     | Before verification loading or coverage validation. |
| `cartulary.requirement_catalog.v1`              | `tools/schemas/cartulary.requirement_catalog.v1.schema.json`              | present           | Requirement owner        | Before verification loading or coverage validation. |
| `cartulary.verification_registry.v2`            | `tools/schemas/cartulary.verification_registry.v2.schema.json`            | present           | Verification registry    | Before catalog compilation or evidence routing. |
| `cartulary.verification_contract.v2`            | `tools/schemas/cartulary.verification_contract.v2.schema.json`            | present           | Verification owner       | Before catalog compilation or evidence routing. |
| `cartulary.test_owner_registry.v1`              | `tools/schemas/cartulary.test_owner_registry.v1.schema.json`              | present           | Test catalog owner       | Before owner manifest loading. |
| `cartulary.test_family_manifest.v2`             | `tools/schemas/cartulary.test_family_manifest.v2.schema.json`             | present           | Test family owners       | Before selection, topology generation, or audit. |
| `cartulary.test_runner_registry.v1`             | `tools/schemas/cartulary.test_runner_registry.v1.schema.json`             | present           | Runner registry          | Before selector resolution or adapter invocation. |
| `cartulary.test_evidence_accounting.v1`         | `tools/schemas/cartulary.test_evidence_accounting.v1.schema.json`         | present           | Owner-aware target summaries | Before target-summary success. |
| `cartulary.test_evidence_audit_summary.v1`      | `tools/schemas/cartulary.test_evidence_audit_summary.v1.schema.json`      | present           | Owner evidence audit     | Before audit success. |
| `cartulary.test_owner_explanation.v1`           | `tools/schemas/cartulary.test_owner_explanation.v1.schema.json`           | present           | Owner diagnostics        | Before target-local JSON output. |
| `cartulary.task_guide_summary.v2`               | `tools/schemas/cartulary.task_guide_summary.v2.schema.json`               | present           | Task guidance            | Before target-local JSON output. |
| `cartulary.test_catalog_check_summary.v1`       | `tools/schemas/cartulary.test_catalog_check_summary.v1.schema.json`       | present           | Catalog validator        | Before private catalog-check success. |
| `cartulary.executable_input_policy.v1`          | `tools/schemas/cartulary.executable_input_policy.v1.schema.json`          | present           | Executable input boundary owner | Before executable-input policy validation. |
| `cartulary.govulncheck_findings.v1`             | `tools/schemas/cartulary.govulncheck_findings.v1.schema.json`             | present           | Govulncheck wrapper      | Before failure classification or target-summary security rollup consumes findings. |
| `cartulary.test_services.lease.v1`              | `tools/schemas/cartulary.test_services.lease.v1.schema.json`              | present           | Service suite            | Before attach or cleanup relies on lease. |
| `cartulary.test_services.lifecycle.v2`          | `tools/schemas/cartulary.test_services.lifecycle.v2.schema.json`          | present           | Service suite            | During service lifecycle JSONL validation. |
| `cartulary.test_services.scope.v1`              | `tools/schemas/cartulary.test_services.scope.v1.schema.json`              | present           | Service suite            | Before scheduler failure propagation consumes service-suite diagnostics. |
| `cartulary.web_e2e_stack.v4`                    | `tools/schemas/cartulary.web_e2e_stack.v4.schema.json`                    | present           | Browser session lifecycle | Before browser target starts Playwright. |
| `cartulary.browser_startup_event.v1`             | `tools/schemas/cartulary.browser_startup_event.v1.schema.json`             | present           | Browser session lifecycle | For each append-only startup transition. |
| `cartulary.browser_startup_diagnostics.v2`       | `tools/schemas/cartulary.browser_startup_diagnostics.v2.schema.json`       | present           | Browser session lifecycle | Once at terminal ready or failed state. |
| `cartulary.browser_group_result.v2`              | `tools/schemas/cartulary.browser_group_result.v2.schema.json`              | present           | Browser evidence adapter | Before browser group evidence is accepted. |
| `cartulary.browser_target_result.v1`             | `tools/schemas/cartulary.browser_target_result.v1.schema.json`             | present           | Browser evidence finalizer | Before browser target evidence is accepted. |
| `cartulary.local_object_store_proxy_start_attempt.v1` | `tools/schemas/cartulary.local_object_store_proxy_start_attempt.v1.schema.json` | present | Local development proxy lifecycle | Before a startup attempt is recovered or promoted. |
| `cartulary.local_object_store_proxy_lease.v1`    | `tools/schemas/cartulary.local_object_store_proxy_lease.v1.schema.json`    | present           | Local development proxy lifecycle | Before reuse or signaling. |
| `cartulary.local_object_store_proxy_health.v1`   | `tools/schemas/cartulary.local_object_store_proxy_health.v1.schema.json`   | present           | Local development proxy lifecycle | During ownership and configuration proof. |
| `cartulary.test.runtime_identity.v1`             | `tools/schemas/cartulary.test.runtime_identity.v1.schema.json`             | present           | Browser stack            | During backend identity readiness probing. |
| `cartulary.test.runtime_reset.v1`               | `tools/schemas/cartulary.test.runtime_reset.v1.schema.json`               | present           | Reset route/wrapper      | Before browser reset success is accepted. |
| `cartulary.test.clock_control.v1`               | `tools/schemas/cartulary.test.clock_control.v1.schema.json`               | present           | Test clock route         | Before a fixed, offset, reset, or state clock-control response is accepted. |
| `cartulary.test.public_error_fault.v1`          | `tools/schemas/cartulary.test.public_error_fault.v1.schema.json`          | present           | Browser stack            | Before an armed public-error fault is accepted. |
| `cartulary.test.network_flow_fault_control.v1`  | `tools/schemas/cartulary.test.network_flow_fault_control.v1.schema.json`  | present           | Network Flow fault-control route | Before an armed Network Flow commit or worker fault is accepted. |
| `cartulary.test.network_flow_randomness_control.v1` | `tools/schemas/cartulary.test.network_flow_randomness_control.v1.schema.json` | present       | Network Flow randomness-control route | Before an armed deterministic Network Flow random stream is accepted. |
| `cartulary.test.network_flow_auth_transition_control.v1` | `tools/schemas/cartulary.test.network_flow_auth_transition_control.v1.schema.json` | present | Network Flow auth-transition control route | Before an armed Network Flow auth-transition control is accepted. |
| `cartulary.test.network_flow_audit_assertion_control.v1` | `tools/schemas/cartulary.test.network_flow_audit_assertion_control.v1.schema.json` | present | Network Flow audit-assertion control route | Before an armed Network Flow audit-count or replay assertion is accepted. |
| `cartulary.fixture_report.v1`                   | `tools/schemas/cartulary.fixture_report.v1.schema.json`                   | present           | Fixture report target    | Before machine JSON is emitted.           |
| `cartulary.network_flow_fixture_manifest.v2`    | `tools/schemas/cartulary.network_flow_fixture_manifest.v2.schema.json`    | present           | Network Flow fixture manifest validator | Before a Network Flow fixture manifest is selected for behavior execution. |
| `cartulary.network_flow_fixture_scenario.v2`    | `tools/schemas/cartulary.network_flow_fixture_scenario.v2.schema.json`    | present           | Network Flow fixture scenario validator | Before a Network Flow fixture scenario is selected for behavior execution. |
| `cartulary.graph_projection_fixture_manifest.v3` | `tools/schemas/cartulary.graph_projection_fixture_manifest.v3.schema.json` | present | Graph Projection fixture manifest validator | Before a Graph Projection behavior fixture is selected for execution. |
| `cartulary.network_flow_timezone_ruleset_provenance.v2` | `tools/schemas/cartulary.network_flow_timezone_ruleset_provenance.v2.schema.json` | present | Network Flow timezone provenance validator | During JSON shape checks and before timestamp fixtures are accepted. |
| `cartulary.agent_finalize_summary.v3`           | `tools/schemas/cartulary.agent_finalize_summary.v3.schema.json`           | present           | Agent finalizer          | Before `agent-finalize` exits.            |
| `cartulary.cache.readiness.v1`                  | `tools/schemas/cartulary.cache.readiness.v1.schema.json`                  | present           | Readiness cache helper   | Before a readiness cache record or retained cache artifact is accepted. |
| `cartulary.cache.build_artifact.v1`             | `tools/schemas/cartulary.cache.build_artifact.v1.schema.json`             | present           | Build-artifact cache helper | Before a build cache record or retained cache artifact is accepted. |
| `cartulary.cache.static_analysis.v1`            | `tools/schemas/cartulary.cache.static_analysis.v1.schema.json`            | present           | Static-analysis cache helper | Before a static-analysis cache record or retained cache artifact is accepted. |
| `cartulary.agent_finalize_action_cache_record.v1` | `tools/schemas/cartulary.agent_finalize_action_cache_record.v1.schema.json` | present         | Agent finalizer action cache | Before an `agent-finalize` action-cache hit is accepted. |
| `cartulary.execution_topology_render_cache.v1`  | `tools/schemas/cartulary.execution_topology_render_cache.v1.schema.json`  | present           | Execution-topology renderer | Before cached render content is reused.   |
| `cartulary.frontend_accessibility_summary.v4`   | `tools/schemas/cartulary.frontend_accessibility_summary.v4.schema.json`   | present           | Browser accessibility target | Before `browser-e2e-a11y` success.    |
| `cartulary.release_readiness_evidence.v2`       | `tools/schemas/cartulary.release_readiness_evidence.v2.schema.json`       | present           | Release-readiness aggregation | Before `release-readiness-evidence` success. |
| `cartulary.frontend_visual_fixture_registry.v5` | `tools/schemas/cartulary.frontend_visual_fixture_registry.v5.schema.json` | present           | Semantic frontend visual fixture registry and one-to-one design-contract projection validation | During JSON shape checks, catalog checks, and visual readiness validation. |
| `cartulary.frontend_claim_publication_review.v1` | `tools/schemas/cartulary.frontend_claim_publication_review.v1.schema.json` | present          | Conditional frontend claim-publication review metadata; no default target emits it | Before any future or explicit frontend claim-review artifact is accepted as Core 05-routed release evidence. |
| `cartulary.otel_conformance_summary.v1`         | `tools/schemas/cartulary.otel_conformance_summary.v1.schema.json`         | present           | OpenTelemetry conformance target | Before `otel-conformance` success. |

Adoption and continued conformance for `cartulary.testing_harness.current.v2` require live repository verification of every row whose `Status` is `present`. Each declared attachment path MUST exist, parse as a JSON schema, reject unknown top-level fields unless the schema declares an explicit extension container, and validate at least one positive fixture and one negative fixture. Missing or malformed required attachments make the v2 profile nonconforming; they MUST NOT be reclassified as future attachments. A schema absent from this table is unsupported by current harness validation; old retained JSON remains manually inspectable but MUST NOT be interpreted as current evidence.

Current owner-slice producers and retained-plan readers MUST accept only `cartulary.test_slice_plan.v2`; producers MUST NOT dual-emit an older form or provide a compatibility alias. Validation MUST cover owner ID, selection mode, dependency scope, completion scope, sorted unique requested and resolved row IDs, target and command consistency, work-unit identity uniqueness, dependency completion-key closure, registered resource names and claims, exact runtime-binary IDs, expected artifacts, and finalizers. The schema MUST close top-level, selection, work-unit, dependency, resource-claim, runtime-binary, expected-artifact, and finalizer objects with `additionalProperties=false`. Every input-valid invocation, including one whose child later fails, MUST retain a valid plan before setup or child execution. A rejected input or zero-row selection MUST NOT create a plan or start setup.

Current conformance MUST be proven from current schema-owned artifacts. Historical artifact inspection is non-conformance troubleshooting and is not an acceptance criterion for Sections 1-17.

**TH-HARNESS-REQ-251**
Schema-owned artifacts MUST be closed by default. Unknown top-level fields are invalid unless the schema declares an explicit extension container.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-025

**TH-HARNESS-REQ-252**
Every retained summary artifact MUST include normalized `result_root`, `run_id`, `run_root`, `target`, `output_mode`, public `exit_code`, primary `failure_class`, primary `failure_reason`, `started_at`, and `completed_at`. Timestamps MUST be RFC3339 UTC strings with non-null values.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-015

Public and machine summaries MUST serialize `run_root` once per summary and MUST
express every structured retained harness artifact reference relative to that
`run_root`. A structured retained harness artifact reference MUST validate against
`cartulary.harness_artifact_ref.v1` and MUST contain `role`, `path_kind`, and `path`.
`path_kind` is exactly `file` or `directory`. A file reference MUST also contain one
`format` from `json`, `jsonl`, `log`, `markdown`, `sarif`, or `text`; a directory
reference MUST NOT contain `format`. Paths use normalized POSIX separators, MUST be
non-empty and relative to `run_root`, and MUST reject absolute paths, empty segments,
`.` segments, and `..` traversal.

Repository source paths are not retained harness artifact references. A schema that
needs both retained evidence and repository sources MUST use separate `artifact_refs`
and `source_refs` or `source_files` fields. Repository source paths MUST be normalized
repo-relative paths and MUST NOT be dereferenced by retained-run diagnostics as though
they were run artifacts.

Before dereferencing a retained harness artifact reference, a consumer MUST prove
run-root containment, reject symlink path components, require the declared filesystem
kind, and apply its owner-defined bounded read policy. A mismatch is
`failure_class=artifact`, `failure_reason=artifact_error`, public exit `11`. Consumers
MUST validate all selected references before emitting referenced content.

The current schema cut is atomic. `cartulary.tool_run_summary.v5`,
`cartulary.same_run_helper_artifact_ref.v2`,
`cartulary.test_services.lifecycle.v2`,
`cartulary.test_evidence_accounting.v1`,
`cartulary.frontend_accessibility_summary.v4`, and
`cartulary.release_readiness_evidence.v2` are the only current versions of those
families. Current producers and consumers MUST NOT dual-emit, dual-read, alias, or
translate the replaced versions.

**TH-HARNESS-REQ-253**
Every schema-owned artifact MUST include `schema_id`. A schema-owned artifact MAY include `extensions` only when its schema declares that field. When present, `extensions` MUST be an object keyed by reverse-DNS or `cartulary.*` extension keys. Consumers MUST ignore unknown extension keys and MUST NOT derive required behavior from an unknown extension key. Adding a new required top-level member or changing the meaning of an existing member requires a new schema ID. Extension data is supplemental only; any value required for conformance, drift, timing, cleanup, scheduling, or failure classification MUST be a declared schema member.

Supplemental service-backed extension data under `extensions["cartulary.service_backed"]` MUST normalize extension-level `readiness_status`, `teardown_status`, and `leak_status` to `pass`, `fail`, or `unknown`. These extension rollups MUST be derived from canonical service lifecycle artifacts and MUST NOT expose raw lifecycle tokens such as `succeeded`, `cleanup_failed`, or `skipped_no_lease` as pass/fail status fields. The schema-owned scheduler and service artifacts remain authoritative when extension data is absent or `unknown`.

Supplemental security extension data under `extensions["cartulary.security"]` MAY summarize scanner artifacts retained by the current target. When that extension contains a Govulncheck rollup, it MUST contain `govulncheck.status`, `govulncheck.finding_count`, `govulncheck.blocking_count`, and `govulncheck.blocking_vulnerability_ids` derived only from the current target's schema-valid `cartulary.govulncheck_findings.v1` artifacts. Omission is caller-interchangeable because the findings artifact and normalized failure fields remain authoritative.
Verified by: TH-HARNESS-AC-000

**TH-HARNESS-REQ-254**
Generated-drift replay MUST be driven by a declared scratch-input manifest. The drift checker MUST copy every declared generator runtime input into its scratch tree before invoking generation, including shared harness helper scripts used by generator wrappers. A missing declared scratch input MUST fail as an artifact error with a diagnostic naming the missing path, not as a tool-specific module, import, or Make lookup failure.
Verified by: TH-HARNESS-AC-000

**TH-HARNESS-REQ-255**
`browser-e2e-a11y` MUST emit exactly one retained `cartulary.frontend_accessibility_summary.v4` artifact for a completed accessibility target attempt. The artifact MUST contain active catalog accessibility rows only and MUST be a JSON object with `schema_id`, `rows[]`, `scenarios[]`, `keyboard_matrix[]`, `state_communication_checks[]`, `contrast_checks[]`, `violations[]`, and `artifact_refs[]`. Scenario status fields MUST use only `pass`, `fail`, `missing`, or `skipped`; check result fields MUST use only `pass` or `fail`; nested objects MUST be schema-closed. Inactive or unauthorized rows MUST NOT appear in this row-evidence artifact.

`browser-e2e-visual` MUST select active visual catalog rows whose Playwright selector stage is `visual`. Direct execution runs the full target inventory; an owner slice constrains selection to the resolved owner rows. Exact Playwright title patterns MUST come from catalog selectors. Matching screenshots remain implementation-readiness evidence and MUST NOT be inferred from snapshot filenames, deleted ledgers, or visual fixture registry text alone.

`browser-e2e-visual-update` MUST use the same visual row selection, runtime-profile session grouping, and service lifecycle as direct `browser-e2e-visual`, with Playwright snapshot update mode enabled for every selected group. The current refresh profile is the pinned Linux x86_64 environment and browser/toolchain pins owned by this NLSpec; other host profiles are unsupported for committed refreshes. The target MUST remain helper-only, MUST NOT be selected by `check`, `test`, `ci`, release gates, or either owner-slice command, and MUST NOT emit passing `browser-e2e-visual` target or owner-accounting evidence. Its authored writes are limited to committed Playwright visual goldens under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`. A refresh record MUST name its accepted trigger, affected row and fixture IDs, changed golden paths, capture-contract changes or their explicit absence, reviewer outcome, and the later ordinary visual validation root. A refresh is complete only after changed images are reviewed and a later `browser-e2e-visual` run passes with screenshot comparisons active. Refresh artifacts remain implementation-readiness evidence and MUST NOT satisfy product conformance, design conformance, release, or Core 05 publication gates.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-022

Frontend visual fixture identity MUST be semantic, immutable, and local to the active owner-catalog row and exact Playwright scenario that uses it. When a retained support contract names a fixture, its ID MUST match `visual.fixture.<semantic_name>` and MUST NOT encode a frontend namespace, delivery phase, ordinal, or legacy row ID. Fixture-support metadata MUST declare `capture_scope` as `full_viewport`, `selector`, or `region`; selector captures MUST use a stable data-attribute selector. Selector-only non-grid specimens MAY declare `scroll_normalization.kind="not_applicable"` with a reason instead of a workbook-grid anchor. A fixture with `no_dynamic_regions=true` MUST keep `dynamic_masks=[]`. Missing, retired, placeholder, or ambiguous fixtures MUST NOT remain in the active fixture population and MUST NOT close an owner row from generic text, inferred snapshot names, or ad hoc scenario titles. Visual selection and accounting MUST come only from catalog rows, never from a separate fixture registry.

For semantic-identity validation, fixture identity means values that route, select,
isolate, retain, or name test scenarios and artifacts, including fixture and seed
IDs, transaction and incident-key prefixes, actor/session ownership metadata,
scenario IDs, snapshot and golden paths, and retained artifact names. It does not
mean arbitrary product payload content merely because a test created the value.
Identity-bearing fixture values MUST be behavior-centric and MUST NOT embed a
retired delivery identity or a current owner, family, or row ID. Catalog identities
belong in selectors and retained evidence; they MUST NOT be copied into rendered
workbook content as replacement fixture labels. Product-owned uses of the word
`phase` and constructed negative fixtures that prove obsolete identity rejection
remain valid. Semantic validation MUST inspect declared identity-bearing
constructors and metadata rather than apply a repository-wide word ban.

The `visual.fixture.default_timeline_workbook_shell` fixture is the Default Timeline workbook shell. Its closure evidence MUST come from the exact active `module.workbook` visual catalog row and stable Playwright scenario declared for that fixture; the retained screenshot MUST be Playwright output from the running app-owned workbook shell with browser and operating-system chrome excluded. External concept images, generated mockups, browser-chrome screenshots, or design-reference bitmaps MAY inform review, but they MUST NOT be committed as the frontend golden, cited as row-accounting evidence, or used as an alternate visual fixture source.

Frontend visual support specimens MUST remain separate from default-shell evidence. Grid-adapter and token/theme screenshots classified as implementation or design support MUST NOT close `visual.fixture.default_timeline_workbook_shell`, substitute for the Default Timeline workbook-shell fixture, or become product-conformance evidence. One catalog row MUST NOT close another row unless its verification contract explicitly names an adopted consolidation relationship and the accounting artifact records that relationship.

Browser row accounting MUST close a catalog row only from the exact stable scenario IDs and diagnostic titles in its registered Playwright selector. A shell support row may use a target-level registered command result because its selector is the command ID. No other runner-wide or target-wide pass may close a row whose exact selected cases are absent. Failing or incomplete mapped targets MUST leave the row failed or infrastructure-failed rather than closed.

Public Make-owned verification targets MUST treat `unmapped` as unexpected executed test inventory. If a target exits successfully while its own accounting section has `counts.unmapped > 0`, target-summary generation MUST fail the target with `failure_class="harness"` and `failure_reason="test_accounting_unmapped"`. The retained `unmapped` and `unmapped_failed` count fields remain diagnostic fields, but successful public targets require `unmapped=0`. Canonical evidence MUST map through one active catalog row by the exact registered selector. Intentional residual support coverage MUST be declared in the verification registry and catalog with an owner and support profile; a separate unowned or filename-pattern classification is invalid.

Claim-publication routing is inactive unless a verification contract uses `behavior_class=claim_publication`, the profile and row posture resolve to the same active Core 05 claim, and the required publication predicate and artifact bundle validate. `claim_posture=informative` MAY retain engineering measurement but MUST NOT satisfy claim-bearing publication. The existing `benchmark-claim-check` target remains the Core 05 benchmark-manifest validator; ordinary implementation or informative evidence MUST NOT activate it. When its default manifest is absent, its no-claim pass MUST NOT be cited as Core 05 publication evidence.

**TH-HARNESS-REQ-256**
`explain-run` MUST diagnose retained aggregate run roots that contain `run-summary.json` and retained public tool-run roots that contain at least one `<target>/tool-run-summary.json`. Tool-run diagnostics MUST NOT require a synthetic aggregate `run-summary.json`. When a tool-run target also emits a command-specific summary artifact, such as `agent-finalize/finalize-summary.json`, `explain-run` MUST surface a bounded human summary of that artifact and retain `DETAIL=logs` access to target and child logs when `TARGET=<target>` is supplied. `DETAIL=accounting` MUST group retained test inventory by target, owner, family, evidence class, row ID, and file so module authors can triage unmapped or residual coverage without scraping raw summary JSON.

`explain-run DETAIL=logs` MUST accept only current
`cartulary.tool_run_summary.v5` structured log references. File references are
readable only for `format=log|text`. Directory references enumerate direct regular
`*.log` files in deterministic lexical order. Before emitting any referenced content,
the diagnostic MUST validate every selected reference, reject symlinks and traversal,
and enforce at most 4,096 files, 16 MiB per file, and 256 MiB in aggregate. An unsafe,
missing, or kind-mismatched reference fails with `artifact/artifact_error`, exit `11`,
without partial log replay. Non-current tool-summary schemas are unsupported rather
than translated or scanned for historical schema-specific diagnostics.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-019

**TH-HARNESS-REQ-257**
Cache records and retained cache-state artifacts are schema-owned harness artifacts only for the cache profile that produced them. A readiness or build cache record MUST identify the cache schema, scope, profile ID, state, reason code, key digest, input digest, output digest, record path, cacheable outputs, non-cacheable side effects, and timestamp. An `agent-finalize` action-cache record MUST identify the action ID, command ID, action contract version, input profile, key digest, cache schema, input/output digests, and output paths. An execution-topology render cache record MUST identify the renderer, generator version, Node version, input digest, and rendered content digests.

Cache records MAY be retained in `.cache/cartulary/*` and MAY be referenced by compact retained run-root cache artifacts such as `<target>/*-cache-*.json`. A run-root cache artifact proves only cache behavior for the current target attempt; it MUST NOT substitute for the target's required summary, child log, generated-drift verdict, security-scan output, service lifecycle evidence, or scheduler summary. Investigation tools MAY display cache state, but conformance evidence MUST cite the public target summary and required target artifacts rather than the local cache record.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-015, TH-HARNESS-AC-028

Operator runtime-binary consumers MUST retain a bounded provenance artifact for scheduler-selected operator scenario work. That artifact MUST name `CARTULARY_OPERATOR_BIN` as scheduler-produced, identify the `build-operator` producer target, record the normalized binary path, record the executable file digest, and reference the retained `cartulary.cache.build_artifact.v1` artifact and output digest from the producing target. A valid build-artifact cache hit MAY satisfy the producer target under this section, but the binary consumer MUST NOT be marked as scheduler `reused` merely because the binary was already built.
Verified by: TH-HARNESS-AC-037

**TH-HARNESS-REQ-268**
The owner-first schema families are `cartulary.requirement_registry.v1`,
`cartulary.requirement_catalog.v1`, `cartulary.verification_registry.v2`,
`cartulary.verification_contract.v2`, `cartulary.test_owner_registry.v1`,
`cartulary.test_family_manifest.v2`, `cartulary.test_runner_registry.v1`,
`cartulary.test_slice_plan.v2`,
`cartulary.test_slice_scheduler_summary.v1`,
`cartulary.test_evidence_root_manifest.v1`,
`cartulary.test_evidence_accounting.v1`,
`cartulary.test_evidence_audit_summary.v1`,
`cartulary.test_owner_summary.v1`,
`cartulary.test_owner_explanation.v1`, `cartulary.task_guide_summary.v2`, and
`cartulary.test_catalog_check_summary.v1`. Each is a required current
attachment and rejects old-family schema IDs.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-071

**TH-HARNESS-REQ-269**
An owner-slice target MUST retain `<target>/tool-run-summary.json`, `<target>/test-slice-plan.json`, `<target>/test-slice-scheduler-summary.json`, `<target>/owners/<owner-id>/test-evidence-accounting.json`, and `<target>/owners/<owner-id>/test-owner-summary.json`. `test-evidence-audit` MUST retain `<target>/tool-run-summary.json` and `<target>/test-evidence-audit-summary.json`. `explain-test-owner` and `task-guide` are read-only and MUST NOT create retained artifacts.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-270**
The slice plan MUST contain `selection_mode`, `completion_scope`, `dependency_scope`, `owner_id`, `requested_row_ids`, `resolved_row_ids`, all semantic identity digests from TH-HARNESS-REQ-152, normalized work units, dependencies, resource claims, finalizers, and expected artifact schemas. `selection_mode` is `default_owner` or `exact_rows`; `completion_scope` is `full_owner` or `selected_subset`; `dependency_scope` is `all` or `service_backed`. Arrays MUST be sorted and duplicate-free.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-064

**TH-HARNESS-REQ-271**
`cartulary.test_evidence_accounting.v1` MUST contain one expected row record and exactly one terminal observed row record for every resolved row. Each record identifies owner, family, verification IDs, runner, selector digest, execution profiles, attempt summaries, terminal state, duration, failure fields when applicable, and redacted artifact references. A missing, duplicate, unexpected, or incompatible row record is `scheduler_accounting_error`.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-272**
`cartulary.test_owner_summary.v1` MUST aggregate only the resolved row inventory. It MUST expose selected, passed, failed, infrastructure-failed, dependency-skipped, cancelled, and authorized-skipped counts; unused inputs; primary failure; artifact references; and whether the invocation closed `full_owner` or `selected_subset`. A selected subset MUST NOT be represented as full-owner closure.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-276**
Every successful public target that selects active catalog rows MUST finalize paired
`<target>/owners/<owner-id>/test-evidence-accounting.json` and
`<target>/owners/<owner-id>/test-owner-summary.json` shards before target-summary
success. Go and Vitest adapters MUST prove the exact registered symbol or title
inventory, Playwright adapters MUST prove the exact selected row results, and a
shell row MUST prove the selected target command identity. Aggregate process or
target success alone MUST NOT close an absent, duplicate, failed, or unmapped
selected case. A target-wide diagnostic test outside the selected catalog
partition MUST NOT be promoted into owner evidence or broaden the retained row
scope; exact-row execution MUST reject such an unexpected case.

Target-level selection scope is an atomic tuple: `all`, `default_check`, or exact
sorted row IDs. Only `rows` scope carries row IDs. A partial or contradictory tuple
MUST fail before owner evidence is written. Repeated target-summary aggregation MAY
reuse already-written shards only when their source, catalog, verification, target,
run, owner, and exact selected-row identities match; it MUST NOT broaden or narrow
the retained scope. The target tool-run summary MUST reference every paired owner
shard. Browser targets additionally retain the target browser-owner index.
Verified by: TH-HARNESS-AC-065, TH-HARNESS-AC-066, TH-HARNESS-AC-071

**TH-HARNESS-REQ-273**
`cartulary.test_evidence_audit_summary.v1` MUST identify every required target/root,
every supplied unused target/root, the exact common compatibility tuple, required
rows by target partition, accepted artifacts, rejected artifacts with reasons, and
the final owner closure result. The audit MUST fail when any required
`(owner_id, target_id, row_id)` obligation lacks exactly one compatible successful
terminal record.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-274**
`cartulary.test_owner_explanation.v1` MUST expose owner, manifest, families, row counts, runner and evidence distributions, profile distributions, default-check participation, and exact commands. `cartulary.task_guide_summary.v2` MUST expose `role="module-author"`, owner, focused owner slice, applicable generated/drift gates, evidence-derived broader gates, and release gate when required. Neither schema may contain a delivery-phase selector.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-071

**TH-HARNESS-REQ-275**
`cartulary.executable_input_policy.v1` is a required current schema attachment.
It uses Draft 2020-12, a constant `schema_id`, closed keys, required fields, and
`additionalProperties=false`; its policy validator rejects unknown fields,
unsorted inputs, or unsafe restricted-root declarations before scanning
executable behavior.
Verified by: TH-HARNESS-AC-067, TH-HARNESS-AC-071

**TH-HARNESS-REQ-277**
The required current harness-observability schema attachments are
`cartulary.harness_execution_context.v2`,
`cartulary.harness_invocation_start.v1`,
`cartulary.harness_observability_index.v1`,
`cartulary.harness_trace_bundle.v1`,
`cartulary.harness_hotspot_summary.v1`,
`cartulary.harness_sequence_event.v1`,
`cartulary.harness_public_target_duration_baselines.v2`,
and `cartulary.harness_performance_evidence_roots.v2`. The v1 execution-context
and performance schemas remain attached only so the closed
`retained_v1_reference_migration` reader can validate immutable historical
reference evidence; normal baseline and candidate operation MUST reject v1
manifests. Each schema MUST be Draft
2020-12, require its exact schema ID, close unknown fields, and validate before
the corresponding artifact or check succeeds.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-074, TH-HARNESS-AC-079

**TH-HARNESS-REQ-278**
The observability index path is
`<run-root>/_shared/harness-observability/observability-index.json`. It MUST
contain the run ID, generated status, sorted required, excluded, and
out-of-scope target
dispositions, a run-relative execution-context reference and digest, sorted
invocation records, source artifact digests, and any bounded partial-generation
diagnostic. Each invocation record identifies one
directory named by a stable invocation ID and references its trace bundle,
OTLP trace payload, OTLP metric payload, and hotspot summary by normalized
run-relative path plus lowercase SHA-256.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-074

**TH-HARNESS-REQ-279**
The trace bundle is a deterministic projection of native evidence. It MUST
contain one nonzero 32-hex trace ID, one nonzero 16-hex span ID per span, exactly
one root span, valid parent references or closed links, RFC3339 start and end
timestamps, nonnegative durations, a closed span class and status, sorted
source artifact references, and closed safe attributes. Stable input bytes MUST
produce byte-identical normalized trace, metric, and hotspot output. Native
evidence remains authoritative when a projection disagrees.
Verified by: TH-HARNESS-AC-074, TH-HARNESS-AC-075

**TH-HARNESS-REQ-280**
The OTLP JSON artifacts MUST be valid JSON encodings of
`ExportTraceServiceRequest` and `ExportMetricsServiceRequest` under the adopted
protocol baseline. Their resource uses `service.name=cartulary.harness`; their
only instrumentation scope is `cartulary.harness.execution` with the adopted
harness contract version, null schema URL, and no scope attributes. The trace
payload MUST contain the same trace/span IDs, parentage, timestamps, statuses,
and allowed attributes as the native trace bundle. Metric names and dimensions
are closed by Section 10.5.
Verified by: TH-HARNESS-AC-075, TH-HARNESS-AC-076

**TH-HARNESS-REQ-281**
`sequence-events.jsonl` is retained below the aggregate target directory. Each
record validates as `cartulary.harness_sequence_event.v1`; sequence-local `seq`
starts at one and has no gaps. Events are exactly `sequence_started`,
`step_eligible`, `step_started`, `step_finished`, `step_skipped`,
`sequence_finished`, or `sequence_interrupted`. Each terminal step event
records normalized status, exit code, wall and monotonic boundaries, step ID,
target, dependencies, and execution mode without raw command or environment
data.
Verified by: TH-HARNESS-AC-074, TH-HARNESS-AC-077

**TH-HARNESS-REQ-282**
Harness-observability artifacts are owner-only diagnostic secret-bearing
artifacts. A partial local generation MUST retain a schema-valid index with
`status=partial` and a closed diagnostic class; it MUST NOT replace a prior
product, harness, cleanup, or artifact result. A normal invocation MAY report
the partial diagnostic as a warning. The explicit observability check MUST
reject it.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-078

**TH-HARNESS-REQ-283**
The top-level wrapper MUST retain
`<run-root>/_shared/harness-invocation-start.json` as
`cartulary.harness_invocation_start.v1` before prerequisite work begins. The
artifact MUST bind the run, target, start timestamp, and a sorted snapshot of
the recursively expanded target-to-prerequisite edges from the generated task
surface. The marker and terminal summary MUST occupy the same frozen run root;
a marker in a sibling generated run ID is artifact-incomplete. Child wrappers
MUST NOT replace the top-level boundary. The terminal
`<run-root>/_shared/harness-observability/execution-context.json` MUST retain
whether that boundary was present and the same edge snapshot. The invocation
root interval MUST use this boundary and the terminal result; target parentage
MUST use the retained edges, explicit summary children, sequence edges, or
scheduler edges and MUST NOT be inferred from temporal containment. A required
performance root without the top-level boundary is artifact-incomplete.

The top-level wrapper MUST retain
`<run-root>/_shared/harness-observability/execution-context.json` as
`cartulary.harness_execution_context.v1` before derived observability output is
accepted. The context MUST contain the exact run, invocation, public target and
command identities; commit and source-snapshot digest; clean or dirty source
state; host, toolchain, externally available capacity, workload/evidence, and
execution-policy digests; start, end, terminal status, interruption, retry
count, warm eligibility, and a bounded sorted contamination-reason set. Source
identity MUST include a sorted retained measurement-contract catalog for every
required public target and separately bound check-internal measurement subject,
with command ID, canonical inputs, measurement profile, eligibility, gates,
workload/evidence digest, and target-scoped execution-policy digest. Externally
available capacity MUST NOT incorporate target-declared
logical scheduler limits; those limits belong to the target-scoped execution
policy. Source
artifact references MUST be normalized relative to the retained run, so an
owner-only root may be moved outside the checkout without exposing absolute
paths. Derived artifacts and all qualification checks MUST consume this retained
context and MUST NOT recompute historical profiles from the current checkout.
Roots without this context or with a mismatched context digest remain
diagnostic-only.
Verified by: TH-HARNESS-AC-073, TH-HARNESS-AC-074, TH-HARNESS-AC-079

### 8.1 Artifact Families

| Artifact family                                      | Producer                                        | Path under run root                                             | Schema policy                                                 | Ordering and nullability                                                              | Retention and cleanup                                        |
| ---------------------------------------------------- | ----------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| Harness invocation start                             | Top-level public preflight                      | `_shared/harness-invocation-start.json`                         | `cartulary.harness_invocation_start.v1`                       | Required run, target, start timestamp, and sorted explicit target-prerequisite edges | Owner-only diagnostic; removed with the run root.             |
| Tool-run summary                                     | Centralized wrappers                            | `<target>/tool-run-summary.json` or the summary directory closed by the target's Section 8.1 row | `cartulary.tool_run_summary.v5`                               | Required non-null timestamps, target, exit code, output mode, artifact refs, failures | Retained; removed by cleanup only under default result root. |
| Fallow static reports                                | `frontend-fallow-static`                        | `frontend-fallow-static/fallow/*` and `frontend-fallow-static/fallow-static-summary.json` | `cartulary.fallow_static_summary.v2` for the normalized summary; raw JSON, SARIF, Markdown, stdout, stderr, and `resolved-fallowrc.json` files are diagnostic-only | Report names, statuses, issue counts, artifact refs, resolved-config refs, baseline state, and enforcement state in schema-defined order | Retained as run-root artifacts; generated/source roots and Fallow config or baseline inputs are not cleanup candidates. |
| Owner summary                                        | Owner accounting handlers                       | `<target>/owners/<owner-id>/test-owner-summary.json`            | `cartulary.test_owner_summary.v1`                             | Stable owner/row/runner/status/count fields                                             | Retained.                                                    |
| Vitest failure details                               | Vitest wrappers                                 | `<target>/raw/<row-id>/vitest-failure-details.json`             | `cartulary.vitest_failure_details.v1`                         | Runner JSON reference, stdout/stderr refs, failed owner path, title, message, source, raw messages, diagnostic tags, and first app frame | Retained when a Vitest runner report exists; absent sidecar uses owner-summary fallback. |
| Target summary                                       | Target summary generator                        | `<target>/target-summary.json`                                  | `cartulary.test_target_summary.v4`                            | Child/totals rollups ordered by registry order                                        | Retained.                                                    |
| Run summary                                          | Run summary generator                           | `run-summary.json` or aggregate dir                             | `cartulary.test_run_summary.v6`                               | Work units and artifact dirs ordered deterministically                                | Retained.                                                    |
| Same-run helper artifact reference                   | Run summary generator                           | `_shared/same-run-helper-artifacts/*.json`                      | `cartulary.same_run_helper_artifact_ref.v2`                    | Helper target, producer work-unit ID, declared inputs, producer artifact digests, consumer refs, same-run scope, and non-scheduler-reuse accounting | Retained with the run root; diagnostic only.                 |
| Harness execution context                            | Top-level wrapper                               | `_shared/harness-observability/execution-context.json`          | `cartulary.harness_execution_context.v2`                       | Immutable invocation, source, environment, normalized target-policy projection, terminal, eligibility, and contamination identity | Owner-only diagnostic; removed with the run root. |
| Harness observability index                          | Observability finalizer                         | `_shared/harness-observability/observability-index.json`        | `cartulary.harness_observability_index.v1`                     | Context ref/digest, required/excluded/out-of-scope targets, invocation refs, source digests, status, and bounded diagnostic in schema order | Owner-only diagnostic; removed with the run root. |
| Harness invocation bundle                            | Observability finalizer                         | `_shared/harness-observability/<invocation-id>/*`               | Native trace and hotspot schemas plus adopted OTLP JSON shape | Deterministic IDs, source refs, spans, metrics, paths, gaps, and digests | Owner-only diagnostic; removed with the run root. |
| Harness sequence event stream                        | Sequence scheduler                              | `<aggregate-target>/sequence-events.jsonl`                      | `cartulary.harness_sequence_event.v1`                          | Strict sequence-local order with one terminal event per started or skipped step | Retained with aggregate target evidence. |
| Harness public-target duration baselines             | Performance maintenance target                  | `tools/harness_public_target_duration_baselines.json`           | `cartulary.harness_public_target_duration_baselines.v2`        | Target-local provenance, timing source, policy projection, discarded warm-up, two measured samples, statistics, gates, and 48-target portfolio total in target order | Checked-in maintenance artifact; regenerated only by its Make target. |
| Scheduler summary                                    | Scheduler                                       | `<target>/scheduler-summary.json`                               | Scheduler summary schema by scheduler type                    | Work units by manifest ordinal; resources by registry order                           | Retained.                                                    |
| Scheduler event stream                               | Scheduler                                       | `<target>/scheduler-events.jsonl`                               | `cartulary.scheduler_event.v7`                                | `seq` strictly increases with no gaps                                                 | Retained.                                                    |
| Scheduler progress summary                           | Scheduler reporter                              | `<target>/progress-summary.log`                                 | diagnostic-only                                               | Bounded progress snapshots                                                            | Retained.                                                    |
| Scheduler pressure summary                           | Scheduler reporter                              | `<target>/pressure-summary.json`                                | `cartulary.scheduler_pressure_summary.v4`                      | Closed current-profile fields are defined below; ordering is target, lane, resource, fixture-class, row fixture pressure, execution-family fixture pressure, fixture proofs, fixture-tier proof artifacts, readiness attribution, and slowest-work lexical order after producer timestamps are normalized | Retained.                                                    |
| Govulncheck findings                                 | Govulncheck wrapper                             | `<target>/<row-id>/govulncheck-findings.json`                    | `cartulary.govulncheck_findings.v1`                            | Finding IDs and counts in deterministic order; symbol-reachable findings are blocking | Retained with the row; promoted by target summary as artifact refs and supplemental security extension data. |
| Agent finalizer summary                              | Agent finalizer                                 | `agent-finalize/finalize-summary.json`                          | `cartulary.agent_finalize_summary.v3`                         | Ordered actions, private substeps, skipped work, cache state, updated files, `RESULTS_DIR`, child artifact refs | Retained.                                                    |
| Cache state artifacts                                | Cache helpers and agent finalizer               | `<target>/*-cache-*.json` when emitted; records under `.cache/cartulary/*` | `cartulary.cache.readiness.v1`, `cartulary.cache.build_artifact.v1`, `cartulary.cache.static_analysis.v1`, `cartulary.agent_finalize_action_cache_record.v1`, or `cartulary.execution_topology_render_cache.v1` | Profile ID, cache state/reason, key digest, input digest, output digest, output paths, and record path in schema-defined order | Run-root artifacts retained; default cache records removed only by `make distclean`. |
| Runtime binary provenance                            | Go target runner                                | `_shared/<execution-family>/runtime-binaries.json` when an aggregate declares runtime binaries | diagnostic-only                                               | Runtime binary ID, scheduler-owned consumer env, producer target, normalized path, file digest, build-artifact ref, and output digest | Retained with the shared Go report. |
| Service scope summary                                | Service suite                                   | `_shared/test-services/<suite-id>/service-scope.json`            | `cartulary.test_services.scope.v1`                         | Suite identity, target, preflight, failure, cleanup, and service summaries closed by Section 11 | Retained; cleanup may append diagnostics.                    |
| Service lifecycle event stream                       | Service suite                                   | `_shared/test-services/<suite-id>/lifecycle-events.jsonl`        | `cartulary.test_services.lifecycle.v2`                        | `seq` strictly increases; transitions match Section 11.2                               | Retained; not cleanup proof.                                |
| Browser startup events                               | Browser session lifecycle                       | `_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/startup-events.jsonl` | `cartulary.browser_startup_event.v1` | Exact suite/session/profile identity and validated append-only state transitions | Retained for the session; lifecycle adapter is sole writer. |
| Browser startup diagnostics                          | Browser session lifecycle                       | `_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/startup-diagnostics.json` | `cartulary.browser_startup_diagnostics.v2` | Immutable terminal state, event reference/digest, classification, redaction-safe message, origins, and artifact references | Retained for the session; group and target evidence consume by reference. |
| Browser stack metadata                               | Browser session lifecycle                       | `_shared/test-services/<suite-id>/browser-sessions/<browser-session-id>/stack-v4.json` | `cartulary.web_e2e_stack.v4` | Immutable suite/session/mode/profile identity, service scope, database, object-store namespace, backend/frontend process proofs, build digest, fixture, diagnostic, lease, and readiness bindings | Retained for current-run attach admission. |
| Browser group result                                 | Browser evidence adapter                        | `<target>/browser-groups/<group-id>/browser-group-result.json` | `cartulary.browser_group_result.v2` | Exact selected rows, terminal observations, and ordered session artifact references/digests | Retained for target accounting. |
| Browser target result                                | Browser evidence finalizer                      | `<target>/browser-target-result.json` | `cartulary.browser_target_result.v1` | Ordered group-result references/digests and deduplicated session artifact references/digests | Retained for target accounting. |
| Local object-store proxy attempt, lease, and health  | Local development proxy lifecycle               | owner-only `.cartulary/runtime/object-store-proxy/` state and loopback health endpoint | `cartulary.local_object_store_proxy_start_attempt.v1`, `cartulary.local_object_store_proxy_lease.v1`, `cartulary.local_object_store_proxy_health.v1` | Canonical nonsecret configuration, instance identity, boot-aware process proof, and readiness state | Development-only; never browser or product evidence. |
| Reset response/status/state                          | Reset route/wrapper                             | `reset-boundary/*.json`, `*.status`, `*.state-reset`            | `cartulary.test.runtime_reset.v1` for reset data              | Reset ID, table list, migration/admin flags, object count required                    | Retained for browser target.                                 |
| Test clock-control response                          | Test clock route                               | clock-control transcript or target-owned clock-control dir       | `cartulary.test.clock_control.v1`                             | Clock mode, current RFC3339 timestamp, offset seconds, and fixed timestamp when mode is fixed | Retained only by the target or fixture transcript that controls the clock; never production API evidence. |
| Network Flow fault-control response                  | Network Flow fault-control route                | Network Flow fixture transcript or target-owned fault-control dir | `cartulary.test.network_flow_fault_control.v1`                | Fault ID, exact boundary token, fault kind, optional safe error code, optional correlation key, and `consume_once=true` | Retained only by the target or fixture transcript that arms the fault; never production API evidence. |
| Network Flow randomness-control response             | Network Flow randomness-control route           | Network Flow fixture transcript or target-owned randomness-control dir | `cartulary.test.network_flow_randomness_control.v1`           | Control ID, exact stream token, value kind, value count, remaining count, `consume_once=true`, and `exhaustion="fail_closed"` | Retained only by the target or fixture transcript that arms deterministic fixture randomness; never production API evidence. |
| Network Flow auth-transition-control response        | Network Flow auth-transition-control route      | Network Flow fixture transcript or target-owned auth-transition-control dir | `cartulary.test.network_flow_auth_transition_control.v1`      | Control ID, exact boundary token, transition kind, actor ref, incident ref, resource kind/ref, hidden response kind, optional correlation key, `must_not_disclose_resource=true`, and `consume_once=true` | Retained only by the target or fixture transcript that arms route-time authorization or hidden-resource assertions; never production API evidence. |
| Network Flow audit-assertion-control response        | Network Flow audit-assertion-control route      | Network Flow fixture transcript or target-owned audit-assertion-control dir | `cartulary.test.network_flow_audit_assertion_control.v1`     | Assertion ID, assertion kind, event code, operation ref, actor ref, incident ref, resource kind/ref, baseline count, expected final count, expected replay increment, optional correlation key, and `consume_once=true` | Retained only by the target or fixture transcript that arms exact-count or replay-silence assertions; never product audit evidence by itself. |
| Frontend accessibility summary                       | Browser accessibility target                    | `browser-e2e-a11y/accessibility/frontend-accessibility-summary.json` | `cartulary.frontend_accessibility_summary.v4`                  | Active `rows[]`, `scenarios[]`, `keyboard_matrix[]`, `state_communication_checks[]`, `contrast_checks[]`, `violations[]`, and `artifact_refs[]` in schema-defined order | Retained for browser target.                                 |
| Test evidence accounting                            | Owner-aware target summaries                    | `<target>/owners/<owner-id>/test-evidence-accounting.json`           | `cartulary.test_evidence_accounting.v1`                        | Selection identity, semantic digests, exact expected and terminal row records, attempts, and required-target closure | Retained for target; target/tool-run summaries reference every owner shard instead of duplicating row details. |
| Release-readiness evidence                           | Release-readiness aggregation                   | `release-readiness-evidence/release-readiness-evidence.json`        | `cartulary.release_readiness_evidence.v2`                      | Evidence records with explicit owner refs, evidence class, product conformance effect, Core 05 publication effect, release-gate effect, run root, artifact refs, and status | Retained for release-readiness target; target/tool-run summaries reference the artifact. |
| Network Flow fixture manifest                        | Network Flow fixture manifest validator         | `fixtures/network-flow/<fixture_id>/manifest.json`                  | `cartulary.network_flow_fixture_manifest.v2`                   | Fixture identity, source files, expected artifacts, transcript files, per-file SHA-256 values, and aggregate bundle hashes in canonical sorted order | Source fixture roots are committed and immutable after freeze; run-local materializations are retained under the selected target's run root. |
| Generated manifest summaries                         | Generation/drift scripts                        | tool-specific target dirs                                       | JSON schemas declared by generated artifacts                  | Unknown fields rejected where shape tools enforce closure                             | Generated files remain checked in; summaries retained.       |
| Logs                                                 | Shell, Go, scheduler, browser, service wrappers | target log dirs                                                 | diagnostic-only unless producer declares schema               | Logs are text after redaction; empty logs may be omitted                              | Retained unless cleanup removes result root.                 |
| Coverage reports                                     | Go/frontend/test tools                          | tool-specific coverage paths                                    | diagnostic-only                                               | No current schema-owned field contract; retained only as tool diagnostic output       | Removed by `make clean` when under registered paths.         |
| Playwright screenshots, videos, traces, HTML reports | Playwright                                      | Playwright report/test-results dirs                             | diagnostic-only secret-bearing                                | No current schema-owned field contract; retained only as Playwright diagnostic output | Removed by `make clean` when under registered paths.         |
| Visual snapshots and goldens                         | Browser/fixture tools                           | source and tool-specific dirs                                   | validation-only; helper refresh is owned by `browser-e2e-visual-update` | No current schema-owned diagnostic schema contract; helper refresh writes tool-specific committed PNGs only | Refresh is helper-only and is not validation evidence until `browser-e2e-visual` passes. |

**TH-HARNESS-REQ-258**
Every target named by a sequence step's `produces_summary_targets[]` MUST retain `<target>/target-summary.json` in the selected run root before the sequence aggregate emits its run summary or aggregate target summary. A target's `<target>/tool-run-summary.json` remains the wrapper-owned tool-run summary. Command-specific reports retained by the target, such as SeaweedFS compatibility or release-gate reports, MUST NOT substitute for `target-summary.json` when the target is a sequence-produced summary target.

Artifact references follow the ownership direction. A target tool-run summary MUST reference artifacts owned by that target and MAY reference its nested scheduler or owner-partition artifacts. Only the aggregate target whose identity equals the enclosing root tool summary's target MAY reference the root `run-summary.json` and root `tool-run-summary.json`; every leaf or nested target MUST omit those parent artifacts. A child target's artifact inventory therefore MUST be identical whether it finalizes before or after the enclosing aggregate artifacts exist.
Verified by: TH-HARNESS-AC-015, TH-HARNESS-AC-023, TH-HARNESS-AC-027

Nested scheduler targets MUST expose their scheduler artifacts under their own target directory even when a parent aggregate also references them. `check-service-backed` MUST retain first-class `check-service-backed/scheduler-summary.json`, `check-service-backed/scheduler-events.jsonl`, and `check-service-backed/pressure-summary.json`; the pressure summary MUST report backend/browser lane timing, fixture class counts, resource-claim counts, planned child totals, executed child totals, and slowest child work. Parent `check` artifacts MAY link to those nested artifacts, but investigation tools MUST NOT require callers to mine a large parent scheduler summary to diagnose `check-service-backed`.

In the current profile, `pressure-summary.json` is a required retained diagnostic artifact with schema-owned field closure. It MUST be a JSON object that validates against `cartulary.scheduler_pressure_summary.v4` before scheduler target success. Older pressure-summary schema IDs are unsupported; diagnostics MAY report their paths as `unsupported_schema` but MUST NOT interpret them as current evidence. Current-profile scheduler work reuse is not adopted: `reused_accounting_counts.reused` MUST be `0`, and the scheduler MUST emit truthful nonnegative `executed` and `skipped` counts where those values are derivable from scheduler records. `readiness_attribution_counts`, `readiness_attribution_duration_ms`, and `readiness_attribution_units` MUST be derived only from scheduler-readable `work_units[].readiness_attribution` metadata. Empty readiness attribution means the selected schedule contains no readiness source metadata to attribute, not that readiness was free. The schema validates harness diagnostics only; the artifact MUST NOT be cited as product-conformance or Core 05 claim-publication evidence.

| Field | Required type | Meaning | Omission/null rule |
| --- | --- | --- | --- |
| `schema_id` | string | Schema-owned producer marker; value is `cartulary.scheduler_pressure_summary.v4`. | MUST be present and non-null. |
| `target` | string | Public or scheduler target whose retained directory contains the artifact. | MUST be present and non-empty. |
| `scheduler_kind` | string | Scheduler family that produced the artifact, using the Section 10.1 closed scheduler-kind values. | MUST be present and non-empty. |
| `status` | string | Final scheduler status, using the same status token as the scheduler summary. | MUST be present and non-empty. |
| `total_work_units` | nonnegative integer | Planned child total after schedule normalization. | MUST be present; zero is allowed only for an empty diagnostic fixture schedule. |
| `completed_work_units` | nonnegative integer | Executed child total observed by the scheduler reporter. | MUST be present and MUST NOT exceed `total_work_units`. |
| `scheduler_total_duration_ms` | nonnegative integer | Scheduler wall-clock duration in milliseconds. | MUST be present. |
| `target_counts` | object | Counts by normalized child target or lane key; keys are non-empty strings and values are nonnegative integers. | MUST be present; an empty object means no child count was attributable. |
| `lane_duration_ms` | object | Backend and browser lane timing by normalized lane key; keys are non-empty strings and values are nonnegative integer milliseconds. | MUST be present; an empty object means no lane timing was attributable. |
| `resource_claim_counts` | object | Aggregate logical resource-claim counts by Section 10.2 resource name. | MUST be present; an empty object means no resource claim was observed. |
| `fixture_class_counts` | object | Counts by closed fixture class: `migration_scratch`, `template_clone`, `package_reset`, `transaction_or_shared_postgres`, or `none`. | MUST be present; absent classes have count `0`. |
| `row_fixture_pressure` | array | Row-level fixture pressure aggregates keyed by `target`, `row_id`, `execution_family`, and effective fixture class, with `work_unit_count` and `duration_ms`. | MUST be present; an empty array means no scheduler work had row attribution. |
| `execution_family_fixture_pressure` | array | Execution-family fixture pressure aggregates keyed by `target`, `execution_family`, and effective fixture class, with `work_unit_count` and `duration_ms`. | MUST be present; an empty array means no scheduler work had execution-family attribution. |
| `fixture_proof_records` | array | Retained fixture-proof diagnostics for candidate row or symbol decisions. Each record MUST name `target`, `row_id`, `execution_family`, optional `symbol`, effective `fixture_policy`, `proof_kind`, `proof_status`, optional `proof_ref`, `reason`, and optional `dirty_tables[]`. | MUST be present; an empty array means no candidate proof diagnostics were emitted for the target. |
| `fixture_tier_proofs` | array | Retained `cartulary.fixture_tier_proof.v2` objects for schema-owned fixture-tier proof decisions accepted by the current scheduler reporter. Each object MUST identify target, owner ID, row ID, execution family, optional symbol, effective fixture policy, proof kind/status, execution boundary, observed surfaces, reset surface, and final verdict. | MUST be present; an empty array means no retained fixture-tier proof artifacts were emitted for the target. |
| `slowest_work_units` | array | Slowest work-unit diagnostics ordered by descending `duration_ms`, then `label` lexical order. Each item MUST include at least `id`, `label`, `status`, and `duration_ms`. | MUST be present; an empty array is allowed only when no work unit executed. |
| `reused_accounting_counts` | object | Reuse accounting counts; current closed keys are `executed`, `reused`, and `skipped`, each a nonnegative integer. | MUST be present with all three keys; `reused` MUST be `0` until scheduler work reuse is owner-adopted. |
| `readiness_attribution_counts` | object | Readiness attribution counts by scheduler-declared readiness class. Keys are non-empty strings and values are nonnegative integers. | MUST be present; an empty object means the selected schedule has no readiness source metadata to attribute. |
| `readiness_attribution_duration_ms` | object | Duration totals by scheduler-declared readiness class. Keys are non-empty strings and values are nonnegative integer milliseconds. | MUST be present; an empty object means the selected schedule has no readiness source metadata to attribute. |
| `readiness_attribution_units` | array | Unit-level readiness/provisioning records derived from manifest metadata. Each item MUST name `id`, `label`, `timing_role`, `readiness_class`, `duration_ms`, `warm_threshold_ms`, and `warm_status`; `reason` MUST be present when declared by the manifest. | MUST be present; empty only when no completed work unit has scheduler-readable readiness metadata. |
| `generated_at` | RFC 3339 UTC string | Time the pressure summary object was generated. | MUST be present and non-empty. |

**TH-HARNESS-REQ-259**
Same-run helper artifact reuse is adopted only for helper/setup artifacts produced earlier in the same retained run root. A consumer MUST retain a `cartulary.same_run_helper_artifact_ref.v2` object before reporting helper reuse. The object MUST name the current `run_id` and `run_root`, the helper target, the producer work-unit ID, declared input artifact refs, producer artifact refs with SHA-256 digests, consumer refs, an output digest, `reuse_scope="same_run_only"`, `accounting_mode` of `derived` or `helper_reused`, and `scheduler_reused=false`.

The producer artifacts referenced by a same-run helper artifact ref MUST resolve under the current run root and MUST exist before the consumer aggregate succeeds. Missing artifacts, paths outside the current run root, malformed refs, digest mismatches, or retained-run references from any prior run MUST fail closed as artifact/configuration errors rather than silently falling back to old evidence. A same-run helper artifact ref MAY reduce duplicate helper/setup work inside one aggregate, but it MUST NOT skip any selected production-conformance row, product test, browser/live-state row, service-backed row, drift verdict, security verdict, cleanup, runtime reset, scratch migration apply, object-store mutation, or destructive-operation safeguard.

Same-run helper artifact refs are not cache records and are not scheduler work reuse. Aggregates MAY expose them in run summaries and `explain-run`, but scheduler summaries and pressure summaries MUST continue to report current-profile scheduler `reused` counts as `0` unless a later NLSpec revision adopts scheduler work reuse separately.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-025, TH-HARNESS-AC-028, TH-HARNESS-AC-048

### 8.2 Network Flow Fixture Manifest Contract

**TH-HARNESS-REQ-262**
Network Flow conformance fixtures MUST use a directory-scoped manifest at `fixtures/network-flow/<fixture_id>/manifest.json`. The `<fixture_id>` directory name and manifest `fixture_id` MUST be identical, MUST use the full `NF-FIX-###-slug` identifier, and MUST NOT be inferred from source filenames, display names, route labels, generated output names, or test titles. Legacy single-file fixture locators are not canonical manifest identity.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-263**
A Network Flow fixture manifest MUST validate as
`cartulary.network_flow_fixture_manifest.v2`, MUST be schema-closed at every
object boundary, and MUST declare `profile_id="network_flow_activity"`,
`manifest_version=2`, `freeze.status`, `freeze.revision`, `source_files[]`,
`expected_artifacts[]`, `transcript_files[]`, `source_bundle_sha256`, and
`expected_bundle_sha256`. It MUST omit acceptance IDs, verification IDs,
copied requirements, execution selectors, phase selectors, and specification
provenance. Each listed scenario file MUST validate as
`cartulary.network_flow_fixture_scenario.v2` and contain only fixture identity
and a human-readable behavior summary. File arrays MUST be ordered by
`logical_path` ascending by Unicode code point, and each listed file MUST
declare exact byte `size_bytes` and lowercase hex SHA-256 of the committed file
bytes.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-264**
The Network Flow source bundle digest algorithm is `network_flow_fixture_bundle_hash_v1`. For each `source_files[]` entry in manifest order, the validator hashes the UTF-8 frame `logical_path`, a NUL byte, lowercase `sha256`, a NUL byte, decimal `size_bytes`, and LF. `source_bundle_sha256` MUST equal the SHA-256 of the concatenated frames. `expected_bundle_sha256` uses the same frame algorithm over `expected_artifacts[]` followed by `transcript_files[]`, preserving each array's manifest order. A missing file, extra unlisted file, digest mismatch, size mismatch, unsorted list, duplicate path, absolute path, symlink, or traversal path MUST fail before product code starts.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-265**
Only `freeze.status="frozen"` Network Flow manifests may close Network Flow
behavior evidence. A frozen manifest is append-only by revision: any byte
change to source files, expected artifacts, transcript files, or aggregate
digests requires a new `freeze.revision` and a tracker entry that names the
changed fixture. Draft manifests MAY exist during fixture authoring, but public
verification targets MUST report them as blocked rather than treating draft
bytes as current evidence.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-266**
Network Flow fixture execution MUST materialize manifest-listed files into a run-local read-only input workspace under the selected result root. Product tests MUST read the run-local materialization, not mutate the committed fixture directory. Expected artifacts and transcript files are read-only comparison inputs. The runner MUST retain a bounded execution summary that names selected fixture IDs, manifest file SHA-256, source and expected bundle SHA-256 values, materialized input root, produced artifact refs, and comparison status. This summary is harness evidence only and does not define Network Flow product behavior.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-267**
Network Flow fixture manifests carry integrity and execution data; they do not
route evidence or own domain semantics. Test-catalog rows route executable
fixture tests through the Network Flow behavior verification. A manifest MUST
NOT redefine import parsing, row identity, cursor behavior, graph behavior,
indicator binding, authorization, retention, audit occurrence rules, or
generated contract shape.
Verified by: TH-HARNESS-AC-049

### 8.3 Agent Finalizer

**TH-HARNESS-REQ-260**
`agent-finalize` is a harness-maintenance finalizer. It refreshes and validates deterministic harness-maintenance artifacts before a caller runs explicit verification. It MUST NOT be described or implemented as a verification gate, test runner, cleanup target, code-generation workflow, migration workflow, release gate, security gate, build gate, browser E2E surface, or benchmark-claim surface.
Verified by: TH-HARNESS-AC-019

**TH-HARNESS-REQ-261**
`agent-finalize` exposes exactly one semantic operation: finalize harness-maintenance artifacts before explicit verification. Its public input surface is `make agent-finalize` with optional `RESULTS_DIR=<run_root>` and the ordinary output-mode controls. Callers MUST NOT select finalizer substeps by child target name.

`agent-finalize` MUST derive its execution plan from the closed finalizer action registry below.

| Action ID | Requires `RESULTS_DIR` | Mutating | Cache eligible | Input profile ID | Action contract version | Required behavior | Allowed output |
| --- | ---: | ---: | ---: | --- | --- | --- | --- |
| `scheduler_drift_validation` | yes | no | yes | `agent_finalize.scheduler_drift_validation.v1` | `v1` | Validate scheduler event ordering and warm-check timing health against the retained run before retained-run mutations are allowed. | Finalizer summary and child summaries. |
| `generated_structure_refresh` | no | yes | yes | `agent_finalize.generated_structure_refresh.v2` | `v2` | Refresh catalog-derived topology and schedules, then verify no unsupported drift remains. | Finalizer summary, child summaries, updated-file list. |
| `schema_shape_validation` | no | no | yes | `agent_finalize.schema_shape_validation.v1` | `v1` | Validate harness-owned JSON shape and schema attachments needed by the finalizer path. | Finalizer summary and child summaries. |
| `duration_baseline_refresh` | yes | yes | yes | `agent_finalize.duration_baseline_refresh.v3` | `v3` | Refresh only advisory harness duration-baseline artifacts from a successful, uncontaminated retained run. | Finalizer summary, child summaries, updated-file list. |
| `duration_baseline_coverage` | no | no | yes | `agent_finalize.duration_baseline_coverage.v1` | `v1` | Verify that required advisory duration-baseline entries exist or are explicitly defaulted. | Finalizer summary and child summaries. |
| `duration_baseline_drift_validation` | yes | no | yes | `agent_finalize.duration_baseline_drift_validation.v2` | `v2` | Validate advisory duration-baseline freshness against the retained run without promoting the baselines to benchmark claims or product performance conformance evidence. | Finalizer summary and child summaries. |

The implementation MAY realize an action by invoking one or more Make targets or scripts. Child target names are not part of the `agent-finalize` public contract unless this NLSpec explicitly promotes them.
Verified by: TH-HARNESS-AC-019, TH-HARNESS-AC-020

`agent-finalize` MAY reuse a prior result for a cache-eligible finalizer action only at the finalizer `action_id` boundary. It MUST NOT cache or identify work by child Make target name, script path, package command, or raw tool invocation. A cache hit MUST prove that the action is cache-eligible, the action's input profile is closed, declared repo inputs and implementation bytes match, relevant tool identity and environment values match, declared output artifacts already exist and match the cached output digest, the cache record validates against the current cache schema, and any required retained-run evidence digest matches after retained-run validation succeeds.

Action-cache use MUST be disabled when `CI=1` unless a later adopted harness revision defines CI cache semantics. It MUST be disabled locally when `CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1`. Cache records MUST live under `.cache/cartulary/agent-finalize-action-cache` by default, MUST validate against `cartulary.agent_finalize_action_cache_record.v1` before reuse, and MUST be treated as local acceleration evidence only, not product verification evidence.

A cache hit MUST mark the action `execution_state="reused"` in `agent-finalize/finalize-summary.json`. It MUST NOT report reused work as zero-duration executed child work. Substeps skipped because of an action-cache hit MUST be marked skipped with a cache-hit skipped reason and MUST NOT require child logs from the current run. The action cache state MUST be reported with closed states `hit`, `miss`, `bypass`, `disabled`, `corrupt`, or `ineligible`, plus a closed reason code and the cache key, input profile, input digest, output digest, action contract version, cache schema ID, and record path when a key was computed.

Cache records that are missing or whose inputs changed MUST miss and run normally. Cache records that are malformed, fail schema/key validation, or disagree with current digests MUST be reported as `corrupt`; they may run normally only when ordinary action execution is safe, and they MUST never produce success by reuse. Missing or changed required outputs MUST miss and run normal refresh or validation behavior.

When `RESULTS_DIR` is set, `agent-finalize` MUST validate the supplied retained run and run retained-run scheduler health before any mutating generated-artifact or duration-baseline refresh. A valid finalizer `RESULTS_DIR` is an existing retained full warm `make check` run root that identifies a successful, uncontaminated run and contains the artifact families required by the selected refresh and warm scheduler checks. The retained root MUST be the latest successful full warm `make check` retained run under the same retained-results parent unless `ALLOW_OLDER_RESULTS_DIR=1` is supplied and recorded in the finalizer summary. Without that override, an older retained root MUST fail before mutating refresh work. The retained root MUST contain a passing `check/tool-run-summary.json`, `check/scheduler-summary.json`, and `check/scheduler-events.jsonl`; service-backed-only, owner-slice, browser-only, and other partial run roots MUST NOT be accepted as finalizer retained-run maintenance input. It MUST also contain paired, schema-valid `test-evidence-accounting.json` and `test-owner-summary.json` owner partitions. Every consumed partition MUST record successful row closure, match its summary identity, and match the current source-snapshot, catalog, and verification-contract digests; runtime, resource, and fixture profile digests MAY differ between partitions but MUST match within each accounting/summary pair. The finalizer MUST reject missing paths, non-directory paths, missing full-check markers, failed retained target summaries, missing required scheduler/target/owner-accounting families, incompatible owner evidence, contaminated service timing evidence, older retained roots without explicit override, and retained evidence that cannot support `scheduler-summary-timing-drift TARGET=check`. Rejection before a mutating refresh MUST fail with `configuration_error` for invalid caller input or `artifact_error` for unsafe retained evidence.

After retained-run preflight succeeds, the finalizer MUST freeze the accepted pre-mutation source-snapshot identity for every retained-evidence child action. A later child that reads owner accounting after an earlier baseline writer intentionally changes tracked maintenance bytes MUST compare that accounting with the frozen identity, not a newly computed mid-transaction snapshot. The frozen identity is private finalizer context: it MUST be scoped to the exact normalized retained root, MUST fail closed when malformed or paired with a different root, and MUST NOT alter standalone public baseline commands, which continue to compare retained evidence with the current source snapshot.

When `agent-finalize/finalize-summary.json` records a normalized child failure, the enclosing shell step and public target summaries MUST promote that failure class, reason, headline, and child target. The generic nonzero shell wrapper failure MUST NOT outrank or duplicate the normalized finalizer failure as `unknown_failure`; raw wrapper stdout and stderr remain retained log artifacts.

For retained finalization, `duration_baseline_refresh` MUST complete before `generated_structure_refresh`. The baseline action MUST own only the four advisory baseline writers; it MUST NOT invoke generation or generated drift or claim generated topology/schedule outputs in its action-cache boundary. `generated_structure_refresh` MUST then invoke `generate` and `generate-drift` exactly once against the stabilized baseline bytes. A retained finalizer MUST NOT invoke the same public child target twice in one prepared run identity across an intentional tracked-source mutation.
Verified by: TH-HARNESS-AC-019

Duration-baseline refreshes remain advisory harness planning data and MUST NOT become benchmark claims or product performance conformance evidence. Browser entry baselines MUST use semantic catalog row IDs, and scheduler work-unit baselines MUST use current generated work-unit IDs. Phase-keyed, `E-*`, and `FE-*` duration identities are unsupported. A baseline refresh MUST consume only compatible successful owner and scheduler evidence; obsolete entries MUST be removed or ignored rather than carried forward by compatibility translation.
Verified by: TH-HARNESS-AC-019

Before any selected finalizer action can mutate tracked generated or baseline files, `agent-finalize` MUST snapshot or stage the tracked worktree state it is allowed to change. If any later selected action fails, the finalizer MUST restore tracked files to the start-of-run state before writing `finalize-summary.json`; pre-existing tracked user changes MUST be preserved, and untracked files MUST NOT be deleted as part of rollback. The finalizer summary MUST report rollback status with `mutation_rollback.status`, `restored_file_count`, `restored_files[]`, and `error`.
Verified by: TH-HARNESS-AC-019

When `agent-finalize` mutates tracked generated or baseline artifacts and completes successfully, those promoted mutations MUST be explicit in `finalize-summary.json` through `generated.updated_file_count` and `updated_files[]`. Audit or handoff records that cite an `agent-finalize` run MUST distinguish pre-existing worktree changes from finalizer-caused updates. A finalizer update MUST NOT be treated as silent remediation merely because the command succeeded.
Verified by: TH-HARNESS-AC-019

The `generated_structure_refresh` action MUST invoke the ordinary `generate` and `generate-drift` ownership surfaces. `agent-finalize` MUST NOT invoke `format`, `migration-drift`, `test-fast`, `test`, `check`, `ci`, `release-check`, browser E2E targets, security scan targets, build targets, `clean`, `distclean`, or `benchmark-claim-check`.
Verified by: TH-HARNESS-AC-019

`agent-finalize` MUST be fail-fast and resumable. It MUST stop at the first failed substep, preserve completed substeps, mark later selected substeps as skipped-after-failure, retain child logs and child summaries when available, and propagate the normalized child failure class and reason when a failed child summary is readable. Summary-write or cleanup-reporting failures MUST be reported without masking an earlier primary child failure.
Verified by: TH-HARNESS-AC-019

`agent-finalize` MUST retain `agent-finalize/finalize-summary.json` with schema ID `cartulary.agent_finalize_summary.v3`. The summary MUST include selected and not-selected actions, private substeps, skipped work, per-action cache state, updated files, `RESULTS_DIR`, failure records, and child artifact references. Human summary output MUST include one compact `[FINALIZE]` line before the ordinary `[RESULT]` line and MUST expose reused-action and cache-hit counts when cache evidence exists. Machine output MUST remain exactly one `cartulary.tool_run_summary.v5` JSON object; callers MUST use the `finalize_summary` artifact reference to read command-specific finalizer details.
Verified by: TH-HARNESS-AC-004, TH-HARNESS-AC-019

Check scheduler summaries that include service sessions MUST report service-suite setup timing separately from child test timing. Each `service_sessions[]` entry MUST include `setup_duration_ms`, `ready_at_monotonic_ms`, `child_work_started_at_monotonic_ms`, and `cleanup_duration_ms`; fields MAY be `null` only when the corresponding lifecycle segment did not run or did not reach readiness. Duration baselines for backend, browser, and service-backed child work MUST be derived from child work-unit timings, not from service-suite setup or cleanup timing.

Scheduler-owned service-session startup MUST retain a minimal redacted service-session environment diagnostic before Docker preflight or managed-service startup begins. If startup fails before a lease can be written, the service-session summary MAY report no lease path only when service cleanup is explicitly recorded as `skipped_no_lease` and the failure artifact points to the current-run `service-scope.json`. Summary artifact references MUST NOT point investigators only at paths that cannot exist for the observed startup stage.

## 9. Failure Classes and Exit Codes

**TH-HARNESS-REQ-300**
Public Make-owned wrappers MUST expose exact normalized public exit codes according to the failure-reason table below in retained summaries and compact failure output. Raw child process exit codes MAY be preserved in summaries but MUST NOT define the normalized public exit code except where `child_target_failure` explicitly delegates to a normalized child failure class.
Verified by: TH-HARNESS-AC-014

Public exit-code selection is reason-based. Wrappers MUST derive the
normalized public exit code from the normalized `failure_reason` and primary-failure
rules in this section, not from the raw process status of a child command. The current GNU Make invocation binding may return GNU Make's executor failure status for a failed recipe; callers that require reason-specific failure codes MUST read the retained `tool-run-summary.json` or compact failure line rather than treating the outer `make` process status as the normalized public exit code.

Scheduler summaries MUST propagate normalized failures from every completed failed work unit whose child target summary is available. For a failed `service_session` work unit, a current-run `cartulary.test_services.scope.v1` artifact is also an authoritative child diagnostic source when startup fails before a child target summary exists. Scheduler failure collection MUST read the schema-valid service scope from the same retained run root, emit the service failure as an ordinary `failures[]` record, and point its artifact reference at `service-scope.json`. The scheduler's own fallback classification is used only when no child summary or current-run service diagnostic exists, the diagnostic is unreadable, or the failure belongs to scheduler orchestration rather than completed child target work. The summary's primary failure still follows Section 9.1 ordering, but `failures[]`, `failure_classes`, and `failure_reasons` MUST retain all completed failed work units. Dependency-skipped completion rows MUST NOT be counted as independent failures when the underlying failed work unit is already represented. A child target assertion failure therefore remains `failure_class=product` and `failure_reason=test_assertion_failure` at the scheduler summary layer, while a concurrent config, harness, artifact, timing, infra, or security failure remains visible in the same scheduler and aggregate summaries.

Every failed retained summary that carries the standard failure fields MUST expose both a non-null `failure_class` and a non-null `failure_reason`. Passing summaries MUST expose no primary failure. A generic shell-wrapper exit such as `command exited with status 1` is diagnostic wrapper evidence when a tool runner has already emitted a classified failure for the same target; it MUST NOT become the primary failure or an independent primary harness failure.

Post-summary scheduler validation failures, including scheduler event, timing, critical-path, summary, or accounting drift detected after child work has completed, MUST be normalized as `failure_class=harness`, `failure_reason=scheduler_accounting_error`, and public exit code `11`. They MUST NOT fall through as caller `configuration_error` merely because they are detected by the scheduler runner.

Failure classification uses two layers:

- `failure_class`: coarse stable grouping for humans and automation.
- `failure_reason`: detailed snake-case reason for diagnosis, exit-code mapping, and handoff.

| Failure class | Meaning                                                                                 |
| ------------- | --------------------------------------------------------------------------------------- |
| `product`     | The product behavior under test failed after harness setup completed.                   |
| `security`    | A blocking security scanner finding was reported after harness setup completed.         |
| `config`      | Caller input, environment, manifest, or local tool configuration was invalid or missing. |
| `infra`       | Required backing infrastructure failed preflight, startup, readiness, or capacity.      |
| `harness`     | Harness orchestration, fixture, scheduler, child aggregation, or cleanup failed.        |
| `artifact`    | Required retained evidence was missing, malformed, invalid, or unsafe.                 |
| `timing`      | A deadline, timeout, or timing-accounting guard failed.                                |
| `interrupted` | The command was cancelled or interrupted.                                               |
| `unknown`     | The wrapper could not classify the failure.                                             |

| Failure reason                | Default class | Trigger                                                            |                                    Public exit code |
| ----------------------------- | ------------- | ------------------------------------------------------------------ | --------------------------------------------------: |
| success                       | none          | No failure                                                         |                                                 `0` |
| `usage_error`                 | `config`      | Invalid arguments, missing required flags, unsupported output mode |                                                 `2` |
| `configuration_error`         | `config`      | Missing/invalid tool, path, env, config, manifest, resource limit  |                                                 `2` |
| `preflight_error`             | `infra`       | Docker/platform/tool preflight fails before managed services       |                                                 `3` |
| `service_start_error`         | `infra`       | Backing service or browser process fails to start                  |                                                 `3` |
| `service_readiness_timeout`   | `infra`       | Started service fails readiness before deadline                    |                                                 `3` |
| `fixture_error`               | `harness`     | DB/bucket/template/reset/janitor/fixture operation or shape validation fails |                                      `3` |
| `resource_conflict`           | `infra`       | Logical resource, port, lock, DB/bucket name, or host conflict     |                                                 `4` |
| `test_assertion_failure`      | `product`     | Test runner assertion fails after harness setup                    |                                                `10` |
| `security_finding`            | `security`    | Blocking Govulncheck vulnerability or enforcing Gosec finding after scanner setup; Fallow uses this reason only if a later adopted security-scan profile selects it |                         `1` |
| `child_target_failure`        | `harness`     | Aggregate child exits nonzero                                      |                         normalized child class exit |
| `tool_diagnostic_failure`     | `harness`     | ShellCheck/Biome/Fallow-style static-analysis, formatter, linter, or tool diagnostic failure after setup |                         `1` |
| `scheduler_accounting_error`  | `harness`     | Manifest, summary, timing, event, or accounting mismatch           |                                                `11` |
| `boundary_policy_violation`   | `harness`     | Executable code or validation reads documentation outside the exact Section 15 exception registry |                         `11` |
| `test_accounting_unmapped`    | `harness`     | Public target completed with executed tests that were neither mapped nor intentionally classified |                  `11` |
| `artifact_error`              | `artifact`    | Required artifact missing, invalid, unredacted, or schema-invalid  |                                                `11` |
| `cleanup_error`               | `harness`     | Cleanup command/finalizer/leak check/reaper scheduling fails       |         `12` when no earlier primary failure exists |
| `duration_baseline_drift`     | `timing`      | Explicit duration-baseline or warm scheduler timing drift check fails |                                             `13` |
| `timeout_failure`             | `timing`      | Command, readiness, watchdog, cleanup, or lock exceeds deadline    |                                                `13` |
| `cancelled_or_interrupted`    | `interrupted` | Signal, cancellation, abort                                        | `130` for SIGINT, `143` for SIGTERM, otherwise `15` |
| `unknown_failure`             | `unknown`     | Failure cannot be classified                                       |                                                 `1` |

Default human output MUST expose bounded failure fields for a failed public target before GNU Make's generic recipe failure line can be the only visible diagnostic. It MUST omit full failure records unless verbose output is requested. The canonical compact shape is:

```text
failure_class=infra reason=service_readiness_timeout failed=<unit>
```

Full failure records belong in retained JSON summaries and investigation commands.

For `frontend-fallow-static`, missing Fallow tools, missing install state, invalid `.fallowrc.json`, invalid reachability owner input, missing owner-declared reachability files, or invalid public asset references MUST map to `configuration_error`; enforcing static findings MUST map to `tool_diagnostic_failure`; malformed raw reports or invalid normalized summaries MUST map to `artifact_error`. Non-blocking Fallow findings MAY be retained as warning evidence without failing the target.

**TH-HARNESS-REQ-305**
Missing or unknown owners, malformed `ROWS`, duplicate rows, cross-owner rows, explicit non-service rows on `service-backed-test-slice`, invalid worker bounds, invalid target-local JSON, removed inputs, and zero-row selections MUST use `failure_class=config`, `failure_reason=usage_error`, and exit `2` before setup.
Verified by: TH-HARNESS-AC-064

**TH-HARNESS-REQ-306**
Invalid owner manifests, verification contracts, schema IDs, runner definitions, selector resolution, profile references, duplicate case ownership, and policy registries MUST use `failure_class=config`, `failure_reason=configuration_error`, and exit `2` before child execution.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-063

**TH-HARNESS-REQ-307**
A product assertion after successful setup uses `test_assertion_failure` and exit `10`. Setup, service, fixture, browser, and scheduler infrastructure failures retain their existing Section 9 reasons. Missing, duplicate, or contradictory terminal row records use `scheduler_accounting_error` and exit `11`.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-308**
Missing, stale, mixed, ambiguous, unsafe, or semantically incompatible retained evidence uses `artifact_error` and exit `11`. Invalid caller path syntax remains `usage_error`; syntactically valid paths with incompatible contents are artifact failures.
Verified by: TH-HARNESS-AC-066

**TH-HARNESS-REQ-309**
An unauthorized documentation read detected by the Section 15 policy uses `boundary_policy_violation` and exit `11`. An invalid documentation exception or semantic allowlist is `configuration_error` before the scan. The policy failure MUST name the consumer and normalized path without disclosing document contents.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-310**
Configuration failures occur before setup and therefore precede no child failure. Once child execution begins, TH-HARNESS-REQ-304 primary-failure precedence remains unchanged. Cleanup failure MUST NOT replace an earlier primary failure; it becomes exit `12` only when no earlier failure exists.
Verified by: TH-HARNESS-AC-065

### 9.1 Primary Failure Selection

```text
select_primary_failure(failures):
  if failures is empty:
    return success
  if any non-cleanup failure exists:
    return non-cleanup failure by:
      1. failure-class precedence from TH-HARNESS-REQ-304,
      2. top-level command lifecycle order from TH-HARNESS-REQ-304,
      3. scheduler event sequence if scheduler-owned,
      4. child target registry order if aggregate-owned,
      5. artifact path lexical order,
      6. failure reason lexical order
  return earliest cleanup failure by cleanup step order
```

**TH-HARNESS-REQ-301**
Cleanup failure after an earlier product or operational failure MUST be recorded but MUST NOT override the public exit code selected for the earlier primary failure.
Verified by: TH-HARNESS-AC-014

**TH-HARNESS-REQ-302**
Harness setup, readiness, fixture, artifact, scheduler, timeout, and cleanup failures MUST NOT use `failure_class=product`. A failing assertion after successful harness setup MUST be classified with `failure_class=product` and `failure_reason=test_assertion_failure`.

Vitest and Playwright per-test timeouts after the test runner has reached product execution are product test failures and MUST be classified as `failure_class=product` with `failure_reason=test_assertion_failure`. Harness-owned watchdogs, command deadlines, lock deadlines, service readiness deadlines, and cleanup deadlines remain operational failures and MUST use `failure_reason=timeout_failure` or `failure_reason=service_readiness_timeout` according to the failure-reason table.

Vitest assertion summaries that contain reporter stack-formatting markers such as `STACK_TRACE_ERROR` MUST preserve actionable assertion context when the runner report provides it. The retained diagnostic MUST keep the assertion title, owner path, raw runner-report reference, reproduce command, and a diagnostic tag such as `vitest_stack_trace_error`; the stack-formatting marker MUST NOT replace an available assertion message as the primary human diagnostic.

Vitest wrappers MAY retain `cartulary.vitest_failure_details.v1` sidecars under their raw artifact directory. When present and schema-valid, owner summaries MUST prefer a matching sidecar assertion message over the runner JSON's stack-formatting marker and MUST retain both the runner JSON and sidecar references in failure details. Missing sidecars remain compatible; summaries MUST continue to use the runner JSON diagnostic fallback when no sidecar entry is available.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-014, TH-HARNESS-AC-021

**TH-HARNESS-REQ-303**
A scheduler MUST preserve the first failed work unit and its retained detail as `failed_work_unit` and `failed_work_unit_detail` even when later sibling work drains and also fails. Scheduler summaries MUST include a bounded `observed_failed_work_units[]` array containing completed nonzero-exit work units in finish order, including later drained sibling failures; that array is diagnostic and MUST NOT rewrite the first failed work unit. Human failure output, scheduler summaries, target summaries, and tool-run summaries MUST choose a primary headline and public exit code from the primary-failure rules without contradicting `failed_work_unit` when the failed work unit has retained classified child evidence.

Scheduler `critical_path_wall_duration_ms` is the scheduler timing envelope and MUST equal the scheduler total duration for every emitted scheduler summary, including failed schedules. When a failed schedule completes no successful work unit, `critical_path_units[]` MUST be empty and `critical_path_terminal_unit` MUST be `null`; the failed path remains represented by `failed_work_unit_detail` and `observed_failed_work_units[]`.
Verified by: TH-HARNESS-AC-014, TH-HARNESS-AC-024

**TH-HARNESS-REQ-304**
Primary-failure precedence is closed. Failure-class precedence is exactly: `product`, `security`, `config`, `infra`, `harness`, `artifact`, `timing`, `interrupted`, `unknown`. Top-level command lifecycle order is exactly: wrapper identity, output-mode resolution, configuration resolution, result-root/run-ID resolution, redaction initialization, semantic target behavior, artifact validation, cleanup or finalizers, public output emission.

When class and lifecycle step tie, scheduler-owned failures order by scheduler event sequence; aggregate-owned child failures order by public child target registry order; artifact failures order by normalized artifact path lexical order; remaining ties order by `failure_reason` lexical order. A cleanup or finalizer failure MUST NOT override an earlier non-cleanup primary failure.
Verified by: TH-HARNESS-AC-014, TH-HARNESS-AC-032

## 10. Scheduler Contract

**TH-HARNESS-REQ-350**
Scheduler manifests are normative scheduler inputs. A scheduler target MUST validate manifest schema, work-unit IDs, dependencies, resource claims, finalizers, output schemas, and timing settings before starting child work.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-024

**TH-HARNESS-REQ-360**
Owner selection begins after Section 5 input validation. For `test-slice`, explicit `ROWS` selects exactly those rows; omitted `ROWS` selects every active executable row owned by `OWNER`. For `service-backed-test-slice`, explicit `ROWS` MUST all reference runtime profiles with `managed_services_required=true`; omitted `ROWS` selects every active owner row with that property. `default_check` MUST NOT affect either omission rule. A zero-row result is `usage_error`.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-065

**TH-HARNESS-REQ-361**
Selection MUST set `selection_mode="default_owner"` for omitted `ROWS` and `selection_mode="exact_rows"` otherwise. `completion_scope` is `full_owner` for omitted selection over the command's complete dependency scope and `selected_subset` for explicit rows. `dependency_scope` is `all` for `test-slice` and `service_backed` for `service-backed-test-slice`.
Verified by: TH-HARNESS-AC-064, TH-HARNESS-AC-065

**TH-HARNESS-REQ-362**
Before setup, the planner MUST resolve every verification reference, runner, selector, runtime profile, resource profile, fixture profile, logical resource, and expected artifact schema. It MUST reject zero or multiple selector resolution, overlapping active ownership, dependency cycles, missing completion keys, and incompatible profile combinations. No child may start after a preflight failure.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-065

**TH-HARNESS-REQ-363**
The planner MUST sort resolved rows by `row_id`, construct deterministic work-unit and dependency IDs, derive exact runtime-binary prerequisites from the authored execution topology, and retain a valid immutable `cartulary.test_slice_plan.v2` before setup. The same semantic inputs MUST produce byte-identical plan content except for fields that the schema explicitly classifies as diagnostic timestamps or run identity.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-071

**TH-HARNESS-REQ-364**
Every resolved row MUST end in exactly one state:

| State | Required meaning |
| --- | --- |
| `passed` | Every required selected case executed and passed, and accounting is complete. |
| `failed` | A product assertion failed or a runner emitted an unauthorized skip. |
| `infrastructure_failed` | Setup, service, fixture, browser, or harness infrastructure prevented execution. |
| `skipped_dependency` | A declared failed dependency prevented the row from starting. |
| `cancelled` | A supported signal or explicit cancellation stopped execution. |
| `skipped_authorized` | One exact unexpired verification-level skip authorization applies. |

No other terminal token is current conformance.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-365**
Owner-slice success requires every resolved row to be `passed` or validly `skipped_authorized`. Missing records, duplicate records, dependency skips, cancellation, infrastructure failure, and unauthorized skips fail the invocation. A row passes only when every required selected case passes and required artifacts validate.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-366**
The current profile performs no automatic product-test retry. Existing service-startup and readiness retry policies in Section 11 remain unchanged. If a runner or service creates more than one attempt, every attempt MUST be retained and only the final row state participates in aggregation.
Verified by: TH-HARNESS-AC-065

**TH-HARNESS-REQ-367**
Both owner-slice commands use `stop_on_first_failure=false`. Running independent work drains after a failure; work whose dependency failed does not start. Cancellation MUST propagate through the existing scheduler contract. Finalizers and cleanup MUST always run, subject to the existing destructive-safety guards.

The owner-slice commands and direct browser-stage commands MUST use the shared scheduler family `test_slice`. Owner plans MUST retain `plan_semantic_digest` and `scheduler_semantic_digest`. These digests exclude invocation timestamps and run IDs, and identical semantic selections, profiles, work units, dependencies, resource claims, timeouts, and finalizers MUST produce identical digests. Browser scheduler units MUST preserve the catalog-derived browser session group, runtime profile, exact selected group set, stage serialization resource, and evidence target. A browser evidence finalizer MUST run after every session reaches a terminal state, and the parent target summary MUST run after every evidence finalizer reaches a terminal state.

Resource admission is FIFO for conflicting claims: a later ready unit MUST NOT take capacity needed by an earlier ready resource-blocked unit when their claims overlap. Ready units with disjoint claims MAY proceed. Product work has one scheduler attempt. A dependency failure emits `skipped_dependency`; an interrupt emits `cancelled`; an ordinary scheduler watchdog emits `infrastructure_failed` with `timeout_failure`. Finalizers become ready when every declared dependency is terminal, whether successful, failed, or dependency-skipped. They run in deterministic manifest order after ordinary running work drains. Cleanup failure is primary only when no earlier higher-precedence failure exists.

The normative command algorithm is:

```text
validate command inputs
resolve OWNER
parse and validate explicit ROWS when present

if command is test-slice:
    selected = explicit rows, otherwise all active owner rows

if command is service-backed-test-slice:
    if explicit:
        reject any row not requiring managed services
        selected = explicit rows
    otherwise:
        selected = all active owner rows requiring managed services

reject empty selected set
resolve selectors and profiles
reject unresolved, multiply resolved, or overlapping cases
sort rows by row_id
emit immutable plan
execute with stop_on_first_failure=false
drain running work
do not start dependency-blocked work
always run finalizers
emit one terminal record per selected row
```
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-065

**TH-HARNESS-REQ-368**
Runner results MUST map exact executed cases back to one catalog row by the preflight-resolved selector inventory. An executed case owned by zero or multiple rows is `scheduler_accounting_error`. A runner-wide pass cannot close a row whose exact cases are absent.
Verified by: TH-HARNESS-AC-063, TH-HARNESS-AC-065

**TH-HARNESS-REQ-369**
The scheduler and evidence-accounting layers MUST preserve owner, family, verification, selector, profile, and evidence-class identity through work-unit expansion, batching, sharding, cancellation, finalization, and summaries. Batching MAY reduce process count but MUST NOT merge or weaken row identity.
Verified by: TH-HARNESS-AC-060, TH-HARNESS-AC-061, TH-HARNESS-AC-065

**TH-HARNESS-REQ-352**
Check-scheduler dependency declarations MUST account for readiness work once, at the scheduler layer. A scheduled work unit whose actual child path or retained-summary emission requires pnpm-managed workspace Node packages, including root Node tool dependencies, harness schema-validation dependencies, or frontend dependencies, MUST depend on `check-frontend-install`; a scheduled work unit whose actual child path requires a build artifact or service image MUST depend on the scheduler-modeled readiness unit that produces that artifact or image. A scheduled work unit whose selected behavior and retained-summary emission do not require installed pnpm-managed workspace packages MUST NOT depend directly or indirectly on `FRONTEND_INSTALL_STAMP`. The default fast `check-harness-smoke` child behavior MUST require only the Node runtime and harness source inputs needed by the selected fast smoke checks, though its retained summary emission MAY rely on scheduler-modeled pnpm package readiness.
Verified by: TH-HARNESS-AC-006

If a later profile promotes `frontend-fallow-static` into default `check`, it MUST be a direct scheduler work unit with `check-frontend-install` readiness and the static Node resource shape, not a hidden prerequisite of another frontend target.

The public `check` wrapper MUST NOT run substantial frontend install, build, service-image, or browser readiness work outside scheduler accounting. It MAY perform only minimal runner bootstrap needed to start the scheduler process, including pinned repo-local Node readiness for `run-check-schedule.mjs`, plus fail-fast configuration validation that does not provision dependencies or build artifacts. When scheduler event or summary schema validation depends on pnpm-managed Node packages, the check scheduler MAY defer its own early event validation until the scheduler-visible package readiness unit completes, but it MUST validate deferred records before final summary emission. Frontend install, backend build, migration build, service-image build, service-image warmup, and browser readiness that are required by default `check` work MUST appear as scheduler-visible units in retained scheduler summaries.

### 10.1 Scheduler Manifest Fields

The canonical scheduler input schema is `cartulary.scheduler_manifest.v2`.
`tools/scheduler_manifest.json` is the committed generated scheduler input for
check, sequence, service-backed, and test-slice scheduler families. Family-
specific source forms such as check schedule metadata, service-backed
`work_unit_sources[]`, Go shard expansion, and browser group expansion MAY exist
only as upstream authoring inputs. Scheduler runners MUST NOT accept those
family-specific source forms as runtime scheduler manifests. Historical
Non-current scheduler manifest schema IDs are unsupported for runtime and diagnostic
consumption; current producers emit only v2.

The owner-slice planner output is separately owned by `cartulary.test_slice_plan.v2`. Its required `selection` object contains `selection_mode`, `owner_id`, `dependency_scope`, `completion_scope`, `requested_row_ids`, and `resolved_row_ids`. Each work unit includes the sorted exact `runtime_binary_ids` required by its selected catalog families. The enum axes are orthogonal;
target names or command IDs MUST NOT be overloaded to infer omitted selection
metadata. Requested and resolved arrays MUST be sorted and unique, and every
resolved row MUST map to scheduled work or to an inherently target-wide check
whose selected closure remains attributable to that row.

| Field                              | Type                 | Required | Default                      | Rule                                                        |
| ---------------------------------- | -------------------- | -------: | ---------------------------- | ----------------------------------------------------------- |
| `schema_id`                        | string               |      yes | none                         | Must be `cartulary.scheduler_manifest.v2`.                  |
| `generated`                        | object               |      yes | none                         | Generator and authoring-input provenance.                   |
| `schedules[]`                      | array                |      yes | none                         | Normalized scheduler inputs.                                |
| `schedules[].target`               | string               |      yes | none                         | Public target or scheduler target identity.                 |
| `schedules[].scheduler_kind`       | string               |      yes | none                         | `check`, `sequence`, `service_backed`, or `test_slice`; future families require a later adopted schema/spec revision. |
| `schedules[].capacity_profile`     | string               |      yes | none                         | Registry-backed capacity profile name.                      |
| `schedules[].resource_limits`      | object               |      yes | none                         | Logical resource limits or `auto` policies.                 |
| `schedules[].stop_on_first_failure` | boolean             |      yes | none                         | Check scheduler: `true`; service-backed and test-slice schedulers: `false`. |
| `schedules[].progress_tick_seconds` | integer             |      yes | `30`                         | Must be `5..300`; affects reporting only.                   |
| `schedules[].validate_timing`      | boolean              |      yes | `true`                       | Must be `true` for conformance runs.                        |
| `schedules[].summary_groups`       | array                |       no | `[]`                         | Summary grouping policy for scheduler output.               |
| `schedules[].work_units[]`         | array                |      yes | none                         | Ordered by normalized manifest ordinal.                     |
| `work_units[].id`                  | string               |       no | `target`                     | Unique within schedule when defaulted.                      |
| `work_units[].command`             | object               |      yes | none                         | Structured descriptor resolved by the scheduler runner.     |
| `work_units[].priority`            | integer              |       no | `0`                          | Higher integer wins among ready work.                       |
| `work_units[].weight_ms`           | positive integer     |      yes | none                         | Advisory duration estimate only.                            |
| `work_units[].needs[]`             | string array         |       no | `[]`                         | Completion keys required before start.                      |
| `work_units[].completion_keys[]`   | string array         |       no | `[work_unit.id]`             | Added on success.                                           |
| `work_units[].failure_keys[]`      | string array         |       no | completion keys              | Added on failure.                                           |
| `work_units[].complete_on_failure` | boolean              |       no | `false`                      | Adds completion keys even when command fails.               |
| `work_units[].make_prerequisite_policy` | enum `run` or `skip` | required for `make_target` | none | Controls whether scheduler-owned `make_target` execution runs or suppresses the target's recursive Make prerequisites; omitted policy is invalid. |
| `work_units[].resource_claims`     | object               |      yes | `{}`                         | Logical claims only.                                        |
| `work_units[].env`                 | object               |       no | `{}`                         | Scheduler-owned child environment values; MUST NOT override scheduler-owned harness identity variables. |
| `work_units[].readiness_attribution` | object             |       no | none                         | Scheduler-readable readiness/provisioning attribution metadata. When present it MUST declare `timing_role`, `readiness_class`, `warm_threshold_ms`, and `reason`; producers MUST NOT infer readiness attribution from work-unit names. |
| `work_units[].browser_session_group` | string             |       no | stage target                  | Browser stack/session identity shared by compatible browser work units. |
| `work_units[].browser_session_isolation_reason` | string |       no | none                          | Required explanation when authored browser topology deliberately separates otherwise compatible work. |
| `work_units[].browser_session_finalizer` | boolean        |       no | `true`                        | Whether a browser stage completion unit stops its session. Shared projection sessions MUST use a separate `browser_session_finalizer` work unit instead of coupling one target's summary to all groups. |
| `work_units[].shard_names[]`     | string array         | required for `go_shard_finalize` | none               | Selected Go shard names that the aggregate finalizer is allowed to require. For `go_shard_finalize`, this list MUST be non-empty and MUST match the finalizer's `go_shard:<name>` needs after prefix removal. |
| `work_units[].retained_resource_claims` | object         |       no | `{}`                         | Claims kept after work-unit exit until explicit release.    |
| `work_units[].release_retained_resource_claims` | object |       no | `{}`                         | Retained claims to release after work-unit exit.            |
| `schedules[].finalizers[]`         | array                |      yes | `[]`                         | Always run after scheduler drains or stops.                 |

Supported `work_units[].command.type` values are `make_target`,
`service_session_start`, `browser_stage_session_start`, `browser_group`,
`browser_stage_complete`, `browser_session_finalizer`, `go_shard`, `go_shard_finalize`, and
`service_complete`. Dependency-gated aggregate work, including Go shard
aggregation, MUST remain in `work_units[]` and MUST NOT be modeled as an
unconditional scheduler finalizer.

For every command type, fields not listed for that type are forbidden. Scheduler manifest validation MUST reject unknown command types, missing required fields, forbidden fields, wrong field types, omitted `make_prerequisite_policy` on `make_target` work units, and `make_prerequisite_policy` values outside `run` or `skip` before starting child work.

| `command.type` | Required command fields | Optional fields and defaults | Forbidden command fields | Start and success behavior |
| --- | --- | --- | --- | --- |
| `make_target` | `target` | `service_target` only when joining a service session. | `shard`, `browser_stage`, `group_id` | Starts `make <target>` under scheduler-owned environment. Success requires the target's declared summary and artifact policy. When `make_prerequisite_policy=run`, recursive Make prerequisites run normally; when `skip`, the scheduler injects the prerequisite-skip environment only after declared scheduler dependencies have modeled the required readiness work. |
| `service_session_start` | `service_target` | none | `target`, `shard`, `browser_stage`, `group_id` | Starts the owned service session, emits lease and lifecycle evidence, and retains service-stack claims until service completion or finalizers release them. |
| `browser_stage_session_start` | `service_target`, `browser_stage` | none | `target`, `shard`, `group_id` | Starts the browser session group for an existing service session, releases its transient process claim after startup, and retains declared browser lifecycle claims until the session finalizer releases them. |
| `browser_group` | `service_target`, `browser_stage`, `group_id` | none | `target`, `shard` | Runs the selected browser group. Success emits group evidence and the work unit's completion key. |
| `browser_stage_complete` | `service_target`, `browser_stage` | none | `target`, `shard`, `group_id` | Aggregates completed browser groups and emits the stage target summary. For shared projection sessions it MUST depend only on its own browser groups and MUST NOT stop or release the shared session. |
| `browser_session_finalizer` | `service_target`, `browser_session_group` | none | `target`, `shard`, `group_id`, `browser_stage` | Stops a shared browser session and releases its retained browser-stack and stage-lane claims after every group in the session has finished. |
| `go_shard` | `target`, `shard`, `service_target` | `complete_on_failure=false` unless explicitly declared on the work unit. | `browser_stage`, `group_id` | Runs one Go shard under the service session. Product assertion failures map as product failures; setup and runtime failures map through Section 9. |
| `go_shard_finalize` | `target`, `service_target` | top-level `work_units[].shard_names[]` is required and is not a command field. | `shard`, `browser_stage`, `group_id` | Aggregates summaries for scheduler-selected shards and emits the target summary. Missing or inconsistent evidence for a selected shard is an artifact or scheduler-accounting failure; shards omitted by the scheduler selection MUST NOT be required by this finalizer. |
| `service_complete` | `service_target` | none | `target`, `shard`, `browser_stage`, `group_id` | Terminates the owned service lease after dependent work completes and emits its completion key only after teardown succeeds. End-of-schedule cleanup remains an idempotent failure/interruption fallback and MUST NOT be the ordinary teardown path when later work depends on the service-complete key. |

`weight_ms` is an advisory scheduling estimate. It MUST NOT be treated as a logical resource claim, timeout, benchmark claim, pass/fail threshold, or product performance conformance statement.

Nested child-runner concurrency is not advisory. A work unit whose command launches a worker pool MUST keep its declared resource claims and scheduled child worker budget aligned according to Section 5. Direct public targets may expose different developer-loop defaults only when the scheduler-owned invocation remains deterministic and resource-accounted.

Retained resource claims represent continuing logical capacity pressure after a work unit exits. They MUST NOT be used to preserve historical ownership for resources that no longer constrain future work. A browser stage session MAY retain browser-stack and stage-lane claims while the stage remains live, but it MUST release its generic process claim after startup and MUST NOT retain broad database or object-store capacity solely because the stage used those services during readiness. The browser-stack resource owns persistent backend/frontend lifecycle capacity; the process resource owns transient scheduler child admission.

Every `go_shard` scheduler work unit MUST be executable by its declared `target` through the shared Go shard-plan contract. Scheduler generation MUST fail before writing a manifest when a work unit assigns an authoritative/raw shard to `backend-integration-support`, a support shard to `backend-integration`, or any shard name that the target runner cannot resolve for the same target.

Go catalog rows that carry multiple exact Go symbols MAY declare `scenario_symbols` as a closed mapping from stable scenario IDs to those declared symbols. Each scenario MUST remain a scheduler-visible exact-symbol evidence item carrying the row ID, scenario ID, evidence class, and owner ID. Compatible items MAY share a deterministic service Go shard under TH-HARNESS-REQ-357; otherwise the shard identity MUST include the scenario ID. Harness smoke assertions MUST derive membership and shard identity from the catalog and shared Go shard-plan contract rather than maintaining copies. Scenario-scoped planning MUST NOT drop selected row evidence or rerun every scenario through an unaccounted aggregate. Rows that omit `scenario_symbols` retain the same exact-symbol evidence and compatibility rules.

For the `check` scheduler, a service-backed suite session MUST become eligible as soon as its readiness prerequisites are satisfied. Browser build artifacts such as the harness-profile server and migration binaries MUST be dependencies of only the browser stage sessions that require them, not prerequisites of the shared service-suite readiness unit. A browser stage MUST depend on the `build-server-harness` producer that supplies its declared `server-harness` runtime binary; it MUST NOT substitute the independent deployable `build-server` gate. Backend service-backed shards that use the suite template database MAY depend only on the service-session readiness completion key when they do not require those browser build artifacts.

For the `check` scheduler, frontend install, pinned tool bootstrap, binary builds, and service-image warmup MUST be first-class scheduler work units with completion keys when selected downstream work consumes them. Scheduler-invoked child targets MAY suppress recursive Make prerequisite setup only when `make_prerequisite_policy=skip` and their declared readiness keys are already satisfied. Direct public Make target invocation MUST continue to run its normal prerequisites.

Every Make target used as an authored aggregate-sequence step MUST apply the same prerequisite suppression policy even when the target is an internal helper. Its prerequisites MUST be emitted through a conditional recipe prelude rather than unconditional Make graph edges: direct invocation runs the prerequisites, while a scheduler-owned invocation may skip them only after the sequence DAG has established the corresponding readiness dependency. A sequence MUST NOT repeat prerequisite builds after an owning `check` or `build` step has completed them.

A generated recipe MAY declare `prerequisite_jobs` only when its non-Node
prerequisites are independent artifact producers that are safe to admit through
one recursive Make graph. The value MUST be an integer from `2` through `8`,
MUST NOT exceed the number of non-Node prerequisites, and is supported only for
public recipes or authored sequence-step recipes whose prerequisite output is
centralized. The renderer MUST pass the closed value as `--jobs=<value>` to the
single recursive Make invocation. Pinned Node readiness remains a separately
serialized precondition and is excluded from this count. Recipes without the
field remain serial. Ambient `MAKEFLAGS`, parent environment, or public inputs
MUST NOT expand this target-owned bound, and scheduler-owned prerequisite skip
policy MUST suppress the entire recursive block.

The current `migration-drift` recipe owns two independent readiness producers,
the migrate binary and pinned Goose binary, and therefore declares
`prerequisite_jobs=2`. The current `deployable-shape` recipe owns the server,
migrate, and operator binary artifacts through one shared recursive graph and
declares `prerequisite_jobs=3`; that shared graph preserves Make dependency
deduplication for their embedded-web inputs. The standup smoke targets continue
to depend on `deployable-shape` rather than repeating its artifact inventory.

After input validation and shared Postgres readiness, `migration-drift` MUST run
its empty-database apply-to-head scenario and penultimate-boundary upgrade
scenario concurrently. Each scenario MUST own a distinct scratch database and
working directory. The parent MUST await both scenarios, retain both results,
and replay their captured standard output and standard error in authored order:
empty database first, then penultimate boundary. If both fail, the empty
scenario is the primary failure; a penultimate-only failure retains its own
status. Cleanup MUST drain live scenario children and drop both scratch
databases on pass, failure, or interruption. Scenario parallelism MUST NOT
change migration selection, lineage verification, built-binary use, or the
single-migration best-available-boundary fallback.

Prerequisites that are the step's owned child work are not readiness and MUST
NOT be skipped. In particular, `lint-go` runs its format, vet, and staticcheck
children through its prerequisite prelude; the lint preflight establishes
tool readiness but does not replace those child results.

An aggregate sequence invoked by a scheduler work unit MUST treat its authored
per-step prerequisite policy as authoritative. A normal serial or parallel sequence
step MUST clear inherited scheduler prerequisite-skip state before invoking its Make
target. Only a sequence step explicitly declared with `skip_prerequisites=true` MAY
inject prerequisite-skip and prerequisite-satisfied state. An outer composite gate's
`make_prerequisite_policy=skip` MUST NOT suppress prerequisites owned by ordinary
inner sequence steps.

A scheduled work unit that inspects, embeds, serves, signs, packages, or otherwise consumes a generated build artifact MUST depend on the scheduler-visible work unit that produces or proves that artifact. This includes static/security conformance scanners that inspect built bundle roots. A build-artifact cache hit MAY satisfy the producer work unit only under the Section 8 cache contract; it MUST NOT allow an artifact consumer to start without the producer completion key. Cache-profile smoke fixtures MUST create every declared file and directory output for the profile they exercise; partial-output fixtures MUST model a rebuild or fail-closed path and MUST NOT accept success by incomplete output.

The current check schedule MUST model embedded web asset preparation as first-class build-artifact readiness. The `embedded-web-assets` producer work unit MUST depend on `build-web`; both `build-server` and `build-server-harness` MUST depend on `embedded-web-assets`; and scheduled `backend-unit` MUST depend on `embedded-web-assets` while its selected Go package graph can compile `internal/platform/httpapi/webassets`. Direct public target invocation MAY keep ordinary Make prerequisites and fallback embedded assets, but scheduler-selected work MUST use producer completion keys instead of ambient source-tree stability.

Readiness work that materially affects timing MUST NOT be hidden behind another scheduler unit's Make prerequisites. The current profile MUST model `testservices-build` separately from `test-service-images`; future frontend or web asset builds that materially contribute to a downstream readiness unit MUST receive the same first-class treatment. Warm eligibility checks MAY reject retained timing evidence when readiness units show provisioning-heavy work above their warm thresholds.

Runtime binary injection is scheduler-owned. The current runtime-binary registry is closed and data-driven. Each entry MUST declare the runtime binary ID, producer target, producer output variable, consumer environment variable, and repo-relative default output path used by scheduler-owned consumer env wiring. The current registry contains exactly four entries:

| ID | Producer target | Producer output variable | Consumer env | Default output path |
| --- | --- | --- | --- | --- |
| `server` | `build-server` | `SERVER_BIN` | `CARTULARY_SERVER_BIN` | `server` |
| `server-harness` | `build-server-harness` | `SERVER_HARNESS_BIN` | `CARTULARY_SERVER_HARNESS_BIN` | `server-harness` |
| `migrate` | `build-migrate` | `MIGRATE_BIN` | `CARTULARY_MIGRATE_BIN` | `migrate` |
| `operator` | `build-operator` | `OPERATOR_BIN` | `CARTULARY_OPERATOR_BIN` | `operator` |

Catalog rows MAY declare one or more runtime-binary registry IDs through their referenced runtime profile. A scheduler-selected Go shard or process unit that includes such a row MUST depend on every declared registry producer before starting the consumer process and MUST pass only the declared registry consumer environments with the registry default output paths unless a later adopted registry rule declares another scheduler-owned source. Producer inputs MUST NOT be forwarded to arbitrary child processes. A process test that starts the server MUST resolve `server-harness`; browser stacks consume the declared `server-harness` and `migrate` artifacts through their Make-owned lifecycle adapter.

Declaring or consuming a runtime binary is not itself a resource-isolation claim. Scheduler-expanded Go shards MUST claim only their real logical contention, such as one `process` slot plus database, object-store, CPU, and I/O fixture claims when applicable. A runtime-binary consumer MUST NOT claim the full `process` capacity merely because it consumes `operator`; any future full-lane isolation requirement needs an adopted owner rule that names the isolation reason or a dedicated resource.

Omitted producer variables resolve to the registry’s Make defaults. Omitted consumer environments are invalid for canonical process or browser work. Raw direct Go tests MUST fail with a concise public-Make instruction; retained harness evidence MUST NOT rely on hidden nested builds. Public `check` and unrelated targets MUST reject command-line producer or consumer overrides before child work; inherited undeclared values MUST be stripped unless the selected runtime-binary registry entry reintroduces the scheduler-owned consumer environment.

The runtime consumer MUST verify that its declared consumer environment is non-empty, normalized as a filesystem path, an existing regular executable file, and not a symlink path. Missing, non-regular, non-executable, or caller-supplied public-command-line values are `configuration_error`, exit `2`, before product assertions. A mismatch between the consumed binary and the producer build-artifact reference or output digest is `artifact_error`, exit `11`.
Verified by: TH-HARNESS-AC-024, TH-HARNESS-AC-037

**TH-HARNESS-REQ-398**
Default `check-service-backed` browser work MUST consume the same generated
`browser_session_group`, runtime-profile, and service-requirement identities as
the selected browser batch plan. The check scheduler MUST NOT create
check-specific session aliases or merge sessions merely because their current
requirements appear compatible. Browser work shares a stack only when the
authored browser topology resolves it to the same session group and runtime
profile. A shared browser session MUST preserve per-target summaries,
per-target completion keys, cleanup on failure, and redaction-safe session
artifacts. Reset boundaries MUST be explicit scheduler work or an explicit
isolation reason; target-name conventions MUST NOT be used as the sharing
contract.

The service-backed `webserver-backed` stage owns exactly two session lanes so
its default and Network Flow-claimed sessions may overlap. They remain separate
stacks with distinct runtime profiles, fixture identities, ports, and cleanup;
the Network Flow session retains its immutable isolation reason. A session lane
MUST remain held until all groups attached to that session are terminal and its
session finalizer completes.

Service-backed browser stage dependencies generated by the schedule renderer MUST be owned by `tools/execution_topology_manifest.json`, not by hard-coded renderer or validator lists. The current ordinary measurement policy is `service_backed_schedules.defaults.browser_stage_generated_needs.measurement.selected_peer_stages=["webserver-backed","stateful","visual","a11y"]`; the renderer and topology validator MUST consume that same owner input and fail closed when a selected peer browser stage is not explicitly listed with an owner reason.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-018

**TH-HARNESS-REQ-399**
Every scheduled `browser_group` that runs inside one service-backed scheduler runtime MUST receive a scheduler-owned worker-admin slot range. The ranges for all concurrently schedulable browser groups in that service-backed runtime MUST be non-overlapping and contiguous from offset `0`, and every group MUST receive `CARTULARY_PLAYWRIGHT_WORKER_COUNT` equal to the total slot count plus `CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET` equal to the start of its range. A group that launches more than one Playwright worker MUST own a range with at least that many slots. Direct public browser leaf targets MAY keep the direct isolated offset default from Section 5, but scheduled browser groups MUST fail before Playwright product assertions when the count, offset, or range is missing or invalid.

Every scheduled non-measurement `browser_group` MUST claim exactly two
scheduler-family CPU tokens, one scheduler-family I/O token, and one `process`
slot. A measurement group retains its limit-sized CPU and I/O isolation and
also claims exactly one `process` slot. A retained browser stage session
releases its startup process slot once the backend/frontend stack is ready;
`browser_stack` remains its exact live-stack capacity claim.
A browser group process or CPU claim MUST NOT be dropped merely because the
group attaches to an existing browser session; the Playwright/Chromium child
is independent host pressure and participates in ordinary resource admission.

Every retained `browser_stage_session` MUST retain exactly one `browser_stack`
slot and MUST NOT retain a generic `process` slot, even when the authored stage
startup uses `limit` claims to establish an isolation boundary. Limit-sized
startup claims MUST release down to the exact browser-stack and stage-lane
lifecycle claims before a dependent browser group is admitted. Measurement
exclusivity belongs to the measurement group's full CPU, I/O, Postgres, and
object-store claims; a retained session MUST NOT retain its dependent child's
process capacity or make that child infeasible.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-018

### 10.2 Logical Resource Registry

The current-conformance logical resource registry is closed by TH-HARNESS-REQ-353 below. Earlier summary tables MUST NOT be used as resource-bound authority.

Scheduler resource claims MUST distinguish reset-heavy, clone-heavy, transaction-heavy, fixture-start-heavy, browser, build, and static work so one expensive class cannot monopolize broadly shared I/O capacity. In the default check profile, reset-heavy service-backed shards MUST be capped below total `host_io`/`go_io` capacity while `postgres_reset` remains a separate bottleneck. Owner topology inputs MAY declare per-execution-family claims only through registered resource profiles. A fixture-sensitive profile MUST use the narrowest registered resource that models its contention and MUST NOT inflate broad `host_io`, `go_io`, or `process` claims when a closed fixture resource exists. A caller override MAY reduce total capacity only when the normalized schedule still has a feasible non-deadlocking assignment.

**TH-HARNESS-REQ-353**
The logical resource registry is closed by the table below. `tools/scheduler_resource_registry.json` is a mirror of this current-conformance registry and MUST NOT independently add resources, change bounds, change override inputs, or redefine auto policies.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-030

| Resource | Schedulers | Default limit | Auto policy | Override input | Min | Max | Display/order rule | Omission behavior |
| --- | --- | --- | --- | --- | ---: | ---: | --- | --- |
| `host_cpu` | `check`, `sequence` | none | `host_cpu` | `CHECK_HOST_CPU_JOBS` for `check` only | 1 | 256 | display order `10` | resolve by scheduler-specific auto policy |
| `host_io` | `check`, `sequence` | none | `host_io` | `CHECK_HOST_IO_JOBS` for `check` only | 1 | 256 | display order `20` | resolve by scheduler-specific auto policy |
| `suite_service_stack` | `check` | `1` | none | none | 1 | 256 | display order `30` | use default `1` |
| `migration_scratch_postgres` | `check` | `1` | none | none | 1 | 256 | display order `40` | use default `1` |
| `go_cpu` | `service_backed`, `test_slice` | none | `service_backed_go_cpu` | `CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT` | 1 | 256 | display order `110` | resolve by auto policy |
| `go_io` | `service_backed`, `test_slice` | none | `service_backed_go_io` | `CARTULARY_SERVICE_BACKED_GO_IO_LIMIT` | 1 | 256 | display order `120` | resolve by auto policy |
| `browser_stack` | `check`, `service_backed`, `test_slice` | none | `service_backed_browser_stack` | `CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT` | 1 | 256 | display order `130` | resolve by auto policy |
| `object_store` | `check`, `service_backed`, `test_slice` | `32` | none | none | 1 | 256 | display order `140` | use default `32` |
| `seaweedfs_fixture` | `check`, `service_backed`, `test_slice` | `2` | none | none | 1 | 8 | display order `145` | use default `2` |
| `postgres` | `check`, `service_backed`, `test_slice` | `32` | none | none | 1 | 256 | display order `150` | use default `32` |
| `process` | `check`, `sequence`, `service_backed`, `test_slice` | none | `host_process_slots` | none | 1 | 256 | display order `160` | resolve by scheduler-specific auto policy |
| `postgres_reset` | `check`, `service_backed`, `test_slice` | none | `service_backed_postgres_reset` | `CARTULARY_SERVICE_BACKED_POSTGRES_RESET_LIMIT` | 1 | 8 | display order `170` | resolve by auto policy |
| `postgres_clone` | `check`, `service_backed`, `test_slice` | none | `service_backed_postgres_clone` | `CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT` | 1 | 8 | display order `175` | resolve by auto policy |
| `browser_stage_*` | `check`, `service_backed`, `test_slice` | `1` | none | manifest positive integer only | 1 | 8 | display order `135`, then resource name lexical order | use generated default `1` unless manifest declares another positive value; current schedules declare `browser_stage_stateful=3` and `browser_stage_webserver_backed=2` |

Resource override inputs accept positive decimal integers only. An override below the largest declared claim for that resource, above the resource maximum, or incompatible with a feasible non-deadlocking schedule is a `configuration_error`, exit `2`, before child work.

Current capacity profiles are `check_default` for the `check` scheduler, `sequence_adaptive` for the `sequence` scheduler, `service_backed_full` and `service_backed_backend` for the `service_backed` scheduler, and `test_slice_default` for the `test_slice` scheduler. Capacity profiles are registry-owned shortcuts only; they MUST NOT authorize resources outside the scheduler family named by the profile. Owner-input validation MUST reject unknown profile names, profile names attached to the wrong scheduler family, and auto-policy names outside the current closed set before schedule execution begins.

Go shard scheduler-profile resource claims are a closed shared policy. For the `check` scheduler, CPU and I/O claims use `host_cpu` and `host_io`; for `service_backed` and `test_slice`, they use `go_cpu` and `go_io`. Profile claims are: default or `balanced` = CPU `1`, I/O `1`; `cpu_heavy` = CPU `2`, I/O `1`; `io_heavy` = CPU `1`, I/O `2`; `transaction_heavy` = CPU `1`, I/O `1`; `reset_heavy` = CPU `1`, I/O `2`, `postgres_reset` `1`; and `clone_heavy` = CPU `1`, I/O `2`, `postgres_clone` `1`.

**TH-HARNESS-REQ-354**
Auto resource policies are closed by the following algorithms. `available_parallelism` is `os.availableParallelism()` when the runtime exposes it and otherwise the host CPU count; in either case it is floored at `1`. `clamp(value, min, max)` returns `min(max(value, min), max)`.

| Auto policy | Required algorithm |
| --- | --- |
| `host_cpu` | For `check`, `clamp(ceil(available_parallelism * 0.7), 1, 256)`. For `sequence`, `max(largest declared host_cpu claim, floor(available_parallelism * 0.85))`, bounded to the registry limits. |
| `host_io` | For `check`, `max(host_cpu, largest declared host_io claim in the normalized provisional work-unit set)`. For `sequence`, `max(host_cpu, available_parallelism, largest declared host_io claim)`, bounded to the registry limits. |
| `host_process_slots` | For `sequence`, `max(largest declared process claim, clamp(floor(available_parallelism / 3), 2, 8))`, bounded to the registry limits. For `check`, `service_backed`, and `test_slice`, return `max(largest declared process claim, clamp(floor(available_parallelism / 2), 2, 12))`, bounded to the registry limits. |
| `service_backed_go_cpu` | If no Go shard units exist, return `1`. Otherwise compute `total_weight=sum(max(1, weight_ms))`, `max_weight=max(max(1, weight_ms))`, `weighted_concurrency=ceil(total_weight / max(30000, max_weight))`, `host_concurrency=max(2, available_parallelism - 1)` when `available_parallelism <= 4` and `floor(available_parallelism * 0.75)` otherwise; return `clamp(max(4, min(host_concurrency, weighted_concurrency)), 4, 16)`. |
| `service_backed_go_io` | If no Go shard units exist, return `1`. Otherwise count Go shard scheduler profiles and compute `profile_concurrency=balanced + transaction_heavy + 2*io_heavy + 2*clone_heavy + 2*reset_heavy + ceil(cpu_heavy/2)`; return `clamp(max(6, go_cpu + 2, profile_concurrency), 6, 24)`. |
| `service_backed_browser_stack` | Count distinct `browser_stage_session` work units that claim `browser_stack`, keyed by `browser_session_group`; if no retained session starters exist, count distinct `browser_stage_*` resource lanes in the normalized provisional work-unit set. If the resulting demand count is `0`, return `1`. Otherwise return `max(1, min(demand_count, stack_claiming_unit_count when nonzero, process limit when set, max selected CPU limit when set))`, where the selected CPU limit is `host_cpu` for `check` and `go_cpu` for `service_backed` or `test_slice`. |
| `service_backed_postgres_clone` | Let CPU and I/O bounds be the selected scheduler CPU and I/O resource limits. If neither bound is positive, return default `6`. Otherwise return `max(1, min(8, max(6, floor(min(positive CPU/I/O bounds)/2))))`. |
| `service_backed_postgres_reset` | Let the I/O bound be `host_io` for `check` and `go_io` for `service_backed` or `test_slice`. If the bound is absent or non-positive, return default `4`. Otherwise return `max(1, min(8, floor(io_bound/3)))`. |

Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-018, TH-HARNESS-AC-030

**TH-HARNESS-REQ-355**
Harness performance acceptance MUST measure the unchanged public `make check`
command on one documented host and scheduler-capacity profile. A qualifying
window consists of one successful unmeasured warm run followed by five
consecutive successful measured runs on byte-identical authored inputs and
generated outputs. External process wall time is the controlling measurement;
the nearest-rank p90 of five runs, which is the maximum, MUST be strictly less
than 120 seconds. Every measured run MUST retain the same public target and
`command_id` inventory, the same required catalog-row inventory and evidence
routes, zero missing or unmapped required evidence, deterministic scheduler
timing, valid artifacts, and successful cleanup.

Failed, interrupted, stale, retry-contaminated, input-mismatched, capacity-
mismatched, or otherwise contaminated runs MUST be retained and classified but
MUST NOT enter the qualifying window. Any authored-input or generated-output
change restarts the warm run and all five measured runs. Performance evidence is
harness evidence only; it MUST NOT be cited as product conformance, release
readiness, benchmark publication, or Core 05 claim-publication evidence.
Verified by: TH-HARNESS-AC-059

**TH-HARNESS-REQ-356**
An unsharded pure Go target MAY execute multiple exact-symbol evidence rows in
one execution family only when every row has compatible package selection,
runtime-binary requirements, fixture policy and budget, and isolation policy.
The owner catalog remains the evidence-routing owner. Consolidation MUST preserve
every row ID, verification ID, evidence class, selected symbol, profile, and
default-check disposition, and the Go JSON report MUST still prove
each exact selected symbol. Raw package-wide selectors remain separate from
exact-symbol families. A missing, ambiguous, duplicate, or unexpectedly
selected symbol MUST fail closed; it MUST NOT be repaired by inferred row
membership or treated as derived success.

Private execution-family IDs MAY be replaced atomically when no continuing
external consumer is demonstrated. Current producers, consumers, diagnostics,
fixtures, and duration accounting MUST move together; no alias, dual reader,
dual emitter, or forwarding shim is permitted for the retired private IDs.
Verified by: TH-HARNESS-AC-060

**TH-HARNESS-REQ-357**
Service-backed exact-symbol Go evidence items MAY share one scheduler shard only
when they have the same target, package selection, runtime-binary set, fixture
policy and complete fixture budget, isolation policy, scheduler resource
profile, and authoritative or support evidence class. Raw package selectors,
isolated items, and incompatible items remain separate. Within each compatible
execution family the planner MUST sort by descending authored-input duration
estimate with stable row ID and symbol tie-breakers, then first-fit pack using
the target's closed exact-symbol shard profile. The default profile admits at
most eight exact symbols and at most 12,000 milliseconds of estimated test work
per shard. `backend-process` uses at most 16 exact symbols and at most 24,000
milliseconds because its compatible process-lifecycle symbols share expensive
Go package startup and contend for the same service fixture when split into too
many physical children. An individual item whose owner estimate exceeds its
target profile's estimated-work limit MUST remain alone and MUST NOT be split
or hidden. Package and command startup overhead remains part of the shard's
scheduler weight but MUST NOT be charged once per item when deciding whether
compatible work shares a process.

Every packed item MUST retain its row ID, scenario ID when declared, symbol,
owner, coverage class, fixture proof, duration key, and runtime-binary
requirements in the shard plan and emitted evidence. The runner MUST use an
anchored exact-symbol selector and MUST fail closed on unexpected, duplicate,
missing, crashed, or incomplete output; unresolved rows MUST remain missing.
Packing and generated shard names MUST be byte-deterministic from authored
inputs and duration owners. Private family and shard IDs MAY be replaced
atomically without aliases or dual readers, but public targets, command IDs,
result-root structure, summaries, failure classes, and cleanup behavior MUST
not change.
Verified by: TH-HARNESS-AC-061

**TH-HARNESS-REQ-358**
Scheduler-manifest growth gates MUST remain meaningful when an optimization
reduces the number of ordinary work units. The report MUST retain total bytes,
bytes per work unit, overall p95, and maximum serialized-unit measurements, but
MUST gate p95 independently for ordinary units and for the structurally wide
`aggregate_finalize`, `browser_group`, and `browser_stage_complete` kinds.
Ordinary p95 MUST be at most 1,500 bytes; structurally-wide p95 MUST be at most
5,000 bytes; the overall maximum MUST remain at most 12 KiB; total bytes per
unit and the 25-unit ordinary synthetic fixture MUST each remain at most 1,600
bytes. Scratch renders MUST be byte-identical. Adding or removing large numbers
of compact Go shards MUST NOT by itself conceal or create a growth failure by
moving a different work-unit kind across one global percentile boundary.
Verified by: TH-HARNESS-AC-061

**TH-HARNESS-REQ-359**
A migration-scratch helper attached to a scheduler-owned service suite MUST
create a fresh empty database and replay the selected migration path exactly as
before, but its test cleanup MUST close every client and record the scratch
database as suite-retained. The scheduler-owned service teardown MUST remain
responsible for removing the owned stack and all retained scratch state. A
standalone helper without an active attached suite MUST still drop its scratch
database during test cleanup. This lifecycle consolidation MUST preserve fresh
database identity, fixture attribution and budgets, migration and failure
evidence, deterministic cleanup, and fail-closed teardown; it MUST NOT reuse a
scratch database across tests or runs, leave state outside the owned stack, or
treat a retained database as completed cleanup before service teardown passes.
Verified by: TH-HARNESS-AC-061

**TH-HARNESS-REQ-370**
Every scheduler kind, including `sequence`, MUST emit
`cartulary.scheduler_event.v7` evidence that represents each work unit with its
manifest ordinal, priority, declared dependency edges, normalized resource
claims, eligibility instant, start instant, and terminal instant and state.
Eligibility is the first scheduler iteration in which all declared dependencies
are complete or the unit becomes terminally dependency-blocked. Queue wait is
elapsed time from eligibility to start. A unit that never starts records its
terminal dependency or cancellation state without a fabricated execution
duration. All boundaries use the scheduler process's monotonic clock; child
processes MUST NOT synthesize monotonic values from wall-clock timestamps.
Verified by: TH-HARNESS-AC-074, TH-HARNESS-AC-077

**TH-HARNESS-REQ-371**
For every eligible but unstarted interval, the scheduler MUST retain the closed
wait reason `resources`, `earlier_overlapping_ready`, `capacity`, or
`scheduler_stop`, plus sorted blocking logical resources and blocking unit IDs.
The event stream MUST retain wait-start and wait-end boundaries even when the
interval is zero. Resource-blocked duration is the union of intervals
attributable to one logical resource; overlapping observations MUST NOT be
summed.
Verified by: TH-HARNESS-AC-074, TH-HARNESS-AC-077

**TH-HARNESS-REQ-372**
The actual dependency critical path is computed over observed dependency edges
using each node's queue-wait and execution intervals. For each node, path cost
is its attributable duration plus the maximum predecessor path cost, with
manifest ordinal and stable ID tie-breakers. The hotspot summary MUST name the
ordered path as `critical_path` and its duration as
`actual_dependency_critical_path_ms`. It MUST NOT rewrite
or reinterpret scheduler `critical_path_wall_duration_ms`, which remains the
full scheduler envelope under TH-HARNESS-REQ-303.
Verified by: TH-HARNESS-AC-074

**TH-HARNESS-REQ-373**
Unattributed envelope time is parent duration minus the union of directly
attributed child intervals clipped to the parent interval. Negative results
clamp to zero and set a bounded clock/accounting warning. Child durations MUST
NOT be summed when they overlap. Sequence-step, scheduler-work, service,
runner, report-collation, and finalizer intervals participate only under their
direct parent.
Verified by: TH-HARNESS-AC-074

**TH-HARNESS-REQ-374**
Task-surface sequences that declare dependency and resource topology MUST run
through the existing scheduler engine under `scheduler_kind=sequence`. Serial
sequences compile to the same engine as one-edge-per-step DAGs. Public command
and summary metadata remains owned by the task-surface owner; dependency,
resource claim, priority, capacity, and finalizer metadata remains owned by
execution topology. Validation MUST require an exact one-to-one binding before
execution and MUST reject duplicate target steps, cycles, unknown dependencies,
unknown resources, and infeasible claims. `check` is the first dependency of `ci`
and `release-check`; post-check work may overlap only as declared. Cancellation,
process-group interruption, first-failure selection by observed completion,
running-sibling drain, dependency skips, finalizers, and public summary order
remain governed by Sections 9 and 10. A shell launcher, when retained, MUST only
validate argv and launch this engine; it MUST NOT implement scheduling policy.

The `sequence_adaptive` profile owns `host_cpu`, `host_io`, and `process`
capacity and MUST ignore inherited `CHECK_HOST_CPU_JOBS` and
`CHECK_HOST_IO_JOBS`; those variables remain direct inputs to the nested
`check` scheduler only. Every sequence work profile MUST claim one `process`
slot. The closed sequence work profiles are: `small_check` = CPU/I/O `1/1`,
`script` = `2/2`, `cpu_analysis` = `4/1`, `artifact_generation` = `2/4`,
`build` = `6/3`, `service_validation` = `2/4`, and the parallel gosec-only
`security_analysis` profile = full resolved CPU capacity and I/O `1`.
`nested_browser_validation` = CPU/I/O `4/4`. Each also claims process `1`.
Parallel Make jobs are `1` for small and service work and equal to the profile
CPU claim for the other profiles.
`nested_check` claims the entire
resolved CPU and I/O budget and forwards those exact values as
`CHECK_HOST_CPU_JOBS` and `CHECK_HOST_IO_JOBS`. `nested_service_validation`
claims CPU/I/O `2/4` and forwards those exact values as
`CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT` and
`CARTULARY_SERVICE_BACKED_GO_IO_LIMIT`. Forwarding is registry-owned,
observable in the scheduler summary, and MUST reject an unclaimed source,
unknown target resource, duplicate child environment variable, or mapping to
an environment input other than the target resource's registered capacity
input.
`nested_browser_validation` claims CPU/I/O `4/4` and uses the same
registry-owned service-backed forwarding. Its four CPU tokens admit two
non-measurement browser groups concurrently while each group retains the exact
CPU-two claim.

Source-boundary conformance commands MUST index each repository tree once per
invocation and reuse immutable file text across independent checks. Rewalking
or rereading the same authored tree for each predicate is forbidden when a
single filtered index provides identical fail-closed coverage. Repo-local tool
and package caches are excluded from authored-source traversal by explicit
directory identity; their content cannot satisfy or violate source ownership.
In the default `check` schedule, the OTel source-boundary work unit MUST wait
for `check-service-backed` completion before admission. It retains the ordinary
CPU/I/O `1/1` claim and static-validation priority; this phase edge prevents a
subsecond repository scan from contending with the service-backed critical
path without granting it exclusive capacity or artificial scheduling priority.
The current aggregate DAGs are closed as follows. `lint` MUST establish Node,
frontend-install, and shell-tool readiness once before starting the sequence;
its eight lint, boundary, script, Markdown, shell, and frontend-typecheck steps
then have no dependency edges. `ci` MUST start only `check`; after it succeeds,
`harness-contract`, `deployable-shape`, and `duration-baseline-drift-suite`
and the bounded `go-gosec-audit` are mutually independent. Both scans and the CI `deployable-shape` step MUST
suppress recursive prerequisites because successful `check` has established
their toolchain, frontend, and binary readiness. `release-check` MUST
start only `check`; after it succeeds, harness contract, security audit,
license generation, build, SeaweedFS compatibility, and
release-browser readiness are independent as capacity permits. Both scans
reuse the readiness established by `check`. `sbom` depends
on `license-report`; `seaweedfs-release-gate` depends on compatibility,
license, and SBOM; `deployable-shape` depends on build; and
the release `deployable-shape` step MUST suppress the binary prerequisites
established by that build dependency. `release-readiness-evidence` depends on every release branch. Authored summary
order remains stable and does not follow completion order.

The advisory gosec audit owns one physical `repository` scan. Its rules are the
stable union of the former runtime and support rules, and its package selection
is the minimal union of runtime patterns and inventory-derived support patterns;
a broad runtime pattern subsumes narrower support descendants. The consolidated
scan MUST preserve every prior check and MAY add the stricter runtime-only rule
to support code because the audit remains non-failing. Under the
`security_analysis` sequence profile, the scan receives the scheduler's exact
resolved CPU claim as `GOMAXPROCS`; the full-capacity claim prevents unrelated
CPU work from oversubscribing it. The scheduler MUST strip an inherited value
and inject the claim for every sequence child. Direct execution derives the
same limit from online host capacity. Output and profile metadata identify the
single repository scan. The audit reads and validates its support inventory
through a contract check rather than runtime discovery. The authored broad
package roots MUST subsume every support-inventory root; adding an uncovered
root fails the harness contract until the audit plan is extended.
Verified by: TH-HARNESS-AC-077, TH-HARNESS-AC-078

**TH-HARNESS-REQ-375**
The execution-topology row for `release-browser-readiness` itself owns logical
`browser_stack` capacity two; no parent target or environment override may alter
that schedule identity. The reference profile uses the same isolated schedule
with capacity one, and the candidate profile changes only the retained
execution-policy projection to capacity two plus a parent-forwarded CPU/I/O
budget of `4/4`. The parent budget MUST use `nested_browser_validation`; a
smaller forwarded CPU budget that serializes CPU-two browser groups is not the
adopted transition. The schedule produces five
profile-compatible sessions and the existing support, visual, accessibility,
and aggregate summaries. Direct public browser leaves retain isolated stack
behavior. A session beyond the compatible default schedule requires an
immutable `browser_session_isolation_reason`; resets, runtime profiles, visual
comparison mode, redaction, and cleanup are unchanged.
Verified by: TH-HARNESS-AC-077, TH-HARNESS-AC-078

### 10.5 Harness Observability Metrics and Performance Acceptance

**TH-HARNESS-REQ-376**
The current harness metric registry is closed to
`cartulary.harness.invocation.duration`,
`cartulary.harness.dependency.critical_path`,
`cartulary.harness.scheduler.queue_wait`,
`cartulary.harness.scheduler.resource_blocking`, and
`cartulary.harness.invocation.unattributed`. Duration units are milliseconds.
Allowed dimensions are the closed safe
target, command, family, runner, work-unit kind, status, timing bucket, wait
reason, and logical-resource tokens present in the native bundle. Run IDs,
paths, process IDs, test output, and raw symbol names are forbidden metric
dimensions.
Verified by: TH-HARNESS-AC-075, TH-HARNESS-AC-076

**TH-HARNESS-REQ-377**
A qualifying target/provider performance window contains one consecutive
successful warm-up observation followed by exactly two consecutive successful
measured observations. The warm-up root is retained and validated but excluded
from statistics. For two values, the median is their arithmetic midpoint and
the median absolute deviation is the arithmetic midpoint of their absolute
deviations from that median. All three roots in one window MUST match provider,
command ID, canonical inputs, timing source, workload/evidence digest, host
profile, externally available capacity, toolchain profile, target execution
policy, commit, source-snapshot digest, and clean source state. Different
reference targets MAY bind immutable windows from different clean snapshots;
all final candidate windows MUST share one clean frozen commit and source
snapshot. Across one target's reference and candidate windows, every retained
field above except commit and source snapshot MUST match unless the target's
contract declares one exact normalized policy transition. Backend-unit and its
synthetic finalizer permit only grouped capture plus parallel report emission;
`lint`, `ci`, and `release-check` permit only serial-reference to
topology-owned-DAG policy; release browser readiness permits only child-owned
browser-stack capacity one to two; `browser-e2e-webserver-backed` permits only
its two isolated runtime-profile sessions to change from one shared stage lane
to two lanes. Comparing only an opaque changed digest is
insufficient: current evidence MUST retain and compare the normalized policy
projection. Let `m` be the median duration and `d` the median absolute
deviation. The no-regression limit is `m + max(1000, 3*d, 0.05*m)`
milliseconds. A required hotspot improves only when its candidate median is at
least `max(1000, 3*d, 0.10*m)` milliseconds below the reference median.
The normalized execution-policy projection MUST use the Section 3.6 semantic
JSON encoding. Its `execution_policy_sha256` value is the same SHA-256 digest
without the `sha256:` prefix. Producers and validators MUST use that one
recursively key-sorted, I-JSON-safe encoding; formatted bytes, object insertion
order, and a second policy-specific canonicalizer are forbidden digest inputs.
Failed, interrupted, stale, retried, source- or capacity-mismatched observations
are retained as rejected evidence and do not enter either set. Performance
acceptance MUST consume retained execution contexts, record every rejected root
and reason, and MUST NOT derive a historical profile from the current checkout
or trust manifest-declared digests over the root's retained context.
Verified by: TH-HARNESS-AC-079

**TH-HARNESS-REQ-378**
Required improvement gates are backend-unit total wall time, backend native
`report_collation` interval union, release browser-readiness wall time, and
release-check wall time. Every other required public testing command uses the
no-regression limit. The three exact timing sources are the complete public
invocation envelope, an exact-once aggregate scheduler-work envelope, and the
interval union of native backend-unit `report_collation` spans. Aggregate runs
MAY provide leaf samples only when scheduler evidence proves the same command,
canonical inputs, workload, capacity contract, and one exact occurrence;
otherwise the command MUST be run directly. A v2 evidence-roots manifest
deduplicates explicit reference or candidate windows and binds each target to
one window and timing source. `mode=baseline` contains only reference windows
and bindings; `mode=comparison` also contains candidate windows and bindings.
All 48 required public targets MUST pass their individual gate, and
`public_entrypoint_portfolio_total_ms`, the sum of those 48 candidate medians,
MUST be strictly less than the corresponding reference sum. The internal
release-browser row and synthetic backend-finalizer row are required gates but
are excluded from that 48-target sum. Default `make check` MUST NOT enforce
these two-sample drift gates.
TH-HARNESS-REQ-355 remains the independent controlling five-run `make check`
acceptance.
Verified by: TH-HARNESS-AC-079

**TH-HARNESS-REQ-379**
Compatible pure backend-unit exact-symbol groups MUST be partitioned directly
by normalized package selection, sorted runtime-binary set, complete fixture
profile, fixture policy and budget, isolation policy, and authoritative or
support evidence class, then execute with
`min(group_count,clamp(floor(available_parallelism/4),1,8))` workers. Each Go
child receives scheduler-owned `GOMAXPROCS=max(1,floor(available_parallelism/workers))`.
This child scheduler partition applies only to backend-unit grouped capture.
Other backend targets retain the available host parallelism assigned to their
own scheduler unit; backend finalization MAY reuse the bounded worker-count
formula without rewriting capture-child `GOMAXPROCS`.
All exact symbols in one
compatible package/runtime/fixture/isolation/evidence-class group MAY share one
Go JSON process. Raw package selectors remain separate. Each physical Go report
MUST be parsed once; immutable family-projection requests then execute through
the same host-derived worker pool. Each worker MUST initialize the output
runtime once and process multiple requests sequentially; starting one cold
output runtime per family is forbidden. Target owner-evidence finalization MUST
load one immutable catalog/accounting context and reuse it for every owner
partition in that target. Production MUST NOT read or mutate ambient environment
to control this private worker limit. Stable row ordering,
target-summary ordering, and Section 9 primary-failure ordering MUST be applied
after all bounded work settles. Capture or report-worker failure MUST preserve
any already successful row evidence and select one primary failure through the
existing taxonomy.
Verified by: TH-HARNESS-AC-078

**TH-HARNESS-REQ-380**
`harness-public-target-duration-baselines` is the sole writer of
`tools/harness_public_target_duration_baselines.json`. It MUST accept only an
exact v2 `mode=baseline` manifest with one accepted warm-up and two accepted
measured roots per target/provider window, independently verify each context and
bundle, derive samples from the binding's exact timing source, compute the
Section 10.5 target and portfolio statistics, write deterministic normalized
bytes, and retain a bounded maintenance summary. A non-publication
`retained_v1_reference_migration` path MAY accept an immutable v1 reference root
without the later invocation marker only when its terminal native summaries,
scheduler evidence where applicable, complete artifacts, clean state, canonical
inputs, and timing boundaries all validate. Migration mode is forbidden for
candidate evidence. When a migrated v1 baseline row contains only its historical
execution-policy digest, unchanged-policy comparison MUST recompute that exact
legacy formatted-JSON digest over the strict candidate's retained normalized
projection. The bridge MUST first prove that the baseline wrapper and row digest
agree, MUST remain isolated to `retained_v1_reference_migration`, and MUST NOT
replace the Section 10.5 semantic digest in a v2 context or strict candidate.
Declared policy transitions remain governed by their exact normalized
transition checks and MUST NOT use the digest-only bridge. The writer MUST
reject cold, dirty, failed, interrupted,
retried, duplicate, profile-mismatched, missing-command, and wrong-cardinality
roots. Hand editing, partial refresh, inferred roots, newest-run selection, and
v1 manifests in normal operation are forbidden.
Verified by: TH-HARNESS-AC-079

`work_units[].timeout_seconds` is optional and, when present, MUST be an integer from `1` through `3600`. It is a scheduler-owned watchdog around the whole child process group, not a product-performance assertion. Expiry MUST terminate the child group, retain the partial redacted log, record `failure_class=timing` and `failure_reason=timeout_failure`, return `13`, drain already-running independent work, mark dependency-blocked work `skipped_dependency`, and then run finalizers. A finalizer has its own timeout and MUST NOT inherit the aborted work signal. Omitting the field delegates deadlines to the narrower service, browser, runner, or child-target contract. Product assertions MUST have exactly one scheduler attempt; scheduler watchdog expiry MUST NOT create a retry.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-030, TH-HARNESS-AC-065

### 10.3 Scheduling Algorithm

```text
pending = work_units in manifest order
running = empty map
completed_keys = empty set
failed_keys = empty set
primary_failure = null
scheduler_stopped = false
emit scheduler_started

while pending is not empty or running is not empty:
  ready = pending units whose dependencies are all in completed_keys
          and whose dependencies are not in failed_keys

  for unit in ready by priority DESC, weight_ms DESC, manifest ordinal ASC, id ASC:
    if scheduler_stopped:
      break
    if any earlier ready unit is resource-blocked and overlaps unit.resource_claims:
      continue
    if resources_available(unit.resource_claims):
      start unit
      remove unit from pending
      add unit to running
      emit unit_started

  if running is empty:
    if pending contains units with failed dependencies:
      mark them skipped in manifest order
      remove them from pending
      continue
    fail with scheduler_accounting_error for deadlock or impossible resources

  wait until one or more running units finish or the progress tick fires

  if progress tick fires:
    emit progress event
    continue

  process finished units by:
    1. observed_monotonic_finished_at ascending,
    2. manifest ordinal ascending

  for each finished unit:
    release non-retained resources
    record status, logs, duration, completion keys, failure keys
    if success or unit.complete_on_failure:
      add completion_keys
    if failure:
      add failure_keys
      set primary_failure if null
      if stop_on_first_failure:
        scheduler_stopped = true

run finalizers in manifest order after all running units drain
release retained resource claims
emit scheduler summary, progress summary, and scheduler_complete
validate timing when validate_timing is true; validation failures are `scheduler_accounting_error`
exit with selected primary failure
```

Finalizer failure becomes primary only when no earlier non-finalizer failure exists.

Work units skipped because `stop_on_first_failure` has selected a primary failed work unit, or because their dependencies include a failed key, are propagation records. They MUST be retained in scheduler and target summaries with their skipped reason and failed dependency, but they MUST NOT be reported as additional root failures.

Dependencies outrank priority: a work unit is not ready until its dependencies are satisfied. Priority affects only ready work and MUST NOT preempt work that is already running.

For the `check` scheduler, priority assignments MUST preserve the service-backed critical path once service readiness exists. A ready `check-service-backed` service session and its expanded browser stage, browser group, Go shard, backend make target, aggregate finalizer, and service-complete child work MUST have higher priority than post-build local evidence, static validation, catalog validation, and drift validation work. Build readiness and service-image readiness MAY remain above service-backed child work when those dependencies are still required to create valid service-backed evidence. Lower-priority ready work MAY start only when it does not overlap the resource claims of an earlier ready service-backed child that is resource-blocked.

Priority reservation MUST NOT create scheduler deadlock around retained
lifecycle claims. When no child process is running and at least one ready work
unit fits the current resource limits, the scheduler MUST admit the earliest
such unit even when an earlier resource-blocked unit reserved an overlapping
resource. This liveness fallback applies only to an otherwise idle scheduler;
ordinary admission while work is running continues to preserve the priority
reservation rule above.

### 10.4 Event Ordering

| Event field             | Rule                                                                         |
| ----------------------- | ---------------------------------------------------------------------------- |
| `schema_id`             | `cartulary.scheduler_event.v7`.                                              |
| `target`                | Public target or scheduler target identity.                                  |
| `scheduler_kind`        | Scheduler family `check`, `service_backed`, `test_slice`, or `sequence`.     |
| `seq`                   | Starts at `1`, increments by `1`, no gaps.                                   |
| `event`                 | Compact event token such as `scheduler-started`, `unit-started`, or `progress`. |
| `monotonic_ms`          | Non-decreasing scheduler-relative monotonic time.                            |
| `emitted_at`            | RFC3339 UTC. Wall-clock regressions require a `clock-skew` marker event.     |
| Work-unit ordering      | Manifest ordinal unless completion tie rule applies.                         |
| Completion tie          | `observed_monotonic_finished_at` ascending, then manifest ordinal ascending. |
| Artifact ordering       | Lexicographic by normalized artifact path.                                   |
| Resource ordering       | Registry display order, then lexicographic fallback.                         |

## 11. Service and Fixture Lifecycle

**TH-HARNESS-REQ-400**
Service-backed targets MUST run in exactly one service mode: `owned` or `attach`.
Verified by: TH-HARNESS-AC-007

| Mode     | Selection rule                                       | Missing variables                                                      | Ownership                                                                     |
| -------- | ---------------------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `owned`  | `CARTULARY_TEST_SERVICES_ACTIVE` omitted or not `1`. | Not applicable.                                                        | Harness starts and cleans suite resources.                                    |
| `attach` | `CARTULARY_TEST_SERVICES_ACTIVE=1`.                  | Any missing required attach variable fails with `configuration_error`. | Harness uses supplied services and MUST NOT delete container-level resources. |

### 11.0 Postgres Fixture Model

**TH-HARNESS-REQ-405**
Postgres fixture selection MUST be intent-based so catalog growth does not turn fixture setup into hidden critical-path cost. Registered fixture profiles are the sole owner for row fixture policy, fixture budget, and clone/reset exception reason. Helper code MUST fail closed when service-backed rows lack a resolvable fixture profile or when the profile omits an applicable PostgreSQL policy or budget.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-018

| Fixture class       | Policy tokens                         | Intended use                                                                                 | Guardrails                                                                 |
| ------------------- | ------------------------------------- | -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `transaction`       | `transaction`                         | Store-layer tests that use the narrow `postgres.DB` interface and can roll back all writes.  | MUST NOT use package reset; records transaction events instead of resets.   |
| `reusable_database` | `package_reset`, `group_clone`        | Route/runtime tests proven safe to reuse a package or group DB with a closed reset surface.  | Package resets MUST declare a closed `dirty_tables` set when targeted, or an explicit reset reason when broad reset is intentional; grouped clones MUST declare why shared committed state is safe. |
| `isolated_clone`    | `template_clone`                      | Tests needing committed route/runtime state, DB identity, schema mutation, process/runtime isolation, or unsafe reset state. | Explicit clone profiles MUST declare why clone isolation is required. |
| `migration_scratch` | `migration_scratch`, migration helper | Migration runner, upgrade, downgrade, and backfill-path assertions only.                     | Current-head schema assertions MUST use the schema template instead.        |

The suite template database MUST be named from the suite ID and a schema hash derived from sorted migration SQL inputs plus the migration runner identity. Fixture events and summaries MUST include `schema_hash`, `fixture_class`, and `reuse_group` when a package or group reuse key exists. Attach mode MUST fail before child execution when an advertised `CARTULARY_PGTEST_SCHEMA_HASH` does not match the local migration schema hash.

Service-backed Go fixture profiles MUST explicitly declare `fixture_policy.postgres` and `fixture_budget.postgres`. The current fixture-budget fields are `max_template_clones`, `max_group_clones`, `max_package_resets`, `max_transactions`, `max_migration_scratch`, `dirty_tables`, and `reset_conformance`. Policy selection MUST prefer `transaction` for rollback-safe tests, `package_reset` for committed-state tests with a known bounded reset surface, `group_clone` for shared seeded packages, and `template_clone` only for schema mutation, process lifecycle, migrations, destructive residue, or isolation-sensitive tests. Multi-symbol Go rows MAY reference closed symbol-specific fixture-profile overrides keyed only by symbols declared in the row selector. Overrides MUST NOT introduce symbols, widen the selected row set, or change target selection. Explicit `template_clone`, `group_clone`, broad `package_reset`, and `migration_scratch` profiles MUST carry the corresponding human reason and closed reason-code field. The current closed reason codes are `committed_cross_connection_visibility`, `database_identity`, `process_lifecycle`, `schema_mutation`, `destructive_residue`, `shared_seeded_state`, `bounded_reset_surface`, and `migration_scratch`. Stale or policy-inapplicable reason fields MUST fail profile validation.

| Policy token | Required reason field | Admissible reason-code values |
| --- | --- | --- |
| `template_clone` | `template_clone_reason`, `template_clone_reason_code` | `committed_cross_connection_visibility`, `database_identity`, `process_lifecycle`, `schema_mutation`, `destructive_residue` |
| `group_clone` | `group_clone_reason`, `group_clone_reason_code` | `shared_seeded_state` |
| `package_reset` | `package_reset_reason`, `package_reset_reason_code` when explicit broad or targeted reset evidence is declared | `bounded_reset_surface` |
| `migration_scratch` | `migration_scratch_reason`, `migration_scratch_reason_code` | `migration_scratch` |
| `transaction` | none | none |

`migration_scratch_reason` MUST justify migration, boundary, replay, upgrade, or backfill coverage. Store-layer rows that are transaction-safe MUST use `transaction` and MUST NOT use package reset. A reusable package reset MUST declare a foreign-key-closed targeted table set; a broad mutable reset requires an explicit reason and accepted proof. Backend-store package reset remains unsupported. Safe committed route groups MUST use `group_clone` when the accepted proof shows grouping preserves isolation and assertion clarity. `max_group_clones` counts physical `group-reused` database creations, not catalog rows. Nested subtests, child HTTP runtimes, or adapter-variation cases under one `group_clone` row MUST either reuse a parent-scoped group database or be separately budgeted; they MUST NOT multiply databases behind one shared-state reason.

Fixture-tier lowering MUST be backed by a closed proof model before row evidence is relabeled. A retained proof artifact MUST validate as `cartulary.fixture_tier_proof.v2` and MUST identify `target`, `owner_id`, `row_id`, optional `symbol`, `execution_family`, `effective_fixture_policy`, `proof_kind`, `proof_status`, optional `proof_ref`, human `reason`, `execution_boundary`, `observed_surfaces`, `reset_surface`, and `final_verdict`. Current proof statuses and verdicts are `accepted`, `retained`, and `blocked`. Proof artifacts MUST NOT reduce selected rows, enable product-test reuse, make historical artifacts current, or change accepted fixture-profile tokens.

The current proof admission rules are closed:

| Policy token | Required execution boundary | Required proof status before lower-tier admission | Required reset/surface closure |
| --- | --- | --- | --- |
| `transaction` | `rollback_transaction` | `accepted` for row-specific lowering or symbol override proof. | The proof MUST show a rollback-scoped store fixture or equivalent transaction boundary, no committed cross-connection visibility requirement, no WebSocket observer residue, and no package reset event. |
| `package_reset` | `committed_package_reset` | `accepted` before any product row may use package reset; raw reset-conformance helper rows are implementation-support only. | The proof MUST include FK-closed `dirty_tables[]` matching `fixture_budget.postgres.dirty_tables`, preserved `goose_db_version`, captured before/after row counts, route-idempotency cleanup or non-applicability, no schema mutation, no process restart dependency, and `object_store=none` or proven object cleanup. Backend-store package reset remains unsupported. |
| `group_clone` | `shared_group_database` | `accepted` when grouped committed state is intentional. | The proof MUST name the shared seeded state, the grouping key, and why all grouped symbols can share committed state without hiding per-test isolation requirements. Nested child runtimes MUST reuse the parent group database or be separately budgeted. |
| `template_clone` | `isolated_template_clone` | `retained` when clone isolation is intentionally preserved. | The proof MUST name the clone-only boundary being preserved, such as process lifecycle, schema mutation, database identity, destructive residue, WebSocket/cross-connection observer semantics, unsafe object-store effects, or unproven cleanup. |
| `migration_scratch` | `migration_scratch_database` | `retained` for migration-path evidence. | The proof MUST stay scoped to migration, replay, upgrade, downgrade, boundary, or backfill coverage and MUST NOT be used for current-head product runtime rows. |

Proof must account for the whole harness execution boundary, not only the product row's narrow assertion. Route/runtime proof MUST include auth/session/bootstrap side effects, `route_idempotency`, jobs, object-store buckets or prefixes, process lifecycle, WebSocket observers, cross-connection observers, and any schema or migration side effects when those surfaces are observed or required. A "no durable rows" product assertion is insufficient for reset admission unless every harness-created durable side effect in the same execution boundary is captured and reset-safe. WebSocket, cross-connection observer, process lifecycle, recovery, and operator-binary rows are clone-only unless proof records committed observer setup, teardown, no pending events, no cross-test connection residue, and no object/job residue.

Service-backed fixture shape validation MUST fail unplanned `template_clone` use, unplanned `group_clone` use, unplanned transaction use, forbidden package resets, migration scratch overuse, and declared structural fixture-count overuse. Structural overuse diagnostics for template clones, group clones, and transactions MUST identify bounded actual source details, including top-level test, full test name, caller package or file when available, reuse group when present, actual count, declared budget, and planned manifest symbols. Explicit clone rows MUST carry a clone reason, package resets MUST stay below the current warm target of `30` reset operations and `60000ms` executed reset time, and retained fixture reports MUST distinguish transaction, package reset, group clone, and template clone events so warm scheduler health can identify fixture pressure instead of hiding it inside shard time. Fixture duration evidence is diagnostic and advisory in the default local `check` path; explicit duration drift targets, warm scheduler timing checks, and `agent-finalize RESULTS_DIR=<dir>` own timing freshness. Raising `postgres_clone` capacity is valid only as a measured capacity calibration; it MUST NOT substitute for converting safe tests to transaction, template clone, or group clone policy.

### 11.1 Lifecycle Machine Contract

**TH-HARNESS-REQ-403**
A lifecycle machine is normative only when this NLSpec explicitly labels it normative. A representational lifecycle diagram MUST be labeled non-normative, MUST cite its owning requirements, and MUST NOT add behavior. A normative harness lifecycle machine MUST define scope, instance key, closed state set, closed event set, terminal states, transition table, guard precedence, failure mapping, authoritative state derivation, observable evidence, and conformance criteria. Illegal transitions MUST NOT mutate state, MUST fail closed with Section 9 failure classification, and MUST emit retained evidence. State-advancing artifact writes MUST be atomic and idempotent, or the machine MUST define guardrails that prevent unsafe re-execution. Parent lifecycle logic MUST depend only on child terminal status and retained artifacts, not on child in-memory state.
Verified by: TH-HARNESS-AC-017

Implementations MAY realize a normative lifecycle machine with ordinary control flow, tables, generated code, or a state-machine library. The runtime mechanism is not normative. The closed states, events, transitions, failure mapping, and observable evidence are normative.

Normative lifecycle-machine state and event names MUST be ASCII `lower_snake_case`. A transition table is closed by default: any `(state, event)` pair not listed by the owning machine is illegal.

### 11.2 Normative Service Suite Lifecycle Machine

**TH-HARNESS-REQ-404**
The service suite lifecycle machine is normative for every service-backed suite in `owned` or `attach` mode. The machine ID is `test_services_suite_lifecycle_v1`. The machine instance key is `suite_id`. The authoritative transition record is `_shared/test-services/<suite-id>/lifecycle-events.jsonl`, where every line MUST validate as `cartulary.test_services.lifecycle.v2`. The current state is `requested` before the first lifecycle event and otherwise the `to_state` of the last valid lifecycle event. A missing, malformed, non-sequential, or transition-invalid lifecycle event stream after a suite directory or lease exists MUST fail closed with `failure_class=artifact` and `failure_reason=artifact_error`. The service lease remains cleanup-proof evidence and MUST NOT be interpreted as a transition log.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-017

Lifecycle event `seq` starts at `1`, increments by `1`, and has no gaps. Events MUST be processed in emitted sequence order. When competing conditions are observed before the next event is emitted, guard precedence is:

| State           | Precedence rule                                      |
| --------------- | ---------------------------------------------------- |
| `starting`      | `startup_failed` before `readiness_passed`.          |
| `running_child` | `interrupt_received` before `child_started` before `child_finished` when multiple child signals are observed before the next event is emitted. |
| `cleaning`      | `cleanup_failed` before `cleanup_succeeded`.         |
| all others      | The transition table has at most one allowed event.  |

#### States

| State           | Kind         | Invariants                                                                 | Observable signals                                                                 |
| --------------- | ------------ | -------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `requested`     | initial      | Suite setup has been requested but no lifecycle event has been emitted.    | No lifecycle event stream exists for the selected `suite_id`.                      |
| `starting`      | intermediate | Suite setup or attach-mode validation is in progress; child work has not started. | Latest lifecycle event has `to_state=starting`; lease or startup diagnostics exist when the suite writes them. |
| `ready`         | intermediate | Required services or supplied attach endpoints have passed readiness.      | Latest lifecycle event has `to_state=ready`; readiness diagnostics are retained when produced. |
| `running_child` | intermediate | One or more child work units are executing under the suite.                 | Latest lifecycle event has `to_state=running_child`; the event references child logs or target artifacts when known and reports `active_child_count`. |
| `interrupted`   | intermediate | Cancellation or interruption was observed while child work was active.     | Latest lifecycle event has `to_state=interrupted` and `failure_reason=cancelled_or_interrupted`. |
| `cleaning`      | intermediate | Owned teardown or attach-mode diagnostic finalization is in progress.      | Latest lifecycle event has `to_state=cleaning`; lease `cleanup_state=in_progress` when a lease exists. |
| `cleaned`       | terminal     | Required cleanup or attach-mode finalization completed.                   | Latest lifecycle event has `to_state=cleaned`; lease `cleanup_state=completed` or `deferred`. |
| `failed_start`  | terminal     | Startup, attach validation, preflight, or readiness failed before child work started. | Latest lifecycle event has `to_state=failed_start`; failure summary records the Section 9 reason. |
| `cleanup_failed` | terminal   | Cleanup or finalization failed and retained proof remains for investigation or stale janitor handling. | Latest lifecycle event has `to_state=cleanup_failed`; lease `cleanup_state=failed` when a lease exists. |

#### Events

| Event                | Definition                                                                 |
| -------------------- | -------------------------------------------------------------------------- |
| `start_services`     | Begin owned suite startup or attach-mode suite validation.                 |
| `readiness_passed`   | All readiness predicates required by Section 11.4 passed before deadline.  |
| `startup_failed`     | Startup, attach validation, preflight, fixture preparation, or readiness failed before child work started. |
| `child_started`      | A child target, child command, or scheduler work unit started under a ready suite. |
| `child_finished`     | An active child target, child command, or scheduler work unit exited and its status was recorded. |
| `interrupt_received` | The wrapper observed cancellation or process interruption while child work was active. |
| `cleanup_started`    | Teardown, cleanup, or attach-mode diagnostic finalization started.         |
| `cleanup_succeeded`  | Teardown, cleanup, or attach-mode diagnostic finalization completed.       |
| `cleanup_failed`     | Teardown, cleanup, or attach-mode diagnostic finalization failed.          |

#### Transition Rules

| From state      | Event                | Guard                                      | To state         | Required actions                                                              | Failure mapping                                                                 | Observable evidence |
| --------------- | -------------------- | ------------------------------------------ | ---------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------- | ------------------- |
| `requested`     | `start_services`     | Configuration resolved; `suite_id` allocated; no prior lifecycle event exists. | `starting`       | Create suite directory as needed; write initial lease or attach diagnostic when applicable; append lifecycle event. | If setup cannot begin, fail before mutation with `configuration_error`, `preflight_error`, or `fixture_error` according to the failed predicate. | Lifecycle event `seq=1`; lease or diagnostic artifact refs when present. |
| `starting`      | `readiness_passed`   | All required readiness predicates passed.  | `ready`          | Append lifecycle event and retain readiness diagnostics when produced.         | none                                                                            | Lifecycle event with readiness artifact refs. |
| `starting`      | `startup_failed`     | Any startup, attach validation, preflight, fixture, or readiness predicate failed before child start. | `failed_start`   | Append lifecycle event; record failure summary; terminate known partial resources or leave proof for stale janitor. | `preflight_error`, `service_start_error`, `service_readiness_timeout`, `fixture_error`, or `configuration_error` according to Sections 9 and 11.4. | Lifecycle event with failure fields and proof artifact refs. |
| `ready`         | `child_started`      | Child key is non-empty and not already active. | `running_child`  | Append lifecycle event; set `active_child_count=1`; retain child log or target artifact refs when known. | Child start failure before process launch is `child_target_failure` or `fixture_error` according to the wrapper boundary. | Lifecycle event with child key, `active_child_count=1`, and child artifact refs when known. |
| `running_child` | `child_started`      | Child key is non-empty and not already active. | `running_child`  | Append lifecycle event; increment `active_child_count`; retain child log or target artifact refs when known. | Duplicate or missing child key is an illegal transition. | Lifecycle event with child key and incremented `active_child_count`. |
| `running_child` | `child_finished`     | Active child key is known, child status has been recorded, at least two children remain active before this event, and no interruption wins by guard precedence. | `running_child` | Append lifecycle event; decrement `active_child_count`; retain child status and artifacts. | Unknown child key or negative active count is an illegal transition. Child failure is recorded for primary failure selection by Section 9.1; the lifecycle state itself is not terminal. | Lifecycle event with child key, child status artifact refs, and decremented `active_child_count`. |
| `running_child` | `child_finished`     | Active child key is known, child status has been recorded, exactly one child remains active before this event, and no interruption wins by guard precedence. | `ready` | Append lifecycle event; set `active_child_count=0`; retain child status and artifacts. | Unknown child key or negative active count is an illegal transition. Child failure is recorded for primary failure selection by Section 9.1; the lifecycle state itself is not terminal. | Lifecycle event with child key, child status artifact refs, and `active_child_count=0`. |
| `running_child` | `interrupt_received` | Cancellation or signal wins by guard precedence. | `interrupted`    | Append lifecycle event and preserve child/interruption diagnostics when available; report current `active_child_count`. | `failure_class=interrupted`, `failure_reason=cancelled_or_interrupted`.         | Lifecycle event with interruption fields and `active_child_count`. |
| `ready`         | `cleanup_started`    | No child is running and cleanup/finalization is required. | `cleaning`       | Set lease `cleanup_state=in_progress` when a lease exists; append lifecycle event. | Cleanup start failure is recorded as `cleanup_error` without deleting unproven resources. | Lifecycle event and updated lease. |
| `interrupted`   | `cleanup_started`    | Interruption has been recorded and cleanup/finalization is required. | `cleaning`       | Set lease `cleanup_state=in_progress` when a lease exists; append lifecycle event. | Primary interruption failure is preserved by Section 9.1.                       | Lifecycle event and updated lease. |
| `cleaning`      | `cleanup_succeeded`  | Required cleanup or finalization completed. | `cleaned`        | Set lease `cleanup_state=completed` or `deferred`; append lifecycle event.     | Success unless an earlier primary failure exists.                               | Lifecycle event and final lease. |
| `cleaning`      | `cleanup_failed`     | Cleanup or finalization failed.            | `cleanup_failed` | Set lease `cleanup_state=failed`; append lifecycle event; retain proof for stale janitor. | `cleanup_error`; Section 9.1 decides whether it becomes the public exit-code driver. | Lifecycle event, final lease, and cleanup diagnostics. |

Any listed event presented in an unlisted state MUST append a lifecycle event with `transition_status=illegal`, `from_state` equal to `to_state`, `failure_class=harness`, and `failure_reason=scheduler_accounting_error`, then fail without mutating suite state. An unrecognized event token MUST be rejected before lifecycle mutation and MUST NOT be appended to the schema-valid lifecycle event stream. A terminal state MUST reject every later event as illegal.

Lifecycle failure events MUST carry normalized Section 9 failure fields. `startup_failed`, `interrupt_received`, `cleanup_failed`, and illegal-transition lifecycle events MUST set non-null `failure_class` and `failure_reason` in the lifecycle stream. Non-failure lifecycle events MUST preserve `failure_class=null` and `failure_reason=null` unless a later schema revision adds a narrower event-specific diagnostic field.

**TH-HARNESS-REQ-406**
The service-suite lifecycle active-child counter is normative. `child_started`, `child_finished`, and `interrupt_received` lifecycle events MUST include `active_child_count`. `ready + child_started` sets the count to `1`; `running_child + child_started` increments it; `running_child + child_finished` decrements it and remains in `running_child` while the count is greater than `0`; `running_child + child_finished` transitions to `ready` when the count becomes `0`. Negative active counts, missing child identity, duplicate `child_started` for the same active child key, and `child_finished` for an unknown active child key are illegal transitions under Section 11.2.
Verified by: TH-HARNESS-AC-017, TH-HARNESS-AC-033

**TH-HARNESS-REQ-411**
Every current-run service suite MUST retain `_shared/test-services/<suite-id>/service-scope.json` as `cartulary.test_services.scope.v1` after the suite artifact directory exists and before scheduler classification consumes service diagnostics. The scope summary MUST include `schema_id`, `target`, `run_id`, `suite_id`, `artifact_dir`, preflight, cleanup, service, fixture, and started-service summaries. When a service suite fails, `failure` MUST be present and MUST include normalized non-null `failure_class` and `failure_reason` fields. Startup preflight failures map to `infra/preflight_error`; service startup failures map to `infra/service_start_error`; service readiness deadline failures map to `infra/service_readiness_timeout`; fixture preparation failures map to `harness/fixture_error` unless Section 9 assigns the failed predicate to `artifact_error`; cleanup failures map to `harness/cleanup_error`.

Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-014, TH-HARNESS-AC-017, TH-HARNESS-AC-025

**TH-HARNESS-REQ-410**
An authoritative service-backed Go row MUST make its PostgreSQL fixture policy
explicit both in its referenced fixture profile and at the helper call site.
The helper call-site policy MUST select exactly one of transaction, package
reset, group clone, isolated template clone, or migration scratch and MUST
agree with the scheduler-resolved fixture-profile policy. Missing or conflicting
policy MUST fail before database preparation.

Package reset MUST be exposed only through an explicitly named reset helper
and remains admissible only with the closed reset proof, dirty-table surface,
reason, and budget required by TH-HARNESS-REQ-405. A generic helper MUST NOT
silently fall back to package reset, group clone, or template clone. Non-row
implementation-support tests MAY use an explicitly selected isolated clone,
transaction, or migration scratch without a product row, but the
selected intent must still be visible at the call site.

Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-056

### 11.3 Lease Fields

Lease files MUST be written before child work starts, MUST be redacted before retention, and MUST be written atomically as a complete JSON file. A lease is evidence for cleanup only when its resource proof matches the actual resource state; cleanup MUST verify labels, prefixes, generated names, or equivalent proof and MUST NOT trust the lease path alone.

| Field              | Type                                                               |                        Required |
| ------------------ | ------------------------------------------------------------------ | ------------------------------: |
| `schema_id`        | string, `cartulary.test_services.lease.v1`                         |                             yes |
| `lease_id`         | non-empty opaque lease identifier                                  |                             yes |
| `suite_id`         | 24 lowercase hex chars                                             |                             yes |
| `target`           | string                                                             |                             yes |
| `mode`             | `owned` or `attach`                                                |                             yes |
| `ownership_mode`   | `owned` or `attach`                                                |                             yes |
| `result_root`      | normalized path                                                    |                             yes |
| `run_id`           | normalized run ID                                                  |                             yes |
| `run_root`         | normalized run-root path                                           |                             yes |
| `owner_pid`        | integer process ID for the owning wrapper                          |                             yes |
| `created_at`       | RFC3339 UTC                                                        |                             yes |
| `heartbeat_at`     | RFC3339 UTC                                                        |                              no |
| `expires_at`       | RFC3339 UTC                                                        |                              no |
| `resources[]`      | redacted resource records with service kind, logical ID, and proof | yes in owned mode, may be empty |
| `proof_labels`     | object of required labels used to prove container ownership        |           yes for container use |
| `proof_prefixes`   | object of generated DB/bucket/path prefixes used to prove ownership | yes for DB, bucket, or path use |
| `cleanup_state`    | `not_started`, `in_progress`, `completed`, `failed`, or `deferred` |                             yes |

### 11.4 Readiness Deadlines

Browser owned-stack readiness is an ownership predicate, not only an HTTP availability predicate. The backend is ready only when all of the following hold before the deadline:

- the wrapper-started backend process group is still alive;
- the selected backend port is owned by a process in that process group, when the platform exposes listener process metadata;
- the token-protected `GET /api/v1/test/runtime/identity` route returns `cartulary.test.runtime_identity.v1`, `runtime_marker="harness-owned"`, `test_routes_enabled=true`, and a server process ID;
- the backend process group remains alive after the identity probe.

The canonical browser E2E frontend startup mode is a built preview: `build-web` MUST complete before browser stack startup, `apps/web/dist/index.html` MUST exist before the frontend process starts, and the wrapper MUST launch a non-watching preview command rather than the Vite dev server. Missing built frontend artifacts are `configuration_error`, exit `2`; they are not service-readiness failures.

The frontend is ready only when the wrapper-started preview process group is still alive, the selected frontend port is owned by a process in that process group when listener process metadata is available, the frontend HTTP probe succeeds, the process group remains alive after the probe, and stack metadata records `frontend_mode="preview"` and `frontend_command_kind="vite-preview"`. A stale or unrelated listener MUST NOT satisfy browser readiness.

| Resource                     | Deadline | Poll interval | Failure reason                       |
| ---------------------------- | -------: | ------------: | ------------------------------------ |
| Docker preflight             |    `15s` |          `1s` | `preflight_error`                    |
| Postgres container readiness |   `180s` |       `500ms` | `service_readiness_timeout`          |
| local object-store readiness |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Template DB migration        |   `180s` |           n/a | `fixture_error`                      |
| Browser backend readiness    |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Browser frontend readiness   |   `120s` |       `500ms` | `service_readiness_timeout`          |
| Reset route success          |    `30s` |           n/a | `fixture_error` or `timeout_failure` |

### 11.5 Retry and Teardown Rules

**TH-HARNESS-REQ-401**
No hidden startup retry is allowed. Retry is allowed only when a resource row declares `max_attempts`, bounded backoff, retryable failure reasons, and an overall deadline. Readiness polling within a deadline is not a retry.
Verified by: TH-HARNESS-AC-007

**TH-HARNESS-REQ-402**
Owned teardown order MUST be: browser child processes, browser fixtures, reset-tainted runtime roots, test databases, object buckets or prefixes, service containers, lease finalization. Attach mode MUST record diagnostics but MUST NOT delete container-level resources or external services.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-010

Destructive reset, cleanup, attach-mode service mutation, and non-idempotent operations MUST NOT be retried unless a resource row explicitly declares the operation safe to retry.

| Resource operation          | `max_attempts` | Backoff | Retryable failure reasons                                  | Overall deadline | Safe retry scope                                  |
| --------------------------- | -------------: | ------- | ---------------------------------------------------------- | ---------------- | ------------------------------------------------- |
| Docker preflight            |            `1` | none    | none                                                       | `15s`            | none                                              |
| Postgres owned startup      |            `3` | `500ms` | transient Docker startup or transport failure before readiness polling begins | attempt startup only; readiness deadline is Section 11.4 `180s` | Failed attempt container is terminated first.     |
| object-store owned startup  |            `2` | `250ms` | transient Docker startup or transport failure before readiness polling begins | attempt startup only; readiness deadline is Section 11.4 `120s` | Failed attempt container is terminated first.     |
| Template DB migration       |            `1` | none    | none                                                       | `180s`           | none                                              |
| Browser backend startup     |            `1` | none    | none                                                       | `120s`           | readiness polling only                            |
| Browser frontend startup    |            `1` | none    | none                                                       | `120s`            | strict-port conflicts fail as `resource_conflict` |
| Runtime reset route         |            `1` | none    | none                                                       | `30s`            | none                                              |
| Owned teardown and cleanup  |            `1` | none    | none                                                       | cleanup-specific | cleanup records failure and leaves proof for janitor |

Stale janitor cleanup of previously owned service containers is a proof-gated startup preflight maintenance step, not authoritative product evidence. Once ownership proof and current-suite exclusion pass, Docker `not found` and Docker "removal already in progress" results MUST be accepted as idempotent cleanup outcomes and MUST NOT fail the new suite. Concurrent removal MUST be retained as deferred cleanup diagnostics. Docker daemon/list failures, unsafe ownership proof, and non-idempotent removal failures remain blocking startup-preflight failures.

**TH-HARNESS-REQ-412**
If Docker container deletion for a proven stale owned service container returns `context deadline exceeded`, cancellation, or an equivalent bounded remove timeout, startup preflight MUST perform one bounded post-delete recheck before deciding whether to fail. The recheck MUST use Docker container state for the same container ID after the original ownership proof and current-suite exclusion have already passed. If the recheck proves the container is gone, the outcome is idempotent removal and counts as removed. If the recheck proves the container is in Docker `removing` or `dead` state, the outcome is deferred idempotent cleanup and counts as deferred. If the recheck cannot read Docker state, or proves the same container is still present without `removing` or `dead` state, the original deletion timeout remains a blocking startup-preflight failure with `failure_class=infra` and `failure_reason=preflight_error`.

This timeout acceptance is proof-gated only. A stale-container delete timeout without a successful `not found`, `removing`, or `dead` recheck MUST NOT be accepted.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-010, TH-HARNESS-AC-017

Attach mode MAY write diagnostic records and lease observations. It MUST NOT delete externally supplied services, containers, databases, buckets, or object prefixes.

For browser owned stacks, a listener conflict detected before process startup or during backend/frontend process bind/startup maps to `resource_conflict`. Backend or frontend process exit before readiness maps to `service_start_error` only when retained startup diagnostics and logs do not identify a listener, port, lock, or other resource conflict. A live owned process that does not satisfy its readiness predicates before the deadline maps to `service_readiness_timeout`. Suite-admin login failures after owned readiness has been proven are no longer treated as readiness failures.

Startup retry windows and readiness deadlines are separate. If a Postgres or object-store startup attempt reaches readiness polling and then the Section 11.4 readiness deadline expires, the operation MUST NOT retry that service. The failure MUST be `failure_class=infra`, `failure_reason=service_readiness_timeout`, and public exit `3`. Browser backend startup, browser frontend startup, runtime reset, and cleanup have `max_attempts=1`; their polling or operation deadlines do not create retry attempts. Browser ports are dynamically allocated before process startup, but a later strict-port collision is terminal `failure_class=infra`, `failure_reason=resource_conflict`; silently changing the admitted port or replacing terminal startup evidence is forbidden.

### 11.6 Duration Baselines

Duration baselines are advisory scheduler planning data only. They MUST NOT become benchmark claims, product performance conformance, timeout policy, or evidence that product behavior is fast enough.

Baseline values MUST be positive integer `weight_ms` values derived only from successful, uncontaminated retained runs. Missing entries MUST use explicit default weights and MUST be reported as defaulted, not silently ignored.

Baseline refresh MUST reject contaminated evidence, including failed scheduler runs, service startup retries, service failures, reset taint, missing timing events, or interrupted runs.

For row-keyed browser E2E duration baselines, refresh MUST join retained timing evidence to the active catalog by row ID. Refresh MAY replace stale stored file/title metadata with the active registered selector metadata for that row ID. Planning and drift validation MUST remain strict and reject stale selector metadata instead of silently using it.

Mutating `browser-e2e-duration-baselines` refresh MUST require an explicit, existing `RESULTS_DIR` retained run selector and MUST reject omitted or ambiguous retained evidence. Read-only `browser-e2e-duration-baseline-drift` MAY use the current retained run default only where the public input matrix declares that default; any caller-supplied `RESULTS_DIR` MUST still resolve to an existing retained root or run root.

Raw Go duration baselines for package-level harness suites MUST be stored and checked by raw package baseline key. Current shard planning, coverage, and drift validation MUST NOT use an aggregate raw-suite duration as a fallback for missing package baselines, and current shard plans MUST NOT emit legacy aggregate baseline keys.

`agent-finalize RESULTS_DIR=<dir>` MUST perform retained-run validation before any duration-baseline refresh action mutates committed baseline artifacts. That validation MUST reject failed, incomplete, contaminated, or non-warm retained evidence before the first mutating refresh substep starts.

Duration-baseline drift checks MAY fail only for severe stale planning. Compact drift diagnostics MUST include `subject`, `planned_ms`, `actual_ms`, `ratio`, and `kind`.

Warm scheduler health checks MAY consume retained timing artifacts from a successful warm-ready run. Such checks MUST remain harness-maintenance evidence and MUST NOT be described as claim-bearing product benchmark evidence. When a warm `check` artifact is evaluated, the check MUST fail if default local `check-service-backed` includes ordinary browser measurement work, if hidden provisioning prevents warm eligibility, if `check-service-backed` exceeds the configured warm budget, or if non-isolated backend/browser peer lanes exceed the configured balance ratio by more than the bounded materiality floor. Unless the caller supplies a different value, the supported WSL2 hard warm-maintenance budget for `check-service-backed` is `155000ms`, the balance ratio is `1.25`, and the materiality floor is `5000ms`.

### 11.7 Network Flow Fixture Materialization Lifecycle

**TH-HARNESS-REQ-407**
Network Flow fixture source roots are committed fixture inputs, not service-owned mutable state. A harness runner MUST NOT write into `fixtures/network-flow/**` during validation, preview, apply, graph, cursor, indicator-link, or transcript comparison work. Any generated, staged, normalized, or copied fixture material MUST live under the current run root and MUST be removed only by ordinary result-root cleanup. `make clean`, service teardown, stale janitors, database reset, and object-store reset MUST NOT delete committed Network Flow fixture roots.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-408**
Before a Network Flow fixture is materialized, the runner MUST validate the manifest schema, path safety, per-file byte hashes, aggregate bundle hashes, frozen status when selected for conformance, and owner routing. Materialization MUST use only manifest-listed files, MUST reject symlinks and traversal paths, and MUST make product execution observe a read-only run-local copy. A failed pre-materialization check is `failure_class=artifact`, `failure_reason=artifact_error`; product code MUST NOT start for that fixture.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-409**
Network Flow fixture materialization participates in the existing service lifecycle only as fixture preparation. If materialization fails before child work starts, the service-suite lifecycle records `startup_failed` with `fixture_error` or `artifact_error` according to Section 9 ownership, preserves diagnostics, and performs ordinary owned teardown. If child product work has started, later comparison failures are product or artifact failures according to the owning assertion, but the committed fixture root remains immutable.
Verified by: TH-HARNESS-AC-049

### 11.8 Local Development Object-Store Proxy

**TH-HARNESS-REQ-413**
The local development object-store proxy is development-only and MAY retain
`127.0.0.1:8333`; browser evidence MUST never depend on it. Proxy lifecycle
operations MUST serialize through one OS advisory operation lock. State,
operation metadata, startup attempts, ready leases, and per-instance logs MUST
reside beneath one owner-only repo runtime root as non-symlink regular files.
State publication uses an owner-only temporary file, file `fsync`, atomic
rename, and parent-directory `fsync`.

The launcher MUST publish a secure
`cartulary.local_object_store_proxy_start_attempt.v1` before spawn. The child
MUST synchronously bind the exact loopback listener, then publish its instance
identity and full process proof before serving normal proxy traffic. The proof
contains Linux boot ID, PID, `/proc/<pid>/stat` start-time ticks, effective UID,
executable device and inode, and SHA-256 of `/proc/<pid>/exe`. Promotion to
`cartulary.local_object_store_proxy_lease.v1` is atomic and permitted only after
the five-second bind/identity handshake, closed
`cartulary.local_object_store_proxy_health.v1` identity, and the separately
bounded object-store plus exact-CORS probe all pass.

Recovery or reuse requires matching process, executable, health, listener,
instance, and canonical nonsecret configuration proofs. A fully proven
configuration mismatch is gracefully restarted. A stale or legacy PID file is
untrusted metadata and MUST never authorize signaling. An unproven listener is
an immediate `resource_conflict`. Signaling requires `pidfd_open`, complete
proof revalidation after opening the pidfd, `pidfd_send_signal`, and confirmed
termination through the pidfd. Unsupported pidfd behavior fails closed without
PID-only fallback. A startup attempt abandoned before process proof may be
discarded only when no listener occupies the configured endpoint.

The upstream origin is canonicalized and MUST reject userinfo, query, and
fragment; listener configuration MUST be an explicit loopback IP. Configuration
fingerprints include only canonical nonsecret values. Development proxy health
is loopback-only implementation support and MUST NOT become a product API,
production deployment surface, browser attachment input, or product evidence.
Verified by: TH-HARNESS-AC-007, TH-HARNESS-AC-010

## 12. Test-Only Harness Routes

**TH-HARNESS-REQ-450**
Test-only harness routes are harness routes. Runtime-control routes include `POST /api/v1/test/runtime/reset`, `GET /api/v1/test/runtime/identity`, `POST /api/v1/test/clock/set`, `POST /api/v1/test/clock/reset`, `GET /api/v1/test/clock/state`, `POST /api/v1/test/runtime/public-error-faults`, `POST /api/v1/test/runtime/network-flow-faults`, `POST /api/v1/test/runtime/network-flow-randomness`, `POST /api/v1/test/runtime/network-flow-auth-transitions`, and `POST /api/v1/test/runtime/network-flow-audit-assertions`. Fixture routes include `POST /api/v1/test/incidents/{incident_id}/saved-views/system`. Any future `/api/v1/test/*` or `/ws/v1/test/*` route that observes or mutates harness runtime state or fixture state is also a test-only harness route. These routes MUST be unavailable unless every enablement predicate below is satisfied. They MUST NOT be documented as production API behavior.

Runtime-linked implementations of Section 12 routes that are registered by application binaries are owned by the platform harness runtime adapter boundary. Private implementation package paths are not public route contracts. Non-test runtime code MUST NOT import broad test-helper packages to register or execute Section 12 routes once a platform harness runtime adapter exists for that route family.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-013

### 12.1 Enablement

**TH-HARNESS-REQ-480**
An in-process application test server MUST receive an explicit test-route
mode. The closed modes are disabled, harness-owned, and custom environment.
An omitted, empty, or unknown mode MUST fail before runtime construction.
Harness-owned mode is admissible only when the test exercises a guarded
test-route contract or registers an owner test-route contribution; ordinary
product-route tests MUST use disabled mode. Custom-environment mode is limited
to negative configuration, authorization, host, origin, token, and process
composition tests that supply the complete environment under test.

This requirement changes private test-helper setup only. It does not change
the production enablement predicates below.

Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-056

| Predicate                      | Required value                                                                      |
| ------------------------------ | ----------------------------------------------------------------------------------- |
| `CARTULARY_ENABLE_TEST_ROUTES` | Exact `1`.                                                                          |
| `CARTULARY_TEST_ROUTE_TOKEN`   | Non-empty string with at least 128 bits of entropy, generated by the harness stack. |
| Runtime ownership              | Server started by a Make-owned browser or test harness stack.                       |
| Production/default runtime     | Test-only harness routes are not registered.                                         |

**TH-HARNESS-REQ-453**
The test-route authorization header name is exactly `X-Cartulary-Test-Route-Token`. Harness-generated route tokens MUST be 32 bytes from a cryptographically secure pseudorandom generator encoded as unpadded base64url, producing exactly 43 ASCII characters. Attach-mode supplied tokens MUST be ASCII visible characters, MUST be length `43..512`, MUST contain no whitespace, MUST NOT equal `test`, `token`, `secret`, `password`, or `changeme`, and MUST NOT be a repeated single-character string. Missing, malformed, or weak attach-mode tokens MUST fail before test-route registration with `failure_class=config`, `failure_reason=configuration_error`, and public exit code `2`.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-035

### 12.2 Authorization

| Request condition                                     | Behavior                                                     |
| ----------------------------------------------------- | ------------------------------------------------------------ |
| Test-only harness route not enabled                   | Return ordinary not-found behavior.                          |
| Test-only harness route enabled, missing token header | `403`, `error.code=test_route_forbidden`.                    |
| Test-only harness route enabled, wrong token header   | `403`, `error.code=test_route_forbidden`.                    |
| Test-only harness route enabled, correct token header | Evaluate request after host/origin boundary checks.          |
| Cookie-authenticated request without token            | Forbidden; session auth does not authorize test-only routes. |
| Bearer/session/bootstrap-token request without token  | Forbidden; product auth does not authorize test-only routes. |

CSRF does not apply because cookie authentication is not accepted as authorization for test-only harness routes. Incident roles, session cookies, bearer sessions, bootstrap tokens, and `deployment_admin` authority do not bypass the test-route token requirement.

**TH-HARNESS-REQ-451**
When test routes are enabled and a harness-owned API or browser origin is configured, test-only harness routes MUST reject requests whose request origin is not the harness-declared browser origin or harness-declared API origin, or whose request host does not match the harness-owned API origin. Same-process health and readiness probes MUST use explicitly declared non-destructive health endpoints rather than test-only harness routes. Test-only harness routes MUST NOT enable permissive CORS. A rejected origin or host MUST fail before any runtime-control mutation or fixture mutation with `403`, `error.code=test_route_forbidden`.
Verified by: TH-HARNESS-AC-008

### 12.2.1 Runtime Identity

`GET /api/v1/test/runtime/identity` is a harness test route with the same enablement and authorization predicates as the reset route. It returns `cartulary.test.runtime_identity.v1`, the harness runtime marker, `test_routes_enabled=true`, and the server process ID. It MUST NOT return the route token or any database, object-store, credential, or session secret. Browser wrappers and Playwright global setup use this route to prove that the selected API origin belongs to the current harness-owned backend before destructive reset or suite-admin login work begins.

### 12.2.2 Test Clock

**TH-HARNESS-REQ-460**
The test clock routes are harness test routes with the same enablement, host/origin, and token authorization predicates as the reset route. The route family consists of `POST /api/v1/test/clock/set`, `POST /api/v1/test/clock/reset`, and `GET /api/v1/test/clock/state`. They MAY control or observe the harness-owned runtime clock for scenarios that need deterministic authentication, cursor TTL, timestamp, timezone, source-retention, or session-expiry timing. Because that clock can feed security-sensitive product decisions in harness-owned runtimes, missing token, wrong token, product session credentials alone, wrong host, missing origin when origins are configured, malformed origin, or unapproved origin MUST fail before clock mutation or state disclosure with `403`, `error.code=test_route_forbidden`.
Verified by: TH-HARNESS-AC-035, TH-HARNESS-AC-051

**TH-HARNESS-REQ-461**
`POST /api/v1/test/clock/set` accepts a JSON object with exactly one of `fixed_now` or `offset_seconds`. `fixed_now` MUST be an RFC3339/RFC3339Nano timestamp and is normalized to UTC. `offset_seconds` MUST set the clock to wall time plus that offset and MUST clear any fixed clock. Unknown members, invalid JSON, non-object JSON, trailing JSON, missing both command fields, or specifying both command fields MUST fail with `400`, `error.code=invalid_mutation_payload`.
Verified by: TH-HARNESS-AC-035, TH-HARNESS-AC-051

**TH-HARNESS-REQ-462**
`POST /api/v1/test/clock/reset` accepts no body or `{}` and restores wall-clock mode by clearing both fixed time and offset. Unknown members, invalid JSON, non-object JSON, or trailing JSON MUST fail with `400`, `error.code=invalid_mutation_payload`. `GET /api/v1/test/clock/state` MUST return the current test-clock state without mutation.
Verified by: TH-HARNESS-AC-051

**TH-HARNESS-REQ-463**
Successful test-clock set, reset, and state responses MUST return `cartulary.test.clock_control.v1` in the standard success envelope. The response MUST include `schema_id`, `mode`, `now`, and `offset_seconds`; `mode` is one of `wall`, `offset`, or `fixed`; `now` is the current harness clock in RFC3339Nano UTC form; `fixed_now` appears only when `mode="fixed"`. The response MUST NOT include the test-route token, cookies, product session credentials, database credentials, object-store credentials, host filesystem paths, or private runtime state.
Verified by: TH-HARNESS-AC-051

**TH-HARNESS-REQ-464**
Runtime reset MUST clear any registered in-memory test clock before reset success is accepted, restoring `mode="wall"` and `offset_seconds=0`. A fixture or target that changes the test clock MUST either run in an owned runtime that is torn down afterward or restore wall mode before subsequent unrelated product work starts. Network Flow fixtures that rely on cursor TTL, safe-digest rotation windows, source-retention expiry, soft-delete timing, timezone fold/gap interpretation, uptime-derived timestamps, or timestamp ordinal boundaries MUST record the selected clock-control response in their transcript and MUST cite the adopted product owner requirement that defines the expected time behavior.
Verified by: TH-HARNESS-AC-051

### 12.2.3 Saved-View System Fixture

`POST /api/v1/test/incidents/{incident_id}/saved-views/system` is a harness fixture route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY seed one incident-bound `scope='system'` saved-view fixture per successful request for browser scenarios that must distinguish implementation-owned saved-view configurations from contract-backed system views. It MUST NOT be exposed as production API behavior, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or creating any saved-view row.

This fixture route MUST create only saved-view fixture rows through the saved-view store path. It MUST NOT expose arbitrary SQL execution, projection mutation, generic fixture mutation, or caller-supplied saved-view identity, owner, timestamps, version, or scope. The route fixes `scope='system'`, fixes `owner_user_id=null`, derives `incident_id` from the path, and returns the normal saved-view resource in the standard success envelope with HTTP `201`.

### 12.2.4 Public-Error Fault Control

`POST /api/v1/test/runtime/public-error-faults` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm a one-shot public error envelope for the next exact ordinary `/api/v1/` request whose method and path match the request body. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any fault.

The armed fault is in-memory harness runtime state. It is consumed at the service boundary before the ordinary route handler runs. Matching is exact on uppercase HTTP method and request path; query strings and fragments are not part of the accepted path. Faults MUST apply only to paths that start with `/api/v1/` and MUST NOT apply to paths that start with `/api/v1/test/`. Test-control routes therefore cannot fault themselves or other test-only harness controls.

The request body MUST be a JSON object with exactly the fields below.

| Field          | Required | Behavior                                                                                             |
| -------------- | -------- | ---------------------------------------------------------------------------------------------------- |
| `method`       | yes      | Non-empty HTTP method; normalized to uppercase for exact matching.                                   |
| `path`         | yes      | Exact ordinary public route path beginning `/api/v1/` and not beginning `/api/v1/test/`; no query.   |
| `status`       | yes      | Public error status from `400` through `599`.                                                        |
| `code`         | yes      | Non-empty public error code.                                                                         |
| `message`      | no       | Public error message to place in the public envelope.                                                |
| `retryable`    | no       | Public retryability flag; omitted values default to `false`.                                         |
| `details`      | no       | Public error details object. Consumers MUST render only details keys allowlisted by product UI code. |
| `consume_once` | yes      | Must be `true`; persistent or multi-consume faults are not accepted.                                 |

Unknown members, missing required fields, non-object JSON, invalid JSON, status outside `400..599`, empty `code`, a path outside ordinary `/api/v1/`, a path under `/api/v1/test/`, a path with query or fragment, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_public_error_fault_request`.

Successful arming MUST return HTTP `201` with `cartulary.test.public_error_fault.v1` in the standard success envelope. The response MUST include a generated `fault_id`, normalized `method`, exact `path`, `status`, `code`, `retryable`, and `consume_once=true`. The response MUST NOT include the test-route token, configured origins, cookies, product session credentials, database credentials, object-store credentials, or private runtime state.

The next exact ordinary public-route match MUST return a standard public error envelope using the armed `status`, `code`, `message`, `retryable`, and `details`, with the request's public `request_id`. After that response, the fault MUST be consumed and the same request match MUST reach the ordinary route handler unless another fault is armed.

**TH-HARNESS-REQ-454**
At most one public-error fault may be armed per harness-owned runtime. A request to arm a second fault while one is pending MUST fail before replacing the pending fault with HTTP `409`, `error.code=test_public_error_fault_already_armed`. Runtime reset MUST clear any pending fault before reset success is accepted. A consumed fault MUST be removed before the fault response is written, so a retry of the same ordinary request reaches the ordinary route handler unless another fault has been armed.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-035

### 12.2.5 Network Flow Fault Control

**TH-HARNESS-REQ-455**
`POST /api/v1/test/runtime/network-flow-faults` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm a one-shot Network Flow commit or worker fault for a named harness boundary consumed by Network Flow tests. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any fault.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-456**
Network Flow fault boundaries are closed tokens owned by the harness. They identify where an opted-in Network Flow test implementation checks for an armed fault; they do not define product state-machine semantics. The supported boundary tokens are exactly:

| Boundary token                                                     | Harness use |
| ------------------------------------------------------------------ | ----------- |
| `network_flow.import.before_owner_prepare`                         | Fault before Network Flow table or diagnostic state is prepared for the shared unit of work. |
| `network_flow.import.after_owner_prepare`                          | Fault after Network Flow owner state has been prepared but before later participants are prepared. |
| `network_flow.import.after_indicator_prepare`                      | Fault after indicator create/dedupe or binding participants have prepared their writes. |
| `network_flow.import.after_audit_prepare`                          | Fault after transactional audit occurrences have been prepared. |
| `network_flow.import.after_idempotency_prepare`                    | Fault after idempotency-success state has been prepared. |
| `network_flow.import.after_terminal_publication_prepare`           | Fault after terminal import-result publication has been prepared. |
| `network_flow.import.before_transaction_commit`                    | Fault immediately before the shared transaction commit. |
| `network_flow.import.after_transaction_commit_before_reply`        | Fault after the shared transaction commits but before the apply caller receives the terminal response. |
| `network_flow.worker.before_handler_start`                         | Fault before a durable Network Flow worker handler starts owner work. |
| `network_flow.worker.before_apply_start`                           | Fault before the worker starts an apply attempt. |
| `network_flow.worker.before_cancellation_check`                    | Fault before the worker observes a cancellation gate. |
| `network_flow.worker.before_final_commit`                          | Fault before the worker reaches the final shared transaction commit. |
| `network_flow.worker.after_final_commit_before_terminal_publication` | Fault after final commit and before terminal-result publication or recovery reconciliation. |
| `network_flow.worker.after_terminal_publication_before_ack`        | Fault after terminal publication and before the worker acknowledges durable completion. |
| `network_flow.worker.before_replay_reconciliation`                 | Fault before worker recovery reconciles an already-committed operation with terminal publication. |

Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-457**
The request body MUST be a JSON object with exactly the fields below.

| Field             | Required | Behavior |
| ----------------- | -------- | -------- |
| `boundary`        | yes      | One of the closed Network Flow boundary tokens in TH-HARNESS-REQ-456. |
| `fault_kind`      | yes      | One of `return_error`, `panic`, `cancel_context`, `worker_crash`, or `worker_cancel`; worker-only kinds are accepted only with `network_flow.worker.*` boundaries. |
| `error_code`      | conditional | Required only for `fault_kind="return_error"`; a lowercase safe diagnostic token matching `^[a-z][a-z0-9_]{1,127}$`. |
| `correlation_key` | no       | Optional ASCII token matching `^[A-Za-z0-9._:-]{1,128}$`; when supplied, consumption requires the same key. |
| `consume_once`    | yes      | Must be `true`; persistent or multi-consume faults are not accepted. |

Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported boundary, unsupported fault kind, worker-only fault kind on an import boundary, invalid or misplaced `error_code`, invalid `correlation_key`, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_network_flow_fault_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_fault_control.v1` in the standard success envelope. The response MUST include a generated `fault_id`, exact `boundary`, exact `fault_kind`, optional `error_code`, optional `correlation_key`, and `consume_once=true`. The response MUST NOT include the test-route token, configured origins, cookies, product session credentials, database credentials, object-store credentials, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-458**
The armed Network Flow fault is in-memory harness runtime state and is consumed only by an opted-in Network Flow test implementation at the exact named boundary. A boundary mismatch MUST leave the fault pending. If the fault has a `correlation_key`, a missing or different consumer key MUST leave the fault pending. A consumed fault MUST be removed before applying its effect so retry, replay, or recovery reaches ordinary behavior unless another fault is armed. If no Network Flow fault registry is registered in the harness-owned runtime, boundary checks MUST be no-ops.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-459**
At most one Network Flow fault may be armed per harness-owned runtime. A request to arm a second fault while one is pending MUST fail before replacing the pending fault with HTTP `409`, `error.code=test_network_flow_fault_already_armed`. Runtime reset MUST clear any pending registered Network Flow fault before reset success is accepted. Network Flow fault controls MAY be used to prove all-or-nothing final commit, worker crash, cancellation, terminal-publication replay, and recovery behavior only when the executed fixture also cites the adopted product owner requirement that defines the expected state.
Verified by: TH-HARNESS-AC-050

### 12.2.6 Network Flow Deterministic Randomness Control

**TH-HARNESS-REQ-465**
`POST /api/v1/test/runtime/network-flow-randomness` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm one deterministic random stream for opted-in Network Flow fixture code that needs repeatable IDs, nonces, key IDs, digest salts, or intentional collision values. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any stream.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-466**
Network Flow deterministic-randomness streams are closed harness tokens. They identify fixture-only injection points and do not define product identity semantics, digest algorithms, cursor algorithms, Graph Projection semantics, or public API compatibility. The supported stream tokens are exactly:

| Stream token                             | Harness use |
| ---------------------------------------- | ----------- |
| `network_flow.table_id`                  | Deterministic table identity and table-name collision fixtures. |
| `network_flow.row_id`                    | Deterministic row identity fixtures. |
| `network_flow.diagnostic_id`             | Deterministic diagnostic identity and diagnostic-order fixtures. |
| `network_flow.import_job_id`             | Deterministic import/apply job identity fixtures. |
| `network_flow.import_source_ref`         | Deterministic opaque import-source reference fixtures. |
| `network_flow.cursor_nonce`              | Deterministic cursor nonce, replay, TTL, and rotation fixtures. |
| `network_flow.safe_digest_nonce`         | Deterministic safe-digest salt or key-bound comparison fixtures without exposing production secrets. |
| `network_flow.graph_invocation_id`       | Deterministic ephemeral Graph Projection invocation fixtures. |

Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-467**
The request body MUST be a JSON object with exactly the fields below.

| Field          | Required | Behavior |
| -------------- | -------- | -------- |
| `stream`       | yes      | One of the closed Network Flow stream tokens in TH-HARNESS-REQ-466. |
| `value_kind`   | yes      | One of `uuid`, `token`, or `hex_bytes`. |
| `values`       | yes      | Ordered deterministic values, length `1..256`; duplicate values are allowed only to exercise collision behavior. |
| `consume_once` | yes      | Must be `true`; persistent or multi-consume values are not accepted. |
| `exhaustion`   | yes      | Must be `fail_closed`; an armed stream exhausted by fixture code MUST fail the fixture rather than silently falling back to production randomness. |

For `value_kind="uuid"`, each value MUST be canonical lowercase UUID text. For `value_kind="token"`, each value MUST match `^[A-Za-z0-9._:-]{1,128}$`. For `value_kind="hex_bytes"`, each value MUST be lowercase, even-length hex text no longer than 512 characters.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-468**
Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported stream, unsupported value kind, an empty or oversized `values` array, a value that does not match the selected `value_kind`, `consume_once` other than `true`, or `exhaustion` other than `fail_closed` MUST fail with `400`, `error.code=invalid_network_flow_randomness_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_randomness_control.v1` in the standard success envelope. The response MUST include a generated `control_id`, exact `stream`, exact `value_kind`, `value_count`, `remaining_count`, `consume_once=true`, and `exhaustion="fail_closed"`. The response MUST NOT include deterministic values, the test-route token, configured origins, cookies, product session credentials, database credentials, object-store credentials, production secret material, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-469**
At most one deterministic-randomness sequence may be armed for a stream in a harness-owned runtime. A request to arm a second sequence for the same stream while the stream is registered, including after all values have been consumed, MUST fail before replacing the registered sequence with HTTP `409`, `error.code=test_network_flow_random_stream_already_armed`. Consuming a stream MUST return values in request order exactly once. A missing stream MUST be a no-op for opted-in consumers, but an armed stream that is exhausted or consumed with the wrong `value_kind` MUST fail closed. Runtime reset MUST clear any registered Network Flow deterministic-randomness streams before reset success is accepted.
Verified by: TH-HARNESS-AC-052

### 12.2.7 Network Flow Authorization-Transition Control

**TH-HARNESS-REQ-470**
`POST /api/v1/test/runtime/network-flow-auth-transitions` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm one fixture-only authorization-transition control for opted-in Network Flow tests that need a route-time authorization change, hidden-resource assertion, cursor authorization recheck, or extension-resource invalidation trigger. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any transition.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-471**
Network Flow authorization-transition boundaries, transition kinds, resource kinds, and hidden-response kinds are closed harness tokens. They identify fixture-only injection and assertion points and do not define product authorization semantics, incident membership rules, hidden-resource status codes, cursor algorithms, WebSocket contracts, or public API compatibility. The supported boundary tokens are exactly `network_flow.route.before_authorization`, `network_flow.route.after_authorization_before_lookup`, `network_flow.route.after_lookup_before_response`, `network_flow.cursor.before_authorization_recheck`, `network_flow.websocket.before_invalidation_publish`, and `network_flow.fixture.after_transition`. The supported `transition_kind` values are exactly `incident_membership_revoked`, `incident_membership_restored`, `incident_soft_deleted`, `network_flow_table_soft_deleted`, `network_flow_table_renamed`, `session_revoked`, and `extension_claim_removed`. The supported `resource_kind` values are exactly `incident`, `network_flow_table`, `network_flow_cursor`, `network_flow_graph`, `network_flow_contributors`, and `network_flow_workspace`. The supported `hidden_response_kind` values are exactly `not_found`, `forbidden_without_resource`, `empty_collection`, `cursor_rejected`, `extension_profile_not_claimed`, and `invalidation_event`.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-472**
The request body MUST be a JSON object with exactly the fields below.

| Field                        | Required | Behavior |
| ---------------------------- | -------- | -------- |
| `boundary`                   | yes      | One of the closed Network Flow auth-transition boundary tokens in TH-HARNESS-REQ-471. |
| `transition_kind`            | yes      | One of the closed transition kinds in TH-HARNESS-REQ-471. |
| `actor_ref`                  | yes      | Safe fixture actor reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `incident_ref`               | yes      | Safe fixture incident reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `resource_kind`              | yes      | One of the closed resource kinds in TH-HARNESS-REQ-471. |
| `resource_ref`               | yes      | Safe fixture resource reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `hidden_response_kind`       | yes      | One of the closed hidden-response kinds in TH-HARNESS-REQ-471. |
| `must_not_disclose_resource` | yes      | Must be `true`; controls that allow resource disclosure are not accepted. |
| `correlation_key`            | no       | Optional safe fixture correlation key matching `^[A-Za-z0-9._:-]{1,128}$`; when supplied, consumption requires the same key. |
| `consume_once`               | yes      | Must be `true`; persistent or multi-consume transitions are not accepted. |

Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-473**
Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported boundary, unsupported transition kind, unsupported resource kind, unsupported hidden-response kind, unsafe refs, `must_not_disclose_resource` other than `true`, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_network_flow_auth_transition_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_auth_transition_control.v1` in the standard success envelope. The response MUST include a generated `control_id`, exact boundary, transition kind, actor ref, incident ref, resource kind/ref, hidden response kind, optional correlation key, `must_not_disclose_resource=true`, and `consume_once=true`. The response MUST NOT include product session credentials, membership row IDs, role grants, route-token material, raw hidden resource details, database credentials, object-store credentials, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-474**
An armed Network Flow auth-transition control is in-memory harness runtime state and is consumed only by an opted-in Network Flow test implementation at the exact boundary, actor ref, incident ref, resource ref, and optional correlation key. A mismatch MUST leave the control pending. A consumed control MUST be removed before the fixture applies the transition or hidden-resource assertion, so retry/replay reaches ordinary behavior unless another control has been armed. A request to arm a duplicate exact boundary/actor/incident/resource tuple while one is pending MUST fail before replacement with HTTP `409`, `error.code=test_network_flow_auth_transition_already_armed`; independent tuples MAY be armed concurrently. Runtime reset MUST clear any registered Network Flow auth-transition controls before reset success is accepted.
Verified by: TH-HARNESS-AC-053

### 12.2.8 Network Flow Audit Assertion Control

**TH-HARNESS-REQ-475**
`POST /api/v1/test/runtime/network-flow-audit-assertions` is a harness test route with the same enablement, host/origin, and token authorization predicates as the reset route. The route MAY arm one fixture-only audit assertion for opted-in Network Flow tests that need exact domain audit occurrence counts, zero-occurrence cases, or no-additional-occurrence replay checks. It MUST NOT be exposed as production API behavior, MUST NOT be listed in production OpenAPI, MUST NOT accept ordinary session, role, CSRF, bearer, bootstrap-token, or `deployment_admin` authorization as a substitute for the test-route token, and MUST fail host/origin or token checks before decoding the body or arming any assertion.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-476**
Network Flow audit assertion kinds, event codes, and resource kinds are closed harness tokens. They identify fixture-only assertion points and do not define product audit semantics, audit storage, operation authorization, event payload shape, or public API compatibility. The supported `assertion_kind` values are exactly `exact_count`, `zero_occurrences`, and `no_audit_replay`. The supported `event_code` values are exactly `network_flow_table_created`, `network_flow_table_renamed`, `network_flow_table_soft_deleted`, `network_flow_graph_query_executed`, `network_flow_indicator_binding_created`, and `network_flow_indicator_binding_reused`. The supported `resource_kind` values are exactly `network_flow_table`, `network_flow_graph`, `network_flow_indicator_binding`, and `network_flow_import`.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-477**
The request body MUST be a JSON object with exactly the fields below.

| Field                       | Required | Behavior |
| --------------------------- | -------- | -------- |
| `assertion_kind`            | yes      | One of the closed assertion kinds in TH-HARNESS-REQ-476. |
| `event_code`                | yes      | One of the closed Network Flow audit event codes in TH-HARNESS-REQ-476. |
| `operation_ref`             | yes      | Safe fixture operation reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `actor_ref`                 | yes      | Safe fixture actor reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `incident_ref`              | yes      | Safe fixture incident reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `resource_kind`             | yes      | One of the closed resource kinds in TH-HARNESS-REQ-476. |
| `resource_ref`              | yes      | Safe fixture resource reference matching `^[A-Za-z0-9._:-]{1,128}$`. |
| `baseline_count`            | yes      | Non-negative count before the exercised product operation, maximum `1000000`. |
| `expected_final_count`      | yes      | Exact count expected after the exercised product operation, maximum `1000000`. |
| `expected_replay_increment` | yes      | Exact additional occurrence count expected from committed replay, maximum `1000000`; no-audit replay assertions require `0`. |
| `correlation_key`           | no       | Optional safe fixture correlation key matching `^[A-Za-z0-9._:-]{1,128}$`; when supplied, consumption requires the same key. |
| `consume_once`              | yes      | Must be `true`; persistent or multi-consume assertions are not accepted. |

`expected_final_count` MUST be greater than or equal to `baseline_count`. For `assertion_kind="zero_occurrences"`, `baseline_count`, `expected_final_count`, and `expected_replay_increment` MUST all be `0`. For `assertion_kind="no_audit_replay"`, `expected_replay_increment` MUST be `0`.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-478**
Unknown members, missing required fields, non-object JSON, invalid JSON, unsupported assertion kind, unsupported event code, unsupported resource kind, unsafe refs, negative or oversized counts, `expected_final_count < baseline_count`, assertion-kind count-rule violations, or `consume_once` other than `true` MUST fail with `400`, `error.code=invalid_network_flow_audit_assertion_request`. Successful arming MUST return HTTP `201` with `cartulary.test.network_flow_audit_assertion_control.v1` in the standard success envelope. The response MUST include a generated `assertion_id`, exact assertion kind, event code, operation ref, actor ref, incident ref, resource kind/ref, baseline count, expected final count, expected replay increment, optional correlation key, and `consume_once=true`. The response MUST NOT include product session credentials, raw audit payloads, raw source data, safe-digest key material, cursor tokens, membership row IDs, role grants, database credentials, object-store credentials, raw fixture source paths, or private runtime state.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-479**
An armed Network Flow audit assertion is in-memory harness runtime state and is consumed only by an opted-in Network Flow test implementation at the exact event code, operation ref, resource kind/ref, and optional correlation key. A mismatch MUST leave the assertion pending. A consumed assertion MUST be removed before the fixture compares the observed product audit counts, so retry/replay reaches ordinary behavior unless another assertion has been armed. A request to arm a duplicate exact event/operation/resource tuple while one is pending MUST fail before replacement with HTTP `409`, `error.code=test_network_flow_audit_assertion_already_armed`; independent tuples MAY be armed concurrently. Runtime reset MUST clear any registered Network Flow audit assertions before reset success is accepted.
Verified by: TH-HARNESS-AC-054

### 12.3 Runtime Reset Request Body

| Body                                | Behavior                                        |
| ----------------------------------- | ----------------------------------------------- |
| No body                             | Accepted.                                       |
| `{}` JSON object                    | Accepted.                                       |
| JSON object with any members        | `400`, `error.code=invalid_test_reset_request`. |
| Non-object JSON                     | `400`, `error.code=invalid_test_reset_request`. |
| Invalid JSON with JSON content type | `400`, `error.code=invalid_test_reset_request`. |

### 12.4 Saved-View System Fixture Request Body

The saved-view system fixture request body MUST be a JSON object with exactly the fixture fields below.

| Field            | Required | Behavior                                                                                                                                 |
| ---------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `view_schema_id` | yes      | Stable workbook `view_schema_id`; unknown or empty values fail with `400`, `error.code=invalid_mutation_payload`.                       |
| `display_name`   | yes      | Saved-view display name; normalized with ordinary saved-view display-name normalization.                                                   |
| `query_json`     | yes      | Persisted saved-view query JSON; normalized with ordinary saved-view persisted-query normalization for the selected `view_schema_id`.      |
| `layout_json`    | no       | Saved-view layout JSON; omitted values receive the ordinary normalized default layout for the selected `view_schema_id`.                  |

Unknown members, including `scope`, `owner_user_id`, `saved_view_id`, `incident_id`, `created_at`, `updated_at`, and `saved_view_version`, MUST fail with `400`, `error.code=invalid_mutation_payload`. Missing required fields, non-object JSON, invalid JSON, invalid `view_schema_id`, invalid `display_name`, invalid persisted query shape, or invalid layout shape MUST fail through the ordinary saved-view mutation error envelope with `400`, `error.code=invalid_mutation_payload`, and field/reason details when available.

Successful fixture creation MUST return a saved-view resource with `scope='system'`, `owner_user_id=null`, the path `incident_id`, a generated `saved_view_id`, normalized `display_name`, normalized `query_json`, normalized `layout_json`, and store-managed timestamps/version. The returned resource MAY be visible through ordinary saved-view list behavior because `scope='system'` is visible fixture data; that visibility does not make the fixture route a production create API and does not change the ordinary saved-view create rule that rejects `scope='system'`.

### 12.5 Runtime Reset Concurrency and Timeout

| Condition            | Behavior                                                                                    |
| -------------------- | ------------------------------------------------------------------------------------------- |
| No reset active      | Acquire reset lock and run reset.                                                           |
| Reset already active | `409`, `error.code=test_runtime_reset_in_progress`.                                         |
| Reset exceeds `30s`  | `503`, `error.code=test_runtime_reset_timeout`; response includes failed action when known. |

### 12.6 Runtime Reset Algorithm and Partial Failure

**TH-HARNESS-REQ-452**
The reset route MUST preserve migration metadata, restore the active deployment admin, truncate mutable public-schema runtime state, clear route idempotency state, clear registered in-memory test-clock state, clear in-memory public-error fault state, clear registered in-memory Network Flow fault state, clear registered in-memory Network Flow deterministic-randomness state, clear registered in-memory Network Flow auth-transition state, clear registered in-memory Network Flow audit-assertion state, and clear the configured object store bucket or prefix for the harness-owned runtime.

The database reset table set is selected by this algorithm:

```text
select_reset_tables(database):
  query information_schema.tables
  keep rows where table_schema = "public"
  keep rows where table_type = "BASE TABLE"
  reject table_name = "goose_db_version"
  order table_name ascending
  return table_name list
```

The reset MUST execute `TRUNCATE TABLE public.<table> ... RESTART IDENTITY CASCADE` for the selected table list inside one database transaction, using identifier-safe table quoting. If the selected table list is empty, the truncate step is a successful no-op and the bootstrap restoration step still runs. After truncate, bootstrap restoration MUST restore exactly one active deployment admin and exactly one bootstrap marker before commit. The `goose_db_version` row count before reset MUST equal the count after reset and MUST be nonzero. The `route_idempotency` row count after reset MUST be `0`.

Database table truncation and bootstrap restoration MUST execute in one database transaction when the database supports that transaction shape. Object-store deletion occurs after the database transaction commits. The route MUST NOT claim rollback across object-store deletion.

The reset response MUST include `tables_reset` in sorted order, post-reset counts, `object_count_after`, `partial_failure`, and `failed_action` when a failed action exists. `tables_reset` MUST be the exact output of `select_reset_tables(database)` for the reset attempt.

| Surface | Selection rule | Mutation | Success proof | Failure status/code |
| --- | --- | --- | --- | --- |
| Migration metadata | `public.goose_db_version` only | Preserve; never truncate | before and after counts equal and nonzero | `500`, `error.code=test_runtime_reset_failed` |
| Public mutable DB tables | `select_reset_tables(database)` | Truncate with `RESTART IDENTITY CASCADE` inside the reset transaction | `tables_reset` sorted and post-reset mutable counts are zero except bootstrap-restored rows | `500`, `error.code=test_runtime_reset_failed` |
| Bootstrap admin state | Ordinary bootstrap preflight inside the reset transaction | Restore active deployment admin and bootstrap marker | exactly one active deployment admin and exactly one bootstrap marker | `500`, `error.code=test_runtime_reset_failed` |
| Route idempotency | `public.route_idempotency` when selected by the table algorithm | Truncate with other mutable tables | post-reset `route_idempotency` count equals `0` | `500`, `error.code=test_runtime_reset_failed` |
| Public-error fault state | In-memory harness runtime fault slot | Clear pending fault before success is accepted | no pending fault remains after reset | `500`, `error.code=test_runtime_reset_failed` when clear hook fails |
| Configured object-store bucket or prefix | Harness-owned object-store configuration for the runtime | Delete every object after DB transaction commit | `object_count_after=0` | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true` |

| Failure point                                      | Required response                                                                                                  |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Before DB transaction commit                       | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=false` unless prior mutation occurred.             |
| After DB commit and before object cleanup complete | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true`.                                             |
| Object cleanup partial deletion                    | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true`, include `object_count_after` if measurable. |
| Bootstrap admin not restored                       | `500`, `error.code=test_runtime_reset_failed`, `partial_failure=true`.                                             |

A browser wrapper receiving `partial_failure=true` MUST mark the owned stack tainted and restart it before further browser child work.
Verified by: TH-HARNESS-AC-008, TH-HARNESS-AC-034

### 12.7 Runtime Reset Success Readiness

Success requires all of the following:

- migration metadata preserved;
- active deployment admin restored;
- mutable incident/product tables empty;
- route idempotency table empty;
- object count after reset equals `0`;
- response validates against `cartulary.test.runtime_reset.v1`.

## 13. Cleanup and Destructive Safety

**TH-HARNESS-REQ-500**
Cleanup is destructive. Cleanup commands MUST delete only paths or resources satisfying the exact ownership predicates in this section.
Verified by: TH-HARNESS-AC-009, TH-HARNESS-AC-010

`make clean` and `make distclean` are repo-local cleanup commands. They MUST NOT remove caller-supplied external result roots and MUST NOT stop local Compose services. Local service teardown belongs to `make services-down`, not to repo-local cleanup.
Frontend dependency install state is a coupled repo-local artifact set. `make clean` MUST preserve installed dependency roots for local loop speed, while `make distclean` MUST remove the repo-local pnpm store, frontend install stamps, and root/workspace `node_modules` directories together so stale package-manager metadata cannot survive without its store.
Harness cache state is repo-local acceleration state. `make clean` MUST preserve default `.cache/cartulary/*` cache roots so ordinary cleanup does not erase warm local readiness. `make distclean` MUST remove default cache roots under `.cache/cartulary/`, including readiness, build-artifact, agent-finalize action-cache, and execution-topology render cache roots. Neither cleanup target may remove caller-supplied cache directories outside the repository.
Fallow configs, rule packs, schemas, and any future reviewed baselines under repository source/tool roots are harness inputs, not cleanup-owned artifacts. Fallow run-root outputs are ordinary retained artifacts and may be removed only through result-root cleanup predicates.

### 13.1 Path Algorithm

```text
normalize_cleanup_candidate(path):
  reject empty string
  reject NUL
  reject "/"
  reject "."
  reject ".."
  reject any caller-supplied segment equal to ".."
  reject backslash on POSIX conformance hosts
  resolve relative paths against repository root
  reject absolute paths outside repository root
  reject protected repository roots named in the table below when they are named as cleanup candidates
  lstat path
  if path is symlink:
    unlink symlink object only
    MUST NOT follow target
  if path is directory:
    remove directory tree only after every traversed entry remains under the candidate root by lexical path and lstat traversal
```

The protected repository root set is closed in the current profile:

| Protected root | Protection rule |
| --- | --- |
| `.git` | Reject when named directly as a cleanup candidate. |
| `docs` | Reject when named directly as a cleanup candidate. |
| `cmd` | Reject when named directly as a cleanup candidate. |
| `internal` | Reject when named directly as a cleanup candidate. |
| `apps` | Reject when named directly as a cleanup candidate. |
| `packages` | Reject when named directly as a cleanup candidate. |
| `contracts` | Reject when named directly as a cleanup candidate. |
| `db/migrations` | Reject when named directly as a cleanup candidate. |
| `db/queries` | Reject when named directly as a cleanup candidate. |
| `configs` | Reject when named directly as a cleanup candidate. |
| `scripts` | Reject when named directly as a cleanup candidate. |
| `tools` | Reject when named directly as a cleanup candidate. |
| `go.mod` | Reject when named directly as a cleanup candidate. |
| `go.sum` | Reject when named directly as a cleanup candidate. |
| `package.json` | Reject when named directly as a cleanup candidate. |
| `pnpm-lock.yaml` | Reject when named directly as a cleanup candidate. |
| `pnpm-workspace.yaml` | Reject when named directly as a cleanup candidate. |

A child path under a protected root MAY be removed only when Section 13.2 or another adopted cleanup table explicitly lists that exact path or path family as cleanup-owned. Missing cleanup-owned paths are successful no-ops. A path that is both protected and cleanup-owned MUST use the narrower cleanup-owned row; broad ancestor deletion remains rejected.

### 13.2 Cleanup Scope

| Command               |      Removes default result root? | Removes custom `CARTULARY_TEST_RESULTS_DIR`? | Removes default `.cache/cartulary` cache roots? | Removes external Go caches? | Stops Docker/Compose globally? |
| --------------------- | --------------------------------: | -------------------------------------------: | ---------------------------------------------: | --------------------------: | -----------------------------: |
| `make clean`          | yes, only default registered path |                                           no |                                             no |                          no |                             no |
| `make distclean`      | yes, only default registered path |                                           no |                                            yes |                          no |                             no |
| `make services-down`  |                                no |                                           no |                                             no |                          no | no; stops only this repo's local Compose services and preserves named volumes |
| Service-suite cleanup |        only suite-owned artifacts |                                           no |                                             no |                          no |                             no |
| Stale janitor         |        proof-gated resources only |                                           no |                                             no |                          no |                             no |

`make distclean` owns removal of `.pnpm-store`, the repository-root `node_modules` directory, workspace package `node_modules` directories under `apps/web` and `packages/*`, and default repo-local cache roots under `.cache/cartulary/`. Missing workspace dependency roots or cache roots are not cleanup failures.

### 13.2.1 Local Service And Data Reset Scope

`make services-down` MUST stop only the local Compose services declared for repository development and MUST preserve named volumes. It MUST NOT pass a Compose volume-removal flag. `db-down` is not a current public command binding; new and existing automation MUST use `services-down`.

`make db-migrate` MUST apply the repository migration line to the local development database without dropping, recreating, or truncating that database and without resetting, deleting, initializing, or inspecting object storage. It MAY start Postgres to perform the migration. It MUST use the same `migrate up` application surface as deployable migration execution and MUST surface migration-remediation reports without rewriting them. It MUST NOT overwrite an inherited managed Postgres DSN selected by the config's database `service_ref`; for the default local development service it MAY derive a local Compose DSN only when the selected DSN environment value is unset.

`make db-reset` MUST recreate only the local development database and rerun migrations. It MAY start Postgres to perform the reset, but it MUST NOT reset, delete, or inspect object storage. A real `db-reset` MUST reject before Compose, database, migration, or object-store commands unless `CARTULARY_DESTRUCTIVE_CONFIRM=db-reset` was supplied on the Make command line.

`make dev` MUST start the backend process and prove backend readiness before starting the frontend process. If backend readiness fails because the database is behind the current line, the dev-stack diagnostic MUST direct the caller to `make db-migrate`; if the backend log reports `prod_ddl_rebaseline_v1` with `historical_migration_lineage`, the diagnostic MUST direct the caller to reset the local database or use an owner-approved export/import path. The frontend process MUST NOT start after a backend readiness failure.

`make object-store-reset` MUST clear only objects in the configured local object-store bucket and MUST leave the bucket present afterward. In the current implementation profile the local object store is SeaweedFS S3, and the public command and command ID are provider-neutral. A real `object-store-reset` MUST reject before Compose or object-store commands unless `CARTULARY_DESTRUCTIVE_CONFIRM=object-store-reset` was supplied on the Make command line.

### 13.3 Stale Janitor Thresholds

| Resource        | Completed-run predicate                                         | Uncompleted stale predicate                | Active-resource rule                                                                     |
| --------------- | --------------------------------------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------- |
| Database        | Completed summary or lease cleanup state older than 15 minutes. | Lease or metadata older than 24 hours.     | Active connections may be terminated only after proof predicate passes.                  |
| Bucket          | Completed summary or lease cleanup state older than 15 minutes. | Lease or metadata older than 24 hours.     | Delete only generated bucket/prefix with proof metadata.                                 |
| Container       | Completed summary or lease cleanup state older than 15 minutes. | Harness Docker labels older than 24 hours. | Running container may be stopped only if proof predicate passes and label owner matches. |
| Browser fixture | Completed target summary older than 15 minutes.                 | Fixture metadata older than 6 hours.       | Delete only generated fixture directory with ownership metadata.                         |
| Browser process/session | Completed session lease older than 15 minutes.          | Session lease older than 6 hours with matching runtime-root marker and process command/env proof. | Running processes may be stopped only when PGID, runtime root, command/env proof, and lease identity all match; a port listener alone is never sufficient proof. |

For container cleanup, an already-deleting Docker resource is treated as deferred successful cleanup only after the same proof predicates pass. This compatibility rule exists to make repeated service-backed public targets reproducible under Docker's asynchronous removal lifecycle; it MUST NOT broaden cleanup authority to unlabelled, current-suite, or externally owned containers.

### 13.4 Dry-Run Contract

| Setting                                        | Behavior                                                                                                                    |
| ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| `CARTULARY_CLEANUP_DRY_RUN` omitted or not `1` | Cleanup or reset may delete resources satisfying predicates and confirmation rules.                                         |
| `CARTULARY_CLEANUP_DRY_RUN=1`                  | Cleanup MUST emit deletion candidates and reasons and MUST NOT start services, stop services, delete files, delete DBs, delete bucket objects, delete buckets, delete containers, or delete browser fixtures. |

Dry-run output MUST include normalized path or resource identity, proof predicate, action that would be taken, and rejection reason for retained candidates. For human destructive targets, the dry-run line format is:

```text
DRY-RUN <action> <normalized-identity> <proof-or-rejection-reason>
```

`CARTULARY_DESTRUCTIVE_CONFIRM` is ignored for dry-runs. Inherited environment values MUST NOT satisfy the confirmation predicate for public Make targets; only Make command-line values are valid confirmation sources.

### 13.5 Parent-Death and Reaper Rule

Immediate cleanup after parent death is not guaranteed. The conformance guarantee is:

- owned resources carry enough lease or proof metadata for later stale janitor evaluation;
- detached reaper scheduling is optional unless the command declares it;
- if a detached reaper is scheduled, it writes `reaper-scheduled.json` with lease ID, started-at timestamp, target resources, and timeout.

## 14. Platform and CI Support

**TH-HARNESS-REQ-550**
The current conformance support matrix is closed by this section. A target may be run elsewhere, but unsupported environments MUST NOT be used for current harness conformance claims.
Verified by: TH-HARNESS-AC-012

| Environment/tool                                   | Current conformance status                 | Required evidence                                                                                                                             |
| -------------------------------------------------- | ------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Linux x86_64 with Docker Engine and Docker Compose | required                                   | Full acceptance matrix.                                                                                                                       |
| WSL2 Ubuntu with Docker Desktop integration        | supported compatibility profile            | `doctor`, `test-fast`, `check`, browser E2E smoke.                                                                                            |
| macOS                                              | unsupported for current conformance        | None.                                                                                                                                         |
| Windows native                                     | unsupported                                | None; use WSL2 profile.                                                                                                                       |
| Hosted CI provider                                 | provider-neutral only                      | `make ci`; no annotation/upload claims.                                                                                                       |
| Podman/Podman Compose                              | unsupported                                | None.                                                                                                                                         |
| Docker                                             | required for service-backed targets        | Missing Docker yields `preflight_error`.                                                                                                      |
| Docker Compose                                     | required for local Compose targets         | Missing Compose yields `preflight_error` when Compose is absent and `service_readiness_timeout` when Compose-started service readiness fails. |
| Go/Node/pnpm/Playwright/Bash utilities             | required as pinned by repository procedure | Version mismatch yields `configuration_error`.                                                                                                |
| Linux inotify capacity                             | required only for Vite dev surfaces        | Low watcher limits or exhausted watcher usage MUST be diagnosed before Vite dev startup; release/browser E2E preview paths MUST NOT require this preflight. |

Harness diagnostics MAY report Linux inotify `max_user_watches`, `max_user_instances`, best-effort current watcher usage, and bounded operator diagnostics. The harness MUST NOT mutate host sysctl settings.

## 15. Security and Redaction

**TH-HARNESS-REQ-600**
Centralized summaries, machine output, and retained logs captured by harness wrappers MUST be redacted before retention and before stdout emission.
Verified by: TH-HARNESS-AC-011

**TH-HARNESS-REQ-601**
Redaction MUST be applied to captured stdout, stderr, wrapper diagnostics, machine JSON, retained logs, service env dumps, and summary artifacts before those bytes are written outside a private runtime working file or emitted to stdout/stderr. A redaction failure MUST fail the public target with `failure_class=artifact`, `failure_reason=artifact_error`, and public exit code `11` unless an earlier primary failure is preserved by Section 9.1.
Verified by: TH-HARNESS-AC-011

**TH-HARNESS-REQ-602**
The redaction algorithm MUST apply to both keys and values after decoding structured JSON where possible and to raw text otherwise. Matching MUST be case-insensitive for key names and header names. At minimum, the algorithm MUST redact:

- variables, JSON keys, HTTP headers, and CLI arguments whose names match the secret pattern table;
- URL userinfo and DSN password segments;
- bearer-token, session-cookie, JWT, private-key, and object-store credential forms in raw text;
- `Authorization`, `Cookie`, `Set-Cookie`, and `X-Cartulary-Test-Route-Token` values.
Verified by: TH-HARNESS-AC-011

Structured redaction MUST preserve schema-owned container shapes and scalar types unless the value itself is secret. Object and array fields such as `service_sessions`, `browser_stage_sessions`, `session_target`, `cleanup_status`, `lease_file`, and timing fields MUST NOT be replaced merely because their names contain a secret-related substring. Secret key matching MUST use exact or anchored credential-name patterns rather than broad substring matching that can redact structural diagnostics.
Verified by: TH-HARNESS-AC-000, TH-HARNESS-AC-011

**TH-HARNESS-REQ-604**
Structured secret-key matching is closed. Before comparing a structured key name, the redactor MUST uppercase it and replace every non-ASCII-alphanumeric run with one `_`, then trim leading and trailing `_`. The resulting token is secret-bearing only when it equals or has an anchored credential suffix/prefix equivalent to one of: `PASSWORD`, `PASS`, `PWD`, `TOKEN`, `JWT`, `BEARER`, `API_KEY`, `ACCESS_KEY`, `SECRET_KEY`, `AUTHORIZATION`, `COOKIE`, `SET_COOKIE`, or `X_CARTULARY_TEST_ROUTE_TOKEN`. Substring-only matches such as `session_target` containing `token` across token boundaries MUST NOT redact the field.

Raw-text redaction MUST apply after structured redaction to these closed families: URL userinfo (`scheme://userinfo@host`), PostgreSQL-style DSN password segments (`password=` or `:password@` credential forms), bearer authorization headers, JWT-like three-part base64url tokens, PEM private-key blocks, and S3-compatible access-key or secret-key assignments. Structured redaction MUST preserve object and array shapes and preserve numeric, boolean, and null scalar types unless that scalar value itself is secret. A redaction write or validation failure maps to `failure_class=artifact`, `failure_reason=artifact_error`, and public exit `11` unless Section 9.1 preserves an earlier primary failure.
Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-036

**TH-HARNESS-REQ-605**
Runtime and support security scans MUST derive their default support-root
exclusions and support package patterns from the schema-validated test-support
inventory required by TH-HARNESS-REQ-067. Every registered Go support root
MUST be included in the support scan. A root MAY be excluded from the runtime
profile only when the inventory classifies it as test support; a package that
is compiled into a production binary, including a module-owned harness-control
package, MUST remain in the runtime profile.

An unknown `internal/**/testsupport` root, a registered root missing from the
support profile, a duplicate path, or a path excluded from both profiles MUST
fail before the security tool starts. Publicly declared security overrides
remain governed by the Section 5 input registry and MUST NOT silently replace
inventory validation.

Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-056

SeaweedFS strict release evidence MUST derive its redaction scan input set from
the current release-evidence run, the current `seaweedfs-compatibility` report,
and the current backend-process backup/restore row evidence. Strict release
targets MUST run `backend-process` and `seaweedfs-compatibility` as current-run
prerequisites. Stable copies, historical delivery-shaped paths, newest-run
fallback evidence, retained `services-up` reports, and retired source-migration
evidence MUST NOT satisfy the gate. Missing selected child artifacts are
blocking findings and MUST NOT be replaced with fallback evidence.

**TH-HARNESS-REQ-675**
The current profile contains no legacy-source object-storage transition product
behavior or release-support evidence. The authored task surface, verification
catalog, families, topology, occurrence policy, release evidence, and release
gates MUST contain no dedicated transition target, row, or replacement owner.
SeaweedFS release evidence instead covers
current S3 compatibility, backup integrity and restoration, operator-private
artifact redaction, and source-owner storage-reference preservation.
Historical result roots MAY contain retired migration artifacts but MUST NOT be
selected as current conformance or release evidence.
Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-015

The operational recovery smoke MUST build its uniquely tagged application image exactly once before Compose startup. Startup MUST consume that completed image without requesting a second build. Removing duplicate build orchestration MUST NOT change readiness, backup, restore-verification, route-absence, diagnostic, or cleanup behavior.
Verified by: TH-HARNESS-AC-011, TH-HARNESS-AC-015

**TH-HARNESS-REQ-603**
Retained run roots and target artifact directories MUST be created with owner-only permissions on POSIX conformance hosts unless the caller explicitly supplied a custom result root whose permissions cannot be narrowed without changing ownership. Required summary artifacts and retained logs MUST be written with owner-read/write permissions. A custom result root that is world-writable without the sticky bit, or that cannot protect newly created files from other users on the host, MUST fail preflight with `configuration_error`.
Verified by: TH-HARNESS-AC-003, TH-HARNESS-AC-011, TH-HARNESS-AC-015

Screenshots, videos, traces, visual geometry diagnostics, and Playwright HTML reports are diagnostic secret-bearing artifacts. They MUST NOT be described as safe to upload or publish without separate review. Browser visual targets MAY retain compact geometry diagnostics for workbook screenshot failures, including scroll metrics, visible field keys, required element rectangles, active element identity, and inspector state. Those diagnostics are harness mechanics only; they MUST NOT define product UI behavior or supplement the bounded visual-snapshot refresh authority in TH-HARNESS-REQ-255.

Workbook visual regression tests that capture an outer grid shell while driving an inner grid scrollport MUST normalize and verify both layers before assertion. The screenshot-target shell MUST be reset to `scrollLeft=0` and `scrollTop=0` for left/default viewport captures unless a test explicitly declares a different shell-scroll contract, while the owned grid scrollport MUST be normalized to the test's declared scroll or anchor state. Anchor-based captures that intentionally frame off-screen workbook columns are explicit shell-scroll contracts and MUST still reset stale shell state before computing their deterministic offset. The diagnostic record MUST identify the screenshot target and both shell and scrollport metrics; exact human wording is non-normative. This normalization is harness mechanics only; it does not promote refresh output into product conformance, design conformance, release, or Core 05 publication evidence.

**TH-HARNESS-REQ-606**
`tools/executable_input_policy.json` validates as
`cartulary.executable_input_policy.v1`. It declares closed restricted roots,
standalone documentation-maintenance sources, and machine evidence roots.
Product, catalog, security, release, conformance, and generation commands have
no documentation-read exception mechanism.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-607**
The executable-input boundary scans executable and validation sources plus
machine evidence configuration without opening, statting, resolving, hashing,
or enumerating a restricted documentation root. Direct and joined restricted
paths fail the gate. Verification contracts, test-family manifests, generated
artifacts, and production fixtures do not accept documentation references.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-608**
Boundary tests use neutral synthetic root names and files. Standalone
documentation maintenance may read documentation only through the explicitly
classified `lint-markdown` command, which is excluded from product, CI,
conformance, release, performance, and claim-publication aggregates.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-609**
The semantic-identity scan MUST detect delivery-shaped path components, filenames, symbols, titles, selector values, variables, schema IDs, artifact names, and target IDs case-insensitively. The implementation MAY keep closed detection patterns, but MUST NOT publish or consume a runtime exception registry.
Verified by: TH-HARNESS-AC-067, TH-HARNESS-AC-070

**TH-HARNESS-REQ-610**
Every semantic-identity violation in an executable source, live catalog selector, generated topology identity, public input, schema identity, harness/test path, or active fixture MUST fail the gate. No allowlist, path exception, compatibility classification, or line-number exception is permitted. Product vocabulary that legitimately uses the word “phase” without encoding delivery order remains outside the delivery-shaped patterns and is owned by its product specification.
Verified by: TH-HARNESS-AC-067, TH-HARNESS-AC-070

**TH-HARNESS-REQ-611**
An ambiguous semantic match blocks closure. The scanner MUST report the matched token and normalized location without opening or copying unrelated document content.
Verified by: TH-HARNESS-AC-067

**TH-HARNESS-REQ-612**
Harness observability may emit only the closed attributes and metrics in
Sections 8 and 10. It MUST NOT emit raw commands or arguments, environment
names or values, absolute or source-provided paths, hostnames, process IDs,
headers, URLs, SQL text, runner output, error messages, stack traces, test
symbols, product-authored values, or artifact contents. Failures use only the
existing low-cardinality failure class and failure reason tokens.
The trace bundle MAY identify a digested source artifact by its normalized,
repository-relative retained-artifact path; that reconstruction identity MUST
NOT be copied into OTLP attributes or metrics.
Verified by: TH-HARNESS-AC-075, TH-HARNESS-AC-076

**TH-HARNESS-REQ-613**
Ordinary harness processes MUST ignore inherited variables whose names begin
with `OTEL_` for harness-observability behavior. They MUST NOT initialize an
OTel SDK, autoconfiguration provider, exporter, resource detector, or default
localhost endpoint. The explicit export command receives only its declared
Make inputs through argv; product and browser child processes MUST NOT receive
the export endpoint or header material.
Verified by: TH-HARNESS-AC-076

**TH-HARNESS-REQ-614**
`HARNESS_OTLP_ENDPOINT` MUST be an absolute URL with scheme `https`, except that
`http` is allowed for `localhost`, `127.0.0.0/8`, or `[::1]`. User information,
query, fragment, non-HTTP schemes, encoded authority ambiguity, and redirects
are invalid. The exporter appends `/v1/traces` and `/v1/metrics` to the base
path exactly once and applies a ten-second request timeout with no automatic
retry. Invalid endpoint, selection, or header configuration is
`configuration_error`, exit `2`; collector resolution, connection, timeout,
redirect, or non-success delivery failure is a bounded
`tool_diagnostic_failure`, exit `1`, of the export command only. Neither class
may modify source evidence.
Verified by: TH-HARNESS-AC-076

**TH-HARNESS-REQ-615**
An optional header file MUST be a non-symlink regular file no larger than 64
KiB and owner-readable only; on POSIX, group or other permission bits are
invalid. Its JSON value is a closed object of at most 32 ASCII header names to
string values of at most 4096 bytes each. `host`, `content-length`,
`content-type`, and connection-management headers are forbidden. The path and
all values are redacted from output and retained diagnostics. Header files and
selected retained runs are never modified.
Verified by: TH-HARNESS-AC-076

### 15.1 Secret Pattern Table

| Secret class               | Match rule                                                                          | Redaction token                      |
| -------------------------- | ----------------------------------------------------------------------------------- | ------------------------------------ |
| Passwords                  | Exact or anchored variable/key names for `PASSWORD`, `PASS`, `PWD`, or equivalent credential suffixes | `[REDACTED:password]`                |
| Tokens                     | Exact or anchored variable/key names for `TOKEN`, `JWT`, `BEARER`, cookie headers, or equivalent credential suffixes | `[REDACTED:token]`                   |
| API or access keys         | Exact or anchored `API_KEY`, `ACCESS_KEY`, `SECRET_KEY`, or credential-context key names | `[REDACTED:key]`                     |
| DSNs/URLs with credentials | URL userinfo or DSN password segment present                                        | `[REDACTED:dsn]`                     |
| Object-store credentials   | S3-compatible access key or secret key variables                                    | `[REDACTED:object-store-credential]` |
| Private keys               | PEM private-key block markers                                                       | `[REDACTED:private-key]`             |

### 15.2 Artifact Redaction Table

| Artifact class                      | Redaction requirement                                                                                                      |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Tool/run/target/scheduler summaries | Redact before write.                                                                                                       |
| Machine stdout JSON                 | Redact before stdout.                                                                                                      |
| Captured child stdout/stderr logs   | Redact before retention.                                                                                                   |
| Service env files and env dumps     | Store only redacted credential values unless the file is required for child execution and kept under private runtime root. |
| Browser screenshots/videos/traces   | Diagnostic secret-bearing; not safe for publication.                                                                       |
| Visual geometry diagnostics         | Diagnostic secret-bearing; not safe for publication.                                                                       |
| Playwright HTML reports             | Diagnostic secret-bearing; not safe for publication.                                                                       |
| CI logs                             | Redact using the same token rules before harness-controlled emission.                                                      |

## 16. Integration with Product Specifications

**TH-HARNESS-REQ-650**
Harness verification contracts and catalog rows route behavior classes and
executable selectors only. They MUST NOT copy or resolve product requirement
IDs, acceptance criteria, specification lifecycle status, deleted delivery
registries, or ledgers, and they MUST NOT redefine product behavior under test.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-016

**TH-HARNESS-REQ-651**
Support-profile tests, helper commands, raw aggregate suites, and direct package scripts MUST NOT be counted as authoritative product-conformance evidence unless an adopted verification contract and canonical Make target select them and emit required evidence artifacts.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-652**
Load, login bursts, service resets, artificial stress margins, and browser harness stress tests are harness-only unless product specs explicitly adopt them. Browser login bursts MUST NOT be used as the sole evidence for Core 04 session-cap semantics when backend or integration evidence can prove victim selection and revocation delivery directly.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-653**
Timing-sensitive browser evidence for asynchronous socket behavior MUST prove the relevant sender readiness, receiver readiness, event identity, and diagnostic capture boundary before starting the measured interaction. A timed assertion MUST measure the product event under test, not page navigation, socket establishment, route cleanup, or waiter attachment.
Verified by: TH-HARNESS-AC-013

**TH-HARNESS-REQ-654**
Harness recovery verification MUST be product-owner-subordinate. A harness target, catalog row, retained artifact, task-surface entry, topology entry, scheduler work unit, or runtime-binary injection rule MUST NOT redefine recovery CLI authorization, logical command grammar, result schema, progress schema, timeout defaults, exit-code mapping, safe-output rules, restore-target preflight, journal behavior, or public route absence.

Any row claiming Core 04 AC-402, AC-427, or AC-428 MUST cite the Core-owned recovery requirements it verifies and MUST treat the recovery CLI as `deployment_admin`-irrelevant. A test that requires `deployment_admin`, session cookies, bearer tokens, CSRF, browser Origin, incident role, common-job authorization, WebSocket authorization, or public HTTP route access for recovery CLI invocation MUST NOT count as recovery conformance evidence.

Legacy tests that characterize a repository implementation requiring `deployment_admin` for recovery commands are negative evidence only. They MUST NOT close product-conformance rows, MUST NOT use blocked or stale status as current closure, and MUST NOT verify AC-402, AC-427, or AC-428.

Verified by: TH-HARNESS-AC-043

**TH-HARNESS-REQ-655**
Operator recovery conformance rows MUST map implementation-owned executable or wrapper behavior to exactly one Core logical command from Core 01 §12.2.1 before product assertions execute.

The current valid conformance command set is closed:

| Logical command | Operation token |
| --- | --- |
| `operator backup inspect latest` | `backup_inspect_latest` |
| `operator backup create` | `backup_create` |
| `operator restore latest` | `restore_latest` |
| `operator restore-verify latest` | `restore_verify_latest` |
| `operator restore-verify due` | `restore_verify_due` |

Compatibility aliases are not a second public recovery contract. A recovery invocation that is not one of the five canonical logical commands is negative evidence only and MUST fail current conformance routing.

Verified by: TH-HARNESS-AC-044

**TH-HARNESS-REQ-656**
Restore-workbook-probe evidence MUST cite a product owner for the probe behavior. A recovery catalog row MAY use the fixture only when Core 01 or an adopted recovery NLSpec owns the probe and the verification contract cites that owner.

The harness MAY route, schedule, and retain evidence for the probe, but it MUST NOT define workbook query semantics, selected workbook surfaces, source-row eligibility, pass/fail reason codes, or operator error mapping by itself. If no owner-defined probe contract exists for a claimed row, the harness MUST report the row as blocked or unsupported rather than inventing semantics from test fixtures, filenames, or package names.

Verified by: TH-HARNESS-AC-047

**TH-HARNESS-REQ-657**
Network Flow fixture manifests are byte-freeze and execution-input artifacts.
They MUST omit product requirement IDs, acceptance IDs, verification IDs,
copied owner text, and test or phase selectors. Product semantics remain in the
adopted Network Flow, Core, and Graph Projection specifications. The test
catalog routes fixture execution through the Network Flow behavior
verification without copying specification identities into the fixture.
Verified by: TH-HARNESS-AC-049

**TH-HARNESS-REQ-658**
Network Flow fault controls are harness mechanics for exercising adopted
product-owned commit, worker, cancellation, replay, and recovery behavior. A
fixture or target MAY use `cartulary.test.network_flow_fault_control.v1` only
from a row routed through the Network Flow behavior verification. Fault-control
boundary names, fault kinds, correlation keys, and route responses MUST NOT be
cited as independent product semantics, public API compatibility, Core 05
publication evidence, or performance evidence.
Verified by: TH-HARNESS-AC-050

**TH-HARNESS-REQ-659**
Network Flow clock-control evidence is harness mechanics for exercising adopted
product-owned time behavior. A fixture or target MAY use
`cartulary.test.clock_control.v1` only from a row routed through the Network
Flow behavior verification. Clock route responses, wall-clock mode names,
offsets, and fixed timestamps MUST NOT be cited as independent product
semantics, public API compatibility, Core 05 publication evidence, or
performance evidence.
Verified by: TH-HARNESS-AC-051

**TH-HARNESS-REQ-660**
Network Flow deterministic-randomness evidence is harness mechanics for
exercising adopted product-owned identity, nonce, digest, collision, ordering,
and replay behavior. A fixture or target MAY use
`cartulary.test.network_flow_randomness_control.v1` only from a row routed
through the Network Flow behavior verification. Stream names, value kinds,
collision values, response counts, and fail-closed exhaustion behavior MUST
NOT be cited as independent product semantics, public API compatibility, Core
05 publication evidence, production secret management, or performance
evidence.
Verified by: TH-HARNESS-AC-052

**TH-HARNESS-REQ-661**
Network Flow auth-transition evidence is harness mechanics for exercising
adopted product-owned route authorization, hidden-resource, cursor recheck, and
extension-resource invalidation behavior. A fixture or target MAY use
`cartulary.test.network_flow_auth_transition_control.v1` only from a row
routed through the Network Flow behavior verification. Boundary names,
transition kinds, safe fixture refs, hidden-response kinds, correlation keys,
and route responses MUST NOT be cited as independent product semantics, public
API compatibility, production authorization policy, Core 05 publication
evidence, or performance evidence.
Verified by: TH-HARNESS-AC-053

**TH-HARNESS-REQ-662**
Network Flow audit-assertion evidence is harness mechanics for exercising
adopted product-owned domain audit occurrence counts, transactional audit
boundaries, exact idempotency replay behavior, and no-audit failure cases. A
fixture or target MAY use
`cartulary.test.network_flow_audit_assertion_control.v1` only from a row routed
through the Network Flow behavior verification. Assertion kinds, event codes,
safe fixture refs, baseline counts, expected final counts, replay increments,
correlation keys, and route responses MUST NOT be cited as independent product
semantics, public API compatibility, audit storage design, Core 05 publication
evidence, or performance evidence.
Verified by: TH-HARNESS-AC-054

**TH-HARNESS-REQ-663**
Network Flow generation and fixture integrity MUST be checked by their owning
validators without an acceptance-accounting manifest. Contract-family
validation owns generation lifecycle and generated-output closure; the
generated-drift manifest owns scratch-copy coverage; task-surface validation
owns public targets; fixture validation owns containment, regular-file and
symlink policy, resource bounds, exact sizes, per-file digests, bundle digests,
and frozen revision state. None of those validators may parse a specification,
count acceptance IDs, resolve prose locators, or infer product completeness
from static metadata.
Verified by: TH-HARNESS-AC-055

**TH-HARNESS-REQ-664**
Platform harness runtime owns generic guarded-route authorization, control
contribution registration, reset-hook orchestration, and centralized
redaction. Product- or extension-specific control registries, request
validation, pending state, consume semantics, and product dependency adapters
MUST be owned by the corresponding module in a runtime-scanned package.

A module control contribution MUST provide only its guarded route registrars,
reset hook, and explicitly typed dependency adapters to the generic platform
boundary. Binary composition MAY register that contribution only when test
routes are enabled. Moving a control between implementation packages MUST
preserve its route path, schema ID, guard ordering, disclosure behavior,
redaction, pending-state conflict rules, one-shot consumption, and runtime
reset behavior. Test-support scan exclusions MUST NOT hide a package compiled
into the server binary.

Verified by: TH-HARNESS-AC-050, TH-HARNESS-AC-052, TH-HARNESS-AC-053,
TH-HARNESS-AC-054, TH-HARNESS-AC-056

**TH-HARNESS-REQ-666**
Harness tests and catalog rows MUST NOT carry product requirements, acceptance
criteria, deleted delivery-phase maps, or ledgers, and they MUST NOT redefine
product behavior. Support-only rows and direct package scripts cannot close
product conformance merely because a verification contract or canonical Make
target selects them; retained execution evidence remains subject to human
review against the adopted product specification.
Verified by: TH-HARNESS-AC-013, TH-HARNESS-AC-062, TH-HARNESS-AC-069

**TH-HARNESS-REQ-667**
Catalog ownership is determined in this order: owner of the normative verification postcondition; for a cross-module integration, owner of the externally visible postcondition or primary durable mutation; for mechanism-only evidence, owner of the platform or harness mechanism. Other participants are collaborators. If these rules do not produce one owner, catalog adoption is blocked. Filename, package, runner, and maintainer identity MUST NOT break a tie.
Verified by: TH-HARNESS-AC-062, TH-HARNESS-AC-068

**TH-HARNESS-REQ-668**
Evidence-class gate applicability is closed:

Every active row requires one full-owner `test-slice` partition in addition to the
evidence-class gate below. An explicit selected-subset slice remains valid execution
evidence but cannot close a full-owner audit.

| Evidence class | Required gate family |
| --- | --- |
| `unit` | Owner slice and the unit or fast gate named by the verification contract. |
| `integration` | Owner slice and the integration gate named by the verification contract. |
| `browser` | The Playwright stage's webserver-backed, stateful, or support public gate. |
| `accessibility` | `browser-e2e-a11y`. |
| `visual` | Visual golden digest validation and `browser-e2e-visual`. |
| `measurement` | `browser-e2e-measurement`; informative unless an active claim authorizes publication. |
| `static` | The exact static public target named by the verification contract. |
| `security` | The exact security public target named by the verification contract. |
| `release` | `release-check`. |

Every static or security verification contract MUST name one current public target. An unknown or private target is invalid.
Verified by: TH-HARNESS-AC-069

**TH-HARNESS-REQ-669**
A gate is `not_applicable_zero_rows` only when the selected owner has zero active rows of the corresponding evidence class or Playwright stage. A required row cannot be declared inapplicable by a caller, tracker, guide, or implementation heuristic. Exact applicability MUST be generated from the catalog and verification contracts.
Verified by: TH-HARNESS-AC-069

**TH-HARNESS-REQ-670**
An owner evidence audit MUST require one compatible successful terminal record for every active owner row and every gate derived by TH-HARNESS-REQ-668. A broad passing `check` root does not prove an explicit support, visual, accessibility, or measurement target ran. Missing explicit roots are blockers when such rows exist.
Verified by: TH-HARNESS-AC-066, TH-HARNESS-AC-069

**TH-HARNESS-REQ-671**
`default_check=true` means only that `make check` selects the row through generated topology. It does not change owner-slice omission, evidence-class meaning, claim posture, or release applicability. Generated topology MUST fail when it omits a default row or selects a non-default row without an explicit aggregate or projection contract.
Verified by: TH-HARNESS-AC-006, TH-HARNESS-AC-069

**TH-HARNESS-REQ-672**
Release-readiness aggregation MUST consume current v2 owner summaries and evidence-accounting artifacts only. It MUST preserve owner and evidence class, reject historical phase schemas, and keep implementation, informative, conformance, and Core 05 publication effects distinct.
Verified by: TH-HARNESS-AC-066, TH-HARNESS-AC-070

**TH-HARNESS-REQ-673**
The migration crosswalk is temporary implementation evidence and MUST NOT be a runtime alias, selector source, compatibility reader, or permanent catalog dependency. After its final reconciliation report is retained, the crosswalk is removed; the immutable owner catalog remains the only executable ownership source.
Verified by: TH-HARNESS-AC-068, TH-HARNESS-AC-070

**TH-HARNESS-REQ-674**
R01 through R09 and external research sources in Section 18 are rationale and supporting evidence only. Current behavior is adopted only through the requirements and tables in Sections 1 through 17. An implementation MUST NOT infer a missing default or policy from a research report.
Verified by: TH-HARNESS-AC-016, TH-HARNESS-AC-071

## 17. Acceptance Criteria / Definition of Done

The acceptance matrix is the harness Definition of Done: it defines what must
be true at the observable boundary. It is not a mandatory one-row-per-test,
one-row-per-verification, or machine completeness map. Implementations MAY
combine compatible scenarios in one test or use multiple tests for one
scenario, provided the resulting behavior evidence is unambiguous.

Owner-slice acceptance under TH-HARNESS-AC-063 through TH-HARNESS-AC-065 MUST cover omitted and exact-row plans, complete owner and service-backed dependency scopes, and target-wide checks whose closure remains limited to resolved rows. Negative fixtures MUST cover missing and unknown owners, blank tokens, normalized duplicates, cross-owner IDs, inactive rows, unmapped rows, zero-row owners, and a non-service-backed row requested by `service-backed-test-slice`. Each rejected selection MUST exit `2` before setup or child launch. The retained plan, retained scheduler summary, and target-local JSON scheduler summary MUST contain identical selection identity, and every selected-subset fixture MUST prove that row rollup is not treated as full-owner completion.

Cutover acceptance MUST execute in this order:

1. Review specification ownership, internal references, and open decisions
   editorially; executable tools do not inspect specification content.
2. Run `make lint-markdown` as documentation maintenance only.
3. Run schema and negative-fixture validation through `make json-shape-check`.
4. Run `make generated-artifact-policy-check`.
5. Run `make harness-contract` and inspect the public command/input surface.
6. Regenerate only through `make generate`, then run `make generate-drift`.
7. Run focused owner slices and catalog, selector, accounting, and audit fixtures.
8. Run every gate generated by TH-HARNESS-REQ-668, including frontend, browser, accessibility, visual, measurement, security, and release row evidence.
9. Run `make agent-finalize` without retained-run maintenance when no eligible full warm root exists.
10. Produce one successful full warm `make check` root, run `make agent-finalize RESULTS_DIR=<that-exact-root>`, repeat the broad verification affected by finalization, and run `make release-check`.
11. Confirm zero active v1 targets, inputs, schema readers, catalog readers, artifact writers, generated drift, and unresolved migration items.

A later step MUST NOT compensate for an earlier failure, and evidence from different source/catalog/verification identities MUST NOT be combined across steps.

| ID                | Requirement owner  | Scope                            | Setup fixture                                                                | Invocation                                                                                | Expected exit/status                                           | Stdout                                                 | Stderr                                                       | Required artifacts                                                                                 | Negative case                                                                | Cleanup expectation                                     |
| ----------------- | ------------------ | -------------------------------- | ---------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- | -------------------------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------- |
| TH-HARNESS-AC-000 | Section 8          | Schema validation                | Any public target that emits required JSON                                   | Target named by the fixture                                                               | Success only if JSON validates                                 | Per Section 7                                          | Per Section 7                                                | Every emitted required JSON artifact validates against Section 8 schema attachments                | Inject schema-invalid required summary                                       | No extra cleanup beyond target contract                 |
| TH-HARNESS-AC-001 | Sections 1, 4      | Command registry                 | Current tree                                                                 | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` plus registry parity checker    | `0` when registry matches exactly                              | Bounded report                                         | Empty on success                                             | Public target registry parity report                                                               | Extra/missing public target fails                                            | none                                                    |
| TH-HARNESS-AC-002 | Section 5          | Config precedence                | Fixture target with CLI, Make var, env, manifest, config, default candidates | Dedicated config resolver test target or unit harness                                     | `0`                                                            | Machine or bounded summary                             | Empty on success                                             | Resolver summary showing CLI > Make var > env > manifest > config file > default                   | Non-positive scheduler limit exits `2` with `configuration_error`            | no child work                                           |
| TH-HARNESS-AC-003 | Sections 5, 6      | Result root, run ID, and prepared identity | No child work required                                                       | Invalid result root, invalid run ID, unsafe custom result root, and complete/partial prepared-identity fixtures | `2` for invalid or partial identity; `0` for complete prepared reuse | Empty or failure JSON according to target output class | Bounded config diagnostic                                    | Failure summary when wrapper starts; retained root preflight rejects unsafe custom permissions and partial prepared identity before writes | Slash, backslash, whitespace, `.`, `..`, existing non-empty unprepared run dir, world-writable custom root, partial prepared tuple, non-`1` marker, and target mismatch all fail | no child work and no artifact creation for rejected identity |
| TH-HARNESS-AC-004 | Sections 7, 8      | Machine output accepted          | Toolchain ready; explicit result root/run ID                                 | `CARTULARY_OUTPUT_MODE=machine make backend-unit`; `... make test-fast`; `... make check` | Target status                                                  | Exactly one JSON object plus LF                        | Empty after wrapper starts                                   | `cartulary.tool_run_summary.v5` and target artifacts                                               | Progress prose or duplicate JSON fails                                       | normal target cleanup                                   |
| TH-HARNESS-AC-005 | Section 7          | Machine output rejected          | No child work                                                                | `CARTULARY_OUTPUT_MODE=machine make clean`; `... make dev`; `... make help`               | `2`                                                            | Empty                                                  | Bounded `usage_error` diagnostic                             | None required                                                                                      | Child work starts despite rejection                                          | no deletion or service start                            |
| TH-HARNESS-AC-006 | Section 10         | Scheduler determinism            | Controlled manifest with simultaneous child completions and scheduled browser groups | Run scheduler fixture twice with same manifest; validate generated browser worker-admin slot ranges | `0`                                                            | Bounded summary or machine object                      | Empty on success                                             | Byte-identical scheduler events after dynamic timestamp normalization allowed only by schema rules; browser group worker slots are explicit, contiguous, and non-overlapping | Event sequence differs; browser worker slot env is missing or overlaps       | finalizers run                                          |
| TH-HARNESS-AC-007 | Section 11         | Service modes                    | Owned and attach fixtures                                                    | Owned service target; attach target missing one required var                              | owned success; attach failure `2`                              | Bounded summary                                        | Empty on owned success; config diagnostic for attach failure | Owned lease before child work; attach failure summary                                              | Attach mode deletes container-level resource                                 | owned teardown recorded                                 |
| TH-HARNESS-AC-008 | Section 12         | Test-only harness routes         | Browser test runtime with test route token and saved-view fixture inputs     | Reset route success, saved-view system fixture success, auth rejection, origin/host rejection, concurrent reset, timeout, partial failure fixtures | Expected HTTP statuses from Section 12                         | HTTP JSON response                                     | n/a                                                          | Reset response validates schema; saved-view fixture response is a normal saved-view resource with `scope='system'`; tainted stack marker on partial failure; no permissive CORS headers | Default runtime exposes any test route, wrong host/origin reaches mutation, product auth bypasses the test token, saved-view fixture accepts caller-supplied scope/owner/identity, or wildcard CORS is emitted | tainted stack restarted before further work             |
| TH-HARNESS-AC-009 | Section 13         | Cleanup and destructive reset guard | Synthetic registry with safe and unsafe paths; fake Compose, database, migration, and object-store commands | Cleanup guard unit; `CARTULARY_CLEANUP_DRY_RUN=1 make clean`; dry-run and missing-confirmation invocations for `services-down`, `db-reset`, and `object-store-reset` | `0` for safe dry-run; nonzero for unsafe synthetic path or missing destructive confirmation | Dry-run lines match format                             | Bounded guard or confirmation diagnostic before mutation      | Candidate list, guard evidence, and command-shape evidence for confirmed local resets                              | Empty path, `/`, `.`, `..`, traversal, protected root, outside-repo path, symlink-following, inherited-env-only destructive confirmation, object-store reset touching another bucket, or `services-down` removing volumes accepted | no deletion, service start, or service stop in dry-run  |
| TH-HARNESS-AC-010 | Section 13         | Stale janitor proof gates        | Fake DB, bucket, container, and browser fixtures with/without proof          | Focused stale-janitor tests                                                               | `0`                                                            | Bounded summary                                        | Empty on success                                             | Evidence that unproven resources retained and proven stale fixtures deleted only outside dry-run   | Resource lacking generated name/proof deleted                                | unproven resources retained                             |
| TH-HARNESS-AC-011 | Section 15         | Redaction                        | Fake DSN, object-store secret, token, header, cookie, CLI arg, nested JSON, structural session fields, and private-key fixtures | Redaction unit plus one wrapper log capture                                               | `0`; redaction/write failure exits `11` unless Section 9.1 preserves an earlier primary failure | No unredacted secret in machine JSON                   | No unredacted secret in captured stderr                      | Summaries/logs contain required redaction tokens, preserve schema-owned object/array fields, and use owner-read/write file modes | Any secret pattern appears unredacted, required structural fields are replaced by redaction tokens, or required retained log is group/world-readable | none                                                    |
| TH-HARNESS-AC-012 | Section 14         | Platform matrix                  | Platform claim checker fixture                                               | Platform matrix checker                                                                   | `0` for allowed profiles; nonzero for unsupported claim        | Bounded summary                                        | Diagnostic on unsupported claim                              | Matrix report                                                                                      | macOS/Windows-native/Podman claimed as current conformance                   | none                                                    |
| TH-HARNESS-AC-013 | Sections 9, 16     | Product versus harness failure   | One known failing assertion, one harness setup failure, and one browser strict-port conflict fixture | Canonical test target under each fixture                                                  | Product failure exits `10`; setup failure exits Section 9 code; strict-port conflict retains its existing resource-conflict exit | Failure headline names class and reason                | Bounded diagnostic                                           | Target/tool, owner, and scheduler summaries agree on `resource_conflict` | Setup failure classified as product, or strict-port conflict classified as `service_start_error` | harness cleanup attempted                               |
| TH-HARNESS-AC-014 | Section 9          | Exit-code matrix                 | Controlled failure fixtures                                                  | Exit matrix test target                                                                   | Exact Section 9 code for every class                           | Per output mode                                        | Per output mode                                              | Failure summaries with primary failure selection                                                   | Cleanup failure overrides earlier product failure                            | cleanup failure recorded but primary exit preserved     |
| TH-HARNESS-AC-015 | Sections 6, 8      | Retained artifact identity       | Explicit result root/run ID plus generated default identity fixtures for public node-tool and owner-slice targets | `CARTULARY_TEST_RESULTS_DIR=<dir> CARTULARY_TEST_RUN_ID=<id> make backend-unit`; direct generated-ID public targets | `0`                                                            | Summary names run root                                 | Empty                                                        | Artifacts under one `<dir>/<id>` with target, run ID, run root, invocation marker, and terminal summary; retained run roots and target dirs are owner-only on POSIX hosts | Preflight marker and summary use sibling generated IDs, newest-run fallback is accepted as proof, or retained directories are group/world-accessible | custom absolute result root not removed by `make clean` |
| TH-HARNESS-AC-016 | Sections 1, 2, 18, 19 | Editorial and boundary closure | Revised document                                                             | Human owner review; `make lint-markdown` checks Markdown quality only                      | Review complete; Markdown lint `0`                             | Existing Markdown-lint output                         | Existing Markdown-lint diagnostic                            | Human review records owner conflicts or open decisions; no executable artifact consumes this document | A specification conflict is silently resolved by a machine projection or Markdown lint is cited as product conformance | none                                                    |
| TH-HARNESS-AC-017 | Section 11         | Lifecycle-machine conformance    | Service-suite fixtures for happy path, startup failure, interrupted child, cleanup failure, illegal transition, and crash/rerun | Lifecycle-machine conformance target or unit harness                                      | Happy path `0`; failure fixtures use exact Section 9 code      | Bounded summary or machine object                      | Empty on happy path; bounded diagnostic on failure fixture | `cartulary.test_services.lifecycle.v2` stream with sequential events, valid transitions, terminal state, Section 9 failure mapping, and cleanup proof behavior | Unlisted `(state,event)` mutates state, terminal state accepts later event, or lifecycle stream validates with a sequence gap | normal suite cleanup; unproven resources retained       |
| TH-HARNESS-AC-018 | Sections 4, 10, 11 | Warm scheduler health            | Retained warm-ready `check` fixture plus over-budget, cold-provisioning, measurement-in-default-check, skewed-lane, shared-browser-session, unexpected-reuse, and fixture-budget fixtures | `make scheduler-summary-timing-drift RESULTS_DIR=<dir> TARGET=check SCHEDULER_WARM_CHECK_BUDGET_MS=60000 SCHEDULER_WARM_CHECK_BALANCE_RATIO=1.25` | Success only for warm-eligible, in-budget, balanced fixtures                    | Bounded summary                                        | Bounded diagnostic on failure fixture                  | Scheduler and target summaries identify `check-service-backed` wall time, evaluated lanes, browser session groups, unexpected reused accounting, readiness attribution, and fixture class counts | Measurement stage, undeclared extra browser session, hidden provisioning, unexplained reused work, unplanned clone/reset, or skewed non-isolated lane passes unnoticed | none                                                    |
| TH-HARNESS-AC-019 | Section 8.2        | Agent finalizer                  | Fake Make fixture plus valid, missing, failed, incomplete, contaminated, non-warm retained run, action-cache hit/miss/disabled/corrupt/input-change/output-change fixtures | `make agent-finalize`; `RESULTS_DIR=<dir> make agent-finalize`; `CARTULARY_OUTPUT_MODE=machine make agent-finalize` | Success for coherent maintenance inputs; fail-fast for first failed action substep or invalid retained run; cache hits only for eligible closed-profile actions | One `[FINALIZE]` line then bounded result/artifact lines; machine emits one JSON object | Bounded failure diagnostic naming failed action/substep | `agent-finalize/finalize-summary.json`, per-action `execution_state` and cache state, child summaries/logs when executed, and `finalize_summary` artifact ref | Excluded targets run, mutation starts after invalid `RESULTS_DIR`, cache hit bypasses retained-run validation, corrupt cache produces success, machine output requires log parsing, semantic action IDs are absent, or skipped-after-failure work is absent | No cleanup or destructive command is run               |
| TH-HARNESS-AC-020 | Section 4          | Public target semantic value     | Current target registry plus one synthetic shallow-wrapper fixture           | Registry semantic-value checker                                                           | Success only when every public target declares at least one semantic behavior and every declared behavior has an owner section | Bounded report                                         | Empty on success                                             | Semantic-value parity report                                                                    | Target with only child command aliases and no semantic behavior passes        | none                                                    |
| TH-HARNESS-AC-021 | Sections 5, 8, 10   | Frontend-unit harness stability  | Current check topology plus delayed jsdom workbook-row, inspector-subject, row-history selector, controlled-input helper fixtures, and Vitest failure sidecar fixtures | Topology contract tests; `make frontend-unit`; constrained `CHECK_HOST_CPU_JOBS=2 make check` | Success when scheduled Vitest workers and resource claims match, shared helpers tolerate bounded async hydration, inspector readiness matches stable subject identity without retrying activation, and sidecar diagnostics are preferred when present | Bounded summary                                        | Empty on success                                             | Scheduler manifest shows `frontend-unit` `host_cpu=2` and `VITEST_MAX_WORKERS=2`; frontend helper, inspector-subject, and selector-policy tests pass; `frontend-unit` retains `cartulary.vitest_failure_details.v1` when `runner.json` exists; synthetic `STACK_TRACE_ERROR` plus sidecar reports the sidecar assertion message | `frontend-unit` can run more workers than it claims, row or inspector waits use unbounded/ad hoc selectors, inspector readiness retries activation or omits the expected and mounted subject identity, row-history waits use display labels instead of `history_item_ref`, helper diagnostics omit mounted row identity, sidecar schema drift is accepted, or `STACK_TRACE_ERROR` replaces an available sidecar assertion message | none                                                    |
| TH-HARNESS-AC-022 | Sections 1, 4, 5, 8, 10 | Frontend readiness harness metadata | Active frontend catalog rows plus missing accessibility target, raw-only artifact, missing visual fixture, stale accounting, support-specimen substitution, concept-image substitution, cross-owner selection, and removed-preflight fixtures | Task-surface report, catalog/JSON validation, `explain-test-owner`, accessibility/visual fixture checks, and evidence-accounting validation | Success only when active owner rows, exact selectors, profiles, semantic fixture support, normalized accessibility summary, and selected-row accounting closure are valid | Bounded report or explanation | Empty on success; bounded configuration diagnostic | Public registry includes `browser-e2e-a11y`; catalog owns frontend rows; semantic fixture metadata owns capture scope; `visual.fixture.default_timeline_workbook_shell` closes only from its exact active catalog scenario and app-owned screenshot; support specimens and concept images cannot close it; v4 accessibility and v1 evidence-accounting schemas validate | Filename/title inference, cross-owner closure, stale accounting, support substitution, raw third-party output, or removed preflight target passes | no child browser work required for metadata fixtures |
| TH-HARNESS-AC-023 | Sections 4, 7      | Registry output-class and side-effect parity | Current NLSpec registry, `tools/task_surface_manifest.json`, rendered task-surface report, and output-class matrix | Registry parity checker and malformed registry fixtures | Success only when target membership, output class, artifact policy, schema policy, and side effects agree exactly | Bounded parity report | Empty on success; bounded drift diagnostic on failure | Public registry parity report with side-effect and output-class comparison | Missing `side_effects[]`, `none` plus another class, undeclared side effect, or output-class drift passes | none |
| TH-HARNESS-AC-024 | Section 10         | Scheduler command-shape closure | One live generated scheduler fixture for each required command type plus malformed command descriptors, prerequisite-policy fixtures, and artifact-consumer dependency fixtures, with optional command types validated when present | Scheduler manifest schema and shape checks | Success only for closed command shapes, required and closed `make_prerequisite_policy` values for `make_target`, and declared artifact producer/consumer DAG edges | Bounded summary | Empty on success; bounded shape diagnostic on failure | Shape-check evidence naming command type, required fields, forbidden fields, wrong-type failures, prerequisite policy, and artifact dependency parity | Missing required field, forbidden field, unknown command type, wrong field type, omitted or invalid prerequisite policy, or missing build-artifact producer edge passes | no child work |
| TH-HARNESS-AC-025 | Section 8          | Schema attachment closure | Every Section 8 `present` schema attachment plus positive and negative fixtures | Schema attachment policy check and schema fixture validation | Success only when every present attachment exists, parses, is top-level closed or extension-container closed, and validates fixtures | Bounded summary | Empty on success; bounded schema diagnostic on failure | Schema attachment report | Missing schema, malformed schema, open top-level schema without extension container, or fixture-blind schema passes | none |
| TH-HARNESS-AC-026 | Sections 1, 16     | Core test-route traceability | Core 04, verification contracts, and active catalog rows | Requirement-ID uniqueness check and `REQ-04-109` citation classifier | Success only when `REQ-04-109` means test-only runtime-control route security and public-origin behavior cites `REQ-04-110` | Bounded traceability report | Empty on success; bounded traceability diagnostic on failure | Duplicate-ID report and citation classification report | Public route, WebSocket origin, evidence-handle, or deployment-origin row cites `REQ-04-109` | none |
| TH-HARNESS-AC-027 | Section 4          | Public registry source-of-truth parity | Current public Make surface, task-surface manifest, execution topology, generated Make include, scheduler reachability, projection metadata, and prose registry | Registry source-of-truth parity checker | Success only when every public target exists in the NLSpec registry and `tools/task_surface_manifest.json`, no removed target appears in public generated mirrors, no public target is introduced only by topology, generated Make, or prose, every advertised `check` inclusion is reachable as direct work/aggregate/projection, and support-only internal targets do not advertise default `check` unless scheduled | Bounded parity report | Empty on success; bounded drift diagnostic on failure | Public-source parity report with default-check projection labels and removed-target exclusions | Public target appears only in execution topology, generated Make includes, or prose; removed target remains public; full direct browser target is counted when only a smoke projection ran; support-only target advertises unselected `check` membership | none |
| TH-HARNESS-AC-028 | Sections 4, 5, 8, 13 | Local cache profiles | Cache helper fixture plus representative readiness, build, finalizer, and render-cache fixtures | Cache-helper smoke tests; cold/hot readiness and build target runs; `make agent-finalize`; render drift checks; cleanup fixture | Success only when first run misses, second run hits, disable/force/corrupt/missing-output cases execute or fail safely, summaries remain emitted, and scheduler accounting reports no undeclared reuse | Bounded summary | Empty on success; bounded diagnostic on invalid cache state | Cache records validate against Section 8 schemas; run-root cache artifacts show state/reason/record path; public summaries remain present | Security, drift, service readiness, runtime reset, cleanup, destructive guard, browser/service-backed live-state work, or aggregate success is accepted by cache reuse; missing output succeeds by reuse; cache artifact is cited as product evidence | `make clean` preserves default cache roots; `make distclean` removes default cache roots |
| TH-HARNESS-AC-029 | Sections 1, 5     | Public input matrix closure | Current NLSpec input matrix and task-surface metadata | Input-contract parity checker | Success only when every public target input in task-surface metadata appears in the NLSpec matrix and every NLSpec matrix row mirrors metadata or an explicitly documented NLSpec default override | Bounded parity report | Empty on success; bounded drift diagnostic on mismatch | Public input matrix parity report | Public target accepts an input absent from the NLSpec matrix, or NLSpec row omits default, bound, empty-string, invalid, summary, or forwarding behavior | none |
| TH-HARNESS-AC-030 | Section 10        | Scheduler defaults and auto policies | Fixture schedules for each resource and auto policy, plus override and impossible-resource fixtures | Scheduler resource-resolution tests | Success only when every fixed default, override bound, auto formula, omission rule, and impossible-resource failure matches Section 10 | Bounded summary | Empty on success; bounded diagnostic on mismatch | Resource-limit source and resolved-limit evidence in scheduler summaries | `auto` resolves differently, override above max passes, omission lacks default/auto behavior, or impossible resources start child work | no child work beyond fixture scheduler |
| TH-HARNESS-AC-031 | Section 8         | Scheduler diagnostic artifact closure | Current scheduler artifact families plus positive and negative pressure-summary and fixture-tier proof fixtures | Schema/artifact policy checker and schema fixture validation | Success only when `pressure-summary.json` is schema-owned by present `cartulary.scheduler_pressure_summary.v4`, embeds retained proof objects that validate as `cartulary.fixture_tier_proof.v2`, validates before scheduler success, and rejects unknown or missing required fields | Bounded report | Empty on success; bounded diagnostic on mismatch | Artifact policy report naming pressure-summary status and schema fixture results | Missing `cartulary.scheduler_pressure_summary.v4` or `cartulary.fixture_tier_proof.v2` schema, open top-level pressure-summary fields, open fixture proof fields, or unvalidated scheduler success passes | none |
| TH-HARNESS-AC-032 | Section 9         | Primary failure determinism | Simultaneous failure fixtures covering class, lifecycle, event, target, path, and reason ties | Exit matrix and primary-failure unit tests | Success only when selected primary failure and public exit follow Section 9.1 and TH-HARNESS-REQ-304 exactly | Bounded summary | Bounded diagnostic on mismatch | Failure summary with ordered candidate failures and selected primary | Cleanup overrides earlier non-cleanup failure, or tie order differs across runs | cleanup failure recorded when fixture creates one |
| TH-HARNESS-AC-033 | Section 11        | Concurrent lifecycle | Service-suite lifecycle fixtures with overlapping child work, duplicate child start, unknown child finish, and interruption | Lifecycle-machine conformance target or unit harness | Success only when active-child counts transition legally and illegal duplicate/unknown events fail closed | Bounded summary or machine object | Empty on happy path; bounded diagnostic for illegal fixtures | Lifecycle stream with `active_child_count`, child keys, legal transitions, and terminal state | Concurrent child start is rejected, unknown finish mutates state, duplicate start passes, or active count becomes negative | normal suite cleanup; unproven resources retained |
| TH-HARNESS-AC-034 | Section 12        | Runtime reset closure | Reset fixtures with public mutable tables, migration metadata, bootstrap admin, route idempotency, object-store objects, partial failure, and pending public-error fault | Test-runtime reset route tests | Success only when selected table ordering, migration metadata preservation, bootstrap restoration, route-idempotency clearing, object cleanup, partial-failure response, and fault clearing match Section 12 | HTTP JSON response | n/a | Reset response validates schema and includes sorted `tables_reset`, post-reset counts, object count, partial failure fields when applicable | Unsorted or incomplete table set, migration metadata truncation, missing bootstrap admin, route idempotency residue, uncleared fault, or object residue passes | tainted stack restarted before further work |
| TH-HARNESS-AC-035 | Section 12        | Test-route edge closure | Weak token, malformed token, missing/wrong header, pending-fault conflict, consumed-fault retry, and reset-clears-fault fixtures | Test-only route unit/integration tests | Expected HTTP status or startup/config failure for every Section 12 token and fault edge case | HTTP JSON response where route starts | n/a | Route guard failures use `test_route_forbidden`; duplicate pending faults use `test_public_error_fault_already_armed`; startup configuration failures use `configuration_error` | Weak token starts, product auth bypasses token, second fault replaces pending fault, consumed fault remains armed, or reset does not clear fault | runtime reset clears fault state |
| TH-HARNESS-AC-036 | Sections 13, 15   | Cleanup and redaction closure | Protected-root cleanup fixtures, cleanup-owned child paths, structured secret keys, raw secret text, and structural field names | Cleanup guard and redaction tests | Success only when protected-root attempts fail, cleanup-owned paths succeed or no-op when missing, exact key/raw-text redaction applies, and schema-owned structures are preserved | Bounded summary | Empty on success; bounded diagnostic on mismatch | Cleanup guard report and redacted summary/log fixtures | Protected root deletion passes, missing cleanup-owned path fails, secret leaks, or structural fields are over-redacted | no deletion outside cleanup-owned fixtures |
| TH-HARNESS-AC-037 | Sections 5, 8, 10 | Operator runtime binary injection | Current topology plus missing, non-executable, digest-mismatch, undeclared-input, and raw-Go fallback fixtures | `make build-operator`; scheduler-selected operator Go work; `make check OPERATOR_BIN=/tmp/x`; `make check CARTULARY_OPERATOR_BIN=/tmp/x` | Producer succeeds with declared output path; consumer fails before product assertions for invalid injection; undeclared public inputs exit `2`; digest mismatch exits `11` | Bounded summary | Empty on success; bounded config/artifact diagnostic on mismatch | `build-operator` tool summary and build-artifact cache artifact; operator aggregate `runtime-binaries.json`; Go runner logs contain no nested `make build-operator` for scheduler-selected operator work | Hidden nested operator builds, arbitrary child forwarding, missing producer dependency, missing runtime-binary provenance, or binary cache hit marked as scheduler `reused` passes | no extra cleanup beyond build output contract |
| TH-HARNESS-AC-038 | Section 4.1A      | Helper ownership registry | Current helper registry plus allowed facade and forbidden catalog/execution/accounting/scheduler-internal import fixtures | Harness import-boundary contract tests | Success only when every helper family is classified exactly once and no unclassified private catalog, execution, accounting, backend, browser, scheduler, duration, diagnostics, generated-artifact, or finalization helper is imported by non-owner code | Bounded report | Empty on success; bounded import-boundary diagnostic on mismatch | Import-boundary report naming helper family, facade key, source, and target | Non-owner code imports private helpers after a facade exists, or a family is unclassified or duplicated | none |
| TH-HARNESS-AC-039 | Section 4.1A      | Compatibility paths | Current tree plus semantic-boundary and unknown-owner fixtures | Caller inventory, import-boundary contract tests, and relevant public target characterization | Success only when historical backend, frontend, scheduler, and execution paths are absent as compatibility paths, no temporary redirects are accepted, current facades are owned exactly once, semantic boundary rules reject bypasses, generated/task-surface metadata does not reference removed paths, and public Make behavior remains unchanged | Bounded report | Empty on success; bounded semantic ownership diagnostic on mismatch | Authored helper-owner validation and import-boundary report naming current facade key, source, and target | A forwarding shim is added without a demonstrated consumer, a facade is missing or duplicated, unknown owner roots or cross-owner private imports pass, removed paths return in generated metadata, or public behavior changes after deletion | none |
| TH-HARNESS-AC-040 | Section 4.1A      | Govulncheck findings ownership | Govulncheck JSON stream fixtures covering no findings, package/module findings, symbol findings, redaction, and malformed JSON | Static-analysis security findings tests and `go-vulncheck` when toolchain is ready | Success only when static-analysis/security ownership proves identical normalized findings, redaction, exit mapping, and artifact behavior after helper movement | Bounded summary | Empty on success; bounded parse/security diagnostic on failure | `govulncheck-findings.json` validates `cartulary.govulncheck_findings.v1` with deterministic finding order | Backend path remains a supported import, redaction changes, symbol findings stop blocking, or exit mapping drifts | temp files removed |
| TH-HARNESS-AC-041 | Section 4.1A      | Migration/schema validator ownership | Migration history, schema object ownership, scratch migration, and JSON-shape fixtures | `json-shape-check`, `migration-drift`, and database-contract drift tests | Success only when `json-shape-check` and `migration-drift` retain manifest schema validation, scratch apply behavior, diagnostics, and failure classification after helper movement | Bounded summary | Empty on success; bounded drift diagnostic on failure | Migration/schema manifest validation summaries and migration drift retained artifacts | Schema IDs change, diagnostics drift unexpectedly, scratch DB cleanup changes, or validators become backend execution behavior owners | scratch DB cleanup per migration-drift contract |
| TH-HARNESS-AC-042 | Section 4.1A      | Duration retained-run safety | Coverage, drift, update, and finalizer retained-run fixtures including failed, partial, missing-artifact, stale, contaminated, ambiguous, valid drift, and valid full warm check roots | Duration baseline and finalizer tests | Success only when coverage is read-only, drift is read-only, update rejects invalid retained evidence before mutation, and `agent-finalize` accepts only valid retained full warm `make check` roots for mutating duration refresh | Bounded summary | Empty on success; bounded retained-run diagnostic on failure | Baseline files unchanged for invalid evidence; finalizer summary records retained-run validation before mutation | Mutating update starts before retained-run validation, partial run is accepted for finalizer refresh, invalid evidence writes baselines, or drift mutates files | invalid fixtures leave tracked baselines unchanged |
| TH-HARNESS-AC-043 | Section 16        | Recovery evidence routing | Core recovery verification contracts plus legacy `deployment_admin` negative fixtures | Catalog evidence-routing validation and affected recovery target planning | Success only when recovery CLI invocation is `deployment_admin`-irrelevant and legacy gated tests are negative-only and cannot close AC-402, AC-427, or AC-428 | Bounded report | Empty on success; bounded routing diagnostic | Report names row IDs, verification IDs, evidence class, and negative status | A gated recovery CLI test counts as product conformance or stale status closes a current row | none |
| TH-HARNESS-AC-044 | Section 16        | Canonical operator command mapping | Operator scenario rows, runtime-binary rows, and compatibility-alias fixtures | `make build-operator` plus scheduler-selected operator work or command-mapping validation | Success only when implementation-owned executable behavior maps to exactly the five Core logical commands, compatibility aliases are negative-only, and Core-owned final stdout, optional progress, timeout/default, target-config, and exit-code behavior validate | Bounded summary or machine result | Empty on success; bounded command-mapping diagnostic on mismatch | Operator runtime-binary provenance and retained operator result/progress artifacts for selected work | Old command names, compatibility aliases, or alternate JSON envelopes pass as a second public contract | no extra cleanup beyond build output contract |
| TH-HARNESS-AC-045 | Section 4.1A      | Migration-evidence classification | Migration manifest, embedded SQL source, goose ledger, schema-object ownership, and recovery catalog fixtures | `json-shape-check`, database-contract drift validation, and catalog classification checks | Success only when migration-history evidence routes through database-contract or migration-evidence ownership and cannot close operator-recovery conformance | Bounded report | Empty on success; bounded classification diagnostic | Migration/schema summaries and catalog-row owner classification | A migration-evidence operator wrapper counts as AC-428 recovery proof | none |
| TH-HARNESS-AC-046 | Section 4.1A      | Moved-test accounting | Test-path movement fixtures covering catalog, task-surface, topology, generated schedules, and runtime-binary profiles | JSON/catalog validation, task-surface report, and generated drift | Success only when owner inputs change before generated artifacts and outputs regenerate through Make | Bounded report | Empty on success; bounded accounting diagnostic | Owner-input diff plus drift/schema summaries | A moved test hand-edits generated output or changes profiles/scheduling without NLSpec coverage | none |
| TH-HARNESS-AC-047 | Section 16        | Restore-workbook-probe owner routing | Recovery probe catalog rows with owner-cited and owner-missing fixtures | Verification routing and affected owner planning | Success only when probe evidence cites Core 01 or an adopted recovery owner and harness fixtures do not define semantics | Bounded report | Empty on success; bounded owner-routing diagnostic | Probe routing report with verification ID or blocked status | Fixture, filename, or package inference closes recovery proof | none |
| TH-HARNESS-AC-048 | Section 8         | Same-run helper artifact closure | Helper aggregate fixtures plus valid, scheduler-reused, missing-digest, and old-run-scope schema fixtures | Run-summary helper fixture, `explain-run`, schema validation, and `make json-shape-check` | Success only when helper refs validate, resolve under the current run root, expose producer artifact digests and consumer refs, report helper reuse without scheduler `reused`, and fail closed for old-run or malformed refs | Bounded run summary and `explain-run` lines | Bounded schema/artifact diagnostic on mismatch | `cartulary.same_run_helper_artifact_ref.v2` retained under `_shared/same-run-helper-artifacts`; run summary links helper refs; pressure/scheduler reused counts remain current-profile zero | Old retained helper artifacts are accepted as fresh, helper reuse is reported as scheduler `reused`, malformed refs pass, or selected conformance rows are skipped | none |
| TH-HARNESS-AC-049 | Sections 8, 11, 16 | Network Flow fixture manifests | Positive frozen fixture manifest and scenario with committed source, expected, and transcript bytes plus negative fixtures for forbidden provenance fields, unsorted paths, missing files, symlink/traversal paths, digest mismatch, and draft selection | Network Flow fixture/scenario validators, schema validation, `make json-shape-check`, and the Network Flow behavior tests that execute fixtures | Success only when every selected manifest and scenario validates, hashes match exact bytes, committed fixture roots are not mutated, run-local materialization is used, frozen-only selection is enforced, and no acceptance, verification, requirement, selector, or specification-provenance field is accepted | Bounded summary naming fixture IDs and bundle digests | Bounded schema/artifact diagnostic on mismatch | `cartulary.network_flow_fixture_manifest.v2` and `cartulary.network_flow_fixture_scenario.v2` validation evidence plus retained execution summary with manifest SHA-256, source and expected bundle digests, materialized input root, produced artifact refs, and comparison status | Fixture identity inferred from an untyped filename, unlisted file accepted, committed bytes mutated, forbidden provenance accepted, draft fixture selected as current evidence, or digest/order/path mismatch reaches product code | Run-local materialization removed by result-root cleanup; committed fixture bytes retained |
| TH-HARNESS-AC-050 | Sections 12, 16    | Network Flow fault controls | Disabled route, token/host/origin edge cases, invalid boundary/kind/correlation/error-code bodies, pending conflict, exact boundary/correlation consume, reset-clears-fault, and ownerless selector fixtures | Network Flow fault-control route tests, schema validation, and Network Flow fixture targets that select commit or worker fault controls | Success only when the test route is unavailable by default, guard failures happen before mutation, request validation is closed, one pending fault is consumed once by exact boundary and optional correlation key, reset clears pending faults, and fault evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.network_flow_fault_control.v1` route response plus fixture transcript naming boundary, fault kind, correlation scope, owner refs, pre/post state counts, and replay or recovery result where product work executed | Product auth bypasses the token, a second fault replaces a pending fault, wrong boundary or correlation consumes a fault, reset leaves a pending fault, or fault-control-only evidence closes a product row | runtime reset clears pending fault state |
| TH-HARNESS-AC-051 | Sections 12, 16    | Test clock controls | Disabled route, token/host/origin edge cases, fixed clock, offset clock, reset clock, read-only state, invalid payloads, runtime-reset-clears-clock, and ownerless Network Flow time selector fixtures | Test clock route tests, reset route tests, schema validation, and Network Flow fixture targets that select time-sensitive rows | Success only when test clock routes are unavailable by default, guard failures happen before mutation/disclosure, set accepts exactly one command, reset restores wall mode, state is non-mutating, runtime reset clears registered clock state, responses validate as `cartulary.test.clock_control.v1`, and clock evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.clock_control.v1` route response plus fixture transcript naming mode, fixed/offset selection, owner refs, pre/post state, and product time result where product work executed | Product auth bypasses the token, fixed time survives runtime reset, state mutates the clock, invalid payload changes clock state, or clock-control-only evidence closes a product row | runtime reset or explicit clock reset restores wall mode |
| TH-HARNESS-AC-052 | Sections 12, 16    | Network Flow deterministic randomness controls | Disabled route, token/host/origin edge cases, invalid stream/kind/value/exhaustion bodies, duplicate value collision fixture, same-stream pending conflict, fail-closed exhaustion, reset-clears-randomness, and ownerless selector fixtures | Network Flow randomness-control route tests, schema validation, and Network Flow fixture targets that select deterministic ID, nonce, digest, collision, ordering, or replay controls | Success only when the test route is unavailable by default, guard failures happen before mutation/disclosure, request validation is closed, duplicate fixture values are preserved in order, response data never echoes values, same-stream replacement is rejected, exhausted or wrong-kind streams fail closed, reset clears registered streams, and randomness evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.network_flow_randomness_control.v1` route response plus fixture transcript naming stream, value kind, value count, owner refs, pre/post state counts, and product result where product work executed | Product auth bypasses the token, deterministic values leak in responses/logs, a second sequence replaces a pending stream, duplicate values are deduplicated, exhaustion silently falls back to production randomness, reset leaves registered stream state, or randomness-control-only evidence closes a product row | runtime reset clears registered deterministic-randomness streams |
| TH-HARNESS-AC-053 | Sections 12, 16    | Network Flow auth-transition controls | Disabled route, token/host/origin edge cases, invalid boundary/transition/resource/hidden-response/ref bodies, exact transition consume, correlation-scoped consume, independent tuple arming, duplicate tuple conflict, reset-clears-transitions, and ownerless selector fixtures | Network Flow auth-transition route tests, schema validation, and Network Flow fixture targets that select route authorization, cursor recheck, hidden-resource, or invalidation controls | Success only when the test route is unavailable by default, guard failures happen before mutation/disclosure, request validation is closed, exact boundary/actor/incident/resource/correlation matching consumes once, mismatches leave pending state, independent tuples can coexist, duplicate tuple replacement is rejected, reset clears registered transitions, responses never disclose hidden resource details, and auth-transition evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.network_flow_auth_transition_control.v1` route response plus fixture transcript naming boundary, transition kind, actor ref, incident ref, resource kind/ref, hidden response kind, owner refs, pre/post auth state, and product result where product work executed | Product auth bypasses the token, a mismatch consumes a transition, a duplicate replaces pending state, hidden resource identifiers leak in route output/logs, reset leaves registered transition state, or auth-transition-only evidence closes a product row | runtime reset clears registered auth-transition controls |
| TH-HARNESS-AC-054 | Sections 12, 16    | Network Flow audit assertion controls | Disabled route, token/host/origin edge cases, invalid assertion/event/resource/ref/count bodies, exact-count assertion, zero-occurrence assertion, no-audit replay assertion, exact consume, correlation-scoped consume, independent tuple arming, duplicate tuple conflict, reset-clears-assertions, and ownerless selector fixtures | Network Flow audit-assertion route tests, schema validation, and Network Flow fixture targets that select transactional audit, replay, failure, graph-success, or binding-count controls | Success only when the test route is unavailable by default, guard failures happen before mutation/disclosure, request validation is closed, exact event/operation/resource/correlation matching consumes once, mismatches leave pending state, independent tuples can coexist, duplicate tuple replacement is rejected, count rules fail closed, reset clears registered assertions, responses never disclose raw audit payloads or secret material, and audit-assertion evidence is owner-routed | HTTP JSON response and bounded fixture summary | Bounded error envelope on mismatch | `cartulary.test.network_flow_audit_assertion_control.v1` route response plus fixture transcript naming assertion kind, event code, operation ref, actor ref, incident ref, resource kind/ref, baseline count, expected final count, expected replay increment, owner refs, observed product counts, and product result where product work executed | Product auth bypasses the token, a mismatch consumes an assertion, a duplicate replaces pending state, invalid count rules arm a control, raw audit payloads or secret material leak, reset leaves registered assertion state, or audit-assertion-only evidence closes a product row | runtime reset clears registered audit assertions |
| TH-HARNESS-AC-055 | Sections 8, 16     | Network Flow generation and fixture integrity | Contract-family registry, generated outputs, generated-drift scratch manifest, task surface, fixture manifests, and scenario files, plus negative cases for invalid lifecycle, missing generated output, missing scratch input, missing public target, forbidden provenance, and unsafe fixture path | Contract-family, generated-artifact, task-surface, fixture/scenario, and generated-drift validators; `make json-shape-check`; `make harness-contract`; `make generate-drift` | Success only when each owner validates its own closed input, the active Network Flow family has no activation dependency, official generation reproduces outputs, scratch replay contains required typed inputs, public targets resolve, and fixture integrity rejects unsafe or forbidden metadata without parsing specifications | Bounded owner-specific summaries | Bounded schema/artifact diagnostic on mismatch | Retained JSON-shape, harness-contract, generated-policy, and generate-drift summaries | Static acceptance accounting, specification parsing, missing typed input, unsafe path, private raw target, or hand-authored generated output closes evidence | no child product work beyond the separately routed behavior tests |
| TH-HARNESS-AC-056 | Sections 4, 11, 12, 15, 16 | Test-support ownership and explicit policy | Registered shared and owner-local support roots, unknown-root fixtures, in-process server mode fixtures, fixture-policy agreement fixtures, security-profile rendering fixtures, and module control contributions | Support-inventory validator, service-backed guard tests, server-helper tests, PostgreSQL helper tests, static-analysis wrapper tests, backend boundary check, and affected product-control route tests | Success only when support roots are owner-named and registered exactly once, phase-shaped active helpers are absent, route mode and database policy are explicit, manifest and call-site fixture policy agree, every support root is security-scanned, runtime-compiled controls stay in runtime scans, and module contributions preserve guarded behavior | Bounded ownership/policy summary | Bounded configuration, boundary, or security diagnostic | Validated test-support inventory plus ordinary target summaries from the affected checks | Unknown/duplicate support root, zero-value route mode, fixture-policy mismatch, unscanned support root, phase-shaped active helper, or runtime control hidden by support exclusions passes | no cleanup beyond the selected test/service contract |
| TH-HARNESS-AC-057 | Section 4.1B | Runtime-binary and private-runner closure | Production and harness server profiles, runtime-binary registry fixtures, invalid harness-only environment, missing/nonregular/nonexecutable/digest-mismatched injection, nested-build attempts, legacy runner aliases, and static shell-import fixtures | Server build-profile tests, runtime-binary validator, harness import-boundary checks, and private runner usage tests | Success only when the repository exposes exactly the three deployable identities, the harness server remains a non-deployable build profile, production rejects harness-only inputs, every black-box row receives an exact validated scheduler artifact, and every legacy or contextless runner path fails closed | Bounded build and contract summaries | Bounded configuration, boundary, or usage diagnostic | Runtime-binary registry validation, build summaries, injection provenance, and import-boundary report | A fourth deployable appears, production consumes a harness-only input, a nested build fallback executes, an injected binary lacks identity proof, or a legacy alias succeeds | build outputs and run-local injected binaries follow their target cleanup contracts |
| TH-HARNESS-AC-058 | Sections 4, 5, 8 | Task-surface ownership and generated Make density | Authored task-surface owner, execution topology, generated task-surface projection, shared Make runtime, thin Make bindings, and synthetic-growth fixtures | Task-surface owner/schema validation, public registry parity, generation drift, density validation, and public wrapper characterization | Success only when public command metadata has one authored machine owner, topology cannot redefine it, generated v15 parity holds, unsupported `alias` profiles and per-variable origin transport are rejected, every size/line/growth budget passes, and public target behavior is unchanged | Bounded ownership and density report | Bounded configuration or generated-drift diagnostic | Owner/projection digest relationship plus byte, line-length, repeated-expansion, and synthetic-growth metrics | A generated projection becomes an owner, topology embeds task metadata, dense global input plumbing returns, budgets regress, or public behavior changes | generated scratch outputs removed |
| TH-HARNESS-AC-059 | Sections 10, 17 | `make check` performance acceptance | Documented host/capacity context, unchanged authored-input digest, one warm run, five measured runs, and contaminated-run fixtures | Canonical `make check` for every run plus retained-run inspection and external process-wall measurement | Success only when all five measured runs pass and the nearest-rank p90/maximum process wall time is below 120 seconds without target, row, evidence, scheduler, artifact, failure, or cleanup drift | Existing bounded `check` output | Existing bounded failure output; contamination is classified separately | Five current run roots with process and scheduler timing, exact target/row/coverage parity, zero missing/unmapped evidence, deterministic generated output, and cleanup proof | One anomalous run alone passes; a failed, stale, mismatched, retried, or reduced-coverage run enters the window; required work moves outside `check` | normal `check` cleanup succeeds for every run |
| TH-HARNESS-AC-060 | Sections 10, 16 | Pure Go execution-family consolidation | Backend-unit catalog rows plus compatible, incompatible, duplicate-symbol, over-selection, missing-symbol, partial-output, and attribution fixtures | Target-plan validation, Go runner fixtures, `make backend-unit`, exact owner slices, and generated drift | Success only when compatible exact-symbol rows share execution without changing identity/routing and every symbol is proven exactly once | Existing bounded target summary | Bounded row/symbol or artifact diagnostic | Per-row evidence plus before/after execution-family count | A row drops, an unowned test selects, raw/exact selectors merge, incompatible profiles merge, or a shim remains | ordinary target cleanup |
| TH-HARNESS-AC-061 | Sections 10, 16 | Deterministic service Go batching | Service-backed catalog rows and duration owners plus compatible, incompatible, isolated, oversized, partial-output, exact-failure, and growth fixtures | Shard-plan smoke, scheduler matrix, owner slices, Go runners, profile agreement, generated drift, and `make check` timing | Success only when compatible items share deterministic shards without changing row/scenario/owner coverage, profiles, attribution, artifacts, or cleanup | Existing bounded shard and target summaries | Bounded planning, resource, row/symbol, or artifact diagnostic | Shard metadata, process counts, plan digest, profile parity, and current evidence | Packing crosses a boundary, selects unowned work, loses a row, weakens a claim, accepts partial output, or is nondeterministic | ordinary target and service cleanup |
| TH-HARNESS-AC-062 | Sections 3, 8 | Owner, verification, and identifier closure | Valid registries plus missing, extra, null, duplicate, unordered, unknown, Unicode, overlength, and phase-token fixtures | Catalog schema and reference validator | `0` only for a closed, unique, fully resolved active catalog | Bounded catalog summary | Bounded configuration diagnostic | Catalog and verification semantic digests plus schema summaries | Unknown field, unresolved reference, zero-row owner, recycled ID, or invalid ordering passes | no child work |
| TH-HARNESS-AC-063 | Sections 3, 10 | Runner selectors and profiles | Exact Go, Vitest, Playwright, and shell selectors plus zero, multiple, overlap, traversal, symlink, glob, regex, and unknown-profile fixtures | `test-catalog-check` selector/profile validation | `0` only when every case resolves exactly once and every profile is compatible | Bounded catalog summary | Bounded selector/profile diagnostic | Selector-resolution and profile-resolution report | Arbitrary executable, ambiguous selector, inline resource override, or symlink escape passes | no child work |
| TH-HARNESS-AC-064 | Sections 4, 5, 7 | Owner command inputs and output | Missing/valid owners; omitted/blank/duplicate/cross-owner rows; worker `0`, `1`, `16`, `17`; JSON omitted, empty, `1`, invalid; retired inputs | Public command contract fixtures for all five owner commands | Exact Section 5 success or exit `2` before setup | Human output or exact command JSON plus LF | Bounded usage diagnostic | Plan, explanation, guide, or audit artifacts exactly where required | Default-check narrows omitted rows, JSON and machine combine, or old input is ignored | no setup for rejected inputs |
| TH-HARNESS-AC-065 | Sections 9, 10 | Terminal row accounting | Pass, assertion, setup, dependency, cancellation, authorized/expired skip, duplicate/missing record, concurrent failure, and cleanup fixtures | Owner-slice scheduler matrix | Success only for passed or authorized-skipped rows; exact Section 9 exits otherwise | Bounded scheduler summary | Bounded primary diagnostic | One terminal record per resolved row, all attempts, finalizer evidence | Unauthorized skip passes, cleanup masks primary, retry hides failure, or selected row lacks record | finalizers always run |
| TH-HARNESS-AC-066 | Sections 6, 8, 16 | Evidence compatibility and freshness | Matching and mismatched source/catalog/verification/profile digests; missing, duplicate, extra, mixed, old, dirty, and unsafe roots | `test-evidence-audit` fixtures | `0` only for one compatible complete owner evidence set | Bounded audit result | Bounded usage or artifact diagnostic | `cartulary.test_evidence_audit_summary.v1` with used/unused/rejected roots | TTL, newest fallback, mixed snapshot, duplicate candidate, or broad-check inference passes | retained inputs unchanged |
| TH-HARNESS-AC-067 | Sections 3, 6, 15 | Executable-input and semantic boundaries | Neutral direct/joined restricted-root fixtures plus product-phase, execution-step, delivery-phase, and ambiguous fixtures | Executable-input policy and semantic identity scan | `0` only when executable sources and machine evidence contain no restricted input | Bounded policy summary | Bounded boundary or configuration diagnostic | Closed policy validation summary | Validator consumes a restricted root, machine evidence contains a documentation path, or a live selector retains delivery identity | no document inspection or mutation |
| TH-HARNESS-AC-068 | Sections 3, 4, 16 | Migration reconciliation | Frozen 456 backend, 87 frontend, and 5 Graph identities plus duplicate, missing, consolidation, deletion, and new-row fixtures | Baseline/crosswalk validator and final totals report | `0` only when all 548 identities have one terminal disposition and every new row has authorization | Bounded totals by source, disposition, and owner | Bounded reconciliation diagnostic | Baseline digest, crosswalk digest, assertion-preservation and owner-review evidence | Graph row omitted, old identity repeated, unauthorized row added, or deletion lacks owner proof | temporary crosswalk removed only after report retention |
| TH-HARNESS-AC-069 | Sections 3, 16 | Evidence-to-gate applicability | Owners with each evidence class, zero-row classes, informative measurement, claim measurement, and unknown target fixtures | Generated applicability matrix and owner audit | `0` only when every active row maps to exact required gates | Bounded applicability summary | Bounded routing diagnostic | Generated owner/evidence/gate matrix | Human applicability skip, informative evidence closes release, or private target satisfies gate | none |
| TH-HARNESS-AC-070 | Sections 1, 4, 15, 16 | Atomic v1 retirement | Current tree plus old target, variable, schema, artifact, reader, alias, dual-write, and delivery-identity fixtures | Task-surface parity, semantic scan, schema attachment validation, and repository reference scan | `0` only when no active v1 surface remains | Bounded retirement summary | Bounded compatibility diagnostic | Removed-identity report and v2 registry parity | Any v1 public command, reader, fallback, phase catalog, or ledger remains live | generated outputs regenerated through Make |
| TH-HARNESS-AC-071 | Sections 1-17 | Harness v2 parity and adoption | Revised NLSpec, schemas, catalog, topology, task surface, generated outputs, focused owner evidence, warm check, and release evidence | Editorial lint, `json-shape-check`, generated-policy/drift, harness contract, owner slices, `agent-finalize`, warm `check`, and `release-check` | Success only when every required target passes from one coherent v2 change set | Existing bounded summaries | Exact failing target diagnostic | Current v2 schemas, task surface, catalog, topology, run roots, finalizer, and release artifacts | Partial adoption, missing schema, generated drift, unresolved requirement, or historical evidence closes a gate | ordinary target cleanup succeeds |
| TH-HARNESS-AC-072 | Sections 2, 4 | Observability ownership and coverage | Complete public task surface; required, excluded, out-of-scope, omission, overlap, unknown, duplicate-sequence, unowned-reason, parameterized-profile, and application-boundary fixtures | Task-surface validation, harness contract, and OTel conformance | Every public target has exactly one explicit disposition, every required target has one canonical measurement profile, and every public testing entry point is required | Bounded coverage summary | Bounded owner or boundary diagnostic | Authored policy and generated projection digests | Default or target-name inference, omission, overlap, ownerless disposition, unstable input profile, duplicate sequence occurrence, or application scope widening passes | none |
| TH-HARNESS-AC-073 | Sections 4, 8 | Local observability lifecycle | Direct, aggregate, external-root, source-change, dirty, retry, failure, interruption, contamination, tamper, and partial-generation fixtures | Top-level context capture, observability finalizer, and explicit observability check | One retained context and trace per top-level required invocation; complete bundles validate independently of the checkout; partial bundles warn normally and fail the explicit check | Bounded result and artifact refs | Bounded artifact diagnostic | Context, index, per-invocation artifacts, and source digests | Current checkout changes historical identity, absolute path leaks, normal result changes, tampered evidence is rewritten, or partial output passes explicit validation | selected retained run byte-for-byte unchanged by the check |
| TH-HARNESS-AC-074 | Sections 8, 10 | Trace graph and timing accounting | Direct/nested sequence, scheduler dependency, all wait reasons, overlapping-child, blocker, service, runner, clock-skew, finalizer, failure, and interruption fixtures | Golden in-memory reconstruction and schema checks | Explicit parentage and links, interval unions, queue waits, resource blocking, dependency critical path, and unattributed time match exact expected values | Bounded trace summary | Bounded graph/accounting diagnostic | Native trace and hotspot bundle | Temporal containment invents parentage, child/finalizer durations are summed across overlap, dependency edges disappear, or scheduler envelope semantics change | scratch output removed |
| TH-HARNESS-AC-075 | Sections 8, 10, 15 | OTLP shape and privacy | Valid bundle plus hostile paths, commands, environment, credentials, SQL-like text, symbols, output, and error strings | OTLP decoding, allowlist validation, and redaction scan | Trace and metric payloads decode, match native identity, and contain only registered names/attributes | Bounded shape summary | Bounded privacy or shape diagnostic | OTLP trace/metric payload digests | Forbidden literal or unknown attribute reaches payload | scratch capture removed |
| TH-HARNESS-AC-076 | Sections 2, 4, 15 | Explicit exporter containment | Disabled ordinary runs, hostile `OTEL_*`, exact and ambiguous selection, valid HTTPS/loopback endpoints, invalid URLs, header permissions/shapes, redirect, timeout, and receiver failure | Fake collector and process-network fixtures | Ordinary runs make no telemetry request; explicit export sends exact payloads; configuration exits `2`; delivery exits `1` | Bounded count and endpoint class | Bounded configuration or exporter diagnostic with secrets absent | Fake collector request digests | Newest-run fallback, inherited env egress, redirect follow, timeout drift, header leak, failure-code collapse, or source mutation passes | fake collector stopped; source run unchanged |
| TH-HARNESS-AC-077 | Sections 8, 10 | Sequence and browser scheduling | Serial, parallel, simultaneous failure, dependency failure, interruption, generic resource contention, cycles, unknown dependencies, release dependency, and capacity-one/two browser fixtures | Shared scheduler matrix, generated topology, and browser lifecycle tests | DAG order, deterministic event bytes and failure order, capacity, summary projection, process-group cancellation, resets, isolated leaves, finalizers, and cleanup match the adopted contract | Existing bounded sequence/browser summaries | Existing bounded scheduler diagnostic | Scheduler v7 events, scheduler summaries, and leaf target summaries | Shell owns scheduler policy, release dependency starts early, more than declared stacks overlap, sibling leaks, or direct leaf shares a stack | all started stacks cleaned |
| TH-HARNESS-AC-078 | Sections 9, 10 | Backend and finalizer optimization parity | Compatible/incompatible direct compatibility keys, raw and isolated selectors, missing/extra/duplicate/partial Go JSON, capture and worker exceptions, concurrent failure, and emission fixtures | Backend target plan, runner fixtures, backend-unit, and artifact comparison | Process grouping and bounded concurrency reduce work while every row, partial success, artifact, and primary failure remains exact | Existing bounded target summary | Bounded row, artifact, or failure diagnostic | Before/after process and interval-union finalizer accounting | Family-level inference merges rows, raw selector merges, partial success drops, artifact order changes, or race masks failure | ordinary target cleanup |
| TH-HARNESS-AC-079 | Sections 4, 8, 10, 17 | Public-target performance acceptance | Target/provider windows with one discarded warm-up and exactly two measured roots; midpoint median and MAD; native, invocation, and exact aggregate timing; ordering, duplicate, failed, cold, retried, dirty, every provenance mismatch, missing-command, parameterized-profile, exact policy transition, four hotspots, 48 individual gates, portfolio-total gate, migration isolation, and independent check-window fixtures | Baseline writer, performance checker, retained-run inspection, and canonical check acceptance | Exact Section 10.5 formulas and cardinality pass for every required public command and hotspot, the portfolio total decreases, migration is reference-only, and TH-HARNESS-REQ-355 remains satisfied | Bounded performance summary | `duration_baseline_drift`, exit `13`, with every rejected-root reason | Verified root refs and bindings, timing sources, policy projections, medians, deviations, limits, portfolio delta, and verdicts | Hand-edited baseline, inferred/newest root, mismatched or failed run, summed finalizers, opaque policy change, blanket percentage, or moved required work passes | retained inputs unchanged; baseline writes only after complete validation |
| TH-HARNESS-AC-080 | Sections 5, 9, 17 | Versioned focus-continuity observation | Delayed mention mutation whose focused control unmounts, whose semantic Timeline row is initially offscreen, and whose projection delivers `N-1`, then response version `N`; instrumented postcondition helper | Focus-continuity model and React fixtures plus the exact live browser row through its canonical owner slice | One attempt succeeds only after the stable source row reaches at least `N` and the application restores the row-local focus target; helper interaction counters remain zero | Existing bounded unit/browser summary | Product assertion diagnostic with expected/rendered versions, target presence, active-element identity, mounted row IDs, and scroll geometry | Ordered lifecycle generations, response source identity/version, rendered version observations, and focus/viewport observation | `N-1` settles continuity, the helper focuses or scrolls after mutation, user interruption still permits focus stealing, or a later pass replaces the failed invocation | ordinary target cleanup; historical failed evidence remains unchanged |

### 17.1 Requirement-to-Acceptance Traceability

| Requirement range         | Owner section                      | Acceptance criteria                                     |
| ------------------------- | ---------------------------------- | ------------------------------------------------------- |
| `TH-HARNESS-REQ-001..049` | Status, scope, authority, owner model | TH-HARNESS-AC-013, TH-HARNESS-AC-015, TH-HARNESS-AC-016, TH-HARNESS-AC-026, TH-HARNESS-AC-029, TH-HARNESS-AC-062, TH-HARNESS-AC-063, TH-HARNESS-AC-066, TH-HARNESS-AC-071, TH-HARNESS-AC-072 |
| `TH-HARNESS-REQ-050..099` | Public command surface             | TH-HARNESS-AC-001, TH-HARNESS-AC-004, TH-HARNESS-AC-005, TH-HARNESS-AC-018, TH-HARNESS-AC-020, TH-HARNESS-AC-023, TH-HARNESS-AC-027, TH-HARNESS-AC-028, TH-HARNESS-AC-038, TH-HARNESS-AC-039, TH-HARNESS-AC-040, TH-HARNESS-AC-041, TH-HARNESS-AC-042, TH-HARNESS-AC-045, TH-HARNESS-AC-046, TH-HARNESS-AC-056, TH-HARNESS-AC-058, TH-HARNESS-AC-064, TH-HARNESS-AC-070, TH-HARNESS-AC-071, TH-HARNESS-AC-072, TH-HARNESS-AC-073, TH-HARNESS-AC-076, TH-HARNESS-AC-079 |
| `TH-HARNESS-REQ-100..149` | Configuration                      | TH-HARNESS-AC-002, TH-HARNESS-AC-003, TH-HARNESS-AC-021, TH-HARNESS-AC-028, TH-HARNESS-AC-029, TH-HARNESS-AC-064 |
| `TH-HARNESS-REQ-150..199` | Result roots and artifact identity | TH-HARNESS-AC-003, TH-HARNESS-AC-015, TH-HARNESS-AC-064, TH-HARNESS-AC-066, TH-HARNESS-AC-067, TH-HARNESS-AC-071 |
| `TH-HARNESS-REQ-200..249` | Output modes                       | TH-HARNESS-AC-004, TH-HARNESS-AC-005, TH-HARNESS-AC-023 |
| `TH-HARNESS-REQ-250..299` | Artifacts and schemas              | TH-HARNESS-AC-000, TH-HARNESS-AC-004, TH-HARNESS-AC-015, TH-HARNESS-AC-019, TH-HARNESS-AC-025, TH-HARNESS-AC-028, TH-HARNESS-AC-031, TH-HARNESS-AC-048, TH-HARNESS-AC-049, TH-HARNESS-AC-062, TH-HARNESS-AC-064, TH-HARNESS-AC-065, TH-HARNESS-AC-066, TH-HARNESS-AC-071, TH-HARNESS-AC-073, TH-HARNESS-AC-074, TH-HARNESS-AC-075, TH-HARNESS-AC-079 |
| `TH-HARNESS-REQ-300..349` | Failure and exit codes             | TH-HARNESS-AC-013, TH-HARNESS-AC-014, TH-HARNESS-AC-032, TH-HARNESS-AC-064, TH-HARNESS-AC-065, TH-HARNESS-AC-066, TH-HARNESS-AC-067 |
| `TH-HARNESS-REQ-350..399` | Scheduler                          | TH-HARNESS-AC-006, TH-HARNESS-AC-018, TH-HARNESS-AC-021, TH-HARNESS-AC-024, TH-HARNESS-AC-030, TH-HARNESS-AC-059, TH-HARNESS-AC-060, TH-HARNESS-AC-061, TH-HARNESS-AC-063, TH-HARNESS-AC-064, TH-HARNESS-AC-065, TH-HARNESS-AC-074, TH-HARNESS-AC-077, TH-HARNESS-AC-078, TH-HARNESS-AC-079 |
| `TH-HARNESS-REQ-400..449` | Services                           | TH-HARNESS-AC-007, TH-HARNESS-AC-010, TH-HARNESS-AC-017, TH-HARNESS-AC-033, TH-HARNESS-AC-049, TH-HARNESS-AC-056 |
| `TH-HARNESS-REQ-450..499` | Reset route                        | TH-HARNESS-AC-008, TH-HARNESS-AC-034, TH-HARNESS-AC-035, TH-HARNESS-AC-050, TH-HARNESS-AC-051, TH-HARNESS-AC-052, TH-HARNESS-AC-053, TH-HARNESS-AC-054, TH-HARNESS-AC-056 |
| `TH-HARNESS-REQ-500..549` | Cleanup                            | TH-HARNESS-AC-009, TH-HARNESS-AC-010, TH-HARNESS-AC-028, TH-HARNESS-AC-036 |
| `TH-HARNESS-REQ-550..599` | Platform                           | TH-HARNESS-AC-012                                       |
| `TH-HARNESS-REQ-600..649` | Security and redaction             | TH-HARNESS-AC-003, TH-HARNESS-AC-011, TH-HARNESS-AC-015, TH-HARNESS-AC-036, TH-HARNESS-AC-056, TH-HARNESS-AC-067, TH-HARNESS-AC-075, TH-HARNESS-AC-076 |
| `TH-HARNESS-REQ-650..699` | Product integration                | TH-HARNESS-AC-013, TH-HARNESS-AC-016, TH-HARNESS-AC-026, TH-HARNESS-AC-043, TH-HARNESS-AC-044, TH-HARNESS-AC-047, TH-HARNESS-AC-049, TH-HARNESS-AC-050, TH-HARNESS-AC-051, TH-HARNESS-AC-052, TH-HARNESS-AC-053, TH-HARNESS-AC-054, TH-HARNESS-AC-055, TH-HARNESS-AC-056, TH-HARNESS-AC-062, TH-HARNESS-AC-066, TH-HARNESS-AC-068, TH-HARNESS-AC-069, TH-HARNESS-AC-070, TH-HARNESS-AC-071, TH-HARNESS-AC-080 |

## 18. Sources and Evidence Limits

This section is traceability and evidence posture. It does not add current conformance behavior.

Primary repository evidence used to shape this NLSpec includes:

- `testing-harness-nlspec.md`, prior draft;
- `nlspec-spec.md`, NLSpec standard;
- Core 00 through Core 04 for product-conformance authority;
- Core 05 for claim-publication separation;
- `docs/domain.md` for vocabulary and owner navigation;
- implementation and testing guides for repository command-surface and harness context;
- `Makefile`, `tools/task_surface_manifest.json`, generated task-surface includes, scheduler manifests, and schema files when present in the repository.
- research reports R01 through R09 under `docs/research/` for state-boundary, handoff, stable-identity, cancellation, integrity, governance, extension-registry, and public-contract rationale only.

The imported standards named by TH-HARNESS-REQ-011 were evaluated for current status as of 2026-07-16. Their primary references are JSON Schema Draft 2020-12, RFC 8785 plus verified errata, RFC 3339, FIPS 180-4, NIST SP 800-218, OWASP Path Traversal, and CWE-22. Research and security guidance explain the chosen boundaries; only the exact imported algorithms and constraints stated in Sections 1 through 17 are normative.

The following evidence categories remain non-normative in this document unless promoted by a requirement above:

| Evidence category                                                | Current role                                                                      |
| ---------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| Recovery docs under `docs/testing-harness-spec-recovery-docs/**` | Historical traceability and diagnostic context.                                   |
| Raw package scripts and tool output                              | Developer convenience or child-command evidence.                                  |
| Script paths, generated Make include names, helper binaries, `internal_helper`, `check_internal`, priority-band names, and generator-only constants | Implementation details unless a requirement above explicitly promotes one. |
| Playwright screenshots, videos, traces, HTML reports             | Diagnostic secret-bearing artifacts.                                              |
| Hosted CI provider workflows                                     | Outside current conformance unless provider source is supplied and later adopted. |
| Visual snapshot refresh diagnostics                              | The helper workflow is current under Sections 6, 8, 11, 15, and 17; raw Playwright image/report internals remain non-normative. |

Exact numeric constants are normative only when they protect security, cleanup safety, bounded output, or deterministic scheduling. Other numeric values in generated manifests, helper names, priority bands, and generator-only constants are implementation details unless this NLSpec gives them a requirement.

The editorial lint for TH-HARNESS-AC-016 rejects the forbidden evidence markers listed in this non-normative section when they appear in Sections 1 through 17. The forbidden markers are: `TODO`, `source_limited`, `source-limited`, `source-observed`, `current code`, `selected evidence`, `recovery evidence`, and `maintainer_decision_required`.

## 19. Future Decisions Outside Current Conformance

The items below are explicitly outside the current conformance profile. They do not block implementation of the current harness contract.

Adoption of `cartulary.testing_harness.current.v2` adopts only Sections 1 through 17 and the current conformance rules explicitly listed there. It MUST NOT adopt any Section 19 future area as current harness conformance, product conformance, provider-specific hosted CI behavior, Playwright diagnostic schema stability, or Core 05 claim-publication evidence. The helper-only visual refresh contract is current only to the extent defined in Sections 6, 8, 11, 15, and 17.

| Future area                                                 | Current treatment                    | Future adoption requirement                                                                  |
| ----------------------------------------------------------- | ------------------------------------ | -------------------------------------------------------------------------------------------- |
| macOS certification                                         | Unsupported for current conformance. | Add platform profile, exact toolchain matrix, and acceptance evidence.                       |
| Windows-native support                                      | Unsupported for current conformance. | Add platform profile separate from WSL2.                                                     |
| Podman/Podman Compose                                       | Unsupported for current conformance. | Add service fixture compatibility profile and cleanup proof.                                 |
| Hosted CI annotations/uploads/artifact-retention dashboards | Provider-neutral `make ci` only.     | Add provider workflow source and provider-specific contract.                                 |
| Playwright report/trace/video/screenshot and visual-geometry diagnostic schemas | Diagnostic-only.                     | Adopt exact Playwright version/schema family or wrapper schema.                              |
| Benchmark-publication harness integration                   | Not part of harness conformance.     | Add Core 05-compatible benchmark manifest and claim-publication profile.                     |
| Scheduler, cross-run helper, or product test-result artifact reuse | Not adopted; selected work still executes in local `check`. Current same-run helper refs are limited to helper/setup artifacts produced under the same run root and do not count as scheduler reuse. | Define reusable artifact cache provenance, allowed work-unit classes, retained artifact schema, reused accounting, bypass controls, CI policy, revised TH-HARNESS-AC-018 behavior, and explicit exclusions for security, drift, services, cleanup, destructive safeguards, browser/live-state work, and runtime reset. |
