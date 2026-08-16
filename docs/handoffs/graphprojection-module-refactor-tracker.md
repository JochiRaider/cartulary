# Graph Projection Production Activation Tracker and Handoff

## 1. Controlling Posture

- **Target subsystem:** `internal/modules/graphprojection`
- **Controlling artifact:** `docs/handoffs/graphprojection-module-refactor-tracker.md`
- **Iteration:** GP2, production retained-Graph activation and v1 retirement
- **Status:** `COMPLETE`; GP2-S00 through GP2-S11 are complete and no slice is
  active
- **Planning baseline:** clean commit
  `df30974f44b18f0667498ef90555fd17846444b9`
- **Prior iteration:** GP-S00 through GP-S08 completed at `df30974f`
- **Implementation baseline:** commit
  `df30974f44b18f0667498ef90555fd17846444b9` plus the pre-existing staged
  tracker revision; no other worktree change existed when GP2-S00 began
- **Executed order:**
  `GP2-S00 → GP2-S01 → GP2-S02 → GP2-S03 → GP2-S04 → GP2-S05 → GP2-S06 → GP2-S07 → GP2-S08 → GP2-S09 → GP2-S10 → GP2-S11`

This tracker controls the user-authorized implementation. Planned behavior does
not become conformance authority until its applicable owner is adopted. Every
slice is a separate workstream, and the tracker checkpoint for one slice MUST
complete before the next slice begins.

The authority order remains:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 within their named scopes.
3. Versioned machine projections of those owners.
4. Current repository implementation and tests as evidence.
5. This tracker as sequencing, decision, validation, and handoff control only.

`docs/research/nlspec-spec.md` supplied writing guidance for behavioral
completeness, unambiguous interfaces, explicit defaults, conceptual fidelity,
spec economy, recreatability, and binary acceptance criteria. It did not supply
Graph Projection requirements and MUST NOT be treated as an instruction source
or runtime authority.

### Statement classes

| Class | Meaning | Effect in this tracker |
| --- | --- | --- |
| Current fact | Verified owner or repository state at `df30974f` | Baseline evidence; changes require an adopted owner and passing implementation evidence |
| Planned decision | User-authorized target for GP2 | Directs owner drafting and slice design but is not conformance authority before adoption |
| Adopted requirement | Text later accepted in a named owner | Governs implementation and machine projections after its adoption slice |
| Completion evidence | Terminal command result, run root, artifact, or operational proof | Permits a slice to move from `IN PROGRESS` to `DONE` |

## 2. Prior Iteration Baseline

GP-S00 through GP-S08 completed the first Graph Projection remediation. That
history is frozen at commit `df30974f` and is summarized here instead of being
repeated as a second active plan.

| Prior outcome | Completed state |
| --- | --- |
| Focused evidence routing | All direct Graph contract tests and fixture wrappers have owner-routed terminal evidence |
| PostgreSQL store structure | The retained store was decomposed by concern without changing behavior |
| Root facade structure | Ephemeral, retained, query, invalidation, and error concerns were separated by file |
| Helper-surface review | Adapter exports were preserved because narrowing them would have introduced cycles or duplication |
| Ordinary production posture | The full retained store remained explicitly dormant and guarded against accidental composition |
| Recovery contracts | Graph-owned registry, binding, result, protected evidence, and exact legacy empty-registry support were added |
| Recovery execution | A separate Graph participant clears the five legacy derived tables and rebuilds typed registry candidates |
| Final validation | `make check` passed 767/767 at run root `20260815T172858Z-p1576876` |

The final prior Markdown checkpoint passed at run root
`20260815T173839Z-p1755608`. The first iteration introduced no ordinary retained
producer, Graph route, Graph cursor configuration, Graph worker, or database
migration. Its current source registry is exactly empty, so current restore
behavior is successful clear-only.

## 3. Current Facts and Production Gaps

| ID | Current fact | Production or maintenance risk | GP2 disposition |
| --- | --- | --- | --- |
| GP2-G01 | Network Flow is the only active projection consumer and calls `project_ephemeral` | The large retained product has no producer or complete user workflow | Make Network Flow the source owner of incident-shared saved-graph declarations |
| GP2-G02 | Network Flow v1 explicitly forbids retained Graph views | Activating the store without owner adoption would contradict the current profile | Adopt the later retained lifecycle, authorization, cleanup, job, recovery, and UI contract before code |
| GP2-G03 | No production caller constructs `postgresstore.Store` | Store behavior is integration-tested but operationally dead | Replace it with narrow completed-result publication, read/traversal, and lease adapters |
| GP2-G04 | The retained facade returns an accepted summary only after synchronous computation has already finished | Attempt state is misleading and process failure can strand `accepted` or `computing` rows | Move attempt lifecycle, retry, cancellation, and crash recovery to common jobs |
| GP2-G05 | Reporting binds random `projection_run_id` values | Recovery intentionally mints fresh run IDs, so rebuilt results cannot preserve durable Reporting references | Bind consumers to deterministic content-derived `projection_result_id` |
| GP2-G06 | Graph owns a private cursor codec, raw cursor key, lifecycle idempotency, and per-run invalidation | These duplicate platform and Network Flow capabilities and enlarge the security surface | Reuse Network Flow cursor protection, common route idempotency, source-owner retirement, and common jobs |
| GP2-G07 | `projection_input` contains caller-supplied `requested_at`, `requested_by`, no-op `custom_config`, and retention fields that do nothing for ephemeral calls | Trust, digest, and compatibility boundaries are unnecessarily broad | Remove them from v2 input; keep actor, time, job, and retention policy in owning application contracts |
| GP2-G08 | Graph v1 retains non-emitted historical aliases and broad exported adapter/test seams | Dead vocabulary and test-only hooks burden every future revision | Remove aliases, hooks, old DTOs, old errors, and adapter exports after all consumers cut over |
| GP2-G09 | The restore registry is empty and the legacy compatibility path is open-ended | Production activation cannot reconstruct saved graphs; indefinite compatibility blocks schema cleanup | Add one typed Network Flow registration and retire the exact old bridge after evidence-gated rollout |
| GP2-G10 | Reporting validates Graph bindings but renders placeholder labels instead of selected Graph data | The nominal consumer does not deliver the adopted product value | Resolve exact vertices/edges, apply selection and redaction, and render real Graph diagrams |
| GP2-G11 | The Network Analysis UI offers exploratory ephemeral graphs only | There is no saved-graph discovery, job, lifecycle, or Reporting handoff workflow | Add a complete accessible saved-graph workflow while preserving unsaved exploration |
| GP2-G12 | Graph limits permit very large in-memory inputs and output collections | Activation without bounded concurrency, cancellation, and performance evidence risks resource exhaustion | Bind the first producer to Network Flow limits and require operational load and fault evidence |

### Compatibility classification

| Surface | Classification | Continuing value | Planned end state |
| --- | --- | --- | --- |
| Ephemeral Network Flow Graph behavior | Migrate | Active user-facing behavior | Preserve semantics through a deliberate v2 contract change and updated generated clients |
| Deterministic projection algorithms | Carry forward | Core reusable capability | Keep behind a cohesive pure projection boundary |
| Reporting Graph selection | Replace incomplete implementation | Material future product value | Bind to stable v2 results and render actual selected data |
| Retained Graph lifecycle facade | Delete | No production caller and duplicates common jobs | No v2 create/refresh facade or accepted/computing Graph state |
| `postgresstore` | Delete | No production composition | Narrow v2 result adapters only |
| Graph cursor and cursor key | Delete | Network Flow already owns rotated cursor protection | No Graph-specific deployment secret |
| Graph idempotency table | Delete | Common route and job idempotency already exist | No v2 Graph idempotency table |
| Per-run invalidation and configurable retention | Delete | No adopted producer needs the generic surface | Source-owner retirement plus reachability/lease-based result cleanup |
| v1 runtime parser and output | Delete | A dual path would permanently couple the production foundation to pre-activation bytes | v2 only after in-repository consumers cut over |
| Exact pre-activation backup bridge | Temporarily preserve | Protects the bounded empty-registry backup posture during rollout | Delete only after GP2-S10 operational evidence passes |
| v1 source and fixtures | Delete after cutover | Git retains historical evidence | No build, test, generator, or runtime dependency |

## 4. Planned Product and Architecture Contract

Everything in this section is a planned GP2 decision until GP2-S01 adopts it in
the applicable owner documents.

### 4.1 Ownership and composition

- Graph Projection owns deterministic admission, normalization, derivation,
  completed-result identity, canonical output, publication invariants, exact
  result reads, bounded traversal, and recovery reconstruction semantics.
- Network Flow owns saved-graph declarations, source enumeration, source
  validity, incident authorization, public routes, common-job admission,
  display names, semantic queries, source-table consequences, auditing, and the
  browser workflow.
- Reporting owns release-tuple admission, source-owner binding dispatch,
  selection rules, result leases for active render work, redaction, diagram
  materialization, release output, and its public failures.
- Recovery owns admitted operation identity, backup selection, participant
  ordering, serving readiness, terminal evidence, replay, and indeterminate
  target handling.
- Common jobs own materialization attempt identity, queueing, retry,
  cancellation, execution lease, progress, terminal state, and route replay.
- Application assembly injects borrowed database and common-job capabilities.
  Constructors MUST start no hidden worker and MUST close no borrowed resource.
- Graph Projection remains authorization- and transport-agnostic. No public
  Graph HTTP or WebSocket family is introduced.

### 4.2 Graph Projection v2

The runtime contract becomes exactly `graph_projection.v2`. The repository will
not ship a v1 runtime reader, writer, translator, feature flag, or content
negotiation path.

V2 removes these v1 members and concepts:

- `requested_at` and `requested_by` from projection input;
- `projection_config.retention_policy`;
- `projection_config.custom_config`;
- accepted/computing run state and accepted-run responses;
- caller idempotency within Graph Projection;
- list-view cursors and Graph-owned cursor cryptography;
- graph-view and projection-run invalidation operations;
- failed-run inspection and configurable retained-run history;
- non-emitted historical validation and lifecycle aliases.

Operation time, authenticated actor, job identity, request identity, source-owner
identity, and cleanup policy are explicit invocation context owned by the
calling application. They do not enter normalized projection input or its
source/configuration digests unless an adopted identity table explicitly names
them.

`projection_result_id` is a lowercase generated identifier with prefix
`gpres_`. Its suffix is SHA-256 over the v2 length-framed tuple:

```text
graph_view_id
source_owner_id
source_snapshot_id
projection_version
normalized_configuration_sha256
normalized_source_sha256
canonical_output_sha256
```

The tuple prefix and framing algorithm are adopted once in the v2 Graph owner.
Server timestamps and common `job_id` do not participate. The same admitted
input and output under the same projection version MUST reproduce the same
result ID during retry and Recovery. Publishing an existing ID with different
bytes or digests is an invariant failure, never an overwrite.

### 4.3 Authoritative and derived storage

Migration `00032` is additive and creates:

- authoritative `network_flow_graph_views`;
- derived `graph_projection_results`;
- derived `graph_projection_result_vertices`;
- derived `graph_projection_result_edges`;
- operational derived `graph_projection_result_leases`.

An authoritative Network Flow graph-view declaration contains, at minimum:

- stable `graph_view_id` and Graph view key;
- `incident_id`, normalized display name, and declaration state;
- canonical semantic query with explicit table IDs and its SHA-256 digest;
- desired Network Flow source-snapshot identity;
- selected v2 result identity and its exact binding tuple when available;
- declaration version, creator, created time, and updated time;
- latest materialization job reference and closed last-failure classification.

Only `active` and `retired` are authoritative declaration states. Pending,
running, failed, and succeeded are common-job/materialization outcomes, not
Graph view states. A failed refresh preserves the prior selected result. A
retired declaration is excluded from ordinary reads and recovery candidates.

Migration `00033` runs only after the v2 cutover. It fails before destructive DDL
unless all five legacy Graph tables and every persisted v1 Reporting Graph
reference are empty. It then drops:

```text
graph_projection_edges
graph_projection_idempotency
graph_projection_runs
graph_projection_vertices
graph_projection_views
```

No row translator, inferred binding, compatibility view, trigger mirror, or
dual write is permitted. Non-empty state blocks the cutover and requires an
owner-approved remediation outside this plan.

### 4.4 Network Flow public contract

The planned Network Flow route family is:

```text
GET|POST /api/v1/incidents/{incident_id}/network-flow/graph-views
GET|PATCH|DELETE /api/v1/incidents/{incident_id}/network-flow/graph-views/{graph_view_id}
POST /api/v1/incidents/{incident_id}/network-flow/graph-views/{graph_view_id}/refresh
GET /api/v1/incidents/{incident_id}/network-flow/graph-views/{graph_view_id}/result
POST /api/v1/incidents/{incident_id}/network-flow/graph-views/{graph_view_id}/contributors/query
```

The visibility and mutation policy is closed:

| Operation | Authorized roles | Required mutation identity | Result |
| --- | --- | --- | --- |
| List, get, result, contributors | Any current incident member | None | Incident-shared visible active declarations only |
| Create | `editor`, `admin` | `client_txn_id` | `202` common job |
| Rename | `editor`, `admin` | `client_txn_id`, `base_graph_view_version` | Updated declaration or exact replay |
| Refresh | `editor`, `admin` | `client_txn_id`, `base_graph_view_version` | `202` common job |
| Retire | `reviewer`, `admin` | `client_txn_id`, `base_graph_view_version` | Retired declaration or exact replay |

Create persists a normalized display name and the canonical current semantic
query. Any `all_active_tables` exploratory scope is materialized to its exact
ordered table IDs before persistence. Duplicate display names are allowed;
ordering uses normalized display name, then `graph_view_id`. List and
contributors continuation use the existing Network Flow cursor protector and
common paging envelope.

Create and refresh enqueue job kind
`network_flow_activity.graph_view_materialize_v1` for worker kind
`network_flow_activity.graph_view_worker_v1`. The common job success contains one
resource reference with kind `network_flow_graph_view`, ID equal to
`graph_view_id`, and the canonical same-origin Network Flow route.

The worker reads and validates the declaration, constructs v2 input through the
Network Flow source adapter, computes outside a database transaction, and then
atomically publishes the result and compare-and-swaps the declaration. Lost
responses and retried handlers converge on the exact selected result. Source
table soft deletion or lost membership prevents ordinary result exposure and
invalidates affected active declarations through Network Flow-owned behavior.

### 4.5 Reporting v2 consumption

`source_projection_ref.v2` replaces `projection_run_id` with
`projection_result_id` and adds `source_owner_id`. It carries exactly the Graph
view, result, source-snapshot, projection-version, configuration-digest,
source-digest, and output-digest binding needed for immutable consumption.

Reporting resolves refs through a typed source-owner registry. The current
registry contains exactly the Network Flow owner. Within release admission and
the caller-owned transaction, that adapter proves incident ownership,
declaration visibility, selected-result identity, immutable source boundary,
and every digest. It acquires a result lease before the render job can observe
the payload.

The render path reads the exact leased result and implements the adopted
`explicit_refs`, `neighborhood`, and `all_with_bounds` selection rules using
Graph output order. Missing or duplicate objects, stale owner bindings, digest
mismatch, limit overflow, or lease loss fail closed through Reporting's Graph
failure family. Existing Reporting redaction is applied before diagram text or
artifacts are produced. The current label-only Graph placeholder is deleted.
Terminal render handling releases the lease only after the durable Reporting
outcome no longer depends on Graph rows.

### 4.6 Recovery and compatibility

The Network Flow backup participant adds graph-view declarations as
authoritative state. The Graph restore registry changes from empty to one
active typed Network Flow registration. Recovery restores authoritative state
first, then clears and reconstructs the four v2 Graph-derived tables in
canonical order before common workers resume.

Active declarations with a selected completed result at the backup consistency
point are rebuild candidates. Retired declarations and declarations without a
published result are not reconstructed. Every rebuilt result MUST reproduce its
original `projection_result_id`, digests, counts, vertex IDs, and edge IDs.
Mismatch fails readiness; it is never repaired by changing an authoritative
reference.

The only temporary historical allowance is the exact packaged pre-activation
empty-registry backup binding. Under the v2 schema it performs clear-only over
the four current derived tables, enumerates no v1 input, creates no Network Flow
declaration, and admits no other historical digest. GP2-S10 removes it only
after all of these are recorded:

1. A v2 deployment produced a fresh backup.
2. That backup passed isolated restore and readiness verification.
3. The supported backup and journal inventory contains no artifact requiring
   the old binding.
4. The rollback decision for the prior binary/backup pair is explicit.
5. The regenerated recovery binding and post-removal restore evidence pass.

## 5. Workstream and Slice Plan

### Workstream dependencies

Each implementation slice is its own workstream. The dependency chain is
strictly linear so owner, contract, persistence, runtime, consumer, recovery,
cutover, UI, and operational evidence cannot overtake one another.

| Workstream | Goal | Depends on | Feeds |
| --- | --- | --- | --- |
| GP2-S00 | Baseline, compatibility, and rollback evidence | Prior GP-S08 | GP2-S01 |
| GP2-S01 | Owner adoption and contradiction closure | GP2-S00 | GP2-S02 |
| GP2-S02 | V2 machine projections and generated contracts | GP2-S01 | GP2-S03 |
| GP2-S03 | Additive v2 storage and narrow adapters | GP2-S02 | GP2-S04 |
| GP2-S04 | Pure Graph v2 engine and result identity | GP2-S03 | GP2-S05 |
| GP2-S05 | Network Flow producer, jobs, routes, and contributions | GP2-S04 | GP2-S06 |
| GP2-S06 | Reporting exact Graph consumer | GP2-S05 | GP2-S07 |
| GP2-S07 | Backup and Recovery v2 | GP2-S06 | GP2-S08 |
| GP2-S08 | V1 runtime and schema cutover | GP2-S07 | GP2-S09 |
| GP2-S09 | Saved-graph browser workflow | GP2-S08 | GP2-S10 |
| GP2-S10 | Production hardening and compatibility retirement | GP2-S09 | GP2-S11 |
| GP2-S11 | Final validation and handoff | GP2-S10 | None |

### Slice tracker

| Slice | Status | Depends on | Completion outcome | Next action |
| --- | --- | --- | --- | --- |
| GP2-S00 | `DONE` | Prior GP-S08 | Baseline and every legacy surface are classified with rollback evidence | None |
| GP2-S01 | `DONE` | GP2-S00 | Owner documents adopt one coherent v2 contract | None |
| GP2-S02 | `DONE` | GP2-S01 | Machine contracts and generated projections express v2 exactly | None |
| GP2-S03 | `DONE` | GP2-S02 | Additive v2 storage and narrow adapters pass atomicity tests | None |
| GP2-S04 | `DONE` | GP2-S03 | Pure v2 projection and deterministic completed-result identity pass | None |
| GP2-S05 | `DONE` | GP2-S04 | Network Flow declarations, jobs, routes, and source contribution pass | None |
| GP2-S06 | `DONE` | GP2-S05 | Reporting renders exact redacted Graph selections | None |
| GP2-S07 | `DONE` | GP2-S06 | V2 backup and Recovery rebuild exact stable results | None |
| GP2-S08 | `DONE` | GP2-S07 | V1 runtime and legacy storage code are absent | None |
| GP2-S09 | `DONE` | GP2-S08 | Complete accessible saved-graph UI and browser workflows pass | None |
| GP2-S10 | `DONE` | GP2-S09 | Production bounds pass and the historical bridge is removed after isolated v2 restore evidence | None |
| GP2-S11 | `DONE` | GP2-S10 | Full validation and final handoff are complete | None |

### GP2-S00 — Baseline and retirement evidence

- Rebaseline this tracker to the exact implementation commit used for work.
- Re-audit every Graph production consumer, exported symbol, legacy table,
  Reporting v1 ref, recovery binding, journal decoder, fixture family, generated
  contract, deployed backup reference, and configuration surface.
- Query only approved disposable/local environments for table and reference
  emptiness. Do not infer production emptiness from source composition.
- Record the exact prior binary and verified backup pair available for rollback.
- Classify every surface as `migrate`, `replace`, `delete`, or
  `temporarily_preserve`; no `unknown` item may enter GP2-S01.

Exit: the compatibility inventory is complete, non-empty legacy state blocks
the destructive cutover path, and the tracker records commands, results, risks,
and the owner decision required for every exception.

#### GP2-S00 completion evidence

The repository baseline is exact. No deployment connector or production data
plane is attached to this workspace, so this execution makes no claim about an
external database. That is a closed `not_applicable_to_repository_execution`
classification, not inferred emptiness: migration `00033` MUST retain its
transactional zero-state preflight, and a deployer MUST capture the exact
pre-cutover backup before applying it.

| Inventoried surface | Evidence at implementation baseline | Classification |
| --- | --- | --- |
| Active Graph producer | Network Flow alone constructs the ephemeral service; no production code constructs `postgresstore.Store` | `migrate` |
| Durable consumer | Reporting decodes `source_projection_ref.v1` and reads a same-transaction binding through `postgresbinding` | `replace` |
| Recovery | Current registry is exactly `entries=[]`; binding ID is `graphprojection.restore_rebuild.current_empty_registry.v1` with registry digest `2d3718cc1f88f809ca4ea5c4efedc1a8108c8ef28b95ad939b0f7666429f7fe1` | `temporarily_preserve` |
| Legacy SQL | Migration `00022` owns the five v1 tables; authored SQL and service-backed tests are the only retained-store users | `delete` after preflight |
| Generated consumers | Network Flow schemas/generated TypeScript, OpenAPI, Graph fixtures, Reporting contracts, and Recovery artifacts contain v1 shapes | `migrate` or `replace` through owner generation |
| Configuration | No production Graph cursor key, retained-store flag, or hidden worker configuration exists | `delete`; introduce no replacement Graph secret |
| Migration numbering | Repository head already contains `00030_evidence_upload_leases.sql` and `00031_evidence_blob_cleanup_claims.sql` | Reserve `00032` and `00033` |
| Extension namespace | Claimed profile ID is `network_flow_activity`; extension job kinds are profile-prefixed | Use `network_flow_activity.graph_view_materialize_v1` and `network_flow_activity.graph_view_worker_v1` |
| Runtime rollback | Exact pre-cutover source/binary is commit `df30974f44b18f0667498ef90555fd17846444b9`; repository fixture evidence uses the current empty-registry backup contract | Before `00033`, use the matching old binary only with its compatible schema; after `00033`, restore the exact pre-cutover backup into a replacement target before starting that binary |
| External deployment state | No external environment is connected or authorized in this repository task | `not_applicable_to_repository_execution`; deployment preflight remains mandatory |

Commands used for the inventory were read-only `git status`, `git rev-parse`,
`git diff`, migration filename enumeration, exact `rg` searches over runtime,
contracts, generated consumers, recovery artifacts, and extension profile
declarations, plus `jq` inspection of the packaged registry and binding. No
unknown repository surface remains. Non-empty external legacy state is an exact
deployment blocker, never a compatibility-translation trigger.

### GP2-S01 — Adopt owner specifications

- Adopt Graph Projection 2.0 and the next Network Flow revision.
- Amend Reporting, Report Composition, Core 00/Core 01/Core 04, extension
  discovery, common-job resource vocabulary, backup/recovery ownership, and
  applicable operating guidance.
- Define every v2 member, variant, default, bound, role, state, error mapping,
  digest, ordering rule, transaction boundary, job transition, lease rule,
  restore outcome, UI state, and legacy cutoff exactly once in its owner.
- Remove v1 behavior from current owner text rather than appending contradictory
  exceptions. Historical rationale belongs in the tracker or an appendix.
- Run the two-implementer, recreatability, Definition-of-Done trace, and Core
  projection re-audit checks.

Exit: owners are adopted, current-profile extension declarations are coherent,
no contradiction remains, and `make lint-markdown` passes.

#### GP2-S01 completion evidence

Graph Projection 2.0.0 now owns pure deterministic computation, immutable
result/object identity, narrow result capabilities, leases, four-table Recovery,
and the exact rollout bridge boundary. Network Flow 3.0.0/state 2 owns saved
declarations, routes, roles, idempotency, materialization generations, Common
Jobs, quotas, UI states, Reporting participation, backup, and Recovery source
state. Reporting 1.3.0 and Report Composition 1.2.0 consume
`source_projection_ref.v2`, exact leased result data, and post-redaction graph
models. Extensions 0.8.0 and Core 00/01/04 project major 3, state migration,
worker/job/participant/rebuild, and Recovery v2 ownership. Domain vocabulary and
rollout guidance distinguish declarations, results, attempts, and diagrams.

The two-implementer review has one source-owner implementation path and one
pure-engine path with no shared lifecycle ambiguity. The recreatability review
found every new route, role, member, state, digest, job, lease, quota, rebuild,
and rollback decision in an adopted owner. The Definition-of-Done trace maps
Graph §11, Network Flow §23, Reporting graph acceptance, and Extensions generic
admission to the planned verification families. The Core projection re-audit
keeps workbook projection rebuild separate from
`graphprojection.restore_rebuild.v2`.

| Command | Result | Evidence |
| --- | --- | --- |
| Affected-owner `make task-guide` and `make explain-test-owner` commands | PASS | Graph 6 rows, Network Flow 60, Reporting 12, Recovery 26, web Network Flow 31 |
| Owner contradiction searches | PASS | No current v1 Graph owner/ref/rebuild or Network Flow major-2 declaration remains outside baseline handoff history |
| `git diff --check` | PASS | No whitespace errors before the Markdown gate |
| `make lint-markdown` | PASS | Run root `20260815T220136Z-p1824229` |

Product tests, generation, migration, service-backed, browser, finalizer, and
broad gates were skipped because GP2-S01 changes owner and supporting documents
only. They become mandatory in the slices that project or implement these
owners.

### GP2-S02 — Machine contracts and generation

- Add authored v2 Graph input, result, binding, error, registry, Recovery, and
  fixture projections.
- Add Network Flow graph-view route, resource, mutation, job, cursor binding,
  source contribution, backup family, state-presence, and generated-client
  projections.
- Add Reporting `source_projection_ref.v2` and exact source-owner registry
  projections.
- Add new owner-routed unit, integration, browser, accessibility, stateful, and
  measurement rows before relying on focused evidence.
- Runtime code consumes generated or embedded machine projections and never
  reads Markdown.

Exit: strict-shape, canonical digest, ordering, selector resolution,
`make generate`, drift, generated-policy, JSON-shape, and boundary checks pass.

#### GP2-S02 completion evidence

Graph Projection is now an active generated contract family with strict v2
semantic-input/result schemas and a deterministic empty-result golden fixture.
Network Flow projects contract major 3, the complete eight-operation saved-graph
route family, the closed saved-graph resource and mutation schemas, saved-graph
errors, and the v2 nested ephemeral result. Its extension projection now owns
state version 2, the exact `1 → 2` migration, five authoritative backup
families, the materialization job/worker/resource tuple, Reporting
participation, and the derived-state rebuild contribution.

Reporting's public contract now carries the exact v2 result tuple and no v1 run
reference. Recovery publishes a one-entry Network Flow source registry, a
four-table v2 rebuild binding/result contract, and a digest-distinct exact-v1
historical bridge. Generated Go, TypeScript, OpenAPI, extension, import,
Network Flow, and Recovery projections were recreated through `make generate`.
The standard browser client admits only Network Flow major 3; major 2 is not a
supported compatibility path.

The JSON-shape verifier was advanced with the contract instead of weakened: it
now requires the Graph Projection family, Network Flow major 3, all eight new
routes, all six new error families, and correct recursive closure handling for
JSON Schema `properties` maps. Authored test-family routing covers Graph,
Network Flow, Extensions, Reporting, Recovery, and the browser's generated
major-selection boundary.

The first focused run exposed two related transitional conditions. The old
Graph runtime Recovery row consumed the newly projected current v2 constants
before GP2-S07, so the v2 contract assertion was moved into a dedicated
contract-projection row rather than prematurely activating Recovery runtime.
The first Network Flow and Recovery focused builds also found test-helper name
collisions; both were corrected. The first JSON-shape runs correctly failed on
the major-2, 11-route, and 42-error verifier inventory; the verifier was updated
to the adopted major-3 closure. No unrelated failure remains.

| Command | Result | Evidence |
| --- | --- | --- |
| Affected-owner `make task-guide` commands | PASS | Graph, Network Flow, Extensions, Reporting, and Recovery selected focused owner slices; `web.networkflow` reports 31 Vitest rows |
| Focused Graph contract row | PASS | `20260815T222848Z-p1863491` |
| Focused Network Flow contract row | PASS | `20260815T222848Z-p1863495` |
| Focused Extensions contract row | PASS | `20260815T222650Z-p1852171` |
| Focused Reporting contract row | PASS | `20260815T222650Z-p1852186` |
| Focused Recovery contract row | PASS | `20260815T222848Z-p1863507` |
| Focused browser major-selection row | PASS | `20260815T223622Z-p1912225` |
| `make generate` | PASS | `20260815T223420Z-p1899594` |
| `make generate-drift` | PASS | `20260815T223434Z-p1902620` |
| `make generated-artifact-policy-check` | PASS | `20260815T223434Z-p1902621` |
| `make json-shape-check` | PASS | `20260815T223434Z-p1902653` |
| `make frontend-import-boundary-check` | PASS | `20260815T223434Z-p1902964` |
| `make format` | PASS | `20260815T223614Z-p1908695` |
| `git diff --check` | PASS | No whitespace errors before the tracker checkpoint |
| `make lint-markdown` | PASS | Final checkpoint run root `20260815T223726Z-p1913086` |

Database migration, storage atomicity, service-backed persistence, runtime
engine, producer, full Reporting render, isolated Recovery, browser workflow,
hardening, finalizer, and broad checks were skipped because their owning slices
remain GP2-S03 through GP2-S11. The compatibility posture remains deliberate:
v1 runtime is still present only until the coordinated cutover, but no major-2
browser client support or v1 payload translation was introduced.

### GP2-S03 — Additive v2 storage

- Add migration `00032` with `network_flow_graph_views` and the four v2 Graph tables.
- Keep authoritative declaration SQL in Network Flow and derived result SQL in
  a Graph child adapter.
- Implement narrow borrowed-transaction publication, exact-result read,
  vertex/edge read, bounded traversal, lease acquire/release, and reachability
  cleanup ports.
- Publish one completed result, vertices, and edges atomically. Re-publishing
  exact bytes is idempotent; same identity with different bytes fails.
- Preserve all five legacy tables unchanged in this additive slice.

Exit: migration head, empty upgrade, rollback, atomicity, duplicate,
referential, lease, and SQL-owner boundary tests pass.

#### GP2-S03 completion evidence

Migration `00032` now creates the authoritative
`network_flow_graph_views` table and four immutable derived Graph result tables.
The declaration deliberately has no result foreign key, so Recovery can restore
Network Flow authority before rebuilding Graph results. All five v1 Graph tables
remain present. The Down path removes only the new tables and trigger function;
a dedicated scratch-database test proves `32 → 31 → 32` while preserving all
legacy tables.

The Graph root exports only narrow v2 result capabilities and contract types.
The PostgreSQL child adapter publishes an envelope, vertices, and edges through
a borrowed transaction; exact byte replay is idempotent and a same-ID byte
mismatch is an invariant error. Separate query capabilities provide exact
binding reads, canonical ordered vertices/edges, bounded traversal, leases, and
reachability cleanup. Network Flow owns its declaration store, including
duplicate display-name ordering and an all-or-null selected-result binding.

Adding the migration exposed a Recovery catalog omission during generation.
The catalog now classifies the declaration as authoritative and the four result
tables as rebuildable derived state, while the PostgreSQL snapshot fixture adds
only the authoritative declaration. This increased the authored table inventory
from 113 to 118 and the required backup-table inventory from 83 to 84. The
Recovery v2 table names were also reconciled to their adopted owner names; the
exact historical v1 bridge remains unchanged for its later evidence-gated
retirement.

| Command | Result | Evidence |
| --- | --- | --- |
| Graph and Network Flow `make task-guide` and `make explain-test-owner` | PASS | Graph reports 8 rows/5 service-backed; Network Flow reports 61 rows/24 service-backed |
| Graph migration-head and result-adapter service rows | PASS | Run root `20260815T225449Z-p1939883` |
| Network Flow declaration persistence service row | PASS | Run root `20260815T225520Z-p1943272` |
| Graph v2 contract-projection row | PASS | Run root `20260815T225552Z-p1946663` |
| Migration rollback/reapply service row | PASS | Run root `20260815T225744Z-p1960700` |
| `make migration-drift` | PASS | Empty and penultimate scratch upgrade; run root `20260815T225552Z-p1946606` |
| `make backend-module-boundary-check` | PASS | Run root `20260815T225552Z-p1946779` |
| `make generate` | PASS | Final S03 generation run root `20260815T225732Z-p1957758` |
| `make format` | PASS | Run root `20260815T225725Z-p1954236` |
| `git diff --check` | PASS | No whitespace errors before the tracker checkpoint |
| `make lint-markdown` | PASS | Final checkpoint run root `20260815T225909Z-p1962472` |

Compatibility remains additive and reversible in this slice. No v1 runtime,
table, parser, or Reporting reference was removed, and no dual writer or
translation path was added. Producer/job semantics, Reporting lease use,
Recovery execution, v1 deletion, UI behavior, load/fault evidence, finalizer,
and broad checks remain deferred to their owning later slices. The remaining
S03 risk is therefore rollout-only: migration `00032` is ready but unused until
GP2-S05 composes the producer.

### GP2-S04 — Graph v2 engine and application boundary

- Add the v2 parser and pure projection service without a v1 translation path.
- Separate trusted invocation metadata from normalized projection bytes.
- Produce deterministic `projection_result_id`, output digest, validation
  identity, result ordering, and safe errors.
- Replace the optional-repository `Service` with explicit constructors and
  cohesive projection/result ports.
- Preserve the active ephemeral semantic result through the v2 Network Flow
  adapter while changing its declared schema intentionally.

Exit: v2 golden fixtures, retry determinism, cancellation, resource limits,
Unicode/JSON admission, output identity, safe errors, and Graph boundary tests
pass.

#### GP2-S04 completion evidence

The root Graph package now exposes a strict, pure `EngineV2` boundary. Trusted
`graph_view_id` and `source_owner_id` arrive through invocation context rather
than semantic JSON. Configuration and source bytes are independently
normalized, canonical output excludes recursive identity fields, and the
completed result ID uses unsigned 64-bit big-endian length framing over the
adopted identity transcript. The engine rejects removed and unknown fields,
orders rules and results deterministically, supports numeric normalization and
cancellation, and emits safe closed errors without source-authored values.

Network Flow's active unsaved graph route now derives its trusted graph-view ID
and returns the v2 result directly. No response translator or legacy response
parser was retained. The major-3 frontend availability fixture and Graph query
fixture were cut over with the generated client. Migration `00032`'s edge
direction constraint was corrected to the adopted `directed`, `undirected`,
and `bidirectional` values; the mixed-owner migration validation now assigns
Graph result objects to Graph Projection and declaration objects to Network
Flow.

| Command | Result | Evidence |
| --- | --- | --- |
| `make task-guide ROLE=module-author OWNER=module.graphprojection` | PASS | Focused Graph guidance selected unit and service-backed slices |
| `make task-guide ROLE=module-author OWNER=module.networkflow` | PASS | Focused Network Flow guidance selected unit and service-backed slices |
| `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_pure_determinism` | PASS | Run root `20260815T232439Z-p2053992` |
| `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58` | PASS | Run root `20260815T232617Z-p2060485` |
| `make test-slice OWNER=module.networkflow ROWS=module.networkflow.frontend_integration.verify_production_network_flow_grids_read_only_b_f662be335e` | PASS | Run root `20260815T232403Z-p2052499` |
| `make service-backed-test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.migration_reset,module.graphprojection.storage.result_v2_atomicity_and_queries` | PASS | Run root `20260815T232439Z-p2054030` |
| `make migration-drift` | PASS | Run root `20260815T232439Z-p2053949` |
| `make backend-module-boundary-check` | PASS | Run root `20260815T232617Z-p2060655` |
| `make frontend-import-boundary-check` | PASS | Run root `20260815T232617Z-p2060674` |
| `make generate` | PASS | Run root `20260815T232629Z-p2062023` |
| `make generate-drift` | PASS | Run root `20260815T232652Z-p2068759` |
| `make generated-artifact-policy-check` | PASS | Run root `20260815T232652Z-p2068751` |
| `make json-shape-check` | PASS | Run root `20260815T232652Z-p2068744` |
| `make format` | PASS | Run root `20260815T232643Z-p2065084` |
| `git diff --check` | PASS | No whitespace errors before the tracker checkpoint |
| `make lint-markdown` | PASS | Run root `20260815T232758Z-p2072860` |

Two deliberately broader probes failed at run roots
`20260815T231701Z-p2003336` and `20260815T231701Z-p2003334`. The Graph owner
probe reached the still-v1 restore runtime after the current generated binding
became v2; GP2-S07 owns that current Recovery implementation and the exact
historical dispatcher. The Network Flow probe reached the adopted state-v2 and
job declarations before their canonical worker, migrator, and producer are
composed; GP2-S05 owns those changes. These are forward-slice dependencies, not
failures of the S04 engine or ephemeral cutover. They remain visible and MUST
not be waived at their owning exits.

### GP2-S05 — Network Flow producer and common job

- Implement incident-shared declaration create/list/get/rename/refresh/retire
  behavior and route idempotency.
- Persist explicit table IDs and the canonical semantic query; never persist a
  visual label, cursor token, browser selection, or raw Graph input as authority.
- Register `network_flow_activity.graph_view_materialize_v1` and worker kind
  `network_flow_activity.graph_view_worker_v1` in the claimed profile's
  common-job catalog and runner composition.
- Compute outside a transaction, then publish the result, compare-and-swap the
  declaration, append owner audit/change evidence, and complete the job through
  the established atomic finalization pattern.
- Preserve a prior selected result on failed refresh. Detect stale declaration
  version, table soft deletion, source-snapshot change, cancellation, execution
  loss, retry, and lost terminal response without exposing partial output.
- Add the active Network Flow restore-source registration and liveness provider.

Exit: authorization, CSRF, replay, conflict, source validity, job lifecycle,
crash recovery, audit privacy, cursor, route, process composition, and
service-backed tests pass.

#### GP2-S05 completion evidence

Network Flow now owns an incident-shared saved-graph declaration lifecycle and
all eight major-3 operations. Mutations use Common route idempotency, require
the adopted roles and optimistic `base_graph_view_version`, enforce active,
retained, and nonterminal-job quotas, and retain stable IDs independently of
duplicate display names. Rename advances only `graph_view_version`; create and
refresh advance `materialization_generation`. Source-table deletion and
retirement clear ordinary result exposure without deleting immutable result
rows that a later Reporting lease may protect.

The claimed extension profile registers the exact Common Job and worker kinds.
The worker composes the authoritative source and projects outside a database
transaction, observes cancellation on the first check and at least every 1,024
processed checks, then atomically publishes immutable result data, performs a
generation/source/job compare-and-swap, appends terminal proof, and completes
the Common Job. A retried semantic refresh converges on the same result ID;
terminal route replay rehydrates the original accepted graph-view envelope
after Common Jobs replaces the retained idempotency payload with its terminal
resource. A failed or stale refresh cannot overwrite the prior selected result.

The Network Flow state owner now counts the fifth `graph_views` family and the
server composes the exact state `1 → 2` migration and v2 pending/final
validators. Verification exposed test wiring that still named the retired v1
validator: the broad state row timed out at
`20260816T000517Z-p2205829` because its concurrency evidence waited forever
after the migration-unavailable error. The fixture now exercises the real v2
migration and validator, the wait is bounded, and the Network Flow-specific
and concurrency cases are separate authored work units. The corrected full
state-admission matrix passes. This was related transitional test debt, not an
external or waived failure.

The contract error and effective-limit names were reconciled to the adopted
owner. The JSON-shape verifier was updated rather than weakening the route
contract. Two intermediate `make generate` runs failed on an unsorted new row
and overlapping selector (`20260816T001655Z-p2210913` and
`20260816T001743Z-p2214309`); both authored routing defects were corrected and
the final generation passed. JSON-shape runs at
`20260816T002258Z-p2248262` and `20260816T002319Z-p2249128` found the stale
error inventory and then the expected generated-topology digest; both were
corrected through the owning verifier and `make generate`.

| Command | Result | Evidence |
| --- | --- | --- |
| `make task-guide ROLE=module-author OWNER=module.networkflow` | PASS | Focused Network Flow guidance selected unit and service-backed slices |
| Graph v2 cancellation/determinism row | PASS | Run root `20260816T000344Z-p2177955` |
| Network Flow v3 contract and route-error rows | PASS | Run root `20260816T000433Z-p2181949` |
| Saved-graph lifecycle route row | PASS | Run root `20260816T000347Z-p2178420` |
| Declaration, lifecycle, and packaged-process service rows | PASS | Run root `20260816T000436Z-p2182411` |
| Network Flow v2 state and bounded concurrency rows | PASS | Run root `20260816T002052Z-p2233928` |
| Full Extensions state-admission matrix | PASS | Run root `20260816T002208Z-p2240679` |
| `make generate` | PASS | Final run root `20260816T002332Z-p2249653` |
| `make generate-drift` | PASS | Run root `20260816T002239Z-p2242099` |
| `make migration-drift` | PASS | Run root `20260816T002248Z-p2244981` |
| `make generated-artifact-policy-check` | PASS | Run root `20260816T002257Z-p2247846` |
| `make json-shape-check` | PASS | Run root `20260816T002342Z-p2252615` |
| `make backend-module-boundary-check` | PASS | Run root `20260816T002352Z-p2253584` |
| `make frontend-import-boundary-check` | PASS | Run root `20260816T002354Z-p2253987` |
| `make format` | PASS | Final run root `20260816T002205Z-p2237147` |
| `git diff --check` | PASS | No whitespace errors before the tracker checkpoint |
| `make lint-markdown` | PASS | Final checkpoint run root `20260816T002507Z-p2254906` |

Compatibility remains intentionally one-way: the active ephemeral route and
saved-graph routes emit v2 only, major-2 clients receive no saved-graph
workspace, and there is no v1 payload translator, Graph-owned job facade,
cursor key, or new idempotency mechanism. Migration `00032` remains additive;
the v1 facade and five v1 tables remain only for the coordinated GP2-S08
cutover. Reporting exact consumption, Recovery rebuild, destructive deletion,
the saved-graph browser workflow, load/fault gates, finalizer, and broad check
remain in their later owning slices. The principal unresolved risk entering
GP2-S06 is that immutable results have no durable consumer lease until
Reporting is cut over.

### GP2-S06 — Reporting's real Graph consumer

- Replace v1 ref decoding and `postgresbinding` with `source_projection_ref.v2`
  and the typed source-owner registry.
- Validate source owner, incident, declaration, exact selected result, source
  boundary, and digests inside release admission's transaction.
- Acquire a Graph result lease before the queued render can observe its payload.
- Materialize `explicit_refs`, `neighborhood`, and `all_with_bounds` from exact
  Graph result data; preserve timeline behavior unchanged.
- Apply Reporting redaction and bounds before diagram bytes. Remove placeholder
  Graph labels and any fallback that substitutes snapshot fields for Graph
  content.
- Release leases only through terminal success, terminal failure, cancellation,
  or safe orphan cleanup after the consumer outcome is durable.

Exit: real graph diagrams, every selection mode, same-transaction visibility,
stale/missing/duplicate/digest failures, redaction, retry, lease loss, and
terminal cleanup pass.

#### GP2-S06 completion evidence

Reporting now admits only the complete `source_projection_ref.v2` tuple. The
route decoder rejects unknown, removed, malformed, unsorted, duplicate-view,
non-Network-Flow, non-v2, and non-`gpres_` inputs. Release admission dispatches
the tuple through an exact source-owner registry, and the Network Flow adapter
validates incident ownership, the active declaration, selected result, source
snapshot, and all digests through the release transaction. It acquires the
`snapshot_reporting` render lease in that same transaction before the Common
Job payload or idempotency result can become observable.

The worker renews that lease before reading ordered immutable vertices and
edges. Result identity is rechecked at the registry boundary so a defective
provider cannot substitute a different tuple. Durable terminal success,
render failure, cancellation, and ordinary failed completion release the job's
leases only after the terminal transition is committed. Missing, stale,
unselected, digest-mismatched, and lease-lost states map to the closed Reporting
reason vocabulary without exposing source values.

Graph diagrams now select real result objects using closed, fully materialized
`explicit_refs`, `neighborhood`, and `all_with_bounds` rules. Selection follows
projection output order, validates duplicate/missing refs and edge endpoints,
enforces the 80-vertex/160-edge render boundary, and emits deterministic
overflow evidence for an explicit `summarize` policy. Mermaid and SVG are
derived from the selected topology rather than graph-view/run placeholders.
Network Flow has not adopted release-safe classifications for raw endpoint and
flow values, so Reporting uses only the adopted ordinal fallback for internal
labels and fails external graph release closed instead of copying raw values or
inventing a disclosure policy. This is the durable security boundary until a
future owner version adds typed releasable Graph label fields.

The service-backed adapter test publishes an exact result and authoritative
selected declaration, proves the lease is visible inside the borrowed release
transaction, renews and reads the exact ordered result, rejects a stale source
boundary, commits terminal lease deletion, and then proves further reads fail.
The runtime assembly injects exactly the Network Flow provider; Reporting has
no PostgreSQL binding import or legacy Graph reader.

Related failures found and corrected during the slice were:

- the first focused run, `20260816T003441Z-p2264198`, exposed one missed
  determinism-render parameter and v1-only binding tests;
- run `20260816T003852Z-p2273317` exposed a fixture result ID that did not use
  the required `gpres_` plus SHA-256 shape;
- service runs `20260816T004723Z-p2292731` and
  `20260816T004813Z-p2294771` exposed, respectively, an incorrectly declared
  fixture capability and the distinction between source direction `forward`
  and persisted Graph direction `directed`.

All four failures were caused by the S06 cutover or its new evidence and were
fixed without compatibility aliases or weakened validation. An attempted
`make backend-import-boundary-check` was a command-selection error because no
such public target exists; `make help-all` identified and the checkpoint ran
the canonical `make backend-module-boundary-check` target. No product gate was
skipped.

| Command | Result | Evidence |
| --- | --- | --- |
| `make task-guide ROLE=module-author OWNER=module.reporting` | PASS | Selected the focused unit and service-backed owner slices |
| Exact Graph selection/render row | PASS | Run root `20260816T004714Z-p2291820` |
| Exact Graph admission/lease service row | PASS | Run root `20260816T004902Z-p2299898` |
| `make test-slice OWNER=module.reporting` | PASS, 16/16 | Final run root `20260816T005106Z-p2310003` |
| `make service-backed-test-slice OWNER=module.reporting` | PASS, 8/8 | Final run root `20260816T005242Z-p2322503` |
| App-server publication/composition row | PASS | Run root `20260816T005021Z-p2305475` |
| `make generate` | PASS | Run root `20260816T005140Z-p2314295` |
| `make generate-drift` | PASS | Run root `20260816T005158Z-p2317353` |
| `make generated-artifact-policy-check` | PASS | Run root `20260816T005158Z-p2317372` |
| `make json-shape-check` | PASS | Run root `20260816T005158Z-p2317406` |
| `make backend-module-boundary-check` | PASS | Run root `20260816T005213Z-p2321872` |
| `make frontend-import-boundary-check` | PASS | Run root `20260816T005158Z-p2317738` |
| `make format` | PASS | Final run root `20260816T005103Z-p2306477` |
| Reporting v1 runtime symbol scan | PASS | No non-test v1 ref, binding reader, or `postgresbinding` match |
| `git diff --check` | PASS | No whitespace errors before the tracker checkpoint |
| `make lint-markdown` | PASS | Checkpoint run root `20260816T005350Z-p2324086` |

Compatibility is intentionally breaking and closed: no Reporting v1 reference
decoder, heuristic translation, latest-result lookup, placeholder graph path,
or dual source registry remains. Persisted v1-reference absence and physical
v1 table deletion remain destructive GP2-S08 preflight gates. Recovery still
uses its earlier runtime binding, so GP2-S07 is the only authorized next slice.
The primary risk entering S07 is that backups can preserve Network Flow
declarations but cannot yet rebuild and reconcile their selected v2 results and
Reporting leases before readiness.

### GP2-S07 — Recovery and backup integration

- Add Network Flow declarations to its authoritative backup codec and extension
  state families.
- Freeze the one-entry Graph source registry and matching implementation binding.
- Restore Network Flow authoritative state before Graph reconstruction; keep
  common workers quiescent until readiness.
- Clear exactly the four v2 derived tables and rebuild eligible declarations in
  canonical order with exact result identities and digests.
- Treat pending, failed, and retired declarations according to the adopted v2
  candidate matrix; do not infer a completed result.
- Reconcile or fail nonterminal Reporting job leases before workers resume.
- Preserve the exact old empty-registry backup bridge only as defined in
  section 4.6.

Exit: v2 clear-only, active rebuild, declaration exclusions, exact Reporting
reference preservation, rollback, indeterminate commit, cancellation, lease
loss, terminal replay, safe failures, and source/import/SQL boundaries pass.

#### GP2-S07 completion evidence

Network Flow saved-graph declarations are now the fifth authoritative Network
Flow backup family. Graph Projection publishes one current v2 Recovery
contribution for exactly the four v2 derived tables, freezes the one-entry
Network Flow source registry and matching implementation binding, and retains
the historical empty-v1-registry pair only behind exact schema and digest
dispatch. Current restore rejects the legacy artifact pair and historical
restore rejects supplied or forged artifacts.

The assembled participant enumerates active declarations with selected exact
bindings, reconstructs the immutable Network Flow source boundary through the
pure v2 engine, proves the complete result tuple, and atomically replaces only
the four v2 derived tables. The five v1 tables are deliberately preserved for
GP2-S08. Post-commit reads prove result, vertex, edge, and lease counts plus
every exact result binding. Publication failure rolls back the clear, and an
uncertain commit remains an indeterminate-target failure.

Recovery now composes owner-specific derived-state reconciliation through a
narrow Graph writer port. Common Jobs releases copied execution attempts while
preserving durable status, retry count, and retry schedule. Network Flow
strictly validates restored graph-materialization payloads. Reporting strictly
validates every restored nonterminal release reference against the rebuilt
binding set, recreates deterministic `snapshot_reporting` leases, and fails
readiness if any exact referenced result is unavailable. An active route-backed
scenario proves identical result, vertex, and edge IDs after replacement and
proves both Network Flow and Reporting jobs plus the Reporting lease are
reconciled before readiness.

| Gate | Result | Evidence |
| --- | --- | --- |
| Initial Recovery owner probe | Expected related failure, 22/38 | Run root `20260816T005633Z-p2327004`; exposed absent v2 contributions, stale codecs, and missing source binding |
| Recovery catalog contribution row | PASS | Run root `20260816T005936Z-p2373695` |
| Current/historical artifact dispatcher row | PASS | Run root `20260816T010923Z-p2384092` |
| Intermediate Recovery owner probe | Expected related failure, 34/38 | Run root `20260816T011007Z-p2386890`; remaining binding and operator assertions were corrected |
| Recovery targeted unit/integration rows | PASS | Run root `20260816T011411Z-p2441238` |
| Intermediate Graph owner probes | Expected related failures | Run roots `20260816T010931Z-p2384618`, `20260816T011727Z-p2444500`, and `20260816T011810Z-p2446776`; v1 test assumptions and the inbound-boundary assertion were migrated |
| New deterministic restore row | PASS | Run root `20260816T012420Z-p2465505` |
| Intermediate `make generate` attempts | Expected related failures | Run roots `20260816T012226Z-p2452885` and `20260816T012343Z-p2459258`; invalid authored row naming was corrected before generation |
| `make generate` | PASS | Run root `20260816T012407Z-p2462479` |
| Canonical Graph row after boundary migration | PASS | Run root `20260816T013409Z-p2471929`; supersedes related failure root `20260816T013316Z-p2470734` |
| Graph restore service/publication rows | PASS, 4/4 | Run root `20260816T013953Z-p2482848` |
| Active saved-graph identity/job/lease restore row | PASS, 3/3 | Run root `20260816T013859Z-p2479976`; supersedes related diagnostic roots `20260816T013705Z-p2476274` and `20260816T013752Z-p2478083` |
| Recovery assembly/evidence row | PASS, 3/3 | Run root `20260816T013926Z-p2481391` |
| `make test-slice OWNER=module.graphprojection` | PASS, 12/12 | Run root `20260816T014030Z-p2484625` |
| `make service-backed-test-slice OWNER=module.graphprojection` | PASS, 7/7 | Run root `20260816T014102Z-p2486919` |
| `make test-slice OWNER=module.recovery` | PASS, 38/38 | Run root `20260816T014112Z-p2488303` |
| `make service-backed-test-slice OWNER=module.recovery` | PASS, 26/26 | Run root `20260816T014227Z-p2532099` |
| `make test-slice OWNER=module.reporting` | PASS, 16/16 | Run root `20260816T014459Z-p2578697` |
| `make service-backed-test-slice OWNER=module.reporting` | PASS, 8/8 | Run root `20260816T014459Z-p2578727` |
| `make test-slice OWNER=module.networkflow` | PASS, 76/76 | Final run root `20260816T014834Z-p2678629`; supersedes the related stale-limit-fixture failure root `20260816T014459Z-p2578674` and targeted repair root `20260816T014824Z-p2678048` |
| `make service-backed-test-slice OWNER=module.networkflow` | PASS, 39/39 | Run root `20260816T014459Z-p2578691` |
| `make format` | PASS | Final run root `20260816T015206Z-p2723134` |
| `make generate-drift` | PASS, 4/4 | Run root `20260816T014434Z-p2573938` |
| `make generated-artifact-policy-check` | PASS, 3/3 | Run root `20260816T014434Z-p2574011` |
| `make json-shape-check` | PASS, 3/3 | Run root `20260816T014434Z-p2573950` |
| `make backend-module-boundary-check` | PASS, 3/3 | Run root `20260816T014434Z-p2574223` |
| `git diff --check` | PASS | No whitespace errors before the tracker checkpoint |
| `make lint-markdown` | PASS | Final checkpoint run root `20260816T015237Z-p2727818` |

Compatibility remains deliberately bounded. The exact historical empty-v1
registry and binding are not a runtime v1 path and remain only until the fresh
v2 backup and isolated-restore inventory gate in GP2-S10. The v1 runtime,
fixtures, and five v1 tables remain solely for the ordered GP2-S08 cutover.
No heuristic artifact translation, declaration inference, result-ID rewrite,
or copied lease row is admitted. Migration and v1-reference emptiness remain
the destructive GP2-S08 preflight; maximum-size and fresh-backup operational
proof remain GP2-S10 work rather than unresolved S07 correctness gaps.

### GP2-S08 — V1 cutover and dead-code deletion

- Switch Network Flow ephemeral output, saved Graph results, Reporting refs,
  Recovery, generated clients, tests, and fixtures to v2.
- Delete the retained create/refresh facade, broad repository, query and
  invalidation services, old lifecycle/query DTOs and errors, accepted/computing
  states, Graph cursor, Graph idempotency, run retention, publication hooks,
  `postgresstore`, `postgresbinding`, old store fixtures, and obsolete helper
  exports.
- Delete v1 runtime contracts and non-emitted aliases. Git history is the only
  historical source; no build or test target retains them.
- Add migration `00033` with zero-state preflight and removal of the legacy five
  tables.
- Strengthen boundary guards so deleted surfaces cannot return through aliases,
  forwarding packages, copied algorithms, dormant flags, or compatibility
  views.

Exit: consumer and symbol searches are empty, the old tables are absent at
head, non-empty preflight fails closed, no v1 runtime bytes are accepted or
emitted, and all backend owner slices pass.

#### GP2-S08 completion evidence

Migration `00033_graph_projection_v1_removal.sql` now owns the destructive
cutover. Its transactional preflight rejects any row in the five legacy Graph
tables, any persisted Reporting release v1 reference, and any Reporting job
payload containing v1 or malformed Graph references before issuing DDL. An
empty cutover removes all five tables. Its down section recreates only the
empty schema needed by disposable migration tests; it cannot recover data and
is not a supported production rollback.

The v1 facade, broad repository, binding/store packages, private cursor and
idempotency implementations, lifecycle/query services, aliases, hooks, legacy
tests, and 36-fixture family are deleted. The pure engine no longer constructs
synthetic v1 run state. Current Recovery classifies exactly the four v2 derived
tables and emits v2 terminal journal evidence. The historical empty-v1
registry/binding/result artifacts remain byte-exact and reachable only through
their exact artifact dispatcher until GP2-S10.

The repository execution has no authorized external deployment data plane, so
external row emptiness remains `not_applicable_to_repository_execution` rather
than inferred. The migration preflight is the executable deployment gate, and
the required pre-cutover binary remains commit
`df30974f44b18f0667498ef90555fd17846444b9`. After `00033`, rollback is an
exact pre-cutover database and object-store restore into a replacement target
using that matching old binary/schema pair, never an in-place binary rollback
or a reverse translation of v2 identities.

| Gate | Result | Evidence |
| --- | --- | --- |
| Initial migration row | Expected related failure | Run root `20260816T020046Z-p2736020`; exposed driver DETAIL handling and multi-command fixture seeding, both corrected without weakening the fail-closed reason |
| Migration preflight/head row | PASS, 3/3 | Run root `20260816T020204Z-p2742138`; proves empty cutover, mechanical empty rollback/reapply, legacy-row failure, Reporting release/job v1 failure, malformed-job failure, and no partial DDL |
| Initial `make generate` | Expected related failure | Run root `20260816T021421Z-p2763582`; version 33 lacked an authored source-owner assignment |
| Intermediate `make generate` | Expected related failure | Run root `20260816T021610Z-p2765302`; exposed and led to repair of an accidental truncated task-surface diagnostic while preserving only the intended target deletion |
| `make generate` after cutover projection | PASS | Run root `20260816T021709Z-p2768695` |
| Initial Graph owner checkpoint | Expected related failure, 7/8 | Run root `20260816T022103Z-p2780302`; the new deletion guard treated an empty workspace directory as a live Go package and was corrected to inspect source-bearing packages |
| Initial Recovery owner checkpoint | Expected related failure, 36/38 | Run root `20260816T022214Z-p2783714`; exposed the stale current catalog digest and v1 journal schema reference after live v1 table retirement |
| Final generated contract projection | PASS | Run root `20260816T022429Z-p2836899`; current catalog, v2 binding/result, terminal journal, migration lineage, task surface, and v2 Graph golden are aligned |
| `make test-slice OWNER=module.graphprojection` | PASS, 8/8 | Run root `20260816T022444Z-p2839900` |
| `make service-backed-test-slice OWNER=module.graphprojection` | PASS, 5/5 | Run root `20260816T022904Z-p2937649` |
| `make test-slice OWNER=module.recovery` | PASS, 38/38 | Run root `20260816T022444Z-p2839902` |
| `make service-backed-test-slice OWNER=module.recovery` | PASS, 26/26 | Run root `20260816T022904Z-p2937652` |
| `make test-slice OWNER=module.reporting` | PASS, 16/16 | Run root `20260816T022635Z-p2887907` |
| `make service-backed-test-slice OWNER=module.reporting` | PASS, 8/8 | Run root `20260816T022955Z-p2973999` |
| `make test-slice OWNER=module.networkflow` | PASS, 76/76 | Run root `20260816T022635Z-p2887901` |
| `make service-backed-test-slice OWNER=module.networkflow` | PASS, 39/39 | Run root `20260816T022955Z-p2973997` |
| `make format` | PASS | Final run root `20260816T023213Z-p3013105` |
| `make generate-drift` | PASS, 4/4 | Run root `20260816T023228Z-p3016985` |
| `make migration-drift` | PASS, 5/5 | Run root `20260816T023228Z-p3017044` |
| `make generated-artifact-policy-check` | PASS, 3/3 | Run root `20260816T023228Z-p3017030` |
| `make json-shape-check` | PASS, 3/3 | Run root `20260816T023228Z-p3017068` |
| `make backend-module-boundary-check` | PASS, 3/3 | Run root `20260816T023229Z-p3017444` |
| `make frontend-import-boundary-check` | PASS, 2/2 | Run root `20260816T023229Z-p3017470` |
| `git diff --check` | PASS | No whitespace errors before the tracker checkpoint |
| `make lint-markdown` | PASS | Checkpoint run root `20260816T023409Z-p3024622` |

The Graph boundary row now permanently rejects root imports of PostgreSQL,
HTTP, authentication, or Common Jobs; rejects removed v1 packages/imports; and
asserts one current four-table v2 Recovery contribution. Repository scans find
no production `postgresstore`, `postgresbinding`, fixture harness, v1 Graph
table model, private cursor/idempotency implementation, or old Graph lifecycle
symbol outside the exact historical Recovery bridge and migration history.
Duplicate v1 response parsing and payload translation are intentionally absent.

### GP2-S09 — Complete saved-graph UI

- Extend the current Network Analysis Graph panel, rather than creating a
  separate Graph application shell.
- Keep exploratory ephemeral graphs explicitly unsaved.
- Add saved-graph discovery and selection, save dialog, normalized display
  name, common-job progress, load/retry states, rename, refresh, retire,
  contributor drill-down, and permission-aware controls.
- Use generated clients and schemas only. Do not duplicate result identity,
  error mapping, role policy, cursor decoding, or canonical query logic in the
  browser.
- Reconcile stale requests by incident, graph view, result identity, and request
  generation. Preserve focus and announce job/error transitions accessibly.

Exit: frontend unit, type, accessibility, stale-response, focus-restoration,
authorization-loss, stateful, measurement, and browser workflows pass.

#### GP2-S09 completion evidence

The existing Network Analysis Graph workspace now presents an explicit choice
between unsaved exploration and incident-shared saved graphs. The saved surface
uses only generated Network Flow major-3 types and validators and implements
list, create, exact-result read, rename, refresh, retire, contributor query,
durable job status, last-safe-result behavior, empty/error/retry states, and
role-aware controls. Viewer, editor, reviewer, and administrator capabilities
match the adopted matrix. Request-generation guards discard stale list, result,
and contributor responses across incident or selection changes.

Dialogs use the shared Network Flow modal focus contract, including focus trap,
Escape dismissal, disabled dismissal during mutation, and trigger focus return.
Contributor close restores focus to the selected graph object. Terminal job
polling replaces queued/running notices with succeeded, failed, or cancelled
outcomes, retaining explicit last-safe-result wording. Both unsaved and saved
viewers mount at most 500 vertices and 1,000 edges and expose bounded previous
and next navigation. Saved objects sort by their stable display labels before
pagination so random source identities cannot make browser order unstable.

The supported-load browser fixture contains 501 deterministic vertices and
1,001 deterministic edges. Its measurement row proves exact first-page mounts
of 500 and 1,000 and exact second-page mounts of one each. The stateful browser
row proves create, job completion, exact-result viewing, contributor lookup,
rename, last-safe refresh, and retirement against the real server. The existing
Network Flow accessibility and visual rows now include the saved-result surface.

| Gate | Result | Evidence |
| --- | --- | --- |
| Initial test-catalog formatting attempts | Expected related artifact failures | Authored row and collaborator arrays were ASCII-sorted before any test execution |
| Initial full web owner run | Cancelled related diagnostic, 23/34 | Run root `20260816T024901Z-p3048042`; an unstable callback caused a disabled-controller render loop and was replaced with the stable workspace error callback |
| Intermediate saved lifecycle rows | Expected related failures | Run roots `20260816T025304Z-p3053005`, `20260816T025419Z-p3058065`, and `20260816T025649Z-p3067566`; ambiguous selectors and asynchronous role reload expectations were corrected without weakening behavior |
| Saved lifecycle, role, and render-bound rows | PASS | Final targeted roots `20260816T025453Z-p3062432`, `20260816T025716Z-p3072020`, and `20260816T025649Z-p3067566` for the render-bound row |
| Initial stateful route projection | Expected related artifact failure | The new Playwright row lacked its generated browser group; `make generate` projected it before execution |
| Saved-graph stateful browser row | PASS, 13/13 | Run root `20260816T025906Z-p3080031` |
| Initial saved-graph accessibility row | Expected related assertion failure | Run root `20260816T030232Z-p3114515`; Shift+Tab correctly skipped the disabled submit button and wrapped to Cancel, so the assertion was corrected to the actual focus contract |
| Saved-graph accessibility row | PASS, 12/12 | Run root `20260816T030409Z-p3143642` |
| Initial ordinary visual reconciliation | Expected related snapshot failure | Run root `20260816T030559Z-p3172783`; the adopted Graph sub-surface controls intentionally made the contributor golden stale, with zero ambiguous emitted mappings and zero missing emitted goldens |
| Intermediate visual update | Expected related assertion failure | Run root `20260816T031057Z-p3229858`; terminal status and terminal notice both matched a non-exact locator after polling was corrected |
| `make browser-e2e-visual-update` | PASS, 14/14 | Accepted refresh root `20260816T032017Z-p3323589`; reviewed changes are `network-flow-analysis-graph-contributors-linux.png` and new `network-flow-analysis-saved-graph-result-linux.png`; viewport, zoom, masks, and screenshot scope are unchanged, while one Graph-specific dynamic identity mask was added |
| Saved-graph DOM measurement row | PASS, 12/12 | Run root `20260816T031922Z-p3298234`; retained attachment records the 501/1,001 logical and 500/1,000 mounted limits |
| `make browser-e2e-visual` | PASS, 14/14 | Final current-source run root `20260816T033617Z-p3559167` |
| `make test-slice OWNER=web.networkflow` | PASS, 36/36 | Run root `20260816T032419Z-p3379351` |
| `make test-slice OWNER=module.networkflow` | PASS, 78/78 | Run root `20260816T032444Z-p3383603` |
| `make service-backed-test-slice OWNER=module.networkflow` | PASS, 41/41 | Run root `20260816T032717Z-p3429901` |
| Initial `make frontend-unit` | Expected related failure, 386/389 | Run root `20260816T032944Z-p3470388`; contract-major, source-ownership, and shared-selector projections were corrected |
| Targeted protocol and architecture policy rows | PASS | Run roots `20260816T033326Z-p3512762` and `20260816T033331Z-p3513273` |
| `make frontend-unit` | PASS, 389/389 | Run root `20260816T033339Z-p3513984` |
| `make frontend-typecheck` | PASS, 2/2 | Final run root `20260816T033316Z-p3512320` |
| `make frontend-import-boundary-check` | PASS, 2/2 | Run root `20260816T033532Z-p3555265` |
| `make format` | PASS, 2/2 | Final run root `20260816T033312Z-p3508809` |
| `make generate` | PASS | Final run root `20260816T031914Z-p3295306` |
| `make generate-drift` | PASS, 4/4 | Run root `20260816T033608Z-p3555809` |
| `make generated-artifact-policy-check` | PASS, 3/3 | Run root `20260816T033616Z-p3558678` |
| `git diff --check` | PASS | No whitespace errors before the tracker checkpoint |
| `make lint-markdown` | PASS | Checkpoint run root `20260816T034113Z-p3588329`; final follow-up root `20260816T034138Z-p3589391` |

Compatibility remains intentionally one-way. The web application consumes the
major-3 saved-graph contract directly and contains no major-2 response parser,
v1 Graph payload translator, duplicate role policy, or client-side result
identity implementation. Existing unsaved exploration remains because it is an
active and distinct product workflow. The historical empty-v1 Recovery bridge
is unchanged and remains the only temporary compatibility path pending GP2-S10.

### GP2-S10 — Production hardening and compatibility retirement

- Prove bounded current-profile producer input, vertices, edges, traversal,
  selection, concurrency, memory, transaction duration, and job queue pressure.
- Prove cancellation latency, retry exhaustion, shutdown, execution loss,
  process restart, lease cleanup, restore quiescence, and safe telemetry.
- Verify that logs, traces, metrics, job summaries, cursor failures, audits, and
  Recovery evidence expose no raw Graph input, source values, SQL, secrets,
  cursor internals, stack text, or unredacted report content.
- Produce and isolate-restore a fresh v2 backup, complete the supported artifact
  inventory, and record the prior binary/backup rollback decision.
- Remove the exact historical bridge, regenerate the recovery binding, and
  repeat Recovery evidence only after every section 4.6 gate passes.

Exit: operational budgets pass, rollback and restore evidence are explicit,
and no historical Graph binding remains in the current binary.

#### GP2-S10 completion checkpoint

Status: `DONE`.

The production envelope and compatibility retirement are complete. The pure
engine now has an executable maximum-bound case at exactly 100,000 vertices
and 250,000 edges, while its existing cancellation contract observes context
at least every 1,024 processed items and around publication. Network Flow has
an explicit transactional test proving that a stale generation rolls back the
staged immutable result, while a concurrent rename changes only
`graph_view_version` and does not block publication of the still-current
materialization generation.

Common Jobs supervision proves the one-handler-per-process composition,
bounded concurrency, retry exhaustion, shutdown, lost-attempt handling, and
lease-loss behavior. Extension finalization proves cancellation before commit,
atomic success/failure proof, and indeterminate-commit fail-closed handling.
Graph result storage proves atomic immutable publication, digest consistency,
ordered exact reads and traversal, active-lease protection, expired-lease
cleanup, and reachability cleanup. Reporting owner evidence proves exact lease
acquire/read/renew/release and safe redaction; Network Flow safe-value evidence
proves raw values do not enter logs, telemetry, or audit payloads.

Recovery captured a fresh current backup whose Network Flow graph-view family
and Graph restore artifacts are v2, restored it into the canonical isolated
operator target, and passed readiness. Only after that evidence passed, the
exact empty-v1-registry schemas, fixtures, dispatcher, digest admission,
generated constants, runtime compatibility fields, and source-only restore
seams were removed. Current restore now rejects that retired registry digest,
and repository scans find no packaged or runtime Graph restore v1 binding.

Compatibility is intentionally closed. There is no current-binary v1 Graph
reader, dispatcher, payload translator, or major-2 Network Flow consumer.
After migration `00033`, rollback is not an in-place binary rollback: operators
must restore the exact pre-cutover database and object backup into a replacement
target and use the matching old binary/schema pair, as recorded in
`docs/guides/graph_projection_v2_rollout.md`.

Files changed specifically for this checkpoint include:

```text
apps/web/e2e/{extensions.stateful.spec.ts,network-flow.spec.ts}
contracts/recovery/{index.json,fixtures/backup-integrity-manifest.v3.json}
contracts/recovery/{graph-projection-restore-*.v1.schema.json,fixtures/graph-projection-restore-*.v1.json} (deleted)
docs/{graph_projection_nlspec.md,guides/graph_projection_v2_rollout.md}
internal/gen/contractrecovery/{artifacts_gen.go,graph_projection_restore_binding_gen.go}
internal/modules/graphprojection/{engine_v2_test.go,restore_contract.go,restore_contract_test.go,restore_service.go,restore_service_test.go}
internal/modules/networkflow/graph_view_store_test.go
internal/modules/recovery/{restore.go,vnext_codec.go,vnext_codec_test.go,vnext_graph_restore_artifacts_test.go}
internal/modules/recovery/restorecontract/graphprojection.go
tools/contractgen/{main.go,recovery_validation.go}
tools/test_families/{module.graphprojection.json,module.networkflow.json,module.recovery.json}
```

Validation evidence:

| Gate | Result | Evidence |
| --- | --- | --- |
| Affected owner guides and inventories | PASS | `make task-guide ROLE=module-author` and `make explain-test-owner` for Graph Projection, Network Flow, Recovery, and Reporting |
| Maximum Graph v2 semantic bounds | PASS, 1/1 | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.v2_maximum_semantic_bounds`; run root `20260816T035601Z-p3614463` |
| Stale-generation atomic rollback and rename-safe publication | PASS, 3/3 | Final `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.store.graph_view_declaration_persistence`; run root `20260816T040255Z-p3633068` |
| Atomic result read/traversal/lease/cleanup | PASS, 3/3 | Focused Graph result row; run root `20260816T040350Z-p3634909` |
| Bounded job concurrency, retry, shutdown, and lease loss | PASS, 3/3 | `platform.jobs.integration.supervised_runner`; run root `20260816T040350Z-p3634886` |
| Atomic job finalization, cancellation, and crash posture | PASS, 3/3 | Extension finalization row; run root `20260816T040350Z-p3634899` |
| V2 capture/restore and retired-registry rejection | PASS, 1/1 | Recovery vNext row; run root `20260816T040350Z-p3634911` |
| Fresh backup and canonical isolated operator restore | PASS, 8/8 | Recovery operator process row; run root `20260816T040420Z-p3640812` |
| Graph Projection unit / service-backed | PASS, 9/9 and 5/5 | Run roots `20260816T040524Z-p3674645` and `20260816T040946Z-p3788247` |
| Network Flow unit / service-backed | PASS with corrected assertion, then 41/41 | Unit run `20260816T040524Z-p3674659` passed 76/78 and exposed the saved-result/name ownership assertion; corrected row passed 13/13 at `20260816T040844Z-p3763232`; service-backed root `20260816T040946Z-p3788246` |
| Recovery unit / service-backed | PASS, 38/38 and 26/26 | Run roots `20260816T040524Z-p3674666` and `20260816T040946Z-p3788264` |
| Reporting unit / service-backed | PASS, 16/16 and 8/8 | Run roots `20260816T041313Z-p3876242` and `20260816T041313Z-p3876264` |
| Backend module boundary | PASS, 3/3 | Run root `20260816T041313Z-p3876388` |
| Stateful browser | PASS, 36/36 | Initial root `20260816T041350Z-p3883289` exposed one stale major-2 browser expectation; current-only correction passed at `20260816T042049Z-p3957269` |
| Browser measurement | PASS, 29/29 | Run root `20260816T041350Z-p3883294` |
| Generation, migration, generated policy, and JSON shape drift | PASS | Run roots `20260816T041221Z-p3867716`, `20260816T041221Z-p3867754`, `20260816T041221Z-p3867747`, and `20260816T041221Z-p3867767` |
| Formatting | PASS, 2/2 | Final run root `20260816T042043Z-p3953696` |
| Historical bridge and major-2 consumer scan | PASS | No runtime, packaged Recovery v1 Graph artifact, legacy dispatcher symbol, or `supported_contract_majors: [2]` consumer remains; migration rollback tests are the only intentional legacy-table references |
| `git diff --check` | PASS | No whitespace errors at the S10 checkpoint |
| `make lint-markdown` | PASS | Checkpoint root `20260816T042457Z-p3987659`; follow-up root `20260816T042524Z-p3988715` |

Two related diagnostic failures are retained rather than hidden. The initial
Network Flow publication row failed because its new fixture referenced a
nonexistent Common Job; seeding the required authoritative job satisfied the
`00032` foreign key and the unchanged CAS behavior passed. The first full
Network Flow unit/browser sweep found declaration display-name assertions
inside the immutable result container, which was corrected to assert the
declaration heading. The first full stateful target then found an unmodified
major-2 client-support expectation, which was updated to the adopted major 3.
All superseding focused and broad evidence is green.

Residual operational risk is capacity calibration rather than correctness:
the enforced semantic maxima and concurrency quotas are proven, but deployment
operators must continue using privacy-safe telemetry to tune worker resources
for their hardware and incident mix. No compatibility bridge or deferred
correctness work remains.

### GP2-S11 — Final validation and handoff

Status: `DONE`.

The final slice completed the owner-routed verification matrix, repaired the
cross-owner gaps found by the first broad run, reran the affected owner slices,
and passed independent broad and release gates. No requirement was weakened,
no safe error was broadened, and no compatibility alias or v1 execution path
was introduced to make a gate pass.

#### Final gap dispositions

| Gap | Final disposition | Completion evidence |
| --- | --- | --- |
| GP2-G01 | `CLOSED` | Network Flow v3 owns saved-graph declarations, APIs, authorization, jobs, composition, backup state, and the browser workflow |
| GP2-G02 | `CLOSED` | Adopted Graph v2, Network Flow v3/state v2, Reporting, Extensions, Recovery, and Core changes agree with generated contracts and pass generation drift |
| GP2-G03 | `CLOSED` | The broad v1 store/facade is deleted; narrow publication, exact-read, traversal, lease, cleanup, and restore adapters pass unit and service-backed evidence |
| GP2-G04 | `CLOSED` | Common Jobs owns materialization attempts while generation-aware atomic publication preserves one deterministic semantic result |
| GP2-G05 | `CLOSED` | Deterministic `projection_result_id` and the complete digest tuple survive retry, restart, Reporting binding, backup, and isolated Recovery |
| GP2-G06 | `CLOSED` | Graph-owned cursors, idempotency, job state, retirement coordination, and deployment configuration are absent from the current runtime |
| GP2-G07 | `CLOSED` | Strict v2 admission rejects removed and unknown caller fields; equivalent semantic inputs have stable identities |
| GP2-G08 | `CLOSED` | V1 aliases, hooks, exports, fixtures, facades, binding/store packages, and runtime parser paths are deleted |
| GP2-G09 | `CLOSED` | Network Flow declarations are authoritative backup state, the four v2 derived tables rebuild exactly, and the historical empty-v1 bridge is removed |
| GP2-G10 | `CLOSED` | Reporting resolves exact ordered graph data, leases it, applies classification/redaction, and emits real bounded Mermaid/SVG artifacts |
| GP2-G11 | `CLOSED` | The Network Flow workspace covers saved-graph list/create/rename/refresh/retire/result/contributor/job states with bounded accessible visualization |
| GP2-G12 | `CLOSED` | Quotas, single-handler concurrency, 1,024-item cancellation cadence, cleanup/lease rules, safe telemetry, and maximum/fault evidence pass |

#### S11 implementation and integration corrections

- Added complete Extensions owner routing for Network Flow graph state, worker,
  job, and Reporting participant declarations, plus exact publication-catalog
  expectations.
- Closed the frontend selector-policy gap for the saved-graph heading and
  updated browser consumers to use the typed selector surface.
- Removed an aggregate browser reset that invalidated immutable seeded
  measurement fixtures and added a topology invariant preventing recurrence.
- Made destructive schema projection remove indexes, constraints, and triggers
  transitively with their dropped relation in both the catalog generator and
  JSON-shape parity validator. Historical migrations may legitimately own no
  final object after a later destructive migration.
- Added policy-compliant full indexes for both saved-graph foreign keys and
  aligned v2 immutable-result table grants with the repository privilege
  classes while database triggers continue rejecting semantic updates.
- Updated the canonical migration hash and exact operator evidence digest only
  after the final authored `00032` bytes stabilized.
- Made Network Flow Incident Bundle portability require and inspect all five
  authoritative state families, including saved-graph declarations.
- Removed unused Graph helpers, normalized safe error construction, and updated
  the authored harness target count after the retired fixture target was
  deleted.

#### S11 focused validation

| Layer | Result | Run root or evidence |
| --- | --- | --- |
| Graph Projection unit / service-backed | PASS, 9/9 and 5/5 | `20260816T055426Z-p809845`; service evidence `20260816T043608Z-p4161616` |
| Network Flow unit / service-backed | PASS, 78/78 and 41/41 | `20260816T055426Z-p809864`; `20260816T055751Z-p904592` |
| Reporting unit / service-backed | PASS, 16/16 and 8/8 | `20260816T055426Z-p809875`; service evidence `20260816T043608Z-p4161644` |
| Recovery unit / service-backed | PASS, 38/38 and 26/26 | `20260816T055426Z-p809886`; service evidence `20260816T043608Z-p4161679` |
| Extensions unit / service-backed | PASS, 40/40 and 26/26 | `20260816T043514Z-p4134555`; `20260816T043608Z-p4161659` |
| Database migrations unit / service-backed | PASS, 17/17 and 8/8 | `20260816T055314Z-p754489`; `20260816T060016Z-p945154` |
| Incident Bundles unit / service-backed | PASS, 26/26 and 15/15 | `20260816T055314Z-p754468`; `20260816T055214Z-p748781` |
| Operator and server unit composition | PASS, 17/17 and 45/45 | `20260816T055314Z-p754474`; `20260816T055314Z-p754479` |
| PostgreSQL privilege matrix | PASS, 3/3 | `20260816T060058Z-p947089` |
| Common Jobs operational telemetry | PASS, 3/3 | `20260816T054632Z-p727837` |
| Harness command/evidence/catalog rows | PASS, 1/1 each | `20260816T054600Z-p725542`, `20260816T054600Z-p725557`, and `20260816T054600Z-p725564` |
| Network Flow frontend owner | PASS, 36/36 | `20260816T042623Z-p3992471` |

#### S11 browser, drift, security, and broad validation

| Command | Result | Run root |
| --- | --- | --- |
| `make browser-e2e` | PASS, 67/67 | `20260816T051539Z-p455871` |
| `make browser-e2e-stateful` | PASS, 36/36 | `20260816T044520Z-p207234` |
| `make browser-e2e-a11y` | PASS, 14/14 | `20260816T044520Z-p207489` |
| `make browser-e2e-measurement` | PASS, 29/29 | `20260816T044520Z-p207260` |
| `make browser-e2e-visual` | PASS, 14/14 | `20260816T044520Z-p207342` |
| `make format` | PASS, 2/2 | `20260816T060135Z-p948589` |
| `make generate` | PASS | `20260816T060404Z-p962918` |
| `make generate-drift` | PASS, 4/4 | `20260816T060427Z-p966438` |
| `make migration-drift` | PASS, 5/5 | `20260816T060427Z-p966477` |
| `make generated-artifact-policy-check` | PASS, 3/3 | `20260816T060427Z-p966471` |
| `make json-shape-check` | PASS, 3/3 | `20260816T060418Z-p965864` |
| `make lint-go` | PASS | Direct target returned exit 0; this target emitted no retained run root |
| `make backend-module-boundary-check` | PASS, 3/3 | `20260816T060520Z-p1006472` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `20260816T060443Z-p977897` |
| `make go-gosec-targeted` | PASS, 4/4 | `20260816T060443Z-p978071` |
| `make go-vulncheck` | PASS, 4/4 | `20260816T060443Z-p978061` |
| `make frontend-typecheck` | PASS, 2/2 | `20260816T060533Z-p1007129` |
| `make frontend-unit` | PASS, 389/389 | `20260816T060533Z-p1007138` |
| `make lint-biome` | PASS, 2/2 | `20260816T060533Z-p1007170` |
| `make agent-finalize` | PASS, 1/1 | `20260816T060735Z-p1049421` |
| `make check` | PASS, 776/776 | `20260816T060753Z-p1052302` |
| `make release-check` | PASS, 965/965 | `20260816T061242Z-p1197640` |
| `git diff --check` | PASS | No whitespace errors at the final checkpoint |
| `make lint-markdown` | PASS | `20260816T064026Z-p1454723`; final follow-up rerun after recording this evidence also passed |

`RESULTS_DIR` was unset for `make agent-finalize`, so retained successful-run
maintenance was intentionally skipped; the finalizer itself passed and the
subsequent full `check` and `release-check` produced fresh retained evidence.

#### Related failures and their resolution

| Failure evidence | Classification and resolution |
| --- | --- |
| Initial Extensions unit root `20260816T042623Z-p3992399` and focused roots through `20260816T043332Z-p4127766` | Related authoring gaps in verification routing and exact declaration expectations; fixed in authored test-family inputs and Extensions tests, superseded by 40/40 and 26/26 passes |
| Frontend unit root `20260816T044009Z-p116751` | Related missing typed selector; fixed in UI contracts and component/browser consumers, superseded by 389/389 and browser passes |
| Aggregate browser roots `20260816T044520Z-p207252` and `20260816T045300Z-p387183` | Related topology reset invalidated fixture-backed Timeline measurement state; reset removed and guarded by a regression invariant, superseded by 67/67 |
| First broad root `20260816T052339Z-p515879` | Related 756/776 result exposed stale migration/evidence digests, incomplete publication expectations, dropped-object projection, v2 type privileges, five-family portability, one harness count, lint, and telemetry timing; every affected focused row and owner slice passed before the final broad rerun |
| DDL/ACL diagnostic roots `20260816T054632Z-p727834` and `20260816T054957Z-p740830` | Related least-privilege parity and uncovered saved-graph foreign-key indexes; authored migration fixed, then privilege and DDL parity passed |
| JSON-shape roots `20260816T060145Z-p952266` and `20260816T060331Z-p962168` | Related destructive-migration projection assumptions; dependent-object removal and legitimate empty historical allocations fixed, superseded by 3/3 |
| Parallel service-backed roots `20260816T054632Z-p727830` and `20260816T054632Z-p727844` | Harness-only shared service-image race caused by concurrent Make invocations; rerun sequentially with passing DDL and Incident Bundle roots |

One command-selection diagnostic used nonexistent target
`backend-import-boundary-check`; `make help-all` identified the public target
`backend-module-boundary-check`, which passed. One early focused invocation used
two invalid database row IDs and failed before executing product tests; the
owner catalog supplied the correct rows, which passed. Neither diagnostic is a
product failure.

#### Compatibility and residual risk

The final runtime is Graph Projection v2, Network Flow contract major 3/state
version 2, Reporting exact-reference v2, and Recovery v2 only. The historical
empty-v1 bridge, dual parser, five v1 tables, v1 runtime facade, and legacy
generated clients are absent. Migration `00032` remains the additive v2
introduction and `00033` is the destructive v1 removal. After `00033`, rollback
is replacement-target restoration of the exact pre-cutover backup using the
matching old binary/schema pair; in-place old-binary rollback is unsupported.

There is no residual correctness or temporary compatibility work. The only
remaining operational risk is environment-specific capacity calibration at the
adopted maximum graph sizes. The runtime bounds, quotas, cancellation cadence,
privacy-safe telemetry, and worker limit make that risk observable and bounded;
operators should tune resources from deployment evidence without changing the
semantic contract.

Exit: GP2-S00 through GP2-S11 are `DONE`, all binary criteria pass, no v1
runtime or historical Graph binding remains, and this tracker is the complete
handoff rather than an alternate owner specification.

## 6. Validation Matrix

### Required public targets

| Validation layer | Required command or target family | Required slices |
| --- | --- | --- |
| Documentation | `make lint-markdown` | Every slice checkpoint |
| Formatting | `make format` for touched Go/frontend sources | GP2-S03 through GP2-S10 as applicable |
| Generation | `make generate` | GP2-S02 and every later authored projection change |
| Generated drift | `make generate-drift` | GP2-S02, GP2-S07, GP2-S08, GP2-S10, GP2-S11 |
| Generated policy | `make generated-artifact-policy-check` | Same as generation drift |
| Contract shape | `make json-shape-check` | Contract and final slices |
| Migration history | `make migration-drift` | GP2-S03, GP2-S08, GP2-S11 |
| Backend boundaries | `make backend-module-boundary-check` | Every implementation slice |
| Graph unit | `make test-slice OWNER=module.graphprojection` | GP2-S03 through GP2-S11 |
| Graph services | `make service-backed-test-slice OWNER=module.graphprojection` | Storage, Recovery, cutover, and final slices |
| Network Flow unit | `make test-slice OWNER=module.networkflow` | GP2-S04 through GP2-S11 |
| Network Flow services | `make service-backed-test-slice OWNER=module.networkflow` | Producer, Recovery, cutover, hardening, and final slices |
| Reporting unit | `make test-slice OWNER=module.reporting` | GP2-S06 through GP2-S11 |
| Reporting services | `make service-backed-test-slice OWNER=module.reporting` | GP2-S06 through GP2-S11 |
| Recovery unit | `make test-slice OWNER=module.recovery` | GP2-S07 through GP2-S11 |
| Recovery services | `make service-backed-test-slice OWNER=module.recovery` | GP2-S07 through GP2-S11 |
| Network Flow frontend | `make test-slice OWNER=web.networkflow` | GP2-S02, GP2-S09, GP2-S11 |
| Frontend type/unit | `make frontend-typecheck` and `make frontend-unit` | GP2-S09, GP2-S11 |
| Browser behavior | `make browser-e2e` | GP2-S09, GP2-S11 |
| Stateful browser | `make browser-e2e-stateful` | GP2-S09, GP2-S10, GP2-S11 |
| Accessibility | `make browser-e2e-a11y` | GP2-S09, GP2-S11 |
| Measurement | `make browser-e2e-measurement` | GP2-S09, GP2-S10, GP2-S11 |
| Finalizer | `make agent-finalize` | GP2-S11 before broad gate |
| Broad gate | `make check` | GP2-S11 |

Each slice MUST first run `make task-guide ROLE=module-author OWNER=<owner-id>`
and `make explain-test-owner OWNER=<owner-id>` for every affected owner. New
tests MUST be added through authored test-family inputs and generated through
Make; generated topology MUST NOT be hand-edited.

### Required acceptance families

- V2 strict input members, omission/null rules, removed-member rejection,
  canonical ordering, and digest transcripts.
- Stable result identity across retry, job replay, process restart, and Recovery.
- Publication atomicity, same-ID byte mismatch, reference integrity, and lease
  reachability.
- Saved-graph create, rename, refresh, retire, list, get, result, contributor,
  cursor, authorization, conflict, CSRF, and idempotency behavior.
- Common-job cancellation, execution loss, retry exhaustion, prior-result
  preservation, and lost terminal response.
- Reporting source-owner dispatch, binding validation, three selection modes,
  redaction, limits, lease lifecycle, and real graph artifact output.
- Backup source declaration, exact rebuild, exclusions, readiness, rollback,
  indeterminate outcome, terminal replay, and worker quiescence.
- Migration `00032` clean application and migration `00033` empty-state success plus
  non-empty fail-closed evidence.
- Deleted v1 symbols, packages, tables, routes, flags, config, schemas, and
  runtime byte rejection.
- UI unsaved/saved distinction, job progress, retry, stale response, focus,
  keyboard, permission, stateful restart, and bounded DOM behavior.
- Safe errors, logs, telemetry, audits, cursors, job evidence, and Recovery
  evidence under sensitive source and dependency failures.

## 7. Tracker Update Protocol

After every future slice, before the next slice begins:

1. Run the slice's narrow required verification.
2. Update the slice status, files changed, substantive changes, commands,
   results, run roots, failures, compatibility impact, risks, and next action.
3. Update the gap and compatibility tables when a disposition changes.
4. Record any skipped conditional check with its exact reason.
5. Run `git diff --check`.
6. Run `make lint-markdown` and record its run root.
7. Begin the next slice only after the tracker checkpoint is complete.

Allowed statuses are `PLANNED`, `IN PROGRESS`, `BLOCKED`, and `DONE`.
`BLOCKED` requires an exact missing owner decision, external artifact, or failed
gate. `DONE` requires all exit criteria and terminal evidence; implementation
presence alone is insufficient.

## 8. Binary Completion Criteria

The GP2 iteration is complete only when all conditions are true:

- Graph Projection 2.0 and every affected owner are adopted and contradiction
  free.
- The runtime accepts and emits only `graph_projection.v2`.
- Network Flow has one production-composed incident-shared saved-graph producer
  using common durable jobs.
- Common jobs, not Graph, own attempt state, retry, cancellation, and crash
  recovery.
- Durable consumers bind deterministic `projection_result_id`, never random
  attempt identity.
- Reporting renders exact selected and redacted Graph vertices and edges.
- Recovery reproduces all stable result identities before workers resume.
- The four v2 derived tables and authoritative Network Flow declaration table
  have exact owner boundaries and passing migration evidence.
- The five legacy Graph tables, `postgresstore`, `postgresbinding`, old facade,
  cursor, idempotency, invalidation, run-history, hooks, aliases, and fixtures
  are absent.
- No public Graph route, authorization implementation, hidden worker, private
  cursor key, dual writer, compatibility view, or v1 translation remains.
- The complete saved-graph UI passes unit, browser, stateful, accessibility, and
  measurement evidence.
- Production concurrency, memory, cancellation, transaction, recovery, lease,
  telemetry, and shutdown budgets pass.
- A fresh v2 backup passes isolated restore, the artifact inventory is closed,
  and the historical empty-registry bridge is removed.
- `make agent-finalize` and the final `make check` pass.
- The final tracker handoff contains exact evidence and no unclassified risk.

## 9. Current Slice Handoff

### GP2-S11 and iteration result

GP2-S11 and the complete GP2 iteration are `DONE`. Graph Projection is a pure,
deterministic v2 engine; Network Flow owns saved-graph declarations and their
complete lifecycle; Reporting consumes and leases exact immutable results;
Recovery rebuilds those results from authoritative Network Flow state; and the
existing Network Flow workspace exposes the full bounded, accessible workflow.

The final owner, service, browser, drift, security, broad, and release gates
pass. GP2-G01 through GP2-G12 are closed. There is no active slice, deferred
compatibility bridge, or unclassified correctness risk.

### Files changed in this step

```text
apps/web/e2e/{extensions.stateful.spec.ts,network-flow.spec.ts}
apps/web/src/networkFlow/NetworkFlowSavedGraphPanel.tsx
packages/ui-contracts/src/networkFlowSelectors.ts
db/migrations/00032_graph_projection_v2.sql
internal/app/extensionassembly/publication_catalog_test.go
internal/app/operator/operator_migration_evidence_test.go
internal/modules/database_migrations/{catalog_characterization_test.go,rebaseline_manifest_integration_test.go}
internal/modules/extensions/{contract_test.go,coordinator_test.go,protocol_test.go}
internal/modules/graphprojection/{canonical.go,engine_v2.go,input.go}
internal/modules/graphprojection/postgresresult/store.go
internal/modules/networkflow/{extension_state.go,graph_restore_source.go,portability_state.go,reporting_graph_source.go}
internal/modules/reporting/graph_source.go
internal/testutil/pgtest/pgtest_test.go
tools/contractgen/recovery_validation.go
tools/database-migrations/generate-catalog-projections.mjs
tools/execution_topology_manifest.json
tools/execution_topology_render_index.json
tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership.mjs
tools/harness/generated-artifacts/tests/test-execution-topology.mjs
tools/harness/tests/contract-suite-support.mjs
tools/schema_object_ownership_manifest.json
tools/test_families/module.extensions.json
generated topology, task-surface, and contract projections updated by make generate
docs/handoffs/graphprojection-module-refactor-tracker.md
```

Earlier slice sections in this tracker record the complete owner, contract,
implementation, migration, generated, Recovery, Reporting, and web file groups
for GP2-S00 through GP2-S10. This list is the additional S11 validation and
integration-repair delta, not a replacement for those checkpoints.

### Explicitly unchanged

- Lockfiles, external deployment data, and deployment configuration
- The adopted semantic maxima of 100,000 vertices and 250,000 edges
- Network Flow quotas of 32 active and 128 retained declarations, four
  nonterminal graph jobs per incident, and one materialization handler per
  serving process
- The replacement-target rollback procedure adopted before destructive cutover

### Final validation

The GP2-S11 completion checkpoint records the exact narrow and broad commands,
run roots, related failures, superseding passes, skipped retained-run
maintenance, and final compatibility classification. Terminal broad evidence
is `make check` 776/776 at `20260816T060753Z-p1052302` and
`make release-check` 965/965 at `20260816T061242Z-p1197640`.

### Current next action

None. Preserve this tracker as the final handoff record. Any future Graph
Projection or saved-graph phase begins with a new owner-first plan and must not
reopen v1 compatibility implicitly.
