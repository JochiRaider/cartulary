# GP4 — Legacy Retirement and Production Simplification

## 1. Controlling Posture

- **Target subsystem:** `internal/modules/graphprojection`
- **Controlling artifact:** `docs/handoffs/graphprojection-module-refactor-tracker.md`
- **Current posture:** GP4 — Legacy Retirement and Production Simplification
- **Status:** GP2 and GP3 are frozen and complete; `GP4-S00` through
  `GP4-S08` are `DONE`, `GP4-S09` is `READY`, and `GP4-S10` is blocked
  on their declared predecessors
- **GP4 planning baseline:** clean commit
  `4d519e21f05d9df69f4c08c1d9a9ef4054e0fb8c`
- **GP3 planning baseline:** clean commit
  `d5b5d4fd3d4e8d046fb375e1b9225e4d496e519d`
- **Completed iteration:** GP2-S00 through GP2-S11 completed at `d5b5d4fd`
- **Earlier completed iteration:** GP-S00 through GP-S08 completed at
  `df30974f`
- **Planned GP4 order:**
  `GP4-S00 → GP4-S01 → GP4-S02 → GP4-S03 → GP4-S04 → GP4-S05 → GP4-S06 → GP4-S07 → GP4-S08 → GP4-S09 → GP4-S10`

This tracker controls the user-authorized implementation. Planned behavior does
not become conformance authority until its applicable owner is adopted. Every
slice is a separate workstream, and the tracker checkpoint for one slice MUST
complete before the next slice begins.

The authority order is:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 within their named scopes.
3. `docs/domain.md` for vocabulary and owner navigation within its stated
   boundary.
4. Versioned machine projections of adopted owners.
5. Current repository implementation and tests as evidence.
6. This tracker as sequencing, decision, validation, and handoff control only;
   it is not product or conformance authority.

`docs/research/nlspec-spec.md` supplied writing guidance for behavioral
completeness, unambiguous interfaces, explicit defaults, conceptual fidelity,
spec economy, recreatability, and binary acceptance criteria. It did not supply
Graph Projection requirements and MUST NOT be treated as an instruction source
or runtime authority.

### Statement classes

| Class | Meaning | Effect in this tracker |
| --- | --- | --- |
| Current fact | Verified owner or repository state at the stated iteration baseline | Baseline evidence; changes require an adopted owner and passing implementation evidence |
| Planned decision | User-authorized target for the current iteration | Directs owner drafting and slice design but is not conformance authority before adoption |
| Adopted requirement | Text later accepted in a named owner | Governs implementation and machine projections after its adoption slice |
| Completion evidence | Terminal command result, run root, artifact, or operational proof | Permits a slice to move from `IN PROGRESS` to `DONE` |

Sections 2 through 9 are the frozen GP2 completion record. Section 10 is the
frozen GP3 completion record. They retain their commands, run roots,
compatibility decisions, risks, and handoff evidence and must not be
reinterpreted as active GP4 requirements. Section 11 is the active GP4 plan.
Where the frozen record describes inventory-gated v1 or historical Recovery
compatibility, the GP4 planned decision supersedes that posture prospectively;
it does not rewrite what GP2 or GP3 implemented or proved.

## 2. Frozen GP2 — Prior Iteration Baseline

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

## 3. Frozen GP2 — Facts and Production Gaps

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

## 4. Frozen GP2 — Product and Architecture Contract

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

## 5. Frozen GP2 — Workstream and Slice Plan

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

## 6. Frozen GP2 — Validation Matrix

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

## 7. Frozen GP2 — Tracker Update Protocol

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

## 8. Frozen GP2 — Binary Completion Criteria

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

## 9. Frozen GP2 — Final Slice Handoff

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

### GP2 closeout next action

None. Preserve this tracker as the final handoff record. Any future Graph
Projection or saved-graph phase begins with a new owner-first plan and must not
reopen v1 compatibility implicitly.

## 10. GP3 — Production Readiness and Time-Bucketed Graphs

Everything in Sections 10.2 through 10.7 is a planned GP3 decision until the
applicable owner adopts it in GP3-S01. Current owners and implementation remain
the conformance baseline until then. GP3 does not reopen Graph Projection v1 or
change the completed GP2 evidence above.

### 10.1 GP3-S00 baseline and inventory

GP3 planning began from clean commit
`d5b5d4fd3d4e8d046fb375e1b9225e4d496e519d`. `git status --short` emitted no
paths before this tracker edit. The planning inspection covered:

```text
docs/handoffs/graphprojection-module-refactor-tracker.md
docs/domain.md
docs/research/nlspec-spec.md
docs/graph_projection_nlspec.md
docs/network-flow-activity-nlspec.md
docs/opentelemetry-instrumentation-nlspec.md
docs/extension-subsystem-nlspec.md
docs/guides/graph_projection_v2_rollout.md
internal/modules/graphprojection/postgresresult/store.go
internal/modules/networkflow/{api.go,graph.go,graph_view_jobs.go,module.go,reporting_graph_source.go}
internal/platform/jobs/{manager.go,runner.go,telemetry.go}
internal/app/server/runtime_assembly.go
db/migrations/00032_graph_projection_v2.sql
```

`docs/domain.md` was used only for current vocabulary and owner navigation.
`docs/research/nlspec-spec.md` was used only for specification-writing guidance:
behavioral completeness, explicit defaults, conceptual fidelity, spec economy,
recreatability, and testable completion. Neither document supplied unadopted
GP3 product behavior.

| Verified current fact | Production or expansion consequence |
| --- | --- |
| `postgresresult.Cleaner.DeleteUnreachableResults` has unit/service evidence but production has no constructor or scheduled caller | Superseded results and expired leases can accumulate indefinitely |
| The cleanup method accepts one caller-supplied global reachable-ID list and does not scope deletion by source owner | A future producer could have its results deleted by another owner's incomplete reachability set |
| Graph composition calls `ListRowsForTables`, filters and sorts a complete row slice, and retains contributing rows on vertices and edges | Memory grows with selected source rows rather than bounded result size; contributor lookup repeats the same construction |
| Common Jobs provides eight global attempt slots and has no per-worker active-attempt limit | Multiple maximum graph jobs can run concurrently and starve unrelated job kinds despite one registered graph handler |
| Generic Jobs telemetry exposes running count, duration, attempts, expiry, and lease-renewal failures | Queue wait, graph phase cost, result volume, cleanup progress, and cleanup backlog are not observable |
| Network Flow owner Table 20-A says omitted limits use default-and-maximum values, while runtime defaults are generally lower and the claimed configuration namespace has no resource-limit member | Owner, discovery, runtime, and deployment capacity policy disagree |
| The maximum Graph Projection fixture proves semantic completion at 100,000 vertices and 250,000 edges but publishes no retained deployment hardware envelope | Semantic correctness is proven; production capacity remains environment-specific |
| Network Flow's only explicit Graph product backlog item is time-bucketed output | Temporal growth has an owner-recognized boundary but no current schema, identity, limit, Recovery, Reporting, or UI contract |

### 10.2 Target ownership and interfaces

| Surface | GP3 planned decision |
| --- | --- |
| Graph Projection | Retain `graph_projection.v2`; add no temporal product semantics, transport, configuration, public route, or worker. Replace cleanup's global reachability input with source-owner-scoped transaction-safe capabilities. |
| Network Flow | Advance to public contract major `4`, state version `3`, and configuration contract major `2`. Own temporal graph semantics, effective resource policy, source iteration, result cleanup orchestration, and product UI. |
| Semantic queries | Introduce `cartulary.network_flow.graph_semantic_query.v2` with default and `time_bucket_v1` variants. New ephemeral queries and saved-graph creates accept v2 only. |
| Existing declarations | Continue to read, refresh, back up, restore, and return persisted semantic-query v1 declarations and their exact results. They remain a justified state-compatibility variant, not a public new-write protocol. |
| State migration | `network_flow_activity.state_2_to_3_v1` expands the admitted declaration union while preserving all five authoritative families byte-for-byte. It creates no job, rewrites no query, and changes no result identity. |
| Public routes | Retain the existing `/api/v1/incidents/{incident_id}/network-flow` route root and saved-graph paths. Generate only a major-4 client for current browser use. Add no operator route. |
| Reporting | Keep exact `source_projection_ref.v2`; route typed Network Flow graph label components through Reporting redaction. Internal output uses deterministic post-redaction labels, while external Graph release remains fail-closed without a separately adopted allow rule. |
| Recovery | Publish `graphprojection.restore_rebuild.v3`, which rebuilds admitted v1 default and v2 default/temporal Network Flow declarations into Graph Projection v2 results. Retain the exact v2 dispatcher read-only while a supported retained pre-GP3 backup references it. |
| Persistence | Add only `00034_graph_projection_cleanup_indexes.sql`. Migrations `00032` and `00033` remain immutable. |
| Worker control | Add a generated worker-runtime contract with `max_active_attempts_per_process`. The graph worker declares `1`; every existing worker explicitly declares the limit preserving its current effective behavior. |

#### 10.2.1 Time-bucket contract

`time_bucket_v1` has the following exact semantics:

- Both `time_range.start_utc` and `time_range.end_utc` are required and define
  the half-open interval `[start,end)`; `start` must be earlier than `end`.
- `bucket_width_seconds` is exactly one of `60`, `300`, `900`, `3600`,
  `21600`, or `86400`.
- Bucket boundaries are UTC instants aligned to the Unix epoch. Implementations
  use mathematical floor division for negative epochs and never use local time,
  locale, DST rules, or calendar-month arithmetic.
- With width `W`, `first_bucket_start=floor(start/W)*W`,
  `bucket_end_exclusive=ceil(end/W)*W`, and
  `bucket_count=(bucket_end_exclusive-first_bucket_start)/W`. An `end` exactly
  on a boundary does not create an extra bucket.
- A source row participates exactly when
  `flow_start_utc >= start && flow_start_utc < end`. It belongs to the bucket
  containing `flow_start_utc`; counters are never duplicated or prorated.
- Bucketed edges group by bucket start followed by the existing default key:
  source endpoint, destination endpoint, IP protocol, destination-port
  presence, and destination-port value.
- `network_flow_bucket_edge_id_v1` returns `nfbe_` plus 64 lowercase SHA-256
  hex characters over a length-framed transcript containing the algorithm
  domain, incident ID, canonical bucket start and end, endpoint IDs, protocol,
  and the existing destination-port presence/value pair. Including the end
  prevents equal-start buckets of different widths from sharing an identity.
- The Graph Projection adapter uses source relationship kind
  `network_flow.bucketed_flow_edge.v1` and projection version
  `network_flow_activity.time_bucket.v1`. Each edge contains canonical
  `bucket_start_utc` and `bucket_end_utc` properties.
- `time_buckets[]` is Network Flow response metadata, never Graph Projection
  result state or a new persistence family. It appears only on the temporal
  result variant, is ordered by bucket start, and includes every intersecting
  bucket including empty buckets. Each item contains its bounds and aggregate
  unique-vertex, edge, and contributing-row counts without duplicating edge
  payloads.
- A bucket-edge contributor selector returns only rows assigned to the exact
  bucket and edge key. A vertex selector returns the qualifying rows touching
  that endpoint across the requested interval.
- Total result vertices and edges remain subject to effective Network Flow
  limits and Graph Projection's immutable maxima. No partial bucket or graph is
  returned on a limit failure.

The v2 graph-query digest uses a new length-framed
`cartulary.network_flow.graph_query_digest.v2` transcript that binds the v2
semantic-query schema and complete normalized aggregation variant. Deployment
limits and lower per-request overrides remain excluded. The existing source
snapshot algorithm may continue to bind the new graph digest without changing
its own version because the graph digest is already an explicit framed input.

Major-4 contributor selectors always use Network Flow source IDs, never
projected `ed_` IDs. Responses provide canonical selector objects: vertex ID
plus endpoint value; default edge ID plus both endpoint values, protocol, and
destination-port presence/value; or bucket edge ID plus that key and bucket
bounds. The server recomputes the supplied ID and binds the complete selector
into cursor identity before using fixed selector-aware SQL predicates.

#### 10.2.2 Resource configuration

Configuration contract major 2 adds optional
`network_flow_activity.resource_limits`. It is forbidden for an unclaimed
profile. Unknown members, explicit null, non-integers, out-of-range values, and
invalid active/retained relationships fail before any listener or worker
starts. Omitted members use the defaults below. Runtime receives one immutable
effective set and MUST NOT clamp or repair invalid values. Public source-profile
discovery returns every effective limit. Request overrides may only lower
vertices, edges, example references, contributing rows, and time buckets and do
not enter semantic identity.

| Limit | Minimum | Default | Maximum |
| --- | ---: | ---: | ---: |
| Active tables per incident | 1 | 128 | 128 |
| Retained tables per incident | 1 | 512 | 512 |
| Selected tables per query | 1 | 16 | 64 |
| CSV columns | 1 | 256 | 512 |
| Header scalar bytes | 1 | 256 | 256 |
| Raw cell scalar bytes | 1 | 4,096 | 16,384 |
| Rows per CSV | 1 | 250,000 | 5,000,000 |
| Accepted rows per table | 1 | 250,000 | 5,000,000 |
| Retained rejected-row diagnostics | 0 | 10,000 | 100,000 |
| Filters per query | 0 | 16 | 16 |
| Sorts per query | 0 | 8 | 8 |
| Query page limit | 1 | 500 | 1,000 |
| Graph vertices | 1 | 5,000 | 100,000 |
| Graph edges | 0 | 10,000 | 250,000 |
| Active graph views per incident | 1 | 32 | 32 |
| Retained graph views per incident | 1 | 128 | 128 |
| Nonterminal graph jobs per incident | 1 | 4 | 4 |
| Example row references per edge | 0 | 10 | 100 |
| Binding source-row references | 1 | 16 | 1,000 |
| Aggregate counter digits | 1 | 39 | 128 |
| Contributing graph rows | 1 | 250,000 | 5,000,000 |
| Time buckets | 1 | 256 | 1,024 |
| Materialization timeout seconds | 30 | 300 | 3,600 |

The materialization deadline starts when the saved-graph handler begins and
covers source validation, scanning, aggregation, and projection. Expiry before
final publication produces a terminal safe timeout and preserves the prior
selected result. Once final transactional publication begins before the
deadline, existing indeterminate-commit reconciliation governs the outcome;
timeout MUST NOT reverse a possibly committed result. Raising a deployment
above the conservative defaults requires retained GP3 capacity evidence;
configuration acceptance is not a hardware-independent performance claim.

### 10.3 Compatibility and rollback policy

| Surface | Classification | GP3 outcome |
| --- | --- | --- |
| Graph Projection v2 engine/result identity | Preserve | Temporal behavior is expressed through Network Flow input and projection-version selection; no Graph v3 protocol is introduced |
| Network Flow public major 3 | Replace for current clients | Current browser and generated consumers move directly to major 4; no dual generated client or response parser is retained |
| Persisted semantic-query v1 declarations | Preserve with continuing value | Read, refresh, return, back up, restore, and rebuild; reject for new creates and remove only after a later zero-inventory gate |
| New semantic-query v2 | Add | Required for new ephemeral queries and saved-graph creates; supports default and time-bucket variants |
| Network Flow state v2 | Migrate without rewriting | State `2 -> 3` preserves all authoritative bytes and expands validation to the v1/v2 query union |
| Reporting exact references v2 | Preserve | Full Graph binding already distinguishes temporal results through projection version and digests |
| Graph Recovery v2 algorithm | Preserve narrowly while referenced | Current rebuild registration becomes v3; retain the exact v2 dispatcher read-only until no supported retained pre-GP3 backup references it |
| Database migrations `00032` and `00033` | Preserve | Add `00034`; perform no destructive GP3 schema operation |
| Runtime limit constants | Replace as authority | Validated effective configuration becomes the runtime source; constants provide only owner-declared defaults and maxima |

Before rollout, capture and verify an exact pre-GP3 backup with the old binary,
state version, schema head, and artifact digests. Migration `00034` is additive,
but an old binary cannot interpret state version 3 or v2 semantic queries. After
any v2 declaration is persisted, rollback uses the exact pre-GP3 backup in a
replacement target with the prior binary; it is not an in-place binary
rollback. V1 declaration support is justified by authoritative installed state
and must not expand into Graph Projection v1 or a v1 new-write API.

### 10.4 GP3 gap remediation register

#### GP3-G01 — Limit owner/runtime contradiction

- **Remediation:** Separate conservative defaults from immutable maxima, add
  the typed resource-limit configuration above, inject it through application
  assembly, and correct the header default that currently exceeds the adopted
  ceiling.
- **Areas:** Specification, configuration contracts, implementation, tests,
  generated discovery, and operational documentation.
- **Rationale:** Undocumented hardcoded reductions and an impossible owner
  configuration promise prevent operators from knowing the actual envelope.
- **Long-term benefit:** One explicit capacity boundary supports safe tuning and
  later resource dimensions without scattered constants.
- **Compatibility:** Network Flow public major 4 and configuration major 2 make
  the changed effective-default contract explicit. Conservative runtime values
  remain defaults rather than being silently raised.
- **Unresolved risk:** Deployments may unknowingly underdeliver owner behavior or
  raise workloads beyond proven capacity.
- **Validation:** Startup, inactive-profile, null, unknown-key, minimum,
  maximum, cross-limit, discovery, request-override, and assembly tests agree on
  every effective value.

#### GP3-G02 — Whole-scope row materialization

- **Remediation:** Replace `ListRowsForTables` on Graph paths with an ordered,
  cancellation-aware iterator and incremental aggregation. Enforce contributing
  rows, vertices, edges, counter digits, examples, and cancellation while
  streaming.
- **Areas:** Network Flow specification, implementation, persistence adapters,
  and tests.
- **Rationale:** Result bounds do not protect memory when millions of source
  rows are first loaded, copied, sorted, and retained.
- **Long-term benefit:** Memory follows admitted aggregate/result state instead
  of total table scope and can support later aggregation modes.
- **Compatibility:** Streaming preserves exact persisted-v1 bytes, ordering,
  IDs, digests, examples, and errors below limits. New v2 default queries retain
  default meaning but intentionally receive v2 digest, source-snapshot, and
  result identities. Contributing-row rejection is a major-4 behavior.
- **Unresolved risk:** A permitted query can exhaust process memory before
  existing result limits execute.
- **Validation:** Golden equivalence, skew, multi-table order, maximum-row,
  `limit+1`, cancellation, allocation-shape, and database-error tests pass.

#### GP3-G03 — Contributor queries rebuild entire graphs

- **Remediation:** Revalidate the semantic digest and tables, validate the
  canonical source-ID-plus-key selector returned by major-4 responses, then use
  a selector-aware ordered scan/page path. Retain only one page and cursor
  state; do not construct unrelated graph objects or store all contributing
  rows.
- **Areas:** Network Flow specification, implementation, cursor behavior, and
  tests.
- **Rationale:** Contributor navigation is a bounded read and should not repeat
  full graph memory cost.
- **Long-term benefit:** Stable contributor latency and a reusable selection
  boundary for default and temporal edges.
- **Compatibility:** Major-4 replaces the public selector shape while retaining
  authorization, contributor ordering, and cursor behavior. Vertex, default
  edge, and temporal edge selectors bind the complete canonical key.
- **Unresolved risk:** A small contributor page can require maximum graph memory
  and fail independently of its requested limit.
- **Validation:** Vertex, default-edge, bucket-edge, continuation, stale digest,
  source deletion, authorization loss, and bounded-memory tests pass.

#### GP3-G04 — Global concurrency admits multiple large Graph jobs

- **Remediation:** Project a worker-runtime contract, acquire both global and
  per-worker capacity before claiming an execution, and continue scanning past
  saturated worker kinds. Declare one active graph attempt per process.
- **Areas:** Common Jobs, Extensions and owner specifications, generated
  contracts, implementation, and tests.
- **Rationale:** Registration count does not constrain active attempts; waiting
  inside a handler would still consume scarce global slots.
- **Long-term benefit:** Explicit resource isolation and fair future worker
  expansion without module-local semaphores.
- **Compatibility:** No public job state changes. Existing workers explicitly
  retain their effective concurrency; the Graph limit corrects an ambiguous GP2
  operational claim.
- **Unresolved risk:** Graph jobs can multiply peak memory and starve imports,
  Reporting, or other background work.
- **Validation:** Same-kind saturation, mixed-kind fairness, notifications,
  recovery scans, retry, lease renewal, shutdown, and component-loss tests pass.

#### GP3-G05 — Result cleanup is dormant

- **Remediation:** Network Flow owns a private
  `network_flow_activity.graph_result_cleanup.v1` dispatcher. Application
  assembly starts it only after readiness, runs one immediate sweep and then a
  five-minute base cadence, coalesces ticks, and closes it during ordered
  shutdown. `has_more` schedules a non-overlapping five-second paced
  continuation; transient failure uses bounded retry rather than a busy loop.
- **Areas:** Network Flow and application lifecycle specifications,
  implementation, tests, and operations.
- **Rationale:** A tested port without a production caller provides no storage
  lifecycle.
- **Long-term benefit:** Derived storage remains self-maintaining without a
  public route, Common Job, or Graph-owned hidden worker.
- **Compatibility:** No route or result behavior changes. Cleanup affects only
  unreachable derived rows.
- **Unresolved risk:** Refreshes and retirement grow result/object storage
  indefinitely.
- **Validation:** Readiness ordering, immediate/cadenced execution, paced drain,
  transient failure, unexpected dispatcher termination, clean shutdown, and
  restart pass.

#### GP3-G06 — Cleanup is not future-owner-safe

- **Remediation:** Select candidates by exact `source_owner_id`; lock result
  rows, query Network Flow selected bindings in the same borrowed transaction,
  recheck unexpired leases, then atomically delete only unselected candidates.
  Use `FOR UPDATE SKIP LOCKED` and enforce result-before-declaration lock order
  for cleanup, publication replay, and Reporting lease admission.
- **Areas:** Graph port specification, Network Flow implementation, PostgreSQL
  adapter, migration, and tests.
- **Rationale:** A deployment-wide caller-supplied reachable list is unbounded,
  race-prone, and unsafe when another producer is added.
- **Long-term benefit:** Each future source owner can clean only its own derived
  results through the same narrow protocol.
- **Compatibility:** Replace the unused cleanup signature directly; no
  production caller requires an adapter.
- **Unresolved risk:** A future owner can suffer cross-owner deletion, or a
  concurrent publication/lease can race cleanup.
- **Validation:** Cross-owner isolation, publication races, selected-result
  preservation, new/renewed lease races, multi-instance exclusion, and rollback
  pass.

#### GP3-G07 — Lease and deletion maintenance is insufficiently bounded

- **Remediation:** Delete at most 1,000 expired leases per transaction and one
  result envelope plus its cascading objects per result transaction. A sweep
  drains at most eight results or 30 seconds, whichever comes first, and reports
  remaining work for a paced continuation.
- **Areas:** Graph storage contract, implementation, migration/indexing, tests,
  and operations.
- **Rationale:** One unbounded expired-lease statement or multi-result cascade
  can create long locks and unpredictable shutdown.
- **Long-term benefit:** Restart-safe, measurable maintenance with bounded
  transaction scope.
- **Compatibility:** Leases retain their existing expiry semantics; no semantic
  result retention interval is introduced.
- **Unresolved risk:** Cleanup itself can become a database pressure event.
- **Validation:** Batch boundaries, `has_more`, large cascades, statement
  failure, cancellation, crash rollback, and eventual drain pass.

#### GP3-G08 — Operational visibility is incomplete

- **Remediation:** Add Common Jobs queued-count and queue-wait-duration signals;
  add a Network Flow instrumentation scope for materialization phase duration,
  source rows, result vertices/edges/buckets, cleanup outcome/duration/deletion,
  eligible backlog, and oldest eligible result age measured from `published_at`.
  The signal MUST NOT claim to measure time since a result became unreachable,
  because that transition is not persisted.
- **Areas:** OpenTelemetry and Common Jobs specifications, registries,
  application observers, implementation, tests, and operational guidance.
- **Rationale:** Generic terminal duration cannot distinguish queue pressure,
  source composition, projection, publication, or failed maintenance.
- **Long-term benefit:** Capacity and failure diagnosis remain vendor-neutral and
  safe as graph modes grow.
- **Compatibility:** Additive internal telemetry only. Closed attributes are
  operation, phase, graph mode, result, job kind, and safe error class; no
  incident, graph, result, digest, row, label, property, SQL, or raw error value
  is emitted.
- **Unresolved risk:** Operators cannot distinguish slow computation, queue
  starvation, database publication, or cleanup lag.
- **Validation:** Registry, signal shape, disabled exporter, error mapping,
  cardinality, sentinel privacy, and telemetry self-failure tests pass.

#### GP3-G09 — Capacity evidence is not deployment-calibrated

- **Remediation:** Define reproducible default, raised, semantic-maximum,
  high-cardinality, dense-edge, single-edge-skew, maximum-bucket, cancellation,
  cleanup, and mixed-job workloads. Retain wall time, allocations, peak RSS,
  database work, output bytes, and environment identity.
- **Areas:** Testing Harness inputs, tests, operational documentation, and
  handoff evidence.
- **Rationale:** A functional maximum-size test is not a hardware capacity
  claim.
- **Long-term benefit:** Deployments can select effective limits from evidence
  without weakening stable semantic maxima.
- **Compatibility:** No project-wide SLO is invented. A deployment that fails a
  raised profile lowers effective configuration rather than changing identity
  or output semantics.
- **Unresolved risk:** Operators may choose limits that cause OOM, excessive
  queueing, or shutdown overruns.
- **Validation:** Retained artifacts are reproducible and bind commit,
  toolchain, workload, hardware/runtime profile, effective limits, and terminal
  outcome.

#### GP3-G10 — Time-bucketed graph behavior is undefined

- **Remediation:** Adopt and implement the complete §10.2.1 query, digest,
  bucket, edge identity, result, contributor, error, and limit contract in
  Network Flow.
- **Areas:** Network Flow specification, contracts, implementation, tests, and
  domain documentation.
- **Rationale:** Temporal grouping is source-owner product meaning, not a reason
  to make the pure Graph engine time-aware.
- **Long-term benefit:** One explicit temporal model can grow to later graph
  analysis without corrupting default edge identity.
- **Compatibility:** Contract major 4 and schema/digest v2 make the new semantics
  explicit. Default v1 persisted results remain exact.
- **Unresolved risk:** Local bucket interpretations would diverge across saved
  graphs, contributors, Reporting, and Recovery.
- **Validation:** Boundary alignment, negative epochs, exact-boundary end,
  empty buckets, row assignment, no counter duplication, limits, IDs, digests,
  examples, cancellation, and property tests pass.

#### GP3-G11 — Temporal persistence and exact consumers lack migration rules

- **Remediation:** Adopt Network Flow state 3's v1/v2 declaration union,
  Recovery v3 mixed rebuild, exact read-only v2 dispatch for still-supported
  backups, and Reporting redaction candidates while preserving exact reference
  v2.
- **Areas:** Network Flow, Extensions, Recovery, Reporting, contracts,
  implementation, tests, and backup guidance.
- **Rationale:** Temporal output is not production-ready if authoritative
  declarations outlive their parser or restored results change identity.
- **Long-term benefit:** Exact mixed-generation restore and consumer reuse with
  no Graph v1 resurrection or heuristic translation.
- **Compatibility:** Read-old/write-new is intentional. Later v1 declaration or
  Recovery v2 removal needs both zero installed v1 declarations and zero
  supported retained backup references.
- **Unresolved risk:** Upgrades can strand saved graphs or reports, and restores
  can claim readiness with different objects.
- **Validation:** State `2 -> 3`, mixed list/read/refresh, backup, isolated
  restore, exact result/object identity, Reporting lease, and readiness pass.

#### GP3-G12 — No complete temporal and operational workflow

- **Remediation:** Add major-4 default/time-bucket controls, complete-range and
  width validation, bucket navigation, saved graph lifecycle integration,
  contributor pivots, bounded visualization, capacity guidance, cleanup
  response, and rollout/rollback procedures.
- **Areas:** Web implementation, public contracts, browser tests, design/domain
  documentation, and operations.
- **Rationale:** Backend semantics and maintenance without discoverable user and
  operator workflows remain incomplete product behavior.
- **Long-term benefit:** Temporal analysis fits the existing Network Analysis
  workspace and production operations remain repeatable.
- **Compatibility:** Consume only the major-4 client. Do not retain a major-3
  browser decoder, parallel application shell, or public cleanup control.
- **Unresolved risk:** Users cannot understand temporal scope, large results, or
  stale state, and operators cannot deploy or recover the feature safely.
- **Validation:** Unit, stateful browser, accessibility, measurement, visual,
  backup/restore, rollout, and rollback drills pass.

### 10.5 Linear workstream plan

All workstreams are strictly linear. A slice begins only after its predecessor
passes narrow validation and this tracker contains the predecessor's completed
checkpoint.

| Slice | Status | Depends on | Exit outcome | Next action |
| --- | --- | --- | --- | --- |
| GP3-S00 | `DONE` | GP2 complete | GP3 plan, inventory, decisions, and document validation are recorded | GP3-S01 |
| GP3-S01 | `DONE` | GP3-S00 | Every affected owner adopts a complete, contradiction-free contract | GP3-S02 |
| GP3-S02 | `DONE` | GP3-S01 | Contracts, generated artifacts, and verification routing reproduce owners without drift | GP3-S03 |
| GP3-S03 | `DONE` | GP3-S02 | Effective resource configuration and discovery are exact and fail closed | GP3-S04 |
| GP3-S04 | `DONE` | GP3-S03 | Graph and contributor source processing is ordered, streaming, and bounded | GP3-S05 |
| GP3-S05 | `DONE` | GP3-S04 | Per-worker admission isolates Graph jobs without starving other work | GP3-S06 |
| GP3-S06 | `DONE` | GP3-S05 | Production result/lease cleanup is source-safe, bounded, and lifecycle-composed | GP3-S07 |
| GP3-S07 | `DONE` | GP3-S06 | Queue, materialization, and cleanup signals are complete and privacy-safe | GP3-S08 |
| GP3-S08 | `DONE` | GP3-S07 | Default and temporal Graph backend behavior passes exact conformance | GP3-S09 |
| GP3-S09 | `DONE` | GP3-S08 | Mixed declaration Recovery and temporal Reporting pass exact-result gates | GP3-S10 |
| GP3-S10 | `DONE` | GP3-S09 | Major-4 temporal workflow passes frontend and browser gates | GP3-S11 |
| GP3-S11 | `DONE` | GP3-S10 | Retained capacity and rollout evidence supports the selected production profile | GP3-S12 |
| GP3-S12 | `DONE` | GP3-S11 | All GP3 gaps close and final narrow, broad, and release gates pass | None |

#### GP3-S00 — Baseline and tracker control

- Reconcile the completed GP2 handoff with the GP3 plan without deleting or
  rewriting GP2 evidence.
- Record current implementation seams, owner contradictions, compatibility
  decisions, migration number, slice ordering, and validation policy.
- Change only this tracker during the document-update step.

Exit: this tracker is the complete controlling plan; Markdown validation and
diff hygiene pass; GP3-S00 is `DONE`; only GP3-S01 is next.

#### GP3-S01 — Owner adoption

- Adopt Network Flow major 4, state 3, configuration 2, semantic-query v2,
  temporal identities, effective limits, streaming behavior, cleanup ownership,
  telemetry, and rollback semantics.
- Adopt source-owner-scoped cleanup in Graph Projection without changing the
  semantic `graph_projection.v2` engine contract.
- Adopt worker runtime capacity in the appropriate Core/Common Jobs and
  Extensions owners; adopt the OpenTelemetry signals, Reporting temporal
  behavior, Recovery v3, and affected Core owner facts.
- Update `docs/domain.md` only for the adopted temporal-graph vocabulary while
  preserving saved declaration, immutable result, job attempt, and Reporting
  diagram distinctions.

Exit: owner review and Markdown validation prove a complete, non-contradictory
requirements chain. No machine projection is treated as authority.

#### GP3-S02 — Contract projection and verification routing

- Author v2 public and semantic schemas, v1/v2 persisted unions, major-4 routes
  and errors, state `2 -> 3`, configuration major 2, worker runtime contracts,
  telemetry registries, Recovery v3 artifacts, and temporal fixtures.
- Update Core recognition, Extensions fragments/bindings, client support, and
  all affected verification rows.
- Regenerate Go, TypeScript, OpenAPI, topology, schedules, and other managed
  outputs only through Make-owned workflows.

Exit: generation, generated drift, JSON shape, generated-artifact policy,
contract shape, and import-boundary checks pass.

#### GP3-S03 — Resource configuration

- Resolve the configuration before runtime composition and inject one immutable
  effective Network Flow limit set through application assembly.
- Enforce inactive-profile, explicit-null, type, minimum, maximum, unknown-key,
  and cross-limit failures before listeners or workers start.
- Return the complete effective limits from source-profile discovery and apply
  lower request overrides without changing graph identity.

Exit: Network Flow, Extensions, platform configuration, application startup,
and discovery tests pass for defaults, valid overrides, and every failure class.

#### GP3-S04 — Bounded source and contributor pipeline

- Replace complete source-row slices with stable ordered iteration and
  incremental default aggregation.
- Store aggregate counts and bounded example refs rather than contributing row
  slices on every vertex and edge.
- Return canonical contributor selector objects, recompute their source IDs,
  and implement selector-aware iteration and page construction without
  reconstructing unrelated graph objects.

Exit: persisted-v1 output/digest equivalence, intentional v2 identity fixtures,
multi-table ordering, selector validation, cursor, authorization, stale-source,
maximum-row, allocation, limit, and cancellation evidence passes.

#### GP3-S05 — Worker isolation and backpressure

- Project per-worker capacity into runtime selection and acquire worker plus
  global capacity before execution claim.
- Continue durable selection past a saturated worker kind and keep notification
  delivery best-effort while periodic scans remain authoritative.
- Preserve existing job state, lease, retry, shutdown, and fatal component-loss
  semantics.

Exit: graph concurrency is exactly one per serving process; mixed workers make
progress; restart, retry, renewal, cancellation, and shutdown tests pass.

#### GP3-S06 — Durable result and lease cleanup

- Add migration `00034` with only owner-approved candidate-selection indexes.
- Implement source-owner candidate selection, selected-binding checks, lease
  rechecks, bounded expired-lease deletion, one-result atomic cascades, and the
  shared result-before-declaration lock order.
- Compose the private dispatcher after serving readiness and stop it in reverse
  application lifecycle order. Use five-minute base cadence and paced
  five-second continuations only while bounded work reports `has_more`.

Exit: migration drift, transaction invariants, cross-owner isolation,
publication/lease races, multi-instance exclusion, bounded work, crash rollback,
readiness, and shutdown pass.

#### GP3-S07 — Operational telemetry

- Add `cartulary.jobs.queued` and queue-wait duration by closed job kind.
- Add Network Flow materialization phase duration and source/result volume
  signals plus cleanup operation, duration, deletion, backlog, and oldest-age
  signals.
- Keep the pure Graph engine free of telemetry dependencies; application
  composition injects observers.

Exit: signal registry, exact attributes, success/failure/cancel/timeout,
export-disabled behavior, bounded cardinality, and privacy sentinel tests pass.

#### GP3-S08 — Time-bucket backend

- Implement v2 admission/normalization, digest v2, bucket arithmetic, bucket
  edge identity including both bucket bounds, streaming temporal aggregation,
  Network Flow response-owned bucket summaries, saved materialization,
  contributors, limits, and safe errors.
- Use the unchanged Graph Projection v2 engine with mode-specific Network Flow
  mapping, relationship kind, properties, and projection version.
- Preserve internal v1 declaration execution and prohibit v1 new creates.

Exit: golden, boundary, empty-bucket, negative-epoch, DST-independence,
maximum-bucket, identity, property, retry, stale-generation, cancellation, and
fuzz/property tests pass.

#### GP3-S09 — Recovery and Reporting

- Publish the Recovery v3 registration and rebuild mixed v1/v2 declarations
  from authoritative Network Flow state while workers remain quiescent.
- Retain exact Recovery v2 dispatch read-only while supported retained backups
  reference it; do not translate or rewrite its artifacts.
- Require exact result, vertex, edge, bucket, digest, job, and lease
  reconciliation before readiness.
- Route typed Network Flow graph label components through Reporting redaction.
  Internal labels are deterministic post-redaction values with ordinal fallback;
  external Graph release remains fail-closed without a separately adopted rule.

Exit: state migration, mixed backup, isolated restore, exact identity,
readiness, selection, redaction, render bounds, retry, and lease reconciliation
pass.

#### GP3-S10 — Web product completion

- Regenerate and consume only the major-4 client.
- Add default/time-bucket selection, complete-range validation, fixed width
  choices, ordered bucket navigation, empty-bucket state, saved lifecycle,
  job/failure state, and contributor pivots inside Network Analysis.
- Mount no more than 500 vertices and 1,000 edges and expose a selected bucket
  or bounded summary for larger exact results.

Exit: frontend type/unit/import-boundary, stateful browser, keyboard,
accessibility, measurement, and visual checks pass for all roles and lifecycle
states.

#### GP3-S11 — Capacity and rollout certification

- Run the adopted default, raised, semantic-maximum, skew, dense, temporal,
  cleanup, cancellation, crash, queue-pressure, and shutdown workloads.
- Retain measurements with commit, toolchain, workload, hardware/runtime
  profile, effective limits, artifacts, and outcome. Lower deployment limits
  when a raised profile fails; do not change stable Graph semantics to make a
  benchmark pass.
- Produce a fresh state-v3 backup, restore mixed v1/v2 declarations in
  isolation, and exercise major-4 rollout plus replacement-target rollback.

Exit: retained evidence supports the selected deployment limits; restore and
rollback drills pass; no hardware-independent SLO or unclassified operational
risk is claimed.

#### GP3-S12 — Final validation and handoff

- Run affected owner guides and narrow slices before formatting, generation,
  drift, configuration, migration, generated-artifact, JSON-shape, security,
  telemetry, frontend, and browser gates.
- Run `make agent-finalize` before broad verification, passing a retained
  successful `RESULTS_DIR` when available and recording the unset case
  otherwise.
- Run `make check` and `make release-check` and complete every gap, file,
  command, run root, compatibility, rollback, residual-risk, and skipped-check
  disposition in this tracker.

Exit: GP3-G01 through GP3-G12 are closed with evidence, every slice is `DONE`,
no temporary compatibility path is unclassified, and this tracker is the final
handoff.

### 10.6 GP3 validation and checkpoint policy

After GP3-S00, every slice begins with:

```text
make task-guide ROLE=module-author OWNER=<affected-owner>
make explain-test-owner OWNER=<affected-owner>
```

The implementer then runs the narrowest routed `make test-slice` or
`make service-backed-test-slice` rows for every affected owner. Required owner
families across GP3 are `module.graphprojection`, `module.networkflow`,
`platform.jobs`, `module.extensions`, `platform.config`, `platform.telemetry`,
`module.reporting`, `module.recovery`, `app.server`, and `web.networkflow` as
their slices become affected.

| Validation layer | Required point |
| --- | --- |
| `git diff --check` and `make lint-markdown` | Every checkpoint |
| `make format` | Every implementation slice with authored Go or frontend changes |
| `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check` | Contract/projection changes and final gate |
| `make migration-drift` | GP3-S06 and final gate |
| Backend/frontend import-boundary checks | Every affected implementation slice |
| Owner unit/service-backed slices | Before advancing every affected owner |
| Frontend type/unit plus browser/stateful/a11y/measurement/visual | GP3-S10 onward as routed |
| Capacity and retained-run explanation | GP3-S11 |
| `make agent-finalize`, `make check`, `make release-check` | GP3-S12 in that order |

After each slice and before the next begins, record status, files, substantive
changes, exact commands, run roots/artifacts, failures and relatedness,
compatibility impact, residual risks, skipped conditional checks, and the one
next action. Allowed statuses remain `PLANNED`, `IN PROGRESS`, `BLOCKED`, and
`DONE`. An owner, configuration, migration, privacy, restore, rollback, or final
gate failure marks the slice `BLOCKED` rather than advancing.

### 10.7 GP3 binary completion criteria

GP3 is complete only when all of the following are true:

- Network Flow major 4, state 3, and configuration major 2 are adopted and
  projected without drift.
- Graph Projection remains a pure v2 engine with source-owner-scoped cleanup
  ports and no temporal product policy or hidden worker.
- Effective Network Flow resource limits are validated, discoverable, injected,
  and bounded by immutable maxima.
- Default and contributor Graph processing is streaming and does not retain
  whole-scope contributing-row collections.
- The graph worker has exactly one active attempt per process without consuming
  waiting global job slots or starving other workers.
- Expired leases and unreachable Network Flow results are cleaned in bounded,
  multi-instance-safe transactions after readiness and before shutdown.
- Queue, materialization, and cleanup health are observable through closed,
  privacy-safe telemetry.
- V2 default and time-bucket queries have exact schemas, digests, identities,
  limits, contributors, saved materialization, and deterministic results.
- V1 declarations remain supported only for installed authoritative state and
  cannot be created through the current public write contract.
- Recovery v3 reconstructs mixed v1/v2 results exactly before readiness, and
  Reporting renders classified/redacted temporal edges through unchanged exact
  references v2.
- The major-4 browser workflow is complete, accessible, stateful, and bounded.
- Retained capacity evidence supports the selected deployment profile without
  claiming a universal SLO.
- Fresh backup, isolated restore, rollout, and replacement-target rollback
  drills pass.
- Final narrow, broad, and release gates pass and every gap has a terminal
  disposition.

### 10.8 GP3-S00 completion checkpoint

GP3-S00 changed only this tracker. It preserved the GP2 completion record,
established the owner-first GP3 plan, classified existing semantic-query v1
state as justified read/refresh/restore compatibility, reserved migration
`00034`, and fixed strict slice sequencing through the final handoff.

The final reconciliation distinguishes exact persisted-v1 identity from
intentional v2 identity, keeps `time_buckets[]` in the Network Flow response
instead of Graph result state, includes both bucket bounds in temporal-edge
identity, closes major-4 contributor selector keys, uses paced cleanup drain,
names only measurable cleanup eligibility age, and retains exact Recovery v2
dispatch solely while supported retained backups reference it.

#### Files changed

```text
docs/handoffs/graphprojection-module-refactor-tracker.md
```

#### Explicitly unchanged

- Adopted owner specifications, including `docs/domain.md`
- Contracts, generated roots, migrations, implementation, tests, and lockfiles
- Graph Projection v2 behavior and the GP2 production database state
- GP2 commands, run roots, compatibility outcomes, and release evidence

#### Validation

- `git diff --check` and `git diff --cached --check` — PASS with no output.
- `make lint-markdown` — PASS. Final reconciliation run root:
  `.cartulary/test-results/20260816T125233Z-p1539914`; summary:
  `adhoc/lint-markdown/tool-run-summary.json`.
- Broader product, generation, migration, service-backed, frontend, browser,
  security, and release checks were not run because GP3-S00 is intentionally a
  tracker-only planning checkpoint and changes no governing or executable
  artifact.

#### Current next action

Begin GP3-S01 owner adoption. Do not author contracts, migrations, generated
artifacts, or implementation until every affected owner accepts the complete
major-4/state-3/configuration-2, cleanup, concurrency, telemetry, temporal,
Reporting, Recovery, and rollback contract.

### 10.9 GP3-S01 completion checkpoint

Status: `DONE`.

GP3-S01 adopted one coherent owner chain before changing any machine
projection. Network Flow is now public major 4, durable state 3, and
configuration major 2. Its owner defines semantic-query v2, exact epoch-aligned
temporal aggregation and identity, canonical Network Flow selectors, ordered
streaming and contributor paging, immutable effective limits, materialization
timeout, bounded cleanup, privacy-safe telemetry, mixed-generation Recovery,
Reporting redaction, the bounded browser workflow, and replacement-target
rollback.

The companion owners now agree that Graph Projection remains a pure v2 engine;
cleanup uses borrowed source-owner-scoped transactions; Jobs consumes generated
per-worker capacity and reserves global plus worker slots before claim;
Extensions publishes the runtime mapping; Core owns current recognition,
discovery, client support, Recovery dispatch, and private dispatcher lifecycle;
OpenTelemetry owns the closed Jobs and Network Flow signals; and Reporting
keeps `source_projection_ref.v2` while treating Network Flow label components
as redaction candidates. The former resource table contradiction was removed:
Table 20-A is the single min/default/max registry and its header scalar default
is exactly 256.

#### Files changed

```text
docs/domain.md
docs/extension-subsystem-nlspec.md
docs/graph_projection_nlspec.md
docs/network-flow-activity-nlspec.md
docs/opentelemetry-instrumentation-nlspec.md
docs/reporting-subsystem-nlspec.md
docs/spec/00_document_set_status_and_precedence.md
docs/spec/01_architecture_storage_and_view_contracts.md
docs/spec/03_workbook_interaction_collaboration_and_workflows.md
docs/spec/04_security_deployment_and_conformance.md
docs/spec/I_projection_authority_boundary_and_characterization.md
docs/handoffs/graphprojection-module-refactor-tracker.md
```

#### Compatibility and migration impact

- New public Network Flow graph writes are major-4 semantic-query v2 only.
  Persisted v1 declarations retain exact read, refresh, backup, restore,
  rebuild, source-snapshot, result, and byte compatibility.
- V2 default queries intentionally cross the digest-v2 identity boundary even
  though their default graph meaning matches v1.
- State `2 -> 3` is byte-preserving. Recovery v3 is current; the exact v2
  dispatcher is retained read-only only for supported retained pre-GP3 backup
  catalogs.
- Graph Projection remains `graph_projection.v2`, Reporting references remain
  `source_projection_ref.v2`, migrations `00032` and `00033` remain immutable,
  and GP3 reserves only additive migration `00034`.
- After the first v2 declaration, rollback is replacement-target restore from
  the exact pre-GP3 backup with the prior binary; in-place old-binary rollback
  is unsupported.

#### Validation

- `make task-guide ROLE=module-author OWNER=<owner-id>` — PASS for
  `module.networkflow`, `module.graphprojection`, `platform.jobs`,
  `module.extensions`, `platform.telemetry`, `module.reporting`,
  `module.recovery`, and `app.server`.
- `make explain-test-owner OWNER=<owner-id>` — PASS for the same eight owners;
  every owner resolved to its current manifest and focused target.
- `git diff --check` and `git diff --cached --check` — PASS with no output.
- `make lint-markdown` — PASS. Owner-adoption run root:
  `.cartulary/test-results/20260816T131042Z-p1551964`; summary:
  `adhoc/lint-markdown/tool-run-summary.json`. Tracker-checkpoint follow-up root:
  `.cartulary/test-results/20260816T131232Z-p1553347`.
- Product unit, service-backed, generation, migration, frontend, browser,
  security, broad, and release checks were intentionally not run in this
  owner-only slice. They cannot validate the newly adopted behavior before the
  downstream contracts and implementation exist; S02 through S12 own those
  gates.

No owner, configuration, migration, privacy, restore, rollback, or release gate
failed. No generated root, contract, migration, implementation source, test,
lockfile, or retained runtime artifact changed in S01. The only residual risk is
the expected temporary projection lag between the newly adopted owners and the
pre-S02 machine artifacts; it is classified and bounded by the strict linear
sequence.

#### Current next action

Execute GP3-S02 contract projection and verification routing. Do not begin
resource-configuration implementation until all affected schemas, owner facts,
generated outputs, fixtures, and test routes reproduce the adopted owners
without drift.

### 10.10 GP3-S02 completion checkpoint

Status: `DONE`.

GP3-S02 projected the adopted GP3 owner chain without moving temporal policy
into Graph Projection. Network Flow is represented as public major 4 with a
closed semantic-query v2 union, mixed persisted v1/v2 declaration reads,
canonical Network Flow contributor selectors, response-owned temporal bucket
summaries, the complete configuration-major-2 effective-limit registry, and
the new safe error reasons. The unchanged Graph Projection v2 contract now has
a separate source-owner-scoped borrowed-transaction maintenance projection.

Extensions generates an exact worker-runtime artifact for every worker kind,
including a limit of one for the graph worker and eight for existing workers.
Its Network Flow descriptor, client support, state-3 admission/migration,
implementation binding, Reporting participant, and Recovery contribution are
version-closed. Recovery publishes the current mixed-query v3 source registry,
binding, result, and terminal journal evidence while retaining the exact v2
artifacts and journal v3 as historical readers. The Jobs and Network Flow
signal inventory is projected through a closed OpenTelemetry registry. All
managed Go, TypeScript, import, topology, and contract outputs were regenerated
through `make generate`.

#### Files changed

Authored projection inputs and schemas:

```text
contracts/extensions/{build,fragments,profiles,specification}/**
contracts/extensions/dependencies.json
contracts/graph-projection/{index.json,storage-maintenance.v1.json}
contracts/imports/index.json
contracts/network-flow/{index.json,routes.v1.json,errors.v1.json}
contracts/network-flow/{schemas.v2.json,frontend-entrypoints.v4.json}
contracts/network-flow/{graph-semantics.v2.json,resource-limits.v2.json}
contracts/network-flow/{resource-limits-config.v2.schema.json,reporting-graph-source.v2.schema.json}
contracts/otel/cartulary_signal_registry.v1.json
contracts/protocol-ts/frontend-entrypoints.v2.json
contracts/recovery/{index.json,graph-projection-restore-*.v3.schema.json}
contracts/recovery/operator-recovery-journal-payload.v4.schema.json
contracts/recovery/fixtures/{graph-projection-restore-*.v3.json,operator-recovery-journal-payload.v4.json,recovery-state-catalog.v1.json}
tools/contractgen/{extensions_generation.go,extensions_validation.go,main.go,recovery_validation.go}
tools/harness/generated-artifacts/check-json-shapes.mjs
tools/otel/check-otel-conformance.mjs
tools/schemas/cartulary.network_flow_*.schema.json
tools/schemas/cartulary.otel_signal_registry.v1.schema.json
tools/harness_schema_attachments.json
```

Verification routing and assertions:

```text
internal/modules/extensions/contract_test.go
internal/modules/graphprojection/v2_contract_projection_test.go
internal/modules/networkflow/v3_contract_projection_test.go
internal/modules/recovery/vnext_graph_restore_artifacts_test.go
tools/test_families/module.extensions.json
tools/test_families/module.networkflow.json
tools/test_families/module.recovery.json
```

Generated outputs changed only through the Make-owned generator:

```text
internal/gen/contract{extensions,graphprojection,imports,networkflow,recovery}/**
internal/gen/importtargetregistry/registry_gen.go
packages/protocol-ts/src/generated/**
tools/execution_topology_render_index.json
```

The S01 owner documents and this tracker remain part of the cumulative GP3
worktree. Removed v1/index-v2/frontend-v3 source paths were replaced by their
new current versions rather than edited as generated compatibility aliases.
No lockfile, migration, product database schema, or Graph Projection engine
algorithm changed in S02.

#### Compatibility and migration impact

- Public major-4 graph requests and creates project semantic-query v2 only;
  v1 schemas remain registered for exact installed declaration reads,
  refreshes, results, backup, restore, and rebuild.
- Unchanged rename, refresh, and retire request shapes keep their v1 schema
  identities while their enclosing routes and results are projected into
  major 4. Current browser support points only to
  `network_flow_activity.standard.v4`.
- V2 default queries intentionally use digest v2 and therefore new source
  snapshot/result identities. Persisted v1 declarations retain their original
  digest, projection version `network_flow_activity.v1`, and result bytes.
- Recovery v3 accepts semantic-query v1 and v2. Exact v2 binding, registry, and
  result artifacts plus journal payload v3 remain strict historical readers;
  current terminal evidence uses journal payload v4 and rebuild result v3.
- State migration `2 -> 3` is projected as byte-preserving. Runtime migration,
  configuration, worker admission, temporal execution, Recovery dispatch, and
  Reporting consumption remain deliberately assigned to S03 through S09.

#### Validation

- All ten affected owner guides and test-owner explanations passed for
  `module.graphprojection`, `module.networkflow`, `platform.jobs`,
  `module.extensions`, `platform.config`, `platform.telemetry`,
  `module.reporting`, `module.recovery`, `app.server`, and `web.networkflow`.
- `make format` — PASS. Latest S02 formatting root:
  `.cartulary/test-results/20260816T135309Z-p1635842`.
- `make generate` — PASS. Final S02 generation root:
  `.cartulary/test-results/20260816T135412Z-p1642298`.
- `make generate-drift` — PASS, 4/4 units, root
  `.cartulary/test-results/20260816T135446Z-p1645861`.
- `make generated-artifact-policy-check` — PASS, 3/3 units, root
  `.cartulary/test-results/20260816T135446Z-p1645880`.
- `make json-shape-check` — PASS, 3/3 units, root
  `.cartulary/test-results/20260816T135446Z-p1645897`.
- `make frontend-import-boundary-check` — PASS, 2/2 units, root
  `.cartulary/test-results/20260816T135446Z-p1646151`.
- `git diff --check` and `git diff --cached --check` — PASS with no output.
- `make lint-markdown` — PASS. Completion-content run root:
  `.cartulary/test-results/20260816T135607Z-p1650243`; summary:
  `adhoc/lint-markdown/tool-run-summary.json`.
- `make test-slice OWNER=module.extensions
  ROWS=module.extensions.unit.network_flow_v4_state_jobs_reporting_projection`
  — PASS, root `.cartulary/test-results/20260816T135222Z-p1631866`.
- `make test-slice OWNER=module.graphprojection
  ROWS=module.graphprojection.storage.v2_contract_projection` — PASS, root
  `.cartulary/test-results/20260816T135319Z-p1639395`.
- `make test-slice OWNER=module.networkflow
  ROWS=module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58`
  — PASS, root `.cartulary/test-results/20260816T135319Z-p1639396`.
- `make test-slice OWNER=module.recovery
  ROWS=module.recovery.unit.graph_restore_v3_contract_projection` — PASS, root
  `.cartulary/test-results/20260816T135426Z-p1645264`.

Early generation failures at
`.cartulary/test-results/20260816T134127Z-p1598056`,
`.cartulary/test-results/20260816T134203Z-p1600147`, and
`.cartulary/test-results/20260816T134257Z-p1602224` respectively exposed a
missing major-4 import owner, Recovery canonical-byte hash handling, and an
overly open conditional schema. The first JSON-shape run at
`.cartulary/test-results/20260816T134408Z-p1610636` then rejected that
conditional object. The sources were corrected to exact catalog references,
canonical hashes, and closed default/temporal union variants; the final
generation and JSON-shape runs above supersede all four failures.

Initial Graph Projection and Network Flow assertion runs at
`.cartulary/test-results/20260816T135222Z-p1631854` and
`.cartulary/test-results/20260816T135222Z-p1631859` found assertion-key errors
in the new tests, not contract defects. Corrected assertions pass in the final
narrow roots above. The broader pre-S09 Recovery runtime row failed at
`.cartulary/test-results/20260816T135329Z-p1640382` because runtime assembly
still constructs the old restore catalog; its v3 contract assertion passed in
that run and in the dedicated final row. This is expected implementation lag,
not an S02 restore gate: S09 owns current/historical dispatcher execution and
must supersede the retained failure before any restore, broad, or release gate.

Service-backed, migration, security, browser, capacity, backup/restore drill,
broad `make check`, and release checks were skipped because S02 changes only
machine projections, generated artifacts, and their narrow contract evidence.
Their mandatory owning slices remain S03 through S12. No owner,
configuration, migration, privacy, restore, rollback, or release gate required
for S02 failed.

The residual risk is explicit transitional projection lag: generated current
contracts now lead product implementation until the strictly ordered S03-S10
slices consume them. The v1 and Recovery-v2 readers are the only classified
compatibility paths; no major-3 browser client, Graph Projection v1 surface,
public cleanup route, or alternative temporal persistence was introduced.

#### Current next action

Execute GP3-S03 resource configuration. Do not begin streaming,
worker-admission, cleanup, telemetry, temporal, Recovery/Reporting, or web
implementation early.

### 10.11 GP3-S03 completion checkpoint

Status: `DONE`.

GP3-S03 makes configuration major 2 the only runtime defaulting boundary for
Network Flow capacity policy. Application assembly now injects one validated,
immutable `EffectiveLimits` value containing all 23 adopted limits. Network
Flow runtime paths no longer normalize, clamp, or repair partial limit values;
the module rejects an invalid injected value. The header-scalar default is now
the adopted 256-byte value rather than the contradictory 1,024-byte value.

The closed `network_flow_activity.resource_limits` table supports TOML and
environment-object projection, distinguishes omitted members from legitimate
zero minima, and rejects explicit null, non-object and non-integer values,
unknown keys, values outside every adopted range, and active/retained
relationship violations. Configuration snapshots deep-copy pointer-backed
overrides, and source-profile discovery returns every resolved value. Existing
Graph request overrides are checked against the injected deployment value,
including the adopted zero minima for edges and example references; they
remain outside Graph identity. The contributing-row and time-bucket override
members are projected and injected now and become executable only with the
semantic-query-v2 decoder in GP3-S08.

Current state-3 startup also registers the byte-preserving
`network_flow_activity.migrate_state_2_to_3_v1` step and v3 validators required
by the generated current profile binding. This admits the new state version
without rewriting authoritative bytes. Mixed declaration validation, Recovery
v3 execution, and Reporting consumption remain in GP3-S09.

#### Files changed

Runtime policy and composition:

```text
internal/modules/networkflow/{api.go,configuration.go,module.go,store.go}
internal/modules/networkflow/{csv_parser.go,import_facade.go,query.go,query_filter.go}
internal/modules/networkflow/{indicator_link.go,graph.go,graph_view_routes.go}
internal/modules/networkflow/{resources.go,routes.go,graph_restore_source.go}
internal/app/configassembly/{configuration.go,deployment.go}
internal/app/server/runtime_assembly.go
internal/modules/extensions/coordinator.go
```

Verification and routing:

```text
internal/modules/networkflow/{configuration_test.go,network_flow_contract_test.go}
internal/modules/networkflow/{network_flow_unit_test.go,routes_integration_test.go,store_test.go}
internal/app/configassembly/configuration_test.go
internal/app/server/runtime_test.go
tools/test_families/module.networkflow.json
tools/execution_topology_render_index.json
```

Generated roots changed only through `make generate`; no generated file,
lockfile, migration, Graph Projection engine algorithm, or public cleanup
surface was hand-edited.

#### Compatibility and migration impact

- Configuration major 2 makes the 256-byte header default and all deployment
  overrides explicit. An invalid formerly repaired partial runtime value now
  fails closed, which is the intentional compatibility break.
- Public discovery advances to the projected v2 response and exposes all 23
  effective values. Effective deployment limits and lower request overrides do
  not enter Graph query identity.
- Existing persisted semantic-query-v1 declarations and result bytes are not
  rewritten. State `2 -> 3` is byte-preserving and creates no job or result.
- The internal store constructor now requires an explicit effective value.
  Test and restore composition choose defaults explicitly rather than relying
  on hidden fallback behavior.
- A zero-edge or zero-example request override is now accepted as required by
  the adopted minima. No legacy positive-only behavior is retained because it
  had no continuing semantic value.

#### Validation

- Required `make task-guide ROLE=module-author OWNER=...` and
  `make explain-test-owner OWNER=...` probes passed for
  `module.networkflow`, `module.extensions`, `platform.config`, and
  `app.server` before implementation.
- Network Flow configuration, effective-limit, current Graph override, and v4
  contract rows — PASS, 3/3, root
  `.cartulary/test-results/20260816T141711Z-p1705560`.
- Platform configuration claim, owner projection, and inactive-profile rows —
  PASS, 3/3, root `.cartulary/test-results/20260816T141711Z-p1705578`.
- Full `platform.config` owner slice — PASS, 15/15, root
  `.cartulary/test-results/20260816T141848Z-p1716477`.
- Extensions inactive-configuration and generated-binding admission rows —
  PASS, 2/2, root `.cartulary/test-results/20260816T141724Z-p1707415`.
- Application startup and runner-assembly rows — PASS, 2/2, root
  `.cartulary/test-results/20260816T141711Z-p1705598`.
- Configured effective-limit discovery through the service-backed application
  runtime — PASS, 3/3 units, root
  `.cartulary/test-results/20260816T141532Z-p1696967`.
- `make format` — PASS. Final pre-checkpoint root
  `.cartulary/test-results/20260816T141756Z-p1708328`.
- `make generate` — PASS, root
  `.cartulary/test-results/20260816T141654Z-p1702445`.
- `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T141804Z-p1712128`.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T141804Z-p1712137`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T141804Z-p1712192`.
- `make backend-module-boundary-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T141804Z-p1712447`.
- `git diff --check` and `git diff --cached --check` — PASS before this
  checkpoint edit.
- `make lint-markdown` — PASS, root
  `.cartulary/test-results/20260816T142250Z-p1766362`.

Early related failures were resolved without weakening the owner policy.
Formatting root `.cartulary/test-results/20260816T140540Z-p1657291` exposed a
syntax error; Network Flow root
`.cartulary/test-results/20260816T140612Z-p1664534` exposed a duplicate test
binding; platform-config root
`.cartulary/test-results/20260816T140649Z-p1666002` exposed the coordinator's
stale implementation-binding-v1 parser; Network Flow root
`.cartulary/test-results/20260816T141106Z-p1683556` exposed a boundary test that
did not keep paired active/retained limits valid; and service-backed root
`.cartulary/test-results/20260816T141418Z-p1691466` exposed missing runtime
registration for the adopted byte-preserving state-3 migration. The final
narrow roots above supersede each failure. Two attempted commands used
nonexistent row IDs and stopped at catalog admission without test execution or
a retained run root; both were immediately replaced by generated row IDs.

The broadened `module.networkflow` diagnostic at
`.cartulary/test-results/20260816T141848Z-p1716476` passed 75/79 units. Its
three product-test failures are classified predecessor projection lag outside
the S03 gate: the saved lifecycle still constructs the pre-GP3 Recovery source
registry assigned to S09, while the frontend integration and visual golden
still consume the major-3 discovery shape assigned to S10. No S03 resource
configuration, discovery, startup, request-limit, or assembly row failed in
that run. S09 and S10 must supersede this diagnostic before their respective
owner gates and before final validation.

Migration, security, telemetry, cleanup, temporal, capacity, backup/restore,
rollback, broad `make check`, and release checks were not S03 gates and remain
assigned to S06 through S12. The only S03 residual is deliberate sequencing:
the two temporal-specific lower overrides cannot execute until S08 installs
semantic-query v2; there is no alternative decoder or compatibility path.

#### Current next action

Execute GP3-S04 bounded source and contributor pipeline. Do not begin worker
admission, cleanup, telemetry, temporal execution, Recovery/Reporting, web, or
capacity work before S04 is `DONE`.

### 10.12 GP3-S04 completion checkpoint

Status: `DONE`.

GP3-S04 replaces whole-scope Graph source materialization with one canonical,
cancellation-aware database iterator. The iterator preserves caller-supplied
table order through `unnest(... WITH ORDINALITY)` and then orders rows by the
existing contributor keyset. Default aggregation now retains vertices, edges,
aggregate counts, contributing table/fingerprint sets, counter sums, and only
the configured number of example rows. Contributing-row, vertex, edge, and
counter limits stop iteration at the first disallowed row rather than after a
complete source copy.

Contributor queries now accept only canonical Network Flow source selectors.
The service recomputes `nff_...` vertex and edge identities from their complete
source key, rejects ID/key mismatches, resolves active source tables and the
query digest on every page, and applies a fixed selector predicate in SQL.
Only `limit + 1` matching rows and cursor state are retained. Cursor identity
binds the canonical selector, graph digest, actor, session, incident, order,
and page limit. The old ID-only contributor fallback and complete graph
reconstruction path were removed because neither offers continuing value in
the major-4 contract.

Indicator linking still accepts its separately owned persisted-v1 graph
selector. It uses the bounded aggregate only to validate that existing
selector identity, then scans the fixed source predicate and retains no more
than `max_binding_source_row_refs`. Explicit `row_refs` remain bounded by that
same deployment limit. Graph Projection remains an unchanged pure v2 engine.

The semantic-query-v2 default-mode digest foundation is length-framed and has
a frozen identity fixture. It is intentionally not selected by a public route
until GP3-S08 admits semantic-query v2 and its temporal variant. Request limits
remain excluded from both identities.

#### Files changed

Runtime and persistence:

```text
internal/modules/networkflow/digest.go
internal/modules/networkflow/graph.go
internal/modules/networkflow/indicator_link.go
internal/modules/networkflow/store.go
```

Verification and routing:

```text
internal/modules/networkflow/graph_streaming_test.go
internal/modules/networkflow/routes_integration_test.go
tools/test_families/module.networkflow.json
tools/execution_topology_render_index.json
```

Generated topology changed only through `make generate`. No generated root,
lockfile, migration, Graph Projection engine algorithm, cleanup surface, job
admission policy, telemetry path, or frontend file was hand-edited in S04.

#### Compatibility and migration impact

- Existing persisted semantic-query-v1 declarations continue to use the v1
  digest, source-snapshot identity, projection identity, ordering, aggregate
  meaning, and canonical output bytes. The streaming v1 aggregate is frozen by
  a golden fixture.
- The canonical contributor selector is an intentional major-4 break. The
  prior ID-only request is not retained as a fallback because it requires full
  graph reconstruction and cannot provide a durable selection key for temporal
  edges. The public schema constants advance in S08 and the browser advances
  in S10; the intermediate worktree is not a releasable compatibility point.
- Default semantic-query-v2 uses a distinct digest and therefore distinct
  downstream snapshot/result identities when S08 activates it. That boundary
  is deliberate even though default graph meaning is unchanged.
- No persisted declaration or result is rewritten, and no database migration
  is required by this slice.

#### Validation

- Required `make task-guide ROLE=module-author OWNER=...` and
  `make explain-test-owner OWNER=...` probes passed for
  `module.networkflow` and `module.graphprojection` before implementation.
- Streaming aggregate, retained-cardinality, v1 golden, v2 digest, limit+1,
  canonical-selector, and database-error tests — PASS, root
  `.cartulary/test-results/20260816T145239Z-p1863247`.
- Ordered multi-table source iteration, cancellation, contributing-row limit,
  vertex/default-edge selectors, continuation, stale digest, ID/key mismatch,
  source removal, and authorization-loss integration — PASS, 3/3, root
  `.cartulary/test-results/20260816T144828Z-p1840198`.
- Adjacent Network Flow unit and graph-view store rows — PASS, 4/4, root
  `.cartulary/test-results/20260816T144910Z-p1841798`; service-backed
  graph-view persistence — PASS, 3/3, root
  `.cartulary/test-results/20260816T144940Z-p1843309`.
- All unaffected Graph Projection engine and storage rows — PASS, 8/8, root
  `.cartulary/test-results/20260816T145038Z-p1846934`; the service-backed
  storage subset — PASS, 5/5, root
  `.cartulary/test-results/20260816T145046Z-p1848292`.
- `make format` — PASS, final root
  `.cartulary/test-results/20260816T145220Z-p1856754`.
- `make generate` — PASS, final root
  `.cartulary/test-results/20260816T145226Z-p1860252`.
- `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T145305Z-p1864245`.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T145305Z-p1864247`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T145305Z-p1864249`.
- `make backend-module-boundary-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T145305Z-p1864309`.
- `git diff --check` and `git diff --cached --check` — PASS.
- `make lint-markdown` — PASS, root
  `.cartulary/test-results/20260816T145427Z-p1868455`.

Early S04 failures were resolved without retaining compatibility shortcuts.
Unit roots `.cartulary/test-results/20260816T143655Z-p1792807` and
`.cartulary/test-results/20260816T144431Z-p1824036` recorded the deliberately
unfrozen v1 and v2 fixture digests before they were reviewed and fixed.
Service-backed root `.cartulary/test-results/20260816T144541Z-p1829457`
showed that a removed source correctly returns the more precise existing
`network_flow_table_not_active/soft_deleted` result; the assertion was
corrected and the final service-backed root supersedes it. One combined
`make format generate test-slice ...` invocation failed at task-surface
admission because `OWNER` is invalid for `format`; each public target was then
run separately and passed.

The broadened Graph Projection diagnostic at
`.cartulary/test-results/20260816T144947Z-p1844465` passed 8/9 units. Its only
failed row is the already classified GP3-S09 `source_registry_mismatch`: the
generated current Recovery source registry is v3 while runtime exact mixed
declaration dispatch remains v2. No Graph engine, Graph storage, streaming,
selector, or contributor row failed; the explicit unaffected roots above
supersede the diagnostic for S04. S09 must supersede the Recovery row before
its owner gate and final validation.

Temporal selector/result wiring, time-bucket aggregation, and public
semantic-query-v2 activation remain solely in S08. Mixed saved-declaration
Recovery remains in S09, browser consumption remains in S10, and environment
capacity/RSS evidence remains in S11. These are sequenced dependencies rather
than alternative S04 paths. Migration, cleanup, worker isolation, telemetry,
security, backup/restore, rollback, broad `make check`, and release checks were
not S04 gates and remain assigned to S05 through S12.

#### Current next action

Execute GP3-S05 worker isolation and backpressure. Do not begin cleanup,
telemetry, temporal execution, Recovery/Reporting, web, or capacity work before
S05 is `DONE`.

### 10.13 GP3-S05 completion checkpoint

Status: `DONE`.

GP3-S05 removes the handwritten job-kind-to-worker switch from application
assembly. The Extensions coordinator now parses, retains, copy-projects, and
publishes each generated `cartulary.extension_worker_runtime_contract.v1` with
its profile ID, worker kind, sorted job-kind set, and exact process maximum.
Both the complete recognized Jobs catalog and the claim-filtered runnable
selection are derived from those generated contracts. Startup rejects missing,
duplicate, unsorted, unknown, cross-profile, cross-handler, zero-capacity, and
incomplete assignments rather than inferring or defaulting them.

Jobs recovery candidates now carry job ID, job kind, and the catalog-derived
handler identity. The runner asks persistence only for job kinds whose worker
has capacity in the current process. It reserves a numbered global attempt slot
and the generated worker slot together under one lock before durable claim,
then releases both exactly once on claim miss, claim error, identity mismatch,
handler completion, failure, cancellation, ownership loss, or graceful
conditional release. Notification hints resolve the same catalog candidate and
remain best-effort; initial and periodic durable scans remain authoritative.

Recovery scanning continues after one worker becomes saturated. This is
important when an older full recovery batch contains only Graph jobs: the
runner recomputes the non-saturated job-kind set and performs a bounded follow-
up query so another worker is not hidden behind that batch. Production global
concurrency remains `8`, existing worker maxima remain `8`, and
`network_flow_activity.graph_view_worker_v1` is exactly `1` per serving
process. No handler-local semaphore or deployment configuration was added.

#### Files changed

Generated-contract admission and application projection:

```text
internal/modules/extensions/coordinator.go
internal/app/extensionassembly/{jobs.go,publication_catalog.go}
internal/app/server/runtime_assembly.go
```

Jobs runtime:

```text
internal/platform/jobs/definition_catalog.go
internal/platform/jobs/durable_persistence.go
internal/platform/jobs/runner.go
```

Verification and harness support:

```text
internal/modules/extensions/{coordinator_test.go,contract_test.go}
internal/app/extensionassembly/publication_catalog_test.go
internal/app/server/extensions_publication_characterization_test.go
internal/platform/jobs/{composition_test.go,jobs_test.go,runner_supervision_test.go}
internal/testutil/collaborationsupport/intents.go
tools/test_families/platform.jobs.json
tools/execution_topology_render_index.json
```

The generated worker contracts themselves were adopted and produced in S01
and S02. S05 consumes them without editing generated roots, lockfiles,
migrations, product schemas, public job state, or Network Flow handlers.

#### Compatibility and migration impact

- Public and durable job resource shapes, statuses, retry accounting, leases,
  progress, cancellation, and terminal results are unchanged.
- Existing workers retain effective process capacity `8`; only the Graph-view
  worker receives its adopted isolation maximum `1`. The global maximum stays
  `8`, so non-Graph throughput is not reduced by a waiting Graph handler.
- Candidate discovery is a private persistence/runner contract change. Stored
  handler identity is still validated against the immutable recognized catalog
  before claim, and no job row is rewritten.
- Test composition must now provide explicit worker runtime contracts. This is
  an intentional fail-closed constructor change; hidden test or production
  defaults would conceal incomplete packaging.

#### Validation

- Required `make task-guide ROLE=module-author OWNER=...` and
  `make explain-test-owner OWNER=...` probes passed for `platform.jobs`,
  `module.extensions`, `module.networkflow`, and `app.server` before
  implementation.
- Complete Jobs owner slice — PASS, 18/18, root
  `.cartulary/test-results/20260816T151357Z-p1954095`.
- Complete service-backed Jobs owner slice — PASS, 14/14, root
  `.cartulary/test-results/20260816T151435Z-p1956975`.
- The focused supervised-runner row proves exact Graph capacity one, existing
  global concurrency, mixed-kind progress behind 101 older Graph jobs,
  notification recovery after slot release, retry/renewal, claim/recovery,
  shutdown, and fatal-loss invariants — PASS, 3/3, root
  `.cartulary/test-results/20260816T151134Z-p1941740`.
- Extensions registry, worker publication, immutability, Network Flow v4, and
  generated-capacity rows — PASS, 3/3, root
  `.cartulary/test-results/20260816T151747Z-p1973343`.
- Application publication and exact policy catalog rows — PASS, 2/2, root
  `.cartulary/test-results/20260816T151346Z-p1953505`.
- Claimed Network Flow server composition and contributor workflow — PASS,
  3/3, root `.cartulary/test-results/20260816T150830Z-p1930299`.
- Unclaimed optional-profile server composition remains quiescent — PASS,
  3/3, root `.cartulary/test-results/20260816T151602Z-p1963483`.
- Inactive extension job reconciliation with the explicit worker selection —
  PASS, 3/3, root `.cartulary/test-results/20260816T151450Z-p1958162`.
- `make format` — PASS, final root
  `.cartulary/test-results/20260816T151738Z-p1969815`.
- `make generate` — PASS, root
  `.cartulary/test-results/20260816T150439Z-p1887403`.
- `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T151810Z-p1974129`.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T151810Z-p1974131`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T151810Z-p1974133`.
- `make backend-module-boundary-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T151810Z-p1974193`.
- `git diff --check` and `git diff --cached --check` passed before this
  checkpoint edit.
- Final `git diff --check` and `git diff --cached --check` — PASS.
- Final `make lint-markdown` — PASS, root
  `.cartulary/test-results/20260816T152010Z-p1978352`.

Focused service-backed root
`.cartulary/test-results/20260816T150453Z-p1890383` exposed that the first
runner revision skipped the synchronous storage probe while the dequeue gate
was closed. Recovery selection now validates storage regardless of gate state;
claim scheduling remains gated, and the final runner root supersedes the
failure. Root `.cartulary/test-results/20260816T150958Z-p1936149` exposed a
test notification sent in the small interval after the handler became inactive
but before its runner reservation was released. The final test retries the
best-effort hint while durable recovery remains authoritative and proves the
successor is admitted; no wait or semaphore was added to product handlers.

The broadened Extensions diagnostic at
`.cartulary/test-results/20260816T150716Z-p1902556` passed 35/40 units. Its
stale major-3 and two-entry inactive-configuration expectations were corrected
and now pass in the explicit Extensions root above. Its remaining failures are
strictly sequenced: generic state-runtime fixtures do not yet install Network
Flow state-3 final validation and concurrent migration behavior assigned to
S09, while the stateful browser still consumes major 3 and is assigned to S10.
No worker contract, Jobs claim, runner, server composition, or mixed-kind row
failed in the final S05 evidence.

Jobs queued/queue-wait telemetry remains in S07. Cleanup, temporal execution,
mixed Recovery/Reporting, browser completion, capacity certification,
backup/restore, rollback, broad `make check`, and release checks were not S05
gates and remain assigned to S06 through S12. No alternative worker mapping or
compatibility path remains.

#### Current next action

Execute GP3-S06 durable cleanup. Begin with the required owner routing, then
add only migration `00034`, replace global reachability cleanup with the
borrowed-transaction source-owner protocol, repair the shared lock order, and
compose the private bounded Network Flow dispatcher. Do not begin telemetry,
temporal execution, Recovery/Reporting, web, or capacity work before S06 is
`DONE`.

### 10.14 GP3-S06 completion checkpoint

Status: `DONE`.

GP3-S06 replaces the unbounded global reachability-cleanup seam with a narrow
borrowed-transaction maintenance contract. Graph Projection now deletes at
most 1,000 expired leases in one transaction, locks one oldest candidate for
one exact source owner using `(published_at, projection_result_id)` and
`FOR UPDATE SKIP LOCKED`, permits a live-lease recheck, and deletes at most the
locked result envelope plus its owned cascades. A private continuation key
allows one source owner to traverse selected results without turning an
unbounded reachable-ID set into an application input.

Network Flow owns the policy and orchestration. One sweep captures a single
observation time, drains one bounded lease batch, and processes at most eight
result transactions or 30 seconds. Each result transaction locks the result
first, locks and checks all Network Flow declarations selecting it, rechecks
live leases, and deletes only when both authoritative reachability tests are
empty. The graph-view publication store and Reporting lease admission now
enforce the same result-before-declaration order. Publication replay also
handles the legal race in which a conflicting result vanishes while an
`ON CONFLICT DO NOTHING` insertion waits: it performs one bounded reinsertion
only after an exact locked read proves absence, without weakening digest,
binding, or byte-equivalence checks.

The private dispatcher is constructed only for the claimed Network Flow
profile, starts after serving readiness, runs an immediate sweep, uses the
adopted five-minute cadence, paces continuation at five seconds while work
reports `has_more`, retries transient failures after a bounded 30 seconds,
coalesces ticks through one serialized loop, and cancels in-flight work during
reverse-order shutdown. Restart is supported. A panic or other unexpected loop
loss reaches the fatal component-loss sink, and its completion channel closes
even if that sink panics. No public cleanup route, operator command, common job,
retention interval, or Graph worker was added.

#### Files changed

Migration and generated database projections:

```text
db/migrations/00034_graph_projection_cleanup_indexes.sql
tools/database-migrations/generate-catalog-projections.mjs
tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership.mjs
tools/migration_history_manifest.json
tools/schema_object_ownership_manifest.json
```

Graph Projection maintenance boundary and PostgreSQL implementation:

```text
internal/modules/graphprojection/result_v2_ports.go
internal/modules/graphprojection/postgresresult/store.go
internal/modules/graphprojection/postgresresult/store_test.go
internal/modules/graphprojection/postgresresult/cleanup_test.go
internal/modules/graphprojection/graph_projection_migration_test.go
```

Network Flow cleanup policy, dispatcher, lock ordering, and verification:

```text
internal/modules/networkflow/graph_result_cleanup.go
internal/modules/networkflow/graph_result_cleanup_dispatcher.go
internal/modules/networkflow/graph_result_cleanup_dispatcher_test.go
internal/modules/networkflow/graph_result_cleanup_integration_test.go
internal/modules/networkflow/graph_result_cleanup_race_integration_test.go
internal/modules/networkflow/graph_result_cleanup_test_bridge_test.go
internal/modules/networkflow/graph_view_store.go
internal/modules/networkflow/reporting_graph_source.go
```

Application lifecycle and routed verification:

```text
internal/app/server/runtime_assembly.go
internal/app/server/evidence_cleanup_telemetry_test.go
internal/app/server/evidence_composition_characterization_test.go
internal/app/server/runtime_integration_test.go
tools/test_families/module.graphprojection.json
tools/test_families/module.networkflow.json
tools/test_families/app.server.json
tools/execution_topology_render_index.json
```

The topology and database manifests changed only through `make generate` after
their authored catalogs were updated. Generated roots and lockfiles were not
hand-edited. Migrations `00032_graph_projection_v2.sql` and
`00033_graph_projection_v1_removal.sql` are byte-unchanged.

#### Compatibility and migration impact

- Migration `00034_graph_projection_cleanup_indexes.sql` is additive and adds
  only the owner/order candidate index. Its rollback drops only that index.
- Persisted query declarations, result identities, canonical result bytes,
  lease expiry meaning, publication selection, and exact v1 execution remain
  unchanged. Cleanup affects only expired leases and unselected, unleased
  derived results owned by `network_flow_activity`.
- The unused global `ReachabilityCleanerV2` signature is intentionally removed.
  Retaining it would preserve a cross-owner, unbounded, race-prone API with no
  continuing value. Future source owners can use the same narrow Graph
  maintenance primitives while owning their own authoritative selection test.
- Publication's bounded vanished-conflict retry is an internal race repair, not
  an identity or overwrite compatibility path. Any surviving row with unequal
  binding, digest, or bytes remains an invariant conflict.
- Application composition gains a private claimed-profile lifecycle component.
  Unclaimed deployments remain quiescent, and there is no externally callable
  cleanup behavior to migrate.

#### Validation

- Required `make task-guide ROLE=module-author OWNER=...` and
  `make explain-test-owner OWNER=...` probes passed for
  `module.graphprojection`, `module.networkflow`, `module.reporting`, and
  `app.server` before implementation.
- Complete focused Graph Projection cleanup and storage slice — PASS, 6/6,
  root `.cartulary/test-results/20260816T155711Z-p2110516`; corresponding
  service-backed slice — PASS, 5/5, root
  `.cartulary/test-results/20260816T155835Z-p2113720`.
- Network Flow cleanup, bounded sweep, lifecycle, and adjacent unit rows —
  PASS, 7/7, final root
  `.cartulary/test-results/20260816T155911Z-p2116293`; corresponding
  service-backed rows — PASS, 5/5, root
  `.cartulary/test-results/20260816T155844Z-p2114899`.
- The dispatcher lifecycle row after completion-channel hardening — PASS, 1/1,
  root `.cartulary/test-results/20260816T160154Z-p2128502`.
- Reporting exact-result lease admission and lock-order rows — PASS, 3/3, root
  `.cartulary/test-results/20260816T155919Z-p2117461`; service-backed Reporting
  rows — PASS, 3/3, root
  `.cartulary/test-results/20260816T155210Z-p2056589`.
- Application readiness and reverse-shutdown lifecycle rows — PASS, 5/5, root
  `.cartulary/test-results/20260816T155949Z-p2118861`; unclaimed-profile
  service-backed composition — PASS, 3/3, root
  `.cartulary/test-results/20260816T155327Z-p2063773`; claimed Network Flow
  actual-server composition — PASS, 3/3, root
  `.cartulary/test-results/20260816T155757Z-p2112284`.
- The race-focused source-owner cleanup row — PASS, root
  `.cartulary/test-results/20260816T154745Z-p2043614`; the bounded cleanup
  integration row — PASS, root
  `.cartulary/test-results/20260816T154540Z-p2034630`.
- `make format` — PASS, final root
  `.cartulary/test-results/20260816T160150Z-p2125013`.
- `make generate` — PASS, final root
  `.cartulary/test-results/20260816T155514Z-p2100558`.
- `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T160206Z-p2129477`.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T160206Z-p2129503`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T160206Z-p2129527`.
- `make migration-drift` — PASS, 5/5, root
  `.cartulary/test-results/20260816T160206Z-p2129581`.
- `make backend-module-boundary-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T160206Z-p2129964`.
- `make go-gosec-targeted` — PASS, 4/4, root
  `.cartulary/test-results/20260816T160206Z-p2130164`.
- `git diff --check` and `git diff --cached --check` passed before this
  checkpoint edit.
- Mandatory checkpoint `git diff --check` and `git diff --cached --check` —
  PASS.
- Mandatory checkpoint `make lint-markdown` — PASS, root
  `.cartulary/test-results/20260816T160332Z-p2163635`.

Focused roots `.cartulary/test-results/20260816T153015Z-p1989818`,
`.cartulary/test-results/20260816T153015Z-p1989823`, and
`.cartulary/test-results/20260816T153015Z-p1989830` exposed a removed import and
stale interface assertion during the cleanup-port replacement; both were
corrected before behavioral validation. Race root
`.cartulary/test-results/20260816T154300Z-p2023372` exposed the legal vanished
`ON CONFLICT` row and led to the bounded exact-read/reinsert repair. Reporting
root `.cartulary/test-results/20260816T154617Z-p2036563` exposed a stale-source
classification-order regression; the result-first locking path now preserves
the adopted stale-source-before-digest result.

The first generation attempt at
`.cartulary/test-results/20260816T153744Z-p2005906`, with diagnostic root
`.cartulary/test-results/20260816T153827Z-p2007804`, correctly rejected
migration 34 without a catalog owner. The authored migration owner was added.
JSON root `.cartulary/test-results/20260816T155421Z-p2099099` then exposed the
drift validator's incomplete owner map; after that authored validator was
updated, root `.cartulary/test-results/20260816T155503Z-p2100042` reported only
the expected stale generated topology. The final `make generate` and passing
JSON-shape root supersede all three failures.

The tests cover exact 1,000-plus-one lease batches, active-lease exclusion,
source-owner isolation, oldest ordering, cursor continuation, concurrent
`SKIP LOCKED` selection, selected-result preservation, large cascades, rollback,
eight-result and elapsed-time sweep bounds, cancellation, eventual drain,
publication and Reporting lease races, immediate/cadenced/continued/retried
dispatch, non-overlap, restart, shutdown cancellation, and fatal loop loss.

S07 telemetry is deliberately absent from these new operations until its
registry and privacy workstream. S08 temporal results, S09 mixed Recovery and
Reporting label semantics, S10 browser behavior, S11 retained capacity and
rollout drills, and S12 broad `make check`, release, full security, backup,
restore, and rollback gates were not S06 gates. They remain sequenced work,
not unclassified cleanup risk.

#### Current next action

Execute GP3-S07 operational telemetry. Begin with the required owner routing,
then project Jobs queued/queue-wait and Network Flow materialization and cleanup
signals through privacy-safe observers. Do not begin temporal execution,
Recovery/Reporting, web, or capacity work before S07 is `DONE`.

### 10.15 GP3-S07 completion checkpoint

Status: `DONE`.

GP3-S07 adds the missing operational boundaries without coupling the pure
Graph Projection engine to telemetry. Jobs now exposes current queued work by
the immutable catalog-backed job kind and records queue wait exactly once,
after a durable claim commits, from
`COALESCE(handler_next_attempt_at, submitted_at)` to the claim instant. Claim
misses and errors emit no wait. An impossible future eligibility timestamp is
an invariant violation rather than a clamped measurement. The queued gauge
includes retry-delayed work and emits the complete closed catalog plus
`unknown`, so absent kinds remain observable as zero without exposing raw
stored values.

Network Flow now owns a small semantic observer port containing only closed
operation, phase, graph mode, outcome, safe error class, aggregate counts, and
durations. Ephemeral and saved graph composition report source validation,
source scan, Graph Projection, and final publication at the boundaries where
those outcomes become known. Successful compositions report contributing-row,
vertex, and edge counts; S08 will populate the already-projected temporal
bucket count. Cleanup reports operation outcome, duration, expired-lease and
projection-result deletions, and a post-sweep health snapshot derived from the
same exact source-owner selection-and-lease predicate used for deletion. The
oldest age is measured from `published_at`; no signal calls it time
unreachable.

Application assembly adapts the semantic port to the
`cartulary.network_flow` OpenTelemetry scope. The adapter owns instruments,
spans, static safe logs, attribute filtering, and last-valid cleanup gauges.
It is composed only for the claimed profile. Network Flow invokes the port
through panic-contained wrappers, and instrument-construction, no-SDK,
observer, collection, logger, or exporter failure has no product return path.
No identifier, source-owner token, cursor, digest, endpoint, label, row,
property, SQL, database object name, or raw error is representable in the
semantic observations.

The S02 OpenTelemetry registry projection had used draft signal names and
vocabularies that disagreed with the adopted OpenTelemetry owner. This slice
repairs that downstream artifact and its conformance expectations to the exact
owner names, units, instrument kinds, attributes, scopes, and closed values.
`OTEL-CORPUS-019` now binds the Jobs queue and Network Flow graph/cleanup
requirements, including disabled-export and product-invariance assertions.

#### Files changed

Jobs queue telemetry and verification:

```text
internal/platform/jobs/{manager.go,telemetry.go,durable_persistence.go}
internal/platform/jobs/{telemetry_test.go,telemetry_integration_test.go}
tools/test_families/platform.jobs.json
```

Network Flow semantic boundaries, cleanup health, and verification:

```text
internal/modules/networkflow/{graph.go,graph_view_jobs.go,module.go,routes.go}
internal/modules/networkflow/{graph_telemetry.go,graph_telemetry_test.go}
internal/modules/networkflow/{graph_result_cleanup.go,graph_result_cleanup_dispatcher.go}
internal/modules/networkflow/graph_result_cleanup_integration_test.go
tools/test_families/module.networkflow.json
```

Application and platform OpenTelemetry projection:

```text
internal/app/server/{runtime_assembly.go,network_flow_telemetry.go,network_flow_telemetry_test.go}
internal/platform/telemetry/{accessors.go,privacy.go,registry.go}
internal/platform/telemetry/{accessors_test.go,privacy_test.go,registry_test.go}
internal/platform/telemetry/testsupport/capture.go
contracts/otel/cartulary_signal_registry.v1.json
tools/otel/check-otel-conformance.mjs
internal/testutil/golden/otel/corpus_manifest.json
internal/testutil/golden/otel/cases/OTEL-CORPUS-009/input.json
internal/testutil/golden/otel/cases/OTEL-CORPUS-015/input.json
internal/testutil/golden/otel/cases/OTEL-CORPUS-019/**
tools/test_families/app.server.json
tools/execution_topology_render_index.json
```

The topology projection changed only through `make generate`. No generated
Go or TypeScript product contract, migration, lockfile, Graph Projection
engine code, public response, or persistence identity changed in S07.

#### Compatibility and migration impact

- The changes are additive internal telemetry. Public APIs, job state,
  retries, selected graph results, cleanup eligibility, declaration bytes, and
  Graph Projection v2 behavior are unchanged.
- Existing worker kinds retain their closed job-kind identities. The new
  queued and wait measurements consume the generated runtime catalog and add
  no hardcoded compatibility map.
- Deployments with telemetry disabled or no SDK retain identical product
  behavior. Telemetry consumers must adopt the owner-exact signal names and
  `canceled` spelling; the discarded draft names never formed an adopted or
  released compatibility surface.
- Cleanup health is a current observation, not persisted state. A failed or
  invalid health read leaves the prior valid gauges intact and cannot fail a
  sweep.

#### Validation

- Required `make task-guide ROLE=module-author OWNER=...` and
  `make explain-test-owner OWNER=...` probes passed for `platform.jobs`,
  `module.networkflow`, `platform.telemetry`, and `app.server` before
  implementation.
- Complete Jobs unit slice — PASS, 18/18, root
  `.cartulary/test-results/20260816T162731Z-p2221219`; complete Jobs
  service-backed slice — PASS, 14/14, root
  `.cartulary/test-results/20260816T162759Z-p2226169`. The final queue-focused
  unit and service rows pass at
  `.cartulary/test-results/20260816T163005Z-p2234789` and
  `.cartulary/test-results/20260816T163014Z-p2241301`.
- Complete platform telemetry slice — PASS, 10/10, root
  `.cartulary/test-results/20260816T162731Z-p2221243`.
- Network Flow aggregation, cleanup lifecycle, and telemetry unit rows — PASS,
  3/3, root `.cartulary/test-results/20260816T162731Z-p2221224`; the expanded
  semantic telemetry boundary passes at
  `.cartulary/test-results/20260816T163005Z-p2234799`.
- Network Flow bounded contributor, cleanup, cleanup-race, and graph-view
  persistence service rows — PASS, 6/6, root
  `.cartulary/test-results/20260816T163221Z-p2312404`.
- Complete application server unit slice — PASS, 45/45, root
  `.cartulary/test-results/20260816T163049Z-p2245330`; the telemetry and cleanup
  lifecycle row also passes independently at
  `.cartulary/test-results/20260816T163005Z-p2234808`.
- `make otel-conformance` — PASS, 14/14, root
  `.cartulary/test-results/20260816T163005Z-p2234635`.
- `make format` — PASS, root
  `.cartulary/test-results/20260816T162947Z-p2228262`.
- `make generate` — PASS, root
  `.cartulary/test-results/20260816T162950Z-p2231714`.
- `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T163049Z-p2245221`.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T163049Z-p2245262`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T163050Z-p2246608`.
- `make backend-module-boundary-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T163049Z-p2245493`.
- `make go-gosec-targeted` — PASS, 4/4, root
  `.cartulary/test-results/20260816T163051Z-p2247452`.
- `git diff --check` and `git diff --cached --check` passed before this
  checkpoint edit.
- Mandatory checkpoint `git diff --check` and `git diff --cached --check` —
  PASS.
- Mandatory checkpoint `make lint-markdown` — PASS, root
  `.cartulary/test-results/20260816T163509Z-p2314648`.

The first OpenTelemetry conformance run at
`.cartulary/test-results/20260816T162323Z-p2200159` and its immediate rerun at
`.cartulary/test-results/20260816T162510Z-p2205723` rejected only non-canonical
serialization/key order in the newly authored empty normalized corpus files.
Canonical one-line serialization corrected the authored fixtures; the final
14/14 root above supersedes both failures.

A combined Network Flow service run at
`.cartulary/test-results/20260816T162731Z-p2221234` passed all six S07/S04/S06
rows and failed only
`module.networkflow.integration.saved_graph_lifecycle_v2` while constructing
the not-yet-implemented current Recovery-v3 source registry. This is the same
explicit S02-to-S09 projection lag recorded in §10.10, not an S07 telemetry
gate. S09 must supersede it before any restore, broad, or release gate; the
passing six-row root above is the final S07 service evidence.

S08 temporal execution, S09 Recovery/Reporting, S10 browser behavior, S11
capacity and rollout drills, and S12 broad `make check`, release, backup,
restore, and rollback gates were skipped because their strictly ordered owning
slices have not begun. No S07 owner, security, privacy, migration, or telemetry
gate failed. The only residual telemetry work is S08 supplying exact temporal
bucket cardinality and timeout outcomes through the already-closed observer
boundary when those semantics exist.

#### Current next action

Execute GP3-S08 time-bucket backend. Begin with required Network Flow and Graph
Projection owner routing, then implement v2 admission, identity, arithmetic,
streaming aggregation, response-owned bucket summaries, saved materialization,
contributors, limits, and errors. Do not begin Recovery/Reporting, web, or
capacity work before S08 is `DONE`.

### 10.16 GP3-S08 completion checkpoint

Status: `DONE`.

GP3-S08 implements time-bucketed graphs as Network Flow semantics while keeping
Graph Projection v2 unchanged and policy-free. Current public graph writes now
accept only the closed semantic-query-v2 union. Its default variant preserves
the existing overlap meaning behind the intentional graph-query-digest-v2
identity boundary; its temporal variant requires a complete half-open range and
one of the six adopted widths. Persisted v1 declarations retain their exact v1
decoder, digest, source snapshot, result bytes, refresh, contributor, and
rebuild behavior. They cannot be created through the current public API.

Temporal arithmetic uses Unix-epoch-aligned UTC buckets, mathematical floor
division before the epoch, exact-boundary exclusive ends, and one
`flow_start_utc` assignment per accepted row. Network Flow streams rows into
incremental aggregates, preserves bounded examples, enforces contributing-row,
bucket, vertex, edge, example, and counter-digit limits, and gives temporal
edges `nfbe_` identities that bind both bucket bounds and the complete default
edge key. It sends Graph Projection the adopted relationship kind, projection
version, and canonical bucket properties; the engine remains the unchanged
`graph_projection.v2` implementation.

V2 responses expose canonical Network Flow vertex, default-edge, and temporal
edge selectors. The server recomputes every source ID, applies fixed
selector-aware SQL, and binds the complete selector, digest, authorization
scope, order, and page limit into contributor cursors. Saved-result reads derive
`time_buckets[]` from the exact stored projection result and declaration rather
than persisting a second result index. Empty buckets remain present, while
bucket vertex, edge, and contributing-row counts are reconstructed exactly.
Saved contributor continuation additionally binds the graph view, selected
result, actor, session, and incident, so a refresh or authorization change
fails closed.

The materialization deadline now starts when the saved-graph handler begins and
covers validation, scanning, aggregation, and projection. Expiry before final
publication produces a terminal canonical timeout without changing the prior
selected result. Once final transactional publication begins before the
deadline, the deadline is detached and the existing indeterminate-commit
reconciliation remains authoritative.

#### Files changed

Network Flow execution and persistence adapters:

```text
internal/modules/networkflow/{graph.go,graph_temporal.go,graph_response_v2.go}
internal/modules/networkflow/{store.go,indicator_link.go,graph_view_jobs.go}
internal/modules/networkflow/{graph_view_routes.go,graph_telemetry.go}
```

Temporal, saved lifecycle, contributor, cancellation, and timeout evidence:

```text
internal/modules/networkflow/{graph_temporal_test.go,graph_materialization_timeout_test.go}
internal/modules/networkflow/{routes_integration_test.go,graph_view_store_test.go}
internal/modules/networkflow/{network_flow_unit_test.go,network_flow_contract_test.go}
internal/modules/networkflow/v3_contract_projection_test.go
tools/test_families/module.networkflow.json
```

Contract closure and generated projections:

```text
contracts/network-flow/{schemas.v2.json,routes.v1.json}
tools/harness/generated-artifacts/check-json-shapes.mjs
tools/execution_topology_render_index.json
internal/gen/contractnetworkflow/artifacts_gen.go
packages/protocol-ts/src/generated/{network-flow-types.ts,network-flow-validators.ts}
```

Generated files changed only through `make generate`. No Graph Projection
engine algorithm, migration, lockfile, cleanup schedule, public cleanup route,
or persisted bucket table was introduced by S08.

#### Compatibility and migration impact

- Current public graph query, saved-create, result, and contributor payloads
  move to their adopted major-4 schema variants. Unchanged rename, refresh, and
  retire request shapes retain their existing schema identities.
- New default and temporal writes use digest v2 and intentionally receive new
  source-snapshot and result identities. Default graph meaning is preserved,
  but v1 and v2 result identity is deliberately not conflated.
- Installed v1 declarations remain byte-exact executable state. The public v2
  decoder rejects new v1 writes, avoiding a permanent dual-write surface.
- Contributor clients must echo the canonical Network Flow selector returned
  by the result. Graph Projection `ed_` values are never accepted as source
  selectors; default and temporal source edge IDs are `nff_` and `nfbe_`.
- Request overrides may only lower deployment limits and remain outside graph
  identity. Temporal bucket width and normalized range remain inside identity.
- No schema migration is required. Network Flow extension-state admission and
  mixed-generation Recovery remain S09 work, so broad restore and release
  compatibility are not yet claimed.

#### Validation

- Required `make task-guide ROLE=module-author OWNER=...` and
  `make explain-test-owner OWNER=...` probes passed for
  `module.networkflow` and `module.graphprojection` before implementation.
- Focused Network Flow temporal arithmetic, streaming, canonical selector,
  effective-limit, and telemetry unit rows — PASS, 5/5, root
  `.cartulary/test-results/20260816T172918Z-p2486547`.
- Bounded contributor, time-bucket saved lifecycle, and graph-view declaration
  persistence service rows — PASS, 5/5 execution units, root
  `.cartulary/test-results/20260816T172927Z-p2488738`.
- Graph Projection v2 maximum-bound, pure-determinism, and contract-projection
  rows — PASS, 3/3, root
  `.cartulary/test-results/20260816T173020Z-p2491030`.
- Earlier focused temporal and saved-lifecycle roots passed at
  `.cartulary/test-results/20260816T172622Z-p2477116`,
  `.cartulary/test-results/20260816T172627Z-p2477826`, and
  `.cartulary/test-results/20260816T172318Z-p2463191` after the relevant
  arithmetic, exact-result reconstruction, cursor, cancellation, and limit
  changes.
- `make format` — PASS, root
  `.cartulary/test-results/20260816T172857Z-p2480043`.
- `make generate` — PASS, root
  `.cartulary/test-results/20260816T172903Z-p2483571`.
- `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T173037Z-p2491666`.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T173051Z-p2494608`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T173055Z-p2495059`.
- `make frontend-import-boundary-check` — PASS, 2/2, root
  `.cartulary/test-results/20260816T173101Z-p2495583`.
- `make backend-module-boundary-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T173120Z-p2523105`.
- `make go-gosec-targeted` — PASS, 4/4, root
  `.cartulary/test-results/20260816T173104Z-p2496010`.
- Pre-checkpoint `git diff --check` and `git diff --cached --check` — PASS.
- Pre-checkpoint `make lint-markdown` — PASS, root
  `.cartulary/test-results/20260816T173137Z-p2523664`.

The first temporal fixture attempts failed at
`.cartulary/test-results/20260816T165626Z-p2334878` because the authored harness
selector does not admit active `Fuzz*` execution, at
`.cartulary/test-results/20260816T165729Z-p2344902` because a test used the
wrong actor field, and at
`.cartulary/test-results/20260816T165815Z-p2349666` because the deliberately
empty edge-ID golden had not yet been captured. The routed deterministic
property-seed test plus the compiled fuzz target, corrected field, and fixed
`nfbe_` fixture supersede those failures.

The initial bounded integration run at
`.cartulary/test-results/20260816T170105Z-p2356035` found one old public v1 test
request; the major-4 request correction passes in the final focused roots. The
generation/shape attempts at
`.cartulary/test-results/20260816T172359Z-p2467907` and
`.cartulary/test-results/20260816T172429Z-p2468623` exposed a stale continuation
expectation and then its expected generated topology drift; the final
generation, drift, and JSON-shape roots supersede both. A final Graph Projection
probe first used three nonexistent row aliases and failed as an artifact-routing
error; `make explain-test-owner` identified the owned row IDs and the final 3/3
root above supersedes that invocation.

The broad Network Flow attempt at
`.cartulary/test-results/20260816T171143Z-p2394741` passed 75/87 rows. Its
backend contract failures were corrected and pass in the final focused roots;
the remaining failures are the explicitly sequenced S09 Recovery registration
and S10 browser/client work, not S08 gates. The existing saved-graph lifecycle
also reaches only the same S09 `source_registry_mismatch` gate at
`.cartulary/test-results/20260816T170548Z-p2375539`.

Active fuzz execution, mixed-generation restore, Reporting redaction, browser,
capacity, rollout/rollback, broad `make check`, and release verification were
not run because the harness exposes deterministic property evidence for this
slice and the other gates belong to S09 through S12. There is no unresolved S08
owner, implementation, security, privacy, identity, or persistence failure.

Post-checkpoint `git diff --check` and `git diff --cached --check` — PASS.
Post-checkpoint `make lint-markdown` — PASS, root
`.cartulary/test-results/20260816T173249Z-p2524866`.

#### Current next action

Execute GP3-S09 Recovery and Reporting. Begin with the required Recovery,
Reporting, Network Flow, Graph Projection, and application owner routing; then
admit mixed v1/v2 declarations through state 3, publish exact Recovery v3 while
retaining read-only historical v2 dispatch, and route temporal labels only
through typed Reporting redaction. Do not begin web, capacity, or final handoff
work before S09 is `DONE`.

### 10.17 GP3-S09 completion checkpoint

Status: `DONE`.

GP3-S09 advances Network Flow extension state from 2 to 3 with a byte-preserving
admission step. The owner validator now scans saved declarations in stable ID
order, admits only canonical semantic-query v1 or v2 bytes, recomputes their
digests, and checks that any selected projection version agrees with the saved
aggregation mode. The generic Extensions coordinator invokes that validation
through a narrow read-only family capability and remains unaware of Network
Flow tables or graph semantics.

Recovery now publishes `graphprojection.restore_rebuild.v3`, a v3 source
registry, and a v3 implementation binding that admit mixed v1/v2 declarations.
Fresh backup artifacts carry the v3 contract. The exact v2 registry, binding,
algorithm, result schema, and artifact decoder remain available only as a
read-only historical dispatcher for supported retained backup sets; no v2
artifact is translated or widened. Restore completion records the catalog
digest from the exact dispatched graph artifact, so current and historical
evidence reconcile against their own immutable generation before readiness.

Reporting retains `source_projection_ref.v2` and now receives a typed,
owner-neutral label-candidate contract from Network Flow. The source validates
and renews the exact selected result lease before deriving components solely
from exact result object state. Internal rendering applies retained redaction
components to endpoint labels and to default or canonical temporal edge labels.
Removed or missing components use deterministic ordinals; external Network
Flow graph release remains fail-closed. No identifier, digest, property bag, or
raw source value bypasses Reporting redaction.

#### Files changed

Recovery contracts, generation, dispatch, and evidence:

```text
contracts/recovery/{index.json,operator-recovery-journal-payload.v4.schema.json}
contracts/recovery/fixtures/{recovery-state-catalog.v1.json,graph-projection-restore-implementation-binding.v3.json}
contracts/recovery/fixtures/{graph-projection-restore-rebuild-result.v3.json,graph-projection-restore-source-registry.v3.json}
contracts/recovery/fixtures/operator-recovery-journal-payload.v4.json
internal/gen/contractrecovery/{artifacts_gen.go,graph_projection_restore_binding_gen.go}
internal/modules/graphprojection/{restore_contract.go,restore_service.go,recovery_state.go}
internal/modules/recovery/{restore.go,vnext_codec.go}
internal/modules/recovery/application/evidence.go
internal/modules/recovery/restorecontract/graphprojection.go
internal/app/recoveryassembly/{graphprojection_restore.go,state_catalog_test.go}
tools/contractgen/{main.go,recovery_validation.go}
```

State-3 and mixed-declaration admission:

```text
internal/platform/extensionstore/store.go
internal/modules/extensions/{state.go,state_service_test.go,state_test_adapter_test.go}
internal/app/extensionassembly/state_store.go
internal/modules/networkflow/{extension_state.go,extension_state_v3_integration_test.go}
internal/modules/networkflow/{graph_restore_source.go,graph_view_store_test.go}
tools/test_families/{module.extensions.json,module.networkflow.json}
```

Reporting typed candidates and redacted rendering:

```text
contracts/network-flow/reporting-graph-source.v2.schema.json
internal/modules/reporting/graphsourcecontract/contract.go
internal/modules/networkflow/{reporting_graph_source.go,reporting_graph_source_integration_test.go}
internal/modules/reporting/{graph_source.go,render_bundle.go,graph_result_render_test.go}
tools/test_families/{module.graphprojection.json,module.recovery.json}
```

Generated files changed only through `make generate`. S09 adds no database
migration, public route, Graph Projection engine policy, persisted bucket index,
external Reporting allow rule, or compatibility write path.

#### Compatibility and migration impact

- State `2 -> 3` is an admission migration: it validates the expanded persisted
  query union and does not rewrite declaration JSON, digests, selected result
  bindings, or result envelopes.
- Persisted v1 and v2 declarations rebuild through the current v3 dispatcher.
  Exact pre-GP3 v2 backup artifacts remain readable only through their frozen
  v2 catalog/binding/algorithm combination and continue to emit the v2 closed
  error vocabulary.
- Fresh backups and restore journals move to the v3 graph rebuild and v4 journal
  contracts. A catalog and codec from different generations cannot be mixed.
- Internal report diagrams gain deterministic redacted endpoint and temporal
  labels. External graph release remains unavailable, so S09 introduces no new
  disclosure compatibility surface.
- Historical v2 removal remains conditioned on zero supported retained backup
  sets. Persisted v1 removal separately requires zero installed v1 declarations
  and a new owner decision.

#### Validation

- Required `make task-guide ROLE=module-author OWNER=...` and
  `make explain-test-owner OWNER=...` probes passed for `module.recovery`,
  `module.reporting`, `module.networkflow`, `module.graphprojection`, and
  `app.server` before implementation.
- Graph Projection mixed v1/v2 rebuild and exact historical-v2 dispatch row —
  PASS, 1/1, root
  `.cartulary/test-results/20260816T181444Z-p2625125`.
- Recovery v3 projection, historical artifact dispatch, exact state catalog,
  and frozen-catalog rows — PASS, 4/4, root
  `.cartulary/test-results/20260816T182131Z-p2649466`; the focused v3/historical
  pair also passed 2/2 at
  `.cartulary/test-results/20260816T181444Z-p2625130`.
- Extensions state `2 -> 3` admission and final validation — PASS, 3/3, root
  `.cartulary/test-results/20260816T181815Z-p2639900`.
- Network Flow mixed v1/v2 saved-declaration validation and byte-preservation —
  PASS, 3/3, root
  `.cartulary/test-results/20260816T182040Z-p2647738`.
- Reporting exact selection, typed candidate validation, internal redaction,
  temporal rendering, ordinal fallback, and external fail-closed unit row —
  PASS, 1/1, root
  `.cartulary/test-results/20260816T182131Z-p2649461`.
- Reporting exact-result lease lifecycle — PASS, 3/3, root
  `.cartulary/test-results/20260816T182139Z-p2650824`.
- Network Flow default and temporal saved-graph lifecycle, including mixed
  restore and exact result/object reconciliation — PASS, 4/4, root
  `.cartulary/test-results/20260816T182207Z-p2652166`.
- Recovery typed terminal evidence and narrow restore participant — PASS, 3/3,
  root `.cartulary/test-results/20260816T182237Z-p2653673`.
- Application extension-readiness admission gate — PASS, 1/1, root
  `.cartulary/test-results/20260816T182306Z-p2655018`.
- `make format` — PASS, root
  `.cartulary/test-results/20260816T182505Z-p2723038`.
- `make generate` — PASS, root
  `.cartulary/test-results/20260816T181947Z-p2643161`.
- `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T182325Z-p2655619`.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T182325Z-p2655540`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T182325Z-p2655662`.
- `make migration-drift` — PASS, 5/5, root
  `.cartulary/test-results/20260816T182340Z-p2659393`.
- `make frontend-import-boundary-check` — PASS, 2/2, root
  `.cartulary/test-results/20260816T182340Z-p2659625`.
- `make backend-module-boundary-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T182433Z-p2689528`.
- `make go-gosec-targeted` — PASS, 4/4, root
  `.cartulary/test-results/20260816T182433Z-p2689583`.
- Pre-checkpoint `git diff --check` and `git diff --cached --check` — PASS.
- Pre-checkpoint `make lint-markdown` — PASS, root
  `.cartulary/test-results/20260816T182514Z-p2726650`.

The first current-binding generation at
`.cartulary/test-results/20260816T175223Z-p2570679` exposed the intended
Recovery catalog digest change; the v3 binding and fixtures were regenerated
against that exact catalog and the final generation root supersedes it. The
first Graph Projection restore row at
`.cartulary/test-results/20260816T181237Z-p2623625` exposed a test expectation
that would have widened the frozen v2 error union; the corrected exact-v2
assertion passes in the final restore root.

Concurrent state-row attempts at
`.cartulary/test-results/20260816T181504Z-p2629023` and
`.cartulary/test-results/20260816T181504Z-p2629026`, followed by focused roots
`.cartulary/test-results/20260816T181705Z-p2637956`,
`.cartulary/test-results/20260816T181841Z-p2641288`, and
`.cartulary/test-results/20260816T181959Z-p2646121`, exposed only fixture and
routing defects: a multi-command prepared statement, an out-of-order synthetic
timestamp, the wrong Postgres fixture capability, and an overlong synthetic row
ID. The final state roots supersede them. The first security closure root
`.cartulary/test-results/20260816T182340Z-p2659807` reported a bounds-analysis
false positive in a hand-written equal-slice loop; using the standard-library
equality primitive removed the ambiguity and the final security root passes.
An attempted nonexistent `backend-import-boundary-check` target was corrected
through `make help-all` to the canonical backend target recorded above.

Capacity workloads, browser workflows, rollout/rollback drills, broad
`make check`, and release verification were skipped because they belong to the
strictly subsequent S10 through S12 workstreams. No S09 owner, compatibility,
readiness, identity, lease, privacy, security, migration, or restore gate is
unresolved.

Post-checkpoint `git diff --check` and `git diff --cached --check` — PASS.
Post-checkpoint `make lint-markdown` — PASS, root
`.cartulary/test-results/20260816T182626Z-p2727838`.

#### Current next action

Execute GP3-S10 Web product completion. Begin with required
`web.networkflow`, Network Flow, and application owner routing; then consume
only major 4, add accessible temporal controls and bucket navigation, preserve
the saved lifecycle and contributor pivots, and bound mounted graph objects to
500 vertices and 1,000 edges. Do not begin capacity or final handoff work before
S10 is `DONE`.

### 10.18 GP3-S10 completion checkpoint

Status: `DONE`.

GP3-S10 moves the browser to the Network Flow major-4 boundary without a
parallel major-3 client. Ephemeral Graph queries, saved-graph creates, and
initial contributor requests now write semantic-query v2 and echo the canonical
Network Flow selectors supplied by the server. Unchanged rename, refresh,
retire, and cursor-continuation shapes retain their owner-approved schema IDs.
The adapter reads current v2 Graph responses and the mixed saved-result union,
so exact retained v1 results remain displayable while all new writes use v2.

Unsaved exploration now exposes the closed default/time-bucket choice, all six
allowed widths, and complete-range validation before transport. Temporal
results navigate the response-owned bucket summaries, show canonical
`[start,end)` bounds and per-bucket counts, and mount no objects for an empty
bucket. Vertex, default-edge, and temporal-edge contributor pivots use the
canonical selector objects rather than Graph Projection `ed_...` identities.

Saved Graph results support the same temporal bucket navigation and canonical
contributor selection while preserving rename, refresh, last-safe result,
failure, retire, and role behavior. Both exploration and saved results page
large selections and mount at most 500 vertices and 1,000 edges at once, with
explicit capacity guidance. A real application browser scenario now creates a
two-bucket temporal graph, navigates its empty bucket by keyboard, saves it,
waits for materialization, opens contributors, renames, refreshes while retaining
the prior result, and retires the declaration.

#### Files changed

Major-4 browser boundary and controllers:

```text
packages/protocol-ts/src/entrypoints/network-flow.ts
packages/protocol-ts/src/entrypoints/entrypoints.test.ts
packages/protocol-ts/src/index.test.ts
apps/web/src/services/{networkFlowContractAdapter.ts,networkFlowContractAdapter.test.ts}
apps/web/src/extensions/extensionAvailability.test.ts
apps/web/src/testing/extensionAvailabilityTestSupport.ts
apps/web/src/networkFlow/{networkFlowClient.ts,networkFlowClient.test.ts}
apps/web/src/networkFlow/useNetworkFlowGraphController.ts
```

Temporal, saved-result, bounded-rendering, and load-fixture behavior:

```text
apps/web/src/networkFlow/NetworkAnalysisWorkspace.tsx
apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx
apps/web/src/networkFlow/NetworkFlowSavedGraphPanel.tsx
apps/web/src/networkFlow/NetworkFlowGridLoadFixture.tsx
```

Browser, visual, and routed evidence:

```text
apps/web/e2e/{network-flow.spec.ts,extensions.stateful.spec.ts}
apps/web/e2e/workbook.visual.spec.ts-snapshots/network-flow-analysis-graph-contributors-linux.png
tools/test_families/web.networkflow.json
tools/execution_topology_render_index.json
```

The topology projection changed only through `make generate`. No generated
protocol source, lockfile, backend route, database schema, Graph Projection
engine policy, cleanup control, or alternate application shell was hand-edited
in S10.

#### Compatibility and migration impact

- The browser advertises and accepts only Network Flow contract major 4. A
  major-3 discovery response is intentionally unsupported; no hidden dual
  decoder or content negotiation remains.
- New Graph and saved-Graph writes use semantic-query v2 and therefore the
  intentional v2 digest/identity boundary. Current v2 saved results and exact
  historical v1 saved results are both readable through the major-4 response
  union; authoritative persisted bytes are not translated.
- Contributor requests now echo complete canonical source selectors and bind
  page limits. Graph Projection edge IDs are no longer sent as Network Flow
  selector identities.
- Rendering caps bound mounted browser objects, not server result identity or
  deployment policy. Operators choose deployable backend limits from S11
  evidence; the UI does not silently clamp an accepted query.
- The visual refresh changed exactly
  `network-flow-analysis-graph-contributors-linux.png`. Its accepted trigger is
  the adopted major-4 aggregation control block; owner row
  `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`
  and fixture `visual.fixture.claimed_network_analysis_workspace_states` own the
  capture. Viewport, zoom, masks, scroll normalization, and screenshot scope are
  unchanged. Review found no unexpected clipping, overflow, typography, or
  contributor-drawer change.

#### Validation

- Required `make task-guide ROLE=module-author OWNER=...` and
  `make explain-test-owner OWNER=...` probes passed for `web.networkflow`,
  `module.networkflow`, and `app.server` before implementation.
- `make frontend-typecheck` — PASS, 2/2, root
  `.cartulary/test-results/20260816T185001Z-p2814181`.
- `make frontend-unit` — PASS, 390/390, root
  `.cartulary/test-results/20260816T185016Z-p2814764`.
- `make frontend-import-boundary-check` — PASS, 2/2, root
  `.cartulary/test-results/20260816T185215Z-p2856656`.
- Full routed `web.networkflow` slice, including the new temporal controls,
  canonical bucket selector, empty bucket, saved lifecycle, role, and render
  bound rows — PASS, 37/37, root
  `.cartulary/test-results/20260816T185251Z-p2860981`.
- `make browser-e2e-stateful` — PASS, 36/36, root
  `.cartulary/test-results/20260816T185845Z-p2927705`.
- Real-browser keyboard bucket navigation is included in that stateful result;
  focused unit keyboard and focus-restoration coverage also passed within the
  390/390 frontend root.
- `make browser-e2e-a11y` — PASS, 14/14, root
  `.cartulary/test-results/20260816T190130Z-p2959211`.
- `make browser-e2e-measurement` — PASS, 29/29, root
  `.cartulary/test-results/20260816T191643Z-p3111057`. The retained rows include
  ordinary Network Flow DOM growth and saved-result object-model ceiling
  evidence.
- `make browser-e2e-visual-update` — PASS, 14/14, root
  `.cartulary/test-results/20260816T191213Z-p3054964`; the refreshed Network
  Flow image was inspected against its prior image and failed-run diff.
- `make browser-e2e-visual` after review — PASS, 14/14, root
  `.cartulary/test-results/20260816T191457Z-p3082826`.
- `make format` — PASS, final root
  `.cartulary/test-results/20260816T192335Z-p3150980`.
- `make generate` — PASS, root
  `.cartulary/test-results/20260816T184938Z-p2811056`.
- `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T185223Z-p2857048`.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T185236Z-p2860011`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T185242Z-p2860485`.
- Pre-checkpoint `git diff --check` and `git diff --cached --check` — PASS.
- Pre-checkpoint `make lint-markdown` — PASS, root
  `.cartulary/test-results/20260816T192346Z-p3154610`.

The initial full web owner run at
`.cartulary/test-results/20260816T182819Z-p2731626` failed 21 of 36 rows because
the browser still consumed major 3, which was the intended S10 starting gap.
The major-4 cutover passed 36/36 before the temporal row was added at
`.cartulary/test-results/20260816T183932Z-p2752897`, and the final 37/37 root
supersedes it. Typecheck root
`.cartulary/test-results/20260816T184419Z-p2758682` exposed three test-fixture
narrowing errors; the final typecheck root supersedes it. An initial
`frontend-unit` invocation stopped before tests because the new catalog row was
not ASCII-sorted; the row was reordered and both complete 390/390 runs passed.

Stateful browser root
`.cartulary/test-results/20260816T185302Z-p2861515` proved the major-4 temporal
request and response but exposed pointer interception at the viewport edge; the
scenario now activates the native button through focus and Enter, adding real
keyboard coverage without forcing a click. Root
`.cartulary/test-results/20260816T185603Z-p2895005` then passed the complete
temporal lifecycle and exposed only one separate stale major-3 packaged-client
expectation; the final 36/36 root supersedes both.

The first measurement root
`.cartulary/test-results/20260816T190256Z-p2986305` passed all three Network
Flow rows but failed an unrelated Timeline blank-row paint observation. The
clean 29/29 retry supersedes it. Initial visual root
`.cartulary/test-results/20260816T190959Z-p3026990` failed only the expected
major-4 control-block screenshot delta. Reconciliation reported the capture as
active and unambiguous with no missing golden. The update target also produced
one unrelated nondeterministic public-error image change; review showed only an
unmasked synthetic actor digit and that file was restored, leaving exactly the
accepted Network Flow golden change. Final visual validation supersedes the
failed comparison.

The browser umbrella, capacity workloads, fresh backup/restore/rollback drills,
`make check`, and `make release-check` were not duplicated in S10 because they
belong to S11 and S12. No S10 contract, compatibility, accessibility,
measurement, visual, or product-workflow risk remains unclassified.

#### Current next action

Execute GP3-S11 capacity and rollout certification. Begin with required owner
routing, then retain default, raised, semantic-maximum diagnostic,
dense/skewed/temporal, cleanup, cancellation, crash, mixed-queue,
browser-response, and shutdown evidence. Capture a fresh state-3 backup and run
isolated restore plus replacement-target rollback drills. Do not begin final
handoff before S11 is `DONE`.

### 10.19 GP3-S11 completion checkpoint

GP3-S11 selects the conservative configuration-major-2 defaults for the GP3
reference deployment. The retained backend capacity artifact binds the dirty
source snapshot, baseline commit, toolchain, system, execution graph, Go
runtime, host CPU/memory profile, every effective limit, workload cardinality,
wall time, allocations, process peak RSS, projection input/result bytes, and
outcome. Raised and semantic-maximum observations are explicitly diagnostic;
they do not publish a portable capacity claim or select an override.

The release-tier capacity row runs the default skew and exact default-dense
limits, a raised dense profile, the full five-million contributing-row maximum,
the default 256-bucket temporal maximum, and the 1,024-bucket semantic maximum.
The first retained run showed that a raised dense workload could exceed the
generic 256 MiB scheduler declaration even while passing. S11 therefore adds
the honest `backend_capacity_isolated` profile with a 1 GiB memory claim and a
one-hour runner timeout. It does not raise any product limit. The final run
observed:

| Workload | Rows | Vertices | Edges | Buckets | Wall ms | Allocated bytes | Process peak RSS | Result bytes | Selection |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| default skew | 250,000 | 2 | 1 | 0 | 350 | 505,674,056 | 39,038,976 | 12,948 | selected default |
| default dense | 250,000 | 5,000 | 10,000 | 0 | 1,674 | 1,811,564,240 | 307,978,240 | 21,694,707 | selected default |
| raised dense | 500,000 | 10,000 | 20,000 | 0 | 3,500 | 3,624,492,608 | 604,905,472 | 43,385,475 | diagnostic only |
| temporal default | 250,000 | 2 | 256 | 256 | 489 | 730,758,568 | 604,905,472 | 468,829 | selected default |
| maximum rows | 5,000,000 | 2 | 1 | 0 | 7,121 | 10,128,069,528 | 604,905,472 | 12,957 | diagnostic only |
| maximum buckets | 1,024 | 2 | 1,024 | 1,024 | 120 | 120,598,136 | 604,905,472 | 1,829,713 | diagnostic only |

Peak RSS is the process high-water mark and is therefore cumulative within the
single isolated row. Database statements and rows are recorded as zero for
these in-memory ordered-iterator/Graph-engine measurements; the separately
routed service-backed rows below supply the PostgreSQL scan, cleanup, race, and
publication evidence. This separation prevents an in-memory metric from being
misrepresented as database work.

#### S11 implementation, Recovery, and operations

- Added `internal/modules/networkflow/graph_capacity_test.go`. It streams rows
  without retaining a source slice, projects every workload through the real
  Graph Projection v2 adapter, validates exact result cardinality and effective
  limits, and writes a private `0600` capacity artifact only beneath the
  Make-owned step artifact directory.
- Added the release-tier informative row
  `module.networkflow.measurement.backend_graph_capacity_certification`, the
  `backend_capacity_isolated` topology resource profile, and the matching
  closed schema/catalog admission updates. Generated topology was refreshed
  through `make generate`.
- Corrected the bounded contributor integration row from a transaction fixture
  to the dedicated PostgreSQL fixture its application runtime actually owns.
  This is verification-routing repair, not a product behavior change.
- Extended the saved-graph lifecycle integration to install a persisted v1
  declaration, materialize it through the supported compatibility refresh
  path, retain a new v2 declaration beside it, and prove Recovery v3 rebuilds
  both result and object identities exactly. New public v1 creation remains
  rejected.
- Extended the vNext capture/restore codec drill with exact mixed v1/v2 Network
  Flow declaration rows. Its fresh state-3 backup uses the v3 Graph source
  registry and implementation binding, restores into a fresh isolated atomic
  target, and writes a private `0600` proof containing the backup identity,
  integrity digest, catalog/codec digests, and both semantic-query digests.
- Added `docs/guides/network_flow_gp3_rollout_and_rollback.md` and linked it
  from the Graph Projection v2 rollout guide. It selects defaults, defines
  evidence requirements, gives the additive `00034`/state-3/major-4 rollout
  order, and makes replacement-target restoration mandatory after state 3 or
  any v2 declaration. It forbids in-place downgrade, translation, identity
  reverse-mapping, migration-history edits, and manual Graph reconstruction.
- The operator restore row exercises the production backup command, fresh
  isolated replacement target, exact confirmed backup identity, Graph and
  workbook rebuild, readiness, and durable terminal replay. Exact historical
  Recovery v2 artifact dispatch remains separately proven and read-only. A
  deployment's exact old binary and pre-GP3 database/object-store backup are
  necessarily rollout inputs; the guide makes their capture and isolated
  verification a fail-closed prerequisite rather than inventing repository
  evidence for an external target.

Files added or substantively changed by S11 are:

- `internal/modules/networkflow/graph_capacity_test.go`;
- `internal/modules/networkflow/routes_integration_test.go`;
- `internal/modules/recovery/vnext_codec_test.go`;
- `tools/execution_topology_manifest.json`;
- `tools/execution_topology_render_index.json`;
- `tools/harness/test-catalog/test-catalog.mjs`;
- `tools/schemas/cartulary.harness_work_graph_owner.v1.schema.json`;
- `tools/schemas/cartulary.test_family_manifest.v5.schema.json`;
- `tools/test_families/module.networkflow.json`;
- `docs/guides/graph_projection_v2_rollout.md`; and
- `docs/guides/network_flow_gp3_rollout_and_rollback.md`.

#### S11 validation and retained evidence

- Required `module.networkflow`, `module.recovery`, `platform.jobs`, and
  `app.server` task guides and owner explanations — PASS before implementation;
  Network Flow reported 74 rows/33 service-backed, Recovery 27/14, Jobs 16/12,
  and application server 34/22.
- Final backend capacity certification — PASS, 1/1, root
  `.cartulary/test-results/20260816T194200Z-p3179780`; retained artifact
  `unit-artifacts/go-5e69a7753cb72544-263e427db896-001/network-flow-capacity-evidence.json`.
- Network Flow PostgreSQL cleanup, cleanup races, contributor cancellation and
  paging, mixed v1/v2 saved lifecycle, and temporal saved lifecycle — PASS,
  7/7, root `.cartulary/test-results/20260816T194347Z-p3182930`.
- Network Flow bounded aggregation, cleanup dispatcher scheduling/fatal-loss,
  temporal arithmetic/identity, timeout, conservation, and property rows —
  PASS, 3/3, root
  `.cartulary/test-results/20260816T194416Z-p3184717`.
- Jobs durable claim recovery, cancellation lifecycle, exact per-worker
  capacity, mixed-kind fairness, retry, renewal, recovery scan, and supervised
  shutdown — PASS, 5/5, root
  `.cartulary/test-results/20260816T194431Z-p3185394`.
- Application-server cleanup readiness/reverse shutdown, telemetry containment,
  lease loss, fatal loop loss, and admission drain — PASS, 2/2, root
  `.cartulary/test-results/20260816T194502Z-p3187307`.
- Fresh vNext state-3 mixed-generation backup, exact atomic restore, current v3
  artifact closure, and exact historical v2 dispatch — PASS, 1/1, root
  `.cartulary/test-results/20260816T194034Z-p3172243`; retained drill proof
  `unit-artifacts/go-33e24180c636aff8-37c4cb98f9a4-001/gp3-mixed-graph-backup-restore-drill.json`.
- Production operator backup/create/inspect, isolated replacement-target
  restore, Graph/workbook rebuild, restore verification, target admission, and
  terminal replay — PASS, 8/8, root
  `.cartulary/test-results/20260816T194518Z-p3188238`.
- Focused real-browser saved-graph response/render ceiling — PASS, 12/12 work
  units, root `.cartulary/test-results/20260816T194617Z-p3220401`. The complete
  S10 measurement result remains
  `.cartulary/test-results/20260816T191643Z-p3111057` and final browser gates
  repeat in S12.
- `make format` — PASS, root
  `.cartulary/test-results/20260816T194123Z-p3173057`.
- Final `make generate` — PASS, root
  `.cartulary/test-results/20260816T194732Z-p3247905`.
- Final `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T194746Z-p3250875`.
- `make test-catalog-check` — PASS with no separate run root emitted.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T194811Z-p3254288`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T194815Z-p3254733`.

The first Recovery codec run at
`.cartulary/test-results/20260816T194017Z-p3169788` failed during compilation
because the new proof writer omitted the `strings` import; the final codec root
supersedes it. The first combined Network Flow service run at
`.cartulary/test-results/20260816T194248Z-p3180543` exposed the pre-existing
transaction/dedicated-fixture routing mismatch for the application-runtime
contributor row. The catalog was repaired and the final 7/7 root supersedes it.
The first post-repair drift root
`.cartulary/test-results/20260816T194715Z-p3244743` correctly reported the stale
generated topology index; the final generation and 4/4 drift roots supersede
it. The initial passing capacity root
`.cartulary/test-results/20260816T193501Z-p3162862` was retained but superseded
after its 569 MiB high-water observation motivated the honest isolated
scheduler profile.

No public response, Graph Projection v2 behavior, database migration, or
selected deployment limit changes in S11. Persisted v1 execution and exact
Recovery v2 dispatch remain deliberate compatibility paths with explicit
zero-inventory retirement conditions. Raised profiles remain configurable but
unselected. Capacity evidence is host-specific; the guide forbids treating it
as a universal SLO. Full browser umbrellas, broad validation, security, final
backup evidence reconciliation, `make agent-finalize`, `make check`, and
`make release-check` remain assigned to S12.

#### Current next action

Post-checkpoint `git diff --check` and `git diff --cached --check` passed.
`make lint-markdown` passed at
`.cartulary/test-results/20260816T194944Z-p3255693`. GP3-S11 is `DONE`; no
capacity, backup, restore, replacement-target, compatibility, or operational
risk is unclassified.

Execute GP3-S12 validation and handoff. Begin with every affected owner guide
and narrow slice, then run the mandated formatting, generation, drift,
migration, security, telemetry, frontend, browser, boundary, finalizer, broad
check, and release sequence. Reconcile all gaps and conditional deployment
evidence in this tracker before marking the iteration complete.

### 10.20 GP3-S12 final validation and handoff

GP3-S12 and the complete GP3 iteration are `DONE`. Network Flow major 4, state
3, and configuration major 2 now own temporal semantics and production policy;
Graph Projection remains a pure deterministic v2 engine. Every GP3 workstream
is complete, every final gate passes, and no active slice remains.

#### Gap closure

| Gap | Status | Final disposition and evidence |
| --- | --- | --- |
| GP3-G01 | `CLOSED` | S03 established one immutable, fail-closed effective-limit value across configuration, startup, discovery, and request overrides; final configuration and server owner slices pass. |
| GP3-G02 | `CLOSED` | S04 replaced whole-scope graph reads with ordered cancellation-aware iteration and bounded incremental aggregates; final Network Flow unit and PostgreSQL slices pass. |
| GP3-G03 | `CLOSED` | S04 added canonical source selectors and selector-aware ordered paging with cursor binding and one-page retention; final Network Flow slices pass. |
| GP3-G04 | `CLOSED` | S05 generated worker runtime mappings and atomically reserves global and per-worker capacity; final Jobs unit and PostgreSQL slices pass. |
| GP3-G05 | `CLOSED` | S06 composes the private cleanup dispatcher after readiness with immediate, cadenced, paced, retry, fatal-loss, and reverse-shutdown behavior; final Network Flow and server slices pass. |
| GP3-G06 | `CLOSED` | S06 uses exact source-owner candidate locking and transaction-local selected-binding and lease checks with the common result-first lock order; migration and race gates pass. |
| GP3-G07 | `CLOSED` | S06 bounds lease batches, result cascades, sweep count, duration, and continuation; batch, rollback, multi-instance, and eventual-drain evidence passes. |
| GP3-G08 | `CLOSED` | S07 adds closed, privacy-safe Jobs queue and Network Flow materialization/cleanup telemetry; owner, conformance, privacy, no-export, and self-failure gates pass. |
| GP3-G09 | `CLOSED` | S11 retains reproducible default, raised, semantic-maximum, skew, dense, temporal, cleanup, cancellation, queue, and shutdown evidence and selects only conservative defaults. |
| GP3-G10 | `CLOSED` | S08 implements the adopted half-open range, epoch-aligned buckets, negative-epoch arithmetic, exact edge/digest identity, metadata, limits, contributors, and failures; temporal unit, integration, fuzz, and property evidence passes. |
| GP3-G11 | `CLOSED` | S09/S11 prove byte-preserving state 3 admission, exact mixed v1/v2 Recovery v3 rebuild, historical v2 dispatch, Reporting redaction, leases, fresh backup, isolated restore, and replacement-target rollback. |
| GP3-G12 | `CLOSED` | S10 completes the major-4 temporal controls, bucket navigation, bounded rendering, saved lifecycle, contributor pivots, failures, accessibility, measurement, visual behavior, and operator guidance; final frontend and browser gates pass. |

#### S12 implementation and compatibility reconciliation

The first broad gate exposed final verification debt that narrow GP3 owners did
not own. S12 corrected it structurally:

- Extension characterization now counts both state-3 proofs, the concurrent
  admission fixture registers the complete `1 -> 2 -> 3` migration chain, and
  the fixed-clock finalizer fixture no longer submits jobs in its own future.
- The intended additive `00034` migration is reflected in the canonical
  migration-schema hash and exact operator evidence digest.
- Static analysis removed obsolete whole-slice filtering, semantic-composition,
  worker-listing, and restore-test assignments left behind by the streaming,
  generated-runtime, and Recovery refactors. New error text follows the Go
  error contract rather than suppressing diagnostics.

Files substantively changed by S12, beyond the implementation already recorded
in S00 through S11, are:

```text
internal/app/operator/operator_migration_evidence_test.go
internal/modules/database_migrations/catalog_characterization_test.go
internal/modules/extensions/contract_test.go
internal/modules/extensions/state_service_test.go
internal/modules/graphprojection/restore_service_test.go
internal/modules/networkflow/graph.go
internal/modules/networkflow/graph_result_cleanup.go
internal/modules/networkflow/graph_result_cleanup_dispatcher.go
internal/modules/networkflow/query.go
internal/platform/extensionstore/finalizer_integration_test.go
internal/platform/jobs/definition_catalog.go
internal/testutil/pgtest/pgtest_test.go
tools/contractgen/recovery_validation.go
docs/handoffs/graphprojection-module-refactor-tracker.md
```

The compatibility posture is final and explicit:

- New public writes use semantic-query v2 and contract major 4. Persisted v1
  declarations remain readable, refreshable, returnable, back-up-able,
  restorable, and rebuildable without byte or identity translation.
- Exact Recovery v2 dispatch remains read-only for supported pre-GP3 backups.
  Removing either compatibility path requires zero installed v1 declarations,
  zero supported retained backup references, and a new owner decision.
- Migration `00034` is the only GP3 database migration; `00032` and `00033`
  remain immutable. After state 3 or a v2 declaration, rollback is an exact
  pre-GP3 backup restored into a replacement target with the prior binary.
- The selected deployment profile remains the conservative defaults. Raised
  and semantic-maximum evidence is diagnostic and is not a portable SLO.
- External Network Flow graph release remains fail-closed. GP3 adds no public
  cleanup route, Graph temporal protocol, dual browser client, or in-place
  downgrade path.

#### Final affected-owner evidence

Every affected owner guide and explanation passed before its S12 slice. The
final routed owner results are:

| Owner | Unit/static result | Service-backed result |
| --- | --- | --- |
| `module.graphprojection` | 10/10, `.cartulary/test-results/20260816T195137Z-p3261971` | 6/6, `.cartulary/test-results/20260816T195219Z-p3264428` |
| `module.networkflow` | 89/89, `.cartulary/test-results/20260816T195230Z-p3265609` | 47/47, `.cartulary/test-results/20260816T195511Z-p3311962` |
| `platform.jobs` | 18/18, `.cartulary/test-results/20260816T195738Z-p3352647` | 14/14, `.cartulary/test-results/20260816T195811Z-p3355211` |
| `module.extensions` | 40/40, `.cartulary/test-results/20260816T200134Z-p3386666` | 26/26, `.cartulary/test-results/20260816T200224Z-p3413459` |
| `platform.config` | 15/15, `.cartulary/test-results/20260816T200311Z-p3438190` | No service-backed rows |
| `platform.telemetry` | 10/10, `.cartulary/test-results/20260816T200319Z-p3439819` | No service-backed rows |
| `module.reporting` | 16/16, `.cartulary/test-results/20260816T200333Z-p3445880` | 8/8, `.cartulary/test-results/20260816T200404Z-p3450858` |
| `module.recovery` | 39/39, `.cartulary/test-results/20260816T200414Z-p3452026` | 26/26, `.cartulary/test-results/20260816T200525Z-p3491056` |
| `app.server` | 45/45, `.cartulary/test-results/20260816T200616Z-p3526497` | 33/33, `.cartulary/test-results/20260816T200714Z-p3554818` |
| `web.networkflow` | 37/37, `.cartulary/test-results/20260816T200804Z-p3578949` | No service-backed rows |
| `module.database_migrations` | 17/17, `.cartulary/test-results/20260816T204656Z-p4139962` | 8/8, `.cartulary/test-results/20260816T204738Z-p4143604` |
| `app.operator` | 17/17, `.cartulary/test-results/20260816T204821Z-p4145525` | 9/9, `.cartulary/test-results/20260816T204903Z-p4167505` |

#### Final cross-cutting evidence

- `make format` — PASS, final root
  `.cartulary/test-results/20260816T204552Z-p4122953`.
- `make generate` — PASS, root
  `.cartulary/test-results/20260816T200829Z-p3586877`.
- `make generate-drift` — PASS, 4/4, root
  `.cartulary/test-results/20260816T200840Z-p3589821`.
- `make generated-artifact-policy-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T200850Z-p3592752`.
- `make json-shape-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T200854Z-p3593197`.
- `make migration-drift` — PASS, 5/5, root
  `.cartulary/test-results/20260816T200903Z-p3593686`.
- `make go-vulncheck` — PASS, 4/4, root
  `.cartulary/test-results/20260816T200915Z-p3596520`.
- `make go-gosec-targeted` — PASS, 4/4, root
  `.cartulary/test-results/20260816T200939Z-p3597389`.
- `make go-gosec-audit` — PASS, 4/4, root
  `.cartulary/test-results/20260816T200951Z-p3623596`.
- `make otel-conformance` — PASS, 14/14, root
  `.cartulary/test-results/20260816T201001Z-p3638753`.
- `make frontend-typecheck` — PASS, 2/2, root
  `.cartulary/test-results/20260816T201011Z-p3642892`.
- `make frontend-unit` — PASS, 390/390, root
  `.cartulary/test-results/20260816T201023Z-p3643438`.
- `make frontend-import-boundary-check` — PASS, 2/2, root
  `.cartulary/test-results/20260816T201213Z-p3681344`.
- `make lint-biome` — PASS, 2/2, root
  `.cartulary/test-results/20260816T201220Z-p3681798`.
- `make browser-e2e` — PASS, 67/67, root
  `.cartulary/test-results/20260816T201227Z-p3682320`.
- `make browser-e2e-webserver-backed` — PASS, 62/62, root
  `.cartulary/test-results/20260816T202002Z-p3739653`.
- `make browser-e2e-stateful` — PASS, 36/36, root
  `.cartulary/test-results/20260816T202455Z-p3821599`.
- `make browser-e2e-a11y` — PASS, 14/14, root
  `.cartulary/test-results/20260816T202718Z-p3850837`.
- `make browser-e2e-measurement` — PASS, 29/29, final root
  `.cartulary/test-results/20260816T212622Z-p449485`.
- `make browser-e2e-visual` — PASS, 14/14, root
  `.cartulary/test-results/20260816T203531Z-p3917617`.
- `make backend-module-boundary-check` — PASS, 3/3, root
  `.cartulary/test-results/20260816T203715Z-p3944854`.
- Final `make agent-finalize
  RESULTS_DIR=.cartulary/test-results/20260816T204956Z-p4190652` — PASS, 1/1,
  root `.cartulary/test-results/20260816T213326Z-p489247`.
- `make check` — PASS, 784/784, root
  `.cartulary/test-results/20260816T204956Z-p4190652`.
- `make release-check` — PASS, 974/974, root
  `.cartulary/test-results/20260816T213414Z-p492324`.

The first S12 Extension owner root
`.cartulary/test-results/20260816T195823Z-p3356392` failed 3 of 40 rows and
exposed the fixture/characterization drift corrected above. The focused repair
root `.cartulary/test-results/20260816T200103Z-p3384915` passed 5/5 before the
complete owner rerun. The first broad root
`.cartulary/test-results/20260816T203739Z-p3948125` failed five work units: four
new static-analysis findings plus the additive-migration hash/evidence changes,
with the migration test repeated by the raw integration sweep. Final lint,
owner, and broad results supersede it.

The first release root
`.cartulary/test-results/20260816T205838Z-p187726` passed 972/974 but an
unrelated Timeline blank-row paint predicate did not qualify before timeout.
Its committed row existed; only the versioned/visible summary match was late.
The canonical measurement retry passed 29/29 and the complete release retry
passed 974/974, so the failure is classified as transient environment
qualification and is superseded without a product or threshold change.

The first two S12 finalizer invocations used no `RESULTS_DIR` because a
qualifying successful full warm check did not yet exist; retained-run
maintenance was therefore skipped as required. After the successful full check
existed, the final invocation bound that run and passed. No visual-update gate
was run in S12 because S12 introduced no intentional visual delta; the final
visual comparison passed against the reviewed S10 golden. No external
production deployment or customer backup was fabricated: the exact old binary,
pre-GP3 backup, environment-specific limit selection, and replacement target
remain fail-closed deployment prerequisites in the rollout guide.

#### Final residual-risk and next-action disposition

Capacity measurements remain environment-specific, persisted-v1 and Recovery
v2 compatibility remain inventory-gated, and production rollout requires an
operator-controlled pre-GP3 backup and replacement target. These are explicit
operational constraints, not unclassified remediation gaps. There is no
remaining temporary path, failed correctness/security/privacy/restore/rollback
gate, or deferred GP3 implementation.

Final checkpoint `git diff --check` and `git diff --cached --check` passed.
`make lint-markdown` passed at
`.cartulary/test-results/20260816T215251Z-p721352`. The next action is none.
Preserve this tracker as the final GP3 handoff; any later temporal mode or
compatibility removal begins with a new owner decision and plan.

## 11. GP4 — Legacy Retirement and Production Simplification

### 11.1 Iteration purpose and control boundary

GP4 removes compatibility and abstraction weight that no longer improves the
future Graph Projection product. It makes the retained projection engine,
result capabilities, and Recovery integration small enough to understand and
strict enough to operate as production infrastructure. Clean ownership and
closed current contracts take precedence over retaining historical runtime
paths.

This section is the active implementation plan. Its decisions are
user-authorized planning decisions until GP4-S01 adopts them in the applicable
owners. The adopted owners then govern implementation and their machine
projections. This tracker controls sequence, checkpoints, evidence, and
handoff; it does not independently define product conformance.

GP4-S00 is a document-only slice. It changes this tracker and no product code,
contract projection, generated artifact, database object, or owner document.
Implementation begins only after GP4-S01 adopts the owner changes.

### 11.2 Planning baseline and inspected evidence

The GP4 planning baseline is clean commit
`4d519e21f05d9df69f4c08c1d9a9ef4054e0fb8c`. The worktree was clean before
this tracker edit. `make test-slice OWNER=module.graphprojection` passed all
8 routed rows at
`.cartulary/test-results/20260829T212324Z-p2764224`.

The planning inspection covered:

| Evidence | Planning use |
| --- | --- |
| `docs/graph_projection_nlspec.md` | Current Graph Projection 2.1.0 ownership, pure-v2 engine, result identity, traversal, persistence, and Recovery obligations |
| `docs/network-flow-activity-nlspec.md` | Current contract-major-4/state-3 declaration ownership, semantic-query compatibility, materialization, and Graph restore source behavior |
| Adopted Recovery, Extensions, Core 00, and Core 04 sections | Recovery selection, extension state admission, owner recognition, module-boundary adoption, and acceptance routing |
| `docs/domain.md` | Graph Projection, saved graph, result, job, Network Flow, and Reporting vocabulary and owner navigation |
| `docs/research/nlspec-spec.md` | NLSpec drafting quality guidance only; it is not a Graph Projection instruction source or behavioral authority |
| `contracts/graph-projection/**`, `contracts/network-flow/**`, `contracts/recovery/**`, and `contracts/extensions/**` | Current typed versions, fixtures, compatibility generations, and generator inputs downstream of adopted owners |
| `internal/modules/graphprojection/**` and its PostgreSQL adapters | Engine, semantic, result, restore, export, option, limit, and package-topology evidence |
| Network Flow, Recovery, Reporting, and application assembly consumers | Actual callers, compatibility dispatch, source ownership, result consumption, and composition seams |

Tests, generators, and generated artifacts remain evidence and projections.
They do not supersede the adopted owner documents. No production or
verification code may begin depending on this tracker or another Markdown
file.

### 11.3 Controlling compatibility and identity decisions

GP4 makes a deliberate hard support cut:

- Removal does not wait for an installed-declaration inventory, retained-backup
  inventory, feature flag, staged translator, or dual-reader period.
- Semantic-query v1 declarations and Graph Recovery v2/v3 artifacts become
  unsupported. The runtime, generated client, source registry, dispatcher,
  fixtures, and active generator branches that exist only for them are
  removed.
- Unsupported persisted state fails closed before mutation. GP4 does not
  silently delete, reinterpret, translate, or rewrite a v1 declaration, and it
  does not partially advance the extension state version.
- Unsupported backup bindings fail selection or binding admission before any
  Graph result table is cleared, rebuilt, reconciled, or published. There is no
  fallback to an older dispatcher.
- Git history is the historical source. Active contracts and tests do not keep
  executable fossils merely to prove that an old package or identifier once
  existed.

The hard cut does not authorize unrelated identity churn. GP4 preserves:

- `graph_projection.v2` as the sole Graph engine and result contract identity;
- exact normalized configuration, normalized source, canonical output, vertex,
  edge, source snapshot, and projection-result identity algorithms for inputs
  that are already valid v2 inputs;
- `cartulary.reporting.source_projection_ref.v2` and exact-result Reporting
  consumption;
- bounded traversal and its identity-bearing consumer capability;
- Graph result publication, exact reads, leases, and source-owner cleanup;
- tables and authored migrations `00032`, `00033`, and `00034` without DDL or
  stored-shape changes; and
- the existing Network Flow route-family root where only the versioned graph
  resource payload changes.

GP4 introduces no Graph Projection v3 protocol, in-place data rewrite, public
Graph route family, Graph-owned authorization, generic job facade, or backup
translation service.

### 11.4 Production gap register

| ID | Final status | Baseline evidence | Risk | Required GP4 closure |
| --- | --- | --- | --- | --- |
| GP4-G01 | `CLOSED` | Network Flow current resources accept semantic-query v1 or v2, and state 3 can contain either | Every engine and recovery revision carries an open-ended union | Advance to contract major 5/state 4, admit v2 only, and remove the v1 runtime |
| GP4-G02 | `CLOSED` | Recovery v3 composes mixed-v1/v2 rebuilding and an exact historical v2 dispatcher | Backup compatibility controls current assembly and generator topology | Publish current-only Recovery v4 and reject all old Graph bindings before mutation |
| GP4-G03 | `CLOSED` | Restore contracts, orchestration, and state contribution live in the pure engine root package | Package exports and imports obscure ownership and invite future coupling | Move restore concerns to a dedicated Graph-owned `restore` subpackage |
| GP4-G04 | `CLOSED` | `EngineV2` is an empty type with a constructor and method facade | A state-free abstraction creates lifecycle and extension expectations it cannot satisfy | Replace it with one package function and keep private helpers private |
| GP4-G05 | `CLOSED` | V2 input is adapted through old `graph_view_key`, `op`, relationship-mapping, identifier, and field-name shapes | Compatibility-shaped internals make every new phase harder to add safely | Use one native-v2 internal model with trusted invocation context outside configuration |
| GP4-G06 | `CLOSED` | Runtime merge tokens and projected-type checks are narrower and differently named than the adopted Graph owner | Valid owner-defined configurations can fail or receive unintended semantics | Implement the complete closed v2 type and merge matrices |
| GP4-G07 | `CLOSED` | Integer, string, label, and identifier checks mix IEEE-safe-number and rune ceilings with owner byte/int64 rules | Boundary behavior can disagree across admission, normalization, and execution | Centralize signed-int64, finite-number, and UTF-8 byte rules |
| GP4-G08 | `CLOSED` | Limits contain stale fields and duplicate hard-coded output ceilings | Operators and maintainers cannot identify the actual semantic boundary | Keep one private semantic limit registry projected from the Graph owner |
| GP4-G09 | `CLOSED` | Unused provider interfaces, options, eligibility hooks, clock injection, derivation exports, and optional constructors remain | Dead seams enlarge the supported API and allow invalid production composition | Delete unused seams and make required dependencies constructor-mandatory |
| GP4-G10 | `CLOSED` | A nil restore reconciler can report success without reconciling jobs and leases | Recovery can return ready with incomplete postconditions | Require reconciliation and keep clearing, rebuild, reconciliation, and publication transactional |
| GP4-G11 | `CLOSED` | Broad engine tests and one empty golden do not cover the owner semantic matrix | Refactors can preserve test green while violating typed behavior | Add exhaustive generated or table-driven owner-routed semantic evidence |
| GP4-G12 | `CLOSED` | Negative tests retain lists of removed v1 packages and symbols forever | Deletion leaves permanent maintenance code | Replace fossils with positive import, export, contract, and composition allowlists |

### 11.5 Target owner and contract state

GP4-S01 adopts the following owner state before any projection or runtime edit:

- Graph Projection advances from 2.1.0 to 2.2.0. The result protocol remains
  v2; the revision closes native-v2 internal semantics, merge/type behavior,
  package ownership, required restore dependencies, and the current-only
  Recovery contribution.
- Network Flow advances from 4.0.1 to 5.0.0, public contract major 5, and
  extension state 4. Semantic query remains
  `cartulary.network_flow.graph_semantic_query.v2`.
- Recovery publishes `graphprojection.restore_rebuild.v4`, its corresponding
  source-registry, implementation-binding, and rebuild-result v4 schemas, and
  no historical Graph dispatcher list.
- The Network Flow extension fragment advances to contract major 5, current
  state 4, minimum migratable state 3, v4 Graph rebuild contribution, and the
  state-4 validation algorithm.
- A new adopted `docs/decisions/graphprojection-module-boundary.md` owns only
  internal package, import, constructor, lifecycle, port, and compatibility
  topology. Core 00 reserves the next unclaimed owner-recognition requirement
  (`REQ-00-076` at this baseline), and Core 04 reserves the next unclaimed
  acceptance criterion (`AC-568` at this baseline). If either identifier is no
  longer free when S01 starts, S01 must stop and reconcile the owner catalog
  rather than silently choose a conflicting identifier.
- Core recognition, traceability, Recovery, Extensions, and Network Flow owner
  text are revised in the same adoption slice so no adopted document points to
  the removed compatibility posture.

Network Flow state transition is exact:

1. Fresh claims initialize directly at state 4.
2. State versions 1 and 2 are outside the new minimum migratable range and are
   rejected without running their historical migration algorithms.
3. State 3 may have no migration ledger because it was freshly initialized, or
   it may have the exact complete 1→2 and 2→3 ledger because it was migrated
   under the prior owner. Extensions validates the latter through inert
   `MigrationLedgerDefinition` facts containing only profile/lineage identity,
   migration identity, from/to versions, and the frozen definition digest.
   These facts have no apply or validation algorithm and never become runnable
   migration registrations. A partial or otherwise invalid historical ledger
   fails preflight.
4. `network_flow_activity.state_3_to_4_v1` locks the claimed profile state and
   validates every saved Graph declaration in the same transaction.
5. A declaration whose semantic-query schema is not v2 causes the transition
   to fail atomically. No declaration byte, digest, result reference, or state
   version changes, and the claimed profile remains unavailable at state 3.
6. When all declarations are v2-compatible, the transition performs no row
   rewrite and advances only the extension state metadata to 4.
7. The state-4 validator admits only v2 declarations. The v1 decoder is not
   retained as a migration helper.

Versioned Network Flow resource schemas that directly or transitively expose
the former v1/v2 query union advance from graph-resource v2 to v3. This includes
the saved Graph view and each request, response, accepted summary, mutation
result, list item, or job terminal payload that embeds that view shape.
Unchanged route paths, request identifiers, semantic-query v2, and schemas that
do not reference the changed resource remain byte-identical and retain their
current identifiers. Generated TypeScript entrypoints advance to contract
major 5; no major-4 compatibility client remains in the active workspace.

Recovery v4 is current-only:

- its source registry admits only the Network Flow v2 semantic-query schema;
- its implementation binding names only
  `graphprojection.restore_rebuild.v4` and has no historical dispatch field or
  admits that field only as an exact empty list if the generic Recovery schema
  requires it;
- selection rejects Graph v2/v3 bindings as unsupported before opening the
  Graph participant transaction;
- the old v2/v3 schemas, fixtures, generated codec registrations, recovery
  generations, source constructors, frozen aliases, and validation branches
  are removed from the active build; and
- current v4 restore always reconciles nonterminal jobs and result leases
  before returning ready.

### 11.6 Target module and runtime design

#### Package ownership

`internal/modules/graphprojection` becomes the pure deterministic engine and
completed-result model. It may depend only on the standard library and its own
private files. It must not import Recovery, PostgreSQL, HTTP, authentication,
jobs, Network Flow, Reporting, or another module.

`internal/modules/graphprojection/restore` owns Graph restore contracts, source
registry construction, rebuild orchestration, and the Graph recovery-state
contribution. `postgresrestore` implements its database transaction and writer
ports. Recovery's public compatibility package may alias the new restore
contract only where an actual cross-owner port remains necessary; it must not
re-export dead historical types. Application recovery assembly composes the
current Network Flow source registration, current v4 binding, mandatory
reconciler, and database writer.

The boundary test becomes a positive allowlist of root imports, root exports,
restore exports, adapter interfaces, and application composition. It does not
scan for an ever-growing list of v1 directory or symbol names.

#### Engine entrypoint and identity

The public pure entrypoint becomes:

```go
func ProjectV2(
    ctx context.Context,
    invocation InvocationContextV2,
    rawInput []byte,
) (ProjectionResultV2, error)
```

`EngineV2`, `NewEngineV2`, and the method facade are removed. Invocation holds
trusted graph-view and source-owner context. User configuration does not gain
or receive an injected `graph_view_key`. Relationship mappings have one
internal home. Filter operators remain `operator` throughout parsing,
validation, normalization, and evaluation; no `op` adapter or old field-name
alias remains. Graph-view ID derivation is caller-owned and moves to Network
Flow with exact byte-for-byte identity fixtures.

Numbers are normalized once before validation and execution. Integer accepts
the complete signed-int64 domain. Number accepts finite JSON numeric values and
rejects overflow, NaN, and infinities. Textual ceilings use UTF-8 byte length,
not rune count. Identifier, timestamp, scalar, array, and field-path checks are
shared by admission and execution so they cannot drift.

#### Closed semantic matrix

The engine implements only the owner vocabulary:

- exact field paths use `source_entity_kind`, `source_relationship_kind`, and
  the adopted property and temporal paths; old `kind` aliases are rejected;
- projected types are boolean, integer, number, string, timestamp, identifier,
  and every adopted array counterpart;
- `single_value` requires all admitted canonical values to be equal;
- `first` and `last` select by canonical contributor order;
- `min` and `max` accept only mutually comparable scalar types, compare
  numeric values numerically, and compare string, timestamp, and identifier
  values by their adopted canonical order;
- `sum` accepts integer or number values, checks signed-int64 overflow, and
  rejects a non-finite numeric result;
- `count` returns the integer count of admitted values;
- `set` accepts array-valued projections, flattens one level, removes values
  equal by canonical bytes, and emits canonical order;
- `ordered_list` accepts array-valued projections, concatenates in canonical
  contributor order, and preserves duplicates; and
- invalid type/merge combinations fail validation before derivation and never
  publish a partial result.

Null, omitted, wildcard/default mapping, contributor ordering, cancellation,
and temporal behavior continue to follow the adopted Graph 2.2.0 algorithms.
The implementation does not preserve an old token as an undocumented synonym.

#### Ports and adapters

Remove `ResultPublisherV2`, `ExactResultReaderV2`, `ResultLeaseWriterV2`, and
`ResultMaintenanceV2` when their only users are compile-time assertions.
Consumers depend on the smallest concrete capability or a consumer-owned port.
Unexport `DeriveProjectionResultIDV2`, `ResourceLimits`, restore registry
builders, and other helpers without a cross-package production consumer.

Remove unused `RestoreServiceOptions.Now`, unconditional candidate-validity
hooks, and optional restore constructors. `postgresrestore.New` requires both
the database and reconciler; no constructor can produce a writer that skips
reconciliation. The source enumerator returns only eligible current candidates,
and restore validates them rather than accepting a caller-supplied always-true
eligibility result.

Keep publication, exact read, traversal, lease, cleanup, and restore adapter
capabilities. Split the PostgreSQL result adapter into capability-focused files
only when this reduces file-level coupling; do not introduce new interfaces or
change transaction boundaries merely to obtain smaller files. One private
Graph semantic-limit registry replaces stale fields and duplicated literals.

### 11.7 Slice ledger

| Slice | Current status | Depends on | Required terminal outcome |
| --- | --- | --- | --- |
| GP4-S00 — Baseline and plan | `DONE` | GP3 complete | Tracker contains the verified baseline, hard-cut decision, target topology, ordered slices, gates, and binary completion criteria |
| GP4-S01 — Adopt owner decisions | `DONE` | S00 | Graph 2.2.0, Network Flow 5.0.0/state 4, Recovery v4, boundary ADR, Core recognition, and traceability are adopted together |
| GP4-S02 — Project contracts | `DONE` | S01 | Machine projections and generated outputs encode only the adopted current contracts; drift and shape checks pass |
| GP4-S03 — Retire Network Flow v1 | `DONE` | S02 | State 3→4 is fail-closed and byte-preserving for v2; all active v1 runtime/client paths are absent |
| GP4-S04 — Cut Recovery to v4 | `DONE` | S03 | Current v4 rebuild passes; v2/v3 Graph backup paths fail before mutation and have no active dispatcher |
| GP4-S05 — Correct package ownership | `DONE` | S04 | Pure engine root and Graph-owned restore subpackage satisfy the adopted positive boundary |
| GP4-S06 — Remove dead engine seams | `DONE` | S05 | Native-v2 entrypoint, model, limits, constructors, and minimal exports replace compatibility-shaped scaffolding |
| GP4-S07 — Complete v2 semantics | `DONE` | S06 | Full field/type/merge/numeric/byte matrix passes with stable valid-v2 identities |
| GP4-S08 — Harden persistence | `DONE` | S07 | Publication, traversal, leases, cleanup, and mandatory-reconciler restore pass transaction and failure evidence |
| GP4-S09 — Integrate and clean up | `DONE` | S08 | All consumers use current APIs; orphaned compatibility code, contracts, fixtures, guides, and fossil tests are absent |
| GP4-S10 — Final verification and handoff | `DONE` | S09 | Every affected owner and broad required gate passes; all gaps close and the tracker records terminal evidence |

Every slice is independently reviewable. A slice changes from `BLOCKED` only
after its predecessor records terminal evidence here. A failing gate is fixed
structurally within the current slice; it is not waived, converted into a
permanent exception, or deferred behind a compatibility shim.

### 11.8 Slice implementation and checkpoints

#### GP4-S01 — Adopt owner decisions

- Draft and adopt the Graph, Network Flow, Recovery, Extensions, Core, and
  traceability changes described in section 11.5.
- Adopt the module-boundary ADR without assigning Graph Projection a new domain
  context or public transport surface.
- Reconcile all owner references to v1 declarations, mixed restore, historical
  dispatch, minimum state 1, contract major 4, and the former package topology.
- Run Markdown lint and owner/reference consistency checks. No runtime or
  machine projection changes belong in this slice.

##### GP4-S01 checkpoint — DONE (2026-08-29)

Owner adoption is complete. Graph Projection is adopted at 2.2.0 with the
native-v2 semantic matrix, pure-root/restore-subpackage boundary, single
`ProjectV2` entrypoint, one semantic limit registry, and current-only mandatory-
reconciliation Recovery v4. Network Flow is adopted at 5.0.0, contract major
5, state 4, minimum migratable state 3, semantic-query v2 only, graph-resource
v3, and a forward-only rollout. Extensions 0.10.0 now separates executable
migration definitions from inert committed-ledger verification facts. Core 00
REQ-00-076 adopts the Graph module-boundary decision, Core 01 selects only the
v4 participant, Core 04 AC-568 owns terminal evidence, and Appendix F traces
the new owner boundary. `docs/domain.md` was inspected and requires no edit:
its saved graph, projection result, materialization job, Reporting diagram, and
workbook projection vocabulary already matches the adopted state.

The required S01 correction is adopted explicitly. Valid state-3 Network Flow
installations may have no historical ledger or the exact complete 1→2 and 2→3
ledger. Those transition identities and digests remain only in
`MigrationLedgerDefinition` projections used to authenticate stored history;
their former apply and validator algorithms are not packaged or executable.
State 1/2 remains unsupported, and only 3→4 can execute.

Files changed in this checkpoint are
`docs/graph_projection_nlspec.md`,
`docs/network-flow-activity-nlspec.md`,
`docs/extension-subsystem-nlspec.md`,
`docs/decisions/graphprojection-module-boundary.md`, Core 00, Core 01, Core 04,
Appendix F, Appendix I, and this tracker. The active contract paths in §11.2
were corrected to `contracts/graph-projection` and `contracts/network-flow`.
No product code, authored machine contract, generated artifact, SQL, migration,
database shape, public route, Graph protocol/result identity, Reporting
reference, or domain vocabulary changed in S01.

Gap disposition at this checkpoint: GP4-G01 through GP4-G12 are owner-adopted
and remain implementation-open for S02 through S09. S01 closes the decision
ambiguity for every gap; it does not claim machine projection or runtime
closure. Public compatibility impact is the adopted Network Flow major-5 hard
cut and graph-resource-v3 response family. Internal compatibility impact is the
future package/API cut and mandatory restore dependency. State-4 commit will be
the forward-only runtime rollback boundary; S01 itself is documentation-only
and can be reverted without data impact.

Validation passed:

- uniqueness searches found exactly one `REQ-00-076`, `AC-568`, and
  `EXT-REQ-237`, with `EXT-AC-160` linked from the owner and acceptance table;
- reviewed current-owner searches found no v3 Graph Recovery promise, Graph
  2.1.0 current claim, Network Flow `contract_major=4`, or
  `network_flow_activity@4` outside frozen earlier tracker history;
- `git diff --check` and `git diff --cached --check` passed; and
- `make lint-markdown` passed at
  `.cartulary/test-results/20260829T223104Z-p2805174`, superseding the
  pre-checkpoint pass at
  `.cartulary/test-results/20260829T222915Z-p2803661`.

No failure occurred. Residual risk is deliberate sequencing risk: authored
machine projections and runtime still encode the prior contracts until S02 and
later slices. S02 is now `READY`; all later slices remain `BLOCKED`.

#### GP4-S02 — Project contracts

- Update authored Graph, Network Flow, Recovery, Extensions, verification, and
  generator inputs; regenerate generated roots instead of hand-editing them.
- Add graph-resource v3 schemas only for wire shapes affected by removal of the
  query union. Preserve unchanged schema IDs and exact route identifiers.
- Make Recovery v4 the sole current Graph generation and delete authored v2/v3
  compatibility inputs only after generator references have moved.
- Update contract indices, fixtures, digest registries, schema allowlists,
  verification rows, and generated clients atomically. Generation, drift,
  generated-artifact policy, and JSON-shape checks must all pass.

##### GP4-S02 checkpoint — DONE (2026-08-29)

The adopted owner decisions are now executable machine facts. Graph Projection
projects owner version 2.2.0 and one authored semantic registry containing the
closed twelve-type merge matrix, signed-int64/finite-number/UTF-8 boundaries,
semantic and traversal ceilings, and cancellation cadence. The former duplicate
index limits are gone. Network Flow projects contract major 5 from
`schemas.v3.json`, retains semantic-query v2 and unchanged request schemas where
their shape did not change, advances every saved-view-bearing response to v3,
and generates only the major-5 browser entrypoint. Route paths and identifiers
are unchanged.

Extensions now projects exactly one executable Network Flow migration, 3→4,
and a distinct `migration_ledger_definitions` collection containing inert 1→2,
2→3, and 3→4 identity/from/to/digest facts. The old apply and validation
algorithms are absent from implementation bindings and dependency imports.
The generated state registry says current state 4 and minimum migratable state
3. Recovery projects one capture-current generation,
`recovery.current.workbook_owned.graph_v4`, with only v4 source registry,
implementation binding, rebuild result, journal payload, and semantic-query-v2
facts. Authored Graph v2/v3 bindings, registries, results, catalogs, journal
fixtures, generator inputs, and generated constants were removed. The current
catalog's four Graph rebuild rows now select
`graphprojection.restore_rebuild.v4`.

Authored files changed in this slice comprise the Graph, Network Flow,
Extensions, Imports, Protocol TypeScript, and Recovery contract families;
their schemas and fixtures; `tools/contractgen` Extensions/Recovery generation
and validation; JSON-shape and schema-attachment inputs; the two focused
contract-projection tests and their Network Flow test-family selector; and the
minimum current-only restore contract consumers required for those projection
tests to compile. `make generate` updated the declared generated Go,
TypeScript, extension, import-target, task-surface, and topology outputs. No
generated root was hand-edited.

The minimum consumer refresh removes production use of historical generated
Graph bindings from the Graph contract, Recovery compatibility port,
Recovery runner dispatch, and recovery assembly. The full Recovery selection
and pre-mutation behavior remains S04 work. A pre-S04 mixed-generation test is
temporarily compile-isolated with test-only aliases; those aliases and the
obsolete test are explicitly scheduled for deletion in S04 and are not
production registrations, generated facts, or callable product compatibility.

Gap disposition at this checkpoint: GP4-G01, G02, G06, G07, G08, G11, and G12
have their required machine-contract direction projected but remain open until
their runtime and exhaustive evidence slices. GP4-G03 through G05 and G09/G10
remain implementation-open. No gap is claimed closed from contract generation
alone.

Compatibility impact is intentional and breaking: current Network Flow is
major 5, saved-view-bearing resources are v3, and Graph Recovery artifacts are
v4 only. Route paths, Graph Projection v2/result v2 identities, database
shapes, migrations 00032–00034, Reporting references, and unrelated Recovery
schema history are unchanged. S02 writes no database state, so its rollback
boundary is source/build artifact rollback; the forward-only operational
boundary begins only after S03 commits state 4.

Terminal validation passed:

- `make format` at
  `.cartulary/test-results/20260829T225839Z-p2861598`;
- `make generate` at
  `.cartulary/test-results/20260829T225906Z-p2866225`;
- `make generate-drift` at
  `.cartulary/test-results/20260829T225923Z-p2869119`;
- `make generated-artifact-policy-check` at
  `.cartulary/test-results/20260829T225923Z-p2869148`;
- `make json-shape-check` at
  `.cartulary/test-results/20260829T225923Z-p2869155`;
- `make toolchain-drift` at
  `.cartulary/test-results/20260829T225923Z-p2869180`;
- `make frontend-import-boundary-check` at
  `.cartulary/test-results/20260829T225923Z-p2869482`;
- the focused Graph contract row at
  `.cartulary/test-results/20260829T225937Z-p2873679`; and
- the focused Network Flow contract row at
  `.cartulary/test-results/20260829T225937Z-p2873685`; and
- `git diff --check`, `git diff --cached --check`, and
  `make lint-markdown` at
  `.cartulary/test-results/20260829T230036Z-p2874571`.

Related failures were fixed rather than waived. Initial `make generate` runs
at `20260829T224747Z-p2815956`, `20260829T224848Z-p2819403`, and
`20260829T224911Z-p2821081` exposed, respectively, schema-source ordering, an
old schema path, and an old owner-contract reference. Initial JSON-shape runs
at `20260829T225049Z-p2828869` and `20260829T225220Z-p2836416` exposed a stale
major assertion and the still-current contributor-continuation schema omitted
from the public registry. Initial focused rows at
`20260829T225341Z-p2843870`, `20260829T225710Z-p2850019`,
`20260829T225743Z-p2855448`, and `20260829T225809Z-p2860202` exposed stale
current-consumer references to deleted generated historical constants. Every
failure was change-related and is superseded by the passes above.

Residual risk is confined to the planned sequencing boundary: the runtime has
not yet implemented state 3→4, removed semantic-query-v1 behavior, or completed
the S04 pre-mutation old-artifact rejection proof. S03 is now `READY`; S04 and
later slices remain `BLOCKED`.

#### GP4-S03 — Retire Network Flow v1

- Implement the exact state 3→4 algorithm in section 11.5 and set the profile's
  minimum migratable version to 3.
- Remove semantic-query v1 decode, normalization, digest, list/get, refresh,
  mutation, import/export, backup, frontend, and test-helper branches.
- Move graph-view ID derivation into Network Flow and retain exact default and
  temporal identity fixtures.
- Prove v2 rows advance without byte changes and a single v1 row blocks the
  entire transition without deletion, rewrite, or version advance.

##### GP4-S03 checkpoint — DONE (2026-08-29)

Network Flow now has one executable state step,
`network_flow_activity.migrate_state_3_to_4_v1`, and one current validator,
`network_flow_activity.validate_state_v4`. Fresh profiles initialize directly
at state 4. Stored states 1 and 2 fail preflight before an apply or validator
can execute. State-3 profiles admit both owner-valid history shapes: no prior
ledger, or the exact inert 1→2 and 2→3 ledger prefix. Generic Extensions state
coordination authenticates those stored rows through
`MigrationLedgerDefinition` while executable lookup remains limited to the
3→4 `MigrationDefinition`; a ledger fact cannot become runnable by presence in
the history registry.

The 3→4 step validates every active declaration through Network Flow's current
semantic-query-v2 validator before state metadata or ledger mutation. Valid v2
declaration bytes and digests are not rewritten. Any semantic-query-v1
declaration rejects the whole transaction with declaration bytes, declaration
digests, result bindings, migration ledger, and state metadata unchanged. The
backend parser, response composer, temporal/digest selection, restore-source
enumerator, public resource schemas, browser adapter, generated-client
consumers, and active test helpers now accept or emit only semantic-query v2
and graph-resource v3. The browser admits contract major 5 only.

Graph-view identity derivation moved out of the Graph adapter capability and
into a private Network Flow helper that preserves the existing length-framed
SHA-256 algorithm. Existing default, temporal, projection-result, and semantic
digest fixtures remain exact. The Graph root's obsolete exported identity
helper is intentionally left for S06, where all dead engine exports are
removed together; it no longer has a production Network Flow caller.

Files changed in this checkpoint are
`internal/modules/extensions/catalogs.go`,
`internal/modules/extensions/state.go`,
`internal/modules/extensions/state_service_test.go`,
`internal/app/server/runtime_assembly.go`, the Network Flow graph runtime,
response, temporal, restore-source, extension-state, route, adapter, and
contract-projection files under `internal/modules/networkflow`,
`apps/web/src/services/networkFlowContractAdapter.ts` and its tests,
`apps/web/src/extensions`, `apps/web/src/testing`, the Network Flow browser and
E2E fixtures, `packages/protocol-ts/src/entrypoints/network-flow.ts`,
`packages/protocol-ts/src/index.test.ts`, the Extensions, Network Flow, and web
Network Flow test-family catalogs, and generated projections refreshed by
`make generate`. Recovery state-catalog, v4 binding/result/journal, generation,
and backup-integrity fixtures were corrected together after the full owner
slice exposed a stale S02 v3 digest beneath the already-v4 table algorithms;
the generated Recovery projections were regenerated from those authored
facts.

Gap disposition at this checkpoint: GP4-G01 is runtime-closed except for final
cross-owner cleanup evidence in S09/S10. GP4-G05 closes Network Flow's v1 union
and Graph-view-ID adapter coupling, while its native-v2 Graph model remains S06
work. GP4-G11 gains state-transition, byte-preservation, identity, backend,
service-backed, and browser unit evidence; the exhaustive Graph semantic rows
remain S07. GP4-G02's machine facts are v4-only, but old-artifact selection and
test-fossil removal remain S04. All other gaps retain their declared later
slice.

Compatibility impact is the adopted hard cut: runtime and generated clients
accept Network Flow contract major 5, graph-resource v3, and semantic-query v2
only. No translator, dual reader/writer, fallback, flag, or inventory gate was
added. Valid state-3/v2 data advances without declaration or identity changes;
v1 data remains stored and blocks availability pending separate owner-approved
remediation. Routes, Graph v2 result bytes and IDs, Reporting references,
tables, migrations 00032–00034, and migration hashes are unchanged. State-4
commit is the forward-only rollback boundary; after it commits, a major-4
binary is not a supported downgrade.

Terminal validation passed:

- `make format` at
  `.cartulary/test-results/20260829T233941Z-p3429765`;
- `make generate` at
  `.cartulary/test-results/20260829T233948Z-p3433888`;
- `make test-slice OWNER=module.extensions` at
  `.cartulary/test-results/20260829T232041Z-p3108153`;
- `make test-slice OWNER=module.networkflow` at
  `.cartulary/test-results/20260829T233102Z-p3225605` (34/34);
- `make service-backed-test-slice OWNER=module.extensions` at
  `.cartulary/test-results/20260829T233316Z-p3286143` (22/22), plus the final
  state-v4 row at
  `.cartulary/test-results/20260829T234001Z-p3436756` (3/3);
- `make service-backed-test-slice OWNER=module.networkflow` at
  `.cartulary/test-results/20260829T233407Z-p3327993` (28/28);
- the focused browser major-5 row at
  `.cartulary/test-results/20260829T233835Z-p3428119` (2/2);
- `make frontend-typecheck` at
  `.cartulary/test-results/20260829T231827Z-p3073002`;
- `make frontend-unit` at
  `.cartulary/test-results/20260829T234001Z-p3436830` (390/390); and
- `make frontend-import-boundary-check` at
  `.cartulary/test-results/20260829T233618Z-p3385766`; and
- `git diff --check`, `git diff --cached --check`, and
  `make lint-markdown` at
  `.cartulary/test-results/20260829T234346Z-p3491309`.

Related failures were fixed rather than waived. The first full Network Flow
slice at `20260829T232130Z-p3150365` exposed one stale contract assertion and
the S02 Recovery catalog/binding digest mismatch. Generate failures at
`20260829T232601Z-p3209611`, `20260829T232901Z-p3212160`,
`20260829T233001Z-p3215309`, and `20260829T233022Z-p3216950` traced the complete
catalog, binding, result, generation, and backup-integrity digest cascade; the
pass at `20260829T233948Z-p3433888` supersedes them. The frontend-unit run at
`20260829T233618Z-p3385759` found the one remaining major-4 browser allowlist;
the 390/390 pass supersedes it. An ASCII-order preflight failure after adding
the state-1/2 test-family selector was change-related and corrected before the
terminal runs.

Residual risk is the planned Recovery execution boundary: test-only obsolete
Graph restore generations and old-artifact cases still exist until S04 proves
pre-mutation rejection and deletes them. No active production v1 Network Flow
path remains. S04 is now `READY`; S05 and later slices remain `BLOCKED`.

#### GP4-S04 — Cut Recovery to v4

- Compose only current Network Flow v2 source registration and v4 Graph
  binding; remove historical v2 and pre-workbook v3 constructors and catalogs.
- Delete old Graph codec registrations, generated artifacts, dispatcher
  branches, frozen-selection aliases, and generator validation.
- Validate selection and binding before the participant transaction mutates a
  target. Unsupported old artifacts return the adopted closed error.
- Prove exact v4 rebuild, rollback on failure, mandatory job/lease
  reconciliation, and unchanged target state for old bindings.

##### GP4-S04 checkpoint — DONE (2026-08-29)

Recovery now loads exactly one generated recovery generation,
`recovery.current.workbook_owned.graph_v4`. The generation registry requires
one capture-current entry and rejects empty, multiple, non-current, malformed,
or mixed Graph bindings. Graph source and implementation proofs are classified
only by their current artifact kinds and must then match the selected v4
schema IDs, canonical digests, and bytes. The former v2/v3 schema constants,
historical generation indexing, fallback dispatch tests, codecs, test-only
Graph aliases, and v2 contract-generator validator are removed.

The Restore runner now resolves and validates the complete verification
evidence, including the selected Graph registry and implementation binding,
before recording or entering PostgreSQL/object restore mutation. The vNext
restore service independently repeats artifact admission before its atomic
target callback. Tests inject both v2 and v3 Graph schema bindings and prove
the closed `ErrVNextBackup` result, zero old-artifact body reads at direct
selection, and zero target atomic-mutation calls. At the Graph participant
boundary, unknown binding digests return
`unsupported_graph_restore_generation` before source enumeration,
reconciliation, or publication.

The v4 positive path remains deterministic and transaction-scoped: the exact
current registry admits only Network Flow semantic-query v2, the exact binding
selects `graphprojection.restore_rebuild.v4`, failures retain the existing
rollback behavior, and successful service-backed reconstruction preserves its
ready result, reconciliation counts, and postcondition. No compatibility
translator or fallback was introduced for old artifacts.

Files changed in this checkpoint are
`internal/modules/recovery/generation.go` and its tests,
`internal/modules/recovery/vnext_codec.go`,
`internal/modules/recovery/vnext_codec_test.go`,
`internal/modules/recovery/vnext_graph_restore_artifacts_test.go`,
`internal/modules/recovery/restore.go`,
`internal/modules/graphprojection/restore_service.go` and its tests,
`tools/contractgen/recovery_validation.go`, the Graph and Recovery test-family
manifests, `tools/backend_module_boundaries.json`, generated task/topology
projections, and this tracker. Authored Recovery v4 schemas and fixtures did
not change in S04.

Gap disposition at this checkpoint: GP4-G02 is closed for active contracts,
generation selection, binding admission, runtime dispatch, rollback, and
pre-mutation old-artifact rejection; S09/S10 retain only final residue and
broad verification. GP4-G10 has current reconciliation/rollback evidence but
remains open until S08 makes the dependency unconstructably mandatory and
completes injected-phase/postcondition coverage. GP4-G12 removes the obsolete
historical-generation guard and dispatch tests; positive package/API
allowlists remain S05/S06. Other gaps retain their declared later slice.

Compatibility impact is the intentional Recovery hard cut: Graph v2 and v3
backup bindings are unsupported, are not read through a translator, and fail
before target mutation. Unrelated owner codecs and historical schema identities
inside the one current Recovery generation are unchanged. Current v4 backup
bytes, Graph v2 result identities, Network Flow rows, Reporting references,
routes, database shapes, migrations, and migration hashes are unchanged. S04
writes no deployment state; rollback is a source/binary rollback before a v4
restore is attempted. An old artifact requires separate owner-authorized
remediation, not downgrade dispatch.

Terminal validation passed:

- `make format` at
  `.cartulary/test-results/20260829T235527Z-p3587453`;
- `make generate` at
  `.cartulary/test-results/20260829T235531Z-p3591529`;
- `make generate-drift` at
  `.cartulary/test-results/20260829T235955Z-p3722280`;
- `make json-shape-check` at
  `.cartulary/test-results/20260829T235955Z-p3722290`;
- `make test-slice OWNER=module.graphprojection` at
  `.cartulary/test-results/20260829T235553Z-p3595612` (8/8);
- `make service-backed-test-slice OWNER=module.graphprojection` at
  `.cartulary/test-results/20260829T235649Z-p3613175` (6/6);
- `make test-slice OWNER=module.recovery` at
  `.cartulary/test-results/20260829T235229Z-p3512354` (24/24), with the final
  current-only artifact/target-mutation row at
  `.cartulary/test-results/20260829T235543Z-p3594349`;
- `make service-backed-test-slice OWNER=module.recovery` at
  `.cartulary/test-results/20260829T235744Z-p3630326` (19/19); and
- `make test-slice OWNER=app.operator` at
  `.cartulary/test-results/20260829T235905Z-p3684628` (12/12); and
- `git diff --check`, `git diff --cached --check`, and
  `make lint-markdown` at
  `.cartulary/test-results/20260830T000110Z-p3726113`.

Related failures were fixed or classified with exact evidence. The first
focused Graph row at `20260829T235128Z-p3503255` expected the obsolete generic
binding-unavailable code; it now expects the adopted unsupported-generation
code and is superseded by the 8/8 pass. The parallel Graph/Recovery invocation
at `20260829T235229Z-p3512353` lost a shared
`tmp/test-service-images/warm.stamp.tmp` rename while Recovery passed; this was
a harness concurrency race unrelated to product behavior, and the standalone
Graph rerun passed 8/8 at `20260829T235553Z-p3595612`.

Residual risk is now package ownership and API shape rather than Recovery
compatibility: restore behavior still resides in the Graph root, Recovery
aliases that root's restore types, and the service retains options scheduled
for removal. S05 is now `READY`; S06 and later slices remain `BLOCKED`.

#### GP4-S05 — Correct package ownership

- Move restore service, restore contract, recovery-state contribution, and
  their tests into `internal/modules/graphprojection/restore` without changing
  v4 wire bytes.
- Update PostgreSQL restore, Recovery compatibility ports, Network Flow source
  registration, revision assembly, and application recovery assembly.
- Enforce the pure root and restore topology with positive import/export tests.
- Remove transitional aliases within this slice; do not retain both old and new
  package paths.

##### GP4-S05 checkpoint — DONE (2026-08-29)

Restore ownership now resides exclusively in
`internal/modules/graphprojection/restore`. Its contract, service,
recovery-state contribution, and tests moved atomically; the former root files
were deleted rather than retained as forwarding aliases. The Graph root now
contains deterministic projection behavior only. Its positive boundary test
admits only standard-library imports or a future owner-private
`graphprojection/internal` path, and its package-topology allowlist names the
three owned subpackages. The restore package has an exact positive allowlist
for its production files and non-standard dependencies, plus v4-only
recovery-state contribution evidence.

PostgreSQL restore realization, Network Flow source registration, Recovery
selection/orchestration, recovery-state catalog assembly, and application
recovery assembly now import the Graph-owned restore package. Recovery retains
exactly one minimal cross-owner alias,
`restorecontract.GraphProjectionParticipant`; registry, request, result,
binding, source-state, error, algorithm, and table facts are no longer
re-exported through Recovery. Callers that need those facts depend on their
Graph owner directly. Application server/operator composition builds through
the single new path, and no production or test reference remains to the
deleted root restore surface.

Files changed in this checkpoint are the moved files under
`internal/modules/graphprojection/restore`, deleted root restore files,
`internal/modules/graphprojection/boundary_guard_test.go`, PostgreSQL restore,
Network Flow restore source and integration evidence, Recovery orchestration
and its minimal compatibility port, recovery application evidence,
`internal/app/recoveryassembly`, the Graph test-family manifest, and this
tracker. Generated product contracts and v4 wire fixtures did not change in
S05; the generated task surface was refreshed only because verification
routing now points at the new package.

Gap disposition at this checkpoint: GP4-G03 is closed for production
ownership, imports, package paths, positive boundaries, assembly, and removal
of transitional aliases. GP4-G12 now positively enforces the root package
topology and restore dependency surface; the remaining export/constructor
allowlists belong to S06. GP4-G04, GP4-G05, GP4-G08, and GP4-G09 remain open
for the engine/API cleanup in S06. Other gaps retain their declared later
slice.

Compatibility impact is internal package-path churn only. Recovery v4 schema
bytes, registry and binding digests, result identities, Network Flow default
and temporal identities, routes, persisted rows, SQL, migrations, and database
shapes are unchanged. The move writes no deployment state and can be rolled
back as one source/binary change before S06; no old package alias is available
for an incremental internal rollout.

Terminal validation passed:

- focused pure-root and restore boundary rows at
  `.cartulary/test-results/20260830T000817Z-p3737994` and
  `.cartulary/test-results/20260830T000854Z-p3743116`;
- `make format` at
  `.cartulary/test-results/20260830T001430Z-p3808747`;
- `make generate` at
  `.cartulary/test-results/20260830T001055Z-p3748589`;
- `make generate-drift` at
  `.cartulary/test-results/20260830T002508Z-p4100375`;
- `make test-slice OWNER=module.graphprojection` at
  `.cartulary/test-results/20260830T001622Z-p3868294` (9/9);
- `make test-slice OWNER=module.recovery` at
  `.cartulary/test-results/20260830T001721Z-p3885847` (24/24);
- `make test-slice OWNER=module.networkflow` at
  `.cartulary/test-results/20260830T002222Z-p4016669` (34/34); and
- `make build-server` and `make build-operator` at
  `.cartulary/test-results/20260830T002430Z-p4075047` and
  `.cartulary/test-results/20260830T002430Z-p4075053`.

Related failures were fixed without restoring aliases. The first restore row
failed on three private root test helpers after the move and is superseded by
`20260830T000854Z-p3743116`. Recovery runs
`20260830T001108Z-p3751379` and `20260830T001439Z-p3812920` found stale
assembly test calls and an incorrectly inferred new API name; the exact catalog
row passed at `20260830T001612Z-p3867832` and the full owner passed 24/24.
Network Flow run `20260830T001904Z-p3940579` found the same stale test alias
surface; the focused restore lifecycle passed at
`20260830T002127Z-p3999241` and the full owner then passed 34/34. All failures
were change-related compile failures and are superseded by the listed passing
runs.

Residual risk is now API shape, not ownership: the new restore service still
uses the stateless `EngineV2` facade and an options struct, the root exports
compatibility-shaped v2 structures and duplicated limits, and constructors
still permit seams scheduled for removal. S06 is now `READY`; S07 and later
slices remain `BLOCKED`.

#### GP4-S06 — Remove dead engine seams

- Replace all callers with `ProjectV2`, remove `EngineV2`, and make the internal
  request/config/filter model native v2.
- Remove duplicate mappings, adapters, stale limits, unused provider
  interfaces, clock and eligibility hooks, optional constructors, and exports
  without production consumers.
- Centralize exact owner limits and replace duplicated output-ceiling literals.
- Run compile, boundary, Graph, Network Flow, Recovery, and application assembly
  slices before deleting the last compatibility symbol.

##### GP4-S06 checkpoint — DONE (2026-08-29)

`ProjectV2(ctx, invocation, semanticInput)` is now the sole exported root
function and engine entrypoint. `EngineV2`, `NewEngineV2`, the method facade,
and exported result identity and graph-view identity helpers are gone. Network
Flow and restore call the pure function directly; Network Flow retains the
private graph-view identity derivation it already owned. A positive root
function/method allowlist now fails on an extra engine entrypoint or missing
required result capability.

The private admission model is current-v2-shaped. Trusted `graph_view_id` is no
longer injected into semantic configuration, the request has one relationship-
mapping collection, filter predicates use `operator` without `op` conversion,
and the redundant per-request projection-schema member and v1 identifier
branching are gone. Closed-schema tests reject `graph_view_key` and `op`, while
a two-invocation test proves trusted values do not affect normalized
configuration or source digests and still participate in result identity.
Existing empty, reordered, maximum-bound, Network Flow default, and temporal
identities remain exact.

The Graph runtime now has one private semantic-limit registry. Stale cursor and
list limits, widened 500,000/1,000,000 output ceilings, duplicated declaration,
label, and property-key ceilings, and the exported `ResourceLimits` accessor
are gone. Contract-projection evidence compares every owner-projected semantic
limit to its runtime member and checks the projected label/string byte limits.
Effective projected output ceilings are consistently 100,000 vertices and
250,000 edges, and cancellation cadence comes from the same registry.

Compile-assertion-only publication/read/lease/maintenance interfaces were
removed. Restore service options, clock/engine/supported-generation seams,
candidate-validity hooks, historical-shaped registry/binding maps, and the
unconsumed generic registry constructor were removed. Restore construction is
current-only. `postgresrestore.New` now requires both a database and derived-
state reconciler, rejects either nil dependency, and `ReplaceAll` also fails
closed if an invalid in-package value bypasses construction. The previous
optional constructor and conditional reconciliation branch no longer exist.
Positive restore API and constructor tests describe the supported surface.

Files changed in this checkpoint are the Graph engine/input/type/object/
validation/aggregation/limit/result-port sources and tests; Network Flow's
Graph adapter; the Graph restore contract, service, boundary and semantic
tests; PostgreSQL restore/result adapters and constructor tests; application
recovery assembly; the Graph verification-family manifest and generated task
surface; and this tracker. No owner contract, schema, fixture, generated client,
SQL, migration, route, or persisted shape changed in S06.

Gap disposition at this checkpoint: GP4-G04 and GP4-G05 are closed. GP4-G08
is closed for runtime registry structure, exact machine-projection equality,
stale-limit removal, output ceilings, and cancellation cadence; S07 retains
the boundary behavior matrix. GP4-G09 is closed for the identified dead
interfaces, options, hooks, generation seams, constructors, and mandatory
constructor dependencies. GP4-G12 is closed for positive root/restore function
and constructor boundaries; final residue review remains S09/S10. GP4-G10 now
has an unconstructably mandatory reconciler but remains open until S08 proves
every reconciliation phase, rollback, postcondition, and readiness invariant.
Other gaps retain their declared later slice.

Compatibility impact is an intentional atomic internal Go API break. All
callers moved to `ProjectV2`, the current restore constructor, and the required
reconciler in this slice. There is no compatibility facade. Graph protocol and
Recovery v4 bytes, result/graph-view identities, current Network Flow behavior,
Reporting references, routes, storage, migrations, and migration hashes are
unchanged. S06 writes no deployment state; rollback requires one source/binary
rollback before S07 and cannot mix old and new internal package APIs.

Terminal validation passed:

- `make format` at
  `.cartulary/test-results/20260830T004846Z-p174171`;
- `make generate` at
  `.cartulary/test-results/20260830T004056Z-p4163371`;
- `make generate-drift` at
  `.cartulary/test-results/20260830T004850Z-p178270`;
- `make test-slice OWNER=module.graphprojection` at
  `.cartulary/test-results/20260830T004247Z-p4193519` (9/9);
- `make test-slice OWNER=module.networkflow` at
  `.cartulary/test-results/20260830T004354Z-p17208` (34/34);
- `make test-slice OWNER=module.recovery` at
  `.cartulary/test-results/20260830T004607Z-p76326` (24/24);
- `make test-slice OWNER=module.reporting` at
  `.cartulary/test-results/20260830T004736Z-p131525` (6/6); and
- `make build-server` and `make build-operator` at
  `.cartulary/test-results/20260830T004816Z-p148896` and
  `.cartulary/test-results/20260830T004816Z-p148902`.

Related failures were fixed structurally. Focused run
`20260830T003256Z-p4121107` found one private constructor call after the dead
export was removed and is superseded by `20260830T003315Z-p4121763`.
Runs `20260830T003712Z-p4130870` and `20260830T004104Z-p4166121` found imports
made unused by the native-v2 cleanup. Run `20260830T004127Z-p4174976` found an
overstrict new test expectation that trusted context must change canonical
semantic output; the adopted invariant requires only stable semantic digests
and distinct result identity. The corrected full Graph run passed 9/9. All
these failures were change-related and are superseded by the listed terminal
evidence.

Residual risk is semantic completeness rather than API structure: the runtime
still has the former safe-integer ceiling, incomplete projected-type and merge
behavior, and remaining rune-count paths. S07 is now `READY`; S08 and later
slices remain `BLOCKED`.

#### GP4-S07 — Complete v2 semantics

- Implement the closed field, type, merge, numeric, byte, null/default, and
  wildcard matrices from section 11.6.
- Generate or table-drive exhaustive cases from authored contract facts; tests
  must not read Markdown.
- Add signed-int64 limits, numeric overflow/non-finite, multibyte boundary,
  invalid pairing, reorder, cancellation, and no-partial-result evidence.
- Re-run exact existing Network Flow default and temporal identity fixtures;
  any valid-v2 identity change blocks the slice unless an owner decision is
  reopened.

##### GP4-S07 checkpoint — DONE (2026-08-29)

The Graph runtime now implements the owner-projected twelve-type matrix
exactly. Boolean, integer, number, string, timestamp, identifier, and all six
array types admit only their declared merge behaviors. `single_value` uses
canonical equality; `first` and `last` consume canonical contributor order;
scalar extrema use exact numeric, timestamp, or canonical text ordering;
integer sum detects signed-int64 overflow; number sum rejects non-finite
results; integer `count` counts admitted values; and array `set` and
`ordered_list` implement their distinct flattening, ordering, and duplicate
rules. Invalid pairs are fatal before derivation, and merge conflicts return no
partial projection result.

Numeric and text admission now uses one path throughout schema admission,
defaults, filters, direct mappings, aggregation, canonicalization, and
execution. Integer lexemes admit the full signed-int64 range and reject either
overflow; number lexemes and sums must remain finite. Identifiers, labels,
strings, property keys, and field paths use UTF-8 byte ceilings from the
private runtime registry, invalid UTF-8 is rejected before parsing, and flat
array cardinality is an owner-projected semantic limit. Property and metadata
object keys are validated consistently rather than only when a later mapping
references them.

The authored semantic registry now contains the array-item ceiling in addition
to the complete field/type/merge, text, traversal, output, validation, and
cancellation facts. A generated non-empty semantic golden projects direct and
aggregated vertices and fixes the result ID, all three semantic/output digests,
canonical contributor ordering, integer sum, first-value, and array-set output.
The new positive semantic verification row reads generated contract artifacts,
not Markdown, and exhaustively proves every valid and invalid matrix pair,
empty/single/multiple inputs, repeated values, overflow, non-finite results,
flat arrays, null/default/omit behavior, wildcard application, source reorder,
multibyte boundaries, invalid UTF-8, and absence of partial output. Existing
cancellation evidence remains routed in the pure-engine row.

Files changed in this checkpoint are the authored Graph semantic registry,
contract index, and non-empty golden fixture; generated Graph contract
artifacts; Graph aggregation, engine admission, input, limit, projection,
semantic, validation, contract-projection, and engine/semantic tests; the Graph
verification-family manifest and generated task surface; and this tracker. No
Network Flow resource bytes, public route, Recovery v4 contract, Reporting
reference, SQL, DDL, migration, migration hash, persistence transaction, or
deployment state changed in S07.

Gap disposition at this checkpoint: GP4-G06 and GP4-G07 are closed. GP4-G08
is closed for the owner-projected semantic and text registries, runtime
equality, array ceiling, cancellation cadence, and removal of duplicate engine
admission literals. GP4-G11 is closed for deterministic non-empty golden,
matrix, identity, boundary, cancellation, and failure evidence routed to
`module.graphprojection`; final harness ownership review remains S10. The
semantic portions of GP4-G05 are closed with native-v2 values from admission
through output. GP4-G10 remains open for S08 transaction-phase reconciliation
evidence; all other residual work retains its declared slice.

Compatibility impact is intentionally asymmetric. Previously rejected
signed-int64 values are now owner-valid. Multibyte text that exceeded the
adopted byte ceiling while passing a rune-count check is now rejected, as are
undocumented type/merge combinations. Currently valid Network Flow default and
temporal graph-view IDs, projection result IDs, and digests remain exact. No
translator, legacy token synonym, compatibility switch, data rewrite, or
rollback migration was introduced. S07 writes no persisted state and may be
rolled back only with the matching source/binary set before a later state-4
deployment writes owner-current data.

Terminal validation passed:

- `make format` at
  `.cartulary/test-results/20260830T010427Z-p199766`;
- `make generate` at
  `.cartulary/test-results/20260830T010431Z-p203866`;
- the new semantic row at
  `.cartulary/test-results/20260830T010446Z-p206696`;
- `make test-slice OWNER=module.graphprojection` at
  `.cartulary/test-results/20260830T010500Z-p207212` (9/9);
- the Network Flow default and temporal identity rows at
  `.cartulary/test-results/20260830T010617Z-p225753`;
- `make generate-drift` at
  `.cartulary/test-results/20260830T010748Z-p227731`;
- `make generated-artifact-policy-check` at
  `.cartulary/test-results/20260830T010759Z-p230738`; and
- `make json-shape-check` at
  `.cartulary/test-results/20260830T010803Z-p231207`; and
- tracker-scoped `git diff --check`, `git diff --cached --check`, and
  `make lint-markdown` at
  `.cartulary/test-results/20260830T010908Z-p232066`.

The initial semantic-row run
`.cartulary/test-results/20260830T010307Z-p198544` failed only because the new
golden intentionally contained `PENDING` identity placeholders. The observed
identities were copied into the authored fixture, regenerated, and superseded
by the passing semantic and full-owner runs above. This was a change-related
test-fixture completion failure, not an environment failure or product waiver.

Residual risk is now concentrated in persistence: required reconciliation is
constructable only with both dependencies, but S08 must still prove every
clear, rebuild, publication, job, lease, postcondition, readiness, collision,
traversal, cleanup, retry, lock-order, and multi-instance invariant. S08 is now
`READY`; S09 and S10 remain `BLOCKED`.

#### GP4-S08 — Harden persistence

- Make reconciliation a required restore dependency and preserve atomic clear,
  rebuild, publish, job, lease, and readiness postconditions.
- Retain and test result publication collision handling, exact read, bounded
  traversal, lease acquisition/release, cleanup ordering, batching,
  cancellation, and retry behavior.
- Decompose the result store by capability if the review confirms it reduces
  coupling, while preserving SQL, lock order, transaction scope, and public
  behavior.
- Prove no DDL or migration hash changed.

##### GP4-S08 checkpoint — DONE (2026-08-29)

Restore publication now fails with an exact invalid-result cause for malformed
plans or invalid reconciliation counts instead of misclassifying them as
cancellation. The required reconciler runs after immutable results are staged
and before postcondition verification or commit. The writer verifies aggregate
result, vertex, edge, and lease counts, reloads every exact binding, and compares
the complete stored result, ordered vertex, and ordered edge bytes both inside
the mutation transaction and again after commit. Only after the post-commit
proof succeeds can the restore service publish `ready`.

New phase-injection evidence proves rollback with byte-identical prior target
state when clear, immutable publication, derived-state reconciliation,
reconciliation count validation, or transaction postconditions fail. Separate
tests prove that a commit failure is classified as indeterminate whether the
write did or did not become visible, and that post-commit verification failure
is also indeterminate. Invalid proof digests, rebuilt bindings, job counts, or
lease counts cannot produce readiness. The assembled Network Flow lifecycle
continues to prove in one transaction that Network Flow and Reporting
nonterminal jobs become reclaimable, Reporting leases are recreated for exact
rebuilt bindings, and the resulting Graph object identities remain unchanged.

Result persistence evidence now covers idempotent byte-identical publication,
identity collision rejection, rollback without partial child rows, exact read,
binding mismatch, not-found behavior, bounded ordered vertex reads, outgoing,
incoming, both-direction, zero-depth, kind-filtered, and one-over traversal,
lease acquisition, renewal, release, expired-lease batching, owner-scoped
cleanup, selected/leased-result preservation, cascade cleanup, cancellation,
continuation, rollback, skip-locked multi-transaction behavior, and concurrent
publication and lease-admission races. The result store remains one file: its
capabilities are already separate concrete types with one shared validation
and identity layer, so a mechanical file split would not reduce runtime
coupling or transaction complexity.

The executable semantic/traversal registry moved into Graph's Go-internal
`semanticlimits` package. The root and PostgreSQL adapter now consume the same
facts without adding a public root accessor, and public output constants alias
the one private literal source. The positive root boundary admits exactly this
private dependency and the topology test admits exactly this private package.
Storage-maintenance tests compare the expired-lease batch and Network Flow
sweep/cadence values with the generated Graph storage contract.

Files changed in this checkpoint are Graph's private semantic-limit package,
root limit aliases and positive boundaries, result port and PostgreSQL result
adapter/tests, PostgreSQL restore writer/tests, restore service proof tests,
Network Flow storage-contract projection test, the Graph verification-family
manifest and generated task surface, and this tracker. No SQL, query, DDL,
migration, migration hash, table/index shape, Recovery v4 artifact, Graph result
identity algorithm, Network Flow route, Reporting reference, or browser
resource changed in S08.

Gap disposition at this checkpoint: GP4-G10 is closed. Reconciliation is
mandatory at construction and execution, every precommit phase rolls back,
every postcommit uncertainty is indeterminate, exact postconditions gate
readiness, and the real source-owner jobs and Reporting leases are reconciled
atomically. GP4-G08 is closed end to end for shared engine and traversal facts
and contract-checked storage-maintenance bounds. The persistence portions of
GP4-G02, GP4-G09, and GP4-G11 are closed with current-only publication,
required dependencies, and routed transaction/race evidence. S09 retains only
consumer cleanup, active-guide cleanup, generated/browser integration, and the
reviewed historical search; S10 retains terminal verification and handoff.

Compatibility impact is internal hardening only. Invalid internal plans now
receive a more accurate cause, and restore never reports ready without an exact
committed proof. Valid current v4 restores, Graph result bytes and identities,
Network Flow selected bindings, Reporting leases, table shapes, lock order, and
transaction boundaries are unchanged. S08 creates no migration or deploy-time
state transition. Rollback requires the matching source/binary set; an
indeterminate production commit must be resolved by the existing closed
Recovery retry/evidence workflow rather than assumed successful or failed.

Terminal validation passed:

- `make task-guide ROLE=module-author OWNER=module.graphprojection` selected
  the focused and service-backed Graph slices;
- restore phase/publication evidence at
  `.cartulary/test-results/20260830T011524Z-p240822` (3/3);
- result, lease, traversal, and cleanup evidence at
  `.cartulary/test-results/20260830T011807Z-p263807` (4/4);
- Network Flow cleanup, race, and full saved-graph restore lifecycle evidence
  at `.cartulary/test-results/20260830T011913Z-p281580` (5/5);
- Reporting exact-result lease lifecycle evidence at
  `.cartulary/test-results/20260830T012006Z-p299555` (3/3);
- `make migration-drift` at
  `.cartulary/test-results/20260830T012048Z-p316616` (5/5), with no authored
  migration or query diff and migrations `00032`, `00033`, and `00034`
  unchanged;
- `make test-slice OWNER=module.graphprojection` at
  `.cartulary/test-results/20260830T012157Z-p324130` (9/9);
- `make service-backed-test-slice OWNER=module.graphprojection` at
  `.cartulary/test-results/20260830T012251Z-p341752` (6/6);
- the Network Flow runtime/storage-maintenance projection row at
  `.cartulary/test-results/20260830T012352Z-p359215`;
- `make generate` and `make generate-drift` at
  `.cartulary/test-results/20260830T012401Z-p359723` and
  `.cartulary/test-results/20260830T012413Z-p362533`; and
- final `make format` and the positive root-boundary row at
  `.cartulary/test-results/20260830T012444Z-p365764` and
  `.cartulary/test-results/20260830T012448Z-p369892`; and
- tracker-scoped diff checks and `make lint-markdown` at
  `.cartulary/test-results/20260830T012619Z-p370949`.

The first S08 `make format` invocation found the newly extended verification
test list was not ASCII-sorted and stopped before emitting a retained run root.
The manifest was sorted and the failure was superseded by all listed format,
generation, focused, and service-backed passing evidence. It was a
change-related authored-artifact error, not an environment limitation or
waiver.

Residual risk is now integration residue: current code and storage behavior
are proven, but active generated/browser references, guides, comments,
fixtures, and historical identifiers must be reviewed and either removed or
listed as declared immutable/frozen exceptions. S09 is now `READY`; S10 remains
`BLOCKED`.

#### GP4-S09 — Integrate and clean up

- Update Network Flow, Recovery, Reporting regressions, Extensions, server and
  recovery assembly, generated clients, browser surfaces, and operator guides
  to current contracts only.
- Delete orphaned schemas, fixtures, constructors, adapters, aliases, tests,
  comments, and rollout guidance that promise v1 or old-backup support.
- Search authored and generated active roots for old graph query, state,
  recovery, engine, adapter, and package identifiers. Allow matches only in
  immutable SQL history and frozen tracker history, with each exception listed
  at the checkpoint.
- Run all affected owner and browser slices and record failures, fixes, and
  superseding roots.

##### GP4-S09 checkpoint — DONE (2026-08-29)

Every production consumer now resolves the current Graph API and the dedicated
Graph restore package directly. Network Flow and the generated TypeScript
client expose contract major 5, graph resources v3, and semantic-query v2 only.
Recovery and operator assembly expose the single v4 Graph registry, binding,
result, and journal. Reporting continues to consume exact Graph v2 result
references, and Extensions exposes state 4 plus inert migration-ledger proof
facts without a historical execution algorithm. Server, operator, browser,
stateful, measurement, accessibility, and visual composition all pass with
these exact surfaces.

The active operator guide now describes Graph 2.2.0, Network Flow major 5/state
4, Recovery v4, the hard current-only cut, isolated restore, and forward-only
rollback. The obsolete GP3 rollout guide was deleted. One Reporting owner
sentence was made phase-neutral, and the superseded Network Flow table no
longer preserves the retired semantic-query schema identifier; §27 remains the
sole current conflict authority. Graph restore test helpers were renamed for
the current protocol. Network Flow contract evidence now asserts the complete
positive Graph schema allowlist, while unsupported declaration and old-artifact
tests construct their inputs without turning deleted identifiers into a
permanent static fossil.

The exact removal scan covers old engine constructors, test helpers, semantic-
query schema IDs, Graph restore v2/v3 algorithms and schema IDs, major-4
frontend entrypoints, the former Network Flow schema bundle, and the deleted
GP3 guide. It returns no match in active owners, guides, decisions, production,
tests, contracts, generators, generated clients, or immutable SQL. The only
permitted matches are frozen entries in this controlling tracker: GP2's v2
restore and `EngineV2` history; GP3's v3 restore, v2 schema/major-4 client, and
rollout-guide history; and the GP4 baseline/gap/removal record that explains
their deletion. Git remains the source of deleted artifact history.

Files changed in this checkpoint are the current Graph rollout guide, deleted
GP3 rollout guide, Network Flow and Reporting owners, Network Flow Graph
contract/state/route/store test support, Recovery generation validation test,
Graph restore service test, Extensions verification-accounting test, and this
tracker. All broader production, contract, generated, frontend, and assembly
changes were introduced by their owning earlier slices and were reviewed here
as one current-only integration. No SQL, DDL, migration, migration hash, public
route, Graph protocol/result identity, Reporting reference, authorization
surface, domain vocabulary, or persisted table shape changed in S09.

Gap disposition at this checkpoint: GP4-G01 and GP4-G02 are closed across
active clients, generated artifacts, browser composition, operator guidance,
and old-name residue. GP4-G03 through GP4-G10 remain closed under their earlier
owner/runtime/persistence evidence. GP4-G11 is closed for active verification
routing and browser integration. GP4-G12 is closed with positive Graph root,
restore, schema, constructor, and composition allowlists plus the one terminal
search inventory above. S10 retains only the complete repository verification
and handoff closure; it does not retain an implementation remediation.

Compatibility impact remains the adopted hard cut. The removed GP3 guide and
historical identifiers provide no runtime compatibility loss beyond the owner-
adopted major-5/state-4 and Recovery-v4 boundary. A target that commits state 4
cannot run a major-4 binary; saved v1 declarations and old Graph backups remain
untouched and require separately authorized remediation. S09 writes no
deployment state. Source rollback must restore the complete matching binary,
contract, generated-client, and guide set; state 4 is not downgraded in place.

Terminal validation passed:

- generation, drift, generated-policy, JSON-shape, and toolchain gates at
  `.cartulary/test-results/20260830T013301Z-p379650`,
  `.cartulary/test-results/20260830T013310Z-p382364`,
  `.cartulary/test-results/20260830T013318Z-p385309`,
  `.cartulary/test-results/20260830T013319Z-p385720`, and
  `.cartulary/test-results/20260830T013322Z-p386132`;
- focused Graph, Network Flow, Recovery, Extensions, Reporting, server,
  operator, and web Network Flow slices at
  `.cartulary/test-results/20260830T013330Z-p386574` (9/9),
  `.cartulary/test-results/20260830T013421Z-p404056` (34/34),
  `.cartulary/test-results/20260830T013626Z-p462642` (24/24),
  `.cartulary/test-results/20260830T013923Z-p564484` (24/24),
  `.cartulary/test-results/20260830T014017Z-p606657` (6/6),
  `.cartulary/test-results/20260830T014058Z-p623929` (24/24),
  `.cartulary/test-results/20260830T014154Z-p666311` (12/12), and
  `.cartulary/test-results/20260830T014230Z-p704031` (37/37);
- service-backed Graph, Network Flow, Recovery, Extensions, Reporting, server,
  and operator slices at
  `.cartulary/test-results/20260830T014251Z-p705705` (6/6),
  `.cartulary/test-results/20260830T014346Z-p722872` (28/28),
  `.cartulary/test-results/20260830T014552Z-p780621` (19/19),
  `.cartulary/test-results/20260830T014711Z-p834860` (22/22),
  `.cartulary/test-results/20260830T014759Z-p876505` (4/4),
  `.cartulary/test-results/20260830T014841Z-p893533` (17/17), and
  `.cartulary/test-results/20260830T014938Z-p935206` (9/9);
- frontend typecheck, 390/390 unit tests, import boundary, and Biome at
  `.cartulary/test-results/20260830T015019Z-p972618`,
  `.cartulary/test-results/20260830T015029Z-p973104`,
  `.cartulary/test-results/20260830T015032Z-p973470`, and
  `.cartulary/test-results/20260830T015035Z-p973879`;
- browser E2E, webserver-backed, stateful, measurement, accessibility, and
  visual targets at `.cartulary/test-results/20260830T015042Z-p974379`
  (51/51), `.cartulary/test-results/20260830T015558Z-p1048844` (60/60),
  `.cartulary/test-results/20260830T020144Z-p1105285` (34/34),
  `.cartulary/test-results/20260830T020412Z-p1153965` (22/22),
  `.cartulary/test-results/20260830T020836Z-p1210524` (12/12), and
  `.cartulary/test-results/20260830T020959Z-p1254415` (12/12); and
- final formatting, backend boundary, migration drift, exact-name scans, and
  tracker-scoped diff checks at
  `.cartulary/test-results/20260830T013919Z-p560373`,
  `.cartulary/test-results/20260830T021158Z-p1299153`, and
  `.cartulary/test-results/20260830T021200Z-p1299491`; and
- checkpoint Markdown lint at
  `.cartulary/test-results/20260830T021417Z-p1303333`.

Focused Extensions run
`.cartulary/test-results/20260830T013810Z-p517969` passed 23/24 and exposed a
stale static selector count after the state-1/2 preflight test was added to the
owner row. The positive verification-accounting expectation was corrected from
two to three selectors and is superseded by the 24/24 root above. This was a
change-related test-accounting failure, not a product behavior or environment
waiver.

Residual risk is limited to terminal repository-wide regression, static,
security, finalizer, check, and release evidence. S10 is now `READY`; no
implementation or compatibility cleanup remains blocked behind it.

#### GP4-S10 — Final verification and handoff

- Run the final verification matrix in section 11.9 from a clean generated
  state and reconcile every failure rather than carrying a waiver.
- Run `make agent-finalize` with the successful full warm-check `RESULTS_DIR`;
  if no qualifying run exists, record the prescribed skipped-maintenance
  reason and obtain a qualifying run before final completion.
- Close every GP4 gap, record the final compatibility search exceptions,
  document any environment-only limitation, and mark all slices `DONE` only
  after `make check` and `make release-check` pass.

##### GP4-S10 checkpoint and terminal handoff — DONE (2026-08-29)

GP4 is complete. All twelve gaps in section 11.4 are `CLOSED`, every slice from
S00 through S10 is `DONE`, every binary completion criterion is satisfied, and
there is no successor implementation action. The final tree has one current
Graph Projection subsystem: Graph owner 2.2.0 and protocol v2, Network Flow
owner 5.0.0/contract major 5/state 4 with semantic-query v2 only, Recovery v4
only, a pure root with `ProjectV2`, a dedicated Graph restore subpackage,
native-v2 internal semantics, one private semantic-limit registry, mandatory
restore reconciliation, and positive architectural allowlists.

Final gap disposition is exact:

- GP4-G01 is closed by the atomic state 3→4 hard cut, v2-only runtime/client,
  and inert earlier ledger-proof facts;
- GP4-G02 is closed by the single v4 Recovery generation and pre-mutation old-
  artifact rejection;
- GP4-G03 is closed by Graph-owned restore and the dependency-pure engine root;
- GP4-G04 is closed by the sole `ProjectV2` engine entrypoint;
- GP4-G05 is closed by native-v2 names, shapes, trusted context, and private
  Network Flow graph-view identity;
- GP4-G06 and GP4-G07 are closed by the exhaustive type/merge, int64, finite-
  number, UTF-8 byte, array, null/default, wildcard, order, and cancellation
  implementation and evidence;
- GP4-G08 is closed by the owner-projected registry and shared private runtime
  facts;
- GP4-G09 is closed by removal of dead interfaces, options, hooks,
  constructors, adapters, and the final three unused private helpers;
- GP4-G10 is closed by mandatory transactional job/lease reconciliation,
  exact postconditions, and readiness gating;
- GP4-G11 is closed by the non-empty golden and routed semantic, identity,
  persistence, failure, browser, broad-check, and release evidence; and
- GP4-G12 is closed by positive root, package, export, schema, constructor, and
  composition allowlists plus the terminal reviewed search.

The complete change set spans adopted owner/Core/ADR/traceability documents;
Graph, Network Flow, Extensions, and Recovery contracts and generated Go/
TypeScript projections; Graph engine, restore, PostgreSQL result/restore, and
tests; Network Flow state, graph, restore-source, routes, frontend adapter, and
browser tests; Recovery selection/codec and application assembly; Reporting,
server, and operator integration evidence; verification routing, schemas,
generator validation, and operator guidance. The former major-4 client, old
Graph Recovery schemas/fixtures/generations, root restore implementation, and
obsolete GP3 guide are deleted. Generated roots were changed only through
`make generate`.

Explicitly unchanged surfaces are the Graph v2 wire/result and object identity
algorithms for current valid inputs; default and temporal Network Flow graph-
view, projection, and digest fixtures; `/api/v1/.../network-flow` route paths;
Reporting exact-result references; public authorization and domain vocabulary;
all Graph/Network Flow persisted table shapes; authored SQL; migrations 00032,
00033, and 00034 and their hashes; and the next legal migration boundary. The
terminal diff contains no authored migration or query change.

Public compatibility is the adopted intentional hard cut. Graph resources use
the major-5 generated client and saved-view resource v3; semantic-query v1 and
Graph Recovery v2/v3 have no current reader, writer, translator, dispatcher,
fallback, or compatibility view. Internal Go callers moved atomically to
`ProjectV2`, `graphprojection/restore`, native-v2 structures, and mandatory
constructors. Stored unsupported declarations remain byte-identical and make
the profile unavailable; unsupported backups fail before mutation. Once state
4 commits, a major-4 binary is not a downgrade candidate. Rollback requires an
exact pre-rollout database/object-store backup and matching former binary in a
fresh replacement target, never an in-place state or migration-history edit.

The terminal exact-name scan returns no match in active owners, guides,
decisions, production, tests, contracts, generators, generated clients, or
immutable SQL. Permitted matches remain only in this frozen tracker: GP2's v2
restore and engine-facade history; GP3's v3 restore, schema/client, and guide
history; and GP4's baseline and removal rationale. No environment-only
limitation, skipped required target, waiver, unexplained generated diff, or
unresolved product failure remains.

Terminal verification passed:

- final formatting, generation, drift, generated-artifact policy, JSON shape,
  toolchain, migration, backend boundary, frontend type/unit/import, and first
  static floor at `.cartulary/test-results/20260830T021509Z-p1305515`,
  `.cartulary/test-results/20260830T021513Z-p1309603`,
  `.cartulary/test-results/20260830T021522Z-p1312323`,
  `.cartulary/test-results/20260830T021530Z-p1315259`,
  `.cartulary/test-results/20260830T021531Z-p1315670`,
  `.cartulary/test-results/20260830T021534Z-p1316082`,
  `.cartulary/test-results/20260830T021535Z-p1316456`,
  `.cartulary/test-results/20260830T021543Z-p1319401`,
  `.cartulary/test-results/20260830T021545Z-p1319790`,
  `.cartulary/test-results/20260830T021555Z-p1320273`, and
  `.cartulary/test-results/20260830T021558Z-p1320645`;
- terminal lint, Biome, scripts, Markdown, shell, and vulnerability gates at
  `.cartulary/test-results/20260830T021737Z-p1359535`,
  `.cartulary/test-results/20260830T021747Z-p1365006`,
  `.cartulary/test-results/20260830T021749Z-p1365449`,
  `.cartulary/test-results/20260830T021750Z-p1365827`,
  `.cartulary/test-results/20260830T021751Z-p1366701`, and
  `.cartulary/test-results/20260830T021753Z-p1367598`;
- corrected final format, generation/drift, targeted gosec, audit gosec, and
  lint at `.cartulary/test-results/20260830T021906Z-p1388002`,
  `.cartulary/test-results/20260830T021910Z-p1392091`,
  `.cartulary/test-results/20260830T021919Z-p1394847`,
  `.cartulary/test-results/20260830T021928Z-p1397883`,
  `.cartulary/test-results/20260830T021938Z-p1428264`, and
  `.cartulary/test-results/20260830T021944Z-p1445751`;
- `make test-fast` at
  `.cartulary/test-results/20260830T022000Z-p1451420` (442/442);
- the qualifying warm `make check` at
  `.cartulary/test-results/20260830T022115Z-p1464493` (672/672);
- `make agent-finalize
  RESULTS_DIR=.cartulary/test-results/20260830T022115Z-p1464493` at
  `.cartulary/test-results/20260830T022629Z-p1581402` (1/1);
- the terminal `make check` at
  `.cartulary/test-results/20260830T022651Z-p1584427` (672/672); and
- `make release-check` at
  `.cartulary/test-results/20260830T023202Z-p1698980` (838/838); and
- post-handoff tracker diff checks and Markdown lint at
  `.cartulary/test-results/20260830T024650Z-p1915746`.

Two terminal failures were change-related and fully superseded. Lint root
`.cartulary/test-results/20260830T021601Z-p1321085` found three unused private
helpers left after native-v2 admission replaced their compatibility-era call
paths; the functions were deleted and the final lint roots pass. Targeted
gosec root `.cartulary/test-results/20260830T021813Z-p1368434` found fragile
indexed access through three parallel one-entry Recovery expectation slices;
the generator now validates one exact current generation directly, and both
targeted and audit scans pass. Neither correction restores compatibility,
widens input, changes generated bytes, or changes persisted behavior.

Residual risk is only the intentional operational incompatibility already
adopted: installed v1 declarations and old Graph backups need separately
authorized remediation, and state 4 is forward-only. There is no remaining GP4
implementation, specification, validation, documentation, or handoff risk.

### 11.9 Verification matrix

Each implementation slice starts with the current task guide and the narrowest
owner rows. Expected final coverage is:

| Concern | Required public Make evidence |
| --- | --- |
| Direct owners | `make task-guide ROLE=module-author OWNER=<owner>` followed by unit and service-backed slices for `module.graphprojection`, `module.networkflow`, `module.recovery`, and `module.extensions` |
| Consumers and assembly | Routed slices for `module.reporting`, `app.server`, `app.operator`, `web.networkflow`, and database migrations when selected by the owner guides |
| Generation and authored shapes | `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make toolchain-drift`, and `make migration-drift` |
| Static and security | `make format`, `make lint`, `make lint-biome`, `make lint-scripts`, `make lint-markdown`, `make lint-shell`, `make go-vulncheck`, `make go-gosec-targeted`, and `make go-gosec-audit` |
| Frontend and browser | `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, and all applicable browser E2E, webserver, stateful, measurement, accessibility, and visual targets |
| Boundaries and release | Backend module-boundary checks selected by the task surface, `make agent-finalize`, `make check`, and `make release-check` |

GP4-S00 itself requires only `git diff --check`, `git diff --cached --check`,
and `make lint-markdown`, because it is a tracker-only change. Later slices
must record the exact command, result, run root, relatedness of any failure,
and superseding evidence.

The semantic and compatibility acceptance suite must include:

- every adopted field path, scalar and array projected type, and merge mode;
- valid and invalid type/merge pairs, nulls, defaults, wildcard mappings,
  canonical contributor order, repeated values, and empty inputs;
- signed-int64 minimum/maximum and overflow, finite-number boundaries, and
  multibyte UTF-8 values immediately below, at, and above every byte ceiling;
- deterministic output under source and mapping reorderings, prompt context
  cancellation, and absence of partial publication on error;
- exact current Network Flow default and temporal projection IDs and digests;
- state 3→4 compatible success and incompatible atomic failure;
- current Recovery v4 success, injected rollback, mandatory reconciliation,
  and pre-mutation rejection of v2/v3 Graph artifacts; and
- result collision, read, traversal, lease, cleanup, transaction, lock-order,
  retry, and multi-instance persistence behavior.

### 11.10 Binary completion criteria

GP4 is complete only when all of the following are true:

1. Adopted owners and machine projections agree on Graph 2.2.0, Network Flow
   major 5/state 4, semantic-query v2 only, and Recovery v4 only.
2. No active runtime, generated client, generator branch, source registry,
   dispatcher, fixture, or test helper accepts semantic-query v1 or Graph
   Recovery v2/v3.
3. State 3 v2-only profiles advance byte-for-byte; a v1 declaration blocks the
   transition without data or state-version mutation.
4. Unsupported old Graph backups fail before any target mutation and cannot
   resolve through a historical fallback.
5. The Graph root is a pure deterministic package; restore ownership resides
   in its dedicated subpackage and application assembly supplies all required
   dependencies explicitly.
6. `ProjectV2` is the only engine entrypoint, and the old constructor, method
   facade, config/filter adapters, duplicate fields, dead interfaces, optional
   reconciler, hooks, and stale limit surface are absent.
7. The complete adopted v2 semantic matrix passes, including signed-int64,
   finite-number, UTF-8 byte, array, merge, ordering, and cancellation cases.
8. Existing valid-v2 Network Flow projection and graph-view identities remain
   exact, as do Reporting result references and traversal capability.
9. Result publication, read, traversal, lease, cleanup, and v4 restore behavior
   pass owner-routed unit and service-backed evidence without a DDL or migration
   change.
10. Positive boundary tests pass, old-name searches contain only reviewed
    immutable SQL or frozen tracker history, generated outputs are clean, and
    no active guide promises removed support.
11. Every required final owner, generation, drift, security, frontend, browser,
    boundary, finalizer, broad check, and release target has terminal passing
    evidence recorded in this tracker.

### 11.11 GP4 current next action

GP4-S00 through GP4-S10 are `DONE`. All twelve gaps are `CLOSED`; focused,
service-backed, semantic, persistence, frontend, browser, static, security,
finalizer, repository-wide, and release evidence is terminal and passing.

Current next action: none.
