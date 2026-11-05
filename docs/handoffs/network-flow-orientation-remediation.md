# Orientation-aware Network Flow implementation handoff

Status: the three remaining Network Flow specification/validation findings and
the graph algorithm naming inconsistency are closed. Commit-scoped qualification,
all five complete owner audits and human specification-to-projection approval
remain recorded for their original frozen source. The naming follow-up has fresh
focused verification on its successor source, recorded in the final section.
The changes are ready to commit within that scope. Earlier orientation and
Imports amendment evidence remains historical. Full release/package
qualification, deployment, database reset, historical-data conversion and
committing are outside this execution's scope.

## Adopted semantics and compatibility

[Network Flow 9.0.0](../network-flow-activity-nlspec.md) owns neutral canonical
endpoint A/B tuples, explicit direction, ordinal-bound derivation evidence,
queries, indicators, aggregation and current-state admission. Cisco Flow Search
uses exporter-reported client/server orientation; its counters are conversation
totals. Missing and incomplete roles remain observations with unknown direction.
No direction is inferred from endpoint position, ports, counters or protocol.

Public major 9 and durable/minimum state 6 form a clean boundary. The supported
profiles are `cisco_sna_flow_search_csv_v2` and `cisco_sna_netflow_csv_v2`.
The original source/destination profile establishes declared direction without
asserting client/server roles. Generic Imports orchestration and Graph
Projection's v2 engine protocol retain their existing contracts. Coordinated
Network Flow recovery and source-consumer bindings have successor versions.

The authored [contract index](../../contracts/network-flow/index.json) records
the exact current public schemas. Its major changes include mapping registry
v3, row/query/diagnostic/indicator formats v2, graph semantic query and results
v3, saved graph formats v5, presentation v3 and frontend entrypoints v9.
Extensions 0.14.8 binds current state/codec admission; Graph Projection 2.2.2
records the coordinated source/recovery dependency without changing its engine
direction vocabulary. Database migration 38 establishes the neutral storage
constraints and the coordinated Reporting retained-input boundary.

[Reporting 1.5.0](../reporting-subsystem-nlspec.md) preserves graph direction in
its diagram model, Mermaid and manual SVG. Undirected and bidirectional edges
are rejected for sequence diagrams. Frozen admitted inputs remain independent
of subsequent source availability; existing disclosure and external-release
restrictions still apply. Core 00 records the coordinated owner adoption.

Auth and Jobs expose read-only presence probes for pre-migration admission.
Network Flow owns interpretation of its route/job families. Imports owns
retained mappings, plans and apply outcomes. Reporting owns frozen-input codec
admission. No owner reaches into another owner's private runtime storage to
repair or convert historical payloads.

## Implementation review map

| Workstream | Primary review entry points | Result |
| --- | --- | --- |
| W1 | [Network Flow specification](../network-flow-activity-nlspec.md), [contract index](../../contracts/network-flow/index.json), Core 00 coordinated adoption | One current neutral-endpoint contract; successor profiles and affected schema/algorithm versions. |
| W2 | [Mapping registry](../../contracts/network-flow/mapping-registry.v3.json), [CSV normalization](../../internal/modules/networkflow/csv_parser.go) | Closed ordinal-bound derivations, fingerprint coverage and source-input diagnostics. |
| W3 | [Endpoint derivation](../../internal/modules/networkflow/endpoint_direction.go), [storage](../../internal/modules/networkflow/store.go), [migration 38](../../db/migrations/00038_network_flow_orientation.sql) | Canonical tuples, explicit direction/evidence, digest-before-identity ordering and constrained durable state. |
| W4 | [Query predicates](../../internal/modules/networkflow/query_filter.go), [SQL queries](../../internal/modules/networkflow/query_sql.go), [indicator links](../../internal/modules/networkflow/indicator_link.go) | Neutral and established-direction predicates, stable pagination and endpoint-slot links. |
| W5 | [Semantic selectors](../../internal/modules/networkflow/edge_semantics.go), [graph composition](../../internal/modules/networkflow/graph.go), [contributor reads](../../internal/modules/networkflow/graph_source_read_store.go) | Shared grouping/identity/contributor semantics; unknown edges remain undirected. |
| W6 | [Observation presentation](../../apps/web/src/networkFlow/networkFlowObservation.ts), [exploration navigation](../../apps/web/src/networkFlow/networkFlowExplorationNavigation.ts), [presentation contract](../../contracts/network-flow/presentation.v3.json) | Typed endpoint/direction labels, accessible navigation, preview interpretation and conversation counters. |
| W7 | [Network Flow Reporting source](../../internal/modules/networkflow/reporting_graph_source.go), [diagram direction](../../internal/modules/reporting/diagram_direction.go), [graph redaction](../../internal/modules/reporting/graph_redaction.go) | Direction survives freezing, redaction and serialization; sequence selections reject unsupported topology. |
| W8 | [Migration admission](../../internal/modules/networkflow/migration_admission.go), [current-state admission](../../internal/modules/networkflow/orientation_admission.go), [graph restore](../../internal/modules/networkflow/graph_restore_source.go) | Read-only compatibility rejection, current codecs/bindings and transactional recovery. |
| W9 | [Independent fixture corpus](../../fixtures/network-flow-orientation/v1), [service oracle](../../internal/modules/networkflow/orientation_corpus_integration_test.go) | Original source bytes, explicit successor expectations and multiplicity-preserving integration checks. |

Network Flow acceptance criteria NF-AC-122 through NF-AC-127 collect the new
semantic obligations. Existing authorization, resource-bound, cancellation,
replay and disclosure requirements continue to apply. Tests and generated
projections do not replace human review of specification fidelity.

## Fixtures and independent expectations

Original CSV bytes remain unchanged, including the original Cisco example's
CRLF line endings. The Cisco example is NF-FIX-029 revision 7. The other affected
frozen fixtures have explicit successor revisions in their manifests and
executable consumers. Previous expectations remain repository history.

The incident corpus is in `fixtures/network-flow-orientation/v1`, with source
hashes, per-observation expectations, per-edge memberships and provenance in
JSON. Its independent calculation uses CSV parsing and IP ordering rather than
product normalization or captured application output. Revision-authoring Python
files are maintenance evidence, not runtime or verification inputs; they do not
read documentation. No test, generator or admission path consumes this handoff.

| Dataset | Observations | Subject-server roles | Unknown observations | Semantic edges |
| --- | ---: | ---: | ---: | ---: |
| Day 1 | 2,000 | 308 | 120 | 651 |
| Day 2 | 2,000 | 282 | 120 | 660 |
| Overlap | 500 | 77 | 44 | 324 |
| Original example | 1,200 | 182 | 30 | 875 |

The original example's 30 unknown observations form 26 undirected edges.
Unknown observation counts and edge counts are different quantities. Overlap
is deliberately retained; import does not deduplicate observations.

Across the three incident exports, the service-level oracle requires 4,500
observations, 3,903,637,992 bytes and 4,412,353 packets. It compares accepted
observations, default edge memberships, cross-table aggregation and paginated
contributors against independently authored JSON expectations.

The timeline, identities, hosts and aliases remain a separate fixture-adoption
workstream. This change does not certify their workflow coverage.

## Operational transition

1. Retain the previous matching application, browser assets, registry, database,
   object store, deployment configuration and required key material under the
   existing access controls. Verify historical access and a restorable backup
   before any operational transition.
2. Quiesce previous writers. Do not point the new release at historical Network
   Flow state to upgrade it. Pre-migration and claim admission reject incompatible
   state without repair or deletion, including retained artifacts without tables.
3. Provision fresh compatible database/object storage through the normal
   supported operational workflow. Publish matching server, browser assets and
   generated registries together. Initialize current state and verify readiness.
4. Re-import original sources using current profiles and newly approved mappings.
   Review preview sample counts, direction and diagnostics. Verify complete
   import totals after apply. Recreate saved graphs and indicator links explicitly;
   old row IDs, selectors, cursors and approvals are incompatible.
5. Keep historical outputs and workbooks available in their original environment.
   New imports create new evidence identities. Do not relabel frozen historical
   reports or silently replace old links with new ones.
6. Roll back by restoring or returning to the previous matching environment after
   quiescing new writers. Preserve any new-state evidence for later handling.
   Migration 38 has no downgrade path; do not run a reverse migration or delete
   new state to make an old binary start.

## Qualification and review gate

The orientation qualification baseline, before the Imports responsiveness
amendment, is
`sha256:b1737727e1eef82189658cd6b63588367672363c339aebfc6089e9e0ef15531d`
on dirty checkout `e6ca992e634147fd257f8a7462026a80053cd1d8`.
`orientation-finalize-15` completed successfully before this source was frozen.

`CARTULARY_TEST_RUN_ID=orientation-package-final-05 make package-inspect CHECK=all`
completed with native Make exit 0: 18 work units passed, with all 27 cleanup
steps completed. The inspected bundle is
[2350ebff…668b.tar](../../.cartulary/release-artifacts/packages/2350ebffef7cbb606caa4b0958801bc94789867a296769dd0fe9945d4b06668b.tar).
Its manifest SHA-256 is
`2350ebffef7cbb606caa4b0958801bc94789867a296769dd0fe9945d4b06668b`;
archive SHA-256 is
`b6a7787e1cf488f691ea52d007a8cb64783d6de027246dae4b502128703b8218`.
The [retained inspection receipt](../../.cartulary/test-results/orientation-package-final-05/package-release/artifacts/inspection.json)
binds the exact application, PostgreSQL, SeaweedFS and operator artifacts.
Installation without a build toolchain and rejection of an occupied destination
passed. The inventory disposition remains
`content_complete_review_unestablished`; package inspection does not establish
human inventory review or deployment approval.

Focused implementation runs are retained under `.cartulary/test-results` with
explicit `orientation-*` run IDs. Earlier-source results remain diagnostic
history and do not establish current-source acceptance. Finalization and source
freeze passed before the following portfolio; subsequent handoff edits affect
documentation only.

Human review must assess fidelity between adopted requirements and authored /
generated contracts, review fixture provenance and independent calculations,
and review the operational release package. This is separate from passing tests.
Deployment remains a separate operational action.

### Orientation baseline verification portfolio

Every invocation below reached native Make exit 0. All listed work units passed,
with no cancelled or skipped units, and every cleanup operation completed. Counts
are harness work units, not individual assertions. Each linked manifest retains
the exact command target, declared inputs, source identity and toolchain/system
digests; sibling `run-summary.json` and `cleanup-results.json` retain results.

For each row, the invocation used its run ID as the environment variable
`CARTULARY_TEST_RUN_ID`. For `test-slice`, `OWNER` and the complete comma-separated
`ROWS` argument are recorded verbatim in the linked manifest's `declared_inputs`.
The command form was `make test-slice OWNER=<OWNER> ROWS=<ROWS>`. This avoids
substituting a broad owner run for the exact, nonredundant complement selections.

| Retained run / exact input manifest | Make command or selection | Passed units | Completed cleanup |
| --- | --- | ---: | ---: |
| [orientation-finalize-15](../../.cartulary/test-results/orientation-finalize-15/run-manifest.json) | `make agent-finalize` | 1 | 10 |
| [orientation-check-final-02](../../.cartulary/test-results/orientation-check-final-02/run-manifest.json) | `make check` | 332 | 485 |
| [orientation-nf-consumers-final-01](../../.cartulary/test-results/orientation-nf-consumers-final-01/run-manifest.json) | `test-slice`, `module.networkflow`, 35 rows | 25 | 52 |
| [orientation-nf-capacity-final-03](../../.cartulary/test-results/orientation-nf-capacity-final-03/run-manifest.json) | `test-slice`, `module.networkflow`, 42 integration rows | 11 | 36 |
| [orientation-reporting-measurement-final-03](../../.cartulary/test-results/orientation-reporting-measurement-final-03/run-manifest.json) | `test-slice`, `module.reporting`, render/snapshot progress | 17 | 32 |
| [orientation-imports-measurement-final-02](../../.cartulary/test-results/orientation-imports-measurement-final-02/run-manifest.json) | `test-slice`, `module.imports`, CSV apply and XLSX apply/discovery | 19 | 36 |
| [orientation-auth-complement-final-02](../../.cartulary/test-results/orientation-auth-complement-final-02/run-manifest.json) | `test-slice`, `module.auth`, 42 rows | 24 | 38 |
| [orientation-extensions-complement-final-02](../../.cartulary/test-results/orientation-extensions-complement-final-02/run-manifest.json) | `test-slice`, `module.extensions`, 2 rows | 12 | 23 |
| [orientation-recovery-complement-final-02](../../.cartulary/test-results/orientation-recovery-complement-final-02/run-manifest.json) | `test-slice`, `module.recovery`, 1 row | 12 | 23 |
| [orientation-imports-complement-final-02](../../.cartulary/test-results/orientation-imports-complement-final-02/run-manifest.json) | `test-slice`, `module.imports`, 5 browser rows | 13 | 26 |
| [orientation-migration-complement-final-02](../../.cartulary/test-results/orientation-migration-complement-final-02/run-manifest.json) | `test-slice`, `module.database_migrations`, 9 rows | 5 | 12 |
| [orientation-app-server-complement-final-03](../../.cartulary/test-results/orientation-app-server-complement-final-03/run-manifest.json) | `test-slice`, `app.server`, 34 rows | 21 | 27 |
| [orientation-fixture-identity-final-01](../../.cartulary/test-results/orientation-fixture-identity-final-01/run-manifest.json) | `test-slice`, `harness.browser`, `harness.browser.unit.native_fixtures` | 1 | 10 |
| [orientation-visual-final-02](../../.cartulary/test-results/orientation-visual-final-02/run-manifest.json) | `make browser-e2e SUITE=visual` | 18 | 35 |
| [orientation-visual-final-03](../../.cartulary/test-results/orientation-visual-final-03/run-manifest.json) | `make browser-e2e SUITE=visual` | 18 | 34 |
| [orientation-package-final-05](../../.cartulary/test-results/orientation-package-final-05/run-manifest.json) | `make package-inspect CHECK=all` | 18 | 27 |

`make check` selected 1,693 catalog rows and used the managed cache: 78 hits,
99 misses and 155 bypassed units. It is not claimed as 332 uncached executions.
The Network Flow complements include bounded large-import integration, exact
fixture/contributor checks, browser workflows, accessibility, process behavior
and all four explicit Network Flow browser measurements. The full visual suite
was run twice to satisfy the golden-maintenance gate. The harness fixture-key
repair received its explicit native-fixtures test plus real Reporting/Imports
measurement coverage; no full `harness.browser` owner audit is claimed.

The shared compatibility identity for this portfolio includes toolchain digest
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and system digest
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.

### Orientation baseline complete-owner evidence audits

These 14 audits passed with native Make exit 0, no rejected evidence roots, and
all 10 cleanup operations completed per audit. They establish coverage of 545
catalog rows against the frozen source and current catalog/routing digests.
They do not establish human projection fidelity.

The exact command form was
`make test-evidence-audit OWNER=<owner> EVIDENCE_ROOTS_FILE=<manifest>`, with the
audit run ID in `CARTULARY_TEST_RUN_ID`. Each sibling `run-manifest.json` records
the original manifest path. The linked durable audit contains every required
partition and accepted run root/row list, so evidence selection does not depend
on retaining the temporary input-manifest files.

| Owner | Required rows | Retained audit |
| --- | ---: | --- |
| `module.networkflow` | 144 | [orientation-audit-nf-final-01](../../.cartulary/test-results/orientation-audit-nf-final-01/test-evidence-audit/audit.json) |
| `web.networkflow` | 66 | [orientation-audit-web-final-02](../../.cartulary/test-results/orientation-audit-web-final-02/test-evidence-audit/audit.json) |
| `module.reporting` | 45 | [orientation-audit-reporting-final-01](../../.cartulary/test-results/orientation-audit-reporting-final-01/test-evidence-audit/audit.json) |
| `module.extensions` | 44 | [orientation-audit-extensions-final-02](../../.cartulary/test-results/orientation-audit-extensions-final-02/test-evidence-audit/audit.json) |
| `module.recovery` | 44 | [orientation-audit-recovery-final-02](../../.cartulary/test-results/orientation-audit-recovery-final-02/test-evidence-audit/audit.json) |
| `module.auth` | 72 | [orientation-audit-auth-final-01](../../.cartulary/test-results/orientation-audit-auth-final-01/test-evidence-audit/audit.json) |
| `platform.jobs` | 19 | [orientation-audit-jobs-final-02](../../.cartulary/test-results/orientation-audit-jobs-final-02/test-evidence-audit/audit.json) |
| `module.database_migrations` | 19 | [orientation-audit-migration-final-02](../../.cartulary/test-results/orientation-audit-migration-final-02/test-evidence-audit/audit.json) |
| `app.migrate` | 2 | [orientation-audit-app-migrate-final-02](../../.cartulary/test-results/orientation-audit-app-migrate-final-02/test-evidence-audit/audit.json) |
| `app.server` | 50 | [orientation-audit-app-server-final-02](../../.cartulary/test-results/orientation-audit-app-server-final-02/test-evidence-audit/audit.json) |
| `module.graphprojection` | 11 | [orientation-audit-graph-final-02](../../.cartulary/test-results/orientation-audit-graph-final-02/test-evidence-audit/audit.json) |
| `platform.postgres` | 6 | [orientation-audit-postgres-final-02](../../.cartulary/test-results/orientation-audit-postgres-final-02/test-evidence-audit/audit.json) |
| `platform.telemetry` | 11 | [orientation-audit-telemetry-final-02](../../.cartulary/test-results/orientation-audit-telemetry-final-02/test-evidence-audit/audit.json) |
| `harness.generated_artifacts` | 12 | [orientation-audit-generated-final-03](../../.cartulary/test-results/orientation-audit-generated-final-03/test-evidence-audit/audit.json) |

For the orientation baseline, no complete `module.imports` audit was run: its
CSV-discovery measurement remained unqualified. The passing three-row measurement run must not be presented as
covering that fourth measurement. Full CI was not repeated after the final
source freeze; the nonredundant `make check`, complements, owner audits, explicit
visuals and package inspection above are the final-source portfolio. Earlier
CI results belong to older source and are not used for this acceptance claim.

### Supporting repairs found during qualification

- Large imports now write accepted observations through bounded `CopyFrom`
  batches in the existing owner transaction. Both a 128-row ceiling and a
  1-MiB estimated payload ceiling apply; an already-admitted oversized row is
  flushed alone. The original transaction deadline remains unchanged. Late
  invalid-row and SQL-constraint failures prove rollback after earlier batches
  have been written.
- Measurement fixture allocations derive a bounded opaque key from the exact
  builder/profile/snapshot tuple. The previous concatenation exceeded the
  retained service-scope schema's length limit for Reporting. Tuple identity,
  consumer sharing and distinct-input separation remain tested; no schema limit
  was widened.
- CSV apply measurement preparation now expects target-neutral discovery's
  explicit absence of tabular metadata. XLSX still verifies its discovered
  20,000-row, two-column geometry. The fixed input populations are unchanged.
- Qualification also repaired stale fixture/admission expectations, current
  SQL-owner grants, a test transaction cleanup race, telemetry test-support
  placement, and deployable-shape assertions for the current private bootstrap
  asset format. These repairs retain authorization, transaction, disclosure and
  runtime-boundary checks.
- The packaged-image disabled-cryptography probe now recognizes the server's
  sanitized startup-boundary message instead of requiring raw error text. It
  checks exact per-binary rejection exits as well, so timeouts or unrelated
  process failures cannot qualify. The initial package run passed archive,
  deployable-shape and operational-recovery inspection but failed the two
  consumers of this stale shared assertion; all cleanup completed.
- The next standup check exposed a real startup-diagnostic gap: Core 04 requires
  typed PostgreSQL admission failures to map through
  `invalid_deployment_config`, but the server emitted only a generic runtime
  setup log. Server startup now translates the closed PostgreSQL reason into
  a structured deployment diagnostic, discarding upstream text. A unit case
  wraps the typed failure in private data and proves that only the safe
  classification survives; the package probe checks the structured wrong-role
  certificate rejection, exact exit and unchanged cryptographic format.
  Server startup also preserves the migration owner's fixed incompatible-state
  diagnostic while discarding any wrapping cause. The package probes require
  exact exits for both incompatible and missing crypto identity and continue to
  compare database counts and filesystem digests before and after rejection.

The final-source server complement's second invocation failed infrastructure
preflight while independent managed builds overlapped. Go attempted to embed
`.browser-package-1779431750` while another asset publication removed that
temporary file. The 13 affected process selectors did not execute; the run's
unit and browser partitions passed and all 30 cleanup steps completed. Its root
is `orientation-app-server-complement-final-02`, native Make exit 2 / child exit
3, normalized `infra/preflight_error`. This was an execution-order failure;
no product assertion failed. Subsequent managed build/test invocations are
serialized. The fresh standalone `orientation-app-server-complement-final-03`
passed with native Make exit 0. No deadline, assertion or source was changed to
address the overlap; the failed root remains excluded from acceptance evidence.

### Outstanding generic Imports qualification conflict

The [earlier diagnostic run](../../.cartulary/test-results/orientation-imports-measurement-final-01/run-summary.json)
`orientation-imports-measurement-final-01` failed CSV discovery's existing
`perf.import_csv_discovery.visible_with_cancel.v1` criterion. Retained network
evidence shows the exact admitted Job running and then succeeding in 13 ms.
The criterion requires a queued/running Job with enabled cancellation while
the editor accepts text across two consecutive frames; successful early
completion explicitly does not qualify. No artificial delay, enlarged fixture,
response hold, retry or terminal-state substitution was introduced.

That run used earlier source `95ba28a4…`, exited Make 2 / child 10, and completed
all 40 cleanup operations. Its CSV-apply setup assertion also failed; the
target-neutral metadata correction described above resolved that assertion.
On final source, CSV apply and XLSX apply/discovery passed in
`orientation-imports-measurement-final-02`, with native Make exit 0 and all 36
cleanup operations completed. CSV discovery was deliberately not rerun unchanged
to seek a lucky pass. The 13-ms observation is diagnostic evidence from the
earlier source, not a current-source measurement claim.

This is a conflict between the existing generic performance criterion and the
current target-neutral CSV discovery workload, separate from orientation
normalization. On 2026-11-05, the user selected "Keep it as a documented release
blocker" rather than amending the generic Imports criterion to accept successful
completion plus responsive editing. At that point the criterion remained
unchanged and no waiver or amendment was authorized. The user subsequently
authorized the progress-or-success amendment for all four Imports measurements,
with scoped Imports qualification. The original decision remains historical;
the amendment and its current evidence are recorded separately below.

### Diagnostic history and remaining handoff actions

The failed package invocations below all used earlier executable source,
returned native Make exit 2 / child 10, and completed all 27 cleanup operations.
They are excluded from acceptance. Each run root retains its summary, exact
inputs and probe artifacts under `.cartulary/test-results/<run-id>`.

| Run ID | Source prefix | Failure and disposition |
| --- | --- | --- |
| `orientation-package-final-01` | `c9dd7ddf` | Disabled-crypto probe expected raw text; corrected to the safe boundary message and exact exits. |
| `orientation-package-final-02` | `072823f9` | Server lost typed PostgreSQL admission classification; added safe closed-reason translation and regression coverage. |
| `orientation-package-final-03` | `f9db0bc3` | Probe parsed only the first line of formatted JSON; now validates the complete diagnostic suffix. |
| `orientation-package-final-04` | `b81c1599` | Server lost the fixed incompatible-crypto-state diagnostic; preserves only the owner sentinel, never its wrapping cause. |

The current-source package run `orientation-package-final-05` passed all checks.
Its safe rejection probes retain exact exit, unchanged state and confidentiality
assertions. The separate build-overlap failure and Imports criterion conflict
are documented above with their distinct dispositions.

The later authorized amendment requires matching measurement evidence and a
complete Imports owner audit before its performance evidence gate can close.
It does not establish release readiness for the changed source. Any executable change
requires regenerated projections where applicable, finalization and a fresh
compatible evidence portfolio. Do not reuse this source's evidence as proof for
a changed implementation.

Reviewers must confirm specification/projection fidelity, independent fixture
provenance/calculation, and release-inventory disposition. Operations must verify
historical-environment retention and backups before the fresh-state transition.
No historical database conversion, deployment or automatic cleanup is part of
this implementation. Other IR fixture adoption remains separate.

## Visual refresh assessment

The adopted endpoint and direction presentation authorized the golden refresh.
`orientation-visual-baseline-01` retained a complete ordinary baseline: 48 rows,
47 passed, with only the Network Analysis visual row failing on its expected
presentation changes. Reconciliation accounted for all 255 captures and goldens,
with no missing or orphaned entries. `orientation-visual-update-01` completed
successfully through the transactional Make updater.

The affected row is
`module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`,
using fixture `visual.fixture.claimed_network_analysis_workspace_states`.
The renderer profile, viewport declarations, normalization, masks, zoom and
capture scopes were not changed. Eight PNGs under
`apps/web/e2e/visual-goldens/` changed; the empty workspace remained unchanged.

| Golden suffix after `network-flow-analysis-` | Assessed change |
| --- | --- |
| `accepted-inspector-linux.png` | Neutral endpoint headers and inspector; counter explanation; preserved usable grid area. |
| `compact-saved-graphs-linux.png` | Current source profile metadata and usable compact saved-graph controls. |
| `delete-dialog-linux.png` | Saved edge label states declared direction; existing deletion confirmation remains readable. |
| `filtered-empty-grid-linux.png` | Endpoint A/B, ports and direction headers; empty-grid action remains visible. |
| `graph-contributors-linux.png` | Declared direction label and neutral endpoint-link actions agree. |
| `narrow-query-controls-linux.png` | Query controls wrap; endpoint and counter explanation remain readable. |
| `rejected-diagnostics-linux.png` | Invalid IP diagnostic identifies the derivation input, not a normalized endpoint slot. |
| `saved-graph-result-linux.png` | Saved graph uses the same explicit direction label as unsaved exploration. |

All eight changed PNGs were imported into the artifacts-mode UI-review harness,
verified against bundle byte counts and SHA-256 digests, and inspected at native
resolution on 2026-11-05. The review session stopped successfully with complete
cleanup. These image imports provided no DOM or axe observations; the interactive
HTML report was not inspected. Browser accessibility tests provide separate
behavioral evidence. Two fresh ordinary validations on final executable source
`b1737727…` passed: `orientation-visual-final-02` and
`orientation-visual-final-03`, each using `make browser-e2e SUITE=visual` with its
run ID in `CARTULARY_TEST_RUN_ID`. Both native Make exits were 0. Each run
accounted for all 255 captures and committed goldens, with zero missing,
orphaned or ambiguous mappings. All cleanup completed: 35 steps in the first
run and 34 in the second. The earlier
`orientation-visual-final-01` also passed but belongs to a prior source snapshot
and is excluded from final-source owner acceptance.


## Imports progress-or-success amendment — 2026-11-05

The maintainer subsequently authorized the amendment for all four Imports
measurements and selected scoped Imports qualification, without repeating the
complete Network Flow release portfolio. Core 00 §1.5.2 records the adoption;
Core 04 REQ-04-158, AC-027 and the Imports portion of AC-046 own the behavior.

The new predicates are `perf.import_csv_discovery.responsive.v2`,
`perf.import_xlsx_discovery.responsive.v2`,
`perf.import_csv_apply.responsive.v2` and
`perf.import_xlsx_apply.responsive.v2`. Discovery/apply performance contracts,
contract schemas and nested proofs are v2. Current readers reject retired v1
Imports proofs and predicates. Common observation v4 and summary/aggregate v5
envelope formats remain unchanged, with current nested proof registration.

Each observation records `stop_kind=progress_with_cancel` with a queued/running
Job and enabled cancellation, or `stop_kind=succeeded` with a visibly successful,
noncancelable Job and no enabled Cancel control. Both require the same focused,
editable Timeline binding and prescribed uncommitted edit across two consecutive
frames. A branch transition restarts that frame count. Success checks the actual
rendered status text; transport success or a status attribute alone is insufficient.
The measured interval remains at or below 1,000 ms from product-action acceptance,
including request/preflight work. Fixture inputs, populations, background traffic,
sampling and the bounded diagnostic timeout are unchanged.

The shared observer permits the success policy only for Imports. Reporting,
Evidence and Reference Pack continue to require cancelable progress. Independent
functional cancellation obligations remain in force. The new proof discriminant
is carried through collection, schema admission, finalization and retained
artifact validation; synthetic IDs remain restricted to the existing closed
proof locations. No raw transaction IDs, payloads or edited text are retained.

The only product markup change marks the existing Import Assistant status text
for visible observation. Displayed copy, layout, runtime APIs, storage and
Network Flow direction semantics are unchanged. No visual-golden refresh,
database migration, historical conversion or deployment was performed.

### Amendment source and qualification

Frozen executable source:
`sha256:812ff0258fa3167bfdb00b7c1db8a0e7070539670994392c9a14c89688c9232c`
on dirty checkout `e6ca992e634147fd257f8a7462026a80053cd1d8`.
`imports-responsive-finalize-02` passed with native Make exit 0 and no generated
changes. Retained-run maintenance was skipped because no compatible successful
full check root was supplied; no older source's result was substituted.

All automated amendment qualification passed on this source. The complete
`module.imports` audit accepted all 50 required catalog rows with no rejected
roots. Each command below reached native Make exit 0 and completed every cleanup
operation. No work unit was failed, cancelled or skipped in these accepted runs.
Counts are harness work units, not individual assertions.

Each command used the linked run ID in the `CARTULARY_TEST_RUN_ID` environment
variable. For `test-slice`, the manifest retains the exact `OWNER` and `ROWS`
arguments under `declared_inputs`; the command is `make test-slice` with those
arguments. The audit manifest's original path is retained in its run manifest;
the audit artifact records all accepted roots and row lists independently of
that temporary file.

| Retained run / exact input manifest | Command or selection | Passed units | Completed cleanup |
| --- | --- | ---: | ---: |
| [imports-responsive-finalize-02](../../.cartulary/test-results/imports-responsive-finalize-02/run-manifest.json) | `make agent-finalize` | 1 | 10 |
| [imports-responsive-check-01](../../.cartulary/test-results/imports-responsive-check-01/run-manifest.json) | `make check` | 332 | 563 |
| [imports-responsive-browser-01](../../.cartulary/test-results/imports-responsive-browser-01/run-manifest.json) | `test-slice`, `module.imports`, five missing browser rows | 13 | 26 |
| [imports-responsive-proof-final-01](../../.cartulary/test-results/imports-responsive-proof-final-01/run-manifest.json) | `test-slice`, `harness.browser`, `harness.browser.behavior.artifact_assurance` | 2 | 11 |
| [imports-responsive-measurement-01](../../.cartulary/test-results/imports-responsive-measurement-01/run-manifest.json) | `test-slice`, `module.imports`, all four measurements | 21 | 40 |
| [imports-responsive-audit-02](../../.cartulary/test-results/imports-responsive-audit-02/run-manifest.json) | `make test-evidence-audit OWNER=module.imports EVIDENCE_ROOTS_FILE=/tmp/imports-responsive-evidence.json` | 1 | 10 |

`make check` selected 1,693 catalog rows and reported 155 cache bypasses, 177
misses and zero hits. Its coverage includes the Imports observer regressions and
strict Reporting, Evidence and Reference Pack observer regressions. The separate
artifact-assurance row tests v2 proof collection, threshold admission, retained
finalization and redaction. Earlier focused runs and generated/schema checks are
diagnostic/development evidence; these compatible final-source roots establish
the acceptance portfolio.

The [complete Imports audit](../../.cartulary/test-results/imports-responsive-audit-02/test-evidence-audit/audit.json)
binds source `812ff025…`, catalog digest
`sha256:8d8fa05c9468b557057760dea4821bedc9d553990a435aefc254433d7eaf0f4d`
and routing digest
`sha256:0cc024cfe83a641048134573136d5eb575870adc4e3de7c2102f11e9d62cb66f`.
Toolchain and system digests remain those recorded in the orientation baseline.

### Real measurement results

All four finalized summaries are qualified, with one observation each, zero
scheduler overlap and the unchanged 1,000-ms threshold. Values below are rounded
to one decimal millisecond; the linked aggregate and observations retain the
original precision. These measure action acceptance through the second qualifying
frame with the editor active, not server Job duration.

| Predicate suffix after `perf.import_` | Elapsed ms | Stop branch | Observed Job status | Cancel enabled |
| --- | ---: | --- | --- | --- |
| `csv_discovery.responsive.v2` | 212.7 | `succeeded` | `succeeded` | false |
| `xlsx_discovery.responsive.v2` | 361.4 | `progress_with_cancel` | `queued` | true |
| `csv_apply.responsive.v2` | 250.9 | `progress_with_cancel` | `running` | true |
| `xlsx_apply.responsive.v2` | 195.8 | `progress_with_cancel` | `queued` | true |

The [measurement aggregate](../../.cartulary/test-results/imports-responsive-measurement-01/browser-e2e-measurement/frontend-measurement-aggregate.json)
and its referenced current summaries/observations retain branch, timing, exact
synthetic identity, fixture and cleanup evidence. No branch was forced; no action
retry, warm-up, worker delay or response hold was introduced.

Agent source review traced the adopted branches and unchanged bounds through
both authored contracts, closed proof variants, generated consumers, the shared
observer, serialization, profile admission, collection and retained validation.
Active v1 Imports predicate/proof bindings are retired; predecessor identities
appear only in rejection tests and history. This review and the passing tests do
not establish human projection-fidelity review. The performance evidence gap is
resolved; final amendment acceptance remains pending that human review. Full
Network Flow release qualification, package reinspection and visual qualification
on this changed source were deliberately outside the selected follow-up scope.

### Amendment diagnostic history

- `imports-responsive-harness-01` failed its browser artifact-assurance row with
  native Make exit 2 / child exit 10. The collector's synthetic-ID allowlist
  still named the old v1 predicates, so it rejected a v2 proof's permitted
  `record_id`. The correction updates only the exact Imports predicate bindings;
  private-field and wrong-location rejection remain enforced. All 22 cleanup
  operations completed. The focused `imports-responsive-proof-02` rerun passed
  after the repair; final-source evidence is recorded separately.
- `imports-responsive-finalize-01` exited Make 2 / child 1 during JSON-shape
  validation because the collector edit made the generated topology digest stale.
  It performed no generated mutations and completed all 10 cleanup operations.
  `make generate` under `imports-responsive-generate-02` refreshed the projections;
  `imports-responsive-finalize-02` then passed. No validator was relaxed.

- `imports-responsive-audit-01` exited Make 2 / child 2 because the input manifest's
  targets were not ASCII-sorted. All 10 cleanup operations completed. Sorting the
  same target/root entries corrected the invocation input; the fresh
  `imports-responsive-audit-02` audit passed without source changes or test reruns.

These roots are under `.cartulary/test-results/<run-id>`. Failed or earlier-source
runs remain diagnostic history, not current acceptance evidence. The prior
Network Flow audits, visuals and package receipt above remain bound to their
original source; this amendment does not establish release readiness for the
changed source.

## Validation-gap completion — 2026-11-05

The maintainer authorized completing the three review findings and qualifying
the resulting source for commit. [Core 00 §1.5.3](../spec/00_document_set_status_and_precedence.md)
records the adopted correction. Public major 9, durable state 6 and minimum
state 6 remain unchanged. Migration 00038 is completed directly because it is
uncommitted and unreleased; its catalog/history projections are regenerated.
No compatibility migration, retained-state rewrite or legacy reader is added.

### Closing implementation and review map

| Finding | Completed correction | Review and regression inputs |
| --- | --- | --- |
| Restore admitted noncanonical endpoint text | NF-REQ-035 now defines exact IPv4/IPv6 spelling, including mapped IPv6 and first-longest zero compression. Import normalizes valid source spelling; retained writes, claim and restore reject alternate spelling. One Go endpoint helper serves write/claim validation; one SQL endpoint-pair predicate serves CHECK constraints and the existing owner restore invariant. | [Migration 38](../../db/migrations/00038_network_flow_orientation.sql), [endpoint helper](../../internal/modules/networkflow/endpoint_direction.go), [independent vectors](../../fixtures/network-flow-validation/v1/endpoints.json), [storage tests](../../internal/modules/networkflow/canonical_endpoint_integration_test.go), [atomic restore tests](../../internal/modules/recovery/postgres_integrity_integration_test.go). |
| Unrelated errors suppressed orientation conflicts | Scalar results retain operand validity; timestamp and role conflicts run independently before a single rejected-row decision. Normalization constructs only accepted endpoints/evidence. One role-conflict predicate also protects defensive derivation. | NF-REQ-080e and §12.4; [CSV validation](../../internal/modules/networkflow/csv_parser.go), [combined-error tests](../../internal/modules/networkflow/orientation_diagnostics_test.go). |
| Normative diagnostic identity still named v1 | Tables 6-B/12-E and NF-REQ-113a now specify the existing v2 algorithm and exact NUL-delimited transcript. Related domain query/bucket versions are reconciled. Runtime identities are unchanged. | [Independent v2/v1 known answers](../../fixtures/network-flow-validation/v1/diagnostic-ids.json), [identity admission tests](../../internal/modules/networkflow/validation_vectors_test.go). |

IPv4 canonical validation combines an explicit four-octet, no-leading-zero
grammar with PostgreSQL address/range validation. IPv6 is formatted from parsed
address bytes under the owner rules, independent of PostgreSQL display text.
Malformed and null inputs return false. Constraints require `IS TRUE`; restore
rejects `IS NOT TRUE`. Functions retain invoker privileges, fixed search paths
and narrowly granted execution. Recovery acquires no normalization or live
profile dependency. The valid restore test applies claim validators only after
structural restore has committed.

The endpoint vectors are independently authored accepted/rejected cases shared
by Go and SQL. Diagnostic hashes were independently calculated from literal
transcripts; tests consume the frozen answers rather than generating expected
IDs through production helpers. Original incident CSV bytes and accepted-row
expectations are unchanged. New fixtures and executable expectations are outside
Markdown; no executable consumer depends on this handoff or the specification.

### Qualification repair and diagnostic history

Two broader checks exposed the same 100,001-row Imports/Network Flow apply
failure. The public error was `owner_apply_validation_failed`. A test-only
decorator using the existing application composition seam now retains only
closed timeout/cancellation booleans, without source values or raw database
errors. The subsequent service-backed diagnostic run established
`transaction_timeout=true`: the existing 30-second owner transaction expired.
The smaller canonical-state and diagnostic regressions passed throughout.

Removing per-address exception subtransactions alone did not resolve the broad
failure. The final SQL implementation also removes repeated aggregate queries
from direction-evidence object-shape checks and replaces array-query iteration
with an index loop bounded by the derivation's four or six inputs. Closed shape,
types, hashes, attribution, roles and ordinal uniqueness remain enforced. An
explicit `bigint` input check aligns retained ordinals with Go's `int64` bound.
[Storage regressions](../../internal/modules/networkflow/direction_evidence_integration_test.go)
cover missing/extra members, malformed inputs, duplicate and overflowing
ordinals. No transaction deadline, fixture size, measurement threshold, golden,
worker duration or harness resource budget was relaxed.

The isolated 100,001-row test passed in 22.81 seconds after this change; earlier
isolated observations were 26.02–29.35 seconds. These are diagnostic test
durations, not a new performance acceptance contract or controlled benchmark.
The large import also passed in `nf-validation-check-03` under the full check
workload. That invocation had one unrelated catalog-projection failure and is
not passing acceptance evidence.

The catalog failure exposed raw SQL regular-expression scans treating the
comment words `constraint path` as a real constraint. Generation and static
catalog validation now share a small owner-local lexical projection that masks
comments, quoted values and routine bodies while preserving source offsets.
Migration policy inspection remains separate and continues inspecting routine
bodies. The boundary regression covers nested comments, escaped quotes, tagged
routine bodies, Unicode offsets and malformed tokens. Fresh-database catalog
parity passes. No comment was reworded to conceal the defect.

The failed development runs remain retained: `nf-validation-check-01` and
`nf-validation-check-02` failed the large import; the diagnostic service run
`nf-validation-imports-services-diagnostics-01` established its transaction
timeout. `nf-validation-check-03` failed catalog parity. The first lexical test
exposed an escape-string prefix mismatch, corrected before its second passing
run. Schema check 04 exposed the second raw SQL scan; check 05 then rejected
stale topology inputs, requiring regeneration. Each failure was investigated
before the next invocation. All these failed runs exited Make 2 (harness 10)
and completed cleanup; they do not fill current-source acceptance obligations.

### Pre-follow-up source and acceptance disposition

Finalization `nf-validation-finalize-04` passed with native Make exit 0. The
executable source is dirty at commit
`e6ca992e634147fd257f8a7462026a80053cd1d8`, with harness source digest
`sha256:fc975641e456db417dcf6d436314887692209aec00d24d5babf7033d5c191c85`.
Commit qualification passed for that source. All three findings have closing
regression evidence, the selected portfolio and all five complete owner audits
passed, and the human review below is approved. Recommend committing this
reviewed source. No commit or deployment was performed. Full Network Flow
release/package qualification on this source remains outstanding by scope.

Human projection-fidelity review is **approved**. On 2026-11-05 the maintainer
answered “Reviewed and approved” for frozen executable source
`fc975641e456db417dcf6d436314887692209aec00d24d5babf7033d5c191c85`, covering
the canonical endpoint rules, operand-local diagnostics, diagnostic-ID v2
transcript and generated migration/catalog bindings linked in the review map.
This is a recorded human disposition, separate from adoption, generation and
automated test results. Executable changes after this source require a fresh
compatibility and review assessment.

Existing rejected-row diagnostics and receipts remain immutable. To obtain the
additional required conflict diagnostics, start a fresh preview/import from the
original source in compatible state. Do not repair stored rows, diagnostics or
backups. Historical access/re-import and rollback retain the operational
boundaries above: rollback uses the previous matching application/database
environment, never a downgrade of the new database.


### Retained commit-qualification evidence for the pre-follow-up source

All provider invocations below passed with native Make exit **0**, no failed or
skipped work units, and every cleanup step completed. The five audits accepted
only the frozen source above; none combined historical or incompatible roots.
Work units are scheduler units, not individual test counts. Exact selections
are retained in each root's `run-manifest.json` and reproduced below.

| Run and retained summary | Scope | Passed work units | Completed cleanup steps |
| --- | --- | ---: | ---: |
| [nf-validation-check-04](../../.cartulary/test-results/nf-validation-check-04/run-summary.json) | make check | 332 | 563 |
| [nf-validation-networkflow-backend-01](../../.cartulary/test-results/nf-validation-networkflow-backend-01/run-summary.json) | Network Flow backend partitions, 49 rows | 19 | 46 |
| [nf-validation-networkflow-browser-01](../../.cartulary/test-results/nf-validation-networkflow-browser-01/run-summary.json) | Network Flow browser partitions, 31 rows | 25 | 52 |
| [nf-validation-recovery-coverage-01](../../.cartulary/test-results/nf-validation-recovery-coverage-01/run-summary.json) | Recovery browser, 1 row | 12 | 23 |
| [nf-validation-migrations-coverage-01](../../.cartulary/test-results/nf-validation-migrations-coverage-01/run-summary.json) | Migration unit partition, 9 rows | 5 | 12 |
| [nf-validation-extensions-coverage-01](../../.cartulary/test-results/nf-validation-extensions-coverage-01/run-summary.json) | Extensions stateful browser, 2 rows | 12 | 23 |
| [nf-validation-imports-coverage-01](../../.cartulary/test-results/nf-validation-imports-coverage-01/run-summary.json) | Imports browser/measurements, 9 rows | 24 | 47 |
| [nf-validation-boundaries-final-01](../../.cartulary/test-results/nf-validation-boundaries-final-01/run-summary.json) | Shared harness boundary regression | 4 | 13 |

The evidence-audit contract requires a complete owner/target partition in one
provider. Consequently the supplemental portfolio repeats 43 Network Flow
integration rows and eight migration unit rows already covered by `make check`.
This is the minimum partition completion needed by the existing audit contract;
all other supplemental product selections add missing obligations. Network Flow
backend/browser selections were split to respect the 4,096-character `ROWS`
bound. The source, workload, audit contract and golden expectations were not
changed to avoid these obligations.

| Complete owner audit | Accepted rows | Accepted target partitions | Rejected evidence |
| --- | ---: | ---: | ---: |
| [module.networkflow](../../.cartulary/test-results/nf-validation-audit-networkflow-01/test-evidence-audit/audit.json) | 149 | 11 | 0 |
| [module.recovery](../../.cartulary/test-results/nf-validation-audit-recovery-01/test-evidence-audit/audit.json) | 44 | 5 | 0 |
| [module.database_migrations](../../.cartulary/test-results/nf-validation-audit-migrations-01/test-evidence-audit/audit.json) | 19 | 2 | 0 |
| [module.extensions](../../.cartulary/test-results/nf-validation-audit-extensions-01/test-evidence-audit/audit.json) | 44 | 4 | 0 |
| [module.imports](../../.cartulary/test-results/nf-validation-audit-imports-01/test-evidence-audit/audit.json) | 50 | 5 | 0 |

All five audit commands exited 0 and each completed all 10 cleanup steps. The
union closes **306 required owner rows**, including the newly registered cases.
The explicit Network Flow visual row passed against existing goldens; no golden
refresh was performed. The existing Cisco corpus passed its independent source
hash, per-observation, graph/contributor and total expectations unchanged.

Closing regression evidence:

- Canonical endpoint parity and storage rejection: [Go vectors](../../.cartulary/test-results/nf-validation-check-04/rows/module.networkflow.unit.canonical_endpoint_vectors.json), [SQL/direct writes](../../.cartulary/test-results/nf-validation-check-04/rows/module.networkflow.integration.canonical_endpoint_storage.json), and [atomic restore/claim validation](../../.cartulary/test-results/nf-validation-check-04/rows/module.recovery.integration.reset_preserves_crypto_format_identity.json).
- Independent cross-field diagnostics, preview/two-pass apply equivalence and retention prefixes: [combined-error row](../../.cartulary/test-results/nf-validation-check-04/rows/module.networkflow.unit.orientation_diagnostic_dependencies.json).
- Exact v2 diagnostic IDs and predecessor rejection: [known-answer row](../../.cartulary/test-results/nf-validation-check-04/rows/module.networkflow.unit.diagnostic_id_v2_vectors.json).
- Strict direction-evidence validation after the performance repair: [SQL shape and ordinal cases](../../.cartulary/test-results/nf-validation-check-04/rows/module.networkflow.integration.direction_evidence_storage.json).
- Catalog lexical boundary and fresh database parity: [shared harness regression](../../.cartulary/test-results/nf-validation-boundaries-final-01/rows/harness.command_surface.behavior.owner_boundaries.json) and [migration catalog row](../../.cartulary/test-results/nf-validation-check-04/rows/module.database_migrations.integration.production_ddl_v3_recurrence.json).

The current [Imports aggregate](../../.cartulary/test-results/nf-validation-imports-coverage-01/browser-e2e-measurement/frontend-measurement-aggregate.json)
is qualified with zero scheduler overlap, zero warm-ups, one observation per
predicate, the prescribed 20,000-row populations and unchanged 1,000-ms limits:

| Predicate | Elapsed ms | Observed branch | Job status | Cancel enabled |
| --- | ---: | --- | --- | --- |
| `perf.import_csv_apply.responsive.v2` | 236.2 | `progress_with_cancel` | `running` | `true` |
| `perf.import_csv_discovery.responsive.v2` | 186.4 | `succeeded` | `succeeded` | `false` |
| `perf.import_xlsx_apply.responsive.v2` | 276.7 | `progress_with_cancel` | `running` | `true` |
| `perf.import_xlsx_discovery.responsive.v2` | 359.8 | `progress_with_cancel` | `queued` | `true` |

Both permitted branches occurred without forcing either. Successful completion
is not cancellation evidence; functional cancellation remains independently
covered by the complete Imports audit.

### Exact final commands

All commands ran from the repository root with fresh run IDs. The retained
manifests contain the admitted defaults, exact row selections and source
identity. Audit manifests supplied under `/tmp` are invocation inputs; the
retained audit reports preserve their accepted target/root bindings and exact
required rows. They are not compatibility aliases for historical evidence.

Pre-freeze development/generation commands (successful, native exit 0):

```sh
CARTULARY_TEST_RUN_ID=nf-validation-spec-01 make lint-markdown
CARTULARY_TEST_RUN_ID=nf-validation-generate-09 make generate
CARTULARY_TEST_RUN_ID=nf-validation-generated-check-04 make source-check CHECK=generated
CARTULARY_TEST_RUN_ID=nf-validation-schema-check-06 make source-check CHECK=schema
CARTULARY_TEST_RUN_ID=nf-validation-finalize-04 make agent-finalize
```

The final `make check` also verifies current generated/schema projections after
finalization. The development runs above do not substitute for the compatible
current-source audit providers.

<details>
<summary>Literal current-source qualification commands and row selections</summary>

```sh
CARTULARY_TEST_RUN_ID=nf-validation-check-04 make check
CARTULARY_TEST_RUN_ID=nf-validation-networkflow-backend-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.aggregate_bounds_fixture_compares_legal_source_c_7934db1cd2,module.networkflow.integration.approved_mapping_apply_rejects_forbidden_variant_c2e96f179f,module.networkflow.integration.atomic_fixture_compares_every_commit_boundary_ro_a4cdb920d4,module.networkflow.integration.audit_fixture_compares_publication_binding_and_g_5d2f1cd15b,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.integration.canonical_endpoint_storage,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.canonical_migration_preserves_immutable_results_c092f26ac0,module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2,module.networkflow.integration.cross_table_graph_fixture_compares_exact_importe_be02f892b9,module.networkflow.integration.direction_evidence_storage,module.networkflow.integration.effective_resource_limit_discovery,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.graph_result_cleanup,module.networkflow.integration.graph_result_cleanup_races,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9,module.networkflow.integration.indicator_mismatch_fixture_compares_authorizatio_eb91e61505,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.large_limits_fixture_compares_valid_imported_row_470de5552b,module.networkflow.integration.mapping_candidate_public_preview_and_approval_re_145ec783fa,module.networkflow.integration.mapping_policy_null_and_empty_requests_reject_wi_c863b856b4,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.naming_fixture_compares_imports_filename_normali_d840a6c98d,module.networkflow.integration.network_flow_routes_remain_unclaimed_by_default_e8dabf0ee1,module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488,module.networkflow.integration.pagination_recovery,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388,module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5,module.networkflow.integration.resource_intent_failure_rolls_back_source_mutation_8a4c3d2e1f,module.networkflow.integration.resource_limits_fixture_compares_application_con_77d0adda0f,module.networkflow.integration.restored_network_flow_state_admits_metadata_and_0985ff0a17,module.networkflow.integration.retention_fixture_compares_soft_delete_incident_570ae9035a,module.networkflow.integration.saved_graph_cutover_v6,module.networkflow.integration.saved_graph_lifecycle_v2,module.networkflow.integration.source_changed_fixture_compares_every_source_ver_213b5f7eab,module.networkflow.integration.stale_graph_fixture_compares_soft_deletion_contr_5ee08e6e49,module.networkflow.integration.table_lifecycle_admission_replay,module.networkflow.integration.time_bucket_saved_graph_lifecycle,module.networkflow.measurement.backend_graph_capacity_certification,module.networkflow.process.the_packaged_standalone_server_composes_the_netw_400a31ad27,module.networkflow.support_integration.network_flow_audit_assertion_route_disabled_by_d_611b731957,module.networkflow.support_integration.network_flow_auth_transition_route_disabled_by_d_07779f71d9,module.networkflow.support_integration.network_flow_fault_route_disabled_by_default_8b0741837f,module.networkflow.support_integration.network_flow_randomness_route_disabled_by_defaul_a63c6ef32a
CARTULARY_TEST_RUN_ID=nf-validation-networkflow-browser-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.accessibility.exploration_navigation,module.networkflow.accessibility.indicator_link_keyboard_recovery,module.networkflow.accessibility.verify_claimed_network_analysis_keyboard_focus_a_c81856fdc1,module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179,module.networkflow.browser.import_recovery,module.networkflow.browser.import_retained_draft,module.networkflow.browser.indicator_link_recovery,module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619,module.networkflow.browser.network_flow_selector_covers_default_graph_mode_fc65f0ecfd,module.networkflow.browser.network_flow_selector_covers_importing_nf_fix_00_094b0f779b,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_e_346ec21519,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_v_171429ad9e,module.networkflow.browser.network_flow_selector_covers_soft_deleting_a_tab_537fbd618e,module.networkflow.browser.network_flow_selector_covers_the_filter_selector_12b0cf1448,module.networkflow.browser.network_flow_selector_covers_when_the_extension_4ce49e0d6f,module.networkflow.browser.network_flow_selector_covers_when_the_extension_c90fca78ee,module.networkflow.browser.query_authoring_integrity,module.networkflow.browser.table_lifecycle_continuity,module.networkflow.browser.verify_claimed_network_analysis_discovery_import_f977242343,module.networkflow.browser_stateful.exploration_navigation,module.networkflow.browser_stateful.pagination_expiry_bridge,module.networkflow.browser_stateful.pagination_recovery,module.networkflow.browser_stateful.saved_graph_deferred_navigation,module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle,module.networkflow.browser_stateful.saved_graph_read_recovery,module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf,module.networkflow.measurement.exploration_graph_dom_ceiling,module.networkflow.measurement.measure_bounded_dom_growth_and_semantic_reachabi_297b6a7322,module.networkflow.measurement.network_flow_selector_covers_the_actual_accepted_f68f4c7a57,module.networkflow.measurement.saved_graph_dom_ceiling,module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6
CARTULARY_TEST_RUN_ID=nf-validation-recovery-coverage-01 make test-slice OWNER=module.recovery ROWS=module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0
CARTULARY_TEST_RUN_ID=nf-validation-migrations-coverage-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.fail_closed_source_and_remediation_contract,module.database_migrations.unit.migration_evidence_projection_semantics,module.database_migrations.unit.migration_evidence_source_audit,module.database_migrations.unit.owner_private_provider_policy,module.database_migrations.unit.production_ddl_v3_recurrence,module.database_migrations.unit.provider_source_and_safe_error_isolation,module.database_migrations.unit.public_surface_and_reader_capability,module.database_migrations.unit.schema_bootstrap_guard,module.database_migrations.unit.test_harness_targeted_operation_validation
CARTULARY_TEST_RUN_ID=nf-validation-extensions-coverage-01 make test-slice OWNER=module.extensions ROWS=module.extensions.browser_stateful.authorization_invalidation_retires_stale_extensi_e73d3390e3,module.extensions.browser_stateful.bc015_availability_continuity_d538000c38
CARTULARY_TEST_RUN_ID=nf-validation-imports-coverage-01 make test-slice OWNER=module.imports ROWS=module.imports.browser.base_workbook_omits_unclaimed_import_assistant_6f271c8b42,module.imports.browser.claimed_import_assistant_production_workflow_986c2117ee,module.imports.browser.density_containment,module.imports.browser.keyboard_recovery,module.imports.browser.partial_outcomes,module.imports.measurement.csv_apply,module.imports.measurement.csv_discovery,module.imports.measurement.xlsx_apply,module.imports.measurement.xlsx_discovery
CARTULARY_TEST_RUN_ID=nf-validation-boundaries-final-01 make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries
CARTULARY_TEST_RUN_ID=nf-validation-audit-networkflow-01 make test-evidence-audit OWNER=module.networkflow EVIDENCE_ROOTS_FILE=/tmp/nf-validation-evidence-module.networkflow.json
CARTULARY_TEST_RUN_ID=nf-validation-audit-recovery-01 make test-evidence-audit OWNER=module.recovery EVIDENCE_ROOTS_FILE=/tmp/nf-validation-evidence-module.recovery.json
CARTULARY_TEST_RUN_ID=nf-validation-audit-migrations-01 make test-evidence-audit OWNER=module.database_migrations EVIDENCE_ROOTS_FILE=/tmp/nf-validation-evidence-module.database_migrations.json
CARTULARY_TEST_RUN_ID=nf-validation-audit-extensions-01 make test-evidence-audit OWNER=module.extensions EVIDENCE_ROOTS_FILE=/tmp/nf-validation-evidence-module.extensions.json
CARTULARY_TEST_RUN_ID=nf-validation-audit-imports-01 make test-evidence-audit OWNER=module.imports EVIDENCE_ROOTS_FILE=/tmp/nf-validation-evidence-module.imports.json
```

</details>

No selected commit-qualification obligation remains skipped. Full release/package
qualification, package inspection and deployment on the resulting source remain
unperformed by explicit scope; earlier release evidence retains its original
source identity. There is no unresolved implementation or specification finding
from this remediation. Historical environments and retained evidence are
unchanged; no database reset, state conversion, commit or deployment occurred.

Documentation-only validation `CARTULARY_TEST_RUN_ID=nf-validation-spec-02 make lint-markdown`
passed with native exit 0. After correcting its retained-summary link, the final
documentation command is `CARTULARY_TEST_RUN_ID=nf-validation-spec-03 make lint-markdown`,
with its [retained summary](../../.cartulary/test-results/nf-validation-spec-03/adhoc/lint-markdown/tool-run-summary.json).

## Graph algorithm naming follow-up — 2026-11-05

The final review found a nonblocking projection inconsistency: the authored
[graph semantics contract](../../contracts/network-flow/graph-semantics.v3.json)
used `network_flow_edge_id_v2`, while NF-REQ-135a names the algorithm
`network_flow_flow_edge_id_v2`. The maintainer instructed correction,
regeneration, focused verification and this handoff update, without committing.

The contract now uses the adopted name. `make generate` refreshed the embedded
JSON and artifact SHA-256 in
[the generated Go projection](../../internal/gen/contractnetworkflow/artifacts_gen.go).
Those are the only two executable files changed from the preceding reviewed
snapshot. The domain separator, framing, ordered hash inputs and edge-ID prefixes
are unchanged. This is a projection-name correction, with no algorithm version,
public-major, state-version or migration change and no compatibility alias.

The successor executable snapshot is
`sha256:f012186771c4ce35bc94983d887316431b83f0f2751b68623cba56a1232e7883`,
dirty at commit `e6ca992e634147fd257f8a7462026a80053cd1d8`.
Finalization left that snapshot unchanged. The focused test invocation and both
post-finalization source checks carry this same source identity. The previous
full check, five complete owner audits and human approval remain bound to
`sha256:fc975641e456db417dcf6d436314887692209aec00d24d5babf7033d5c191c85`;
they have not been relabeled as successor-source evidence or fresh human review.

Verification was selected through `make task-guide OWNER=module.networkflow`
and its coverage discovery. The three existing selected rows cover the graph
contract projection, bounded default aggregation/selectors, and temporal
aggregation/identity fixtures. No new test duplicates the metadata assignment.

| Invocation and retained summary | Outcome | Completed cleanup steps |
| --- | --- | ---: |
| [nf-edge-name-generate-01](../../.cartulary/test-results/nf-edge-name-generate-01/generate/tool-run-summary.json) | Generation passed; native Make exit 0. | Not a graph invocation. |
| [nf-edge-name-focused-01](../../.cartulary/test-results/nf-edge-name-focused-01/run-summary.json) | All three selected rows passed in one work unit; native Make exit 0. | 10 |
| [nf-edge-name-finalize-01](../../.cartulary/test-results/nf-edge-name-finalize-01/run-summary.json) | Finalization passed; native Make exit 0. | 10 |
| [nf-edge-name-generated-01](../../.cartulary/test-results/nf-edge-name-generated-01/run-summary.json) | Generated-source checks passed, five work units; native Make exit 0. | 14 |
| [nf-edge-name-schema-01](../../.cartulary/test-results/nf-edge-name-schema-01/run-summary.json) | Schema checks passed, three work units; native Make exit 0. | 12 |

Exact commands, executed serially from the repository root with fresh run IDs:

```sh
CARTULARY_TEST_RUN_ID=nf-edge-name-generate-01 make generate
CARTULARY_TEST_RUN_ID=nf-edge-name-focused-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58
CARTULARY_TEST_RUN_ID=nf-edge-name-finalize-01 make agent-finalize
CARTULARY_TEST_RUN_ID=nf-edge-name-generated-01 make source-check CHECK=generated
CARTULARY_TEST_RUN_ID=nf-edge-name-schema-01 make source-check CHECK=schema
CARTULARY_TEST_RUN_ID=nf-edge-name-docs-01 make lint-markdown
```

The final documentation check's exit and outcome are retained in its
[tool summary](../../.cartulary/test-results/nf-edge-name-docs-01/adhoc/lint-markdown/tool-run-summary.json).

The naming finding is closed. This follow-up selects focused qualification for a
metadata correction; it does not repeat the full check, browser/visual or Imports
measurements, or complete-owner audits. No runtime algorithm or UI behavior was
modified to require those broader obligations. Their retained results remain
evidence for their original source, and full release/package qualification of
the successor source remains outstanding. No commit or deployment occurred.
